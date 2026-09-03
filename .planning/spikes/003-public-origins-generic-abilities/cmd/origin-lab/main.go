package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"ai-lang/public-origins-generic-abilities/origins"
)

type report struct {
	Schema       int                      `json:"schema"`
	FixturePass  int                      `json:"fixture_pass"`
	FixtureFail  int                      `json:"fixture_fail"`
	ProducerPass int                      `json:"producer_rejection_pass"`
	ProducerFail int                      `json:"producer_rejection_fail"`
	Generated    generatedReport          `json:"generated"`
	Encodings    []origins.EncodingMetric `json:"encodings"`
	Scale        scaleReport              `json:"scale"`
	Signatures   map[string][]string      `json:"signatures,omitempty"`
	ElapsedNanos int64                    `json:"elapsed_nanos"`
}

type generatedReport struct {
	Cases         int                 `json:"cases"`
	Mismatch      *origins.Comparison `json:"mismatch,omitempty"`
	MinimalInput  *origins.CallCase   `json:"minimal_input,omitempty"`
	InjectedFault bool                `json:"injected_fault"`
	Detected      bool                `json:"detected"`
}

type scaleReport struct {
	Origins            int   `json:"origins"`
	InterfaceBytes     int   `json:"interface_bytes"`
	ValueOriginTokens  int   `json:"value_origin_tokens"`
	RegionTokens       int   `json:"region_tokens"`
	ConsumerCheckNanos int64 `json:"consumer_check_nanos"`
}

func main() {
	fixturesPath := flag.String("fixtures", "fixtures/cases.json", "path to fixture JSON")
	inject := flag.Bool("inject-origin-bug", false, "omit one union origin to prove mismatch detection")
	showSignatures := flag.Bool("show-signatures", false, "include all three rendered signature sets")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()

	started := time.Now()
	file, err := origins.LoadFixtures(*fixturesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	public, producerDiagnostics := origins.ExportInterface(file, origins.ProducerOptions{})
	if len(producerDiagnostics) != 0 {
		fmt.Fprintln(os.Stderr, "honest fixture rejected:", producerDiagnostics)
		os.Exit(2)
	}
	public, err = origins.RoundTripInterface(public)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	implementations, err := origins.FunctionMap(file.Functions)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	output := report{Schema: 1}
	for _, call := range file.Cases {
		comparison := origins.Compare(public, implementations, call)
		matches := comparison.Agreement && comparison.Consumer.Valid == call.ExpectValid
		if call.ExpectCode != "" {
			matches = matches && comparison.Consumer.Diagnostic != nil && comparison.Consumer.Diagnostic.Code == call.ExpectCode
		}
		if matches {
			output.FixturePass++
		} else {
			output.FixtureFail++
		}
	}
	for _, function := range file.DishonestFunctions {
		if origins.VerifyProducer(function, function.Body) != nil {
			output.ProducerPass++
		} else {
			output.ProducerFail++
		}
	}

	generated := origins.GenerateCases(file.Functions)
	if *inject {
		search := origins.FindInjectedMismatch(file)
		output.Generated = generatedReport{Cases: search.Programs, Mismatch: search.Mismatch, MinimalInput: search.Input, InjectedFault: true, Detected: search.Mismatch != nil}
	} else {
		for _, call := range generated {
			comparison := origins.Compare(public, implementations, call)
			if !comparison.Agreement {
				copy := call
				output.Generated = generatedReport{Cases: len(generated), Mismatch: &comparison, MinimalInput: &copy}
				break
			}
		}
		if output.Generated.Cases == 0 {
			output.Generated.Cases = len(generated)
		}
	}

	if *showSignatures {
		output.Signatures = map[string][]string{}
	}
	for _, approach := range []string{origins.ValueOrigins, origins.ExplicitRegions, origins.CallbackOnly} {
		lines, metric := origins.RenderAll(file.Functions, approach)
		output.Encodings = append(output.Encodings, metric)
		if *showSignatures {
			output.Signatures[approach] = lines
		}
	}

	large := origins.LargeUnionFunction(1000)
	largeFile := origins.FixtureFile{Schema: 1, Types: file.Types, Functions: []origins.Function{large}}
	largePublic, diagnostics := origins.ExportInterface(largeFile, origins.ProducerOptions{})
	if len(diagnostics) != 0 {
		fmt.Fprintln(os.Stderr, diagnostics)
		os.Exit(2)
	}
	interfaceBytes, _ := json.Marshal(largePublic)
	call := largeCall(large)
	checkStarted := time.Now()
	checked := origins.ConsumerCheck(largePublic, call)
	checkNanos := time.Since(checkStarted).Nanoseconds()
	_, valueMetric := origins.RenderAll([]origins.Function{large}, origins.ValueOrigins)
	_, regionMetric := origins.RenderAll([]origins.Function{large}, origins.ExplicitRegions)
	output.Scale = scaleReport{Origins: 1000, InterfaceBytes: len(interfaceBytes), ValueOriginTokens: valueMetric.LexicalTokens, RegionTokens: regionMetric.LexicalTokens, ConsumerCheckNanos: checkNanos}
	if !checked.Valid {
		output.FixtureFail++
	}
	output.ElapsedNanos = time.Since(started).Nanoseconds()

	var data []byte
	if *pretty {
		data, err = json.MarshalIndent(output, "", "  ")
	} else {
		data, err = json.Marshal(output)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(string(data))
	if output.FixtureFail != 0 || output.ProducerFail != 0 || (!*inject && output.Generated.Mismatch != nil) || *inject {
		os.Exit(1)
	}
}

func largeCall(function origins.Function) origins.CallCase {
	call := origins.CallCase{ID: "SCALE-001", Function: function.ID, Arguments: map[string]string{}, Steps: []origins.Step{{Kind: "use_result"}}}
	for _, parameter := range function.Params {
		binding := "arg_" + parameter.Name
		call.Arguments[parameter.Name] = binding
		call.Bindings = append(call.Bindings, origins.Binding{ID: binding, Type: parameter.Type})
	}
	return call
}

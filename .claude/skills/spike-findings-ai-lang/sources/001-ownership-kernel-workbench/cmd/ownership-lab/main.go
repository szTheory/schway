package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"ai-lang/ownership-kernel-workbench/ownership"
)

type report struct {
	Schema       int                        `json:"schema"`
	Fixtures     []ownership.Comparison     `json:"fixtures"`
	FixturePass  int                        `json:"fixture_pass"`
	FixtureFail  int                        `json:"fixture_fail"`
	FlowFixtures []ownership.FlowComparison `json:"flow_fixtures"`
	FlowPass     int                        `json:"flow_pass"`
	FlowFail     int                        `json:"flow_fail"`
	Search       ownership.SearchResult     `json:"search"`
	ElapsedNanos int64                      `json:"elapsed_nanos"`
}

func main() {
	fixturesPath := flag.String("fixtures", "fixtures/cases.json", "path to fixture JSON")
	flowFixturesPath := flag.String("flow-fixtures", "fixtures/flow-cases.json", "path to control-flow fixture JSON")
	maxDepth := flag.Int("max-depth", 4, "maximum generated operation depth")
	injectBug := flag.Bool("inject-checker-bug", false, "allow moves while shared to prove mismatch detection")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()

	started := time.Now()
	file, err := ownership.LoadFixtures(*fixturesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	flowFile, err := ownership.LoadFlowFixtures(*flowFixturesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	options := ownership.CheckerOptions{AllowMoveWhileShared: *injectBug}
	output := report{Schema: 1, Fixtures: make([]ownership.Comparison, 0, len(file.Fixtures))}
	for _, fixture := range file.Fixtures {
		comparison := ownership.Compare(fixture.Program, options)
		output.Fixtures = append(output.Fixtures, comparison)
		matchesExpectation := comparison.Agreement && comparison.Oracle.Valid == fixture.ExpectValid
		if !fixture.ExpectValid && comparison.Oracle.Diagnostic != nil {
			matchesExpectation = matchesExpectation && comparison.Oracle.Diagnostic.Code == fixture.ExpectCode
		}
		if matchesExpectation {
			output.FixturePass++
		} else {
			output.FixtureFail++
		}
	}
	for _, fixture := range flowFile.Fixtures {
		comparison := ownership.CompareFlow(fixture.Program, ownership.FlowOptions{})
		output.FlowFixtures = append(output.FlowFixtures, comparison)
		matchesExpectation := comparison.Agreement && comparison.Analyzer.Valid == fixture.ExpectValid
		if !fixture.ExpectValid && comparison.Analyzer.Diagnostic != nil {
			matchesExpectation = matchesExpectation && comparison.Analyzer.Diagnostic.Code == fixture.ExpectCode
		}
		if matchesExpectation {
			output.FlowPass++
		} else {
			output.FlowFail++
		}
	}
	output.Search = ownership.FindMismatch(*maxDepth, options)
	output.ElapsedNanos = time.Since(started).Nanoseconds()

	var bytes []byte
	if *pretty {
		bytes, err = json.MarshalIndent(output, "", "  ")
	} else {
		bytes, err = json.Marshal(output)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Println(string(bytes))
	if output.FixtureFail != 0 || output.FlowFail != 0 || (!*injectBug && output.Search.Mismatch != nil) || (*injectBug && output.Search.Mismatch == nil) {
		os.Exit(1)
	}
}

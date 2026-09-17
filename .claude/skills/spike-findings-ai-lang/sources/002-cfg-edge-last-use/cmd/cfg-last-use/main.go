package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"ai-lang/cfg-edge-last-use/cfg"
)

type report struct {
	Schema       int             `json:"schema"`
	FixturePass  int             `json:"fixture_pass"`
	FixtureFail  int             `json:"fixture_fail"`
	Fixtures     []fixtureReport `json:"fixtures"`
	Generated    generatedReport `json:"generated"`
	LargeCFG     largeCFGReport  `json:"large_cfg"`
	ElapsedNanos int64           `json:"elapsed_nanos"`
}

type fixtureReport struct {
	Program       string `json:"program"`
	Agreement     bool   `json:"agreement"`
	AnalyzerValid bool   `json:"analyzer_valid"`
	OracleValid   bool   `json:"oracle_valid"`
	Paths         int    `json:"paths"`
	Endpoints     int    `json:"endpoints"`
}

type generatedReport struct {
	Programs int             `json:"programs"`
	Mismatch *cfg.Comparison `json:"mismatch,omitempty"`
	Input    *cfg.Program    `json:"minimal_input,omitempty"`
	Injected bool            `json:"injected_fault"`
	Detected bool            `json:"detected"`
}

type largeCFGReport struct {
	Valid               bool `json:"valid"`
	Blocks              int  `json:"blocks"`
	Edges               int  `json:"edges"`
	Loans               int  `json:"loans"`
	Endpoints           int  `json:"endpoints"`
	TransferEvaluations int  `json:"transfer_evaluations"`
}

func main() {
	fixturesPath := flag.String("fixtures", "fixtures/cases.json", "path to CFG fixture JSON")
	injectBug := flag.Bool("inject-edge-bug", false, "omit edge endpoints to prove mismatch detection")
	pretty := flag.Bool("pretty", true, "pretty-print JSON")
	flag.Parse()

	started := time.Now()
	file, err := cfg.LoadFixtures(*fixturesPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	options := cfg.Options{OmitEdgeEnds: *injectBug}
	output := report{Schema: 1, Fixtures: []fixtureReport{}}
	for _, fixture := range file.Fixtures {
		comparison := cfg.Compare(fixture.Program, options, 3)
		output.Fixtures = append(output.Fixtures, fixtureReport{
			Program: comparison.Program, Agreement: comparison.Agreement,
			AnalyzerValid: comparison.AnalyzerValid, OracleValid: comparison.OracleValid,
			Paths: len(comparison.Paths), Endpoints: len(comparison.Analysis.Endpoints),
		})
		matches := comparison.Agreement && comparison.AnalyzerValid == fixture.ExpectValid
		if matches {
			output.FixturePass++
		} else {
			output.FixtureFail++
		}
	}
	search := cfg.FindMismatch(options)
	output.Generated = generatedReport{
		Programs: search.Programs, Mismatch: search.Mismatch, Input: search.MinimalInput,
		Injected: *injectBug, Detected: search.Mismatch != nil,
	}
	large := cfg.Analyze(cfg.LargeDiamondChain(500), cfg.Options{})
	output.LargeCFG = largeCFGReport{
		Valid: large.Valid, Blocks: large.Blocks, Edges: large.Edges, Loans: large.Loans,
		Endpoints: len(large.Endpoints), TransferEvaluations: large.TransferEvaluations,
	}
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
	if output.FixtureFail != 0 || (!*injectBug && output.Generated.Mismatch != nil) || (*injectBug && output.Generated.Mismatch == nil) || !output.LargeCFG.Valid {
		os.Exit(1)
	}
}

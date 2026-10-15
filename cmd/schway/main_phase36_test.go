package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase36PublicCommand struct {
	Schema     string                `json:"schema"`
	Command    string                `json:"command"`
	Status     string                `json:"status"`
	Executions []execution.Execution `json:"executions"`
}

func TestPhase36PublicHelperConformance(t *testing.T) {
	path := testsupport.ProjectPath("examples", "helper_calls", "public.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read public helper witness: %v", err)
	}
	if strings.Count(string(source), "unused") != 1 {
		t.Fatal("public helper witness must keep its Byte parameter unused")
	}
	for _, engine := range []string{"interpreter", "native"} {
		t.Run(engine, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"--json", "run", "--engine=" + engine, path})
			if code != exitSuccess || stderr != "" {
				t.Fatalf("public %s helper run code=%d stderr=%q stdout=%q", engine, code, stderr, stdout)
			}
			var command phase36PublicCommand
			if err := json.Unmarshal([]byte(stdout), &command); err != nil {
				t.Fatalf("decode public %s helper result: %v; output=%q", engine, err, stdout)
			}
			if command.Command != "run" || command.Status != "pass" || len(command.Executions) != 1 {
				t.Fatalf("public %s helper schema=%q command=%q status=%q executions=%d", engine, command.Schema, command.Command, command.Status, len(command.Executions))
			}
			got := command.Executions[0]
			if got.Outcome.Kind != execution.OutcomeReturned || got.Outcome.Value != "1994146192" {
				t.Fatalf("public %s projection outcome=%+v, want returned value 1994146192", engine, got.Outcome)
			}
			calls, returns := 0, 0
			for _, event := range got.Events {
				if event.Kind == "function.called" && strings.Contains(event.CalleeFunctionID, "crc_step:fn:crc_step") {
					calls++
				}
				if event.Kind == "function.returned" {
					returns++
				}
			}
			if calls != 2 || returns != 3 || len(got.LiveResources) != 0 {
				t.Fatalf("public %s helper calls=%d returns=%d live=%v, want 2 helper calls, 3 returns, and no live resources", engine, calls, returns, got.LiveResources)
			}
		})
	}

	for _, entry := range []struct {
		name string
		file string
	}{
		{name: "local-u64-entry", file: "crc_step.schway"},
		{name: "imported-u64-entry", file: "importer.schway"},
	} {
		t.Run(entry.name, func(t *testing.T) {
			entryPath := testsupport.ProjectPath("examples", "helper_calls", entry.file)
			outputs := map[string]string{}
			for _, engine := range []string{"interpreter", "native"} {
				code, stdout, stderr := captureLangRun(t, []string{"--json", "run", "--engine=" + engine, entryPath})
				if code != 64 || stderr != "" || !strings.Contains(stdout, "tool.run_input_required") {
					t.Fatalf("%s %s U64 entry without input: code=%d stdout=%q stderr=%q, want tool.run_input_required exit 64", entry.name, engine, code, stdout, stderr)
				}
				if strings.Contains(stdout, "module.imports_not_loaded") {
					t.Fatalf("%s %s refused its import instead of reaching the input-required result: %s", entry.name, engine, stdout)
				}
				outputs[engine] = stdout
			}
			if outputs["interpreter"] != outputs["native"] {
				t.Fatalf("%s U64 input-required output differs by engine:\n%s\n%s", entry.name, outputs["interpreter"], outputs["native"])
			}
		})
	}
}

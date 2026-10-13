package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase34SurfaceRoute(t *testing.T) {
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	empty := t.TempDir()
	if err := os.Chdir(empty); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousWD); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
	for _, args := range [][]string{{"surface", "--json"}, {"--json", "surface"}} {
		var first string
		for repeat := 0; repeat < 2; repeat++ {
			code, stdout, stderr := captureLangRun(t, args)
			if code != exitcode.Success {
				t.Fatalf("%v exit = %d, want 0; stdout=%q stderr=%q", args, code, stdout, stderr)
			}
			if stderr != "" {
				t.Fatalf("%v wrote stderr %q", args, stderr)
			}
			if !strings.HasSuffix(stdout, "\n") || strings.HasSuffix(stdout, "\n\n") {
				t.Fatalf("%v output must have exactly one trailing newline, got %q", args, stdout)
			}
			var document map[string]json.RawMessage
			if err := json.Unmarshal([]byte(stdout), &document); err != nil {
				t.Fatalf("%v output is not standalone JSON: %v", args, err)
			}
			var schema string
			if err := json.Unmarshal(document["schema"], &schema); err != nil || schema != "schway.surface/1" {
				t.Fatalf("%v schema = %q, want schway.surface/1 (decode error %v)", args, schema, err)
			}
			if len(document) != 10 {
				t.Fatalf("%v top-level field count = %d, want standalone ten-field manifest", args, len(document))
			}
			for _, key := range []string{"schema", "coordinates", "types", "operators", "statements", "builtins", "calls", "restrictions", "limits", "refusals"} {
				if _, ok := document[key]; !ok {
					t.Errorf("%v missing manifest field %q", args, key)
				}
			}
			if repeat == 0 {
				first = stdout
			} else if stdout != first {
				t.Fatalf("%v output changed between identical invocations", args)
			}
		}
	}
	if entries, err := os.ReadDir(empty); err != nil || len(entries) != 0 {
		t.Fatalf("source-free route created files in empty working directory: entries=%v err=%v", entries, err)
	}
	for _, args := range [][]string{
		{"surface"},
		{"surface", "--json", "source.schway"},
		{"surface", "--json", "--json"},
		{"surface", "--json", "--depth=3"},
		{"surface", "--json", "--kind=symbol"},
		{"surface", "--json", "--cursor=x"},
		{"surface", "--json", "--expand"},
		{"surface", "source.schway", "--json"},
		{"--json", "surface", ""},
		{"--json", "--json", "surface"},
	} {
		code, stdout, stderr := captureLangRun(t, args)
		if problems := phase34SurfaceRawArgumentProblems(args, code, false); len(problems) != 0 {
			t.Errorf("invalid raw args %q: %v (stdout=%q stderr=%q)", args, problems, stdout, stderr)
		}
		if strings.Contains(stdout+stderr, `"schema":"schway.surface/1"`) {
			t.Errorf("invalid raw args %q were accepted as a surface request", args)
		}
		if containsRaw(args, "--json") {
			var usage struct {
				Command string `json:"command"`
				Status  string `json:"status"`
			}
			if err := json.Unmarshal([]byte(stdout), &usage); err != nil || usage.Command != "usage" || usage.Status != "usage_error" || stderr != "" {
				t.Errorf("invalid JSON raw args %q changed usage streams: usage=%+v err=%v stdout=%q stderr=%q", args, usage, err, stdout, stderr)
			}
		} else if stdout != "" || stderr == "" || !strings.Contains(stderr, "usage:") {
			t.Errorf("invalid text raw args %q changed usage streams: stdout=%q stderr=%q", args, stdout, stderr)
		}
	}

	// Simulate the old global-strip-first order locally. It accepts the raw
	// flag after removal, and the raw-argument checker must name that bypass.
	bypassArgs := []string{"surface", "--json", "--depth=3"}
	stripped, _, ok := extractDepth(append([]string(nil), bypassArgs...))
	if !ok {
		t.Fatal("global depth extractor rejected the armed bypass control")
	}
	mutantCode, mutantStdout, mutantStderr := captureLangRun(t, stripped)
	if mutantCode != exitcode.Success || mutantStderr != "" || !strings.Contains(mutantStdout, `"schema": "schway.surface/1"`) {
		t.Fatalf("stripped-raw-flag control did not reach accepted route: code=%d stdout=%q stderr=%q", mutantCode, mutantStdout, mutantStderr)
	}
	if problems := phase34SurfaceRawArgumentProblems(bypassArgs, mutantCode, false); !phase32Contains(problems, "unexpected_flag_acceptance") {
		t.Fatalf("bypass control problems = %v, want unexpected_flag_acceptance", problems)
	}
	cleanCode, cleanOut, cleanErr := captureLangRun(t, bypassArgs)
	if problems := phase34SurfaceRawArgumentProblems(bypassArgs, cleanCode, false); len(problems) != 0 || cleanCode != exitcode.Usage || cleanOut == "" || cleanErr != "" {
		t.Fatalf("raw flag validator did not restore after bypass control: problems=%v code=%d stdout=%q stderr=%q", problems, cleanCode, cleanOut, cleanErr)
	}
}

func TestPhase34SurfaceRouteInventory(t *testing.T) {
	mainSource, err := os.ReadFile(testsupport.ProjectPath("cmd", "schway", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	routes := phase32EnumerateRoutes(string(mainSource))
	entries := phase32EnumerateEntries(testsupport.ProjectPath("examples"))
	testNames := phase32ScanTestNames(testsupport.ProjectPath("cmd"), testsupport.ProjectPath("internal"))
	for _, key := range []string{"surface", "--json surface"} {
		if !containsRaw(routes, key) {
			t.Errorf("missing_route %s: route census = %v", key, routes)
		}
		row, found := phase34SurfaceRouteRow(phase32RouteRows, key)
		if !found || row.Modules != phase32NoEntrySource || !strings.Contains(row.Reason, "TestPhase34SurfaceRoute") {
			t.Errorf("wrong_module_class %s: row=%+v found=%t", key, row, found)
		}
	}
	if problems := phase32RouteAuditProblems(routes, entries, phase32RouteRows, testNames); len(problems) != 0 {
		t.Fatalf("closed route inventory problems: %v", problems)
	}
	for _, key := range []string{"surface", "--json surface"} {
		removed := phase32RouteAuditProblems(routes, entries, phase32WithoutRow(phase32RouteRows, key), testNames)
		if !phase32HasPrefix(removed, "missing_route "+key) {
			t.Errorf("removing %s control problems = %v, want missing_route", key, removed)
		}
		misclassified := append([]phase32RouteRow(nil), phase32RouteRows...)
		for i := range misclassified {
			if misclassified[i].Key == key {
				misclassified[i].Modules = phase32ModuleAware
			}
		}
		if problems := phase32ModuleProblems(misclassified, check.ModuleAwareRoutes(), nil); !phase32HasPrefix(problems, "wrong_module_class "+key) {
			t.Errorf("misclassification control problems = %v, want wrong_module_class %s", problems, key)
		}
	}
}

func phase34SurfaceRouteRow(rows []phase32RouteRow, key string) (phase32RouteRow, bool) {
	for _, row := range rows {
		if row.Key == key {
			return row, true
		}
	}
	return phase32RouteRow{}, false
}

func phase34SurfaceRawArgumentProblems(args []string, exit int, acceptedForm bool) []string {
	if acceptedForm && exit != exitcode.Success {
		return []string{"expected_surface_acceptance"}
	}
	if !acceptedForm && exit == exitcode.Success {
		return []string{"unexpected_flag_acceptance"}
	}
	if !acceptedForm && exit != exitcode.Usage {
		return []string{"wrong_usage_exit"}
	}
	return nil
}

func containsRaw(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
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
		if code != exitcode.Usage {
			t.Errorf("invalid raw args %q exit=%d, want usage %d (stdout=%q stderr=%q)", args, code, exitcode.Usage, stdout, stderr)
		}
		if strings.Contains(stdout+stderr, `"schema":"schway.surface/1"`) {
			t.Errorf("invalid raw args %q were accepted as a surface request", args)
		}
	}
}

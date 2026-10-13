package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
)

func TestPhase34SurfaceRoute(t *testing.T) {
	for _, args := range [][]string{{"surface", "--json"}, {"--json", "surface"}} {
		code, stdout, stderr := captureLangRun(t, args)
		if code != exitcode.Success {
			t.Fatalf("%v exit = %d, want 0; stdout=%q stderr=%q", args, code, stdout, stderr)
		}
		if stderr != "" {
			t.Fatalf("%v wrote stderr %q", args, stderr)
		}
		if !strings.HasSuffix(stdout, "\n") || strings.Count(stdout, "\n") != 1 {
			t.Fatalf("%v output must have exactly one trailing newline, got %q", args, stdout)
		}
		var document struct {
			Schema string `json:"schema"`
		}
		if err := json.Unmarshal([]byte(stdout), &document); err != nil {
			t.Fatalf("%v output is not standalone JSON: %v; output=%q", args, err, stdout)
		}
		if document.Schema != "schway.surface/1" {
			t.Fatalf("%v schema = %q, want schway.surface/1", args, document.Schema)
		}
	}
}

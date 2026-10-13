package session_test

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase34ImportedDiagnosticPresentation(t *testing.T) {
	entry := writeModuleFixture(t, map[string]string{
		"main.schway": "module app.main\n\nimport lib {\n  fn calc\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = calc(n)\n  r\n}\n",
		"lib.schway":  "module lib\n\nexport {\n  fn calc\n}\n\nfn calc(n: U64) -> U64 {\n  let d = n + @\n  d\n}\n",
	})

	reads := map[string]int{}
	restore := modload.SetReadObserverForTest(func(module string) { reads[module]++ })
	defer restore()

	result, err := session.CheckCommandFile(entry)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusInvalid || len(result.Diagnostics) == 0 {
		t.Fatalf("CheckCommandFile status=%q diagnostics=%v, want an imported syntax error", result.Status, result.Diagnostics)
	}
	if reads[""] != 1 || reads["lib"] != 1 || len(reads) != 2 {
		t.Fatalf("loader reads = %v, want entry and lib exactly once", reads)
	}

	encoded := phase30ResultJSON(t, result)
	var wire struct {
		Diagnostics []struct {
			Unit        string           `json:"unit"`
			PrimarySpan diagnostic.Span `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(encoded), &wire); err != nil {
		t.Fatal(err)
	}
	if len(wire.Diagnostics) == 0 {
		t.Fatalf("JSON has no diagnostics: %s", encoded)
	}
	problem := wire.Diagnostics[0]
	if problem.Unit != "lib" {
		t.Fatalf("JSON unit = %q, want dotted module name lib: %s", problem.Unit, encoded)
	}
	if problem.PrimarySpan.Line != 8 || problem.PrimarySpan.Col != 13 {
		t.Fatalf("imported primary location = %d:%d, want 8:13", problem.PrimarySpan.Line, problem.PrimarySpan.Col)
	}
	line := "  let d = n + @"
	wantStart := strings.Index(line, "@")
	if problem.PrimarySpan.Start != wantStart+len("module lib\n\nexport {\n  fn calc\n}\n\nfn calc(n: U64) -> U64 {\n") || problem.PrimarySpan.End != problem.PrimarySpan.Start+1 {
		t.Fatalf("imported byte span = [%d:%d], want the @ byte", problem.PrimarySpan.Start, problem.PrimarySpan.End)
	}

	human, err := protocol.Human(result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(human, "lib.schway:8:13") || !strings.Contains(human, line) {
		t.Fatalf("human output does not show imported path, position, and source line:\n%s", human)
	}
	if strings.Contains(human, filepath.Dir(entry)) || strings.Contains(encoded, filepath.Dir(entry)) {
		t.Fatalf("diagnostic presentation leaked a canonical absolute path:\n%s\n%s", human, encoded)
	}
}

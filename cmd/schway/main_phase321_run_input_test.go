package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPhase321RunInputRequiredCLI pins D-01/D-06 through the CLI: both engines
// exit 64 with identical JSON and neither echoes the path or parameter.
func TestPhase321RunInputRequiredCLI(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture_main.schway")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	outputs := map[string]string{}
	for _, engine := range []string{"interpreter", "native"} {
		code, stdout, stderr := captureLangRun(t, []string{"--json", "run", "--engine=" + engine, path})
		if code != 64 {
			t.Fatalf("%s: exit %d, want 64 (stdout %s stderr %s)", engine, code, stdout, stderr)
		}
		if !strings.Contains(stdout, "tool.run_input_required") {
			t.Fatalf("%s: missing code in %s", engine, stdout)
		}
		if strings.Contains(stdout+stderr, dir) || strings.Contains(stdout+stderr, "n:") {
			t.Fatalf("%s: output echoed path or parameter: %s", engine, stdout)
		}
		outputs[engine] = stdout
	}
	if outputs["interpreter"] != outputs["native"] {
		t.Fatalf("engines print different JSON:\n%s\n%s", outputs["interpreter"], outputs["native"])
	}
}

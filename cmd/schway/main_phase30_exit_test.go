package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
)

// TestPhase30ModuleRefusalExitCodes pins the exit class of a refused module
// program on every module-aware route: an entry whose import names a missing
// module is a source refusal (exitcode.Invalid) that names module.not_found.
// The armed control runs the same four routes on an entry path that does not
// exist, which is operational (exitcode.Operational) and never names
// module.not_found, so each row's assertion can fail.
func TestPhase30ModuleRefusalExitCodes(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.schway")
	source := "module app.main\n\nimport missing {\n  fn f\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  value\n}\n"
	if err := os.WriteFile(entry, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	absent := filepath.Join(dir, "absent.schway")

	routes := []struct {
		name string
		args func(path string) []string
	}{
		{"check", func(path string) []string { return []string{"check", path} }},
		{"build", func(path string) []string { return []string{"build", path, "--output", filepath.Join(dir, "artifact")} }},
		{"run-interpreter", func(path string) []string { return []string{"run", "--engine=interpreter", path} }},
		{"run-native", func(path string) []string { return []string{"run", "--engine=native", path} }},
	}

	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, route.args(entry))
			if code != exitcode.Invalid {
				t.Fatalf("exit code = %d, want exitcode.Invalid (%d)\nstdout=%q stderr=%q", code, exitcode.Invalid, stdout, stderr)
			}
			if !strings.Contains(stdout+stderr, "module.not_found") {
				t.Fatalf("output does not name module.not_found: stdout=%q stderr=%q", stdout, stderr)
			}
		})
	}

	t.Run("control-operational-fault", func(t *testing.T) {
		for _, route := range routes {
			code, stdout, stderr := captureLangRun(t, route.args(absent))
			if code != exitcode.Operational {
				t.Errorf("%s on a missing entry: exit code = %d, want exitcode.Operational (%d)\nstdout=%q stderr=%q", route.name, code, exitcode.Operational, stdout, stderr)
			}
			if code == exitcode.Invalid || strings.Contains(stdout+stderr, "module.not_found") {
				t.Errorf("%s on a missing entry must not look like a refused import: code=%d stdout=%q stderr=%q", route.name, code, stdout, stderr)
			}
		}
	})
}

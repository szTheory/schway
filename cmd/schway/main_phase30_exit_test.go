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

// TestPhase30CapRefusalExitsInvalid pins CR-01 on the CLI: an over-cap module
// program exits exitcode.Invalid naming the cap code, on every module-aware
// route, never module.imports_not_loaded. The within-cap control names no cap.
func TestPhase30CapRefusalExitsInvalid(t *testing.T) {
	write := func(imports int) string {
		dir := t.TempDir()
		var entry strings.Builder
		entry.WriteString("module app.main\n\n")
		for index := 1; index <= imports; index++ {
			name := "m" + strings.Repeat("x", index)
			entry.WriteString("import " + name + " {\n  fn f\n}\n\n")
			lib := "module " + name + "\n\nexport {\n  fn f\n}\n\nfn f(n: U64) -> U64 {\n  n\n}\n"
			if err := os.WriteFile(filepath.Join(dir, name+".schway"), []byte(lib), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		entry.WriteString("export {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  n\n}\n")
		path := filepath.Join(dir, "main.schway")
		if err := os.WriteFile(path, []byte(entry.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	for _, args := range [][]string{{"check"}, {"debug-map"}} {
		t.Run(args[0], func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, append(append([]string{}, args...), write(16)))
			out := stdout + stderr
			if code != exitcode.Invalid || !strings.Contains(out, "module.count_limit") || strings.Contains(out, "module.imports_not_loaded") {
				t.Fatalf("16 imports: exit %d, output %q; want exitcode.Invalid naming module.count_limit", code, out)
			}
		})
	}
	t.Run("control-within-cap", func(t *testing.T) {
		_, stdout, stderr := captureLangRun(t, []string{"check", write(2)})
		if strings.Contains(stdout+stderr, "module.count_limit") {
			t.Fatalf("a within-cap program names module.count_limit: %q", stdout+stderr)
		}
	})
}

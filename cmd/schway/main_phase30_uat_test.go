package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
)

// TestPhase30UATRefusalTranscriptsCLI pins the 30-UAT.md checkpoints 2-5
// through the public `check` command: each refused module program exits
// exitcode.Invalid and names its own code (and `unit=` where the refusal is
// inside an imported file). Each row has an accepted sibling that differs only
// in the refused fact, so a row whose refusal stops firing, or a sibling that
// starts being refused, fails here.
func TestPhase30UATRefusalTranscriptsCLI(t *testing.T) {
	const entry = "module app\n\nimport lib {\n  fn %s\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = %s(n)\n  r\n}\n"
	main := func(name string) string { return strings.ReplaceAll(entry, "%s", name) }
	lib := func(header, body string) string {
		return "module " + header + "\n\nexport {\n  fn api\n}\n\nfn api(n: U64) -> U64 {\n  let r = " + body + "(n)\n  r\n}\n\nfn helper(n: U64) -> U64 {\n  n\n}\n"
	}
	cycleA := "module a\n\nimport b {\n  fn g\n}\n\nexport {\n  fn f\n}\n\nfn f(n: U64) -> U64 {\n  let r = g(n)\n  r\n}\n"
	cycleB := func(importA bool) string {
		head := "module b\n\n"
		if importA {
			head += "import a {\n  fn f\n}\n\n"
		}
		return head + "export {\n  fn g\n}\n\nfn g(n: U64) -> U64 {\n  n\n}\n"
	}
	cycleMain := strings.ReplaceAll(main("f"), "import lib", "import a")
	chain := func(depth int) map[string]string {
		files := map[string]string{}
		for i := 1; i <= depth; i++ {
			name := "m" + strings.Repeat("x", i)
			fn := "f" + strings.Repeat("x", i)
			body := "fn " + fn + "(n: U64) -> U64 {\n  n\n}\n"
			imports := ""
			if i < depth {
				next := strings.Repeat("x", i+1)
				imports = "import m" + next + " {\n  fn f" + next + "\n}\n\n"
				body = "fn " + fn + "(n: U64) -> U64 {\n  let r = f" + next + "(n)\n  r\n}\n"
			}
			files[name+".schway"] = "module " + name + "\n\n" + imports + "export {\n  fn " + fn + "\n}\n\n" + body
		}
		files["main.schway"] = strings.ReplaceAll(main("fx"), "import lib", "import mx")
		return files
	}

	type program struct {
		files   map[string]string
		symlink bool
	}
	rows := []struct {
		name     string
		refused  program
		accepted program
		want     []string
	}{
		{
			name:     "private-name",
			refused:  program{files: map[string]string{"main.schway": main("helper"), "lib.schway": lib("lib", "helper")}},
			accepted: program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "helper")}},
			want:     []string{"module.not_exported", `does not export "helper"`},
		},
		{
			name:     "import-cycle",
			refused:  program{files: map[string]string{"main.schway": cycleMain, "a.schway": cycleA, "b.schway": cycleB(true)}},
			accepted: program{files: map[string]string{"main.schway": cycleMain, "a.schway": cycleA, "b.schway": cycleB(false)}},
			want:     []string{"module.import_cycle", "a -> b -> a"},
		},
		{
			name:     "header-path-mismatch",
			refused:  program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("other", "helper")}},
			accepted: program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "helper")}},
			want:     []string{"module.path_mismatch"},
		},
		{
			name:     "symlinked-module",
			refused:  program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "helper")}, symlink: true},
			accepted: program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "helper")}},
			want:     []string{"module.symlink"},
		},
		{
			name:     "dotdot-path",
			refused:  program{files: map[string]string{"main.schway": strings.ReplaceAll(main("api"), "import lib", "import ..lib"), "lib.schway": lib("lib", "helper")}},
			accepted: program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "helper")}},
			want:     []string{"syntax.expected_module_name"},
		},
		{
			name:     "imported-file-error-names-unit",
			refused:  program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "nope")}},
			accepted: program{files: map[string]string{"main.schway": main("api"), "lib.schway": lib("lib", "helper")}},
			want:     []string{"core.call_callee_unresolved", "unit=lib"},
		},
		{
			name:     "depth-limit",
			refused:  program{files: chain(9)},
			accepted: program{files: chain(8)},
			want:     []string{"module.depth_limit", "limit of 8 modules below the entry"},
		},
	}

	write := func(t *testing.T, p program) string {
		t.Helper()
		dir := t.TempDir()
		for name, content := range p.files {
			path := filepath.Join(dir, name)
			if p.symlink && name == "lib.schway" {
				outside := filepath.Join(t.TempDir(), name)
				if err := os.WriteFile(outside, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Skipf("env:symlink unavailable on this host: %v", err)
				}
				continue
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return filepath.Join(dir, "main.schway")
	}

	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"check", write(t, row.refused)})
			out := stdout + stderr
			if code != exitcode.Invalid {
				t.Fatalf("refused program: exit %d, want exitcode.Invalid (%d)\n%s", code, exitcode.Invalid, out)
			}
			for _, want := range row.want {
				if !strings.Contains(out, want) {
					t.Fatalf("refused program output does not contain %q:\n%s", want, out)
				}
			}
			code, stdout, stderr = captureLangRun(t, []string{"check", write(t, row.accepted)})
			if code != exitcode.Success || !strings.Contains(stdout, "check pass") {
				t.Fatalf("accepted sibling: exit %d, want check pass\n%s%s", code, stdout, stderr)
			}
		})
	}
}

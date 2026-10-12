package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase30CopyExampleDir copies the files of examples/phase30 into a TempDir.
func phase30CopyExampleDir(t *testing.T) string {
	t.Helper()
	src := testsupport.ProjectPath("examples", "phase30")
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, entry.Name()), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// TestPhase30AdlerModuleAppCLI is the public runnable gain (SC1, D-17): the
// Phase 29 Adler-32 program with the library imported from its own module
// file answers `schway app run -- N` with the pinned checksum for every case
// id, through `schway build --manifest`. An armed control mutates the library
// copy and observes a different answer; `schway run` keeps the exit class the
// Phase 29 fixture gives under the same command.
func TestPhase30AdlerModuleAppCLI(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable on this host")
	}
	source := testsupport.ProjectPath("examples", "phase30", "adler32_cli.schway")
	manifest := testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), "adler32")
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("module public build code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for caseID := uint64(0); caseID <= 9; caseID++ {
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok {
			t.Fatalf("case %d has no pinned checksum", caseID)
		}
		input := strconv.FormatUint(caseID, 10)
		t.Run("case-"+input, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", input})
			expected := strconv.FormatUint(want, 10) + "\n"
			if code != 0 || stdout != expected || stderr != "" {
				t.Fatalf("case %s code=%d stdout=%q stderr=%q; want code=0 stdout=%q", input, code, stdout, stderr, expected)
			}
		})
	}
	t.Run("unknown-case", func(t *testing.T) {
		code, stdout, _ := captureLangRun(t, []string{"app", "run", artifact, "--", "10"})
		if code != 65 || stdout != "" {
			t.Fatalf("unknown case code=%d stdout=%q; want code=65 and empty stdout", code, stdout)
		}
	})
	t.Run("check", func(t *testing.T) {
		code, _, stderr := captureLangRun(t, []string{"check", source})
		if code != 0 {
			t.Fatalf("check code=%d stderr=%q; want 0", code, stderr)
		}
	})
	phase29Source := testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway")
	for _, test := range []struct{ engine, code string }{
		{"interpreter", "tool.run_failed"},
		{"native", "native.tool_failure"},
	} {
		t.Run("run-"+test.engine, func(t *testing.T) {
			for _, program := range []string{source, phase29Source} {
				code, stdout, stderr := captureLangRun(t, []string{"run", "--engine=" + test.engine, program})
				if code != 3 || !strings.Contains(stdout+stderr, test.code) {
					t.Fatalf("run --engine=%s %s code=%d stdout=%q stderr=%q; want exit 3 and %s", test.engine, filepath.Base(program), code, stdout, stderr, test.code)
				}
			}
		})
	}
	t.Run("mutated-library", func(t *testing.T) {
		dir := phase30CopyExampleDir(t)
		libraryPath := filepath.Join(dir, "adler32.schway")
		library, err := os.ReadFile(libraryPath)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(library), "var a = 1\n") != 1 {
			t.Fatal("library no longer has exactly one `var a = 1` line")
		}
		mutated := strings.Replace(string(library), "var a = 1\n", "var a = 2\n", 1)
		if err := os.WriteFile(libraryPath, []byte(mutated), 0o600); err != nil {
			t.Fatal(err)
		}
		mutatedArtifact := filepath.Join(t.TempDir(), "adler32-mutated")
		code, stdout, stderr := captureLangRun(t, []string{"build", filepath.Join(dir, "adler32_cli.schway"), "--manifest", manifest, "--output", mutatedArtifact})
		if code != 0 || stdout != "built "+mutatedArtifact+"\n" || stderr != "" {
			t.Fatalf("mutated build code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		want, _ := testsupport.Phase29FixtureExpected(0)
		pinned := strconv.FormatUint(want, 10) + "\n"
		code, stdout, stderr = captureLangRun(t, []string{"app", "run", mutatedArtifact, "--", "0"})
		if code != 0 || stderr != "" {
			t.Fatalf("mutated run code=%d stdout=%q stderr=%q", code, stdout, stderr)
		}
		if stdout == pinned {
			t.Fatalf("mutated library still printed the pinned value %q; the comparison is inert", pinned)
		}
	})
}

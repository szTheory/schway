package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase29FixtureAppCLI is the public runnable gain (D-15, D-16): one
// `schway build --manifest` artifact answers `schway app run -- N` with the
// pinned Adler-32 checksum for every case id, an unknown case exits 65 with
// nothing printed, and `schway run` keeps its Phase 28 classes for the same
// program (D-06).
func TestPhase29FixtureAppCLI(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable on this host")
	}
	source := testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway")
	manifest := testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), "adler32")
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("fixture public build code=%d stdout=%q stderr=%q", code, stdout, stderr)
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
	for _, test := range []struct{ engine, code string }{
		{"interpreter", "tool.run_failed"},
		{"native", "native.tool_failure"},
	} {
		t.Run("run-"+test.engine+"-unchanged", func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"run", "--engine=" + test.engine, source})
			if code != 3 || !strings.Contains(stdout+stderr, test.code) {
				t.Fatalf("run --engine=%s code=%d stdout=%q stderr=%q; want exit 3 and %s (D-06 unchanged)", test.engine, code, stdout, stderr, test.code)
			}
		})
	}
}

// TestPhase29OOBControlAppCLI pins the README out-of-bounds control through
// the public CLI: `schway build --manifest` then `schway app run -- N` reads
// one byte past the end and exits 65 with the exact stderr and empty stdout.
func TestPhase29OOBControlAppCLI(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable on this host")
	}
	source := testsupport.ProjectPath("examples", "phase29", "byte_index_oob.schway")
	manifest := testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), "byte-index-oob")
	code, stdout, stderr := captureLangRun(t, []string{"build", source, "--manifest", manifest, "--output", artifact})
	if code != 0 || stdout != "built "+artifact+"\n" || stderr != "" {
		t.Fatalf("oob control public build code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, input := range []string{"0", "2"} {
		t.Run("case-"+input, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", input})
			if code != 65 || stdout != "" || stderr != "schway: byte index out of bounds\n" {
				t.Fatalf("case %s code=%d stdout=%q stderr=%q; want code=65, empty stdout, byte index stderr", input, code, stdout, stderr)
			}
		})
	}
}

var phase29ReadmeCaseRow = regexp.MustCompile("(?m)^\\| ([0-9]+) \\| [^|]+ \\| ([0-9]+) \\|$")

// TestPhase29ExampleReadmeCaseTable keeps examples/phase29/README.md in step
// with the pinned oracle table: every case row names the pinned checksum, and
// the documented out-of-bounds stderr and exit code match the tested ones.
func TestPhase29ExampleReadmeCaseTable(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("examples", "phase29", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)
	seen := map[uint64]bool{}
	for _, m := range phase29ReadmeCaseRow.FindAllStringSubmatch(readme, -1) {
		caseID, _ := strconv.ParseUint(m[1], 10, 64)
		documented, _ := strconv.ParseUint(m[2], 10, 64)
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok || documented != want {
			t.Errorf("README case %d documents %d; pinned %d (known=%v)", caseID, documented, want, ok)
		}
		seen[caseID] = true
	}
	for _, caseID := range testsupport.Phase29FixtureCaseIDs() {
		if !seen[caseID] {
			t.Errorf("README case table has no row for case %d", caseID)
		}
	}
	for _, claim := range []string{"`schway: byte index out of bounds`", "exit code is\n65", "and exit 65"} {
		if !strings.Contains(readme, claim) {
			t.Errorf("README is missing %q", claim)
		}
	}
}

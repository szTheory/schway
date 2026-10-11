package main

import (
	"os/exec"
	"path/filepath"
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
		t.Skip("clang unavailable on this host")
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

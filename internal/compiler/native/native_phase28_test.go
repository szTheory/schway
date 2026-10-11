package native

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestUnreachableFunctionProgramRuns is the Phase 28 runnable gain (ROADMAP
// SC3, DX-18): programs with functions `main` never calls check, build and
// run instead of failing, and an ambiguous entry is a source refusal (exit 2).
func TestUnreachableFunctionProgramRuns(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skipf("env:clang toolchain unavailable on this host: %v", err)
	}
	binary := testsupport.BuildCLI(t)

	t.Run("unused_function", func(t *testing.T) {
		source := testsupport.ProjectPath("examples", "phase28", "unused_function.schway")
		if result := testsupport.RunCLI(t, binary, nil, "check", source); result.Exit != 0 {
			t.Fatalf("check exit=%d stdout=%s", result.Exit, result.Stdout)
		}
		artifact := filepath.Join(t.TempDir(), "app")
		if result := testsupport.RunCLI(t, binary, nil, "build", source, "--output", artifact); result.Exit != 0 {
			t.Fatalf("build exit=%d stdout=%s stderr=%s", result.Exit, result.Stdout, result.Stderr)
		}
		for _, input := range []string{"7", "0"} {
			result := testsupport.RunCLI(t, binary, nil, "app", "run", artifact, "--", input)
			if result.Exit != 0 || string(result.Stdout) != input+"\n" {
				t.Fatalf("app run %s: exit=%d stdout=%q stderr=%q", input, result.Exit, result.Stdout, result.Stderr)
			}
		}
	})
}

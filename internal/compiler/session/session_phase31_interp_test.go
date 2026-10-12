package session_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase31InterpFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.bin")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func phase31InterpCheck(t *testing.T, path string) session.CheckResult {
	t.Helper()
	checked, err := session.CheckFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics for %s: %+v", path, checked.Diagnostics)
	}
	return checked
}

// TestPhase31ChecksumInterpreter runs the checksum example in the interpreter
// from an independent Go host reader (D-07, D-09, D-16).
func TestPhase31ChecksumInterpreter(t *testing.T) {
	checked := phase31InterpCheck(t, testsupport.ProjectPath("examples", "checksum", "checksum.schway"))

	t.Run("wikipedia", func(t *testing.T) {
		run, err := interp.RunWithFileBytes(checked.Program, "main", phase31InterpFile(t, "Wikipedia"), testsupport.ChecksumReadFile)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != "300286872" || len(run.LiveResources) != 0 {
			t.Fatalf("outcome = %+v live=%v, want returned 300286872 with no live resources", run.Outcome, run.LiveResources)
		}
		released := 0
		for _, event := range run.Events {
			if event.Kind == "resource.released" {
				released++
			}
		}
		if released != 1 {
			t.Fatalf("release events = %d, want 1", released)
		}
	})

	t.Run("missing", func(t *testing.T) {
		run, err := interp.RunWithFileBytes(checked.Program, "main", filepath.Join(t.TempDir(), "absent.bin"), testsupport.ChecksumReadFile)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeTypedFailure || run.Outcome.Value != "FileUnavailable" || len(run.LiveResources) != 0 {
			t.Fatalf("outcome = %+v live=%v, want typed_failure FileUnavailable", run.Outcome, run.LiveResources)
		}
	})

	t.Run("unknown-alternative-refused", func(t *testing.T) {
		_, err := interp.RunWithFileBytes(checked.Program, "main", "x", func(string) ([]byte, string) { return nil, "NotAnAlternative" })
		if err == nil || !strings.Contains(err.Error(), "NotAnAlternative") {
			t.Fatalf("err = %v, want refusal naming the unknown alternative", err)
		}
	})
}

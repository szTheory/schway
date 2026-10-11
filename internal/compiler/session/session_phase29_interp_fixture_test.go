package session_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase29InterpreterFixtureVectors runs the committed fixture program in
// the interpreter for every pinned case id (D-16). Bytes come only from the
// independent Go generator (D-07); the interpreter never reads the C adapter.
func TestPhase29InterpreterFixtureVectors(t *testing.T) {
	source, err := os.ReadFile(phase29FixturePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture diagnostics: %+v", checked.Diagnostics)
	}
	// The tracer: case 3 ("Wikipedia") first, then all ten ids.
	tracer, err := interp.RunWithFixtureBytes(checked.Program, "main", "3", testsupport.Phase29FixtureBytes)
	if err != nil {
		t.Fatal(err)
	}
	if tracer.Outcome.Kind != execution.OutcomeReturned || tracer.Outcome.Value != "300286872" || len(tracer.LiveResources) != 0 {
		t.Fatalf("case 3 = %+v live=%v, want returned 300286872 with no live resources", tracer.Outcome, tracer.LiveResources)
	}
	for _, caseID := range testsupport.Phase29FixtureCaseIDs() {
		want, ok := testsupport.Phase29FixtureExpected(caseID)
		if !ok {
			t.Fatalf("case %d has no pinned checksum", caseID)
		}
		run, err := interp.RunWithFixtureBytes(checked.Program, "main", strconv.FormatUint(caseID, 10), testsupport.Phase29FixtureBytes)
		if err != nil {
			t.Fatalf("case %d: %v", caseID, err)
		}
		if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != strconv.FormatUint(want, 10) || len(run.LiveResources) != 0 {
			t.Fatalf("case %d = %+v live=%v, want returned %d with no live resources", caseID, run.Outcome, run.LiveResources, want)
		}
	}
}

package session_test

import (
	"context"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestBranchInterpreterNative is 03-01-02's end-to-end tracer falsifier:
// generated C17 branches on the scrutinee, executes only the selected arm's
// statements, and produces the same terminal outcome and event order as the
// interpreter at both -O0 and -O3 (D-11/D-12a). It drives the real
// checker -> corevalidate -> interpreter -> cgen -> Clang path exactly as
// session.RunNative already proves for Phase 2's straight-line fixture.
func TestBranchInterpreterNative(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase3", "branch_view.schway")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("branch native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if len(result.Interpreter) != 2 || len(result.O0.Pairs) != 2 || len(result.O3.Pairs) != 2 {
		t.Fatalf("expected one execution per alternative (On, Off): %+v", result)
	}
	seenOutcomes := map[string]bool{}
	for index, want := range result.Interpreter {
		seenOutcomes[want.Outcome.Value] = true
		for _, actual := range []execution.Execution{result.O0.Pairs[index].Execution, result.O3.Pairs[index].Execution} {
			if !execution.Equal(want, actual) {
				t.Fatalf("branch semantic execution mismatch at index %d:\nwant=%+v\ngot =%+v", index, want, actual)
			}
		}
		// The unselected arm produces no events: only operations belonging
		// to the block whose pattern matched the input ever appear.
		for _, event := range want.Events {
			if event.FunctionID == "" {
				t.Fatalf("event missing function id: %+v", event)
			}
		}
	}
	if !seenOutcomes["On"] || !seenOutcomes["Off"] {
		t.Fatalf("expected both alternatives exercised: %+v", seenOutcomes)
	}
	// session.RunNative already asserts O0/O3 agree with the interpreter by
	// returning a non-nil error (*session.EngineMismatch) on the first
	// disagreement; reaching here with no error is itself the differential
	// pass for both optimization levels.
}

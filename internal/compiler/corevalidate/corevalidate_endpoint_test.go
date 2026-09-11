package corevalidate_test

import (
	"reflect"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// validBranchLoanProgram is validBranchProgram's counterpart for a branch
// fixture that actually carries a loan (validBranchProgram's own
// branchTestSource has none): 03-03's shipped
// branch_one_arm_shared_accept.lang produces exactly one real
// core.LoanEndpoint, giving TestLoanEndpointMismatchRejected something to
// corrupt.
func validBranchLoanProgram(t *testing.T) core.Program {
	t.Helper()
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase3", "branch_one_arm_shared_accept.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("baseline branch-loan program rejected: %+v", result.Problems)
	}
	return cloneCoreProgram(t, checked.Program)
}

// TestLoanEndpointMismatchRejected is 03-04-01's own falsifier at the full
// Validate() level: a real, checker-produced branch program's own declared
// LoanEndpoints is corrupted (moved to a different block, dropped, or
// invented for a nonexistent loan) and must be rejected purely because
// recomputeLoanEndpoints disagrees, not because any other invariant caught
// it first.
func TestLoanEndpointMismatchRejected(t *testing.T) {
	baseline := validBranchLoanProgram(t)
	linear := baseline.Functions[0].Linear
	if len(linear.LoanEndpoints) == 0 {
		t.Fatalf("expected the baseline branch fixture to declare at least one loan endpoint: %+v", linear)
	}

	tests := []struct {
		name string
		edit func(*core.Program)
	}{
		{
			name: "endpoint moved to a different block",
			edit: func(program *core.Program) {
				endpoints := program.Functions[0].Linear.LoanEndpoints
				endpoints[0].BlockID = program.Functions[0].Linear.Blocks[len(program.Functions[0].Linear.Blocks)-1].ID
			},
		},
		{
			name: "endpoint dropped entirely",
			edit: func(program *core.Program) {
				program.Functions[0].Linear.LoanEndpoints = program.Functions[0].Linear.LoanEndpoints[1:]
			},
		},
		{
			name: "endpoint invented for a loan that does not exist",
			edit: func(program *core.Program) {
				linear := program.Functions[0].Linear
				linear.LoanEndpoints = append(append([]core.LoanEndpoint(nil), linear.LoanEndpoints...), core.LoanEndpoint{
					ID:               linear.LoanEndpoints[0].ID + ":invented",
					LoanID:           program.Functions[0].ID + ":loan:invented",
					Kind:             "point",
					BlockID:          linear.LoanEndpoints[0].BlockID,
					AfterOperationID: linear.LoanEndpoints[0].AfterOperationID,
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneCoreProgram(t, baseline)
			test.edit(&mutated)
			result := corevalidate.Validate(mutated)
			if result.Valid {
				t.Fatalf("mutation %q was accepted", test.name)
			}
			if len(result.Problems) == 0 || result.Problems[0].Code != "core.loan_endpoint_mismatch" {
				t.Fatalf("mutation %q code=%v want=core.loan_endpoint_mismatch", test.name, result.Problems)
			}
		})
	}
}

// TestLoanEndpointsAccessorMatchesInternalComputation is plan 10-08 Task 1's
// core proof (D-10-52): Result.LoanEndpoints exposes recomputeLoanEndpoints's
// existing computation without re-deriving it a second way. On a real,
// checker-produced, Valid branch-loan program, the checked-and-agreeing
// declared LoanEndpoints (function.Linear.LoanEndpoints) already equals what
// recomputeLoanEndpoints independently computed -- that agreement is exactly
// what Valid: true means for this invariant -- so the accessor's own output
// must equal the declared set too.
func TestLoanEndpointsAccessorMatchesInternalComputation(t *testing.T) {
	program := validBranchLoanProgram(t)
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("baseline branch-loan program rejected: %+v", result.Problems)
	}

	function := program.Functions[0]
	declared := function.Linear.LoanEndpoints
	if len(declared) == 0 {
		t.Fatalf("expected the baseline branch fixture to declare at least one loan endpoint: %+v", function.Linear)
	}

	got := result.LoanEndpoints()[function.ID]
	if !reflect.DeepEqual(got, declared) {
		t.Fatalf("LoanEndpoints()[%s] = %+v, want %+v (the checker's own declared, agreeing set)", function.ID, got, declared)
	}
}

// TestLoanEndpointsAccessorStableAcrossRepeatedCalls proves the accessor's
// output is deterministic: two successive calls on the same Result return
// byte-identical results, and mutating one call's returned map/slices never
// affects the other -- the accessor copies out rather than sharing storage.
func TestLoanEndpointsAccessorStableAcrossRepeatedCalls(t *testing.T) {
	program := validBranchLoanProgram(t)
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("baseline branch-loan program rejected: %+v", result.Problems)
	}

	first := result.LoanEndpoints()
	second := result.LoanEndpoints()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("two successive LoanEndpoints() calls diverged: %+v vs %+v", first, second)
	}

	// Mutate the first call's result; the second call (already taken) and a
	// fresh third call must both be unaffected.
	for functionID := range first {
		first[functionID] = append(first[functionID], core.LoanEndpoint{ID: "mutated"})
		break
	}
	third := result.LoanEndpoints()
	if reflect.DeepEqual(first, third) {
		t.Fatal("mutating a returned map affected a subsequent LoanEndpoints() call -- accessor is not copying out")
	}
	if !reflect.DeepEqual(second, third) {
		t.Fatalf("a call unaffected by mutation still diverged: %+v vs %+v", second, third)
	}
}

// TestLoanEndpointsAccessorEmptyForLoanFreeFunction proves a function with
// no loans -- here, a straight-line function with no Blocks at all --
// yields an empty (nil), non-panicking slice, with an entry still present
// for its own function ID.
func TestLoanEndpointsAccessorEmptyForLoanFreeFunction(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase07", "call_basic.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}

	result := corevalidate.Validate(checked.Program)
	if !result.Valid {
		t.Fatalf("call_basic.lang rejected: %+v", result.Problems)
	}

	endpoints := result.LoanEndpoints()
	if len(endpoints) != len(checked.Program.Functions) {
		t.Fatalf("LoanEndpoints() returned %d entries, want one per function (%d)", len(endpoints), len(checked.Program.Functions))
	}
	for _, function := range checked.Program.Functions {
		got, ok := endpoints[function.ID]
		if !ok {
			t.Fatalf("LoanEndpoints() missing an entry for function %s", function.ID)
		}
		if len(got) != 0 {
			t.Fatalf("LoanEndpoints()[%s] = %+v, want empty (loan-free function)", function.ID, got)
		}
	}
}

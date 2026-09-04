package corevalidate_test

import (
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

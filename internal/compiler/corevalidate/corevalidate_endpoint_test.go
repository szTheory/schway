package corevalidate_test

import (
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// validBranchLoanProgram is validBranchProgram's counterpart for a branch
// fixture that actually carries a loan (validBranchProgram's own
// branchTestSource has none): 03-03's shipped
// branch_one_arm_shared_accept.schway produces exactly one real
// core.LoanEndpoint, giving TestLoanEndpointMismatchRejected something to
// corrupt.
func validBranchLoanProgram(t *testing.T) core.Program {
	t.Helper()
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase3", "branch_one_arm_shared_accept.schway"))
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
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase07", "call_basic.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", checked.Diagnostics)
	}

	result := corevalidate.Validate(checked.Program)
	if !result.Valid {
		t.Fatalf("call_basic.schway rejected: %+v", result.Problems)
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

// TestDerivePeerSignatureModeMutantPairing is plan 10-08 Task 3's own
// D-10-41 mutant pairing, the COMPLEMENTARY half of plan 10-01's own
// move-as-copy test (interp_test.go#TestMoveAsCopyMutationKilled). Naming
// which peer alone catches which mutant, in both directions, is what makes
// "independent" falsifiable rather than rhetorical:
//
//	| Mutant                                                                        | Catches                                                                              | Blind, and why                                                                                          |
//	|--------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------|
//	| interp: a Move treated as a Copy (plan 10-01, interp.go's moveAsCopyForTest)     | interp only (TestMoveAsCopyMutationKilled)                                            | check/corevalidate execute nothing; a Move's own STATIC type is unaffected by whether the RUNTIME treats it as a copy |
//	| corevalidate: parameterContract.Mode hardcoded to the wrong value (this test, corevalidate.go's parameterContractModeOverrideForTest) | the criterion-4 differential (corevalidate's own peer signature vs. originvalidate's declared contract) | interp, which never reads a mode string at all -- it executes core.Operations, never contract metadata    |
//
// This is D-10-40's own escape hatch made concrete: ParameterContract.Mode
// is hardcoded "owned" everywhere in production (D-07-01's cardinality-1
// grammar), so a NATURAL input can never disagree by construction. This
// seeded fault is what makes criterion 4's ownership-fact half non-vacuous.
//
// Reachable only from package corevalidate_test (this file): the seam
// (corevalidate.SetParameterContractModeOverrideForTest, export_test.go)
// is a _test.go-only symbol, invisible to any OTHER package's own test
// binary -- session_peer_gate_test.go's own doc comment cross-references
// this test rather than duplicating it for that reason.
func TestDerivePeerSignatureModeMutantPairing(t *testing.T) {
	fixture := testsupport.ProjectPath("testdata", "phase07", "call_basic.schway")
	checked, err := session.CheckFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("call_basic.schway unexpectedly refused: %+v", checked.Diagnostics)
	}
	var function core.Function
	for _, candidate := range checked.Program.Functions {
		if candidate.Name == "main" {
			function = candidate
		}
	}
	if function.ID == "" {
		t.Fatal("call_basic.schway has no function named main")
	}

	declared, err := originvalidate.BuildInterface(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var declaredMode string
	for _, signature := range declared.Functions {
		if signature.ID == function.ID {
			declaredMode = signature.Parameters[0].Mode
		}
	}
	if declaredMode == "" {
		t.Fatalf("declared contract has no parameter mode for %s", function.ID)
	}

	baseline := corevalidate.Validate(checked.Program)
	if !baseline.Valid {
		t.Fatalf("call_basic.schway unexpectedly corevalidate-rejected: %+v", baseline.Problems)
	}
	baselinePeer, ok := baseline.PeerSignatures()[function.ID]
	if !ok {
		t.Fatalf("peer signatures missing %s", function.ID)
	}
	if baselinePeer.Parameters[0].Mode != declaredMode {
		t.Fatalf("baseline: peer mode %q disagrees with the declared contract %q before any fault -- this test's own fixture assumption is wrong", baselinePeer.Parameters[0].Mode, declaredMode)
	}

	baselineExecution, err := interp.Run(checked.Program, function.Name, "7")
	if err != nil {
		t.Fatalf("interp.Run before the fault: %v", err)
	}
	baselineBytes, err := interp.CanonicalBytes(baselineExecution)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("seeded fault: the differential against the declared contract catches it", func(t *testing.T) {
		restore := corevalidate.SetParameterContractModeOverrideForTest(func() string { return "shared" })
		defer restore()

		mutated := corevalidate.Validate(checked.Program)
		if !mutated.Valid {
			t.Fatalf("mutated: corevalidate unexpectedly rejected the program outright: %+v", mutated.Problems)
		}
		mutatedPeer, ok := mutated.PeerSignatures()[function.ID]
		if !ok {
			t.Fatal("mutated: peer signatures missing the function")
		}
		if mutatedPeer.Parameters[0].Mode == declaredMode {
			t.Fatal("seeded fault did not change the peer's reported parameter mode -- the differential has nothing to catch")
		}
	})

	t.Run("interp stays blind: canonical bytes unchanged under the same fault", func(t *testing.T) {
		restore := corevalidate.SetParameterContractModeOverrideForTest(func() string { return "shared" })
		defer restore()

		mutatedExecution, err := interp.Run(checked.Program, function.Name, "7")
		if err != nil {
			t.Fatalf("interp.Run under the fault: %v", err)
		}
		mutatedBytes, err := interp.CanonicalBytes(mutatedExecution)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(baselineBytes, mutatedBytes) {
			t.Fatal("interp's canonical bytes changed under a fault it never consults (a mode string) -- interp should be structurally blind to this")
		}
	})
}

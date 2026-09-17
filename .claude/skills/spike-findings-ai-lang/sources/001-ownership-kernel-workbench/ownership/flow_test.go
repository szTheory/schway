package ownership

import "testing"

func TestFlowFixturesMatchExpectations(t *testing.T) {
	file, err := LoadFlowFixtures("../fixtures/flow-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range file.Fixtures {
		fixture := fixture
		t.Run(fixture.Program.ID, func(t *testing.T) {
			comparison := CompareFlow(fixture.Program, FlowOptions{})
			if !comparison.Agreement || comparison.Analyzer.Valid != fixture.ExpectValid {
				t.Fatalf("unexpected comparison: %+v", comparison)
			}
			if !fixture.ExpectValid && comparison.Analyzer.Diagnostic.Code != fixture.ExpectCode {
				t.Fatalf("diagnostic=%q, want %q", comparison.Analyzer.Diagnostic.Code, fixture.ExpectCode)
			}
		})
	}
}

func op(operation Operation) FlowStep {
	return FlowStep{Operation: &operation}
}

func branch(thenSteps, elseSteps []FlowStep) FlowStep {
	return FlowStep{Branch: &FlowBranch{ID: "choice", Then: thenSteps, Else: elseSteps}}
}

func loop(body []FlowStep) FlowStep {
	return FlowStep{Loop: &FlowLoop{ID: "repeat", Body: body}}
}

func TestBranchJoinRejectsUseAfterConditionalMove(t *testing.T) {
	program := FlowProgram{ID: "FLOW-MOVE-001", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		branch([]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})}, nil),
		op(Operation{Kind: ReadPlace, Place: "p"}),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
	if comparison.Analyzer.Diagnostic.Code != "ownership.maybe_uninitialized" {
		t.Fatalf("diagnostic=%q, want ownership.maybe_uninitialized", comparison.Analyzer.Diagnostic.Code)
	}
}

func TestBranchJoinAcceptsTargetMovedOnEveryPath(t *testing.T) {
	program := FlowProgram{ID: "FLOW-MOVE-002", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		branch(
			[]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})},
			[]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})},
		),
		op(Operation{Kind: ReadPlace, Place: "q"}),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || !comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}

func TestBranchJoinAllowsDeadConditionalMove(t *testing.T) {
	program := FlowProgram{ID: "FLOW-MOVE-003", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		branch([]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})}, nil),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || !comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}

func TestBranchJoinRejectsMaybeInitializedTarget(t *testing.T) {
	program := FlowProgram{ID: "FLOW-MOVE-004", Steps: []FlowStep{
		branch([]FlowStep{op(Operation{Kind: Declare, Place: "q"})}, nil),
		op(Operation{Kind: Declare, Place: "p"}),
		op(Operation{Kind: Move, Place: "p", Target: "q"}),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}

func TestNestedBranchesAgreeWithPathExpansion(t *testing.T) {
	program := FlowProgram{ID: "FLOW-NESTED-001", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		branch(
			[]FlowStep{branch(
				[]FlowStep{op(Operation{Kind: ReadPlace, Place: "p"})},
				[]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})},
			)},
			[]FlowStep{op(Operation{Kind: ReadPlace, Place: "p"})},
		),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || !comparison.Analyzer.Valid || len(comparison.Paths) != 3 {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}

func TestFlowFaultInjectionDisagreesWithPathOracle(t *testing.T) {
	program := FlowProgram{ID: "FLOW-INJECT-001", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		branch([]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})}, nil),
		op(Operation{Kind: ReadPlace, Place: "p"}),
	}}
	comparison := CompareFlow(program, FlowOptions{AllowMaybeUse: true})
	if comparison.Agreement || !comparison.Analyzer.Valid {
		t.Fatalf("fault injection was not detected: %+v", comparison)
	}
}

func TestSmallBranchProductAgreesWithPathExpansion(t *testing.T) {
	choices := []*Operation{
		nil,
		{Kind: ReadPlace, Place: "p"},
		{Kind: Move, Place: "p", Target: "q"},
		{Kind: Release, Place: "p"},
	}
	checked := 0
	for _, thenChoice := range choices {
		for _, elseChoice := range choices {
			for _, suffixChoice := range choices {
				thenSteps := []FlowStep{}
				if thenChoice != nil {
					thenSteps = append(thenSteps, op(*thenChoice))
				}
				elseSteps := []FlowStep{}
				if elseChoice != nil {
					elseSteps = append(elseSteps, op(*elseChoice))
				}
				steps := []FlowStep{op(Operation{Kind: Declare, Place: "p"}), branch(thenSteps, elseSteps)}
				if suffixChoice != nil {
					steps = append(steps, op(*suffixChoice))
				}
				comparison := CompareFlow(FlowProgram{ID: "FLOW-GENERATED", Steps: steps}, FlowOptions{})
				checked++
				if !comparison.Agreement {
					t.Fatalf("mismatch after %d cases: %+v", checked, comparison)
				}
			}
		}
	}
	if checked != 64 {
		t.Fatalf("checked=%d, want 64", checked)
	}
}

func TestReadOnlyLoopReachesStableOwnershipState(t *testing.T) {
	program := FlowProgram{ID: "FLOW-LOOP-001", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		loop([]FlowStep{op(Operation{Kind: ReadPlace, Place: "p"})}),
		op(Operation{Kind: ReadPlace, Place: "p"}),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || !comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}

func TestRepeatedMoveLoopIsRejected(t *testing.T) {
	program := FlowProgram{ID: "FLOW-LOOP-002", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		loop([]FlowStep{op(Operation{Kind: Move, Place: "p", Target: "q"})}),
	}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
}

func TestAmbiguousFlowStepIsRejectedByAnalyzerAndPaths(t *testing.T) {
	operation := Operation{Kind: Declare, Place: "p"}
	program := FlowProgram{ID: "FLOW-MALFORMED-001", Steps: []FlowStep{{
		Operation: &operation,
		Branch:    &FlowBranch{ID: "also-a-branch"},
	}}}
	comparison := CompareFlow(program, FlowOptions{})
	if !comparison.Agreement || comparison.Analyzer.Valid {
		t.Fatalf("unexpected comparison: %+v", comparison)
	}
	if comparison.Analyzer.Diagnostic.Code != "ownership.invalid_flow_step" {
		t.Fatalf("diagnostic=%q, want ownership.invalid_flow_step", comparison.Analyzer.Diagnostic.Code)
	}
}

func TestFlowLoanJoinMismatchedOwnerIsRejectedConservatively(t *testing.T) {
	program := FlowProgram{ID: "FLOW-LOAN-MISMATCH-001", Steps: []FlowStep{
		op(Operation{Kind: Declare, Place: "p"}),
		op(Operation{Kind: Declare, Place: "q"}),
		branch(
			[]FlowStep{op(Operation{Kind: BorrowShared, Place: "p", Loan: "view"})},
			[]FlowStep{op(Operation{Kind: BorrowShared, Place: "q", Loan: "view"})},
		),
	}}
	result := AnalyzeFlow(program, FlowOptions{})
	if result.Valid || result.Diagnostic.Code != "ownership.loan_join_mismatch" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

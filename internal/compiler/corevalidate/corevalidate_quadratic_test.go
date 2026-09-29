package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// TestValidatorVerdictsUnchanged is 03-04-02's own pin: before
// loanChainIndex replaced replayStraightLine/replayBlocks' per-operation
// loansForPlace copy-and-rescan (D-02-03/Q2(b)), every one of these fixtures
// produced the exact accept/reject verdict and first-problem code recorded
// here. The replacement changes ONLY how loanLastUse is computed, never the
// law it feeds — a regression in the replacement would show up as a changed
// verdict or a changed first code, not as a changed cost.
func TestValidatorVerdictsUnchanged(t *testing.T) {
	tests := []struct {
		name      string
		program   core.Program
		wantValid bool
		wantCode  string
	}{
		{name: "owned move-and-return", program: ownedProgram(), wantValid: true},
		{name: "shared borrow, copy, then move", program: borrowedProgram(), wantValid: true},
		{
			name: "reborrow observed after move is rejected",
			program: reborrowProgram(core.OpBorrowShared, true),
			wantValid: false, wantCode: "core.move_while_borrowed",
		},
		{name: "reborrow with no later use is accepted", program: reborrowProgram(core.OpBorrowShared, false), wantValid: true},
		{
			name: "copy of a loan observed after move is rejected",
			program: reborrowProgram(core.OpCopy, true),
			wantValid: false, wantCode: "core.move_while_borrowed",
		},
		{name: "copy of a loan with no later use is accepted", program: reborrowProgram(core.OpCopy, false), wantValid: true},
		{name: "branch-shaped function with no loans", program: validBranchProgram(t), wantValid: true},
		{name: "branch-shaped function with an edge-specific loan", program: validBranchLoanProgram(t), wantValid: true},
		{name: "exclusive-borrow function", program: validExclusiveBorrowProgram(t), wantValid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := corevalidate.Validate(test.program)
			if result.Valid != test.wantValid {
				t.Fatalf("valid=%v want=%v problems=%+v", result.Valid, test.wantValid, result.Problems)
			}
			if !test.wantValid {
				if len(result.Problems) == 0 || result.Problems[0].Code != test.wantCode {
					t.Fatalf("code=%v want=%s", result.Problems, test.wantCode)
				}
			}
		})
	}
}

// scaleReborrowProgram is scaleProgram's reborrow-chain counterpart:
// facts-1 sequential reborrows of the previous loan (`let review_i = borrow
// review_{i-1}`), each formally still live at the final return (the return
// reads the LAST reborrow, so the fixpoint over loanLastUse must walk the
// whole ancestry chain to place every loan's last use correctly) — the
// shape D-02-03's original quadratic blowup actually requires (a series
// built only from independent, non-chained loans never reaches it).
func scaleReborrowProgram(facts int) core.Program {
	if facts < 2 {
		panic("scale facts must include the owner and at least one reborrow")
	}
	functionID := "s1:scale:fn:reborrow"
	typeID := functionID + ":type:0"
	ownerID := functionID + ":place:0"
	linear := &core.LinearBody{
		ID: functionID + ":linear",
		Types: []core.TypeFact{{
			ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
			Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			NegativeWitnesses: []core.AbilityWitness{},
		}},
		Places:     make([]core.Place, 0, facts),
		Operations: make([]core.LinearOperation, 0, facts),
	}
	linear.Places = append(linear.Places, core.Place{ID: ownerID, Name: "owner", TypeID: typeID})
	previous := ownerID
	for index := 0; index < facts-1; index++ {
		placeID := functionID + ":place:" + itoaFacts(index+1)
		linear.Places = append(linear.Places, core.Place{ID: placeID, Name: "review", TypeID: typeID})
		linear.Operations = append(linear.Operations, core.LinearOperation{
			ID: functionID + ":op:" + itoaFacts(index), PointID: functionID + ":point:linear:" + itoaFacts(index),
			Kind: core.OpBorrowShared, SourceID: previous, TargetID: placeID, LoanID: functionID + ":loan:" + itoaFacts(index), TypeID: typeID,
		})
		previous = placeID
	}
	last := facts - 1
	linear.Operations = append(linear.Operations, core.LinearOperation{
		ID: functionID + ":op:" + itoaFacts(last), PointID: functionID + ":point:linear:" + itoaFacts(last),
		Kind: core.OpReturn, SourceID: previous, TypeID: typeID,
	})
	return core.Program{
		Schema: core.Schema1, Module: "scale", ModuleID: "s1:scale:module:reborrow",
		Functions: []core.Function{{
			ID: functionID, Name: "reborrow", EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter: core.Parameter{ID: ownerID, Name: "owner", Type: "Byte"}, ReturnType: "Byte", Linear: linear,
		}},
	}
}

func itoaFacts(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// TestValidatorReborrowChainIsLinear is 03-04-02's honest-work falsifier:
// the validator's counted Checks grows LINEARLY (asserted as a growth
// ratio, not an exact formula or wall-clock time) across a long reborrow
// chain of 101, 1,001, and 10,001 facts. A reintroduction of the
// per-operation loansForPlace copy-and-rescan this task removes would show
// up here as super-linear growth even though TestCoreValidationWorkSeries
// (built from independent, non-chained loans) would stay unaffected — this
// is the D-10 reachability question applied to a cost test: a series with
// no chain in it cannot exercise the shape that produced the quadratic cost
// in the first place.
func TestValidatorReborrowChainIsLinear(t *testing.T) {
	samples := []int{101, 1001, 10001}
	checks := make([]int, len(samples))
	for index, facts := range samples {
		program := scaleReborrowProgram(facts)
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("facts=%d rejected: %+v", facts, result)
		}
		checks[index] = result.Checks
	}
	// A linear series scales work/facts to within a small, fixed band
	// across a 100x growth in input size; a quadratic regression would
	// blow this ratio up by roughly the same 100x factor.
	baseline := float64(checks[0]) / float64(samples[0])
	for index, facts := range samples {
		ratio := float64(checks[index]) / float64(facts)
		if ratio > baseline*3 {
			t.Fatalf("facts=%d checks=%d ratio=%.2f exceeds linear band (baseline ratio=%.2f) -- growth looks super-linear", facts, checks[index], ratio, baseline)
		}
	}
}

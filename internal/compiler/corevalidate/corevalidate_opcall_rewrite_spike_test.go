package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// findFirstOpCallForTest returns the (functionIndex, operationIndex) of the
// FIRST core.OpCall operation in program, walking functions in declaration
// order and, within a function, operations in declaration order. Fails the
// test if no OpCall exists -- Q-01 requires a real call site to rewrite.
func findFirstOpCallForTest(t *testing.T, program core.Program) (functionIndex, operationIndex int) {
	t.Helper()
	for fi, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for oi, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall {
				return fi, oi
			}
		}
	}
	t.Fatal("expected testdata/phase07/deep_diamond_acyclic.schway to contain at least one core.OpCall operation")
	return 0, 0
}

// TestQ01CoreLevelOpCallToOpCopyRewrite is Phase 11's Q-01 pre-planning
// experiment (11-CONTEXT.md, D-11-29): does corevalidate accept a core-level
// rewrite of a single OpCall operation to OpCopy, with ID/SourceID/TargetID
// preserved byte-for-byte and CalleeID cleared? This decides whether QLT-05
// (plan 11-08's reducer) can express "drop a call site" as a single-field
// core-level edit, or whether corevalidate's independent re-derivation
// refuses the rewritten shape -- narrowing 11-08 to a RefusedShapes-gated
// scope instead.
//
// The test asserts EXACTLY ONE of the two possible outcomes below, never a
// disjunction -- an either/or assertion would be satisfied whichever way
// corevalidate answers, which is the vacuity this spike exists to avoid.
func TestQ01CoreLevelOpCallToOpCopyRewrite(t *testing.T) {
	program := loadCheckedProgram(t, "phase07", "deep_diamond_acyclic.schway")

	fi, oi := findFirstOpCallForTest(t, program)
	original := program.Functions[fi].Linear.Operations[oi]
	if original.Kind != core.OpCall {
		t.Fatalf("expected the located operation to be core.OpCall, got %v", original.Kind)
	}

	rewritten := cloneCheckedProgramForTest(t, program)
	target := &rewritten.Functions[fi].Linear.Operations[oi]
	if target.ID != original.ID || target.SourceID != original.SourceID || target.TargetID != original.TargetID {
		t.Fatal("clone diverged from original before the rewrite was even applied")
	}
	target.Kind = core.OpCopy
	target.CalleeID = ""
	// ID, SourceID and TargetID are untouched -- this is a single-field
	// edit (Kind) plus clearing CalleeID, nothing else.

	result := corevalidate.Validate(rewritten)

	// Q-01 BRANCH A (accepted): corevalidate's independent re-derivation
	// (peerDeriveOriginFacts has no core.OpCall case -- the Phase 10
	// carry-forward gap this experiment probes) never observes the callee
	// edge at all once it has been rewritten to OpCopy, so it raises no
	// problem against this fixture. This is a single-outcome assertion,
	// not a disjunction: if corevalidate's answer ever flips to refusing
	// the rewrite, this assertion goes red rather than silently continuing
	// to claim BRANCH A, and PHASE-11-DEBT.md's recorded verdict must be
	// revisited alongside plan 11-08's branch selection.
	if !result.Valid {
		t.Fatalf("Q-01 verdict changed: corevalidate now REFUSES the core-level OpCall to OpCopy rewrite (problems: %+v) -- this test asserted BRANCH A (accepted) at authoring time; update PHASE-11-DEBT.md and re-evaluate plan 11-08's branch selection", result.Problems)
	}
	t.Log("Q-01 verdict: BRANCH A (accepted) -- corevalidate accepted the core-level OpCall to OpCopy rewrite")
}

// TestQ01RewrittenProgramStillChecks proves Task 1's rewrite is a
// single-field edit: the function set, operation count, and every
// operation's ID are unchanged between the pre-rewrite and rewritten
// program. This is what makes Q-01's verdict attributable to the Kind
// change alone, never to collateral damage from the rewrite mechanism
// itself.
func TestQ01RewrittenProgramStillChecks(t *testing.T) {
	program := loadCheckedProgram(t, "phase07", "deep_diamond_acyclic.schway")
	fi, oi := findFirstOpCallForTest(t, program)

	rewritten := cloneCheckedProgramForTest(t, program)
	rewritten.Functions[fi].Linear.Operations[oi].Kind = core.OpCopy
	rewritten.Functions[fi].Linear.Operations[oi].CalleeID = ""

	if len(rewritten.Functions) != len(program.Functions) {
		t.Fatalf("function count changed: before=%d after=%d", len(program.Functions), len(rewritten.Functions))
	}
	for i := range program.Functions {
		before := program.Functions[i]
		after := rewritten.Functions[i]
		if before.ID != after.ID {
			t.Fatalf("function %d ID changed: before=%q after=%q", i, before.ID, after.ID)
		}
		if before.Linear == nil || after.Linear == nil {
			if before.Linear != after.Linear {
				t.Fatalf("function %d Linear presence changed", i)
			}
			continue
		}
		if len(before.Linear.Operations) != len(after.Linear.Operations) {
			t.Fatalf("function %d operation count changed: before=%d after=%d", i, len(before.Linear.Operations), len(after.Linear.Operations))
		}
		for j := range before.Linear.Operations {
			beforeOp := before.Linear.Operations[j]
			afterOp := after.Linear.Operations[j]
			if beforeOp.ID != afterOp.ID {
				t.Fatalf("function %d operation %d ID changed: before=%q after=%q", i, j, beforeOp.ID, afterOp.ID)
			}
		}
	}
}

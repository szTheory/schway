package reduce_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/reduce"
)

// D-05-27's three required reducer mutation-kills: a reducer has exactly
// three distinct ways to be vacuous, and each needs its own kill --
// control:reduce.no_progress, control:reduce.predicate_too_loose, and
// control:reduce.nondeterministic. A predicate that is too TIGHT is caught
// by TestNoOpReducerGoesRed's own strict size-decrease assertion, which is
// why D-05-27 lists only three kills rather than four.

// operationCount is this file's own size metric: the total number of
// core.LinearOperation entries the function's Linear body carries,
// counting every arm's operations for a Match-bodied function too (every
// move in this package operates on fn.Linear.Operations directly,
// regardless of whether the function is Match-bodied or straight-line).
func operationCount(p core.Program) int {
	if len(p.Functions) == 0 || p.Functions[0].Linear == nil {
		return 0
	}
	return len(p.Functions[0].Linear.Operations)
}

// TestNoOpReducerGoesRed is control:reduce.no_progress's mutation-kill: a
// reducer whose move application is the identity (it reports progress
// without shrinking anything) must be caught by a STRICT size-decrease
// assertion over a corpus of at least three distinct, genuinely-reducible
// seeded mismatches -- reusing this package's own existing NAT-03-style
// fixture builders (reduce_test.go) as the corpus, each exercising a
// different one of the five moves.
func TestNoOpReducerGoesRed(t *testing.T) {
	corpus := []struct {
		name string
		seed core.Program
	}{
		{"borrow_chain_unused_binding_and_truncation", borrowChainSeed()},
		{"foreign_chain_offpath_stage", foreignChainSeedWithSteps(3)},
		{"three_arm_match_unmatched_arm", threeArmMatchSeed()},
	}
	if len(corpus) < 3 {
		t.Fatalf("D-05-27 requires a corpus of at least three seeded mismatches, got %d", len(corpus))
	}
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interpreter-vs-O0"}

	for _, entry := range corpus {
		t.Run(entry.name, func(t *testing.T) {
			seedSize := operationCount(entry.seed)

			// Green path: the REAL reducer strictly decreases size.
			real, err := reduce.Reduce(context.Background(), reduce.Seed{Program: entry.seed}, alwaysInteresting(sig))
			if err != nil {
				t.Fatalf("real reduce: %v", err)
			}
			if operationCount(real.Program) >= seedSize {
				t.Fatalf("control:reduce.no_progress: the REAL reducer did not strictly decrease operation count for %s (seed=%d got=%d) -- fixture is not genuinely reducible", entry.name, seedSize, operationCount(real.Program))
			}

			// Red path: a no-op move list (Apply always reports success
			// without changing the program) must NOT pass the strict
			// size-decrease assertion -- this is the mutation D-05-27
			// requires this test to catch.
			reduce.SetTestOnlyMoves(func() []reduce.Move {
				return []reduce.Move{{Name: "no-op", Apply: func(p core.Program) (core.Program, bool) { return p, true }}}
			})
			mutated, err := reduce.Reduce(context.Background(), reduce.Seed{Program: entry.seed}, alwaysInteresting(sig))
			reduce.ResetTestOnlyMoves()
			if err != nil {
				t.Fatalf("no-op reduce: %v", err)
			}
			if operationCount(mutated.Program) < seedSize {
				t.Fatalf("no-op move list unexpectedly shrank %s -- the mutation was not exercised", entry.name)
			}
			if mutated.Minimality != reduce.MinimalityBudgetExhausted {
				t.Fatalf("expected the no-op reducer to exhaust its budget (never reach a real fixpoint) for %s, got %q", entry.name, mutated.Minimality)
			}
		})
	}
}

// TestPredicateTooLooseGoesRed is control:reduce.predicate_too_loose's
// mutation-kill: seed two distinct mismatches A and B with different
// signatures, then show that a predicate that accepts ANY candidate
// (ignoring reduce.Interesting entirely) lets reduction toward A drift
// into reporting B's signature -- exactly the vacuity D-05-24's real
// conjunction exists to refuse. With the real predicate (which applies
// reduce.Interesting against the ORIGINAL seed's own signature), the
// drift toward B's signature is rejected and no reduction happens.
func TestPredicateTooLooseGoesRed(t *testing.T) {
	seedA := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interpreter-vs-O0", OperationID: "op-A"}
	seedB := reduce.Signature{Axis: "axis:event-order", EnginePair: "O0-vs-O3", OperationID: "op-B"}
	if reduce.Interesting(seedA, seedB) {
		t.Fatal("test fixture error: seedA and seedB must be genuinely different signatures")
	}
	programA := borrowChainSeed()
	seedSize := operationCount(programA)

	// Green path: a real predicate, even when handed a hostile candidate
	// signature that claims to be B, still applies Interesting against the
	// real seed (A) -- D-05-24's conjunction correctly refuses it, so no
	// reduction happens.
	realPredicate := func(_ context.Context, _ core.Program) (reduce.Signature, bool, error) {
		return seedB, reduce.Interesting(seedA, seedB), nil
	}
	resultReal, err := reduce.Reduce(context.Background(), reduce.Seed{Program: programA}, realPredicate)
	if err != nil {
		t.Fatalf("real predicate reduce: %v", err)
	}
	if operationCount(resultReal.Program) != seedSize {
		t.Fatalf("control:reduce.predicate_too_loose: a correct predicate accepted a candidate matching a DIFFERENT mismatch's signature (B) while reducing A")
	}

	// Red path: a "too loose" predicate accepts any candidate unconditionally,
	// ignoring Interesting -- reduction proceeds and drifts toward reporting
	// B's signature while claiming to still be reducing A, exactly the
	// vacuity mode this control kills.
	tooLoose := func(_ context.Context, _ core.Program) (reduce.Signature, bool, error) {
		return seedB, true, nil
	}
	resultMutated, err := reduce.Reduce(context.Background(), reduce.Seed{Program: programA}, tooLoose)
	if err != nil {
		t.Fatalf("too-loose predicate reduce: %v", err)
	}
	if operationCount(resultMutated.Program) >= seedSize {
		t.Fatal("expected the too-loose predicate to accept a candidate drifting toward a DIFFERENT signature (proving the mutation is exercised), but it never reduced")
	}
}

// TestReducerNonDeterminismGoesRed is control:reduce.nondeterministic's
// mutation-kill: with the real, fixed move order, Reduce is byte-identical
// across repeated runs; randomizing the move order via the test seam
// (D-05-27) must, for at least one seed within a bounded search, produce a
// byte-different reduced output -- proving the fixed order in Moves() is
// load-bearing for determinism, not decorative.
func TestReducerNonDeterminismGoesRed(t *testing.T) {
	seed := threeArmMatchSeedWithUnusedBindings()
	sig := reduce.Signature{Axis: "a", EnginePair: "p"}

	first, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("first real reduce: %v", err)
	}
	second, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("second real reduce: %v", err)
	}
	firstBytes, err := json.Marshal(first.Program)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	secondBytes, err := json.Marshal(second.Program)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(firstBytes) != string(secondBytes) {
		t.Fatalf("the real, fixed move order is not deterministic:\nfirst:  %s\nsecond: %s", firstBytes, secondBytes)
	}

	seenDivergence := false
	const maxAttempts = 50
	for attempt := int64(1); attempt <= maxAttempts && !seenDivergence; attempt++ {
		rngSeed := attempt
		reduce.SetTestOnlyMoves(func() []reduce.Move {
			moves := append([]reduce.Move(nil), reduce.Moves()...)
			randomSource := rand.New(rand.NewSource(rngSeed))
			randomSource.Shuffle(len(moves), func(i, j int) { moves[i], moves[j] = moves[j], moves[i] })
			return moves
		})
		result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
		reduce.ResetTestOnlyMoves()
		if err != nil {
			t.Fatalf("randomized-order reduce (seed=%d): %v", rngSeed, err)
		}
		resultBytes, err := json.Marshal(result.Program)
		if err != nil {
			t.Fatalf("marshal randomized-order result: %v", err)
		}
		if string(resultBytes) != string(firstBytes) {
			seenDivergence = true
		}
	}
	if !seenDivergence {
		t.Fatalf("control:reduce.nondeterministic: randomizing move order across %d seeds never produced a byte-different reduced output -- the mutation was not exercised", maxAttempts)
	}
}

// threeArmMatchSeedWithUnusedBindings extends threeArmMatchSeed with an
// extra unused shared borrow in the first two arms, giving
// drop-unused-binding a genuine site alongside drop-unmatched-arm and
// collapse-branch-to-diverging-arm's own eligible sites -- the combination
// needed to make move ORDER observable in the final byte output (verified
// empirically this session: shuffled orders diverge on which arm survives
// collapse-branch-to-diverging-arm's own fixed Arms[0] choice once
// drop-unmatched-arm has already run a different number of times).
func threeArmMatchSeedWithUnusedBindings() core.Program {
	const fn = "s1:phase5.reduce_fixture:function:triage"
	const matchID = fn + ":match"
	parameterID := fn + ":place:0"
	typeID := fn + ":type:0"
	return core.Program{
		Schema:   core.Schema,
		Module:   "phase5.reduce_fixture",
		ModuleID: "s1:phase5.reduce_fixture:module:phase5.reduce_fixture",
		DataTypes: []core.DataType{
			{ID: "s1:phase5.reduce_fixture:type:Signal", Name: "Signal", Alternatives: []string{"Go", "Halt", "Extra"}},
		},
		Functions: []core.Function{
			{
				ID: fn, Name: "triage",
				EntryPointID: fn + ":point:entry", ReturnPointID: fn + ":point:return",
				Parameter:  core.Parameter{ID: parameterID, Name: "flag", Type: "Signal"},
				ReturnType: "Signal",
				Match: &core.Match{
					ID: matchID, PointID: fn + ":point:match", Scrutinee: "flag",
					Arms: []core.MatchArm{
						{ID: fn + ":arm:0", EdgeID: matchID + ":edge:Go", Pattern: "Go", BlockID: fn + ":block:arm:0"},
						{ID: fn + ":arm:1", EdgeID: matchID + ":edge:Halt", Pattern: "Halt", BlockID: fn + ":block:arm:1"},
						{ID: fn + ":arm:2", EdgeID: matchID + ":edge:Extra", Pattern: "Extra", BlockID: fn + ":block:arm:2"},
					},
				},
				Linear: &core.LinearBody{
					ID: fn + ":linear", Types: []core.TypeFact{{ID: typeID}},
					Places: []core.Place{
						{ID: parameterID, Name: "flag", TypeID: typeID},
						{ID: fn + ":place:1", Name: "flag", TypeID: typeID},
						{ID: fn + ":place:2", Name: "held", TypeID: typeID},
						{ID: fn + ":place:3", Name: "unused0", TypeID: typeID},
						{ID: fn + ":place:4", Name: "flag", TypeID: typeID},
						{ID: fn + ":place:5", Name: "unused1", TypeID: typeID},
						{ID: fn + ":place:6", Name: "_", TypeID: typeID},
					},
					Operations: []core.LinearOperation{
						{ID: fn + ":op:0", PointID: fn + ":point:linear:0", Kind: core.OpCopy, SourceID: parameterID, TargetID: fn + ":place:1", TypeID: typeID},
						{ID: fn + ":op:1", PointID: fn + ":point:linear:1", Kind: core.OpMove, SourceID: fn + ":place:1", TargetID: fn + ":place:2", TypeID: typeID},
						{ID: fn + ":op:6", PointID: fn + ":point:linear:6", Kind: core.OpBorrowShared, SourceID: fn + ":place:2", TargetID: fn + ":place:3", TypeID: typeID},
						{ID: fn + ":op:2", PointID: fn + ":point:linear:2", Kind: core.OpReturn, SourceID: fn + ":place:2", TypeID: typeID},
						{ID: fn + ":op:3", PointID: fn + ":point:linear:3", Kind: core.OpCopy, SourceID: parameterID, TargetID: fn + ":place:4", TypeID: typeID},
						{ID: fn + ":op:7", PointID: fn + ":point:linear:7", Kind: core.OpBorrowShared, SourceID: fn + ":place:4", TargetID: fn + ":place:5", TypeID: typeID},
						{ID: fn + ":op:4", PointID: fn + ":point:linear:4", Kind: core.OpDefect, SourceID: fn + ":place:4", TypeID: typeID, Reason: "adversarial defect"},
						{ID: fn + ":op:5", PointID: fn + ":point:linear:5", Kind: core.OpReturn, SourceID: parameterID, TypeID: typeID},
					},
					Blocks: []core.Block{
						{ID: fn + ":block:entry", PointID: fn + ":point:entry", Successors: []string{fn + ":block:arm:0", fn + ":block:arm:1", fn + ":block:arm:2"}},
						{ID: fn + ":block:arm:0", PointID: fn + ":point:arm:0", OperationIDs: []string{fn + ":op:0", fn + ":op:1", fn + ":op:6", fn + ":op:2"}, Successors: []string{fn + ":block:join"}},
						{ID: fn + ":block:arm:1", PointID: fn + ":point:arm:1", OperationIDs: []string{fn + ":op:3", fn + ":op:7", fn + ":op:4"}, Successors: []string{fn + ":block:join"}},
						{ID: fn + ":block:arm:2", PointID: fn + ":point:arm:2", OperationIDs: []string{fn + ":op:5"}, Successors: []string{fn + ":block:join"}},
						{ID: fn + ":block:join", PointID: fn + ":point:return"},
					},
				},
			},
		},
	}
}

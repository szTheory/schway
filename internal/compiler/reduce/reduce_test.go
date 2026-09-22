package reduce_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/reduce"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// ---------------------------------------------------------------------
// Fixture builders -- small, hand-built core.Program values. Reduce's
// moves operate on core.Program structure directly (D-05-23), so these
// fixtures do not go through the parser/checker; only ProjectSource's own
// round-trip test (reduce_project_test.go) drives the real front end.
// ---------------------------------------------------------------------

// borrowChainSeed mirrors testdata/phase5/restrict_borrow.lang's shape: a
// straight-line chain of borrows over one parameter, with an UNUSED
// trailing borrow inserted so TestReduceHasExactlyFiveMoves' sibling tests
// have something for drop-unused-binding to remove.
func borrowChainSeed() core.Program {
	const fn = "s1:phase5.reduce_fixture:function:touch"
	return core.Program{
		Schema:   core.Schema,
		Module:   "phase5.reduce_fixture",
		ModuleID: "s1:phase5.reduce_fixture:module:phase5.reduce_fixture",
		Functions: []core.Function{
			{
				ID: fn, Name: "touch",
				EntryPointID: fn + ":point:entry", ReturnPointID: fn + ":point:return",
				Parameter:  core.Parameter{ID: fn + ":place:0", Name: "buffer", Type: "Buffer"},
				ReturnType: "Buffer",
				Linear: &core.LinearBody{
					ID:     fn + ":linear",
					Types:  []core.TypeFact{{ID: fn + ":type:0"}},
					Places: []core.Place{{ID: fn + ":place:0", Name: "buffer", TypeID: fn + ":type:0"}},
					Operations: []core.LinearOperation{
						{ID: fn + ":op:0", PointID: fn + ":point:linear:0", Kind: core.OpBorrowExclusive, SourceID: fn + ":place:0", TargetID: fn + ":place:1", TypeID: fn + ":type:0"},
						{ID: fn + ":op:1", PointID: fn + ":point:linear:1", Kind: core.OpBorrowShared, SourceID: fn + ":place:1", TargetID: fn + ":place:2", TypeID: fn + ":type:0"},
						// unused: nothing reads place:3
						{ID: fn + ":op:2", PointID: fn + ":point:linear:2", Kind: core.OpBorrowShared, SourceID: fn + ":place:1", TargetID: fn + ":place:3", TypeID: fn + ":type:0"},
						{ID: fn + ":op:3", PointID: fn + ":point:linear:3", Kind: core.OpReturn, SourceID: fn + ":place:2", TypeID: fn + ":type:0"},
					},
				},
			},
		},
	}
}

// foreignChainSeed mirrors testdata/phase5/tail_collapse_release_ladder.lang's
// shape: three successive foreign acquisitions to the same declared symbol,
// all released in reverse order, function returns its own parameter.
func foreignChainSeed() core.Program {
	return foreignChainSeedWithSteps(3)
}

func foreignChainSeedWithSteps(steps int) core.Program {
	const fn = "s1:phase5.reduce_fixture:function:main"
	const module = "phase5.reduce_fixture"
	parameterID := fn + ":place:0"
	var places []core.Place
	places = append(places, core.Place{ID: parameterID, Name: "request", TypeID: fn + ":type:0"})
	var ops []core.LinearOperation
	callIDs := make([]string, steps)
	for i := 0; i < steps; i++ {
		okPlace := fn + ":place:" + itoa(i+1)
		errPlace := fn + ":place:" + itoa(steps+1+i)
		callID := fn + ":op:" + itoa(i)
		callIDs[i] = callID
		places = append(places, core.Place{ID: okPlace, Name: "v" + itoa(i), TypeID: fn + ":type:0"})
		ops = append(ops, core.LinearOperation{
			ID: callID, PointID: fn + ":point:linear:" + itoa(i), Kind: core.OpForeignCall,
			SourceID: parameterID, TargetID: okPlace, ErrTargetID: errPlace, TypeID: fn + ":type:0",
			OkEdgeID: fn + ":edge:step:" + itoa(i) + ":ok", ErrEdgeID: fn + ":edge:step:" + itoa(i) + ":err",
			Allocator: "libc_malloc",
		})
	}
	// releases in reverse completion order
	nextIdx := steps
	for i := steps - 1; i >= 0; i-- {
		ops = append(ops, core.LinearOperation{
			ID: fn + ":op:" + itoa(nextIdx), PointID: fn + ":point:linear:" + itoa(nextIdx),
			Kind: core.OpRelease, SourceID: callIDsTarget(places, i), TypeID: fn + ":type:0",
			ReleasesOperationID: callIDs[i], Allocator: "libc_malloc",
		})
		nextIdx++
	}
	ops = append(ops, core.LinearOperation{
		ID: fn + ":op:" + itoa(nextIdx), PointID: fn + ":point:linear:" + itoa(nextIdx),
		Kind: core.OpReturn, SourceID: parameterID, TypeID: fn + ":type:0",
	})

	successBlock := core.Block{ID: fn + ":block:success", PointID: fn + ":point:success"}
	for _, op := range ops {
		successBlock.OperationIDs = append(successBlock.OperationIDs, op.ID)
	}

	return core.Program{
		Schema:   core.Schema,
		Module:   module,
		ModuleID: "s1:" + module + ":module:" + module,
		DataTypes: []core.DataType{
			{ID: "s1:" + module + ":type:AcquireError", Name: "AcquireError", Alternatives: []string{"OpenFailed"}},
		},
		Functions: []core.Function{
			{
				ID: fn, Name: "main",
				EntryPointID: fn + ":point:entry", ReturnPointID: fn + ":point:return",
				Parameter:       core.Parameter{ID: parameterID, Name: "request", Type: "Byte"},
				ReturnType:      "Byte",
				ForeignContract: &core.ForeignContract{Symbol: "lang_res_open", Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden", Fails: "AcquireError"},
				Linear: &core.LinearBody{
					ID: fn + ":linear", Types: []core.TypeFact{{ID: fn + ":type:0"}},
					Places: places, Operations: ops,
					Blocks: []core.Block{successBlock},
				},
			},
		},
	}
}

func callIDsTarget(places []core.Place, index int) string {
	// v{index} place is at slice offset index+1 (place:0 is the parameter)
	return places[index+1].ID
}

func itoa(n int) string {
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

// alwaysInteresting is a fixed-signature predicate: every candidate reports
// the same Signature as accepted, so tests can exercise a move in
// isolation without any real comparator behind it.
func alwaysInteresting(sig reduce.Signature) reduce.Predicate {
	return func(_ context.Context, _ core.Program) (reduce.Signature, bool, error) {
		return sig, true, nil
	}
}

func neverInteresting() reduce.Predicate {
	return func(_ context.Context, _ core.Program) (reduce.Signature, bool, error) {
		return reduce.Signature{}, false, nil
	}
}

// countingPredicate wraps a delegate and counts invocations.
type countingPredicate struct {
	calls    int
	delegate reduce.Predicate
}

func (c *countingPredicate) predicate() reduce.Predicate {
	return func(ctx context.Context, p core.Program) (reduce.Signature, bool, error) {
		c.calls++
		return c.delegate(ctx, p)
	}
}

// ---------------------------------------------------------------------
// Task 1 tests
// ---------------------------------------------------------------------

// TestReduceHasExactlyFiveMoves is Phase 11's own flip of this test's
// former name (D-11-29): the two whole-program moves are now PREPENDED to
// the original five, so Moves() returns exactly seven entries. The name
// is kept (not renamed) so this test's own git history stays attached to
// the assertion it has always made -- "Moves() returns a fixed, named
// order" -- even as the fixed order itself grows by two.
func TestReduceHasExactlyFiveMoves(t *testing.T) {
	moves := reduce.Moves()
	if len(moves) != 7 {
		t.Fatalf("expected exactly 7 moves (2 new whole-program + 5 original), got %d", len(moves))
	}
	want := []string{
		"drop-call-site",
		"drop-orphan-function",
		"drop-unused-binding",
		"drop-unmatched-arm",
		"drop-offpath-foreign-stage",
		"collapse-branch-to-diverging-arm",
		"truncate-to-minimal-prefix",
	}
	for i, name := range want {
		if moves[i].Name != name {
			t.Fatalf("move %d: got %q, want %q", i, moves[i].Name, name)
		}
	}
}

func TestReduceIsDeterministic(t *testing.T) {
	seed := borrowChainSeed()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interpreter-vs-O0", OperationID: seed.Functions[0].Linear.Operations[3].ID}

	first, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("first reduce: %v", err)
	}
	second, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("second reduce: %v", err)
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
		t.Fatalf("reduce is not deterministic:\nfirst:  %s\nsecond: %s", firstBytes, secondBytes)
	}
	if first.Source != second.Source {
		t.Fatalf("projected source is not deterministic:\nfirst:  %s\nsecond: %s", first.Source, second.Source)
	}
}

// TestReduceRespectsBudget engineers a seed with far more removable
// foreign-chain stages (100) than MaxReductionAttempts (64), and an
// always-accepting predicate, so Reduce keeps making genuine progress
// every cycle yet still cannot reach a true fixpoint before the budget is
// spent.
func TestReduceRespectsBudget(t *testing.T) {
	seed := foreignChainSeedWithSteps(100)
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(reduce.Signature{Axis: "a", EnginePair: "p", CausalRole: "c"}))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if result.Minimality != reduce.MinimalityBudgetExhausted {
		t.Fatalf("expected budget_exhausted, got %q", result.Minimality)
	}
	if result.Attempts != reduce.AttemptsPerFunction {
		t.Fatalf("expected Attempts == %d, got %d", reduce.AttemptsPerFunction, result.Attempts)
	}
}

func TestReduceCountsWork(t *testing.T) {
	seed := foreignChainSeed()
	counting := &countingPredicate{delegate: func(_ context.Context, _ core.Program) (reduce.Signature, bool, error) {
		return reduce.Signature{}, false, nil
	}}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, counting.predicate())
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if result.TotalRecomputedWork != counting.calls {
		t.Fatalf("TotalRecomputedWork (%d) does not match the number of check+compile+execute cycles observed by the fake predicate (%d)", result.TotalRecomputedWork, counting.calls)
	}
	if counting.calls == 0 {
		t.Fatal("expected the fake predicate to be invoked at least once")
	}
}

func TestReducePackageNeverTouchesSourceText(t *testing.T) {
	for _, file := range []string{"reduce.go", "predicate.go"} {
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("reading %s: %v", file, err)
		}
		for _, banned := range []string{"compiler/syntax", "compiler/ast"} {
			if strings.Contains(string(contents), banned) {
				t.Fatalf("%s must not import %q -- this package operates on core.Program only, never source text", file, banned)
			}
		}
	}
}

// ---------------------------------------------------------------------
// Move-specific unit tests
// ---------------------------------------------------------------------

func TestDropUnusedBindingRemovesOnlyUnreadOperation(t *testing.T) {
	seed := borrowChainSeed()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", OperationID: seed.Functions[0].Linear.Operations[3].ID}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	for _, op := range result.Program.Functions[0].Linear.Operations {
		if op.TargetID == seed.Functions[0].Linear.Operations[2].TargetID {
			t.Fatalf("expected the unused borrow (place:3) to have been dropped, still present: %+v", op)
		}
	}
	if len(result.Program.Functions[0].Linear.Operations) >= len(seed.Functions[0].Linear.Operations) {
		t.Fatalf("expected the operation count to shrink, got %d (seed had %d)", len(result.Program.Functions[0].Linear.Operations), len(seed.Functions[0].Linear.Operations))
	}
}

func TestDropOffpathForeignStageRemovesACompletedPair(t *testing.T) {
	seed := foreignChainSeed()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "does-not-move"}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	callCount := 0
	releaseCount := 0
	for _, op := range result.Program.Functions[0].Linear.Operations {
		switch op.Kind {
		case core.OpForeignCall:
			callCount++
		case core.OpRelease:
			releaseCount++
		}
	}
	if callCount != 0 || releaseCount != 0 {
		t.Fatalf("expected every acquire/release pair to be reducible away under an always-interesting predicate, got %d calls / %d releases", callCount, releaseCount)
	}
}

func TestDropUnmatchedArmNeverDropsBelowTwoArms(t *testing.T) {
	seed := threeArmMatchSeed()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "any"}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if result.Program.Functions[0].Match != nil && len(result.Program.Functions[0].Match.Arms) < 2 {
		t.Fatalf("drop-unmatched-arm must never reduce a match below 2 arms, got %d", len(result.Program.Functions[0].Match.Arms))
	}
}

func TestCollapseBranchToDivergingArmProducesStraightLineBody(t *testing.T) {
	seed := twoArmMatchSeed()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "arm-0"}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	fn := result.Program.Functions[0]
	if fn.Match != nil {
		t.Fatalf("expected the match to be collapsed away, got %+v", fn.Match)
	}
	if fn.Linear == nil {
		t.Fatal("expected a linear body to remain after collapse")
	}
}

func TestPhase17ReducerTwoTypeFacts(t *testing.T) {
	seed := twoArmMatchSeed()
	function := &seed.Functions[0]
	parameterTypeID := function.ID + ":type:parameter"
	returnTypeID := function.ID + ":type:return"
	function.Linear.Types = []core.TypeFact{{ID: parameterTypeID}, {ID: returnTypeID}}
	function.Linear.Places[0].TypeID = parameterTypeID
	// Every surviving branch operation is return-side typed: whichever arm a
	// prior narrowing move leaves behind, a first-operation fallback would
	// corrupt the parameter projection immediately.
	for index := range function.Linear.Operations {
		function.Linear.Operations[index].TypeID = returnTypeID
	}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, func(_ context.Context, candidate core.Program) (reduce.Signature, bool, error) {
		return reduce.Signature{Axis: "phase17", EnginePair: "pair", CausalRole: "two-types"}, candidate.Functions[0].Match == nil, nil
	})
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	got := result.Program.Functions[0]
	if got.Match != nil || got.Linear == nil {
		t.Fatalf("expected collapsed linear projection, got %+v", got)
	}
	if got.Linear.Places[0].TypeID != parameterTypeID {
		t.Fatalf("parameter TypeID = %q, want %q", got.Linear.Places[0].TypeID, parameterTypeID)
	}
	if got.Linear.Operations[0].TypeID != returnTypeID {
		t.Fatalf("operation TypeID = %q, want return-side %q", got.Linear.Operations[0].TypeID, returnTypeID)
	}
}

func TestPhase17ReducerOperationOrderIndependent(t *testing.T) {
	seed := twoArmMatchSeed()
	function := &seed.Functions[0]
	parameterTypeID := function.ID + ":type:parameter"
	returnTypeID := function.ID + ":type:return"
	function.Linear.Types = []core.TypeFact{{ID: parameterTypeID}, {ID: returnTypeID}}
	function.Linear.Places[0].TypeID = parameterTypeID
	for index := range function.Linear.Operations {
		function.Linear.Operations[index].TypeID = returnTypeID
	}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, func(_ context.Context, candidate core.Program) (reduce.Signature, bool, error) {
		return reduce.Signature{Axis: "phase17", EnginePair: "pair", CausalRole: "reordered"}, candidate.Functions[0].Match == nil, nil
	})
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if got := result.Program.Functions[0].Linear.Places[0].TypeID; got != parameterTypeID {
		t.Fatalf("parameter TypeID changed with operation order: got %q want %q", got, parameterTypeID)
	}
}

func TestPhase17ReducerMissingFactFailsClosed(t *testing.T) {
	seed := twoArmMatchSeed()
	seed.Functions[0].Linear.Places[0].TypeID = ""
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(reduce.Signature{Axis: "phase17", EnginePair: "pair", CausalRole: "missing-parameter-fact"}))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if result.Program.Functions[0].Match == nil {
		t.Fatalf("reducer collapsed a match with no explicit parameter TypeID: %+v", result.Program.Functions[0].Linear)
	}
}

func TestPhase17ReducerForeignChainRetainsReturnFact(t *testing.T) {
	seed := recheck(t, realForeignChainSource)
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(reduce.Signature{Axis: "phase17", EnginePair: "pair", CausalRole: "foreign-return"}))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	linear := result.Program.Functions[0].Linear
	if len(linear.Types) != 2 {
		t.Fatalf("reduced plain return retained %d type facts, want parameter and return: %+v", len(linear.Types), linear.Types)
	}
	if linear.Types[0].ID != result.Program.Functions[0].ID+":type:0" || linear.Types[1].ID != result.Program.Functions[0].ID+":type:1" {
		t.Fatalf("reduced signature facts = %+v, want type:0/type:1", linear.Types)
	}
	if linear.Types[1].Shape.Constructor != result.Program.Functions[0].ReturnType {
		t.Fatalf("return fact = %+v, declared return = %q", linear.Types[1], result.Program.Functions[0].ReturnType)
	}
}

func TestTruncateToMinimalPrefixShortensStraightLineTail(t *testing.T) {
	seed := borrowChainSeedNoUnusedTail()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "first-borrow"}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if len(result.Program.Functions[0].Linear.Operations) >= len(seed.Functions[0].Linear.Operations) {
		t.Fatalf("expected truncation to shrink the operation count, got %d (seed had %d)", len(result.Program.Functions[0].Linear.Operations), len(seed.Functions[0].Linear.Operations))
	}
}

// TestReduceRejectsMultiFunctionSeed is Phase 11's own flip of this test's
// former name (D-11-30): a multi-function seed is no longer rejected --
// reduce_multifunction_test.go's TestReduceMultiFunctionSeed and
// TestReduceSingleFunctionOutputUnchanged carry the real Phase 11
// coverage; this smoke test just proves the old hard error is gone.
//
// WR-01: this test formerly passed Seed{Program: seed} with
// EntryFunctionID left at its zero value and asserted only err == nil,
// which is precisely the usage hazard Seed.Validate now refuses -- and it
// never inspected result.Program.Functions to confirm the nominated entry
// survived. It now supplies a real entry ID and asserts survival, so the
// test witnesses the property it always claimed to.
func TestReduceRejectsMultiFunctionSeed(t *testing.T) {
	seed := borrowChainSeed()
	second := foreignChainSeed()
	seed.Functions = append(seed.Functions, second.Functions...)
	entryID := seed.Functions[0].ID
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed, EntryFunctionID: entryID}, alwaysInteresting(reduce.Signature{}))
	if err != nil {
		t.Fatalf("expected no error for a multi-function seed (D-11-30), got %v", err)
	}
	found := false
	for _, fn := range result.Program.Functions {
		if fn.ID == entryID {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the nominated entry function %q to survive reduction, got %d function(s)", entryID, len(result.Program.Functions))
	}
}

// ---------------------------------------------------------------------
// Additional fixtures
// ---------------------------------------------------------------------

// borrowChainSeedNoUnusedTail is like borrowChainSeed but every borrow feeds
// the next, and the final return reads the LAST borrow's target directly --
// truncate-to-minimal-prefix can therefore shorten the tail by dropping the
// last borrow and rewiring the return to the second-to-last.
func borrowChainSeedNoUnusedTail() core.Program {
	const fn = "s1:phase5.reduce_fixture:function:touch2"
	return core.Program{
		Schema:   core.Schema,
		Module:   "phase5.reduce_fixture",
		ModuleID: "s1:phase5.reduce_fixture:module:phase5.reduce_fixture",
		Functions: []core.Function{
			{
				ID: fn, Name: "touch2",
				EntryPointID: fn + ":point:entry", ReturnPointID: fn + ":point:return",
				Parameter:  core.Parameter{ID: fn + ":place:0", Name: "buffer", Type: "Buffer"},
				ReturnType: "Buffer",
				Linear: &core.LinearBody{
					ID:     fn + ":linear",
					Types:  []core.TypeFact{{ID: fn + ":type:0"}},
					Places: []core.Place{{ID: fn + ":place:0", Name: "buffer", TypeID: fn + ":type:0"}},
					Operations: []core.LinearOperation{
						{ID: fn + ":op:0", PointID: fn + ":point:linear:0", Kind: core.OpBorrowExclusive, SourceID: fn + ":place:0", TargetID: fn + ":place:1", TypeID: fn + ":type:0"},
						{ID: fn + ":op:1", PointID: fn + ":point:linear:1", Kind: core.OpBorrowShared, SourceID: fn + ":place:1", TargetID: fn + ":place:2", TypeID: fn + ":type:0"},
						{ID: fn + ":op:2", PointID: fn + ":point:linear:2", Kind: core.OpReturn, SourceID: fn + ":place:2", TypeID: fn + ":type:0"},
					},
				},
			},
		},
	}
}

// twoArmMatchSeed mirrors testdata/phase5/defect_dies_by_signal.lang's
// shape: a Signal-typed parameter, one arm that takes/returns, one arm
// that defects.
func twoArmMatchSeed() core.Program {
	const fn = "s1:phase5.reduce_fixture:function:triage"
	const matchID = fn + ":match"
	parameterID := fn + ":place:0"
	typeID := fn + ":type:0"

	goAliasOp := fn + ":op:0"
	goAliasPlace := fn + ":place:1"
	goTakeOp := fn + ":op:1"
	goTakePlace := fn + ":place:2"
	goReturnOp := fn + ":op:2"
	goPad := fn + ":place:3"

	haltAliasOp := fn + ":op:3"
	haltAliasPlace := fn + ":place:4"
	haltDefectOp := fn + ":op:4"
	haltPad := fn + ":place:5"

	return core.Program{
		Schema:   core.Schema,
		Module:   "phase5.reduce_fixture",
		ModuleID: "s1:phase5.reduce_fixture:module:phase5.reduce_fixture",
		DataTypes: []core.DataType{
			{ID: "s1:phase5.reduce_fixture:type:Signal", Name: "Signal", Alternatives: []string{"Go", "Halt"}},
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
					},
				},
				Linear: &core.LinearBody{
					ID: fn + ":linear", Types: []core.TypeFact{{ID: typeID}},
					Places: []core.Place{
						{ID: parameterID, Name: "flag", TypeID: typeID},
						{ID: goAliasPlace, Name: "flag", TypeID: typeID},
						{ID: goTakePlace, Name: "held", TypeID: typeID},
						{ID: goPad, Name: "_", TypeID: typeID},
						{ID: haltAliasPlace, Name: "flag", TypeID: typeID},
						{ID: haltPad, Name: "_", TypeID: typeID},
					},
					Operations: []core.LinearOperation{
						{ID: goAliasOp, PointID: fn + ":point:linear:0", Kind: core.OpCopy, SourceID: parameterID, TargetID: goAliasPlace, TypeID: typeID},
						{ID: goTakeOp, PointID: fn + ":point:linear:1", Kind: core.OpMove, SourceID: goAliasPlace, TargetID: goTakePlace, TypeID: typeID},
						{ID: goReturnOp, PointID: fn + ":point:linear:2", Kind: core.OpReturn, SourceID: goTakePlace, TypeID: typeID},
						{ID: haltAliasOp, PointID: fn + ":point:linear:3", Kind: core.OpCopy, SourceID: parameterID, TargetID: haltAliasPlace, TypeID: typeID},
						{ID: haltDefectOp, PointID: fn + ":point:linear:4", Kind: core.OpDefect, SourceID: haltAliasPlace, TypeID: typeID, Reason: "adversarial defect"},
					},
					Blocks: []core.Block{
						{ID: fn + ":block:entry", PointID: fn + ":point:entry", Successors: []string{fn + ":block:arm:0", fn + ":block:arm:1"}},
						{ID: fn + ":block:arm:0", PointID: fn + ":point:arm:0", OperationIDs: []string{goAliasOp, goTakeOp, goReturnOp}, Successors: []string{fn + ":block:join"}},
						{ID: fn + ":block:arm:1", PointID: fn + ":point:arm:1", OperationIDs: []string{haltAliasOp, haltDefectOp}, Successors: []string{fn + ":block:join"}},
						{ID: fn + ":block:join", PointID: fn + ":point:return"},
					},
					Edges: []core.Edge{
						{ID: fn + ":edge:entry:arm:0", FromBlockID: fn + ":block:entry", ToBlockID: fn + ":block:arm:0", Pattern: "Go"},
						{ID: fn + ":edge:arm:0:join", FromBlockID: fn + ":block:arm:0", ToBlockID: fn + ":block:join", Pattern: "Go"},
						{ID: fn + ":edge:entry:arm:1", FromBlockID: fn + ":block:entry", ToBlockID: fn + ":block:arm:1", Pattern: "Halt"},
						{ID: fn + ":edge:arm:1:join", FromBlockID: fn + ":block:arm:1", ToBlockID: fn + ":block:join", Pattern: "Halt"},
					},
				},
			},
		},
	}
}

// threeArmMatchSeed is a synthetic (never-emitted-by-check.Program) 3-arm
// match used only to exercise drop-unmatched-arm's own distinct behavior
// from collapse-branch-to-diverging-arm's 2-arm case.
func threeArmMatchSeed() core.Program {
	seed := twoArmMatchSeed()
	fn := &seed.Functions[0]
	thirdArmOp := fn.ID + ":op:5"
	thirdArmPlace := fn.ID + ":place:6"
	fn.Match.Arms = append(fn.Match.Arms, core.MatchArm{
		ID: fn.ID + ":arm:2", EdgeID: fn.Match.ID + ":edge:Extra", Pattern: "Extra", BlockID: fn.ID + ":block:arm:2",
	})
	fn.Linear.Operations = append(fn.Linear.Operations, core.LinearOperation{
		ID: thirdArmOp, PointID: fn.ID + ":point:linear:5", Kind: core.OpReturn, SourceID: fn.Parameter.ID, TypeID: fn.Linear.Types[0].ID,
	})
	fn.Linear.Places = append(fn.Linear.Places, core.Place{ID: thirdArmPlace, Name: "_", TypeID: fn.Linear.Types[0].ID})
	fn.Linear.Blocks = append(fn.Linear.Blocks, core.Block{
		ID: fn.ID + ":block:arm:2", PointID: fn.ID + ":point:arm:2", OperationIDs: []string{thirdArmOp}, Successors: []string{fn.ID + ":block:join"},
	})
	return seed
}

// ---------------------------------------------------------------------
// Task 3 tests: ProjectSource's round-trip correspondence with the real
// shipped front end (D-05-23).
// ---------------------------------------------------------------------

// recheck re-parses and re-checks source through the REAL shipped front end
// -- the round-trip authority every test in this section depends on.
func recheck(t *testing.T, source string) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse failed for projected source:\n%s\ndiagnostics: %+v", source, parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) > 0 {
		t.Fatalf("check failed for projected source:\n%s\ndiagnostics: %+v", source, result.Diagnostics)
	}
	return result.Program
}

// structurallyEqual compares two core.Program values on operation ordinals
// and block graph -- the correspondence D-05-23 requires -- by comparing
// their full serialized bytes (every field this reducer touches is already
// part of that ordinal/block-graph identity; a byte comparison is the
// strictest, least-assumption-laden form of "structurally equal" available
// without re-implementing a bespoke semantic diff).
func structurallyEqual(t *testing.T, got, want core.Program) {
	t.Helper()
	gotBytes, err := json.Marshal(stripSpans(got))
	if err != nil {
		t.Fatalf("marshal got: %v", err)
	}
	wantBytes, err := json.Marshal(stripSpans(want))
	if err != nil {
		t.Fatalf("marshal want: %v", err)
	}
	if string(gotBytes) != string(wantBytes) {
		t.Fatalf("programs are not structurally equal:\nrechecked: %s\nreduced:   %s", gotBytes, wantBytes)
	}
}

// stripSpans zeroes every diagnostic.Span this comparison should ignore --
// "structurally equal on operation ordinals and block graph" (the plan's
// own wording) never meant "byte-identical source positions". A reduced
// program's spans point into the ORIGINAL seed's source text, while a
// fresh recheck's spans point into the freshly-projected source text; the
// two texts are never byte-identical (different whitespace, comments,
// binding names survive reduction), so their spans are expected to differ
// even when the programs are otherwise structurally identical.
func stripSpans(p core.Program) core.Program {
	encoded, err := json.Marshal(p)
	if err != nil {
		return p
	}
	var clone core.Program
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return p
	}
	for i := range clone.DataTypes {
		clone.DataTypes[i].Span = diagnostic.Span{}
	}
	for i := range clone.Functions {
		clone.Functions[i].Span = diagnostic.Span{}
	}
	return clone
}

// realStraightLineSource mirrors testdata/phase5/restrict_borrow.lang's
// shape.
const realStraightLineSource = `module phase5.reduce_roundtrip

export {
  fn touch
}

fn touch(buffer: Buffer) -> Buffer {
  let first = borrow mut buffer
  let second = borrow first
  second
}
`

// realForeignChainSource mirrors testdata/phase5/tail_collapse_release_ladder.lang's
// shape (narrowed to two steps).
const realForeignChainSource = `module phase5.reduce_roundtrip

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let a = try lang_res_open(request)
  let b = try lang_res_open(request)
  request
}
`

// realMatchSource mirrors testdata/phase5/defect_dies_by_signal.lang's
// shape.
const realMatchSource = `module phase5.reduce_roundtrip

export {
  type Signal
  fn triage
}

data Signal =
  | Go
  | Halt

fn triage(flag: Signal) -> Signal {
  match flag {
    Go => {
      let held = take flag
      held
    }
    Halt => {
      defect "adversarial defect: dies by signal"
    }
  }
}
`

func TestProjectedSourceCorrespondsToReducedCore(t *testing.T) {
	t.Run("straight_line_borrow_chain_unreduced", func(t *testing.T) {
		seed := recheck(t, realStraightLineSource)
		source := reduce.ProjectSource(seed)
		rechecked := recheck(t, source)
		structurallyEqual(t, rechecked, seed)
	})

	t.Run("foreign_resource_chain_unreduced", func(t *testing.T) {
		seed := recheck(t, realForeignChainSource)
		source := reduce.ProjectSource(seed)
		rechecked := recheck(t, source)
		structurallyEqual(t, rechecked, seed)
	})

	t.Run("foreign_resource_chain_after_offpath_drop", func(t *testing.T) {
		// Run the real reduce loop (drop-offpath-foreign-stage) over a
		// genuine checked seed, then require the REDUCED result to still
		// round-trip -- this is the actual end-to-end path plan 05-12 wires
		// up, not just the unreduced seed.
		seed := recheck(t, realForeignChainSource)
		sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "does-not-move"}
		result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
		if err != nil {
			t.Fatalf("reduce: %v", err)
		}
		if result.Source == "" || strings.HasPrefix(result.Source, "// reduce: projection unsupported") {
			t.Fatalf("expected a supported projection, got: %s", result.Source)
		}
		rechecked := recheck(t, result.Source)
		structurallyEqual(t, rechecked, result.Program)
	})

	t.Run("collapsed_adt_match_projection_is_documented_unsupported", func(t *testing.T) {
		// Reduce a two-arm ADT-typed (Signal) match down via
		// collapse-branch-to-diverging-arm. The CORE-level collapse itself
		// succeeds (reduction never touches source text), but this
		// project's checker only ever admits a declared ADT parameter/return
		// type inside an EXHAUSTIVE match body -- there is no partial-match
		// syntax -- so a collapsed ADT-typed function has no valid,
		// re-checkable source projection under the current grammar.
		// ProjectSource reports this explicitly rather than emitting
		// uncheckable or silently-wrong source (documented narrowing, see
		// reduce.go's ProjectSource doc comment and 05-10-SUMMARY.md).
		seed := recheck(t, realMatchSource)
		sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "go-arm"}
		result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
		if err != nil {
			t.Fatalf("reduce: %v", err)
		}
		if result.Program.Functions[0].Match != nil {
			t.Fatal("expected the match to have collapsed away at the core level for this fixture")
		}
		if !strings.HasPrefix(result.Source, "// reduce: projection unsupported") {
			t.Fatalf("expected the documented unsupported-projection marker for a collapsed ADT-typed match, got: %s", result.Source)
		}
	})
}

func TestProjectedSourceIsDeterministic(t *testing.T) {
	seed := foreignChainSeed()
	first := reduce.ProjectSource(seed)
	second := reduce.ProjectSource(seed)
	if first != second {
		t.Fatalf("projection is not deterministic:\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestResultSourceAlwaysMatchesResultProgram(t *testing.T) {
	seed := foreignChainSeed()
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "does-not-move"}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	reprojected := reduce.ProjectSource(result.Program)
	if reprojected != result.Source {
		t.Fatalf("Result.Source does not match a fresh projection of Result.Program:\nResult.Source: %s\nreprojected:   %s", result.Source, reprojected)
	}
}

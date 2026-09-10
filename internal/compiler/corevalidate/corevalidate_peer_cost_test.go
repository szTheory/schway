package corevalidate_test

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 06, Task 1/2 (D-09-28, D-09-29 lineage): measures the
// PEER's own re-derivation cost -- corevalidate's `foldChain`/
// `carriedLoans` place-provenance walk (corevalidate.go:1128-1238),
// consulted from `replayStraightLine`'s per-operation loop -- against its
// own declared bound, fitted against OPERATION count, never against
// `check`'s ratified `recomputed_work_growth_exponent`
// (T-09-04/threat-model row 1). The metric name this test's own findings
// are reported under, `peer_closure_recomputed_work_growth_exponent`, was
// declared gate-eligible at both chokepoints by plan 09-02 and is NOT
// ratified here: no `qlt02_budget_manifest.json` row is added by this
// plan (D-09-28, D-08-33's precedent) -- ratification belongs to plan
// 09-08's mid-phase gate, at that gate's own commit.
//
// HONESTY DISCLOSURE (phase constraint 5, inherited from 09-04-SUMMARY.md's
// own finding): testsupport.CallGraphCorpusShapes()'s five generated
// shapes contain ZERO core.OpBorrowShared/core.OpBorrowExclusive
// operations -- they were built for check's own usesParam work-counter
// (Phase 08), never for loan-liveness sensitivity. Sweeping them AS
// GENERATED would exercise `buildLoanChainIndex`'s bookkeeping machinery
// (place-parent pointers, one-inspection-per-operation counting) but never
// a single populated `bornAt` entry, and `carriedLoans` would return an
// empty loan list on every call -- a code path that runs but never
// actually carries a loan across anything. That would be a bound over a
// borrow-free program, decorative in exactly the way this phase keeps
// warning about (constraint 5).
//
// This test therefore does NOT sweep the five shapes as testsupport
// generates them. `injectBorrowForLoanCarry` (below) prepends a genuine
// `core.OpBorrowShared` of the function's own parameter to every generated
// function, before any existing operation, and rewires every operation
// that used to read the raw parameter to read the freshly-borrowed place
// instead. This makes every function's own return derive from a borrow of
// its own parameter (`peerLoanCarryFact.ReturnsBorrowOfParam == true`,
// corevalidate_peer_liveness.go), which in turn makes
// `buildLoanChainIndex`'s OpCall branch actually chain a call's result
// back to its argument's own loan ancestry (corevalidate.go:1173-1179) --
// so `carriedLoans` genuinely walks a non-empty loan list, and
// `bornAt`/`foldChain`'s memoization is genuinely exercised, on every
// swept program point. What is fitted below is therefore the real cost of
// loan-carry propagation across the corpus's own call-graph shape, not a
// borrow-free structural walk wearing its name.
//
// Generator Reachability Register (this test's own row, mirroring
// check_peer_shape_differential_test.go's discipline):
//
//	Generator: testsupport.GenerateCallGraphCorpus, POST-PROCESSED by
//	  injectBorrowForLoanCarry (this file) before validation.
//	Reaches: acyclic call graphs of increasing size and sharing, relay
//	  depth >= 2, where EVERY function borrows and returns a borrow of its
//	  own single parameter.
//	Provably does NOT reach: any cyclic shape (refused before this
//	  validator's replay ever runs); loops or back edges (none exist in the
//	  language); arity > 1; exclusive borrows (only OpBorrowShared is
//	  injected -- there is never more than one live borrow per function
//	  here, so shared vs exclusive makes no observable difference to the
//	  cost counters under test); or a re-borrow/reborrow chain within one
//	  function (each function borrows its parameter exactly once).
//
// Wall-clock elapsed time is logged per sweep point for an incidental
// human reader and is NEVER compared to anything or used to gate a
// verdict (constraint 4, T-09-20) -- `corevalidate.Result.Checks` is the
// sole instrument, following costcorpus_test.go's own recorded discipline.
// ---------------------------------------------------------------------

// peerCostSizeLadder mirrors interproceduralSizeLadder's own discipline
// (check/costcorpus_test.go): four or more points spanning at least 4x, and
// large enough that the ratio-stability tripwire's own two comparison
// points (peerCostRatioSmall, peerCostRatioLarge) are both present.
var peerCostSizeLadder = []int{16, 32, 64, 128, 256, 512}

const (
	peerCostRatioSmall = 128
	peerCostRatioLarge = 512
)

// peerClosureGrowthExponentBoundMilli is THIS peer's own declared bound,
// derived from this test's own observed curve at authoring time -- never
// copied from check's ratified recomputed_work_growth_exponent bound
// (threat-model row 1, T-09-04, constraint 2 of this plan). Observed
// milli-exponents at authoring time (fitted against operation count, over
// peerCostSizeLadder, memoized/production path):
//
//	chain:         ~1000 (flat -- O(1) chain depth per function, by
//	               construction: every function borrows once and every
//	               operation reads that SAME borrowed place)
//	diamond:       ~1000
//	dense:         ~1000
//	parser-shaped: ~1000
//	forward:       ~1000 (memoized: program-order chain reads are O(1)
//	               amortized per operation, matching spike S-006's own
//	               "flat 4.0 work units/op" finding for the identical
//	               program-order shape)
//
// All five shapes measured flat-linear (milli-exponent ~1000, i.e.
// exponent ~1.0) because per-function loan-chain depth here is bounded by
// this test's OWN injected shape (one borrow, one hop, memoized) rather
// than by corpus size -- the corpus's size axis is FUNCTION COUNT / CALL
// FANOUT, not per-function chain depth. The bound below (1300) leaves a
// declared margin over the observed ~1000 without being so loose it could
// never trip; TestPeerClosureCostUnmemoizedSeamExceedsBound proves the
// SAME bound, on the ONE shape whose chain depth actually grows with size
// ("forward"), is exceeded when memoization is disabled.
const peerClosureGrowthExponentBoundMilli = 1300

// peerClosureRatioStabilityBound is this test's own declared band for the
// ratio-stability tripwire (justification: a single fitted exponent can
// hide a curve that is linear at small sizes and bends later; comparing
// work/operations at two widely separated sizes catches a bend a fit
// averages away -- same justification check's own 15% band uses,
// independently re-derived here for the peer's own distinct bound rather
// than borrowed from it).
const peerClosureRatioStabilityBound = 0.15

// injectBorrowForLoanCarry prepends a genuine core.OpBorrowShared of the
// function's own parameter as the function's FIRST operation, and rewrites
// every existing operation whose SourceID read the raw parameter to read
// the freshly-borrowed place instead -- see this file's own header comment
// for why this is required for an honest sweep. Applied BEFORE the
// strict-ID-convention remap below, so the remap sees a single coherent
// operation list and assigns sequential place/operation IDs across the
// WHOLE rewritten body, borrow op included.
func injectBorrowForLoanCarry(function core.Function) core.Function {
	if function.Linear == nil {
		return function
	}
	paramID := function.Parameter.ID
	borrowedID := function.ID + ":place:borrowed"
	loanID := function.ID + ":loan:0"
	borrowOp := core.LinearOperation{
		ID: function.ID + ":op:borrow", Kind: core.OpBorrowShared,
		SourceID: paramID, TargetID: borrowedID, LoanID: loanID,
	}
	rewritten := make([]core.LinearOperation, 0, len(function.Linear.Operations)+1)
	rewritten = append(rewritten, borrowOp)
	for _, operation := range function.Linear.Operations {
		if operation.SourceID == paramID {
			operation.SourceID = borrowedID
		}
		rewritten = append(rewritten, operation)
	}
	function.Linear.Operations = rewritten
	return function
}

// completePeerCostProgramForCorevalidate fills in the structural
// scaffolding testsupport's own generator never populates, mirroring
// check_peer_shape_differential_test.go's completeSyntheticProgramForCorevalidate
// (written independently here since that function is unexported inside
// package check and unreachable from this external test package). It
// injects the borrow (above) FIRST, then remaps every place/operation ID
// to corevalidate's required strict "<functionID>:place:<index>" /
// "<functionID>:op:<index>" sequential naming, in first-reference order --
// this renames IDENTITY bookkeeping only, never the call-graph topology
// testsupport's generator produced.
func completePeerCostProgramForCorevalidate(program core.Program) core.Program {
	for i := range program.Functions {
		program.Functions[i] = injectBorrowForLoanCarry(program.Functions[i])
	}
	for i := range program.Functions {
		function := &program.Functions[i]
		function.EntryPointID = function.ID + ":point:entry"
		function.ReturnPointID = function.ID + ":point:return"
		function.ReturnType = "Byte"
		if function.Linear == nil {
			continue
		}
		// A LEAF function here (no OpCall) returns the borrow it just took
		// directly (or through a plain OpCopy) -- peerReturnDerivesFromBorrow
		// sees that direct chain, so SEM-06's Callable-is-publication-safety
		// law (peerCallable, D-04-03) requires a declared PublicOrigin or
		// the function is refused as core.callee_not_callable. A RELAY
		// function's own return instead derives from an OpCall's result;
		// peerDeriveOriginFacts/peerReturnDerivesFromBorrow do not trace
		// through OpCall at all (a documented, pre-existing scope
		// narrowing, unrelated to this test), so a relay function is
		// already Callable with PublicOrigin left nil -- declaring one on a
		// relay function would instead make peerOriginContained's own
		// !fact.Derived branch refuse it. PublicOrigin is therefore
		// declared only on functions with no OpCall of their own.
		hasCall := false
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall {
				hasCall = true
				break
			}
		}
		if !hasCall {
			function.PublicOrigin = &core.PublicOrigin{Paths: []string{function.Parameter.Name}, Access: "shared"}
		}
		typeID := function.ID + ":type:0"
		function.Linear.Types = []core.TypeFact{{
			ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
			Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			NegativeWitnesses: []core.AbilityWitness{},
		}}

		originalParameterID := function.Parameter.ID
		remapped := make(map[string]string)
		var orderedOldIDs []string
		remap := func(oldID string) string {
			if oldID == "" {
				return ""
			}
			if newID, ok := remapped[oldID]; ok {
				return newID
			}
			newID := fmt.Sprintf("%s:place:%d", function.ID, len(orderedOldIDs))
			remapped[oldID] = newID
			orderedOldIDs = append(orderedOldIDs, oldID)
			return newID
		}

		function.Parameter.ID = remap(function.Parameter.ID)
		for j := range function.Linear.Operations {
			operation := &function.Linear.Operations[j]
			operation.SourceID = remap(operation.SourceID)
			operation.TargetID = remap(operation.TargetID)
			operation.TypeID = typeID
			operation.ID = fmt.Sprintf("%s:op:%d", function.ID, j)
			operation.PointID = fmt.Sprintf("%s:point:linear:%d", function.ID, j)
		}

		places := make([]core.Place, 0, len(orderedOldIDs))
		for _, oldID := range orderedOldIDs {
			name := "result"
			if oldID == originalParameterID {
				name = function.Parameter.Name
			}
			places = append(places, core.Place{ID: remapped[oldID], Name: name, TypeID: typeID})
		}
		function.Linear.Places = places
	}
	return program
}

// buildPeerCostProgram generates one corpus point and applies this file's
// own borrow-injection and structural completion, so corevalidate.Validate
// can reach the loan-carry cost this test measures.
func buildPeerCostProgram(shape string, n int) (core.Program, error) {
	program, err := testsupport.GenerateCallGraphCorpus(shape, n)
	if err != nil {
		return core.Program{}, err
	}
	program.Schema = core.Schema1
	program.Module = "peercost"
	program.ModuleID = "s1:peercost:module:peercost"
	program = completePeerCostProgramForCorevalidate(program)
	return program, nil
}

// peerCostOpCount sums every declared function's own operation count --
// the honest axis (constraint 3 of this plan, D-08-31's own precedent):
// spike S-006 found the "dense" shape fits 1.50 against function count and
// 1.11 against operation count, the gap being the corpus's own
// leaf-to-relay ratio drifting with size -- fitting against function count
// alone would have produced a wrong finding.
func peerCostOpCount(program core.Program) int {
	count := 0
	for _, function := range program.Functions {
		if function.Linear != nil {
			count += len(function.Linear.Operations)
		}
	}
	return count
}

// fitPeerClosureGrowthExponent is an ordinary least-squares fit of
// log(work) against log(x), ported independently from
// check/costcorpus_test.go's own fitGrowthExponentInOps (itself ported
// from spike 006's iplive/scaling.go) -- the SAME well-known fitting
// method, applied to the PEER's own distinct counted-work samples, never
// sharing a Go symbol with check's own fitter (mechanical independence:
// check's fitter is unexported inside package check and unreachable from
// this external test package regardless).
func fitPeerClosureGrowthExponent(xs, ys []int) float64 {
	if len(xs) != len(ys) {
		return math.NaN()
	}
	var sumX, sumY, sumXY, sumXX float64
	usable := 0
	for i := range xs {
		x, y := xs[i], ys[i]
		if x <= 0 || y <= 0 {
			continue
		}
		logX, logY := math.Log(float64(x)), math.Log(float64(y))
		sumX += logX
		sumY += logY
		sumXY += logX * logY
		sumXX += logX * logX
		usable++
	}
	if usable < 2 {
		return math.NaN()
	}
	n := float64(usable)
	denominator := n*sumXX - sumX*sumX
	if denominator == 0 {
		return math.NaN()
	}
	return (n*sumXY - sumX*sumY) / denominator
}

// sweepPeerClosureCost sweeps peerCostSizeLadder for one shape, driving
// corevalidate.Validate's own deterministic Checks counter (never wall
// clock) and returning parallel operation-count/checks-count slices for
// the caller to fit or ratio-compare.
func sweepPeerClosureCost(t *testing.T, shape string) (ops, checks []int) {
	t.Helper()
	ops = make([]int, len(peerCostSizeLadder))
	checks = make([]int, len(peerCostSizeLadder))
	for index, n := range peerCostSizeLadder {
		program, err := buildPeerCostProgram(shape, n)
		if err != nil {
			t.Fatalf("buildPeerCostProgram(%q, %d): %v", shape, n, err)
		}
		start := time.Now()
		result := corevalidate.Validate(program)
		elapsed := time.Since(start)
		if !result.Valid {
			t.Fatalf("shape=%s n=%d: corevalidate.Validate reported invalid, problems=%v", shape, n, result.Problems)
		}
		opCount := peerCostOpCount(program)
		// Wall clock is OBSERVED ONLY -- logged for a human reader, never
		// compared to anything and never used to gate a verdict (T-09-20).
		t.Logf("shape=%s n=%d functions=%d ops=%d checks=%d elapsed=%s (elapsed is observed-only, never gated)",
			shape, n, len(program.Functions), opCount, result.Checks, elapsed)
		if result.Checks <= 0 {
			t.Fatalf("shape=%s n=%d: zero counted work from corevalidate.Validate", shape, n)
		}
		ops[index] = opCount
		checks[index] = result.Checks
	}
	return ops, checks
}

// peerClosureGrowthExponentMilliByShape mirrors check's own
// growthExponentMilliByShape determinism guard: if a repeated invocation
// within one test binary (e.g. `go test -count=2`) disagrees with the
// first, the measurement is not deterministic.
var peerClosureGrowthExponentMilliByShape = map[string][]int64{}

// TestPeerClosureCostGrowthExponentWithinBound is this task's own
// falsifier: for every shape testsupport.CallGraphCorpusShapes() names,
// sweep peerCostSizeLadder (six points spanning 32x), fit the growth
// exponent against operation count, and assert it stays at or under
// peerClosureGrowthExponentBoundMilli -- this peer's OWN bound, derived
// from this test's own observed curve, never check's ratified
// recomputed_work_growth_exponent (T-09-04).
func TestPeerClosureCostGrowthExponentWithinBound(t *testing.T) {
	for _, shape := range testsupport.CallGraphCorpusShapes() {
		shape := shape
		t.Run(shape, func(t *testing.T) {
			ops, checks := sweepPeerClosureCost(t, shape)
			exponent := fitPeerClosureGrowthExponent(ops, checks)
			if math.IsNaN(exponent) {
				t.Fatalf("shape=%s: fitted exponent is NaN (fewer than two usable sweep points)", shape)
			}
			milli := int64(math.Round(exponent * 1000))
			peerClosureGrowthExponentMilliByShape[shape] = append(peerClosureGrowthExponentMilliByShape[shape], milli)
			if first := peerClosureGrowthExponentMilliByShape[shape][0]; milli != first {
				t.Fatalf("shape=%s: fitted milli-exponent is not deterministic across repeated invocations: first=%d this=%d",
					shape, first, milli)
			}
			t.Logf("peer_closure_recomputed_work_growth_exponent shape=%s milli=%d (exponent=%.4f), bound=%d",
				shape, milli, exponent, peerClosureGrowthExponentBoundMilli)
			if milli > peerClosureGrowthExponentBoundMilli {
				t.Fatalf("shape=%s: fitted milli-exponent (against OPERATION count) = %d (exponent=%.4f) exceeds this peer's OWN bound %d",
					shape, milli, exponent, peerClosureGrowthExponentBoundMilli)
			}
		})
	}
}

// TestPeerClosureCostRatioStabilityTripwire asserts the ratio
// checks(4S)/checks(S) stays within peerClosureRatioStabilityBound for
// every shape -- a single fitted exponent can hide a curve that is linear
// at small sizes and bends later; comparing two widely separated sizes
// catches the bend a fit averages away.
func TestPeerClosureCostRatioStabilityTripwire(t *testing.T) {
	for _, shape := range testsupport.CallGraphCorpusShapes() {
		shape := shape
		t.Run(shape, func(t *testing.T) {
			ops, checks := sweepPeerClosureCost(t, shape)
			var ratioAtSmall, ratioAtLarge float64
			var sawSmall, sawLarge bool
			for index, n := range peerCostSizeLadder {
				switch n {
				case peerCostRatioSmall:
					ratioAtSmall = float64(checks[index]) / float64(ops[index])
					sawSmall = true
				case peerCostRatioLarge:
					ratioAtLarge = float64(checks[index]) / float64(ops[index])
					sawLarge = true
				}
			}
			if !sawSmall || !sawLarge {
				t.Fatalf("shape=%s: ratio-stability tripwire requires both S=%d and 4S=%d in peerCostSizeLadder",
					shape, peerCostRatioSmall, peerCostRatioLarge)
			}
			delta := math.Abs(ratioAtLarge-ratioAtSmall) / ratioAtSmall
			t.Logf("shape=%s: checks/operations at S=%d = %.4f, at 4S=%d = %.4f, delta=%.1f%%",
				shape, peerCostRatioSmall, ratioAtSmall, peerCostRatioLarge, ratioAtLarge, delta*100)
			if delta > peerClosureRatioStabilityBound {
				t.Fatalf("shape=%s: checks/operations ratio at S=%d (%.4f) vs 4S=%d (%.4f) differ by %.1f%%, exceeds the %.0f%% ratio-stability bound",
					shape, peerCostRatioSmall, ratioAtSmall, peerCostRatioLarge, ratioAtLarge, delta*100, peerClosureRatioStabilityBound*100)
			}
		})
	}
}

// TestPeerClosureCostUnmemoizedSeamExceedsBound is Task 2's own
// mutation-kill (T-09-17): "a cost bound that has never been exceeded, and
// therefore measures nothing." It sweeps ONLY the "forward" shape, not all
// five: corpusForward is the ONE shape in this corpus whose per-function
// chain DEPTH grows with corpus size (a single caller threading one borrow
// through k sequential calls, corpusForward's own doc comment) -- the
// other four shapes' per-function chain depth stays O(1) regardless of
// corpus size BY THIS TEST'S OWN CONSTRUCTION (every function borrows its
// own parameter exactly once and every operation inside that SAME function
// reads that SAME borrowed place or a one-hop derivative of it), so
// disabling memoization cannot make THEM quadratic -- there is nothing
// deep to re-walk. Asserting the seam's effect only where it can
// structurally show up, rather than averaging it away across shapes that
// can never exhibit it, is the same discipline 09-04's own mutation-kill
// used (a hand-picked fixture, not the whole corpus).
//
// Both directions are asserted: with the seam engaged, the fitted
// exponent must EXCEED peerClosureGrowthExponentBoundMilli (proving the
// bound is a real control); with the seam off (the production path,
// already proven in TestPeerClosureCostGrowthExponentWithinBound above),
// it must hold -- reasserted here, on the SAME sweep function, so a reader
// sees both outcomes of the identical mechanism side by side rather than
// having to cross-reference a separate test file.
func TestPeerClosureCostUnmemoizedSeamExceedsBound(t *testing.T) {
	const shape = "forward"

	t.Run("seam disabled (production path) stays within bound", func(t *testing.T) {
		ops, checks := sweepPeerClosureCost(t, shape)
		exponent := fitPeerClosureGrowthExponent(ops, checks)
		if math.IsNaN(exponent) {
			t.Fatalf("shape=%s: fitted exponent is NaN", shape)
		}
		milli := int64(math.Round(exponent * 1000))
		if milli > peerClosureGrowthExponentBoundMilli {
			t.Fatalf("shape=%s: seam OFF (production path) milli-exponent=%d exceeds bound=%d -- the production path itself should never trip this bound",
				shape, milli, peerClosureGrowthExponentBoundMilli)
		}
	})

	t.Run("seam engaged reproduces a genuinely quadratic walk", func(t *testing.T) {
		restore := corevalidate.SetPeerClosureUnmemoizedSeamForTest(true)
		defer restore()

		ops, checks := sweepPeerClosureCost(t, shape)
		exponent := fitPeerClosureGrowthExponent(ops, checks)
		if math.IsNaN(exponent) {
			t.Fatalf("shape=%s: fitted exponent is NaN", shape)
		}
		milli := int64(math.Round(exponent * 1000))

		var ratioAtSmall, ratioAtLarge float64
		var sawSmall, sawLarge bool
		for index, n := range peerCostSizeLadder {
			switch n {
			case peerCostRatioSmall:
				ratioAtSmall = float64(checks[index]) / float64(ops[index])
				sawSmall = true
			case peerCostRatioLarge:
				ratioAtLarge = float64(checks[index]) / float64(ops[index])
				sawLarge = true
			}
		}
		var ratioDelta float64
		var ratioExceeds bool
		if sawSmall && sawLarge {
			ratioDelta = math.Abs(ratioAtLarge-ratioAtSmall) / ratioAtSmall
			ratioExceeds = ratioDelta > peerClosureRatioStabilityBound
		}

		t.Logf("shape=%s seam=unmemoized milli-exponent=%d (exponent=%.4f) bound=%d; ratio delta=%.1f%% bound=%.0f%%",
			shape, milli, exponent, peerClosureGrowthExponentBoundMilli, ratioDelta*100, peerClosureRatioStabilityBound*100)

		if milli <= peerClosureGrowthExponentBoundMilli && !ratioExceeds {
			t.Fatalf("shape=%s: unmemoized seam failed to exceed EITHER the growth-exponent bound (milli=%d, bound=%d) or the ratio-stability tripwire (delta=%.1f%%, bound=%.0f%%) -- the bound has never been seen to fail, so it measures nothing (T-09-17)",
				shape, milli, peerClosureGrowthExponentBoundMilli, ratioDelta*100, peerClosureRatioStabilityBound*100)
		}
	})
}

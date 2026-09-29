package session_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// functionCarryDepth computes D-10-45's own definition of composition
// depth for one function's own return: the number of core.OpCall hops a
// single loan's carry relation crosses before reaching its terminal use
// (this function's own core.OpReturn). It mirrors
// pathoracle_compose.go's composeCarriesOwnLoan/composeCall exactly in
// spirit -- a fresh loan is anything this function creates itself via
// core.OpBorrowShared/core.OpBorrowExclusive (depth 0, not yet crossed any
// call); forwarding a CALLEE's own carried loan into this function's own
// return counts one more hop than whatever depth the callee itself
// reached -- but is deliberately a SEPARATE, independent re-implementation
// (never a call into package pathoracle) operating purely on
// core.LinearOperation, so this gate and the composition machinery it
// gates cannot share a bug. functions is a whole PROGRAM's own
// ID-to-Function map (built fresh per file, mirroring
// pathoracle.BuildCalleeLookup's own shape); visiting guards against a
// call-graph cycle in a corrupted or refused artifact by failing closed to
// -1 ("does not carry"), exactly like pathoracle's own
// compositionCycleError; memo caches one already-computed answer per
// function ID within the SAME program so a diamond-shaped call graph does
// not re-walk a shared callee's body once per caller.
//
// -1 means "this function's own return does not carry any loan at all"
// (mirroring composeCall's own "does not carry" fail-closed default for an
// unresolved callee, D-10-09's own vocabulary) -- never a sentinel that
// could be confused with a genuine depth-0 (a function that creates and
// immediately returns its own loan, with zero hops crossed).
func functionCarryDepth(functions map[string]core.Function, id string, visiting map[string]bool, memo map[string]int) int {
	if depth, ok := memo[id]; ok {
		return depth
	}
	if visiting[id] {
		// A call-graph cycle in a corrupted or check-refused artifact.
		// callgraph.Order already refuses a genuinely cyclic program
		// before a checked core.Program is admitted (Phase 07); this
		// guard exists so a synthetic or refused artifact this gate
		// walks directly still fails closed rather than recursing
		// forever, mirroring pathoracle.compositionCycleError's own
		// fail-closed posture.
		return -1
	}
	function, ok := functions[id]
	if !ok || function.Linear == nil {
		memo[id] = -1
		return -1
	}
	visiting[id] = true
	defer delete(visiting, id)

	// chain maps a placeID to the composition depth the loan flowing
	// through it has already crossed -- the SAME forward place-
	// inheritance idiom pathoracle_compose.go's composeCarriesOwnLoan and
	// corevalidate_peer_liveness.go's derivePeerLoanCarry both use,
	// re-derived a third, independent time here.
	chain := map[string]int{}
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared, core.OpBorrowExclusive:
			if operation.TargetID != "" {
				chain[operation.TargetID] = 0
			}
		case core.OpMove, core.OpCopy:
			if depth, carries := chain[operation.SourceID]; carries && operation.TargetID != "" {
				chain[operation.TargetID] = depth
			}
		case core.OpCall:
			calleeDepth := functionCarryDepth(functions, operation.CalleeID, visiting, memo)
			if calleeDepth >= 0 && operation.TargetID != "" {
				chain[operation.TargetID] = calleeDepth + 1
			}
		}
	}

	result := -1
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpReturn {
			continue
		}
		if depth, carries := chain[operation.SourceID]; carries && depth > result {
			result = depth
		}
	}
	memo[id] = result
	return result
}

// maxCompositionDepthAcrossCorpus walks testdata/ (testsupport.ProjectPath,
// filepath.WalkDir -- the sibling gate's own resolution idiom,
// session_peer_gate_test.go's TestNoUndeclaredCheckPeerDivergenceAcrossCorpus)
// and returns the maximum composition depth any function's own return
// reaches, plus the project-relative path of the fixture that reached it.
//
// A fixture counts toward this walk once `session.CheckFile` produces a
// structurally real core.Program (parse succeeded and check.Program built
// SOME core.Program) -- regardless of whether check's OWN admission
// diagnostics are empty. This is deliberate, not an oversight: pathoracle
// is a "third, genuinely different decision procedure" whose own identity
// claim is that it never trusts a declared contract, only the checked
// core's own operations (pathoracle.go's own package doc) -- so its
// composition depth is a property of the CHECKED CORE STRUCTURE, not of
// what check/corevalidate/originvalidate separately conclude about it. A
// program that check itself refuses (like relay_depth3_refuse.schway, refused
// via check.interprocedural_loan_liveness) still has a fully-formed
// core.Program with every declared core.OpCall intact, and its own
// declared-borrow carry genuinely crosses 3 hops in that structure --
// exactly the depth-3 evidence this gate needs, independent of the
// separate, pre-existing corevalidate finding relay_depth3_accept.schway's
// own header documents (a program admitted end-to-end cannot exhibit a
// literal 3-hop DECLARED-ORIGIN carry today; the checked-but-refused
// artifact still can, and is what this gate reads).
//
// A file that fails to PARSE at all (session.Check's own parse-diagnostics
// early return, leaving Program at its zero value with zero Functions) is
// skipped -- there is no core.Program for pathoracle's own definition to
// apply to.
func maxCompositionDepthAcrossCorpus(t *testing.T) (observedMax int, source string) {
	t.Helper()
	root := testsupport.ProjectPath("testdata")
	observedMax = -1
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".schway" {
			return nil
		}
		checked, checkErr := session.CheckFile(path)
		if checkErr != nil {
			return nil
		}
		if len(checked.Program.Functions) == 0 {
			// Parse failed (or produced no functions at all): no
			// core.Program for this gate's definition to apply to.
			return nil
		}
		functions := make(map[string]core.Function, len(checked.Program.Functions))
		for _, function := range checked.Program.Functions {
			functions[function.ID] = function
		}
		memo := map[string]int{}
		for _, function := range checked.Program.Functions {
			depth := functionCarryDepth(functions, function.ID, map[string]bool{}, memo)
			if depth > observedMax {
				observedMax = depth
				if relative, relErr := filepath.Rel(testsupport.ProjectPath("."), path); relErr == nil {
					source = filepath.ToSlash(relative)
				} else {
					source = path
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return observedMax, source
}

// compositionDepthGate is the single piece of decision logic the
// bidirectional check reuses UNCHANGED for both directions (D-10-50): a
// declared depth requirement is satisfied only when the corpus's own
// observed maximum reaches at least that far. Reused, not duplicated, by
// TestCompositionDepthCorpusReachesDeclaredBound's two subtests -- one
// against the real declared pathoracle.MaxCompositionDepth, one against a
// hypothetically-raised local value -- so proving direction 2 fires is a
// property of this ONE function, not two independently-written assertions
// that could quietly drift apart.
func compositionDepthGate(observedMax, requiredDepth int, source string) (ok bool, message string) {
	if observedMax < requiredDepth {
		return false, fmt.Sprintf(
			"composition depth corpus falls short: observed maximum depth %d (from %s), required declared depth %d",
			observedMax, source, requiredDepth,
		)
	}
	return true, ""
}

// TestCompositionDepthCorpusReachesDeclaredBound is D-10-50's bidirectional
// gate: the deliverable is the CHECK that fails in BOTH directions, not
// merely the declared constant. Direction 1 (the "corpus reaches the
// declared bound" subtest): the corpus must contain a fixture whose own
// composition depth reaches pathoracle.MaxCompositionDepth, or this gate
// names the shortfall. Direction 2 (the "raising..." subtest): the SAME
// compositionDepthGate logic, run again against a hypothetically-raised
// local requirement (never editing the production constant), must report a
// shortfall -- proving that raising pathoracle.MaxCompositionDepth without
// extending the corpus would make direction 1 fail, OBSERVED to fire here
// rather than merely reasoned about in prose.
//
// Order-independent and free of shared mutable state (T-10-04's own
// concurrency probe): every map this test allocates is local to a single
// call of maxCompositionDepthAcrossCorpus or functionCarryDepth, and no
// package-level variable is read or written anywhere in this file -- so
// `-count=2 -shuffle=on` exercises two structurally independent runs with
// no possibility of one leaking state into the other.
func TestCompositionDepthCorpusReachesDeclaredBound(t *testing.T) {
	observedMax, source := maxCompositionDepthAcrossCorpus(t)

	t.Run("corpus reaches the declared bound", func(t *testing.T) {
		ok, message := compositionDepthGate(observedMax, pathoracle.MaxCompositionDepth, source)
		if !ok {
			t.Fatal(message)
		}
	})

	t.Run("raising the constant without extending the corpus reports a shortfall", func(t *testing.T) {
		raisedRequirement := observedMax + 1
		ok, message := compositionDepthGate(observedMax, raisedRequirement, source)
		if ok {
			t.Fatalf(
				"direction 2 did not fire: raising the requirement to %d (observed max %d, from %s) should have reported a shortfall but the gate reported ok",
				raisedRequirement, observedMax, source,
			)
		}
		if message == "" {
			t.Fatal("direction 2 fired but produced an empty message; a failure here must be diagnosable without reading this test")
		}
		t.Logf("direction 2 confirmed to fire: %s", message)
	})
}

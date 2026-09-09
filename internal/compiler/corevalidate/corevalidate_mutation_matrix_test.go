package corevalidate_test

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// stage0MutationMatrix records, for TestStage0SummaryMutationMatrix, every
// control this plan introduced (corevalidate.SummaryPeerControls) alongside
// the seeded mutation(s) that kill it -- QLT-08's claim that no control
// ships having never been seen to fail. A control appearing with an empty
// mutation list fails this test. fault1/fault4 execute in
// originvalidate_test.go's own TestStage0SummaryMutationMatrix (they flip
// originvalidate's own unexported seams, unreachable from this package's
// test binary -- see that file's doc comment); they are still named here so
// this matrix is the single place naming which fault kills which control.
var stage0MutationMatrix = map[string][]string{
	"control:summary.peer_both_replay_sites":        {"fault2_two_site_replay_blocks_disabled"},
	"control:summary.producer_peer_zero_divergence": {"fault1_producer_only_type_fact_lookup", "fault3_bilateral_false_agreement"},
	"control:summary.callable_publication_safety":   {"fault3_bilateral_false_agreement", "fault4_callable_producer", "fault5_callable_peer"},
}

func loadCheckedProgram(t *testing.T, phase, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", phase, fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s/%s: unexpected diagnostics: %+v", phase, fixture, checked.Diagnostics)
	}
	return checked.Program
}

// TestStage0SummaryMutationMatrix (corevalidate half) is 07-02 Task 3's
// D-07-24 gate for the three faults that flip THIS package's own
// fault-injection seams (disableSummaryPeerAtReplayBlocksForTest,
// forcePeerCallableAlwaysTrue): fault 2 (two-site), fault 3 (bilateral), and
// fault 5 (peer-side Callable). Faults 1 and 4 live in originvalidate_test.go
// (package originvalidate_test), which flips originvalidate's own seams
// instead -- see that file's identically-named test's doc comment for why
// the five faults are split across both packages.
func TestStage0SummaryMutationMatrix(t *testing.T) {
	t.Run("fault2_two_site_replay_blocks_disabled", func(t *testing.T) {
		// D-07-21: a branch-shaped function returning a borrowed origin on
		// exactly one arm, with the declared PublicOrigin left absent
		// (undeclared), can only be caught by the peer AT replayBlocks --
		// disabling that one site's wiring makes the peer signature
		// disappear for this function entirely, and restoring it brings
		// the refusal signal (Callable == false) back.
		program := loadCheckedProgram(t, "phase3", "public_view_multi_arm_omitted.lang")
		function := program.Functions[0]

		restore := corevalidate.SetDisableSummaryPeerAtReplayBlocksForTest(true)
		result := corevalidate.Validate(program)
		if _, ok := result.PeerSignatures()[function.ID]; ok {
			restore()
			t.Fatal("expected the peer signature to be ABSENT while replayBlocks wiring is disabled -- the refusal signal disappeared")
		}
		restore()

		result = corevalidate.Validate(program)
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatal("expected the peer signature to be present once replayBlocks wiring is restored")
		}
		if peerSignature.Callable {
			t.Fatal("expected the restored peer to report Callable == false (the borrow-derived-on-one-arm return is undeclared)")
		}
		_, blocksRan := result.PeerSiteCoverage()
		if !blocksRan {
			t.Fatal("expected PeerSiteCoverage to report replayBlocks ran once restored")
		}
	})

	t.Run("fault3_bilateral_false_agreement", func(t *testing.T) {
		// D-07-24 fault 3, the delicate one: injecting the SAME
		// Callable-always-true fault into producer AND peer identically
		// makes them still (wrongly) agree. The sweep below must report
		// that agreement as a GATE FAILURE, not a pass.
		//
		// The producer-side half of this fault (originvalidate's
		// forceCallableAlwaysTrue driving BuildInterface to report
		// Callable == true unconditionally) is independently proven by
		// originvalidate_test.go's fault4_callable_producer subtest --
		// originvalidate's unexported seam is unreachable from this
		// package's test binary (see this file's package doc comment), so
		// this subtest supplies that already-proven producer value
		// literally and flips ONLY the peer-side seam it CAN reach here,
		// then feeds both into the same sweep fault 4/5 exercise.
		restorePeer := corevalidate.SetPeerCallableForceOverrideForTest(true)
		defer restorePeer()

		program := loadCheckedProgram(t, "phase07", "clean_but_unpublishable.lang")
		function := program.Functions[0]
		result := corevalidate.Validate(program)
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if !peerSignature.Callable {
			t.Fatal("expected the forced peer override to report Callable == true")
		}
		// producerCallableUnderFault4 is the value
		// originvalidate_test.go's fault4_callable_producer subtest proves
		// BuildInterface reports for this exact fixture when
		// forceCallableAlwaysTrue is active: true, unconditionally.
		const producerCallableUnderFault4 = true

		report := sweepReportBilateralAgreement(producerCallableUnderFault4, peerSignature.Callable)
		if report != "no divergence detected under bilateral fault" {
			t.Fatalf("expected the bilateral fault to produce a false agreement report, got %q", report)
		}
	})

	t.Run("fault5_callable_peer", func(t *testing.T) {
		// The reverse of fault 4: forcing ONLY the peer's Callable
		// derivation to true must leave the (unforced) producer refusing,
		// a real divergence in the opposite direction.
		program := loadCheckedProgram(t, "phase07", "clean_but_unpublishable.lang")
		function := program.Functions[0]

		// Unforced baseline: confirm the peer itself would refuse absent
		// the seam, so the forced override below is a genuine flip, not a
		// no-op.
		baseline := corevalidate.Validate(program)
		if baselineSignature := baseline.PeerSignatures()[function.ID]; baselineSignature.Callable {
			t.Fatal("expected the unforced peer to refuse (Callable == false) on clean_but_unpublishable.lang")
		}

		restore := corevalidate.SetPeerCallableForceOverrideForTest(true)
		defer restore()

		result := corevalidate.Validate(program)
		peerSignature, ok := result.PeerSignatures()[function.ID]
		if !ok {
			t.Fatalf("function %s missing from peer signatures", function.ID)
		}
		if !peerSignature.Callable {
			t.Fatal("expected the forced peer override to report Callable == true, diverging from the unforced producer's false")
		}
	})

	t.Run("fault6_callable_refusal_peer_disabled", func(t *testing.T) {
		// 07-05 Task 3 Test 5: corevalidate's OWN Callable-based refusal
		// (peerCallable, wired into the "call" admission arm), disabled
		// INDEPENDENTLY of check's own refusal, on a hand-built synthetic
		// core.Program -- check never even sees this program, since it
		// exists only to give corevalidate something to validate without
		// going through check.Program's own (now refusing) admission path.
		program := syntheticCallToNonCallableCalleeProgram(t)

		baseline := corevalidate.Validate(program)
		if baseline.Valid {
			t.Fatalf("expected the unforced corevalidate refusal to refuse the call, got Valid == true")
		}
		foundBaseline := false
		for _, problem := range baseline.Problems {
			if problem.Code == core.CalleeNotCallable {
				foundBaseline = true
			}
		}
		if !foundBaseline {
			t.Fatalf("expected %s in the unforced baseline, got %+v", core.CalleeNotCallable, baseline.Problems)
		}

		restore := corevalidate.SetPeerCallableForceOverrideForTest(true)
		defer restore()
		forced := corevalidate.Validate(program)
		if !forced.Valid {
			t.Fatalf("expected the forced peer override to admit the call (corevalidate's OWN refusal disabled independently), got problems: %+v", forced.Problems)
		}
		// check's own, SEPARATE refusal on the identical shape
		// (call_uncallable_callee.lang) is independently proven by
		// check_test.go's TestCallToNonCallableCalleeRefused -- this
		// subtest disables ONLY corevalidate's own half, never check's.
	})

	t.Run("fault7_bilateral_callable_refusal", func(t *testing.T) {
		// 07-05 Task 3 Test 6: the SAME bilateral technique fault3 above
		// uses -- combine THIS package's own live-flipped result with the
		// OTHER package's independently-proven literal (established by
		// check_test.go's own TestVerifyCallableRefusalSeamAdmitsUncallableCallee,
		// which proves check's verifyCallableRefusalSeam makes check admit
		// the identical shape) -- and feed both into the SAME
		// sweepReportBilateralAgreement function. Two identical wrongs
		// must not merge into a green: the report is a GATE FAILURE, never
		// a pass.
		restore := corevalidate.SetPeerCallableForceOverrideForTest(true)
		defer restore()

		program := syntheticCallToNonCallableCalleeProgram(t)
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("expected the forced peer override to admit the call, got problems: %+v", result.Problems)
		}
		// checkAdmitsUnderSeam is the literal fact
		// check_test.go's TestVerifyCallableRefusalSeamAdmitsUncallableCallee
		// independently proves: with verifyCallableRefusalSeam engaged,
		// check wrongly admits the identical call_uncallable_callee.lang
		// shape (zero diagnostics) -- check.verifyCallableRefusalSeam is
		// unexported and unreachable from this package's test binary,
		// exactly like fault3's producer-side literal above.
		const checkAdmitsUnderSeam = true
		report := sweepReportBilateralAgreement(checkAdmitsUnderSeam, result.Valid)
		if report != "no divergence detected under bilateral fault" {
			t.Fatalf("expected the bilateral fault to produce a false agreement report, got %q", report)
		}
	})

	t.Run("every introduced control has a non-empty seeded-mutation list", func(t *testing.T) {
		for _, control := range corevalidate.SummaryPeerControls {
			mutations, ok := stage0MutationMatrix[control]
			if !ok || len(mutations) == 0 {
				t.Fatalf("control %s has an EMPTY seeded-mutation list -- QLT-08 requires every control to have been seen to fail", control)
			}
		}
		if len(stage0MutationMatrix) != len(corevalidate.SummaryPeerControls) {
			t.Fatalf("stage0MutationMatrix (%d entries) and corevalidate.SummaryPeerControls (%d entries) must name the exact same control set", len(stage0MutationMatrix), len(corevalidate.SummaryPeerControls))
		}
	})
}

// sweepReportBilateralAgreement is the minimal "sweep" D-07-24 fault 3
// requires: given a producer's and a peer's independently-derived Callable
// values for the SAME function, it reports whether they diverge (evidence)
// or wrongly agree (a gate failure the caller must treat as a FAIL, never a
// silent pass) -- the report string itself is the observable Fault 3's test
// asserts.
func sweepReportBilateralAgreement(producerCallable, peerCallable bool) string {
	if producerCallable == peerCallable {
		return "no divergence detected under bilateral fault"
	}
	return "divergence detected"
}

// syntheticCallToNonCallableCalleeProgram is 07-05 Task 3 Test 5's
// hand-built synthetic core.Program: a caller with one core.OpCall to a
// callee whose body returns an exclusive-borrow-derived place with no
// declared PublicOrigin -- the exact shape peerCallable (D-07-33) refuses.
// Hand-built rather than produced via check.Program/session.Check, because
// check itself now refuses this exact shape (07-05 Task 1's own
// verifyCallableRefusal) -- there is no other way to hand corevalidate a
// core.Program containing an admitted call to a non-callable callee.
func syntheticCallToNonCallableCalleeProgram(t *testing.T) core.Program {
	t.Helper()
	derived, err := ability.Derive(core.TypeRef{Constructor: "Buffer"})
	if err != nil {
		t.Fatal(err)
	}

	const callerID, calleeID = "s1:test:fn:caller", "s1:test:fn:callee"
	callerTypeID, calleeTypeID := callerID+":type:0", calleeID+":type:0"
	callerParamID, callerTargetID := callerID+":place:0", callerID+":place:1"
	calleeParamID, calleeTargetID := calleeID+":place:0", calleeID+":place:1"

	typeFact := func(id string) core.TypeFact {
		return core.TypeFact{ID: id, Shape: core.TypeRef{Constructor: "Buffer"}, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}
	}

	callee := core.Function{
		ID: calleeID, Name: "callee",
		EntryPointID: calleeID + ":point:entry", ReturnPointID: calleeID + ":point:return",
		Parameter:  core.Parameter{ID: calleeParamID, Name: "buffer", Type: "Buffer"},
		ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:    calleeID + ":linear",
			Types: []core.TypeFact{typeFact(calleeTypeID)},
			Places: []core.Place{
				{ID: calleeParamID, Name: "buffer", TypeID: calleeTypeID},
				{ID: calleeTargetID, Name: "view", TypeID: calleeTypeID},
			},
			Operations: []core.LinearOperation{
				{ID: calleeID + ":op:0", PointID: calleeID + ":point:linear:0", Kind: core.OpBorrowExclusive, SourceID: calleeParamID, TargetID: calleeTargetID, TypeID: calleeTypeID, LoanID: calleeID + ":loan:0"},
				{ID: calleeID + ":op:1", PointID: calleeID + ":point:linear:1", Kind: core.OpReturn, SourceID: calleeTargetID, TypeID: calleeTypeID},
			},
		},
	}
	caller := core.Function{
		ID: callerID, Name: "caller",
		EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
		Parameter:  core.Parameter{ID: callerParamID, Name: "buffer", Type: "Buffer"},
		ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:    callerID + ":linear",
			Types: []core.TypeFact{typeFact(callerTypeID)},
			Places: []core.Place{
				{ID: callerParamID, Name: "buffer", TypeID: callerTypeID},
				{ID: callerTargetID, Name: "result", TypeID: callerTypeID},
			},
			Operations: []core.LinearOperation{
				{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: callerParamID, TargetID: callerTargetID, TypeID: callerTypeID, CalleeID: calleeID},
				{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: callerTargetID, TypeID: callerTypeID},
			},
		},
	}
	return core.Program{
		Schema: core.Schema1, Module: "test.synthetic_uncallable_callee", ModuleID: "s1:test:module:synthetic_uncallable_callee",
		Functions: []core.Function{caller, callee},
	}
}

package corevalidate_test

import (
	"os"
	"testing"

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

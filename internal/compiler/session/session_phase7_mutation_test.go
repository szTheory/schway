package session_test

import (
	"context"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

// TestPhase7DispatchControlsMutationKilledPhase07Lane is Task 3 Test 2
// (D-07-41/D-07-42, QLT-08): the phase07 lane's OWN kill, distinct from
// core_test.go's in-process control kill (which has its own,
// package-internal TestPhase7DispatchControlsMutationKilled) -- A-05
// requires each control to carry its own literal fixture list and its own
// proof, never sharing one implementation two independent controls could
// both silently rot alongside.
//
// The four-beat body (pathoracle_test.go:268-295's precedent): assert
// clean, save/override/defer-restore, assert an observable effect, assert
// the specific code.
func TestPhase7DispatchControlsMutationKilledPhase07Lane(t *testing.T) {
	// Beat 1: assert clean. The real lane, with its real fixtures and its
	// real required-kinds check, passes and fires all three Phase 07
	// controls (already covered by TestVerifyPhase7ControlsAndWork above;
	// re-asserted here as this test's own "before" baseline).
	clean, err := session.VerifyPhase7ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase7ControlsAndWork (clean) returned an error: %v", err)
	}
	if clean.Status != protocol.StatusPass {
		t.Fatalf("VerifyPhase7ControlsAndWork (clean) status = %s, want pass", clean.Status)
	}

	// Beat 2: override the lane's fixture list to an OpCall-free fixture
	// (testdata/phase07/clean_but_unpublishable.schway, checked-clean but
	// carrying no core.OpCall at all), restored via defer.
	restoreFixtures := session.SetPhase07LaneDispatchFixturesForTest([]string{"clean_but_unpublishable.schway"})
	defer restoreFixtures()

	// Beat 3: assert an observable effect. With the real required-kinds
	// check still in force (core.OpCall required) but the corpus now
	// unable to produce it, the lane MUST fail.
	withoutCallFixture, err := session.VerifyPhase7ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase7ControlsAndWork (OpCall-free corpus) returned an error: %v", err)
	}
	if withoutCallFixture.Status == protocol.StatusPass {
		t.Fatal("expected the phase07 lane to fail once its corpus contains no core.OpCall-producing fixture, got pass")
	}

	// Beat 4: assert the specific code -- the exact verify.control_missing
	// diagnostic naming the phase07 lane's own control, never a bare
	// non-nil-status check.
	foundControlMissing := false
	for _, diag := range withoutCallFixture.Diagnostics {
		if diag.Code == "verify.control_missing" && strings.Contains(diag.Message, "control:kind.exhaustive_dispatch.phase07_lane") {
			foundControlMissing = true
		}
	}
	if !foundControlMissing {
		t.Fatalf("expected a verify.control_missing diagnostic naming control:kind.exhaustive_dispatch.phase07_lane, got diagnostics=%+v", withoutCallFixture.Diagnostics)
	}

	// Now the mutation: with the required-kinds seam ALSO overridden to
	// exclude core.OpCall, the SAME OpCall-free corpus passes -- proving
	// the required-kinds loop, not the per-fixture site calls above it, is
	// what makes this lane load-bearing (the phase07-lane analog of
	// Test 1's in-process kill).
	restoreKinds := session.SetPhase07LaneRequiredKindsForTest([]core.OperationKind{})
	defer restoreKinds()
	mutated, err := session.VerifyPhase7ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase7ControlsAndWork (mutated) returned an error: %v", err)
	}
	if mutated.Status != protocol.StatusPass {
		t.Fatalf("expected the mutated phase07 lane (core.OpCall excluded from required kinds) to pass on the OpCall-free corpus, got status=%s diagnostics=%+v", mutated.Status, mutated.Diagnostics)
	}
}

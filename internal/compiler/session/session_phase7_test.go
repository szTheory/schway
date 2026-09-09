package session_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestVerifyPhase7ControlsAndWork runs Phase 07's own control-and-work
// gate end-to-end and asserts every required control fires with nonzero
// RecomputedWork, mirroring TestVerifyPhase5ControlsAndWork's /
// TestVerifyPhase6ControlsAndWork's own shape.
func TestVerifyPhase7ControlsAndWork(t *testing.T) {
	result, err := session.VerifyPhase7ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase7ControlsAndWork returned an error: %v", err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("VerifyPhase7ControlsAndWork status = %s, want pass; lanes=%+v diagnostics=%+v", result.Status, result.Lanes, result.Diagnostics)
	}
	for _, required := range session.Phase7RequiredControls() {
		found := false
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("VerifyPhase7ControlsAndWork never fired required control %s", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			t.Fatalf("lane %s reported zero RecomputedWork", lane.ID)
		}
	}
}

// phase7VerifierScriptText reads scripts/verify-phase7.sh's own text,
// mirroring phase5VerifierScriptText's / phase6VerifierScriptText's
// precedent exactly.
func phase7VerifierScriptText(t testing.TB) string {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase7.sh"))
	if err != nil {
		t.Fatalf("read scripts/verify-phase7.sh: %v", err)
	}
	return string(data)
}

// phase7OwnControlIdentifiers extracts the exact set of `control:` tokens
// from the script's OWN required-control block (the `for control in ...`
// loop following the "Phase 07's own required-control set" comment),
// mirroring phase5OwnControlIdentifiers's / phase6OwnControlIdentifiers's
// anchor-plus-block-extraction technique exactly.
func phase7OwnControlIdentifiers(t testing.TB, text string) []string {
	t.Helper()
	anchor := "Phase 07's own required-control set"
	anchorIndex := strings.Index(text, anchor)
	if anchorIndex == -1 {
		t.Fatalf("scripts/verify-phase7.sh is missing its own required-control-set comment anchor")
	}
	rest := text[anchorIndex:]
	doneIndex := strings.Index(rest, "\ndone")
	if doneIndex == -1 {
		t.Fatalf("scripts/verify-phase7.sh's own required-control block has no closing done")
	}
	block := rest[:doneIndex]
	var identifiers []string
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "\\"))
		if strings.HasPrefix(trimmed, "control:") {
			identifiers = append(identifiers, trimmed)
		}
	}
	return identifiers
}

// TestPhase7RequiredControlsMatchScript asserts set equality, in both
// directions, between session.Phase7RequiredControls() and the
// identifiers named in scripts/verify-phase7.sh's own required-control
// block (T-07-24/D-07-41): a control added to one and forgotten in the
// other fails here rather than silently drifting apart.
func TestPhase7RequiredControlsMatchScript(t *testing.T) {
	text := phase7VerifierScriptText(t)
	fromScript := phase7OwnControlIdentifiers(t, text)
	scriptSet := make(map[string]bool, len(fromScript))
	for _, control := range fromScript {
		scriptSet[control] = true
	}
	sessionSet := make(map[string]bool, len(session.Phase7RequiredControls()))
	for _, control := range session.Phase7RequiredControls() {
		sessionSet[control] = true
	}
	for control := range sessionSet {
		if !scriptSet[control] {
			t.Fatalf("session.Phase7RequiredControls() names %s, which the script's own required-control block does not", control)
		}
	}
	for control := range scriptSet {
		if !sessionSet[control] {
			t.Fatalf("the script's required-control block names %s, which session.Phase7RequiredControls() does not", control)
		}
	}
	if len(scriptSet) != len(sessionSet) {
		t.Fatalf("script control set (%d) and session control set (%d) differ in size: script=%v session=%v", len(scriptSet), len(sessionSet), fromScript, session.Phase7RequiredControls())
	}
}

// controlsWithRecordedMutationKill is Task 3 Test 5's own authoritative
// kill registry (D-07-41): one entry per control in
// session.Phase7RequiredControls(), each already proven load-bearing by a
// named seeded-mutation test IN THE PLAN THAT INTRODUCED IT -- see each
// Control* identifier's own doc comment in session_phase7.go for the
// exact test name. This is the completeness claim's other half: not a
// self-authored table that could omit a control from both this list and
// Phase7RequiredControls() without either going red, but a list compared
// against that authoritative function by EXACT set equality below.
var controlsWithRecordedMutationKill = []string{
	// 07-04: core_test.go's TestPhase7DispatchControlsMutationKilled,
	// session_phase7_mutation_test.go's own lane kill, and interp/cgen's
	// TestOpCallGroupedArmMutationKilled.
	session.ControlKindExhaustiveDispatchPhase07InProcess,
	session.ControlKindExhaustiveDispatchPhase07Lane,
	session.ControlDispatchRecognizedNotExecuted,
	// 07-05: check_test.go's TestCallAdmissionBodyBlindControl and
	// TestVerifyCallableRefusalSeamAdmitsUncallableCallee /
	// corevalidate's fault6/fault7 subtests.
	session.ControlCallAdmissionBodyBlind,
	session.ControlCallCallableRefusal,
	// 07-06: callgraph_test.go's TestCallGraphMutationMatrix subtests and
	// TestRotationRemovalMutationKilled / TestWitnessSelectionRemovalMutationKilled.
	session.ControlCallGraphGrayReentry,
	session.ControlCallGraphSelfEdge,
	session.ControlCallGraphUnresolvedEdgeRefused,
	session.ControlCallGraphCycleIDDeterministic,
	session.ControlCallGraphCauseBound,
	// 07-07: corevalidate_cycle_peer_test.go's
	// TestCyclePeerMutationMatrix and check_test.go's
	// TestCheckCycleRefusalIndependentOfCorevalidatePeer /
	// TestForeignSymbolShadowingFixtureRefusedAndEdgeMutationKilled, and
	// this file's own TestPhase7ControlsAreMutationKilled (below).
	session.ControlCorevalidateCyclePeerIndependent,
	session.ControlCorevalidateCyclePeerGrayReentry,
	session.ControlCallGraphForeignShadowingEdgePreserved,
	session.ControlPhase07ControlsAreMutationKilled,
}

// phase7ControlsExactSetEqual is the completeness check's own comparison
// primitive (D-07-41): returns an error unless required and killed name
// EXACTLY the same set of controls, sorted so the error message is
// deterministic. Exercised directly, with FABRICATED inputs, by
// TestPhase7ControlsCompletenessMetaMutation below -- proving the
// comparison itself is load-bearing, not merely present: a control added
// to the required list with no recorded kill must fail, and a kill
// recorded for a control absent from the required list must also fail.
func phase7ControlsExactSetEqual(required, killed []string) error {
	requiredSet := make(map[string]bool, len(required))
	for _, control := range required {
		requiredSet[control] = true
	}
	killedSet := make(map[string]bool, len(killed))
	for _, control := range killed {
		killedSet[control] = true
	}
	var missingKill, extraKill []string
	for control := range requiredSet {
		if !killedSet[control] {
			missingKill = append(missingKill, control)
		}
	}
	for control := range killedSet {
		if !requiredSet[control] {
			extraKill = append(extraKill, control)
		}
	}
	sort.Strings(missingKill)
	sort.Strings(extraKill)
	if len(missingKill) != 0 || len(extraKill) != 0 {
		return fmt.Errorf("phase07 control completeness mismatch: required-but-unkilled=%v killed-but-not-required=%v", missingKill, extraKill)
	}
	return nil
}

// TestPhase7ControlsAreMutationKilled is Task 3 Test 5: the phase-wide
// completeness matrix, derived from session.Phase7RequiredControls() (the
// authoritative list, itself held in exact set equality with
// scripts/verify-phase7.sh by TestPhase7RequiredControlsMatchScript above)
// and compared by EXACT set equality against controlsWithRecordedMutationKill.
// Every control introduced across 07-01 through 07-07 must appear in both,
// with a non-empty seeded-mutation kill -- an empty mutation set is
// non-compliance, not a vacuous pass (QLT-08).
func TestPhase7ControlsAreMutationKilled(t *testing.T) {
	if err := phase7ControlsExactSetEqual(session.Phase7RequiredControls(), controlsWithRecordedMutationKill); err != nil {
		t.Fatal(err)
	}
}

// TestPhase7ControlsCompletenessMetaMutation is Task 3 Test 5's own
// meta-mutation proof: the completeness comparison itself is falsifiable.
// Uses FABRICATED control lists, never touching production state, so this
// test cannot itself corrupt session.Phase7RequiredControls() for any
// sibling test.
func TestPhase7ControlsCompletenessMetaMutation(t *testing.T) {
	t.Run("a required control with no recorded kill fails completeness", func(t *testing.T) {
		required := append(append([]string{}, session.Phase7RequiredControls()...), "control:phase07.fake_new_control_no_kill")
		if err := phase7ControlsExactSetEqual(required, controlsWithRecordedMutationKill); err == nil {
			t.Fatal("mutation (adding an unkilled control) had no observable effect: expected completeness to fail")
		}
	})
	t.Run("a recorded kill for a control absent from the required list fails completeness", func(t *testing.T) {
		killed := append(append([]string{}, controlsWithRecordedMutationKill...), "control:phase07.fake_extra_kill_not_required")
		if err := phase7ControlsExactSetEqual(session.Phase7RequiredControls(), killed); err == nil {
			t.Fatal("mutation (recording an extra kill) had no observable effect: expected completeness to fail")
		}
	})
}

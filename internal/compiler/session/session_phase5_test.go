package session_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPhase5RequiredControlsMatchExportedConstants guards against literal
// spelling drift between Phase5RequiredControls()'s own flat string list
// (grep target for scripts/verify-phase5.sh's parity block) and the
// exported control-ID constants used to fire those same controls
// elsewhere in this package (ControlAliasFalseNoAlias,
// ControlDiagnosticRejectProgramIDEquivalence, the four sanitize
// controls).
func TestPhase5RequiredControlsMatchExportedConstants(t *testing.T) {
	named := map[string]bool{
		session.ControlAliasFalseNoAlias:                    true,
		session.ControlDiagnosticRejectProgramIDEquivalence: true,
		session.ControlSanitizeRetainedPointer:              true,
		session.ControlSanitizeUBSanNoRecover:               true,
		session.ControlSanitizeAllocatorMismatch:            true,
		session.ControlSanitizeUseAfterFree:                 true,
	}
	for literal := range named {
		found := false
		for _, required := range session.Phase5RequiredControls() {
			if required == literal {
				found = true
			}
		}
		if !found {
			t.Fatalf("exported control constant %q is not present, byte-identical, in Phase5RequiredControls()", literal)
		}
	}
}

func TestVerifyPhase5ControlsAndWork(t *testing.T) {
	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("expected a pass result, got status=%q diagnostics=%+v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}
	for _, required := range session.Phase5RequiredControls() {
		found := false
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("required control %q never fired in any lane: %+v", required, result.Lanes)
		}
	}
}

func TestPhase5ControlsAllHaveNonzeroWork(t *testing.T) {
	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("expected a pass result, got status=%q lanes=%+v", result.Status, result.Lanes)
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			t.Fatalf("lane %s fired with zero RecomputedWork -- a control that fires with zero counted work is exactly as unproven as one that never fired", lane.ID)
		}
	}
}

// TestPhase5ToolMissingIsNamedObligationNotPass is T-05-33's own
// falsifier: with a broken ClangPath, every sanitizer control must still
// appear as a named, gate-visible operational obligation -- never omitted,
// and never rendered as a pass.
func TestPhase5ToolMissingIsNamedObligationNotPass(t *testing.T) {
	original := session.Phase5ClangPathOverrideForTest
	session.Phase5ClangPathOverrideForTest = testsupport.ProjectPath("testdata", "does-not-exist", "clang-broken")
	defer func() { session.Phase5ClangPathOverrideForTest = original }()

	result, err := session.VerifyPhase5ControlsAndWork(context.Background())
	if err != nil {
		t.Fatalf("VerifyPhase5ControlsAndWork returned an unexpected hard error for a broken toolchain: %v", err)
	}
	if result.Status == "pass" {
		t.Fatal("expected a non-pass result with a broken ClangPath, got a pass")
	}
	for _, required := range []string{
		session.ControlSanitizeRetainedPointer,
		session.ControlSanitizeUseAfterFree,
		session.ControlSanitizeAllocatorMismatch,
		session.ControlSanitizeUBSanNoRecover,
	} {
		found := false
		var laneStatus string
		for _, lane := range result.Lanes {
			for _, control := range lane.Controls {
				if control == required {
					found = true
					laneStatus = lane.Status
				}
			}
		}
		if !found {
			t.Fatalf("sanitizer control %q was not named as an obligation with a broken toolchain: %+v", required, result.Lanes)
		}
		if laneStatus == "pass" {
			t.Fatalf("sanitizer control %q rendered as a pass with a broken toolchain", required)
		}
	}
}

// TestSanitizeLaneNotInEditOrCheck asserts D-05-17's cost-placement rule
// at the CLI dispatch layer: neither the `check` command (session.Check/
// session.CheckCommandFile, wired from cmd/lang/main.go's runCheck) nor
// any edit-loop path references the sanitizer lane or the Phase 5 gate
// that invokes it.
func TestSanitizeLaneNotInEditOrCheck(t *testing.T) {
	mainSource, err := os.ReadFile(testsupport.ProjectPath("cmd", "lang", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"VerifyPhase5SanitizeLane", "VerifyPhase5ControlsAndWork"} {
		if strings.Contains(string(mainSource), forbidden) {
			t.Fatalf("cmd/lang/main.go must not reference %s outside the verify/release path", forbidden)
		}
	}

	checkSource, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session.go"))
	if err != nil {
		t.Fatal(err)
	}
	// session.go declares Check/CheckCommandFile (the `check` pipeline) and
	// must not itself reference the sanitizer lane or the Phase 5 gate --
	// those live only in the new session_phase5*.go files this plan adds,
	// which `check` never imports into its own call graph.
	for _, forbidden := range []string{"VerifyPhase5SanitizeLane", "VerifyPhase5ControlsAndWork"} {
		if strings.Contains(string(checkSource), forbidden) {
			t.Fatalf("session.go must not reference %s", forbidden)
		}
	}
}

// TestNAT03MutationsCiteExistingPrograms is T-05-36's own falsifier,
// enumerated over the WHOLE table rather than sampled: every row with
// Subjected: true must cite a CorpusProgram path that exists on disk,
// including the two rows plan 05-07 left marked PENDING-05-08 (now closed
// by plan 05-08's fixtures landing).
func TestNAT03MutationsCiteExistingPrograms(t *testing.T) {
	rows := session.NAT03Mutations()
	subjected := 0
	for _, row := range rows {
		if !row.Subjected {
			continue
		}
		subjected++
		path := testsupport.ProjectPath(row.CorpusProgram)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("row %s cites a CorpusProgram that does not exist on disk: %s (%v)", row.ControlID, row.CorpusProgram, err)
		}
	}
	if subjected != 6 {
		t.Fatalf("expected exactly 6 subjected rows, got %d", subjected)
	}
}

// TestNAT03SanitizerRowMovesItsClaimedAxis closes D-05-22 for the row
// plan 05-07 could not assert (control:native.sanitize.allocator_mismatch,
// PENDING-05-08 until this plan): Phase5AssertMutationMovesAnAxis succeeds
// against plan 05-08's testdata/phase5/allocator_mismatch.lang, asserting
// the row's own claimed axis specifically.
func TestNAT03SanitizerRowMovesItsClaimedAxis(t *testing.T) {
	var target *session.NAT03Mutation
	for _, row := range session.NAT03Mutations() {
		row := row
		if row.ControlID == session.ControlSanitizeAllocatorMismatch {
			target = &row
		}
	}
	if target == nil {
		t.Fatal("no NAT-03 row declares control:native.sanitize.allocator_mismatch")
	}
	if !target.Subjected {
		t.Fatalf("expected the allocator-mismatch row to be Subjected: true, got %+v", target)
	}
	if err := session.Phase5AssertMutationMovesAnAxis(context.Background(), *target); err != nil {
		t.Fatalf("allocator-mismatch row did not move its claimed axis: %v", err)
	}
}

// TestNoNAT03RowRemainsPending refuses a gate whose citation is unclosed:
// no NAT-03 row's OWN declaration in session_phase5_alias.go may still
// carry plan 05-07's pending marker comment now that plan 05-08's fixtures
// exist. Checks for the exact marker-comment line, not any prose mention
// of the string PENDING-05-08 elsewhere in the file (e.g. a stale error
// message), matching the plan's own "delete exactly the two marker
// comments, nothing else" scope.
func TestNoNAT03RowRemainsPending(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase5_alias.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(source), "\n") {
		if strings.TrimSpace(line) == "// PENDING-05-08" {
			t.Fatalf("session_phase5_alias.go still carries a PENDING-05-08 marker comment: %q", line)
		}
	}
}

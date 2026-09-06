package session_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// coordinatedLieSourcePath/coordinatedLieCorePath name D-05-30's adversarial
// artifact pair (plan 05-13): an ordinary, checker-accepted Lang source
// program and a HAND-AUTHORED core artifact naming the exact same
// module/function identity but claiming a different ownership-transfer
// outcome. Neither file is derived from the other -- that absence of a real
// translation step is the whole point of the demonstration.
func coordinatedLieSourcePath() string {
	return testsupport.ProjectPath("testdata", "phase5", "coordinated_lie.lang")
}

func coordinatedLieCorePath() string {
	return testsupport.ProjectPath("testdata", "phase5", "coordinated_lie.core.json")
}

// loadCoordinatedLieCore reads and decodes the hand-authored core artifact.
// It performs no validation of its own -- callers run corevalidate.Validate
// explicitly, exactly as they would for any other independently-produced
// core.Program.
func loadCoordinatedLieCore(t *testing.T) core.Program {
	t.Helper()
	raw, err := os.ReadFile(coordinatedLieCorePath())
	if err != nil {
		t.Fatalf("reading coordinated_lie.core.json: %v", err)
	}
	var program core.Program
	if err := json.Unmarshal(raw, &program); err != nil {
		t.Fatalf("decoding coordinated_lie.core.json: %v", err)
	}
	return program
}

// checkCoordinatedLieSource runs the SOURCE half of the pair through
// check+originvalidate exactly as any other Lang program would be admitted,
// returning the checker-derived core.Program (the source's own honest
// claim).
func checkCoordinatedLieSource(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(coordinatedLieSourcePath())
	if err != nil {
		t.Fatalf("reading coordinated_lie.lang: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("coordinated_lie.lang: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("coordinated_lie.lang: originvalidate.ValidatePublished rejected the source-derived program: %+v", problems)
	}
	return checked.Program
}

// TestCoordinatedLieArtifactsBothValidate is D-05-30's reachability
// demonstration: coordinated_lie.lang independently passes check and
// originvalidate.ValidatePublished on its own terms, coordinated_lie.core.json
// independently passes corevalidate.Validate on ITS own terms, and the two
// artifacts -- which name the identical module/function identity -- assert
// DIFFERENT ownership-transfer claims about the same buffer parameter: the
// source moves it (full ownership transfer, never returning access to the
// original place), the hand-authored core only ever shares it (a borrow,
// never a move). Neither validator ever compares one artifact against the
// other; asserting that disagreement explicitly here is the whole point of
// the demonstration.
func TestCoordinatedLieArtifactsBothValidate(t *testing.T) {
	sourceProgram := checkCoordinatedLieSource(t)

	coreProgram := loadCoordinatedLieCore(t)
	validated := corevalidate.Validate(coreProgram)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected the hand-authored coordinated_lie.core.json on its own terms: %+v", validated.Problems)
	}

	if len(sourceProgram.Functions) != 1 || sourceProgram.Functions[0].Linear == nil || len(sourceProgram.Functions[0].Linear.Operations) == 0 {
		t.Fatalf("coordinated_lie.lang: expected exactly one straight-line function with at least one operation, got %+v", sourceProgram.Functions)
	}
	if len(coreProgram.Functions) != 1 || coreProgram.Functions[0].Linear == nil || len(coreProgram.Functions[0].Linear.Operations) == 0 {
		t.Fatalf("coordinated_lie.core.json: expected exactly one straight-line function with at least one operation, got %+v", coreProgram.Functions)
	}

	sourceModuleID := sourceProgram.ModuleID
	coreModuleID := coreProgram.ModuleID
	if sourceModuleID != coreModuleID {
		t.Fatalf("expected the pair to name the SAME module identity (that is what makes this a coordinated lie, not just two unrelated programs): source=%q core=%q", sourceModuleID, coreModuleID)
	}

	sourceClaim := sourceProgram.Functions[0].Linear.Operations[0].Kind
	coreClaim := coreProgram.Functions[0].Linear.Operations[0].Kind
	if sourceClaim != core.OpMove {
		t.Fatalf("expected coordinated_lie.lang's own checker-derived claim to be a move (full ownership transfer), got %q", sourceClaim)
	}
	if coreClaim != core.OpBorrowShared {
		t.Fatalf("expected coordinated_lie.core.json's hand-authored claim to be a shared borrow (never a move), got %q", coreClaim)
	}
	if sourceClaim == coreClaim {
		t.Fatalf("expected the source-derived claim and the hand-authored core claim to DIFFER -- both were %q, which is not a coordinated lie, just an accurate translation", sourceClaim)
	}
}

// TestCoordinatedLieEscapeIsDeclared asserts EscapeCoordinatedSourceToCoreFalseClaim
// appears in VerifyCoordinatedLieEscape's own returned LaneResult.ExpectedEscapes
// -- the Phase 5 expected-escape set THIS PLAN establishes (plan 05-14 owns
// folding it into the shipped session.Phase5ExpectedEscapes(), a dependency
// recorded in this plan's summary and NOT performed here, to keep the Go
// control/escape sets and the shell gate script from ever seeing a
// transiently divergent pair).
func TestCoordinatedLieEscapeIsDeclared(t *testing.T) {
	result, err := session.VerifyCoordinatedLieEscape(context.Background())
	if err != nil {
		t.Fatalf("VerifyCoordinatedLieEscape: %v", err)
	}
	declared := false
	for _, escape := range result.ExpectedEscapes {
		if escape == session.EscapeCoordinatedSourceToCoreFalseClaim {
			declared = true
		}
	}
	if !declared {
		t.Fatalf("expected %s in VerifyCoordinatedLieEscape's ExpectedEscapes, got %+v", session.EscapeCoordinatedSourceToCoreFalseClaim, result.ExpectedEscapes)
	}
}

// TestCoordinatedLieEscapeIsNeverDetected asserts the escape identifier
// never appears in any phase's shipped required-control set
// (session.AllShippedControlIDs(), the union of Phase4RequiredControls()
// and Phase5RequiredControls() -- qlt01.go's own single source of truth for
// "any control this repository currently ships") so it can never be
// claimed as covered by a control that happens to share its name.
func TestCoordinatedLieEscapeIsNeverDetected(t *testing.T) {
	for _, control := range session.AllShippedControlIDs() {
		if control == session.EscapeCoordinatedSourceToCoreFalseClaim {
			t.Fatalf("escape %s must never appear in the shipped required-control set, but it does", session.EscapeCoordinatedSourceToCoreFalseClaim)
		}
	}
}

// TestCoordinatedLiePassesTheGateUnderTheNamedEscape is D-05-30's
// demonstration proper: the gate passes the adversarial pair (no error, a
// "pass" status, zero fired controls) AND the result explicitly names the
// escape it attributes that pass to.
func TestCoordinatedLiePassesTheGateUnderTheNamedEscape(t *testing.T) {
	result, err := session.VerifyCoordinatedLieEscape(context.Background())
	if err != nil {
		t.Fatalf("expected the gate to pass the adversarial pair, got error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("expected LaneResult.Status = %q, got %q", "pass", result.Status)
	}
	if len(result.Fired) != 0 {
		t.Fatalf("expected no control to have fired on the adversarial pair, got %+v", result.Fired)
	}
	if result.RecomputedWork == 0 {
		t.Fatal("expected nonzero RecomputedWork -- a lane that inspects nothing cannot claim to have verified anything")
	}
	found := false
	for _, escape := range result.ExpectedEscapes {
		if escape == session.EscapeCoordinatedSourceToCoreFalseClaim {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected the pass to be attributed to %s, got ExpectedEscapes=%+v", session.EscapeCoordinatedSourceToCoreFalseClaim, result.ExpectedEscapes)
	}
}

// TestBothPhase5EscapesAreVisible asserts both Phase 5 escapes -- plan
// 05-08's escape:callback-invocation-unsubjected (already shipped in
// session.Phase5ExpectedEscapes()) and this plan's
// escape:coordinated-source-to-core-false-claim (declared by
// VerifyCoordinatedLieEscape, not yet folded into
// session.Phase5ExpectedEscapes() -- that fold is plan 05-14's job) --
// appear together in the union this test constructs, and that NEITHER
// appears in session.AllShippedControlIDs(). This documents the full set
// plan 05-14 will ship without prematurely editing session_phase5.go here.
func TestBothPhase5EscapesAreVisible(t *testing.T) {
	escapes := append([]string{}, session.Phase5ExpectedEscapes()...)
	escapes = append(escapes, session.EscapeCoordinatedSourceToCoreFalseClaim)

	wantCallback := false
	wantCoordinatedLie := false
	for _, escape := range escapes {
		if escape == session.EscapeCallbackInvocationUnsubjected {
			wantCallback = true
		}
		if escape == session.EscapeCoordinatedSourceToCoreFalseClaim {
			wantCoordinatedLie = true
		}
	}
	if !wantCallback || !wantCoordinatedLie {
		t.Fatalf("expected both Phase 5 escapes present, got %+v", escapes)
	}

	shipped := make(map[string]bool, len(session.AllShippedControlIDs()))
	for _, control := range session.AllShippedControlIDs() {
		shipped[control] = true
	}
	for _, escape := range escapes {
		if shipped[escape] {
			t.Fatalf("escape %s must never appear in the shipped required-control set", escape)
		}
	}
}

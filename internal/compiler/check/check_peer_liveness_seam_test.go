package check

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// TestPeerLoanCarrySeamDisabledCheckStillRefuses is Phase 09's own D-09-24/
// D-09-25 companion-assertion test, the NEW direction (a corevalidate-side
// seam, check asserted still-refuses): a seeded fault that merely makes two
// peers diverge proves nothing on its own -- that is precisely the false
// positive Knight & Leveson (1986) documented for N-version programming,
// where independently written implementations correlate on faults because
// the spec, the requirements understanding, or (here) a shared code path is
// what actually failed, not genuine independent agreement. A seam inside
// code BOTH peers depend on would also produce a divergence, so divergence
// alone is not evidence of independence.
//
// The gate this test proves is therefore: seed a fault in corevalidate's
// OWN peer-derivation ONLY (SetDisablePeerLoanCarryConsultForTest, an
// unexported package-level seam corevalidate_peer_liveness.go declares --
// unreachable from this package by construction, D-09-25's own "the seam's
// very scoping is the proof the code paths are separate" argument), and
// assert THREE things: (1) check's own independently-derived answer for
// relay_escort_witness.schway -- a wholly unrelated program from a wholly
// unrelated code path -- is completely UNCHANGED by a fault seeded entirely
// inside the OTHER peer; (2) with the seam ON, check and corevalidate now
// DISAGREE on testdata/phase08/twin_a_accept.schway (check still admits it;
// corevalidate, its own consult disabled, reverts to its pre-Phase-09
// unconditional-propagation shape and wrongly refuses it again); (3) with
// the seam OFF, the two agree on that same program. Only the combination of
// (2) and (3) -- disagreement WITH the seam, agreement WITHOUT it, on the
// SAME program -- is evidence the peer's own consult is what was doing the
// work; (1) alone rules out a shared-code correlated failure.
func TestPeerLoanCarrySeamDisabledCheckStillRefuses(t *testing.T) {
	// (1) check's own independent answer, unrelated program, unaffected by
	// a fault seeded entirely inside corevalidate's own package.
	restore := corevalidate.SetDisablePeerLoanCarryConsultForTest(true)
	witnessResult := Program(mustParseProgram(t, readPhase07Fixture(t, "relay_escort_witness.schway")))
	restore()
	if len(witnessResult.Diagnostics) != 1 || witnessResult.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected check to still refuse relay_escort_witness.schway with check.interprocedural_loan_liveness regardless of corevalidate's own seam, got %+v", witnessResult.Diagnostics)
	}

	twinSource := readPhase08Fixture(t, "twin_a_accept.schway")

	// (2) seam ON: check admits, corevalidate now wrongly refuses again --
	// the two DISAGREE.
	restore = corevalidate.SetDisablePeerLoanCarryConsultForTest(true)
	twinResult := Program(mustParseProgram(t, twinSource))
	if len(twinResult.Diagnostics) != 0 {
		restore()
		t.Fatalf("expected check to admit twin_a_accept.schway, got %+v", twinResult.Diagnostics)
	}
	seamOnCoreResult := corevalidate.Validate(twinResult.Program)
	restore()
	if seamOnCoreResult.Valid {
		t.Fatal("expected corevalidate to wrongly refuse twin_a_accept.schway again with its own consult seam disabled, got Valid == true")
	}
	found := false
	for _, problem := range seamOnCoreResult.Problems {
		if problem.Code == "core.move_while_borrowed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected core.move_while_borrowed with the seam engaged, got %+v", seamOnCoreResult.Problems)
	}

	// (3) seam OFF: check admits, corevalidate independently agrees.
	twinResultAgain := Program(mustParseProgram(t, twinSource))
	if len(twinResultAgain.Diagnostics) != 0 {
		t.Fatalf("expected check to admit twin_a_accept.schway, got %+v", twinResultAgain.Diagnostics)
	}
	seamOffCoreResult := corevalidate.Validate(twinResultAgain.Program)
	if !seamOffCoreResult.Valid {
		t.Fatalf("expected corevalidate to independently agree with the seam OFF, got Problems %+v", seamOffCoreResult.Problems)
	}
}

// TestInterproceduralLivenessSeamCheckDisabledCorevalidateStillRefuses is
// D-09-24's EXISTING direction, generalized from Callable
// (TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses,
// check_test.go) to the interprocedural liveness fact: with check's OWN
// admission disabled (disableInterproceduralLoanLivenessForTest engaged),
// relay_escort_witness.schway is admitted by check -- but the resulting
// core.Program is INDEPENDENTLY still refused by corevalidate's own
// peer-derived loan-carry consult (corevalidate_peer_liveness.go), which
// never consults check's summaries, its seam, or its diagnostics at all.
func TestInterproceduralLivenessSeamCheckDisabledCorevalidateStillRefuses(t *testing.T) {
	defer func() { disableInterproceduralLoanLivenessForTest = false }()
	disableInterproceduralLoanLivenessForTest = true
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "relay_escort_witness.schway")))
	disableInterproceduralLoanLivenessForTest = false
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected check's seam to admit relay_escort_witness.schway, got %+v", result.Diagnostics)
	}
	coreResult := corevalidate.Validate(result.Program)
	if coreResult.Valid {
		t.Fatal("expected corevalidate to independently still refuse, got Valid == true")
	}
	found := false
	for _, problem := range coreResult.Problems {
		if problem.Code == "core.move_while_borrowed" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected core.move_while_borrowed, got %+v", coreResult.Problems)
	}
}

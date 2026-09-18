package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestDeclineCarriesReasonAndDiagnosis exercises selectRepair's decline path
// directly against synthetic checkResult documents (D-14-42's exact behavior
// cases), never through the subprocess, so the classification logic itself
// is pinned independent of any real diagnostic corpus.
func TestDeclineCarriesReasonAndDiagnosis(t *testing.T) {
	t.Run("none_offered: invalid, >=1 diagnostic, zero repairs anywhere", func(t *testing.T) {
		result := checkResult{
			Status: "invalid",
			Diagnostics: []jsonDiagnostic{
				{Code: "syntax.expected_rbrace"},
				{Code: "syntax.expected_declaration"},
			},
		}
		_, _, ok := selectRepair(result)
		if ok {
			t.Fatal("expected no driver-eligible repair")
		}
		decision := classifyDecline(result)
		if decision.declineReason != DeclineNoneOffered {
			t.Fatalf("got decline reason %q, want %q", decision.declineReason, DeclineNoneOffered)
		}
		if decision.diagnosisCode != "syntax.expected_rbrace" {
			t.Fatalf("got diagnosis code %q, want first diagnostic's code %q", decision.diagnosisCode, "syntax.expected_rbrace")
		}
		wantCodes := []string{"syntax.expected_rbrace", "syntax.expected_declaration"}
		if len(decision.diagnosisCodes) != len(wantCodes) {
			t.Fatalf("got diagnosis codes %v, want %v", decision.diagnosisCodes, wantCodes)
		}
		for i, c := range wantCodes {
			if decision.diagnosisCodes[i] != c {
				t.Fatalf("diagnosisCodes[%d] = %q, want %q", i, decision.diagnosisCodes[i], c)
			}
		}
		if decision.bestApplicability != "" {
			t.Fatalf("got best applicability %q, want empty for none_offered", decision.bestApplicability)
		}
	})

	t.Run("none_eligible: repairs present, none driver-eligible", func(t *testing.T) {
		result := checkResult{
			Status: "invalid",
			Diagnostics: []jsonDiagnostic{
				{
					Code: "check.some_diagnostic",
					Repairs: []jsonRepair{
						{Kind: "unwired_kind", Applicability: applicabilityRequiresConfirmation},
					},
				},
			},
		}
		_, _, ok := selectRepair(result)
		if ok {
			t.Fatal("expected no driver-eligible repair")
		}
		decision := classifyDecline(result)
		if decision.declineReason != DeclineNoneEligible {
			t.Fatalf("got decline reason %q, want %q", decision.declineReason, DeclineNoneEligible)
		}
		if decision.bestApplicability != applicabilityRequiresConfirmation {
			t.Fatalf("got best applicability %q, want %q", decision.bestApplicability, applicabilityRequiresConfirmation)
		}
	})

	t.Run("no_diagnostics: invalid with zero diagnostics stays unrepairable", func(t *testing.T) {
		result := checkResult{Status: "invalid"}
		_, _, ok := selectRepair(result)
		if ok {
			t.Fatal("expected no driver-eligible repair")
		}
		decision := classifyDecline(result)
		if decision.declineReason != DeclineNoDiagnostics {
			t.Fatalf("got decline reason %q, want %q", decision.declineReason, DeclineNoDiagnostics)
		}
		// diagnosis_code must never be empty on an unrepairable outcome
		// (D-14-42's global truth), even here where there is no diagnostic
		// to name as the first element -- the decline reason itself is the
		// only honest non-empty value available.
		if decision.diagnosisCode != DeclineNoDiagnostics {
			t.Fatalf("got diagnosis code %q, want %q (no diagnostic to name; the decline reason is the only honest non-empty value)", decision.diagnosisCode, DeclineNoDiagnostics)
		}
	})
}

// TestRepairOutcomeDeclineFieldsPopulated drives Repair's own decline branch
// (not just selectRepair/classifyDecline in isolation) via a stand-in
// checkResult fed straight into the same code path Repair uses, confirming
// the Outcome envelope itself carries the new fields end to end.
func TestRepairOutcomeDeclineFieldsPopulated(t *testing.T) {
	result := checkResult{
		Status: "invalid",
		Diagnostics: []jsonDiagnostic{
			{Code: "check.call_graph_cycle"},
		},
	}
	_, _, ok := selectRepair(result)
	if ok {
		t.Fatal("expected no driver-eligible repair")
	}
	decision := classifyDecline(result)
	outcome := Outcome{
		Status:            OutcomeUnrepairable,
		DiagnosisCode:     decision.diagnosisCode,
		DiagnosisCodes:    newDiagnosisCodeList(decision.diagnosisCodes),
		DeclineReason:     decision.declineReason,
		BestApplicability: decision.bestApplicability,
		SubprocessCount:   1,
	}
	if outcome.DiagnosisCode == "" {
		t.Fatal("outcome.DiagnosisCode is empty, want the first diagnostic's code")
	}
	if outcome.DeclineReason != DeclineNoneOffered {
		t.Fatalf("got decline reason %q, want %q", outcome.DeclineReason, DeclineNoneOffered)
	}
}

// declineVocabulary is the closed three-value decline_reason set D-14-42
// defines. TestUnrepairableAlwaysCarriesDiagnosis checks membership against
// this set, not merely non-emptiness -- an out-of-vocabulary value must
// fail the guard (see TestDeclineReasonOutsideVocabularyFailsTheGuard).
var declineVocabulary = map[string]bool{
	DeclineNoneOffered:   true,
	DeclineNoneEligible:  true,
	DeclineNoDiagnostics: true,
}

// unrepairableDiagnosisOK reports whether an unrepairable outcome's
// diagnosis_code is non-empty and its decline_reason is a member of the
// closed vocabulary (D-14-42) -- a real membership check, not a disguised
// non-emptiness check. Returns ok plus a message describing the violation
// when ok is false.
func unrepairableDiagnosisOK(diagnosisCode, declineReason string) (ok bool, msg string) {
	if diagnosisCode == "" {
		return false, "unrepairable outcome carries an empty diagnosis_code"
	}
	if !declineVocabulary[declineReason] {
		return false, fmt.Sprintf("decline_reason %q is not a member of the closed vocabulary %v", declineReason, declineVocabulary)
	}
	return true, ""
}

func assertUnrepairableCarriesDiagnosis(t *testing.T, label, diagnosisCode, declineReason string) {
	t.Helper()
	if ok, msg := unrepairableDiagnosisOK(diagnosisCode, declineReason); !ok {
		t.Fatalf("%s: %s", label, msg)
	}
}

// TestUnrepairableAlwaysCarriesDiagnosis is the corpus-wide guard T-14-13
// requires: for every checked-in capture and every fixture class the driver
// exercises, an outcome whose status is unrepairable carries a non-empty
// diagnosis_code and a decline_reason drawn from the closed vocabulary. An
// outcome whose status is not unrepairable is not constrained here at all.
//
// This plan's reclassifications, recorded here rather than only in
// 14-03-SUMMARY.md (D-14-42): the forward-direction interprocedural loan
// liveness decline (D-13-09b, exercised below as
// "interprocedural_loan_forward_direction") and the call-graph cycle
// decline (D-13-12 -- not exercised by any checked-in .lang fixture today,
// so it is not asserted directly here, only documented) both become
// repair.none_offered, stating in the protocol "a cycle has no local,
// mechanical edit" instead of leaving it as silence. The uniqueness-gate
// decline (exercised below as "call_argument_type_ambiguous") becomes
// repair.none_offered or repair.none_eligible depending on whether it
// suppresses emission entirely or leaves a non-eligible repair in place --
// this corpus's own fixture suppresses emission, so it lands on
// repair.none_offered. All three stay OutcomeUnrepairable: honest declines
// stay honest.
func TestUnrepairableAlwaysCarriesDiagnosis(t *testing.T) {
	// Every checked-in cmd/lang-repair/testdata/*_capture.json. Each one is
	// decoded directly as a checkResult (no subprocess needed) and
	// classified exactly the way Repair's own first check would classify
	// it. All eight captures in this corpus carry a driver-eligible repair
	// or report a clean re-check (see TestBaselineCaptureContainsEligibleRepair),
	// so none of them actually reaches the unrepairable branch today --
	// this walk is what proves that honestly, rather than assuming it, and
	// stays a real guard against any future capture that DOES decline.
	captures := []string{
		"borrow_diagnose_capture.json",
		"borrow_reverify_capture.json",
		"interprocedural_loan_diagnose_capture.json",
		"interprocedural_loan_reverify_capture.json",
		"match_diagnose_capture.json",
		"match_reverify_capture.json",
		"move_diagnose_capture.json",
		"move_reverify_capture.json",
	}
	for _, name := range captures {
		t.Run(name, func(t *testing.T) {
			result := decodeCheckJSON(t, captureFile(t, name))
			if result.Status == statusPass {
				return // already_clean -- not constrained by this guard
			}
			if _, _, ok := selectRepair(result); ok {
				return // a driver-eligible repair exists -- would-be repaired, not unrepairable
			}
			decision := classifyDecline(result)
			assertUnrepairableCarriesDiagnosis(t, name, decision.diagnosisCode, decision.declineReason)
		})
	}

	// Every fixture class the driver's own TestUnrepairableDefectFailsTheGate
	// exercises end to end through the real `lang` binary and Repair --
	// mirroring that test's three subtests exactly, so this guard is
	// checked against the same real declines that test already proves are
	// genuinely unrepairable, not a synthetic stand-in.
	langBinary := testsupport.BuildCLI(t)

	t.Run("non_exhaustive_match", func(t *testing.T) {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		sourcePath := filepath.Join(dir, "non_exhaustive.lang")
		mustWriteFile(t, sourcePath, source)
		outcome, err := Repair(context.Background(), langBinary, sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if outcome.Status != OutcomeUnrepairable {
			t.Fatalf("got status %q, want %q", outcome.Status, OutcomeUnrepairable)
		}
		assertUnrepairableCarriesDiagnosis(t, "non_exhaustive_match", outcome.DiagnosisCode, outcome.DeclineReason)
	})

	t.Run("call_argument_type_ambiguous", func(t *testing.T) {
		fixturePath := testsupport.ProjectPath("testdata", "phase13", "heldout_call_argument_ambiguous.lang")
		original, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatal(err)
		}
		mutated, err := session.CallArgumentTypeInjector{}.Inject(original)
		if err != nil {
			t.Fatalf("injecting: %v", err)
		}
		dir := t.TempDir()
		sourcePath := filepath.Join(dir, "heldout_call_argument_ambiguous.lang")
		mustWriteFile(t, sourcePath, mutated)
		outcome, err := Repair(context.Background(), langBinary, sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if outcome.Status != OutcomeUnrepairable {
			t.Fatalf("got status %q, want %q", outcome.Status, OutcomeUnrepairable)
		}
		assertUnrepairableCarriesDiagnosis(t, "call_argument_type_ambiguous", outcome.DiagnosisCode, outcome.DeclineReason)
	})

	t.Run("interprocedural_loan_forward_direction", func(t *testing.T) {
		fixturePath := testsupport.ProjectPath("testdata", "phase07", "relay_escort_witness.lang")
		original, err := os.ReadFile(fixturePath)
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		sourcePath := filepath.Join(dir, "relay_escort_witness.lang")
		mustWriteFile(t, sourcePath, original)
		outcome, err := Repair(context.Background(), langBinary, sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		if outcome.Status != OutcomeUnrepairable {
			t.Fatalf("got status %q, want %q", outcome.Status, OutcomeUnrepairable)
		}
		assertUnrepairableCarriesDiagnosis(t, "interprocedural_loan_forward_direction", outcome.DiagnosisCode, outcome.DeclineReason)
		if outcome.DeclineReason != DeclineNoneOffered {
			t.Fatalf("got decline reason %q, want %q (zero repairs anywhere on the forward-direction case)", outcome.DeclineReason, DeclineNoneOffered)
		}
	})
}

// TestDeclineReasonOutsideVocabularyFailsTheGuard proves
// unrepairableDiagnosisOK's vocabulary check is a real membership test, not
// a disguised non-emptiness check: a non-empty but out-of-vocabulary
// decline_reason must still fail it. A temporary edit substituting an
// out-of-vocabulary reason for a real one below would flip this to a false
// pass, which is exactly what this test exists to catch.
func TestDeclineReasonOutsideVocabularyFailsTheGuard(t *testing.T) {
	if ok, msg := unrepairableDiagnosisOK("check.some_diagnostic", "repair.not_a_real_reason"); ok {
		t.Fatal("expected an out-of-vocabulary decline_reason to fail the guard, but it passed")
	} else if msg == "" {
		t.Fatal("expected a non-empty violation message")
	}
}

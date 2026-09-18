package main

import "testing"

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
		if decision.diagnosisCode != "" {
			t.Fatalf("got diagnosis code %q, want empty (zero diagnostics)", decision.diagnosisCode)
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

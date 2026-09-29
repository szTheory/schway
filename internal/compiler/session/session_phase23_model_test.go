package session_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase23ModelReportUsesModelOnlyEvidenceVocabulary(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase22", "identity.schway"))
	if err != nil {
		t.Fatal(err)
	}
	cases := session.ReplayCases{
		Schema: session.ReplayCasesSchema,
		Cases: []session.ReplayCase{{
			ID:        "independent-modeled-byte-41",
			Kind:      session.ReplayCaseVerifierModel,
			Operation: "fixture.value",
			Expected:  session.ReplayExpected{Kind: execution.OutcomeReturned, Value: "65"},
			ForeignOutcomes: []session.ReplayForeignOutcome{{
				Operation: "fixture.value", Type: "U64", Value: "65",
			}},
		}},
	}
	report, err := session.VerifyApplicationCases(context.Background(), source, cases, native.Runner{})
	if err != nil {
		t.Fatalf("verify model case: %v", err)
	}
	if !report.Verified || len(report.Cases) != 1 || report.Cases[0].Status != session.ReplayStatusPass {
		t.Fatalf("model report did not pass: %+v", report)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{`"evidence_scope":"verifier_model_only"`, `"actual_host_io":false`, `"physical_cleanup":false`} {
		if !strings.Contains(string(encoded), required) {
			t.Errorf("model report omits required evidence boundary %s: %s", required, encoded)
		}
	}
}

func TestPhase23SourceReplayRefusesLocalCAndPathIO(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase23", "file_byte.schway"))
	if err != nil {
		t.Fatal(err)
	}
	cases := session.ReplayCases{
		Schema: session.ReplayCasesSchema,
		Cases: []session.ReplayCase{{
			ID: "ordinary-source-replay", Kind: session.ReplayCaseSource, Input: "65",
			Expected: session.ReplayExpected{Kind: execution.OutcomeReturned, Value: "65"}, ForeignOutcomes: []session.ReplayForeignOutcome{},
		}},
	}
	report, err := session.VerifyApplicationCases(context.Background(), source, cases, native.Runner{})
	if err != nil {
		t.Fatalf("local-C replay should be a scoped refusal in the report: %v", err)
	}
	caseReport := report.Cases[0]
	if caseReport.Status != session.ReplayStatusUnsupported || caseReport.ActualHostIO || caseReport.PhysicalCleanup {
		t.Fatalf("source replay crossed the model/native boundary: %+v", caseReport)
	}
	if !strings.Contains(caseReport.Diagnostic, "U64-to-U64") {
		t.Fatalf("source replay refusal does not identify the unsupported path entry: %q", caseReport.Diagnostic)
	}
}

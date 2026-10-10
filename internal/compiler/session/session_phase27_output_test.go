package session_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27OutputEvidenceSeparation(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "fizzbuzz.schway")
	want, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "fizzbuzz-15.out"))
	if err != nil {
		t.Fatal(err)
	}
	outputLimitedArtifact := filepath.Join(t.TempDir(), "fizzbuzz-output-limited")
	func() {
		restoreOutputLimit := cgen.SetApplicationOutputLimitForTest(4)
		defer restoreOutputLimit()
		_, diagnostics, buildErr := session.BuildApplicationFile(context.Background(), sourcePath, outputLimitedArtifact, native.DefaultRunner())
		if buildErr != nil || len(diagnostics) != 0 {
			t.Fatalf("BuildApplicationFile with lowered output limit diagnostics=%+v err=%v", diagnostics, buildErr)
		}
	}()
	outputReportPath := filepath.Join(t.TempDir(), "fizzbuzz-output-failure-evidence.json")
	var outputStdout, outputStderr bytes.Buffer
	outputOutcome, outputReport, err := native.RunApplicationWithEvidence(context.Background(), outputLimitedArtifact, "3", outputReportPath, native.EvidenceEvents, &outputStdout, &outputStderr)
	// The native app stops at the first failed write (like the interpreter's
	// terminal defect), so it must not publish a success-shaped capture: the
	// evidence is honestly incomplete while the process outcome stays exit 65.
	var outputToolError *native.ToolError
	if !errors.As(err, &outputToolError) || outputToolError.Code != "native.evidence_incomplete" || outputOutcome.Kind != native.RunExited || outputOutcome.ExitCode != 65 || outputReport.ProcessOutcome != outputOutcome || outputReport.CaptureStatus != native.EvidenceStatusIncomplete || outputReport.Execution != nil || outputReport.Verified || !bytes.Equal(outputStdout.Bytes(), []byte("1\n2\n")) || outputStderr.String() != "schway: application output limit exceeded\n" {
		t.Fatalf("output-limited app reported success-shaped evidence: outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outputOutcome, outputReport, outputStdout.Bytes(), outputStderr.String(), err)
	}

	evidenceLimitedArtifact := filepath.Join(t.TempDir(), "fizzbuzz-evidence-limited")
	func() {
		restoreEvidenceLimit := cgen.SetApplicationEvidenceOutputLimitForTest(4)
		defer restoreEvidenceLimit()
		_, diagnostics, buildErr := session.BuildApplicationFile(context.Background(), sourcePath, evidenceLimitedArtifact, native.DefaultRunner())
		if buildErr != nil || len(diagnostics) != 0 {
			t.Fatalf("BuildApplicationFile with lowered evidence limit diagnostics=%+v err=%v", diagnostics, buildErr)
		}
	}()

	reportPath := filepath.Join(t.TempDir(), "fizzbuzz-evidence.json")
	var stdout, stderr bytes.Buffer
	outcome, report, err := native.RunApplicationWithEvidence(context.Background(), evidenceLimitedArtifact, "15", reportPath, native.EvidenceEvents, &stdout, &stderr)
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" {
		t.Fatalf("evidence result report=%+v err=%v, want separate evidence-capacity verdict", report, err)
	}
	if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusCapacityExhausted || !bytes.Equal(stdout.Bytes(), want) || stderr.Len() != 0 {
		t.Fatalf("evidence capture changed app result: outcome=%+v report=%+v stdout length=%d stderr=%q, want successful child and exact %d app bytes", outcome, report, stdout.Len(), stderr.String(), len(want))
	}
}

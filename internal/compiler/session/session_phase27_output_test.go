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
	artifact := filepath.Join(t.TempDir(), "fizzbuzz")
	restoreEvidenceLimit := cgen.SetApplicationEvidenceOutputLimitForTest(4)
	defer restoreEvidenceLimit()
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}

	reportPath := filepath.Join(t.TempDir(), "fizzbuzz-evidence.json")
	var stdout, stderr bytes.Buffer
	outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, "15", reportPath, native.EvidenceEvents, &stdout, &stderr)
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" {
		t.Fatalf("evidence result report=%+v err=%v, want separate evidence-capacity verdict", report, err)
	}
	if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusCapacityExhausted || !bytes.Equal(stdout.Bytes(), want) || stderr.Len() != 0 {
		t.Fatalf("evidence capture changed app result: outcome=%+v report=%+v stdout length=%d stderr=%q, want successful child and exact %d app bytes", outcome, report, stdout.Len(), stderr.String(), len(want))
	}
}

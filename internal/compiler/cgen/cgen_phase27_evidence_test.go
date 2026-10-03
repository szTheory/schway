package cgen_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27EvidenceCapacity(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "fizzbuzz.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "fizzbuzz-15.out"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("FizzBuzz check diagnostics: %+v", checked.Diagnostics)
	}

	normalC, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	const eventCapacityPrefix = "#define SCHWAY_EVENT_CAPACITY "
	if !strings.Contains(normalC, eventCapacityPrefix) {
		t.Fatal("generated application omitted its bounded event capacity")
	}
	eventLimitedC := replaceGeneratedEventCapacity(t, normalC)
	assertPhase27CapacityExhausted(t, source, eventLimitedC, "event-capacity", want)

	var evidenceLimitedC string
	func() {
		restoreEvidenceLimit := cgen.SetApplicationEvidenceOutputLimitForTest(8)
		defer restoreEvidenceLimit()
		evidenceLimitedC, err = cgen.EmitApplication(checked.Program)
	}()
	if err != nil {
		t.Fatalf("EmitApplication with lower evidence byte capacity: %v", err)
	}
	assertPhase27CapacityExhausted(t, source, evidenceLimitedC, "evidence-byte-capacity", want)
}

func replaceGeneratedEventCapacity(t *testing.T, generated string) string {
	t.Helper()
	const prefix = "#define SCHWAY_EVENT_CAPACITY "
	start := strings.Index(generated, prefix)
	if start < 0 {
		t.Fatal("generated application has no event capacity definition")
	}
	end := strings.IndexByte(generated[start:], '\n')
	if end < 0 {
		t.Fatal("generated event capacity definition is unterminated")
	}
	end += start
	original := generated[start:end]
	if original == prefix+"1u" {
		t.Fatal("generated event capacity was already one; lowered-capacity control is inert")
	}
	return strings.Replace(generated, original, prefix+"1u", 1)
}

func assertPhase27CapacityExhausted(t *testing.T, source []byte, generated, name string, want []byte) {
	t.Helper()
	artifact := filepath.Join(t.TempDir(), name)
	if _, err := native.BuildApplication(context.Background(), source, generated, artifact); err != nil {
		t.Fatalf("BuildApplication(%s): %v", name, err)
	}
	reportPath := filepath.Join(t.TempDir(), name+"-report.json")
	var stdout, stderr bytes.Buffer
	outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, "15", reportPath, native.EvidenceEvents, &stdout, &stderr)
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" {
		t.Fatalf("%s result=%+v outcome=%+v err=%v; want explicit evidence-capacity exhaustion", name, report, outcome, err)
	}
	if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusCapacityExhausted || report.Verified || report.Execution != nil || !bytes.Equal(stdout.Bytes(), want) || stderr.Len() != 0 {
		t.Fatalf("%s capacity conflated app and evidence outcomes: outcome=%+v report=%+v stdout length=%d stderr=%q", name, outcome, report, stdout.Len(), stderr.String())
	}
}

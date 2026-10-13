package native

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
)

func TestPhase33CaptureReadBounded(t *testing.T) {
	t.Run("exactly-limit", func(t *testing.T) {
		capture := phase22CompleteCapture(t)
		limit := len(capture) + 64
		padded := capture[:1] + strings.Repeat(" ", 64) + capture[1:]
		artifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'app\\n'\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(padded)))
		runner := DefaultRunner()
		runner.evidenceLimit = limit
		outcome, report, err := runPhase33Capture(t, runner, artifact)
		if err != nil || outcome.Kind != RunExited || outcome.ExitCode != 0 || report.CaptureStatus != EvidenceStatusComplete || report.Execution == nil {
			t.Fatalf("exactly-limit outcome=%+v report=%+v err=%v; want complete capture at %d bytes", outcome, report, err, limit)
		}
	})

	t.Run("limit-plus-one", func(t *testing.T) {
		capture := phase22CompleteCapture(t)
		limit := len(capture) + 64
		padded := capture[:1] + strings.Repeat(" ", 64) + capture[1:]
		artifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'app\\n'\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(padded)))
		runner := DefaultRunner()
		runner.evidenceLimit = limit - 1
		outcome, report, err := runPhase33Capture(t, runner, artifact)
		var toolError *ToolError
		if !errors.As(err, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" || outcome.Kind != RunExited || outcome.ExitCode != 0 || report.CaptureStatus != EvidenceStatusCapacityExhausted {
			t.Fatalf("limit-plus-one outcome=%+v report=%+v err=%v; want native.evidence_capacity_exhausted", outcome, report, err)
		}
	})

	t.Run("sparse-1gib-allocation-bounded", func(t *testing.T) {
		capture := phase22CompleteCapture(t)
		baselineArtifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'app\\n'\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(capture)))
		baselineDelta, baselineOutcome, baselineReport, baselineErr := runPhase33MeasuredCapture(t, DefaultRunner(), baselineArtifact)
		if baselineErr != nil || baselineOutcome.Kind != RunExited || baselineOutcome.ExitCode != 0 || baselineReport.CaptureStatus != EvidenceStatusComplete {
			t.Fatalf("baseline capture outcome=%+v report=%+v err=%v; want complete", baselineOutcome, baselineReport, baselineErr)
		}

		sparsePath := filepath.Join(t.TempDir(), "sparse-capture.json")
		file, err := os.Create(sparsePath)
		if err != nil {
			t.Fatalf("precondition: sparse capture construction: create: %v", err)
		}
		// D-16 sparse allocation control: file.Truncate(1<<30) constructs a 1 GiB sparse file.
		if err := file.Truncate(1 << 30); err != nil {
			_ = file.Close()
			t.Fatalf("precondition: sparse capture construction: truncate: %v", err)
		}
		if err := file.Close(); err != nil {
			t.Fatalf("precondition: sparse capture construction: close: %v", err)
		}
		info, err := os.Stat(sparsePath)
		if err != nil || info.Size() != 1<<30 {
			t.Fatalf("precondition: sparse capture construction: stat=%v size=%v; want 1 GiB", err, info)
		}
		sparseArtifact := writePhase22EvidenceScript(t, fmt.Sprintf("printf 'app\\n'\nln -s %s \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(sparsePath)))
		sparseDelta, sparseOutcome, sparseReport, sparseErr := runPhase33MeasuredCapture(t, DefaultRunner(), sparseArtifact)
		info, err = os.Stat(sparsePath)
		if err != nil || info.Size() != 1<<30 {
			t.Fatalf("precondition: sparse capture construction: post-run stat=%v size=%v; want 1 GiB", err, info)
		}
		var toolError *ToolError
		if !errors.As(sparseErr, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" || sparseOutcome.Kind != RunExited || sparseOutcome.ExitCode != 0 || sparseReport.CaptureStatus != EvidenceStatusCapacityExhausted {
			t.Fatalf("sparse capture outcome=%+v report=%+v err=%v; want native.evidence_capacity_exhausted", sparseOutcome, sparseReport, sparseErr)
		}
		difference := int64(sparseDelta) - int64(baselineDelta)
		t.Logf("baseline TotalAlloc delta=%d sparse TotalAlloc delta=%d difference=%d", baselineDelta, sparseDelta, difference)
		if difference >= 16*execution.MaxApplicationEvidenceBytes {
			t.Fatalf("sparse capture allocation difference=%d; want < %d bytes (baseline=%d sparse=%d)", difference, 16*execution.MaxApplicationEvidenceBytes, baselineDelta, sparseDelta)
		}
	})
}

func runPhase33Capture(t *testing.T, runner Runner, artifact string) (RunOutcome, ApplicationEvidenceReport, error) {
	t.Helper()
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	outcome, report, err := runner.RunApplicationWithEvidence(context.Background(), artifact, "7", filepath.Join(root, "capture-report.json"), EvidenceEvents, &stdout, &stderr)
	return outcome, report, err
}

func runPhase33MeasuredCapture(t *testing.T, runner Runner, artifact string) (uint64, RunOutcome, ApplicationEvidenceReport, error) {
	t.Helper()
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	outcome, report, err := runPhase33Capture(t, runner, artifact)
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	return after.TotalAlloc - before.TotalAlloc, outcome, report, err
}

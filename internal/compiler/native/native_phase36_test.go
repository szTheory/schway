package native_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

const nativePhase36CaptureOutcomeSource = `module phase36.capture_outcome

export {
  fn main
}

fn increment(value: U64) -> U64 {
  let one = 1
  let result = value + one
  result
}

fn main(input: U64) -> Unit {
  var index = 0
  var result = input + 18446744073709546615
  write_u64(input)
  write_newline()
  while index < 5000 {
    result = increment(result)
    index = index + 1
  }
}
`

func TestPhase36CaptureOutcomeSeparation(t *testing.T) {
	checked := session.Check([]byte(nativePhase36CaptureOutcomeSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	generated = nativePhase36ReplaceDefine(t, generated, "SCHWAY_INVOCATION_CAPACITY", "1u")
	artifact := filepath.Join(t.TempDir(), "phase36-capture-outcome")
	if _, err := native.BuildApplication(context.Background(), []byte(nativePhase36CaptureOutcomeSource), generated, artifact); err != nil {
		t.Fatalf("BuildApplication: %v", err)
	}

	t.Run("successful-work-after-capture-exhaustion", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		outcome, report, runErr := native.DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "0", filepath.Join(t.TempDir(), "report.json"), native.EvidenceEvents, &stdout, &stderr)
		var toolError *native.ToolError
		if !errors.As(runErr, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusCapacityExhausted || report.Execution != nil || stdout.String() != "0\n" || stderr.Len() != 0 {
			t.Fatalf("capture exhaustion changed successful work: outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
		}
	})

	t.Run("late-defect-after-capture-exhaustion", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		outcome, report, runErr := native.DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "1", filepath.Join(t.TempDir(), "report.json"), native.EvidenceEvents, &stdout, &stderr)
		var toolError *native.ToolError
		if !errors.As(runErr, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusCapacityExhausted || report.Execution != nil || stdout.String() != "1\n" || stderr.String() != "schway: U64 addition overflow\n" {
			t.Fatalf("capture exhaustion hid the later defect: outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
		}
	})

	t.Run("capture-disabled-more-than-4096-dynamic-calls", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		outcome, report, runErr := native.DefaultRunner().RunApplicationWithEvidence(context.Background(), artifact, "0", filepath.Join(t.TempDir(), "report.json"), native.EvidenceDisabled, &stdout, &stderr)
		if runErr != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusDisabled || stdout.String() != "0\n" || stderr.Len() != 0 {
			t.Fatalf("disabled evidence affected 5000 dynamic calls: outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
		}
	})
}

func nativePhase36ReplaceDefine(t *testing.T, generated, name, value string) string {
	t.Helper()
	prefix := "#define " + name + " "
	start := strings.Index(generated, prefix)
	if start < 0 {
		t.Fatalf("generated C lacks %s", name)
	}
	end := strings.IndexByte(generated[start:], '\n')
	if end < 0 {
		t.Fatalf("generated %s definition has no line ending", name)
	}
	end += start
	return generated[:start] + prefix + value + generated[end:]
}

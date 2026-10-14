package session_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase35FocusedValueAnswers(t *testing.T) {
	const source = `module phase35.application
export { fn main }
fn main(x: U64) -> U64 {
  let y = 0x21
  let witness = x ^ (y << 3) & 0xFF
  witness
}
`
	const input = "256"
	const want = "264"

	interpreted, diagnostics, err := session.RunInterpreterApplication([]byte(source), input)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("interpreter application diagnostics=%+v err=%v", diagnostics, err)
	}
	if interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || interpreted.Execution.Outcome.Value != want || string(interpreted.Output) != want+"\n" {
		t.Fatalf("interpreter application outcome=%+v output=%q, want returned %s and %q", interpreted.Execution.Outcome, interpreted.Output, want, want+"\n")
	}

	artifact := filepath.Join(t.TempDir(), "phase35-application")
	if _, diagnostics, err := session.BuildApplication(context.Background(), []byte(source), artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build application diagnostics=%+v err=%v", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("native application outcome=%+v stdout=%q stderr=%q err=%v", outcome, stdout.String(), stderr.String(), err)
	}
	if stdout.String() != want+"\n" {
		t.Fatalf("native application stdout=%q, want %q", stdout.String(), want+"\n")
	}
}

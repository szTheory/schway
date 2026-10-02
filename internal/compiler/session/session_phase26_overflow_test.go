package session_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase26OverflowProcessOutcome(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "phase26", "checked_add_overflow.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("overflow witness check diagnostics: %+v", checked.Diagnostics)
	}
	interpreted, err := interp.Run(checked.Program, "main", "1")
	if err != nil || interpreted.Outcome.Kind != execution.OutcomeDefect || interpreted.Outcome.Value != "" {
		t.Fatalf("interpreter overflow outcome=%+v err=%v, want empty defect", interpreted.Outcome, err)
	}
	if len(interpreted.Events) == 0 {
		t.Fatal("interpreter overflow omitted its event")
	}
	event := interpreted.Events[len(interpreted.Events)-1]
	if event.Kind != "function.defected" || event.Output != "U64 addition overflow" || event.SourcePlace == "" {
		t.Fatalf("interpreter overflow event=%+v, want source-attributed overflow reason", event)
	}

	artifact := filepath.Join(t.TempDir(), "checked_add_overflow")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "1", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway: U64 addition overflow\n" {
		t.Fatalf("native overflow outcome=%+v stdout=%q stderr=%q err=%v, want exit 65 and fixed stderr only", outcome, stdout.String(), stderr.String(), err)
	}

	var toolError *native.ToolError
	missing := filepath.Join(filepath.Dir(artifact), "missing-artifact")
	stdout.Reset()
	stderr.Reset()
	outcome, err = native.RunApplication(context.Background(), missing, "1", &stdout, &stderr)
	if !errors.As(err, &toolError) || toolError == nil {
		t.Fatalf("missing artifact outcome=%+v err=%v, want a native ToolError", outcome, err)
	}
	if outcome.Kind == native.RunExited && outcome.ExitCode == 65 {
		t.Fatalf("missing artifact was misclassified as a child overflow: %+v", outcome)
	}
}

package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase26TracerAppSum10(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sum_to_n check diagnostics: %+v", checked.Diagnostics)
	}
	interpreted, err := interp.Run(checked.Program, "main", "10")
	if err != nil {
		t.Fatalf("interp.Run(main, 10): %v", err)
	}
	if interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != "55" {
		t.Fatalf("interpreter outcome = %+v, want returned 55", interpreted.Outcome)
	}

	artifact := filepath.Join(t.TempDir(), "sum_to_n")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil {
		t.Fatalf("BuildApplicationFile: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("sum_to_n build diagnostics: %+v", diagnostics)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "10", &stdout, &stderr)
	if err != nil {
		t.Fatalf("RunApplication: %v", err)
	}
	if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != "55\n" || stderr.Len() != 0 {
		t.Fatalf("app outcome=%+v stdout=%q stderr=%q, want exit 0, 55 newline, empty stderr", outcome, stdout.String(), stderr.String())
	}
}

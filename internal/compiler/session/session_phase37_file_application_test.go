package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase37InterpreterFileApplication(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "crc32", "crc32.schway")
	checked, err := session.CheckFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source diagnostics = %+v", checked.Diagnostics)
	}
	inputPath := filepath.Join(t.TempDir(), "crc-input.bin")
	if err := os.WriteFile(inputPath, []byte("123456789"), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := interp.RunWithFileBytes(checked.Program, "main", inputPath, func(path string) ([]byte, string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, "FileUnavailable"
		}
		return data, ""
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Outcome.Kind != execution.OutcomeReturned || result.Outcome.Value != "3421780262" {
		t.Fatalf("outcome = %+v, want returned 3421780262", result.Outcome)
	}
	if len(result.Events) != 0 {
		t.Fatalf("capture-disabled file application retained %d events; want zero", len(result.Events))
	}
}

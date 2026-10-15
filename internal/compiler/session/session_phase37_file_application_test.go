package session_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"time"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase37InterpreterFileApplication(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "crc32", "crc32.schway")
	inputPath := filepath.Join(t.TempDir(), "crc-input.bin")
	run := func(name string, input []byte, want uint64) {
		t.Helper()
		if err := os.WriteFile(inputPath, input, 0o600); err != nil {
			t.Fatal(err)
		}
		started := time.Now()
		result, diagnostics, err := session.RunInterpreterFileApplication(sourcePath, inputPath)
		elapsed := time.Since(started)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("%s diagnostics = %+v", name, diagnostics)
		}
		if result.Execution.Outcome.Kind != execution.OutcomeReturned || result.Execution.Outcome.Value != strconv.FormatUint(want, 10) {
			t.Fatalf("%s outcome = %+v, want returned %d", name, result.Execution.Outcome, want)
		}
		if got, expected := string(result.Output), strconv.FormatUint(want, 10)+"\n"; got != expected {
			t.Fatalf("%s output = %q, want %q", name, got, expected)
		}
		if result.Execution.Events != nil || len(result.Execution.Events) != 0 {
			t.Fatalf("%s retained events: len=%d cap=%d; want nil and zero", name, len(result.Execution.Events), cap(result.Execution.Events))
		}
		if name == "maximum-65536-ff" {
			t.Logf("maximum-case elapsed=%s input_bytes=%d host=%s/%s go=%s", elapsed, len(input), runtime.GOOS, runtime.GOARCH, runtime.Version())
		}
	}
	run("123456789", []byte("123456789"), 3421780262)
	run("maximum-65536-ff", bytes.Repeat([]byte{0xff}, 65536), 3735780942)
}

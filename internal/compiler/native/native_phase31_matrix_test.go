package native_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase31MatrixDeadline = 5 * time.Second

// phase31MatrixInterp projects an interpreter run to (exit, stdout, stderr).
func phase31MatrixInterp(t *testing.T, checked session.CheckResult, path string) (int, string, string, interp.Execution) {
	t.Helper()
	run, err := interp.RunWithFileBytes(checked.Program, "main", path, testsupport.ChecksumReadFile)
	if err != nil {
		t.Fatal(err)
	}
	switch run.Outcome.Kind {
	case execution.OutcomeReturned:
		return 0, run.Outcome.Value + "\n", "", run
	case execution.OutcomeTypedFailure:
		return 65, "", "FileError." + run.Outcome.Value + "\n", run
	}
	t.Fatalf("unexpected interpreter outcome %+v", run.Outcome)
	return 0, "", "", run
}

// TestPhase31CaseMatrix runs one shared case list on the native engine (public
// build + RunApplication) and on the interpreter, comparing both against the
// literal expectations (D-07, D-08, D-09, D-14). No -short gate.
func TestPhase31CaseMatrix(t *testing.T) {
	phase29RequireClang(t)
	source := testsupport.ProjectPath("examples", "checksum", "checksum.schway")
	manifest := testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), "checksum")
	if _, diagnostics, err := session.BuildApplicationFile(context.Background(), source, artifact, native.DefaultRunner(), manifest); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build checksum: diagnostics=%+v err=%v", diagnostics, err)
	}
	checked, err := session.CheckFile(source)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check checksum: %+v %v", checked.Diagnostics, err)
	}
	for _, c := range testsupport.ChecksumCaseMatrix() {
		t.Run(c.Name, func(t *testing.T) {
			path := testsupport.MaterializeChecksumCase(t, t.TempDir(), c)
			assertRow := func(t *testing.T, exit int, stdout, stderr string) {
				t.Helper()
				if exit != c.Exit || stdout != c.Stdout || stderr != c.Stderr {
					t.Fatalf("(exit,stdout,stderr) = (%d,%q,%q), want (%d,%q,%q)", exit, stdout, stderr, c.Exit, c.Stdout, c.Stderr)
				}
				if strings.Contains(stdout+stderr, testsupport.ChecksumSecret) {
					t.Fatalf("output names the path: %q %q", stdout, stderr)
				}
				if (stdout != "") == (stderr != "") {
					t.Fatalf("want exactly one non-empty stream, got stdout=%q stderr=%q", stdout, stderr)
				}
			}
			t.Run("native", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), phase31MatrixDeadline)
				defer cancel()
				var stdout, stderr bytes.Buffer
				outcome, err := native.RunApplication(ctx, artifact, path, &stdout, &stderr)
				if err != nil {
					t.Fatal(err)
				}
				if outcome.Kind != native.RunExited {
					t.Fatalf("outcome = %+v, want exit", outcome)
				}
				assertRow(t, outcome.ExitCode, stdout.String(), stderr.String())
			})
			t.Run("interpreter", func(t *testing.T) {
				type result struct {
					exit           int
					stdout, stderr string
					run            interp.Execution
				}
				done := make(chan result, 1)
				go func() {
					exit, stdout, stderr, run := phase31MatrixInterp(t, checked, path)
					done <- result{exit, stdout, stderr, run}
				}()
				select {
				case r := <-done:
					assertRow(t, r.exit, r.stdout, r.stderr)
					if len(r.run.LiveResources) != 0 {
						t.Fatalf("live resources = %v, want none", r.run.LiveResources)
					}
					if c.Exit == 0 {
						released := 0
						for _, event := range r.run.Events {
							if event.Kind == "resource.released" {
								released++
							}
						}
						if released != 1 {
							t.Fatalf("release events = %d, want 1", released)
						}
					}
				case <-time.After(phase31MatrixDeadline):
					t.Fatal("interpreter row exceeded the deadline")
				}
			})
		})
	}
}

// TestPhase31CaseMatrixNativeBoundaries pins the pre-launch refusal and the
// FIFO deadline on the native route (D-09, T-31-18, T-31-20).
func TestPhase31CaseMatrixNativeBoundaries(t *testing.T) {
	phase29RequireClang(t)
	source := testsupport.ProjectPath("examples", "checksum", "checksum.schway")
	manifest := testsupport.ProjectPath("examples", "checksum", "file_bytes.bindings.json")
	artifact := filepath.Join(t.TempDir(), "checksum")
	if _, diagnostics, err := session.BuildApplicationFile(context.Background(), source, artifact, native.DefaultRunner(), manifest); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build checksum: diagnostics=%+v err=%v", diagnostics, err)
	}
	t.Run("nul-argument", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		_, err := native.RunApplication(context.Background(), artifact, "phase31-secret\x00tail", &stdout, &stderr)
		if err == nil {
			t.Fatal("embedded NUL argument was not refused before launch")
		}
		if stdout.Len() != 0 || stderr.Len() != 0 || strings.Contains(stdout.String()+stderr.String(), "phase31-secret") {
			t.Fatalf("refusal wrote output: stdout=%q stderr=%q", stdout.String(), stderr.String())
		}
	})
	t.Run("fifo-deadline", func(t *testing.T) {
		var fifo testsupport.ChecksumCase
		for _, c := range testsupport.ChecksumCaseMatrix() {
			if c.Kind == "fifo" {
				fifo = c
			}
		}
		path := testsupport.MaterializeChecksumCase(t, t.TempDir(), fifo)
		ctx, cancel := context.WithTimeout(context.Background(), phase31MatrixDeadline)
		defer cancel()
		start := time.Now()
		var stdout, stderr bytes.Buffer
		outcome, err := native.RunApplication(ctx, artifact, path, &stdout, &stderr)
		if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stderr.String() != "FileError.FileUnavailable\n" {
			t.Fatalf("fifo outcome=%+v err=%v stderr=%q", outcome, err, stderr.String())
		}
		if time.Since(start) >= phase31MatrixDeadline {
			t.Fatalf("fifo row took %v", time.Since(start))
		}
	})
}

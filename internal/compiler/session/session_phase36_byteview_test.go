package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase36ByteViewLoopCaller = `fn identity(value: U64) -> U64 {
  let result = value + 1
  result
}
fn adler32(bytes: ByteView) -> U64 {
  var index = 0
  var answer = 0
  while index < byte_len(bytes) {
    answer = answer + identity(byte_at(bytes, index))
    index = index + 1
  }
  answer
}
`

func TestPhase36NativeByteViewLoopHelpers(t *testing.T) {
	// Independent byte arithmetic: sum(b"Wikipedia") = 919; nine increments = 928.
	phase36RunByteViewComposition(t, phase36ByteViewLoopCaller, "3", "928", 9)
}

func TestPhase36NativeFileByteViewLoopHelpers(t *testing.T) {
	dir := phase31CopyDir(t, testsupport.ProjectPath("examples", "checksum"))
	library := filepath.Join(dir, "adler32.schway")
	original, err := os.ReadFile(library)
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(original), "fn adler32(")
	if start < 0 {
		t.Fatal("missing copied ByteView library")
	}
	if err := os.WriteFile(library, []byte(string(original[:start])+phase36ByteViewLoopCaller), 0o600); err != nil {
		t.Fatal(err)
	}
	entry := filepath.Join(dir, "checksum.schway")
	checked := phase36CheckFile(t, entry)
	helper := phase36FunctionID(checked.Program.Functions, "identity")
	input := phase31CopyInput(t, "Wikipedia")
	interpreted, err := interp.RunWithFileBytes(checked.Program, "main", input, testsupport.ChecksumReadFile)
	if err != nil || interpreted.Outcome.Kind != execution.OutcomeReturned || interpreted.Outcome.Value != "928" || len(interpreted.LiveResources) != 0 || phase36CallsTo(interpreted.Events, helper) != 9 {
		t.Fatalf("interpreter file loop outcome=%+v live=%v calls=%d err=%v", interpreted.Outcome, interpreted.LiveResources, phase36CallsTo(interpreted.Events, helper), err)
	}
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	for _, lane := range []struct{ name, flags string }{{"O0", "-O0"}, {"O3", "-O3"}, {"O3-LTO", "-O3 -flto"}} {
		t.Run(lane.name, func(t *testing.T) {
			buildDir := t.TempDir()
			wrapper := filepath.Join(buildDir, "clang-lane")
			if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nexec "+phase36ShellQuote(clang)+" \"$@\" "+lane.flags+"\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			runner := native.DefaultRunner()
			runner.ClangPath = wrapper
			artifact := filepath.Join(buildDir, "file-loop")
			if _, diagnostics, err := session.BuildApplicationFile(context.Background(), entry, artifact, runner, filepath.Join(dir, "file_bytes.bindings.json")); err != nil || len(diagnostics) != 0 {
				t.Fatalf("file loop build diagnostics=%+v err=%v", diagnostics, err)
			}
			var stdout, stderr bytes.Buffer
			outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, input, filepath.Join(buildDir, "report.json"), native.EvidenceEvents, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != "928\n" || stderr.Len() != 0 || report.CaptureStatus != native.EvidenceStatusComplete || report.Execution == nil || report.Execution.Outcome.Value != "928" || phase36CallsTo(report.Execution.Events, helper) != 9 {
				t.Fatalf("native file loop outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), err)
			}
			if err := executionpeer.Validate(checked.Program, *report.Execution); err != nil {
				t.Fatalf("independent file loop causal peer: %v", err)
			}
			t.Logf("native file wrapper %s literal=928 helper_calls=9 capture=complete causal_peer=accepted", lane.name)
		})
	}
}

func TestPhase36ByteViewArgumentOrderMutationControls(t *testing.T) {
	root := testsupport.ProjectPath()
	implementation := filepath.Join(root, "internal/compiler/cgen/cgen_program.go")
	original, err := os.ReadFile(implementation)
	if err != nil {
		t.Fatal(err)
	}
	const before = `for _, block := range function.Linear.Blocks {
		fmt.Fprintf(out, "%s:\n", label(block.ID))
		terminated := false`
	const after = `for _, block := range function.Linear.Blocks {
		// Mutant: invoke a helper before evaluating its byte_at argument.
		if function.Parameter.Type == "ByteView" {
			block.OperationIDs = append([]string(nil), block.OperationIDs...)
			for callIndex, id := range block.OperationIDs {
				if operations[id].Kind != core.OpCall { continue }
				for argumentIndex := 0; argumentIndex < callIndex; argumentIndex++ {
					if operations[block.OperationIDs[argumentIndex]].Kind == core.OpBytesAt {
						copy(block.OperationIDs[argumentIndex+1:callIndex+1], block.OperationIDs[argumentIndex:callIndex])
						block.OperationIDs[argumentIndex] = id
						break
					}
				}
			}
		}
		fmt.Fprintf(out, "%s:\n", label(block.ID))
		terminated := false`
	mutated, counts := phase36ApplyEngineMutation(string(original), []phase36EngineMutationEdit{{before: before, after: after}})
	if len(counts) != 1 || counts[0] != 1 || mutated == string(original) {
		t.Fatalf("argument-order mutation matches=%v", counts)
	}
	for _, lane := range []string{"O0", "O3", "O3-LTO"} {
		t.Run(lane, func(t *testing.T) {
			selector := "^TestPhase36NativeByteViewCallerArgumentBoundary$/^at-length$/^" + lane + "$"
			target := "TestPhase36NativeByteViewCallerArgumentBoundary/at-length/" + lane
			for _, control := range []struct {
				name   string
				source []byte
				fail   bool
			}{
				{"fresh-no-op", original, false}, {"early-helper", []byte(mutated), true}, {"restored", original, false},
			} {
				output, childErr := phase36RunEngineMutationChild(t, root, implementation, control.source, selector, nil)
				if control.fail {
					if childErr == nil {
						t.Fatal("early-helper argument-order mutant survived")
					}
					phase36RequireEngineMutationChild(t, output, target, "FAIL", "native helper invocation events=1 want 0")
				} else {
					if childErr != nil {
						t.Fatalf("%s %s: %v\n%s", lane, control.name, childErr, output)
					}
					phase36RequireEngineMutationChild(t, output, target, "PASS", "")
				}
				t.Logf("%s %s reached native helper-entry assertion", lane, control.name)
			}
		})
	}
}

// The application shell intentionally does not publish normal defect capture.
// This test-only exit observer serializes the actual recorder at exit, including
// exit(65), so zero helper-entry events is measured rather than inferred from
// build success or generated-text order. It leaves execution and events intact.
func phase36RunByteViewComposition(t *testing.T, caller, input, want string, wantCalls int) {
	t.Helper()
	source := phase29InlineFixture(t, caller)
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("ByteView caller diagnostics: %+v", checked.Diagnostics)
	}
	helperID := phase36FunctionID(checked.Program.Functions, "identity")
	if helperID == "" {
		t.Fatal("missing identity helper")
	}
	interpreted, err := interp.RunWithFixtureBytes(checked.Program, "main", input, testsupport.Phase29FixtureBytes)
	if err != nil {
		t.Fatal(err)
	}
	if phase36CallsTo(interpreted.Events, helperID) != wantCalls {
		t.Fatalf("interpreter helper calls=%d want %d", phase36CallsTo(interpreted.Events, helperID), wantCalls)
	}
	if want == "" {
		if interpreted.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(interpreted.Events) != "byte index out of bounds" || len(interpreted.LiveResources) != 1 {
			t.Fatalf("interpreter bounds outcome=%+v live=%v", interpreted.Outcome, interpreted.LiveResources)
		}
	} else if interpreted.Outcome.Kind != execution.OutcomeReturned || interpreted.Outcome.Value != want || len(interpreted.LiveResources) != 0 {
		t.Fatalf("interpreter value=%+v live=%v want %s", interpreted.Outcome, interpreted.LiveResources, want)
	}
	generated, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("ByteView composition EmitApplication: %v", err)
	}
	const mainStart = "int main(int argc, char **argv) {\n"
	if strings.Count(generated, mainStart) != 1 {
		t.Fatal("expected one application main")
	}
	observer := `static void phase36_observe_events(void) {
  const char *path = getenv("SCHWAY_PHASE36_EVENTS");
  FILE *stream;
  if (path == NULL) return;
  stream = fopen(path, "wb");
  if (stream == NULL) abort();
  schway_output_stream = stream;
  schway_output_count = 0u;
  schway_output_limit = SCHWAY_EVIDENCE_OUTPUT_LIMIT;
  if (!schway_write_literal("[") || !schway_write_events() || !schway_write_literal("]")) abort();
  if (fclose(stream) != 0) abort();
}
`
	observed := strings.Replace(generated, mainStart, observer+mainStart+"  if (atexit(phase36_observe_events) != 0) abort();\n", 1)
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	manifest := testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json")
	for _, lane := range []struct{ name, flags string }{{"O0", "-O0"}, {"O3", "-O3"}, {"O3-LTO", "-O3 -flto"}} {
		t.Run(lane.name, func(t *testing.T) {
			dir := t.TempDir()
			wrapper := filepath.Join(dir, "clang-lane")
			if err := os.WriteFile(wrapper, []byte("#!/bin/sh\nexec "+phase36ShellQuote(clang)+" \"$@\" "+lane.flags+"\n"), 0o700); err != nil {
				t.Fatal(err)
			}
			runner := native.DefaultRunner()
			runner.ClangPath = wrapper
			artifact := filepath.Join(dir, "byteview-composition")
			if _, diagnostics, err := session.BuildApplication(context.Background(), source, artifact, runner, manifest); err != nil || len(diagnostics) != 0 {
				t.Fatalf("session native %s BuildApplication diagnostics=%+v err=%v", lane.name, diagnostics, err)
			}
			var plainStdout, plainStderr bytes.Buffer
			plainOutcome, err := native.RunApplication(context.Background(), artifact, input, &plainStdout, &plainStderr)
			wantExit, wantStdout, wantStderr := 0, want+"\n", ""
			if want == "" {
				wantExit, wantStdout, wantStderr = 65, "", "schway: byte index out of bounds\n"
			}
			if err != nil || plainOutcome.Kind != native.RunExited || plainOutcome.ExitCode != wantExit || plainStdout.String() != wantStdout || plainStderr.String() != wantStderr {
				t.Fatalf("unmodified session application outcome=%+v stdout=%q stderr=%q err=%v", plainOutcome, plainStdout.String(), plainStderr.String(), err)
			}
			if _, err := runner.BuildApplication(context.Background(), source, observed, artifact, manifest); err != nil {
				t.Fatalf("native %s BuildApplication: %v", lane.name, err)
			}
			trace := filepath.Join(dir, "events.json")
			t.Setenv("SCHWAY_PHASE36_EVENTS", trace)
			var stdout, stderr bytes.Buffer
			evidence := filepath.Join(dir, "report.json")
			outcome, report, runErr := native.RunApplicationWithEvidence(context.Background(), artifact, input, evidence, native.EvidenceEvents, &stdout, &stderr)
			if want == "" {
				// Existing application failure capture is incomplete; actual process
				// exit/streams and the separate observer own this assertion.
				if outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway: byte index out of bounds\n" || report.CaptureStatus != native.EvidenceStatusIncomplete {
					t.Fatalf("native bounds outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
				}
			} else {
				if runErr != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != want+"\n" || stderr.Len() != 0 || report.CaptureStatus != native.EvidenceStatusComplete || report.Execution == nil || report.Execution.Outcome.Value != want {
					t.Fatalf("native value outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
				}
				if err := executionpeer.Validate(checked.Program, *report.Execution); err != nil {
					t.Fatalf("independent native composition causal peer: %v", err)
				}
			}
			data, err := os.ReadFile(trace)
			if err != nil {
				t.Fatal(err)
			}
			var events []execution.Event
			if err := json.Unmarshal(data, &events); err != nil {
				t.Fatal(err)
			}
			if got := phase36CallsTo(events, helperID); got != wantCalls {
				t.Fatalf("native helper invocation events=%d want %d", got, wantCalls)
			}
			seen := map[string]bool{}
			for _, event := range events {
				if event.Kind == "function.returned" && event.FunctionID == helperID {
					if seen[event.Invocation] {
						t.Fatalf("reused helper invocation %q", event.Invocation)
					}
					seen[event.Invocation] = true
				}
			}
			if len(seen) != wantCalls {
				t.Fatalf("native helper returned invocations=%d want %d", len(seen), wantCalls)
			}
			t.Logf("native %s literal=%s exit=%d helper_calls=%d distinct_helper_returns=%d", lane.name, fmt.Sprint(want), outcome.ExitCode, wantCalls, len(seen))
		})
	}
}

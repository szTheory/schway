package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase36HelperAnswersAndDefects(t *testing.T) {
	fixture := phase36ReadHelperValues(t)

	for _, row := range []struct{ file, caseName string }{
		{"crc_step.schway", "local-four-step-loop"},
		{"importer.schway", "imported-four-step-loop"},
	} {
		row := row
		t.Run(row.caseName, func(t *testing.T) {
			path := testsupport.ProjectPath("examples", "helper_calls", row.file)
			phase36RunSourceFile(t, path, "305419897", phase36FindHelperCase(t, fixture, row.caseName))
		})
	}

	stepSource := phase36SingleStepSource()
	for _, name := range []string{
		"single-step-odd-one", "single-step-even-two", "single-step-high-even",
		"single-step-high-odd", "single-step-u32-maximum",
	} {
		row := phase36FindHelperCase(t, fixture, name)
		t.Run(name, func(t *testing.T) {
			phase36RunSource(t, stepSource, row.Input, row)
		})
	}

	pairsSource := phase36ScalarPairsSource()
	for _, name := range []string{"scalar-pairs-true-branch", "scalar-pairs-false-branch"} {
		row := phase36FindHelperCase(t, fixture, name)
		t.Run(name, func(t *testing.T) {
			phase36RunSource(t, pairsSource, row.Input, row)
		})
	}

	for _, row := range phase36DefectPrograms(t, fixture) {
		row := row
		t.Run(row.caseName, func(t *testing.T) {
			phase36RunSource(t, row.source, row.expected.Input, row.expected)
		})
	}
}

func TestPhase36HelperOptimizationMatrix(t *testing.T) {
	fixture := phase36ReadHelperValues(t)
	programs := []struct {
		name     string
		program  func(*testing.T) session.CheckResult
		caseName string
	}{
		{"local", func(t *testing.T) session.CheckResult {
			return phase36CheckFile(t, testsupport.ProjectPath("examples", "helper_calls", "crc_step.schway"))
		}, "local-four-step-loop"},
		{"imported", func(t *testing.T) session.CheckResult {
			return phase36CheckFile(t, testsupport.ProjectPath("examples", "helper_calls", "importer.schway"))
		}, "imported-four-step-loop"},
	}
	lanes := []struct {
		name         string
		optimization string
		lto          bool
	}{
		{"O0", "-O0", false},
		{"O3", "-O3", false},
		{"O3-LTO", "-O3", true},
	}
	answerRuns := 0
	for _, program := range programs {
		program := program
		t.Run(program.name, func(t *testing.T) {
			checked := program.program(t)
			entry, err := callgraph.EntryFunction(checked.Program)
			if err != nil {
				t.Fatalf("entry function: %v", err)
			}
			pruned, err := callgraph.PruneToReachable(checked.Program, entry.ID)
			if err != nil {
				t.Fatalf("prune reachable program: %v", err)
			}
			want := phase36FindHelperCase(t, fixture, program.caseName)
			interpreted, err := interp.RunApplication(pruned, entry.Name, want.Input)
			if err != nil {
				t.Fatalf("interpreter application: %v", err)
			}
			if interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || interpreted.Execution.Outcome.Value != want.ExpectedDecimal || string(interpreted.Output) != want.ExpectedStdout {
				t.Fatalf("interpreter outcome=%+v stdout=%q, want returned %s and %q", interpreted.Execution.Outcome, interpreted.Output, want.ExpectedDecimal, want.ExpectedStdout)
			}
			generated, err := cgen.EmitNative(pruned)
			if err != nil {
				t.Fatalf("EmitNative: %v", err)
			}
			for _, lane := range lanes {
				lane := lane
				t.Run(lane.name, func(t *testing.T) {
					runner := native.DefaultRunner().EnableCommandRecording()
					runner.LTO = lane.lto
					got, err := runner.Run(context.Background(), generated, lane.optimization, []string{want.Input})
					if err != nil || len(got.Pairs) != 1 {
						t.Fatalf("native %s pairs=%d err=%v", lane.name, len(got.Pairs), err)
					}
					if err := phase36CompareExecution("native "+lane.name, got.Pairs[0].Execution, want); err != nil {
						t.Fatal(err)
					}
					if lane.lto {
						commands := runner.LastCommandLines()
						if len(commands) == 0 {
							t.Fatal("O3-LTO lane recorded no compiler command")
						}
						for _, command := range commands {
							if !phase36HasArgument(command, "-flto") || !phase36HasArgument(command, "-O3") {
								t.Fatalf("O3-LTO compiler command lacks -O3 -flto: %v", command)
							}
						}
					}
					answerRuns++
				})
			}
		})
	}
	if answerRuns != 6 {
		t.Fatalf("native helper answer lane executions=%d, want 6 (two programs x three lanes)", answerRuns)
	}

	defects := phase36DefectPrograms(t, fixture)
	for _, row := range defects {
		row := row
		if row.caseName != "u64-helper-overflow" && row.caseName != "bool-helper-zero-remainder" {
			continue
		}
		t.Run(row.caseName, func(t *testing.T) {
			checked := session.Check([]byte(row.source))
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("defect program diagnostics: %+v", checked.Diagnostics)
			}
			entry, err := callgraph.EntryFunction(checked.Program)
			if err != nil {
				t.Fatalf("entry function: %v", err)
			}
			pruned, err := callgraph.PruneToReachable(checked.Program, entry.ID)
			if err != nil {
				t.Fatalf("prune reachable defect program: %v", err)
			}
			interpreted, err := interp.RunApplication(pruned, entry.Name, row.expected.Input)
			if err != nil || interpreted.Execution.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(interpreted.Execution.Events) != row.expected.DefectReason {
				t.Fatalf("interpreter defect outcome=%+v err=%v", interpreted.Execution.Outcome, err)
			}
			generated, err := cgen.EmitNative(pruned)
			if err != nil {
				t.Fatalf("EmitNative defect program: %v", err)
			}
			for _, lane := range lanes {
				lane := lane
				t.Run(lane.name, func(t *testing.T) {
					runner := native.DefaultRunner().EnableCommandRecording()
					runner.Expect = native.ExpectDefect
					runner.LTO = lane.lto
					got, err := runner.Run(context.Background(), generated, lane.optimization, []string{row.expected.Input})
					if err != nil || len(got.Pairs) != 1 {
						t.Fatalf("native %s defect pairs=%d err=%v", lane.name, len(got.Pairs), err)
					}
					helperName := map[string]string{"u64-helper-overflow": "add_one", "bool-helper-zero-remainder": "remainder_to_bool"}[row.caseName]
					helperID := phase36FunctionID(checked.Program.Functions, helperName)
					if helperID == "" || phase36CallsTo(got.Pairs[0].Execution.Events, helperID) == 0 {
						t.Fatalf("native %s defect trace does not prove entry into helper %q", lane.name, helperName)
					}
					if err := phase36CompareExecution("native "+lane.name, got.Pairs[0].Execution, row.expected); err != nil {
						t.Fatal(err)
					}
					if lane.lto {
						commands := runner.LastCommandLines()
						if len(commands) == 0 {
							t.Fatal("O3-LTO defect lane recorded no compiler command")
						}
						for _, command := range commands {
							if !phase36HasArgument(command, "-flto") || !phase36HasArgument(command, "-O3") {
								t.Fatalf("O3-LTO defect command lacks -O3 -flto: %v", command)
							}
						}
					}
				})
			}
		})
	}
}

func TestPhase36HelperUBSan(t *testing.T) {
	fixture := phase36ReadHelperValues(t)
	clangPath, err := exec.LookPath("clang")
	if err != nil {
		t.Fatalf("required installed Clang is unavailable: %v", err)
	}
	wrapper := filepath.Join(t.TempDir(), "clang-ubsan")
	script := "#!/bin/sh\nexec " + phase36ShellQuote(clangPath) + " -fsanitize=undefined -fno-sanitize-recover=all \"$@\"\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatalf("write UBSan Clang wrapper: %v", err)
	}
	runner := native.DefaultRunner()
	runner.ClangPath = wrapper
	checked := phase36CheckFile(t, testsupport.ProjectPath("examples", "helper_calls", "crc_step.schway"))
	entry, err := callgraph.EntryFunction(checked.Program)
	if err != nil {
		t.Fatalf("local entry function: %v", err)
	}
	pruned, err := callgraph.PruneToReachable(checked.Program, entry.ID)
	if err != nil {
		t.Fatalf("prune local program: %v", err)
	}
	generated, err := cgen.EmitNative(pruned)
	if err != nil {
		t.Fatalf("EmitNative local CRC program: %v", err)
	}
	positive := phase36FindHelperCase(t, fixture, "local-four-step-loop")
	got, err := runner.Run(context.Background(), generated, "-O3", []string{positive.Input})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("UBSan local helper pairs=%d err=%v", len(got.Pairs), err)
	}
	if err := phase36CompareExecution("native UBSan", got.Pairs[0].Execution, positive); err != nil {
		t.Fatal(err)
	}
	for _, row := range phase36DefectPrograms(t, fixture) {
		if row.caseName != "u64-helper-overflow" && row.caseName != "bool-helper-zero-remainder" {
			continue
		}
		checked := session.Check([]byte(row.source))
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("UBSan defect diagnostics: %+v", checked.Diagnostics)
		}
		entry, err := callgraph.EntryFunction(checked.Program)
		if err != nil {
			t.Fatalf("UBSan defect entry: %v", err)
		}
		pruned, err := callgraph.PruneToReachable(checked.Program, entry.ID)
		if err != nil {
			t.Fatalf("prune UBSan defect program: %v", err)
		}
		generated, err := cgen.EmitNative(pruned)
		if err != nil {
			t.Fatalf("EmitNative UBSan defect program: %v", err)
		}
		defectRunner := runner
		defectRunner.Expect = native.ExpectDefect
		got, err := defectRunner.Run(context.Background(), generated, "-O3", []string{row.expected.Input})
		if err != nil || len(got.Pairs) != 1 {
			t.Fatalf("UBSan %s pairs=%d err=%v", row.caseName, len(got.Pairs), err)
		}
		if err := phase36CompareExecution("native UBSan", got.Pairs[0].Execution, row.expected); err != nil {
			t.Fatal(err)
		}
	}
}

type phase36HelperFixture struct {
	Schema string                 `json:"schema"`
	Cases  []phase36HelperCaseRow `json:"cases"`
}

type phase36HelperCaseRow struct {
	Source          string `json:"source"`
	Case            string `json:"case"`
	Input           string `json:"input"`
	ExpectedDecimal string `json:"expected_decimal"`
	ExpectedStdout  string `json:"expected_stdout"`
	ExpectedStderr  string `json:"expected_stderr"`
	ExpectedExit    int    `json:"expected_exit"`
	DefectReason    string `json:"defect_reason"`
	Calculation     string `json:"calculation"`
}

func phase36ReadHelperValues(t *testing.T) phase36HelperFixture {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase36", "helper-values.json"))
	if err != nil {
		t.Fatalf("read independent Phase 36 helper values: %v", err)
	}
	var fixture phase36HelperFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode independent Phase 36 helper values: %v", err)
	}
	if fixture.Schema != "schway.phase36.helper-values/1" || len(fixture.Cases) < 14 {
		t.Fatalf("invalid Phase 36 helper fixture: schema=%q cases=%d", fixture.Schema, len(fixture.Cases))
	}
	return fixture
}

func phase36FindHelperCase(t *testing.T, fixture phase36HelperFixture, name string) phase36HelperCaseRow {
	t.Helper()
	for _, row := range fixture.Cases {
		if row.Case == name {
			return row
		}
	}
	t.Fatalf("independent helper fixture has no case %q", name)
	return phase36HelperCaseRow{}
}

func phase36SingleStepSource() string {
	return `module phase36.single_step

export {
  fn main
}

fn crc_step(state: U64) -> U64 {
  var next = state >> 1
  if state & 1 == 1 {
    next = next ^ 0xEDB88320
  } else {
    next = state >> 1
  }
  next
}

fn main(state: U64) -> U64 {
  let result = crc_step(state)
  result
}
`
}

func phase36ScalarPairsSource() string {
	return `module phase36.scalar_pairs

export {
  fn main
}

fn u64_u64(value: U64) -> U64 {
  let result = value + 2
  result
}

fn u64_bool(value: U64) -> Bool {
  let result = value < 8
  result
}

fn bool_u64(value: Bool) -> U64 {
  var result = 11
  if value {
    result = 7
  } else {
    result = 11
  }
  result
}

fn bool_bool(value: Bool) -> Bool {
  let result = value
  result
}

fn main(seed: U64) -> U64 {
  let predicate = u64_bool(u64_u64(seed + 2))
  let number = bool_u64(predicate)
  let flipped = bool_bool(predicate)
  var result = 0
  if flipped {
    result = number + 2
  } else {
    result = number + 23
  }
  if predicate {
    result = result + 3
  } else {
    result = result + 2
  }
  result
}
`
}

func phase36DefectPrograms(t *testing.T, fixture phase36HelperFixture) []struct {
	caseName string
	source   string
	expected phase36HelperCaseRow
} {
	return []struct {
		caseName string
		source   string
		expected phase36HelperCaseRow
	}{
		{"u64-helper-overflow", `module phase36.u64_overflow
export { fn main }
fn add_one(value: U64) -> U64 {
  let result = value + 1
  result
}
fn main(value: U64) -> U64 {
  let result = add_one(value)
  result
}
`, phase36FindHelperCase(t, fixture, "u64-helper-overflow")},
		{"bool-helper-overflow", `module phase36.bool_overflow
export { fn main }
fn overflow_to_bool(value: U64) -> Bool {
  let incremented = value + 1
  let result = incremented == 0
  result
}
fn main(value: U64) -> U64 {
  var result = 0
  if overflow_to_bool(value) {
    result = 1
  } else {
    result = 2
  }
  result
}
`, phase36FindHelperCase(t, fixture, "bool-helper-overflow")},
		{"u64-helper-zero-remainder", `module phase36.u64_remainder
export { fn main }
fn even_remainder(value: U64) -> U64 {
  let result = value % 2
  result
}
fn divide_by_helper(value: U64) -> U64 {
  let divisor = even_remainder(value)
  let result = value % divisor
  result
}
fn main(value: U64) -> U64 {
  let result = divide_by_helper(value)
  result
}
`, phase36FindHelperCase(t, fixture, "u64-helper-zero-remainder")},
		{"bool-helper-zero-remainder", `module phase36.bool_remainder
export { fn main }
fn even_remainder(value: U64) -> U64 {
  let result = value % 2
  result
}
fn remainder_to_bool(value: U64) -> Bool {
  let divisor = even_remainder(value)
  let remainder = value % divisor
  let result = remainder == 0
  result
}
fn main(value: U64) -> U64 {
  var result = 0
  if remainder_to_bool(value) {
    result = 1
  } else {
    result = 2
  }
  result
}
`, phase36FindHelperCase(t, fixture, "bool-helper-zero-remainder")},
		{"nested-helper-overflow", `module phase36.nested_overflow
export { fn main }
fn identity(value: U64) -> U64 {
  let result = value
  result
}
fn outer(value: U64) -> U64 {
  let result = value + 1
  result
}
fn main(value: U64) -> U64 {
  let result = outer(identity(value))
  result
}
`, phase36FindHelperCase(t, fixture, "nested-helper-overflow")},
		{"late-loop-helper-overflow", `module phase36.late_loop_overflow
export { fn main }
fn bump(value: U64) -> U64 {
  let result = value + 1
  result
}
fn main(seed: U64) -> U64 {
  var value = seed
  var i = 0
  while i < 3 {
    value = bump(value)
    i = i + 1
  }
  value
}
`, phase36FindHelperCase(t, fixture, "late-loop-helper-overflow")},
	}
}

func phase36RunSource(t *testing.T, source, input string, expected phase36HelperCaseRow) {
	t.Helper()
	interpreted, diagnostics, err := session.RunInterpreterApplication([]byte(source), input)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("interpreter application diagnostics=%+v err=%v", diagnostics, err)
	}
	if err := phase36CompareExecution("interpreter", interpreted.Execution, expected); err != nil {
		t.Fatal(err)
	}
	if string(interpreted.Output) != expected.ExpectedStdout {
		t.Fatalf("interpreter stdout=%q, want %q", interpreted.Output, expected.ExpectedStdout)
	}
	artifact := filepath.Join(t.TempDir(), "phase36-application")
	if _, diagnostics, err := session.BuildApplication(context.Background(), []byte(source), artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplication diagnostics=%+v err=%v", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != expected.ExpectedExit || stdout.String() != expected.ExpectedStdout || stderr.String() != expected.ExpectedStderr {
		t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v, want exit=%d stdout=%q stderr=%q", outcome, stdout.String(), stderr.String(), err, expected.ExpectedExit, expected.ExpectedStdout, expected.ExpectedStderr)
	}
}

func phase36RunSourceFile(t *testing.T, sourcePath, input string, expected phase36HelperCaseRow) {
	t.Helper()
	checked := phase36CheckFile(t, sourcePath)
	entry, err := callgraph.EntryFunction(checked.Program)
	if err != nil {
		t.Fatalf("entry function: %v", err)
	}
	pruned, err := callgraph.PruneToReachable(checked.Program, entry.ID)
	if err != nil {
		t.Fatalf("prune reachable program: %v", err)
	}
	interpreted, err := interp.RunApplication(pruned, entry.Name, input)
	if err != nil {
		t.Fatalf("interpreter application: %v", err)
	}
	if err := phase36CompareExecution("interpreter", interpreted.Execution, expected); err != nil || string(interpreted.Output) != expected.ExpectedStdout {
		t.Fatalf("interpreter outcome=%+v stdout=%q: %v", interpreted.Execution.Outcome, interpreted.Output, err)
	}
	artifact := filepath.Join(t.TempDir(), "phase36-file-application")
	if _, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != expected.ExpectedExit || stdout.String() != expected.ExpectedStdout || stderr.String() != expected.ExpectedStderr {
		t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v", outcome, stdout.String(), stderr.String(), err)
	}
}

func phase36CheckFile(t *testing.T, path string) session.CheckResult {
	t.Helper()
	checked, err := session.CheckFile(path)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check source file %s: err=%v diagnostics=%+v", path, err, checked.Diagnostics)
	}
	return checked
}

func phase36CompareExecution(engine string, got execution.Execution, expected phase36HelperCaseRow) error {
	if expected.ExpectedExit == 0 {
		if got.Outcome.Kind != execution.OutcomeReturned || got.Outcome.Value != expected.ExpectedDecimal {
			return fmt.Errorf("%s outcome=%+v, want returned %s", engine, got.Outcome, expected.ExpectedDecimal)
		}
		return nil
	}
	if got.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(got.Events) != expected.DefectReason {
		return fmt.Errorf("%s outcome=%+v defect_reason=%q, want %q", engine, got.Outcome, phase36DefectEvent(got.Events), expected.DefectReason)
	}
	return nil
}

func phase36DefectEvent(events []execution.Event) string {
	for index := len(events) - 1; index >= 0; index-- {
		if events[index].Kind == "function.defected" {
			return events[index].Output
		}
	}
	return ""
}

func phase36HasArgument(arguments []string, want string) bool {
	for _, argument := range arguments {
		if argument == want {
			return true
		}
	}
	return false
}

func phase36ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func TestPhase36ByteViewCallerArgumentBoundary(t *testing.T) {
	const caller = `fn identity(value: U64) -> U64 {
  let result = value + 1
  result
}

fn adler32(bytes: ByteView) -> U64 {
  let index = byte_len(bytes)
  var answer = 0
  answer = identity(byte_at(bytes, index))
  let later = answer + 2
  later
}
`
	validCaller := strings.Replace(caller, "let index = byte_len(bytes)", "let index = 0", 1)
	for _, tc := range []struct {
		name   string
		callee string
		want   string
	}{
		{"valid-index-enters-helper", validCaller, "90"},
		{"out-of-bounds-argument-stops-before-helper", caller, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := phase29InlineFixture(t, tc.callee)
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("ByteView caller source diagnostics: %+v", checked.Diagnostics)
			}
			identityID := phase36FunctionID(checked.Program.Functions, "identity")
			if identityID == "" {
				t.Fatal("checked caller fixture has no identity helper")
			}
			run, err := interp.RunWithFixtureBytes(checked.Program, "main", "3", testsupport.Phase29FixtureBytes)
			if err != nil {
				t.Fatalf("fixture interpreter run: %v", err)
			}
			calls := phase36CallsTo(run.Events, identityID)
			if strings.Contains(tc.name, "out-of-bounds") {
				if run.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(run.Events) != "byte index out of bounds" || len(run.LiveResources) != 1 || calls != 0 {
					t.Fatalf("caller OOB outcome=%+v live=%v helper_calls=%d, want byte-index defect with live owner and zero helper calls", run.Outcome, run.LiveResources, calls)
				}
			} else if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != tc.want || len(run.LiveResources) != 0 || calls != 1 {
				t.Fatalf("valid caller outcome=%+v live=%v helper_calls=%d, want %s and one helper call", run.Outcome, run.LiveResources, calls, tc.want)
			}

		})
	}
}

// TestPhase36NativeByteViewCallerArgumentRefusal records a known native
// boundary: the interpreter proves byte_at faults before identity is entered,
// but the native application emitter currently refuses this scalar helper
// call inside a ByteView caller. Passing this test documents the refusal; it
// does not satisfy the native caller-ordering requirement.
func TestPhase36NativeByteViewCallerArgumentRefusal(t *testing.T) {
	const caller = `fn identity(value: U64) -> U64 {
  let result = value + 1
  result
}

fn adler32(bytes: ByteView) -> U64 {
  let index = byte_len(bytes)
  var answer = 0
  answer = identity(byte_at(bytes, index))
  let later = answer + 2
  later
}
`
	source := phase29InlineFixture(t, caller)
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("ByteView caller source diagnostics: %+v", checked.Diagnostics)
	}
	identityID := phase36FunctionID(checked.Program.Functions, "identity")
	interpreted, err := interp.RunWithFixtureBytes(checked.Program, "main", "3", testsupport.Phase29FixtureBytes)
	if err != nil || interpreted.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(interpreted.Events) != "byte index out of bounds" || len(interpreted.LiveResources) != 1 || phase36CallsTo(interpreted.Events, identityID) != 0 {
		t.Fatalf("interpreter caller control outcome=%+v live=%v helper_calls=%d err=%v", interpreted.Outcome, interpreted.LiveResources, phase36CallsTo(interpreted.Events, identityID), err)
	}
	manifest := testsupport.ProjectPath("examples", "phase29", "fixture_bytes.bindings.json")
	_, diagnostics, buildErr := session.BuildApplication(context.Background(), source, filepath.Join(t.TempDir(), "byteview-caller"), native.DefaultRunner(), manifest)
	if len(diagnostics) != 0 {
		t.Fatalf("native ByteView caller diagnostics: %+v", diagnostics)
	}
	const refusal = `:fn:adler32:op:4": scalar call emission requires dynamic invocation support`
	if buildErr == nil || !strings.Contains(buildErr.Error(), refusal) {
		t.Fatalf("native ByteView caller build error=%v, want explicit cgen refusal containing %q", buildErr, refusal)
	}
	t.Logf("native semantic caller ordering remains blocked: reproduction=identity(byte_at(bytes, byte_len(bytes))) after caller fixture acquisition; refusal=%q", buildErr)
}

func TestPhase36LegacyByteViewCalleeBounds(t *testing.T) {
	fixtureSource := phase29InlineFixture(t, phase36LegacyOOBAdler32())
	checkedFixture := session.Check(fixtureSource)
	if len(checkedFixture.Diagnostics) != 0 {
		t.Fatalf("legacy fixture OOB source diagnostics: %+v", checkedFixture.Diagnostics)
	}
	fixtureRun, err := interp.RunWithFixtureBytes(checkedFixture.Program, "main", "3", testsupport.Phase29FixtureBytes)
	if err != nil || fixtureRun.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(fixtureRun.Events) != "byte index out of bounds" || len(fixtureRun.LiveResources) != 1 {
		t.Fatalf("legacy fixture callee bounds outcome=%+v live=%v err=%v, want byte-index defect with live owner", fixtureRun.Outcome, fixtureRun.LiveResources, err)
	}
	validFixture, err := interp.RunWithFixtureBytes(phase29CheckedFixture(t).Program, "main", "3", testsupport.Phase29FixtureBytes)
	if err != nil || validFixture.Outcome.Kind != execution.OutcomeReturned || validFixture.Outcome.Value != "300286872" || len(validFixture.LiveResources) != 0 {
		t.Fatalf("restored fixture control outcome=%+v live=%v err=%v, want returned 300286872 with no live owner", validFixture.Outcome, validFixture.LiveResources, err)
	}

	dir := phase31CopyDir(t, testsupport.ProjectPath("examples", "checksum"))
	libraryPath := filepath.Join(dir, "adler32.schway")
	originalLibrary, err := os.ReadFile(libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(originalLibrary), "fn adler32(")
	if start < 0 {
		t.Fatal("copied legacy adler32 library has no function")
	}
	mutatedLibrary := string(originalLibrary[:start]) + phase36LegacyOOBAdler32()
	if err := os.WriteFile(libraryPath, []byte(mutatedLibrary), 0o600); err != nil {
		t.Fatalf("write OOB library copy: %v", err)
	}
	entryPath := filepath.Join(dir, "checksum.schway")
	checkedFile := phase36CheckFile(t, entryPath)
	fileInput := phase31CopyInput(t, "x")
	fileRun, err := interp.RunWithFileBytes(checkedFile.Program, "main", fileInput, testsupport.ChecksumReadFile)
	if err != nil || fileRun.Outcome.Kind != execution.OutcomeDefect || phase36DefectEvent(fileRun.Events) != "byte index out of bounds" || len(fileRun.LiveResources) != 1 {
		t.Fatalf("legacy file callee bounds outcome=%+v live=%v err=%v, want byte-index defect with live owner", fileRun.Outcome, fileRun.LiveResources, err)
	}

	manifestPath := filepath.Join(dir, "file_bytes.bindings.json")
	retainedArtifact := filepath.Join(t.TempDir(), "checksum-oob-retained")
	if _, diagnostics, err := session.BuildApplicationFile(context.Background(), entryPath, retainedArtifact, native.DefaultRunner(), manifestPath); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile OOB diagnostics=%+v err=%v", diagnostics, err)
	}
	var retainedStdout, retainedStderr bytes.Buffer
	retainedOutcome, err := native.RunApplication(context.Background(), retainedArtifact, fileInput, &retainedStdout, &retainedStderr)
	if err != nil || retainedOutcome.Kind != native.RunExited || retainedOutcome.ExitCode != 65 || retainedStdout.Len() != 0 || retainedStderr.String() != "schway: byte index out of bounds\n" {
		t.Fatalf("retained OOB app outcome=%+v stdout=%q stderr=%q err=%v, want exit 65 and exact bounds reason", retainedOutcome, retainedStdout.String(), retainedStderr.String(), err)
	}

	generated, err := cgen.EmitApplication(checkedFile.Program)
	if err != nil {
		t.Fatalf("EmitApplication OOB program: %v", err)
	}
	// Runner.Run compiles in a private temporary directory and has no binding
	// manifest staging hook. Point this test-only generated translation unit at
	// the copied adapter header while keeping the adapter C as a separate unit.
	headerPath := filepath.ToSlash(filepath.Join(dir, "adapter.h"))
	const localAdapterInclude = `#include "local/adapter.h"`
	if strings.Count(generated, localAdapterInclude) != 1 {
		t.Fatalf("generated C should include the legacy adapter header once")
	}
	generated = strings.Replace(generated, localAdapterInclude, `#include "`+headerPath+`"`, 1)
	clangPath, err := exec.LookPath("clang")
	if err != nil {
		t.Fatalf("required installed Clang is unavailable: %v", err)
	}
	for _, lane := range []struct {
		name         string
		optimization string
		lto          bool
	}{
		{"O0", "-O0", false},
		{"O3", "-O3", false},
		{"O3-LTO", "-O3", true},
	} {
		lane := lane
		t.Run("native-"+lane.name, func(t *testing.T) {
			laneDir := t.TempDir()
			generatedPath := filepath.Join(laneDir, "program.c")
			if err := os.WriteFile(generatedPath, []byte(generated), 0o600); err != nil {
				t.Fatalf("write %s generated C: %v", lane.name, err)
			}
			binaryPath := filepath.Join(laneDir, "checksum-oob-lane")
			compileArgs := []string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", lane.optimization}
			if lane.lto {
				compileArgs = append(compileArgs, "-flto")
			}
			compileArgs = append(compileArgs, generatedPath, filepath.Join(dir, "adapter.c"), "-o", binaryPath)
			compileCmd := exec.CommandContext(context.Background(), clangPath, compileArgs...)
			var compileStdout, compileStderr bytes.Buffer
			compileCmd.Stdout, compileCmd.Stderr = &compileStdout, &compileStderr
			if err := compileCmd.Run(); err != nil {
				t.Fatalf("compile native %s application with %v: %v stdout=%q stderr=%q", lane.name, compileArgs, err, compileStdout.String(), compileStderr.String())
			}
			if lane.lto && (!phase36HasArgument(compileArgs, "-O3") || !phase36HasArgument(compileArgs, "-flto")) {
				t.Fatalf("native O3-LTO command lacks required flags: %v", compileArgs)
			}
			cmd := exec.Command(binaryPath, fileInput)
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			runErr := cmd.Run()
			exitCode := -1
			if runErr != nil {
				var exitErr *exec.ExitError
				if !errors.As(runErr, &exitErr) {
					t.Fatalf("run %s application binary: %v", lane.name, runErr)
				}
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 0
			}
			if exitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway: byte index out of bounds\n" {
				t.Fatalf("native %s application exit=%d stdout=%q stderr=%q, want exit 65 and exact bounds reason", lane.name, exitCode, stdout.String(), stderr.String())
			}
		})
	}

	if err := os.WriteFile(libraryPath, originalLibrary, 0o600); err != nil {
		t.Fatalf("restore library control: %v", err)
	}
	restored := phase36CheckFile(t, entryPath)
	restoredRun, err := interp.RunWithFileBytes(restored.Program, "main", fileInput, testsupport.ChecksumReadFile)
	if err != nil || restoredRun.Outcome.Kind != execution.OutcomeReturned || restoredRun.Outcome.Value != "7929977" || len(restoredRun.LiveResources) != 0 {
		t.Fatalf("restored file control outcome=%+v live=%v err=%v, want returned 7929977 with no live owner", restoredRun.Outcome, restoredRun.LiveResources, err)
	}
	validArtifact := filepath.Join(t.TempDir(), "checksum-restored")
	if _, diagnostics, err := session.BuildApplicationFile(context.Background(), entryPath, validArtifact, native.DefaultRunner(), manifestPath); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile restored diagnostics=%+v err=%v", diagnostics, err)
	}
	var validStdout, validStderr bytes.Buffer
	validOutcome, err := native.RunApplication(context.Background(), validArtifact, fileInput, &validStdout, &validStderr)
	if err != nil || validOutcome.Kind != native.RunExited || validOutcome.ExitCode != 0 || validStdout.String() != "7929977\n" || validStderr.Len() != 0 {
		t.Fatalf("restored native control outcome=%+v stdout=%q stderr=%q err=%v", validOutcome, validStdout.String(), validStderr.String(), err)
	}
}

func TestPhase36NativeCallCaptureOutcomeSeparation(t *testing.T) {
	fixture := phase36ReadHelperValues(t)
	localPath := testsupport.ProjectPath("examples", "helper_calls", "crc_step.schway")
	localBytes, err := os.ReadFile(localPath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(localBytes)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("local helper diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication repeated helper: %v", err)
	}
	generated = phase36ReplaceDefine(t, generated, "SCHWAY_INVOCATION_CAPACITY", "1u")
	successArtifact := filepath.Join(t.TempDir(), "helper-capture-success")
	if _, err := native.BuildApplication(context.Background(), localBytes, generated, successArtifact); err != nil {
		t.Fatalf("BuildApplication repeated helper: %v", err)
	}
	late := phase36FindDefectProgram(t, fixture, "late-loop-helper-overflow")
	lateChecked := session.Check([]byte(late.source))
	if len(lateChecked.Diagnostics) != 0 {
		t.Fatalf("late helper defect diagnostics: %+v", lateChecked.Diagnostics)
	}
	lateGenerated, err := cgen.EmitApplication(lateChecked.Program)
	if err != nil {
		t.Fatalf("EmitApplication late defect: %v", err)
	}
	lateGenerated = phase36ReplaceDefine(t, lateGenerated, "SCHWAY_INVOCATION_CAPACITY", "1u")
	defectArtifact := filepath.Join(t.TempDir(), "helper-capture-defect")
	if _, err := native.BuildApplication(context.Background(), []byte(late.source), lateGenerated, defectArtifact); err != nil {
		t.Fatalf("BuildApplication late helper defect: %v", err)
	}

	success := phase36FindHelperCase(t, fixture, "local-four-step-loop")
	defect := phase36FindHelperCase(t, fixture, "late-loop-helper-overflow")
	for _, tc := range []struct {
		name     string
		artifact string
		input    string
		want     phase36HelperCaseRow
	}{
		{"success", successArtifact, success.Input, success},
		{"late-defect", defectArtifact, defect.Input, defect},
	} {
		tc := tc
		t.Run(tc.name+"-capture-exhausted", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			outcome, report, runErr := native.DefaultRunner().RunApplicationWithEvidence(context.Background(), tc.artifact, tc.input, filepath.Join(t.TempDir(), "report.json"), native.EvidenceEvents, &stdout, &stderr)
			var toolError *native.ToolError
			if !errors.As(runErr, &toolError) || toolError.Code != "native.evidence_capacity_exhausted" || outcome.Kind != native.RunExited || outcome.ExitCode != tc.want.ExpectedExit || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusCapacityExhausted || report.Execution != nil || stdout.String() != tc.want.ExpectedStdout || stderr.String() != tc.want.ExpectedStderr {
				t.Fatalf("capture exhausted app outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
			}
		})
		t.Run(tc.name+"-capture-disabled", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			outcome, report, runErr := native.DefaultRunner().RunApplicationWithEvidence(context.Background(), tc.artifact, tc.input, filepath.Join(t.TempDir(), "report.json"), native.EvidenceDisabled, &stdout, &stderr)
			if runErr != nil || outcome.Kind != native.RunExited || outcome.ExitCode != tc.want.ExpectedExit || report.ProcessOutcome != outcome || report.CaptureStatus != native.EvidenceStatusDisabled || stdout.String() != tc.want.ExpectedStdout || stderr.String() != tc.want.ExpectedStderr {
				t.Fatalf("capture-disabled app outcome=%+v report=%+v stdout=%q stderr=%q err=%v", outcome, report, stdout.String(), stderr.String(), runErr)
			}
		})
	}
}

func phase36LegacyOOBAdler32() string {
	return `fn adler32(bytes: ByteView) -> U64 {
  let index = byte_len(bytes)
  let result = 0 + byte_at(bytes, index)
  result
}
`
}

func phase29CheckedFixture(t *testing.T) session.CheckResult {
	t.Helper()
	return phase36CheckFile(t, phase29FixturePath)
}

func phase36FunctionID(functions []core.Function, name string) string {
	for _, function := range functions {
		if function.Name == name {
			return function.ID
		}
	}
	return ""
}

func phase36CallsTo(events []execution.Event, calleeID string) int {
	count := 0
	for _, event := range events {
		if event.Kind == "function.called" && event.CalleeFunctionID == calleeID {
			count++
		}
	}
	return count
}

func phase36FindDefectProgram(t *testing.T, fixture phase36HelperFixture, name string) struct {
	caseName string
	source   string
	expected phase36HelperCaseRow
} {
	t.Helper()
	for _, row := range phase36DefectPrograms(t, fixture) {
		if row.caseName == name {
			return row
		}
	}
	t.Fatalf("unknown Phase 36 defect program %q", name)
	return struct {
		caseName string
		source   string
		expected phase36HelperCaseRow
	}{}
}

func phase36ReplaceDefine(t *testing.T, generated, name, value string) string {
	t.Helper()
	prefix := "#define " + name + " "
	start := strings.Index(generated, prefix)
	if start < 0 {
		t.Fatalf("generated C lacks %s", name)
	}
	end := strings.IndexByte(generated[start:], '\n')
	if end < 0 {
		t.Fatalf("generated C definition %s has no line ending", name)
	}
	end += start
	return generated[:start] + prefix + value + generated[end:]
}

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

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
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

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
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
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

func TestPhase35BitwiseOptimizationMatrix(t *testing.T) {
	fixture := phase35ReadValueVectors(t)
	lanes := []struct {
		name         string
		optimization string
		runner       native.Runner
	}{
		{name: "O0", optimization: "-O0", runner: native.DefaultRunner()},
		{name: "O3", optimization: "-O3", runner: native.DefaultRunner()},
	}
	ltoRunner := native.DefaultRunner().EnableCommandRecording()
	ltoRunner.LTO = true
	lanes = append(lanes, struct {
		name         string
		optimization string
		runner       native.Runner
	}{name: "O3-flto", optimization: "-O3", runner: ltoRunner})
	var ltoCommandCount int
	for _, vector := range fixture.Cases {
		vector := vector
		t.Run(vector.Name, func(t *testing.T) {
			checked := phase35CheckValueSource(t, phase35VectorSource(vector))
			interpreted, err := interp.Run(checked.Program, "main", "0")
			if err != nil {
				t.Fatalf("interpreter execution: %v", err)
			}
			if err := phase35CompareEngineValue("interpreter", interpreted, vector.ExpectedDecimal); err != nil {
				t.Fatal(err)
			}
			generated, err := cgen.EmitNative(checked.Program)
			if err != nil {
				t.Fatalf("EmitNative: %v", err)
			}
			for _, lane := range lanes {
				lane := lane
				t.Run(lane.name, func(t *testing.T) {
					got, err := lane.runner.Run(context.Background(), generated, lane.optimization, []string{"0"})
					if err != nil || len(got.Pairs) != 1 {
						t.Fatalf("native %s execution pairs=%d err=%v", lane.name, len(got.Pairs), err)
					}
					if err := phase35CompareEngineValue("native "+lane.name, got.Pairs[0].Execution, vector.ExpectedDecimal); err != nil {
						t.Fatal(err)
					}
					if lane.name == "O3-flto" {
						ltoCommandCount++
					}
				})
			}
		})
	}
	if ltoCommandCount != len(fixture.Cases) {
		t.Fatalf("O3-LTO executions=%d, want one native execution for each of %d authored vectors", ltoCommandCount, len(fixture.Cases))
	}
	commands := ltoRunner.LastCommandLines()
	if len(commands) != len(fixture.Cases) {
		t.Fatalf("LTO compiler commands=%d, want %d", len(commands), len(fixture.Cases))
	}
	for index, command := range commands {
		if !phase35HasArgument(command, "-flto") || !phase35HasArgument(command, "-O3") {
			t.Fatalf("LTO command %d did not reach Clang with -O3 -flto: %v", index, command)
		}
	}
}

func TestPhase35ShiftUBSan(t *testing.T) {
	fixture := phase35ReadValueVectors(t)
	clangPath, err := exec.LookPath("clang")
	if err != nil {
		t.Fatalf("required installed Clang is unavailable: %v", err)
	}
	wrapper := filepath.Join(t.TempDir(), "clang-ubsan")
	script := "#!/bin/sh\nexec " + phase35ShellQuote(clangPath) + " -fsanitize=undefined -fno-sanitize-recover=all \"$@\"\n"
	if err := os.WriteFile(wrapper, []byte(script), 0o700); err != nil {
		t.Fatalf("write focused UBSan Clang wrapper: %v", err)
	}
	runner := native.DefaultRunner()
	runner.ClangPath = wrapper
	shiftVectors := 0
	required := map[string]bool{
		"left-count-31": false, "left-count-32": false, "left-count-63": false,
		"left-zero-count-maximum": false, "left-high-bit-truncates": false,
		"right-count-31-logical": false, "right-count-32-logical": false, "right-count-63-logical": false,
		"right-zero-count-maximum": false, "right-high-bit-count-63": false,
	}
	for _, vector := range fixture.Cases {
		if vector.Operation != "shift_left" && vector.Operation != "shift_right" {
			continue
		}
		vector := vector
		shiftVectors++
		if _, ok := required[vector.Name]; ok {
			required[vector.Name] = true
		}
		t.Run(vector.Name, func(t *testing.T) {
			checked := phase35CheckValueSource(t, phase35VectorSource(vector))
			generated, err := cgen.EmitNative(checked.Program)
			if err != nil {
				t.Fatalf("EmitNative: %v", err)
			}
			got, err := runner.Run(context.Background(), generated, "-O3", []string{"0"})
			if err != nil || len(got.Pairs) != 1 {
				t.Fatalf("native UBSan shift execution pairs=%d err=%v", len(got.Pairs), err)
			}
			if err := phase35CompareEngineValue("UBSan native", got.Pairs[0].Execution, vector.ExpectedDecimal); err != nil {
				t.Fatal(err)
			}
		})
	}
	if shiftVectors == 0 {
		t.Fatal("independent literal fixture contains no shift vectors for the sanitizer lane")
	}
	for name, seen := range required {
		if !seen {
			t.Errorf("required UBSan boundary vector %q was absent", name)
		}
	}
}

func TestPhase35EngineWrongAnswerControls(t *testing.T) {
	fixture := phase35ReadValueVectors(t)
	firstByOperation := make(map[string]phase35ValueVector)
	for _, vector := range fixture.Cases {
		if vector.Operation != "" {
			if _, exists := firstByOperation[vector.Operation]; !exists {
				firstByOperation[vector.Operation] = vector
			}
		}
	}
	wantOperations := []string{"xor", "and", "or", "shift_left", "shift_right"}
	for _, operation := range wantOperations {
		vector, ok := firstByOperation[operation]
		if !ok {
			t.Fatalf("fixture has no independent answer vector for %s", operation)
		}
		t.Run(operation, func(t *testing.T) {
			checked := phase35CheckValueSource(t, phase35VectorSource(vector))
			interpreted, err := interp.Run(checked.Program, "main", "0")
			if err != nil {
				t.Fatalf("interpreter execution: %v", err)
			}
			if err := phase35CompareEngineValue("interpreter", interpreted, vector.ExpectedDecimal); err != nil {
				t.Fatal(err)
			}
			generated, err := cgen.EmitNative(checked.Program)
			if err != nil {
				t.Fatalf("EmitNative: %v", err)
			}
			nativeRun, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"0"})
			if err != nil || len(nativeRun.Pairs) != 1 {
				t.Fatalf("native execution pairs=%d err=%v", len(nativeRun.Pairs), err)
			}
			actual := nativeRun.Pairs[0].Execution
			if err := phase35CompareEngineValue("native O0", actual, vector.ExpectedDecimal); err != nil {
				t.Fatal(err)
			}
			wrong := phase35WrongAnswer(vector.ExpectedDecimal)
			mutatedInterpreter := interpreted
			mutatedInterpreter.Outcome.Value = wrong
			if err := phase35CompareEngineValue("interpreter", mutatedInterpreter, vector.ExpectedDecimal); err == nil || !strings.Contains(err.Error(), "interpreter") {
				t.Fatalf("%s interpreter wrong-answer control was not attributed: %v", operation, err)
			}
			mutatedNative := actual
			mutatedNative.Outcome.Value = wrong
			if err := phase35CompareEngineValue("native O0", mutatedNative, vector.ExpectedDecimal); err == nil || !strings.Contains(err.Error(), "native O0") {
				t.Fatalf("%s native wrong-answer control was not attributed: %v", operation, err)
			}
			restoredInterpreter, err := interp.Run(checked.Program, "main", "0")
			if err != nil || phase35CompareEngineValue("interpreter restored", restoredInterpreter, vector.ExpectedDecimal) != nil {
				t.Fatalf("%s interpreter answer did not restore: err=%v execution=%+v", operation, err, restoredInterpreter)
			}
			restoredNative, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"0"})
			if err != nil || len(restoredNative.Pairs) != 1 || phase35CompareEngineValue("native restored", restoredNative.Pairs[0].Execution, vector.ExpectedDecimal) != nil {
				t.Fatalf("%s native answer did not restore: err=%v result=%+v", operation, err, restoredNative)
			}
		})
	}
}

type phase35ValueVectorFixture struct {
	Schema string               `json:"schema"`
	Cases  []phase35ValueVector `json:"cases"`
}

type phase35ValueVector struct {
	Name            string `json:"name"`
	Operation       string `json:"op"`
	Left            string `json:"left"`
	Right           string `json:"right"`
	Count           string `json:"count"`
	Expression      string `json:"expression"`
	X               string `json:"x"`
	Y               string `json:"y"`
	ExpectedDecimal string `json:"expected_decimal"`
}

func phase35ReadValueVectors(t *testing.T) phase35ValueVectorFixture {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "value-vectors.json"))
	if err != nil {
		t.Fatalf("read committed Phase 35 value vectors: %v", err)
	}
	var fixture phase35ValueVectorFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode committed Phase 35 value vectors: %v", err)
	}
	if fixture.Schema != "schway.phase35.u64-value-vectors/1" || len(fixture.Cases) == 0 {
		t.Fatalf("invalid Phase 35 vector fixture schema or empty cases: %+v", fixture)
	}
	return fixture
}

func phase35VectorSource(vector phase35ValueVector) string {
	body := ""
	if vector.Expression != "" {
		body = fmt.Sprintf("  let x = %s\n  let y = %s\n  let result = %s\n", vector.X, vector.Y, vector.Expression)
	} else {
		operator := map[string]string{"xor": "^", "and": "&", "or": "|"}[vector.Operation]
		if operator != "" {
			body = fmt.Sprintf("  let left = %s\n  let right = %s\n  let result = left %s right\n", vector.Left, vector.Right, operator)
		} else {
			operator = map[string]string{"shift_left": "<<", "shift_right": ">>"}[vector.Operation]
			if operator == "" {
				return ""
			}
			body = fmt.Sprintf("  let left = %s\n  let result = left %s %s\n", vector.Left, operator, vector.Count)
		}
	}
	return "module phase35.session_vector\nexport { fn main }\nfn main(unused: U64) -> U64 {\n" + body + "  result\n}\n"
}

func phase35CheckValueSource(t *testing.T, source string) session.CheckResult {
	t.Helper()
	if source == "" {
		t.Fatal("value vector has no recognized source operation")
	}
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("literal vector source failed checking: diagnostics=%+v functions=%d\n%s", checked.Diagnostics, len(checked.Program.Functions), source)
	}
	return checked
}

func phase35CompareEngineValue(engine string, got execution.Execution, want string) error {
	if got.Outcome.Kind != execution.OutcomeReturned || got.Outcome.Value != want {
		return fmt.Errorf("%s literal answer mismatch: got %+v, want returned %s", engine, got.Outcome, want)
	}
	return nil
}

func phase35WrongAnswer(expected string) string {
	if expected == "0" {
		return "1"
	}
	return "0"
}

func phase35HasArgument(arguments []string, want string) bool {
	for _, argument := range arguments {
		if argument == want {
			return true
		}
	}
	return false
}

func phase35ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27ExactAnswerMatrix(t *testing.T) {
	fizzSource, err := os.ReadFile(testsupport.ProjectPath("examples", "fizzbuzz.schway"))
	if err != nil {
		t.Fatal(err)
	}
	fifteen, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "fizzbuzz-15.out"))
	if err != nil {
		t.Fatal(err)
	}
	thousand, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "fizzbuzz-1000.out"))
	if err != nil {
		t.Fatal(err)
	}
	fizzArtifact := filepath.Join(t.TempDir(), "fizzbuzz")
	if _, diagnostics, err := session.BuildApplication(context.Background(), fizzSource, fizzArtifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build FizzBuzz diagnostics=%+v err=%v", diagnostics, err)
	}
	fizzCases := []struct {
		input string
		want  []byte
	}{{"0", []byte{}}, {"1", []byte("1\n")}, {"15", fifteen}, {"1000", thousand}}
	for _, test := range fizzCases {
		t.Run("fizzbuzz-"+test.input, func(t *testing.T) {
			interpreted, diagnostics, err := session.RunInterpreterApplication(fizzSource, test.input)
			if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || interpreted.Execution.Outcome.Value != "" || !bytes.Equal(interpreted.Output, test.want) {
				t.Fatalf("interpreter diagnostics=%+v outcome=%+v stdout=%q err=%v; want Unit and independently pinned %d bytes", diagnostics, interpreted.Execution.Outcome, interpreted.Output, err, len(test.want))
			}
			var stdout, stderr bytes.Buffer
			outcome, err := native.RunApplication(context.Background(), fizzArtifact, test.input, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || !bytes.Equal(stdout.Bytes(), test.want) || stderr.Len() != 0 {
				t.Fatalf("native outcome=%+v stdout length=%d stderr=%q err=%v; want exit 0 and independently pinned %d bytes", outcome, stdout.Len(), stderr.String(), err, len(test.want))
			}
			if !bytes.Equal(interpreted.Output, stdout.Bytes()) {
				t.Fatalf("engines differ after each matched the fixed answer: interpreter=%d native=%d bytes", len(interpreted.Output), stdout.Len())
			}
		})
	}
	interpretedFizzLimit, diagnostics, err := session.RunInterpreterApplication(fizzSource, "1001")
	if err != nil || len(diagnostics) != 0 || interpretedFizzLimit.Execution.Outcome.Kind != execution.OutcomeDefect || len(interpretedFizzLimit.Output) != 0 || !phase27HasDefectReason(interpretedFizzLimit.Execution.Events, "input exceeds 1000") {
		t.Fatalf("FizzBuzz interpreter 1001 diagnostics=%+v execution=%+v output=%q err=%v; want pre-output input defect", diagnostics, interpretedFizzLimit.Execution, interpretedFizzLimit.Output, err)
	}
	var fizzStdout, fizzStderr bytes.Buffer
	fizzOutcome, err := native.RunApplication(context.Background(), fizzArtifact, "1001", &fizzStdout, &fizzStderr)
	if err != nil || fizzOutcome.Kind != native.RunExited || fizzOutcome.ExitCode != 65 || fizzStdout.Len() != 0 || fizzStderr.String() != "schway app: input exceeds 1000\n" {
		t.Fatalf("FizzBuzz native 1001 outcome=%+v stdout=%q stderr=%q err=%v", fizzOutcome, fizzStdout.String(), fizzStderr.String(), err)
	}

	sumSource, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	sumArtifact := filepath.Join(t.TempDir(), "sum_to_n")
	if _, diagnostics, err := session.BuildApplication(context.Background(), sumSource, sumArtifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build sum_to_n diagnostics=%+v err=%v", diagnostics, err)
	}
	sumCases := []struct {
		input string
		value string
		want  []byte
	}{{"0", "0", []byte("0\n")}, {"10", "55", []byte("55\n")}, {"1000", "500500", []byte("500500\n")}}
	for _, test := range sumCases {
		t.Run("sum-"+test.input, func(t *testing.T) {
			interpreted, diagnostics, err := session.RunInterpreterApplication(sumSource, test.input)
			if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || interpreted.Execution.Outcome.Value != test.value || !bytes.Equal(interpreted.Output, test.want) {
				t.Fatalf("interpreter diagnostics=%+v outcome=%+v stdout=%q err=%v; want return %s and literal bytes %q", diagnostics, interpreted.Execution.Outcome, interpreted.Output, err, test.value, test.want)
			}
			var stdout, stderr bytes.Buffer
			outcome, err := native.RunApplication(context.Background(), sumArtifact, test.input, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || !bytes.Equal(stdout.Bytes(), test.want) || stderr.Len() != 0 {
				t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v; want exit 0 and literal bytes %q", outcome, stdout.Bytes(), stderr.String(), err, test.want)
			}
			if !bytes.Equal(interpreted.Output, stdout.Bytes()) {
				t.Fatalf("engines differ after each matched the fixed answer: interpreter=%q native=%q", interpreted.Output, stdout.Bytes())
			}
		})
	}
	interpretedSumLimit, diagnostics, err := session.RunInterpreterApplication(sumSource, "1001")
	if err != nil || len(diagnostics) != 0 || interpretedSumLimit.Execution.Outcome.Kind != execution.OutcomeDefect || len(interpretedSumLimit.Output) != 0 || !phase27HasDefectReason(interpretedSumLimit.Execution.Events, "input exceeds 1000") {
		t.Fatalf("sum interpreter 1001 diagnostics=%+v execution=%+v output=%q err=%v; want pre-output input defect", diagnostics, interpretedSumLimit.Execution, interpretedSumLimit.Output, err)
	}
	var sumStdout, sumStderr bytes.Buffer
	sumOutcome, err := native.RunApplication(context.Background(), sumArtifact, "1001", &sumStdout, &sumStderr)
	if err != nil || sumOutcome.Kind != native.RunExited || sumOutcome.ExitCode != 65 || sumStdout.Len() != 0 || sumStderr.String() != "schway app: input exceeds 1000\n" {
		t.Fatalf("sum native 1001 outcome=%+v stdout=%q stderr=%q err=%v", sumOutcome, sumStdout.String(), sumStderr.String(), err)
	}
}

func TestPhase27ReachedWrongRemainder(t *testing.T) {
	assertPhase27FizzMutationKilled(t, "wrong-remainder", "by_fifteen == 0", "by_fifteen == 1")
}

func TestPhase27ReachedWrongArithmetic(t *testing.T) {
	assertPhase27FizzMutationKilled(t, "wrong-increment", "i = i + 1", "i = i + 2")
}

func TestPhase27SkippedIteration(t *testing.T) {
	assertPhase27FizzMutationKilled(t, "skipped-first-iteration", "var i = 0", "var i = 1")
}

func assertPhase27FizzMutationKilled(t *testing.T, name, old, replacement string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "fizzbuzz.schway"))
	if err != nil {
		t.Fatal(err)
	}
	expected, err := os.ReadFile(testsupport.ProjectPath("examples", "phase27", "fizzbuzz-15.out"))
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(source), old, replacement, 1)
	if mutated == string(source) {
		t.Fatalf("%s mutation did not find its source witness %q", name, old)
	}
	mutatedSource := []byte(mutated)
	checked := session.Check(mutatedSource)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s reached mutation diagnostics: %+v", name, checked.Diagnostics)
	}
	artifact := filepath.Join(t.TempDir(), name)
	if _, diagnostics, err := session.BuildApplication(context.Background(), mutatedSource, artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("%s mutation did not build: diagnostics=%+v err=%v", name, diagnostics, err)
	}
	interpreted, diagnostics, err := session.RunInterpreterApplication(mutatedSource, "15")
	if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || bytes.Equal(interpreted.Output, expected) {
		t.Fatalf("%s mutation was not reached by the interpreter: diagnostics=%+v outcome=%+v stdout=%q err=%v", name, diagnostics, interpreted.Execution.Outcome, interpreted.Output, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "15", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stderr.Len() != 0 || bytes.Equal(stdout.Bytes(), expected) {
		t.Fatalf("%s mutation was not reached by native execution: outcome=%+v stdout=%q stderr=%q err=%v", name, outcome, stdout.Bytes(), stderr.String(), err)
	}
}

func TestPhase27OverflowZeroControls(t *testing.T) {
	overflowSource, err := os.ReadFile(testsupport.ProjectPath("examples", "phase26", "checked_add_overflow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	overflowArtifact := filepath.Join(t.TempDir(), "checked_add_overflow")
	if _, diagnostics, err := session.BuildApplication(context.Background(), overflowSource, overflowArtifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build checked-add witness diagnostics=%+v err=%v", diagnostics, err)
	}
	const maxU64 = "18446744073709551615"
	for _, test := range []struct {
		input      string
		wantKind   string
		wantValue  string
		wantOutput []byte
		wantExit   int
		wantStderr string
		wantDefect string
	}{{"0", execution.OutcomeReturned, maxU64, []byte(maxU64 + "\n"), 0, "", ""}, {"1", execution.OutcomeDefect, "", []byte{}, 65, "schway: U64 addition overflow\n", "U64 addition overflow"}} {
		interpreted, diagnostics, err := session.RunInterpreterApplication(overflowSource, test.input)
		if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != test.wantKind || interpreted.Execution.Outcome.Value != test.wantValue || !bytes.Equal(interpreted.Output, test.wantOutput) {
			t.Fatalf("checked-add input=%s interpreter diagnostics=%+v execution=%+v output=%q err=%v", test.input, diagnostics, interpreted.Execution, interpreted.Output, err)
		}
		if test.wantDefect != "" && !phase27HasDefectReason(interpreted.Execution.Events, test.wantDefect) {
			t.Fatalf("checked-add input=%s interpreter omitted exact defect %q: %+v", test.input, test.wantDefect, interpreted.Execution.Events)
		}
		var stdout, stderr bytes.Buffer
		outcome, err := native.RunApplication(context.Background(), overflowArtifact, test.input, &stdout, &stderr)
		if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != test.wantExit || !bytes.Equal(stdout.Bytes(), test.wantOutput) || stderr.String() != test.wantStderr {
			t.Fatalf("checked-add input=%s native outcome=%+v stdout=%q stderr=%q err=%v", test.input, outcome, stdout.Bytes(), stderr.String(), err)
		}
	}

	zeroSource := []byte("module phase27_dynamic_zero\nexport { fn main }\nfn main(divisor: U64) -> U64 {\n  let dividend = 17\n  var result = 0\n  if divisor == 0 {\n    let remainder = dividend % divisor\n    result = remainder\n  } else {\n    result = 0\n  }\n  result\n}\n")
	zeroArtifact := filepath.Join(t.TempDir(), "dynamic-zero")
	if _, diagnostics, err := session.BuildApplication(context.Background(), zeroSource, zeroArtifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build dynamic-zero witness diagnostics=%+v err=%v", diagnostics, err)
	}
	zeroResult, diagnostics, err := session.RunInterpreterApplication(zeroSource, "0")
	if err != nil || len(diagnostics) != 0 || zeroResult.Execution.Outcome.Kind != execution.OutcomeDefect || len(zeroResult.Output) != 0 || !phase27HasDefectReason(zeroResult.Execution.Events, "U64 remainder by zero") {
		t.Fatalf("dynamic-zero interpreter diagnostics=%+v execution=%+v output=%q err=%v", diagnostics, zeroResult.Execution, zeroResult.Output, err)
	}
	var zeroStdout, zeroStderr bytes.Buffer
	zeroOutcome, err := native.RunApplication(context.Background(), zeroArtifact, "0", &zeroStdout, &zeroStderr)
	if err != nil || zeroOutcome.Kind != native.RunExited || zeroOutcome.ExitCode != 65 || zeroStdout.Len() != 0 || zeroStderr.String() != "schway: U64 remainder by zero\n" {
		t.Fatalf("dynamic-zero native outcome=%+v stdout=%q stderr=%q err=%v", zeroOutcome, zeroStdout.String(), zeroStderr.String(), err)
	}
}

func TestPhase27RepeatedEventIdentity(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "fizzbuzz.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("FizzBuzz check diagnostics: %+v", checked.Diagnostics)
	}
	var copyID string
	for _, function := range checked.Program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCopy {
				copyID = operation.ID + ":event"
				break
			}
		}
	}
	if copyID == "" {
		t.Fatal("FizzBuzz source has no loop snapshot copy event")
	}
	for _, test := range []struct {
		input string
		want  []uint64
	}{{"0", []uint64{}}, {"1", []uint64{0}}, {"2", []uint64{0, 1}}} {
		result, err := interp.RunApplication(checked.Program, "main", test.input)
		if err != nil || result.Execution.Outcome.Kind != execution.OutcomeReturned {
			t.Fatalf("input=%s interpreter execution=%+v err=%v", test.input, result.Execution, err)
		}
		var got []uint64
		var invocation string
		for _, event := range result.Execution.Events {
			if event.ID != copyID {
				continue
			}
			if invocation == "" {
				invocation = event.Invocation
			} else if invocation != event.Invocation {
				t.Fatalf("one FizzBuzz entry copy site changed invocation identity: %q then %q", invocation, event.Invocation)
			}
			got = append(got, event.Occurrence)
		}
		if len(got) != len(test.want) {
			t.Fatalf("input=%s copy occurrences=%v, want %v", test.input, got, test.want)
		}
		for index := range got {
			if got[index] != test.want[index] {
				t.Fatalf("input=%s copy occurrences=%v, want %v", test.input, got, test.want)
			}
		}
	}
}

func TestPhase27LegacyDocumentCompatibility(t *testing.T) {
	sumSource, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	cases := session.ReplayCases{
		Schema: session.ReplayCasesSchema,
		Cases: []session.ReplayCase{{
			ID: "sum-to-ten", Kind: session.ReplayCaseSource, Input: "10",
			Expected: session.ReplayExpected{Kind: execution.OutcomeReturned, Value: "55"}, ForeignOutcomes: []session.ReplayForeignOutcome{},
		}},
	}
	report, err := session.VerifyApplicationCases(context.Background(), sumSource, cases, native.DefaultRunner())
	if err != nil || !report.Verified || report.Status != session.ReplayStatusPass || len(report.Cases) != 1 || report.Cases[0].Status != session.ReplayStatusPass {
		t.Fatalf("legacy U64 execution-document verification report=%+v err=%v", report, err)
	}
	for _, engine := range []string{"interpreter", "O0", "O3"} {
		if actual, ok := report.Cases[0].Engines[engine]; !ok || actual.Outcome.Kind != execution.OutcomeReturned || actual.Outcome.Value != "55" {
			t.Fatalf("legacy execution document lacks expected %s value: %+v", engine, report.Cases[0].Engines)
		}
	}

	fizzSource, err := os.ReadFile(testsupport.ProjectPath("examples", "fizzbuzz.schway"))
	if err != nil {
		t.Fatal(err)
	}
	fizz, diagnostics, err := session.RunInterpreterApplication(fizzSource, "2")
	if err != nil || len(diagnostics) != 0 || fizz.Execution.Outcome.Kind != execution.OutcomeReturned || fizz.Execution.Outcome.Value != "" || !bytes.Equal(fizz.Output, []byte("1\n2\n")) {
		t.Fatalf("Unit app output surface diagnostics=%+v execution=%+v output=%q err=%v", diagnostics, fizz.Execution, fizz.Output, err)
	}
	for _, event := range fizz.Execution.Events {
		if strings.Contains(event.Output, "application output") || event.Kind == "application.wrote" {
			t.Fatalf("application bytes leaked into the execution document: %+v", event)
		}
	}
}

func phase27HasDefectReason(events []execution.Event, reason string) bool {
	for _, event := range events {
		if event.Kind == "function.defected" && event.Output == reason {
			return true
		}
	}
	return false
}

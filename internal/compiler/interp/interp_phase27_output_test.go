package interp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase27OutputExactLimit(t *testing.T) {
	fizzBuzz := checkedPhase27OutputProgram(t, "fizzbuzz.schway")
	fizzResult, err := runApplicationWithOutputLimit(fizzBuzz, "main", "2", 4)
	if err != nil || fizzResult.Execution.Outcome.Kind != execution.OutcomeReturned || !bytes.Equal(fizzResult.Output, []byte("1\n2\n")) {
		t.Fatalf("FizzBuzz exact-limit result=%+v err=%v, want returned and four exact bytes", fizzResult, err)
	}

	sum := checkedPhase27OutputProgram(t, "sum_to_n.schway")
	sumResult, err := runApplicationWithOutputLimit(sum, "main", "10", 3)
	if err != nil || sumResult.Execution.Outcome.Kind != execution.OutcomeReturned || !bytes.Equal(sumResult.Output, []byte("55\n")) {
		t.Fatalf("sum exact-limit result=%+v err=%v, want returned 55 LF at three bytes", sumResult, err)
	}
}

func TestPhase27OutputCrossingWrite(t *testing.T) {
	fizzBuzz := checkedPhase27OutputProgram(t, "fizzbuzz.schway")
	fizzResult, err := runApplicationWithOutputLimit(fizzBuzz, "main", "3", 4)
	if err != nil || fizzResult.Execution.Outcome.Kind != execution.OutcomeDefect || !bytes.Equal(fizzResult.Output, []byte("1\n2\n")) {
		t.Fatalf("FizzBuzz crossing result=%+v err=%v, want defect after the completed four-byte prefix", fizzResult, err)
	}
	assertPhase27OutputLimitDefect(t, fizzResult.Execution)

	sum := checkedPhase27OutputProgram(t, "sum_to_n.schway")
	sumResult, err := runApplicationWithOutputLimit(sum, "main", "10", 2)
	if err != nil || sumResult.Execution.Outcome.Kind != execution.OutcomeDefect || !bytes.Equal(sumResult.Output, []byte("55")) {
		t.Fatalf("sum crossing result=%+v err=%v, want decimal chunk then failed LF", sumResult, err)
	}
	assertPhase27OutputLimitDefect(t, sumResult.Execution)
}

func TestPhase27OutputProductionLimit(t *testing.T) {
	if execution.MaxApplicationOutputBytes != 65536 {
		t.Fatalf("production application output limit=%d, want 65536", execution.MaxApplicationOutputBytes)
	}
}

func checkedPhase27OutputProgram(t *testing.T, name string) core.Program {
	t.Helper()
	sourcePath := filepath.Join(interpProjectRoot(), "examples", name)
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("%s parse diagnostics: %+v", name, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s check diagnostics: %+v", name, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("%s core validation: %+v", name, validated.Problems)
	}
	return validated.Program()
}

func assertPhase27OutputLimitDefect(t *testing.T, result Execution) {
	t.Helper()
	if result.Outcome.Kind != execution.OutcomeDefect || len(result.Events) == 0 {
		t.Fatalf("output-limit execution=%+v, want terminal defect event", result)
	}
	event := result.Events[len(result.Events)-1]
	if event.Kind != "function.defected" || event.Output != "application output limit exceeded" {
		t.Fatalf("output-limit terminal event=%+v, want stable output-limit reason", event)
	}
}

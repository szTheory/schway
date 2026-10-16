package cgen_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase371NoCallScalarDefectNativeConformance(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sum_to_n checker diagnostics: %+v", checked.Diagnostics)
	}
	var entry core.Function
	var defectFunction core.Function
	var defect core.LinearOperation
	calls := 0
	for _, function := range checked.Program.Functions {
		if function.Name == "main" {
			entry = function
		}
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall {
				calls++
			}
			if operation.Kind == core.OpDefect {
				defectFunction = function
				defect = operation
			}
		}
	}
	if calls != 0 {
		t.Fatalf("checked sum_to_n contains %d OpCall operations, want a permanent no-call witness", calls)
	}
	if entry.ID == "" || defectFunction.ID != entry.ID || defect.ID == "" {
		t.Fatalf("checked entry/defect identity missing: entry=%+v defect_function=%q defect=%+v", entry, defectFunction.ID, defect)
	}
	const wantReason = "input exceeds 1000"
	if defect.Reason != wantReason {
		t.Fatalf("checked defect reason=%q, want literal %q", defect.Reason, wantReason)
	}
	assertDefect := func(t *testing.T, run execution.Execution, schema string) {
		t.Helper()
		if run.Schema != schema || run.Outcome.Kind != execution.OutcomeDefect || run.Outcome.Value != "" || run.Events == nil || run.LiveResources == nil || len(run.LiveResources) != 0 {
			t.Fatalf("execution=%+v, want one complete %s defect document without a value or live resources", run, schema)
		}
		terminal := run.Events[len(run.Events)-1]
		if terminal.Kind != "function.defected" || terminal.ID != defect.ID+":event:defected" || terminal.FunctionID != defectFunction.ID || terminal.Output != wantReason || terminal.SourcePlace != defect.SourceID || terminal.TypeID != defect.TypeID {
			t.Fatalf("terminal event=%+v, want checked identity and literal reason %q", terminal, wantReason)
		}
	}
	interpreted, err := interp.Run(checked.Program, "main", "1001")
	if err != nil {
		t.Fatalf("interpreter defect: %v", err)
	}
	assertDefect(t, interpreted, execution.Schema1)

	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("native emission: %v", err)
	}
	for _, lane := range []struct {
		name, optimization string
		lto                bool
	}{{"O0", "-O0", false}, {"O3", "-O3", false}, {"O3-LTO", "-O3", true}} {
		t.Run(lane.name, func(t *testing.T) {
			runner := native.DefaultRunner()
			runner.LTO = lane.lto
			runner.Expect = native.ExpectDefect
			result, runErr := runner.Run(context.Background(), generated, lane.optimization, []string{"1001"})
			if runErr != nil {
				var toolErr *native.ToolError
				if errors.As(runErr, &toolErr) && toolErr.Code == "native.run_failed" && strings.Contains(runErr.Error(), "exit status 65") {
					t.Fatalf("native conformance received application defect exit 65 instead of a structured defect record: %v", runErr)
				}
				t.Fatalf("native defect run failed before its semantic assertion: %v", runErr)
			}
			if len(result.Pairs) != 1 || result.Pairs[0].Input != "1001" {
				t.Fatalf("native defect result pairs=%+v, want exactly one complete record for input 1001", result.Pairs)
			}
			assertDefect(t, result.Pairs[0].Execution, execution.Schema2)
			if err := executionpeer.Validate(checked.Program, result.Pairs[0].Execution); err != nil {
				t.Fatalf("independent peer rejected checked no-call defect: %v", err)
			}
		})
	}

	valueRunner := native.DefaultRunner()
	valueRunner.Expect = native.ExpectValue
	valueResult, err := valueRunner.Run(context.Background(), generated, "-O0", []string{"5"})
	if err != nil || len(valueResult.Pairs) != 1 || valueResult.Pairs[0].Execution.Outcome.Value != "15" {
		t.Fatalf("native sum_to_n(5)=%+v err=%v, want returned 15", valueResult, err)
	}
	interpretedValue, err := interp.Run(checked.Program, "main", "5")
	if err != nil || interpretedValue.Outcome.Kind != execution.OutcomeReturned || interpretedValue.Outcome.Value != "15" {
		t.Fatalf("interpreter sum_to_n(5)=%+v err=%v, want returned 15", interpretedValue, err)
	}
	if err := executionpeer.Validate(checked.Program, valueResult.Pairs[0].Execution); err != nil {
		t.Fatalf("independent peer rejected native returned-value control: %v", err)
	}

	appSource := checked.Program
	appC, err := cgen.EmitApplication(appSource)
	if err != nil {
		t.Fatalf("application emission: %v", err)
	}
	artifact := filepath.Join(t.TempDir(), "sum-to-n-app")
	if _, err := native.BuildApplication(context.Background(), source, appC, artifact); err != nil {
		t.Fatalf("application build: %v", err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "1001", &stdout, &stderr)
	if err != nil || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway app: input exceeds 1000\n" {
		t.Fatalf("application outcome=%+v stdout=%q stderr=%q err=%v, want empty stdout, exit 65 and exact reason", outcome, stdout.String(), stderr.String(), err)
	}

	unitChecked := session.Check([]byte("module phase371.unit_defect\nexport { fn main }\nfn main(input: U64) -> Unit { write_u64(input) }\n"))
	if len(unitChecked.Diagnostics) != 0 {
		t.Fatalf("Unit refusal control was not checker-admitted: %+v", unitChecked.Diagnostics)
	}
	if _, err := cgen.EmitNative(unitChecked.Program); err == nil || !strings.Contains(err.Error(), "Unit C return is supported only by the application shell") {
		t.Fatalf("native Unit conformance refusal changed: %v", err)
	}
}

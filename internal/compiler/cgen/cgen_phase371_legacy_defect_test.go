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

func TestPhase371CountedDefectReasonBytes(t *testing.T) {
	reasons := []struct {
		name  string
		value string
	}{
		{name: "ascii", value: "plain ASCII reason"},
		{name: "control_then_hex_digit", value: "\x01A"},
		{name: "nul_then_a", value: "\x00A"},
		{name: "nul_then_z", value: "\x00Z"},
		{name: "other_controls", value: "\x02\r\t\x7f"},
		{name: "non_ascii", value: "café 雪"},
		{name: "replacement_character", value: "\uFFFD"},
		{name: "literal_backslash_escapes", value: `\u0000\x41`},
	}
	witnesses := []struct {
		name string
		make func([]byte) []byte
	}{
		{name: "no_call", make: phase371NoCallReasonSource},
		{name: "dynamic_entry", make: func(reason []byte) []byte { return phase371DynamicReasonSource("entry", reason) }},
		{name: "dynamic_helper", make: func(reason []byte) []byte { return phase371DynamicReasonSource("helper", reason) }},
	}
	lanes := []struct {
		name, optimization string
		lto                bool
	}{{"O0", "-O0", false}, {"O3", "-O3", false}, {"O3-LTO", "-O3", true}}
	for _, witness := range witnesses {
		t.Run(witness.name, func(t *testing.T) {
			for reasonIndex, reason := range reasons {
				t.Run(reason.name, func(t *testing.T) {
					source := witness.make([]byte(reason.value))
					checked := session.Check(source)
					if len(checked.Diagnostics) != 0 {
						t.Fatalf("valid UTF-8 reason source refused: %+v", checked.Diagnostics)
					}
					defect, functionID := phase371FindDefect(checked.Program)
					if defect.ID == "" || defect.Reason != reason.value {
						t.Fatalf("checked reason=%q, want exact %q", defect.Reason, reason.value)
					}
					callCount := 0
					for _, function := range checked.Program.Functions {
						if function.Linear == nil {
							continue
						}
						for _, operation := range function.Linear.Operations {
							if operation.Kind == core.OpCall {
								callCount++
							}
						}
					}
					if witness.name == "no_call" && callCount != 0 {
						t.Fatalf("no-call witness has %d checked OpCall operations", callCount)
					}
					if witness.name != "no_call" && callCount == 0 {
						t.Fatal("dynamic witness did not check any OpCall operation")
					}
					input := "7"
					if witness.name == "no_call" {
						input = "1001"
					}
					interpreted, err := interp.Run(checked.Program, "main", input)
					if err != nil || interpreted.Outcome.Kind != execution.OutcomeDefect || len(interpreted.Events) == 0 || interpreted.Events[len(interpreted.Events)-1].Output != reason.value {
						t.Fatalf("interpreter did not preserve source reason: execution=%+v err=%v", interpreted, err)
					}
					generated, err := cgen.EmitNative(checked.Program)
					if err != nil {
						t.Fatalf("EmitNative: %v", err)
					}
					for _, lane := range lanes {
						// Keep every optimization tier visible for the hex-escape and NUL cases;
						// the remaining valid reasons exercise the same counted path at O0.
						if reasonIndex > 3 && lane.name != "O0" {
							continue
						}
						t.Run(lane.name, func(t *testing.T) {
							runner := native.DefaultRunner()
							runner.LTO = lane.lto
							runner.Expect = native.ExpectDefect
							result, runErr := runner.Run(context.Background(), generated, lane.optimization, []string{input})
							if runErr != nil || len(result.Pairs) != 1 {
								t.Fatalf("native %s defect: pairs=%d err=%v", lane.name, len(result.Pairs), runErr)
							}
							run := result.Pairs[0].Execution
							if run.Schema != execution.Schema2 || run.Outcome.Kind != execution.OutcomeDefect || len(run.Events) == 0 || run.Events[len(run.Events)-1].Output != reason.value {
								t.Fatalf("native %s lost reason bytes: %+v, want %q", lane.name, run, reason.value)
							}
							if run.Events[len(run.Events)-1].FunctionID != functionID || run.Events[len(run.Events)-1].ID != defect.ID+":event:defected" {
								t.Fatalf("native terminal identity=%+v, want checked function/op", run.Events[len(run.Events)-1])
							}
							if err := executionpeer.Validate(checked.Program, run); err != nil {
								t.Fatalf("independent peer rejected native reason witness: %v", err)
							}
						})
					}
					appC, err := cgen.EmitApplication(checked.Program)
					if err != nil {
						t.Fatalf("EmitApplication: %v", err)
					}
					artifact := filepath.Join(t.TempDir(), "reason-app")
					if _, err := native.BuildApplication(context.Background(), source, appC, artifact); err != nil {
						t.Fatalf("BuildApplication: %v", err)
					}
					var stdout, stderr bytes.Buffer
					outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
					wantStderr := append([]byte("schway app: "), []byte(reason.value)...)
					wantStderr = append(wantStderr, '\n')
					if err != nil || outcome.ExitCode != 65 || stdout.Len() != 0 || !bytes.Equal(stderr.Bytes(), wantStderr) {
						t.Fatalf("application outcome=%+v stdout=%q stderr=%q err=%v, want empty stdout, exit65 and exact %q", outcome, stdout.Bytes(), stderr.Bytes(), err, wantStderr)
					}
				})
			}
		})
	}

	// Exercise the same counted structured-output path under ASan/UBSan. The
	// report records whether this host supplied a working sanitizer runtime.
	sanitizerSource := phase371NoCallReasonSource([]byte("\x00A"))
	checked := session.Check(sanitizerSource)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sanitizer source refused: %+v", checked.Diagnostics)
	}
	sanitizerC, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("sanitizer EmitNative: %v", err)
	}
	sanitizerRunner := native.DefaultRunner()
	probe, err := sanitizerRunner.ProbeSanitizerRuntime(context.Background())
	if err != nil {
		var toolErr *native.ToolError
		if errors.As(err, &toolErr) && toolErr.Code == "native.tool_missing" {
			t.Logf("focused sanitizer host availability: unavailable (%s)", toolErr.Code)
			return
		} else {
			t.Fatalf("focused sanitizer runtime probe failed: report=%+v err=%v", probe, err)
		}
	}
	if probe.CheckKind != "heap-use-after-free" || probe.Sanitizer != "address" {
		t.Fatalf("sanitizer host probe=%+v, want address-sanitizer smoke finding", probe)
	}
	report, err := sanitizerRunner.RunSanitized(context.Background(), sanitizerC, []string{"1001"})
	if err != nil || report.CheckKind != "" || report.ExitCode != 0 {
		t.Fatalf("focused sanitizer report=%+v err=%v, want clean counted reason path", report, err)
	}
	t.Logf("focused sanitizer host availability: available (smoke=%s/%s); counted reason path clean (exit=%d)", probe.Sanitizer, probe.CheckKind, report.ExitCode)
}

func phase371FindDefect(program core.Program) (core.LinearOperation, string) {
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpDefect {
				return operation, function.ID
			}
		}
	}
	return core.LinearOperation{}, ""
}

func phase371NoCallReasonSource(reason []byte) []byte {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		panic(err)
	}
	return bytes.Replace(source, []byte("input exceeds 1000"), reason, 1)
}

func phase371DynamicReasonSource(location string, reason []byte) []byte {
	prefix := []byte("module phase371.dynamic_reason\nexport { fn main }\n")
	if location == "entry" {
		prefix = append(prefix, []byte("fn identity(input: U64) -> U64 { input }\nfn main(input: U64) -> U64 {\n var result = identity(input)\n if input == 7 { defect \"")...)
		prefix = append(prefix, reason...)
		return append(prefix, []byte("\" } else { result = result + 1 }\n result\n}\n")...)
	}
	prefix = append(prefix, []byte("fn stop(input: U64) -> U64 {\n var result = input\n if input == 7 { defect \"")...)
	prefix = append(prefix, reason...)
	prefix = append(prefix, []byte("\" } else { result = result + 1 }\n result\n}\nfn main(input: U64) -> U64 { let result = stop(input)\n result\n}\n")...)
	return prefix
}

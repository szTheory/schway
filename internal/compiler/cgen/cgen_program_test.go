package cgen_test

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPublicDispatchUsesOnlyEmitProgram parses the production source so the
// sole-route claim cannot be satisfied by a comment or an incidental runtime
// result. Each public API may validate first, but its only emitter call is the
// whole-program implementation.
func TestPublicDispatchUsesOnlyEmitProgram(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testsupport.ProjectPath("internal", "compiler", "cgen", "cgen.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Emit", "EmitNative"} {
		var declaration *ast.FuncDecl
		for _, candidate := range file.Decls {
			function, ok := candidate.(*ast.FuncDecl)
			if ok && function.Name.Name == name {
				declaration = function
				break
			}
		}
		if declaration == nil {
			t.Fatalf("missing public dispatcher %s", name)
		}
		var emitterCalls []string
		ast.Inspect(declaration.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if identifier, ok := call.Fun.(*ast.Ident); ok && len(identifier.Name) >= 4 && identifier.Name[:4] == "emit" {
				emitterCalls = append(emitterCalls, identifier.Name)
			}
			return true
		})
		if len(emitterCalls) != 1 || emitterCalls[0] != "emitProgram" {
			t.Fatalf("%s emitter calls = %v, want only emitProgram", name, emitterCalls)
		}
	}
}

// TestPhase21LegacyEmitterBodiesRetired keeps the three cut private lowering
// implementations out of production source. The live whole-program emitter
// retains the refusal boundary for those families.
func TestPhase21LegacyEmitterBodiesRetired(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testsupport.ProjectPath("internal", "compiler", "cgen", "cgen.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	retired := map[string]bool{
		"emitLinearForeign":                true,
		"emitLinearBorrowedByPointer":      true,
		"emitLinearBorrowedByPointerPlain": true,
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil {
			continue
		}
		if retired[function.Name.Name] {
			t.Errorf("retired lowering body %s remains in production cgen.go", function.Name.Name)
		}
	}
}

// checkedPhase07Program parses the tracked adversarial fixture through the
// normal checker and independent core peer; the 61-node assertion below is
// therefore derived from the checked source fixture, not restated data.
func checkedPhase07Program(t *testing.T, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("%s: corevalidate rejected: %+v", fixture, validated.Problems)
	}
	return validated.Program()
}

// TestProgramOrdinaryLinearTracer drives the existing whole-program emitter
// directly on the phase-2 tracked-transfer fixture.  It deliberately avoids
// the public EmitNative dispatcher: this is the N=1 tracer for the surviving
// emitter law, including its schema-2 event/invocation machinery.
func TestProgramOrdinaryLinearTracer(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("owned_transfer.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("owned_transfer.schway: corevalidate rejected: %+v", validated.Problems)
	}
	program := validated.Program()

	generated, err := cgen.EmitProgramForTest(program)
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"01020304"})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("direct emitProgram native run: pairs=%d err=%v", len(got.Pairs), err)
	}
	want, err := interp.Run(program, "relay", "01020304")
	if err != nil {
		t.Fatalf("interpreter: %v", err)
	}
	// The interpreter remains schema-1 for this legacy fixture. The program
	// emitter's schema-2 contract adds the entry invocation identity without
	// changing the underlying ordinary operation sequence.
	want.Schema = execution.Schema2
	invocation, err := execution.FormatInvocation(program.Functions[0].ID, nil)
	if err != nil {
		t.Fatalf("format entry invocation: %v", err)
	}
	for index := range want.Events {
		want.Events[index].Schema = execution.Schema2
		want.Events[index].Invocation = invocation
	}
	if got.Pairs[0].Execution.Schema != execution.Schema2 {
		t.Fatalf("direct emitProgram schema=%q, want %q", got.Pairs[0].Execution.Schema, execution.Schema2)
	}
	if !execution.Equal(want, got.Pairs[0].Execution) {
		t.Fatalf("direct emitProgram differs from interpreter:\nwant: %+v\ngot:  %+v", want, got.Pairs[0].Execution)
	}
}

func TestPhase19U64NativeExactWidthAndOutput(t *testing.T) {
	const maximum = "18446744073709551615"
	fixture, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"0", maximum} {
		t.Run(value, func(t *testing.T) {
			source := strings.Replace(string(fixture), "= 42", "= "+value, 1)
			checked := session.Check([]byte(source))
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
			}
			validated := corevalidate.Validate(checked.Program)
			if !validated.Valid {
				t.Fatalf("corevalidate rejected: %+v", validated.Problems)
			}
			generated, err := cgen.EmitNative(validated.Program())
			if err != nil {
				t.Fatalf("EmitNative: %v", err)
			}
			for _, fragment := range []string{"#include <stdint.h>", "uint64_t", "UINT64_C(" + value + ")", "#if !defined(UINT64_MAX)"} {
				if !strings.Contains(generated, fragment) {
					t.Fatalf("generated C does not contain %q", fragment)
				}
			}
			got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"7"})
			if err != nil || len(got.Pairs) != 1 {
				t.Fatalf("native run: pairs=%d err=%v", len(got.Pairs), err)
			}
			if got.Pairs[0].Execution.Outcome.Value != value {
				t.Fatalf("native result = %q, want canonical decimal string %q", got.Pairs[0].Execution.Outcome.Value, value)
			}
		})
	}
}

func TestPhase19ExactWidthTargetGuardRejectsMissingMacro(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	generated = strings.Replace(generated, "#include <stdint.h>\n", "#include <stdint.h>\n#undef UINT64_MAX\n", 1)
	if _, cleanup, err := native.DefaultRunner().CompileOnly(context.Background(), generated, "-O0"); err == nil {
		cleanup()
		t.Fatal("expected deterministic missing-exact-width-macro control to fail compilation")
	} else if !strings.Contains(err.Error(), "Schway U64 requires exact-width uint64_t support") {
		t.Fatalf("compile error does not identify exact-width guard: %v", err)
	}
}

func TestPhase19U64NativeEntryInput(t *testing.T) {
	const maximum = "18446744073709551615"
	source := []byte("module phase19.u64_input\n\nexport { fn main }\n\nfn main(input: U64) -> U64 { input }\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{maximum})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("native max input: pairs=%d err=%v", len(got.Pairs), err)
	}
	if got.Pairs[0].Execution.Outcome.Value != maximum {
		t.Fatalf("native result = %q, want %q", got.Pairs[0].Execution.Outcome.Value, maximum)
	}
}

// TestProgramBranchTracer drives the checked branch fixture through the
// surviving whole-program writer. Both switch alternatives are exercised, and
// arm-local linear operations share the schema-2 event buffer rather than
// producing an arm-local document.
func TestProgramBranchTracer(t *testing.T) {
	tests := []struct {
		fixture  string
		function string
		inputs   []string
	}{
		{fixture: "borrowed_view.schway", function: "choose", inputs: []string{"On", "Off"}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("%s: unexpected diagnostics: %+v", test.fixture, checked.Diagnostics)
			}
			program := checked.Program
			generated, err := cgen.EmitProgramForTest(program)
			if err != nil {
				t.Fatalf("direct emitProgram: %v", err)
			}
			invocation, err := execution.FormatInvocation(program.Functions[0].ID, nil)
			if err != nil {
				t.Fatalf("format entry invocation: %v", err)
			}
			for _, input := range test.inputs {
				got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{input})
				if err != nil || len(got.Pairs) != 1 {
					t.Fatalf("direct emitProgram native run input %q: pairs=%d err=%v", input, len(got.Pairs), err)
				}
				want, err := interp.Run(program, test.function, input)
				if err != nil {
					t.Fatalf("interpreter input %q: %v", input, err)
				}
				want.Schema = execution.Schema2
				for index := range want.Events {
					want.Events[index].Schema = execution.Schema2
					want.Events[index].Invocation = invocation
				}
				if !execution.Equal(want, got.Pairs[0].Execution) {
					t.Fatalf("direct emitProgram input %q differs from interpreter:\nwant: %+v\ngot:  %+v", input, want, got.Pairs[0].Execution)
				}
			}
		})
	}
}

// TestProgramMatchPayloadLowering drives the checked Phase 12 payload tracer
// through the surviving program emitter. Both alternatives must retain their
// construct/destructure events and receive the schema-2 invocation identity.
func TestProgramMatchPayloadLowering(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase12", "payload_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("payload_tracer.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	program := checked.Program
	generated, err := cgen.EmitProgramForTest(program)
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	invocation, err := execution.FormatInvocation(program.Functions[0].ID, nil)
	if err != nil {
		t.Fatalf("format entry invocation: %v", err)
	}
	for _, input := range []string{"Ok", "Err"} {
		got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{input})
		if err != nil || len(got.Pairs) != 1 {
			t.Fatalf("native run input %q: pairs=%d err=%v", input, len(got.Pairs), err)
		}
		want, err := interp.Run(program, "identity", input)
		if err != nil {
			t.Fatalf("interpreter input %q: %v", input, err)
		}
		want.Schema = execution.Schema2
		for index := range want.Events {
			want.Events[index].Schema = execution.Schema2
			want.Events[index].Invocation = invocation
		}
		if !execution.Equal(want, got.Pairs[0].Execution) {
			t.Fatalf("input %q differs from interpreter:\nwant: %+v\ngot:  %+v", input, want, got.Pairs[0].Execution)
		}
	}
}

// TestProgramMatchUsesCheckerLayout proves that the emitted tagged record is
// a projection of check.PayloadRecordLayout rather than a local layout law.
func TestProgramMatchUsesCheckerLayout(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase12", "payload_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("payload_tracer.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitProgramForTest(checked.Program)
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	var outcome core.DataType
	for _, dataType := range checked.Program.DataTypes {
		if dataType.Name == "Outcome" {
			outcome = dataType
			break
		}
	}
	layout := check.PayloadRecordLayout(outcome)
	for _, field := range layout.Fields {
		if !strings.Contains(generated, field.CType+" "+field.Name+";") {
			t.Fatalf("generated tagged record omits checker field %s %s:\n%s", field.CType, field.Name, generated)
		}
	}
	if strings.Index(generated, layout.Fields[1].Name) > strings.Index(generated, layout.Fields[2].Name) {
		t.Fatalf("generated tagged record does not preserve checker field order:\n%s", generated)
	}
}

// TestProgramMatchDefectEventPrecedesAbort verifies the schema-2 terminal
// record is visible to the native runner before the defect helper aborts.
func TestProgramMatchDefectEventPrecedesAbort(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "defect_terminal.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("defect_terminal.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitProgramForTest(checked.Program)
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	runner := native.DefaultRunner()
	runner.Expect = native.ExpectDefect
	got, err := runner.Run(context.Background(), generated, "-O0", []string{"Halt"})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("defect native run: pairs=%d err=%v", len(got.Pairs), err)
	}
	events := got.Pairs[0].Execution.Events
	if len(events) == 0 || events[len(events)-1].Kind != "function.defected" || events[len(events)-1].Output != "halt requested" {
		t.Fatalf("defect event was not the terminal visible event: %+v", events)
	}
}

// TestProgramNoreturnExemptionIsNarrow prevents attributes from spreading
// beyond the generated abort-only defect helper.
func TestProgramNoreturnExemptionIsNarrow(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "defect_terminal.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("defect_terminal.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitProgramForTest(checked.Program)
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	if strings.Count(generated, "_Noreturn") != 1 || !strings.Contains(generated, "_Noreturn static void schway_defect") {
		t.Fatalf("_Noreturn must appear only on schway_defect:\n%s", generated)
	}
	if strings.Index(generated, "schway_write_literal(\"{\\\"schema\\\":\\\"lang.execution/2") > strings.Index(generated, "schway_defect(\"halt requested\")") {
		t.Fatalf("defect document must be emitted before schway_defect:\n%s", generated)
	}
}

// TestProgramLiveResourcesAreDerived rejects a fixed resource-tail literal:
// even though ordinary linear programs presently derive no live resources,
// their schema-2 document must be rendered from the surviving emitter's
// derived collection.
func TestProgramLiveResourcesAreDerived(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("owned_transfer.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("owned_transfer.schway: corevalidate rejected: %+v", validated.Problems)
	}
	generated, err := cgen.EmitProgramForTest(validated.Program())
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	if !strings.Contains(generated, "schway_write_live_resources") {
		t.Fatalf("schema-2 resource tail bypasses a derived-value writer:\n%s", generated)
	}
}

// TestProgramLiveResourceDerivationIsNotInert mutates the emitter-owned
// derivation and observes the resulting native document, so a future tail
// shortcut cannot leave the derivation present but unused.
func TestProgramLiveResourceDerivationIsNotInert(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", "owned_transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("owned_transfer.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("owned_transfer.schway: corevalidate rejected: %+v", validated.Problems)
	}
	restore := cgen.SetProgramLiveResourcesForTest([]string{"resource:seed"})
	defer restore()
	generated, err := cgen.EmitProgramForTest(validated.Program())
	if err != nil {
		t.Fatalf("direct emitProgram: %v", err)
	}
	if !strings.Contains(generated, "resource:seed") {
		t.Fatalf("seeded derived resources did not move generated serialization:\n%s", generated)
	}
}

// syntheticInvocationProgram produces valid, straight-line source whose
// shared-callee unfolding is exact: node11 has 4095 occurrences; root adds
// either one child (4096 total) or node0 as its second child (4097 total).
func syntheticInvocationProgram(t *testing.T, nodes int) core.Program {
	t.Helper()
	if nodes != 4096 && nodes != 4097 {
		t.Fatalf("syntheticInvocationProgram only models the pinned boundaries, got %d", nodes)
	}
	var source strings.Builder
	source.WriteString("module phase15.synthetic\n\nexport {\n  fn root\n}\n\n")
	source.WriteString("fn node0(value: Byte) -> Byte {\n  value\n}\n\n")
	for level := 1; level <= 11; level++ {
		fmt.Fprintf(&source, "fn node%d(value: Byte) -> Byte {\n  let first = node%d(value)\n  let second = node%d(value)\n  second\n}\n\n", level, level-1, level-1)
	}
	source.WriteString("fn root(value: Byte) -> Byte {\n  let first = node11(value)\n")
	if nodes == 4097 {
		source.WriteString("  let second = node0(value)\n  second\n")
	} else {
		source.WriteString("  first\n")
	}
	source.WriteString("}\n")
	checked := session.Check([]byte(source.String()))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("synthetic %d-node source unexpectedly refused: %+v\n%s", nodes, checked.Diagnostics, source.String())
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("synthetic %d-node source failed core validation: %+v", nodes, validated.Problems)
	}
	return validated.Program()
}

func TestInvocationPathTableDeepDiamondMeasures61(t *testing.T) {
	program := checkedPhase07Program(t, "deep_diamond_acyclic.schway")
	got, err := cgen.InvocationPathNodeCountForTest(program)
	if err != nil {
		t.Fatalf("preflight deep_diamond_acyclic.schway: %v", err)
	}
	if got != 61 {
		t.Fatalf("deep_diamond_acyclic.schway invocation nodes = %d, want 61 (T0=1, T1=5, T2=13, T3=29, T4=61)", got)
	}
	if _, err := cgen.EmitNative(program); err != nil {
		t.Fatalf("EmitNative after 61-node preflight: %v", err)
	}
}

// TestDeepDiamondExecutesAcrossNativeOptimizationTiers is the runtime
// counterpart to the compile-only 61-occurrence preflight assertion above.
// Its oracle is derived from the interpreter: every native tier must retain
// the full occurrence-weighted event stream, not merely avoid crashing.
func TestDeepDiamondExecutesAcrossNativeOptimizationTiers(t *testing.T) {
	program := checkedPhase07Program(t, "deep_diamond_acyclic.schway")
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatal(err)
	}
	want, err := interp.Run(program, entry.Name, "7")
	if err != nil {
		t.Fatalf("interp.Run: %v", err)
	}
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	if wantCapacity := fmt.Sprintf("#define SCHWAY_EVENT_CAPACITY %du", len(want.Events)); !strings.Contains(generated, wantCapacity) {
		t.Fatalf("generated capacity is not occurrence-weighted; want %q", wantCapacity)
	}
	wantBytes, err := cgen.ExecutionOutputSizeForTest(program)
	if err != nil {
		t.Fatal(err)
	}
	for _, optimization := range []string{"-O0", "-O3"} {
		t.Run(optimization, func(t *testing.T) {
			got, runErr := native.DefaultRunner().Run(context.Background(), generated, optimization, []string{"7"})
			if runErr != nil || len(got.Pairs) != 1 {
				t.Fatalf("native run: pairs=%d err=%v", len(got.Pairs), runErr)
			}
			if !execution.Equal(want, got.Pairs[0].Execution) {
				t.Fatalf("native evidence differs from interpreter: want %d events, got %d", len(want.Events), len(got.Pairs[0].Execution.Events))
			}
			if got.OutputBytes != wantBytes {
				t.Fatalf("native output bytes=%d, preflight estimate=%d", got.OutputBytes, wantBytes)
			}
		})
	}
}

// TestSchema2ExecutionOutputBoundIsPreflighted pins the separately diagnosed
// byte contract at N-1/N. The fixture is intentionally larger than the legacy
// generic process-stream cap, so acceptance cannot silently inherit 64 KiB.
func TestSchema2ExecutionOutputBoundIsPreflighted(t *testing.T) {
	program := checkedPhase07Program(t, "deep_diamond_acyclic.schway")
	observed, err := cgen.ExecutionOutputSizeForTest(program)
	if err != nil {
		t.Fatal(err)
	}
	if observed <= native.MaxStreamBytes {
		t.Fatalf("fixture size %d must exercise beyond legacy stream limit %d", observed, native.MaxStreamBytes)
	}

	restore := cgen.SetExecutionOutputLimitForTest(observed - 1)
	_, err = cgen.EmitNative(program)
	restore()
	bound, ok := cgen.ExecutionOutputExceededError(err)
	if !ok || bound.Code() != "cgen.execution_output_exceeded" || bound.Limit() != observed-1 || bound.Observed() != observed {
		t.Fatalf("want exact output-bound refusal at N-1, got %v", err)
	}
	if cgen.InvocationSerializationReachedForTest() {
		t.Fatal("output-bound refusal reached C serialization")
	}

	restore = cgen.SetExecutionOutputLimitForTest(observed)
	generated, err := cgen.EmitNative(program)
	restore()
	if err != nil {
		t.Fatalf("exact-bound document refused: %v", err)
	}
	if !strings.Contains(generated, fmt.Sprintf("#define SCHWAY_OUTPUT_LIMIT %du", observed)) {
		t.Fatalf("generated writer does not carry exact tested limit %d", observed)
	}
}

// TestSchema2PayloadOutcomeBoundIsPreflighted keeps the terminal payload
// writer's maximum representation inside the size preflight contract. With a
// limit one byte below that conservative document size, emission must refuse
// before generated-C serialization begins.
func TestSchema2PayloadOutcomeBoundIsPreflighted(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase12", "payload_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("payload_tracer.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	observed, err := cgen.ExecutionOutputSizeForTest(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if observed <= 1 {
		t.Fatalf("payload document size %d is too small for a boundary control", observed)
	}

	restore := cgen.SetExecutionOutputLimitForTest(observed - 1)
	_, err = cgen.EmitNative(checked.Program)
	restore()
	bound, ok := cgen.ExecutionOutputExceededError(err)
	if !ok || bound.Limit() != observed-1 || bound.Observed() != observed {
		t.Fatalf("want payload output-bound refusal at N-1, got %v", err)
	}
	if cgen.InvocationSerializationReachedForTest() {
		t.Fatal("payload output-bound refusal reached C serialization")
	}
}

func TestInvocationPreflightOrdering(t *testing.T) {
	t.Run("call_graph_refusal_precedes_expansion", func(t *testing.T) {
		restore := cgen.SetInvocationPreflightBypassForTest(false)
		defer restore()
		program := core.Program{Functions: []core.Function{
			{ID: "fn:a", Linear: &core.LinearBody{Operations: []core.LinearOperation{{ID: "op:missing", Kind: core.OpCall, CalleeID: "fn:missing"}}}},
			{ID: "fn:b", Linear: &core.LinearBody{}},
		}}
		if _, err := cgen.EmitProgramForTest(program); err == nil {
			t.Fatal("expected call-graph refusal")
		} else if _, ok := callgraph.UnresolvedCalleeError(err); !ok {
			t.Fatalf("want callgraph unresolved-callee error, got %v", err)
		}
		if cgen.InvocationSerializationReachedForTest() {
			t.Fatal("call-graph refusal reached C serialization")
		}
	})
	t.Run("entry_refusal_precedes_expansion", func(t *testing.T) {
		restore := cgen.SetInvocationPreflightBypassForTest(false)
		defer restore()
		program := core.Program{Functions: []core.Function{{ID: "fn:a", Linear: &core.LinearBody{}}, {ID: "fn:b", Linear: &core.LinearBody{}}}}
		if _, err := cgen.EmitProgramForTest(program); err == nil {
			t.Fatal("expected entry refusal")
		} else if _, ok := callgraph.EntryAmbiguousError(err); !ok {
			t.Fatalf("want callgraph entry-ambiguity error, got %v", err)
		}
		if cgen.InvocationSerializationReachedForTest() {
			t.Fatal("entry refusal reached C serialization")
		}
	})
}

func TestUnsupportedProgramShapePrecedesSchema2Preflight(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture must check clean: %+v", checked.Diagnostics)
	}

	_, err = cgen.EmitProgramForTest(checked.Program)
	if err == nil {
		t.Fatal("expected by-pointer program shape to be refused")
	}
	if !strings.Contains(err.Error(), "by-pointer bodies are not supported by whole-program native emission this phase") {
		t.Fatalf("expected structural refusal before schema-2 preflight, got: %v", err)
	}
	if cgen.InvocationSerializationReachedForTest() {
		t.Fatal("unsupported program shape reached C serialization")
	}
}

// TestProgramBorrowedByPointerDisposition makes the whole-program admission
// boundary answer the human-reviewed Plan 16-05 decision, rather than a
// locally assumed default.  cut-m004 applies to the pointer-specialized
// family in every program cardinality: adding an ordinary caller must not
// turn the legacy single-function lowering into an admitted program shape.
func TestProgramBorrowedByPointerDisposition(t *testing.T) {
	decision, err := os.ReadFile(testsupport.ProjectPath(".planning", "milestones", "M003-phases", "16-branch-match-emitter-port", "16-05-SUMMARY.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(decision), "Selected cut-m004") {
		t.Fatalf("Plan 16-05 decision record does not select cut-m004: %q", decision)
	}

	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture must check clean: %+v", checked.Diagnostics)
	}

	program := checked.Program
	program.Functions = append(program.Functions, core.Function{
		ID: "fn:ordinary-caller",
		Linear: &core.LinearBody{Operations: []core.LinearOperation{{
			ID: "op:call-restrict", Kind: core.OpCall, CalleeID: program.Functions[0].ID,
		}}},
	})
	_, err = cgen.EmitProgramForTest(program)
	if err == nil || !strings.Contains(err.Error(), "by-pointer bodies are not supported by whole-program native emission this phase") {
		t.Fatalf("cut-m004 must refuse a by-pointer family in a multi-function program before preflight, got %v", err)
	}
	if cgen.InvocationSerializationReachedForTest() {
		t.Fatal("cut-m004 pointer refusal reached C serialization")
	}
}

// TestProgramBranchValidationOrder pins the admission law after branch bodies
// became supported: graph/entry errors win first, unsupported shapes still
// stop before preflight/serialization, and an admitted branch reaches the
// output-bound preflight when its seeded limit is too small.
func TestProgramBranchValidationOrder(t *testing.T) {
	branchSource, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "borrowed_view.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(branchSource)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("branch fixture must check clean: %+v", checked.Diagnostics)
	}

	t.Run("graph_and_entry_precede_branch_inspection", func(t *testing.T) {
		program := checked.Program
		program.Functions = append([]core.Function(nil), checked.Program.Functions...)
		program.Functions = append(program.Functions, core.Function{ID: "fn:second", Linear: &core.LinearBody{}})
		_, err := cgen.EmitProgramForTest(program)
		if err == nil {
			t.Fatal("expected ambiguous entry refusal")
		}
		if _, ok := callgraph.EntryAmbiguousError(err); !ok {
			t.Fatalf("entry validation did not precede branch inspection: %v", err)
		}
		if cgen.InvocationSerializationReachedForTest() {
			t.Fatal("entry refusal reached C serialization")
		}
	})

	t.Run("foreign_shape_precedes_preflight", func(t *testing.T) {
		program := checked.Program
		program.Functions = append([]core.Function(nil), checked.Program.Functions...)
		program.Functions[0].ForeignContract = &core.ForeignContract{}
		reset := cgen.SetExecutionOutputLimitForTest(execution.MaxDocumentBytes)
		defer reset()
		_, err := cgen.EmitProgramForTest(program)
		if err == nil || !strings.Contains(err.Error(), "multi-function foreign contracts are not supported by native emission this phase") {
			t.Fatalf("want named foreign shape refusal, got %v", err)
		}
		if cgen.InvocationSerializationReachedForTest() {
			t.Fatal("foreign shape refusal reached C serialization")
		}
	})

	t.Run("branch_preflight_seed_is_not_inert", func(t *testing.T) {
		restore := cgen.SetExecutionOutputLimitForTest(1)
		defer restore()
		_, err := cgen.EmitProgramForTest(checked.Program)
		bound, ok := cgen.ExecutionOutputExceededError(err)
		if !ok || bound.Limit() != 1 {
			t.Fatalf("seeded branch output limit did not reach preflight: %v", err)
		}
		if cgen.InvocationSerializationReachedForTest() {
			t.Fatal("branch output preflight refusal reached C serialization")
		}
	})
}

func TestInvocationPathTableBoundary(t *testing.T) {
	accepted := syntheticInvocationProgram(t, 4096)
	if _, err := cgen.EmitNative(accepted); err != nil {
		t.Fatalf("4096-node program refused: %v", err)
	}

	overflow := syntheticInvocationProgram(t, 4097)
	_, err := cgen.EmitNative(overflow)
	if err == nil {
		t.Fatal("4097-node program was admitted")
	}
	cap, ok := cgen.InvocationPathTableExceededError(err)
	if !ok || cap.Code() != "cgen.invocation_path_table_exceeded" {
		t.Fatalf("want cgen.invocation_path_table_exceeded, got %v", err)
	}
	if cap.Entry() == "" || cap.Limit() != 4096 || cap.ObservedAtLeast() != 4097 || cap.FirstOverflowCallOperationID() == "" || cap.FirstOverflowCalleeFunctionID() == "" || cap.FirstOverflowParentIndex() < 0 {
		t.Fatalf("incomplete overflow facts: entry=%q limit=%d observed=%d parent=%d call=%q callee=%q", cap.Entry(), cap.Limit(), cap.ObservedAtLeast(), cap.FirstOverflowParentIndex(), cap.FirstOverflowCallOperationID(), cap.FirstOverflowCalleeFunctionID())
	}
	_, repeated := cgen.EmitNative(overflow)
	if repeated == nil || repeated.Error() != err.Error() {
		t.Fatalf("first-overflow context was not deterministic:\nfirst: %v\nnext:  %v", err, repeated)
	}
}

func TestInvocationPreflightGuardIsNotInert(t *testing.T) {
	overflow := syntheticInvocationProgram(t, 4097)
	restore := cgen.SetInvocationPreflightBypassForTest(false)
	if _, err := cgen.EmitNative(overflow); err == nil {
		restore()
		t.Fatal("restored preflight admitted 4097 nodes")
	}
	if cgen.InvocationSerializationReachedForTest() {
		restore()
		t.Fatal("restored preflight reached forbidden serialization path")
	}
	restore()

	restore = cgen.SetInvocationPreflightBypassForTest(true)
	defer restore()
	if _, err := cgen.EmitNative(overflow); err != nil {
		t.Fatalf("bypassed preflight did not reach emission: %v", err)
	}
	if !cgen.InvocationSerializationReachedForTest() {
		t.Fatal("bypassed preflight did not reach forbidden serialization path")
	}
}

// phase11CheckedProgram checks and independently re-validates a
// testdata/phase11 fixture, returning the validated core.Program (peer
// re-derivation applied, exactly like every other differential path in
// this project).
func phase11CheckedProgram(t *testing.T, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("%s: corevalidate rejected: %+v", fixture, validated.Problems)
	}
	return validated.Program()
}

// TestEmitProgram is Task 1's own emission-shape assertion (D-11-01/D-11-03):
// the emitted C for multi_function_entry_basic.schway contains two function
// definitions, two prototypes preceding them, exactly one `int main(`, and
// compiles under the project's existing native build flags.
func TestEmitProgram(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.schway")

	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}

	if strings.Count(generated, "int main(") != 1 {
		t.Fatalf("expected exactly one int main(, got source:\n%s", generated)
	}
	prototypeIndexMain := strings.Index(generated, "static unsigned char SCHWAY_MAIN(unsigned char, unsigned int);")
	prototypeIndexIdentity := strings.Index(generated, "static unsigned char SCHWAY_IDENTITY(unsigned char, unsigned int);")
	definitionIndexMain := strings.Index(generated, "static unsigned char SCHWAY_MAIN(unsigned char schway_value_")
	definitionIndexIdentity := strings.Index(generated, "static unsigned char SCHWAY_IDENTITY(unsigned char schway_value_")
	if prototypeIndexMain < 0 || prototypeIndexIdentity < 0 || definitionIndexMain < 0 || definitionIndexIdentity < 0 {
		t.Fatalf("expected both prototypes and both definitions present, got source:\n%s", generated)
	}
	if prototypeIndexMain > definitionIndexMain || prototypeIndexMain > definitionIndexIdentity {
		t.Fatalf("expected every prototype before every definition, got source:\n%s", generated)
	}
	if prototypeIndexIdentity > definitionIndexMain || prototypeIndexIdentity > definitionIndexIdentity {
		t.Fatalf("expected every prototype before every definition, got source:\n%s", generated)
	}

	runner := native.DefaultRunner()
	if _, err := runner.Run(context.Background(), generated, "-O0", []string{"7"}); err != nil {
		t.Fatalf("generated C did not compile/link/run: %v", err)
	}
}

func TestPhase18ResultComputedMatchNativeReturn(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "result_computed_match.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	runner := native.DefaultRunner()
	output, err := runner.Run(context.Background(), generated, "-O0", []string{"Raw"})
	if err != nil {
		t.Fatalf("native execution: %v", err)
	}
	if len(output.Pairs) != 1 || output.Pairs[0].Execution.Outcome.Value != "Accepted" {
		t.Fatalf("native result = %+v, want Accepted", output.Pairs)
	}
}

func TestProgramInvocationIndexThreading(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.schway")
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	for _, want := range []string{
		"static unsigned char SCHWAY_MAIN(unsigned char, unsigned int);",
		"static unsigned char SCHWAY_IDENTITY(unsigned char, unsigned int);",
		"SCHWAY_MAIN(schway_entry_input, 0u)",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated C is missing threaded invocation index %q:\n%s", want, generated)
		}
	}
}

func TestPhase17ProgramTwoTypePrototype(t *testing.T) {
	generated := phase17TwoTypeGeneratedC(t)
	phase17RequireTwoTypeC(t, generated)
}

func TestPhase17ProgramTwoTypeDefinition(t *testing.T) {
	generated := phase17TwoTypeGeneratedC(t)
	for _, want := range []string{
		"static SCHWAY_RESULT SCHWAY_CLASSIFY(SCHWAY_RESOURCE value, unsigned int invocation_index)",
		"static SCHWAY_RESULT SCHWAY_MAIN(SCHWAY_RESOURCE schway_value_resource, unsigned int invocation_index)",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated C misses two-type definition %q:\n%s", want, generated)
		}
	}
}

func TestPhase17ProgramTwoTypeCall(t *testing.T) {
	generated := phase17TwoTypeGeneratedC(t)
	if !strings.Contains(generated, "SCHWAY_RESULT schway_value_result = SCHWAY_CLASSIFY(") {
		t.Fatalf("generated C misses Result call target with Resource argument:\n%s", generated)
	}
}

func TestPhase17ProgramTwoTypeEntryIO(t *testing.T) {
	generated := phase17TwoTypeGeneratedC(t)
	for _, want := range []string{
		"SCHWAY_RESOURCE schway_entry_input;",
		"SCHWAY_RESULT schway_entry_output = SCHWAY_MAIN(schway_entry_input, 0u);",
		"SCHWAY_RESULT_name(schway_entry_output)",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated C misses two-type entry position %q:\n%s", want, generated)
		}
	}
}

func TestPhase17ProgramTypePairMutation(t *testing.T) {
	generated := phase17TwoTypeGeneratedC(t)
	phase17RequireTwoTypeC(t, generated)
	for _, mutation := range []struct {
		name, old, new string
	}{
		{"prototype", "static SCHWAY_RESULT SCHWAY_CLASSIFY(SCHWAY_RESOURCE, unsigned int);", "static SCHWAY_RESOURCE SCHWAY_CLASSIFY(SCHWAY_RESOURCE, unsigned int);"},
		{"call target", "SCHWAY_RESULT schway_value_result = SCHWAY_CLASSIFY", "SCHWAY_RESOURCE schway_value_result = SCHWAY_CLASSIFY"},
		{"entry output", "SCHWAY_RESULT schway_entry_output = SCHWAY_MAIN", "SCHWAY_RESOURCE schway_entry_output = SCHWAY_MAIN"},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			mutated := strings.Replace(generated, mutation.old, mutation.new, 1)
			if mutated == generated {
				t.Fatalf("mutation did not find its intended C position: %q", mutation.old)
			}
			if phase17TwoTypeCProblem(mutated) == "" {
				t.Fatalf("seeded %s swap passed structural guard", mutation.name)
			}
		})
	}
}

func phase17TwoTypeGeneratedC(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check tracer: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative tracer: %v", err)
	}
	return generated
}

func phase17RequireTwoTypeC(t *testing.T, generated string) {
	t.Helper()
	if problem := phase17TwoTypeCProblem(generated); problem != "" {
		t.Fatalf("generated C violates Resource -> Result type pair: %s:\n%s", problem, generated)
	}
}

func phase17TwoTypeCProblem(generated string) string {
	for _, want := range []string{
		"static SCHWAY_RESULT SCHWAY_CLASSIFY(SCHWAY_RESOURCE, unsigned int);",
		"static SCHWAY_RESULT SCHWAY_MAIN(SCHWAY_RESOURCE, unsigned int);",
		"SCHWAY_RESULT schway_value_result = SCHWAY_CLASSIFY(",
		"SCHWAY_RESOURCE schway_entry_input;",
		"SCHWAY_RESULT schway_entry_output = SCHWAY_MAIN(schway_entry_input, 0u);",
	} {
		if !strings.Contains(generated, want) {
			return "missing " + want
		}
	}
	return ""
}

func TestParentIndexedChildLookup(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_diamond_call.schway")
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	if !strings.Contains(generated, "static const char *schway_invocations[]") {
		t.Fatalf("generated C has no literal invocation table:\n%s", generated)
	}
	if !strings.Contains(generated, ":fn:main:op:0#0/s1:phase11.multi_function_diamond_call:fn:left:op:0#0") ||
		!strings.Contains(generated, ":fn:main:op:1#0/s1:phase11.multi_function_diamond_call:fn:right:op:0#0") {
		t.Fatalf("generated C does not retain distinct shared-leaf occurrences:\n%s", generated)
	}
	if !strings.Contains(generated, "schway_child_index_") {
		t.Fatalf("generated C has no parent-indexed child lookup:\n%s", generated)
	}
}

func TestInvocationTableEmissionIsDeterministic(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_diamond_call.schway")
	first, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("first EmitNative: %v", err)
	}
	second, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("second EmitNative: %v", err)
	}
	if first != second {
		t.Fatal("invocation table emission is not byte-identical")
	}
}

func TestProgramWritesExecutionSchema2(t *testing.T) {
	generated, err := cgen.EmitNative(phase11CheckedProgram(t, "multi_function_entry_basic.schway"))
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	for _, want := range []string{"{\\\"schema\\\":\\\"lang.execution/2\\\"", ",\\\"invocation\\\":", ",\\\"callee_function_id\\\":"} {
		if !strings.Contains(generated, want) {
			t.Fatalf("schema-2 generated C is missing %q:\n%s", want, generated)
		}
	}
}

func TestNativeFunctionCalledPreorder(t *testing.T) {
	source := mustRead(t, "multi_function_diamond_call.schway")
	interpreted, diags, err := session.RunInterpreter(source)
	if err != nil || len(diags) != 0 || len(interpreted) != 1 {
		t.Fatalf("RunInterpreter: executions=%d diagnostics=%v error=%v", len(interpreted), diags, err)
	}
	nativeResult, diags, err := session.RunNative(context.Background(), source, native.DefaultRunner())
	if err != nil || len(diags) != 0 || len(nativeResult.O0.Pairs) != 1 {
		t.Fatalf("RunNative: pairs=%d diagnostics=%v error=%v", len(nativeResult.O0.Pairs), diags, err)
	}
	got := nativeResult.O0.Pairs[0].Execution
	if !execution.Equal(interpreted[0], got) {
		t.Fatalf("native /2 evidence disagrees with interpreter:\nwant=%+v\ngot=%+v", interpreted[0], got)
	}
	called := 0
	for index, event := range got.Events {
		if event.Kind != "function.called" {
			continue
		}
		called++
		if event.Invocation == "" || event.CalleeFunctionID == "" {
			t.Fatalf("call edge %d omits /2 ownership: %+v", index, event)
		}
		if index+1 >= len(got.Events) || got.Events[index+1].FunctionID != event.CalleeFunctionID {
			t.Fatalf("call edge %d is not immediately before callee evidence: %+v", index, got.Events)
		}
	}
	if called != 4 {
		t.Fatalf("function.called edges = %d, want 4", called)
	}
}

func TestNativeFunctionCalledProjectionRemoval(t *testing.T) {
	withRight := mustRead(t, "multi_function_diamond_call.schway")
	withoutRight := []byte(strings.Replace(string(withRight), "fn left(value: Byte) -> Byte {\n  let result = leaf(value)\n  result\n}", "fn left(value: Byte) -> Byte {\n  value\n}", 1))
	run := func(source []byte) []execution.Event {
		t.Helper()
		result, diags, err := session.RunNative(context.Background(), source, native.DefaultRunner())
		if err != nil || len(diags) != 0 || len(result.O0.Pairs) != 1 {
			t.Fatalf("RunNative: pairs=%d diagnostics=%v error=%v", len(result.O0.Pairs), diags, err)
		}
		var edges []execution.Event
		for _, event := range result.O0.Pairs[0].Execution.Events {
			if event.Kind == "function.called" {
				edges = append(edges, event)
			}
		}
		return edges
	}
	if got, want := len(run(withRight))-len(run(withoutRight)), 1; got != want {
		t.Fatalf("removing one source call changed call-edge projection by %d, want %d", got, want)
	}
}

func TestLegacyEventWritersFrozen(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "toggle.schway"))
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "generated.golden.c"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Check: %+v", checked.Diagnostics)
	}
	generated, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if generated != string(golden) {
		t.Fatal("legacy /0 event writer bytes changed")
	}
}

// TestEmitProgramEndToEndAgreesWithInterpreterAtO0 drives
// multi_function_entry_basic.schway through session's own run path on both
// engines and asserts the two lang.execution/1 documents are equal --
// end to end through session, not only through a unit test calling cgen
// directly.
func TestEmitProgramEndToEndAgreesWithInterpreterAtO0(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", "multi_function_entry_basic.schway"))
	if err != nil {
		t.Fatal(err)
	}

	interpreted, diags, err := session.RunInterpreter(source)
	if err != nil {
		t.Fatalf("RunInterpreter: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("RunInterpreter diagnostics: %v", diags)
	}
	if len(interpreted) != 1 {
		t.Fatalf("expected exactly one interpreter execution, got %d", len(interpreted))
	}

	result, diags, err := session.RunNative(context.Background(), source, native.DefaultRunner())
	if err != nil {
		t.Fatalf("RunNative: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("RunNative diagnostics: %v", diags)
	}
	if len(result.O0.Pairs) != 1 {
		t.Fatalf("expected exactly one -O0 pair, got %d", len(result.O0.Pairs))
	}

	if !execution.Equal(interpreted[0], result.O0.Pairs[0].Execution) {
		t.Fatalf("interpreter and native -O0 disagree:\ninterp=%+v\nnative=%+v", interpreted[0], result.O0.Pairs[0].Execution)
	}
}

// TestEmitProgramZeroCallEdges pins the empty-call-edge structural case
// (PLAN.md Task 3): multi_function_zero_call.schway declares two functions
// with ZERO core.OpCall operations anywhere in the program. Both are
// therefore in-degree-zero candidates with an EMPTY reachable closure each
// -- a genuine, unbreakable tie under callgraph.EntryFunction's own
// documented closure-size tie-break (never a guess, per D-11-05's
// "MUST NOT guess an entry function when zero or many in-degree-zero roots
// exist" prohibition). The non-degenerate behavior this fixture's own
// header comment and 11-03-SUMMARY.md's Deviations section document is a
// clean, named core.EntryAmbiguous refusal -- never a crash, never a
// silent Functions[0]-style guess, and never a wrong-answer emission. This
// deliberately diverges from this task's own literal "assert emission
// succeeds" wording (see the Deviations section for the full reasoning):
// that wording is unsatisfiable for a truly zero-call, two-function
// program under a non-guessing resolver, and D-11-05's fail-closed
// invariant is prioritized.
func TestEmitProgramZeroCallEdges(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_zero_call.schway")

	_, err := cgen.EmitNative(program)
	if err == nil {
		t.Fatal("expected emitProgram to refuse an ambiguous-entry program, got nil error")
	}
	ambiguous, ok := callgraph.EntryAmbiguousError(err)
	if !ok {
		t.Fatalf("expected a callgraph entry-ambiguity refusal, got: %v", err)
	}
	if ambiguous.Code() != core.EntryAmbiguous {
		t.Fatalf("expected code %q, got %q", core.EntryAmbiguous, ambiguous.Code())
	}
	roots := ambiguous.Roots()
	if len(roots) != 2 {
		t.Fatalf("expected exactly two candidate roots, got %v", roots)
	}

	// The SAME refusal, from the SAME single resolver, must also surface
	// through session's own run path (D-11-05's "one resolver" property):
	// the interpreter and the native binary must never be given a chance
	// to silently disagree about which function this program even is.
	if _, _, err := session.RunInterpreter(mustRead(t, "multi_function_zero_call.schway")); err == nil {
		t.Fatal("expected RunInterpreter to refuse the ambiguous-entry program")
	}
}

// TestEmitProgramForwardDefinedCallee proves D-11-03's ordering claim: the
// entry calls a function declared LATER in the source
// (multi_function_forward_callee.schway), and emitProgram writes every
// prototype before any definition, so the forward reference is legal C17.
func TestEmitProgramForwardDefinedCallee(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_forward_callee.schway")

	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	if !strings.Contains(generated, "static unsigned char SCHWAY_LATER(unsigned char, unsigned int);") {
		t.Fatalf("expected SCHWAY_LATER's own forward prototype, got source:\n%s", generated)
	}

	runner := native.DefaultRunner()
	if _, err := runner.Run(context.Background(), generated, "-O0", []string{"7"}); err != nil {
		t.Fatalf("generated C did not compile/link/run: %v", err)
	}
}

// TestEmitProgramRefusesRecursiveProgramsUpstream builds a self-recursive
// and a mutually-recursive core.Program directly (mirroring
// callgraph_test.go's own synthetic-program convention) and asserts each
// is refused by check.Program with core.CallGraphCycle BEFORE emitProgram
// is ever reachable -- the refusal is asserted by CODE, not merely by
// absence of emitted output.
func TestEmitProgramRefusesRecursiveProgramsUpstream(t *testing.T) {
	// plan 14-07: this fixture is checked into the repo
	// (testdata/phase07/cycle_self.schway); a read failure here is repo
	// corruption, not a legitimate skip condition, so it fails loudly
	// rather than silently skipping.
	selfRecursive, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "cycle_self.schway"))
	if err != nil {
		t.Fatalf("testdata/phase07/cycle_self.schway: %v", err)
	}
	checked := session.Check(selfRecursive)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("expected self-recursive program to be refused by check")
	}
	if checked.Diagnostics[0].Code != core.CallGraphCycle {
		t.Fatalf("expected %q, got %q", core.CallGraphCycle, checked.Diagnostics[0].Code)
	}

	mutual, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "cycle_unreachable.schway"))
	if err != nil {
		t.Fatal(err)
	}
	mutualChecked := session.Check(mutual)
	if len(mutualChecked.Diagnostics) == 0 {
		t.Fatal("expected mutually-recursive program to be refused by check")
	}
	if mutualChecked.Diagnostics[0].Code != core.CallGraphCycle {
		t.Fatalf("expected %q, got %q", core.CallGraphCycle, mutualChecked.Diagnostics[0].Code)
	}
}

// TestEmitProgramSingleFunctionBytesUnchanged is D-11-01's own freeze
// guard: for a representative testdata/phase1 fixture (the same
// source/golden pairing TestExistingEmittersAreByteIdentical already
// pins), Emit still routes to the single-function path -- the additive
// Emit/EmitNative dispatch fork for N>1 never fires for N==1 -- and its
// output is byte-identical to the committed golden, comparing against the
// golden itself rather than a freshly generated string.
func TestEmitProgramSingleFunctionBytesUnchanged(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "toggle.schway"))
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "generated.golden.c"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 {
		t.Fatalf("expected this golden's own fixture to stay single-function, got %d", len(checked.Program.Functions))
	}
	generated, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatalf("Emit: %v", err)
	}
	if generated != string(golden) {
		t.Fatalf("single-function output moved:\nwant:\n%s\ngot:\n%s", golden, generated)
	}
}

func mustRead(t *testing.T, fixture string) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", fixture))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

// attributeSetCommentLine returns the generated call-boundary
// attribute-set comment line from generated, failing the test if absent.
func attributeSetCommentLine(t *testing.T, generated string) string {
	t.Helper()
	for _, line := range strings.Split(generated, "\n") {
		if strings.Contains(line, "call-boundary attribute set") {
			return line
		}
	}
	t.Fatalf("no call-boundary attribute set comment found in:\n%s", generated)
	return ""
}

// TestEmittedAttributeSetIsExplicitlyEmpty is NAT-05's own satisfaction
// test (D-11-09/D-11-10): the emitted C for the tracer fixture names its
// call-boundary attribute set as explicitly empty, cites D-11-09, and
// cgen.ScanForBannedAttributes agrees the emitted artifact carries no
// banned token at all.
func TestEmittedAttributeSetIsExplicitlyEmpty(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.schway")
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}

	comment := attributeSetCommentLine(t, generated)
	if !strings.Contains(comment, "EMPTY") {
		t.Fatalf("expected the comment to name the set as EMPTY, got: %q", comment)
	}
	if !strings.Contains(comment, "D-11-09") {
		t.Fatalf("expected the comment to cite D-11-09, got: %q", comment)
	}
	if found := cgen.ScanForBannedAttributes(generated); len(found) != 0 {
		t.Fatalf("expected no banned attributes in the emitted artifact, found: %v", found)
	}
}

// TestEmittedAttributeSetCommentIsDerivedNotLiteral proves the rendered
// comment tracks emitCallBoundaryAttributeSet's own derivation rather than
// restating a hand-written literal beside it (D-04-12): injecting a
// non-empty set through the derivation's own test seam changes the
// rendered comment's content.
func TestEmittedAttributeSetCommentIsDerivedNotLiteral(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.schway")

	baseline, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative (baseline): %v", err)
	}
	baselineComment := attributeSetCommentLine(t, baseline)

	restore := cgen.SetCallBoundaryAttributeSetForTest([]string{"probe:injected-attribute"})
	defer restore()
	mutated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative (mutated derivation): %v", err)
	}
	mutatedComment := attributeSetCommentLine(t, mutated)

	if baselineComment == mutatedComment {
		t.Fatalf("expected the rendered comment to change when the derivation returns a non-empty set, got identical comment: %q", baselineComment)
	}
	if !strings.Contains(mutatedComment, "probe:injected-attribute") {
		t.Fatalf("expected the injected derivation's entry to appear in the rendered comment, got: %q", mutatedComment)
	}
}

// TestEmittedAttributeSetOrderIsStable proves the attribute-set statement
// is byte-identical across a same-program re-emission and a declaration-
// order-permuted variant of the same program (must_have: byte-stable
// order regardless of function declaration order).
func TestEmittedAttributeSetOrderIsStable(t *testing.T) {
	// The former gate corpus intentionally contains a pointer-specialized
	// helper and is now a cut-m004 refusal. This ordinary multi-function
	// caller still exercises declaration-order stability without weakening
	// the whole-program admission boundary.
	program := phase11CheckedProgram(t, "multi_function_entry_basic.schway")

	first, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative (first): %v", err)
	}
	second, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative (second): %v", err)
	}
	if first != second {
		t.Fatal("expected re-emitting the same program twice to be byte-identical")
	}

	permuted := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID}
	for index := len(program.Functions) - 1; index >= 0; index-- {
		permuted.Functions = append(permuted.Functions, program.Functions[index])
	}
	permutedGenerated, err := cgen.EmitNative(permuted)
	if err != nil {
		t.Fatalf("EmitNative (permuted): %v", err)
	}

	firstComment := attributeSetCommentLine(t, first)
	permutedComment := attributeSetCommentLine(t, permutedGenerated)
	if firstComment != permutedComment {
		t.Fatalf("expected the same attribute-set statement regardless of declaration order:\nfirst=%q\npermuted=%q", firstComment, permutedComment)
	}
}

// TestCallSitesEmitNoArithmeticConversion is NAT-05's precision backstop:
// Lang parameters are Byte/Buffer by value with no arithmetic in the
// language, so no call site in the gate corpus's emitted C may introduce
// an arithmetic conversion.
func TestCallSitesEmitNoArithmeticConversion(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.schway")
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}

	arithmeticTokens := []string{"+", "-", "*", "/", "%", "<<", ">>", "(int)", "(unsigned", "(signed", "(double)", "(float)"}
	for _, line := range strings.Split(generated, "\n") {
		commentIndex := strings.Index(line, "/* call:")
		if commentIndex < 0 {
			continue
		}
		code := line[:commentIndex]
		for _, token := range arithmeticTokens {
			if strings.Contains(code, token) {
				t.Fatalf("call site emits an arithmetic-conversion token %q: %q", token, line)
			}
		}
	}
}

func TestPhase22ApplicationEmitterSharesBodyAndSeparatesOutputShell(t *testing.T) {
	source := []byte("module phase22.identity\n\nexport { fn main }\n\nfn main(input: U64) -> U64 { input }\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate problems: %+v", validated.Problems)
	}

	application, err := cgen.EmitApplication(validated.Program())
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	conformance, err := cgen.EmitNative(validated.Program())
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	appMain := strings.LastIndex(application, "\nint main(")
	conformanceMain := strings.LastIndex(conformance, "\nint main(")
	if appMain < 0 || conformanceMain < 0 {
		t.Fatal("generated output is missing main")
	}
	appBodyStart := strings.LastIndex(application[:appMain], "\nstatic ")
	conformanceBodyStart := strings.LastIndex(conformance[:conformanceMain], "\nstatic ")
	if appBodyStart < 0 || conformanceBodyStart < 0 || application[appBodyStart:appMain] != conformance[conformanceBodyStart:conformanceMain] {
		t.Fatal("application entry changed the checked function-body lowering")
	}
	appEntry := application[appMain:]
	parse := strings.Index(appEntry, "schway_parse_u64_decimal(argv[1]")
	call := strings.Index(appEntry, "schway_entry_output = ")
	if parse < 0 || call < 0 || parse >= call {
		t.Fatalf("generated application does not validate input before the Lang body:\n%s", appEntry)
	}
	if !strings.Contains(appEntry, "strlen(argv[1]) > 4096u") || !strings.Contains(appEntry, "schway_write_u64_plain(schway_entry_output)") || !strings.Contains(appEntry, "schway_write_literal(\"\\n\")") {
		t.Fatalf("application shell is missing its transport bound or plain decimal output:\n%s", appEntry)
	}
	evidenceBranch := strings.Index(appEntry, "const char *schway_evidence_path = getenv(\"SCHWAY_APP_EVIDENCE_PATH\")")
	if evidenceBranch < 0 || strings.Contains(appEntry[:evidenceBranch], "lang.execution/2") || strings.Contains(appEntry[:evidenceBranch], "schway_write_events()") {
		t.Fatalf("application shell serializes compiler evidence before the private capture branch:\n%s", appEntry)
	}
	if !strings.Contains(appEntry[evidenceBranch:], "schway_output_stream = schway_evidence_file") || !strings.Contains(appEntry[evidenceBranch:], "schway_write_events()") {
		t.Fatalf("application shell does not direct captured events to the private file:\n%s", appEntry)
	}
	if !strings.Contains(application, "static FILE *schway_output_stream = NULL;") || !strings.Contains(application, "schway_output_stream != NULL ? schway_output_stream : stdout") {
		t.Fatal("application evidence writer must use a portable runtime stdout fallback")
	}
	if !strings.Contains(conformance[conformanceMain:], "lang.execution/2") || !strings.Contains(conformance[conformanceMain:], "schway_write_events()") {
		t.Fatal("conformance emitter lost its execution-document shell")
	}
}

func TestPhase22ApplicationEmitterRefusesOtherEntryShapes(t *testing.T) {
	for _, source := range []string{
		"module phase22.byte_result\n\nexport { fn main }\n\nfn main(input: Byte) -> Byte {\n  input\n}\n",
		"module phase22.byte_input\n\nexport { fn main }\n\nfn main(input: Byte) -> U64 {\n  let count = 42\n  count\n}\n",
	} {
		checked := session.Check([]byte(source))
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("unexpected checker diagnostics: %+v", checked.Diagnostics)
		}
		if _, err := cgen.EmitApplication(checked.Program); err == nil {
			t.Fatalf("EmitApplication accepted unsupported entry shape for %q", source)
		}
	}
}

func TestPhase23PhysicalDestructorControlsKeepEventsPlausible(t *testing.T) {
	generated, err := cgen.EmitApplication(phase23ProgramForEmitter(t))
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	if !strings.Contains(generated, "lang.execution/2") || !strings.Contains(generated, `schway_record_event("function.returned"`) {
		t.Fatal("compiler-side event capture lost its ordinary schema-2 returned execution record")
	}
	if !strings.Contains(generated, "schway_file_byte_release(") {
		t.Fatal("generated app no longer carries the physical destructor call targeted by native mutation controls")
	}
	if strings.Contains(generated, `schway_record_event("resource.released"`) || strings.Contains(generated, `schway_record_event("allocation.freed"`) {
		t.Fatal("compiler events claim physical destructor behavior instead of remaining semantic records")
	}
}

func phase23ProgramForEmitter(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase23", "file_byte.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("file_byte.schway checker diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("file_byte.schway core validation problems: %+v", validated.Problems)
	}
	return validated.Program()
}

func mutatePhase23Entry(program *core.Program, mutate func(*core.Function)) {
	program.Functions = append([]core.Function(nil), program.Functions...)
	for index := range program.Functions {
		if program.Functions[index].Name != "main" || program.Functions[index].Linear == nil {
			continue
		}
		function := program.Functions[index]
		linear := *function.Linear
		linear.Operations = append([]core.LinearOperation(nil), function.Linear.Operations...)
		function.Linear = &linear
		mutate(&function)
		program.Functions[index] = function
		return
	}
	panic("Phase 23 emitter fixture has no linear main")
}

func mutatePhase23Contract(program *core.Program, operationIndex int, mutate func(*core.ForeignOperationContract)) {
	mutatePhase23Entry(program, func(function *core.Function) {
		if operationIndex < 0 || operationIndex >= len(function.Linear.Operations) {
			panic("Phase 23 emitter fixture operation index out of range")
		}
		operation := function.Linear.Operations[operationIndex]
		if operation.Foreign == nil {
			panic("Phase 23 emitter fixture operation has no checked foreign contract")
		}
		contract := *operation.Foreign
		mutate(&contract)
		operation.Foreign = &contract
		function.Linear.Operations[operationIndex] = operation
	})
}

func requirePhase23EmitterRefusalBeforeSerialization(t *testing.T, program core.Program, name string) {
	t.Helper()
	generated, err := cgen.EmitProgramForTest(program)
	if err == nil {
		t.Fatalf("%s candidate reached native emission; generated C length=%d", name, len(generated))
	}
	if generated != "" {
		t.Fatalf("%s refusal returned partial C (%d bytes): %v", name, len(generated), err)
	}
	if cgen.InvocationSerializationReachedForTest() {
		t.Fatalf("%s candidate was refused only after C serialization began: %v", name, err)
	}
}

func TestPhase23DiscardRefusedBeforeCSerialization(t *testing.T) {
	program := phase23ProgramForEmitter(t)
	mutatePhase23Entry(&program, func(function *core.Function) {
		operations := function.Linear.Operations
		// The acquire remains, but no consumer or release is allowed to erase
		// its successful owner result from the candidate program.
		function.Linear.Operations = []core.LinearOperation{operations[0], operations[3]}
	})
	requirePhase23EmitterRefusalBeforeSerialization(t, program, "discarded acquisition")
}

func TestPhase23SourceRefusalBeforeCSerialization(t *testing.T) {
	program := phase23ProgramForEmitter(t)
	generated, err := cgen.EmitProgramForTest(program)
	if err != nil || generated == "" || !cgen.InvocationSerializationReachedForTest() {
		t.Fatalf("valid local acquire/borrow/release path was not emitted: bytes=%d err=%v", len(generated), err)
	}

	for _, tc := range []struct {
		name   string
		mutate func(*core.Function)
	}{
		{
			name: "owner copy",
			mutate: func(function *core.Function) {
				operations := function.Linear.Operations
				copyOwner := core.LinearOperation{ID: "candidate:copy", Kind: core.OpCopy, SourceID: operations[0].TargetID, TargetID: "candidate:copy-result", TypeID: operations[0].TypeID}
				function.Linear.Operations = []core.LinearOperation{operations[0], copyOwner, operations[1], operations[2], operations[3]}
			},
		},
		{
			name: "moved-from use",
			mutate: func(function *core.Function) {
				operations := function.Linear.Operations
				moveOwner := core.LinearOperation{ID: "candidate:move", Kind: core.OpMove, SourceID: operations[0].TargetID, TargetID: "candidate:moved-owner", TypeID: operations[0].TypeID}
				function.Linear.Operations = []core.LinearOperation{operations[0], moveOwner, operations[1], operations[2], operations[3]}
			},
		},
		{
			name: "owner escape",
			mutate: func(function *core.Function) {
				operations := function.Linear.Operations
				operations[3].SourceID = operations[0].TargetID
				function.Linear.Operations = operations
			},
		},
		{
			name: "cross-call transfer",
			mutate: func(function *core.Function) {
				operations := function.Linear.Operations
				transfer := core.LinearOperation{ID: "candidate:transfer", Kind: core.OpCall, SourceID: operations[0].TargetID, TargetID: "candidate:transfer-result", TypeID: operations[0].TypeID}
				function.Linear.Operations = []core.LinearOperation{operations[0], transfer, operations[1], operations[2], operations[3]}
			},
		},
		{
			name: "unsupported exit",
			mutate: func(function *core.Function) {
				operations := function.Linear.Operations
				operations[3].Kind = core.OpFail
				function.Linear.Operations = operations
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := phase23ProgramForEmitter(t)
			mutatePhase23Entry(&candidate, tc.mutate)
			requirePhase23EmitterRefusalBeforeSerialization(t, candidate, tc.name)
		})
	}
}

func TestPhase23OperationContractRefusalBeforeCSerialization(t *testing.T) {
	valid := phase23ProgramForEmitter(t)
	if _, err := cgen.EmitProgramForTest(valid); err != nil {
		t.Fatalf("valid three-operation owner contract refused: %v", err)
	}
	if !cgen.InvocationSerializationReachedForTest() {
		t.Fatal("valid local owner contract did not reach C serialization")
	}

	for _, tc := range []struct {
		name           string
		operationIndex int
		mutate         func(*core.ForeignOperationContract)
	}{
		{"acquire status type", 0, func(contract *core.ForeignOperationContract) { contract.Fails = "WrongAcquireError" }},
		{"acquire symbol", 0, func(contract *core.ForeignOperationContract) { contract.Symbol = "schway_other_acquire" }},
		{"acquire mode", 0, func(contract *core.ForeignOperationContract) { contract.Mode = "borrow" }},
		{"acquire ABI type", 0, func(contract *core.ForeignOperationContract) { contract.ABIType = "wrong_acquire_fn" }},
		{"acquire operand type", 0, func(contract *core.ForeignOperationContract) { contract.ParameterType = "FileByteOwner" }},
		{"acquire result type", 0, func(contract *core.ForeignOperationContract) { contract.ResultType = "U64" }},
		{"acquire allocator", 0, func(contract *core.ForeignOperationContract) { contract.Allocator = "another_allocator" }},
		{"use status type", 1, func(contract *core.ForeignOperationContract) { contract.Fails = "WrongUseError" }},
		{"use symbol", 1, func(contract *core.ForeignOperationContract) { contract.Symbol = "schway_other_use" }},
		{"use mode", 1, func(contract *core.ForeignOperationContract) { contract.Mode = "consume" }},
		{"use ABI type", 1, func(contract *core.ForeignOperationContract) { contract.ABIType = "wrong_use_fn" }},
		{"use operand type", 1, func(contract *core.ForeignOperationContract) { contract.ParameterType = "PathToken" }},
		{"use result type", 1, func(contract *core.ForeignOperationContract) { contract.ResultType = "FileByteOwner" }},
		{"release symbol", 2, func(contract *core.ForeignOperationContract) { contract.Symbol = "schway_other_release" }},
		{"release mode", 2, func(contract *core.ForeignOperationContract) { contract.Mode = "borrow" }},
		{"release failure type", 2, func(contract *core.ForeignOperationContract) { contract.Fails = "ReleaseError" }},
		{"release ABI type", 2, func(contract *core.ForeignOperationContract) { contract.ABIType = "wrong_release_fn" }},
		{"release operand type", 2, func(contract *core.ForeignOperationContract) { contract.ParameterType = "PathToken" }},
		{"release result type", 2, func(contract *core.ForeignOperationContract) { contract.ResultType = "U64" }},
		{"allocator mismatch", 2, func(contract *core.ForeignOperationContract) { contract.Allocator = "another_allocator" }},
		{"destructor pairing", 0, func(contract *core.ForeignOperationContract) { contract.Release = "schway_other_release" }},
		{"unwind policy", 1, func(contract *core.ForeignOperationContract) { contract.Unwind = "allowed" }},
		{"nonlocal exit policy", 1, func(contract *core.ForeignOperationContract) { contract.NonlocalExit = "allowed" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			candidate := phase23ProgramForEmitter(t)
			mutatePhase23Contract(&candidate, tc.operationIndex, tc.mutate)
			requirePhase23EmitterRefusalBeforeSerialization(t, candidate, tc.name)
		})
	}
}

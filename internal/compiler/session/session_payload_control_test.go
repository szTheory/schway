package session_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPayloadLayoutMutationRefused proves control:payload.layout_mismatch
// (D-12-37): compiling the generated payload conformance unit
// (cgen.EmitPayloadConformance) against the transposed frozen fixture
// testdata/phase12/payload_layout_mismatch.golden.c is refused at compile
// time under the project's existing -Werror flag set, reporting the
// distinct "native.conformance_failed" code -- exactly mirroring
// TestLayoutMutationIsCompileTimeRefusal's own shape for the foreign-contract
// layout control.
//
// This control is NECESSARY for the C-side layout obligation and
// STRUCTURALLY INCAPABLE of catching a wrong-slot READ for a correct tag --
// see PayloadLayoutMutationRunner's own doc comment. D-12-38's
// TestPayloadSlotSwapMutationKilled (below) is the control that covers that
// gap.
func TestPayloadLayoutMutationRefused(t *testing.T) {
	runner := session.PayloadLayoutMutationRunner{
		Runner:      native.DefaultRunner(),
		DataType:    session.PayloadProbeDataType(),
		FixturePath: testsupport.ProjectPath("testdata", "phase12", "payload_layout_mismatch.golden.c"),
	}
	err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected the transposed payload fixture to be refused at compile time")
	}
	var toolErr *native.ToolError
	if !errors.As(err, &toolErr) || toolErr.Code != "native.conformance_failed" {
		t.Fatalf("expected native.conformance_failed, got %v", err)
	}
}

// TestPayloadLayoutMutationAttacksFrozenFixtureOnly companions the refusal
// test above: pointing FixturePath at a correctly-ordered (untransposed)
// private header -- written to a throwaway temp file, never committed to
// testdata, mirroring TestLayoutMutationAttacksFrozenFixtureOnly's own
// shape -- compiles cleanly, proving the runner's verdict depends solely on
// FixturePath's own content rather than always refusing regardless of
// input. The runner type itself carries no field of a generated-source
// shape (Runner, DataType, FixturePath only).
func TestPayloadLayoutMutationAttacksFrozenFixtureOnly(t *testing.T) {
	correctPath := filepath.Join(t.TempDir(), "schway_payload_layout_probe_correct.h")
	correctHeader := []byte(`#ifndef SCHWAY_PAYLOAD_LAYOUT_PROBE_PRIVATE_H
#define SCHWAY_PAYLOAD_LAYOUT_PROBE_PRIVATE_H
typedef struct PayloadProbe_payload {
  unsigned char tag;
  unsigned char field_First;
  unsigned char field_Second;
} PayloadProbe_payload;
#endif
`)
	if err := os.WriteFile(correctPath, correctHeader, 0o600); err != nil {
		t.Fatal(err)
	}
	runner := session.PayloadLayoutMutationRunner{Runner: native.DefaultRunner(), DataType: session.PayloadProbeDataType(), FixturePath: correctPath}
	if err := runner.Run(context.Background()); err != nil {
		t.Fatalf("expected the untransposed fixture to conform, got %v", err)
	}
	value := reflect.ValueOf(runner)
	if value.NumField() != 3 {
		t.Fatalf("PayloadLayoutMutationRunner grew an unexpected field: %+v", runner)
	}
}

// isolateNativeFunction extracts ONE function (by name) plus every
// core.DataType it transitively references (by parameter/return type name,
// and by any AlternativeDetail's PayloadType naming a sibling data type)
// out of a checked, multi-function core.Program, and re-validates the
// result. Native emission does not support multi-function branch bodies
// this phase (D-11-52/D-12-32); payload_borrow_interaction.schway's own
// design deliberately combines payload and borrow machinery as TWO
// functions in ONE checked program specifically because they cannot
// combine in one function yet (12-03-SUMMARY.md), so driving either
// function alone through native emission requires this isolation step --
// it changes no operation, place, or type fact, only which subset of an
// already-checked program's declarations are carried into the emitted
// translation unit.
func isolateNativeFunction(t *testing.T, program core.Program, functionName string) core.Program {
	t.Helper()
	var target core.Function
	found := false
	for _, function := range program.Functions {
		if function.Name == functionName {
			target = function
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("isolateNativeFunction: function %q not found", functionName)
	}
	byName := make(map[string]core.DataType, len(program.DataTypes))
	for _, dataType := range program.DataTypes {
		byName[dataType.Name] = dataType
	}
	needed := make(map[string]bool, len(program.DataTypes))
	var visit func(name string)
	visit = func(name string) {
		if needed[name] {
			return
		}
		dataType, ok := byName[name]
		if !ok {
			return
		}
		needed[name] = true
		for _, detail := range dataType.AlternativeDetails {
			if detail.PayloadType != "" {
				visit(detail.PayloadType)
			}
		}
	}
	visit(target.Parameter.Type)
	visit(target.ReturnType)
	dataTypes := make([]core.DataType, 0, len(needed))
	for _, dataType := range program.DataTypes {
		if needed[dataType.Name] {
			dataTypes = append(dataTypes, dataType)
		}
	}
	isolated := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, DataTypes: dataTypes, Functions: []core.Function{target}}
	validated := corevalidate.Validate(isolated)
	if !validated.Valid {
		t.Fatalf("isolateNativeFunction: %q: core validation failed after isolation: %+v", functionName, validated.Problems)
	}
	return validated.Program()
}

// runFunctionOkExecutions drives functionName's "Ok" input (both fixtures
// declare an Outcome{Ok(Buffer), Err(Fault)} identity-shaped function with
// this exact input) through the interpreter and both native optimization
// levels, returning the three comparable execution.Execution documents
// keyed by engine name -- the same shape Phase5CompareEngines-based
// controls (session_phase5_mismatch.go's ReduceSeededAliasMismatch) drive
// directly rather than through RunNativeFile's own internal execution.Equal
// short-circuit, so this test can name which AXIS a disagreement (if any)
// falls on.
func runFunctionOkExecutions(t *testing.T, ctx context.Context, program core.Program, functionName string, runner native.Runner) map[string]execution.Execution {
	t.Helper()
	cSource, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("%s: EmitNative: %v", functionName, err)
	}
	interpreted, err := interp.Run(program, functionName, "Ok")
	if err != nil {
		t.Fatalf("%s: interp.Run(Ok): %v", functionName, err)
	}
	o0, err := runner.Run(ctx, cSource, "-O0", []string{"Ok"})
	if err != nil {
		t.Fatalf("%s: -O0 run: %v", functionName, err)
	}
	o3, err := runner.Run(ctx, cSource, "-O3", []string{"Ok"})
	if err != nil {
		t.Fatalf("%s: -O3 run: %v", functionName, err)
	}
	if len(o0.Pairs) != 1 || len(o3.Pairs) != 1 {
		t.Fatalf("%s: expected exactly one execution per optimization level, got -O0=%d -O3=%d", functionName, len(o0.Pairs), len(o3.Pairs))
	}
	interpreted, err = session.ProjectExecutionSchema2(program, interpreted)
	if err != nil {
		t.Fatalf("%s: interpreter schema-2 projection: %v", functionName, err)
	}
	o0Execution, err := session.ProjectExecutionSchema2(program, o0.Pairs[0].Execution)
	if err != nil {
		t.Fatalf("%s: -O0 schema-2 projection: %v", functionName, err)
	}
	o3Execution, err := session.ProjectExecutionSchema2(program, o3.Pairs[0].Execution)
	if err != nil {
		t.Fatalf("%s: -O3 schema-2 projection: %v", functionName, err)
	}
	return map[string]execution.Execution{
		"interpreter": interpreted,
		"O0":          o0Execution,
		"O3":          o3Execution,
	}
}

// checkedProgram runs relativeParts through session.Check and fatals on any
// diagnostic, returning the checked core.Program.
func checkedProgram(t *testing.T, relativeParts ...string) core.Program {
	t.Helper()
	path := testsupport.ProjectPath(relativeParts...)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", filepath.Join(relativeParts...), err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", filepath.Join(relativeParts...), checked.Diagnostics)
	}
	return checked.Program
}

// TestPayloadSlotSwapMutationKilled is D-12-38's decisive control: a
// reverted production-hunk mutation seeding a bug in cgen's
// OpConstructPayload codegen that writes the source payload into a
// DIFFERENT alternative's struct field than the one the (correct) tag
// names. Two beats, per D-12-38:
//
//  1. Mutated: with cgen.SetPayloadSlotSwapForTest(true) engaged, drive
//     plan 02's tracer fixture (payload_tracer.schway, "Ok" input) through
//     interpreter/-O0/-O3 and compare via session.Phase5CompareEngines,
//     asserting on session.AxisTerminalOutcome specifically.
//  2. Companion (unmutated): drive plan 03's payload_borrow_interaction.schway
//     and assert the comparator reports AGREEMENT -- discriminating real
//     value-identity checking from a harness that screams on any diff.
//
// WR-02: the mutated beat asserts, via
// cgen.PayloadSlotSwapInjectedWriteCount, that engaging the seam actually
// injected a wrong-slot write at least once. This proves the MUTATION WAS
// APPLIED. The schema-2 terminal outcome now includes the runtime payload,
// so this control must observe the wrong-slot corruption on the terminal
// outcome axis.
func TestPayloadSlotSwapMutationKilled(t *testing.T) {
	ctx := context.Background()
	runner := native.DefaultRunner()

	t.Run("mutated", func(t *testing.T) {
		restore := cgen.SetPayloadSlotSwapForTest(true)
		defer restore()

		program := checkedProgram(t, "testdata", "phase12", "payload_tracer.schway")
		engines := runFunctionOkExecutions(t, ctx, program, "identity", runner)
		injected := cgen.PayloadSlotSwapInjectedWriteCount()
		if injected < 1 {
			t.Fatalf("WR-02: the D-12-38 fault-injection seam found no alternative-mismatch target and therefore injected nothing (injected write count = %d), so this beat proved nothing about the mutation -- check whether payload_tracer.schway's data type still declares two payload-carrying alternatives", injected)
		}
		t.Logf("WR-02: fault-injection seam injected %d wrong-slot write(s)", injected)
		compareErr := session.Phase5CompareEngines("payload_tracer.schway(mutated)", engines)
		var disagreement *session.Phase5EngineDisagreement
		if !errors.As(compareErr, &disagreement) || disagreement.Axis != session.AxisTerminalOutcome {
			t.Fatalf("payload_tracer.schway(mutated): comparison error = %v, want exact %s disagreement", compareErr, session.AxisTerminalOutcome)
		}
		t.Logf("mutation KILLED on %s: %v", session.AxisTerminalOutcome, disagreement)
	})

	t.Run("companion_unmutated_agreement", func(t *testing.T) {
		// D-12-32/D-11-52: native emission does not support multi-function
		// branch bodies this phase, and payload_borrow_interaction.schway
		// deliberately declares TWO functions (choose, identity) in one
		// checked program (12-03-SUMMARY.md). isolateNativeFunction extracts
		// just "identity" (plus the Outcome/Fault data types it needs) so
		// this companion beat can still drive a genuinely DIFFERENT fixture
		// file than the mutated beat above, unmutated, through the same
		// native/interpreter comparison.
		multiFunction := checkedProgram(t, "testdata", "phase12", "payload_borrow_interaction.schway")
		program := isolateNativeFunction(t, multiFunction, "identity")
		engines := runFunctionOkExecutions(t, ctx, program, "identity", runner)
		if compareErr := session.Phase5CompareEngines("payload_borrow_interaction.schway(unmutated)", engines); compareErr != nil {
			t.Fatalf("payload_borrow_interaction.schway: expected agreement absent the mutation, got: %v", compareErr)
		}
	})
}

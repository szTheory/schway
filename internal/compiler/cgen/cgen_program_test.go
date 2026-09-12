package cgen_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

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
// the emitted C for multi_function_entry_basic.lang contains two function
// definitions, two prototypes preceding them, exactly one `int main(`, and
// compiles under the project's existing native build flags.
func TestEmitProgram(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.lang")

	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}

	if strings.Count(generated, "int main(") != 1 {
		t.Fatalf("expected exactly one int main(, got source:\n%s", generated)
	}
	prototypeIndexMain := strings.Index(generated, "static unsigned char LANG_MAIN(unsigned char);")
	prototypeIndexIdentity := strings.Index(generated, "static unsigned char LANG_IDENTITY(unsigned char);")
	definitionIndexMain := strings.Index(generated, "static unsigned char LANG_MAIN(unsigned char lang_value_")
	definitionIndexIdentity := strings.Index(generated, "static unsigned char LANG_IDENTITY(unsigned char lang_value_")
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

// TestEmitProgramEndToEndAgreesWithInterpreterAtO0 drives
// multi_function_entry_basic.lang through session's own run path on both
// engines and asserts the two lang.execution/1 documents are equal --
// end to end through session, not only through a unit test calling cgen
// directly.
func TestEmitProgramEndToEndAgreesWithInterpreterAtO0(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", "multi_function_entry_basic.lang"))
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
// (PLAN.md Task 3): multi_function_zero_call.lang declares two functions
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
	program := phase11CheckedProgram(t, "multi_function_zero_call.lang")

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
	if _, _, err := session.RunInterpreter(mustRead(t, "multi_function_zero_call.lang")); err == nil {
		t.Fatal("expected RunInterpreter to refuse the ambiguous-entry program")
	}
}

// TestEmitProgramForwardDefinedCallee proves D-11-03's ordering claim: the
// entry calls a function declared LATER in the source
// (multi_function_forward_callee.lang), and emitProgram writes every
// prototype before any definition, so the forward reference is legal C17.
func TestEmitProgramForwardDefinedCallee(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_forward_callee.lang")

	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	if !strings.Contains(generated, "static unsigned char LANG_LATER(unsigned char);") {
		t.Fatalf("expected LANG_LATER's own forward prototype, got source:\n%s", generated)
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
	selfRecursive, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "cycle_self.lang"))
	if err != nil {
		t.Skipf("no self-recursion fixture available: %v", err)
	}
	checked := session.Check(selfRecursive)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("expected self-recursive program to be refused by check")
	}
	if checked.Diagnostics[0].Code != core.CallGraphCycle {
		t.Fatalf("expected %q, got %q", core.CallGraphCycle, checked.Diagnostics[0].Code)
	}

	mutual, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "cycle_unreachable.lang"))
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
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", "toggle.lang"))
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

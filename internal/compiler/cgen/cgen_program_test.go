package cgen_test

import (
	"context"
	"fmt"
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
	program := checkedPhase07Program(t, "deep_diamond_acyclic.lang")
	got, err := cgen.InvocationPathNodeCountForTest(program)
	if err != nil {
		t.Fatalf("preflight deep_diamond_acyclic.lang: %v", err)
	}
	if got != 61 {
		t.Fatalf("deep_diamond_acyclic.lang invocation nodes = %d, want 61 (T0=1, T1=5, T2=13, T3=29, T4=61)", got)
	}
	if _, err := cgen.EmitNative(program); err != nil {
		t.Fatalf("EmitNative after 61-node preflight: %v", err)
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
	prototypeIndexMain := strings.Index(generated, "static unsigned char LANG_MAIN(unsigned char, unsigned int);")
	prototypeIndexIdentity := strings.Index(generated, "static unsigned char LANG_IDENTITY(unsigned char, unsigned int);")
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

func TestProgramInvocationIndexThreading(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_entry_basic.lang")
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	for _, want := range []string{
		"static unsigned char LANG_MAIN(unsigned char, unsigned int);",
		"static unsigned char LANG_IDENTITY(unsigned char, unsigned int);",
		"LANG_MAIN(lang_entry_input, 0u)",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("generated C is missing threaded invocation index %q:\n%s", want, generated)
		}
	}
}

func TestParentIndexedChildLookup(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_diamond_call.lang")
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	if !strings.Contains(generated, "static const char *lang_invocations[]") {
		t.Fatalf("generated C has no literal invocation table:\n%s", generated)
	}
	if !strings.Contains(generated, ":fn:main:op:0#0/s1:phase11.multi_function_diamond_call:fn:left:op:0#0") ||
		!strings.Contains(generated, ":fn:main:op:1#0/s1:phase11.multi_function_diamond_call:fn:right:op:0#0") {
		t.Fatalf("generated C does not retain distinct shared-leaf occurrences:\n%s", generated)
	}
	if !strings.Contains(generated, "lang_child_index_") {
		t.Fatalf("generated C has no parent-indexed child lookup:\n%s", generated)
	}
}

func TestInvocationTableEmissionIsDeterministic(t *testing.T) {
	program := phase11CheckedProgram(t, "multi_function_diamond_call.lang")
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
	if !strings.Contains(generated, "static unsigned char LANG_LATER(unsigned char, unsigned int);") {
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
	// plan 14-07: this fixture is checked into the repo
	// (testdata/phase07/cycle_self.lang); a read failure here is repo
	// corruption, not a legitimate skip condition, so it fails loudly
	// rather than silently skipping.
	selfRecursive, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "cycle_self.lang"))
	if err != nil {
		t.Fatalf("testdata/phase07/cycle_self.lang: %v", err)
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
	program := phase11CheckedProgram(t, "multi_function_entry_basic.lang")
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
	program := phase11CheckedProgram(t, "multi_function_entry_basic.lang")

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
	program := phase11CheckedProgram(t, "multi_function_gate_corpus.lang")

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
	program := phase11CheckedProgram(t, "multi_function_gate_corpus.lang")
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

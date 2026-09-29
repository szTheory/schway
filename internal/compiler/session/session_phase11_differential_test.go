package session_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase11DifferentialCorpus is the directory every fixture this file drives
// lives in -- the same testdata/phase11 corpus plans 11-03/11-04 committed.
func phase11DifferentialCorpus() string {
	return testsupport.ProjectPath("testdata", "phase11")
}

// phase11EntryInput derives the fixed, deterministic single-value input
// this file drives every comparable fixture with, selecting on the
// resolved entry function's own declared parameter type -- the same
// Byte/Buffer convention session.go's own interpreterInputs (D-11-05) and
// every existing Phase 4/5 differential in this package already establish.
// The default branch fails loudly rather than silently deriving a wrong
// input for a parameter type this file's corpus does not use.
func phase11EntryInput(t *testing.T, parameterType string) string {
	t.Helper()
	switch parameterType {
	case "Byte":
		return "7"
	case "Buffer":
		return "01020304"
	default:
		t.Fatalf("phase11EntryInput: unsupported entry parameter type %q", parameterType)
		return ""
	}
}

// phase11EntryFunction returns the checked, validated program's own
// resolved entry core.Function by name -- the same lookup
// TestPhase11SuppressionIsDiffLocal (session_phase11_gate_test.go) already
// establishes as this package's convention for recovering a core.Function
// by name from an already-checked core.Program.
func phase11EntryFunction(t *testing.T, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("phase11EntryFunction: function %q not found in program", name)
	return core.Function{}
}

// phase11CheckedFixture reads and checks fixture from
// phase11DifferentialCorpus, returning the validated core.Program and its
// resolved entry name via session.Phase4CheckedProgram (D-11-05: entry
// resolution goes through callgraph.EntryFunction, never Functions[0]).
func phase11CheckedFixture(t *testing.T, fixture string) (core.Program, string) {
	t.Helper()
	program, entryName, err := session.Phase4CheckedProgram(phase11DifferentialCorpus(), fixture)
	if err != nil {
		t.Fatalf("%s: Phase4CheckedProgram: %v", fixture, err)
	}
	return program, entryName
}

// phase16CheckedFixture uses the same checker and canonical entry resolver as
// the Phase 11 corpus while accepting the three retained N=1 fixtures. The
// native route selected below is deliberately direct emitProgram, not either
// public dispatcher.
func phase16CheckedFixture(t *testing.T, fixture string) (core.Program, string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath(fixture))
	if err != nil {
		t.Fatalf("%s: read: %v", fixture, err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: check diagnostics: %+v", fixture, checked.Diagnostics)
	}
	entry, err := callgraph.EntryFunction(checked.Program)
	if err != nil {
		t.Fatalf("%s: entry: %v", fixture, err)
	}
	return checked.Program, entry.Name
}

func phase16EntryInput(t *testing.T, fixture, parameterType string) string {
	t.Helper()
	if parameterType == "Switch" {
		switch fixture {
		case "testdata/phase1/toggle.schway":
			return "Off"
		case "testdata/phase3/borrowed_view.schway":
			return "On"
		}
	}
	return phase11EntryInput(t, parameterType)
}

// phase16ProjectInterpreterSchema2 is a comparison-boundary projection for
// the direct-program gate only. The N=1 interpreter legitimately retains /0
// or /1 while direct emitProgram emits /2, so the projection supplies the
// one entry invocation and removes legacy event-only input/output decoration
// before the independent /2 peer and all-pairs comparator inspect semantics.
func phase16ProjectInterpreterSchema2(t *testing.T, program core.Program, document execution.Execution) execution.Execution {
	t.Helper()
	projected, err := session.ProjectExecutionSchema2(program, document)
	if err != nil {
		t.Fatalf("project interpreter schema: %v", err)
	}
	return projected
}

type phase11NativeCSupplier func(core.Program) (string, error)

// phase11RunFourTiers drives program's resolved entry through the
// interpreter and all three native tiers -- -O0, -O3, and -O3 with -flto --
// for the SAME single input, returning all four execution.Execution
// documents keyed by engine name, ready for session.Phase5CompareEngines.
//
// D-11-25: because every Lang function this phase's cgen.emitProgram
// emits lands in ONE translation unit (D-11-24, native.Runner's single
// generated program.c, deliberately not widened this phase), `-flto` is
// inert by construction for Lang-to-Lang code in this corpus -- it is run
// here because running it costs nothing and a future translation-unit
// split would make it live, NOT because this differential exercises real
// cross-TU LTO inlining. The existing LTO lane's own non-inertness
// (session_phase5.go's phase5RunInterpreterO0O3LTOLane) is borrowed
// entirely from a foreign translation-unit boundary
// (inline_across_foreign.schway), which this pure Lang-to-Lang corpus has
// none of. NAT-07's hand-written control (plan 11-02) is what proves the
// TIER itself can be exploited when a real cross-TU boundary exists.
func phase11RunFourTiers(t *testing.T, ctx context.Context, program core.Program, entryName, input string) map[string]execution.Execution {
	return phase11RunFourTiersWithSupplier(t, ctx, program, entryName, input, cgen.EmitNative, "public EmitNative")
}

// phase11RunFourTiersWithSupplier preserves the interpreter and native runner
// lanes while allowing a narrow caller-selected C supplier. The supplier is
// named in every native failure so a direct-program gate cannot accidentally
// become evidence for public dispatch.
func phase11RunFourTiersWithSupplier(t *testing.T, ctx context.Context, program core.Program, entryName, input string, supplier phase11NativeCSupplier, route string) map[string]execution.Execution {
	t.Helper()

	interpreted, err := interp.Run(program, entryName, input)
	if err != nil {
		t.Fatalf("interp.Run(%s): %v", entryName, err)
	}

	cSource, err := supplier(program)
	if err != nil {
		t.Fatalf("%s: native C: %v", route, err)
	}

	runner := native.DefaultRunner()
	o0, err := runner.Run(ctx, cSource, "-O0", []string{input})
	if err != nil || len(o0.Pairs) != 1 {
		t.Fatalf("%s -O0 run failed: err=%v result=%+v", route, err, o0)
	}
	o3, err := runner.Run(ctx, cSource, "-O3", []string{input})
	if err != nil || len(o3.Pairs) != 1 {
		t.Fatalf("%s -O3 run failed: err=%v result=%+v", route, err, o3)
	}
	ltoRunner := runner
	ltoRunner.LTO = true
	o3lto, err := ltoRunner.Run(ctx, cSource, "-O3", []string{input})
	if err != nil || len(o3lto.Pairs) != 1 {
		t.Fatalf("%s -O3 -flto run failed: err=%v result=%+v", route, err, o3lto)
	}

	return map[string]execution.Execution{
		"interpreter": interpreted,
		"O0":          o0.Pairs[0].Execution,
		"O3":          o3.Pairs[0].Execution,
		"O3-LTO":      o3lto.Pairs[0].Execution,
	}
}

// phase11CompareFourTiers is this file's own end-to-end drive-and-compare
// step: check the named fixture, run all four tiers over its resolved
// entry with a fixed, type-derived input, and assert the comparator's
// five axes (axis:terminal-outcome, axis:event-order,
// axis:resource-ledger, axis:exit-status-signal, axis:diagnostic-id --
// the COMPARATOR'S five axes, distinct from the QLT-03 shape register's
// own five-axis taxonomy referenced elsewhere in this phase) agree for
// EVERY tier pair via session.Phase5CompareEngines, not merely
// interpreter-vs-O0. Returns the program and the engines map so a caller
// needing an additional structural assertion (e.g. UnreachableFunction's
// event-absence check) does not have to re-derive either.
func phase11CompareFourTiers(t *testing.T, ctx context.Context, fixture string) (core.Program, map[string]execution.Execution) {
	t.Helper()
	program, entryName := phase11CheckedFixture(t, fixture)
	entry := phase11EntryFunction(t, program, entryName)
	input := phase11EntryInput(t, entry.Parameter.Type)
	engines := phase11RunFourTiers(t, ctx, program, entryName, input)
	if err := session.Phase5CompareProgramEngines(fixture, program, engines); err != nil {
		t.Fatalf("%s: four-tier disagreement: %v", fixture, err)
	}
	return program, engines
}

func phase16CompareDirectProgramFourTiers(t *testing.T, ctx context.Context, fixture string) (core.Program, map[string]execution.Execution, string) {
	t.Helper()
	program, entryName := phase16CheckedFixture(t, fixture)
	entry := phase11EntryFunction(t, program, entryName)
	input := phase16EntryInput(t, fixture, entry.Parameter.Type)
	directC, err := cgen.EmitProgramNativeForTest(program)
	if err != nil {
		t.Fatalf("%s: direct emitProgram(..., true): %v", fixture, err)
	}
	if !strings.Contains(directC, "schway.execution/2") {
		t.Fatalf("%s: direct emitProgram(..., true) did not supply schema-2 C", fixture)
	}
	engines := phase11RunFourTiersWithSupplier(t, ctx, program, entryName, input, func(core.Program) (string, error) {
		return directC, nil
	}, "direct emitProgram(..., true)")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, program, engines["interpreter"])
	// Bare matches have no core.LinearOperation ID for executionpeer to
	// classify, so this direct gate deliberately uses the all-pairs semantic
	// comparator rather than pretending the peer can validate a synthetic
	// bare-match return operation. Branches with linear arms keep their peer
	// coverage in the existing program-emitter tests.
	if err := session.Phase5CompareEngines(fixture, engines); err != nil {
		t.Fatalf("%s: direct emitProgram four-tier disagreement: %v", fixture, err)
	}
	return program, engines, directC
}

// TestPhase16DirectProgramFourTierDifferential keeps semantic evidence
// independent from the byte/provenance gates: each native optimization lane
// compiles C supplied directly by emitProgram(..., true), rather than C from
// the retained public N=1 dispatcher.
func TestPhase16DirectProgramFourTierDifferential(t *testing.T) {
	ctx := context.Background()
	for _, test := range []struct {
		name, fixture string
	}{
		{name: "Toggle", fixture: "testdata/phase1/toggle.schway"},
		{name: "OwnedTransfer", fixture: "testdata/phase2/owned_transfer.schway"},
		{name: "BorrowedView", fixture: "testdata/phase3/borrowed_view.schway"},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			_, engines, _ := phase16CompareDirectProgramFourTiers(t, ctx, test.fixture)
			if len(engines) != 4 {
				t.Fatalf("%s: direct route supplied %d engine documents, want 4", test.fixture, len(engines))
			}
		})
	}
}

// TestPhase16EmitterPortSemanticGuardIsNotInert mutates a compared semantic
// document after its direct-program C bytes have been fixed. The C digest is
// therefore still internally consistent, yet the independent comparator must
// reject the changed terminal-outcome axis; this is why a golden hash cannot
// serve as the semantic witness for its own update.
func TestPhase16EmitterPortSemanticGuardIsNotInert(t *testing.T) {
	program, engines, directC := phase16CompareDirectProgramFourTiers(t, context.Background(), "testdata/phase2/owned_transfer.schway")
	before := sha256.Sum256([]byte(directC))
	mutated := engines["O3"]
	mutated.Outcome.Value = "phase16-seeded-semantic-mutation"
	engines["O3"] = mutated
	after := sha256.Sum256([]byte(directC))
	if before != after {
		t.Fatalf("direct-program C unexpectedly changed during document mutation: before=%s after=%s", hex.EncodeToString(before[:]), hex.EncodeToString(after[:]))
	}
	err := session.Phase5CompareProgramEngines("phase16-seeded-semantic-mutation", program, engines)
	if err == nil {
		t.Fatal("seeded semantic mutation stayed green")
	}
	var disagreement *session.Phase5EngineDisagreement
	if !errors.As(err, &disagreement) || disagreement.Axis != session.AxisTerminalOutcome {
		t.Fatalf("seeded semantic mutation must change terminal-outcome comparator axis, got %v", err)
	}
}

// TestPhase16Schema2ProjectionRejectsNativeFactMutation proves that expected
// schema-2 facts remain derived from the checked program and interpreter
// document. In particular, no native invocation or defect reason can be
// copied back into the expected projection to conceal a broken emitter.
func TestPhase16Schema2ProjectionRejectsNativeFactMutation(t *testing.T) {
	program, entryName := phase16CheckedFixture(t, "testdata/phase5/defect_dies_by_signal.schway")
	interpreted, err := interp.Run(program, entryName, "Halt")
	if err != nil {
		t.Fatalf("interp.Run: %v", err)
	}
	expected := phase16ProjectInterpreterSchema2(t, program, interpreted)
	if len(expected.Events) == 0 {
		t.Fatal("fixture produced no expected events")
	}
	mutatedInvocation := expected
	mutatedInvocation.Events = append([]execution.Event(nil), expected.Events...)
	mutatedInvocation.Events[0].Invocation = "native-injected-invocation"
	if execution.Equal(expected, mutatedInvocation) {
		t.Fatal("native invocation mutation was accepted by schema-2 comparison")
	}
	mutatedReason := expected
	mutatedReason.Events = append([]execution.Event(nil), expected.Events...)
	foundDefect := false
	for index := range mutatedReason.Events {
		if mutatedReason.Events[index].Kind == "function.defected" {
			mutatedReason.Events[index].Output = "native-injected-defect-reason"
			foundDefect = true
			break
		}
	}
	if !foundDefect {
		t.Fatal("fixture produced no defect event")
	}
	if execution.Equal(expected, mutatedReason) {
		t.Fatal("native defect-reason mutation was accepted by schema-2 comparison")
	}
}

// TestPhase11InterproceduralDifferential is NAT-06's own four-tier
// interprocedural differential (PLAN.md Task 2): the interpreter and all
// three native tiers must agree, on the comparator's five axes, over the
// full testdata/phase11 corpus (plans 11-03/11-04's own fixtures plus this
// plan's two new ones), including the structural edge cases NAT-06's
// must_haves require -- zero call edges, an unreachable function, a
// forward-defined callee, a diamond with a shared leaf, and a diverging
// callee.
func TestPhase11InterproceduralDifferential(t *testing.T) {
	ctx := context.Background()

	// AllComparableFixtures sweeps every testdata/phase11/*.schway fixture
	// that resolves to a unique entry and is expected to agree across all
	// four tiers -- every existing corpus member except the cut M004
	// multi_function_gate_corpus.schway (refusal is covered by
	// probe:TestPhase16M004CorpusRefusal) and multi_function_zero_call.schway,
	// whose own genuinely ambiguous entry
	// (ZeroCallEdges, below) means it never reaches any tier at all, by
	// design (D-11-05's "never guess" prohibition).
	t.Run("AllComparableFixtures", func(t *testing.T) {
		for _, fixture := range []string{
			"multi_function_entry_basic.schway",
			"multi_function_forward_callee.schway",
			"multi_function_unreachable.schway",
			"multi_function_relay_depth2.schway",
		} {
			t.Run(fixture, func(t *testing.T) {
				phase11CompareFourTiers(t, ctx, fixture)
			})
		}
	})

	// ZeroCallEdges: multi_function_zero_call.schway declares two functions
	// with ZERO core.OpCall operations anywhere in the program. Both are
	// therefore in-degree-zero candidates with an EMPTY reachable closure
	// each -- a genuine, unbreakable tie under callgraph.EntryFunction's
	// own documented closure-size tie-break (11-03-SUMMARY.md Deviation 3;
	// cgen_program_test.go's own TestEmitProgramZeroCallEdges). No tier can
	// ever produce an execution.Execution document for this program at
	// all: entry resolution itself refuses first, before check.Program's
	// own admission even matters. "Agrees on all four tiers" for this
	// fixture is therefore satisfied by every tier-reaching call site
	// (RunInterpreter, and the ONE shared cgen.EmitNative lowering step
	// every native tier's own compile is downstream of) refusing with the
	// IDENTICAL named core.EntryAmbiguous code -- never one tier silently
	// succeeding while another disagrees, never a crash, never a
	// Functions[0]-style guess. This deliberately diverges from a literal
	// reading of "agrees" as "produces four comparable execution
	// documents", for the same reason 11-03-SUMMARY.md's own Deviation 3
	// diverged from this task's ancestor plan's literal wording: that
	// reading is unsatisfiable for a genuinely tied, zero-call,
	// two-function program under a non-guessing resolver.
	t.Run("ZeroCallEdges", func(t *testing.T) {
		const fixture = "multi_function_zero_call.schway"
		source := phase11ReadFixture(t, fixture)

		if _, _, err := session.RunInterpreter(source); err == nil {
			t.Fatal("expected RunInterpreter to refuse the ambiguous-entry program")
		} else if ambiguous, ok := callgraph.EntryAmbiguousError(err); !ok || ambiguous.Code() != core.EntryAmbiguous {
			t.Fatalf("expected a callgraph entry-ambiguity refusal from RunInterpreter, got: %v", err)
		}

		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("expected %s to check cleanly (the refusal is at entry resolution, not check), got diagnostics: %+v", fixture, checked.Diagnostics)
		}
		if _, err := cgen.EmitNative(checked.Program); err == nil {
			t.Fatal("expected cgen.EmitNative to refuse the ambiguous-entry program")
		} else if ambiguous, ok := callgraph.EntryAmbiguousError(err); !ok || ambiguous.Code() != core.EntryAmbiguous {
			t.Fatalf("expected a callgraph entry-ambiguity refusal from EmitNative (the one lowering step every native tier's compile shares), got: %v", err)
		}
	})

	// UnreachableFunction: multi_function_unreachable.schway's `orphan` is
	// declared but never called and never exported. It must still agree
	// across all four tiers (main's own resolution and execution are
	// unaffected by orphan's presence), AND its absence from the event
	// stream must be asserted BY FUNCTION ID -- never inferred from a
	// matching total event count, which could pass vacuously if two
	// distinct functions happened to contribute the same number of events.
	t.Run("UnreachableFunction", func(t *testing.T) {
		const fixture = "multi_function_unreachable.schway"
		program, engines := phase11CompareFourTiers(t, ctx, fixture)

		var orphanID string
		for _, function := range program.Functions {
			if function.Name == "orphan" {
				orphanID = function.ID
			}
		}
		if orphanID == "" {
			t.Fatalf("%s: expected a declared function named orphan", fixture)
		}
		for engineName, doc := range engines {
			for _, event := range doc.Events {
				if event.FunctionID == orphanID {
					t.Fatalf("%s: engine %q emitted an event for the unreachable function %q: %+v", fixture, engineName, orphanID, event)
				}
			}
		}
	})

	// ForwardDefinedCallee: multi_function_forward_callee.schway's `main`
	// calls `later`, declared AFTER it in source order -- must agree
	// across all four tiers.
	t.Run("ForwardDefinedCallee", func(t *testing.T) {
		phase11CompareFourTiers(t, ctx, "multi_function_forward_callee.schway")
	})

	// DiamondSharedLeaf: multi_function_diamond_call.schway's `main` calls
	// two distinct callees (`left`, `right`) that both call a shared leaf
	// (`leaf`) -- must agree across all four tiers.
	//
	// The shared leaf is deliberately invoked from two distinct call sites.
	// Schema /2 preserves the old static IDs but makes the full
	// (Invocation, ID) identity pair unique, so this exact fixture is the
	// permanent four-tier occurrence-identity gate.
	t.Run("DiamondSharedLeaf", func(t *testing.T) {
		const fixture = "multi_function_diamond_call.schway"
		program, entryName := phase11CheckedFixture(t, fixture)
		entry := phase11EntryFunction(t, program, entryName)
		input := phase11EntryInput(t, entry.Parameter.Type)
		engines := phase11RunFourTiers(t, ctx, program, entryName, input)
		if len(engines) != 4 {
			t.Fatalf("%s: expected all four engine documents, got %d", fixture, len(engines))
		}
		for _, name := range []string{"interpreter", "O0", "O3", "O3-LTO"} {
			document, ok := engines[name]
			if !ok {
				t.Fatalf("%s: missing %s document", fixture, name)
			}
			seen := make(map[string]bool, len(document.Events))
			for _, event := range document.Events {
				pair := event.Invocation + "\x00" + event.ID
				if seen[pair] {
					t.Fatalf("%s: %s duplicated occurrence identity %q", fixture, name, pair)
				}
				seen[pair] = true
			}
		}
		// This validates each document with the independent /2 peer (including
		// strict preorder and caller-owned call edges) before comparing every
		// one of the six engine pairs.
		if err := session.Phase5CompareProgramEngines(fixture, program, engines); err != nil {
			t.Fatalf("%s: four-tier occurrence evidence disagreed: %v", fixture, err)
		}
	})

	// DivergingCallee: a fixture whose callee does not return normally
	// (terminates via `defect`) should produce the same terminal outcome
	// and the same event prefix on all four tiers, asserting
	// axis:exit-status-signal specifically.
	//
	// NOT EXPRESSIBLE at this maturity (honest "not expressible" per this
	// task's own escape hatch, not a silently absent case): every existing
	// `defect` terminator in this codebase (defect_terminal.schway,
	// defect_dies_by_signal.schway) is reached through a `core.Match` arm --
	// there is no other Lang construct capable of reaching `defect` (no
	// arithmetic, no `if`, no loops, this milestone's own established
	// maturity ceiling). cgen.emitProgram's own doc comment and its own
	// explicit runtime refusal ("multi-function branch bodies are not
	// supported by native emission this phase") make ANY Match-bodied
	// function in a multi-function program unemittable by the whole-program
	// assembler this phase built (11-03-SUMMARY.md: "straight-line bodies
	// only ... no core.Match"). A diverging callee therefore cannot be
	// lowered to native at all in a multi-function program this phase --
	// not merely a fixture-authoring inconvenience, a structural
	// consequence of D-11-02's own scope decision (the six single-function
	// emitters, including emitMatch, are deliberately not deleted or
	// generalized this phase). Recorded in PHASE-11-DEBT.md.
	t.Run("DivergingCallee", func(t *testing.T) {
		t.Skip("not expressible this phase: a diverging callee requires a core.Match-bodied function (the only construct that can reach `defect` at this maturity), and cgen.emitProgram explicitly refuses any Match-bodied function in a multi-function program (\"multi-function branch bodies are not supported by native emission this phase\", D-11-02) -- see PHASE-11-DEBT.md, witnessed by probe:TestLTOInertnessOnMultiFunctionEmission (internal/compiler/session/witness_registry_test.go)")
	})
}

// phase11ReadFixture reads fixture's raw source bytes from
// phase11DifferentialCorpus, for the one subtest (ZeroCallEdges) that must
// drive session.RunInterpreter and session.Check directly rather than
// through the checked-program helper (an ambiguous-entry program's own
// refusal happens AFTER check.Program's own clean admission, so this
// subtest needs the raw source to prove that ordering).
func phase11ReadFixture(t *testing.T, fixture string) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", fixture))
	if err != nil {
		t.Fatalf("%s: %v", fixture, err)
	}
	return source
}

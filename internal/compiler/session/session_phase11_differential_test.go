package session_test

import (
	"context"
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
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
// (inline_across_foreign.lang), which this pure Lang-to-Lang corpus has
// none of. NAT-07's hand-written control (plan 11-02) is what proves the
// TIER itself can be exploited when a real cross-TU boundary exists.
func phase11RunFourTiers(t *testing.T, ctx context.Context, program core.Program, entryName, input string) map[string]execution.Execution {
	t.Helper()

	interpreted, err := interp.Run(program, entryName, input)
	if err != nil {
		t.Fatalf("interp.Run(%s): %v", entryName, err)
	}

	cSource, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("cgen.EmitNative: %v", err)
	}

	runner := native.DefaultRunner()
	o0, err := runner.Run(ctx, cSource, "-O0", []string{input})
	if err != nil || len(o0.Pairs) != 1 {
		t.Fatalf("-O0 run failed: err=%v result=%+v", err, o0)
	}
	o3, err := runner.Run(ctx, cSource, "-O3", []string{input})
	if err != nil || len(o3.Pairs) != 1 {
		t.Fatalf("-O3 run failed: err=%v result=%+v", err, o3)
	}
	ltoRunner := runner
	ltoRunner.LTO = true
	o3lto, err := ltoRunner.Run(ctx, cSource, "-O3", []string{input})
	if err != nil || len(o3lto.Pairs) != 1 {
		t.Fatalf("-O3 -flto run failed: err=%v result=%+v", err, o3lto)
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
	if err := session.Phase5CompareEngines(fixture, engines); err != nil {
		t.Fatalf("%s: four-tier disagreement: %v", fixture, err)
	}
	return program, engines
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

	// AllComparableFixtures sweeps every testdata/phase11/*.lang fixture
	// that resolves to a unique entry and is expected to agree across all
	// four tiers -- every existing corpus member except
	// multi_function_zero_call.lang, whose own genuinely ambiguous entry
	// (ZeroCallEdges, below) means it never reaches any tier at all, by
	// design (D-11-05's "never guess" prohibition).
	t.Run("AllComparableFixtures", func(t *testing.T) {
		for _, fixture := range []string{
			"multi_function_entry_basic.lang",
			"multi_function_forward_callee.lang",
			"multi_function_unreachable.lang",
			"multi_function_gate_corpus.lang",
			"multi_function_relay_depth2.lang",
		} {
			t.Run(fixture, func(t *testing.T) {
				phase11CompareFourTiers(t, ctx, fixture)
			})
		}
	})

	// ZeroCallEdges: multi_function_zero_call.lang declares two functions
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
		const fixture = "multi_function_zero_call.lang"
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

	// UnreachableFunction: multi_function_unreachable.lang's `orphan` is
	// declared but never called and never exported. It must still agree
	// across all four tiers (main's own resolution and execution are
	// unaffected by orphan's presence), AND its absence from the event
	// stream must be asserted BY FUNCTION ID -- never inferred from a
	// matching total event count, which could pass vacuously if two
	// distinct functions happened to contribute the same number of events.
	t.Run("UnreachableFunction", func(t *testing.T) {
		const fixture = "multi_function_unreachable.lang"
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

	// ForwardDefinedCallee: multi_function_forward_callee.lang's `main`
	// calls `later`, declared AFTER it in source order -- must agree
	// across all four tiers.
	t.Run("ForwardDefinedCallee", func(t *testing.T) {
		phase11CompareFourTiers(t, ctx, "multi_function_forward_callee.lang")
	})

	// DiamondSharedLeaf: multi_function_diamond_call.lang's `main` calls
	// two distinct callees (`left`, `right`) that both call a shared leaf
	// (`leaf`) -- must agree across all four tiers.
	//
	// FINDING (Rule 1-adjacent, architectural, flagged for human review,
	// NOT silently fixed): this fixture is the FIRST program in this
	// repository ever actually EXECUTED (as opposed to merely checked --
	// testdata/phase07/deep_diamond_acyclic.lang's own shared-leaf diamond
	// is a check-only corpus member, never driven through interp.Run or
	// native emission) whose call graph invokes the SAME callee from TWO
	// DISTINCT static call sites. Both interp.Run and cgen.emitProgram
	// derive an executed function's `function.returned` event ID from that
	// function's OWN STATIC OpReturn operation ID (interp.go's
	// terminalOutcome, cgen_program.go's emitProgramFunction) -- an
	// identity that is per-DECLARATION, not per-INVOCATION. `leaf` is
	// invoked twice in one run (once via `left`, once via `right`), so it
	// emits the IDENTICAL `function.returned` event ID twice. interp.Run
	// itself performs no duplicate-ID validation and returns such a
	// document successfully; native.go's OWN decode-time validator
	// (validateExecution, "duplicate execution event id") correctly
	// refuses to trust ANY decoded execution document -- native or
	// otherwise -- carrying two events with the same ID, since ID
	// uniqueness is exactly the invariant this project's own causal-event
	// tracking (D-04's own event-identity convention) depends on. The
	// refusal is IDENTICAL and consistent across -O0, -O3, and -O3 -flto
	// (all three share the one cgen.EmitNative lowering; the divergence is
	// structural, not optimizer-dependent), so there is no cross-tier
	// DISAGREEMENT here -- but there is no four-tier AGREEMENT on a
	// comparable execution document either, since three of the four tiers
	// never produce one. Fixing this for real requires giving each
	// function's events a per-INVOCATION-unique identity (e.g. threading a
	// call-site-qualified suffix through both interp.Run's frame stack and
	// cgen.emitProgram's per-function event-ID derivation, independently,
	// so the two engines' identity schemes keep agreeing) -- a change to
	// interp.go and cgen_program.go, neither of which is in this plan's own
	// declared files_modified, and a big enough change to both engines'
	// shared event-identity convention that it needs its own reviewed
	// plan, not a same-task patch. Recorded as new debt (PHASE-11-DEBT.md)
	// rather than silently worked around by weakening this test's own
	// assertions or quietly picking an easier "diamond" that never
	// actually re-invokes the shared leaf (which would misreport a
	// narrower proof as the wider one Task 1's own prohibition forbids).
	t.Run("DiamondSharedLeaf", func(t *testing.T) {
		const fixture = "multi_function_diamond_call.lang"
		program, entryName := phase11CheckedFixture(t, fixture)
		entry := phase11EntryFunction(t, program, entryName)
		input := phase11EntryInput(t, entry.Parameter.Type)

		interpreted, err := interp.Run(program, entryName, input)
		if err != nil {
			t.Fatalf("%s: interp.Run: %v", fixture, err)
		}
		if interpreted.Outcome.Kind != execution.OutcomeReturned {
			t.Fatalf("%s: interpreter: expected outcome kind %q, got %q", fixture, execution.OutcomeReturned, interpreted.Outcome.Kind)
		}
		seen := make(map[string]bool, len(interpreted.Events))
		duplicated := false
		for _, event := range interpreted.Events {
			if seen[event.ID] {
				duplicated = true
			}
			seen[event.ID] = true
		}
		if !duplicated {
			t.Fatalf("%s: expected the interpreter's own document to exhibit the documented duplicate-event-ID finding; if this now passes, the underlying gap has been fixed and this test (and PHASE-11-DEBT.md's matching row) should be updated, not left stale", fixture)
		}

		cSource, err := cgen.EmitNative(program)
		if err != nil {
			t.Fatalf("%s: cgen.EmitNative unexpectedly refused a shared-leaf diamond at the LOWERING step (expected it to lower cleanly and fail only at native decode time): %v", fixture, err)
		}
		runner := native.DefaultRunner()
		for _, optimization := range []string{"-O0", "-O3"} {
			if _, err := runner.Run(ctx, cSource, optimization, []string{input}); err == nil {
				t.Fatalf("%s: expected %s to refuse with a duplicate-event-ID decode error, got a clean run -- the documented finding may have been fixed without updating this test", fixture, optimization)
			}
		}
		ltoRunner := runner
		ltoRunner.LTO = true
		if _, err := ltoRunner.Run(ctx, cSource, "-O3", []string{input}); err == nil {
			t.Fatalf("%s: expected -O3 -flto to refuse with a duplicate-event-ID decode error, got a clean run", fixture)
		}
	})

	// DivergingCallee: a fixture whose callee does not return normally
	// (terminates via `defect`) should produce the same terminal outcome
	// and the same event prefix on all four tiers, asserting
	// axis:exit-status-signal specifically.
	//
	// NOT EXPRESSIBLE at this maturity (honest "not expressible" per this
	// task's own escape hatch, not a silently absent case): every existing
	// `defect` terminator in this codebase (defect_terminal.lang,
	// defect_dies_by_signal.lang) is reached through a `core.Match` arm --
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
		t.Skip("not expressible this phase: a diverging callee requires a core.Match-bodied function (the only construct that can reach `defect` at this maturity), and cgen.emitProgram explicitly refuses any Match-bodied function in a multi-function program (\"multi-function branch bodies are not supported by native emission this phase\", D-11-02) -- see PHASE-11-DEBT.md")
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

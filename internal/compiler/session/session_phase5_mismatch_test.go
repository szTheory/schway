package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/reduce"
	"github.com/szTheory/schway/internal/compiler/session"
)

// ---------------------------------------------------------------------
// Phase 11 plan 11-09 (D-11-33): foreignCallSequenceFor's two-knower guard
// against reduction slippage. Fixtures are built by hand exactly like
// reduce_multifunction_test.go's own multi-function fixtures -- this
// file's own moves operate on core.Program structure directly, never
// through the parser/checker.
// ---------------------------------------------------------------------

const (
	fcsMainID   = "s1:phase11.fcs_fixture:function:main"
	fcsWorkerID = "s1:phase11.fcs_fixture:function:worker"
)

// foreignCallSequenceFixture is a two-function program: main calls worker
// once, and worker -- a NON-entry function -- makes a single foreign
// acquisition. Before this plan's fix, foreignCallSequenceFor returned nil
// unconditionally for any multi-function program; this fixture is exactly
// the shape that fix must cover.
func foreignCallSequenceFixture() core.Program {
	const module = "phase11.fcs_fixture"
	typeID := fcsMainID + ":type:0"
	return core.Program{
		Schema:   core.Schema,
		Module:   module,
		ModuleID: "s1:" + module + ":module:" + module,
		Functions: []core.Function{
			{
				ID: fcsMainID, Name: "main",
				EntryPointID: fcsMainID + ":point:entry", ReturnPointID: fcsMainID + ":point:return",
				Parameter:  core.Parameter{ID: fcsMainID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:    fcsMainID + ":linear",
					Types: []core.TypeFact{{ID: typeID}},
					Places: []core.Place{
						{ID: fcsMainID + ":place:0", Name: "value", TypeID: typeID},
						{ID: fcsMainID + ":place:1", Name: "result", TypeID: typeID},
					},
					Operations: []core.LinearOperation{
						{ID: fcsMainID + ":op:0", PointID: fcsMainID + ":point:linear:0", Kind: core.OpCall, SourceID: fcsMainID + ":place:0", TargetID: fcsMainID + ":place:1", TypeID: typeID, CalleeID: fcsWorkerID},
						{ID: fcsMainID + ":op:1", PointID: fcsMainID + ":point:linear:1", Kind: core.OpReturn, SourceID: fcsMainID + ":place:1", TypeID: typeID},
					},
				},
			},
			{
				ID: fcsWorkerID, Name: "worker",
				EntryPointID: fcsWorkerID + ":point:entry", ReturnPointID: fcsWorkerID + ":point:return",
				Parameter:       core.Parameter{ID: fcsWorkerID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType:      "Byte",
				ForeignContract: &core.ForeignContract{Symbol: "acquire_resource", Allocator: "malloc"},
				Linear: &core.LinearBody{
					ID:    fcsWorkerID + ":linear",
					Types: []core.TypeFact{{ID: typeID}},
					Places: []core.Place{
						{ID: fcsWorkerID + ":place:0", Name: "value", TypeID: typeID},
						{ID: fcsWorkerID + ":place:1", Name: "acquired", TypeID: typeID},
					},
					Operations: []core.LinearOperation{
						{ID: fcsWorkerID + ":op:0", PointID: fcsWorkerID + ":point:linear:0", Kind: core.OpForeignCall, SourceID: fcsWorkerID + ":place:0", TargetID: fcsWorkerID + ":place:1", TypeID: typeID},
						{ID: fcsWorkerID + ":op:1", PointID: fcsWorkerID + ":point:linear:1", Kind: core.OpReturn, SourceID: fcsWorkerID + ":place:1", TypeID: typeID},
					},
				},
			},
		},
	}
}

// foreignCallSequenceFixtureExecution is a hand-built O0-shaped
// execution.Execution consistent with foreignCallSequenceFixture's own
// static structure: one "foreign.called" event, FunctionID naming worker,
// mirroring exactly what a real -O0 run of this program would record
// (interp.go/native.go's own shared event-kind convention).
func foreignCallSequenceFixtureExecution() execution.Execution {
	return execution.Execution{
		Schema: execution.Schema1,
		Events: []execution.Event{
			{Schema: execution.Schema1, ID: fcsWorkerID + ":op:0:event", Kind: "foreign.called", FunctionID: fcsWorkerID, SourcePlace: fcsWorkerID + ":place:0", TargetPlace: fcsWorkerID + ":place:1"},
		},
	}
}

// TestForeignCallSequenceCoversAllFunctions asserts the static sequence is
// non-nil and contains a foreign acquisition made by a NON-entry function
// -- the exact bug D-11-33 fixes: a nil that looked like agreement.
func TestForeignCallSequenceCoversAllFunctions(t *testing.T) {
	program := foreignCallSequenceFixture()
	sequence, err := session.StaticForeignCallSequenceForTest(program)
	if err != nil {
		t.Fatalf("StaticForeignCallSequenceForTest: %v", err)
	}
	if sequence == nil {
		t.Fatal("expected a non-nil static foreign-call sequence for a multi-function program with a foreign acquisition in a non-entry function")
	}
	found := false
	for _, symbol := range sequence {
		if symbol == "acquire_resource" {
			found = true
		}
	}
	if !found {
		t.Fatalf("sequence %v does not contain worker's own foreign acquisition", sequence)
	}
}

// TestForeignCallSequenceStaticAndDynamicAgree asserts the two independent
// derivations produce equal sequences on the multi-function corpus.
func TestForeignCallSequenceStaticAndDynamicAgree(t *testing.T) {
	program := foreignCallSequenceFixture()
	run := foreignCallSequenceFixtureExecution()
	sequence, err := session.ForeignCallSequenceForTest(program, run)
	if err != nil {
		t.Fatalf("ForeignCallSequenceForTest: %v", err)
	}
	static, err := session.StaticForeignCallSequenceForTest(program)
	if err != nil {
		t.Fatalf("StaticForeignCallSequenceForTest: %v", err)
	}
	dynamic := session.DynamicForeignCallSequenceForTest(program, run)
	if len(static) != len(dynamic) {
		t.Fatalf("static %v and dynamic %v derivations disagree in length", static, dynamic)
	}
	for i := range static {
		if static[i] != dynamic[i] {
			t.Fatalf("static %v and dynamic %v derivations disagree at index %d", static, dynamic, i)
		}
	}
	if len(sequence) == 0 {
		t.Fatal("expected a non-empty agreed sequence")
	}
}

// TestForeignCallSequenceDetectsDroppedCallToForeignCallee constructs the
// "after" state of a reduction that drops main's only call to worker
// (drop-call-site rewrites the OpCall to OpCopy) and, once worker has no
// remaining callers, removes worker entirely (drop-orphan-function) --
// exactly what Reduce's own pass structure produces once both whole-
// program moves run to exhaustion (11-08-SUMMARY.md). The two whole-
// program moves are modeled directly here (not invoked via reduce's own
// unexported functions, which are not visible outside package reduce's
// own test binary) since this test is about foreignCallSequenceFor's own
// reaction to the resulting program shape, not about the moves'
// correctness (already covered in reduce_multifunction_test.go). Before
// D-11-33's fix, this sequence would have been nil on BOTH sides (any
// multi-function program returned nil unconditionally) -- a real dropped
// call to a foreign-acquiring callee that never moved the comparison at
// all. After the fix, it must move.
func TestForeignCallSequenceDetectsDroppedCallToForeignCallee(t *testing.T) {
	before := foreignCallSequenceFixture()
	beforeRun := foreignCallSequenceFixtureExecution()
	beforeSequence, err := session.ForeignCallSequenceForTest(before, beforeRun)
	if err != nil {
		t.Fatalf("ForeignCallSequenceForTest(before): %v", err)
	}
	if len(beforeSequence) == 0 {
		t.Fatal("expected a non-empty sequence before the drop")
	}

	after := foreignCallSequenceFixture()
	after.Functions[0].Linear.Operations[0] = core.LinearOperation{
		ID: fcsMainID + ":op:0", PointID: fcsMainID + ":point:linear:0",
		Kind: core.OpCopy, SourceID: fcsMainID + ":place:0", TargetID: fcsMainID + ":place:1", TypeID: fcsMainID + ":type:0",
	}
	after.Functions = after.Functions[:1] // drop-orphan-function: worker has zero remaining callers.
	afterRun := execution.Execution{Schema: execution.Schema1}

	afterSequence, err := session.ForeignCallSequenceForTest(after, afterRun)
	if err != nil {
		t.Fatalf("ForeignCallSequenceForTest(after): %v", err)
	}
	if len(afterSequence) != 0 {
		t.Fatalf("expected an empty sequence after dropping the only call to the foreign-acquiring callee, got %v", afterSequence)
	}
	if len(beforeSequence) == len(afterSequence) {
		t.Fatal("expected the sequence to change after the drop -- this is the slippage D-11-33 fixes")
	}
}

// TestForeignCallSequenceSeededDisagreement proves the agreement check
// between the two derivations is load-bearing rather than trivially true:
// with the dynamic derivation perturbed via this plan's own fault-
// injection seam, foreignCallSequenceFor must FAIL.
func TestForeignCallSequenceSeededDisagreement(t *testing.T) {
	restore := session.SetForeignCallSequenceDisagreementForTest(func(dynamic []string) []string {
		return append(append([]string(nil), dynamic...), "seeded-disagreement")
	})
	defer restore()

	program := foreignCallSequenceFixture()
	run := foreignCallSequenceFixtureExecution()
	if _, err := session.ForeignCallSequenceForTest(program, run); err == nil {
		t.Fatal("expected foreignCallSequenceFor to fail once the dynamic derivation is seeded to disagree with the static one")
	}
}

// TestMismatchDocumentEmittedOnSeededDivergence drives plan 05-07's
// AliasFactMutationRunner to a REAL divergence over
// testdata/phase5/false_restrict_hoist.schway and asserts a lang.mismatch/0
// document is emitted with a non-empty reduced_source, a diverging_axis
// matching the comparator's own verdict, and a minimality of fixpoint or
// budget_exhausted (D-05-26).
func TestMismatchDocumentEmittedOnSeededDivergence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	document, err := session.ReduceSeededAliasMismatch(ctx)
	if err != nil {
		requirePhase16M004Refusal(t, err, "by-pointer")
		return
	}
	if document.Schema != reduce.MismatchSchema {
		t.Fatalf("schema = %q, want %q", document.Schema, reduce.MismatchSchema)
	}
	if document.DivergingAxis != session.AxisTerminalOutcome {
		t.Fatalf("diverging_axis = %q, want the comparator's own verdict %q", document.DivergingAxis, session.AxisTerminalOutcome)
	}
	if document.ReducedSource == "" {
		t.Fatal("expected a non-empty reduced_source")
	}
	if document.Minimality != reduce.MinimalityFixpoint && document.Minimality != reduce.MinimalityBudgetExhausted {
		t.Fatalf("minimality = %q, want %q or %q", document.Minimality, reduce.MinimalityFixpoint, reduce.MinimalityBudgetExhausted)
	}
	if document.EnginePair == "" {
		t.Fatal("expected a non-empty engine_pair")
	}
	if document.EvidenceID == "" {
		t.Fatal("expected a non-empty evidence_id (the additive discretion field)")
	}
}

// TestMismatchDocumentIsSelfSufficient asserts the emitted document's
// causal_chain and event_window are both non-empty for the seeded
// divergence -- the repair-agent requirement that a fix be proposable from
// this document alone, without opening the full execution log (D-05-26).
func TestMismatchDocumentIsSelfSufficient(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	document, err := session.ReduceSeededAliasMismatch(ctx)
	if err != nil {
		requirePhase16M004Refusal(t, err, "by-pointer")
		return
	}
	if len(document.CausalChain) == 0 {
		t.Fatal("expected a non-empty causal_chain")
	}
	if len(document.EventWindow) == 0 {
		t.Fatal("expected a non-empty event_window")
	}
	for engine, events := range document.EventWindow {
		if len(events) == 0 {
			t.Fatalf("engine %q has an empty event window", engine)
		}
		if len(events) > reduce.EventWindowSize {
			t.Fatalf("engine %q has %d events, want at most %d (bounded)", engine, len(events), reduce.EventWindowSize)
		}
	}
	if len(document.Causes) == 0 {
		t.Fatal("expected at least one cause naming what diverged")
	}
}

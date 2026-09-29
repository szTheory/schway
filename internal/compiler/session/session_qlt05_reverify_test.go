package session_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/reduce"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 11 plan 11-09 (D-11-32/D-11-37): QLT-05's strict, cold-start
// re-verification. Fixtures are built by hand, mirroring
// reduce_multifunction_test.go's own multiFunctionCallSeed shape --
// package reduce's own unexported moves and test fixtures are not visible
// from this package, so this file builds its own.
// ---------------------------------------------------------------------

const (
	qlt05CallerID = "s1:phase11.qlt05_fixture:function:caller"
	qlt05CalleeID = "s1:phase11.qlt05_fixture:function:callee"
)

// qlt05MultiFunctionSeed is a two-function seed: caller calls callee once
// (a single, eligible core.OpCall edge) and returns callee's result --
// exactly the shape Task 1 of 11-08 engineered as its own Q-01 spike
// shape. No Match, no ForeignContract: the only moves eligible against it
// are the two whole-program moves plus truncate-to-minimal-prefix, all of
// which -- per 11-08-SUMMARY.md's own TestOperationIDsUnchangedByWholeProgramMoves
// -- never renumber a SURVIVING operation's own ID. That is what makes
// this fixture's own entry-function terminal OpReturn ID a stable,
// content-derived fact across the whole reduction (used below as this
// file's own deterministic test double's OperationID).
func qlt05MultiFunctionSeed() core.Program {
	const module = "phase11.qlt05_fixture"
	typeID := qlt05CallerID + ":type:0"
	return core.Program{
		Schema:   core.Schema,
		Module:   module,
		ModuleID: "s1:" + module + ":module:" + module,
		Functions: []core.Function{
			{
				ID: qlt05CallerID, Name: "caller",
				EntryPointID: qlt05CallerID + ":point:entry", ReturnPointID: qlt05CallerID + ":point:return",
				Parameter:  core.Parameter{ID: qlt05CallerID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:    qlt05CallerID + ":linear",
					Types: []core.TypeFact{{ID: typeID}},
					Places: []core.Place{
						{ID: qlt05CallerID + ":place:0", Name: "value", TypeID: typeID},
						{ID: qlt05CallerID + ":place:1", Name: "result", TypeID: typeID},
					},
					Operations: []core.LinearOperation{
						{ID: qlt05CallerID + ":op:0", PointID: qlt05CallerID + ":point:linear:0", Kind: core.OpCall, SourceID: qlt05CallerID + ":place:0", TargetID: qlt05CallerID + ":place:1", TypeID: typeID, CalleeID: qlt05CalleeID},
						{ID: qlt05CallerID + ":op:1", PointID: qlt05CallerID + ":point:linear:1", Kind: core.OpReturn, SourceID: qlt05CallerID + ":place:1", TypeID: typeID},
					},
				},
			},
			{
				ID: qlt05CalleeID, Name: "callee",
				EntryPointID: qlt05CalleeID + ":point:entry", ReturnPointID: qlt05CalleeID + ":point:return",
				Parameter:  core.Parameter{ID: qlt05CalleeID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:         qlt05CalleeID + ":linear",
					Types:      []core.TypeFact{{ID: typeID}},
					Places:     []core.Place{{ID: qlt05CalleeID + ":place:0", Name: "value", TypeID: typeID}},
					Operations: []core.LinearOperation{{ID: qlt05CalleeID + ":op:0", PointID: qlt05CalleeID + ":point:linear:0", Kind: core.OpReturn, SourceID: qlt05CalleeID + ":place:0", TypeID: typeID}},
				},
			},
		},
	}
}

// alwaysInterestingPredicate accepts every candidate reduce.Reduce
// proposes, driving the seed to its own natural fixpoint deterministically
// -- this file's own tests care about QLT05Reverify's comparison logic,
// not about the reducer's own interestingness policy (covered elsewhere).
func alwaysInterestingPredicate(ctx context.Context, candidate core.Program) (reduce.Signature, bool, error) {
	return reduce.Signature{}, true, nil
}

// qlt05EntryReturnOperationID walks program for entryID's own function and
// returns its first OpReturn operation's ID -- the stable, content-derived
// fact qlt05StructuralDeriver keys its Signature on.
func qlt05EntryReturnOperationID(program core.Program, entryID string) (string, bool) {
	for _, fn := range program.Functions {
		if fn.ID != entryID || fn.Linear == nil {
			continue
		}
		for _, op := range fn.Linear.Operations {
			if op.Kind == core.OpReturn {
				return op.ID, true
			}
		}
	}
	return "", false
}

// qlt05StructuralDeriver is this file's own deterministic test double for
// session.QLT05SignatureDeriver: it computes a Signature purely as a
// function of the candidate program's own structure (entryID's terminal
// OpReturn operation ID), never compiling or running anything. This
// exercises QLT05Reverify's OWN comparison, cold-start, and idempotence
// properties in isolation -- properties orthogonal to HOW a fresh
// Signature is obtained -- while production wiring instead threads
// qlt05RealSignatureDeriver's real corevalidate+cgen+native pipeline (see
// session_phase5_mismatch.go's own doc comment on that function).
func qlt05StructuralDeriver(entryID string) session.QLT05SignatureDeriver {
	return func(ctx context.Context, candidate core.Program) (reduce.Signature, bool, error) {
		operationID, ok := qlt05EntryReturnOperationID(candidate, entryID)
		if !ok {
			return reduce.Signature{}, false, nil
		}
		return reduce.Signature{
			Axis:        "axis:terminal-outcome",
			EnginePair:  "interp-vs-O0",
			OperationID: operationID,
		}, true, nil
	}
}

// TestQLT05Reverification reduces a multi-function seed to its own
// fixpoint, re-derives a fresh Signature from the reduced program via a
// cold-start-shaped deriver, and asserts every strictly-compared field
// still agrees with the seed's own signature (derived from the SAME
// program before reduction) and that control:reduce.reverified is
// reported.
func TestQLT05Reverification(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	seedProgram := qlt05MultiFunctionSeed()
	deriver := qlt05StructuralDeriver(qlt05CallerID)
	seedSignature, ok, err := deriver(ctx, seedProgram)
	if err != nil || !ok {
		t.Fatalf("deriving seed signature: ok=%v err=%v", ok, err)
	}

	result, err := reduce.Reduce(ctx, reduce.Seed{Program: seedProgram, EntryFunctionID: qlt05CallerID}, alwaysInterestingPredicate)
	if err != nil {
		t.Fatalf("reduce.Reduce: %v", err)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("expected the whole-program moves to remove the now-uncalled callee, got %d functions", len(result.Program.Functions))
	}

	verdict, err := session.QLT05Reverify(ctx, seedSignature, result.Program, deriver, t.TempDir())
	if err != nil {
		t.Fatalf("QLT05Reverify: %v", err)
	}
	if !verdict.Verified {
		t.Fatalf("expected re-verification to pass, got Reason=%q", verdict.Reason)
	}
	if verdict.Control != session.ControlReduceReverified {
		t.Fatalf("Control = %q, want %q", verdict.Control, session.ControlReduceReverified)
	}
}

// TestQLT05ReverificationRejectsCausalRoleFallback constructs a reduced
// program whose OperationID differs from the seed's but whose CausalRole
// still matches, and asserts re-verification FAILS with
// reduce.reverification_signature_drift -- the assertion that converts
// HDD's slippage trap into a falsifiable claim.
func TestQLT05ReverificationRejectsCausalRoleFallback(t *testing.T) {
	ctx := context.Background()
	sharedCausalRole := "the acquisition whose release is observed last"
	seed := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interp-vs-O0", OperationID: "opA", CausalRole: sharedCausalRole}
	deriver := func(context.Context, core.Program) (reduce.Signature, bool, error) {
		return reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interp-vs-O0", OperationID: "opB", CausalRole: sharedCausalRole}, true, nil
	}

	verdict, err := session.QLT05Reverify(ctx, seed, core.Program{}, deriver, t.TempDir())
	if err != nil {
		t.Fatalf("QLT05Reverify: %v", err)
	}
	if verdict.Verified {
		t.Fatal("expected re-verification to FAIL when OperationID differs, even though CausalRole still matches")
	}
	if !strings.Contains(verdict.Reason, "reduce.reverification_signature_drift") {
		t.Fatalf("Reason = %q, want it to contain the exact literal %q", verdict.Reason, "reduce.reverification_signature_drift")
	}
}

// TestQLT05ReverificationIsColdStart runs re-verification with the cache
// root pointed at an empty t.TempDir() and asserts it still reaches a
// verdict -- proving no cached artifact or search-time verdict is
// consumed: if re-verification depended on any pre-populated content
// under cacheRoot, an empty temp directory would starve it.
func TestQLT05ReverificationIsColdStart(t *testing.T) {
	ctx := context.Background()
	seed := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interp-vs-O0", OperationID: "op"}
	deriver := func(context.Context, core.Program) (reduce.Signature, bool, error) {
		return reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "interp-vs-O0", OperationID: "op"}, true, nil
	}

	emptyCacheRoot := t.TempDir()
	entries, err := os.ReadDir(emptyCacheRoot)
	if err != nil {
		t.Fatalf("reading empty cache root: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected an empty cache root, found %d entries", len(entries))
	}

	verdict, err := session.QLT05Reverify(ctx, seed, core.Program{}, deriver, emptyCacheRoot)
	if err != nil {
		t.Fatalf("QLT05Reverify with an empty cache root: %v", err)
	}
	if !verdict.Verified {
		t.Fatalf("expected re-verification to still reach a passing verdict from an empty cache root, got Reason=%q", verdict.Reason)
	}
}

// TestQLT05ReverificationIsIdempotent re-verifies the SAME reduced program
// twice and asserts identical verdicts, and that the reduced program
// itself is never mutated by re-verification.
func TestQLT05ReverificationIsIdempotent(t *testing.T) {
	ctx := context.Background()
	seedProgram := qlt05MultiFunctionSeed()
	deriver := qlt05StructuralDeriver(qlt05CallerID)
	seedSignature, ok, err := deriver(ctx, seedProgram)
	if err != nil || !ok {
		t.Fatalf("deriving seed signature: ok=%v err=%v", ok, err)
	}

	result, err := reduce.Reduce(ctx, reduce.Seed{Program: seedProgram, EntryFunctionID: qlt05CallerID}, alwaysInterestingPredicate)
	if err != nil {
		t.Fatalf("reduce.Reduce: %v", err)
	}

	before, err := json.Marshal(result.Program)
	if err != nil {
		t.Fatalf("marshaling reduced program before re-verification: %v", err)
	}

	first, err := session.QLT05Reverify(ctx, seedSignature, result.Program, deriver, t.TempDir())
	if err != nil {
		t.Fatalf("QLT05Reverify (first): %v", err)
	}
	second, err := session.QLT05Reverify(ctx, seedSignature, result.Program, deriver, t.TempDir())
	if err != nil {
		t.Fatalf("QLT05Reverify (second): %v", err)
	}
	if first != second {
		t.Fatalf("expected identical verdicts across two re-verification runs, got %+v and %+v", first, second)
	}

	after, err := json.Marshal(result.Program)
	if err != nil {
		t.Fatalf("marshaling reduced program after re-verification: %v", err)
	}
	if string(before) != string(after) {
		t.Fatal("expected the reduced program to be unmodified by re-verification")
	}
}

// TestMismatchDocumentSchemaUnchanged asserts MismatchDocument's schema
// identifier is still lang.mismatch/0 and its field set is unchanged
// (D-05-39/D-11-37): QLT-05 is a gate claim about the reducer's own
// output, not a document field, so declaring control:reduce.reverified
// must never bump this schema.
func TestMismatchDocumentSchemaUnchanged(t *testing.T) {
	if reduce.MismatchSchema != "lang.mismatch/0" {
		t.Fatalf("MismatchSchema = %q, want %q", reduce.MismatchSchema, "lang.mismatch/0")
	}
	want := []string{
		"Schema", "DivergingAxis", "EnginePair", "DivergingOperationID",
		"ReducedCore", "ReducedSource", "Minimality", "TotalRecomputedWork",
		"ReductionAttempts", "EventWindow", "CausalChain", "Causes", "EvidenceID",
	}
	documentType := reflect.TypeOf(reduce.MismatchDocument{})
	if documentType.NumField() != len(want) {
		t.Fatalf("MismatchDocument has %d fields, want exactly %d: %v", documentType.NumField(), len(want), want)
	}
	for i, name := range want {
		if documentType.Field(i).Name != name {
			t.Fatalf("field %d = %q, want %q (schema field set must not change without a /1 bump)", i, documentType.Field(i).Name, name)
		}
	}
}

// ---------------------------------------------------------------------
// Task 3 (D-11-34): the anti-vacuity gate. control:reduce.no_progress was
// written against a single-function reducer and does not cover a
// multi-function reduction that drops zero calls and zero functions --
// this is the assertion that closes that gap.
// ---------------------------------------------------------------------

// qlt05GateFixturePath resolves testdata/phase11/multi_function_reduce_gate.schway.
const qlt05GateFixturePath = "testdata/phase11/multi_function_reduce_gate.schway"

// qlt05AppliedMovesContains reports whether result.AppliedMoves contains
// moveName -- the anti-vacuity assertion itself: a reduction that dropped
// zero calls and zero functions re-verifies perfectly and would go green
// without this check.
func qlt05AppliedMovesContains(result reduce.Result, moveName string) bool {
	for _, applied := range result.AppliedMoves {
		if applied == moveName {
			return true
		}
	}
	return false
}

// loadQLT05GateSeed checks and validates the gate fixture, resolving its
// entry function exactly as ReduceSeededAliasMismatch's own production
// call site does (callgraph.EntryFunction, never a guess).
func loadQLT05GateSeed(t *testing.T) reduce.Seed {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath(qlt05GateFixturePath))
	if err != nil {
		t.Fatalf("reading gate fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("gate fixture failed to check: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("gate fixture rejected by corevalidate: %+v", validated.Problems)
	}
	program := validated.Program()
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatalf("resolving entry function: %v", err)
	}
	return reduce.Seed{Program: program, EntryFunctionID: entry.ID}
}

// TestQLT05GateIsNonVacuous runs the reduction on the engineered gate
// fixture (helper is the provably removable function, per the fixture's
// own header) and asserts AppliedMoves actually contains
// "drop-orphan-function" -- the difference between a gate and a
// decoration.
func TestQLT05GateIsNonVacuous(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	seed := loadQLT05GateSeed(t)
	result, err := reduce.Reduce(ctx, seed, alwaysInterestingPredicate)
	if err != nil {
		t.Fatalf("reduce.Reduce: %v", err)
	}
	if !qlt05AppliedMovesContains(result, "drop-orphan-function") {
		t.Fatalf("AppliedMoves = %v, want it to contain %q -- the gate fixture's own removable function was never dropped", result.AppliedMoves, "drop-orphan-function")
	}
}

// TestQLT05EmptyReductionFailsAntiVacuity feeds a single-function seed on
// which NO whole-program move can ever apply (dropOrphanFunction refuses
// outright when len(Functions) <= 1, and there is no OpCall for
// dropCallSite to target) and asserts the anti-vacuity assertion FAILS --
// exercising the complementary side: asserting the gate passes when a
// move applied proves nothing about whether it would have noticed that
// none did.
func TestQLT05EmptyReductionFailsAntiVacuity(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const fn = "s1:phase11.qlt05_no_move_fixture:function:trivial"
	typeID := fn + ":type:0"
	trivial := core.Program{
		Schema:   core.Schema,
		Module:   "phase11.qlt05_no_move_fixture",
		ModuleID: "s1:phase11.qlt05_no_move_fixture:module:phase11.qlt05_no_move_fixture",
		Functions: []core.Function{
			{
				ID: fn, Name: "trivial",
				EntryPointID: fn + ":point:entry", ReturnPointID: fn + ":point:return",
				Parameter:  core.Parameter{ID: fn + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:         fn + ":linear",
					Types:      []core.TypeFact{{ID: typeID}},
					Places:     []core.Place{{ID: fn + ":place:0", Name: "value", TypeID: typeID}},
					Operations: []core.LinearOperation{{ID: fn + ":op:0", PointID: fn + ":point:linear:0", Kind: core.OpReturn, SourceID: fn + ":place:0", TypeID: typeID}},
				},
			},
		},
	}

	result, err := reduce.Reduce(ctx, reduce.Seed{Program: trivial, EntryFunctionID: fn}, alwaysInterestingPredicate)
	if err != nil {
		t.Fatalf("reduce.Reduce: %v", err)
	}
	if qlt05AppliedMovesContains(result, "drop-orphan-function") {
		t.Fatalf("AppliedMoves = %v unexpectedly contains %q on a seed with no eligible whole-program move", result.AppliedMoves, "drop-orphan-function")
	}
}

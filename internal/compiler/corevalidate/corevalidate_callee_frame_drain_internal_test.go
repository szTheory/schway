package corevalidate

import (
	"fmt"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/core"
)

// --- Plan 10-05 Task 2: peerCalleeFrameDrained (D-10-33/D-10-34, SEM-09) ---

// drainInternalByteType is the shared TypeFact every hand-built function
// below uses -- these tests drive peerCalleeFrameDrained directly (same
// package, no validator needed: it is a pure predicate), never through
// corevalidate.Validate's full pipeline, to isolate this ONE invariant from
// every other structural check a real fixture would also exercise.
var drainInternalByteType = core.TypeFact{ID: "drain:type:byte"}

// TestPeerCalleeFrameDrainedAdmitsReleasedAcquisition: an acquisition
// discharged by a matching OpRelease anywhere in the function is drained.
func TestPeerCalleeFrameDrainedAdmitsReleasedAcquisition(t *testing.T) {
	function := core.Function{
		ID: "fn:released", Parameter: core.Parameter{ID: "fn:released:place:0"},
		Linear: &core.LinearBody{
			Operations: []core.LinearOperation{
				{ID: "fn:released:op:0", Kind: core.OpForeignCall, TargetID: "fn:released:place:1", OkEdgeID: "e:ok", ErrEdgeID: "e:err"},
				{ID: "fn:released:op:1", Kind: core.OpRelease, SourceID: "fn:released:place:1", ReleasesOperationID: "fn:released:op:0"},
				{ID: "fn:released:op:2", Kind: core.OpReturn, SourceID: "fn:released:place:0"},
			},
			Edges: []core.Edge{
				{ID: "e:ok", ToBlockID: "block:ok"},
				{ID: "e:err", ToBlockID: "block:err"},
			},
		},
	}
	if !peerCalleeFrameDrained(&function) {
		t.Fatal("expected a released acquisition to be reported drained")
	}
}

// TestPeerCalleeFrameDrainedAdmitsReturnEscapedAcquisition:
// foreign_acquire_one.lang's own shape -- no release exists at all, but the
// acquired value is the function's own returned value directly.
func TestPeerCalleeFrameDrainedAdmitsReturnEscapedAcquisition(t *testing.T) {
	function := core.Function{
		ID: "fn:escaped", Parameter: core.Parameter{ID: "fn:escaped:place:0"},
		Linear: &core.LinearBody{
			Operations: []core.LinearOperation{
				{ID: "fn:escaped:op:0", Kind: core.OpForeignCall, TargetID: "fn:escaped:place:1", OkEdgeID: "e:ok", ErrEdgeID: "e:err"},
				{ID: "fn:escaped:op:1", Kind: core.OpReturn, SourceID: "fn:escaped:place:1"},
			},
			Edges: []core.Edge{
				{ID: "e:ok", ToBlockID: "block:ok"},
				{ID: "e:err", ToBlockID: "block:err"},
			},
		},
	}
	if !peerCalleeFrameDrained(&function) {
		t.Fatal("expected a return-escaped acquisition (no release, direct return) to be reported drained")
	}
}

// TestPeerCalleeFrameDrainedAdmitsReturnEscapeThroughMoveChain: the acquired
// value flows through a Move hop before it is returned -- still an escape,
// traced forward exactly like peerParameterEscapesOwned's own established
// shape.
func TestPeerCalleeFrameDrainedAdmitsReturnEscapeThroughMoveChain(t *testing.T) {
	function := core.Function{
		ID: "fn:moved_escape", Parameter: core.Parameter{ID: "fn:moved_escape:place:0"},
		Linear: &core.LinearBody{
			Operations: []core.LinearOperation{
				{ID: "fn:moved_escape:op:0", Kind: core.OpForeignCall, TargetID: "fn:moved_escape:place:1", OkEdgeID: "e:ok", ErrEdgeID: "e:err"},
				{ID: "fn:moved_escape:op:1", Kind: core.OpMove, SourceID: "fn:moved_escape:place:1", TargetID: "fn:moved_escape:place:2"},
				{ID: "fn:moved_escape:op:2", Kind: core.OpReturn, SourceID: "fn:moved_escape:place:2"},
			},
			Edges: []core.Edge{
				{ID: "e:ok", ToBlockID: "block:ok"},
				{ID: "e:err", ToBlockID: "block:err"},
			},
		},
	}
	if !peerCalleeFrameDrained(&function) {
		t.Fatal("expected an acquisition escaping through a Move chain to be reported drained")
	}
}

// TestPeerCalleeFrameDrainedAdmitsDiscardedAcquisition:
// discard_because.lang's own shape -- an acquisition whose OkEdgeID and
// ErrEdgeID converge on the SAME ToBlockID is check's own structural
// encoding of `discard ... because`, exempt from the drain obligation
// entirely: no release, no return escape, still admitted.
func TestPeerCalleeFrameDrainedAdmitsDiscardedAcquisition(t *testing.T) {
	function := core.Function{
		ID: "fn:discarded", Parameter: core.Parameter{ID: "fn:discarded:place:0"},
		Linear: &core.LinearBody{
			Operations: []core.LinearOperation{
				{ID: "fn:discarded:op:0", Kind: core.OpForeignCall, TargetID: "fn:discarded:place:1", OkEdgeID: "e:ok", ErrEdgeID: "e:err"},
				{ID: "fn:discarded:op:1", Kind: core.OpReturn, SourceID: "fn:discarded:place:0"},
			},
			Edges: []core.Edge{
				{ID: "e:ok", ToBlockID: "block:success"},
				{ID: "e:err", ToBlockID: "block:success"},
			},
		},
	}
	if !peerCalleeFrameDrained(&function) {
		t.Fatal("expected a discarded acquisition (converging ok/err edges) to be reported drained")
	}
}

// TestPeerCalleeFrameDrainedAdmitsNoLinearBody: a pure lang.core/0 match
// function (Linear == nil) trivially drains -- it acquires nothing.
func TestPeerCalleeFrameDrainedAdmitsNoLinearBody(t *testing.T) {
	function := core.Function{ID: "fn:match_only"}
	if !peerCalleeFrameDrained(&function) {
		t.Fatal("expected a function with no Linear body to be reported drained")
	}
}

// TestPeerCalleeFrameDrainedRefusesAbandonedAcquisition: a TRACKED
// acquisition (divergent ok/err edges, exactly the `try` shape) with no
// release anywhere and no return escape -- genuinely live in this frame at
// every one of its own terminating returns.
func TestPeerCalleeFrameDrainedRefusesAbandonedAcquisition(t *testing.T) {
	function := core.Function{
		ID: "fn:abandoned", Parameter: core.Parameter{ID: "fn:abandoned:place:0"},
		Linear: &core.LinearBody{
			Operations: []core.LinearOperation{
				{ID: "fn:abandoned:op:0", Kind: core.OpForeignCall, TargetID: "fn:abandoned:place:1", OkEdgeID: "e:ok", ErrEdgeID: "e:err"},
				// Returns the function's own PARAMETER, never the acquired
				// place -- exactly checkResourceLifecycle's own terminal
				// return shape (never the acquired resource itself).
				{ID: "fn:abandoned:op:1", Kind: core.OpReturn, SourceID: "fn:abandoned:place:0"},
			},
			Edges: []core.Edge{
				{ID: "e:ok", ToBlockID: "block:ok"},
				{ID: "e:err", ToBlockID: "block:err"},
			},
		},
	}
	if peerCalleeFrameDrained(&function) {
		t.Fatal("expected a genuinely abandoned tracked acquisition to be reported undrained")
	}
}

// See TestPeerCalleeFrameDrainedRefusesGenuinelyAbandonedAcquisition in
// corevalidate_test.go (the external test package) for the full
// corevalidate.Validate end-to-end proof, driven off a REAL checked fixture
// (session.Check) rather than a hand-built core.Program here: this
// package's own structural admission gate (operation ID/PointID ordinals,
// entry-block PointID identity, terminal-block reachability, and more)
// requires many more invariants to satisfy by hand than the bare predicate
// tests above need, and a real fixture already satisfies all of them.

// TestPeerCalleeFrameDrainedEvaluatedOncePerDeclaration is Task 2's own
// falsifier for the once-per-declaration claim, structurally mirroring
// check_test.go's own TestSummaryDerivationIsOnePassPerFunction: a diamond
// call graph (one caller, two relays, one shared leaf reached through BOTH
// relays) must evaluate peerCalleeFrameDrained exactly once per declared
// function -- including the shared leaf, reached by two call sites --
// instrumented via peerCalleeFrameDrainedObserved rather than merely
// inferred from the final Result.
func TestPeerCalleeFrameDrainedEvaluatedOncePerDeclaration(t *testing.T) {
	// typeIDFor mirrors linearStructural's own core.type_order convention
	// ({functionID}:type:{index}): every function declares its OWN Byte
	// type fact at index 0, never a fact shared across functions.
	typeIDFor := func(functionID string) string { return functionID + ":type:0" }
	// byteTypeFact derives Byte's REAL granted/negative abilities through
	// ability.Derive rather than hand-guessing them, mirroring
	// mustDeriveTypeFactForTest's identical precedent in the external test
	// package: a hardcoded ability list would silently drift from the real
	// derivation and trip core.ability_mismatch, a structural check
	// unrelated to what this test intends to exercise.
	byteTypeFact := func(functionID string) core.TypeFact {
		shape := core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}}
		derived, err := ability.Derive(shape)
		if err != nil {
			t.Fatalf("ability.Derive(Byte): %v", err)
		}
		return core.TypeFact{ID: typeIDFor(functionID), Shape: shape, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}
	}
	// op builds one flat-body operation with both the ID and PointID
	// ordinal conventions linearStructural's own core.operation_order check
	// requires ({functionID}:op:{index} / {functionID}:point:linear:{index}).
	op := func(functionID string, index int, kind core.OperationKind, sourceID, targetID, calleeID string) core.LinearOperation {
		return core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: kind, SourceID: sourceID, TargetID: targetID, TypeID: typeIDFor(functionID), CalleeID: calleeID,
		}
	}
	leafID := "diamond:fn:leaf"
	leaf := core.Function{
		ID: leafID, Name: "leaf", EntryPointID: leafID + ":point:entry", ReturnPointID: leafID + ":point:return",
		Parameter: core.Parameter{ID: leafID + ":place:0", Name: "v", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: leafID + ":linear", Types: []core.TypeFact{byteTypeFact(leafID)},
			Places: []core.Place{
				{ID: leafID + ":place:0", Name: "v", TypeID: typeIDFor(leafID)},
				{ID: leafID + ":place:1", Name: "out", TypeID: typeIDFor(leafID)},
			},
			Operations: []core.LinearOperation{
				op(leafID, 0, core.OpCopy, leafID+":place:0", leafID+":place:1", ""),
				op(leafID, 1, core.OpReturn, leafID+":place:1", "", ""),
			},
		},
	}
	relay := func(id, calleeID string) core.Function {
		return core.Function{
			ID: id, Name: id, EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
			Parameter: core.Parameter{ID: id + ":place:0", Name: "v", Type: "Byte"}, ReturnType: "Byte",
			Linear: &core.LinearBody{
				ID: id + ":linear", Types: []core.TypeFact{byteTypeFact(id)},
				Places: []core.Place{
					{ID: id + ":place:0", Name: "v", TypeID: typeIDFor(id)},
					{ID: id + ":place:1", Name: "out", TypeID: typeIDFor(id)},
				},
				Operations: []core.LinearOperation{
					op(id, 0, core.OpCall, id+":place:0", id+":place:1", calleeID),
					op(id, 1, core.OpReturn, id+":place:1", "", ""),
				},
			},
		}
	}
	relay1 := relay("diamond:fn:relay1", leaf.ID)
	relay2 := relay("diamond:fn:relay2", leaf.ID)
	callerID := "diamond:fn:caller"
	caller := core.Function{
		ID: callerID, Name: "caller", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
		Parameter: core.Parameter{ID: callerID + ":place:0", Name: "v", Type: "Byte"}, ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: callerID + ":linear", Types: []core.TypeFact{byteTypeFact(callerID)},
			Places: []core.Place{
				{ID: callerID + ":place:0", Name: "v", TypeID: typeIDFor(callerID)},
				{ID: callerID + ":place:1", Name: "r1", TypeID: typeIDFor(callerID)},
				{ID: callerID + ":place:2", Name: "r2", TypeID: typeIDFor(callerID)},
			},
			Operations: []core.LinearOperation{
				op(callerID, 0, core.OpCall, callerID+":place:0", callerID+":place:1", relay1.ID),
				op(callerID, 1, core.OpCopy, callerID+":place:1", callerID+":place:2", ""),
				op(callerID, 2, core.OpReturn, callerID+":place:2", "", ""),
			},
		},
	}
	// A second call site to relay2 too, so the leaf is reached through TWO
	// distinct call chains (caller->relay1->leaf and caller->relay2->leaf),
	// even though only relay1 is actually wired into caller's own body
	// above -- relay2 is declared (its OWN declaration must still be
	// evaluated exactly once) but not called by this program's entry path,
	// proving the invariant's count tracks DECLARATIONS, never reachability
	// or call-site count.
	program := core.Program{
		Schema: core.Schema1, Module: "diamond", ModuleID: "diamond:module",
		Functions: []core.Function{caller, relay1, relay2, leaf},
	}

	defer func() { peerCalleeFrameDrainedObserved = nil }()
	counts := map[string]int{}
	peerCalleeFrameDrainedObserved = func(functionID string) { counts[functionID]++ }

	result := Validate(program)
	if !result.Valid {
		t.Fatalf("expected the diamond call graph to validate, got %+v", result.Problems)
	}

	for _, function := range program.Functions {
		if counts[function.ID] != 1 {
			t.Fatalf("expected %q to be evaluated exactly once, got %d (counts: %+v)", function.ID, counts[function.ID], counts)
		}
	}
}

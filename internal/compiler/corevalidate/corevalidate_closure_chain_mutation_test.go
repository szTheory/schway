package corevalidate_test

import (
	"encoding/json"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// cloneCheckedProgramForTest deep-copies a core.Program via a JSON
// round-trip, so a test can mutate one function's declared facts (e.g.
// PublicOrigin) without disturbing the original.
func cloneCheckedProgramForTest(t *testing.T, program core.Program) core.Program {
	t.Helper()
	data, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("marshal for clone: %v", err)
	}
	var clone core.Program
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatalf("unmarshal for clone: %v", err)
	}
	return clone
}

// findFunctionIDByName returns the function ID matching name within
// program, failing the test if none is found.
func findFunctionIDByName(t *testing.T, program core.Program, name string) string {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function.ID
		}
	}
	t.Fatalf("expected program to declare a function named %q", name)
	return ""
}

// callBasicVariant returns testdata/phase07/call_basic.schway's checked
// program, optionally with `identity` (the callee `main` calls) carrying a
// declared ForeignContract it did not have before -- a signature-affecting
// change to the callee's OWN published facts, which corevalidate's peer
// derivation (derivePeerSignature) reads directly from
// function.ForeignContract, exactly like originvalidate.BuildInterface
// does. This deliberately mutates ForeignContract rather than PublicOrigin
// (Phase 07's original choice): identity's body is an ordinary owned
// passthrough with no borrow at all, so fabricating a PublicOrigin the body
// never derives is now (Phase 09, D-09-16) correctly refused by
// peerOriginContained -- exactly the false-agreement class this phase
// closes, not a shape this helper should exercise by accident. No
// core.OpForeignCall operation exists in identity's body, so none of
// corevalidate's ForeignContract field-shape checks (all gated on an
// actual OpForeignCall being present) fire for this synthetic contract; it
// exists purely to make Foreign (part of the peer's signature and
// ClosureDigest preimage) differ between variants.
func callBasicVariant(t *testing.T, calleeHasDeclaredForeign bool) (program core.Program, mainID, identityID string) {
	t.Helper()
	program = loadCheckedProgram(t, "phase07", "call_basic.schway")
	mainID = findFunctionIDByName(t, program, "main")
	identityID = findFunctionIDByName(t, program, "identity")
	if calleeHasDeclaredForeign {
		program = cloneCheckedProgramForTest(t, program)
		for i := range program.Functions {
			if program.Functions[i].ID == identityID {
				program.Functions[i].ForeignContract = &core.ForeignContract{Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden"}
			}
		}
	}
	return program, mainID, identityID
}

// peerClosureDigestFor validates program and returns caller's peer
// ClosureDigest, failing the test if validation did not succeed or the
// peer signature is missing.
func peerClosureDigestFor(t *testing.T, program core.Program, functionID string) string {
	t.Helper()
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("expected the program to validate, got problems: %+v", result.Problems)
	}
	signature, ok := result.PeerSignatures()[functionID]
	if !ok {
		t.Fatalf("function %s missing from peer signatures", functionID)
	}
	return signature.ClosureDigest
}

// TestPeerClosureDigestEmptyCalleesMutationKilled is 07-08 Task 2 Test 4's
// own peer-side kill (QLT-08, D-07-41): with
// corevalidate.SetClosureDigestEmptyCalleesForTest(true) engaged,
// chainPeerClosureDigests chains every function as if it had no callees --
// main's own peer ClosureDigest stops reacting to identity's (main's
// callee) signature changing. Restoring the seam restores the property.
func TestPeerClosureDigestEmptyCalleesMutationKilled(t *testing.T) {
	base, mainID, _ := callBasicVariant(t, false)
	mutatedCallee, _, _ := callBasicVariant(t, true)

	restore := corevalidate.SetClosureDigestEmptyCalleesForTest(true)
	baseDigestUnderFault := peerClosureDigestFor(t, base, mainID)
	mutatedDigestUnderFault := peerClosureDigestFor(t, mutatedCallee, mainID)
	restore()
	if baseDigestUnderFault != mutatedDigestUnderFault {
		t.Fatal("mutation (forcing empty callee pairs) had no observable effect: main's peer digest still changed when identity's signature changed")
	}

	baseDigestRestored := peerClosureDigestFor(t, base, mainID)
	mutatedDigestRestored := peerClosureDigestFor(t, mutatedCallee, mainID)
	if baseDigestRestored == mutatedDigestRestored {
		t.Fatal("expected the callee-changes-invalidates-caller property to be restored once the seam is disabled")
	}
}

// TestPeerClosureDigestDiscoveryOrderMutationKilled is 07-08 Task 2 Test
// 5's own peer-side kill: with
// corevalidate.SetClosureDigestDiscoveryOrderForTest(true) engaged,
// chainPeerClosureDigests chains in plain program.Functions declaration
// order instead of checkCallGraphAcyclic's own proven-correct postorder.
// call_basic.schway declares `identity` (the callee) BEFORE `main` (the
// caller), so this particular fixture cannot observe the mutation (the
// declaration order happens to already be callee-before-caller) -- the
// assertion instead directly compares the correctly- and
// incorrectly-ordered chained digest for main against each other on the
// SAME program, over a program where the two orders differ by
// construction: main's own preimage, chained with identity's
// STILL-EMPTY digest (declaration-order fault, since main precedes
// identity in the constructed slice below) must differ from main's
// preimage chained with identity's real digest.
func TestPeerClosureDigestDiscoveryOrderMutationKilled(t *testing.T) {
	program, mainID, identityID := callBasicVariant(t, false)
	// Reorder so main (the caller) is declared BEFORE identity (the
	// callee) -- call_basic.schway itself already declares identity first,
	// so this reorder is what makes the declaration-order fault
	// observable at all.
	reordered := cloneCheckedProgramForTest(t, program)
	var mainFunction, identityFunction core.Function
	for _, function := range reordered.Functions {
		switch function.ID {
		case mainID:
			mainFunction = function
		case identityID:
			identityFunction = function
		}
	}
	reordered.Functions = []core.Function{mainFunction, identityFunction}

	correctDigest := peerClosureDigestFor(t, reordered, mainID)

	restore := corevalidate.SetClosureDigestDiscoveryOrderForTest(true)
	defer restore()
	mutatedDigest := peerClosureDigestFor(t, reordered, mainID)

	if correctDigest == mutatedDigest {
		t.Fatal("mutation (declaration-order chaining) had no observable effect: main's peer digest was unchanged despite being chained before identity's digest existed")
	}
}

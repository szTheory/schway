package reduce_test

import (
	"context"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/reduce"
)

// coder is the Code()-carrying shape both callgraph's entry refusal and
// reduce's own seed refusal implement, asserted structurally here rather
// than by importing either concrete type.
type coder interface{ Code() string }

// TestReduceRefusesMultiFunctionSeedWithEmptyEntryID is WR-01's primary
// guard: the exact zero-value usage a caller gets by writing
// reduce.Seed{Program: p} and forgetting the entry fact. Before the guard
// this returned a nil error and a silently entry-less program.
func TestReduceRefusesMultiFunctionSeedWithEmptyEntryID(t *testing.T) {
	seed := multiFunctionCallSeed()
	_, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(reduce.Signature{}))
	if err == nil {
		t.Fatal("expected a multi-function seed with an empty EntryFunctionID to be refused (WR-01), got nil error")
	}
	refusal, ok := reduce.SeedEntryInvalidError(err)
	if !ok {
		t.Fatalf("expected reduce.SeedEntryInvalidError to type-assert the refusal, got %T: %v", err, err)
	}
	if got := refusal.Code(); got != core.SeedEntryInvalid {
		t.Fatalf("expected refusal code %q, got %q", core.SeedEntryInvalid, got)
	}
	if got := refusal.Supplied(); got != "" {
		t.Fatalf("expected the refusal to witness the empty supplied ID, got %q", got)
	}
	declared := refusal.Declared()
	if len(declared) != len(seed.Functions) {
		t.Fatalf("expected the refusal to witness all %d declared function IDs, got %d: %v", len(seed.Functions), len(declared), declared)
	}
}

// TestReduceRefusesMultiFunctionSeedWithUnknownEntryID covers the stale or
// misspelled ID half of WR-01 -- a caller that supplies an entry fact that
// its own program does not contain.
func TestReduceRefusesMultiFunctionSeedWithUnknownEntryID(t *testing.T) {
	seed := multiFunctionCallSeed()
	_, err := reduce.Reduce(
		context.Background(),
		reduce.Seed{Program: seed, EntryFunctionID: multiFnCallerID + ":typo"},
		alwaysInteresting(reduce.Signature{}),
	)
	if err == nil {
		t.Fatal("expected a multi-function seed with an unknown EntryFunctionID to be refused (WR-01), got nil error")
	}
	refusal, ok := reduce.SeedEntryInvalidError(err)
	if !ok {
		t.Fatalf("expected reduce.SeedEntryInvalidError to type-assert the refusal, got %T: %v", err, err)
	}
	if got := refusal.Supplied(); got != multiFnCallerID+":typo" {
		t.Fatalf("expected the refusal to witness the supplied ID verbatim, got %q", got)
	}
}

// TestReduceAcceptsSingleFunctionSeedWithoutEntryID pins Seed.Validate's
// deliberate exemption: dropOrphanFunction returns early on a one-function
// program, so the entry fact is unreachable and no hazard exists. Every
// pre-Phase-11 caller and every committed single-function golden depends on
// this staying permissive -- the guard must add a refusal exactly where the
// hazard is, and nowhere else.
func TestReduceAcceptsSingleFunctionSeedWithoutEntryID(t *testing.T) {
	seed := borrowChainSeed()
	if len(seed.Functions) != 1 {
		t.Fatalf("fixture precondition: expected a one-function seed, got %d", len(seed.Functions))
	}
	if _, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(reduce.Signature{})); err != nil {
		t.Fatalf("expected a single-function seed with no EntryFunctionID to remain accepted, got %v", err)
	}
}

// TestSeedEntryHazardIsReal is this guard's ANTI-VACUITY control: it proves
// the refusal protects against a real deletion, not a hypothetical one. It
// drives dropOrphanFunction directly with an empty entry ID -- exactly the
// package state Reduce would have been in before WR-01's fix -- against a
// program whose genuine entry is in-degree-zero, and asserts the entry IS
// deleted. If this ever stops deleting the entry, the guard above has
// become vacuous and this test fails, forcing the guard to be re-justified
// rather than silently kept as decoration.
func TestSeedEntryHazardIsReal(t *testing.T) {
	seed := multiFunctionCallSeed()

	// Precondition: the caller is the program's real entry and nothing
	// calls it, which is what makes it look exactly like an orphan.
	for _, fn := range seed.Functions {
		if fn.Linear == nil {
			continue
		}
		for _, op := range fn.Linear.Operations {
			if op.Kind == core.OpCall && op.CalleeID == multiFnCallerID {
				t.Fatalf("fixture precondition: %q must have zero in-edges to witness the hazard", multiFnCallerID)
			}
		}
	}

	reduced, applied := reduce.DropOrphanFunctionForTest(seed, "")
	if !applied {
		t.Fatal("expected drop-orphan-function to apply with an empty entry ID (the WR-01 hazard); it did not")
	}
	for _, fn := range reduced.Functions {
		if fn.ID == multiFnCallerID {
			t.Fatal("expected the unexempted entry function to be deleted, witnessing why Seed.Validate must refuse; it survived")
		}
	}

	// The same move with the entry ID correctly supplied must spare it --
	// the two halves together show the exemption is what does the work.
	spared, _ := reduce.DropOrphanFunctionForTest(seed, multiFnCallerID)
	found := false
	for _, fn := range spared.Functions {
		if fn.ID == multiFnCallerID {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a correctly supplied entry ID to exempt the entry function from deletion")
	}
}

// TestSeedEntryRefusalMessageIsByteStable pins the refusal's witness
// ordering: Declared() is sorted, so permuting the seed's own function
// declaration order cannot change the message text. This mirrors
// callgraph.entryAmbiguousError's identical byte-stability guarantee.
func TestSeedEntryRefusalMessageIsByteStable(t *testing.T) {
	forward := multiFunctionCallSeed()
	reversed := multiFunctionCallSeed()
	for i, j := 0, len(reversed.Functions)-1; i < j; i, j = i+1, j-1 {
		reversed.Functions[i], reversed.Functions[j] = reversed.Functions[j], reversed.Functions[i]
	}

	forwardErr := reduce.Seed{Program: forward}.Validate()
	reversedErr := reduce.Seed{Program: reversed}.Validate()
	if forwardErr == nil || reversedErr == nil {
		t.Fatal("expected both permutations to be refused")
	}
	if forwardErr.Error() != reversedErr.Error() {
		t.Fatalf("expected a byte-stable refusal message across declaration orders:\n forward: %s\nreversed: %s", forwardErr.Error(), reversedErr.Error())
	}
	if !strings.Contains(forwardErr.Error(), "reduce.seed_entry_invalid") {
		t.Fatalf("expected the refusal message to carry its own stable prefix, got %q", forwardErr.Error())
	}
}

// TestSeedEntryRefusalDispatchesLikeCallgraphRefusal asserts the shape
// claim the doc comments make: reduce's seed refusal and callgraph's entry
// refusal both answer Code(), so a caller can dispatch on either the same
// way without knowing the concrete type.
func TestSeedEntryRefusalDispatchesLikeCallgraphRefusal(t *testing.T) {
	err := reduce.Seed{Program: multiFunctionCallSeed()}.Validate()
	if err == nil {
		t.Fatal("expected a refusal")
	}
	c, ok := err.(coder)
	if !ok {
		t.Fatalf("expected the refusal to expose Code(), got %T", err)
	}
	if c.Code() != core.SeedEntryInvalid {
		t.Fatalf("expected Code() == %q, got %q", core.SeedEntryInvalid, c.Code())
	}
}

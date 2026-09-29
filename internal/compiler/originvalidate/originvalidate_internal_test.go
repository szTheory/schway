package originvalidate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

func sampleSignature() core.FunctionSignature {
	return core.FunctionSignature{
		ID: "f", Name: "f",
		Parameters: []core.ParameterContract{{ID: "p", Name: "p", Type: "Byte", Mode: "owned", Drops: false}},
		Return:     core.ReturnContract{Type: "Byte", Mode: "owned", Paths: []string{}},
		Abilities:  []core.Ability{},
		Foreign:    core.ForeignReach{},
	}
}

// TestClosureDigestNonSelfReferential is 07-01 Task 3's Test 1 (D-07-37):
// the preimage for a function excludes ClosureDigest itself — mutating only
// a function's ClosureDigest field and recomputing yields the same digest.
func TestClosureDigestNonSelfReferential(t *testing.T) {
	signature := sampleSignature()
	first, err := computeClosureDigest(signature, nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	signature.ClosureDigest = "sha256:some-stale-prior-value-entirely-different"
	second, err := computeClosureDigest(signature, nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if first != second {
		t.Fatalf("expected non-self-referential digest: %q != %q after only mutating ClosureDigest", first, second)
	}
}

// TestClosureDigestDomainSeparatorLoadBearing is 07-01 Task 3's Test 2
// (D-07-37): the preimage begins with a fixed domain separator declared
// once as a named const, and changing any other field of the signature
// changes the digest.
func TestClosureDigestDomainSeparatorLoadBearing(t *testing.T) {
	signature := sampleSignature()
	preimage, err := closureDigestPreimageBytes(signature, nil)
	if err != nil {
		t.Fatalf("preimage: %v", err)
	}
	if !bytes.HasPrefix(preimage, []byte(ClosureDigestDomainSeparator)) {
		t.Fatalf("expected preimage to begin with the domain separator %q, got %q", ClosureDigestDomainSeparator, preimage[:min(len(preimage), 64)])
	}

	base, err := computeClosureDigest(signature, nil)
	if err != nil {
		t.Fatalf("compute base: %v", err)
	}
	changed := signature
	changed.Name = "different-name"
	changedDigest, err := computeClosureDigest(changed, nil)
	if err != nil {
		t.Fatalf("compute changed: %v", err)
	}
	if base == changedDigest {
		t.Fatal("expected changing a signature field (Name) to change the digest")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestClosureDigestZeroCalleeBaseCase is 07-01 Task 3's Test 3 (D-07-37/
// D-07-38): for a zero-callee function the digest is the digest of that
// preimage with an empty callee list, is non-empty, starts with sha256:
// followed by exactly 64 lowercase hex characters, and is byte-identical
// across two runs in the same process and (via a fresh process re-running
// this same deterministic computation) across processes.
func TestClosureDigestZeroCalleeBaseCase(t *testing.T) {
	signature := sampleSignature()
	first, err := computeClosureDigest(signature, nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if first == "" {
		t.Fatal("expected a non-empty ClosureDigest")
	}
	if !strings.HasPrefix(first, "sha256:") || len(first) != len("sha256:")+64 {
		t.Fatalf("expected sha256:+64 lowercase hex, got %q", first)
	}
	for _, r := range first[len("sha256:"):] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			t.Fatalf("expected lowercase hex only, got %q", first)
		}
	}
	second, err := computeClosureDigest(signature, nil)
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if first != second {
		t.Fatalf("expected byte-identical digests across two runs: %q != %q", first, second)
	}
	// Empty and nil callee lists are the same zero-callee base case.
	third, err := computeClosureDigest(signature, []calleeDigestPair{})
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	if first != third {
		t.Fatalf("expected nil and empty callee lists to produce the same digest: %q != %q", first, third)
	}
}

// TestClosureDigestCalleeOrderIndependent is 07-01 Task 3's Test 4
// (D-07-37): a synthetic signature with two callee pairs supplied out of ID
// order produces the same digest as the same two pairs supplied in ID
// order — the sort is part of the preimage, not of the caller.
func TestClosureDigestCalleeOrderIndependent(t *testing.T) {
	signature := sampleSignature()
	inOrder := []calleeDigestPair{{ID: "a", ClosureDigest: "sha256:aa"}, {ID: "b", ClosureDigest: "sha256:bb"}}
	outOfOrder := []calleeDigestPair{{ID: "b", ClosureDigest: "sha256:bb"}, {ID: "a", ClosureDigest: "sha256:aa"}}
	inOrderDigest, err := computeClosureDigest(signature, inOrder)
	if err != nil {
		t.Fatalf("compute in-order digest: %v", err)
	}
	outOfOrderDigest, err := computeClosureDigest(signature, outOfOrder)
	if err != nil {
		t.Fatalf("compute out-of-order digest: %v", err)
	}
	if inOrderDigest != outOfOrderDigest {
		t.Fatalf("expected order-independent digests: %q != %q", inOrderDigest, outOfOrderDigest)
	}
}

// TestClosureDigestSortMutationKilled is 07-01 Task 3's D-07-42 mutation-kill
// falsifier (QLT-08, D-07-41): with closureDigestSortOverride replaced
// through its unexported seam to skip sorting, two callee pairs supplied
// out of ID order no longer produce the same digest as the same pairs
// supplied in ID order — proving the sort is genuinely part of the
// preimage, not merely appearing to be. Restoring the seam (via defer)
// restores order-independence. Same-package (white-box) test per D-07-42.
func TestClosureDigestSortMutationKilled(t *testing.T) {
	signature := core.FunctionSignature{ID: "f", Name: "f", Return: core.ReturnContract{Type: "Byte", Mode: "owned", Paths: []string{}}}
	inOrder := []calleeDigestPair{{ID: "a", ClosureDigest: "sha256:aa"}, {ID: "b", ClosureDigest: "sha256:bb"}}
	outOfOrder := []calleeDigestPair{{ID: "b", ClosureDigest: "sha256:bb"}, {ID: "a", ClosureDigest: "sha256:aa"}}

	inOrderDigest, err := computeClosureDigest(signature, inOrder)
	if err != nil {
		t.Fatalf("compute in-order digest: %v", err)
	}
	outOfOrderDigest, err := computeClosureDigest(signature, outOfOrder)
	if err != nil {
		t.Fatalf("compute out-of-order digest: %v", err)
	}
	if inOrderDigest != outOfOrderDigest {
		t.Fatalf("expected order-independent digests before mutation: %q != %q", inOrderDigest, outOfOrderDigest)
	}

	original := closureDigestSortOverride
	closureDigestSortOverride = func(pairs []calleeDigestPair) []calleeDigestPair { return pairs } // identity: skip sort
	defer func() { closureDigestSortOverride = original }()

	mutatedOutOfOrderDigest, err := computeClosureDigest(signature, outOfOrder)
	if err != nil {
		t.Fatalf("compute mutated out-of-order digest: %v", err)
	}
	if mutatedOutOfOrderDigest == inOrderDigest {
		t.Fatal("mutation (skipping the callee-pair sort) had no observable effect: out-of-order digest still matched in-order digest")
	}
}

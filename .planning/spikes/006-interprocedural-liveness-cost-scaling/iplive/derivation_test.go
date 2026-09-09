package iplive

import (
	"testing"
)

// TestForwardChainStaysLinearInProgramOrder pins the precondition the linear
// result rests on: a body whose operations are in PROGRAM ORDER converges in a
// constant number of derivation passes no matter how long its forwarding chain
// is.
func TestForwardChainStaysLinearInProgramOrder(t *testing.T) {
	var last float64
	for _, n := range []int{8, 32, 128, 512} {
		program, err := Generate(ShapeForward, n)
		if err != nil {
			t.Fatalf("generate: %v", err)
		}
		point, err := Measure(ShapeForward, program, 40000000)
		if err != nil {
			t.Fatalf("n=%d: %v", n, err)
		}
		last = float64(point.Memo.Work) / float64(point.Ops)
		t.Logf("ordered n=%4d ops=%5d memo=%8d perOp=%.1f", point.Functions, point.Ops, point.Memo.Work, last)
	}
	if last > 8 {
		t.Fatalf("program-order forwarding chain cost %.1f work/op at k=512; the constant-pass property is gone", last)
	}
}

// TestReversedBodyIsQuadratic is the other half of the same finding, and the
// reason it is stated as an INVARIANT rather than a property of the mechanism:
// list the same chain in reverse and the re-scan derivation needs one pass per
// link, which is quadratic in body length. Real core.LinearOperation sequences
// are in program order, so Phase 08 may rely on this -- but any later pass that
// reorders operations before summary derivation gives the bound back.
func TestReversedBodyIsQuadratic(t *testing.T) {
	first, last := 0.0, 0.0
	for _, k := range []int{8, 32, 128, 512} {
		program := ReversedForward(k)
		memo := NewMemoProvider(program, 0)
		analysis, err := Analyze(program, memo)
		if err != nil {
			t.Fatalf("k=%d: %v", k, err)
		}
		perOp := float64(analysis.Work) / float64(program.OpCount())
		if first == 0 {
			first = perOp
		}
		last = perOp
		t.Logf("reversed k=%4d ops=%5d memo=%8d perOp=%.1f", k, program.OpCount(), analysis.Work, perOp)
	}
	if last < first*10 {
		t.Fatalf("expected the reversed body to degrade (%.1f -> %.1f work/op); if it no longer does, the derivation changed and the README's invariant note must be revised", first, last)
	}
}

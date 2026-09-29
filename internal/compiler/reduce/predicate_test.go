package reduce_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/reduce"
)

func baseSignature() reduce.Signature {
	return reduce.Signature{
		Axis:                "axis:terminal-outcome",
		OperationID:         "s1:m:function:main:op:3",
		CausalRole:          "",
		EnginePair:          "interpreter-vs-O0",
		ForeignCallSequence: []string{"schway_res_open(request)", "schway_res_open(request)"},
	}
}

func TestPredicateAcceptsIdenticalSignature(t *testing.T) {
	seed := baseSignature()
	candidate := baseSignature()
	if !reduce.Interesting(seed, candidate) {
		t.Fatal("expected an identical signature to be accepted -- the predicate must not be vacuously rejecting everything")
	}
}

func TestPredicateRejectsDifferentAxis(t *testing.T) {
	seed := baseSignature()
	candidate := baseSignature()
	candidate.Axis = "axis:event-order"
	if reduce.Interesting(seed, candidate) {
		t.Fatal("expected a different axis to be rejected")
	}
}

func TestPredicateRejectsDifferentOperation(t *testing.T) {
	seed := baseSignature()
	candidate := baseSignature()
	candidate.OperationID = "s1:m:function:main:op:9"
	// no shared causal role either, so the fallback cannot rescue it
	if reduce.Interesting(seed, candidate) {
		t.Fatal("expected a different operation (with no matching causal role) to be rejected")
	}
}

func TestPredicateRejectsDifferentEnginePair(t *testing.T) {
	seed := baseSignature()
	candidate := baseSignature()
	candidate.EnginePair = "interpreter-vs-O3"
	if reduce.Interesting(seed, candidate) {
		t.Fatal("expected a different engine pair to be rejected")
	}
}

func TestPredicateAcceptsShiftedPositionSameCausalRole(t *testing.T) {
	seed := baseSignature()
	seed.CausalRole = "acquisition-released-last"
	candidate := baseSignature()
	candidate.CausalRole = "acquisition-released-last"
	// operation identity shifted under reduction (position moved), but the
	// causal role -- D-05-24's one permitted relaxation -- still matches.
	candidate.OperationID = "s1:m:function:main:op:1"
	if !reduce.Interesting(seed, candidate) {
		t.Fatal("expected a shifted-position candidate with a matching causal role to be accepted")
	}
}

func TestPredicateRejectsForeignCallSequenceDrift(t *testing.T) {
	seedSequence := []string{"schway_res_open(request)", "schway_arena_open(request)"}
	cases := map[string][]string{
		"reordered":              {"schway_arena_open(request)", "schway_res_open(request)"},
		"changed_symbol":         {"schway_res_open(request)", "schway_retained_touch(request)"},
		"changed_argument_shape": {"schway_res_open(other)", "schway_arena_open(request)"},
	}
	for name, sequence := range cases {
		t.Run(name, func(t *testing.T) {
			seed := baseSignature()
			seed.ForeignCallSequence = append([]string{}, seedSequence...)
			candidate := baseSignature()
			candidate.ForeignCallSequence = sequence
			if reduce.Interesting(seed, candidate) {
				t.Fatalf("expected foreign call sequence drift (%s) to be rejected", name)
			}
		})
	}
}

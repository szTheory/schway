package corevalidate

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// TestCheckCallTypeContractEmptyParameterTypeRefuses is 07-09 Task 2 Test 4's
// remaining fail-closed sub-case (edge 2): a callee whose declared
// Parameter.Type is the empty string refuses the argument-type predicate.
// This drives checkCallTypeContract directly, same-package, rather than
// through corevalidate.Validate's full pipeline: a structurally-valid
// core.Program can never carry an empty core.Function.Parameter.Type
// (linearStructural's own core.parameter_mismatch law requires it to equal
// the function's own type fact Shape.Constructor, which is never empty for
// a known constructor), so this shape is unreachable end to end -- exactly
// mirroring check.call_return_type_unrepresentable's own
// unreachable-from-source status (see PHASE-07-DEBT.md).
func TestCheckCallTypeContractEmptyParameterTypeRefuses(t *testing.T) {
	v := &validator{}
	source := core.Place{ID: "s1:test:fn:caller:place:0", TypeID: "s1:test:fn:caller:type:0"}
	target := core.Place{ID: "s1:test:fn:caller:place:1", TypeID: "s1:test:fn:caller:type:0"}
	types := map[string]core.TypeFact{
		"s1:test:fn:caller:type:0": {ID: "s1:test:fn:caller:type:0", Shape: core.TypeRef{Constructor: "Byte"}},
	}
	places := map[string]core.Place{target.ID: target}
	callee := core.Function{ID: "s1:test:fn:callee", Parameter: core.Parameter{Type: ""}, ReturnType: "Byte"}
	operation := core.LinearOperation{ID: "s1:test:fn:caller:op:0", TargetID: target.ID}

	if v.checkCallTypeContract(callee, source, operation, places, types) {
		t.Fatal("expected refusal for an empty declared parameter type, got admission")
	}
	if len(v.problems) == 0 || v.problems[0].Code != core.CallArgumentTypeMismatch {
		t.Fatalf("expected %s, got %+v", core.CallArgumentTypeMismatch, v.problems)
	}
}

// TestCheckCallTypeContractEmptyReturnTypeRefuses is the return-type
// counterpart: a callee whose declared ReturnType is the empty string
// refuses the target/return predicate, with the argument predicate
// clearing first (so this test proves the SECOND gate independently, not
// merely a repeat of the first).
func TestCheckCallTypeContractEmptyReturnTypeRefuses(t *testing.T) {
	v := &validator{}
	source := core.Place{ID: "s1:test:fn:caller:place:0", TypeID: "s1:test:fn:caller:type:0"}
	target := core.Place{ID: "s1:test:fn:caller:place:1", TypeID: "s1:test:fn:caller:type:0"}
	types := map[string]core.TypeFact{
		"s1:test:fn:caller:type:0": {ID: "s1:test:fn:caller:type:0", Shape: core.TypeRef{Constructor: "Byte"}},
	}
	places := map[string]core.Place{target.ID: target}
	callee := core.Function{ID: "s1:test:fn:callee", Parameter: core.Parameter{Type: "Byte"}, ReturnType: ""}
	operation := core.LinearOperation{ID: "s1:test:fn:caller:op:0", TargetID: target.ID}

	if v.checkCallTypeContract(callee, source, operation, places, types) {
		t.Fatal("expected refusal for an empty declared return type, got admission")
	}
	if len(v.problems) == 0 || v.problems[0].Code != core.CallReturnTypeMismatch {
		t.Fatalf("expected %s, got %+v", core.CallReturnTypeMismatch, v.problems)
	}
}

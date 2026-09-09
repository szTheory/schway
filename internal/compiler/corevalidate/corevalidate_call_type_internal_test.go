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

// TestCheckCallTypeContractArgumentPeerSeamKilled is 07-09 Task 3's
// per-control fail-closed mutation kill for
// control:call.argument_type_matches_parameter, isolated from the
// return-type predicate by driving checkCallTypeContract directly rather
// than through corevalidate.Validate's full pipeline: a genuinely
// mismatched argument (caller type Buffer, callee Parameter.Type Byte)
// with a MATCHING return type (callee ReturnType Buffer, same as the
// caller) is wrongly admitted when disableCallArgumentTypePeerForTest is
// engaged, and refused once it is disengaged. The matching return type is
// what isolates this control's own kill: with a mismatched return type
// too, the still-active return predicate would refuse regardless of this
// seam, proving nothing about THIS control specifically (exactly the
// coupling TestCallTypePeerMutationMatrix's own doc comment documents for
// corevalidate.Validate's full pipeline).
func TestCheckCallTypeContractArgumentPeerSeamKilled(t *testing.T) {
	defer func() { disableCallArgumentTypePeerForTest = false }()
	source := core.Place{ID: "s1:test:fn:caller:place:0", TypeID: "s1:test:fn:caller:type:0"}
	target := core.Place{ID: "s1:test:fn:caller:place:1", TypeID: "s1:test:fn:caller:type:0"}
	types := map[string]core.TypeFact{
		"s1:test:fn:caller:type:0": {ID: "s1:test:fn:caller:type:0", Shape: core.TypeRef{Constructor: "Buffer"}},
	}
	places := map[string]core.Place{target.ID: target}
	callee := core.Function{ID: "s1:test:fn:callee", Parameter: core.Parameter{Type: "Byte"}, ReturnType: "Buffer"}
	operation := core.LinearOperation{ID: "s1:test:fn:caller:op:0", TargetID: target.ID}

	disableCallArgumentTypePeerForTest = true
	v := &validator{}
	if !v.checkCallTypeContract(callee, source, operation, places, types) {
		t.Fatalf("expected the seam to wrongly admit the mismatched argument, got %+v", v.problems)
	}

	disableCallArgumentTypePeerForTest = false
	restored := &validator{}
	if restored.checkCallTypeContract(callee, source, operation, places, types) {
		t.Fatal("expected the restored predicate to refuse the mismatched argument")
	}
	if len(restored.problems) == 0 || restored.problems[0].Code != core.CallArgumentTypeMismatch {
		t.Fatalf("expected %s restored once the seam is disengaged, got %+v", core.CallArgumentTypeMismatch, restored.problems)
	}
}

// TestCheckCallTypeContractReturnPeerSeamKilled is the return-type
// counterpart, isolated the same way: a matching argument type (caller
// and callee Parameter.Type both Byte) with a mismatched return type
// (callee ReturnType Buffer) is wrongly admitted when
// disableCallReturnTypePeerForTest is engaged, and refused once it is
// disengaged.
func TestCheckCallTypeContractReturnPeerSeamKilled(t *testing.T) {
	defer func() { disableCallReturnTypePeerForTest = false }()
	source := core.Place{ID: "s1:test:fn:caller:place:0", TypeID: "s1:test:fn:caller:type:0"}
	target := core.Place{ID: "s1:test:fn:caller:place:1", TypeID: "s1:test:fn:caller:type:0"}
	types := map[string]core.TypeFact{
		"s1:test:fn:caller:type:0": {ID: "s1:test:fn:caller:type:0", Shape: core.TypeRef{Constructor: "Byte"}},
	}
	places := map[string]core.Place{target.ID: target}
	callee := core.Function{ID: "s1:test:fn:callee", Parameter: core.Parameter{Type: "Byte"}, ReturnType: "Buffer"}
	operation := core.LinearOperation{ID: "s1:test:fn:caller:op:0", TargetID: target.ID}

	disableCallReturnTypePeerForTest = true
	v := &validator{}
	if !v.checkCallTypeContract(callee, source, operation, places, types) {
		t.Fatalf("expected the seam to wrongly admit the mismatched return type, got %+v", v.problems)
	}

	disableCallReturnTypePeerForTest = false
	restored := &validator{}
	if restored.checkCallTypeContract(callee, source, operation, places, types) {
		t.Fatal("expected the restored predicate to refuse the mismatched return type")
	}
	if len(restored.problems) == 0 || restored.problems[0].Code != core.CallReturnTypeMismatch {
		t.Fatalf("expected %s restored once the seam is disengaged, got %+v", core.CallReturnTypeMismatch, restored.problems)
	}
}

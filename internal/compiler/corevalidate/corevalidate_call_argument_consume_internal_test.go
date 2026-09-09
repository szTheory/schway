package corevalidate

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// TestConsumeCallArgumentDerivesFromOwnShapeNotRecordedAbilities is 07-11
// Task 2 Test 4 (no shared implementation): consumeCallArgument's copy
// decision follows its OWN v.derive/deriveAbility re-derivation from
// types[operation.TypeID].Shape, never a producer-recorded
// TypeFact.Abilities list. This drives consumeCallArgument directly,
// same-package, rather than through corevalidate.Validate's full
// pipeline: a structurally-valid core.Program can never carry a
// TypeFact.Abilities list that disagrees with its own Shape (Validate's
// pre-existing core.ability_mismatch law, corevalidate.go, already
// enforces equality between the two), so a genuinely divergent input is
// unreachable end to end -- exactly checkCallTypeContract's own
// unreachable-from-a-valid-program precedent
// (corevalidate_call_type_internal_test.go). Byte's Shape structurally
// re-derives AbilityCopy regardless of what a hand-corrupted Abilities
// list says, proving the decision reads Shape, not the list.
func TestConsumeCallArgumentDerivesFromOwnShapeNotRecordedAbilities(t *testing.T) {
	v := &validator{}
	sourceID := "s1:test:fn:caller:place:0"
	typeID := "s1:test:fn:caller:type:0"
	types := map[string]core.TypeFact{
		typeID: {
			ID:    typeID,
			Shape: core.TypeRef{Constructor: "Byte"},
			// Deliberately built WITHOUT core.AbilityCopy, as if a producer
			// had (wrongly) recorded Byte as non-copyable. Shape.Constructor
			// stays "Byte", so v.derive must still grant AbilityCopy.
			Abilities: []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		},
	}
	initialized := map[string]bool{sourceID: true}
	operation := core.LinearOperation{ID: "s1:test:fn:caller:op:0", SourceID: sourceID, TypeID: typeID}

	if !v.consumeCallArgument(operation, types, initialized) {
		t.Fatalf("expected consumeCallArgument to succeed, got problems %+v", v.problems)
	}
	if !initialized[sourceID] {
		t.Fatal("expected a Byte argument to stay initialized (copied, not consumed) based on its OWN Shape-derived ability, ignoring the recorded Abilities list")
	}
}

// TestConsumeCallArgumentConsumesNonCopyableFromOwnShape is the
// non-copyable counterpart: a Buffer type fact whose Abilities list was
// (wrongly) recorded WITH core.AbilityCopy is still consumed, because
// Buffer's Shape structurally denies AbilityCopy regardless of the
// recorded list.
func TestConsumeCallArgumentConsumesNonCopyableFromOwnShape(t *testing.T) {
	v := &validator{}
	sourceID := "s1:test:fn:caller:place:0"
	typeID := "s1:test:fn:caller:type:0"
	types := map[string]core.TypeFact{
		typeID: {
			ID:        typeID,
			Shape:     core.TypeRef{Constructor: "Buffer"},
			Abilities: []core.Ability{core.AbilityCopy, core.AbilityShare},
		},
	}
	initialized := map[string]bool{sourceID: true}
	operation := core.LinearOperation{ID: "s1:test:fn:caller:op:0", SourceID: sourceID, TypeID: typeID}

	if !v.consumeCallArgument(operation, types, initialized) {
		t.Fatalf("expected consumeCallArgument to succeed, got problems %+v", v.problems)
	}
	if initialized[sourceID] {
		t.Fatal("expected a Buffer argument to be consumed based on its OWN Shape-derived ability, ignoring a wrongly-recorded Abilities list")
	}
}

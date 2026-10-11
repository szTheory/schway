package corevalidate

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ability"
	"github.com/szTheory/schway/internal/compiler/core"
)

var phase29Abilities = []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape}

// phase29AbilityDisagreements compares ability.Derive with the independent
// corevalidate.deriveAbility for every ability (D-17, D-23). A non-empty flip
// inverts corevalidate's granted bit for that one ability, which is the armed
// mismatch control: it proves the comparison can actually report a difference.
func phase29AbilityDisagreements(t *testing.T, shape core.TypeRef, flip core.Ability) []string {
	t.Helper()
	derived, err := ability.Derive(shape)
	if err != nil {
		t.Fatalf("ability.Derive(%v): %v", shape, err)
	}
	var disagreements []string
	for _, candidate := range phase29Abilities {
		granted := false
		for _, g := range derived.Granted {
			if g == candidate {
				granted = true
			}
		}
		var path []string
		for _, witness := range derived.NegativeWitnesses {
			if witness.Ability == candidate {
				path = witness.Path
			}
		}
		peerGranted, peerPath, known := deriveAbility(shape, candidate, 0, nil)
		if !known {
			disagreements = append(disagreements, fmt.Sprintf("%s: corevalidate does not know the shape", candidate))
			continue
		}
		if candidate == flip {
			peerGranted = !peerGranted
		}
		if granted != peerGranted || (!granted && !reflect.DeepEqual(append([]string(nil), path...), append([]string(nil), peerPath...))) {
			disagreements = append(disagreements, fmt.Sprintf("%s: ability.Derive granted=%v path=%v, corevalidate granted=%v path=%v", candidate, granted, path, peerGranted, peerPath))
		}
	}
	return disagreements
}

func TestPhase29ByteViewAbilityAgreement(t *testing.T) {
	shape := core.TypeRef{Constructor: "ByteView"}

	if got := phase29AbilityDisagreements(t, shape, ""); len(got) != 0 {
		t.Fatalf("ByteView derivations disagree: %v", got)
	}

	// D-17 asserted directly so two derivations cannot agree on a wrong answer.
	derived, err := ability.Derive(shape)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(derived.Granted, []core.Ability{core.AbilityDrop}) {
		t.Fatalf("ability.Derive(ByteView).Granted = %v, want [drop]", derived.Granted)
	}
	for _, denied := range []core.Ability{core.AbilityCopy, core.AbilityShare, core.AbilitySend, core.AbilityEscape} {
		granted, path, known := deriveAbility(shape, denied, 0, nil)
		if !known || granted || !reflect.DeepEqual(path, []string{"ByteView"}) {
			t.Fatalf("deriveAbility(ByteView, %s) = %v %v %v, want denied with path [ByteView]", denied, granted, path, known)
		}
		found := false
		for _, witness := range derived.NegativeWitnesses {
			if witness.Ability == denied {
				found = true
				if !reflect.DeepEqual(witness.Path, []string{"ByteView"}) {
					t.Fatalf("ability.Derive witness for %s = %v, want [ByteView]", denied, witness.Path)
				}
			}
		}
		if !found {
			t.Fatalf("ability.Derive has no negative witness for %s", denied)
		}
	}
	if granted, _, known := deriveAbility(shape, core.AbilityDrop, 0, nil); !known || !granted {
		t.Fatalf("deriveAbility(ByteView, drop) = %v known=%v, want granted", granted, known)
	}

	// Armed mismatch control: each single flipped ability is reported alone.
	for _, flipped := range phase29Abilities {
		got := phase29AbilityDisagreements(t, shape, flipped)
		if len(got) != 1 || len(got[0]) < len(flipped) || got[0][:len(flipped)] != string(flipped) {
			t.Fatalf("flip %s: disagreements = %v, want exactly one naming %s", flipped, got, flipped)
		}
	}
}

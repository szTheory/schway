package ability

import (
	"fmt"

	"github.com/codename-lang/lang/internal/compiler/core"
)

var order = [...]core.Ability{
	core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape,
}

const (
	maxAbilityDepth = 64
	maxAbilityNodes = 4096
)

type abilitySet struct {
	copy   bool
	drop   bool
	share  bool
	send   bool
	escape bool
}

func (set abilitySet) has(candidate core.Ability) bool {
	switch candidate {
	case core.AbilityCopy:
		return set.copy
	case core.AbilityDrop:
		return set.drop
	case core.AbilityShare:
		return set.share
	case core.AbilitySend:
		return set.send
	case core.AbilityEscape:
		return set.escape
	default:
		return false
	}
}

// combineStructural is the single production law for sealed aggregate shapes.
// It is deliberately package-private: callers may request derivation for a
// TypeRef, but cannot manufacture an ability set or install a global rule.
func combineStructural(children []abilitySet) abilitySet {
	result := abilitySet{copy: true, drop: true, share: true, send: true, escape: true}
	for _, child := range children {
		result.copy = result.copy && child.copy
		result.drop = result.drop && child.drop
		result.share = result.share && child.share
		result.send = result.send && child.send
		result.escape = result.escape && child.escape
	}
	return result
}

type derivedShape struct {
	set       abilitySet
	witnesses map[core.Ability][]string
}

type deriver struct {
	combine func([]abilitySet) abilitySet
	depth   int
	nodes   int
	// sealedLeaves names declared field-less nominal data types (Phase 3):
	// a constructor present here is treated as a sealed structural leaf
	// granting all five abilities, exactly like Byte, rather than falling
	// through to the unknown-constructor error. Nil for every Phase 1/2
	// caller, which only ever derives Byte/Buffer/Box/Pair shapes.
	sealedLeaves map[string]bool
}

type Result struct {
	Granted           []core.Ability
	NegativeWitnesses []core.AbilityWitness
}

// Derive applies the sealed primitive and constructor rules. There is no
// production API for supplying or combining arbitrary ability masks.
func Derive(shape core.TypeRef) (Result, error) {
	return DeriveSealed(shape, nil)
}

// DeriveSealed is Derive extended with a set of declared field-less nominal
// data type names (Phase 3). A match scrutinee's own type (e.g. an
// alternative-only data type with no payload) is not one of the built-in
// primitive/constructor shapes ability.go already understands, but its
// alternatives carry no data, so it is trivially copy/drop/share/send/
// escape-safe — exactly as sealed as Byte. sealedLeaves lets callers name
// which constructors qualify without ability.go importing the checker's
// declared-type table.
func DeriveSealed(shape core.TypeRef, sealedLeaves map[string]bool) (Result, error) {
	d := deriver{combine: combineStructural, sealedLeaves: sealedLeaves}
	derived, err := d.derive(shape)
	if err != nil {
		return Result{}, err
	}
	result := Result{Granted: []core.Ability{}, NegativeWitnesses: []core.AbilityWitness{}}
	for _, candidate := range order {
		if derived.set.has(candidate) {
			result.Granted = append(result.Granted, candidate)
		} else {
			result.NegativeWitnesses = append(result.NegativeWitnesses, core.AbilityWitness{
				Ability: candidate,
				Path:    append([]string(nil), derived.witnesses[candidate]...),
			})
		}
	}
	return result, nil
}

func (d *deriver) derive(shape core.TypeRef) (derivedShape, error) {
	return d.deriveAt(shape, 1)
}

func (d *deriver) deriveAt(shape core.TypeRef, depth int) (derivedShape, error) {
	d.depth = depth
	d.nodes++
	if depth > maxAbilityDepth || d.nodes > maxAbilityNodes {
		return derivedShape{}, fmt.Errorf("type expression exceeds ability limits")
	}

	if d.sealedLeaves[shape.Constructor] {
		if len(shape.Arguments) != 0 {
			return derivedShape{}, fmt.Errorf("%s takes no type arguments", shape.Constructor)
		}
		return derivedShape{
			set:       abilitySet{copy: true, drop: true, share: true, send: true, escape: true},
			witnesses: map[core.Ability][]string{},
		}, nil
	}

	switch shape.Constructor {
	case "Byte", "U64", "PathToken", "Unit":
		if len(shape.Arguments) != 0 {
			return derivedShape{}, fmt.Errorf("%s takes no type arguments", shape.Constructor)
		}
		return derivedShape{
			set:       abilitySet{copy: true, drop: true, share: true, send: true, escape: true},
			witnesses: map[core.Ability][]string{},
		}, nil
	case "Buffer":
		if len(shape.Arguments) != 0 {
			return derivedShape{}, fmt.Errorf("Buffer takes no type arguments")
		}
		// Buffer is a shareable, noncopyable resource: an immutable shared
		// loan observes the buffer without duplicating ownership, so `share`
		// is granted while `copy` remains denied with its witness intact.
		return derivedShape{
			set: abilitySet{drop: true, share: true, send: true, escape: true},
			witnesses: map[core.Ability][]string{
				core.AbilityCopy: {"Buffer"},
			},
		}, nil
	case "FileByteOwner":
		if len(shape.Arguments) != 0 {
			return derivedShape{}, fmt.Errorf("FileByteOwner takes no type arguments")
		}
		return derivedShape{
			set:       abilitySet{drop: true, share: true, send: true, escape: true},
			witnesses: map[core.Ability][]string{core.AbilityCopy: {"FileByteOwner"}},
		}, nil
	case "Box":
		if len(shape.Arguments) != 1 {
			return derivedShape{}, fmt.Errorf("Box takes one type argument")
		}
		return d.deriveStructural(shape.Arguments, []string{"Box.value"}, depth)
	case "Pair":
		if len(shape.Arguments) != 2 {
			return derivedShape{}, fmt.Errorf("Pair takes two type arguments")
		}
		return d.deriveStructural(shape.Arguments, []string{"Pair.left", "Pair.right"}, depth)
	default:
		return derivedShape{}, fmt.Errorf("unknown type constructor %q", shape.Constructor)
	}
}

func (d *deriver) deriveStructural(arguments []core.TypeRef, fields []string, depth int) (derivedShape, error) {
	children := make([]derivedShape, len(arguments))
	sets := make([]abilitySet, len(arguments))
	for index, argument := range arguments {
		child, err := d.deriveAt(argument, depth+1)
		if err != nil {
			return derivedShape{}, err
		}
		children[index] = child
		sets[index] = child.set
	}
	if d.combine == nil {
		return derivedShape{}, fmt.Errorf("structural ability combiner is not configured")
	}
	set := d.combine(sets)
	witnesses := make(map[core.Ability][]string)
	for _, candidate := range order {
		if set.has(candidate) {
			continue
		}
		for index, child := range children {
			if child.set.has(candidate) {
				continue
			}
			witnesses[candidate] = append([]string{fields[index]}, child.witnesses[candidate]...)
			break
		}
	}
	return derivedShape{set: set, witnesses: witnesses}, nil
}

func Has(result Result, candidate core.Ability) bool {
	for _, granted := range result.Granted {
		if granted == candidate {
			return true
		}
	}
	return false
}

package ability

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	compilerast "github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func fabricatedSet(mask int) abilitySet {
	return abilitySet{
		copy: mask&1 != 0, drop: mask&2 != 0, share: mask&4 != 0,
		send: mask&8 != 0, escape: mask&16 != 0,
	}
}

func observedMask(set abilitySet) int {
	mask := 0
	if set.copy {
		mask |= 1
	}
	if set.drop {
		mask |= 2
	}
	if set.share {
		mask |= 4
	}
	if set.send {
		mask |= 8
	}
	if set.escape {
		mask |= 16
	}
	return mask
}

func TestBoxAllAbilityMasks(t *testing.T) {
	for mask := 0; mask < 32; mask++ {
		input := fabricatedSet(mask)
		got := combineStructural([]abilitySet{input})
		if observedMask(got) != mask {
			t.Fatalf("Box mask %05b became %05b", mask, observedMask(got))
		}
		if input != fabricatedSet(mask) {
			t.Fatalf("combiner mutated Box input %05b", mask)
		}
	}
}

func TestPairAllAbilityMasks(t *testing.T) {
	for left := 0; left < 32; left++ {
		for right := 0; right < 32; right++ {
			got := observedMask(combineStructural([]abilitySet{fabricatedSet(left), fabricatedSet(right)}))
			want := 0
			for bit := 0; bit < 5; bit++ {
				leftHas := left&(1<<bit) != 0
				rightHas := right&(1<<bit) != 0
				if leftHas && rightHas {
					want |= 1 << bit
				}
			}
			if got != want {
				t.Fatalf("Pair masks %05b/%05b: got %05b want independent conjunction %05b", left, right, got, want)
			}
		}
	}
	firstChildLaw := true
	unionLaw := true
	forwardImplicationLaw := true
	reverseImplicationLaw := true
	for _, pair := range [][2]int{{1, 0}, {0, 1}, {5, 18}, {18, 5}} {
		got := observedMask(combineStructural([]abilitySet{fabricatedSet(pair[0]), fabricatedSet(pair[1])}))
		want := pair[0] & pair[1]
		if got != want {
			t.Fatalf("asymmetric conjunction mismatch: left=%05b right=%05b got=%05b want=%05b", pair[0], pair[1], got, want)
		}
		firstChildLaw = firstChildLaw && got == pair[0]
		unionLaw = unionLaw && got == pair[0]|pair[1]
		forwardImplicationLaw = forwardImplicationLaw && got == (^pair[0]|pair[1])&31
		reverseImplicationLaw = reverseImplicationLaw && got == (^pair[1]|pair[0])&31
	}
	if firstChildLaw || unionLaw || forwardImplicationLaw || reverseImplicationLaw {
		t.Fatalf("asymmetric cases did not distinguish structural conjunction: first=%t union=%t forward-implication=%t reverse-implication=%t", firstChildLaw, unionLaw, forwardImplicationLaw, reverseImplicationLaw)
	}
}

func TestTypeRefDerivationUsesStructuralCombiner(t *testing.T) {
	type call struct{ masks []int }
	var calls []call
	spy := func(children []abilitySet) abilitySet {
		masks := make([]int, len(children))
		for index, child := range children {
			masks[index] = observedMask(child)
		}
		calls = append(calls, call{masks: masks})
		return combineStructural(children)
	}
	d := deriver{combine: spy}
	shape := core.TypeRef{Constructor: "Pair", Arguments: []core.TypeRef{
		{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Buffer"}}},
		{Constructor: "Byte"},
	}}
	result, err := d.derive(shape)
	if err != nil {
		t.Fatal(err)
	}
	if result.set.copy || result.set.share || !result.set.drop || !result.set.send || !result.set.escape {
		t.Fatalf("sealed derivation returned wrong independent set: %+v", result.set)
	}
	if len(calls) != 2 || !reflect.DeepEqual(calls[0].masks, []int{26}) || !reflect.DeepEqual(calls[1].masks, []int{26, 31}) {
		t.Fatalf("Box/Pair path bypassed request-local combiner: %+v", calls)
	}
	if d.nodes != 4 {
		t.Fatalf("derivation counted %d nodes, want 4", d.nodes)
	}

	reverse, err := Derive(core.TypeRef{Constructor: "Pair", Arguments: []core.TypeRef{{Constructor: "Buffer"}, {Constructor: "Byte"}}})
	if err != nil || Has(reverse, core.AbilityCopy) || Has(reverse, core.AbilityShare) {
		t.Fatalf("reverse asymmetric Pair admitted first-child/union/implication behavior: result=%+v err=%v", reverse, err)
	}
}

func TestNegativeWitnessPath(t *testing.T) {
	shape := core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Pair", Arguments: []core.TypeRef{
		{Constructor: "Byte"},
		{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Buffer"}}},
	}}}}
	result, err := Derive(shape)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Box.value", "Pair.right", "Box.value", "Buffer"}
	for _, witness := range result.NegativeWitnesses {
		if witness.Ability == core.AbilityCopy || witness.Ability == core.AbilityShare {
			if !reflect.DeepEqual(witness.Path, want) {
				t.Fatalf("%s witness=%v want smallest stable path %v", witness.Ability, witness.Path, want)
			}
		}
	}
	if _, err := Derive(core.TypeRef{Constructor: "Unknown"}); err == nil {
		t.Fatal("unknown constructor was accepted")
	}
	tooDeep := core.TypeRef{Constructor: "Byte"}
	for index := 0; index < maxAbilityDepth+1; index++ {
		tooDeep = core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{tooDeep}}
	}
	if _, err := Derive(tooDeep); err == nil || !strings.Contains(err.Error(), "limits") {
		t.Fatalf("recursive-depth input was not rejected by a stable limit: %v", err)
	}
}

func TestArbitraryMasksRemainTestPrivate(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), testsupport.ProjectPath("internal", "compiler", "ability", "ability.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for name := range file.Scope.Objects {
		if goast.IsExported(name) && (strings.Contains(strings.ToLower(name), "mask") || strings.Contains(strings.ToLower(name), "combine") || name == "AbilitySet") {
			t.Fatalf("production ability package exported arbitrary structural authority %q", name)
		}
	}
	for _, surface := range []reflect.Type{
		reflect.TypeOf(core.TypeRef{}), reflect.TypeOf(core.TypeFact{}), reflect.TypeOf(core.Program{}),
		reflect.TypeOf(compilerast.TypeRef{}), reflect.TypeOf(compilerast.Program{}),
	} {
		for index := 0; index < surface.NumField(); index++ {
			name := strings.ToLower(surface.Field(index).Name)
			if strings.Contains(name, "mask") || strings.Contains(name, "root") {
				t.Fatalf("%s exposes fabricated ability authority through field %s", surface, surface.Field(index).Name)
			}
		}
	}
	for _, path := range [][]string{{"internal", "compiler", "check", "check.go"}, {"internal", "compiler", "core", "core.go"}, {"internal", "compiler", "ast", "ast.go"}} {
		source, err := parser.ParseFile(token.NewFileSet(), testsupport.ProjectPath(path...), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for name := range source.Scope.Objects {
			lower := strings.ToLower(name)
			if strings.Contains(lower, "mask") || strings.Contains(lower, "abilityroot") || strings.Contains(lower, "combinestructural") {
				t.Fatalf("%s accepts or exports arbitrary ability authority %q", strings.Join(path, "/"), name)
			}
		}
	}
}

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
	// Buffer denies only copy, so Box<Buffer> observes as 11110b = 30 and the
	// Pair conjunction with Byte (31) keeps every bit except copy.
	if result.set.copy || !result.set.share || !result.set.drop || !result.set.send || !result.set.escape {
		t.Fatalf("sealed derivation returned wrong independent set: %+v", result.set)
	}
	if len(calls) != 2 || !reflect.DeepEqual(calls[0].masks, []int{30}) || !reflect.DeepEqual(calls[1].masks, []int{30, 31}) {
		t.Fatalf("Box/Pair path bypassed request-local combiner: %+v", calls)
	}
	if d.nodes != 4 {
		t.Fatalf("derivation counted %d nodes, want 4", d.nodes)
	}

	reverse, err := Derive(core.TypeRef{Constructor: "Pair", Arguments: []core.TypeRef{{Constructor: "Buffer"}, {Constructor: "Byte"}}})
	if err != nil || Has(reverse, core.AbilityCopy) || !Has(reverse, core.AbilityShare) {
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
	sawCopy := false
	for _, witness := range result.NegativeWitnesses {
		if witness.Ability == core.AbilityCopy {
			sawCopy = true
			if !reflect.DeepEqual(witness.Path, want) {
				t.Fatalf("%s witness=%v want smallest stable path %v", witness.Ability, witness.Path, want)
			}
			continue
		}
		t.Fatalf("Buffer denies only copy, but %s was withheld with witness %v", witness.Ability, witness.Path)
	}
	if !sawCopy {
		t.Fatal("nested Buffer did not withhold copy")
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

// TestShareIsUniversallyGrantedAfterBufferShare records, as an executable
// fact, that no source-reachable type withholds share once Buffer grants it.
// The four constructors are enumerated exhaustively to depth 4: Byte and
// Buffer both grant share, and Box/Pair only conjoin their children. The
// checker's ownership.borrow_requires_share gate is therefore unreachable
// from source today and stands as defence in depth against the source-blind
// validator's core.ability.share_denied; if a future constructor withholds
// share, this test fails and the gate becomes reachable.
func TestShareIsUniversallyGrantedAfterBufferShare(t *testing.T) {
	shapes := []core.TypeRef{{Constructor: "Byte"}, {Constructor: "Buffer"}}
	for depth := 0; depth < 3; depth++ {
		next := append([]core.TypeRef(nil), shapes...)
		for _, left := range shapes {
			next = append(next, core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{left}})
			for _, right := range shapes {
				next = append(next, core.TypeRef{Constructor: "Pair", Arguments: []core.TypeRef{left, right}})
			}
		}
		shapes = next
	}
	for _, shape := range shapes {
		result, err := Derive(shape)
		if err != nil {
			t.Fatalf("%+v: %v", shape, err)
		}
		if !Has(result, core.AbilityShare) {
			t.Fatalf("%+v withholds share: the borrow gate is now source-reachable and needs a negative fixture", shape)
		}
	}
	if len(shapes) < 100 {
		t.Fatalf("enumeration collapsed to %d shapes", len(shapes))
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

// TestNominalLeafAbilityMaskIsExhaustive is the falsifiable claim that
// alternative count and naming cannot affect the derived mask for a
// declared field-less nominal data type (Phase 3's sealed leaf, D-10). The
// existing TestBoxAllAbilityMasks/TestPairAllAbilityMasks enumerate
// structural COMBINATORS; a nominal leaf has no combinator and, before this
// test, had no input-varying coverage at all — the claim "any declared
// field-less type grants all five abilities regardless of shape" was
// assumed, not exercised.
func TestNominalLeafAbilityMaskIsExhaustive(t *testing.T) {
	names := []string{"Switch", "Signal", "X", "VeryLongAlternativeTypeName", "A1", "switch_"}
	alternativeCounts := []int{1, 2, 3, 5, 16}
	for _, name := range names {
		for _, count := range alternativeCounts {
			sealed := map[string]bool{name: true}
			result, err := DeriveSealed(core.TypeRef{Constructor: name}, sealed)
			if err != nil {
				t.Fatalf("name=%q alternatives=%d: unexpected error: %v", name, count, err)
			}
			for _, wanted := range []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape} {
				if !Has(result, wanted) {
					t.Fatalf("name=%q alternatives=%d: sealed leaf denied %s, want all five granted: %+v", name, count, wanted, result)
				}
			}
			if len(result.NegativeWitnesses) != 0 {
				t.Fatalf("name=%q alternatives=%d: sealed leaf carries witnesses: %+v", name, count, result.NegativeWitnesses)
			}
		}
	}
	// A constructor NOT in the sealed set stays unknown — the leaf
	// treatment is opt-in per name, never a blanket fallback.
	if _, err := DeriveSealed(core.TypeRef{Constructor: "Unsealed"}, map[string]bool{"Switch": true}); err == nil {
		t.Fatal("an unsealed, unknown constructor was silently accepted")
	}
	// A sealed name with type arguments is rejected: a field-less nominal
	// type takes none, exactly like Byte.
	if _, err := DeriveSealed(core.TypeRef{Constructor: "Switch", Arguments: []core.TypeRef{{Constructor: "Byte"}}}, map[string]bool{"Switch": true}); err == nil {
		t.Fatal("a sealed leaf with type arguments was silently accepted")
	}
}

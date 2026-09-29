package check_test

import (
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ability"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestUnexecutableShapeRejectedWithSpan closes D-02-09/D-07 (03-02-03):
// testdata/phase2/ability_shapes.schway type-checked cleanly through Phase 2
// but every engine died spanless at exit 3, because Box/Pair have no native
// C lowering (cgen.linearInput only maps Byte/Buffer). check now refuses
// each such function with a span-bearing, repair-bearing diagnostic naming
// the offending constructor, instead of admitting it into the checked core.
func TestUnexecutableShapeRejectedWithSpan(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase2", "ability_shapes.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 3 {
		t.Fatalf("want 3 diagnostics (one per unexecutable-shape function), got %+v", checked.Diagnostics)
	}
	for _, problem := range checked.Diagnostics {
		if problem.Code != "check.unexecutable_shape" {
			t.Fatalf("unexpected code %q: %+v", problem.Code, problem)
		}
		if problem.Schema != "lang.diagnostic/1" {
			t.Fatalf("unexecutable-shape rejection must join the repair-bearing taxonomy: %+v", problem)
		}
		if problem.Primary.Start == 0 && problem.Primary.End == 0 {
			t.Fatalf("unexecutable-shape diagnostic carries no span: %+v", problem)
		}
		if len(problem.Repairs) != 1 || problem.Repairs[0].Kind != "use_executable_shape" {
			t.Fatalf("unexecutable-shape diagnostic missing its executable-shapes repair: %+v", problem)
		}
		hasType, hasConstructor := false, false
		for _, cause := range problem.Causes {
			if cause.Kind == "type" && cause.Detail != "" {
				hasType = true
			}
			if cause.Kind == "constructor" && (cause.Detail == "Box" || cause.Detail == "Pair") {
				hasConstructor = true
			}
		}
		if !hasType || !hasConstructor {
			t.Fatalf("unexecutable-shape diagnostic missing type/constructor causes: %+v", problem)
		}
	}
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("a rejected function must not be admitted for execution: %+v", checked.Program.Functions)
	}
}

// TestAbilityFactsSurviveExecutionRejection proves check.go's new execution-
// admission gate (D-02-09/D-07) is about REFUSING EXECUTION, not about
// withholding ability derivation: the exact three shapes
// testdata/phase2/ability_shapes.schway exercises still derive their sealed
// structural ability facts and negative witnesses cleanly through the
// underlying ability package, unaffected by checkLinear's new gate — the
// gate runs strictly after ability.Derive already succeeded (check.go's
// checkLinear).
func TestAbilityFactsSurviveExecutionRejection(t *testing.T) {
	tests := []struct {
		name      string
		shape     core.TypeRef
		granted   []core.Ability
		witnesses []core.AbilityWitness
	}{
		{
			name:      "Box<Byte>",
			shape:     core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Byte", Arguments: []core.TypeRef{}}}},
			granted:   []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			witnesses: []core.AbilityWitness{},
		},
		{
			name:      "Box<Buffer>",
			shape:     core.TypeRef{Constructor: "Box", Arguments: []core.TypeRef{{Constructor: "Buffer", Arguments: []core.TypeRef{}}}},
			granted:   []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			witnesses: []core.AbilityWitness{{Ability: core.AbilityCopy, Path: []string{"Box.value", "Buffer"}}},
		},
		{
			name: "Pair<Byte,Buffer>",
			shape: core.TypeRef{Constructor: "Pair", Arguments: []core.TypeRef{
				{Constructor: "Byte", Arguments: []core.TypeRef{}},
				{Constructor: "Buffer", Arguments: []core.TypeRef{}},
			}},
			granted:   []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			witnesses: []core.AbilityWitness{{Ability: core.AbilityCopy, Path: []string{"Pair.right", "Buffer"}}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ability.Derive(test.shape)
			if err != nil {
				t.Fatalf("ability derivation failed for an unexecutable-but-well-typed shape: %v", err)
			}
			if !reflect.DeepEqual(result.Granted, test.granted) || !reflect.DeepEqual(result.NegativeWitnesses, test.witnesses) {
				t.Fatalf("%s ability facts diverged: got granted=%+v witnesses=%+v, want granted=%+v witnesses=%+v",
					test.name, result.Granted, result.NegativeWitnesses, test.granted, test.witnesses)
			}
		})
	}
}

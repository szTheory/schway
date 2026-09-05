package corevalidate_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestOwnershipMutationMatrix(t *testing.T) {
	valid := ownedProgram()
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid owned core rejected: %+v", result)
	}

	tests := []struct {
		name string
		code string
		edit func(*core.Program)
	}{
		{"forged copy operation", "core.ability.copy_denied", func(program *core.Program) {
			program.Functions[0].Linear.Operations[0].Kind = core.OpCopy
		}},
		{"forged positive ability", "core.ability_mismatch", func(program *core.Program) {
			program.Functions[0].Linear.Types[0].Abilities = append([]core.Ability{core.AbilityCopy}, program.Functions[0].Linear.Types[0].Abilities...)
		}},
		{"duplicate operation id", "core.duplicate_operation_id", func(program *core.Program) {
			program.Functions[0].Linear.Operations[1].ID = program.Functions[0].Linear.Operations[0].ID
		}},
		{"unknown place", "core.unknown_place", func(program *core.Program) {
			program.Functions[0].Linear.Operations[0].SourceID = "s1:test:fn:relay:place:absent"
		}},
		{"unknown type", "core.unknown_type", func(program *core.Program) {
			program.Functions[0].Linear.Places[0].TypeID = "s1:test:fn:relay:type:absent"
		}},
		{"unknown loan", "core.unknown_loan", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[0].LoanID = ""
		}},
		{"omitted move", "core.operation_order", func(program *core.Program) {
			operations := program.Functions[0].Linear.Operations
			program.Functions[0].Linear.Operations = append([]core.LinearOperation(nil), operations[1:]...)
		}},
		{"illegal loan transition", "core.move_while_borrowed", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			operations := program.Functions[0].Linear.Operations
			firstID, firstPoint := operations[1].ID, operations[1].PointID
			secondID, secondPoint := operations[2].ID, operations[2].PointID
			operations[1], operations[2] = operations[2], operations[1]
			operations[1].ID, operations[1].PointID = firstID, firstPoint
			operations[2].ID, operations[2].PointID = secondID, secondPoint
		}},
		{"inconsistent final claim", "core.final_claim_mismatch", func(program *core.Program) {
			program.Functions[0].Linear.Operations[1].SourceID = program.Functions[0].Parameter.ID
		}},
		{"self move target", "core.invalid_target", func(program *core.Program) {
			program.Functions[0].Linear.Operations[0].TargetID = program.Functions[0].Linear.Operations[0].SourceID
		}},
		{"target overwrite", "core.invalid_target", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[1].TargetID = program.Functions[0].Linear.Operations[0].TargetID
		}},
		{"skipped target ordinal", "core.invalid_target", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[0].TargetID = program.Functions[0].Linear.Places[2].ID
		}},
		{"reused borrow target", "core.invalid_target", func(program *core.Program) {
			borrow := borrowedProgram()
			*program = borrow
			program.Functions[0].Linear.Operations[1].Kind = core.OpBorrowShared
			program.Functions[0].Linear.Operations[1].LoanID = program.Functions[0].ID + ":loan:1"
			program.Functions[0].Linear.Operations[1].TargetID = program.Functions[0].Linear.Operations[0].TargetID
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneProgram(t, valid)
			test.edit(&mutated)
			result := corevalidate.Validate(mutated)
			if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != test.code {
				t.Fatalf("mutation result=%+v, want first code %q", result, test.code)
			}
		})
	}
}

func TestArbitraryMaskCannotEnterCoreValidation(t *testing.T) {
	encoded, err := json.Marshal(ownedProgram())
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"mask", "arbitrary", "bits"} {
		if strings.Contains(strings.ToLower(string(encoded)), forbidden) {
			t.Fatalf("serialized core contains arbitrary ability representation %q: %s", forbidden, encoded)
		}
	}
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"compiler/ast", "compiler/check", "compiler/interp", "compiler/cgen", "compiler/session", "compiler/ability"} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("validator imports forbidden producer/engine package %q", forbidden)
		}
	}
}

func TestValidationOwnsTrustBoundaryCopy(t *testing.T) {
	program := ownedProgram()
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("valid core rejected: %+v", result)
	}
	program.Functions[0].Linear.Types[0].Shape.Constructor = "Forged"
	program.Functions[0].Linear.Operations[0].Kind = core.OpCopy
	if got := result.Program(); !reflect.DeepEqual(got, ownedProgram()) {
		t.Fatalf("validated core aliases caller storage:\ngot=%+v", got)
	}
	first := result.Program()
	first.Functions[0].Linear.Places[0].Name = "mutated"
	if second := result.Program(); second.Functions[0].Linear.Places[0].Name == "mutated" {
		t.Fatal("Program returned mutable validator-owned storage")
	}
}

func TestUnvalidatedCoreCannotExecute(t *testing.T) {
	forged := ownedProgram()
	forged.Functions[0].Linear.Types[0].Abilities = append([]core.Ability{core.AbilityCopy}, forged.Functions[0].Linear.Types[0].Abilities...)
	if _, err := interp.Run(forged, "relay", "01020304"); err == nil || !strings.Contains(err.Error(), "core.ability_mismatch") {
		t.Fatalf("unvalidated forged core reached interpreter: %v", err)
	}
}

func TestCoreValidationWorkSeries(t *testing.T) {
	for _, facts := range []int{101, 1001, 10001} {
		program := scaleProgram(facts)
		result := corevalidate.Validate(program)
		if !result.Valid {
			t.Fatalf("facts=%d rejected: %+v", facts, result)
		}
		want := corevalidate.LinearWorkLimit(facts)
		if result.Checks != want {
			t.Fatalf("facts=%d checks=%d want exact linear formula %d", facts, result.Checks, want)
		}
	}
}

func TestOwnedClaimReorderRejected(t *testing.T) {
	program := ownedProgram()
	before, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	operations := program.Functions[0].Linear.Operations
	operations[0], operations[1] = operations[1], operations[0]
	after, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	if sha256.Sum256(before) == sha256.Sum256(after) {
		t.Fatal("semantic operation/final-claim reorder did not change identity")
	}
	result := corevalidate.Validate(program)
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.operation_order" {
		t.Fatalf("reordered final claim survived: %+v", result)
	}
}

func TestCoordinatedSourceCoreEscapeIsNamed(t *testing.T) {
	falseClaim := ownedProgram()
	function := &falseClaim.Functions[0]
	function.Parameter.Type = "Byte"
	function.ReturnType = "Byte"
	fact := &function.Linear.Types[0]
	fact.Shape = core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}}
	fact.Abilities = []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape}
	fact.NegativeWitnesses = []core.AbilityWitness{}
	function.Linear.Operations[0].Kind = core.OpCopy

	result := corevalidate.Validate(falseClaim)
	if !result.Valid {
		t.Fatalf("internally consistent coordinated false claim was incorrectly reported as detected: %+v", result)
	}
	if corevalidate.KnownEscape != "escape:coordinated-source-core-lie" {
		t.Fatalf("unexpected proof-boundary name %q", corevalidate.KnownEscape)
	}
}

func ownedProgram() core.Program {
	functionID := "s1:test:fn:relay"
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	targetID := functionID + ":place:1"
	return core.Program{
		Schema: core.Schema1, Module: "test", ModuleID: "s1:test:module:test",
		Functions: []core.Function{{
			ID: functionID, Name: "relay", EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter: core.Parameter{ID: parameterID, Name: "buffer", Type: "Buffer"}, ReturnType: "Buffer",
			Linear: &core.LinearBody{
				ID: functionID + ":linear",
				Types: []core.TypeFact{{
					ID: typeID, Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
					Abilities:         []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
					NegativeWitnesses: []core.AbilityWitness{{Ability: core.AbilityCopy, Path: []string{"Buffer"}}},
				}},
				Places: []core.Place{{ID: parameterID, Name: "buffer", TypeID: typeID}, {ID: targetID, Name: "delivered", TypeID: typeID}},
				Operations: []core.LinearOperation{
					{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpMove, SourceID: parameterID, TargetID: targetID, TypeID: typeID},
					{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: targetID, TypeID: typeID},
				},
			},
		}},
	}
}

func borrowedProgram() core.Program {
	program := ownedProgram()
	function := &program.Functions[0]
	linear := function.Linear
	function.Parameter.Type = "Byte"
	function.ReturnType = "Byte"
	linear.Types[0] = core.TypeFact{
		ID: linear.Types[0].ID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
	owner := function.Parameter.ID
	view := function.ID + ":place:1"
	observed := function.ID + ":place:2"
	delivered := function.ID + ":place:3"
	linear.Places = []core.Place{
		{ID: owner, Name: "buffer", TypeID: linear.Types[0].ID},
		{ID: view, Name: "view", TypeID: linear.Types[0].ID},
		{ID: observed, Name: "observed", TypeID: linear.Types[0].ID},
		{ID: delivered, Name: "delivered", TypeID: linear.Types[0].ID},
	}
	linear.Operations = []core.LinearOperation{
		{ID: function.ID + ":op:0", PointID: function.ID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: owner, TargetID: view, LoanID: function.ID + ":loan:0", TypeID: linear.Types[0].ID},
		{ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1", Kind: core.OpCopy, SourceID: view, TargetID: observed, TypeID: linear.Types[0].ID},
		{ID: function.ID + ":op:2", PointID: function.ID + ":point:linear:2", Kind: core.OpMove, SourceID: owner, TargetID: delivered, TypeID: linear.Types[0].ID},
		{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpReturn, SourceID: delivered, TypeID: linear.Types[0].ID},
	}
	return program
}

// reborrowProgram derives a second place from the loan target (a reborrow when
// derivation is OpBorrowShared, a copy of the loan when it is OpCopy) and then
// observes that derived place after the owner is moved. Loan liveness must be
// transitive for this to be refused: associating loan:0 only with its
// immediate target expires it at op:1 and admits the move at op:2.
func reborrowProgram(derivation core.OperationKind, observeAfterMove bool) core.Program {
	program := borrowedProgram()
	function := &program.Functions[0]
	linear := function.Linear
	owner := function.Parameter.ID
	view := function.ID + ":place:1"
	review := function.ID + ":place:2"
	delivered := function.ID + ":place:3"
	observed := function.ID + ":place:4"
	typeID := linear.Types[0].ID
	linear.Places = []core.Place{
		{ID: owner, Name: "buffer", TypeID: typeID},
		{ID: view, Name: "view", TypeID: typeID},
		{ID: review, Name: "review", TypeID: typeID},
		{ID: delivered, Name: "delivered", TypeID: typeID},
		{ID: observed, Name: "observed", TypeID: typeID},
	}
	derivedLoan := ""
	if derivation == core.OpBorrowShared {
		derivedLoan = function.ID + ":loan:1"
	}
	linear.Operations = []core.LinearOperation{
		{ID: function.ID + ":op:0", PointID: function.ID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: owner, TargetID: view, LoanID: function.ID + ":loan:0", TypeID: typeID},
		{ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1", Kind: derivation, SourceID: view, TargetID: review, LoanID: derivedLoan, TypeID: typeID},
		{ID: function.ID + ":op:2", PointID: function.ID + ":point:linear:2", Kind: core.OpMove, SourceID: owner, TargetID: delivered, TypeID: typeID},
	}
	if observeAfterMove {
		linear.Operations = append(linear.Operations,
			core.LinearOperation{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpCopy, SourceID: review, TargetID: observed, TypeID: typeID},
			core.LinearOperation{ID: function.ID + ":op:4", PointID: function.ID + ":point:linear:4", Kind: core.OpReturn, SourceID: delivered, TypeID: typeID},
		)
	} else {
		linear.Places = linear.Places[:4]
		linear.Operations = append(linear.Operations,
			core.LinearOperation{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpReturn, SourceID: delivered, TypeID: typeID},
		)
	}
	return program
}

func TestTransitiveLoanBlocksMove(t *testing.T) {
	for _, testCase := range []struct {
		name       string
		derivation core.OperationKind
	}{
		{name: "reborrow of a loan", derivation: core.OpBorrowShared},
		{name: "copy of a loan", derivation: core.OpCopy},
	} {
		blocked := corevalidate.Validate(reborrowProgram(testCase.derivation, true))
		if blocked.Valid || len(blocked.Problems) == 0 || blocked.Problems[0].Code != "core.move_while_borrowed" {
			t.Fatalf("%s: owner moved while a transitively derived loan was still observed: %+v", testCase.name, blocked)
		}
		unused := corevalidate.Validate(reborrowProgram(testCase.derivation, false))
		if !unused.Valid {
			t.Fatalf("%s: transitive loan liveness over-blocked a loan with no later use: %+v", testCase.name, unused)
		}
	}
}

func scaleProgram(facts int) core.Program {
	if facts < 1 {
		panic("scale facts must include the final claim")
	}
	functionID := "s1:scale:fn:copy"
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	linear := &core.LinearBody{
		ID: functionID + ":linear",
		Types: []core.TypeFact{{
			ID: typeID, Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
			Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
			NegativeWitnesses: []core.AbilityWitness{},
		}},
		Places:     make([]core.Place, 0, facts),
		Operations: make([]core.LinearOperation, 0, facts),
	}
	linear.Places = append(linear.Places, core.Place{ID: parameterID, Name: "source", TypeID: typeID})
	for index := 0; index < facts-1; index++ {
		targetID := fmt.Sprintf("%s:place:%d", functionID, index+1)
		linear.Places = append(linear.Places, core.Place{ID: targetID, Name: fmt.Sprintf("copy%d", index), TypeID: typeID})
		linear.Operations = append(linear.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: core.OpCopy, SourceID: parameterID, TargetID: targetID, TypeID: typeID,
		})
	}
	last := facts - 1
	linear.Operations = append(linear.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, last), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, last),
		Kind: core.OpReturn, SourceID: parameterID, TypeID: typeID,
	})
	return core.Program{
		Schema: core.Schema1, Module: "scale", ModuleID: "s1:scale:module:scale",
		Functions: []core.Function{{
			ID: functionID, Name: "copy", EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter: core.Parameter{ID: parameterID, Name: "source", Type: "Byte"}, ReturnType: "Byte", Linear: linear,
		}},
	}
}

func cloneProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	encoded, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.Program
	if err := json.Unmarshal(encoded, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func foreignAcquireProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestForeignRefusalsAreIndependentlyDerived proves corevalidate re-derives
// D-04-16's missing-policy refusal and D-04-02's call-target refusal purely
// from the core artifact, never from check's own AST-level facts: each
// mutation below is fed DIRECTLY to corevalidate.Validate, bypassing
// check.Program entirely, so the refusal cannot be riding on check's own
// gate.
func TestForeignRefusalsAreIndependentlyDerived(t *testing.T) {
	valid := foreignAcquireProgram(t)
	if result := corevalidate.Validate(valid); !result.Valid {
		t.Fatalf("valid foreign-call core rejected: %+v", result)
	}

	t.Run("missing unwind policy", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.Unwind = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.unwind_policy_undeclared" {
			t.Fatalf("expected foreign.unwind_policy_undeclared, got %+v", result)
		}
	})

	t.Run("missing nonlocal_exit policy", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		mutated.Functions[0].ForeignContract.NonlocalExit = ""
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "foreign.unwind_policy_undeclared" {
			t.Fatalf("expected foreign.unwind_policy_undeclared, got %+v", result)
		}
	})

	t.Run("call target resolves to a declared Lang function", func(t *testing.T) {
		mutated := cloneProgram(t, valid)
		// Introduce a second, otherwise-inert function whose name collides
		// with the foreign contract's symbol -- the same shape check.go's
		// own admission gate refuses at the AST level, re-derived here
		// purely from core.Function.Name identity.
		decoy := cloneProgram(t, valid).Functions[0]
		decoy.ID = decoy.ID + ":decoy"
		decoy.Name = mutated.Functions[0].ForeignContract.Symbol
		decoy.ForeignContract = nil
		decoy.Linear.ID = decoy.ID + ":linear"
		for index := range decoy.Linear.Places {
			decoy.Linear.Places[index].ID = decoy.ID + fmt.Sprintf(":place:%d", index)
		}
		mutated.Functions = append(mutated.Functions, decoy)
		result := corevalidate.Validate(mutated)
		if result.Valid || result.Problems[0].Code != "core.call_target_not_foreign" {
			t.Fatalf("expected core.call_target_not_foreign, got %+v", result)
		}
	})
}

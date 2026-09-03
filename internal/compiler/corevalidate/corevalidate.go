// Package corevalidate independently validates inert typed-core facts before
// an execution engine consumes them. It intentionally does not know source or
// reuse checker/interpreter authorization code.
package corevalidate

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/codename-lang/lang/internal/compiler/core"
)

const maxTypeDepth = 64

type Problem struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

type Result struct {
	Valid    bool      `json:"valid"`
	Problems []Problem `json:"problems,omitempty"`
	Checks   int       `json:"checks"`
	program  core.Program
}

// Program returns a content-owned copy of the validated program. Callers can
// neither mutate the validator's copy nor race validation by retaining slices.
func (r Result) Program() core.Program { return cloneProgram(r.program) }

// Validate is deliberately source-blind. It proves internal consistency of a
// typed-core statement, not that a coordinated producer translated source
// truthfully.
func Validate(input core.Program) Result {
	owned := cloneProgram(input)
	v := validator{program: owned}
	v.run()
	return Result{Valid: len(v.problems) == 0, Problems: v.problems, Checks: v.checks, program: owned}
}

type validator struct {
	program  core.Program
	checks   int
	problems []Problem
}

func (v *validator) check(ok bool, code, detail string) bool {
	v.checks++
	if ok {
		return true
	}
	if len(v.problems) == 0 {
		v.problems = append(v.problems, Problem{Code: code, Detail: detail})
	}
	return false
}

func (v *validator) run() {
	if !v.check(v.program.Schema == core.Schema || v.program.Schema == core.Schema1, "core.schema", v.program.Schema) {
		return
	}
	if !v.check(v.program.Module != "" && v.program.ModuleID != "", "core.module", "missing module identity") {
		return
	}
	functionIDs := make(map[string]struct{}, len(v.program.Functions))
	for index := range v.program.Functions {
		function := &v.program.Functions[index]
		if !v.unique(functionIDs, function.ID, "core.duplicate_function_id") {
			return
		}
		if !v.check(function.HasClosedBody(), "core.invalid_body", function.ID) {
			return
		}
		if function.Linear != nil {
			if !v.check(v.program.Schema == core.Schema1, "core.schema", "linear body requires lang.core/1") || !v.linear(function) {
				return
			}
		} else if !v.check(v.program.Schema == core.Schema, "core.schema", "match body requires lang.core/0") || !v.match(function) {
			return
		}
	}
}

func (v *validator) match(function *core.Function) bool {
	match := function.Match
	if !v.check(match.ID != "" && match.PointID != "" && function.EntryPointID != "" && function.ReturnPointID != "", "core.invalid_id", function.ID) {
		return false
	}
	armIDs := make(map[string]struct{}, len(match.Arms))
	edgeIDs := make(map[string]struct{}, len(match.Arms))
	for _, arm := range match.Arms {
		if !v.unique(armIDs, arm.ID, "core.duplicate_arm_id") || !v.unique(edgeIDs, arm.EdgeID, "core.duplicate_edge_id") {
			return false
		}
	}
	return true
}

func (v *validator) linear(function *core.Function) bool {
	linear := function.Linear
	if !v.check(linear.ID == function.ID+":linear", "core.linear_id", linear.ID) ||
		!v.check(function.EntryPointID == function.ID+":point:entry" && function.ReturnPointID == function.ID+":point:return", "core.point_id", function.ID) {
		return false
	}

	types := make(map[string]core.TypeFact, len(linear.Types))
	for index, fact := range linear.Types {
		if !v.uniqueType(types, fact) {
			return false
		}
		if !v.check(fact.ID == fmt.Sprintf("%s:type:%d", function.ID, index), "core.type_order", fact.ID) {
			return false
		}
		abilities, witnesses, ok := v.derive(fact.Shape, 0)
		if !ok {
			return false
		}
		if !v.check(reflect.DeepEqual(fact.Abilities, abilities) && reflect.DeepEqual(fact.NegativeWitnesses, witnesses), "core.ability_mismatch", fact.ID) {
			return false
		}
	}

	places := make(map[string]core.Place, len(linear.Places))
	for index, place := range linear.Places {
		if !v.uniquePlace(places, place) {
			return false
		}
		if !v.check(place.ID == fmt.Sprintf("%s:place:%d", function.ID, index), "core.place_order", place.ID) {
			return false
		}
		if _, ok := types[place.TypeID]; !v.check(ok, "core.unknown_type", place.TypeID) {
			return false
		}
	}
	parameter, ok := places[function.Parameter.ID]
	if !v.check(ok, "core.unknown_place", function.Parameter.ID) ||
		!v.check(parameter.Name == function.Parameter.Name, "core.parameter_mismatch", function.Parameter.ID) {
		return false
	}
	parameterType := types[parameter.TypeID]
	if !v.check(parameterType.Shape.Constructor == function.Parameter.Type, "core.parameter_mismatch", function.Parameter.ID) {
		return false
	}

	operationIDs := make(map[string]struct{}, len(linear.Operations))
	pointIDs := make(map[string]struct{}, len(linear.Operations))
	loanIDs := make(map[string]struct{})
	for index, operation := range linear.Operations {
		if !v.unique(operationIDs, operation.ID, "core.duplicate_operation_id") || !v.unique(pointIDs, operation.PointID, "core.duplicate_point_id") {
			return false
		}
		if !v.check(operation.ID == fmt.Sprintf("%s:op:%d", function.ID, index) && operation.PointID == fmt.Sprintf("%s:point:linear:%d", function.ID, index), "core.operation_order", operation.ID) {
			return false
		}
		if _, ok := places[operation.SourceID]; !v.check(ok, "core.unknown_place", operation.SourceID) {
			return false
		}
		if _, ok := types[operation.TypeID]; !v.check(ok, "core.unknown_type", operation.TypeID) {
			return false
		}
		if operation.Kind != core.OpReturn {
			if _, ok := places[operation.TargetID]; !v.check(ok, "core.unknown_place", operation.TargetID) {
				return false
			}
		}
		if operation.Kind == core.OpBorrowShared {
			if !v.unique(loanIDs, operation.LoanID, "core.unknown_loan") {
				return false
			}
		} else if !v.check(operation.LoanID == "", "core.unknown_loan", operation.LoanID) {
			return false
		}
	}
	return v.replay(function, types, places)
}

func (v *validator) replay(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	operations := function.Linear.Operations
	initialized := map[string]bool{function.Parameter.ID: true}
	loanOwner := make(map[string]string)
	loanTarget := make(map[string]string)
	loanLastUse := make(map[string]int)
	for index, operation := range operations {
		for loanID, targetID := range loanTarget {
			if operation.SourceID == targetID {
				loanLastUse[loanID] = index
			}
		}
		if operation.Kind == core.OpBorrowShared {
			loanOwner[operation.LoanID] = operation.SourceID
			loanTarget[operation.LoanID] = operation.TargetID
			loanLastUse[operation.LoanID] = index
		}
	}

	returned := false
	for index, operation := range operations {
		source := places[operation.SourceID]
		if !v.check(source.TypeID == operation.TypeID, "core.type_mismatch", operation.ID) {
			return false
		}
		if !initialized[operation.SourceID] {
			code := "core.place_uninitialized"
			if operation.Kind == core.OpReturn {
				code = "core.final_claim_mismatch"
			}
			return v.check(false, code, operation.SourceID)
		}
		switch operation.Kind {
		case core.OpCopy:
			if !hasAbility(types[operation.TypeID], core.AbilityCopy) {
				return v.check(false, "core.ability.copy_denied", operation.TypeID)
			}
			if !v.targetMatches(operation, places) {
				return false
			}
			initialized[operation.TargetID] = true
		case core.OpBorrowShared:
			if !hasAbility(types[operation.TypeID], core.AbilityShare) {
				return v.check(false, "core.ability.share_denied", operation.TypeID)
			}
			if !v.targetMatches(operation, places) {
				return false
			}
			initialized[operation.TargetID] = true
		case core.OpMove:
			for loanID, ownerID := range loanOwner {
				if ownerID == operation.SourceID && loanLastUse[loanID] >= index {
					return v.check(false, "core.move_while_borrowed", operation.ID)
				}
			}
			if !v.targetMatches(operation, places) {
				return false
			}
			initialized[operation.SourceID] = false
			initialized[operation.TargetID] = true
		case core.OpReturn:
			if index != len(operations)-1 || returned || operation.TargetID != "" || operation.TypeID != source.TypeID || types[source.TypeID].Shape.Constructor != function.ReturnType {
				return v.check(false, "core.final_claim_mismatch", operation.ID)
			}
			returned = true
		default:
			return v.check(false, "core.unknown_operation", string(operation.Kind))
		}
	}
	return v.check(returned, "core.final_claim_mismatch", function.ID)
}

func (v *validator) targetMatches(operation core.LinearOperation, places map[string]core.Place) bool {
	return v.check(places[operation.TargetID].TypeID == operation.TypeID, "core.type_mismatch", operation.TargetID)
}

func (v *validator) derive(shape core.TypeRef, depth int) ([]core.Ability, []core.AbilityWitness, bool) {
	v.checks++
	if depth > maxTypeDepth {
		v.problems = append(v.problems, Problem{Code: "core.type_limit", Detail: shape.Constructor})
		return nil, nil, false
	}
	order := []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape}
	granted := make([]core.Ability, 0, len(order))
	witnesses := make([]core.AbilityWitness, 0)
	for _, requested := range order {
		ok, path, known := deriveAbility(shape, requested, depth)
		if !known {
			v.problems = append(v.problems, Problem{Code: "core.unknown_type_constructor", Detail: shape.Constructor})
			return nil, nil, false
		}
		if ok {
			granted = append(granted, requested)
		} else {
			witnesses = append(witnesses, core.AbilityWitness{Ability: requested, Path: path})
		}
	}
	return granted, witnesses, true
}

func deriveAbility(shape core.TypeRef, requested core.Ability, depth int) (bool, []string, bool) {
	if depth > maxTypeDepth {
		return false, nil, false
	}
	switch shape.Constructor {
	case "Byte":
		if len(shape.Arguments) != 0 {
			return false, nil, false
		}
		return true, nil, true
	case "Buffer":
		if len(shape.Arguments) != 0 {
			return false, nil, false
		}
		if requested == core.AbilityCopy || requested == core.AbilityShare {
			return false, []string{"Buffer"}, true
		}
		return true, nil, true
	case "Box":
		if len(shape.Arguments) != 1 {
			return false, nil, false
		}
		ok, path, known := deriveAbility(shape.Arguments[0], requested, depth+1)
		if !ok && known {
			path = append([]string{"Box.value"}, path...)
		}
		return ok, path, known
	case "Pair":
		if len(shape.Arguments) != 2 {
			return false, nil, false
		}
		for index, argument := range shape.Arguments {
			ok, path, known := deriveAbility(argument, requested, depth+1)
			if !known {
				return false, nil, false
			}
			if !ok {
				field := "Pair.left"
				if index == 1 {
					field = "Pair.right"
				}
				return false, append([]string{field}, path...), true
			}
		}
		return true, nil, true
	default:
		return false, nil, false
	}
}

func (v *validator) unique(set map[string]struct{}, id, code string) bool {
	v.checks++
	if id == "" {
		if len(v.problems) == 0 {
			v.problems = append(v.problems, Problem{Code: code, Detail: "empty id"})
		}
		return false
	}
	if _, exists := set[id]; exists {
		if len(v.problems) == 0 {
			v.problems = append(v.problems, Problem{Code: code, Detail: id})
		}
		return false
	}
	set[id] = struct{}{}
	return true
}

func (v *validator) uniqueType(set map[string]core.TypeFact, fact core.TypeFact) bool {
	v.checks++
	if fact.ID == "" {
		v.problems = append(v.problems, Problem{Code: "core.duplicate_type_id", Detail: "empty id"})
		return false
	}
	if _, exists := set[fact.ID]; exists {
		v.problems = append(v.problems, Problem{Code: "core.duplicate_type_id", Detail: fact.ID})
		return false
	}
	set[fact.ID] = fact
	return true
}

func (v *validator) uniquePlace(set map[string]core.Place, place core.Place) bool {
	v.checks++
	if place.ID == "" {
		v.problems = append(v.problems, Problem{Code: "core.duplicate_place_id", Detail: "empty id"})
		return false
	}
	if _, exists := set[place.ID]; exists {
		v.problems = append(v.problems, Problem{Code: "core.duplicate_place_id", Detail: place.ID})
		return false
	}
	set[place.ID] = place
	return true
}

func hasAbility(fact core.TypeFact, requested core.Ability) bool {
	for _, granted := range fact.Abilities {
		if granted == requested {
			return true
		}
	}
	return false
}

func cloneProgram(program core.Program) core.Program {
	encoded, err := json.Marshal(program)
	if err != nil {
		panic(fmt.Sprintf("core validation clone: %v", err))
	}
	var clone core.Program
	if err := json.Unmarshal(encoded, &clone); err != nil {
		panic(fmt.Sprintf("core validation clone: %v", err))
	}
	return clone
}

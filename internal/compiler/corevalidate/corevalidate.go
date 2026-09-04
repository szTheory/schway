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

// KnownEscape names the exact boundary this validator cannot prove. A producer
// that coordinates a false source claim with internally consistent core facts
// remains outside source-blind validation.
const KnownEscape = "escape:coordinated-source-core-lie"

// LinearWorkLimit is the exact declared count for the canonical scale shape:
// one sealed Byte fact, one parameter plus one target per copy, and facts-1
// copies followed by one final return claim.
func LinearWorkLimit(facts int) int { return 16*facts + 13 }

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
	if !v.check(len(v.program.Functions) > 0, "core.function_missing", v.program.ModuleID) {
		return
	}
	dataNames := make(map[string]core.DataType, len(v.program.DataTypes))
	dataIDs := make(map[string]struct{}, len(v.program.DataTypes))
	for _, dataType := range v.program.DataTypes {
		if !v.unique(dataIDs, dataType.ID, "core.duplicate_type_id") || !v.check(dataType.Name != "", "core.unknown_type", dataType.ID) {
			return
		}
		if _, exists := dataNames[dataType.Name]; !v.check(!exists, "core.duplicate_type_name", dataType.Name) {
			return
		}
		alternatives := make(map[string]struct{}, len(dataType.Alternatives))
		for _, alternative := range dataType.Alternatives {
			if !v.unique(alternatives, alternative, "core.duplicate_alternative") {
				return
			}
		}
		dataNames[dataType.Name] = dataType
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
		switch {
		case function.Match != nil && function.Linear != nil:
			// A match function whose arms carry linear bodies (Phase 3): the
			// arm-body facts are additive on lang.core/1, exactly like a
			// straight-line linear body, so the same schema requirement
			// applies. This is the third branch of the body/schema
			// cross-lock (03-PATTERNS I-7): linear-only requires /1,
			// match-only requires /0, match-with-arm-bodies also requires
			// /1.
			if !v.check(v.program.Schema == core.Schema1, "core.schema", "a match with an arm body requires lang.core/1") || !v.matchBranch(function, dataNames) {
				return
			}
		case function.Linear != nil:
			if !v.check(v.program.Schema == core.Schema1, "core.schema", "linear body requires lang.core/1") || !v.linear(function) {
				return
			}
		default:
			if !v.check(v.program.Schema == core.Schema, "core.schema", "match body requires lang.core/0") || !v.match(function, dataNames) {
				return
			}
		}
	}
}

// matchBranch validates a match function whose arms carry linear bodies. It
// validates the match-shaped facts (arm/edge identity, pattern coverage,
// exhaustiveness) independently of the checker exactly as match() does, then
// falls through to linear() for the flattened operations, places, types, and
// the new Block/Edge facts the arms lowered into.
func (v *validator) matchBranch(function *core.Function, dataNames map[string]core.DataType) bool {
	match := function.Match
	if !v.check(match.ID != "" && match.PointID != "" && function.EntryPointID != "" && function.ReturnPointID != "", "core.invalid_id", function.ID) {
		return false
	}
	dataType, knownType := dataNames[function.Parameter.Type]
	if !v.check(knownType && function.ReturnType == function.Parameter.Type, "core.unknown_type", function.Parameter.Type) ||
		!v.check(match.Scrutinee == function.Parameter.Name, "core.unknown_place", match.Scrutinee) {
		return false
	}
	alternatives := make(map[string]struct{}, len(dataType.Alternatives))
	for _, alternative := range dataType.Alternatives {
		alternatives[alternative] = struct{}{}
	}
	armIDs := make(map[string]struct{}, len(match.Arms))
	edgeIDs := make(map[string]struct{}, len(match.Arms))
	patterns := make(map[string]struct{}, len(match.Arms))
	for _, arm := range match.Arms {
		if !v.unique(armIDs, arm.ID, "core.duplicate_arm_id") || !v.unique(edgeIDs, arm.EdgeID, "core.duplicate_edge_id") {
			return false
		}
		if !v.unique(patterns, arm.Pattern, "core.duplicate_pattern") {
			return false
		}
		if _, known := alternatives[arm.Pattern]; !v.check(known, "core.unknown_alternative", arm.ID) {
			return false
		}
		// This phase's scope decision (checkBranch): a match that carries
		// any arm body requires every arm to carry one. Value and BlockID
		// are therefore mutually exclusive per arm, and BlockID must be
		// present here — a bare Value in a branch-shaped function is
		// rejected rather than silently tolerated.
		if !v.check(arm.Value == "" && arm.BlockID != "", "core.invalid_body", arm.ID) {
			return false
		}
	}
	if !v.check(len(patterns) == len(alternatives), "core.final_claim_mismatch", match.ID) {
		return false
	}
	return v.linear(function)
}

func (v *validator) match(function *core.Function, dataNames map[string]core.DataType) bool {
	match := function.Match
	if !v.check(match.ID != "" && match.PointID != "" && function.EntryPointID != "" && function.ReturnPointID != "", "core.invalid_id", function.ID) {
		return false
	}
	armIDs := make(map[string]struct{}, len(match.Arms))
	edgeIDs := make(map[string]struct{}, len(match.Arms))
	dataType, knownType := dataNames[function.Parameter.Type]
	if !v.check(knownType && function.ReturnType == function.Parameter.Type, "core.unknown_type", function.Parameter.Type) ||
		!v.check(match.Scrutinee == function.Parameter.Name, "core.unknown_place", match.Scrutinee) {
		return false
	}
	alternatives := make(map[string]struct{}, len(dataType.Alternatives))
	for _, alternative := range dataType.Alternatives {
		alternatives[alternative] = struct{}{}
	}
	patterns := make(map[string]struct{}, len(match.Arms))
	for _, arm := range match.Arms {
		if !v.unique(armIDs, arm.ID, "core.duplicate_arm_id") || !v.unique(edgeIDs, arm.EdgeID, "core.duplicate_edge_id") {
			return false
		}
		if !v.unique(patterns, arm.Pattern, "core.duplicate_pattern") {
			return false
		}
		_, patternKnown := alternatives[arm.Pattern]
		_, valueKnown := alternatives[arm.Value]
		if !v.check(patternKnown && valueKnown, "core.unknown_alternative", arm.ID) {
			return false
		}
	}
	return v.check(len(patterns) == len(alternatives), "core.final_claim_mismatch", match.ID)
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
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			if !v.unique(loanIDs, operation.LoanID, "core.unknown_loan") {
				return false
			}
		} else if !v.check(operation.LoanID == "", "core.unknown_loan", operation.LoanID) {
			return false
		}
	}
	if len(linear.Blocks) > 0 || len(linear.Edges) > 0 {
		if !v.blocksAndEdges(function, operationIDs) {
			return false
		}
	}
	return v.replay(function, types, places)
}

// blocksAndEdges independently validates the Phase 3 CFG facts: every block
// and edge ID is unique and non-empty, every block's OperationIDs and every
// edge's endpoints resolve into facts this function already declared, and
// every operation ID is claimed by exactly one block (an operation that
// belongs to no block, or to two, is rejected rather than silently ignored).
// This is deliberately the same "unique + referential closure" shape as
// every other ordinal-bearing slice in this file (types/places/operations),
// applied to the two new slices T-03-01/T-03-04 introduce.
func (v *validator) blocksAndEdges(function *core.Function, operationIDs map[string]struct{}) bool {
	linear := function.Linear
	blockIDs := make(map[string]struct{}, len(linear.Blocks))
	claimedOperations := make(map[string]struct{}, len(linear.Operations))
	for _, block := range linear.Blocks {
		if !v.unique(blockIDs, block.ID, "core.duplicate_block_id") {
			return false
		}
		for _, opID := range block.OperationIDs {
			if _, ok := operationIDs[opID]; !v.check(ok, "core.unknown_operation_reference", opID) {
				return false
			}
			if _, already := claimedOperations[opID]; !v.check(!already, "core.duplicate_operation_id", opID) {
				return false
			}
			claimedOperations[opID] = struct{}{}
		}
	}
	for _, operation := range linear.Operations {
		if _, claimed := claimedOperations[operation.ID]; !v.check(claimed, "core.unknown_block", operation.ID) {
			return false
		}
	}
	edgeIDs := make(map[string]struct{}, len(linear.Edges))
	for _, edge := range linear.Edges {
		if !v.unique(edgeIDs, edge.ID, "core.duplicate_edge_id") {
			return false
		}
		if _, ok := blockIDs[edge.FromBlockID]; !v.check(ok, "core.unknown_block", edge.FromBlockID) {
			return false
		}
		if _, ok := blockIDs[edge.ToBlockID]; !v.check(ok, "core.unknown_block", edge.ToBlockID) {
			return false
		}
	}
	for _, block := range linear.Blocks {
		for _, successor := range block.Successors {
			if _, ok := blockIDs[successor]; !v.check(ok, "core.unknown_block", successor) {
				return false
			}
		}
	}
	return true
}

func (v *validator) replay(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	if len(function.Linear.Blocks) > 0 {
		return v.replayBlocks(function, types, places)
	}
	return v.replayStraightLine(function, types, places)
}

func (v *validator) replayStraightLine(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	operations := function.Linear.Operations
	initialized := map[string]bool{function.Parameter.ID: true}
	produced := map[string]bool{function.Parameter.ID: true}
	loanOwner := make(map[string]string)
	// loansForPlace is transitive: a place produced from a loan-derived place
	// carries every loan its source carried. Recording only the immediate
	// borrow target lets a reborrow or a copy of a loan expire the original
	// loan one operation early and admit a move while it is still observable.
	loansForPlace := make(map[string][]string)
	loanLastUse := make(map[string]int)
	for index, operation := range operations {
		v.checks++ // inspect each operation once while finding final loan uses
		carried := append([]string(nil), loansForPlace[operation.SourceID]...)
		for _, loanID := range carried {
			loanLastUse[loanID] = index
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			loanOwner[operation.LoanID] = operation.SourceID
			loanLastUse[operation.LoanID] = index
			carried = append(carried, operation.LoanID)
		}
		if operation.Kind != core.OpReturn && len(carried) > 0 {
			loansForPlace[operation.TargetID] = carried
		}
	}
	ownerBlockedUntil := make(map[string]int)
	for loanID, ownerID := range loanOwner {
		v.checks++ // consolidate each declared loan exactly once
		if loanLastUse[loanID] > ownerBlockedUntil[ownerID] {
			ownerBlockedUntil[ownerID] = loanLastUse[loanID]
		}
	}
	// ownerLiveSharedUntil/ownerLiveExclusiveUntil independently re-derive
	// the five-row conflict matrix (T-03-07): unlike ownerBlockedUntil (a
	// single whole-function aggregate, used only for the move-while-
	// borrowed law, which is safe because a place can never be re-borrowed
	// after it is moved), these two maps are updated incrementally in
	// program order as pass 2 below encounters each new loan, so a loan
	// created later in the function can never appear to "block" an earlier
	// one. This is a materially different mechanism from check.go's
	// per-owner active-loan set with immediate per-index expiry (D-12): it
	// derives conflict from an aggregated running high-water mark per
	// access mode, not from a checker-shaped set of currently-live loan
	// objects.
	ownerLiveSharedUntil := make(map[string]int)
	ownerLiveExclusiveUntil := make(map[string]int)

	returned := false
	for index, operation := range operations {
		source := places[operation.SourceID]
		if !v.check(source.TypeID == operation.TypeID, "core.type_mismatch", operation.ID) {
			return false
		}
		if !v.check(initialized[operation.SourceID], finalOrTransitionCode(operation.Kind), operation.SourceID) {
			return false
		}
		v.checks++ // dispatch one independently authorized transition
		switch operation.Kind {
		case core.OpCopy:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityCopy), "core.ability.copy_denied", operation.TypeID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpBorrowShared:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityShare), "core.ability.share_denied", operation.TypeID) {
				return false
			}
			if until, blocked := ownerLiveExclusiveUntil[operation.SourceID]; !v.check(!blocked || until < index, "core.borrow_conflict", operation.ID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if lastUse := loanLastUse[operation.LoanID]; lastUse > ownerLiveSharedUntil[operation.SourceID] {
				ownerLiveSharedUntil[operation.SourceID] = lastUse
			}
		case core.OpBorrowExclusive:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityShare), "core.ability.share_denied", operation.TypeID) {
				return false
			}
			sharedUntil, sharedBlocked := ownerLiveSharedUntil[operation.SourceID]
			exclusiveUntil, exclusiveBlocked := ownerLiveExclusiveUntil[operation.SourceID]
			conflict := (sharedBlocked && sharedUntil >= index) || (exclusiveBlocked && exclusiveUntil >= index)
			if !v.check(!conflict, "core.borrow_conflict", operation.ID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if lastUse := loanLastUse[operation.LoanID]; lastUse > ownerLiveExclusiveUntil[operation.SourceID] {
				ownerLiveExclusiveUntil[operation.SourceID] = lastUse
			}
		case core.OpMove:
			blockedUntil, hasLoan := ownerBlockedUntil[operation.SourceID]
			if !v.check(!hasLoan || blockedUntil < index, "core.move_while_borrowed", operation.ID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.SourceID] = false
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
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

// replayBlocks is replayStraightLine's counterpart for a branch-shaped
// function (T-03-04's independent re-derivation of the arm-body case). The
// initialized/produced/loan bookkeeping is identical and safe to run exactly
// the same way over the whole flat Operations list, because every arm's
// places are distinct global ordinals — one arm's move can never be observed
// by a sibling arm that never executes at the same time. The one law that
// must change is OpReturn: a branch-shaped function has one return PER BLOCK
// (the block the mutually-exclusive selected arm executes), not one return
// for the whole function, so "last operation in the whole slice" becomes
// "last operation in its own block", and "returned" becomes a per-block set
// rather than one flag.
func (v *validator) replayBlocks(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	linear := function.Linear
	operations := linear.Operations
	lastOperationOfBlock := make(map[string]string, len(linear.Blocks))
	blockOfOperation := make(map[string]string, len(operations))
	for _, block := range linear.Blocks {
		for index, opID := range block.OperationIDs {
			blockOfOperation[opID] = block.ID
			if index == len(block.OperationIDs)-1 {
				lastOperationOfBlock[block.ID] = opID
			}
		}
	}

	initialized := map[string]bool{function.Parameter.ID: true}
	produced := map[string]bool{function.Parameter.ID: true}
	loanOwner := make(map[string]string)
	loansForPlace := make(map[string][]string)
	loanLastUse := make(map[string]int)
	for index, operation := range operations {
		v.checks++ // inspect each operation once while finding final loan uses
		carried := append([]string(nil), loansForPlace[operation.SourceID]...)
		for _, loanID := range carried {
			loanLastUse[loanID] = index
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			loanOwner[operation.LoanID] = operation.SourceID
			loanLastUse[operation.LoanID] = index
			carried = append(carried, operation.LoanID)
		}
		if operation.Kind != core.OpReturn && len(carried) > 0 {
			loansForPlace[operation.TargetID] = carried
		}
	}
	ownerBlockedUntil := make(map[string]int)
	for loanID, ownerID := range loanOwner {
		v.checks++ // consolidate each declared loan exactly once
		if loanLastUse[loanID] > ownerBlockedUntil[ownerID] {
			ownerBlockedUntil[ownerID] = loanLastUse[loanID]
		}
	}
	// See replayStraightLine's identical declaration for why these two maps
	// (rather than a single whole-function aggregate) independently
	// re-derive the five-row conflict matrix.
	ownerLiveSharedUntil := make(map[string]int)
	ownerLiveExclusiveUntil := make(map[string]int)

	returnedBlocks := make(map[string]bool, len(linear.Blocks))
	for index, operation := range operations {
		source := places[operation.SourceID]
		if !v.check(source.TypeID == operation.TypeID, "core.type_mismatch", operation.ID) {
			return false
		}
		if !v.check(initialized[operation.SourceID], finalOrTransitionCode(operation.Kind), operation.SourceID) {
			return false
		}
		v.checks++ // dispatch one independently authorized transition
		switch operation.Kind {
		case core.OpCopy:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityCopy), "core.ability.copy_denied", operation.TypeID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpBorrowShared:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityShare), "core.ability.share_denied", operation.TypeID) {
				return false
			}
			if until, blocked := ownerLiveExclusiveUntil[operation.SourceID]; !v.check(!blocked || until < index, "core.borrow_conflict", operation.ID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if lastUse := loanLastUse[operation.LoanID]; lastUse > ownerLiveSharedUntil[operation.SourceID] {
				ownerLiveSharedUntil[operation.SourceID] = lastUse
			}
		case core.OpBorrowExclusive:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityShare), "core.ability.share_denied", operation.TypeID) {
				return false
			}
			sharedUntil, sharedBlocked := ownerLiveSharedUntil[operation.SourceID]
			exclusiveUntil, exclusiveBlocked := ownerLiveExclusiveUntil[operation.SourceID]
			conflict := (sharedBlocked && sharedUntil >= index) || (exclusiveBlocked && exclusiveUntil >= index)
			if !v.check(!conflict, "core.borrow_conflict", operation.ID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if lastUse := loanLastUse[operation.LoanID]; lastUse > ownerLiveExclusiveUntil[operation.SourceID] {
				ownerLiveExclusiveUntil[operation.SourceID] = lastUse
			}
		case core.OpMove:
			blockedUntil, hasLoan := ownerBlockedUntil[operation.SourceID]
			if !v.check(!hasLoan || blockedUntil < index, "core.move_while_borrowed", operation.ID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.SourceID] = false
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpReturn:
			blockID, known := blockOfOperation[operation.ID]
			if !v.check(known && lastOperationOfBlock[blockID] == operation.ID, "core.final_claim_mismatch", operation.ID) {
				return false
			}
			if !v.check(!returnedBlocks[blockID] && operation.TargetID == "" && operation.TypeID == source.TypeID && types[source.TypeID].Shape.Constructor == function.ReturnType, "core.final_claim_mismatch", operation.ID) {
				return false
			}
			returnedBlocks[blockID] = true
		default:
			return v.check(false, "core.unknown_operation", string(operation.Kind))
		}
	}
	for _, block := range linear.Blocks {
		if len(block.OperationIDs) == 0 {
			continue
		}
		if !v.check(returnedBlocks[block.ID], "core.final_claim_mismatch", block.ID) {
			return false
		}
	}
	return true
}

func (v *validator) targetMatches(function *core.Function, operationIndex int, operation core.LinearOperation, places map[string]core.Place, produced map[string]bool) bool {
	target := places[operation.TargetID]
	expected := fmt.Sprintf("%s:place:%d", function.ID, operationIndex+1)
	return v.check(
		operation.TargetID == expected && operation.TargetID != operation.SourceID && !produced[operation.TargetID] && target.TypeID == operation.TypeID,
		"core.invalid_target",
		operation.TargetID,
	)
}

func (v *validator) derive(shape core.TypeRef, depth int) ([]core.Ability, []core.AbilityWitness, bool) {
	nodes, bounded := boundedTypeNodes(shape)
	v.checks += nodes
	if depth > maxTypeDepth || !bounded {
		v.problems = append(v.problems, Problem{Code: "core.type_limit", Detail: shape.Constructor})
		return nil, nil, false
	}
	order := []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape}
	granted := make([]core.Ability, 0, len(order))
	witnesses := make([]core.AbilityWitness, 0)
	sealed := v.sealedLeaves()
	for _, requested := range order {
		ok, path, known := deriveAbility(shape, requested, depth, sealed)
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

// sealedLeaves names every declared data type in this program (Phase 3): a
// field-less nominal alternative type (e.g. a match scrutinee's own type) is
// trivially copy/drop/share/send/escape-safe, so it is re-derived here as a
// sealed structural leaf, exactly as check.go's ability.DeriveSealed does on
// the checker side — an independent implementation of the same law, per D-12.
func (v *validator) sealedLeaves() map[string]bool {
	names := make(map[string]bool, len(v.program.DataTypes))
	for _, dataType := range v.program.DataTypes {
		names[dataType.Name] = true
	}
	return names
}

func boundedTypeNodes(shape core.TypeRef) (int, bool) {
	type item struct {
		shape core.TypeRef
		depth int
	}
	stack := []item{{shape: shape}}
	count := 0
	for len(stack) > 0 {
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		count++
		if count > 4096 || current.depth > maxTypeDepth {
			return count, false
		}
		for _, argument := range current.shape.Arguments {
			stack = append(stack, item{shape: argument, depth: current.depth + 1})
		}
	}
	return count, true
}

func finalOrTransitionCode(kind core.OperationKind) string {
	if kind == core.OpReturn {
		return "core.final_claim_mismatch"
	}
	return "core.place_uninitialized"
}

func deriveAbility(shape core.TypeRef, requested core.Ability, depth int, sealed map[string]bool) (bool, []string, bool) {
	if depth > maxTypeDepth {
		return false, nil, false
	}
	if sealed[shape.Constructor] {
		if len(shape.Arguments) != 0 {
			return false, nil, false
		}
		return true, nil, true
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
		// Buffer denies only copy: a shared loan observes without duplicating
		// ownership, so share is granted. Derived independently of the
		// ability package's structural combiner.
		if requested == core.AbilityCopy {
			return false, []string{"Buffer"}, true
		}
		return true, nil, true
	case "Box":
		if len(shape.Arguments) != 1 {
			return false, nil, false
		}
		ok, path, known := deriveAbility(shape.Arguments[0], requested, depth+1, sealed)
		if !ok && known {
			path = append([]string{"Box.value"}, path...)
		}
		return ok, path, known
	case "Pair":
		if len(shape.Arguments) != 2 {
			return false, nil, false
		}
		for index, argument := range shape.Arguments {
			ok, path, known := deriveAbility(argument, requested, depth+1, sealed)
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

// Package corevalidate independently validates inert typed-core facts before
// an execution engine consumes them. It intentionally does not know source or
// reuse checker/interpreter authorization code.
package corevalidate

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

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
//
// The +1 over the pre-03-04 formula (16*facts+13) is loanChainIndex's own
// honest, once-per-place counted work (D-05): every operation in this scale
// shape reads from the SAME parameter place (never a derived one), so
// carriedLoans visits exactly one NEW place — the parameter itself — no
// matter how large facts grows, and memoizes it after that single visit.
// The constant is therefore facts-independent by construction, not an
// unaccounted bump: TestValidatorReborrowChainIsLinear separately proves the
// chain-walk's own cost scales with chain depth when the scale shape
// actually has one to walk.
func LinearWorkLimit(facts int) int { return 16*facts + 14 }

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
	// v.linear itself independently re-derives and compares LoanEndpoints
	// (T-03-13/D-12) once Blocks are present — see its own doc comment.
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
		// OpFail and OpDefect are terminators alongside OpReturn (D-04-04/
		// D-04-15): none of the three ever carries a TargetID, since none
		// produces an ordinary place -- OpReturn ends the function, OpFail
		// ends the err block, OpDefect ends an arm block by aborting.
		if operation.Kind != core.OpReturn && operation.Kind != core.OpFail && operation.Kind != core.OpRelease && operation.Kind != core.OpDefect {
			if _, ok := places[operation.TargetID]; !v.check(ok, "core.unknown_place", operation.TargetID) {
				return false
			}
		}
		if operation.Kind == core.OpForeignCall {
			if _, ok := places[operation.ErrTargetID]; !v.check(ok, "core.unknown_place", operation.ErrTargetID) {
				return false
			}
			if !v.check(operation.OkEdgeID != "" && operation.ErrEdgeID != "", "core.foreign_call_edges_missing", operation.ID) {
				return false
			}
			if !v.check(function.ForeignContract != nil && function.ForeignContract.Symbol != "", "core.foreign_contract_missing", operation.ID) {
				return false
			}
			// D-04-16, independently derived: check.go refuses a missing
			// unwind/nonlocal_exit policy at admission time by inspecting
			// ast.ForeignPolicy (never emitting a core artifact for such a
			// symbol at all); this validator never reads that AST-level
			// fact -- it re-derives the SAME refusal purely from the two
			// policy string fields core.ForeignContract itself carries, so
			// a corrupted core artifact that skipped check.go's gate is
			// still caught here.
			if !v.check(function.ForeignContract.Unwind != "" && function.ForeignContract.NonlocalExit != "", "foreign.unwind_policy_undeclared", operation.ID) {
				return false
			}
			// D-04-02, independently derived: check.go refuses a callee that
			// resolves to a Lang function name at parse-resolution time (an
			// AST-level, pre-core fact this validator never sees). This
			// validator re-derives the same refusal by a materially
			// different mechanism -- a straight name collision scan against
			// every OTHER function this core.Program itself declares --
			// reading only the core artifact, never check's own name table.
			if !v.check(!isDeclaredFunctionName(v.program.Functions, function.ID, function.ForeignContract.Symbol), "core.call_target_not_foreign", operation.ID) {
				return false
			}
			// Phase 4 plan 03 (D-04-12/FFI-01): every foreign obligation
			// category must be present -- no default value that would let an
			// omission pass as a declaration -- and the declared Layout must
			// be internally consistent. This is independently derived from
			// check's own admission gate: it reads only the four flat string
			// fields and the Layout struct core.ForeignContract itself
			// carries, never check's AST-level foreignSymbolInfo.
			if !v.check(function.ForeignContract.InitializedState != "" && function.ForeignContract.Capture != "" && function.ForeignContract.Retention != "" && function.ForeignContract.Aliasing != "", "foreign.obligation_undeclared", operation.ID) {
				return false
			}
			if !v.foreignLayoutConsistent(function.ForeignContract.Layout, operation.ID) {
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
	if !v.replay(function, types, places) {
		return false
	}
	if len(linear.Blocks) > 0 {
		// T-03-13/D-12: independently re-derive the declared loan-endpoint
		// set (recomputeLoanEndpoints) and require whole-value equality
		// against what the producer declared. A moved, dropped, or invented
		// endpoint is caught here, not by trusting the checker's own
		// dataflow. Gated on Blocks (never populated for a straight-line
		// body this phase) rather than on the match/linear dispatch shape,
		// so this runs for any branch-shaped function regardless of which
		// run() case dispatched here.
		return v.loanEndpointsMatch(function)
	}
	return true
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
	// D-04-09: the only producer of an OpFail is an err edge. check.go never
	// emits one elsewhere, by construction, but a hand-corrupted core.Program
	// routing any other edge into a fail-terminated block must still be
	// refused here, independently -- exactly the same defense-in-depth
	// posture releaseAllocatorMatches already applies to a release's declared
	// allocator.
	operationsByID := make(map[string]core.LinearOperation, len(linear.Operations))
	for _, operation := range linear.Operations {
		operationsByID[operation.ID] = operation
	}
	incoming := make(map[string][]core.Edge, len(linear.Edges))
	for _, edge := range linear.Edges {
		incoming[edge.ToBlockID] = append(incoming[edge.ToBlockID], edge)
	}
	for _, block := range linear.Blocks {
		if len(block.OperationIDs) == 0 {
			continue
		}
		last := operationsByID[block.OperationIDs[len(block.OperationIDs)-1]]
		if last.Kind != core.OpFail {
			continue
		}
		edges := incoming[block.ID]
		if !v.check(len(edges) > 0, "core.fail_edge_missing", last.ID) {
			return false
		}
		for _, edge := range edges {
			if !v.check(edge.Pattern == "err", "core.fail_reached_without_err_edge", last.ID) {
				return false
			}
		}
	}
	return true
}

// loanChainIndex is corevalidate's own place-provenance bookkeeping: for
// every place produced by an operation, parent records the place it derived
// from (its SourceID) and bornAt records the loan a place was freshly
// created FOR, if the producing operation was a borrow. This is basic
// data-flow bookkeeping intrinsic to the core artifact itself (any correct
// reader has to know which loans a place carries), not the liveness
// algorithm under test — replayStraightLine/replayBlocks still decide
// conflicts and endpoints their own way once carriedLoans answers "what does
// this place currently carry".
//
// carriedLoans replaces the O(n)-per-operation append-and-copy this file
// used to do (D-02-03/Q2(b), validator half): instead of materializing a
// fresh copy of a place's whole loan ancestry on every reference, it walks
// the parent chain once per place, recursively, and memoizes the result —
// so a place already visited (directly or as another place's ancestor)
// answers in O(1) on every later reference, and the chain as a whole is
// walked in full at most once per function.
type loanChainIndex struct {
	bornAt map[string]string
	parent map[string]string
	memo   map[string][]string
	checks *int
}

func buildLoanChainIndex(operations []core.LinearOperation, checks *int) *loanChainIndex {
	idx := &loanChainIndex{
		bornAt: make(map[string]string, len(operations)),
		parent: make(map[string]string, len(operations)),
		memo:   make(map[string][]string, len(operations)),
		checks: checks,
	}
	for _, operation := range operations {
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			idx.bornAt[operation.TargetID] = operation.LoanID
		}
		if operation.TargetID != "" {
			idx.parent[operation.TargetID] = operation.SourceID
		}
	}
	return idx
}

// carriedLoans walks the parent chain ITERATIVELY, not recursively, and
// stops the moment it revisits a place already on the current walk. A
// self-referencing or cyclic parent pointer can only arise from a corrupted
// core artifact (e.g. a borrow target retargeted onto its own source) — this
// package's ordinary place/target validation rejects that elsewhere in the
// same pass — but this helper must never crash on one: a recursive walk
// would recurse forever (or exhaust the goroutine stack) on exactly that
// input, which is precisely the kind of adversarial input a source-blind
// validator must stay defined against.
func (idx *loanChainIndex) carriedLoans(placeID string) []string {
	if placeID == "" {
		return nil
	}
	if cached, ok := idx.memo[placeID]; ok {
		return cached
	}
	var path []string
	visited := make(map[string]bool, 1)
	current := placeID
	for current != "" {
		if cached, ok := idx.memo[current]; ok {
			return idx.foldChain(path, cached)
		}
		if visited[current] {
			return idx.foldChain(path, nil) // cyclic parent pointer: stop, do not recurse
		}
		visited[current] = true
		path = append(path, current)
		current = idx.parent[current]
	}
	return idx.foldChain(path, nil)
}

// foldChain memoizes every place on path, walking from the outermost
// (furthest from the original query) back to the innermost, each place's
// loan list being its own birth loan (if any) prepended to whatever loans
// the rest of the chain already carries.
func (idx *loanChainIndex) foldChain(path []string, tail []string) []string {
	loans := tail
	for i := len(path) - 1; i >= 0; i-- {
		place := path[i]
		*idx.checks++ // one unit per NEWLY visited place while walking its ancestry, once ever
		var here []string
		if loanID, born := idx.bornAt[place]; born {
			here = append(here, loanID)
		}
		here = append(here, loans...)
		idx.memo[place] = here
		loans = here
	}
	return loans
}

// blockReach computes, for every declared block, the set of OTHER blocks
// reachable by following one or more successor edges — a one-shot
// reachability closure, deliberately not the iterative worklist fixpoint
// check.go's loanLivenessFixpoint uses. Each root's breadth-first search
// visits a block at most once (a visited-once frontier, never re-enqueued),
// so the computation terminates in bounded time even over an
// (illegitimately) cyclic declared graph — unlike the checker, which
// explicitly detects and rejects a cycle before analysis, this closure has
// no need to: a cyclic declaration just makes every block in the cycle
// mutually reachable, which can never make a dishonest core artifact pass
// the endpoint comparison it would otherwise fail.
func blockReach(blocks []core.Block, index map[string]int) []map[int]bool {
	reach := make([]map[int]bool, len(blocks))
	for i, block := range blocks {
		visited := make(map[int]bool, len(blocks))
		queue := append([]string(nil), block.Successors...)
		for len(queue) > 0 {
			id := queue[0]
			queue = queue[1:]
			successorIndex, known := index[id]
			if !known || visited[successorIndex] {
				continue
			}
			visited[successorIndex] = true
			queue = append(queue, blocks[successorIndex].Successors...)
		}
		reach[i] = visited
	}
	return reach
}

// recomputeLoanEndpoints independently re-derives a branch-shaped function's
// loan-endpoint set directly from the declared Blocks/Edges/Operations, by a
// mechanism check.go's loanLivenessFixpoint does not use (T-03-13/D-12).
// loanLivenessFixpoint iterates a worklist, re-evaluating each block's
// transfer function against its successors' live-in sets until nothing
// changes (a converging monotone dataflow). This pass never iterates to
// convergence at all. Instead it:
//
//  1. Materializes the full "use relation" up front — every (loanID,
//     blockIdx, ordinal) at which some operation actually references that
//     loan (through its own place-provenance chain) — as an explicit list,
//     not folded into a converging live-in/live-out set.
//  2. Computes the reachability closure of the declared successor relation
//     (blockReach) once.
//  3. Reduces, per loan: a block's local last reference is a POINT endpoint
//     exactly when no direct successor of that block, nor anything
//     reachable from it, ever references the loan again (needsAt is a pure
//     membership test against the materialized relation, not a value that
//     converges). A block with more than one successor, where the loan is
//     needed by at least one successor's reachable future but not by
//     another, gets an EDGE endpoint on exactly the successor edge that
//     does not need it. A loan with no reference anywhere past its own
//     creation ends at its own birth operation (the same fallback
//     materializeLoanEndpoints uses when nothing else claims a loan).
//
// Neither derivation is obtainable from the other by renaming: one converges
// a set under repeated re-evaluation; the other computes one static closure
// and one static reduction over an explicitly materialized relation, with no
// shared helper, no import, and no worklist anywhere in this file. On any
// acyclic graph (OWN-03's scope this phase) the two are mathematically
// forced to agree, because both compute the same underlying partial order of
// "happens before, reachably" — so an honest producer's declared endpoints
// always match, while a corrupted endpoint set (moved, dropped, or invented)
// is caught by a genuinely independent recomputation, not by an oracle that
// could share a wrong law with the producer (WR-02).
func (v *validator) recomputeLoanEndpoints(function *core.Function) []core.LoanEndpoint {
	linear := function.Linear
	if len(linear.Blocks) == 0 {
		return nil
	}

	blockIndex := make(map[string]int, len(linear.Blocks))
	for index, block := range linear.Blocks {
		blockIndex[block.ID] = index
	}
	edgeIDByPair := make(map[[2]string]string, len(linear.Edges))
	for _, edge := range linear.Edges {
		edgeIDByPair[[2]string{edge.FromBlockID, edge.ToBlockID}] = edge.ID
	}
	reach := blockReach(linear.Blocks, blockIndex)

	operationBlock := make(map[string]int, len(linear.Operations))
	operationOrdinal := make(map[string]int, len(linear.Operations))
	for _, block := range linear.Blocks {
		for ordinal, opID := range block.OperationIDs {
			operationBlock[opID] = blockIndex[block.ID]
			operationOrdinal[opID] = ordinal
		}
	}

	chain := buildLoanChainIndex(linear.Operations, &v.checks)

	type birthFact struct {
		blockIdx int
		ordinal  int
		opID     string
	}
	born := make(map[string]birthFact)
	for _, operation := range linear.Operations {
		v.checks++ // one inspection per operation while materializing loan births
		if operation.Kind != core.OpBorrowShared && operation.Kind != core.OpBorrowExclusive {
			continue
		}
		blockIdx, known := operationBlock[operation.ID]
		if !known {
			continue // referential-closure gaps are already rejected by blocksAndEdges
		}
		born[operation.LoanID] = birthFact{blockIdx: blockIdx, ordinal: operationOrdinal[operation.ID], opID: operation.ID}
	}

	type reference struct {
		blockIdx int
		ordinal  int
		opID     string
	}
	referencesByLoan := make(map[string][]reference)
	for _, operation := range linear.Operations {
		blockIdx, known := operationBlock[operation.ID]
		if !known {
			continue
		}
		for _, loanID := range chain.carriedLoans(operation.SourceID) {
			v.checks++ // one inspection per materialized (operation, referenced loan) pair
			referencesByLoan[loanID] = append(referencesByLoan[loanID], reference{blockIdx: blockIdx, ordinal: operationOrdinal[operation.ID], opID: operation.ID})
		}
	}

	var endpoints []core.LoanEndpoint
	for loanID, birth := range born {
		v.checks++ // one reduction pass per declared loan

		referencedAt := make(map[int]bool, len(linear.Blocks))
		lastOrdinalAt := make(map[int]int, len(linear.Blocks))
		lastOpAt := make(map[int]string, len(linear.Blocks))
		for _, ref := range referencesByLoan[loanID] {
			if !referencedAt[ref.blockIdx] || ref.ordinal > lastOrdinalAt[ref.blockIdx] {
				lastOrdinalAt[ref.blockIdx] = ref.ordinal
				lastOpAt[ref.blockIdx] = ref.opID
			}
			referencedAt[ref.blockIdx] = true
		}
		needsAt := func(blockIdx int) bool {
			if referencedAt[blockIdx] {
				return true
			}
			for target := range referencedAt {
				if reach[blockIdx][target] {
					return true
				}
			}
			return false
		}

		scope := []int{birth.blockIdx}
		for target := range reach[birth.blockIdx] {
			scope = append(scope, target)
		}
		for _, blockIdx := range scope {
			v.checks++ // one classification decision per block in this loan's reachable scope
			block := linear.Blocks[blockIdx]
			liveOut := false
			for _, successorID := range block.Successors {
				if needsAt(blockIndex[successorID]) {
					liveOut = true
					break
				}
			}
			if len(block.Successors) > 1 && liveOut {
				for _, successorID := range block.Successors {
					if needsAt(blockIndex[successorID]) {
						continue
					}
					id := edgeIDByPair[[2]string{block.ID, successorID}]
					endpoints = append(endpoints, core.LoanEndpoint{ID: id + ":" + loanID, LoanID: loanID, Kind: "edge", EdgeID: id})
				}
			}
			if referencedAt[blockIdx] && !liveOut {
				endpoints = append(endpoints, core.LoanEndpoint{
					ID: fmt.Sprintf("%s:point:%s:%d:%s", function.ID, block.ID, lastOrdinalAt[blockIdx], loanID),
					LoanID: loanID, Kind: "point", BlockID: block.ID, AfterOperationID: lastOpAt[blockIdx],
				})
			}
		}

		if len(referencesByLoan[loanID]) == 0 {
			// Fallback: a loan never referenced after its own creation ends
			// at its own birth operation — the same degenerate case
			// materializeLoanEndpoints' `!recorded[operation.LoanID]` branch
			// covers on the checker side.
			block := linear.Blocks[birth.blockIdx]
			endpoints = append(endpoints, core.LoanEndpoint{
				ID: fmt.Sprintf("%s:point:%s:%d:%s", function.ID, block.ID, birth.ordinal, loanID),
				LoanID: loanID, Kind: "point", BlockID: block.ID, AfterOperationID: birth.opID,
			})
		}
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].ID < endpoints[j].ID })
	return endpoints
}

func (v *validator) loanEndpointsMatch(function *core.Function) bool {
	recomputed := v.recomputeLoanEndpoints(function)
	return v.check(reflect.DeepEqual(recomputed, function.Linear.LoanEndpoints), "core.loan_endpoint_mismatch", function.ID)
}

func (v *validator) replay(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	if len(function.Linear.Blocks) > 0 {
		return v.replayBlocks(function, types, places)
	}
	return v.replayStraightLine(function, types, places)
}

func (v *validator) replayStraightLine(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	operations := function.Linear.Operations
	operationsByID := make(map[string]core.LinearOperation, len(operations))
	for _, operation := range operations {
		operationsByID[operation.ID] = operation
	}
	initialized := map[string]bool{function.Parameter.ID: true}
	produced := map[string]bool{function.Parameter.ID: true}
	loanOwner := make(map[string]string)
	// loanLastUse is transitive: a place produced from a loan-derived place
	// carries every loan its source carried. Recording only the immediate
	// borrow target lets a reborrow or a copy of a loan expire the original
	// loan one operation early and admit a move while it is still observable.
	// D-02-03/Q2(b), validator half: this used to re-copy a place's whole
	// loan-ancestry list on every operation (a per-operation append-and-copy
	// that grows with chain depth). loanChainIndex.carriedLoans replaces
	// that with a memoized parent-pointer chain walked at most once per
	// place across the whole function.
	loanLastUse := make(map[string]int)
	chain := buildLoanChainIndex(operations, &v.checks)
	for index, operation := range operations {
		v.checks++ // inspect each operation once while finding final loan uses
		for _, loanID := range chain.carriedLoans(operation.SourceID) {
			loanLastUse[loanID] = index
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			loanOwner[operation.LoanID] = operation.SourceID
			loanLastUse[operation.LoanID] = index
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
		case core.OpForeignCall:
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			errTarget, errKnown := places[operation.ErrTargetID]
			if !v.check(errKnown && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
				return false
			}
			initialized[operation.ErrTargetID] = true
			produced[operation.ErrTargetID] = true
		case core.OpFail:
			if !v.check(index == len(operations)-1 && !returned && operation.TargetID == "", "core.final_claim_mismatch", operation.ID) {
				return false
			}
			returned = true
		case core.OpDefect:
			// A terminator alongside OpReturn/OpFail (D-04-15): never the
			// last operation of anything but its own straight-line body (this
			// case exists only so control:kind.exhaustive_dispatch's table
			// finds it handled here too -- a straight-line, block-less body
			// can never actually carry a defect this phase).
			if !v.check(index == len(operations)-1 && !returned && operation.TargetID == "" && operation.Reason != "", "core.final_claim_mismatch", operation.ID) {
				return false
			}
			returned = true
		case core.OpRelease:
			// A non-terminal transition (D-04-07): it discharges an
			// already-produced acquisition and produces no new place. The
			// reverse-order release sequence itself is independently
			// rederived and compared in checkReleaseOrder, not here --
			// replay only needs to confirm the discharged place is a real,
			// already-initialized place, which the generic pre-switch
			// checks above already established.
			if !v.check(operation.ReleasesOperationID != "", "core.release_target_unknown", operation.ID) {
				return false
			}
			if !v.releaseAllocatorMatches(operation, operationsByID) {
				return false
			}
		default:
			return v.check(false, "core.unknown_operation", string(operation.Kind))
		}
	}
	return v.check(returned, "core.final_claim_mismatch", function.ID)
}

// releaseAllocatorMatches independently re-derives T-04-14's
// allocator-identity requirement: re-fetch the OpForeignCall this release
// names (never trusting check's own bookkeeping) and require its Allocator
// to match. A release naming an acquisition whose allocator differs is
// refused with foreign.release_allocator_mismatch.
func (v *validator) releaseAllocatorMatches(operation core.LinearOperation, operationsByID map[string]core.LinearOperation) bool {
	acquisition, ok := operationsByID[operation.ReleasesOperationID]
	if !ok {
		// An unresolvable ReleasesOperationID (e.g. an "invented" release
		// naming an operation that does not exist) is caught by
		// checkReleaseOrder's own sequence comparison, not here -- this
		// check's job is narrower: given a REAL acquisition, its allocator
		// must match.
		return true
	}
	return v.check(acquisition.Allocator == operation.Allocator, "foreign.release_allocator_mismatch", operation.ID)
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
	operationsByID := make(map[string]core.LinearOperation, len(operations))
	for _, operation := range operations {
		operationsByID[operation.ID] = operation
	}
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
	// See replayStraightLine's identical declaration for why a memoized
	// parent-pointer chain (loanChainIndex), not a per-operation copy of the
	// accumulated loan list, computes loanLastUse here (D-02-03/Q2(b)).
	loanLastUse := make(map[string]int)
	chain := buildLoanChainIndex(operations, &v.checks)
	for index, operation := range operations {
		v.checks++ // inspect each operation once while finding final loan uses
		for _, loanID := range chain.carriedLoans(operation.SourceID) {
			loanLastUse[loanID] = index
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			loanOwner[operation.LoanID] = operation.SourceID
			loanLastUse[operation.LoanID] = index
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
		case core.OpForeignCall:
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			errTarget, errKnown := places[operation.ErrTargetID]
			if !v.check(errKnown && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
				return false
			}
			initialized[operation.ErrTargetID] = true
			produced[operation.ErrTargetID] = true
		case core.OpFail:
			// OpFail is replayBlocks' err-edge terminator (D-04-04): unlike
			// OpReturn it never carries the function's own ReturnType --
			// the whole point of the edge is that its payload is a
			// DIFFERENT, independently declared failure ADT. Every other
			// per-block terminal requirement (last operation of its block,
			// no target, exactly one terminal per block) still applies.
			blockID, known := blockOfOperation[operation.ID]
			if !v.check(known && lastOperationOfBlock[blockID] == operation.ID, "core.final_claim_mismatch", operation.ID) {
				return false
			}
			if !v.check(!returnedBlocks[blockID] && operation.TargetID == "" && operation.TypeID == source.TypeID, "core.final_claim_mismatch", operation.ID) {
				return false
			}
			returnedBlocks[blockID] = true
		case core.OpDefect:
			// A per-block terminator alongside OpReturn/OpFail (D-04-15): the
			// only route into a "defect" terminal outcome, admissible only in
			// a match arm's terminal position this phase (checkBranch never
			// produces it elsewhere). It carries a required non-empty Reason
			// and never a TargetID.
			blockID, known := blockOfOperation[operation.ID]
			if !v.check(known && lastOperationOfBlock[blockID] == operation.ID, "core.final_claim_mismatch", operation.ID) {
				return false
			}
			if !v.check(!returnedBlocks[blockID] && operation.TargetID == "" && operation.Reason != "", "core.final_claim_mismatch", operation.ID) {
				return false
			}
			returnedBlocks[blockID] = true
		case core.OpRelease:
			// A non-terminal transition (D-04-07), never the last operation
			// of a genuinely terminal block (an err/success block always
			// ends in OpFail/OpReturn) -- see the exhaustive-dispatch note
			// on replayStraightLine's identical case for why replay's own
			// obligation here is narrow: confirm the release names a real
			// discharge target, and leave the reverse-order sequence itself
			// to the independent rederivation in checkReleaseOrder.
			if !v.check(operation.ReleasesOperationID != "", "core.release_target_unknown", operation.ID) {
				return false
			}
			if !v.releaseAllocatorMatches(operation, operationsByID) {
				return false
			}
		default:
			return v.check(false, "core.unknown_operation", string(operation.Kind))
		}
	}
	if !v.checkReleaseOrder(function, operationsByID) {
		return false
	}
	for _, block := range linear.Blocks {
		if len(block.OperationIDs) == 0 {
			continue
		}
		lastOpID := block.OperationIDs[len(block.OperationIDs)-1]
		if operationsByID[lastOpID].Kind == core.OpForeignCall {
			// A block whose last operation is OpForeignCall forks into its
			// two successor edges instead of terminating itself (D-04-04);
			// only a block whose last operation is an actual terminator
			// (OpReturn/OpFail) must appear in returnedBlocks.
			continue
		}
		if !v.check(returnedBlocks[block.ID], "core.final_claim_mismatch", block.ID) {
			return false
		}
	}
	return true
}

// checkReleaseOrder independently rederives the reverse-order release
// sequence D-04-07 requires. For every block ending in OpFail or OpReturn, it
// walks BACKWARD over the declared block/edge graph starting from that
// block's own incoming edge, collecting every completed (not-yet-discharged)
// OpForeignCall acquisition it passes -- a discovery order that is already
// reverse-of-completion order, because the walk moves from the fail/return
// point back toward the entry. A failure block's own triggering acquisition
// is excluded (its incoming edge has Pattern "err", so includeThis starts
// false); a success block's immediate predecessor IS included (its incoming
// edge has Pattern "ok", so it completed). This is materially different from
// check.go's forward accumulation (D-12/D-12a): it never reads check's own
// accumulated list or any field check uses to communicate it, only the
// block/edge graph and the operations check.go already emitted. Only a call
// whose ok and err edges target DIFFERENT blocks is tracked -- a `discard`'s
// converging ok/err edges mark its resource as untracked for release this
// plan, a documented narrowing shared with check.go's own accumulation.
func (v *validator) checkReleaseOrder(function *core.Function, operationsByID map[string]core.LinearOperation) bool {
	linear := function.Linear
	edgesByTo := make(map[string][]core.Edge, len(linear.Edges))
	okEdgeInto := make(map[string]core.Edge, len(linear.Edges))
	for _, edge := range linear.Edges {
		edgesByTo[edge.ToBlockID] = append(edgesByTo[edge.ToBlockID], edge)
		if edge.Pattern == "ok" {
			okEdgeInto[edge.ToBlockID] = edge
		}
	}
	// tracked names exactly the acquisitions SOME OpRelease in this function
	// discharges -- the same rule check.go's own materialization and cgen's
	// runtime ledger both use, so a discard's untracked acquisition (no
	// OpRelease ever names it) and the 04-01 tracer's acquisition (no
	// OpRelease exists in that shape at all) are both correctly excluded,
	// even though a tracer call's ok/err edges also diverge.
	tracked := make(map[string]bool, len(linear.Operations))
	for _, operation := range linear.Operations {
		v.checks++ // one inspection per operation while locating tracked acquisitions
		if operation.Kind == core.OpRelease && operation.ReleasesOperationID != "" {
			tracked[operation.ReleasesOperationID] = true
		}
	}
	callInBlock := make(map[string]core.LinearOperation, len(linear.Blocks))
	for _, block := range linear.Blocks {
		for _, opID := range block.OperationIDs {
			operation := operationsByID[opID]
			if operation.Kind == core.OpForeignCall && tracked[operation.ID] {
				callInBlock[block.ID] = operation
			}
		}
	}

	rederive := func(startEdge core.Edge) []core.LinearOperation {
		var expected []core.LinearOperation
		includeThis := startEdge.Pattern == "ok"
		currentBlockID := startEdge.FromBlockID
		for {
			v.checks++ // one inspection per block visited while walking backward
			if includeThis {
				if op, ok := callInBlock[currentBlockID]; ok {
					expected = append(expected, op)
				}
			}
			edge, ok := okEdgeInto[currentBlockID]
			if !ok {
				break
			}
			currentBlockID = edge.FromBlockID
			includeThis = true
		}
		return expected
	}

	for _, block := range linear.Blocks {
		if len(block.OperationIDs) == 0 {
			continue
		}
		lastOp := operationsByID[block.OperationIDs[len(block.OperationIDs)-1]]
		if lastOp.Kind != core.OpFail && lastOp.Kind != core.OpReturn {
			continue
		}
		v.checks++ // one reduction pass per terminal block
		incoming := edgesByTo[block.ID]
		if len(incoming) != 1 {
			continue
		}
		expected := rederive(incoming[0])
		var actual []core.LinearOperation
		for _, opID := range block.OperationIDs {
			if operation := operationsByID[opID]; operation.Kind == core.OpRelease {
				actual = append(actual, operation)
			}
		}
		if !v.check(len(expected) == len(actual), "core.release_order_mismatch", block.ID) {
			return false
		}
		for index, want := range expected {
			if !v.check(actual[index].ReleasesOperationID == want.ID, "core.release_order_mismatch", actual[index].ID) {
				return false
			}
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

// isDeclaredFunctionName reports whether name matches the Name of any OTHER
// function this core.Program declares (excludeFunctionID excludes the
// caller's own function, since a Lang function's own name is never its own
// foreign contract's symbol in any honest artifact). D-04-01: OpForeignCall
// is the only call surface -- a foreign contract naming a real Lang
// function is exactly the inadmissible shape D-04-02 refuses.
// foreignLayoutConsistent independently validates a declared core.RecordLayout's
// internal consistency (D-04-12/task 04-03-01): the field list must be
// non-empty, offsets must be strictly ascending in declaration order, no two
// fields may share a name, and every field's offset+size must fit within the
// record's own declared size. This reads only the core artifact's own
// Layout struct -- never the AST, never check's own foreignSymbolInfo -- so
// a hand-corrupted core.Program is caught exactly the same way a corrupted
// checker output would be.
func (v *validator) foreignLayoutConsistent(layout *core.RecordLayout, detail string) bool {
	if !v.check(layout != nil && len(layout.Fields) > 0 && layout.ForeignTypeName != "", "foreign.layout_invalid", detail) {
		return false
	}
	seenNames := make(map[string]struct{}, len(layout.Fields))
	previousOffset := -1
	for _, field := range layout.Fields {
		v.checks++ // one inspection per declared layout field
		if _, duplicate := seenNames[field.Name]; !v.check(field.Name != "" && !duplicate, "foreign.layout_invalid", detail) {
			return false
		}
		seenNames[field.Name] = struct{}{}
		if !v.check(field.Offset > previousOffset, "foreign.layout_invalid", detail) {
			return false
		}
		previousOffset = field.Offset
		if !v.check(field.Size > 0 && field.Alignment > 0 && field.Offset+field.Size <= layout.Size, "foreign.layout_invalid", detail) {
			return false
		}
	}
	return v.check(layout.Size > 0 && layout.Alignment > 0, "foreign.layout_invalid", detail)
}

func isDeclaredFunctionName(functions []core.Function, excludeFunctionID, name string) bool {
	for _, candidate := range functions {
		if candidate.ID == excludeFunctionID {
			continue
		}
		if candidate.Name == name {
			return true
		}
	}
	return false
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

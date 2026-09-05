package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/ability"
	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

// testOnlyForceUniformLoanJoin is a fault-injection seam for
// TestUniformJoinPlacementFlipsBothVerdicts (03-03-02): when true, every
// arm-body loan's computed last use is forced to len(body.Bindings) -- as
// if edge-specific placement had been deleted and every loan ended
// uniformly at the join regardless of which edge actually needs it --
// instead of whatever discoverLoanLastUses would otherwise compute. It is
// read only by analyzeArmBody, never by analyzeStraightLine, so the
// straight-line exhaustive differential (TestOwnershipSequenceExhaustive)
// is untouched by its existence. This is a test-only seam, not a
// production code path: a test that ships its own seam cannot be
// falsified by reverting a production hunk (the seam would simply never
// engage), so this specific claim is falsified by direct mutation instead
// -- flipping this variable and observing a verdict change IS the
// falsifying action, recorded verbatim in the owning plan's summary
// (D-09, mirroring 02-VALIDATION.md's recorded residual-weakness
// precedent for fault-injection seams).
var testOnlyForceUniformLoanJoin = false

type Result struct {
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
	Work        int
}

func Program(program ast.Program) Result {
	result := Result{Program: core.Program{Schema: core.Schema, Module: program.Module, ModuleID: semanticID(program.Module, "module", program.Module)}}
	hasMatch, hasLinear := false, false
	for _, function := range program.Funcs {
		if function.Body.Linear != nil {
			hasLinear = true
		} else {
			hasMatch = true
		}
	}
	if hasMatch && hasLinear {
		span := diagnostic.Span{}
		if len(program.Funcs) > 0 {
			span = program.Funcs[0].Span
		}
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
			"core.mixed_body_versions",
			span,
			"a module cannot mix match and linear function bodies until function-level core versioning is defined",
		))
		return result
	}
	types := make(map[string]core.DataType)
	for _, declaration := range program.Data {
		alternatives := make([]string, 0, len(declaration.Alternatives))
		for _, alternative := range declaration.Alternatives {
			alternatives = append(alternatives, alternative.Name)
		}
		dataType := core.DataType{ID: semanticID(program.Module, "type", declaration.Name), Name: declaration.Name, Alternatives: alternatives, Span: declaration.Span}
		types[declaration.Name] = dataType
		result.Program.DataTypes = append(result.Program.DataTypes, dataType)
	}

	// Phase 4: the foreign symbol table and the Lang function-name set both
	// exist purely so a fallible call's callee can be resolved against one or
	// the other (D-04-01/D-04-02) -- a callee resolving into functionNames
	// rather than foreignSymbols is core.call_target_not_foreign, never an
	// ordinary unknown-name error.
	foreignSymbols, foreignDiagnostics := collectForeignSymbols(program)
	if len(foreignDiagnostics) > 0 {
		result.Diagnostics = append(result.Diagnostics, foreignDiagnostics...)
		return result
	}
	functionNames := make(map[string]bool, len(program.Funcs))
	for _, function := range program.Funcs {
		functionNames[function.Name] = true
	}

	for _, function := range program.Funcs {
		functionID := semanticID(program.Module, "fn", function.Name)
		if !function.Body.HasClosedVariant() {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("core.invalid_body", function.Span, "function must have exactly one body variant"))
			continue
		}
		// OWN-04 scope fence: a match-bodied (S1) function — bare-arm or
		// arm-body/branch alike — must not return a borrowed view this phase.
		// This is deliberately checked before either match path below so the
		// rejection carries a named, span-bearing cause instead of falling
		// through to the generic type.return_mismatch the bare-match path
		// would otherwise produce for a return type it does not recognize.
		if function.Body.Linear == nil && function.ReturnOrigin != nil {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
				"ownership.match_borrowed_return_unsupported", function.ReturnOrigin.Span,
				"a match-bodied function cannot declare a borrowed return origin this phase",
			))
			continue
		}
		if function.Body.Linear != nil {
			var checked core.Function
			var diagnostics []diagnostic.Diagnostic
			var work int
			if hasTryCall(function.Body.Linear) {
				checked, diagnostics, work = checkFallibleLinear(functionID, function, foreignSymbols, functionNames, types)
			} else {
				checked, diagnostics, work = checkLinear(program.Module, functionID, function)
			}
			result.Work += work
			result.Diagnostics = append(result.Diagnostics, diagnostics...)
			if len(diagnostics) == 0 {
				result.Program.Schema = core.Schema1
				result.Program.Functions = append(result.Program.Functions, checked)
			}
			continue
		}
		matchID := semanticID(program.Module, "match", function.Name)
		dataType, ok := types[function.Parameter.Type.Constructor]
		if !ok {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.unknown", function.Parameter.Span, "unknown parameter type"))
			continue
		}
		if !sameType(function.ReturnType, function.Parameter.Type) {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.return_mismatch", function.Span, "S1 match result must have the parameter type"))
			continue
		}
		if function.Body.Scrutinee != function.Parameter.Name {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("name.unknown_scrutinee", function.Body.Span, "match scrutinee is not the function parameter"))
			continue
		}
		for _, arm := range function.Body.Arms {
			if !arm.HasClosedVariant() {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("core.invalid_body", arm.Span, "match arm must have exactly one value form"))
				continue
			}
		}
		hasArmBody := false
		for _, arm := range function.Body.Arms {
			if arm.Body != nil {
				hasArmBody = true
				break
			}
		}
		if hasArmBody {
			checked, diagnostics, work := checkBranch(program.Module, functionID, matchID, function, dataType, sealedNames(types))
			result.Work += work
			result.Diagnostics = append(result.Diagnostics, diagnostics...)
			if len(diagnostics) == 0 {
				result.Program.Schema = core.Schema1
				result.Program.Functions = append(result.Program.Functions, checked)
			}
			continue
		}
		seen := make(map[string]bool)
		arms := make([]core.MatchArm, 0, len(function.Body.Arms))
		for index, arm := range function.Body.Arms {
			if seen[arm.Pattern] {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.subsumed", arm.Span, "alternative is already matched"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Pattern) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Value) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.invalid_variant", arm.Span, "match result is not an alternative of the return type"))
				continue
			}
			seen[arm.Pattern] = true
			arms = append(arms, core.MatchArm{
				ID: fmt.Sprintf("%s:arm:%d", functionID, index), EdgeID: fmt.Sprintf("%s:edge:%s", matchID, arm.Pattern),
				Pattern: arm.Pattern, Value: arm.Value,
			})
		}
		missing := make([]string, 0)
		for _, alternative := range dataType.Alternatives {
			if !seen[alternative] {
				missing = append(missing, alternative)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			causes := make([]diagnostic.Cause, 0, len(missing))
			for _, name := range missing {
				causes = append(causes, diagnostic.Cause{Kind: "missing_alternative", Detail: name})
			}
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes...))
		}
		result.Program.Functions = append(result.Program.Functions, core.Function{
			ID: functionID, Name: function.Name,
			EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter:  core.Parameter{ID: semanticID(program.Module, "parameter", function.Name+"."+function.Parameter.Name), Name: function.Parameter.Name, Type: function.Parameter.Type.Constructor},
			ReturnType: function.ReturnType.Constructor,
			Match:      &core.Match{ID: matchID, PointID: functionID + ":point:match", Scrutinee: function.Body.Scrutinee, Arms: arms},
			Span:       function.Span,
		})
	}
	return result
}

func sealedNames(types map[string]core.DataType) map[string]bool {
	names := make(map[string]bool, len(types))
	for name := range types {
		names[name] = true
	}
	return names
}

// maxBlocksPerFunction bounds T-03-01's CFG-shape denial-of-service surface
// at the lowering layer: entry block, one block per body arm, and the join
// block. Sized generously above anything this phase's fixtures need.
//
// Deliberately unreachable from real source today, exactly like
// ownership.borrow_requires_share (check_test.go's nonShareableTypeFact):
// the parser's own maxArmsPerMatch (64) caps arm count below this bound
// (2+64 = 66 < 128), so no real `.lang` program can ever trigger this arm
// of checkBranch through syntax.Parse. It exists as defense-in-depth for a
// future relaxation of the parser cap, and is exercised directly by a
// synthetic ast.Program in TestArmBodyLimits (check_branch_test.go), per
// D-10: an unreachable gate must be named as such, not assumed correct.
const maxBlocksPerFunction = 128

// checkBranch lowers a match function whose arms hold full linear bodies
// into a single core.Function that carries BOTH Match (arm identity and
// pattern/edge bookkeeping) and Linear (the flattened operations plus the
// Blocks/Edges the arms lower into). Per this phase's scope decision, a
// match that carries any arm body requires every arm to carry one — bare and
// body arms are never interleaved in the same function — so the native
// switch/case lowering never needs a bare-alternative fallback case inside a
// block-shaped function, and the interpreter/validator dispatch stays a
// simple "every arm has a BlockID" invariant rather than a per-arm union.
func checkBranch(module, functionID, matchID string, function ast.FuncDecl, dataType core.DataType, sealed map[string]bool) (core.Function, []diagnostic.Diagnostic, int) {
	parameterType := coreType(function.Parameter.Type)
	derived, err := ability.DeriveSealed(parameterType, sealed)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}, typeNodeCount(parameterType)
	}
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	typeFact := core.TypeFact{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}
	linear := &core.LinearBody{
		ID:         functionID + ":linear",
		Types:      []core.TypeFact{typeFact},
		Places:     []core.Place{{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}},
		Operations: []core.LinearOperation{},
	}

	work := typeNodeCount(parameterType)
	nextIndex := 0
	seen := make(map[string]bool)
	arms := make([]core.MatchArm, 0, len(function.Body.Arms))
	blocks := make([]core.Block, 0, len(function.Body.Arms)+2)
	edges := make([]core.Edge, 0, len(function.Body.Arms)*2)
	var diagnostics []diagnostic.Diagnostic
	entryBlockID := functionID + ":block:entry"
	joinBlockID := functionID + ":block:join"
	armBlockIDs := make([]string, 0, len(function.Body.Arms))
	armCFGBlocks := make([]cfgBlockSpec, 0, len(function.Body.Arms))
	armEdgeIDs := make(map[string]string, len(function.Body.Arms))

	for index, arm := range function.Body.Arms {
		if seen[arm.Pattern] {
			diagnostics = append(diagnostics, diagnostic.Error("match.subsumed", arm.Span, "alternative is already matched"))
			continue
		}
		if !contains(dataType.Alternatives, arm.Pattern) {
			diagnostics = append(diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
			continue
		}
		if arm.Body == nil {
			diagnostics = append(diagnostics, diagnostic.Error(
				"core.mixed_arm_forms", arm.Span,
				"a match with any arm body requires every arm to carry a body this phase",
			))
			continue
		}
		seen[arm.Pattern] = true
		if len(blocks)+2 > maxBlocksPerFunction {
			diagnostics = append(diagnostics, diagnostic.Error("check.arm_body_limit", arm.Span, "function exceeds the declared block limit"))
			continue
		}

		armBlockID := fmt.Sprintf("%s:block:arm:%d", functionID, index)
		aliasOpID := fmt.Sprintf("%s:op:%d", functionID, nextIndex)
		aliasPointID := fmt.Sprintf("%s:point:linear:%d", functionID, nextIndex)
		aliasPlaceID := fmt.Sprintf("%s:place:%d", functionID, nextIndex+1)
		aliasOp := core.LinearOperation{
			ID: aliasOpID, PointID: aliasPointID, Kind: core.OpCopy, SourceID: parameterID, TargetID: aliasPlaceID, TypeID: typeID,
		}
		linear.Places = append(linear.Places, core.Place{ID: aliasPlaceID, Name: function.Parameter.Name, TypeID: typeID})
		linear.Operations = append(linear.Operations, aliasOp)
		armOperationIDs := []string{aliasOpID}
		armOps := []core.LinearOperation{aliasOp}
		nextIndex++
		work++

		support := analyzeArmBody(functionID, nextIndex, function.Parameter.Name, aliasPlaceID, arm.Body.Span, typeFact, arm.Body)
		if support.Diagnostic != nil {
			diagnostics = append(diagnostics, *support.Diagnostic)
			continue
		}
		linear.Places = append(linear.Places, support.Places...)
		linear.Operations = append(linear.Operations, support.Operations...)
		for _, operation := range support.Operations {
			armOperationIDs = append(armOperationIDs, operation.ID)
		}
		armOps = append(armOps, support.Operations...)
		nextIndex += len(support.Operations)
		work += support.Work

		blocks = append(blocks, core.Block{
			ID: armBlockID, PointID: fmt.Sprintf("%s:point:arm:%d", functionID, index),
			OperationIDs: armOperationIDs, Successors: []string{joinBlockID},
		})
		armBlockIDs = append(armBlockIDs, armBlockID)
		armEdgeToJoinID := fmt.Sprintf("%s:edge:arm:%d:join", functionID, index)
		edges = append(edges,
			core.Edge{ID: fmt.Sprintf("%s:edge:entry:arm:%d", functionID, index), FromBlockID: entryBlockID, ToBlockID: armBlockID, Pattern: arm.Pattern},
			core.Edge{ID: armEdgeToJoinID, FromBlockID: armBlockID, ToBlockID: joinBlockID, Pattern: arm.Pattern},
		)
		armEdgeIDs[armBlockID+"->"+joinBlockID] = armEdgeToJoinID
		armCFGBlocks = append(armCFGBlocks, cfgBlockSpec{id: armBlockID, operations: armOps, successors: []string{joinBlockID}})
		arms = append(arms, core.MatchArm{
			ID: fmt.Sprintf("%s:arm:%d", functionID, index), EdgeID: fmt.Sprintf("%s:edge:%s", matchID, arm.Pattern),
			Pattern: arm.Pattern, BlockID: armBlockID,
		})
	}

	missing := make([]string, 0)
	for _, alternative := range dataType.Alternatives {
		if !seen[alternative] {
			missing = append(missing, alternative)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		causes := make([]diagnostic.Cause, 0, len(missing))
		for _, name := range missing {
			causes = append(causes, diagnostic.Cause{Kind: "missing_alternative", Detail: name})
		}
		diagnostics = append(diagnostics, diagnostic.Error("match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes...))
	}
	if len(diagnostics) > 0 {
		return core.Function{}, diagnostics, work
	}

	blocks = append([]core.Block{{ID: entryBlockID, PointID: functionID + ":point:entry", OperationIDs: []string{}, Successors: armBlockIDs}}, blocks...)
	blocks = append(blocks, core.Block{ID: joinBlockID, PointID: functionID + ":point:return", OperationIDs: []string{}, Successors: []string{}})
	linear.Blocks = blocks
	linear.Edges = edges

	// Backward worklist loan-liveness dataflow (03-03, Q2): the join block
	// carries no operations, so this always converges to a per-arm-block
	// local answer under this phase's topology, but the fixpoint machinery
	// itself is general (see materializeLoanEndpoints's own multi-successor
	// test coverage in check_test.go).
	cfgBlocks := append(append([]cfgBlockSpec(nil), armCFGBlocks...), cfgBlockSpec{id: joinBlockID, successors: nil})
	fixpoint, err := loanLivenessFixpoint(functionID, cfgBlocks)
	if err != nil {
		diagnostics = append(diagnostics, diagnostic.Error("check.cfg_back_edge", function.Body.Span, err.Error()))
		return core.Function{}, diagnostics, work
	}
	edgeIDLookup := func(fromBlockID, toBlockID string) string {
		if id, ok := armEdgeIDs[fromBlockID+"->"+toBlockID]; ok {
			return id
		}
		return fmt.Sprintf("%s:edge:%s:%s", functionID, fromBlockID, toBlockID)
	}
	linear.LoanEndpoints = materializeLoanEndpoints(functionID, cfgBlocks, edgeIDLookup, fixpoint)
	work += fixpoint.work

	return core.Function{
		ID: functionID, Name: function.Name,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:  core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor},
		ReturnType: function.ReturnType.Constructor,
		Match:      &core.Match{ID: matchID, PointID: functionID + ":point:match", Scrutinee: function.Body.Scrutinee, Arms: arms},
		Linear:     linear,
		Span:       function.Span,
	}, nil, work
}

// ---------------------------------------------------------------------
// 03-03: backward worklist loan liveness over the per-function CFG.
//
// This gives checkBranch's arm blocks their LoanEndpoint records (point vs
// edge) via a genuinely backward, block-local-transfer, worklist-to-a-
// fixpoint mechanism (Q2), independent of discoverLoanLastUses (kept only
// as 03-05's future path-oracle building block, per this plan's own
// prohibition against deleting it). Scope note (documented deviation): the
// straight-line path (checkLinear/analyzeStraightLine) is NOT rewired onto
// this pass -- 03-06's already-shipped TestPhase3FieldsAreOmittedWhenAbsent
// requires loan_endpoints stay absent from every Phase 1/2 program's
// serialized core, and every Phase 2 fixture is a straight-line body, so
// populating LoanEndpoints there would move those already-verified bytes.
// discoverLoanLastUses therefore still drives conflict/expiry decisions in
// both analyzeStraightLine and analyzeArmBody (verdicts provably unchanged,
// since neither of those functions is touched by this section); this
// dataflow is the sole producer of the observable core.LoanEndpoint records,
// wired only into checkBranch's arm blocks, where new (not previously
// shipped) fixtures exercise it.
// ---------------------------------------------------------------------

// cfgBlockSpec is the minimal per-block shape the backward worklist
// dataflow consumes: a stable ID, the block's own operations in program
// order, and its successor block IDs (edges out, unlabeled here -- the
// caller reattaches the real edge identity when materializing
// core.LoanEndpoint records).
type cfgBlockSpec struct {
	id         string
	operations []core.LinearOperation
	successors []string
}

// loanLivenessResult is the fixpoint's output: per-block live-in loan sets
// and the total counted work (one unit per transfer-function evaluation,
// one further unit per worklist reinsertion -- D-05/03-03-03).
type loanLivenessResult struct {
	liveIn map[string]map[string]bool
	work   int
}

// placeLoanChain is the O(1)-per-operation derivation this pass builds once
// per function, in a single forward pass over every block's operations
// concatenated in a stable order. A reborrow (`let review = borrow view`)
// is simultaneously a live view of its OWN new loan AND every loan its
// source place already carried -- the same transitive-liveness law
// discoverLoanLastUses' inheritance encodes, here expressed over place IDs
// instead of binding names. Rather than materializing that full ancestry as
// a list per place (an append-copy per operation that reintroduces exactly
// the Θ(N²) blowup this pass replaces, D-02-03/Q2(b)), it is kept as a
// linked chain: latestLoan[placeID] names only the FRESHEST loan a place is
// a view of, and parentLoan[loanID] names the loan immediately BEFORE it in
// the reborrow chain (absent for a loan with no reborrow ancestor). Walking
// the chain is deferred to blockLoanLiveness's own backward scan, which
// short-circuits the moment it reaches an already-recorded loan -- so the
// chain is walked in full at most ONCE per function (each loan visited and
// recorded exactly once across the whole block), bounding total per-block
// work to O(operations), not O(operations × chain depth).
type placeLoanChain struct {
	latestLoan map[string]string
	parentLoan map[string]string
}

func derivePlaceLoans(operations []core.LinearOperation) placeLoanChain {
	chain := placeLoanChain{
		latestLoan: make(map[string]string, len(operations)),
		parentLoan: make(map[string]string, len(operations)),
	}
	for _, operation := range operations {
		inherited := chain.latestLoan[operation.SourceID]
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			chain.parentLoan[operation.LoanID] = inherited
			if operation.TargetID != "" {
				chain.latestLoan[operation.TargetID] = operation.LoanID
			}
			continue
		}
		if inherited != "" && operation.TargetID != "" {
			chain.latestLoan[operation.TargetID] = inherited
		}
	}
	return chain
}

// loanBlockUse is one loan's last reference found within a single block by
// blockLoanLiveness's backward scan.
type loanBlockUse struct {
	loanID         string
	operationIndex int
	operationID    string
}

// blockLoanLiveness is the per-block transfer function: given a block's
// operations in program order, the function-global place->loan chain
// (derivePlaceLoans), and the set of loans already known live on exit
// (liveOut, the union of every successor's live-in set), it walks the
// operations BACKWARD. A reference to a place walks that place's loan chain
// from its freshest loan upward, marking each unrecorded ancestor live and
// recording the reference as its most-recent (first found, walking
// backward) use, stopping as soon as an already-recorded loan is reached
// (the amortized-linear short-circuit). Reaching a loan's own birth
// operation (OpBorrowShared/OpBorrowExclusive) removes it from the live
// set, since nothing earlier in program order can be "after" its creation.
// The returned live set is what must be live entering the block. The
// returned work count is one unit per operation inspected PLUS one unit per
// chain-ancestor step actually walked -- honest per-operation counting
// (D-05), not a flat per-block unit: a reintroduced unbounded chain walk
// (the Θ(N²) shape this pass replaces) would show up here as work growing
// faster than operation count, which TestReborrowChainWorkIsLinear asserts
// directly against.
func blockLoanLiveness(operations []core.LinearOperation, chain placeLoanChain, liveOut map[string]bool) ([]loanBlockUse, map[string]bool, int) {
	live := make(map[string]bool, len(liveOut))
	for loan := range liveOut {
		live[loan] = true
	}
	recorded := make(map[string]bool, len(live))
	var uses []loanBlockUse
	work := 0
	for index := len(operations) - 1; index >= 0; index-- {
		work++ // one unit per operation inspected
		operation := operations[index]
		for loan := chain.latestLoan[operation.SourceID]; loan != "" && !recorded[loan]; loan = chain.parentLoan[loan] {
			work++ // one unit per chain-ancestor step walked
			live[loan] = true
			recorded[loan] = true
			uses = append(uses, loanBlockUse{loanID: loan, operationIndex: index, operationID: operation.ID})
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			if !recorded[operation.LoanID] {
				recorded[operation.LoanID] = true
				uses = append(uses, loanBlockUse{loanID: operation.LoanID, operationIndex: index, operationID: operation.ID})
			}
			delete(live, operation.LoanID)
		}
	}
	return uses, live, work
}

func loanSetsEqual(a, b map[string]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for loan := range a {
		if !b[loan] {
			return false
		}
	}
	return true
}

// loanLivenessFixpoint computes backward monotone dataflow over the finite
// lattice of live loan IDs per block boundary (Q2), iterated with a
// worklist to a fixpoint. blocks must be given in a stable order; every
// successor ID must resolve to a block in the same slice. A cycle (a block
// reachable from itself by following successors) is rejected fail-closed --
// OWN-03 is scoped to acyclic CFGs this phase (T-03-11).
func loanLivenessFixpoint(functionID string, blocks []cfgBlockSpec) (loanLivenessResult, error) {
	byID := make(map[string]cfgBlockSpec, len(blocks))
	order := make([]string, 0, len(blocks))
	var allOps []core.LinearOperation
	for _, block := range blocks {
		byID[block.id] = block
		order = append(order, block.id)
		allOps = append(allOps, block.operations...)
	}
	placeLoan := derivePlaceLoans(allOps)

	const (
		unvisited = 0
		inWork    = 1
		done      = 2
	)
	state := make(map[string]int, len(blocks))
	var walk func(id string) error
	walk = func(id string) error {
		switch state[id] {
		case inWork:
			return fmt.Errorf("check.cfg_back_edge: block %q participates in a cycle", id)
		case done:
			return nil
		}
		state[id] = inWork
		for _, successor := range byID[id].successors {
			if err := walk(successor); err != nil {
				return err
			}
		}
		state[id] = done
		return nil
	}
	for _, id := range order {
		if err := walk(id); err != nil {
			return loanLivenessResult{}, err
		}
	}

	predecessors := make(map[string][]string, len(blocks))
	for _, block := range blocks {
		for _, successor := range block.successors {
			predecessors[successor] = append(predecessors[successor], block.id)
		}
	}

	liveIn := make(map[string]map[string]bool, len(blocks))
	for _, id := range order {
		liveIn[id] = map[string]bool{}
	}

	queue := append([]string(nil), order...)
	queued := make(map[string]bool, len(blocks))
	for _, id := range order {
		queued[id] = true
	}

	work := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		queued[id] = false
		work++ // one transfer-function evaluation

		block := byID[id]
		liveOut := map[string]bool{}
		for _, successor := range block.successors {
			for loan := range liveIn[successor] {
				liveOut[loan] = true
			}
		}
		_, newLiveIn, transferWork := blockLoanLiveness(block.operations, placeLoan, liveOut)
		work += transferWork
		if !loanSetsEqual(liveIn[id], newLiveIn) {
			liveIn[id] = newLiveIn
			for _, predecessor := range predecessors[id] {
				if !queued[predecessor] {
					queue = append(queue, predecessor)
					queued[predecessor] = true
					work++ // one worklist reinsertion
				}
			}
		}
	}
	return loanLivenessResult{liveIn: liveIn, work: work}, nil
}

// materializeLoanEndpoints turns the fixpoint's converged live-in sets into
// core.LoanEndpoint records. Two, mutually exclusive, kinds are produced per
// loan per block:
//
//   - A POINT endpoint where the loan is referenced inside a block and does
//     NOT survive to that block's own live-out (liveOut, the union of every
//     successor's live-in) -- its last reference is genuinely inside this
//     block, so it ends at a point.
//   - An EDGE endpoint only at a genuine successor DIVERGENCE: a block with
//     more than one successor, where the loan is needed by at least one
//     successor (present in liveOut) but NOT by a specific other successor
//     (absent from that successor's own live-in). That is exactly the edge
//     the loan ends on -- the loan is still tracked along the successor(s)
//     that need it (and will receive its own point/edge endpoint further
//     downstream, wherever it is finally consumed), while the diverging
//     edge is where a program on THAT path may safely mutate/move the owner
//     (T-03-06's literal claim, and 03-03-02's fixture-pair falsifier).
//
// A single-successor block therefore never produces an edge endpoint for a
// loan flowing through it unremarked -- there is no divergence to record --
// matching checkBranch's current topology (every arm block has exactly one
// successor, the join), where genuine edge endpoints require a loan born
// before a real fan-out, which today's arm-body lowering does not yet
// produce from real source (D-10; proven possible in the general case by
// TestEdgeSpecificLiveOut).
func materializeLoanEndpoints(functionID string, blocks []cfgBlockSpec, edgeID func(fromBlockID, toBlockID string) string, result loanLivenessResult) []core.LoanEndpoint {
	var placeLoanAll []core.LinearOperation
	for _, block := range blocks {
		placeLoanAll = append(placeLoanAll, block.operations...)
	}
	placeLoan := derivePlaceLoans(placeLoanAll)

	var endpoints []core.LoanEndpoint
	for _, block := range blocks {
		liveOut := map[string]bool{}
		for _, successor := range block.successors {
			for loan := range result.liveIn[successor] {
				liveOut[loan] = true
			}
		}

		if len(block.successors) > 1 {
			for loan := range liveOut {
				for _, successor := range block.successors {
					if result.liveIn[successor][loan] {
						continue
					}
					endpoints = append(endpoints, core.LoanEndpoint{
						ID:     edgeID(block.id, successor) + ":" + loan,
						LoanID: loan, Kind: "edge", EdgeID: edgeID(block.id, successor),
					})
				}
			}
		}

		uses, _, _ := blockLoanLiveness(block.operations, placeLoan, liveOut)
		for _, use := range uses {
			if liveOut[use.loanID] {
				continue // survives past this block; its endpoint lives elsewhere
			}
			endpoints = append(endpoints, core.LoanEndpoint{
				ID:               fmt.Sprintf("%s:point:%s:%d:%s", functionID, block.id, use.operationIndex, use.loanID),
				LoanID:           use.loanID, Kind: "point", BlockID: block.id, AfterOperationID: use.operationID,
			})
		}
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].ID < endpoints[j].ID })
	return endpoints
}

// analyzeArmBody analyzes one match arm's linear body using the same
// straight-line ownership machinery as analyzeStraightLine, but numbers
// every place/operation/point/loan id starting at startIndex within the
// function's single flat Operations list rather than restarting at zero.
// This keeps every arm's operations, once concatenated in arm order,
// satisfying the exact same global ordinal invariant corevalidate already
// enforces for a non-branching linear body (place N is produced by
// operation N-1). The scrutinee is not re-declared here: parameterPlaceID
// names the place the caller already seeded (the implicit per-arm alias
// copy), so every arm's bindings are checked against an alias that only that
// arm can move, and one arm's move can never be observed as a false
// use-after-move by a sibling arm that never runs at the same time (mutually
// exclusive control flow — the two arms' places never collide because their
// IDs are distinct global ordinals). Deliberately a separate function from
// analyzeStraightLine, duplicating rather than reusing it, so the Phase 1/2
// straight-line path (checkLinear) carries zero risk from this addition.
func analyzeArmBody(functionID string, startIndex int, parameterName, parameterPlaceID string, parameterSpan diagnostic.Span, typeFact core.TypeFact, body *ast.LinearBody) ownershipSupport {
	result := ownershipSupport{
		Places: []core.Place{}, Operations: []core.LinearOperation{}, LoanFinalUses: []loanFinalUseFact{}, States: []ownershipStateFact{},
		Work: len(body.Bindings) + 1,
	}
	loanUses := discoverLoanLastUses(parameterName, body)
	if testOnlyForceUniformLoanJoin {
		// Fault-injection seam for TestUniformJoinPlacementFlipsBothVerdicts
		// (03-03-02): force every loan's computed last use to the arm's own
		// join point (len(body.Bindings)), exactly as if edge-specific
		// placement had been deleted and every loan ended uniformly at the
		// join regardless of which edge actually needs it. See the doc
		// comment on testOnlyForceUniformLoanJoin for why this is falsified
		// by direct mutation rather than a production revert.
		for index, use := range loanUses {
			use.index = len(body.Bindings)
			loanUses[index] = use
		}
	}
	for index, binding := range body.Bindings {
		if binding.RHS.Kind == "borrow" || binding.RHS.Kind == "borrow_mut" {
			result.LoanFinalUses = append(result.LoanFinalUses, loanFinalUseFact{
				LoanID: fmt.Sprintf("%s:loan:%d", functionID, startIndex+index), Binding: binding.Name, OperationIndex: startIndex + loanUses[index].index,
			})
		}
	}
	places := map[string]*placeState{
		parameterName: {place: core.Place{ID: parameterPlaceID, Name: parameterName, TypeID: typeFact.ID}, declared: parameterSpan, initialized: true},
	}
	activeLoans := make(map[string]map[string]*loanState)
	expiringLoans := make(map[int][]*loanState)
	endLoans := func(index int) {
		for _, loan := range expiringLoans[index] {
			if loans := activeLoans[loan.ownerID]; loans != nil {
				delete(loans, loan.id)
				if len(loans) == 0 {
					delete(activeLoans, loan.ownerID)
				}
			}
		}
	}
	fail := func(problem diagnostic.Diagnostic) ownershipSupport {
		result.DiagnosticCode = problem.Code
		result.Diagnostic = &problem
		return result
	}
	useAfterMove := func(span diagnostic.Span, state *placeState) diagnostic.Diagnostic {
		causes := []diagnostic.Cause{
			{Kind: "declared_here", Span: spanPointer(state.declared)},
			{Kind: "moved_here", Span: state.movedAt},
			{Kind: "place", Detail: state.place.ID},
			{Kind: "transfer_target", Detail: state.moveTargetID},
			{Kind: "type", Detail: state.place.TypeID},
		}
		return diagnostic.ErrorWithRepairs(
			"ownership.use_after_move", span, "value was used after ownership transferred", causes,
			diagnostic.Repair{Kind: "use_transfer_target", Detail: state.moveTargetID},
			diagnostic.Repair{Kind: "move_use_before_transfer"},
		)
	}
	for index, binding := range body.Bindings {
		result.Work++
		global := startIndex + index
		source, ok := places[binding.RHS.Source]
		if !ok {
			return fail(diagnostic.Error("name.unknown", binding.RHS.Span, "binding source is unknown"))
		}
		if !source.initialized {
			return fail(useAfterMove(binding.RHS.Span, source))
		}
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, global+1), Name: binding.Name, TypeID: source.place.TypeID}
		kind := core.OpCopy
		var loan *loanState
		switch binding.RHS.Kind {
		case "take":
			if loans := activeLoans[source.place.ID]; len(loans) > 0 {
				loanIDs := make([]string, 0, len(loans))
				for loanID := range loans {
					loanIDs = append(loanIDs, loanID)
				}
				sort.Strings(loanIDs)
				blocking := loans[loanIDs[0]]
				causes := []diagnostic.Cause{
					{Kind: "borrow_created_here", Span: spanPointer(blocking.borrowedAt)},
					{Kind: "borrow_used_later", Span: spanPointer(blocking.lastUseSpan)},
					{Kind: "loan", Detail: blocking.id},
					{Kind: "owner", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.move_while_borrowed", binding.RHS.Span, "cannot transfer ownership while a future-used shared loan is live", causes,
					diagnostic.Repair{Kind: "move_after_last_borrow_use"},
				))
			}
			kind = core.OpMove
			source.initialized = false
			source.movedAt = spanPointer(binding.RHS.Span)
			source.moveTargetID = target.ID
		case "borrow":
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			if blocking := conflictingLoan(activeLoans[source.place.ID], "shared"); blocking != nil {
				return fail(borrowConflictDiagnostic(binding.RHS.Span, blocking, source.place.ID, source.place.TypeID))
			}
			kind = core.OpBorrowShared
			use := loanUses[index]
			loan = &loanState{
				id: fmt.Sprintf("%s:loan:%d", functionID, global), ownerID: source.place.ID, access: "shared",
				borrowedAt: binding.RHS.Span, lastUse: use.index, lastUseSpan: use.span,
			}
			if activeLoans[source.place.ID] == nil {
				activeLoans[source.place.ID] = make(map[string]*loanState)
			}
			activeLoans[source.place.ID][loan.id] = loan
			expiringLoans[loan.lastUse] = append(expiringLoans[loan.lastUse], loan)
		case "borrow_mut":
			// An exclusive loan is gated on the same share-ability requirement
			// as a shared loan: the ability that permits observation without
			// duplicating ownership is exactly what an exclusive loan also
			// needs. No source-reachable Phase 1/2/3 type withholds share
			// (see nonShareableTypeFact, check_test.go), so this gate — like
			// the identical one immediately above for shared borrows — is
			// exercised only synthetically today (D-10).
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			if blocking := conflictingLoan(activeLoans[source.place.ID], "exclusive"); blocking != nil {
				return fail(borrowConflictDiagnostic(binding.RHS.Span, blocking, source.place.ID, source.place.TypeID))
			}
			kind = core.OpBorrowExclusive
			use := loanUses[index]
			loan = &loanState{
				id: fmt.Sprintf("%s:loan:%d", functionID, global), ownerID: source.place.ID, access: "exclusive",
				borrowedAt: binding.RHS.Span, lastUse: use.index, lastUseSpan: use.span,
			}
			if activeLoans[source.place.ID] == nil {
				activeLoans[source.place.ID] = make(map[string]*loanState)
			}
			activeLoans[source.place.ID][loan.id] = loan
			expiringLoans[loan.lastUse] = append(expiringLoans[loan.lastUse], loan)
		default:
			if !hasTypeAbility(typeFact, core.AbilityCopy) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityCopy)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.transfer_requires_take", binding.RHS.Span, "noncopyable binding requires explicit take", causes,
					diagnostic.Repair{Kind: "insert_take"},
				))
			}
		}
		result.Places = append(result.Places, target)
		places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true}
		result.Operations = append(result.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, global), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, global),
			Kind: kind, SourceID: source.place.ID, TargetID: target.ID, LoanID: loanID(loan), TypeID: source.place.TypeID,
		})
		endLoans(index)
		result.States = append(result.States, ownershipSnapshot(index, places, activeLoans))
	}
	result.Work++
	returned, ok := places[body.Result]
	if !ok {
		return fail(diagnostic.Error("name.unknown", body.Span, "linear result is unknown"))
	}
	if !returned.initialized {
		return fail(useAfterMove(body.Span, returned))
	}
	global := startIndex + len(body.Bindings)
	result.Operations = append(result.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, global), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, global),
		Kind: core.OpReturn, SourceID: returned.place.ID, TypeID: returned.place.TypeID,
	})
	// A straight-line body's Return is always the last operation, so its
	// missing target never leaves a gap in the flat Places array (nothing
	// after it needs a place index). A branch's Return is NOT the last
	// operation in the whole function's flat list — the next arm's own
	// operations follow it — so corevalidate's place-order invariant
	// (place N is produced by operation N-1) would otherwise desync at
	// exactly this point. Padding with one unreferenced place per arm's
	// Return keeps every operation, including Return, consuming exactly
	// one place-index slot, restoring the same array-position coupling
	// analyzeStraightLine's callers already rely on.
	result.Places = append(result.Places, core.Place{
		ID: fmt.Sprintf("%s:place:%d", functionID, global+1), Name: "_", TypeID: returned.place.TypeID,
	})
	endLoans(len(body.Bindings))
	result.States = append(result.States, ownershipSnapshot(len(body.Bindings), places, activeLoans))
	return result
}

func checkLinear(module, functionID string, function ast.FuncDecl) (core.Function, []diagnostic.Diagnostic, int) {
	parameterType := coreType(function.Parameter.Type)
	if !sameType(function.ReturnType, function.Parameter.Type) {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.return_mismatch", function.Span, "linear result must have the parameter type")}, typeNodeCount(parameterType)
	}
	// OWN-04: the borrowed-view return case relaxes nothing about type
	// identity above (the underlying type must still match the parameter
	// type) — it adds a new legal declaration alongside the unchanged
	// ordinary case, resolved on the syntactic discriminant
	// (function.ReturnOrigin != nil) rather than on the return type's shape.
	// A borrowed return with a path other than the function's own single
	// parameter is a causal, span-bearing rejection: this reduced language
	// has one parameter per function and no field-path-bearing executable
	// shape, so any other path can never be honest.
	if function.ReturnOrigin != nil && function.ReturnOrigin.Path != function.Parameter.Name {
		causes := []diagnostic.Cause{
			{Kind: "declared_path", Detail: function.ReturnOrigin.Path},
			{Kind: "only_legal_path", Detail: function.Parameter.Name},
		}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error(
			"origin.unknown_path", function.ReturnOrigin.Span, "borrowed-view origin path must name the function's own parameter", causes...,
		)}, typeNodeCount(parameterType)
	}
	derived, err := ability.Derive(parameterType)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}, typeNodeCount(parameterType)
	}
	typeID := functionID + ":type:0"
	if !executableShape(parameterType) {
		// Ability derivation above already ran and produced facts for this
		// shape (derived.Granted/derived.NegativeWitnesses) — this gate
		// closes D-02-09 by refusing EXECUTION admission, not by
		// withholding ability derivation, so the shape stays available for
		// ability evidence (see TestAbilityFactsSurviveExecutionRejection).
		// Only Byte and Buffer have a native lowering this phase
		// (cgen.linearInput); every other shape (Box, Pair, or a nominal
		// data type used in a straight-line body) type-checks and derives
		// abilities cleanly but cannot be run by any engine, so it must be
		// refused here with a causal span rather than dying spanless
		// downstream on every engine (D-07).
		causes := []diagnostic.Cause{
			{Kind: "type", Detail: typeID},
			{Kind: "constructor", Detail: parameterType.Constructor},
		}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
			"check.unexecutable_shape", function.Parameter.Span,
			"parameter shape has no native execution lowering this phase", causes,
			diagnostic.Repair{Kind: "use_executable_shape", Detail: "Byte or Buffer"},
		)}, typeNodeCount(parameterType)
	}
	parameterID := functionID + ":place:0"
	linear := &core.LinearBody{
		ID:         functionID + ":linear",
		Types:      []core.TypeFact{{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}},
		Places:     []core.Place{{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}},
		Operations: []core.LinearOperation{},
	}
	support := analyzeStraightLine(functionID, function.Parameter.Name, function.Parameter.Span, linear.Types[0], function.Body.Linear)
	if support.Diagnostic != nil {
		return core.Function{}, []diagnostic.Diagnostic{*support.Diagnostic}, support.Work
	}
	linear.Places = support.Places
	linear.Operations = support.Operations
	var publicOrigin *core.PublicOrigin
	if function.ReturnOrigin != nil {
		publicOrigin = &core.PublicOrigin{Paths: []string{function.ReturnOrigin.Path}, Access: function.ReturnOrigin.Access}
	}
	return core.Function{
		ID: functionID, Name: function.Name, EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter: core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor}, ReturnType: function.ReturnType.Constructor,
		Linear: linear, PublicOrigin: publicOrigin, Span: function.Span,
	}, nil, support.Work
}

// ---------------------------------------------------------------------
// Phase 4: fallible foreign call admission (D-04-01/D-04-02/D-04-04/D-04-05).
//
// This section is deliberately a separate, narrow path rather than a
// generalization of checkLinear/analyzeStraightLine: it handles exactly the
// one shape this plan's tracer proves -- a straight-line function whose sole
// binding is `try <foreign symbol>(<parameter>)`, immediately returned. A
// richer shape (ordinary bindings before or after the call, multiple calls)
// is out of scope this plan and is refused with a named diagnostic rather
// than silently mishandled. Keeping this fully separate from checkLinear
// means the Phase 1-3 straight-line path carries zero risk from this
// addition (D-04-23's byte-identity requirement), exactly as checkBranch
// stayed separate from analyzeStraightLine in Phase 3.
// ---------------------------------------------------------------------

// maxForeignBlocksPerProgram, maxForeignSymbolsPerBlock, and
// maxForeignParametersPerSymbol bound T-04-05's foreign declaration surface,
// derived from the existing declaration/function caps (syntax.go's
// maxDeclarations family) rather than an arbitrary round number: a program
// already cannot declare more than a few thousand top-level items, so a
// foreign surface bounded well below that is fail-closed, not merely
// advisory.
const (
	maxForeignBlocksPerProgram     = 64
	maxForeignSymbolsPerBlockCheck = 64
	maxForeignParametersPerSymbol  = 1
)

// foreignSymbolInfo is check.go's own resolved view of one declared foreign
// symbol: the policy keys collectForeignSymbols found, by name, so
// checkFallibleLinear (and, in a later task, the admission gate refusing a
// missing unwind/nonlocal_exit policy) can ask "was this key present"
// without re-scanning ast.ForeignPolicy each time.
type foreignSymbolInfo struct {
	Name         string
	Parameter    ast.Parameter
	ReturnType   ast.TypeRef
	Allocator    string
	HasAllocator bool
	Unwind       string
	HasUnwind    bool
	NonlocalExit string
	HasNonlocal  bool
	Fails        string
	HasFails     bool
	Span         diagnostic.Span
}

// collectForeignSymbols builds the module-wide foreign symbol table from
// every declared `foreign C {}` block, applying T-04-05's caps fail-closed.
// It does not refuse a symbol for a missing unwind/nonlocal_exit policy --
// that admission gate is checkFallibleLinear's job (D-04-16), fired only for
// a symbol an actual call resolves to, so a declared-but-never-called
// under-specified symbol does not block an unrelated program.
func collectForeignSymbols(program ast.Program) (map[string]foreignSymbolInfo, []diagnostic.Diagnostic) {
	if len(program.Foreign) > maxForeignBlocksPerProgram {
		return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_block_limit", program.Foreign[0].Span, "program exceeds the declared foreign block limit")}
	}
	symbols := make(map[string]foreignSymbolInfo)
	for _, block := range program.Foreign {
		if len(block.Symbols) > maxForeignSymbolsPerBlockCheck {
			return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_symbol_limit", block.Span, "foreign block exceeds the declared symbol limit")}
		}
		for _, symbol := range block.Symbols {
			if len(symbol.Policies) > maxForeignPoliciesPerSymbolCheck {
				return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_policy_limit", symbol.Span, "foreign symbol exceeds the declared policy limit")}
			}
			info := foreignSymbolInfo{Name: symbol.Name, Parameter: symbol.Parameter, ReturnType: symbol.ReturnType, Span: symbol.Span}
			for _, policy := range symbol.Policies {
				switch policy.Key {
				case "unwind":
					info.Unwind, info.HasUnwind = policy.Value, true
				case "nonlocal_exit":
					info.NonlocalExit, info.HasNonlocal = policy.Value, true
				case "allocator":
					info.Allocator, info.HasAllocator = policy.Value, true
				case "fails":
					info.Fails, info.HasFails = policy.Value, true
				}
			}
			symbols[symbol.Name] = info
		}
	}
	return symbols, nil
}

// maxForeignPoliciesPerSymbolCheck mirrors the parser's own
// maxForeignPoliciesPerSymbol cap (syntax package): declared again here,
// independently, rather than imported, so check.go's own admission surface
// is bounded even if a future caller constructs an ast.Program directly
// (bypassing the parser).
const maxForeignPoliciesPerSymbolCheck = 32

// hasTryCall reports whether a linear body's bindings contain a fallible
// foreign call consumer (D-04-06's `try` or `discard ... because` forms). A
// function with no such binding takes the entirely unchanged checkLinear
// path.
func hasTryCall(body *ast.LinearBody) bool {
	for _, binding := range body.Bindings {
		if binding.RHS.Kind == "try_call" || binding.RHS.Kind == "discard_call" {
			return true
		}
	}
	return false
}

// checkFallibleLinear dispatches between the two shapes this phase supports.
// The tracer shape (04-01) is a single `try` binding immediately returned.
// The resource-lifecycle shape (04-02, D-04-07) is a sequence of one or more
// try/discard foreign-call bindings whose result is the function's own
// parameter -- every acquired resource is released before return, so the
// parameter (never moved by a foreign call) is always what comes back.
func checkFallibleLinear(functionID string, function ast.FuncDecl, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (core.Function, []diagnostic.Diagnostic, int) {
	parameterType := coreType(function.Parameter.Type)
	if !sameType(function.ReturnType, function.Parameter.Type) {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.return_mismatch", function.Span, "linear result must have the parameter type")}, typeNodeCount(parameterType)
	}
	derived, err := ability.Derive(parameterType)
	if err != nil {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.unknown", function.Parameter.Span, err.Error())}, typeNodeCount(parameterType)
	}
	typeID := functionID + ":type:0"
	work := typeNodeCount(parameterType) + 1
	if !executableShape(parameterType) {
		causes := []diagnostic.Cause{{Kind: "type", Detail: typeID}, {Kind: "constructor", Detail: parameterType.Constructor}}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
			"check.unexecutable_shape", function.Parameter.Span,
			"parameter shape has no native execution lowering this phase", causes,
			diagnostic.Repair{Kind: "use_executable_shape", Detail: "Byte or Buffer"},
		)}, work
	}

	body := function.Body.Linear
	if len(body.Bindings) == 1 && body.Bindings[0].RHS.Kind == "try_call" && body.Result == body.Bindings[0].Name {
		return checkForeignTracer(functionID, function, parameterType, derived, typeID, work, body.Bindings[0], foreignSymbols, functionNames, dataTypes)
	}
	if len(body.Bindings) > 0 && body.Result == function.Parameter.Name && everyBindingIsFallible(body.Bindings) {
		return checkResourceLifecycle(functionID, function, parameterType, derived, typeID, work, foreignSymbols, functionNames, dataTypes)
	}
	return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error(
		"check.foreign_call_shape_unsupported", body.Span,
		"this phase supports only a single fallible foreign call immediately returned, or a sequence of try/discard foreign calls whose result is the function's own parameter",
	)}, work
}

func everyBindingIsFallible(bindings []ast.Binding) bool {
	for _, binding := range bindings {
		if binding.RHS.Kind != "try_call" && binding.RHS.Kind != "discard_call" {
			return false
		}
	}
	return true
}

// resolveForeignStep independently resolves one step's declared foreign
// symbol and its failure-ADT type fact, shared by both checkForeignTracer and
// checkResourceLifecycle so the two admission gates (D-04-02/D-04-16) and the
// argument-shape rule stay in exact lockstep between the tracer's single-call
// shape and the resource-lifecycle chain.
func resolveForeignStep(binding ast.Binding, functionParameterName string, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (foreignSymbolInfo, core.TypeRef, ability.Result, *diagnostic.Diagnostic) {
	if len(binding.RHS.Arguments) != maxForeignParametersPerSymbol || binding.RHS.Arguments[0] != functionParameterName {
		problem := diagnostic.Error("name.unknown", binding.RHS.Span, "foreign call argument must be the function's own parameter")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	symbol, isForeign := foreignSymbols[binding.RHS.Callee]
	if !isForeign {
		if functionNames[binding.RHS.Callee] {
			causes := []diagnostic.Cause{{Kind: "callee", Detail: binding.RHS.Callee}}
			problem := diagnostic.ErrorWithRepairs(
				"core.call_target_not_foreign", binding.RHS.Span,
				"a fallible call's target must be a declared foreign symbol, not a Lang function", causes,
				diagnostic.Repair{Kind: "declare_foreign_symbol"},
			)
			return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
		}
		problem := diagnostic.Error("name.unknown", binding.RHS.Span, "foreign call target is unknown")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	if problem := missingForeignPolicyDiagnostic(symbol); problem != nil {
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, problem
	}
	errDataType, ok := dataTypes[symbol.Fails]
	if !symbol.HasFails || !ok || len(errDataType.Alternatives) == 0 {
		problem := diagnostic.Error("type.unknown", binding.Span, "foreign symbol's declared failure type is unknown or has no alternatives")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	errShape := core.TypeRef{Constructor: errDataType.Name}
	errDerived, derr := ability.DeriveSealed(errShape, map[string]bool{errDataType.Name: true})
	if derr != nil {
		problem := diagnostic.Error("type.unknown", binding.Span, derr.Error())
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	return symbol, errShape, errDerived, nil
}

// checkForeignTracer lowers the 04-01 shape: a straight-line function whose
// sole binding is a fallible foreign call, immediately returned on the ok
// edge. It produces a core.Function whose Linear body carries three blocks
// (entry/ok/err) and two edges, exactly mirroring checkBranch's block/edge
// shape but keyed on a fallible call rather than a match arm.
func checkForeignTracer(functionID string, function ast.FuncDecl, parameterType core.TypeRef, derived ability.Result, typeID string, work int, tryBinding ast.Binding, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (core.Function, []diagnostic.Diagnostic, int) {
	symbol, errShape, errDerived, problem := resolveForeignStep(tryBinding, function.Parameter.Name, foreignSymbols, functionNames, dataTypes)
	if problem != nil {
		return core.Function{}, []diagnostic.Diagnostic{*problem}, work
	}

	parameterID := functionID + ":place:0"
	okPlaceID := functionID + ":place:1"
	errPlaceID := functionID + ":place:2"
	errTypeID := functionID + ":type:1"
	entryBlockID := functionID + ":block:entry"
	okBlockID := functionID + ":block:ok"
	errBlockID := functionID + ":block:err"
	okEdgeID := functionID + ":edge:entry:ok"
	errEdgeID := functionID + ":edge:entry:err"
	callOpID := functionID + ":op:0"
	returnOpID := functionID + ":op:1"
	failOpID := functionID + ":op:2"

	linear := &core.LinearBody{
		ID: functionID + ":linear",
		Types: []core.TypeFact{
			{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses},
			{ID: errTypeID, Shape: errShape, Abilities: errDerived.Granted, NegativeWitnesses: errDerived.NegativeWitnesses},
		},
		Places: []core.Place{
			{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID},
			{ID: okPlaceID, Name: tryBinding.Name, TypeID: typeID},
			{ID: errPlaceID, Name: "_err", TypeID: errTypeID},
		},
		Operations: []core.LinearOperation{
			{
				ID: callOpID, PointID: functionID + ":point:linear:0", Kind: core.OpForeignCall,
				SourceID: parameterID, TargetID: okPlaceID, TypeID: typeID,
				OkEdgeID: okEdgeID, ErrEdgeID: errEdgeID, ErrTargetID: errPlaceID,
			},
			{ID: returnOpID, PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: okPlaceID, TypeID: typeID},
			{ID: failOpID, PointID: functionID + ":point:linear:2", Kind: core.OpFail, SourceID: errPlaceID, TypeID: errTypeID},
		},
		Blocks: []core.Block{
			{ID: entryBlockID, PointID: functionID + ":point:entry", OperationIDs: []string{callOpID}, Successors: []string{okBlockID, errBlockID}},
			{ID: okBlockID, PointID: functionID + ":point:ok", OperationIDs: []string{returnOpID}},
			{ID: errBlockID, PointID: functionID + ":point:err", OperationIDs: []string{failOpID}},
		},
		Edges: []core.Edge{
			{ID: okEdgeID, FromBlockID: entryBlockID, ToBlockID: okBlockID, Pattern: "ok"},
			{ID: errEdgeID, FromBlockID: entryBlockID, ToBlockID: errBlockID, Pattern: "err"},
		},
	}

	contract := &core.ForeignContract{
		Symbol: symbol.Name, Allocator: symbol.Allocator, Unwind: symbol.Unwind, NonlocalExit: symbol.NonlocalExit, Fails: symbol.Fails,
	}

	return core.Function{
		ID: functionID, Name: function.Name,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:       core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor},
		ReturnType:      function.ReturnType.Constructor,
		Linear:          linear,
		ForeignContract: contract,
		Span:            function.Span,
	}, nil, work
}

// resourceStep is check.go's own resolved bookkeeping for one step of a
// resource-lifecycle function: the block/place/type identity a step was
// assigned, plus the information the (D-04-07) reverse-order release
// materialization needs once the step completes.
type resourceStep struct {
	kind       string // "try_call" | "discard_call"
	symbol     foreignSymbolInfo
	callOpID   string
	okPlaceID  string
	errPlaceID string
	errTypeID  string
	blockID    string
	okEdgeID   string
	errEdgeID  string
}

// checkResourceLifecycle lowers a sequence of N try/discard foreign-call
// bindings (D-04-07/D-04-06) into N single-operation call blocks, one
// terminal err block per try_call step (releasing every completed try_call
// acquisition that precedes it, reverse of completion order), and one
// terminal success block (releasing every completed try_call acquisition, in
// full reverse order, then returning the function's own parameter -- never
// moved by any OpForeignCall, so it is always available to return once every
// acquired resource has been released).
//
// Order is decided by this file's own forward accumulation over the step
// sequence (mirroring the `expiringLoans[loan.lastUse] = append(...)` shape
// already established for loans in this file), never by map iteration: a
// release list derived from the emitter's own control-flow structure rather
// than from this materialized list is exactly the defect the three-
// acquisition fixture and the transposition mutation (Task 04-02-03) exist to
// catch.
func checkResourceLifecycle(functionID string, function ast.FuncDecl, parameterType core.TypeRef, derived ability.Result, typeID string, work int, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (core.Function, []diagnostic.Diagnostic, int) {
	steps := function.Body.Linear.Bindings
	n := len(steps)
	parameterID := functionID + ":place:0"

	types := []core.TypeFact{{ID: typeID, Shape: parameterType, Abilities: derived.Granted, NegativeWitnesses: derived.NegativeWitnesses}}
	places := make([]core.Place, 1+n, 1+2*n)
	places[0] = core.Place{ID: parameterID, Name: function.Parameter.Name, TypeID: typeID}
	infos := make([]resourceStep, n)
	typeCounter := 1

	// corevalidate's targetMatches requires an OpForeignCall at flat
	// operation index i to produce place:(i+1) EXACTLY (the same
	// index-plus-one invariant every other transition operation obeys) --
	// so every step's ok place is assigned first, contiguously, in call
	// order (place:1..place:n), and every step's err place (never
	// index-checked -- only referenced by ErrTargetID) is appended
	// afterward, in step order, at place:(n+1)..place:(2n).
	for i, binding := range steps {
		symbol, errShape, errDerived, problem := resolveForeignStep(binding, function.Parameter.Name, foreignSymbols, functionNames, dataTypes)
		if problem != nil {
			return core.Function{}, []diagnostic.Diagnostic{*problem}, work
		}
		errTypeID := fmt.Sprintf("%s:type:%d", functionID, typeCounter)
		typeCounter++
		types = append(types, core.TypeFact{ID: errTypeID, Shape: errShape, Abilities: errDerived.Granted, NegativeWitnesses: errDerived.NegativeWitnesses})

		okPlaceID := fmt.Sprintf("%s:place:%d", functionID, i+1)
		errPlaceID := fmt.Sprintf("%s:place:%d", functionID, n+1+i)
		places[i+1] = core.Place{ID: okPlaceID, Name: binding.Name, TypeID: typeID}
		places = append(places, core.Place{ID: errPlaceID, Name: fmt.Sprintf("_err%d", i), TypeID: errTypeID})

		blockID := functionID + ":block:entry"
		if i > 0 {
			blockID = fmt.Sprintf("%s:block:step:%d", functionID, i)
		}
		infos[i] = resourceStep{
			kind: binding.RHS.Kind, symbol: symbol, callOpID: fmt.Sprintf("%s:op:%d", functionID, i),
			okPlaceID: okPlaceID, errPlaceID: errPlaceID, errTypeID: errTypeID, blockID: blockID,
		}
		work++
	}

	successBlockID := functionID + ":block:success"
	var blocks []core.Block
	var edges []core.Edge
	operations := make([]core.LinearOperation, n)
	for i := 0; i < n; i++ {
		okTarget := successBlockID
		if i+1 < n {
			okTarget = infos[i+1].blockID
		}
		errTarget := okTarget
		if infos[i].kind == "try_call" {
			errTarget = fmt.Sprintf("%s:block:err:%d", functionID, i)
		}
		infos[i].okEdgeID = fmt.Sprintf("%s:edge:step:%d:ok", functionID, i)
		infos[i].errEdgeID = fmt.Sprintf("%s:edge:step:%d:err", functionID, i)
		successors := []string{okTarget}
		if errTarget != okTarget {
			successors = append(successors, errTarget)
		}
		pointID := fmt.Sprintf("%s:point:step:%d", functionID, i)
		if i == 0 {
			// pathoracle/originvalidate require the entry block's PointID to
			// equal the function's own EntryPointID exactly (the same
			// convention checkForeignTracer and checkBranch already use).
			pointID = functionID + ":point:entry"
		}
		blocks = append(blocks, core.Block{
			ID: infos[i].blockID, PointID: pointID,
			OperationIDs: []string{infos[i].callOpID}, Successors: successors,
		})
		edges = append(edges,
			core.Edge{ID: infos[i].okEdgeID, FromBlockID: infos[i].blockID, ToBlockID: okTarget, Pattern: "ok"},
			core.Edge{ID: infos[i].errEdgeID, FromBlockID: infos[i].blockID, ToBlockID: errTarget, Pattern: "err"},
		)
		operations[i] = core.LinearOperation{
			ID: infos[i].callOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, i), Kind: core.OpForeignCall,
			SourceID: parameterID, TargetID: infos[i].okPlaceID, TypeID: typeID,
			OkEdgeID: infos[i].okEdgeID, ErrEdgeID: infos[i].errEdgeID, ErrTargetID: infos[i].errPlaceID,
		}
	}

	// completed accumulates try_call steps in completion order as the
	// forward walk below reaches each one -- the single materialization
	// point D-04-07 requires. discard_call steps are never appended: their
	// acquired resource is not tracked for release this plan (a documented
	// narrowing -- see the plan's flagged_assumptions).
	var completed []resourceStep
	releaseOps := func(from []resourceStep) []core.LinearOperation {
		ops := make([]core.LinearOperation, 0, len(from))
		for j := len(from) - 1; j >= 0; j-- {
			acquired := from[j]
			opID := fmt.Sprintf("%s:op:%d", functionID, len(operations)+len(ops))
			ops = append(ops, core.LinearOperation{
				ID: opID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, len(operations)+len(ops)),
				Kind: core.OpRelease, SourceID: acquired.okPlaceID, TypeID: typeID, ReleasesOperationID: acquired.callOpID,
			})
		}
		return ops
	}

	for i := 0; i < n; i++ {
		if infos[i].kind != "try_call" {
			continue
		}
		errOps := releaseOps(completed)
		failOpIndex := len(operations) + len(errOps)
		failOpID := fmt.Sprintf("%s:op:%d", functionID, failOpIndex)
		errOps = append(errOps, core.LinearOperation{
			ID: failOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, failOpIndex),
			Kind: core.OpFail, SourceID: infos[i].errPlaceID, TypeID: infos[i].errTypeID,
		})
		errOpIDs := make([]string, len(errOps))
		for idx, op := range errOps {
			errOpIDs[idx] = op.ID
		}
		blocks = append(blocks, core.Block{
			ID: fmt.Sprintf("%s:block:err:%d", functionID, i), PointID: fmt.Sprintf("%s:point:err:%d", functionID, i), OperationIDs: errOpIDs,
		})
		operations = append(operations, errOps...)
		completed = append(completed, infos[i])
	}

	successOps := releaseOps(completed)
	returnOpIndex := len(operations) + len(successOps)
	returnOpID := fmt.Sprintf("%s:op:%d", functionID, returnOpIndex)
	successOps = append(successOps, core.LinearOperation{
		ID: returnOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, returnOpIndex),
		Kind: core.OpReturn, SourceID: parameterID, TypeID: typeID,
	})
	successOpIDs := make([]string, len(successOps))
	for idx, op := range successOps {
		successOpIDs[idx] = op.ID
	}
	blocks = append(blocks, core.Block{ID: successBlockID, PointID: functionID + ":point:success", OperationIDs: successOpIDs})
	operations = append(operations, successOps...)

	linear := &core.LinearBody{
		ID: functionID + ":linear", Types: types, Places: places, Operations: operations, Blocks: blocks, Edges: edges,
	}
	first := infos[0].symbol
	contract := &core.ForeignContract{Symbol: first.Name, Allocator: first.Allocator, Unwind: first.Unwind, NonlocalExit: first.NonlocalExit, Fails: first.Fails}

	return core.Function{
		ID: functionID, Name: function.Name,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:       core.Parameter{ID: parameterID, Name: function.Parameter.Name, Type: parameterType.Constructor},
		ReturnType:      function.ReturnType.Constructor,
		Linear:          linear,
		ForeignContract: contract,
		Span:            function.Span,
	}, nil, work
}

// missingForeignPolicyDiagnostic is Task 4's admission gate stub (D-04-16):
// a foreign declaration missing its unwind or nonlocal_exit policy is
// refused, independently, by both check and corevalidate, with no default
// value. It is implemented in Task 4; this plan's tracer fixture always
// supplies both policies, so this always returns nil today.
func missingForeignPolicyDiagnostic(symbol foreignSymbolInfo) *diagnostic.Diagnostic {
	missing := ""
	switch {
	case !symbol.HasUnwind:
		missing = "unwind"
	case !symbol.HasNonlocal:
		missing = "nonlocal_exit"
	default:
		return nil
	}
	causes := []diagnostic.Cause{
		{Kind: "foreign_symbol", Detail: symbol.Name},
		{Kind: "missing_policy", Detail: missing},
	}
	problem := diagnostic.ErrorWithRepairs(
		"foreign.unwind_policy_undeclared", symbol.Span,
		"a foreign symbol must declare both an unwind and a nonlocal_exit policy, with no default", causes,
		diagnostic.Repair{Kind: "declare_unwind_policy"},
	)
	return &problem
}

type loanFinalUseFact struct {
	LoanID         string
	Binding        string
	OperationIndex int
}

type ownershipStateFact struct {
	OperationIndex    int
	InitializedPlaces []string
	ActiveLoans       []string
}

type ownershipSupport struct {
	Places         []core.Place
	Operations     []core.LinearOperation
	LoanFinalUses  []loanFinalUseFact
	States         []ownershipStateFact
	Work           int
	DiagnosticCode string
	Diagnostic     *diagnostic.Diagnostic
}

type placeState struct {
	place        core.Place
	declared     diagnostic.Span
	initialized  bool
	movedAt      *diagnostic.Span
	moveTargetID string
}

func analyzeStraightLine(functionID, parameterName string, parameterSpan diagnostic.Span, typeFact core.TypeFact, body *ast.LinearBody) ownershipSupport {
	parameterID := functionID + ":place:0"
	result := ownershipSupport{
		Places:     []core.Place{{ID: parameterID, Name: parameterName, TypeID: typeFact.ID}},
		Operations: []core.LinearOperation{}, LoanFinalUses: []loanFinalUseFact{}, States: []ownershipStateFact{},
		Work: typeNodeCount(typeFact.Shape) + len(body.Bindings) + 1,
	}
	loanUses := discoverLoanLastUses(parameterName, body)
	for index, binding := range body.Bindings {
		if binding.RHS.Kind == "borrow" || binding.RHS.Kind == "borrow_mut" {
			result.LoanFinalUses = append(result.LoanFinalUses, loanFinalUseFact{
				LoanID: fmt.Sprintf("%s:loan:%d", functionID, index), Binding: binding.Name, OperationIndex: loanUses[index].index,
			})
		}
	}
	places := map[string]*placeState{
		parameterName: {place: result.Places[0], declared: parameterSpan, initialized: true},
	}
	activeLoans := make(map[string]map[string]*loanState)
	expiringLoans := make(map[int][]*loanState)
	endLoans := func(index int) {
		for _, loan := range expiringLoans[index] {
			if loans := activeLoans[loan.ownerID]; loans != nil {
				delete(loans, loan.id)
				if len(loans) == 0 {
					delete(activeLoans, loan.ownerID)
				}
			}
		}
	}
	fail := func(problem diagnostic.Diagnostic) ownershipSupport {
		result.DiagnosticCode = problem.Code
		result.Diagnostic = &problem
		return result
	}
	useAfterMove := func(span diagnostic.Span, state *placeState) diagnostic.Diagnostic {
		causes := []diagnostic.Cause{
			{Kind: "declared_here", Span: spanPointer(state.declared)},
			{Kind: "moved_here", Span: state.movedAt},
			{Kind: "place", Detail: state.place.ID},
			{Kind: "transfer_target", Detail: state.moveTargetID},
			{Kind: "type", Detail: state.place.TypeID},
		}
		return diagnostic.ErrorWithRepairs(
			"ownership.use_after_move", span, "value was used after ownership transferred", causes,
			diagnostic.Repair{Kind: "use_transfer_target", Detail: state.moveTargetID},
			diagnostic.Repair{Kind: "move_use_before_transfer"},
		)
	}
	for index, binding := range body.Bindings {
		result.Work++
		source, ok := places[binding.RHS.Source]
		if !ok {
			return fail(diagnostic.Error("name.unknown", binding.RHS.Span, "binding source is unknown"))
		}
		if !source.initialized {
			return fail(useAfterMove(binding.RHS.Span, source))
		}
		target := core.Place{ID: fmt.Sprintf("%s:place:%d", functionID, index+1), Name: binding.Name, TypeID: source.place.TypeID}
		kind := core.OpCopy
		var loan *loanState
		switch binding.RHS.Kind {
		case "take":
			if loans := activeLoans[source.place.ID]; len(loans) > 0 {
				loanIDs := make([]string, 0, len(loans))
				for loanID := range loans {
					loanIDs = append(loanIDs, loanID)
				}
				sort.Strings(loanIDs)
				blocking := loans[loanIDs[0]]
				causes := []diagnostic.Cause{
					{Kind: "borrow_created_here", Span: spanPointer(blocking.borrowedAt)},
					{Kind: "borrow_used_later", Span: spanPointer(blocking.lastUseSpan)},
					{Kind: "loan", Detail: blocking.id},
					{Kind: "owner", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.move_while_borrowed", binding.RHS.Span, "cannot transfer ownership while a future-used shared loan is live", causes,
					diagnostic.Repair{Kind: "move_after_last_borrow_use"},
				))
			}
			kind = core.OpMove
			source.initialized = false
			source.movedAt = spanPointer(binding.RHS.Span)
			source.moveTargetID = target.ID
		case "borrow":
			// A shared loan requires the share ability, exactly as the
			// implicit-copy path below requires copy. Without this gate the
			// checker emits core.OpBorrowShared that the source-blind
			// validator rejects as core.ability.share_denied, turning an
			// invalid program into a spanless operational failure.
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			if blocking := conflictingLoan(activeLoans[source.place.ID], "shared"); blocking != nil {
				return fail(borrowConflictDiagnostic(binding.RHS.Span, blocking, source.place.ID, source.place.TypeID))
			}
			kind = core.OpBorrowShared
			use := loanUses[index]
			loan = &loanState{
				id: fmt.Sprintf("%s:loan:%d", functionID, index), ownerID: source.place.ID, access: "shared",
				borrowedAt: binding.RHS.Span, lastUse: use.index, lastUseSpan: use.span,
			}
			if activeLoans[source.place.ID] == nil {
				activeLoans[source.place.ID] = make(map[string]*loanState)
			}
			activeLoans[source.place.ID][loan.id] = loan
			expiringLoans[loan.lastUse] = append(expiringLoans[loan.lastUse], loan)
		case "borrow_mut":
			// An exclusive loan is gated on the same share-ability requirement
			// as a shared loan (see the comment on the identical gate above).
			if !hasTypeAbility(typeFact, core.AbilityShare) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityShare)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.borrow_requires_share", binding.RHS.Span, "type does not grant the share ability required to borrow", causes,
					diagnostic.Repair{Kind: "use_take_instead"},
				))
			}
			if blocking := conflictingLoan(activeLoans[source.place.ID], "exclusive"); blocking != nil {
				return fail(borrowConflictDiagnostic(binding.RHS.Span, blocking, source.place.ID, source.place.TypeID))
			}
			kind = core.OpBorrowExclusive
			use := loanUses[index]
			loan = &loanState{
				id: fmt.Sprintf("%s:loan:%d", functionID, index), ownerID: source.place.ID, access: "exclusive",
				borrowedAt: binding.RHS.Span, lastUse: use.index, lastUseSpan: use.span,
			}
			if activeLoans[source.place.ID] == nil {
				activeLoans[source.place.ID] = make(map[string]*loanState)
			}
			activeLoans[source.place.ID][loan.id] = loan
			expiringLoans[loan.lastUse] = append(expiringLoans[loan.lastUse], loan)
		default:
			if !hasTypeAbility(typeFact, core.AbilityCopy) {
				causes := []diagnostic.Cause{
					{Kind: "declared_here", Span: spanPointer(source.declared)},
					{Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityCopy)},
					{Kind: "place", Detail: source.place.ID},
					{Kind: "type", Detail: source.place.TypeID},
				}
				return fail(diagnostic.ErrorWithRepairs(
					"ownership.transfer_requires_take", binding.RHS.Span, "noncopyable binding requires explicit take", causes,
					diagnostic.Repair{Kind: "insert_take"},
				))
			}
		}
		result.Places = append(result.Places, target)
		places[binding.Name] = &placeState{place: target, declared: binding.Span, initialized: true}
		result.Operations = append(result.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: kind, SourceID: source.place.ID, TargetID: target.ID, LoanID: loanID(loan), TypeID: source.place.TypeID,
		})
		endLoans(index)
		result.States = append(result.States, ownershipSnapshot(index, places, activeLoans))
	}
	result.Work++
	returned, ok := places[body.Result]
	if !ok {
		return fail(diagnostic.Error("name.unknown", body.Span, "linear result is unknown"))
	}
	if !returned.initialized {
		return fail(useAfterMove(body.Span, returned))
	}
	ordinal := len(body.Bindings)
	result.Operations = append(result.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, ordinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, ordinal),
		Kind: core.OpReturn, SourceID: returned.place.ID, TypeID: returned.place.TypeID,
	})
	endLoans(ordinal)
	result.States = append(result.States, ownershipSnapshot(ordinal, places, activeLoans))
	return result
}

type loanUse struct {
	index int
	span  diagnostic.Span
}

type loanState struct {
	id      string
	ownerID string
	// access is "shared" or "exclusive" (Phase 3). It decides which rows of
	// the five-row conflict matrix apply when a new loan is created on the
	// same owner: shared-vs-shared never conflicts, every other combination
	// does. See conflictingLoan.
	access      string
	borrowedAt  diagnostic.Span
	lastUse     int
	lastUseSpan diagnostic.Span
}

// conflictingLoan selects the loan (if any) among an owner's currently active
// loans that conflicts with a newly-created loan of newAccess, using the same
// deterministic-first-by-sorted-ID selection the existing move_while_borrowed
// gate already uses (so two runs never disagree on which loan a diagnostic
// blames). Shared-plus-shared overlap is never a conflict; every other
// combination (shared+exclusive, exclusive+shared, exclusive+exclusive) is.
func conflictingLoan(candidates map[string]*loanState, newAccess string) *loanState {
	if len(candidates) == 0 {
		return nil
	}
	ids := make([]string, 0, len(candidates))
	for id, loan := range candidates {
		if newAccess == "shared" && loan.access == "shared" {
			continue
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Strings(ids)
	return candidates[ids[0]]
}

// borrowConflictDiagnostic builds the ownership.borrow_conflict rejection:
// span-bearing causes first (mirroring move_while_borrowed's construction),
// then ID-bearing detail causes (loan/owner/type), then a repair — the exact
// shape check.go's other ownership diagnostics already use.
func borrowConflictDiagnostic(span diagnostic.Span, blocking *loanState, ownerID, ownerTypeID string) diagnostic.Diagnostic {
	causes := []diagnostic.Cause{
		{Kind: "borrow_created_here", Span: spanPointer(blocking.borrowedAt)},
		{Kind: "borrow_used_later", Span: spanPointer(blocking.lastUseSpan)},
		{Kind: "loan", Detail: blocking.id},
		{Kind: "owner", Detail: ownerID},
		{Kind: "type", Detail: ownerTypeID},
	}
	return diagnostic.ErrorWithRepairs(
		"ownership.borrow_conflict", span, "cannot create a loan while a conflicting loan is live", causes,
		diagnostic.Repair{Kind: "create_loan_after_conflicting_loan_ends"},
	)
}

// discoverLoanLastUses extends every loan to its last transitively derived
// use. Association is inherited: a binding derived from a loan-derived
// binding stays associated with the same loans, so a reborrow
// (`let c = borrow b`) or a copy of a loan (`let c = b`) keeps extending the
// original owner's blocked window. Associating only the immediate borrow
// target expires the loan one hop early and admits a move while the loan is
// still observable.
func discoverLoanLastUses(parameterName string, body *ast.LinearBody) map[int]loanUse {
	uses := make(map[int]loanUse)
	visible := map[string]int{parameterName: -1}
	loansForBinding := make(map[int][]int)
	for index, binding := range body.Bindings {
		var inherited []int
		if sourceBinding, ok := visible[binding.RHS.Source]; ok {
			for _, loanIndex := range loansForBinding[sourceBinding] {
				if use, tracked := uses[loanIndex]; tracked {
					use.index = index
					use.span = binding.RHS.Span
					uses[loanIndex] = use
				}
				inherited = append(inherited, loanIndex)
			}
		}
		visible[binding.Name] = index
		if binding.RHS.Kind == "borrow" || binding.RHS.Kind == "borrow_mut" {
			uses[index] = loanUse{index: index, span: binding.RHS.Span}
			inherited = append(inherited, index)
		}
		loansForBinding[index] = inherited
	}
	if resultBinding, ok := visible[body.Result]; ok {
		for _, loanIndex := range loansForBinding[resultBinding] {
			if use, tracked := uses[loanIndex]; tracked {
				use.index = len(body.Bindings)
				use.span = body.Span
				uses[loanIndex] = use
			}
		}
	}
	return uses
}

func ownershipSnapshot(index int, places map[string]*placeState, activeLoans map[string]map[string]*loanState) ownershipStateFact {
	initialized := make([]string, 0, len(places))
	for _, place := range places {
		if place.initialized {
			initialized = append(initialized, place.place.ID)
		}
	}
	sort.Strings(initialized)
	loans := make([]string, 0)
	for _, ownerLoans := range activeLoans {
		for loanID := range ownerLoans {
			loans = append(loans, loanID)
		}
	}
	sort.Strings(loans)
	return ownershipStateFact{OperationIndex: index, InitializedPlaces: initialized, ActiveLoans: loans}
}

func hasTypeAbility(fact core.TypeFact, wanted core.Ability) bool {
	for _, candidate := range fact.Abilities {
		if candidate == wanted {
			return true
		}
	}
	return false
}

// executableShape reports whether a linear (non-match) function's parameter
// type has a native execution lowering this phase. cgen.linearInput only
// maps Byte and Buffer; every other shape (Box, Pair, or a nominal data type
// used in a straight-line body) type-checks and derives abilities cleanly
// but has no engine that can run it (D-02-09/D-07). Match-based functions
// (checkBranch included) are unaffected: their parameter is always a
// declared nominal alternative type, always executable via emitMatch/
// emitBranch's enum-based lowering, so this gate is only consulted from
// checkLinear.
func executableShape(value core.TypeRef) bool {
	return value.Constructor == "Byte" || value.Constructor == "Buffer"
}

func typeNodeCount(value core.TypeRef) int {
	count := 1
	for _, argument := range value.Arguments {
		count += typeNodeCount(argument)
	}
	return count
}

func missingAbilityDetail(fact core.TypeFact, requested core.Ability) string {
	for _, witness := range fact.NegativeWitnesses {
		if witness.Ability == requested {
			return string(requested) + ":" + strings.Join(witness.Path, ".")
		}
	}
	return string(requested)
}

func loanID(loan *loanState) string {
	if loan == nil {
		return ""
	}
	return loan.id
}

func spanPointer(span diagnostic.Span) *diagnostic.Span {
	copy := span
	return &copy
}

func coreType(value ast.TypeRef) core.TypeRef {
	result := core.TypeRef{Constructor: value.Constructor, Arguments: make([]core.TypeRef, 0, len(value.Arguments))}
	for _, argument := range value.Arguments {
		result.Arguments = append(result.Arguments, coreType(argument))
	}
	return result
}

func sameType(left, right ast.TypeRef) bool {
	return typeKey(coreType(left)) == typeKey(coreType(right))
}

func typeKey(value core.TypeRef) string {
	if len(value.Arguments) == 0 {
		return value.Constructor
	}
	parts := make([]string, 0, len(value.Arguments))
	for _, argument := range value.Arguments {
		parts = append(parts, typeKey(argument))
	}
	return value.Constructor + "<" + strings.Join(parts, ",") + ">"
}

func semanticID(module, kind, name string) string { return "s1:" + module + ":" + kind + ":" + name }

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

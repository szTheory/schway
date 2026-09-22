// Package reduce implements Phase 5's core-level program reducer (INT-02):
// deterministic moves applied to fixpoint or a bounded work budget,
// operating on the typed core.Program artifact ONLY -- the operation list
// and block graph -- never on source text (D-05-23). This project's
// identity discipline is function-local semantic ordinals, never source
// offsets, which is exactly why a text- or offset-based reducer (ddmin,
// C-Reduce) is the wrong tool here: it would collide head-on with that
// discipline and could emit structurally invalid candidates. See
// predicate.go for the interestingness oracle this package's Reduce
// function is driven by, and ProjectSource (below) for the reduced SOURCE
// case, which is always a pretty-printed projection of the reduced core
// produced by the SAME reduction run -- never independently re-parsed and
// re-reduced.
//
// Phase 11 (D-11-29/D-11-30) widens Reduce to accept a multi-function
// Seed: two whole-program moves (drop-call-site, drop-orphan-function) run
// FIRST, as HDD's coarse level, followed by the original five per-function
// moves looped over every function in the seed. reduce deliberately never
// imports the call-graph package (D-11-29): that package's own consumer
// set is a documented, load-bearing independence boundary this package
// stays outside of, so Seed.EntryFunctionID is caller-supplied rather than
// re-derived here, and the per-function loops iterate Functions in the
// caller-supplied slice order (already deterministic -- Go slice order is
// stable), never re-sorted by a locally-computed graph order.
package reduce

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// AttemptsPerFunction is Phase 11's derived work-budget factor (D-11-31),
// replacing the flat MaxReductionAttempts = 64 (D-05-25) it supersedes: the
// per-attempt bound applied ONCE PER FUNCTION in the seed, plus one more
// attempt per call site (derivedAttemptBound below). A one-function
// zero-call seed's derived bound is therefore exactly
// AttemptsPerFunction*1+0 == 64, byte-identical to the old flat bound, so
// every existing single-function reduction's Attempts count is unchanged.
const AttemptsPerFunction = 64

// derivedAttemptBound computes Reduce's bounded work budget for the given
// program (D-11-31): AttemptsPerFunction*len(Functions) + callSiteCount.
// Following check.loanLivenessBound's own "+1 floor" lesson (a formula that
// can compute zero for some legal shape must be rejected, not shipped) --
// this formula floors at AttemptsPerFunction itself: the degenerate case a
// literal product could reach is a program with zero functions, which is
// not a legal seed under any of this package's own invariants, but the
// floor makes that fact true by construction rather than by hoping no
// caller ever passes one.
func derivedAttemptBound(p core.Program) int {
	bound := AttemptsPerFunction*len(p.Functions) + countCallSites(p)
	if bound < AttemptsPerFunction {
		return AttemptsPerFunction
	}
	return bound
}

// countCallSites counts every core.OpCall operation across every function
// in the program -- the derived bound's per-call-site share (D-11-31).
func countCallSites(p core.Program) int {
	count := 0
	for _, fn := range p.Functions {
		if fn.Linear == nil {
			continue
		}
		for _, op := range fn.Linear.Operations {
			if op.Kind == core.OpCall {
				count++
			}
		}
	}
	return count
}

// Minimality is a closed two-value field (D-05-25): any other string is a
// bug, never a third state quietly introduced later. Exhausting the
// derived attempt budget is NOT a lane failure -- it sets this to
// MinimalityBudgetExhausted instead of MinimalityFixpoint, the mechanical
// form of INT-02's own "smallest KNOWN case" hedge, made auditable rather
// than silently degrading. budget_exhausted is never reclassified as a
// refusal and no third state is introduced (D-11-31).
const (
	MinimalityFixpoint        = "fixpoint"
	MinimalityBudgetExhausted = "budget_exhausted"
)

// Move is one of the five fixed reduction moves. Apply proposes exactly one
// candidate transformation of the given core.Program -- the first eligible
// site, in a fixed deterministic scan order -- and reports whether it found
// one. Moves operate on core.Program structure ONLY; no function in this
// package accepts, returns, or parses source text.
type Move struct {
	Name  string
	Apply func(core.Program) (core.Program, bool)
}

// Moves returns EXACTLY the seven reduction moves, in this EXACT order:
// the two Phase 11 whole-program moves (drop-call-site,
// drop-orphan-function) are PREPENDED as HDD's coarse level, running to
// exhaustion before intra-function narrowing begins, followed by the
// original five per-function moves in their original relative order. The
// order is the determinism guarantee (D-05-23/D-11-29) and must never be
// sorted, shuffled, or made configurable. drop-call-site always runs
// before drop-orphan-function becomes eligible: a callee only becomes
// orphanable after its last call site is dropped, and Reduce's own pass
// structure (each move's inner loop runs to exhaustion before the next
// move starts) guarantees that ordering without any extra bookkeeping.
func Moves() []Move {
	return []Move{
		{Name: "drop-call-site", Apply: dropCallSite},
		{Name: "drop-orphan-function", Apply: dropOrphanFunction},
		{Name: "drop-unused-binding", Apply: dropUnusedBinding},
		{Name: "drop-unmatched-arm", Apply: dropUnmatchedArm},
		{Name: "drop-offpath-foreign-stage", Apply: dropOffpathForeignStage},
		{Name: "collapse-branch-to-diverging-arm", Apply: collapseBranchToDivergingArm},
		{Name: "truncate-to-minimal-prefix", Apply: truncateToMinimalPrefix},
	}
}

// Seed is Reduce's Phase 11 input (D-11-30): a whole core.Program plus the
// caller-resolved entry function ID. The entry function is CALLER-SUPPLIED
// and never minted as a new core schema field -- core.Function carries no
// export or access field, and Callable lives on core.FunctionSignature in
// the interface artifact, not on core.Function itself. Production callers
// resolve EntryFunctionID via callgraph.EntryFunction before constructing
// a Seed; reduce itself never imports callgraph (D-11-29).
type Seed struct {
	Program         core.Program
	EntryFunctionID string
}

// entryFunctionID is set once per Reduce call (D-11-30) and read by
// dropOrphanFunction: the one function dropOrphanFunction must never
// remove. Move.Apply's signature is core.Program-only (Moves()' own fixed
// shape, D-05-23) so the five untouched per-function moves never need
// Seed threaded through them; only drop-orphan-function needs the entry
// fact, and Reduce is not re-entrant or concurrent by design (matching
// every other Phase 5 lane -- pure functions over immutable core.Program
// values, no goroutines), so a single package-level value set at the top
// of Reduce and read here is safe, mirroring testOnlyMoves' own seam
// shape above.
var entryFunctionID string

// testOnlyMoves is D-05-27's fault-injection seam: when non-nil, Reduce
// uses this move list instead of Moves()' fixed order. Test-only, exposed
// via export_test.go's SetTestOnlyMoves/ResetTestOnlyMoves following this
// repository's testOnlyForceUniformLoanJoin precedent (check.go) -- a seam
// that never gates production behaviour until proven safe. Production
// code must never set this; movesToApply falls back to the real Moves()
// whenever it is nil, which is its permanent default value.
var testOnlyMoves func() []Move

// movesToApply is Reduce's own move-source indirection: testOnlyMoves when
// set (TestNoOpReducerGoesRed's identity-move override,
// TestReducerNonDeterminismGoesRed's randomized-order override), otherwise
// the real, fixed Moves() order every production caller observes.
func movesToApply() []Move {
	if testOnlyMoves != nil {
		return testOnlyMoves()
	}
	return Moves()
}

// Result is one completed reduction run. Source is always populated inside
// Reduce from Result.Program -- D-05-23's binding rule that the reduced
// source case and the reduced core case always correspond, because both
// come from the very same run.
type Result struct {
	Program             core.Program
	Source              string
	Minimality          string
	Attempts            int
	TotalRecomputedWork int
	AppliedMoves        []string
}

// Predicate decides whether a candidate core.Program is still "interesting"
// -- see Interesting (predicate.go) for the D-05-24 conjunction a real
// Predicate implementation is expected to apply.
type Predicate func(ctx context.Context, candidate core.Program) (Signature, bool, error)

// attemptWork is the fixed per-attempt charge summed into
// Result.TotalRecomputedWork: one check+compile+execute cycle is one unit
// of RecomputedWork, the same currency every other Phase 5 lane uses (no
// second cost unit is introduced here).
const attemptWork = 1

// Reduce applies the moves greedily in fixed order to fixpoint: it calls
// move 1's Apply repeatedly -- each accepted candidate narrows the program
// further -- until Apply reports no further eligible site OR a produced
// candidate is rejected by interesting, then advances to move 2, and so
// on. On reaching the last move, if any move applied during that pass,
// Reduce restarts from move 1; a full pass across every move with no
// accepted candidate is a fixpoint. Reduce also stops, non-erroring, once
// its derived attempt budget (derivedAttemptBound, D-11-31) has been
// spent -- each call to interesting is one such cycle, whether or not it
// is accepted.
func Reduce(ctx context.Context, seed Seed, interesting Predicate) (Result, error) {
	if interesting == nil {
		return Result{}, fmt.Errorf("reduce: interesting predicate must not be nil")
	}
	// WR-01: fail closed BEFORE any move runs. dropOrphanFunction exempts
	// exactly one function from deletion by ID, and a program's real entry
	// is by construction in-degree-zero -- so an unmatched EntryFunctionID
	// would let this reducer silently delete the very function the seed is
	// about. Seed.Validate (seed_validate.go) refuses instead; see its doc
	// comment for why single-function seeds are exempt.
	if err := seed.Validate(); err != nil {
		return Result{}, err
	}

	current := cloneProgram(seed.Program)
	entryFunctionID = seed.EntryFunctionID
	maxAttempts := derivedAttemptBound(current)
	var applied []string
	attempts := 0
	totalWork := 0
	minimality := MinimalityFixpoint

pass:
	for {
		progressedThisPass := false
		for _, move := range movesToApply() {
			for {
				if attempts >= maxAttempts {
					minimality = MinimalityBudgetExhausted
					break pass
				}
				candidate, ok := move.Apply(current)
				if !ok {
					break
				}
				attempts++
				_, keep, err := interesting(ctx, candidate)
				if err != nil {
					return Result{}, err
				}
				totalWork += attemptWork
				if !keep {
					break
				}
				current = candidate
				applied = append(applied, move.Name)
				progressedThisPass = true
			}
		}
		if !progressedThisPass {
			break
		}
	}

	return Result{
		Program:             current,
		Source:              ProjectSource(current),
		Minimality:          minimality,
		Attempts:            attempts,
		TotalRecomputedWork: totalWork,
		AppliedMoves:        applied,
	}, nil
}

// ---------------------------------------------------------------------
// Move 0a: drop-call-site (D-11-29, Q-01 BRANCH A -- a live move)
// ---------------------------------------------------------------------

// dropCallSite rewrites the FIRST eligible core.OpCall operation (program
// order across functions, then within a function's own Operations order)
// to core.OpCopy, preserving SourceID and TargetID byte-for-byte and
// clearing CalleeID. No renumbering, no place churn, no block surgery --
// every operation ID stays exactly what it was, which is what keeps
// Signature.OperationID valid across this move and keeps plan 11-09's
// strict-field-equality re-verification safe to demand
// (TestQ01CoreLevelOpCallToOpCopyRewrite settled Q-01 BRANCH A: corevalidate
// accepts this exact rewrite). A call whose argument's own declared type
// differs from the call's own result type ("type-changing calls",
// RefusedShapes()) is skipped -- OpCopy cannot express a type change, so
// rewriting it would emit an ill-typed candidate rather than a smaller
// one.
func dropCallSite(p core.Program) (core.Program, bool) {
	for fi, fn := range p.Functions {
		if fn.Linear == nil {
			continue
		}
		placeType := placeTypesByID(fn)
		for oi, op := range fn.Linear.Operations {
			if op.Kind != core.OpCall {
				continue
			}
			if placeType[op.SourceID] != op.TypeID {
				continue
			}
			out := cloneProgram(p)
			target := &out.Functions[fi].Linear.Operations[oi]
			target.Kind = core.OpCopy
			target.CalleeID = ""
			return out, true
		}
	}
	return p, false
}

// placeTypesByID maps a function's own Places to their declared TypeID,
// used by dropCallSite to detect a type-changing call.
func placeTypesByID(fn core.Function) map[string]string {
	m := make(map[string]string, len(fn.Linear.Places))
	for _, place := range fn.Linear.Places {
		m[place.ID] = place.TypeID
	}
	return m
}

// ---------------------------------------------------------------------
// Move 0b: drop-orphan-function (D-11-29, Q-01 BRANCH A -- a live move)
// ---------------------------------------------------------------------

// dropOrphanFunction deletes the FIRST non-entry function (in Functions
// declaration order) with zero in-edges over all OpCall.CalleeID in the
// program. The function named by entryFunctionID (set once per Reduce
// call from Seed.EntryFunctionID) is never eligible -- it is the one
// function that can never be orphaned. A callee only becomes orphanable
// after its last call site is dropped; Reduce's own pass structure (each
// move's inner loop runs to exhaustion before the next move starts)
// guarantees drop-call-site has already run to exhaustion for this pass
// before this move's own inner loop ever starts.
func dropOrphanFunction(p core.Program) (core.Program, bool) {
	if len(p.Functions) <= 1 {
		return p, false
	}
	called := map[string]bool{}
	for _, fn := range p.Functions {
		if fn.Linear == nil {
			continue
		}
		for _, op := range fn.Linear.Operations {
			if op.Kind == core.OpCall && op.CalleeID != "" {
				called[op.CalleeID] = true
			}
		}
	}
	for fi, fn := range p.Functions {
		if fn.ID == entryFunctionID {
			continue
		}
		if called[fn.ID] {
			continue
		}
		out := cloneProgram(p)
		out.Functions = append(append([]core.Function{}, out.Functions[:fi]...), out.Functions[fi+1:]...)
		return out, true
	}
	return p, false
}

// ---------------------------------------------------------------------
// Move 1: drop-unused-binding
// ---------------------------------------------------------------------

// dropUnusedBinding removes the FIRST (in Operations order, looped over
// every function in the seed in declaration order) pure value-producing
// operation (copy/move/borrow) whose target place is never read as any
// other operation's SourceID anywhere in that same function. Foreign
// acquisitions/releases are explicitly excluded -- move 3 owns removing a
// completed acquire/release pair as a unit, since an acquisition's target
// being unused does not make its (still-observable) release side effect
// droppable on its own.
func dropUnusedBinding(p core.Program) (core.Program, bool) {
	for fi, fn := range p.Functions {
		if fn.Linear == nil {
			continue
		}
		used := usedSourceIDs(fn)
		for i, op := range fn.Linear.Operations {
			if !isPureValueKind(op.Kind) {
				continue
			}
			if op.TargetID == "" || used[op.TargetID] {
				continue
			}
			out := cloneProgram(p)
			removeOperationAt(&out.Functions[fi], i)
			return out, true
		}
	}
	return p, false
}

// ---------------------------------------------------------------------
// Move 2: drop-unmatched-arm
// ---------------------------------------------------------------------

// dropUnmatchedArm removes the first match arm when a match has MORE THAN
// TWO arms, keeping the match well-formed (a two-arm match is exactly the
// branch shape collapse-branch-to-diverging-arm (move 4) owns collapsing;
// this move never reduces a match below two arms, since a single surviving
// arm is a different move's job and this project's checker requires
// exhaustive dispatch over a data type's declared alternatives).
func dropUnmatchedArm(p core.Program) (core.Program, bool) {
	for fi, fn := range p.Functions {
		if fn.Match == nil || len(fn.Match.Arms) <= 2 {
			continue
		}
		out := cloneProgram(p)
		outFn := &out.Functions[fi]
		removed := outFn.Match.Arms[0]
		outFn.Match.Arms = append([]core.MatchArm{}, outFn.Match.Arms[1:]...)
		if removed.BlockID != "" && outFn.Linear != nil {
			block := findBlock(outFn.Linear.Blocks, removed.BlockID)
			if block != nil {
				removedOpIDs := append([]string(nil), block.OperationIDs...)
				outFn.Linear.Operations = removeOpsByID(outFn.Linear.Operations, removedOpIDs)
				outFn.Linear.Blocks = removeBlock(outFn.Linear.Blocks, removed.BlockID)
				outFn.Linear.Edges = removeEdgesTouching(outFn.Linear.Edges, removed.BlockID)
				for i := range outFn.Linear.Blocks {
					outFn.Linear.Blocks[i].Successors = removeString(outFn.Linear.Blocks[i].Successors, removed.BlockID)
				}
			}
		}
		return out, true
	}
	return p, false
}

// ---------------------------------------------------------------------
// Move 3: drop-offpath-foreign-stage
// ---------------------------------------------------------------------

// dropOffpathForeignStage removes the FIRST completed foreign
// acquisition/release pair (an OpForeignCall together with the OpRelease
// that names it via ReleasesOperationID). The move itself has no notion of
// "on" vs "off" the diverging causal path -- that is exactly what the
// interestingness predicate decides by rejecting a candidate that moved the
// diverging operation or engine pair (D-05-24); this move is only a
// deterministic candidate generator, always proposing the earliest
// completed stage in program order.
// dropOffpathForeignStage removes the FIRST completed acquisition (in
// program order) from a resource-lifecycle foreign-call chain, rebuilding
// the entire chain from scratch via rebuildForeignChain -- the exact same
// algorithm check.go's checkResourceLifecycle uses to derive a chain from
// source. Regenerating from scratch (rather than surgically patching the
// existing block/edge graph) is what keeps a reduced candidate always
// referentially consistent: RES-01 requires every completed acquisition to
// be discharged by a release on EVERY path that completes after it (this
// project's real corpus release-ladders always have at least one, often
// two: an error-path release and a success-path release), so ad hoc
// deletion of "the first matching release" would leave the other dangling.
// This move's own notion of "completed" is simply "present" -- every call
// this project's checker admits is completed on at least one path, so any
// remaining call is always a legal removal target; interestingness (not
// this move) is what decides whether removing it disturbs the diverging
// causal path (D-05-24).
func dropOffpathForeignStage(p core.Program) (core.Program, bool) {
	for fi, fn := range p.Functions {
		if fn.Linear == nil || fn.ForeignContract == nil || len(fn.Linear.Blocks) == 0 {
			continue
		}
		steps := decodeForeignSteps(fn)
		if len(steps) == 0 {
			continue
		}
		parameterFact := fn.Linear.Types[0]
		var returnFact *core.TypeFact
		if p.Schema == core.Schema1 {
			parameter, returned, ok := directionalSignatureFacts(fn)
			if !ok {
				// A /1 reduction may not invent a distinct return fact. A
				// distinct return without a declared fact is refused here
				// instead of inheriting an unrelated failure-operation fact.
				continue
			}
			parameterFact, returnFact = parameter, &returned
		}
		remaining := steps[1:]

		out := cloneProgram(p)
		outFn := &out.Functions[fi]
		if len(remaining) == 0 {
			// No acquisitions survive -- this function is now exactly a plain
			// parameter-in/parameter-out straight-line body, and the
			// ForeignContract fact that justified the foreign surface is
			// dead (a fresh check of the corresponding plain source would
			// never re-derive it).
			outFn.ForeignContract = nil
			types := []core.TypeFact{parameterFact}
			if returnFact != nil {
				types = append(types, *returnFact)
			}
			outFn.Linear = &core.LinearBody{
				ID:    fn.ID + ":linear",
				Types: types,
				Places: []core.Place{
					{ID: fn.Parameter.ID, Name: fn.Parameter.Name, TypeID: parameterFact.ID},
				},
				Operations: []core.LinearOperation{
					{ID: fn.ID + ":op:0", PointID: fn.ID + ":point:linear:0", Kind: core.OpReturn, SourceID: fn.Parameter.ID, TypeID: parameterFact.ID},
				},
			}
			if len(p.Functions) == 1 {
				// Single-function seed (the byte-identical pre-Phase-11
				// behaviour this move must preserve exactly): the
				// program-level DataTypes this one function's foreign
				// surface justified are now dead too.
				out.DataTypes = nil
			}
			// Multi-function seed: DataTypes is a program-level fact that
			// may still be referenced by a SIBLING function's own
			// parameter/return type, so it is deliberately left untouched
			// here rather than risk breaking a sibling that still needs it.
			return out, true
		}
		failShape := core.TypeRef{Constructor: fn.ForeignContract.Fails}
		failGranted := []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape}
		outFn.Linear = rebuildForeignChain(fn.ID, fn.Parameter, parameterFact, returnFact, failShape, failGranted, remaining)
		return out, true
	}
	return p, false
}

// directionalSignatureFacts recovers the two declared facts that a /1 reduced
// source function will re-mint. A /0 artifact's type:1 may instead be an
// error fact, so callers must invoke this only for a /1 program. A distinct
// return with no explicit matching fact fails closed rather than inheriting
// an unrelated failure-operation fact.
func directionalSignatureFacts(fn core.Function) (core.TypeFact, core.TypeFact, bool) {
	if fn.Linear == nil {
		return core.TypeFact{}, core.TypeFact{}, false
	}
	parameterTypeID := ""
	for _, place := range fn.Linear.Places {
		if place.ID == fn.Parameter.ID {
			parameterTypeID = place.TypeID
			break
		}
	}
	if parameterTypeID == "" {
		return core.TypeFact{}, core.TypeFact{}, false
	}
	var parameterFact core.TypeFact
	parameterFound := false
	for _, fact := range fn.Linear.Types {
		if fact.ID == parameterTypeID {
			parameterFact, parameterFound = fact, true
			break
		}
	}
	if !parameterFound || parameterFact.Shape.Constructor != fn.Parameter.Type {
		return core.TypeFact{}, core.TypeFact{}, false
	}
	returnTypeID := fn.ID + ":type:1"
	for _, fact := range fn.Linear.Types {
		if fact.ID == returnTypeID && fact.Shape.Constructor == fn.ReturnType {
			return parameterFact, fact, true
		}
	}
	return core.TypeFact{}, core.TypeFact{}, false
}

// foreignStepInfo is the minimal per-step fact rebuildForeignChain needs:
// the surviving binding's own place name (every step in this project's
// real corpus calls its function's single declared symbol with its single
// parameter as the sole argument -- see reduce.go's package doc and
// 05-PATTERNS.md's own note that core.Function carries exactly one
// ForeignContract per function).
type foreignStepInfo struct {
	bindingName string
	allocator   string
}

// decodeForeignSteps recovers the ordered call sequence from a
// resource-lifecycle LinearBody. checkResourceLifecycle always places the n
// OpForeignCall operations FIRST, in call order, before any OpFail/
// OpRelease/OpReturn (core.go's LinearOperation slice is built call-by-call
// before any release/fail/return is appended) -- so filtering Operations
// for OpForeignCall, preserving order, exactly recovers the step sequence
// regardless of how many releases/fails have already been reduced away
// elsewhere.
func decodeForeignSteps(fn core.Function) []foreignStepInfo {
	nameByID := make(map[string]string, len(fn.Linear.Places))
	for _, place := range fn.Linear.Places {
		nameByID[place.ID] = place.Name
	}
	var steps []foreignStepInfo
	for _, op := range fn.Linear.Operations {
		if op.Kind != core.OpForeignCall {
			continue
		}
		steps = append(steps, foreignStepInfo{bindingName: nameByID[op.TargetID], allocator: op.Allocator})
	}
	return steps
}

// rebuildForeignChain reconstructs a resource-lifecycle LinearBody for the
// given surviving steps, replicating checkResourceLifecycle's own ID
// scheme and block/edge shape exactly (place:1..n for the ok bindings,
// place:(n+1)..2n for the err places, block:entry/block:step:i/
// block:err:i/block:success, releases in reverse completion order on both
// the success path and each try_call's own error path) -- see check.go's
// checkResourceLifecycle, the algorithm this function is a deliberate,
// narrowed peer of (every step here is always a "try_call" to the SAME
// declared symbol, this project's real corpus never uses discard_call).
func rebuildForeignChain(functionID string, parameter core.Parameter, parameterType core.TypeFact, returnType *core.TypeFact, failShape core.TypeRef, failGranted []core.Ability, steps []foreignStepInfo) *core.LinearBody {
	n := len(steps)
	parameterID := parameter.ID
	typeID := parameterType.ID

	types := make([]core.TypeFact, 1, 1+n)
	types[0] = parameterType
	if returnType != nil {
		types = append(types, *returnType)
	}

	type builtStep struct {
		callID, okPlaceID, errPlaceID, errTypeID, blockID, okEdgeID, errEdgeID string
		allocator                                                              string
	}
	built := make([]builtStep, n)
	okPlaces := make([]core.Place, n)
	errPlaces := make([]core.Place, n)
	for i, step := range steps {
		errTypeID := fmt.Sprintf("%s:type:%d", functionID, len(types))
		types = append(types, core.TypeFact{ID: errTypeID, Shape: failShape, Abilities: failGranted, NegativeWitnesses: []core.AbilityWitness{}})
		okPlaceID := fmt.Sprintf("%s:place:%d", functionID, i+1)
		errPlaceID := fmt.Sprintf("%s:place:%d", functionID, n+1+i)
		okPlaces[i] = core.Place{ID: okPlaceID, Name: step.bindingName, TypeID: typeID}
		errPlaces[i] = core.Place{ID: errPlaceID, Name: fmt.Sprintf("_err%d", i), TypeID: errTypeID}
		blockID := functionID + ":block:entry"
		if i > 0 {
			blockID = fmt.Sprintf("%s:block:step:%d", functionID, i)
		}
		built[i] = builtStep{
			callID: fmt.Sprintf("%s:op:%d", functionID, i), okPlaceID: okPlaceID, errPlaceID: errPlaceID, errTypeID: errTypeID,
			blockID: blockID, allocator: step.allocator,
		}
	}
	places := make([]core.Place, 0, 1+2*n)
	places = append(places, core.Place{ID: parameterID, Name: parameter.Name, TypeID: typeID})
	places = append(places, okPlaces...)
	places = append(places, errPlaces...)

	successBlockID := functionID + ":block:success"
	var blocks []core.Block
	var edges []core.Edge
	operations := make([]core.LinearOperation, n)
	for i := 0; i < n; i++ {
		okTarget := successBlockID
		if i+1 < n {
			okTarget = built[i+1].blockID
		}
		errTarget := fmt.Sprintf("%s:block:err:%d", functionID, i)
		built[i].okEdgeID = fmt.Sprintf("%s:edge:step:%d:ok", functionID, i)
		built[i].errEdgeID = fmt.Sprintf("%s:edge:step:%d:err", functionID, i)
		pointID := fmt.Sprintf("%s:point:step:%d", functionID, i)
		if i == 0 {
			pointID = functionID + ":point:entry"
		}
		blocks = append(blocks, core.Block{
			ID: built[i].blockID, PointID: pointID,
			OperationIDs: []string{built[i].callID}, Successors: []string{okTarget, errTarget},
		})
		edges = append(edges,
			core.Edge{ID: built[i].okEdgeID, FromBlockID: built[i].blockID, ToBlockID: okTarget, Pattern: "ok"},
			core.Edge{ID: built[i].errEdgeID, FromBlockID: built[i].blockID, ToBlockID: errTarget, Pattern: "err"},
		)
		operations[i] = core.LinearOperation{
			ID: built[i].callID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, i), Kind: core.OpForeignCall,
			SourceID: parameterID, TargetID: built[i].okPlaceID, TypeID: typeID,
			OkEdgeID: built[i].okEdgeID, ErrEdgeID: built[i].errEdgeID, ErrTargetID: built[i].errPlaceID,
			Allocator: built[i].allocator,
		}
	}

	var completed []builtStep
	releaseOps := func(from []builtStep) []core.LinearOperation {
		ops := make([]core.LinearOperation, 0, len(from))
		for j := len(from) - 1; j >= 0; j-- {
			acquired := from[j]
			opID := fmt.Sprintf("%s:op:%d", functionID, len(operations)+len(ops))
			ops = append(ops, core.LinearOperation{
				ID: opID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, len(operations)+len(ops)),
				Kind: core.OpRelease, SourceID: acquired.okPlaceID, TypeID: typeID, ReleasesOperationID: acquired.callID,
				Allocator: acquired.allocator,
			})
		}
		return ops
	}
	for i := 0; i < n; i++ {
		errOps := releaseOps(completed)
		failOpIndex := len(operations) + len(errOps)
		failOpID := fmt.Sprintf("%s:op:%d", functionID, failOpIndex)
		errOps = append(errOps, core.LinearOperation{
			ID: failOpID, PointID: fmt.Sprintf("%s:point:linear:%d", functionID, failOpIndex),
			Kind: core.OpFail, SourceID: built[i].errPlaceID, TypeID: built[i].errTypeID,
		})
		errOpIDs := make([]string, len(errOps))
		for idx, op := range errOps {
			errOpIDs[idx] = op.ID
		}
		blocks = append(blocks, core.Block{
			ID: fmt.Sprintf("%s:block:err:%d", functionID, i), PointID: fmt.Sprintf("%s:point:err:%d", functionID, i), OperationIDs: errOpIDs,
		})
		operations = append(operations, errOps...)
		completed = append(completed, built[i])
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

	return &core.LinearBody{
		ID: functionID + ":linear", Types: types, Places: places, Operations: operations, Blocks: blocks, Edges: edges,
	}
}

// ---------------------------------------------------------------------
// Move 4: collapse-branch-to-diverging-arm
// ---------------------------------------------------------------------

// collapseBranchToDivergingArm collapses a two-arm match to its first arm's
// body, discarding the Match wrapper and the other arm entirely. It always
// tries Arms[0] (deterministic, fixed order); if that candidate is later
// rejected by interesting, this move simply stops applying for the current
// pass rather than trying Arms[1] -- a documented narrowing (see
// 05-10-SUMMARY.md) that keeps the move a pure, side-effect-free function
// of its input program. checkBranch's own convention prefixes every arm
// body with a parameter-aliasing OpCopy; this move strips that alias and
// renumbers the remaining operations/places to analyzeStraightLine's own
// ID scheme, so a collapsed result is structurally indistinguishable from
// what check.Program would produce for the equivalent straight-line
// source -- the correspondence ProjectSource's round-trip test depends on.
func collapseBranchToDivergingArm(p core.Program) (core.Program, bool) {
	for fi, fn := range p.Functions {
		if fn.Match == nil || len(fn.Match.Arms) != 2 || fn.Linear == nil {
			continue
		}
		arm := fn.Match.Arms[0]
		if arm.BlockID == "" {
			continue
		}
		block := findBlock(fn.Linear.Blocks, arm.BlockID)
		if block == nil {
			continue
		}
		armOps := opsByIDs(fn.Linear.Operations, block.OperationIDs)
		if len(armOps) == 0 {
			continue
		}
		rest := armOps
		aliasTargetID := ""
		if armOps[0].Kind == core.OpCopy && armOps[0].SourceID == fn.Parameter.ID {
			aliasTargetID = armOps[0].TargetID
			rest = armOps[1:]
		}
		if len(rest) == 0 {
			continue
		}

		origPlaceNames := make(map[string]string, len(fn.Linear.Places))
		parameterTypeID := ""
		for _, place := range fn.Linear.Places {
			origPlaceNames[place.ID] = place.Name
			if place.ID == fn.Parameter.ID {
				parameterTypeID = place.TypeID
			}
		}
		// The parameter TypeID is a source fact, not an inference from the
		// first surviving operation. Without it this projection cannot retain
		// directional parameter/return facts, so refuse this reduction rather
		// than silently substituting an operation (which may be a return).
		if parameterTypeID == "" {
			continue
		}

		newOps, newPlaces := renumberStraightLine(fn.ID, fn.Parameter, parameterTypeID, origPlaceNames, rest, aliasTargetID)

		out := cloneProgram(p)
		outFn := &out.Functions[fi]
		outFn.Match = nil
		outFn.Linear = &core.LinearBody{
			ID:         fn.ID + ":linear",
			Types:      fn.Linear.Types,
			Places:     newPlaces,
			Operations: newOps,
		}
		return out, true
	}
	return p, false
}

// renumberStraightLine renumbers a sequence of operations (with any
// alias-parameter reference redirected to the real parameter place) into
// analyzeStraightLine's own ID scheme: place:0 is the parameter, place:N
// for N>=1 is the target of the (N-1)th surviving operation, op:N/
// point:linear:N are assigned contiguously in order.
func renumberStraightLine(functionID string, parameter core.Parameter, parameterTypeID string, origPlaceNames map[string]string, ops []core.LinearOperation, aliasTargetID string) ([]core.LinearOperation, []core.Place) {
	idMap := map[string]string{parameter.ID: parameter.ID}
	if aliasTargetID != "" {
		idMap[aliasTargetID] = parameter.ID
	}
	places := []core.Place{{ID: parameter.ID, Name: parameter.Name, TypeID: parameterTypeID}}
	newOps := make([]core.LinearOperation, 0, len(ops))
	for index, op := range ops {
		newOp := op
		newOp.ID = fmt.Sprintf("%s:op:%d", functionID, index)
		newOp.PointID = fmt.Sprintf("%s:point:linear:%d", functionID, index)
		if mapped, ok := idMap[op.SourceID]; ok {
			newOp.SourceID = mapped
		}
		if op.TargetID != "" {
			newTargetID := fmt.Sprintf("%s:place:%d", functionID, index+1)
			idMap[op.TargetID] = newTargetID
			newOp.TargetID = newTargetID
			places = append(places, core.Place{ID: newTargetID, Name: origPlaceNames[op.TargetID], TypeID: op.TypeID})
		}
		newOps = append(newOps, newOp)
	}
	return newOps, places
}

// ---------------------------------------------------------------------
// Move 5: truncate-to-minimal-prefix
// ---------------------------------------------------------------------

// truncateToMinimalPrefix is scoped to a flat straight-line body (no
// Blocks, no Match): it repeatedly drops the operation immediately
// preceding the terminator, as long as that operation's target is EXACTLY
// what the terminator reads -- rewiring the terminator to read what the
// removed operation itself read. This is truncation to the minimal prefix
// still reaching whatever the terminator (and, transitively, the diverging
// operation, if it lies further back) needs; it stops the moment the value
// feeding the terminator is not the immediately preceding operation's own
// output, since removing further requires a deeper dependency analysis
// this move does not attempt.
func truncateToMinimalPrefix(p core.Program) (core.Program, bool) {
	for fi, fn := range p.Functions {
		if fn.Linear == nil || fn.Match != nil || len(fn.Linear.Blocks) > 0 {
			continue
		}
		ops := fn.Linear.Operations
		if len(ops) < 2 {
			continue
		}
		terminatorIdx := len(ops) - 1
		terminator := ops[terminatorIdx]
		if !isTerminatorKind(terminator.Kind) {
			continue
		}
		candidateIdx := terminatorIdx - 1
		removed := ops[candidateIdx]
		if removed.TargetID == "" || removed.TargetID != terminator.SourceID {
			continue
		}

		out := cloneProgram(p)
		outFn := &out.Functions[fi]
		newOps := make([]core.LinearOperation, 0, len(ops)-1)
		newOps = append(newOps, outFn.Linear.Operations[:candidateIdx]...)
		newTerminator := outFn.Linear.Operations[terminatorIdx]
		newTerminator.SourceID = removed.SourceID
		newTerminator.TypeID = removed.TypeID
		newOps = append(newOps, newTerminator)
		outFn.Linear.Operations = newOps
		outFn.Linear.Places = removePlace(outFn.Linear.Places, removed.TargetID)
		return out, true
	}
	return p, false
}

// ---------------------------------------------------------------------
// ProjectSource -- D-05-23's reduced-source-case projection
// ---------------------------------------------------------------------

// ProjectSource produces the reduced SOURCE case as a pretty-printed
// projection of the reduced core -- D-05-23's binding rule that the source
// case and the core case always correspond, because the source is derived
// from the SAME core this run reduced to, never independently re-parsed
// and re-reduced. It is deterministic: the same core always yields
// byte-identical text.
//
// Coverage is deliberately narrow (documented in 05-10-SUMMARY.md): exactly
// the two straight-line body shapes this project's actual Phase 5 corpus
// exercises -- a flat borrow/copy/take chain (no Blocks, no Match), and a
// foreign resource-lifecycle chain (Blocks present, a single
// ForeignContract, every call to that function's own declared symbol),
// PLUS (Phase 11, D-11-29) a plain OpCall to a sibling function. A program
// outside those shapes (e.g. one where collapse-branch has not yet
// eliminated its Match) projects to an explicit, clearly-marked
// unsupported-shape comment (RefusedShapes(), reduce.projection_unsupported)
// rather than guessing at invalid or non-corresponding source text. The
// single-function case (len(program.Functions) == 1) is projected via the
// exact pre-Phase-11 code path, byte-identical to every existing
// single-function reduction's own projected source; a multi-function
// program routes through projectMultiFunctionSource instead.
func ProjectSource(program core.Program) string {
	if len(program.Functions) == 0 {
		return unsupportedProjection(fmt.Sprintf("expected at least one function, got %d", len(program.Functions)))
	}
	if len(program.Functions) == 1 {
		return projectSingleFunctionSource(program, program.Functions[0])
	}
	return projectMultiFunctionSource(program)
}

func projectSingleFunctionSource(program core.Program, fn core.Function) string {
	if fn.Match != nil {
		return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + "match-bodied functions are not yet projected to source (see 05-10-SUMMARY.md)")
	}
	if fn.Linear == nil {
		return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + "function has no linear body")
	}
	if isDeclaredDataType(program, fn.Parameter.Type) || isDeclaredDataType(program, fn.ReturnType) {
		// This project's checker only admits a user-declared ADT parameter
		// or return type inside an EXHAUSTIVE match (S1) body -- a plain
		// straight-line Linear body is only ever checker-admitted for the
		// executable primitive shapes (Byte/Buffer). A reduced program
		// whose match has collapsed away (move 4) but whose parameter/return
		// type is still an ADT therefore has no valid, re-checkable source
		// projection under the current grammar (there is no partial-match
		// syntax): projecting it would either be rejected by the real front
		// end (a non-exhaustive match) or silently wrong (a fabricated
		// exhaustive match not derived from the reduced core). Documented
		// narrowing (05-10-SUMMARY.md) -- reduction at the CORE level is
		// unaffected; only this SOURCE projection is unsupported here.
		return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + "a collapsed ADT-typed match has no exhaustive, re-checkable source projection under the current grammar")
	}
	if fn.ForeignContract != nil && hasForeignCall(fn.Linear.Operations) {
		return projectForeignChainSource(program, fn)
	}
	return projectStraightLineSource(program, fn)
}

// projectMultiFunctionSource projects a multi-function program (Phase 11,
// D-11-29): every function must be a Match-free, non-foreign, non-ADT
// straight-line body (the same admitted shapes as the single-function
// path, minus the foreign-chain case, which this projector does not yet
// combine with sibling functions) -- any function outside that shape
// refuses the WHOLE projection via the same unsupportedProjection helper,
// naming the offending function, rather than silently omitting it.
func projectMultiFunctionSource(program core.Program) string {
	var exports strings.Builder
	var dataDecls strings.Builder
	declaredDataTypes := map[string]bool{}
	nameByFunctionID := make(map[string]string, len(program.Functions))
	for _, fn := range program.Functions {
		nameByFunctionID[fn.ID] = fn.Name
	}
	var bodies strings.Builder
	for _, fn := range program.Functions {
		if fn.Match != nil {
			return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + fmt.Sprintf("function %q: match-bodied functions are not yet projected to source in a multi-function program", fn.Name))
		}
		if fn.Linear == nil {
			return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + fmt.Sprintf("function %q has no linear body", fn.Name))
		}
		if isDeclaredDataType(program, fn.Parameter.Type) || isDeclaredDataType(program, fn.ReturnType) {
			return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + fmt.Sprintf("function %q: a collapsed ADT-typed match has no exhaustive, re-checkable source projection under the current grammar", fn.Name))
		}
		if fn.ForeignContract != nil && hasForeignCall(fn.Linear.Operations) {
			return unsupportedProjection(refusalTag(RefusalProjectionUnsupported) + fmt.Sprintf("function %q: foreign resource-lifecycle chains are not yet projected to source in a multi-function program", fn.Name))
		}
		fmt.Fprintf(&exports, "  fn %s\n", fn.Name)
		for _, dt := range []string{fn.Parameter.Type, fn.ReturnType} {
			if declaredDataTypes[dt] {
				continue
			}
			decl := renderDataType(program, dt)
			if decl != "" {
				declaredDataTypes[dt] = true
				dataDecls.WriteString(decl)
			}
		}
		bodies.WriteString(projectMultiFunctionFunctionBody(fn, nameByFunctionID))
	}
	return fmt.Sprintf("module %s\n\nexport {\n%s}\n\n%s%s", program.Module, exports.String(), dataDecls.String(), bodies.String())
}

// projectMultiFunctionFunctionBody projects one function's own declaration
// (never the module header/exports, which projectMultiFunctionSource emits
// once for the whole program). It extends projectStraightLineSource's own
// operation-kind switch with core.OpCall, projected as a plain call
// expression to the resolved callee's name (never its ID) -- the source
// syntax this project's own multi-function corpus already uses (see e.g.
// testdata/phase07/deep_diamond_acyclic.lang's "let result = leaf(value)").
func projectMultiFunctionFunctionBody(fn core.Function, nameByFunctionID map[string]string) string {
	nameByID := map[string]string{fn.Parameter.ID: fn.Parameter.Name}
	for _, place := range fn.Linear.Places {
		nameByID[place.ID] = place.Name
	}
	var body strings.Builder
	terminal := "  " + fn.Parameter.Name + "\n"
	for _, op := range fn.Linear.Operations {
		switch op.Kind {
		case core.OpBorrowShared:
			fmt.Fprintf(&body, "  let %s = borrow %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpBorrowExclusive:
			fmt.Fprintf(&body, "  let %s = borrow mut %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpMove:
			fmt.Fprintf(&body, "  let %s = take %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpCopy:
			fmt.Fprintf(&body, "  let %s = %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpCall:
			fmt.Fprintf(&body, "  let %s = %s(%s)\n", nameByID[op.TargetID], nameByFunctionID[op.CalleeID], nameByID[op.SourceID])
		case core.OpReturn:
			terminal = "  " + nameByID[op.SourceID] + "\n"
		case core.OpDefect:
			terminal = fmt.Sprintf("  defect %q\n", op.Reason)
		}
	}
	return fmt.Sprintf("fn %s(%s: %s) -> %s {\n%s%s}\n\n", fn.Name, fn.Parameter.Name, fn.Parameter.Type, fn.ReturnType, body.String(), terminal)
}

func isDeclaredDataType(program core.Program, typeName string) bool {
	for _, dataType := range program.DataTypes {
		if dataType.Name == typeName {
			return true
		}
	}
	return false
}

func hasForeignCall(ops []core.LinearOperation) bool {
	for _, op := range ops {
		if op.Kind == core.OpForeignCall {
			return true
		}
	}
	return false
}

func unsupportedProjection(reason string) string {
	return "// reduce: projection unsupported for this program shape -- " + reason + "\n"
}

// ---------------------------------------------------------------------
// RefusedShapes() -- the enumerated register of shapes reduce provably
// declines (D-11-35). Refusing more than the bare minimum is acceptable
// HERE precisely because the refusals are named: every RefusalID below is
// one of the three new IDs this plan introduces, following the
// check.loan_liveness_bound_exceeded naming convention.
// ---------------------------------------------------------------------

// The three new refusal IDs (D-11-35), following
// check.loan_liveness_bound_exceeded's own naming convention.
const (
	// RefusalSeedShapeUnsupported names a core.Program shape Reduce's own
	// move set declines to transform (as opposed to a SOURCE projection
	// refusal, which is RefusalProjectionUnsupported below).
	RefusalSeedShapeUnsupported = "reduce.seed_shape_unsupported"
	// RefusalProjectionUnsupported names every ProjectSource refusal --
	// the machine-checkable tag refusalTag embeds into unsupportedProjection's
	// own reason text, never a second wrapper.
	RefusalProjectionUnsupported = "reduce.projection_unsupported"
	// RefusalReverificationSignatureDrift is reserved for plan 11-09's own
	// strict-field-equality re-verification refusal (D-11-35): reduce's
	// own move set never produces this refusal itself, since
	// re-verification happens downstream of this package's own moves.
	RefusalReverificationSignatureDrift = "reduce.reverification_signature_drift"
)

// refusalTag formats a refusal ID as a machine-checkable prefix for a
// ProjectSource reason string, so a caller (or a test) can recover WHICH
// named refusal fired from the returned comment text without a second,
// parallel API. This is embedded INTO the single reason argument
// unsupportedProjection already wraps -- never a second wrapper around
// its output.
func refusalTag(id string) string {
	return "[" + id + "] "
}

// RefusedShape is one entry in RefusedShapes()' register (D-11-35): a
// shape name, the refusal ID it is declined under, a one-sentence reason,
// and whether the shape can actually be constructed and exercised at this
// language's current maturity. Disclosed == true means the shape is
// either refused upstream of this package (recursive call graphs, refused
// by callgraph.Order before a seed ever reaches Reduce) or is a design
// commitment never to build a mechanism (flaky-predicate tolerance,
// D-11-36) or a reservation for a future plan (D-11-35's
// reverification_signature_drift) -- TestRefusedShapesAreActuallyRefused
// exercises every NON-disclosed entry directly; a disclosed entry's own
// Reason states explicitly why it cannot be constructed here.
type RefusedShape struct {
	Name      string
	RefusalID string
	Reason    string
	Disclosed bool
}

// RefusedShapes returns the enumerated, test-asserted register of shapes
// reduce provably declines (D-11-35). Refusing more than the bare minimum
// is acceptable here precisely because the refusals are named.
func RefusedShapes() []RefusedShape {
	return []RefusedShape{
		{
			Name:      "Match-bodied source projection",
			RefusalID: RefusalProjectionUnsupported,
			Reason:    "ProjectSource has no exhaustive-match source syntax to project a Match-bodied function into (05-10-SUMMARY.md); the move set narrows match arms but never eliminates the Match wrapper down to a re-checkable straight-line projection on its own.",
		},
		{
			Name:      "ADT-typed collapsed match projection",
			RefusalID: RefusalProjectionUnsupported,
			Reason:    "A collapsed ADT-typed match (parameter or return type is a declared data type) has no exhaustive, re-checkable source projection under the current grammar -- there is no partial-match syntax.",
		},
		{
			Name:      "Type-changing calls",
			RefusalID: RefusalSeedShapeUnsupported,
			Reason:    "An OpCall whose argument's own declared type differs from the call's own result type cannot be rewritten to OpCopy without changing a type, so drop-call-site skips it rather than emitting an ill-typed copy.",
		},
		{
			Name:      "Recursive call graphs",
			RefusalID: RefusalSeedShapeUnsupported,
			Reason:    "A cyclic call graph is refused upstream by callgraph.Order's own cycle check (SEM-07, core.call_graph_cycle) before a seed ever reaches Reduce; reduce has no independent cycle detector of its own to exercise.",
			Disclosed: true,
		},
		{
			Name:      "Flaky or nondeterministic predicates",
			RefusalID: RefusalSeedShapeUnsupported,
			Reason:    "Tolerance for a flaky predicate -- re-invoking it on rejection, accepting a candidate only when most of several invocations agree, or loosening the interestingness comparison to an approximate match -- is explicitly not built (D-11-36); a flaky anti-vacuity control is a criterion-3 defect to escalate, never a reduce-level refusal to construct.",
			Disclosed: true,
		},
		{
			Name:      "Re-verification signature drift (reserved for plan 11-09)",
			RefusalID: RefusalReverificationSignatureDrift,
			Reason:    "Reserved for plan 11-09's own strict-field-equality re-verification refusal; reduce's own move set never produces this refusal itself.",
			Disclosed: true,
		},
	}
}

func projectStraightLineSource(program core.Program, fn core.Function) string {
	nameByID := map[string]string{fn.Parameter.ID: fn.Parameter.Name}
	for _, place := range fn.Linear.Places {
		nameByID[place.ID] = place.Name
	}
	var body strings.Builder
	terminal := "  " + fn.Parameter.Name + "\n"
	for _, op := range fn.Linear.Operations {
		switch op.Kind {
		case core.OpBorrowShared:
			fmt.Fprintf(&body, "  let %s = borrow %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpBorrowExclusive:
			fmt.Fprintf(&body, "  let %s = borrow mut %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpMove:
			fmt.Fprintf(&body, "  let %s = take %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpCopy:
			fmt.Fprintf(&body, "  let %s = %s\n", nameByID[op.TargetID], nameByID[op.SourceID])
		case core.OpReturn:
			terminal = "  " + nameByID[op.SourceID] + "\n"
		case core.OpDefect:
			terminal = fmt.Sprintf("  defect %q\n", op.Reason)
		}
	}
	dataDecl := renderDataType(program, fn.Parameter.Type)
	if fn.ReturnType != fn.Parameter.Type {
		dataDecl += renderDataType(program, fn.ReturnType)
	}
	var out strings.Builder
	fmt.Fprintf(&out, "module %s\n\nexport {\n  fn %s\n}\n\n%sfn %s(%s: %s) -> %s {\n%s%s}\n",
		program.Module, fn.Name, dataDecl, fn.Name, fn.Parameter.Name, fn.Parameter.Type, fn.ReturnType, body.String(), terminal)
	return out.String()
}

// renderDataType emits a `data Name = | Alt ...` declaration for typeName
// when it names one of program.DataTypes (a user-declared ADT, e.g. a
// match scrutinee's type) -- omitted entirely for a built-in primitive
// type (Byte/Buffer), which never appears in program.DataTypes.
func renderDataType(program core.Program, typeName string) string {
	for _, dataType := range program.DataTypes {
		if dataType.Name != typeName || len(dataType.Alternatives) == 0 {
			continue
		}
		var out strings.Builder
		fmt.Fprintf(&out, "data %s =\n", dataType.Name)
		for _, alt := range dataType.Alternatives {
			fmt.Fprintf(&out, "  | %s\n", alt)
		}
		out.WriteString("\n")
		return out.String()
	}
	return ""
}

func projectForeignChainSource(program core.Program, fn core.Function) string {
	contract := fn.ForeignContract
	nameByID := map[string]string{fn.Parameter.ID: fn.Parameter.Name}
	for _, place := range fn.Linear.Places {
		nameByID[place.ID] = place.Name
	}
	var body strings.Builder
	resultName := fn.Parameter.Name
	for _, op := range fn.Linear.Operations {
		switch op.Kind {
		case core.OpForeignCall:
			fmt.Fprintf(&body, "  let %s = try %s(%s)\n", nameByID[op.TargetID], contract.Symbol, nameByID[op.SourceID])
		case core.OpReturn:
			resultName = nameByID[op.SourceID]
		}
	}
	if body.Len() == 0 {
		return unsupportedProjection("foreign-chain body has no surviving OpForeignCall step")
	}

	failName, failAlt := failDataType(program, contract.Fails)

	var foreign strings.Builder
	fmt.Fprintf(&foreign, "foreign C {\n\n  fn %s(%s: %s) -> %s {\n", contract.Symbol, fn.Parameter.Name, fn.Parameter.Type, fn.Parameter.Type)
	if contract.Unwind != "" {
		fmt.Fprintf(&foreign, "    unwind: %s\n", contract.Unwind)
	}
	if contract.NonlocalExit != "" {
		fmt.Fprintf(&foreign, "    nonlocal_exit: %s\n", contract.NonlocalExit)
	}
	if contract.Allocator != "" {
		fmt.Fprintf(&foreign, "    allocator: %q\n", contract.Allocator)
	}
	if failName != "" {
		fmt.Fprintf(&foreign, "    fails: %s\n", failName)
	}
	foreign.WriteString("  }\n}\n")

	var out strings.Builder
	fmt.Fprintf(&out, "module %s\n\nexport {\n  fn %s\n}\n\n%s\ndata %s =\n  | %s\n\nfn %s(%s: %s) -> %s {\n%s  %s\n}\n",
		program.Module, fn.Name, foreign.String(), failName, failAlt, fn.Name, fn.Parameter.Name, fn.Parameter.Type, fn.ReturnType, body.String(), resultName)
	return out.String()
}

func failDataType(program core.Program, name string) (string, string) {
	for _, dataType := range program.DataTypes {
		if dataType.Name == name && len(dataType.Alternatives) > 0 {
			return dataType.Name, dataType.Alternatives[0]
		}
	}
	if name == "" {
		return "", ""
	}
	return name, name + "Failed"
}

// ---------------------------------------------------------------------
// Shared structural helpers
// ---------------------------------------------------------------------

func cloneProgram(p core.Program) core.Program {
	encoded, err := json.Marshal(p)
	if err != nil {
		return p
	}
	var out core.Program
	if err := json.Unmarshal(encoded, &out); err != nil {
		return p
	}
	return out
}

func isPureValueKind(k core.OperationKind) bool {
	switch k {
	case core.OpBorrowShared, core.OpBorrowExclusive, core.OpCopy, core.OpMove:
		return true
	}
	return false
}

func isTerminatorKind(kind core.OperationKind) bool {
	for _, k := range core.TerminatorKinds() {
		if k == kind {
			return true
		}
	}
	return false
}

func usedSourceIDs(fn core.Function) map[string]bool {
	used := map[string]bool{}
	if fn.Linear == nil {
		return used
	}
	for _, op := range fn.Linear.Operations {
		if op.SourceID != "" {
			used[op.SourceID] = true
		}
	}
	return used
}

func removeOperationAt(fn *core.Function, index int) {
	op := fn.Linear.Operations[index]
	fn.Linear.Operations = append(append([]core.LinearOperation{}, fn.Linear.Operations[:index]...), fn.Linear.Operations[index+1:]...)
	if op.TargetID != "" {
		fn.Linear.Places = removePlace(fn.Linear.Places, op.TargetID)
	}
	for bi := range fn.Linear.Blocks {
		fn.Linear.Blocks[bi].OperationIDs = removeString(fn.Linear.Blocks[bi].OperationIDs, op.ID)
	}
}

func removePlace(places []core.Place, id string) []core.Place {
	if id == "" {
		return places
	}
	out := make([]core.Place, 0, len(places))
	for _, place := range places {
		if place.ID != id {
			out = append(out, place)
		}
	}
	return out
}

func removeString(values []string, value string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v != value {
			out = append(out, v)
		}
	}
	return out
}

func findBlock(blocks []core.Block, id string) *core.Block {
	for i := range blocks {
		if blocks[i].ID == id {
			return &blocks[i]
		}
	}
	return nil
}

func opsByIDs(ops []core.LinearOperation, ids []string) []core.LinearOperation {
	byID := make(map[string]core.LinearOperation, len(ops))
	for _, op := range ops {
		byID[op.ID] = op
	}
	out := make([]core.LinearOperation, 0, len(ids))
	for _, id := range ids {
		if op, ok := byID[id]; ok {
			out = append(out, op)
		}
	}
	return out
}

func removeOpsByID(ops []core.LinearOperation, ids []string) []core.LinearOperation {
	remove := make(map[string]bool, len(ids))
	for _, id := range ids {
		remove[id] = true
	}
	out := make([]core.LinearOperation, 0, len(ops))
	for _, op := range ops {
		if !remove[op.ID] {
			out = append(out, op)
		}
	}
	return out
}

func removeBlock(blocks []core.Block, id string) []core.Block {
	out := make([]core.Block, 0, len(blocks))
	for _, b := range blocks {
		if b.ID != id {
			out = append(out, b)
		}
	}
	return out
}

func removeEdgesTouching(edges []core.Edge, blockID string) []core.Edge {
	out := make([]core.Edge, 0, len(edges))
	for _, e := range edges {
		if e.FromBlockID == blockID || e.ToBlockID == blockID {
			continue
		}
		out = append(out, e)
	}
	return out
}

func removeIndices(ops []core.LinearOperation, indices []int) []core.LinearOperation {
	remove := make(map[int]bool, len(indices))
	for _, i := range indices {
		remove[i] = true
	}
	out := make([]core.LinearOperation, 0, len(ops)-len(indices))
	for i, op := range ops {
		if !remove[i] {
			out = append(out, op)
		}
	}
	return out
}

func scrubBlocks(blocks []core.Block, opIDs []string) []core.Block {
	remove := make(map[string]bool, len(opIDs))
	for _, id := range opIDs {
		remove[id] = true
	}
	out := make([]core.Block, len(blocks))
	for i, b := range blocks {
		filtered := make([]string, 0, len(b.OperationIDs))
		for _, id := range b.OperationIDs {
			if !remove[id] {
				filtered = append(filtered, id)
			}
		}
		b.OperationIDs = filtered
		out[i] = b
	}
	return out
}

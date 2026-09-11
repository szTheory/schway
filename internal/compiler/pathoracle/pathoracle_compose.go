package pathoracle

import (
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// CalleeLookup is the narrowest possible callee-resolution capability
// composition needs (D-10-13): given a callee's declared function ID,
// return its own core.Function (the same Blocks/Edges/Operations view
// RecomputeEndpoints already trusts for its top-level function), or false
// if unresolved. Deliberately NOT a bare whole-program value -- composition needs
// one function at a time, keyed by ID, and nothing else a fuller capability
// would expose. This is the same SHAPE as originvalidate's own
// calleeOriginFact widening (D-10-13's own independent-but-matching
// decision): least-privilege parameter passing at both re-derivers.
type CalleeLookup func(functionID string) (core.Function, bool)

// MaxCompositionDepth is QLT-04's declared, bounded composition depth: the
// number of core.OpCall hops a single loan-or-published-origin's carry
// relation crosses before reaching its terminal use -- a chain of d+1
// functions (D-10-45). This definition is ADOPTED, not invented: it is
// already load-bearing at corevalidate_peer_liveness.go:64-77
// (derivePeerLoanCarry's own doc comment, "depth-2, depth-3, and beyond all
// resolve correctly from the SAME single forward pass") and in the
// relay_depthN_* fixture-naming convention Phases 08 and 09 established.
//
// Why not less than 2 (D-10-46): depth 2 is the established NECESSITY
// FLOOR. D-09-49 Q2 (09-CONTEXT.md) proved depth-1-only evidence is
// insufficient to demonstrate transitivity across an arbitrary number of
// call hops -- a depth-1 corpus cannot distinguish "the derivation reads a
// callee's own recursively-composed answer" from "it happens to work for
// the one-hop case" (the same generalization Phase 07's own
// diamond-vs-chain cycle-mutation finding required, 07-06-SUMMARY.md).
// QLT-04 cannot declare less than 2 without contradicting that prior
// finding.
//
// Why 3, not merely 2 (D-10-46): necessity-minimum plus one sufficiency
// margin, in the spirit of CBMC's `--unwind N` plus an unwinding assertion
// -- a bound AND a check that raising it finds nothing new. What forced 3
// specifically (D-10-47): before this phase there was NO depth-3 fixture
// anywhere in this tree (`grep -rn 'depth3\|depth_3'` over testdata/ and
// internal/ returned nothing), while corevalidate_peer_liveness.go's own
// prose already asserted depth-2, depth-3, and beyond all resolve from the
// same pass -- prose with no fixture behind it until plan 10-07's own
// relay_depth3_accept.lang/relay_depth3_refuse.lang pair and
// TestCompositionDepthCorpusReachesDeclaredBound's bidirectional gate
// (D-10-50) closed that gap.
//
// The direction, contrasted with MaxPaths (D-10-12): MaxPaths (pathoracle.go)
// sits deliberately ABOVE its real reachable maximum (4096 declared against
// a 64-arm real ceiling) so it never fires on any program the parser can
// produce today -- a purely defensive ceiling. MaxCompositionDepth sits AT
// the declared depth instead: a genuinely reachable bound, not a ceiling
// held safely out of reach. The two constants therefore report which one
// fired independently (distinct Code()s), and a reader must not
// pattern-match one constant's rationale onto the other.
//
// The product space this constant's corpus exercises is small and
// genuinely exhaustible today (T-10-04's boundary-crossing case space):
//
//	Dimension                  | Cardinality | Note
//	ParameterContract.Mode     | 1           | NAMED EXCLUSION (D-07-01):
//	                                           unreachable at anything but
//	                                           "owned"; arity fixed at 1
//	                                           (core.go:245-256, the plan's
//	                                           own 191-198 citation is stale
//	                                           against the current file --
//	                                           corrected here, mirroring
//	                                           D-10-40's own stale-citation
//	                                           precedent). The whole
//	                                           dimension collapses -- this is
//	                                           STATED, never a silently
//	                                           pruned dimension.
//	ReturnContract.Mode        | 3           | three legal values
//	                                           (core.go:270-285)
//	LoanEndpoint.Kind          | 2           | "point" / "edge"
//	                                           (core.go:790-797)
//	verdict                    | 2           | accept / refuse
//
// Roughly a dozen structurally distinct boundary-crossing cases, not
// thousands.
const MaxCompositionDepth = 3

// compositionDepthError is returned when composing across core.OpCall would
// recurse deeper than MaxCompositionDepth call frames. Mirrors
// pathCapError's shape exactly (D-10-12): a distinct Code() so a consumer
// can tell which cap fired.
type compositionDepthError struct {
	functionID string
	limit      int
}

func (e *compositionDepthError) Error() string {
	return fmt.Sprintf("pathoracle.composition_depth_exceeded: composing into function %q would exceed the declared cap of %d call frames", e.functionID, e.limit)
}

// Code reports the stable diagnostic code for a composition-depth-cap
// rejection -- distinct from pathCapError's "pathoracle.path_count_exceeded"
// so the two caps are independently identifiable.
func (e *compositionDepthError) Code() string { return "pathoracle.composition_depth_exceeded" }

// CompositionDepthError type-asserts err as the composition-depth-cap
// rejection, mirroring PathCapError's own extractor shape exactly.
func CompositionDepthError(err error) (*compositionDepthError, bool) {
	e, ok := err.(*compositionDepthError)
	return e, ok
}

// compositionCycleError is composition's OWN guard against mutual
// recursion across the call edge (D-10-15), distinct from EnumeratePaths's
// existing intra-function back-edge handling: that guard catches a cycle
// WITHIN one function's own declared Blocks/Successors; this one catches a
// cycle ACROSS function boundaries during composition (function A calls B
// calls A). This is fail-closed defence for a CORRUPTED core artifact only:
// callgraph.Order (internal/compiler/callgraph) already refuses a
// call-graph cycle before a checked program is ever admitted, so this guard is
// never expected to fire against a correctly-produced artifact -- it exists
// so a synthetic or corrupted core.Function handed to this package directly
// (bypassing callgraph entirely, as several of this package's own tests do)
// still fails closed rather than recursing forever.
type compositionCycleError struct{ functionID string }

func (e *compositionCycleError) Error() string {
	return fmt.Sprintf("pathoracle.composition_cycle: function %q participates in a call cycle during composition", e.functionID)
}

func (e *compositionCycleError) Code() string { return "pathoracle.composition_cycle" }

// CompositionCycleError type-asserts err as the composition cycle-guard
// rejection.
func CompositionCycleError(err error) (*compositionCycleError, bool) {
	e, ok := err.(*compositionCycleError)
	return e, ok
}

// forceContractHopForTest is the D-10-55 seeded-fault seam. Production
// composition always recurses into the callee and re-derives, PER CALLEE
// PATH, whether that path's own return value carries a loan the callee
// itself created (D-10-09). When this func is non-nil, composeCall instead
// consults it as a stubbed contract-hop fact -- mirroring
// peerLoanCarryFact.ReturnsBorrowOfParam's per-FUNCTION shape
// (corevalidate_peer_liveness.go) -- and reports the SAME single answer for
// every one of the callee's paths, regardless of which concrete path would
// actually be taken: this collapses the per-path split real composition
// preserves (D-10-14). Deliberately unexported with no exported
// Set/Force/Disable name in this file: package pathoracle_test (or any
// other importer) cannot reach it directly, only this package's own
// export_test.go bridge can -- Go's package boundary is the actual
// independence proof, not merely an assertion that happens to pass. Never
// seed at a point check, corevalidate, or originvalidate could ever read
// from; this seam is reachable only from within package pathoracle.
var forceContractHopForTest func(calleeID string) bool

// composeCall re-derives, for one core.OpCall operation, the distinct
// answers the callee's OWN concrete paths give to "does this path's return
// value carry a loan the callee itself created" (D-10-09/D-10-11): it
// recurses into the callee via this package's OWN EnumeratePaths, the SAME
// mechanism this package already uses for its top-level function, and asks
// per callee path whether a loan the callee created via its own
// core.OpBorrowShared/core.OpBorrowExclusive survives to that path's
// core.OpReturn. It never consults the callee's DECLARED return contract
// (unlike check.derivePlaceLoans / corevalidate.buildLoanChainIndex's
// per-function summary hop): a callee whose declared type claims a borrow
// but whose own concrete body never creates one reports false here, and a
// callee that creates one on only SOME of its paths reports differently per
// path -- exactly the distinction D-10-14 says a contract hop cannot make.
// The returned slice has one entry per callee path (EnumeratePaths' own
// deterministic order); work is the oracle's own counted work performed
// composing this one call.
func composeCall(operation core.LinearOperation, lookupCallee CalleeLookup, depth int, callStack map[string]bool) (carries []bool, work int, err error) {
	if forceContractHopForTest != nil {
		return []bool{forceContractHopForTest(operation.CalleeID)}, 0, nil
	}
	if depth >= MaxCompositionDepth {
		return nil, 0, &compositionDepthError{functionID: operation.CalleeID, limit: MaxCompositionDepth}
	}
	if callStack[operation.CalleeID] {
		return nil, 0, &compositionCycleError{functionID: operation.CalleeID}
	}
	if lookupCallee == nil {
		// No callee-lookup capability was supplied at all: fail closed to
		// "does not carry", exactly like corevalidate's own unknown-callee
		// default (T-09-02) -- never over-approximate in the permitting
		// direction merely because composition was not wired up.
		return []bool{false}, 0, nil
	}
	callee, ok := lookupCallee(operation.CalleeID)
	if !ok || callee.Linear == nil || len(callee.Linear.Blocks) == 0 {
		// An unresolved callee (a corrupted artifact) or a bodyless one
		// (production never populates Blocks for a straight-line function,
		// 03-03's documented scope decision, mirrored here) carries no
		// loan of its own into the caller.
		return []bool{false}, 0, nil
	}

	calleeIdx := indexBlocks(callee)
	entryBlockID := ""
	for _, block := range callee.Linear.Blocks {
		if block.PointID == callee.EntryPointID {
			entryBlockID = block.ID
			break
		}
	}
	if entryBlockID == "" {
		return nil, 0, &noEntryBlockError{functionID: callee.ID}
	}

	calleeBlockPaths, err := EnumeratePaths(callee.ID, entryBlockID, calleeIdx, MaxPaths)
	if err != nil {
		return nil, 0, err
	}

	nextStack := make(map[string]bool, len(callStack)+1)
	for id := range callStack {
		nextStack[id] = true
	}
	nextStack[operation.CalleeID] = true

	total := 0
	carries = make([]bool, 0, len(calleeBlockPaths))
	for _, calleeBlockIDs := range calleeBlockPaths {
		carriesOwnLoan, w, err := composeCarriesOwnLoan(callee, calleeIdx, calleeBlockIDs, lookupCallee, depth, nextStack)
		total += w
		if err != nil {
			return nil, total, err
		}
		carries = append(carries, carriesOwnLoan)
	}
	return carries, total, nil
}

// composeCarriesOwnLoan walks ONE concrete callee path forward and reports
// whether the value flowing into that path's own core.OpReturn traces
// (through the SAME forward place-inheritance chain linearizePath already
// uses) back to a loan the callee itself created on this path via
// core.OpBorrowShared/core.OpBorrowExclusive. It is deliberately NOT seeded
// with anything the caller's own argument carries (D-10-11): composition's
// cross-boundary contribution is using THIS per-path fact as the gate for
// whether the CALLER's own place-chain propagates across the call boundary
// -- composeLinearizeCaller does that gating -- mirroring
// derivePlaceLoans'/buildLoanChainIndex's own per-function gate, except
// re-derived per callee path instead of trusted from a declared contract.
// A nested core.OpCall on the callee's own path recurses through
// composeCall again at depth+1, so composition genuinely nests to
// MaxCompositionDepth frames rather than stopping at one hop.
func composeCarriesOwnLoan(callee core.Function, idx blockIndex, blockIDs []string, lookupCallee CalleeLookup, depth int, callStack map[string]bool) (bool, int, error) {
	chain := map[string]bool{} // placeID -> traces to a loan created on this path
	work := 0
	sawTerminator := false
	returnValueID := ""

	for _, blockID := range blockIDs {
		for _, operation := range idx.operations[blockID] {
			work++
			if isTerminatorKind(operation.Kind) {
				sawTerminator = true
			}
			switch operation.Kind {
			case core.OpBorrowShared, core.OpBorrowExclusive:
				if operation.TargetID != "" {
					chain[operation.TargetID] = true
				}
			case core.OpCall:
				nestedCarries, w, err := composeCall(operation, lookupCallee, depth+1, callStack)
				work += w
				if err != nil {
					return false, work, err
				}
				anyCarry := chain[operation.SourceID]
				for _, c := range nestedCarries {
					if c {
						anyCarry = true
						break
					}
				}
				if anyCarry && operation.TargetID != "" {
					chain[operation.TargetID] = true
				}
			default:
				if chain[operation.SourceID] && operation.TargetID != "" {
					chain[operation.TargetID] = true
				}
			}
			if operation.Kind == core.OpReturn {
				returnValueID = operation.SourceID
			}
		}
	}

	if !sawTerminator {
		return false, work, &unterminatedLoanError{functionID: callee.ID, loanID: "<composition>"}
	}
	return chain[returnValueID], work, nil
}

// composedVariant is one concrete way a caller path can conclude once every
// core.OpCall on it has been resolved to one of its callee's own concrete
// answers (D-10-09): the caller's own loanState map for this specific
// resolution, exactly the shape linearizePath already returns for a
// call-free path.
type composedVariant struct {
	states map[string]loanState
}

func cloneChain(chain map[string][]string) map[string][]string {
	clone := make(map[string][]string, len(chain))
	for k, v := range chain {
		clone[k] = append([]string(nil), v...)
	}
	return clone
}

func cloneLoanStates(states map[string]loanState) map[string]loanState {
	clone := make(map[string]loanState, len(states))
	for k, v := range states {
		clone[k] = v
	}
	return clone
}

// hasOpCall reports whether any operation along blockIDs is a core.OpCall --
// RecomputeEndpoints uses this to decide, per top-level caller path,
// whether the ordinary call-free linearizePath suffices (byte-identical
// behavior to before this plan for every such path) or whether the
// composing replay below is needed.
func hasOpCall(idx blockIndex, blockIDs []string) bool {
	for _, blockID := range blockIDs {
		for _, operation := range idx.operations[blockID] {
			if operation.Kind == core.OpCall {
				return true
			}
		}
	}
	return false
}

// composeLinearizeCaller replays ONE caller path forward exactly like
// linearizePath, except at each core.OpCall it forks into one variant per
// distinct answer composeCall reports for that call (D-10-09): whether the
// callee's OWN concrete path (never a declared contract) carries a loan
// into its return gates whether the CALLER's own place-chain propagates
// across the call boundary, mirroring derivePlaceLoans'/
// buildLoanChainIndex's per-function gate, re-derived per callee path
// instead of trusted from a declared contract. Forward replay still runs
// exactly ONCE over each concretely spliced continuation (D-10-09/D-10-11):
// no convergence loop, fixpoint, or relation closure is introduced anywhere
// here -- every fork is a genuine, distinct concrete continuation, and each
// one is replayed forward exactly once.
func composeLinearizeCaller(functionID string, idx blockIndex, blockIDs []string, lookupCallee CalleeLookup) ([]composedVariant, int, error) {
	type liveVariant struct {
		chain  map[string][]string
		states map[string]loanState
	}
	variants := []liveVariant{{chain: map[string][]string{}, states: map[string]loanState{}}}
	work := 0
	sawTerminator := false

	for _, blockID := range blockIDs {
		operations := idx.operations[blockID]
		for opIndex, operation := range operations {
			work++
			if isTerminatorKind(operation.Kind) {
				sawTerminator = true
			}

			if operation.Kind == core.OpCall {
				carries, w, err := composeCall(operation, lookupCallee, 0, map[string]bool{functionID: true})
				work += w
				if err != nil {
					return nil, work, err
				}
				next := make([]liveVariant, 0, len(variants)*len(carries))
				for _, v := range variants {
					inherited := v.chain[operation.SourceID]
					for _, loanID := range inherited {
						state, known := v.states[loanID]
						if !known {
							return nil, work, &unterminatedLoanError{functionID: functionID, loanID: loanID}
						}
						state.lastBlockID, state.lastIndex, state.lastOpID = blockID, opIndex, operation.ID
						v.states[loanID] = state
					}
					for _, carriesLoan := range carries {
						forked := liveVariant{chain: cloneChain(v.chain), states: cloneLoanStates(v.states)}
						if carriesLoan && operation.TargetID != "" && len(inherited) > 0 {
							forked.chain[operation.TargetID] = append([]string(nil), inherited...)
						}
						next = append(next, forked)
					}
				}
				variants = next
				continue
			}

			for i := range variants {
				v := &variants[i]
				inherited := v.chain[operation.SourceID]
				for _, loanID := range inherited {
					state, known := v.states[loanID]
					if !known {
						return nil, work, &unterminatedLoanError{functionID: functionID, loanID: loanID}
					}
					state.lastBlockID, state.lastIndex, state.lastOpID = blockID, opIndex, operation.ID
					v.states[loanID] = state
				}
				if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
					v.states[operation.LoanID] = loanState{bornBlockID: blockID, lastBlockID: blockID, lastIndex: opIndex, lastOpID: operation.ID}
					inherited = append(append([]string(nil), inherited...), operation.LoanID)
				}
				if operation.TargetID != "" && len(inherited) > 0 {
					v.chain[operation.TargetID] = inherited
				}
			}
		}
	}

	if !sawTerminator {
		for _, v := range variants {
			if len(v.states) > 0 {
				loanIDs := make([]string, 0, len(v.states))
				for id := range v.states {
					loanIDs = append(loanIDs, id)
				}
				sort.Strings(loanIDs)
				return nil, work, &unterminatedLoanError{functionID: functionID, loanID: loanIDs[0]}
			}
		}
	}

	result := make([]composedVariant, len(variants))
	for i, v := range variants {
		result[i] = composedVariant{states: v.states}
	}
	return result, work, nil
}

// synthesizeComposedEndpoints handles a caller path whose composition
// across an OpCall produced more than one concrete continuation (D-10-14):
// a loan whose composed variants disagree on whether it survives past the
// call cannot be reported as a single, agreed death position the way
// RecomputeEndpoints's ordinary same-path bookkeeping assumes (that
// assumption is exactly what pathoracle.inconsistent_path_death exists to
// enforce for a genuine bug elsewhere -- composed divergence is not that).
// The variant(s) that keep propagating the loan reach its own eventual,
// LATEST death position along this caller path -- reported as an ordinary
// point endpoint, exactly like any same-path death. Every OTHER, EARLIER
// death position some composed variant reaches is reported as an edge
// endpoint at a composition-synthetic edge identity, distinct per distinct
// early position -- the per-path split TRU-03 requires and a contract hop
// cannot express (D-10-14).
func synthesizeComposedEndpoints(functionID string, blockIDs []string, variants []map[string]loanState) []core.LoanEndpoint {
	positionOf := func(blockID string) int {
		for i, id := range blockIDs {
			if id == blockID {
				return i
			}
		}
		return -1
	}

	loanIDSet := map[string]bool{}
	for _, states := range variants {
		for loanID := range states {
			loanIDSet[loanID] = true
		}
	}
	loanIDs := make([]string, 0, len(loanIDSet))
	for id := range loanIDSet {
		loanIDs = append(loanIDs, id)
	}
	sort.Strings(loanIDs)

	var endpoints []core.LoanEndpoint
	for _, loanID := range loanIDs {
		distinct := map[[2]int]loanState{}
		var order [][2]int
		for _, states := range variants {
			state, ok := states[loanID]
			if !ok {
				continue
			}
			key := [2]int{positionOf(state.lastBlockID), state.lastIndex}
			if _, seen := distinct[key]; !seen {
				distinct[key] = state
				order = append(order, key)
			}
		}
		if len(order) == 0 {
			continue
		}
		sort.Slice(order, func(i, j int) bool {
			if order[i][0] != order[j][0] {
				return order[i][0] < order[j][0]
			}
			return order[i][1] < order[j][1]
		})

		latestKey := order[len(order)-1]
		latest := distinct[latestKey]
		endpoints = append(endpoints, core.LoanEndpoint{
			ID:               fmt.Sprintf("%s:point:%s:%d:%s", functionID, latest.lastBlockID, latest.lastIndex, loanID),
			LoanID:           loanID, Kind: "point", BlockID: latest.lastBlockID, AfterOperationID: latest.lastOpID,
		})

		for i, key := range order[:len(order)-1] {
			earlier := distinct[key]
			edgeID := fmt.Sprintf("%s:compose-edge:%s:%d:%d", functionID, earlier.lastBlockID, earlier.lastIndex, i)
			endpoints = append(endpoints, core.LoanEndpoint{
				ID: edgeID + ":" + loanID, LoanID: loanID, Kind: "edge", EdgeID: edgeID,
			})
		}
	}
	return endpoints
}

// Package corevalidate independently validates inert typed-core facts before
// an execution engine consumes them. It intentionally does not know source or
// reuse checker/interpreter authorization code.
package corevalidate

import (
	"crypto/sha256"
	"encoding/hex"
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
//
// The +1*facts over the pre-Phase-07 formula (16*facts+14) is D-07-29's
// CalleeID kind-exclusivity check: one new v.check per operation, run
// unconditionally regardless of kind, confirming a non-OpCall operation
// (every operation in this scale shape) leaves CalleeID empty.
//
// The +1 flat term (from 17*facts+14 to 17*facts+15) is 07-07's own
// whole-program cycle peer (checkCallGraphAcyclic): it runs exactly ONCE
// per Validate call, not once per fact/operation, so it is a constant
// addition regardless of facts -- mechanical, expected, and not a
// per-operation work-formula change (mirrors the precedent already
// recorded at this file's own TestCoreValidationWorkSeries call site).
func LinearWorkLimit(facts int) int { return 17*facts + 15 }

type Problem struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

type Result struct {
	Valid    bool      `json:"valid"`
	Problems []Problem `json:"problems,omitempty"`
	Checks   int       `json:"checks"`
	program  core.Program

	// peerSignatures/peerRanStraightLine/peerRanBlocks carry 07-02's
	// independent (D-07-20/D-07-22) structural summary-peer results --
	// see PeerSignatures/PeerSiteCoverage below.
	peerSignatures      map[string]core.FunctionSignature
	peerRanStraightLine bool
	peerRanBlocks       bool
}

// Program returns a content-owned copy of the validated program. Callers can
// neither mutate the validator's copy nor race validation by retaining slices.
func (r Result) Program() core.Program { return cloneProgram(r.program) }

// PeerSignatures returns corevalidate's own independent (D-07-20/D-07-22)
// structural re-derivation of every function's lang.interface/1 signature
// summary, keyed by function ID. This is NOT originvalidate.BuildInterface's
// output -- it is a materially different, separately-written computation
// over the SAME checked core.Program, read from the validator's own
// already-replayed type/place facts rather than a fresh exact-ID type-fact
// lookup (corevalidate.go's doc-comment template, corevalidate.go:352-361).
// A test harness with access to both packages compares this against
// originvalidate.BuildInterface(program); this package never imports
// originvalidate to make that comparison itself. Callers get a fresh copy of
// the map on every call.
func (r Result) PeerSignatures() map[string]core.FunctionSignature {
	out := make(map[string]core.FunctionSignature, len(r.peerSignatures))
	for id, signature := range r.peerSignatures {
		out[id] = signature
	}
	return out
}

// PeerSiteCoverage reports whether the summary peer actually ran at
// replayStraightLine and at replayBlocks during this Validate call (D-07-21).
// A test asserts BOTH are true when the input program exercises both replay
// shapes, rather than inferring coverage from which fixtures happened to be
// supplied -- wiring the peer into only one site is the literal D-02-03/
// D-03-01 repeat this instrumentation exists to catch.
func (r Result) PeerSiteCoverage() (straightLine, blocks bool) {
	return r.peerRanStraightLine, r.peerRanBlocks
}

// Validate is deliberately source-blind. It proves internal consistency of a
// typed-core statement, not that a coordinated producer translated source
// truthfully.
func Validate(input core.Program) Result {
	owned := cloneProgram(input)
	v := validator{program: owned}
	v.run()
	return Result{
		Valid: len(v.problems) == 0, Problems: v.problems, Checks: v.checks, program: owned,
		peerSignatures:      v.peerSignatures,
		peerRanStraightLine: v.peerRanStraightLine,
		peerRanBlocks:       v.peerRanBlocks,
	}
}

type validator struct {
	program  core.Program
	checks   int
	problems []Problem

	// peerSignatures/peerRanStraightLine/peerRanBlocks are populated by
	// recordSummaryPeer, called from BOTH replayStraightLine and
	// replayBlocks (D-07-21) -- see Result.PeerSignatures/PeerSiteCoverage.
	peerSignatures      map[string]core.FunctionSignature
	peerRanStraightLine bool
	peerRanBlocks       bool

	// peerAdjacency/peerPostorder are checkCallGraphAcyclic's own byproduct
	// (07-08, D-07-38/D-07-22): the SAME deduped adjacency it already
	// builds to prove acyclicity, plus the raw DFS postorder (callee
	// FINISHES before caller) its traversal already produces for free --
	// chainPeerClosureDigests below consumes both to chain ClosureDigest
	// over corevalidate's OWN, independently-derived route, never
	// callgraph's or originvalidate's.
	peerAdjacency map[string][]string
	peerPostorder []string
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

// pendingReplayEntry defers one function's per-function replay (D-07-19/
// T-07-46) until every declared function's own structural facts have been
// independently validated, so run()'s whole-program cycle peer can run
// once, program-wide, strictly between "all functions structurally valid"
// and "any function replayed" -- see the call site's own comment in run().
type pendingReplayEntry struct {
	function *core.Function
	types    map[string]core.TypeFact
	places   map[string]core.Place
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
	pending := make([]pendingReplayEntry, 0, len(v.program.Functions))
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
			if !v.check(v.program.Schema == core.Schema1, "core.schema", "a match with an arm body requires lang.core/1") {
				return
			}
			types, places, ok := v.matchBranchStructural(function, dataNames)
			if !ok {
				return
			}
			pending = append(pending, pendingReplayEntry{function: function, types: types, places: places})
		case function.Linear != nil:
			if !v.check(v.program.Schema == core.Schema1, "core.schema", "linear body requires lang.core/1") {
				return
			}
			types, places, ok := v.linearStructural(function)
			if !ok {
				return
			}
			pending = append(pending, pendingReplayEntry{function: function, types: types, places: places})
		default:
			if !v.check(v.program.Schema == core.Schema, "core.schema", "match body requires lang.core/0") || !v.match(function, dataNames) {
				return
			}
		}
	}

	// D-07-19/T-07-46: corevalidate's own whole-program cycle peer runs
	// HERE -- after every declared function's own structural facts (types,
	// places, operation identity/order, block/edge referential closure,
	// all validated above by linearStructural/matchBranchStructural) have
	// been independently checked, but strictly BEFORE any function's own
	// per-function replay runs below. Running it earlier would mean
	// traversing a program whose own operations have not yet been checked
	// for internal consistency; running it later (interleaved with, or
	// after, per-function replay) would let an unrelated per-function
	// replay error mask a real cycle, so the very same program could
	// report two different first errors depending purely on function
	// declaration order. See checkCallGraphAcyclic's own doc comment for
	// the peer's independence rationale (D-07-19).
	if !disableCyclePeerForTest {
		if !v.checkCallGraphAcyclic() {
			return
		}
	}

	for _, entry := range pending {
		if !v.replay(entry.function, entry.types, entry.places) {
			return
		}
		if len(entry.function.Linear.Blocks) > 0 {
			// T-03-13/D-12: independently re-derive the declared
			// loan-endpoint set and require whole-value equality against
			// what the producer declared -- see linearStructural's
			// identical historical comment (moved here unchanged by this
			// plan's two-pass split).
			if !v.loanEndpointsMatch(entry.function) {
				return
			}
		}
	}

	// D-07-38/D-07-12 (07-08): chain the peer's own ClosureDigest
	// re-derivation ONLY after every function's peer signature has been
	// recorded (v.match()'s direct call for /0-only functions, plus every
	// pending entry's replayStraightLine/replayBlocks call above) and ONLY
	// on the success path -- run() returns early, above, before ever
	// reaching here, on any structural or replay problem. v.peerPostorder
	// is populated by checkCallGraphAcyclic, so this is gated identically
	// to that call.
	if !disableCyclePeerForTest {
		v.chainPeerClosureDigests()
	}
}

// disableCyclePeerForTest is Task 3's D-07-42 independent-disable seam
// (QLT-08 Test 2/Test 3): when true, corevalidate's own whole-program
// cycle peer (checkCallGraphAcyclic) is skipped entirely, proving check's
// own callgraph-based refusal is wholly independent of this peer ever
// running -- the bilateral row engages this alongside callgraph's own
// disableCycleDetectionForTest to prove the sweep reports "no divergence
// detected under bilateral fault" and fails the gate rather than passing
// it. Exposed to other packages (callgraph_test's cross-package bilateral
// tests) via corevalidate_export_test.go's SetDisableCyclePeerForTest.
// false (the production default) means "run the peer for real".
var disableCyclePeerForTest bool

// SetDisableCyclePeerForTest is Task 3 Test 2/Test 3's cross-package
// bilateral-fault seam. Go's build model excludes "_test.go" files from a
// normal package import, so a same-package-only unexported var (the shape
// every other Phase 07 fault-injection seam in this file uses) cannot be
// reached by check's own bilateral-fault test, which imports this package
// as an ordinary dependency. This is the one, deliberately minimal,
// clearly-named exception (D-07-42): production-visible, but a documented
// test-only no-op unless a test explicitly calls it, and always restored
// via the returned closure. Never called from any production code path in
// this repository.
func SetDisableCyclePeerForTest(disabled bool) (restore func()) {
	previous := disableCyclePeerForTest
	disableCyclePeerForTest = disabled
	return func() { disableCyclePeerForTest = previous }
}

// peerGrayVsVisitedMutationForTest is Task 3 Test 4's own fault-injection
// seam, mirroring callgraph.grayVsVisitedMutationForTest exactly but on
// THIS package's own, independently written traversal (D-07-19): when
// true, a legitimately revisited BLACK (finished, shared) node is wrongly
// treated the same as a GRAY (on-stack) re-entry, so an acyclic diamond
// with a shared leaf is wrongly refused. Proves the peer's own
// gray-versus-visited distinction is independently load-bearing, not
// merely present (T-07-41) -- a second derivation nobody has seen fail is
// not evidence. false (the production default) means "only a gray
// re-entry is treated as a cycle".
var peerGrayVsVisitedMutationForTest bool

// checkCallGraphAcyclic is corevalidate's OWN, independently written
// three-color (white/gray/black) depth-first search over v.program's whole
// call graph (D-07-19). It reads nothing from check's callgraph package --
// TestValidatorImportsStayIndependent's forbidden-suffix list forbids this
// file from ever importing it -- and, in this plan's own tests, is
// exercised EXCLUSIVELY against hand-built synthetic core.Program values
// via syntheticProgram, never through the parser or check. Independence
// here comes from a DISJOINT REACHABLE INPUT SPACE (parser-reachable for
// check, synthetic-only for this peer), never from a different algorithm
// (D-07-19): this traversal mirrors pathoracle.backEdgeError's posture
// exactly -- rather than trusting check.go's own callgraph.Order to have
// caught a cycle first, a corrupted or synthetic core.Program check never
// saw is still refused here.
//
// Roots are ALL declared functions, never merely an entry point: a
// synthetic artifact may declare no entry point at all, and a cycle among
// functions unreachable from any entry point must still be refused
// (fail-closed). Edges come from core.LinearOperation.CalleeID on
// operations with Kind == core.OpCall, enumerated across a function's
// WHOLE Linear.Operations list -- never reconstructed from Block/Successor
// position (D-07-28) -- so a cycle-closing OpCall inside a match arm is
// found exactly like one in a straight-line body. An OpCall whose CalleeID
// resolves to no declared function refuses with core.CallCalleeUnresolved
// (D-07-45) rather than being silently dropped from the graph -- a dropped
// edge is precisely how a cycle escapes detection here too.
//
// It emits the SAME inert string constant core.CallGraphCycle that
// check's callgraph package emits, with no span (D-07-15/D-07-35): the
// code is shared, verbatim, as a string constant; the derivation
// producing it is not. The peer has no spans and needs none -- its job is
// refusal, not diagnostics.
func (v *validator) checkCallGraphAcyclic() bool {
	declared := make(map[string]bool, len(v.program.Functions))
	for _, function := range v.program.Functions {
		declared[function.ID] = true
	}

	adjacency := make(map[string][]string, len(declared))
	for id := range declared {
		adjacency[id] = nil
	}
	// edgeSeen dedupes a caller calling the SAME callee via more than one
	// OpCall (07-08, D-07-38): a duplicate edge changes nothing about
	// cycle presence, but chainPeerClosureDigests below reuses this exact
	// adjacency to build closure-digest callee pairs, and a duplicate
	// pair there would diverge from originvalidate's own deduped pairs
	// (D-07-37) -- mirrors callgraph.buildAdjacency's identical dedup.
	edgeSeen := make(map[[2]string]bool)
	for _, function := range v.program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind != core.OpCall {
				continue
			}
			if operation.CalleeID == "" {
				// D-07-29's CalleeID kind-exclusivity check (an OpCall's
				// CalleeID must be non-empty) is enforced during replay,
				// not by this whole-program peer, which runs before
				// replay and only concerns itself with edges it can
				// actually form. Skip so replay's own dedicated
				// core.callee_id_missing refusal still fires, unmasked.
				continue
			}
			if !declared[operation.CalleeID] {
				if disableCalleeResolutionCheckForTest {
					// Mirrors check.checkCallGraphAcyclic's own identical
					// suppression under verifyCallInvariantsSeam (07-06):
					// this seam's whole point is that corevalidate's
					// resolves-to-a-declared-function re-derivation is
					// suppressed EVERYWHERE it appears in this package,
					// including this newest, independent site, so the one
					// seeded mutation keeps every admission arm it names
					// failing together (TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal).
					continue
				}
				return v.check(false, core.CallCalleeUnresolved, operation.CalleeID)
			}
			key := [2]string{function.ID, operation.CalleeID}
			if edgeSeen[key] {
				continue
			}
			edgeSeen[key] = true
			adjacency[function.ID] = append(adjacency[function.ID], operation.CalleeID)
		}
	}
	// Sorted adjacency (and, below, sorted roots) is what keeps this
	// peer's refuse/accept verdict deterministic under -count=2
	// -shuffle=on, exactly like callgraph.buildAdjacency's identical
	// discipline -- a map-backed traversal cannot pass by luck.
	for id := range adjacency {
		sort.Strings(adjacency[id])
	}
	v.peerAdjacency = adjacency
	roots := make([]string, 0, len(adjacency))
	for id := range adjacency {
		roots = append(roots, id)
	}
	sort.Strings(roots)

	const (
		peerWhite = iota
		peerGray
		peerBlack
	)
	colorOf := make(map[string]int, len(adjacency))
	type peerFrame struct {
		id    string
		index int
	}
	for _, root := range roots {
		if colorOf[root] != peerWhite {
			continue
		}
		stack := []peerFrame{{id: root}}
		colorOf[root] = peerGray
		for len(stack) > 0 {
			top := &stack[len(stack)-1]
			children := adjacency[top.id]
			if top.index < len(children) {
				child := children[top.index]
				top.index++
				switch colorOf[child] {
				case peerWhite:
					colorOf[child] = peerGray
					stack = append(stack, peerFrame{id: child})
				case peerGray:
					// A back edge to an on-stack (gray) node is a real
					// cycle -- including a self-edge (function.ID ==
					// operation.CalleeID), which is caught here too since
					// a root is marked gray before its own children are
					// ever visited.
					return v.check(false, core.CallGraphCycle, child)
				case peerBlack:
					// A legitimately revisited shared node in an acyclic
					// subgraph (a diamond's shared leaf) -- not a cycle.
					// Task 3 Test 4's mutation collapses exactly this
					// distinction.
					if peerGrayVsVisitedMutationForTest {
						return v.check(false, core.CallGraphCycle, child)
					}
				}
			} else {
				colorOf[top.id] = peerBlack
				// 07-08, D-07-38: a node is appended here only once every
				// callee it can reach has already finished (this is plain
				// DFS postorder, callee-before-caller) -- the exact order
				// chainPeerClosureDigests needs to chain bottom-up, free as
				// a byproduct of the SAME traversal that already proves
				// acyclicity, with no second pass over the graph.
				v.peerPostorder = append(v.peerPostorder, top.id)
				stack = stack[:len(stack)-1]
			}
		}
	}
	return v.check(true, core.CallGraphCycle, "")
}

// chainPeerClosureDigests is 07-08's own independent re-derivation of
// D-07-38's chained ClosureDigest (D-07-12/D-07-22): it consumes
// checkCallGraphAcyclic's OWN adjacency and postorder (never callgraph's
// or originvalidate's -- this package imports neither), so a chaining bug
// here is a genuine divergence from the producer, not two callers of one
// implementation. It runs strictly after checkCallGraphAcyclic has proven
// the whole program's call graph acyclic: the chain terminates only on a
// DAG (D-07-38).
//
// v.peerPostorder already lists every function callee-before-caller (see
// the doc comment at its append site above), so this walks it in the
// order recorded -- no reversal needed here, unlike
// originvalidate.BuildInterface's consumption of callgraph.Order's own
// array (which is handed back caller-first, for Phase 08's fixpoint, and
// must be walked backward to recover this same property).
func (v *validator) chainPeerClosureDigests() {
	order := v.peerPostorder
	if closureDigestDiscoveryOrderForTest {
		// Task 2 Test 5's own seam (D-07-41/D-07-42): chain in plain
		// declaration order instead of the proven-correct postorder, so a
		// caller processed before its callee reads that callee's
		// still-empty ClosureDigest -- an observably different digest,
		// proving the ordering is genuinely load-bearing here too, not
		// merely on originvalidate's side.
		order = make([]string, 0, len(v.program.Functions))
		for _, function := range v.program.Functions {
			order = append(order, function.ID)
		}
	}
	for _, id := range order {
		signature, ok := v.peerSignatures[id]
		if !ok {
			continue
		}
		var pairs []peerCalleeDigestPair
		if !closureDigestEmptyCalleesForTest {
			// 07-12 (CR-03/PVG-02): join this caller's own Foreign/Fails
			// with every callee's ALREADY-JOINED peer signature, strictly
			// before peerComputeClosureDigest below -- v.peerPostorder
			// already lists every function callee-before-caller, so each
			// callee's own joined value is already final by the time its
			// caller reads it. This reads v.peerAdjacency/v.peerSignatures
			// (corevalidate's OWN postorder and OWN independently-derived
			// signatures), never originvalidate's chainOrder or
			// calleeIDsForClosureDigest -- a materially different source
			// from the producer's, per CR-03's own coordinated-blindness
			// finding.
			for _, calleeID := range v.peerAdjacency[id] {
				calleeSignature := v.peerSignatures[calleeID]
				pairs = append(pairs, peerCalleeDigestPair{ID: calleeID, ClosureDigest: calleeSignature.ClosureDigest})
				if !disableForeignClosureJoinPeerForTest {
					signature.Foreign = peerJoinForeignReach(signature.Foreign, calleeSignature.Foreign)
				}
				if !disableFailsClosureJoinPeerForTest {
					signature.Fails = peerJoinFails(signature.Fails, calleeSignature.Fails)
				}
			}
		}
		digest, err := peerComputeClosureDigest(signature, pairs)
		if err != nil {
			// Unreachable in practice (json.Marshal over a
			// core.FunctionSignature cannot fail): fail closed by leaving
			// this function's ClosureDigest at its zero value rather than
			// propagating an error run() has no return path for.
			continue
		}
		signature.ClosureDigest = digest
		v.peerSignatures[id] = signature
	}
}

// disableForeignClosureJoinPeerForTest and disableFailsClosureJoinPeerForTest
// are 07-12's D-07-41/D-07-42 fault-injection seams for THIS package's
// independent Foreign/Fails closure join (control:summary.
// foreign_reach_closure_derived / control:summary.fails_closure_derived,
// peer side): unexported, false by production default, restored via defer
// in every test that engages them -- mirroring originvalidate's identically
// purposed foreignClosureJoinSeam/failsClosureJoinSeam on the producer
// side, with no shared implementation.
var (
	disableForeignClosureJoinPeerForTest bool
	disableFailsClosureJoinPeerForTest   bool
)

// peerJoinReachPolicy independently re-derives originvalidate's
// joinReachPolicy (07-12, CR-03/PVG-02): "" is the identity; equal
// non-empty values merge; "forbidden" wins over "permitted" in either
// argument order; any other disagreement -- unreachable while the
// vocabulary stays exactly {"", "permitted", "forbidden"} -- resolves to
// the shared core.ForeignReachConflict schema sentinel. This package
// deliberately shares no helper function with originvalidate (only the
// core.ForeignReachConflict constant itself, a schema value, not a
// derivation) -- see peerJoinForeignReach's own doc comment.
func peerJoinReachPolicy(into, from string) string {
	switch {
	case into == "":
		return from
	case from == "" || into == from:
		return into
	case into == "forbidden" || from == "forbidden":
		return "forbidden"
	default:
		return core.ForeignReachConflict
	}
}

// peerJoinAllocatorName independently re-derives originvalidate's
// joinAllocatorName: "" is the identity; equal non-empty names merge; two
// DIFFERENT non-empty names have no vocabulary ordering and resolve to
// core.ForeignReachConflict.
func peerJoinAllocatorName(into, from string) string {
	switch {
	case into == "":
		return from
	case from == "" || into == from:
		return into
	default:
		return core.ForeignReachConflict
	}
}

// peerJoinForeignReach is corevalidate's OWN, independently written
// worst-case lattice join over core.ForeignReach (07-12, CR-03/PVG-02):
// same SEMANTICS as originvalidate.joinForeignReach (empty is the
// identity; equal merges; forbidden beats permitted; a genuine allocator
// disagreement resolves to core.ForeignReachConflict), different
// implementation, in a different package, sharing no helper function --
// 07-REVIEW.md CR-03 found producer and peer reproducing the SAME local
// read verbatim, so this join is written independently by design, never
// imported from originvalidate (enforced by
// TestForeignClosureJoinPeersIndependent and the grep-based import-
// boundary assertions this package already carries).
func peerJoinForeignReach(into, from core.ForeignReach) core.ForeignReach {
	return core.ForeignReach{
		Allocator:    peerJoinAllocatorName(into.Allocator, from.Allocator),
		Unwind:       peerJoinReachPolicy(into.Unwind, from.Unwind),
		NonlocalExit: peerJoinReachPolicy(into.NonlocalExit, from.NonlocalExit),
	}
}

// peerJoinFails is corevalidate's OWN, independently written join over
// FunctionSignature.Fails (07-12, CR-03/PVG-02): "" is the identity; equal
// merges; two different non-empty values keep the EXISTING (caller-
// nearest) value, mirroring originvalidate.joinFails' semantics exactly
// but written independently. Folded over v.peerAdjacency[id], which is
// NOT sorted by ID the way originvalidate's calleeIDsForClosureDigest is
// (checkCallGraphAcyclic's own adjacency-build sorts each function's
// callee list by ID too -- see its own doc comment -- so this fold is
// equally deterministic in production, by the same sorted-iteration
// discipline, independently applied).
func peerJoinFails(into, from string) string {
	if into == "" {
		return from
	}
	return into
}

// closureDigestDiscoveryOrderForTest and closureDigestEmptyCalleesForTest
// are Task 2's own D-07-41/D-07-42 fault-injection seams for THIS
// package's independent chained-digest re-derivation (mirroring
// originvalidate's identically-purposed closureDigestDiscoveryOrderOverride/
// closureDigestEmptyCalleesOverride on the producer side): unexported,
// false by production default, restored via defer in every test that
// engages them.
var (
	closureDigestDiscoveryOrderForTest bool
	closureDigestEmptyCalleesForTest   bool
)

// closureDigestDomainSeparator duplicates
// originvalidate.ClosureDigestDomainSeparator's exact literal value
// (D-07-22): this package must never import originvalidate (enforced by
// TestValidatorImportsStayIndependent), so an independent literal, kept
// byte-identical by convention, is the only way this peer's chained
// digest can be compared byte-for-byte against the producer's -- the same
// "shared verbatim as an inert string constant, never a shared
// derivation" posture core.CallGraphCycle already holds across these two
// packages' cycle refusals.
const closureDigestDomainSeparator = "lang.closure_digest/1\x00"

// peerCalleeDigestPair mirrors originvalidate's calleeDigestPair
// field-for-field (same JSON tags, same declaration order) so
// json.Marshal produces byte-identical output given byte-identical
// values -- this package's own independent implementation of the SAME
// preimage shape (D-07-37), never a shared type.
type peerCalleeDigestPair struct {
	ID            string `json:"id"`
	ClosureDigest string `json:"closure_digest"`
}

func peerSortCalleeDigestPairs(pairs []peerCalleeDigestPair) []peerCalleeDigestPair {
	sorted := append([]peerCalleeDigestPair(nil), pairs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	return sorted
}

// peerClosureDigestPreimageBytes independently re-derives
// originvalidate.closureDigestPreimageBytes' exact preimage shape
// (domain separator, then signature with ClosureDigest zeroed, then
// callee pairs sorted by ID) by this package's own route.
func peerClosureDigestPreimageBytes(signature core.FunctionSignature, callees []peerCalleeDigestPair) ([]byte, error) {
	signature.ClosureDigest = ""
	payload := struct {
		Signature core.FunctionSignature `json:"signature"`
		Callees   []peerCalleeDigestPair `json:"callees"`
	}{Signature: signature, Callees: peerSortCalleeDigestPairs(callees)}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	preimage := make([]byte, 0, len(closureDigestDomainSeparator)+len(body))
	preimage = append(preimage, []byte(closureDigestDomainSeparator)...)
	preimage = append(preimage, body...)
	return preimage, nil
}

func peerComputeClosureDigest(signature core.FunctionSignature, callees []peerCalleeDigestPair) (string, error) {
	preimage, err := peerClosureDigestPreimageBytes(signature, callees)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(preimage)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// matchBranch validates a match function whose arms carry linear bodies. It
// validates the match-shaped facts (arm/edge identity, pattern coverage,
// exhaustiveness) independently of the checker exactly as match() does, then
// falls through to linear() for the flattened operations, places, types, and
// the new Block/Edge facts the arms lowered into.
// matchBranchStructural is matchBranch's structural half (D-07-19's
// two-pass split): it validates every match-shaped structural fact exactly
// as matchBranch always did, then falls through to linearStructural for
// the flattened operations/places/types/Block/Edge facts -- but returns
// BEFORE replay, so run()'s whole-program cycle peer can run once, between
// every function's own structural validation and any function's replay.
func (v *validator) matchBranchStructural(function *core.Function, dataNames map[string]core.DataType) (map[string]core.TypeFact, map[string]core.Place, bool) {
	match := function.Match
	if !v.check(match.ID != "" && match.PointID != "" && function.EntryPointID != "" && function.ReturnPointID != "", "core.invalid_id", function.ID) {
		return nil, nil, false
	}
	dataType, knownType := dataNames[function.Parameter.Type]
	if !v.check(knownType && function.ReturnType == function.Parameter.Type, "core.unknown_type", function.Parameter.Type) ||
		!v.check(match.Scrutinee == function.Parameter.Name, "core.unknown_place", match.Scrutinee) {
		return nil, nil, false
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
			return nil, nil, false
		}
		if !v.unique(patterns, arm.Pattern, "core.duplicate_pattern") {
			return nil, nil, false
		}
		if _, known := alternatives[arm.Pattern]; !v.check(known, "core.unknown_alternative", arm.ID) {
			return nil, nil, false
		}
		// This phase's scope decision (checkBranch): a match that carries
		// any arm body requires every arm to carry one. Value and BlockID
		// are therefore mutually exclusive per arm, and BlockID must be
		// present here — a bare Value in a branch-shaped function is
		// rejected rather than silently tolerated.
		if !v.check(arm.Value == "" && arm.BlockID != "", "core.invalid_body", arm.ID) {
			return nil, nil, false
		}
	}
	if !v.check(len(patterns) == len(alternatives), "core.final_claim_mismatch", match.ID) {
		return nil, nil, false
	}
	return v.linearStructural(function)
}

func (v *validator) match(function *core.Function, dataNames map[string]core.DataType) bool {
	// D-07-25: a pure lang.core/0 match function (no Linear body at all)
	// never reaches replayStraightLine/replayBlocks, but
	// originvalidate.BuildInterface still emits a FunctionSignature for it
	// (Callable trivially true: RecomputeOrigin reports not-ok when
	// function.Linear is nil, so PublishProblemsFor reports no problem).
	// Record the peer signature here too, with nil types/places, so the
	// whole-corpus zero-divergence claim (D-07-25) covers /0-only programs
	// as well, not only linear-bodied ones. This is not one of D-07-21's two
	// replay sites, so it never sets peerRanStraightLine/peerRanBlocks.
	if v.peerSignatures == nil {
		v.peerSignatures = make(map[string]core.FunctionSignature)
	}
	v.peerSignatures[function.ID] = derivePeerSignature(function, nil, nil)
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

// linearStructural is linear's structural half (D-07-19's two-pass
// split): types, places, operation identity/order/reference consistency,
// and (when Blocks/Edges are declared) their referential closure -- every
// check linear always performed BEFORE its own replay call. It returns
// BEFORE replay and loan-endpoint recomputation, which linear composes
// back on afterward, so run() can insert its whole-program cycle peer
// strictly between "every function structurally valid" and "any function
// replayed".
func (v *validator) linearStructural(function *core.Function) (map[string]core.TypeFact, map[string]core.Place, bool) {
	linear := function.Linear
	if !v.check(linear.ID == function.ID+":linear", "core.linear_id", linear.ID) ||
		!v.check(function.EntryPointID == function.ID+":point:entry" && function.ReturnPointID == function.ID+":point:return", "core.point_id", function.ID) {
		return nil, nil, false
	}

	types := make(map[string]core.TypeFact, len(linear.Types))
	for index, fact := range linear.Types {
		if !v.uniqueType(types, fact) {
			return nil, nil, false
		}
		if !v.check(fact.ID == fmt.Sprintf("%s:type:%d", function.ID, index), "core.type_order", fact.ID) {
			return nil, nil, false
		}
		abilities, witnesses, ok := v.derive(fact.Shape, 0)
		if !ok {
			return nil, nil, false
		}
		if !v.check(reflect.DeepEqual(fact.Abilities, abilities) && reflect.DeepEqual(fact.NegativeWitnesses, witnesses), "core.ability_mismatch", fact.ID) {
			return nil, nil, false
		}
	}

	places := make(map[string]core.Place, len(linear.Places))
	for index, place := range linear.Places {
		if !v.uniquePlace(places, place) {
			return nil, nil, false
		}
		if !v.check(place.ID == fmt.Sprintf("%s:place:%d", function.ID, index), "core.place_order", place.ID) {
			return nil, nil, false
		}
		if _, ok := types[place.TypeID]; !v.check(ok, "core.unknown_type", place.TypeID) {
			return nil, nil, false
		}
	}
	parameter, ok := places[function.Parameter.ID]
	if !v.check(ok, "core.unknown_place", function.Parameter.ID) ||
		!v.check(parameter.Name == function.Parameter.Name, "core.parameter_mismatch", function.Parameter.ID) {
		return nil, nil, false
	}
	parameterType := types[parameter.TypeID]
	if !v.check(parameterType.Shape.Constructor == function.Parameter.Type, "core.parameter_mismatch", function.Parameter.ID) {
		return nil, nil, false
	}

	operationIDs := make(map[string]struct{}, len(linear.Operations))
	pointIDs := make(map[string]struct{}, len(linear.Operations))
	loanIDs := make(map[string]struct{})
	for index, operation := range linear.Operations {
		if !v.unique(operationIDs, operation.ID, "core.duplicate_operation_id") || !v.unique(pointIDs, operation.PointID, "core.duplicate_point_id") {
			return nil, nil, false
		}
		if !v.check(operation.ID == fmt.Sprintf("%s:op:%d", function.ID, index) && operation.PointID == fmt.Sprintf("%s:point:linear:%d", function.ID, index), "core.operation_order", operation.ID) {
			return nil, nil, false
		}
		if _, ok := places[operation.SourceID]; !v.check(ok, "core.unknown_place", operation.SourceID) {
			return nil, nil, false
		}
		if _, ok := types[operation.TypeID]; !v.check(ok, "core.unknown_type", operation.TypeID) {
			return nil, nil, false
		}
		// OpFail and OpDefect are terminators alongside OpReturn (D-04-04/
		// D-04-15): none of the three ever carries a TargetID, since none
		// produces an ordinary place -- OpReturn ends the function, OpFail
		// ends the err block, OpDefect ends an arm block by aborting.
		if operation.Kind != core.OpReturn && operation.Kind != core.OpFail && operation.Kind != core.OpRelease && operation.Kind != core.OpDefect {
			if _, ok := places[operation.TargetID]; !v.check(ok, "core.unknown_place", operation.TargetID) {
				return nil, nil, false
			}
		}
		if operation.Kind == core.OpForeignCall {
			if _, ok := places[operation.ErrTargetID]; !v.check(ok, "core.unknown_place", operation.ErrTargetID) {
				return nil, nil, false
			}
			if !v.check(operation.OkEdgeID != "" && operation.ErrEdgeID != "", "core.foreign_call_edges_missing", operation.ID) {
				return nil, nil, false
			}
			if !v.check(function.ForeignContract != nil && function.ForeignContract.Symbol != "", "core.foreign_contract_missing", operation.ID) {
				return nil, nil, false
			}
			// Phase 4 plan 12 (D-04-12/FFI-01, 04-VERIFICATION.md gap 2):
			// check.go refuses a malformed symbol at parse-resolution time by
			// inspecting the AST, an AST-level fact this validator never
			// sees; this validator re-derives an equivalent refusal purely
			// from the one flat string field core.ForeignContract itself
			// carries. The refusal exists because cgen splices this exact
			// string into generated C at three sites plus a header comment,
			// so an unaudited value is arbitrary C source injection at the
			// boundary the phase goal calls audited. Pass operation.ID, not
			// the Symbol itself, as the detail: Problem.Detail is serialized
			// into diagnostic JSON, and echoing an attacker-controlled
			// string containing newlines or quotes into an output channel is
			// the same class of defect this check exists to close.
			if !v.check(validCIdentifier(function.ForeignContract.Symbol), "foreign.symbol_not_identifier", operation.ID) {
				return nil, nil, false
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
				return nil, nil, false
			}
			// 04-13 (04-VERIFICATION.md gap 2b, FFI-01): check.go now refuses
			// a hostile policy value at source-admission time by inspecting
			// ast.ForeignPolicy.Value's span-bearing AST node (Task 1), an
			// AST-level fact this validator never sees. This validator
			// re-derives an equivalent refusal purely from the three flat
			// string fields core.ForeignContract itself carries, so a
			// corrupted core artifact that skipped check.go's gate is still
			// caught. The check immediately above is a PRESENCE claim
			// (non-empty); this is the SHAPE claim (C-identifier), and it
			// exists because cgen.EmitForeignHeader splices these exact three
			// strings raw into C comments at cgen.go:1333-1335, in a unit
			// session.go hands to native.Runner.CompileConformanceUnit for
			// real compilation. Placed AFTER foreign.unwind_policy_undeclared
			// so an OMITTED unwind/nonlocal_exit policy keeps reporting that
			// existing code, and AFTER foreign.symbol_not_identifier so
			// 04-12's audit still fires first for a hostile Symbol. Passes
			// operation.ID, never a field value, for the same
			// diagnostic-JSON-echo reason as every sibling check in this
			// block.
			if !v.check(validCIdentifier(function.ForeignContract.Allocator) && validCIdentifier(function.ForeignContract.Unwind) && validCIdentifier(function.ForeignContract.NonlocalExit), "foreign.policy_value_not_identifier", operation.ID) {
				return nil, nil, false
			}
			// D-04-02, independently derived: check.go refuses a callee that
			// resolves to a Lang function name at parse-resolution time (an
			// AST-level, pre-core fact this validator never sees). This
			// validator re-derives the same refusal by a materially
			// different mechanism -- a straight name collision scan against
			// every OTHER function this core.Program itself declares --
			// reading only the core artifact, never check's own name table.
			if !v.check(!isDeclaredFunctionName(v.program.Functions, function.ID, function.ForeignContract.Symbol), "core.call_target_not_foreign", operation.ID) {
				return nil, nil, false
			}
			// Phase 4 plan 03 (D-04-12/FFI-01): every foreign obligation
			// category must be present -- no default value that would let an
			// omission pass as a declaration -- and the declared Layout must
			// be internally consistent. This is independently derived from
			// check's own admission gate: it reads only the four flat string
			// fields and the Layout struct core.ForeignContract itself
			// carries, never check's AST-level foreignSymbolInfo.
			if !v.check(function.ForeignContract.InitializedState != "" && function.ForeignContract.Capture != "" && function.ForeignContract.Retention != "" && function.ForeignContract.Aliasing != "", "foreign.obligation_undeclared", operation.ID) {
				return nil, nil, false
			}
			// 04-13 Task 3 (04-VERIFICATION.md gap 2b, FFI-01): audit every
			// REMAINING core.ForeignContract string field cgen splices into
			// generated C -- Fails, InitializedState, Capture, Retention,
			// Aliasing (cgen.go:1336-1348's comment block) and, when a
			// Layout is declared, Layout.ForeignTypeName and each
			// Layout.Fields[].Name / .CType (spliced as REAL C tokens into
			// EmitForeignConformance's _Static_assert operands, not merely
			// into a comment). Placed immediately after
			// foreign.obligation_undeclared so an OMITTED obligation keeps
			// reporting that existing code; this is the SHAPE/safety claim
			// for the fields the presence check above does not inspect.
			// Passes operation.ID, never a field value, for the same
			// diagnostic-JSON-echo reason as every sibling check in this
			// block.
			if !v.check(foreignContractFieldsCSafe(function.ForeignContract), "foreign.contract_field_not_c_safe", operation.ID) {
				return nil, nil, false
			}
			if !v.foreignLayoutConsistent(function.ForeignContract.Layout, operation.ID) {
				return nil, nil, false
			}
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			if !v.unique(loanIDs, operation.LoanID, "core.unknown_loan") {
				return nil, nil, false
			}
		} else if !v.check(operation.LoanID == "", "core.unknown_loan", operation.LoanID) {
			return nil, nil, false
		}
	}
	if len(linear.Blocks) > 0 || len(linear.Edges) > 0 {
		if !v.blocksAndEdges(function, operationIDs) {
			return nil, nil, false
		}
	}
	return types, places, true
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
	// core.terminal_block_unreachable (04-VERIFICATION.md gap 1, Task 1
	// option A): a structural peer of the OpFail incoming-edge check below,
	// independent of checkReleaseOrder itself. Every OpReturn- or
	// OpFail-terminated block that is NOT the function's own entry block
	// must have at least one incoming edge -- a terminal block with zero
	// incoming edges is unreachable by construction, and unreachable code
	// is exactly the shape whose release order checkReleaseOrder cannot
	// rederive (it has no incoming edge to walk backward from). This does
	// NOT bound the count from above: an upper bound of "exactly one" was
	// tried and refused a legitimate program (`discard <call> because
	// "..."` merges its ok/err edges into one successor block, which is
	// terminal and has TWO incoming edges) -- see 04-REVIEW-FIX.md's CR-01
	// entry. The function's entry block is identified by PointID, matching
	// EntryPointID exactly, the same convention check.go's own producers
	// (checkBranch, checkForeignTracer, checkResourceLifecycle) already use
	// for the entry block -- not by position in the Blocks slice.
	for _, block := range linear.Blocks {
		if len(block.OperationIDs) == 0 {
			continue
		}
		last := operationsByID[block.OperationIDs[len(block.OperationIDs)-1]]
		if last.Kind != core.OpFail && last.Kind != core.OpReturn {
			continue
		}
		v.checks++ // one inspection per terminal block, independent of checkReleaseOrder's own per-terminal-block reduction
		if block.PointID != function.EntryPointID {
			if !v.check(len(incoming[block.ID]) > 0, "core.terminal_block_unreachable", block.ID) {
				return false
			}
		}
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
					ID:     fmt.Sprintf("%s:point:%s:%d:%s", function.ID, block.ID, lastOrdinalAt[blockIdx], loanID),
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
				ID:     fmt.Sprintf("%s:point:%s:%d:%s", function.ID, block.ID, birth.ordinal, loanID),
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
	// D-07-21: the summary peer runs at THIS replay site, using the
	// types/places this straight-line replay already computed -- one of the
	// two required wiring points (see replayBlocks' identical call).
	v.recordSummaryPeer(function, types, places, false)
	operations := function.Linear.Operations
	operationsByID := make(map[string]core.LinearOperation, len(operations))
	for _, operation := range operations {
		operationsByID[operation.ID] = operation
	}
	// D-07-29: CalleeID is fail-closed and kind-exclusive, independently
	// re-derived here rather than trusting check's own bookkeeping.
	// declaredFunctionIDs backs the third refusal -- a CalleeID naming no
	// declared function (D-07-45) -- with the whole program's own function
	// ID set, never a name-keyed lookup.
	declaredFunctionIDs := make(map[string]bool, len(v.program.Functions))
	for _, fn := range v.program.Functions {
		declaredFunctionIDs[fn.ID] = true
	}
	// functionByID backs the NEW (07-05 Task 3) Callable refusal below: the
	// callee's own core.Function value, from this program's own function
	// set, is peerCallable's only input -- never check's signature table.
	functionByID := make(map[string]core.Function, len(v.program.Functions))
	for _, fn := range v.program.Functions {
		functionByID[fn.ID] = fn
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
		// D-07-29: CalleeID is populated only on an OpCall -- every other
		// kind must leave it empty, exactly like Allocator/ReleasesOperationID
		// before it. Checked once here, for every operation regardless of
		// kind, rather than duplicated per non-OpCall case arm below.
		if !v.check(operation.Kind == core.OpCall || operation.CalleeID == "", "core.callee_id_kind_exclusive", operation.ID) {
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
		case core.OpCall:
			// D-07-29/D-07-45: a call's CalleeID is fail-closed (never empty
			// on an OpCall -- the kind-exclusive check above already refused
			// the reverse) and must name a function this program actually
			// declares. An unresolvable CalleeID is never silently dropped;
			// it is refused here with its own typed identity, distinct from
			// the (07-06) cycle code.
			if !v.check(operation.CalleeID != "", "core.callee_id_missing", operation.ID) {
				return false
			}
			if !v.check(disableCalleeResolutionCheckForTest || declaredFunctionIDs[operation.CalleeID], core.CallCalleeUnresolved, operation.CalleeID) {
				return false
			}
			// 07-05 Task 3/D-07-34: corevalidate's OWN, independently
			// re-derived half of SEM-06 -- peerCallable (D-07-33's narrowed
			// re-derivation, 07-02), consulted here on the CALLEE'S OWN
			// core.Function value from this program's own function set,
			// never on check's signature table and never by asking check.
			// See verifyCallableRefusal (check.go) for the other, separate
			// derivation of this exact fact.
			if callee, ok := functionByID[operation.CalleeID]; ok {
				if !v.check(peerCallable(&callee), core.CalleeNotCallable, operation.CalleeID) {
					return false
				}
				if !v.checkCallTypeContract(callee, source, operation, places, types) {
					return false
				}
			}
			if !v.consumeCallArgument(operation, types, initialized) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
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
	// D-07-21: the summary peer runs at THIS replay site too, using the
	// types/places this branch-shaped replay computed -- see
	// replayStraightLine's identical call. A peer wired into only one site
	// reproduces the literal D-02-03/D-03-01 repeat. disableSummaryPeerAtReplayBlocksForTest
	// is fault 2's D-07-42 seam (07-02 Task 3): production always calls
	// recordSummaryPeer here; the test temporarily skips it to prove a
	// branch-shaped-only borrow-derived return becomes undetectable when
	// only this site's wiring is missing.
	if !disableSummaryPeerAtReplayBlocksForTest {
		v.recordSummaryPeer(function, types, places, true)
	}
	linear := function.Linear
	operations := linear.Operations
	lastOperationOfBlock := make(map[string]string, len(linear.Blocks))
	blockOfOperation := make(map[string]string, len(operations))
	operationsByID := make(map[string]core.LinearOperation, len(operations))
	for _, operation := range operations {
		operationsByID[operation.ID] = operation
	}
	// See replayStraightLine's identical declaration (D-07-29/D-07-45).
	declaredFunctionIDs := make(map[string]bool, len(v.program.Functions))
	for _, fn := range v.program.Functions {
		declaredFunctionIDs[fn.ID] = true
	}
	// See replayStraightLine's identical declaration (07-05 Task 3/D-07-34).
	functionByID := make(map[string]core.Function, len(v.program.Functions))
	for _, fn := range v.program.Functions {
		functionByID[fn.ID] = fn
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
		// See replayStraightLine's identical check (D-07-29): CalleeID is
		// kind-exclusive, checked once per operation regardless of kind.
		if !v.check(operation.Kind == core.OpCall || operation.CalleeID == "", "core.callee_id_kind_exclusive", operation.ID) {
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
		case core.OpCall:
			// See replayStraightLine's identical case (D-07-29/D-07-45).
			if !v.check(operation.CalleeID != "", "core.callee_id_missing", operation.ID) {
				return false
			}
			if !v.check(disableCalleeResolutionCheckForTest || declaredFunctionIDs[operation.CalleeID], core.CallCalleeUnresolved, operation.CalleeID) {
				return false
			}
			// See replayStraightLine's identical case (07-05 Task 3/D-07-34).
			if callee, ok := functionByID[operation.CalleeID]; ok {
				if !v.check(peerCallable(&callee), core.CalleeNotCallable, operation.CalleeID) {
					return false
				}
				// See replayStraightLine's identical case (07-09).
				if !v.checkCallTypeContract(callee, source, operation, places, types) {
					return false
				}
			}
			// See replayStraightLine's identical case (07-11/PVG-01/CR-01).
			if !v.consumeCallArgument(operation, types, initialized) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
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

// recordSummaryPeer is 07-02's D-07-21 wiring point: called from BOTH
// replayStraightLine (viaBlocks=false) and replayBlocks (viaBlocks=true),
// storing derivePeerSignature's independently-derived summary for function
// and marking which replay shape actually ran it. Called unconditionally at
// each site's entry (not gated on the replay's own pass/fail), because the
// signature summary a producer would publish is a fact about the function's
// declared shape, independent of whether THIS validator run additionally
// finds a different problem elsewhere in the same program.
func (v *validator) recordSummaryPeer(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place, viaBlocks bool) {
	if v.peerSignatures == nil {
		v.peerSignatures = make(map[string]core.FunctionSignature)
	}
	v.peerSignatures[function.ID] = derivePeerSignature(function, types, places)
	if viaBlocks {
		v.peerRanBlocks = true
	} else {
		v.peerRanStraightLine = true
	}
}

// SummaryPeerControls names every control 07-02 introduces (its PLAN.md's
// <artifacts_this_phase_produces>), in a form 07-07's phase-wide
// completeness matrix can compare against by exact set equality. This is
// plain read-only data, never a fault-injection seam -- D-07-42's
// unexported-mutable-var constraint does not apply to it.
var SummaryPeerControls = []string{
	"control:summary.peer_both_replay_sites",
	"control:summary.producer_peer_zero_divergence",
	"control:summary.callable_publication_safety",
}

// derivePeerSignature is corevalidate's independent (D-07-20/D-07-22)
// structural re-derivation of one function's lang.interface/1 signature
// summary. It names what originvalidate.BuildInterface does (strips every
// function body and packages the producer's own already-proven facts into a
// summary) and the materially different mechanism this one uses instead: it
// reads types via a direct map lookup into the validator's own already-
// replayed type-fact map (built once per Validate() call from
// function.Linear.Types, keyed by fact ID) rather than BuildInterface's
// fresh linear scan over function.Linear.Types for every call, and it
// derives parameter-drop and origin-omission facts by forward set-
// propagation over function.Linear.Operations rather than
// originvalidate's backward per-return walk (see peerParameterEscapesOwned/
// peerReturnDerivesFromBorrow). It never imports originvalidate and never
// calls BuildInterface or PublishProblemsFor (enforced by
// TestValidatorImportsStayIndependent, corevalidate_endpoint_internal_test.go) --
// two callers of one implementation cannot diverge by construction, so
// sharing the implementation here would make this "peer" agreement theater,
// not evidence (D-07-20).
//
// D-07-33: Callable is narrowed to independently re-deriving ONLY the
// core.origin_omitted refusal class. For core.origin_understated,
// core.origin_access_mismatch, and foreign-origin-omitted, this peer does
// not compute an independent answer at all -- it can only ever falsely
// agree with whatever the producer declares for those classes this phase.
// See PHASE-07-DEBT.md's D-07-33 entry; Phase 09 closes this.
func derivePeerSignature(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) core.FunctionSignature {
	// places is accepted (not merely function+types) so this signature
	// matches recordSummaryPeer's call sites at both replay sites, which
	// already have it in scope -- no field derived by this peer needs Place
	// facts yet; a future field can gain access without a signature change.
	_ = places

	var abilities []core.Ability
	if fact, ok := types[function.ID+":type:0"]; ok {
		abilities = fact.Abilities
	} else {
		abilities = []core.Ability{}
	}
	hasDropAbility := false
	for _, ability := range abilities {
		if ability == core.AbilityDrop {
			hasDropAbility = true
			break
		}
	}

	parameterContract := core.ParameterContract{
		ID: function.Parameter.ID, Name: function.Parameter.Name, Type: function.Parameter.Type,
		// D-07-01: today's grammar has exactly one parameter form.
		Mode:  "owned",
		Drops: hasDropAbility && !peerParameterEscapesOwned(function),
	}

	returnContract := core.ReturnContract{Type: function.ReturnType}
	switch {
	case function.PublicOrigin == nil:
		returnContract.Mode = "owned"
		returnContract.Paths = []string{}
	case function.PublicOrigin.Access == "shared":
		returnContract.Mode = "shared"
		returnContract.Paths = function.PublicOrigin.Paths
	default:
		returnContract.Mode = function.PublicOrigin.Access
		returnContract.Paths = function.PublicOrigin.Paths
	}
	returnContract.Fresh = returnContract.Mode == "owned" && hasDropAbility

	foreignReach := core.ForeignReach{}
	fails := ""
	if function.ForeignContract != nil {
		foreignReach = core.ForeignReach{
			Allocator: function.ForeignContract.Allocator, Unwind: function.ForeignContract.Unwind,
			NonlocalExit: function.ForeignContract.NonlocalExit,
		}
		fails = function.ForeignContract.Fails
	}

	return core.FunctionSignature{
		ID: function.ID, Name: function.Name,
		Parameters: []core.ParameterContract{parameterContract},
		Return:     returnContract,
		Abilities:  abilities,
		Callable:   peerCallable(function),
		Fails:      fails,
		Foreign:    foreignReach,
	}
}

// peerParameterEscapesOwned is peerParameterContract's independent
// re-derivation of originvalidate.parameterEscapesOwned, by a materially
// different mechanism (D-07-22): FORWARD set-propagation from the parameter
// over function.Linear.Operations (an OpMove/OpCopy hop propagates
// membership to its target) rather than a backward per-return walk with
// cycle detection. It reports whether the parameter is directly moved or
// copied, with no intervening borrow/foreign hop, all the way out through
// some OpReturn.
func peerParameterEscapesOwned(function *core.Function) bool {
	if function.Linear == nil {
		return false
	}
	tracesToParameter := map[string]bool{function.Parameter.ID: true}
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpMove && operation.Kind != core.OpCopy {
			continue
		}
		if tracesToParameter[operation.SourceID] {
			tracesToParameter[operation.TargetID] = true
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpReturn && tracesToParameter[operation.SourceID] {
			return true
		}
	}
	return false
}

// peerReturnDerivesFromBorrow is peerCallable's independent re-derivation of
// the core.origin_omitted refusal class ONLY (D-07-33's narrowed scope), by
// a materially different mechanism (D-07-22) from
// originvalidate.RecomputeOriginPerReturn's backward per-return walk:
// FORWARD set-propagation from the parameter over
// function.Linear.Operations, marking a place borrow-derived when it is the
// target of an OpBorrowShared/OpBorrowExclusive hop from an already-tracked
// place, or of a plain OpMove/OpCopy hop from one. It reports presence only
// (does this function return SOME borrow-derived place), never an access
// mode -- narrowed Callable only needs presence, not the mode
// RecomputeOrigin's first-hop-wins rule additionally derives.
//
// D-07-33: this deliberately does NOT walk through OpForeignCall the way
// originvalidate.checkForeignOriginOmitted/RecomputeOriginPerReturn do for a
// declared borrow/retain foreign contract. foreign-origin-omitted is one of
// the three classes this peer does not independently re-derive this phase
// (see peerCallable's doc comment and TestPeerDoesNotRederiveNarrowedClasses).
func peerReturnDerivesFromBorrow(function *core.Function) bool {
	if function.Linear == nil {
		return false
	}
	// paramTrace tracks places that are the parameter itself, or reach it
	// through a pure Move/Copy chain -- check.go's arm lowering copies the
	// match scrutinee (the parameter) into a fresh place before each arm
	// borrows it, so the borrow hop's SourceID is usually a copy of the
	// parameter, never the parameter's own place ID directly.
	paramTrace := map[string]bool{function.Parameter.ID: true}
	derived := make(map[string]bool)
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared, core.OpBorrowExclusive:
			if paramTrace[operation.SourceID] || derived[operation.SourceID] {
				derived[operation.TargetID] = true
			}
		case core.OpMove, core.OpCopy:
			if paramTrace[operation.SourceID] {
				paramTrace[operation.TargetID] = true
			}
			if derived[operation.SourceID] {
				derived[operation.TargetID] = true
			}
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpReturn && derived[operation.SourceID] {
			return true
		}
	}
	return false
}

// peerCallable is D-07-33's narrowed independent re-derivation of Callable.
// It only independently answers the core.origin_omitted class: when no
// public origin is declared, it reports Callable false iff the body itself
// returns a borrow-derived place. When a public origin IS declared, this
// peer does not check whether the declaration is understated, access-
// mismatched, or a foreign-origin omission -- it reports Callable true
// unconditionally, which is only ever a FALSE agreement with the producer
// for those three classes (D-07-33, declared in PHASE-07-DEBT.md, closed in
// Phase 09).
// disableSummaryPeerAtReplayBlocksForTest is fault 2's D-07-42 fault-
// injection seam (07-02 Task 3): production always wires the summary peer
// into replayBlocks (D-07-21); the test temporarily disables ONLY this
// site, proving a branch-shaped function's borrow-derived-on-exactly-one-arm
// return is undetectable through replayStraightLine alone -- the literal
// D-02-03/D-03-01 repeat this wiring exists to catch. false (the
// always-wired production default) means "call recordSummaryPeer here".
var disableSummaryPeerAtReplayBlocksForTest bool

// disableCalleeResolutionCheckForTest is Task 2's D-07-41/D-07-42 seam
// (QLT-08): when true, corevalidate's own independent re-derivation that an
// OpCall's CalleeID names a declared function is skipped, at BOTH replay
// sites. It never suppresses the kind-exclusivity check (empty CalleeID on
// an OpCall, or a non-empty one on any other kind) -- only the
// resolves-to-a-declared-function predicate, mirroring check's own
// verifyCallInvariantsSeam so the two independent sites can each be shown to
// fail under the same seeded mutation, in their own package's own test
// (D-07-42's cross-package split, following 07-02's precedent). false (the
// production default) means "check declaredFunctionIDs for real".
var disableCalleeResolutionCheckForTest bool

// forcePeerCallableAlwaysTrue is D-07-42's fault-injection seam for
// TestStage0SummaryMutationMatrix's faults 3 and 5 (07-02 Task 3): mirrors
// originvalidate's forceCallableAlwaysTrue on this package's own peer
// derivation. false (the always-real production default) means "use
// peerCallable's real narrowed derivation".
var forcePeerCallableAlwaysTrue bool

// disableCallArgumentTypePeerForTest is 07-09 Task 2's D-07-42 fault-
// injection seam: when true, the OpCall replay arm's independent argument-
// type refusal (core.CallArgumentTypeMismatch) is skipped, admitting a
// mismatched argument type anyway. Unexported, false in production, set
// only via SetDisableCallArgumentTypePeerForTest (export_test.go) by a
// same-package test that defers the restore immediately. Its purpose is to
// prove the peer refusal is load-bearing ON ITS OWN: Task 2's
// TestCallTypePeersIndependentOfCheck disables check's own argument-type
// gate through ITS seam and shows corevalidate still refuses; Task 3's
// mutation matrix disables THIS seam and shows corevalidate stops
// refusing, proving this predicate -- not merely check's -- is what was
// catching the defect.
var disableCallArgumentTypePeerForTest bool

// disableCallReturnTypePeerForTest is 07-09 Task 2's D-07-42 fault-
// injection seam, the target/return-type counterpart of
// disableCallArgumentTypePeerForTest above: when true, the OpCall replay
// arm's independent return-type refusal (core.CallReturnTypeMismatch) is
// skipped. Unexported, false in production, set only via
// SetDisableCallReturnTypePeerForTest.
var disableCallReturnTypePeerForTest bool

// disableCallArgumentConsumePeerForTest is 07-11's D-07-41/D-07-42
// fault-injection seam (PVG-01/CR-01, QLT-08) for
// control:call.argument_consumed_when_noncopyable: when true,
// consumeCallArgument never move-marks a call argument, regardless of its
// copy ability -- reproducing the pre-07-11 hole from the peer's own side,
// independently of check's identical seam (callArgumentConsumeSeam,
// check.go). Unexported, false in production, set only via
// SetDisableCallArgumentConsumePeerForTest (export_test.go) by a
// same-package test that defers the restore immediately.
var disableCallArgumentConsumePeerForTest bool

// forceCallArgumentConsumePeerForTest is 07-11's D-07-41/D-07-42
// fault-injection seam for control:call.copyable_argument_not_consumed:
// when true, consumeCallArgument move-marks a call argument REGARDLESS of
// its copy ability -- the over-refusal fault for the non-refusing
// direction, so a copyable argument's second use is wrongly refused with
// core.place_uninitialized. Unexported, false in production, set only via
// SetForceCallArgumentConsumePeerForTest.
var forceCallArgumentConsumePeerForTest bool

// consumeCallArgument is 07-11's independent peer half of the call-site
// ownership consume rule (PVG-01/CR-01): a call transfers its argument, and
// corevalidate decides whether that transfer costs the caller its binding
// using ONLY its own derivation -- v.derive over the emitted core
// artifact's own types[operation.TypeID].Shape (its own in-package
// deriveAbility/sealedLeaves) -- never check's recorded Abilities list and
// never check's ability package. operation.TypeID is already asserted equal
// to the argument place's own TypeID by the generic
// `source.TypeID == operation.TypeID` check every operation passes before
// its kind-specific switch runs (this function's own caller), so deriving
// from operation.TypeID IS deriving from the argument's own type.
//
// A non-copyable argument is consumed by mirroring the OpMove arm verbatim
// (initialized[operation.SourceID] = false); the SECOND use of that same
// source -- including a second call -- is then refused by the pre-existing
// generic initialized[...]/finalOrTransitionCode gate at the top of the
// per-operation loop, which yields core.place_uninitialized. A copyable
// argument is left untouched, exactly as the OpCopy arm leaves its own
// source. An unresolvable type fact or an unknown type constructor is the
// REFUSING case: v.derive itself appends a core.unknown_type_constructor
// problem and reports failure, which this function propagates by returning
// false -- fail-closed, never a silent admit.
func (v *validator) consumeCallArgument(operation core.LinearOperation, types map[string]core.TypeFact, initialized map[string]bool) bool {
	if disableCallArgumentConsumePeerForTest {
		return true
	}
	granted, _, known := v.derive(types[operation.TypeID].Shape, 0)
	if !known {
		return false
	}
	if forceCallArgumentConsumePeerForTest || !abilityGranted(granted, core.AbilityCopy) {
		initialized[operation.SourceID] = false
	}
	return true
}

// abilityGranted reports whether wanted appears in a v.derive-returned
// granted-abilities slice -- corevalidate's own independent structural
// re-derivation, never check's TypeFact.Abilities list or hasAbility.
func abilityGranted(granted []core.Ability, wanted core.Ability) bool {
	for _, candidate := range granted {
		if candidate == wanted {
			return true
		}
	}
	return false
}

// checkCallTypeContract is 07-09's independent peer half of the call
// argument/return type contract (D-07-09/SEM-05/T-07-09-03) -- the pair of
// facts 07-VERIFICATION.md found missing entirely from corevalidate's
// OpCall replay. It is consulted from BOTH OpCall replay sites
// (replayStraightLine and replayBlocks) on the callee's OWN core.Function
// value (from THIS program's own functionByID, never check's
// callSignatureTable and never a value obtained by asking check anything).
// Sharing this helper across corevalidate's own two replay sites does not
// weaken the independence property the plan requires: the property is
// independence FROM check, never internal deduplication within this one
// package, and check.go's resolveCallBinding has no reachable path into
// this or any other corevalidate symbol.
//
// The two predicates are structurally independent of check's own gate:
//  1. It lives entirely in package corevalidate, importing nothing from
//     the checker package.
//  2. It reads the callee's own declared Parameter.Type/ReturnType off
//     core.Function -- a plain field on this program's own data, never a
//     published interface signature type and never a value check produced.
//  3. It resolves the ARGUMENT and TARGET places' own types by walking
//     places and this function's own type facts (place.TypeID ->
//     types[TypeID].Shape.Constructor) -- a materially different
//     mechanism from check's name-keyed AST-derived callee-contract table
//     lookup (resolveCallBinding, check.go).
//  4. It shares no function, type, or constant with check that expresses
//     either comparison. The two sides' code strings
//     (core.CallArgumentTypeMismatch/core.CallReturnTypeMismatch versus
//     check.call_argument_type_mismatch/check.call_return_type_unrepresentable)
//     are already distinct by the checkpoint's ratification, which makes
//     accidental coupling visible in any diagnostics list.
//
// Both predicates fail closed: an unresolvable place, an unresolvable
// type fact, or an empty constructor or contract string refuses -- absence
// is never the permitting case. 07-VERIFICATION.md found that both
// derivations shared one blind spot (the caller's own argument TypeID,
// asked nothing of the callee), so the two-peer safety net gave zero
// protection against this defect class. A shared helper WITH check here
// would rebuild exactly that; this helper shares nothing with it.
func (v *validator) checkCallTypeContract(callee core.Function, source core.Place, operation core.LinearOperation, places map[string]core.Place, types map[string]core.TypeFact) bool {
	argumentConstructor := ""
	if fact, ok := types[source.TypeID]; ok {
		argumentConstructor = fact.Shape.Constructor
	}
	argumentTypeMatches := argumentConstructor != "" && callee.Parameter.Type != "" && argumentConstructor == callee.Parameter.Type
	if !v.check(disableCallArgumentTypePeerForTest || argumentTypeMatches, core.CallArgumentTypeMismatch, operation.ID) {
		return false
	}
	targetConstructor := ""
	if targetPlace, ok := places[operation.TargetID]; ok {
		if fact, ok := types[targetPlace.TypeID]; ok {
			targetConstructor = fact.Shape.Constructor
		}
	}
	targetTypeMatches := targetConstructor != "" && callee.ReturnType != "" && targetConstructor == callee.ReturnType
	return v.check(disableCallReturnTypePeerForTest || targetTypeMatches, core.CallReturnTypeMismatch, operation.ID)
}

func peerCallable(function *core.Function) bool {
	if forcePeerCallableAlwaysTrue {
		return true
	}
	if function.PublicOrigin == nil {
		return !peerReturnDerivesFromBorrow(function)
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
	okEdgeInto := make(map[string][]core.Edge, len(linear.Edges))
	for _, edge := range linear.Edges {
		edgesByTo[edge.ToBlockID] = append(edgesByTo[edge.ToBlockID], edge)
		if edge.Pattern == "ok" {
			okEdgeInto[edge.ToBlockID] = append(okEdgeInto[edge.ToBlockID], edge)
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

	// rederive walks BACKWARD over okEdgeInto, a map built directly from
	// linear.Edges with no acyclicity precheck anywhere earlier in Validate
	// (blocksAndEdges, above, checks ID uniqueness and referential closure
	// only). A cyclic ok-edge chain can only arise from a corrupted core
	// artifact -- exactly the adversarial input a source-blind validator
	// must stay defined against, the same rationale loanChainIndex.carriedLoans
	// and blockReach already state for their own walks. The visited set below
	// matches their idiom: a re-visited block is a HARD REFUSAL
	// (core.release_order_cyclic), never a silent truncation of expected --
	// a truncated expected would still be compared against actual and could
	// ACCEPT a corrupted program, which is worse than the hang this guard
	// replaces. The boolean return propagates the refusal to every call site.
	var rederive func(startEdge core.Edge, visited map[string]bool) ([]core.LinearOperation, bool)
	rederive = func(startEdge core.Edge, visited map[string]bool) ([]core.LinearOperation, bool) {
		var expected []core.LinearOperation
		includeThis := startEdge.Pattern == "ok"
		currentBlockID := startEdge.FromBlockID
		for {
			if visited[currentBlockID] {
				v.check(false, "core.release_order_cyclic", currentBlockID)
				return nil, false
			}
			visited[currentBlockID] = true
			v.checks++ // one inspection per block visited while walking backward
			if includeThis {
				if op, ok := callInBlock[currentBlockID]; ok {
					expected = append(expected, op)
				}
			}
			candidates := okEdgeInto[currentBlockID]
			if len(candidates) == 0 {
				break
			}
			if len(candidates) == 1 {
				currentBlockID = candidates[0].FromBlockID
				includeThis = true
				continue
			}
			// Interior merge: okEdgeInto used to be a singular map whose
			// last-writer-wins assignment discarded any but the final
			// declared ok edge into this block, so a genuinely disagreeing
			// history was never rederived or compared. The per-terminal-block
			// loop below already applies the discipline of checking EVERY
			// incoming edge rather than skipping or choosing one; this branch
			// extends that same discipline one hop earlier, into the backward
			// walk itself. check.go's honest lowering never produces two ok
			// edges into one block, so this is defense against a
			// hand-corrupted artifact -- exactly corevalidate's stated
			// purpose. Each candidate walks with its own COPY of visited: a
			// shared map would make a legitimate diamond that reconverges on
			// a shared ancestor look like a cycle, while starting each branch
			// from empty would let a two-block cycle recurse forever.
			var agreed []core.LinearOperation
			for index, candidate := range candidates {
				branchVisited := make(map[string]bool, len(visited)+1)
				for blockID, seen := range visited {
					branchVisited[blockID] = seen
				}
				tail, ok := rederive(candidate, branchVisited)
				if !ok {
					return nil, false
				}
				if index == 0 {
					agreed = tail
					continue
				}
				if !v.check(sameReleaseHistory(agreed, tail), "core.release_order_merge_mismatch", currentBlockID) {
					return nil, false
				}
			}
			expected = append(expected, agreed...)
			return expected, true
		}
		return expected, true
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
		// A terminal block is not required to have exactly one incoming
		// edge -- e.g. `discard ... because ...` legitimately merges its
		// ok/err edges into the same successor block, and that is fine
		// whenever neither path leaves a tracked acquisition needing
		// release. What this rederivation cannot tolerate is SKIPPING the
		// check for a merge point: every incoming edge must be walked
		// backward independently and the block's single, fixed release
		// list must match what each of them expects. If a hand-corrupted
		// program merged two chains with genuinely different completed-
		// acquisition sets, the actual release list can match at most one
		// of them -- so checking against every incoming edge, not skipping
		// the block, is what closes the gap CR-01 identified.
		if !v.check(len(incoming) > 0, "core.release_order_indeterminate", block.ID) {
			return false
		}
		var actual []core.LinearOperation
		for _, opID := range block.OperationIDs {
			if operation := operationsByID[opID]; operation.Kind == core.OpRelease {
				actual = append(actual, operation)
			}
		}
		for _, edge := range incoming {
			expected, ok := rederive(edge, make(map[string]bool, len(linear.Blocks)))
			if !ok {
				return false
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
	}
	return true
}

// sameReleaseHistory compares two independently rederived release histories
// at an interior merge. Two histories agree only when they have the same
// length and, at every index, the same operation ID -- the same discipline
// the per-terminal-block loop above already applies to its own comparison,
// extended one hop earlier. It performs no counted work of its own; all
// counted work for checkReleaseOrder stays inside checkReleaseOrder.
func sameReleaseHistory(left, right []core.LinearOperation) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].ID != right[index].ID {
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

// validCIdentifier audits the exact shape `cgen` splices unsanitized into an
// extern declaration and a call-expression callee (cgen.go:352, cgen.go:778,
// and the call-expression callee that reuses foreignExternName's result).
// Every other emitted identifier is routed through cName/cLocal, which
// replace every character outside [A-Za-z0-9_]; Symbol deliberately cannot
// take that route, because foreignExternName must match a real exported
// symbol in the byte-frozen foreign translation unit verbatim (cgen.go:
// 771-777) -- so the audit has to happen on the INPUT rather than on the
// output. This is a hand-rolled byte loop, matching cName/cLocal's existing
// idiom, rather than a regexp import: the predicate is exactly
// ^[A-Za-z_][A-Za-z0-9_]*$, checked byte-by-byte (not rune-by-rune) so a
// multibyte rune is rejected by its individual bytes.
func validCIdentifier(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	if !(first >= 'A' && first <= 'Z' || first >= 'a' && first <= 'z' || first == '_') {
		return false
	}
	for index := 1; index < len(name); index++ {
		b := name[index]
		if !(b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9' || b == '_') {
			return false
		}
	}
	return true
}

// commentSafe audits the shape 04-13 Task 3 requires of a
// core.ForeignContract string field cgen.EmitForeignHeader splices raw into
// a C COMMENT (cgen.go:1336-1348: fails, initialized_state, capture,
// retention, aliasing). An EMPTY string is comment-safe -- emptiness is a
// separate concern, already covered by foreign.obligation_undeclared and
// foreign.layout_invalid, and Layout.Fields[].CType is legitimately empty
// (cgen substitutes "unsigned char"). Refuses the two-byte sequence "*/" or
// "/*", any byte below 0x20, 0x7F, or any byte at or above 0x80. "/*" is
// refused alongside "*/" (not merely the closing sequence) because this
// project compiles generated C with warnings-as-errors, and a "/*" nested
// inside an already-open comment is a diagnosed condition on most
// compilers -- admitting it would turn a hostile value into a build break
// rather than a clean refusal.
func commentSafe(value string) bool {
	if value == "" {
		return true
	}
	for index := 0; index < len(value); index++ {
		b := value[index]
		if b < 0x20 || b == 0x7F || b >= 0x80 {
			return false
		}
		if index+1 < len(value) {
			pair := value[index : index+2]
			if pair == "*/" || pair == "/*" {
				return false
			}
		}
	}
	return true
}

// validCTypeExpression audits Layout.Fields[].CType, which
// EmitForeignConformance splices into a _Static_assert operand as a REAL C
// TOKEN sequence (cgen.go: the sizeof(%s)/_Alignof(%s) operands), not into
// a comment -- comment-safety alone is insufficient here. An EMPTY string is
// valid (cgen substitutes "unsigned char" at emission). Otherwise the value
// must be one or more C identifiers separated by exactly one single space,
// with no leading or trailing space and no double space: a hand-rolled byte
// scan splits on a single 0x20 byte and requires every part to satisfy
// validCIdentifier (no strings.Split import, matching this package's
// existing hand-rolled-byte-loop idiom). This exists because the sole
// honest value, "unsigned char" (check.go's standardForeignLayout), itself
// contains a space, so the plain identifier rule used for the three policy
// values would wrongly refuse it.
func validCTypeExpression(value string) bool {
	if value == "" {
		return true
	}
	start := 0
	for index := 0; index <= len(value); index++ {
		if index == len(value) || value[index] == ' ' {
			if !validCIdentifier(value[start:index]) {
				return false
			}
			start = index + 1
		}
	}
	return true
}

// foreignContractFieldsCSafe is the single predicate covering every
// core.ForeignContract string field EmitForeignHeader's and
// EmitForeignConformance's remaining splice sites reach (cgen.go:1336-1348's
// comment block, plus EmitForeignConformance's _Static_assert operands over
// Layout.ForeignTypeName/Fields[].Name/Fields[].CType) that Task 2's
// identifier audit does not already cover (Symbol, Allocator, Unwind and
// NonlocalExit are audited separately). A future emitter adding a new
// splice site over one of these fields has this doc comment and this
// function as its named place to extend.
func foreignContractFieldsCSafe(contract *core.ForeignContract) bool {
	// Alias is Phase 4 plan 06's additive aliasing-obligation field
	// (D-04-28); D-05-36 closes the carried D-04-33 audit gap by giving it
	// the exact same comment-safety check as its Aliasing sibling. This is
	// an independent re-derivation of cgen's own unsafeForeignContractField
	// Alias check (D-12): neither file imports the other.
	if !commentSafe(contract.Fails) || !commentSafe(contract.InitializedState) || !commentSafe(contract.Capture) || !commentSafe(contract.Retention) || !commentSafe(contract.Aliasing) || !commentSafe(contract.Alias) {
		return false
	}
	if contract.Layout == nil {
		return true
	}
	if !validCIdentifier(contract.Layout.ForeignTypeName) {
		return false
	}
	for _, field := range contract.Layout.Fields {
		if !validCIdentifier(field.Name) || !validCTypeExpression(field.CType) {
			return false
		}
	}
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

// ---------------------------------------------------------------------
// D-05-03b/D-05-04: independent re-derivation of an emitted attribute's
// justification. This section shares no helper with the checker's own
// alias-fact derivation or the C emitter's own emission logic -- D-12's
// three independent derivations.
// ---------------------------------------------------------------------

// AttributeClaim is corevalidate's OWN local decoding of one lang.foreign/0
// sidecar emitted_attributes entry (D-05-04/D-12): declared here, never
// imported from cgen.EmittedAttribute, so this validator's refusal logic
// never shares a type -- let alone a helper -- with the producer it audits.
type AttributeClaim struct {
	Attr        string
	CoreNode    string
	Parameter   string
	JustifiedBy string
}

// AttributeUnjustifiedError is returned by ValidateEmittedAttributes when an
// entry's attribute name is not one D-05-01 proves justifiable, or its
// claimed justification does not match the independently re-derived one.
// Code is always "core.attribute_unjustified" (D-05-03b) -- joining the
// lang.diagnostic/1 taxonomy without moving any existing ID (D-05-39).
type AttributeUnjustifiedError struct {
	Code   string
	Detail string
}

func (e *AttributeUnjustifiedError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Detail) }

// recomputeAliasJustifications independently re-derives, for every function
// in program whose by-pointer shape (D-05-01's whole-call exclusivity
// condition) is present, the loan ID that justifies an emitted `restrict`
// on its sole parameter -- keyed by "<function id>/<parameter id>". It uses
// corevalidate's OWN existing independent loan-endpoint re-derivation
// (recomputeLoanEndpoints's reachability-closure-plus-reduction mechanism),
// fed a synthetic single-block CFG view of the function's own straight-line
// operations (recomputeLoanEndpoints itself only walks a function whose
// Linear.Blocks are already populated) -- never check's fixpoint, and
// shares no helper with check or cgen (D-12).
func recomputeAliasJustifications(program core.Program) map[string]string {
	justifications := make(map[string]string)
	v := &validator{program: program}
	for _, function := range program.Functions {
		if function.Linear == nil || function.Match != nil || function.PublicOrigin != nil || len(function.Linear.Blocks) > 0 {
			continue
		}
		operations := function.Linear.Operations
		if len(operations) == 0 {
			continue
		}
		first := operations[0]
		if first.Kind != core.OpBorrowExclusive || first.SourceID != function.Parameter.ID || first.TargetID == "" {
			continue
		}
		current := first.TargetID
		terminatorIndex := -1
		valid := true
		for index := 1; index < len(operations); index++ {
			operation := operations[index]
			if operation.SourceID == function.Parameter.ID || operation.SourceID != current {
				valid = false
				break
			}
			if operation.Kind == core.OpReturn {
				terminatorIndex = index
				break
			}
			if operation.TargetID == "" {
				valid = false
				break
			}
			current = operation.TargetID
		}
		if !valid || terminatorIndex != len(operations)-1 {
			continue
		}
		terminator := operations[terminatorIndex]

		opIDs := make([]string, len(operations))
		for index, operation := range operations {
			opIDs[index] = operation.ID
		}
		synthetic := function
		synthetic.Linear = &core.LinearBody{
			ID: function.Linear.ID, Types: function.Linear.Types, Places: function.Linear.Places,
			Operations: operations,
			Blocks:     []core.Block{{ID: function.ID + ":block:straight", OperationIDs: opIDs, Successors: nil}},
		}
		endpoints := v.recomputeLoanEndpoints(&synthetic)
		for _, endpoint := range endpoints {
			if endpoint.LoanID != first.LoanID {
				continue
			}
			if endpoint.Kind == "point" && endpoint.AfterOperationID == terminator.ID {
				justifications[function.ID+"/"+function.Parameter.ID] = first.LoanID
			}
			break
		}
	}
	return justifications
}

// ValidateEmittedAttributes independently re-derives, from program's own
// loan facts alone, the justification for every claimed emitted attribute,
// and refuses any entry whose name is not proven-justifiable or whose
// claimed justification does not match the re-derived one (D-05-03b/D-05-04).
// It never reads cgen's EmittedAttributes as its own evidence: attributes is
// corevalidate's OWN AttributeClaim decoding of the sidecar's JSON, and this
// function's own re-derivation (recomputeAliasJustifications) shares no
// helper with check's deriveAliasFacts or cgen's emission logic (D-12).
func ValidateEmittedAttributes(program core.Program, attributes []AttributeClaim) error {
	justifications := recomputeAliasJustifications(program)
	for _, attribute := range attributes {
		if attribute.Attr != "restrict" {
			return &AttributeUnjustifiedError{Code: "core.attribute_unjustified", Detail: fmt.Sprintf("attribute %q is not a proven-justifiable attribute", attribute.Attr)}
		}
		expected, ok := justifications[attribute.CoreNode+"/"+attribute.Parameter]
		if !ok || expected == "" {
			return &AttributeUnjustifiedError{Code: "core.attribute_unjustified", Detail: fmt.Sprintf("no independently re-derived justification for %s/%s", attribute.CoreNode, attribute.Parameter)}
		}
		if attribute.JustifiedBy != expected {
			return &AttributeUnjustifiedError{Code: "core.attribute_unjustified", Detail: fmt.Sprintf("claimed justification %q does not match re-derived %q for %s/%s", attribute.JustifiedBy, expected, attribute.CoreNode, attribute.Parameter)}
		}
	}
	return nil
}

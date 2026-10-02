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
	"strconv"

	"github.com/szTheory/schway/internal/compiler/core"
)

const maxTypeDepth = 64

// phase17ReturnLookupFaultForTest is a deliberately narrow, default-off
// mutation seam for the core peer's return-side lookup. It is consulted only
// after the parameter contract has already been derived, so it cannot become
// an alternate path for parameter Drops or for a producer-supplied contract.
var phase17ReturnLookupFaultForTest bool

// SetPhase17ReturnLookupFaultForTest toggles only corevalidate's local return
// ability lookup. The control carries no type or contract value; callers can
// observe the peer's independently-derived signature without being given a
// derived return fact. Its restore closure is idempotent for defer/Cleanup
// callers and always restores the state that preceded this invocation.
func SetPhase17ReturnLookupFaultForTest(enabled bool) (restore func()) {
	previous := phase17ReturnLookupFaultForTest
	phase17ReturnLookupFaultForTest = enabled
	restored := false
	return func() {
		if restored {
			return
		}
		restored = true
		phase17ReturnLookupFaultForTest = previous
	}
}

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
//
// The further +1 flat term (from 17*facts+15 to 17*facts+16) is Plan
// 10-05 Task 2's own peerCalleeFrameDrained invariant (D-10-33/D-10-34): one
// v.check call inside derivePeerSignature, run exactly ONCE PER FUNCTION
// DECLARATION (this file's own scale fixtures declare exactly one
// function), never once per fact/operation -- the identical "constant
// addition, not a per-operation change" reasoning as 07-07's own entry
// immediately above.
func LinearWorkLimit(facts int) int { return 17*facts + 17 }

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

	// peerConsultedFields carries Phase 09 Plan 06's own D-09-30 disclosure
	// fact -- see PeerConsultedFields below.
	peerConsultedFields map[string]bool

	// peerLoanEndpoints carries plan 10-08's own D-10-52 seam: the SAME
	// []core.LoanEndpoint value loanEndpointsMatch already computed via
	// recomputeLoanEndpoints during this Validate call, keyed by function
	// ID -- never re-derived a second way -- see LoanEndpoints below.
	peerLoanEndpoints map[string][]core.LoanEndpoint
}

// Program returns a content-owned copy of the validated program. Callers can
// neither mutate the validator's copy nor race validation by retaining slices.
func (r Result) Program() core.Program { return cloneProgram(r.program) }

// LoanEndpoints returns corevalidate's own independently-recomputed
// core.LoanEndpoint set for every function in the validated program, keyed
// by function ID -- the missing seam D-10-52 names. `check` already
// materializes the identical shape via materializeLoanEndpoints
// (check.go:1726) and pathoracle.RecomputeEndpoints is already exported, so
// this was the only unreachable-from-outside computation among the three
// static peers; comparing only admit/refuse in a differential would be a
// regression from information already in this codebase.
//
// This is NOT a second derivation: every value here is the SAME
// recomputeLoanEndpoints result loanEndpointsMatch already computed and
// checked, byte-for-byte, during this same Validate call -- captured into
// v.peerLoanEndpoints as it was produced, never recomputed afterward
// against a fresh, unpopulated validator (which would silently drop
// v.peerLoanCarry's interprocedural OpCall facts and diverge from what
// Validate actually checked).
//
// A function with no Blocks (a straight-line linear body, or a match-only
// function with no Linear body at all) has an empty (nil) slice rather
// than panicking or being omitted -- every function ID in the program gets
// an entry.
//
// Built by iterating r.program.Functions -- the checker's own
// deterministically-ordered slice -- rather than ranging any Go map to
// construct the result, so the output is byte-identical across repeated
// calls (TestLoanEndpointsAccessorStableAcrossRepeatedCalls). Callers get a
// fresh copy on every call; mutating the returned map or its slices never
// affects this Result.
func (r Result) LoanEndpoints() map[string][]core.LoanEndpoint {
	out := make(map[string][]core.LoanEndpoint, len(r.program.Functions))
	for _, function := range r.program.Functions {
		endpoints := r.peerLoanEndpoints[function.ID]
		out[function.ID] = append([]core.LoanEndpoint(nil), endpoints...)
	}
	return out
}

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

// PeerConsultedFields returns corevalidate's own record of every
// callee-signature field name its interprocedural derivation consulted
// during this Validate call (D-09-30) -- a fresh, sorted copy on every
// call, mirroring PeerSignatures' own contract. This exists so a test
// harness with access to both this package and check can compare this set
// against check's own interproceduralConsultObserved record and assert
// they are IDENTICAL; this package never imports check to make that
// comparison itself. Matching sets constrain what each peer may READ
// (SEM-05 body-blindness), never how it derives -- proven on OUTPUTS,
// exactly as PeerSignatures/ClosureDigest byte-equality already are.
func (r Result) PeerConsultedFields() []string {
	out := make([]string, 0, len(r.peerConsultedFields))
	for field := range r.peerConsultedFields {
		out = append(out, field)
	}
	sort.Strings(out)
	return out
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
		peerConsultedFields: v.peerConsultedFields,
		peerLoanEndpoints:   v.peerLoanEndpoints,
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

	// peerConsultedFields accumulates every callee-signature field name
	// this validator's interprocedural derivation consulted during this
	// Validate call (D-09-30) -- see recordPeerConsult and
	// Result.PeerConsultedFields.
	peerConsultedFields map[string]bool

	// peerLoanEndpoints carries plan 10-08's own D-10-52 seam: the SAME
	// []core.LoanEndpoint value loanEndpointsMatch already computed via
	// recomputeLoanEndpoints during this Validate call, keyed by function
	// ID -- never re-derived a second way -- see Result.LoanEndpoints.
	peerLoanEndpoints map[string][]core.LoanEndpoint

	// peerAdjacency/peerPostorder are checkCallGraphAcyclic's own byproduct
	// (07-08, D-07-38/D-07-22): the SAME deduped adjacency it already
	// builds to prove acyclicity, plus the raw DFS postorder (callee
	// FINISHES before caller) its traversal already produces for free --
	// chainPeerClosureDigests below consumes both to chain ClosureDigest
	// over corevalidate's OWN, independently-derived route, never
	// callgraph's or originvalidate's.
	peerAdjacency map[string][]string
	peerPostorder []string

	// peerLoanCarry is Phase 09's own interprocedural loan-liveness fact
	// (D-09-01/D-09-02), folded into this SAME postorder substrate:
	// chainPeerLoanCarry (corevalidate_peer_liveness.go) populates this map
	// by walking v.peerPostorder callee-before-caller, exactly like
	// chainPeerClosureDigests above. Consulted by buildLoanChainIndex's
	// OpCall branch before propagating a loan across a call boundary. Never
	// persisted across Validate calls (D-09-06).
	peerLoanCarry map[string]peerLoanCarryFact
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
		// D-12-09's second mechanism: core.NewDataType is the only permitted
		// constructor for a DataType carrying payload alternatives, and it
		// enforces name-set equality between Alternatives and
		// AlternativeDetails AT CONSTRUCTION. That guard cannot protect a
		// program that arrives already built -- decoded from JSON, produced
		// by a future pass, or assembled in a test -- bypassing the
		// constructor entirely. This arm re-asserts the SAME invariant
		// corpus-wide, independently of whether NewDataType was ever called:
		// every AlternativeDetails entry must name an alternative present in
		// Alternatives, and no name may be named twice. Protobuf's oneof and
		// FlatBuffers' union-type-vector pairing are kept in sync by
		// generated accessors for exactly this reason -- a name present in
		// one list and absent from the other is representable but invalid.
		detailNames := make(map[string]struct{}, len(dataType.AlternativeDetails))
		for _, detail := range dataType.AlternativeDetails {
			if _, present := alternatives[detail.Name]; !v.check(present, "core.alternative_details_desynchronized", detail.Name) {
				return
			}
			if !v.unique(detailNames, detail.Name, "core.alternative_details_desynchronized") {
				return
			}
		}
		dataNames[dataType.Name] = dataType
	}
	functionIDs := make(map[string]struct{}, len(v.program.Functions))
	pending := make([]pendingReplayEntry, 0, len(v.program.Functions))
	requiresSchema1 := false
	for index := range v.program.Functions {
		function := &v.program.Functions[index]
		if function.Linear != nil {
			requiresSchema1 = true
			break
		}
	}
	if v.program.Schema == core.Schema1 && !requiresSchema1 {
		if len(v.problems) == 0 {
			v.problems = append(v.problems, Problem{Code: "core.schema", Detail: "lang.core/1 requires a linear body"})
		}
		return
	}
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
			if !v.check(v.program.Schema == core.Schema || v.program.Schema == core.Schema1, "core.schema", "match body requires lang.core/0 or lang.core/1") || !v.match(function, dataNames) {
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
		// D-09-01: chainPeerLoanCarry must run here, immediately after
		// v.peerPostorder is proven acyclic and populated, and strictly
		// BEFORE the pending-replay loop below -- unlike
		// chainPeerClosureDigests (which only needs to run after every
		// function's peer signature is recorded), buildLoanChainIndex
		// already reads v.peerLoanCarry from INSIDE replayStraightLine/
		// replayBlocks/recomputeLoanEndpoints, all of which execute as part
		// of the pending-replay loop immediately below. Calling this any
		// later would read a nil map and fail closed on every call.
		v.chainPeerLoanCarry()
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

	if !v.validateTransferredOwner() {
		return
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
	scrutineeType := function.Parameter.Type
	knownScrutinee := match.Scrutinee == function.Parameter.Name
	computedScrutinee := false
	if function.Linear != nil && match.ScrutineeID != "" {
		knownScrutinee = false
		for _, place := range function.Linear.Places {
			if place.ID != match.ScrutineeID {
				continue
			}
			if place.Name != match.Scrutinee {
				continue
			}
			for _, fact := range function.Linear.Types {
				if fact.ID == place.TypeID {
					scrutineeType = fact.Shape.Constructor
					knownScrutinee = true
					computedScrutinee = place.ID != function.Parameter.ID
					break
				}
			}
		}
		if computedScrutinee && !v.computedScrutineeDefinedInEntry(function, match.ScrutineeID) {
			return nil, nil, v.check(false, "core.unknown_place", match.Scrutinee)
		}
	}
	dataType, knownScrutineeType := dataNames[scrutineeType]
	_, knownReturnType := dataNames[function.ReturnType]
	if !v.check(knownScrutineeType && knownReturnType, "core.unknown_type", scrutineeType) ||
		!v.check(knownScrutinee, "core.unknown_place", match.Scrutinee) {
		return nil, nil, false
	}
	alternatives := make(map[string]struct{}, len(dataType.Alternatives))
	for _, alternative := range dataType.Alternatives {
		alternatives[alternative] = struct{}{}
	}
	armIDs := make(map[string]struct{}, len(match.Arms))
	edgeIDs := make(map[string]struct{}, len(match.Arms))
	patterns := make(map[string]struct{}, len(match.Arms))
	valuePlaceIDs := make(map[string]core.MatchArm, len(match.Arms))
	seenValuePlaceIDs := make(map[string]struct{}, len(match.Arms))
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
		// A branch body ordinarily carries only BlockID. A differing
		// scrutinee/return type can additionally carry Value and an
		// edge-bound ValuePlaceID; the place is initialized only when this
		// arm is selected and is consumed by its terminal OpReturn.
		if !v.check(arm.BlockID != "" && ((arm.Value == "" && arm.ValuePlaceID == "") || (arm.Value != "" && arm.ValuePlaceID != "")), "core.invalid_body", arm.ID) {
			return nil, nil, false
		}
		if arm.ValuePlaceID != "" {
			returnDataType := dataNames[function.ReturnType]
			if !v.check(containsAlternative(returnDataType, arm.Value), "core.unknown_alternative", arm.Value) ||
				!v.unique(seenValuePlaceIDs, arm.ValuePlaceID, "core.duplicate_place_id") {
				return nil, nil, false
			}
			valuePlaceIDs[arm.ValuePlaceID] = arm
		}
	}
	if !v.check(len(patterns) == len(alternatives), "core.final_claim_mismatch", match.ID) {
		return nil, nil, false
	}
	types, places, valid := v.linearStructural(function)
	if !valid {
		return nil, nil, false
	}
	for valuePlaceID, arm := range valuePlaceIDs {
		place, exists := places[valuePlaceID]
		fact, typeKnown := types[place.TypeID]
		if !v.check(exists && typeKnown && fact.Shape.Constructor == function.ReturnType, "core.return_type_mismatch", valuePlaceID) {
			return nil, nil, false
		}
		armBlock := ""
		for _, candidate := range match.Arms {
			if candidate.ID == arm.ID {
				armBlock = candidate.BlockID
				break
			}
		}
		var matchingReturns int
		for _, block := range function.Linear.Blocks {
			for _, operationID := range block.OperationIDs {
				operation := findLinearOperation(function.Linear.Operations, operationID)
				if operation.SourceID != valuePlaceID {
					continue
				}
				if block.ID == armBlock && operation.Kind == core.OpReturn && operation.TypeID == place.TypeID {
					matchingReturns++
				} else {
					v.check(false, "core.return_type_mismatch", valuePlaceID)
					return nil, nil, false
				}
			}
		}
		if !v.check(matchingReturns == 1, "core.return_type_mismatch", valuePlaceID) {
			return nil, nil, false
		}
	}
	return types, places, true
}

func containsAlternative(dataType core.DataType, name string) bool {
	for _, alternative := range dataType.Alternatives {
		if alternative == name {
			return true
		}
	}
	return false
}

func findLinearOperation(operations []core.LinearOperation, id string) core.LinearOperation {
	for _, operation := range operations {
		if operation.ID == id {
			return operation
		}
	}
	return core.LinearOperation{}
}

// computedScrutineeDefinedInEntry independently proves that a non-parameter
// match place is produced by the straight-line prefix in the branch entry
// block. Merely finding a matching Place in the flattened arm list would let
// forged core borrow a place that does not exist until after dispatch.
func (v *validator) computedScrutineeDefinedInEntry(function *core.Function, placeID string) bool {
	if function.Linear == nil || placeID == "" {
		return false
	}
	entryID := function.ID + ":block:entry"
	operations := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operations[operation.ID] = operation
	}
	for _, block := range function.Linear.Blocks {
		if block.ID != entryID || len(block.Successors) == 0 || len(block.OperationIDs) == 0 {
			continue
		}
		for _, operationID := range block.OperationIDs {
			operation, exists := operations[operationID]
			if !exists {
				return false
			}
			if operation.TargetID == placeID || operation.Kind == core.OpDestructurePayload && operation.PayloadTargetID == placeID {
				return true
			}
		}
	}
	return false
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
	v.peerSignatures[function.ID] = v.derivePeerSignature(function, nil, nil)
	match := function.Match
	if !v.check(match.ID != "" && match.PointID != "" && function.EntryPointID != "" && function.ReturnPointID != "", "core.invalid_id", function.ID) {
		return false
	}
	armIDs := make(map[string]struct{}, len(match.Arms))
	edgeIDs := make(map[string]struct{}, len(match.Arms))
	scrutineeType := function.Parameter.Type
	knownScrutinee := match.Scrutinee == function.Parameter.Name
	if function.Linear != nil {
		knownScrutinee = false
		for _, place := range function.Linear.Places {
			if match.ScrutineeID != "" && place.ID != match.ScrutineeID {
				continue
			}
			if match.ScrutineeID == "" && place.Name != match.Scrutinee {
				continue
			}
			if place.Name != match.Scrutinee {
				continue
			}
			for _, fact := range function.Linear.Types {
				if fact.ID == place.TypeID {
					scrutineeType = fact.Shape.Constructor
					knownScrutinee = true
					break
				}
			}
		}
	}
	dataType, knownScrutineeType := dataNames[scrutineeType]
	returnDataType, knownReturnType := dataNames[function.ReturnType]
	if !v.check(knownScrutineeType && knownReturnType, "core.unknown_type", scrutineeType) ||
		!v.check(knownScrutinee, "core.unknown_place", match.Scrutinee) {
		return false
	}
	alternatives := make(map[string]struct{}, len(dataType.Alternatives))
	for _, alternative := range dataType.Alternatives {
		alternatives[alternative] = struct{}{}
	}
	returnAlternatives := make(map[string]struct{}, len(returnDataType.Alternatives))
	for _, alternative := range returnDataType.Alternatives {
		returnAlternatives[alternative] = struct{}{}
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
		_, valueKnown := returnAlternatives[arm.Value]
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
		wantTypeID := fmt.Sprintf("%s:type:%d", function.ID, index)
		literalU64ID := fact.ID == function.ID+":type:u64" && fact.Shape.Constructor == "U64" && len(fact.Shape.Arguments) == 0
		if !v.check(fact.ID == wantTypeID || literalU64ID, "core.type_order", fact.ID) {
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
		if operation.Kind == core.OpConst {
			if !v.check(operation.SourceID == "", "core.source_kind_exclusive", operation.ID) {
				return nil, nil, false
			}
		} else if _, ok := places[operation.SourceID]; !v.check(ok, "core.unknown_place", operation.SourceID) {
			return nil, nil, false
		}
		if _, ok := types[operation.TypeID]; !v.check(ok, "core.unknown_type", operation.TypeID) {
			return nil, nil, false
		}
		// OpFail and OpDefect are terminators alongside OpReturn (D-04-04/
		// D-04-15): none of the three ever carries a TargetID, since none
		// produces an ordinary place -- OpReturn ends the function, OpFail
		// ends the err block, OpDefect ends an arm block by aborting.
		// OpDestructurePayload joins the terminator exemption above (Phase
		// 12, D-12-10): it names its own produced place via PayloadTargetID,
		// never the ordinary TargetID, which it leaves empty -- checked
		// separately immediately below.
		if operation.Kind != core.OpReturn && operation.Kind != core.OpFail && operation.Kind != core.OpRelease && operation.Kind != core.OpDefect && operation.Kind != core.OpDestructurePayload && operation.Kind != core.OpScalarStore && operation.Kind != core.OpBranch {
			if _, ok := places[operation.TargetID]; !v.check(ok, "core.unknown_place", operation.TargetID) {
				return nil, nil, false
			}
		}
		if operation.Kind == core.OpDestructurePayload {
			if !v.check(operation.TargetID == "", "core.invalid_target", operation.ID) {
				return nil, nil, false
			}
			if _, ok := places[operation.PayloadTargetID]; !v.check(ok, "core.unknown_place", operation.PayloadTargetID) {
				return nil, nil, false
			}
		}
		switch operation.Kind {
		case core.OpAddChecked, core.OpLessU64:
			right, known := places[operation.RightID]
			want := "U64"
			if operation.Kind == core.OpLessU64 {
				want = "Bool"
			}
			if !v.check(known && types[places[operation.SourceID].TypeID].Shape.Constructor == "U64" && types[right.TypeID].Shape.Constructor == "U64" && types[operation.TypeID].Shape.Constructor == want && operation.StoreTargetID == "" && operation.TrueEdgeID == "" && operation.FalseEdgeID == "", "core.scalar_operation_invalid", operation.ID) {
				return nil, nil, false
			}
		case core.OpScalarStore:
			target, known := places[operation.StoreTargetID]
			if !v.check(known && target.Mutable && target.ID != function.Parameter.ID && target.TypeID == places[operation.SourceID].TypeID && operation.TargetID == "" && operation.RightID == "" && operation.TrueEdgeID == "" && operation.FalseEdgeID == "", "core.scalar_store_invalid", operation.ID) {
				return nil, nil, false
			}
		case core.OpBranch:
			if !v.check(types[operation.TypeID].Shape.Constructor == "Bool" && operation.TargetID == "" && operation.RightID == "" && operation.StoreTargetID == "" && operation.TrueEdgeID != "" && operation.FalseEdgeID != "", "core.scalar_branch_invalid", operation.ID) {
				return nil, nil, false
			}
		}
		if operation.Kind == core.OpForeignCall {
			if operation.Foreign != nil {
				if !v.validateLocalForeignOperation(function, operation, types, places) {
					return nil, nil, false
				}
			} else {
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
		} else if operation.Foreign != nil {
			if operation.Kind != core.OpRelease || !v.validateLocalForeignOperation(function, operation, types, places) {
				return nil, nil, false
			}
		}
		if operation.Kind == core.OpCall && (operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "") {
			errTarget, known := places[operation.ErrTargetID]
			if !v.check(operation.ErrTargetID != "" && operation.OkEdgeID != "" && operation.ErrEdgeID != "" && known && errTarget.TypeID != "" && len(linear.Blocks) != 0, "core.call_error_edges_invalid", operation.ID) {
				return nil, nil, false
			}
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			if !v.unique(loanIDs, operation.LoanID, "core.unknown_loan") {
				return nil, nil, false
			}
		} else if !v.check(operation.LoanID == "" && (operation.Kind == core.OpConst || operation.ConstU64 == ""), "core.unknown_loan", operation.LoanID) {
			return nil, nil, false
		}
		if operation.Kind == core.OpConst {
			value, err := strconv.ParseUint(operation.ConstU64, 10, 64)
			shape := types[operation.TypeID].Shape
			if !v.check(err == nil && strconv.FormatUint(value, 10) == operation.ConstU64 && operation.TargetID != "" && operation.TypeID != "" && shape.Constructor == "U64" && len(shape.Arguments) == 0 && operation.CalleeID == "" && operation.PayloadType == "" && operation.PayloadTargetID == "" && operation.ErrTargetID == "" && operation.OkEdgeID == "" && operation.ErrEdgeID == "" && operation.ReleasesOperationID == "" && operation.Allocator == "" && operation.Reason == "", "core.constant_invalid", operation.ID) {
				return nil, nil, false
			}
		}
		if operation.Kind != core.OpAddChecked && operation.Kind != core.OpLessU64 && operation.RightID != "" || operation.Kind != core.OpScalarStore && operation.StoreTargetID != "" || operation.Kind != core.OpBranch && (operation.TrueEdgeID != "" || operation.FalseEdgeID != "") {
			if !v.check(false, "core.scalar_field_kind_exclusive", operation.ID) {
				return nil, nil, false
			}
		}
	}
	if len(linear.Blocks) > 0 || len(linear.Edges) > 0 {
		if !v.blocksAndEdges(function, operationIDs, types, places) {
			return nil, nil, false
		}
	}
	return types, places, true
}

func (v *validator) validateLocalForeignOperation(function *core.Function, operation core.LinearOperation, types map[string]core.TypeFact, places map[string]core.Place) bool {
	contract := operation.Foreign
	if contract == nil {
		return false
	}
	if !v.check(validCIdentifier(contract.Symbol) && validCIdentifier(contract.ABIType), "foreign.operation_identifier_invalid", operation.ID) {
		return false
	}
	if !v.check(contract.Unwind == "forbidden" && contract.NonlocalExit == "forbidden", "foreign.operation_exit_policy_invalid", operation.ID) {
		return false
	}
	source, sourceOK := places[operation.SourceID]
	if !v.check(sourceOK && types[source.TypeID].Shape.Constructor == contract.ParameterType, "foreign.operation_parameter_mismatch", operation.ID) {
		return false
	}
	switch contract.Mode {
	case "acquire":
		target, targetOK := places[operation.TargetID]
		if operation.Kind != core.OpForeignCall || !targetOK || types[target.TypeID].Shape.Constructor != contract.ResultType || contract.Symbol != "schway_file_byte_acquire" || contract.ABIType != "schway_file_byte_acquire_fn" || contract.ParameterType != "PathToken" || contract.ResultType != "FileByteOwner" || contract.Fails != "AcquireError" || contract.Allocator != "libc_malloc" || contract.Release != "schway_file_byte_release" || operation.Allocator != contract.Allocator || !validLocalForeignErrorEdges(operation, contract, types, places) {
			return v.check(false, "foreign.operation_contract_invalid", operation.ID)
		}
	case "borrow":
		target, targetOK := places[operation.TargetID]
		if operation.Kind != core.OpForeignCall || !targetOK || types[target.TypeID].Shape.Constructor != contract.ResultType || contract.Symbol != "schway_file_byte_use" || contract.ABIType != "schway_file_byte_use_fn" || contract.ParameterType != "FileByteOwner" || contract.ResultType != "U64" || contract.Fails != "UseError" || contract.Allocator != "" || contract.Release != "" || operation.Allocator != "" || !validLocalForeignErrorEdges(operation, contract, types, places) {
			return v.check(false, "foreign.operation_contract_invalid", operation.ID)
		}
	case "consume":
		if operation.Kind != core.OpRelease || contract.Symbol != "schway_file_byte_release" || contract.ABIType != "schway_file_byte_release_fn" || contract.ParameterType != "FileByteOwner" || contract.ResultType != "Unit" || contract.Fails != "" || contract.Allocator != "libc_malloc" || contract.Release != "" || operation.Allocator != contract.Allocator || operation.TargetID != "" {
			return v.check(false, "foreign.operation_contract_invalid", operation.ID)
		}
	default:
		return v.check(false, "foreign.operation_mode_invalid", operation.ID)
	}
	if contract.Fails != "" && !validCIdentifier(contract.Fails) {
		return v.check(false, "foreign.operation_failure_invalid", operation.ID)
	}
	if contract.Allocator != "" && !validCIdentifier(contract.Allocator) {
		return v.check(false, "foreign.operation_allocator_invalid", operation.ID)
	}
	if contract.Release != "" && !validCIdentifier(contract.Release) {
		return v.check(false, "foreign.operation_release_invalid", operation.ID)
	}
	if isDeclaredFunctionName(v.program.Functions, function.ID, contract.Symbol) {
		return v.check(false, "core.call_target_not_foreign", operation.ID)
	}
	return true
}

func validLocalForeignErrorEdges(operation core.LinearOperation, contract *core.ForeignOperationContract, types map[string]core.TypeFact, places map[string]core.Place) bool {
	hasTarget := operation.ErrTargetID != ""
	hasEdges := operation.OkEdgeID != "" || operation.ErrEdgeID != ""
	if !hasTarget && !hasEdges {
		return true
	}
	target, ok := places[operation.ErrTargetID]
	return hasTarget && operation.OkEdgeID != "" && operation.ErrEdgeID != "" && ok && contract.Fails != "" && types[target.TypeID].Shape.Constructor == contract.Fails
}

// validateTransferredOwner derives a Phase 24 obligation from the helper's acquire operation,
// follows its OpReturn into the caller's OpCall, and requires one post-borrow release
// that names that exact acquisition. It never seeds state from release operations.
func (v *validator) validateTransferredOwner() bool {
	if phase24ErrorFunctionsPresent(v.program) {
		return v.validateRepeatedPhase24Owner()
	}
	type located struct {
		function *core.Function
		op       core.LinearOperation
	}
	var acquire located
	var helper *core.Function
	for i := range v.program.Functions {
		f := &v.program.Functions[i]
		if f.Linear == nil {
			continue
		}
		for _, op := range f.Linear.Operations {
			if op.Foreign != nil && op.Foreign.Mode == "acquire" {
				if acquire.function != nil {
					return v.check(false, "core.owner_transfer_acquire", op.ID)
				}
				acquire = located{f, op}
				helper = f
			}
		}
	}
	if helper == nil || helper.ReturnType != "FileByteOwner" {
		return true
	}
	if helper.Linear == nil || acquire.op.Kind != core.OpForeignCall || acquire.op.Foreign == nil || acquire.op.Foreign.Allocator == "" || acquire.op.Foreign.Release == "" {
		return v.check(false, "core.owner_transfer_helper", helper.ID)
	}
	var helperReturn *core.LinearOperation
	for i := range helper.Linear.Operations {
		op := &helper.Linear.Operations[i]
		if op.Kind == core.OpReturn {
			if helperReturn != nil {
				return v.check(false, "core.owner_transfer_return", op.ID)
			}
			helperReturn = op
		} else if op.Kind != core.OpForeignCall && op.Kind != core.OpFail {
			return v.check(false, "core.owner_transfer_helper_op", op.ID)
		}
	}
	if helperReturn == nil || helperReturn.SourceID != acquire.op.TargetID || acquire.op.ErrTargetID == "" || acquire.op.OkEdgeID == "" || acquire.op.ErrEdgeID == "" {
		return v.check(false, "core.owner_transfer_return", helper.ID)
	}
	var caller *core.Function
	var call, borrow, release, ret *core.LinearOperation
	for i := range v.program.Functions {
		f := &v.program.Functions[i]
		if f.Linear == nil {
			continue
		}
		for j := range f.Linear.Operations {
			op := &f.Linear.Operations[j]
			if op.Kind == core.OpCall && op.CalleeID == helper.ID {
				if caller != nil {
					return v.check(false, "core.owner_transfer_call", op.ID)
				}
				caller = f
				call = op
			}
		}
	}
	if caller == nil || caller.ReturnType != "U64" || call.TargetID == "" || call.TypeID == "" {
		return v.check(false, "core.owner_transfer_call", helper.ID)
	}
	for i := range caller.Linear.Operations {
		op := &caller.Linear.Operations[i]
		switch {
		case op.Kind == core.OpCopy && op.SourceID == call.TargetID:
			return v.check(false, "core.owner_transfer_copy", op.ID)
		case op.Foreign != nil && op.Foreign.Mode == "borrow":
			if borrow != nil {
				return v.check(false, "core.owner_transfer_borrow", op.ID)
			}
			borrow = op
		case op.Kind == core.OpRelease:
			if release != nil {
				return v.check(false, "core.owner_transfer_release", op.ID)
			}
			release = op
		case op.Kind == core.OpReturn:
			if ret != nil {
				return v.check(false, "core.owner_transfer_return", op.ID)
			}
			ret = op
		}
	}
	if borrow == nil || release == nil || ret == nil || borrow.SourceID != call.TargetID || release.SourceID != call.TargetID || release.ReleasesOperationID != acquire.op.ID || release.Foreign == nil || borrow.Foreign == nil || release.Foreign.Mode != "consume" || release.Foreign.Symbol != acquire.op.Foreign.Release || release.Allocator != acquire.op.Foreign.Allocator || release.Foreign.Allocator != acquire.op.Foreign.Allocator {
		return v.check(false, "core.owner_transfer_discharge", caller.ID)
	}
	if !(indexOperation(caller.Linear.Operations, borrow.ID) < indexOperation(caller.Linear.Operations, release.ID) && indexOperation(caller.Linear.Operations, release.ID) < indexOperation(caller.Linear.Operations, ret.ID)) {
		return v.check(false, "core.owner_transfer_order", caller.ID)
	}
	// The integrated utility may return the scalar through exactly two
	// independently checked copy helpers. Keep this proof deliberately
	// syntactic and local: arbitrary call results, extra calls, and reordered
	// chains do not inherit the direct borrowed-result admission above.
	if ret.SourceID != borrow.TargetID {
		var helperCalls []core.LinearOperation
		for _, op := range caller.Linear.Operations {
			if op.Kind == core.OpCall && op.ID != call.ID {
				helperCalls = append(helperCalls, op)
			}
		}
		if len(helperCalls) != 2 || !(indexOperation(caller.Linear.Operations, borrow.ID) < indexOperation(caller.Linear.Operations, helperCalls[0].ID) && indexOperation(caller.Linear.Operations, helperCalls[0].ID) < indexOperation(caller.Linear.Operations, helperCalls[1].ID) && indexOperation(caller.Linear.Operations, helperCalls[1].ID) < indexOperation(caller.Linear.Operations, release.ID) && indexOperation(caller.Linear.Operations, release.ID) < indexOperation(caller.Linear.Operations, ret.ID)) {
			return v.check(false, "core.owner_transfer_order", caller.ID)
		}
		calleeByID := make(map[string]*core.Function, len(v.program.Functions))
		for i := range v.program.Functions {
			calleeByID[v.program.Functions[i].ID] = &v.program.Functions[i]
		}
		shared, exclusive := calleeByID[helperCalls[0].CalleeID], calleeByID[helperCalls[1].CalleeID]
		if !isU64CopyHelper(shared, core.OpBorrowShared) || !isU64CopyHelper(exclusive, core.OpBorrowExclusive) || helperCalls[0].SourceID != borrow.TargetID || helperCalls[1].SourceID != helperCalls[0].TargetID || ret.SourceID != helperCalls[1].TargetID {
			return v.check(false, "core.owner_transfer_discharge", caller.ID)
		}
	}
	for _, op := range helper.Linear.Operations {
		if op.Kind == core.OpRelease {
			return v.check(false, "core.owner_transfer_early_release", op.ID)
		}
	}
	return true
}

func isU64CopyHelper(function *core.Function, borrowKind core.OperationKind) bool {
	if function == nil || function.Linear == nil || function.ReturnType != "U64" || function.Parameter.Type != "U64" || len(function.Linear.Operations) != 3 {
		return false
	}
	borrow, copy, ret := function.Linear.Operations[0], function.Linear.Operations[1], function.Linear.Operations[2]
	return borrow.Kind == borrowKind && borrow.SourceID == function.Parameter.ID && borrow.TargetID != "" &&
		copy.Kind == core.OpCopy && copy.SourceID == borrow.TargetID && copy.TargetID != "" &&
		ret.Kind == core.OpReturn && ret.SourceID == copy.TargetID
}

func phase24ErrorFunctionsPresent(program core.Program) bool {
	if program.Module != "phase24.error" {
		return false
	}
	names := map[string]bool{}
	for _, function := range program.Functions {
		names[function.Name] = true
	}
	return names["acquire"] && names["probe"] && names["main"]
}

func (v *validator) validateRepeatedPhase24Owner() bool {
	functions := make(map[string]*core.Function, len(v.program.Functions))
	for index := range v.program.Functions {
		functions[v.program.Functions[index].Name] = &v.program.Functions[index]
	}
	helper, probe, caller := functions["acquire"], functions["probe"], functions["main"]
	if helper == nil || probe == nil || caller == nil || helper.Linear == nil || probe.Linear == nil || caller.Linear == nil {
		return v.check(false, "core.owner_transfer_helper", "phase24 error functions")
	}
	var acquire *core.LinearOperation
	for index := range helper.Linear.Operations {
		operation := &helper.Linear.Operations[index]
		if operation.Kind == core.OpForeignCall && operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
			if acquire != nil {
				return v.check(false, "core.owner_transfer_acquire", operation.ID)
			}
			acquire = operation
		}
	}
	if acquire == nil || helper.Parameter.Type != "PathToken" || helper.ReturnType != "FileByteOwner" || acquire.Kind != core.OpForeignCall || acquire.Foreign == nil || acquire.Foreign.Allocator != "libc_malloc" || acquire.Foreign.Release != "schway_file_byte_release" {
		return v.check(false, "core.owner_transfer_helper", helper.ID)
	}
	returned := 0
	for _, operation := range helper.Linear.Operations {
		if operation.Kind == core.OpReturn {
			returned++
			if operation.SourceID != acquire.TargetID {
				return v.check(false, "core.owner_transfer_return", operation.ID)
			}
		}
		if operation.Kind == core.OpRelease {
			return v.check(false, "core.owner_transfer_early_release", operation.ID)
		}
	}
	if !v.check(returned == 1, "core.owner_transfer_return", helper.ID) {
		return false
	}
	acquireCalls, probeCalls := phase24CallsTo(caller, helper.ID), phase24CallsTo(caller, probe.ID)
	probeAcquireCalls := phase24CallsTo(probe, helper.ID)
	if !v.check(len(acquireCalls) == 2 && len(probeCalls) == 1 && len(probeAcquireCalls) == 1, "core.owner_transfer_call", helper.ID) {
		return false
	}
	for _, operation := range append(append(append([]core.LinearOperation{}, acquireCalls...), probeCalls...), probeAcquireCalls...) {
		if !v.validatePhase24CallEdges(operationOwnerFunction(operation, caller, probe), operation, "ResourceError") {
			return false
		}
	}
	// Confirm the calls are separate checked activation sites, even though
	// the helper's acquisition operation is one static core operation.
	seenCalls := map[string]bool{}
	for _, operation := range append(append(append([]core.LinearOperation{}, acquireCalls...), probeCalls...), probeAcquireCalls...) {
		if !v.check(operation.ID != "" && !seenCalls[operation.ID], "core.owner_transfer_call", operation.ID) {
			return false
		}
		seenCalls[operation.ID] = true
	}
	if !v.phase24ProbeCleanup(probe, probeAcquireCalls[0], *acquire, "ResourceError") {
		return false
	}
	return v.phase24CallerCleanup(caller, acquireCalls[0], acquireCalls[1], probeCalls[0], *acquire)
}

func phase24CallsTo(function *core.Function, calleeID string) []core.LinearOperation {
	var calls []core.LinearOperation
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == calleeID {
			calls = append(calls, operation)
		}
	}
	return calls
}

func operationOwnerFunction(operation core.LinearOperation, caller, probe *core.Function) *core.Function {
	for _, candidate := range []*core.Function{caller, probe} {
		for _, existing := range candidate.Linear.Operations {
			if existing.ID == operation.ID {
				return candidate
			}
		}
	}
	return nil
}

func (v *validator) validatePhase24CallEdges(function *core.Function, call core.LinearOperation, errorType string) bool {
	if function == nil || call.Kind != core.OpCall || call.ErrTargetID == "" || call.OkEdgeID == "" || call.ErrEdgeID == "" {
		return v.check(false, "core.call_error_edges_invalid", call.ID)
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	types := make(map[string]core.TypeFact, len(function.Linear.Types))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact
	}
	errTarget, known := places[call.ErrTargetID]
	if !v.check(known && types[errTarget.TypeID].Shape.Constructor == errorType, "core.call_error_type_mismatch", call.ID) {
		return false
	}
	if !v.check(v.phase24ErrorEnvelopeIncludes(call.CalleeID, errorType), "core.call_error_envelope_mismatch", call.ID) {
		return false
	}
	operationBlock := ""
	for _, block := range function.Linear.Blocks {
		for _, operationID := range block.OperationIDs {
			if operationID == call.ID {
				if operationBlock != "" {
					return v.check(false, "core.duplicate_operation_id", call.ID)
				}
				operationBlock = block.ID
			}
		}
	}
	ok, errEdge := phase24FindEdge(function.Linear.Edges, call.OkEdgeID), phase24FindEdge(function.Linear.Edges, call.ErrEdgeID)
	return v.check(operationBlock != "" && ok.ID != "" && errEdge.ID != "" && ok.FromBlockID == operationBlock && errEdge.FromBlockID == operationBlock && ok.Pattern == "ok" && errEdge.Pattern == "err" && phase24BlockHasSuccessor(function.Linear.Blocks, operationBlock, ok.ToBlockID) && phase24BlockHasSuccessor(function.Linear.Blocks, operationBlock, errEdge.ToBlockID), "core.call_error_edges_invalid", call.ID)
}

func (v *validator) phase24ErrorEnvelopeIncludes(calleeID, envelopeName string) bool {
	var callee *core.Function
	for index := range v.program.Functions {
		if v.program.Functions[index].ID == calleeID {
			callee = &v.program.Functions[index]
			break
		}
	}
	if callee == nil || callee.Linear == nil {
		return false
	}
	constructorByType := map[string]string{}
	for _, fact := range callee.Linear.Types {
		constructorByType[fact.ID] = fact.Shape.Constructor
	}
	failureTypes := map[string]bool{}
	for _, operation := range callee.Linear.Operations {
		if operation.Kind == core.OpFail {
			failureTypes[constructorByType[operation.TypeID]] = true
		}
		if operation.Foreign != nil && operation.Foreign.Fails != "" {
			failureTypes[operation.Foreign.Fails] = true
		}
	}
	if len(failureTypes) == 0 {
		return false
	}
	envelope := map[string]bool{}
	envelopeValid := false
	for _, dataType := range v.program.DataTypes {
		if dataType.Name == envelopeName && len(dataType.AlternativeDetails) == 0 {
			envelopeValid = true
			for _, alternative := range dataType.Alternatives {
				envelope[alternative] = true
			}
		}
	}
	if !envelopeValid || len(envelope) == 0 {
		return false
	}
	for failureType := range failureTypes {
		var alternatives []string
		for _, dataType := range v.program.DataTypes {
			if dataType.Name == failureType && len(dataType.AlternativeDetails) == 0 {
				alternatives = dataType.Alternatives
			}
		}
		if len(alternatives) == 0 {
			return false
		}
		for _, alternative := range alternatives {
			if !envelope[alternative] {
				return false
			}
		}
	}
	return true
}

func phase24FindEdge(edges []core.Edge, id string) core.Edge {
	for _, edge := range edges {
		if edge.ID == id {
			return edge
		}
	}
	return core.Edge{}
}

func phase24BlockHasSuccessor(blocks []core.Block, from, to string) bool {
	for _, block := range blocks {
		if block.ID != from {
			continue
		}
		for _, successor := range block.Successors {
			if successor == to {
				return true
			}
		}
	}
	return false
}

func (v *validator) phase24ProbeCleanup(probe *core.Function, acquireCall, acquisition core.LinearOperation, errorType string) bool {
	places := make(map[string]core.Place, len(probe.Linear.Places))
	types := make(map[string]core.TypeFact, len(probe.Linear.Types))
	for _, place := range probe.Linear.Places {
		places[place.ID] = place
	}
	for _, fact := range probe.Linear.Types {
		types[fact.ID] = fact
	}
	var use *core.LinearOperation
	for index := range probe.Linear.Operations {
		operation := &probe.Linear.Operations[index]
		if operation.Foreign != nil && operation.Foreign.Mode == "borrow" {
			if use != nil {
				return v.check(false, "core.owner_transfer_borrow", operation.ID)
			}
			use = operation
		}
	}
	if use == nil || use.Foreign.Fails != "UseError" || use.ErrTargetID == "" || use.OkEdgeID == "" || use.ErrEdgeID == "" || use.SourceID != acquireCall.TargetID {
		return v.check(false, "core.owner_transfer_borrow", probe.ID)
	}
	errPlace, exists := places[use.ErrTargetID]
	if !v.check(exists && types[errPlace.TypeID].Shape.Constructor == "UseError", "core.owner_transfer_error", use.ID) {
		return false
	}
	if !v.phase24TerminalCleanup(probe, phase24FindEdge(probe.Linear.Edges, use.OkEdgeID).ToBlockID, []string{acquireCall.TargetID}, use.TargetID, core.OpReturn, acquisition.ID) ||
		!v.phase24TerminalCleanup(probe, phase24FindEdge(probe.Linear.Edges, use.ErrEdgeID).ToBlockID, []string{acquireCall.TargetID}, use.ErrTargetID, core.OpFail, acquisition.ID) {
		return false
	}
	acquireError := phase24FindEdge(probe.Linear.Edges, acquireCall.ErrEdgeID)
	if !v.check(acquireError.ID != "" && v.phase24TerminalCleanup(probe, acquireError.ToBlockID, nil, acquireCall.ErrTargetID, core.OpFail, acquisition.ID), "core.owner_transfer_failed_acquire", acquireCall.ID) {
		return false
	}
	return true
}

func (v *validator) phase24CallerCleanup(caller *core.Function, callA, callB, callProbe, acquisition core.LinearOperation) bool {
	if !v.phase24TerminalCleanup(caller, phase24FindEdge(caller.Linear.Edges, callA.OkEdgeID).ToBlockID, nil, "", core.OpCall, acquisition.ID) {
		return false
	}
	if !v.phase24TerminalCleanup(caller, phase24FindEdge(caller.Linear.Edges, callA.ErrEdgeID).ToBlockID, nil, callA.ErrTargetID, core.OpFail, acquisition.ID) {
		return false
	}
	if !v.phase24TerminalCleanup(caller, phase24FindEdge(caller.Linear.Edges, callB.ErrEdgeID).ToBlockID, []string{callA.TargetID}, callB.ErrTargetID, core.OpFail, acquisition.ID) {
		return false
	}
	if !v.phase24TerminalCleanup(caller, phase24FindEdge(caller.Linear.Edges, callProbe.ErrEdgeID).ToBlockID, []string{callB.TargetID, callA.TargetID}, callProbe.ErrTargetID, core.OpFail, acquisition.ID) {
		return false
	}
	return v.phase24TerminalCleanup(caller, phase24FindEdge(caller.Linear.Edges, callProbe.OkEdgeID).ToBlockID, []string{callB.TargetID, callA.TargetID}, callProbe.TargetID, core.OpReturn, acquisition.ID)
}

func (v *validator) phase24TerminalCleanup(function *core.Function, blockID string, owners []string, terminalSource string, terminal core.OperationKind, acquisitionID string) bool {
	var block *core.Block
	for index := range function.Linear.Blocks {
		if function.Linear.Blocks[index].ID == blockID {
			block = &function.Linear.Blocks[index]
			break
		}
	}
	if block == nil || len(block.OperationIDs) != len(owners)+1 {
		return v.check(false, "core.owner_transfer_cleanup", blockID)
	}
	operations := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operations[operation.ID] = operation
	}
	for index, owner := range owners {
		release := operations[block.OperationIDs[index]]
		if release.Kind != core.OpRelease || release.SourceID != owner || release.ReleasesOperationID != acquisitionID || release.Foreign == nil || release.Foreign.Mode != "consume" || release.Allocator != "libc_malloc" {
			return v.check(false, "core.owner_transfer_cleanup_order", release.ID)
		}
	}
	last := operations[block.OperationIDs[len(block.OperationIDs)-1]]
	if last.Kind != terminal || terminalSource != "" && last.SourceID != terminalSource {
		return v.check(false, "core.owner_transfer_cleanup", last.ID)
	}
	if terminal == core.OpFail && last.TypeID == "" || terminal == core.OpReturn && last.SourceID == "" {
		return v.check(false, "core.owner_transfer_cleanup", last.ID)
	}
	return true
}

func indexOperation(operations []core.LinearOperation, id string) int {
	for i, op := range operations {
		if op.ID == id {
			return i
		}
	}
	return -1
}

// localOwnerLifecycleValid seeds obligations from every acquire operation,
// then replays borrow and release transitions in source order. It never
// discovers an owner by scanning the release list.
func localOwnerLifecycleValid(function *core.Function) bool {
	if function == nil || function.Linear == nil || function.Parameter.Type != "PathToken" || function.ReturnType != "U64" || len(function.Linear.Blocks) != 0 || len(function.Linear.Edges) != 0 {
		return false
	}
	type ownerState struct {
		placeID  string
		resultID string
		contract *core.ForeignOperationContract
		live     bool
		borrowed bool
		released bool
	}
	owners := map[string]ownerState{}
	activeByPlace := map[string]string{}
	var returnedFrom string
	foreignModes := make([]string, 0, 3)
	for _, operation := range function.Linear.Operations {
		contract := operation.Foreign
		if contract == nil {
			if operation.Kind == core.OpReturn {
				if returnedFrom != "" {
					return false
				}
				returnedFrom = operation.SourceID
			} else {
				return false
			}
			continue
		}
		foreignModes = append(foreignModes, contract.Mode)
		switch contract.Mode {
		case "acquire":
			if operation.Kind != core.OpForeignCall || operation.SourceID != function.Parameter.ID || operation.TargetID == "" || owners[operation.ID].contract != nil || activeByPlace[operation.TargetID] != "" {
				return false
			}
			owners[operation.ID] = ownerState{placeID: operation.TargetID, contract: contract, live: true}
			activeByPlace[operation.TargetID] = operation.ID
		case "borrow":
			acquireID, found := activeByPlace[operation.SourceID]
			owner, known := owners[acquireID]
			if operation.Kind != core.OpForeignCall || !found || !known || !owner.live || owner.borrowed || owner.released || operation.TargetID == "" {
				return false
			}
			owner.borrowed = true
			owner.resultID = operation.TargetID
			owners[acquireID] = owner
		case "consume":
			owner, known := owners[operation.ReleasesOperationID]
			if operation.Kind != core.OpRelease || !known || !owner.live || !owner.borrowed || owner.released || owner.placeID != operation.SourceID || owner.contract.Release != contract.Symbol || owner.contract.Allocator != contract.Allocator {
				return false
			}
			owner.live = false
			owner.released = true
			owners[operation.ReleasesOperationID] = owner
			delete(activeByPlace, operation.SourceID)
		default:
			return false
		}
	}
	if len(owners) != 1 || len(foreignModes) != 3 || foreignModes[0] != "acquire" || foreignModes[1] != "borrow" || foreignModes[2] != "consume" || len(activeByPlace) != 0 {
		return false
	}
	for _, owner := range owners {
		if owner.live || !owner.borrowed || !owner.released || owner.resultID == "" || returnedFrom != owner.resultID {
			return false
		}
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
func (v *validator) blocksAndEdges(function *core.Function, operationIDs map[string]struct{}, types map[string]core.TypeFact, places map[string]core.Place) bool {
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
	blockForOperation := make(map[string]string, len(operationIDs))
	for _, block := range linear.Blocks {
		for _, opID := range block.OperationIDs {
			blockForOperation[opID] = block.ID
		}
	}
	for _, operation := range linear.Operations {
		if operation.Kind != core.OpBranch {
			continue
		}
		var trueEdge, falseEdge *core.Edge
		for index := range linear.Edges {
			edge := &linear.Edges[index]
			if edge.ID == operation.TrueEdgeID {
				trueEdge = edge
			}
			if edge.ID == operation.FalseEdgeID {
				falseEdge = edge
			}
		}
		valid := trueEdge != nil && falseEdge != nil && trueEdge.FromBlockID == blockForOperation[operation.ID] && falseEdge.FromBlockID == blockForOperation[operation.ID] && trueEdge.Pattern == "true" && falseEdge.Pattern == "false"
		if !v.check(valid, "core.scalar_branch_edges_invalid", operation.ID) {
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
	return v.validateScalarFixedPoint(function, types, places)
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

// buildLoanChainIndex builds the parent-pointer chain used by
// carriedLoans. loanCarry is Phase 09's own peer-derived interprocedural
// loan-liveness fact (D-09-03): before this unconditionally sets
// idx.parent[operation.TargetID] = operation.SourceID for an OpCall, it
// first consults loanCarry[operation.CalleeID] -- if the callee's own
// forward-derived fact reports it does NOT return a borrow of its own
// parameter, the call's result is a fresh, owned identity, and the chain
// deliberately breaks at this call boundary (the target's entry is simply
// never written, so carriedLoans can never walk past it into the
// argument's own loan ancestry). Every other operation kind is untouched.
func buildLoanChainIndex(operations []core.LinearOperation, checks *int, loanCarry map[string]peerLoanCarryFact) *loanChainIndex {
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
		if operation.Kind == core.OpCall && !disablePeerLoanCarryConsultForTest && !forcePeerLoanCarryTrueForTest && !loanCarry[operation.CalleeID].ReturnsBorrowOfParam {
			// The callee's own declared contract does not return a borrow
			// of its parameter (or the callee is unknown -- a corrupted
			// artifact fails closed to "does not carry", T-09-02): do not
			// chain this call's target back to its argument at all.
			continue
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

// peerClosureUnmemoizedSeamForTest is Phase 09 Plan 06's own QLT-08
// mutation-kill seam for the peer's cost bound
// (peer_closure_recomputed_work_growth_exponent, D-09-28): when true,
// foldChain (below) skips writing idx.memo entirely, so carriedLoans'
// fast-path memo hit (corevalidate.go:1207) can never trigger and every
// later reference to a place already on some prior walk's path re-walks
// its FULL parent chain from scratch instead of stopping at an
// already-folded prefix. This specifically reproduces a genuinely
// quadratic walk (not merely a slower one) for any shape whose per-
// function chain DEPTH grows with corpus size -- see
// TestPeerClosureCostUnmemoizedSeamExceedsBound's own doc comment for why
// the "forward" corpus shape (corpusForward, a single caller threading one
// borrow through k sequential calls) is the one shape this seam is
// asserted against: querying the i-th call's own source without a memo
// shortcut costs O(i) (walking all the way back to the borrowed root), so
// the SUM over k calls is O(k^2) -- exactly S-006's own "quadratic in body
// length" finding for the unmemoized arm, reproduced here inside the
// shipped validator rather than in a workbench. The read-half seam
// (skipping foldChain's write) was chosen over a carriedLoans-read-skip
// because a write-skip is the minimal change that defeats memoization
// while leaving every other invariant (bornAt, parent, the walk's own
// cycle guard) untouched -- a read-skip would additionally have to
// reimplement "pretend this specific lookup missed" without disturbing
// the walk that follows it. Unexported, false in production; set only via
// SetPeerClosureUnmemoizedSeamForTest (export_test.go) by a test that
// defers the restore immediately.
var peerClosureUnmemoizedSeamForTest bool

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
		if !peerClosureUnmemoizedSeamForTest {
			idx.memo[place] = here
		}
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

	chain := buildLoanChainIndex(linear.Operations, &v.checks, v.peerLoanCarry)

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
	// D-10-52: capture the SAME recomputed value Result.LoanEndpoints later
	// exposes, as it is produced -- never a second, freshly re-derived call.
	if v.peerLoanEndpoints == nil {
		v.peerLoanEndpoints = make(map[string][]core.LoanEndpoint)
	}
	v.peerLoanEndpoints[function.ID] = recomputed
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
	chain := buildLoanChainIndex(operations, &v.checks, v.peerLoanCarry)
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
		if !v.check(operation.Kind == core.OpConst || operation.Kind != core.OpCall || source.TypeID == operation.TypeID || function.Parameter.Type != function.ReturnType, "core.type_mismatch", operation.ID) {
			return false
		}
		if !v.check(operation.Kind == core.OpConst || initialized[operation.SourceID], finalOrTransitionCode(operation.Kind), operation.SourceID) {
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
		case core.OpConst:
			if !scalarTargetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpCopy:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityCopy), "core.ability.copy_denied", operation.TypeID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpAddChecked, core.OpLessU64:
			right, known := places[operation.RightID]
			want := "U64"
			if operation.Kind == core.OpLessU64 {
				want = "Bool"
			}
			if !v.check(known && initialized[operation.RightID] && types[right.TypeID].Shape.Constructor == "U64" && types[operation.TypeID].Shape.Constructor == want, "core.scalar_operation_invalid", operation.ID) {
				return false
			}
			if !scalarTargetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID], produced[operation.TargetID] = true, true
		case core.OpScalarStore:
			target, known := places[operation.StoreTargetID]
			if !v.check(known && target.Mutable && initialized[operation.SourceID] && target.TypeID == source.TypeID && operation.StoreTargetID != function.Parameter.ID, "core.scalar_store_invalid", operation.ID) {
				return false
			}
			initialized[operation.StoreTargetID] = true
		case core.OpBranch:
			if !v.check(types[source.TypeID].Shape.Constructor == "Bool", "core.scalar_branch_invalid", operation.ID) {
				return false
			}
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
			if !v.check(types[source.TypeID].Shape.Constructor == function.ReturnType, core.ReturnTypeMismatch, operation.ID) {
				return false
			}
			if index != len(operations)-1 || returned || operation.TargetID != "" || operation.TypeID != source.TypeID {
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
			if !v.consumeCallArgument(operation, source.TypeID, types, initialized) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if operation.ErrTargetID != "" {
				errTarget, known := places[operation.ErrTargetID]
				if !v.check(known && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
					return false
				}
				initialized[operation.ErrTargetID] = true
				produced[operation.ErrTargetID] = true
			}
		case core.OpForeignCall:
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if operation.Foreign == nil {
				errTarget, errKnown := places[operation.ErrTargetID]
				if !v.check(errKnown && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
					return false
				}
				initialized[operation.ErrTargetID] = true
				produced[operation.ErrTargetID] = true
			}
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
	if !v.check(returned, "core.final_claim_mismatch", function.ID) {
		return false
	}
	hasCall := false
	for _, operation := range operations {
		hasCall = hasCall || operation.Kind == core.OpCall
	}
	for _, operation := range operations {
		if operation.Foreign != nil && function.ReturnType != "FileByteOwner" && !hasCall {
			return v.check(localOwnerLifecycleValid(function), "core.local_owner_lifecycle", function.ID)
		}
	}
	return true
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
	if function.Match != nil {
		for _, arm := range function.Match.Arms {
			if arm.ValuePlaceID != "" {
				initialized[arm.ValuePlaceID] = true // edge-bound value, selected before its arm replay
			}
		}
	}
	produced := map[string]bool{function.Parameter.ID: true}
	loanOwner := make(map[string]string)
	// See replayStraightLine's identical declaration for why a memoized
	// parent-pointer chain (loanChainIndex), not a per-operation copy of the
	// accumulated loan list, computes loanLastUse here (D-02-03/Q2(b)).
	loanLastUse := make(map[string]int)
	chain := buildLoanChainIndex(operations, &v.checks, v.peerLoanCarry)
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
		if !v.check(operation.Kind == core.OpConst || operation.Kind != core.OpCall || source.TypeID == operation.TypeID || function.Parameter.Type != function.ReturnType, "core.type_mismatch", operation.ID) {
			return false
		}
		if !v.check(operation.Kind == core.OpConst || initialized[operation.SourceID], finalOrTransitionCode(operation.Kind), operation.SourceID) {
			return false
		}
		// See replayStraightLine's identical check (D-07-29): CalleeID is
		// kind-exclusive, checked once per operation regardless of kind.
		if !v.check(operation.Kind == core.OpCall || operation.CalleeID == "", "core.callee_id_kind_exclusive", operation.ID) {
			return false
		}
		v.checks++ // dispatch one independently authorized transition
		switch operation.Kind {
		case core.OpConst:
			validTarget := false
			if hasScalarCFG(function) {
				validTarget = scalarTargetMatches(function, index, operation, places, produced)
			} else {
				validTarget = v.targetMatches(function, index, operation, places, produced)
			}
			if !validTarget {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpAddChecked, core.OpLessU64:
			right, known := places[operation.RightID]
			want := "U64"
			if operation.Kind == core.OpLessU64 {
				want = "Bool"
			}
			if !v.check(known && initialized[operation.RightID] && types[source.TypeID].Shape.Constructor == "U64" && types[right.TypeID].Shape.Constructor == "U64" && types[operation.TypeID].Shape.Constructor == want, "core.scalar_operation_invalid", operation.ID) {
				return false
			}
			if !scalarTargetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID], produced[operation.TargetID] = true, true
		case core.OpScalarStore:
			target, known := places[operation.StoreTargetID]
			if !v.check(known && target.Mutable && initialized[operation.SourceID] && target.TypeID == source.TypeID && operation.StoreTargetID != function.Parameter.ID, "core.scalar_store_invalid", operation.ID) {
				return false
			}
			initialized[operation.StoreTargetID] = true
		case core.OpBranch:
			blockID, known := blockOfOperation[operation.ID]
			if !v.check(known && lastOperationOfBlock[blockID] == operation.ID && types[source.TypeID].Shape.Constructor == "Bool", "core.scalar_branch_invalid", operation.ID) {
				return false
			}
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
		case core.OpConstructPayload:
			// D-12-14: construction consumes its own source place (the
			// payload the arm's own binder names), mirroring OpMove's own
			// SourceID clearing -- independently re-derived here, never
			// copy-pasted from check's own emission logic (D-09-02).
			// targetMatches is not reused here: it additionally requires
			// target.TypeID == operation.TypeID, which does not hold for
			// this kind BY DESIGN -- operation.TypeID names the SOURCE
			// payload's type (the universal pre-switch law above), while
			// the constructed TARGET is genuinely, and deliberately, of the
			// outer ADT's own different type.
			expected := fmt.Sprintf("%s:place:%d", function.ID, index+1)
			if !v.check(operation.TargetID == expected && operation.TargetID != operation.SourceID && !produced[operation.TargetID], "core.invalid_target", operation.TargetID) {
				return false
			}
			initialized[operation.SourceID] = false
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
		case core.OpDestructurePayload:
			// D-12-10/D-12-14: this kind names its own produced place via
			// PayloadTargetID (never the ordinary TargetID, which stays
			// empty and is already checked absent by matchBranchStructural),
			// and independently re-derives D-12-29's "dropped exactly once"
			// static accounting by clearing the scrutinee alias's own
			// liveness, mirroring OpMove's SourceID-clearing shape.
			expected := fmt.Sprintf("%s:place:%d", function.ID, index+1)
			target, exists := places[operation.PayloadTargetID]
			if !v.check(exists && operation.PayloadTargetID == expected && operation.PayloadTargetID != operation.SourceID && !produced[operation.PayloadTargetID] && target.TypeID != "", "core.invalid_target", operation.PayloadTargetID) {
				return false
			}
			initialized[operation.SourceID] = false
			initialized[operation.PayloadTargetID] = true
			produced[operation.PayloadTargetID] = true
		case core.OpReturn:
			if !v.check(types[source.TypeID].Shape.Constructor == function.ReturnType, core.ReturnTypeMismatch, operation.ID) {
				return false
			}
			blockID, known := blockOfOperation[operation.ID]
			if !v.check(known && lastOperationOfBlock[blockID] == operation.ID, "core.final_claim_mismatch", operation.ID) {
				return false
			}
			if !v.check(!returnedBlocks[blockID] && operation.TargetID == "" && operation.TypeID == source.TypeID, "core.final_claim_mismatch", operation.ID) {
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
			if !v.consumeCallArgument(operation, source.TypeID, types, initialized) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			if operation.ErrTargetID != "" {
				errTarget, known := places[operation.ErrTargetID]
				if !v.check(known && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
					return false
				}
				initialized[operation.ErrTargetID] = true
				produced[operation.ErrTargetID] = true
			}
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
		// A computed terminal match has a real linear prefix in its entry
		// block. The block then dispatches through the function's existing
		// match edges, so its last prefix operation is intentionally not a
		// function-terminal operation.
		if function.Match != nil && block.ID == function.ID+":block:entry" && len(block.Successors) > 0 {
			continue
		}
		lastOpID := block.OperationIDs[len(block.OperationIDs)-1]
		lastOperation := operationsByID[lastOpID]
		if lastOperation.Kind == core.OpBranch || lastOperation.Kind == core.OpForeignCall || (lastOperation.Kind == core.OpCall && lastOperation.OkEdgeID != "" && lastOperation.ErrEdgeID != "") || (len(block.Successors) > 0 && lastOperation.Kind != core.OpReturn && lastOperation.Kind != core.OpFail && lastOperation.Kind != core.OpDefect) {
			// A fallible foreign call or function call forks into its declared
			// successor edges instead of terminating the function (D-04-04 and
			// Phase 24 typed calls). Only a block whose last operation is an
			// actual terminator (OpReturn/OpFail) must appear in returnedBlocks.
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
	v.peerSignatures[function.ID] = v.derivePeerSignature(function, types, places)
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
// peerConsultObserved is Phase 09 Plan 06's own D-09-30 disclosure-proof
// instrumentation seam, mirroring check.go:641's interproceduralConsultObserved
// exactly: when non-nil, invoked with the function's own ID and the exact
// consulted field name, at every point this derivation reads a
// callee-signature-shaped field. nil in production: zero cost, zero
// allocation. Production visibility of WHICH fields were consulted is
// instead carried on the validator itself (recordPeerConsult, below) and
// surfaced through Result.PeerConsultedFields() -- unlike check's own
// seam, this fact must be readable from an ordinary (non-test) Validate
// call, since PeerConsultedFields is an exported, always-available
// accessor, not a test-only observation.
var peerConsultObserved func(calleeID, field string)

// peerCalleeFrameDrainedObserved is Plan 10-05 Task 2's own disclosure-proof
// instrumentation seam, mirroring peerConsultObserved's identical shape
// (D-09-30): when non-nil, invoked with the function's own ID at every
// derivePeerSignature call, immediately before peerCalleeFrameDrained's own
// predicate runs -- letting a same-package test assert this invariant is
// evaluated EXACTLY ONCE PER DECLARED FUNCTION, independent of call-site
// count (D-10-33's own must_have). nil in production: zero cost, zero
// allocation.
var peerCalleeFrameDrainedObserved func(functionID string)

// parameterContractModeOverrideForTest is plan 10-08 Task 3's own D-10-41
// seeded-fault seam: the COMPLEMENTARY half of the mutant pairing plan
// 10-01 landed for interp's own move-as-copy mutation (moveAsCopyForTest,
// interp.go) -- that seam proves ONLY interp can catch a mutation
// invisible to both static peers; this one proves the inverse direction,
// that criterion 4's differential against corevalidate's own declared
// contract catches a corrupted signature interp is structurally blind to,
// because interp never consults a mode string at all. Production always
// derives parameterContract.Mode as the hardcoded "owned" literal
// (D-07-01: today's grammar has exactly one parameter form, so a natural
// input can never disagree by construction -- D-10-40). When this func is
// non-nil, a same-package test may substitute a wrong value instead.
// Deliberately unexported with no exported Set/Force name in this file:
// package corevalidate_test (or any other importer) cannot reach it
// directly, only this package's own export_test.go bridge can.
var parameterContractModeOverrideForTest func() string

// parameterContractMode reports the effective parameter contract mode:
// parameterContractModeOverrideForTest() when set (test-only), else the
// production "owned" literal.
func parameterContractMode() string {
	if parameterContractModeOverrideForTest != nil {
		return parameterContractModeOverrideForTest()
	}
	return "owned"
}

// recordPeerConsult is the single choke point every consult of a
// callee-signature-shaped field must pass through: it accumulates field
// into v.peerConsultedFields (surfaced production-side via
// Result.PeerConsultedFields()) and, if a same-package test has installed
// peerConsultObserved, also invokes it with the same (calleeID, field)
// pair -- mirroring check's own per-call-site observation shape for
// symmetry, even though this package's own closed-set test only needs the
// flat set.
func (v *validator) recordPeerConsult(calleeID, field string) {
	if v.peerConsultedFields == nil {
		v.peerConsultedFields = make(map[string]bool)
	}
	v.peerConsultedFields[field] = true
	if peerConsultObserved != nil {
		peerConsultObserved(calleeID, field)
	}
}

func (v *validator) derivePeerSignature(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) core.FunctionSignature {
	// places is accepted (not merely function+types) so this signature
	// matches recordSummaryPeer's call sites at both replay sites, which
	// already have it in scope -- no field derived by this peer needs Place
	// facts yet; a future field can gain access without a signature change.
	_ = places

	// Phase 09 Plan 06, Task 3 (D-09-30): this is the peer's own
	// per-function summary-derivation site, structurally mirroring
	// check.go:594-609's own consult of a callee's declared
	// return.mode/parameters[0].mode -- run once per declared function,
	// unconditionally, regardless of whether that function is ever
	// actually called (exactly check's own discipline). check reads these
	// two fields from originvalidate's independently-built interface;
	// corevalidate independently RE-DERIVES the equivalent facts here
	// (returnContract.Mode/parameterContract.Mode, below) rather than
	// reading them from a stored declaration -- but the DISCLOSED FIELD
	// NAME is the same closed vocabulary either way, because both
	// mechanisms answer the identical SEM-05 body-blind question ("what is
	// this function's own declared return/parameter access mode"), just by
	// different means. Recording both here, not only at an OpCall site,
	// mirrors check's own recording point exactly (functionID names the
	// function whose OWN fields these are, matching interproceduralConsultObserved's
	// convention -- later, when some OTHER function calls this one, this
	// recorded consult is what answers "what did the peer read about its
	// callee").
	v.recordPeerConsult(function.ID, "parameters[0].mode")
	v.recordPeerConsult(function.ID, "return.mode")

	// Plan 10-05 Task 2 (D-10-33/D-10-34, SEM-09): promote "the callee's
	// frame is drained of live resources before it pops" to a
	// corevalidate-checked invariant on the callee's own SIGNATURE, checked
	// HERE -- once per function declaration, in the same place as this
	// function's other per-declaration checks -- never re-derived by a
	// caller per call site. See core.CalleeFrameNotDrained's own doc
	// comment for why this is a materially different, coarser fact than
	// checkReleaseOrder's existing path-sensitive reordering check.
	if peerCalleeFrameDrainedObserved != nil {
		peerCalleeFrameDrainedObserved(function.ID)
	}
	v.check(peerCalleeFrameDrained(function), core.CalleeFrameNotDrained, function.ID)

	parameterAbilities := peerParameterAbilities(function, types, places)
	parameterHasDropAbility := abilityGranted(parameterAbilities, core.AbilityDrop)

	parameterContract := core.ParameterContract{
		ID: function.Parameter.ID, Name: function.Parameter.Name, Type: function.Parameter.Type,
		// D-07-01: today's grammar has exactly one parameter form.
		Mode:  parameterContractMode(),
		Drops: parameterHasDropAbility && !peerParameterEscapesOwned(function),
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
	returnAbilities := peerReturnAbilities(function, types)
	returnContract.Fresh = returnContract.Mode == "owned" && abilityGranted(returnAbilities, core.AbilityDrop)

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
		Abilities:  parameterAbilities,
		Callable:   peerCallable(function),
		Fails:      fails,
		Foreign:    foreignReach,
	}
}

// peerParameterAbilities reads only the declared parameter place's own type
// fact. It intentionally does not infer parameter ownership from the return
// fact: the two directions are separate facts once a function may return a
// different type.
func peerParameterAbilities(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) []core.Ability {
	parameter, ok := places[function.Parameter.ID]
	if !ok {
		return nil
	}
	fact, ok := types[parameter.TypeID]
	if !ok || fact.Shape.Constructor != function.Parameter.Type {
		return nil
	}
	return fact.Abilities
}

// peerReturnAbilities reads corevalidate's own return type fact. New
// directional programs use :type:1; the :type:0 fallback preserves the
// established one-fact core schema for older same-type programs. The fault is
// applied here, after parameter derivation, and nowhere else.
func peerReturnAbilities(function *core.Function, types map[string]core.TypeFact) []core.Ability {
	returnFact, ok := types[function.ID+":type:1"]
	if !ok || returnFact.Shape.Constructor != function.ReturnType {
		returnFact, ok = types[function.ID+":type:0"]
	}
	if !ok || returnFact.Shape.Constructor != function.ReturnType || phase17ReturnLookupFaultForTest {
		return nil
	}
	return returnFact.Abilities
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
// the core.origin_omitted refusal class (the PublicOrigin == nil branch),
// by a materially different mechanism (D-07-22) from
// originvalidate.RecomputeOriginPerReturn's backward per-return walk:
// FORWARD set-propagation from the parameter over
// function.Linear.Operations, marking a place borrow-derived when it is the
// target of an OpBorrowShared/OpBorrowExclusive hop from an already-tracked
// place, or of a plain OpMove/OpCopy hop from one. It reports presence only
// (does this function return SOME borrow-derived place), never an access
// mode; it is now a thin wrapper over peerDeriveOriginFacts (D-09-17), which
// Phase 09 extended with the access-mode payload peerOriginContained
// consumes for the declared-origin branch (see peerCallable's doc comment
// and TestPeerRederivesFormerlyNarrowedClasses, which closes D-07-33).
//
// This deliberately does NOT walk through OpForeignCall the way
// originvalidate.checkForeignOriginOmitted/RecomputeOriginPerReturn do for a
// declared borrow/retain foreign contract (D-09-20): that class is decided
// separately, by peerForeignOriginOmitted's own bounded, function-local
// check.
func peerReturnDerivesFromBorrow(function *core.Function) bool {
	return peerDeriveOriginFacts(function).Derived
}

// peerOriginFact is peerDeriveOriginFacts' single-place answer for one
// function: whether its returned place is derived from the parameter
// through a borrow hop, and if so, which access mode the hop that produced
// the RETURNED PLACE ITSELF established (empty when nothing is derived).
type peerOriginFact struct {
	Derived bool
	Access  string
}

// peerDeriveOriginFacts is peerReturnDerivesFromBorrow's access-mode-
// carrying generalization (D-09-17): the SAME single forward pass over
// function.Linear.Operations that function already performs, with its
// `derived map[string]bool` widened to `derived map[string]string`
// carrying the access mode ("shared" or "exclusive") a place was derived
// with. This remains an INDEPENDENT re-derivation by a materially
// different mechanism from originvalidate's RecomputeOriginPerReturn /
// walkReturnOrigin (originvalidate.go:130-218): that is a BACKWARD walk
// from each return, combined across returns by a separate combination law;
// this is a single FORWARD set-propagation pass with no backward walk and
// no combination law across returns at all. Equivalence to
// RecomputeOriginPerReturn's own documented "first-seen-hop-wins" rule (the
// hop NEAREST the return decides) falls out structurally here rather than
// by an explicit backward scan: every borrow op writes its OWN kind into
// its OWN freshly-produced target place (guarded so an already-set place's
// mode is never overwritten -- first-hop-wins per place), so the mode read
// back at the return is always the mode of whichever operation most
// recently PRODUCED the returned place -- exactly the hop nearest the
// return, never an earlier one a later reborrow superseded.
// TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn proves this
// equivalence directly against real fixtures rather than assuming it
// (D-09-19).
//
// peerCallable's peerOriginContained (below) then performs a CONTAINMENT
// check of the DECLARED origin against this fact -- never a recomputation
// of originvalidate's backward combination law -- which is what preserves
// the peer's independence from a third layer (originvalidate) that already
// exists (D-09-19). This walk also deliberately never crosses an
// OpForeignCall hop (D-09-20): the foreign class is decided separately, by
// peerForeignOriginOmitted's own bounded, function-local check.
func peerDeriveOriginFacts(function *core.Function) peerOriginFact {
	if function.Linear == nil {
		return peerOriginFact{}
	}
	// paramTrace tracks places that are the parameter itself, or reach it
	// through a pure Move/Copy chain -- check.go's arm lowering copies the
	// match scrutinee (the parameter) into a fresh place before each arm
	// borrows it, so the borrow hop's SourceID is usually a copy of the
	// parameter, never the parameter's own place ID directly.
	paramTrace := map[string]bool{function.Parameter.ID: true}
	derived := make(map[string]string)
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared:
			if paramTrace[operation.SourceID] || derived[operation.SourceID] != "" {
				if derived[operation.TargetID] == "" {
					derived[operation.TargetID] = "shared"
				}
			}
		case core.OpBorrowExclusive:
			if paramTrace[operation.SourceID] || derived[operation.SourceID] != "" {
				if derived[operation.TargetID] == "" {
					derived[operation.TargetID] = "exclusive"
				}
			}
		case core.OpMove, core.OpCopy:
			if operation.Kind == core.OpCopy && peerIsCopiedU64(function, operation) {
				// A copy of the admitted scalar representation reads an
				// independent U64 value. The loan still has this OpCopy as a
				// use (so access/liveness derivation reaches it), but the copied
				// target no longer carries the parameter's borrow origin.
				continue
			}
			if paramTrace[operation.SourceID] {
				paramTrace[operation.TargetID] = true
			}
			if mode := derived[operation.SourceID]; mode != "" && derived[operation.TargetID] == "" {
				derived[operation.TargetID] = mode
			}
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpReturn {
			continue
		}
		if mode := derived[operation.SourceID]; mode != "" {
			return peerOriginFact{Derived: true, Access: mode}
		}
	}
	return peerOriginFact{}
}

// peerIsCopiedU64 is corevalidate's narrow value-copy boundary for Phase 25.
// U64 is a sealed scalar with Copy ability; copying a value read through a
// borrow ends the origin chain, while the preceding OpCopy remains an access
// to the borrowed place for loan-conflict and liveness checks. Do not infer
// this behavior from a caller's desired family or from a serialized pointer
// ABI claim.
func peerIsCopiedU64(function *core.Function, operation core.LinearOperation) bool {
	if function == nil || function.Linear == nil || operation.Kind != core.OpCopy || operation.TypeID == "" {
		return false
	}
	for _, fact := range function.Linear.Types {
		if fact.ID == operation.TypeID {
			return fact.Shape.Constructor == "U64" && len(fact.Shape.Arguments) == 0 && abilityGranted(fact.Abilities, core.AbilityCopy)
		}
	}
	return false
}

// disablePeerOriginContainmentForTest is Task 1's D-09-19 fault-injection
// seam, the origin-containment counterpart of forcePeerCallableAlwaysTrue:
// when true, peerCallable's declared-origin branch skips
// peerOriginContained entirely and falls back to the old unconditional
// agreement, reproducing D-07-33's pre-Phase-09 false agreement on the
// three formerly-narrowed classes. Unexported, false in production, set
// only via SetDisablePeerOriginContainmentForTest (export_test.go) by a
// same-package test that defers the restore immediately -- QLT-08's proof
// that the containment check is load-bearing, not vacuous
// (TestPeerOriginContainmentDisabledFalselyAgreesAgain).
var disablePeerOriginContainmentForTest bool

// peerOriginContained is peerCallable's independent CONTAINMENT check for a
// function with a DECLARED PublicOrigin (D-09-19): it derives the peer's
// OWN facts via peerDeriveOriginFacts, then checks that the DECLARATION
// covers them, rather than recomputing originvalidate's backward
// combination law -- the distinction that preserves independence from a
// third layer (originvalidate) that already exists. It reaches the same
// VERDICT as originvalidate.PublishProblemsFor's three declared-origin
// checks (domain, path containment, access agreement) without copying
// their exact order, since the peer is under no obligation to reach the
// verdict the same way.
func peerOriginContained(function *core.Function) bool {
	declared := function.PublicOrigin
	if declared.Access != "shared" && declared.Access != "exclusive" {
		// Domain check: a declared Access outside the closed set can never
		// be matched by any derivation, so a mutated summary cannot declare
		// a conflicting value and have it pass by accident.
		return false
	}
	fact := peerDeriveOriginFacts(function)
	if !fact.Derived {
		return false
	}
	// Under arity 1, PublicOrigin.Paths has exactly one possible legal
	// value (function.Parameter.Name) whenever anything is derivable at
	// all (D-09-18), so path containment degenerates to a membership test
	// -- written as containment (a loop over declared.Paths) rather than a
	// single equality, so widening arity past 1 (D-07-07) does not
	// silently change this function's semantics.
	covered := false
	for _, path := range declared.Paths {
		if path == function.Parameter.Name {
			covered = true
			break
		}
	}
	if !covered {
		return false
	}
	return declared.Access == fact.Access
}

// disablePeerForeignOriginPeerForTest is Task 2's D-09-20 fault-injection
// seam: when true, peerCallable skips peerForeignOriginOmitted entirely.
// Unexported, false in production, set only via
// SetDisableForeignOriginPeerForTest (export_test.go) by a same-package
// test that defers the restore immediately -- QLT-08's proof that the
// foreign class is load-bearing, not vacuous
// (TestPeerOriginContainmentDisabledFalselyAgreesAgain's foreign-class
// half).
var disablePeerForeignOriginPeerForTest bool

// peerForeignOriginOmitted is peerCallable's independent, BOUNDED
// re-derivation of the core.foreign_origin_omitted class (D-09-20): it
// consults ONLY this one function's own function.ForeignContract, never
// originvalidate's checkForeignOriginOmitted, never a whole-program
// callee-contract table. originvalidate's own checkForeignOriginOmitted
// (originvalidate.go:282-345) is already a ~55-line backward walk scoped to
// one function's own ForeignContract.Alias -- proof by existence that a
// narrow, function-local foreign-crossing check is possible without
// depending on spike 005's PARTIAL FFI-provenance surface. This function
// performs the SAME bounded check with the peer's own FORWARD propagation
// shape instead: seeded from the foreign call's own target place,
// propagated through OpMove/OpCopy/OpBorrowShared/OpBorrowExclusive hops,
// read at OpReturn -- never a backward sourceOf walk, and never by
// importing or calling originvalidate. Crossing the OpForeignCall hop HERE
// does not reopen the general-propagation prohibition
// peerDeriveOriginFacts observes: this check is function-local, bounded to
// one function's own declared contract, not a general cross-function walk.
func peerForeignOriginOmitted(function *core.Function) bool {
	if function.ForeignContract == nil {
		return false
	}
	alias := function.ForeignContract.Alias
	if alias != "borrow" && alias != "retain" {
		return false
	}
	if function.Linear == nil || function.PublicOrigin != nil {
		return false
	}
	fromForeign := make(map[string]bool)
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpForeignCall {
			fromForeign[operation.TargetID] = true
			continue
		}
		switch operation.Kind {
		case core.OpMove, core.OpCopy, core.OpBorrowShared, core.OpBorrowExclusive:
			if fromForeign[operation.SourceID] {
				fromForeign[operation.TargetID] = true
			}
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpReturn && fromForeign[operation.SourceID] {
			return true
		}
	}
	return false
}

// peerCallable is corevalidate's independent re-derivation of Callable
// (D-04-03's full publication-safety predicate). D-07-33 originally
// narrowed this to the core.origin_omitted class alone, leaving the other
// three of PublishProblemsFor's four refusal classes as a provable FALSE
// agreement; Phase 09 closes that narrowing (D-09-16: the literal OWN-08
// reading that stops at origin_omitted is rejected -- Callable is defined
// project-wide as the full predicate). It now independently re-derives all
// four classes:
//  1. Foreign-origin-omitted (peerForeignOriginOmitted, D-09-20), consulted
//     first, matching PublishProblemsFor's own early return.
//  2. core.origin_omitted, when no public origin is declared: Callable is
//     false iff the body itself returns a borrow-derived place
//     (peerReturnDerivesFromBorrow).
//     3 and 4. Path containment (core.origin_understated) and access
//     agreement (core.origin_access_mismatch), when a public origin IS
//     declared: peerOriginContained compares the declaration against the
//     peer's own forward-derived facts (D-09-19) -- a containment check,
//     never a recomputation of originvalidate's backward combination law.
//
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
// WR-01 (07-REVIEW): that equivalence is NOT an IR invariant local to
// OpCall the way it is for OpCopy/OpMove/OpBorrow* (where TargetID's type
// and SourceID's type are definitionally the same value) -- for OpCall,
// operation.TypeID is the call's declared RETURN type (derived on the
// producer side from the callee's declared return type), and it only
// equals the argument's own SourceID type because of two OTHER,
// separately-checked facts: the generic source.TypeID == operation.TypeID
// law above (which is not OpCall-specific and could in principle be
// loosened or bypassed for one kind without this function noticing), and
// D-07-09's constraint that a function's declared return type equals its
// own declared parameter type. sourceTypeID is threaded through
// explicitly by this function's two callers (each already holding
// places[operation.SourceID].TypeID as `source.TypeID`) so this function
// can assert the equivalence itself, right where the derivation happens,
// instead of trusting a caller-side law it cannot see -- if either
// constraint above is ever weakened, this assertion fails closed here
// rather than silently deriving copyability from the wrong type.
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
func (v *validator) consumeCallArgument(operation core.LinearOperation, sourceTypeID string, types map[string]core.TypeFact, initialized map[string]bool) bool {
	if disableCallArgumentConsumePeerForTest {
		return true
	}
	// A call must name a target when its return type differs from its argument
	// type. This also keeps the direct, target-less test seam fail-closed.
	if operation.TargetID == "" && operation.TypeID != sourceTypeID {
		v.problems = append(v.problems, Problem{Code: "core.type_mismatch", Detail: operation.ID})
		return false
	}
	// WR-01 (07-REVIEW): OpCall-specific defense -- operation.TypeID (the
	// call's declared return type) must equal sourceTypeID (the argument
	// place's own declared type) before the ability derivation below is
	// allowed to key off operation.TypeID. This is NOT redundant with the
	// generic source.TypeID == operation.TypeID check every operation kind
	// already passes (corevalidate.go, replayStraightLine/replayBlocks):
	// that check is generic across all kinds and could be loosened or
	// bypassed for OpCall specifically without this function ever knowing.
	// A divergence here ties directly to D-07-09 (a function's declared
	// return type must equal its own declared parameter type) -- if that
	// constraint is ever relaxed, this assertion is what catches it.
	granted, _, known := v.derive(types[sourceTypeID].Shape, 0)
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

// peerCalleeFrameDrained is Plan 10-05 Task 2's per-function-declaration
// predicate (D-10-33/D-10-34, core.CalleeFrameNotDrained's own doc comment
// has the full rationale). Every OpForeignCall acquisition this function's
// own body declares is exempt from the drain obligation entirely when its
// own OkEdgeID/ErrEdgeID resolve to the SAME ToBlockID -- checkResourceLifecycle's
// own structural encoding (D-04-06) of `discard ... because`: converging
// ok/err edges are check's own declaration that this acquisition's outcome,
// and by extension its resource, is deliberately advisory and untracked,
// mirroring checkReleaseOrder's own identical "only a call whose ok and err
// edges target DIFFERENT blocks is tracked" convention exactly (never a
// SEPARATE signal this peer invented). Every other (tracked) acquisition
// must be either (a) released by SOME OpRelease anywhere in the function,
// or (b) ownership-transferred OUT via a terminating OpReturn --
// foreign_acquire_one.schway's own `handle` shape, where the acquired value
// is returned directly and needs no release at all, since the resource
// becomes the CALLER's obligation the instant it crosses the return
// boundary. (b) is traced forward through a pure Move/Copy chain, seeded
// from the acquisition's own TARGET place, mirroring
// peerParameterEscapesOwned's own established forward set-propagation shape
// exactly (D-07-22) -- just seeded from an acquisition's target instead of
// the function's own parameter. A tracked acquisition satisfying NEITHER
// (a) nor (b) is reported false: it is genuinely abandoned, live in this
// frame at every one of its own terminating returns.
//
// A function with no Linear body at all (a pure lang.core/0 match function)
// trivially drains (it acquires nothing), and a straight-line body with no
// OpForeignCall at all is likewise trivially true, matching every
// pre-Phase-4 function's own admitted shape (D-04-23: this invariant
// changes nothing for a program that never acquires a foreign resource).
// This is deliberately DIFFERENT from checkReleaseOrder's own path-
// sensitive backward-walk reordering (which only orders releases that
// ALREADY exist somewhere in the function): the two checks are independent
// and neither subsumes the other.
func peerCalleeFrameDrained(function *core.Function) bool {
	if function.Linear == nil {
		return true
	}
	linear := function.Linear

	edgesByID := make(map[string]core.Edge, len(linear.Edges))
	for _, edge := range linear.Edges {
		edgesByID[edge.ID] = edge
	}

	released := make(map[string]bool, len(linear.Operations))
	returnedSources := make(map[string]bool)
	for _, operation := range linear.Operations {
		switch operation.Kind {
		case core.OpRelease:
			if operation.ReleasesOperationID != "" {
				released[operation.ReleasesOperationID] = true
			}
		case core.OpReturn:
			returnedSources[operation.SourceID] = true
		}
	}

	for _, acquisition := range linear.Operations {
		if acquisition.Kind != core.OpForeignCall {
			continue
		}
		// A checked borrowed result is an ordinary scalar observation, and
		// a consuming foreign operation discharges an argument rather than
		// acquiring a frame-owned resource. Only explicit contracts authorize
		// these exemptions: legacy core with a nil Foreign contract and
		// unknown modes remain conservatively tracked as acquisitions.
		if acquisition.Foreign != nil && (acquisition.Foreign.Mode == "borrow" || acquisition.Foreign.Mode == "consume") {
			continue
		}
		if okEdge, okKnown := edgesByID[acquisition.OkEdgeID]; okKnown {
			if errEdge, errKnown := edgesByID[acquisition.ErrEdgeID]; errKnown && okEdge.ToBlockID == errEdge.ToBlockID {
				continue
			}
		}
		if released[acquisition.ID] {
			continue
		}
		reaches := map[string]bool{acquisition.TargetID: true}
		for _, operation := range linear.Operations {
			if operation.Kind != core.OpMove && operation.Kind != core.OpCopy {
				continue
			}
			if reaches[operation.SourceID] {
				reaches[operation.TargetID] = true
			}
		}
		escapesViaReturn := false
		for source := range returnedSources {
			if reaches[source] {
				escapesViaReturn = true
				break
			}
		}
		if !escapesViaReturn {
			return false
		}
	}
	return true
}

func peerCallable(function *core.Function) bool {
	if forcePeerCallableAlwaysTrue {
		return true
	}
	if !disablePeerForeignOriginPeerForTest && peerForeignOriginOmitted(function) {
		return false
	}
	if function.PublicOrigin == nil {
		return !peerReturnDerivesFromBorrow(function)
	}
	return disablePeerOriginContainmentForTest || peerOriginContained(function)
}

type releaseWitness struct {
	acquisitionID string
	ownerPlaceID  string
	checkSource   bool
}

// returnedResourceAcquisition finds a direct foreign acquisition whose
// returned owner crosses a declared function boundary. The caller's OpCall
// becomes the dynamic owner event; the callee acquisition operation remains
// the stable release reference.
func returnedResourceAcquisition(function *core.Function) (core.LinearOperation, bool) {
	if function == nil || function.Linear == nil {
		return core.LinearOperation{}, false
	}
	var acquisition core.LinearOperation
	acquisitionCount, returnCount := 0, 0
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpForeignCall && operation.Foreign != nil && operation.Foreign.Mode == "acquire" {
			acquisition = operation
			acquisitionCount++
		}
		if operation.Kind == core.OpReturn {
			returnCount++
			if operation.SourceID != acquisition.TargetID || operation.TypeID != acquisition.TypeID {
				return core.LinearOperation{}, false
			}
		}
	}
	return acquisition, acquisitionCount == 1 && returnCount == 1 && acquisition.TargetID != ""
}

// checkReleaseOrder independently rederives the reverse-order release
// sequence D-04-07 requires. For every block ending in OpFail or OpReturn, it
// walks BACKWARD over the declared block/edge graph starting from that
// block's own incoming edge, collecting every completed (not-yet-discharged)
// successful acquisition it passes -- a direct foreign acquire or an owning
// OpCall whose callee returns that acquisition. Discovery order is already
// reverse-of-completion order because the walk moves from the fail/return
// point back toward the entry. A failure block's own triggering acquisition
// is excluded (its incoming edge has Pattern "err", so includeThis starts
// false); a success block's immediate predecessor IS included (its incoming
// edge has Pattern "ok", so it completed). This is materially different from
// check.go's forward accumulation (D-12/D-12a): it never reads check's own
// accumulated list or any field check uses to communicate it, only the
// block/edge graph, the operations check.go emitted, and the callee's own
// acquisition-and-return operations. Only a call whose ok and err edges
// target DIFFERENT blocks is tracked -- a `discard`'s converging ok/err edges
// mark its resource as untracked for release this plan, a documented
// narrowing shared with check.go's own accumulation.
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
	functionByID := make(map[string]*core.Function, len(v.program.Functions))
	for index := range v.program.Functions {
		functionByID[v.program.Functions[index].ID] = &v.program.Functions[index]
	}
	callInBlock := make(map[string][]releaseWitness, len(linear.Blocks))
	for _, block := range linear.Blocks {
		// A block can contain more than one successful acquisition. Record
		// them in reverse operation order because cleanup discharges the most
		// recently completed owner first.
		for index := len(block.OperationIDs) - 1; index >= 0; index-- {
			operation, ok := operationsByID[block.OperationIDs[index]]
			if !ok {
				continue
			}
			switch {
			case operation.Kind == core.OpForeignCall && tracked[operation.ID]:
				callInBlock[block.ID] = append(callInBlock[block.ID], releaseWitness{acquisitionID: operation.ID})
			case operation.Kind == core.OpCall && operation.TargetID != "":
				callee := functionByID[operation.CalleeID]
				acquisition, returnsOwner := returnedResourceAcquisition(callee)
				if returnsOwner && tracked[acquisition.ID] {
					callInBlock[block.ID] = append(callInBlock[block.ID], releaseWitness{acquisitionID: acquisition.ID, ownerPlaceID: operation.TargetID, checkSource: true})
				}
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
	var rederive func(startEdge core.Edge, visited map[string]bool) ([]releaseWitness, bool)
	rederive = func(startEdge core.Edge, visited map[string]bool) ([]releaseWitness, bool) {
		var expected []releaseWitness
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
				expected = append(expected, callInBlock[currentBlockID]...)
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
			var agreed []releaseWitness
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
				sourceMatches := !want.checkSource || actual[index].SourceID == want.ownerPlaceID
				if !v.check(actual[index].ReleasesOperationID == want.acquisitionID && sourceMatches, "core.release_order_mismatch", actual[index].ID) {
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
func sameReleaseHistory(left, right []releaseWitness) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
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

func scalarTargetMatches(function *core.Function, operationIndex int, operation core.LinearOperation, places map[string]core.Place, produced map[string]bool) bool {
	target := places[operation.TargetID]
	return target.ID != "" && operation.TargetID != operation.SourceID && !produced[operation.TargetID] && target.TypeID == operation.TypeID
}

func hasScalarCFG(function *core.Function) bool {
	if function == nil || function.Linear == nil {
		return false
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpAddChecked || operation.Kind == core.OpLessU64 || operation.Kind == core.OpScalarStore || operation.Kind == core.OpBranch {
			return true
		}
	}
	return false
}

type scalarAbstractValue struct {
	initialized bool
	known       bool
	isBool      bool
	u64         uint64
	boolean     bool
}

func (v *validator) validateScalarFixedPoint(function *core.Function, types map[string]core.TypeFact, places map[string]core.Place) bool {
	if !hasScalarCFG(function) {
		return true
	}
	linear := function.Linear
	for _, place := range linear.Places {
		constructor := types[place.TypeID].Shape.Constructor
		if !v.check(constructor == "U64" || constructor == "Bool", "core.scalar_backedge_authority", place.ID) {
			return false
		}
	}
	blocks := append([]core.Block(nil), linear.Blocks...)
	sort.Slice(blocks, func(i, j int) bool { return blocks[i].ID < blocks[j].ID })
	operations := make(map[string]core.LinearOperation, len(linear.Operations))
	for _, operation := range linear.Operations {
		operations[operation.ID] = operation
	}
	out := make(map[string]map[string]scalarAbstractValue, len(blocks))
	reachable := make(map[string]bool, len(blocks))
	const transferLimit = 65536
	transfers := 0
	for sweep := 0; ; sweep++ {
		changed := false
		for _, block := range blocks {
			incoming := map[string]scalarAbstractValue{}
			knownIncoming := false
			if block.PointID == function.EntryPointID {
				incoming[function.Parameter.ID] = scalarAbstractValue{initialized: true, isBool: false}
				knownIncoming = true
			}
			for _, edge := range linear.Edges {
				if edge.ToBlockID != block.ID || !reachable[edge.FromBlockID] {
					continue
				}
				if !knownIncoming {
					incoming = cloneScalarState(out[edge.FromBlockID])
					knownIncoming = true
					continue
				}
				incoming = mergeScalarState(incoming, out[edge.FromBlockID], places)
			}
			if !knownIncoming {
				continue
			}
			state := cloneScalarState(incoming)
			for _, operationID := range block.OperationIDs {
				transfers++
				if transfers > transferLimit {
					return v.check(false, "core.scalar_analysis_limit", function.ID)
				}
				operation := operations[operationID]
				source := state[operation.SourceID]
				switch operation.Kind {
				case core.OpConst:
					value, err := strconv.ParseUint(operation.ConstU64, 10, 64)
					if err != nil {
						return v.check(false, "core.scalar_constant_invalid", operation.ID)
					}
					state[operation.TargetID] = scalarAbstractValue{initialized: true, known: true, u64: value}
				case core.OpAddChecked:
					right := state[operation.RightID]
					if !v.check(source.initialized && right.initialized, "core.place_uninitialized", operation.ID) {
						return false
					}
					value := scalarAbstractValue{initialized: true}
					if source.known && right.known && ^uint64(0)-source.u64 >= right.u64 {
						value.known, value.u64 = true, source.u64+right.u64
					}
					state[operation.TargetID] = value
				case core.OpLessU64:
					right := state[operation.RightID]
					if !v.check(source.initialized && right.initialized, "core.place_uninitialized", operation.ID) {
						return false
					}
					value := scalarAbstractValue{initialized: true, isBool: true}
					if source.known && right.known {
						value.known, value.boolean = true, source.u64 < right.u64
					}
					state[operation.TargetID] = value
				case core.OpScalarStore:
					if !v.check(source.initialized, "core.place_uninitialized", operation.ID) {
						return false
					}
					state[operation.StoreTargetID] = scalarAbstractValue{initialized: true, known: source.known, isBool: types[places[operation.StoreTargetID].TypeID].Shape.Constructor == "Bool", u64: source.u64, boolean: source.boolean}
				case core.OpBranch, core.OpReturn, core.OpDefect:
					if !v.check(source.initialized, "core.place_uninitialized", operation.ID) {
						return false
					}
				default:
					return v.check(false, "core.scalar_backedge_authority", operation.ID)
				}
			}
			if !reachable[block.ID] || !sameScalarState(out[block.ID], state) {
				out[block.ID] = state
				reachable[block.ID] = true
				changed = true
			}
		}
		if !changed {
			return true
		}
		if sweep >= transferLimit {
			return v.check(false, "core.scalar_analysis_limit", function.ID)
		}
	}
}

func cloneScalarState(state map[string]scalarAbstractValue) map[string]scalarAbstractValue {
	clone := make(map[string]scalarAbstractValue, len(state))
	for id, value := range state {
		clone[id] = value
	}
	return clone
}

func mergeScalarState(left, right map[string]scalarAbstractValue, places map[string]core.Place) map[string]scalarAbstractValue {
	merged := make(map[string]scalarAbstractValue, len(places))
	for id := range places {
		a, b := left[id], right[id]
		value := scalarAbstractValue{initialized: a.initialized && b.initialized, isBool: a.isBool && b.isBool}
		if value.initialized && a.known && b.known && a.isBool == b.isBool && a.u64 == b.u64 && a.boolean == b.boolean {
			value.known, value.u64, value.boolean = true, a.u64, a.boolean
		}
		merged[id] = value
	}
	return merged
}

func sameScalarState(left, right map[string]scalarAbstractValue) bool {
	if len(left) != len(right) {
		return false
	}
	for id, value := range left {
		if right[id] != value {
			return false
		}
	}
	return true
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
	case "Byte", "U64", "Bool":
		if len(shape.Arguments) != 0 {
			return false, nil, false
		}
		return true, nil, true
	case "PathToken", "Unit":
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
	case "FileByteOwner":
		if len(shape.Arguments) != 0 {
			return false, nil, false
		}
		if requested == core.AbilityCopy {
			return false, []string{"FileByteOwner"}, true
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

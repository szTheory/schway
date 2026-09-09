package session

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// Phase7RequiredControls is the complete Phase 07 required-control list as
// of plan 07-04 (D-07-41), mirroring Phase4RequiredControls' /
// Phase5RequiredControls' shape verbatim: a flat slice of exact identifier
// strings, one per line, no computation. scripts/verify-phase7.sh
// duplicates this list VERBATIM, and TestPhase7RequiredControlsMatchScript
// asserts the two are set-equal in both directions, so the gate cannot
// silently shrink by dropping a control from either copy (T-07-24).
//
// All three identifiers are fired together by this file's own
// lane:kind-exhaustive-dispatch-phase07 (VerifyPhase7ControlsAndWork): the
// CLI-observable lane drives the SAME phase-07 fixtures and the SAME
// recognized-not-executed assertion core_test.go's in-process control
// (control:kind.exhaustive_dispatch.phase07_in_process) and Task 3's
// mutation-kill test (control:dispatch.recognized_not_executed) exercise
// in-process -- so a CLI run of this lane genuinely proves all three
// claims, not merely the lane's own identity
// (control:kind.exhaustive_dispatch.phase07_lane). Recording all three
// here, rather than only the lane's own identifier, is what lets 07-07's
// phase-wide completeness matrix compare against a single authoritative
// list without a second, divergent registry.
func Phase7RequiredControls() []string {
	return []string{
		ControlKindExhaustiveDispatchPhase07InProcess,
		ControlKindExhaustiveDispatchPhase07Lane,
		ControlDispatchRecognizedNotExecuted,
		ControlCallAdmissionBodyBlind,
		ControlCallGraphGrayReentry,
		ControlCallGraphSelfEdge,
		ControlCallGraphUnresolvedEdgeRefused,
		ControlCallGraphCycleIDDeterministic,
		ControlCallGraphCauseBound,
		ControlCallCallableRefusal,
		ControlCorevalidateCyclePeerIndependent,
		ControlCorevalidateCyclePeerGrayReentry,
		ControlCallGraphForeignShadowingEdgePreserved,
		ControlPhase07ControlsAreMutationKilled,
		ControlSummaryClosureDigestChained,
		ControlSummaryClosureDigestReversePostorder,
		ControlCallArgumentTypeMatchesParameter,
		ControlCallTargetTypeFromCalleeReturn,
		ControlCheckPeerConsulted,
		ControlInterfacePeerRefusalIsInvalid,
		ControlCallArgumentConsumedWhenNoncopyable,
		ControlCallCopyableArgumentNotConsumed,
		ControlSummaryForeignReachClosureDerived,
		ControlSummaryFailsClosureDerived,
	}
}

// Control identifiers introduced by plan 07-04 (D-07-41), following the
// session-package's control:<area>.<specific> naming convention.
const (
	// ControlKindExhaustiveDispatchPhase07InProcess names
	// internal/compiler/core/core_test.go's
	// TestAllOperationKindsHandledAtEverySite fixture list once it carries
	// core.OpCall: the in-process control (A-05) exercised via `go test`,
	// never repurposing Phase 4's own control:kind.exhaustive_dispatch
	// identifier or fixture list (T-07-23).
	ControlKindExhaustiveDispatchPhase07InProcess = "control:kind.exhaustive_dispatch.phase07_in_process"
	// ControlKindExhaustiveDispatchPhase07Lane names this file's own
	// CLI-observable lane (`lang verify testdata/phase07`), the phase-07
	// sibling of Phase 4's lane:kind-exhaustive-dispatch, with its own
	// distinct lane ID and its own phase07DispatchFixtures list.
	ControlKindExhaustiveDispatchPhase07Lane = "control:kind.exhaustive_dispatch.phase07_lane"
	// ControlDispatchRecognizedNotExecuted names the D-07-39 claim both
	// controls assert: a function containing core.OpCall is RECOGNIZED at
	// the interp site (interp.Run returns the named interp.ErrCallUnsupported,
	// never a crash and never a silently faked success) but Phase 07 defines
	// no call-stack execution semantics for it. Task 3's mutation-kill test
	// (TestPhase7DispatchControlsMutationKilled) proves this assertion is
	// load-bearing: folding core.OpCall into a grouped copy/move/borrow arm
	// makes it go red.
	ControlDispatchRecognizedNotExecuted = "control:dispatch.recognized_not_executed"
	// ControlCallAdmissionBodyBlind names 07-05 Task 3's SEM-05 structural
	// claim, made falsifiable rather than asserted: check's "call"
	// admission arm (verifyCallableRefusal) never reaches a callee's own
	// core.Function.Linear/Match -- only the pre-body signature table
	// (D-07-34). Killed by check_test.go's own
	// TestCallAdmissionBodyBlindControl (the body-read seam) -- an
	// in-process Go-test proof, same category as
	// ControlDispatchRecognizedNotExecuted above. This lane's own two
	// fixtures (both Callable) exercise the SAME production admission arm
	// through a real CLI path with zero divergence, but neither fixture
	// contains a REFUSED call -- declared here, never implied, exactly
	// like A-02's cgen declaration above.
	ControlCallAdmissionBodyBlind = "control:call.admission_body_blind"
	// ControlCallCallableRefusal names 07-05 Task 3's SEM-06 two-peer
	// discipline: a call to a non-publishable callee (Callable == false,
	// D-04-03) is refused independently at BOTH check (the signature
	// table's Callable consult) and corevalidate (peerCallable, D-07-33's
	// narrowed re-derivation), with the bilateral case failing the gate
	// rather than passing it. Killed by check_test.go's
	// TestVerifyCallableRefusalSeamAdmitsUncallableCallee and
	// corevalidate_mutation_matrix_test.go's fault6/fault7 subtests --
	// in-process Go-test proofs. Declared, not implied: this lane's own
	// fixtures contain no refused call (both are Callable), so the lane
	// exercises the CONSULTING mechanism (with a permitting verdict), not
	// the refusing branch.
	ControlCallCallableRefusal = "control:call.callable_refusal"
	// ControlCallGraphGrayReentry names 07-06 Task 3's headline mutation
	// kill (QLT-08): the on-stack (gray) re-entry distinguisher, never a
	// plain visited set, is what lets callgraph.Order accept a
	// diamond-laden acyclic graph while still refusing a real cycle.
	// Killed by callgraph_test.go's own TestCallGraphMutationMatrix
	// (gray_reentry subtest), which observes
	// testdata/phase07/deep_diamond_acyclic.lang wrongly refused the
	// instant the gray-versus-visited seam is engaged -- an in-process
	// Go-test proof, same category as ControlCallAdmissionBodyBlind
	// above. This lane's own dispatch fixtures now include
	// deep_diamond_acyclic.lang (it accepts); cycle_self.lang is never
	// added to a fixture list that expects a clean check.
	ControlCallGraphGrayReentry = "control:callgraph.gray_reentry"
	// ControlCallGraphSelfEdge names 07-06 Task 3's self-recursion control:
	// excluding a callee == caller edge would silently legalize direct
	// recursion. Killed by TestCallGraphMutationMatrix's self_edge
	// subtest, which observes testdata/phase07/cycle_self.lang wrongly
	// accepted the instant that seam is engaged.
	ControlCallGraphSelfEdge = "control:callgraph.self_edge"
	// ControlCallGraphUnresolvedEdgeRefused names 07-06 Task 3's
	// dropped-edge control (D-07-45): a CalleeID resolving to no declared
	// function must refuse, never be silently dropped from the graph --
	// a dropped edge is how a cycle escapes detection. Killed by
	// TestCallGraphMutationMatrix's unresolved_edge_refused subtest.
	ControlCallGraphUnresolvedEdgeRefused = "control:callgraph.unresolved_edge_refused"
	// ControlCallGraphCycleIDDeterministic names 07-06 Task 2's
	// determinism controls (D-07-16/D-07-43): canonical rotation and
	// cross-cycle lexicographically-smallest witness selection, each
	// independently load-bearing. Killed by callgraph_test.go's
	// TestRotationRemovalMutationKilled and
	// TestWitnessSelectionRemovalMutationKilled.
	ControlCallGraphCycleIDDeterministic = "control:callgraph.cycle_id_deterministic"
	// ControlCallGraphCauseBound names 07-06 Task 2's D-07-15 bound: the
	// core.call_graph_cycle diagnostic caps emitted cycle_member causes at
	// callgraph.MaxCycleCauses (32), reporting
	// callgraph.TruncatedCycleBound on overflow, while the traversal
	// itself stays unbounded. Verified by check_test.go's
	// TestCallGraphCycleBoundedAt32Causes and
	// TestCallGraphCycleTruncatesAt33Members.
	ControlCallGraphCauseBound = "control:callgraph.cause_bound"
	// ControlCorevalidateCyclePeerIndependent names 07-07 Task 1/Task 3's
	// D-07-19 claim: corevalidate's own, independently written cycle
	// traversal (over a disjoint, synthetic-only reachable input space)
	// refuses a cycle even with check's own callgraph-based refusal
	// disabled. Killed by
	// check_test.go's TestCheckCycleRefusalIndependentOfCorevalidatePeer.
	ControlCorevalidateCyclePeerIndependent = "control:corevalidate.cycle_peer_independent"
	// ControlCorevalidateCyclePeerGrayReentry names 07-07 Task 3 Test 4's
	// own gray-versus-visited seam on corevalidate's independent
	// traversal -- a second derivation nobody has seen fail is not
	// evidence (T-07-41). Killed by
	// corevalidate_cycle_peer_test.go's TestCyclePeerMutationMatrix.
	ControlCorevalidateCyclePeerGrayReentry = "control:corevalidate.cycle_peer_gray_reentry"
	// ControlCallGraphForeignShadowingEdgePreserved names 07-07 Task 2/
	// Task 3's T-07-43 claim: a declared Lang function whose name shadows
	// a foreign symbol still contributes its own CalleeID graph edge
	// (D-07-30's resolution-time precedence), never silently dropped as a
	// foreign call would be. Killed by check_test.go's
	// TestForeignSymbolShadowingFixtureRefusedAndEdgeMutationKilled.
	ControlCallGraphForeignShadowingEdgePreserved = "control:callgraph.foreign_shadowing_edge_preserved"
	// ControlPhase07ControlsAreMutationKilled names 07-07 Task 3 Test 5's
	// own completeness claim (D-07-41): every control in
	// Phase7RequiredControls() has a recorded seeded-mutation kill,
	// compared by exact set equality against an authoritative list, never
	// a self-authored registry. Killed by session_phase7_test.go's own
	// TestPhase7ControlsAreMutationKilledMetaMutation.
	ControlPhase07ControlsAreMutationKilled = "control:phase07.controls_are_mutation_killed"
	// ControlSummaryClosureDigestChained names 07-08 Task 1/Task 2's
	// D-07-12/D-07-38 claim: ClosureDigest is a real Merkle chain over
	// callee SUMMARY digests (never a callee body), and changing a
	// callee's own published signature changes its caller's ClosureDigest
	// -- the only thing that closes the stale-but-self-consistent hole a
	// per-unit hash leaves. Killed by originvalidate's
	// TestClosureDigestEmptyCalleesMutationKilled (producer side) and
	// corevalidate's TestPeerClosureDigestEmptyCalleesMutationKilled (peer
	// side).
	ControlSummaryClosureDigestChained = "control:summary.closure_digest_chained"
	// ControlSummaryClosureDigestReversePostorder names 07-08's D-07-38
	// ordering claim: the digest chain is computed only over a graph
	// already proven acyclic, in an order where every callee is finished
	// before its caller (callgraph.Order's own reverse postorder, walked
	// backward). Killed by originvalidate's
	// TestClosureDigestDiscoveryOrderMutationKilled (producer side) and
	// corevalidate's TestPeerClosureDigestDiscoveryOrderMutationKilled
	// (peer side).
	ControlSummaryClosureDigestReversePostorder = "control:summary.closure_digest_reverse_postorder"
	// ControlCallArgumentTypeMatchesParameter names 07-09's argument-type
	// admission gate (D-07-09/SEM-05): a call's argument type must equal
	// the callee's declared parameter type, checked independently at both
	// admission layers -- check.resolveCallBinding (from a pre-body
	// AST-derived callee-contract table) and corevalidate's OpCall replay
	// (from this program's own places/type-facts and functionByID, with
	// no helper shared with check). Killed on the check side by
	// check_test.go's TestCallArgumentTypeCheckMutationKilled and on the
	// peer side by corevalidate_test.go's TestCallTypePeerMutationMatrix.
	// This lane's own fixtures (call_basic.lang,
	// call_from_both_match_arms.lang, deep_diamond_acyclic.lang) all
	// ADMIT -- declared, not implied: the lane exercises the CONSULTING
	// mechanism with a permitting verdict, never the refusing branch,
	// exactly like ControlCallAdmissionBodyBlind and
	// ControlCallCallableRefusal above.
	ControlCallArgumentTypeMatchesParameter = "control:call.argument_type_matches_parameter"
	// ControlCallTargetTypeFromCalleeReturn names 07-09's target-type
	// derivation claim (T-07-09-02): OpCall's TargetID.TypeID is derived
	// from the callee's declared return type resolved against the
	// caller's own type facts, fail-closed when unresolvable, never
	// copied from the caller's argument place. Killed on the check side
	// by check_test.go's TestCallReturnTypeDerivationMutationKilled and
	// on the peer side by corevalidate_test.go's
	// TestCallTypePeerMutationMatrix. Currently reachable from source only
	// in the ADMITTING direction (every function's declared return type
	// equals its declared parameter type this phase, D-07-09's language
	// surface constraint) -- the REFUSING direction
	// (check.call_return_type_unrepresentable /
	// core.CallReturnTypeMismatch) is mutation-killed through a seeded
	// seam and corevalidate's synthetic input space, never through a
	// .lang fixture; see PHASE-07-DEBT.md. This lane's own fixtures all
	// admit, exactly as ControlCallArgumentTypeMatchesParameter above.
	ControlCallTargetTypeFromCalleeReturn = "control:call.target_type_from_callee_return"
	// ControlCheckPeerConsulted names 07-10 Task 1's D-07-10 union-rule
	// claim (07-REVIEW.md CR-04 / PVG-03): CheckCommandFile consults
	// corevalidate.Validate on every source it admits and reports the
	// peer's own refusal code, never silently discarding it. Killed by
	// checkCommandPeerSeam (session.go) and
	// session_peer_gate_test.go's TestCheckCommandPeerConsultMutationKilled.
	ControlCheckPeerConsulted = "control:check.peer_consulted"
	// ControlInterfacePeerRefusalIsInvalid names 07-10 Task 2's claim: a
	// corevalidate refusal on the `interface export` / `interface core`
	// paths is reported as protocol.StatusInvalid carrying the peer's own
	// code, never as tool.operation_failed/exit 3 with the code
	// discarded. Killed by interfacePeerRefusalSeam (session.go) and
	// session_peer_gate_test.go's TestInterfacePeerRefusalMutationKilled.
	ControlInterfacePeerRefusalIsInvalid = "control:interface.peer_refusal_is_invalid"
	// ControlCallArgumentConsumedWhenNoncopyable names 07-11's consume-on-call
	// claim (07-VERIFICATION.md PVG-01 / 07-REVIEW.md CR-01): a call
	// transfers its argument, and a NON-COPYABLE argument is consumed
	// (move-marked), checked independently at both admission layers --
	// check.resolveCallBinding (from the argument's own core.TypeFact via
	// the ability package) and corevalidate's OpCall replay (from the
	// emitted core artifact's own types[operation.TypeID].Shape via
	// corevalidate's own deriveAbility, no helper shared with check). Killed
	// on the check side by check_test.go's
	// TestCallArgumentConsumeMutationKilled and on the peer side by
	// corevalidate_test.go's TestCallConsumePeerMutationMatrix.
	// call_argument_used_twice.lang is this control's own standing negative
	// control -- declared, not implied.
	ControlCallArgumentConsumedWhenNoncopyable = "control:call.argument_consumed_when_noncopyable"
	// ControlCallCopyableArgumentNotConsumed names 07-11's non-refusing-
	// direction claim: a COPYABLE call argument is copied, never consumed
	// -- a copyable value's second use through a call must not be wrongly
	// refused. Killed on the check side by check_test.go's
	// TestCallArgumentConsumeOverRefusalMutationKilled and on the peer side
	// by corevalidate_test.go's TestCallConsumePeerMutationMatrix.
	// call_argument_used_once.lang and call_basic.lang are this control's
	// own admitting fixtures.
	ControlCallCopyableArgumentNotConsumed = "control:call.copyable_argument_not_consumed"
	// ControlSummaryForeignReachClosureDerived names 07-12's Foreign
	// closure-join claim (07-VERIFICATION.md PVG-02 / 07-REVIEW.md CR-03):
	// core.FunctionSignature.Foreign is closure-derived over the acyclic
	// call graph, joined worst-case (empty is the identity; equal merges;
	// "forbidden" beats "permitted"; a genuine allocator disagreement
	// resolves to the declared core.ForeignReachConflict sentinel), before
	// ClosureDigest is computed -- checked independently at both the
	// producer (originvalidate.BuildInterface's chainOrder loop) and the
	// peer (corevalidate's own v.peerPostorder/v.peerAdjacency walk in
	// chainPeerClosureDigests), with no helper shared between the two.
	// Killed on the producer side by originvalidate's
	// TestForeignClosureJoinMutationKilled and on the peer side by
	// corevalidate's TestForeignClosureJoinPeerMutationMatrix.
	// call_fallible_foreign_reach.lang is this control's own standing
	// witness -- declared, not implied.
	ControlSummaryForeignReachClosureDerived = "control:summary.foreign_reach_closure_derived"
	// ControlSummaryFailsClosureDerived names 07-12's Fails closure-join
	// claim (07-VERIFICATION.md PVG-02 / 07-REVIEW.md CR-03):
	// core.FunctionSignature.Fails is closure-derived over the same acyclic
	// call graph and the same before-ClosureDigest placement, joined
	// first-non-empty-wins (a disclosed imprecision, D-07-53: Fails is a
	// single string and cannot express a union of two distinct error
	// types), checked independently at both the producer and the peer with
	// no shared helper. Killed on the producer side by originvalidate's
	// TestFailsClosureJoinMutationKilled and on the peer side by
	// corevalidate's TestForeignClosureJoinPeerMutationMatrix.
	// call_fallible_foreign_reach.lang is this control's own standing
	// witness too.
	ControlSummaryFailsClosureDerived = "control:summary.fails_closure_derived"
)

// phase07DispatchFixtures is the literal, hand-maintained, phase-scoped
// fixture list (A-05) this lane drives through corevalidate, pathoracle,
// originvalidate, interp, and cgen -- the CLI-observable sibling of
// core_test.go's own fixtures slice. Phase 07 does not extend or reuse
// Phase 4's dispatchFixtures (session.go); it carries its own.
// deep_diamond_acyclic.lang (07-06 Task 3) is wired in here because it
// accepts (checks clean, bounded) -- a refusing fixture like
// cycle_self.lang or cycle_mutual.lang must never be added to a list this
// lane expects a clean check from.
var phase07DispatchFixtures = []string{"call_basic.lang", "call_from_both_match_arms.lang", "deep_diamond_acyclic.lang"}

// phase07LaneDispatchFixturesOverride and phase07LaneRequiredKindsOverride
// are Task 3's D-07-41/D-07-42 fault-injection seams for the phase07 lane
// (Test 2, control:kind.exhaustive_dispatch.phase07_lane's own kill):
// unexported, nil by default (production callers never set these), and
// exercised only via session_phase7_export_test.go's setters from the
// external session_test package -- the same shape as
// corevalidate.disableCalleeResolutionCheckForTest (07-02). A nil override
// means "use the real default"; a non-nil (possibly empty) override
// replaces it for the duration of the test.
var (
	phase07LaneDispatchFixturesOverride []string
	phase07LaneRequiredKindsOverride    []core.OperationKind
)

// phase07LaneDispatchFixtures resolves the override seam above, or the
// real default when unset.
func phase07LaneDispatchFixtures() []string {
	if phase07LaneDispatchFixturesOverride != nil {
		return phase07LaneDispatchFixturesOverride
	}
	return phase07DispatchFixtures
}

// phase07LaneRequiredKinds resolves the override seam above, or the real
// default ([]core.OperationKind{core.OpCall}) when unset.
func phase07LaneRequiredKinds() []core.OperationKind {
	if phase07LaneRequiredKindsOverride != nil {
		return phase07LaneRequiredKindsOverride
	}
	return []core.OperationKind{core.OpCall}
}

// phase07FunctionHasOpCall mirrors core_test.go's functionHasOpCall (a
// deliberate, small duplication across the in-process and CLI-observable
// controls, matching A-05's requirement that each control carry its own
// literal fixture/logic rather than share a single implementation two
// independent proofs could both silently rot alongside).
func phase07FunctionHasOpCall(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall {
			return true
		}
	}
	return false
}

// phase07LinearProbeInput mirrors core_test.go's linearProbeInput: a
// straight-line function's sole declared parameter type selects a fixed,
// hand-picked probe input string for interp.Run. Both testdata/phase07
// fixtures' straight-line functions declare a Byte parameter.
func phase07LinearProbeInput(function core.Function) (string, bool) {
	switch function.Parameter.Type {
	case "Byte":
		return "7", true
	case "Buffer":
		return "01020304", true
	default:
		return "", false
	}
}

// VerifyPhase7ControlsAndWork is Phase 07's own control-and-work gate
// (D-07-41), reachable through `lang verify testdata/phase07` via
// cmd/lang/main.go's isPhase7Corpus dispatch (mirroring isPhase5Corpus/
// isPhase6Corpus). It drives phase07DispatchFixtures through check,
// corevalidate, pathoracle, originvalidate, interp, and cgen exactly like
// core_test.go's in-process control, and requires core.OpCall to be
// encountered by at least one operation across the corpus.
//
// D-07-39 -- recognition, not execution: where a function contains a
// core.OpCall, interp.Run returning the named interp.ErrCallUnsupported is
// treated as "handled" (recognized), never as a lane failure. Any OTHER
// interp error still fails this lane.
//
// A-02 -- cgen's runtime behaviour for core.OpCall is exercised by NEITHER
// this lane NOR core_test.go's in-process control, because cgen.Emit hard-
// fails on len(program.Functions) != 1 (cgen.go:22) before its OpCall arm
// could ever run, and no legal OpCall-bearing program in this phase's
// corpus has exactly one function. This lane's own `if len(program.Functions)
// == 1` gate below is therefore never taken for either phase07 fixture --
// declared here, in PHASE-07-DEBT.md, and in 07-04-SUMMARY.md, never implied.
func VerifyPhase7ControlsAndWork(ctx context.Context) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("verify", protocol.StatusPass)

	addLane := func(id, status string, controls []string, work int, laneStarted time.Time) {
		result.Lanes = append(result.Lanes, protocol.Lane{
			Schema: protocol.LaneSchema1, ID: id, Status: status,
			Controls: append([]string{}, controls...), RecomputedWork: work,
			ElapsedNS: time.Since(laneStarted).Nanoseconds(), PeakRSSStatus: protocol.PeakRSSUnavailable,
		})
		result.Metrics.RecomputedWork += work
	}
	fail := func(status, code, message string) (protocol.Result, error) {
		result.Status = status
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error(code, diagnostic.Span{}, message))
		result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
		return result.Finalize(), nil
	}

	corpus := nat03CorpusPath("testdata/phase07")
	laneStarted := time.Now()
	encounteredKinds := make(map[core.OperationKind]bool)
	dispatchWork := 0
	for _, fixtureName := range phase07LaneDispatchFixtures() {
		fixtureSource, fixtureErr := readBoundedFile(filepath.Join(corpus, fixtureName), syntax.MaxSourceBytes)
		if fixtureErr != nil {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", fixtureName)
		}
		fixtureChecked := Check(fixtureSource)
		if len(fixtureChecked.Diagnostics) != 0 {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
			return fail(protocol.StatusInvalid, "verify.fixture_rejected", fixtureName)
		}
		dispatchWork += fixtureChecked.Work
		fixtureValidated := corevalidate.Validate(fixtureChecked.Program)
		if !fixtureValidated.Valid {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+fixtureValidated.Checks, laneStarted)
			return fail(protocol.StatusInvalid, "verify.fixture_rejected", fixtureName)
		}
		dispatchWork += fixtureValidated.Checks
		dispatchProgram := fixtureValidated.Program()
		for _, function := range dispatchProgram.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encounteredKinds[operation.Kind] = true
				}
				if function.Linear.ID != "" {
					if _, _, oracleErr := pathoracle.RecomputeEndpoints(function); oracleErr != nil {
						addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
						return fail(protocol.StatusOperational, "verify.control_incomplete", "pathoracle dispatch error for "+fixtureName)
					}
				}
			}
			_ = originvalidate.RecomputeOriginPerReturn(function)
			dispatchWork++

			hasCall := phase07FunctionHasOpCall(function)
			switch {
			case function.Match != nil:
				for _, arm := range function.Match.Arms {
					if _, interpErr := interp.Run(dispatchProgram, function.Name, arm.Pattern); interpErr != nil {
						// D-07-39: recognized, not executed -- see the
						// doc comment above.
						if !hasCall || !errors.Is(interpErr, interp.ErrCallUnsupported) {
							addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
							return fail(protocol.StatusOperational, "verify.control_incomplete", "interp dispatch error for "+fixtureName)
						}
					}
					dispatchWork++
				}
			case function.Linear != nil:
				input, ok := phase07LinearProbeInput(function)
				if ok {
					if _, interpErr := interp.Run(dispatchProgram, function.Name, input); interpErr != nil {
						// D-07-39: recognized, not executed -- see the
						// doc comment above.
						if !hasCall || !errors.Is(interpErr, interp.ErrCallUnsupported) {
							addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
							return fail(protocol.StatusOperational, "verify.control_incomplete", "interp dispatch error for "+fixtureName)
						}
					}
					dispatchWork++
				}
			}
		}
		// A-02: this gate is never taken for either phase07 fixture -- see
		// the doc comment above.
		if len(dispatchProgram.Functions) == 1 {
			if _, cgenErr := cgen.Emit(dispatchProgram); cgenErr != nil {
				addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
				return fail(protocol.StatusOperational, "verify.control_incomplete", "cgen dispatch error for "+fixtureName)
			}
			dispatchWork++
		}
	}
	for _, kind := range phase07LaneRequiredKinds() {
		if !encounteredKinds[kind] {
			addLane("lane:kind-exhaustive-dispatch-phase07", "fail", nil, dispatchWork+1, laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_missing", "control:kind.exhaustive_dispatch.phase07_lane (kind "+string(kind)+" never encountered)")
		}
	}
	addLane("lane:kind-exhaustive-dispatch-phase07", "pass", Phase7RequiredControls(), dispatchWork+1, laneStarted)

	for _, required := range Phase7RequiredControls() {
		if !hasControl(result.Lanes, required) {
			return fail(protocol.StatusInvalid, "verify.control_missing", required)
		}
	}
	for _, lane := range result.Lanes {
		if lane.RecomputedWork == 0 {
			return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID)
		}
	}
	result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
	return result.Finalize(), nil
}

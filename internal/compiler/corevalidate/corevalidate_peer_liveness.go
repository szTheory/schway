package corevalidate

import "github.com/szTheory/schway/internal/compiler/core"

// This file is Phase 09's own interprocedural loan-liveness derivation
// (D-09-01, D-09-02, D-09-04): corevalidate's independent re-derivation of
// the fact `check`'s buildInterproceduralSummaries/deriveFunctionUsesParam
// (check.go:579-728) exists to answer, "does this function's return value
// derive from a borrow of its own parameter, transitively through calls".
//
// What is reused from Phase 07 is SUBSTRATE, never derivation: the
// postorder ordering plumbing (v.peerPostorder, built once by
// checkCallGraphAcyclic's own independent DFS, corevalidate.go:130-138) and
// the memoized callee-before-caller consumption idiom
// chainPeerClosureDigests already uses (corevalidate.go:495-549). Nothing
// about the DERIVATION itself -- the actual forward set-propagation walk
// below -- is shared with, or borrowed from, that substrate; this is
// precisely the conflation a reviewer will probe first, so it is named
// here explicitly rather than left implicit.
//
// The derivation itself is a FORWARD set-propagation over
// function.Linear.Operations (D-09-02), matching the idiom
// peerReturnDerivesFromBorrow/peerParameterEscapesOwned already use
// (corevalidate.go:2061-2131) -- and structurally OPPOSITE check's backward
// memoized walk over callgraph.Order (deriveFunctionUsesParam,
// check.go:686-728). Independence is proven by derivation method plus
// import boundary, never by file or package location (D-09-02): this file
// imports only core, never callgraph, check, ast, originvalidate,
// interp, cgen, session, or ability (TestPeerLivenessFileImportsStayIndependent).

// peerLoanCarryFact is corevalidate's own interprocedural loan-liveness
// fact for one function: whether some value this function returns derives
// from a borrow of the function's own parameter, and therefore "carries" a
// loan across any call site whose result is assigned that call's return. A
// struct, not a bare bool, because a later plan widens this with an
// access-mode payload (shared vs exclusive), following the same
// first-hop-wins idiom peerReturnDerivesFromBorrow already establishes.
type peerLoanCarryFact struct {
	ReturnsBorrowOfParam bool
}

// derivePeerLoanCarry is the single deterministic forward pass over
// function.Linear.Operations that computes one function's own
// peerLoanCarryFact, given every callee's ALREADY-FINAL fact in callee
// (chainPeerLoanCarry guarantees this by walking v.peerPostorder
// callee-before-caller, exactly like chainPeerClosureDigests). It mirrors
// peerReturnDerivesFromBorrow's exact shape: a paramTrace set seeded with
// the parameter's own place, a derived set built by one forward scan, and
// a second short scan checking OpReturn -- no recursion, no worklist, no
// inner fixpoint loop, and no import of callgraph.
//
// The OpCall case is the one addition beyond peerReturnDerivesFromBorrow's
// own shape: a call's result (TargetID) enters derived when the CALLEE's
// own fact says it returns a borrow of ITS parameter, AND the call's
// argument (SourceID) is itself something this function already tracks as
// parameter-derived. An unknown CalleeID (absent from callee, which can
// only happen for a corrupted core artifact -- checkCallGraphAcyclic
// refuses a genuinely cyclic or dangling-edge program before this ever
// runs) is treated as NOT carrying: fail-closed for a corrupted artifact
// (T-09-02), and unreachable in practice because chainPeerLoanCarry's own
// postorder walk finalizes every callee first.
//
// Transitivity across arbitrary call depth (D-09-49 Q2) composes through
// this single pass with no fixpoint iteration at all: because
// v.peerPostorder is callee-before-caller BY CONSTRUCTION (it is this
// package's own DFS postorder byproduct, corevalidate.go:130-138 -- never
// check's callgraph.Order, which is caller-before-callee and requires
// buildInterproceduralSummaries to walk it backward, D-08-42), every
// callee's fact is already final and non-changing by the time a caller
// reads it here. A caller therefore reads a finalized callee fact exactly
// once, regardless of how many hops separate the caller from the function
// that actually originates the borrow -- depth-2 (b forwards c's result),
// depth-3, and beyond all resolve correctly from the SAME single forward
// pass per function, chained only through the postorder ordering, never
// through re-walking a callee's own body a second time.
func derivePeerLoanCarry(function *core.Function, callee map[string]peerLoanCarryFact) peerLoanCarryFact {
	if function.Linear == nil {
		return peerLoanCarryFact{}
	}
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
		case core.OpCall:
			if (paramTrace[operation.SourceID] || derived[operation.SourceID]) && callee[operation.CalleeID].ReturnsBorrowOfParam {
				derived[operation.TargetID] = true
			}
		}
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpReturn && derived[operation.SourceID] {
			return peerLoanCarryFact{ReturnsBorrowOfParam: true}
		}
	}
	return peerLoanCarryFact{}
}

// chainPeerLoanCarry folds derivePeerLoanCarry into the SAME postorder
// substrate chainPeerClosureDigests already consumes (corevalidate.go's
// v.peerPostorder), read in the SAME loop shape: iterate the postorder
// (callee-before-caller by construction), and for each function ID derive
// its own fact from every callee's already-stored entry before storing the
// caller's own. Called from run(), immediately after checkCallGraphAcyclic
// has populated v.peerPostorder and strictly before any replay consumes
// v.peerLoanCarry via buildLoanChainIndex -- see the call site's own
// comment in run() for why this ordering point differs from
// chainPeerClosureDigests' own (later) call site. Rebuilt fresh on every
// Validate call, never persisted (D-09-06).
func (v *validator) chainPeerLoanCarry() {
	functionByID := make(map[string]*core.Function, len(v.program.Functions))
	for index := range v.program.Functions {
		functionByID[v.program.Functions[index].ID] = &v.program.Functions[index]
	}
	v.peerLoanCarry = make(map[string]peerLoanCarryFact, len(v.peerPostorder))
	for _, id := range v.peerPostorder {
		function, ok := functionByID[id]
		if !ok {
			continue
		}
		v.peerLoanCarry[id] = derivePeerLoanCarry(function, v.peerLoanCarry)
	}
}

// disablePeerLoanCarryConsultForTest is Phase 09's own D-09-25/D-09-26
// fault-injection seam (QLT-08): when true, buildLoanChainIndex's OpCall
// branch never breaks the chain at a call boundary at all -- reproducing
// the pre-Phase-09 unconditional-propagation shape from the peer's own
// side, independently of check's own interprocedural liveness law.
// Unexported, false in production; set only via
// SetDisablePeerLoanCarryConsultForTest by a test that defers the restore
// immediately.
var disablePeerLoanCarryConsultForTest bool

// forcePeerLoanCarryTrueForTest is Phase 09's own D-09-25/D-09-26
// fault-injection seam, the "always carrying" counterpart of
// disablePeerLoanCarryConsultForTest above: when true, every OpCall is
// treated as though its callee's fact reports ReturnsBorrowOfParam ==
// true, regardless of what derivePeerLoanCarry actually derived --
// reintroducing the two retired peerDivergenceExpected fixtures'
// over-refusal from a DIFFERENT mechanism than disabling the consult
// entirely, so the two seams can each be shown to fail the mutation-kill
// independently. Unexported, false in production; set only via
// SetForcePeerLoanCarryTrueForTest.
var forcePeerLoanCarryTrueForTest bool

// SetDisablePeerLoanCarryConsultForTest is D-09-25/D-09-26's cross-package
// fault-injection seam for disablePeerLoanCarryConsultForTest above. Go's
// build model excludes "_test.go" files from a normal package import, so a
// same-package-only unexported var (the shape corevalidate_peer_liveness_test.go
// itself uses to set the var directly) cannot be reached by check's own
// companion-assertion test (check_peer_liveness_seam_test.go), which
// imports this package as an ordinary dependency -- exactly the same
// cross-package exception SetDisableCyclePeerForTest above documents
// (D-07-42). Production-visible, but a documented test-only no-op unless a
// test explicitly calls it, and always restored via the returned closure.
// Never called from any production code path in this repository.
func SetDisablePeerLoanCarryConsultForTest(disable bool) (restore func()) {
	previous := disablePeerLoanCarryConsultForTest
	disablePeerLoanCarryConsultForTest = disable
	return func() { disablePeerLoanCarryConsultForTest = previous }
}

// SetForcePeerLoanCarryTrueForTest is forcePeerLoanCarryTrueForTest's own
// cross-package counterpart, for the same reason
// SetDisablePeerLoanCarryConsultForTest exists.
func SetForcePeerLoanCarryTrueForTest(force bool) (restore func()) {
	previous := forcePeerLoanCarryTrueForTest
	forcePeerLoanCarryTrueForTest = force
	return func() { forcePeerLoanCarryTrueForTest = previous }
}

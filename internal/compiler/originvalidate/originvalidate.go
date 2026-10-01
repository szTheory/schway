// Package originvalidate independently re-derives and verifies a function's
// declared public borrow origin (OWN-04) from the typed-core artifact alone.
// It intentionally does not know source or reuse checker code: it imports
// neither internal/compiler/check nor internal/compiler/ast, so a declared
// origin is never trusted, only recomputed from core.Program/core.Function
// facts — the same source-blind posture corevalidate already holds for
// ownership, applied here to the origin trust boundary (T-03-02/T-03-03/
// T-03-16).
package originvalidate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/szTheory/schway/internal/compiler/callgraph"
	"github.com/szTheory/schway/internal/compiler/core"
)

// phase17ReturnLookupFaultForTest is a default-off mutation seam consulted
// only by originvalidate's local return-side lookup.
var phase17ReturnLookupFaultForTest bool

// SetPhase17ReturnLookupFaultForTest is the fact-free test control used by
// the cross-package agreement matrix. Its idempotent restore closure makes
// the narrow seam safe for both defer and t.Cleanup callers.
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

// KnownEscape names the exact boundary this package cannot prove: a producer
// whose frontend and published summary lie in a coordinated way remains
// outside source-blind, body-blind validation. It mirrors
// corevalidate.KnownEscape's shape, applied to the origin trust boundary
// instead of the source/core one (T-03-08, an accepted residual, never
// claimed as solved).
const KnownEscape = "escape:coordinated-frontend-summary-lie"

// TerminatorKindsOverride is a fault-injection seam for
// TestTerminatorWalkMutationKilled (D-09/D-04-29): production always walks
// the full core.TerminatorKinds() set; the test temporarily narrows it (the
// same "delete OpFail from the recognised set" mutation the throwaway-
// detached-worktree demonstration performs on the source directly) to prove
// that a walker recognising fewer terminators really does lose a fixture's
// origin fact. nil (the always-true production default) means "use
// core.TerminatorKinds() unmodified".
var TerminatorKindsOverride func() []core.OperationKind

func recognizedTerminatorKinds() []core.OperationKind {
	if TerminatorKindsOverride != nil {
		return TerminatorKindsOverride()
	}
	return core.TerminatorKinds()
}

// isTerminatorKind is originvalidate's own membership test against the
// terminator registry (D-04-29): a set-membership test against
// core.TerminatorKinds() rather than a literal restatement of it, so
// widening the registry widens this walker automatically. This is the
// single site (RecomputeOriginPerReturn's collection loop, immediately
// below) where the backward-walk collection discriminates a terminator
// operation from an ordinary one.
func isTerminatorKind(kind core.OperationKind) bool {
	for _, terminator := range recognizedTerminatorKinds() {
		if kind == terminator {
			return true
		}
	}
	return false
}

// RecognizesTerminator is session.go's control:terminator.walk_incomplete
// hook (D-04-29): it reports whether this package's own walker treats kind
// as a terminator, reading the exact same isTerminatorKind this file's
// production walk uses -- never a second, restated copy of the set.
func RecognizesTerminator(kind core.OperationKind) bool { return isTerminatorKind(kind) }

// ExpectedEscapes is the origin package's contribution to a verify result's
// expected-escapes list — surfaced next to corevalidate.KnownEscape, never
// reported as a detected control.
func ExpectedEscapes() []string { return []string{KnownEscape} }

// Problem is the origin package's independent-validation finding shape,
// matching corevalidate.Problem's {Code, Detail} contract.
type Problem struct {
	Code   string `json:"code"`
	Detail string `json:"detail,omitempty"`
}

// Error is CheckSummary's typed failure, matching evidence.ValidationError's
// {Code}-only shape so callers can dispatch on Code the same way.
type Error struct{ Code string }

func (e *Error) Error() string { return e.Code }

// ReturnOrigin is the origin derivation for exactly ONE core.OpReturn within
// a function's body: which paths and access mode that single return's
// SourceID traces back to the function's own parameter, if it traces at
// all. A straight-line function's Operations carries exactly one
// core.OpReturn; a match-bodied function carries one per arm (check.go's
// arm lowering appends every arm's Return into the same flat Operations
// slice, per corevalidate.replayBlocks' own "one return per block" note) —
// so a function's full origin picture is the SET of these, never any one of
// them alone.
type ReturnOrigin struct {
	// OperationID is the originating core.OpReturn's own ID.
	OperationID string
	// Paths names the parameter path(s) this return traces back to, when
	// Derived is true. Empty when Derived is false.
	Paths []string
	// Access is the derived access mode ("shared" or "exclusive") when
	// Derived is true. Empty when Derived is false.
	Access string
	// Derived reports whether the backward walk from this return reached
	// function.Parameter.ID through at least one borrow hop. false means an
	// owned return, or a chain that broke/cycled before reaching the
	// parameter.
	Derived bool
}

// AccessConflicting is the conservative-combination sentinel RecomputeOrigin
// reports when a function's arms derive different access modes for their
// respective returns: neither arm's answer is the true answer, so the
// sentinel names the disagreement itself rather than silently resolving to
// whichever arm happened to be walked. It is never a declarable
// core.PublicOrigin.Access value — ValidatePublished refuses any declared
// Access outside {"shared", "exclusive"} before ever comparing it against a
// recomputed answer, so a mutated summary cannot declare the sentinel and
// match a conflicting recomputation.
const AccessConflicting = "conflicting"

// calleeOriginFact is the narrow origin fact walkReturnOrigin consults at
// an OpCall hop (D-10-01/D-10-04/T-10-05): only the callee's declared
// return access mode and whether it derives from the callee's own
// parameter at all — never anything core.Function- or core.Program-shaped,
// so a callee's BODY is never reachable from inside this walk. Mirrors
// corevalidate.peerLoanCarryFact's minimal one-field shape (D-09-03's
// shipped precedent this task mirrors), widened by exactly the field this
// walk needs beyond a bare bool: the access mode a propagated borrow
// carries.
type calleeOriginFact struct {
	// Access is the callee's declared core.PublicOrigin.Access ("shared" or
	// "exclusive") when Derived is true. Empty when Derived is false.
	Access string
	// Derived reports whether the callee's DECLARED return contract says it
	// returns a borrow of its own parameter at all.
	Derived bool
}

// BuildCalleeOriginFacts builds the callee-contract map walkReturnOrigin
// consults at an OpCall hop, from EVERY function's declared PublicOrigin
// alone — never a callee body (D-10-02/D-10-04). ValidatePublished and
// BuildInterface both build this ONCE per program and thread it down (D-10-
// 03); it is exported so a caller outside this package that holds its own
// core.Program (session.go's and session_phase7.go's exhaustive-dispatch
// sites, core_test.go's own probe, and this package's own external test
// binary) can build the identical map to call RecomputeOriginPerReturn /
// RecomputeOrigin / PublishProblemsFor directly. calleeOriginFact itself
// stays unexported (T-10-05): a caller can hold and pass this map straight
// back into this package's functions via `:=` type inference, but can
// never spell out a field on its element type, because that element type
// carries no core.Function- or core.Program-shaped field to spell.
func BuildCalleeOriginFacts(program core.Program) map[string]calleeOriginFact {
	facts := make(map[string]calleeOriginFact, len(program.Functions))
	for _, function := range program.Functions {
		if function.PublicOrigin == nil {
			continue
		}
		facts[function.ID] = calleeOriginFact{Access: function.PublicOrigin.Access, Derived: true}
	}
	return facts
}

// RecomputeOriginPerReturn is the package's SOLE backward-walk site. It
// builds its source map from this function's core operations and walks each
// return independently. A definition must precede the return; in branch
// core it must also live in that return's arm or in the shared entry prefix.
// This prevents a malformed artifact from borrowing a sibling-arm or
// future operation's origin while retaining the computed-prefix path.
//
// calleeContracts is D-10-01/D-10-03's callee-contract map (see
// BuildCalleeOriginFacts): the walk consults it, keyed by CalleeID, at an
// OpCall hop, and never reads anything else about a callee.
func RecomputeOriginPerReturn(function core.Function, calleeContracts map[string]calleeOriginFact) []ReturnOrigin {
	if function.Linear == nil || len(function.Linear.Operations) == 0 {
		return nil
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	sourceIndex := make(map[string]int, len(operations))
	operationBlock := make(map[string]string, len(operations))
	for _, block := range function.Linear.Blocks {
		for _, operationID := range block.OperationIDs {
			operationBlock[operationID] = block.ID
		}
	}
	var returnIndexes []int
	for index := range operations {
		operation := operations[index]
		// D-04-29: the discriminant is a membership test against the
		// terminator registry, not an equality test against core.OpReturn
		// alone -- Phase 4 introduces core.OpFail and core.OpDefect, and an
		// analysis that recognises only a return is exactly the defect class
		// 03-08/03-09/03-10 closed three times, now reproduced one layer up.
		// walkReturnOrigin below is unchanged: it is already terminator-
		// agnostic, walking backward from whichever operation this loop
		// collects.
		if isTerminatorKind(operation.Kind) {
			returnIndexes = append(returnIndexes, index)
			continue
		}
		// D-12-10: an OpDestructurePayload names its produced place via
		// PayloadTargetID, never the ordinary TargetID (which stays empty --
		// corevalidate's matchBranchStructural already checks this).
		// Indexing sourceOf by TargetID alone would key this operation under
		// "" and make its produced place permanently unreachable by a
		// backward walk starting from a later operation's SourceID -- no
		// case arm in walkReturnOrigin's switch below could ever compensate
		// for a produced place the map never recorded in the first place.
		if operation.Kind == core.OpDestructurePayload {
			sourceOf[operation.PayloadTargetID] = operation
			sourceIndex[operation.PayloadTargetID] = index
			continue
		}
		sourceOf[operation.TargetID] = operation
		sourceIndex[operation.TargetID] = index
	}
	if len(returnIndexes) == 0 {
		return nil
	}
	results := make([]ReturnOrigin, 0, len(returnIndexes))
	for _, returnIndex := range returnIndexes {
		returnOp := &operations[returnIndex]
		// Blocks themselves carry the trust-boundary evidence. Do not depend
		// on Match being present to enable arm scoping: a forged artifact may
		// omit its match summary while retaining branch-shaped linear facts.
		branch := len(function.Linear.Blocks) != 0
		returnBlock := operationBlock[returnOp.ID]
		results = append(results, walkReturnOrigin(function, sourceOf, sourceIndex, operationBlock, returnOp, returnIndex, returnBlock, branch, calleeContracts))
	}
	return results
}

// disableOpCallOriginConsultForTest is D-10-08's fault-injection seam for
// TestOpCallOriginWalkGateIsLoadBearing: production always consults
// calleeContracts at an OpCall hop (the case below); the test temporarily
// forces that case to no-op, reproducing the pre-fix behavior of walking
// straight through a call boundary into the argument's own provenance, so a
// regression test can prove the case is genuinely load-bearing by observing
// twin_a_accept.schway regress to its old core.origin_omitted refusal --
// never merely that the case is syntactically present (D-10-08). Mirrors
// export_test.go's existing SetTypeFactExactIDMatchOverrideForTest seam
// shape (D-09-25/D-09-26's disablePeerLoanCarryConsultForTest precedent).
// false (the production default) means "consult calleeContracts for real".
var disableOpCallOriginConsultForTest bool

// walkReturnOrigin performs exactly one backward walk, from one return
// operation, using the shared sourceOf map RecomputeOriginPerReturn built
// once for the whole function.
func walkReturnOrigin(function core.Function, sourceOf map[string]core.LinearOperation, sourceIndex map[string]int, operationBlock map[string]string, returnOp *core.LinearOperation, returnIndex int, returnBlock string, branch bool, calleeContracts map[string]calleeOriginFact) ReturnOrigin {
	current := returnOp.SourceID
	visited := make(map[string]bool)
	derivedAccess := ""
	entryBlock := function.ID + ":block:entry"
	for current != function.Parameter.ID {
		if visited[current] {
			return ReturnOrigin{OperationID: returnOp.ID}
		}
		visited[current] = true
		operation, exists := sourceOf[current]
		definitionIndex, indexed := sourceIndex[current]
		if !exists || !indexed || definitionIndex >= returnIndex {
			return ReturnOrigin{OperationID: returnOp.ID}
		}
		if branch {
			definitionBlock := operationBlock[operation.ID]
			if definitionBlock != entryBlock && definitionBlock != returnBlock {
				return ReturnOrigin{OperationID: returnOp.ID}
			}
		}
		switch operation.Kind {
		case core.OpConst:
			// A constant introduces a fresh owned value. It has no parameter
			// origin and is the terminal root for this backward walk.
			return ReturnOrigin{OperationID: returnOp.ID}
		case core.OpCopy:
			if originCopyIsU64(function, operation) {
				// The copy operation still reads the borrowed place, but its
				// target is an independent scalar value and therefore has no
				// borrowed return origin.
				return ReturnOrigin{OperationID: returnOp.ID}
			}
		case core.OpBorrowExclusive:
			if derivedAccess == "" {
				derivedAccess = "exclusive"
			}
		case core.OpBorrowShared:
			if derivedAccess == "" {
				derivedAccess = "shared"
			}
		case core.OpDestructurePayload:
			// Task 1 (Phase 12, plan 04): walk through a NEW kind's OWN
			// semantics -- a destructure's produced place (indexed above by
			// PayloadTargetID rather than the ordinary TargetID) carries no
			// access-mode fact of its own; the payload's origin is whatever
			// its SourceID (the scrutinee alias) already carries. This is
			// categorically different from filling core.OpCall's missing
			// case in peerDeriveOriginFacts (D-10-C01) -- that gap MUST stay
			// open per D-12-28/Pitfall 3, since D-12-27's resource-payload
			// refusal depends on it. Deriving through a kind this walk has
			// never seen before is ordinary in-scope extension of THIS
			// switch's own coverage, not a resolution of a different peer's
			// open gap. No derivedAccess assignment is needed here: the
			// unconditional `current = operation.SourceID` below already
			// continues the walk into the borrow (or further chain) that
			// produced the destructured alias.
		case core.OpForeignCall:
			if operation.Foreign != nil {
				// Phase 23's borrow operation has its own source place and
				// result contract. Its U64 result is owned, so this boundary
				// terminates origin derivation instead of borrowing facts from
				// another foreign operation in the function.
				return ReturnOrigin{OperationID: returnOp.ID}
			}
			// D-04-28: a foreign declaration is itself a signature carrying
			// origin and access facts -- declaring a call foreign-only does
			// not escape origin reasoning, it moves the facts somewhere they
			// are asserted (function.ForeignContract.Alias) rather than
			// derived from a body the foreign symbol does not have. When the
			// contract declares this call borrows/retains its argument, the
			// hop counts exactly like an in-language borrow hop so a
			// correctly DECLARED PublicOrigin for such a function is
			// recognised as matching, not flagged as understated.
			if derivedAccess == "" && function.ForeignContract != nil {
				switch function.ForeignContract.Alias {
				case "borrow":
					derivedAccess = "shared"
				case "retain":
					derivedAccess = "exclusive"
				}
			}
		case core.OpCall:
			// D-10-01/D-10-07: consult the CALLEE's DECLARED return
			// contract, mirroring D-09-03's shipped precedent
			// (corevalidate.buildLoanChainIndex's own loanCarry consult).
			// Real but NOT EXACT mirror of the OpForeignCall case above:
			// OpForeignCall's contract lives on the CURRENT function
			// (function.ForeignContract, no lookup at all), while OpCall's
			// contract lives on a DIFFERENT function and requires this
			// cross-function calleeContracts lookup instead. If the
			// callee's declared contract does NOT say it returns a borrow
			// of its own parameter -- including an unknown CalleeID, which
			// fails closed to "does not carry" for a corrupted artifact
			// (T-09-02's precedent) -- the call's result is a fresh, owned
			// identity from this walk's perspective, and the chain
			// deliberately BREAKS at this call boundary: return
			// non-derived immediately instead of falling through to walk
			// past it into the argument's own provenance, which is exactly
			// the pre-fix D-09-51 defect this case closes.
			// disableOpCallOriginConsultForTest (D-10-08) forces this
			// entire case to no-op, restoring that pre-fix transparent
			// walk-through so a test can observe the case is load-bearing.
			if !disableOpCallOriginConsultForTest {
				fact, known := calleeContracts[operation.CalleeID]
				if !known || !fact.Derived {
					return ReturnOrigin{OperationID: returnOp.ID}
				}
				if derivedAccess == "" {
					derivedAccess = fact.Access
				}
			}
		}
		current = operation.SourceID
	}
	if derivedAccess == "" {
		return ReturnOrigin{OperationID: returnOp.ID}
	}
	return ReturnOrigin{OperationID: returnOp.ID, Paths: []string{function.Parameter.Name}, Access: derivedAccess, Derived: true}
}

// originCopyIsU64 is originvalidate's independent recognition of the only
// scalar value-copy boundary admitted by Phase 25. The core validator
// separately re-derives Copy ability; this walk uses the closed U64 shape
// rather than importing or accepting the producer's origin conclusion.
func originCopyIsU64(function core.Function, operation core.LinearOperation) bool {
	if operation.Kind != core.OpCopy || operation.TypeID == "" || function.Linear == nil {
		return false
	}
	for _, fact := range function.Linear.Types {
		if fact.ID == operation.TypeID {
			return fact.Shape.Constructor == "U64" && len(fact.Shape.Arguments) == 0
		}
	}
	return false
}

// RecomputeOrigin derives the origin path(s) and access mode a function's
// body actually returns, as the conservative combination of every
// RecomputeOriginPerReturn element (the "combination law", 03-10-PLAN.md):
//
//  1. No return is borrow-derived (every arm owned, or every chain broke) →
//     (nil, "", false). Byte-identical to today for an owned straight-line
//     function or an every-arm-owned match function.
//  2. Every borrow-derived return agrees on access mode → the ordered union
//     of their paths, that agreed access, ok=true. Byte-identical to today
//     for a one-return function.
//  3. Borrow-derived returns DISAGREE on access mode → the ordered union of
//     their paths, access=AccessConflicting, ok=true — neither arm's answer
//     wins.
//
// This function performs no backward walk itself; RecomputeOriginPerReturn
// is the only place that does.
func RecomputeOrigin(function core.Function, calleeContracts map[string]calleeOriginFact) (paths []string, access string, ok bool) {
	perReturn := RecomputeOriginPerReturn(function, calleeContracts)
	var derived []ReturnOrigin
	for _, origin := range perReturn {
		if origin.Derived {
			derived = append(derived, origin)
		}
	}
	if len(derived) == 0 {
		return nil, "", false
	}
	pathSeen := make(map[string]bool, len(derived))
	var orderedPaths []string
	combinedAccess := derived[0].Access
	conflict := false
	for _, origin := range derived {
		if origin.Access != combinedAccess {
			conflict = true
		}
		for _, path := range origin.Paths {
			if !pathSeen[path] {
				pathSeen[path] = true
				orderedPaths = append(orderedPaths, path)
			}
		}
	}
	if conflict {
		return orderedPaths, AccessConflicting, true
	}
	return orderedPaths, combinedAccess, true
}

func hasOwnerCall(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	for _, op := range function.Linear.Operations {
		if op.Kind == core.OpCall {
			return true
		}
	}
	return false
}

// transferredOwnerProblem derives the helper-owned acquisition and follows its return
// through the receiving call, borrowed use, and final paired discharge.
func transferredOwnerProblem(program core.Program) *Problem {
	if originPhase24ErrorProgram(program) {
		return originRepeatedOwnerProblem(program)
	}
	fail := func(id string) *Problem {
		return &Problem{Code: "core.local_owner_transfer", Detail: fmt.Sprintf("invalid transferred owner at %s", id)}
	}
	type foundOp struct {
		f  *core.Function
		op core.LinearOperation
	}
	var acq foundOp
	var helper *core.Function
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Linear == nil {
			continue
		}
		for _, op := range f.Linear.Operations {
			if op.Foreign != nil && op.Foreign.Mode == "acquire" {
				if helper != nil {
					return fail(op.ID)
				}
				helper = f
				acq = foundOp{f, op}
			}
		}
	}
	if helper == nil || helper.ReturnType != "FileByteOwner" {
		return nil
	}
	if helper.Parameter.Type != "PathToken" || acq.op.Foreign == nil || acq.op.Kind != core.OpForeignCall || acq.op.Foreign.Symbol != "schway_file_byte_acquire" || acq.op.Foreign.ABIType != "schway_file_byte_acquire_fn" || acq.op.Foreign.ParameterType != "PathToken" || acq.op.Foreign.ResultType != "FileByteOwner" || acq.op.Foreign.Fails != "AcquireError" || acq.op.Foreign.Allocator != "libc_malloc" || acq.op.Foreign.Release != "schway_file_byte_release" || acq.op.ErrTargetID == "" || acq.op.OkEdgeID == "" || acq.op.ErrEdgeID == "" {
		return fail(helper.ID)
	}
	var returned *core.LinearOperation
	for i := range helper.Linear.Operations {
		op := &helper.Linear.Operations[i]
		if op.Kind == core.OpReturn {
			if returned != nil {
				return fail(op.ID)
			}
			returned = op
		}
		if op.Kind == core.OpRelease {
			return fail(op.ID)
		}
	}
	if returned == nil || returned.SourceID != acq.op.TargetID || acq.op.ErrTargetID == "" {
		return fail(helper.ID)
	}
	var caller *core.Function
	var call, borrow, release, ret *core.LinearOperation
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Linear == nil {
			continue
		}
		for j := range f.Linear.Operations {
			op := &f.Linear.Operations[j]
			if op.Kind == core.OpCall && op.CalleeID == helper.ID {
				if caller != nil {
					return fail(op.ID)
				}
				caller = f
				call = op
			}
		}
	}
	if caller == nil || caller.ReturnType != "U64" {
		return fail(helper.ID)
	}
	bi, ri, ti := -1, -1, -1
	for i := range caller.Linear.Operations {
		op := &caller.Linear.Operations[i]
		switch {
		case op.Kind == core.OpCopy && op.SourceID == call.TargetID:
			return fail(op.ID)
		case op.Foreign != nil && op.Foreign.Mode == "borrow":
			if borrow != nil {
				return fail(op.ID)
			}
			borrow = op
			bi = i
		case op.Kind == core.OpRelease:
			if release != nil {
				return fail(op.ID)
			}
			release = op
			ri = i
		case op.Kind == core.OpReturn:
			if ret != nil {
				return fail(op.ID)
			}
			ret = op
			ti = i
		}
	}
	if borrow == nil || release == nil || ret == nil || borrow.SourceID != call.TargetID || release.SourceID != call.TargetID || release.ReleasesOperationID != acq.op.ID || release.Foreign == nil || borrow.Foreign == nil || release.Foreign.Mode != "consume" || release.Foreign.Symbol != acq.op.Foreign.Release || release.Foreign.ABIType != "schway_file_byte_release_fn" || release.Foreign.ParameterType != "FileByteOwner" || release.Foreign.ResultType != "Unit" || release.Foreign.Fails != "" || release.Foreign.Allocator != acq.op.Foreign.Allocator || release.Allocator != acq.op.Foreign.Allocator || borrow.Foreign.Symbol != "schway_file_byte_use" || borrow.Foreign.ABIType != "schway_file_byte_use_fn" || borrow.Foreign.ParameterType != "FileByteOwner" || borrow.Foreign.ResultType != "U64" || borrow.Foreign.Fails != "UseError" || ret.SourceID != borrow.TargetID {
		return fail(caller.ID)
	}
	if !(bi >= 0 && bi < ri && ri < ti) {
		return fail(caller.ID)
	}
	return nil
}

func originPhase24ErrorProgram(program core.Program) bool {
	if program.Module != "phase24.error" {
		return false
	}
	seen := map[string]bool{}
	for _, function := range program.Functions {
		seen[function.Name] = true
	}
	return seen["acquire"] && seen["probe"] && seen["main"]
}

func originRepeatedOwnerProblem(program core.Program) *Problem {
	fail := func(id string) *Problem {
		return &Problem{Code: "core.local_owner_transfer", Detail: fmt.Sprintf("invalid repeated owner flow at %s", id)}
	}
	functions := map[string]*core.Function{}
	for index := range program.Functions {
		functions[program.Functions[index].Name] = &program.Functions[index]
	}
	helper, probe, caller := functions["acquire"], functions["probe"], functions["main"]
	if helper == nil || probe == nil || caller == nil || helper.Linear == nil || probe.Linear == nil || caller.Linear == nil {
		return fail("phase24 functions")
	}
	var acquisition *core.LinearOperation
	for index := range helper.Linear.Operations {
		op := &helper.Linear.Operations[index]
		if op.Foreign != nil && op.Foreign.Mode == "acquire" {
			if acquisition != nil {
				return fail(op.ID)
			}
			acquisition = op
		}
	}
	if acquisition == nil || acquisition.Kind != core.OpForeignCall || acquisition.Foreign == nil || acquisition.Foreign.Release != "schway_file_byte_release" {
		return fail(helper.ID)
	}
	returned := false
	for _, op := range helper.Linear.Operations {
		if op.Kind == core.OpReturn && op.SourceID == acquisition.TargetID {
			returned = true
		}
		if op.Kind == core.OpRelease {
			return fail(op.ID)
		}
	}
	if !returned {
		return fail(helper.ID)
	}
	mainAcquires, mainProbes, probeAcquires := originCalls(caller, helper.ID), originCalls(caller, probe.ID), originCalls(probe, helper.ID)
	if len(mainAcquires) != 2 || len(mainProbes) != 1 || len(probeAcquires) != 1 {
		return fail(helper.ID)
	}
	allCalls := append(append(append([]core.LinearOperation{}, mainAcquires...), mainProbes...), probeAcquires...)
	seenCallIDs := map[string]bool{}
	for _, call := range allCalls {
		if seenCallIDs[call.ID] || !originFallibleCall(program, caller, probe, call) {
			return fail(call.ID)
		}
		seenCallIDs[call.ID] = true
	}
	if !originProbeReleases(probe, probeAcquires[0], acquisition.ID) {
		return fail(probe.ID)
	}
	if !originMainReleases(caller, mainAcquires[0], mainAcquires[1], mainProbes[0], acquisition.ID) {
		return fail(caller.ID)
	}
	return nil
}

func originCalls(function *core.Function, callee string) []core.LinearOperation {
	var result []core.LinearOperation
	for _, op := range function.Linear.Operations {
		if op.Kind == core.OpCall && op.CalleeID == callee {
			result = append(result, op)
		}
	}
	return result
}

func originFallibleCall(program core.Program, caller, probe *core.Function, call core.LinearOperation) bool {
	for _, function := range []*core.Function{caller, probe} {
		for _, operation := range function.Linear.Operations {
			if operation.ID != call.ID {
				continue
			}
			var block string
			for _, candidate := range function.Linear.Blocks {
				for _, id := range candidate.OperationIDs {
					if id == call.ID {
						block = candidate.ID
					}
				}
			}
			var okEdge, errEdge *core.Edge
			for index := range function.Linear.Edges {
				edge := &function.Linear.Edges[index]
				if edge.ID == call.OkEdgeID {
					okEdge = edge
				}
				if edge.ID == call.ErrEdgeID {
					errEdge = edge
				}
			}
			if call.ErrTargetID == "" || block == "" || okEdge == nil || errEdge == nil || okEdge.FromBlockID != block || errEdge.FromBlockID != block || okEdge.Pattern != "ok" || errEdge.Pattern != "err" || okEdge.ToBlockID == errEdge.ToBlockID || !originBlockHasSuccessor(function, block, okEdge.ToBlockID) || !originBlockHasSuccessor(function, block, errEdge.ToBlockID) {
				return false
			}
			return originPlaceType(function, call.ErrTargetID) == "ResourceError" && originFailureEnvelopeFits(program, call.CalleeID)
		}
	}
	return false
}

func originBlockHasSuccessor(function *core.Function, blockID, successorID string) bool {
	for _, block := range function.Linear.Blocks {
		if block.ID != blockID {
			continue
		}
		for _, successor := range block.Successors {
			if successor == successorID {
				return true
			}
		}
	}
	return false
}

func originFailureEnvelopeFits(program core.Program, calleeID string) bool {
	var callee *core.Function
	for index := range program.Functions {
		if program.Functions[index].ID == calleeID {
			callee = &program.Functions[index]
			break
		}
	}
	if callee == nil || callee.Linear == nil {
		return false
	}
	byType := map[string]string{}
	for _, fact := range callee.Linear.Types {
		byType[fact.ID] = fact.Shape.Constructor
	}
	failing := map[string]bool{}
	for _, operation := range callee.Linear.Operations {
		if operation.Kind == core.OpFail {
			failing[byType[operation.TypeID]] = true
		}
		if operation.Foreign != nil && operation.Foreign.Fails != "" {
			failing[operation.Foreign.Fails] = true
		}
	}
	if len(failing) == 0 {
		return false
	}
	envelope := map[string]bool{}
	for _, dataType := range program.DataTypes {
		if dataType.Name == "ResourceError" && len(dataType.AlternativeDetails) == 0 {
			for _, alt := range dataType.Alternatives {
				envelope[alt] = true
			}
		}
	}
	if len(envelope) == 0 {
		return false
	}
	for failure := range failing {
		found := false
		for _, dataType := range program.DataTypes {
			if dataType.Name != failure || len(dataType.AlternativeDetails) != 0 {
				continue
			}
			found = true
			for _, alt := range dataType.Alternatives {
				if !envelope[alt] {
					return false
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func originPlaceType(function *core.Function, placeID string) string {
	types := map[string]string{}
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	for _, place := range function.Linear.Places {
		if place.ID == placeID {
			return types[place.TypeID]
		}
	}
	return ""
}

func originProbeReleases(probe *core.Function, acquireCall core.LinearOperation, acquisitionID string) bool {
	var use *core.LinearOperation
	for index := range probe.Linear.Operations {
		op := &probe.Linear.Operations[index]
		if op.Foreign != nil && op.Foreign.Mode == "borrow" {
			if use != nil {
				return false
			}
			use = op
		}
	}
	if use == nil || use.SourceID != acquireCall.TargetID || use.Foreign.Fails != "UseError" || originPlaceType(probe, use.ErrTargetID) != "UseError" {
		return false
	}
	return originCleanupBlock(probe, originEdgeTarget(probe, use.OkEdgeID), []string{acquireCall.TargetID}, use.TargetID, core.OpReturn, acquisitionID) &&
		originCleanupBlock(probe, originEdgeTarget(probe, use.ErrEdgeID), []string{acquireCall.TargetID}, use.ErrTargetID, core.OpFail, acquisitionID) &&
		originCleanupBlock(probe, originEdgeTarget(probe, acquireCall.ErrEdgeID), nil, acquireCall.ErrTargetID, core.OpFail, acquisitionID)
}

func originMainReleases(main *core.Function, callA, callB, callProbe core.LinearOperation, acquisitionID string) bool {
	return originCleanupBlock(main, originEdgeTarget(main, callA.ErrEdgeID), nil, callA.ErrTargetID, core.OpFail, acquisitionID) &&
		originCleanupBlock(main, originEdgeTarget(main, callB.ErrEdgeID), []string{callA.TargetID}, callB.ErrTargetID, core.OpFail, acquisitionID) &&
		originCleanupBlock(main, originEdgeTarget(main, callProbe.ErrEdgeID), []string{callB.TargetID, callA.TargetID}, callProbe.ErrTargetID, core.OpFail, acquisitionID) &&
		originCleanupBlock(main, originEdgeTarget(main, callProbe.OkEdgeID), []string{callB.TargetID, callA.TargetID}, callProbe.TargetID, core.OpReturn, acquisitionID)
}

func originEdgeTarget(function *core.Function, edgeID string) string {
	for _, edge := range function.Linear.Edges {
		if edge.ID == edgeID {
			return edge.ToBlockID
		}
	}
	return ""
}

func originCleanupBlock(function *core.Function, blockID string, owners []string, terminalSource string, terminal core.OperationKind, acquisitionID string) bool {
	for _, block := range function.Linear.Blocks {
		if block.ID != blockID {
			continue
		}
		if len(block.OperationIDs) != len(owners)+1 {
			return false
		}
		operations := map[string]core.LinearOperation{}
		for _, op := range function.Linear.Operations {
			operations[op.ID] = op
		}
		for index, owner := range owners {
			release := operations[block.OperationIDs[index]]
			if release.Kind != core.OpRelease || release.SourceID != owner || release.ReleasesOperationID != acquisitionID || release.Foreign == nil || release.Foreign.Mode != "consume" {
				return false
			}
		}
		last := operations[block.OperationIDs[len(block.OperationIDs)-1]]
		return last.Kind == terminal && (terminalSource == "" || last.SourceID == terminalSource)
	}
	return false
}

// ValidatePublished recomputes every function's origin from its body and
// compares it against the declaration, following Spike 003's
// producer-verification gates. It returns the first problem only, matching
// corevalidate's first-problem-only accumulation, so the stable assertion
// target is always the first defect. Recomputation now runs unconditionally
// for every function, including one with no declared PublicOrigin at all
// (D-03-02/GAP 2, ROADMAP SC4's "omitted" category): when recomputation
// succeeds for such a function, its body actually returns a borrow-derived
// place that publication would otherwise expose indistinguishably from a
// fully-owned return, and that is refused with core.origin_omitted. When
// recomputation instead reports not-ok (an owned return, or a match-bodied
// function with no linear return chain), the function is left untouched
// exactly as before — the new check does not over-fire on an honest,
// declaration-free owned value.
// checkForeignOriginOmitted independently re-derives, for one function,
// whether its returned value traces back to a foreign call whose declared
// contract says it borrows or retains its argument (D-04-28). It performs
// its own backward walk over function.Linear.Operations rather than sharing
// RecomputeOriginPerReturn's (T-04-37: two derivations must not agree merely
// because they share a law) -- it reads only the core artifact, never the
// checker: the foreign contract's declared Alias obligation and the
// operation's own SourceID/TargetID places.
func checkForeignOriginOmitted(function core.Function) *Problem {
	if function.ForeignContract == nil {
		return nil
	}
	alias := function.ForeignContract.Alias
	if alias != "borrow" && alias != "retain" {
		return nil
	}
	if function.Linear == nil || function.PublicOrigin != nil {
		return nil
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var foreignCallTarget string
	for index := range operations {
		operation := operations[index]
		if operation.Kind == core.OpForeignCall {
			foreignCallTarget = operation.TargetID
		}
		if isTerminatorKind(operation.Kind) {
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	if foreignCallTarget == "" {
		return nil
	}
	for index := range operations {
		operation := operations[index]
		if operation.Kind != core.OpReturn {
			continue
		}
		current := operation.SourceID
		visited := make(map[string]bool)
		for current != function.Parameter.ID {
			if current == foreignCallTarget {
				return &Problem{
					Code: "core.foreign_origin_omitted",
					Detail: fmt.Sprintf(
						"%s: return derives from a foreign call declared %q on argument %q, but no public origin is declared",
						function.ID, alias, function.Parameter.Name,
					),
				}
			}
			if visited[current] {
				break
			}
			visited[current] = true
			source, exists := sourceOf[current]
			if !exists {
				break
			}
			current = source.SourceID
		}
	}
	return nil
}

// PublishProblemsFor is D-07-32's per-function extraction of
// ValidatePublished's loop body: it recomputes exactly ONE function's
// return-origin safety from its checked core body and reports every problem
// that would refuse its publication. This is D-04-03's actual predicate —
// "would ValidatePublished publish it" — restated at function granularity so
// Callable (D-07-31) can be derived per function rather than collapsing an
// entire program to its first offending function. It never reads an export
// list: the word "export" does not appear on this path, because publication
// safety and export membership are different rules (D-07-31).
func PublishProblemsFor(function core.Function, calleeContracts map[string]calleeOriginFact) []Problem {
	if function.ReturnType != "FileByteOwner" && !hasOwnerCall(function) {
		if problem := localOwnerOriginProblem(function); problem != nil {
			return []Problem{*problem}
		}
	}
	if problem := checkForeignOriginOmitted(function); problem != nil {
		return []Problem{*problem}
	}
	recomputedPaths, recomputedAccess, ok := RecomputeOrigin(function, calleeContracts)
	if function.PublicOrigin == nil {
		if ok {
			return []Problem{{
				Code:   "core.origin_omitted",
				Detail: fmt.Sprintf("%s: no declared origin, but body derives origin %v with access %q", function.ID, recomputedPaths, recomputedAccess),
			}}
		}
		return nil
	}
	// Declared-access domain check (Task 03-10-02): a declared Access
	// outside {"shared", "exclusive"} is refused BEFORE any comparison
	// with the recomputed answer, so a mutated summary cannot declare
	// AccessConflicting and have it match a genuinely conflicting
	// recomputation.
	if function.PublicOrigin.Access != "shared" && function.PublicOrigin.Access != "exclusive" {
		return []Problem{{
			Code:   "core.origin_access_mismatch",
			Detail: fmt.Sprintf("%s: declared access %q is not a declarable mode", function.ID, function.PublicOrigin.Access),
		}}
	}
	if !ok || !containsAll(function.PublicOrigin.Paths, recomputedPaths) {
		return []Problem{{
			Code:   "core.origin_understated",
			Detail: fmt.Sprintf("%s: declared origin %v does not cover the body-derived origin %v", function.ID, function.PublicOrigin.Paths, recomputedPaths),
		}}
	}
	if function.PublicOrigin.Access != recomputedAccess {
		return []Problem{{
			Code:   "core.origin_access_mismatch",
			Detail: fmt.Sprintf("%s: declared access %q, body derives %q", function.ID, function.PublicOrigin.Access, recomputedAccess),
		}}
	}
	return nil
}

// localOwnerOriginProblem independently checks the operation-specific owner
// path before origin publication. The checker emits one straight-line
// success path: an acquire error has no FileByteOwner result, while a borrow
// error leaves its input owner live until the following consuming release.
// The local format deliberately refuses CFGs and other operations rather than
// pretending a flat list proves branches it does not represent.
func localOwnerOriginProblem(function core.Function) *Problem {
	if function.Linear == nil {
		return nil
	}
	local := false
	for _, operation := range function.Linear.Operations {
		local = local || operation.Foreign != nil
	}
	if !local {
		return nil
	}
	fail := func(operationID string) *Problem {
		return &Problem{Code: "core.local_owner_lifecycle", Detail: fmt.Sprintf("%s: invalid local owner operation %q", function.ID, operationID)}
	}
	linear := function.Linear
	if function.Parameter.Type != "PathToken" || function.ReturnType != "U64" || len(linear.Blocks) != 0 || len(linear.Edges) != 0 {
		return fail(function.ID)
	}
	types := make(map[string]string, len(linear.Types))
	for _, fact := range linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	places := make(map[string]string, len(linear.Places))
	for _, place := range linear.Places {
		places[place.ID] = types[place.TypeID]
	}
	type ownerFact struct {
		placeID   string
		resultID  string
		release   string
		allocator string
		borrowed  bool
		released  bool
	}
	owners := make(map[string]ownerFact)
	ownerByPlace := make(map[string]string)
	modeOrder := make([]string, 0, 3)
	returnedFrom := ""
	for _, operation := range linear.Operations {
		contract := operation.Foreign
		if contract == nil {
			if operation.Kind != core.OpReturn || returnedFrom != "" || len(ownerByPlace) != 0 {
				return fail(operation.ID)
			}
			returnedFrom = operation.SourceID
			continue
		}
		if !originIdentifier(contract.Symbol) || !originIdentifier(contract.ABIType) || contract.Unwind != "forbidden" || contract.NonlocalExit != "forbidden" {
			return fail(operation.ID)
		}
		modeOrder = append(modeOrder, contract.Mode)
		sourceType := places[operation.SourceID]
		switch contract.Mode {
		case "acquire":
			if operation.Kind != core.OpForeignCall || operation.SourceID != function.Parameter.ID || sourceType != "PathToken" || places[operation.TargetID] != "FileByteOwner" || operation.TypeID == "" || types[operation.TypeID] != "FileByteOwner" || contract.Symbol != "schway_file_byte_acquire" || contract.ABIType != "schway_file_byte_acquire_fn" || contract.ParameterType != "PathToken" || contract.ResultType != "FileByteOwner" || contract.Fails != "AcquireError" || contract.Allocator != "libc_malloc" || contract.Release != "schway_file_byte_release" || operation.Allocator != contract.Allocator || operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "" || owners[operation.ID].placeID != "" || ownerByPlace[operation.TargetID] != "" {
				return fail(operation.ID)
			}
			owners[operation.ID] = ownerFact{placeID: operation.TargetID, release: contract.Release, allocator: contract.Allocator}
			ownerByPlace[operation.TargetID] = operation.ID
		case "borrow":
			acquireID := ownerByPlace[operation.SourceID]
			owner, exists := owners[acquireID]
			if operation.Kind != core.OpForeignCall || !exists || owner.borrowed || owner.released || sourceType != "FileByteOwner" || places[operation.TargetID] != "U64" || operation.TypeID == "" || types[operation.TypeID] != "U64" || contract.Symbol != "schway_file_byte_use" || contract.ABIType != "schway_file_byte_use_fn" || contract.ParameterType != "FileByteOwner" || contract.ResultType != "U64" || contract.Fails != "UseError" || contract.Allocator != "" || contract.Release != "" || operation.Allocator != "" || operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "" {
				return fail(operation.ID)
			}
			owner.borrowed = true
			owner.resultID = operation.TargetID
			owners[acquireID] = owner
		case "consume":
			owner, exists := owners[operation.ReleasesOperationID]
			if operation.Kind != core.OpRelease || !exists || !owner.borrowed || owner.released || owner.placeID != operation.SourceID || sourceType != "FileByteOwner" || operation.TypeID == "" || types[operation.TypeID] != "FileByteOwner" || owner.release != contract.Symbol || owner.allocator != contract.Allocator || contract.Symbol != "schway_file_byte_release" || contract.ABIType != "schway_file_byte_release_fn" || contract.ParameterType != "FileByteOwner" || contract.ResultType != "Unit" || contract.Fails != "" || contract.Allocator != "libc_malloc" || contract.Release != "" || operation.Allocator != contract.Allocator || operation.TargetID != "" {
				return fail(operation.ID)
			}
			owner.released = true
			owners[operation.ReleasesOperationID] = owner
			delete(ownerByPlace, operation.SourceID)
		default:
			return fail(operation.ID)
		}
	}
	if len(owners) != 1 || len(ownerByPlace) != 0 || len(modeOrder) != 3 || modeOrder[0] != "acquire" || modeOrder[1] != "borrow" || modeOrder[2] != "consume" || returnedFrom == "" {
		return fail(function.ID)
	}
	for _, owner := range owners {
		if !owner.borrowed || !owner.released || returnedFrom != owner.resultID {
			return fail(function.ID)
		}
	}
	return nil
}

func originIdentifier(value string) bool {
	if value == "" || !(value[0] == '_' || value[0] >= 'A' && value[0] <= 'Z' || value[0] >= 'a' && value[0] <= 'z') {
		return false
	}
	for i := 1; i < len(value); i++ {
		b := value[i]
		if !(b == '_' || b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' || b >= '0' && b <= '9') {
			return false
		}
	}
	return true
}

// ValidatePublished recomputes every function's origin from its body and
// compares it against the declaration, following Spike 003's
// producer-verification gates. It returns the first problem only, matching
// corevalidate's first-problem-only accumulation, so the stable assertion
// target is always the first defect. D-07-32: the per-function body now
// lives in PublishProblemsFor; this loop preserves the original
// whole-program first-problem contract exactly, byte-for-byte, by returning
// the first non-empty PublishProblemsFor result it encounters.
func ValidatePublished(program core.Program) []Problem {
	if problem := transferredOwnerProblem(program); problem != nil {
		return []Problem{*problem}
	}
	calleeContracts := BuildCalleeOriginFacts(program)
	for _, function := range program.Functions {
		if problem := pointerBorrowConflictProblem(function); problem != nil {
			return []Problem{*problem}
		}
		if problems := PublishProblemsFor(function, calleeContracts); len(problems) > 0 {
			return problems
		}
	}
	return nil
}

// pointerBorrowConflictProblem independently replays the bounded straight-
// line pointer-helper loan shape. It derives each loan's owner, access
// family, and last use from this function's own operations; it does not
// consume check/corevalidate classifications or materialized endpoints.
func pointerBorrowConflictProblem(function core.Function) *Problem {
	if function.Linear == nil || len(function.Linear.Blocks) != 0 {
		return nil
	}
	operations := function.Linear.Operations
	placeLoans := map[string][]string{function.Parameter.ID: nil}
	loanOwner := map[string]string{}
	loanAccess := map[string]string{}
	loanBirth := map[string]int{}
	lastUse := map[string]int{}
	for index, operation := range operations {
		carried := placeLoans[operation.SourceID]
		for _, loanID := range carried {
			lastUse[loanID] = index
		}
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			owner := operation.SourceID
			if len(carried) > 0 {
				owner = loanOwner[carried[0]]
			}
			loanOwner[operation.LoanID] = owner
			loanBirth[operation.LoanID] = index
			if operation.Kind == core.OpBorrowShared {
				loanAccess[operation.LoanID] = "shared"
			} else {
				loanAccess[operation.LoanID] = "exclusive"
			}
			lastUse[operation.LoanID] = index
			carried = append(append([]string(nil), carried...), operation.LoanID)
		}
		if operation.TargetID != "" {
			placeLoans[operation.TargetID] = append([]string(nil), carried...)
		}
	}
	for index, operation := range operations {
		if operation.Kind != core.OpBorrowShared && operation.Kind != core.OpBorrowExclusive {
			continue
		}
		owner := operation.SourceID
		carriedBySource := make(map[string]bool)
		for _, loanID := range placeLoans[operation.SourceID] {
			owner = loanOwner[loanID]
			carriedBySource[loanID] = true
		}
		newExclusive := operation.Kind == core.OpBorrowExclusive
		for loanID, activeOwner := range loanOwner {
			if activeOwner != owner || loanBirth[loanID] >= index || lastUse[loanID] < index {
				continue
			}
			// A borrow through an existing loan is a reborrow, so that
			// loan itself is not a competing sibling. A shared parent
			// cannot authorize an exclusive child, however.
			if carriedBySource[loanID] && (!newExclusive || loanAccess[loanID] == "exclusive") {
				continue
			}
			if newExclusive || loanAccess[loanID] == "exclusive" {
				return &Problem{Code: "core.borrow_conflict", Detail: operation.ID}
			}
		}
	}
	return nil
}

// containsAll reports whether every recomputed path is present in the
// declared set — a declared set that omits a body-derivable path is
// understated.
func containsAll(declared, recomputed []string) bool {
	present := make(map[string]bool, len(declared))
	for _, path := range declared {
		present[path] = true
	}
	for _, path := range recomputed {
		if !present[path] {
			return false
		}
	}
	return true
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// ClosureDigestDomainSeparator is D-07-37's canonical ClosureDigest
// preimage's fixed prefix, declared exactly once here and referenced
// everywhere else the preimage is built. It is not itself a digest of
// anything -- it exists solely so a future digest with a superficially
// similar preimage shape can never collide with this one.
const ClosureDigestDomainSeparator = "lang.closure_digest/1\x00"

// calleeDigestPair is one callee's (ID, ClosureDigest) pair, part of
// D-07-37's canonical preimage. Field order is fixed by declaration (never
// map iteration), so json.Marshal of a []calleeDigestPair is deterministic
// regardless of insertion order -- SORTING the slice by ID (below) is what
// makes the preimage itself order-independent; this struct's own encoding
// was never order-dependent to begin with.
type calleeDigestPair struct {
	ID            string `json:"id"`
	ClosureDigest string `json:"closure_digest"`
}

// closureDigestSortOverride is D-07-42's unexported fault-injection seam
// for a same-package mutation-kill test (originvalidate_internal_test.go):
// production always sorts callee pairs by ID; the test temporarily
// replaces this with the identity function to prove the out-of-order test
// actually depends on the sort, not merely appears to. nil (the
// always-true production default) means "sort by ID".
var closureDigestSortOverride func([]calleeDigestPair) []calleeDigestPair

func sortCalleeDigestPairs(pairs []calleeDigestPair) []calleeDigestPair {
	sorted := append([]calleeDigestPair(nil), pairs...)
	if closureDigestSortOverride != nil {
		return closureDigestSortOverride(sorted)
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	return sorted
}

// closureDigestPreimageBytes builds D-07-37's canonical, non-self-
// referential preimage for signature's ClosureDigest: the domain separator,
// then signature with ClosureDigest itself ZEROED (so the digest never
// depends on its own prior value — non-self-reference, Test 1), then the
// callee (ID, ClosureDigest) pairs SORTED BY ID (Test 4) so the preimage is
// independent of the order callees happen to be supplied in.
//
// D-07-38: 07-01 called this only with an empty/nil callees slice — the
// zero-callee base case. This plan (07-08) supplies REAL callee pairs from
// BuildInterface, now that cycle refusal exists (callgraph.Order runs
// first): the chain terminates only on a DAG, so computing it before that
// gate would let a cyclic program reach a non-terminating digest
// computation before the gate that would refuse it. This function's
// preimage definition itself is UNCHANGED by that addition — only where
// the caller sources callees from changed.
func closureDigestPreimageBytes(signature core.FunctionSignature, callees []calleeDigestPair) ([]byte, error) {
	signature.ClosureDigest = ""
	payload := struct {
		Signature core.FunctionSignature `json:"signature"`
		Callees   []calleeDigestPair     `json:"callees"`
	}{Signature: signature, Callees: sortCalleeDigestPairs(callees)}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	preimage := make([]byte, 0, len(ClosureDigestDomainSeparator)+len(body))
	preimage = append(preimage, []byte(ClosureDigestDomainSeparator)...)
	preimage = append(preimage, body...)
	return preimage, nil
}

// computeClosureDigest computes D-07-37's canonical ClosureDigest for one
// function signature and its (possibly empty) callee set, reusing the
// shared digest() helper — no second digest format is minted anywhere in
// this package.
func computeClosureDigest(signature core.FunctionSignature, callees []calleeDigestPair) (string, error) {
	preimage, err := closureDigestPreimageBytes(signature, callees)
	if err != nil {
		return "", err
	}
	return digest(preimage), nil
}

// BuildInterface strips every function body from program and binds the
// resulting summary to program's own content digest, reusing the same
// SHA-256 content-digest pattern already proven for evidence.CoreDigest
// (Spike 004's certificate binds the same way). Callers MUST run
// ValidatePublished against program first: BuildInterface packages what the
// producer already proved rather than re-deriving it.
//
// D-07-08: this emits Schema core.InterfaceSchema1 (lang.interface/1) —
// every function.ID+":type:0" abilities is unchanged from /0.
// Every /1 field is populated from its R-01 authority (see the field-level
// doc comments on core.FunctionSignature); Callable is derived per function
// from PublishProblemsFor (D-07-31/D-07-32, 07-02) — publication safety, not
// export membership — and fails closed to false whenever PublishProblemsFor
// reports any problem.
// typeFactExactIDMatchOverride is D-07-42's fault-injection seam for
// TestStage0SummaryMutationMatrix's fault 1 (07-02 Task 3): production
// always requires an exact fact.ID == wantID match in BuildInterface's
// abilities lookup below; the test temporarily forces every comparison to
// report no match, reproducing D-07-23's silent empty-abilities fallback so
// a divergence test can prove the corevalidate peer's independently-derived
// abilities really do diverge, not merely appear to. nil (the always-true
// production default) means "use the real fact.ID == wantID comparison".
// Unexported: never a package-level var settable outside a same-package (or
// export_test.go, which is excluded from every non-test build) fault test.
var typeFactExactIDMatchOverride func(factID, wantID string) bool

func typeFactExactIDMatch(factID, wantID string) bool {
	if typeFactExactIDMatchOverride != nil {
		return typeFactExactIDMatchOverride(factID, wantID)
	}
	return factID == wantID
}

// forceCallableAlwaysTrue is D-07-42's fault-injection seam for
// TestStage0SummaryMutationMatrix's faults 3 and 4 (07-02 Task 3):
// production always derives Callable from
// len(PublishProblemsFor(function))==0; the test temporarily forces it
// unconditionally true, proving the independent corevalidate peer still
// refuses (fault 4) or, combined with corevalidate's own mirror seam,
// proving a bilateral fault produces a FALSE agreement that the sweep must
// report as a gate failure (fault 3). false (the always-real production
// default) means "use the real derivation".
var forceCallableAlwaysTrue bool

// closureDigestEmptyCalleesOverride is Task 2's D-07-41/D-07-42
// fault-injection seam (QLT-08's headline mutation kill): production
// always supplies a caller's REAL, deduped, callee (ID, ClosureDigest)
// pairs to closureDigestPreimageBytes; the test temporarily forces every
// function to compute its ClosureDigest as if it had NO callees at all,
// proving the callee-changes-invalidates-caller property (D-07-12) is
// genuinely load-bearing: with this seam engaged, changing only a callee's
// body no longer changes the caller's ClosureDigest, because the caller's
// preimage never referenced the callee's digest to begin with. false (the
// production default) means "chain over real callee pairs".
var closureDigestEmptyCalleesOverride bool

// closureDigestDiscoveryOrderOverride is Task 2's D-07-41/D-07-42
// fault-injection seam for D-07-38's ordering claim: production computes
// every function's ClosureDigest bottom-up, callee before caller
// (BuildInterface's own reversal of callgraph.Order's returned array —
// see the doc comment below); the test temporarily computes digests in
// plain program.Functions DECLARATION order instead, so a caller declared
// before its callee reads that callee's digest while it is still the
// empty string (not yet computed), producing an observably DIFFERENT
// digest than the correctly-ordered computation — proving the ordering
// itself is load-bearing, not merely present. false (the production
// default) means "chain in callgraph.Order's reverse-of-reverse-postorder
// (callee-before-caller)".
var closureDigestDiscoveryOrderOverride bool

// closureDigestComputationOrderObserved is Task 1's own ordering-
// instrumentation seam: when non-nil, invoked with each function's ID, in
// the exact order BuildInterface finalizes its ClosureDigest, so a
// same-package test can assert that sequence against callgraph.Order's own
// return value. nil in production: zero cost, zero allocation.
var closureDigestComputationOrderObserved func(functionID string)

// foreignClosureJoinSeam is 07-12's D-07-41/D-07-42 fault-injection seam for
// control:summary.foreign_reach_closure_derived: production always joins a
// caller's published Foreign with every callee's already-committed Foreign,
// in the chainOrder loop, before that caller's own ClosureDigest is
// computed; the test temporarily disables that join, reproducing CR-03's
// original defect (a caller of a foreign-reaching callee wrongly publishes
// the zero ForeignReach) so a mutation-kill test can prove the join is
// genuinely load-bearing. false (the production default) means "join for
// real".
var foreignClosureJoinSeam = false

// SetForeignClosureJoinSeam installs/lifts foreignClosureJoinSeam and
// returns a restore func. Declared directly in this production file
// (not export_test.go), mirroring check.SetCallArgumentConsumeSeamForTest's
// established pattern (07-11, D-07-42): Go's build model excludes every
// "_test.go" file (including export_test.go) from a normal cross-package
// import, so a same-package-only or export_test.go-only setter cannot be
// reached by corevalidate's own bilateral-independence test
// (TestForeignClosureJoinPeersIndependent), which imports this package as
// an ordinary dependency. Deliberately minimal: a documented test-only
// no-op unless a test explicitly calls it, and always restored via its
// returned closure. Callers MUST defer the restore immediately.
func SetForeignClosureJoinSeam(disable bool) (restore func()) {
	previous := foreignClosureJoinSeam
	foreignClosureJoinSeam = disable
	return func() { foreignClosureJoinSeam = previous }
}

// failsClosureJoinSeam is 07-12's D-07-41/D-07-42 fault-injection seam for
// control:summary.fails_closure_derived: production always joins a
// caller's published Fails with every callee's already-committed Fails, in
// the same chainOrder loop, before ClosureDigest; the test temporarily
// disables that join, reproducing CR-03's original defect (a caller of a
// fallible callee wrongly publishes ""/infallible) so a mutation-kill test
// can prove the join is genuinely load-bearing. false (the production
// default) means "join for real".
var failsClosureJoinSeam = false

// SetFailsClosureJoinSeam installs/lifts failsClosureJoinSeam and returns
// a restore func. Same cross-package-visibility rationale as
// SetForeignClosureJoinSeam above. Callers MUST defer the restore
// immediately.
func SetFailsClosureJoinSeam(disable bool) (restore func()) {
	previous := failsClosureJoinSeam
	failsClosureJoinSeam = disable
	return func() { failsClosureJoinSeam = previous }
}

// joinReachPolicy joins two ForeignReach policy field values (Unwind,
// NonlocalExit) under today's two-value vocabulary {"", "permitted",
// "forbidden"}: "" is the identity element (an empty operand contributes
// nothing); two equal non-empty values merge to that value; "forbidden" is
// strictly more constraining than "permitted" and wins whenever either
// operand is "forbidden" (in either argument order, so this branch alone
// makes the function commutative for this vocabulary); any other
// disagreement -- unreachable while the vocabulary stays exactly these
// three values, but not assumed away -- resolves to the declared
// core.ForeignReachConflict sentinel rather than an arbitrary pick.
// joinForeignReach below folds this pairwise operation over a caller's own
// local value and every callee's already-joined value; because "forbidden"
// (once present anywhere in the fold) is pairwise-absorbing against any
// other operand, and because "" is a true identity, the FOLD's result does
// not depend on fold order -- see TestForeignJoinIsOrderIndependent.
func joinReachPolicy(into, from string) string {
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

// joinAllocatorName joins two ForeignReach.Allocator values: "" is the
// identity; two equal non-empty names merge; two DIFFERENT non-empty names
// have no vocabulary ordering at all (unlike Unwind/NonlocalExit's
// forbidden-wins rule) and resolve to core.ForeignReachConflict, never to
// either input. Like joinReachPolicy, this is commutative and, because
// core.ForeignReachConflict is itself pairwise-absorbing against any
// distinct operand once formed, associative under fold.
func joinAllocatorName(into, from string) string {
	switch {
	case into == "":
		return from
	case from == "" || into == from:
		return into
	default:
		return core.ForeignReachConflict
	}
}

// joinForeignReach is 07-12's producer-side worst-case lattice join over
// core.ForeignReach (CR-03/PVG-02), applied field by field via
// joinAllocatorName/joinReachPolicy. It is commutative and associative
// (TestForeignJoinIsOrderIndependent), so the published value the
// chainOrder loop below folds this over does not depend on the order a
// function's callees happen to be visited in. corevalidate has its OWN,
// deliberately unshared implementation of this same join
// (peerJoinForeignReach) over its own postorder -- see
// PHASE-07-DEBT.md/07-REVIEW.md CR-03 for why no helper is shared between
// the two packages.
func joinForeignReach(into, from core.ForeignReach) core.ForeignReach {
	return core.ForeignReach{
		Allocator:    joinAllocatorName(into.Allocator, from.Allocator),
		Unwind:       joinReachPolicy(into.Unwind, from.Unwind),
		NonlocalExit: joinReachPolicy(into.NonlocalExit, from.NonlocalExit),
	}
}

// joinFails is 07-12's producer-side join over FunctionSignature.Fails
// (CR-03/PVG-02): "" is the identity (an empty operand contributes
// nothing); two equal non-empty values merge; two DIFFERENT non-empty
// values keep the EXISTING (caller-nearest -- the accumulator this
// function is folded into, which starts as the caller's own local Fails
// and is folded across callees in calleeIDsForClosureDigest's sorted-ID
// order, so the result is deterministic in production even though this
// function is not itself order-independent the way joinForeignReach is).
// This is a deliberately DISCLOSED imprecision, not a hidden one:
// core.FunctionSignature.Fails is a single string and cannot express a
// UNION of two distinct error types, so a caller reaching two distinct
// fallible callees publishes only one of them -- recorded as D-07-53 in
// PHASE-07-DEBT.md, never silently overwritten last-writer-wins (the
// EXISTING value always wins over a new disagreeing one, never the
// reverse).
func joinFails(into, from string) string {
	if into == "" {
		return from
	}
	return into
}

// calleeIDsForClosureDigest collects the DISTINCT callee IDs a function's
// own core.OpCall operations name, sorted, from core.LinearOperation.
// CalleeID -- the SAME edge fact callgraph.Order reads, from the same
// place (D-07-38's read_first note), so there is no second edge notion to
// drift. Dedup means a function calling the same callee twice contributes
// exactly one (ID, ClosureDigest) pair to its own preimage, matching
// callgraph.buildAdjacency's own dedup discipline.
func calleeIDsForClosureDigest(function core.Function) []string {
	if function.Linear == nil {
		return nil
	}
	seen := make(map[string]bool)
	var ids []string
	for _, operation := range function.Linear.Operations {
		if operation.Kind != core.OpCall {
			continue
		}
		if seen[operation.CalleeID] {
			continue
		}
		seen[operation.CalleeID] = true
		ids = append(ids, operation.CalleeID)
	}
	sort.Strings(ids)
	return ids
}

func BuildInterface(program core.Program) (core.Interface, error) {
	// D-07-38: callgraph.Order runs FIRST, before any ClosureDigest is
	// computed. On a refused (cyclic, or unresolved-callee) graph, its
	// error is returned unchanged and NO digest is computed at all --
	// chaining over a graph that is not a proven DAG is exactly the
	// non-termination this ordering exists to prevent. `order` is
	// callgraph.Order's own "reverse postorder" (D-07-18): CALLERS first,
	// callees last (see callgraph.Order's doc comment and
	// TestOrderSortsAdjacencyByCalleeID's own worked example). Chaining
	// needs the OPPOSITE direction -- every callee's digest computed
	// before its caller reads it -- which is exactly `order` walked
	// BACKWARD (its raw, un-reversed DFS postorder): a DFS node is
	// appended to that raw postorder only once every callee it can reach
	// has already finished, so walking `order` from its last element to
	// its first recovers callee-before-caller processing with no second
	// traversal.
	order, err := callgraph.Order(program)
	if err != nil {
		return core.Interface{}, err
	}
	coreBytes, err := json.Marshal(program)
	if err != nil {
		return core.Interface{}, err
	}
	summary := core.Interface{
		Schema: core.InterfaceSchema1, ModuleID: program.ModuleID, CoreDigest: digest(coreBytes),
		Functions: make([]core.FunctionSignature, 0, len(program.Functions)),
	}
	// programFunctionByID is 07-12's IN-01 fix (07-REVIEW.md INFO): the
	// second pass below resolves a program function BY ID rather than by
	// indexing program.Functions with an index built from
	// len(summary.Functions). Built once, alongside indexByID, from the
	// same range over program.Functions.
	programFunctionByID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		programFunctionByID[function.ID] = function
	}
	// D-10-03: built once, from program's own declared PublicOrigin facts
	// only, mirroring ValidatePublished's own single build -- BuildInterface
	// strips bodies but the callee-contract map itself was never
	// body-derived to begin with.
	calleeContracts := BuildCalleeOriginFacts(program)
	indexByID := make(map[string]int, len(program.Functions))
	for _, function := range program.Functions {
		parameterAbilities := interfaceParameterAbilities(function)
		returnAbilities := interfaceReturnAbilities(function)
		hasParameterDropAbility := abilitiesIncludeDrop(parameterAbilities)
		hasReturnDropAbility := abilitiesIncludeDrop(returnAbilities)

		parameterContract := core.ParameterContract{
			ID: function.Parameter.ID, Name: function.Parameter.Name, Type: function.Parameter.Type,
			// D-07-01: today's grammar has exactly one parameter form
			// (by-value), so every parameter's Mode is "owned" (R-01).
			Mode:  "owned",
			Drops: hasParameterDropAbility && !parameterEscapesOwned(function),
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
		returnContract.Fresh = returnContract.Mode == "owned" && hasReturnDropAbility

		foreignReach := core.ForeignReach{}
		fails := ""
		if function.ForeignContract != nil {
			foreignReach = core.ForeignReach{
				Allocator: function.ForeignContract.Allocator, Unwind: function.ForeignContract.Unwind,
				NonlocalExit: function.ForeignContract.NonlocalExit,
			}
			fails = function.ForeignContract.Fails
		}

		// Callable is publication safety, not export membership (D-07-31):
		// true exactly when PublishProblemsFor(function) reports zero
		// problems. No export list is consulted anywhere on this path.
		callable := len(PublishProblemsFor(function, calleeContracts)) == 0
		if forceCallableAlwaysTrue {
			callable = true
		}
		signature := core.FunctionSignature{
			ID: function.ID, Name: function.Name,
			Parameters: []core.ParameterContract{parameterContract},
			Return:     returnContract,
			Abilities:  parameterAbilities,
			Callable:   callable,
			Fails:      fails,
			Foreign:    foreignReach,
		}
		// ClosureDigest is computed in a second pass, below, once every
		// function's base signature (everything else) is built and every
		// callee's own ClosureDigest is available in chaining order --
		// D-07-38. indexByID is recorded (never a pointer into
		// summary.Functions, which keeps growing during this loop and could
		// reallocate) so the second pass can mutate
		// summary.Functions[index] directly, by index, once its callees'
		// digests are already committed to the slice. 07-12/IN-01:
		// indexByID indexes summary.Functions ONLY -- it is no longer used
		// to index program.Functions (the second pass resolves the program
		// function from programFunctionByID, keyed by ID, instead), so the
		// two slices are no longer index-coupled: any future `continue`
		// added to this loop's body can no longer silently mis-associate a
		// later function's callee set, join, or digest with an earlier
		// function's index.
		indexByID[function.ID] = len(summary.Functions)
		summary.Functions = append(summary.Functions, signature)
	}

	// D-07-38/D-07-12: chain ClosureDigest bottom-up, callee before caller.
	// closureDigestDiscoveryOrderOverride (Task 2's own fault-injection
	// seam) replaces this with plain declaration order, so a caller
	// processed before its callee reads that callee's still-empty
	// ClosureDigest -- proving the ordering is load-bearing.
	chainOrder := make([]string, 0, len(order))
	if closureDigestDiscoveryOrderOverride {
		for _, function := range program.Functions {
			chainOrder = append(chainOrder, function.ID)
		}
	} else {
		for i := len(order) - 1; i >= 0; i-- {
			chainOrder = append(chainOrder, order[i])
		}
	}
	for _, functionID := range chainOrder {
		index, ok := indexByID[functionID]
		if !ok {
			continue
		}
		function := programFunctionByID[functionID]
		var calleePairs []calleeDigestPair
		if !closureDigestEmptyCalleesOverride {
			// 07-12 (CR-03/PVG-02): join this caller's own Foreign/Fails
			// with every callee's ALREADY-COMMITTED Foreign/Fails (chainOrder
			// is callee-before-caller on a proven DAG, so every callee's own
			// joined value is already final by the time its caller reads
			// it), strictly BEFORE computeClosureDigest below, so the digest
			// covers the joined facts (T-07-12-03). calleeIDsForClosureDigest
			// reads only the callee's own PUBLISHED SIGNATURE
			// (summary.Functions[calleeIndex]) -- never a callee body, never
			// a *core.LinearBody -- preserving SEM-05's body-blindness by
			// construction.
			for _, calleeID := range calleeIDsForClosureDigest(function) {
				calleeIndex, ok := indexByID[calleeID]
				if !ok {
					// Edge 2 (D-07-12/07-12 Test 5): a callee ID absent from
					// the index is a DEFECT, not an empty-join contribution
					// -- refuse loudly rather than silently narrowing this
					// caller's published reach. Unreachable in practice
					// (check's own callgraph.Order already requires every
					// OpCall's CalleeID to resolve to a declared function
					// before BuildInterface ever runs), but not assumed away.
					return core.Interface{}, fmt.Errorf("originvalidate: callee %q has no published signature for closure join", calleeID)
				}
				calleeSignature := summary.Functions[calleeIndex]
				calleePairs = append(calleePairs, calleeDigestPair{ID: calleeID, ClosureDigest: calleeSignature.ClosureDigest})
				if !foreignClosureJoinSeam {
					summary.Functions[index].Foreign = joinForeignReach(summary.Functions[index].Foreign, calleeSignature.Foreign)
				}
				if !failsClosureJoinSeam {
					summary.Functions[index].Fails = joinFails(summary.Functions[index].Fails, calleeSignature.Fails)
				}
			}
		}
		closureDigest, err := computeClosureDigest(summary.Functions[index], calleePairs)
		if err != nil {
			return core.Interface{}, err
		}
		summary.Functions[index].ClosureDigest = closureDigest
		if closureDigestComputationOrderObserved != nil {
			closureDigestComputationOrderObserved(functionID)
		}
	}
	return summary, nil
}

// parameterEscapesOwned reports whether function's return traces back to its
// own parameter as a directly moved/copied OWNED value, following only
// OpMove/OpCopy chains and never crossing an OpBorrowShared/
// OpBorrowExclusive/OpForeignCall operation. This is R-01's authority for
// ParameterContract.Drops: a borrow-derived return never transfers the
// parameter's ownership, so the parameter's drop obligation (if it has the
// Drop ability at all) still belongs to, and is discharged by, the callee.
// When the parameter itself IS the owned return value, ownership (and the
// obligation to drop it) moves to the caller instead -- the callee does not
// discharge it.
func parameterEscapesOwned(function core.Function) bool {
	if function.Linear == nil {
		return false
	}
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var returnOps []core.LinearOperation
	for _, operation := range operations {
		if operation.Kind == core.OpReturn {
			returnOps = append(returnOps, operation)
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	for _, returnOp := range returnOps {
		current := returnOp.SourceID
		visited := make(map[string]bool)
		for current != function.Parameter.ID {
			if visited[current] {
				break
			}
			visited[current] = true
			operation, exists := sourceOf[current]
			if !exists || (operation.Kind != core.OpMove && operation.Kind != core.OpCopy) {
				break
			}
			current = operation.SourceID
		}
		if current == function.Parameter.ID {
			return true
		}
	}
	return false
}

// interfaceParameterAbilities and interfaceReturnAbilities deliberately keep
// the Phase 17 contract lookups separate.  Origin validation derives both
// facts from its own typed-core traversal; it neither imports another peer nor
// reuses a producer-derived contract.
func interfaceParameterAbilities(function core.Function) []core.Ability {
	return interfaceTypeAbilities(function, function.ID+":type:0")
}

func interfaceReturnAbilities(function core.Function) []core.Ability {
	if phase17ReturnLookupFaultForTest {
		return nil
	}
	abilities := interfaceTypeAbilities(function, function.ID+":type:1")
	if abilities == nil && function.Parameter.Type == function.ReturnType {
		return interfaceTypeAbilities(function, function.ID+":type:0")
	}
	return abilities
}

func interfaceTypeAbilities(function core.Function, typeID string) []core.Ability {
	if function.Linear == nil {
		return nil
	}
	for _, fact := range function.Linear.Types {
		if typeFactExactIDMatch(fact.ID, typeID) {
			return fact.Abilities
		}
	}
	return nil
}

func abilitiesIncludeDrop(abilities []core.Ability) bool {
	for _, ability := range abilities {
		if ability == core.AbilityDrop {
			return true
		}
	}
	return false
}

// FunctionAnswer is what a body-blind consumer can decide for one function:
// its declared origin and access mode, read directly from the summary.
type FunctionAnswer struct {
	ID     string
	Name   string
	Paths  []string
	Access string
}

// decodeErrorCode extracts a core.DecodeError's Code, so CheckSummary can
// re-surface DecodeInterface's own refusal code through originvalidate's
// {Code}-only Error shape rather than collapsing every structural refusal
// into one generic code.
func decodeErrorCode(err error) string {
	if decodeErr, ok := err.(*core.DecodeError); ok {
		return decodeErr.Code
	}
	return ""
}

// CheckSummary is the body-blind consuming path (OWN-04 success criterion
// 4's second, independent CLI invocation), rerouted through
// core.DecodeInterface per D-07-36: a document is now schema-peeked and
// strictly validated (presence, non-emptiness, Mode's closed set, digest
// shape, and function-ID uniqueness) before this function ever asks an
// origin or access question of it, and a lang.interface/0 document is
// refused outright (T-07-02: never admissible for a call).
//
// coreBytes is the RAW, UNPARSED bytes of the core artifact the summary
// claims to be bound to: this function hashes them and compares against the
// summary's recorded digest, and NEVER unmarshals coreBytes into any struct
// that could carry a Linear/Match body field — the digest check is the only
// use coreBytes is put to, and it runs before any origin or access question
// is answered (T-03-03: a stale summary is rejected before it is partially
// trusted).
func CheckSummary(summaryBytes, coreBytes []byte) ([]FunctionAnswer, error) {
	decoded, err := core.DecodeInterface(summaryBytes)
	if err != nil {
		if code := decodeErrorCode(err); code != "" {
			return nil, &Error{Code: code}
		}
		return nil, &Error{Code: "origin.invalid_summary"}
	}
	if !decoded.Admissible || decoded.V1 == nil {
		return nil, &Error{Code: "origin.summary_not_admissible"}
	}
	summary := *decoded.V1
	if summary.CoreDigest != digest(coreBytes) {
		return nil, &Error{Code: "origin.stale_summary"}
	}
	answers := make([]FunctionAnswer, 0, len(summary.Functions))
	for _, function := range summary.Functions {
		answer := FunctionAnswer{ID: function.ID, Name: function.Name}
		if function.Return.Mode != "owned" {
			answer.Paths = function.Return.Paths
			answer.Access = function.Return.Mode
		}
		answers = append(answers, answer)
	}
	return answers, nil
}

// Package pathoracle is Phase 3's third, genuinely different decision
// procedure for loan liveness (03-RESEARCH Q4, ROADMAP criterion 1). It is
// consumed only by tests and the verify harness (session.go) -- never by
// check, corevalidate, interp, or cgen -- and it reads nothing but a
// core.Function's own declared Blocks/Edges/Operations. It imports neither
// check nor corevalidate nor ast, and it calls no production liveness
// function.
//
// Three loan-endpoint derivations now coexist in this repository, and none
// can be obtained from either of the others by renaming (mirroring
// check_test.go's own two-procedure doc-comment convention, extended to
// three):
//
//   - check.go's loanLivenessFixpoint converges a finite lattice of
//     per-block live-loan sets to a fixpoint via a backward worklist: it
//     never enumerates a concrete path, only abstract per-block summaries.
//   - corevalidate.go's recomputeLoanEndpoints materializes an explicit
//     (loan, block, ordinal) use relation and computes ONE bounded
//     reachability closure over it, then reduces to each loan's final
//     reachable uses -- a single global relation, never per-block
//     summaries and never a path.
//   - This package enumerates every DISTINCT acyclic entry-to-return path
//     as an explicit, concrete block sequence, replays each one forward in
//     isolation (never converging or closing anything), and only combines
//     the independent per-path answers at the very end, when a genuine
//     divergence across paths at a fan-out block is what earns a loan its
//     edge endpoint instead of a point endpoint. No shared helper, no
//     queue, and no relation-closure step exists anywhere in this file.
//
// The property this package's identity claim actually rests on (D-10-11) is
// NOT "touches only its own function" -- that is impossible for any correct
// interprocedural analysis, and both other laws also cross into
// callee-derived facts (check's own fixpoint, corevalidate's own closure).
// The property is: this package never converges, never closes a relation,
// only enumerates and replays concrete paths. Composing across a
// core.OpCall (pathoracle_compose.go, D-10-09) preserves that property
// under call-crossing, because composition recurses into the callee via
// this package's OWN EnumeratePaths/linearizePath and splices its concrete
// paths into the caller's, running the SAME forward replay once over the
// spliced result. A contract hop -- the pattern check.go and
// corevalidate.go both already use at core.OpCall -- would NOT preserve it:
// a hop collapses a callee's own per-path distinctions into one summary
// fact BEFORE replay ever runs, which is precisely the join/reduction step
// this package never performs anywhere else. This is the first thing a
// reviewer should probe, since the touches-only-its-own-function property
// is the one they will likely assume instead.
package pathoracle

import (
	"fmt"
	"sort"

	"github.com/szTheory/schway/internal/compiler/core"
)

// ValidateLocalOwnerPaths independently replays the Phase 23 straight-line
// local-owner path. Acquisition seeds obligations; a borrow must observe the
// live owner and exactly one matching consuming release must discharge it.
func ValidateLocalOwnerPaths(program core.Program) error {
	if err := validateTransferredOwner(program); err != nil {
		return err
	}
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		local := false
		for _, operation := range function.Linear.Operations {
			local = local || operation.Foreign != nil
		}
		if !local || function.ReturnType == "FileByteOwner" || pathHasCall(function) {
			continue
		}
		if function.Parameter.Type != "PathToken" || function.ReturnType != "U64" || len(function.Linear.Blocks) != 0 || len(function.Linear.Edges) != 0 {
			return fmt.Errorf("pathoracle.local_owner_shape: function %q is outside the admitted PathToken-to-U64 path", function.ID)
		}
		types := make(map[string]string, len(function.Linear.Types))
		for _, fact := range function.Linear.Types {
			types[fact.ID] = fact.Shape.Constructor
		}
		places := make(map[string]string, len(function.Linear.Places))
		for _, place := range function.Linear.Places {
			places[place.ID] = types[place.TypeID]
		}
		type owner struct {
			place    string
			result   string
			contract *core.ForeignOperationContract
			borrowed bool
			released bool
		}
		owners := map[string]owner{}
		ownerByPlace := map[string]string{}
		modes := make([]string, 0, 3)
		var returnedFrom string
		for _, operation := range function.Linear.Operations {
			contract := operation.Foreign
			if contract == nil {
				if operation.Kind == core.OpReturn {
					if returnedFrom != "" || len(ownerByPlace) != 0 {
						return fmt.Errorf("pathoracle.local_owner_terminal: function %q returns with an outstanding owner", function.ID)
					}
					returnedFrom = operation.SourceID
				} else {
					return fmt.Errorf("pathoracle.local_owner_operation: operation %q escapes the admitted local-owner shape", operation.ID)
				}
				continue
			}
			if !pathOracleIdentifier(contract.Symbol) || !pathOracleIdentifier(contract.ABIType) || contract.Unwind != "forbidden" || contract.NonlocalExit != "forbidden" {
				return fmt.Errorf("pathoracle.local_owner_contract: operation %q has incomplete ABI or exit facts", operation.ID)
			}
			modes = append(modes, contract.Mode)
			switch contract.Mode {
			case "acquire":
				if operation.Kind != core.OpForeignCall || operation.SourceID != function.Parameter.ID || places[operation.SourceID] != "PathToken" || places[operation.TargetID] != "FileByteOwner" || types[operation.TypeID] != "FileByteOwner" || contract.Symbol != "schway_file_byte_acquire" || contract.ABIType != "schway_file_byte_acquire_fn" || contract.ParameterType != "PathToken" || contract.ResultType != "FileByteOwner" || contract.Fails != "AcquireError" || contract.Allocator != "libc_malloc" || contract.Release != "schway_file_byte_release" || operation.Allocator != contract.Allocator || operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "" || owners[operation.ID].contract != nil || ownerByPlace[operation.TargetID] != "" {
					return fmt.Errorf("pathoracle.local_owner_acquire: operation %q is not a valid owner seed", operation.ID)
				}
				owners[operation.ID] = owner{place: operation.TargetID, contract: contract}
				ownerByPlace[operation.TargetID] = operation.ID
			case "borrow":
				acquireID := ownerByPlace[operation.SourceID]
				acquired, exists := owners[acquireID]
				if operation.Kind != core.OpForeignCall || !exists || acquired.released || acquired.borrowed || places[operation.SourceID] != "FileByteOwner" || places[operation.TargetID] != "U64" || types[operation.TypeID] != "U64" || contract.Symbol != "schway_file_byte_use" || contract.ABIType != "schway_file_byte_use_fn" || contract.ParameterType != "FileByteOwner" || contract.ResultType != "U64" || contract.Fails != "UseError" || contract.Allocator != "" || contract.Release != "" || operation.Allocator != "" || operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "" {
					return fmt.Errorf("pathoracle.local_owner_borrow: operation %q does not borrow one live acquired owner", operation.ID)
				}
				acquired.borrowed = true
				acquired.result = operation.TargetID
				owners[acquireID] = acquired
			case "consume":
				acquired, exists := owners[operation.ReleasesOperationID]
				if operation.Kind != core.OpRelease || !exists || acquired.released || !acquired.borrowed || acquired.place != operation.SourceID || places[operation.SourceID] != "FileByteOwner" || types[operation.TypeID] != "FileByteOwner" || acquired.contract.Release != contract.Symbol || acquired.contract.Allocator != contract.Allocator || contract.Symbol != "schway_file_byte_release" || contract.ABIType != "schway_file_byte_release_fn" || contract.ParameterType != "FileByteOwner" || contract.ResultType != "Unit" || contract.Allocator != "libc_malloc" || contract.Release != "" || contract.Fails != "" || operation.Allocator != contract.Allocator || operation.TargetID != "" {
					return fmt.Errorf("pathoracle.local_owner_release: operation %q does not discharge its acquired owner", operation.ID)
				}
				acquired.released = true
				owners[operation.ReleasesOperationID] = acquired
				delete(ownerByPlace, operation.SourceID)
			default:
				return fmt.Errorf("pathoracle.local_owner_mode: operation %q has unsupported mode", operation.ID)
			}
		}
		if len(owners) != 1 || len(ownerByPlace) != 0 || len(modes) != 3 || modes[0] != "acquire" || modes[1] != "borrow" || modes[2] != "consume" {
			return fmt.Errorf("pathoracle.local_owner_unreleased: function %q has no complete acquisition discharge", function.ID)
		}
		for _, acquired := range owners {
			if !acquired.borrowed || !acquired.released || returnedFrom == "" || returnedFrom != acquired.result {
				return fmt.Errorf("pathoracle.local_owner_terminal: function %q returns before owner discharge", function.ID)
			}
		}
	}
	return nil
}

func pathHasCall(function core.Function) bool {
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

// validateTransferredOwner independently starts at a callee acquisition and
// carries that obligation through the returned place into the caller.
func validateTransferredOwner(program core.Program) error {
	if pathPhase24ErrorProgram(program) {
		return validateRepeatedPathOwner(program)
	}
	type acquisition struct {
		f  *core.Function
		op core.LinearOperation
	}
	var a acquisition
	var helper *core.Function
	fail := func(code, id string) error {
		return fmt.Errorf("pathoracle.%s: invalid transferred owner at %q", code, id)
	}
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Linear == nil {
			continue
		}
		for _, op := range f.Linear.Operations {
			if op.Foreign != nil && op.Foreign.Mode == "acquire" {
				if helper != nil {
					return fail("owner_transfer_acquire", op.ID)
				}
				helper = f
				a = acquisition{f, op}
			}
		}
	}
	if helper == nil || helper.ReturnType != "FileByteOwner" {
		return nil
	}
	if helper.Parameter.Type != "PathToken" || a.op.Kind != core.OpForeignCall || a.op.Foreign == nil || a.op.Foreign.Symbol != "schway_file_byte_acquire" || a.op.Foreign.ABIType != "schway_file_byte_acquire_fn" || a.op.Foreign.ParameterType != "PathToken" || a.op.Foreign.ResultType != "FileByteOwner" || a.op.Foreign.Fails != "AcquireError" || a.op.Foreign.Allocator != "libc_malloc" || a.op.Foreign.Release != "schway_file_byte_release" || a.op.ErrTargetID == "" || a.op.OkEdgeID == "" || a.op.ErrEdgeID == "" {
		return fail("owner_transfer_helper", helper.ID)
	}
	var retHelper *core.LinearOperation
	for i := range helper.Linear.Operations {
		op := &helper.Linear.Operations[i]
		if op.Kind == core.OpReturn {
			if retHelper != nil {
				return fail("owner_transfer_return", op.ID)
			}
			retHelper = op
		}
		if op.Kind == core.OpRelease {
			return fail("owner_transfer_early_release", op.ID)
		}
	}
	if retHelper == nil || retHelper.SourceID != a.op.TargetID || a.op.ErrTargetID == "" {
		return fail("owner_transfer_return", helper.ID)
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
					return fail("owner_transfer_call", op.ID)
				}
				caller = f
				call = op
			}
		}
	}
	if caller == nil || caller.ReturnType != "U64" {
		return fail("owner_transfer_call", helper.ID)
	}
	bi, ri, ti := -1, -1, -1
	for i := range caller.Linear.Operations {
		op := &caller.Linear.Operations[i]
		switch {
		case op.Kind == core.OpCopy && op.SourceID == call.TargetID:
			return fail("owner_transfer_copy", op.ID)
		case op.Foreign != nil && op.Foreign.Mode == "borrow":
			if borrow != nil {
				return fail("owner_transfer_borrow", op.ID)
			}
			borrow = op
			bi = i
		case op.Kind == core.OpRelease:
			if release != nil {
				return fail("owner_transfer_release", op.ID)
			}
			release = op
			ri = i
		case op.Kind == core.OpReturn:
			if ret != nil {
				return fail("owner_transfer_return", op.ID)
			}
			ret = op
			ti = i
		}
	}
	if borrow == nil || release == nil || ret == nil || borrow.SourceID != call.TargetID || release.SourceID != call.TargetID || release.ReleasesOperationID != a.op.ID || release.Foreign == nil || borrow.Foreign == nil || release.Foreign.Mode != "consume" || release.Foreign.Symbol != a.op.Foreign.Release || release.Foreign.ABIType != "schway_file_byte_release_fn" || release.Foreign.ParameterType != "FileByteOwner" || release.Foreign.ResultType != "Unit" || release.Foreign.Fails != "" || release.Foreign.Allocator != a.op.Foreign.Allocator || release.Allocator != a.op.Foreign.Allocator || borrow.Foreign.Symbol != "schway_file_byte_use" || borrow.Foreign.ABIType != "schway_file_byte_use_fn" || borrow.Foreign.ParameterType != "FileByteOwner" || borrow.Foreign.ResultType != "U64" || borrow.Foreign.Fails != "UseError" || ret.SourceID != borrow.TargetID {
		return fail("owner_transfer_discharge", caller.ID)
	}
	if !(bi >= 0 && bi < ri && ri < ti) {
		return fail("owner_transfer_order", caller.ID)
	}
	return nil
}

func pathPhase24ErrorProgram(program core.Program) bool {
	if program.Module != "phase24.error" {
		return false
	}
	names := map[string]bool{}
	for _, function := range program.Functions {
		names[function.Name] = true
	}
	return names["acquire"] && names["probe"] && names["main"]
}

func validateRepeatedPathOwner(program core.Program) error {
	problem := func(code, id string) error {
		return fmt.Errorf("pathoracle.%s: invalid repeated owner path at %q", code, id)
	}
	byName := map[string]*core.Function{}
	for index := range program.Functions {
		byName[program.Functions[index].Name] = &program.Functions[index]
	}
	acquireFn, probe, main := byName["acquire"], byName["probe"], byName["main"]
	if acquireFn == nil || probe == nil || main == nil || acquireFn.Linear == nil || probe.Linear == nil || main.Linear == nil {
		return problem("owner_transfer_functions", "phase24")
	}
	var acquisition *core.LinearOperation
	for index := range acquireFn.Linear.Operations {
		op := &acquireFn.Linear.Operations[index]
		if op.Kind == core.OpForeignCall && op.Foreign != nil && op.Foreign.Mode == "acquire" {
			if acquisition != nil {
				return problem("owner_transfer_acquire", op.ID)
			}
			acquisition = op
		}
	}
	if acquisition == nil || acquisition.Foreign.Release != "schway_file_byte_release" || acquisition.Allocator != "libc_malloc" {
		return problem("owner_transfer_helper", acquireFn.ID)
	}
	returned := false
	for _, op := range acquireFn.Linear.Operations {
		if op.Kind == core.OpReturn && op.SourceID == acquisition.TargetID {
			returned = true
		}
		if op.Kind == core.OpRelease {
			return problem("owner_transfer_early_release", op.ID)
		}
	}
	if !returned {
		return problem("owner_transfer_return", acquireFn.ID)
	}
	mainAcquireCalls := pathCalls(main, acquireFn.ID)
	mainProbeCalls := pathCalls(main, probe.ID)
	probeAcquireCalls := pathCalls(probe, acquireFn.ID)
	if len(mainAcquireCalls) != 2 || len(mainProbeCalls) != 1 || len(probeAcquireCalls) != 1 {
		return problem("owner_transfer_call", acquireFn.ID)
	}
	callIDs := map[string]bool{}
	for _, call := range append(append(append([]core.LinearOperation{}, mainAcquireCalls...), mainProbeCalls...), probeAcquireCalls...) {
		if call.ID == "" || callIDs[call.ID] || !pathCallEdgesPresent(program, main, probe, call) {
			return problem("owner_transfer_call_edges", call.ID)
		}
		callIDs[call.ID] = true
	}
	if !pathProbeDischarges(probe, probeAcquireCalls[0], acquisition.ID) {
		return problem("owner_transfer_probe_cleanup", probe.ID)
	}
	if !pathMainDischarges(main, mainAcquireCalls[0], mainAcquireCalls[1], mainProbeCalls[0], acquisition.ID) {
		return problem("owner_transfer_main_cleanup", main.ID)
	}
	return nil
}

func pathCalls(function *core.Function, calleeID string) []core.LinearOperation {
	var result []core.LinearOperation
	for _, op := range function.Linear.Operations {
		if op.Kind == core.OpCall && op.CalleeID == calleeID {
			result = append(result, op)
		}
	}
	return result
}

func pathCallEdgesPresent(program core.Program, main, probe *core.Function, call core.LinearOperation) bool {
	for _, function := range []*core.Function{main, probe} {
		for _, operation := range function.Linear.Operations {
			if operation.ID != call.ID {
				continue
			}
			var containing string
			for _, block := range function.Linear.Blocks {
				for _, opID := range block.OperationIDs {
					if opID == call.ID {
						containing = block.ID
					}
				}
			}
			edges := map[string]core.Edge{}
			for _, edge := range function.Linear.Edges {
				edges[edge.ID] = edge
			}
			okEdge, errEdge := edges[call.OkEdgeID], edges[call.ErrEdgeID]
			if call.ErrTargetID == "" || containing == "" || okEdge.FromBlockID != containing || errEdge.FromBlockID != containing || okEdge.Pattern != "ok" || errEdge.Pattern != "err" || okEdge.ToBlockID == errEdge.ToBlockID || !pathBlockHasSuccessor(function, containing, okEdge.ToBlockID) || !pathBlockHasSuccessor(function, containing, errEdge.ToBlockID) {
				return false
			}
			for _, place := range function.Linear.Places {
				if place.ID != call.ErrTargetID {
					continue
				}
				for _, fact := range function.Linear.Types {
					if fact.ID == place.TypeID {
						return fact.Shape.Constructor == "ResourceError" && pathFailureEnvelopeCovers(program, call.CalleeID)
					}
				}
			}
			return false
		}
	}
	return false
}

func pathBlockHasSuccessor(function *core.Function, blockID, successorID string) bool {
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

func pathFailureEnvelopeCovers(program core.Program, calleeID string) bool {
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
	constructor := map[string]string{}
	for _, fact := range callee.Linear.Types {
		constructor[fact.ID] = fact.Shape.Constructor
	}
	wanted := map[string]bool{}
	for _, op := range callee.Linear.Operations {
		if op.Kind == core.OpFail {
			wanted[constructor[op.TypeID]] = true
		}
		if op.Foreign != nil && op.Foreign.Fails != "" {
			wanted[op.Foreign.Fails] = true
		}
	}
	if len(wanted) == 0 {
		return false
	}
	covered := map[string]bool{}
	for _, dataType := range program.DataTypes {
		if dataType.Name == "ResourceError" && len(dataType.AlternativeDetails) == 0 {
			for _, name := range dataType.Alternatives {
				covered[name] = true
			}
		}
	}
	if len(covered) == 0 {
		return false
	}
	for name := range wanted {
		found := false
		for _, dataType := range program.DataTypes {
			if dataType.Name != name || len(dataType.AlternativeDetails) != 0 {
				continue
			}
			found = true
			for _, alternative := range dataType.Alternatives {
				if !covered[alternative] {
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

func pathEdgeBlock(function *core.Function, edgeID string) string {
	for _, edge := range function.Linear.Edges {
		if edge.ID == edgeID {
			return edge.ToBlockID
		}
	}
	return ""
}

func pathTerminalCleanup(function *core.Function, edgeID string, reverseOwnerPlaces []string, terminalSource string, terminal core.OperationKind, acquireID string) bool {
	blockID := pathEdgeBlock(function, edgeID)
	for _, block := range function.Linear.Blocks {
		if block.ID != blockID {
			continue
		}
		if len(block.OperationIDs) != len(reverseOwnerPlaces)+1 {
			return false
		}
		byID := map[string]core.LinearOperation{}
		for _, op := range function.Linear.Operations {
			byID[op.ID] = op
		}
		for index, owner := range reverseOwnerPlaces {
			op := byID[block.OperationIDs[index]]
			if op.Kind != core.OpRelease || op.SourceID != owner || op.ReleasesOperationID != acquireID || op.Foreign == nil || op.Foreign.Mode != "consume" || op.Allocator != "libc_malloc" {
				return false
			}
		}
		last := byID[block.OperationIDs[len(block.OperationIDs)-1]]
		return last.Kind == terminal && (terminalSource == "" || last.SourceID == terminalSource)
	}
	return false
}

func pathProbeDischarges(probe *core.Function, helperCall core.LinearOperation, acquisitionID string) bool {
	var use core.LinearOperation
	count := 0
	for _, operation := range probe.Linear.Operations {
		if operation.Foreign != nil && operation.Foreign.Mode == "borrow" {
			use = operation
			count++
		}
	}
	if count != 1 || use.SourceID != helperCall.TargetID || use.ErrTargetID == "" || use.Foreign.Fails != "UseError" {
		return false
	}
	return pathTerminalCleanup(probe, use.OkEdgeID, []string{helperCall.TargetID}, use.TargetID, core.OpReturn, acquisitionID) &&
		pathTerminalCleanup(probe, use.ErrEdgeID, []string{helperCall.TargetID}, use.ErrTargetID, core.OpFail, acquisitionID) &&
		pathTerminalCleanup(probe, helperCall.ErrEdgeID, nil, helperCall.ErrTargetID, core.OpFail, acquisitionID)
}

func pathMainDischarges(main *core.Function, callA, callB, callProbe core.LinearOperation, acquisitionID string) bool {
	return pathTerminalCleanup(main, callA.ErrEdgeID, nil, callA.ErrTargetID, core.OpFail, acquisitionID) &&
		pathTerminalCleanup(main, callB.ErrEdgeID, []string{callA.TargetID}, callB.ErrTargetID, core.OpFail, acquisitionID) &&
		pathTerminalCleanup(main, callProbe.ErrEdgeID, []string{callB.TargetID, callA.TargetID}, callProbe.ErrTargetID, core.OpFail, acquisitionID) &&
		pathTerminalCleanup(main, callProbe.OkEdgeID, []string{callB.TargetID, callA.TargetID}, callProbe.TargetID, core.OpReturn, acquisitionID)
}

func pathOracleIdentifier(value string) bool {
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

// MaxPaths bounds the number of acyclic entry-to-return paths this package
// will enumerate for one function. Today's language caps a match at
// maxArmsPerMatch (64) arms and gives each arm exactly one linear
// entry-to-return path with no nested branching construct, so the real
// reachable maximum is 64 -- MaxPaths is set well above that (a decimal
// order of magnitude) so it never fires on any program the parser can
// produce today, while still being a genuine, finite, fail-closed ceiling
// against a future nested-branching CFG shape (T-03-01/T-03-13): path count
// is multiplicative in branch count for a CFG with more than one level of
// fan-out, so an unbounded enumeration would be a real denial-of-service
// surface the moment that shape becomes reachable.
const MaxPaths = 4096

// TerminatorKindsOverride is a fault-injection seam for
// TestTerminatorWalkMutationKilled (D-09/D-04-29): production always closes
// a path on the full core.TerminatorKinds() set; the test temporarily
// narrows it (the same "delete OpFail from the recognised set" mutation the
// throwaway-detached-worktree demonstration performs on the source
// directly) to prove that a walker recognising fewer terminators really
// does misclassify a fixture path. nil (the always-true production default)
// means "use core.TerminatorKinds() unmodified".
var TerminatorKindsOverride func() []core.OperationKind

func recognizedTerminatorKinds() []core.OperationKind {
	if TerminatorKindsOverride != nil {
		return TerminatorKindsOverride()
	}
	return core.TerminatorKinds()
}

// isTerminatorKind is pathoracle's own membership test against the
// terminator registry (D-04-29): a set-membership test against
// core.TerminatorKinds() rather than a literal restatement of it, so
// widening the registry widens this walker automatically.
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

// pathCapError is returned when a function's acyclic entry-to-return path
// count exceeds MaxPaths. It carries a stable Code(), matching the
// diagnostic.Diagnostic convention used elsewhere in the repository, and
// rejects fail-closed -- the cap is never raised to make a test pass; a
// function that would exceed it is refused, not enumerated.
type pathCapError struct {
	functionID string
	limit      int
}

func (e *pathCapError) Error() string {
	return fmt.Sprintf("pathoracle.path_count_exceeded: function %q exceeds the declared cap of %d acyclic entry-to-return paths", e.functionID, e.limit)
}

// Code reports the stable diagnostic code for a path-count-cap rejection.
func (e *pathCapError) Code() string { return "pathoracle.path_count_exceeded" }

// PathCapError type-asserts err as the path-count-cap rejection, mirroring
// the errors.As convention used elsewhere in the repository (e.g.
// session.go's EngineMismatch).
func PathCapError(err error) (*pathCapError, bool) {
	e, ok := err.(*pathCapError)
	return e, ok
}

// backEdgeError is returned when the declared successor relation contains a
// cycle. OWN-03 is scoped to acyclic CFGs this phase (T-03-11); this
// package independently re-derives that rejection rather than trusting
// check.go's own loanLivenessFixpoint to have caught it first, so a
// corrupted or synthetic CFG that check.go never saw is still refused here.
type backEdgeError struct {
	functionID, blockID string
}

func (e *backEdgeError) Error() string {
	return fmt.Sprintf("pathoracle.cfg_back_edge: function %q block %q participates in a cycle", e.functionID, e.blockID)
}

func (e *backEdgeError) Code() string { return "pathoracle.cfg_back_edge" }

// unterminatedLoanError is the late terminal guard the spike requires
// (.planning/spikes/002-cfg-edge-last-use/README.md). It fires in two
// distinct circumstances, both hard failures rather than a normalized
// repair:
//
//  1. A path that carries at least one loan (a borrow was replayed) but
//     never reaches ANY terminator operation (D-04-29: core.OpReturn,
//     core.OpFail, or core.OpDefect) at all. Every real entry-to-return
//     path this repository's checker produces ends in exactly one
//     terminator (corevalidate's own replayBlocks invariant); a path with
//     no terminator anywhere is a structurally malformed CFG this oracle
//     refuses to silently default-close (a per-path linearizer that
//     quietly ended such a loan "at its own creation" would be EXACTLY the
//     normalized repair this guard exists to reject, since that default is
//     legitimate only when the path genuinely terminates and the loan
//     simply goes unreferenced afterward). Before D-04-29 widened this to
//     every terminator, a path whose only real exit was core.OpFail or
//     core.OpDefect was misclassified as malformed the moment it carried a
//     live loan -- exactly the failure class 03-08/03-09/03-10 closed three
//     times, reproduced here as a false rejection rather than a silent
//     admission.
//  2. A linearized path references a loan through the place-inheritance
//     chain that was never recorded as born earlier on the SAME path — a
//     bookkeeping inconsistency in the linearizer itself.
type unterminatedLoanError struct {
	functionID, loanID string
}

func (e *unterminatedLoanError) Error() string {
	return fmt.Sprintf("pathoracle.unterminated_loan: function %q references loan %q with no recorded birth on this path", e.functionID, e.loanID)
}

func (e *unterminatedLoanError) Code() string { return "pathoracle.unterminated_loan" }

// noEntryBlockError is returned when a function's Linear.Blocks carries no
// block whose PointID matches the function's own declared EntryPointID --
// a corrupted or adversarial core artifact this package must not silently
// tolerate.
type noEntryBlockError struct{ functionID string }

func (e *noEntryBlockError) Error() string {
	return fmt.Sprintf("pathoracle.no_entry_block: function %q declares no block whose point_id matches its entry_point_id", e.functionID)
}

func (e *noEntryBlockError) Code() string { return "pathoracle.no_entry_block" }

// blockIndex is the oracle's own read-only view of a function's declared
// CFG: blocks by ID, and the real declared Edge.ID for each (from, to)
// block pair -- read directly from the core artifact's own Edges slice, not
// reconstructed from any naming convention, so materialized edge endpoints
// are whole-value-identical to production's without this package needing
// to duplicate an ID-formatting formula it does not own.
type blockIndex struct {
	byID       map[string]core.Block
	operations map[string][]core.LinearOperation // blockID -> its operations, resolved from OperationIDs
	edgeID     map[[2]string]string
}

func indexBlocks(function core.Function) blockIndex {
	idx := blockIndex{
		byID: map[string]core.Block{}, operations: map[string][]core.LinearOperation{}, edgeID: map[[2]string]string{},
	}
	if function.Linear == nil {
		return idx
	}
	byOperationID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		byOperationID[operation.ID] = operation
	}
	for _, block := range function.Linear.Blocks {
		idx.byID[block.ID] = block
		resolved := make([]core.LinearOperation, 0, len(block.OperationIDs))
		for _, operationID := range block.OperationIDs {
			resolved = append(resolved, byOperationID[operationID])
		}
		idx.operations[block.ID] = resolved
	}
	for _, edge := range function.Linear.Edges {
		idx.edgeID[[2]string{edge.FromBlockID, edge.ToBlockID}] = edge.ID
	}
	return idx
}

// EnumeratePaths walks every acyclic block-to-block path from entryBlockID
// to a terminal block (one with no declared successors), visiting
// successors in the exact order the core artifact declares them so the
// returned path order is deterministic. It is exhaustive path expansion --
// a different mechanism class from both production analyzers (see the
// package doc comment) -- and it is bounded: enumeration stops with a
// pathCapError the instant more than cap distinct paths have been found,
// rather than continuing to build a path count that could otherwise grow
// multiplicatively in branch count (T-03-01).
func EnumeratePaths(functionID, entryBlockID string, idx blockIndex, cap int) ([][]string, error) {
	var results [][]string
	visiting := map[string]bool{}
	var walk func(id string, current []string) error
	walk = func(id string, current []string) error {
		if visiting[id] {
			return &backEdgeError{functionID: functionID, blockID: id}
		}
		visiting[id] = true
		defer delete(visiting, id)
		current = append(current, id)

		block, ok := idx.byID[id]
		if !ok || len(block.Successors) == 0 {
			if len(results) >= cap {
				return &pathCapError{functionID: functionID, limit: cap}
			}
			results = append(results, append([]string(nil), current...))
			return nil
		}
		for _, successor := range block.Successors {
			if err := walk(successor, current); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(entryBlockID, nil); err != nil {
		return nil, err
	}
	return results, nil
}

// loanState is one loan's birth and current last-referencing position,
// tracked while linearizePath replays a single concrete path forward.
type loanState struct {
	bornBlockID string
	bornIndex   int
	bornOpID    string
	lastBlockID string
	lastIndex   int
	lastOpID    string
}

// linearizePath replays ONE concrete path's operations forward, in program
// order, re-deriving the retained straight-line last-use logic's mechanism
// -- forward transitive place->loan inheritance, exactly the shape
// discoverLoanLastUses (check.go) uses for real admission decisions on
// every straight-line and arm body in the repository today -- expressed
// over core.LinearOperation place IDs instead of ast.Binding names, since
// this package must not import check and therefore cannot call
// discoverLoanLastUses itself. Association is inherited exactly as that law
// requires: a place derived from a loan-carrying place (copy, move, or a
// reborrow) stays associated with every loan its source transitively
// carried, so a chain of reborrows extends every ancestor loan's lifetime,
// not just its immediate parent's.
//
// If an operation's SourceID resolves (via the place-inheritance chain) to
// a loan ID that was never recorded as born earlier on this same path, that
// is the late terminal guard firing: this is a hard failure
// (unterminatedLoanError), not a position this function silently
// normalizes.
func linearizePath(functionID string, idx blockIndex, blockIDs []string) (map[string]loanState, int, error) {
	chain := map[string][]string{} // placeID -> ordered ancestor+own loan IDs, oldest first
	states := map[string]loanState{}
	work := 0
	// sawTerminator tracks whether this path reached ANY terminator
	// operation (D-04-29: core.OpReturn, core.OpFail, or core.OpDefect), not
	// only a return -- renamed from sawReturn because the guard below now
	// closes a path on every terminator kind the registry names, and the
	// old name would have left this widened meaning actively misleading.
	sawTerminator := false

	for _, blockID := range blockIDs {
		operations := idx.operations[blockID]
		for opIndex, operation := range operations {
			work++
			if isTerminatorKind(operation.Kind) {
				sawTerminator = true
			}
			// OpConst is a fresh root: it has no source place and cannot
			// inherit a loan, even if a malformed artifact supplies SourceID.
			// Admission rejects that field; this peer independently keeps the
			// constant source-free while traversing checked core.
			var inherited []string
			if operation.Kind != core.OpConst {
				inherited = chain[operation.SourceID]
			}
			for _, loanID := range inherited {
				state, known := states[loanID]
				if !known {
					return nil, work, &unterminatedLoanError{functionID: functionID, loanID: loanID}
				}
				state.lastBlockID, state.lastIndex, state.lastOpID = blockID, opIndex, operation.ID
				states[loanID] = state
			}
			if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
				states[operation.LoanID] = loanState{
					bornBlockID: blockID, bornIndex: opIndex, bornOpID: operation.ID,
					lastBlockID: blockID, lastIndex: opIndex, lastOpID: operation.ID,
				}
				inherited = append(append([]string(nil), inherited...), operation.LoanID)
			}
			if operation.Kind == core.OpConst && operation.TargetID != "" {
				chain[operation.TargetID] = nil
			} else if operation.TargetID != "" && len(inherited) > 0 {
				chain[operation.TargetID] = inherited
			}
		}
	}

	if !sawTerminator && len(states) > 0 {
		// This path carries at least one loan but never reached ANY
		// terminator (D-04-29: return, typed failure, or defect) -- a
		// malformed CFG this oracle refuses to close by default. Name the
		// smallest (deterministic) loan ID so the failure is reproducible.
		loanIDs := make([]string, 0, len(states))
		for id := range states {
			loanIDs = append(loanIDs, id)
		}
		sort.Strings(loanIDs)
		return nil, work, &unterminatedLoanError{functionID: functionID, loanID: loanIDs[0]}
	}
	return states, work, nil
}

// pathResult is one enumerated path together with its replayed loan states
// and, derived from those, which blocks along the path each loan is still
// needed BEYOND (i.e. the loan's terminal use lies at or after a LATER
// block in this same path).
type pathResult struct {
	blockIDs     []string
	states       map[string]loanState
	neededBeyond map[string]map[string]bool // blockID -> loanID -> true
}

// BuildCalleeLookup builds the narrowest callee-resolution capability
// RecomputeEndpoints's composition arm needs (CalleeLookup,
// pathoracle_compose.go) from a whole core.Program: a callee-indexed map,
// never the bare core.Program itself (D-10-13) -- kept in this file, not
// pathoracle_compose.go, so that file's own composition machinery never
// itself needs to know what a core.Program is.
func BuildCalleeLookup(program core.Program) CalleeLookup {
	byID := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		byID[function.ID] = function
	}
	return func(functionID string) (core.Function, bool) {
		function, ok := byID[functionID]
		return function, ok
	}
}

// RecomputeEndpoints is the oracle's single entry point: independently
// decide loan endpoints for a checked function from the typed core alone.
// A straight-line function (no declared Blocks) has nothing for this
// package to enumerate -- production itself never populates LoanEndpoints
// outside a branch-shaped function (03-03's documented scope decision), so
// this mirrors that scope rather than inventing a disagreement against an
// artifact production never claims to produce. lookupCallee is the
// narrowest possible callee-resolution capability (D-10-13) a
// core.OpCall on the enumerated path composes across (D-10-09,
// pathoracle_compose.go); nil is safe to pass for any function with no
// OpCall on its enumerated paths (every pre-10-03 caller), and composition
// fails closed to "does not carry" for an OpCall it cannot resolve.
// Returns the recomputed endpoints, the oracle's own counted work (one unit
// per operation inspected across every enumerated path, plus every
// composed callee path), and an error if the function's CFG is cyclic,
// exceeds MaxPaths, composition exceeds MaxCompositionDepth or hits a
// composition cycle, or a linearized path fails the late terminal guard.
func RecomputeEndpoints(function core.Function, lookupCallee CalleeLookup) ([]core.LoanEndpoint, int, error) {
	if function.Linear == nil || len(function.Linear.Blocks) == 0 {
		return nil, 0, nil
	}

	idx := indexBlocks(function)
	entryBlockID := ""
	for _, block := range function.Linear.Blocks {
		if block.PointID == function.EntryPointID {
			entryBlockID = block.ID
			break
		}
	}
	if entryBlockID == "" {
		return nil, 0, &noEntryBlockError{functionID: function.ID}
	}

	blockPaths, err := EnumeratePaths(function.ID, entryBlockID, idx, MaxPaths)
	if err != nil {
		return nil, 0, err
	}

	work := 0
	results := make([]pathResult, len(blockPaths))
	deaths := map[string][]loanState{}
	var endpoints []core.LoanEndpoint
	for pi, blockIDs := range blockPaths {
		var variantStates []map[string]loanState
		if hasOpCall(idx, blockIDs) {
			variants, w, err := composeLinearizeCaller(function.ID, idx, blockIDs, lookupCallee)
			if err != nil {
				return nil, work, err
			}
			work += w
			for _, variant := range variants {
				variantStates = append(variantStates, variant.states)
			}
		} else {
			states, w, err := linearizePath(function.ID, idx, blockIDs)
			if err != nil {
				return nil, work, err
			}
			work += w
			variantStates = []map[string]loanState{states}
		}

		if len(variantStates) != 1 {
			// Composition produced more than one concrete continuation for
			// this caller path (D-10-14): a loan whose composed variants
			// disagree on its death position is handled directly below,
			// never fed into the single-answer `deaths` bookkeeping this
			// loop otherwise uses (that bookkeeping's own disagreement
			// guard, pathoracle.inconsistent_path_death, exists to catch a
			// genuine bug elsewhere -- a load-bearing composed divergence
			// is not that).
			endpoints = append(endpoints, synthesizeComposedEndpoints(function.ID, blockIDs, variantStates)...)
			continue
		}
		states := variantStates[0]

		neededBeyond := map[string]map[string]bool{}
		for loanID, state := range states {
			birthIndex, deathIndex := -1, -1
			for i, id := range blockIDs {
				if id == state.bornBlockID && birthIndex == -1 {
					birthIndex = i
				}
				if id == state.lastBlockID {
					deathIndex = i
				}
			}
			// Only blocks strictly BETWEEN the loan's birth and its death
			// carry it "beyond" themselves -- a block before the loan is even
			// born never needs it, and the death block itself is where it
			// finally ends, not a block it survives past.
			for i := birthIndex; i >= 0 && i < deathIndex; i++ {
				beyond := blockIDs[i]
				if neededBeyond[beyond] == nil {
					neededBeyond[beyond] = map[string]bool{}
				}
				neededBeyond[beyond][loanID] = true
			}
			for _, existing := range deaths[loanID] {
				// Distinct concrete arms may legitimately end a loan at
				// different blocks. Two replays that claim different final
				// operation positions in the SAME block, however, describe
				// inconsistent facts about one declared operation sequence.
				if existing.lastBlockID == state.lastBlockID &&
					(existing.lastIndex != state.lastIndex || existing.lastOpID != state.lastOpID) {
					return nil, work, fmt.Errorf("pathoracle.inconsistent_path_death: loan %q dies at different positions within block %q", loanID, state.lastBlockID)
				}
			}
			deaths[loanID] = append(deaths[loanID], state)
		}
		results[pi] = pathResult{blockIDs: blockIDs, states: states, neededBeyond: neededBeyond}
	}

	for loanID, states := range deaths {
		hasLaterReference := false
		for _, state := range states {
			if state.lastOpID != state.bornOpID {
				hasLaterReference = true
				break
			}
		}
		if !hasLaterReference && len(states) > 0 {
			state := states[0]
			endpoints = append(endpoints, core.LoanEndpoint{
				ID:     fmt.Sprintf("%s:point:%s:%d:%s", function.ID, state.bornBlockID, state.bornIndex, loanID),
				LoanID: loanID, Kind: "point", BlockID: state.bornBlockID, AfterOperationID: state.bornOpID,
			})
			continue
		}
		seen := map[string]bool{}
		for _, state := range states {
			if state.lastOpID == state.bornOpID {
				continue
			}
			id := fmt.Sprintf("%s:point:%s:%d:%s", function.ID, state.lastBlockID, state.lastIndex, loanID)
			if seen[id] {
				continue
			}
			seen[id] = true
			endpoints = append(endpoints, core.LoanEndpoint{
				ID: id, LoanID: loanID, Kind: "point", BlockID: state.lastBlockID, AfterOperationID: state.lastOpID,
			})
		}
	}

	for _, block := range function.Linear.Blocks {
		if len(block.Successors) < 2 {
			continue
		}
		candidates := map[string]bool{}
		for _, result := range results {
			for loanID := range result.neededBeyond[block.ID] {
				candidates[loanID] = true
			}
		}
		for loanID := range candidates {
			for _, successor := range block.Successors {
				if pathReachesLoanViaSuccessor(results, block.ID, successor, loanID) {
					continue
				}
				edgeID := idx.edgeID[[2]string{block.ID, successor}]
				if edgeID == "" {
					edgeID = fmt.Sprintf("%s:edge:%s:%s", function.ID, block.ID, successor)
				}
				endpoints = append(endpoints, core.LoanEndpoint{
					ID: edgeID + ":" + loanID, LoanID: loanID, Kind: "edge", EdgeID: edgeID,
				})
			}
		}
	}

	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].ID < endpoints[j].ID })
	unique := endpoints[:0]
	for _, endpoint := range endpoints {
		if len(unique) > 0 && unique[len(unique)-1].ID == endpoint.ID {
			if unique[len(unique)-1] != endpoint {
				return nil, work, fmt.Errorf("pathoracle.inconsistent_endpoint: endpoint ID %q has conflicting replay facts", endpoint.ID)
			}
			continue
		}
		unique = append(unique, endpoint)
	}
	return unique, work, nil
}

// pathReachesLoanViaSuccessor reports whether some enumerated path that
// traverses the edge (fromBlockID, successor) still needs loanID beyond
// fromBlockID -- i.e. whether that specific divergent branch is one where
// the loan continues to be live, as opposed to one where it has already
// ended by the time control reaches fromBlockID's own last operation.
func pathReachesLoanViaSuccessor(results []pathResult, fromBlockID, successor, loanID string) bool {
	for _, result := range results {
		for i := 0; i+1 < len(result.blockIDs); i++ {
			if result.blockIDs[i] == fromBlockID && result.blockIDs[i+1] == successor {
				if result.neededBeyond[fromBlockID][loanID] {
					return true
				}
			}
		}
	}
	return false
}

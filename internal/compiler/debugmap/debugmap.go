// Package debugmap is the bounded, semantic-only half of the debug-lineage
// experiment `wiki/debug-evidence-symbolication-and-proof.md` names as the
// "Next bounded experiment" (D-01..D-04, 03-RESEARCH Q7). It joins a source
// span to a typed-core identity to the operation, point, or edge identity
// that identity produced, for one checked program, and reports absence
// honestly rather than inventing a value.
//
// Scope fence (D-03), recorded here so it is auditable by name rather than
// implied: of the wiki note's six-step experiment, this package implements
// ONLY steps 1, the semantic half of step 2, and 4. The following are
// REJECTED, deliberately, this phase:
//
//   - The NATIVE-DEBUG-INFO half of step 2 (DWARF, CodeView, or even a
//     minimal `#line` directive emitted by cgen) — D-03 names DWARF/CodeView
//     emission as explicitly out of scope, and neither OWN-03 nor OWN-04
//     needs native debug info to hold.
//   - Step 3's PANIC/SEGFAULT capture half — any crash-capture mechanism is
//     the storage-and-symbolication subsystem D-03 forbids, even a minimal
//     one touches signal-handling/core-dump-adjacent code. Step 3's
//     ownership-diagnostic capture half is satisfied by step 4 instead,
//     using the existing diagnostic.Diagnostic machinery — no new capture
//     path is added for it.
//   - Step 5 (fault injection across stale-symbol, inlining, redaction,
//     truncation, and wrong-build-ID defects) — it presupposes the native
//     debug-info artifacts the rejected halves of steps 2 and 3 would have
//     produced. With those rejected, step 5 has nothing to mutate this
//     phase.
//   - Step 6 (compile/capture/symbolication latency measurement) — only
//     meaningful once step 2's native half and steps 3/5 exist, rejected for
//     the same reason. The semantic-side emission cost and output bytes
//     instead piggyback on the existing recomputed-work/output-bytes
//     metrics at no extra cost (see debugmap.Build's Work return and the
//     CLI command's completeCommand wiring).
//
// Bounding (D-04, D-15): MaxEntries and MaxOutputBytes are declared caps
// that fail closed with a stable code rather than truncating; Build accepts
// a context.Context and periodically checks its deadline; every entry
// increments the returned work count. No new spawned process and no file
// I/O beyond what the compiler already does is introduced by this package.
package debugmap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

// Schema versions this package's artifact independently of every other
// schema in the compiler. It is a new constant beside its peers
// (core.Schema1, evidence.Schema1, core.InterfaceSchema, ...) — no existing
// schema constant is edited to introduce it.
const Schema = "schway.debug-map/0"

// Availability is the honest three-valued report D-04 requires: a lineage
// entry never fabricates a value it does not have.
type Availability string

const (
	// Available means the map found and joined a real source span, core ID,
	// and operation/point/edge identity.
	Available Availability = "available"
	// OptimizedOut is reserved for a future native-lowering consumer of this
	// map that can observe a fact the semantic layer produced but a later
	// lowering step removed. No path in this phase's semantic-only scope
	// produces it — it exists so the schema does not need to change when a
	// future, in-scope consumer needs it — but the enumeration itself is
	// part of the honest-reporting contract D-04 requires, so it is declared
	// now rather than added silently later.
	OptimizedOut Availability = "optimized_out"
	// NotCaptured means the map has no entry for the requested identity:
	// never a guess, never a nearest-neighbour span.
	NotCaptured Availability = "not_captured"
)

// MaxEntries bounds the number of lineage entries a single Build call may
// produce. Exceeding it fails closed with ErrEntryCapExceeded rather than
// silently truncating the map — a truncated debug map is worse than none,
// since it would look complete to a consumer.
const MaxEntries = 4096

// MaxOutputBytes bounds the serialized size of a Map's Entries. Sized like
// evidence.MaxManifestBytes — this is a small, bounded side artifact, not
// source. Exceeding it fails closed with ErrOutputCeilingExceeded.
const MaxOutputBytes = 1 << 20

// Error is this package's stable typed failure, matching the {Code}-only
// shape evidence.ValidationError and originvalidate.Error already use, so
// callers can dispatch on Code identically.
type Error struct{ Code string }

func (e *Error) Error() string { return e.Code }

// ErrEntryCapExceeded and ErrOutputCeilingExceeded are the two fail-closed
// codes D-04's caps require.
var (
	ErrEntryCapExceeded      = &Error{Code: "debugmap.entry_cap_exceeded"}
	ErrOutputCeilingExceeded = &Error{Code: "debugmap.output_ceiling_exceeded"}
	ErrDeadlineExceeded      = &Error{Code: "debugmap.deadline_exceeded"}
)

// Entry is one joined lineage fact: a source span, the typed-core function
// identity it belongs to, and the operation/point/edge identity that source
// construct produced. ID is a function-local semantic ordinal
// ("{functionID}:debug:{n}"), never a source byte offset, matching the
// existing convention (core.Block/Edge/LoanEndpoint IDs) — SourceSpan is
// carried as data, not as identity.
type Entry struct {
	ID           string          `json:"id"`
	CoreID       string          `json:"core_id"`
	SourceSpan   diagnostic.Span `json:"source_span"`
	OperationID  string          `json:"operation_id,omitempty"`
	PointID      string          `json:"point_id,omitempty"`
	Kind         string          `json:"kind"`
	Availability Availability    `json:"availability"`
}

// Map is the schema-versioned side table Build produces.
type Map struct {
	Schema  string  `json:"schema"`
	Entries []Entry `json:"entries"`
}

// kindFor names an entry by the law it joins, following the repo's existing
// convention of naming things after what they mean rather than an internal
// code.
func kindFor(operationKind core.OperationKind) string {
	switch operationKind {
	case core.OpMove:
		return "move"
	case core.OpBorrowShared:
		return "borrow_shared"
	case core.OpBorrowExclusive:
		return "borrow_exclusive"
	case core.OpCopy:
		return "copy"
	case core.OpReturn:
		return "return"
	default:
		return "unknown"
	}
}

// Build independently joins astProgram's own span-bearing bindings against
// checkedProgram's already-lowered operations, for every straight-line
// function and every branch (match-arm-body) function checkedProgram
// carries. It relies on the same 1:1 positional correspondence check.go's
// own analyzeStraightLine/analyzeArmBody establish between
// ast.LinearBody.Bindings[i] and core.LinearOperation at index i (verified
// directly against check.go before this package was written, not assumed) —
// it does not re-run or import check, so the join is read-only over two
// already-produced artifacts, never a second admission decision.
//
// Build is deterministic, bounded, and counts one unit of work per entry.
// Exceeding MaxEntries fails closed with ErrEntryCapExceeded rather than
// truncating. ctx's deadline is checked once per function; a canceled or
// expired context fails closed with ErrDeadlineExceeded.
func Build(ctx context.Context, astProgram ast.Program, checkedProgram core.Program) (Map, int, error) {
	result := Map{Schema: Schema, Entries: make([]Entry, 0, 16)}
	work := 0

	astByName := make(map[string]ast.FuncDecl, len(astProgram.Funcs))
	for _, function := range astProgram.Funcs {
		astByName[function.Name] = function
	}

	appendEntry := func(functionID, coreID string, span diagnostic.Span, operationID, pointID, kind string) error {
		if ctx.Err() != nil {
			return ErrDeadlineExceeded
		}
		if len(result.Entries) >= MaxEntries {
			return ErrEntryCapExceeded
		}
		result.Entries = append(result.Entries, Entry{
			ID:     fmt.Sprintf("%s:debug:%d", functionID, len(result.Entries)),
			CoreID: coreID, SourceSpan: span, OperationID: operationID, PointID: pointID,
			Kind: kind, Availability: Available,
		})
		work++
		return nil
	}

	for _, function := range checkedProgram.Functions {
		astFunction, ok := astByName[function.Name]
		if !ok {
			continue
		}
		if ctx.Err() != nil {
			return Map{}, work, ErrDeadlineExceeded
		}

		switch {
		case function.Linear != nil && function.Match == nil:
			// Straight-line body: ast bindings and core operations correspond
			// 1:1 by index, with one trailing OpReturn beyond the bindings.
			if astFunction.Body.Linear == nil {
				continue
			}
			if err := joinBindings(appendEntry, function.ID, astFunction.Body.Linear.Bindings, astFunction.Body.Linear.Span, function.Linear.Operations); err != nil {
				return Map{}, work, err
			}
		case function.Match != nil:
			// Branch body: each arm with a Body joins its own bindings against
			// the operations recorded in the arm's own block.
			operationsByID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
			for _, operation := range function.Linear.Operations {
				operationsByID[operation.ID] = operation
			}
			blocksByID := make(map[string]core.Block, len(function.Linear.Blocks))
			for _, block := range function.Linear.Blocks {
				blocksByID[block.ID] = block
			}
			for armIndex, arm := range function.Match.Arms {
				if arm.BlockID == "" || armIndex >= len(astFunction.Body.Arms) {
					continue
				}
				astArm := astFunction.Body.Arms[armIndex]
				if astArm.Body == nil {
					continue
				}
				block, found := blocksByID[arm.BlockID]
				if !found {
					continue
				}
				armOperations := make([]core.LinearOperation, 0, len(block.OperationIDs))
				for _, operationID := range block.OperationIDs {
					if operation, ok := operationsByID[operationID]; ok {
						armOperations = append(armOperations, operation)
					}
				}
				// checkBranch prepends one implicit alias OpCopy (the scrutinee
				// copied into a fresh per-arm place, 03-01's per-arm aliasing
				// decision) ahead of the arm body's own bindings -- verified
				// directly against checkBranch before this package was written.
				// It corresponds to no source span, so it is skipped rather than
				// joined to a fabricated one; the remaining operations line up
				// 1:1 with the arm's own bindings exactly like the straight-line
				// case, with one trailing OpReturn.
				if len(armOperations) > 0 {
					armOperations = armOperations[1:]
				}
				if err := joinBindings(appendEntry, function.ID, astArm.Body.Bindings, astArm.Body.Span, armOperations); err != nil {
					return Map{}, work, err
				}
			}
		}
	}

	encoded, err := json.Marshal(result.Entries)
	if err != nil {
		return Map{}, work, err
	}
	if len(encoded) > MaxOutputBytes {
		return Map{}, work, ErrOutputCeilingExceeded
	}
	return result, work, nil
}

type appendFunc func(functionID, coreID string, span diagnostic.Span, operationID, pointID, kind string) error

func joinBindings(appendEntry appendFunc, functionID string, bindings []ast.Binding, resultSpan diagnostic.Span, operations []core.LinearOperation) error {
	for index, binding := range bindings {
		if index >= len(operations) {
			break
		}
		operation := operations[index]
		if err := appendEntry(functionID, functionID, binding.Span, operation.ID, operation.PointID, kindFor(operation.Kind)); err != nil {
			return err
		}
	}
	if len(operations) > len(bindings) {
		final := operations[len(operations)-1]
		if final.Kind == core.OpReturn {
			if err := appendEntry(functionID, functionID, resultSpan, final.ID, final.PointID, kindFor(final.Kind)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Resolve answers the honest-absence half of D-04: given an operation or
// point identity read off a diagnostic.Diagnostic cause or an
// interp.Execution event (both already carry core-produced IDs, e.g.
// "{functionID}:op:{n}" or the same with a ":event"/":event:returned" event
// suffix stripped by the caller), it returns the joined Entry if the map
// captured one, and an honest NotCaptured entry — never a guess, never a
// fabricated span — otherwise.
func Resolve(built Map, operationID string) Entry {
	for _, entry := range built.Entries {
		if entry.OperationID == operationID {
			return entry
		}
	}
	return Entry{OperationID: operationID, Availability: NotCaptured}
}

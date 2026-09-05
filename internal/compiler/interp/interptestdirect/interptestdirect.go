// Package interptestdirect holds interp.RunLinearBlockDirect (task
// 04-07-02's narrow, test-only entry point) and its supporting helpers,
// moved out of the interp package proper (WR-01/WR-02 of the Phase 4 code
// review): the function's own doc comment states plainly that it "is not
// used by any production CLI path", so it should not ship as a discoverable,
// callable, exported member of the interp package's real production API
// surface. Living in its own package makes that boundary a compile-time
// fact -- a caller has to explicitly import interptestdirect, not merely
// interp -- rather than a disclaimer in a doc comment alone.
//
// This package duplicates a handful of small interp-package-private
// helpers (findFunction, ownedEvent, liveResourceList) rather than
// exporting them from interp to share, matching this project's own
// established convention elsewhere (D-12: independently re-deriving shared
// facts across packages instead of coupling packages through shared
// helpers) -- see interp.foreignFailureLiteralDirect's own comment for the
// precedent this package continues.
package interptestdirect

import (
	"fmt"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

// RunLinearBlockDirect is task 04-07-02's narrow, test-only entry point:
// unlike interp.Run (which always begins at ":block:entry" and always
// simulates a foreign call's success, since the interpreter cannot actually
// call C), this executes a single, already-validated block's own
// operations directly, given the acquisitions already considered live on
// entry. It exists because the project's frozen foreign translation unit
// always succeeds at real runtime (D-04-10), so a resource-lifecycle
// function's OWN failure block is otherwise unreachable through Run's
// public entry point. It supports exactly the shape check.go ever produces
// for such a block -- zero or more OpRelease operations, then a single
// terminal OpFail -- matching cgen.go's emitForeignReleasesAndFail's own
// supported shape, so the SAME block compiles to real, genuinely
// executable C (dead code today given the frozen TU's real success-only
// behavior, but valid C) that a test-only foreign object double can make
// live. This is not used by any production CLI path; it exists solely so
// the interpreter's real OpRelease/OpFail event-emission logic can be
// compared, genuinely executed, against the real native build's genuinely
// executed same block.
func RunLinearBlockDirect(program core.Program, functionName, blockID string, precedingCallIDs []string, failingCallID string) (execution.Execution, error) {
	function, ok := findFunction(program, functionName)
	if !ok {
		return execution.Execution{}, fmt.Errorf("function %q is absent from checked core", functionName)
	}
	if function.Linear == nil {
		return execution.Execution{}, fmt.Errorf("function %q has no linear body", function.ID)
	}
	operationsByID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operationsByID[operation.ID] = operation
	}
	var target *core.Block
	for index := range function.Linear.Blocks {
		if function.Linear.Blocks[index].ID == blockID {
			target = &function.Linear.Blocks[index]
			break
		}
	}
	if target == nil {
		return execution.Execution{}, fmt.Errorf("function %q has no block %q", function.ID, blockID)
	}
	live := make(map[string]bool, len(precedingCallIDs))
	var liveOrder []string
	events := make([]execution.Event, 0, len(precedingCallIDs)+len(target.OperationIDs))
	// precedingCallIDs names the OpForeignCall operations that, per this
	// engine's own documented always-succeeds stub, already completed
	// (and their acquisitions became live) before control reached blockID
	// -- mirroring interp's own runLinearBlocks OpForeignCall case exactly,
	// so the resulting event sequence is byte-identical in shape to what
	// Run would have produced had it genuinely reached this block through
	// the real walker, not merely the release/fail suffix.
	for _, callID := range precedingCallIDs {
		call, known := operationsByID[callID]
		if !known || call.Kind != core.OpForeignCall {
			return execution.Execution{}, fmt.Errorf("preceding call %q is not a known OpForeignCall", callID)
		}
		events = append(events, execution.Event{
			Schema: execution.Schema1, ID: call.ID + ":event", Kind: "foreign.called", FunctionID: function.ID,
			SourcePlace: call.SourceID, TargetPlace: call.TargetID, TypeID: call.TypeID,
		})
		live[callID] = true
		liveOrder = append(liveOrder, callID)
	}
	// cgen.go's emitLinearForeign records "foreign.called" unconditionally,
	// immediately after the real call returns, BEFORE inspecting .ok -- so
	// the failing call's own attempt is recorded exactly like any other,
	// just never marked live (its acquisition never completed).
	if failingCallID != "" {
		failingCall, known := operationsByID[failingCallID]
		if !known || failingCall.Kind != core.OpForeignCall {
			return execution.Execution{}, fmt.Errorf("failing call %q is not a known OpForeignCall", failingCallID)
		}
		events = append(events, execution.Event{
			Schema: execution.Schema1, ID: failingCall.ID + ":event", Kind: "foreign.called", FunctionID: function.ID,
			SourcePlace: failingCall.SourceID, TargetPlace: failingCall.TargetID, TypeID: failingCall.TypeID,
		})
	}
	for _, operationID := range target.OperationIDs {
		operation, known := operationsByID[operationID]
		if !known {
			return execution.Execution{}, fmt.Errorf("block %q references unknown operation %q", blockID, operationID)
		}
		switch operation.Kind {
		case core.OpRelease:
			live[operation.ReleasesOperationID] = false
			events = append(events, ownedEvent(function, operation, "resource.released"))
		case core.OpFail:
			literal, literalErr := foreignFailureLiteralDirect(program, function)
			if literalErr != nil {
				return execution.Execution{}, literalErr
			}
			events = append(events, execution.Event{
				Schema: execution.Schema1, ID: operation.ID + ":event:failed", Kind: "function.failed",
				FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
			})
			return execution.Execution{Schema: execution.Schema1, Outcome: execution.Outcome{Kind: "typed_failure", Value: literal}, Events: events, LiveResources: liveResourceList(live, liveOrder)}, nil
		default:
			return execution.Execution{}, fmt.Errorf("block %q has an unsupported operation kind %q for direct execution", blockID, operation.Kind)
		}
	}
	return execution.Execution{}, fmt.Errorf("block %q has no terminal OpFail", blockID)
}

// foreignFailureLiteralDirect independently re-derives cgen.go's own
// foreignFailureLiteral and interp.go's own foreignFailureLiteralDirect
// (not imported -- matching this project's established convention of
// independently re-deriving shared facts across packages rather than
// sharing helpers, D-12): the err edge's payload is a place of an ordinary
// declared nullary ADT, but this phase has no case-analysis syntax to pick
// a specific alternative at the failure site, so every engine reports the
// SAME compile-time-known first alternative of the function's declared
// Fails type.
func foreignFailureLiteralDirect(program core.Program, function core.Function) (string, error) {
	if function.ForeignContract == nil {
		return "", fmt.Errorf("function %q has no foreign contract", function.ID)
	}
	for _, dataType := range program.DataTypes {
		if dataType.Name == function.ForeignContract.Fails {
			if len(dataType.Alternatives) == 0 {
				return "", fmt.Errorf("foreign failure type %q has no alternatives", dataType.Name)
			}
			return dataType.Alternatives[0], nil
		}
	}
	return "", fmt.Errorf("foreign failure type %q is not declared", function.ForeignContract.Fails)
}

// findFunction is interp.findFunction, independently re-derived here (see
// package doc) rather than exported and shared.
func findFunction(program core.Program, name string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.Name == name {
			return function, true
		}
	}
	return core.Function{}, false
}

// ownedEvent is interp.ownedEvent, independently re-derived here (see
// package doc) rather than exported and shared.
func ownedEvent(function core.Function, operation core.LinearOperation, kind string) execution.Event {
	return execution.Event{
		Schema: execution.Schema1, ID: operation.ID + ":event", Kind: kind, FunctionID: function.ID,
		SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
	}
}

// liveResourceList is interp.liveResourceList, independently re-derived
// here (see package doc) rather than exported and shared: it projects the
// live-tracking map into the deterministic slice execution.Execution
// carries, acquisition op IDs still marked live in first-acquired order,
// returning an empty (never nil) slice when nothing is live.
func liveResourceList(live map[string]bool, order []string) []string {
	result := []string{}
	for _, id := range order {
		if live[id] {
			result = append(result, id)
		}
	}
	return result
}

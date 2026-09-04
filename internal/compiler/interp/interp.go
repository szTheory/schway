package interp

import (
	"fmt"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

const Schema = execution.Schema0

type Outcome = execution.Outcome
type Event = execution.Event
type Execution = execution.Execution

func Run(program core.Program, functionName, input string) (Execution, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return Execution{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	function, ok := findFunction(program, functionName)
	if !ok {
		return Execution{}, fmt.Errorf("function %q is absent from checked core", functionName)
	}
	if !function.HasClosedBody() {
		return Execution{}, fmt.Errorf("function %q has invalid body union", functionName)
	}
	if function.Linear != nil && function.Match == nil {
		return runLinear(function, input)
	}
	for _, arm := range function.Match.Arms {
		if arm.Pattern != input {
			continue
		}
		if arm.BlockID == "" {
			event := Event{
				Schema: Schema, ID: arm.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, Input: input, Output: arm.Value,
			}
			return Execution{
				Schema: Schema, Outcome: Outcome{Kind: "returned", Value: arm.Value},
				Events: []Event{event}, LiveResources: []string{},
			}, nil
		}
		return runBranchArm(function, arm, input)
	}
	return Execution{}, fmt.Errorf("checked match %q has no arm for %q", function.Match.ID, input)
}

// runBranchArm walks only the operations of the selected arm's block, in
// core order, and emits their events — the D-12a consumer this phase adds
// alongside check, corevalidate, and cgen. The unselected arms' operations
// are never visited, so they produce no events (03-01-02's behavior clause).
func runBranchArm(function core.Function, arm core.MatchArm, input string) (Execution, error) {
	var block core.Block
	found := false
	for _, candidate := range function.Linear.Blocks {
		if candidate.ID == arm.BlockID {
			block, found = candidate, true
			break
		}
	}
	if !found {
		return Execution{}, fmt.Errorf("branch arm %q references unknown block %q", arm.ID, arm.BlockID)
	}
	operations := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operations[operation.ID] = operation
	}
	values := map[string]string{function.Parameter.ID: input}
	events := make([]Event, 0, len(block.OperationIDs))
	for _, operationID := range block.OperationIDs {
		operation, known := operations[operationID]
		if !known {
			return Execution{}, fmt.Errorf("block %q references unknown operation %q", block.ID, operationID)
		}
		value, initialized := values[operation.SourceID]
		if !initialized {
			return Execution{}, fmt.Errorf("operation %q reads uninitialized place %q", operation.ID, operation.SourceID)
		}
		switch operation.Kind {
		case core.OpCopy:
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.copied"))
		case core.OpMove:
			delete(values, operation.SourceID)
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.transferred"))
		case core.OpBorrowShared:
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.borrowed"))
		case core.OpBorrowExclusive:
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.borrowed_exclusive"))
		case core.OpReturn:
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
			})
			return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "returned", Value: value}, Events: events, LiveResources: []string{}}, nil
		default:
			return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	return Execution{}, fmt.Errorf("branch arm %q has no return operation", arm.ID)
}

func CanonicalBytes(execution Execution) ([]byte, error) {
	return execution2bytes(execution)
}

func execution2bytes(value Execution) ([]byte, error) { return execution.CanonicalBytes(value) }

func runLinear(function core.Function, input string) (Execution, error) {
	values := map[string]string{function.Parameter.ID: input}
	events := make([]Event, 0, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		value, initialized := values[operation.SourceID]
		if !initialized {
			return Execution{}, fmt.Errorf("operation %q reads uninitialized place %q", operation.ID, operation.SourceID)
		}
		switch operation.Kind {
		case core.OpCopy:
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.copied"))
		case core.OpMove:
			delete(values, operation.SourceID)
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.transferred"))
		case core.OpBorrowShared:
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.borrowed"))
		case core.OpBorrowExclusive:
			values[operation.TargetID] = value
			events = append(events, ownedEvent(function, operation, "value.borrowed_exclusive"))
		case core.OpReturn:
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
			})
			return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "returned", Value: value}, Events: events, LiveResources: []string{}}, nil
		default:
			return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	return Execution{}, fmt.Errorf("linear body %q has no return operation", function.Linear.ID)
}

func ownedEvent(function core.Function, operation core.LinearOperation, kind string) Event {
	return Event{
		Schema: execution.Schema1, ID: operation.ID + ":event", Kind: kind, FunctionID: function.ID,
		SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
	}
}

func findFunction(program core.Program, name string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.Name == name {
			return function, true
		}
	}
	return core.Function{}, false
}

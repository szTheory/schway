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
		case core.OpForeignCall:
			// No match arm can produce a foreign call this phase (checkBranch
			// does not admit `try` inside an arm body) -- this case exists
			// solely so control:kind.exhaustive_dispatch's six-site table
			// finds every kind handled at every site, per D-04-22.
			values[operation.TargetID] = value
			values[operation.ErrTargetID] = "err"
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event", Kind: "foreign.called", FunctionID: function.ID,
				SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
			})
		case core.OpFail:
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event:failed", Kind: "function.failed",
				FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
			})
			return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "typed_failure", Value: value}, Events: events, LiveResources: []string{}}, nil
		case core.OpReturn:
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
			})
			return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "returned", Value: value}, Events: events, LiveResources: []string{}}, nil
		case core.OpRelease:
			// No match arm can produce a release this phase either -- named
			// here for the same six-site exhaustive-dispatch reason as
			// OpForeignCall above.
			events = append(events, ownedEvent(function, operation, "resource.released"))
		case core.OpDefect:
			// D-04-15: a real, reachable, abort-only terminal outcome. It
			// performs no release and reads no live-resource accounting this
			// phase (no arm can carry a foreign acquisition), so
			// LiveResources stays the same empty-but-never-nil shape every
			// other arm terminator here uses.
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event:defected", Kind: "function.defected",
				FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID, Output: operation.Reason,
			})
			return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""}, Events: events, LiveResources: []string{}}, nil
		default:
			return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
		}
	}
	return Execution{}, fmt.Errorf("branch arm %q has no return operation", arm.ID)
}

// runLinearBlocks walks a fallible-call function's Blocks/Edges (D-04-04),
// starting at the entry block, executing each block's operations in order.
// An OpForeignCall never terminates its block by itself -- per Claude's
// Discretion (04-PATTERNS Pattern 3 note), the interpreter cannot actually
// call C, so it models the call as a fixed literal-outcome stub that always
// succeeds this phase, unconditionally following the ok edge. Proving
// engine disagreement on a genuine failure path is deferred to a later
// plan; this phase proves the dispatch shape exists and both engines agree
// on the success path the shipped fixture actually exercises.
func runLinearBlocks(function core.Function, input string) (Execution, error) {
	operations := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operations[operation.ID] = operation
	}
	blocks := make(map[string]core.Block, len(function.Linear.Blocks))
	for _, block := range function.Linear.Blocks {
		blocks[block.ID] = block
	}
	edges := make(map[string]core.Edge, len(function.Linear.Edges))
	for _, edge := range function.Linear.Edges {
		edges[edge.ID] = edge
	}
	// tracked names exactly the acquisitions this function's own OpRelease
	// operations discharge somewhere (D-04-07): computed once, up front, so
	// a function with no OpRelease at all (the 04-01 tracer shape, and a
	// `discard`'s own untracked acquisition) never populates LiveResources,
	// preserving that shape's pre-plan-02 empty-slice behavior exactly.
	tracked := make(map[string]bool)
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpRelease && operation.ReleasesOperationID != "" {
			tracked[operation.ReleasesOperationID] = true
		}
	}

	values := map[string]string{function.Parameter.ID: input}
	events := make([]Event, 0, len(function.Linear.Operations))
	// live is Phase 4 plan 02's real resource accounting (D-04-07),
	// replacing the hardcoded empty LiveResources slice: a resource is
	// tracked live the moment its acquisition's ok edge is taken, and
	// discharged when the OpRelease naming it (by ReleasesOperationID) runs.
	// liveOrder preserves first-acquired order so a leaked-resource report
	// is deterministic rather than a function of map iteration.
	live := make(map[string]bool)
	var liveOrder []string
	currentID := function.ID + ":block:entry"
	for {
		block, known := blocks[currentID]
		if !known {
			return Execution{}, fmt.Errorf("block %q is unknown", currentID)
		}
		var forked *core.LinearOperation
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
			case core.OpForeignCall:
				values[operation.TargetID] = value
				values[operation.ErrTargetID] = "err"
				events = append(events, Event{
					Schema: execution.Schema1, ID: operation.ID + ":event", Kind: "foreign.called", FunctionID: function.ID,
					SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
				})
				// The interpreter always simulates success this phase (a
				// documented discretionary stub -- it cannot actually call
				// C), so the acquisition unconditionally becomes live here,
				// on the ok path, exactly mirroring D-04-07's rule that only
				// a completed acquisition is ever tracked for release.
				if tracked[operation.ID] && !live[operation.ID] {
					live[operation.ID] = true
					liveOrder = append(liveOrder, operation.ID)
				}
				forkedOperation := operation
				forked = &forkedOperation
			case core.OpRelease:
				live[operation.ReleasesOperationID] = false
				events = append(events, ownedEvent(function, operation, "resource.released"))
			case core.OpReturn:
				events = append(events, Event{
					Schema: execution.Schema1, ID: operation.ID + ":event:returned", Kind: "function.returned",
					FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
				})
				return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "returned", Value: value}, Events: events, LiveResources: liveResourceList(live, liveOrder)}, nil
			case core.OpFail:
				events = append(events, Event{
					Schema: execution.Schema1, ID: operation.ID + ":event:failed", Kind: "function.failed",
					FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID,
				})
				return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "typed_failure", Value: value}, Events: events, LiveResources: liveResourceList(live, liveOrder)}, nil
			case core.OpDefect:
				// D-04-15/D-04-18: a defect path populates LiveResources with
				// every still-live acquisition and performs no release -- the
				// SAME liveResourceList accounting OpFail/OpReturn use, just
				// never followed by a release call.
				events = append(events, Event{
					Schema: execution.Schema1, ID: operation.ID + ":event:defected", Kind: "function.defected",
					FunctionID: function.ID, SourcePlace: operation.SourceID, TypeID: operation.TypeID, Output: operation.Reason,
				})
				return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""}, Events: events, LiveResources: liveResourceList(live, liveOrder)}, nil
			default:
				return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
			}
		}
		if forked != nil {
			edge, known := edges[forked.OkEdgeID]
			if !known {
				return Execution{}, fmt.Errorf("foreign call %q references unknown ok edge %q", forked.ID, forked.OkEdgeID)
			}
			currentID = edge.ToBlockID
			continue
		}
		if len(block.Successors) == 1 {
			currentID = block.Successors[0]
			continue
		}
		return Execution{}, fmt.Errorf("block %q has no terminator and an ambiguous successor set", block.ID)
	}
}

// liveResourceList projects the live-tracking map into the deterministic
// slice Execution.LiveResources carries: acquisition op IDs still marked
// live, in first-acquired order. Returns an empty (never nil) slice when
// nothing is live, matching every pre-plan-02 terminal record's shape.
func liveResourceList(live map[string]bool, order []string) []string {
	result := []string{}
	for _, id := range order {
		if live[id] {
			result = append(result, id)
		}
	}
	return result
}

func CanonicalBytes(execution Execution) ([]byte, error) {
	return execution2bytes(execution)
}

func execution2bytes(value Execution) ([]byte, error) { return execution.CanonicalBytes(value) }

func runLinear(function core.Function, input string) (Execution, error) {
	// A straight-line body that carries Blocks/Edges is Phase 4's fallible-
	// call shape (D-04-04): the flat Operations list alone is not enough to
	// know which edge to follow, so it is walked block-by-block instead.
	// Every pre-Phase-4 linear (non-Match) function leaves Blocks empty --
	// only checkFallibleLinear ever populates it for a Match-less function --
	// so this branch changes nothing for any existing program.
	if len(function.Linear.Blocks) > 0 {
		return runLinearBlocks(function, input)
	}
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

package executionpeer

import (
	"errors"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
)

func TestPhase35PureOperationCausalControls(t *testing.T) {
	kinds := []core.OperationKind{
		core.OpBitXorU64,
		core.OpBitAndU64,
		core.OpBitOrU64,
		core.OpShiftLeftU64,
		core.OpShiftRightU64,
	}
	if got := eventFreeOperationKinds(); len(got) != len(kinds) {
		t.Fatalf("event-free inventory has %d kinds, want %d", len(got), len(kinds))
	}

	program := phase35PureProgram(kinds)
	root, err := execution.FormatInvocation("entry", nil)
	if err != nil {
		t.Fatal(err)
	}
	clean := execution.Execution{
		Schema:  execution.Schema2,
		Outcome: execution.Outcome{Kind: execution.OutcomeReturned},
		Events: []execution.Event{{
			Schema: execution.Schema2, ID: "entry:return:event:returned", Kind: "function.returned",
			FunctionID: "entry", Invocation: root,
		}},
	}
	if err := Validate(program, clean); err != nil {
		t.Fatalf("event-free operations should admit the function's return event: %v", err)
	}

	for _, kind := range kinds {
		kind := kind
		opID := "entry:" + string(kind)
		for _, control := range []struct {
			name string
			id   string
			kind string
		}{
			{name: "normal-step", id: opID + ":event", kind: "value.copied"},
			{name: "defect", id: opID + ":event:defected", kind: "function.defected"},
		} {
			t.Run(string(kind)+"/"+control.name, func(t *testing.T) {
				document := phase35ForgedDocument(root, opID, control.id, control.kind)
				assertPhase35CausalCode(t, Validate(program, document), "pure_operation_event")

				indexed, err := buildIndex(program)
				if err != nil {
					t.Fatal(err)
				}
				withoutKind := *indexed
				withoutKind.eventFreeKinds = make(map[core.OperationKind]bool, len(indexed.eventFreeKinds)-1)
				for indexedKind, present := range indexed.eventFreeKinds {
					if present && indexedKind != kind {
						withoutKind.eventFreeKinds[indexedKind] = true
					}
				}
				if _, stillPresent := withoutKind.eventFreeKinds[kind]; stillPresent {
					t.Fatal("negative control failed to remove operation from event-free inventory")
				}
				assertPhase35CausalCode(t, withoutKind.validate(document), "unknown_kind")

				// A fresh public validation confirms the real inventory remains intact.
				assertPhase35CausalCode(t, Validate(program, document), "pure_operation_event")
			})
		}
	}
}

func phase35PureProgram(kinds []core.OperationKind) core.Program {
	operations := make([]core.LinearOperation, 0, len(kinds)+1)
	for _, kind := range kinds {
		operations = append(operations, core.LinearOperation{
			ID: "entry:" + string(kind), Kind: kind,
			SourceID: "entry:left", RightID: "entry:right", TargetID: "entry:result", TypeID: "u64",
		})
	}
	operations = append(operations, core.LinearOperation{ID: "entry:return", Kind: core.OpReturn, SourceID: "entry:result", TypeID: "u64"})
	return core.Program{Functions: []core.Function{{
		ID: "entry", Name: "entry", Parameter: core.Parameter{ID: "entry:arg", Type: "U64"}, ReturnType: "U64",
		Linear: &core.LinearBody{
			ID:    "entry:body",
			Types: []core.TypeFact{{ID: "u64", Shape: core.TypeRef{Constructor: "U64"}}},
			Places: []core.Place{
				{ID: "entry:left", TypeID: "u64"},
				{ID: "entry:right", TypeID: "u64"},
				{ID: "entry:result", TypeID: "u64"},
			},
			Operations: operations,
		},
	}}}
}

func phase35ForgedDocument(invocation, operationID, eventID, kind string) execution.Execution {
	outcome := execution.OutcomeReturned
	if kind == "function.defected" {
		outcome = execution.OutcomeDefect
	}
	return execution.Execution{
		Schema:  execution.Schema2,
		Outcome: execution.Outcome{Kind: outcome},
		Events: []execution.Event{{
			Schema: execution.Schema2, ID: eventID, Kind: kind,
			FunctionID: "entry", Invocation: invocation,
		}},
	}
}

func assertPhase35CausalCode(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Validate() = nil, want executionpeer.%s", want)
	}
	var peerError *Error
	if !errors.As(err, &peerError) {
		t.Fatalf("Validate() error %T = %v, want *Error", err, err)
	}
	if got := peerError.Code(); got != "executionpeer."+want {
		t.Fatalf("Validate() code = %q, want executionpeer.%s", got, want)
	}
}

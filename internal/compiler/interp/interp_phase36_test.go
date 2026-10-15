package interp_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase36RepeatedCallSource = `module phase36.runtime_calls
export { fn main }
fn bump(input: U64) -> U64 { let result = input + 2 result }
fn main(input: U64) -> U64 {
  var total = 0
  var index = 0
  while index < 3 {
    total = bump(total + index)
    index = index + 1
  }
  total
}
`

func TestPhase36InterpreterRepeatedCallTracer(t *testing.T) {
	checked := session.Check([]byte(phase36RepeatedCallSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("repeated-call source diagnostics: %+v", checked.Diagnostics)
	}
	main := phase36Function(t, checked.Program, "main")
	bump := phase36Function(t, checked.Program, "bump")
	call, returns := phase36CallAndReturn(t, main, bump)
	result, err := interp.RunApplication(checked.Program, "main", "0")
	if err != nil {
		t.Fatalf("RunApplication: %v", err)
	}
	if result.Execution.Outcome.Kind != execution.OutcomeReturned || result.Execution.Outcome.Value != "9" || string(result.Output) != "9\n" {
		t.Fatalf("actual application result = outcome %+v, output %q; want returned 9 and 9 newline", result.Execution.Outcome, result.Output)
	}
	if result.Execution.Schema != execution.Schema2 {
		t.Fatalf("execution schema = %q, want %q for invocation evidence", result.Execution.Schema, execution.Schema2)
	}
	root, err := execution.FormatInvocation(main.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantChildren := make([]string, 3)
	for ordinal := range wantChildren {
		wantChildren[ordinal], err = execution.FormatInvocation(main.ID, []execution.InvocationSegment{{OpCallID: call.ID, Ordinal: ordinal}})
		if err != nil {
			t.Fatal(err)
		}
	}
	var called []execution.Event
	var terminals []execution.Event
	for _, event := range result.Execution.Events {
		if event.ID == call.ID+":event:called" {
			called = append(called, event)
		}
		if event.ID == returns.ID+":event:returned" {
			terminals = append(terminals, event)
		}
	}
	if len(called) != len(wantChildren) || len(terminals) != len(wantChildren) {
		t.Fatalf("activation event counts: called=%d returned=%d, want 3 each", len(called), len(terminals))
	}
	for index, childPath := range wantChildren {
		if called[index].Invocation != root || called[index].CalleeFunctionID != bump.ID || called[index].Occurrence != uint64(index) {
			t.Fatalf("call event[%d] = %+v, want parent %q, callee %q, occurrence %d", index, called[index], root, bump.ID, index)
		}
		if terminals[index].Invocation != childPath || terminals[index].FunctionID != bump.ID || terminals[index].Occurrence != 0 {
			t.Fatalf("callee terminal[%d] = %+v, want child %q with a fresh event occurrence", index, terminals[index], childPath)
		}
	}
}

func phase36Function(t *testing.T, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("checked function %q is absent", name)
	return core.Function{}
}

func phase36CallAndReturn(t *testing.T, caller, callee core.Function) (core.LinearOperation, core.LinearOperation) {
	t.Helper()
	var call, returned core.LinearOperation
	for _, operation := range caller.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == callee.ID {
			if call.ID != "" {
				t.Fatalf("caller has more than one static call to %q", callee.ID)
			}
			call = operation
		}
	}
	for _, operation := range callee.Linear.Operations {
		if operation.Kind == core.OpReturn {
			returned = operation
		}
	}
	if call.ID == "" || returned.ID == "" {
		t.Fatalf("call/return operations missing: call=%s return=%s", call.ID, returned.ID)
	}
	return call, returned
}


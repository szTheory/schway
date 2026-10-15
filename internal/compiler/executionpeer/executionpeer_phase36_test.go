package executionpeer_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
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

func TestPhase36DynamicCallOrdinalControls(t *testing.T) {
	program, run, call, bump, main := phase36RepeatedCallRun(t)
	if run.Execution.Outcome.Kind != execution.OutcomeReturned || run.Execution.Outcome.Value != "9" || string(run.Output) != "9\n" {
		t.Fatalf("interpreter result = %+v, output %q; want returned 9", run.Execution.Outcome, run.Output)
	}
	root, err := execution.FormatInvocation(main.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	for ordinal := 0; ordinal < 3; ordinal++ {
		child, err := execution.FormatInvocation(main.ID, []execution.InvocationSegment{{OpCallID: call.ID, Ordinal: ordinal}})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, event := range run.Execution.Events {
			if event.FunctionID == bump.ID && event.Invocation == child {
				found = true
			}
		}
		if !found {
			t.Fatalf("no observed callee event at expected child activation %q", child)
		}
	}
	if err := executionpeer.Validate(program, run.Execution); err != nil {
		t.Fatalf("independent causal peer refused repeated dynamic activations rooted at %q: %v", root, err)
	}
}

func TestPhase36CallCausalAttributionControls(t *testing.T) {
	program, run, call, bump, _ := phase36RepeatedCallRun(t)
	if err := executionpeer.Validate(program, run.Execution); err != nil {
		t.Fatalf("peer rejected the unmodified interpreter trace: %v", err)
	}
	var called, terminal int
	for _, event := range run.Execution.Events {
		if event.ID == call.ID+":event:called" {
			called++
			if event.CalleeFunctionID != bump.ID || event.FunctionID == bump.ID {
				t.Fatalf("call attribution = %+v, want caller-to-%q edge", event, bump.ID)
			}
		}
		if event.Kind == "function.returned" && event.FunctionID == bump.ID {
			terminal++
			parsed, err := execution.ParseInvocation(event.Invocation)
			if err != nil || len(parsed.Segments) != 1 || parsed.Segments[0].OpCallID != call.ID || parsed.Segments[0].Ordinal != terminal-1 {
				t.Fatalf("callee terminal does not match dynamic call activation: event=%+v parsed=%+v err=%v", event, parsed, err)
			}
		}
	}
	if called != 3 || terminal != 3 {
		t.Fatalf("causal call/terminal observations = %d/%d, want 3/3", called, terminal)
	}
}

func phase36RepeatedCallRun(t *testing.T) (core.Program, interp.ApplicationResult, core.LinearOperation, core.Function, core.Function) {
	t.Helper()
	checked := session.Check([]byte(phase36RepeatedCallSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("repeated-call source diagnostics: %+v", checked.Diagnostics)
	}
	var main, bump core.Function
	for _, function := range checked.Program.Functions {
		switch function.Name {
		case "main":
			main = function
		case "bump":
			bump = function
		}
	}
	if main.ID == "" || bump.ID == "" {
		t.Fatalf("functions absent from checked program: main=%q bump=%q", main.ID, bump.ID)
	}
	var call core.LinearOperation
	for _, operation := range main.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == bump.ID {
			call = operation
		}
	}
	if call.ID == "" {
		t.Fatal("main has no checked direct call to bump")
	}
	result, err := interp.RunApplication(checked.Program, "main", "0")
	if err != nil {
		t.Fatalf("RunApplication: %v", err)
	}
	return checked.Program, result, call, bump, main
}

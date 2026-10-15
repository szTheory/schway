package executionpeer_test

import (
	"errors"
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

const phase36CausalCallSource = `module phase36.causal_calls
export { fn main }
fn bump(input: U64) -> U64 { let result = input + 1 result }
fn decoy(input: U64) -> U64 { let result = input + 2 result }
fn main(input: U64) -> U64 {
  var total = 0
  var index = 0
  while index < 3 {
    total = bump(total + index)
    index = index + 1
  }
  let last = bump(total)
  let finish = decoy(last)
  finish
}
`

const phase36CausalOverflowSource = `module phase36.causal_overflow
export { fn main }
fn overflow_step(input: U64) -> U64 { let result = input + 1 result }
fn main(input: U64) -> U64 {
  var value = input
  var index = 0
  while index < 2 {
    value = overflow_step(value)
    index = index + 1
  }
  value
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
	phase36CallCausalForgeryControls(t)
}

func phase36CallCausalForgeryControls(t *testing.T) {
	t.Helper()
	program, clean, main, bump, decoy := phase36CausalCallRun(t)
	if err := executionpeer.Validate(program, clean.Execution); err != nil {
		t.Fatalf("fresh positive trace rejected: %v", err)
	}
	if err := executionpeer.Validate(program, clonePhase36Execution(clean.Execution)); err != nil {
		t.Fatalf("no-op trace copy rejected: %v", err)
	}
	partial := clonePhase36Execution(clean.Execution)
	partial.Events = partial.Events[:len(partial.Events)-1]
	if err := executionpeer.Validate(program, partial); err == nil {
		t.Fatal("peer accepted an incomplete execution prefix as complete")
	}
	if err := executionpeer.Validate(program, clonePhase36Execution(clean.Execution)); err != nil {
		t.Fatalf("restored full trace rejected after partial-prefix control: %v", err)
	}
	mainCalls := phase36Calls(main)
	var loopCall, laterCall, decoyCall core.LinearOperation
	for _, call := range mainCalls {
		switch call.CalleeID {
		case bump.ID:
			if loopCall.ID == "" {
				loopCall = call
			} else {
				laterCall = call
			}
		case decoy.ID:
			decoyCall = call
		}
	}
	if loopCall.ID == "" || laterCall.ID == "" || decoyCall.ID == "" {
		t.Fatalf("expected two bump sites and one decoy site: %+v", mainCalls)
	}
	mutants := []struct {
		name, code string
		mutate     func(*execution.Execution)
	}{
		{"ordinal-gap", "unknown_invocation", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool {
				return event.FunctionID == bump.ID && event.Kind == "function.returned"
			})
			document.Events[index].Invocation = phase36Path(t, main.ID, loopCall.ID, 7)
		}},
		{"ordinal-duplicate", "duplicate_pair", func(document *execution.Execution) {
			indices := phase36FindEvents(*document, func(event execution.Event) bool {
				return event.FunctionID == bump.ID && event.Kind == "function.returned"
			})
			if len(indices) < 2 {
				t.Fatalf("trace has only %d bump returns", len(indices))
			}
			document.Events[indices[1]].Invocation = phase36Path(t, main.ID, loopCall.ID, 0)
		}},
		{"ordinal-reordered", "unknown_invocation", func(document *execution.Execution) {
			indices := phase36FindEvents(*document, func(event execution.Event) bool {
				return event.FunctionID == bump.ID && event.Invocation == phase36Path(t, main.ID, loopCall.ID, 0) && event.Kind == "function.returned"
			})
			if len(indices) == 0 {
				t.Fatal("no first loop callee terminal")
			}
			document.Events[indices[0]].Invocation = phase36Path(t, main.ID, loopCall.ID, 1)
		}},
		{"wrong-parent", "unknown_invocation", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events[index].Invocation = phase36Path(t, main.ID, decoyCall.ID, 0)
		}},
		{"wrong-static-site", "call_attribution", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events[index].ID = laterCall.ID + ":event:called"
		}},
		{"wrong-callee", "callee", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events[index].CalleeFunctionID = decoy.ID
		}},
		{"wrong-source-place", "call_attribution", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events[index].SourcePlace = "forged-source"
		}},
		{"wrong-target-place", "call_attribution", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events[index].TargetPlace = "forged-target"
		}},
		{"wrong-result-type", "call_attribution", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events[index].TypeID = "forged-type"
		}},
		{"missing-call", "unknown_invocation", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool { return event.ID == loopCall.ID+":event:called" })
			document.Events = append(document.Events[:index], document.Events[index+1:]...)
		}},
		{"missing-return", "preorder", func(document *execution.Execution) {
			index := phase36FindEvent(t, *document, func(event execution.Event) bool {
				return event.FunctionID == bump.ID && event.Kind == "function.returned"
			})
			document.Events = append(document.Events[:index], document.Events[index+1:]...)
		}},
		{"repeated-terminal", "closed_activation", func(document *execution.Execution) {
			last := document.Events[len(document.Events)-1]
			last.Occurrence++
			document.Events = append(document.Events, last)
		}},
	}
	for _, mutant := range mutants {
		mutant := mutant
		t.Run(mutant.name, func(t *testing.T) {
			if err := executionpeer.Validate(program, clean.Execution); err != nil {
				t.Fatalf("fresh pre-mutant positive rejected: %v", err)
			}
			noOp := clonePhase36Execution(clean.Execution)
			if err := executionpeer.Validate(program, noOp); err != nil {
				t.Fatalf("fresh no-op negative control rejected: %v", err)
			}
			forged := clonePhase36Execution(clean.Execution)
			mutant.mutate(&forged)
			var peerError *executionpeer.Error
			if err := executionpeer.Validate(program, forged); !errors.As(err, &peerError) || peerError.Class != mutant.code {
				t.Fatalf("mutant Validate() = %v, want executionpeer.%s", err, mutant.code)
			}
			if err := executionpeer.Validate(program, clonePhase36Execution(clean.Execution)); err != nil {
				t.Fatalf("restored positive rejected after %s: %v", mutant.name, err)
			}
		})
	}
	wrongValue := clonePhase36Execution(clean.Execution)
	wrongValue.Outcome.Value = "999"
	if err := executionpeer.Validate(program, wrongValue); err != nil {
		t.Fatalf("causal peer should accept the structurally plausible false-success trace; numeric replay is not its claim: %v", err)
	}
	if clean.Execution.Outcome.Value == wrongValue.Outcome.Value {
		t.Fatal("numeric interpreter claim is not discriminating from the false-success trace")
	}

	overflowProgram, overflowExecution := phase36CausalOverflowRun(t)
	if err := executionpeer.Validate(overflowProgram, overflowExecution); err != nil {
		t.Fatalf("peer rejected actual nested checked-add defect trace: %v", err)
	}
	forgedDefect := clonePhase36Execution(overflowExecution)
	forgedDefect.Events[len(forgedDefect.Events)-1].Output = "forged arithmetic reason"
	if err := executionpeer.Validate(overflowProgram, forgedDefect); err == nil {
		t.Fatal("peer accepted a forged checked-add defect reason")
	}
	postDefect := clonePhase36Execution(overflowExecution)
	postDefect.Events = append(postDefect.Events, postDefect.Events[0])
	if err := executionpeer.Validate(overflowProgram, postDefect); err == nil {
		t.Fatal("peer accepted an event after terminal defect")
	}
	if err := executionpeer.Validate(overflowProgram, overflowExecution); err != nil {
		t.Fatalf("restored actual defect trace rejected: %v", err)
	}
	large := execution.Execution{
		Schema: execution.Schema2, Outcome: execution.Outcome{Kind: execution.OutcomeReturned},
		Events: make([]execution.Event, execution.MaxDocumentBytes/128+1),
	}
	var capacityError *executionpeer.Error
	if err := executionpeer.Validate(program, large); !errors.As(err, &capacityError) || capacityError.Class != "document_capacity" {
		t.Fatalf("over-cap observed event document = %v, want executionpeer.document_capacity", err)
	}
	if err := executionpeer.Validate(program, clonePhase36Execution(clean.Execution)); err != nil {
		t.Fatalf("restored positive after document capacity control rejected: %v", err)
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

func phase36CausalCallRun(t *testing.T) (core.Program, interp.ApplicationResult, core.Function, core.Function, core.Function) {
	t.Helper()
	checked := session.Check([]byte(phase36CausalCallSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("causal-call source diagnostics: %+v", checked.Diagnostics)
	}
	var main, bump, decoy core.Function
	for _, function := range checked.Program.Functions {
		switch function.Name {
		case "main":
			main = function
		case "bump":
			bump = function
		case "decoy":
			decoy = function
		}
	}
	run, err := interp.RunApplication(checked.Program, "main", "0")
	if err != nil {
		t.Fatalf("causal-call RunApplication: %v", err)
	}
	return checked.Program, run, main, bump, decoy
}

func phase36Calls(function core.Function) []core.LinearOperation {
	var calls []core.LinearOperation
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall {
			calls = append(calls, operation)
		}
	}
	return calls
}

func phase36Path(t *testing.T, entry, callID string, ordinal int) string {
	t.Helper()
	path, err := execution.FormatInvocation(entry, []execution.InvocationSegment{{OpCallID: callID, Ordinal: ordinal}})
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func phase36FindEvent(t *testing.T, document execution.Execution, match func(execution.Event) bool) int {
	t.Helper()
	for index, event := range document.Events {
		if match(event) {
			return index
		}
	}
	t.Fatal("target event was not found")
	return -1
}

func phase36FindEvents(document execution.Execution, match func(execution.Event) bool) []int {
	var indices []int
	for index, event := range document.Events {
		if match(event) {
			indices = append(indices, index)
		}
	}
	return indices
}

func clonePhase36Execution(document execution.Execution) execution.Execution {
	clone := document
	clone.Events = append([]execution.Event(nil), document.Events...)
	clone.LiveResources = append([]string(nil), document.LiveResources...)
	return clone
}

func phase36CausalOverflowRun(t *testing.T) (core.Program, execution.Execution) {
	t.Helper()
	checked := session.Check([]byte(phase36CausalOverflowSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("causal overflow source diagnostics: %+v", checked.Diagnostics)
	}
	run, err := interp.Run(checked.Program, "main", "18446744073709551614")
	if err != nil {
		t.Fatalf("causal overflow Run: %v", err)
	}
	return checked.Program, run
}

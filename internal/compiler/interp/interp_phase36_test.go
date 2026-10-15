package interp_test

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
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

const phase36CallSemanticsSource = `module phase36.call_semantics
export { fn main }
fn u64_to_u64(input: U64) -> U64 { let result = input + 1 result }
fn u64_to_bool(input: U64) -> Bool { let result = input < 4 result }
fn bool_to_u64(input: Bool) -> U64 {
  var result = 0
  if input { result = 7 } else { result = 8 }
  result
}
fn bool_to_bool(input: Bool) -> Bool { let result = input result }
fn bump(input: U64) -> U64 { let result = input + 1 result }
fn twice(input: U64) -> U64 {
  var result = input
  var index = 0
  while index < 2 {
    result = bump(result)
    index = index + 1
  }
  result
}
fn main(input: U64) -> U64 {
  let first = twice(input)
  let second = twice(first)
  let third = bump(second)
  let fourth = bump(third)
  fourth
}
fn overflow_step(input: U64) -> U64 { let result = input + 1 result }
fn overflow_main(input: U64) -> U64 {
  var value = input
  var index = 0
  while index < 2 {
    value = overflow_step(value)
    index = index + 1
  }
  let later = bump(value)
  later
}
fn remainder_step(input: U64) -> U64 {
  let divisor = input % 2
  let result = input % divisor
  result
}
fn remainder_main(input: U64) -> U64 {
  var value = input
  var index = 0
  while index < 2 {
    value = remainder_step(value)
    index = index + 1
  }
  let later = bump(value)
  later
}
`

const phase36LongCallSource = `module phase36.long_calls
export { fn main }
fn add_one(input: U64) -> U64 { let result = input + 1 result }
fn main(input: U64) -> U64 {
  write_u64(input)
  write_newline()
  var value = input
  var index = 0
  while index < 5001 {
    value = add_one(value)
    index = index + 1
  }
  value
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

func TestPhase36InterpreterCallAnswersAndDefects(t *testing.T) {
	checked := session.Check([]byte(phase36CallSemanticsSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("call semantics source diagnostics: %+v", checked.Diagnostics)
	}
	for _, test := range []struct {
		name, function, input, want string
	}{
		{"U64-to-U64", "u64_to_u64", "41", "42"},
		{"U64-to-Bool true", "u64_to_bool", "3", "true"},
		{"U64-to-Bool false", "u64_to_bool", "4", "false"},
		{"Bool-to-U64 true", "bool_to_u64", "true", "7"},
		{"Bool-to-U64 false", "bool_to_u64", "false", "8"},
		{"Bool-to-Bool", "bool_to_bool", "false", "false"},
	} {
		t.Run(test.name, func(t *testing.T) {
			result, err := interp.Run(checked.Program, test.function, test.input)
			if err != nil {
				t.Fatalf("Run(%s): %v", test.function, err)
			}
			if result.Outcome.Kind != execution.OutcomeReturned || result.Outcome.Value != test.want {
				t.Fatalf("Run(%s, %q) = %+v, want returned %q", test.function, test.input, result.Outcome, test.want)
			}
		})
	}

	mainResult, err := interp.RunApplication(checked.Program, "main", "1")
	if err != nil {
		t.Fatalf("nested helper loop application: %v", err)
	}
	if mainResult.Execution.Outcome.Value != "7" || string(mainResult.Output) != "7\n" {
		t.Fatalf("nested/helper-local loops = %+v, output %q; want 7", mainResult.Execution.Outcome, mainResult.Output)
	}
	main := phase36Function(t, checked.Program, "main")
	twice := phase36Function(t, checked.Program, "twice")
	bump := phase36Function(t, checked.Program, "bump")
	var mainTwiceSites, mainBumpSites []core.LinearOperation
	for _, operation := range main.Linear.Operations {
		if operation.Kind != core.OpCall {
			continue
		}
		if operation.CalleeID == twice.ID {
			mainTwiceSites = append(mainTwiceSites, operation)
		}
		if operation.CalleeID == bump.ID {
			mainBumpSites = append(mainBumpSites, operation)
		}
	}
	if len(mainTwiceSites) != 2 || len(mainBumpSites) != 2 {
		t.Fatalf("expected two sites per callee: twice=%+v bump=%+v", mainTwiceSites, mainBumpSites)
	}
	root, err := execution.FormatInvocation(main.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	mainPaths := map[string]string{}
	for _, operation := range append(mainTwiceSites, mainBumpSites...) {
		path, err := execution.FormatInvocation(main.ID, []execution.InvocationSegment{{OpCallID: operation.ID, Ordinal: 0}})
		if err != nil {
			t.Fatal(err)
		}
		mainPaths[operation.ID] = path
		found := false
		for _, event := range mainResult.Execution.Events {
			if event.ID == operation.ID+":event:called" {
				found = true
				if event.Invocation != root || event.Occurrence != 0 {
					t.Fatalf("per-site activation event = %+v, want root %q and occurrence zero", event, root)
				}
			}
		}
		if !found {
			t.Fatalf("no call event for static site %q", operation.ID)
		}
	}
	var innerCall core.LinearOperation
	for _, operation := range twice.Linear.Operations {
		if operation.Kind == core.OpCall && operation.CalleeID == bump.ID {
			innerCall = operation
		}
	}
	if innerCall.ID == "" {
		t.Fatal("twice helper has no bump call site")
	}
	for _, parentSite := range mainTwiceSites {
		parentPath := mainPaths[parentSite.ID]
		var actualChildren []string
		var actualOccurrences []uint64
		for _, event := range mainResult.Execution.Events {
			if event.ID == innerCall.ID+":event:called" && event.Invocation == parentPath {
				actualOccurrences = append(actualOccurrences, event.Occurrence)
				child, err := execution.FormatInvocation(main.ID, []execution.InvocationSegment{
					{OpCallID: parentSite.ID, Ordinal: 0},
					{OpCallID: innerCall.ID, Ordinal: len(actualChildren)},
				})
				if err != nil {
					t.Fatal(err)
				}
				actualChildren = append(actualChildren, child)
			}
		}
		if len(actualChildren) != 2 || actualOccurrences[0] != 0 || actualOccurrences[1] != 1 {
			t.Fatalf("helper-local loop under %q children=%v event occurrences=%v; want #0/#1 and occurrences 0/1", parentPath, actualChildren, actualOccurrences)
		}
	}

	for _, test := range []struct {
		function, input, callName, reason string
	}{
		{"overflow_main", "18446744073709551614", "overflow_step", execution.DefectReasonU64AddOverflow},
		{"remainder_main", "1", "remainder_step", execution.DefectReasonU64RemainderByZero},
	} {
		t.Run(test.function, func(t *testing.T) {
			result, err := interp.Run(checked.Program, test.function, test.input)
			if err != nil {
				t.Fatalf("Run(%s): %v", test.function, err)
			}
			if result.Outcome.Kind != execution.OutcomeDefect {
				t.Fatalf("Run(%s) outcome = %+v, want defect", test.function, result.Outcome)
			}
			var defect *execution.Event
			for index := range result.Events {
				if result.Events[index].Kind == "function.defected" {
					defect = &result.Events[index]
				}
			}
			if defect == nil || defect.Output != test.reason {
				t.Fatalf("Run(%s) defect event = %+v, want reason %q", test.function, defect, test.reason)
			}
			callFunction := phase36Function(t, checked.Program, test.function)
			callee := phase36Function(t, checked.Program, test.callName)
			call, _ := phase36CallAndReturn(t, callFunction, callee)
			wantPath, err := execution.FormatInvocation(callFunction.ID, []execution.InvocationSegment{{OpCallID: call.ID, Ordinal: 1}})
			if err != nil {
				t.Fatal(err)
			}
			if defect.Invocation != wantPath {
				t.Fatalf("later defect invocation = %q, want second dynamic helper path %q", defect.Invocation, wantPath)
			}
			for _, event := range result.Events {
				if event.FunctionID == callFunction.ID && event.Kind == "function.called" && event.ID != call.ID+":event:called" {
					t.Fatalf("caller continued into later sibling after defect: %+v", event)
				}
			}
		})
	}
}

func TestPhase36InterpreterEvidenceBoundaries(t *testing.T) {
	checked := session.Check([]byte(phase36RepeatedCallSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("evidence-boundary source diagnostics: %+v", checked.Diagnostics)
	}
	clean, err := interp.RunApplication(checked.Program, "main", "0")
	if err != nil {
		t.Fatalf("unbounded reference RunApplication: %v", err)
	}
	if clean.Execution.Outcome.Value != "9" || string(clean.Output) != "9\n" {
		t.Fatalf("reference outcome/output = %+v / %q, want 9", clean.Execution.Outcome, clean.Output)
	}
	pathBytes, eventBytes := phase36EvidenceBytes(t, clean.Execution.Events)
	if pathBytes == 0 || eventBytes == 0 {
		t.Fatalf("reference evidence byte costs = paths %d, events %d", pathBytes, eventBytes)
	}

	t.Run("path-exact", func(t *testing.T) {
		restore := interp.SetEvidenceLimitsForTest(pathBytes, execution.MaxDocumentBytes, execution.MaxDocumentBytes, 2147483647)
		defer restore()
		result, err := interp.RunApplication(checked.Program, "main", "0")
		if err != nil || !execution.Equal(result.Execution, clean.Execution) || string(result.Output) != "9\n" {
			t.Fatalf("exact path budget result = %+v, output %q, err %v", result.Execution, result.Output, err)
		}
	})
	t.Run("path-plus-one-exhausts", func(t *testing.T) {
		restore := interp.SetEvidenceLimitsForTest(pathBytes-1, execution.MaxDocumentBytes, execution.MaxDocumentBytes, 2147483647)
		defer restore()
		result, err := interp.RunApplication(checked.Program, "main", "0")
		if !errors.Is(err, interp.ErrEvidenceCapacityExhausted) || len(result.Execution.Events) != 0 {
			t.Fatalf("path plus one err/events = %v / %d, want named evidence exhaustion and no partial trace", err, len(result.Execution.Events))
		}
		if result.Execution.Outcome.Value != "9" || string(result.Output) != "9\n" {
			t.Fatalf("path exhaustion lost actual result/output: %+v / %q", result.Execution.Outcome, result.Output)
		}
	})
	t.Run("event-exact", func(t *testing.T) {
		restore := interp.SetEvidenceLimitsForTest(execution.MaxDocumentBytes, eventBytes, execution.MaxDocumentBytes, 2147483647)
		defer restore()
		result, err := interp.RunApplication(checked.Program, "main", "0")
		if err != nil || !execution.Equal(result.Execution, clean.Execution) || string(result.Output) != "9\n" {
			t.Fatalf("exact event budget result = %+v, output %q, err %v", result.Execution, result.Output, err)
		}
	})
	t.Run("event-plus-one-exhausts", func(t *testing.T) {
		restore := interp.SetEvidenceLimitsForTest(execution.MaxDocumentBytes, eventBytes-1, execution.MaxDocumentBytes, 2147483647)
		defer restore()
		result, err := interp.RunApplication(checked.Program, "main", "0")
		if !errors.Is(err, interp.ErrEvidenceCapacityExhausted) || len(result.Execution.Events) != 0 {
			t.Fatalf("event plus one err/events = %v / %d, want named evidence exhaustion and no partial trace", err, len(result.Execution.Events))
		}
		if result.Execution.Outcome.Value != "9" || string(result.Output) != "9\n" {
			t.Fatalf("event exhaustion lost actual result/output: %+v / %q", result.Execution.Outcome, result.Output)
		}
	})
	t.Run("ordinal-before-wrap", func(t *testing.T) {
		restore := interp.SetEvidenceLimitsForTest(execution.MaxDocumentBytes, execution.MaxDocumentBytes, execution.MaxDocumentBytes, 1)
		defer restore()
		result, err := interp.RunApplication(checked.Program, "main", "0")
		if !errors.Is(err, interp.ErrEvidenceCapacityExhausted) || len(result.Execution.Events) != 0 {
			t.Fatalf("near-limit ordinal err/events = %v / %d, want named evidence exhaustion and no partial trace", err, len(result.Execution.Events))
		}
		if result.Execution.Outcome.Value != "9" || string(result.Output) != "9\n" {
			t.Fatalf("ordinal exhaustion stopped numeric work: %+v / %q", result.Execution.Outcome, result.Output)
		}
	})
	t.Run("more-than-4096-work-and-late-defect", func(t *testing.T) {
		checked := session.Check([]byte(phase36LongCallSource))
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("long-call source diagnostics: %+v", checked.Diagnostics)
		}
		const input = "18446744073709546615"
		restore := interp.SetEvidenceLimitsForTest(4096, 4096, 8192, 2)
		defer restore()
		result, err := interp.RunApplication(checked.Program, "main", input)
		if !errors.Is(err, interp.ErrEvidenceCapacityExhausted) || len(result.Execution.Events) != 0 {
			t.Fatalf("post-exhaustion err/events = %v / %d, want named evidence error and no partial trace", err, len(result.Execution.Events))
		}
		if result.Execution.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("execution after capture exhaustion = %+v, want the 5001st checked addition to defect", result.Execution.Outcome)
		}
		if !strings.Contains(err.Error(), execution.DefectReasonU64AddOverflow) {
			t.Fatalf("evidence error lost the bounded terminal reason: %v", err)
		}
		if string(result.Output) != input+"\n" {
			t.Fatalf("capture exhaustion lost already-produced application output: %q", result.Output)
		}
	})
	t.Run("fixture-route-preserves-execution-on-exhaustion", func(t *testing.T) {
		source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway"))
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		start := strings.Index(text, "fn adler32(")
		if start < 0 {
			t.Fatal("fixture source no longer has its expected adler32/main layout")
		}
		relativeEnd := strings.Index(text[start:], "\nfn main(")
		if relativeEnd < 0 {
			t.Fatal("fixture source no longer has its expected adler32/main layout")
		}
		end := start + relativeEnd
		replacement := `fn adler32(bytes: ByteView) -> U64 {
  var value = 0
  var index = 0
  while index < 2 {
    value = bump(value)
    index = index + 1
  }
  value
}
fn bump(input: U64) -> U64 { let result = input + 1 result }
`
		checked := session.Check([]byte(text[:start] + replacement + text[end:]))
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("fixture-route source diagnostics: %+v", checked.Diagnostics)
		}
		restore := interp.SetEvidenceLimitsForTest(execution.MaxDocumentBytes, execution.MaxDocumentBytes, execution.MaxDocumentBytes, 0)
		defer restore()
		result, err := interp.RunWithFixtureBytes(checked.Program, "main", "3", testsupport.Phase29FixtureBytes)
		if !errors.Is(err, interp.ErrEvidenceCapacityExhausted) || len(result.Events) != 0 {
			t.Fatalf("fixture-route exhaustion = execution %+v, err %v; want named error and no partial trace", result, err)
		}
		if result.Outcome.Kind != execution.OutcomeReturned || result.Outcome.Value != "2" {
			t.Fatalf("fixture route lost completed helper result: %+v, want returned 2", result.Outcome)
		}
	})
}

func phase36EvidenceBytes(t *testing.T, events []execution.Event) (int, int) {
	t.Helper()
	paths := make(map[string]bool)
	for _, event := range events {
		if event.Invocation != "" {
			paths[event.Invocation] = true
		}
	}
	pathBytes := 0
	for path := range paths {
		pathBytes += len(path)
	}
	eventBytes := 0
	for index, event := range events {
		encoded, err := json.Marshal(event)
		if err != nil {
			t.Fatal(err)
		}
		eventBytes += len(encoded)
		if index > 0 {
			eventBytes++
		}
	}
	return pathBytes, eventBytes
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

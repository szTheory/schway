package check_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase36ScalarCallLoweringAndTypePairs(t *testing.T) {
	source := []byte(`module phase36.scalar_calls
export { fn main }
fn identity_u64(input: U64) -> U64 { let result = input result }
fn less_than_one(input: U64) -> Bool { let result = input < 1 result }
fn bool_number(input: Bool) -> U64 { let result = 1 result }
fn identity_bool(input: Bool) -> Bool { let result = input result }
fn inner(input: U64) -> U64 { let result = input + 1 result }
fn outer(input: U64) -> U64 { let result = input + 2 result }
fn sibling(input: U64) -> U64 { let result = input + 3 result }
fn index_value(input: U64) -> U64 { let result = input result }
fn dynamic_zero(input: U64) -> U64 { let result = 0 result }
fn call_u64_u64(input: U64) -> U64 { let result = identity_u64(input) result }
fn call_u64_bool(input: U64) -> Bool { let result = less_than_one(input) result }
fn call_bool_u64(input: Bool) -> U64 { let result = bool_number(input) result }
fn call_bool_bool(input: Bool) -> Bool { let result = identity_bool(input) result }
fn compose(input: U64) -> U64 {
  let result = outer(inner(input + 1)) + sibling(input + 2)
  result
}
fn contexts(input: U64) -> U64 {
  var result = inner(input)
  result = sibling(result + 1)
  if less_than_one(result) {
    result = inner(result)
  } else {
    result = sibling(result)
  }
  result
}
fn byte_index(view: ByteView) -> U64 {
  var result = 0
  result = byte_at(view, index_value(result))
  result
}
fn divide_by_dynamic_zero(input: U64) -> U64 {
  let result = input % dynamic_zero(input)
  result
}
fn main(input: U64) -> U64 { let result = call_u64_u64(input) result }
`)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("scalar call source did not parse: %+v", parsed.Diagnostics)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("well-typed scalar helper call source was refused: %+v", checked.Diagnostics)
	}

	functionIDs := make(map[string]string, len(checked.Program.Functions))
	for _, function := range checked.Program.Functions {
		functionIDs[function.Name] = function.ID
	}
	for _, pair := range []struct {
		caller string
		callee string
		result string
	}{
		{"call_u64_u64", "identity_u64", "U64"},
		{"call_u64_bool", "less_than_one", "Bool"},
		{"call_bool_u64", "bool_number", "U64"},
		{"call_bool_bool", "identity_bool", "Bool"},
	} {
		caller := phase36Function(t, checked.Program, pair.caller)
		callee := phase36Function(t, checked.Program, pair.callee)
		calls := phase36CallOperations(caller)
		if len(calls) != 1 {
			t.Fatalf("%s call operations = %+v, want one OpCall", pair.caller, calls)
		}
		call := calls[0]
		if call.SourceID != caller.Parameter.ID || call.CalleeID != callee.ID || call.TargetID == "" || call.TargetID == call.SourceID {
			t.Fatalf("%s call identity/source/target = %+v, parameter=%s callee=%s", pair.caller, call, caller.Parameter.ID, callee.ID)
		}
		places := make(map[string]core.Place, len(caller.Linear.Places))
		for _, place := range caller.Linear.Places {
			places[place.ID] = place
		}
		place, exists := places[call.TargetID]
		if !exists || place.TypeID != call.TypeID {
			t.Fatalf("%s call target place is absent or disagrees with its TypeID: %+v", pair.caller, call)
		}
		constructor := ""
		for _, fact := range caller.Linear.Types {
			if fact.ID == call.TypeID {
				constructor = fact.Shape.Constructor
			}
		}
		if constructor != pair.result || callee.ReturnType != pair.result {
			t.Fatalf("%s call result constructor=%q callee return=%q, want %q", pair.caller, constructor, callee.ReturnType, pair.result)
		}
	}
	_ = functionIDs

	compose := phase36Function(t, checked.Program, "compose")
	composeCalls := phase36CallOperations(compose)
	if len(composeCalls) != 3 {
		t.Fatalf("nested/sibling call count = %d, want 3: %+v", len(composeCalls), composeCalls)
	}
	wantCallers := []string{functionIDs["inner"], functionIDs["outer"], functionIDs["sibling"]}
	for index, call := range composeCalls {
		if call.CalleeID != wantCallers[index] {
			t.Fatalf("call order[%d] = %q, want %q", index, call.CalleeID, wantCallers[index])
		}
		for earlier := 0; earlier < index; earlier++ {
			if call.TargetID == composeCalls[earlier].TargetID {
				t.Fatalf("call results reused static target %q", call.TargetID)
			}
		}
	}
	if composeCalls[1].SourceID != composeCalls[0].TargetID {
		t.Fatalf("outer call did not consume the completed inner result: inner=%+v outer=%+v", composeCalls[0], composeCalls[1])
	}
	composePositions := make(map[string]int, len(compose.Linear.Operations))
	for index, operation := range compose.Linear.Operations {
		composePositions[operation.ID] = index
	}
	for _, call := range composeCalls {
		if sourceIndex, targetIndex := composePositions[call.SourceID], composePositions[call.ID]; sourceIndex >= targetIndex {
			t.Fatalf("call %s ran before its argument producer %s", call.ID, call.SourceID)
		}
	}

	contexts := phase36Function(t, checked.Program, "contexts")
	contextCalls := phase36CallOperations(contexts)
	if len(contextCalls) != 5 {
		t.Fatalf("initializer/assignment/condition/branch calls = %+v, want 5 calls", contextCalls)
	}
	branches := 0
	for _, operation := range contexts.Linear.Operations {
		if operation.Kind == core.OpBranch {
			branches++
			if operation.SourceID != contextCalls[2].TargetID {
				t.Fatalf("condition branch reads %q, want Bool call result %q", operation.SourceID, contextCalls[2].TargetID)
			}
		}
	}
	if branches != 1 {
		t.Fatalf("condition branch count = %d, want 1", branches)
	}

	byteIndex := phase36Function(t, checked.Program, "byte_index")
	indexCalls := phase36CallOperations(byteIndex)
	var byteRead core.LinearOperation
	for _, operation := range byteIndex.Linear.Operations {
		if operation.Kind == core.OpBytesAt {
			byteRead = operation
		}
	}
	if len(indexCalls) != 1 || byteRead.Kind != core.OpBytesAt || byteRead.RightID != indexCalls[0].TargetID {
		t.Fatalf("call in byte_at index was not materialized before the byte read: calls=%+v byte_at=%+v", indexCalls, byteRead)
	}

	divide := phase36Function(t, checked.Program, "divide_by_dynamic_zero")
	zeroCalls := phase36CallOperations(divide)
	var remainder core.LinearOperation
	for _, operation := range divide.Linear.Operations {
		if operation.Kind == core.OpRemainderU64 {
			remainder = operation
		}
	}
	if len(zeroCalls) != 1 || remainder.Kind != core.OpRemainderU64 || remainder.RightID != zeroCalls[0].TargetID {
		t.Fatalf("dynamic helper result did not remain an initialized unknown divisor: call=%+v remainder=%+v", zeroCalls, remainder)
	}

	byteSource := []byte(`module phase36.byte_call_use
export { fn main }
fn identity_u64(input: U64) -> U64 { let result = input result }
fn main(input: Byte) -> U64 {
  var result = 0
  result = identity_u64(input)
  result
}
`)
	byteUse := session.Check(byteSource)
	if len(byteUse.Diagnostics) != 1 || byteUse.Diagnostics[0].Code != "type.scalar_byte_parameter_used" {
		t.Fatalf("unused Byte bridge allowed its formal inside a call argument: %+v", byteUse.Diagnostics)
	}
}

func phase36Function(t *testing.T, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("checked function %q not found", name)
	return core.Function{}
}

func phase36CallOperations(function core.Function) []core.LinearOperation {
	var calls []core.LinearOperation
	if function.Linear == nil {
		return calls
	}
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpCall {
			calls = append(calls, operation)
		}
	}
	return calls
}

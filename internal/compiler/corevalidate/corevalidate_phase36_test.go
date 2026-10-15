package corevalidate_test

import (
	"encoding/json"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const phase36RawCallSource = `module phase36.raw_calls
export { fn main }
fn u64_u64(value: U64) -> U64 { let result = value result }
fn u64_bool(value: U64) -> Bool { let result = value < 9 result }
fn bool_u64(value: Bool) -> U64 { let result = 13 result }
fn bool_bool(value: Bool) -> Bool { let result = value result }
fn main(input: U64) -> U64 {
  let seed = 7
  let copied = input
  let below = u64_bool(input)
  let same = bool_bool(below)
  let from_bool = bool_u64(same)
  let from_u64 = u64_u64(copied)
  var current = seed
  if same {
    current = bool_u64(same)
  } else {
    current = bool_u64(below)
  }
  while current < 4 {
    current = u64_u64(current)
  }
  let paired = u64_bool(from_u64)
  var final = current
  if paired {
    final = from_bool
  } else {
    final = current
  }
  final
}
`

func TestPhase36RawCallShapeAndFreshness(t *testing.T) {
	program := phase36RawCallProgram(t)
	main := phase36RawCallFunction(t, &program, "main")
	callees := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		callees[function.ID] = function
	}
	pairs := map[string]bool{}
	callCount := 0
	for _, operation := range main.Linear.Operations {
		if operation.Kind != core.OpCall {
			continue
		}
		callCount++
		callee, ok := callees[operation.CalleeID]
		if !ok {
			t.Fatalf("test fixture has unresolved call: %+v", operation)
		}
		pairs[callee.Parameter.Type+"->"+callee.ReturnType] = true
		if operation.SourceID == "" || operation.TargetID == "" || operation.SourceID == operation.TargetID {
			t.Fatalf("call lacks distinct initialized argument/result places: %+v", operation)
		}
		if operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "" || operation.Foreign != nil {
			t.Fatalf("infallible scalar call carries unrelated authority: %+v", operation)
		}
	}
	for _, pair := range []string{"U64->U64", "U64->Bool", "Bool->U64", "Bool->Bool"} {
		if !pairs[pair] {
			t.Errorf("raw fixture does not contain signature pair %s (calls=%d)", pair, callCount)
		}
	}
	if callCount < 6 {
		t.Fatalf("branch and loop fixture lost call producers: got %d calls", callCount)
	}
	if got := corevalidate.Validate(program); !got.Valid {
		t.Fatalf("corevalidate refused well-typed raw scalar calls and repeated static branch/loop producers: %+v", got.Problems)
	}
}

func TestPhase36RawCallRefusalRemovalControls(t *testing.T) {
	program := phase36RawCallProgram(t)
	if got := corevalidate.Validate(program); !got.Valid {
		t.Fatalf("clean raw call baseline refused: %+v", got.Problems)
	}
	noop := phase36CloneRawCallProgram(t, program)
	if got := corevalidate.Validate(noop); !got.Valid {
		t.Fatalf("no-op copy changed raw call verdict: %+v", got.Problems)
	}

	// Keep the call's result contract intact while replacing its U64 argument
	// with a Bool place. This is accepted only when corevalidate's own argument
	// contract seam is removed; source checking and the other peer are absent.
	argumentMutant := phase36CloneRawCallProgram(t, program)
	argumentMain := phase36RawCallFunction(t, &argumentMutant, "main")
	argumentCall := phase36FindRawCall(t, argumentMain, "u64_bool")
	argumentCall.SourceID = phase36RawCallPlace(t, argumentMain, "below")
	if got := corevalidate.Validate(argumentMutant); got.Valid {
		t.Fatal("corevalidate accepted a Bool argument for a U64 parameter")
	}
	restoreArgument := corevalidate.SetDisableCallArgumentTypePeerForTest(true)
	opened := corevalidate.Validate(argumentMutant)
	restoreArgument()
	if !opened.Valid {
		t.Fatalf("argument-contract removal did not reach malformed raw-call acceptance: %+v", opened.Problems)
	}
	if restored := corevalidate.Validate(argumentMutant); restored.Valid {
		t.Fatal("argument contract stayed disabled after restore")
	}

	// Keep the argument valid while pointing at a U64-returning declaration and
	// retaining the original Bool target. Only the independent return contract
	// distinguishes this malformed raw call.
	returnMutant := phase36CloneRawCallProgram(t, program)
	returnMain := phase36RawCallFunction(t, &returnMutant, "main")
	returnCall := phase36FindRawCall(t, returnMain, "u64_bool")
	returnCall.CalleeID = phase36RawCallFunction(t, &returnMutant, "u64_u64").ID
	if got := corevalidate.Validate(returnMutant); got.Valid {
		t.Fatal("corevalidate accepted a U64 callee return into a Bool result place")
	}
	restoreReturn := corevalidate.SetDisableCallReturnTypePeerForTest(true)
	opened = corevalidate.Validate(returnMutant)
	restoreReturn()
	if !opened.Valid {
		t.Fatalf("return-contract removal did not reach malformed raw-call acceptance: %+v", opened.Problems)
	}
	if restored := corevalidate.Validate(returnMutant); restored.Valid {
		t.Fatal("return contract stayed disabled after restore")
	}

	// A static producer cannot redefine an earlier call result. One call site
	// in the loop is still valid when executed repeatedly at runtime; this
	// mutant changes its static target to an earlier producer.
	freshnessMutant := phase36CloneRawCallProgram(t, program)
	freshMain := phase36RawCallFunction(t, &freshnessMutant, "main")
	calls := phase36RawCalls(freshMain)
	if len(calls) < 2 {
		t.Fatal("fixture needs two call producers for freshness control")
	}
	last := calls[len(calls)-1]
	last.TargetID = calls[0].TargetID
	if got := corevalidate.Validate(freshnessMutant); got.Valid {
		t.Fatal("corevalidate accepted reuse of a prior static OpCall target")
	}
}

func phase36RawCallProgram(t *testing.T) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(phase36RawCallSource))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 36 raw call fixture did not parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 5 {
		t.Fatalf("Phase 36 raw call fixture did not lower: diagnostics=%+v functions=%d", checked.Diagnostics, len(checked.Program.Functions))
	}
	return checked.Program
}

func phase36CloneRawCallProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	data, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.Program
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func phase36RawCallFunction(t *testing.T, program *core.Program, name string) *core.Function {
	t.Helper()
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	t.Fatalf("missing raw call function %q", name)
	return nil
}

func phase36FindRawCall(t *testing.T, function *core.Function, calleeName string) *core.LinearOperation {
	t.Helper()
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == core.OpCall && operation.CalleeID == function.Module+"."+calleeName || operation.Kind == core.OpCall && len(operation.CalleeID) >= len(calleeName) && operation.CalleeID[len(operation.CalleeID)-len(calleeName):] == calleeName {
			return operation
		}
	}
	t.Fatalf("missing call to %q in %s", calleeName, function.ID)
	return nil
}

func phase36RawCallPlace(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, place := range function.Linear.Places {
		if place.Name == name {
			return place.ID
		}
	}
	t.Fatalf("missing raw call place %q", name)
	return ""
}

func phase36RawCalls(function *core.Function) []*core.LinearOperation {
	var result []*core.LinearOperation
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == core.OpCall {
			result = append(result, &function.Linear.Operations[index])
		}
	}
	return result
}

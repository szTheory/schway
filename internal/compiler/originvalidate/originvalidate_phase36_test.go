package originvalidate_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const phase36OriginCallSource = `module phase36.origin_calls
export { fn main }
fn u64_u64(value: U64) -> U64 { let result = value result }
fn u64_bool(value: U64) -> Bool { let result = value < 9 result }
fn bool_u64(value: Bool) -> U64 { let result = 13 result }
fn bool_bool(value: Bool) -> Bool { let result = value result }
fn zero(value: U64) -> U64 { let result = 0 result }
fn view_u64(value: ByteView) -> U64 { let result = 3 result }
fn main(input: U64) -> U64 {
  let copied = input
  let first = u64_u64(copied)
  let flag = u64_bool(input)
  let second = bool_u64(flag)
  let same = bool_bool(flag)
  let dynamic = zero(input)
  let one = 1
  let remainder = one % dynamic
  var current = first
  if same {
    current = bool_u64(same)
  } else {
    current = bool_u64(flag)
  }
  var count = 0
  while count < 2 {
    current = u64_u64(current)
    count = count + 1
  }
  let after_loop = u64_bool(current)
  if after_loop {
    current = bool_u64(after_loop)
  } else {
    current = second
  }
  let result = remainder + current
  result
}
`

func TestPhase36OriginCallFreshnessAndTransfer(t *testing.T) {
	program := phase36OriginCallProgram(t)
	main := phase36OriginCallFunction(t, &program, "main")
	contracts := originvalidate.BuildCalleeOriginFacts(program)
	if problems := originvalidate.PublishProblemsFor(*main, contracts); len(problems) != 0 {
		t.Fatalf("origin peer refused valid raw scalar calls, branch/loop assignments, or unknown zero-helper result: %+v", problems)
	}

	pairs := map[string]bool{}
	callees := make(map[string]core.Function, len(program.Functions))
	for _, function := range program.Functions {
		callees[function.ID] = function
	}
	for _, operation := range main.Linear.Operations {
		if operation.Kind != core.OpCall {
			continue
		}
		callee, ok := callees[operation.CalleeID]
		if !ok {
			t.Fatalf("raw origin call did not resolve from program contracts: %+v", operation)
		}
		pairs[callee.Parameter.Type+"->"+callee.ReturnType] = true
		if operation.TargetID == "" || operation.SourceID == operation.TargetID || operation.Foreign != nil || operation.LoanID != "" || operation.ErrTargetID != "" || operation.OkEdgeID != "" || operation.ErrEdgeID != "" {
			t.Fatalf("raw scalar call has missing target or unrelated authority: %+v", operation)
		}
	}
	for _, pair := range []string{"U64->U64", "U64->Bool", "Bool->U64", "Bool->Bool"} {
		if !pairs[pair] {
			t.Errorf("raw origin fixture lacks scalar pair %s", pair)
		}
	}
}

func TestPhase36OriginCallRefusalRemovalControls(t *testing.T) {
	base := phase36OriginCallProgram(t)
	main := phase36OriginCallFunction(t, &base, "main")
	contracts := originvalidate.BuildCalleeOriginFacts(base)
	if got := originvalidate.PublishProblemsFor(*main, contracts); len(got) != 0 {
		t.Fatalf("clean origin call baseline refused: %+v", got)
	}
	copy := phase36CloneOriginCallProgram(t, base)
	copyMain := phase36OriginCallFunction(t, &copy, "main")
	if got := originvalidate.PublishProblemsFor(*copyMain, originvalidate.BuildCalleeOriginFacts(copy)); len(got) != 0 {
		t.Fatalf("no-op origin call copy changed the verdict: %+v", got)
	}

	for _, row := range []struct {
		name string
		edit func(*testing.T, *core.Program, *core.Function, *core.LinearOperation)
	}{
		{name: "wrong argument type", edit: func(t *testing.T, _ *core.Program, caller *core.Function, call *core.LinearOperation) {
			call.SourceID = phase36OriginCallPlace(t, caller, "flag")
		}},
		{name: "wrong result contract", edit: func(t *testing.T, program *core.Program, _ *core.Function, call *core.LinearOperation) {
			call.CalleeID = phase36OriginCallFunction(t, program, "bool_u64").ID
		}},
		{name: "missing callee", edit: func(_ *testing.T, _ *core.Program, _ *core.Function, call *core.LinearOperation) {
			call.CalleeID = ""
		}},
		{name: "missing argument", edit: func(_ *testing.T, _ *core.Program, _ *core.Function, call *core.LinearOperation) {
			call.SourceID = ""
		}},
		{name: "missing result", edit: func(_ *testing.T, _ *core.Program, _ *core.Function, call *core.LinearOperation) {
			call.TargetID = ""
		}},
		{name: "extra authority field", edit: func(t *testing.T, _ *core.Program, caller *core.Function, call *core.LinearOperation) {
			call.RightID = phase36OriginCallPlace(t, caller, "input")
		}},
		{name: "uninitialized argument", edit: func(_ *testing.T, _ *core.Program, caller *core.Function, call *core.LinearOperation) {
			typeID := caller.Linear.Places[0].TypeID
			id := fmt.Sprintf("%s:place:%d", caller.ID, len(caller.Linear.Places))
			caller.Linear.Places = append(caller.Linear.Places, core.Place{ID: id, Name: "uninitialized", TypeID: typeID})
			call.SourceID = id
		}},
		{name: "prior call target reuse", edit: func(t *testing.T, _ *core.Program, caller *core.Function, call *core.LinearOperation) {
			call.TargetID = phase36OriginCallCall(t, caller, "bool_bool").TargetID
		}},
		{name: "prior constant target reuse", edit: func(t *testing.T, _ *core.Program, caller *core.Function, _ *core.LinearOperation) {
			phase36OriginCallLast(t, caller, "u64_u64").TargetID = phase36OriginCallPlace(t, caller, "one")
		}},
		{name: "prior copy target reuse", edit: func(t *testing.T, _ *core.Program, caller *core.Function, _ *core.LinearOperation) {
			phase36OriginCallLast(t, caller, "u64_u64").TargetID = phase36OriginCallPlace(t, caller, "copied")
		}},
		{name: "prior mutable-store target reuse", edit: func(t *testing.T, _ *core.Program, caller *core.Function, _ *core.LinearOperation) {
			phase36OriginCallLast(t, caller, "u64_u64").TargetID = phase36OriginCallPlace(t, caller, "current")
		}},
		{name: "ByteView scalar signature", edit: func(t *testing.T, program *core.Program, caller *core.Function, call *core.LinearOperation) {
			call = phase36OriginCallCall(t, caller, "u64_u64")
			viewCallee := phase36OriginCallFunction(t, program, "view_u64")
			call.CalleeID = viewCallee.ID
			call.SourceID = phase36AppendOriginPlace(t, caller, viewCallee, "ByteView", "view")
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			mutant := phase36CloneOriginCallProgram(t, base)
			caller := phase36OriginCallFunction(t, &mutant, "main")
			call := phase36OriginCallCall(t, caller, "u64_bool")
			row.edit(t, &mutant, caller, call)
			mutantContracts := originvalidate.BuildCalleeOriginFacts(mutant)
			if got := originvalidate.PublishProblemsFor(*caller, mutantContracts); len(got) == 0 {
				t.Fatal("origin peer accepted a forged raw scalar call")
			}
			restore := originvalidate.SetDisableScalarAnalysisLawForTest(true)
			defer restore()
			opened := originvalidate.PublishProblemsFor(*caller, mutantContracts)
			restore()
			if len(opened) != 0 {
				t.Fatalf("origin-owned refusal-removal control did not reach its malformed-IR acceptance assertion: %+v", opened)
			}
			if restored := originvalidate.PublishProblemsFor(*caller, mutantContracts); len(restored) == 0 {
				t.Fatal("origin call contract stayed disabled after restore")
			}
			if clean := originvalidate.PublishProblemsFor(*phase36OriginCallFunction(t, &copy, "main"), originvalidate.BuildCalleeOriginFacts(copy)); len(clean) != 0 {
				t.Fatalf("positive control no longer passes after restoring refusal law: %+v", clean)
			}
		})
	}
}

func phase36OriginCallProgram(t *testing.T) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(phase36OriginCallSource))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 36 origin fixture parse failed: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) == 0 {
		t.Fatalf("Phase 36 origin fixture lowering failed: diagnostics=%+v", checked.Diagnostics)
	}
	return checked.Program
}

func phase36CloneOriginCallProgram(t *testing.T, program core.Program) core.Program {
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

func phase36OriginCallFunction(t *testing.T, program *core.Program, name string) *core.Function {
	t.Helper()
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	t.Fatalf("missing origin call function %q", name)
	return nil
}

func phase36OriginCallCall(t *testing.T, function *core.Function, callee string) *core.LinearOperation {
	return phase36OriginCallAt(t, function, callee, 0)
}

func phase36OriginCallLast(t *testing.T, function *core.Function, callee string) *core.LinearOperation {
	t.Helper()
	var matches []*core.LinearOperation
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == core.OpCall && len(operation.CalleeID) >= len(callee) && operation.CalleeID[len(operation.CalleeID)-len(callee):] == callee {
			matches = append(matches, operation)
		}
	}
	if len(matches) == 0 {
		t.Fatalf("missing origin call to %q", callee)
	}
	return matches[len(matches)-1]
}

func phase36OriginCallAt(t *testing.T, function *core.Function, callee string, occurrence int) *core.LinearOperation {
	t.Helper()
	var matches []*core.LinearOperation
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		// The module-qualified ID suffix is sufficient here and keeps this
		// helper independent from any checker signature table.
		if operation.Kind == core.OpCall && len(operation.CalleeID) >= len(callee) && operation.CalleeID[len(operation.CalleeID)-len(callee):] == callee {
			matches = append(matches, operation)
		}
	}
	if occurrence >= 0 && occurrence < len(matches) {
		return matches[occurrence]
	}
	t.Fatalf("missing origin call to %q occurrence %d", callee, occurrence)
	return nil
}

func phase36OriginCallPlace(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, place := range function.Linear.Places {
		if place.Name == name {
			return place.ID
		}
	}
	t.Fatalf("missing origin call place %q", name)
	return ""
}

func phase36AppendOriginPlace(t *testing.T, function, typeSource *core.Function, constructor, name string) string {
	t.Helper()
	var typeFact core.TypeFact
	found := false
	for _, candidate := range typeSource.Linear.Types {
		if candidate.Shape.Constructor == constructor {
			typeFact, found = candidate, true
			break
		}
	}
	if !found {
		t.Fatalf("missing %s type fact in %s", constructor, typeSource.ID)
	}
	typeFact.ID = fmt.Sprintf("%s:type:%d", function.ID, len(function.Linear.Types))
	function.Linear.Types = append(function.Linear.Types, typeFact)
	placeID := fmt.Sprintf("%s:place:%d", function.ID, len(function.Linear.Places))
	function.Linear.Places = append(function.Linear.Places, core.Place{ID: placeID, Name: name, TypeID: typeFact.ID})
	return placeID
}

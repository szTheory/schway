package pathoracle

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase36PathCallSource = `module phase36.path_calls
export { fn main }
fn u64_u64(value: U64) -> U64 { let result = value result }
fn u64_bool(value: U64) -> Bool { let result = value < 9 result }
fn bool_u64(value: Bool) -> U64 { let result = 13 result }
fn bool_bool(value: Bool) -> Bool { let result = value result }
fn view_u64(value: ByteView) -> U64 { let result = 3 result }
fn main(input: U64) -> U64 {
  let first = u64_u64(input)
  let flag = u64_bool(input)
  let from_bool = bool_u64(flag)
  let same = bool_bool(flag)
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
    current = from_bool
  }
  current
}
`

func TestPhase36ScalarCallAuthorityBoundaries(t *testing.T) {
	program := phase36PathCallProgram(t)
	main := phase36PathCallFunction(t, &program, "main")
	if err := ValidateLocalOwnerPaths(program); err != nil {
		t.Fatalf("path oracle refused valid scalar helper pairs and branch/loop stores: %v", err)
	}
	for _, pair := range []struct{ name, callee string }{{"U64-to-U64", "u64_u64"}, {"U64-to-Bool", "u64_bool"}, {"Bool-to-U64", "bool_u64"}, {"Bool-to-Bool", "bool_bool"}} {
		if phase36PathCallLast(main, pair.callee) == nil {
			t.Errorf("checked path fixture is missing %s call", pair.name)
		}
	}

	caller := phase36PathLoanCaller(t, phase36PathCallFunction(t, &program, "u64_u64").ID)
	program.Functions = append(program.Functions, caller)
	if err := ValidateLocalOwnerPaths(program); err != nil {
		t.Fatalf("source loan through a by-value scalar call was refused: %v", err)
	}
	endpoints, _, err := RecomputeEndpoints(caller, BuildCalleeLookup(program))
	if err != nil {
		t.Fatalf("path oracle could not replay a loaned scalar argument: %v", err)
	}
	if got := phase36PathLoanEndpointAfter(t, endpoints, "loan:source"); got != "op:call" {
		t.Fatalf("call did not remain the last source-loan use: endpoint follows %q", got)
	}

	noOp := phase36ClonePathProgram(t, program)
	if err := ValidateLocalOwnerPaths(noOp); err != nil {
		t.Fatalf("no-op raw program copy changed the path-oracle verdict: %v", err)
	}

	for _, row := range []struct {
		name   string
		mutate func(*testing.T, *core.Program, *core.Function)
	}{
		{name: "wrong argument type", mutate: func(t *testing.T, _ *core.Program, function *core.Function) {
			phase36PathCallAt(function, "u64_bool", 0).SourceID = phase36PathPlace(t, function, "same")
		}},
		{name: "wrong result signature", mutate: func(t *testing.T, program *core.Program, function *core.Function) {
			phase36PathCallAt(function, "u64_bool", 0).CalleeID = phase36PathCallFunction(t, program, "bool_u64").ID
		}},
		{name: "loan field on value call", mutate: func(_ *testing.T, _ *core.Program, function *core.Function) {
			phase36PathCallAt(function, "u64_bool", 0).LoanID = "forged-loan"
		}},
		{name: "prior static call target reuse", mutate: func(t *testing.T, _ *core.Program, function *core.Function) {
			phase36PathCallLast(function, "u64_bool").TargetID = phase36PathCallAt(function, "bool_bool", 0).TargetID
		}},
		{name: "resource result type", mutate: func(t *testing.T, _ *core.Program, function *core.Function) {
			call := phase36PathCallAt(function, "u64_u64", 0)
			typeID, _ := phase36AppendPathTypePlace(function, "FileByteOwner", "resource_target")
			call.TargetID = phase36PathPlace(t, function, "resource_target")
			call.TypeID = typeID
		}},
		{name: "ByteView scalar call", mutate: func(t *testing.T, program *core.Program, function *core.Function) {
			call := phase36PathCallAt(function, "u64_u64", 0)
			callee := phase36PathCallFunction(t, program, "view_u64")
			call.CalleeID = callee.ID
			_, sourceID := phase36AppendPathTypePlace(function, "ByteView", "forged_view")
			call.SourceID = sourceID
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			mutant := phase36ClonePathProgram(t, phase36PathCallProgram(t))
			function := phase36PathCallFunction(t, &mutant, "main")
			row.mutate(t, &mutant, function)
			if err := ValidateLocalOwnerPaths(mutant); err == nil {
				t.Fatal("path oracle accepted a malformed raw scalar call")
			}
		})
	}

	// The pre-existing fixture-owner boundary still owns its sole mixed
	// ByteView -> U64 call and rejects mutations through that closed matcher.
	owner := phase36FixtureOwnerProgram(t)
	if err := ValidateLocalOwnerPaths(owner); err != nil {
		t.Fatalf("legacy fixture-owner call no longer passes its owner boundary: %v", err)
	}
	if err := ValidateFixtureOwnerPaths(owner); err != nil {
		t.Fatalf("legacy fixture-owner seam rejected its checked program: %v", err)
	}

	forged := phase36ClonePathProgram(t, program)
	borrower := phase36PathCallFunction(t, &forged, "u64_u64")
	phase36ForgeScalarCalleeLoan(t, borrower)
	forgedCaller := phase36PathLoanCaller(t, borrower.ID)
	for index := range forged.Functions {
		if forged.Functions[index].Name == forgedCaller.Name {
			forged.Functions[index] = forgedCaller
		}
	}
	if err := ValidateLocalOwnerPaths(forged); err == nil {
		t.Fatal("path oracle accepted a scalar helper body with forged loan-return authority")
	}
	restore := func() func() {
		previous := disableScalarCallContextForTest
		disableScalarCallContextForTest = true
		return func() { disableScalarCallContextForTest = previous }
	}()
	if err := ValidateLocalOwnerPaths(forged); err != nil {
		restore()
		t.Fatalf("scalar-context refusal-removal seam did not reach the malformed callee: %v", err)
	}
	forgedEndpoints, _, err := RecomputeEndpoints(forgedCaller, BuildCalleeLookup(forged))
	if err != nil {
		restore()
		t.Fatalf("bypassed oracle could not reproduce forged result ancestry: %v", err)
	}
	if got := phase36PathLoanEndpointAfter(t, forgedEndpoints, "loan:source"); got != "op:return" {
		restore()
		t.Fatalf("forged scalar result ancestry did not carry the source loan to return, got %q", got)
	}
	restore()
	if err := ValidateLocalOwnerPaths(forged); err == nil {
		t.Fatal("scalar-context law remained disabled after restore")
	}
	if err := ValidateLocalOwnerPaths(program); err != nil {
		t.Fatalf("positive scalar-call baseline failed after restoring the refusal law: %v", err)
	}

	cyclic := phase36PathLoanCaller(t, phase36PathCallFunction(t, &program, "u64_u64").ID)
	cyclic.Linear.Edges = []core.Edge{{ID: "edge:back", FromBlockID: "block:entry", ToBlockID: "block:entry"}}
	cyclic.Linear.Blocks[0].Successors = []string{"block:entry"}
	if _, _, err := RecomputeEndpoints(cyclic, BuildCalleeLookup(program)); err == nil {
		t.Fatal("path oracle allowed a scalar call to grant back-edge loan authority")
	} else if _, ok := err.(*backEdgeError); !ok {
		t.Fatalf("cyclic raw path failed for an unrelated reason: %v", err)
	}
}

func phase36PathCallProgram(t *testing.T) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(phase36PathCallSource))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 36 path fixture parse failed: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) == 0 {
		t.Fatalf("Phase 36 path fixture lowering failed: diagnostics=%+v", checked.Diagnostics)
	}
	return checked.Program
}

func phase36ClonePathProgram(t *testing.T, program core.Program) core.Program {
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

func phase36PathCallFunction(t *testing.T, program *core.Program, name string) *core.Function {
	t.Helper()
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	t.Fatalf("missing path call function %q", name)
	return nil
}

func phase36PathCallAt(function *core.Function, callee string, occurrence int) *core.LinearOperation {
	var calls []*core.LinearOperation
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == core.OpCall && len(operation.CalleeID) >= len(callee) && operation.CalleeID[len(operation.CalleeID)-len(callee):] == callee {
			calls = append(calls, operation)
		}
	}
	if occurrence < 0 || occurrence >= len(calls) {
		return nil
	}
	return calls[occurrence]
}

func phase36PathCallLast(function *core.Function, callee string) *core.LinearOperation {
	var calls []*core.LinearOperation
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == core.OpCall && len(operation.CalleeID) >= len(callee) && operation.CalleeID[len(operation.CalleeID)-len(callee):] == callee {
			calls = append(calls, operation)
		}
	}
	if len(calls) == 0 {
		return nil
	}
	return calls[len(calls)-1]
}

func phase36PathPlace(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, place := range function.Linear.Places {
		if place.Name == name {
			return place.ID
		}
	}
	t.Fatalf("missing path call place %q", name)
	return ""
}

func phase36AppendPathTypePlace(function *core.Function, constructor, name string) (string, string) {
	typeID := fmt.Sprintf("%s:type:%d", function.ID, len(function.Linear.Types))
	function.Linear.Types = append(function.Linear.Types, core.TypeFact{ID: typeID, Shape: core.TypeRef{Constructor: constructor}})
	placeID := fmt.Sprintf("%s:place:%d", function.ID, len(function.Linear.Places))
	function.Linear.Places = append(function.Linear.Places, core.Place{ID: placeID, Name: name, TypeID: typeID})
	return typeID, placeID
}

func phase36PathLoanCaller(t *testing.T, calleeID string) core.Function {
	t.Helper()
	const (
		typeU64 = "type:u64"
		param   = "place:argument"
		borrow  = "place:borrowed"
		result  = "place:result"
	)
	operations := []core.LinearOperation{
		{ID: "op:borrow", Kind: core.OpBorrowShared, SourceID: param, TargetID: borrow, TypeID: typeU64, LoanID: "loan:source"},
		{ID: "op:call", Kind: core.OpCall, SourceID: borrow, TargetID: result, TypeID: typeU64, CalleeID: calleeID},
		{ID: "op:return", Kind: core.OpReturn, SourceID: result, TypeID: typeU64},
	}
	return core.Function{
		ID: "function:phase36_loan_caller", Name: "loan_caller", Parameter: core.Parameter{ID: param, Name: "argument", Type: "U64"},
		ReturnType: "U64", EntryPointID: "point:entry",
		Linear: &core.LinearBody{
			ID: "linear:loan_caller", Types: []core.TypeFact{{ID: typeU64, Shape: core.TypeRef{Constructor: "U64"}}},
			Places:     []core.Place{{ID: param, Name: "argument", TypeID: typeU64}, {ID: borrow, Name: "borrowed", TypeID: typeU64}, {ID: result, Name: "result", TypeID: typeU64}},
			Operations: operations,
			Blocks:     []core.Block{{ID: "block:entry", PointID: "point:entry", OperationIDs: []string{"op:borrow", "op:call", "op:return"}}},
		},
	}
}

func phase36PathLoanEndpointAfter(t *testing.T, endpoints []core.LoanEndpoint, loanID string) string {
	t.Helper()
	for _, endpoint := range endpoints {
		if endpoint.LoanID == loanID && endpoint.Kind == "point" {
			return endpoint.AfterOperationID
		}
	}
	t.Fatalf("no path-oracle endpoint found for %s: %+v", loanID, endpoints)
	return ""
}

func phase36ForgeScalarCalleeLoan(t *testing.T, function *core.Function) {
	t.Helper()
	var result string
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == core.OpCopy {
			operation.Kind = core.OpBorrowShared
			operation.LoanID = "loan:forged-helper"
			result = operation.TargetID
			break
		}
	}
	if result == "" {
		t.Fatal("scalar helper fixture has no copy producer to forge")
	}
	function.PublicOrigin = nil
	ids := make([]string, len(function.Linear.Operations))
	for index, operation := range function.Linear.Operations {
		ids[index] = operation.ID
	}
	function.Linear.Blocks = []core.Block{{ID: function.ID + ":block:entry", PointID: function.EntryPointID, OperationIDs: ids}}
}

func phase36FixtureOwnerProgram(t *testing.T) core.Program {
	t.Helper()
	path := testsupport.ProjectPath("examples", "phase29", "adler32_fixture.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture owner source parse failed: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture owner source check failed: %+v", checked.Diagnostics)
	}
	return checked.Program
}

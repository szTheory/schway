package originvalidate_test

import (
	"encoding/json"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// TestPhase35OriginPeerBaseline pins the existing OpConst publication path
// before the independent Phase 35 immediate and transfer laws are added.
func TestPhase35OriginPeerBaseline(t *testing.T) {
	program := phase35OriginCheckedCore(t, "module phase35.origin_peer_baseline\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let value = 42\n  value\n}\n")
	function := program.Functions[0]
	if function.Linear == nil || len(function.Linear.Operations) < 2 || function.Linear.Operations[0].Kind != core.OpConst || function.Linear.Operations[0].ConstU64 != "42" {
		t.Fatalf("baseline OpConst shape changed: %+v", function.Linear)
	}
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("originvalidate stopped accepting the existing OpConst program: %+v", problems)
	}
}

// TestPhase35ShiftImmediateAdmission makes originvalidate admit and recompute
// a valid immediate shift, then checks forged raw fields independently. The
// peer sees only IR facts and makes no source-token provenance claim.
func TestPhase35ShiftImmediateAdmission(t *testing.T) {
	program := phase35OriginCheckedCore(t, "module phase35.origin_shift\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let limit = 16\n  let flag = input < limit\n  let shifted = input << (3)\n  shifted\n}\n")
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("valid immediate shift was rejected by origin publication: %+v", problems)
	}
	function := program.Functions[0]
	shift := phase35OriginFindOperation(t, &function, core.OpShiftLeftU64)
	types := phase35OriginTypes(&function)
	places := phase35OriginPlaces(&function)
	if shift.ConstU64 != "3" || shift.RightID != "" || types[shift.TypeID] != "U64" || types[places[shift.SourceID].TypeID] != "U64" || types[places[shift.TargetID].TypeID] != "U64" || shift.TargetID == function.Parameter.ID || shift.TargetID == shift.SourceID {
		t.Fatalf("raw shift lacks a fresh U64 target and canonical immediate: %+v", *shift)
	}
	if phase35OriginHasRemainder(&function) {
		t.Fatalf("pure shift fixture unexpectedly contains a remainder operation")
	}
	if returns := originvalidate.RecomputeOriginPerReturn(function, nil); len(returns) != 1 || returns[0].OperationID != phase35OriginOperationID(t, &function, core.OpReturn) {
		t.Fatalf("pure shift did not reach per-return origin recomputation: %+v", returns)
	}
	if problems := originvalidate.PublishProblemsFor(function, nil); len(problems) != 0 {
		t.Fatalf("pure shift without a remainder did not publish cleanly: %+v", problems)
	}

	flagID := phase35OriginPlaceID(t, &function, "flag")
	inputID := function.Parameter.ID
	returnIndex := phase35OriginOperationIndex(t, &function, core.OpReturn)
	tests := []struct {
		name string
		edit func(*core.Function, *core.LinearOperation)
		want string
	}{
		{name: "missing immediate", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "" }, want: "core.scalar_operation_invalid"},
		{name: "leading zero", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "03" }, want: "core.scalar_operation_invalid"},
		{name: "signed immediate", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "+3" }, want: "core.scalar_operation_invalid"},
		{name: "hex immediate", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "0x3" }, want: "core.scalar_operation_invalid"},
		{name: "count 64", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "64" }, want: "core.scalar_operation_invalid"},
		{name: "maximum u64", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "18446744073709551615" }, want: "core.scalar_operation_invalid"},
		{name: "overflow", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "18446744073709551616" }, want: "core.scalar_operation_invalid"},
		{name: "extra right place", edit: func(_ *core.Function, op *core.LinearOperation) { op.RightID = inputID }, want: "core.scalar_operation_invalid"},
		{name: "wrong source type", edit: func(_ *core.Function, op *core.LinearOperation) { op.SourceID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "wrong result type", edit: func(function *core.Function, op *core.LinearOperation) {
			op.TypeID = phase35OriginTypeID(t, function, "Bool")
		}, want: "core.scalar_operation_invalid"},
		{name: "parameter target", edit: func(function *core.Function, op *core.LinearOperation) { op.TargetID = function.Parameter.ID }, want: "core.scalar_operation_invalid"},
		{name: "source aliases target", edit: func(_ *core.Function, op *core.LinearOperation) { op.TargetID = op.SourceID }, want: "core.scalar_operation_invalid"},
		{name: "store field", edit: func(_ *core.Function, op *core.LinearOperation) { op.StoreTargetID = inputID }, want: "core.scalar_operation_invalid"},
		{name: "edge field", edit: func(_ *core.Function, op *core.LinearOperation) { op.TrueEdgeID = "forged-edge" }, want: "core.scalar_operation_invalid"},
		{name: "callee field", edit: func(_ *core.Function, op *core.LinearOperation) { op.CalleeID = "forged-callee" }, want: "core.scalar_operation_invalid"},
		{name: "payload field", edit: func(_ *core.Function, op *core.LinearOperation) { op.PayloadType = "forged-payload" }, want: "core.scalar_operation_invalid"},
		{name: "foreign field", edit: func(_ *core.Function, op *core.LinearOperation) { op.Foreign = &core.ForeignOperationContract{} }, want: "core.scalar_operation_invalid"},
		{name: "unrelated const field", edit: func(function *core.Function, _ *core.LinearOperation) {
			function.Linear.Operations[returnIndex].ConstU64 = "3"
		}, want: "core.scalar_operation_invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutant := phase35CloneOriginProgram(program)
			mutantFunction := &mutant.Functions[0]
			mutantShift := phase35OriginFindOperation(t, mutantFunction, core.OpShiftLeftU64)
			test.edit(mutantFunction, mutantShift)
			problems := originvalidate.PublishProblemsFor(*mutantFunction, nil)
			if !phase35OriginHasProblem(problems, test.want) {
				t.Fatalf("raw shift mutant problems=%+v, want independent refusal %q", problems, test.want)
			}
		})
	}
	if problems := originvalidate.PublishProblemsFor(function, nil); len(problems) != 0 {
		t.Fatalf("clean origin observation failed after copied controls: %+v", problems)
	}
}

// TestPhase35UnusedByteIRAdmission independently confines Byte to the unused
// formal and keeps all local computation in the U64/Bool subset.
func TestPhase35UnusedByteIRAdmission(t *testing.T) {
	program := phase35OriginCheckedCore(t, "module phase35.origin_byte\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let value = 42\n  let flag = value < 64\n  let shifted = value << 1\n  shifted\n}\n")
	function := program.Functions[0]
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("originvalidate rejected valid unused-Byte bridge: %+v", problems)
	}
	shiftIndex := phase35OriginOperationIndex(t, &function, core.OpShiftLeftU64)
	lessIndex := phase35OriginOperationIndex(t, &function, core.OpLessU64)
	returnIndex := phase35OriginOperationIndex(t, &function, core.OpReturn)
	tests := []struct {
		name string
		edit func(*core.Function)
		want string
	}{
		{name: "read as shift source", edit: func(fn *core.Function) { fn.Linear.Operations[shiftIndex].SourceID = fn.Parameter.ID }, want: "core.scalar_byte_parameter_used"},
		{name: "copy from formal", edit: func(fn *core.Function) {
			op := &fn.Linear.Operations[shiftIndex]
			op.Kind = core.OpCopy
			op.SourceID = fn.Parameter.ID
			op.RightID = ""
			op.ConstU64 = ""
		}, want: "core.scalar_byte_parameter_used"},
		{name: "read as right operand", edit: func(fn *core.Function) { fn.Linear.Operations[lessIndex].RightID = fn.Parameter.ID }, want: "core.scalar_byte_parameter_used"},
		{name: "write into formal", edit: func(fn *core.Function) { fn.Linear.Operations[shiftIndex].TargetID = fn.Parameter.ID }, want: "core.scalar_byte_parameter_used"},
		{name: "store into formal", edit: func(fn *core.Function) {
			op := &fn.Linear.Operations[shiftIndex]
			op.Kind = core.OpScalarStore
			op.TargetID = ""
			op.RightID = ""
			op.ConstU64 = ""
			op.StoreTargetID = fn.Parameter.ID
		}, want: "core.scalar_byte_parameter_used"},
		{name: "return formal", edit: func(fn *core.Function) { fn.Linear.Operations[returnIndex].SourceID = fn.Parameter.ID }, want: "core.scalar_byte_parameter_used"},
		{name: "payload use", edit: func(fn *core.Function) { fn.Linear.Operations[returnIndex].PayloadTargetID = fn.Parameter.ID }, want: "core.scalar_byte_parameter_used"},
		{name: "wrong return type", edit: func(fn *core.Function) { fn.ReturnType = "Byte" }, want: "core.scalar_operation_invalid"},
		{name: "non-scalar local", edit: func(fn *core.Function) { fn.Linear.Places[1].TypeID = phase35OriginTypeID(t, fn, "Byte") }, want: "core.scalar_operation_invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutant := phase35CloneOriginProgram(program)
			mutantFunction := &mutant.Functions[0]
			test.edit(mutantFunction)
			problems := originvalidate.PublishProblemsFor(*mutantFunction, nil)
			if !phase35OriginHasProblem(problems, test.want) {
				t.Fatalf("Byte-use mutant problems=%+v, want independent refusal %q", problems, test.want)
			}
		})
	}
	if problems := originvalidate.PublishProblemsFor(function, nil); len(problems) != 0 {
		t.Fatalf("clean origin bridge failed after copied controls: %+v", problems)
	}
}

func phase35OriginCheckedCore(t *testing.T, source string) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 35 origin fixture did not parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("Phase 35 origin fixture did not check: diagnostics=%+v functions=%d", checked.Diagnostics, len(checked.Program.Functions))
	}
	return checked.Program
}

func phase35CloneOriginProgram(program core.Program) core.Program {
	data, err := json.Marshal(program)
	if err != nil {
		panic(err)
	}
	var clone core.Program
	if err := json.Unmarshal(data, &clone); err != nil {
		panic(err)
	}
	return clone
}

func phase35OriginFindOperation(t *testing.T, function *core.Function, kind core.OperationKind) *core.LinearOperation {
	t.Helper()
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == kind {
			return &function.Linear.Operations[index]
		}
	}
	t.Fatalf("missing %s operation in %+v", kind, function.Linear.Operations)
	return nil
}

func phase35OriginOperationIndex(t *testing.T, function *core.Function, kind core.OperationKind) int {
	t.Helper()
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == kind {
			return index
		}
	}
	t.Fatalf("missing %s operation in %+v", kind, function.Linear.Operations)
	return -1
}

func phase35OriginOperationID(t *testing.T, function *core.Function, kind core.OperationKind) string {
	t.Helper()
	return phase35OriginFindOperation(t, function, kind).ID
}

func phase35OriginHasRemainder(function *core.Function) bool {
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpRemainderU64 {
			return true
		}
	}
	return false
}

func phase35OriginPlaces(function *core.Function) map[string]core.Place {
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	return places
}

func phase35OriginTypes(function *core.Function) map[string]string {
	types := make(map[string]string, len(function.Linear.Types))
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	return types
}

func phase35OriginPlaceID(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, place := range function.Linear.Places {
		if place.Name == name {
			return place.ID
		}
	}
	t.Fatalf("missing place %q", name)
	return ""
}

func phase35OriginTypeID(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, fact := range function.Linear.Types {
		if fact.Shape.Constructor == name {
			return fact.ID
		}
	}
	t.Fatalf("missing %s type fact", name)
	return ""
}

func phase35OriginHasProblem(problems []originvalidate.Problem, code string) bool {
	for _, problem := range problems {
		if problem.Code == code {
			return true
		}
	}
	return false
}

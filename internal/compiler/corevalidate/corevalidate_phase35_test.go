package corevalidate

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// TestPhase35CorePeerBaseline pins the existing OpConst path before the
// independent Phase 35 immediate and transfer laws are added.
func TestPhase35CorePeerBaseline(t *testing.T) {
	program := phase35CheckedCore(t, "module phase35.core_peer_baseline\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let value = 42\n  value\n}\n")
	function := program.Functions[0]
	if function.Linear == nil || len(function.Linear.Operations) < 2 || function.Linear.Operations[0].Kind != core.OpConst || function.Linear.Operations[0].ConstU64 != "42" {
		t.Fatalf("baseline OpConst shape changed: %+v", function.Linear)
	}
	if result := Validate(program); !result.Valid {
		t.Fatalf("corevalidate stopped accepting the existing OpConst program: %+v", result.Problems)
	}
}

// TestPhase35ShiftImmediateAdmission proves this peer accepts the raw IR
// immediate contract and re-derives malformed amounts, types, targets and
// fields. The producer's AST provenance is intentionally outside this peer's
// authority.
func TestPhase35ShiftImmediateAdmission(t *testing.T) {
	program := phase35CheckedCore(t, "module phase35.core_shift\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let limit = 16\n  let flag = input < limit\n  let shifted = input << (3)\n  shifted\n}\n")
	if got := Validate(program); !got.Valid {
		t.Fatalf("valid raw shift was rejected: %+v", got.Problems)
	}
	function := &program.Functions[0]
	shift := phase35FindOperation(t, function, core.OpShiftLeftU64)
	places := phase35PlacesByID(function)
	types := phase35TypesByID(function)
	if shift.ConstU64 != "3" || shift.RightID != "" || types[shift.TypeID] != "U64" || places[shift.SourceID].TypeID != shift.TypeID || places[shift.TargetID].TypeID != shift.TypeID || shift.TargetID == function.Parameter.ID || shift.TargetID == shift.SourceID {
		t.Fatalf("raw shift does not have a fresh U64 target and canonical immediate: op=%+v", *shift)
	}
	if !hasScalarCFG(function) {
		t.Fatal("shift operation did not select the scalar analysis route")
	}

	flagID := phase35PlaceID(t, function, "flag")
	inputID := function.Parameter.ID
	returnIndex := phase35OperationIndex(t, function, core.OpReturn)
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
			op.TypeID = phase35TypeID(t, function, "Bool")
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
		}, want: "core.scalar_field_kind_exclusive"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutant := cloneProgram(program)
			mutantFunction := &mutant.Functions[0]
			mutantShift := phase35FindOperation(t, mutantFunction, core.OpShiftLeftU64)
			test.edit(mutantFunction, mutantShift)
			got := Validate(mutant)
			if got.Valid || !phase35HasCoreProblem(got.Problems, test.want) {
				t.Fatalf("raw shift mutant verdict=%+v, want refusal %q", got, test.want)
			}
		})
	}
	if got := Validate(program); !got.Valid {
		t.Fatalf("clean raw shift failed after copied controls: %+v", got.Problems)
	}
}

// TestPhase35UnusedByteIRAdmission checks the bridge at the raw-core boundary.
// Only the formal may be Byte; local places and the returned value remain in
// the independently admitted U64/Bool scalar subset.
func TestPhase35UnusedByteIRAdmission(t *testing.T) {
	program := phase35CheckedCore(t, "module phase35.core_byte\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let value = 42\n  let flag = value < 64\n  let shifted = value << 1\n  shifted\n}\n")
	function := &program.Functions[0]
	if !scalarByteFormalUnused(function) {
		t.Fatal("checker-produced unused Byte bridge does not satisfy the narrow IR law")
	}
	if got := Validate(program); !got.Valid {
		t.Fatalf("valid unused-Byte bridge was rejected: %+v", got.Problems)
	}
	for _, place := range function.Linear.Places {
		if place.ID == function.Parameter.ID {
			if phase35TypesByID(function)[place.TypeID] != "Byte" {
				t.Fatalf("formal lost its Byte type: %+v", place)
			}
			continue
		}
		if kind := phase35TypesByID(function)[place.TypeID]; kind != "U64" && kind != "Bool" {
			t.Fatalf("non-formal place %q has unsupported type %q", place.ID, kind)
		}
	}

	shiftIndex := phase35OperationIndex(t, function, core.OpShiftLeftU64)
	lessIndex := phase35OperationIndex(t, function, core.OpLessU64)
	returnIndex := phase35OperationIndex(t, function, core.OpReturn)
	useTests := []struct {
		name string
		edit func(*core.Function)
	}{
		{name: "read as shift source", edit: func(fn *core.Function) { fn.Linear.Operations[shiftIndex].SourceID = fn.Parameter.ID }},
		{name: "copy from formal", edit: func(fn *core.Function) {
			op := &fn.Linear.Operations[shiftIndex]
			op.Kind = core.OpCopy
			op.SourceID = fn.Parameter.ID
			op.RightID = ""
			op.ConstU64 = ""
		}},
		{name: "read as right operand", edit: func(fn *core.Function) { fn.Linear.Operations[lessIndex].RightID = fn.Parameter.ID }},
		{name: "write into formal", edit: func(fn *core.Function) { op := &fn.Linear.Operations[shiftIndex]; op.TargetID = fn.Parameter.ID }},
		{name: "store into formal", edit: func(fn *core.Function) {
			op := &fn.Linear.Operations[shiftIndex]
			op.Kind = core.OpScalarStore
			op.TargetID = ""
			op.RightID = ""
			op.ConstU64 = ""
			op.StoreTargetID = fn.Parameter.ID
		}},
		{name: "return formal", edit: func(fn *core.Function) { fn.Linear.Operations[returnIndex].SourceID = fn.Parameter.ID }},
		{name: "payload use", edit: func(fn *core.Function) { fn.Linear.Operations[returnIndex].PayloadTargetID = fn.Parameter.ID }},
		{name: "wrong return type", edit: func(fn *core.Function) { fn.ReturnType = "Byte" }},
		{name: "non-scalar local", edit: func(fn *core.Function) { fn.Linear.Places[1].TypeID = phase35TypeID(t, fn, "Byte") }},
	}
	for _, test := range useTests {
		t.Run(test.name, func(t *testing.T) {
			mutant := cloneProgram(program)
			mutantFunction := &mutant.Functions[0]
			test.edit(mutantFunction)
			got := Validate(mutant)
			if got.Valid || !phase35HasCoreProblem(got.Problems, "core.scalar_byte_parameter_used") {
				t.Fatalf("Byte-use mutant verdict=%+v, want core.scalar_byte_parameter_used", got)
			}
		})
	}
	if got := Validate(program); !got.Valid {
		t.Fatalf("clean unused-Byte bridge failed after copied controls: %+v", got.Problems)
	}
}

func phase35CheckedCore(t *testing.T, source string) core.Program {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 35 raw-core fixture did not parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("Phase 35 raw-core fixture did not check: diagnostics=%+v functions=%d", checked.Diagnostics, len(checked.Program.Functions))
	}
	return checked.Program
}

func phase35FindOperation(t *testing.T, function *core.Function, kind core.OperationKind) *core.LinearOperation {
	t.Helper()
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == kind {
			return &function.Linear.Operations[index]
		}
	}
	t.Fatalf("missing %s operation in %+v", kind, function.Linear.Operations)
	return nil
}

func phase35OperationIndex(t *testing.T, function *core.Function, kind core.OperationKind) int {
	t.Helper()
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == kind {
			return index
		}
	}
	t.Fatalf("missing %s operation in %+v", kind, function.Linear.Operations)
	return -1
}

func phase35PlacesByID(function *core.Function) map[string]core.Place {
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	return places
}

func phase35TypesByID(function *core.Function) map[string]string {
	types := make(map[string]string, len(function.Linear.Types))
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	return types
}

func phase35PlaceID(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, place := range function.Linear.Places {
		if place.Name == name {
			return place.ID
		}
	}
	t.Fatalf("missing place %q", name)
	return ""
}

func phase35TypeID(t *testing.T, function *core.Function, name string) string {
	t.Helper()
	for _, fact := range function.Linear.Types {
		if fact.Shape.Constructor == name {
			return fact.ID
		}
	}
	t.Fatalf("missing %s type fact", name)
	return ""
}

func phase35HasCoreProblem(problems []Problem, code string) bool {
	for _, problem := range problems {
		if problem.Code == code {
			return true
		}
	}
	return false
}

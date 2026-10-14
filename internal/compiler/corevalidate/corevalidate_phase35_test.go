package corevalidate

import (
	"os"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase35LegacyByteCompatibility keeps pre-bridge admission independent
// of the scalar bridge's unused-formal restriction.
func TestPhase35LegacyByteCompatibility(t *testing.T) {
	for _, path := range []string{
		"testdata/phase2/implicit_copy.schway",
		"testdata/phase07/call_basic.schway",
		"testdata/phase4/foreign_acquire_one.schway",
		"testdata/phase4/nonlocal_exit_probe.schway",
		"testdata/phase11/multi_function_zero_call.schway",
	} {
		t.Run(path, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(path))
			if err != nil {
				t.Fatal(err)
			}
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("legacy fixture parse: %+v", parsed.Diagnostics)
			}
			checked := check.Program(parsed.Program)
			if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) == 0 {
				t.Fatalf("legacy fixture admission: %+v", checked.Diagnostics)
			}
			for _, function := range checked.Program.Functions {
				if function.Parameter.Type != "Byte" || function.Linear == nil {
					continue
				}
				for _, fact := range function.Linear.Types {
					if fact.ID == function.ID+":type:2" && fact.Shape.Constructor == "Bool" ||
						fact.ID == function.ID+":type:3" && fact.Shape.Constructor == "Unit" {
						t.Fatalf("legacy fixture unexpectedly carries structured scalar type authority: %+v", fact)
					}
				}
			}
			if result := Validate(checked.Program); !result.Valid {
				t.Fatalf("legacy Byte core admission changed: %+v", result.Problems)
			}
		})
	}
}

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
	for _, kind := range []core.OperationKind{core.OpShiftLeftU64, core.OpShiftRightU64} {
		for _, amount := range []string{"0", "63"} {
			t.Run(string(kind)+"/count-"+amount, func(t *testing.T) {
				boundary := cloneProgram(program)
				boundaryShift := phase35FindOperation(t, &boundary.Functions[0], core.OpShiftLeftU64)
				boundaryShift.Kind = kind
				boundaryShift.ConstU64 = amount
				if got := Validate(boundary); !got.Valid {
					t.Fatalf("raw %s immediate %s should be valid without AST provenance: %+v", kind, amount, got.Problems)
				}
			})
		}
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
		{name: "leading whitespace", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = " 3" }, want: "core.scalar_operation_invalid"},
		{name: "trailing whitespace", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "3 " }, want: "core.scalar_operation_invalid"},
		{name: "count 64", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "64" }, want: "core.scalar_operation_invalid"},
		{name: "maximum u64", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "18446744073709551615" }, want: "core.scalar_operation_invalid"},
		{name: "overflow", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "18446744073709551616" }, want: "core.scalar_operation_invalid"},
		{name: "extra right place", edit: func(_ *core.Function, op *core.LinearOperation) { op.RightID = inputID }, want: "core.scalar_operation_invalid"},
		{name: "wrong source type", edit: func(_ *core.Function, op *core.LinearOperation) { op.SourceID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "wrong result type", edit: func(function *core.Function, op *core.LinearOperation) {
			op.TypeID = phase35TypeID(t, function, "Bool")
		}, want: "core.scalar_operation_invalid"},
		{name: "wrong result place type", edit: func(function *core.Function, op *core.LinearOperation) {
			for index := range function.Linear.Places {
				if function.Linear.Places[index].ID == op.TargetID {
					function.Linear.Places[index].TypeID = phase35TypeID(t, function, "Bool")
					phase35RewireShiftConsumers(function, op.TargetID, inputID)
					return
				}
			}
			t.Fatalf("shift target %q is absent", op.TargetID)
		}, want: "core.scalar_operation_invalid"},
		{name: "parameter target", edit: func(function *core.Function, op *core.LinearOperation) {
			phase35RewireShiftConsumers(function, op.TargetID, function.Parameter.ID)
			op.TargetID = function.Parameter.ID
		}, want: "core.scalar_operation_invalid"},
		{name: "source aliases target", edit: func(function *core.Function, op *core.LinearOperation) {
			phase35RewireShiftConsumers(function, op.TargetID, op.SourceID)
			op.TargetID = op.SourceID
		}, want: "core.scalar_operation_invalid"},
		{name: "uninitialized U64 source", edit: func(function *core.Function, op *core.LinearOperation) {
			placeID := function.ID + ":place:" + strconv.Itoa(len(function.Linear.Places))
			function.Linear.Places = append(function.Linear.Places, core.Place{ID: placeID, TypeID: phase35TypeID(t, function, "U64")})
			op.SourceID = placeID
			function.Linear.Operations[returnIndex].SourceID = inputID
		}, want: "core.scalar_shift_amount_invalid"},
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
			if got.Valid {
				t.Fatalf("ADMISSION_ACCEPTED: raw shift mutant was accepted; want refusal %q", test.want)
			}
			if !phase35HasCoreProblem(got.Problems, test.want) {
				t.Fatalf("ADMISSION_DIAGNOSTIC_MISMATCH: raw shift mutant verdict=%+v, want refusal %q", got, test.want)
			}
		})
	}
	if got := Validate(program); !got.Valid {
		t.Fatalf("clean raw shift failed after copied controls: %+v", got.Problems)
	}
}

// Keep raw admission edits independent of a later copy/return's own laws.
func phase35RewireShiftConsumers(function *core.Function, oldTarget, replacement string) {
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].SourceID == oldTarget {
			function.Linear.Operations[index].SourceID = replacement
		}
	}
}

// TestPhase35ShiftTargetFreshness reaches a shift target that aliases an
// earlier U64 definition while keeping the later consumer attached to that
// target. This isolates target freshness from a downstream uninitialized read.
func TestPhase35ShiftTargetFreshness(t *testing.T) {
	program := phase35CheckedCore(t, "module phase35.shift_target\nexport { fn main }\nfn main(input: U64) -> U64 {\n  var reused = input\n  let limit = 16\n  let flag = input < limit\n  let prior = input >> 1\n  let shifted = input << 3\n  shifted\n}\n")
	base := &program.Functions[0]
	for _, kind := range []core.OperationKind{core.OpShiftLeftU64, core.OpShiftRightU64} {
		for _, row := range []struct {
			name   string
			target string
		}{
			{name: "prior constant", target: phase35FindOperation(t, base, core.OpConst).TargetID},
			{name: "prior copy", target: phase35PlaceID(t, base, "prior")},
			{name: "prior mutable store", target: phase35PlaceID(t, base, "reused")},
			{name: "prior shift", target: phase35FindOperation(t, base, core.OpShiftRightU64).TargetID},
		} {
			t.Run(string(kind)+"/"+row.name, func(t *testing.T) {
				mutant := cloneProgram(program)
				function := &mutant.Functions[0]
				shift := phase35FindOperation(t, function, core.OpShiftLeftU64)
				oldTargetID := shift.TargetID
				shift.Kind, shift.TargetID = kind, row.target
				for index := range function.Linear.Operations {
					if function.Linear.Operations[index].SourceID == oldTargetID {
						function.Linear.Operations[index].SourceID = row.target
					}
				}
				got := Validate(mutant)
				if got.Valid {
					t.Fatalf("ADMISSION_ACCEPTED: initialized shift target reuse; want core.invalid_target")
				}
				if !phase35HasCoreProblem(got.Problems, "core.invalid_target") {
					t.Fatalf("ADMISSION_DIAGNOSTIC_MISMATCH: initialized shift target reuse: verdict=%+v, want core.invalid_target", got)
				}
			})
		}
	}
}

// TestPhase35ShiftProducerShapes distinguishes fresh expression temporaries
// from mutable assignment destinations and repeated execution of a loop body.
func TestPhase35ShiftProducerShapes(t *testing.T) {
	for _, row := range []struct {
		name string
		body string
	}{
		{name: "assignment", body: "  var value = input\n  value = value << 1\n  value = value >> 1\n  value\n"},
		{name: "branch", body: "  var value = input\n  if input < 16 {\n    value = value << 1\n  } else {\n    value = value >> 1\n  }\n  value\n"},
		{name: "loop", body: "  var value = input\n  var i = 0\n  while i < 3 {\n    value = value << 1\n    value = value >> 1\n    i = i + 1\n  }\n  value\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			program := phase35CheckedCore(t, "module phase35.shift_shapes\nexport { fn main }\nfn main(input: U64) -> U64 {\n"+row.body+"}\n")
			if row.name == "loop" {
				backEdge := false
				for _, edge := range program.Functions[0].Linear.Edges {
					backEdge = backEdge || edge.Pattern == "back"
				}
				if !backEdge {
					t.Fatal("loop control did not produce a backedge")
				}
			}
			result := Validate(program)
			if !result.Valid {
				t.Fatalf("producer shift shape refused: %+v", result.Problems)
			}
			if result.PeerSignatures()[program.Functions[0].ID].ClosureDigest == "" {
				t.Fatal("producer control stopped replay without finishing validation")
			}
		})
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

	t.Run("constant-only fake bridge", func(t *testing.T) {
		constant := phase35CheckedCore(t, "module phase35.constant_byte\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let value = 42\n  value\n}\n")
		base := &constant.Functions[0]
		for _, operation := range base.Linear.Operations {
			if operation.Kind != core.OpConst && operation.Kind != core.OpReturn {
				t.Fatalf("constant-only discriminator control contains %s", operation.Kind)
			}
		}
		if result := Validate(constant); !result.Valid {
			t.Fatalf("clean constant-only bridge refused: %+v", result.Problems)
		}
		for _, row := range []struct {
			name string
			edit func(*core.Function)
			want string
		}{
			{name: "used formal without scalar instruction", edit: func(fn *core.Function) {
				op := &fn.Linear.Operations[0]
				op.Kind, op.SourceID, op.ConstU64 = core.OpCopy, fn.Parameter.ID, ""
			}, want: "core.scalar_byte_parameter_used"},
			{name: "wrong return without scalar instruction", edit: func(fn *core.Function) {
				fn.ReturnType = "Byte"
			}, want: "core.scalar_byte_parameter_used"},
			{name: "non-scalar local without scalar instruction", edit: func(fn *core.Function) {
				fn.Linear.Places[1].TypeID = phase35TypeID(t, fn, "Byte")
			}, want: "core.scalar_byte_parameter_used"},
			{name: "unreachable used formal", edit: func(fn *core.Function) {
				// This flat operation is deliberately absent from every block.
				// Applicability and the unused law must inspect it nevertheless.
				op := fn.Linear.Operations[0]
				op.ID, op.PointID = fn.ID+":op:unreachable", fn.ID+":point:unreachable"
				op.Kind, op.SourceID, op.ConstU64 = core.OpCopy, fn.Parameter.ID, ""
				fn.Linear.Operations = append(fn.Linear.Operations, op)
			}, want: "core.scalar_byte_parameter_used"},
			{name: "foreign metadata cannot exempt used formal", edit: func(fn *core.Function) {
				fn.ForeignContract = &core.ForeignContract{Symbol: "forged"}
				op := &fn.Linear.Operations[0]
				op.Kind, op.SourceID, op.ConstU64 = core.OpCopy, fn.Parameter.ID, ""
			}, want: "core.scalar_byte_parameter_used"},
			{name: "call opcode cannot exempt used formal", edit: func(fn *core.Function) {
				op := &fn.Linear.Operations[0]
				op.Kind, op.SourceID, op.ConstU64 = core.OpCall, fn.Parameter.ID, ""
				op.CalleeID = "forged-callee"
			}, want: "core.scalar_byte_parameter_used"},
		} {
			t.Run(row.name, func(t *testing.T) {
				mutant := cloneProgram(constant)
				row.edit(&mutant.Functions[0])
				result := Validate(mutant)
				if result.Valid || !phase35HasCoreProblem(result.Problems, row.want) {
					t.Fatalf("fake constant-only bridge escaped %s: %+v", row.want, result)
				}
			})
		}
	})
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

// TestPhase35ScalarTransferMutants compares each independently derived
// bitwise/shift result with a separate expected constant before reaching the
// static zero-remainder sink. Input mutations remain copy-local controls.
func TestPhase35ScalarTransferMutants(t *testing.T) {
	program := phase35CheckedCore(t, "module phase35.core_transfer\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let left = 0\n  let right = input\n  let actual = left ^ right\n  let expected = 0\n  let mismatch = actual ^ expected\n  let one = 1\n  let remainder = one % mismatch\n  remainder\n}\n")
	if len(program.Functions) != 1 {
		t.Fatalf("transfer fixture functions=%d", len(program.Functions))
	}
	base := &program.Functions[0]
	leftID := phase35PlaceID(t, base, "left")
	rightID := phase35PlaceID(t, base, "right")
	expectedID := phase35PlaceID(t, base, "expected")
	type vector struct {
		name        string
		kind        core.OperationKind
		left, right uint64
		amount      string
		want        uint64
		faultRight  uint64
		faultAmount string
	}
	vectors := []vector{
		{name: "xor zero", kind: core.OpBitXorU64, left: 5, right: 5, want: 0, faultRight: 6},
		{name: "and zero", kind: core.OpBitAndU64, left: 1, right: 2, want: 0, faultRight: 1},
		{name: "or overlapping bit", kind: core.OpBitOrU64, left: 1, right: 1, want: 1, faultRight: 2},
		{name: "shift left zero", kind: core.OpShiftLeftU64, left: 1 << 63, amount: "1", want: 0, faultAmount: "0"},
		{name: "shift right one", kind: core.OpShiftRightU64, left: 1 << 63, amount: "63", want: 1, faultAmount: "0"},
	}
	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			mutant := cloneProgram(program)
			function := &mutant.Functions[0]
			phase35SetConstTarget(t, function, leftID, tc.left)
			phase35SetConstTarget(t, function, rightID, tc.right)
			phase35SetConstTarget(t, function, expectedID, tc.want)
			op := phase35OperationForSource(t, function, core.OpBitXorU64, leftID)
			op.Kind = tc.kind
			op.SourceID = leftID
			if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
				op.RightID = ""
				op.ConstU64 = tc.amount
			} else {
				op.RightID = rightID
				op.ConstU64 = ""
			}
			result := Validate(mutant)
			if result.Valid || !phase35HasCoreProblem(result.Problems, "core.scalar_remainder_zero") {
				t.Fatalf("%s transfer failed to reach the static zero-remainder sink: problems=%+v", tc.name, result.Problems)
			}

			faulted := cloneProgram(mutant)
			faultFunction := &faulted.Functions[0]
			if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
				phase35FindOperation(t, faultFunction, tc.kind).ConstU64 = tc.faultAmount
			} else {
				phase35SetConstTarget(t, faultFunction, rightID, tc.faultRight)
			}
			if got := Validate(faulted); !got.Valid {
				t.Fatalf("wrong-result control still reached the zero sink: %+v", got.Problems)
			}
			if restored := Validate(mutant); restored.Valid || !phase35HasCoreProblem(restored.Problems, "core.scalar_remainder_zero") {
				t.Fatalf("restored %s transfer no longer reaches the named sink: %+v", tc.name, restored.Problems)
			}
		})
	}

	cfgProgram := phase35CheckedCore(t, "module phase35.core_transfer_cfg\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let left = 0\n  let right = input\n  var actual = input\n  if input < 8 {\n    actual = left ^ right\n  } else {\n    actual = left ^ right\n  }\n  let expected = 0\n  let mismatch = actual ^ expected\n  let one = 1\n  let remainder = one % mismatch\n  remainder\n}\n")
	cfgBase := &cfgProgram.Functions[0]
	cfgLeftID := phase35PlaceID(t, cfgBase, "left")
	cfgRightID := phase35PlaceID(t, cfgBase, "right")
	cfgExpectedID := phase35PlaceID(t, cfgBase, "expected")
	cfgInputID := cfgBase.Parameter.ID
	for _, tc := range vectors {
		t.Run("cfg/"+tc.name, func(t *testing.T) {
			mutant := cloneProgram(cfgProgram)
			function := &mutant.Functions[0]
			if len(function.Linear.Blocks) < 3 {
				t.Fatalf("transfer fixture did not retain a branch CFG: blocks=%d", len(function.Linear.Blocks))
			}
			phase35SetConstTarget(t, function, cfgLeftID, tc.left)
			phase35SetConstTarget(t, function, cfgRightID, tc.right)
			phase35SetConstTarget(t, function, cfgExpectedID, tc.want)
			operations := phase35OperationsForSource(function, core.OpBitXorU64, cfgLeftID)
			if len(operations) != 2 {
				t.Fatalf("CFG fixture has %d tested branch operations, want 2", len(operations))
			}
			for _, operation := range operations {
				operation.Kind = tc.kind
				operation.SourceID = cfgLeftID
				if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
					operation.RightID = ""
					operation.ConstU64 = tc.amount
				} else {
					operation.RightID = cfgRightID
					operation.ConstU64 = ""
				}
			}
			result := Validate(mutant)
			if result.Valid || !phase35HasCoreProblem(result.Problems, "core.scalar_remainder_zero") {
				t.Fatalf("CFG %s transfer failed to reach the static zero-remainder sink: %+v", tc.name, result.Problems)
			}

			faulted := cloneProgram(mutant)
			faultOps := phase35OperationsForSource(&faulted.Functions[0], tc.kind, cfgLeftID)
			if len(faultOps) != 2 {
				t.Fatalf("fault copy has %d branch operations, want 2", len(faultOps))
			}
			for _, operation := range faultOps {
				if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
					operation.ConstU64 = tc.faultAmount
				} else {
					operation.RightID = cfgInputID
				}
			}
			if got := Validate(faulted); !got.Valid {
				t.Fatalf("CFG wrong-result control still reached the zero sink: %+v", got.Problems)
			}
			if restored := Validate(mutant); restored.Valid || !phase35HasCoreProblem(restored.Problems, "core.scalar_remainder_zero") {
				t.Fatalf("restored CFG %s transfer no longer reaches the named sink: %+v", tc.name, restored.Problems)
			}
		})
	}
}

func phase35OperationForSource(t *testing.T, function *core.Function, kind core.OperationKind, sourceID string) *core.LinearOperation {
	t.Helper()
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == kind && operation.SourceID == sourceID {
			return operation
		}
	}
	t.Fatalf("missing %s operation from source %q", kind, sourceID)
	return nil
}

func phase35OperationsForSource(function *core.Function, kind core.OperationKind, sourceID string) []*core.LinearOperation {
	var operations []*core.LinearOperation
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == kind && operation.SourceID == sourceID {
			operations = append(operations, operation)
		}
	}
	return operations
}

// TestPhase35AnalysisTriggers pins each pure operator as an independent
// scalar-analysis trigger and exercises shifts inside a real branch CFG.
func TestPhase35AnalysisTriggers(t *testing.T) {
	operators := []struct {
		name string
		expr string
		kind core.OperationKind
	}{
		{name: "xor", expr: "input ^ 1", kind: core.OpBitXorU64},
		{name: "and", expr: "input & 1", kind: core.OpBitAndU64},
		{name: "or", expr: "input | 1", kind: core.OpBitOrU64},
		{name: "left shift", expr: "input << 1", kind: core.OpShiftLeftU64},
		{name: "right shift", expr: "input >> 1", kind: core.OpShiftRightU64},
	}
	for _, tc := range operators {
		t.Run(tc.name, func(t *testing.T) {
			program := phase35CheckedCore(t, "module phase35.core_trigger\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let result = "+tc.expr+"\n  result\n}\n")
			function := &program.Functions[0]
			if phase35FindOperation(t, function, tc.kind) == nil || !hasScalarCFG(function) {
				t.Fatalf("%s did not independently select core scalar analysis", tc.name)
			}
			if result := Validate(program); !result.Valid {
				t.Fatalf("pure %s trigger without remainder failed validation: %+v", tc.name, result.Problems)
			}
		})
	}

	branch := phase35CheckedCore(t, "module phase35.core_shift_cfg\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let flag = input < 8\n  var shifted = input\n  if flag {\n    shifted = input << 3\n  } else {\n    shifted = input >> 2\n  }\n  shifted\n}\n")
	function := &branch.Functions[0]
	if len(function.Linear.Blocks) < 3 || !hasScalarCFG(function) {
		t.Fatalf("shift branch fixture did not produce a scalar CFG: blocks=%d", len(function.Linear.Blocks))
	}
	if result := Validate(branch); !result.Valid {
		t.Fatalf("branch shift fixed-point transfer failed validation: %+v", result.Problems)
	}
	t.Run("field-laws", phase35CoreScalarFieldLaws)
}

// phase35CoreScalarFieldLaws keeps wrong-kind immediates and unrelated
// raw operation fields behind corevalidate's own refusal rows.
func phase35CoreScalarFieldLaws(t *testing.T) {
	program := phase35CheckedCore(t, "module phase35.core_fields\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let one = 1\n  let sum = input + one\n  let less = sum < 8\n  let equal = sum == 8\n  let xor = sum ^ one\n  let anded = xor & one\n  let ored = anded | one\n  let left = ored << 1\n  let right = left >> 1\n  let remainder = right % 7\n  remainder\n}\n")
	kindRows := []core.OperationKind{
		core.OpAddChecked,
		core.OpLessU64,
		core.OpEqualU64,
		core.OpRemainderU64,
		core.OpBitXorU64,
		core.OpBitAndU64,
		core.OpBitOrU64,
	}
	for _, kind := range kindRows {
		t.Run("ConstU64 forbidden on "+string(kind), func(t *testing.T) {
			mutant := cloneProgram(program)
			op := phase35FindOperation(t, &mutant.Functions[0], kind)
			op.ConstU64 = "1"
			got := Validate(mutant)
			if got.Valid || !phase35HasCoreProblem(got.Problems, "core.scalar_field_kind_exclusive") {
				t.Fatalf("wrong-kind ConstU64 on %s was not refused exactly: %+v", kind, got.Problems)
			}
		})
	}

	flagID := phase35PlaceID(t, &program.Functions[0], "less")
	type mutation struct {
		name string
		edit func(*core.Function, *core.LinearOperation)
		want string
	}
	mutations := []mutation{
		{name: "wrong source Bool", edit: func(_ *core.Function, op *core.LinearOperation) { op.SourceID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "wrong right Bool", edit: func(_ *core.Function, op *core.LinearOperation) { op.RightID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "wrong result type", edit: func(fn *core.Function, op *core.LinearOperation) { op.TypeID = phase35TypeID(t, fn, "Bool") }, want: "core.scalar_operation_invalid"},
		{name: "missing right operand", edit: func(_ *core.Function, op *core.LinearOperation) { op.RightID = "" }, want: "core.scalar_operation_invalid"},
		{name: "target is parameter", edit: func(fn *core.Function, op *core.LinearOperation) { op.TargetID = fn.Parameter.ID }, want: "core.place_uninitialized"},
		{name: "target aliases source", edit: func(_ *core.Function, op *core.LinearOperation) { op.TargetID = op.SourceID }, want: "core.place_uninitialized"},
		{name: "wrong target type", edit: func(_ *core.Function, op *core.LinearOperation) { op.TargetID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "store field", edit: func(fn *core.Function, op *core.LinearOperation) { op.StoreTargetID = fn.Parameter.ID }, want: "core.scalar_operation_invalid"},
		{name: "true edge field", edit: func(_ *core.Function, op *core.LinearOperation) { op.TrueEdgeID = "forged-edge" }, want: "core.scalar_operation_invalid"},
		{name: "false edge field", edit: func(_ *core.Function, op *core.LinearOperation) { op.FalseEdgeID = "forged-edge" }, want: "core.scalar_operation_invalid"},
		{name: "loan field", edit: func(_ *core.Function, op *core.LinearOperation) { op.LoanID = "forged-loan" }, want: "core.scalar_operation_invalid"},
		{name: "foreign field", edit: func(_ *core.Function, op *core.LinearOperation) { op.Foreign = &core.ForeignOperationContract{} }, want: "core.scalar_operation_invalid"},
		{name: "ok edge field", edit: func(_ *core.Function, op *core.LinearOperation) { op.OkEdgeID = "forged-edge" }, want: "core.scalar_operation_invalid"},
		{name: "error edge field", edit: func(_ *core.Function, op *core.LinearOperation) { op.ErrEdgeID = "forged-edge" }, want: "core.scalar_operation_invalid"},
		{name: "error target field", edit: func(fn *core.Function, op *core.LinearOperation) { op.ErrTargetID = fn.Parameter.ID }, want: "core.scalar_operation_invalid"},
		{name: "release field", edit: func(_ *core.Function, op *core.LinearOperation) { op.ReleasesOperationID = "forged-operation" }, want: "core.scalar_operation_invalid"},
		{name: "allocator field", edit: func(_ *core.Function, op *core.LinearOperation) { op.Allocator = "forged-allocator" }, want: "core.scalar_operation_invalid"},
		{name: "reason field", edit: func(_ *core.Function, op *core.LinearOperation) { op.Reason = "forged-reason" }, want: "core.scalar_operation_invalid"},
		{name: "callee field", edit: func(_ *core.Function, op *core.LinearOperation) { op.CalleeID = "forged-callee" }, want: "core.scalar_operation_invalid"},
		{name: "payload type field", edit: func(_ *core.Function, op *core.LinearOperation) { op.PayloadType = "forged-payload" }, want: "core.scalar_operation_invalid"},
		{name: "payload target field", edit: func(fn *core.Function, op *core.LinearOperation) { op.PayloadTargetID = fn.Parameter.ID }, want: "core.scalar_operation_invalid"},
		{name: "text source field", edit: func(_ *core.Function, op *core.LinearOperation) { op.TextSource = "\"forged\"" }, want: "core.scalar_operation_invalid"},
		{name: "text bytes field", edit: func(_ *core.Function, op *core.LinearOperation) { op.TextBytes = []byte("forged") }, want: "core.scalar_operation_invalid"},
		{name: "text length field", edit: func(_ *core.Function, op *core.LinearOperation) { op.TextLength = 6 }, want: "core.scalar_operation_invalid"},
		{name: "uninitialized U64 source", edit: func(fn *core.Function, op *core.LinearOperation) {
			op.SourceID = phase35PlaceID(t, fn, "right")
		}, want: "core.place_uninitialized"},
	}
	for _, kind := range []core.OperationKind{core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64} {
		for _, row := range mutations {
			t.Run(string(kind)+"/"+row.name, func(t *testing.T) {
				mutant := cloneProgram(program)
				function := &mutant.Functions[0]
				op := phase35FindOperation(t, function, kind)
				row.edit(function, op)
				got := Validate(mutant)
				if got.Valid || !phase35HasCoreProblem(got.Problems, row.want) {
					t.Fatalf("%s %s fault was not refused as %s: %+v", kind, row.name, row.want, got.Problems)
				}
			})
		}
	}
	if got := Validate(program); !got.Valid {
		t.Fatalf("clean field-law control failed after faults: %+v", got.Problems)
	}

	byteProgram := phase35CheckedCore(t, "module phase35.core_byte_source\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let value = 42\n  let flag = value < 64\n  let shifted = value << 1\n  let first = shifted ^ value\n  let second = first & value\n  let third = second | value\n  third\n}\n")
	for _, kind := range []core.OperationKind{core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64} {
		t.Run(string(kind)+"/wrong-source-Byte", func(t *testing.T) {
			mutant := cloneProgram(byteProgram)
			function := &mutant.Functions[0]
			phase35FindOperation(t, function, kind).SourceID = function.Parameter.ID
			got := Validate(mutant)
			if got.Valid || !phase35HasCoreProblem(got.Problems, "core.scalar_byte_parameter_used") {
				t.Fatalf("%s Byte source passed the unused-formal bridge: %+v", kind, got.Problems)
			}
		})
	}
}

func phase35SetConstTarget(t *testing.T, function *core.Function, target string, value uint64) {
	t.Helper()
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.TargetID != target {
			continue
		}
		operation.Kind = core.OpConst
		operation.SourceID = ""
		operation.RightID = ""
		operation.ConstU64 = strconv.FormatUint(value, 10)
		operation.StoreTargetID = ""
		operation.ErrTargetID = ""
		operation.PayloadTargetID = ""
		operation.PayloadType = ""
		operation.CalleeID = ""
		operation.Foreign = nil
		return
	}
	t.Fatalf("missing operation defining target %q", target)
}

func phase35OperationsOfKind(function *core.Function, kind core.OperationKind) []*core.LinearOperation {
	var operations []*core.LinearOperation
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == kind {
			operations = append(operations, &function.Linear.Operations[index])
		}
	}
	return operations
}

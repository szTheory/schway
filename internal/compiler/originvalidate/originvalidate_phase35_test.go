package originvalidate_test

import (
	"encoding/json"
	"strconv"
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
	for _, kind := range []core.OperationKind{core.OpShiftLeftU64, core.OpShiftRightU64} {
		for _, amount := range []string{"0", "63"} {
			t.Run(string(kind)+"/count-"+amount, func(t *testing.T) {
				boundary := phase35CloneOriginProgram(program)
				boundaryShift := phase35OriginFindOperation(t, &boundary.Functions[0], core.OpShiftLeftU64)
				boundaryShift.Kind = kind
				boundaryShift.ConstU64 = amount
				if problems := originvalidate.PublishProblemsFor(boundary.Functions[0], nil); len(problems) != 0 {
					t.Fatalf("origin peer refused raw %s immediate %s: %+v", kind, amount, problems)
				}
			})
		}
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
		{name: "leading whitespace", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = " 3" }, want: "core.scalar_operation_invalid"},
		{name: "trailing whitespace", edit: func(_ *core.Function, op *core.LinearOperation) { op.ConstU64 = "3 " }, want: "core.scalar_operation_invalid"},
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

// TestPhase35ScalarTransferMutants drives each independently implemented
// origin transfer into its own static zero-remainder refusal, then proves a
// copy-local wrong-result fault removes that observation and restoration
// reaches it again.
func TestPhase35ScalarTransferMutants(t *testing.T) {
	program := phase35OriginCheckedCore(t, "module phase35.origin_transfer\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let left = 0\n  let right = input\n  let zero = left ^ right\n  let one = 1\n  let remainder = one % zero\n  remainder\n}\n")
	base := &program.Functions[0]
	leftID := phase35OriginPlaceID(t, base, "left")
	rightID := phase35OriginPlaceID(t, base, "right")
	type vector struct {
		name        string
		kind        core.OperationKind
		left, right uint64
		amount      string
		faultRight  uint64
		faultAmount string
	}
	vectors := []vector{
		{name: "xor zero", kind: core.OpBitXorU64, left: 5, right: 5, faultRight: 6},
		{name: "and zero", kind: core.OpBitAndU64, left: 1, right: 2, faultRight: 1},
		{name: "or zero", kind: core.OpBitOrU64, left: 0, right: 0, faultRight: 1},
		{name: "shift left zero", kind: core.OpShiftLeftU64, left: 1 << 63, amount: "1", faultAmount: "0"},
		{name: "shift right zero", kind: core.OpShiftRightU64, left: 1, amount: "1", faultAmount: "0"},
	}
	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			mutant := phase35CloneOriginProgram(program)
			function := &mutant.Functions[0]
			phase35OriginSetConstTarget(t, function, leftID, tc.left)
			phase35OriginSetConstTarget(t, function, rightID, tc.right)
			op := phase35OriginFindOperation(t, function, core.OpBitXorU64)
			op.Kind = tc.kind
			op.SourceID = leftID
			if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
				op.RightID = ""
				op.ConstU64 = tc.amount
			} else {
				op.RightID = rightID
				op.ConstU64 = ""
			}
			if problems := originvalidate.PublishProblemsFor(*function, nil); !phase35OriginHasProblem(problems, "core.scalar_remainder_zero") {
				t.Fatalf("origin %s transfer missed its static zero-remainder sink: %+v", tc.name, problems)
			}

			faulted := phase35CloneOriginProgram(mutant)
			faultFunction := &faulted.Functions[0]
			if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
				phase35OriginFindOperation(t, faultFunction, tc.kind).ConstU64 = tc.faultAmount
			} else {
				phase35OriginSetConstTarget(t, faultFunction, rightID, tc.faultRight)
			}
			if problems := originvalidate.PublishProblemsFor(*faultFunction, nil); phase35OriginHasProblem(problems, "core.scalar_remainder_zero") {
				t.Fatalf("origin %s wrong-result control still reached zero sink: %+v", tc.name, problems)
			}
			if problems := originvalidate.PublishProblemsFor(*function, nil); !phase35OriginHasProblem(problems, "core.scalar_remainder_zero") {
				t.Fatalf("restored origin %s transfer did not reach zero sink: %+v", tc.name, problems)
			}
		})
	}

	cfgProgram := phase35OriginCheckedCore(t, "module phase35.origin_transfer_cfg\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let left = 0\n  let right = input\n  var divisor = input\n  if input < 8 {\n    divisor = left ^ right\n  } else {\n    divisor = left ^ right\n  }\n  let one = 1\n  let remainder = one % divisor\n  remainder\n}\n")
	cfgBase := &cfgProgram.Functions[0]
	cfgLeftID := phase35OriginPlaceID(t, cfgBase, "left")
	cfgRightID := phase35OriginPlaceID(t, cfgBase, "right")
	cfgInputID := cfgBase.Parameter.ID
	for _, tc := range vectors {
		t.Run("cfg/"+tc.name, func(t *testing.T) {
			mutant := phase35CloneOriginProgram(cfgProgram)
			function := &mutant.Functions[0]
			if len(function.Linear.Blocks) < 3 {
				t.Fatalf("origin transfer fixture did not retain branch blocks: %d", len(function.Linear.Blocks))
			}
			phase35OriginSetConstTarget(t, function, cfgLeftID, tc.left)
			phase35OriginSetConstTarget(t, function, cfgRightID, tc.right)
			operations := phase35OriginOperationsOfKind(function, core.OpBitXorU64)
			if len(operations) != 2 {
				t.Fatalf("origin CFG has %d tested branch operations, want 2", len(operations))
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
			if problems := originvalidate.PublishProblemsFor(*function, nil); !phase35OriginHasProblem(problems, "core.scalar_remainder_zero") {
				t.Fatalf("origin CFG %s transfer missed its zero sink: %+v", tc.name, problems)
			}

			faulted := phase35CloneOriginProgram(mutant)
			faultOps := phase35OriginOperationsOfKind(&faulted.Functions[0], tc.kind)
			if len(faultOps) != 2 {
				t.Fatalf("origin CFG fault copy has %d tested operations, want 2", len(faultOps))
			}
			for _, operation := range faultOps {
				if tc.kind == core.OpShiftLeftU64 || tc.kind == core.OpShiftRightU64 {
					operation.ConstU64 = tc.faultAmount
				} else {
					operation.RightID = cfgInputID
				}
			}
			if problems := originvalidate.PublishProblemsFor(faulted.Functions[0], nil); phase35OriginHasProblem(problems, "core.scalar_remainder_zero") {
				t.Fatalf("origin CFG %s wrong-result control still reached sink: %+v", tc.name, problems)
			}
			if problems := originvalidate.PublishProblemsFor(*function, nil); !phase35OriginHasProblem(problems, "core.scalar_remainder_zero") {
				t.Fatalf("restored origin CFG %s transfer did not reach sink: %+v", tc.name, problems)
			}
		})
	}
}

// TestPhase35AnalysisTriggers exercises origin admission, scalar trigger and
// per-return recomputation for every pure operation when no remainder exists.
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
			program := phase35OriginCheckedCore(t, "module phase35.origin_trigger\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let result = "+tc.expr+"\n  result\n}\n")
			function := &program.Functions[0]
			if phase35OriginHasRemainder(function) {
				t.Fatalf("pure %s fixture unexpectedly contains a remainder", tc.name)
			}
			if phase35OriginFindOperation(t, function, tc.kind).Kind != tc.kind {
				t.Fatalf("missing pure %s operation", tc.name)
			}
			if problems := originvalidate.PublishProblemsFor(*function, nil); len(problems) != 0 {
				t.Fatalf("origin did not publish valid pure %s without remainder: %+v", tc.name, problems)
			}
			returns := originvalidate.RecomputeOriginPerReturn(*function, nil)
			if len(returns) != 1 || returns[0].OperationID != phase35OriginOperationID(t, function, core.OpReturn) || returns[0].Derived {
				t.Fatalf("pure %s missed owned per-return recomputation: %+v", tc.name, returns)
			}

			mutant := phase35CloneOriginProgram(program)
			mutantFunction := &mutant.Functions[0]
			phase35OriginFindOperation(t, mutantFunction, tc.kind).SourceID = phase35OriginPlaceID(t, mutantFunction, "result")
			if problems := originvalidate.PublishProblemsFor(*mutantFunction, nil); !phase35OriginHasProblem(problems, "core.scalar_analysis_invalid") {
				t.Fatalf("malformed no-remainder %s did not reach origin scalar trigger: %+v", tc.name, problems)
			}
			if returns := originvalidate.RecomputeOriginPerReturn(*mutantFunction, nil); len(returns) != 0 {
				t.Fatalf("origin recomputed returns despite failed %s scalar transfer: %+v", tc.name, returns)
			}
		})
	}

	branch := phase35OriginCheckedCore(t, "module phase35.origin_shift_cfg\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let flag = input < 8\n  var shifted = input\n  if flag {\n    shifted = input << 3\n  } else {\n    shifted = input >> 2\n  }\n  shifted\n}\n")
	function := &branch.Functions[0]
	if len(function.Linear.Blocks) < 3 || phase35OriginHasRemainder(function) {
		t.Fatalf("origin no-remainder shift CFG not reached: blocks=%d remainder=%v", len(function.Linear.Blocks), phase35OriginHasRemainder(function))
	}
	if problems := originvalidate.PublishProblemsFor(*function, nil); len(problems) != 0 {
		t.Fatalf("origin refused pure shift CFG publication: %+v", problems)
	}
	if returns := originvalidate.RecomputeOriginPerReturn(*function, nil); len(returns) != 1 || returns[0].OperationID != phase35OriginOperationID(t, function, core.OpReturn) {
		t.Fatalf("origin did not recompute return through pure shift CFG: %+v", returns)
	}
	t.Run("field-laws", phase35OriginScalarFieldLaws)
}

func phase35OriginScalarFieldLaws(t *testing.T) {
	t.Helper()
	program := phase35OriginCheckedCore(t, "module phase35.origin_fields\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let one = 1\n  let sum = input + one\n  let less = sum < 8\n  let equal = sum == 8\n  let xor = sum ^ one\n  let anded = xor & one\n  let ored = anded | one\n  let left = ored << 1\n  let right = left >> 1\n  let remainder = right % 7\n  remainder\n}\n")
	kindRows := []core.OperationKind{core.OpAddChecked, core.OpLessU64, core.OpEqualU64, core.OpRemainderU64, core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64}
	for _, kind := range kindRows {
		t.Run("ConstU64 forbidden on "+string(kind), func(t *testing.T) {
			mutant := phase35CloneOriginProgram(program)
			phase35OriginFindOperation(t, &mutant.Functions[0], kind).ConstU64 = "1"
			if problems := originvalidate.PublishProblemsFor(mutant.Functions[0], nil); !phase35OriginHasProblem(problems, "core.scalar_operation_invalid") {
				t.Fatalf("origin admitted wrong-kind ConstU64 on %s: %+v", kind, problems)
			}
		})
	}

	flagID := phase35OriginPlaceID(t, &program.Functions[0], "less")
	type mutation struct {
		name string
		edit func(*core.Function, *core.LinearOperation)
		want string
	}
	mutations := []mutation{
		{name: "wrong source Bool", edit: func(_ *core.Function, op *core.LinearOperation) { op.SourceID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "wrong right Bool", edit: func(_ *core.Function, op *core.LinearOperation) { op.RightID = flagID }, want: "core.scalar_operation_invalid"},
		{name: "wrong result type", edit: func(fn *core.Function, op *core.LinearOperation) { op.TypeID = phase35OriginTypeID(t, fn, "Bool") }, want: "core.scalar_operation_invalid"},
		{name: "missing right operand", edit: func(_ *core.Function, op *core.LinearOperation) { op.RightID = "" }, want: "core.scalar_operation_invalid"},
		{name: "target is parameter", edit: func(fn *core.Function, op *core.LinearOperation) { op.TargetID = fn.Parameter.ID }, want: "core.scalar_operation_invalid"},
		{name: "target aliases source", edit: func(_ *core.Function, op *core.LinearOperation) { op.TargetID = op.SourceID }, want: "core.scalar_operation_invalid"},
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
		{name: "uninitialized U64 source", edit: func(fn *core.Function, op *core.LinearOperation) { op.SourceID = phase35OriginPlaceID(t, fn, "right") }, want: "core.scalar_analysis_invalid"},
	}
	for _, kind := range []core.OperationKind{core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64} {
		for _, row := range mutations {
			t.Run(string(kind)+"/"+row.name, func(t *testing.T) {
				mutant := phase35CloneOriginProgram(program)
				function := &mutant.Functions[0]
				row.edit(function, phase35OriginFindOperation(t, function, kind))
				problems := originvalidate.PublishProblemsFor(*function, nil)
				if !phase35OriginHasProblem(problems, row.want) {
					t.Fatalf("origin %s %s was not refused as %s: %+v", kind, row.name, row.want, problems)
				}
			})
		}
	}
	if problems := originvalidate.PublishProblemsFor(program.Functions[0], nil); len(problems) != 0 {
		t.Fatalf("clean origin field-law program failed after controls: %+v", problems)
	}

	byteProgram := phase35OriginCheckedCore(t, "module phase35.origin_byte_source\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let value = 42\n  let first = value ^ 1\n  let second = first & 1\n  let third = second | 1\n  third\n}\n")
	for _, kind := range []core.OperationKind{core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64} {
		t.Run(string(kind)+"/wrong-source-Byte", func(t *testing.T) {
			mutant := phase35CloneOriginProgram(byteProgram)
			function := &mutant.Functions[0]
			phase35OriginFindOperation(t, function, kind).SourceID = function.Parameter.ID
			if problems := originvalidate.PublishProblemsFor(*function, nil); !phase35OriginHasProblem(problems, "core.scalar_byte_parameter_used") {
				t.Fatalf("origin accepted %s Byte source: %+v", kind, problems)
			}
		})
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

func phase35OriginSetConstTarget(t *testing.T, function *core.Function, target string, value uint64) {
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
	t.Fatalf("missing origin operation defining target %q", target)
}

func phase35OriginOperationsOfKind(function *core.Function, kind core.OperationKind) []*core.LinearOperation {
	var operations []*core.LinearOperation
	for index := range function.Linear.Operations {
		if function.Linear.Operations[index].Kind == kind {
			operations = append(operations, &function.Linear.Operations[index])
		}
	}
	return operations
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

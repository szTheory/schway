package check

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/surface"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase35ShiftAdmission(t *testing.T) {
	const source = "module phase35.shift_admission\nexport { fn main }\nfn main(x: U64) -> U64 {\n  let result = x << (3)\n  result\n}\n"
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("literal-shift tracer did not parse: %+v", parsed.Diagnostics)
	}
	checked := Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("literal-shift tracer was not checked: diagnostics=%+v functions=%+v", checked.Diagnostics, checked.Program.Functions)
	}
	function := checked.Program.Functions[0]
	if function.Linear == nil {
		t.Fatalf("literal-shift tracer has no linear core: %+v", function)
	}
	var shift *core.LinearOperation
	constValues := 0
	for index := range function.Linear.Operations {
		operation := &function.Linear.Operations[index]
		if operation.Kind == core.OpConst {
			constValues++
			if operation.ConstU64 == "3" {
				t.Fatalf("literal shift amount was lowered to an ordinary constant place: %+v", *operation)
			}
		}
		if operation.Kind == core.OpShiftLeftU64 {
			if shift != nil {
				t.Fatal("literal-shift tracer emitted more than one shift")
			}
			shift = operation
		}
	}
	if shift == nil {
		t.Fatalf("grouped literal did not produce OpShiftLeftU64: %+v", function.Linear.Operations)
	}
	if shift.ID != function.ID+":op:0" || shift.PointID != function.ID+":point:linear:0" || shift.ConstU64 != "3" ||
		shift.RightID != "" || shift.SourceID != function.Parameter.ID || shift.TargetID == "" || shift.TypeID != function.Linear.Types[0].ID {
		t.Fatalf("shift core shape is not canonical: %+v", *shift)
	}
	if constValues != 0 {
		t.Fatalf("left-shift tracer produced %d OpConst operations; the literal amount must have no place", constValues)
	}
	if shift.LoanID != "" || shift.Foreign != nil || shift.OkEdgeID != "" || shift.ErrEdgeID != "" || shift.ErrTargetID != "" ||
		shift.ReleasesOperationID != "" || shift.Allocator != "" || shift.Reason != "" || shift.CalleeID != "" ||
		shift.PayloadType != "" || shift.PayloadTargetID != "" || shift.TextSource != "" || len(shift.TextBytes) != 0 ||
		shift.TextLength != 0 || shift.StoreTargetID != "" || shift.TrueEdgeID != "" || shift.FalseEdgeID != "" {
		t.Fatalf("shift populated fields outside the existing immediate contract: %+v", *shift)
	}

	refusal, ok := surface.LookupRefusal("surface.unsupported_shift_amount")
	if !ok {
		t.Fatal("surface.unsupported_shift_amount is absent from the refusal catalog")
	}
	for _, tc := range []struct {
		name   string
		amount string
		code   string
	}{
		{name: "name", amount: "amount", code: refusal.Code},
		{name: "computed group", amount: "(1 + 2)", code: refusal.Code},
		{name: "decimal 64", amount: "64", code: refusal.Code},
		{name: "hex 64", amount: "0X40", code: refusal.Code},
		{name: "maximum U64", amount: "18446744073709551615", code: refusal.Code},
		{name: "overflow", amount: "18446744073709551616", code: "type.invalid_u64"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := "let result = x << " + tc.amount
			if tc.name == "name" {
				body = "let amount = 3\n  " + body
			}
			badSource := fmt.Sprintf("module phase35.shift_bad\nexport { fn main }\nfn main(x: U64) -> U64 {\n  %s\n  result\n}\n", body)
			badParse := syntax.Parse([]byte(badSource))
			if len(badParse.Diagnostics) != 0 {
				t.Fatalf("%s source did not reach checking: %+v", tc.name, badParse.Diagnostics)
			}
			bad := Program(badParse.Program)
			if len(bad.Diagnostics) == 0 {
				t.Fatalf("SOURCE_ADMISSION_ACCEPTED: %s shift amount produced a checked function", tc.name)
			}
			if len(bad.Diagnostics) != 1 || bad.Diagnostics[0].Code != tc.code {
				t.Fatalf("SOURCE_ADMISSION_DIAGNOSTIC_MISMATCH: %s amount diagnostics=%+v, want exactly %q", tc.name, bad.Diagnostics, tc.code)
			}
			if tc.code == refusal.Code && bad.Diagnostics[0].Message != refusal.Hint {
				t.Fatalf("%s shift refusal message=%q, want catalog hint %q", tc.name, bad.Diagnostics[0].Message, refusal.Hint)
			}
			if strings.Contains(bad.Diagnostics[0].Message, "repair") || len(bad.Diagnostics[0].Repairs) != 0 {
				t.Fatalf("%s shift refusal unexpectedly offers a repair: %+v", tc.name, bad.Diagnostics[0])
			}
			if len(bad.Program.Functions) != 0 {
				t.Fatalf("%s rejected amount produced a function: %+v", tc.name, bad.Program.Functions)
			}
		})
	}

	for _, tc := range []struct {
		name   string
		amount string
		want   string
	}{
		{name: "decimal zero", amount: "0", want: "0"},
		{name: "decimal 63", amount: "63", want: "63"},
		{name: "hex 63", amount: "0X3F", want: "63"},
	} {
		t.Run("accepted/"+tc.name, func(t *testing.T) {
			acceptedSource := fmt.Sprintf("module phase35.shift_ok\nexport { fn main }\nfn main(x: U64) -> U64 {\n  let result = x >> %s\n  result\n}\n", tc.amount)
			acceptedParse := syntax.Parse([]byte(acceptedSource))
			if len(acceptedParse.Diagnostics) != 0 {
				t.Fatalf("%s source did not parse: %+v", tc.name, acceptedParse.Diagnostics)
			}
			accepted := Program(acceptedParse.Program)
			if len(accepted.Diagnostics) != 0 || len(accepted.Program.Functions) != 1 {
				t.Fatalf("%s shift was refused: %+v", tc.name, accepted.Diagnostics)
			}
			var got *core.LinearOperation
			for index := range accepted.Program.Functions[0].Linear.Operations {
				operation := &accepted.Program.Functions[0].Linear.Operations[index]
				if operation.Kind == core.OpShiftRightU64 {
					got = operation
				}
			}
			if got == nil || got.ConstU64 != tc.want || got.RightID != "" {
				t.Fatalf("%s shift immediate=%+v, want canonical %q and no RHS place", tc.name, got, tc.want)
			}
		})
	}
}

func TestPhase35NumericRoutePolicy(t *testing.T) {
	for _, tc := range []struct {
		name string
		rhs  string
		code string
	}{
		{name: "negative syntax owner", rhs: "x << -1", code: "surface.unsupported_shift_amount"},
		{name: "malformed numeric owner", rhs: "x << 0x_FF", code: "syntax.malformed_numeric_literal"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "module phase35.numeric_route\nexport { fn main }\nfn main(x: U64) -> U64 {\n  let result = " + tc.rhs + "\n  result\n}\n"
			parsed := syntax.Parse([]byte(source))
			if len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Code != tc.code {
				t.Fatalf("%s diagnostics=%+v, want the owning stage code %q", tc.name, parsed.Diagnostics, tc.code)
			}
			if tc.code == "surface.unsupported_shift_amount" {
				refusal, ok := surface.LookupRefusal(tc.code)
				if !ok || parsed.Diagnostics[0].Message != refusal.Hint {
					t.Fatalf("negative shift diagnostic=%+v, want registered refusal hint %q", parsed.Diagnostics[0], refusal.Hint)
				}
			}
			if len(Program(parsed.Program).Program.Functions) != 0 {
				t.Fatalf("%s source produced a checked function", tc.name)
			}
		})
	}
}

func TestPhase35ScalarTransferMutants(t *testing.T) {
	type vector struct {
		name      string
		kind      core.OperationKind
		peer      core.OperationKind
		left      uint64
		right     uint64
		amount    uint64
		want      uint64
		usesRight bool
		faultAt   *uint64
	}
	amountOne := uint64(1)
	vectors := []vector{
		{name: "xor high-bit pattern", kind: core.OpBitXorU64, peer: core.OpBitAndU64, left: 0xA5, right: 0x3C, want: 0x99, usesRight: true},
		{name: "and high-bits to all-zero", kind: core.OpBitAndU64, peer: core.OpBitOrU64, left: 0xF000000000000000, right: 0x0F00000000000000, want: 0, usesRight: true},
		{name: "or overlapping high-bit", kind: core.OpBitOrU64, peer: core.OpBitXorU64, left: 0x8000000000000005, right: 5, want: 0x8000000000000005, usesRight: true},
		{name: "left shift all-zero", kind: core.OpShiftLeftU64, peer: core.OpShiftRightU64, left: 0x8000000000000000, amount: 1, want: 0},
		{name: "right shift high-bit", kind: core.OpShiftRightU64, peer: core.OpShiftLeftU64, left: 0x8000000000000000, amount: 63, want: 1},
		{name: "left shift count zero", kind: core.OpShiftLeftU64, peer: core.OpShiftRightU64, left: 0x8000000000000000, amount: 0, want: 0x8000000000000000, faultAt: &amountOne},
		{name: "left shift count 31", kind: core.OpShiftLeftU64, peer: core.OpShiftRightU64, left: 1, amount: 31, want: 1 << 31},
		{name: "left shift count 32", kind: core.OpShiftLeftU64, peer: core.OpShiftRightU64, left: 1, amount: 32, want: 1 << 32},
		{name: "left shift count 63", kind: core.OpShiftLeftU64, peer: core.OpShiftRightU64, left: 1, amount: 63, want: 1 << 63},
		{name: "right shift count zero", kind: core.OpShiftRightU64, peer: core.OpShiftLeftU64, left: 0x8000000000000000, amount: 0, want: 0x8000000000000000, faultAt: &amountOne},
		{name: "right shift count 31", kind: core.OpShiftRightU64, peer: core.OpShiftLeftU64, left: 0x8000000000000000, amount: 31, want: 1 << 32},
		{name: "right shift count 32", kind: core.OpShiftRightU64, peer: core.OpShiftLeftU64, left: 0x8000000000000000, amount: 32, want: 1 << 31},
		{name: "right shift count 63", kind: core.OpShiftRightU64, peer: core.OpShiftLeftU64, left: 0x8000000000000000, amount: 63, want: 1},
	}
	for _, tc := range vectors {
		t.Run(tc.name, func(t *testing.T) {
			function, testedID, sinkID := phase35TransferSink(tc.kind, tc.left, tc.right, tc.amount, tc.want, tc.usesRight)
			_, code, category := checkScalarFixedPointDetailed(function)
			if code != "check.scalar_remainder_zero" || category != sinkID {
				t.Fatalf("checker did not derive expected %s fact into static-zero sink: code=%q category=%q", tc.name, code, category)
			}

			// Fault a copied core observation by substituting a peer operator.
			// No shared implementation hook or production state is mutated.
			faulted := function
			faulted.Linear = new(core.LinearBody)
			*faulted.Linear = *function.Linear
			faulted.Linear.Types = append([]core.TypeFact(nil), function.Linear.Types...)
			faulted.Linear.Places = append([]core.Place(nil), function.Linear.Places...)
			faulted.Linear.Blocks = append([]core.Block(nil), function.Linear.Blocks...)
			faulted.Linear.Edges = append([]core.Edge(nil), function.Linear.Edges...)
			faulted.Linear.Operations = append([]core.LinearOperation(nil), function.Linear.Operations...)
			found := false
			for index := range faulted.Linear.Operations {
				if faulted.Linear.Operations[index].ID == testedID {
					if tc.faultAt != nil {
						faulted.Linear.Operations[index].ConstU64 = strconv.FormatUint(*tc.faultAt, 10)
					} else {
						faulted.Linear.Operations[index].Kind = tc.peer
					}
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("fault control could not find tested operation %q", testedID)
			}
			_, faultCode, faultCategory := checkScalarFixedPointDetailed(faulted)
			if faultCode != "" || faultCategory != "" {
				t.Fatalf("peer transfer fault reached an unrelated checker failure instead of changing the named sink: code=%q category=%q", faultCode, faultCategory)
			}
			// A clean rerun proves the copied fault did not alter the fixture.
			if _, code, category = checkScalarFixedPointDetailed(function); code != "check.scalar_remainder_zero" || category != sinkID {
				t.Fatalf("restored baseline failed after the copy-local fault: code=%q category=%q", code, category)
			}
		})
	}
}

// phase35TransferSink builds a minimal core observation whose tested result
// is compared with an independently supplied constant, then feeds that
// equality-to-zero into the checker's static remainder-divisor safety sink.
func phase35TransferSink(kind core.OperationKind, left, right, amount, want uint64, usesRight bool) (core.Function, string, string) {
	const (
		typeID     = "phase35:type:u64"
		inputID    = "phase35:place:input"
		leftID     = "phase35:place:left"
		rightID    = "phase35:place:right"
		actualID   = "phase35:place:actual"
		wantID     = "phase35:place:want"
		zeroID     = "phase35:place:zero"
		oneID      = "phase35:place:one"
		remID      = "phase35:place:remainder"
		testedID   = "phase35:op:tested"
		sinkID     = "phase35:op:zero-sink"
		entryID    = "phase35:block:entry"
		entryPoint = "phase35:point:entry"
	)
	places := []core.Place{
		{ID: inputID, Name: "input", TypeID: typeID},
		{ID: leftID, Name: "left", TypeID: typeID},
		{ID: rightID, Name: "right", TypeID: typeID},
		{ID: actualID, Name: "actual", TypeID: typeID},
		{ID: wantID, Name: "want", TypeID: typeID},
		{ID: zeroID, Name: "zero", TypeID: typeID},
		{ID: oneID, Name: "one", TypeID: typeID},
		{ID: remID, Name: "remainder", TypeID: typeID},
	}
	constOp := func(id, target string, value uint64) core.LinearOperation {
		return core.LinearOperation{ID: id, PointID: "phase35:point:" + id, Kind: core.OpConst, TargetID: target, ConstU64: strconv.FormatUint(value, 10), TypeID: typeID}
	}
	ops := []core.LinearOperation{
		constOp("phase35:op:left", leftID, left),
	}
	if usesRight {
		ops = append(ops, constOp("phase35:op:right", rightID, right))
	}
	tested := core.LinearOperation{ID: testedID, PointID: "phase35:point:tested", Kind: kind, SourceID: leftID, TargetID: actualID, TypeID: typeID}
	if usesRight {
		tested.RightID = rightID
	} else {
		tested.ConstU64 = strconv.FormatUint(amount, 10)
	}
	ops = append(ops, tested, constOp("phase35:op:want", wantID, want),
		core.LinearOperation{ID: "phase35:op:compare", PointID: "phase35:point:compare", Kind: core.OpBitXorU64, SourceID: actualID, RightID: wantID, TargetID: zeroID, TypeID: typeID},
		constOp("phase35:op:one", oneID, 1),
		core.LinearOperation{ID: sinkID, PointID: "phase35:point:sink", Kind: core.OpRemainderU64, SourceID: oneID, RightID: zeroID, TargetID: remID, TypeID: typeID},
		core.LinearOperation{ID: "phase35:op:return", PointID: "phase35:point:return", Kind: core.OpReturn, SourceID: remID, TypeID: typeID},
	)
	operationIDs := make([]string, len(ops))
	for index := range ops {
		operationIDs[index] = ops[index].ID
	}
	function := core.Function{
		ID:            "phase35:transfer",
		Name:          "transfer",
		EntryPointID:  entryPoint,
		ReturnPointID: "phase35:point:return",
		Parameter:     core.Parameter{ID: inputID, Name: "input", Type: "U64"},
		ReturnType:    "U64",
		Linear: &core.LinearBody{
			ID:         "phase35:linear",
			Types:      []core.TypeFact{{ID: typeID, Shape: core.TypeRef{Constructor: "U64"}}},
			Places:     places,
			Operations: ops,
			Blocks:     []core.Block{{ID: entryID, PointID: entryPoint, OperationIDs: operationIDs, Successors: []string{}}},
		},
	}
	return function, testedID, sinkID
}

func TestPhase35UnusedByteSourceAdmission(t *testing.T) {
	const source = "module phase35.byte_witness\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let x = 0x100\n  let y = 0x21\n  let ready = x < y\n  let witness = x ^ ((y << 3) & 0xFF)\n  let alternate = (x ^ (y << 3)) & 0xFF\n  let combined = x | y\n  witness\n}\n"
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 || len(parsed.Program.Funcs) != 1 {
		t.Fatalf("unused-Byte witness did not parse: %+v", parsed.Diagnostics)
	}
	functionAST := parsed.Program.Funcs[0]
	if functionAST.Body.Linear == nil || len(functionAST.Body.Linear.Statements) != 7 {
		t.Fatalf("witness was not routed through structured statements: %+v", functionAST.Body.Linear)
	}
	expression := func(name string) *ast.Expr {
		t.Helper()
		for _, statement := range functionAST.Body.Linear.Statements {
			if statement.Name == name {
				return statement.Expr
			}
		}
		t.Fatalf("missing %q expression", name)
		return nil
	}
	witness, alternate := expression("witness"), expression("alternate")
	if witness == nil || witness.Kind != "bit_xor" || witness.Right == nil || witness.Right.Kind != "bit_and" || witness.Right.Left == nil || witness.Right.Left.Kind != "shift_left" {
		t.Fatalf("source witness association changed: %+v", witness)
	}
	if alternate == nil || alternate.Kind != "bit_and" || alternate.Left == nil || alternate.Left.Kind != "bit_xor" {
		t.Fatalf("alternate-mask control association changed: %+v", alternate)
	}
	checked := Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("unused Byte with U64/Bool locals was refused: diagnostics=%+v functions=%+v", checked.Diagnostics, checked.Program.Functions)
	}
	function := checked.Program.Functions[0]
	if function.Parameter.Type != "Byte" || function.Linear == nil {
		t.Fatalf("accepted bridge did not preserve the Byte formal: %+v", function)
	}
	wantTypes := []string{"Byte", "U64", "Bool", "Unit"}
	if len(function.Linear.Types) != len(wantTypes) {
		t.Fatalf("bridge type facts=%+v", function.Linear.Types)
	}
	for index, want := range wantTypes {
		if function.Linear.Types[index].Shape.Constructor != want {
			t.Fatalf("bridge type fact %d=%q, want %q", index, function.Linear.Types[index].Shape.Constructor, want)
		}
	}
	places := make(map[string]core.Place, len(function.Linear.Places))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	bytePlace := places[function.Parameter.ID]
	if function.Parameter.ID == "" || function.Linear.Types[0].ID != bytePlace.TypeID || function.Linear.Types[0].Shape.Constructor != "Byte" {
		t.Fatalf("formal does not have its dedicated Byte fact: parameter=%+v place=%+v", function.Parameter, bytePlace)
	}
	seenU64, seenBool, seenXor, seenAnd, seenOr, seenShift := false, false, false, false, false, false
	for _, place := range function.Linear.Places {
		typeName := ""
		for _, fact := range function.Linear.Types {
			if fact.ID == place.TypeID {
				typeName = fact.Shape.Constructor
			}
		}
		seenU64 = seenU64 || typeName == "U64"
		seenBool = seenBool || typeName == "Bool"
	}
	for _, operation := range function.Linear.Operations {
		for _, reference := range []string{operation.SourceID, operation.RightID, operation.TargetID, operation.StoreTargetID, operation.ErrTargetID, operation.PayloadTargetID} {
			if reference == function.Parameter.ID && reference != "" {
				t.Fatalf("operation captures the unused Byte formal: %+v", operation)
			}
			if place, exists := places[reference]; exists && reference != "" && place.TypeID == bytePlace.TypeID && place.ID != function.Parameter.ID {
				t.Fatalf("operation references a Byte-typed scalar place: op=%+v place=%+v", operation, place)
			}
		}
		switch operation.Kind {
		case core.OpBitXorU64:
			seenXor = true
		case core.OpBitAndU64:
			seenAnd = true
		case core.OpBitOrU64:
			seenOr = true
		case core.OpShiftLeftU64:
			seenShift = true
		}
	}
	if !seenU64 || !seenBool || !seenXor || !seenAnd || !seenOr || !seenShift {
		t.Fatalf("witness core is incomplete: U64=%v Bool=%v xor=%v and=%v or=%v shift=%v ops=%+v", seenU64, seenBool, seenXor, seenAnd, seenOr, seenShift, function.Linear.Operations)
	}
	byteUnit := syntax.Parse([]byte("module phase35.byte_unit\nexport { fn main }\nfn main(unused: Byte) -> Unit {\n  var seed = 0\n}\n"))
	if len(byteUnit.Diagnostics) != 0 {
		t.Fatalf("Byte/Unit signature control did not parse: %+v", byteUnit.Diagnostics)
	}
	byteUnitCheck := Program(byteUnit.Program)
	if len(byteUnitCheck.Diagnostics) != 1 || byteUnitCheck.Diagnostics[0].Code != "type.scalar_signature" || byteUnitCheck.Diagnostics[0].Message != "unused Byte scalar input requires a U64 result" {
		t.Fatalf("Byte/Unit signature was not refused by its precise owner: %+v", byteUnitCheck.Diagnostics)
	}

	for _, tc := range []struct {
		name string
		body string
	}{
		{name: "read and copy", body: "var seed = 0\n  let copied = unused\n  copied"},
		{name: "result", body: "var seed = 0\n  unused"},
		{name: "assignment target", body: "var seed = 0\n  unused = seed\n  seed"},
		{name: "builtin operand", body: "var seed = 0\n  let size = byte_len(unused)\n  size"},
		{name: "write operand", body: "var seed = 0\n  write_u64(unused)\n  seed"},
		{name: "unreachable branch", body: "var seed = 0\n  if (1 < 0) { let hidden = unused } else { let other = seed }\n  seed"},
	} {
		t.Run("refused/"+tc.name, func(t *testing.T) {
			badSource := "module phase35.byte_bad\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  " + tc.body + "\n}\n"
			badParse := syntax.Parse([]byte(badSource))
			if len(badParse.Diagnostics) != 0 {
				t.Fatalf("%s source did not reach checking: %+v", tc.name, badParse.Diagnostics)
			}
			bad := Program(badParse.Program)
			if len(bad.Diagnostics) != 1 || bad.Diagnostics[0].Code != "type.scalar_byte_parameter_used" {
				t.Fatalf("%s diagnostics=%+v, want exactly the Byte-use source refusal", tc.name, bad.Diagnostics)
			}
			if len(bad.Program.Functions) != 0 {
				t.Fatalf("%s refused bridge produced a function: %+v", tc.name, bad.Program.Functions)
			}
		})
	}

	// Synthetic AST controls cover source locations that the current parser
	// does not combine with a structured body but that the prepass must guard.
	for _, tc := range []struct {
		name string
		body *ast.LinearBody
	}{
		{name: "legacy binding source", body: &ast.LinearBody{Bindings: []ast.Binding{{RHS: ast.RHS{Source: "unused", Span: diagnostic.Span{Start: 2, End: 8}}}}}},
		{name: "call callee and argument", body: &ast.LinearBody{Bindings: []ast.Binding{{RHS: ast.RHS{Callee: "unused", Arguments: []string{"unused"}, Span: diagnostic.Span{Start: 3, End: 9}}}}}},
		{name: "nested terminal match body", body: &ast.LinearBody{TerminalMatch: &ast.MatchExpr{Scrutinee: "tag", Arms: []ast.MatchArm{{Body: &ast.LinearBody{Result: "unused", ResultSpan: diagnostic.Span{Start: 4, End: 10}}}}}}},
		{name: "implicit defect source", body: &ast.LinearBody{Statements: []ast.Statement{{Kind: "defect", Span: diagnostic.Span{Start: 5, End: 11}}}}},
	} {
		if _, found := scalarByteParameterUse(tc.body, "unused"); !found {
			t.Errorf("strict prepass missed %s", tc.name)
		}
	}

	// The statement-free Byte literal route remains the legacy owner and keeps
	// its established Const/Return operation sequence.
	legacyParsed := syntax.Parse([]byte("module phase35.byte_legacy\nexport { fn main }\nfn main(input: Byte) -> U64 {\n  let value = 7\n  value\n}\n"))
	legacy := Program(legacyParsed.Program)
	if len(legacy.Diagnostics) != 0 || len(legacy.Program.Functions) != 1 {
		t.Fatalf("legacy Byte literal route changed: diagnostics=%+v", legacy.Diagnostics)
	}
	legacyOps := legacy.Program.Functions[0].Linear.Operations
	if len(legacyOps) != 2 || legacyOps[0].Kind != core.OpConst || legacyOps[0].ConstU64 != "7" || legacyOps[1].Kind != core.OpReturn {
		t.Fatalf("legacy Byte operations changed: %+v", legacyOps)
	}
}

package check

import (
	"fmt"
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
			if len(bad.Diagnostics) != 1 || bad.Diagnostics[0].Code != tc.code {
				t.Fatalf("%s amount diagnostics=%+v, want exactly %q", tc.name, bad.Diagnostics, tc.code)
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

func TestPhase35UnusedByteSourceAdmission(t *testing.T) {
	const source = "module phase35.byte_witness\nexport { fn main }\nfn main(unused: Byte) -> U64 {\n  var seed = 0\n  let x = 0x100\n  let y = 0x21\n  let ready = x < y\n  let witness = x ^ ((y << 3) & 0xFF)\n  let alternate = (x ^ (y << 3)) & 0xFF\n  witness\n}\n"
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 || len(parsed.Program.Funcs) != 1 {
		t.Fatalf("unused-Byte witness did not parse: %+v", parsed.Diagnostics)
	}
	functionAST := parsed.Program.Funcs[0]
	if functionAST.Body.Linear == nil || len(functionAST.Body.Linear.Statements) != 6 {
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
	seenU64, seenBool, seenXor, seenAnd, seenShift := false, false, false, false, false
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
		case core.OpShiftLeftU64:
			seenShift = true
		}
	}
	if !seenU64 || !seenBool || !seenXor || !seenAnd || !seenShift {
		t.Fatalf("witness core is incomplete: U64=%v Bool=%v xor=%v and=%v shift=%v ops=%+v", seenU64, seenBool, seenXor, seenAnd, seenShift, function.Linear.Operations)
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

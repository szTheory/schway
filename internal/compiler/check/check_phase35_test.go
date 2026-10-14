package check

import (
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
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

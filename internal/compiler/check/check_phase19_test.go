package check_test

import (
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/session"
)

func TestPhase19LiteralAdmission(t *testing.T) {
	for _, literal := range []string{"42", "0x2a", "0b10_1010", "0", "18446744073709551615"} {
		t.Run(literal, func(t *testing.T) {
			source := []byte("module phase19.literal_admission\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let value = " + literal + "\n  value\n}\n")
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("valid literal %q was refused: %+v", literal, checked.Diagnostics)
			}
			if len(checked.Program.Functions) != 1 {
				t.Fatalf("checked functions=%d, want 1", len(checked.Program.Functions))
			}
			linear := checked.Program.Functions[0].Linear
			if linear == nil || len(linear.Operations) != 2 || linear.Operations[0].Kind != core.OpConst || linear.Operations[0].ConstU64 == "" {
				t.Fatalf("literal did not lower to OpConst: %+v", linear)
			}
			if literal == "0" && linear.Operations[0].ConstU64 != "0" {
				t.Fatalf("zero canonicalized to %q", linear.Operations[0].ConstU64)
			}
			if literal == "0x2a" || literal == "0b10_1010" {
				if linear.Operations[0].ConstU64 != "42" {
					t.Fatalf("%s canonicalized to %q, want 42", literal, linear.Operations[0].ConstU64)
				}
			}
		})
	}
}

func TestPhase19LiteralRange(t *testing.T) {
	const literal = "18446744073709551616"
	source := []byte("module phase19.literal_range\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let value = " + literal + "\n  value\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("U64 max+1 was admitted")
	}
	first := checked.Diagnostics[0]
	if first.Code != "check.literal_out_of_range" || string(source[first.Primary.Start:first.Primary.End]) != literal {
		t.Fatalf("overflow diagnostic=%+v, want full literal span and check.literal_out_of_range", first)
	}
	if len(checked.Program.Functions) != 0 {
		t.Fatalf("overflow produced executable functions: %+v", checked.Program.Functions)
	}
}

func TestPhase19LiteralType(t *testing.T) {
	source := []byte("module phase19.literal_type\nexport {\n  fn main\n}\nfn main(input: Byte) -> Byte {\n  let value = 42\n  value\n}\n")
	checked := session.Check(source)
	for _, problem := range checked.Diagnostics {
		if problem.Code == "type.return_mismatch" {
			return
		}
	}
	t.Fatalf("U64 literal was contextually converted to Byte; diagnostics=%+v", checked.Diagnostics)
}

func TestPhase19LiteralFrontier(t *testing.T) {
	checked := session.Check([]byte("module phase19.literal_frontier\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let count = 42\n  count\n}\n"))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("valid literal frontier remains refused: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].Linear == nil || checked.Program.Functions[0].Linear.Operations[0].Kind != core.OpConst {
		t.Fatalf("literal frontier did not reach typed OpConst: %+v", checked.Program.Functions)
	}
}

func TestPhase19NumericRefusalFrontiers(t *testing.T) {
	malformed := []byte("module phase19.malformed_frontier\nexport {\n  fn main\n}\nfn main(input: Byte) -> U64 {\n  let count = 0x_FF\n  count\n}\n")
	checked := session.Check(malformed)
	if len(checked.Diagnostics) == 0 || !strings.HasPrefix(checked.Diagnostics[0].Code, "syntax.") {
		t.Fatalf("malformed literal lost syntax refusal: %+v", checked.Diagnostics)
	}
	if checked.Diagnostics[0].Code != "syntax.malformed_numeric_literal" || string(malformed[checked.Diagnostics[0].Primary.Start:checked.Diagnostics[0].Primary.End]) != "0x_FF" {
		t.Fatalf("malformed literal diagnostic has wrong code/span: %+v", checked.Diagnostics[0])
	}
}

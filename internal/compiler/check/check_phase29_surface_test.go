package check_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase29ReservedBuiltinNames(t *testing.T) {
	for _, name := range []string{"byte_len", "byte_at"} {
		t.Run(name, func(t *testing.T) {
			source := "module phase29.reserved\n\nexport {\n  fn " + name + "\n}\n\nfn " + name + "(n: U64) -> U64 {\n  var x = n\n  x\n}\n"
			parsed := syntax.Parse([]byte(source))
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("parse: %+v", parsed.Diagnostics)
			}
			result := check.Program(parsed.Program)
			if len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "check.reserved_builtin_name" {
				t.Fatalf("diagnostics = %+v", result.Diagnostics)
			}
		})
	}
}

// A byte_len call in a generic linear body is not silently resolved as a
// user call: the let-first scalar path owns it, and a body that cannot reach
// that path is refused rather than accepted.
func TestPhase29ByteReadNotResolvedAsUserCall(t *testing.T) {
	source := "module phase29.reserved\n\nexport {\n  fn probe\n}\n\nfn probe(n: U64) -> U64 {\n  let x = byte_len(n)\n  x\n}\n"
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) == 0 {
		result := check.Program(parsed.Program)
		if len(result.Diagnostics) == 0 {
			t.Fatal("byte_len over a non-ByteView operand was accepted")
		}
	}
}

package session_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27WriteAdmission(t *testing.T) {
	source := phase27WriteSource()
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Unit write source diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].ReturnType != "Unit" {
		t.Fatalf("checked return mode = %+v, want explicit Unit", checked.Program.Functions)
	}
	function := checked.Program.Functions[0]
	var writes, newlines, unitReturns int
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpWriteU64:
			writes++
		case core.OpWriteNewline:
			newlines++
		case core.OpReturn:
			if operation.SourceID == "" && operation.TypeID == typeIDFor(t, function.Linear.Types, "Unit") {
				unitReturns++
			}
		}
	}
	if writes != 1 || newlines != 1 || unitReturns != 1 {
		t.Fatalf("typed writes/Unit return = %d/%d/%d; want 1/1/1", writes, newlines, unitReturns)
	}
	wrongOperand := []byte(strings.Replace(string(source), "write_u64(n)", "write_u64(n < 1)", 1))
	wrong := session.Check(wrongOperand)
	if !hasDiagnostic(wrong.Diagnostics, "type.scalar_write_operand") {
		t.Fatalf("Bool write_u64 operand was not rejected: %+v", wrong.Diagnostics)
	}
	if wrong.Diagnostics[0].Primary.End <= wrong.Diagnostics[0].Primary.Start {
		t.Fatalf("wrong operand diagnostic has no source span: %+v", wrong.Diagnostics[0])
	}

	wrongNewlineArity := []byte(strings.Replace(string(source), "write_newline()", "write_newline(n)", 1))
	newline := session.Check(wrongNewlineArity)
	if len(newline.Diagnostics) == 0 || newline.Diagnostics[0].Primary.End <= newline.Diagnostics[0].Primary.Start {
		t.Fatalf("write_newline argument was not rejected with a span: %+v", newline.Diagnostics)
	}

	legacy, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if old := session.Check(legacy); len(old.Diagnostics) != 0 || old.Program.Functions[0].ReturnType != "U64" {
		t.Fatalf("existing U64 app return mode changed: %+v", old.Diagnostics)
	}
}

func phase27WriteSource() []byte {
	return []byte("module phase27\nexport { fn main }\nfn main(n: U64) -> Unit {\n  if n < 1001 {\n    write_u64(n)\n    write_newline()\n  } else {\n    defect \"input exceeds 1000\"\n  }\n}\n")
}

func typeIDFor(t *testing.T, types []core.TypeFact, name string) string {
	t.Helper()
	for _, fact := range types {
		if fact.Shape.Constructor == name {
			return fact.ID
		}
	}
	t.Fatalf("type %s missing from checked facts", name)
	return ""
}

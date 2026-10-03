package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase27TracerFizzBuzz1(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "fizzbuzz.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}

	interpreted, diagnostics, err := session.RunInterpreterApplication(source, "1")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("interpreter application diagnostics=%+v err=%v", diagnostics, err)
	}
	if interpreted.Execution.Outcome.Kind != "returned" || interpreted.Execution.Outcome.Value != "" || len(interpreted.Execution.Events) != 1 || interpreted.Execution.Events[0].Kind != "function.returned" || string(interpreted.Output) != "1\n" {
		t.Fatalf("interpreter execution=%+v output=%q, want Unit and literal 1 LF", interpreted.Execution, interpreted.Output)
	}

	artifact := filepath.Join(t.TempDir(), "fizzbuzz")
	_, diagnostics, err = session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v (the tracer requires installed Clang)", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "1", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != "1\n" || stderr.Len() != 0 {
		t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v, want exit 0, literal 1 LF, empty stderr", outcome, stdout.String(), stderr.String(), err)
	}
}

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

func TestPhase27WritePeerAdmission(t *testing.T) {
	checked := session.Check(phase27WriteSource())
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Unit write source diagnostics: %+v", checked.Diagnostics)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("core peer refused valid writes: %+v", peer.Problems)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("origin peer refused Unit entry: %+v", problems)
	}

	for _, check := range []struct {
		name   string
		mutate func(*core.Program)
	}{
		{name: "wrong-write-type", mutate: func(program *core.Program) {
			boolType := typeIDFor(t, program.Functions[0].Linear.Types, "Bool")
			var boolPlace string
			for _, place := range program.Functions[0].Linear.Places {
				if place.TypeID == boolType {
					boolPlace = place.ID
					break
				}
			}
			for i := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[i]
				if op.Kind == core.OpWriteU64 {
					op.SourceID = boolPlace
				}
			}
		}},
		{name: "newline-operand", mutate: func(program *core.Program) {
			for i := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[i]
				if op.Kind == core.OpWriteNewline {
					op.SourceID = program.Functions[0].Parameter.ID
				}
			}
		}},
		{name: "forged-unit-result", mutate: func(program *core.Program) {
			for i := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[i]
				if op.Kind == core.OpReturn {
					op.SourceID = program.Functions[0].Parameter.ID
				}
			}
		}},
	} {
		t.Run(check.name, func(t *testing.T) {
			forged := checked.Program
			forged.Functions = append([]core.Function(nil), checked.Program.Functions...)
			forged.Functions[0].Linear = clonePhase27Linear(checked.Program.Functions[0].Linear)
			check.mutate(&forged)
			if peer := corevalidate.Validate(forged); peer.Valid {
				t.Fatal("core peer accepted a forged write or Unit return fact")
			}
			if problems := originvalidate.ValidatePublished(forged); len(problems) == 0 {
				t.Fatal("origin peer accepted a forged write or Unit return fact")
			}
		})
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

func clonePhase27Linear(linear *core.LinearBody) *core.LinearBody {
	if linear == nil {
		return nil
	}
	clone := *linear
	clone.Types = append([]core.TypeFact(nil), linear.Types...)
	clone.Places = append([]core.Place(nil), linear.Places...)
	clone.Operations = append([]core.LinearOperation(nil), linear.Operations...)
	clone.Blocks = append([]core.Block(nil), linear.Blocks...)
	clone.Edges = append([]core.Edge(nil), linear.Edges...)
	return &clone
}

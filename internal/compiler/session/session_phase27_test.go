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
	"github.com/szTheory/schway/internal/compiler/syntax"
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

func TestPhase27TracerBoundaries(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "fizzbuzz.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(t.TempDir(), "fizzbuzz")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	for _, test := range []struct {
		input       string
		wantOutcome string
		wantCode    int
		wantStdout  string
		wantStderr  string
	}{{"0", "returned", 0, "", ""}, {"1001", "defect", 65, "", "schway app: input exceeds 1000\n"}} {
		t.Run("input-"+test.input, func(t *testing.T) {
			interpreted, diagnostics, err := session.RunInterpreterApplication(source, test.input)
			if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != test.wantOutcome || len(interpreted.Output) != 0 {
				t.Fatalf("interpreter diagnostics=%+v execution=%+v output=%q err=%v", diagnostics, interpreted.Execution, interpreted.Output, err)
			}
			var stdout, stderr bytes.Buffer
			outcome, err := native.RunApplication(context.Background(), artifact, test.input, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != test.wantCode || stdout.String() != test.wantStdout || stderr.String() != test.wantStderr {
				t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v, want exit %d stdout=%q stderr=%q", outcome, stdout.String(), stderr.String(), err, test.wantCode, test.wantStdout, test.wantStderr)
			}
		})
	}

	sumPath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	sumSource, err := os.ReadFile(sumPath)
	if err != nil {
		t.Fatal(err)
	}
	sumResult, diagnostics, err := session.RunInterpreterApplication(sumSource, "0")
	if err != nil || len(diagnostics) != 0 || string(sumResult.Output) != "0\n" || sumResult.Execution.Outcome.Value != "0" {
		t.Fatalf("sum interpreter diagnostics=%+v execution=%+v output=%q err=%v, want returned 0 and 0 LF", diagnostics, sumResult.Execution, sumResult.Output, err)
	}
	sumArtifact := filepath.Join(t.TempDir(), "sum_to_n")
	_, diagnostics, err = session.BuildApplicationFile(context.Background(), sumPath, sumArtifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("sum BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	var sumStdout, sumStderr bytes.Buffer
	sumOutcome, err := native.RunApplication(context.Background(), sumArtifact, "0", &sumStdout, &sumStderr)
	if err != nil || sumOutcome.Kind != native.RunExited || sumOutcome.ExitCode != 0 || sumStdout.String() != "0\n" || sumStderr.Len() != 0 {
		t.Fatalf("sum native outcome=%+v stdout=%q stderr=%q err=%v, want 0 LF", sumOutcome, sumStdout.String(), sumStderr.String(), err)
	}

	u64Source := []byte("module phase27_u64_write\nexport { fn main }\nfn main(n: U64) -> U64 {\n  if n < 1001 {\n    write_u64(n)\n    write_newline()\n  } else {\n    defect \"input exceeds 1000\"\n  }\n  n\n}\n")
	u64Result, diagnostics, err := session.RunInterpreterApplication(u64Source, "2")
	if err != nil || len(diagnostics) != 0 || u64Result.Execution.Outcome.Value != "2" || string(u64Result.Output) != "2\n2\n" {
		t.Fatalf("U64 write interpreter diagnostics=%+v execution=%+v output=%q err=%v, want explicit and automatic lines", diagnostics, u64Result.Execution, u64Result.Output, err)
	}
	u64Path := filepath.Join(t.TempDir(), "u64_write.schway")
	if err := os.WriteFile(u64Path, u64Source, 0o600); err != nil {
		t.Fatal(err)
	}
	u64Artifact := filepath.Join(t.TempDir(), "u64_write")
	_, diagnostics, err = session.BuildApplicationFile(context.Background(), u64Path, u64Artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("U64 write BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	var u64Stdout, u64Stderr bytes.Buffer
	u64Outcome, err := native.RunApplication(context.Background(), u64Artifact, "2", &u64Stdout, &u64Stderr)
	if err != nil || u64Outcome.Kind != native.RunExited || u64Outcome.ExitCode != 0 || u64Stdout.String() != "2\n2\n" || u64Stderr.Len() != 0 {
		t.Fatalf("U64 write native outcome=%+v stdout=%q stderr=%q err=%v, want explicit and automatic lines", u64Outcome, u64Stdout.String(), u64Stderr.String(), err)
	}

	first := session.Format(source)
	second := session.Format(first.Canonical)
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) || !bytes.Contains(first.Canonical, []byte("// bound before output")) || !bytes.Contains(first.Canonical, []byte("// emit current value")) {
		t.Fatalf("Unit/write formatting did not preserve comments and stabilize: first=%q second=%q diagnostics=%+v/%+v", first.Canonical, second.Canonical, first.Diagnostics, second.Diagnostics)
	}
	for _, token := range syntax.Parse(first.Canonical).Tree.Tokens {
		if token.Kind == syntax.TokenIdentifier && (token.Text == "write_u64" || token.Text == "write_newline") && (token.Span.End <= token.Span.Start || string(first.Canonical[token.Span.Start:token.Span.End]) != token.Text) {
			t.Errorf("write token lost its exact source span: %+v", token)
		}
	}
	malformed := []byte(strings.Replace(string(source), "write_u64(i)", "write_u64(,)", 1))
	parsed := syntax.Parse(malformed)
	if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Primary.End <= parsed.Diagnostics[0].Primary.Start {
		t.Fatalf("malformed write did not recover with a stable span: %+v", parsed.Diagnostics)
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

func TestPhase27LiteralWriteAdmission(t *testing.T) {
	source := []byte(`module phase27_literal_write
export { fn main }
fn main(n: U64) -> Unit {
  if n < 1001 {
    write_text("Fizz")
    write_text("quote: \" slash: \\ nul: \x00 newline: \n unicode: \u03bb")
    write_text("")
  } else {
    defect "input exceeds 1000"
  }
}
`)
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("literal-only write diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].ReturnType != "Unit" {
		t.Fatalf("literal-write program = %+v, want one Unit function", checked.Program.Functions)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("core peer refused exact literal write facts: %+v", peer.Problems)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("origin peer refused exact literal write facts: %+v", problems)
	}
	wantSources := []string{`"Fizz"`, `"quote: \" slash: \\ nul: \x00 newline: \n unicode: \u03bb"`, `""`}
	wantBytes := [][]byte{
		[]byte("Fizz"),
		[]byte("quote: \" slash: \\ nul: \x00 newline: \n unicode: λ"),
		{},
	}
	var writes []core.LinearOperation
	for _, operation := range checked.Program.Functions[0].Linear.Operations {
		if operation.Kind == core.OpWriteText {
			writes = append(writes, operation)
		}
	}
	if len(writes) != len(wantSources) {
		t.Fatalf("write_text operation count = %d, want %d", len(writes), len(wantSources))
	}
	for index, operation := range writes {
		if operation.TextSource != wantSources[index] || !bytes.Equal(operation.TextBytes, wantBytes[index]) || operation.TextLength != uint64(len(wantBytes[index])) {
			t.Errorf("literal %d source=%q bytes=%v length=%d, want source=%q bytes=%v length=%d", index, operation.TextSource, operation.TextBytes, operation.TextLength, wantSources[index], wantBytes[index], len(wantBytes[index]))
		}
		if operation.SourceID != "" || operation.TargetID != "" || operation.TypeID != typeIDFor(t, checked.Program.Functions[0].Linear.Types, "Unit") {
			t.Errorf("literal %d has runtime operand/result or non-Unit type: %+v", index, operation)
		}
	}
	for _, fact := range checked.Program.Functions[0].Linear.Types {
		if fact.Shape.Constructor == "String" {
			t.Fatal("literal write introduced a runtime String type")
		}
	}
	for _, mutation := range []struct {
		name   string
		mutate func(*testing.T, *core.Program)
	}{
		{name: "length-mismatch", mutate: func(t *testing.T, program *core.Program) {
			t.Helper()
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpWriteText {
					op.TextLength++
					return
				}
			}
			t.Fatal("write_text operation missing")
		}},
		{name: "payload-mismatch", mutate: func(t *testing.T, program *core.Program) {
			t.Helper()
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpWriteText && len(op.TextBytes) > 0 {
					op.TextBytes[0] ^= 1
					return
				}
			}
			t.Fatal("non-empty write_text operation missing")
		}},
		{name: "nonliteral-provenance", mutate: func(t *testing.T, program *core.Program) {
			t.Helper()
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpWriteText {
					op.TextSource = "n"
					return
				}
			}
			t.Fatal("write_text operation missing")
		}},
		{name: "wrong-operation-kind", mutate: func(t *testing.T, program *core.Program) {
			t.Helper()
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpWriteText {
					op.Kind = core.OpWriteU64
					return
				}
			}
			t.Fatal("write_text operation missing")
		}},
		{name: "embedded-zero-length", mutate: func(t *testing.T, program *core.Program) {
			t.Helper()
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpWriteText && bytes.IndexByte(op.TextBytes, 0) >= 0 {
					op.TextLength--
					return
				}
			}
			t.Fatal("embedded-zero write_text operation missing")
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			forged := checked.Program
			forged.Functions = append([]core.Function(nil), checked.Program.Functions...)
			forged.Functions[0].Linear = clonePhase27Linear(checked.Program.Functions[0].Linear)
			mutation.mutate(t, &forged)
			if peer := corevalidate.Validate(forged); peer.Valid {
				t.Fatal("core peer accepted forged literal bytes, length, or provenance")
			}
			if problems := originvalidate.ValidatePublished(forged); len(problems) == 0 {
				t.Fatal("origin peer accepted forged literal bytes, length, or provenance")
			}
		})
	}

	dynamic := []byte(strings.Replace(string(source), `write_text("Fizz")`, `write_text(n)`, 1))
	dynamicResult := session.Check(dynamic)
	if !hasDiagnostic(dynamicResult.Diagnostics, "type.scalar_write_literal") {
		t.Fatalf("dynamic text operand was not refused: %+v", dynamicResult.Diagnostics)
	}
	dynamicStart := strings.Index(string(dynamic), "write_text(n)") + len("write_text(")
	if dynamicResult.Diagnostics[0].Primary.Start != dynamicStart || dynamicResult.Diagnostics[0].Primary.End != dynamicStart+1 {
		t.Fatalf("dynamic operand diagnostic span=%+v, want only %q", dynamicResult.Diagnostics[0].Primary, "n")
	}
	compound := []byte(strings.Replace(string(source), `write_text("Fizz")`, `write_text("F" + "izz")`, 1))
	compoundResult := session.Check(compound)
	if !hasDiagnostic(compoundResult.Diagnostics, "type.scalar_write_literal") {
		t.Fatalf("general text expression was not refused: %+v", compoundResult.Diagnostics)
	}
	compoundStart := strings.Index(string(compound), `"F" + "izz"`)
	if compoundResult.Diagnostics[0].Primary.Start != compoundStart || compoundResult.Diagnostics[0].Primary.End != compoundStart+len(`"F" + "izz"`) {
		t.Fatalf("compound text diagnostic span=%+v, want full expression at %d", compoundResult.Diagnostics[0].Primary, compoundStart)
	}

	malformed := []byte("module phase27_malformed_text\nexport { fn main }\nfn main(n: U64) -> Unit {\n  write_text(\"\\q\")\n}\n")
	malformedResult := session.Check(malformed)
	if !hasDiagnostic(malformedResult.Diagnostics, "type.scalar_write_literal") {
		t.Fatalf("malformed escape was not refused: %+v", malformedResult.Diagnostics)
	}
	malformedStart := strings.Index(string(malformed), `"\q"`)
	if malformedResult.Diagnostics[0].Primary.Start != malformedStart || malformedResult.Diagnostics[0].Primary.End != malformedStart+4 {
		t.Fatalf("malformed escape diagnostic span=%+v, want full quoted token at %d", malformedResult.Diagnostics[0].Primary, malformedStart)
	}
}

func TestPhase27LiteralEscapeFormat(t *testing.T) {
	source := []byte(`module phase27_literal_format
export { fn main }
fn main(n: U64) -> Unit {
  // This comment remains attached to the literal write.
  write_text( "a\"b\\c\x00\n\u03bb" ) // trailing comment stays visible
}
`)
	first := session.Format(source)
	second := session.Format(first.Canonical)
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) {
		t.Fatalf("literal formatting did not stabilize: first=%q second=%q diagnostics=%+v/%+v", first.Canonical, second.Canonical, first.Diagnostics, second.Diagnostics)
	}
	if !bytes.Contains(first.Canonical, []byte(`write_text("a\"b\\c\x00\n\u03bb")`)) || !bytes.Contains(first.Canonical, []byte("// This comment remains attached to the literal write.")) || !bytes.Contains(first.Canonical, []byte("// trailing comment stays visible")) || first.Canonical[len(first.Canonical)-1] != '\n' {
		t.Fatalf("canonical source changed quoted spelling, comments, or final LF: %q", first.Canonical)
	}
	for _, token := range syntax.Parse(source).Tree.Tokens {
		if token.Kind == syntax.TokenString && (token.Span.End <= token.Span.Start || string(source[token.Span.Start:token.Span.End]) != token.Text) {
			t.Errorf("quoted token lost exact source span: %+v", token)
		}
	}

	malformedEscape := []byte("module phase27_format_recovery\nexport { fn main }\nfn main(n: U64) -> Unit {\n  write_text(\"\\q\")\n}\n")
	recovered := session.Format(malformedEscape)
	if !bytes.Contains(recovered.Canonical, []byte(`write_text("\q")`)) || len(recovered.Diagnostics) != 0 {
		t.Fatalf("formatter did not preserve malformed quoted token for checker refusal: canonical=%q diagnostics=%+v", recovered.Canonical, recovered.Diagnostics)
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
	for index := range clone.Operations {
		clone.Operations[index].TextBytes = append([]byte(nil), linear.Operations[index].TextBytes...)
	}
	clone.Blocks = append([]core.Block(nil), linear.Blocks...)
	clone.Edges = append([]core.Edge(nil), linear.Edges...)
	return &clone
}

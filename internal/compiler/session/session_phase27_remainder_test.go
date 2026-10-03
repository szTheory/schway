package session_test

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase27OperatorSyntax(t *testing.T) {
	source := []byte("module phase27_operators\nexport { fn main }\nfn main(n: U64) -> Unit {\n  // remainder binds more tightly than equality\n  if n % 5 == 0 {\n    write_u64(n)\n  } else {\n    write_newline()\n  }\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("operator source diagnostics: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].ReturnType != "Unit" {
		t.Fatalf("checked program functions: %+v", checked.Program.Functions)
	}
	function := checked.Program.Functions[0]
	places := make(map[string]core.Place, len(function.Linear.Places))
	types := make(map[string]string, len(function.Linear.Types))
	for _, place := range function.Linear.Places {
		places[place.ID] = place
	}
	for _, fact := range function.Linear.Types {
		types[fact.ID] = fact.Shape.Constructor
	}
	var equal, remainder *core.LinearOperation
	for index := range function.Linear.Operations {
		op := &function.Linear.Operations[index]
		switch op.Kind {
		case core.OpEqualU64:
			equal = op
		case core.OpRemainderU64:
			remainder = op
		}
	}
	if equal == nil || remainder == nil {
		t.Fatalf("typed equality/remainder operations missing: %+v", function.Linear.Operations)
	}
	if types[places[equal.TargetID].TypeID] != "Bool" || types[places[remainder.TargetID].TypeID] != "U64" || equal.SourceID != remainder.TargetID {
		t.Fatalf("operator result types or precedence are wrong: equal=%+v remainder=%+v", equal, remainder)
	}

	first := session.Format(source)
	second := session.Format(first.Canonical)
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) || !bytes.Contains(first.Canonical, []byte("// remainder binds more tightly than equality")) || !bytes.Contains(first.Canonical, []byte("n % 5 == 0")) {
		t.Fatalf("operator formatting changed or lost source: first=%q second=%q diagnostics=%+v/%+v", first.Canonical, second.Canonical, first.Diagnostics, second.Diagnostics)
	}
	for _, token := range syntax.Parse(first.Canonical).Tree.Tokens {
		if token.Kind != syntax.TokenPercent && token.Kind != syntax.TokenEqualEqual {
			continue
		}
		if token.Span.End <= token.Span.Start || string(first.Canonical[token.Span.Start:token.Span.End]) != token.Text {
			t.Errorf("operator %q lost its source span: %+v", token.Text, token.Span)
		}
	}

	wrongType := []byte("module phase27_wrong_operator_type\nexport { fn main }\nfn main(n: U64) -> U64 {\n  var flag = n < 3\n  var bad = flag % 2\n  n\n}\n")
	if wrong := session.Check(wrongType); len(wrong.Diagnostics) == 0 || wrong.Diagnostics[0].Code != "type.scalar_operand" {
		t.Fatalf("Bool remainder operand was not rejected: %+v", wrong.Diagnostics)
	}
	malformed := []byte(strings.Replace(string(source), "n % 5 == 0", "n % )", 1))
	parsed := syntax.Parse(malformed)
	if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Primary.End <= parsed.Diagnostics[0].Primary.Start {
		t.Fatalf("malformed operator expression did not recover with a span: %+v", parsed.Diagnostics)
	}
}

func TestPhase27StaticZero(t *testing.T) {
	source := []byte("module phase27_static_zero\nexport { fn main }\nfn main(n: U64) -> U64 {\n  let rem = n % 0\n  n\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != "type.scalar_remainder_zero" {
		t.Fatalf("literal zero remainder divisor was not rejected: %+v", checked.Diagnostics)
	}
	zeroOffset := strings.Index(string(source), "n % 0") + len("n % ")
	if checked.Diagnostics[0].Primary.Start != zeroOffset || checked.Diagnostics[0].Primary.End != zeroOffset+1 {
		t.Fatalf("literal zero diagnostic span=%+v, want divisor byte [%d,%d)", checked.Diagnostics[0].Primary, zeroOffset, zeroOffset+1)
	}
}

func TestPhase27ScalarZero(t *testing.T) {
	knownZero := []byte("module phase27_scalar_zero\nexport { fn main }\nfn main(n: U64) -> U64 {\n  var zero = 0\n  let copied_zero = zero\n  let rem = n % copied_zero\n  n\n}\n")
	checked := session.Check(knownZero)
	if len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != "type.scalar_remainder_zero" {
		t.Fatalf("copied known-zero remainder divisor was not rejected: %+v", checked.Diagnostics)
	}
	divisorOffset := strings.Index(string(knownZero), "n % copied_zero") + len("n % ")
	if checked.Diagnostics[0].Primary.Start != divisorOffset || checked.Diagnostics[0].Primary.End != divisorOffset+len("copied_zero") {
		t.Fatalf("copied zero diagnostic span=%+v, want divisor [%d,%d)", checked.Diagnostics[0].Primary, divisorOffset, divisorOffset+len("copied_zero"))
	}

	unknown := []byte("module phase27_unknown_zero\nexport { fn main }\nfn main(n: U64) -> U64 {\n  var divisor = n\n  let rem = n % divisor\n  rem\n}\n")
	admitted := session.Check(unknown)
	if len(admitted.Diagnostics) != 0 {
		t.Fatalf("unknown divisor was not admitted for guarded execution: %+v", admitted.Diagnostics)
	}
	var remainders int
	for _, operation := range admitted.Program.Functions[0].Linear.Operations {
		if operation.Kind == core.OpRemainderU64 {
			remainders++
		}
	}
	if remainders != 1 {
		t.Fatalf("unknown-divisor core contains %d remainder operations, want one", remainders)
	}
}

func TestPhase27EqualityRemainder(t *testing.T) {
	source := []byte("module phase27_remainder_runtime\nexport { fn main }\nfn main(n: U64) -> U64 {\n  let dividend = 17\n  let remainder = dividend % n\n  var result = 0\n  if remainder == 2 {\n    result = remainder\n  } else {\n    result = 0\n  }\n  result\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("dynamic remainder source diagnostics: %+v", checked.Diagnostics)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("core peer refused valid equality/remainder operations: %+v", peer.Problems)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("origin peer refused valid equality/remainder operations: %+v", problems)
	}

	artifact := filepath.Join(t.TempDir(), "phase27-remainder")
	if _, diagnostics, err := session.BuildApplication(context.Background(), source, artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplication diagnostics=%+v err=%v", diagnostics, err)
	}
	for _, test := range []struct{ input, want string }{{"5", "2"}, {"6", "0"}} {
		t.Run("input-"+test.input, func(t *testing.T) {
			interpreted, diagnostics, err := session.RunInterpreterApplication(source, test.input)
			if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != "returned" || interpreted.Execution.Outcome.Value != test.want || string(interpreted.Output) != test.want+"\n" {
				t.Fatalf("interpreter diagnostics=%+v outcome=%+v output=%q err=%v, want %s", diagnostics, interpreted.Execution.Outcome, interpreted.Output, err, test.want)
			}
			var stdout, stderr bytes.Buffer
			outcome, err := native.RunApplication(context.Background(), artifact, test.input, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != test.want+"\n" || stderr.Len() != 0 {
				t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v, want %q", outcome, stdout.String(), stderr.String(), err, test.want+"\n")
			}
		})
	}
}

func TestPhase27ForgedRemainder(t *testing.T) {
	source := []byte("module phase27_forged_remainder\nexport { fn main }\nfn main(n: U64) -> U64 {\n  let divisor = 5\n  let remainder = n % divisor\n  var result = 0\n  if remainder == 2 {\n    result = remainder\n  } else {\n    result = 0\n  }\n  result\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source diagnostics: %+v", checked.Diagnostics)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("core peer refused valid source before mutations: %+v", peer.Problems)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("origin peer refused valid source before mutations: %+v", problems)
	}

	for _, mutation := range []struct {
		name  string
		apply func(*testing.T, *core.Program)
	}{
		{name: "bool-operand", apply: func(t *testing.T, program *core.Program) {
			t.Helper()
			boolType := typeIDFor(t, program.Functions[0].Linear.Types, "Bool")
			boolPlace := ""
			for _, place := range program.Functions[0].Linear.Places {
				if place.TypeID == boolType {
					boolPlace = place.ID
					break
				}
			}
			if boolPlace == "" {
				t.Fatal("source has no Bool place for forged operand")
			}
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpRemainderU64 {
					op.SourceID = boolPlace
					return
				}
			}
			t.Fatal("remainder operation missing")
		}},
		{name: "bool-result", apply: func(t *testing.T, program *core.Program) {
			t.Helper()
			boolType := typeIDFor(t, program.Functions[0].Linear.Types, "Bool")
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpRemainderU64 {
					op.TypeID = boolType
					return
				}
			}
			t.Fatal("remainder operation missing")
		}},
		{name: "known-zero-divisor", apply: func(t *testing.T, program *core.Program) {
			t.Helper()
			for index := range program.Functions[0].Linear.Operations {
				op := &program.Functions[0].Linear.Operations[index]
				if op.Kind == core.OpConst && op.ConstU64 == "5" {
					op.ConstU64 = "0"
					return
				}
			}
			t.Fatal("known nonzero divisor constant missing")
		}},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			forged := checked.Program
			forged.Functions = append([]core.Function(nil), checked.Program.Functions...)
			forged.Functions[0].Linear = clonePhase27Linear(checked.Program.Functions[0].Linear)
			mutation.apply(t, &forged)
			if peer := corevalidate.Validate(forged); peer.Valid {
				t.Fatal("core peer accepted forged remainder operand/result facts")
			}
			if problems := originvalidate.ValidatePublished(forged); len(problems) == 0 {
				t.Fatal("origin peer accepted forged remainder operand/result facts")
			}
		})
	}
}

func TestPhase27DynamicZero(t *testing.T) {
	source := []byte("module phase27_dynamic_zero\nexport { fn main }\nfn main(divisor: U64) -> U64 {\n  let dividend = 17\n  var result = 0\n  if divisor == 0 {\n    let remainder = dividend % divisor\n    result = remainder\n  } else {\n    result = 0\n  }\n  result\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("dynamic zero witness diagnostics: %+v", checked.Diagnostics)
	}
	cSource, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("EmitApplication: %v", err)
	}
	assertZeroGuardOrder := func(t *testing.T, generated string) {
		t.Helper()
		moduloIndex := strings.LastIndex(generated, " % ")
		prefixEnd := len(generated)
		if moduloIndex >= 0 {
			prefixEnd = moduloIndex
		}
		zeroCompareIndex := strings.LastIndex(generated[:prefixEnd], "== UINT64_C(0)")
		guardIndex := -1
		if zeroCompareIndex >= 0 {
			guardIndex = strings.LastIndex(generated[:zeroCompareIndex], "if (")
		}
		reasonIndex := strings.Index(generated, "U64 remainder by zero")
		if moduloIndex < 0 || zeroCompareIndex < 0 || guardIndex < 0 || reasonIndex < guardIndex || reasonIndex >= moduloIndex {
			t.Fatalf("generated remainder does not guard the divisor before C %%: modulo=%d guard=%d reason=%d\n%s", moduloIndex, guardIndex, reasonIndex, generated)
		}
	}
	assertZeroGuardOrder(t, cSource)
	if !strings.Contains(cSource, `fputs("schway: U64 remainder by zero\n", stderr);`) || !strings.Contains(cSource, "exit(65);") {
		t.Fatalf("generated application lacks the public zero-divisor failure: %s", cSource)
	}
	executionC, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative execution witness: %v", err)
	}
	assertZeroGuardOrder(t, executionC)

	conformanceInterpreter, err := interp.Run(checked.Program, "main", "0")
	if err != nil || conformanceInterpreter.Outcome.Kind != execution.OutcomeDefect || len(conformanceInterpreter.Events) == 0 || conformanceInterpreter.Events[len(conformanceInterpreter.Events)-1].Output != "U64 remainder by zero" {
		t.Fatalf("conformance interpreter outcome=%+v events=%+v err=%v", conformanceInterpreter.Outcome, conformanceInterpreter.Events, err)
	}
	projected, err := session.ProjectExecutionSchema2(checked.Program, conformanceInterpreter)
	if err != nil || executionpeer.Validate(checked.Program, projected) != nil {
		t.Fatalf("projected interpreter remainder defect=%+v err=%v", projected, err)
	}
	conformanceRunner := native.DefaultRunner()
	conformanceRunner.Expect = native.ExpectDefect
	conformanceNative, err := conformanceRunner.Run(context.Background(), executionC, "-O0", []string{"0"})
	if err != nil || len(conformanceNative.Pairs) != 1 {
		t.Fatalf("native conformance remainder defect pairs=%d err=%v", len(conformanceNative.Pairs), err)
	}
	if err := executionpeer.Validate(checked.Program, conformanceNative.Pairs[0].Execution); err != nil {
		t.Fatalf("execution peer rejected native remainder defect: %v", err)
	}
	if !execution.Equal(projected, conformanceNative.Pairs[0].Execution) {
		t.Fatalf("interpreter/native conformance remainder documents differ: interpreter=%+v native=%+v", projected, conformanceNative.Pairs[0].Execution)
	}

	interpreted, diagnostics, err := session.RunInterpreterApplication(source, "0")
	if err != nil || len(diagnostics) != 0 || interpreted.Execution.Outcome.Kind != "defect" || len(interpreted.Output) != 0 {
		t.Fatalf("interpreter diagnostics=%+v outcome=%+v output=%q err=%v, want pre-output defect", diagnostics, interpreted.Execution.Outcome, interpreted.Output, err)
	}
	foundReason := false
	for _, event := range interpreted.Execution.Events {
		if event.Kind == "function.defected" && event.Output == "U64 remainder by zero" {
			foundReason = true
		}
	}
	if !foundReason {
		t.Fatalf("interpreter execution has no exact remainder-zero reason: %+v", interpreted.Execution.Events)
	}

	artifact := filepath.Join(t.TempDir(), "phase27-dynamic-zero")
	if _, diagnostics, err := session.BuildApplication(context.Background(), source, artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplication diagnostics=%+v err=%v", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "0", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway: U64 remainder by zero\n" {
		t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v, want exit 65, empty stdout, exact diagnostic", outcome, stdout.String(), stderr.String(), err)
	}
}

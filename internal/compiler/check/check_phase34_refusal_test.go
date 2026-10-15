package check_test

import (
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func phase34MergedProgram(t *testing.T, entry, library []byte) ast.Program {
	t.Helper()
	parse := func(source []byte) ast.Program {
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("module fixture failed to parse: %+v", parsed.Diagnostics)
		}
		return parsed.Program
	}
	merged := parse(entry)
	lib := parse(library)
	merged.Units = []ast.Unit{{Name: "", Imports: merged.Imports, Exports: merged.Exports}}
	for _, function := range lib.Funcs {
		function.Module = lib.Module
		merged.Funcs = append(merged.Funcs, function)
	}
	merged.Units = append(merged.Units, ast.Unit{Name: lib.Module, Imports: lib.Imports, Exports: lib.Exports})
	return merged
}

func TestPhase34UnsupportedSurfaceRefusals(t *testing.T) {
	for _, test := range []struct {
		name   string
		source string
		form   string
	}{
		{
			name:   "unbound-true-linear-tail",
			source: "module phase34.unbound_true\nexport { fn main }\nfn main(n: U64) -> U64 { true }\n",
			form:   "true",
		},
		{
			name:   "unbound-false-linear-tail",
			source: "module phase34.unbound_false\nexport { fn main }\nfn main(n: U64) -> U64 { false }\n",
			form:   "false",
		},
		{
			name:   "unbound-true-scalar-tail",
			source: "module phase34.unbound_true_scalar\nexport { fn main }\nfn main(n: U64) -> U64 { var result = n true }\n",
			form:   "true",
		},
		{
			name:   "unbound-bool-match-arm-tail",
			source: "module phase34.unbound_arm\nexport { fn main }\ndata Choice = | One\ndata Result = | Yes\nfn main(choice: Choice) -> Result { match choice { One => { true } } }\n",
			form:   "true",
		},
	} {
		source := []byte(test.source)
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("%s fixture must reach checking, parser diagnostics=%+v", test.name, parsed.Diagnostics)
		}
		result := session.Check(source)
		start := strings.LastIndex(test.source, test.form)
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "surface.unsupported_bool_literal_tail" ||
			result.Diagnostics[0].Primary.Start != start || result.Diagnostics[0].Primary.End != start+len(test.form) {
			t.Errorf("%s result = %+v, want the exact unbound Bool-tail refusal at [%d:%d]", test.name, result.Diagnostics, start, start+len(test.form))
		}
		if len(result.Program.Functions) != 0 {
			t.Errorf("%s admitted a rejected Bool-looking tail: %+v", test.name, result.Program.Functions)
		}
	}
}

func TestPhase34BoolSpellingsRemainIdentifiers(t *testing.T) {
	for _, test := range []struct {
		name       string
		source     string
		functionID string
	}{
		{"true-parameter", "module phase34.true_parameter\nexport { fn main }\nfn main(true: U64) -> U64 { true }\n", "s1:phase34.true_parameter:fn:main"},
		{"false-parameter", "module phase34.false_parameter\nexport { fn main }\nfn main(false: U64) -> U64 { false }\n", "s1:phase34.false_parameter:fn:main"},
		{"true-scalar-parameter", "module phase34.true_scalar\nexport { fn main }\nfn main(true: U64) -> U64 { let result = true + 1 result }\n", "s1:phase34.true_scalar:fn:main"},
		{"false-scalar-binding", "module phase34.false_scalar\nexport { fn main }\nfn main(n: U64) -> U64 { let false = n + 1 false }\n", "s1:phase34.false_scalar:fn:main"},
		{"true-binding", "module phase34.true_binding\nexport { fn main }\nfn main(n: U64) -> U64 { let true = n true }\n", "s1:phase34.true_binding:fn:main"},
		{"false-binding", "module phase34.false_binding\nexport { fn main }\nfn main(n: U64) -> U64 { let false = n false }\n", "s1:phase34.false_binding:fn:main"},
		{"mixed-bindings", "module phase34.mixed_bindings\nexport { fn main }\nfn main(false: U64) -> U64 { let true = false true }\n", "s1:phase34.mixed_bindings:fn:main"},
	} {
		result := session.Check([]byte(test.source))
		if len(result.Diagnostics) != 0 {
			t.Errorf("%s changed its pre-existing checked meaning: %+v", test.name, result.Diagnostics)
		}
		if len(result.Program.Functions) != 1 || result.Program.Functions[0].ID != test.functionID {
			t.Errorf("%s changed core function identity: %+v, want %q", test.name, result.Program.Functions, test.functionID)
		}
	}

	entry := []byte("module phase34.importer\nimport phase34.bool_names { fn true }\nexport { fn main }\nfn main(n: U64) -> U64 { let result = true(n) result }\n")
	library := []byte("module phase34.bool_names\nexport { fn true }\nfn true(value: U64) -> U64 { value }\n")
	merged := phase34MergedProgram(t, entry, library)
	imported := check.Program(merged)
	if len(imported.Diagnostics) != 0 {
		t.Fatalf("imported function named true changed its call meaning: %+v", imported.Diagnostics)
	}
	if len(imported.Program.Functions) != 2 {
		t.Fatalf("imported function named true changed core function count: %+v", imported.Program.Functions)
	}
	if imported.Program.Functions[0].ID != "s1:phase34.importer:fn:main" || imported.Program.Functions[1].ID != "s1:phase34.bool_names:fn:true" {
		t.Fatalf("imported true changed semantic IDs: %+v", imported.Program.Functions)
	}

	wrongType := session.Check([]byte("module phase34.bound_wrong_type\nexport { fn main }\nfn main(n: U64) -> U64 { var true = n < 2 true }\n"))
	if len(wrongType.Diagnostics) != 1 || wrongType.Diagnostics[0].Code != "type.scalar_result" {
		t.Errorf("bound true with the wrong result type = %+v, want type.scalar_result", wrongType.Diagnostics)
	}
	uninitialized := session.Check([]byte("module phase34.bound_moved\nexport { fn main }\nfn main(true: Buffer) -> Buffer { let moved = take true true }\n"))
	if len(uninitialized.Diagnostics) != 1 || uninitialized.Diagnostics[0].Code != "ownership.use_after_move" {
		t.Errorf("bound moved true result = %+v, want ownership.use_after_move", uninitialized.Diagnostics)
	}
	unknown := session.Check([]byte("module phase34.other_unknown\nexport { fn main }\nfn main(n: U64) -> U64 { missing }\n"))
	if len(unknown.Diagnostics) != 1 || unknown.Diagnostics[0].Code != "name.unknown" {
		t.Errorf("unrelated unknown result = %+v, want name.unknown", unknown.Diagnostics)
	}

	if len(wrongType.Program.Functions) != 0 || len(uninitialized.Program.Functions) != 0 || len(unknown.Program.Functions) != 0 {
		t.Errorf("a rejected bound/unknown-name control reached core: wrong=%+v moved=%+v unknown=%+v", wrongType.Program, uninitialized.Program, unknown.Program)
	}
}

func TestPhase34CallArityBoundaries(t *testing.T) {
	call := session.Check([]byte("module phase34.call_arity\nexport { fn main }\nfn identity(value: U64) -> U64 { value }\nfn main(n: U64) -> U64 { let result = identity(n, n) result }\n"))
	if len(call.Diagnostics) != 1 || call.Diagnostics[0].Code != "check.call_arity_unsupported" {
		t.Fatalf("ordinary two-argument call = %+v, want check.call_arity_unsupported", call.Diagnostics)
	}
	if call.Diagnostics[0].Code == "surface.unsupported_parameter_arity" {
		t.Fatal("call-site arity was classified as declaration arity")
	}

	byteAt := session.Check([]byte("module phase34.byte_at\nexport { fn read fn main }\nfn read(bytes: ByteView) -> U64 { let result = byte_at(bytes, 0) result }\nfn main(n: U64) -> U64 { n }\n"))
	if len(byteAt.Diagnostics) != 0 || len(byteAt.Program.Functions) != 2 {
		t.Fatalf("byte_at(view, index) lost its admitted builtin arity: diagnostics=%+v program=%+v", byteAt.Diagnostics, byteAt.Program)
	}
	var byteAtBody *core.LinearBody
	for _, function := range byteAt.Program.Functions {
		if function.Name == "read" {
			byteAtBody = function.Linear
		}
	}
	if byteAtBody == nil {
		t.Fatalf("byte_at reader has no checked scalar body: %+v", byteAt.Program.Functions)
	}
	foundByteAt := false
	for _, operation := range byteAtBody.Operations {
		if operation.Kind == core.OpBytesAt {
			foundByteAt = true
		}
	}
	if !foundByteAt {
		t.Fatalf("byte_at(view, index) did not retain its core byte-read operation: %+v", byteAtBody)
	}

	callExpression := session.Check([]byte("module phase34.scalar_call_expression\nexport { fn main }\nfn identity(value: U64) -> U64 { value }\nfn main(n: U64) -> U64 { let result = identity(n) + 1 result }\n"))
	if len(callExpression.Diagnostics) != 0 {
		t.Fatalf("one-argument helper expression was refused: %+v", callExpression.Diagnostics)
	}
	var callCount int
	for _, function := range callExpression.Program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCall {
				callCount++
			}
		}
	}
	if callCount != 1 {
		t.Fatalf("one-argument helper expression produced %d OpCall operations, want one", callCount)
	}
}

func TestPhase34RefusalHintsHaveNoRepairs(t *testing.T) {
	for _, test := range []struct {
		name       string
		source     string
		form       string
		spanOffset int
		spanLength int
		code       string
		hint       string
	}{
		{name: "logical-and", source: "module phase34.and\nexport { fn main }\nfn main(n: U64) -> U64 { var result = n if n < 2 && n < 3 { result = n } else { result = n } result }\n", form: "&&", code: "surface.unsupported_logical_and", hint: "Logical && is unsupported; use explicit if/else control flow."},
		{name: "logical-or", source: "module phase34.or\nexport { fn main }\nfn main(flag: Bool) -> Bool { flag || false }\n", form: "||", code: "surface.unsupported_logical_or", hint: "Logical || is unsupported; use explicit if/else control flow."},
		{name: "logical-not", source: "module phase34.not\nexport { fn main }\nfn main(flag: Bool) -> Bool { !flag }\n", form: "!", code: "surface.unsupported_logical_not", hint: "Logical ! is unsupported; use explicit if/else control flow."},
		{name: "parameter-arity", source: "module phase34.arity\nexport { fn main }\nfn pair(left: U64, right: U64) -> U64 { left }\nfn main(n: U64) -> U64 { n }\n", form: "right: U64", code: "surface.unsupported_parameter_arity", hint: "Declarations require exactly one parameter; pass one admitted value."},
		{name: "if-without-else", source: "module phase34.if_no_else\nexport { fn main }\nfn main(n: U64) -> U64 { var result = n if n < 1 { result = n } result }\n", form: "if n < 1 { result = n }", code: "surface.unsupported_if_without_else", hint: "An if statement requires an else block."},
		{name: "bool-tail", source: "module phase34.bool_tail\nexport { fn main }\nfn main(n: U64) -> U64 { true }\n", form: "true", code: "surface.unsupported_bool_literal_tail", hint: "Bool literals are unsupported in result tails; return an initialized value allowed by the declared result type."},
		{name: "subtraction", source: "module phase34.subtraction\nexport { fn main }\nfn main(n: U64) -> U64 { n - 1 }\n", form: "n - 1", spanOffset: 2, spanLength: 1, code: "surface.unsupported_subtraction", hint: "Subtraction is unsupported; no unchecked emulation is recommended."},
	} {
		result := session.Check([]byte(test.source))
		start := strings.LastIndex(test.source, test.form)
		if start >= 0 {
			start += test.spanOffset
		}
		spanLength := test.spanLength
		if spanLength == 0 {
			spanLength = len(test.form)
		}
		if len(result.Diagnostics) != 1 {
			t.Errorf("%s diagnostics = %+v, want exactly one refusal", test.name, result.Diagnostics)
			continue
		}
		problem := result.Diagnostics[0]
		if problem.Code != test.code || problem.Message != test.hint || len(problem.Repairs) != 0 {
			t.Errorf("%s diagnostic = %+v, want independent code %q, hint %q, and no repairs", test.name, problem, test.code, test.hint)
		}
		if problem.Primary.Start != start || problem.Primary.End != start+spanLength {
			t.Errorf("%s span = %+v, want [%d:%d]", test.name, problem.Primary, start, start+spanLength)
		}
	}
}

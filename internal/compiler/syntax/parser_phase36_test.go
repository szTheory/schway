package syntax_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase36ScalarCallGrammar(t *testing.T) {
	source := []byte("module phase36.calls\nexport {\n  fn evaluate\n}\nfn evaluate(input: U64) -> U64 {\n  let direct = helper(input + 3)\n  let nested = helper(other(input))\n  let contextual = helper(true)\n  write_u64(helper(input) + helper(other(input)))\n  direct\n}\n")
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("well-formed named scalar calls produced diagnostics: %+v functions=%+v", parsed.Diagnostics, parsed.Program.Funcs)
	}
	if !bytes.Equal(parsed.Tree.Bytes(), source) {
		t.Fatalf("lossless parse changed call source bytes: %q", parsed.Tree.Bytes())
	}
	if len(parsed.Program.Funcs) != 1 || parsed.Program.Funcs[0].Body.Linear == nil {
		t.Fatalf("call program did not produce one scalar function: %+v", parsed.Program.Funcs)
	}
	statements := parsed.Program.Funcs[0].Body.Linear.Statements
	if len(statements) != 4 {
		t.Fatalf("parsed call statement count = %d, want 4", len(statements))
	}
	direct := statements[0].Expr
	if direct == nil || direct.Kind != "call" || direct.Name != "helper" || direct.Right != nil || direct.Left == nil || direct.Left.Kind != "add" {
		t.Fatalf("direct call with computed argument = %+v", direct)
	}
	nested := statements[1].Expr
	if nested == nil || nested.Kind != "call" || nested.Name != "helper" || nested.Left == nil || nested.Left.Kind != "call" || nested.Left.Name != "other" {
		t.Fatalf("nested call did not use Left for its sole argument: %+v", nested)
	}
	contextual := statements[2].Expr
	if contextual == nil || contextual.Kind != "call" || contextual.Left == nil || contextual.Left.Kind != "name" || contextual.Left.Name != "true" {
		t.Fatalf("contextual true identifier changed inside call argument: %+v", contextual)
	}
	ordered := statements[3].Expr
	if ordered == nil || ordered.Kind != "add" || ordered.Left == nil || ordered.Left.Kind != "call" || ordered.Right == nil || ordered.Right.Kind != "call" {
		t.Fatalf("calls did not bind as scalar-expression atoms: %+v", ordered)
	}
	if ordered.Left.Name != "helper" || ordered.Right.Name != "helper" || ordered.Right.Left.Kind != "call" {
		t.Fatalf("sibling/nested call tree = %+v", ordered)
	}

	for _, tc := range []struct {
		name string
		expr string
		code string
	}{
		{name: "zero arguments", expr: "helper()", code: "syntax.scalar_call_arity"},
		{name: "two arguments", expr: "helper(input, other(input))", code: "syntax.scalar_call_arity"},
		{name: "trailing argument separator", expr: "helper(input,)", code: "syntax.scalar_call_arity"},
		{name: "missing argument", expr: "helper(", code: "syntax.scalar_call_arity"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := phase36CallWriteSource(tc.expr)
			malformed := syntax.Parse([]byte(source))
			if !bytes.Equal(malformed.Tree.Bytes(), []byte(source)) {
				t.Fatalf("arity recovery changed lossless bytes: %q", malformed.Tree.Bytes())
			}
			if !phase36HasSyntaxDiagnostic(malformed.Diagnostics, tc.code) {
				t.Fatalf("invalid call arity %q lacked %s: %+v", tc.expr, tc.code, malformed.Diagnostics)
			}
		})
	}
}

func TestPhase36ScalarCallDepthAndRecovery(t *testing.T) {
	for _, depth := range []int{63, 64, 65} {
		t.Run(fmt.Sprint(depth), func(t *testing.T) {
			expression := strings.Repeat("helper(", depth) + "input" + strings.Repeat(")", depth)
			parsed := syntax.Parse([]byte(phase36CallSource(expression)))
			refusals := 0
			for _, problem := range parsed.Diagnostics {
				if problem.Code == "surface.unsupported_expression_depth" {
					refusals++
				}
			}
			if depth <= syntax.MaxExpressionDepth && refusals != 0 {
				t.Fatalf("call nesting %d refused within the shared limit: %+v", depth, parsed.Diagnostics)
			}
			if depth > syntax.MaxExpressionDepth && refusals != 1 {
				t.Fatalf("call nesting %d depth refusals = %d, diagnostics=%+v", depth, refusals, parsed.Diagnostics)
			}
		})
	}

	mixed := strings.Repeat("helper((", 31) + "byte_at(view, input)" + strings.Repeat("))", 31)
	for _, tc := range []struct {
		name       string
		expression string
		wantRefuse bool
	}{
		{name: "mixed depth at 64", expression: "(" + mixed + ")"},
		{name: "mixed depth at 65", expression: "((" + mixed + "))", wantRefuse: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := syntax.Parse([]byte(phase36ByteViewSource(tc.expression)))
			refusals := 0
			for _, problem := range parsed.Diagnostics {
				if problem.Code == "surface.unsupported_expression_depth" {
					refusals++
				}
			}
			if tc.wantRefuse && refusals != 1 {
				t.Fatalf("mixed depth refusal count=%d, diagnostics=%+v", refusals, parsed.Diagnostics)
			}
			if !tc.wantRefuse && refusals != 0 {
				t.Fatalf("mixed depth at the limit refused: %+v", parsed.Diagnostics)
			}
		})
	}

	malformedSource := []byte("module phase36.recovery\nexport {\n  fn damaged\n  fn sibling\n}\nfn damaged(input: U64) -> U64 {\n  let value = helper(other(input)\nfn sibling(input: U64) -> U64 {\n  input\n}\n")
	malformed := syntax.Parse(malformedSource)
	if !bytes.Equal(malformed.Tree.Bytes(), malformedSource) {
		t.Fatalf("malformed nested call changed source bytes: %q", malformed.Tree.Bytes())
	}
	if len(malformed.Diagnostics) == 0 {
		t.Fatal("missing nested call delimiters were accepted")
	}
	if len(malformed.Program.Funcs) != 2 || malformed.Program.Funcs[1].Name != "sibling" {
		t.Fatalf("nested call recovery lost following sibling: %+v", malformed.Program.Funcs)
	}
}

func TestPhase36ScalarCallFormatting(t *testing.T) {
	source := []byte("module phase36.format\nexport {\n  fn evaluate\n}\nfn evaluate(input: U64) -> U64 {\nlet result=helper(\n// keep this note\ninput+3)\nresult\n}\n")
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("formatter input parse diagnostics: %+v", parsed.Diagnostics)
	}
	formatted := syntax.Format(parsed.Tree)
	if !bytes.Contains(formatted, []byte("// keep this note")) || !bytes.Contains(formatted, []byte("helper(")) {
		t.Fatalf("canonical output lost call spelling/comment: %q", formatted)
	}
	reparsed := syntax.Parse(formatted)
	if len(reparsed.Diagnostics) != 0 {
		t.Fatalf("formatted call did not reparse: %+v\n%s", reparsed.Diagnostics, formatted)
	}
	if again := syntax.Format(reparsed.Tree); !bytes.Equal(again, formatted) {
		t.Fatalf("call formatting is not idempotent:\nfirst: %s\nagain: %s", formatted, again)
	}
	if got := reparsed.Program.Funcs[0].Body.Linear.Statements[0].Expr; got == nil || got.Kind != "call" || got.Left.Kind != "add" {
		t.Fatalf("formatted AST call shape = %+v", got)
	}
}

func phase36CallSource(expression string) string {
	return "module phase36.call_expr\nexport {\n  fn evaluate\n}\nfn evaluate(input: U64) -> U64 {\n  let result = " + expression + "\n  result\n}\n"
}

func phase36CallWriteSource(expression string) string {
	return "module phase36.call_write\nexport {\n  fn evaluate\n}\nfn evaluate(input: U64) -> U64 {\n  write_u64(" + expression + ")\n  input\n}\n"
}

func phase36ByteViewSource(expression string) string {
	return "module phase36.byte_view\nexport {\n  fn evaluate\n}\nfn evaluate(view: ByteView) -> U64 {\n  let result = " + expression + "\n  result\n}\n"
}

func phase36HasSyntaxDiagnostic(diagnostics []diagnostic.Diagnostic, code string) bool {
	for _, problem := range diagnostics {
		if problem.Code == code {
			return true
		}
	}
	return false
}

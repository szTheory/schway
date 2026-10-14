package syntax_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/surface"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase35ScalarGrammar(t *testing.T) {
	parsed, expression := parsePhase35Expression(t, "x ^ (x << 3) & 0xFF")
	if got, want := phase35ExpressionShape(expression), "bit_xor(name:x,bit_and(shift_left(name:x,number:3),number:0xFF))"; got != want {
		t.Fatalf("locked grouped witness = %s, want %s", got, want)
	}
	if got := parsed.Tree.Bytes(); !bytes.Contains(got, []byte("x ^ (x << 3) & 0xFF")) {
		t.Fatalf("lossless source bytes lost the witness spelling: %q", got)
	}
	if expression.Right.Left.Span.Start != strings.Index(string(parsed.Tree.Source), "(x << 3)") ||
		expression.Right.Left.Span.End != strings.Index(string(parsed.Tree.Source), " & 0xFF") {
		t.Fatalf("transparent group did not extend the inner expression span: %+v", expression.Right.Left.Span)
	}
	if !bytes.Equal(parsed.Tree.Bytes(), parsed.Tree.Source) {
		t.Fatalf("lossless tree bytes differ from the source: got %q want %q", parsed.Tree.Bytes(), parsed.Tree.Source)
	}

	_, alternate := parsePhase35Expression(t, "(x ^ (x << 3)) & 0xFF")
	if got := phase35ExpressionShape(alternate); got != "bit_and(bit_xor(name:x,shift_left(name:x,number:3)),number:0xFF)" {
		t.Fatalf("discriminating alternate grouping = %s", got)
	}
	if phase35ExpressionShape(alternate) == phase35ExpressionShape(expression) {
		t.Fatal("the alternate grouping collapsed to the locked witness tree")
	}

	_, expression = parsePhase35Expression(t, "x & x ^ x | x")
	if got, want := phase35ExpressionShape(expression), "bit_or(bit_xor(bit_and(name:x,name:x),name:x),name:x)"; got != want {
		t.Fatalf("locked precedence = %s, want %s", got, want)
	}

	_, expression = parsePhase35Expression(t, "x ^ x ^ x")
	if got, want := phase35ExpressionShape(expression), "bit_xor(bit_xor(name:x,name:x),name:x)"; got != want {
		t.Fatalf("same-tier operators did not associate left: %s", got)
	}
	for _, tc := range []struct {
		expr string
		kind string
	}{
		{expr: "x ^ x", kind: "bit_xor"},
		{expr: "x & x", kind: "bit_and"},
		{expr: "x | x", kind: "bit_or"},
	} {
		_, equal := parsePhase35Expression(t, tc.expr)
		if equal.Kind != tc.kind || equal.Left.Name != "x" || equal.Right.Name != "x" {
			t.Fatalf("equal operands for %s = %+v", tc.expr, equal)
		}
	}
	_, expression = parsePhase35Expression(t, "x << 1 >> 2")
	if got, want := phase35ExpressionShape(expression), "shift_right(shift_left(name:x,number:1),number:2)"; got != want {
		t.Fatalf("shift tier did not associate left: %s", got)
	}

	_, expression = parsePhase35Expression(t, "x << ((3))")
	if expression.Kind != "shift_left" || expression.Right.Kind != "number" || expression.Right.Value != "3" {
		t.Fatalf("transparent numeric groups changed the amount leaf: %+v", expression)
	}
	_, expression = parsePhase35Expression(t, "x << (1 + 2)")
	if expression.Right.Kind != "add" {
		t.Fatalf("computed grouped expression lost its non-literal identity: %+v", expression.Right)
	}

	for _, tc := range []struct {
		name string
		expr string
		code string
	}{
		{name: "logical and", expr: "x && x", code: "surface.unsupported_logical_and"},
		{name: "logical or", expr: "x || x", code: "surface.unsupported_logical_or"},
		{name: "logical not", expr: "!x", code: "surface.unsupported_logical_not"},
		{name: "negative shift amount", expr: "x << -1", code: "surface.unsupported_shift_amount"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := syntax.Parse([]byte(phase35Source(tc.expr)))
			if !hasPhase35Diagnostic(parsed.Diagnostics, tc.code) {
				t.Fatalf("diagnostics do not preserve %s ownership: %+v", tc.code, parsed.Diagnostics)
			}
		})
	}
	logicalOrSource := []byte(phase35Source("x || x"))
	logicalOrFormatted := syntax.Format(syntax.Parse(logicalOrSource).Tree)
	if !bytes.Contains(logicalOrFormatted, []byte("x || x")) || bytes.Contains(logicalOrFormatted, []byte("x | | x")) {
		t.Fatalf("formatter split logical OR into bitwise separators: %q", logicalOrFormatted)
	}
	if refusal, ok := surface.LookupRefusal("surface.unsupported_shift_amount"); !ok || refusal.Hint != "Shift amounts must be integer literals from 0 to 63; write a literal in that range." {
		t.Fatalf("shift amount refusal = %+v, found=%v", refusal, ok)
	}

	for _, tc := range []struct {
		name string
		expr string
		code string
	}{
		{name: "empty group", expr: "()", code: "syntax.expected_scalar_operand"},
		{name: "missing operand", expr: "x ^", code: "syntax.expected_scalar_operand"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := phase35Source(tc.expr)
			if tc.name == "missing operand" {
				source = strings.Replace(source, "  result\n}\n", "}\n", 1)
			}
			parsed := syntax.Parse([]byte(source))
			if !hasPhase35Diagnostic(parsed.Diagnostics, tc.code) {
				t.Fatalf("diagnostics do not retain %s recovery: %+v", tc.code, parsed.Diagnostics)
			}
		})
	}

	for _, expression := range []string{"x < < x", "x < /* separated */ < x", "x > > x"} {
		spaced := syntax.Parse([]byte(phase35Source(expression)))
		if len(spaced.Program.Funcs) != 1 || spaced.Program.Funcs[0].Body.Linear == nil || findPhase35Expr(spaced.Program.Funcs[0].Body.Linear.Statements, "shift_left") || findPhase35Expr(spaced.Program.Funcs[0].Body.Linear.Statements, "shift_right") {
			t.Fatalf("noncontiguous angle tokens were parsed as a shift for %q: %+v", expression, spaced.Program.Funcs)
		}
	}
}

func TestPhase35FormatterRoundTrip(t *testing.T) {
	source := []byte("module phase35.format\n\nexport {\n  fn evaluate\n}\n\nfn evaluate(x: U64) -> U64 {\n  let result = x ^ (x << 3) & 0xFF // keep this comment\n  result\n}\n\nfn text(x: U64) -> Unit {\n  write_text(\">> remains text\")\n}\n")
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %+v", parsed.Diagnostics)
	}
	if !bytes.Equal(parsed.Tree.Bytes(), source) {
		t.Fatalf("lossless tree changed source bytes: got %q want %q", parsed.Tree.Bytes(), source)
	}
	first := syntax.Format(parsed.Tree)
	if !bytes.Contains(first, []byte("x ^ (x << 3) & 0xFF")) || !bytes.Contains(first, []byte("// keep this comment")) || !bytes.Contains(first, []byte("\">> remains text\"")) {
		t.Fatalf("formatter changed operators, comments, or string spelling: %q", first)
	}
	reparsed := syntax.Parse(first)
	if len(reparsed.Diagnostics) != 0 {
		t.Fatalf("formatted parse: %+v\n%s", reparsed.Diagnostics, first)
	}
	if got, want := phase35ExpressionShape(reparsed.Program.Funcs[0].Body.Linear.Statements[0].Expr), "bit_xor(name:x,bit_and(shift_left(name:x,number:3),number:0xFF))"; got != want {
		t.Fatalf("formatted AST = %s, want %s", got, want)
	}
	second := syntax.Format(reparsed.Tree)
	if !bytes.Equal(first, second) {
		t.Fatalf("format is not idempotent:\nfirst:  %q\nsecond: %q", first, second)
	}

	generic := []byte("module phase35.generic\n\nexport {\n  type Choice\n  fn relay\n  fn choose\n}\n\ndata Choice =\n  | On\n  | Off\n\nfn relay(value: Box<Box<U64>>) -> Box<Box<U64>> {\n  let held = take value\n  held\n}\n\nfn choose(flag: Choice) -> Choice {\n  match flag {\n    On => flag\n    Off => flag\n  }\n}\n")
	genericParsed := syntax.Parse(generic)
	if len(genericParsed.Diagnostics) != 0 {
		t.Fatalf("nested generic parse: %+v", genericParsed.Diagnostics)
	}
	formattedGeneric := syntax.Format(genericParsed.Tree)
	if !bytes.Contains(formattedGeneric, []byte("Box<Box<U64>>")) {
		t.Fatalf("nested generic closing angles drifted: %q", formattedGeneric)
	}
	if again := syntax.Format(syntax.Parse(formattedGeneric).Tree); !bytes.Equal(formattedGeneric, again) {
		t.Fatalf("nested generic formatting is not idempotent:\nfirst:  %q\nsecond: %q", formattedGeneric, again)
	}
}

func parsePhase35Expression(t *testing.T, expression string) (syntax.ParseResult, *ast.Expr) {
	t.Helper()
	parsed := syntax.Parse([]byte(phase35Source(expression)))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse %q: %+v", expression, parsed.Diagnostics)
	}
	if len(parsed.Program.Funcs) != 1 || parsed.Program.Funcs[0].Body.Linear == nil || len(parsed.Program.Funcs[0].Body.Linear.Statements) != 1 {
		t.Fatalf("parse %q did not produce one scalar let: %+v", expression, parsed.Program.Funcs)
	}
	return parsed, parsed.Program.Funcs[0].Body.Linear.Statements[0].Expr
}

func phase35Source(expression string) string {
	return fmt.Sprintf("module phase35.scalar\n\nexport {\n  fn evaluate\n}\n\nfn evaluate(x: U64) -> U64 {\n  let result = %s\n  result\n}\n", expression)
}

func phase35ExpressionShape(expression *ast.Expr) string {
	if expression == nil {
		return "<nil>"
	}
	if expression.Kind == "name" {
		return "name:" + expression.Name
	}
	if expression.Kind == "number" || expression.Kind == "text" {
		return expression.Kind + ":" + expression.Value
	}
	return expression.Kind + "(" + phase35ExpressionShape(expression.Left) + "," + phase35ExpressionShape(expression.Right) + ")"
}

func hasPhase35Diagnostic(diagnostics []diagnostic.Diagnostic, code string) bool {
	for _, problem := range diagnostics {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func findPhase35Expr(statements []ast.Statement, kind string) bool {
	var visit func(*ast.Expr) bool
	visit = func(expression *ast.Expr) bool {
		return expression != nil && (expression.Kind == kind || visit(expression.Left) || visit(expression.Right))
	}
	for _, statement := range statements {
		if visit(statement.Expr) || findPhase35Expr(statement.Then, kind) || findPhase35Expr(statement.Else, kind) {
			return true
		}
	}
	return false
}

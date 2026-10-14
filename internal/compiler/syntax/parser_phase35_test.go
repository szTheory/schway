package syntax_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/surface"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
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

func TestPhase35ExpressionDepth(t *testing.T) {
	if got := syntax.CurrentAuthoringBounds().ExpressionDepth; got != 64 {
		t.Fatalf("expression depth bound = %d, want 64", got)
	}
	if got := syntax.CurrentAuthoringBounds().TypeDepth; got != 64 {
		t.Fatalf("expression bound changed the independent type depth: %d", got)
	}
	if refusal, ok := surface.LookupRefusal("surface.unsupported_expression_depth"); !ok || refusal.Hint != "Expression nesting exceeds the parser limit of 64; reduce nested groups or byte_at indices." {
		t.Fatalf("expression-depth refusal = %+v, found=%v", refusal, ok)
	}

	for _, tc := range []struct {
		name  string
		build func(int) string
	}{
		{name: "groups", build: func(depth int) string { return strings.Repeat("(", depth) + "x" + strings.Repeat(")", depth) }},
		{name: "byte_at indices", build: phase35NestedIndexExpr},
		{name: "mixed entries", build: phase35MixedDepthExpr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, depth := range []int{63, 64, 65, 128} {
				t.Run(fmt.Sprint(depth), func(t *testing.T) {
					inputType := "U64"
					if strings.Contains(tc.build(depth), "byte_at") {
						inputType = "ByteView"
					}
					parsed := syntax.Parse([]byte(phase35DepthSource(tc.build(depth), inputType)))
					gotRefusals := phase35DiagnosticCount(parsed.Diagnostics, "surface.unsupported_expression_depth")
					if depth <= 64 && gotRefusals != 0 {
						t.Fatalf("depth %d was refused: %+v", depth, parsed.Diagnostics)
					}
					if depth > 64 && gotRefusals != 1 {
						t.Fatalf("depth %d expression-depth refusal count=%d, diagnostics=%+v", depth, gotRefusals, parsed.Diagnostics)
					}
					if depth == 65 && tc.name == "groups" {
						source := phase35DepthSource(tc.build(depth), inputType)
						wantStart := strings.Index(source, "let result = ") + len("let result = ") + 65
						var found bool
						for _, problem := range parsed.Diagnostics {
							if problem.Code == "surface.unsupported_expression_depth" {
								found = problem.Primary == (diagnostic.Span{Start: wantStart, End: wantStart + 1})
							}
						}
						if !found {
							t.Fatalf("65th group entry refusal span did not identify its operand: %+v", parsed.Diagnostics)
						}
					}
				})
			}
		})
	}

	deepMalformed := "module phase35.depth\nexport {\n  fn damaged\n  fn sibling\n}\nfn damaged(input: U64) -> U64 {\n  let value = " + strings.Repeat("(", 2048) + "x\nfn sibling(input: U64) -> U64 {\n  input\n}\n"
	malformed := syntax.Parse([]byte(deepMalformed))
	if !hasPhase35Diagnostic(malformed.Diagnostics, "surface.unsupported_expression_depth") {
		t.Fatalf("very deep malformed group chain lacks its bounded-depth refusal: %+v", malformed.Diagnostics)
	}
	if !phase35HasFunction(malformed.Program.Funcs, "sibling") {
		t.Fatalf("recovery did not preserve the following declaration sibling: %+v", malformed.Program.Funcs)
	}
	if len(malformed.Diagnostics) > syntax.CurrentAuthoringBounds().PrimaryDiagnostics+1 {
		t.Fatalf("malformed group chain exceeded bounded diagnostic work: %d", len(malformed.Diagnostics))
	}

	eofSource := "module phase35.depth_eof\nexport {\n  fn damaged\n}\nfn damaged(input: U64) -> U64 {\n  let value = " + strings.Repeat("(", 65) + "x"
	eof := syntax.Parse([]byte(eofSource))
	if !hasPhase35Diagnostic(eof.Diagnostics, "surface.unsupported_expression_depth") {
		t.Fatalf("EOF depth refusal was lost: %+v", eof.Diagnostics)
	}

	logicalNot := syntax.Parse([]byte(phase35DepthSource(strings.Repeat("!", 1024)+"x", "U64")))
	if !hasPhase35Diagnostic(logicalNot.Diagnostics, "surface.unsupported_logical_not") || len(logicalNot.Diagnostics) > syntax.CurrentAuthoringBounds().PrimaryDiagnostics+1 {
		t.Fatalf("long logical-not chain was not processed iteratively within diagnostic limits: %+v", logicalNot.Diagnostics)
	}
	unsupportedBinary := syntax.Parse([]byte(phase35DepthSource(strings.TrimSuffix(strings.Repeat("x - ", 4096), " - ")+" - x", "U64")))
	if len(unsupportedBinary.Diagnostics) != syntax.CurrentAuthoringBounds().PrimaryDiagnostics+1 || unsupportedBinary.Diagnostics[len(unsupportedBinary.Diagnostics)-1].Code != "syntax.too_many_errors" {
		t.Fatalf("long unsupported binary chain did not stop at the parser's diagnostic work cap: %+v", unsupportedBinary.Diagnostics)
	}

	phase35AssertNumericDelta(t)
}

func TestPhase35GroupedNegativeShiftAmountRefusal(t *testing.T) {
	for _, tc := range []struct {
		name string
		expr string
		kind string
	}{
		{name: "grouped", expr: "x << (-1)", kind: "shift_left"},
		{name: "nested groups", expr: "x >> ((-1))", kind: "shift_right"},
		{name: "comments between tokens", expr: "x << (// before sign\n - // before digits\n 1 )", kind: "shift_left"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := phase35Source(tc.expr)
			parsed := syntax.Parse([]byte(source))
			if len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Code != "surface.unsupported_shift_amount" {
				t.Fatalf("grouped negative shift amount diagnostics = %+v, want only surface.unsupported_shift_amount", parsed.Diagnostics)
			}
			expressionStart := strings.Index(source, "let result = ") + len("let result = ")
			minus := expressionStart + strings.Index(source[expressionStart:], "-")
			if parsed.Diagnostics[0].Primary != (diagnostic.Span{Start: minus, End: minus + 1}) {
				t.Fatalf("shift refusal span = %+v, want minus at %d", parsed.Diagnostics[0].Primary, minus)
			}
			expression := parsed.Program.Funcs[0].Body.Linear.Statements[0].Expr
			if expression.Kind != tc.kind || expression.Right == nil || expression.Right.Kind != "number" || expression.Right.Value != "-1" {
				t.Fatalf("grouped negative amount AST = %+v", expression)
			}
		})
	}

	recovery := syntax.Parse([]byte(phase35Source("x - y << (-1)")))
	if len(recovery.Diagnostics) != 2 ||
		recovery.Diagnostics[0].Code != "surface.unsupported_subtraction" ||
		recovery.Diagnostics[1].Code != "surface.unsupported_shift_amount" {
		t.Fatalf("grouped negative shift amount during suffix recovery = %+v", recovery.Diagnostics)
	}

	tooDeep := "x << " + strings.Repeat("(", syntax.CurrentAuthoringBounds().ExpressionDepth+1) + "-1" + strings.Repeat(")", syntax.CurrentAuthoringBounds().ExpressionDepth+1)
	parsed := syntax.Parse([]byte(phase35Source(tooDeep)))
	if phase35DiagnosticCount(parsed.Diagnostics, "surface.unsupported_expression_depth") != 1 || hasPhase35Diagnostic(parsed.Diagnostics, "surface.unsupported_shift_amount") {
		t.Fatalf("over-depth negative amount bypassed the expression bound: %+v", parsed.Diagnostics)
	}
}

func TestPhase35RecoveryDeltaControls(t *testing.T) {
	phase34AssertFrozenRecoveryInputs(t)
	corpus, phase34After := loadPhase34RecoveryCorpus(t)
	phase35After := loadPhase35RecoveryAfter(t)
	historical := fixtureAfterByID(t, phase34After.Drafts, "crc32-reconstructed")
	draft := fixtureByID(t, corpus.Drafts, "crc32-reconstructed")
	parsed := syntax.Parse([]byte(draft.Source))
	got := observePhase34Recovery(parsed)
	if reason := comparePhase35RecoveryAfter(phase35After, historical, got); reason != "" {
		t.Fatalf("literal Phase 35 recovery projection failed before controls: %s", reason)
	}

	missing := phase35After
	missing.Amendments = nil
	if reason := comparePhase35RecoveryAfter(missing, historical, got); reason != "amendment_count" {
		t.Fatalf("missing amendment reached %q, want amendment_count", reason)
	}
	t.Log("missing Phase 35 amendment reached amendment_count")
	if reason := comparePhase35RecoveryAfter(phase35After, historical, got); reason != "" {
		t.Fatalf("missing amendment control did not restore: %s", reason)
	}
	t.Log("missing Phase 35 amendment control restored")

	wrongRow := phase35After
	wrongRow.Amendments = append([]phase35RecoveryAmendment(nil), phase35After.Amendments...)
	wrongRow.Amendments[0].Fixture = "word-counter-reconstructed"
	if reason := comparePhase35RecoveryAfter(wrongRow, historical, got); reason != "wrong_fixture" {
		t.Fatalf("wrong amendment row reached %q, want wrong_fixture", reason)
	}
	t.Log("wrong Phase 35 amendment row reached wrong_fixture")
	if reason := comparePhase35RecoveryAfter(phase35After, historical, got); reason != "" {
		t.Fatalf("wrong row control did not restore: %s", reason)
	}
	t.Log("wrong Phase 35 amendment row control restored")

	wrongBefore := phase35After
	wrongBefore.Amendments = append([]phase35RecoveryAmendment(nil), phase35After.Amendments...)
	wrongBefore.Amendments[0].Phase34Projection = clonePhase35Projection(phase35After.Amendments[0].Phase34Projection)
	wrongBefore.Amendments[0].Phase34Projection.Diagnostics[0].PrimarySpan.Start++
	if reason := comparePhase35RecoveryAfter(wrongBefore, historical, got); reason != "wrong_phase34_projection" {
		t.Fatalf("wrong historical projection reached %q, want wrong_phase34_projection", reason)
	}
	t.Log("wrong Phase 34 projection reached wrong_phase34_projection")
	if reason := comparePhase35RecoveryAfter(phase35After, historical, got); reason != "" {
		t.Fatalf("wrong before control did not restore: %s", reason)
	}
	t.Log("wrong Phase 34 projection control restored")

	wrongAfter := phase35After
	wrongAfter.Amendments = append([]phase35RecoveryAmendment(nil), phase35After.Amendments...)
	wrongAfter.Amendments[0].Phase35Projection.Diagnostics = []phase34RecoveryDiagnostic{{Code: "syntax.expected_rbrace", PrimarySpan: phase34RecoverySpan{Start: 184, End: 185}, Causes: []phase34RecoveryCause{}}}
	wrongAfter.Amendments[0].Phase35Projection.DiagnosticCount = 1
	if reason := comparePhase35RecoveryAfter(wrongAfter, historical, got); reason != "wrong_phase35_projection" {
		t.Fatalf("wrong current projection reached %q, want wrong_phase35_projection", reason)
	}
	t.Log("wrong Phase 35 projection reached wrong_phase35_projection")
	if reason := comparePhase35RecoveryAfter(phase35After, historical, got); reason != "" {
		t.Fatalf("wrong after control did not restore: %s", reason)
	}
	t.Log("wrong Phase 35 projection control restored")

	wrongSibling := got
	wrongSibling.RecoveredSiblingNames = []string{"unexpected"}
	if reason := comparePhase35RecoveryAfter(phase35After, historical, wrongSibling); reason != "wrong_phase35_projection" {
		t.Fatalf("wrong sibling reached %q, want wrong_phase35_projection", reason)
	}
	t.Log("wrong Phase 35 sibling reached wrong_phase35_projection")
	if reason := comparePhase35RecoveryAfter(phase35After, historical, got); reason != "" {
		t.Fatalf("wrong sibling control did not restore: %s", reason)
	}
	t.Log("wrong Phase 35 sibling control restored")
}

func phase35NestedIndexExpr(depth int) string {
	expression := "0"
	for range depth {
		expression = "byte_at(input, " + expression + ")"
	}
	return expression
}

func phase35MixedDepthExpr(depth int) string {
	expression := "x"
	for entry := 0; entry < depth; entry++ {
		if entry%2 == 0 {
			expression = "(" + expression + ")"
		} else {
			expression = "byte_at(input, " + expression + ")"
		}
	}
	return expression
}

func phase35DepthSource(expression, inputType string) string {
	return fmt.Sprintf("module phase35.depth\nexport {\n  fn evaluate\n}\nfn evaluate(input: %s) -> U64 {\n  let result = %s\n  result\n}\n", inputType, expression)
}

func phase35DiagnosticCount(diagnostics []diagnostic.Diagnostic, code string) int {
	count := 0
	for _, problem := range diagnostics {
		if problem.Code == code {
			count++
		}
	}
	return count
}

func phase35HasFunction(functions []ast.FuncDecl, name string) bool {
	for _, function := range functions {
		if function.Name == name {
			return true
		}
	}
	return false
}

func phase35AssertNumericDelta(t *testing.T) {
	t.Helper()
	type depthOutcome struct {
		Depth int    `json:"depth"`
		Parse string `json:"parse"`
		Check string `json:"check,omitempty"`
		Emit  string `json:"emit,omitempty"`
	}
	var delta struct {
		Schema     string `json:"schema"`
		Date       string `json:"date"`
		Historical struct {
			Fixture  string         `json:"fixture"`
			Revision string         `json:"revision"`
			Indices  []depthOutcome `json:"nested_index"`
		} `json:"historical"`
		AfterState struct {
			Limit       int            `json:"expression_depth_limit"`
			RefusalCode string         `json:"refusal_code"`
			Indices     []depthOutcome `json:"nested_index"`
			Groups      []depthOutcome `json:"groups"`
			Mixed       []depthOutcome `json:"mixed"`
		} `json:"after_state"`
	}
	contents, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "numeric-delta.json"))
	if err != nil {
		t.Fatalf("read numeric resource delta: %v", err)
	}
	if err := json.Unmarshal(contents, &delta); err != nil {
		t.Fatalf("decode numeric resource delta: %v", err)
	}
	if delta.Schema != "phase35.numeric-delta/1" || delta.Date != "2026-10-14" || delta.Historical.Fixture != "numeric-policy-baseline.json" || delta.Historical.Revision != "c4a2eec1ab47a6e5eabc7ff7ff9093e21854aa4b" {
		t.Fatalf("invalid historical resource delta header: %+v", delta)
	}
	wantDepths := []int{63, 64, 65, 128}
	if len(delta.Historical.Indices) != len(wantDepths) {
		t.Fatalf("historical nested-index observations=%+v", delta.Historical.Indices)
	}
	for index, wantDepth := range wantDepths {
		row := delta.Historical.Indices[index]
		if row.Depth != wantDepth || row.Parse != "admitted" || row.Check != "admitted" || row.Emit != "emitted" {
			t.Fatalf("historical depth row %d changed: %+v", wantDepth, row)
		}
	}
	if delta.AfterState.Limit != 64 || delta.AfterState.RefusalCode != "surface.unsupported_expression_depth" {
		t.Fatalf("invalid after-state depth policy: %+v", delta.AfterState)
	}
	for _, rows := range [][]depthOutcome{delta.AfterState.Indices, delta.AfterState.Groups, delta.AfterState.Mixed} {
		if len(rows) != len(wantDepths) {
			t.Fatalf("after-state boundary rows=%+v", rows)
		}
		for index, wantDepth := range wantDepths {
			wantOutcome := "admitted"
			if wantDepth > delta.AfterState.Limit {
				wantOutcome = "refused"
			}
			if rows[index].Depth != wantDepth || rows[index].Parse != wantOutcome {
				t.Fatalf("after-state boundary row %d changed: %+v", wantDepth, rows[index])
			}
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

func TestPhase35FormatterPreservesSeparatedShiftTokens(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		comment    string
	}{
		{name: "whitespace", expression: "x > > x"},
		{name: "line comment", expression: "x > // separated angles\n  > x", comment: "// separated angles"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(phase35Source(tc.expression))
			parsed := syntax.Parse(source)
			if len(parsed.Program.Funcs) != 1 || parsed.Program.Funcs[0].Body.Linear == nil || findPhase35Expr(parsed.Program.Funcs[0].Body.Linear.Statements, "shift_right") {
				t.Fatalf("source unexpectedly contains a right shift for %q: %+v", tc.expression, parsed.Program.Funcs)
			}

			formatted := syntax.Format(parsed.Tree)
			if bytes.Contains(formatted, []byte(">>")) {
				t.Fatalf("formatter joined separated right angles for %q: %q", tc.expression, formatted)
			}
			if tc.comment != "" && !bytes.Contains(formatted, []byte(tc.comment)) {
				t.Fatalf("formatter dropped the separating comment: %q", formatted)
			}
			reparsed := syntax.Parse(formatted)
			if len(reparsed.Program.Funcs) != 1 || reparsed.Program.Funcs[0].Body.Linear == nil || findPhase35Expr(reparsed.Program.Funcs[0].Body.Linear.Statements, "shift_right") {
				t.Fatalf("formatted source reparsed as a right shift: %q (%+v)", formatted, reparsed.Program.Funcs)
			}
			if again := syntax.Format(reparsed.Tree); !bytes.Equal(formatted, again) {
				t.Fatalf("recovery formatting is not idempotent:\nfirst:  %q\nsecond: %q", formatted, again)
			}
		})
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

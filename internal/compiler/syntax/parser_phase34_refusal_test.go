package syntax_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/surface"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase34UnsupportedSurfaceRefusals(t *testing.T) {
	wantCatalog := []surface.Refusal{
		{Code: "surface.unsupported_logical_and", Form: "&&", Hint: "Logical && is unsupported; use explicit if/else control flow."},
		{Code: "surface.unsupported_logical_or", Form: "||", Hint: "Logical || is unsupported; use explicit if/else control flow."},
		{Code: "surface.unsupported_logical_not", Form: "!", Hint: "Logical ! is unsupported; use explicit if/else control flow."},
		{Code: "surface.unsupported_parameter_arity", Form: "more than one declaration parameter", Hint: "Declarations require exactly one parameter; pass one admitted value."},
		{Code: "surface.unsupported_if_without_else", Form: "if statement without else block", Hint: "An if statement requires an else block."},
		{Code: "surface.unsupported_bool_literal_tail", Form: "unbound true or false in a result tail", Hint: "Bool literals are unsupported in result tails; return an initialized value allowed by the declared result type."},
		{Code: "surface.unsupported_subtraction", Form: "-", Hint: "Subtraction is unsupported; no unchecked emulation is recommended."},
	}
	if got := surface.Refusals(); !reflect.DeepEqual(got, wantCatalog) {
		t.Fatalf("surface.Refusals() = %+v, want the independent seven-entry catalog %+v", got, wantCatalog)
	}
	for _, want := range wantCatalog {
		if got, ok := surface.LookupRefusal(want.Code); !ok || got != want {
			t.Errorf("surface.LookupRefusal(%q) = %+v, %v; want %+v, true", want.Code, got, ok, want)
		}
	}

	const source = "module phase34.logical_and\nexport { fn main }\nfn main(n: U64) -> U64 { var result = n if n < 2 && n < 3 { result = n } else { result = n } result }\n"
	const hint = "Logical && is unsupported; use explicit if/else control flow."

	parsed := syntax.Parse([]byte(source))
	if !bytes.Equal(parsed.Tree.Bytes(), []byte(source)) {
		t.Fatal("refused logical-and source did not remain lossless")
	}
	if len(parsed.Diagnostics) != 1 {
		t.Fatalf("logical-and diagnostics = %+v, want one dedicated refusal", parsed.Diagnostics)
	}
	problem := parsed.Diagnostics[0]
	start := strings.Index(source, "&&")
	if problem.Code != "surface.unsupported_logical_and" || problem.Message != hint || problem.Primary.Start != start || problem.Primary.End != start+2 {
		t.Fatalf("logical-and diagnostic = %+v, want code surface.unsupported_logical_and at [%d:%d] with hint %q", problem, start, start+2, hint)
	}

	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "surface.unsupported_logical_and" {
		t.Fatalf("session.Check diagnostics = %+v, want only the logical-and refusal", checked.Diagnostics)
	}
	if !reflect.DeepEqual(checked.Program, core.Program{}) {
		t.Fatalf("session.Check admitted rejected syntax: %+v", checked.Program)
	}
}

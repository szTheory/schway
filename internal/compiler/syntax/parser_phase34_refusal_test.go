package syntax_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
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

	tests := []struct {
		name       string
		source     string
		form       string
		spanOffset int
		spanLength int
		code       string
		hint       string
	}{
		{
			name:   "logical-and-result-tail",
			source: "module phase34.and_tail\nexport { fn main }\nfn main(flag: Bool) -> Bool { flag && true }\n",
			form:   "&&", code: "surface.unsupported_logical_and",
			hint: "Logical && is unsupported; use explicit if/else control flow.",
		},
		{
			name:   "logical-or-result-tail",
			source: "module phase34.or_tail\nexport { fn main }\nfn main(flag: Bool) -> Bool { flag || false }\n",
			form:   "||", code: "surface.unsupported_logical_or",
			hint: "Logical || is unsupported; use explicit if/else control flow.",
		},
		{
			name:   "logical-not-result-tail",
			source: "module phase34.not_tail\nexport { fn main }\nfn main(flag: Bool) -> Bool { !flag }\n",
			form:   "!", code: "surface.unsupported_logical_not",
			hint: "Logical ! is unsupported; use explicit if/else control flow.",
		},
		{
			name:   "subtraction-expression",
			source: "module phase34.subtraction\nexport { fn main }\nfn main(n: U64) -> U64 { n - 1 }\n",
			form:   "n - 1", spanOffset: 2, spanLength: 1, code: "surface.unsupported_subtraction",
			hint: "Subtraction is unsupported; no unchecked emulation is recommended.",
		},
		{
			name:   "schway-declaration-arity",
			source: "module phase34.schway_arity\nexport { fn main }\nfn combine(left: U64, right: U64) -> U64 { left }\nfn main(n: U64) -> U64 { n }\n",
			form:   "right: U64", code: "surface.unsupported_parameter_arity",
			hint: "Declarations require exactly one parameter; pass one admitted value.",
		},
		{
			name:   "foreign-declaration-arity",
			source: "module phase34.foreign_arity\nexport { fn main }\nforeign C { fn combine(left: U64, right: U64) -> U64 { abi: \"C\" } }\nfn main(n: U64) -> U64 { n }\n",
			form:   "right: U64", code: "surface.unsupported_parameter_arity",
			hint: "Declarations require exactly one parameter; pass one admitted value.",
		},
		{
			name:   "if-without-else",
			source: "module phase34.if_no_else\nexport { fn main }\nfn main(n: U64) -> U64 { var result = n if n < 1 { result = n } result }\n",
			form:   "if n < 1 { result = n }", code: "surface.unsupported_if_without_else",
			hint: "An if statement requires an else block.",
		},
	}
	for _, test := range tests {
		parsed := syntax.Parse([]byte(test.source))
		if !bytes.Equal(parsed.Tree.Bytes(), []byte(test.source)) {
			t.Errorf("%s changed lossless source bytes", test.name)
		}
		start := strings.Index(test.source, test.form)
		if start >= 0 {
			start += test.spanOffset
		}
		spanLength := test.spanLength
		if spanLength == 0 {
			spanLength = len(test.form)
		}
		want := diagnostic.Span{Start: start, End: start + spanLength}
		if start < 0 || len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Code != test.code || parsed.Diagnostics[0].Message != test.hint || parsed.Diagnostics[0].Primary != want {
			t.Errorf("%s diagnostics = %+v, want only %s at [%d:%d] with hint %q", test.name, parsed.Diagnostics, test.code, want.Start, want.End, test.hint)
		}
	}

	nearMisses := []struct {
		name   string
		source string
	}{
		{
			name:   "spaced-ampersands",
			source: "module phase34.spaced_and\nexport { fn main }\nfn main(n: U64) -> U64 { var result = n if n < 2 & & n < 3 { result = n } else { result = n } result }\n",
		},
		{
			name:   "spaced-pipes",
			source: "module phase34.spaced_or\nexport { fn main }\nfn main(flag: Bool) -> Bool { flag | | false }\n",
		},
		{
			name:   "missing-then-block-closer",
			source: "module phase34.missing_then_closer\nexport { fn main }\nfn main(n: U64) -> U64 { if n < 1 { result = n }\n",
		},
		{
			name:   "strings-comments-arrows-and-ownership",
			source: "module phase34.controls // && || ! - ->\nexport { fn main }\ndata Choice = Left | Right\nfn main(buffer: Buffer) -> Buffer { defect \"&& || ! - -> & & | |\" }\n",
		},
	}
	for _, test := range nearMisses {
		parsed := syntax.Parse([]byte(test.source))
		for _, problem := range parsed.Diagnostics {
			if strings.HasPrefix(problem.Code, "surface.unsupported_") {
				t.Errorf("%s was reclassified as a surface refusal: %+v", test.name, problem)
			}
		}
	}
}

package syntax_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func TestPhase35NumericOwnership(t *testing.T) {
	for _, literal := range []string{"0XfF", "0XF_FF", "0xA_B", "18446744073709551615"} {
		for _, route := range []string{"legacy", "scalar"} {
			t.Run(route+"/valid/"+literal, func(t *testing.T) {
				source := phase35NumericRouteSource(route, literal)
				parsed := syntax.Parse(source)
				if len(parsed.Diagnostics) != 0 {
					t.Fatalf("valid literal %q diagnostics=%+v", literal, parsed.Diagnostics)
				}
				if !bytes.Equal(parsed.Tree.Bytes(), source) {
					t.Fatalf("valid literal changed lossless source bytes: got %q want %q", parsed.Tree.Bytes(), source)
				}
				var value string
				var span diagnostic.Span
				function := parsed.Program.Funcs[0]
				if route == "legacy" {
					binding := function.Body.Linear.Bindings[0]
					value, span = binding.RHS.Source, binding.RHS.Span
				} else {
					expression := function.Body.Linear.Statements[0].Expr
					if expression == nil || expression.Kind != "number" {
						t.Fatalf("scalar literal AST = %+v", expression)
					}
					value, span = expression.Value, expression.Span
				}
				if value != literal || string(source[span.Start:span.End]) != literal {
					t.Fatalf("valid literal spelling/span changed: value=%q span=%+v", value, span)
				}
				formatted := syntax.Format(parsed.Tree)
				if !bytes.Contains(formatted, []byte("= "+literal)) {
					t.Fatalf("formatter changed valid numeric spelling %q: %q", literal, formatted)
				}
				if again := syntax.Format(syntax.Parse(formatted).Tree); !bytes.Equal(formatted, again) {
					t.Fatalf("valid numeric format is not idempotent: first=%q second=%q", formatted, again)
				}
			})
		}
	}

	malformed := []string{
		"0x", "0X", "0xG", "0Xg", "0x1G", "0b2", "0x_FF", "0xFF_",
		"0xFF__0", "0xFFu64", "42abc", "0xFFλ",
	}
	for _, literal := range malformed {
		t.Run("lex/"+literal, func(t *testing.T) {
			tokens, diagnostics := syntax.Lex([]byte(literal))
			if len(tokens) != 2 || tokens[0].Kind != syntax.TokenUnknown || tokens[0].Text != literal || tokens[0].Span.Start != 0 || tokens[0].Span.End != len(literal) {
				t.Fatalf("Lex(%q) candidate = %+v; want one full-span unknown token", literal, tokens)
			}
			if len(diagnostics) != 1 || diagnostics[0].Code != "syntax.malformed_numeric_literal" || diagnostics[0].Primary.Start != 0 || diagnostics[0].Primary.End != len(literal) {
				t.Fatalf("Lex(%q) diagnostics = %+v; want one full-span malformed-numeric cause", literal, diagnostics)
			}
		})
		for _, route := range []string{"legacy", "scalar"} {
			t.Run(route+"/malformed/"+literal, func(t *testing.T) {
				source := phase35MalformedNumericRouteSource(route, literal)
				parsed := syntax.Parse(source)
				if !bytes.Equal(parsed.Tree.Bytes(), source) {
					t.Fatalf("rejected literal changed lossless source bytes: got %q want %q", parsed.Tree.Bytes(), source)
				}
				if len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Code != "syntax.malformed_numeric_literal" {
					t.Fatalf("rejected literal %q has duplicated or unrelated parser causes: %+v", literal, parsed.Diagnostics)
				}
				problem := parsed.Diagnostics[0]
				if string(source[problem.Primary.Start:problem.Primary.End]) != literal || problem.Primary.End-problem.Primary.Start != len(literal) {
					t.Fatalf("rejected literal diagnostic span %+v does not cover complete %q", problem.Primary, literal)
				}
				formatted := syntax.Format(parsed.Tree)
				if !bytes.Contains(formatted, []byte("write_u64("+literal+")")) {
					t.Fatalf("formatter changed rejected numeric spelling %q: %q", literal, formatted)
				}
				reparsed := syntax.Parse(formatted)
				if len(reparsed.Diagnostics) != 1 || reparsed.Diagnostics[0].Code != "syntax.malformed_numeric_literal" {
					t.Fatalf("formatted rejected literal %q lost exact lexical ownership: %+v", formatted, reparsed.Diagnostics)
				}
			})
		}
	}

	if tokens, diagnostics := syntax.Lex([]byte("0x")); len(tokens) != 2 || len(diagnostics) != 1 || diagnostics[0].Primary != (diagnostic.Span{Start: 0, End: 2}) {
		t.Fatalf("empty hexadecimal prefix at EOF lost its full candidate span: tokens=%+v diagnostics=%+v", tokens, diagnostics)
	}

	closerSource := []byte("module phase35.numeric_closer\nexport { fn main }\nfn main(input: U64) -> Unit { write_u64(0x) }\n")
	closer := syntax.Parse(closerSource)
	if len(closer.Diagnostics) != 1 || closer.Diagnostics[0].Code != "syntax.malformed_numeric_literal" {
		t.Fatalf("malformed numeric ownership consumed or duplicated the caller's closer: %+v", closer.Diagnostics)
	}

	independentSource := []byte("module phase35.numeric_independent\nexport { fn main }\nfn main(input: Byte) -> U64 {\n  let bad = 0x\n  let also_bad = @\n  input\n}\n")
	independent := syntax.Parse(independentSource)
	malformedStart := strings.Index(string(independentSource), "0x")
	unknownStart := strings.Index(string(independentSource), "@")
	if !hasPhase35NumericDiagnostic(independent.Diagnostics, "syntax.malformed_numeric_literal", malformedStart, malformedStart+2) {
		t.Fatalf("independent malformed numeric cause was lost: %+v", independent.Diagnostics)
	}
	if !hasPhase35NumericDiagnostic(independent.Diagnostics, "syntax.expected_binding_source", unknownStart, unknownStart+1) {
		t.Fatalf("unrelated unknown-token parser cause was filtered: %+v", independent.Diagnostics)
	}

	var limited strings.Builder
	limited.WriteString("module phase35.numeric_limit\nexport { fn main }\nfn main(input: Byte) -> Unit {\n")
	for index := 0; index < syntax.CurrentAuthoringBounds().PrimaryDiagnostics+2; index++ {
		fmt.Fprintf(&limited, "  let bad_%d = 0x\n", index)
	}
	limited.WriteString("}\n")
	limitedParse := syntax.Parse([]byte(limited.String()))
	wantDiagnostics := syntax.CurrentAuthoringBounds().PrimaryDiagnostics + 1
	if len(limitedParse.Diagnostics) != wantDiagnostics || limitedParse.Diagnostics[len(limitedParse.Diagnostics)-1].Code != "syntax.too_many_errors" {
		t.Fatalf("malformed candidate stream lost bounded truncation marker: count=%d diagnostics=%+v", len(limitedParse.Diagnostics), limitedParse.Diagnostics)
	}
	for _, problem := range limitedParse.Diagnostics[:len(limitedParse.Diagnostics)-1] {
		if problem.Code != "syntax.malformed_numeric_literal" {
			t.Fatalf("bounded malformed stream contains a duplicate parser expectation: %+v", limitedParse.Diagnostics)
		}
	}
}

func phase35NumericRouteSource(route, literal string) []byte {
	parameter := "Byte"
	if route == "scalar" {
		parameter = "U64"
	}
	return []byte(fmt.Sprintf("module phase35.numeric.%s\nexport { fn main }\nfn main(input: %s) -> U64 {\n  let value = %s\n  value\n}\n", route, parameter, literal))
}

func phase35MalformedNumericRouteSource(route, literal string) []byte {
	parameter := "Byte"
	if route == "scalar" {
		parameter = "U64"
	}
	return []byte(fmt.Sprintf("module phase35.numeric.%s\nexport { fn main }\nfn main(input: %s) -> Unit { write_u64(%s) }\n", route, parameter, literal))
}

func hasPhase35NumericDiagnostic(diagnostics []diagnostic.Diagnostic, code string, start, end int) bool {
	for _, problem := range diagnostics {
		if problem.Code == code && problem.Primary.Start == start && problem.Primary.End == end {
			return true
		}
	}
	return false
}

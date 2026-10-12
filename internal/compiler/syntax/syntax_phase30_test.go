package syntax_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// phase30Item is the span-free projection of one ast.ImportItem.
type phase30Item struct{ Kind, Name string }

func phase30Items(items []ast.ImportItem) []phase30Item {
	result := make([]phase30Item, len(items))
	for index, item := range items {
		result[index] = phase30Item{Kind: item.Kind, Name: item.Name}
	}
	return result
}

const phase30Tracer = "module examples.phase30.adler32_cli\nimport adler32 { fn adler32 }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n"

// TestPhase30ImportGrammar pins the D-01/D-02 import grammar. The tracer row
// is the D-01 example; the refusal rows are added with the refusals.
func TestPhase30ImportGrammar(t *testing.T) {
	for _, test := range []struct {
		name    string
		source  string
		module  string
		items   []phase30Item
		imports int
	}{
		{"tracer D-01 example", phase30Tracer, "adler32", []phase30Item{{"fn", "adler32"}}, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			parsed := syntax.Parse([]byte(test.source))
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("diagnostics: %v", diagnosticIdentity(parsed.Diagnostics))
			}
			if len(parsed.Program.Imports) != test.imports {
				t.Fatalf("imports = %+v, want %d", parsed.Program.Imports, test.imports)
			}
			declaration := parsed.Program.Imports[0]
			if declaration.Module != test.module {
				t.Fatalf("module = %q, want %q", declaration.Module, test.module)
			}
			if got := phase30Items(declaration.Items); len(got) != len(test.items) || got[0] != test.items[0] {
				t.Fatalf("items = %+v, want %+v", got, test.items)
			}
			if got := test.source[declaration.Span.Start:declaration.Span.End]; got != "import adler32 { fn adler32 }" {
				t.Fatalf("declaration span covers %q", got)
			}
			if got := test.source[declaration.ModuleSpan.Start:declaration.ModuleSpan.End]; got != "adler32" {
				t.Fatalf("module span covers %q", got)
			}
		})
	}
}

// TestPhase30FormatImport pins the canonical import layout from hand-written
// messy spacing and requires the formatter to be a fixed point on it.
func TestPhase30FormatImport(t *testing.T) {
	messy := "module examples.phase30.adler32_cli   import   adler32 {   fn   adler32 }\n\n\n  export { fn main }\nfn main(n: U64) -> U64 { n }\n"
	want := "module examples.phase30.adler32_cli\n\nimport adler32 {\n  fn adler32\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  n\n}\n"
	parsed := syntax.Parse([]byte(messy))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %v", diagnosticIdentity(parsed.Diagnostics))
	}
	first := syntax.Format(parsed.Tree)
	if string(first) != want {
		t.Fatalf("canonical layout:\n%q\nwant:\n%q", first, want)
	}
	again := syntax.Parse(first)
	if len(again.Diagnostics) != 0 || string(syntax.Format(again.Tree)) != want {
		t.Fatalf("formatter is not idempotent over imports: %q", syntax.Format(again.Tree))
	}
}

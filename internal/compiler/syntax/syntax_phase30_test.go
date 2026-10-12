package syntax_test

import (
	"bytes"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
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

// phase30Codes returns the diagnostic codes in emission order.
func phase30Codes(diagnostics []diagnostic.Diagnostic) []string { return diagnosticCodes(diagnostics) }

func phase30HasPrefix(diagnostics []diagnostic.Diagnostic, prefix string) bool {
	for _, problem := range diagnostics {
		if strings.HasPrefix(problem.Code, prefix) {
			return true
		}
	}
	return false
}

func phase30Apply(t *testing.T, source string, repair diagnostic.Repair) string {
	t.Helper()
	if repair.Span == nil {
		t.Fatalf("repair %s has no span", repair.Kind)
	}
	return source[:repair.Span.Start] + repair.Replacement + source[repair.Span.End:]
}

// TestPhase30ImportRefusals pins every D-02 and D-25 parse-time refusal that
// carries no repair, plus the empty-edge and no-import edge.
func TestPhase30ImportRefusals(t *testing.T) {
	const header = "module m\n"
	const footer = "export { fn main }\nfn main(n: U64) -> U64 { n }\n"
	for _, test := range []struct {
		name   string
		source string
		code   string
		exact  bool // exactly one diagnostic, with this code
	}{
		{"empty item list", header + "import a {}\n" + footer, "syntax.expected_import_name", true},
		{"kind without a name", header + "import a { fn }\n" + footer, "syntax.expected_import_name", true},
		{"comma separated items", header + "import a { fn x, fn y }\n" + footer, "syntax.", false},
		{"alias form", header + "import a { fn x as y }\n" + footer, "syntax.", false},
		{"glob form", header + "import a { * }\n" + footer, "syntax.", false},
		{"import as identifier", "module m\nexport { fn main }\nfn main(import: U64) -> U64 { import }\n", "syntax.", false},
		{"qualified call", header + "import a { fn f }\nexport { fn main }\nfn main(n: U64) -> U64 {\n  let r = a.f(n)\n  r\n}\n", "syntax.", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			parsed := syntax.Parse([]byte(test.source))
			if test.exact {
				if got := phase30Codes(parsed.Diagnostics); len(got) != 1 || got[0] != test.code {
					t.Fatalf("codes = %v, want exactly [%s]", got, test.code)
				}
				return
			}
			if !phase30HasPrefix(parsed.Diagnostics, test.code) {
				t.Fatalf("codes = %v, want a %s* refusal", phase30Codes(parsed.Diagnostics), test.code)
			}
		})
	}

	t.Run("empty item list is refused at the brace", func(t *testing.T) {
		source := header + "import a {}\n" + footer
		parsed := syntax.Parse([]byte(source))
		brace := strings.Index(source, "}")
		if len(parsed.Diagnostics) != 1 || parsed.Diagnostics[0].Primary.Start != brace {
			t.Fatalf("diagnostics = %v, want one at the closing brace %d", diagnosticIdentity(parsed.Diagnostics), brace)
		}
	})
	t.Run("no import declarations leaves Imports nil", func(t *testing.T) {
		parsed := syntax.Parse([]byte(header + footer))
		if len(parsed.Diagnostics) != 0 || parsed.Program.Imports != nil {
			t.Fatalf("diagnostics=%v imports=%+v", diagnosticIdentity(parsed.Diagnostics), parsed.Program.Imports)
		}
	})
	t.Run("type items stay in the AST", func(t *testing.T) {
		parsed := syntax.Parse([]byte(header + "import a { type T  fn f }\n" + footer))
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("diagnostics: %v", diagnosticIdentity(parsed.Diagnostics))
		}
		got := phase30Items(parsed.Program.Imports[0].Items)
		if len(got) != 2 || got[0] != (phase30Item{"type", "T"}) || got[1] != (phase30Item{"fn", "f"}) {
			t.Fatalf("items = %+v", got)
		}
	})
	t.Run("merged-program caps are exported", func(t *testing.T) {
		if syntax.MaxFunctions != 1024 || syntax.MaxDeclarations != 4096 || syntax.MaxImportsPerModule != 16 {
			t.Fatalf("caps = %d %d %d", syntax.MaxFunctions, syntax.MaxDeclarations, syntax.MaxImportsPerModule)
		}
	})
}

// TestPhase30ImportRepairs pins the only two import repairs: moving a late
// import above the export block, and deleting a subset duplicate (D-04, D-19).
func TestPhase30ImportRepairs(t *testing.T) {
	t.Run("import after export carries one rotate repair", func(t *testing.T) {
		source := "module x\nexport { fn main }\nimport a { fn f }\nfn main(n: U64) -> U64 { n }\n"
		parsed := syntax.Parse([]byte(source))
		if got := phase30Codes(parsed.Diagnostics); len(got) != 1 || got[0] != "syntax.import_after_export" {
			t.Fatalf("codes = %v", got)
		}
		repairs := parsed.Diagnostics[0].Repairs
		if len(repairs) != 1 || repairs[0].Kind != "move_import_above_export" || !diagnostic.DriverEligible(repairs[0]) {
			t.Fatalf("repairs = %+v", repairs)
		}
		exportStart, importEnd := strings.Index(source, "export"), strings.Index(source, "}\nfn main")+1
		if repairs[0].Span.Start != exportStart || repairs[0].Span.End != importEnd {
			t.Fatalf("repair span = %+v, want [%d,%d)", *repairs[0].Span, exportStart, importEnd)
		}
		fixed := phase30Apply(t, source, repairs[0])
		again := syntax.Parse([]byte(fixed))
		if len(again.Diagnostics) != 0 {
			t.Fatalf("applied repair does not parse clean: %q diagnostics=%v", fixed, diagnosticIdentity(again.Diagnostics))
		}
		if len(again.Program.Imports) != 1 || again.Program.Imports[0].Module != "a" || phase30Items(again.Program.Imports[0].Items)[0] != (phase30Item{"fn", "f"}) {
			t.Fatalf("imports after repair = %+v", again.Program.Imports)
		}
		if len(parsed.Program.Imports) != 1 {
			t.Fatalf("late import was not kept for recovery: %+v", parsed.Program.Imports)
		}
	})

	t.Run("subset duplicate carries one delete repair", func(t *testing.T) {
		source := "module x\nimport a { fn f  fn g }\nimport a { fn g }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n"
		parsed := syntax.Parse([]byte(source))
		if got := phase30Codes(parsed.Diagnostics); len(got) != 1 || got[0] != "module.duplicate_import" {
			t.Fatalf("codes = %v", got)
		}
		repairs := parsed.Diagnostics[0].Repairs
		if len(repairs) != 1 || repairs[0].Kind != "delete_duplicate_import" || repairs[0].Replacement != "\n" || !diagnostic.DriverEligible(repairs[0]) {
			t.Fatalf("repairs = %+v", repairs)
		}
		later := strings.Index(source, "import a { fn g }")
		if parsed.Diagnostics[0].Primary.Start != later {
			t.Fatalf("primary = %+v, want the later declaration at %d", parsed.Diagnostics[0].Primary, later)
		}
		if len(parsed.Diagnostics[0].Causes) == 0 {
			t.Fatalf("no cause names the earlier declaration")
		}
		fixed := phase30Apply(t, source, repairs[0])
		again := syntax.Parse([]byte(fixed))
		if len(again.Diagnostics) != 0 || len(again.Program.Imports) != 1 || len(again.Program.Imports[0].Items) != 2 {
			t.Fatalf("applied repair: %q diagnostics=%v imports=%+v", fixed, diagnosticIdentity(again.Diagnostics), again.Program.Imports)
		}
	})

	t.Run("a duplicate that adds a name has no repair", func(t *testing.T) {
		source := "module x\nimport a { fn f }\nimport a { fn f  fn g }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n"
		parsed := syntax.Parse([]byte(source))
		if got := phase30Codes(parsed.Diagnostics); len(got) != 1 || got[0] != "module.duplicate_import" {
			t.Fatalf("codes = %v", got)
		}
		if len(parsed.Diagnostics[0].Repairs) != 0 {
			t.Fatalf("repairs = %+v, want none", parsed.Diagnostics[0].Repairs)
		}
	})

	t.Run("no other import refusal carries a repair", func(t *testing.T) {
		for _, source := range []string{
			"module x\nimport a {}\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n",
			"module x\nimport a { fn f, fn g }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n",
			"module x\nimport ..a { fn f }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n",
		} {
			for _, problem := range syntax.Parse([]byte(source)).Diagnostics {
				if len(problem.Repairs) != 0 {
					t.Fatalf("%q: %s carries repairs %+v", source, problem.Code, problem.Repairs)
				}
			}
		}
	})
}

// phase30GeneratedImportProgram builds one program with 0..4 import
// declarations of 1..3 items in a random (not alphabetical) order, with mixed
// spacing, newlines and line comments at item and declaration boundaries.
func phase30GeneratedImportProgram(random *rand.Rand) []byte {
	modules := []string{"zeta", "alpha.core", "mid_module", "b.c.d", "omega"}
	random.Shuffle(len(modules), func(left, right int) { modules[left], modules[right] = modules[right], modules[left] })
	names := []string{"run", "adler32", "digest", "x", "next_value"}
	gap := func() string {
		switch random.Intn(4) {
		case 0:
			return " "
		case 1:
			return "\n"
		case 2:
			return "  \n\n  "
		default:
			return " // note\n"
		}
	}
	var out strings.Builder
	out.WriteString("module gen.phase30" + gap())
	for index := 0; index < random.Intn(5); index++ {
		out.WriteString("import" + gap()[:1] + modules[index] + " {" + gap())
		random.Shuffle(len(names), func(left, right int) { names[left], names[right] = names[right], names[left] })
		for item := 0; item <= random.Intn(3); item++ {
			kind := "fn"
			if random.Intn(5) == 0 {
				kind = "type"
			}
			out.WriteString(kind + " " + names[item] + gap())
		}
		out.WriteString("}" + gap())
	}
	out.WriteString("export {" + gap() + "fn main" + gap() + "}\nfn main(n: U64) -> U64 { n }\n")
	return []byte(out.String())
}

func phase30ImportShape(program ast.Program) []string {
	var shape []string
	for _, declaration := range program.Imports {
		line := declaration.Module + ":"
		for _, item := range declaration.Items {
			line += " " + item.Kind + " " + item.Name
		}
		shape = append(shape, line)
	}
	return shape
}

// TestPhase30GeneratedImportRoundTrips runs the idempotence, lossless-CST and
// semantic-token lanes over 200 deterministic generated import programs.
func TestPhase30GeneratedImportRoundTrips(t *testing.T) {
	random := rand.New(rand.NewSource(0x30c0de))
	withImports := 0
	for caseID := 0; caseID < 200; caseID++ {
		source := phase30GeneratedImportProgram(random)
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatalf("case %d lost CST bytes: %q", caseID, source)
		}
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("case %d did not parse: %q diagnostics=%v", caseID, source, diagnosticIdentity(parsed.Diagnostics))
		}
		if len(parsed.Program.Imports) > 0 {
			withImports++
		}
		canonical := syntax.Format(parsed.Tree)
		reparsed := syntax.Parse(canonical)
		if len(reparsed.Diagnostics) != 0 {
			t.Fatalf("case %d canonical form does not reparse: %q diagnostics=%v", caseID, canonical, diagnosticIdentity(reparsed.Diagnostics))
		}
		if !bytes.Equal(canonical, syntax.Format(reparsed.Tree)) {
			t.Fatalf("case %d is not a fixed point:\nfirst=%q\nsecond=%q", caseID, canonical, syntax.Format(reparsed.Tree))
		}
		if !reflect.DeepEqual(semanticTokens(parsed.Tree), semanticTokens(reparsed.Tree)) {
			t.Fatalf("case %d changed token identity\nsource=%q\ncanonical=%q", caseID, source, canonical)
		}
		if !reflect.DeepEqual(phase30ImportShape(parsed.Program), phase30ImportShape(reparsed.Program)) {
			t.Fatalf("case %d changed import order or items\nsource=%q\ncanonical=%q", caseID, source, canonical)
		}
		if !reflect.DeepEqual(comments(parsed.Tree), comments(reparsed.Tree)) {
			t.Fatalf("case %d changed comments\nsource=%q\ncanonical=%q", caseID, source, canonical)
		}
	}
	if withImports < 100 {
		t.Fatalf("generator produced imports in only %d of 200 programs", withImports)
	}
}

func FuzzPhase30ImportParseFormat(f *testing.F) {
	f.Add([]byte(phase30Tracer))
	f.Add([]byte("module m\nimport a {}\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n"))
	f.Add([]byte("module m\nexport { fn main }\nimport a { fn f }\nfn main(n: U64) -> U64 { n }\n"))
	f.Add([]byte("module m\nimport a { fn f }\nimport a { fn f }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n"))
	f.Add([]byte("module m\nimport b.c {\n  type T // c\n  fn g\n}\nimport a { fn f }\nexport { fn main }\nfn main(n: U64) -> U64 { n }\n"))
	f.Fuzz(func(t *testing.T, source []byte) {
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatal("CST round trip changed fuzz input")
		}
		if len(parsed.Diagnostics) > 21 {
			t.Fatalf("unbounded diagnostics: %d", len(parsed.Diagnostics))
		}
		if len(parsed.Diagnostics) != 0 {
			return
		}
		first := syntax.Format(parsed.Tree)
		formatted := syntax.Parse(first)
		if len(formatted.Diagnostics) != 0 {
			t.Fatalf("formatter produced invalid source: %q diagnostics=%v", first, diagnosticIdentity(formatted.Diagnostics))
		}
		if second := syntax.Format(formatted.Tree); !bytes.Equal(first, second) {
			t.Fatalf("formatter is not idempotent:\nfirst=%q\nsecond=%q", first, second)
		}
	})
}

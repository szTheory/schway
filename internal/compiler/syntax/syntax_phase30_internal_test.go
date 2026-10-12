package syntax

import (
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

func phase30Program(imports string) []byte {
	return []byte("module m\n" + imports + "export { fn main }\nfn main(n: U64) -> U64 { n }\n")
}

func phase30HasCode(diagnostics []diagnostic.Diagnostic, code string) int {
	count := 0
	for _, problem := range diagnostics {
		if problem.Code == code {
			count++
		}
	}
	return count
}

func phase30Syntax(diagnostics []diagnostic.Diagnostic) bool {
	for _, problem := range diagnostics {
		if strings.HasPrefix(problem.Code, "syntax.") {
			return true
		}
	}
	return false
}

// TestPhase30ImportPathSyntaxRefused pins T-30-04: the shapes `..x`, `/etc/x`,
// `a/b` and `"a"` are unrepresentable as module paths, refused by the lexer or
// parser before any filesystem access. Each of the first three has its own
// armed control: admitModulePathFormForTest engages exactly that one shape and
// the refusal must disappear, so no refusal rests on a sibling accepted fixture.
func TestPhase30ImportPathSyntaxRefused(t *testing.T) {
	if admitModulePathFormForTest != "" {
		t.Fatalf("production seam value = %q, want empty", admitModulePathFormForTest)
	}
	for _, test := range []struct {
		name, form, path, code string
	}{
		{"dotdot", "dotdot", "..x", "syntax.expected_module_name"},
		{"absolute", "absolute", "/etc/x", "syntax.unexpected_byte"},
		{"slash", "slash", "a/b", "syntax.unexpected_byte"},
		{"quoted", "", `"a"`, "syntax.expected_module_name"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := phase30Program("import " + test.path + " { fn x }\n")
			refused := Parse(source)
			if phase30HasCode(refused.Diagnostics, test.code) == 0 {
				t.Fatalf("seam off: codes = %v, want %s", codesOf(refused.Diagnostics), test.code)
			}
			if !phase30Syntax(refused.Diagnostics) {
				t.Fatalf("seam off: no syntax.* refusal in %v", codesOf(refused.Diagnostics))
			}
			if test.form == "" {
				return
			}
			func() {
				admitModulePathFormForTest = test.form
				defer func() { admitModulePathFormForTest = "" }()
				armed := Parse(source)
				if len(armed.Diagnostics) != 0 {
					t.Fatalf("form %q armed: diagnostics = %v, want none", test.form, codesOf(armed.Diagnostics))
				}
				if len(armed.Program.Imports) != 1 || armed.Program.Imports[0].Module != test.path {
					t.Fatalf("form %q armed: imports = %+v, want module %q", test.form, armed.Program.Imports, test.path)
				}
			}()
			if again := Parse(source); phase30HasCode(again.Diagnostics, test.code) == 0 {
				t.Fatalf("seam not restored: codes = %v", codesOf(again.Diagnostics))
			}
		})
	}
	t.Run("a form admits only its own shape", func(t *testing.T) {
		admitModulePathFormForTest = "slash"
		defer func() { admitModulePathFormForTest = "" }()
		if parsed := Parse(phase30Program("import /etc/x { fn x }\n")); len(parsed.Diagnostics) == 0 {
			t.Fatalf("slash form also admitted an absolute path")
		}
		if parsed := Parse(phase30Program("import ..x { fn x }\n")); len(parsed.Diagnostics) == 0 {
			t.Fatalf("slash form also admitted a dotdot path")
		}
	})
}

func codesOf(diagnostics []diagnostic.Diagnostic) []string {
	result := make([]string, len(diagnostics))
	for index, problem := range diagnostics {
		result[index] = problem.Code
	}
	return result
}

func phase30Imports(count int) string {
	var out strings.Builder
	for index := 0; index < count; index++ {
		fmt.Fprintf(&out, "import m%d { fn f }\n", index)
	}
	return out.String()
}

// TestPhase30ImportsLimitParse pins D-13/D-25: 16 imports (duplicates
// counted) parse, the 17th is refused at parse time, and the exceeds seam is
// observed to fail the exact-N row.
func TestPhase30ImportsLimitParse(t *testing.T) {
	if got := Parse(phase30Program(phase30Imports(16))); phase30HasCode(got.Diagnostics, "module.imports_limit") != 0 || len(got.Program.Imports) != 16 {
		t.Fatalf("16 imports: codes = %v imports = %d", codesOf(got.Diagnostics), len(got.Program.Imports))
	}
	source := phase30Program(phase30Imports(17))
	got := Parse(source)
	if phase30HasCode(got.Diagnostics, "module.imports_limit") != 1 {
		t.Fatalf("17 imports: codes = %v, want exactly one module.imports_limit", codesOf(got.Diagnostics))
	}
	seventeenth := strings.Index(string(source), "import m16")
	for _, problem := range got.Diagnostics {
		if problem.Code == "module.imports_limit" && problem.Primary.Start != seventeenth {
			t.Fatalf("limit primary = %+v, want the 17th declaration at %d", problem.Primary, seventeenth)
		}
	}
	t.Run("duplicates are counted", func(t *testing.T) {
		duplicates := strings.Repeat("import same { fn f }\n", 17)
		if phase30HasCode(Parse(phase30Program(duplicates)).Diagnostics, "module.imports_limit") != 1 {
			t.Fatalf("17 duplicate imports were not counted against the cap")
		}
	})
	t.Run("off-by-one control fails the exact-N row", func(t *testing.T) {
		exceedsOffByOneForTest = true
		defer func() { exceedsOffByOneForTest = false }()
		if phase30HasCode(Parse(phase30Program(phase30Imports(16))).Diagnostics, "module.imports_limit") == 0 {
			t.Fatalf("with the seam engaged, 16 imports must be refused (the exact-N row fails)")
		}
	})
}

package modload

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// writeModules writes files (slash-separated relative path to content) under a
// fresh TempDir and returns the directory. Fixtures are generated, never
// committed (D-27).
func writeModules(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// loadDir loads dir/main.schway.
func loadDir(t *testing.T, dir string) (Set, []diagnostic.Diagnostic) {
	t.Helper()
	set, diags, err := Load(filepath.Join(dir, "main.schway"))
	if err != nil {
		t.Fatal(err)
	}
	return set, diags
}

func codesOf(diags []diagnostic.Diagnostic) []string {
	codes := make([]string, 0, len(diags))
	for _, d := range diags {
		codes = append(codes, d.Code)
	}
	return codes
}

func hasCode(diags []diagnostic.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func findCode(diags []diagnostic.Diagnostic, code string) (diagnostic.Diagnostic, bool) {
	for _, d := range diags {
		if d.Code == code {
			return d, true
		}
	}
	return diagnostic.Diagnostic{}, false
}

const (
	entryHeader = "module app.main\n\n"
	entryExport = "export {\n  fn main\n}\n\n"
)

// entryUsing builds an entry that imports each named module's `fn f` and calls
// nothing, so import resolution is the only thing under test.
func entryUsing(modules ...string) string {
	var out strings.Builder
	out.WriteString(entryHeader)
	for _, module := range modules {
		out.WriteString("import " + module + " {\n  fn f\n}\n\n")
	}
	out.WriteString(entryExport)
	out.WriteString("fn main(n: U64) -> U64 {\n  n\n}\n")
	return out.String()
}

// libSource builds a module named name exporting fn f and importing each of
// imports.
func libSource(name string, imports ...string) string {
	var out strings.Builder
	out.WriteString("module " + name + "\n\n")
	for _, module := range imports {
		out.WriteString("import " + module + " {\n  fn f\n}\n\n")
	}
	out.WriteString("export {\n  fn f\n}\n\nfn f(n: U64) -> U64 {\n  n\n}\n")
	return out.String()
}

func TestPhase30LoadOneImport(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"main.schway": "module app.main\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let r = twice(n)\n  r\n}\n",
		"lib.schway":  "module lib\n\nexport {\n  fn twice\n}\n\nfn twice(n: U64) -> U64 {\n  n\n}\n",
	})
	set, diags := loadDir(t, dir)
	if len(diags) != 0 {
		t.Fatalf("diagnostics: %+v", diags)
	}
	type fn struct{ name, module string }
	var funcs []fn
	for _, f := range set.Program.Funcs {
		funcs = append(funcs, fn{f.Name, f.Module})
	}
	if want := []fn{{"main", ""}, {"twice", "lib"}}; !reflect.DeepEqual(funcs, want) {
		t.Fatalf("funcs = %v, want %v", funcs, want)
	}
	var units []string
	for _, unit := range set.Program.Units {
		units = append(units, unit.Name)
	}
	if want := []string{"", "lib"}; !reflect.DeepEqual(units, want) {
		t.Fatalf("units = %q, want %q", units, want)
	}
	if len(set.Trees) != 2 {
		t.Fatalf("trees = %d, want 2", len(set.Trees))
	}
	if want := len(set.Trees[""].Tokens) + len(set.Trees["lib"].Tokens); set.Work != want || want == 0 {
		t.Fatalf("work = %d, want %d", set.Work, want)
	}
	if len(set.Sources) != 2 || set.Sources[0].Module != "" || set.Sources[1].Module != "lib" {
		t.Fatalf("sources = %+v", set.Sources)
	}
}

// TestPhase30NoImportLoadIsSingleFileParse pins SC4/D-15: an entry with no
// import declaration loads as exactly its single-file parse.
func TestPhase30NoImportLoadIsSingleFileParse(t *testing.T) {
	checked := 0
	for _, root := range []string{"../../../examples", "../../../testdata"} {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".schway") {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			want := syntax.Parse(data)
			for _, token := range want.Tree.Tokens {
				if token.Kind == syntax.TokenImport {
					return nil
				}
			}
			set, diags, err := Load(path)
			if err != nil {
				t.Errorf("%s: %v", path, err)
				return nil
			}
			checked++
			if !reflect.DeepEqual(set.Program, want.Program) {
				t.Errorf("%s: program differs from the single-file parse", path)
			}
			if !reflect.DeepEqual(diags, want.Diagnostics) {
				t.Errorf("%s: diagnostics differ from the single-file parse", path)
			}
			if !reflect.DeepEqual(set.Trees[""].Tokens, want.Tree.Tokens) || len(set.Trees) != 1 {
				t.Errorf("%s: tree differs from the single-file parse", path)
			}
			if set.Program.Units != nil {
				t.Errorf("%s: Units must be nil for an import-free program", path)
			}
			if set.Work != len(want.Tree.Tokens) {
				t.Errorf("%s: work = %d, want %d", path, set.Work, len(want.Tree.Tokens))
			}
			for _, function := range set.Program.Funcs {
				if function.Module != "" {
					t.Errorf("%s: FuncDecl.Module = %q, want empty", path, function.Module)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if checked < 20 {
		t.Fatalf("only %d import-free programs checked", checked)
	}
}

var _ = ast.Program{}

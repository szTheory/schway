package modload

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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

// readCounts installs a read observer and returns the live counter plus the
// restore func.
func readCounts() (map[string]int, func()) {
	counts := map[string]int{}
	return counts, SetReadObserverForTest(func(module string) { counts[module]++ })
}

func importCycles(diags []diagnostic.Diagnostic) []diagnostic.Diagnostic {
	var out []diagnostic.Diagnostic
	for _, d := range diags {
		if d.Code == "module.import_cycle" {
			out = append(out, d)
		}
	}
	return out
}

func TestPhase30ImportCycles(t *testing.T) {
	rows := []struct {
		name    string
		files   map[string]string
		message string
		unit    string
	}{
		{"self", map[string]string{"main.schway": entryUsing("a"), "a.schway": libSource("a", "a")}, "import cycle: a -> a", "a"},
		{"length-2", map[string]string{"main.schway": entryUsing("a"), "a.schway": libSource("a", "b"), "b.schway": libSource("b", "a")}, "import cycle: a -> b -> a", "b"},
		{"length-3", map[string]string{"main.schway": entryUsing("a"), "a.schway": libSource("a", "b"), "b.schway": libSource("b", "c"), "c.schway": libSource("c", "a")}, "import cycle: a -> b -> c -> a", "c"},
		{"length-3-entered-at-b", map[string]string{"main.schway": entryUsing("b"), "a.schway": libSource("a", "b"), "b.schway": libSource("b", "c"), "c.schway": libSource("c", "a")}, "import cycle: a -> b -> c -> a", "a"},
		{"length-3-entered-at-c", map[string]string{"main.schway": entryUsing("c"), "a.schway": libSource("a", "b"), "b.schway": libSource("b", "c"), "c.schway": libSource("c", "a")}, "import cycle: a -> b -> c -> a", "b"},
		{"via-entry", map[string]string{"main.schway": entryUsing("a"), "a.schway": libSource("a", "main")}, "import cycle: a -> app.main -> a", "a"},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			_, diags := loadDir(t, writeModules(t, row.files))
			cycles := importCycles(diags)
			if len(cycles) != 1 || len(diags) != 1 {
				t.Fatalf("diagnostics = %v, want exactly one module.import_cycle", codesOf(diags))
			}
			if cycles[0].Message != row.message {
				t.Fatalf("message = %q, want %q", cycles[0].Message, row.message)
			}
			if cycles[0].Unit != row.unit {
				t.Fatalf("unit = %q, want the importer %q", cycles[0].Unit, row.unit)
			}
			if len(cycles[0].Causes) == 0 {
				t.Fatalf("cycle carries no per-hop causes")
			}
		})
	}

	t.Run("diamond-is-not-a-cycle-and-the-visited-mutant-refuses-it", func(t *testing.T) {
		files := map[string]string{
			"main.schway": entryUsing("a", "b"), "a.schway": libSource("a", "c"),
			"b.schway": libSource("b", "c"), "c.schway": libSource("c"),
		}
		dir := writeModules(t, files)
		if _, diags := loadDir(t, dir); len(diags) != 0 {
			t.Fatalf("diamond refused: %v", codesOf(diags))
		}
		visitedInsteadOfColorForTest = true
		defer func() { visitedInsteadOfColorForTest = false }()
		if _, diags := loadDir(t, dir); len(importCycles(diags)) == 0 {
			t.Fatalf("visitedInsteadOfColorForTest did not falsely refuse the diamond")
		}
	})
}

func TestPhase30GraphRefusals(t *testing.T) {
	good := map[string]string{"main.schway": entryUsing("lib"), "lib.schway": libSource("lib")}
	with := func(name, content string) map[string]string {
		files := map[string]string{}
		for key, value := range good {
			files[key] = value
		}
		files[name] = content
		return files
	}
	t.Run("sibling-accepted", func(t *testing.T) {
		if _, diags := loadDir(t, writeModules(t, good)); len(diags) != 0 {
			t.Fatalf("accepted fixture refused: %v", codesOf(diags))
		}
	})

	armed := func(t *testing.T, code string, files map[string]string, check func(t *testing.T, d diagnostic.Diagnostic, files map[string]string)) {
		t.Helper()
		dir := writeModules(t, files)
		_, diags := loadDir(t, dir)
		d, ok := findCode(diags, code)
		if !ok {
			t.Fatalf("%s not observed; got %v", code, codesOf(diags))
		}
		if check != nil {
			check(t, d, files)
		}
		skipGraphCheckForTest = code
		defer func() { skipGraphCheckForTest = "" }()
		if _, diags := loadDir(t, dir); hasCode(diags, code) {
			t.Fatalf("%s still observed with skipGraphCheckForTest engaged", code)
		}
	}

	t.Run("not_found", func(t *testing.T) {
		files := map[string]string{"main.schway": entryUsing("missing")}
		armed(t, "module.not_found", files, func(t *testing.T, d diagnostic.Diagnostic, files map[string]string) {
			if d.Unit != "" {
				t.Fatalf("unit = %q, want the entry importer", d.Unit)
			}
			source := files["main.schway"]
			start := strings.Index(source, "missing")
			if d.Primary.Start != start || d.Primary.End != start+len("missing") {
				t.Fatalf("span = %+v, want the import's module path at %d", d.Primary, start)
			}
		})
	})
	t.Run("path_mismatch", func(t *testing.T) {
		files := with("lib.schway", strings.Replace(libSource("lib"), "module lib", "module other", 1))
		armed(t, "module.path_mismatch", files, func(t *testing.T, d diagnostic.Diagnostic, files map[string]string) {
			if d.Unit != "" || len(d.Causes) != 1 || d.Causes[0].Unit != "lib" || d.Causes[0].Span == nil {
				t.Fatalf("attribution = %+v", d)
			}
			span := d.Causes[0].Span
			if got := files["lib.schway"][span.Start:span.End]; got != "other" {
				t.Fatalf("cause span covers %q, want the header name", got)
			}
		})
	})
	t.Run("duplicate", func(t *testing.T) {
		files := map[string]string{
			"main.schway": strings.Replace(entryUsing("lib"), "module app.main", "module lib", 1),
			"lib.schway":  libSource("lib"),
		}
		armed(t, "module.duplicate", files, func(t *testing.T, d diagnostic.Diagnostic, files map[string]string) {
			if d.Unit != "" {
				t.Fatalf("unit = %q, want the importer", d.Unit)
			}
		})
	})
	t.Run("entry-header-is-never-checked-against-its-path", func(t *testing.T) {
		if _, diags := loadDir(t, writeModules(t, good)); hasCode(diags, "module.path_mismatch") {
			t.Fatalf("entry header app.main was checked against main.schway")
		}
	})
	for _, name := range []string{"_x", "a__b", "b_"} {
		t.Run("invalid_import_name/"+name, func(t *testing.T) {
			source := strings.Replace(entryUsing("lib"), "fn f", "fn "+name, 1)
			_, diags := loadDir(t, writeModules(t, with("main.schway", source)))
			d, ok := findCode(diags, "module.invalid_import_name")
			if !ok {
				t.Fatalf("not refused; got %v", codesOf(diags))
			}
			start := strings.Index(source, "fn "+name)
			if d.Primary.Start > start+3 || d.Primary.End < start+3+len(name) {
				t.Fatalf("span %+v does not cover the item %q at %d", d.Primary, name, start)
			}
		})
	}
	t.Run("imported_data_unsupported", func(t *testing.T) {
		source := libSource("lib") + "\ndata Thing =\n  | One\n"
		_, diags := loadDir(t, writeModules(t, with("lib.schway", source)))
		d, ok := findCode(diags, "module.imported_data_unsupported")
		if !ok || d.Unit != "lib" {
			t.Fatalf("got %+v, want imported_data_unsupported in lib", diags)
		}
	})
	t.Run("imported_foreign_unsupported", func(t *testing.T) {
		source := libSource("lib") + "\nforeign C {\n  fn c_one(n: U64) -> U64 {\n    unwind: forbidden\n    nonlocal_exit: forbidden\n    mode: \"pure\"\n    abi: c_one_fn\n  }\n}\n"
		_, diags := loadDir(t, writeModules(t, with("lib.schway", source)))
		d, ok := findCode(diags, "module.imported_foreign_unsupported")
		if !ok || d.Unit != "lib" {
			t.Fatalf("got %+v, want imported_foreign_unsupported in lib", diags)
		}
	})
	t.Run("parse-error-unit", func(t *testing.T) {
		broken := strings.Replace(libSource("lib"), "export {\n  fn f\n}", "export {\n  fn f\n", 1)
		dir := writeModules(t, with("lib.schway", broken))
		_, diags := loadDir(t, dir)
		var syntaxDiag *diagnostic.Diagnostic
		for index := range diags {
			if strings.HasPrefix(diags[index].Code, "syntax.") {
				syntaxDiag = &diags[index]
				break
			}
		}
		if syntaxDiag == nil || syntaxDiag.Unit != "lib" {
			t.Fatalf("parse diagnostics = %+v, want Unit lib", diags)
		}
		restore := SetSkipParseUnitStampForTest(true)
		defer restore()
		_, diags = loadDir(t, dir)
		for _, d := range diags {
			if strings.HasPrefix(d.Code, "syntax.") && d.Unit != "" {
				t.Fatalf("Unit = %q with the stamp skipped, want empty", d.Unit)
			}
		}
	})
	t.Run("not_found-text-leak-control", func(t *testing.T) {
		dir := writeModules(t, map[string]string{"main.schway": entryUsing("missing")})
		_, diags := loadDir(t, dir)
		d, _ := findCode(diags, "module.not_found")
		if strings.Contains(d.Message, dir) {
			t.Fatalf("production message leaks the absolute path: %q", d.Message)
		}
		restore := SetFormatErrorTextForTest(true)
		defer restore()
		_, diags = loadDir(t, dir)
		d, _ = findCode(diags, "module.not_found")
		if !strings.Contains(d.Message, dir) {
			t.Fatalf("formatErrorTextForTest did not append the candidate path: %q", d.Message)
		}
	})
}

func diamondFixture() map[string]string {
	return map[string]string{
		"main.schway": entryUsing("a", "b"), "a.schway": libSource("a", "c"),
		"b.schway": libSource("b", "c"), "c.schway": libSource("c"),
	}
}

func TestPhase30DiamondLoadsOnce(t *testing.T) {
	dir := writeModules(t, diamondFixture())
	counts, restore := readCounts()
	set, diags := loadDir(t, dir)
	restore()
	if len(diags) != 0 {
		t.Fatalf("diagnostics: %v", codesOf(diags))
	}
	if counts["c"] != 1 {
		t.Fatalf("c read %d times, want 1", counts["c"])
	}
	inC := 0
	for _, function := range set.Program.Funcs {
		if function.Module == "c" {
			inC++
		}
	}
	if inC != 1 || len(set.Program.Units) != 4 {
		t.Fatalf("c functions = %d, units = %d; want 1 and 4", inC, len(set.Program.Units))
	}

	t.Run("diamondTwice-control", func(t *testing.T) {
		counts, restoreObserver := readCounts()
		defer restoreObserver()
		restoreSeam := SetDiamondTwiceForTest(true)
		defer restoreSeam()
		loadDir(t, dir)
		if counts["c"] != 2 {
			t.Fatalf("c read %d times with the mutant engaged, want 2", counts["c"])
		}
	})
}

func TestPhase30LoaderReadsEachFileOnce(t *testing.T) {
	fixtures := map[string]map[string]string{
		"diamond":   diamondFixture(),
		"via-entry": {"main.schway": entryUsing("a"), "a.schway": libSource("a", "main")},
	}
	for name, files := range fixtures {
		t.Run(name, func(t *testing.T) {
			counts, restore := readCounts()
			defer restore()
			loadDir(t, writeModules(t, files))
			if counts[""] != 1 {
				t.Fatalf("entry read %d times, want 1", counts[""])
			}
			if len(counts) != len(files) {
				t.Fatalf("observed %v for %d files", counts, len(files))
			}
			for module, count := range counts {
				if count != 1 {
					t.Fatalf("module %q read %d times, want 1", module, count)
				}
			}
		})
	}
	t.Run("reread-control", func(t *testing.T) {
		counts, restore := readCounts()
		defer restore()
		rereadForTest = true
		defer func() { rereadForTest = false }()
		loadDir(t, writeModules(t, fixtures["diamond"]))
		twice := false
		for module, count := range counts {
			if module != "" && count == 2 {
				twice = true
			}
		}
		if !twice {
			t.Fatalf("rereadForTest produced counts %v, want some imported module at 2", counts)
		}
	})
}

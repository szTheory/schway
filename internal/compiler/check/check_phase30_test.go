package check

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// mergeForTest assembles the merged-program shape plan 30-04 builds, without
// importing modload: entry functions first with Module "", then each library
// (sorted by name) with every function's Module set; Units[0] is the entry.
func mergeForTest(t *testing.T, entry []byte, libs map[string][]byte) ast.Program {
	t.Helper()
	parse := func(source []byte) ast.Program {
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("fixture failed to parse: %+v\n%s", parsed.Diagnostics, source)
		}
		return parsed.Program
	}
	merged := parse(entry)
	merged.Units = []ast.Unit{{Name: "", Imports: merged.Imports, Exports: merged.Exports}}
	names := make([]string, 0, len(libs))
	for name := range libs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		library := parse(libs[name])
		for _, function := range library.Funcs {
			function.Module = name
			merged.Funcs = append(merged.Funcs, function)
		}
		merged.Units = append(merged.Units, ast.Unit{Name: name, Imports: library.Imports, Exports: library.Exports})
	}
	return merged
}

const (
	phase30LibTwice = "module lib\n\nexport {\n  fn twice\n}\n\nfn twice(value: Byte) -> Byte {\n  value\n}\n"
	phase30EntryApp = "module app.main\n\nimport lib {\n  fn twice\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let result = twice(value)\n  result\n}\n"
)

func phase30Codes(diagnostics []diagnostic.Diagnostic) []string {
	codes := make([]string, 0, len(diagnostics))
	for _, d := range diagnostics {
		codes = append(codes, d.Code)
	}
	return codes
}

func phase30JSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func phase30FunctionByID(t *testing.T, program core.Program, id string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.ID == id {
			return function
		}
	}
	t.Fatalf("no core function %s in %v", id, func() []string {
		ids := []string{}
		for _, function := range program.Functions {
			ids = append(ids, function.ID)
		}
		return ids
	}())
	return core.Function{}
}

// TestPhase30CheckMergedProgram is the tracer: an entry calling an imported
// exported function checks clean into a core program carrying module-qualified
// IDs and the module facts corevalidate re-derives visibility from.
func TestPhase30CheckMergedProgram(t *testing.T) {
	result := Program(mergeForTest(t, []byte(phase30EntryApp), map[string][]byte{"lib": []byte(phase30LibTwice)}))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected a clean check, got %+v", result.Diagnostics)
	}
	main := phase30FunctionByID(t, result.Program, "s1:app.main:fn:main")
	twice := phase30FunctionByID(t, result.Program, "s1:lib:fn:twice")
	if twice.Module != "lib" || !twice.Exported {
		t.Fatalf("twice must carry Module lib and Exported, got module=%q exported=%v", twice.Module, twice.Exported)
	}
	if main.Module != "" || main.Exported {
		t.Fatalf("the entry main must carry no module facts, got module=%q exported=%v", main.Module, main.Exported)
	}
	want := []core.ModuleImport{{From: "", Module: "lib", Name: "twice"}}
	if got := phase30JSON(t, result.Program.Imports); got != phase30JSON(t, want) {
		t.Fatalf("Program.Imports = %s, want %s", got, phase30JSON(t, want))
	}
	var callee string
	for _, operation := range main.Linear.Operations {
		if operation.Kind == core.OpCall {
			callee = operation.CalleeID
		}
	}
	if callee != "s1:lib:fn:twice" {
		t.Fatalf("main's OpCall CalleeID = %q, want s1:lib:fn:twice", callee)
	}
	t.Run("a private function stays out of the export facts", func(t *testing.T) {
		lib := phase30LibTwice + "\nfn hidden(value: Byte) -> Byte {\n  value\n}\n"
		got := Program(mergeForTest(t, []byte(phase30EntryApp), map[string][]byte{"lib": []byte(lib)}))
		if len(got.Diagnostics) != 0 {
			t.Fatalf("expected a clean check, got %+v", got.Diagnostics)
		}
		if hidden := phase30FunctionByID(t, got.Program, "s1:lib:fn:hidden"); hidden.Exported || hidden.Module != "lib" {
			t.Fatalf("hidden must be Module lib and not Exported, got %+v", hidden)
		}
	})
}

// TestPhase30CheckRefusesUnloadedImports pins D-22: check.Program fails closed
// on an import-bearing program that no loader resolved.
func TestPhase30CheckRefusesUnloadedImports(t *testing.T) {
	parsed := syntax.Parse([]byte(phase30EntryApp))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse: %+v", parsed.Diagnostics)
	}
	got := Program(parsed.Program)
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "module.imports_not_loaded" {
		t.Fatalf("expected exactly module.imports_not_loaded, got %+v", got.Diagnostics)
	}
	if want := parsed.Program.Imports[0].Span; got.Diagnostics[0].Primary != want {
		t.Fatalf("primary span = %+v, want the first import declaration %+v", got.Diagnostics[0].Primary, want)
	}

	restore := SetIgnoreUnloadedImportsForTest(true)
	defer restore()
	controlled := Program(parsed.Program)
	if len(controlled.Diagnostics) == 0 {
		t.Fatal("with the guard skipped the call must still be unresolved")
	}
	for _, d := range controlled.Diagnostics {
		if d.Code == "module.imports_not_loaded" {
			t.Fatalf("the control must skip the guard, got %+v", controlled.Diagnostics)
		}
	}
	restore()
	if again := Program(parsed.Program); len(again.Diagnostics) != 1 || again.Diagnostics[0].Code != "module.imports_not_loaded" {
		t.Fatalf("restore must re-arm the guard, got %+v", again.Diagnostics)
	}
}

func phase30Entry(imports string) []byte {
	return []byte("module app.main\n\n" + imports + "\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  value\n}\n")
}

func phase30Lib(exports string, functions ...string) []byte {
	return []byte("module lib\n\nexport {\n" + exports + "}\n\n" + strings.Join(functions, "\n"))
}

func phase30Fn(name string) string {
	return "fn " + name + "(value: Byte) -> Byte {\n  value\n}\n"
}

// TestPhase30PrivateNameRefused pins MOD-04 / SC2: importing a name the target
// does not export is refused at the item with a hint cause in the target, and
// no repair.
func TestPhase30PrivateNameRefused(t *testing.T) {
	libs := map[string][]byte{"lib": phase30Lib("  fn shown\n", phase30Fn("shown"), phase30Fn("hidden"))}
	entry := []byte(phase30EntryHidden)
	program := mergeForTest(t, entry, libs)
	got := Program(program)
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "module.not_exported" {
		t.Fatalf("expected one module.not_exported, got %+v", got.Diagnostics)
	}
	d := got.Diagnostics[0]
	if d.Unit != "" {
		t.Fatalf("the diagnostic belongs to the importer (entry), got unit %q", d.Unit)
	}
	item := program.Imports[0].Items[0]
	if d.Primary != item.Span {
		t.Fatalf("primary = %+v, want the import item %+v", d.Primary, item.Span)
	}
	if len(d.Causes) != 1 || d.Causes[0].Kind != "export_list" || d.Causes[0].Unit != "lib" {
		t.Fatalf("expected one export_list cause in lib, got %+v", d.Causes)
	}
	if len(d.Repairs) != 0 {
		t.Fatalf("no repair is offered, got %+v", d.Repairs)
	}

	t.Run("control: admitting the private import is observable", func(t *testing.T) {
		defer func() { admitPrivateImportForTest = false }()
		admitPrivateImportForTest = true
		admitted := Program(mergeForTest(t, entry, libs))
		if len(admitted.Diagnostics) != 0 {
			t.Fatalf("the control must admit the private import, got %+v", admitted.Diagnostics)
		}
	})

	t.Run("empty export block", func(t *testing.T) {
		empty := map[string][]byte{"lib": []byte("module lib\n\nexport {\n}\n\n" + phase30Fn("hidden"))}
		out := Program(mergeForTest(t, entry, empty))
		if len(out.Diagnostics) != 1 || out.Diagnostics[0].Code != "module.not_exported" {
			t.Fatalf("expected module.not_exported for an empty export block, got %+v", out.Diagnostics)
		}
		if len(out.Diagnostics[0].Causes) != 1 || out.Diagnostics[0].Causes[0].Span != nil {
			t.Fatalf("an empty export block has no span to point at, got %+v", out.Diagnostics[0].Causes)
		}
	})

	t.Run("names are byte exact", func(t *testing.T) {
		exact := map[string][]byte{"lib": phase30Lib("  fn foo\n", phase30Fn("foo"))}
		out := Program(mergeForTest(t, phase30Entry("import lib {\n  fn Foo\n}\n"), exact))
		if len(out.Diagnostics) != 1 || out.Diagnostics[0].Code != "module.not_exported" {
			t.Fatalf("expected module.not_exported for Foo vs foo, got %+v", out.Diagnostics)
		}
	})

	t.Run("a private name cannot be called by its bare name", func(t *testing.T) {
		caller := []byte("module app.main\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let result = hidden(value)\n  result\n}\n")
		out := Program(mergeForTest(t, caller, libs))
		if len(out.Diagnostics) == 0 {
			t.Fatal("calling another module's private function by bare name must be refused")
		}
		for _, d := range out.Diagnostics {
			if d.Unit != "" {
				t.Fatalf("the refusal belongs to the caller (entry), got unit %q", d.Unit)
			}
		}
	})

	t.Run("same private name in two modules coexists", func(t *testing.T) {
		both := map[string][]byte{"lib": phase30Lib("  fn shown\n", phase30Fn("shown"), phase30Fn("helper"))}
		entryHelper := []byte("module app.main\n\nimport lib {\n  fn shown\n}\n\nexport {\n  fn main\n}\n\n" + phase30Fn("helper") + "\nfn main(value: Byte) -> Byte {\n  value\n}\n")
		out := Program(mergeForTest(t, entryHelper, both))
		if len(out.Diagnostics) != 0 {
			t.Fatalf("two private helpers must coexist, got %+v", out.Diagnostics)
		}
		a := phase30FunctionByID(t, out.Program, "s1:app.main:fn:helper")
		b := phase30FunctionByID(t, out.Program, "s1:lib:fn:helper")
		if a.ID == b.ID {
			t.Fatalf("the helpers must have distinct IDs, got %s", a.ID)
		}
	})
}

const phase30EntryHidden = "module app.main\n\nimport lib {\n  fn hidden\n}\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  value\n}\n"

// TestPhase30ImportCollisions pins D-04 / D-20 / D-02: each collision shape is
// refused by name, names both origins, and carries no repair.
func TestPhase30ImportCollisions(t *testing.T) {
	libF := phase30Lib("  fn f\n", phase30Fn("f"))
	libA := []byte("module a\n\nexport {\n  fn f\n}\n\n" + phase30Fn("f"))
	libB := []byte("module b\n\nexport {\n  fn f\n}\n\n" + phase30Fn("f"))

	localSource := []byte("module app.main\n\nimport lib {\n  fn f\n}\n\nexport {\n  fn main\n}\n\n" + phase30Fn("f") + "\nfn main(value: Byte) -> Byte {\n  value\n}\n")
	twiceSource := phase30Entry("import a {\n  fn f\n}\n\nimport b {\n  fn f\n}\n")

	requireOne := func(t *testing.T, got []diagnostic.Diagnostic, code string) diagnostic.Diagnostic {
		t.Helper()
		if len(got) != 1 || got[0].Code != code {
			t.Fatalf("expected exactly %s, got %+v", code, got)
		}
		if len(got[0].Repairs) != 0 {
			t.Fatalf("%s must carry no repair, got %+v", code, got[0].Repairs)
		}
		return got[0]
	}

	t.Run("import vs local function", func(t *testing.T) {
		d := requireOne(t, Program(mergeForTest(t, localSource, map[string][]byte{"lib": libF})).Diagnostics, "module.import_collision")
		if !strings.Contains(d.Message, "module lib") || !strings.Contains(d.Message, "this module") {
			t.Fatalf("the message must name both origins, got %q", d.Message)
		}
		if len(d.Causes) != 1 || d.Causes[0].Kind != "local_declaration" || d.Causes[0].Span == nil {
			t.Fatalf("expected a cause at the local declaration, got %+v", d.Causes)
		}
	})

	t.Run("import vs import", func(t *testing.T) {
		program := mergeForTest(t, twiceSource, map[string][]byte{"a": libA, "b": libB})
		d := requireOne(t, Program(program).Diagnostics, "module.import_collision")
		if d.Primary != program.Imports[1].Items[0].Span {
			t.Fatalf("primary = %+v, want the second item %+v", d.Primary, program.Imports[1].Items[0].Span)
		}
		if len(d.Causes) != 1 || d.Causes[0].Span == nil || *d.Causes[0].Span != program.Imports[0].Items[0].Span {
			t.Fatalf("expected a cause at the first item, got %+v", d.Causes)
		}
		if !strings.Contains(d.Message, "module a") || !strings.Contains(d.Message, "module b") {
			t.Fatalf("the message must name both modules, got %q", d.Message)
		}
	})

	t.Run("control: skipping the collision check is observable", func(t *testing.T) {
		defer func() { skipImportCollisionForTest = false }()
		skipImportCollisionForTest = true
		for name, source := range map[string][]byte{"local": localSource, "import": twiceSource} {
			libs := map[string][]byte{"lib": libF, "a": libA, "b": libB}
			for _, d := range Program(mergeForTest(t, source, libs)).Diagnostics {
				if d.Code == "module.import_collision" {
					t.Fatalf("%s: the control must not refuse as import_collision, got %+v", name, d)
				}
			}
		}
	})

	t.Run("reserved builtin", func(t *testing.T) {
		for _, name := range []string{"byte_len", "byte_at", "write_text"} {
			lib := map[string][]byte{"lib": phase30Lib("  fn "+name+"\n", phase30Fn("hold"))}
			d := requireOne(t, Program(mergeForTest(t, phase30Entry("import lib {\n  fn "+name+"\n}\n"), lib)).Diagnostics, "module.import_collision")
			if !strings.Contains(d.Message, "reserved builtin") || !strings.Contains(d.Message, "module lib") {
				t.Fatalf("%s: the message must name both origins, got %q", name, d.Message)
			}
		}
	})

	t.Run("main is not importable", func(t *testing.T) {
		lib := map[string][]byte{"lib": phase30Lib("  fn main\n", phase30Fn("main"))}
		requireOne(t, Program(mergeForTest(t, phase30Entry("import lib {\n  fn main\n}\n"), lib)).Diagnostics, "module.main_not_importable")
	})

	t.Run("type items are unsupported", func(t *testing.T) {
		requireOne(t, Program(mergeForTest(t, phase30Entry("import lib {\n  type T\n}\n"), map[string][]byte{"lib": libF})).Diagnostics, "module.type_import_unsupported")
	})

	t.Run("an unloaded target fails closed", func(t *testing.T) {
		requireOne(t, Program(mergeForTest(t, phase30Entry("import gone {\n  fn f\n}\n"), nil)).Diagnostics, "module.imports_not_loaded")
	})
}

// TestPhase30UnitStampedPerFunction pins D-11: a diagnostic produced while
// checking a function of unit U carries Unit = U, and the same failure in two
// units has two IDs.
func TestPhase30UnitStampedPerFunction(t *testing.T) {
	bad := func(name string) string {
		return "fn " + name + "(value: Byte) -> Byte {\n  let result = missing(value)\n  result\n}\n"
	}
	libSource := []byte("module lib\n\nexport {\n  fn shown\n}\n\n" + bad("shown"))
	entry := []byte("module app.main\n\nimport lib {\n  fn shown\n}\n\nexport {\n  fn main\n}\n\n" + bad("main"))
	program := mergeForTest(t, entry, map[string][]byte{"lib": libSource})

	result := Program(program)
	var libDiagnostic, entryDiagnostic *diagnostic.Diagnostic
	for index := range result.Diagnostics {
		switch result.Diagnostics[index].Unit {
		case "lib":
			libDiagnostic = &result.Diagnostics[index]
		case "":
			entryDiagnostic = &result.Diagnostics[index]
		}
	}
	if libDiagnostic == nil || entryDiagnostic == nil {
		t.Fatalf("expected one diagnostic per unit, got %+v", result.Diagnostics)
	}
	if libDiagnostic.Code != entryDiagnostic.Code {
		t.Fatalf("the same failure must have the same code, got %s and %s", libDiagnostic.Code, entryDiagnostic.Code)
	}
	if libDiagnostic.ID == entryDiagnostic.ID {
		t.Fatalf("the same check failure in two units must have two IDs, both %s", libDiagnostic.ID)
	}

	t.Run("entry diagnostics keep the pre-change ID", func(t *testing.T) {
		recomputed := diagnostic.Error(entryDiagnostic.Code, entryDiagnostic.Primary, entryDiagnostic.Message, entryDiagnostic.Causes...)
		if entryDiagnostic.ID != recomputed.ID {
			t.Fatalf("the entry ID must be the unit-free identity %s, got %s", recomputed.ID, entryDiagnostic.ID)
		}
	})

	t.Run("control: skipping the stamp leaves the unit empty", func(t *testing.T) {
		restore := SetSkipUnitStampingForTest(true)
		defer restore()
		for _, d := range Program(program).Diagnostics {
			if d.Unit != "" {
				t.Fatalf("the control must leave Unit empty, got %+v", d)
			}
		}
	})

	t.Run("a call-graph cycle in a library is attributed to the library", func(t *testing.T) {
		cyclic := []byte("module lib\n\nexport {\n  fn shown\n}\n\nfn shown(value: Byte) -> Byte {\n  let result = other(value)\n  result\n}\n\nfn other(value: Byte) -> Byte {\n  let result = shown(value)\n  result\n}\n")
		out := Program(mergeForTest(t, phase30Entry("import lib {\n  fn shown\n}\n"), map[string][]byte{"lib": cyclic}))
		if len(out.Diagnostics) != 1 || out.Diagnostics[0].Code != core.CallGraphCycle || out.Diagnostics[0].Unit != "lib" {
			t.Fatalf("expected a lib-attributed cycle, got %+v", out.Diagnostics)
		}
	})
}

// TestPhase30ImportItemOrderIrrelevant pins the ordering edge: permuting the
// items inside an import block and the order of import declarations changes
// neither the core JSON nor the diagnostics.
func TestPhase30ImportItemOrderIrrelevant(t *testing.T) {
	libA := []byte("module a\n\nexport {\n  fn fa\n  fn fb\n}\n\n" + phase30Fn("fa") + "\n" + phase30Fn("fb"))
	libB := []byte("module b\n\nexport {\n  fn gb\n}\n\n" + phase30Fn("gb"))
	main := "\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let one = fa(value)\n  let two = fb(one)\n  let three = gb(two)\n  three\n}\n"
	variants := []string{
		"module app.main\n\nimport a {\n  fn fa\n  fn fb\n}\n\nimport b {\n  fn gb\n}\n" + main,
		"module app.main\n\nimport b {\n  fn gb\n}\n\nimport a {\n  fn fb\n  fn fa\n}\n" + main,
	}
	var encoded []string
	for _, variant := range variants {
		result := Program(mergeForTest(t, []byte(variant), map[string][]byte{"a": libA, "b": libB}))
		if len(result.Diagnostics) != 0 {
			t.Fatalf("expected a clean check, got %+v", result.Diagnostics)
		}
		encoded = append(encoded, phase30JSON(t, result.Program)+phase30JSON(t, result.Diagnostics))
	}
	if encoded[0] != encoded[1] {
		t.Fatalf("permuting import order changed the output:\n%s\n%s", encoded[0], encoded[1])
	}
	want := []core.ModuleImport{{Module: "a", Name: "fa"}, {Module: "a", Name: "fb"}, {Module: "b", Name: "gb"}}
	result := Program(mergeForTest(t, []byte(variants[1]), map[string][]byte{"a": libA, "b": libB}))
	if phase30JSON(t, result.Program.Imports) != phase30JSON(t, want) {
		t.Fatalf("Program.Imports must be sorted, got %s", phase30JSON(t, result.Program.Imports))
	}
}

// phase30LocalFileByteEntry is the Phase 23 local-file-byte entry program
// (a corpus program whose `main` checkLocalFileByteEntry admits) with one
// import added, so the entry has a real unit scope beside its foreign table.
func phase30LocalFileByteEntry(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../examples/phase23/file_byte.schway")
	if err != nil {
		t.Fatal(err)
	}
	const header = "module phase23.file_byte\n"
	if !strings.HasPrefix(string(source), header) {
		t.Fatalf("unexpected corpus header %q", source[:40])
	}
	return []byte(header + "\nimport lib {\n  fn twice\n}\n" + strings.TrimPrefix(string(source), header))
}

// TestPhase30LibraryPrivateMainNotEntryShaped pins D-29 / D-20: a library
// module may define a private `main`; it is an ordinary non-entry function.
// Entry-unit diagnostics and the entry result are identical with and without
// it, it is never an entry root, and no refusal code exists for defining it.
func TestPhase30LibraryPrivateMainNotEntryShaped(t *testing.T) {
	entry := phase30LocalFileByteEntry(t)
	entryID := "s1:phase23.file_byte:fn:main"
	libBase := "module lib\n\nexport {\n  fn twice\n}\n\n" + phase30Fn("twice")
	pathMain := "\nfn main(path: PathToken) -> U64 {\n  let owner = try schway_file_byte_acquire(path)\n  let value = try schway_file_byte_use(owner)\n  value\n}\n"
	ordinaryMain := "\nfn main(value: Byte) -> Byte {\n  value\n}\n"

	entryUnit := func(diagnostics []diagnostic.Diagnostic) string {
		kept := []diagnostic.Diagnostic{}
		for _, d := range diagnostics {
			if d.Unit == "" {
				kept = append(kept, d)
			}
		}
		return phase30JSON(t, kept)
	}
	check := func(libSource string) Result {
		return Program(mergeForTest(t, entry, map[string][]byte{"lib": []byte(libSource)}))
	}
	entryMainJSON := func(result Result) string {
		return phase30JSON(t, phase30FunctionByID(t, result.Program, entryID))
	}

	without := check(libBase)
	if len(without.Diagnostics) != 0 {
		t.Fatalf("the baseline must check clean, got %+v", without.Diagnostics)
	}
	for _, variant := range []struct{ name, source string }{
		{"ordinary private main", libBase + ordinaryMain},
		{"path token main in the acquire/use shape", libBase + pathMain},
	} {
		t.Run(variant.name, func(t *testing.T) {
			with := check(variant.source)
			if entryUnit(with.Diagnostics) != entryUnit(without.Diagnostics) {
				t.Fatalf("entry-unit diagnostics changed:\n%s\n%s", entryUnit(with.Diagnostics), entryUnit(without.Diagnostics))
			}
			if entryMainJSON(with) != entryMainJSON(without) {
				t.Fatalf("the entry main's core function changed:\n%s\n%s", entryMainJSON(with), entryMainJSON(without))
			}
			// The only function treated as the entry `main` is the entry's.
			for _, function := range with.Program.Functions {
				if function.Name == "main" && function.ID != entryID && function.Linear != nil {
					for _, operation := range function.Linear.Operations {
						if operation.Kind == core.OpForeignCall {
							t.Fatalf("%s was lowered as an entry-shaped foreign main", function.ID)
						}
					}
				}
			}
			// D-29: there is no refusal code for defining a library main.
			for _, d := range with.Diagnostics {
				if strings.HasPrefix(d.Code, "module.") {
					t.Fatalf("defining a library main must not produce a module.* refusal, got %+v", d)
				}
			}
		})
	}

	t.Run("an ordinary private main checks clean with its own ID", func(t *testing.T) {
		with := check(libBase + ordinaryMain)
		if len(with.Diagnostics) != 0 {
			t.Fatalf("expected a clean check, got %+v", with.Diagnostics)
		}
		libMain := phase30FunctionByID(t, with.Program, "s1:lib:fn:main")
		if libMain.Module != "lib" || libMain.Exported {
			t.Fatalf("lib main must be a private lib function, got %+v", libMain)
		}
	})

	t.Run("control: program-wide recognizer scope makes lib main entry-shaped", func(t *testing.T) {
		defer func() { programWideRecognizerScopeForTest = false }()
		defaultResult := check(libBase + pathMain)
		entryShaped := func(result Result) []string {
			ids := []string{}
			for _, function := range result.Program.Functions {
				if function.Name != "main" || function.Linear == nil {
					continue
				}
				for _, operation := range function.Linear.Operations {
					if operation.Kind == core.OpForeignCall {
						ids = append(ids, function.ID)
						break
					}
				}
			}
			sort.Strings(ids)
			return ids
		}
		if got := entryShaped(defaultResult); len(got) != 1 || got[0] != entryID {
			t.Fatalf("by default exactly the entry main is entry-shaped, got %v", got)
		}
		programWideRecognizerScopeForTest = true
		wide := check(libBase + pathMain)
		got := entryShaped(wide)
		if len(got) != 2 {
			t.Fatalf("the control must let lib main be recognized as an entry shape, got %v", got)
		}
		// The control moves the entry RESULT (the set of entry-shaped core
		// functions grows from the entry main alone to include lib main) and
		// the lib-unit diagnostics (lib main is admitted instead of refused as
		// a non-entry PathToken function). It does not move the entry-unit
		// diagnostics, which are empty in both: lib main never reaches the
		// entry's own checks, so the entry-result difference is the observable.
		libRefusals := func(result Result) int {
			count := 0
			for _, d := range result.Diagnostics {
				if d.Unit == "lib" {
					count++
				}
			}
			return count
		}
		if libRefusals(defaultResult) == 0 || libRefusals(wide) != 0 {
			t.Fatalf("the control must change how lib main is treated: default lib diagnostics %d, program-wide %d", libRefusals(defaultResult), libRefusals(wide))
		}
	})
}

// TestPhase30ExportedButUndeclaredImportRefused pins WR-03: importing a name the
// target exports but never declares is refused at the import item, and no
// ModuleImport fact is produced. The armed control admits it and shows the
// dangling fact.
func TestPhase30ExportedButUndeclaredImportRefused(t *testing.T) {
	libs := map[string][]byte{"lib": phase30Lib("  fn ghost\n", phase30Fn("other"))}
	entry := phase30Entry("import lib {\n  fn ghost\n}\n")
	got := Program(mergeForTest(t, entry, libs))
	if len(got.Diagnostics) != 1 || got.Diagnostics[0].Code != "module.not_exported" {
		t.Fatalf("expected one module.not_exported for the undeclared export, got %+v", got.Diagnostics)
	}
	d := got.Diagnostics[0]
	if d.Unit != "" || len(d.Causes) != 1 || d.Causes[0].Kind != "export_list" || d.Causes[0].Unit != "lib" {
		t.Fatalf("diagnostic must belong to the importer with an export_list cause in lib: %+v", d)
	}
	if len(got.Program.Imports) != 0 {
		t.Fatalf("a refused import must leave no ModuleImport fact, got %+v", got.Program.Imports)
	}

	t.Run("control: skipping the declaration check leaves a dangling fact", func(t *testing.T) {
		skipDeclaredImportCheckForTest = true
		defer func() { skipDeclaredImportCheckForTest = false }()
		admitted := Program(mergeForTest(t, entry, libs))
		for _, d := range admitted.Diagnostics {
			if d.Code == "module.not_exported" {
				t.Fatalf("the control must not refuse the import: %+v", admitted.Diagnostics)
			}
		}
	})
}

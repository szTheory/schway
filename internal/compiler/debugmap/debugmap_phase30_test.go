package debugmap

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const (
	phase30DebugLib = "module lib\n\nexport {\n  fn helper\n  fn other\n}\n\nfn helper(value: Byte) -> Byte {\n  value\n}\n\nfn other(value: Byte) -> Byte {\n  let result = helper(value)\n  result\n}\n"
	// The entry's own `helper` has a different body (two bindings) from lib's.
	phase30DebugEntry = "module app.main\n\nimport lib {\n  fn other\n}\n\nexport {\n  fn main\n}\n\nfn helper(value: Byte) -> Byte {\n  let first = value\n  let second = first\n  second\n}\n\nfn main(value: Byte) -> Byte {\n  let one = helper(value)\n  let two = other(one)\n  two\n}\n"
)

// phase30MergedProgram assembles the merged ast by hand in the shape plan
// 30-04's modload produces (functions stamped with their module, entry first)
// and runs check on it.
func phase30MergedProgram(t *testing.T, entry string, libs map[string]string) (ast.Program, core.Program) {
	t.Helper()
	parse := func(source string) ast.Program {
		parsed := syntax.Parse([]byte(source))
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
	checked := check.Program(merged)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("merged fixture failed check: %+v", checked.Diagnostics)
	}
	return merged, checked.Program
}

func phase30Build(t *testing.T, astProgram ast.Program, coreProgram core.Program) Map {
	t.Helper()
	built, _, err := Build(context.Background(), astProgram, coreProgram)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return built
}

// entriesFor returns the entries whose function lies in the named module's
// function with the given name.
func phase30Entries(built Map, coreProgram core.Program, module, name string) []Entry {
	var functionID string
	for _, function := range coreProgram.Functions {
		if function.Module == module && function.Name == name {
			functionID = function.ID
		}
	}
	var out []Entry
	for _, entry := range built.Entries {
		if functionID != "" && strings.HasPrefix(entry.CoreID, functionID) && entry.CoreID == functionID {
			out = append(out, entry)
		}
	}
	return out
}

// TestPhase30DebugMapUnit proves D-11: the debug map joins ast and core
// functions by (module, name), so two same-named functions of different
// modules never cross-join, and every entry carries its module as Unit. The
// control engages bareNameJoinForTest and observes the wrong join.
func TestPhase30DebugMapUnit(t *testing.T) {
	astProgram, coreProgram := phase30MergedProgram(t, phase30DebugEntry, map[string]string{"lib": phase30DebugLib})
	built := phase30Build(t, astProgram, coreProgram)

	libHelper := phase30Entries(built, coreProgram, "lib", "helper")
	entryHelper := phase30Entries(built, coreProgram, "", "helper")
	// lib.helper: one return (no bindings); app helper: two bindings + return.
	if len(libHelper) != 1 || len(entryHelper) != 3 {
		t.Fatalf("want 1 lib helper entry and 3 entry helper entries, got %d and %d", len(libHelper), len(entryHelper))
	}
	for _, entry := range libHelper {
		if entry.Unit != "lib" || !strings.Contains(entry.OperationID, "lib") {
			t.Fatalf("lib helper entry must carry Unit lib and an s1:lib: operation ID, got %+v", entry)
		}
	}
	for _, entry := range entryHelper {
		if entry.Unit != "" || strings.Contains(entry.OperationID, ":lib:") {
			t.Fatalf("entry helper must carry the empty Unit and an entry-module operation ID, got %+v", entry)
		}
	}
	for _, entry := range built.Entries {
		if strings.Contains(entry.CoreID, ":lib:") != (entry.Unit == "lib") {
			t.Fatalf("Unit disagrees with the function's module: %+v", entry)
		}
	}

	t.Run("control: bare-name join cross-joins same-named functions", func(t *testing.T) {
		bareNameJoinForTest = true
		defer func() { bareNameJoinForTest = false }()
		wrong := phase30Build(t, astProgram, coreProgram)
		libSpan := func(m Map) any { return phase30Entries(m, coreProgram, "lib", "helper")[0].SourceSpan }
		if libSpan(wrong) == libSpan(built) {
			t.Fatal("control: the bare-name join must give lib helper the entry helper's AST body, moving its source span")
		}
	})

	t.Run("a no-import program is unaffected by the module-keyed join", func(t *testing.T) {
		astSingle, coreSingle := phase30MergedProgram(t, phase30SingleSource, nil)
		keyed := phase30Build(t, astSingle, coreSingle)
		bareNameJoinForTest = true
		defer func() { bareNameJoinForTest = false }()
		bare := phase30Build(t, astSingle, coreSingle)
		a, _ := json.Marshal(keyed)
		b, _ := json.Marshal(bare)
		if string(a) != string(b) || len(keyed.Entries) == 0 {
			t.Fatalf("no-import debug map bytes must be identical: %s vs %s", a, b)
		}
		for _, entry := range keyed.Entries {
			if entry.Unit != "" {
				t.Fatalf("a no-import entry must have no unit, got %+v", entry)
			}
		}
	})
}

const phase30SingleSource = "module app.single\n\nexport {\n  fn main\n}\n\nfn main(value: Byte) -> Byte {\n  let copy = value\n  copy\n}\n"

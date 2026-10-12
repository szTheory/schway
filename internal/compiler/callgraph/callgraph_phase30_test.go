package callgraph

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

func phase30Function(id, name, module string, callees ...string) core.Function {
	function := core.Function{
		ID: id, Name: name, Module: module, EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter: core.Parameter{ID: id + ":place:0", Name: "value", Type: "Byte"},
		Linear:    &core.LinearBody{ID: id + ":linear"},
	}
	for index, callee := range callees {
		function.Linear.Operations = append(function.Linear.Operations, core.LinearOperation{
			ID: id + ":op:" + string(rune('0'+index)), Kind: core.OpCall, CalleeID: callee,
		})
	}
	return function
}

// TestPhase30EntryIgnoresLibraryMain proves EntryFunction considers only
// entry-module (Module "") functions, so an imported library's `main` or an
// unimported library export never makes the entry ambiguous (T-30-26). The
// controls engage entryModuleFilterOffForTest and observe the refusal the
// filter prevents.
func TestPhase30EntryIgnoresLibraryMain(t *testing.T) {
	t.Run("library main does not compete with the entry main", func(t *testing.T) {
		// The library main precedes the entry main in program order.
		program := core.Program{Functions: []core.Function{
			phase30Function("s1:lib:fn:main", "main", "lib"),
			phase30Function("s1:app:fn:main", "main", "", "s1:lib:fn:helper"),
			phase30Function("s1:lib:fn:helper", "helper", "lib"),
		}}
		entry, err := EntryFunction(program)
		if err != nil || entry.ID != "s1:app:fn:main" {
			t.Fatalf("want the entry-module main, got %q err=%v", entry.ID, err)
		}

		entryModuleFilterOffForTest = true
		defer func() { entryModuleFilterOffForTest = false }()
		_, err = EntryFunction(program)
		ambiguous, ok := EntryAmbiguousError(err)
		if !ok || !ambiguous.duplicateMain {
			t.Fatalf("control: with the filter off a library main must yield the duplicate-main refusal, got %v", err)
		}
	})

	t.Run("an unimported library export is not a root", func(t *testing.T) {
		// No `main` anywhere: the roots decide. `extra` is a library export
		// with in-degree zero; without the filter it ties with the entry.
		program := core.Program{Functions: []core.Function{
			phase30Function("s1:app:fn:start", "start", "", "s1:app:fn:step"),
			phase30Function("s1:app:fn:step", "step", ""),
			phase30Function("s1:lib:fn:extra", "extra", "lib"),
		}}
		entry, err := EntryFunction(program)
		if err != nil || entry.ID != "s1:app:fn:start" {
			t.Fatalf("want entry start, got %q err=%v", entry.ID, err)
		}

		// Equal closure sizes make the unfiltered tie ambiguous.
		tied := core.Program{Functions: []core.Function{
			phase30Function("s1:app:fn:start", "start", ""),
			phase30Function("s1:lib:fn:extra", "extra", "lib"),
		}}
		if entry, err := EntryFunction(tied); err != nil || entry.ID != "s1:app:fn:start" {
			t.Fatalf("want entry start next to an unimported export, got %q err=%v", entry.ID, err)
		}
		entryModuleFilterOffForTest = true
		defer func() { entryModuleFilterOffForTest = false }()
		if _, err := EntryFunction(tied); err == nil {
			t.Fatal("control: with the filter off the unimported export must tie with the entry")
		}
	})

	t.Run("a no-import program keeps its entry choice", func(t *testing.T) {
		program := core.Program{Functions: []core.Function{
			phase30Function("fn:a", "a", "", "fn:b"),
			phase30Function("fn:b", "b", ""),
		}}
		entry, err := EntryFunction(program)
		if err != nil || entry.ID != "fn:a" {
			t.Fatalf("want fn:a, got %q err=%v", entry.ID, err)
		}
	})
}

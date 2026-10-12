package interp

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
)

// TestPhase30FindFunctionPrefersEntryModule proves a name resolves to the
// entry-module function first, so a same-named library function never shadows
// the entry (T-30-26). The control engages entryModulePreferenceOffForTest and
// observes the earlier library main win.
func TestPhase30FindFunctionPrefersEntryModule(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		{ID: "s1:lib:fn:main", Name: "main", Module: "lib"},
		{ID: "s1:lib:fn:f", Name: "f", Module: "lib"},
		{ID: "s1:app:fn:main", Name: "main"},
	}}

	if got, ok := findFunction(program, "main"); !ok || got.ID != "s1:app:fn:main" {
		t.Fatalf("want the entry-module main, got %q ok=%v", got.ID, ok)
	}
	if got, ok := findFunction(program, "f"); !ok || got.ID != "s1:lib:fn:f" {
		t.Fatalf("a library-only name must still resolve, got %q ok=%v", got.ID, ok)
	}
	if _, ok := findFunction(program, "missing"); ok {
		t.Fatal("an unknown name must not resolve")
	}

	entryModulePreferenceOffForTest = true
	defer func() { entryModulePreferenceOffForTest = false }()
	if got, ok := findFunction(program, "main"); !ok || got.ID != "s1:lib:fn:main" {
		t.Fatalf("control: with the preference off the first match in program order wins, got %q ok=%v", got.ID, ok)
	}
}

package callgraph_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestEntryFunctionResolvesUniqueRoot proves EntryFunction resolves the
// single in-degree-zero function on both a real multi-function tracer
// fixture and a real, deeper acyclic diamond corpus, with no error.
func TestEntryFunctionResolvesUniqueRoot(t *testing.T) {
	basic := checkedPhase11Fixture(t, "multi_function_entry_basic.lang")
	entry, err := callgraph.EntryFunction(basic)
	if err != nil {
		t.Fatalf("multi_function_entry_basic.lang: unexpected error: %v", err)
	}
	if entry.Name != "main" {
		t.Fatalf("multi_function_entry_basic.lang: want entry %q, got %q", "main", entry.Name)
	}

	diamond := checkedPhase07Fixture(t, "deep_diamond_acyclic.lang")
	diamondEntry, err := callgraph.EntryFunction(diamond)
	if err != nil {
		t.Fatalf("deep_diamond_acyclic.lang: unexpected error: %v", err)
	}
	if diamondEntry.Name != "top4" {
		t.Fatalf("deep_diamond_acyclic.lang: want entry %q, got %q", "top4", diamondEntry.Name)
	}
}

// TestEntryFunctionRefusesManyRoots proves the many-roots refusal is named,
// byte-stable, and independent of declaration order: three functions that
// call nothing and are called by nothing are three genuinely tied
// in-degree-zero candidates (an all-empty-closure tie, per
// EntryFunction's own documented tie-break), refused rather than guessed.
// The SAME program built with its three functions declared in two
// different orders must produce byte-identical error text -- the
// determinism claim; comparing one message to itself would prove nothing.
func TestEntryFunctionRefusesManyRoots(t *testing.T) {
	forward := core.Program{Functions: []core.Function{
		leafFunction("fn:a"), leafFunction("fn:b"), leafFunction("fn:c"),
	}}
	reversed := core.Program{Functions: []core.Function{
		leafFunction("fn:c"), leafFunction("fn:b"), leafFunction("fn:a"),
	}}

	_, forwardErr := callgraph.EntryFunction(forward)
	_, reversedErr := callgraph.EntryFunction(reversed)
	if forwardErr == nil || reversedErr == nil {
		t.Fatalf("expected both declaration orders to refuse, got forward=%v reversed=%v", forwardErr, reversedErr)
	}
	forwardAmbiguous, ok := callgraph.EntryAmbiguousError(forwardErr)
	if !ok {
		t.Fatalf("forward: expected an entry-ambiguity refusal, got: %v", forwardErr)
	}
	reversedAmbiguous, ok := callgraph.EntryAmbiguousError(reversedErr)
	if !ok {
		t.Fatalf("reversed: expected an entry-ambiguity refusal, got: %v", reversedErr)
	}
	if forwardAmbiguous.Code() != core.EntryAmbiguous || reversedAmbiguous.Code() != core.EntryAmbiguous {
		t.Fatalf("expected code %q on both, got forward=%q reversed=%q", core.EntryAmbiguous, forwardAmbiguous.Code(), reversedAmbiguous.Code())
	}
	if forwardErr.Error() != reversedErr.Error() {
		t.Fatalf("declaration order moved the refusal's own error text:\nforward:  %s\nreversed: %s", forwardErr.Error(), reversedErr.Error())
	}
	roots := forwardAmbiguous.Roots()
	if len(roots) != 3 {
		t.Fatalf("expected three candidate roots, got %v", roots)
	}
	if !sort.StringsAreSorted(roots) {
		t.Fatalf("expected candidate roots in sorted order, got %v", roots)
	}
}

// TestEntryFunctionRefusesZeroRoots constructs a core.Program whose
// functions form a closed call structure with no in-degree-zero node at
// all (a two-function mutual cycle, each called exactly once by the
// other) and asserts the same named refusal -- no panic, no
// index-out-of-range.
func TestEntryFunctionRefusesZeroRoots(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		syntheticFunction("fn:a", "fn:b", "fn:a:op:0"),
		syntheticFunction("fn:b", "fn:a", "fn:b:op:0"),
	}}
	_, err := callgraph.EntryFunction(program)
	if err == nil {
		t.Fatal("expected a refusal, got nil")
	}
	ambiguous, ok := callgraph.EntryAmbiguousError(err)
	if !ok {
		t.Fatalf("expected an entry-ambiguity refusal, got: %v", err)
	}
	if ambiguous.Code() != core.EntryAmbiguous {
		t.Fatalf("expected code %q, got %q", core.EntryAmbiguous, ambiguous.Code())
	}
	if len(ambiguous.Roots()) != 0 {
		t.Fatalf("expected zero candidate roots, got %v", ambiguous.Roots())
	}
}

// TestEntryFunctionRefusesEmptyProgram proves a core.Program with zero
// functions returns the named refusal, not a zero-value core.Function
// silently treated as success.
func TestEntryFunctionRefusesEmptyProgram(t *testing.T) {
	function, err := callgraph.EntryFunction(core.Program{})
	if err == nil {
		t.Fatalf("expected a refusal, got function=%+v", function)
	}
	ambiguous, ok := callgraph.EntryAmbiguousError(err)
	if !ok {
		t.Fatalf("expected an entry-ambiguity refusal, got: %v", err)
	}
	if ambiguous.Code() != core.EntryAmbiguous {
		t.Fatalf("expected code %q, got %q", core.EntryAmbiguous, ambiguous.Code())
	}
}

// TestEntryFunctionEmitsUnreachableFunctions is the adopted contract for a
// declared-but-uncalled, non-entry function (D-11-05, this phase's PLAN.md
// must_haves): multi_function_unreachable.lang declares `main` (calls
// `callee`), `callee`, and `orphan` (declared, never called, never
// exported). `orphan`'s own in-degree is zero too, exactly like main's --
// EntryFunction's contract, stated on its own doc comment, is that main's
// non-empty reachable closure ({callee}) strictly exceeds orphan's empty
// one, so main resolves as the unique entry: the unreachable function does
// NOT change entry resolution or (see cgen_program_test.go's own
// end-to-end fixtures) the execution document.
func TestEntryFunctionEmitsUnreachableFunctions(t *testing.T) {
	program := checkedPhase11Fixture(t, "multi_function_unreachable.lang")
	entry, err := callgraph.EntryFunction(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Name != "main" {
		t.Fatalf("want entry %q, got %q", "main", entry.Name)
	}
	found := false
	for _, function := range program.Functions {
		if function.Name == "orphan" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected the unreachable function \"orphan\" to still be present in the checked core.Program")
	}
}

// TestEntryFunctionAgreesWithSingleExport (D-11-06) iterates every real
// .lang fixture under testdata/phase1 through testdata/phase07 whose
// single declared export names a function, and asserts
// EntryFunction(program).Name equals that export's own name -- the
// independent re-derivation that lets this phase avoid adding an
// Exported/EntryFunctionID field to core.Program.
func TestEntryFunctionAgreesWithSingleExport(t *testing.T) {
	directories := []string{"phase1", "phase2", "phase3", "phase4", "phase5", "phase6", "phase07"}
	checkedCount := 0
	for _, directory := range directories {
		dir := testsupport.ProjectPath("testdata", directory)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", directory, entry.Name(), err)
			}
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				continue // malformed-on-purpose negative-control fixture
			}
			if len(parsed.Program.Exports) != 1 || parsed.Program.Exports[0].Kind != "fn" {
				continue // a data-type export, or not exactly one export
			}
			result := check.Program(parsed.Program)
			if len(result.Diagnostics) != 0 {
				continue // refused-on-purpose negative-control fixture
			}
			entryFunction, err := callgraph.EntryFunction(result.Program)
			if err != nil {
				t.Fatalf("%s/%s: EntryFunction refused an admitted program: %v", directory, entry.Name(), err)
			}
			wantName := parsed.Program.Exports[0].Name
			if entryFunction.Name != wantName {
				t.Fatalf("%s/%s: export names %q, EntryFunction resolved %q", directory, entry.Name(), wantName, entryFunction.Name)
			}
			checkedCount++
		}
	}
	if checkedCount == 0 {
		t.Fatal("expected at least one real fixture to be exercised")
	}
	t.Logf("EntryFunction agreed with ast.Export on %d real fixtures", checkedCount)
}

// TestNoCoreProgramEntryField performs a source scan of
// internal/compiler/core/core.go's own core.Function struct declaration
// (never a reflection check on a constructed value) and fails if it
// declares an "Exported" or "EntryFunctionID" field -- D-11-06's guard
// against a future phase silently reintroducing a schema field this phase
// deliberately avoided.
func TestNoCoreProgramEntryField(t *testing.T) {
	path := testsupport.ProjectPath("internal", "compiler", "core", "core.go")
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	ast.Inspect(parsed, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != "Function" {
			return true
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}
		found = true
		for _, field := range structType.Fields.List {
			for _, name := range field.Names {
				if name.Name == "Exported" || name.Name == "EntryFunctionID" {
					t.Fatalf("core.Function declares a %q field -- D-11-06 requires this stay a call-graph-derived fact, never a schema field", name.Name)
				}
			}
		}
		return true
	})
	if !found {
		t.Fatal("did not find core.Function's own struct declaration to scan")
	}
}

// TestSessionRunSitesDoNotIndexFunctionsZero (the adopted assumption-delta
// invariant, PLAN.md's <assumption_delta_decision>) parses
// internal/compiler/session/session.go and asserts that the bodies of
// RunInterpreter, RunNative, and interpreterInputs contain no
// "Functions[0]" index expression.
func TestSessionRunSitesDoNotIndexFunctionsZero(t *testing.T) {
	path := testsupport.ProjectPath("internal", "compiler", "session", "session.go")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, source, 0)
	if err != nil {
		t.Fatal(err)
	}
	targets := map[string]bool{"RunInterpreter": true, "RunNative": true, "interpreterInputs": true}
	checked := 0
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || !targets[function.Name.Name] || function.Body == nil {
			continue
		}
		checked++
		start := fileSet.Position(function.Body.Pos()).Offset
		end := fileSet.Position(function.Body.End()).Offset
		body := string(source[start:end])
		if strings.Contains(body, "Functions[0]") {
			t.Fatalf("%s's own body indexes Functions[0]:\n%s", function.Name.Name, body)
		}
	}
	if checked != len(targets) {
		t.Fatalf("expected to find all %d target functions, found %d", len(targets), checked)
	}
}

// checkedPhase07Fixture and checkedPhase11Fixture check a real fixture
// through this project's normal admission path and fail the test if it is
// unexpectedly rejected.

func checkedPhase11Fixture(t *testing.T, name string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase11", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed to parse: %+v", name, parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture %q unexpectedly rejected: %+v", name, result.Diagnostics)
	}
	return result.Program
}

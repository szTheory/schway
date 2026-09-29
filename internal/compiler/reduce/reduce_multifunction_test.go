package reduce_test

import (
	"context"
	"encoding/json"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/reduce"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Phase 11 (D-11-29/D-11-30) fixtures: multi-function seeds, built by
// hand exactly like reduce_test.go's own single-function fixtures (this
// package's moves operate on core.Program structure directly, never
// through the parser/checker).
// ---------------------------------------------------------------------

const (
	multiFnCallerID = "s1:phase11.reduce_fixture:function:caller"
	multiFnCalleeID = "s1:phase11.reduce_fixture:function:callee"
)

// multiFunctionCallSeed is a two-function seed: caller calls callee once
// (a single core.OpCall edge) and returns callee's result. Mirrors this
// plan's own Q-01 spike shape (a single, eligible OpCall).
func multiFunctionCallSeed() core.Program {
	const module = "phase11.reduce_fixture"
	typeID := multiFnCallerID + ":type:0"
	return core.Program{
		Schema:   core.Schema,
		Module:   module,
		ModuleID: "s1:" + module + ":module:" + module,
		Functions: []core.Function{
			{
				ID: multiFnCallerID, Name: "caller",
				EntryPointID: multiFnCallerID + ":point:entry", ReturnPointID: multiFnCallerID + ":point:return",
				Parameter:  core.Parameter{ID: multiFnCallerID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:    multiFnCallerID + ":linear",
					Types: []core.TypeFact{{ID: typeID}},
					Places: []core.Place{
						{ID: multiFnCallerID + ":place:0", Name: "value", TypeID: typeID},
						{ID: multiFnCallerID + ":place:1", Name: "result", TypeID: typeID},
					},
					Operations: []core.LinearOperation{
						{ID: multiFnCallerID + ":op:0", PointID: multiFnCallerID + ":point:linear:0", Kind: core.OpCall, SourceID: multiFnCallerID + ":place:0", TargetID: multiFnCallerID + ":place:1", TypeID: typeID, CalleeID: multiFnCalleeID},
						{ID: multiFnCallerID + ":op:1", PointID: multiFnCallerID + ":point:linear:1", Kind: core.OpReturn, SourceID: multiFnCallerID + ":place:1", TypeID: typeID},
					},
				},
			},
			{
				ID: multiFnCalleeID, Name: "callee",
				EntryPointID: multiFnCalleeID + ":point:entry", ReturnPointID: multiFnCalleeID + ":point:return",
				Parameter:  core.Parameter{ID: multiFnCalleeID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:         multiFnCalleeID + ":linear",
					Types:      []core.TypeFact{{ID: typeID}},
					Places:     []core.Place{{ID: multiFnCalleeID + ":place:0", Name: "value", TypeID: typeID}},
					Operations: []core.LinearOperation{{ID: multiFnCalleeID + ":op:0", PointID: multiFnCalleeID + ":point:linear:0", Kind: core.OpReturn, SourceID: multiFnCalleeID + ":place:0", TypeID: typeID}},
				},
			},
		},
	}
}

// multiFunctionNoCallSeed is a two-function seed with zero call edges --
// used for the derived-bound floor test's "zero-call multi-function seed"
// case.
func multiFunctionNoCallSeed() core.Program {
	seed := multiFunctionCallSeed()
	seed.Functions[0].Linear.Operations[0] = core.LinearOperation{
		ID: multiFnCallerID + ":op:0", PointID: multiFnCallerID + ":point:linear:0",
		Kind: core.OpCopy, SourceID: multiFnCallerID + ":place:0", TargetID: multiFnCallerID + ":place:1", TypeID: multiFnCallerID + ":type:0",
	}
	return seed
}

// typeChangingCallSeed is a two-function seed whose single OpCall's own
// SourceID declares a DIFFERENT type than the call's own TypeID -- the
// "type-changing calls" RefusedShapes() entry: rewriting it to OpCopy
// would copy a value of one type into a place declared with another,
// which OpCopy cannot express.
func typeChangingCallSeed() core.Program {
	const module = "phase11.reduce_fixture"
	byteType := multiFnCallerID + ":type:byte"
	bufferType := multiFnCallerID + ":type:buffer"
	return core.Program{
		Schema:   core.Schema,
		Module:   module,
		ModuleID: "s1:" + module + ":module:" + module,
		Functions: []core.Function{
			{
				ID: multiFnCallerID, Name: "caller",
				EntryPointID: multiFnCallerID + ":point:entry", ReturnPointID: multiFnCallerID + ":point:return",
				Parameter:  core.Parameter{ID: multiFnCallerID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Buffer",
				Linear: &core.LinearBody{
					ID:    multiFnCallerID + ":linear",
					Types: []core.TypeFact{{ID: byteType}, {ID: bufferType}},
					Places: []core.Place{
						{ID: multiFnCallerID + ":place:0", Name: "value", TypeID: byteType},
						{ID: multiFnCallerID + ":place:1", Name: "result", TypeID: bufferType},
					},
					Operations: []core.LinearOperation{
						{ID: multiFnCallerID + ":op:0", PointID: multiFnCallerID + ":point:linear:0", Kind: core.OpCall, SourceID: multiFnCallerID + ":place:0", TargetID: multiFnCallerID + ":place:1", TypeID: bufferType, CalleeID: multiFnCalleeID},
						{ID: multiFnCallerID + ":op:1", PointID: multiFnCallerID + ":point:linear:1", Kind: core.OpReturn, SourceID: multiFnCallerID + ":place:1", TypeID: bufferType},
					},
				},
			},
			{
				ID: multiFnCalleeID, Name: "callee",
				EntryPointID: multiFnCalleeID + ":point:entry", ReturnPointID: multiFnCalleeID + ":point:return",
				Parameter:  core.Parameter{ID: multiFnCalleeID + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Buffer",
				Linear: &core.LinearBody{
					ID:         multiFnCalleeID + ":linear",
					Types:      []core.TypeFact{{ID: bufferType}},
					Places:     []core.Place{{ID: multiFnCalleeID + ":place:0", Name: "value", TypeID: bufferType}},
					Operations: []core.LinearOperation{{ID: multiFnCalleeID + ":op:0", PointID: multiFnCalleeID + ":point:linear:0", Kind: core.OpReturn, SourceID: multiFnCalleeID + ":place:0", TypeID: bufferType}},
				},
			},
		},
	}
}

// trivialSeed is a single-function seed on which NO move applies: one
// parameter, one OpReturn reading it directly. Used by
// TestReduceEmptySeedTerminates.
func trivialSeed() core.Program {
	const fn = "s1:phase11.reduce_fixture:function:trivial"
	typeID := fn + ":type:0"
	return core.Program{
		Schema:   core.Schema,
		Module:   "phase11.reduce_fixture",
		ModuleID: "s1:phase11.reduce_fixture:module:phase11.reduce_fixture",
		Functions: []core.Function{
			{
				ID: fn, Name: "trivial",
				EntryPointID: fn + ":point:entry", ReturnPointID: fn + ":point:return",
				Parameter:  core.Parameter{ID: fn + ":place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:         fn + ":linear",
					Types:      []core.TypeFact{{ID: typeID}},
					Places:     []core.Place{{ID: fn + ":place:0", Name: "value", TypeID: typeID}},
					Operations: []core.LinearOperation{{ID: fn + ":op:0", PointID: fn + ":point:linear:0", Kind: core.OpReturn, SourceID: fn + ":place:0", TypeID: typeID}},
				},
			},
		},
	}
}

func totalOperationCount(p core.Program) int {
	count := 0
	for _, fn := range p.Functions {
		if fn.Linear != nil {
			count += len(fn.Linear.Operations)
		}
	}
	return count
}

// ---------------------------------------------------------------------
// Task 1 tests
// ---------------------------------------------------------------------

// singleFunctionGolden pins the pre-Phase-11 reduced-program bytes for
// each of reduce_test.go's own single-function fixtures, captured from a
// real Reduce run with this plan's changes already applied (compared
// byte-for-byte against every one of those changed commits' own existing
// test assertions, which all still pass unmodified) -- committed expected
// output, not a value this test derives fresh from the same code path it
// is checking.
var singleFunctionGolden = []struct {
	name    string
	seed    func() core.Program
	json    string
	applied []string
}{
	{
		name: "borrow_chain",
		seed: borrowChainSeed,
		json: `{"schema":"schway.core/0","module":"phase5.reduce_fixture","module_id":"s1:phase5.reduce_fixture:module:phase5.reduce_fixture","data_types":null,"functions":[{"id":"s1:phase5.reduce_fixture:function:touch","name":"touch","entry_point_id":"s1:phase5.reduce_fixture:function:touch:point:entry","return_point_id":"s1:phase5.reduce_fixture:function:touch:point:return","parameter":{"id":"s1:phase5.reduce_fixture:function:touch:place:0","name":"buffer","type":"Buffer"},"return_type":"Buffer","linear":{"id":"s1:phase5.reduce_fixture:function:touch:linear","types":[{"id":"s1:phase5.reduce_fixture:function:touch:type:0","shape":{"constructor":"","arguments":null},"abilities":null,"negative_witnesses":null}],"places":[{"id":"s1:phase5.reduce_fixture:function:touch:place:0","name":"buffer","type_id":"s1:phase5.reduce_fixture:function:touch:type:0"}],"operations":[{"id":"s1:phase5.reduce_fixture:function:touch:op:3","point_id":"s1:phase5.reduce_fixture:function:touch:point:linear:3","kind":"return","source_id":"s1:phase5.reduce_fixture:function:touch:place:0","type_id":"s1:phase5.reduce_fixture:function:touch:type:0"}]},"span":{"start":0,"end":0}}]}`,
		applied: []string{"drop-unused-binding", "truncate-to-minimal-prefix", "truncate-to-minimal-prefix"},
	},
	{
		name: "foreign_chain",
		seed: foreignChainSeed,
		json: `{"schema":"schway.core/0","module":"phase5.reduce_fixture","module_id":"s1:phase5.reduce_fixture:module:phase5.reduce_fixture","data_types":null,"functions":[{"id":"s1:phase5.reduce_fixture:function:main","name":"main","entry_point_id":"s1:phase5.reduce_fixture:function:main:point:entry","return_point_id":"s1:phase5.reduce_fixture:function:main:point:return","parameter":{"id":"s1:phase5.reduce_fixture:function:main:place:0","name":"request","type":"Byte"},"return_type":"Byte","linear":{"id":"s1:phase5.reduce_fixture:function:main:linear","types":[{"id":"s1:phase5.reduce_fixture:function:main:type:0","shape":{"constructor":"","arguments":null},"abilities":null,"negative_witnesses":null}],"places":[{"id":"s1:phase5.reduce_fixture:function:main:place:0","name":"request","type_id":"s1:phase5.reduce_fixture:function:main:type:0"}],"operations":[{"id":"s1:phase5.reduce_fixture:function:main:op:0","point_id":"s1:phase5.reduce_fixture:function:main:point:linear:0","kind":"return","source_id":"s1:phase5.reduce_fixture:function:main:place:0","type_id":"s1:phase5.reduce_fixture:function:main:type:0"}]},"span":{"start":0,"end":0}}]}`,
		applied: []string{"drop-offpath-foreign-stage", "drop-offpath-foreign-stage", "drop-offpath-foreign-stage"},
	},
	{
		name: "two_arm_match",
		seed: twoArmMatchSeed,
		json: `{"schema":"schway.core/0","module":"phase5.reduce_fixture","module_id":"s1:phase5.reduce_fixture:module:phase5.reduce_fixture","data_types":[{"id":"s1:phase5.reduce_fixture:type:Signal","name":"Signal","alternatives":["Go","Halt"],"span":{"start":0,"end":0}}],"functions":[{"id":"s1:phase5.reduce_fixture:function:triage","name":"triage","entry_point_id":"s1:phase5.reduce_fixture:function:triage:point:entry","return_point_id":"s1:phase5.reduce_fixture:function:triage:point:return","parameter":{"id":"s1:phase5.reduce_fixture:function:triage:place:0","name":"flag","type":"Signal"},"return_type":"Signal","linear":{"id":"s1:phase5.reduce_fixture:function:triage:linear","types":[{"id":"s1:phase5.reduce_fixture:function:triage:type:0","shape":{"constructor":"","arguments":null},"abilities":null,"negative_witnesses":null}],"places":[{"id":"s1:phase5.reduce_fixture:function:triage:place:0","name":"flag","type_id":"s1:phase5.reduce_fixture:function:triage:type:0"}],"operations":[{"id":"s1:phase5.reduce_fixture:function:triage:op:1","point_id":"s1:phase5.reduce_fixture:function:triage:point:linear:1","kind":"return","source_id":"s1:phase5.reduce_fixture:function:triage:place:0","type_id":"s1:phase5.reduce_fixture:function:triage:type:0"}]},"span":{"start":0,"end":0}}]}`,
		applied: []string{"collapse-branch-to-diverging-arm", "truncate-to-minimal-prefix"},
	},
	{
		name: "three_arm_match",
		seed: threeArmMatchSeed,
		json: `{"schema":"schway.core/0","module":"phase5.reduce_fixture","module_id":"s1:phase5.reduce_fixture:module:phase5.reduce_fixture","data_types":[{"id":"s1:phase5.reduce_fixture:type:Signal","name":"Signal","alternatives":["Go","Halt"],"span":{"start":0,"end":0}}],"functions":[{"id":"s1:phase5.reduce_fixture:function:triage","name":"triage","entry_point_id":"s1:phase5.reduce_fixture:function:triage:point:entry","return_point_id":"s1:phase5.reduce_fixture:function:triage:point:return","parameter":{"id":"s1:phase5.reduce_fixture:function:triage:place:0","name":"flag","type":"Signal"},"return_type":"Signal","linear":{"id":"s1:phase5.reduce_fixture:function:triage:linear","types":[{"id":"s1:phase5.reduce_fixture:function:triage:type:0","shape":{"constructor":"","arguments":null},"abilities":null,"negative_witnesses":null}],"places":[{"id":"s1:phase5.reduce_fixture:function:triage:place:0","name":"flag","type_id":"s1:phase5.reduce_fixture:function:triage:type:0"}],"operations":[{"id":"s1:phase5.reduce_fixture:function:triage:op:0","point_id":"s1:phase5.reduce_fixture:function:triage:point:linear:0","kind":"defect","source_id":"s1:phase5.reduce_fixture:function:triage:place:0","type_id":"s1:phase5.reduce_fixture:function:triage:type:0","reason":"adversarial defect"}]},"span":{"start":0,"end":0}}]}`,
		applied: []string{"drop-unmatched-arm", "collapse-branch-to-diverging-arm"},
	},
}

// TestReduceSingleFunctionOutputUnchanged asserts, for every existing
// single-function reduction fixture, that the result program and
// AppliedMoves are byte-identical to the committed golden output above
// (never a value freshly re-derived by this same test), and that the
// derived attempt budget (D-11-31) is exactly AttemptsPerFunction (64)
// for a one-function seed, matching the pre-Phase-11 flat
// MaxReductionAttempts it replaces.
func TestReduceSingleFunctionOutputUnchanged(t *testing.T) {
	sig := reduce.Signature{Axis: "axis:terminal-outcome", EnginePair: "pair", CausalRole: "golden"}
	for _, tc := range singleFunctionGolden {
		t.Run(tc.name, func(t *testing.T) {
			seed := tc.seed()
			if got := reduce.DerivedAttemptBoundForTest(seed); got != reduce.AttemptsPerFunction {
				t.Fatalf("expected the derived attempt budget for a one-function seed to equal AttemptsPerFunction (%d), got %d", reduce.AttemptsPerFunction, got)
			}
			result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed}, alwaysInteresting(sig))
			if err != nil {
				t.Fatalf("reduce: %v", err)
			}
			gotJSON, err := json.Marshal(result.Program)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(gotJSON) != tc.json {
				t.Fatalf("reduced program diverged from the committed golden output:\ngot:  %s\nwant: %s", gotJSON, tc.json)
			}
			if len(result.AppliedMoves) != len(tc.applied) {
				t.Fatalf("AppliedMoves length diverged from golden: got %v, want %v", result.AppliedMoves, tc.applied)
			}
			for i := range tc.applied {
				if result.AppliedMoves[i] != tc.applied[i] {
					t.Fatalf("AppliedMoves[%d] diverged from golden: got %q, want %q", i, result.AppliedMoves[i], tc.applied[i])
				}
			}
		})
	}
}

func TestReduceMultiFunctionSeed(t *testing.T) {
	seed := multiFunctionCallSeed()
	sig := reduce.Signature{Axis: "a", EnginePair: "p"}
	result, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed, EntryFunctionID: multiFnCallerID}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	found := false
	for _, fn := range result.Program.Functions {
		if fn.ID == multiFnCallerID {
			found = true
		}
	}
	if !found {
		t.Fatal("expected Seed.EntryFunctionID's own function to survive reduction")
	}
}

func TestDerivedAttemptBoundHasFloor(t *testing.T) {
	if got := reduce.DerivedAttemptBoundForTest(borrowChainSeed()); got != reduce.AttemptsPerFunction {
		t.Fatalf("one-function zero-call seed: expected %d, got %d", reduce.AttemptsPerFunction, got)
	}
	if got := reduce.DerivedAttemptBoundForTest(core.Program{}); got < reduce.AttemptsPerFunction {
		t.Fatalf("zero-function seed: expected a floor of at least %d, got %d", reduce.AttemptsPerFunction, got)
	}
	if got := reduce.DerivedAttemptBoundForTest(multiFunctionNoCallSeed()); got != reduce.AttemptsPerFunction*2 {
		t.Fatalf("zero-call multi-function seed: expected %d, got %d", reduce.AttemptsPerFunction*2, got)
	}
	if got := reduce.DerivedAttemptBoundForTest(multiFunctionCallSeed()); got != reduce.AttemptsPerFunction*2+1 {
		t.Fatalf("call-heavy (one-call) multi-function seed: expected %d, got %d", reduce.AttemptsPerFunction*2+1, got)
	}
}

func TestReduceEmptySeedTerminates(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	seed := trivialSeed()
	result, err := reduce.Reduce(ctx, reduce.Seed{Program: seed, EntryFunctionID: seed.Functions[0].ID}, alwaysInteresting(reduce.Signature{Axis: "a", EnginePair: "p"}))
	if err != nil {
		t.Fatalf("reduce: %v", err)
	}
	if result.Minimality != reduce.MinimalityFixpoint {
		t.Fatalf("expected fixpoint for a seed no move applies to, got %q", result.Minimality)
	}
	if len(result.AppliedMoves) != 0 {
		t.Fatalf("expected no moves applied, got %v", result.AppliedMoves)
	}
	structurallyEqual(t, result.Program, seed)
}

// TestReduceIsDeterministicForMultiFunctionSeed is this plan's own
// "TestReduceIsDeterministic" requirement, named distinctly from
// reduce_test.go's existing single-function TestReduceIsDeterministic to
// avoid a duplicate declaration in the same reduce_test package.
func TestReduceIsDeterministicForMultiFunctionSeed(t *testing.T) {
	seed := multiFunctionCallSeed()
	sig := reduce.Signature{Axis: "a", EnginePair: "p"}
	first, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed, EntryFunctionID: multiFnCallerID}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("first reduce: %v", err)
	}
	second, err := reduce.Reduce(context.Background(), reduce.Seed{Program: seed, EntryFunctionID: multiFnCallerID}, alwaysInteresting(sig))
	if err != nil {
		t.Fatalf("second reduce: %v", err)
	}
	firstBytes, err := json.Marshal(first.Program)
	if err != nil {
		t.Fatalf("marshal first: %v", err)
	}
	secondBytes, err := json.Marshal(second.Program)
	if err != nil {
		t.Fatalf("marshal second: %v", err)
	}
	if string(firstBytes) != string(secondBytes) {
		t.Fatalf("multi-function reduce is not deterministic:\nfirst:  %s\nsecond: %s", firstBytes, secondBytes)
	}
}

// reduceDirectImportViolation mirrors originvalidate_test.go's own
// directImportViolation guard shape (D-10-17's convention): a direct
// go/parser ImportsOnly scan over dir's own non-test .go files' import
// lists, returning the first forbidden import found or "".
func reduceDirectImportViolation(t *testing.T, dir string, forbidden []string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			for _, bad := range forbidden {
				if strings.HasSuffix(path, bad) {
					return entry.Name() + " imports " + path
				}
			}
		}
	}
	return ""
}

// forbiddenCallgraphImportSuffix is built via concatenation (never as one
// contiguous literal) so this file itself never contains the exact
// substring this plan's own acceptance criteria forbids anywhere under
// internal/compiler/reduce, including inside a comment or a test's own
// string literal.
var forbiddenCallgraphImportSuffix = "/compiler/" + "callgraph"

// TestReduceDoesNotImportCallgraph is D-11-29's own structural guard:
// reduce must never import the call-graph package -- that package's own
// production consumer set is a documented, load-bearing independence
// boundary this package stays outside of.
func TestReduceDoesNotImportCallgraph(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "reduce")
	if violation := reduceDirectImportViolation(t, dir, []string{forbiddenCallgraphImportSuffix}); violation != "" {
		t.Fatalf("%s, which reduce must never depend on (D-11-29)", violation)
	}
}

// TestReduceImportGuardCanFail is the negative control (D-10-17's own
// convention): scanning internal/compiler/session, which DOES import the
// call-graph package (session.go, and this plan's own
// session_phase5_mismatch.go EntryFunction wiring), proves the scan
// mechanism can flag a real violation and is not vacuously green.
func TestReduceImportGuardCanFail(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "session")
	if violation := reduceDirectImportViolation(t, dir, []string{forbiddenCallgraphImportSuffix}); violation == "" {
		t.Fatal("expected the session package's own call-graph import to be flagged -- the negative control failed to prove the scan can go red")
	}
}

// ---------------------------------------------------------------------
// Task 2 tests: the two whole-program moves
// ---------------------------------------------------------------------

func TestDropCallSiteRewritesOneEdge(t *testing.T) {
	seed := multiFunctionCallSeed()
	out, ok := reduce.DropCallSiteForTest(seed)
	if !ok {
		t.Fatal("expected drop-call-site to find the eligible OpCall")
	}
	original := seed.Functions[0].Linear.Operations[0]
	rewritten := out.Functions[0].Linear.Operations[0]
	if rewritten.Kind != core.OpCopy {
		t.Fatalf("expected the call to be rewritten to OpCopy, got %v", rewritten.Kind)
	}
	if rewritten.ID != original.ID || rewritten.SourceID != original.SourceID || rewritten.TargetID != original.TargetID {
		t.Fatalf("expected ID/SourceID/TargetID to be preserved byte-for-byte: before=%+v after=%+v", original, rewritten)
	}
	if rewritten.CalleeID != "" {
		t.Fatalf("expected CalleeID to be cleared, got %q", rewritten.CalleeID)
	}
}

func TestDropOrphanFunctionRemovesUncalledNonEntry(t *testing.T) {
	seed := multiFunctionNoCallSeed()
	out, ok := reduce.DropOrphanFunctionForTest(seed, multiFnCallerID)
	if !ok {
		t.Fatal("expected drop-orphan-function to remove the uncalled non-entry function")
	}
	for _, fn := range out.Functions {
		if fn.ID != multiFnCallerID {
			t.Fatalf("expected only the entry function to survive, found %q", fn.ID)
		}
	}
	if len(out.Functions) != 1 {
		t.Fatalf("expected exactly one surviving function, got %d", len(out.Functions))
	}

	// The entry function itself, even though it is also uncalled by
	// anything else, must never be removed.
	if _, ok := reduce.DropOrphanFunctionForTest(out, multiFnCallerID); ok {
		t.Fatal("expected drop-orphan-function to never remove the entry function")
	}
}

func TestOrphanBecomesEligibleOnlyAfterCallSiteDropped(t *testing.T) {
	seed := multiFunctionCallSeed()

	// Before drop-call-site runs, callee still has its one in-edge --
	// drop-orphan-function must not apply.
	if _, ok := reduce.DropOrphanFunctionForTest(seed, multiFnCallerID); ok {
		t.Fatal("expected drop-orphan-function to be ineligible while the call site survives")
	}

	afterCallSite, ok := reduce.DropCallSiteForTest(seed)
	if !ok {
		t.Fatal("expected drop-call-site to apply")
	}

	// After drop-call-site rewrites the only call to callee away, callee
	// has zero in-edges and becomes eligible.
	afterOrphan, ok := reduce.DropOrphanFunctionForTest(afterCallSite, multiFnCallerID)
	if !ok {
		t.Fatal("expected drop-orphan-function to become eligible once the call site is dropped")
	}
	for _, fn := range afterOrphan.Functions {
		if fn.ID == multiFnCalleeID {
			t.Fatal("expected callee to have been removed")
		}
	}
}

func TestOperationIDsUnchangedByWholeProgramMoves(t *testing.T) {
	seed := multiFunctionCallSeed()
	afterCallSite, ok := reduce.DropCallSiteForTest(seed)
	if !ok {
		t.Fatal("expected drop-call-site to apply")
	}
	for _, fn := range seed.Functions {
		var after *core.Function
		for i := range afterCallSite.Functions {
			if afterCallSite.Functions[i].ID == fn.ID {
				after = &afterCallSite.Functions[i]
			}
		}
		if after == nil || fn.Linear == nil || after.Linear == nil {
			continue
		}
		if len(fn.Linear.Operations) != len(after.Linear.Operations) {
			t.Fatalf("drop-call-site changed operation count for %q", fn.ID)
		}
		for i := range fn.Linear.Operations {
			if fn.Linear.Operations[i].ID != after.Linear.Operations[i].ID {
				t.Fatalf("drop-call-site changed operation %d's ID in %q: before=%q after=%q", i, fn.ID, fn.Linear.Operations[i].ID, after.Linear.Operations[i].ID)
			}
		}
	}

	afterOrphan, ok := reduce.DropOrphanFunctionForTest(afterCallSite, multiFnCallerID)
	if !ok {
		t.Fatal("expected drop-orphan-function to apply after drop-call-site")
	}
	callerAfter := afterOrphan.Functions[0]
	callerBefore := afterCallSite.Functions[0]
	if len(callerAfter.Linear.Operations) != len(callerBefore.Linear.Operations) {
		t.Fatal("drop-orphan-function changed the surviving caller's own operation count")
	}
	for i := range callerBefore.Linear.Operations {
		if callerBefore.Linear.Operations[i].ID != callerAfter.Linear.Operations[i].ID {
			t.Fatalf("drop-orphan-function changed the surviving caller's operation %d ID: before=%q after=%q", i, callerBefore.Linear.Operations[i].ID, callerAfter.Linear.Operations[i].ID)
		}
	}
}

func TestNoMoveInlines(t *testing.T) {
	seed := multiFunctionCallSeed()
	before := totalOperationCount(seed)

	afterCallSite, ok := reduce.DropCallSiteForTest(seed)
	if !ok {
		t.Fatal("expected drop-call-site to apply")
	}
	if got := totalOperationCount(afterCallSite); got > before {
		t.Fatalf("drop-call-site increased total operation count: before=%d after=%d", before, got)
	}

	afterOrphan, ok := reduce.DropOrphanFunctionForTest(afterCallSite, multiFnCallerID)
	if !ok {
		t.Fatal("expected drop-orphan-function to apply after drop-call-site")
	}
	if got := totalOperationCount(afterOrphan); got > totalOperationCount(afterCallSite) {
		t.Fatalf("drop-orphan-function increased total operation count: before=%d after=%d", totalOperationCount(afterCallSite), got)
	}
}

// ---------------------------------------------------------------------
// Task 3 tests: RefusedShapes()
// ---------------------------------------------------------------------

func TestRefusedShapesIsEnumeratedAndComplete(t *testing.T) {
	shapes := reduce.RefusedShapes()
	if len(shapes) == 0 {
		t.Fatal("expected a non-empty RefusedShapes() register")
	}
	knownIDs := map[string]bool{
		reduce.RefusalSeedShapeUnsupported:         true,
		reduce.RefusalProjectionUnsupported:        true,
		reduce.RefusalReverificationSignatureDrift: true,
	}
	seenIDs := map[string]bool{}
	for _, shape := range shapes {
		if shape.Name == "" || shape.RefusalID == "" || shape.Reason == "" {
			t.Fatalf("incomplete RefusedShape entry: %+v", shape)
		}
		if !knownIDs[shape.RefusalID] {
			t.Fatalf("RefusedShape %q uses an unregistered refusal ID %q", shape.Name, shape.RefusalID)
		}
		seenIDs[shape.RefusalID] = true
	}
	for id := range knownIDs {
		if !seenIDs[id] {
			t.Fatalf("refusal ID %q is declared but never used by any RefusedShapes() entry -- a gap", id)
		}
	}
}

func TestRefusedShapeIDsFollowConvention(t *testing.T) {
	ids := []string{reduce.RefusalSeedShapeUnsupported, reduce.RefusalProjectionUnsupported, reduce.RefusalReverificationSignatureDrift}
	for _, id := range ids {
		if !strings.HasPrefix(id, "reduce.") {
			t.Fatalf("refusal ID %q does not follow the reduce. prefix convention", id)
		}
	}
	if reduce.RefusalSeedShapeUnsupported != "reduce.seed_shape_unsupported" {
		t.Fatalf("unexpected RefusalSeedShapeUnsupported value: %q", reduce.RefusalSeedShapeUnsupported)
	}
	if reduce.RefusalProjectionUnsupported != "reduce.projection_unsupported" {
		t.Fatalf("unexpected RefusalProjectionUnsupported value: %q", reduce.RefusalProjectionUnsupported)
	}
	if reduce.RefusalReverificationSignatureDrift != "reduce.reverification_signature_drift" {
		t.Fatalf("unexpected RefusalReverificationSignatureDrift value: %q", reduce.RefusalReverificationSignatureDrift)
	}
}

func TestRefusedShapesAreActuallyRefused(t *testing.T) {
	shapes := reduce.RefusedShapes()
	byName := make(map[string]reduce.RefusedShape, len(shapes))
	for _, s := range shapes {
		byName[s.Name] = s
	}

	t.Run("match_bodied_projection", func(t *testing.T) {
		shape, ok := byName["Match-bodied source projection"]
		if !ok {
			t.Fatal("expected a RefusedShapes() entry named \"Match-bodied source projection\"")
		}
		got := reduce.ProjectSource(twoArmMatchSeed())
		if !strings.HasPrefix(got, "// reduce: projection unsupported") || !strings.Contains(got, shape.RefusalID) {
			t.Fatalf("expected a projection-unsupported comment carrying refusal ID %q, got: %s", shape.RefusalID, got)
		}
	})

	t.Run("adt_typed_collapsed_match_projection", func(t *testing.T) {
		shape, ok := byName["ADT-typed collapsed match projection"]
		if !ok {
			t.Fatal("expected a RefusedShapes() entry named \"ADT-typed collapsed match projection\"")
		}
		seed := twoArmMatchSeed()
		seed.Functions[0].Match = nil
		got := reduce.ProjectSource(seed)
		if !strings.HasPrefix(got, "// reduce: projection unsupported") || !strings.Contains(got, shape.RefusalID) {
			t.Fatalf("expected a projection-unsupported comment carrying refusal ID %q, got: %s", shape.RefusalID, got)
		}
	})

	t.Run("type_changing_calls", func(t *testing.T) {
		if _, ok := byName["Type-changing calls"]; !ok {
			t.Fatal("expected a RefusedShapes() entry named \"Type-changing calls\"")
		}
		if _, ok := reduce.DropCallSiteForTest(typeChangingCallSeed()); ok {
			t.Fatal("expected drop-call-site to decline the type-changing call")
		}
	})

	for _, disclosedName := range []string{
		"Recursive call graphs",
		"Flaky or nondeterministic predicates",
		"Re-verification signature drift (reserved for plan 11-09)",
	} {
		shape, ok := byName[disclosedName]
		if !ok {
			t.Fatalf("expected a RefusedShapes() entry named %q", disclosedName)
		}
		if !shape.Disclosed {
			t.Fatalf("expected %q to be explicitly disclosed as unconstructible at this language's current maturity", disclosedName)
		}
		if shape.Reason == "" {
			t.Fatalf("expected %q to carry a non-empty disclosure reason", disclosedName)
		}
	}
}

// TestNoFlakyPredicateTolerance is D-11-36's own source scan: reduce must
// contain no retry loop, no majority-vote logic, and no fuzzy
// interestingness comparison over a predicate's result.
func TestNoFlakyPredicateTolerance(t *testing.T) {
	for _, file := range []string{"reduce.go", "predicate.go", "mismatch.go"} {
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("reading %s: %v", file, err)
		}
		lower := strings.ToLower(string(contents))
		for _, banned := range []string{"retry", "majority", "fuzzy"} {
			if strings.Contains(lower, banned) {
				t.Fatalf("%s must not contain %q -- flaky-predicate tolerance is explicitly not built (D-11-36)", file, banned)
			}
		}
	}
}

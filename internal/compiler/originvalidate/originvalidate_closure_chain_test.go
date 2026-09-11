package originvalidate_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// deepDiamondProgram is 07-06 Task 3's own multi-level, shared-leaf corpus
// (testdata/phase07/deep_diamond_acyclic.lang): four chained diamonds, each
// with a node calling two distinct successors that both reach a shared
// successor -- exactly the shape a closure-digest chain needs to exercise
// depth and callee sharing (D-07-38's read_first note).
func deepDiamondProgram(t testing.TB) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "deep_diamond_acyclic.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("deep_diamond_acyclic.lang: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func signaturesByID(summary core.Interface) map[string]core.FunctionSignature {
	out := make(map[string]core.FunctionSignature, len(summary.Functions))
	for _, function := range summary.Functions {
		out[function.ID] = function
	}
	return out
}

// TestClosureDigestComputationOrderIsCalleeBeforeCaller is Task 1 Test 1
// (D-07-38): BuildInterface computes every function's ClosureDigest in an
// order where every callee is finished before its caller -- the exact
// reverse of callgraph.Order's own returned array (which hands back
// callers first, callees last -- see callgraph.Order's own doc comment and
// TestOrderSortsAdjacencyByCalleeID's worked example). A test asserts the
// observed computation order equals callgraph.Order's return value,
// reversed, element for element.
func TestClosureDigestComputationOrderIsCalleeBeforeCaller(t *testing.T) {
	program := deepDiamondProgram(t)

	callgraphOrder, err := callgraph.Order(program)
	if err != nil {
		t.Fatalf("callgraph.Order: %v", err)
	}
	wantComputationOrder := make([]string, len(callgraphOrder))
	for i, id := range callgraphOrder {
		wantComputationOrder[len(callgraphOrder)-1-i] = id
	}

	var gotComputationOrder []string
	restore := originvalidate.SetClosureDigestComputationOrderObservedForTest(func(functionID string) {
		gotComputationOrder = append(gotComputationOrder, functionID)
	})
	defer restore()

	if _, err := originvalidate.BuildInterface(program); err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}

	if !reflect.DeepEqual(gotComputationOrder, wantComputationOrder) {
		t.Fatalf("digest computation order mismatch:\ngot=%v\nwant (reverse of callgraph.Order)=%v", gotComputationOrder, wantComputationOrder)
	}
}

// TestClosureDigestZeroCalleeLeafUnperturbedByContext is Task 1 Test 2: a
// zero-callee function's ClosureDigest is unchanged whether it is computed
// standalone or inside a larger program -- the chaining arm never perturbs
// a leaf, because a leaf's own preimage callee list is empty either way.
func TestClosureDigestZeroCalleeLeafUnperturbedByContext(t *testing.T) {
	program := deepDiamondProgram(t)
	var leaf core.Function
	found := false
	for _, function := range program.Functions {
		if function.Name == "leaf" {
			leaf, found = function, true
			break
		}
	}
	if !found {
		t.Fatal("expected deep_diamond_acyclic.lang to declare a function named leaf")
	}

	fullSummary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(full program): %v", err)
	}
	inContext, ok := signaturesByID(fullSummary)[leaf.ID]
	if !ok {
		t.Fatalf("leaf %s missing from full-program summary", leaf.ID)
	}

	standalone := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, Functions: []core.Function{leaf}}
	standaloneSummary, err := originvalidate.BuildInterface(standalone)
	if err != nil {
		t.Fatalf("BuildInterface(standalone leaf): %v", err)
	}
	if len(standaloneSummary.Functions) != 1 {
		t.Fatalf("expected exactly one function in the standalone summary, got %d", len(standaloneSummary.Functions))
	}

	if inContext.ClosureDigest != standaloneSummary.Functions[0].ClosureDigest {
		t.Fatalf("expected leaf's ClosureDigest to be unperturbed by context: in-context=%q standalone=%q", inContext.ClosureDigest, standaloneSummary.Functions[0].ClosureDigest)
	}
}

// TestClosureDigestSharedLeafConsistentAcrossParents is Task 1 Test 3: on
// deep_diamond_acyclic.lang, mid1a and mid1b are distinct parents that both
// call the shared leaf `leaf`. Their own digests differ from each other
// (different IDs/Names), but BOTH must change identically in response to a
// change in leaf's own published signature -- proving the shared leaf's
// digest genuinely propagates into both parents' preimages, rather than
// only one of them (or neither).
func TestClosureDigestSharedLeafConsistentAcrossParents(t *testing.T) {
	program := deepDiamondProgram(t)
	leafIndex, mid1aID, mid1bID := -1, "", ""
	for i, function := range program.Functions {
		switch function.Name {
		case "leaf":
			leafIndex = i
		case "mid1a":
			mid1aID = function.ID
		case "mid1b":
			mid1bID = function.ID
		}
	}
	if leafIndex == -1 || mid1aID == "" || mid1bID == "" {
		t.Fatal("expected deep_diamond_acyclic.lang to declare leaf, mid1a, and mid1b")
	}

	baseSummary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(base): %v", err)
	}
	baseByID := signaturesByID(baseSummary)
	mid1aBase, mid1bBase := baseByID[mid1aID], baseByID[mid1bID]
	if mid1aBase.ClosureDigest == mid1bBase.ClosureDigest {
		t.Fatalf("expected mid1a and mid1b (distinct functions) to have distinct ClosureDigests, both got %q", mid1aBase.ClosureDigest)
	}

	// Mutate ONLY leaf's own published signature (a declared origin
	// changes its Return.Mode/Paths -- BuildInterface packages whatever is
	// present, regardless of whether a real checker would have derived it,
	// exactly as corevalidate_mutation_matrix_test.go's own synthetic
	// programs do), leaving every other function's core.Function struct
	// byte-identical.
	mutated := cloneProgramForTest(t, program)
	mutated.Functions[leafIndex].PublicOrigin = &core.PublicOrigin{Paths: []string{mutated.Functions[leafIndex].Parameter.Name}, Access: "shared"}

	mutatedSummary, err := originvalidate.BuildInterface(mutated)
	if err != nil {
		t.Fatalf("BuildInterface(mutated leaf): %v", err)
	}
	mutatedByID := signaturesByID(mutatedSummary)
	mid1aMutated, mid1bMutated := mutatedByID[mid1aID], mutatedByID[mid1bID]

	if mid1aMutated.ClosureDigest == mid1aBase.ClosureDigest {
		t.Fatal("expected mid1a's ClosureDigest to change when the shared leaf's signature changed")
	}
	if mid1bMutated.ClosureDigest == mid1bBase.ClosureDigest {
		t.Fatal("expected mid1b's ClosureDigest to change when the shared leaf's signature changed")
	}
}

// TestBuildInterfaceRefusesCyclicGraphComputingNoDigest is Task 1 Test 4
// (D-07-38): BuildInterface runs callgraph.Order first and, on a cyclic
// graph, returns that error unchanged and computes NO digest at all --
// never attempting to chain over a graph that is not a proven DAG.
// Constructed directly against a hand-built, forged core.Program (never
// through the parser or check, which already refuses a cycle on its own
// admission path) -- BuildInterface must behave correctly on a program no
// producer would ever hand it, exactly like callgraph.Order's own T-07-35
// posture.
func TestBuildInterfaceRefusesCyclicGraphComputingNoDigest(t *testing.T) {
	program := mutualCycleProgramForTest()
	summary, err := originvalidate.BuildInterface(program)
	if err == nil {
		t.Fatalf("expected BuildInterface to refuse a cyclic program, got a summary: %+v", summary)
	}
	if _, ok := callgraph.CycleError(err); !ok {
		t.Fatalf("expected callgraph's own cycle error, got %v (%T)", err, err)
	}
	if len(summary.Functions) != 0 {
		t.Fatalf("expected no digest computed on a cyclic program, got %d function signatures", len(summary.Functions))
	}
}

// TestClosureDigestDeterministicAcrossRuns is Task 1 Test 5 (the
// in-process half; go test -count=2 -shuffle=on covers the cross-run half
// of the same claim at the package level): running BuildInterface twice on
// the identical checked program yields byte-identical ClosureDigest values
// for every function.
func TestClosureDigestDeterministicAcrossRuns(t *testing.T) {
	program := deepDiamondProgram(t)
	first, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface (first run): %v", err)
	}
	second, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface (second run): %v", err)
	}
	firstByID, secondByID := signaturesByID(first), signaturesByID(second)
	if len(firstByID) != len(secondByID) {
		t.Fatalf("expected the same function count across runs: %d vs %d", len(firstByID), len(secondByID))
	}
	for id, firstSignature := range firstByID {
		secondSignature, ok := secondByID[id]
		if !ok {
			t.Fatalf("function %s missing from second run", id)
		}
		if firstSignature.ClosureDigest != secondSignature.ClosureDigest {
			t.Fatalf("function %s: ClosureDigest differs across runs: %q vs %q", id, firstSignature.ClosureDigest, secondSignature.ClosureDigest)
		}
	}
}

// TestClosureDigestDeterministicAcrossCorpus is Task 1's acceptance
// criterion "two runs and two processes produce byte-identical
// ClosureDigest values for every function of every testdata/phase07
// accepting fixture": over every testdata/phase07 fixture that checks
// clean, BuildInterface run twice on the identical checked program
// produces byte-identical ClosureDigest values for every function. The
// cross-PROCESS half of the same claim is a property of computeClosureDigest
// being a pure function of its inputs (sha256 over a sorted, map-free
// preimage) -- go test -count=2 -shuffle=on (a fresh process boundary
// is not needed to observe a map-iteration-order or goroutine-scheduling
// leak, since none of this package's own code depends on either) already
// exercises this determinism claim at the package level.
func TestClosureDigestDeterministicAcrossCorpus(t *testing.T) {
	dir := testsupport.ProjectPath("testdata", "phase07")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	checkedAny := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
			continue
		}
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			continue
		}
		first, err := originvalidate.BuildInterface(checked.Program)
		if err != nil {
			t.Fatalf("%s: BuildInterface (first run): %v", entry.Name(), err)
		}
		second, err := originvalidate.BuildInterface(checked.Program)
		if err != nil {
			t.Fatalf("%s: BuildInterface (second run): %v", entry.Name(), err)
		}
		firstByID, secondByID := signaturesByID(first), signaturesByID(second)
		for id, firstSignature := range firstByID {
			checkedAny = true
			secondSignature, ok := secondByID[id]
			if !ok {
				t.Fatalf("%s: function %s missing from second run", entry.Name(), id)
			}
			if firstSignature.ClosureDigest != secondSignature.ClosureDigest {
				t.Fatalf("%s: function %s: ClosureDigest differs across runs: %q vs %q", entry.Name(), id, firstSignature.ClosureDigest, secondSignature.ClosureDigest)
			}
		}
	}
	if !checkedAny {
		t.Fatal("expected at least one accepting testdata/phase07 fixture to be compared")
	}
}

// TestClosureDigestSummaryCarriesNoEdgeList is Task 1 Test 6 (D-07-11):
// over a multi-function program, the marshalled lang.interface/1 document
// contains no callee_id field at all -- the closure-digest chain reads
// call-graph edges to BUILD a preimage, but never publishes them. Edges
// stay in callgraph; the summary carries only closure-derived scalars.
func TestClosureDigestSummaryCarriesNoEdgeList(t *testing.T) {
	program := deepDiamondProgram(t)
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if bytes.Contains(encoded, []byte(`"callee_id"`)) {
		t.Fatalf("interface summary leaked a callee_id edge field: %s", encoded)
	}
}

// TestCalleeChangeInvalidatesCallerClosureDigest is Task 2 Test 1 -- the
// standing verdict made falsifiable, and the entire reason ClosureDigest
// exists: changing ONLY the callee's published signature changes the
// caller's ClosureDigest, while the caller's own core.Function is
// byte-identical across both runs.
func TestCalleeChangeInvalidatesCallerClosureDigest(t *testing.T) {
	base := straightLineCallProgramForTest(false)
	baseSummary, err := originvalidate.BuildInterface(base)
	if err != nil {
		t.Fatalf("BuildInterface(base): %v", err)
	}

	mutated := straightLineCallProgramForTest(true)
	mutatedSummary, err := originvalidate.BuildInterface(mutated)
	if err != nil {
		t.Fatalf("BuildInterface(mutated): %v", err)
	}

	callerFunction, calleeFunction := findByName(t, base, "caller"), findByName(t, base, "callee")
	mutatedCallerFunction := findByName(t, mutated, "caller")
	if !reflect.DeepEqual(callerFunction, mutatedCallerFunction) {
		t.Fatalf("expected caller's own core.Function to be byte-identical across both runs:\nbase=%+v\nmutated=%+v", callerFunction, mutatedCallerFunction)
	}

	baseByID, mutatedByID := signaturesByID(baseSummary), signaturesByID(mutatedSummary)
	if baseByID[calleeFunction.ID].ClosureDigest == mutatedByID[calleeFunction.ID].ClosureDigest {
		t.Fatal("expected the callee's own ClosureDigest to change when its published signature changed")
	}
	if baseByID[callerFunction.ID].ClosureDigest == mutatedByID[callerFunction.ID].ClosureDigest {
		t.Fatal("expected the caller's ClosureDigest to change when its callee's signature changed -- this is the whole point of the chain (D-07-12)")
	}
}

// TestUnrelatedFunctionChangeDoesNotInvalidateCallerClosureDigest is Task 2
// Test 2: the chain is precise, not a whole-program hash. Changing a
// function the caller never calls leaves the caller's ClosureDigest
// unchanged.
func TestUnrelatedFunctionChangeDoesNotInvalidateCallerClosureDigest(t *testing.T) {
	base := straightLineCallProgramForTest(false)
	baseSummary, err := originvalidate.BuildInterface(base)
	if err != nil {
		t.Fatalf("BuildInterface(base): %v", err)
	}

	mutated := cloneProgramForTest(t, base)
	for i := range mutated.Functions {
		if mutated.Functions[i].Name == "unrelated" {
			mutated.Functions[i].PublicOrigin = &core.PublicOrigin{Paths: []string{mutated.Functions[i].Parameter.Name}, Access: "exclusive"}
		}
	}
	mutatedSummary, err := originvalidate.BuildInterface(mutated)
	if err != nil {
		t.Fatalf("BuildInterface(mutated): %v", err)
	}

	callerFunction := findByName(t, base, "caller")
	baseByID, mutatedByID := signaturesByID(baseSummary), signaturesByID(mutatedSummary)
	if baseByID[callerFunction.ID].ClosureDigest != mutatedByID[callerFunction.ID].ClosureDigest {
		t.Fatal("expected the caller's ClosureDigest to stay unchanged when an UNRELATED function (never called by the caller) changed -- the chain must not degenerate into a whole-program hash")
	}
}

// foreignReachOnlyCallProgramForTest is a 10-02 (D-10-01) variant of
// straightLineCallProgramForTest that isolates the SAME
// caller-calls-callee shape but toggles ONLY callee's ForeignContract
// (never PublicOrigin): once walkReturnOrigin gained a real core.OpCall
// case, toggling callee's declared origin ALSO flips caller's own
// Callable bit (caller's body directly forwards its own parameter into
// the call, so a callee that genuinely declares a borrow-of-parameter
// return makes caller's own undeclared return newly derived, hence
// core.origin_omitted) -- a second, uncontrolled variable
// straightLineCallProgramForTest's own doc comment did not anticipate
// when it was authored (it predates this plan's cross-call propagation).
// Toggling ForeignContract instead changes callee's published Foreign/Fails
// (and therefore its own ClosureDigest) while leaving PublicOrigin nil in
// both variants, so calleeContracts never carries callee at all and
// caller's own Callable/Return.Mode stay constant across both variants --
// exactly the isolation TestClosureDigestEmptyCalleesMutationKilled needs.
func foreignReachOnlyCallProgramForTest(calleeHasForeignReach bool) core.Program {
	const callerID, calleeID = "s1:test:fn:caller", "s1:test:fn:callee"
	caller := straightLineFunction(callerID, calleeID)
	callee := straightLineFunction(calleeID, "")
	if calleeHasForeignReach {
		callee.ForeignContract = &core.ForeignContract{
			Symbol: "probe", Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden",
		}
	}
	return core.Program{
		Schema: core.Schema1, Module: "test.closure_chain_empty_callees", ModuleID: "s1:test:module:closure_chain_empty_callees",
		Functions: []core.Function{caller, callee},
	}
}

// TestClosureDigestEmptyCalleesMutationKilled is Task 2 Test 4 (QLT-08,
// D-07-41): with closureDigestEmptyCalleesOverride engaged, the
// callee-changes-invalidates-caller property (the previous test) goes
// red -- the caller's digest stops reacting to its callee changing,
// because the caller's own preimage never referenced any callee pair to
// begin with. Restoring the seam restores the property.
func TestClosureDigestEmptyCalleesMutationKilled(t *testing.T) {
	base := foreignReachOnlyCallProgramForTest(false)
	mutatedCallee := foreignReachOnlyCallProgramForTest(true)
	callerFunction := findByName(t, base, "caller")

	restore := originvalidate.SetClosureDigestEmptyCalleesOverrideForTest(true)
	baseSummary, err := originvalidate.BuildInterface(base)
	if err != nil {
		restore()
		t.Fatalf("BuildInterface(base): %v", err)
	}
	mutatedSummary, err := originvalidate.BuildInterface(mutatedCallee)
	restore()
	if err != nil {
		t.Fatalf("BuildInterface(mutated): %v", err)
	}

	baseByID, mutatedByID := signaturesByID(baseSummary), signaturesByID(mutatedSummary)
	if baseByID[callerFunction.ID].ClosureDigest != mutatedByID[callerFunction.ID].ClosureDigest {
		t.Fatal("mutation (forcing empty callee pairs) had no observable effect: caller's digest still changed when its callee changed")
	}

	// Restored: the property comes back.
	restoredBaseSummary, err := originvalidate.BuildInterface(base)
	if err != nil {
		t.Fatalf("BuildInterface(base, restored): %v", err)
	}
	restoredMutatedSummary, err := originvalidate.BuildInterface(mutatedCallee)
	if err != nil {
		t.Fatalf("BuildInterface(mutated, restored): %v", err)
	}
	restoredBaseByID, restoredMutatedByID := signaturesByID(restoredBaseSummary), signaturesByID(restoredMutatedSummary)
	if restoredBaseByID[callerFunction.ID].ClosureDigest == restoredMutatedByID[callerFunction.ID].ClosureDigest {
		t.Fatal("expected the callee-changes-invalidates-caller property to be restored once the seam is disabled")
	}
}

// TestClosureDigestDiscoveryOrderMutationKilled is Task 2 Test 5
// (D-07-38's ordering claim, D-07-41): with
// closureDigestDiscoveryOrderOverride engaged, a program is chained in
// plain declaration order instead of the proven-correct callee-before-
// caller order. On straightLineCallProgramForTest, caller is declared
// BEFORE callee, so under the mutation the caller reads callee's
// still-empty ClosureDigest -- an observably DIFFERENT (and wrong) digest
// than the correctly-ordered computation -- proving the ordering itself is
// load-bearing.
func TestClosureDigestDiscoveryOrderMutationKilled(t *testing.T) {
	program := straightLineCallProgramForTest(false)
	callerFunction := findByName(t, program, "caller")
	if declarationIndex(program, "caller") > declarationIndex(program, "callee") {
		t.Fatal("test fixture assumption violated: caller must be declared before callee")
	}

	correct, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(correct order): %v", err)
	}

	restore := originvalidate.SetClosureDigestDiscoveryOrderOverrideForTest(true)
	defer restore()
	mutated, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(discovery order): %v", err)
	}

	correctByID, mutatedByID := signaturesByID(correct), signaturesByID(mutated)
	if correctByID[callerFunction.ID].ClosureDigest == mutatedByID[callerFunction.ID].ClosureDigest {
		t.Fatal("mutation (declaration-order chaining) had no observable effect: caller's digest was unchanged despite being chained before its callee's digest existed")
	}
}

// --- test fixtures -----------------------------------------------------

func findByName(t testing.TB, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("expected program to declare a function named %q", name)
	return core.Function{}
}

func declarationIndex(program core.Program, name string) int {
	for i, function := range program.Functions {
		if function.Name == name {
			return i
		}
	}
	return -1
}

func cloneProgramForTest(t testing.TB, program core.Program) core.Program {
	t.Helper()
	data, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("marshal for clone: %v", err)
	}
	var clone core.Program
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatalf("unmarshal for clone: %v", err)
	}
	return clone
}

// straightLineOperations builds one function's whole straight-line body: a
// single OpReturn of its own parameter, or (when calleeID is non-empty) an
// OpCall to calleeID followed by an OpReturn of the call's result.
func straightLineOperations(functionID, calleeID string) []core.LinearOperation {
	if calleeID == "" {
		return []core.LinearOperation{
			{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpReturn, SourceID: functionID + ":place:0"},
		}
	}
	return []core.LinearOperation{
		{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpCall, SourceID: functionID + ":place:0", TargetID: functionID + ":place:1", CalleeID: calleeID},
		{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpReturn, SourceID: functionID + ":place:1"},
	}
}

func straightLineFunction(id, calleeID string) core.Function {
	return core.Function{
		ID: id, Name: id[len("s1:test:fn:"):],
		EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter:  core.Parameter{ID: id + ":place:0", Name: "value", Type: "Byte"},
		ReturnType: "Byte",
		Linear:     &core.LinearBody{ID: id + ":linear", Operations: straightLineOperations(id, calleeID)},
	}
}

// straightLineCallProgramForTest is Task 2's own hand-built, forged
// three-function core.Program (never through the parser or check, exactly
// like corevalidate_mutation_matrix_test.go's syntheticCallToNonCallableCalleeProgram):
// `caller` (declared FIRST) calls `callee`; `unrelated` calls nothing and
// is never called. Building it directly, rather than through a `.lang`
// fixture, gives full control over exactly which function's published
// signature differs between two otherwise-identical programs.
// calleeHasDeclaredOrigin toggles ONLY callee's PublicOrigin (a
// signature-affecting fact BuildInterface reads directly, standing in for
// "the callee's body changed" -- BuildInterface never validates that a
// declared origin is body-derivable; that is ValidatePublished's job, not
// exercised here).
func straightLineCallProgramForTest(calleeHasDeclaredOrigin bool) core.Program {
	const callerID, calleeID, unrelatedID = "s1:test:fn:caller", "s1:test:fn:callee", "s1:test:fn:unrelated"
	caller := straightLineFunction(callerID, calleeID)
	callee := straightLineFunction(calleeID, "")
	if calleeHasDeclaredOrigin {
		callee.PublicOrigin = &core.PublicOrigin{Paths: []string{"value"}, Access: "shared"}
	}
	unrelated := straightLineFunction(unrelatedID, "")
	return core.Program{
		Schema: core.Schema1, Module: "test.closure_chain", ModuleID: "s1:test:module:closure_chain",
		Functions: []core.Function{caller, callee, unrelated},
	}
}

// mutualCycleProgramForTest hand-builds a two-function core.Program whose
// OpCall edges form a genuine cycle (a calls b, b calls a) -- never
// constructible through the parser or check, which already refuse it on
// their own admission path (D-07-45/SEM-07). Exercised directly against
// callgraph.Order/BuildInterface exactly like callgraph_test.go's own
// synthetic cycle programs (T-07-35).
func mutualCycleProgramForTest() core.Program {
	const aID, bID = "s1:test:fn:a", "s1:test:fn:b"
	a := straightLineFunction(aID, bID)
	b := straightLineFunction(bID, aID)
	return core.Program{
		Schema: core.Schema1, Module: "test.mutual_cycle", ModuleID: "s1:test:module:mutual_cycle",
		Functions: []core.Function{a, b},
	}
}

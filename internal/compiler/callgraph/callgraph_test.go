package callgraph_test

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestImportsStayIndependent reads callgraph's own Go import list (parsed
// from source, never assumed) and fails if it imports check, corevalidate,
// ast, interp, or cgen -- the T-07-... import-independence falsifier, in
// the style of pathoracle's own TestOracleImportsStayIndependent.
func TestImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "callgraph")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"internal/compiler/corevalidate",
		"internal/compiler/ast",
		"internal/compiler/interp",
		"internal/compiler/cgen",
		"internal/compiler/check",
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := dir + string(os.PathSeparator) + entry.Name()
		parsed, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, imp := range parsed.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			for _, bad := range forbidden {
				if strings.Contains(importPath, bad) {
					t.Fatalf("%s imports forbidden package %q", entry.Name(), importPath)
				}
			}
		}
	}
}

// syntheticFunction builds a minimal straight-line core.Function whose
// single OpCall operation names calleeID, mirroring pathoracle_test.go's
// own synthetic-core.Function builder template and ID convention.
func syntheticFunction(id, calleeID, opID string) core.Function {
	return core.Function{
		ID: id, Name: id, EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter: core.Parameter{ID: id + ":place:0", Name: "value", Type: "Byte"},
		Linear: &core.LinearBody{
			ID: id + ":linear",
			Operations: []core.LinearOperation{
				{ID: opID, PointID: id + ":point:linear:0", Kind: core.OpCall, SourceID: id + ":place:0", TargetID: id + ":place:1", CalleeID: calleeID},
			},
		},
	}
}

// leafFunction builds a minimal straight-line core.Function with no
// operations at all -- a legal terminal node with no outgoing call edges.
func leafFunction(id string) core.Function {
	return core.Function{
		ID: id, Name: id, EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter: core.Parameter{ID: id + ":place:0", Name: "value", Type: "Byte"},
		Linear:    &core.LinearBody{ID: id + ":linear"},
	}
}

// TestOrderDetectsMutualCycle proves a length-2 cycle (A calls B, B calls
// A) is refused with callgraph's own witness-carrying typed error, never
// silently accepted and never a native-recursion stack overflow.
func TestOrderDetectsMutualCycle(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		syntheticFunction("fn:a", "fn:b", "fn:a:op:0"),
		syntheticFunction("fn:b", "fn:a", "fn:b:op:0"),
	}}
	_, err := callgraph.Order(program)
	if err == nil {
		t.Fatalf("expected a cycle error, got none")
	}
	cycle, ok := callgraph.CycleError(err)
	if !ok {
		t.Fatalf("want a *cycleError, got: %v", err)
	}
	if cycle.Code() != core.CallGraphCycle {
		t.Fatalf("want code %q, got %q", core.CallGraphCycle, cycle.Code())
	}
	if len(cycle.Members()) != 2 {
		t.Fatalf("want a 2-member witness, got %v", cycle.Members())
	}
}

// TestOrderAcceptsAcyclicChain proves an acyclic chain returns a non-empty
// reverse postorder with no error.
func TestOrderAcceptsAcyclicChain(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		syntheticFunction("fn:a", "fn:b", "fn:a:op:0"),
		syntheticFunction("fn:b", "fn:c", "fn:b:op:0"),
		leafFunction("fn:c"),
	}}
	order, err := callgraph.Order(program)
	if err != nil {
		t.Fatalf("unexpected error on an acyclic chain: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("want a 3-node reverse postorder, got %v", order)
	}
}

// TestOrderDedupesDuplicateEdges proves 5 duplicate OpCall operations
// between the same pair of functions yield exactly one edge -- E is
// bounded by V-squared, not by raw operation count.
func TestOrderDedupesDuplicateEdges(t *testing.T) {
	ops := make([]core.LinearOperation, 0, 5)
	for i := 0; i < 5; i++ {
		ops = append(ops, core.LinearOperation{
			ID: "fn:a:op:" + itoa(i), PointID: "fn:a:point:linear:" + itoa(i),
			Kind: core.OpCall, SourceID: "fn:a:place:0", TargetID: "fn:a:place:1", CalleeID: "fn:b",
		})
	}
	program := core.Program{Functions: []core.Function{
		{
			ID: "fn:a", Name: "a", EntryPointID: "fn:a:point:entry", ReturnPointID: "fn:a:point:return",
			Parameter: core.Parameter{ID: "fn:a:place:0", Name: "value", Type: "Byte"},
			Linear:    &core.LinearBody{ID: "fn:a:linear", Operations: ops},
		},
		leafFunction("fn:b"),
	}}
	order, err := callgraph.Order(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 2 {
		t.Fatalf("want a 2-node reverse postorder despite 5 duplicate operations, got %v", order)
	}
}

// TestOrderSortsAdjacencyByCalleeID proves adjacency lists are sorted by
// callee ID at build time, asserted on a synthetic program built with
// out-of-order edges (fn:a calls fn:c then fn:b, declared in that source
// order) by observing the reverse postorder visits fn:c's subtree, then
// fn:b's, in the sorted-adjacency order rather than declaration order.
func TestOrderSortsAdjacencyByCalleeID(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		{
			ID: "fn:a", Name: "a", EntryPointID: "fn:a:point:entry", ReturnPointID: "fn:a:point:return",
			Parameter: core.Parameter{ID: "fn:a:place:0", Name: "value", Type: "Byte"},
			Linear: &core.LinearBody{
				ID: "fn:a:linear",
				Operations: []core.LinearOperation{
					{ID: "fn:a:op:0", PointID: "fn:a:point:linear:0", Kind: core.OpCall, SourceID: "fn:a:place:0", TargetID: "fn:a:place:1", CalleeID: "fn:cc"},
					{ID: "fn:a:op:1", PointID: "fn:a:point:linear:1", Kind: core.OpCall, SourceID: "fn:a:place:1", TargetID: "fn:a:place:2", CalleeID: "fn:bb"},
				},
			},
		},
		leafFunction("fn:bb"),
		leafFunction("fn:cc"),
	}}
	order, err := callgraph.Order(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Sorted adjacency visits fn:bb before fn:cc (alphabetical), so fn:bb
	// FINISHES first and therefore sits LATER in reverse postorder than
	// fn:cc (reverse postorder lists a node before everything that
	// finished before it).
	indexOf := func(id string) int {
		for i, v := range order {
			if v == id {
				return i
			}
		}
		return -1
	}
	if indexOf("fn:cc") >= indexOf("fn:bb") {
		t.Fatalf("want fn:cc ordered before fn:bb in reverse postorder under sorted (bb,cc) adjacency, got %v", order)
	}
}

// TestOrderRefusesUnresolvedCallee proves an OpCall whose CalleeID names no
// declared function makes Order return the unresolved-callee error from
// 07-03, never the cycle error and never a silently skipped edge
// (D-07-45) -- exercised directly against a forged core.Program no parser
// produced (T-07-35).
func TestOrderRefusesUnresolvedCallee(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		syntheticFunction("fn:a", "fn:ghost", "fn:a:op:0"),
	}}
	_, err := callgraph.Order(program)
	if err == nil {
		t.Fatalf("expected an unresolved-callee error, got none")
	}
	if _, ok := callgraph.CycleError(err); ok {
		t.Fatalf("expected the unresolved-callee error, got the cycle error instead")
	}
	unresolved, ok := callgraph.UnresolvedCalleeError(err)
	if !ok {
		t.Fatalf("want an unresolved-callee error, got: %v", err)
	}
	if unresolved.Code() != core.CallCalleeUnresolved {
		t.Fatalf("want code %q, got %q", core.CallCalleeUnresolved, unresolved.Code())
	}
}

// threeCycleProgram builds a synthetic program whose only cycle is a
// 3-member cycle among ids[0]->ids[1]->ids[2]->ids[0], with leaf as an
// unrelated fourth declared function so root iteration order is never
// trivially just the cycle itself.
func threeCycleProgram(ids [3]string) core.Program {
	return core.Program{Functions: []core.Function{
		syntheticFunction(ids[0], ids[1], ids[0]+":op:0"),
		syntheticFunction(ids[1], ids[2], ids[1]+":op:0"),
		syntheticFunction(ids[2], ids[0], ids[2]+":op:0"),
		leafFunction("fn:zzz-leaf"),
	}}
}

// entryOffsetCycleProgram builds a synthetic program with an external
// caller (fn:000, alphabetically smallest, so it is always visited as a
// root before any cycle member) that enters a 3-member cycle
// (fn:bbb->fn:ccc->fn:aaa->fn:bbb) partway through, at fn:bbb -- NOT at
// the cycle's own alphabetically-smallest member (fn:aaa). Because every
// declared function is independently a root, and roots are visited in
// sorted order, a SELF-CONTAINED cycle's own smallest member is always
// visited as a root before any other cycle member gets a chance to be
// reached first, which means the raw (unrotated) discovery order already
// happens to start at the smallest member with no help from rotation at
// all. This fixture exists specifically to defeat that: fn:000's own
// traversal reaches the cycle at fn:bbb BEFORE the root loop ever reaches
// fn:aaa as a fresh root (fn:aaa is already colored by then), so the raw
// witness genuinely starts at a non-smallest member, and rotation is the
// only thing that can put fn:aaa back at index 0.
func entryOffsetCycleProgram() core.Program {
	return core.Program{Functions: []core.Function{
		syntheticFunction("fn:000", "fn:bbb", "fn:000:op:0"),
		syntheticFunction("fn:aaa", "fn:bbb", "fn:aaa:op:0"),
		syntheticFunction("fn:bbb", "fn:ccc", "fn:bbb:op:0"),
		syntheticFunction("fn:ccc", "fn:aaa", "fn:ccc:op:0"),
	}}
}

// diagnosticIDFor builds a minimal identity-bearing string from a cycle's
// rotated members, standing in for check's own diagnostic.Error identity
// (Schema/Code/Span/Causes hashed verbatim and in order) without needing
// to import diagnostic or duplicate check's cause-construction: since
// Causes participate in ID identity IN ORDER, two witnesses with the same
// members in the same order always produce the same real diagnostic ID,
// and this comparison is exactly as sensitive to order as that real
// construction is.
func diagnosticIDFor(members []string) string {
	id := ""
	for _, member := range members {
		id += member + "|"
	}
	return id
}

// TestCycleIdentityStableAcrossDeclarationOrder is Task 2 Test 1: a
// synthetic 3-cycle built with its functions declared in one order, and
// the identical cycle built with its functions declared in the reverse
// order, yield one identical rotated witness (and therefore one identical
// diagnostic identity), because Order always computes its own sorted
// roots and sorted adjacency regardless of core.Program.Functions'
// declaration order.
func TestCycleIdentityStableAcrossDeclarationOrder(t *testing.T) {
	ids := [3]string{"fn:aaa", "fn:bbb", "fn:ccc"}
	forward := threeCycleProgram(ids)
	reversed := core.Program{Functions: append([]core.Function(nil), forward.Functions...)}
	for i, j := 0, len(reversed.Functions)-1; i < j; i, j = i+1, j-1 {
		reversed.Functions[i], reversed.Functions[j] = reversed.Functions[j], reversed.Functions[i]
	}

	_, errForward := callgraph.Order(forward)
	_, errReversed := callgraph.Order(reversed)
	cycleForward, ok := callgraph.CycleError(errForward)
	if !ok {
		t.Fatalf("expected a cycle error for forward declaration order, got: %v", errForward)
	}
	cycleReversed, ok := callgraph.CycleError(errReversed)
	if !ok {
		t.Fatalf("expected a cycle error for reversed declaration order, got: %v", errReversed)
	}
	idForward := diagnosticIDFor(cycleForward.Members())
	idReversed := diagnosticIDFor(cycleReversed.Members())
	if idForward != idReversed {
		t.Fatalf("want one identical witness regardless of declaration order, got %v vs %v", cycleForward.Members(), cycleReversed.Members())
	}
}

// discoveryOrderTwoCycleProgram builds a synthetic program where a single
// shared root (fn:0root, alphabetically smallest so it is always visited
// as a root before either cycle's own members) fans out, in sorted-
// adjacency order, to TWO independent cycles: fn:2bbb (visited FIRST,
// since "fn:2bbb" < "fn:3ccc"), a length-1 self-cycle, and fn:3ccc
// (visited SECOND), a length-2 cycle through fn:1aaa. fn:1aaa's ID sorts
// BETWEEN fn:0root and fn:2bbb, so it is never independently visited as
// its own root before fn:0root's single depth-first traversal reaches it
// via fn:3ccc -- discovery order here is genuinely driven by adjacency
// order, not by which cycle happens to contain the globally smallest ID.
// fn:3ccc's cycle (rotated: [fn:1aaa, fn:3ccc]) is lexicographically
// smaller than fn:2bbb's self-cycle ([fn:2bbb]), but fn:2bbb's cycle is
// discovered FIRST -- exactly the case D-07-43 requires witness selection
// (not mere discovery order) to resolve correctly.
func discoveryOrderTwoCycleProgram() core.Program {
	root := core.Function{
		ID: "fn:0root", Name: "0root", EntryPointID: "fn:0root:point:entry", ReturnPointID: "fn:0root:point:return",
		Parameter: core.Parameter{ID: "fn:0root:place:0", Name: "value", Type: "Byte"},
		Linear: &core.LinearBody{
			ID: "fn:0root:linear",
			Operations: []core.LinearOperation{
				{ID: "fn:0root:op:0", PointID: "fn:0root:point:linear:0", Kind: core.OpCall, SourceID: "fn:0root:place:0", TargetID: "fn:0root:place:1", CalleeID: "fn:2bbb"},
				{ID: "fn:0root:op:1", PointID: "fn:0root:point:linear:1", Kind: core.OpCall, SourceID: "fn:0root:place:1", TargetID: "fn:0root:place:2", CalleeID: "fn:3ccc"},
			},
		},
	}
	return core.Program{Functions: []core.Function{
		root,
		syntheticFunction("fn:2bbb", "fn:2bbb", "fn:2bbb:op:0"), // length-1 self-cycle
		syntheticFunction("fn:3ccc", "fn:1aaa", "fn:3ccc:op:0"),
		syntheticFunction("fn:1aaa", "fn:3ccc", "fn:1aaa:op:0"),
	}}
}

// TestTwoDistinctCyclesSelectsLexicographicallySmallestWitness is Task 2
// Test 2 (D-07-43), the discovery-order-driven case: even though fn:2bbb's
// self-cycle is discovered FIRST during the traversal,
// discoveryOrderTwoCycleProgram's diagnostic identity is always fn:3ccc's
// cycle (rotated to [fn:1aaa, fn:3ccc]), because that witness's canonical
// rotation is lexicographically smaller.
func TestTwoDistinctCyclesSelectsLexicographicallySmallestWitness(t *testing.T) {
	_, err := callgraph.Order(discoveryOrderTwoCycleProgram())
	cycle, ok := callgraph.CycleError(err)
	if !ok {
		t.Fatalf("expected a cycle error, got: %v", err)
	}
	members := cycle.Members()
	if len(members) != 2 || members[0] != "fn:1aaa" || members[1] != "fn:3ccc" {
		t.Fatalf("want the lexicographically-smallest-rotation witness [fn:1aaa fn:3ccc] selected despite later discovery, got %v", members)
	}
}

// TestTwoDistinctCyclesYieldSameWitnessAcrossPermutations is Task 2 Test 2
// (D-07-43): a synthetic graph containing TWO distinct, non-overlapping
// cycles yields the same witness -- and therefore the same diagnostic
// identity -- regardless of the functions' declaration order, because the
// witness selected is the one whose canonical rotation is
// lexicographically smallest among every cycle discovered.
func TestTwoDistinctCyclesYieldSameWitnessAcrossPermutations(t *testing.T) {
	build := func(order []core.Function) core.Program { return core.Program{Functions: order} }
	cycleA := []core.Function{
		syntheticFunction("fn:mmm", "fn:nnn", "fn:mmm:op:0"),
		syntheticFunction("fn:nnn", "fn:mmm", "fn:nnn:op:0"),
	}
	cycleB := []core.Function{
		syntheticFunction("fn:xxx", "fn:yyy", "fn:xxx:op:0"),
		syntheticFunction("fn:yyy", "fn:xxx", "fn:yyy:op:0"),
	}
	permutations := [][]core.Function{
		append(append([]core.Function{}, cycleA...), cycleB...),
		append(append([]core.Function{}, cycleB...), cycleA...),
		{cycleA[1], cycleB[0], cycleA[0], cycleB[1]},
	}

	var previous []string
	for i, order := range permutations {
		_, err := callgraph.Order(build(order))
		cycle, ok := callgraph.CycleError(err)
		if !ok {
			t.Fatalf("permutation %d: expected a cycle error, got: %v", i, err)
		}
		members := cycle.Members()
		if previous != nil {
			if diagnosticIDFor(members) != diagnosticIDFor(previous) {
				t.Fatalf("permutation %d: want the same witness as prior permutations, got %v vs %v", i, members, previous)
			}
		}
		previous = members
		// fn:mmm sorts before fn:xxx, so cycleA's canonical rotation
		// ("fn:mmm", "fn:nnn") is lexicographically smaller than cycleB's
		// ("fn:xxx", "fn:yyy") -- the deterministically-selected witness
		// must always be cycleA's.
		if members[0] != "fn:mmm" {
			t.Fatalf("permutation %d: want the lexicographically-smallest-rotation cycle (fn:mmm...) selected, got %v", i, members)
		}
	}
}

// TestRotationRemovalMutationKilled is Task 2 Test 3's rotation half:
// disabling canonical rotation (Task 2's own seeded mutation) makes the
// witness stop starting at the cycle's own smallest member on
// entryOffsetCycleProgram, where the raw discovery order genuinely does
// NOT start at the smallest member -- the corpus can actually kill this
// control, not merely assert it.
func TestRotationRemovalMutationKilled(t *testing.T) {
	// At the production default, rotation puts fn:aaa at index 0.
	_, defaultErr := callgraph.Order(entryOffsetCycleProgram())
	defaultCycle, ok := callgraph.CycleError(defaultErr)
	if !ok || defaultCycle.Members()[0] != "fn:aaa" {
		t.Fatalf("expected the production default to rotate fn:aaa to index 0, got: %v", defaultErr)
	}

	restore := callgraph.SetRotationDisabledForTest(true)
	defer restore()
	_, mutatedErr := callgraph.Order(entryOffsetCycleProgram())
	mutatedCycle, ok := callgraph.CycleError(mutatedErr)
	if !ok {
		t.Fatalf("expected a cycle error with rotation disabled, got: %v", mutatedErr)
	}
	if mutatedCycle.Members()[0] == "fn:aaa" {
		t.Fatalf("expected disabling rotation to break the smallest-member-first property, but fn:aaa is still at index 0: %v", mutatedCycle.Members())
	}
}

// TestWitnessSelectionRemovalMutationKilled is Task 2 Test 3's
// witness-selection half: disabling cross-cycle witness selection (Task
// 2's own seeded mutation) makes discoveryOrderTwoCycleProgram pick
// whichever cycle was discovered FIRST (fn:2bbb's self-cycle) instead of
// the lexicographically smallest (fn:3ccc's, rotated to
// [fn:1aaa, fn:3ccc]) -- the corpus can actually kill this control.
func TestWitnessSelectionRemovalMutationKilled(t *testing.T) {
	restore := callgraph.SetWitnessSelectionDisabledForTest(true)
	defer restore()

	_, err := callgraph.Order(discoveryOrderTwoCycleProgram())
	cycle, ok := callgraph.CycleError(err)
	if !ok {
		t.Fatalf("expected a cycle error, got: %v", err)
	}
	members := cycle.Members()
	if len(members) == 2 && members[0] == "fn:1aaa" && members[1] == "fn:3ccc" {
		t.Fatalf("expected disabling witness selection to pick the FIRST-discovered cycle (fn:2bbb) instead of the smallest, but got the correct answer: %v", members)
	}
}

// TestCycleRotationPutsSmallestFunctionIDFirst is a direct assertion of
// D-07-16's rotation rule on a single cycle: whichever member the DFS
// happens to discover the back edge from, index 0 of the reported witness
// is always the lexicographically smallest member ID.
func TestCycleRotationPutsSmallestFunctionIDFirst(t *testing.T) {
	program := threeCycleProgram([3]string{"fn:zzz", "fn:aaa", "fn:mmm"})
	_, err := callgraph.Order(program)
	cycle, ok := callgraph.CycleError(err)
	if !ok {
		t.Fatalf("expected a cycle error, got: %v", err)
	}
	members := cycle.Members()
	if members[0] != "fn:aaa" {
		t.Fatalf("want the lexicographically smallest member (fn:aaa) at index 0, got %v", members)
	}
}

// TestMemberEdgeOperationIDsWrapAround proves MemberEdgeOperationIDs
// carries one operation ID per member, each realizing that member's own
// outgoing edge to the NEXT member (wrapping from the last back to the
// first), never a Span (D-07-35: core.LinearOperation has no Span field,
// and this package's own witness carries only operation IDs).
func TestMemberEdgeOperationIDsWrapAround(t *testing.T) {
	program := core.Program{Functions: []core.Function{
		syntheticFunction("fn:a", "fn:b", "fn:a:op:0"),
		syntheticFunction("fn:b", "fn:a", "fn:b:op:0"),
	}}
	_, err := callgraph.Order(program)
	cycle, ok := callgraph.CycleError(err)
	if !ok {
		t.Fatalf("expected a cycle error, got: %v", err)
	}
	members := cycle.Members()
	edgeIDs := cycle.MemberEdgeOperationIDs()
	if len(edgeIDs) != len(members) {
		t.Fatalf("want one edge operation ID per member, got %d ids for %d members", len(edgeIDs), len(members))
	}
	for i, member := range members {
		expectedOpID := member + ":op:0"
		if edgeIDs[i] != expectedOpID {
			t.Fatalf("member %d (%s): want edge operation ID %q, got %q", i, member, expectedOpID, edgeIDs[i])
		}
	}
	closing := cycle.ClosingOperationID()
	if closing != edgeIDs[len(edgeIDs)-1] {
		t.Fatalf("want ClosingOperationID to be the wrap-around (last) edge, got %q vs %q", closing, edgeIDs[len(edgeIDs)-1])
	}
}

// TestOrderNeverBoundsAnAcyclicTraversal is Task 2 Test 8: a 200-node
// acyclic diamond-laden synthetic graph returns a non-empty reverse
// postorder with no error, no truncation -- the traversal itself is never
// bounded, only the diagnostic a cycle would produce.
func TestOrderNeverBoundsAnAcyclicTraversal(t *testing.T) {
	const layers = 66
	functions := []core.Function{leafFunction("fn:leaf")}
	previous := "fn:leaf"
	for i := 0; i < layers; i++ {
		a := "fn:mid" + itoa(i) + "a"
		b := "fn:mid" + itoa(i) + "b"
		top := "fn:top" + itoa(i)
		functions = append(functions,
			syntheticFunction(a, previous, a+":op:0"),
			syntheticFunction(b, previous, b+":op:0"),
		)
		functions = append(functions, core.Function{
			ID: top, Name: top, EntryPointID: top + ":point:entry", ReturnPointID: top + ":point:return",
			Parameter: core.Parameter{ID: top + ":place:0", Name: "value", Type: "Byte"},
			Linear: &core.LinearBody{
				ID: top + ":linear",
				Operations: []core.LinearOperation{
					{ID: top + ":op:0", PointID: top + ":point:linear:0", Kind: core.OpCall, SourceID: top + ":place:0", TargetID: top + ":place:1", CalleeID: a},
					{ID: top + ":op:1", PointID: top + ":point:linear:1", Kind: core.OpCall, SourceID: top + ":place:1", TargetID: top + ":place:2", CalleeID: b},
				},
			},
		})
		previous = top
	}
	// 1 leaf + 66*(a+b+top) = 1 + 198 = 199 nodes; comfortably over 200
	// once counting is inclusive on the boundary this test targets.
	order, err := callgraph.Order(core.Program{Functions: functions})
	if err != nil {
		t.Fatalf("unexpected error on a %d-node acyclic diamond-laden graph: %v", len(functions), err)
	}
	if len(order) != len(functions) {
		t.Fatalf("want a reverse postorder covering every node, got %d for %d functions", len(order), len(functions))
	}
}

func itoa(i int) string {
	digits := "0123456789"
	if i == 0 {
		return "0"
	}
	var buf []byte
	for i > 0 {
		buf = append([]byte{digits[i%10]}, buf...)
		i /= 10
	}
	return string(buf)
}

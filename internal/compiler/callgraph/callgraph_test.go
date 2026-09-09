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

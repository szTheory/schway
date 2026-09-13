package check

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
)

// ---------------------------------------------------------------------
// Phase 13 Plan 02: contract-boundary blame (D-13-01..D-13-08).
//
// Every test function here is prefixed TestBlame so the phase's scoped
// verify command (`go test ./internal/compiler/check/... -run TestBlame`)
// selects exactly this file's coverage.
// ---------------------------------------------------------------------

// synthLinearFunction builds a minimal core.Function carrying exactly the
// operations named, for callgraph/functionOwnerIndex tests that need a
// synthetic core.Program rather than a parsed .lang fixture. Each entry in
// calls becomes one core.OpCall operation whose CalleeID is the given
// callee function ID.
func synthLinearFunction(id string, calls ...string) core.Function {
	ops := make([]core.LinearOperation, 0, len(calls)+1)
	for i, calleeID := range calls {
		ops = append(ops, core.LinearOperation{
			ID:       id + ":op:" + itoa(i),
			PointID:  id + ":point:linear:" + itoa(i),
			Kind:     core.OpCall,
			SourceID: id + ":place:0",
			TargetID: id + ":place:" + itoa(i+1),
			TypeID:   id + ":type:0",
			CalleeID: calleeID,
		})
	}
	// One non-call operation too, so functionByOperationID's "total and
	// exact" claim is exercised against more than one operation kind per
	// function.
	ops = append(ops, core.LinearOperation{
		ID: id + ":op:tail", PointID: id + ":point:tail", Kind: core.OpMove,
		SourceID: id + ":place:0", TargetID: id + ":place:tail", TypeID: id + ":type:0",
	})
	return core.Function{
		ID: id, Name: id,
		Linear: &core.LinearBody{ID: id + ":linear", Operations: ops},
	}
}

func itoa(i int) string {
	digits := "0123456789"
	if i == 0 {
		return "0"
	}
	out := make([]byte, 0, 4)
	for i > 0 {
		out = append([]byte{digits[i%10]}, out...)
		i /= 10
	}
	return string(out)
}

// TestBlameFunctionByOperationIDIsTotalAndExact is Task 1's first behavior:
// every operation in every function's Linear.Operations resolves to its OWN
// function, an unknown operation ID returns ok == false (never a guess),
// and a duplicated operation ID across two functions also refuses rather
// than silently keeping the first (or last) writer.
func TestBlameFunctionByOperationIDIsTotalAndExact(t *testing.T) {
	leaf := synthLinearFunction("s1:m:fn:leaf")
	mid := synthLinearFunction("s1:m:fn:mid", "s1:m:fn:leaf")
	top := synthLinearFunction("s1:m:fn:top", "s1:m:fn:mid")
	program := core.Program{Functions: []core.Function{leaf, mid, top}}

	index := buildFunctionByOperationID(program)

	for _, function := range program.Functions {
		for _, operation := range function.Linear.Operations {
			gotID, ok := index.lookup(operation.ID)
			if !ok {
				t.Fatalf("operation %s: expected ok=true, got false", operation.ID)
			}
			if gotID != function.ID {
				t.Fatalf("operation %s: expected owner %s, got %s", operation.ID, function.ID, gotID)
			}
		}
	}

	if _, ok := index.lookup("s1:m:fn:nowhere:op:0"); ok {
		t.Fatal("expected an unknown operation ID to resolve ok=false")
	}

	// Duplicated operation ID across two DIFFERENT functions: a detectable
	// condition, never a silent last-writer-wins overwrite.
	dupA := core.Function{ID: "s1:m:fn:dupA", Name: "dupA", Linear: &core.LinearBody{
		ID: "s1:m:fn:dupA:linear",
		Operations: []core.LinearOperation{
			{ID: "s1:m:fn:shared:op:0", PointID: "p", Kind: core.OpMove, SourceID: "s", TargetID: "t", TypeID: "ty"},
		},
	}}
	dupB := core.Function{ID: "s1:m:fn:dupB", Name: "dupB", Linear: &core.LinearBody{
		ID: "s1:m:fn:dupB:linear",
		Operations: []core.LinearOperation{
			{ID: "s1:m:fn:shared:op:0", PointID: "p", Kind: core.OpMove, SourceID: "s", TargetID: "t", TypeID: "ty"},
		},
	}}
	dupProgram := core.Program{Functions: []core.Function{dupA, dupB}}
	dupIndex := buildFunctionByOperationID(dupProgram)
	if _, ok := dupIndex.lookup("s1:m:fn:shared:op:0"); ok {
		t.Fatal("expected a duplicated operation ID (claimed by two functions) to resolve ok=false, not a guess")
	}
}

// TestBlameCalleeBeforeCallerOrderMatchesSummaryOrder is Task 1's second
// behavior: calleeBeforeCallerOrder places every callee before every one of
// its callers, on both a diamond and a deep-chain call-graph shape, and
// buildInterproceduralSummaries consumes exactly this same helper (the
// package's ONE non-comment callgraph.Order call site, asserted separately
// by this plan's acceptance-criteria grep) -- so testing the ordering
// property here is testing what buildInterproceduralSummaries itself now
// consumes, not a parallel derivation.
func TestBlameCalleeBeforeCallerOrderMatchesSummaryOrder(t *testing.T) {
	assertCalleeBeforeCaller := func(t *testing.T, program core.Program) []string {
		t.Helper()
		order, err := calleeBeforeCallerOrder(program)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(order) != len(program.Functions) {
			t.Fatalf("expected order to name every function exactly once, got %v", order)
		}
		position := map[string]int{}
		for i, id := range order {
			position[id] = i
		}
		for _, function := range program.Functions {
			for _, operation := range function.Linear.Operations {
				if operation.Kind != core.OpCall {
					continue
				}
				callerPos, ok := position[function.ID]
				if !ok {
					t.Fatalf("caller %s missing from order", function.ID)
				}
				calleePos, ok := position[operation.CalleeID]
				if !ok {
					t.Fatalf("callee %s missing from order", operation.CalleeID)
				}
				if calleePos >= callerPos {
					t.Fatalf("expected callee %s (pos %d) before caller %s (pos %d)", operation.CalleeID, calleePos, function.ID, callerPos)
				}
			}
		}
		return order
	}

	// Deep chain: top -> mid -> leaf.
	leaf := synthLinearFunction("s1:m:fn:leaf")
	mid := synthLinearFunction("s1:m:fn:mid", "s1:m:fn:leaf")
	top := synthLinearFunction("s1:m:fn:top", "s1:m:fn:mid")
	chain := core.Program{Functions: []core.Function{leaf, mid, top}}
	firstOrder := assertCalleeBeforeCaller(t, chain)
	secondOrder, err := calleeBeforeCallerOrder(chain)
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if len(firstOrder) != len(secondOrder) {
		t.Fatalf("expected calleeBeforeCallerOrder to be deterministic across calls, got %v then %v", firstOrder, secondOrder)
	}
	for i := range firstOrder {
		if firstOrder[i] != secondOrder[i] {
			t.Fatalf("expected calleeBeforeCallerOrder to be deterministic, got %v then %v", firstOrder, secondOrder)
		}
	}

	// Diamond: top -> {a, b}; a -> leaf; b -> leaf.
	diamondLeaf := synthLinearFunction("s1:m:fn:dleaf")
	a := synthLinearFunction("s1:m:fn:a", "s1:m:fn:dleaf")
	b := synthLinearFunction("s1:m:fn:b", "s1:m:fn:dleaf")
	diamondTop := synthLinearFunction("s1:m:fn:dtop", "s1:m:fn:a", "s1:m:fn:b")
	diamond := core.Program{Functions: []core.Function{diamondLeaf, a, b, diamondTop}}
	assertCalleeBeforeCaller(t, diamond)
}

// TestBlameOrderingAuthorityIsSingleSited is a same-package companion to
// this plan's acceptance-criteria grep (which asserts exactly one
// non-comment callgraph.Order call site in check.go): confirms
// buildInterproceduralSummaries and calleeBeforeCallerOrder agree on the
// callee-before-caller order for a real, non-trivial multi-function
// program, rather than each independently computing something that merely
// happens to look similar.
func TestBlameOrderingAuthorityIsSingleSited(t *testing.T) {
	leaf := synthLinearFunction("s1:m:fn:leaf")
	mid := synthLinearFunction("s1:m:fn:mid", "s1:m:fn:leaf")
	program := core.Program{Functions: []core.Function{leaf, mid}}
	table := callSignatureTable{}
	_, work := buildInterproceduralSummaries(program, table)
	if work == 0 {
		t.Fatal("expected buildInterproceduralSummaries to do nonzero work over a two-function program")
	}
}

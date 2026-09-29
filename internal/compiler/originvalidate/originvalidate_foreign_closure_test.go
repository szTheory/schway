package originvalidate_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// fallibleForeignReachProgram is 07-12's own standing witness for
// CR-03/PVG-02 (testdata/phase07/call_fallible_foreign_reach.schway): main
// calls tracer, which try's a foreign C symbol declared allocator
// "libc_malloc", unwind/nonlocal_exit forbidden, fails ProbeError.
func fallibleForeignReachProgram(t testing.TB) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase07", "call_fallible_foreign_reach.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("call_fallible_foreign_reach.schway: unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestForeignReachIsClosureDerived is Task 1 Test 1's Foreign half
// (CR-03/PVG-02): main -- which never itself declares a ForeignContract,
// only calls tracer, which try's a libc-reaching foreign symbol -- now
// publishes the SAME non-empty ForeignReach as tracer, closure-derived
// over the acyclic call graph. Before 07-12, main published the zero
// ForeignReach, which core.go's own doc comment names as a LEGAL value
// meaning "no foreign reach" -- the exact false claim CR-03 found.
func TestForeignReachIsClosureDerived(t *testing.T) {
	program := fallibleForeignReachProgram(t)
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	byID := signaturesByID(summary)
	var tracer, main core.FunctionSignature
	for _, function := range program.Functions {
		switch function.Name {
		case "tracer":
			tracer = byID[function.ID]
		case "main":
			main = byID[function.ID]
		}
	}
	wantForeign := core.ForeignReach{Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "forbidden"}
	if tracer.Foreign != wantForeign {
		t.Fatalf("expected tracer (the direct foreign caller) to publish %+v, got %+v", wantForeign, tracer.Foreign)
	}
	if main.Foreign != wantForeign {
		t.Fatalf("expected main (the TRANSITIVE foreign caller) to publish the joined %+v, got %+v -- this is CR-03/PVG-02: a body-blind consumer of main's summary alone must see the real reach", wantForeign, main.Foreign)
	}
}

// TestFailsIsClosureDerived is Task 1 Test 1's Fails half (CR-03/PVG-02):
// main publishes the same non-empty Fails ("ProbeError") as tracer, never
// the "" empty value core.go documents as meaning infallible.
func TestFailsIsClosureDerived(t *testing.T) {
	program := fallibleForeignReachProgram(t)
	summary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	byID := signaturesByID(summary)
	var tracer, main core.FunctionSignature
	for _, function := range program.Functions {
		switch function.Name {
		case "tracer":
			tracer = byID[function.ID]
		case "main":
			main = byID[function.ID]
		}
	}
	if tracer.Fails != "ProbeError" {
		t.Fatalf("expected tracer to publish fails=ProbeError, got %q", tracer.Fails)
	}
	if main.Fails != "ProbeError" {
		t.Fatalf("expected main to publish the joined fails=ProbeError, got %q -- before 07-12 this was \"\", the exact zero value core.go documents as meaning infallible", main.Fails)
	}
}

// TestForeignReachClosureLeafUnperturbed is Task 1 Test 2: a function with
// no callees at all publishes exactly the Foreign/Fails it would publish
// standalone -- the join's identity element is the empty reach, so a
// leaf's published value is byte-identical whether it is built alone or
// inside a larger program.
func TestForeignReachClosureLeafUnperturbed(t *testing.T) {
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
		t.Fatal("expected deep_diamond_acyclic.schway to declare a function named leaf")
	}
	fullSummary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(full program): %v", err)
	}
	inContext := signaturesByID(fullSummary)[leaf.ID]

	standalone := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, Functions: []core.Function{leaf}}
	standaloneSummary, err := originvalidate.BuildInterface(standalone)
	if err != nil {
		t.Fatalf("BuildInterface(standalone leaf): %v", err)
	}
	if inContext.Foreign != standaloneSummary.Functions[0].Foreign {
		t.Fatalf("expected leaf's Foreign to be unperturbed by context: in-context=%+v standalone=%+v", inContext.Foreign, standaloneSummary.Functions[0].Foreign)
	}
	if inContext.Fails != standaloneSummary.Functions[0].Fails {
		t.Fatalf("expected leaf's Fails to be unperturbed by context: in-context=%q standalone=%q", inContext.Fails, standaloneSummary.Functions[0].Fails)
	}
}

// foreignContractFunction builds a leaf function (no callees) declaring
// contract as its ForeignContract -- hand-built directly, like this
// package's own straightLineFunction/mutualCycleProgramForTest precedent,
// so the join's field-by-field behavior can be driven with fully
// controlled, possibly conflicting, callee reach values.
func foreignContractFunction(id string, contract core.ForeignContract) core.Function {
	fn := straightLineFunction(id, "")
	fn.ForeignContract = &contract
	return fn
}

// multiCallFunction builds a function that calls every ID in calleeIDs, in
// the given order, then returns -- letting a test drive the SAME callee
// SET through DIFFERENT source-level call orders, to prove
// calleeIDsForClosureDigest's dedup-and-sort (and therefore the join that
// folds over it) is independent of the order calls happen to appear in a
// function's body.
func multiCallFunction(id string, calleeIDs []string) core.Function {
	var operations []core.LinearOperation
	last := id + ":place:0"
	for i, calleeID := range calleeIDs {
		target := fmt.Sprintf("%s:place:%d", id, i+1)
		operations = append(operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", id, i), PointID: fmt.Sprintf("%s:point:linear:%d", id, i),
			Kind: core.OpCall, SourceID: last, TargetID: target, CalleeID: calleeID,
		})
		last = target
	}
	operations = append(operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", id, len(calleeIDs)), PointID: fmt.Sprintf("%s:point:linear:%d", id, len(calleeIDs)),
		Kind: core.OpReturn, SourceID: last,
	})
	return core.Function{
		ID: id, Name: id[len("s1:test:fn:"):],
		EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter:  core.Parameter{ID: id + ":place:0", Name: "value", Type: "Byte"},
		ReturnType: "Byte",
		Linear:     &core.LinearBody{ID: id + ":linear", Operations: operations},
	}
}

// TestForeignJoinIsOrderIndependent is Task 1 Test 3/Test 4 (edge 1/3/4,
// D-07-38/SEM-04): joinForeignReach is commutative and associative, so a
// caller's published Foreign does not depend on the order its callees
// happen to be called in. Two programs are built with the SAME two-callee
// set but OPPOSITE call orders in the caller's own body; the two callees'
// reaches disagree on every field (a genuine allocator conflict, and a
// forbidden-vs-permitted disagreement on Unwind) -- driving both edge 1
// (adjacency: equal-merge / more-constraining-wins) and edge 4 (boundary:
// forbidden beats permitted, and two different non-empty allocator names
// resolve to core.ForeignReachConflict, never either input).
func TestForeignJoinIsOrderIndependent(t *testing.T) {
	const callerID, calleeAID, calleeBID = "s1:test:fn:caller", "s1:test:fn:calleea", "s1:test:fn:calleeb"
	calleeA := foreignContractFunction(calleeAID, core.ForeignContract{
		Symbol: "a", Allocator: "libc_malloc", Unwind: "forbidden", NonlocalExit: "", Fails: "ErrA",
	})
	calleeB := foreignContractFunction(calleeBID, core.ForeignContract{
		Symbol: "b", Allocator: "custom_alloc", Unwind: "permitted", NonlocalExit: "forbidden", Fails: "ErrB",
	})

	forward := core.Program{
		Schema: core.Schema1, Module: "test.join_order", ModuleID: "s1:test:module:join_order_forward",
		Functions: []core.Function{multiCallFunction(callerID, []string{calleeAID, calleeBID}), calleeA, calleeB},
	}
	backward := core.Program{
		Schema: core.Schema1, Module: "test.join_order", ModuleID: "s1:test:module:join_order_backward",
		Functions: []core.Function{multiCallFunction(callerID, []string{calleeBID, calleeAID}), calleeA, calleeB},
	}

	forwardSummary, err := originvalidate.BuildInterface(forward)
	if err != nil {
		t.Fatalf("BuildInterface(forward order): %v", err)
	}
	backwardSummary, err := originvalidate.BuildInterface(backward)
	if err != nil {
		t.Fatalf("BuildInterface(backward order): %v", err)
	}
	forwardCaller := signaturesByID(forwardSummary)[callerID]
	backwardCaller := signaturesByID(backwardSummary)[callerID]

	if forwardCaller.Foreign != backwardCaller.Foreign {
		t.Fatalf("expected caller's joined Foreign to be independent of call order: forward=%+v backward=%+v", forwardCaller.Foreign, backwardCaller.Foreign)
	}

	wantForeign := core.ForeignReach{
		Allocator:    core.ForeignReachConflict, // "libc_malloc" vs "custom_alloc": no ordering, resolves to the sentinel.
		Unwind:       "forbidden",               // "forbidden" vs "permitted": forbidden wins.
		NonlocalExit: "forbidden",               // "" vs "forbidden": non-empty wins (identity element).
	}
	if forwardCaller.Foreign != wantForeign {
		t.Fatalf("expected caller's joined Foreign to be %+v, got %+v", wantForeign, forwardCaller.Foreign)
	}
}

// TestSecondPassResolvesProgramFunctionByID is Task 1 Test 7 (07-REVIEW.md
// IN-01): the second pass resolves each callsite's own function BY ID
// (programFunctionByID), never by an index shared with summary.Functions.
// A three-function program is driven through TWO different declaration
// orders for the same function set (unrelated declared before vs after
// caller/callee); each function's own published Foreign/Fails must depend
// ONLY on its own identity and its own callees -- never on its position
// within program.Functions -- proving no function's join can bleed into a
// neighboring function's slot.
func TestSecondPassResolvesProgramFunctionByID(t *testing.T) {
	const callerID, calleeID, unrelatedID = "s1:test:fn:idcaller", "s1:test:fn:idcallee", "s1:test:fn:idunrelated"
	callee := foreignContractFunction(calleeID, core.ForeignContract{Symbol: "c", Allocator: "libc_malloc", Unwind: "forbidden", Fails: "ErrC"})
	caller := multiCallFunction(callerID, []string{calleeID})
	unrelated := foreignContractFunction(unrelatedID, core.ForeignContract{Symbol: "u", Allocator: "other_alloc", Unwind: "permitted", Fails: "ErrU"})

	declaredFirst := core.Program{
		Schema: core.Schema1, Module: "test.id_resolve", ModuleID: "s1:test:module:id_resolve_first",
		Functions: []core.Function{unrelated, caller, callee},
	}
	declaredLast := core.Program{
		Schema: core.Schema1, Module: "test.id_resolve", ModuleID: "s1:test:module:id_resolve_last",
		Functions: []core.Function{caller, callee, unrelated},
	}

	firstSummary, err := originvalidate.BuildInterface(declaredFirst)
	if err != nil {
		t.Fatalf("BuildInterface(unrelated first): %v", err)
	}
	lastSummary, err := originvalidate.BuildInterface(declaredLast)
	if err != nil {
		t.Fatalf("BuildInterface(unrelated last): %v", err)
	}

	firstByID, lastByID := signaturesByID(firstSummary), signaturesByID(lastSummary)

	callerFirst, callerLast := firstByID[callerID], lastByID[callerID]
	if callerFirst.Foreign != callerLast.Foreign || callerFirst.Fails != callerLast.Fails {
		t.Fatalf("caller's published Foreign/Fails depended on declaration order: first=%+v/%q last=%+v/%q", callerFirst.Foreign, callerFirst.Fails, callerLast.Foreign, callerLast.Fails)
	}
	wantCallerForeign := core.ForeignReach{Allocator: "libc_malloc", Unwind: "forbidden"}
	if callerFirst.Foreign != wantCallerForeign || callerFirst.Fails != "ErrC" {
		t.Fatalf("caller should have joined ONLY its own callee (callee)'s reach, got Foreign=%+v Fails=%q", callerFirst.Foreign, callerFirst.Fails)
	}

	unrelatedFirst, unrelatedLast := firstByID[unrelatedID], lastByID[unrelatedID]
	if unrelatedFirst.Foreign != unrelatedLast.Foreign || unrelatedFirst.Fails != unrelatedLast.Fails {
		t.Fatalf("unrelated's published Foreign/Fails depended on declaration order: first=%+v/%q last=%+v/%q", unrelatedFirst.Foreign, unrelatedFirst.Fails, unrelatedLast.Foreign, unrelatedLast.Fails)
	}
	wantUnrelatedForeign := core.ForeignReach{Allocator: "other_alloc", Unwind: "permitted"}
	if unrelatedFirst.Foreign != wantUnrelatedForeign || unrelatedFirst.Fails != "ErrU" {
		t.Fatalf("unrelated (never called by caller) must publish only its OWN local contract, got Foreign=%+v Fails=%q -- an index mis-association would leak caller's/callee's joined facts here", unrelatedFirst.Foreign, unrelatedFirst.Fails)
	}
}

// TestForeignClosureJoinMutationKilled is Task 3's producer-side kill for
// control:summary.foreign_reach_closure_derived: with
// foreignClosureJoinSeam engaged, main WRONGLY publishes the zero
// ForeignReach (CR-03's original defect); restoring the seam restores the
// correct joined value.
func TestForeignClosureJoinMutationKilled(t *testing.T) {
	program := fallibleForeignReachProgram(t)
	var mainID string
	for _, function := range program.Functions {
		if function.Name == "main" {
			mainID = function.ID
		}
	}

	restore := originvalidate.SetForeignClosureJoinSeam(true)
	mutatedSummary, err := originvalidate.BuildInterface(program)
	restore()
	if err != nil {
		t.Fatalf("BuildInterface(seam engaged): %v", err)
	}
	mutatedMain := signaturesByID(mutatedSummary)[mainID]
	if mutatedMain.Foreign != (core.ForeignReach{}) {
		t.Fatalf("mutation (disabling the Foreign join) had no observable effect: main still published %+v", mutatedMain.Foreign)
	}

	restoredSummary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(restored): %v", err)
	}
	restoredMain := signaturesByID(restoredSummary)[mainID]
	if restoredMain.Foreign == (core.ForeignReach{}) {
		t.Fatal("expected the Foreign join to be restored once the seam is disengaged")
	}
}

// TestFailsClosureJoinMutationKilled is Task 3's producer-side kill for
// control:summary.fails_closure_derived: with failsClosureJoinSeam
// engaged, main WRONGLY publishes fails="" (CR-03's original defect);
// restoring the seam restores the correct joined value.
func TestFailsClosureJoinMutationKilled(t *testing.T) {
	program := fallibleForeignReachProgram(t)
	var mainID string
	for _, function := range program.Functions {
		if function.Name == "main" {
			mainID = function.ID
		}
	}

	restore := originvalidate.SetFailsClosureJoinSeam(true)
	mutatedSummary, err := originvalidate.BuildInterface(program)
	restore()
	if err != nil {
		t.Fatalf("BuildInterface(seam engaged): %v", err)
	}
	mutatedMain := signaturesByID(mutatedSummary)[mainID]
	if mutatedMain.Fails != "" {
		t.Fatalf("mutation (disabling the Fails join) had no observable effect: main still published fails=%q", mutatedMain.Fails)
	}

	restoredSummary, err := originvalidate.BuildInterface(program)
	if err != nil {
		t.Fatalf("BuildInterface(restored): %v", err)
	}
	restoredMain := signaturesByID(restoredSummary)[mainID]
	if restoredMain.Fails == "" {
		t.Fatal("expected the Fails join to be restored once the seam is disengaged")
	}
}

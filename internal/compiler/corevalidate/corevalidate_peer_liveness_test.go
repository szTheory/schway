package corevalidate

import (
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// mustCheckFixture parses and checks a testdata fixture through the real
// front end, exactly like check_test.go's own mustParseProgram +
// Program(...) pair, so the two mutation-kill tests below drive the same
// real .schway fixtures the retired peerDivergenceExpected entries named,
// not a hand-built synthetic program.
func mustCheckFixture(t *testing.T, pathParts ...string) check.Result {
	t.Helper()
	relativePath := strings.Join(pathParts, "/")
	source, err := os.ReadFile(testsupport.ProjectPath(pathParts...))
	if err != nil {
		t.Fatalf("read fixture %q: %v", relativePath, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed to parse: %+v", relativePath, parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed check.Program: %+v", relativePath, result.Diagnostics)
	}
	return result
}

// loanCarryFunction builds one structurally-minimal core.Function for
// derivePeerLoanCarry's own tests: a parameter place, plus whatever
// operations the caller supplies, closed with a final core.OpReturn. It
// deliberately reuses NONE of syntheticFunction's machinery (that helper
// always ends a function with a call chain then a return of the last call's
// result) since these tests need to hand-place OpBorrow*/OpMove/OpCopy/OpCall
// operations in specific shapes derivePeerLoanCarry itself must
// discriminate between.
func loanCarryFunction(id string, operations []core.LinearOperation) *core.Function {
	parameterID := id + ":place:0"
	return &core.Function{
		ID: id, Name: id,
		Parameter: core.Parameter{ID: parameterID, Name: "value", Type: "Byte"},
		ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID:         id + ":linear",
			Operations: operations,
		},
	}
}

// TestPeerLoanCarryDerivesForwardFromOperations is Task 1 Test 1 (D-09-02):
// a function whose body borrows its own parameter and returns that borrow is
// ReturnsBorrowOfParam == true; a function whose body copies (moves) its
// parameter into a fresh place and returns THAT is false -- an ordinary
// owned move must never be reported as borrow-derived, matching
// peerReturnDerivesFromBorrow's own identical distinction.
func TestPeerLoanCarryDerivesForwardFromOperations(t *testing.T) {
	borrowing := loanCarryFunction("s1:test:fn:borrowing", []core.LinearOperation{
		{ID: "op0", Kind: core.OpBorrowShared, SourceID: "s1:test:fn:borrowing:place:0", TargetID: "place:1", LoanID: "loan:0"},
		{ID: "op1", Kind: core.OpReturn, SourceID: "place:1"},
	})
	if fact := derivePeerLoanCarry(borrowing, nil); !fact.ReturnsBorrowOfParam {
		t.Fatal("expected a function returning a borrow of its own parameter to report ReturnsBorrowOfParam == true")
	}

	owned := loanCarryFunction("s1:test:fn:owned", []core.LinearOperation{
		{ID: "op0", Kind: core.OpMove, SourceID: "s1:test:fn:owned:place:0", TargetID: "place:1"},
		{ID: "op1", Kind: core.OpReturn, SourceID: "place:1"},
	})
	if fact := derivePeerLoanCarry(owned, nil); fact.ReturnsBorrowOfParam {
		t.Fatal("expected a function that only moves (takes) its own parameter and returns the fresh place to report ReturnsBorrowOfParam == false")
	}
}

// TestPeerLoanCarryIsMemoizedCalleeBeforeCaller is Task 1 Test 2 (D-09-01):
// proves chainPeerLoanCarry's callee-before-caller postorder ordering is
// genuinely load-bearing, not incidentally correct, by running the SAME
// two-function program (caller calls callee, callee borrows and returns its
// own parameter) through chainPeerLoanCarry twice -- once with the correct
// callee-before-caller order, once with the order reversed -- and observing
// the caller's own derived fact flip. If the caller's fact read the
// callee's OWN, ALREADY-FINAL entry (the claim this test exists to check),
// reversing the order (so the caller is processed before anything is
// stored for the callee at all) can only ever make the caller's fact WORSE
// (fail closed to false), never better -- exactly what production
// (checkCallGraphAcyclic's own proven-correct postorder) must never do.
func TestPeerLoanCarryIsMemoizedCalleeBeforeCaller(t *testing.T) {
	calleeID := "s1:test:fn:callee"
	callerID := "s1:test:fn:caller"
	callee := core.Function{
		ID: calleeID, Name: "callee",
		Parameter:  core.Parameter{ID: calleeID + ":place:0", Name: "value", Type: "Byte"},
		ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: calleeID + ":linear",
			Operations: []core.LinearOperation{
				{ID: "op0", Kind: core.OpBorrowShared, SourceID: calleeID + ":place:0", TargetID: calleeID + ":place:1", LoanID: "loan:0"},
				{ID: "op1", Kind: core.OpReturn, SourceID: calleeID + ":place:1"},
			},
		},
	}
	caller := core.Function{
		ID: callerID, Name: "caller",
		Parameter:  core.Parameter{ID: callerID + ":place:0", Name: "value", Type: "Byte"},
		ReturnType: "Byte",
		Linear: &core.LinearBody{
			ID: callerID + ":linear",
			Operations: []core.LinearOperation{
				{ID: "op0", Kind: core.OpCall, SourceID: callerID + ":place:0", TargetID: callerID + ":place:1", CalleeID: calleeID},
				{ID: "op1", Kind: core.OpReturn, SourceID: callerID + ":place:1"},
			},
		},
	}
	program := core.Program{Schema: core.Schema1, Module: "peerloancarry", ModuleID: "s1:peerloancarry:module", Functions: []core.Function{callee, caller}}

	correct := &validator{program: program, peerPostorder: []string{calleeID, callerID}}
	correct.chainPeerLoanCarry()
	if !correct.peerLoanCarry[calleeID].ReturnsBorrowOfParam {
		t.Fatal("expected callee's own entry to be ReturnsBorrowOfParam == true")
	}
	if !correct.peerLoanCarry[callerID].ReturnsBorrowOfParam {
		t.Fatal("expected caller's fact, read from the callee's ALREADY-FINAL entry, to be ReturnsBorrowOfParam == true under the correct callee-before-caller order")
	}

	reversed := &validator{program: program, peerPostorder: []string{callerID, calleeID}}
	reversed.chainPeerLoanCarry()
	if reversed.peerLoanCarry[callerID].ReturnsBorrowOfParam {
		t.Fatal("expected caller's fact to fail closed to false when processed BEFORE the callee's own entry exists -- proves the caller genuinely reads v.peerLoanCarry[calleeID], not a re-walk of the callee's own body")
	}
}

// TestPeerLoanCarryTerminatesOnCorruptedCoreArtifact is Task 1 Test 3
// (T-09-02): a core.Program with a self-referencing OpCall CalleeID (a
// function calling itself) must never hang or panic corevalidate's own
// pipeline -- checkCallGraphAcyclic refuses it as a cycle before
// chainPeerLoanCarry ever runs, and Validate must return promptly with
// Valid == false, never block. syntheticProgram/syntheticFunction (this
// package's own cycle-peer test helpers) build the self-edge shape.
func TestPeerLoanCarryTerminatesOnCorruptedCoreArtifact(t *testing.T) {
	program := syntheticProgram(map[string][]string{"a": {"a"}})
	done := make(chan Result, 1)
	go func() { done <- Validate(program) }()
	select {
	case result := <-done:
		if result.Valid {
			t.Fatal("expected a self-referencing OpCall program to be refused, got Valid == true")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Validate hung on a self-referencing core.Program (T-09-02)")
	}
}

// buildLoanCarryChain builds a straight sequential call chain of length
// len(returnsBorrow), function names f0..f(n-1), f0 declared as the
// deepest LEAF (no callees), each fI (i>0) calling f(i-1) with its own
// parameter and returning that call's result unchanged (no local borrow of
// its own). Each leaf-or-relay function's OWN behavior (does it directly
// borrow-and-return its parameter, vs move-and-return) is controlled by
// returnsBorrow[0] only -- every relay above it (i>0) purely forwards.
func buildLoanCarryChain(t *testing.T, leafReturnsBorrow bool) (functions []core.Function, order []string) {
	t.Helper()
	names := []string{"leaf", "mid", "top"}
	var leafOps []core.LinearOperation
	leafID := "s1:test:fn:" + names[0]
	if leafReturnsBorrow {
		leafOps = []core.LinearOperation{
			{ID: "op0", Kind: core.OpBorrowShared, SourceID: leafID + ":place:0", TargetID: leafID + ":place:1", LoanID: "loan:0"},
			{ID: "op1", Kind: core.OpReturn, SourceID: leafID + ":place:1"},
		}
	} else {
		leafOps = []core.LinearOperation{
			{ID: "op0", Kind: core.OpMove, SourceID: leafID + ":place:0", TargetID: leafID + ":place:1"},
			{ID: "op1", Kind: core.OpReturn, SourceID: leafID + ":place:1"},
		}
	}
	functions = append(functions, core.Function{
		ID: leafID, Name: names[0],
		Parameter:  core.Parameter{ID: leafID + ":place:0", Name: "value", Type: "Byte"},
		ReturnType: "Byte",
		Linear:     &core.LinearBody{ID: leafID + ":linear", Operations: leafOps},
	})
	order = append(order, leafID)

	previousID := leafID
	for _, name := range names[1:] {
		fnID := "s1:test:fn:" + name
		functions = append(functions, core.Function{
			ID: fnID, Name: name,
			Parameter:  core.Parameter{ID: fnID + ":place:0", Name: "value", Type: "Byte"},
			ReturnType: "Byte",
			Linear: &core.LinearBody{
				ID: fnID + ":linear",
				Operations: []core.LinearOperation{
					{ID: "op0", Kind: core.OpCall, SourceID: fnID + ":place:0", TargetID: fnID + ":place:1", CalleeID: previousID},
					{ID: "op1", Kind: core.OpReturn, SourceID: fnID + ":place:1"},
				},
			},
		})
		order = append(order, fnID)
		previousID = fnID
	}
	return functions, order
}

// TestPeerLoanCarryPropagatesAcrossTwoCallHops is Task 2 Test 1 (D-09-49
// Q2): a three-function chain top -> mid -> leaf, where leaf returns a
// borrow of its own parameter and mid purely forwards leaf's own call
// result with no local borrow of its own, yields ReturnsBorrowOfParam ==
// true for BOTH mid and top -- composed through TWO call hops from a
// single forward pass per function, chained only via the postorder
// ordering (never a re-walk of leaf's own body from top). The same chain
// with leaf returning an owned (moved) place yields false for both.
func TestPeerLoanCarryPropagatesAcrossTwoCallHops(t *testing.T) {
	borrowing, order := buildLoanCarryChain(t, true)
	v := &validator{program: core.Program{Functions: borrowing}, peerPostorder: order}
	v.chainPeerLoanCarry()
	if !v.peerLoanCarry[order[1]].ReturnsBorrowOfParam {
		t.Fatalf("expected mid (%s), one hop from the borrowing leaf, to be ReturnsBorrowOfParam == true", order[1])
	}
	if !v.peerLoanCarry[order[2]].ReturnsBorrowOfParam {
		t.Fatalf("expected top (%s), TWO hops from the borrowing leaf, to be ReturnsBorrowOfParam == true -- this is the depth-2 composition claim itself", order[2])
	}

	owned, order2 := buildLoanCarryChain(t, false)
	v2 := &validator{program: core.Program{Functions: owned}, peerPostorder: order2}
	v2.chainPeerLoanCarry()
	if v2.peerLoanCarry[order2[1]].ReturnsBorrowOfParam {
		t.Fatalf("expected mid (%s) to be false when leaf returns an owned place", order2[1])
	}
	if v2.peerLoanCarry[order2[2]].ReturnsBorrowOfParam {
		t.Fatalf("expected top (%s) to be false when leaf returns an owned place", order2[2])
	}
}

// TestPeerLoanCarryOrderingIsLoadBearingAtDepthTwo is Task 2 Test 2's own
// falsifier (D-09-49 Q2): the SAME three-hop borrowing chain
// TestPeerLoanCarryPropagatesAcrossTwoCallHops proves correct under
// chainPeerLoanCarry's real callee-before-caller postorder is run again
// under a DECLARATION-order (caller-before-callee) postorder instead --
// the exact wrong order a non-postorder-aware ("non-transitive") consumer
// would use. mid, one hop from leaf, is a direct victim (its own callee's
// entry does not exist yet when mid is processed); top, two hops away, is
// ALSO wrong, but for a compounded reason (it reads mid's own now-wrong
// entry) -- proving a depth-1-only correct derivation is not sufficient
// evidence of depth-2 composition, exactly the gap a shallow, non-transitive
// implementation could hide behind if only depth-1 were ever tested.
func TestPeerLoanCarryOrderingIsLoadBearingAtDepthTwo(t *testing.T) {
	borrowing, order := buildLoanCarryChain(t, true)
	declarationOrder := []string{order[2], order[1], order[0]} // top, mid, leaf: caller-before-callee, wrong
	v := &validator{program: core.Program{Functions: borrowing}, peerPostorder: declarationOrder}
	v.chainPeerLoanCarry()
	if v.peerLoanCarry[order[1]].ReturnsBorrowOfParam {
		t.Fatal("expected mid's fact to fail closed to false under a caller-before-callee (non-postorder) order -- its own callee's entry does not exist yet when mid is processed")
	}
	if v.peerLoanCarry[order[2]].ReturnsBorrowOfParam {
		t.Fatal("expected top's fact to also fail closed to false under a caller-before-callee order -- it reads mid's own (already wrong) entry")
	}
}

// TestPeerLivenessFileImportsStayIndependent is Task 2 Test 3 (D-09-05):
// corevalidate_peer_liveness.go's own import list, parsed from source
// directly, contains none of the eight forbidden producer/engine packages
// this peer must never depend on.
func TestPeerLivenessFileImportsStayIndependent(t *testing.T) {
	path := testsupport.ProjectPath("internal", "compiler", "corevalidate", "corevalidate_peer_liveness.go")
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{"/compiler/check", "/compiler/ast", "/compiler/originvalidate", "/compiler/callgraph", "/compiler/interp", "/compiler/cgen", "/compiler/session", "/compiler/ability"}
	for _, imported := range file.Imports {
		path := strings.Trim(imported.Path.Value, `"`)
		for _, forbid := range forbidden {
			if strings.HasSuffix(path, forbid) {
				t.Fatalf("corevalidate_peer_liveness.go imports %s, which corevalidate must never depend on", path)
			}
		}
	}
}

// TestPeerLoanCarryForcedTrueReintroducesRetiredDivergence is Task 3(d)'s
// QLT-08 mutation-kill: a control that has never been seen to fail is a
// claim, not evidence. With forcePeerLoanCarryTrueForTest engaged, every
// OpCall is treated as carrying its argument's loan regardless of what
// derivePeerLoanCarry actually derived -- reverting corevalidate to its
// pre-Phase-09 unconditional-propagation shape -- and BOTH retired
// peerDivergenceExpected fixtures (testdata/phase08/twin_a_accept.schway,
// testdata/phase08/relay_depth2_accept.schway) must refuse again with
// core.move_while_borrowed, proving D-09-03's retirement is genuinely this
// mechanism's doing.
func TestPeerLoanCarryForcedTrueReintroducesRetiredDivergence(t *testing.T) {
	for _, fixture := range []string{"testdata/phase08/twin_a_accept.schway", "testdata/phase08/relay_depth2_accept.schway"} {
		result := mustCheckFixture(t, strings.Split(fixture, "/")...)
		restore := SetForcePeerLoanCarryTrueForTest(true)
		coreResult := Validate(result.Program)
		restore()
		if coreResult.Valid {
			t.Fatalf("%s: expected corevalidate to refuse again with forcePeerLoanCarryTrueForTest engaged, got Valid == true", fixture)
		}
		found := false
		for _, problem := range coreResult.Problems {
			if problem.Code == "core.move_while_borrowed" {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: expected core.move_while_borrowed, got %+v", fixture, coreResult.Problems)
		}
	}
}

// TestPeerLoanCarryConsultDisabledReintroducesRetiredDivergence is Task
// 3(d)'s second QLT-08 mutation-kill, the same claim proven through the
// OTHER seam: with disablePeerLoanCarryConsultForTest engaged instead of
// forcePeerLoanCarryTrueForTest, buildLoanChainIndex's OpCall branch never
// even reaches the "does the callee carry" question -- a DIFFERENT code
// path than the force-true seam, but the same externally observable
// reintroduced-divergence result, so the two seams each independently kill
// the mutation they were built to catch.
func TestPeerLoanCarryConsultDisabledReintroducesRetiredDivergence(t *testing.T) {
	for _, fixture := range []string{"testdata/phase08/twin_a_accept.schway", "testdata/phase08/relay_depth2_accept.schway"} {
		result := mustCheckFixture(t, strings.Split(fixture, "/")...)
		restore := SetDisablePeerLoanCarryConsultForTest(true)
		coreResult := Validate(result.Program)
		restore()
		if coreResult.Valid {
			t.Fatalf("%s: expected corevalidate to refuse again with disablePeerLoanCarryConsultForTest engaged, got Valid == true", fixture)
		}
		found := false
		for _, problem := range coreResult.Problems {
			if problem.Code == "core.move_while_borrowed" {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s: expected core.move_while_borrowed, got %+v", fixture, coreResult.Problems)
		}
	}
}
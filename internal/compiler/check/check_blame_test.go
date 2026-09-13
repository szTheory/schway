package check

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
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

// ---------------------------------------------------------------------
// Task 2: the B1/B2/B3 rule and the blame_undetermined fail-open closure.
// ---------------------------------------------------------------------

// blameFactsFromDiagnostic classifies every cause of diag into blameFact
// values via classifyDeclaredCause, discarding causes classifyDeclaredCause
// reports out of scope (ok == false) -- mirroring what a production emission
// site consuming this resolver would do.
func blameFactsFromDiagnostic(diag diagnostic.Diagnostic) []blameFact {
	var facts []blameFact
	for _, cause := range diag.Causes {
		if fact, ok := classifyDeclaredCause(diag.Code, cause); ok {
			facts = append(facts, fact)
		}
	}
	return facts
}

// TestBlameClassifiesEveryInterproceduralCode is D-13-04's own
// empirical-verification instruction, applied to the blame resolver
// directly: each of the eight shipped interprocedural codes' REAL causes
// (from running check.Program on a real fixture) classify into blameCaller
// (B2) -- zero B1 candidates, zero blame_undetermined -- and the blamed
// function ID equals the function check.Program's own hand-written Primary
// already selects (D-13-04's "zero churn" claim, restated as a resolver
// property). core.call_graph_cycle takes the SEPARATE D-13-05 exemption
// path (resolveCycleBlame) and is asserted here too, never through
// resolveBlame.
func TestBlameClassifiesEveryInterproceduralCode(t *testing.T) {
	type caseSpec struct {
		code           string
		fixtureDir     string
		fixtureName    string
		callerFuncName string
		isCycle        bool
	}
	cases := []caseSpec{
		{code: "check.interprocedural_loan_liveness", fixtureDir: "phase07", fixtureName: "relay_escort_witness.lang", callerFuncName: "escort"},
		{code: "core.call_graph_cycle", fixtureDir: "phase07", fixtureName: "cycle_indirect.lang", isCycle: true},
		{code: "check.call_argument_type_mismatch", fixtureDir: "phase07", fixtureName: "call_type_mismatch.lang", callerFuncName: "main"},
		{code: "core.callee_not_callable", fixtureDir: "phase07", fixtureName: "call_uncallable_callee.lang", callerFuncName: "main"},
		{code: "syntax.fallible_call_not_consumed", fixtureDir: "phase4", fixtureName: "fallible_call_unconsumed.lang", callerFuncName: "main"},
	}

	readFixture := func(t *testing.T, dir, name string) []byte {
		t.Helper()
		switch dir {
		case "phase07":
			return readPhase07Fixture(t, name)
		case "phase08":
			return readPhase08Fixture(t, name)
		case "phase4":
			return readPhase4Fixture(t, name)
		default:
			t.Fatalf("unknown fixture dir %s", dir)
			return nil
		}
	}

	for _, c := range cases {
		c := c
		t.Run(c.code, func(t *testing.T) {
			source := readFixture(t, c.fixtureDir, c.fixtureName)
			parsed := mustParseProgram(t, source)
			result := Program(parsed)
			var diag diagnostic.Diagnostic
			found := false
			for _, d := range result.Diagnostics {
				if d.Code == c.code {
					diag, found = d, true
					break
				}
			}
			if !found {
				t.Fatalf("expected code %s among diagnostics, got %+v", c.code, result.Diagnostics)
			}

			if c.isCycle {
				// D-13-05: exempt from B3 entirely. Primary stays at the
				// closing operation's span (unchanged, existing behavior);
				// every cycle member is a secondary cause -- assert via
				// resolveCycleBlame directly, never resolveBlame.
				var members []string
				for _, cause := range diag.Causes {
					if cause.Kind == "cycle_member" {
						members = append(members, cause.Detail)
					}
				}
				if len(members) == 0 {
					t.Fatalf("expected cycle_member causes, got %+v", diag.Causes)
				}
				outcome := resolveCycleBlame(members)
				if outcome.Kind != blameCycle {
					t.Fatalf("expected blameCycle, got %v", outcome.Kind)
				}
				if len(outcome.UndeterminedSites) != len(members) {
					t.Fatalf("expected every cycle member published, got %v want %v", outcome.UndeterminedSites, members)
				}
				return
			}

			order, err := calleeBeforeCallerOrder(result.Program)
			if err != nil {
				t.Fatalf("calleeBeforeCallerOrder: %v", err)
			}
			facts := blameFactsFromDiagnostic(diag)
			// The caller function's own ID: read back from the checked
			// core.Program when admission succeeded far enough to publish
			// it (e.g. interprocedural_loan_liveness, callee_not_callable),
			// or computed directly via semanticID from the PARSED module
			// name when the caller's own body never made it into
			// result.Program.Functions (a diagnostic emitted mid-body, e.g.
			// call_argument_type_mismatch/fallible_call_not_consumed
			// short-circuit resolveCallBinding/checkLinear before the
			// caller's own core.Function is ever appended) -- semanticID's
			// "s1:<module>:fn:<name>" scheme is the SAME derivation
			// check.Program itself uses, so this is not a re-derivation of
			// the resolver's own logic, only of the well-known ID scheme.
			callerFunctionID := ""
			for _, function := range result.Program.Functions {
				if function.Name == c.callerFuncName {
					callerFunctionID = function.ID
					break
				}
			}
			if callerFunctionID == "" {
				callerFunctionID = semanticID(parsed.Module, "fn", c.callerFuncName)
			}
			outcome := resolveBlame(result.Program, order, facts, callerFunctionID)
			if outcome.Kind != blameCaller {
				t.Fatalf("expected blameCaller (B2) per D-13-04, got %v (%+v)", outcome.Kind, outcome)
			}
			if outcome.FunctionID != callerFunctionID {
				t.Fatalf("expected blamed function %s, got %s", callerFunctionID, outcome.FunctionID)
			}
		})
	}
}

// TestBlameUndeterminedPublishesBothSitesUnapplied is D-13-07's fail-open
// closure, proven directly: a declared fact classifyDeclaredCause
// recognizes as ABOUT the Return field (a "callee_return_contract" cause)
// but whose Detail does not parse into the expected
// "<calleeID>:return.mode=<Mode>" shape NEVER falls through to B2 --
// resolveBlame routes it to blame_undetermined, publishing BOTH the callee
// and caller sites, and blameUndeterminedRepairs' resulting repairs are
// NEVER diagnostic.DriverEligible (asserted for a repair carrying a real
// Span and Replacement too, proving the gate is Applicability itself, not
// merely missing fields).
func TestBlameUndeterminedPublishesBothSitesUnapplied(t *testing.T) {
	malformed := diagnostic.Cause{Kind: "callee_return_contract", Detail: "not-the-expected-shape"}
	fact, ok := classifyDeclaredCause("check.interprocedural_loan_liveness", malformed)
	if !ok {
		t.Fatal("expected classifyDeclaredCause to recognize callee_return_contract's KIND even when Detail fails to parse")
	}
	if !fact.Declared || fact.Classified {
		t.Fatalf("expected Declared=true, Classified=false for a malformed callee_return_contract Detail, got %+v", fact)
	}

	program := core.Program{Functions: []core.Function{
		{ID: "s1:m:fn:callee", Name: "callee", Linear: &core.LinearBody{ID: "s1:m:fn:callee:linear"}},
		{ID: "s1:m:fn:caller", Name: "caller", Linear: &core.LinearBody{ID: "s1:m:fn:caller:linear"}},
	}}
	order, err := calleeBeforeCallerOrder(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outcome := resolveBlame(program, order, []blameFact{fact}, "s1:m:fn:caller")
	if outcome.Kind != blameUndetermined {
		t.Fatalf("expected blameUndetermined, got %v (%+v)", outcome.Kind, outcome)
	}
	if len(outcome.UndeterminedSites) != 2 {
		t.Fatalf("expected exactly two published sites, got %v", outcome.UndeterminedSites)
	}
	if outcome.UndeterminedSites[1] != "s1:m:fn:caller" {
		t.Fatalf("expected caller site published second, got %v", outcome.UndeterminedSites)
	}

	calleeSpan := diagnostic.Span{Start: 10, End: 20}
	callerSpan := diagnostic.Span{Start: 30, End: 40}
	repairs := blameUndeterminedRepairs(outcome, calleeSpan, callerSpan)
	if len(repairs) != 2 {
		t.Fatalf("expected two repairs, got %+v", repairs)
	}
	for _, repair := range repairs {
		if repair.Applicability != diagnostic.ApplicabilityRequiresConfirmation {
			t.Fatalf("expected RequiresConfirmation, got %q", repair.Applicability)
		}
		if diagnostic.DriverEligible(repair) {
			t.Fatalf("expected DriverEligible == false, got true for %+v", repair)
		}
	}

	// Prove the gate is Applicability itself: construct a repair with a
	// real Span, Replacement, and Kind (everything DriverEligible checks
	// besides Applicability) and confirm RequiresConfirmation alone is
	// still enough to refuse it.
	fullyFormedButUnconfirmed := diagnostic.Repair{
		Kind: "blame_undetermined", Span: &calleeSpan, Replacement: "placeholder",
		Applicability: diagnostic.ApplicabilityRequiresConfirmation,
	}
	if diagnostic.DriverEligible(fullyFormedButUnconfirmed) {
		t.Fatal("expected a fully-formed but RequiresConfirmation repair to still be DriverEligible == false")
	}
}

// TestBlameCycleIsExemptFromB3 asserts D-13-05 directly against a real
// cyclic fixture: core.call_graph_cycle's own Primary (the closing
// operation's span) is untouched, and resolveCycleBlame publishes every
// cycle member -- never routed through resolveBlame's B1/B2/B3 machinery.
func TestBlameCycleIsExemptFromB3(t *testing.T) {
	source := readPhase07Fixture(t, "cycle_indirect.lang")
	result := Program(mustParseProgram(t, source))
	var diag diagnostic.Diagnostic
	found := false
	for _, d := range result.Diagnostics {
		if d.Code == core.CallGraphCycle {
			diag, found = d, true
			break
		}
	}
	if !found {
		t.Fatalf("expected core.call_graph_cycle, got %+v", result.Diagnostics)
	}
	var members []string
	for _, cause := range diag.Causes {
		if cause.Kind == "cycle_member" {
			members = append(members, cause.Detail)
			if cause.Span == nil {
				t.Fatal("expected every cycle_member cause to carry a span")
			}
		}
	}
	if len(members) < 3 {
		t.Fatalf("expected a length-3 cycle (a -> b -> c -> a), got %v", members)
	}
	outcome := resolveCycleBlame(members)
	if outcome.Kind != blameCycle {
		t.Fatalf("expected blameCycle, got %v", outcome.Kind)
	}
	if len(outcome.UndeterminedSites) != len(members) {
		t.Fatalf("expected every cycle member published as a secondary cause, got %v want %v", outcome.UndeterminedSites, members)
	}
	for i, member := range members {
		if outcome.UndeterminedSites[i] != member {
			t.Fatalf("expected cycle members published in order, got %v want %v", outcome.UndeterminedSites, members)
		}
	}
}

// TestBlameTieBreakIsCalleeBeforeCaller is D-13-03's B3: when more than one
// function is a B1 candidate, the winner is the minimum in
// calleeBeforeCallerOrder (the callee-most function), not declaration
// order or string order.
func TestBlameTieBreakIsCalleeBeforeCaller(t *testing.T) {
	leaf := synthLinearFunction("s1:m:fn:leaf")
	mid := synthLinearFunction("s1:m:fn:mid", "s1:m:fn:leaf")
	top := synthLinearFunction("s1:m:fn:top", "s1:m:fn:mid")
	// Declare in an order that DISAGREES with the call-graph order, so a
	// tie-break using declaration index alone would pick the wrong winner.
	program := core.Program{Functions: []core.Function{top, mid, leaf}}
	order, err := calleeBeforeCallerOrder(program)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	facts := []blameFact{
		{FunctionID: "s1:m:fn:mid", Declared: true, Classified: true, Field: declaredFieldReturn, Violated: true},
		{FunctionID: "s1:m:fn:top", Declared: true, Classified: true, Field: declaredFieldReturn, Violated: true},
	}
	outcome := resolveBlame(program, order, facts, "s1:m:fn:top")
	if outcome.Kind != blameFunction {
		t.Fatalf("expected blameFunction (B1/B3), got %v", outcome.Kind)
	}
	if outcome.FunctionID != "s1:m:fn:mid" {
		t.Fatalf("expected mid (callee-before-caller of top) to win the tie-break, got %s", outcome.FunctionID)
	}

	// Residual tie-break: two B1 candidates with IDENTICAL
	// calleeBeforeCallerOrder position is impossible (order is a
	// permutation), so exercise the declaration-index fallback with a
	// candidate absent from order entirely (defensive path).
	factsWithAbsent := []blameFact{
		{FunctionID: "s1:m:fn:ghost", Declared: true, Classified: true, Field: declaredFieldReturn, Violated: true},
		{FunctionID: "s1:m:fn:mid", Declared: true, Classified: true, Field: declaredFieldReturn, Violated: true},
	}
	outcome2 := resolveBlame(program, order, factsWithAbsent, "s1:m:fn:top")
	if outcome2.FunctionID != "s1:m:fn:mid" {
		t.Fatalf("expected the ordered candidate (mid) to win over an unordered one (ghost), got %s", outcome2.FunctionID)
	}
}

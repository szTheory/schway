package check

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

func TestOwnershipSequenceExhaustive(t *testing.T) {
	// Thirty-six symbols span three declaration names, three operation kinds,
	// and four source selectors: every visible place (at this depth) plus an
	// out-of-scope boundary. This exhausts shadowing and multi-owner shapes.
	const alphabet = 36
	for length := 0; length <= 3; length++ {
		cases := 1
		for index := 0; index < length; index++ {
			cases *= alphabet
		}
		for encoded := 0; encoded < cases; encoded++ {
			body := generatedOwnershipBody(encoded, length, alphabet)
			got := analyzeStraightLine("test:fn", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &body)
			want := oracleStraightLine("test:fn", "owner", byteTypeFact(), &body)
			assertSupportEqual(t, fmt.Sprintf("length=%d case=%d", length, encoded), got, want)
		}
	}

	// The same exhaustive alphabet under a type that withholds share: every
	// borrow must be refused at check time, and production must agree with the
	// oracle on which one is refused first.
	for length := 0; length <= 3; length++ {
		cases := 1
		for index := 0; index < length; index++ {
			cases *= alphabet
		}
		for encoded := 0; encoded < cases; encoded++ {
			body := generatedOwnershipBody(encoded, length, alphabet)
			got := analyzeStraightLine("test:fn", "owner", diagnostic.Span{Start: 1, End: 6}, nonShareableTypeFact(), &body)
			want := oracleStraightLine("test:fn", "owner", nonShareableTypeFact(), &body)
			assertSupportEqual(t, fmt.Sprintf("no-share length=%d case=%d", length, encoded), got, want)
		}
	}
	denied := ast.LinearBody{
		Bindings: []ast.Binding{binding("view", "borrow", "owner", 10)},
		Result:   "view",
		Span:     diagnostic.Span{Start: 10, End: 20},
	}
	refused := analyzeStraightLine("test:no-share", "owner", diagnostic.Span{Start: 1, End: 6}, nonShareableTypeFact(), &denied)
	if refused.DiagnosticCode != "ownership.borrow_requires_share" || refused.Diagnostic == nil {
		t.Fatalf("borrow of a type without share was admitted: %+v", refused)
	}
	if refused.Diagnostic.Schema != "lang.diagnostic/1" || len(refused.Diagnostic.Repairs) != 1 {
		t.Fatalf("borrow gate did not select the repair-bearing schema: %+v", refused.Diagnostic)
	}
	if got, want := refused.Diagnostic.Primary, (diagnostic.Span{Start: 10, End: 13}); got != want {
		t.Fatalf("borrow gate reported span %+v, want the borrow site %+v", got, want)
	}
	if len(refused.Operations) != 0 {
		t.Fatalf("borrow gate emitted core it cannot authorize: %+v", refused.Operations)
	}
	causeKinds := make([]string, 0, len(refused.Diagnostic.Causes))
	for _, cause := range refused.Diagnostic.Causes {
		causeKinds = append(causeKinds, cause.Kind)
	}
	if !reflect.DeepEqual(causeKinds, []string{"declared_here", "missing_ability", "place", "type"}) {
		t.Fatalf("borrow gate causes diverged from the copy gate shape: %v", causeKinds)
	}

	witness := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "owner", 10),
			binding("moved", "take", "owner", 20),
			binding("observed", "read", "view", 30),
		},
		Result: "moved",
		Span:   diagnostic.Span{Start: 10, End: 38},
	}
	got := analyzeStraightLine("test:witness", "owner", diagnostic.Span{Start: 1, End: 6}, bufferTypeFact(), &witness)
	if got.DiagnosticCode != "ownership.move_while_borrowed" || len(got.Operations) != 1 {
		t.Fatalf("four-step witness changed: %+v", got)
	}
	if len(got.LoanFinalUses) != 1 || got.LoanFinalUses[0].OperationIndex != 2 {
		t.Fatalf("later loan read is not load-bearing: %+v", got.LoanFinalUses)
	}
	withoutLaterRead := witness
	withoutLaterRead.Bindings = withoutLaterRead.Bindings[:2]
	withoutLaterRead.Result = "moved"
	withoutLaterRead.Span.End = 28
	legal := analyzeStraightLine("test:unused", "owner", diagnostic.Span{Start: 1, End: 6}, bufferTypeFact(), &withoutLaterRead)
	if legal.DiagnosticCode != "" {
		t.Fatalf("unused loan did not end before move: %+v", legal)
	}
}

// TestLoanLivenessFixpoint proves the backward worklist dataflow correctly
// propagates a loan's liveness across a block boundary: a loan born in b1
// and referenced only by b2 (its successor) must be reported live entering
// b2, must not be reported live entering b1 (b1 itself births it), and the
// worklist must report nonzero counted work (D-05) — the propagation this
// pass exists to make honest.
func TestLoanLivenessFixpoint(t *testing.T) {
	loanOp := core.LinearOperation{ID: "fn:op:0", Kind: core.OpBorrowShared, SourceID: "fn:place:0", TargetID: "fn:place:1", LoanID: "fn:loan:0"}
	b1 := cfgBlockSpec{id: "fn:block:b1", operations: []core.LinearOperation{loanOp}, successors: []string{"fn:block:b2"}}
	useOp := core.LinearOperation{ID: "fn:op:1", Kind: core.OpReturn, SourceID: "fn:place:1"}
	b2 := cfgBlockSpec{id: "fn:block:b2", operations: []core.LinearOperation{useOp}, successors: nil}

	result, err := loanLivenessFixpoint("fn", []cfgBlockSpec{b1, b2})
	if err != nil {
		t.Fatalf("unexpected acyclicity error: %v", err)
	}
	if result.work == 0 {
		t.Fatalf("fixpoint reported zero counted work")
	}
	if len(result.liveIn["fn:block:b1"]) != 0 {
		t.Fatalf("loan born in b1 must not be live entering b1: %+v", result.liveIn)
	}
	if !result.liveIn["fn:block:b2"]["fn:loan:0"] {
		t.Fatalf("loan used in b2 must be live entering b2: %+v", result.liveIn)
	}

	// b1 has exactly one successor: no divergence exists to place an edge
	// endpoint on, so the loan flows through b1 unremarked and is finally
	// consumed as a POINT endpoint in b2, where it is actually referenced.
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints("fn", []cfgBlockSpec{b1, b2}, edgeID, result)
	if len(endpoints) != 1 || endpoints[0].Kind != "point" || endpoints[0].LoanID != "fn:loan:0" {
		t.Fatalf("loan consumed in b2 must materialize as exactly one point endpoint there: %+v", endpoints)
	}
	if endpoints[0].BlockID != "fn:block:b2" || endpoints[0].AfterOperationID != "fn:op:1" {
		t.Fatalf("point endpoint named the wrong block/operation: %+v", endpoints[0])
	}
}

// TestEdgeSpecificLiveOut is the general, genuinely multi-successor proof of
// OWN-03's success criterion 2: a block with TWO successors, where only ONE
// successor references the loan, must materialize exactly one edge endpoint
// (on the successor that needs it), not two, and not a point endpoint. This
// is the property checkBranch's own topology (every arm has exactly one
// successor) cannot yet exercise from real source — proven here directly
// against the algorithm, per D-10.
func TestEdgeSpecificLiveOut(t *testing.T) {
	loanOp := core.LinearOperation{ID: "fn:op:0", Kind: core.OpBorrowShared, SourceID: "fn:place:0", TargetID: "fn:place:1", LoanID: "fn:loan:0"}
	entry := cfgBlockSpec{id: "fn:block:entry", operations: []core.LinearOperation{loanOp}, successors: []string{"fn:block:used", "fn:block:unused"}}
	usedOp := core.LinearOperation{ID: "fn:op:1", Kind: core.OpReturn, SourceID: "fn:place:1"}
	used := cfgBlockSpec{id: "fn:block:used", operations: []core.LinearOperation{usedOp}, successors: nil}
	unusedOp := core.LinearOperation{ID: "fn:op:2", Kind: core.OpReturn, SourceID: "fn:place:0"}
	unused := cfgBlockSpec{id: "fn:block:unused", operations: []core.LinearOperation{unusedOp}, successors: nil}

	blocks := []cfgBlockSpec{entry, used, unused}
	result, err := loanLivenessFixpoint("fn", blocks)
	if err != nil {
		t.Fatalf("unexpected acyclicity error: %v", err)
	}
	if !result.liveIn["fn:block:used"]["fn:loan:0"] {
		t.Fatalf("loan must be live entering the block that references it: %+v", result.liveIn)
	}
	if len(result.liveIn["fn:block:unused"]) != 0 {
		t.Fatalf("loan must NOT be live entering the block that never references it: %+v", result.liveIn)
	}

	// The loan is needed along "used" (it keeps propagating, and is finally
	// consumed there as its own point endpoint) but NOT along "unused" --
	// exactly one edge endpoint must land on the DIVERGING edge
	// (entry->unused, where the omission is), not on the edge that still
	// carries it forward.
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints("fn", blocks, edgeID, result)
	var edgeEndpoints, pointEndpoints []core.LoanEndpoint
	for _, endpoint := range endpoints {
		if endpoint.LoanID != "fn:loan:0" {
			continue
		}
		if endpoint.Kind == "edge" {
			edgeEndpoints = append(edgeEndpoints, endpoint)
		} else {
			pointEndpoints = append(pointEndpoints, endpoint)
		}
	}
	if len(edgeEndpoints) != 1 || edgeEndpoints[0].EdgeID != "fn:block:entry->fn:block:unused" {
		t.Fatalf("want exactly one edge endpoint on the diverging (unused) edge, got %+v", edgeEndpoints)
	}
	if len(pointEndpoints) != 1 || pointEndpoints[0].BlockID != "fn:block:used" {
		t.Fatalf("want exactly one point endpoint where the loan is actually consumed, got %+v", pointEndpoints)
	}
}

// TestBackEdgeRejected proves the fixpoint fails closed on a cyclic CFG
// rather than iterating forever (T-03-11) — a real back edge, if one is ever
// present, is rejected, not silently accepted.
func TestBackEdgeRejected(t *testing.T) {
	a := cfgBlockSpec{id: "fn:block:a", operations: nil, successors: []string{"fn:block:b"}}
	b := cfgBlockSpec{id: "fn:block:b", operations: nil, successors: []string{"fn:block:a"}}
	if _, err := loanLivenessFixpoint("fn", []cfgBlockSpec{a, b}); err == nil {
		t.Fatalf("expected a back-edge rejection, got none")
	}
}

// TestStraightLineEndpointsUnchanged pins the shipped straight-line
// reborrow-while-moved fixture's answer against the NEW backward dataflow,
// run here as a single synthetic block (a straight-line body IS one block —
// checkLinear itself is not rewired onto this pass, see the deviation note
// in check.go, but the algorithm's own answer for this shape must still
// agree with the pre-existing, shipped straight-line analysis before it is
// trusted for checkBranch's arm blocks).
func TestStraightLineEndpointsUnchanged(t *testing.T) {
	// A reborrow chain that never moves the owner: both loans stay live all
	// the way to the returned "review" name, matching the shipped
	// reborrow-transitivity law (testdata/phase2/reborrow_while_moved.lang's
	// ACCEPT half; that fixture itself is a reject control, so this test
	// uses the accepted variant with the trailing move removed).
	body := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "code", 0),
			binding("review", "borrow", "view", 8),
		},
		Result: "review",
		Span:   diagnostic.Span{End: 16},
	}
	support := analyzeStraightLine("test:reborrow", "code", diagnostic.Span{}, bufferTypeFact(), &body)
	if support.DiagnosticCode != "" {
		t.Fatalf("shipped reborrow fixture unexpectedly rejected: %+v", support)
	}
	if len(support.LoanFinalUses) != 2 {
		t.Fatalf("want 2 tracked loans, got %+v", support.LoanFinalUses)
	}
	block := cfgBlockSpec{id: "test:reborrow:block:straight", operations: support.Operations, successors: nil}
	result, err := loanLivenessFixpoint("test:reborrow", []cfgBlockSpec{block})
	if err != nil {
		t.Fatalf("unexpected acyclicity error: %v", err)
	}
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints("test:reborrow", []cfgBlockSpec{block}, edgeID, result)
	byLoan := make(map[string]core.LoanEndpoint, len(endpoints))
	for _, endpoint := range endpoints {
		byLoan[endpoint.LoanID] = endpoint
	}
	for _, final := range support.LoanFinalUses {
		endpoint, ok := byLoan[final.LoanID]
		if !ok {
			t.Fatalf("no endpoint materialized for loan %q: %+v", final.LoanID, endpoints)
		}
		wantOperationID := fmt.Sprintf("test:reborrow:op:%d", final.OperationIndex)
		if endpoint.Kind != "point" || endpoint.AfterOperationID != wantOperationID {
			t.Fatalf("loan %q endpoint diverged from the shipped straight-line answer: got %+v, want point at %q", final.LoanID, endpoint, wantOperationID)
		}
	}
}

func TestOwnershipOracleTracksLoansPerOwner(t *testing.T) {
	body := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "owner", 0),
			binding("other", "read", "owner", 4),
			binding("moved", "take", "other", 8),
			binding("observed", "read", "view", 12),
		},
		Result: "moved",
		Span:   diagnostic.Span{End: 20},
	}
	got := analyzeStraightLine("test:multi", "owner", diagnostic.Span{}, byteTypeFact(), &body)
	want := oracleStraightLine("test:multi", "owner", byteTypeFact(), &body)
	assertSupportEqual(t, "borrow one owner while moving another", got, want)
	if got.DiagnosticCode != "" {
		t.Fatalf("unrelated owner move was blocked: %+v", got)
	}

	shadowed := ast.LinearBody{
		Bindings: []ast.Binding{binding("value", "read", "owner", 0), binding("value", "read", "owner", 4)},
		Result:   "value",
		Span:     diagnostic.Span{End: 12},
	}
	assertSupportEqual(t, "shadowed binding", analyzeStraightLine("test:shadow", "owner", diagnostic.Span{}, byteTypeFact(), &shadowed), oracleStraightLine("test:shadow", "owner", byteTypeFact(), &shadowed))

	shadowedLoan := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "owner", 0),
			binding("view", "read", "owner", 4),
			binding("moved", "take", "owner", 8),
			binding("observed", "read", "view", 12),
		},
		Result: "moved",
		Span:   diagnostic.Span{End: 20},
	}
	got = analyzeStraightLine("test:shadow-loan", "owner", diagnostic.Span{}, byteTypeFact(), &shadowedLoan)
	want = oracleStraightLine("test:shadow-loan", "owner", byteTypeFact(), &shadowedLoan)
	assertSupportEqual(t, "shadowed loan identity", got, want)
	if got.DiagnosticCode != "" {
		t.Fatalf("obsolete shadowed loan blocked move: %+v", got)
	}
}

func TestOwnershipWorkSeries(t *testing.T) {
	for _, operations := range []int{10, 100, 1_000, 10_000} {
		bindings := make([]ast.Binding, operations)
		for index := range bindings {
			bindings[index] = binding(fmt.Sprintf("copy%d", index), "read", "owner", index*2)
		}
		body := ast.LinearBody{Bindings: bindings, Result: "owner", Span: diagnostic.Span{End: operations*2 + 1}}
		got := analyzeStraightLine("test:scale", "owner", diagnostic.Span{}, byteTypeFact(), &body)
		wantWork := 1 + 2*(operations+1)
		if got.DiagnosticCode != "" || len(got.Operations) != operations+1 || got.Work != wantWork {
			t.Fatalf("operations=%d result=%+v want work=%d", operations, got, wantWork)
		}
		if got.Work > 2*len(got.Operations)+1 {
			t.Fatalf("operations=%d exceeded linear bound: work=%d core_ops=%d", operations, got.Work, len(got.Operations))
		}
	}
}

func FuzzOwnershipLinear(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{2, 1, 3})
	f.Add([]byte{1, 0})
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 32 {
			input = input[:32]
		}
		body := generatedOwnershipBodyBytes(input)
		got := analyzeStraightLine("test:fuzz", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &body)
		want := oracleStraightLine("test:fuzz", "owner", byteTypeFact(), &body)
		assertSupportEqual(t, fmt.Sprintf("input=%v", input), got, want)
	})
}

func generatedOwnershipBody(encoded, length, alphabet int) ast.LinearBody {
	input := make([]byte, length)
	for index := range input {
		input[index] = byte(encoded % alphabet)
		encoded /= alphabet
	}
	return generatedOwnershipBodyBytes(input)
}

func generatedOwnershipBodyBytes(input []byte) ast.LinearBody {
	body := ast.LinearBody{Result: "owner"}
	visibleNames := []string{"owner"}
	knownName := map[string]bool{"owner": true}
	names := []string{"value0", "value1", "value2"}
	for index, raw := range input {
		name := names[int(raw)%len(names)]
		kinds := []string{"read", "take", "borrow"}
		kind := kinds[(int(raw)/len(names))%len(kinds)]
		sources := append(append([]string(nil), visibleNames...), "out_of_scope")
		source := sources[(int(raw)/(len(names)*len(kinds)))%len(sources)]
		body.Bindings = append(body.Bindings, binding(name, kind, source, index*4))
		if !knownName[name] {
			knownName[name] = true
			visibleNames = append(visibleNames, name)
		}
	}
	if len(input) > 0 {
		body.Result = visibleNames[int(input[len(input)-1])%len(visibleNames)]
	}
	body.Span = diagnostic.Span{End: len(input)*4 + 1}
	return body
}

func binding(name, kind, source string, start int) ast.Binding {
	span := diagnostic.Span{Start: start, End: start + 3}
	return ast.Binding{Name: name, RHS: ast.RHS{Kind: kind, Source: source, Span: span}, Span: span}
}

func byteTypeFact() core.TypeFact {
	return core.TypeFact{
		ID: "test:fn:type:0", Shape: core.TypeRef{Constructor: "Byte", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{},
	}
}

// nonShareableTypeFact is deliberately synthetic. Every source-reachable type
// constructor (Byte, Buffer, Box, Pair) grants share, so no fixture can reach
// the borrow gate; this fact exercises the law itself. It grants copy so the
// implicit-copy gate cannot mask a missing share gate.
func nonShareableTypeFact() core.TypeFact {
	return core.TypeFact{
		ID: "test:fn:type:0", Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityCopy, core.AbilityDrop, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{{Ability: core.AbilityShare, Path: []string{"Buffer"}}},
	}
}

func bufferTypeFact() core.TypeFact {
	return core.TypeFact{
		ID: "test:fn:type:0", Shape: core.TypeRef{Constructor: "Buffer", Arguments: []core.TypeRef{}},
		Abilities:         []core.Ability{core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape},
		NegativeWitnesses: []core.AbilityWitness{{Ability: core.AbilityCopy, Path: []string{"Buffer"}}},
	}
}

// oracleStraightLine deliberately duplicates the tiny semantic law. It does
// not call production last-use, transition, identity, normalization, or state
// snapshot helpers.
type testOraclePlace struct {
	id          string
	initialized bool
}

type testOracleLoan struct {
	ownerID string
	lastUse int
}

// oracleLoanLastUses derives loan liveness by a method the production pass
// never uses. Production streams forward over the bindings once, carrying an
// inherited loan association from each source binding to the binding it
// produces, and never materializes the derivation relation. The oracle does
// the reverse: it first materializes the whole derivation relation as an
// explicit edge set, then closes each loan under that relation by
// order-independent fixed-point iteration (repeat until nothing new becomes
// reachable), and only afterwards reduces the reachable set to its maximum
// ordinal. Neither derivation can be obtained from the other by renaming, so
// a transitivity error in one cannot be mirrored by the other.
func oracleLoanLastUses(parameterName string, body *ast.LinearBody) map[int]int {
	edges := make(map[int][]int)
	for index := range body.Bindings {
		from, ok := oracleResolveBinding(parameterName, body, body.Bindings[index].RHS.Source, index)
		if ok && from >= 0 {
			edges[from] = append(edges[from], index)
		}
	}
	resultBinding, resultKnown := oracleResolveBinding(parameterName, body, body.Result, len(body.Bindings))

	lastUses := make(map[int]int)
	for loan, candidate := range body.Bindings {
		if candidate.RHS.Kind != "borrow" {
			continue
		}
		reachable := map[int]bool{loan: true}
		for changed := true; changed; {
			changed = false
			for from, targets := range edges {
				if !reachable[from] {
					continue
				}
				for _, to := range targets {
					if !reachable[to] {
						reachable[to] = true
						changed = true
					}
				}
			}
		}
		last := loan
		for index := range reachable {
			if index > last {
				last = index
			}
		}
		if resultKnown && resultBinding >= 0 && reachable[resultBinding] {
			last = len(body.Bindings)
		}
		lastUses[loan] = last
	}
	return lastUses
}

func oracleStraightLine(functionID, parameterName string, typeFact core.TypeFact, body *ast.LinearBody) ownershipSupport {
	lastUses := oracleLoanLastUses(parameterName, body)
	loanOrder := make([]int, 0)
	for index, candidate := range body.Bindings {
		if candidate.RHS.Kind == "borrow" {
			loanOrder = append(loanOrder, index)
		}
	}

	result := ownershipSupport{Operations: []core.LinearOperation{}, LoanFinalUses: []loanFinalUseFact{}, States: []ownershipStateFact{}, Work: typeNodeCountOracle(typeFact.Shape) + len(body.Bindings) + 1}
	for _, index := range loanOrder {
		result.LoanFinalUses = append(result.LoanFinalUses, loanFinalUseFact{LoanID: fmt.Sprintf("%s:loan:%d", functionID, index), Binding: body.Bindings[index].Name, OperationIndex: lastUses[index]})
	}

	parameterID := functionID + ":place:0"
	states := map[string]*testOraclePlace{parameterID: {id: parameterID, initialized: true}}
	visiblePlaces := map[string]*testOraclePlace{parameterName: states[parameterID]}
	activeLoans := make(map[string]testOracleLoan)
	for index, candidate := range body.Bindings {
		result.Work++
		sourceBinding, ok := oracleResolveBinding(parameterName, body, candidate.RHS.Source, index)
		sourceID := parameterID
		if sourceBinding >= 0 {
			sourceID = fmt.Sprintf("%s:place:%d", functionID, sourceBinding+1)
		}
		source := states[sourceID]
		if !ok {
			result.DiagnosticCode = "name.unknown"
			return result
		}
		if !source.initialized {
			result.DiagnosticCode = "ownership.use_after_move"
			return result
		}
		if candidate.RHS.Kind == "take" && hasFutureLoanOracle(activeLoans, source.id, index) {
			result.DiagnosticCode = "ownership.move_while_borrowed"
			return result
		}
		kind := core.OpCopy
		loanID := ""
		switch candidate.RHS.Kind {
		case "take":
			kind = core.OpMove
			source.initialized = false
		case "borrow":
			if !oracleHasAbility(typeFact.Abilities, core.AbilityShare) {
				result.DiagnosticCode = "ownership.borrow_requires_share"
				return result
			}
			kind = core.OpBorrowShared
			loanID = fmt.Sprintf("%s:loan:%d", functionID, index)
			activeLoans[loanID] = testOracleLoan{ownerID: source.id, lastUse: lastUses[index]}
		default:
			if !oracleHasAbility(typeFact.Abilities, core.AbilityCopy) {
				result.DiagnosticCode = "ownership.transfer_requires_take"
				return result
			}
		}
		targetID := fmt.Sprintf("%s:place:%d", functionID, index+1)
		target := &testOraclePlace{id: targetID, initialized: true}
		states[targetID] = target
		visiblePlaces[candidate.Name] = target
		result.Operations = append(result.Operations, core.LinearOperation{
			ID: fmt.Sprintf("%s:op:%d", functionID, index), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, index),
			Kind: kind, SourceID: source.id, TargetID: targetID, LoanID: loanID, TypeID: typeFact.ID,
		})
		for id, loan := range activeLoans {
			if loan.lastUse <= index {
				delete(activeLoans, id)
			}
		}
		result.States = append(result.States, oracleSnapshot(index, visiblePlaces, activeLoans))
	}
	result.Work++
	resultBinding, ok := oracleResolveBinding(parameterName, body, body.Result, len(body.Bindings))
	resultID := parameterID
	if resultBinding >= 0 {
		resultID = fmt.Sprintf("%s:place:%d", functionID, resultBinding+1)
	}
	returned := states[resultID]
	if !ok {
		result.DiagnosticCode = "name.unknown"
		return result
	}
	if !returned.initialized {
		result.DiagnosticCode = "ownership.use_after_move"
		return result
	}
	ordinal := len(body.Bindings)
	result.Operations = append(result.Operations, core.LinearOperation{
		ID: fmt.Sprintf("%s:op:%d", functionID, ordinal), PointID: fmt.Sprintf("%s:point:linear:%d", functionID, ordinal),
		Kind: core.OpReturn, SourceID: returned.id, TypeID: typeFact.ID,
	})
	for id, loan := range activeLoans {
		if loan.lastUse <= ordinal {
			delete(activeLoans, id)
		}
	}
	result.States = append(result.States, oracleSnapshot(ordinal, visiblePlaces, activeLoans))
	return result
}

// oracleResolveBinding is intentionally a different implementation from the
// production forward environment: each use scans declarations backward and
// selects the nearest prior binding, with -1 denoting the parameter.
func oracleResolveBinding(parameterName string, body *ast.LinearBody, name string, before int) (int, bool) {
	for index := before - 1; index >= 0; index-- {
		if body.Bindings[index].Name == name {
			return index, true
		}
	}
	if name == parameterName {
		return -1, true
	}
	return -1, false
}

func oracleSnapshot(index int, places map[string]*testOraclePlace, activeLoans map[string]testOracleLoan) ownershipStateFact {
	initialized := make([]string, 0)
	for _, place := range places {
		if place.initialized {
			initialized = append(initialized, place.id)
		}
	}
	sort.Strings(initialized)
	loans := make([]string, 0, len(activeLoans))
	for id := range activeLoans {
		loans = append(loans, id)
	}
	sort.Strings(loans)
	return ownershipStateFact{OperationIndex: index, InitializedPlaces: initialized, ActiveLoans: loans}
}

func hasFutureLoanOracle(active map[string]testOracleLoan, ownerID string, index int) bool {
	for _, loan := range active {
		if loan.ownerID == ownerID && loan.lastUse > index {
			return true
		}
	}
	return false
}

func oracleHasAbility(abilities []core.Ability, wanted core.Ability) bool {
	for _, candidate := range abilities {
		if candidate == wanted {
			return true
		}
	}
	return false
}

func typeNodeCountOracle(value core.TypeRef) int {
	count := 1
	for _, argument := range value.Arguments {
		count += typeNodeCountOracle(argument)
	}
	return count
}

func assertSupportEqual(t *testing.T, name string, got, want ownershipSupport) {
	t.Helper()
	if !reflect.DeepEqual(got.Operations, want.Operations) ||
		!reflect.DeepEqual(got.LoanFinalUses, want.LoanFinalUses) ||
		!reflect.DeepEqual(got.States, want.States) ||
		got.Work != want.Work || got.DiagnosticCode != want.DiagnosticCode {
		t.Fatalf("%s support mismatch:\ngot  %+v\nwant %+v", name, got, want)
	}
}

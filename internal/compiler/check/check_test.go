package check

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// readTestdataFixture and mustParseProgram let internal (package check)
// tests drive Program(...) directly against a real testdata/phase3
// fixture, exactly as the seeded-fault test needs (it must call
// check.Program with testOnlyForceUniformLoanJoin toggled, which is
// unexported and therefore unreachable from the external check_test
// package the other fixture-driving tests use).
func readTestdataFixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../testdata/phase3/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return source
}

func mustParseProgram(t *testing.T, source []byte) ast.Program {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture failed to parse: %+v", parsed.Diagnostics)
	}
	return parsed.Program
}

func TestOwnershipSequenceExhaustive(t *testing.T) {
	// Forty-eight symbols span three declaration names, FOUR operation kinds
	// (implicit copy, take, shared borrow, exclusive borrow -- 03-05-03/D-10
	// extends the original three-kind alphabet to reach the conflict matrix
	// exhaustively too), and four source selectors: every visible place (at
	// this depth) plus an out-of-scope boundary. This exhausts shadowing,
	// multi-owner shapes, AND every shared/exclusive overlap combination the
	// five-row conflict matrix distinguishes.
	const alphabet = 48
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

// TestUniformJoinPlacementFlipsBothVerdicts is 03-03-02's seeded-fault
// falsifier: with testOnlyForceUniformLoanJoin engaged (every arm-body
// loan's last use forced to the arm's own join point, exactly as if
// edge-specific placement had been deleted), the accept fixture's own
// borrow-then-return arm reports a diagnostic it did not report before —
// proving the accept verdict is load-bearing on edge-specific placement,
// not merely green by construction. This is the plan's own
// must_haves.truths claim ("deleting the edge-specific placement makes ONE
// of those two fixtures flip verdict, and the suite fails") — the fault is
// a test-only seam (see testOnlyForceUniformLoanJoin's doc comment), so it
// is falsified by direct mutation (flipping the variable), not by
// reverting a production hunk.
func TestUniformJoinPlacementFlipsBothVerdicts(t *testing.T) {
	source := readTestdataFixture(t, "branch_one_arm_shared_accept.lang")

	baseline := Program(mustParseProgram(t, source))
	if len(baseline.Diagnostics) != 0 {
		t.Fatalf("accept fixture unexpectedly rejected before the fault is seeded: %+v", baseline.Diagnostics)
	}

	testOnlyForceUniformLoanJoin = true
	defer func() { testOnlyForceUniformLoanJoin = false }()
	faulted := Program(mustParseProgram(t, source))
	if len(faulted.Diagnostics) == 0 {
		t.Fatalf("seeded uniform-join fault left the accept fixture's verdict unchanged — the fault is not load-bearing")
	}
}

// TestOracleDisagreesWithUniformJoinFault is the 03-05-02 mutation kill for
// the uniform-join fault, at the ENDPOINT level rather than the admission
// level TestUniformJoinPlacementFlipsBothVerdicts already proves.
//
// testOnlyForceUniformLoanJoin (03-03's seam) only touches
// discoverLoanLastUses' consumer inside analyzeArmBody -- the admission
// decision (conflict/expiry) -- and never touches loanLivenessFixpoint /
// materializeLoanEndpoints, which is checkBranch's SOLE producer of the
// observable core.LoanEndpoint records this oracle recomputes. Seeding the
// fault therefore does not change the honestly-checked function's
// LoanEndpoints at all (proven first, below) — which is exactly why a
// SEPARATE, endpoint-level corruption is required to exercise the oracle's
// own disagreement path: this test constructs the endpoint set "uniform
// join placement" would have produced had it also corrupted
// materializeLoanEndpoints (every loan's endpoint forced to an EDGE on its
// arm's own edge-to-join, instead of the real POINT where it is actually
// last referenced) and confirms the independently recomputed oracle answer
// disagrees with it, naming the specific loan and edge.
func TestOracleDisagreesWithUniformJoinFault(t *testing.T) {
	source := readTestdataFixture(t, "branch_one_arm_shared_accept.lang")

	honest := Program(mustParseProgram(t, source))
	if len(honest.Diagnostics) != 0 {
		t.Fatalf("accept fixture unexpectedly rejected: %+v", honest.Diagnostics)
	}
	function := honest.Program.Functions[0]
	if len(function.Linear.LoanEndpoints) == 0 {
		t.Fatalf("fixture carries no LoanEndpoints to corrupt: %+v", function.Linear)
	}

	// Reconfirm (independently of TestUniformJoinPlacementFlipsBothVerdicts)
	// that seeding the ADMISSION-level fault does not perturb the
	// endpoint-materializing pass at all: the fault only ever engages inside
	// analyzeArmBody, which checkBranch calls BEFORE loanLivenessFixpoint /
	// materializeLoanEndpoints run on the same, already-lowered operations.
	testOnlyForceUniformLoanJoin = true
	faultedAdmission := Program(mustParseProgram(t, source))
	testOnlyForceUniformLoanJoin = false
	if len(faultedAdmission.Diagnostics) == 0 {
		t.Fatalf("uniform-join admission fault did not reject the fixture as TestUniformJoinPlacementFlipsBothVerdicts requires")
	}

	// Build "what uniform join placement would have produced" at the
	// endpoint level: every loan the fixture actually carries a POINT
	// endpoint for, replaced by an EDGE endpoint on its own arm's
	// edge-to-join -- the literal meaning of "every loan ends uniformly at
	// the join" applied to materializeLoanEndpoints's own output shape.
	edgeToJoin := map[string]string{}
	for _, edge := range function.Linear.Edges {
		if edge.ToBlockID == function.ID+":block:join" {
			edgeToJoin[edge.FromBlockID] = edge.ID
		}
	}
	corrupted := make([]core.LoanEndpoint, 0, len(function.Linear.LoanEndpoints))
	for _, endpoint := range function.Linear.LoanEndpoints {
		edgeID, ok := edgeToJoin[endpoint.BlockID]
		if endpoint.Kind != "point" || !ok {
			corrupted = append(corrupted, endpoint)
			continue
		}
		corrupted = append(corrupted, core.LoanEndpoint{
			ID: edgeID + ":" + endpoint.LoanID, LoanID: endpoint.LoanID, Kind: "edge", EdgeID: edgeID,
		})
	}

	oracleEndpoints, work, err := pathoracle.RecomputeEndpoints(function)
	if err != nil {
		t.Fatalf("unexpected oracle error: %v", err)
	}
	if work == 0 {
		t.Fatalf("oracle reported zero work")
	}
	if reflect.DeepEqual(oracleEndpoints, corrupted) {
		t.Fatalf("oracle agreed with the uniform-join-corrupted endpoint set -- the differential is not load-bearing")
	}

	// The disagreement must name the specific loan and edge: the oracle's
	// own answer for the loan the fixture actually carries a loan for must
	// be a POINT endpoint (not the corrupted EDGE endpoint).
	var loanID string
	for _, endpoint := range function.Linear.LoanEndpoints {
		if endpoint.Kind == "point" {
			loanID = endpoint.LoanID
			break
		}
	}
	if loanID == "" {
		t.Fatalf("fixture carries no point endpoint to name in the disagreement: %+v", function.Linear.LoanEndpoints)
	}
	foundPoint, foundCorruptedEdge := false, false
	for _, endpoint := range oracleEndpoints {
		if endpoint.LoanID == loanID && endpoint.Kind == "point" {
			foundPoint = true
		}
	}
	for _, endpoint := range corrupted {
		if endpoint.LoanID == loanID && endpoint.Kind == "edge" {
			foundCorruptedEdge = true
		}
	}
	if !foundPoint || !foundCorruptedEdge {
		t.Fatalf("disagreement did not name loan %q by both its real point endpoint (%v) and the corrupted edge claim (%v)", loanID, oracleEndpoints, corrupted)
	}
}

// metamorphicArmSource builds a 2-arm branch module whose "On" arm holds two
// STRUCTURALLY INDEPENDENT shared-loan pairs (a/observedA and b/observedB —
// neither pair reads the other, and both loans are on the same owner, which
// is legal since shared+shared never conflicts) followed by a move that
// comes after both loans' last use. order swaps which pair is bound first
// (a genuine reordering of independent bindings); suffix alpha-renames both
// pairs' identifiers (a genuine alpha-rename) — together the metamorphic
// transformation 03-05-02 requires. The transformed program still checks
// clean and still carries exactly the same NUMBER of loans/endpoints; only
// their names and the operation ordinals they land on move.
func metamorphicArmSource(order int, suffix string) []byte {
	nameA, nameB := "a"+suffix, "b"+suffix
	obsA, obsB := "observedA"+suffix, "observedB"+suffix
	first, second := nameA, nameB
	firstObs, secondObs := obsA, obsB
	if order%2 == 1 {
		first, second = nameB, nameA
		firstObs, secondObs = obsB, obsA
	}
	body := fmt.Sprintf("      let %s = borrow flag\n      let %s = %s\n      let %s = borrow flag\n      let %s = %s\n      let moved%s = take flag\n      moved%s\n",
		first, firstObs, first, second, secondObs, second, suffix, suffix)
	return []byte(fmt.Sprintf(`module owned.branch_metamorphic

export {
  type Switch
  fn choose
}

data Switch =
  | On
  | Off

fn choose(flag: Switch) -> Switch {
  match flag {
    On => {
%s    }
    Off => {
      let held = take flag
      held
    }
  }
}
`, body))
}

// TestOracleMetamorphicTrials proves the oracle-vs-production differential
// holds across a fixed-seed batch of syntactic reorderings of independent
// bindings within an arm body (order-independent bindings commute: neither
// reads the other) combined with alpha-renaming of the places involved, and
// that the batch itself did not silently run fewer trials than requested
// (D-10's own "ask what a green property test actually reaches" standard,
// applied to trial COUNT, not just pass/fail).
func TestOracleMetamorphicTrials(t *testing.T) {
	const trials = 64
	ran := 0
	for seed := 0; seed < trials; seed++ {
		ran++
		source := metamorphicArmSource(seed, fmt.Sprintf("%d", seed))
		parsed := syntax.Parse(source)
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("seed=%d: generated source failed to parse: %+v\n%s", seed, parsed.Diagnostics, source)
		}
		result := Program(parsed.Program)
		if len(result.Diagnostics) != 0 {
			t.Fatalf("seed=%d: generated fixture unexpectedly rejected: %+v\n%s", seed, result.Diagnostics, source)
		}
		function := result.Program.Functions[0]
		if len(function.Linear.LoanEndpoints) != 2 {
			t.Fatalf("seed=%d: want exactly 2 loan endpoints (both independent shared loans), got %+v", seed, function.Linear.LoanEndpoints)
		}
		want := append([]core.LoanEndpoint(nil), function.Linear.LoanEndpoints...)
		got, work, err := pathoracle.RecomputeEndpoints(function)
		if err != nil {
			t.Fatalf("seed=%d: unexpected oracle error: %v", seed, err)
		}
		if work == 0 {
			t.Fatalf("seed=%d: oracle reported zero work", seed)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("seed=%d: oracle disagreed with production: got %+v want %+v", seed, got, want)
		}
	}
	if ran != trials {
		t.Fatalf("metamorphic batch ran %d trials, want %d — a generator that silently stops early must fail this assertion", ran, trials)
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

// reborrowChainOperations builds a single block's worth of operations
// forming a chain of n reborrows-of-reborrows (op[i] reborrows op[i-1]'s
// own target), the exact shape that produced D-02-03's recorded quadratic
// regression: each successive loan's chain ancestry grows by one, so a
// derivation that copies its ancestor list per operation is Θ(N²) in n.
func reborrowChainOperations(n int) []core.LinearOperation {
	operations := make([]core.LinearOperation, n)
	source := "chain:place:0"
	for index := 0; index < n; index++ {
		target := fmt.Sprintf("chain:place:%d", index+1)
		operations[index] = core.LinearOperation{
			ID: fmt.Sprintf("chain:op:%d", index), Kind: core.OpBorrowShared,
			SourceID: source, TargetID: target, LoanID: fmt.Sprintf("chain:loan:%d", index),
		}
		source = target
	}
	return operations
}

// TestLivenessWorkScale is 03-03-03's honest-work series (D-05/D-02-03):
// the fixpoint's counted work over a reborrow chain of 10, 100, 1,000, and
// 10,000 operations must grow within a linear factor of operation count,
// not quadratically. The bound (4*n+4) is re-derived here, not bumped: one
// unit per transfer-function evaluation (one per block, here always one
// block so a constant), plus at most one unit per operation for the
// backward scan's own loan-chain walk -- amortized O(1) per operation since
// blockLoanLiveness's recorded-guard stops each chain walk the instant it
// reaches an already-recorded loan (see derivePlaceLoans' doc comment), so
// the total chain-walk work across the whole block is bounded by the
// number of distinct loans, not by chain depth times operation count.
func TestLivenessWorkScale(t *testing.T) {
	for _, n := range []int{10, 100, 1_000, 10_000} {
		block := cfgBlockSpec{id: "chain:block:straight", operations: reborrowChainOperations(n), successors: nil}
		result, err := loanLivenessFixpoint("chain", []cfgBlockSpec{block})
		if err != nil {
			t.Fatalf("n=%d: unexpected acyclicity error: %v", n, err)
		}
		if result.work == 0 {
			t.Fatalf("n=%d: zero counted work", n)
		}
		if bound := 4*n + 4; result.work > bound {
			t.Fatalf("n=%d: work=%d exceeded the re-derived linear bound %d", n, result.work, bound)
		}
	}
}

// TestReborrowChainWorkIsLinear asserts the growth RATIO directly (not just
// a fixed bound): doubling-and-more operation counts must not multiply
// counted work by more than a small constant, which a reintroduced
// quadratic scan would violate (10x operations would cost ~100x work, not
// ~10x).
func TestReborrowChainWorkIsLinear(t *testing.T) {
	series := []int{10, 100, 1_000, 10_000}
	work := make([]int, len(series))
	for index, n := range series {
		block := cfgBlockSpec{id: "chain:block:ratio", operations: reborrowChainOperations(n), successors: nil}
		result, err := loanLivenessFixpoint("chain", []cfgBlockSpec{block})
		if err != nil {
			t.Fatalf("n=%d: unexpected acyclicity error: %v", n, err)
		}
		work[index] = result.work
	}
	for index := 1; index < len(series); index++ {
		operationRatio := float64(series[index]) / float64(series[index-1])
		workRatio := float64(work[index]) / float64(work[index-1])
		if workRatio > operationRatio*2 {
			t.Fatalf("n=%d->%d: operation count grew %.1fx but counted work grew %.1fx (%d->%d) -- looks quadratic",
				series[index-1], series[index], operationRatio, workRatio, work[index-1], work[index])
		}
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
		kinds := []string{"read", "take", "borrow", "borrow_mut"}
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
	access  string
	lastUse int
}

// oracleConflictingLoan independently re-derives the five-row conflict
// matrix (03-02): shared+shared never conflicts; every other overlapping
// combination (shared+exclusive, exclusive+shared, exclusive+exclusive)
// does. This is a direct, deliberately duplicated restatement of the same
// LAW check.go's conflictingLoan encodes, not a call into it -- the oracle
// must never import or call production conflict logic.
func oracleConflictingLoan(active map[string]testOracleLoan, ownerID, newAccess string) bool {
	for _, loan := range active {
		if loan.ownerID != ownerID {
			continue
		}
		if newAccess == "shared" && loan.access == "shared" {
			continue
		}
		return true
	}
	return false
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
		if candidate.RHS.Kind != "borrow" && candidate.RHS.Kind != "borrow_mut" {
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
		if candidate.RHS.Kind == "borrow" || candidate.RHS.Kind == "borrow_mut" {
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
			if oracleConflictingLoan(activeLoans, source.id, "shared") {
				result.DiagnosticCode = "ownership.borrow_conflict"
				return result
			}
			kind = core.OpBorrowShared
			loanID = fmt.Sprintf("%s:loan:%d", functionID, index)
			activeLoans[loanID] = testOracleLoan{ownerID: source.id, access: "shared", lastUse: lastUses[index]}
		case "borrow_mut":
			if !oracleHasAbility(typeFact.Abilities, core.AbilityShare) {
				result.DiagnosticCode = "ownership.borrow_requires_share"
				return result
			}
			if oracleConflictingLoan(activeLoans, source.id, "exclusive") {
				result.DiagnosticCode = "ownership.borrow_conflict"
				return result
			}
			kind = core.OpBorrowExclusive
			loanID = fmt.Sprintf("%s:loan:%d", functionID, index)
			activeLoans[loanID] = testOracleLoan{ownerID: source.id, access: "exclusive", lastUse: lastUses[index]}
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

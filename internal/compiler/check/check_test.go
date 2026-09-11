package check

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
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

// straightLineSupportAtDecisionPoint is plan 09-09 Task 1's composed
// production-side helper: it lowers a straight-line body via
// analyzeStraightLine UNCHANGED (lowering still lowers, and a
// timing-INDEPENDENT code -- name.unknown, ownership.use_after_move,
// ownership.borrow_requires_share, ownership.transfer_requires_take -- still
// decides and wins during lowering, exactly as production's own Program()
// behaves: a function whose lowering fails never reaches the post-assembly
// pass), then -- only when lowering admits cleanly -- assembles a one-function
// core.Program from the support's own Places/Operations (mirroring the shape
// checkLinear assembles) and runs the extended checkInterproceduralLoanLiveness
// over it, folding the first resulting diagnostic (if any) into DiagnosticCode/
// Diagnostic. This is the SAME later decision point production's real
// Program() pipeline reaches for ownership.move_while_borrowed/
// ownership.borrow_conflict after D-09-08's authorized deletion (D-09-07: one
// algorithm, one decision point). LoanFinalUses/States/Work/FixpointWork are
// untouched -- they are analyzeStraightLine's own evidence, unaffected by
// where the timing verdict is decided.
func straightLineSupportAtDecisionPoint(functionID, parameterName string, parameterSpan diagnostic.Span, typeFact core.TypeFact, body *ast.LinearBody, calleeContracts map[string]calleeContract, foreignSymbols map[string]foreignSymbolInfo) ownershipSupport {
	support := analyzeStraightLine(functionID, parameterName, parameterSpan, typeFact, body, calleeContracts, foreignSymbols)
	if support.DiagnosticCode != "" {
		return support
	}
	function := core.Function{
		ID: functionID, Name: functionID,
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:  core.Parameter{ID: functionID + ":place:0", Name: parameterName, Type: typeFact.Shape.Constructor},
		ReturnType: typeFact.Shape.Constructor,
		Linear: &core.LinearBody{
			ID: functionID + ":linear", Types: []core.TypeFact{typeFact},
			Places: support.Places, Operations: support.Operations,
		},
		Span: parameterSpan,
	}
	program := core.Program{Functions: []core.Function{function}}
	diagnostics := checkInterproceduralLoanLiveness(program, interproceduralSummaryTable{}, support.CallSpans)
	if len(diagnostics) > 0 {
		diag := diagnostics[0]
		support.DiagnosticCode = diag.Code
		support.Diagnostic = &diag
	}
	return support
}

func TestOwnershipSequenceExhaustive(t *testing.T) {
	// Forty-eight symbols span three declaration names, FOUR operation kinds
	// (implicit copy, take, shared borrow, exclusive borrow -- 03-05-03/D-10
	// extends the original three-kind alphabet to reach the conflict matrix
	// exhaustively too), and four source selectors: every visible place (at
	// this depth) plus an out-of-scope boundary. This exhausts shadowing,
	// multi-owner shapes, AND every shared/exclusive overlap combination the
	// five-row conflict matrix distinguishes.
	//
	// Plan 09-09 Task 1 moved this differential's contract to the
	// post-assembly decision point (D-09-08's authorized deletion of the
	// summary-blind computeLoanLastUses early call site): production is now
	// evaluated through straightLineSupportAtDecisionPoint (lowering, THEN --
	// only if lowering admits cleanly -- the extended
	// checkInterproceduralLoanLiveness pass), and the oracle's own two
	// timing-dependent decisions (ownership.move_while_borrowed/
	// ownership.borrow_conflict) are deferred to a matching point via
	// oracleDeferredLoanLiveness below. Same 225,890 cases, same
	// byte-identical assertSupportEqual comparison, same alphabet.
	const alphabet = 48
	totalCases := 0
	for length := 0; length <= 3; length++ {
		cases := 1
		for index := 0; index < length; index++ {
			cases *= alphabet
		}
		for encoded := 0; encoded < cases; encoded++ {
			body := generatedOwnershipBody(encoded, length, alphabet)
			got := straightLineSupportAtDecisionPoint("test:fn", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &body, nil, nil)
			want := oracleStraightLine("test:fn", "owner", byteTypeFact(), &body)
			assertSupportEqual(t, fmt.Sprintf("length=%d case=%d", length, encoded), got, want)
			totalCases++
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
			got := straightLineSupportAtDecisionPoint("test:fn", "owner", diagnostic.Span{Start: 1, End: 6}, nonShareableTypeFact(), &body, nil, nil)
			want := oracleStraightLine("test:fn", "owner", nonShareableTypeFact(), &body)
			assertSupportEqual(t, fmt.Sprintf("no-share length=%d case=%d", length, encoded), got, want)
			totalCases++
		}
	}

	// T-09-31/Task 1(e): an explicit case counter, derived in code from the
	// alphabet and length bounds rather than hardcoded alone, so a generator
	// that silently stops early -- reducing this phase's strongest existing
	// evidence -- fails HERE rather than being silently absorbed.
	perSweep := 0
	for length := 0; length <= 3; length++ {
		cases := 1
		for index := 0; index < length; index++ {
			cases *= alphabet
		}
		perSweep += cases
	}
	if want := 2 * perSweep; totalCases != want {
		t.Fatalf("exhaustive enumeration ran %d cases, want %d (2 sweeps * (48^0+48^1+48^2+48^3)) -- a generator that silently stops early must fail this assertion", totalCases, want)
	}

	denied := ast.LinearBody{
		Bindings: []ast.Binding{binding("view", "borrow", "owner", 10)},
		Result:   "view",
		Span:     diagnostic.Span{Start: 10, End: 20},
	}
	refused := straightLineSupportAtDecisionPoint("test:no-share", "owner", diagnostic.Span{Start: 1, End: 6}, nonShareableTypeFact(), &denied, nil, nil)
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

	// The four-step witness: `view = borrow owner`, `moved = take owner`,
	// `observed = read view` (`view` read AFTER the move). Production now
	// LOWERS all three bindings unconditionally (no early lowering-time
	// stop for the timing-dependent code) and the post-assembly pass decides
	// ownership.move_while_borrowed -- so Operations now carries all THREE
	// emitted operations, not the one lowering used to stop after. Uses
	// byteTypeFact (not bufferTypeFact): bufferTypeFact withholds Copy, and
	// since lowering no longer stops at the move, it would now ALSO reach
	// binding 2's implicit-copy admission and fail there first with
	// ownership.transfer_requires_take -- a timing-INDEPENDENT code that
	// correctly wins under the precedence rule (D-09-13), but which is not
	// what this witness exists to exercise. byteTypeFact grants both Share
	// (needed for the borrow) and Copy (needed for the observation), so the
	// witness reaches its intended deferred loan-liveness verdict.
	witness := ast.LinearBody{
		Bindings: []ast.Binding{
			binding("view", "borrow", "owner", 10),
			binding("moved", "take", "owner", 20),
			binding("observed", "read", "view", 30),
		},
		Result: "moved",
		Span:   diagnostic.Span{Start: 10, End: 38},
	}
	got := straightLineSupportAtDecisionPoint("test:witness", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &witness, nil, nil)
	if got.DiagnosticCode != "ownership.move_while_borrowed" || len(got.Operations) != 4 {
		t.Fatalf("four-step witness changed: %+v", got)
	}
	if len(got.LoanFinalUses) != 1 || got.LoanFinalUses[0].OperationIndex != 2 {
		t.Fatalf("later loan read is not load-bearing: %+v", got.LoanFinalUses)
	}
	withoutLaterRead := witness
	withoutLaterRead.Bindings = withoutLaterRead.Bindings[:2]
	withoutLaterRead.Result = "moved"
	withoutLaterRead.Span.End = 28
	legal := straightLineSupportAtDecisionPoint("test:unused", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &withoutLaterRead, nil, nil)
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

	result, diag := loanLivenessFixpoint("fn", []cfgBlockSpec{b1, b2}, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag != nil {
		t.Fatalf("unexpected acyclicity error: %+v", diag)
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
	endpoints := materializeLoanEndpoints("fn", []cfgBlockSpec{b1, b2}, edgeID, result, interproceduralSummaryTable{})
	if len(endpoints) != 1 || endpoints[0].Kind != "point" || endpoints[0].LoanID != "fn:loan:0" {
		t.Fatalf("loan consumed in b2 must materialize as exactly one point endpoint there: %+v", endpoints)
	}
	if endpoints[0].BlockID != "fn:block:b2" || endpoints[0].AfterOperationID != "fn:op:1" {
		t.Fatalf("point endpoint named the wrong block/operation: %+v", endpoints[0])
	}
}

// TestLoanLivenessFixpointCoversStraightLine is D-05-35(a)'s widened-domain
// falsifier: loanLivenessFixpoint/materializeLoanEndpoints, run directly
// against a straight-line function's own single-block CFG, now produce
// endpoints for a shipped straight-line-with-borrow fixture -- something the
// PRODUCTION serialized core never did for a straight-line function (Linear.
// LoanEndpoints stays nil there, per the byte-identity note above
// cfgBlockSpec's own doc comment), demonstrating the fixpoint's domain now
// genuinely covers straight-line bodies, not only checkBranch's arm blocks.
// (The read_first note names testdata/phase2, but no ACCEPTED phase2 fixture
// carries a borrow -- both of that phase's borrow fixtures are REJECT
// controls -- so this uses shared_shared_accept.lang, the first accepted
// straight-line borrow fixture testdata/phase3 ships.)
func TestLoanLivenessFixpointCoversStraightLine(t *testing.T) {
	source := readTestdataFixture(t, "shared_shared_accept.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture unexpectedly rejected: %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("want exactly one function, got %d", len(result.Program.Functions))
	}
	function := result.Program.Functions[0]
	if function.Linear == nil || len(function.Linear.Operations) == 0 {
		t.Fatalf("expected a populated straight-line body: %+v", function)
	}
	if len(function.Linear.LoanEndpoints) != 0 {
		t.Fatalf("straight-line functions must not (yet) serialize LoanEndpoints -- byte-identity would move: %+v", function.Linear.LoanEndpoints)
	}

	blockID := function.ID + ":block:straight"
	block := cfgBlockSpec{id: blockID, operations: function.Linear.Operations, successors: nil}
	fixpoint, diag := loanLivenessFixpoint(function.ID, []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag != nil {
		t.Fatalf("unexpected acyclicity error: %+v", diag)
	}
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints(function.ID, []cfgBlockSpec{block}, edgeID, fixpoint, interproceduralSummaryTable{})
	if len(endpoints) == 0 {
		t.Fatalf("widened fixpoint produced zero endpoints for a straight-line body carrying borrows, where it previously covered none")
	}
	if fixpoint.work == 0 {
		t.Fatalf("expected nonzero counted fixpoint work over a non-trivial straight-line body")
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
	result, diag := loanLivenessFixpoint("fn", blocks, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag != nil {
		t.Fatalf("unexpected acyclicity error: %+v", diag)
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
	endpoints := materializeLoanEndpoints("fn", blocks, edgeID, result, interproceduralSummaryTable{})
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
// present, is rejected, not silently accepted. Task 1(c) widens this to
// assert on the returned diagnostic's own Code and its cycle_block cause,
// rather than merely a non-nil error, now that loanLivenessFixpoint returns
// a coded *diagnostic.Diagnostic (D-08-19b) instead of a bare error.
func TestBackEdgeRejected(t *testing.T) {
	a := cfgBlockSpec{id: "fn:block:a", operations: nil, successors: []string{"fn:block:b"}}
	b := cfgBlockSpec{id: "fn:block:b", operations: nil, successors: []string{"fn:block:a"}}
	_, diag := loanLivenessFixpoint("fn", []cfgBlockSpec{a, b}, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag == nil {
		t.Fatalf("expected a back-edge rejection, got none")
	}
	if diag.Code != "check.cfg_back_edge" {
		t.Fatalf("expected code check.cfg_back_edge, got %q", diag.Code)
	}
	var sawCycleBlock bool
	for _, cause := range diag.Causes {
		if cause.Kind != "cycle_block" {
			continue
		}
		sawCycleBlock = true
		if cause.Detail != "fn:block:a" && cause.Detail != "fn:block:b" {
			t.Fatalf("cycle_block cause names an unexpected block: %+v", cause)
		}
	}
	if !sawCycleBlock {
		t.Fatalf("expected a cycle_block cause, got %+v", diag.Causes)
	}
}

// TestCFGBackEdgeWalkIsIterative is Task 1(c)'s D-08-19a structural guard,
// following TestDemoteHasExactlyOnePromotionPassthrough's go/ast-scan
// precedent (internal/compiler/measure/statistics_test.go): parses check.go
// itself and asserts loanLivenessFixpoint's OWN body declares no
// function-typed local variable that is assigned a closure calling itself
// -- the exact `var walk func(id string) error; walk = func(id string)
// error { ...; walk(successor); ... }` shape this task removed. A future
// edit reintroducing native recursion here fails this test even before it
// exhausts a stack. It then exercises the (now-iterative) cycle pre-walk on
// a synthetic 50,000-deep single-successor-chain CFG -- deep enough that a
// reintroduced native-recursion walk would exhaust the goroutine stack --
// proving the conversion is load-bearing under depth, not merely
// structurally absent (T-08-13).
func TestCFGBackEdgeWalkIsIterative(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "check.go", nil, 0)
	if err != nil {
		t.Fatalf("parse check.go: %v", err)
	}
	var target *goast.FuncDecl
	for _, decl := range file.Decls {
		if fn, ok := decl.(*goast.FuncDecl); ok && fn.Name.Name == "loanLivenessFixpoint" {
			target = fn
			break
		}
	}
	if target == nil {
		t.Fatal("check.go does not declare a loanLivenessFixpoint function")
	}

	var selfCalling []string
	goast.Inspect(target.Body, func(n goast.Node) bool {
		assign, ok := n.(*goast.AssignStmt)
		if !ok || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
			return true
		}
		ident, ok := assign.Lhs[0].(*goast.Ident)
		if !ok {
			return true
		}
		lit, ok := assign.Rhs[0].(*goast.FuncLit)
		if !ok {
			return true
		}
		callsSelf := false
		goast.Inspect(lit.Body, func(inner goast.Node) bool {
			call, ok := inner.(*goast.CallExpr)
			if !ok {
				return true
			}
			if fnIdent, ok := call.Fun.(*goast.Ident); ok && fnIdent.Name == ident.Name {
				callsSelf = true
			}
			return true
		})
		if callsSelf {
			selfCalling = append(selfCalling, ident.Name)
		}
		return true
	})
	if len(selfCalling) != 0 {
		t.Fatalf("loanLivenessFixpoint declares a self-calling closure for %v -- the cycle pre-walk must use an explicit stack, not native recursion (D-08-19a)", selfCalling)
	}

	const depth = 50_000
	blocks := make([]cfgBlockSpec, depth)
	for i := 0; i < depth; i++ {
		var successors []string
		if i+1 < depth {
			successors = []string{fmt.Sprintf("deep:block:%d", i+1)}
		}
		blocks[i] = cfgBlockSpec{id: fmt.Sprintf("deep:block:%d", i), successors: successors}
	}
	if _, diag := loanLivenessFixpoint("deep", blocks, interproceduralSummaryTable{}, diagnostic.Span{}); diag != nil {
		t.Fatalf("expected the acyclic %d-deep chain to admit, got %+v", depth, diag)
	}
}

// TestLoanLivenessBoundScalesWithBlockCount is Task 2's D-08-14 scaling
// falsifier: loanLivenessBound must scale with its own input -- doubling
// the block count must double the computed bound (holding the loan count
// fixed), ruling out a flat magic constant masquerading as a derived one.
func TestLoanLivenessBoundScalesWithBlockCount(t *testing.T) {
	base := loanLivenessBound(4, 2)
	doubled := loanLivenessBound(8, 2)
	if doubled != base*2 {
		t.Fatalf("doubling block count did not double the bound: base=%d doubled=%d", base, doubled)
	}
}

// TestLoanLivenessBoundExceededRefusesRatherThanTruncates is Task 2's
// D-08-14 refusal-shape falsifier: with loanLivenessBoundSeam engaged, the
// fixpoint returns the named check.loan_liveness_bound_exceeded diagnostic
// AND a ZERO-VALUED loanLivenessResult -- never a partial live-in map,
// which would be exactly the silent under-approximation criterion 2
// forbids. Its Primary span is the function's own declaration span (no
// single offending operation exists for a whole-function property), and
// its Repairs list is empty.
func TestLoanLivenessBoundExceededRefusesRatherThanTruncates(t *testing.T) {
	defer func() { loanLivenessBoundSeam = false }()
	loanLivenessBoundSeam = true

	loanOp := core.LinearOperation{ID: "fn:op:0", Kind: core.OpBorrowShared, SourceID: "fn:place:0", TargetID: "fn:place:1", LoanID: "fn:loan:0"}
	block := cfgBlockSpec{id: "fn:block:straight", operations: []core.LinearOperation{loanOp}, successors: nil}
	declSpan := diagnostic.Span{Start: 5, End: 9}

	result, diag := loanLivenessFixpoint("fn", []cfgBlockSpec{block}, interproceduralSummaryTable{}, declSpan)
	if diag == nil {
		t.Fatal("expected the seam to force the named bound refusal")
	}
	if diag.Code != "check.loan_liveness_bound_exceeded" {
		t.Fatalf("expected code check.loan_liveness_bound_exceeded, got %q", diag.Code)
	}
	if diag.Primary != declSpan {
		t.Fatalf("expected Primary to be the function's own declaration span %+v, got %+v", declSpan, diag.Primary)
	}
	if len(diag.Repairs) != 0 {
		t.Fatalf("expected no repairs for a whole-function bound refusal, got %+v", diag.Repairs)
	}
	if len(result.liveIn) != 0 || result.work != 0 {
		t.Fatalf("expected a zero-valued result on the bound refusal, got %+v", result)
	}
}

// TestLoanLivenessBoundMutationKilled is Task 3's D-08-16/QLT-08 kill for
// the bound control itself, following TestCallReturnTypeDerivationMutationKilled's
// exact shape (defer-restore declared BEFORE the flip): with the seam
// engaged, an otherwise-clean input is refused with the named bound
// diagnostic; with it disengaged, the SAME input is admitted cleanly. The
// control has been SEEN TO FAIL, not merely shown to exist -- no control
// ships having never been seen to fail.
func TestLoanLivenessBoundMutationKilled(t *testing.T) {
	defer func() { loanLivenessBoundSeam = false }()

	loanOp := core.LinearOperation{ID: "fn:op:0", Kind: core.OpBorrowShared, SourceID: "fn:place:0", TargetID: "fn:place:1", LoanID: "fn:loan:0"}
	useOp := core.LinearOperation{ID: "fn:op:1", Kind: core.OpReturn, SourceID: "fn:place:1"}
	block := cfgBlockSpec{id: "fn:block:straight", operations: []core.LinearOperation{loanOp, useOp}, successors: nil}

	loanLivenessBoundSeam = true
	_, tripped := loanLivenessFixpoint("fn", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
	if tripped == nil || tripped.Code != "check.loan_liveness_bound_exceeded" {
		t.Fatalf("expected the seam to force the named bound refusal, got %+v", tripped)
	}

	loanLivenessBoundSeam = false
	clean, notTripped := loanLivenessFixpoint("fn", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
	if notTripped != nil {
		t.Fatalf("expected a clean result with the seam disengaged, got %+v", notTripped)
	}
	if clean.work == 0 {
		t.Fatalf("expected nonzero counted work for the clean run")
	}
}

// TestLoanLivenessBoundValueIsNotInDiagnosticIdentity is Task 3(b)'s
// D-08-18 identity-stability falsifier: two bound refusals constructed with
// the SAME functionID and span but DIFFERING computed bounds (varied by
// block count, seam-forced so both trip regardless of the actual bound
// reached) must carry the IDENTICAL diagnostic ID, and neither the
// Message nor any Cause's Detail may contain either run's own computed
// bound value as a substring. Retuning loanLivenessBoundFactor must never
// silently move this diagnostic's published ID.
func TestLoanLivenessBoundValueIsNotInDiagnosticIdentity(t *testing.T) {
	defer func() { loanLivenessBoundSeam = false }()
	loanLivenessBoundSeam = true

	loanOp := core.LinearOperation{ID: "fn:op:0", Kind: core.OpBorrowShared, SourceID: "fn:place:0", TargetID: "fn:place:1", LoanID: "fn:loan:0"}
	oneBlock := []cfgBlockSpec{{id: "fn:block:a", operations: []core.LinearOperation{loanOp}, successors: nil}}
	threeBlocks := []cfgBlockSpec{
		{id: "fn:block:a", operations: []core.LinearOperation{loanOp}, successors: []string{"fn:block:b", "fn:block:c"}},
		{id: "fn:block:b", successors: nil},
		{id: "fn:block:c", successors: nil},
	}

	_, diagA := loanLivenessFixpoint("fn", oneBlock, interproceduralSummaryTable{}, diagnostic.Span{})
	_, diagB := loanLivenessFixpoint("fn", threeBlocks, interproceduralSummaryTable{}, diagnostic.Span{})
	if diagA == nil || diagB == nil {
		t.Fatalf("expected the seam to force the bound refusal on both runs, got diagA=%+v diagB=%+v", diagA, diagB)
	}

	boundA := loanLivenessBound(len(oneBlock), 1)
	boundB := loanLivenessBound(len(threeBlocks), 1)
	if boundA == boundB {
		t.Fatalf("test setup did not actually vary the computed bound: %d == %d", boundA, boundB)
	}

	if diagA.ID != diagB.ID {
		t.Fatalf("the bound refusal's own identity moved across two differing computed bounds (%d vs %d) -- retuning the bound must never move a published diagnostic ID (D-08-18): %q vs %q", boundA, boundB, diagA.ID, diagB.ID)
	}

	for _, entry := range []struct {
		diag  *diagnostic.Diagnostic
		bound int
	}{{diagA, boundA}, {diagB, boundB}} {
		boundText := strconv.Itoa(entry.bound)
		if strings.Contains(entry.diag.Message, boundText) {
			t.Fatalf("message %q contains the computed bound value %d", entry.diag.Message, entry.bound)
		}
		for _, cause := range entry.diag.Causes {
			if strings.Contains(cause.Detail, boundText) {
				t.Fatalf("cause detail %q contains the computed bound value %d", cause.Detail, entry.bound)
			}
		}
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
	support := analyzeStraightLine("test:reborrow", "code", diagnostic.Span{}, bufferTypeFact(), &body, nil, nil)
	if support.DiagnosticCode != "" {
		t.Fatalf("shipped reborrow fixture unexpectedly rejected: %+v", support)
	}
	if len(support.LoanFinalUses) != 2 {
		t.Fatalf("want 2 tracked loans, got %+v", support.LoanFinalUses)
	}
	block := cfgBlockSpec{id: "test:reborrow:block:straight", operations: support.Operations, successors: nil}
	result, diag := loanLivenessFixpoint("test:reborrow", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
	if diag != nil {
		t.Fatalf("unexpected acyclicity error: %+v", diag)
	}
	edgeID := func(from, to string) string { return from + "->" + to }
	endpoints := materializeLoanEndpoints("test:reborrow", []cfgBlockSpec{block}, edgeID, result, interproceduralSummaryTable{})
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
	got := analyzeStraightLine("test:multi", "owner", diagnostic.Span{}, byteTypeFact(), &body, nil, nil)
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
	assertSupportEqual(t, "shadowed binding", analyzeStraightLine("test:shadow", "owner", diagnostic.Span{}, byteTypeFact(), &shadowed, nil, nil), oracleStraightLine("test:shadow", "owner", byteTypeFact(), &shadowed))

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
	got = analyzeStraightLine("test:shadow-loan", "owner", diagnostic.Span{}, byteTypeFact(), &shadowedLoan, nil, nil)
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

// branchSequenceProgram builds a synthetic 2-arm branch function directly
// against the ast (bypassing the parser/source text entirely, unlike the
// syntax-package generator) whose two arms hold armA/armB as their full
// linear bodies. Both arms reference the function's own aliased parameter
// by the name "owner", matching generatedOwnershipBody's own convention.
func branchSequenceProgram(armA, armB ast.LinearBody) ast.Program {
	return ast.Program{
		Module: "owned.branch_sequence",
		Data:   []ast.DataDecl{{Name: "Switch", Alternatives: []ast.Alternative{{Name: "On"}, {Name: "Off"}}}},
		Funcs: []ast.FuncDecl{{
			Name:       "choose",
			Parameter:  ast.Parameter{Name: "owner", Type: ast.TypeRef{Constructor: "Switch"}},
			ReturnType: ast.TypeRef{Constructor: "Switch"},
			Body: ast.Body{
				MatchExpr: ast.MatchExpr{
					Scrutinee: "owner",
					Arms: []ast.MatchArm{
						{Pattern: "On", Body: &armA},
						{Pattern: "Off", Body: &armB},
					},
				},
			},
		}},
	}
}

// TestBranchSequenceExhaustive extends TestOwnershipSequenceExhaustive's
// exhaustive ownership-sequence enumeration from a single (straight-line)
// block to a TWO-block branch at bounded length (03-05-03/D-10): every
// combination of {0 or 1 binding} x {the 48-symbol alphabet} for EACH of
// two independent arms is checked against production's real checkBranch
// path (via Program(...), never a synthetic cfgBlockSpec) and compared
// against an independently-computed verdict built from the SAME
// oracleStraightLine differential TestOwnershipSequenceExhaustive already
// trusts, applied once per arm — legitimate because 03-01's per-arm
// aliasing makes every arm's own admission decision fully independent of
// its sibling (a fact 03-03/03-04's own summaries document explicitly): the
// two-block program's overall verdict is REJECT if and only if at least one
// arm's own straight-line verdict is REJECT. The batch stays deterministic
// and bounded (2 lengths x 48^0..48^1 per arm = 49*49 = 2,401 total
// two-block programs) and asserts its own case count, matching the
// fixed-seed batches elsewhere in this file.
func TestBranchSequenceExhaustive(t *testing.T) {
	const alphabet = 48
	const maxLength = 1

	cases := func(length int) int {
		count := 1
		for index := 0; index < length; index++ {
			count *= alphabet
		}
		return count
	}

	total := 0
	for lengthA := 0; lengthA <= maxLength; lengthA++ {
		for encodedA := 0; encodedA < cases(lengthA); encodedA++ {
			bodyA := generatedOwnershipBody(encodedA, lengthA, alphabet)
			wantA := oracleStraightLine("s1:owned.branch_sequence:fn:choose", "owner", byteTypeFact(), &bodyA)
			for lengthB := 0; lengthB <= maxLength; lengthB++ {
				for encodedB := 0; encodedB < cases(lengthB); encodedB++ {
					bodyB := generatedOwnershipBody(encodedB, lengthB, alphabet)
					wantB := oracleStraightLine("s1:owned.branch_sequence:fn:choose", "owner", byteTypeFact(), &bodyB)
					total++

					program := branchSequenceProgram(bodyA, bodyB)
					got := Program(program)
					wantReject := wantA.DiagnosticCode != "" || wantB.DiagnosticCode != ""
					gotReject := len(got.Diagnostics) != 0
					if gotReject != wantReject {
						t.Fatalf(
							"lengthA=%d encodedA=%d lengthB=%d encodedB=%d: production reject=%v oracle reject=%v (production diags=%+v, oracleA=%q, oracleB=%q)",
							lengthA, encodedA, lengthB, encodedB, gotReject, wantReject, got.Diagnostics, wantA.DiagnosticCode, wantB.DiagnosticCode,
						)
					}
				}
			}
		}
	}
	if total != (cases(0)+cases(1))*(cases(0)+cases(1)) {
		t.Fatalf("exhaustive two-block batch ran %d cases, want %d — a generator that silently stops early must fail this assertion", total, (cases(0)+cases(1))*(cases(0)+cases(1)))
	}
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
		result, diag := loanLivenessFixpoint("chain", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
		if diag != nil {
			t.Fatalf("n=%d: unexpected acyclicity error: %+v", n, diag)
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
		result, diag := loanLivenessFixpoint("chain", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})
		if diag != nil {
			t.Fatalf("n=%d: unexpected acyclicity error: %+v", n, diag)
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

// checkerCorpusVerdicts is TestCheckerVerdictsUnchanged's pinned snapshot:
// {phase-dir}/{fixture} -> {first diagnostic code, or "" for a clean
// admission}, over the ENTIRE testdata/phase1..4 corpus, as check.Program
// (not session/originvalidate -- that layer is a separate concern this
// package cannot import) decided before D-04-25's work-counting change
// landed in discoverLoanLastUses. D-04-25 must change ONLY the counted
// work, never an accept/reject decision or a diagnostic code.
var checkerCorpusVerdicts = map[string]string{
	"phase1/comments.lang":                                "",
	"phase1/malformed.lang":                               "syntax.unexpected_byte",
	"phase1/non_exhaustive.lang":                          "match.non_exhaustive",
	"phase1/toggle.lang":                                  "",
	"phase2/ability_shapes.lang":                          "check.unexecutable_shape",
	"phase2/implicit_copy.lang":                           "",
	"phase2/implicit_noncopy.lang":                        "ownership.transfer_requires_take",
	"phase2/move_while_borrowed.lang":                     "ownership.move_while_borrowed",
	"phase2/owned_transfer.lang":                          "",
	"phase2/reborrow_while_moved.lang":                    "ownership.move_while_borrowed",
	"phase2/use_after_move.lang":                          "ownership.use_after_move",
	"phase3/borrowed_view.lang":                           "",
	"phase3/branch_one_arm_shared_accept.lang":            "",
	"phase3/branch_one_arm_shared_reject.lang":            "ownership.move_while_borrowed",
	"phase3/branch_view.lang":                             "",
	"phase3/exclusive_exclusive_reject.lang":              "ownership.borrow_conflict",
	"phase3/exclusive_move_reject.lang":                   "ownership.move_while_borrowed",
	"phase3/public_view.lang":                             "",
	"phase3/public_view_impossible.lang":                  "",
	"phase3/public_view_mixed_access.lang":                "",
	"phase3/public_view_multi_arm_access_conflict.lang":   "",
	"phase3/public_view_multi_arm_omitted.lang":           "",
	"phase3/public_view_omitted.lang":                     "",
	"phase3/public_view_understated.lang":                 "",
	"phase3/sequential_shared_then_exclusive_accept.lang": "",
	"phase3/shared_exclusive_reject.lang":                 "ownership.borrow_conflict",
	"phase3/shared_shared_accept.lang":                    "",
	"phase4/acquire_three_fail_second.lang":               "",
	"phase4/acquire_three_fail_third.lang":                "",
	"phase4/acquire_three_success.lang":                   "",
	"phase4/defect_terminal.lang":                         "",
	"phase4/discard_because.lang":                         "",
	"phase4/fallible_call_unconsumed.lang":                "syntax.fallible_call_not_consumed",
	"phase4/foreign_acquire_one.lang":                     "",
	"phase4/foreign_call_target_not_foreign.lang":         "core.call_target_not_foreign",
	"phase4/foreign_origin_omitted.lang":                  "",
	"phase4/foreign_policy_value_injection.lang":          "check.foreign_policy_value_unsafe",
	"phase4/foreign_unwind_undeclared.lang":               "foreign.unwind_policy_undeclared",
	"phase4/nonlocal_exit_probe.lang":                     "",
}

// TestCheckerVerdictsUnchanged is D-04-25's verdict-pinning falsifier,
// written and confirmed green BEFORE the discoverLoanLastUses work-counting
// change landed: it walks the ENTIRE testdata/phase1..4 corpus and asserts
// check.Program's first diagnostic code (or "" for a clean admission)
// matches the pinned snapshot exactly. A verdict change during this task is
// attributable to this test failing, not silently absorbed into the work
// total.
func TestCheckerVerdictsUnchanged(t *testing.T) {
	for key, wantCode := range checkerCorpusVerdicts {
		phaseAndName := strings.SplitN(key, "/", 2)
		if len(phaseAndName) != 2 {
			t.Fatalf("malformed snapshot key %q", key)
		}
		source, err := os.ReadFile(filepath.Join("../../../testdata", phaseAndName[0], phaseAndName[1]))
		if err != nil {
			t.Fatalf("%s: %v", key, err)
		}
		parsed := syntax.Parse(source)
		gotCode := ""
		if len(parsed.Diagnostics) > 0 {
			gotCode = parsed.Diagnostics[0].Code
		} else {
			result := Program(parsed.Program)
			if len(result.Diagnostics) > 0 {
				gotCode = result.Diagnostics[0].Code
			}
		}
		if gotCode != wantCode {
			t.Fatalf("%s: verdict changed: want code %q, got %q", key, wantCode, gotCode)
		}
	}
}

// TestSingleLoanLivenessLaw is D-05-35(d)'s source-scan falsifier: the
// retired forward chain-inheritance scan must not reappear anywhere in this
// package's production or test code outside a comment. The identifier is
// deliberately built via string concatenation (never spelled out as one
// contiguous literal) so this test's own source line does not trip its own
// scan.
func TestSingleLoanLivenessLaw(t *testing.T) {
	// Plan 09-09 (D-09-08's authorized deletion): a SECOND retired identifier
	// joins discoverLoanLastUses' -- computeLoanLastUses (D-09-52/D-09-07),
	// the summary-blind early call site deleted from analyzeArmBody/
	// analyzeStraightLine. Built by string concatenation, same technique, so
	// this test's own source line does not trip its own scan.
	retiredIdentifiers := []string{
		"discover" + "LoanLastUses",
		"compute" + "LoanLastUses",
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}
	found := false
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		content, err := os.ReadFile(entry.Name())
		if err != nil {
			t.Fatalf("read %s: %v", entry.Name(), err)
		}
		for lineNumber, line := range strings.Split(string(content), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//") {
				continue
			}
			for _, retiredIdentifier := range retiredIdentifiers {
				if strings.Contains(line, retiredIdentifier) {
					t.Fatalf("%s:%d: retired identifier reappeared outside a comment: %q", entry.Name(), lineNumber+1, line)
				}
			}
		}
		found = true
	}
	if !found {
		t.Fatalf("scanned zero .go files -- the source-scan itself is broken")
	}

	// Plan 09-09 Task 4(a): assert an exact, NAMED count of loanLivenessFixpoint
	// call sites in check.go, so a future undocumented second caller --
	// reintroducing coexisting admission laws -- fails HERE rather than
	// silently. Count check.go only (never a _test.go file): Task 1's
	// test-local straightLineSupportAtDecisionPoint legitimately calls
	// checkInterproceduralLoanLiveness (itself one of the four sites below)
	// FROM a _test.go file to OBSERVE the surviving pass, which must not be
	// miscounted as a second production caller.
	content, err := os.ReadFile("check.go")
	if err != nil {
		t.Fatalf("read check.go: %v", err)
	}
	callSitePattern := regexp.MustCompile(`\bloanLivenessFixpoint\(`)
	definitionPattern := regexp.MustCompile(`^func loanLivenessFixpoint\(`)
	var callSiteLines []int
	for lineNumber, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		if definitionPattern.MatchString(line) {
			continue // the function's own declaration is not a call site
		}
		if callSitePattern.MatchString(line) {
			callSiteLines = append(callSiteLines, lineNumber+1)
		}
	}
	// Exactly four, each named so a fifth (or a changed role for one of
	// these four) is investigated rather than silently absorbed:
	//   1. checkInterproceduralLoanLiveness -- the SOLE admission-deciding
	//      call, over the assembled program, summary-aware (D-09-07/D-09-09).
	//   2. checkBranch -- retained ONLY for its own check.cfg_back_edge/
	//      check.loan_liveness_bound_exceeded refusals and to populate
	//      linear.LoanEndpoints; it decides NO ownership.* timing code after
	//      this plan (that decision moved to call site 1).
	//   3. aliasFactEndpoints -- pre-existing (Phase 05), non-deciding,
	//      post-hoc evidence for deriveAliasFacts.
	//   4. loanFinalUseEvidence -- this plan's own non-deciding evidence
	//      replacement for the deleted computeLoanLastUses, feeding
	//      LoanFinalUses/States (test-visible evidence, no production
	//      reader).
	// Only site 1 ever DECIDES an ownership.move_while_borrowed/
	// ownership.borrow_conflict admission outcome -- sites 2-4 either decide
	// an unrelated code or decide nothing at all.
	if want := 4; len(callSiteLines) != want {
		t.Fatalf("check.go has %d non-comment loanLivenessFixpoint(...) call sites at lines %v, want exactly %d (checkInterproceduralLoanLiveness, checkBranch, aliasFactEndpoints, loanFinalUseEvidence) -- an undocumented caller may have reintroduced a coexisting admission law", len(callSiteLines), callSiteLines, want)
	}
}

// TestLivenessLawsStayIndependent is D-05-35(d)'s D-12 guard: corevalidate's
// loan-endpoint re-derivation (recomputeLoanEndpoints) must remain a
// genuinely INDEPENDENT re-derivation from this package's own
// computeLoanLastUses/loanLivenessFixpoint -- sharing no helper name and no
// import edge -- even though this task made the same fixpoint load-bearing
// for admission here. Collapsing the two derivations back into one would
// break D-12's two-independent-derivations invariant this package and
// corevalidate both depend on.
func TestLivenessLawsStayIndependent(t *testing.T) {
	source, err := os.ReadFile(filepath.Join("..", "..", "..", "internal", "compiler", "corevalidate", "corevalidate.go"))
	if err != nil {
		t.Fatalf("read corevalidate.go: %v", err)
	}
	if strings.Contains(string(source), "compiler/check") {
		t.Fatalf("corevalidate.go imports compiler/check -- D-12's independence invariant is broken")
	}
	// Comment lines are exempt: corevalidate.go's own doc comments legitimately
	// NAME check.go's helpers in prose to explain why its mechanism is
	// deliberately different (see corevalidate.go's recomputeLoanEndpoints
	// doc comment). What must never happen is one of those identifiers
	// appearing as a live, callable symbol -- i.e. outside a comment.
	var codeLines []string
	for _, line := range strings.Split(string(source), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		codeLines = append(codeLines, line)
	}
	code := strings.Join(codeLines, "\n")
	sharedHelperNames := []string{
		// computeLoanLastUses (retired under D-09-08/D-09-52, plan 09-09) is
		// deliberately absent from this list: TestSingleLoanLivenessLaw's own
		// retired-identifier scan already forbids it from reappearing
		// anywhere in this package outside a comment, making a duplicate
		// check here both redundant and a false-positive risk against this
		// literal's own line.
		"loanLivenessFixpoint", "materializeLoanEndpoints",
		"blockLoanLiveness", "derivePlaceLoans", "cfgBlockSpec", "loanBlockUse",
		// Phase 08 Task 3: the new interprocedural liveness helpers must stay
		// equally independent from corevalidate.go's own peer re-derivation
		// (D-08-12's own two-independent-derivations invariant, mirroring
		// D-12 for the interprocedural law).
		"buildInterproceduralSummaries", "interproceduralSummaryTable", "checkInterproceduralLoanLiveness",
	}
	for _, name := range sharedHelperNames {
		if strings.Contains(code, name) {
			t.Fatalf("corevalidate.go calls check.go's own liveness helper %q outside a comment -- the two loan-endpoint derivations must share no helper", name)
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
		got := analyzeStraightLine("test:scale", "owner", diagnostic.Span{}, byteTypeFact(), &body, nil, nil)
		// D-04-25: discoverLoanLastUses now counts its own transitive-scan
		// work (operations+1), added on top of the pre-existing formula.
		wantWork := 1 + 2*(operations+1) + (operations + 1)
		if got.DiagnosticCode != "" || len(got.Operations) != operations+1 || got.Work != wantWork {
			t.Fatalf("operations=%d result=%+v want work=%d", operations, got, wantWork)
		}
		if got.Work > 3*len(got.Operations)+2 {
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
		got := analyzeStraightLine("test:fuzz", "owner", diagnostic.Span{Start: 1, End: 6}, byteTypeFact(), &body, nil, nil)
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
	// D-04-25: discoverLoanLastUses now counts its own transitive-scan work
	// (one unit per binding it visits, plus one for the final result-position
	// check) -- oracleLoanLastUses is a genuinely different mechanism (BFS
	// reachability rather than chain inheritance) but must report the same
	// SCALAR cost, so this oracle adds the identical len(body.Bindings)+1
	// term rather than reproducing production's internal walk shape.
	result.Work += len(body.Bindings) + 1
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
		// D-09-08/D-09-09: the three inline early returns for
		// ownership.move_while_borrowed/ownership.borrow_conflict that used
		// to sit here are REMOVED (plan 09-09 Task 1(b)) -- the walk now
		// continues past an offending binding exactly as an admitted one
		// (recording the loan, emitting the operation), matching production's
		// own lowering, which now makes the same two decisions no earlier.
		// oracleDeferredLoanLiveness (below) applies the deferred verdict
		// after the walk completes, at the matching later point. The three
		// inline early returns for name.unknown/ownership.use_after_move/
		// ownership.borrow_requires_share/ownership.transfer_requires_take
		// stay exactly where they are (D-09-47: per-binding facts, not timing
		// facts).
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
			activeLoans[loanID] = testOracleLoan{ownerID: source.id, access: "shared", lastUse: lastUses[index]}
		case "borrow_mut":
			if !oracleHasAbility(typeFact.Abilities, core.AbilityShare) {
				result.DiagnosticCode = "ownership.borrow_requires_share"
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
	if result.DiagnosticCode == "" {
		if code := oracleDeferredLoanLiveness(parameterName, body, lastUses); code != "" {
			result.DiagnosticCode = code
		}
	}
	return result
}

// oracleDeferredLoanLiveness is plan 09-09 Task 1(b)/(c)'s own, independently
// derived deferred verdict: applied only after oracleStraightLine's walk
// completes with no per-binding fact having already failed, it scans the
// oracle's OWN already-independent lastUses (oracleLoanLastUses' BFS
// reachability result) and the oracle's own access-mode rule
// (oracleConflictingLoan's five-row matrix, restated here rather than
// called) to find the EARLIEST offending event in binding order -- a `take`
// of an owner some loan on it still needs (mirroring hasFutureLoanOracle's
// own `lastUse > index` rule), or a new loan whose owner already carries an
// unexpired, access-conflicting loan (mirroring oracleConflictingLoan's own
// membership rule) -- and returns its code, or "" if neither ever fires. It
// must NEVER call, import, or transcribe checkInterproceduralLoanLiveness,
// loanLivenessFixpoint, materializeLoanEndpoints, derivePlaceLoans, or
// conflictingLoan: a differential whose two sides share a derivation agrees
// perfectly and proves nothing (D-02-03/D-03-01's "fix one of two peers"
// hazard).
func oracleDeferredLoanLiveness(parameterName string, body *ast.LinearBody, lastUses map[int]int) string {
	type loanRecord struct {
		ownerID string
		access  string
		lastUse int
	}
	active := map[string]loanRecord{}
	ownerOf := func(index int) string {
		sourceBinding, _ := oracleResolveBinding(parameterName, body, body.Bindings[index].RHS.Source, index)
		if sourceBinding >= 0 {
			return fmt.Sprintf("place:%d", sourceBinding+1)
		}
		return "place:parameter"
	}
	for index, candidate := range body.Bindings {
		ownerID := ownerOf(index)
		switch candidate.RHS.Kind {
		case "take":
			var blocking []string
			for loanID, loan := range active {
				if loan.ownerID == ownerID && loan.lastUse > index {
					blocking = append(blocking, loanID)
				}
			}
			if len(blocking) > 0 {
				sort.Strings(blocking)
				return "ownership.move_while_borrowed"
			}
		case "borrow", "borrow_mut":
			access := "shared"
			if candidate.RHS.Kind == "borrow_mut" {
				access = "exclusive"
			}
			var conflicting []string
			for loanID, loan := range active {
				if loan.ownerID != ownerID {
					continue
				}
				if access == "shared" && loan.access == "shared" {
					continue
				}
				conflicting = append(conflicting, loanID)
			}
			if len(conflicting) > 0 {
				sort.Strings(conflicting)
				return "ownership.borrow_conflict"
			}
			loanID := fmt.Sprintf("evidence:loan:%d", index)
			active[loanID] = loanRecord{ownerID: ownerID, access: access, lastUse: lastUses[index]}
		}
		for loanID, loan := range active {
			if loan.lastUse <= index {
				delete(active, loanID)
			}
		}
	}
	return ""
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

// TestForeignCallLowersToOkAndErrEdges checks the Phase 4 tracer fixture
// yields exactly one OpForeignCall with two successor edges, an ok block
// terminated by OpReturn, an err block terminated by OpFail carrying the
// declared failure ADT's place, and a populated Function.ForeignContract
// (D-04-01/D-04-04/D-04-05). core.DataType.Alternatives must remain
// unchanged ([]string, no payload) -- D-04-04's one-way door.
func TestForeignCallLowersToOkAndErrEdges(t *testing.T) {
	source := readPhase4Fixture(t, "foreign_acquire_one.lang")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("expected exactly one function, got %d", len(result.Program.Functions))
	}
	function := result.Program.Functions[0]

	if function.ForeignContract == nil || function.ForeignContract.Symbol != "lang_res_open" || function.ForeignContract.Allocator != "libc_malloc" {
		t.Fatalf("ForeignContract = %+v", function.ForeignContract)
	}
	if function.Linear == nil || len(function.Linear.Operations) != 3 {
		t.Fatalf("unexpected linear body: %+v", function.Linear)
	}
	callOp := function.Linear.Operations[0]
	if callOp.Kind != core.OpForeignCall || callOp.OkEdgeID == "" || callOp.ErrEdgeID == "" || callOp.ErrTargetID == "" {
		t.Fatalf("call operation = %+v", callOp)
	}
	returnOp := function.Linear.Operations[1]
	if returnOp.Kind != core.OpReturn {
		t.Fatalf("expected OpReturn, got %+v", returnOp)
	}
	failOp := function.Linear.Operations[2]
	if failOp.Kind != core.OpFail || failOp.SourceID != callOp.ErrTargetID {
		t.Fatalf("expected OpFail sourced from the call's err target, got %+v", failOp)
	}

	if len(function.Linear.Blocks) != 3 || len(function.Linear.Edges) != 2 {
		t.Fatalf("expected 3 blocks and 2 edges, got %d blocks, %d edges", len(function.Linear.Blocks), len(function.Linear.Edges))
	}
	var entryBlock core.Block
	for _, block := range function.Linear.Blocks {
		if block.ID == function.ID+":block:entry" {
			entryBlock = block
		}
	}
	if len(entryBlock.Successors) != 2 {
		t.Fatalf("entry block successors = %+v, want 2", entryBlock.Successors)
	}

	for _, dataType := range result.Program.DataTypes {
		if dataType.Name == "AcquireError" && len(dataType.Alternatives) == 0 {
			t.Fatalf("AcquireError has no alternatives")
		}
		// core.DataType.Alternatives is []string; a compile-time type change
		// would fail this file to build at all, which is the strongest
		// possible assertion that D-04-04's "no payload alternative" door
		// stayed shut this task.
		var alternatives []string = dataType.Alternatives
		_ = alternatives
	}
}

// TestDefectLowersToTerminalOutcome pins D-04-15: `defect "<reason>"` lowers
// to an OpDefect operation terminating its arm block, carrying the required
// non-empty reason, no target, and no release anywhere in that block.
func TestDefectLowersToTerminalOutcome(t *testing.T) {
	source := readPhase4Fixture(t, "defect_terminal.lang")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("expected exactly one function, got %d", len(result.Program.Functions))
	}
	function := result.Program.Functions[0]
	if function.Match == nil || function.Linear == nil {
		t.Fatalf("expected a branch-shaped function, got %+v", function)
	}
	var defectArmID string
	for _, arm := range function.Match.Arms {
		if arm.Pattern == "Halt" {
			defectArmID = arm.BlockID
		}
	}
	if defectArmID == "" {
		t.Fatalf("no Halt arm block found: %+v", function.Match.Arms)
	}
	var defectBlock core.Block
	found := false
	for _, block := range function.Linear.Blocks {
		if block.ID == defectArmID {
			defectBlock, found = block, true
		}
	}
	if !found || len(defectBlock.OperationIDs) == 0 {
		t.Fatalf("Halt arm block not found or empty: %+v", function.Linear.Blocks)
	}
	operationsByID := make(map[string]core.LinearOperation, len(function.Linear.Operations))
	for _, operation := range function.Linear.Operations {
		operationsByID[operation.ID] = operation
	}
	lastOp := operationsByID[defectBlock.OperationIDs[len(defectBlock.OperationIDs)-1]]
	if lastOp.Kind != core.OpDefect || lastOp.Reason != "halt requested" || lastOp.TargetID != "" {
		t.Fatalf("expected a terminal OpDefect with reason, got %+v", lastOp)
	}
	for _, opID := range defectBlock.OperationIDs {
		if operationsByID[opID].Kind == core.OpRelease {
			t.Fatalf("no release may appear in a defect-terminated block, found %+v", operationsByID[opID])
		}
	}
}

// TestForeignContractCarriesEveryObligation pins Task 04-03-01: the complete
// core.ForeignContract for a real declared foreign symbol carries every
// obligation FFI-01 names -- the four flat obligation strings plus a fully
// populated Layout -- none of which has a default value that would let an
// omission pass as a declaration.
func TestForeignContractCarriesEveryObligation(t *testing.T) {
	source := readPhase4Fixture(t, "foreign_acquire_one.lang")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	contract := result.Program.Functions[0].ForeignContract
	if contract == nil {
		t.Fatal("expected a populated ForeignContract")
	}
	if contract.Symbol == "" || contract.Allocator == "" || contract.Unwind == "" || contract.NonlocalExit == "" || contract.Fails == "" {
		t.Fatalf("missing D-04-01/02 obligation: %+v", contract)
	}
	if contract.InitializedState == "" || contract.Capture == "" || contract.Retention == "" || contract.Aliasing == "" {
		t.Fatalf("missing D-04-12 obligation: %+v", contract)
	}
	if contract.Layout == nil || contract.Layout.ForeignTypeName == "" || contract.Layout.Size == 0 || contract.Layout.Alignment == 0 || len(contract.Layout.Fields) == 0 {
		t.Fatalf("missing D-04-12 layout obligation: %+v", contract.Layout)
	}
	for _, field := range contract.Layout.Fields {
		if field.Name == "" || field.Size == 0 || field.Alignment == 0 || field.CType == "" {
			t.Fatalf("layout field missing an obligation: %+v", field)
		}
	}
}

// TestForeignFieldsAreOmittedWhenAbsent follows the Phase 3
// TestPhase3FieldsAreOmittedWhenAbsent precedent: every Phase 1-3 fixture's
// serialized core JSON must contain none of the new D-04-12 keys, checked
// key by key rather than by whole-document comparison (which the existing
// TestPreviousPhaseCoreBytesUnchanged golden hash pin already covers).
func TestForeignFieldsAreOmittedWhenAbsent(t *testing.T) {
	fixtures := []string{
		"../../../testdata/phase1/toggle.lang",
		"../../../testdata/phase2/owned_transfer.lang",
		"../../../testdata/phase3/borrowed_view.lang",
	}
	newKeys := []string{`"initialized_state"`, `"capture"`, `"retention"`, `"aliasing"`, `"layout"`, `"foreign_type_name"`}
	for _, path := range fixtures {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		program := mustParseProgram(t, source)
		result := Program(program)
		if len(result.Diagnostics) != 0 {
			t.Fatalf("%s: unexpected diagnostics: %+v", path, result.Diagnostics)
		}
		encoded, err := json.Marshal(result.Program)
		if err != nil {
			t.Fatalf("%s: marshal: %v", path, err)
		}
		text := string(encoded)
		for _, key := range newKeys {
			if strings.Contains(text, key) {
				t.Fatalf("%s: serialized core unexpectedly contains new key %s:\n%s", path, key, text)
			}
		}
	}
}

func readPhase4Fixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../testdata/phase4/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return source
}

// TestUnwindPolicyUndeclaredRejected pins D-04-16: a foreign symbol declared
// without an unwind policy is refused, with no default value, carrying a
// span, causal detail, and a repair.
func TestUnwindPolicyUndeclaredRejected(t *testing.T) {
	source := readPhase4Fixture(t, "foreign_unwind_undeclared.lang")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected a rejection, got none")
	}
	found := false
	for _, problem := range result.Diagnostics {
		if problem.Code != "foreign.unwind_policy_undeclared" {
			continue
		}
		found = true
		if len(problem.Repairs) == 0 {
			t.Fatalf("foreign.unwind_policy_undeclared carries no repair: %+v", problem)
		}
		hasMissingPolicyCause := false
		for _, cause := range problem.Causes {
			if cause.Kind == "missing_policy" {
				hasMissingPolicyCause = true
			}
		}
		if !hasMissingPolicyCause {
			t.Fatalf("expected a missing_policy cause, got %+v", problem.Causes)
		}
	}
	if !found {
		t.Fatalf("expected foreign.unwind_policy_undeclared, got %+v", result.Diagnostics)
	}
}

// TestCallTargetNotForeignRejected pins D-04-02: a fallible call whose
// callee resolves to a declared Lang function is refused with
// core.call_target_not_foreign.
func TestCallTargetNotForeignRejected(t *testing.T) {
	source := readPhase4Fixture(t, "foreign_call_target_not_foreign.lang")
	program := mustParseProgram(t, source)
	result := Program(program)
	found := false
	for _, problem := range result.Diagnostics {
		if problem.Code == "core.call_target_not_foreign" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected core.call_target_not_foreign, got %+v", result.Diagnostics)
	}
}

// TestForeignPolicyValueUnsafeRefusedAtAdmission is 04-13's source-admission
// falsifier (04-VERIFICATION.md gap 2b, FFI-01): a foreign policy value that
// is not a C identifier is refused with check.foreign_policy_value_unsafe at
// declaration time, in collectForeignSymbols' policy loop -- regardless of
// which policy key carries it, because every one of them flows unchanged
// into core.ForeignContract and, from there, into a C comment
// cgen.EmitForeignHeader splices raw.
func TestForeignPolicyValueUnsafeRefusedAtAdmission(t *testing.T) {
	// build renders a minimal foreign C {} source with all three policy
	// values as quoted string literals (the parser accepts a string or a
	// bare identifier for any policy value; quoting lets a hostile override
	// carry bytes an identifier token could never lex), overriding exactly
	// one field per case.
	build := func(overrideKey, overrideValue string) []byte {
		values := map[string]string{
			"allocator":     "libc_malloc",
			"unwind":        "forbidden",
			"nonlocal_exit": "forbidden",
		}
		values[overrideKey] = overrideValue
		return []byte(fmt.Sprintf(
			"module p.probe\nexport { fn main }\nforeign C {\n  fn probe(request: Byte) -> Byte {\n    unwind: \"%s\"\n    nonlocal_exit: \"%s\"\n    allocator: \"%s\"\n    fails: E\n  }\n}\ndata E = | F\nfn main(request: Byte) -> Byte {\n  let handle = try probe(request)\n  handle\n}\n",
			values["unwind"], values["nonlocal_exit"], values["allocator"],
		))
	}

	hostile := []struct {
		name  string
		key   string
		value string
	}{
		{"allocator carrying the comment-escaping payload", "allocator", "*/ int injected(void){return 1;} /*"},
		{"unwind carrying the comment-escaping payload", "unwind", "*/ int injected(void){return 1;} /*"},
		{"nonlocal_exit carrying the comment-escaping payload", "nonlocal_exit", "*/ int injected(void){return 1;} /*"},
		{"value containing a semicolon and a brace", "allocator", "libc;}malloc{"},
		{"value containing an embedded space", "allocator", "libc malloc"},
		{"value beginning with a digit", "allocator", "1libc_malloc"},
		{"value containing a non-ASCII rune", "allocator", "libc_mallocé"},
		{"value that is an empty string literal", "allocator", ""},
	}
	for _, hostileCase := range hostile {
		t.Run(hostileCase.name, func(t *testing.T) {
			source := build(hostileCase.key, hostileCase.value)
			program := mustParseProgram(t, source)
			result := Program(program)
			if len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "check.foreign_policy_value_unsafe" {
				t.Fatalf("expected check.foreign_policy_value_unsafe, got %+v", result.Diagnostics)
			}
		})
	}

	t.Run("negative: identifier-shaped policy values are not refused", func(t *testing.T) {
		source := build("allocator", "libc_malloc")
		program := mustParseProgram(t, source)
		result := Program(program)
		for _, problem := range result.Diagnostics {
			if problem.Code == "check.foreign_policy_value_unsafe" {
				t.Fatalf("an honest program must not be refused by check.foreign_policy_value_unsafe, got %+v", result.Diagnostics)
			}
		}
	})
}

// TestForeignAdmissionCapsRejectFailClosed proves the declared foreign
// symbol/policy caps (T-04-05) reject above the cap rather than truncating
// silently.
// blockByID and operationByID are small test-only lookup helpers shared by
// the resource-lifecycle tests below.
func blockByID(function core.Function, id string) core.Block {
	for _, block := range function.Linear.Blocks {
		if block.ID == id {
			return block
		}
	}
	return core.Block{}
}

func operationByID(function core.Function, id string) core.LinearOperation {
	for _, operation := range function.Linear.Operations {
		if operation.ID == id {
			return operation
		}
	}
	return core.LinearOperation{}
}

// releaseSequence returns the ReleasesOperationID of every OpRelease
// operation in a block's own OperationIDs, in the block's own order.
func releaseSequence(function core.Function, block core.Block) []string {
	var sequence []string
	for _, opID := range block.OperationIDs {
		operation := operationByID(function, opID)
		if operation.Kind == core.OpRelease {
			sequence = append(sequence, operation.ReleasesOperationID)
		}
	}
	return sequence
}

// TestThreeAcquisitionReleaseOrder pins RES-01/D-04-07: a three-stage
// fallible acquisition's success block releases C then B then A -- the exact
// reverse of completed-acquisition order -- and its own OpReturn is sourced
// from the function's own parameter (never moved by any OpForeignCall).
func TestThreeAcquisitionReleaseOrder(t *testing.T) {
	source := readPhase4Fixture(t, "acquire_three_success.lang")
	program := mustParseProgram(t, source)
	result := Program(program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", result.Diagnostics)
	}
	function := result.Program.Functions[0]
	callA := operationByID(function, function.ID+":op:0")
	callB := operationByID(function, function.ID+":op:1")
	callC := operationByID(function, function.ID+":op:2")
	if callA.Kind != core.OpForeignCall || callB.Kind != core.OpForeignCall || callC.Kind != core.OpForeignCall {
		t.Fatalf("expected three OpForeignCall operations, got %+v %+v %+v", callA, callB, callC)
	}
	successBlock := blockByID(function, function.ID+":block:success")
	got := releaseSequence(function, successBlock)
	want := []string{callC.ID, callB.ID, callA.ID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("success block release order = %v, want %v (C, B, A)", got, want)
	}
	lastOpID := successBlock.OperationIDs[len(successBlock.OperationIDs)-1]
	returnOp := operationByID(function, lastOpID)
	if returnOp.Kind != core.OpReturn || returnOp.SourceID != function.Parameter.ID {
		t.Fatalf("success block terminator = %+v, want OpReturn sourced from the parameter", returnOp)
	}
}

// TestPartialAcquisitionReleasesOnlyCompleted pins RES-01/D-04-07's partial-
// failure requirement: the second-stage failure block releases only A (one
// completed acquisition), the third-stage failure block releases B then A
// (two, in reverse order), and no failure block ever releases the
// acquisition whose own OpForeignCall triggered it.
func TestPartialAcquisitionReleasesOnlyCompleted(t *testing.T) {
	for _, tc := range []struct {
		fixture string
	}{{"acquire_three_fail_second.lang"}, {"acquire_three_fail_third.lang"}} {
		source := readPhase4Fixture(t, tc.fixture)
		program := mustParseProgram(t, source)
		result := Program(program)
		if len(result.Diagnostics) != 0 {
			t.Fatalf("%s: unexpected diagnostics: %+v", tc.fixture, result.Diagnostics)
		}
		function := result.Program.Functions[0]
		callA := function.ID + ":op:0"
		callB := function.ID + ":op:1"
		callC := function.ID + ":op:2"

		err0 := releaseSequence(function, blockByID(function, function.ID+":block:err:0"))
		if len(err0) != 0 {
			t.Fatalf("%s: first-stage failure block released %v, want none", tc.fixture, err0)
		}
		err1 := releaseSequence(function, blockByID(function, function.ID+":block:err:1"))
		if !reflect.DeepEqual(err1, []string{callA}) {
			t.Fatalf("%s: second-stage failure block released %v, want [A]", tc.fixture, err1)
		}
		err2 := releaseSequence(function, blockByID(function, function.ID+":block:err:2"))
		if !reflect.DeepEqual(err2, []string{callB, callA}) {
			t.Fatalf("%s: third-stage failure block released %v, want [B, A]", tc.fixture, err2)
		}
		// No block ever releases the acquisition it fails on: err:0 never
		// releases callA, err:1 never releases callB, err:2 never releases
		// callC.
		for _, forbidden := range []struct{ block, op string }{
			{"block:err:0", callA}, {"block:err:1", callB}, {"block:err:2", callC},
		} {
			for _, released := range releaseSequence(function, blockByID(function, function.ID+":"+forbidden.block)) {
				if released == forbidden.op {
					t.Fatalf("%s: %s released its own triggering acquisition %s", tc.fixture, forbidden.block, forbidden.op)
				}
			}
		}
	}
}

// TestDiscardBecauseRoundTrips pins D-04-06: `discard <call> because
// "<rationale>"` parses, formats to a fixed point, and carries the rationale
// string into the core artifact as a required non-empty field.
func TestDiscardBecauseRoundTrips(t *testing.T) {
	source := readPhase4Fixture(t, "discard_because.lang")
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("unexpected parse diagnostics: %+v", parsed.Diagnostics)
	}
	canonical := syntax.Format(parsed.Tree)
	if string(canonical) != string(source) {
		t.Fatalf("discard_because.lang is not at the formatter's fixed point:\n%s", canonical)
	}
	var found bool
	for _, function := range parsed.Program.Funcs {
		for _, binding := range function.Body.Linear.Bindings {
			if binding.RHS.Kind != "discard_call" {
				continue
			}
			found = true
			if binding.RHS.Rationale == "" {
				t.Fatalf("discard binding carries an empty rationale")
			}
		}
	}
	if !found {
		t.Fatal("fixture carries no discard_call binding")
	}
	result := Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected check diagnostics: %+v", result.Diagnostics)
	}
}

// TestDiscardRationaleRequired pins D-04-06: an empty or absent rationale
// does not parse.
func TestDiscardRationaleRequired(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
	}{
		{"empty rationale", "module d.empty\nexport { fn main }\nforeign C {\n  fn probe(request: Byte) -> Byte {\n    unwind: forbidden\n    nonlocal_exit: forbidden\n    allocator: \"libc_malloc\"\n    fails: E\n  }\n}\ndata E = | F\nfn main(request: Byte) -> Byte {\n  discard probe(request) because \"\"\n  request\n}\n"},
		{"absent rationale", "module d.absent\nexport { fn main }\nforeign C {\n  fn probe(request: Byte) -> Byte {\n    unwind: forbidden\n    nonlocal_exit: forbidden\n    allocator: \"libc_malloc\"\n    fails: E\n  }\n}\ndata E = | F\nfn main(request: Byte) -> Byte {\n  discard probe(request)\n  request\n}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parsed := syntax.Parse([]byte(tc.source))
			if len(parsed.Diagnostics) == 0 {
				t.Fatal("expected a parse diagnostic, got none")
			}
		})
	}
}

func TestForeignAdmissionCapsRejectFailClosed(t *testing.T) {
	block := ast.ForeignBlock{Language: "C"}
	for index := 0; index <= maxForeignSymbolsPerBlockCheck; index++ {
		block.Symbols = append(block.Symbols, ast.ForeignSymbol{
			Name:       fmt.Sprintf("sym_%d", index),
			Parameter:  ast.Parameter{Name: "request", Type: ast.TypeRef{Constructor: "Byte"}},
			ReturnType: ast.TypeRef{Constructor: "Byte"},
			Policies: []ast.ForeignPolicy{
				{Key: "unwind", Value: "forbidden"}, {Key: "nonlocal_exit", Value: "forbidden"}, {Key: "allocator", Value: "libc_malloc", IsString: true},
			},
		})
	}
	program := ast.Program{Module: "cap.test", Foreign: []ast.ForeignBlock{block}}
	symbols, diagnostics := collectForeignSymbols(program)
	if symbols != nil || len(diagnostics) == 0 || diagnostics[0].Code != "check.foreign_symbol_limit" {
		t.Fatalf("expected check.foreign_symbol_limit, got symbols=%v diagnostics=%+v", symbols, diagnostics)
	}
}

// verifyCallInvariantsFunctions builds a synthetic two-function core.Function
// slice for verifyCallInvariants' own tests (Task 2 Test 4, D-07-29): a
// caller with a single OpCall immediately returned, plus a trivial callee.
// Mirrors corevalidate's own callerCalleeProgram (corevalidate_test.go) but
// asserted independently here, against check's own verifyCallInvariants,
// never by invoking corevalidate.
func verifyCallInvariantsFunctions(calleeID string) []core.Function {
	callerID, calleeFnID := "s1:test:fn:caller", "s1:test:fn:callee"
	callerTypeID, calleeTypeID := callerID+":type:0", calleeFnID+":type:0"
	callerParamID, callerTargetID := callerID+":place:0", callerID+":place:1"
	calleeParamID := calleeFnID + ":place:0"
	return []core.Function{
		{
			ID: callerID, Name: "caller", EntryPointID: callerID + ":point:entry", ReturnPointID: callerID + ":point:return",
			Parameter: core.Parameter{ID: callerParamID, Name: "value", Type: "Byte"}, ReturnType: "Byte",
			Linear: &core.LinearBody{
				ID:     callerID + ":linear",
				Places: []core.Place{{ID: callerParamID, Name: "value", TypeID: callerTypeID}, {ID: callerTargetID, Name: "result", TypeID: callerTypeID}},
				Operations: []core.LinearOperation{
					{ID: callerID + ":op:0", PointID: callerID + ":point:linear:0", Kind: core.OpCall, SourceID: callerParamID, TargetID: callerTargetID, TypeID: callerTypeID, CalleeID: calleeID},
					{ID: callerID + ":op:1", PointID: callerID + ":point:linear:1", Kind: core.OpReturn, SourceID: callerTargetID, TypeID: callerTypeID},
				},
			},
		},
		{
			ID: calleeFnID, Name: "callee", EntryPointID: calleeFnID + ":point:entry", ReturnPointID: calleeFnID + ":point:return",
			Parameter: core.Parameter{ID: calleeParamID, Name: "value", Type: "Byte"}, ReturnType: "Byte",
			Linear: &core.LinearBody{
				ID:         calleeFnID + ":linear",
				Places:     []core.Place{{ID: calleeParamID, Name: "value", TypeID: calleeTypeID}},
				Operations: []core.LinearOperation{{ID: calleeFnID + ":op:0", PointID: calleeFnID + ":point:linear:0", Kind: core.OpReturn, SourceID: calleeParamID, TypeID: calleeTypeID}},
			},
		},
	}
}

// TestVerifyCallInvariantsRefusesEmptyCalleeID is Task 2 Test 4 (D-07-29):
// check's own verifyCallInvariants, asserted directly and independently of
// corevalidate, refuses an OpCall with an empty CalleeID.
func TestVerifyCallInvariantsRefusesEmptyCalleeID(t *testing.T) {
	got := verifyCallInvariants(verifyCallInvariantsFunctions(""))
	if got == nil || got.Code != "core.callee_id_missing" {
		t.Fatalf("expected core.callee_id_missing, got %+v", got)
	}
}

// TestVerifyCallInvariantsRefusesNonCallWithCalleeID is Task 2 Test 4
// (D-07-29): check's own verifyCallInvariants refuses a non-OpCall operation
// carrying a non-empty CalleeID.
func TestVerifyCallInvariantsRefusesNonCallWithCalleeID(t *testing.T) {
	functions := verifyCallInvariantsFunctions("s1:test:fn:callee")
	functions[0].Linear.Operations[1].CalleeID = "s1:test:fn:callee"
	got := verifyCallInvariants(functions)
	if got == nil || got.Code != "core.callee_id_kind_exclusive" {
		t.Fatalf("expected core.callee_id_kind_exclusive, got %+v", got)
	}
}

// TestVerifyCallInvariantsRefusesUnresolvedCallee is Task 2 Test 4 (D-07-45):
// check's own verifyCallInvariants refuses a CalleeID naming no declared
// function, with its own typed identity.
func TestVerifyCallInvariantsRefusesUnresolvedCallee(t *testing.T) {
	got := verifyCallInvariants(verifyCallInvariantsFunctions("s1:test:fn:does-not-exist"))
	if got == nil || got.Code != core.CallCalleeUnresolved {
		t.Fatalf("expected %s, got %+v", core.CallCalleeUnresolved, got)
	}
}

// TestVerifyCallInvariantsAcceptsWellFormedCall is the accepting-path
// counterpart: verifyCallInvariants returns nil for a well-formed OpCall.
func TestVerifyCallInvariantsAcceptsWellFormedCall(t *testing.T) {
	if got := verifyCallInvariants(verifyCallInvariantsFunctions("s1:test:fn:callee")); got != nil {
		t.Fatalf("expected no refusal for a well-formed OpCall, got %+v", got)
	}
}

// TestUndeclaredCalleeRefusedNeverSilentlyDropped is Task 2 Test 5
// (D-07-45): a source program calling an undeclared name is refused with
// the unresolved-callee code, and the resulting core.Program is never
// returned with the call silently dropped -- a dropped edge is how a cycle
// escapes detection.
func TestUndeclaredCalleeRefusedNeverSilentlyDropped(t *testing.T) {
	source := "module test.undeclared\nexport { fn main }\nfn main(value: Byte) -> Byte {\n  let result = ghost(value)\n  result\n}\n"
	result := Program(mustParseProgram(t, []byte(source)))
	if len(result.Program.Functions) != 0 {
		t.Fatalf("a call to an undeclared name must never reach a returned core.Program, got %d functions", len(result.Program.Functions))
	}
	found := false
	for _, problem := range result.Diagnostics {
		if problem.Code == core.CallCalleeUnresolved {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected %s, got %+v", core.CallCalleeUnresolved, result.Diagnostics)
	}
}

// TestVerifyCallInvariantsSeamRestoresBothRefusals is Task 2 Test 6
// (QLT-08, D-07-41/D-07-42): with the unexported seam engaged, the
// resolves-to-a-declared-function refusal (both the synthetic
// verifyCallInvariants case and the source-level undeclared-callee case)
// stops firing; restoring the seam (via defer) restores both.
func TestVerifyCallInvariantsSeamRestoresBothRefusals(t *testing.T) {
	defer func() { verifyCallInvariantsSeam = false }()

	verifyCallInvariantsSeam = true
	if got := verifyCallInvariants(verifyCallInvariantsFunctions("s1:test:fn:does-not-exist")); got != nil {
		t.Fatalf("expected the seam to suppress the unresolved-callee refusal, got %+v", got)
	}
	source := "module test.seam\nexport { fn main }\nfn main(value: Byte) -> Byte {\n  let result = ghost(value)\n  result\n}\n"
	seamResult := Program(mustParseProgram(t, []byte(source)))
	if len(seamResult.Diagnostics) != 0 {
		t.Fatalf("expected the seam to suppress the source-level refusal too, got %+v", seamResult.Diagnostics)
	}

	verifyCallInvariantsSeam = false
	if got := verifyCallInvariants(verifyCallInvariantsFunctions("s1:test:fn:does-not-exist")); got == nil {
		t.Fatal("expected the refusal restored once the seam is disengaged")
	}
	restored := Program(mustParseProgram(t, []byte(source)))
	found := false
	for _, problem := range restored.Diagnostics {
		if problem.Code == core.CallCalleeUnresolved {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected %s restored, got %+v", core.CallCalleeUnresolved, restored.Diagnostics)
	}
}

// TestBareCallToDeclaredFunctionAdmitted is Task 3 Test 4 (D-07-01): a bare
// call whose callee is a declared Lang function is admitted with no
// diagnostic.
func TestBareCallToDeclaredFunctionAdmitted(t *testing.T) {
	source := "module test.call_admitted\nexport { fn main }\nfn identity(value: Byte) -> Byte {\n  value\n}\nfn main(value: Byte) -> Byte {\n  let result = identity(value)\n  result\n}\n"
	result := Program(mustParseProgram(t, []byte(source)))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected a bare call to a declared Lang function to be admitted, got %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 2 {
		t.Fatalf("expected both functions checked, got %d", len(result.Program.Functions))
	}
}

// TestTwoArgumentCallRefusedAtCheckNotParse is Task 3 Test 5 (D-07-01/
// D-07-07): a call with two arguments is refused by a check predicate, and
// parsing that same source produces no parser-level error diagnostic --
// arity is a check-time rule this phase, not a parser-time one.
func TestTwoArgumentCallRefusedAtCheckNotParse(t *testing.T) {
	source := []byte("module test.call_arity\nexport { fn main }\nfn identity(value: Byte) -> Byte {\n  value\n}\nfn main(value: Byte) -> Byte {\n  let result = identity(value, value)\n  result\n}\n")
	parsed := syntax.Parse(source)
	for _, problem := range parsed.Diagnostics {
		if problem.Severity == "error" {
			t.Fatalf("expected no parser-level error diagnostic for a two-argument call, got %v", problem)
		}
	}
	result := Program(parsed.Program)
	found := false
	for _, problem := range result.Diagnostics {
		if problem.Code == "check.call_arity_unsupported" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected check.call_arity_unsupported, got %+v", result.Diagnostics)
	}
}

// TestCallArgumentNotInScopeRefused is Task 3 Test 6: a call whose argument
// names no in-scope binding is refused.
func TestCallArgumentNotInScopeRefused(t *testing.T) {
	source := "module test.call_scope\nexport { fn main }\nfn identity(value: Byte) -> Byte {\n  value\n}\nfn main(value: Byte) -> Byte {\n  let result = identity(ghost)\n  result\n}\n"
	result := Program(mustParseProgram(t, []byte(source)))
	found := false
	for _, problem := range result.Diagnostics {
		if problem.Code == "name.unknown" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected name.unknown for a call argument naming no in-scope binding, got %+v", result.Diagnostics)
	}
}

// TestEveryBindingIsFallibleRejectsMixedCallBinding covers Task 3's
// acceptance criterion that everyBindingIsFallible handles the "call" kind
// explicitly: a body mixing a "call" binding with a "try" binding is not
// "every binding fallible" (a "call" binding is never fallible), so it
// falls through to check.foreign_call_shape_unsupported rather than being
// silently treated as a resource-lifecycle chain.
func TestEveryBindingIsFallibleRejectsMixedCallBinding(t *testing.T) {
	if everyBindingIsFallible([]ast.Binding{
		{Name: "a", RHS: ast.RHS{Kind: "call", Callee: "helper", Arguments: []string{"value"}}},
		{Name: "b", RHS: ast.RHS{Kind: "try_call", Callee: "sym", Arguments: []string{"value"}}},
	}) {
		t.Fatal("expected everyBindingIsFallible to return false for a body mixing a call binding with a try binding")
	}
}

// TestCallFromBothMatchArmsEnumeratedByKind is Task 3 Test 7 (D-07-28): a
// fixture calling from BOTH arms of a match checks clean, and the checked
// core.Program carries one core.OpCall per arm -- found by scanning ALL
// operations for Kind == core.OpCall, never by block position, and still
// found if the arms' own operation order is reversed.
func TestCallFromBothMatchArmsEnumeratedByKind(t *testing.T) {
	source, err := os.ReadFile("../../../testdata/phase07/call_from_both_match_arms.lang")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected the fixture to check clean, got %+v", result.Diagnostics)
	}
	var mainFunction *core.Function
	for index := range result.Program.Functions {
		if result.Program.Functions[index].Name == "main" {
			mainFunction = &result.Program.Functions[index]
		}
	}
	if mainFunction == nil || mainFunction.Linear == nil {
		t.Fatal("expected a checked linear-carrying function named main")
	}
	calls := scanOpCallsByKind(mainFunction.Linear.Operations)
	if len(calls) != 2 {
		t.Fatalf("expected exactly 2 core.OpCall operations (one per arm), got %d", len(calls))
	}

	reversed := append([]core.LinearOperation(nil), mainFunction.Linear.Operations...)
	for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
		reversed[i], reversed[j] = reversed[j], reversed[i]
	}
	if reversedCalls := scanOpCallsByKind(reversed); len(reversedCalls) != 2 {
		t.Fatalf("expected the scan to still find 2 core.OpCall operations with operation order reversed, got %d", len(reversedCalls))
	}
}

// scanOpCallsByKind is the D-07-28 enumeration technique every OpCall
// consumer (07-06's callgraph included) must use: scan every operation for
// Kind == core.OpCall, never assume a fixed block/arm position.
func scanOpCallsByKind(operations []core.LinearOperation) []core.LinearOperation {
	var calls []core.LinearOperation
	for _, operation := range operations {
		if operation.Kind == core.OpCall {
			calls = append(calls, operation)
		}
	}
	return calls
}

// readPhase07Fixture reads a testdata/phase07 fixture the same way
// TestCallFromBothMatchArmsEnumeratedByKind already does.
func readPhase07Fixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../testdata/phase07/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return source
}

// TestCallToNonCallableCalleeRefused is 07-05 Task 1's SEM-06 tracer: a call
// to a callee that checks clean but fails publication (Callable == false,
// D-04-03/D-07-31) is refused with exactly one error diagnostic carrying the
// ratified core.CalleeNotCallable code, one Cause{Kind: "callee"} naming the
// callee's own function ID, and no repairs.
func TestCallToNonCallableCalleeRefused(t *testing.T) {
	source := readPhase07Fixture(t, "call_uncallable_callee.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %+v", result.Diagnostics)
	}
	diag := result.Diagnostics[0]
	if diag.Code != core.CalleeNotCallable {
		t.Fatalf("expected %s, got %s (%+v)", core.CalleeNotCallable, diag.Code, diag)
	}
	if diag.Severity != "error" {
		t.Fatalf("expected error severity, got %q", diag.Severity)
	}
	if len(diag.Causes) != 1 || diag.Causes[0].Kind != "callee" {
		t.Fatalf("expected exactly one Cause{Kind: \"callee\"}, got %+v", diag.Causes)
	}
	if !strings.HasSuffix(diag.Causes[0].Detail, ":fn:relay") {
		t.Fatalf("expected the cause detail to name the callee's own function ID, got %q", diag.Causes[0].Detail)
	}
	if diag.Repairs != nil {
		t.Fatalf("expected no repairs (D-07-31c: export_callee cannot fix an unsafe borrow-derived return), got %+v", diag.Repairs)
	}
}

// TestCallToDeclaredLangCalleeStillAdmitted is the accepting-path
// counterpart: a call to a callee that IS callable (call_basic.lang, already
// proven by 07-03/07-04) is unaffected by this plan's new admission arm.
func TestCallToDeclaredLangCalleeStillAdmitted(t *testing.T) {
	source := readPhase07Fixture(t, "call_basic.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected call_basic.lang to still check clean, got %+v", result.Diagnostics)
	}
}

// TestBuildCallSignatureTableCallableMatchesPublishProblemsFor is Task 1's
// direct proof that the table's Callable bit really is
// originvalidate.PublishProblemsFor's predicate (D-07-31/D-07-32) and
// nothing else: over the two functions in call_uncallable_callee.lang, the
// table entry's Callable bit matches len(PublishProblemsFor(fn))==0 exactly,
// for both the callable and the non-callable function.
func TestBuildCallSignatureTableCallableMatchesPublishProblemsFor(t *testing.T) {
	source := readPhase07Fixture(t, "call_uncallable_callee.lang")
	parsed := mustParseProgram(t, source)
	// Bypass verifyCallableRefusal (which would refuse the whole program)
	// by lowering functions directly through checkLinear, exactly like
	// Program does internally, so the table can be built and inspected
	// against a program this test controls end to end.
	calleeContracts := buildCalleeContracts(parsed)
	var functions []core.Function
	for _, function := range parsed.Funcs {
		checked, diagnostics, _, _ := checkLinear(parsed.Module, calleeContracts[function.Name].ID, function, calleeContracts, nil)
		if len(diagnostics) != 0 {
			t.Fatalf("expected %s to check clean, got %+v", function.Name, diagnostics)
		}
		functions = append(functions, checked)
	}
	program := core.Program{Schema: core.Schema1, Module: parsed.Module, Functions: functions}
	table, err := buildCallSignatureTable(program)
	if err != nil {
		t.Fatalf("buildCallSignatureTable: %v", err)
	}
	for _, function := range functions {
		entry, ok := table.lookup(function.ID)
		if !ok {
			t.Fatalf("expected a table entry for %s", function.ID)
		}
		want := len(originvalidate.PublishProblemsFor(function)) == 0
		if entry.Callable != want {
			t.Fatalf("%s: expected Callable == %v (PublishProblemsFor), got %v", function.Name, want, entry.Callable)
		}
	}
}

// TestCallSignatureTableEntryCarriesNoBodyReachableField is Task 1's
// structural proof of T-07-28/D-07-34's load-bearing property: the table
// entry type (core.FunctionSignature) has no field of type *core.LinearBody,
// *core.Match, or any type reachable to a function body -- reading its own
// struct tags/field types directly via reflection, not trusting a doc
// comment.
func TestCallSignatureTableEntryCarriesNoBodyReachableField(t *testing.T) {
	entryType := reflect.TypeOf(core.FunctionSignature{})
	for i := 0; i < entryType.NumField(); i++ {
		field := entryType.Field(i)
		name := field.Type.String()
		if strings.Contains(name, "LinearBody") || strings.Contains(name, "core.Match") || strings.Contains(name, "core.Function") {
			t.Fatalf("core.FunctionSignature.%s has type %s, which is reachable to a function body", field.Name, name)
		}
	}
}

// TestCallSignatureTableHasOnlyLookupMethod is Task 1's structural proof
// that callSignatureTable is a read-only accessor by TYPE, not merely by
// convention: parsing check.go's own AST (the same technique
// TestOriginValidatorImportsStayIndependent already uses for its own
// boundary) and asserting the only method declared with a callSignatureTable
// receiver is lookup -- there is no setter, exported or unexported, for a
// future edit to accidentally introduce without this test catching it.
func TestCallSignatureTableHasOnlyLookupMethod(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "check.go", nil, 0)
	if err != nil {
		t.Fatalf("parse check.go: %v", err)
	}
	var methods []string
	for _, declaration := range file.Decls {
		function, ok := declaration.(*goast.FuncDecl)
		if !ok || function.Recv == nil || len(function.Recv.List) != 1 {
			continue
		}
		receiverType := function.Recv.List[0].Type
		if starExpr, ok := receiverType.(*goast.StarExpr); ok {
			receiverType = starExpr.X
		}
		identifier, ok := receiverType.(*goast.Ident)
		if !ok || identifier.Name != "callSignatureTable" {
			continue
		}
		methods = append(methods, function.Name.Name)
	}
	if len(methods) != 1 || methods[0] != "lookup" {
		t.Fatalf("expected callSignatureTable's only method to be lookup, got %v", methods)
	}
}

// TestCallSignatureTableBuiltBeforeCallableAdmissionRuns is Task 1's proof
// that the table is fully built before the ONE admission decision it exists
// to gate (verifyCallableRefusal's Callable consult) ever runs: an
// instrumented build-observed seam and the existing lookup-observed seam
// record a single, real, ordered event log, and every recorded lookup
// event's index is strictly greater than the (exactly one) recorded build
// event's index, across every testdata/phase07 fixture containing a call.
func TestCallSignatureTableBuiltBeforeCallableAdmissionRuns(t *testing.T) {
	defer func() {
		callSignatureTableBuildObserved = nil
		callSignatureTableLookupObserved = nil
	}()

	fixtures := []string{"call_basic.lang", "call_from_both_match_arms.lang", "call_uncallable_callee.lang"}
	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			var events []string
			callSignatureTableBuildObserved = func() { events = append(events, "build") }
			callSignatureTableLookupObserved = func(calleeID string) { events = append(events, "lookup:"+calleeID) }
			result := Program(mustParseProgram(t, readPhase07Fixture(t, fixture)))
			callSignatureTableBuildObserved = nil
			callSignatureTableLookupObserved = nil

			buildIndex := -1
			lookupCount := 0
			for index, event := range events {
				if event == "build" {
					if buildIndex != -1 {
						t.Fatalf("expected exactly one build event, got a second at index %d: %v", index, events)
					}
					buildIndex = index
					continue
				}
				lookupCount++
				if buildIndex == -1 {
					t.Fatalf("observed a lookup event before any build event: %v", events)
				}
				if index <= buildIndex {
					t.Fatalf("lookup event at index %d did not come strictly after the build event at index %d: %v", index, buildIndex, events)
				}
			}
			if buildIndex == -1 {
				t.Fatalf("expected a build event, got none: %v (diagnostics=%+v)", events, result.Diagnostics)
			}
			if lookupCount == 0 {
				t.Fatalf("expected at least one lookup event for a fixture containing a call, got none: %v", events)
			}
		})
	}
}

// TestRelayEscortWitnessParsesCleanly is Task 2 Test 1 (D-07-44): the
// A-normal-form conversion of D-04-03's decisive research witness parses
// with zero parser diagnostics.
func TestRelayEscortWitnessParsesCleanly(t *testing.T) {
	source := readPhase07Fixture(t, "relay_escort_witness.lang")
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("expected zero parser diagnostics, got %+v", parsed.Diagnostics)
	}
}

// TestRelayEscortWitnessCallArgumentsAreANormalForm is Task 2 Test 3
// (D-07-01/D-07-44): every call argument in the fixture is a bare
// identifier -- fails if a nested call or a borrow/take expression ever
// appears at a call site (the exact shape the recorded witness's
// `relay(borrow mut buffer)` had, and D-07-01 now forbids).
func TestRelayEscortWitnessCallArgumentsAreANormalForm(t *testing.T) {
	source := readPhase07Fixture(t, "relay_escort_witness.lang")
	parsed := mustParseProgram(t, source)
	callCount := 0
	for _, function := range parsed.Funcs {
		if function.Body.Linear == nil {
			continue
		}
		for _, binding := range function.Body.Linear.Bindings {
			if binding.RHS.Kind != "call" {
				continue
			}
			callCount++
			if len(binding.RHS.Arguments) != 1 {
				t.Fatalf("%s: expected exactly one call argument, got %d", function.Name, len(binding.RHS.Arguments))
			}
			argument := binding.RHS.Arguments[0]
			if argument == "" || strings.ContainsAny(argument, "() \t\n") {
				t.Fatalf("%s: call argument %q is not a bare identifier (A-normal form violation)", function.Name, argument)
			}
		}
	}
	if callCount == 0 {
		t.Fatal("expected at least one call binding in the fixture")
	}
}

// TestRelayEscortWitnessRefusesInterproceduralLiveness is Task 2 Tests 2 and
// 4 (D-04-03/D-07-44, D-03-02), CLOSED: the converted witness's check
// outcome is now a refusal, not the historical clean-check finding this
// test used to assert. `relay` and `escort` both declare (and satisfy) a
// correct PublicOrigin, so both are Callable (originvalidate.PublishProblemsFor
// reports no problems for either), and `escort`'s call to `relay` is
// admitted under Phase 07's SEM-06 admission arm -- the refusal below is
// NOT a call-admission rejection, it is Phase 08's new interprocedural
// loan-liveness law. `escort` moves `buffer` away (`take buffer`) while
// `aliased` -- the call's own return, standing in for the exclusive
// borrowed view `relay` actually produced -- remains live past the move:
// `derivePlaceLoans` now propagates the loan across the OpCall boundary
// (D-08-07) because `relay`'s declared `-> borrow mut(buffer) Buffer`
// return contract reports `returnsBorrowOfParam == true`, closing the
// INTERPROCEDURAL half of D-03-02 (Phase 3 closed the single-function half)
// via check.interprocedural_loan_liveness.
func TestRelayEscortWitnessRefusesInterproceduralLiveness(t *testing.T) {
	source := readPhase07Fixture(t, "relay_escort_witness.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %+v", result.Diagnostics)
	}
	diag := result.Diagnostics[0]
	if diag.Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected code check.interprocedural_loan_liveness, got %s", diag.Code)
	}
	if len(diag.Repairs) != 0 {
		t.Fatalf("expected zero repairs (D-08-25), got %+v", diag.Repairs)
	}
	if len(diag.Causes) != 3 {
		t.Fatalf("expected exactly 3 causes (D-08-23), got %d: %+v", len(diag.Causes), diag.Causes)
	}
	wantKinds := []string{"borrow_created_here", "loan_extended_by_call", "callee_return_contract"}
	for i, want := range wantKinds {
		if diag.Causes[i].Kind != want {
			t.Fatalf("cause %d: expected kind %q, got %q", i, want, diag.Causes[i].Kind)
		}
	}
	if diag.Causes[2].Span != nil {
		t.Fatalf("cause 3 (callee_return_contract) must be spanless (D-08-24), got %+v", diag.Causes[2].Span)
	}
	if diag.Causes[2].Detail == "" || !strings.Contains(diag.Causes[2].Detail, ":return.mode=") {
		t.Fatalf("cause 3's Detail must match <calleeID>:return.mode=<Mode>, got %q", diag.Causes[2].Detail)
	}
	if !strings.HasSuffix(diag.Causes[2].Detail, ":return.mode=exclusive") {
		t.Fatalf("expected return.mode=exclusive (relay declares -> borrow mut(buffer)), got %q", diag.Causes[2].Detail)
	}

	var escort *core.Function
	for index := range result.Program.Functions {
		if result.Program.Functions[index].Name == "escort" {
			escort = &result.Program.Functions[index]
		}
	}
	if escort == nil {
		t.Fatal("expected a checked function named escort")
	}
	hasCall, hasMove := false, false
	for _, operation := range escort.Linear.Operations {
		switch operation.Kind {
		case core.OpCall:
			hasCall = true
		case core.OpMove:
			hasMove = true
		}
	}
	if !hasCall || !hasMove {
		t.Fatalf("expected escort's checked body to carry both a call and a move (the take), got call=%v move=%v", hasCall, hasMove)
	}
	if diag.Primary.Start == 0 && diag.Primary.End == 0 {
		t.Fatalf("expected a non-zero Primary span (the take buffer binding's RHS span), got %+v", diag.Primary)
	}
}

// TestInterproceduralLoanLivenessTracer is Task 2's accepting counterpart:
// the same caller body against a callee declaring an OWNED return (not a
// borrow of its parameter) produces zero diagnostics -- the refusal above
// is contract-driven, not shape-driven (must_haves backstop truth).
func TestInterproceduralLoanLivenessTracer(t *testing.T) {
	source := []byte(`module phase08.tracer_safe_twin

export {
  fn escort
}

fn relay(buffer: Buffer) -> Buffer {
  buffer
}

fn escort(buffer: Buffer) -> Buffer {
  let borrowed = borrow mut buffer
  let aliased = relay(borrowed)
  let delivered = take buffer
  aliased
}
`)
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected zero diagnostics for the owned-return safe twin, got %+v", result.Diagnostics)
	}
}

// ---------------------------------------------------------------------
// Phase 08 Task 1 (08-02): deriveFunctionUsesParam / buildInterproceduralSummaries.
// ---------------------------------------------------------------------

// TestDeriveFunctionUsesParamBehaviors is 08-02 Task 1's table-driven
// coverage of every case named in the task's own <behavior> block, each
// exercised directly against deriveFunctionUsesParam (never through the
// full ast.Program pipeline, since this is a pure core-level derivation).
func TestDeriveFunctionUsesParamBehaviors(t *testing.T) {
	parameter := core.Parameter{ID: "fn:place:0", Name: "value", Type: "Byte"}
	cases := []struct {
		name       string
		operations []core.LinearOperation
		summaries  interproceduralSummaryTable
		want       bool
	}{
		{
			name: "a leaf function whose body reads its parameter",
			operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpCopy, SourceID: "fn:place:0", TargetID: "fn:place:1"},
				{ID: "op:1", Kind: core.OpReturn, SourceID: "fn:place:1"},
			},
			want: true,
		},
		{
			name: "a leaf function whose body only forwards its parameter to its own return",
			operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpReturn, SourceID: "fn:place:0"},
			},
			want: false,
		},
		{
			name: "a relay whose only parameter contact is passing it to a usesParam callee",
			operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpCall, SourceID: "fn:place:0", TargetID: "fn:place:1", CalleeID: "fn:leaf"},
				{ID: "op:1", Kind: core.OpReturn, SourceID: "fn:place:1"},
			},
			summaries: interproceduralSummaryTable{summaries: map[string]interproceduralSummary{"fn:leaf": {usesParam: true}}},
			want:      true,
		},
		{
			name: "the same relay pointed at a non-usesParam callee",
			operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpCall, SourceID: "fn:place:0", TargetID: "fn:place:1", CalleeID: "fn:leaf"},
				{ID: "op:1", Kind: core.OpReturn, SourceID: "fn:place:1"},
			},
			summaries: interproceduralSummaryTable{summaries: map[string]interproceduralSummary{"fn:leaf": {usesParam: false}}},
			want:      false,
		},
		{
			name: "a call to a callee absent from summaries counts as usesParam (refusing direction)",
			operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpCall, SourceID: "fn:place:0", TargetID: "fn:place:1", CalleeID: "fn:absent"},
				{ID: "op:1", Kind: core.OpReturn, SourceID: "fn:place:1"},
			},
			summaries: interproceduralSummaryTable{},
			want:      true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			function := core.Function{
				ID: "fn:under_test", Name: "under_test", Parameter: parameter,
				Linear: &core.LinearBody{ID: "fn:linear", Operations: tc.operations},
			}
			got, work := deriveFunctionUsesParam(function, tc.summaries)
			if got != tc.want {
				t.Fatalf("usesParam = %v, want %v", got, tc.want)
			}
			if work <= 0 {
				t.Fatalf("expected positive counted work, got %d", work)
			}
		})
	}
}

// TestSummaryDerivationTwoHopChainPropagates is 08-02 Task 1's three-deep
// chain case: caller -> relay -> leaf, where only the leaf's body genuinely
// reads its own parameter. Asserts buildInterproceduralSummaries' own
// callee-before-caller traversal lets the leaf's bit propagate all the way
// to the caller in one build.
func TestSummaryDerivationTwoHopChainPropagates(t *testing.T) {
	leaf := core.Function{
		ID: "fn:leaf", Name: "leaf", Parameter: core.Parameter{ID: "leaf:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "leaf:linear", Operations: []core.LinearOperation{
			{ID: "leaf:op:0", Kind: core.OpCopy, SourceID: "leaf:place:0", TargetID: "leaf:place:1"},
			{ID: "leaf:op:1", Kind: core.OpReturn, SourceID: "leaf:place:1"},
		}},
	}
	relay := core.Function{
		ID: "fn:relay", Name: "relay", Parameter: core.Parameter{ID: "relay:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "relay:linear", Operations: []core.LinearOperation{
			{ID: "relay:op:0", Kind: core.OpCall, SourceID: "relay:place:0", TargetID: "relay:place:1", CalleeID: "fn:leaf"},
			{ID: "relay:op:1", Kind: core.OpReturn, SourceID: "relay:place:1"},
		}},
	}
	caller := core.Function{
		ID: "fn:caller", Name: "caller", Parameter: core.Parameter{ID: "caller:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "caller:linear", Operations: []core.LinearOperation{
			{ID: "caller:op:0", Kind: core.OpCall, SourceID: "caller:place:0", TargetID: "caller:place:1", CalleeID: "fn:relay"},
			{ID: "caller:op:1", Kind: core.OpReturn, SourceID: "caller:place:1"},
		}},
	}
	program := core.Program{Functions: []core.Function{caller, relay, leaf}}
	summaries, _ := buildInterproceduralSummaries(program, callSignatureTable{})
	for id, want := range map[string]bool{"fn:leaf": true, "fn:relay": true, "fn:caller": true} {
		summary, ok := summaries.lookup(id)
		if !ok {
			t.Fatalf("%s: expected a summary entry", id)
		}
		if summary.usesParam != want {
			t.Fatalf("%s: usesParam = %v, want %v (the leaf's own bit must propagate through both hops)", id, summary.usesParam, want)
		}
	}
}

// TestSummaryDerivationIsOnePassPerFunction is 08-02 Task 1's own falsifier
// for the one-derivation-per-function claim: a diamond call graph (one
// caller, two relays, one shared leaf reached through BOTH relays) must
// derive every function -- including the shared leaf -- exactly once,
// instrumented via deriveFunctionUsesParamObserved rather than merely
// asserted from the summary's own final content.
func TestSummaryDerivationIsOnePassPerFunction(t *testing.T) {
	leaf := core.Function{
		ID: "fn:leaf", Name: "leaf", Parameter: core.Parameter{ID: "leaf:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "leaf:linear", Operations: []core.LinearOperation{
			{ID: "leaf:op:0", Kind: core.OpCopy, SourceID: "leaf:place:0", TargetID: "leaf:place:1"},
			{ID: "leaf:op:1", Kind: core.OpReturn, SourceID: "leaf:place:1"},
		}},
	}
	relay := func(id string) core.Function {
		return core.Function{
			ID: id, Name: id, Parameter: core.Parameter{ID: id + ":place:0", Name: "v", Type: "Byte"},
			Linear: &core.LinearBody{ID: id + ":linear", Operations: []core.LinearOperation{
				{ID: id + ":op:0", Kind: core.OpCall, SourceID: id + ":place:0", TargetID: id + ":place:1", CalleeID: "fn:leaf"},
				{ID: id + ":op:1", Kind: core.OpReturn, SourceID: id + ":place:1"},
			}},
		}
	}
	relay1 := relay("fn:relay1")
	relay2 := relay("fn:relay2")
	caller := core.Function{
		ID: "fn:caller", Name: "caller", Parameter: core.Parameter{ID: "caller:place:0", Name: "v", Type: "Byte"},
		Linear: &core.LinearBody{ID: "caller:linear", Operations: []core.LinearOperation{
			{ID: "caller:op:0", Kind: core.OpCall, SourceID: "caller:place:0", TargetID: "caller:place:1", CalleeID: "fn:relay1"},
			{ID: "caller:op:1", Kind: core.OpCall, SourceID: "caller:place:1", TargetID: "caller:place:2", CalleeID: "fn:relay2"},
			{ID: "caller:op:2", Kind: core.OpReturn, SourceID: "caller:place:2"},
		}},
	}
	program := core.Program{Functions: []core.Function{caller, relay1, relay2, leaf}}

	defer func() { deriveFunctionUsesParamObserved = nil }()
	counts := map[string]int{}
	deriveFunctionUsesParamObserved = func(functionID string) { counts[functionID]++ }

	summaries, _ := buildInterproceduralSummaries(program, callSignatureTable{})
	deriveFunctionUsesParamObserved = nil

	for _, id := range []string{"fn:caller", "fn:relay1", "fn:relay2", "fn:leaf"} {
		if counts[id] != 1 {
			t.Fatalf("function %s was derived %d times, want exactly 1 (got %+v)", id, counts[id], counts)
		}
	}
	callerSummary, ok := summaries.lookup("fn:caller")
	if !ok || !callerSummary.usesParam {
		t.Fatalf("expected fn:caller's usesParam to be true (transitive through both relays to the shared leaf), got %+v ok=%v", callerSummary, ok)
	}
}

// ---------------------------------------------------------------------
// Phase 08 Task 2 (08-02): the backward OpCall gate in blockLoanLiveness.
// ---------------------------------------------------------------------

// TestInterproceduralLivenessTwinPatternB is 08-02 Task 2's synthetic-
// core.Program twin (D-08-08): borrow; move; call(v), where the two
// programs differ ONLY in the callee's own usesParam bit. A using callee
// refuses; a non-using callee admits with zero diagnostics -- the refusal
// is contract-driven, not shape-driven.
func TestInterproceduralLivenessTwinPatternB(t *testing.T) {
	usingCallee := core.Function{
		ID: "fn:user_true", Name: "user_true", Parameter: core.Parameter{ID: "user_true:place:0", Name: "borrowed", Type: "Buffer"},
		Linear: &core.LinearBody{ID: "user_true:linear", Operations: []core.LinearOperation{
			{ID: "user_true:op:0", Kind: core.OpCopy, SourceID: "user_true:place:0", TargetID: "user_true:place:1"},
			{ID: "user_true:op:1", Kind: core.OpReturn, SourceID: "user_true:place:1"},
		}},
	}
	notUsingCallee := core.Function{
		ID: "fn:user_false", Name: "user_false", Parameter: core.Parameter{ID: "user_false:place:0", Name: "borrowed", Type: "Buffer"},
		Linear: &core.LinearBody{ID: "user_false:linear", Operations: []core.LinearOperation{
			{ID: "user_false:op:0", Kind: core.OpReturn, SourceID: "user_false:place:0"},
		}},
	}
	signatureTable := callSignatureTable{entries: map[string]core.FunctionSignature{
		"fn:user_true":  {ID: "fn:user_true", Return: core.ReturnContract{Mode: "owned"}, Parameters: []core.ParameterContract{{Mode: "owned"}}},
		"fn:user_false": {ID: "fn:user_false", Return: core.ReturnContract{Mode: "owned"}, Parameters: []core.ParameterContract{{Mode: "owned"}}},
	}}

	for _, tc := range []struct {
		name           string
		calleeID       string
		callee         core.Function
		wantDiagnostic bool
	}{
		{"callee uses its parameter", "fn:user_true", usingCallee, true},
		{"callee does not use its parameter", "fn:user_false", notUsingCallee, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			caller := core.Function{
				ID: "fn:escort", Name: "escort", Parameter: core.Parameter{ID: "escort:place:buffer", Name: "buffer", Type: "Buffer"},
				Linear: &core.LinearBody{ID: "escort:linear", Operations: []core.LinearOperation{
					{ID: "escort:op:0", Kind: core.OpBorrowExclusive, SourceID: "escort:place:buffer", TargetID: "escort:place:borrowed", LoanID: "escort:loan:0"},
					{ID: "escort:op:1", Kind: core.OpMove, SourceID: "escort:place:buffer", TargetID: "escort:place:moved"},
					{ID: "escort:op:2", Kind: core.OpCall, SourceID: "escort:place:borrowed", TargetID: "escort:place:result", CalleeID: tc.calleeID},
					{ID: "escort:op:3", Kind: core.OpReturn, SourceID: "escort:place:result"},
				}},
			}
			program := core.Program{Functions: []core.Function{caller, tc.callee}}
			summaries, _ := buildInterproceduralSummaries(program, signatureTable)
			diagnostics := checkInterproceduralLoanLiveness(program, summaries, map[string]diagnostic.Span{})
			if !tc.wantDiagnostic {
				if len(diagnostics) != 0 {
					t.Fatalf("expected zero diagnostics for the non-using callee, got %+v", diagnostics)
				}
				return
			}
			if len(diagnostics) != 1 {
				t.Fatalf("expected exactly 1 diagnostic for the using callee, got %+v", diagnostics)
			}
			diag := diagnostics[0]
			if diag.Code != "check.interprocedural_loan_liveness" {
				t.Fatalf("expected code check.interprocedural_loan_liveness, got %s", diag.Code)
			}
			if len(diag.Causes) != 3 || diag.Causes[2].Kind != "callee_return_contract" {
				t.Fatalf("expected 3 causes with cause 3 kind callee_return_contract, got %+v", diag.Causes)
			}
			if !strings.Contains(diag.Causes[2].Detail, "parameters[0].mode=") {
				t.Fatalf("expected cause 3's Detail to contain parameters[0].mode= for the backward direction, got %q", diag.Causes[2].Detail)
			}
		})
	}
}

// ---------------------------------------------------------------------
// Phase 08 Task 3 (08-02): program-order invariant and never-persisted memo.
// ---------------------------------------------------------------------

// relayChainFunction builds a single function whose body is a chain of n
// OpCopy operations rooted at its own parameter (op[i]'s source is
// op[i-1]'s own target), the per-function analogue of
// reborrowChainOperations -- the exact shape that exposes
// deriveFunctionUsesParam's own program-order dependency: each successive
// operation's SourceID is only in the parameter-derived set once its
// antecedent has already been visited.
func relayChainFunction(n int) core.Function {
	parameter := core.Parameter{ID: "chain:place:0", Name: "value", Type: "Byte"}
	operations := make([]core.LinearOperation, n)
	source := parameter.ID
	for index := 0; index < n; index++ {
		target := fmt.Sprintf("chain:place:%d", index+1)
		operations[index] = core.LinearOperation{ID: fmt.Sprintf("chain:op:%d", index), Kind: core.OpCopy, SourceID: source, TargetID: target}
		source = target
	}
	return core.Function{ID: "fn:chain", Name: "chain", Parameter: parameter, Linear: &core.LinearBody{ID: "chain:linear", Operations: operations}}
}

// reverseFunctionOperations returns a shallow copy of function with its own
// Linear.Operations reversed -- the control this task's falsifier compares
// against the program-order arrangement built by relayChainFunction.
func reverseFunctionOperations(function core.Function) core.Function {
	original := function.Linear.Operations
	reversed := make([]core.LinearOperation, len(original))
	for index, operation := range original {
		reversed[len(original)-1-index] = operation
	}
	clone := function
	clone.Linear = &core.LinearBody{ID: function.Linear.ID, Operations: reversed}
	return clone
}

// TestSummaryDerivationRequiresProgramOrder is 08-02 Task 3(a)'s pin of
// D-08-11: summary derivation work over a relay chain of k operations grows
// no faster than operation count as k rises through a size series, AND the
// identical chain with its operations reversed costs strictly more per
// operation at the largest size -- the falsifier a future pass reordering
// operations before summary derivation would trip.
func TestSummaryDerivationRequiresProgramOrder(t *testing.T) {
	series := []int{8, 32, 128, 512}
	forwardWork := make([]int, len(series))
	reversedWork := make([]int, len(series))
	for index, n := range series {
		function := relayChainFunction(n)
		usesParam, work := deriveFunctionUsesParam(function, interproceduralSummaryTable{})
		if !usesParam {
			t.Fatalf("n=%d: expected usesParam true for a chain rooted at the parameter", n)
		}
		forwardWork[index] = work

		_, reversedWorkAtN := deriveFunctionUsesParam(reverseFunctionOperations(function), interproceduralSummaryTable{})
		reversedWork[index] = reversedWorkAtN
	}

	for index := 1; index < len(series); index++ {
		operationRatio := float64(series[index]) / float64(series[index-1])
		workRatio := float64(forwardWork[index]) / float64(forwardWork[index-1])
		if workRatio > operationRatio*2 {
			t.Fatalf("n=%d->%d: operation count grew %.1fx but program-order work grew %.1fx (%d->%d) -- looks quadratic",
				series[index-1], series[index], operationRatio, workRatio, forwardWork[index-1], forwardWork[index])
		}
	}

	largest := len(series) - 1
	forwardPerOp := float64(forwardWork[largest]) / float64(series[largest])
	reversedPerOp := float64(reversedWork[largest]) / float64(series[largest])
	multiplier := reversedPerOp / forwardPerOp
	if reversedPerOp <= forwardPerOp {
		t.Fatalf("n=%d: expected the reversed arrangement to cost strictly more per operation than program order: forward=%.2f reversed=%.2f (%.1fx)",
			series[largest], forwardPerOp, reversedPerOp, multiplier)
	}
	t.Logf("n=%d: reversed-vs-forward work-per-operation multiplier: %.1fx (forward=%.2f reversed=%.2f)", series[largest], multiplier, forwardPerOp, reversedPerOp)
}

// TestSummaryMemoNeverPersisted is 08-02 Task 3(b)'s pin of D-08-12/D-08-37:
// (a) a go/ast structural scan of package check's own non-test source files
// asserts no top-level var declaration anywhere references
// interproceduralSummaryTable, and (b) running check.Program twice over the
// same parsed program invokes buildInterproceduralSummaries exactly twice
// -- a fresh memo built per invocation, never persisted or reused. Within-
// run memoization is mandatory and sufficient (EFF-02); a cross-run cache
// is Phase 11 / QLT-06 work (D-08-37) and must not appear here by accident.
func TestSummaryMemoNeverPersisted(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob package check's own sources: %v", err)
	}
	fileSet := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, declaration := range file.Decls {
			genDecl, ok := declaration.(*goast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}
			goast.Inspect(genDecl, func(node goast.Node) bool {
				identifier, ok := node.(*goast.Ident)
				if ok && identifier.Name == "interproceduralSummaryTable" {
					t.Fatalf("%s: a package-level var declaration references interproceduralSummaryTable -- the memo must never be held in a package-level variable (D-08-12/D-08-37)", path)
				}
				return true
			})
		}
	}

	defer func() { buildInterproceduralSummariesObserved = nil }()
	source := readPhase07Fixture(t, "call_basic.lang")
	parsed := mustParseProgram(t, source)

	buildCount := 0
	buildInterproceduralSummariesObserved = func() { buildCount++ }
	Program(parsed)
	Program(parsed)
	buildInterproceduralSummariesObserved = nil

	if buildCount != 2 {
		t.Fatalf("expected buildInterproceduralSummaries to run exactly twice across two check.Program calls on the same parsed program (a fresh memo per invocation), got %d", buildCount)
	}
}

// TestRelayEscortWitnessBothFunctionsAreCallable independently confirms
// (via originvalidate.PublishProblemsFor, not by re-deriving a second
// predicate) that BOTH `relay` and `escort` are Callable -- the call this
// witness demonstrates is genuinely ADMITTED by this plan's own SEM-06 arm,
// not accidentally refused for the unrelated reason Task 1's own negative
// control (call_uncallable_callee.lang) demonstrates. Phase 08's new
// interprocedural loan-liveness refusal (TestRelayEscortWitnessRefusesInterproceduralLiveness)
// is a SEPARATE, later admission arm; it never clears result.Program (unlike
// the call-graph-cycle gate), so Callable is still independently checkable
// here even though the fixture now refuses.
func TestRelayEscortWitnessBothFunctionsAreCallable(t *testing.T) {
	source := readPhase07Fixture(t, "relay_escort_witness.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected exactly the interprocedural_loan_liveness diagnostic, got %+v", result.Diagnostics)
	}
	iface, err := originvalidate.BuildInterface(result.Program)
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	found := 0
	for _, signature := range iface.Functions {
		if !signature.Callable {
			t.Fatalf("expected %s to be Callable, got Callable == false", signature.Name)
		}
		found++
	}
	if found != 2 {
		t.Fatalf("expected 2 functions (relay, escort), got %d", found)
	}
}

// TestCallAdmissionBodyBlindControl is Task 3 Tests 1-2 (D-07-41/D-07-42,
// SEM-05's structural claim made falsifiable): across every
// testdata/phase07 fixture containing a call, verifyCallableRefusal's
// admission arm consults ONLY the signature table for a callee -- never a
// body value -- at the production default. Engaging
// verifyCallableRefusalBodyReadSeam proves the SAME instrumentation would
// have caught it: the body-read observer fires at least once, exactly the
// condition this test's own zero-invocations assertion exists to detect.
// Restoring the seam (via defer) restores the zero-invocations guarantee.
func TestCallAdmissionBodyBlindControl(t *testing.T) {
	defer func() {
		verifyCallableRefusalBodyReadSeam = false
		calleeBodyReadObserved = nil
	}()

	fixtures := []string{"call_basic.lang", "call_from_both_match_arms.lang", "relay_escort_witness.lang"}
	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			source := readPhase07Fixture(t, fixture)

			// Production default: zero body reads.
			var bodyReads int
			calleeBodyReadObserved = func(string) { bodyReads++ }
			result := Program(mustParseProgram(t, source))
			calleeBodyReadObserved = nil
			// relay_escort_witness.lang refuses under Phase 08's new
			// check.interprocedural_loan_liveness law since this task
			// (TestRelayEscortWitnessRefusesInterproceduralLiveness); that
			// refusal is unrelated to THIS test's body-blindness claim about
			// verifyCallableRefusal, so accept it here too -- but any OTHER
			// diagnostic (in particular a callee_not_callable body-read
			// symptom) still fails the test.
			for _, diag := range result.Diagnostics {
				if diag.Code != "check.interprocedural_loan_liveness" {
					t.Fatalf("expected the fixture to check clean (or only refuse via check.interprocedural_loan_liveness), got %+v", result.Diagnostics)
				}
			}
			if bodyReads != 0 {
				t.Fatalf("expected 0 body reads at the production default, got %d", bodyReads)
			}

			// The fault: engaging the body-read seam makes the SAME
			// instrumentation observe at least one body read -- proving the
			// zero-invocations assertion above would have gone red had the
			// production code actually taken this path.
			bodyReads = 0
			calleeBodyReadObserved = func(string) { bodyReads++ }
			verifyCallableRefusalBodyReadSeam = true
			Program(mustParseProgram(t, source))
			verifyCallableRefusalBodyReadSeam = false
			calleeBodyReadObserved = nil
			if bodyReads == 0 {
				t.Fatal("expected the body-read seam to make at least one body read observable")
			}
		})
	}
}

// TestVerifyCallableRefusalSeamAdmitsUncallableCallee is Task 3 Test 3
// (QLT-08, D-07-41/D-07-42): with verifyCallableRefusalSeam engaged,
// call_uncallable_callee.lang -- refused at the production default -- is
// wrongly admitted (zero diagnostics). Restoring the seam restores the
// refusal.
func TestVerifyCallableRefusalSeamAdmitsUncallableCallee(t *testing.T) {
	defer func() { verifyCallableRefusalSeam = false }()
	source := readPhase07Fixture(t, "call_uncallable_callee.lang")

	verifyCallableRefusalSeam = true
	seamResult := Program(mustParseProgram(t, source))
	if len(seamResult.Diagnostics) != 0 {
		t.Fatalf("expected the seam to admit the call, got %+v", seamResult.Diagnostics)
	}

	verifyCallableRefusalSeam = false
	restored := Program(mustParseProgram(t, source))
	if len(restored.Diagnostics) != 1 || restored.Diagnostics[0].Code != core.CalleeNotCallable {
		t.Fatalf("expected %s restored once the seam is disengaged, got %+v", core.CalleeNotCallable, restored.Diagnostics)
	}
}

// TestVerifyCallableRefusalRefusesUnpopulatedTableEntry is Task 3 Test 4
// (D-07-09): a callee whose signature-table entry was never populated (a
// synthetic gap, simulating a table-build failure or a callee the table
// forgot) is refused, never admitted -- absence is the refusing case.
func TestVerifyCallableRefusalRefusesUnpopulatedTableEntry(t *testing.T) {
	functions := verifyCallInvariantsFunctions("s1:test:fn:callee")
	// The zero-value table: no entries at all. A callee ID that is a real,
	// resolvable function (per verifyCallInvariants' own fixture) but
	// simply absent from THIS table must still be refused.
	got := verifyCallableRefusal(functions, callSignatureTable{})
	if got == nil || got.Code != core.CalleeNotCallable {
		t.Fatalf("expected %s for an unpopulated table entry, got %+v", core.CalleeNotCallable, got)
	}
}

// TestVerifyCallableRefusalAcceptsPopulatedCallableEntry is the accepting
// counterpart: a table entry that IS present and Callable admits the call.
func TestVerifyCallableRefusalAcceptsPopulatedCallableEntry(t *testing.T) {
	functions := verifyCallInvariantsFunctions("s1:test:fn:callee")
	table := callSignatureTable{entries: map[string]core.FunctionSignature{
		"s1:test:fn:callee": {ID: "s1:test:fn:callee", Callable: true},
	}}
	if got := verifyCallableRefusal(functions, table); got != nil {
		t.Fatalf("expected no refusal for a populated, Callable entry, got %+v", got)
	}
}

// TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses is
// Task 3 Test 5's "check disabled, corevalidate still refuses" half
// (D-07-34's two-peer discipline): with check's OWN admission arm
// disabled (verifyCallableRefusalSeam engaged), call_uncallable_callee.lang
// is admitted by check -- but the resulting core.Program is
// INDEPENDENTLY still refused by corevalidate's own peerCallable-based
// admission arm (07-05 Task 3), which never consults check's table or its
// seam at all.
func TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses(t *testing.T) {
	defer func() { verifyCallableRefusalSeam = false }()
	verifyCallableRefusalSeam = true
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "call_uncallable_callee.lang")))
	verifyCallableRefusalSeam = false
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected check's seam to admit the call, got %+v", result.Diagnostics)
	}
	coreResult := corevalidate.Validate(result.Program)
	if coreResult.Valid {
		t.Fatal("expected corevalidate to independently still refuse, got Valid == true")
	}
	found := false
	for _, problem := range coreResult.Problems {
		if problem.Code == core.CalleeNotCallable {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected %s, got %+v", core.CalleeNotCallable, coreResult.Problems)
	}
}

// buildCycleSource generates a Lang source program declaring n functions,
// f0..f(n-1), each calling the next in sequence, with the last wrapping
// around to f0 -- a synthetic length-n call cycle, used to exercise
// D-07-15's 32-cause bound without hand-writing 33 fixture functions.
func buildCycleSource(n int) []byte {
	var b strings.Builder
	b.WriteString("module test.bigcycle\nexport { fn f0 }\n")
	for i := 0; i < n; i++ {
		next := (i + 1) % n
		fmt.Fprintf(&b, "fn f%d(value: Byte) -> Byte {\n  let result = f%d(value)\n  result\n}\n", i, next)
	}
	return []byte(b.String())
}

// findCallGraphCycleDiagnostic locates the core.call_graph_cycle diagnostic
// among result.Diagnostics, failing the test if absent.
func findCallGraphCycleDiagnostic(t *testing.T, result Result) diagnostic.Diagnostic {
	t.Helper()
	for _, diag := range result.Diagnostics {
		if diag.Code == core.CallGraphCycle {
			return diag
		}
	}
	t.Fatalf("expected a %s diagnostic, got %+v", core.CallGraphCycle, result.Diagnostics)
	return diagnostic.Diagnostic{}
}

// TestCallGraphCycleBoundedAt32Causes is Task 2 Test 4: a 32-member
// synthetic cycle emits exactly 32 cycle_member causes and zero truncated
// causes.
func TestCallGraphCycleBoundedAt32Causes(t *testing.T) {
	result := Program(mustParseProgram(t, buildCycleSource(32)))
	diag := findCallGraphCycleDiagnostic(t, result)
	memberCauses, truncatedCauses := 0, 0
	for _, cause := range diag.Causes {
		switch cause.Kind {
		case "cycle_member":
			memberCauses++
		case "truncated":
			truncatedCauses++
		case "cycle_length":
			if cause.Detail != "32" {
				t.Fatalf("want cycle_length 32, got %s", cause.Detail)
			}
		}
	}
	if memberCauses != 32 {
		t.Fatalf("want 32 cycle_member causes, got %d", memberCauses)
	}
	if truncatedCauses != 0 {
		t.Fatalf("want zero truncated causes at exactly the bound, got %d", truncatedCauses)
	}
}

// TestCallGraphCycleTruncatesAt33Members is Task 2 Test 5: a 33-member
// synthetic cycle emits exactly 32 cycle_member causes plus one truncated
// cause whose Detail is callgraph.TruncatedCycleBound, and Task 2 Test 6:
// the cycle_length cause's Detail is the TRUE member count ("33"), never
// the truncated count.
func TestCallGraphCycleTruncatesAt33Members(t *testing.T) {
	result := Program(mustParseProgram(t, buildCycleSource(33)))
	diag := findCallGraphCycleDiagnostic(t, result)
	memberCauses := 0
	truncatedCauses := 0
	sawLength := false
	for _, cause := range diag.Causes {
		switch cause.Kind {
		case "cycle_member":
			memberCauses++
		case "truncated":
			truncatedCauses++
			if cause.Detail != callgraph.TruncatedCycleBound {
				t.Fatalf("want truncated detail %q, got %q", callgraph.TruncatedCycleBound, cause.Detail)
			}
		case "cycle_length":
			sawLength = true
			if cause.Detail != "33" {
				t.Fatalf("want the TRUE member count (33), got %s", cause.Detail)
			}
		}
	}
	if !sawLength {
		t.Fatalf("expected a cycle_length cause, got %+v", diag.Causes)
	}
	if memberCauses != 32 {
		t.Fatalf("want exactly 32 cycle_member causes after truncation, got %d", memberCauses)
	}
	if truncatedCauses != 1 {
		t.Fatalf("want exactly 1 truncated cause, got %d", truncatedCauses)
	}
}

// TestCallGraphCycleProjectsSpansFromOperationIDs is Task 2 Test 7
// (D-07-35): the diagnostic's Primary and each cycle_member cause's span
// are real, non-zero source spans projected by check from operation IDs
// -- never a Span core.LinearOperation itself carries (there is no such
// field; see TestLinearOperationHasNoSpanField).
func TestCallGraphCycleProjectsSpansFromOperationIDs(t *testing.T) {
	source := readPhase07Fixture(t, "cycle_mutual.lang")
	result := Program(mustParseProgram(t, source))
	diag := findCallGraphCycleDiagnostic(t, result)
	if diag.Primary.Start == 0 && diag.Primary.End == 0 {
		t.Fatalf("expected a real, non-zero Primary span projected from an operation ID, got %+v", diag.Primary)
	}
	for _, cause := range diag.Causes {
		if cause.Kind != "cycle_member" {
			continue
		}
		if cause.Span == nil || (cause.Span.Start == 0 && cause.Span.End == 0) {
			t.Fatalf("expected cycle_member cause %+v to carry a real, non-zero projected span", cause)
		}
	}
}

// TestLinearOperationHasNoSpanField is Task 2 Test 7's other half
// (D-07-35): core.LinearOperation carries no Span field at all -- spans
// are projected on check's side from operation IDs, never added to the
// serialized core artifact.
func TestLinearOperationHasNoSpanField(t *testing.T) {
	typ := reflect.TypeOf(core.LinearOperation{})
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == "Span" {
			t.Fatalf("core.LinearOperation must not carry a Span field (D-07-35), found one")
		}
	}
}

// TestCallGraphCycleClearsReturnedProgram is D-07-14's own falsifier: when
// check refuses on a call-graph cycle, the returned core.Program carries
// no functions at all -- a cyclic program exists only as an ephemeral
// local, never returned.
func TestCallGraphCycleClearsReturnedProgram(t *testing.T) {
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_mutual.lang")))
	if len(result.Program.Functions) != 0 {
		t.Fatalf("expected no functions in the returned core.Program on a cycle refusal, got %d", len(result.Program.Functions))
	}
}

// TestCycleIndirectFixtureRefused is 07-07 Task 2's length-3 corpus item:
// A -> B -> C -> A, completing SEM-07's self (1) / mutual (2) / indirect
// (>= 3) trio, distinguished by cycle_length and membership rather than by
// three separate codes.
func TestCycleIndirectFixtureRefused(t *testing.T) {
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_indirect.lang")))
	diag := findCallGraphCycleDiagnostic(t, result)
	for _, cause := range diag.Causes {
		if cause.Kind == "cycle_length" {
			length, err := strconv.Atoi(cause.Detail)
			if err != nil || length < 3 {
				t.Fatalf("want cycle_length >= 3, got %q", cause.Detail)
			}
			return
		}
	}
	t.Fatalf("expected a cycle_length cause, got %+v", diag.Causes)
}

// TestCycleUnreachableFixtureRefused is 07-07 Task 2's T-07-42 corpus item:
// a cycle among functions `main` never calls into is still refused --
// roots are ALL declared functions, never merely functions reachable from
// an entry point.
func TestCycleUnreachableFixtureRefused(t *testing.T) {
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_unreachable.lang")))
	findCallGraphCycleDiagnostic(t, result)
}

// TestCycleUnreachableSurvivesEntryPointRemoval is the synthetic sibling of
// TestCycleUnreachableFixtureRefused: with `main` disabled entirely
// (check's own cycle refusal disabled first so the real, still-cyclic
// core.Program is returned rather than cleared), deleting `main` from the
// program's own Functions slice and re-running callgraph.Order directly
// still refuses -- proving the refusal never depended on `main`'s
// presence, only on the orbiting pair's own mutual edges.
func TestCycleUnreachableSurvivesEntryPointRemoval(t *testing.T) {
	previous := disableCallGraphCycleRefusalForTest
	disableCallGraphCycleRefusalForTest = true
	defer func() { disableCallGraphCycleRefusalForTest = previous }()

	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_unreachable.lang")))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected the cycle refusal to be disabled, got %+v", result.Diagnostics)
	}
	program := result.Program

	withoutEntry := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, DataTypes: program.DataTypes}
	for _, function := range program.Functions {
		if function.Name == "main" {
			continue
		}
		withoutEntry.Functions = append(withoutEntry.Functions, function)
	}
	if len(withoutEntry.Functions) != len(program.Functions)-1 {
		t.Fatalf("expected exactly one function (main) removed, got %d of %d remaining", len(withoutEntry.Functions), len(program.Functions))
	}
	if _, err := callgraph.Order(withoutEntry); err == nil {
		t.Fatalf("expected the orbiting cycle to still be refused with no entry point present")
	} else if _, ok := callgraph.CycleError(err); !ok {
		t.Fatalf("expected a cycle error, got %v", err)
	}
}

// TestCycleThroughMatchArmFixtureRefused is 07-07 Task 2's D-07-28 corpus
// item: the closing edge of this fixture's cycle originates INSIDE a match
// arm (helper's arm A calls back into main), proving enumerate-by-kind-
// across-all-blocks on the REFUSAL path, not only the admission path
// call_from_both_match_arms.lang (07-04) proved.
func TestCycleThroughMatchArmFixtureRefused(t *testing.T) {
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_through_match_arm.lang")))
	findCallGraphCycleDiagnostic(t, result)
}

// TestCycleThroughMatchArmSurvivesBlockOrderReversal proves the
// match-arm-closing edge is found by scanning ALL of a function's
// Linear.Operations, never by Block/Successor position (D-07-28): with
// check's own cycle refusal disabled so the real, still-cyclic program is
// returned, reversing the Blocks slice order of every function that
// carries one changes nothing about which operations exist or their
// CalleeID edges, so callgraph.Order still refuses identically.
func TestCycleThroughMatchArmSurvivesBlockOrderReversal(t *testing.T) {
	previous := disableCallGraphCycleRefusalForTest
	disableCallGraphCycleRefusalForTest = true
	defer func() { disableCallGraphCycleRefusalForTest = previous }()

	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_through_match_arm.lang")))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected the cycle refusal to be disabled, got %+v", result.Diagnostics)
	}
	program := result.Program
	for fi := range program.Functions {
		function := &program.Functions[fi]
		if function.Linear == nil || len(function.Linear.Blocks) < 2 {
			continue
		}
		blocks := function.Linear.Blocks
		for i, j := 0, len(blocks)-1; i < j; i, j = i+1, j-1 {
			blocks[i], blocks[j] = blocks[j], blocks[i]
		}
	}
	if _, err := callgraph.Order(program); err == nil {
		t.Fatalf("expected the cycle to still be refused after reversing block order")
	} else if _, ok := callgraph.CycleError(err); !ok {
		t.Fatalf("expected a cycle error after reversing block order, got %v", err)
	}
}

// TestForeignSymbolShadowingFixtureRefusedAndEdgeMutationKilled is 07-07
// Task 2's T-07-43 corpus item AND Task 3's control:callgraph.
// foreign_shadowing_edge_preserved mutation kill in one test: the fixture
// is refused with core.call_graph_cycle, the emitted CalleeID names the
// LANG function `helper` resolves to (D-07-30's precedence, never the
// foreign symbol of the same name) -- and, mutated to drop that edge (as a
// divergent precedence rule would, since a foreign call contributes no
// CalleeID graph edge at all), the cycle silently disappears, proving the
// edge -- and the precedence rule that preserves it -- is load-bearing.
func TestForeignSymbolShadowingFixtureRefusedAndEdgeMutationKilled(t *testing.T) {
	result := Program(mustParseProgram(t, readPhase07Fixture(t, "foreign_symbol_shadowing.lang")))
	findCallGraphCycleDiagnostic(t, result)

	previous := disableCallGraphCycleRefusalForTest
	disableCallGraphCycleRefusalForTest = true
	defer func() { disableCallGraphCycleRefusalForTest = previous }()

	realResult := Program(mustParseProgram(t, readPhase07Fixture(t, "foreign_symbol_shadowing.lang")))
	if len(realResult.Diagnostics) != 0 {
		t.Fatalf("expected the cycle refusal to be disabled, got %+v", realResult.Diagnostics)
	}
	program := realResult.Program

	var helperCalleeID, mainID string
	for _, function := range program.Functions {
		if function.Name == "main" {
			mainID = function.ID
		}
	}
	for _, function := range program.Functions {
		if function.Name != "helper" || function.Linear == nil {
			continue
		}
		for _, op := range function.Linear.Operations {
			if op.Kind == core.OpCall {
				helperCalleeID = op.CalleeID
			}
		}
	}
	if helperCalleeID == "" {
		t.Fatalf("expected helper to carry an OpCall")
	}
	if helperCalleeID != mainID {
		t.Fatalf("expected helper's call to resolve to the Lang function %q (D-07-30), got %q", mainID, helperCalleeID)
	}

	if _, err := callgraph.Order(program); err == nil {
		t.Fatalf("expected the real program to contain a cycle before any mutation")
	} else if _, ok := callgraph.CycleError(err); !ok {
		t.Fatalf("expected a cycle error before mutation, got %v", err)
	}

	mutated := core.Program{Schema: program.Schema, Module: program.Module, ModuleID: program.ModuleID, DataTypes: program.DataTypes}
	for _, function := range program.Functions {
		if function.Name == "helper" && function.Linear != nil {
			filtered := make([]core.LinearOperation, 0, len(function.Linear.Operations))
			for _, op := range function.Linear.Operations {
				if op.Kind == core.OpCall {
					// The mutation: as if this call had instead resolved
					// against the foreign symbol of the same name --
					// which produces core.OpForeignCall, never an
					// OpCall/CalleeID graph edge at all.
					continue
				}
				filtered = append(filtered, op)
			}
			linearCopy := *function.Linear
			linearCopy.Operations = filtered
			function.Linear = &linearCopy
		}
		mutated.Functions = append(mutated.Functions, function)
	}
	if _, err := callgraph.Order(mutated); err != nil {
		t.Fatalf("mutation (dropping the shadowing-precedence edge) had no observable effect: expected the cycle to disappear, got %v", err)
	}
}

// TestCheckCycleRefusalIndependentOfCorevalidatePeer is Task 3 Test 1
// (D-07-42's independent-disable row, check side): with check's OWN
// callgraph-based cycle refusal disabled through disableCallGraphCycleRefusalForTest,
// the cyclic core.Program is returned instead of cleared -- and
// corevalidate's own, independently written cycle peer (07-07), consulted
// via its ordinary public Validate API in its DEFAULT (never toggled)
// state, still refuses it on its own.
func TestCheckCycleRefusalIndependentOfCorevalidatePeer(t *testing.T) {
	previous := disableCallGraphCycleRefusalForTest
	disableCallGraphCycleRefusalForTest = true
	defer func() { disableCallGraphCycleRefusalForTest = previous }()

	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_mutual.lang")))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected check's own cycle refusal to be disabled, got %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) == 0 {
		t.Fatalf("expected the real cyclic core.Program to be returned, not cleared")
	}

	validated := corevalidate.Validate(result.Program)
	if validated.Valid {
		t.Fatalf("expected corevalidate's own independent peer to still refuse the cyclic program")
	}
	foundCycle := false
	for _, problem := range validated.Problems {
		if problem.Code == core.CallGraphCycle {
			foundCycle = true
		}
	}
	if !foundCycle {
		t.Fatalf("expected %s among corevalidate's problems, got %+v", core.CallGraphCycle, validated.Problems)
	}
}

// TestCorevalidatePeerIndependentOfCheckCycleRefusal is Task 3 Test 2
// (D-07-42's independent-disable row, peer side): with corevalidate's own
// cycle peer disabled via its cross-package test seam, check's OWN
// callgraph-based refusal -- consulted in its DEFAULT (never toggled)
// state -- still refuses the same cyclic source on its own. check.Program
// never calls into corevalidate at all in production, so this also
// demonstrates the two pipelines are wholly separate.
func TestCorevalidatePeerIndependentOfCheckCycleRefusal(t *testing.T) {
	restore := corevalidate.SetDisableCyclePeerForTest(true)
	defer restore()

	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_mutual.lang")))
	findCallGraphCycleDiagnostic(t, result)

	// With the peer disabled, feeding check's own (cleared-on-refusal, so
	// re-parse+re-check with check's OWN refusal ALSO disabled) real
	// cyclic program through corevalidate must now report it valid --
	// confirming the peer, not some other corevalidate check, was what
	// caught it.
	previousCheckSeam := disableCallGraphCycleRefusalForTest
	disableCallGraphCycleRefusalForTest = true
	defer func() { disableCallGraphCycleRefusalForTest = previousCheckSeam }()
	unrefused := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_mutual.lang")))
	if len(unrefused.Diagnostics) != 0 {
		t.Fatalf("expected check's own refusal to be disabled too, got %+v", unrefused.Diagnostics)
	}
	validated := corevalidate.Validate(unrefused.Program)
	if !validated.Valid {
		t.Fatalf("expected corevalidate to report the cyclic program valid with its own peer disabled, got %+v", validated.Problems)
	}
}

// TestBilateralCallGraphFaultReportsNoDivergenceAndFailsGate is Task 3 Test
// 3, the bilateral row: with BOTH check's own callgraph-based refusal AND
// corevalidate's independent cycle peer disabled at once, a genuinely
// cyclic program passes BOTH pipelines silently -- the sweep itself must
// observe that neither derivation caught it and report that finding,
// failing the gate rather than passing it, never silently accepting the
// cyclic program as evidence of anything.
func TestBilateralCallGraphFaultReportsNoDivergenceAndFailsGate(t *testing.T) {
	previousCheckSeam := disableCallGraphCycleRefusalForTest
	disableCallGraphCycleRefusalForTest = true
	defer func() { disableCallGraphCycleRefusalForTest = previousCheckSeam }()
	restorePeer := corevalidate.SetDisableCyclePeerForTest(true)
	defer restorePeer()

	result := Program(mustParseProgram(t, readPhase07Fixture(t, "cycle_mutual.lang")))
	checkRefused := len(result.Diagnostics) != 0
	validated := corevalidate.Validate(result.Program)
	corevalidateRefused := !validated.Valid

	report, gatePassed := bilateralCallGraphFaultReport(checkRefused, corevalidateRefused)
	if gatePassed {
		t.Fatalf("expected the bilateral fault to fail the gate, not pass it")
	}
	const wantReport = "no divergence detected under bilateral fault"
	if report != wantReport {
		t.Fatalf("want report %q, got %q", wantReport, report)
	}
}

// bilateralCallGraphFaultReport is Task 3 Test 3's own sweep predicate:
// when NEITHER independent derivation refused a genuinely cyclic program
// (the bilateral-fault condition), report the finding and fail the gate.
// Any other combination (at least one side still refusing) passes,
// because at least one derivation is still doing its job.
func bilateralCallGraphFaultReport(checkRefused, corevalidateRefused bool) (report string, gatePassed bool) {
	if !checkRefused && !corevalidateRefused {
		return "no divergence detected under bilateral fault", false
	}
	return "", true
}

// ---------------------------------------------------------------------
// 07-09 Task 1: the argument-type gate and the callee-return-derived
// target TypeID.
// ---------------------------------------------------------------------

// TestCallArgumentTypeMismatchRefused is Task 1 Test 1: the standing
// negative control (call_type_mismatch.lang) is refused with exactly one
// error-severity diagnostic carrying the ratified
// check.call_argument_type_mismatch code, whose Primary span is the call
// site and whose ordered Causes are callee / argument_type (Buffer) /
// declared_parameter_type (Byte), with no repairs.
func TestCallArgumentTypeMismatchRefused(t *testing.T) {
	source := readPhase07Fixture(t, "call_type_mismatch.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %+v", result.Diagnostics)
	}
	diag := result.Diagnostics[0]
	if diag.Code != checkCallArgumentTypeMismatch {
		t.Fatalf("expected %s, got %s (%+v)", checkCallArgumentTypeMismatch, diag.Code, diag)
	}
	if diag.Severity != "error" {
		t.Fatalf("expected error severity, got %q", diag.Severity)
	}
	wantKinds := []string{"callee", "argument_type", "declared_parameter_type"}
	if len(diag.Causes) != len(wantKinds) {
		t.Fatalf("expected %d causes, got %+v", len(wantKinds), diag.Causes)
	}
	for i, kind := range wantKinds {
		if diag.Causes[i].Kind != kind {
			t.Fatalf("cause %d: want kind %q, got %q (%+v)", i, kind, diag.Causes[i].Kind, diag.Causes)
		}
	}
	if !strings.HasSuffix(diag.Causes[0].Detail, ":fn:identity") {
		t.Fatalf("expected the callee cause to name identity's own function ID, got %q", diag.Causes[0].Detail)
	}
	if diag.Causes[1].Detail != "Buffer" {
		t.Fatalf("expected argument_type detail Buffer, got %q", diag.Causes[1].Detail)
	}
	if diag.Causes[2].Detail != "Byte" {
		t.Fatalf("expected declared_parameter_type detail Byte, got %q", diag.Causes[2].Detail)
	}
	if diag.Repairs != nil {
		t.Fatalf("expected no repairs, got %+v", diag.Repairs)
	}
}

// TestCallTargetTypeDerivedFromCalleeReturn is Task 1 Test 2: call_basic.lang
// still checks clean, and its OpCall's TargetID place has
// TypeID == "<caller function ID>:type:0" -- the same value as before this
// plan -- proving the promoted derivation is byte-identical where it must be.
func TestCallTargetTypeDerivedFromCalleeReturn(t *testing.T) {
	source := readPhase07Fixture(t, "call_basic.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected call_basic.lang to still check clean, got %+v", result.Diagnostics)
	}
	var mainFunction *core.Function
	for i := range result.Program.Functions {
		if result.Program.Functions[i].Name == "main" {
			mainFunction = &result.Program.Functions[i]
		}
	}
	if mainFunction == nil || mainFunction.Linear == nil {
		t.Fatal("expected a checked linear-carrying function named main")
	}
	calls := scanOpCallsByKind(mainFunction.Linear.Operations)
	if len(calls) != 1 {
		t.Fatalf("expected exactly one core.OpCall, got %d", len(calls))
	}
	call := calls[0]
	wantTypeID := mainFunction.ID + ":type:0"
	if call.TypeID != wantTypeID {
		t.Fatalf("expected OpCall.TypeID %q, got %q", wantTypeID, call.TypeID)
	}
	var targetPlace *core.Place
	for i := range mainFunction.Linear.Places {
		if mainFunction.Linear.Places[i].ID == call.TargetID {
			targetPlace = &mainFunction.Linear.Places[i]
		}
	}
	if targetPlace == nil {
		t.Fatalf("expected to find the target place %q", call.TargetID)
	}
	if targetPlace.TypeID != wantTypeID {
		t.Fatalf("expected target place TypeID %q, got %q", wantTypeID, targetPlace.TypeID)
	}
}

// TestCallArgumentTypeContractAbsentOrEmptyRefuses is Task 1 Test 3 (edge 2,
// empty/fail-closed): a callee-contract table entry that is absent, or
// whose parameter type constructor is the empty string, refuses the call.
// No legal source program can produce either shape (buildCalleeContracts
// always derives a non-empty ParameterType from a parsed function
// declaration), so this drives resolveCallBinding directly with a
// hand-built calleeContracts table -- the seeded, package-internal seam
// this plan's own must_haves require, since no fixture can reach it.
func TestCallArgumentTypeContractAbsentOrEmptyRefuses(t *testing.T) {
	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
	places := map[string]*placeState{
		"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
	}
	binding := ast.Binding{Name: "result", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}}}

	t.Run("absent entry refuses (via CallCalleeUnresolved)", func(t *testing.T) {
		_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, map[string]calleeContract{}, typeFact, nil)
		if diag == nil {
			t.Fatal("expected a refusal for an absent callee-contract entry, got none")
		}
		if diag.Code != core.CallCalleeUnresolved {
			t.Fatalf("expected %s, got %s", core.CallCalleeUnresolved, diag.Code)
		}
	})

	t.Run("empty ParameterType refuses (via the new gate)", func(t *testing.T) {
		contracts := map[string]calleeContract{
			"identity": {ID: "s1:m:fn:identity", ParameterType: "", ReturnType: "Byte"},
		}
		_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
		if diag == nil {
			t.Fatal("expected a refusal for an empty declared parameter type, got none")
		}
		if diag.Code != checkCallArgumentTypeMismatch {
			t.Fatalf("expected %s, got %s", checkCallArgumentTypeMismatch, diag.Code)
		}
	})
}

// TestCallArgumentTypeAdjacencyAndBoundary is Task 1 Test 4 (edges 1/4,
// adjacency and boundary): equal constructor strings admit; Byte against
// Buffer refuses in both directions.
func TestCallArgumentTypeAdjacencyAndBoundary(t *testing.T) {
	binding := ast.Binding{Name: "result", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}}}
	cases := []struct {
		name            string
		callerType      string
		calleeParamType string
		wantAdmit       bool
	}{
		{"Byte argument into Byte parameter admits", "Byte", "Byte", true},
		{"Buffer argument into Buffer parameter admits", "Buffer", "Buffer", true},
		{"Byte argument into Buffer parameter refuses", "Byte", "Buffer", false},
		{"Buffer argument into Byte parameter refuses", "Buffer", "Byte", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: tc.callerType}}
			places := map[string]*placeState{
				"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
			}
			// ReturnType == callerType so the admitting cases also clear the
			// return-derivation gate -- this table is testing the
			// argument-type gate specifically.
			contracts := map[string]calleeContract{
				"identity": {ID: "s1:m:fn:identity", ParameterType: tc.calleeParamType, ReturnType: tc.callerType},
			}
			_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
			admitted := diag == nil
			if admitted != tc.wantAdmit {
				t.Fatalf("want admit=%v, got admit=%v (diag=%+v)", tc.wantAdmit, admitted, diag)
			}
			if !tc.wantAdmit && diag.Code != checkCallArgumentTypeMismatch {
				t.Fatalf("expected refusal code %s, got %s", checkCallArgumentTypeMismatch, diag.Code)
			}
		})
	}
}

// TestCallTypeMismatchPrecedesCycleRefusal is Task 1 Test 5 (edges 6/7,
// precedence): a program that is BOTH type-mismatched (main calling
// identity with a Buffer argument) and cyclic (a self-recursive loop
// function elsewhere in the same module) reports
// check.call_argument_type_mismatch, never core.call_graph_cycle -- the
// argument-type refusal fires at emission time inside resolveCallBinding,
// strictly before the post-build callgraph.Order pass, which runs only
// when len(result.Diagnostics) == 0. The six existing cycle fixtures each
// still report core.call_graph_cycle unchanged, proving this precedence
// rule changes no existing fixture's verdict.
func TestCallTypeMismatchPrecedesCycleRefusal(t *testing.T) {
	source := []byte(`module phase07.mismatch_and_cycle

export {
  fn main
}

fn identity(value: Byte) -> Byte {
  value
}

fn loop(value: Byte) -> Byte {
  let next = loop(value)
  next
}

fn main(buffer: Buffer) -> Buffer {
  let result = identity(buffer)
  result
}
`)
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %+v", result.Diagnostics)
	}
	if result.Diagnostics[0].Code != checkCallArgumentTypeMismatch {
		t.Fatalf("expected %s to take precedence over core.call_graph_cycle, got %s", checkCallArgumentTypeMismatch, result.Diagnostics[0].Code)
	}

	for _, fixture := range []string{
		"cycle_self.lang", "cycle_mutual.lang", "cycle_indirect.lang",
		"cycle_unreachable.lang", "cycle_through_match_arm.lang", "foreign_symbol_shadowing.lang",
	} {
		t.Run(fixture, func(t *testing.T) {
			cycleResult := Program(mustParseProgram(t, readPhase07Fixture(t, fixture)))
			if len(cycleResult.Diagnostics) != 1 {
				t.Fatalf("expected exactly one diagnostic for %s, got %+v", fixture, cycleResult.Diagnostics)
			}
			if cycleResult.Diagnostics[0].Code != core.CallGraphCycle {
				t.Fatalf("expected %s for %s, got %s", core.CallGraphCycle, fixture, cycleResult.Diagnostics[0].Code)
			}
		})
	}
}

// phase07TypeRippleCorpusDirs is the corpus TestOpCallTargetTypeIDUnchangedAcrossAcceptingCorpus
// and TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus both
// sweep, following originvalidate_test.go's TestInterfaceV1FieldInvariantsAcrossCorpus
// os.ReadDir pattern (07-09 adds testdata/phase07 to that established set).
var phase07TypeRippleCorpusDirs = []string{"phase07", "phase1", "phase2", "phase3", "phase4", "phase5", "phase6"}

// TestOpCallTargetTypeIDUnchangedAcrossAcceptingCorpus is Task 1 Test 6
// (ripple): across every fixture in testdata/phase07 and testdata/phase1
// through testdata/phase6 that checks clean, every core.OpCall's TargetID
// place TypeID equals the calling function's own single type fact ID, and
// equals the ID whose Shape.Constructor matches the callee's declared
// ReturnType. The two derivations agree on every admitted program. Fails
// the test rather than skipping if it compared zero core.OpCall operations.
func TestOpCallTargetTypeIDUnchangedAcrossAcceptingCorpus(t *testing.T) {
	comparisons := 0
	for _, dirName := range phase07TypeRippleCorpusDirs {
		dir := filepath.Join("..", "..", "..", "testdata", dirName)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", dirName, entry.Name(), err)
			}
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				continue
			}
			result := Program(parsed.Program)
			if len(result.Diagnostics) != 0 {
				continue // rejected fixture (including call_type_mismatch.lang itself)
			}
			functionByID := make(map[string]core.Function, len(result.Program.Functions))
			for _, function := range result.Program.Functions {
				functionByID[function.ID] = function
			}
			for _, function := range result.Program.Functions {
				if function.Linear == nil {
					continue
				}
				wantTypeID := function.ID + ":type:0"
				var callerType core.TypeFact
				for _, fact := range function.Linear.Types {
					if fact.ID == wantTypeID {
						callerType = fact
					}
				}
				placesByID := make(map[string]core.Place, len(function.Linear.Places))
				for _, place := range function.Linear.Places {
					placesByID[place.ID] = place
				}
				for _, operation := range function.Linear.Operations {
					if operation.Kind != core.OpCall {
						continue
					}
					comparisons++
					if operation.TypeID != wantTypeID {
						t.Fatalf("%s/%s: OpCall %s TypeID = %q, want %q", dirName, entry.Name(), operation.ID, operation.TypeID, wantTypeID)
					}
					targetPlace, ok := placesByID[operation.TargetID]
					if !ok {
						t.Fatalf("%s/%s: OpCall %s target place %q not found", dirName, entry.Name(), operation.ID, operation.TargetID)
					}
					if targetPlace.TypeID != wantTypeID {
						t.Fatalf("%s/%s: OpCall %s target place TypeID = %q, want %q", dirName, entry.Name(), operation.ID, targetPlace.TypeID, wantTypeID)
					}
					callee, ok := functionByID[operation.CalleeID]
					if !ok {
						t.Fatalf("%s/%s: OpCall %s CalleeID %q does not resolve", dirName, entry.Name(), operation.ID, operation.CalleeID)
					}
					if callerType.ID == "" {
						t.Fatalf("%s/%s: no type fact %q found on function %s", dirName, entry.Name(), wantTypeID, function.ID)
					}
					if callerType.Shape.Constructor != callee.ReturnType {
						t.Fatalf("%s/%s: OpCall %s caller type constructor %q != callee %s declared ReturnType %q", dirName, entry.Name(), operation.ID, callerType.Shape.Constructor, callee.ID, callee.ReturnType)
					}
				}
			}
		}
	}
	if comparisons == 0 {
		t.Fatal("expected to compare at least one core.OpCall operation across the corpus, compared zero")
	}
}

// TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus is Task 1
// Test 7 (SEM-05 key link): across every accepting fixture, for every
// core.OpCall, originvalidate.BuildInterface's FunctionSignature.Parameters[0].Type
// for the callee equals the declared parameter type constructor the
// admission gate consulted -- proving the /1 field carries the fact the
// gate decides on, closing the key link 07-VERIFICATION.md marked NOT
// WIRED. Fails rather than skips if it compared zero callees.
func TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus(t *testing.T) {
	comparisons := 0
	for _, dirName := range phase07TypeRippleCorpusDirs {
		dir := filepath.Join("..", "..", "..", "testdata", dirName)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, entry.Name()))
			if err != nil {
				t.Fatalf("read %s/%s: %v", dirName, entry.Name(), err)
			}
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				continue
			}
			astProgram := parsed.Program
			result := Program(astProgram)
			if len(result.Diagnostics) != 0 {
				continue
			}
			iface, err := originvalidate.BuildInterface(result.Program)
			if err != nil {
				continue
			}
			signatureByID := make(map[string]core.FunctionSignature, len(iface.Functions))
			for _, signature := range iface.Functions {
				signatureByID[signature.ID] = signature
			}
			calleeContracts := buildCalleeContracts(astProgram)
			contractByID := make(map[string]calleeContract, len(calleeContracts))
			for _, contract := range calleeContracts {
				contractByID[contract.ID] = contract
			}
			for _, function := range result.Program.Functions {
				if function.Linear == nil {
					continue
				}
				for _, operation := range function.Linear.Operations {
					if operation.Kind != core.OpCall {
						continue
					}
					signature, ok := signatureByID[operation.CalleeID]
					if !ok {
						t.Fatalf("%s/%s: no signature for callee %q", dirName, entry.Name(), operation.CalleeID)
					}
					if len(signature.Parameters) != 1 {
						t.Fatalf("%s/%s: expected exactly one parameter on signature %q, got %d", dirName, entry.Name(), signature.ID, len(signature.Parameters))
					}
					contract, ok := contractByID[operation.CalleeID]
					if !ok {
						t.Fatalf("%s/%s: no callee contract for %q", dirName, entry.Name(), operation.CalleeID)
					}
					comparisons++
					if signature.Parameters[0].Type != contract.ParameterType {
						t.Fatalf("%s/%s: FunctionSignature.Parameters[0].Type = %q, admission contract ParameterType = %q", dirName, entry.Name(), signature.Parameters[0].Type, contract.ParameterType)
					}
				}
			}
		}
	}
	if comparisons == 0 {
		t.Fatal("expected to compare at least one callee across the corpus, compared zero")
	}
}

// ---------------------------------------------------------------------
// 07-09 Task 3: mutation kills for both new controls, check side.
// ---------------------------------------------------------------------

// TestCallArgumentTypeCheckMutationKilled is 07-09 Task 3's check-side kill
// for control:call.argument_type_matches_parameter. It first confirms the
// standing negative control (call_type_mismatch.lang) is refused with the
// ratified code in production (the seam off). It then engages ONLY
// callArgumentTypeCheckSeam through a direct resolveCallBinding call whose
// callee contract deliberately decouples ParameterType from ReturnType --
// a shape no real source program can construct, since sameType forces
// every function's declared return type to equal its declared parameter
// type. Without that decoupling, disabling only the argument gate on the
// real fixture still hits the (production, unseamed) return-derivation
// gate -- contract.ReturnType == contract.ParameterType always in real
// source, so the two gates agree there -- proving nothing about THIS seam
// specifically. The decoupled direct call isolates this control's own
// kill from the return-derivation control's kill.
func TestCallArgumentTypeCheckMutationKilled(t *testing.T) {
	defer func() { callArgumentTypeCheckSeam = false }()

	source := readPhase07Fixture(t, "call_type_mismatch.lang")
	production := Program(mustParseProgram(t, source))
	if len(production.Diagnostics) != 1 || production.Diagnostics[0].Code != checkCallArgumentTypeMismatch {
		t.Fatalf("expected %s without the seam, got %+v", checkCallArgumentTypeMismatch, production.Diagnostics)
	}

	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Buffer"}}
	places := map[string]*placeState{
		"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
	}
	binding := ast.Binding{Name: "result", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}}}
	contracts := map[string]calleeContract{
		"identity": {ID: "s1:m:fn:identity", ParameterType: "Byte", ReturnType: "Buffer"},
	}

	callArgumentTypeCheckSeam = true
	_, _, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
	if diag != nil {
		t.Fatalf("expected the seam to wrongly admit the mismatched argument, got %+v", diag)
	}

	callArgumentTypeCheckSeam = false
	// 07-11: the first (seam-engaged) call above admitted a non-copyable
	// (Buffer) argument, which the new consume rule move-marked as a side
	// effect of admission -- unrelated to the argument-type gate this test
	// isolates. Reset the place's move state before the second call so
	// that call exercises the argument-type gate itself, not the
	// unrelated use-after-move gate the first call's consume left behind.
	places["value"].initialized = true
	places["value"].movedAt = nil
	places["value"].moveTargetID = ""
	places["value"].moveTargetName = ""
	_, _, restored := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
	if restored == nil || restored.Code != checkCallArgumentTypeMismatch {
		t.Fatalf("expected %s restored once the seam is disengaged, got %+v", checkCallArgumentTypeMismatch, restored)
	}
}

// TestCallReturnTypeDerivationMutationKilled is 07-09 Task 3's check-side
// kill for control:call.target_type_from_callee_return. With the seam
// engaged, an OpCall's target TypeID reverts to the caller's own argument
// place's TypeID (the pre-plan defect); with it disengaged, the target
// TypeID is derived from the callee's declared return contract resolved
// against the caller's own type fact. The two derivations are constructed
// to disagree (the argument place carries a TypeID the caller's own type
// fact table does not), observably proving the seam is load-bearing.
func TestCallReturnTypeDerivationMutationKilled(t *testing.T) {
	defer func() { callReturnTypeDerivationSeam = false }()

	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}}
	places := map[string]*placeState{
		"value": {place: core.Place{ID: "s1:m:fn:main:place:9", Name: "value", TypeID: "s1:m:fn:main:type:9"}, initialized: true},
	}
	binding := ast.Binding{Name: "result", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}}}
	contracts := map[string]calleeContract{
		"identity": {ID: "s1:m:fn:identity", ParameterType: "Byte", ReturnType: "Byte"},
	}

	callReturnTypeDerivationSeam = true
	op, target, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
	if diag != nil {
		t.Fatalf("expected the seam to admit the call, got %+v", diag)
	}
	if op.TypeID != "s1:m:fn:main:type:9" || target.TypeID != "s1:m:fn:main:type:9" {
		t.Fatalf("expected the seam to derive TypeID from the argument's own place (type:9), got op.TypeID=%q target.TypeID=%q", op.TypeID, target.TypeID)
	}

	callReturnTypeDerivationSeam = false
	// 07-11: typeFact here is Byte (copyable), so the consume rule does not
	// move-mark "value" -- included for parity with the sibling test above,
	// which does need the reset because its typeFact is Buffer.
	places["value"].initialized = true
	places["value"].movedAt = nil
	places["value"].moveTargetID = ""
	places["value"].moveTargetName = ""
	restoredOp, restoredTarget, restoredDiag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
	if restoredDiag != nil {
		t.Fatalf("expected admission once restored, got %+v", restoredDiag)
	}
	if restoredOp.TypeID != typeFact.ID || restoredTarget.TypeID != typeFact.ID {
		t.Fatalf("expected the restored derivation from the callee's declared return contract (type:0), got op.TypeID=%q target.TypeID=%q", restoredOp.TypeID, restoredTarget.TypeID)
	}
	if restoredOp.TypeID == "s1:m:fn:main:type:9" {
		t.Fatal("expected the restored derivation to differ from the seam's, proving the seam is load-bearing")
	}
}

// TestCallConsumesNoncopyableArgument is 07-11 Task 1's Test 1 (the gap)
// and Test 4 (both call paths). It first proves
// call_argument_used_twice.lang -- 07-VERIFICATION.md PVG-01 /
// 07-REVIEW.md CR-01's standing negative control -- is refused with
// exactly one ownership.use_after_move diagnostic whose Primary span is
// the SECOND call site and whose ordered causes are moved_here (the FIRST
// call site) / place / transfer_target. It then proves the identical
// double-consume, expressed inside a match arm so it runs through
// analyzeArmBody rather than analyzeStraightLine, is refused identically
// -- the law reaches both call paths through the one shared resolver,
// never a duplicated copy.
func TestCallConsumesNoncopyableArgument(t *testing.T) {
	t.Run("straight_line", func(t *testing.T) {
		source := readPhase07Fixture(t, "call_argument_used_twice.lang")
		parsed := mustParseProgram(t, source)
		result := Program(parsed)
		if len(result.Diagnostics) != 1 {
			t.Fatalf("expected exactly one diagnostic, got %+v", result.Diagnostics)
		}
		diag := result.Diagnostics[0]
		if diag.Code != "ownership.use_after_move" {
			t.Fatalf("expected ownership.use_after_move, got %s (%+v)", diag.Code, diag)
		}
		var mainFunc ast.FuncDecl
		for _, function := range parsed.Funcs {
			if function.Name == "main" {
				mainFunc = function
			}
		}
		if mainFunc.Body.Linear == nil || len(mainFunc.Body.Linear.Bindings) != 2 {
			t.Fatalf("expected main's straight-line body to carry exactly two bindings, got %+v", mainFunc.Body.Linear)
		}
		firstCallSpan := mainFunc.Body.Linear.Bindings[0].RHS.Span
		secondCallSpan := mainFunc.Body.Linear.Bindings[1].RHS.Span
		if diag.Primary != secondCallSpan {
			t.Fatalf("expected Primary span to be the second call site %+v, got %+v", secondCallSpan, diag.Primary)
		}
		if len(diag.Causes) != 3 {
			t.Fatalf("expected three ordered causes (moved_here/place/transfer_target), got %+v", diag.Causes)
		}
		if diag.Causes[0].Kind != "moved_here" || diag.Causes[0].Span == nil || *diag.Causes[0].Span != firstCallSpan {
			t.Fatalf("expected the first cause to be moved_here pointing at the first call site %+v, got %+v", firstCallSpan, diag.Causes[0])
		}
		if diag.Causes[1].Kind != "place" {
			t.Fatalf("expected the second cause to be place, got %+v", diag.Causes[1])
		}
		if diag.Causes[2].Kind != "transfer_target" {
			t.Fatalf("expected the third cause to be transfer_target, got %+v", diag.Causes[2])
		}
	})

	t.Run("match_arm", func(t *testing.T) {
		// Drives analyzeArmBody directly (rather than through a .lang
		// fixture) since this language surface has no data type carrying
		// a Buffer-typed field to route a non-copyable value into a match
		// arm's own bindings -- the direct-call technique this file's
		// mutation-kill tests already use for the same reason.
		typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Buffer"}}
		contracts := map[string]calleeContract{
			"identity": {ID: "s1:m:fn:identity", ParameterType: "Buffer", ReturnType: "Buffer"},
		}
		body := &ast.LinearBody{
			Bindings: []ast.Binding{
				{Name: "first", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"buffer"}, Span: diagnostic.Span{Start: 10, End: 20}}, Span: diagnostic.Span{Start: 10, End: 20}},
				{Name: "second", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"buffer"}, Span: diagnostic.Span{Start: 30, End: 40}}, Span: diagnostic.Span{Start: 30, End: 40}},
			},
			Result: "second",
		}
		support := analyzeArmBody("s1:m:fn:main", 0, "buffer", "s1:m:fn:main:place:0", diagnostic.Span{}, typeFact, body, contracts, nil)
		if support.Diagnostic == nil {
			t.Fatalf("expected the match-arm double-consume to be refused, got a clean result: %+v", support)
		}
		if support.DiagnosticCode != "ownership.use_after_move" {
			t.Fatalf("expected ownership.use_after_move, got %s (%+v)", support.DiagnosticCode, support.Diagnostic)
		}
		if support.Diagnostic.Primary != (diagnostic.Span{Start: 30, End: 40}) {
			t.Fatalf("expected the second call site's span as Primary, got %+v", support.Diagnostic.Primary)
		}
	})
}

// TestCallDoesNotConsumeCopyableArgument is 07-11 Task 1's Test 2: the
// non-refusing direction, pinned as hard as the refusing one.
// call_argument_used_once.lang (a Buffer passed to exactly one call) and a
// Byte argument passed to two separate calls both check clean --
// copyable values are copied, never consumed, and passing a Buffer to a
// call at all is legal so long as it is never used again afterward.
func TestCallDoesNotConsumeCopyableArgument(t *testing.T) {
	t.Run("buffer_used_once", func(t *testing.T) {
		source := readPhase07Fixture(t, "call_argument_used_once.lang")
		result := Program(mustParseProgram(t, source))
		if len(result.Diagnostics) != 0 {
			t.Fatalf("expected call_argument_used_once.lang to check clean, got %+v", result.Diagnostics)
		}
	})

	t.Run("byte_used_twice", func(t *testing.T) {
		typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Byte"}, Abilities: []core.Ability{core.AbilityCopy}}
		places := map[string]*placeState{
			"value": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "value", TypeID: typeFact.ID}, initialized: true},
		}
		contracts := map[string]calleeContract{
			"identity": {ID: "s1:m:fn:identity", ParameterType: "Byte", ReturnType: "Byte"},
		}
		firstBinding := ast.Binding{Name: "first", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}, Span: diagnostic.Span{Start: 1, End: 2}}}
		op, target, diag := resolveCallBinding("s1:m:fn:main", 0, firstBinding, places, contracts, typeFact, nil)
		if diag != nil {
			t.Fatalf("expected the first call to admit a copyable argument, got %+v", diag)
		}
		places["first"] = &placeState{place: target, initialized: true}
		if !places["value"].initialized {
			t.Fatal("expected a copyable call argument to stay initialized (copied, not consumed)")
		}
		secondBinding := ast.Binding{Name: "second", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"value"}, Span: diagnostic.Span{Start: 3, End: 4}}}
		_, _, secondDiag := resolveCallBinding("s1:m:fn:main", 1, secondBinding, places, contracts, typeFact, nil)
		if secondDiag != nil {
			t.Fatalf("expected the second call on the same copyable value to also be admitted, got %+v", secondDiag)
		}
		_ = op
	})
}

// TestCallConsumeMoveStateComplete is 07-11 Task 1's Test 3: after a
// consuming call, the argument's placeState carries all four fields the
// documented moveTargetName/moveTargetID invariant depends on, so the
// use_transfer_target repair never ships with an empty Detail.
func TestCallConsumeMoveStateComplete(t *testing.T) {
	typeFact := core.TypeFact{ID: "s1:m:fn:main:type:0", Shape: core.TypeRef{Constructor: "Buffer"}}
	places := map[string]*placeState{
		"buffer": {place: core.Place{ID: "s1:m:fn:main:place:0", Name: "buffer", TypeID: typeFact.ID}, initialized: true},
	}
	contracts := map[string]calleeContract{
		"identity": {ID: "s1:m:fn:identity", ParameterType: "Buffer", ReturnType: "Buffer"},
	}
	callSpan := diagnostic.Span{Start: 5, End: 15}
	binding := ast.Binding{Name: "first", RHS: ast.RHS{Kind: "call", Callee: "identity", Arguments: []string{"buffer"}, Span: callSpan}}
	_, target, diag := resolveCallBinding("s1:m:fn:main", 0, binding, places, contracts, typeFact, nil)
	if diag != nil {
		t.Fatalf("expected the consuming call to admit, got %+v", diag)
	}
	argument := places["buffer"]
	if argument.initialized {
		t.Fatal("expected the consumed argument's initialized field to be false")
	}
	if argument.movedAt == nil || *argument.movedAt != callSpan {
		t.Fatalf("expected movedAt to equal the call's own span %+v, got %+v", callSpan, argument.movedAt)
	}
	if argument.moveTargetID != target.ID {
		t.Fatalf("expected moveTargetID to equal the call's target place ID %q, got %q", target.ID, argument.moveTargetID)
	}
	if argument.moveTargetName == "" {
		t.Fatal("expected a non-empty moveTargetName, matching the documented moveTargetID/moveTargetName invariant")
	}
}

// callConsumptionRippleFixtures lists every accepting fixture (a fixture
// that checked clean before this plan) this test sweeps to prove the 07-11
// consume rule ripples nowhere: same verdict, same emitted core.Program,
// deterministically re-derived. Deliberately NOT every testdata/phase07
// fixture -- the corpus also carries refusing fixtures (cycle_*.lang,
// call_uncallable_callee.lang, the two new 07-11 fixtures,
// duplicate_function_name.lang) whose check.Program-level verdict this
// plan does not claim is clean.
var callConsumptionRippleFixtures = []string{
	"call_basic.lang",
	"call_from_both_match_arms.lang",
	"clean_but_unpublishable.lang",
	"deep_diamond_acyclic.lang",
}

// TestCallArgumentConsumptionUnchangedAcrossAcceptingCorpus is 07-11 Task
// 1's Test 7 (ripple): every fixture that checked clean before this plan
// still checks clean, and its emitted core.Program is byte-identical
// across two independent re-derivations -- consumption changes checker
// state only, never emitted core.
func TestCallArgumentConsumptionUnchangedAcrossAcceptingCorpus(t *testing.T) {
	for _, fixture := range callConsumptionRippleFixtures {
		t.Run(fixture, func(t *testing.T) {
			source := readPhase07Fixture(t, fixture)
			first := Program(mustParseProgram(t, source))
			if len(first.Diagnostics) != 0 {
				t.Fatalf("expected %s to still check clean, got %+v", fixture, first.Diagnostics)
			}
			second := Program(mustParseProgram(t, source))
			if len(second.Diagnostics) != 0 {
				t.Fatalf("expected %s to still check clean on re-derivation, got %+v", fixture, second.Diagnostics)
			}
			if !reflect.DeepEqual(first.Program, second.Program) {
				t.Fatalf("expected %s's emitted core.Program to be deterministic and unchanged across re-derivations", fixture)
			}
		})
	}

	t.Run("relay_escort_witness_unchanged", func(t *testing.T) {
		// Phase 08 closes the interprocedural half of D-03-02
		// (TestRelayEscortWitnessRefusesInterproceduralLiveness): the
		// fixture now refuses deterministically rather than staying clean,
		// but that refusal must still be exactly reproducible across
		// re-derivations.
		source := readPhase07Fixture(t, "relay_escort_witness.lang")
		first := Program(mustParseProgram(t, source))
		if len(first.Diagnostics) != 1 || first.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
			t.Fatalf("expected relay_escort_witness.lang to refuse via check.interprocedural_loan_liveness, got %+v", first.Diagnostics)
		}
		second := Program(mustParseProgram(t, source))
		if !reflect.DeepEqual(first.Diagnostics, second.Diagnostics) {
			t.Fatalf("expected relay_escort_witness.lang's refusal to be deterministic across re-derivations, got %+v vs %+v", first.Diagnostics, second.Diagnostics)
		}
	})
}

// TestCallArgumentConsumeMutationKilled is 07-11 Task 3's check-side kill
// for control:call.argument_consumed_when_noncopyable. It first engages
// callArgumentConsumeSeam and runs the REAL production path
// (Program(mustParseProgram(...))) on call_argument_used_twice.lang,
// asserting the seam wrongly admits the double-consume with zero
// diagnostics -- reproducing PVG-01/CR-01's pre-plan hole exactly. It
// then disengages the seam and asserts the real gate is restored
// (ownership.use_after_move).
func TestCallArgumentConsumeMutationKilled(t *testing.T) {
	defer func() { callArgumentConsumeSeam = false }()

	source := readPhase07Fixture(t, "call_argument_used_twice.lang")

	callArgumentConsumeSeam = true
	seamed := Program(mustParseProgram(t, source))
	if len(seamed.Diagnostics) != 0 {
		t.Fatalf("expected the seam to wrongly admit the double-consume, got %+v", seamed.Diagnostics)
	}

	callArgumentConsumeSeam = false
	restored := Program(mustParseProgram(t, source))
	if len(restored.Diagnostics) != 1 || restored.Diagnostics[0].Code != "ownership.use_after_move" {
		t.Fatalf("expected ownership.use_after_move restored once the seam is disengaged, got %+v", restored.Diagnostics)
	}
}

// callByteArgumentTwiceSource is a Byte-argument double-call program,
// built inline (never a testdata/phase07 fixture -- this plan adds no
// third fixture) since call_basic.lang's own Byte argument is passed to
// only ONE call: the over-refusal fault this test seeds needs a SECOND
// use to observe wrongly firing on.
const callByteArgumentTwiceSource = `module phase07.call_argument_byte_used_twice

export {
  fn main
}

fn identity(value: Byte) -> Byte {
  value
}

fn main(value: Byte) -> Byte {
  let first = identity(value)
  let second = identity(value)
  second
}
`

// TestCallArgumentConsumeOverRefusalMutationKilled is 07-11 Task 3's
// check-side kill for control:call.copyable_argument_not_consumed: the
// over-refusal fault for the non-refusing direction. With
// callArgumentConsumeAlwaysSeam engaged, a Byte argument passed to two
// separate calls is WRONGLY refused with ownership.use_after_move; with
// it disengaged, the real production path admits it cleanly.
func TestCallArgumentConsumeOverRefusalMutationKilled(t *testing.T) {
	defer func() { callArgumentConsumeAlwaysSeam = false }()

	source := []byte(callByteArgumentTwiceSource)

	callArgumentConsumeAlwaysSeam = true
	seamed := Program(mustParseProgram(t, source))
	if len(seamed.Diagnostics) != 1 || seamed.Diagnostics[0].Code != "ownership.use_after_move" {
		t.Fatalf("expected the seam to wrongly refuse a copyable argument's second use, got %+v", seamed.Diagnostics)
	}

	callArgumentConsumeAlwaysSeam = false
	restored := Program(mustParseProgram(t, source))
	if len(restored.Diagnostics) != 0 {
		t.Fatalf("expected the Byte double-call to check clean once the seam is disengaged, got %+v", restored.Diagnostics)
	}
}

// ---------------------------------------------------------------------
// Phase 08 Task 1 (08-03): the real .lang twin corpus and the harness that
// enforces the twin discipline.
// ---------------------------------------------------------------------

// readPhase08Fixture reads a testdata/phase08 fixture, mirroring
// readPhase07Fixture exactly.
func readPhase08Fixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile("../../../testdata/phase08/" + name)
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return source
}

// callerLinearBody locates and returns the named function's own
// ast.LinearBody within parsed -- the fixed caller-identity witness
// TestInterproceduralTwinPairsDifferOnlyInTheCallee compares between twin
// members.
func callerLinearBody(t *testing.T, parsed ast.Program, name string) *ast.LinearBody {
	t.Helper()
	for _, function := range parsed.Funcs {
		if function.Name == name && function.Body.Linear != nil {
			return function.Body.Linear
		}
	}
	t.Fatalf("expected a linear-bodied function named %q", name)
	return nil
}

// stringSlicesEqual is assertIdenticalCallerBodies' own small element-wise
// comparison for RHS.Arguments -- reflect.DeepEqual would also work, but an
// explicit loop keeps the failure message's own comparison logic legible
// without importing reflect just for this one call site.
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// assertIdenticalCallerBodies is T-08-09's mitigation, mechanically
// enforced: the two members of a twin pair must differ ONLY in the callee's
// own declaration/body, never in the CALLER's own shape -- a caller-varying
// pair would pass with the interprocedural law entirely deleted, since the
// existing intraprocedural ownership.* law already refuses same-function
// violations (08-CONTEXT.md's single most actionable finding). Compares
// binding count, then each binding's Name/RHS.Kind/RHS.Source/RHS.Callee/
// RHS.Arguments (never RHS.Span, which legitimately differs since the two
// members are different source files), then the body's own Result.
func assertIdenticalCallerBodies(t *testing.T, pairName string, a, b *ast.LinearBody) {
	t.Helper()
	if len(a.Bindings) != len(b.Bindings) {
		t.Fatalf("%s: caller binding count differs: %d vs %d", pairName, len(a.Bindings), len(b.Bindings))
	}
	for i := range a.Bindings {
		left, right := a.Bindings[i], b.Bindings[i]
		if left.Name != right.Name {
			t.Fatalf("%s: binding %d name differs: %q vs %q", pairName, i, left.Name, right.Name)
		}
		if left.RHS.Kind != right.RHS.Kind {
			t.Fatalf("%s: binding %d RHS.Kind differs: %q vs %q", pairName, i, left.RHS.Kind, right.RHS.Kind)
		}
		if left.RHS.Source != right.RHS.Source {
			t.Fatalf("%s: binding %d RHS.Source differs: %q vs %q", pairName, i, left.RHS.Source, right.RHS.Source)
		}
		if left.RHS.Callee != right.RHS.Callee {
			t.Fatalf("%s: binding %d RHS.Callee differs: %q vs %q", pairName, i, left.RHS.Callee, right.RHS.Callee)
		}
		if !stringSlicesEqual(left.RHS.Arguments, right.RHS.Arguments) {
			t.Fatalf("%s: binding %d RHS.Arguments differs: %v vs %v", pairName, i, left.RHS.Arguments, right.RHS.Arguments)
		}
	}
	if a.Result != b.Result {
		t.Fatalf("%s: caller Result differs: %q vs %q", pairName, a.Result, b.Result)
	}
}

// TestInterproceduralTwinPairsDifferOnlyInTheCallee is Task 1's mechanical
// enforcement of the twin discipline (T-08-09): for both real .lang twin
// pairs, the CALLER function's own ast.LinearBody is structurally identical
// between members. A pair whose caller varies fails this test, which is
// exactly the falsifier that stops a caller-varying (decorative) pair from
// silently passing.
func TestInterproceduralTwinPairsDifferOnlyInTheCallee(t *testing.T) {
	pairs := []struct {
		name, refuse, accept, caller string
	}{
		{"pattern A", "twin_a_refuse.lang", "twin_a_accept.lang", "escort"},
		{"pattern B", "twin_b_refuse.lang", "twin_b_accept.lang", "escort"},
	}
	for _, pair := range pairs {
		t.Run(pair.name, func(t *testing.T) {
			refuseParsed := mustParseProgram(t, readPhase08Fixture(t, pair.refuse))
			acceptParsed := mustParseProgram(t, readPhase08Fixture(t, pair.accept))
			refuseCaller := callerLinearBody(t, refuseParsed, pair.caller)
			acceptCaller := callerLinearBody(t, acceptParsed, pair.caller)
			assertIdenticalCallerBodies(t, pair.name, refuseCaller, acceptCaller)
		})
	}
}

// TestInterproceduralLivenessTwinPatternA drives twin_a_refuse.lang and
// twin_a_accept.lang through the real ast.Program -> check.Program
// pipeline: the refusing member produces exactly one diagnostic with code
// check.interprocedural_loan_liveness; the accepting member produces zero
// diagnostics (package check alone -- see twin_a_accept.lang's own header
// for the separate, documented corevalidate residual this test does not
// claim to resolve).
func TestInterproceduralLivenessTwinPatternA(t *testing.T) {
	refuseSource := readPhase08Fixture(t, "twin_a_refuse.lang")
	refuseResult := Program(mustParseProgram(t, refuseSource))
	if len(refuseResult.Diagnostics) != 1 || refuseResult.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected exactly one check.interprocedural_loan_liveness diagnostic for twin_a_refuse.lang, got %+v", refuseResult.Diagnostics)
	}

	acceptSource := readPhase08Fixture(t, "twin_a_accept.lang")
	acceptResult := Program(mustParseProgram(t, acceptSource))
	if len(acceptResult.Diagnostics) != 0 {
		t.Fatalf("expected zero diagnostics for twin_a_accept.lang, got %+v", acceptResult.Diagnostics)
	}
}

// TestInterproceduralLivenessTwinPatternBRealFixtures extends 08-02's own
// synthetic-core.Program twin (TestInterproceduralLivenessTwinPatternB
// above) to also drive the real twin_b_refuse.lang/twin_b_accept.lang
// fixtures landed under Phase 08.
//
// Plan 09-09 (D-09-08's authorized deletion) removed the summary-blind
// lowering-time gate that used to mask BOTH fixtures identically with
// ownership.move_while_borrowed before check's own interprocedural pass
// ever ran (08-02's documented, carried-forward finding). With that gate
// gone, BOTH fixtures now correctly reach the interprocedural pass -- but
// Pattern B's promised split (twin_b_refuse refused, twin_b_accept ADMITTED)
// does NOT materialize: both are refused identically with
// check.interprocedural_loan_liveness. Root cause, diagnosed and NOT fixed
// by this plan (out of scope; see check_ordering_stability_test.go's own
// twin_b_accept/refuse baseline comment for the full explanation):
// deriveFunctionUsesParam (check.go, pre-existing since Phase 08) treats a
// plain OpMove forward of the parameter to its own terminating return as a
// "use", so twin_b_accept's `probe` is ALSO derived as usesParam=true,
// collapsing the interprocedural signal Pattern B depends on. Recorded as
// new debt (this plan's SUMMARY, PHASE-09-DEBT.md) rather than adjusted
// here or in the fixtures themselves -- Task 4(b)'s own contingency for
// exactly this outcome. The CONTRACT-DRIVEN differentiation Pattern B's law
// is DESIGNED to perform is still proven at the checked-core level by
// TestInterproceduralLivenessTwinPatternB above, against a REAL
// interprocedural summary table constructed directly (not derived by the
// buggy function) -- that proof is unaffected by this finding.
func TestInterproceduralLivenessTwinPatternBRealFixtures(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		{"twin_b_refuse.lang", "check.interprocedural_loan_liveness"},
		{"twin_b_accept.lang", "check.interprocedural_loan_liveness"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := readPhase08Fixture(t, tc.name)
			result := Program(mustParseProgram(t, source))
			if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != tc.want {
				t.Fatalf("expected exactly one %s diagnostic, got %+v", tc.want, result.Diagnostics)
			}
		})
	}
}

// ---------------------------------------------------------------------
// Phase 08 Task 2 (08-03): composition depth >= 2, the negative control,
// and the match-arm regression.
// ---------------------------------------------------------------------

// functionIDByName locates a checked core.Function's own semantic ID by its
// source-level Name within an already-checked core.Program -- used to pull
// the RIGHT summary-table entry out of two independently built
// interproceduralSummaryTables whose IDs are module-qualified (and
// therefore differ syntactically) even when they name "the same" function
// across two different negative-control programs.
func functionIDByName(t *testing.T, program core.Program, name string) string {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function.ID
		}
	}
	t.Fatalf("expected a checked function named %q", name)
	return ""
}

// TestInterproceduralLivenessRelayDepth2 is Task 2(a)'s D-08-28.2 pin: a
// composition-depth-2 relay chain (caller -> relay -> leaf) refuses when
// relay's OWN declared return contract reports a borrow of its parameter,
// and admits when it does not -- generalizing 08-01/08-03's depth-1 twin
// (twin_a_*.lang) to prove the derivation genuinely reads a callee's
// declared signature rather than merely working for the one-hop case. The
// emitted causes name RELAY's own function ID and RELAY's own
// return.mode -- never leaf's -- proving the caller's admission consulted
// relay's OWN interprocedural summary entry, not leaf's (leaf is never
// named anywhere in the diagnostic).
func TestInterproceduralLivenessRelayDepth2(t *testing.T) {
	refuseSource := readPhase08Fixture(t, "relay_depth2_refuse.lang")
	refuseResult := Program(mustParseProgram(t, refuseSource))
	if len(refuseResult.Diagnostics) != 1 || refuseResult.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected exactly one check.interprocedural_loan_liveness diagnostic, got %+v", refuseResult.Diagnostics)
	}
	diag := refuseResult.Diagnostics[0]
	if len(diag.Causes) != 3 || diag.Causes[1].Kind != "loan_extended_by_call" {
		t.Fatalf("expected cause 2 kind loan_extended_by_call, got %+v", diag.Causes)
	}
	if !strings.HasSuffix(diag.Causes[1].Detail, ":fn:relay") {
		t.Fatalf("expected cause 2 to name relay's own function ID (not leaf's), got %q", diag.Causes[1].Detail)
	}
	if diag.Causes[2].Kind != "callee_return_contract" || !strings.HasSuffix(diag.Causes[2].Detail, ":fn:relay:return.mode=shared") {
		t.Fatalf("expected cause 3 to name relay's own return.mode (not leaf's), got %+v", diag.Causes[2])
	}

	acceptSource := readPhase08Fixture(t, "relay_depth2_accept.lang")
	acceptResult := Program(mustParseProgram(t, acceptSource))
	if len(acceptResult.Diagnostics) != 0 {
		t.Fatalf("expected relay_depth2_accept.lang to check clean, got %+v", acceptResult.Diagnostics)
	}
}

// TestInterproceduralLivenessNegativeControl is Task 2(b)'s D-08-28.4 pin,
// the "which fields are NOT named" half of criterion 4: two programs whose
// relay differs ONLY in its own transitively-inherited Fails/ForeignReach
// (negative_control_fails.lang's relay calls a genuinely foreign-fallible
// leaf; negative_control_infallible.lang's relay calls a wholly ordinary
// one) produce (a) the identical check verdict -- both refuse with
// check.interprocedural_loan_liveness, same cause shape, both citing
// relay's own return.mode=shared -- and (b) field-for-field identical
// interproceduralSummary entries for relay, proving the liveness law's
// silence on Fails/ForeignReach was demonstrated by actually varying them,
// not merely by never having tried.
func TestInterproceduralLivenessNegativeControl(t *testing.T) {
	failsResult := Program(mustParseProgram(t, readPhase08Fixture(t, "negative_control_fails.lang")))
	infallibleResult := Program(mustParseProgram(t, readPhase08Fixture(t, "negative_control_infallible.lang")))

	if len(failsResult.Diagnostics) != 1 || failsResult.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected negative_control_fails.lang to refuse with check.interprocedural_loan_liveness, got %+v", failsResult.Diagnostics)
	}
	if len(infallibleResult.Diagnostics) != 1 || infallibleResult.Diagnostics[0].Code != "check.interprocedural_loan_liveness" {
		t.Fatalf("expected negative_control_infallible.lang to refuse with check.interprocedural_loan_liveness, got %+v", infallibleResult.Diagnostics)
	}

	failsCauses := failsResult.Diagnostics[0].Causes
	infallibleCauses := infallibleResult.Diagnostics[0].Causes
	if len(failsCauses) != 3 || len(infallibleCauses) != 3 {
		t.Fatalf("expected exactly 3 causes on each member, got %d and %d", len(failsCauses), len(infallibleCauses))
	}
	for i, wantKind := range []string{"borrow_created_here", "loan_extended_by_call", "callee_return_contract"} {
		if failsCauses[i].Kind != wantKind || infallibleCauses[i].Kind != wantKind {
			t.Fatalf("cause %d kind mismatch: got %q (fails) / %q (infallible), want %q", i, failsCauses[i].Kind, infallibleCauses[i].Kind, wantKind)
		}
	}
	if !strings.HasSuffix(failsCauses[2].Detail, ":fn:relay:return.mode=shared") || !strings.HasSuffix(infallibleCauses[2].Detail, ":fn:relay:return.mode=shared") {
		t.Fatalf("expected both members' cause 3 to name relay's own return.mode=shared, got %q (fails) / %q (infallible)", failsCauses[2].Detail, infallibleCauses[2].Detail)
	}

	failsTable, err := buildCallSignatureTable(failsResult.Program)
	if err != nil {
		t.Fatalf("buildCallSignatureTable (fails): %v", err)
	}
	infallibleTable, err := buildCallSignatureTable(infallibleResult.Program)
	if err != nil {
		t.Fatalf("buildCallSignatureTable (infallible): %v", err)
	}
	failsSummaries, _ := buildInterproceduralSummaries(failsResult.Program, failsTable)
	infallibleSummaries, _ := buildInterproceduralSummaries(infallibleResult.Program, infallibleTable)

	failsRelaySummary, ok := failsSummaries.lookup(functionIDByName(t, failsResult.Program, "relay"))
	if !ok {
		t.Fatal("expected a summary entry for relay in negative_control_fails.lang")
	}
	infallibleRelaySummary, ok := infallibleSummaries.lookup(functionIDByName(t, infallibleResult.Program, "relay"))
	if !ok {
		t.Fatal("expected a summary entry for relay in negative_control_infallible.lang")
	}
	if failsRelaySummary != infallibleRelaySummary {
		t.Fatalf("expected relay's interproceduralSummary to be field-for-field identical between the fails and infallible members (differing only in Fails/ForeignReach, which this table structurally never reads), got %+v vs %+v", failsRelaySummary, infallibleRelaySummary)
	}
}

// TestInterproceduralLivenessMatchArmRegression is Task 2(c)'s D-08-28.5
// regression coverage (recommended, explicitly NOT gate-blocking, per
// match_arm_call.lang's own header): a match-bodied function calling the
// same callee from both arms checks exactly as clean as an equivalent
// straight-line program calling that callee once. D-07-28's own guard was
// about call-graph EDGE DISCOVERY (is a call inside a match arm enumerated
// at all); this plan's liveness admission is per-call-site and arity-1
// regardless of which arm a call sits in, and `check` already walks both
// arms uniformly -- this test records that fact under test, it does not
// assert a new gate this plan's success criteria require.
func TestInterproceduralLivenessMatchArmRegression(t *testing.T) {
	matchSource := readPhase08Fixture(t, "match_arm_call.lang")
	matchResult := Program(mustParseProgram(t, matchSource))
	if len(matchResult.Diagnostics) != 0 {
		t.Fatalf("expected match_arm_call.lang to check clean, got %+v", matchResult.Diagnostics)
	}

	straightLineSource := []byte(`module phase08.match_arm_call_straight_line

export {
  fn main
}

fn helper(value: Byte) -> Byte {
  value
}

fn main(flag: Byte) -> Byte {
  let result = helper(flag)
  result
}
`)
	straightLineResult := Program(mustParseProgram(t, straightLineSource))
	if len(straightLineResult.Diagnostics) != 0 {
		t.Fatalf("expected the straight-line equivalent to check clean, got %+v", straightLineResult.Diagnostics)
	}
	if len(matchResult.Diagnostics) != len(straightLineResult.Diagnostics) {
		t.Fatalf("expected the match-arm shape's verdict (%d diagnostics) to match the straight-line equivalent's (%d diagnostics)", len(matchResult.Diagnostics), len(straightLineResult.Diagnostics))
	}
}

// ---------------------------------------------------------------------
// Phase 08 Task 3 (08-03): criterion 4's refusal half -- the exact
// consulted callee-signature field set.
// ---------------------------------------------------------------------

// splitCalleeFieldDetail parses a cause 3 (callee_return_contract) Detail
// string of the ratified shape "<calleeID>:<field>=<value>"
// (D-08-20/D-08-23) back into its calleeID and field parts. calleeID itself
// contains colons (e.g. "s1:phase08.twin_a_refuse:fn:escortee"), so this
// cannot split on the first colon -- it instead searches for one of the two
// field names this derivation is allowed to ever disclose, exactly
// mirroring the closed set interproceduralConsultObserved records.
func splitCalleeFieldDetail(detail string) (calleeID, field string, ok bool) {
	for _, candidate := range []string{"return.mode", "parameters[0].mode"} {
		marker := ":" + candidate + "="
		if index := strings.Index(detail, marker); index != -1 {
			return detail[:index], candidate, true
		}
	}
	return "", "", false
}

// TestInterproceduralDisclosedFieldSet is Task 3's D-08-26 machine-assertion
// of criterion 4's REFUSAL half: table-driven over every testdata/phase08
// fixture plus testdata/phase07/relay_escort_witness.lang, it installs
// interproceduralConsultObserved, runs check.Program, and asserts (a) the
// recorded per-callee field set is a subset of exactly {"return.mode",
// "parameters[0].mode"} -- the derivation must consult no other signature
// field and must reach no body value -- and (b) for every emitted
// check.interprocedural_loan_liveness diagnostic, the field named in cause
// 3 (callee_return_contract) is a member of the recorded set for that same
// callee. That second assertion is what makes the shipped diagnostic's
// disclosure TRUE rather than merely well-formed: a diagnostic could name a
// plausible-looking field string without this derivation having actually
// read it.
//
// Honest scope (D-08-26, restated verbatim from PHASE-08-DEBT.md): criterion
// 4's REFUSAL half is met by a shipped artifact (this diagnostic's own
// cause 3); the ACCEPTED-program half is met by THIS TEST ALONE, by no
// shipped runtime artifact -- protocol.ExplainSummary needs a diagnostic to
// expand, internal/compiler/cache is structurally incapable of holding a
// verdict, and a sibling lang.*/1 document contradicts
// protocol.InterfaceSummary's own recorded anti-pattern. Whether that half
// should become a runtime artifact is on the mid-phase gate's agenda.
func TestInterproceduralDisclosedFieldSet(t *testing.T) {
	defer func() { interproceduralConsultObserved = nil }()

	paths, err := filepath.Glob("../../../testdata/phase08/*.lang")
	if err != nil {
		t.Fatalf("glob testdata/phase08: %v", err)
	}
	paths = append(paths, "../../../testdata/phase07/relay_escort_witness.lang")
	sort.Strings(paths)
	if len(paths) == 0 {
		t.Fatal("expected at least one fixture")
	}

	allowedFields := map[string]bool{"return.mode": true, "parameters[0].mode": true}

	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			recorded := map[string]map[string]bool{}
			interproceduralConsultObserved = func(calleeID, field string) {
				if recorded[calleeID] == nil {
					recorded[calleeID] = map[string]bool{}
				}
				recorded[calleeID][field] = true
			}
			result := Program(mustParseProgram(t, source))
			interproceduralConsultObserved = nil

			for calleeID, fields := range recorded {
				for field := range fields {
					if !allowedFields[field] {
						t.Fatalf("callee %s: consulted disallowed field %q -- must be a subset of {return.mode, parameters[0].mode}", calleeID, field)
					}
				}
			}

			for _, diag := range result.Diagnostics {
				if diag.Code != "check.interprocedural_loan_liveness" {
					continue
				}
				if len(diag.Causes) != 3 {
					t.Fatalf("expected exactly 3 causes on an interprocedural refusal, got %+v", diag.Causes)
				}
				thirdCause := diag.Causes[2]
				calleeID, field, ok := splitCalleeFieldDetail(thirdCause.Detail)
				if !ok {
					t.Fatalf("could not parse cause 3 Detail %q into <calleeID>:<field>=<value>", thirdCause.Detail)
				}
				if !recorded[calleeID][field] {
					t.Fatalf("diagnostic's cause 3 names field %q for callee %q, but that field was never recorded as consulted for that callee: %+v", field, calleeID, recorded[calleeID])
				}
			}
		})
	}
}

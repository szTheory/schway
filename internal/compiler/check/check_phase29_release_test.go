package check

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// phase29ForgeMain deep-copies the checked fixture main through its canonical
// JSON form and hands the copy to mutate. The checked program is never edited in
// place and no production code is touched.
func phase29ForgeMain(t *testing.T, mutate func(f *core.Function)) core.Function {
	t.Helper()
	program, codes := phase29CheckSource(t, []byte(phase29FixtureSource(t)))
	if len(codes) != 0 {
		t.Fatalf("fixture diagnostics: %v", codes)
	}
	raw, err := json.Marshal(phase29FixtureMain(t, program))
	if err != nil {
		t.Fatal(err)
	}
	var function core.Function
	if err := json.Unmarshal(raw, &function); err != nil {
		t.Fatal(err)
	}
	mutate(&function)
	return function
}

// phase29EarlyRelease moves the fixture's OpRelease before its OpCall, keeping
// the operation IDs attached to their operations.
func phase29EarlyRelease(f *core.Function) {
	ops := f.Linear.Operations
	ops[2], ops[3] = ops[3], ops[2]
}

// TestPhase29ReleaseBeforeViewUseRefused is D-22's seeded fault for check: the
// fixture core with its OpRelease moved before its OpCall is refused with
// check.release_while_view_live; with only that law disabled the observed result
// is the closed-shape refusal instead, so the law is load-bearing.
func TestPhase29ReleaseBeforeViewUseRefused(t *testing.T) {
	if diagnostics := checkFixtureOwnerPostAssembly(phase29ForgeMain(t, func(*core.Function) {})); len(diagnostics) != 0 {
		t.Fatalf("unforged fixture refused: %+v", diagnostics)
	}
	forged := phase29ForgeMain(t, phase29EarlyRelease)
	diagnostics := checkFixtureOwnerPostAssembly(forged)
	if len(diagnostics) == 0 || diagnostics[0].Code != "check.release_while_view_live" {
		t.Fatalf("diagnostics = %+v, want first code check.release_while_view_live", diagnostics)
	}
	causes := map[string]string{}
	for _, cause := range diagnostics[0].Causes {
		causes[cause.Kind] = cause.Detail
	}
	if causes["release"] != forged.Linear.Operations[2].ID || causes["view_read_later"] != forged.Linear.Operations[3].ID {
		t.Fatalf("causes = %v, want the release and the later read operation IDs", causes)
	}
	disableReleaseWhileViewLiveLawForTest = true
	defer func() { disableReleaseWhileViewLiveLawForTest = false }()
	diagnostics = checkFixtureOwnerPostAssembly(forged)
	if len(diagnostics) == 0 || diagnostics[0].Code != "check.fixture_owner_shape_unsupported" {
		t.Fatalf("law disabled: diagnostics = %+v, want the closed-shape code", diagnostics)
	}
}

// phase29RenumberOperations keeps operation and point IDs in position order.
func phase29RenumberOperations(f *core.Function) {
	for index := range f.Linear.Operations {
		f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
		f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
	}
}

// phase29MoveOwner forges an OpMove of the owner into a fresh FixtureBytes place
// between the view borrow and the OpCall. loanView rewrites the view borrow as an
// OpBorrowShared carrying a loan, the only form in which the existing
// move-while-borrowed laws can see the view; copyRead then reads the view by a
// plain OpCopy instead of the OpCall.
func phase29MoveOwner(loanView, copyRead bool) func(f *core.Function) {
	return func(f *core.Function) {
		ops := f.Linear.Operations
		owner := ops[0].TargetID
		moved := f.ID + ":place:3"
		sum := f.ID + ":place:4"
		f.Linear.Places[3].ID = sum
		f.Linear.Places = append(f.Linear.Places[:3], core.Place{ID: moved, Name: "moved", TypeID: f.Linear.Types[1].ID}, f.Linear.Places[3])
		view := ops[1]
		if loanView {
			view.Kind = core.OpBorrowShared
			view.Foreign = nil
			view.LoanID = f.ID + ":loan:1"
			// ByteView holds no share ability, so the loan-form borrow is typed as
			// the shareable owner type.
			view.TypeID = f.Linear.Types[1].ID
			f.Linear.Places[2].TypeID = f.Linear.Types[1].ID
		}
		move := core.LinearOperation{Kind: core.OpMove, SourceID: owner, TargetID: moved, TypeID: f.Linear.Types[1].ID}
		call, release, ret := ops[2], ops[3], ops[4]
		call.TargetID, ret.SourceID = sum, sum
		if copyRead {
			// the view is read by a plain copy, so the loan is not extended through a call
			call.Kind, call.CalleeID = core.OpCopy, ""
		}
		release.SourceID = moved
		f.Linear.Operations = []core.LinearOperation{ops[0], view, move, call, release, ret}
		phase29RenumberOperations(f)
	}
}

// TestPhase29MoveOwnerWhileViewLive forges a move of the owner while the view is
// still read by the later OpCall. No new code is introduced for moves (D-22).
// The fixture view is a foreign borrow, not a loan, so ownership.move_while_borrowed
// does not see it: the fixture-shaped forgery is refused by check's closed
// post-assembly shape (check.fixture_owner_shape_unsupported). The loan-form rows
// (the view borrowed as OpBorrowShared) are refused by the existing codes through
// check's post-assembly loan pass: check.interprocedural_loan_liveness when the
// OpCall extends the loan, "ownership.move_while_borrowed" when a plain copy reads
// it. The D-22 law is not what refuses a move: toggling it changes no row.
func TestPhase29MoveOwnerWhileViewLive(t *testing.T) {
	if diagnostics := checkFixtureOwnerPostAssembly(phase29ForgeMain(t, func(*core.Function) {})); len(diagnostics) != 0 {
		t.Fatalf("accepted fixture core refused: %+v", diagnostics)
	}
	loanForm := func(copyRead bool) []diagnostic.Diagnostic {
		program, codes := phase29CheckSource(t, []byte(phase29FixtureSource(t)))
		if len(codes) != 0 {
			t.Fatalf("fixture diagnostics: %v", codes)
		}
		for index := range program.Functions {
			if program.Functions[index].Name == "main" {
				program.Functions[index] = phase29ForgeMain(t, phase29MoveOwner(true, copyRead))
			}
		}
		table, err := buildCallSignatureTable(program)
		if err != nil {
			t.Fatal(err)
		}
		summaries, _ := buildInterproceduralSummaries(program, table)
		return checkInterproceduralLoanLiveness(program, summaries, map[string]diagnostic.Span{})
	}
	for _, disabled := range []bool{false, true} {
		disableReleaseWhileViewLiveLawForTest = disabled
		shaped := checkFixtureOwnerPostAssembly(phase29ForgeMain(t, phase29MoveOwner(false, false)))
		called, copied := loanForm(false), loanForm(true)
		disableReleaseWhileViewLiveLawForTest = false
		if len(shaped) == 0 || shaped[0].Code != "check.fixture_owner_shape_unsupported" {
			t.Fatalf("law disabled=%v: fixture-shaped move diagnostics = %+v", disabled, shaped)
		}
		if len(called) == 0 || called[0].Code != "check.interprocedural_loan_liveness" {
			t.Fatalf("law disabled=%v: loan-form (call) move diagnostics = %+v", disabled, called)
		}
		if len(copied) == 0 || copied[0].Code != "ownership.move_while_borrowed" {
			t.Fatalf("law disabled=%v: loan-form (copy) move diagnostics = %+v", disabled, copied)
		}
	}
}

// phase29FirstDiagnostic parses source and returns the first diagnostic code and
// start offset, falling back to the checker when parsing succeeds.
func phase29FirstDiagnostic(t *testing.T, source string) (string, int) {
	t.Helper()
	parsed := syntax.Parse([]byte(source))
	if len(parsed.Diagnostics) != 0 {
		return parsed.Diagnostics[0].Code, parsed.Diagnostics[0].Primary.Start
	}
	checked := Program(parsed.Program)
	if len(checked.Diagnostics) == 0 {
		return "", -1
	}
	return checked.Diagnostics[0].Code, checked.Diagnostics[0].Primary.Start
}

// TestPhase29LoopNeighborParseRefusals pins the source-level owner neighbors.
// Every `while` row is a PARSE refusal (D-22), not an ownership proof: the main
// body parser (linearBody) admits no `while`, so the first diagnostic points at
// the `while` token and the checker never sees the loop. The seeded typed-core rows
// of TestPhase29OwnerLoopBlockRefused are the ownership-law proof. Release is not
// user-spellable, so the straight-line explicit release is also a parse refusal,
// and a straight-line `let moved = owner` is a checker refusal by the closed shape.
func TestPhase29LoopNeighborParseRefusals(t *testing.T) {
	base := phase29FixtureSource(t)
	tail := "  let sum = adler32(view)\n  sum\n}\n"
	if !strings.Contains(base, tail) {
		t.Fatal("fixture main tail changed")
	}
	if _, codes := phase29CheckSource(t, []byte(base)); len(codes) != 0 {
		t.Fatalf("accepted fixture control diagnostics: %v", codes)
	}
	cases := []struct {
		name   string
		kind   string
		insert string
		code   string
		at     string // text whose first occurrence the diagnostic must start at; "" skips
	}{
		{"move owner in a while body", "parse refusal (D-22)", "  while case < 1 {\n    let moved = owner\n  }\n", "syntax.expected_linear_result", "while case"},
		{"explicit release in a while body", "parse refusal (D-22)", "  while case < 1 {\n    schway_fixture_bytes_release(owner)\n  }\n", "syntax.expected_linear_result", "while case"},
		{"explicit release in straight-line main (release is not user-spellable)", "parse refusal (D-22)", "  schway_fixture_bytes_release(owner)\n", "syntax.expected_rbrace", "(owner)\n  let sum"},
		{"move owner in straight-line main", "checker refusal", "  let moved = owner\n", "check.fixture_owner_shape_unsupported", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := strings.Replace(base, tail, tc.insert+tail, 1)
			if source == base {
				t.Fatal("insertion did not apply")
			}
			code, start := phase29FirstDiagnostic(t, source)
			if code != tc.code {
				t.Fatalf("%s: first diagnostic = %q, want %q", tc.kind, code, tc.code)
			}
			if tc.at != "" && !strings.HasPrefix(source[start:], strings.SplitN(tc.at, "\n", 2)[0]) {
				t.Fatalf("diagnostic starts at %q, want the text %q", source[start:start+12], tc.at)
			}
		})
	}
}

// phase29LoopRow rebuilds the checked fixture main as a loop-shaped typed core
// (D-22): entry, loop header, loop body closed by an OpBranch back edge to the
// header, exit. row R1 releases the owner in the body between the view borrow and
// the OpCall; R2 releases it in the body after the OpCall (view dead, release on
// the cycle); M1 moves the owner in the body between the view borrow and the
// OpCall. Operation IDs follow program order.
func phase29LoopRow(row string) func(f *core.Function) {
	return func(f *core.Function) {
		ops := f.Linear.Operations
		acquire, view, call, release, ret := ops[0], ops[1], ops[2], ops[3], ops[4]
		types := f.Linear.Types
		boolType := f.ID + ":type:3"
		f.Linear.Types = append(types, core.TypeFact{ID: boolType, Shape: core.TypeRef{Constructor: "Bool"}, Abilities: types[0].Abilities, NegativeWitnesses: types[0].NegativeWitnesses})
		cond1, cond2, moved := f.ID+":place:4", f.ID+":place:5", f.ID+":place:6"
		f.Linear.Places = append(f.Linear.Places,
			core.Place{ID: cond1, Name: "c1", TypeID: boolType},
			core.Place{ID: cond2, Name: "c2", TypeID: boolType},
			core.Place{ID: moved, Name: "moved", TypeID: types[1].ID})
		param := f.Parameter.ID
		eq := func(target string) core.LinearOperation {
			return core.LinearOperation{Kind: core.OpEqualU64, SourceID: param, RightID: param, TargetID: target, TypeID: boolType}
		}
		edge := func(name, from, to, pattern string) core.Edge {
			return core.Edge{ID: f.ID + ":edge:" + name, FromBlockID: f.ID + ":block:" + from, ToBlockID: f.ID + ":block:" + to, Pattern: pattern}
		}
		branch := func(cond, trueEdge, falseEdge string) core.LinearOperation {
			return core.LinearOperation{Kind: core.OpBranch, SourceID: cond, TypeID: boolType, TrueEdgeID: f.ID + ":edge:" + trueEdge, FalseEdgeID: f.ID + ":edge:" + falseEdge}
		}
		ownerOp := release
		if row == "M1" {
			ownerOp = core.LinearOperation{Kind: core.OpMove, SourceID: acquire.TargetID, TargetID: moved, TypeID: types[1].ID}
		}
		var flat []core.LinearOperation
		var groups [][]int
		if row == "R2" {
			flat = []core.LinearOperation{acquire, view, call, eq(cond1), branch(cond1, "hb", "hx"), ownerOp, eq(cond2), branch(cond2, "bh", "bx"), ret}
			groups = [][]int{{0, 1, 2}, {3, 4}, {5, 6, 7}, {8}}
		} else {
			flat = []core.LinearOperation{acquire, view, eq(cond1), branch(cond1, "hb", "hx"), ownerOp, eq(cond2), branch(cond2, "bh", "bx"), call, ret}
			groups = [][]int{{0, 1}, {2, 3}, {4, 5, 6}, {7, 8}}
		}
		f.Linear.Operations = flat
		phase29RenumberOperations(f)
		names := []string{"entry", "header", "body", "exit"}
		successors := [][]string{{"header"}, {"body", "exit"}, {"header", "exit"}, {}}
		f.Linear.Blocks = nil
		for index, group := range groups {
			block := core.Block{ID: f.ID + ":block:" + names[index], PointID: f.ID + ":point:block:" + names[index], Successors: []string{}}
			for _, op := range group {
				block.OperationIDs = append(block.OperationIDs, f.Linear.Operations[op].ID)
			}
			for _, successor := range successors[index] {
				block.Successors = append(block.Successors, f.ID+":block:"+successor)
			}
			f.Linear.Blocks = append(f.Linear.Blocks, block)
		}
		f.Linear.Edges = []core.Edge{edge("hb", "header", "body", "true"), edge("hx", "header", "exit", "false"), edge("bh", "body", "header", "true"), edge("bx", "body", "exit", "false")}
	}
}

// TestPhase29OwnerLoopBlockRefused pins owner release and owner move placed in a
// loop-body block closed by an OpBranch back edge (D-22, BYTE-03). planBaseCode is
// the first code characterized at PLAN_BASE 677e04b in a scratch worktree through
// the only check entry that existed there, loanLivenessFixpoint over
// cfgBlocksForFunction; headCode is the first code of checkFixtureOwnerPostAssembly
// at HEAD. R1 releases the owner before the view's last read, so no PLAN_BASE code
// named the release: the D-22 law now fires first and the row carries
// amendment D-22. R2 and M1 are refused at PLAN_BASE by an existing code and keep it
// (D-24 did not trigger: no peer accepted R2 or M1 at PLAN_BASE).
func TestPhase29OwnerLoopBlockRefused(t *testing.T) {
	cases := []struct {
		name         string
		row          string
		planBaseCode string
		headCode     string
		amendment    string
	}{
		{"R1 release between view borrow and call", "R1", "check.cfg_back_edge", "check.release_while_view_live", "D-22"},
		{"R2 release after the call, on the cycle", "R2", "check.cfg_back_edge", "check.cfg_back_edge", ""},
		{"M1 move between view borrow and call", "M1", "check.cfg_back_edge", "check.cfg_back_edge", ""},
	}
	first := func(row string) string {
		diagnostics := checkFixtureOwnerPostAssembly(phase29ForgeMain(t, phase29LoopRow(row)))
		if len(diagnostics) == 0 {
			return ""
		}
		return diagnostics[0].Code
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.amendment == "" && tc.headCode != tc.planBaseCode {
				t.Fatalf("row without an amendment must keep its PLAN_BASE code")
			}
			if got := first(tc.row); got != tc.headCode {
				t.Fatalf("HEAD first code = %q, want %q", got, tc.headCode)
			}
			disableReleaseWhileViewLiveLawForTest = true
			defer func() { disableReleaseWhileViewLiveLawForTest = false }()
			if got := first(tc.row); got != tc.planBaseCode {
				t.Fatalf("law disabled: first code = %q, want the PLAN_BASE code %q", got, tc.planBaseCode)
			}
		})
	}
}

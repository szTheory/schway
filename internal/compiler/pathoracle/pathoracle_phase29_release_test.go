package pathoracle_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
)

// phase29EarlyRelease moves the fixture's OpRelease before its OpCall.
func phase29EarlyRelease(f *core.Function) {
	ops := f.Linear.Operations
	ops[2], ops[3] = ops[3], ops[2]
}

// TestPhase29ReleaseBeforeViewUseRefused is D-22's seeded fault for pathoracle:
// the fixture core with its OpRelease before its OpCall is refused with
// pathoracle.release_while_view_live; with only that law disabled the observed
// code is the closed replay's own pathoracle.fixture_owner_order, so the law is
// load-bearing.
func TestPhase29ReleaseBeforeViewUseRefused(t *testing.T) {
	if err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, func(*core.Function) {})); err != nil {
		t.Fatalf("unforged fixture refused: %v", err)
	}
	err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, phase29EarlyRelease))
	if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.release_while_view_live:") {
		t.Fatalf("error = %v, want code pathoracle.release_while_view_live", err)
	}
	defer pathoracle.SetDisableReleaseWhileViewLiveLawForTest(true)()
	err = pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, phase29EarlyRelease))
	if err == nil || !strings.HasPrefix(err.Error(), "pathoracle.fixture_owner_order:") {
		t.Fatalf("law disabled: error = %v, want code pathoracle.fixture_owner_order", err)
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
		phase29RenumberLoopOps(f)
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

func phase29RenumberLoopOps(f *core.Function) {
	for index := range f.Linear.Operations {
		f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
		f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
	}
}

// TestPhase29OwnerLoopBlockRefused pins owner release placed in a loop-body block
// closed by an OpBranch back edge (D-22, BYTE-03); moves are not a pathoracle law,
// so there is no M1 row here. planBaseCode is the first code characterized at
// PLAN_BASE 677e04b in a scratch worktree through ValidateFixtureOwnerPaths;
// headCode is the first code at HEAD. R1 releases the owner before the view's last
// read: PLAN_BASE named only the closed shape, so the D-22 law now fires first and
// the row carries amendment D-22. R2 keeps the existing closed-shape code (D-24 did
// not trigger: R2 was refused at PLAN_BASE).
func TestPhase29OwnerLoopBlockRefused(t *testing.T) {
	cases := []struct {
		name         string
		row          string
		planBaseCode string
		headCode     string
		amendment    string
	}{
		{"R1 release between view borrow and call", "R1", "pathoracle.fixture_owner_shape", "pathoracle.release_while_view_live", "D-22"},
		{"R2 release after the call, on the cycle", "R2", "pathoracle.fixture_owner_shape", "pathoracle.fixture_owner_shape", ""},
	}
	first := func(row string) string {
		err := pathoracle.ValidateFixtureOwnerPaths(phase29FixtureForge(t, phase29LoopRow(row)))
		if err == nil {
			return ""
		}
		code, _, _ := strings.Cut(err.Error(), ":")
		return code
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.amendment == "" && tc.headCode != tc.planBaseCode {
				t.Fatalf("row without an amendment must keep its PLAN_BASE code")
			}
			if got := first(tc.row); got != tc.headCode {
				t.Fatalf("HEAD first code = %q, want %q", got, tc.headCode)
			}
			defer pathoracle.SetDisableReleaseWhileViewLiveLawForTest(true)()
			if got := first(tc.row); got != tc.planBaseCode {
				t.Fatalf("law disabled: first code = %q, want the PLAN_BASE code %q", got, tc.planBaseCode)
			}
		})
	}
}

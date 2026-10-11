package corevalidate_test

import (
	"fmt"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
)

// phase29Renumber keeps operation and point IDs in position order so the
// generic operation-order law passes and the forged shape reaches replay.
func phase29Renumber(f *core.Function) {
	for index := range f.Linear.Operations {
		f.Linear.Operations[index].ID = fmt.Sprintf("%s:op:%d", f.ID, index)
		f.Linear.Operations[index].PointID = fmt.Sprintf("%s:point:linear:%d", f.ID, index)
	}
}

// phase29EarlyRelease moves the fixture's OpRelease before its OpCall. The
// release keeps naming the acquire's operation ID (op:0), which renumbering
// leaves unchanged.
func phase29EarlyRelease(f *core.Function) {
	ops := f.Linear.Operations
	ops[2], ops[3] = ops[3], ops[2]
	phase29Renumber(f)
}

// TestPhase29ReleaseBeforeViewUseRefused is D-22's seeded fault for corevalidate:
// the fixture core with its OpRelease before its OpCall is refused with
// core.release_while_view_live; with only that law disabled the observed first
// code is the generic replay refusal core.invalid_target, so the law is
// load-bearing.
func TestPhase29ReleaseBeforeViewUseRefused(t *testing.T) {
	if result := corevalidate.Validate(phase29FixtureForge(t, func(*core.Function) {})); !result.Valid {
		t.Fatalf("unforged fixture refused: %+v", result.Problems)
	}
	result := corevalidate.Validate(phase29FixtureForge(t, phase29EarlyRelease))
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.release_while_view_live" {
		t.Fatalf("problems = %+v, want first code core.release_while_view_live", result.Problems)
	}
	defer corevalidate.SetDisableReleaseWhileViewLiveLawForTest(true)()
	result = corevalidate.Validate(phase29FixtureForge(t, phase29EarlyRelease))
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.invalid_target" {
		t.Fatalf("law disabled: problems = %+v, want first code core.invalid_target", result.Problems)
	}
}

// phase29MoveOwner forges an OpMove of the owner into a fresh FixtureBytes place
// between the view borrow and the OpCall. releaseOf selects which place the
// compiler-minted release then names ("owner" or "moved"); loanView rewrites the
// view borrow as an OpBorrowShared carrying a loan, the only form in which the
// existing move_while_borrowed law can see the view.
func phase29MoveOwner(releaseOf string, loanView bool) func(f *core.Function) {
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
			// the shareable owner type; the move below is refused before the call.
			view.TypeID = f.Linear.Types[1].ID
			f.Linear.Places[2].TypeID = f.Linear.Types[1].ID
		}
		move := core.LinearOperation{Kind: core.OpMove, SourceID: owner, TargetID: moved, TypeID: f.Linear.Types[1].ID}
		call, release, ret := ops[2], ops[3], ops[4]
		call.TargetID, ret.SourceID = sum, sum
		if releaseOf == "moved" {
			release.SourceID = moved
		}
		f.Linear.Operations = []core.LinearOperation{ops[0], view, move, call, release, ret}
		phase29Renumber(f)
	}
}

// TestPhase29MoveOwnerWhileViewLive forges a move of the owner while the view is
// still read by the later OpCall. No new code is introduced for moves (D-22). The
// fixture view is a foreign borrow, not a loan, so the existing
// core.move_while_borrowed law does not see it: the fixture-shaped forgeries are
// refused by the existing initialization law (core.place_uninitialized, release of
// the moved-out owner) and by the closed lifecycle matcher
// (core.fixture_owner_lifecycle, release of the moved place). The loan-form row
// (the view borrowed as OpBorrowShared) pins the existing core.move_while_borrowed.
func TestPhase29MoveOwnerWhileViewLive(t *testing.T) {
	cases := []struct {
		name  string
		kind  string
		forge func(f *core.Function)
		code  string
	}{
		{"fixture view, owner released after the move", "fixture-shaped", phase29MoveOwner("owner", false), "core.place_uninitialized"},
		{"fixture view, moved place released", "fixture-shaped", phase29MoveOwner("moved", false), "core.fixture_owner_lifecycle"},
		{"loan-form view", "loan-form", phase29MoveOwner("moved", true), "core.move_while_borrowed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := corevalidate.Validate(phase29FixtureForge(t, tc.forge))
			if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != tc.code {
				t.Fatalf("%s: problems = %+v, want first code %q", tc.kind, result.Problems, tc.code)
			}
			// The D-22 law is not what refuses a move: toggling it changes nothing.
			defer corevalidate.SetDisableReleaseWhileViewLiveLawForTest(true)()
			toggled := corevalidate.Validate(phase29FixtureForge(t, tc.forge))
			if toggled.Valid || len(toggled.Problems) == 0 || toggled.Problems[0].Code != tc.code {
				t.Fatalf("law disabled: problems = %+v, want the same first code %q", toggled.Problems, tc.code)
			}
		})
	}
}

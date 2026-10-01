package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase25PointerProgram(t *testing.T, dir, fixture string) core.Program {
	t.Helper()
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", dir, fixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	return cloneCoreProgram(t, checked.Program)
}

func TestPhase25PointerFamilyCore(t *testing.T) {
	for _, fixture := range []string{"shared_shared_accept.schway", "sequential_shared_then_exclusive_accept.schway"} {
		program := phase25PointerProgram(t, "phase3", fixture)
		if result := corevalidate.Validate(program); !result.Valid {
			t.Errorf("independent core peer rejected %s: %+v", fixture, result.Problems)
		}
	}

	shared := phase25PointerProgram(t, "phase25", "shared_copy_accept.schway")
	if result := corevalidate.Validate(shared); !result.Valid {
		t.Fatalf("independent core peer rejected shared read/copy: %+v", result.Problems)
	}
	sharedOp := &shared.Functions[0].Linear.Operations[0]
	if sharedOp.Kind != core.OpBorrowShared || sharedOp.LoanID == "" {
		t.Fatalf("fixture lacks an explicit shared loan: %+v", *sharedOp)
	}

	// The same checked places/types remain a valid exclusive read/copy core
	// shape.  Changing only the core family makes this a peer input, not a
	// checker classification or checker-produced authorization result.
	exclusive := phase25PointerProgram(t, "phase25", "shared_copy_accept.schway")
	exclusiveOp := &exclusive.Functions[0].Linear.Operations[0]
	exclusiveOp.Kind = core.OpBorrowExclusive
	if result := corevalidate.Validate(exclusive); !result.Valid {
		t.Fatalf("independent core peer rejected exclusive read/copy shape: %+v", result.Problems)
	}

	conflicting := phase25PointerProgram(t, "phase3", "shared_shared_accept.schway")
	var sharedLoans []int
	for i, operation := range conflicting.Functions[0].Linear.Operations {
		if operation.Kind == core.OpBorrowShared {
			sharedLoans = append(sharedLoans, i)
		}
	}
	if len(sharedLoans) < 2 {
		t.Fatalf("expected two shared loans, got %+v", conflicting.Functions[0].Linear.Operations)
	}
	conflicting.Functions[0].Linear.Operations[sharedLoans[1]].Kind = core.OpBorrowExclusive
	result := corevalidate.Validate(conflicting)
	if result.Valid {
		t.Fatal("core peer accepted a live shared loan overlapping an exclusive access")
	}
	if len(result.Problems) == 0 || result.Problems[0].Code != "core.borrow_conflict" {
		t.Fatalf("conflict mutation produced %+v, want core.borrow_conflict", result.Problems)
	}
}

func TestPhase25U64CopyOriginCore(t *testing.T) {
	program := phase25PointerProgram(t, "phase25", "exclusive_copy_accept.schway")
	function := &program.Functions[0]
	result := corevalidate.Validate(program)
	if !result.Valid {
		t.Fatalf("core peer rejected copied exclusive U64: %+v", result.Problems)
	}
	if signature := result.PeerSignatures()[function.ID]; !signature.Callable {
		t.Fatalf("copied U64 still carries a borrow origin in peer signature: %+v", signature)
	}

	conflicting := phase25PointerProgram(t, "phase25", "exclusive_copy_accept.schway")
	function = &conflicting.Functions[0]
	linear := function.Linear
	if len(linear.Operations) != 3 || linear.Operations[0].Kind != core.OpBorrowExclusive || linear.Operations[1].Kind != core.OpCopy || linear.Operations[2].Kind != core.OpReturn {
		t.Fatalf("exclusive-copy fixture has unexpected operations: %+v", linear.Operations)
	}
	linear.Places[2].Name = "overlap"
	linear.Operations[2].SourceID = function.ID + ":place:3"
	linear.Operations = []core.LinearOperation{
		linear.Operations[0],
		{ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1", Kind: core.OpBorrowShared, SourceID: function.Parameter.ID, TargetID: function.ID + ":place:2", TypeID: linear.Operations[0].TypeID, LoanID: "phase25:test:overlap"},
		{ID: function.ID + ":op:2", PointID: function.ID + ":point:linear:2", Kind: core.OpCopy, SourceID: linear.Operations[0].TargetID, TargetID: function.ID + ":place:3", TypeID: linear.Operations[0].TypeID},
		{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpCopy, SourceID: function.ID + ":place:2", TargetID: function.ID + ":place:4", TypeID: linear.Operations[0].TypeID},
		{ID: function.ID + ":op:4", PointID: function.ID + ":point:linear:4", Kind: core.OpReturn, SourceID: function.ID + ":place:3", TypeID: linear.Operations[0].TypeID},
	}
	linear.Places = append(linear.Places, core.Place{ID: function.ID + ":place:3", Name: "copied", TypeID: linear.Operations[0].TypeID}, core.Place{ID: function.ID + ":place:4", Name: "overlap_copy", TypeID: linear.Operations[0].TypeID})
	result = corevalidate.Validate(conflicting)
	if result.Valid || len(result.Problems) == 0 || result.Problems[0].Code != "core.borrow_conflict" {
		t.Fatalf("core peer did not keep the exclusive loan live through its U64 copy: %+v", result.Problems)
	}
}

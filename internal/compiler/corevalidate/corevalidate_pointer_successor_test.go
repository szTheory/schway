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

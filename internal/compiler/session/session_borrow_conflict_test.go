package session_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// borrowConflictDiagnostic is ownershipDiagnostic's testdata/phase3 twin: it
// reads a Phase 3 fixture, asserts it is rejected with exactly one
// diagnostic of the given code, and confirms the interpreter never reaches
// execution on invalid source (mirroring the Phase 2 helper's contract).
func borrowConflictDiagnostic(t *testing.T, fixture, code string) diagnostic.Diagnostic {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase3", fixture)
	checked, err := session.CheckFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 1 || (code != "" && checked.Diagnostics[0].Code != code) {
		t.Fatalf("%s: got diagnostics %+v, want one %q", fixture, checked.Diagnostics, code)
	}
	executions, runDiagnostics, err := session.RunInterpreterFile(path)
	if err != nil || len(executions) != 0 || len(runDiagnostics) != 1 || runDiagnostics[0].ID != checked.Diagnostics[0].ID {
		t.Fatalf("invalid source reached execution or changed diagnostic: executions=%+v diagnostics=%+v err=%v", executions, runDiagnostics, err)
	}
	return checked.Diagnostics[0]
}

// acceptedBorrowFixture reads a Phase 3 positive fixture, asserts it checks
// clean, and independently re-validates the checked core through
// corevalidate — proving the second admission layer agrees a genuinely
// non-overlapping (or shared/shared) pair of loans is legal.
func acceptedBorrowFixture(t *testing.T, fixture string) session.CheckResult {
	t.Helper()
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase3", fixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("%s: independent validator rejected an accepted program: %+v", fixture, result.Problems)
	}
	return checked
}

// TestSharedSharedOverlapAccepted is Q3 fixture 1: two overlapping shared
// loans on one owner never conflict — multiple simultaneous readers observe
// without mutation, and no gate fires.
func TestSharedSharedOverlapAccepted(t *testing.T) {
	checked := acceptedBorrowFixture(t, "shared_shared_accept.lang")
	function := checked.Program.Functions[0]
	sharedCount := 0
	for _, operation := range function.Linear.Operations {
		if operation.Kind == "borrow_shared" {
			sharedCount++
		}
	}
	// Two direct loans on `buffer` plus two reborrows of those loans: four
	// borrow_shared operations total, proving both original loans were
	// admitted (not silently narrowed to one).
	if sharedCount != 4 {
		t.Fatalf("want 4 borrow_shared operations (2 direct + 2 reborrow), got %d: %+v", sharedCount, function.Linear.Operations)
	}
}

// TestSequentialLoansAccepted is Q3 fixture 5, and directly exercises OWN-03
// success criterion 3: a shared loan whose last use precedes an exclusive
// borrow of the same owner is accepted, proving conflict is decided by
// overlapping liveness rather than lexical position — the shared and
// exclusive loans below are textually adjacent but never overlap in time.
func TestSequentialLoansAccepted(t *testing.T) {
	checked := acceptedBorrowFixture(t, "sequential_shared_then_exclusive_accept.lang")
	function := checked.Program.Functions[0]
	sawShared, sawExclusive := false, false
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case "borrow_shared":
			sawShared = true
		case "borrow_exclusive":
			sawExclusive = true
		}
	}
	if !sawShared || !sawExclusive {
		t.Fatalf("expected both a shared and an exclusive borrow operation: %+v", function.Linear.Operations)
	}
}

// TestBorrowConflictMatrix is Q3's compile-reject side: all three
// conflicting rows (shared+exclusive, exclusive+exclusive, exclusive+move)
// are rejected with a stable, repair-bearing diagnostic code — the fourth
// row (shared+shared) is TestSharedSharedOverlapAccepted's accept case, and
// the fifth (sequential) is TestSequentialLoansAccepted's accept case.
func TestBorrowConflictMatrix(t *testing.T) {
	tests := []struct {
		fixture string
		code    string
	}{
		{"shared_exclusive_reject.lang", "ownership.borrow_conflict"},
		{"exclusive_exclusive_reject.lang", "ownership.borrow_conflict"},
		{"exclusive_move_reject.lang", "ownership.move_while_borrowed"},
	}
	for _, test := range tests {
		t.Run(test.fixture, func(t *testing.T) {
			problem := borrowConflictDiagnostic(t, test.fixture, test.code)
			if problem.Schema != "lang.diagnostic/1" || len(problem.Repairs) == 0 {
				t.Fatalf("%s: rejection must join the repair-bearing taxonomy: %+v", test.fixture, problem)
			}
		})
	}
}

// TestBorrowConflictCauseChain asserts the ownership.borrow_conflict cause
// sequence matches check.go's other ownership diagnostics exactly:
// span-bearing causes first (the blocking loan's creation and later-use
// sites), then ID-bearing detail causes (loan, owner place, owner type).
func TestBorrowConflictCauseChain(t *testing.T) {
	problem := borrowConflictDiagnostic(t, "shared_exclusive_reject.lang", "ownership.borrow_conflict")
	assertCauseKinds(t, problem, "borrow_created_here", "borrow_used_later", "loan", "owner", "type")
	// shared_exclusive_reject.lang's conflicting loan is the NEW exclusive
	// (borrow_mut) one, so it also carries the D-06-24/D-06-25
	// narrow_to_shared_borrow MachineApplicable repair alongside the
	// pre-existing classification-only repair (sorted: c < n).
	assertRepairKinds(t, problem, "create_loan_after_conflicting_loan_ends", "narrow_to_shared_borrow")
}

// TestExclusiveMoveRejected mirrors TestMoveWhileBorrowedDiagnostic exactly,
// but for an exclusive loan: a move of an owner while an exclusive loan is
// live is rejected as ownership.move_while_borrowed, the same code and
// cause shape already shipped for shared loans (Q3 fixture 4).
func TestExclusiveMoveRejected(t *testing.T) {
	problem := borrowConflictDiagnostic(t, "exclusive_move_reject.lang", "ownership.move_while_borrowed")
	assertCauseKinds(t, problem, "borrow_created_here", "borrow_used_later", "loan", "owner", "type")
	assertRepairKinds(t, problem, "move_after_last_borrow_use")
}

package check_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
)

const exclusiveBorrowSource = `module owned.exclusive_borrow_lowering

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let delivered = take buffer
  let observed = view
  delivered
}
`

// TestExclusiveBorrowLowersToCore is 03-02-01's checker falsifier: checking
// `let view = borrow mut buffer` emits exactly one core.OpBorrowExclusive
// operation carrying a non-empty loan ID, gated on the same AbilityShare
// requirement shared borrows are gated on (this fixture's Buffer parameter
// grants share, so the gate is satisfied, not exercised as a rejection —
// see check.go's borrow_mut case for the D-10 note that the rejection path
// is source-unreachable, exactly like the shared case beside it). The
// existing move_while_borrowed law still fires generically for the
// exclusive loan (it is not re-derived per access mode).
func TestExclusiveBorrowLowersToCore(t *testing.T) {
	checked := session.Check([]byte(exclusiveBorrowSource))
	if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.move_while_borrowed" {
		t.Fatalf("expected exactly the pre-existing move-while-borrowed rejection, got %+v", checked.Diagnostics)
	}

	// A variant that never moves the owner while the exclusive loan is live
	// proves the exclusive operation itself lowers and validates cleanly.
	clean := `module owned.exclusive_borrow_clean

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let reviewed = borrow view
  view
}
`
	checkedClean := session.Check([]byte(clean))
	if len(checkedClean.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checkedClean.Diagnostics)
	}
	if len(checkedClean.Program.Functions) != 1 {
		t.Fatalf("want one function, got %d", len(checkedClean.Program.Functions))
	}
	function := checkedClean.Program.Functions[0]
	var exclusive *core.LinearOperation
	for index, operation := range function.Linear.Operations {
		if operation.Kind == core.OpBorrowExclusive {
			exclusive = &function.Linear.Operations[index]
		}
	}
	if exclusive == nil {
		t.Fatalf("expected one borrow_exclusive operation among %+v", function.Linear.Operations)
	}
	if exclusive.LoanID == "" {
		t.Fatalf("exclusive borrow operation carries no loan id: %+v", exclusive)
	}
	if result := corevalidate.Validate(checkedClean.Program); !result.Valid {
		t.Fatalf("independent validator rejected a valid exclusive-borrow program: %+v", result.Problems)
	}
}

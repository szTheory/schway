package session_test

import (
	"context"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
)

const exclusiveBorrowNativeSource = `module owned.exclusive_borrow_native

export {
  fn relay
}

fn relay(code: Byte) -> Byte {
  let view = borrow mut code
  let reviewed = view
  view
}
`

// TestExclusiveBorrowInterpreterNative is 03-02-01's end-to-end tracer
// falsifier (D-12a): interp.runLinear emits an ordered event for the
// exclusive borrow, generated C17 executes the corresponding statement, and
// O0/O3/interpreter agree — proving the exclusive operation appears in the
// execution trace at every consumer, not just the two admission layers.
func TestExclusiveBorrowInterpreterNative(t *testing.T) {
	result, diagnostics, err := session.RunNative(context.Background(), []byte(exclusiveBorrowNativeSource), native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("exclusive borrow native run failed: err=%v diagnostics=%+v", err, diagnostics)
	}
	if len(result.Interpreter) != 1 || len(result.O0.Pairs) != 1 || len(result.O3.Pairs) != 1 {
		t.Fatalf("expected exactly one execution: %+v", result)
	}
	execution := result.Interpreter[0]
	sawExclusive := false
	for _, event := range execution.Events {
		if event.Kind == "value.borrowed_exclusive" {
			sawExclusive = true
			if event.SourcePlace == "" || event.TargetPlace == "" {
				t.Fatalf("exclusive borrow event missing source/target: %+v", event)
			}
		}
	}
	if !sawExclusive {
		t.Fatalf("interpreter trace omitted the exclusive borrow event: %+v", execution.Events)
	}
	// session.RunNative already asserts O0/O3 agree with the interpreter by
	// returning a non-nil *session.EngineMismatch on the first disagreement;
	// reaching here with no error is itself the differential pass at both
	// optimization levels.
}

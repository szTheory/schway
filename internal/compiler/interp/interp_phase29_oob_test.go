package interp_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase29OOBBoundarySource = `module phase29.oob.boundary

export {
  fn at_len
}

fn at_len(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}

fn at_last(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var one = 1
  var i = 0
  var spent = 0
  if n < 1 {
    i = 0
  } else {
    spent = n + 0
    i = 0
    while i + one < n {
      i = i + 1
    }
  }
  var x = byte_at(bytes, i)
  x
}

fn at_zero(bytes: ByteView) -> U64 {
  var x = byte_at(bytes, 0)
  x
}

fn at_max(bytes: ByteView) -> U64 {
  var x = byte_at(bytes, 18446744073709551615)
  x
}
`

// TestPhase29InterpreterOOBBoundary pins the interpreter's bounds semantics
// (D-19): index == length defects, index == length-1 reads, index 0 on an
// empty view defects, and the maximum U64 index defects without wrapping.
func TestPhase29InterpreterOOBBoundary(t *testing.T) {
	checked := session.Check([]byte(phase29OOBBoundarySource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
	}
	rows := []struct {
		name     string
		function string
		input    []byte
		defect   bool
		want     string
	}{
		{"index equals length", "at_len", []byte("ab"), true, ""},
		{"index equals length on empty", "at_len", []byte{}, true, ""},
		{"index length-1 reads", "at_last", []byte("ab"), false, "98"},
		{"last index on empty is index 0", "at_last", []byte{}, true, ""},
		{"index 0 on empty", "at_zero", []byte{}, true, ""},
		{"index 0 on one byte", "at_zero", []byte("a"), false, "97"},
		{"max U64 index does not wrap", "at_max", []byte("a"), true, ""},
		{"max U64 index on empty", "at_max", []byte{}, true, ""},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			run, err := interp.RunByteViewForTest(checked.Program, row.function, row.input)
			if err != nil {
				t.Fatal(err)
			}
			if !row.defect {
				if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != row.want {
					t.Fatalf("outcome = %+v, want returned %s", run.Outcome, row.want)
				}
				return
			}
			if run.Outcome.Kind != execution.OutcomeDefect || run.Outcome.Value != "" {
				t.Fatalf("outcome = %+v, want empty defect", run.Outcome)
			}
			if len(run.Events) == 0 {
				t.Fatal("defect recorded no event")
			}
			final := run.Events[len(run.Events)-1]
			if final.Kind != "function.defected" || final.Output != "byte index out of bounds" || final.SourcePlace == "" {
				t.Fatalf("final event = %+v, want source-attributed function.defected", final)
			}
			for _, event := range run.Events[:len(run.Events)-1] {
				if event.Kind == "function.defected" {
					t.Fatalf("defect is not the last event: %+v", run.Events)
				}
			}
		})
	}
}

package interp

import (
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// phase321Program builds a minimal program whose only readable facts are the
// operation IDs and kinds the abort-like predicate looks up.
func phase321Program(kinds map[string]core.OperationKind) core.Program {
	operations := []core.LinearOperation{}
	for id, kind := range kinds {
		operations = append(operations, core.LinearOperation{ID: id, Kind: kind})
	}
	return core.Program{Functions: []core.Function{{Linear: &core.LinearBody{Operations: operations}}}}
}

func phase321Defect(finalID, output string) Execution {
	return Execution{
		Outcome: execution.Outcome{Kind: execution.OutcomeDefect},
		Events:  []Event{{Kind: "foreign.called", ID: "x:event"}, {Kind: "function.defected", ID: finalID, Output: output}},
	}
}

// TestPhase321AbortLikeDefectPredicate pins the closed set of four abort-like
// operation kinds (T-32.1-05) and every refused terminal (D-09, D-25).
func TestPhase321AbortLikeDefectPredicate(t *testing.T) {
	program := phase321Program(map[string]core.OperationKind{
		"op:bytes": core.OpBytesAt, "op:add": core.OpAddChecked, "op:rem": core.OpRemainderU64,
		"op:defect": core.OpDefect, "op:call": core.OpCall, "op:write": core.OpWriteU64,
		"op:nonlocal": core.OpCall, "op:depth": core.OpCall,
	})
	accepted := []struct{ id, output string }{
		{"op:bytes:event:defected", "byte index out of bounds"},
		{"op:add:event:defected", "U64 addition overflow"},
		{"op:rem:event:defected", "U64 remainder by zero"},
		{"op:defect:event:defected", "boom"},
	}
	for _, row := range accepted {
		if !isAbortLikeDefect(program, phase321Defect(row.id, row.output)) {
			t.Errorf("%s should be abort-like", row.id)
		}
	}
	refused := []struct {
		name   string
		result Execution
	}{
		{"call operation", phase321Defect("op:call:event:defected", "x")},
		{"output operation", phase321Defect("op:write:event:defected", "x")},
		{"nonlocal defect", phase321Defect("op:nonlocal:event:nonlocal_defect", "x")},
		{"call depth", phase321Defect("op:depth:event:call_depth_exceeded", "x")},
		{"unknown operation", phase321Defect("op:missing:event:defected", "x")},
		{"wrong event kind", Execution{Outcome: execution.Outcome{Kind: execution.OutcomeDefect}, Events: []Event{{Kind: "function.returned", ID: "op:add:event:defected"}}}},
		{"returned", Execution{Outcome: execution.Outcome{Kind: execution.OutcomeReturned}, Events: []Event{{Kind: "function.defected", ID: "op:add:event:defected"}}}},
		{"typed failure", Execution{Outcome: execution.Outcome{Kind: execution.OutcomeTypedFailure}, Events: []Event{{Kind: "function.defected", ID: "op:add:event:defected"}}}},
		{"empty events", Execution{Outcome: execution.Outcome{Kind: execution.OutcomeDefect}}},
	}
	for _, row := range refused {
		if isAbortLikeDefect(program, row.result) {
			t.Errorf("%s must stay refused", row.name)
		}
	}
}

// phase321FileSource assembles a single-module checked file-owner entry (D-26:
// in memory, no committed .schway file) around the supplied callee text.
func phase321FileSource(callee string) []byte {
	return phase321FileSourceTail(callee, "  sum\n")
}

// phase321FileSourceTail lets a subtest extend main after the callee returns,
// while the file owner is still live (copy-local owner: released at main's end).
func phase321FileSourceTail(callee, tail string) []byte {
	return []byte(`module phase321.file_defect

export {
  fn main
}

foreign C {
  fn schway_file_bytes_acquire(path: PathToken) -> FileBytes {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "acquire"
    abi: schway_file_bytes_acquire_fn
    allocator: "libc_malloc"
    release: "schway_file_bytes_release"
    fails: FileError
  }

  fn schway_file_bytes_view(owner: FileBytes) -> ByteView {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "borrow"
    abi: schway_file_bytes_view_fn
  }

  fn schway_file_bytes_release(owner: FileBytes) -> Unit {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "consume"
    abi: schway_file_bytes_release_fn
    allocator: "libc_malloc"
  }
}

data FileError =
  | FileTooLarge
  | FileUnavailable
  | AllocFailed

` + callee + `
fn main(path: PathToken) -> U64 {
  let owner = try schway_file_bytes_acquire(path)
  let view = schway_file_bytes_view(owner)
  let sum = adler32(view)
` + tail + `}
`)
}

func phase321CheckFile(t *testing.T, callee string) core.Program {
	t.Helper()
	return phase321CheckFileTail(t, callee, "  sum\n")
}

func phase321CheckFileTail(t *testing.T, callee, tail string) core.Program {
	t.Helper()
	parsed := syntax.Parse(phase321FileSourceTail(callee, tail))
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse: %v near %q", parsed.Diagnostics, phase321FileSourceTail(callee, tail)[780:900])
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check: %v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate: %v", validated.Problems)
	}
	return validated.Program()
}

// TestPhase321AbortLikeDefectFileRoute executes the shared predicate through
// RunWithFileBytes (D-25): defect accepted, baseline clean, call depth refused.
func TestPhase321AbortLikeDefectFileRoute(t *testing.T) {
	read := func(string) ([]byte, string) { return []byte("Wikipedia"), "" }

	t.Run("defect with the file owner live yields a defect execution", func(t *testing.T) {
		program := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 1000 {
    defect "boom"
  } else {
    n = n + 0
  }
  n
}
`)
		result, err := RunWithFileBytes(program, "main", "any", read)
		if err != nil {
			t.Fatalf("abort-like defect rejected on the file route: %v", err)
		}
		if result.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("outcome = %+v, want defect", result.Outcome)
		}
		final := result.Events[len(result.Events)-1]
		if final.Kind != "function.defected" || final.Output != "boom" {
			t.Fatalf("final event = %+v, want function.defected boom", final)
		}
		if len(result.LiveResources) != 1 {
			t.Fatalf("live resources = %v, want exactly the file owner", result.LiveResources)
		}
	})

	t.Run("returning callee baseline", func(t *testing.T) {
		program := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  n
}
`)
		result, err := RunWithFileBytes(program, "main", "any", read)
		if err != nil {
			t.Fatal(err)
		}
		if result.Outcome.Kind != execution.OutcomeReturned || result.Outcome.Value != "9" || len(result.LiveResources) != 0 {
			t.Fatalf("result = %+v live=%v, want returned 9 with no live resources", result.Outcome, result.LiveResources)
		}
	})

	t.Run("call-depth defect stays outside the abort-like set", func(t *testing.T) {
		var b strings.Builder
		n := MaxCallDepth + 1
		b.WriteString("fn adler32(bytes: ByteView) -> U64 {\n  var n = byte_len(bytes)\n  n\n}\n\n")
		for i := 0; i < n; i++ {
			if i == n-1 {
				fmt.Fprintf(&b, "fn link%d(value: U64) -> U64 {\n  value\n}\n\n", i)
				continue
			}
			fmt.Fprintf(&b, "fn link%d(value: U64) -> U64 {\n  let result = link%d(value)\n  result\n}\n\n", i, i+1)
		}
		// The chain is entered from main while the file owner is still live.
		program := phase321CheckFileTail(t, b.String(), "  let deep = link0(sum)\n  deep\n")
		// The checker admits only a PathToken entry with one view call and one
		// trailing U64 -> U64 call, and a ByteView-parameter linear callee is
		// unexecutable, so no checked source can run a call chain while the file
		// owner is live: the owner is released before the trailing call. The
		// real execution therefore proves the predicate verdict on an actual
		// call-depth terminal (white-box coverage of the live case is in
		// TestPhase321AbortLikeDefectPredicate).
		result, err := RunWithFileBytes(program, "main", "any", read)
		if err != nil {
			t.Fatalf("call-depth terminal after owner release: %v", err)
		}
		final := result.Events[len(result.Events)-1]
		if result.Outcome.Kind != execution.OutcomeDefect || !strings.HasSuffix(final.ID, ":event:call_depth_exceeded") {
			t.Fatalf("outcome = %+v final = %+v, want a call-depth defect", result.Outcome, final)
		}
		if len(result.LiveResources) != 0 {
			t.Fatalf("live resources = %v, want the owner already released", result.LiveResources)
		}
		if isAbortLikeDefect(program, result) {
			t.Fatal("call-depth defect must not be classified abort-like")
		}
	})
}

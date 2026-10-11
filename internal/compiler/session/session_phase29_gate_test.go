package session_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase29NeighborBaseline pins the first diagnostic code of the neighbor
// refusals that Phase 29 must not move: ownership expressions inside a loop
// body and Buffer-typed loop signatures. The sources are inline strings so no
// .schway file is added to the corpus.
func TestPhase29NeighborBaseline(t *testing.T) {
	cases := []struct {
		name   string
		source string
		code   string
	}{
		{
			name: "borrow inside a loop",
			source: `module phase29.baseline.borrow_loop
export { fn main }

fn main(n: U64) -> U64 {
  var i = 0
  while i < n {
    let b = borrow n
    i = i + 1
  }
  i
}
`,
			code: "syntax.expected_scalar_operand",
		},
		{
			name: "take inside a loop",
			source: `module phase29.baseline.take_loop
export { fn main }

fn main(n: U64) -> U64 {
  var i = 0
  while i < n {
    let t = take n
    i = i + 1
  }
  i
}
`,
			code: "syntax.expected_scalar_operand",
		},
		{
			name: "Buffer parameter on a loop function",
			source: `module phase29.baseline.buffer_param
export { fn main }

fn main(b: Buffer) -> U64 {
  var i = 0
  while i < 1 {
    i = i + 1
  }
  i
}
`,
			code: "type.scalar_signature",
		},
		{
			name: "Buffer result on a loop function",
			source: `module phase29.baseline.buffer_result
export { fn main }

fn main(n: U64) -> Buffer {
  var i = 0
  while i < n {
    i = i + 1
  }
  i
}
`,
			code: "type.scalar_signature",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := session.Check([]byte(tc.source))
			if len(result.Diagnostics) == 0 {
				t.Fatalf("expected %s, got no diagnostics", tc.code)
			}
			if got := result.Diagnostics[0].Code; got != tc.code {
				t.Fatalf("first diagnostic = %q (%+v), want %q", got, result.Diagnostics, tc.code)
			}
		})
	}
}

func phase29Library(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase29", "adler32_library.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("adler32 library diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

// TestPhase29GateCallable proves flag (a): the adler32 library checks clean,
// originvalidate publishes adler32 as Callable, and the interpreter returns
// the pinned literal for one vector end to end.
func TestPhase29GateCallable(t *testing.T) {
	program := phase29Library(t)
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected the library: %+v", validated.Problems)
	}
	iface, err := originvalidate.BuildInterface(validated.Program())
	if err != nil {
		t.Fatalf("BuildInterface: %v", err)
	}
	found := false
	for _, signature := range iface.Functions {
		if signature.Name != "adler32" {
			continue
		}
		found = true
		if !signature.Callable {
			t.Fatalf("adler32 is not Callable: %+v", signature)
		}
		if len(signature.Parameters) != 1 || signature.Parameters[0].Type != "ByteView" {
			t.Fatalf("adler32 signature parameters = %+v, want one ByteView parameter", signature.Parameters)
		}
	}
	if !found {
		t.Fatal("adler32 is absent from the published interface")
	}
	run, err := interp.RunByteViewForTest(program, "adler32", []byte("a"))
	if err != nil {
		t.Fatalf("RunByteViewForTest: %v", err)
	}
	if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != "6422626" {
		t.Fatalf("adler32(a) outcome = %+v, want returned 6422626", run.Outcome)
	}
}

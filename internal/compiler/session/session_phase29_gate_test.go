package session_test

import (
	"os"
	"strings"
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

const phase29ByteSource = `module phase29.gate.edges

export {
  fn main
}

fn at0(bytes: ByteView) -> U64 {
  var x = byte_at(bytes, 0)
  x
}

fn at1(bytes: ByteView) -> U64 {
  var one = 1
  var x = byte_at(bytes, one)
  x
}

fn len0(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  n
}

fn past(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}

fn last(bytes: ByteView) -> U64 {
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

fn main(n: U64) -> U64 {
  var kept = n
  kept
}
`

func phase29Repeat(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}

// TestPhase29GateInterpreterVectors is the Stage-A interpreter half of the
// Wave-0 gate: the pinned adler32 oracle literals, the unsigned-octet and
// empty-view edges, the bounds defect, and flag (e) event silence.
func TestPhase29GateInterpreterVectors(t *testing.T) {
	program := phase29Library(t)
	// Oracle literals produced 2026-10-11 by Python 3.14.4 `zlib.adler32` and
	// Go 1.24.0 `hash/adler32`; both agree on every row.
	vectors := []struct {
		name  string
		input []byte
		want  string
	}{
		{"empty", []byte{}, "1"},
		{"one zero byte", []byte{0x00}, "65537"},
		{"a", []byte("a"), "6422626"},
		{"Wikipedia", []byte("Wikipedia"), "300286872"},
		{"4096 x 0x00", phase29Repeat(0x00, 4096), "268435457"},
		{"5552 x 0xFF", phase29Repeat(0xFF, 5552), "4052720524"},
		{"high bit pair", []byte{0x80, 0xFF}, "33620352"},
	}
	for _, vector := range vectors {
		t.Run(vector.name, func(t *testing.T) {
			run, err := interp.RunByteViewForTest(program, "adler32", vector.input)
			if err != nil {
				t.Fatal(err)
			}
			if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != vector.want {
				t.Fatalf("outcome = %+v, want returned %s", run.Outcome, vector.want)
			}
			// Flag (e): neither byte op records an event on success.
			for _, event := range run.Events {
				if strings.Contains(event.ID, "bytes_len") || strings.Contains(event.ID, "bytes_at") {
					t.Fatalf("byte op recorded an event: %+v", event)
				}
			}
			bytesOps := map[string]bool{}
			for _, function := range program.Functions {
				if function.Name != "adler32" {
					continue
				}
				for _, operation := range function.Linear.Operations {
					if operation.Kind == core.OpBytesLen || operation.Kind == core.OpBytesAt {
						bytesOps[operation.ID] = true
					}
				}
			}
			if len(bytesOps) == 0 {
				t.Fatal("adler32 lowered no byte operations")
			}
			for _, event := range run.Events {
				if bytesOps[strings.TrimSuffix(event.ID, ":event:defected")] || bytesOps[event.SourcePlace] {
					t.Fatalf("event attributed to a byte operation on success: %+v", event)
				}
			}
		})
	}

	edgeChecked := session.Check([]byte(phase29ByteSource))
	if len(edgeChecked.Diagnostics) != 0 {
		t.Fatalf("edge source diagnostics: %+v", edgeChecked.Diagnostics)
	}
	edges := []struct {
		name     string
		function string
		input    []byte
		want     string
	}{
		{"at0 unsigned 0x80", "at0", []byte{0x80, 0xFF}, "128"},
		{"at1 unsigned 0xFF", "at1", []byte{0x80, 0xFF}, "255"},
		{"len0 empty", "len0", []byte{}, "0"},
		{"last index in range", "last", []byte{0x80, 0xFF}, "255"},
		{"last on single byte", "last", []byte("a"), "97"},
	}
	for _, edge := range edges {
		t.Run(edge.name, func(t *testing.T) {
			run, err := interp.RunByteViewForTest(edgeChecked.Program, edge.function, edge.input)
			if err != nil {
				t.Fatal(err)
			}
			if run.Outcome.Kind != execution.OutcomeReturned || run.Outcome.Value != edge.want {
				t.Fatalf("outcome = %+v, want returned %s", run.Outcome, edge.want)
			}
		})
	}

	for _, input := range [][]byte{{}, []byte("a")} {
		run, err := interp.RunByteViewForTest(edgeChecked.Program, "past", input)
		if err != nil {
			t.Fatal(err)
		}
		if run.Outcome.Kind != execution.OutcomeDefect || run.Outcome.Value != "" {
			t.Fatalf("len %d: outcome = %+v, want empty defect", len(input), run.Outcome)
		}
		if len(run.Events) == 0 {
			t.Fatalf("len %d: defect recorded no event", len(input))
		}
		final := run.Events[len(run.Events)-1]
		if final.Kind != "function.defected" || final.Output != "byte index out of bounds" {
			t.Fatalf("len %d: final event = %+v, want function.defected byte index out of bounds", len(input), final)
		}
	}
}

package session_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
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

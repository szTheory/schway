package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func phase31NeighborSource(params, result, body string) []byte {
	return []byte("module phase29.neighbor\nexport { fn f }\n\nfn f(" + params + ") -> " + result + " {\n" + body + "}\n")
}

// TestPhase31LoopNeighborRefusals pins, through the public CLI, that each of
// the nine Phase 29 loop-function neighbours is refused with exit 2 and the
// same first code TestPhase29NeighborRefusals pins (EVD-14 / D-13).
func TestPhase31LoopNeighborRefusals(t *testing.T) {
	const loopTail = "  var i = 0\n  while i < n {\n    i = i + 1\n  }\n  i\n"
	cases := []struct {
		name   string
		source []byte
		code   string
	}{
		{"borrow inside the loop", phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var i = 0\n  while i < n {\n    let b = borrow n\n    i = i + 1\n  }\n  i\n"), "syntax.expected_scalar_operand"},
		{"take inside the loop", phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var i = 0\n  while i < n {\n    let t = take n\n    i = i + 1\n  }\n  i\n"), "syntax.expected_scalar_operand"},
		{"Buffer parameter", phase31NeighborSource("b: Buffer", "U64", "  var n = 1\n"+loopTail), "type.scalar_signature"},
		{"Buffer result", phase31NeighborSource("bytes: ByteView", "Buffer", "  var n = byte_len(bytes)\n"+loopTail), "type.scalar_signature"},
		{"ByteView result", phase31NeighborSource("bytes: ByteView", "ByteView", "  var n = byte_len(bytes)\n"+loopTail), "type.scalar_signature"},
		{"copying the view", phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  let c = bytes\n"+loopTail), "type.scalar_binding"},
		{"storing the view in a var", phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var s = bytes\n"+loopTail), "type.scalar_binding"},
		{"assigning the view", phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var s = 0\n  s = bytes\n"+loopTail), "type.scalar_assignment"},
		{"returning the view", phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n  var i = 0\n  while i < n {\n    i = i + 1\n  }\n  bytes\n"), "type.scalar_result"},
	}
	check := func(t *testing.T, source []byte) (int, string) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "neighbor.schway")
		if err := os.WriteFile(path, source, 0o600); err != nil {
			t.Fatal(err)
		}
		code, stdout, stderr := captureLangRun(t, []string{"check", path})
		return code, stdout + stderr
	}
	t.Run("positive-control", func(t *testing.T) {
		code, out := check(t, phase31NeighborSource("bytes: ByteView", "U64", "  var n = byte_len(bytes)\n"+loopTail))
		if code != 0 {
			t.Fatalf("var-first loop control exit = %d, output %q, want 0", code, out)
		}
	})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out := check(t, tc.source)
			if code != 2 || !strings.Contains(out, tc.code) {
				t.Fatalf("exit = %d, output %q, want exit 2 naming %s", code, out, tc.code)
			}
		})
	}
}

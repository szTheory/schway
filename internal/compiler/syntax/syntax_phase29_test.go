package syntax_test

import (
	"bytes"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func phase29Source(body string) []byte {
	return []byte("module phase29.surface\n\nexport {\n  fn probe\n}\n\nfn probe(bytes: ByteView) -> U64 {\n" + body + "}\n")
}

func TestPhase29ParseByteReads(t *testing.T) {
	parsed := syntax.Parse(phase29Source("  var n = byte_len(bytes)\n  var x = byte_at(bytes, n + 1)\n  x\n"))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("diagnostics: %+v", parsed.Diagnostics)
	}
	statements := parsed.Program.Funcs[0].Body.Linear.Statements
	if len(statements) != 2 {
		t.Fatalf("statements = %+v", statements)
	}
	if got := statements[0].Expr; got.Kind != "byte_len" || got.Name != "bytes" {
		t.Fatalf("byte_len expr = %+v", got)
	}
	got := statements[1].Expr
	if got.Kind != "byte_at" || got.Name != "bytes" || got.Right == nil || got.Right.Kind != "add" {
		t.Fatalf("byte_at expr = %+v", got)
	}
}

func TestPhase29ByteReadArity(t *testing.T) {
	for _, test := range []struct {
		name, body, code string
	}{
		{"at three operands", "  var x = byte_at(bytes, i, j)\n  x\n", "syntax.byte_at_arity"},
		{"at one operand", "  var x = byte_at(bytes)\n  x\n", "syntax.byte_at_arity"},
		{"at no operands", "  var x = byte_at()\n  x\n", "syntax.byte_at_arity"},
		{"len no operands", "  var x = byte_len()\n  x\n", "syntax.byte_len_arity"},
		{"len two operands", "  var x = byte_len(bytes, i)\n  x\n", "syntax.byte_len_arity"},
	} {
		t.Run(test.name, func(t *testing.T) {
			parsed := syntax.Parse(phase29Source(test.body))
			if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Code != test.code {
				t.Fatalf("diagnostics = %+v, want first %s", parsed.Diagnostics, test.code)
			}
		})
	}
}

func TestPhase29LetFirstByteViewBody(t *testing.T) {
	for _, body := range []string{
		"  let n = byte_len(bytes)\n  n\n",
		"  let x = byte_at(bytes, 0)\n  x\n",
	} {
		parsed := syntax.Parse(phase29Source(body))
		if len(parsed.Diagnostics) != 0 {
			t.Fatalf("%q diagnostics: %+v", body, parsed.Diagnostics)
		}
		linear := parsed.Program.Funcs[0].Body.Linear
		if linear == nil || len(linear.Statements) != 1 || linear.Statements[0].Kind != "let" {
			t.Fatalf("%q did not take the scalar path: %+v", body, linear)
		}
	}
}

func TestPhase29DataParameterRefused(t *testing.T) {
	parsed := syntax.Parse([]byte("module phase29.surface\n\nexport {\n  fn adler32\n}\n\nfn adler32(data: ByteView) -> U64 {\n  var x = 1\n  x\n}\n"))
	if len(parsed.Diagnostics) == 0 || parsed.Diagnostics[0].Code != "syntax.expected_parameter_name" {
		t.Fatalf("diagnostics = %+v", parsed.Diagnostics)
	}
}

func TestPhase29FormatByteReads(t *testing.T) {
	for _, test := range []struct {
		name, input, want string
	}{
		{"comma spacing", "  let x = byte_at(bytes,i)\n  x\n", "  let x = byte_at(bytes, i)\n  x\n"},
		{"extra comma space", "  let x = byte_at(bytes,   0)\n  x\n", "  let x = byte_at(bytes, 0)\n  x\n"},
		{"len inner space", "  let n = byte_len( bytes )\n  n\n", "  let n = byte_len(bytes)\n  n\n"},
		{"canonical len", "  let n = byte_len(bytes)\n  n\n", "  let n = byte_len(bytes)\n  n\n"},
		{"canonical at", "  let x = byte_at(bytes, 0)\n  x\n", "  let x = byte_at(bytes, 0)\n  x\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			first := session.Format(phase29Source(test.input))
			if len(first.Diagnostics) != 0 {
				t.Fatalf("format diagnostics: %+v", first.Diagnostics)
			}
			if want := string(phase29Source(test.want)); string(first.Canonical) != want {
				t.Fatalf("Format = %q, want %q", first.Canonical, want)
			}
			second := session.Format(first.Canonical)
			if !bytes.Equal(first.Canonical, second.Canonical) {
				t.Fatalf("not idempotent: %q vs %q", first.Canonical, second.Canonical)
			}
		})
	}
}

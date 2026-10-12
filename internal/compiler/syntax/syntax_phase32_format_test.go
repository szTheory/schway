package syntax_test

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase32FormatVarBody pins D-10: a bare trailing result identifier on its
// own source line after a `var` body statement stays on its own line, so the
// formatter never fuses it into the preceding assignment (`r = r + a` + `r`
// must not become `r = r + ar`).
func TestPhase32FormatVarBody(t *testing.T) {
	assertRoundTrip := func(t *testing.T, source []byte) []byte {
		t.Helper()
		before := session.Check(source)
		parsed := syntax.Parse(source)
		canonical := syntax.Format(parsed.Tree)
		reparsed := syntax.Parse(canonical)
		if len(reparsed.Diagnostics) != 0 {
			t.Fatalf("canonical form does not reparse: %+v\n%q", reparsed.Diagnostics, canonical)
		}
		if again := syntax.Format(reparsed.Tree); !bytes.Equal(canonical, again) {
			t.Fatalf("not a fixed point: %q vs %q", canonical, again)
		}
		if got, want := semanticTokens(reparsed.Tree), semanticTokens(parsed.Tree); !reflect.DeepEqual(got, want) {
			t.Fatalf("semantic tokens changed:\n got %v\nwant %v", got, want)
		}
		after := session.Check(canonical)
		if got, want := diagnosticCodes(after.Diagnostics), diagnosticCodes(before.Diagnostics); !reflect.DeepEqual(got, want) {
			t.Fatalf("check diagnostics changed: got %v, want %v", got, want)
		}
		return canonical
	}

	for _, test := range []struct {
		name, body, tail string
	}{
		{"identifier-terminal", "  var a = 1\n  var r = 1\n  r = r + a\n  r\n", "r = r + a\n  r\n}\n"},
		{"number-terminal", "  var r = 1\n  r = r % 7\n  r\n", "r = r % 7\n  r\n}\n"},
		{"call-paren-terminal", "  var n = byte_len(bytes)\n  var x = byte_at(bytes, n)\n  x\n", "x = byte_at(bytes, n)\n  x\n}\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			canonical := assertRoundTrip(t, phase29Source(test.body))
			if !strings.HasSuffix(string(canonical), test.tail) {
				t.Fatalf("canonical = %q, want suffix %q", canonical, test.tail)
			}
		})
	}

	t.Run("pre-fix-output-rejected", func(t *testing.T) {
		input := syntax.Parse(phase29Source("  var a = 1\n  var r = 1\n  r = r + a\n  r\n"))
		preFix := syntax.Parse(phase29Source("  var a = 1\n  var r = 1\n  r = r + ar\n"))
		if reflect.DeepEqual(semanticTokens(input.Tree), semanticTokens(preFix.Tree)) {
			t.Fatal("semantic-token equality failed to reject the pre-fix merged output")
		}
	})

	t.Run("committed-sources-as-copies", func(t *testing.T) {
		for _, path := range []string{
			"examples/checksum/adler32.schway",
			"examples/checksum/adler_low.schway",
			"examples/phase29/adler32_fixture.schway",
			"examples/phase29/byte_index_oob.schway",
			"testdata/phase29/adler32_library.schway",
		} {
			t.Run(path, func(t *testing.T) {
				source, err := os.ReadFile(testsupport.ProjectPath(strings.Split(path, "/")...))
				if err != nil {
					t.Fatal(err)
				}
				if len(syntax.Parse(source).Diagnostics) != 0 {
					t.Fatalf("source does not parse: %+v", syntax.Parse(source).Diagnostics)
				}
				assertRoundTrip(t, append([]byte(nil), source...))
			})
		}
	})
}

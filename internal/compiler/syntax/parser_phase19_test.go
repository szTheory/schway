package syntax_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase19NumericLiteralRoundTrip(t *testing.T) {
	base, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	for _, literal := range []string{"0", "00", "08", "42", "4_2", "0x2a", "0x2_A", "0b101010", "0b10_1010"} {
		t.Run(literal, func(t *testing.T) {
			source := []byte(strings.Replace(string(base), "let count = 42", "let count = "+literal, 1))
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 {
				t.Fatalf("parse: %+v", parsed.Diagnostics)
			}
			bindings := parsed.Program.Funcs[0].Body.Linear.Bindings
			if len(bindings) != 1 || bindings[0].RHS.Kind != "numeric_literal" || bindings[0].RHS.Source != literal {
				t.Fatalf("literal RHS = %+v", bindings)
			}
			span := bindings[0].RHS.Span
			if span.End-span.Start != len(literal) || string(source[span.Start:span.End]) != literal {
				t.Fatalf("literal span %+v does not retain %q", span, literal)
			}
			first := session.Format(source)
			second := session.Format(first.Canonical)
			if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) || !bytes.Contains(first.Canonical, []byte("= "+literal)) {
				t.Fatalf("format changed numeric spelling or is not idempotent: first=%q second=%q", first.Canonical, second.Canonical)
			}
		})
	}
}

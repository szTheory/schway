package syntax_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase18ComputedTerminalMatchParsesAfterLinearPrefix(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "computed_match.schway"))
	if err != nil {
		t.Fatalf("read computed-match fixture: %v", err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parse computed terminal match: %+v", parsed.Diagnostics)
	}
	if len(parsed.Program.Funcs) != 1 {
		t.Fatalf("function count = %d, want 1", len(parsed.Program.Funcs))
	}
	body := parsed.Program.Funcs[0].Body.Linear
	if body == nil || body.TerminalMatch == nil {
		t.Fatalf("function body did not preserve its terminal match: %+v", parsed.Program.Funcs[0].Body)
	}
	if len(body.Bindings) != 1 || body.Bindings[0].Name != "computed" {
		t.Fatalf("linear prefix = %+v, want the computed binding", body.Bindings)
	}
	match := body.TerminalMatch
	if match.Scrutinee != "computed" || len(match.Arms) != 2 {
		t.Fatalf("terminal match = %+v, want two arms over computed", match)
	}
	if match.Span.Start <= 0 || match.Span.End <= match.Span.Start || string(source[match.Span.Start:match.Span.Start+5]) != "match" {
		t.Fatalf("terminal match span does not preserve its source range: %+v", match.Span)
	}
}

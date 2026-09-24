package syntax_test

import (
	"os"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase18ComputedTerminalMatchParsesAfterLinearPrefix(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "computed_match.lang"))
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
}

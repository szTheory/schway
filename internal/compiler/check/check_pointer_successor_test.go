package check_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25SharedSource(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", "shared_copy_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("shared-copy fixture failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("ordinary shared read/copy helper was refused: %+v", result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("checked function count=%d, want 1", len(result.Program.Functions))
	}
	function := result.Program.Functions[0]
	if function.Name != "shared_copy" || function.Parameter.Type != "U64" || function.ReturnType != "U64" {
		t.Fatalf("unexpected helper contract: %+v", function)
	}
	if function.Linear == nil || len(function.Linear.Operations) != 2 {
		t.Fatalf("helper must contain one shared borrow and one return: %+v", function.Linear)
	}
	borrow, ret := function.Linear.Operations[0], function.Linear.Operations[1]
	if borrow.Kind != core.OpBorrowShared || borrow.SourceID != function.Parameter.ID || borrow.LoanID == "" {
		t.Fatalf("checked helper lost its explicit shared-family fact: %+v", borrow)
	}
	if ret.Kind != core.OpReturn || ret.SourceID != borrow.TargetID {
		t.Fatalf("helper must return the copied shared value: %+v", ret)
	}
}

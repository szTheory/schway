package corevalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// TestPhase35CorePeerBaseline pins the existing OpConst path before the
// independent Phase 35 immediate and transfer laws are added.
func TestPhase35CorePeerBaseline(t *testing.T) {
	parsed := syntax.Parse([]byte("module phase35.core_peer_baseline\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let value = 42\n  value\n}\n"))
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("baseline source did not parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		t.Fatalf("baseline source did not check: diagnostics=%+v functions=%d", checked.Diagnostics, len(checked.Program.Functions))
	}
	function := checked.Program.Functions[0]
	if function.Linear == nil || len(function.Linear.Operations) < 2 || function.Linear.Operations[0].Kind != core.OpConst || function.Linear.Operations[0].ConstU64 != "42" {
		t.Fatalf("baseline OpConst shape changed: %+v", function.Linear)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("corevalidate stopped accepting the existing OpConst program: %+v", result.Problems)
	}
}

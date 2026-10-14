package originvalidate_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// TestPhase35OriginPeerBaseline pins the existing OpConst publication path
// before the independent Phase 35 immediate and transfer laws are added.
func TestPhase35OriginPeerBaseline(t *testing.T) {
	parsed := syntax.Parse([]byte("module phase35.origin_peer_baseline\nexport { fn main }\nfn main(input: U64) -> U64 {\n  let value = 42\n  value\n}\n"))
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
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("originvalidate stopped accepting the existing OpConst program: %+v", problems)
	}
}

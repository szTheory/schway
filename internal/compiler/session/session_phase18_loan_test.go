package session_test

import (
	"context"
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase18LoanAcrossBranchProduction reruns S-010 through source parsing,
// production endpoint materialization, independent peers, interpreter, and
// every available native tier.
func TestPhase18LoanAcrossBranchProduction(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "loan_across_branch.schway"))
	if err != nil {
		t.Fatalf("read S-010 source: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("production checker refused S-010 source: %+v", checked.Diagnostics)
	}
	t.Logf("S-010 production checker work=%d", checked.Work)
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("independent core admission: %+v", peer.Problems)
	}
	if _, err := originvalidate.BuildInterface(checked.Program); err != nil {
		t.Fatalf("independent origin admission: %v", err)
	}
	var selected *core.Function
	for index := range checked.Program.Functions {
		if checked.Program.Functions[index].Name == "select" {
			selected = &checked.Program.Functions[index]
		}
	}
	if selected == nil || selected.Linear == nil {
		t.Fatal("checked source omitted select's production linear CFG")
	}
	var loanID string
	for _, operation := range selected.Linear.Operations {
		if operation.Kind == core.OpBorrowShared {
			loanID = operation.LoanID
		}
	}
	var sawPoint, sawEdge bool
	for _, endpoint := range selected.Linear.LoanEndpoints {
		if endpoint.LoanID != loanID {
			continue
		}
		switch endpoint.Kind {
		case "point":
			sawPoint = true
		case "edge":
			sawEdge = true
		default:
			t.Fatalf("unexpected production endpoint kind %q", endpoint.Kind)
		}
	}
	if loanID == "" || !sawPoint || !sawEdge {
		t.Fatalf("production endpoint witness missing point/edge kinds: loan=%q endpoints=%+v", loanID, selected.Linear.LoanEndpoints)
	}

	for _, input := range []string{"On", "Off"} {
		engines := phase11RunFourTiersWithSupplier(t, context.Background(), checked.Program, "select", input, cgen.EmitNative, "public EmitNative")
		engines["interpreter"] = phase16ProjectInterpreterSchema2(t, checked.Program, engines["interpreter"])
		if len(engines) != 4 {
			t.Fatalf("input %s produced %d engines, want interpreter, O0, O3, O3-LTO", input, len(engines))
		}
		for name, result := range engines {
			if result.Outcome.Kind != "returned" || result.Outcome.Value != input {
				t.Fatalf("input %s engine %s outcome = %+v", input, name, result.Outcome)
			}
		}
		if err := session.Phase5CompareProgramEngines("testdata/phase18/loan_across_branch.schway:"+input, checked.Program, engines); err != nil {
			t.Fatalf("input %s four-engine comparison: %v", input, err)
		}
	}
}

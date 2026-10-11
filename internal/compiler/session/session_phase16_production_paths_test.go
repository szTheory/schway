package session_test

import (
	"context"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase16ProductionBypassMutationIsKilled(t *testing.T) {
	// Seed the kind of historical-artifact selector a production bypass would
	// add. The source guard must detect the concrete mutation.
	mutated := `func emitAfterRefusal() string { return "testdata/phase16/historical/restrict_borrow.c" }`
	if refs := phase16FrozenCSelectors(mutated); len(refs) == 0 {
		t.Fatalf("seeded frozen-C bypass was not detected: %s", mutated)
	}
}

func TestPhase16ProductionPathsPreserveM004Refusal(t *testing.T) {
	for _, fixture := range []string{
		"testdata/phase4/nonlocal_exit_probe.schway",
		// 2026-10-10 (Phase 28 D-10/D-20): restrict_borrow.schway is origin-refused
		// by check, so admission now reports it as exit 2 before any tool; the
		// check-clean cut fixture keeps the terminal M004 native refusal covered.
		"testdata/phase4/acquire_three_success.schway",
	} {
		t.Run(fixture, func(t *testing.T) {
			result, err := session.RunNativeCommandFile(context.Background(), testsupport.ProjectPath(fixture), native.DefaultRunner())
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != protocol.StatusOperational || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "native.tool_failure" {
				t.Fatalf("production native command did not preserve the terminal M004 refusal: %+v", result)
			}
			if len(result.Executions) != 0 {
				t.Fatalf("production native command returned executions for refused M004 source: %+v", result.Executions)
			}
		})
	}
}

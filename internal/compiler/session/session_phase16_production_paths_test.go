package session_test

import (
	"context"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
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
		"testdata/phase4/nonlocal_exit_probe.lang",
		"testdata/phase5/restrict_borrow.lang",
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

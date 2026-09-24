package session_test

import (
	"context"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/session"
)

var phase18RequiredEngines = []string{"interpreter", "O0", "O3", "O3-LTO"}

// phase18CompareAcceptance models the evidence gate: pairwise agreement is
// meaningful only after all required execution tiers and peer admissions ran.
func phase18CompareAcceptance(program core.Program, engines map[string]execution.Execution) error {
	if len(engines) != len(phase18RequiredEngines) {
		return &session.Phase5EngineDisagreement{Fixture: "phase18-acceptance", Axis: "control:required-engines", Detail: "missing execution tier"}
	}
	for _, name := range phase18RequiredEngines {
		if _, ok := engines[name]; !ok {
			return &session.Phase5EngineDisagreement{Fixture: "phase18-acceptance", Axis: "control:required-engines", Detail: "missing " + name}
		}
	}
	return session.Phase5CompareProgramEngines("phase18-acceptance", program, engines)
}

func TestPhase18ComparatorControlRejectsMissingExecutionTierAndRequiresPeer(t *testing.T) {
	program := checkedProgram(t, "testdata", "phase18", "payload_return.lang")
	engines := phase11RunFourTiersWithSupplier(t, context.Background(), program, "return_payload", "Ok", cgen.EmitProgramNativeForTest, "emitProgram")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, program, engines["interpreter"])
	if err := phase18CompareAcceptance(program, engines); err != nil {
		t.Fatalf("complete accepted control did not agree: %v", err)
	}

	missing := make(map[string]execution.Execution, len(engines)-1)
	for name, document := range engines {
		if name != "O3-LTO" {
			missing[name] = document
		}
	}
	if err := phase18CompareAcceptance(program, missing); err == nil {
		t.Fatal("acceptance control passed with the O3-LTO execution tier omitted")
	}

	peerCalls := 0
	restore := session.SetPhase5Schema2PeerValidatorForTest(func(core.Program, execution.Execution) error {
		peerCalls++
		return nil
	})
	err := phase18CompareAcceptance(program, engines)
	restore()
	if err != nil {
		t.Fatalf("peer-observation control failed: %v", err)
	}
	if peerCalls != len(phase18RequiredEngines) {
		t.Fatalf("Schema 2 peer called %d times, want once for each of %d engines", peerCalls, len(phase18RequiredEngines))
	}
}

func TestPhase18ComparatorAxesRejectSeededDivergence(t *testing.T) {
	for _, control := range []struct {
		name   string
		axis   string
		mutate func(*execution.Execution)
	}{
		{"terminal-outcome", session.AxisTerminalOutcome, func(e *execution.Execution) { e.Outcome.Value = "wrong-payload" }},
		{"event-order", session.AxisEventOrder, func(e *execution.Execution) { e.Events[0].TargetPlace = "wrong-place" }},
		{"resource-ledger", session.AxisResourceLedger, func(e *execution.Execution) { e.LiveResources = []string{"unexpected"} }},
		{"exit-status-signal", session.AxisExitStatusSignal, func(e *execution.Execution) { e.ExitSignaled = true; e.ExitSignal = "SIGABRT" }},
	} {
		t.Run(control.name, func(t *testing.T) {
			left := phase5CompareBaseline()
			right := phase5CompareBaseline()
			control.mutate(&right)
			err := session.Phase5CompareEngines("phase18-axis-control", map[string]execution.Execution{"interpreter": left, "O0": right})
			disagreement, ok := err.(*session.Phase5EngineDisagreement)
			if !ok || disagreement.Axis != control.axis {
				t.Fatalf("seeded comparator divergence = %v, want %s", err, control.axis)
			}
		})
	}
}

package session_test

import (
	"context"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/session"
)

// TestPhase15DiamondFrontierMoved preserves the old static-ID collision as
// evidence while requiring /2 to distinguish its two occurrence contexts.
func TestPhase15DiamondFrontierMoved(t *testing.T) {
	const fixture = "multi_function_diamond_call.lang"
	program, entryName := phase11CheckedFixture(t, fixture)
	entry := phase11EntryFunction(t, program, entryName)
	input := phase11EntryInput(t, entry.Parameter.Type)
	engines := phase11RunFourTiers(t, context.Background(), program, entryName, input)
	for name, document := range engines {
		staticIDs := map[string]string{}
		moved := false
		for _, event := range document.Events {
			if prior, ok := staticIDs[event.ID]; ok && prior != event.Invocation {
				moved = true
			}
			staticIDs[event.ID] = event.Invocation
		}
		if !moved {
			t.Fatalf("%s: expected the historical static-ID collision to be separated by invocation", name)
		}
	}
	if err := session.Phase5CompareProgramEngines(fixture, program, engines); err != nil {
		t.Fatalf("moved frontier was not peer-valid across all four tiers: %v", err)
	}
}

// TestPhase15CollisionGuardIsNotInert restores the old collision in one
// document and proves that the /2 peer gate rejects the duplicate pair.
func TestPhase15CollisionGuardIsNotInert(t *testing.T) {
	program, entryName := phase11CheckedFixture(t, "multi_function_diamond_call.lang")
	entry := phase11EntryFunction(t, program, entryName)
	engines := phase11RunFourTiers(t, context.Background(), program, entryName, phase11EntryInput(t, entry.Parameter.Type))
	broken := engines["interpreter"]
	broken.Events = append([]execution.Event(nil), broken.Events...)
	broken.Events = append(broken.Events, broken.Events[0])
	engines["interpreter"] = broken
	if err := session.Phase5CompareProgramEngines("diamond-collision-control", program, engines); err == nil || !strings.Contains(err.Error(), "executionpeer.duplicate_pair") {
		t.Fatalf("restored duplicate occurrence identity was not refused: %v", err)
	}
}

// corruptExecutionProducerInvocation is the producer-side test seam. It is
// deliberately local to this external session test package and works only on
// serialized execution facts; it does not share the session peer seam below.
func corruptExecutionProducerInvocation(document *execution.Execution) {
	document.Events = append([]execution.Event(nil), document.Events...)
	document.Events[0].Invocation = "not-an-invocation"
}

// TestExecutionProducerFaultIsCaughtByPeer corrupts evidence after a real
// producer run while leaving executionpeer intact. The named peer refusal is
// the observable proving producer and peer do not share this derivation.
func TestExecutionProducerFaultIsCaughtByPeer(t *testing.T) {
	program, entryName := phase11CheckedFixture(t, "multi_function_diamond_call.lang")
	entry := phase11EntryFunction(t, program, entryName)
	engines := phase11RunFourTiers(t, context.Background(), program, entryName, phase11EntryInput(t, entry.Parameter.Type))
	broken := engines["interpreter"]
	corruptExecutionProducerInvocation(&broken)
	engines["interpreter"] = broken

	err := session.Phase5CompareProgramEngines("producer-fault", program, engines)
	if err == nil || !strings.Contains(err.Error(), "executionpeer.malformed_invocation") {
		t.Fatalf("producer corruption was not rejected by the intact peer: %v", err)
	}
}

// TestExecutionPeerAcceptanceFaultIsCaughtByControl proves the other fault
// direction with the exact same fixed malformed bytes. First, a peer-boundary
// acceptance fault falsely blesses matching bad documents; after restoration,
// the intact peer rejects those identical bytes.
func TestExecutionPeerAcceptanceFaultIsCaughtByControl(t *testing.T) {
	program, entryName := phase11CheckedFixture(t, "multi_function_diamond_call.lang")
	entry := phase11EntryFunction(t, program, entryName)
	engines := phase11RunFourTiers(t, context.Background(), program, entryName, phase11EntryInput(t, entry.Parameter.Type))
	for name, document := range engines {
		corruptExecutionProducerInvocation(&document)
		engines[name] = document
	}

	// Register cleanup before enabling the peer-only fault so a failure cannot
	// leak it to a later test.
	restore := session.SetPhase5Schema2PeerValidatorForTest(func(_ core.Program, _ execution.Execution) error { return nil })
	t.Cleanup(restore)
	if err := session.Phase5CompareProgramEngines("peer-fault", program, engines); err != nil {
		t.Fatalf("peer acceptance corruption did not create false acceptance: %v", err)
	}
	restore()
	if err := session.Phase5CompareProgramEngines("peer-fault", program, engines); err == nil || !strings.Contains(err.Error(), "executionpeer.malformed_invocation") {
		t.Fatalf("restored peer did not reject the exact corrupted documents: %v", err)
	}
}

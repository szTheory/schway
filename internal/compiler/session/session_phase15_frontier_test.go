package session_test

import (
	"context"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
)

// TestPhase15DiamondFrontierIsPinned records the pre-/2 diagnostic which
// Phase 15 producer plans must move: shared leaf activations still collide on
// static IDs, and native admission refuses the resulting document.
func TestPhase15DiamondFrontierIsPinned(t *testing.T) {
	const fixture = "multi_function_diamond_call.lang"
	program, entryName := phase11CheckedFixture(t, fixture)
	entry := phase11EntryFunction(t, program, entryName)
	input := phase11EntryInput(t, entry.Parameter.Type)
	interpreted, err := interp.Run(program, entryName, input)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	duplicated := false
	for _, event := range interpreted.Events {
		duplicated = duplicated || seen[event.ID]
		seen[event.ID] = true
	}
	if !duplicated {
		t.Fatal("expected pre-/2 duplicate static event IDs; Phase 15 must move this frontier")
	}
	cSource, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatal(err)
	}
	for _, optimization := range []string{"-O0", "-O3"} {
		_, err := native.DefaultRunner().Run(context.Background(), cSource, optimization, []string{input})
		if err == nil || !strings.Contains(err.Error(), "native.invalid_execution") {
			t.Fatalf("%s: expected native invalid-execution duplicate-ID refusal, got %v", optimization, err)
		}
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

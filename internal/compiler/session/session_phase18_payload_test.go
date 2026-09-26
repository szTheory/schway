package session_test

import (
	"context"
	"errors"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
)

func TestPhase18PayloadPlaceReturn(t *testing.T) {
	program := checkedProgram(t, "testdata", "phase18", "payload_return.lang")
	engines := phase11RunFourTiersWithSupplier(t, context.Background(), program, "return_payload", "Ok", cgen.EmitProgramNativeForTest, "emitProgram")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, program, engines["interpreter"])
	for name, got := range engines {
		if got.Outcome.Kind != "returned" || got.Outcome.Value != "Ok:01020304" {
			t.Fatalf("%s outcome = %+v, want returned payload serialization Ok:01020304", name, got.Outcome)
		}
	}
	if err := session.Phase5CompareProgramEngines("testdata/phase18/payload_return.lang:Ok", program, engines); err != nil {
		t.Fatalf("five-axis payload-return comparison: %v", err)
	}
}

func TestPhase18LongPayloadPlaceReturn(t *testing.T) {
	const tag = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	if len(tag) != 152 {
		t.Fatalf("long alternative name length = %d, want 152", len(tag))
	}
	program := checkedProgram(t, "testdata", "phase18", "long_payload_return.lang")
	engines := phase11RunFourTiersWithSupplier(t, context.Background(), program, "return_payload", tag, cgen.EmitProgramNativeForTest, "emitProgram")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, program, engines["interpreter"])
	for name, got := range engines {
		if got.Outcome.Kind != "returned" || got.Outcome.Value != tag+":01020304" {
			t.Fatalf("%s outcome = %+v, want returned payload serialization %s:01020304", name, got.Outcome, tag)
		}
	}
	if err := session.Phase5CompareProgramEngines("testdata/phase18/long_payload_return.lang:"+tag, program, engines); err != nil {
		t.Fatalf("five-axis long payload-return comparison: %v", err)
	}
}

func TestPhase18WrongSlotMutation(t *testing.T) {
	program := checkedProgram(t, "testdata", "phase18", "payload_return.lang")
	restore := cgen.SetPayloadSlotSwapForTest(true)
	mutated := phase11RunFourTiersWithSupplier(t, context.Background(), program, "return_payload", "Ok", cgen.EmitProgramNativeForTest, "emitProgram mutated")
	restore()
	if injected := cgen.PayloadSlotSwapInjectedWriteCount(); injected < 1 {
		t.Fatalf("wrong-slot mutation injected %d writes, want a positive count", injected)
	}
	mutated["interpreter"] = phase16ProjectInterpreterSchema2(t, program, mutated["interpreter"])
	err := session.Phase5CompareProgramEngines("testdata/phase18/payload_return.lang:Ok(mutated)", program, mutated)
	var disagreement *session.Phase5EngineDisagreement
	if !errors.As(err, &disagreement) || disagreement.Axis != session.AxisTerminalOutcome {
		t.Fatalf("mutated program disagreement = %v, want exact %s divergence", err, session.AxisTerminalOutcome)
	}

}

func TestPhase18LongTagWrongSlotMutation(t *testing.T) {
	const tag = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	program := checkedProgram(t, "testdata", "phase18", "long_payload_return.lang")
	restore := cgen.SetPayloadSlotSwapForTest(true)
	mutated := phase11RunFourTiersWithSupplier(t, context.Background(), program, "return_payload", tag, cgen.EmitProgramNativeForTest, "emitProgram long-tag mutated")
	restore()
	if injected := cgen.PayloadSlotSwapInjectedWriteCount(); injected < 1 {
		t.Fatalf("long-tag wrong-slot mutation injected %d writes, want a positive count", injected)
	}
	mutated["interpreter"] = phase16ProjectInterpreterSchema2(t, program, mutated["interpreter"])
	err := session.Phase5CompareProgramEngines("testdata/phase18/long_payload_return.lang:"+tag+"(mutated)", program, mutated)
	var disagreement *session.Phase5EngineDisagreement
	if !errors.As(err, &disagreement) || disagreement.Axis != session.AxisTerminalOutcome {
		t.Fatalf("mutated long-tag program disagreement = %v, want exact %s divergence", err, session.AxisTerminalOutcome)
	}
}

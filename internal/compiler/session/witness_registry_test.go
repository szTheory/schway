// witness_registry_test.go holds plan 14-07's executed probes -- the
// D-14-23 "executed-probe" witness kind, in the flesh. Each probe asserts
// a STRUCTURAL fact (a diagnostic code, an admission refusal, a
// still-identical structural summary), never a message string, per
// D-14-32's accepted-residual-friction note: a probe over-fit to current
// diagnostic prose would go red on a benign rewording, which is not the
// signal this register exists to catch.
package session_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestB1BlameIsStructurallyUnreachable backs PHASE-13-DEBT.md's D-13-02b
// row (and, sharing the same root cause, D-13-10a's withdrawn
// use_matching_argument row): it compiles a real fixture whose declared
// ReturnType contradicts its own declared Parameter.Type -- the exact
// shape B1's contract-violation blame would need to survive admission to
// ever fire on -- and asserts it is refused BY NAME
// (type.return_mismatch) at admission, before any interprocedural pass
// could see it. If this fixture ever checks clean, B1 blame has become
// reachable and this probe goes red (XPASS), forcing a human to regrade
// D-13-02b and D-13-10a rather than letting the claim decay silently.
func TestB1BlameIsStructurallyUnreachable(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "blame_unreachable_admission_refusal.lang"))
	if err != nil {
		t.Fatal(err)
	}
	result := session.Check(source)
	if len(result.Diagnostics) == 0 {
		t.Fatal("D-13-02b's claim requires this fixture to be refused at admission; it checked clean instead -- B1 blame has become reachable, regrade PHASE-13-DEBT.md's D-13-02b and D-13-10a rows")
	}
	found := false
	for _, d := range result.Diagnostics {
		if d.Code == "type.return_mismatch" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a type.return_mismatch diagnostic (the admission precondition D-13-02b names), got: %+v", result.Diagnostics)
	}
}

// TestD1243ControlIsUnconstructible backs PHASE-12-DEBT.md's D-12-43 row:
// it seeds a real, type-safe wrong-slot payload write via
// cgen.SetPayloadSlotSwapForTest and asserts the resulting divergence is
// STILL invisible to every axis session.Phase5CompareEngines checks --
// the decisive value-divergence control D-12-38 wanted is still
// unconstructible against the current representation. Unlike
// TestPayloadSlotSwapMutationKilled (which logs either outcome and always
// passes), this probe FAILS the moment the mutation becomes visible,
// because that is exactly the day D-12-43's claim needs regrading rather
// than continuing to pass on a stale premise.
func TestD1243ControlIsUnconstructible(t *testing.T) {
	ctx := context.Background()
	runner := native.DefaultRunner()
	restore := cgen.SetPayloadSlotSwapForTest(true)
	defer restore()

	program := checkedProgram(t, "testdata", "phase12", "payload_tracer.lang")
	engines := runFunctionOkExecutions(t, ctx, program, "identity", runner)
	injected := cgen.PayloadSlotSwapInjectedWriteCount()
	if injected < 1 {
		t.Fatalf("D-12-43 probe: the fault-injection seam injected no wrong-slot write (count=%d), so this run proves nothing about the control -- check whether payload_tracer.lang's data type still declares two payload-carrying alternatives", injected)
	}
	if compareErr := session.Phase5CompareEngines("payload_tracer.lang(D-12-43 probe)", engines); compareErr != nil {
		t.Fatalf("D-12-43's decisive wrong-slot value-divergence control has become CONSTRUCTIBLE: %v -- this claim is stale, regrade PHASE-12-DEBT.md's D-12-43 row instead of treating this failure as something to silence", compareErr)
	}
}

// TestLTOInertnessOnMultiFunctionEmission backs PHASE-14-DEBT.md's D-14-45
// row (EVD-07): it checks a real multi-function fixture where one
// function (toggle) has a core.Match body -- exactly the shape
// cgen.emitProgram refuses -- and asserts native emission refuses BY NAME
// ("multi-function branch bodies are not supported by native emission
// this phase"). `-flto`'s whole-program optimizer tier has no
// cross-function boundary to exploit when cgen never emits more than one
// function's worth of a program containing a branch body at all; this
// probe is the executed half of that claim.
func TestLTOInertnessOnMultiFunctionEmission(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "multi_function_match_refusal.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture must check clean: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) < 2 {
		t.Fatalf("fixture must be genuinely multi-function (>= 2 functions), got %d", len(checked.Program.Functions))
	}
	_, emitErr := cgen.EmitNative(checked.Program)
	if emitErr == nil {
		t.Fatal("D-14-45: expected multi-function native emission to refuse a Match-bodied function inside a multi-function program; it succeeded instead -- the -flto multi-function inertness claim may no longer hold, regrade PHASE-14-DEBT.md's D-14-45 row")
	}
	if !strings.Contains(emitErr.Error(), "multi-function branch bodies are not supported by native emission this phase") {
		t.Fatalf("expected the refusal to name the multi-function branch-body restriction, got: %v", emitErr)
	}
}

// TestRetainedPointerEscapeIsStillUnsubjected is the probe
// debtRegisterEscapeRegistry's "callback-invocation-unsubjected" entry
// names (D-14-27): it asserts the retained_pointer NAT-03 row is STILL
// Subjected: false with the same EscapeID -- the structural fact the
// escape declares. If a future plan discharges the escape (subjects the
// row), this probe fails, forcing the escape registry entry to be
// retired rather than left citing a stale premise.
func TestRetainedPointerEscapeIsStillUnsubjected(t *testing.T) {
	var target *session.NAT03Mutation
	for _, row := range session.NAT03Mutations() {
		row := row
		if row.ControlID == "control:native.sanitize.retained_pointer" {
			target = &row
		}
	}
	if target == nil {
		t.Fatal("no NAT-03 row declares control:native.sanitize.retained_pointer")
	}
	if target.Subjected {
		t.Fatal("control:native.sanitize.retained_pointer is now Subjected: true -- the escape has been discharged; retire debtRegisterEscapeRegistry's callback-invocation-unsubjected entry instead of leaving this probe passing on a stale premise")
	}
	if target.EscapeID != "escape:callback-invocation-unsubjected" {
		t.Fatalf("expected EscapeID escape:callback-invocation-unsubjected, got %q", target.EscapeID)
	}
}

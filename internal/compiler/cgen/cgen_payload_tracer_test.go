package cgen_test

import (
	"context"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPayloadTracerThreeEngineAgreement is Phase 12 Plan 02's own tracer
// proof (D-12-01 shape, mirroring TestPhase5ByPointerLoweringThreeEngineAgreement):
// session.RunNativeFile already asserts interpreter/-O0/-O3 agreement
// internally (returning an EngineMismatch error on divergence), so a clean
// run here proves the payload-carrying alternative this plan adds --
// constructed, bound, and re-constructed on a single path -- is
// semantically correct end to end, not merely syntactically distinct. The
// second assertion below proves the run actually reached the payload
// emission path, rather than passing vacuously on a program that never
// exercised it.
func TestPayloadTracerThreeEngineAgreement(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase12", "payload_tracer.lang")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if len(diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", diagnostics)
	}
	if err != nil {
		t.Fatalf("engines disagreed or native run failed: %v", err)
	}
	if len(result.Interpreter) == 0 || len(result.O0.Pairs) == 0 || len(result.O3.Pairs) == 0 {
		t.Fatal("expected at least one execution from every engine")
	}
	if !strings.Contains(result.CSource, "unsigned char tag;") {
		t.Fatalf("expected the emitted C to declare the payload struct's tag field, got:\n%s", result.CSource)
	}
}

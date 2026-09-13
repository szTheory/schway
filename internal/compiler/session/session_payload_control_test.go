package session_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestPayloadLayoutMutationRefused proves control:payload.layout_mismatch
// (D-12-37): compiling the generated payload conformance unit
// (cgen.EmitPayloadConformance) against the transposed frozen fixture
// testdata/phase12/payload_layout_mismatch.golden.c is refused at compile
// time under the project's existing -Werror flag set, reporting the
// distinct "native.conformance_failed" code -- exactly mirroring
// TestLayoutMutationIsCompileTimeRefusal's own shape for the foreign-contract
// layout control.
//
// This control is NECESSARY for the C-side layout obligation and
// STRUCTURALLY INCAPABLE of catching a wrong-slot READ for a correct tag --
// see PayloadLayoutMutationRunner's own doc comment. D-12-38's
// TestPayloadSlotSwapMutationKilled (below) is the control that covers that
// gap.
func TestPayloadLayoutMutationRefused(t *testing.T) {
	runner := session.PayloadLayoutMutationRunner{
		Runner:      native.DefaultRunner(),
		DataType:    session.PayloadProbeDataType(),
		FixturePath: testsupport.ProjectPath("testdata", "phase12", "payload_layout_mismatch.golden.c"),
	}
	err := runner.Run(context.Background())
	if err == nil {
		t.Fatal("expected the transposed payload fixture to be refused at compile time")
	}
	var toolErr *native.ToolError
	if !errors.As(err, &toolErr) || toolErr.Code != "native.conformance_failed" {
		t.Fatalf("expected native.conformance_failed, got %v", err)
	}
}

// TestPayloadLayoutMutationAttacksFrozenFixtureOnly companions the refusal
// test above: pointing FixturePath at a correctly-ordered (untransposed)
// private header -- written to a throwaway temp file, never committed to
// testdata, mirroring TestLayoutMutationAttacksFrozenFixtureOnly's own
// shape -- compiles cleanly, proving the runner's verdict depends solely on
// FixturePath's own content rather than always refusing regardless of
// input. The runner type itself carries no field of a generated-source
// shape (Runner, DataType, FixturePath only).
func TestPayloadLayoutMutationAttacksFrozenFixtureOnly(t *testing.T) {
	correctPath := filepath.Join(t.TempDir(), "lang_payload_layout_probe_correct.h")
	correctHeader := []byte(`#ifndef LANG_PAYLOAD_LAYOUT_PROBE_PRIVATE_H
#define LANG_PAYLOAD_LAYOUT_PROBE_PRIVATE_H
typedef struct PayloadProbe_payload {
  unsigned char tag;
  unsigned char field_First;
  unsigned char field_Second;
} PayloadProbe_payload;
#endif
`)
	if err := os.WriteFile(correctPath, correctHeader, 0o600); err != nil {
		t.Fatal(err)
	}
	runner := session.PayloadLayoutMutationRunner{Runner: native.DefaultRunner(), DataType: session.PayloadProbeDataType(), FixturePath: correctPath}
	if err := runner.Run(context.Background()); err != nil {
		t.Fatalf("expected the untransposed fixture to conform, got %v", err)
	}
	value := reflect.ValueOf(runner)
	if value.NumField() != 3 {
		t.Fatalf("PayloadLayoutMutationRunner grew an unexpected field: %+v", runner)
	}
}

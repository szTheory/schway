package session_test

import (
	"context"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
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

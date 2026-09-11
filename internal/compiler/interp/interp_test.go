package interp

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// interpProjectRoot mirrors nat03ProjectRoot's own technique
// (internal/compiler/session/session_phase5_alias.go): a
// runtime.Caller(0)-anchored resolution, avoiding a testsupport import
// (testsupport pulls in session, and session imports interp -- an interp
// test importing testsupport would be a cycle).
func interpProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// checkedCallBasicProgram checks and corevalidates testdata/phase07/call_basic.lang
// directly (syntax.Parse + check.Program + corevalidate.Validate), never
// through package session -- session imports interp, so an interp test
// importing session would be a cycle.
func checkedCallBasicProgram(t *testing.T) core.Program {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", "phase07", "call_basic.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse: unexpected diagnostics: %v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check: unexpected diagnostics: %v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected: %v", validated.Problems)
	}
	return validated.Program()
}

// TestOpCallGroupedArmMutationKilled is Task 3 Test 3 (D-07-39/D-07-41,
// QLT-08): interp's dedicated core.OpCall arm in runLinear returns the
// named ErrCallUnsupported. With that arm replaced -- through the
// unexported opCallGroupedArmForTest seam -- by the SAME grouped behaviour
// OpCopy uses (a plain value copy, no callee ever run), the
// recognized-not-executed assertion goes red: Run no longer returns
// ErrCallUnsupported for main, and instead returns a successful
// "returned" outcome, exactly as if the call had silently succeeded. This
// is the concrete evidence that a control asserting only "no error" (or
// "any error") -- rather than the EXACT named error -- would certify this
// stub.
//
// Four-beat body (pathoracle_test.go:268-295's precedent): assert clean,
// save/override/defer-restore, assert an observable effect, assert the
// specific code.
func TestOpCallGroupedArmMutationKilled(t *testing.T) {
	program := checkedCallBasicProgram(t)
	runMain := func() (Execution, error) { return Run(program, "main", "7") }

	// Beat 1: assert clean. The real, unmutated arm returns the named
	// ErrCallUnsupported for main (which contains a core.OpCall).
	if _, err := runMain(); !errors.Is(err, ErrCallUnsupported) {
		t.Fatalf("expected the clean (unmutated) run to return ErrCallUnsupported, got: %v", err)
	}

	// Beat 2: override, restored via defer.
	previous := opCallGroupedArmForTest
	opCallGroupedArmForTest = true
	defer func() { opCallGroupedArmForTest = previous }()

	// Beat 3: assert an observable effect. The mutated run no longer
	// errors at all -- it returns a successful outcome, exactly as D-07-39
	// warns a grouped-arm fold would.
	execution, err := runMain()
	if err != nil {
		t.Fatalf("expected the mutated run to succeed (folding OpCall into the grouped copy arm), got error: %v", err)
	}

	// Beat 4: assert the specific code -- the mutated run's outcome is
	// "returned" (a plain successful outcome), never ErrCallUnsupported,
	// proving the recognized-not-executed assertion is genuinely load-
	// bearing rather than incidentally true.
	if execution.Outcome.Kind != "returned" {
		t.Fatalf("expected the mutated run's outcome kind to be %q, got %q", "returned", execution.Outcome.Kind)
	}
}

// TestCallExecutesAcrossOneFrame is Task 1's tracer test (SEM-08, OWN-05b,
// D-10-21/D-10-26): call_basic.lang's main calls identity across a real
// heap frame boundary and gets back identity's own returned value, with at
// least one emitted event attributed to the callee's own function ID
// (D-10-32) -- proof the callee frame genuinely ran rather than being
// faked by a pass-through. The same program run twice must produce
// byte-identical CanonicalBytes (D-10-26): every ordered output the frame
// stack produces comes from an ordered slice, never a Go map range.
func TestCallExecutesAcrossOneFrame(t *testing.T) {
	program := checkedCallBasicProgram(t)

	var calleeID string
	for _, function := range program.Functions {
		if function.Name == "identity" {
			calleeID = function.ID
		}
	}
	if calleeID == "" {
		t.Fatalf("call_basic.lang's checked program has no function named %q", "identity")
	}

	execution, err := Run(program, "main", "7")
	if err != nil {
		t.Fatalf("Run(main, %q) returned an unexpected error: %v", "7", err)
	}
	if execution.Outcome.Kind != "returned" {
		t.Fatalf("expected outcome kind %q, got %q", "returned", execution.Outcome.Kind)
	}
	if execution.Outcome.Value != "7" {
		t.Fatalf("expected the callee's own returned value %q, got %q", "7", execution.Outcome.Value)
	}

	foundCalleeEvent := false
	for _, event := range execution.Events {
		if event.FunctionID == calleeID {
			foundCalleeEvent = true
			break
		}
	}
	if !foundCalleeEvent {
		t.Fatalf("expected at least one event attributed to the callee's own function ID %q; events: %+v", calleeID, execution.Events)
	}

	firstBytes, err := CanonicalBytes(execution)
	if err != nil {
		t.Fatalf("CanonicalBytes: %v", err)
	}
	secondExecution, err := Run(program, "main", "7")
	if err != nil {
		t.Fatalf("second Run(main, %q) returned an unexpected error: %v", "7", err)
	}
	secondBytes, err := CanonicalBytes(secondExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (second run): %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("expected two runs of the same program to produce byte-identical CanonicalBytes, got:\n%s\nvs\n%s", firstBytes, secondBytes)
	}
}

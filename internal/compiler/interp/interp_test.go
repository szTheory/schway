package interp

import (
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

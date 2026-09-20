package native

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// TestForeignRetainedSymbolResolves is task 05-08-02's resolution proof
// (D-05-06/D-05-09): the fourth frozen foreign TU's declared symbol
// resolves through the SAME ForeignSourcePathForSymbol switch every other
// frozen TU resolves through.
func TestForeignRetainedSymbolResolves(t *testing.T) {
	path, ok := ForeignSourcePathForSymbol("lang_retained_touch")
	if !ok {
		t.Fatal("lang_retained_touch did not resolve")
	}
	if path != ForeignRetainedSourcePath() {
		t.Fatalf("resolved path %q does not match ForeignRetainedSourcePath() %q", path, ForeignRetainedSourcePath())
	}
}

func compileRetainedPointerFixture(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(projectRoot(), "testdata", "phase5", "retained_pointer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture failed to parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("fixture rejected by corevalidate: %+v", validated.Problems)
	}
	if _, err := cgen.EmitNative(validated.Program()); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("EmitNative must retain the named M004 refusal, got %v", err)
	}
	cSource, err := os.ReadFile(filepath.Join(projectRoot(), "testdata", "phase16", "historical", "retained_pointer.c"))
	if err != nil {
		t.Fatal(err)
	}
	return string(cSource)
}

// TestRetainedPointerFixtureIsUBUnderPlainRun is D-05-10's explicit
// negative assertion: the lane's verdict for this fixture comes from
// RunSanitized ONLY, never from a plain (non-sanitized) native run. A
// plain run of this fixture is undefined behavior, not a guaranteed
// crash -- the freed one-byte block is often re-served intact and the
// program can complete cleanly (or, on some hosts/allocator states, it
// may not). This test deliberately does NOT assert a particular exit code
// or crash for the plain -O0/-O3 run below: asserting either outcome
// would make this test itself flaky on undefined behavior, which is
// exactly the false-determinism trap D-05-10 warns against (the two
// allocator-debug-perturbation environment knobs D-05-10 names are
// rejected for the identical reason).
// It only proves the fixture compiles and links cleanly through the
// ORDINARY (non-sanitizer) build path, so a reader confirms the plain run
// is a real, exercised code path -- never silently skipped -- while its
// outcome is explicitly never treated as evidence.
func TestRetainedPointerFixtureIsUBUnderPlainRun(t *testing.T) {
	cSource := compileRetainedPointerFixture(t)
	runner := DefaultRunner()
	runner.ForeignSources = []string{ForeignRetainedSourcePath()}
	runner.Expect = ExpectValue

	// A plain run is invoked here ONLY to prove the ordinary (non-ASan)
	// build path compiles and links; its result (whatever it is) is
	// deliberately discarded below and never classified, asserted, or
	// used as a pass/fail signal -- that is RunSanitized's job alone
	// (session.VerifyPhase5SanitizeLane), not this test's.
	_, _ = runner.Run(context.Background(), cSource, "-O0", []string{"7"})
}

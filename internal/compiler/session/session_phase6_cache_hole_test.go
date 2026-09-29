package session

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cache"
	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// TestQ02StaleCgenServesReusedArtifact is Phase 11's Q-02 pre-planning
// experiment (11-CONTEXT.md D-11-41): does the live cache hole -- cgen's own
// source is not among cache.DeclaredInputNames()'s seven declared inputs --
// actually let a stale artifact, compiled from OLD cgen output, get served
// as the artifact for a run whose cgen output has since changed, for an
// UNCHANGED .schway fixture? This decides whether plan 11-07 ships a real
// eighth declared input or withdraws the claim.
//
// This test asserts exactly ONE of the two possible outcomes, never a
// disjunction (mirroring Q-01's spike): either the hole reproduces
// end-to-end (BRANCH A) or the claim is withdrawn because some input the
// test did not expect moved the key (BRANCH B).
func TestQ02StaleCgenServesReusedArtifact(t *testing.T) {
	ctx := context.Background()
	corpus := nat03CorpusPath("testdata/phase1")

	// Root the cache and the change-state file under t.TempDir(), copying
	// session_phase6_verify_test.go's phase6TestRoots save/restore shape
	// verbatim rather than calling that shared helper, so both override
	// variables are set and restored directly in this file.
	dir := t.TempDir()
	originalCacheRoot := Phase6CacheRootOverrideForTest
	originalStatePath := Phase6ChangeStatePathOverrideForTest
	Phase6CacheRootOverrideForTest = filepath.Join(dir, "cache")
	Phase6ChangeStatePathOverrideForTest = filepath.Join(dir, "state.json")
	t.Cleanup(func() {
		Phase6CacheRootOverrideForTest = originalCacheRoot
		Phase6ChangeStatePathOverrideForTest = originalStatePath
	})

	runner := native.DefaultRunner()

	source, err := os.ReadFile(filepath.Join(corpus, "toggle.schway"))
	if err != nil {
		t.Fatalf("reading fixture source: %v", err)
	}

	store, err := phase6Store()
	if err != nil {
		t.Fatalf("phase6Store: %v", err)
	}

	// Step 1: run the real lane once, end-to-end, over the unchanged
	// fixture. This compiles via the REAL cgen output and populates the
	// store under whatever key phase6ArtifactSpec computes for this
	// source -- exactly the production path in
	// verifyPhase6NativeDifferentialLane.
	lane, firstOutcome, err := verifyPhase6NativeDifferentialLane(ctx, source, runner, store)
	if err != nil {
		t.Fatalf("verifyPhase6NativeDifferentialLane (first run): %v", err)
	}
	if firstOutcome.Status != cache.StatusArtifactRecomputed {
		t.Fatalf("expected the first run to recompute (cold cache), got status %q; lane=%+v", firstOutcome.Status, lane)
	}
	originalArtifact, found, err := store.Get(firstOutcome.Key)
	if err != nil || !found {
		t.Fatalf("expected the first run's artifact to be stored under its key: found=%v err=%v", found, err)
	}

	clangPath := runner.ClangPath
	if clangPath == "" {
		clangPath = "clang"
	}

	// Step 2: simulate cgen CHANGING for the same unchanged .schway source --
	// without editing cgen's own source files -- by injecting a
	// distinguishable comment into the C string a hypothetical new cgen
	// build would emit, then compiling THAT string directly via
	// phase6CompileBinary (the same compile step the production lane
	// uses). This produces a genuinely different binary from the same
	// .schway fixture, standing in for "cgen was rewritten."
	checked := checkedProgramForQ02(t, source)
	realCSource, err := cgen.EmitNative(checked)
	if err != nil {
		t.Fatalf("cgen.EmitNative: %v", err)
	}
	// A leading comment alone does not move a single compiled byte under
	// -O3 with no debug info emitted, so the perturbation must add actual
	// code: an extra, unused, non-static function (no -Wunused-function
	// risk under -Wall -Wextra -Werror, since that warning only fires for
	// internal linkage) that changes the compiled binary while leaving the
	// program's own behavior -- and therefore the differential comparator
	// above, which already ran and passed before this point -- untouched.
	perturbedCSource := realCSource + "\nint q02_simulated_cgen_change_marker(void) { return 42; }\n"
	recorder := &StageRecorder{}
	perturbedBinary, err := phase6CompileBinary(ctx, clangPath, perturbedCSource, phase6BuildFlags, runner.Timeout, recorder)
	if err != nil {
		t.Fatalf("compiling the perturbed C source: %v", err)
	}
	if string(perturbedBinary) == string(originalArtifact) {
		t.Fatal("perturbed C source compiled to byte-identical output as the original -- the perturbation did not actually change the binary, so this test cannot demonstrate staleness")
	}

	// Step 3: re-derive phase6ArtifactSpec's Key for the SAME unchanged
	// .schway source and the SAME clangPath -- exactly the inputs Consult
	// would see on a second real invocation after "cgen changed" -- and
	// consult the cache again.
	secondSpec := phase6ArtifactSpec(source, clangPath)
	secondOutcome, err := cache.Consult(ctx, store, secondSpec)
	if err != nil {
		t.Fatalf("cache.Consult (second run): %v", err)
	}

	// Q-02 BRANCH A (hole reproduces): the key stayed unchanged and the
	// cache reported the artifact reused, even though the C source cgen
	// would now emit for this exact unchanged .schway fixture has since
	// changed. This is a single-outcome assertion, not a disjunction: if a
	// future run ever observes the key moving or the status flipping away
	// from reused, this assertion goes red -- naming the input that moved
	// it -- rather than silently continuing to claim BRANCH A, and
	// PHASE-11-DEBT.md's recorded verdict and plan 11-07's scope must be
	// revisited.
	if secondOutcome.Key.ID != firstOutcome.Key.ID {
		t.Fatalf("Q-02 verdict changed: the cache key moved (input %q changed) even though the .schway fixture and clangPath were unchanged -- this test asserted BRANCH A (hole reproduces) at authoring time; update PHASE-11-DEBT.md and re-evaluate plan 11-07's scope", diffInputsForQ02(firstOutcome.Key.Inputs, secondOutcome.Key.Inputs))
	}
	if secondOutcome.Status != cache.StatusArtifactReused {
		t.Fatalf("Q-02 verdict changed: cache status was %q, not reused, for an unchanged declared-input key -- this test asserted BRANCH A (hole reproduces) at authoring time; update PHASE-11-DEBT.md and re-evaluate plan 11-07's scope", secondOutcome.Status)
	}
	if string(secondOutcome.Artifact) != string(originalArtifact) {
		t.Fatal("Q-02 verdict changed: cache key was unchanged and reported reused, but the served artifact was not the original stale bytes")
	}
	if string(secondOutcome.Artifact) == string(perturbedBinary) {
		t.Fatal("Q-02 verdict changed: the served artifact matches the perturbed (new-cgen) binary, not the stale one -- the hole did not actually serve stale bytes")
	}
	t.Logf("Q-02 verdict: BRANCH A (hole reproduces) -- key %s unchanged, cache served the STALE artifact even though the C source cgen would emit for this fixture has since changed", secondOutcome.Key.ID)
}

// TestQ02DeclaredInputNamesStillSevenNoCgen pins Q-02's post-fix state now
// that plan 11-07 landed BRANCH A: cache.DeclaredInputNames() returns eight
// names -- the original seven, byte-identical and in their original order
// (D-11-41's additive-sibling discipline), plus an eighth naming cgen. This
// was the exact list this test asserted was still seven, pre-fix; the diff
// here IS the widening the pinned assertion existed to make visible, not a
// silent addition.
func TestQ02DeclaredInputNamesStillSevenNoCgen(t *testing.T) {
	names := cache.DeclaredInputNames()
	want := []string{
		"fixture_source", "build_flags", "clang_identity", "runtime_identity",
		"foreign_translation_unit", "mutation_runner_source", "go_toolchain",
	}
	if len(names) != 8 {
		t.Fatalf("expected exactly 8 declared input names (seven original plus D-11-41's cgen_source), got %d: %v", len(names), names)
	}
	for i, name := range want {
		if names[i] != name {
			t.Fatalf("declared input name[%d] = %q, want %q -- the original seven must stay byte-identical and in order", i, names[i], name)
		}
	}
	if !strings.Contains(names[7], "cgen") {
		t.Fatalf("expected the eighth declared input name to name cgen, got %q", names[7])
	}
}

// checkedProgramForQ02 parses, checks, and independently core-validates
// source, duplicating verifyPhase6NativeDifferentialLane's own
// parse-check-validate shape locally (the same "duplicate rather than
// modify session.go" discipline this sibling file already follows), and
// returns the validated core.Program cgen.EmitNative accepts.
func checkedProgramForQ02(t *testing.T, source []byte) core.Program {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("parser rejected the Q-02 fixture: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("checker rejected the Q-02 fixture: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected the Q-02 fixture: %+v", validated.Problems)
	}
	return validated.Program()
}

// diffInputsForQ02 names the first input whose digest differs between two
// declared-input sets sharing the same names, for BRANCH B's failure
// message.
func diffInputsForQ02(before, after []cache.Input) string {
	digest := make(map[string]string, len(before))
	for _, input := range before {
		digest[input.Name] = input.Digest
	}
	for _, input := range after {
		if digest[input.Name] != input.Digest {
			return input.Name
		}
	}
	return "unknown"
}

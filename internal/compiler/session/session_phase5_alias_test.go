package session_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// falseRestrictHoistPath is testdata/phase5/false_restrict_hoist.lang's
// stable path, read once per test.
func falseRestrictHoistPath() string {
	return testsupport.ProjectPath("testdata", "phase5", "false_restrict_hoist.lang")
}

func falseRestrictHoistSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(falseRestrictHoistPath())
	if err != nil {
		t.Fatal(err)
	}
	return source
}

// TestFalseRestrictFixtureIsCleanUnmutated is D-05-05's own Task 1
// falsifier: testdata/phase5/false_restrict_hoist.lang must be accepted by
// check/corevalidate unmutated, must derive ZERO alias facts (a proven-
// exclusive parameter would make an injected restrict TRUE rather than
// FALSE), must carry the by-pointer marker exactly once (the mutation
// runner's single target, D-05-05/Task 2) with no `*restrict` qualifier
// anywhere in its unmutated generated C, and must agree across
// interpreter/-O0/-O3 via Phase5CompareEngines.
func TestFalseRestrictFixtureIsCleanUnmutated(t *testing.T) {
	source := falseRestrictHoistSource(t)

	if !strings.Contains(string(source), "// adversarial-target: false-no-alias-hoist") {
		t.Fatal("fixture is missing its adversarial-target header")
	}

	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture failed to parse: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	if len(checked.AliasFacts) != 0 {
		t.Fatalf("want zero alias facts (checker must NOT prove exclusivity), got %+v", checked.AliasFacts)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("fixture rejected by corevalidate: %+v", validated.Problems)
	}

	nativeResult, diagnostics, err := session.RunNative(context.Background(), source, native.DefaultRunner())
	if err != nil {
		t.Fatalf("RunNative error: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics from RunNative: %+v", diagnostics)
	}

	if count := strings.Count(nativeResult.CSource, "*restrict "); count != 0 {
		t.Fatalf("want zero `*restrict` qualifiers in unmutated generated C, got %d:\n%s", count, nativeResult.CSource)
	}
	if count := strings.Count(nativeResult.CSource, "/* lang:by-pointer-param */"); count != 1 {
		t.Fatalf("want exactly one by-pointer-param marker, got %d:\n%s", count, nativeResult.CSource)
	}

	// session.RunNative already asserts interpreter==O0==O3 internally
	// (returning an *session.EngineMismatch otherwise); this additionally
	// exercises Phase5CompareEngines by name, per this task's own literal
	// acceptance criterion.
	if len(nativeResult.Interpreter) != 1 || len(nativeResult.O0.Pairs) != 1 || len(nativeResult.O3.Pairs) != 1 {
		t.Fatalf("expected exactly one execution per engine, got interpreter=%d o0=%d o3=%d",
			len(nativeResult.Interpreter), len(nativeResult.O0.Pairs), len(nativeResult.O3.Pairs))
	}
	engines := map[string]execution.Execution{
		"interpreter": nativeResult.Interpreter[0],
		"o0":          nativeResult.O0.Pairs[0].Execution,
		"o3":          nativeResult.O3.Pairs[0].Execution,
	}
	if err := session.Phase5CompareEngines(falseRestrictHoistPath(), engines); err != nil {
		t.Fatalf("unmutated fixture disagrees across engines: %v", err)
	}
}

// TestAliasFactMutationIsDetected proves control:alias.false_no_alias's
// happy path: injecting restrict onto false_restrict_hoist.lang's unproven
// parameter produces D-05-05's exact signature, interpreter == -O0 != -O3,
// and the runner records at least one optimization attempt plus the clang
// identity string used to build it.
func TestAliasFactMutationIsDetected(t *testing.T) {
	runner := session.NewAliasFactMutationRunner(native.DefaultRunner(), falseRestrictHoistPath())
	if err := session.VerifyAliasFalseNoAlias(context.Background(), runner); err != nil {
		t.Fatalf("expected the alias mutation to be detected cleanly, got: %v", err)
	}
	if runner.RecomputedWork() == 0 {
		t.Fatal("expected the runner to have recorded at least one optimization attempt")
	}
	if runner.ClangVersion() == "" {
		t.Fatal("expected the runner to have recorded a clang identity string")
	}
}

// TestAliasMutationMarkerCountGuard proves the fail-closed marker-count
// guard every sibling mutation runner in this package establishes: zero
// and two matching markers each refuse to mutate, rather than attacking an
// absent or ambiguous target.
func TestAliasMutationMarkerCountGuard(t *testing.T) {
	runner := session.NewAliasFactMutationRunner(native.DefaultRunner(), falseRestrictHoistPath())

	t.Run("zero markers", func(t *testing.T) {
		_, err := runner.Run(context.Background(), "static unsigned char touch(unsigned char *param) { return *param; }", "-O0", nil)
		assertAliasControlInvalid(t, err, "0, want 1")
	})

	t.Run("two markers", func(t *testing.T) {
		source := strings.Join([]string{
			"static unsigned char touch(unsigned char *param) { /* lang:by-pointer-param */",
			"  return *param;",
			"}",
			"static unsigned char touch2(unsigned char *param) { /* lang:by-pointer-param */",
			"  return *param;",
			"}",
		}, "\n")
		_, err := runner.Run(context.Background(), source, "-O0", nil)
		assertAliasControlInvalid(t, err, ">1, want 1")
	})
}

func assertAliasControlInvalid(t *testing.T, err error, wantSubstring string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected an alias control error, got nil")
	}
	var toolError *native.ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.alias_control_invalid" {
		t.Fatalf("expected native.alias_control_invalid, got %v", err)
	}
	if !strings.Contains(toolError.Err.Error(), wantSubstring) {
		t.Fatalf("expected error to contain %q, got %v", wantSubstring, toolError.Err)
	}
}

// TestAliasMutationWithoutDivergenceFailsTheLane is D-05-05's load-bearing
// negative case: restrict_borrow.lang's parameter IS proven exclusive (a
// legitimately-justified restrict, no dormant alias-probe parameter), so
// running it through the SAME runner produces no divergence at all —
// VerifyAliasFalseNoAlias must FAIL the lane, naming the control as
// unexercised, rather than silently pass.
func TestAliasMutationWithoutDivergenceFailsTheLane(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.lang")
	runner := session.NewAliasFactMutationRunner(native.DefaultRunner(), path)
	err := session.VerifyAliasFalseNoAlias(context.Background(), runner)
	if err == nil {
		t.Fatal("expected a no-divergence error, got nil")
	}
	if !strings.Contains(err.Error(), session.ControlAliasFalseNoAlias) {
		t.Fatalf("expected the error to name %s, got: %v", session.ControlAliasFalseNoAlias, err)
	}
}

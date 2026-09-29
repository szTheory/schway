package native_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase5LTOFixtureSource checks and lowers testdata/phase5/inline_across_foreign.schway
// (05-05's own D-05-18a fixture, engineered so a single foreign acquisition's
// result feeds the function's own return -- exactly the shape under which
// `-flto` legally lets Clang inline the separately-compiled foreign TU's tiny
// body at the call site, D-05-24). This package (native_test, external, like
// native_conformance_test.go's existing precedent) is the one place in this
// directory allowed to import session/cgen without an import cycle, since
// session itself imports native.
func phase5LTOFixtureSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "inline_across_foreign.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	// This fixture is deliberately outside the Phase 16 admitted surface.
	// Keep the current public refusal live, then run the historical runner
	// control from the digest-pinned artifact instead of recreating it through
	// a second lowering authority.
	if _, err := cgen.EmitNative(checked.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("EmitNative must retain the named M004 refusal, got %v", err)
	}
	cSource, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "historical", "inline_across_foreign.c"))
	if err != nil {
		t.Fatal(err)
	}
	return string(cSource)
}

// nonPathTokens filters every argument that looks like a filesystem path
// (contains os.PathSeparator) out of a recorded command line, leaving only
// the literal flag tokens in construction order. Every path this Runner
// constructs (source/object/binary paths) lives under a fresh
// os.MkdirTemp directory and is therefore non-deterministic across Run
// calls -- filtering it out is what makes "the recorded argument vector
// matches the pre-change literals element-wise" (D-05-19's acceptance
// criterion) a stable, repeatable assertion rather than a flaky one.
func nonPathTokens(line []string) []string {
	tokens := make([]string, 0, len(line))
	for _, argument := range line {
		if strings.ContainsRune(argument, os.PathSeparator) {
			continue
		}
		tokens = append(tokens, argument)
	}
	return tokens
}

func stringSlicesEqual(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

// TestLTOFlagReachesCompileAndLink proves `-flto` reaches BOTH the per-TU
// foreign compile argument vector and the final link argument vector when
// Runner.LTO is true -- asserting only one of the two would leave the tier
// silently inert on whichever command line went unchecked (D-05-19).
func TestLTOFlagReachesCompileAndLink(t *testing.T) {
	cSource := phase5LTOFixtureSource(t)
	runner := native.DefaultRunner().EnableCommandRecording()
	runner.LTO = true
	runner.ForeignSources = []string{native.ForeignResourceSourcePath()}
	if _, err := runner.Run(context.Background(), cSource, "-O3", []string{"7"}); err != nil {
		t.Fatalf("expected the LTO build+run to succeed, got %v", err)
	}
	lines := runner.LastCommandLines()
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 recorded command lines (foreign compile + link), got %d: %v", len(lines), lines)
	}
	for _, line := range lines {
		found := false
		for _, argument := range line {
			if argument == "-flto" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected -flto in recorded command line, missing from %v", line)
		}
	}
}

// TestLTODisabledLeavesCommandLineUnchanged proves that with LTO false,
// every command line this Runner constructs (the foreign compile and the
// final link) has the identical literal flag shape it had before D-05-19 --
// gating on the LTO field alone must never perturb the pre-Phase-5
// differential's own invocations (D-05-19).
func TestLTODisabledLeavesCommandLineUnchanged(t *testing.T) {
	cSource := phase5LTOFixtureSource(t)
	runner := native.DefaultRunner().EnableCommandRecording()
	runner.ForeignSources = []string{native.ForeignResourceSourcePath()}
	if _, err := runner.Run(context.Background(), cSource, "-O0", []string{"7"}); err != nil {
		t.Fatalf("expected the non-LTO build+run to succeed, got %v", err)
	}
	lines := runner.LastCommandLines()
	if len(lines) != 2 {
		t.Fatalf("expected exactly 2 recorded command lines (foreign compile + link), got %d: %v", len(lines), lines)
	}

	wantForeignCompile := []string{"clang", "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0", "-c", "-o"}
	wantLink := []string{"clang", "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O0", "-o"}

	gotForeignCompile := nonPathTokens(lines[0])
	if !stringSlicesEqual(gotForeignCompile, wantForeignCompile) {
		t.Fatalf("foreign compile command line changed shape:\ngot  %v\nwant %v", gotForeignCompile, wantForeignCompile)
	}
	gotLink := nonPathTokens(lines[1])
	if !stringSlicesEqual(gotLink, wantLink) {
		t.Fatalf("link command line changed shape:\ngot  %v\nwant %v", gotLink, wantLink)
	}
	for _, argument := range lines[0] {
		if argument == "-flto" {
			t.Fatal("expected no -flto in the foreign compile line when LTO is false")
		}
	}
	for _, argument := range lines[1] {
		if argument == "-flto" {
			t.Fatal("expected no -flto in the link line when LTO is false")
		}
	}
}

// TestLTOTierIsNotInert is D-05-38's mutation-kill for this lane: build the
// same two-TU program once with LTO false and once with LTO true and assert
// the resulting binaries differ. A lane whose LTO and non-LTO outputs are
// byte-identical is not exercising LTO at all -- it merely builds, which is
// exactly the false-green shape D-05-24 exists to prevent. This uses
// testdata/phase5/inline_across_foreign.schway, engineered by 05-05
// specifically because its cross-TU inlining is an observable codegen
// difference under -flto (05-05-SUMMARY.md).
func TestLTOTierIsNotInert(t *testing.T) {
	cSource := phase5LTOFixtureSource(t)

	nonLTO := native.DefaultRunner().EnableCommandRecording()
	nonLTO.ForeignSources = []string{native.ForeignResourceSourcePath()}
	if _, err := nonLTO.Run(context.Background(), cSource, "-O3", []string{"7"}); err != nil {
		t.Fatalf("non-LTO -O3 build failed: %v", err)
	}

	withLTO := native.DefaultRunner().EnableCommandRecording()
	withLTO.LTO = true
	withLTO.ForeignSources = []string{native.ForeignResourceSourcePath()}
	if _, err := withLTO.Run(context.Background(), cSource, "-O3", []string{"7"}); err != nil {
		t.Fatalf("LTO -O3 build failed: %v", err)
	}

	nonLTOBinary := nonLTO.LastBinary()
	ltoBinary := withLTO.LastBinary()
	if len(nonLTOBinary) == 0 || len(ltoBinary) == 0 {
		t.Fatalf("expected both binaries to be recorded, got %d and %d bytes", len(nonLTOBinary), len(ltoBinary))
	}
	if bytes.Equal(nonLTOBinary, ltoBinary) {
		t.Fatal("expected the -O3 LTO and non-LTO binaries to differ in codegen (cross-TU inlining), got byte-identical binaries -- the LTO tier is inert")
	}
}

// --- NAT-07 composition-only LTO divergence control (D-11-23/D-11-24/Q-04, plan 11-02) ---
//
// This is a hand-written-C control, NOT emitted by cgen (D-11-24): it exists
// solely to prove the -O3/-flto interprocedural optimizer TIER can be
// exploited at all, before any cgen work on multi-function emission begins.
// native.Runner's own single program.c path is unchanged; the extra
// translation units below are supplied entirely through the existing
// ForeignSources mechanism.
//
// Topology (D-11-23's coordinated two-site mutation across THREE
// translation units, empirically discovered on this host by direct
// experimentation -- "measured, not asserted"):
//
//  1. compositionOnlyCalleeSource: the callee TU. Its one function,
//     schway_composition_probe_read, carries the FALSE `restrict` qualifier
//     on its by-pointer parameter (site 1).
//  2. compositionOnlyWriterSource: a THIRD, entirely separate TU. Its one
//     function, schway_composition_probe_write, performs the aliasing write
//     (site 2). It has no idea a restrict promise exists anywhere.
//  3. compositionOnlyWrapperSource: the coordination point. It calls the
//     callee TWICE -- once before, once after the writer's call -- through
//     its OWN by-pointer parameter (which repeats the restrict qualifier
//     when the config calls for it). Neither call to the callee alone, nor
//     the intervening write alone, is unsound. Only a compiler that INLINES
//     BOTH calls to the callee (which requires seeing across the callee's
//     own translation-unit boundary -- i.e. -flto) has any chance of
//     treating the two reads as one cacheable access and skipping the
//     reload across the write. Critically, the wrapper itself must live in
//     its OWN, fourth translation unit (never merged into the same
//     compilation as main): folding wrapper+main into a single TU let the
//     compiler prove the two pointer arguments are trivially identical and
//     correctly refuse the hoist at every tier including -flto -- the
//     divergence requires the genuine cross-TU-inlining boundary.
//  4. The caller (native.Runner's own program.c, supplied as cSource):
//     binds probe == primary (or a genuinely distinct object, for the
//     non-aliasing configurations) and calls the wrapper.
//
// Three injection configurations, each its own row:
//   - "none":           wrapper's primary carries NO restrict; probe DOES
//     alias primary. Baseline: the write is honestly observed.
//   - "restrict-only":  wrapper's primary carries restrict; probe does NOT
//     alias primary (points at a distinct object). No violation exists, so
//     nothing to exploit.
//   - "restrict+write": wrapper's primary carries restrict; probe DOES
//     alias primary. The genuine composition-only violation.
//
// Each row is compared against its OWN -O0 result (never a single matrix-
// wide reference): "none" and "restrict-only" must stay internally
// consistent (green) at every tier, and "restrict+write" must diverge from
// its own -O0 value at EXACTLY one tier: -O3 with -flto.
const compositionOnlyCalleeSource = `unsigned char schway_composition_probe_read(unsigned char *restrict primary) {
  return *primary;
}
`

const compositionOnlyCalleePlainSource = `unsigned char schway_composition_probe_read(unsigned char *primary) {
  return *primary;
}
`

const compositionOnlyWriterSource = `void schway_composition_probe_write(unsigned char *probe) {
  *probe = 99;
}
`

// compositionOnlyWrapperSource renders the coordination-point translation
// unit. restrictQualifier controls whether BOTH the extern declaration of
// the callee and the wrapper's own by-pointer parameter carry the false
// `restrict` promise -- the two must agree, or `-Wall -Wextra -Werror
// -pedantic` (native.Runner's fixed compile flags) reject the mismatched
// declaration outright.
func compositionOnlyWrapperSource(restrictQualifier bool) string {
	qualifier := ""
	if restrictQualifier {
		qualifier = "restrict "
	}
	return fmt.Sprintf(`extern unsigned char schway_composition_probe_read(unsigned char *%[1]sprimary);
extern void schway_composition_probe_write(unsigned char *probe);
unsigned char schway_composition_probe_wrapper(unsigned char *%[1]sprimary, unsigned char *probe) {
  unsigned char before = schway_composition_probe_read(primary);
  schway_composition_probe_write(probe);
  unsigned char after = schway_composition_probe_read(primary);
  return (unsigned char)(before + after);
}
`, qualifier)
}

// compositionOnlyMainSource renders the caller translation unit --
// native.Runner's own program.c -- for one (restrict, aliased) cell.
// aliased controls whether probe is bound to the SAME object as primary
// (the actual violation input) or a genuinely distinct object (the
// restrict-only row's non-violating input). The program prints a minimal,
// schema-valid `schway.execution/0` document so native.Runner's own
// decodeExecution accepts it -- this hand-written control still goes
// through the SAME execute-and-decode path as every other native.Runner
// caller, per D-11-24 (native.Runner itself is never widened).
func compositionOnlyMainSource(restrictQualifier, aliased bool) string {
	qualifier := ""
	if restrictQualifier {
		qualifier = "restrict "
	}
	probeExpr := "&value"
	preamble := ""
	if !aliased {
		preamble = "  unsigned char other = 0;\n"
		probeExpr = "&other"
	}
	return fmt.Sprintf(`#include <stdio.h>
#include <stdlib.h>
extern unsigned char schway_composition_probe_wrapper(unsigned char *%[1]sprimary, unsigned char *probe);
int main(int argc, char **argv) {
  if (argc != 2) return 64;
  unsigned char value = (unsigned char)atoi(argv[1]);
%[2]s  unsigned char result = schway_composition_probe_wrapper(&value, %[3]s);
  printf("{\"schema\":\"schway.execution/0\",\"outcome\":{\"kind\":\"returned\",\"value\":\"%%u\"},\"events\":[{\"schema\":\"schway.execution/0\",\"id\":\"e0\",\"kind\":\"function.returned\",\"function_id\":\"fn:probe\",\"input\":\"in\",\"output\":\"out\"}],\"live_resources\":[]}\n", (unsigned int)result);
  return 0;
}
`, qualifier, preamble, probeExpr)
}

// writeCompositionSource writes source to a file named name inside dir and
// returns its absolute path, for use as a native.Runner ForeignSources
// entry.
func writeCompositionSource(t *testing.T, dir, name, source string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
	return path
}

// compositionOnlyCellValue builds and runs one (config, tier) cell through
// native.Runner -- the ONLY compiler invocation path this control uses, per
// D-11-24 -- and returns the decoded outcome value. Skips the whole test if
// the resolved clang toolchain is absent (never on a genuine divergence
// finding, per Q-04's own instruction).
func compositionOnlyCellValue(t *testing.T, calleePath, writerPath, wrapperPath, mainSource, optimization string, lto bool) string {
	t.Helper()
	runner := native.DefaultRunner()
	runner.LTO = lto
	runner.ForeignSources = []string{calleePath, writerPath, wrapperPath}
	result, err := runner.Run(context.Background(), mainSource, optimization, []string{"5"})
	if err != nil {
		var toolErr *native.ToolError
		if errors.As(err, &toolErr) && toolErr.Code == "native.tool_missing" {
			t.Skipf("env:clang toolchain unavailable on this host: %v", err)
		}
		t.Fatalf("composition-only cell (optimization=%s lto=%v) failed to build/run: %v", optimization, lto, err)
	}
	if len(result.Pairs) != 1 {
		t.Fatalf("expected exactly 1 execution pair, got %d", len(result.Pairs))
	}
	return result.Pairs[0].Execution.Outcome.Value
}

// TestCompositionOnlyLTODivergence is NAT-07's engineered composition-only
// negative control (D-11-23/D-11-24/D-11-26, Q-04): it proves the
// interprocedural `-O3`/`-flto` optimizer tier can be exploited at all,
// using hand-written C compiled through the same clang toolchain
// native.Runner already drives -- before cgen ever learns to emit more than
// one function. See the topology comment above compositionOnlyCalleeSource
// for the full design.
func TestCompositionOnlyLTODivergence(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skipf("env:clang toolchain unavailable on this host: %v", err)
	}

	dir := t.TempDir()
	calleeRestrictPath := writeCompositionSource(t, dir, "callee_restrict.c", compositionOnlyCalleeSource)
	calleePlainPath := writeCompositionSource(t, dir, "callee_plain.c", compositionOnlyCalleePlainSource)
	writerPath := writeCompositionSource(t, dir, "writer.c", compositionOnlyWriterSource)
	wrapperRestrictPath := writeCompositionSource(t, dir, "wrapper_restrict.c", compositionOnlyWrapperSource(true))
	wrapperPlainPath := writeCompositionSource(t, dir, "wrapper_plain.c", compositionOnlyWrapperSource(false))

	type compositionConfig struct {
		name        string
		calleePath  string
		wrapperPath string
		restrict    bool
		aliased     bool
	}
	configs := []compositionConfig{
		{name: "none", calleePath: calleePlainPath, wrapperPath: wrapperPlainPath, restrict: false, aliased: true},
		{name: "restrict-only", calleePath: calleeRestrictPath, wrapperPath: wrapperRestrictPath, restrict: true, aliased: false},
		{name: "restrict+write", calleePath: calleeRestrictPath, wrapperPath: wrapperRestrictPath, restrict: true, aliased: true},
	}

	type compositionTier struct {
		name         string
		optimization string
		lto          bool
	}
	tiers := []compositionTier{
		{name: "-O0", optimization: "-O0", lto: false},
		{name: "-O1", optimization: "-O1", lto: false},
		{name: "-O3", optimization: "-O3", lto: false},
		{name: "-O3 -flto", optimization: "-O3", lto: true},
	}

	values := make(map[string]map[string]string, len(configs))
	for _, config := range configs {
		mainSource := compositionOnlyMainSource(config.restrict, config.aliased)
		values[config.name] = make(map[string]string, len(tiers))
		for _, tier := range tiers {
			values[config.name][tier.name] = compositionOnlyCellValue(t, config.calleePath, writerPath, config.wrapperPath, mainSource, tier.optimization, tier.lto)
		}
	}

	// Assertion 4 (and the general lane-failure rule, D-11-21): count every
	// cell that diverges from ITS OWN row's -O0 reference -- never a single
	// matrix-wide reference, since "restrict-only" and "restrict+write" are
	// legitimately different scenarios with different correct answers.
	type divergence struct{ config, tier string }
	var diverging []divergence
	for _, config := range configs {
		reference := values[config.name]["-O0"]
		for _, tier := range tiers {
			if tier.name == "-O0" {
				continue
			}
			if values[config.name][tier.name] != reference {
				diverging = append(diverging, divergence{config: config.name, tier: tier.name})
			}
		}
	}

	if len(diverging) == 0 {
		t.Fatalf("LANE FAILURE (D-11-21): the composition-only control produced an ALL-GREEN matrix -- no cell diverged from its own -O0 reference in any row. An unexercised control proves nothing; this must be reported as an escalation, never as NAT-07 satisfied. Observed values: %+v", values)
	}
	if len(diverging) != 1 {
		t.Fatalf("expected EXACTLY ONE diverging cell, got %d: %+v. Full matrix: %+v", len(diverging), diverging, values)
	}
	if diverging[0].config != "restrict+write" || diverging[0].tier != "-O3 -flto" {
		t.Fatalf("the single diverging cell must be config=restrict+write tier=-O3 -flto (per Q-04, a divergence anywhere else -- especially at -O3 WITHOUT LTO -- means the TU split does not buy LTO exclusivity and the topology must be re-examined); got config=%s tier=%s. Full matrix: %+v", diverging[0].config, diverging[0].tier, values)
	}

	// Assertion 3: every single-injection cell, at every tier, is green --
	// restated explicitly (not merely implied by the count above) since this
	// is exactly what makes the divergence composition-only.
	for _, name := range []string{"none", "restrict-only"} {
		reference := values[name]["-O0"]
		for _, tier := range tiers {
			if values[name][tier.name] != reference {
				t.Fatalf("single-injection config %q diverged at tier %q (want green at every tier): got %q, want %q (matches -O0)", name, tier.name, values[name][tier.name], reference)
			}
		}
	}

	// Assertion 5 (Q-04's own stated branch): a red cell at -O3 WITHOUT LTO
	// for the full-injection config would mean the TU split does not buy
	// LTO exclusivity at all -- restated explicitly for a clear failure
	// message distinct from the generic count-based assertions above.
	if values["restrict+write"]["-O3"] != values["restrict+write"]["-O0"] {
		t.Fatalf("restrict+write diverged at -O3 WITHOUT -flto (got %q, want %q matching -O0) -- the TU split does not buy LTO exclusivity; re-examine the topology (Q-04)", values["restrict+write"]["-O3"], values["restrict+write"]["-O0"])
	}

	t.Logf("composition-only matrix (NAT-07/Q-04): %+v", values)
}

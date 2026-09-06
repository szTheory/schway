package native_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// phase5LTOFixtureSource checks and lowers testdata/phase5/inline_across_foreign.lang
// (05-05's own D-05-18a fixture, engineered so a single foreign acquisition's
// result feeds the function's own return -- exactly the shape under which
// `-flto` legally lets Clang inline the separately-compiled foreign TU's tiny
// body at the call site, D-05-24). This package (native_test, external, like
// native_conformance_test.go's existing precedent) is the one place in this
// directory allowed to import session/cgen without an import cycle, since
// session itself imports native.
func phase5LTOFixtureSource(t *testing.T) string {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "inline_across_foreign.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	cSource, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("EmitNative failed: %v", err)
	}
	return cSource
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
// testdata/phase5/inline_across_foreign.lang, engineered by 05-05
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

package cache

import (
	"bytes"
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func writeExecutableFixture(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "fixture-clang.sh")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeMutationRunnerFixture(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "runner.go")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeCgenSourceFixture materializes a tiny synthetic "cgen source
// directory" containing one .go file with the given content, standing in
// for internal/compiler/cgen/*.go without this test depending on the real
// package's current contents.
func writeCgenSourceFixture(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cgen.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func baseArtifactSpec(t *testing.T) ArtifactSpec {
	t.Helper()
	dir := t.TempDir()
	return ArtifactSpec{
		Kind:                     KindCompiledBinary,
		FixtureSource:            []byte("fn main() {}\n"),
		BuildFlags:               "-O0 target=x86_64-apple-darwin",
		ClangPath:                writeExecutableFixture(t, dir, "#!/bin/sh\necho fixture clang version 1.0\n"),
		RuntimeIdentity:          "none",
		ForeignTranslationUnit:   []byte("// frozen TU v1\n"),
		MutationRunnerSourcePath: writeMutationRunnerFixture(t, dir, "package fake\n// mutation runner fixture v1\n"),
		GoToolchain:              "go1.24.0",
		CgenSourceDir:            writeCgenSourceFixture(t, "package cgen\n// cgen source fixture v1\n"),
	}
}

func TestCacheKeyCoversEveryDeclaredInput(t *testing.T) {
	names := DeclaredInputNames()
	want := []string{
		"fixture_source", "build_flags", "clang_identity", "runtime_identity",
		"foreign_translation_unit", "mutation_runner_source", "go_toolchain",
		"cgen_source",
	}
	if len(names) != len(want) {
		t.Fatalf("DeclaredInputNames() = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("DeclaredInputNames()[%d] = %q, want %q (order is D-06-07's determinism guarantee)", i, names[i], want[i])
		}
	}

	ctx := context.Background()
	base := baseArtifactSpec(t)
	baseInputs, err := InputsFor(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	baseKey, err := ComputeKey(baseInputs)
	if err != nil {
		t.Fatal(err)
	}

	covered := map[string]bool{}
	cases := []struct {
		name   string
		mutate func(t *testing.T, spec *ArtifactSpec)
	}{
		{"fixture_source", func(t *testing.T, spec *ArtifactSpec) {
			spec.FixtureSource = []byte("fn main() { changed() }\n")
		}},
		{"build_flags", func(t *testing.T, spec *ArtifactSpec) {
			spec.BuildFlags = "-O3 target=x86_64-apple-darwin"
		}},
		{"clang_identity", func(t *testing.T, spec *ArtifactSpec) {
			spec.ClangPath = writeExecutableFixture(t, t.TempDir(), "#!/bin/sh\necho fixture clang version 2.0\n")
		}},
		{"runtime_identity", func(t *testing.T, spec *ArtifactSpec) {
			spec.RuntimeIdentity = "asan+ubsan+libc++"
		}},
		{"foreign_translation_unit", func(t *testing.T, spec *ArtifactSpec) {
			spec.ForeignTranslationUnit = []byte("// frozen TU v2\n")
		}},
		{"mutation_runner_source", func(t *testing.T, spec *ArtifactSpec) {
			spec.MutationRunnerSourcePath = writeMutationRunnerFixture(t, t.TempDir(), "package fake\n// mutation runner fixture v2\n")
		}},
		{"go_toolchain", func(t *testing.T, spec *ArtifactSpec) {
			spec.GoToolchain = "go1.99.0"
		}},
		{"cgen_source", func(t *testing.T, spec *ArtifactSpec) {
			spec.CgenSourceDir = writeCgenSourceFixture(t, "package cgen\n// cgen source fixture CHANGED\n")
		}},
	}
	for _, testCase := range cases {
		covered[testCase.name] = true
		t.Run(testCase.name, func(t *testing.T) {
			mutated := base
			testCase.mutate(t, &mutated)
			inputs, err := InputsFor(ctx, mutated)
			if err != nil {
				t.Fatal(err)
			}
			key, err := ComputeKey(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if key.ID == baseKey.ID {
				t.Fatalf("perturbing declared input %q did not change the key", testCase.name)
			}
		})
	}
	// Self-invalidating: a declared input name with no perturbation row above
	// must fail this test, so DeclaredInputNames() can never silently grow
	// past this table's coverage.
	for _, name := range names {
		if !covered[name] {
			t.Fatalf("declared input %q has no perturbation row in this table", name)
		}
	}
}

func TestMutationRunnerSourceHashIsADeclaredInput(t *testing.T) {
	ctx := context.Background()
	base := baseArtifactSpec(t)
	baseInputs, err := InputsFor(ctx, base)
	if err != nil {
		t.Fatal(err)
	}
	baseKey, err := ComputeKey(baseInputs)
	if err != nil {
		t.Fatal(err)
	}

	changed := base
	changed.MutationRunnerSourcePath = writeMutationRunnerFixture(t, t.TempDir(), "package fake\n// mutation runner fixture CHANGED\n")
	changedInputs, err := InputsFor(ctx, changed)
	if err != nil {
		t.Fatal(err)
	}
	changedKey, err := ComputeKey(changedInputs)
	if err != nil {
		t.Fatal(err)
	}
	if baseKey.ID == changedKey.ID {
		t.Fatal("a mutant binary built by a changed mutation runner must miss the cache")
	}
}

func selfExecFactory(mode string) commandFactory {
	return func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestCacheProbeHelper", "--", mode)
		cmd.Env = append(os.Environ(), "GO_WANT_CACHE_PROBE_HELPER=1")
		return cmd
	}
}

func TestClangDigestIsNotJustTheVersionString(t *testing.T) {
	dir := t.TempDir()
	binaryA := filepath.Join(dir, "clang-a")
	binaryB := filepath.Join(dir, "clang-b")
	if err := os.WriteFile(binaryA, []byte("binary-bytes-a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryB, []byte("binary-bytes-b"), 0o644); err != nil {
		t.Fatal(err)
	}

	factory := selfExecFactory("same-version")
	digestA, err := probeClangIdentityWithCommand(context.Background(), binaryA, factory)
	if err != nil {
		t.Fatal(err)
	}
	digestB, err := probeClangIdentityWithCommand(context.Background(), binaryB, factory)
	if err != nil {
		t.Fatal(err)
	}
	if digestA == digestB {
		t.Fatal("two probes reporting the same version but different binary bytes must produce different digests")
	}
}

func TestClangDigestProbeIsBounded(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "fake-clang")
	if err := os.WriteFile(binary, []byte("fake-clang-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, testCase := range []struct {
		name     string
		mode     string
		wantCode string
		timeout  time.Duration
	}{
		{name: "stdout truncated", mode: "stdout", wantCode: "cache.probe_stdout_truncated"},
		{name: "stderr truncated", mode: "stderr", wantCode: "cache.probe_stderr_truncated"},
		{name: "timeout", mode: "block", wantCode: "cache.probe_timeout", timeout: 50 * time.Millisecond},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx := context.Background()
			if testCase.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, testCase.timeout)
				defer cancel()
			}
			_, err := probeClangIdentityWithCommand(ctx, binary, selfExecFactory(testCase.mode))
			var probeErr *Error
			if !errors.As(err, &probeErr) || probeErr.Code != testCase.wantCode {
				t.Fatalf("error=%v want code=%s", err, testCase.wantCode)
			}
		})
	}
}

// TestCacheProbeHelper is a self-exec test helper (evidence's
// TestEvidenceToolProbeHelper precedent): it only does anything when
// launched as a subprocess by selfExecFactory above via os.Args[0].
func TestCacheProbeHelper(t *testing.T) {
	if os.Getenv("GO_WANT_CACHE_PROBE_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	switch mode {
	case "stdout":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", MaxProbeBytes+1))
	case "stderr":
		_, _ = os.Stderr.WriteString(strings.Repeat("x", MaxProbeBytes+1))
	case "block":
		time.Sleep(10 * time.Second)
	default:
		_, _ = os.Stdout.WriteString("fixture clang version 1.0\n")
	}
	os.Exit(0)
}

// ---------------------------------------------------------------------
// QLT-06b's structural discharge (D-11-39): two independent, individually
// falsifiable import scans -- a direct go/parser scan and a transitive
// go list -deps scan -- plus their own negative controls, plus a
// mechanical proof that nothing interprocedural is ever cached. Copies
// originvalidate_test.go's three-part shape verbatim rather than inventing
// a new one (D-10-17's own precedent).
// ---------------------------------------------------------------------

// cacheForbiddenImports is QLT-06b's closed forbidden set: cache is a
// dependency-free leaf (cache.go's own package doc) and must never import
// core, corevalidate or originvalidate to obtain a digest -- doing so
// would let a trust-crossing derivation leak into this package through
// exactly the back door D-11-41's cgen_source input might otherwise tempt
// someone to open (asking core for a "compiled form" digest instead of
// hashing cgen's own source bytes directly).
var cacheForbiddenImports = []string{"/compiler/core", "/compiler/originvalidate", "/compiler/corevalidate"}

// cacheDirectImportViolation mirrors originvalidate_test.go's own
// directImportViolation verbatim: it reads dir's own non-test Go source
// files' import lists via go/parser's ImportsOnly mode (never assumed from
// a doc comment) and returns the first forbidden import found, or "" if
// none.
func cacheDirectImportViolation(t *testing.T, dir string, forbidden []string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			for _, bad := range forbidden {
				if strings.HasSuffix(path, bad) {
					return entry.Name() + " imports " + path
				}
			}
		}
	}
	return ""
}

// cacheMaxGoListDepsOutputBytes and cacheGoListDepsTimeout mirror
// originvalidate_test.go's own bounded-writer/timeout convention for the
// `go list -deps` spawn (D-02-01's spawn-safety lint).
const cacheMaxGoListDepsOutputBytes = 1 << 20
const cacheGoListDepsTimeout = 2 * time.Minute

type cacheBoundedGoListWriter struct {
	buffer     bytes.Buffer
	overflowed bool
}

func (w *cacheBoundedGoListWriter) Write(data []byte) (int, error) {
	if w.buffer.Len()+len(data) > cacheMaxGoListDepsOutputBytes {
		w.overflowed = true
		remaining := cacheMaxGoListDepsOutputBytes - w.buffer.Len()
		if remaining > 0 {
			w.buffer.Write(data[:remaining])
		}
		return len(data), nil
	}
	w.buffer.Write(data)
	return len(data), nil
}

// cacheTransitiveImportViolation is the shared suffix-matching predicate
// BOTH the real `go list -deps` scan and its own negative control
// exercise, so the negative control proves the real predicate can fail,
// not a second, drifting copy of it (mirroring
// transitiveImportViolation's own precedent).
func cacheTransitiveImportViolation(deps []string, forbidden []string) string {
	for _, dep := range deps {
		for _, bad := range forbidden {
			if strings.HasSuffix(dep, bad) {
				return dep
			}
		}
	}
	return ""
}

// cacheTransitiveImportsViolation hardens the direct-import scan above:
// nobody adds a forbidden import to cache on purpose; they add a helper
// package that itself imports one. Bounded at 1 MiB and 2 minutes,
// matching this repo's own spawn-safety convention.
func cacheTransitiveImportsViolation(t *testing.T, forbidden []string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), cacheGoListDepsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "github.com/codename-lang/lang/internal/compiler/cache")
	cmd.Dir = testsupport.ProjectPath()
	stdout := &cacheBoundedGoListWriter{}
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	if stdout.overflowed {
		t.Fatalf("go list -deps produced more than %d bytes; dependency listing is implausibly large", cacheMaxGoListDepsOutputBytes)
	}
	deps := strings.Split(strings.TrimSpace(stdout.buffer.String()), "\n")
	return cacheTransitiveImportViolation(deps, forbidden)
}

// TestDeclaredInputNamesStableAndNoInterproceduralImport is QLT-06b's
// primary discharge test (D-11-39), asserting four independent knowers in
// one place:
//
// Knower 1 -- the declared-name list: the first seven names stay
// byte-identical and in their original order (exact-string, order-
// sensitive: a reordering is a failure, never an equivalent list). Under
// BRANCH A the eighth entry exists and names cgen; under BRANCH B the list
// is still exactly seven.
//
// Knowers 2 and 3 -- the direct go/parser scan AND the transitive
// go list -deps scan, both run (the repo's own convention states the
// direct scan is HARDENED by the transitive one, never replaced by it).
func TestDeclaredInputNamesStableAndNoInterproceduralImport(t *testing.T) {
	names := DeclaredInputNames()
	originalSeven := []string{
		"fixture_source", "build_flags", "clang_identity", "runtime_identity",
		"foreign_translation_unit", "mutation_runner_source", "go_toolchain",
	}
	if len(names) < len(originalSeven) {
		t.Fatalf("expected at least the original seven declared names, got %d: %v", len(names), names)
	}
	for i, want := range originalSeven {
		if names[i] != want {
			t.Fatalf("declared input name[%d] = %q, want %q -- the original seven must stay byte-identical and in order (D-11-41's additive-sibling discipline)", i, names[i], want)
		}
	}
	switch len(names) {
	case len(originalSeven):
		// BRANCH B: the D-11-41 fix was withdrawn; the list stays seven.
	case len(originalSeven) + 1:
		if !strings.Contains(names[len(originalSeven)], "cgen") {
			t.Fatalf("expected the eighth declared input to name cgen, got %q", names[len(originalSeven)])
		}
	default:
		t.Fatalf("expected exactly %d (BRANCH B) or %d (BRANCH A) declared input names, got %d: %v", len(originalSeven), len(originalSeven)+1, len(names), names)
	}

	dir := testsupport.ProjectPath("internal", "compiler", "cache")
	if violation := cacheDirectImportViolation(t, dir, cacheForbiddenImports); violation != "" {
		t.Fatalf("%s, which cache (a dependency-free leaf) must never depend on", violation)
	}
	if violation := cacheTransitiveImportsViolation(t, cacheForbiddenImports); violation != "" {
		t.Fatalf("cache transitively imports %s, which it must never depend on", violation)
	}
}

// TestCacheDirectImportGuardCanFail is the direct scan's own negative
// control: a synthetic file containing a forbidden import must actually be
// flagged, proving the guard can go red, not merely that it has never yet
// found anything.
func TestCacheDirectImportGuardCanFail(t *testing.T) {
	dir := t.TempDir()
	content := "package cache\n\nimport (\n\t\"github.com/codename-lang/lang/internal/compiler/core\"\n)\n\nvar _ = core.Program{}\n"
	if err := os.WriteFile(filepath.Join(dir, "synthetic.go"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if violation := cacheDirectImportViolation(t, dir, cacheForbiddenImports); violation == "" {
		t.Fatal("expected the synthetic forbidden-import file to be flagged")
	}
}

// TestCacheTransitiveImportGuardCanFail is the transitive scan's own
// negative control: a synthetic dependency list containing
// internal/compiler/originvalidate must be flagged.
func TestCacheTransitiveImportGuardCanFail(t *testing.T) {
	synthetic := []string{
		"github.com/codename-lang/lang/internal/compiler/cache",
		"github.com/codename-lang/lang/internal/compiler/originvalidate",
	}
	if got := cacheTransitiveImportViolation(synthetic, cacheForbiddenImports); got == "" {
		t.Fatal("expected the synthetic dependency list's forbidden originvalidate entry to be flagged")
	}
}

// TestNoClosureDigestInCache is Knower 4, the mechanical form of D-11-38's
// "cache nothing new": no PRODUCTION file under internal/compiler/cache
// mentions core.FunctionSignature's interprocedural per-function digest
// field (the "Closure"+"Digest" identifier, deliberately built by
// concatenation immediately below rather than spelled as a literal, so
// this very test's own source does not itself trip a source-text scan for
// that identifier), and no declared input name looks closure- or
// call-graph-derived. A whole-program FixtureSource hash strictly
// dominates any call-graph-closure key in a single-unit language, so a
// closure key here could only ever admit MORE cache hits on strictly LESS
// evidence -- a soundness-loosening change this test exists to prevent by
// construction, not by review.
//
// Scoped to non-test files (nonTestGoFiles's own convention, mirrored
// locally): this test's own name and doc comment necessarily discuss the
// forbidden identifier by name, so a whole-directory literal scan
// (including this very _test.go file) would trip on its own assertion
// machinery. The binding guarantee is that PRODUCTION code never keys on
// it; scanning production sources only is what actually proves that,
// where a blanket scan would produce a permanent, unfixable false
// positive from this test's own name.
func TestNoClosureDigestInCache(t *testing.T) {
	forbiddenIdentifier := "Closure" + "Digest"
	dir := testsupport.ProjectPath("internal", "compiler", "cache")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), forbiddenIdentifier) {
			t.Fatalf("%s mentions the forbidden interprocedural digest identifier -- no interprocedural fact may ever be marked cacheable (D-11-38)", entry.Name())
		}
	}
	for _, name := range DeclaredInputNames() {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "closure") || strings.Contains(lower, "call_graph") || strings.Contains(lower, "callgraph") {
			t.Fatalf("declared input %q looks closure- or call-graph-derived, which cache must never key on (D-11-38)", name)
		}
	}
}

// TestCacheKeyIsIdempotent asserts D-11-39's idempotency truth: computing
// the key twice for identical inputs yields a byte-identical Key.ID, and a
// second Consult on unchanged inputs reports reuse rather than
// recomputing.
func TestCacheKeyIsIdempotent(t *testing.T) {
	ctx := context.Background()
	spec := baseArtifactSpec(t)

	inputsA, err := InputsFor(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	keyA, err := ComputeKey(inputsA)
	if err != nil {
		t.Fatal(err)
	}
	inputsB, err := InputsFor(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	keyB, err := ComputeKey(inputsB)
	if err != nil {
		t.Fatal(err)
	}
	if keyA.ID != keyB.ID {
		t.Fatalf("computing the key twice for identical inputs produced different IDs: %s vs %s", keyA.ID, keyB.ID)
	}

	store := &Store{Root: t.TempDir()}
	firstOutcome, err := Consult(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	if firstOutcome.Status != StatusArtifactRecomputed {
		t.Fatalf("expected the first Consult on an empty store to recompute, got %s", firstOutcome.Status)
	}
	if err := store.Put(firstOutcome.Key, []byte("artifact-bytes")); err != nil {
		t.Fatal(err)
	}

	secondOutcome, err := Consult(ctx, store, spec)
	if err != nil {
		t.Fatal(err)
	}
	if secondOutcome.Status != StatusArtifactReused {
		t.Fatalf("expected the second Consult on unchanged inputs to reuse, got %s", secondOutcome.Status)
	}
	if secondOutcome.Key.ID != firstOutcome.Key.ID {
		t.Fatal("second Consult computed a different key for identical inputs")
	}
}

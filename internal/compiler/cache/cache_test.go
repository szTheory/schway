package cache_test

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cache"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// ---------------------------------------------------------------------
// Tracer: ComputeKey -> Put -> Get -> byte-identical artifact back,
// through the real filesystem.
// ---------------------------------------------------------------------

func TestCacheRoundTripsArtifactAndMeta(t *testing.T) {
	store := &cache.Store{Root: t.TempDir()}

	key, err := cache.ComputeKey([]cache.Input{{Name: "fixture_source", Digest: "abc123"}})
	if err != nil {
		t.Fatal(err)
	}
	artifact := []byte("compiled-binary-bytes")
	if err := store.Put(key, artifact); err != nil {
		t.Fatal(err)
	}

	got, found, err := store.Get(key)
	if err != nil || !found {
		t.Fatalf("Get after Put: found=%v err=%v", found, err)
	}
	if string(got) != string(artifact) {
		t.Fatalf("round trip mismatch: got %q want %q", got, artifact)
	}

	missingKey, err := cache.ComputeKey([]cache.Input{{Name: "fixture_source", Digest: "different"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.Get(missingKey); err != nil || found {
		t.Fatalf("Get for an unwritten key: found=%v err=%v", found, err)
	}

	path := store.Path(key)
	if len(key.ID) < 2 || filepath.Base(filepath.Dir(path)) != key.ID[:2] {
		t.Fatalf("Path's second-to-last element must be the key ID's first two hex chars: %s", path)
	}

	metaBytes, err := os.ReadFile(filepath.Join(path, "meta.json"))
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(metaBytes, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["schema"] != cache.Schema || decoded["id"] != key.ID {
		t.Fatalf("meta.json schema/id mismatch: %+v", decoded)
	}
	inputsRaw, ok := decoded["inputs"].([]any)
	if !ok || len(inputsRaw) != len(key.Inputs) {
		t.Fatalf("meta.json inputs length mismatch: %+v", decoded["inputs"])
	}
}

// TestCachePutInterruptedBeforeRenameLeavesNoPartialEntry asserts the
// acceptance criterion that a crash between the temp write and the rename
// leaves no directory that Get reports as found: a temp-named sibling file
// existing alongside no real "artifact"/"meta.json" file must still read as
// absent.
func TestCachePutInterruptedBeforeRenameLeavesNoPartialEntry(t *testing.T) {
	store := &cache.Store{Root: t.TempDir()}
	key, err := cache.ComputeKey([]cache.Input{{Name: "fixture_source", Digest: "abc123"}})
	if err != nil {
		t.Fatal(err)
	}
	dir := store.Path(key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".tmp-artifact-interrupted"), []byte("partial"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, found, err := store.Get(key); err != nil || found {
		t.Fatalf("Get must report absent before the rename completes: found=%v err=%v", found, err)
	}
}

// ---------------------------------------------------------------------
// Structural enforcement (D-06-06): the package cannot hold a verdict.
// ---------------------------------------------------------------------

// forbiddenIdentifierSubstrings names the judgement-shaped vocabulary this
// package's exported surface may never carry (D-06-06). "status" and
// "outcome" are deliberately NOT here: D-06-12 reuses those words for
// cache-reuse REPORTING (artifact_reused/artifact_recomputed/not_cacheable/
// unavailable) -- a fact about what this invocation skipped, never a
// judgement about whether a lane passed. The forbidden list below targets
// the actual verdict/judgement/diagnostic vocabulary this package must
// never carry.
var forbiddenIdentifierSubstrings = []string{
	"verdict",
	"judgement",
	"judgment",
	"diagnostic",
	"passfail",
	"testresult",
	"checkresult",
	"laneresult",
}

func TestCacheExportedSurfaceStoresNoVerdict(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "cache")
	fileSet := token.NewFileSet()
	for _, path := range nonTestGoFiles(t, dir) {
		file, err := parser.ParseFile(fileSet, path, nil, parser.AllErrors)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch decl := node.(type) {
			case *ast.TypeSpec:
				if decl.Name.IsExported() {
					assertIdentifierClean(t, path, decl.Name.Name)
				}
				if structType, ok := decl.Type.(*ast.StructType); ok && structType.Fields != nil {
					for _, field := range structType.Fields.List {
						for _, name := range field.Names {
							assertIdentifierClean(t, path, name.Name)
						}
						assertTypeClean(t, path, field.Type)
					}
				}
			case *ast.FuncDecl:
				if decl.Name.IsExported() {
					assertIdentifierClean(t, path, decl.Name.Name)
				}
			case *ast.ValueSpec:
				for _, name := range decl.Names {
					if name.IsExported() {
						assertIdentifierClean(t, path, name.Name)
					}
				}
			}
			return true
		})
	}
}

// TestOutcomeFieldSetIsPinned closes the one hole the identifier denylist
// cannot: "outcome" and "status" are excluded from
// forbiddenIdentifierSubstrings because this package's own exported
// Outcome/CacheStatus names use them, so a LATER field such as
// Outcome.LaneOutcome or Outcome.CheckStatus would pass
// TestCacheExportedSurfaceStoresNoVerdict unnoticed. Pinning the exact
// field set makes any addition to Outcome a deliberate, reviewed act
// rather than a silent one -- the same self-invalidating-enumeration
// discipline 06-01 used for the result-identity struct.
//
// Outcome may only ever report WHAT THIS INVOCATION SKIPPED. If a future
// field would report whether a lane passed, that is D-06-06's one-way
// violation and this test is where it must stop.
func TestOutcomeFieldSetIsPinned(t *testing.T) {
	want := []string{"Artifact", "Key", "Status"}
	got := exportedFieldNames(t, "Outcome")
	if len(got) != len(want) {
		t.Fatalf("cache.Outcome field set changed: got %v, want exactly %v -- adding a field to Outcome requires re-reading D-06-06 (artifacts, never verdicts)", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("cache.Outcome field set changed: got %v, want exactly %v -- adding a field to Outcome requires re-reading D-06-06 (artifacts, never verdicts)", got, want)
		}
	}
}

// exportedFieldNames parses this package's non-test sources and returns the
// declared field names of typeName, in declaration order. It fails when the
// type is absent, so a rename cannot make the pin above vacuously pass.
func exportedFieldNames(t *testing.T, typeName string) []string {
	t.Helper()
	dir := testsupport.ProjectPath("internal", "compiler", "cache")
	fileSet := token.NewFileSet()
	var names []string
	found := false
	for _, path := range nonTestGoFiles(t, dir) {
		file, err := parser.ParseFile(fileSet, path, nil, parser.AllErrors)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			spec, ok := node.(*ast.TypeSpec)
			if !ok || spec.Name.Name != typeName {
				return true
			}
			structType, ok := spec.Type.(*ast.StructType)
			if !ok || structType.Fields == nil {
				return true
			}
			found = true
			for _, field := range structType.Fields.List {
				for _, name := range field.Names {
					names = append(names, name.Name)
				}
			}
			return false
		})
	}
	if !found {
		t.Fatalf("type %s not found in internal/compiler/cache -- the field-set pin cannot vacuously pass", typeName)
	}
	return names
}

func assertIdentifierClean(t *testing.T, path, name string) {
	t.Helper()
	lower := strings.ToLower(name)
	for _, forbidden := range forbiddenIdentifierSubstrings {
		if strings.Contains(lower, forbidden) {
			t.Fatalf("%s: exported identifier %q contains judgement-shaped substring %q -- cache.go must never hold a verdict (D-06-06)", path, name, forbidden)
		}
	}
}

func assertTypeClean(t *testing.T, path string, expr ast.Expr) {
	t.Helper()
	if ident, ok := expr.(*ast.Ident); ok {
		assertIdentifierClean(t, path, ident.Name)
	}
}

// TestCacheImportsStayIndependent asserts, in the style of corevalidate's
// own TestValidatorImportsStayIndependent, that this package's Go import
// list (parsed from source, not assumed) never imports
// internal/compiler/protocol, internal/compiler/diagnostic, or
// internal/compiler/session -- the exact seam a smuggled-in verdict type
// would have to travel through.
func TestCacheImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "cache")
	fileSet := token.NewFileSet()
	forbidden := []string{"/compiler/protocol", "/compiler/diagnostic", "/compiler/session"}
	for _, path := range nonTestGoFiles(t, dir) {
		file, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			importPath := strings.Trim(imported.Path.Value, `"`)
			for _, suffix := range forbidden {
				if strings.HasSuffix(importPath, suffix) {
					t.Fatalf("%s imports %s, which internal/compiler/cache must never depend on", path, importPath)
				}
			}
		}
	}
}

// ---------------------------------------------------------------------
// Task 3: fail-closed not_cacheable, adjacency, and the cold empty store.
// ---------------------------------------------------------------------

func newArtifactSpecFixture(t *testing.T, seed string) cache.ArtifactSpec {
	t.Helper()
	dir := t.TempDir()
	clangPath := filepath.Join(dir, "fixture-clang.sh")
	if err := os.WriteFile(clangPath, []byte("#!/bin/sh\necho fixture clang version 1.0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	runnerPath := filepath.Join(dir, "runner.go")
	if err := os.WriteFile(runnerPath, []byte("package fake\n// mutation runner fixture: "+seed+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cgenDir := filepath.Join(dir, "cgen")
	if err := os.MkdirAll(cgenDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cgenDir, "cgen.go"), []byte("package cgen\n// cgen source fixture: "+seed+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return cache.ArtifactSpec{
		Kind:                     cache.KindCompiledBinary,
		FixtureSource:            []byte("fn main() { " + seed + "() }\n"),
		BuildFlags:               "-O0 target=x86_64-apple-darwin",
		ClangPath:                clangPath,
		RuntimeIdentity:          "none",
		ForeignTranslationUnit:   []byte("// frozen TU: " + seed + "\n"),
		MutationRunnerSourcePath: runnerPath,
		GoToolchain:              "go1.24.0",
		CgenSourceDir:            cgenDir,
	}
}

func TestCacheStatusVocabularyIsClosed(t *testing.T) {
	statuses := cache.CacheStatuses()
	if len(statuses) != 4 {
		t.Fatalf("expected exactly 4 cache statuses, got %d: %v", len(statuses), statuses)
	}
	allowed := map[string]bool{
		"artifact_reused": true, "artifact_recomputed": true,
		"not_cacheable": true, "unavailable": true,
	}
	for _, status := range statuses {
		if !allowed[status] {
			t.Fatalf("status %q is not one of the four D-06-12 values", status)
		}
		if status == "hit" || status == "miss" {
			t.Fatalf("status vocabulary must never contain hit/miss terminology (D-06-12): it would wrongly imply a verdict was cached")
		}
	}
}

func TestCacheFailsClosedToNotCacheable(t *testing.T) {
	ctx := context.Background()
	root := filepath.Join(t.TempDir(), "does-not-exist-yet")
	store := &cache.Store{Root: root}

	failingSpec := newArtifactSpecFixture(t, "probe-fail")
	failingSpec.ClangPath = filepath.Join(t.TempDir(), "no-such-clang-binary")
	outcome, err := cache.Consult(ctx, store, failingSpec)
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Status != cache.StatusNotCacheable || outcome.Key.ID != "" {
		t.Fatalf("failing Clang probe: got status=%s key=%+v, want StatusNotCacheable with a zero Key", outcome.Status, outcome.Key)
	}
	if _, statErr := os.Stat(root); !os.IsNotExist(statErr) {
		t.Fatalf("Consult must perform no store lookup when not_cacheable: root now exists (stat err=%v)", statErr)
	}

	outcome, err = cache.Consult(ctx, store, cache.ArtifactSpec{Kind: "not-a-real-kind"})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Status != cache.StatusNotCacheable || outcome.Key.ID != "" {
		t.Fatalf("unclassified kind: got status=%s key=%+v, want StatusNotCacheable with a zero Key", outcome.Status, outcome.Key)
	}
}

func TestCacheEqualDeclaredInputsShareOneEntry(t *testing.T) {
	ctx := context.Background()
	specA := newArtifactSpecFixture(t, "shared-seed")
	specB := newArtifactSpecFixture(t, "shared-seed")

	inputsA, err := cache.InputsFor(ctx, specA)
	if err != nil {
		t.Fatal(err)
	}
	inputsB, err := cache.InputsFor(ctx, specB)
	if err != nil {
		t.Fatal(err)
	}
	keyA, err := cache.ComputeKey(inputsA)
	if err != nil {
		t.Fatal(err)
	}
	keyB, err := cache.ComputeKey(inputsB)
	if err != nil {
		t.Fatal(err)
	}
	if keyA.ID != keyB.ID {
		t.Fatalf("byte-identical declared inputs must share one key: %s vs %s", keyA.ID, keyB.ID)
	}

	specC := newArtifactSpecFixture(t, "different-seed")
	inputsC, err := cache.InputsFor(ctx, specC)
	if err != nil {
		t.Fatal(err)
	}
	keyC, err := cache.ComputeKey(inputsC)
	if err != nil {
		t.Fatal(err)
	}
	if keyA.ID == keyC.ID {
		t.Fatal("a spec differing in a declared input must produce a different key (miss)")
	}

	store := &cache.Store{Root: t.TempDir()}
	artifact := []byte("shared-artifact-bytes")
	if err := store.Put(keyA, artifact); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.Get(keyB)
	if err != nil || !found {
		t.Fatalf("second Get for the merged entry: found=%v err=%v", found, err)
	}
	if string(got) != string(artifact) {
		t.Fatalf("merged entry returned different bytes: got %q want %q", got, artifact)
	}
}

func TestCacheEmptyStoreReportsCold(t *testing.T) {
	ctx := context.Background()
	store := &cache.Store{Root: t.TempDir()}
	specs := []cache.ArtifactSpec{
		newArtifactSpecFixture(t, "cold-1"),
		newArtifactSpecFixture(t, "cold-2"),
		{Kind: "unclassified-kind"},
	}
	for index, spec := range specs {
		outcome, err := cache.Consult(ctx, store, spec)
		if err != nil {
			t.Fatalf("case %d: %v", index, err)
		}
		if outcome.Status == cache.StatusArtifactReused {
			t.Fatalf("case %d: a fresh empty store must never report reuse, got %s", index, outcome.Status)
		}
		if outcome.Status != cache.StatusArtifactRecomputed && outcome.Status != cache.StatusNotCacheable {
			t.Fatalf("case %d: unexpected status %s", index, outcome.Status)
		}
	}
}

func TestCacheMismatchedMetaIsTreatedAsAbsent(t *testing.T) {
	store := &cache.Store{Root: t.TempDir()}
	key, err := cache.ComputeKey([]cache.Input{{Name: "fixture_source", Digest: "abc"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(key, []byte("artifact-bytes")); err != nil {
		t.Fatal(err)
	}

	dir := store.Path(key)
	tampered := map[string]any{
		"schema": cache.Schema,
		"id":     key.ID,
		"inputs": []map[string]string{{"name": "fixture_source", "digest": "TAMPERED"}},
	}
	encoded, err := json.Marshal(tampered)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, found, err := store.Get(key); err != nil || found {
		t.Fatalf("meta.json recording a mismatched input list must be treated as absent: found=%v err=%v", found, err)
	}
}

func TestCacheArtifactDigestMismatchIsTreatedAsAbsent(t *testing.T) {
	store := &cache.Store{Root: t.TempDir()}
	key, err := cache.ComputeKey([]cache.Input{{Name: "fixture_source", Digest: "abc"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(key, []byte("trusted-artifact")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(store.Path(key), "artifact"), []byte("tampered-artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, found, err := store.Get(key); err != nil || found {
		t.Fatalf("artifact digest mismatch must be a cache miss: found=%v err=%v", found, err)
	}
}

func nonTestGoFiles(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	return files
}

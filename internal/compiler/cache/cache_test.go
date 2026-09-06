package cache_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
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

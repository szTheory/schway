package check

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------
// Phase 09 Plan 02, Task 1 (D-09-46): the call-graph corpus generator was
// relocated from this package (costcorpus_test.go) to
// internal/compiler/testsupport/callgraphcorpus.go so session's and
// corevalidate's own test binaries could reach it without granting either a
// production import of check. These two tests follow the relocated
// generator to its new home and guard against it drifting back into a
// second copy. Both freely import os/path/filepath -- unlike
// costcorpus_test.go's own TestCostCorpusIsNotParsed, which forbids "os" in
// ITS OWN file specifically because that file must never itself read .schway
// source from disk; this file's job is scanning OTHER files' text, which is
// a different concern entirely.
// ---------------------------------------------------------------------

// relocatedCallGraphCorpusPath is the generator's new home, relative to
// this package's own directory.
const relocatedCallGraphCorpusPath = "../testsupport/callgraphcorpus.go"

// TestRelocatedCallGraphCorpusIsNotParsed extends D-08-34's structural
// guard to the generator's new home: internal/compiler/testsupport's own
// callgraphcorpus.go must never import internal/compiler/syntax or os,
// exactly like costcorpus_test.go's own TestCostCorpusIsNotParsed asserted
// before the relocation -- the "never generated as .schway source through the
// real parser" guarantee must follow the generator wherever it lives.
func TestRelocatedCallGraphCorpusIsNotParsed(t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, relocatedCallGraphCorpusPath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parsing %s for its own import scan: %v", relocatedCallGraphCorpusPath, err)
	}

	forbidden := map[string]bool{
		"github.com/szTheory/schway/internal/compiler/syntax": true,
		"os": true,
	}

	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		spec, ok := n.(*ast.ImportSpec)
		if !ok {
			return true
		}
		path, unquoteErr := strconv.Unquote(spec.Path.Value)
		if unquoteErr != nil {
			t.Fatalf("could not unquote import path %s: %v", spec.Path.Value, unquoteErr)
		}
		if forbidden[path] {
			found = append(found, path)
		}
		return true
	})
	if len(found) != 0 {
		t.Fatalf("%s imports forbidden package(s) %v -- the relocated cost corpus must never be generated as .schway source through the real parser (D-08-34)", relocatedCallGraphCorpusPath, found)
	}
}

// TestCallGraphCorpusGeneratorHasOneHome is D-09-46/T-09-07's own
// mutation-kill: a static scan asserting the identifier "func corpusRelay"
// appears in exactly one .go file across internal/compiler/, so a future
// duplication of the generator (rather than a relocation) fails loudly
// instead of silently reintroducing two-derivations-one-truth for the test
// fixtures.
func TestCallGraphCorpusGeneratorHasOneHome(t *testing.T) {
	root := "../"
	// Built from two literals rather than one, so this test's own source
	// text (which necessarily discusses the needle) is never itself a match
	// for the needle it is scanning for.
	needle := "func " + "corpusRelay"
	selfPath := filepath.Join("check", "costcorpus_relocation_test.go")

	var matches []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(filepath.ToSlash(path), filepath.ToSlash(selfPath)) {
			return nil
		}
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if strings.Contains(string(contents), needle) {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s for %q: %v", root, needle, err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected %q to appear in exactly one file across internal/compiler/, found %d: %v", needle, len(matches), matches)
	}
}

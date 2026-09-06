package session_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// currentLaneSchema is deliberately a test-local constant, not an import of
// protocol.Schema/protocol.LaneSchema1 -- the whole point of this pin is to
// catch drift between what the session package's composite literals
// actually say and what this test independently expects, which a shared
// constant would silently paper over.
const currentLaneSchema = "lang.verify-lane/0"

// expectedLaneSchemaLiteralSitesByFile pins the per-file count of
// currentLaneSchema string literals in internal/compiler/session as of the
// start of Phase 6 (D-06-31), confirmed by direct grep: session.go x8,
// session_phase5.go x1, session_phase5_mismatch.go x2,
// session_phase5_sanitize.go x1, total 12. Every one of these sites is a
// protocol.Lane{Schema: currentLaneSchema, ...} composite literal inside one
// of the 5 independently-implemented addLane closures (VerifyCorpus,
// verifyOwnedCorpus, verifyBorrowedCorpus, verifyForeignCorpus,
// VerifyPhase5ControlsAndWork). The coordinated 06-06 bump to
// "lang.verify-lane/1" must move every single one of these sites at once --
// this test is the mechanical completeness check waiting for it.
var expectedLaneSchemaLiteralSitesByFile = map[string]int{
	"session.go":                 8,
	"session_phase5.go":          1,
	"session_phase5_mismatch.go": 2,
	"session_phase5_sanitize.go": 1,
}

const expectedLaneSchemaLiteralSiteTotal = 12

// TestLaneSchemaLiteralSiteCountIsPinned pins the exact count and per-file
// location of every currentLaneSchema string literal in
// internal/compiler/session, so the coordinated lang.verify-lane/0 -> /1
// bump in 06-06 cannot half-land: a site moved between files, added, or
// removed without updating every other site is caught here instead of
// discovered by a runtime schema mismatch (D-06-31, T-06-03).
func TestLaneSchemaLiteralSiteCountIsPinned(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "session")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}

	fileSet := token.NewFileSet()
	countsByFile := make(map[string]int)
	total := 0

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		file, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		count := 0
		ast.Inspect(file, func(node ast.Node) bool {
			literal, ok := node.(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				return true
			}
			if value == currentLaneSchema {
				count++
			}
			return true
		})
		if count > 0 {
			countsByFile[entry.Name()] = count
		}
		total += count
	}

	if total != expectedLaneSchemaLiteralSiteTotal {
		t.Fatalf("lane schema literal site total is %d, want %d (pinned per-file: %+v; found per-file: %+v) -- a coordinated bump must move ALL sites at once",
			total, expectedLaneSchemaLiteralSiteTotal, expectedLaneSchemaLiteralSitesByFile, countsByFile)
	}

	allFiles := make(map[string]bool, len(expectedLaneSchemaLiteralSitesByFile)+len(countsByFile))
	for name := range expectedLaneSchemaLiteralSitesByFile {
		allFiles[name] = true
	}
	for name := range countsByFile {
		allFiles[name] = true
	}
	names := make([]string, 0, len(allFiles))
	for name := range allFiles {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		want := expectedLaneSchemaLiteralSitesByFile[name]
		got := countsByFile[name]
		if got != want {
			t.Fatalf("%s: found %d lane schema literal sites, want %d -- a coordinated bump must move ALL 12 sites at once", name, got, want)
		}
	}
}

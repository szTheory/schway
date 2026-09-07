package session_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// currentLaneSchema documents the /1 lane-schema string this pin's sites now
// carry, for readers -- it is not used for AST matching. 06-06's coordinated
// bump (D-06-31) moved all 12 sites from the raw "lang.verify-lane/0" string
// literal to the protocol.LaneSchema1 constant reference (the same
// two-constant-coexistence discipline diagnostic.go/evidence.go already use),
// so this pin now counts protocol.LaneSchema1 *identifier references*
// (ast.SelectorExpr) rather than ast.BasicLit string values -- a raw /1
// string literal reintroduced at any site would NOT satisfy this pin, which
// is intentional: the constant reference is now the required shape, not
// merely the current one.
const currentLaneSchema = "lang.verify-lane/1"

// expectedLaneSchemaLiteralSitesByFile pins the per-file count of
// protocol.LaneSchema1 identifier references in internal/compiler/session,
// confirmed by direct grep: session.go x8, session_phase5.go x1,
// session_phase5_mismatch.go x2, session_phase5_sanitize.go x1, total 12.
// Every one of these sites is a protocol.Lane{Schema: protocol.LaneSchema1,
// ...} composite literal inside one of the 5 independently-implemented
// addLane closures (VerifyCorpus, verifyOwnedCorpus, verifyBorrowedCorpus,
// verifyForeignCorpus, VerifyPhase5ControlsAndWork). 06-06's coordinated bump
// moved every single one of these sites at once (D-06-31) -- this test is
// the mechanical completeness check that a coordinated bump moves ALL sites,
// and a future half-landed bump (add/remove/relocate one site without the
// rest) is caught here instead of by a runtime schema mismatch.
//
// session_phase6_verify.go x3 is a DIFFERENT species of entry: not a site
// migrated by the 06-06 bump, but three brand-new lane:native-differential /
// addDeferredLane composite literals this plan (06-07) adds, which
// correctly reference protocol.LaneSchema1 directly (there was never a /0
// literal at these sites to move). [Rule 1/3 deviation: this map and the
// total below were updated from 06-06's original 12 to include these 3 new,
// legitimate sites -- 15 total -- so this pin keeps catching a genuinely
// half-landed /0-to-/1 bump among the ORIGINAL 12 while not treating this
// plan's own new production lanes as drift.]
var expectedLaneSchemaLiteralSitesByFile = map[string]int{
	"session.go":                 8,
	"session_phase5.go":          1,
	"session_phase5_mismatch.go": 2,
	"session_phase5_sanitize.go": 1,
	"session_phase6_verify.go":   3,
}

const expectedLaneSchemaLiteralSiteTotal = 15

// TestLaneSchemaLiteralSiteCountIsPinned pins the exact count and per-file
// location of every protocol.LaneSchema1 reference in
// internal/compiler/session, so the coordinated lang.verify-lane/0 -> /1
// bump in 06-06 cannot half-land: a site moved between files, added, or
// removed without updating every other site is caught here instead of
// discovered by a runtime schema mismatch (D-06-31, T-06-03). A coordinated
// bump moves ALL sites at once -- that is what this pin proves.
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
			selector, ok := node.(*ast.SelectorExpr)
			if !ok || selector.Sel == nil || selector.Sel.Name != "LaneSchema1" {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "protocol" {
				return true
			}
			count++
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

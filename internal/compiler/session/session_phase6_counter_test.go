package session_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// TestNoNoOpCompoundAssignment closes CR-01's whole class rather than its one
// instance. `VerifyPhase6ControlsAndWork` -- the function the shipped gate
// `lang verify testdata/phase6` actually runs -- carried
// `result.Metrics.CacheInputsReusedCount += 0` inside a live branch, so
// FND-04's cache-reuse counter was structurally zero on the one path that is
// gated, while the sibling VerifyPhase6ChangedRisk reported it correctly. A
// no-op compound assignment against a literal zero is never intentional: it
// is a placeholder someone meant to come back to, and it reads as working
// code. Scanning for it is what would have caught this without needing to
// run the expensive native lane.
//
// If a future counter legitimately needs a zero-valued step, express it with
// a named constant so the intent is visible and this scan stays meaningful.
func TestNoNoOpCompoundAssignment(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "session")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	fileSet := token.NewFileSet()
	scanned := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		file, parseErr := parser.ParseFile(fileSet, path, nil, parser.AllErrors)
		if parseErr != nil {
			t.Fatalf("parse %s: %v", path, parseErr)
		}
		scanned++
		ast.Inspect(file, func(node ast.Node) bool {
			assign, ok := node.(*ast.AssignStmt)
			if !ok || len(assign.Rhs) != 1 {
				return true
			}
			switch assign.Tok {
			case token.ADD_ASSIGN, token.SUB_ASSIGN, token.OR_ASSIGN, token.XOR_ASSIGN:
			default:
				return true
			}
			literal, ok := assign.Rhs[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.INT || literal.Value != "0" {
				return true
			}
			t.Errorf("%s:%d: no-op compound assignment `%s 0` -- a counter stepped by literal zero never counts; this is the CR-01 shape that made FND-04's cache-reuse count structurally zero in the gated path",
				name, fileSet.Position(assign.Pos()).Line, assign.Tok)
			return true
		})
	}
	// A scan that walked no files would pass vacuously and prove nothing.
	if scanned == 0 {
		t.Fatalf("scanned no non-test Go files in %s -- the guard cannot vacuously pass", dir)
	}
}

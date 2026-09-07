package session

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

// phase6TestRoots roots the cache and the change-state file under t.TempDir()
// via Phase6CacheRootOverrideForTest/Phase6ChangeStatePathOverrideForTest
// (VerifyPhase6ChangedRisk's own signature has no room for either, mirroring
// Phase5ClangPathOverrideForTest's precedent), and restores both on cleanup.
func phase6TestRoots(t *testing.T) (cacheRoot, statePath string) {
	t.Helper()
	dir := t.TempDir()
	cacheRoot = filepath.Join(dir, "cache")
	statePath = filepath.Join(dir, "state.json")

	originalCache := Phase6CacheRootOverrideForTest
	originalState := Phase6ChangeStatePathOverrideForTest
	Phase6CacheRootOverrideForTest = cacheRoot
	Phase6ChangeStatePathOverrideForTest = statePath
	t.Cleanup(func() {
		Phase6CacheRootOverrideForTest = originalCache
		Phase6ChangeStatePathOverrideForTest = originalState
	})
	return cacheRoot, statePath
}

func phase6FindLane(result protocol.Result, id string) *protocol.Lane {
	for i := range result.Lanes {
		if result.Lanes[i].ID == id {
			return &result.Lanes[i]
		}
	}
	return nil
}

// TestVerifyPhase6ChangedRiskRunsOneSelectedLane is Task 1's tracer: a real
// lane, selected by real changed risk, runs over a real corpus, compiles a
// real artifact through the real cache, and -- run a second time over the
// SAME cache root (but a fresh, deliberately cold change-state path, so the
// lane is selected again rather than deferred) -- reuses that artifact
// (acceptance criterion 5).
func TestVerifyPhase6ChangedRiskRunsOneSelectedLane(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	cacheRoot, _ := phase6TestRoots(t)
	runner := native.DefaultRunner()

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, runner)
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("Status = %q, want pass; diagnostics=%v lanes=%+v", result.Status, result.Diagnostics, result.Lanes)
	}

	lane := phase6FindLane(result, phase6NativeDifferentialLane)
	if lane == nil {
		t.Fatalf("no lane %s in result; lanes=%+v", phase6NativeDifferentialLane, result.Lanes)
	}
	if lane.Schema != protocol.LaneSchema1 {
		t.Errorf("Schema = %q, want %q", lane.Schema, protocol.LaneSchema1)
	}
	if lane.Status != protocol.StatusPass {
		t.Errorf("lane status = %q, want pass", lane.Status)
	}
	if lane.CacheStatus != string(cache.StatusArtifactRecomputed) {
		t.Errorf("first run CacheStatus = %q, want %q", lane.CacheStatus, cache.StatusArtifactRecomputed)
	}

	// Second run, same cache root, fresh (cold) change-state path -- the
	// selector will select the lane again (nothing recorded yet in the new
	// state file), but the artifact cache root is unchanged, so the
	// compiled binary is reused rather than recompiled.
	Phase6ChangeStatePathOverrideForTest = filepath.Join(cacheRoot, "..", "state2.json")

	result2, err := VerifyPhase6ChangedRisk(context.Background(), corpus, runner)
	if err != nil {
		t.Fatalf("second VerifyPhase6ChangedRisk error: %v", err)
	}
	lane2 := phase6FindLane(result2, phase6NativeDifferentialLane)
	if lane2 == nil {
		t.Fatalf("no lane %s in second result; lanes=%+v", phase6NativeDifferentialLane, result2.Lanes)
	}
	if lane2.Status != protocol.StatusPass {
		t.Errorf("second run lane status = %q, want pass (a cache hit changes what is skipped, never what is asserted)", lane2.Status)
	}
	if lane2.CacheStatus != string(cache.StatusArtifactReused) {
		t.Errorf("second run CacheStatus = %q, want %q", lane2.CacheStatus, cache.StatusArtifactReused)
	}
}

// TestVerifyReportsCacheStatusVocabulary asserts every lane VerifyPhase6ChangedRisk
// emits carries a schema-1 lane, a CacheStatus drawn from the closed
// four-value artifact vocabulary (or empty, not-yet-populated), and a
// SelectionReason beginning with one of the three declared prefixes.
func TestVerifyReportsCacheStatusVocabulary(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)
	runner := native.DefaultRunner()

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, runner)
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}
	if len(result.Lanes) == 0 {
		t.Fatal("no lanes in result")
	}

	prefixes := []string{ReasonSelected + ":", ReasonDeferred + ":", ReasonWidened + ":"}
	for _, lane := range result.Lanes {
		if lane.Schema != protocol.LaneSchema1 {
			t.Errorf("lane %s Schema = %q, want %q", lane.ID, lane.Schema, protocol.LaneSchema1)
		}
		if err := protocol.ValidateLaneVocabularies(lane); err != nil {
			t.Errorf("lane %s failed vocabulary validation: %v", lane.ID, err)
		}
		matched := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(lane.SelectionReason, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("lane %s SelectionReason %q does not begin with a declared reason prefix", lane.ID, lane.SelectionReason)
		}
	}
}

// TestNativeDifferentialAssertionOutsideCacheBranch is the plan's own
// required structural assertion (acceptance criterion 4): a go/ast scan of
// verifyPhase6NativeDifferentialLane proves the interpreter-vs-native
// comparator call (execution.Equal) never appears inside an *ast.IfStmt
// whose condition mentions the cache outcome -- not merely "review says it
// doesn't."
func TestNativeDifferentialAssertionOutsideCacheBranch(t *testing.T) {
	fileSet := token.NewFileSet()
	path := nat03CorpusPath("internal/compiler/session/session_phase6_verify.go")
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parser.ParseFile: %v", err)
	}

	var target *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if decl, ok := n.(*ast.FuncDecl); ok && decl.Name.Name == "verifyPhase6NativeDifferentialLane" {
			target = decl
		}
		return true
	})
	if target == nil {
		t.Fatal("verifyPhase6NativeDifferentialLane not found")
	}

	mentionsOutcome := func(cond ast.Expr) bool {
		found := false
		ast.Inspect(cond, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.Ident:
				if node.Name == "outcome" {
					found = true
				}
			case *ast.SelectorExpr:
				if ident, ok := node.X.(*ast.Ident); ok && ident.Name == "outcome" {
					found = true
				}
			}
			return true
		})
		return found
	}

	type span struct{ start, end token.Pos }
	var flagged []span
	ast.Inspect(target, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if mentionsOutcome(ifStmt.Cond) {
			flagged = append(flagged, span{ifStmt.Pos(), ifStmt.End()})
		}
		return true
	})

	insideFlagged := func(pos token.Pos) bool {
		for _, s := range flagged {
			if pos >= s.start && pos < s.end {
				return true
			}
		}
		return false
	}

	ast.Inspect(target, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "execution" || sel.Sel.Name != "Equal" {
			return true
		}
		if insideFlagged(call.Pos()) {
			t.Errorf("execution.Equal call at %v is nested inside an if-statement whose condition mentions the cache outcome", fileSet.Position(call.Pos()))
		}
		return true
	})

	if len(flagged) == 0 {
		t.Fatal("no if-statement in verifyPhase6NativeDifferentialLane mentions the cache outcome -- the structural assertion this test makes is vacuous; the function's shape changed unexpectedly")
	}
}

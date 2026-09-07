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

// TestDeferredLaneNeverRendersPass is Task 2's core structural + runtime
// property (T-06-VERIFY-02): a deferred lane's Status is protocol.StatusDeferred
// and never protocol.StatusPass, enforced structurally (phase6AddDeferredLane
// hardcodes the status and takes no status parameter) as well as at runtime.
func TestDeferredLaneNeverRendersPass(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}

	foundDeferred := false
	for _, lane := range result.Lanes {
		if lane.Status == protocol.StatusDeferred {
			foundDeferred = true
			if lane.Status == protocol.StatusPass {
				t.Fatalf("lane %s reports both deferred and pass -- impossible by construction", lane.ID)
			}
		}
	}
	if !foundDeferred {
		t.Fatal("no deferred lane observed on a cold pure_match run -- only lane:native-differential is wired through this function, so at least one other pure_match lane must render deferred")
	}

	// Structural half: phase6AddDeferredLane must never reference
	// protocol.StatusPass anywhere in its body.
	fileSet := token.NewFileSet()
	path := nat03CorpusPath("internal/compiler/session/session_phase6_verify.go")
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parser.ParseFile: %v", err)
	}
	var target *ast.FuncDecl
	ast.Inspect(file, func(n ast.Node) bool {
		if decl, ok := n.(*ast.FuncDecl); ok && decl.Name.Name == "phase6AddDeferredLane" {
			target = decl
		}
		return true
	})
	if target == nil {
		t.Fatal("phase6AddDeferredLane not found")
	}
	if len(target.Type.Params.List) == 0 {
		t.Fatal("phase6AddDeferredLane has no parameters at all, expected (result, id, reason, coldOrWarm)")
	}
	for _, field := range target.Type.Params.List {
		for _, name := range field.Names {
			if strings.EqualFold(name.Name, "status") {
				t.Fatalf("phase6AddDeferredLane declares a %q parameter -- it must take no status parameter at all, structurally incapable of rendering anything but deferred", name.Name)
			}
		}
	}
	ast.Inspect(target, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		pkg, ok := sel.X.(*ast.Ident)
		if ok && pkg.Name == "protocol" && sel.Sel.Name == "StatusPass" {
			t.Errorf("phase6AddDeferredLane references protocol.StatusPass at %v -- a lane rendered through this closure must never be able to report pass", fileSet.Position(sel.Pos()))
		}
		return true
	})
}

// TestEveryLiveLaneIsAccountedFor forces an unclassified fixture kind
// (Phase6FixtureKindOverrideForTest), which routes every one of
// LiveLaneIDs() through D-06-11's widen-to-everything branch, and asserts
// the Result's emitted lane-ID set is exactly LiveLaneIDs() -- selected and
// actually run (lane:native-differential) or deferred (every other lane) --
// with no lane silently absent (T-06-VERIFY-03).
func TestEveryLiveLaneIsAccountedFor(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)

	original := Phase6FixtureKindOverrideForTest
	Phase6FixtureKindOverrideForTest = "phase6-test-unclassified-kind"
	t.Cleanup(func() { Phase6FixtureKindOverrideForTest = original })

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}

	got := make(map[string]bool, len(result.Lanes))
	for _, lane := range result.Lanes {
		if got[lane.ID] {
			t.Errorf("lane %s appears more than once in the Result", lane.ID)
		}
		got[lane.ID] = true
		if lane.Status != protocol.StatusPass && lane.Status != protocol.StatusMismatch && lane.Status != protocol.StatusDeferred && lane.Status != protocol.StatusOperational {
			t.Errorf("lane %s has unexpected status %q", lane.ID, lane.Status)
		}
	}

	live := LiveLaneIDs()
	if len(got) != len(live) {
		t.Fatalf("Result has %d distinct lanes, want %d (LiveLaneIDs()): got=%v", len(got), len(live), got)
	}
	for _, id := range live {
		if !got[id] {
			t.Errorf("live lane %s is missing from the Result entirely -- a lane must never go silently absent", id)
		}
	}
}

// TestCacheInputsReusedCountIsCounted asserts Metrics.CacheInputsReusedCount
// equals the number of lanes whose CacheStatus is artifact_reused, in the
// same counted-unit currency as RecomputedWork -- zero on the first (cold)
// run, one after a second run reuses the cached artifact.
func TestCacheInputsReusedCountIsCounted(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	cacheRoot, _ := phase6TestRoots(t)
	runner := native.DefaultRunner()

	first, err := VerifyPhase6ChangedRisk(context.Background(), corpus, runner)
	if err != nil {
		t.Fatalf("first VerifyPhase6ChangedRisk error: %v", err)
	}
	if first.Metrics.CacheInputsReusedCount != 0 {
		t.Errorf("first run CacheInputsReusedCount = %d, want 0", first.Metrics.CacheInputsReusedCount)
	}

	Phase6ChangeStatePathOverrideForTest = filepath.Join(cacheRoot, "..", "state-second.json")
	second, err := VerifyPhase6ChangedRisk(context.Background(), corpus, runner)
	if err != nil {
		t.Fatalf("second VerifyPhase6ChangedRisk error: %v", err)
	}

	reused := 0
	for _, lane := range second.Lanes {
		if lane.CacheStatus == string(cache.StatusArtifactReused) {
			reused++
		}
	}
	if second.Metrics.CacheInputsReusedCount != reused {
		t.Errorf("second run CacheInputsReusedCount = %d, want %d (count of artifact_reused lanes)", second.Metrics.CacheInputsReusedCount, reused)
	}
	if second.Metrics.CacheInputsReusedCount != 1 {
		t.Errorf("second run CacheInputsReusedCount = %d, want exactly 1 (lane:native-differential)", second.Metrics.CacheInputsReusedCount)
	}
}

// TestFirstColdRunReportsRecomputedNotReused (FND-04's empty-input edge): a
// first run against a completely fresh cache and state root reports the one
// executed lane as recomputed (never reused) and cold, and zero reused
// artifacts overall.
func TestFirstColdRunReportsRecomputedNotReused(t *testing.T) {
	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)

	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}
	if result.Metrics.CacheInputsReusedCount != 0 {
		t.Errorf("CacheInputsReusedCount = %d, want 0 on a cold run", result.Metrics.CacheInputsReusedCount)
	}
	lane := phase6FindLane(result, phase6NativeDifferentialLane)
	if lane == nil {
		t.Fatal("no lane:native-differential in result")
	}
	if lane.CacheStatus == string(cache.StatusArtifactReused) {
		t.Errorf("CacheStatus = %q on a cold run, must never be artifact_reused", lane.CacheStatus)
	}
	if lane.ColdOrWarm != protocol.LaneCold {
		t.Errorf("ColdOrWarm = %q, want %q on the very first run", lane.ColdOrWarm, protocol.LaneCold)
	}
}

// TestLaneStatusVocabularyIsClosed asserts protocol.LaneStatuses() names the
// exact closed set, and every lane VerifyPhase6ChangedRisk emits reports a
// status drawn from it.
func TestLaneStatusVocabularyIsClosed(t *testing.T) {
	want := []string{protocol.StatusPass, protocol.StatusInvalid, protocol.StatusOperational, protocol.StatusMismatch, protocol.StatusUsage, protocol.StatusDeferred}
	got := protocol.LaneStatuses()
	if len(got) != len(want) {
		t.Fatalf("protocol.LaneStatuses() = %v, want %v", got, want)
	}
	inVocab := make(map[string]bool, len(got))
	for _, status := range got {
		inVocab[status] = true
	}
	for _, status := range want {
		if !inVocab[status] {
			t.Fatalf("protocol.LaneStatuses() missing %q: %v", status, got)
		}
	}

	corpus := nat03CorpusPath("testdata/phase1")
	phase6TestRoots(t)
	result, err := VerifyPhase6ChangedRisk(context.Background(), corpus, native.DefaultRunner())
	if err != nil {
		t.Fatalf("VerifyPhase6ChangedRisk error: %v", err)
	}
	for _, lane := range result.Lanes {
		if !inVocab[lane.Status] {
			t.Errorf("lane %s has out-of-vocabulary status %q", lane.ID, lane.Status)
		}
	}
}

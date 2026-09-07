package session

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cache"
)

// TestRiskLaneRegistryLoads proves the embedded risk_lanes.json parses to a
// non-empty row set, and that the same parse path returns a "risklanes:"
// prefixed error when the bytes are malformed -- mirroring
// LoadQLT01Registry's own error-prefix discipline.
func TestRiskLaneRegistryLoads(t *testing.T) {
	rows, err := LoadRiskLaneRegistry()
	if err != nil {
		t.Fatalf("LoadRiskLaneRegistry() returned an error on the checked-in registry: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("LoadRiskLaneRegistry() returned zero rows on the checked-in registry")
	}

	_, err = parseRiskLaneRegistry([]byte("{not valid json"))
	if err == nil {
		t.Fatal("expected an error unmarshalling malformed JSON through the same code path")
	}
	if len(err.Error()) < len("risklanes:") || err.Error()[:len("risklanes:")] != "risklanes:" {
		t.Fatalf("error %q does not begin with %q", err.Error(), "risklanes:")
	}
}

// TestRiskLaneSelectionMatchesDeclaredKind is Task 1's tracer end: a
// declared kind with one changed declared input returns only the lanes
// declaring that input, and a reason for every lane in the registry for
// that kind -- selected or deferred, never absent.
func TestRiskLaneSelectionMatchesDeclaredKind(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:b", DeclaredInputs: []string{"build_flags"}},
		{FixtureKind: "other", LaneID: "lane:c", DeclaredInputs: []string{"fixture_source"}},
	}
	live := []string{"lane:a", "lane:b", "lane:c"}

	sel := selectLanesFromRows(rows, live, "k", []string{"fixture_source"}, nil)

	if got, want := sel.LaneIDs, []string{"lane:a"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("LaneIDs = %v, want %v", got, want)
	}
	if reason, ok := sel.Reasons["lane:a"]; !ok || reason != "selected: k matched" {
		t.Fatalf("lane:a reason = %q, ok=%v, want %q", reason, ok, "selected: k matched")
	}
	if reason, ok := sel.Reasons["lane:b"]; !ok || reason != "deferred: no declared dependency" {
		t.Fatalf("lane:b reason = %q, ok=%v, want %q", reason, ok, "deferred: no declared dependency")
	}
	if _, ok := sel.Reasons["lane:c"]; ok {
		t.Fatalf("lane:c (a different fixture kind) must not appear in Reasons at all")
	}
}

// TestRiskLaneRegistryLanesAreLive proves every checked-in risk_lanes.json
// row's lane_id appears in the live lane-ID set produced by the shipped
// verify functions -- the half of Task 3's audit this task's own
// acceptance criteria requires be simply true here.
func TestRiskLaneRegistryLanesAreLive(t *testing.T) {
	rows, err := LoadRiskLaneRegistry()
	if err != nil {
		t.Fatalf("LoadRiskLaneRegistry() error: %v", err)
	}
	live := make(map[string]bool)
	for _, id := range LiveLaneIDs() {
		live[id] = true
	}
	for _, row := range rows {
		if !live[row.LaneID] {
			t.Errorf("risk_lanes.json row fixture_kind=%s lane_id=%s is not in LiveLaneIDs()", row.FixtureKind, row.LaneID)
		}
	}
}

// TestRiskLaneReasonConstantsMatchVocabulary asserts the three exported
// reason constants match D-06-12's exact vocabulary prefixes.
func TestRiskLaneReasonConstantsMatchVocabulary(t *testing.T) {
	if ReasonSelected != "selected" {
		t.Errorf("ReasonSelected = %q, want %q", ReasonSelected, "selected")
	}
	if ReasonDeferred != "deferred" {
		t.Errorf("ReasonDeferred = %q, want %q", ReasonDeferred, "deferred")
	}
	if ReasonWidened != "widened" {
		t.Errorf("ReasonWidened = %q, want %q", ReasonWidened, "widened")
	}
}

// TestRiskLaneUnclassifiedFixtureWidensToAllLanes: D-06-11's fail-closed
// branch. A fixture kind the registry does not exhaustively classify
// selects every live lane, each with reason ReasonWidened, and reports
// cache.StatusNotCacheable for the run -- ambiguity never resolves to
// "skip it."
func TestRiskLaneUnclassifiedFixtureWidensToAllLanes(t *testing.T) {
	sel, err := SelectLanes("a-kind-not-in-the-registry", nil, nil)
	if err != nil {
		t.Fatalf("SelectLanes error: %v", err)
	}
	live := LiveLaneIDs()
	if len(sel.LaneIDs) != len(live) {
		t.Fatalf("LaneIDs has %d entries, want %d (the full live lane set): %v", len(sel.LaneIDs), len(live), sel.LaneIDs)
	}
	for _, id := range live {
		if reason := sel.Reasons[id]; reason != ReasonWidened+": undeclared-input risk" {
			t.Errorf("lane %s reason = %q, want widened", id, reason)
		}
	}
	if sel.CacheStatus != cache.StatusNotCacheable {
		t.Errorf("CacheStatus = %q, want %q", sel.CacheStatus, cache.StatusNotCacheable)
	}
}

// TestRiskLaneUndeclaredInputIsNotCacheable: a fixture whose declared-input
// set cannot be fully computed (e.g. the Clang identity probe failed)
// selects every lane for that fixture kind and is not_cacheable for the
// run, even though the kind itself IS classified in the registry.
func TestRiskLaneUndeclaredInputIsNotCacheable(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:b", DeclaredInputs: []string{"build_flags"}},
	}
	live := []string{"lane:a", "lane:b"}
	probeErr := errors.New("clang identity probe failed")

	sel := selectLanesFromRows(rows, live, "k", nil, probeErr)

	want := []string{"lane:a", "lane:b"}
	if len(sel.LaneIDs) != len(want) {
		t.Fatalf("LaneIDs = %v, want %v", sel.LaneIDs, want)
	}
	for _, id := range want {
		if reason := sel.Reasons[id]; reason != ReasonWidened+": undeclared-input risk" {
			t.Errorf("lane %s reason = %q, want widened", id, reason)
		}
	}
	if sel.CacheStatus != cache.StatusNotCacheable {
		t.Errorf("CacheStatus = %q, want %q", sel.CacheStatus, cache.StatusNotCacheable)
	}
}

// TestChangeStateFirstRunIsColdAndSelectsAll: with no state file present at
// all, every lane for a known fixture kind is selected and the run is
// reported cold -- the specified behaviour on any machine's very first
// invocation, not a degraded mode.
func TestChangeStateFirstRunIsColdAndSelectsAll(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "does-not-exist", "risklane-state.json")

	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:b", DeclaredInputs: []string{"build_flags"}},
	}
	live := []string{"lane:a", "lane:b"}
	current := FixtureInputs{"fixture_source": "digest-1", "build_flags": "digest-2"}

	sel, _, cold, err := selectLanesForFixtureFromRows(rows, live, "k", "fixture-1", current, nil, statePath)
	if err != nil {
		t.Fatalf("selectLanesForFixtureFromRows error: %v", err)
	}
	if !cold {
		t.Error("cold = false on a run with no prior state file, want true")
	}
	want := []string{"lane:a", "lane:b"}
	if len(sel.LaneIDs) != len(want) {
		t.Fatalf("LaneIDs = %v, want the full lane set %v", sel.LaneIDs, want)
	}
}

// TestChangeStateCorruptFileWidens: a corrupt/truncated state file widens
// to the full lane set rather than trusting whatever prefix parsed.
func TestChangeStateCorruptFileWidens(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "risklane-state.json")
	if err := os.WriteFile(statePath, []byte(`{"entries": {"fixture-1`), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:b", DeclaredInputs: []string{"build_flags"}},
	}
	live := []string{"lane:a", "lane:b"}
	current := FixtureInputs{"fixture_source": "digest-1", "build_flags": "digest-2"}

	sel, _, cold, err := selectLanesForFixtureFromRows(rows, live, "k", "fixture-1", current, nil, statePath)
	if err != nil {
		t.Fatalf("selectLanesForFixtureFromRows error: %v", err)
	}
	if !cold {
		t.Error("cold = false after a corrupt state file, want true (no partial trust)")
	}
	if len(sel.LaneIDs) != 2 {
		t.Fatalf("LaneIDs = %v, want both lanes selected", sel.LaneIDs)
	}
}

// TestChangeStateFileIsLocalAndGitignored: the default state path is
// derived from os.UserCacheDir(), never from the repository root, and
// .gitignore covers it by name with an adjacent comment.
func TestChangeStateFileIsLocalAndGitignored(t *testing.T) {
	path, err := DefaultChangeStatePath()
	if err != nil {
		t.Fatalf("DefaultChangeStatePath error: %v", err)
	}
	base, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("os.UserCacheDir error: %v", err)
	}
	if !stringsHasPrefix(path, base) {
		t.Errorf("DefaultChangeStatePath() = %q, want a prefix of os.UserCacheDir() (%q)", path, base)
	}

	repoRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd error: %v", err)
	}
	if stringsHasPrefix(path, repoRoot) {
		t.Errorf("DefaultChangeStatePath() = %q must not live under the repository root %q", path, repoRoot)
	}

	gitignore, err := os.ReadFile(nat03CorpusPath(".gitignore"))
	if err != nil {
		t.Fatalf("reading .gitignore: %v", err)
	}
	if !stringsContains(string(gitignore), changeStateFileName) {
		t.Errorf(".gitignore does not contain an entry for %q", changeStateFileName)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func stringsContains(haystack, needle string) bool {
	return len(needle) == 0 || indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// TestRiskLaneChangeStateRecordsOnePerFixtureLane (behaviour Test 4): after
// a run writes the state file, an unchanged fixture defers the lanes whose
// declared inputs did not move, and the state records one entry per
// (fixture, lane) pair.
func TestRiskLaneChangeStateRecordsOnePerFixtureLane(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "risklane-state.json")

	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:b", DeclaredInputs: []string{"build_flags"}},
	}
	live := []string{"lane:a", "lane:b"}
	current := FixtureInputs{"fixture_source": "digest-1", "build_flags": "digest-2"}

	_, next, _, err := selectLanesForFixtureFromRows(rows, live, "k", "fixture-1", current, nil, statePath)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}
	if len(next.Entries) != 2 {
		t.Fatalf("state has %d entries after first run, want 2 (one per fixture,lane pair): %v", len(next.Entries), next.Entries)
	}
	if err := next.Save(statePath); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	sel, _, cold, err := selectLanesForFixtureFromRows(rows, live, "k", "fixture-1", current, nil, statePath)
	if err != nil {
		t.Fatalf("second run error: %v", err)
	}
	if cold {
		t.Error("cold = true on a second run with an unchanged, saved state file")
	}
	if len(sel.LaneIDs) != 0 {
		t.Fatalf("LaneIDs = %v, want none selected (nothing changed)", sel.LaneIDs)
	}
	for _, id := range []string{"lane:a", "lane:b"} {
		if reason := sel.Reasons[id]; reason != ReasonDeferred+": no declared dependency" {
			t.Errorf("lane %s reason = %q, want deferred", id, reason)
		}
	}
}


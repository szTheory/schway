package session

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cache"
	"github.com/szTheory/schway/internal/compiler/testsupport"
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

// TestRiskLaneAuditRefusesStaleLaneID: a row whose lane_id is not in the
// live lane-ID set fails the audit with the stale-reference control.
func TestRiskLaneAuditRefusesStaleLaneID(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:does-not-exist", DeclaredInputs: []string{"fixture_source"}},
	}
	live := []string{"lane:a"}
	declared := []string{"fixture_source"}

	result := AuditRiskLaneRegistry(rows, live, declared)
	if result.Status == "pass" {
		t.Fatal("expected a fail status for a fabricated lane_id")
	}
	if !result.Fired[ControlRiskLaneStaleLaneReference] {
		t.Errorf("Fired[%s] = false, want true", ControlRiskLaneStaleLaneReference)
	}
}

// TestRiskLaneAuditRefusesEmptyRegistry: an empty registry is a hard audit
// failure, not a vacuous pass.
func TestRiskLaneAuditRefusesEmptyRegistry(t *testing.T) {
	result := AuditRiskLaneRegistry(nil, []string{"lane:a"}, []string{"fixture_source"})
	if result.Status == "pass" {
		t.Fatal("an empty registry must not pass the audit")
	}
}

// TestRiskLaneAuditRefusesUndeclaredLane: the audit is bidirectional -- a
// live lane ID with no row in the registry also fails the audit, so a
// newly added lane cannot silently escape declaration.
func TestRiskLaneAuditRefusesUndeclaredLane(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"fixture_source"}},
	}
	live := []string{"lane:a", "lane:orphaned"}
	declared := []string{"fixture_source"}

	result := AuditRiskLaneRegistry(rows, live, declared)
	if result.Status == "pass" {
		t.Fatal("expected a fail status when a live lane has no registry row")
	}
	if !result.Fired[ControlRiskLaneUndeclaredLane] {
		t.Errorf("Fired[%s] = false, want true", ControlRiskLaneUndeclaredLane)
	}
}

// TestRiskLaneAuditRefusesUnknownDeclaredInput: a row naming a declared
// input outside cache.DeclaredInputNames() fails the audit.
func TestRiskLaneAuditRefusesUnknownDeclaredInput(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:a", DeclaredInputs: []string{"not_a_real_declared_input"}},
	}
	live := []string{"lane:a"}
	declared := cache.DeclaredInputNames()

	result := AuditRiskLaneRegistry(rows, live, declared)
	if result.Status == "pass" {
		t.Fatal("expected a fail status for a row naming an undeclared input")
	}
}

// TestRiskLaneAuditPassesCheckedInRegistry proves AuditRiskLaneRegistry on
// the real, checked-in registry against LiveLaneIDs() and
// cache.DeclaredInputNames() returns pass with zero fired controls -- the
// declared table cannot drift from the code that emits the lanes, in
// either direction.
func TestRiskLaneAuditPassesCheckedInRegistry(t *testing.T) {
	rows, err := LoadRiskLaneRegistry()
	if err != nil {
		t.Fatalf("LoadRiskLaneRegistry error: %v", err)
	}
	result := AuditRiskLaneRegistry(rows, LiveLaneIDs(), cache.DeclaredInputNames())
	if result.Status != "pass" {
		t.Fatalf("Status = %q, want pass; Fired = %v", result.Status, result.Fired)
	}
	for control, fired := range result.Fired {
		if fired {
			t.Errorf("control %s fired against the checked-in registry", control)
		}
	}
}

// TestRiskLaneVerifyRegistryRunsCleanly is a thin smoke test over
// VerifyRiskLaneRegistry, mirroring VerifyQLT01Registry's own shape.
func TestRiskLaneVerifyRegistryRunsCleanly(t *testing.T) {
	result, err := VerifyRiskLaneRegistry(context.Background())
	if err != nil {
		t.Fatalf("VerifyRiskLaneRegistry error: %v", err)
	}
	if result.Status != "pass" {
		t.Fatalf("Status = %q, want pass", result.Status)
	}
	if result.RecomputedWork <= 0 {
		t.Errorf("RecomputedWork = %d, want > 0", result.RecomputedWork)
	}
}

// TestRiskLaneDoubleMatchSelectsOnce: a lane matched by two different rows
// (for the same fixture kind) is selected exactly once, with the FIRST
// matching row's reason by registry order.
func TestRiskLaneDoubleMatchSelectsOnce(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:dup", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:dup", DeclaredInputs: []string{"build_flags"}},
	}
	live := []string{"lane:dup"}

	sel := selectLanesFromRows(rows, live, "k", []string{"fixture_source", "build_flags"}, nil)

	count := 0
	for _, id := range sel.LaneIDs {
		if id == "lane:dup" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("lane:dup appears %d times in LaneIDs, want exactly once: %v", count, sel.LaneIDs)
	}
	if len(sel.Reasons) != 1 {
		t.Fatalf("Reasons has %d entries, want exactly 1 (one reason for the deduplicated lane): %v", len(sel.Reasons), sel.Reasons)
	}
}

// TestRiskLaneSelectionOrderIsStable: two lanes comparing equal on the
// selection key appear in a specified, stable order (lexicographic)
// across two invocations.
func TestRiskLaneSelectionOrderIsStable(t *testing.T) {
	rows := []RiskLaneRow{
		{FixtureKind: "k", LaneID: "lane:zeta", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:alpha", DeclaredInputs: []string{"fixture_source"}},
		{FixtureKind: "k", LaneID: "lane:mid", DeclaredInputs: []string{"fixture_source"}},
	}
	live := []string{"lane:zeta", "lane:alpha", "lane:mid"}
	changed := []string{"fixture_source"}

	first := selectLanesFromRows(rows, live, "k", changed, nil)
	second := selectLanesFromRows(rows, live, "k", changed, nil)

	want := []string{"lane:alpha", "lane:mid", "lane:zeta"}
	for i, id := range want {
		if first.LaneIDs[i] != id {
			t.Fatalf("first.LaneIDs = %v, want %v", first.LaneIDs, want)
		}
		if second.LaneIDs[i] != id {
			t.Fatalf("second.LaneIDs = %v, want %v", second.LaneIDs, want)
		}
	}
}

// TestRiskLaneSelectLanesOutputHasNoDuplicates is a direct check of
// SelectLanes's own dedup/sort contract over the real checked-in registry.
func TestRiskLaneSelectLanesOutputHasNoDuplicates(t *testing.T) {
	sel, err := SelectLanes("pure_match", []string{"fixture_source"}, nil)
	if err != nil {
		t.Fatalf("SelectLanes error: %v", err)
	}
	seen := map[string]bool{}
	for i, id := range sel.LaneIDs {
		if seen[id] {
			t.Fatalf("duplicate lane ID %s in LaneIDs: %v", id, sel.LaneIDs)
		}
		seen[id] = true
		if i > 0 && sel.LaneIDs[i-1] > id {
			t.Fatalf("LaneIDs not lexicographically sorted: %v", sel.LaneIDs)
		}
	}
}

// TestRiskLaneFileDoesNotImportProtocol is D-06-28's structural import
// boundary applied here: session_phase6_risklanes.go must never import
// internal/compiler/protocol, keeping this file's wiring independent of
// how a future caller (06-07) attaches Selection to protocol.Lane --
// mirroring corevalidate's own TestValidatorImportsStayIndependent.
func TestRiskLaneFileDoesNotImportProtocol(t *testing.T) {
	path := testsupport.ProjectPath("internal", "compiler", "session", "session_phase6_risklanes.go")
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	for _, imported := range file.Imports {
		path := strings.Trim(imported.Path.Value, `"`)
		if strings.HasSuffix(path, "/compiler/protocol") {
			t.Fatalf("session_phase6_risklanes.go imports %s, which it must never depend on", path)
		}
	}
}

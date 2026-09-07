package session

import (
	"testing"
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


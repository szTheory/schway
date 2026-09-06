package session_test

import (
	"context"
	"os"
	"regexp"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// spikeDirNamePattern matches a spike directory's numeric prefix, e.g.
// "001-ownership-kernel-workbench".
var spikeDirNamePattern = regexp.MustCompile(`^([0-9]{3})-`)

// testGoIdentifierPattern is the acceptance criterion's rejection shape:
// a control_mechanism must not read like a Go test identifier.
var testGoIdentifierPattern = regexp.MustCompile(`^Test[A-Z]`)

func discoverSpikeIDs(t *testing.T) []string {
	t.Helper()
	entries, err := os.ReadDir(testsupport.ProjectPath(".planning", "spikes"))
	if err != nil {
		t.Fatalf("failed to list .planning/spikes: %v", err)
	}
	var ids []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		match := spikeDirNamePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			continue
		}
		ids = append(ids, match[1])
	}
	return ids
}

// TestQLT01RegistryParses is task 1's first verification target: the
// embedded qlt01_registry.json must be valid JSON, non-empty, and every row
// must carry a control_mechanism that is not a Go test identifier.
func TestQLT01RegistryParses(t *testing.T) {
	rows, err := session.LoadQLT01Registry()
	if err != nil {
		t.Fatalf("LoadQLT01Registry() failed: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("qlt01_registry.json parsed to zero rows")
	}
	for _, row := range rows {
		if row.ControlMechanism == "" {
			t.Fatalf("row spike_id=%s control_id=%s has an empty control_mechanism", row.SpikeID, row.ControlID)
		}
		if testGoIdentifierPattern.MatchString(row.ControlMechanism) {
			t.Fatalf("row spike_id=%s control_id=%s's control_mechanism %q looks like a Go test identifier, not a stated hazard", row.SpikeID, row.ControlID, row.ControlMechanism)
		}
		if row.LiveDescendant != nil && row.Waived != nil {
			t.Fatalf("row spike_id=%s control_id=%s carries BOTH live_descendant and waived", row.SpikeID, row.ControlID)
		}
		if row.LiveDescendant == nil && row.Waived == nil {
			t.Fatalf("row spike_id=%s control_id=%s carries NEITHER live_descendant nor waived", row.SpikeID, row.ControlID)
		}
		if row.Waived != nil {
			citation := row.Waived.Citation
			if citation == "" {
				t.Fatalf("row spike_id=%s control_id=%s has an empty waived.citation", row.SpikeID, row.ControlID)
			}
			switch citation {
			case "not relevant", "n/a", "N/A", "none", "None", "superseded", "Superseded":
				t.Fatalf("row spike_id=%s control_id=%s's waived.citation %q is a forbidden free-text placeholder", row.SpikeID, row.ControlID, citation)
			}
		}
	}
}

// TestQLT01RegistryCoversAllFiveSpikes asserts, by set comparison against
// the spike directory listing, that rows exist for all five spike IDs --
// an added spike directory with no rows fails this test (T-05-40).
func TestQLT01RegistryCoversAllFiveSpikes(t *testing.T) {
	spikeIDs := discoverSpikeIDs(t)
	if len(spikeIDs) == 0 {
		t.Fatal("discovered zero spike directories under .planning/spikes")
	}

	rows, err := session.LoadQLT01Registry()
	if err != nil {
		t.Fatalf("LoadQLT01Registry() failed: %v", err)
	}
	covered := make(map[string]bool)
	for _, row := range rows {
		covered[row.SpikeID] = true
	}

	for _, id := range spikeIDs {
		if !covered[id] {
			t.Fatalf("spike %s has a directory under .planning/spikes but no registry row cites it", id)
		}
	}

	var iteration5Found bool
	for _, row := range rows {
		if row.SpikeID == "005" && row.LiveDescendant != nil && row.LiveDescendant.ControlID == "control:foreign.nonlocal_exit_undetected" {
			iteration5Found = true
			if !regexp.MustCompile(`(?i)no sanitizer report|sanitizer.*cannot see|zero.*sanitizer`).MatchString(row.ControlMechanism) {
				t.Fatalf("spike 005's nonlocal-exit row must name the no-sanitizer-report property, got %q", row.ControlMechanism)
			}
		}
	}
	if !iteration5Found {
		t.Fatal("spike 005's iteration-5 (longjmp bypasses cleanup with no sanitizer report) row is missing")
	}
}

// TestQLT01RegistryComplete implements D-05-29's two halves plus the (c)/(d)
// extensions: neither/both disposition, stale control reference, missing
// fixture, and empty registry are all hard failures with distinct messages.
func TestQLT01RegistryComplete(t *testing.T) {
	rows, err := session.LoadQLT01Registry()
	if err != nil {
		t.Fatalf("LoadQLT01Registry() failed: %v", err)
	}
	shipped := session.AllShippedControlIDs()
	if len(shipped) < 2 {
		t.Fatalf("AllShippedControlIDs() returned suspiciously few controls: %v", shipped)
	}
	failures := session.AuditQLT01Registry(rows, shipped)
	if len(failures) != 0 {
		for _, failure := range failures {
			t.Errorf("qlt01 audit failure: %v", failure)
		}
		t.Fatal("TestQLT01RegistryComplete found audit failures in the committed registry")
	}
}

// TestQLT01AuditGoesRedOnStaleControl proves the audit can go red without
// editing the committed registry (D-05-38): an in-memory row citing a
// nonexistent control must be reported.
func TestQLT01AuditGoesRedOnStaleControl(t *testing.T) {
	rows := []session.QLT01Row{
		{
			SpikeID:          "999",
			ControlID:        "synthetic-stale-row",
			ControlMechanism: "a synthetic hazard used only to prove the audit can go red",
			LiveDescendant: &session.QLT01LiveDescendant{
				Fixture:   "testdata/phase5/false_restrict_hoist.lang",
				ControlID: "control:does.not.exist",
			},
		},
	}
	failures := session.AuditQLT01Registry(rows, session.AllShippedControlIDs())
	if len(failures) == 0 {
		t.Fatal("expected AuditQLT01Registry to report a stale control reference, got no failures")
	}
	var found bool
	for _, failure := range failures {
		if failure.Control == session.ControlQLT01StaleControlReference {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a %s failure, got %v", session.ControlQLT01StaleControlReference, failures)
	}
}

// TestQLT01AuditGoesRedOnDualDisposition proves the audit also rejects a
// row carrying both a live descendant and a waiver.
func TestQLT01AuditGoesRedOnDualDisposition(t *testing.T) {
	rows := []session.QLT01Row{
		{
			SpikeID:          "999",
			ControlID:        "synthetic-dual-row",
			ControlMechanism: "a synthetic hazard used only to prove the audit rejects dual dispositions",
			LiveDescendant: &session.QLT01LiveDescendant{
				Fixture:   "testdata/phase5/false_restrict_hoist.lang",
				ControlID: "control:alias.false_no_alias",
			},
			Waived: &session.QLT01Waiver{Reason: "r", Citation: "a specific citation", Owner: "o", Phase: "05"},
		},
	}
	failures := session.AuditQLT01Registry(rows, session.AllShippedControlIDs())
	if len(failures) == 0 {
		t.Fatal("expected AuditQLT01Registry to reject a dual-disposition row, got no failures")
	}
}

// TestQLT01AuditGoesRedOnEmptyRegistry proves an empty registry fails the
// audit rather than passing costlessly.
func TestQLT01AuditGoesRedOnEmptyRegistry(t *testing.T) {
	failures := session.AuditQLT01Registry(nil, session.AllShippedControlIDs())
	if len(failures) == 0 {
		t.Fatal("expected an empty registry to fail the audit, got no failures")
	}
}

// TestQLT01LaneCountsWork asserts VerifyQLT01Registry's RecomputedWork
// equals the expected row-derived value and is strictly greater than zero.
func TestQLT01LaneCountsWork(t *testing.T) {
	rows, err := session.LoadQLT01Registry()
	if err != nil {
		t.Fatalf("LoadQLT01Registry() failed: %v", err)
	}
	expected := 0
	for _, row := range rows {
		expected++
		if row.LiveDescendant != nil {
			expected++
		}
	}

	result, err := session.VerifyQLT01Registry(context.Background())
	if err != nil {
		t.Fatalf("VerifyQLT01Registry() failed: %v", err)
	}
	if result.RecomputedWork <= 0 {
		t.Fatalf("expected strictly positive RecomputedWork, got %d", result.RecomputedWork)
	}
	if result.RecomputedWork != expected {
		t.Fatalf("expected RecomputedWork == %d (row-derived), got %d", expected, result.RecomputedWork)
	}
	if result.Status != "pass" {
		t.Fatalf("expected the committed registry to pass its own lane, got status %q", result.Status)
	}
}

// TestQLT01LaneEmptyRegistryReportsZeroWorkAndFails asserts an empty
// registry produces zero work AND a lane failure -- never zero work and a
// pass, which would be the inert-lane shape.
func TestQLT01LaneEmptyRegistryReportsZeroWorkAndFails(t *testing.T) {
	result := session.QLT01LaneFromRows(nil, session.AllShippedControlIDs())
	if result.RecomputedWork != 0 {
		t.Fatalf("expected zero RecomputedWork for an empty registry, got %d", result.RecomputedWork)
	}
	if result.Status == "pass" {
		t.Fatal("expected an empty registry to fail the lane, got status \"pass\" (zero work with a pass is the inert-lane shape)")
	}
}

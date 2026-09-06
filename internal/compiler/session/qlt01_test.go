package session_test

import (
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

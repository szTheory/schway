package session

import (
	"encoding/json"
	"fmt"
	"os"

	_ "embed"
)

//go:embed qlt01_registry.json
var qlt01RegistryBytes []byte

// QLT01LiveDescendant names the fixture and shipped control: identifier that
// keeps a spike's control mechanism alive in production, per D-05-28.
type QLT01LiveDescendant struct {
	Fixture   string `json:"fixture"`
	ControlID string `json:"control_id"`
}

// QLT01Waiver records why a spike's control mechanism has no live descendant.
// D-05-28: "not relevant" is valid ONLY as a waiver with a specific
// falsifiable citation -- never as free text.
type QLT01Waiver struct {
	Reason   string `json:"reason"`
	Citation string `json:"citation"`
	Owner    string `json:"owner"`
	Phase    string `json:"phase"`
}

// QLT01Row is one registry entry: a spike's control mechanism plus exactly
// one of a live descendant or a waiver.
type QLT01Row struct {
	SpikeID          string               `json:"spike_id"`
	ControlID        string               `json:"control_id"`
	ControlMechanism string               `json:"control_mechanism"`
	LiveDescendant   *QLT01LiveDescendant `json:"live_descendant"`
	Waived           *QLT01Waiver         `json:"waived"`
}

// LoadQLT01Registry parses the embedded QLT-01 control registry.
func LoadQLT01Registry() ([]QLT01Row, error) {
	var rows []QLT01Row
	if err := json.Unmarshal(qlt01RegistryBytes, &rows); err != nil {
		return nil, fmt.Errorf("qlt01: failed to parse embedded registry: %w", err)
	}
	return rows, nil
}

// AllShippedControlIDs is the union of every phase's required-control set
// currently exported by this package, built from the existing per-phase
// functions rather than a hand-copied literal, so it cannot go stale
// independently of Phase4RequiredControls/Phase5RequiredControls.
func AllShippedControlIDs() []string {
	seen := make(map[string]bool)
	var all []string
	for _, control := range Phase4RequiredControls() {
		if !seen[control] {
			seen[control] = true
			all = append(all, control)
		}
	}
	for _, control := range Phase5RequiredControls() {
		if !seen[control] {
			seen[control] = true
			all = append(all, control)
		}
	}
	return all
}

// Control identifiers for the QLT-01 registry audit itself (D-05-29).
const (
	// ControlQLT01RegistryIncomplete fires when a registry row has neither
	// a live descendant nor a waiver, or has both -- exactly one is
	// required (D-05-28/D-05-29 half (a)).
	ControlQLT01RegistryIncomplete = "control:qlt01.registry_incomplete"

	// ControlQLT01StaleControlReference fires when a row's
	// live_descendant.control_id no longer exists in AllShippedControlIDs()
	// -- the "ported fixture that rotted" species of theater (D-05-29 half (b)).
	ControlQLT01StaleControlReference = "control:qlt01.stale_control_reference"
)

// qlt01FreeTextPlaceholders are the free-text waiver citations D-05-28
// explicitly forbids -- "not relevant" is valid only as a waiver with a
// specific falsifiable citation, never as one of these placeholders alone.
var qlt01FreeTextPlaceholders = map[string]bool{
	"not relevant": true,
	"n/a":          true,
	"none":         true,
	"superseded":   true,
}

// QLT01AuditFailure is one completeness-audit violation, naming the
// offending row and which of the four conditions failed.
type QLT01AuditFailure struct {
	SpikeID   string
	ControlID string
	Reason    string
	Control   string
}

func (f QLT01AuditFailure) Error() string {
	return fmt.Sprintf("qlt01 registry row spike_id=%s control_id=%s failed %s: %s", f.SpikeID, f.ControlID, f.Control, f.Reason)
}

// AuditQLT01Registry implements D-05-29's completeness audit over rows:
// (a) neither/both live_descendant and waived is a hard failure;
// (b) a live_descendant.control_id absent from shippedControlIDs is a hard
//
//	failure -- the stale-control-reference check that proves the audit
//	can go red without editing the committed registry;
//
// (c) a live_descendant.fixture path that does not exist on disk is a hard
//
//	failure;
//
// (d) an empty registry is a hard failure -- an empty registry passing a
//
//	completeness audit would be the false-green shape this whole phase
//	is about.
func AuditQLT01Registry(rows []QLT01Row, shippedControlIDs []string) []QLT01AuditFailure {
	var failures []QLT01AuditFailure

	if len(rows) == 0 {
		failures = append(failures, QLT01AuditFailure{
			Reason: "registry is empty", Control: ControlQLT01RegistryIncomplete,
		})
		return failures
	}

	shipped := make(map[string]bool, len(shippedControlIDs))
	for _, id := range shippedControlIDs {
		shipped[id] = true
	}

	for _, row := range rows {
		hasLive := row.LiveDescendant != nil
		hasWaiver := row.Waived != nil

		switch {
		case hasLive && hasWaiver:
			failures = append(failures, QLT01AuditFailure{
				SpikeID: row.SpikeID, ControlID: row.ControlID,
				Reason:  "row has BOTH a live_descendant and a waived disposition; exactly one is required",
				Control: ControlQLT01RegistryIncomplete,
			})
		case !hasLive && !hasWaiver:
			failures = append(failures, QLT01AuditFailure{
				SpikeID: row.SpikeID, ControlID: row.ControlID,
				Reason:  "row has NEITHER a live_descendant nor a waived disposition; exactly one is required",
				Control: ControlQLT01RegistryIncomplete,
			})
		case hasLive:
			if !shipped[row.LiveDescendant.ControlID] {
				failures = append(failures, QLT01AuditFailure{
					SpikeID: row.SpikeID, ControlID: row.ControlID,
					Reason:  fmt.Sprintf("cites control_id %q, which is not in AllShippedControlIDs()", row.LiveDescendant.ControlID),
					Control: ControlQLT01StaleControlReference,
				})
			}
			if row.LiveDescendant.Fixture == "" {
				failures = append(failures, QLT01AuditFailure{
					SpikeID: row.SpikeID, ControlID: row.ControlID,
					Reason: "live_descendant.fixture is empty", Control: ControlQLT01RegistryIncomplete,
				})
			} else if _, err := os.Stat(qlt01RepoPath(row.LiveDescendant.Fixture)); err != nil {
				failures = append(failures, QLT01AuditFailure{
					SpikeID: row.SpikeID, ControlID: row.ControlID,
					Reason:  fmt.Sprintf("live_descendant.fixture %q does not exist on disk: %v", row.LiveDescendant.Fixture, err),
					Control: ControlQLT01RegistryIncomplete,
				})
			}
		case hasWaiver:
			citation := row.Waived.Citation
			if citation == "" || qlt01FreeTextPlaceholders[normalizeQLT01Citation(citation)] {
				failures = append(failures, QLT01AuditFailure{
					SpikeID: row.SpikeID, ControlID: row.ControlID,
					Reason:  fmt.Sprintf("waived.citation %q is empty or a forbidden free-text placeholder", citation),
					Control: ControlQLT01RegistryIncomplete,
				})
			}
		}
	}

	return failures
}

func normalizeQLT01Citation(citation string) string {
	result := make([]rune, 0, len(citation))
	for _, r := range citation {
		if r >= 'A' && r <= 'Z' {
			r += 'a' - 'A'
		}
		result = append(result, r)
	}
	trimmed := string(result)
	for len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t') {
		trimmed = trimmed[1:]
	}
	for len(trimmed) > 0 && (trimmed[len(trimmed)-1] == ' ' || trimmed[len(trimmed)-1] == '\t') {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return trimmed
}

// qlt01RepoPath resolves a repo-relative path against the project root, the
// same way NAT03Mutation.CorpusProgram paths are resolved elsewhere in this
// package, so the audit works regardless of the test binary's working
// directory.
func qlt01RepoPath(relative string) string {
	return nat03CorpusPath(relative)
}

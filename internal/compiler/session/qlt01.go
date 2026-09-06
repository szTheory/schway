package session

import (
	"encoding/json"
	"fmt"

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

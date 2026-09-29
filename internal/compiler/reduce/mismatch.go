// mismatch.go defines lang.mismatch/0 (D-05-26): the document an AI repair
// agent consumes to propose a fix for one interpreter-mismatch observation
// without opening the full execution log. It is a NEW top-level schema at
// /0 per this project's convention -- lang.diagnostic/1 and
// lang.execution/0 identities do not move (D-05-39). See mismatch_test.go
// for the golden round-trip and field-set enforcement tests, and
// mutationkill_test.go for the three reducer vacuity-mode kills D-05-27
// requires.
package reduce

import (
	"encoding/json"
	"fmt"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

// MismatchSchema is lang.mismatch/0's schema identity. Publishing a new
// top-level schema identity is a one-way, published-contract commitment
// (D-05-26): once emitted, consumers -- principally the AI repair agent
// D-05-26 names as this document's primary consumer -- bind to its exact
// field set, and changing the set later means a /1 bump, never an edit.
// No other schema identity moves in this plan (D-05-39): no lang.core/2,
// no lang.evidence/2, no lang.diagnostic/2.
const MismatchSchema = "lang.mismatch/0"

// EventWindowSize is the bounded number of trailing events retained per
// engine side in MismatchDocument.EventWindow (D-05-26/T-05-47): bounded
// so the document stays self-sufficient for a repair agent without
// becoming the full execution trace.
const EventWindowSize = 8

// EventRecord is one bounded, addressable entry in EventWindow -- a
// narrower projection of an engine's own event record (kind, operation
// identity, the function it ran in, and human-legible detail) rather than
// re-exporting an execution engine's own event schema identity as this
// schema's own. Session's own wiring (plan 05-12's
// session_phase5_mismatch.go) is responsible for projecting a real
// execution.Event into an EventRecord, keeping this package free of any
// execution/session dependency (mirroring reduce.Signature's own
// session-decoupling discipline in predicate.go).
type EventRecord struct {
	Kind        string `json:"kind"`
	OperationID string `json:"operation_id,omitempty"`
	FunctionID  string `json:"function_id,omitempty"`
	Detail      string `json:"detail,omitempty"`
}

// CausalStep is one ordered acquisition/borrow/control-edge step on the
// path leading to the diverging operation (D-05-26).
type CausalStep struct {
	Kind        string `json:"kind"`
	OperationID string `json:"operation_id"`
	Place       string `json:"place,omitempty"`
}

// MismatchDocument is lang.mismatch/0.
//
// Field set (Claude's Discretion -- see 05-12-SUMMARY.md's recorded
// discretionary deviation, licensed by 05-CONTEXT.md's "Claude's
// Discretion" block naming "the lang.mismatch/0 field encoding details"):
// D-05-26's twelve fields (schema, diverging_axis, engine_pair,
// diverging_operation_id, reduced_core, reduced_source, minimality,
// total_recomputed_work, reduction_attempts, event_window, causal_chain,
// causes) PLUS an additive evidence_id binding this document back to the
// evidence manifest it was produced from. Schema identity at /0 is
// one-way and D-05-39 pins every other identity against moving, so the
// cost of adding evidence_id now (free, additive) is far lower than
// adding it after publication (a /1 bump this discretion call avoids).
// Causes REUSES lang.diagnostic/1's own []diagnostic.Cause shape
// verbatim, so a repair agent parses one cause format project-wide.
type MismatchDocument struct {
	Schema               string                   `json:"schema"`
	DivergingAxis        string                   `json:"diverging_axis"`
	EnginePair           string                   `json:"engine_pair"`
	DivergingOperationID string                   `json:"diverging_operation_id"`
	ReducedCore          string                   `json:"reduced_core"`
	ReducedSource        string                   `json:"reduced_source"`
	Minimality           string                   `json:"minimality"`
	TotalRecomputedWork  int                      `json:"total_recomputed_work"`
	ReductionAttempts    int                      `json:"reduction_attempts"`
	EventWindow          map[string][]EventRecord `json:"event_window"`
	CausalChain          []CausalStep             `json:"causal_chain"`
	Causes               []diagnostic.Cause       `json:"causes"`
	EvidenceID           string                   `json:"evidence_id"`
}

// NewMismatchDocument constructs a MismatchDocument, refusing (rather than
// silently accepting) a Minimality value outside the two closed values
// Reduce ever produces (D-05-25) -- a third, quietly-introduced minimality
// state is exactly the kind of bug this constructor exists to make
// unconstructable. EventWindow is bounded to the last EventWindowSize
// entries per engine side regardless of how many the caller supplies
// (T-05-47), never the full trace.
func NewMismatchDocument(
	divergingAxis, enginePair, divergingOperationID string,
	reducedProgram core.Program,
	reducedSource, minimality string,
	totalRecomputedWork, reductionAttempts int,
	eventWindow map[string][]EventRecord,
	causalChain []CausalStep,
	causes []diagnostic.Cause,
	evidenceID string,
) (MismatchDocument, error) {
	if minimality != MinimalityFixpoint && minimality != MinimalityBudgetExhausted {
		return MismatchDocument{}, fmt.Errorf("reduce: mismatch document minimality must be %q or %q, got %q", MinimalityFixpoint, MinimalityBudgetExhausted, minimality)
	}
	reducedCoreBytes, err := json.Marshal(reducedProgram)
	if err != nil {
		return MismatchDocument{}, fmt.Errorf("reduce: mismatch document reduced core did not marshal: %w", err)
	}
	return MismatchDocument{
		Schema:               MismatchSchema,
		DivergingAxis:        divergingAxis,
		EnginePair:           enginePair,
		DivergingOperationID: divergingOperationID,
		ReducedCore:          string(reducedCoreBytes),
		ReducedSource:        reducedSource,
		Minimality:           minimality,
		TotalRecomputedWork:  totalRecomputedWork,
		ReductionAttempts:    reductionAttempts,
		EventWindow:          boundEventWindow(eventWindow),
		CausalChain:          append([]CausalStep(nil), causalChain...),
		Causes:               append([]diagnostic.Cause(nil), causes...),
		EvidenceID:           evidenceID,
	}, nil
}

// boundEventWindow clones window, truncating every engine's event slice to
// its own last EventWindowSize entries -- bounded regardless of how many
// the caller supplies (D-05-26/T-05-47).
func boundEventWindow(window map[string][]EventRecord) map[string][]EventRecord {
	bounded := make(map[string][]EventRecord, len(window))
	for engine, events := range window {
		if len(events) > EventWindowSize {
			events = events[len(events)-EventWindowSize:]
		}
		bounded[engine] = append([]EventRecord(nil), events...)
	}
	return bounded
}

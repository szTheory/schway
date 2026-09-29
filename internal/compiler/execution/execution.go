package execution

import (
	"bytes"
	"encoding/json"
)

const (
	Schema0 = "schway.execution/0"
	Schema1 = "schway.execution/1"
	Schema2 = "schway.execution/2"

	// MaxApplicationEvidenceBytes bounds a same-run application capture and
	// its published report independently from conformance execution documents.
	MaxApplicationEvidenceBytes = 64 * 1024
)

// MaxDocumentBytes is the native execution-document capture ceiling. Schema-2
// emission preflights against the same contract; compile diagnostics and
// stderr retain their smaller process-stream bounds.
const MaxDocumentBytes = 16 * 1024 * 1024

// Terminal outcome kinds (D-04-08): a closed named set carried as its own
// axis, separate from the returned value. OutcomeCancelled is reserved but
// deliberately unconstructible this phase -- async/cancellation is
// Out-of-Scope (PROJECT.md), and no engine (interp, cgen, native) may ever
// produce it. Adding real cancellation later is an enum member plus a
// reviewed change, never a quiet retrofit as a typed_failure alternative.
// core_test.go's TestCancelledOutcomeIsUnconstructible asserts this by
// scanning every engine's own source for the literal, not by observing that
// no test happened to produce one.
const (
	OutcomeReturned     = "returned"
	OutcomeTypedFailure = "typed_failure"
	OutcomeDefect       = "defect"
	OutcomeCancelled    = "cancelled"
)

// TerminalOutcomeKinds returns the complete closed set, in declaration
// order -- the single table a fail-closed control can assert every engine's
// emitted Outcome.Kind is drawn from.
func TerminalOutcomeKinds() []string {
	return []string{OutcomeReturned, OutcomeTypedFailure, OutcomeDefect, OutcomeCancelled}
}

type Outcome struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Event struct {
	Schema           string `json:"schema"`
	ID               string `json:"id"`
	Kind             string `json:"kind"`
	FunctionID       string `json:"function_id"`
	Input            string `json:"input,omitempty"`
	Output           string `json:"output,omitempty"`
	SourcePlace      string `json:"source_place,omitempty"`
	TargetPlace      string `json:"target_place,omitempty"`
	TypeID           string `json:"type_id,omitempty"`
	Invocation       string `json:"invocation,omitempty"`
	CalleeFunctionID string `json:"callee_function_id,omitempty"`
}

type Execution struct {
	Schema        string   `json:"schema"`
	Outcome       Outcome  `json:"outcome"`
	Events        []Event  `json:"events"`
	LiveResources []string `json:"live_resources"`

	// AllocatorAddress and WallClockNanos are physical observations D-05-21
	// explicitly excludes from Phase5CompareEngines' comparison (addresses
	// and wall-clock time are two of the five named exclusion categories).
	// Neither field is populated by the interpreter, cgen, or native.Runner
	// production paths today -- both stay the zero value on every real
	// execution document -- so their sole purpose is to give
	// session.Phase5ExcludedComparisonFields a genuine struct-field target
	// to route to, proving the fail-closed field-routing test's excluded
	// branch is load-bearing rather than vacuously empty.
	AllocatorAddress string `json:"allocator_address,omitempty"`
	WallClockNanos   int64  `json:"wall_clock_nanos,omitempty"`

	// ExitSignaled and ExitSignal are D-05-20's exit-status/signal axis
	// data: SC1 silently omitted exit status and signal from its comparison
	// scope, so these carry that fact as its own explicit, independently
	// comparable pair rather than an inference from Outcome.Kind. Neither
	// field is populated by production engines today (a defect's SIGABRT is
	// otherwise fully implied by Outcome.Kind=="defect"); they exist so
	// session.Phase5CompareEngines' axis:exit-status-signal has real data to
	// compare independently of axis:terminal-outcome.
	ExitSignaled bool   `json:"exit_signaled,omitempty"`
	ExitSignal   string `json:"exit_signal,omitempty"`
}

func CanonicalBytes(value Execution) ([]byte, error) { return json.Marshal(value) }

func Equal(left, right Execution) bool {
	leftBytes, leftErr := CanonicalBytes(left)
	rightBytes, rightErr := CanonicalBytes(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftBytes, rightBytes)
}

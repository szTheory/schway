package execution

import (
	"bytes"
	"encoding/json"
)

const (
	Schema0 = "lang.execution/0"
	Schema1 = "lang.execution/1"
)

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
	Schema      string `json:"schema"`
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	FunctionID  string `json:"function_id"`
	Input       string `json:"input,omitempty"`
	Output      string `json:"output,omitempty"`
	SourcePlace string `json:"source_place,omitempty"`
	TargetPlace string `json:"target_place,omitempty"`
	TypeID      string `json:"type_id,omitempty"`
}

type Execution struct {
	Schema        string   `json:"schema"`
	Outcome       Outcome  `json:"outcome"`
	Events        []Event  `json:"events"`
	LiveResources []string `json:"live_resources"`
}

func CanonicalBytes(value Execution) ([]byte, error) { return json.Marshal(value) }

func Equal(left, right Execution) bool {
	leftBytes, leftErr := CanonicalBytes(left)
	rightBytes, rightErr := CanonicalBytes(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftBytes, rightBytes)
}

package session

import (
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

// Phase 11 plan 11-09 (D-11-33) test wrappers: expose foreignCallSequenceFor
// and its two independent knowers directly, for the external
// session_phase5_mismatch_test.go / session_qlt05_reverify_test.go test
// files.

// ForeignCallSequenceForTest exposes foreignCallSequenceFor -- D-11-33's
// two-knower guard -- directly.
func ForeignCallSequenceForTest(program core.Program, o0Run execution.Execution) ([]string, error) {
	return foreignCallSequenceFor(program, o0Run)
}

// StaticForeignCallSequenceForTest exposes Knower 1 (the program-structure
// walk) in isolation.
func StaticForeignCallSequenceForTest(program core.Program) ([]string, error) {
	return staticForeignCallSequenceFor(program)
}

// DynamicForeignCallSequenceForTest exposes Knower 2 (the event-stream
// walk) in isolation.
func DynamicForeignCallSequenceForTest(program core.Program, o0Run execution.Execution) []string {
	return dynamicForeignCallSequenceFor(program, o0Run)
}

// SetForeignCallSequenceDisagreementForTest installs this plan's own
// fault-injection seam (foreignCallSequenceDisagreementForTest): when
// perturb is non-nil, foreignCallSequenceFor applies it to the dynamic
// derivation's result immediately before comparing it against the static
// one, proving the agreement check is load-bearing (able to FAIL) rather
// than trivially true. Unexported, restored immediately by every caller.
func SetForeignCallSequenceDisagreementForTest(perturb func([]string) []string) (restore func()) {
	previous := foreignCallSequenceDisagreementForTest
	foreignCallSequenceDisagreementForTest = perturb
	return func() { foreignCallSequenceDisagreementForTest = previous }
}

// SetQLT03InjectExtraRequiredCellForTest installs Task 3's D-11-50
// fault-injection seam (qlt03InjectExtraRequiredCellForTest): when
// disable is true, AuditQLT03Register's own cell enumeration includes one
// additional cell no committed row can ever satisfy, proving the audit is
// load-bearing (able to FAIL) without editing the committed
// qlt03_shape_register.json. Unexported, false in production, set only by
// a same-package test that defers the restore immediately.
func SetQLT03InjectExtraRequiredCellForTest(inject bool) (restore func()) {
	previous := qlt03InjectExtraRequiredCellForTest
	qlt03InjectExtraRequiredCellForTest = inject
	return func() { qlt03InjectExtraRequiredCellForTest = previous }
}

// setQLT03RegisterBytesForTest temporarily swaps the embedded register
// bytes LoadQLT03Register parses, for tests that need to feed the loader a
// deliberately malformed register (out-of-set mechanism, nonexistent
// falsifier test, both/neither populated) without touching the committed
// qlt03_shape_register.json on disk. Unexported, restored immediately by
// every caller.
func setQLT03RegisterBytesForTest(bytes []byte) (restore func()) {
	previous := qlt03RegisterBytes
	qlt03RegisterBytes = bytes
	return func() { qlt03RegisterBytes = previous }
}

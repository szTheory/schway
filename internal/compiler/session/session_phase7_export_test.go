package session

import (
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
)

// This file is compiled ONLY into test binaries (Go excludes every
// "_test.go" file from a non-test build), mirroring
// corevalidate/export_test.go's own precedent exactly. It exposes plan
// 07-04 Task 3's unexported phase07-lane fault-injection seams
// (phase07LaneDispatchFixturesOverride, phase07LaneRequiredKindsOverride)
// to the external session_test package's mutation-kill test
// (TestPhase7DispatchControlsMutationKilled_Phase07Lane,
// session_phase7_mutation_test.go). Every setter returns a restore func;
// callers MUST defer it immediately.

// SetPhase07LaneDispatchFixturesForTest installs Task 3 Test 2's
// dispatch-fixtures override for the phase07 lane and returns a restore
// func.
func SetPhase07LaneDispatchFixturesForTest(fixtures []string) (restore func()) {
	previous := phase07LaneDispatchFixturesOverride
	phase07LaneDispatchFixturesOverride = fixtures
	return func() { phase07LaneDispatchFixturesOverride = previous }
}

// SetPhase07LaneRequiredKindsForTest installs Task 3 Test 2's
// required-kinds override for the phase07 lane and returns a restore func.
func SetPhase07LaneRequiredKindsForTest(kinds []core.OperationKind) (restore func()) {
	previous := phase07LaneRequiredKindsOverride
	phase07LaneRequiredKindsOverride = kinds
	return func() { phase07LaneRequiredKindsOverride = previous }
}

// PeerRefusalDiagnosticForTest exposes 07-10 Task 1's unexported
// peerRefusalDiagnostic formatting helper to session_test's edge-2
// fail-closed assertion: corevalidate.Validate reporting !Valid with an
// EMPTY Problems slice is structurally impossible from any real program
// (v.problems is only ever populated alongside setting v.valid = false via
// v.check/v.checkErr), so the fail-closed fallback can only be exercised
// with a hand-built corevalidate.Result.
func PeerRefusalDiagnosticForTest(validated corevalidate.Result, moduleID string) diagnostic.Diagnostic {
	return peerRefusalDiagnostic(validated, moduleID)
}

// PeerRefusalUnnamedCodeForTest exposes 07-10's fail-closed fallback code
// constant so tests assert against the named constant rather than a
// duplicated string literal.
const PeerRefusalUnnamedCodeForTest = peerRefusalUnnamedCode

// SetCheckCommandPeerSeamForTest installs 07-10 Task 1's
// control:check.peer_consulted fault-injection seam and returns a restore
// func the caller MUST defer immediately.
func SetCheckCommandPeerSeamForTest(disable bool) (restore func()) {
	previous := checkCommandPeerSeam
	checkCommandPeerSeam = disable
	return func() { checkCommandPeerSeam = previous }
}

// SetInterfacePeerRefusalSeamForTest installs 07-10 Task 2's
// control:interface.peer_refusal_is_invalid fault-injection seam and
// returns a restore func the caller MUST defer immediately.
func SetInterfacePeerRefusalSeamForTest(disable bool) (restore func()) {
	previous := interfacePeerRefusalSeam
	interfacePeerRefusalSeam = disable
	return func() { interfacePeerRefusalSeam = previous }
}

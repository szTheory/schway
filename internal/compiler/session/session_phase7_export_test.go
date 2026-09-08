package session

import "github.com/codename-lang/lang/internal/compiler/core"

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

package callgraph

// This file is compiled ONLY into test binaries (Go excludes every
// "_test.go" file from a non-test build), mirroring
// session/session_phase7_export_test.go's own precedent exactly. It
// exposes Task 2's unexported determinism fault-injection seams
// (rotationDisabledForTest, witnessSelectionDisabledForTest) to the
// external callgraph_test package's mutation-kill tests. Every setter
// returns a restore func; callers MUST defer it immediately.

// SetRotationDisabledForTest installs Task 2 Test 3's rotation-removal
// seam and returns a restore func.
func SetRotationDisabledForTest(disabled bool) (restore func()) {
	previous := rotationDisabledForTest
	rotationDisabledForTest = disabled
	return func() { rotationDisabledForTest = previous }
}

// SetWitnessSelectionDisabledForTest installs Task 2 Test 3's
// witness-selection-removal seam and returns a restore func.
func SetWitnessSelectionDisabledForTest(disabled bool) (restore func()) {
	previous := witnessSelectionDisabledForTest
	witnessSelectionDisabledForTest = disabled
	return func() { witnessSelectionDisabledForTest = previous }
}

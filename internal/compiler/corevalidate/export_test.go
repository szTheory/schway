package corevalidate

// This file is compiled ONLY into test binaries (Go excludes every
// "_test.go" file from a non-test build, so none of these symbols exist in
// the compiled corevalidate library at all -- stronger than D-07-42's
// unexported-at-runtime seams, since production code cannot even reference
// them). It exposes 07-02 Task 3's unexported fault-injection seams
// (forcePeerCallableAlwaysTrue, disableSummaryPeerAtReplayBlocksForTest) to
// the cross-package fault tests in package corevalidate_test, which need
// this package's own peer seams flipped alongside originvalidate's producer
// seams (export_test.go there) in the same bilateral-fault test. Every
// setter returns a restore func; callers MUST defer it immediately.

// SetPeerCallableForceOverrideForTest installs faults 3/5's peer-side
// Callable-always-true seam and returns a restore func.
func SetPeerCallableForceOverrideForTest(force bool) (restore func()) {
	previous := forcePeerCallableAlwaysTrue
	forcePeerCallableAlwaysTrue = force
	return func() { forcePeerCallableAlwaysTrue = previous }
}

// SetDisableSummaryPeerAtReplayBlocksForTest installs fault 2's
// replayBlocks-only peer-wiring-disable seam and returns a restore func.
func SetDisableSummaryPeerAtReplayBlocksForTest(disable bool) (restore func()) {
	previous := disableSummaryPeerAtReplayBlocksForTest
	disableSummaryPeerAtReplayBlocksForTest = disable
	return func() { disableSummaryPeerAtReplayBlocksForTest = previous }
}

// SetDisableCalleeResolutionCheckForTest installs Task 2's (07-03)
// resolves-to-a-declared-function seam and returns a restore func. Callers
// MUST defer the restore immediately -- see the seam's own doc comment.
func SetDisableCalleeResolutionCheckForTest(disable bool) (restore func()) {
	previous := disableCalleeResolutionCheckForTest
	disableCalleeResolutionCheckForTest = disable
	return func() { disableCalleeResolutionCheckForTest = previous }
}

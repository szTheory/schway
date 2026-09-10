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

// SetClosureDigestEmptyCalleesForTest installs 07-08 Task 2's own
// empty-callee-pairs fault-injection seam on THIS package's independent
// chained-digest re-derivation and returns a restore func.
func SetClosureDigestEmptyCalleesForTest(force bool) (restore func()) {
	previous := closureDigestEmptyCalleesForTest
	closureDigestEmptyCalleesForTest = force
	return func() { closureDigestEmptyCalleesForTest = previous }
}

// SetClosureDigestDiscoveryOrderForTest installs 07-08 Task 2's own
// discovery-order (non-postorder) fault-injection seam on THIS package's
// independent chained-digest re-derivation and returns a restore func.
func SetClosureDigestDiscoveryOrderForTest(force bool) (restore func()) {
	previous := closureDigestDiscoveryOrderForTest
	closureDigestDiscoveryOrderForTest = force
	return func() { closureDigestDiscoveryOrderForTest = previous }
}

// SetDisableCallArgumentTypePeerForTest installs 07-09 Task 2's independent
// argument-type peer-disable seam and returns a restore func. Callers MUST
// defer the restore immediately.
func SetDisableCallArgumentTypePeerForTest(disable bool) (restore func()) {
	previous := disableCallArgumentTypePeerForTest
	disableCallArgumentTypePeerForTest = disable
	return func() { disableCallArgumentTypePeerForTest = previous }
}

// SetDisableCallReturnTypePeerForTest installs 07-09 Task 2's independent
// return-type peer-disable seam and returns a restore func. Callers MUST
// defer the restore immediately.
func SetDisableCallReturnTypePeerForTest(disable bool) (restore func()) {
	previous := disableCallReturnTypePeerForTest
	disableCallReturnTypePeerForTest = disable
	return func() { disableCallReturnTypePeerForTest = previous }
}

// SetDisableCallArgumentConsumePeerForTest installs 07-11's independent
// call-argument-consume peer-disable seam (PVG-01/CR-01) and returns a
// restore func. Callers MUST defer the restore immediately.
func SetDisableCallArgumentConsumePeerForTest(disable bool) (restore func()) {
	previous := disableCallArgumentConsumePeerForTest
	disableCallArgumentConsumePeerForTest = disable
	return func() { disableCallArgumentConsumePeerForTest = previous }
}

// SetForceCallArgumentConsumePeerForTest installs 07-11's over-refusal
// fault-injection seam (a copyable argument consumed regardless of its copy
// ability) and returns a restore func. Callers MUST defer the restore
// immediately.
func SetForceCallArgumentConsumePeerForTest(force bool) (restore func()) {
	previous := forceCallArgumentConsumePeerForTest
	forceCallArgumentConsumePeerForTest = force
	return func() { forceCallArgumentConsumePeerForTest = previous }
}

// SetDisableForeignClosureJoinPeerForTest installs 07-12's independent
// Foreign-closure-join peer-disable seam
// (control:summary.foreign_reach_closure_derived, peer side) and returns a
// restore func. Callers MUST defer the restore immediately.
func SetDisableForeignClosureJoinPeerForTest(disable bool) (restore func()) {
	previous := disableForeignClosureJoinPeerForTest
	disableForeignClosureJoinPeerForTest = disable
	return func() { disableForeignClosureJoinPeerForTest = previous }
}

// SetDisableFailsClosureJoinPeerForTest installs 07-12's independent
// Fails-closure-join peer-disable seam
// (control:summary.fails_closure_derived, peer side) and returns a restore
// func. Callers MUST defer the restore immediately.
func SetDisableFailsClosureJoinPeerForTest(disable bool) (restore func()) {
	previous := disableFailsClosureJoinPeerForTest
	disableFailsClosureJoinPeerForTest = disable
	return func() { disableFailsClosureJoinPeerForTest = previous }
}

// SetDisablePeerOriginContainmentForTest installs Task 1's D-09-19
// origin-containment-disable seam and returns a restore func. Callers MUST
// defer the restore immediately.
func SetDisablePeerOriginContainmentForTest(disable bool) (restore func()) {
	previous := disablePeerOriginContainmentForTest
	disablePeerOriginContainmentForTest = disable
	return func() { disablePeerOriginContainmentForTest = previous }
}

// SetDisableForeignOriginPeerForTest installs Task 2's D-09-20
// foreign-origin-omitted-disable seam and returns a restore func. Callers
// MUST defer the restore immediately.
func SetDisableForeignOriginPeerForTest(disable bool) (restore func()) {
	previous := disablePeerForeignOriginPeerForTest
	disablePeerForeignOriginPeerForTest = disable
	return func() { disablePeerForeignOriginPeerForTest = previous }
}

// SetPeerClosureUnmemoizedSeamForTest installs Phase 09 Plan 06's own
// QLT-08 mutation-kill seam (peerClosureUnmemoizedSeamForTest,
// corevalidate.go, beside foldChain) and returns a restore func. Callers
// MUST defer the restore immediately.
func SetPeerClosureUnmemoizedSeamForTest(force bool) (restore func()) {
	previous := peerClosureUnmemoizedSeamForTest
	peerClosureUnmemoizedSeamForTest = force
	return func() { peerClosureUnmemoizedSeamForTest = previous }
}

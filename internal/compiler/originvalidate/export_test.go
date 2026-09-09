package originvalidate

// This file is compiled ONLY into test binaries (Go excludes every
// "_test.go" file from a non-test build, so none of these symbols exist in
// the compiled originvalidate library at all -- stronger than D-07-42's
// unexported-at-runtime seams, since production code cannot even reference
// them). It exposes 07-02 Task 3's unexported fault-injection seams
// (typeFactExactIDMatchOverride, forceCallableAlwaysTrue) to this package's
// OWN external test package (originvalidate_test, same directory) --
// package foo_test can reach only foo's exported identifiers, exactly like
// any other importer, so this is the standard Go export_test.go pattern for
// letting black-box tests reach otherwise-unexported internals. Every
// setter returns a restore func; callers MUST defer it immediately (the
// same discipline pathoracle_test.go's TestTerminatorWalkMutationKilled
// established for its own override seam).

// SetTypeFactExactIDMatchOverrideForTest installs fault 1's seam and
// returns a restore func.
func SetTypeFactExactIDMatchOverrideForTest(override func(factID, wantID string) bool) (restore func()) {
	previous := typeFactExactIDMatchOverride
	typeFactExactIDMatchOverride = override
	return func() { typeFactExactIDMatchOverride = previous }
}

// SetCallableForceOverrideForTest installs faults 3/4's producer-side
// Callable-always-true seam and returns a restore func.
func SetCallableForceOverrideForTest(force bool) (restore func()) {
	previous := forceCallableAlwaysTrue
	forceCallableAlwaysTrue = force
	return func() { forceCallableAlwaysTrue = previous }
}

// SetClosureDigestEmptyCalleesOverrideForTest installs 07-08 Task 2's
// empty-callee-pairs fault-injection seam and returns a restore func.
func SetClosureDigestEmptyCalleesOverrideForTest(force bool) (restore func()) {
	previous := closureDigestEmptyCalleesOverride
	closureDigestEmptyCalleesOverride = force
	return func() { closureDigestEmptyCalleesOverride = previous }
}

// SetClosureDigestDiscoveryOrderOverrideForTest installs 07-08 Task 2's
// discovery-order (non-reverse-postorder) fault-injection seam and returns
// a restore func.
func SetClosureDigestDiscoveryOrderOverrideForTest(force bool) (restore func()) {
	previous := closureDigestDiscoveryOrderOverride
	closureDigestDiscoveryOrderOverride = force
	return func() { closureDigestDiscoveryOrderOverride = previous }
}

// SetClosureDigestComputationOrderObservedForTest installs 07-08 Task 1's
// ordering-instrumentation seam and returns a restore func.
func SetClosureDigestComputationOrderObservedForTest(observer func(functionID string)) (restore func()) {
	previous := closureDigestComputationOrderObserved
	closureDigestComputationOrderObserved = observer
	return func() { closureDigestComputationOrderObserved = previous }
}

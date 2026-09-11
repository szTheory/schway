package pathoracle

// This file is compiled ONLY into test binaries (Go excludes every
// "_test.go" file from a non-test build, so none of these symbols exist in
// the compiled pathoracle library at all). It exposes 10-03 Task 2's D-10-55
// seeded-fault seam (forceContractHopForTest, pathoracle_compose.go) to
// this package's OWN external test package (pathoracle_test, same
// directory) -- package foo_test can reach only foo's exported
// identifiers, exactly like any other importer, so this is the standard Go
// export_test.go pattern for letting black-box tests reach otherwise-
// unexported internals, mirroring originvalidate/export_test.go's own
// SetDisableOpCallOriginConsultForTest precedent exactly. Callers MUST
// defer the returned restore func immediately.

// SetForceContractHopForTest installs/lifts the D-10-55 stubbed
// contract-hop seam and returns a restore func.
func SetForceContractHopForTest(force func(calleeID string) bool) (restore func()) {
	previous := forceContractHopForTest
	forceContractHopForTest = force
	return func() { forceContractHopForTest = previous }
}

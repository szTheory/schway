package session

// This file is compiled ONLY into test binaries. It exposes plan 30-08's
// unexported fault-injection and observation seams to the external
// session_test package, following session_phase7_export_test.go. Every setter
// returns a restore func; callers MUST defer it immediately.

// SetCheckCommandPeerObserverForTest installs an observer that records which
// independent admission peers CheckCommandFile consulted, in order.
func SetCheckCommandPeerObserverForTest(observer func(string)) (restore func()) {
	previous := checkCommandPeerObservedForTest
	checkCommandPeerObservedForTest = observer
	return func() { checkCommandPeerObservedForTest = previous }
}

// SetDebugMapGuardNoImportForTest makes the no-import debug-map branch apply
// the MaxCheckWork guard too: the armed control for the D-15 claim that the
// no-import path stays unguarded.
func SetDebugMapGuardNoImportForTest(on bool) (restore func()) {
	previous := debugMapGuardNoImportForTest
	debugMapGuardNoImportForTest = on
	return func() { debugMapGuardNoImportForTest = previous }
}

// UnitOfFunctionIDForTest exposes unitOfFunctionID.
func UnitOfFunctionIDForTest(id, entryModule string) string {
	return unitOfFunctionID(id, entryModule)
}

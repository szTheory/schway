package session

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

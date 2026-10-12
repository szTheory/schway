package protocol

// SetAlwaysPrintUnitForTest arms the negative-control seam that makes the
// human projection print unit= for entry-module diagnostics. It returns a
// restore function.
func SetAlwaysPrintUnitForTest(on bool) func() {
	previous := alwaysPrintUnitForTest
	alwaysPrintUnitForTest = on
	return func() { alwaysPrintUnitForTest = previous }
}

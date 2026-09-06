package reduce

// Test-only accessors for D-05-27's mutation-kill seam. These make
// testOnlyMoves visible to the external reduce_test package without
// exporting it from the production API -- mirroring cgen's own
// export_test.go convention (internal/compiler/cgen/export_test.go).

// SetTestOnlyMoves overrides the move list Reduce applies. Test-only:
// production code must never call this.
func SetTestOnlyMoves(moves func() []Move) { testOnlyMoves = moves }

// ResetTestOnlyMoves restores Reduce's real, fixed Moves() order. Every
// test that calls SetTestOnlyMoves must defer this.
func ResetTestOnlyMoves() { testOnlyMoves = nil }

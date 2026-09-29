package reduce

import "github.com/szTheory/schway/internal/compiler/core"

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

// Phase 11 (D-11-29) test wrappers: expose the two new whole-program moves
// and the derived attempt budget for direct, isolated testing without
// requiring a full Reduce run.

// DropCallSiteForTest exposes dropCallSite directly.
func DropCallSiteForTest(p core.Program) (core.Program, bool) { return dropCallSite(p) }

// DropOrphanFunctionForTest exposes dropOrphanFunction directly, setting
// (and restoring) the package-level entryFunctionID Reduce would otherwise
// set from Seed.EntryFunctionID.
func DropOrphanFunctionForTest(p core.Program, entryID string) (core.Program, bool) {
	previous := entryFunctionID
	entryFunctionID = entryID
	defer func() { entryFunctionID = previous }()
	return dropOrphanFunction(p)
}

// DerivedAttemptBoundForTest exposes derivedAttemptBound directly.
func DerivedAttemptBoundForTest(p core.Program) int { return derivedAttemptBound(p) }

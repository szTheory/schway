package native

import "testing"

// TestForeignArenaSymbolResolves is task 05-08-01's resolution proof
// (D-05-08): the third frozen foreign TU's declared symbol resolves
// through the SAME ForeignSourcePathForSymbol switch every other frozen TU
// resolves through -- no new plumbing.
func TestForeignArenaSymbolResolves(t *testing.T) {
	path, ok := ForeignSourcePathForSymbol("schway_arena_open")
	if !ok {
		t.Fatal("schway_arena_open did not resolve")
	}
	if path != ForeignArenaSourcePath() {
		t.Fatalf("resolved path %q does not match ForeignArenaSourcePath() %q", path, ForeignArenaSourcePath())
	}
}

// TestExistingForeignSymbolsStillResolve guards against a third TU
// silently displacing either pre-existing frozen TU's own resolution --
// the D-04-10 lesson recorded in ForeignSourcePathForSymbol's own doc
// comment (hardcoding one TU silently breaks linking any program declaring
// another).
func TestExistingForeignSymbolsStillResolve(t *testing.T) {
	resourcePath, ok := ForeignSourcePathForSymbol("schway_res_open")
	if !ok || resourcePath != ForeignResourceSourcePath() {
		t.Fatalf("schway_res_open resolution regressed: path=%q ok=%v", resourcePath, ok)
	}
	nonlocalPath, ok := ForeignSourcePathForSymbol("schway_nonlocal_probe")
	if !ok || nonlocalPath != ForeignNonlocalSourcePath() {
		t.Fatalf("schway_nonlocal_probe resolution regressed: path=%q ok=%v", nonlocalPath, ok)
	}
}

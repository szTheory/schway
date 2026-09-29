package native

import (
	"path/filepath"
	"runtime"
)

// ForeignNonlocalSourcePath resolves the repo-relative path to the second
// frozen, hand-written foreign translation unit (D-04-17/D-04-18): the
// process-root landing pad's reachable witness. Uses the same build-time
// source location convention as ForeignResourceSourcePath, for the same
// reason -- the frozen TU is committed, versioned repository source, not a
// runtime-configurable dependency.
func ForeignNonlocalSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "native", "schway_foreign_nonlocal.c")
}

// ForeignSourcePathForSymbol resolves a declared `foreign C {}` symbol name
// to the frozen translation unit that defines it (D-04-10/D-04-17). This
// A foreign interface identifies its frozen translation unit by symbol
// name. The former executable foreign-body emitter linked one extern per
// function, but whole-program native emission now refuses that body shape;
// this resolver remains for the native fixture/tooling boundary and must not
// be read as emitter admission.
func ForeignSourcePathForSymbol(symbol string) (string, bool) {
	switch symbol {
	case "schway_res_open":
		return ForeignResourceSourcePath(), true
	case "schway_nonlocal_probe":
		return ForeignNonlocalSourcePath(), true
	case "schway_arena_open":
		return ForeignArenaSourcePath(), true
	case "schway_retained_touch":
		return ForeignRetainedSourcePath(), true
	default:
		return "", false
	}
}

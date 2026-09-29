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
// to the first frozen translation unit that defines it (D-04-10/D-04-17).
// Use ForeignSourcePathsForSymbol when linking: historical fixture symbols
// need both their current implementation and the test-only alias adapter.
func ForeignSourcePathForSymbol(symbol string) (string, bool) {
	paths := ForeignSourcePathsForSymbol(symbol)
	if len(paths) == 0 {
		return "", false
	}
	return paths[0], true
}

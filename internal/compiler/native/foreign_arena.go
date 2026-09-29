package native

import (
	"path/filepath"
	"runtime"
)

// ForeignArenaSourcePath resolves the repo-relative path to the third
// frozen, hand-written foreign translation unit (D-05-08): the dynamic
// allocator-mismatch subject. Uses the same build-time source location
// convention as ForeignResourceSourcePath/ForeignNonlocalSourcePath, for
// the same reason -- the frozen TU is committed, versioned repository
// source, not a runtime-configurable dependency.
func ForeignArenaSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "native", "schway_foreign_arena.c")
}

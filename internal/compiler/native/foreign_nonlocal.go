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
	return filepath.Join(root, "native", "lang_foreign_nonlocal.c")
}

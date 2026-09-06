package native

import (
	"path/filepath"
	"runtime"
)

// ForeignRetainedSourcePath resolves the repo-relative path to the fourth
// frozen, hand-written foreign translation unit (D-05-06/D-05-09): the
// hostile retained-pointer / use-after-free variant. Uses the same
// build-time source location convention as ForeignResourceSourcePath, for
// the same reason -- the frozen TU is committed, versioned repository
// source, not a runtime-configurable dependency.
func ForeignRetainedSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "native", "lang_foreign_retained.c")
}

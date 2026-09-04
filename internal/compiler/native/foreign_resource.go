package native

import (
	"path/filepath"
	"runtime"
)

// ForeignResourceSourcePath resolves the repo-relative path to the frozen,
// hand-written foreign translation unit (D-04-10), via THIS FILE's own
// build-time source location rather than the running binary's working
// directory -- exactly the pattern testsupport.ProjectPath already uses for
// tests, generalized to a production code path. This is a deliberate,
// narrow exception to "no filesystem path assumptions at runtime": the
// frozen TU is committed, versioned repository source, not a
// runtime-configurable dependency, so resolving it relative to the
// compiler's own build location is correct regardless of the invoking
// process's CWD.
func ForeignResourceSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "native", "lang_foreign_resource.c")
}

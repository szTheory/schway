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

// ForeignSourcePathForSymbol resolves a declared `foreign C {}` symbol name
// to the frozen translation unit that defines it (D-04-10/D-04-17). This
// project's model is one symbol per foreign-acquiring function
// (cgen.go's emitLinearForeign calls exactly one extern per function), so a
// caller needing to auto-link a foreign-shaped program's frozen TU (see
// session.RunNative) resolves it by symbol name rather than by hardcoding a
// single path -- the moment a second frozen TU exists (this plan adds one),
// hardcoding the first one silently breaks linking any program declaring
// the second.
func ForeignSourcePathForSymbol(symbol string) (string, bool) {
	switch symbol {
	case "lang_res_open":
		return ForeignResourceSourcePath(), true
	case "lang_nonlocal_probe":
		return ForeignNonlocalSourcePath(), true
	default:
		return "", false
	}
}

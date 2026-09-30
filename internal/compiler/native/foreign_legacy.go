package native

import (
	"path/filepath"
	"runtime"
)

// ForeignLegacyResourceAdapterSourcePath resolves the test-only resource
// alias used by frozen Phase 1-5 fixtures.
func ForeignLegacyResourceAdapterSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "testdata", "phase16", "historical", "foreign_phase1_5_legacy_adapter.c")
}

// ForeignLegacyNonlocalAdapterSourcePath resolves the test-only wrapper that
// compiles the current nonlocal implementation against the old landing-pad
// symbol defined by byte-frozen Phase 4/5 C artifacts.
func ForeignLegacyNonlocalAdapterSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "testdata", "phase16", "historical", "foreign_phase1_5_legacy_nonlocal_adapter.c")
}

// ForeignLegacyArenaAdapterSourcePath resolves the test-only allocator alias
// used by the frozen Phase 5 allocator fixture.
func ForeignLegacyArenaAdapterSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "testdata", "phase16", "historical", "foreign_phase1_5_legacy_arena_adapter.c")
}

// ForeignLegacyRetainedAdapterSourcePath resolves the test-only retained
// pointer alias used by frozen Phase 5 fixture inputs.
func ForeignLegacyRetainedAdapterSourcePath() string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(root, "testdata", "phase16", "historical", "foreign_phase1_5_legacy_retained_adapter.c")
}

// ForeignSourcePathsForSymbol returns every translation unit required to
// link a declared foreign symbol. Historical Phase 1-5 fixtures retain their
// original call names and use the adapter only at the native test boundary.
func ForeignSourcePathsForSymbol(symbol string) []string {
	switch symbol {
	case "schway_res_open":
		return []string{ForeignResourceSourcePath()}
	case "schway_nonlocal_probe":
		return []string{ForeignNonlocalSourcePath()}
	case "schway_arena_open":
		return []string{ForeignArenaSourcePath()}
	case "schway_retained_touch":
		return []string{ForeignRetainedSourcePath()}
	case "lang_res_open":
		return ForeignLegacyResourceSourcePaths()
	case "lang_nonlocal_probe":
		return ForeignLegacyNonlocalSourcePaths()
	case "lang_arena_open":
		return ForeignLegacyArenaSourcePaths()
	case "lang_retained_touch":
		return ForeignLegacyRetainedSourcePaths()
	default:
		return nil
	}
}

// ForeignLegacyResourceSourcePaths returns the implementation and test alias
// units used by the frozen Phase 1-5 resource fixtures.
func ForeignLegacyResourceSourcePaths() []string {
	return []string{ForeignResourceSourcePath(), ForeignLegacyResourceAdapterSourcePath()}
}

// ForeignLegacyNonlocalSourcePaths returns the implementation and test alias
// units used by the frozen Phase 4 nonlocal-exit fixture.
func ForeignLegacyNonlocalSourcePaths() []string {
	return []string{ForeignLegacyNonlocalAdapterSourcePath()}
}

// ForeignLegacyPhase4SourcePaths returns the deduplicated legacy adapters
// needed by frozen Phase 4 fixtures that combine resource and nonlocal calls.
func ForeignLegacyPhase4SourcePaths() []string {
	return []string{ForeignLegacyNonlocalAdapterSourcePath(), ForeignResourceSourcePath(), ForeignLegacyResourceAdapterSourcePath()}
}

// ForeignLegacyArenaSourcePaths returns the implementation and test alias
// units used by the frozen Phase 5 allocator fixture.
func ForeignLegacyArenaSourcePaths() []string {
	return []string{ForeignArenaSourcePath(), ForeignLegacyArenaAdapterSourcePath()}
}

// ForeignLegacyRetainedSourcePaths returns the implementation and test alias
// units used by frozen retained-pointer fixture inputs.
func ForeignLegacyRetainedSourcePaths() []string {
	return []string{ForeignRetainedSourcePath(), ForeignLegacyRetainedAdapterSourcePath()}
}

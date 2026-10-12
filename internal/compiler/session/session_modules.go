package session

import (
	"strings"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/modload"
)

// Phase 30 (D-22): every module-aware file route reads its source through the
// one helper below. loadSourcePath is the only modload.Load call in the
// session package; modload itself is the only code that opens a module file.

// debugMapGuardNoImportForTest, when set, applies the MaxCheckWork guard to
// the no-import debug-map branch too. It is the armed control for the D-15
// claim that the no-import path keeps its pre-change, unguarded behavior.
var debugMapGuardNoImportForTest bool

// loadSourcePath loads the entry at path and every module it imports. The
// error is non-nil only when the entry itself cannot be read, which keeps the
// operational class of a missing entry file.
func loadSourcePath(path string) (modload.Set, []diagnostic.Diagnostic, error) {
	return modload.Load(path)
}

// checkSourcePath is the shared file-route check. A program with no import
// declarations is handed to the unchanged byte-taking Check (D-15), so every
// no-import result is byte-identical; a program with imports is checked as one
// merged program. The returned bytes are the ENTRY bytes only: the build
// receipt SourceDigest stays the entry digest (module hashing is Phase 31).
func checkSourcePath(path string) (CheckResult, []byte, error) {
	set, diagnostics, err := loadSourcePath(path)
	if err != nil {
		return CheckResult{}, nil, err
	}
	entry := set.Sources[0].Bytes
	if len(set.Program.Imports) == 0 {
		return Check(entry), entry, nil
	}
	return checkLoaded(set, diagnostics), entry, nil
}

// checkLoaded checks an import-bearing loaded set: loader diagnostics first,
// otherwise check.Program over the merged program with Work summed across every
// module (D-21) under the same MaxCheckWork refusal Check applies.
func checkLoaded(set modload.Set, diagnostics []diagnostic.Diagnostic) CheckResult {
	if len(diagnostics) > 0 {
		return CheckResult{Tree: set.Trees[""], Diagnostics: append([]diagnostic.Diagnostic(nil), diagnostics...), Work: set.Work}
	}
	result := CheckResult{Tree: set.Trees[""], UnitTrees: set.Trees, Work: set.Work}
	checked := check.Program(set.Program)
	result.Work += checked.Work
	if result.Work > MaxCheckWork {
		result.Diagnostics = append(result.Diagnostics, diagnostic.Error("check.work_limit", diagnostic.Span{}, "checking exceeds the declared work limit"))
		return result
	}
	result.Program = checked.Program
	result.Diagnostics = append(result.Diagnostics, checked.Diagnostics...)
	return result
}

// moduleOfSemanticID returns the module segment of an `s1:<module>:...` ID, or
// "" when id has no such form.
func moduleOfSemanticID(id string) string {
	rest, ok := strings.CutPrefix(id, "s1:")
	if !ok {
		return ""
	}
	module, _, found := strings.Cut(rest, ":")
	if !found {
		return ""
	}
	return module
}

// unitOfFunctionID returns the unit a function (or operation, place) ID
// belongs to: the module segment of the ID, or "" for the entry module (whose
// diagnostics carry no unit, D-11).
func unitOfFunctionID(id, entryModule string) string {
	module := moduleOfSemanticID(id)
	if module == "" || module == entryModule {
		return ""
	}
	return module
}

// unitOfDetail finds the first semantic ID in a peer problem detail and returns
// its unit. A detail with no ID, or a moduleID that is not in `s1:` form (so
// the entry module is unknown), yields "".
func unitOfDetail(detail, moduleID string) string {
	entryModule := moduleOfSemanticID(moduleID)
	if entryModule == "" {
		return ""
	}
	index := strings.Index(detail, "s1:")
	if index < 0 {
		return ""
	}
	end := index
	for end < len(detail) && detail[end] != ' ' {
		end++
	}
	return unitOfFunctionID(detail[index:end], entryModule)
}

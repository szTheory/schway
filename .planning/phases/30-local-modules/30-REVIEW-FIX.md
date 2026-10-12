---
phase: 30-local-modules
fixed_at: 2026-10-12T00:00:00Z
review_path: .planning/phases/30-local-modules/30-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 30: Code Review Fix Report

**Fixed at:** 2026-10-12
**Source review:** .planning/phases/30-local-modules/30-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 6 (CR-01, WR-01..WR-06; IN-* out of scope)
- Fixed: 6
- Skipped: 0

Verification ran in an isolated worktree. Each touched package was run in full before its commit: modload, check, corevalidate, session and cmd/schway. All passed. The worktree was fast-forwarded into `phase26-validation` and removed. `testdata/phase30/prechange-corpus.json` was not touched.

## Fixed Issues

### CR-01: Cap diagnostics swallowed, users got module.imports_not_loaded
**Files modified:** `internal/compiler/modload/load.go`, `internal/compiler/session/session_phase30_caps_test.go` (new), `cmd/schway/main_phase30_exit_test.go`
**Commit:** 11d64746
**Applied fix:** `loadGraph` now returns the entry program, imports retained, alongside the cap diagnostics. Callers that branch on `len(Program.Imports)` therefore take the module path and report the D-13 code. Session tests prove `module.count_limit`, `module.depth_limit` and `module.total_bytes_limit` reach `CheckCommandFile` and `DebugMapCommandFile`. They also assert the output never contains `module.imports_not_loaded`. A CLI test covers `check` and `debug-map` with `module.count_limit` and exit 2 (`exitcode.Invalid`). The armed control is a within-cap program that names no cap code. With the fix reverted, all session subtests and the CLI test were observed to fail.

### WR-01: Entry self-import inflated the module count
**Files modified:** `internal/compiler/modload/load.go`, `internal/compiler/modload/caps_test.go`
**Commit:** 105f71b5
**Applied fix:** An import naming the entry file's path-derived name is aliased to the entry at enqueue time. It no longer counts toward the cap. The `process`-time same-file check stays for aliases that only file identity reveals, such as hard links. Test: a 16-module program with a self-import is not refused, and still reports `module.import_cycle`. The control seam `entryAliasAtProcessForTest` was observed to reproduce the false `count_limit`.

### WR-02: Repeated cap diagnostics for the same refused target
**Files modified:** `internal/compiler/modload/load.go`, `internal/compiler/modload/caps_test.go`
**Commit:** 9f7962fb
**Applied fix:** A `refused` set makes each cap-refused target report once. The control seam `repeatCapDiagForTest` was observed to produce at least 3 duplicates.

### WR-03: Exported but undeclared name passed check, leaving a dangling import fact
**Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_phase30_test.go`
**Commit:** 6df11250
**Applied fix:** `resolveImports` now requires the target module to declare a function of the imported name. Otherwise it refuses with `module.not_exported` and an `export_list` cause, and emits no `ModuleImport` fact. Control seam `skipDeclaredImportCheckForTest`.

### WR-04: Import colliding with an entry foreign symbol not refused at the import
**Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_phase30_test.go`
**Commit:** 05fade38
**Applied fix:** For the entry unit, foreign symbol names join the collision set. The import is refused with `module.import_collision` and a `local_declaration` cause. Control seam `skipForeignImportCollisionForTest`.
Not applied: the review also suggested data constructor names. Constructors are capitalized type-level names and not callable fn names, so I could not confirm a real collision path. This is a partial application of the suggestion.

### WR-05: Peer-refusal unit taken from the first s1: ID in the detail
**Files modified:** `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/session/session.go`, `internal/compiler/session/session_modules.go`, `internal/compiler/session/session_phase7_export_test.go`, `internal/compiler/session/session_phase30_modules_test.go`, `testdata/phase16/public-emitter-consumers.json`
**Commit:** 26e905e3
**Applied fix:** `corevalidate.Problem` gained a `FunctionID` field (`json:"-"`, so no wire change). It is set to the caller for the two visibility codes. `peerRefusalDiagnostic` attributes the unit from it and falls back to detail parsing otherwise. A forged-core test covers an entry caller (unit empty) and a library caller (its own unit). The control seam restores detail-text attribution and was observed to blame `lib`. With the fix reverted, the new test was observed to fail. The line-pinned emitter registry entry (`EmitNative`, 3089 to 3098) was updated because `session.go` shifted by 9 lines.
Status: fixed. This is an attribution logic change, so human verification is recommended.

### WR-06: All non-NotExist I/O failures reported as "resolves outside the entry directory"
**Files modified:** `internal/compiler/modload/resolve.go`, `internal/compiler/modload/resolve_test.go`
**Commit:** 093c2fa5
**Applied fix:** `ioRefusalCode` maps only errors containing os.Root's "path escapes from parent" to `module.escape`. Everything else becomes `module.read_failed`. Go 1.24 exposes no sentinel for this error, so the match is on text. I confirmed the exact text against a real `os.Root`. An unrecognized error fails closed as `read_failed`. The leaky-path test now expects `read_failed` for a generic denial and `escape` for a root-escape error. Control seam `ioErrorAlwaysEscapeForTest`.
Status: fixed, but depends on os.Root error text. Re-check on Go upgrades.

## Skipped Issues

None. IN-01 and IN-02 are outside the `critical_warning` scope.

---

_Fixed: 2026-10-12_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_

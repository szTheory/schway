---
phase: 30-local-modules
reviewed: 2026-10-12T00:00:00Z
depth: standard
files_reviewed: 28
files_reviewed_list:
  - internal/compiler/modload/fs.go
  - internal/compiler/modload/resolve.go
  - internal/compiler/modload/load.go
  - internal/compiler/modload/merge.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/token.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/format.go
  - internal/compiler/ast/ast.go
  - internal/compiler/diagnostic/diagnostic.go
  - internal/compiler/protocol/protocol.go
  - internal/compiler/debugmap/debugmap.go
  - internal/compiler/check/check.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/callgraph/callgraph.go
  - internal/compiler/interp/interp.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_admit.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_entry.go
  - internal/compiler/session/session_modules.go
  - cmd/schway-repair/repair.go
  - examples/phase30/adler32.schway
  - examples/phase30/adler32_cli.schway
  - examples/phase30/README.md
findings:
  critical: 1
  warning: 5
  info: 2
  total: 8
status: issues_found
---

# Phase 30: Code Review Report

**Reviewed:** 2026-10-12
**Depth:** standard
**Files Reviewed:** 28 production files (test files spot-checked for seam leaks and vacuous controls)
**Status:** issues_found

## Summary

The resolver, loader, merge, visibility law, C-name mangling and unit attribution are mostly careful. The confinement pipeline checks each path component, requires an exact-case match, checks regular-file status before opening, and compares file identity after opening. One real defect is user-visible, and I reproduced it with the built binary. The loader's three cap refusals (module count, import depth, total bytes) never reach the user. The session layer discards them and reports a misleading `module.imports_not_loaded` instead. The other findings are diagnostic-quality and attribution defects. Each of those still fails closed.

## Critical Issues

### CR-01: Cap diagnostics (count, depth, total bytes) are swallowed; users get `module.imports_not_loaded`

**File:** `internal/compiler/modload/load.go:228-230`, with consumers `internal/compiler/session/session_modules.go:36-40` and `internal/compiler/session/session.go:1851-1853`
**Issue:** When `capDiags` is non-empty, `loadGraph` returns `l.result(ast.Program{})`, so `Set.Program.Imports` is empty. Both consumers use `len(set.Program.Imports) == 0` to mean "single-file program, take the unchanged path":
- `checkSourcePath` ignores the loader diagnostics and calls `Check(entry)`.
- `DebugMapCommandFile` takes its no-import branch and re-parses the entry.

That re-parse sees the imports with no Units, so `check.Program` emits `module.imports_not_loaded` ("imports are resolved only by the module-aware commands..."). That message is wrong, since the command is module-aware. The D-13 codes `module.count_limit`, `module.depth_limit` and `module.total_bytes_limit` are unreachable from every CLI route.

Reproduced:
- An entry with 15 imports (16 modules including the entry) passes `schway check`.
- An entry with 16 imports (17 modules) reports `module.imports_not_loaded [11:32]`, not `module.count_limit`.

The modload-level tests pass because they inspect `Load`'s diagnostics directly. No session-level test covers the cap codes.
**Fix:**
```go
// load.go
if len(l.capDiags) > 0 {
    return l.result(entry.program), l.capDiags // keep Imports so callers take the module path
}
```
In addition, make `checkSourcePath` and `DebugMapCommandFile` stop deciding on `len(set.Program.Imports) == 0` alone. They should treat any non-empty loader diagnostics for an import-bearing entry as the result. Add a session-level test for the count, depth and total-bytes codes.

## Warnings

### WR-01: Entry self-import (and any alias-to-entry) inflates the module count before it is known to be the entry

**File:** `internal/compiler/modload/load.go:291-305` and `319-325`
**Issue:** `enqueueImports` increments `l.count` and runs the cap check when it enqueues. Only `process` later discovers that the target is the entry file (`isEntry`) and does `l.count--`. An entry that imports its own file by module name, plus 15 other modules (16 distinct modules in total), is refused with a false count-limit diagnostic. The same inflated count also affects depth bookkeeping for the alias node.
**Fix:** Do not count an enqueue until it is known to be a distinct module. Either resolve entry identity at enqueue time, or perform the cap check at `process` time with the real count. Alternatively, treat `module == entry header` or the entry's path-derived name as an alias up front.

### WR-02: Repeated cap diagnostics for the same refused target

**File:** `internal/compiler/modload/load.go:291-302`
**Issue:** On a cap failure the code `continue`s without setting `l.visited[target]`. Every later import declaration of that target, in any module, emits another identical `count_limit` or `depth_limit` diagnostic. A diamond over the limit therefore produces N duplicate errors, which is noisy and also eats the primary-diagnostic budget.
**Fix:** Record the target as refused (for example `l.visited[target] = true` plus a `refused` set) so each module is reported once.

### WR-03: An imported name that is exported but not declared passes `check`, leaving a dangling import fact

**File:** `internal/compiler/check/check.go` (`resolveImports`, around lines 2331-2380, and `buildScopedCalleeContracts`, `callee, ok := byModule[...]; if !ok { continue }`)
**Issue:** `resolveImports` verifies that the target's export list names the item. It never verifies that the target module declares a function of that name. Reproduced with `lib` exporting `fn ghost` and no `ghost` declaration:
- `import lib { fn ghost }` passes `schway check`, and `core.Program.Imports` gains a fact for a function that does not exist.
- Calling `ghost(x)` yields `core.call_callee_unresolved`, a downstream core-level code. The user is not told that the import is the problem.

Corevalidate trusts the `Imports` facts, so a dangling fact is a wrong fact in the core output.
**Fix:** In `resolveImports`, after the exported check, require the function to exist for `target.Name`. Otherwise emit a `module.not_exported` or `module.unresolved_import` diagnostic with an `export_list` cause. Do not emit a `ModuleImport` fact for it.

### WR-04: Import name that collides with a foreign symbol of the entry is not refused at the import

**File:** `internal/compiler/check/check.go` (`resolveImports` collision block, around lines 2395-2420)
**Issue:** The local-collision check only consults `localSpan`, which holds functions. In the entry module, foreign symbols are a separate namespace that call sites resolve first. Reproduced by importing a library function named `schway_fixture_bytes_view` into the example CLI: the check fails with `core.call_callee_unresolved [1047:1054]` and no `module.import_collision`. This fails closed but gives the wrong code and no cause.
**Fix:** Add the entry's foreign symbol names (and data constructor names) to the collision set for unit `""`, and emit `module.import_collision` with a `local_declaration` cause.

### WR-05: Peer-refusal unit is taken from the first `s1:` ID in the detail, which is the callee for the visibility codes

**File:** `internal/compiler/session/session_modules.go:88-104` (`unitOfDetail`), `internal/compiler/session/session.go:858-862`, `internal/compiler/corevalidate/corevalidate.go` (`v.check(allowed, code, callee.ID)`)
**Issue:** For `core.callee_not_exported` and `core.callee_not_imported` the detail passed to `check` is `callee.ID`. `unitOfDetail` then attributes the diagnostic to the callee's module. The offending code is in the caller. For example, an entry that calls a private library function is blamed on the library module, which also steers a unit-aware tool (repair, editor) to the wrong file. This is reachable only for forged or bypassed core, but the D-11/D-23 attribution contract is exactly about this. Parsing English detail text for an ID is also fragile.
**Fix:** Have `Problem` carry a structured `FunctionID` (the caller) for the visibility codes, and attribute from that. Do not re-parse the detail string.

### WR-06: Engine-level error classification: all non-NotExist `Lstat`/`Open` failures are reported as "resolves outside the entry directory"

**File:** `internal/compiler/modload/resolve.go:332, 360, 378, 409`
**Issue:** Any `Lstat` or `Open` error other than `ErrNotExist` becomes `module.escape`, including EACCES, ENAMETOOLONG, EMFILE and EIO. The message then says the module "resolves outside the entry directory", which is false and misleads the user about the cause. The error text is deliberately withheld, so the user cannot recover the real cause.
**Fix:** Map only root-escape and symlink-escape errors to `CodeEscape`, and map the rest to `CodeReadFailed` ("could not be read"). Go 1.24 `os.Root` reports escapes with an `errors.Is`-detectable "path escapes from parent" error. Match on that if it is exposed. Otherwise default to `CodeReadFailed`.

## Info

### IN-01: Test-only seams are read on production hot paths as unsynchronized globals

**File:** `internal/compiler/syntax/lexer.go:138` (`admitModulePathFormForTest` read for every identifier), `internal/compiler/syntax/parser.go` (`exceedsOffByOneForTest`), plus the many `*ForTest` booleans in modload, check, cgen and diagnostic
**Issue:** Production lexing and parsing read mutable package globals that tests flip. This is a data race with any `t.Parallel()` test, or with a future concurrent compile. It also leaves test-only branches in the lexer's identifier loop (`/` joining) that the production grammar must never reach.
**Fix:** Keep the seams, but read each one once per `Parse` into the `parser`/lexer struct, or move them behind build-tagged files, so the hot loop does not touch globals.

### IN-02: FIFO control test can leak a blocked goroutine and its seam on failure

**File:** `internal/compiler/modload/modload_phase30_gate_unix_test.go:44-64`
**Issue:** The control sets `openBeforeLstatForTest = true` and resets it manually at the end and in one failure branch, not with `defer`. If `os.OpenFile(pipe, O_WRONLY, 0)` fails, the goroutine blocked in `Open` is never released and stays stuck for the rest of the test binary. The 5 s wait then reports an error, but the leaked goroutine still holds the pipe. The seam can also stay armed if `t.Error` is followed by a panic.
**Fix:** Use `defer func(){ openBeforeLstatForTest = false }()` immediately after setting it. Open the write end with `O_RDWR` (non-blocking on Linux and macOS) so a release always succeeds, and register `t.Cleanup` to close it.

---

_Reviewed: 2026-10-12_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

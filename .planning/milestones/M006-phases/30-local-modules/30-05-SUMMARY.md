---
phase: 30-local-modules
plan: 05
subsystem: compiler-check
tags: [go, check, modules, scopes, import-resolution, body-blind, diagnostics]
requires:
  - phase: 30-local-modules
    provides: 30-00 gate PROCEED and no-import golden; 30-01 ast.Units/ImportItem/FuncDecl.Module; 30-02 Diagnostic.InUnit and Cause.Unit
provides:
  - core.Function.Module, core.Function.Exported, core.Program.Imports, core.ModuleImport
  - check per-unit callee-contract scopes (buildScopedCalleeContracts, signature-only)
  - module.imports_not_loaded, module.not_exported, module.import_collision, module.main_not_importable, module.type_import_unsupported
  - per-function Unit stamping of every check diagnostic (D-11)
  - SetSkipUnitStampingForTest and SetIgnoreUnloadedImportsForTest for the 30-08 session tests
affects: [30-04, 30-06, 30-08]
actuals:
  tokens: 28000
  tasks: 3
  commits: 3
plan_head_before: fcad2d105feadb34139359ed677631969a4fa52b
plan_head_after: 83eacc7bcc49f92ef580588b90d50c8961bf9f8e
commits: 3
tech-stack:
  added: []
  patterns:
    - "one scope per unit, built from declared signatures only; the contract type has no body field"
    - "unexported ...ForTest seams, each with an armed control that is observed to break the claim"
    - "stamp-once-per-function after the function loop instead of per emit site"
key-files:
  created:
    - internal/compiler/check/check_phase30_test.go
    - internal/compiler/check/check_phase30_bodyblind_test.go
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
key-decisions:
  - "Function IDs, parameter IDs and match IDs derive from idModule(program, function): the defining module, or the entry header for the entry. Entry IDs are therefore unchanged."
  - "A library function resolves foreign symbols against an empty table (D-20), and isEntryMain requires Module == \"\" in all six recognizer main tests."
  - "Interprocedural loan liveness and the call-graph cycle diagnostic are stamped from core.Function.Module (resp. the closing operation's owning function), verifyCall* from the function that owns the offending operation."
requirements-completed: [MOD-04, MOD-05]
coverage:
  - id: D1
    description: "An entry calling an imported exported function checks clean into module-qualified IDs, Function.Module/Exported and Program.Imports"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30CheckMergedProgram"
        status: pass
    human_judgment: false
  - id: D2
    description: "Import-bearing programs with no Units fail closed with module.imports_not_loaded"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30CheckRefusesUnloadedImports"
        status: pass
    human_judgment: false
  - id: D3
    description: "not_exported, import_collision (local, import, builtin), main_not_importable and type_import_unsupported are refused by name with no repair"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30PrivateNameRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30ImportCollisions"
        status: pass
    human_judgment: false
  - id: D4
    description: "Per-function Unit stamping, order-irrelevant imports, and a library-private main that is never entry-shaped"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30UnitStampedPerFunction"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30ImportItemOrderIrrelevant"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_phase30_test.go#TestPhase30LibraryPrivateMainNotEntryShaped"
        status: pass
    human_judgment: false
  - id: D5
    description: "Cross-module admission reads signatures only: a body change with the same signature moves nothing, a Callable flip legitimately flips the caller"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase30_bodyblind_test.go#TestPhase30CallAdmissionBodyBlind"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_phase30_bodyblind_test.go#TestPhase30CallableFlipIsASignatureChange"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_phase30_bodyblind_test.go#TestPhase30ScopedContractReadsNoBody"
        status: pass
    human_judgment: false
duration: 20min
completed: 2026-10-12
status: complete
---

# Phase 30 Plan 05: Module-aware check Summary

**check now derives one signature-only callee scope per unit, refuses every import misuse by name with both origins and no repair, stamps each diagnostic with its function's unit, and publishes `Function.Module`/`Exported` and `Program.Imports` for corevalidate, with every no-import byte unchanged.**

## Performance

- **Duration:** about 20 min
- **Started:** 2026-10-12T15:08Z
- **Completed:** 2026-10-12T15:28Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments
- core gains three additive omitempty facts (`Function.Module`, `Function.Exported`, `Program.Imports` with `ModuleImport`); no `json:"-"` was added.
- `buildScopedCalleeContracts` returns a `moduleScope` per unit (own functions plus exported fn imports). `buildCalleeContracts` delegates to its entry scope, so there is one code path and a no-`Units` program gets today's table.
- `resolveImports` runs before the reserved-builtin loop and the function loop. It refuses `module.type_import_unsupported`, `module.main_not_importable`, builtin collisions, `module.imports_not_loaded` (unloaded target), `module.not_exported` (hint cause with `Unit` = target, `export_list`), and local/import collisions. All use `diagnostic.Error(...).InUnit(importer)`, so none carries a repair.
- Callee resolution, `functionNames`, the foreign table and the four entry-shape recognizers now read the calling function's unit scope. A library function sees an empty foreign table (D-20). `isEntryMain` (Module == "") replaces every bare `Name == "main"` test.
- Every ID derived inside an imported function uses its defining module (`idModule`). Diagnostics are stamped once per function after the loop; interprocedural liveness, call-graph cycle, `verifyCallInvariants` and `verifyCallableRefusal` stamp from the owning function's module.
- SC2: the importer's diagnostics, scoped contracts and core JSON are byte-identical across an equal-signature library pair with zero `calleeBodyReadObserved` events; both body-read seams are observed to break it, and a go/ast scan pins the builder.

## Task Commits

1. **Task 1: tracer, module-scoped check and core module facts** - `acc0f68` (feat)
2. **Task 2: import-resolution refusals, unit stamping, library-private main** - production code in `acc0f68`, test `c341d4f` (test)
3. **Task 3: body-blind cross-module admission guard** - `83eacc7` (test, plus the seam block restructure in check.go)

**Plan metadata:** recorded in the following docs commit.

## Files Created/Modified
- `internal/compiler/core/core.go` - `Function.Module`, `Function.Exported`, `Program.Imports`, `ModuleImport`.
- `internal/compiler/check/check.go` - scopes, `resolveImports`, guard, stamping, `isEntryMain`, seams and exported setters.
- `internal/compiler/check/check_phase30_test.go` - `mergeForTest` and the tracer, refusal, stamping, ordering and library-main tests.
- `internal/compiler/check/check_phase30_bodyblind_test.go` - the three body-blind tests.

## Decisions Made
- Callable-flip branch that held: a flip exists. `relay(buffer: Buffer) -> Buffer` keeps the same declared signature while a body change (returning a borrowed view with no declared origin, the `call_uncallable_callee` shape) flips Callable to false. The importer then gets `core.callee_not_callable`, attributed to the entry unit. The pre-body `calleeContract` of the two variants is equal, because Callable is only read from the post-body signature table.
- `TestPhase30LibraryPrivateMainNotEntryShaped` control result: `programWideRecognizerScopeForTest` changed the **entry result**, not the entry-unit diagnostics. Under the control the set of entry-shaped core functions grows from the entry `main` alone (1) to include `s1:lib:fn:main` (2), and the lib-unit `check.local_owner_shape_unsupported` refusal disappears. Entry-unit (Unit "") diagnostics are empty in both, because a library `main` never reaches the entry's own checks. The plan's "entry-unit diagnostics or entry result" disjunction is satisfied by the entry result.
- In the default path a PathToken-shaped library `main` is refused with the general `check.local_owner_shape_unsupported` (any non-entry PathToken function that is not the acquire helper gets it, plus the library's empty foreign table). That is not a main-specific rule, so D-29 holds: no library-`main` code exists and `phase30LibraryMainCode` is absent. An ordinary library `main` (`Byte -> Byte`) checks clean with its own ID.
- Test callee types are `Byte`, not `U64`: scalar (`U64`) functions cannot appear as call targets in the linear call path, so `twice`/`mix` use `Byte -> Byte` as `call_basic.schway` does.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Task 2 production code landed in the Task 1 commit**
- **Found during:** Task 1
- **Issue:** `resolveImports`, the stamping loop, `isEntryMain` and the seams all live in `check.Program`'s single function loop and share one set of edits with the scopes. Splitting them into separate compilable commits would have meant duplicate throwaway edits.
- **Fix:** Committed the shared `check.go`/`core.go` edits with Task 1, along with the Task 1 and Task 2 refusal tests that passed against them. The Task 2 commit adds the library-main test. The Task 2 RED step was therefore not demonstrated as a separate failing commit; each behavior row is instead guarded by an armed control (`admitPrivateImportForTest`, `skipImportCollisionForTest`, `skipUnitStampingForTest`, `programWideRecognizerScopeForTest`) observed to change the result.
- **Files modified:** internal/compiler/check/check.go, internal/compiler/core/core.go
- **Verification:** full check/core packages and the session meta set pass.
- **Committed in:** acc0f68

**2. [Rule 1 - Bug] Seam guard shape for the go/ast scan**
- **Found during:** Task 3
- **Issue:** the first form `if scopedContractBodyReadSeam && callee.Body.Linear != nil` put body selectors in the guard's condition, which a scan keyed on a bare `if scopedContractBodyReadSeam` cannot distinguish from an unguarded read.
- **Fix:** nested the body test inside `if scopedContractBodyReadSeam { ... }`.
- **Files modified:** internal/compiler/check/check.go
- **Committed in:** 83eacc7

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)
**Impact on plan:** None on behavior. The only change is commit granularity for Task 2's production code.

## Issues Encountered
- A stray `cat` with no heredoc waited on stdin during a diagnostic probe and was killed; no repository effect.

## User Setup Required
None - no external service configuration required.

## Known Stubs
None.

## Next Phase Readiness
- 30-04 (modload graph/merge) can hand `check.Program` the merged shape this plan's `mergeForTest` builds. `Units[0]` must be the entry with Name "", and each `Unit.Imports` must carry the decl `Module` equal to the target `Unit.Name`.
- 30-06 corevalidate can re-derive visibility from `Function.Module`, `Function.Exported` and `Program.Imports`.
- 30-08 can arm `check.SetSkipUnitStampingForTest` and `check.SetIgnoreUnloadedImportsForTest` from session tests.

## Self-Check: PASSED

- core.go, check.go, check_phase30_test.go and check_phase30_bodyblind_test.go exist.
- Commits acc0f68, c341d4f and 83eacc7 are ancestors of HEAD.
- `go test ./internal/compiler/core ./internal/compiler/check` and the session meta set (Inventory, Groundedness, ValidationLifecycle, TestPhase30PreChangeGolden, 272/273 identity) pass.

---
*Phase: 30-local-modules*
*Completed: 2026-10-12*

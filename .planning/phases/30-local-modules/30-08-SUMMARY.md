---
phase: 30-local-modules
plan: 08
subsystem: session-routes
tags: [go, modules, session, modload, attribution, inventory, exit-codes]
requires:
  - phase: 30-local-modules
    provides: 30-00 no-import golden; 30-04 modload.Load and test setters; 30-05 check.Program merged-program check; 30-06 module-aware corevalidate, callgraph, interp, debugmap; 30-07 module-qualified C names
provides:
  - session_modules.go: loadSourcePath (the one modload.Load call), checkSourcePath, checkLoaded, unitOfFunctionID
  - check, interpreter run, native run, build, app verify and debug-map read their source through the loader (D-22)
  - CheckResult.UnitTrees; peer refusals carry Unit and locate spans in the owning unit's tree (D-11, D-23)
  - TestPhase30SingleIOSite AST scan pinning modload as the only module IO site (D-10)
affects: [30-09]
actuals:
  tokens: 22800
  tasks: 3
  commits: 4
plan_head_before: 4cda1fec517bc558d1350a17ed24dd7030d6128a
plan_head_after: 75c1a71d950f151075f90e5a1ed2158879873b1e
commits: 4
tech-stack:
  added: []
  patterns:
    - "byte-taking public API kept; file routes call checkSourcePath and an unexported continuation over a CheckResult"
    - "import-free programs take the unchanged Check(entry bytes) path, so no-import output is byte-identical (D-15)"
    - "AST scan of IO call sites keyed by file:function with three negative controls"
key-files:
  created:
    - internal/compiler/session/session_modules.go
    - internal/compiler/session/session_phase30_modules_test.go
    - internal/compiler/session/session_phase30_inventory_test.go
    - internal/compiler/session/session_phase30_export_test.go
    - cmd/schway/main_phase30_exit_test.go
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_app_verify.go
    - internal/compiler/session/session_admit.go
    - internal/compiler/session/session_entry.go
    - internal/compiler/session/session_phase30_gate_test.go
    - internal/compiler/session/session_test.go
    - internal/compiler/modload/modload_phase30_gate_test.go
    - internal/compiler/modload/modload_phase30_gate_unix_test.go
    - internal/compiler/modload/resolve_test.go
    - internal/compiler/modload/resolve_unix_test.go
    - internal/compiler/modload/resolve_other_test.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "buildApplicationChecked takes the ENTRY bytes, so the receipt SourceDigest stays the entry digest; module hashing is not claimed (Phase 31)."
  - "The byte-taking peerRefusalDiagnostic and originPeerRefusalDiagnostic signatures are kept (existing tests call them); the unit comes from the module segment of the ID in the problem, with the entry module read from the program's own module ID."
  - "VerifyApplicationCases gained an unexported twin taking an optional precomputed CheckResult instead of a second public signature."
  - "env:symlink, env:mkfifo and env:case-insensitive-fs join the closed environmental set: they are host capabilities, not debt."
requirements-completed: [MOD-03, MOD-04, MOD-05]
duration: 26min
completed: 2026-10-12
status: complete
coverage:
  - id: D1
    description: "check, interpreter run, native run, build, app verify and debug-map load a two-module program through modload and run it end to end (42 from 21 through the built app)"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30CheckFileThroughLoader"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30ModuleRoutes"
        status: pass
    human_judgment: false
  - id: D2
    description: "An error in an imported file reaches the CLI result with unit=<module> and a span into that file's own bytes (type, parse and peer-origin refusals), with armed controls"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30ImportedFileDiagnosticNamesFile"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30PeerRefusalNamesUnit"
        status: pass
    human_judgment: false
  - id: D3
    description: "Diamond loaded once and emitted once; reversing import order is byte-identical; two private helpers keep two core IDs, two C definitions, two debug-map units, and each of the four consumers is observed to attribute a lib-only fault to s1:lib:fn:helper"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30DiamondLoadsAndEmitsOnce"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30ReversedImportOrderIdenticalC"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30SameBareNameTwoModules"
        status: pass
    human_judgment: false
  - id: D4
    description: "Single-file commands refuse imports by name, module work is summed under MaxCheckWork, the no-import debug-map path stays unguarded, refusals exit 2 and operational faults exit 3"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30SingleFileCommandsRefuseImports"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30ModuleWorkLimit"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30DebugMapNoImportUnguarded"
        status: pass
      - kind: unit
        ref: "cmd/schway/main_phase30_exit_test.go#TestPhase30ModuleRefusalExitCodes"
        status: pass
    human_judgment: false
  - id: D5
    description: "modload is provably the only module IO site and route results never leak host paths"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_inventory_test.go#TestPhase30SingleIOSite"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_modules_test.go#TestPhase30ErrorsNeverLeakAbsolutePaths"
        status: pass
    human_judgment: false
---

# Phase 30 Plan 08: Module-aware session routes Summary

**check, run, build, app verify and debug-map now load module programs through one shared helper (loadSourcePath -> checkSourcePath -> modload.Load), diagnostics and peer refusals name the owning module and point into its own file, and an AST scan pins modload as the only module IO site.**

## Performance

- **Duration:** 26 min
- **Started:** 2026-10-12T15:40:00Z (approximate)
- **Completed:** 2026-10-12T16:06:00Z
- **Tasks:** 3 (plus one guard-fix commit)
- **Files modified:** 17

## Accomplishments

- Every no-import program keeps the byte path: `checkSourcePath` hands the entry bytes to the unchanged `Check`. `TestPhase30PreChangeGolden`, the 272/273 identity tests and the Phase 28 exit manifest stay green; the whole `session` and `cmd/schway` packages pass.
- Unit attribution reaches the CLI: JSON `"unit":"lib"` and the human ` unit=lib [` form for type, parse and peer-origin refusals. Peer refusals derive the unit from the ID in the problem and locate spans in `UnitTrees[module]`, never the entry tree.
- Diamond, reversed-order and same-bare-name invariants hold end to end, each with an armed control observed to break it.
- The one `modload.Load` call in `session` is `loadSourcePath`; `TestPhase30SingleIOSite` classifies every `readBoundedFile`, `os.Open`, `os.OpenFile`, `os.ReadFile`, `os.Stat` and `os.OpenRoot` site, with three negative controls.

## Task Commits

1. **Task 1: tracer, check and interpreter routes load modules** - `8f0dacb` (feat)
2. **Task 2: remaining routes, attribution, diamond, order, same-bare-name, refusals, exit codes** - `9c42e59` (feat)
3. **Task 3: single IO site scan, path hygiene, emitter inventory re-pin** - `e9a5373` (test)
4. **Guard fix (Rule 3)** - `75c1a71` (test)

**Plan metadata:** the docs commit that follows this file.

## Same-bare-name census record

All four rows ran, none skipped. The faithful program is admitted or ordered with both helper IDs present, and the attribution assertion requires `s1:lib:fn:helper` and forbids `s1:app.main:fn:helper`.

| Consumer | Fault (lib helper only) | Observed refusal | Branch |
| --- | --- | --- | --- |
| corevalidate | `return_type` U64 -> Bool | `core.return_type_mismatch`, Detail `s1:lib:fn:helper:op:3` (an operation ID; the plan guessed a place ID, the prefix rule still holds) | containment |
| originvalidate | remainder op `type_id` -> Bool type fact | `core.scalar_operation_invalid`, Detail `s1:lib:fn:helper: invalid remainder_u64 operation ...` | containment |
| pathoracle | foreign contract on the return op | `pathoracle.local_owner_shape: function "s1:lib:fn:helper" is outside the admitted PathToken-to-U64 path` | containment |
| callgraph | self-call op in lib helper | `callgraph.cycle: ... length 1: [s1:lib:fn:helper]` | diff / witness (faithful Order lists both helpers) |

## Decisions Made

See `key-decisions` above. Additionally: the same-bare-name fixture is `main -> entry helper -> lib run -> lib helper (n % 7)`. The linear-body grammar allows one call per function, so the entry helper only forwards; n=20 yields 6 and the interpreter trace names both helper IDs.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Interpreter and native run lanes only drive Byte, Buffer or data entries**
- **Found during:** Task 1 and Task 2 fixtures
- **Issue:** the plan asked `RunInterpreterFile` to return 42 on input 21, but `interpreterInputs` supports no U64 entry (this holds for every program, not only modules), and `run --engine=native` fails the same way.
- **Fix:** U64 results are proven by `CheckFile` then `callgraph.EntryFunction`, `PruneToReachable` and `interp.RunApplication`, plus the built app (`app run -- 21` prints 42). The run routes are exercised on a Byte two-module fixture (`RunInterpreterFile`, `RunInterpreterCommandFile`, `RunNativeCommandFile`).
- **Files modified:** `session_phase30_modules_test.go`
- **Committed in:** 8f0dacb, 9c42e59

**2. [Rule 3 - Blocking] Whole-package session run failed `TestNoSuppressionOutlivesItsWitness`**
- **Found during:** whole-package run before the SUMMARY
- **Issue:** 10 uncited skips and `unix` build constraints from the 30-03/30-04 modload tests, plus one new uncited skip in this plan's symlink row.
- **Fix:** the session symlink row now fails instead of skipping on a unix host; modload skips cite `env:symlink`, `env:mkfifo`, `env:case-insensitive-fs`, added to the closed `debtRegisterEnvironmentalSet` (host capabilities, not debt); the modload unix-only tests use `darwin || linux` (their `!unix` twin became `!darwin && !linux`).
- **Files modified:** `session_test.go`, `modload_phase30_gate_test.go`, `modload_phase30_gate_unix_test.go`, `resolve_test.go`, `resolve_unix_test.go`, `resolve_other_test.go`, `session_phase30_modules_test.go`
- **Verification:** `go test ./internal/compiler/session -count=1` and `go test ./internal/compiler/modload -count=1` pass.
- **Committed in:** 75c1a71

**3. [Plan reading] explain and query are pinned by the fail-closed diagnostic's ID**
- **Issue:** these routes look a diagnostic up by ID in the single-file check. On an import-bearing source the diagnostic they can find is `module.imports_not_loaded`; any other ID gives `explain.diagnostic_not_found` or an `unrecognized` query fact.
- **Fix:** the tests take the ID from `session.Check(entry bytes)` and assert both routes name the refusal. Changing the not-found wording is outside this plan's files and is left as a known limitation (T-30-30 residual): an unknown ID on an import-bearing source is not specially reported.

**Total deviations:** 3 (2 Rule 3, 1 plan reading). **Impact:** no scope creep; behavior of import-free programs is untouched.

## Issues Encountered

- `TestPhase16PublicEmitterConsumerInventory` needed one re-pin as the D-27 amendment allows: `session.go:EmitNative` 3049 -> 3089. Only the line number changed.

## Known Stubs

None.

## Threat Flags

None. No new endpoint, auth path, file access pattern or schema; all module IO stays behind `modload`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 30-09 can consume the module-aware routes. The TempDir fixtures here are generated, so `examples/phase30/` stays the only committed fixture directory.
- Not covered: hosted `go test -race` over `modload` and `session` (the MOD-05 concurrency backstop) runs in CI, not here.

## Self-Check: PASSED

- `session_modules.go`, `session_phase30_modules_test.go`, `session_phase30_inventory_test.go`, `session_phase30_export_test.go`, `main_phase30_exit_test.go` exist.
- Commits `8f0dacb`, `9c42e59`, `e9a5373`, `75c1a71` are ancestors of HEAD.

---
*Phase: 30-local-modules*
*Completed: 2026-10-12*

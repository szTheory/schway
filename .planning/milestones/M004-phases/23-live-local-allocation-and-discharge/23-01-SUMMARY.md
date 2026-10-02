---
phase: 23-live-local-allocation-and-discharge
plan: 01
subsystem: compiler/native
tags: [path-token, ownership, local-c, clang, c17, ffi]

requires: []
provides:
  - Checked PathToken to FileByteOwner acquire, borrow, and consuming release path
  - Retained native app producing 65 and 66 from separate caller-selected files
  - Per-operation C prototype and target record layout checks
affects: [23-02, 23-03, 23-04, 23-06]

actuals:
  tokens: 17086
  tasks: 2
  commits: 4
  plan_head_before: f3fe7e725d070fbc0b332bf0b4762d90f741597f

tech-stack:
  added: []
  patterns:
    - Per-operation foreign ABI contracts carried into core operations
    - Header-level C17 layout assertions compiled by each binding probe

key-files:
  created:
    - examples/phase23/file_byte.lang
    - examples/phase23/adapter.h
    - examples/phase23/adapter.c
    - examples/phase23/file_byte.bindings.json
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/native/bindings.go
    - internal/compiler/session/session.go
    - cmd/lang/main_test.go
    - internal/compiler/native/native_app_test.go

key-decisions:
  - "Keep PathToken as direct bounded argv data and admit only the frozen straight-line owner shape."
  - "Check each C symbol against its own function type and prove record layout in the shared C17 header."

requirements-completed: [FFI-03, RES-04]

coverage:
  - id: D1
    description: "Separate caller-selected 0x41 and 0x42 files print 65 and 66 through the retained app, while the Phase 22 U64 app still prints 7."
    requirement: RES-04
    verification:
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase23PublicFileByte"
        status: pass
    human_judgment: false
  - id: D2
    description: "Generated native code keeps the acquired owner through borrowed use and invokes its matching release before return or use-failure exit."
    requirement: RES-04
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_app_test.go#TestPhase23GeneratedReleaseFollowsBorrowedUse"
        status: pass
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase23PublicFileByte"
        status: pass
    human_judgment: false
  - id: D3
    description: "Acquire, use, and release reject independent prototype mismatches; C17 probes reject an unsupported target record layout."
    requirement: FFI-03
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_app_test.go#TestPhase23OperationABI"
        status: pass
    human_judgment: false

duration: 25 min
completed: 2026-09-27
status: complete
---

# Phase 23 Plan 01: Live Local Allocation and Discharge Summary

**A retained native file-byte app now carries a checked PathToken into a live noncopyable owner, borrows its allocation, and emits a paired C release with per-operation ABI checks.**

## Performance

- **Duration:** 25 min (measured from the earliest retained task event)
- **Started:** 2026-09-27T22:19:42-04:00 (earliest retained execution event; work began earlier)
- **Completed:** 2026-09-27T22:44:36-04:00
- **Tasks:** 2
- **Files modified:** 16

## Accomplishments

- Added a bounded opaque `PathToken` entry and noncopyable `FileByteOwner` with separate checked acquire, borrowed-use, and consuming-release contracts. The checker, core validator, origin validator, path oracle, emitter, and retained build session admit the same narrow success shape.
- Added the explicit local C adapter and manifest. It opens a regular file, reads exactly one byte plus an EOF probe, returns initialized malloc-backed storage, uses that storage for the result, and frees it on generated release.
- Added C17 target layout assertions to the shared header and an exact function-type assertion to every independent native binding probe.
- Extended the public CLI test to show separate files containing `0x41` and `0x42` print `65` and `66`; the existing Phase 22 identity app still prints `7`.

## Task Commits

Each task was committed atomically using a RED test commit followed by its GREEN implementation commit:

1. **Task 1 RED:** `7b9ae3f` — `test(23-01): add public file-byte application test`
2. **Task 1 GREEN:** `d793fa0` — `feat(23-01): add checked local file-byte ownership`
3. **Task 2 RED:** `7916b42` — `test(23-01): cover second file byte and operation ABI mismatches`
4. **Task 2 GREEN:** `3bc4eb7` — `feat(23-01): pin operation ABIs and accept second byte`

## Files Created/Modified

- `examples/phase23/file_byte.lang` — checked source operations and retained app entry.
- `examples/phase23/adapter.h`, `adapter.c`, `file_byte.bindings.json` — C17 records, file adapter, and closed local binding manifest.
- `internal/compiler/core/core.go`, `check/check.go`, `corevalidate/corevalidate.go`, `originvalidate/originvalidate.go`, `pathoracle/pathoracle.go` — operation contract production and independent admission checks.
- `internal/compiler/cgen/cgen.go`, `cgen_program.go`, `session/session.go` — entry transport, narrow native emission, release ordering, and manifest matching.
- `internal/compiler/native/bindings.go` — exact symbol prototype checks in per-symbol Clang probes.
- `cmd/lang/main_test.go`, `internal/compiler/native/native_app_test.go` — public app, generated release order, ABI mismatch, and target layout evidence.

## Decisions Made

- Preserved the existing direct argv transport and admitted only `PathToken -> U64` with the fixed acquire/use/generated-release sequence.
- Put layout assertions in the shared header so each binding probe proves the actual target's owner and result record facts before native build proceeds.

## Deviations from Plan

The declared file inventory omitted two implementation seams: `internal/compiler/ability/ability.go` registers the new type capability, and `internal/compiler/cgen/cgen.go` serializes the admitted entry type. Both changes were required by the planned `PathToken`/`FileByteOwner` path and are included in the actual-files list above. GSD worktree cleanup surfaced these as out-of-declared-scope paths; no additional capability or behavior was introduced.

## Issues Encountered

Git's index lives outside the writable worktree boundary in the shared repository metadata. The sandbox approved normal `git add` and `git commit` operations; all four task commits ran with hooks enabled. The plan-head measurement was stored in `/tmp` because the worktree could not write a ledger into the shared Git metadata directory.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The public success witness and operation ABI checks are ready for the sibling Phase 23 plans. The full acquisition error taxonomy, discard refusal, acquisition-seeded mutation suite, and independent physical allocation observer remain assigned to their respective plans. `FFI-03` and `RES-04` are shared with unfinished sibling plans, so 0 of 2 requirement IDs were ready to mark complete in this worktree.

---
*Phase: 23-live-local-allocation-and-discharge*
*Completed: 2026-09-27*

## Self-Check: PASSED

- All four created example files exist.
- All four task commits are present in Git history.
- SUMMARY.md exists and records the measured plan base and four pre-metadata commits.

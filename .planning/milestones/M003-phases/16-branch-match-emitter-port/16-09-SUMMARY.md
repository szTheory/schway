---
phase: 16-branch-match-emitter-port
plan: 09
subsystem: native-emission
tags: [cgen, schema-2, public-api, cutover]
requires:
  - phase: 16-08
    provides: explicit approval for the one-way public-emitter cutover
provides:
  - sole schema-2 authority for public Emit and EmitNative
  - immutable record of the atomic legacy-public-dispatch removal
affects: [phase-16-consumer-migration, native-emission]
actuals:
  tokens: 1
  tasks: 1
  commits: 1
tech-stack:
  added: []
  patterns:
    - Atomic public-authority cutovers are verified from the named historical commit.
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_n1_convergence_test.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/export_test.go
    - internal/compiler/core/core_test.go
key-decisions:
  - "0607486 is the sole D-01 public-emitter cutover; later plans may migrate consumers but may not restore a legacy fallback."
patterns-established:
  - "Public Emit and EmitNative validate then delegate to emitProgram exactly once."
requirements-completed: [NAT-08, NAT-09]
coverage:
  - id: D1
    description: Public production emission has one schema-2 authority and no legacy public fallback.
    requirement: NAT-08
    verification:
      - kind: unit
        ref: go test ./internal/compiler/cgen ./internal/compiler/core -run 'TestN1|TestProgram|TestPublic|TestPreviousPhaseCoreBytesUnchanged' -count=1
        status: pass
    human_judgment: false
  - id: D2
    description: The atomic cutover is preserved as immutable history and M004 families remain refusal-only.
    requirement: NAT-09
    verification:
      - kind: other
        ref: git show --stat 0607486
        status: pass
    human_judgment: false
duration: recovery-closeout
completed: 2026-09-20
status: complete
---

# Phase 16 Plan 09: Sole Public Emitter Cutover Summary

**The public emitter now validates once and delegates only to schema-2 `emitProgram`; the legacy public-dispatch implementation was removed atomically in `0607486`.**

## Performance

- **Duration:** Recovery closeout
- **Completed:** 2026-09-20
- **Tasks:** 1
- **Files modified:** 0 during recovery

## Accomplishments

- Verified `0607486` as the named indivisible public dispatcher and legacy-deletion commit.
- Confirmed `Emit` and `EmitNative` each route through validated `emitProgram` with no production fallback.
- Passed the focused cgen/core regression gate.

## Task Commits

1. **Task 1: Preserve the completed sole-production dispatcher cutover** — `0607486` (`feat(16-09): cut over public program emission`)

## Files Created/Modified

- `internal/compiler/cgen/cgen.go` — public `Emit` and `EmitNative` route to `emitProgram`.
- `internal/compiler/cgen/cgen_n1_convergence_test.go` — cutover convergence evidence.
- `internal/compiler/cgen/cgen_program_test.go` — schema-2 program-emitter coverage.
- `internal/compiler/cgen/export_test.go` — public-emitter test exposure adjusted with the cutover.
- `internal/compiler/core/core_test.go` — core-level byte/control evidence retained.

## Decisions Made

- Treated the existing `0607486` as the plan's authorized atomic action. No second cutover and no source changes were made during recovery.

## Deviations from Plan

### Recovery closeout

The prior execution created the planned production commit but did not create this required summary. This recovery verified the existing historical commit and wrote the missing audit artifact; it did not repeat the source change.

**Total deviations:** 1 recovery closeout.

## Issues Encountered

- The safe-resume gate detected scoped production commits without a summary, so executor dispatch was halted before any duplicate work could occur.

## User Setup Required

None.

## Next Phase Readiness

Plan 16-10 can now migrate schema-2 comparison projections and inventory evidence without reopening the public-emitter authority decision.

---
*Phase: 16-branch-match-emitter-port*
*Completed: 2026-09-20*

---
phase: 19-numeric-literals-and-opconst
plan: 02
subsystem: interpreter
tags: [u64, scalar-projection, regression-gate]
requires:
  - phase: 19-numeric-literals-and-opconst
    provides: Refused literal fixtures and pinned Wave 1 diagnostic frontier
provides:
  - U64-capable interpreter value representation with canonical decimal projection
  - D-12-18 replay and mutation gate before OpConst routing
affects: [19-03, interpreter, execution-evidence]
actuals:
  tokens: 979
  tasks: 2
  commits: 2
commits: 2
plan_head_before: 2668359
tech-stack:
  added: []
  patterns: [discriminated scalar value representation, golden replay plus seeded mutation]
key-files:
  created:
    - internal/compiler/interp/interp_phase19_test.go
    - .planning/phases/19-numeric-literals-and-opconst/19-SCALAR-GATE.md
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/session/session_phase19_test.go
key-decisions:
  - "D-19-01: Represent U64 independently of host architecture, with its full unsigned range."
  - "D-19-04: Project U64 as canonical decimal while retaining existing scalar bytes."
requirements-completed: [VAL-01, VAL-03]
coverage:
  - id: D1
    description: U64 values project as canonical decimal while legacy scalar and tagged values retain their established projections.
    requirement: VAL-03
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/interp ./internal/compiler/session -run 'TestPhase19(ScalarProjection|LiteralFrontier)|TestPayloadCorpusCharacterizationReplay' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: The existing scalar replay detects a seeded projection fault and the literal tracer remains refused at the Wave 1 frontier.
    requirement: VAL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run 'TestPhase19ScalarGate|TestPayloadCorpusCharacterizationReplayMutationKilled|TestPayloadCorpusCharacterizationReplay' -count=1"
        status: pass
    human_judgment: false
duration: 4min
completed: 2026-09-24
status: complete
---

# Phase 19 Plan 02: U64 Scalar Projection Gate Summary

**Interpreter values now carry a distinct U64 fact, serialize it as canonical decimal, and retain the prior scalar projection proven by the D-12-18 replay.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-24T23:46:38Z
- **Completed:** 2026-09-24T23:50:38Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added a discriminated U64 representation and canonical decimal formatting for zero, 42, and the maximum U64 value.
- Verified existing untagged scalar and tagged projections remain byte-identical; the full pinned payload corpus replay passed.
- Reused the established empty-tag mutation seam and recorded the unchanged-golden result before any OpConst dispatch work.

## Task Commits

1. **Task 1: Widen interpreter values and replay the old scalar corpus** - `b28df87` (feat)
2. **Task 2: Prove the scalar replay still detects a projection fault** - `5c100f5` (test)

## Files Created/Modified

- `internal/compiler/interp/interp.go` - Adds the U64 value discriminator and decimal projection.
- `internal/compiler/interp/interp_phase19_test.go` - Covers U64, legacy scalar, and tagged projections.
- `internal/compiler/session/session_phase19_test.go` - Adds the Phase 19 gate reusing the existing replay mutation test and fixture refusal assertion.
- `.planning/phases/19-numeric-literals-and-opconst/19-SCALAR-GATE.md` - Records replay, mutation, and golden disposition.

## Decisions Made

- Kept the numeric value in an explicit U64 field, separate from the legacy tag and payload fields.
- Kept the existing `String()` empty-tag behavior and returned U64 values in canonical base-10 form.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The initial test-first compile failed because the planned U64 fields did not exist yet; implementation added the fields and the focused tests passed.
- The sandbox denied writes to `.git/index.lock`; the requested per-task commits were then created through the managed escalation. No unrelated files were staged.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 03 can begin with the interpreter value and the unchanged D-12-18 scalar evidence gate established. No `OpConst` dispatch consumer has been modified.

---
*Phase: 19-numeric-literals-and-opconst*
*Completed: 2026-09-24*

## Self-Check: PASSED

- Summary, scalar gate record, and test files exist.
- Task commits `b28df87` and `5c100f5` exist.
- Measured task commit count from `plan_head_before` `2668359` is 2.

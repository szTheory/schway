---
phase: 19-numeric-literals-and-opconst
plan: 01
subsystem: compiler-testing
tags: [numeric-literals, source-diagnostics, fixture-frontier]
requires: []
provides:
  - Three Phase 19 numeric source witnesses pinned to current production refusals
affects: [19-02, 19-03, 19-04]
actuals:
  tokens: 856
  tasks: 2
  commits: 2
commits: 2
plan_head_before: d167bcc248cf9df72653ae6bba0c80ffbbb0f7f5
tech-stack:
  added: []
  patterns: [fixture-backed refusal identity and source-span assertions]
key-files:
  created:
    - testdata/phase19/literal_tracer.lang
    - testdata/phase19/literal_overflow.lang
    - testdata/phase19/literal_malformed.lang
  modified:
    - internal/compiler/session/session_phase19_test.go
key-decisions: []
requirements-completed: [VAL-01]
coverage:
  - id: D1
    description: Direct numeric U64-return source witness retains its current diagnostic frontier.
    requirement: VAL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run 'TestPhase19LiteralFrontier' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Overflow and malformed numeric spellings have separate refused source witnesses.
    requirement: VAL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run 'TestPhase19(LiteralFrontier|NumericRefusalFrontiers)' -count=1"
        status: pass
    human_judgment: false
duration: 4min
completed: 2026-09-24
status: complete
---

# Phase 19 Plan 01: Numeric Literal Refusal Frontier Summary

**Three literal-bearing `.lang` witnesses now pin the compiler’s pre-implementation refusal identities and source spans.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-24T23:41:17Z
- **Completed:** 2026-09-24T23:45:10Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added a direct `let count = 42` U64-return tracer fixture and pinned its current first diagnostic ID, code, and span.
- Added separate above-U64 and malformed radix/separator fixtures; both retain their current production refusal identity and span.
- Kept this wave fixture/test-only so later plans can measure diagnostic movement before expanding compiler production support.

## Task Commits

1. **Task 1: Check in the first literal source witness and pin its current refusal** - `9ec6476` (test)
2. **Task 2: Pin overflow and malformed-literal source witnesses** - `b71c861` (test)

## Files Created/Modified

- `testdata/phase19/literal_tracer.lang` - Direct numeric binding with a U64 return signature.
- `testdata/phase19/literal_overflow.lang` - First decimal spelling above U64 maximum.
- `testdata/phase19/literal_malformed.lang` - Invalid separator placement in a hexadecimal spelling.
- `internal/compiler/session/session_phase19_test.go` - Pins current production refusal IDs and spans for each witness.

## Decisions Made

None - followed the plan as specified.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The default Go build cache under `~/Library/Caches` was not writable in the sandbox. Focused verification passed with `GOCACHE=/tmp/ai-lang-go-cache`.
- The task annotations say `tdd="true"`, but both tasks characterize refusals that already exist in production. Their assertions therefore pass against the pre-existing behavior; no meaningful feature RED/GREEN implementation cycle applies to this fixture-only plan.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 02 can begin with the valid-intent tracer and both negative witnesses committed before any production changes. The pinned diagnostics establish the current frontier for the later movement gate.

---
*Phase: 19-numeric-literals-and-opconst*
*Completed: 2026-09-24*

## Self-Check: PASSED

- All three fixture files and the test file exist.
- Task commits `9ec6476` and `b71c861` exist.
- Both task-specific verification commands passed.

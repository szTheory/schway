---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 01
subsystem: compiler
tags: [syntax, checker, typed-core, independent-validation, output]

# Dependency graph
requires: []
provides:
  - Explicit Unit/no-result scalar application return facts
  - Typed write_u64(U64) and write_newline() core operations
  - Independent core and origin admission of write and Unit return facts
affects: [27-02, 27-03, 27-04, 27-05, 27-06, application-output]

# Actuals
actuals:
  tokens: 8330
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Explicit return mode is recorded in the function signature, not inferred from write events.
    - Output events retain static operation identity and do not alter scalar fixed-point state.

key-files:
  created:
    - internal/compiler/session/session_phase27_test.go
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/core/core.go
    - internal/compiler/check/scalar.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go

key-decisions:
  - "Unit is an explicit scalar return mode; its return operation has no source place."
  - "Decimal and LF writes are separate typed operations; neither creates a result place."
  - "Write events preserve their static operation IDs while remaining state-neutral in scalar fixed-point analysis."

requirements-completed: [APP-08, APP-09]

coverage:
  - id: D1
    description: "Ordinary Unit source admits typed decimal and LF writes, rejects invalid operands with source spans, and preserves U64-returning apps."
    requirement: APP-08
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase27_test.go#TestPhase27WriteAdmission"
        status: pass
    human_judgment: false
  - id: D2
    description: "Core and origin peers independently accept valid write and Unit facts and reject forged operands and return mode."
    requirement: APP-09
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase27_test.go#TestPhase27WritePeerAdmission"
        status: pass
    human_judgment: false

# Metrics
duration: 40 min
started: 2026-10-03T16:12:50Z
completed: 2026-10-03
status: complete
---

# Phase 27 Plan 01: Explicit Unit Writes Summary

**Unit application entries now admit typed decimal and LF writes through source checking and independent core/origin validation.**

## Performance

- **Duration:** 40 min
- **Started:** 2026-10-03T16:12:50Z
- **Completed:** 2026-10-03T16:52:22Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Added an explicit `Unit` result mode and source statements for `write_u64(expr)` and `write_newline()`.
- Lowered each write to a distinct typed core operation and retained stable operation IDs across control-flow loops.
- Added independent core/origin admission checks and reached forged-operation controls; existing scalar carry refusals remain intact.

## Task Commits

Each task was committed atomically:

1. **Task 1: Check Unit source and precise decimal/LF writes** - `25c00cf` (feat)
2. **Task 2: Reject forged write and return-mode facts independently** - `d756e98` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `internal/compiler/session/session_phase27_test.go` - source admission, span errors, independent peer checks, and forged-operation controls.
- `internal/compiler/ast/ast.go`, `internal/compiler/syntax/parser.go` - write statement and Unit-body source representation.
- `internal/compiler/core/core.go`, `internal/compiler/check/scalar.go` - typed write operations and explicit Unit return lowering.
- `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/originvalidate/originvalidate.go` - independent write and return-mode validation.

## Decisions Made

- Unit is encoded as the declared function return type and a source-less Unit return operation.
- Output operations have no result place; `write_u64` carries a U64 source operand and `write_newline` carries none.
- Output events preserve their static operation identity and do not mutate scalar dataflow state.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The first peer mutation control targeted a nonexistent place name. It was corrected to select a place from the independently declared Bool type fact; the focused peer tests then passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 27-02 is ready to add formatting and the first executable FizzBuzz path on top of this admission boundary.
- Focused Phase 27 and Phase 26 scalar refusal checks passed; no full project suite was run.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*

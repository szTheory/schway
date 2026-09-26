---
phase: 16-branch-match-emitter-port
plan: "03"
subsystem: native-emitter
tags: [go, c17, match, payload-layout, schema-2]
requires:
  - phase: 16-02
    provides: program-function branch lowering and schema-2 event infrastructure
provides:
  - Checker-derived tagged payload records in emitProgram
  - Schema-2 payload event attribution and defect-before-abort terminal evidence
affects: [16-06-cutover, native-convergence]
tech-stack:
  added: []
  patterns: [checker-owned payload layout projection, direct defect terminal writer]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
key-decisions:
  - "emitProgram reads check.PayloadRecordLayout and core.AlternativeNameForPayloadType rather than deriving payload fields locally."
  - "The defect reason remains a direct terminal event because the shared buffered event ABI intentionally has no output field."
requirements-completed: [NAT-08]
actuals:
  tokens: 6120
  tasks: 2
  commits: 4
plan_head_before: 7c38df5bb413891a68f29ba0df4429bd243a8b7d
duration: 4min
completed: 2026-09-19
status: complete
coverage:
  - id: D1
    description: "Payload match selection and construction/destruction lower through the program emitter using checker-derived tagged fields."
    requirement: NAT-08
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/cgen -run 'TestProgramMatchPayloadLowering|TestProgramMatchUsesCheckerLayout|TestProgramBranchTracer' -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: "Defect arms write a schema-2 terminal event before abort, with _Noreturn restricted to lang_defect."
    requirement: NAT-08
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/cgen -run 'TestProgramMatchDefectEventPrecedesAbort|TestProgramNoreturnExemptionIsNarrow|TestProgramMatchPayloadLowering' -count=1 -v"
        status: pass
    human_judgment: false
---

# Phase 16 Plan 03: Branch/Match Emitter Port Summary

**The surviving program emitter now lowers checker-shaped tagged payload matches and streams schema-2 defect evidence before its sole abort-only helper.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-19T21:43:38Z
- **Completed:** 2026-09-19T21:47:36Z
- **Tasks:** 2/2
- **Files modified:** 2

## Accomplishments

- Projected tagged record fields and C types from `check.PayloadRecordLayout`, preserving alternative declaration order.
- Resolved payload operation identity through `core.AlternativeNameForPayloadType` and kept construct/destructure events attributed to schema-2 invocations.
- Emitted the defect document, including its reason-bearing `function.defected` event, before calling the narrowly `_Noreturn` `lang_defect` helper.

## Task Commits

1. **Task 1: Port tagged match selection and payload layout**
   - `f7d17dd` test: add failing payload program tests
   - `eec684c` feat: lower payload matches through emitProgram
2. **Task 2: Preserve the defect terminal and narrow `_Noreturn` exemption**
   - `79f36ba` test: add failing program defect tests
   - `b5d5f09` feat: preserve program match defect terminal

## Files Created/Modified

- `internal/compiler/cgen/cgen_program.go` — tagged-layout, payload-operation, and defect-terminal lowering in the program emitter.
- `internal/compiler/cgen/cgen_program_test.go` — native payload and defect behavior plus structural `_Noreturn` controls.

## Decisions Made

- Payload C field names and types are read from the checker-derived layout; `cgen` does not add a second layout law.
- The direct defect tail deliberately leaves the general event buffer unchanged so ordinary event output remains byte-stable and the defect event retains its required reason.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- `go test ./internal/compiler/cgen -count=1` remains red in `TestN1ConvergenceDifferential`: its expectations still require `emitProgram` to reject `toggle.lang` and `defect_terminal.lang`. The newly admitted behavior is intentional and the expectation flip belongs to the explicitly deferred Plan 16-06 cutover; it was not changed here.

## Next Phase Readiness

The direct program-emission path now covers payload and defect match arms. Plan 16-06 can own the required convergence expectation flip and dispatch cutover without altering the M004 by-pointer disposition.

## Self-Check: PASSED

- Confirmed both modified compiler files exist and all four task commits are present in git history.

---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 02
subsystem: compiler
tags: [interpreter, c17, formatter, application-output]

# Dependency graph
requires:
  - phase: 27-01
    provides: Typed decimal/LF write operations and explicit Unit return facts
provides:
  - Runnable Unit FizzBuzz tracer for interpreter and retained native application paths
  - Separate interpreter application bytes alongside the unchanged execution document
  - Exact empty, pre-output refusal, and legacy U64 automatic-print controls
affects: [27-03, application-output, phase27-evidence]

# Actuals
actuals:
  tokens: 6982
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Application stdout is captured separately from interpreter execution evidence.
    - Unit entries emit only explicit writes; declared U64 entries retain their automatic result line.

key-files:
  created:
    - examples/fizzbuzz.schway
  modified:
    - internal/compiler/syntax/format.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase27_test.go

key-decisions:
  - "The interpreter application route returns execution evidence and application bytes as separate fields."
  - "Native and interpreter Unit entries do not append a return value; U64 entries still append their decimal result and LF."
  - "Explicit decimal and LF output uses the existing bounded native writer."

requirements-completed: [APP-08, APP-09]

coverage:
  - id: D1
    description: "Ordinary Unit FizzBuzz source writes exactly 1 plus LF through the interpreter and retained C17/Clang application."
    requirement: APP-08
    verification:
      - kind: e2e
        ref: "internal/compiler/session/session_phase27_test.go#TestPhase27TracerFizzBuzz1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Input 0 stays empty, input 1001 fails before output, U64 return printing remains stable, and Unit/write formatting preserves comments and recovery spans."
    requirement: APP-09
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase27_test.go#TestPhase27TracerBoundaries"
        status: pass
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/syntax -run '^TestPhase27Tracer(Boundaries|FizzBuzz1)$|^TestPhase26(FrontendRecovery|ExactSumMatrix)$' -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 30 min
completed: 2026-10-03
status: complete
---

# Phase 27 Plan 02: Runnable Unit FizzBuzz Tracer Summary

**The Unit FizzBuzz tracer now writes exact application bytes through both the interpreter and retained native path, with U64 return printing preserved.**

## Performance

- **Duration:** 30 min
- **Started:** 2026-10-03T17:00:48Z
- **Completed:** 2026-10-03T17:29:54Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added `examples/fizzbuzz.schway` with the pre-output `n < 1001` guard and pre-tested scalar loop; input 1 writes exactly `1\n`.
- Added an interpreter application runner with a separate byte output field while keeping execution documents free of application output events.
- Extended the C17 application shell for Unit entries and explicit writes through the bounded writer, with no Unit return print.
- Pinned input 0, rejected input 1001, formatter stability/recovery, and the existing U64 automatic-print behavior.

## Task Commits

1. **Task 1: Run input 1 through both application engines** - `cd430c6` (feat)
2. **Task 2: Pin empty, rejected, formatter, and legacy return paths** - `a669602` (test)

**Plan metadata:** pending.

## Files Created/Modified

- `examples/fizzbuzz.schway` - Runnable Unit FizzBuzz tracer source.
- `internal/compiler/syntax/format.go` - Canonical line breaks for typed write statements.
- `internal/compiler/interp/interp.go` - Application result bytes separated from execution evidence.
- `internal/compiler/cgen/cgen_program.go` - Unit application shell and bounded explicit writes.
- `internal/compiler/session/session.go` - Checked source-to-interpreter application route and Unit app admission.
- `internal/compiler/session/session_phase27_test.go` - Exact engine output and boundary controls.

## Decisions Made

- Application bytes are returned separately from `Execution`; write operations do not change the execution document schema or add events.
- The declared return type selects application mode: Unit adds no automatic line, while U64 keeps its decimal result and LF after explicit writes.
- The native body and automatic return path share the existing bounded `schway_write_bytes` output accounting.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The initial U64-with-write probe used a straight-line body that the independent core validator rejects because its terminal CFG has no incoming edge. The control now uses the same guarded CFG shape as the tracer and still verifies that declared U64 return mode adds its automatic result line after explicit writes.

## User Setup Required

None - installed Clang was available and the native tracer selector compiled and ran the retained artifact.

## Next Phase Readiness

- Plan 27-03 can now add equality and remainder on top of a runnable application byte path.
- Focused Phase 27 and Phase 26 selectors passed. No full project suite was run.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*

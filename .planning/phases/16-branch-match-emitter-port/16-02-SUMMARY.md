---
phase: 16-branch-match-emitter-port
plan: 02
subsystem: native-emission
tags: [go, c17, cgen, branch, schema-2, validation-order]
requires:
  - phase: 16-branch-match-emitter-port
    provides: direct ordinary-linear emitProgram tracer and schema-2 resource tail
provides:
  - Branch-body lowering through the shared emitProgram function and event writer
  - Ordered branch-shape admission ahead of schema-2 preflight and serialization
affects: [16-03, 16-06, cgen, native-execution]
actuals:
  tokens: 7227
  tasks: 2
  commits: 4
commits: 4
plan_head_before: 9525b09157a09ed12b678e8bbf925089fbef395a
tech-stack:
  added: []
  patterns: [branch functions return through shared main serializer, preflight refusal observation reset per emission]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/cgen_n1_convergence_test.go
key-decisions:
  - "Admit nullary branch blocks through emitProgram while retaining explicit refusals for match-only, payload, foreign, and N=1 by-pointer shapes."
  - "Keep main as the sole schema-2 document writer; branch helpers only record events and return C values."
requirements-completed: [NAT-08]
coverage:
  - id: D1
    description: Branch arms execute through direct emitProgram with schema-2 events.
    requirement: NAT-08
    verification:
      - kind: integration
        ref: go test ./internal/compiler/cgen -run 'TestProgramBranchTracer|TestProgramWritesExecutionSchema2|TestNativeFunctionCalledPreorder' -count=1 -v
        status: pass
    human_judgment: false
  - id: D2
    description: Branch admission retains graph, shape, and preflight diagnostic ordering before serialization.
    requirement: NAT-08
    verification:
      - kind: unit
        ref: go test ./internal/compiler/cgen -run 'TestProgramBranchValidationOrder|TestUnsupportedProgramShapePrecedesSchema2Preflight|TestInvocationPreflightGuardIsNotInert' -count=1 -v
        status: pass
    human_judgment: false
duration: 8m 26s
completed: 2026-09-19
status: complete
---

# Phase 16 Plan 02: Branch Program Emitter Summary

**Nullary branch blocks now lower through the whole-program C writer, retaining schema-2 event identity and ordered admission failures.**

## Performance

- **Duration:** 8m 26s
- **Started:** 2026-09-19T21:28:22Z
- **Completed:** 2026-09-19T21:36:48Z
- **Tasks:** 2/2
- **Files modified:** 3

## Accomplishments

- Lowered checked branch blocks and arm-local copy/move/borrow/return operations with the existing program-level local allocator, invocation table, and schema-2 event buffer.
- Kept `main` as the single execution-document serializer; branch functions emit no document of their own.
- Pinned graph/entry, foreign, by-pointer, and seeded-output-preflight ordering so refusals cannot reach generated-C serialization.

## Task Commits

1. **Task 1: Port branch block lowering into the program function writer**
   - `fa43b95` test RED: add direct branch tracer
   - `0399ca5` feature GREEN: lower branch bodies in program emitter
2. **Task 2: Mutation-pin validation precedence for newly admitted branch bodies**
   - `d370f35` test RED: pin branch admission ordering
   - `44f714d` feature GREEN: preserve branch admission precedence

## Files Created/Modified

- `internal/compiler/cgen/cgen_program.go` — emits nullary branch functions through the shared whole-program schema-2 pipeline and rejects unadmitted shapes before preflight.
- `internal/compiler/cgen/cgen_program_test.go` — executes both branch outcomes and pins ordering/seed controls.
- `internal/compiler/cgen/cgen_n1_convergence_test.go` — updates the measured convergence characterization now that branch blocks are directly admitted.

## Decisions Made

- Scope direct branch admission to nullary alternatives and arm-local straight-line operations; payload, match-only, foreign, and N=1 by-pointer forms stay explicitly refused for later planned work.
- Reset the serialization observation at each emitter attempt so an earlier successful emission cannot mask a later ordered refusal.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Regression] Updated the existing N=1 convergence characterization.**
- **Found during:** Task 2
- **Issue:** Its branch fixture expected `emitProgram` to refuse; Task 1 intentionally admits that fixture.
- **Fix:** Recorded successful, non-byte-identical branch output as the new measured state.
- **Files modified:** `internal/compiler/cgen/cgen_n1_convergence_test.go`
- **Verification:** `go test ./internal/compiler/cgen -count=1`
- **Committed in:** `44f714d`

**Total deviations:** 1 auto-fixed (Rule 1).

## Issues Encountered

The generic TDD RED-evidence checker currently parses TAP-style records and does not classify Go test output; focused Go test output still established the intentional target-test RED failures before each GREEN commit.

## User Setup Required

None.

## Next Phase Readiness

The program emitter has a tested branch-block spine. Match-only and payload lowering remain deliberately named refusals, ready for their separately planned extensions.

## Self-Check: PASSED

- Verified all three modified compiler files exist.
- Verified task commits `fa43b95`, `0399ca5`, `d370f35`, and `44f714d` exist in git history.

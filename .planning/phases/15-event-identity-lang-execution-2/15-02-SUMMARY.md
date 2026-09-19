---
phase: 15-event-identity-lang-execution-2
plan: "02"
subsystem: interpreter evidence
tags: [interpreter, execution-schema-2, invocation, call-edges]
requires:
  - phase: 15-event-identity-lang-execution-2
    provides: canonical `/2` invocation formatter and execution model
provides:
  - Invocation-qualified interpreter events for multi-function executions
  - Preorder caller-owned `function.called` events
affects: [native event identity, execution validation, session differential]
actuals:
  tokens: 7068
  tasks: 2
  commits: 0
tech-stack:
  added: []
  patterns: [frame-local schema and invocation context, projection-only call-edge assertions]
key-files:
  created: [.planning/phases/15-event-identity-lang-execution-2/15-02-SUMMARY.md]
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_test.go
    - internal/compiler/interp/interp_oracle_golden_test.go
key-decisions:
  - "Select /2 only at the multi-function interpreter producer boundary; direct legacy test drivers remain /1."
  - "Emit a call edge only after resolution and depth admission, immediately before child evidence."
patterns-established:
  - "Frame-local occurrence context flows through every interpreter event constructor."
requirements-completed: [OBS-01, OBS-02, OBS-04]
coverage:
  - id: D1
    description: "Multi-function interpreter events carry canonical activation identity while single-function bytes remain /1."
    requirement: OBS-01
    verification:
      - kind: unit
        ref: "internal/compiler/interp#TestInvocationThreadsThroughAllEventPaths; TestExecutionSchemaSelectionPreservesLegacy"
        status: pass
    human_judgment: false
  - id: D2
    description: "Admitted calls produce one caller-owned preorder edge; rejected calls do not."
    requirement: OBS-02
    verification:
      - kind: unit
        ref: "internal/compiler/interp#TestFunctionCalledPreorderAndOwnership; TestFunctionCalledProjectionRemoval; TestRejectedCallEmitsNoCalledEdge"
        status: pass
    human_judgment: false
duration: 0min
completed: 2026-09-19
status: complete
---

# Phase 15 Plan 02 Summary

**The interpreter now publishes activation-qualified `/2` evidence and preorder caller-to-callee call edges for multi-function executions.**

## Accomplishments

- Added frame-local schema and canonical invocation ancestry to every interpreter event path, including immediate returns and abrupt terminals.
- Kept single-function producer bytes on `/1`; legacy oracle freezing now applies only to those legacy documents.
- Added caller-owned `function.called` edges after successful resolution and depth admission, plus projection, preorder, and rejection controls.

## Verification

Passed:

- `go test ./internal/compiler/interp -run 'TestInvocationThreadsThroughAllEventPaths|TestExecutionSchemaSelectionPreservesLegacy' -count=1 -v`
- `go test ./internal/compiler/interp -run 'TestFunctionCalledPreorderAndOwnership|TestFunctionCalledProjectionRemoval|TestRejectedCallEmitsNoCalledEdge' -count=1 -v`
- `go test ./internal/compiler/interp -count=1`

## Deviations from Plan

The existing interpreter oracle corpus included multi-function `/1` goldens. Its byte-freeze check is now deliberately limited to legacy `/1` producers; `/2` behavior is covered by the new focused tests.

## Issues Encountered

Git metadata is read-only in this sandbox, so `git add`/`git commit` cannot create the required commits (`index.lock: Operation not permitted`). The implementation and this summary remain uncommitted for a caller with Git metadata write access.

## Next Phase Readiness

Interpreter `/2` evidence is ready for the peer validator and native emitter work.

---
*Phase: 15-event-identity-lang-execution-2*
*Completed: 2026-09-19*

---
phase: 36-scalar-helper-calls
plan: "03"
subsystem: compiler-validation
tags: [compiler, corevalidate, originvalidate, pathoracle, scalar-calls, ownership]

requires:
  - phase: 36-02
    provides: Typed scalar OpCall instructions with source, fresh target, type, and callee identifiers.
provides:
  - Independent raw core validation for exact scalar call contracts and static target freshness.
  - Independent origin transfer of scalar call results to initialized, value-unknown states.
  - Narrow path-oracle admission for scalar calls with no result loan ancestry or resource authority.
affects: [36-04, compiler-validation, scalar-calls]

actuals:
  tokens: 17425
  tasks: 3
  commits: 5
commits: 5
plan_head_before: 7b3c334f55d63533ad49b3b8c3768a2cb63aa552
plan_head_after: 36dcfabbb342b3c7944a657ffe14536ed3364694

tech-stack:
  added: []
  patterns:
    - Independent raw-program contract derivation in each validation peer.
    - Static producer-target freshness distinct from repeated dynamic execution and mutable stores.
    - Scalar helper results remain initialized but numerically unknown.

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_phase36_test.go
    - internal/compiler/originvalidate/originvalidate_phase36_test.go
    - internal/compiler/pathoracle/pathoracle_phase36_test.go
    - .planning/phases/36-scalar-helper-calls/36-03-pathoracle-red.json
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/pathoracle/pathoracle.go

key-decisions:
  - "Each peer resolves declared signatures and call authority from its own program facts; no checker verdict or numeric callee summary is trusted."
  - "Call results start initialized and value-unknown, while freshness applies to static producer targets rather than repeated runtime execution or stores."
  - "CALL-01 and CALL-02 remain open until the phase-level obligations are complete."

patterns-established:
  - "Raw OpCall validation: independently check callee, exact scalar signature, initialized source, closed fields, and fresh result target."
  - "Path authority: retain source-loan use at a call while giving the result no loan ancestry; scalar-safe calls do not widen generic OpCall authority."

requirements-completed: []

coverage:
  - id: D1
    description: "Raw core validation accepts only exact scalar helper contracts with initialized inputs and fresh result targets."
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_phase36_test.go#TestPhase36RawCallShapeAndFreshness"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_phase36_test.go#TestPhase36RawCallRefusalRemovalControls"
        status: pass
    human_judgment: false
  - id: D2
    description: "Origin analysis independently transfers call results as initialized unknown scalars and checks static freshness through branches and loops."
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_phase36_test.go#TestPhase36OriginCallFreshnessAndTransfer"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_phase36_test.go#TestPhase36OriginCallRefusalRemovalControls"
        status: pass
    human_judgment: false
  - id: D3
    description: "The path oracle admits only scalar-safe calls, accounts for source use, and refuses result loan ancestry and resource/back-edge authority."
    verification:
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_phase36_test.go#TestPhase36ScalarCallAuthorityBoundaries"
        status: pass
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_phase35_test.go#TestPhase35PureOperationAncestry"
        status: pass
    human_judgment: false

duration: 22min
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 03: Independent Scalar Call Validation Summary

Independent raw validation and ownership analysis now admit exact scalar helper calls while keeping call results value-unknown and outside loan ancestry.

## Performance

- **Duration:** 22 min
- **Started:** 2026-10-15T01:59:18Z
- **Completed:** 2026-10-15T02:21:55Z
- **Tasks:** 3
- **Plan files changed:** 7

## Accomplishments

- `corevalidate` independently checks all four U64/Bool parameter/result pairs, initialized arguments, closed call fields, result types, and static producer-target freshness. Valid branch and loop assignments pass; malformed raw calls are refused without passing through `check`.
- `originvalidate` derives declared call contracts from its own function facts and transfers results as initialized unknown values. Tests cover branch/loop stores, a helper returning zero consumed by remainder without invented numeric propagation, malformed contracts, and freshness controls.
- `pathoracle` admits only typed scalar-safe calls, keeps the argument's source-loan use, and gives the result no loan ancestry or resource/back-edge authority. The existing ByteView owner-wrapper boundary remains covered.

## Verification

All required selectors produced fresh root RUN/PASS evidence:

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase36RawCallShapeAndFreshness|TestPhase36RawCallRefusalRemovalControls)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase36OriginCallFreshnessAndTransfer|TestPhase36OriginCallRefusalRemovalControls)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/pathoracle -run '^(TestPhase36ScalarCallAuthorityBoundaries|TestPhase35PureOperationAncestry)$' -count=1 -v -timeout=600s`
- `git diff --check` passed. No full suite or whole-package test was run.

## TDD Evidence

- Task 1's raw core selector first failed on the newly introduced scalar-call behavior, then passed after independent contract and transfer checks were added. The test-first commit is `1cbf9b6d`.
- Task 2's selector was observed failing with `core.scalar_analysis_invalid` before origin call transfer existed; that scratch failure was not persisted as a separate evidence artifact.
- Task 3's targeted raw authority mutant was converted from Go JSON events to TAP, classified `RED_EVIDENCE_OK`, and saved in `36-03-pathoracle-red.json` with the test-first commit `859c5c05`.

## Task Commits

1. **Task 1: Validate raw scalar call contracts and freshness** — `1cbf9b6d` (test), `b717e4f5` (feat).
2. **Task 2: Derive origin call contracts and unknown results** — `8cd042df` (feat; test and implementation in one commit).
3. **Task 3: Constrain scalar-call path authority** — `859c5c05` (test and RED evidence), `36dcfabb` (feat).

Plan metadata is committed separately after canonical state and roadmap updates.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Record a diagnostic for malformed raw calls with a foreign field**
- **Found during:** Task 1
- **Issue:** A structural `OpCall` path returned early when its `Foreign` field was populated without recording a diagnostic, so malformed raw input could appear valid.
- **Fix:** Route this refusal through the core validator's diagnostic path and add a refusal regression control.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/corevalidate/corevalidate_phase36_test.go`
- **Verification:** `TestPhase36RawCallRefusalRemovalControls` passed in the exact Task 1 selector.
- **Committed in:** `b717e4f5`

**Total deviations:** 1 auto-fixed (1 Rule 1).
**Impact on plan:** The fix closes a fail-open diagnostic path directly exercised by the new raw-call contract tests.

The deviation was recorded and marked fixed as entry 12 in `.planning/WINDOWS.md`.

## Issues Encountered

The first attempt to exercise an origin refusal seam through an internal package test import did not compile. The test was moved to the external `originvalidate_test` package and used the existing narrow test setter; the focused selectors then passed.

## User Setup Required

None.

## Next Phase Readiness

Independent peer and path-oracle gates now accept the Phase 36 scalar-call IR shape. Interpreter/native behavior and the runnable helper remain with later Phase 36 plans. CALL-01 and CALL-02 are intentionally still open pending the remaining phase obligations.

---
*Phase: 36-scalar-helper-calls*
*Completed: 2026-10-15*

## Self-Check: PASSED

The summary exists and all five task commits are ancestors of the current plan HEAD.

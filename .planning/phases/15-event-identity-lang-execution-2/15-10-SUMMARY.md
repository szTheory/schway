---
phase: 15-event-identity-lang-execution-2
plan: 10
subsystem: testing
tags: [go-test, github-actions, ci, package-selection]
requires:
  - phase: 15-event-identity-lang-execution-2
    provides: Phase 15 native admission and session evidence seams
provides:
  - Package-correct cross-platform Phase 15 CI selectors
  - Package-aware CI source pin with a negative mismatch control
affects: [ci, phase-15-verification]
actuals:
  tokens: 1310
  tasks: 2
  commits: 3
tech-stack:
  added: []
  patterns: [Focused Go test selections are pinned to their owning package]
key-files:
  created: []
  modified:
    - .github/workflows/ci.yml
    - internal/compiler/session/session_phase6_test.go
key-decisions:
  - "Keep the evidence aggregate split by owning Go package so no-match success cannot masquerade as test execution."
patterns-established:
  - "CI source pins associate each required test selector with its expected Go package."
requirements-completed: [OBS-04, NAT-10]
coverage:
  - id: D1
    description: The cross-platform aggregate selects the native decoder seam and all four session evidence seams from their owning Go packages.
    requirement: OBS-04
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/native -run 'TestDecodeExecutionSchema2AdmissionSeam' -count=1 -v"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestSchema2ComparisonRequiresPeerVerdict|TestPhase5CompareProgramEnginesPreservesLegacySchemas|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert' -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: The recurring CI source pin accepts the owning package and rejects the same decoder test selected under the session package.
    requirement: NAT-10
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestCIWorkflowSelectionPinsPackageOwnership"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestCIWorkflowRunsCurrentAggregateGate"
        status: pass
    human_judgment: false
duration: 4min
completed: 2026-09-24
status: complete
plan_head_before: 70419441167d15c592082e2ce532b94407647060
---

# Phase 15 Plan 10: Package-Correct CI Evidence Summary

**The Phase 15 CI aggregate now executes each named Go test from its owning package and the source pin rejects package/test mismatches.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-24T03:43:18Z
- **Completed:** 2026-09-24T03:47:06Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Split the native decoder admission seam into a focused `./internal/compiler/native` invocation while retaining the four session seams under `./internal/compiler/session`.
- Extended the aggregate source pin to check package ownership for each required selector.
- Added table-driven positive, negative, session, and missing-selector controls; the historical Phase 6 script and Ubuntu/macOS matrix assertions remain required.

## Task Commits

1. **Task 1: Run the native admission seam from its owning package in the cross-platform aggregate** - `0adaa49` (`fix`)
2. **Task 2 RED: Prove the package mismatch is rejected** - `e37b9a4` (`test`)
3. **Task 2 GREEN: Pin aggregate test names to Go packages** - `b733146` (`fix`)

**Plan commits measured from `plan_head_before`:** 3

## Files Created/Modified

- `.github/workflows/ci.yml` - Runs the native admission seam and session evidence seams with package-correct focused commands.
- `internal/compiler/session/session_phase6_test.go` - Pins required selectors to package ownership and exercises the wrong-package negative control.

## Decisions Made

- Kept CI invocations split by package so Go's successful zero-match behavior cannot hide a skipped seam.

## Deviations from Plan

None - plan executed as written. Task 2 used a RED/GREEN commit sequence; the RED subtest failed on the intended assertion, and the GREEN implementation made all cases pass.

## Issues Encountered

- Local native and session tests passed. Hosted Ubuntu/macOS execution was not observable from the isolated, unpushed worktree; both runners remain configured in the recurring CI matrix, which will execute when the change reaches a CI-triggered ref. No human UAT is required for these deterministic checks.

## User Setup Required

None - no external service configuration is required.

## Next Phase Readiness

- The focused seams and source contract are ready for recurring Linux and macOS CI execution.
- No human verification is required for the deterministic package-selection contract.

---
*Phase: 15-event-identity-lang-execution-2*
*Completed: 2026-09-24*

## Self-Check: PASSED

- Summary file exists at the required phase path.
- Task commits `0adaa49`, `e37b9a4`, and `b733146` are present in Git history.
- The measured plan task commit count is 3, matching `actuals.commits`.

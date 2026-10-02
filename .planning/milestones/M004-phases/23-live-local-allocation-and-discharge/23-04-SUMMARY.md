---
phase: 23-live-local-allocation-and-discharge
plan: 04
subsystem: compiler/native
tags: [c17, file-acquisition, failure-cleanup, path-bounds]

requires:
  - phase: 23-01
    provides: Retained file-byte app and bounded PathToken acquisition seam
provides:
  - Typed C adapter acquisition failures with initialized null-owner records and deterministic partial cleanup
  - Public path-byte admission and bounded acquisition diagnostics for the retained native app
affects: [23-05, 23-06, RES-07]

actuals:
  tokens: 5366.25
  tasks: 2
  commits: 4
  plan_head_before: 581754eea346360ace666787398a95d1a5bafd7d

tech-stack:
  added: []
  patterns:
    - Injected C system-call hooks exercise acquisition failures and cleanup deterministically
    - Every failed acquisition returns a typed status with no published owner

key-files:
  created: []
  modified:
    - examples/phase23/adapter.c
    - internal/compiler/native/native_app.go
    - internal/compiler/native/native_app_test.go
    - cmd/lang/main.go
    - cmd/lang/main_test.go
    - .planning/phases/23-live-local-allocation-and-discharge/23-RESEARCH.md

key-decisions:
  - Keep descriptor closure and partial allocation cleanup inside the C adapter before publishing acquisition failure.
  - Reject embedded NUL in the retained Go runner before launching the generated executable.

requirements-completed: [RES-04, RES-07]

coverage:
  - id: D1
    description: "C acquisition failures return distinct initialized statuses, close opened descriptors, and free partial allocations without publishing an owner."
    requirement: RES-07
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase23Acquire' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The retained public app rejects invalid path bytes before adapter access, bounds typed diagnostics, and preserves the independent 65/66 success witnesses."
    requirement: RES-04
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./cmd/lang -run '^TestPhase23(Acquire|PublicFileByte)' -count=1"
        status: pass
    human_judgment: false

duration: 1h 16m
completed: 2026-09-28
status: complete
---

# Phase 23 Plan 04: Bounded Acquisition and Failure Cleanup Summary

**Native file acquisition now reports distinct bounded failures, frees partial storage before returning, and rejects invalid public path bytes before adapter access.**

## Performance

- **Duration:** 1h 16m
- **Started:** 2026-09-27T23:01:09-04:00
- **Completed:** 2026-09-28T00:17:53-04:00
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added deterministic C adapter fault injection for open, metadata, allocation, read, and close paths; each failure returns its own initialized typed status, closes the descriptor when opened, and frees partial storage.
- Exercised regular-file admission, nonblocking open, EINTR retry, empty and overlong inputs, exact one-byte success, and independent 0x41/0x42 values through native controls.
- Added public path length and embedded-NUL checks with bounded typed diagnostics while preserving the retained application’s 65/66 output behavior.
- Recorded the adapter’s regular-file policy and evidence status in the Phase 23 research notes.

## Task Commits

1. **Task 1 RED:** `4b96f71` — `test(23-04): exercise partial acquisition cleanup failures`
2. **Task 1 GREEN:** `d0a9e31` — `feat(23-04): return typed acquisition failures safely`
3. **Task 2 RED:** `b60ed07` — `test(23-04): cover public path bounds and diagnostics`
4. **Task 2 GREEN:** `7e3ddf2` — `feat(23-04): reject invalid bounded input before launch`

## Files Created/Modified

- `examples/phase23/adapter.c` — injectable POSIX calls, typed failure statuses, descriptor closure, and partial-buffer cleanup.
- `internal/compiler/native/native_app.go` — pre-launch embedded-NUL rejection with a typed input error.
- `internal/compiler/native/native_app_test.go` — compiled C fault-injection controls and boundary tests.
- `cmd/lang/main.go` — bounded CLI failure exit handling for the typed input error.
- `cmd/lang/main_test.go` — public path boundary, diagnostic, and success-witness checks.
- `.planning/phases/23-live-local-allocation-and-discharge/23-RESEARCH.md` — dated implementation and evidence status.

## Decisions Made

- The C adapter owns cleanup for every descriptor and partial allocation it creates; it publishes success only after a successful close.
- Embedded NUL is rejected in the Go runner because process argument transport cannot preserve it to the C adapter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Reject embedded NUL before native process launch**
- **Found during:** Task 2
- **Issue:** The existing runner delegated an embedded-NUL argument to `os/exec`, producing a generic launch failure rather than rejecting the invalid token before adapter access.
- **Fix:** Added a typed `native.input_contains_nul` validation error in the retained runner and mapped it to the bounded input-error CLI channel.
- **Files modified:** `internal/compiler/native/native_app.go`, `cmd/lang/main.go`, `cmd/lang/main_test.go`
- **Verification:** Both focused Phase 23 test commands passed.
- **Committed in:** `7e3ddf2`

**2. GSD worktree scope warning — declared inventory omitted three actual files**
- **Found during:** Wave 2 `worktree.cleanup-wave` reconciliation.
- **Issue:** The plan’s `files_modified` list omitted `internal/compiler/native/native_app.go`, `cmd/lang/main.go`, and this dated update to `23-RESEARCH.md`.
- **Disposition:** The two Go files implement the plan’s required pre-launch path-token rejection; the research update records the observed failure-injection and input-boundary evidence. No new behavior beyond the plan’s acceptance criteria was introduced.
- **Reported by:** GSD cleanup as `scope_out_of_declared` for all three paths; merge succeeded.

**Total deviations:** 1 auto-fixed (Rule 2 - Missing Critical) and 1 plan-inventory correction. **Impact:** Behavioral scope is unchanged; the actual file scope exceeded the declared inventory for the paths listed above.

## Issues Encountered

- The first injected C harness compile exposed missing standard I/O and feature-test declarations; the harness translation unit and guarded feature macro were corrected before the task’s verification passed.
- The first public NUL-path assertion exposed the generic process-launch failure described above; the typed pre-launch rejection fixed it.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 23-04 is complete with both focused verification commands passing. Requirements RES-04 and RES-07 are shared by unfinished Phase 23 plans, so `requirements.ready-ids` reported 0/2 ready and `.planning/REQUIREMENTS.md` was left for the orchestrator’s phase-level update. Plans 23-02, 23-03, 23-05, 23-06, and 23-07 remain in the phase queue.

## Self-Check: PASSED

- SUMMARY exists and all four task commits are present in git history.
- Both plan verification commands passed after implementation.
- `git diff --check` passed.

---
*Phase: 23-live-local-allocation-and-discharge*
*Completed: 2026-09-28*

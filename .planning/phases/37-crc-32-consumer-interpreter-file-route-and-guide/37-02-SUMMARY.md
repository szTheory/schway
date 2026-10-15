---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "02"
subsystem: interpreter
tags: [go, filebytes, interpreter, evidence-capture]
requires:
  - phase: 37-01
    provides: CRC consumers and the admitted FileBytes contract used by this interpreter route
provides:
  - Fixed bounded file reader and module-aware interpreter file-application API
  - Capture-disabled evaluation that retains no optional trace identity or events
  - Independent controls for typed failures, output ordering, defects, cleanup, and capture exhaustion
  - Canonical bad/clean/default proof for the interpreter outcome boundary
affects: [phase-37, interpreter, FileBytes]
actuals:
  tokens: 26292
  tasks: 2
  commits: 5
plan_head_before: aded5aca493cf31f39f56e3ba90797d8169518c7
plan_head_after: ffa8efed7dce87347f69c1465a8c8fc63f297f2c
tech-stack:
  added: []
  patterns:
    - Fixed descriptor reader with actual cap-plus-one read and checked close
    - Capture policy propagated through interpreter frames before evidence work
key-files:
  created:
    - internal/compiler/interp/filebytes_reader.go
    - internal/compiler/interp/filebytes_reader_test.go
    - internal/compiler/interp/interp_phase37_test.go
    - internal/compiler/session/session_file_application.go
    - internal/compiler/session/session_phase37_file_application_test.go
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/prohibition-interpreter-violation.json
  modified:
    - internal/compiler/interp/interp.go
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-02-PLAN.md
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md
key-decisions:
  - Keep the production interpreter API bound to its fixed reader; reader injection remains private to semantic controls.
  - Treat evidence capacity as optional capture status while retaining the actual outcome, prior output, and cleanup semantics.
  - Leave APP-13 and APP-14 requirement completion to parent phase acceptance.
patterns-established:
  - Fixed file capability: validate the checked FileBytes contract at the interpreter boundary, then use one nonblocking regular descriptor and one bounded read.
  - Capture-free evaluation: bypass event, identity, ordinal, and recorder construction while preserving evaluator behavior.
requirements-completed: []
duration: ~20min visible execution window; initial context loading preceded the first recorded task commit
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 02: Interpreter File Route and Capture-Independent Outcomes Summary

**The CRC consumers now run through a fixed interpreter file route without trace retention, while capture exhaustion preserves the real application result.**

## Performance

- **Duration:** Approximately 20 minutes from the first task commit through final focused verification; context loading began earlier.
- **Started:** 2026-10-15T11:01:11-04:00 (first recorded TDD commit; not the beginning of context loading).
- **Completed:** 2026-10-15T11:19:24-04:00.
- **Tasks:** 2/2.
- **Files modified:** 12 plan-owned files.

## Accomplishments

- Added a fixed `ReadFileBytes` path that opens one nonblocking descriptor, confirms it is regular, reads at most 65,537 bytes, follows symlinks, and checks close errors. `session.RunInterpreterFileApplication` performs full independent admission and owner-path validation before pruning; `interp.RunFileApplication` independently checks the exact foreign contract and entry shape before invoking the fixed reader.
- Disabled capture before frame identity, event, ordinal, and recorder work while retaining the existing evaluator. Real files produce `3421780262\n` for `123456789` and `3735780942\n` for 65,536 `0xff` bytes; the latter first ran in 2.449 seconds and measured 2.568 seconds on the final focused rerun (darwin/arm64, Go 1.24.0).
- Added controls for zero/nonzero results, typed `FileError` alternatives, prior output followed by a checked defect, abort-like live-owner policy, descriptor close/error paths, and capture exhaustion. The canonical phase-local proof now binds three subjects and passed bad, clean, and default controls with no skipped roots.

The evidence shows only this plan's witnesses. It does not establish arbitrary-divergence timeout parity or recoverable Go allocator exhaustion; `AllocFailed` is modeled through the private reader seam. The native empty-path CLI behavior was left untouched.

## Verification

- PASS: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/session -run '^(TestPhase37InterpreterFileApplication|TestPhase37CaptureDisabledRetention)$' -count=1 -v -timeout=600s`.
- PASS: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37CaptureDisabledOutcomes|TestPhase37FileReaderBasic|TestPhase36InterpreterEvidenceBoundaries|TestPhase321AbortLikeDefectFileRoute)$' -count=1 -v -timeout=600s`.
- PASS: `node .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs --prove`; all three descriptors had named bad mutations and passing clean/default controls through the canonical producer.

## Task Commits

1. **Task 1 RED:** `a6b4e69e` (`test(37-02): pin interpreter file-app capture behavior`).
2. **Task 1 GREEN:** `a49af9c7` (`feat(37-02): run CRC file apps without trace capture`).
3. **Task 2 RED:** `3868c8f8` (`test(37-02): add independent capture outcome controls`).
4. **Task 2 GREEN:** `d6078b39` (`feat(37-02): preserve interpreter outcomes across capture failure`).
5. **Canonical proof receipt:** `ffa8efed` (`docs(37-02): bind prohibition receipt to implementation head`).

## Decisions Made

- Kept the ordinary production API on the fixed reader and kept reader injection private for deterministic tests.
- Recorded semantic owner release counts separately from physical descriptor-close checks; no host allocator failure recovery is claimed.
- Left phase requirements and global planning metadata for APP-13 and APP-14 pending parent phase acceptance.

## Deviations from Plan

### Auto-added proof binding

**1. [Rule 2 - Missing proof binding] Added the interpreter outcome descriptor to this plan.**
- **Found during:** Task 2 (Keep real outcomes and enabled capture controls).
- **Issue:** The plan requested a real projected interpreter prohibition receipt, but only Plan 37-01 contained prohibition descriptors, so the new witness would not have been traceable to its owning plan.
- **Fix:** Added the flat descriptor and fixture path to Plan 37-02, extended the closed adapter to compare the combined plan projection, and kept the original unresolved projection history intact.
- **Files modified:** `37-02-PLAN.md`, `37-prohibition-check.test.cjs`, `prohibition-interpreter-violation.json`, `37-PROHIBITION-EVIDENCE.md`.
- **Verification:** Canonical bad/clean/default proof passed for all three descriptors.
- **Committed in:** `d6078b39`; refreshed receipt in `ffa8efed`.

**Total deviations:** 1 auto-added proof binding. **Impact:** Scoped planning metadata now identifies the new witness; no runtime scope was added.

## Issues Encountered

The first proof mutation targeted a separate interpreter entry branch and did not fail the intended assertion. The adapter was corrected to overlay the exact `runFileApplicationWithProbe` capacity return; the canonical bad control then failed only at the named prior-output assertion.

## User Setup Required

None.

## Next Phase Readiness

Plan 37-02 is complete and ready for its dependent plan. APP-13 and APP-14 remain pending the parent’s phase-level acceptance; this summary records the concrete witnesses above without marking either full requirement complete.

## Self-Check: PASSED

The summary exists, all five recorded task/proof commits are ancestors of HEAD, and the plan-owned files are present. Shared dirty planning files were excluded from these commits.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

---
phase: 36-scalar-helper-calls
plan: "08"
subsystem: testing
tags: [compatibility, evidence, owner-capture, phase-transition]

requires:
  - phase: 36-07
    provides: Exact independent surface manifest and observed owner roots
provides:
  - Source-bound final compatibility and per-knower acceptance receipt
  - Dated evidence and next-capability amendments in both living documents
affects: [phase-36-audit, phase-37-planning, phase-38-receipts]

actuals:
  tokens: 6747
  tasks: 2
  commits: 2
  plan_head_before: 5474f16ecd3e80cf7cdafec73173b2af48c6556f
  plan_head_after: 92300103d4e64e14718e637c3c5e20281a364164

tech-stack:
  added: []
  patterns:
    - Owner-command logs bind argv, source HEAD, expected roots, exits, and output hashes
    - Current source inspection, fresh checks, and historical receipts remain separate

key-files:
  created:
    - .planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md
  modified:
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md

key-decisions:
  - "Bind the final owner capture to compiler source HEAD d117ca4058fbf6a727cb5aa634768cf6194bac34; later commits changed planning documents only."
  - "Leave CALL-01 and CALL-02 pending for the phase-owned independent gates."

patterns-established:
  - "Acceptance rows cite actual test owners and separate numeric, causal, route, and capture claims."
  - "Living roadmap amendments distinguish current source, new focused checks, and historical receipts."

requirements-completed: []
duration: "not recorded; first-to-last accepted owner capture was 7m 25s"
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 08: Final compatibility and acceptance record Summary

**The integrated helper slice now has source-bound compatibility evidence and a per-knower acceptance record, with unresolved phase assumptions kept visible.**

## Performance

- **Duration:** The executor start marker was not preserved; accepted owner captures span 7m 25s.
- **Started:** Not recorded.
- **Completed:** 2026-10-15T05:30:20Z.
- **Tasks:** 2.
- **Files modified:** 3 task deliverables.
- **Actuals tokens:** 6,747, measured as 26,987 added characters divided by four across the receipt and two living-document amendments.

## Accomplishments

- The frozen compatibility controls passed 15/15 roots at source HEAD `5474f16e`; the tracer feedback rerun passed 15/15 roots at `d117ca40`.
- The amended final fixed-owner capture passed 45/45 expected roots across 12 packages at source HEAD `d117ca4058fbf6a727cb5aa634768cf6194bac34`. Every command exited zero, with no `FAIL` or `SKIP` counted as a pass. The accepted manifest is `/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json`; the receipt links its logs and hashes.
- The acceptance record maps each stable claim to observed test owners and keeps old callee ByteView bounds separate from caller argument bounds. It also keeps source lowering, raw-peer contracts, literal engine outcomes, causal attribution, and optional capture outcomes within their actual evidence boundaries.
- Both living documents received dated capability and dormant-seed dispositions. They distinguish current source inspection, fresh focused checks, and historical receipts, and make no hosted-evidence claim.

## Task Commits

1. **Task 1: Recheck integrated helper additions against immutable no-new-call C and old safety controls** - `d117ca40` (`docs(36-08): capture Phase 36 compatibility evidence`).
2. **Task 2: Record per-knower acceptance and append the living capability/seed transition** - `92300103` (`docs(36-08): record Phase 36 acceptance and transition`).

## Files Created/Modified

- `.planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md` - source-bound frozen compatibility results, current owner captures, claim mapping, unresolved assumptions, and actual log links.
- `.planning/PRODUCT-ROADMAP.md` - dated capability, evidence, ownership, and seed transition.
- `.planning/LANGUAGE-MATURITY.md` - dated source and evidence maturity update.

## Decisions Made

- Keep final capture evidence bound to source HEAD `d117ca40`; commit `92300103` and later plan metadata commits affect planning documents only.
- Leave CALL-01 and CALL-02 unchecked and keep Phase 36 closure with the orchestrator's verification and audit gates.
- Preserve the two descriptorless prohibition assumptions as flagged and unverified; no implementation-level enforcement claim was added.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Documentation defect] Corrected acceptance-record references rejected by the fixed document checker**
- **Found during:** Task 2 (per-knower acceptance and living transition).
- **Issue:** An intermediate final capture found a duplicate dated heading and incorrect test-file paths in the receipt and living-document references.
- **Fix:** Removed the duplicate heading and corrected the paths to the actual owner test files.
- **Files modified:** `.planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`.
- **Verification:** The amended fixed-owner command passed 45/45 roots and the final receipt/document checks; latest manifest: `/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json`.
- **Committed in:** `92300103`.

**Total deviations:** 1 auto-fixed documentation defect.
**Impact on plan:** The correction made the named evidence references and dated transition satisfy the planned receipt contract; no production or test source changed.

## Issues Encountered

The first document-check attempts exposed receipt-reference errors and were corrected before the accepted final capture. No execution blocker remains. The run was local and focused; no whole-package suite, hosted dispatch, or push was performed.

## User Setup Required

None.

## Next Phase Readiness

The acceptance receipt and living-document amendments are ready for independent Phase 36 audit. Requirement status and phase closure remain pending the orchestrator-owned gates; this summary completes Plan 36-08 only. No Phase 37 work was started.

---
*Phase: 36-scalar-helper-calls*
*Completed: 2026-10-15*

## Self-Check: PASSED

- Summary and all three task deliverables exist.
- Task commits `d117ca4058fbf6a727cb5aa634768cf6194bac34` and `92300103d4e64e14718e637c3c5e20281a364164` are ancestors of HEAD.
- `git diff --check` passed.

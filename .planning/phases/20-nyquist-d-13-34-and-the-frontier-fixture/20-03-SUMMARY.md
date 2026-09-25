---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "03"
subsystem: testing
tags: [validation, nyquist, evidence, lifecycle]
requires:
  - phase: 20-02
    provides: Updated M002 records and current groundedness scanner
provides:
  - All eight remaining draft validation documents promoted to validated lifecycle status
  - Non-vacuous live/archive zero-draft lifecycle gate
affects: [phase-20-plan-07, validation-archives]
actuals:
  tokens: 4361
  tasks: 3
  commits: 1
tech-stack:
  added: []
  patterns: [phaseArtifactGlob lifecycle scan, seeded draft control]
key-files:
  created: [internal/compiler/session/verification_groundedness_test.go]
  modified:
    - .planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
    - .planning/phases/17-return-type-parameter-type/17-VALIDATION.md
    - .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
    - .planning/phases/19-numeric-literals-and-opconst/19-VALIDATION.md
    - .planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VALIDATION.md
    - .planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-VALIDATION.md
    - .planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md
    - .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
key-decisions:
  - "Use validated plus nyquist_compliant: false for records with incomplete evidence; preserve Phase 13's previously earned true value."
  - "Keep the live exact frontier and generated reconciliation output for Plan 20-07, which owns cross-plan drift and final pinning."
patterns-established:
  - "Validation lifecycle gates scan live and archived documents and include seeded and empty-scan controls."
requirements-completed: [QLT-10]
coverage:
  - id: D1
    description: "The full live/archive validation set has no draft lifecycle status, enforced by a non-vacuous test."
    requirement: QLT-10
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestPhase20ValidationLifecycle$' -count=1"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestEvidenceReconciliationViewIsCurrent$' -count=1"
        status: pass
    human_judgment: false
duration: 53min
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 03: validation lifecycle summary

**All eight remaining validation records now have a non-draft status, and a live/archive test prevents draft records from returning unnoticed.**

## Performance

- **Duration:** approximately 53 minutes
- **Started:** 2026-09-25T02:55:00Z
- **Completed:** 2026-09-25T03:48:00Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- Promoted the four active-milestone and four archived validation records to `validated`, while retaining incomplete grades as `nyquist_compliant: false`; Phase 13's existing `true` evidence remains intact.
- Added `TestPhase20ValidationLifecycle`, which scans live and archived validation documents through `phaseArtifactGlob`, rejects an empty scan, and proves the draft predicate with seeded controls.
- Kept `.planning/EVIDENCE-RECONCILIATION.md` generator-owned; its byte-current check passes. Plan 20-07 owns reconciling findings caused by the changed archive set and final live frontier.

## Task Commits

- `d67e59f` — eight lifecycle records and the seeded non-vacuous gate.

## Verification

- `TestPhase20ValidationLifecycle`, `TestVerificationGroundednessCorpusIsNotEmpty`, and `TestEvidenceReconciliationViewIsCurrent` passed.
- Relevant Phase 17/18/19 and archived Phase 5/6/12/13 evidence tests were exercised. The combined run remains red on two cross-plan assertions: the checked-in validation corpus pair digest is stale after archive edits, and an exemption test still assumes D-14-45 is `UNOWNED(...)`. Plan 20-07 owns these reconciliation updates.

## Deviations from Plan

None. The final reconciliation view content and the Phase 20 exact pin remain with Plan 20-07 after the remaining plans settle.

## Next Phase Readiness

Zero draft validation statuses are now enforced. Plan 20-07 must reconcile archive-derived evidence, current Phase 20 findings, and the final groundedness pin.

---
*Phase: 20-nyquist-d-13-34-and-the-frontier-fixture*
*Completed: 2026-09-25*

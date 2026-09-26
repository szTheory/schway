---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "02"
subsystem: testing
tags: [go, validation, evidence, groundedness]
requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: groundedness scanner, validation-corpus checks, reconciliation view
provides:
  - Updated M002 Phase 07, 08, and 11 validation maps grounded in current tests
  - Separate historical, execution-start, and post-edit frontier measurements
affects: [phase-20-plan-07, phase-20-plan-03, validation-archives]
actuals:
  tokens: 10666
  tasks: 3
  commits: 1
tech-stack:
  added: []
  patterns: [live test-name inventory, exact command evidence, partial validation status]
key-files:
  created: [.planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-02-SUMMARY.md]
  modified:
    - .planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-VALIDATION.md
    - .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md
    - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-RESEARCH.md
    - internal/compiler/session/verification_groundedness_test.go
key-decisions:
  - "Keep the old exact pin unchanged until Plan 20-07 reconciles the final live frontier."
  - "Treat the retired Phase 08 mechanism as historical disposition and retain partial Nyquist status where material review remains."
  - "Keep 07/08/11 validation records validated with nyquist_compliant false while cross-phase evidence is unresolved."
patterns-established:
  - "Measure and label historical, execution-start, and post-edit frontier snapshots separately."
  - "Use literal Go regexp alternation in single-quoted -run expressions."
requirements-completed: [QLT-10]
coverage:
  - id: D1
    description: "Reconciled the three named validation maps against the current test surface and recorded partial status honestly."
    requirement: QLT-10
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestEvidenceReconciliationViewIsCurrent$' -count=1"
        status: pass
      - kind: integration
        ref: "sh scripts/verify-phase7.sh"
        status: fail
    human_judgment: true
    rationale: "The phase records remain partial pending cross-phase archive reconciliation and Plan 20-07's final exact pin."
duration: 270min
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 02: M002 validation records summary

**Phase 07, 08, and 11 validation maps now name live commands and state exactly where current evidence remains partial.**

## Performance

- **Duration:** approximately 270 minutes, including the coordinated wait for Wave 1 source edits
- **Started:** 2026-09-24T23:04:00Z (execution-start evidence capture)
- **Completed:** 2026-09-25T03:34:37Z
- **Tasks:** 3
- **Files modified:** 5 planned implementation files, plus this summary

## Accomplishments

- Preserved the research-time baseline (R1/R2/R3=0, reconciled R2b=23, raw R2b=24, 639 enforced-tier documents, 708 verification commands) and the independent pre-edit execution-start snapshot (R1=0, R2=7, R3=0, R2b=25, 651 documents, 737 commands). The complete start-time test output is at `/private/tmp/phase20-frontier-start.log`.
- Replaced stale Phase 07 fixture assertions, retired the Phase 08 deleted-mechanism test claim, and repaired the Phase 11 elided commands and invalid alternation syntax. Confirmed test names with `go test -list` and ran the commands used for promoted grades.
- Set all three archived validation records to `validated` while retaining `nyquist_compliant: false`; the Phase 08 manual mid-phase decision and Phase 11 guard-ledger completeness remain partial.
- Preserved the old pinned frontier for Plan 20-07 and recorded a post-edit measurement: R1=0, R2=1, R3=0, R2b=23, 656 enforced-tier documents, 726 commands after adding this summary. The sole remaining R2 is `20-VALIDATION.md:44`, `go test ./internal/compiler/session -run '^TestPhase20ValidationLifecycle$' -count=1`, which is outside this plan's file ownership.

## Task Commits

- `c4278ea` — validation maps, groundedness note, research snapshot, and this summary.

## Files Created/Modified

- `07-VALIDATION.md` — current fixture inventory, corrected Go regex commands, and partial status.
- `08-VALIDATION.md` — historical disposition for the removed `computeLoanLastUses` mechanism, with live mapped test commands and partial status.
- `11-VALIDATION.md` — complete command cells, current test-existence markers, exact guard inventory command, and partial status.
- `20-RESEARCH.md` — historical/start/post-edit snapshots and exact interim finding ownership.
- `verification_groundedness_test.go` — documents why the previous literal pin stays in place until Plan 20-07.

## Decisions Made

- The final literal pin belongs to Plan 20-07; this plan measures the current frontier and leaves it red while the one remaining Phase 20 validation reference awaits reconciliation.
- The generated `.planning/EVIDENCE-RECONCILIATION.md` needed no edit: `TestEvidenceReconciliationViewIsCurrent` passed against the current source register.

## Deviations from Plan

None. The reconciliation view was verified byte-current without changing its generated content. Per the parent executor's shared-index instruction, task commits are pending parent serialization.

## Issues Encountered

The Phase 07 aggregate script remains red due cross-plan changes outside this plan's ownership. The rerun after Plan04's process-safety fix passed the native process guard and targeted Phase 07/08/11 tests, but the repository-wide portion reported:

- `TestLanguageMaturityCountsAreCurrent` and `TestSelfDescribingDocsGuardIsNotInert`: `.planning/LANGUAGE-MATURITY.md` reports 140 programs/4602 lines while the live corpus derives 141/4639.
- `TestValidationGradeBarRowExemptionsAreOwned` and `TestDebtRegisterOwnershipGuardIsNotInert`: test assumptions about D-14-45's former `UNOWNED(...)` cell no longer match the edited Phase 14 debt disposition.
- `TestValidationRowGradesAreEarnedOverArchivedCorpus`: checked-in validation corpus pair digest is stale after archive edits.
- `TestPhase16PublicEmitterConsumerInventory`: `session_phase5_corpus_test.go` now has an `EmitNative` call at line 501 while the registry records the old line 468.
- `TestReconciliationVerdictsCarryTheirObligations`: unresolved current R2 at `20-VALIDATION.md:44`, current Phase 11 R3 before the row command was corrected, plus stale historical reconciliation entries for changed Phase 08/11 validation rows.
- `TestVerificationGroundednessFrontierIsPinned` and `TestVerificationGroundednessThreeClassesAreEmpty`: expected one remaining out-of-scope R2 at `20-VALIDATION.md:44`; current measured counts are included above.
- `TestNoSuppressionOutlivesItsWitness`: four existing native skip sites lack resolvable witnesses, and Plan04's two new Phase 20 cache skips at `session_phase20_cache_test.go:30,42` also need witness references. These are owned outside Plan02.

The same tests also reported failures in mutation fixtures tied to those in-progress D-14-45/archive changes. Plan 07 should remeasure after the remaining Phase 20 archive plans settle; this summary does not claim the global suite or exact pin is green.

## User Setup Required

None.

## Next Phase Readiness

Plan 07 has the exact remaining R2 and the post-edit frontier counts needed to reconcile and pin the final corpus. The Phase 07/08/11 maps now provide runnable targeted commands, but their partial flags should remain until the cross-phase archive, maturity, emitter-registry, and witness findings are resolved.

## Self-Check

Summary and five Plan02 implementation files exist; `git diff --check` passes. Named-file commit `c4278ea` contains the implementation files and summary.

---
*Phase: 20-nyquist-d-13-34-and-the-frontier-fixture*
*Completed: 2026-09-25*

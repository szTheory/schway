---
phase: 21-native-emission-ownership-and-resource-discharge-m004
plan: 06
subsystem: evidence
tags: [groundedness, validation-corpus, reconciliation, M003-archive]

requires:
  - phase: 21-04
    provides: Phase 21 verification gap inventory and bounded evidence contracts
provides:
  - Producer-generated validation corpus record bound to the consumer's exact requested pairs
  - Reconciled M003 groundedness findings and current derived reconciliation view
affects: [phase 21 verification, validation grades, groundedness frontier]

actuals:
  tokens: 1443884
  tasks: 2
  commits: 3
plan_head_before: a467c55a5995f5d11bb29b4c62a4c7d0a6d1ad2a

tech-stack:
  added: []
  patterns:
    - Export requested test pairs from the consumer seam and pass each pair as discrete producer arguments
    - Derive the reconciliation view from the shared register parser and renderer

key-files:
  created:
    - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/deferred-items.md
  modified:
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json
    - .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/EVIDENCE-RECONCILIATION.md
    - internal/compiler/session/verification_groundedness_test.go
    - internal/compiler/session/session_test.go

key-decisions:
  - Use the exported consumer request as the only producer input and preserve observed pass, fail, and skip events.
  - Keep archived reconciliation entries in the authored Phase 14 debt register and regenerate the view from its renderer.
  - Preserve the unrelated Phase 16 requirements archive-path failure for separate repair.

requirements-completed: []

coverage:
  - id: D1
    description: A completed 34-pair validation corpus run is bound to the exact exported request and raw JSONL digest.
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^(TestValidationCorpusPairExportMatchesConsumer|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestCheckedInCorpusRecordRejectsTamperingAndVacuity)$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Measured groundedness findings, R2b landing owners, and the generated reconciliation view agree with the archived M003 corpus.
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^(TestVerificationGroundednessThreeClassesAreEmpty|TestVerificationGroundednessFrontierIsPinned|TestReconciliationVerdictsCarryTheirObligations|TestEvidenceReconciliationViewIsCurrent)$' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: Tagged command classification and Phase 20 archived debt lookup remain grounded in live test targets.
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^(TestDebtRegistersAreWellFormed|TestReconciliationVerdictsCarryTheirObligations|TestEvidenceReconciliationViewIsCurrent|TestVerificationGroundednessThreeClassesAreEmpty|TestVerificationGroundednessFrontierIsPinned|TestGoTestTagsAreIncludedInGroundedness|TestPhase20UnownedDebtPopulation)$' -count=1"
        status: pass
    human_judgment: false

duration: 31min
completed: 2026-09-27
status: complete
---

# Phase 21 Plan 06: Archived Evidence Reconciliation Summary

**The validation grade now derives from a digest-bound 34-pair run, and the Phase 14 groundedness ledger matches the measured M003 archive frontier.**

## Performance

- **Duration:** 31 min, measured from the first timestamped artifact; executor entry time was not captured.
- **Started:** Not captured. First timestamped artifact: 2026-09-26T23:59:55Z.
- **Completed:** 2026-09-27T00:31:03Z
- **Tasks:** 2
- **Files modified or created:** 7, excluding this summary.

## Accomplishments

- Exported the consumer's exact 34 package/pattern pairs and ran them sequentially through `scripts/evidence-run-record.sh`. The manifest records producer revision `4079258`, elapsed time `60.923s`, pair digest `f0b7cf7f48c81745e372dab6999cf4b5c0a2d4385749e41c6b70f8241791ec28`, and record digest `1a8bf2936006a0ed4e9dac4e9ca43db3dca88a9f1564e5cb056756934d1ffbad`.
- Reconciled the moved Phase 14 findings and updated the measured frontier and R2b ownership map. The guards report R1=0, R2=0, R3=0, and 28 owned R2b findings while retaining floors of 706 enforced-tier documents and 765 verification commands.
- Added tag-aware parsing for `go test -tags` commands, repaired the Phase 20 debt test's archived artifact lookup, and added D-14-130 for a live R3 command whose roadmap target moved into the M002 archive. The Phase 14 register now declares and contains 86 items; the reconciliation view is re-derived from the shared renderer.
- Both exact plan acceptance commands pass. The corpus keeps the observed failure and skip outcomes from the real run; they were not converted into passes.

## Task Commits

1. **Task 2: Reconcile full groundedness findings and R2b owners** — `914713a` (`fix`)
2. **Task 2 follow-up: Keep the new reconciliation entry in the debt table** — `4079258` (`fix`)
3. **Task 1: Reproduce the exact validation corpus** — `d1b2499` (`fix`)

Task 2 ran before the final Task 1 corpus generation because the corpus grade guard depends on the repaired archive paths and well-formed reconciliation register. The final producer run followed both Task 2 commits.

**Plan metadata:** This summary and `deferred-items.md` are committed together after the production commits.

## Files Created/Modified

- `testdata/phase16/validation-corpus-run-record.jsonl` — actual producer JSONL, with pair and batch completion witnesses.
- `testdata/phase16/validation-corpus-run-record.manifest.json` — revision, exact pair and record digests, completion state, timestamp, and measured duration.
- `.planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` — moved archive paths, R3 reconciliation D-14-130, and synchronized item count.
- `.planning/EVIDENCE-RECONCILIATION.md` — renderer-generated view of the authored reconciliation register.
- `internal/compiler/session/verification_groundedness_test.go` — current frontier and landing records, including build-tag-aware command resolution.
- `internal/compiler/session/session_test.go` — archive-aware Phase 20 debt artifact lookup.
- `.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/deferred-items.md` — records the unrelated suite failure and pending plan 21-05 corpus failure.

## Decisions Made

- Used the consumer export seam as the sole source of the producer's 34 pairs and kept each package/pattern as a discrete argument.
- Preserved the actual cgen failure from pending plan 21-05 in the corpus; the refreshed record reports observations rather than blessing them.
- Kept the generated reconciliation view derived from the authored debt register rather than editing view rows independently.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated the Phase 20 debt test for the archived artifact path**

- **Found during:** Task 1 corpus grade verification
- **Issue:** `TestPhase20UnownedDebtPopulation` read a removed `.planning/phases/20-.../20-DEBT-BASELINE.md` path, causing validation rows that cite the debt-register test to lose their earned grade.
- **Fix:** Locate the unique Phase 20 baseline with `phaseArtifactGlob` and retain the single-match assertion.
- **Files modified:** `internal/compiler/session/session_test.go`
- **Verification:** `TestPhase20UnownedDebtPopulation` passes in the groundedness acceptance run.
- **Committed in:** `914713a`

**2. [Rule 1 - Bug] Included Go build tags in groundedness command classification**

- **Found during:** Task 2 frontier verification
- **Issue:** The test index omitted custom `go test -tags` constraints, so the Phase 21 tagged LTO command could be classified against the wrong set of test names.
- **Fix:** Parse both `-tags value` and `-tags=value`, build the index with those tags, and add a control covering tagged and untagged classification.
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `TestGoTestTagsAreIncludedInGroundedness` passes with the task 2 controls.
- **Committed in:** `914713a`

**3. [Rule 1 - Bug] Added the D-14-130 debt table row and synchronized count**

- **Found during:** Task 1 corpus grade verification
- **Issue:** The newly measured R3 reconciliation had a detail section but no corresponding Items row, so `TestDebtRegistersAreWellFormed` failed and the corpus correctly retained a failing result.
- **Fix:** Add D-14-130 to the Items table and increment `items` from 85 to 86.
- **Files modified:** `.planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md`
- **Verification:** Debt-register, reconciliation, frontier, and view tests pass; the final corpus grade test passes.
- **Committed in:** `4079258`

**Total deviations:** 3 auto-fixed (2 Rule 1, 1 Rule 3). Execution order was adjusted so reconciliation and archive repairs preceded the final corpus run. **Impact:** Required for the exact archived corpus to pass its grade and groundedness guards; no runtime or emitter scope changed.

## Issues Encountered

- The producer JSONL records a failing `internal/compiler/cgen` event for `TestProgramBorrowedByPointerDisposition` (plus its package-level failure), which is assigned to pending plan 21-05. The failure was preserved as observed and recorded in `deferred-items.md`.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session` ran for 56.8s and failed only `TestPhase16EmitterCutsAreAmendedAndOwned`: it reads the removed root `.planning/REQUIREMENTS.md`. Phase 21 verification criterion 4 already records this archive-path gap; it remains deferred to separate repair.
- No authentication gates or user setup were required.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The 21-06 validation-grade and groundedness gaps are closed by their focused guards. Complete plan 21-05, then refresh Phase 21 verification after both gap plans pass. The deferred full-suite archive-path issue remains visible in `deferred-items.md`.

## Self-Check

PASSED

- All task artifacts and the summary file exist.
- Task commits `914713a`, `4079258`, and `d1b2499` are present; three production commits are measured from `plan_head_before`.
- Both exact plan acceptance commands passed.

---
*Phase: 21-native-emission-ownership-and-resource-discharge-m004*
*Completed: 2026-09-27*

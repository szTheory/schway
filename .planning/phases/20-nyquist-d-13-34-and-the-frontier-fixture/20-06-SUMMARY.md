---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "06"
subsystem: heldout-evidence-and-debt
requires:
  - phase: 20-08
    provides: green timing preflight and phase timing evidence
  - phase: 20-07
    provides: live debt population and five-item cap controls
provides:
  - Replacement held-out move and borrow fixtures with measured structural separation
  - Closed D-13-34 ledger row with exact implementation commit
  - Restored D-06-29 structural inference with its corpus-size limitation retained
  - Phase 20 final validation and verification evidence
affects: [QLT-10, PRC-02, D-06-29, D-13-34]
actuals:
  tasks: 3
  selected_outcome: replace
  open_unowned_debt_items: 4
  full_suite: passed
source_revision: b1961d9
key-files:
  created:
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-06-SUMMARY.md
  modified:
    - testdata/phase6/heldout_move_defect.lang
    - testdata/phase6/heldout_borrow_defect.lang
    - internal/compiler/session/session_phase6_injectors_test.go
    - .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json
key-decisions:
  - "The user selected replacement fixtures at the blocking-human checkpoint on 2026-09-25."
  - "The move marker remains on the middle take so the existing single-pass repair re-verifies the added third hop."
  - "D-13-34 closes at 4658eb1; the later evidence-ledger and pin refresh is committed at b1961d9."
verification:
  - "Phase 6 structural summaries pass for all three classes: move 3 takes vs 2; borrow 4 borrows vs 3; match 3 arms vs 2."
  - "Move and borrow injectors produce their intended single ownership defects; TestRepairDriverFixesEveryDefectClassSinglePass passes."
  - "TestDebtRegistersAreWellFormed, TestPhase20UnownedDebtPopulation, and TestPhase20UnownedDebtCapRule pass; the live set is D-10-C04, D-12-43, D-14-46, D-14-47."
  - "The 33-pair archived evidence record and TestValidationRowGradesAreEarnedOverArchivedCorpus pass; record revision b1961d9, elapsed 108.107s."
  - "GOCACHE=/private/tmp/phase20-gocache GOMAXPROCS=4 go test ./... -count=1 passes on the final implementation/evidence tree."
status: complete
completed: 2026-09-25
---

# Phase 20 Plan 06: Close D-13-34 with Replacement Fixtures

The user selected **replace** at the blocking-human checkpoint. The original alpha-rename weakness is preserved in the chronology, while current held-out inputs now differ structurally from their derivation partners under the unchanged five-component predicate.

The move fixture now has a three-hop take chain, with the injector marker on the middle hop. Keeping the marker there is necessary for the existing repair driver's `use_transfer_target` repair to reverify the remaining chain. The borrow fixture adds a reborrow after the marked shared borrow, keeping the original borrow-conflict trigger intact. Both injectors still produce their expected single diagnostic and a driver-eligible repair.

D-13-34 is closed in the source debt register as `CLOSED(4658eb1)`. The current live open-unowned set is four items, within Plan 07's five-item cap. D-06-29's structural inference is restored for move and borrow, with the explicit residual that three classes and six fixtures do not establish generalization across the language's full program space.

## Verification

- `GOCACHE=/private/tmp/phase20-gocache GOMAXPROCS=4 go test ./internal/compiler/session -run 'TestPhase6(HeldoutPairs|DefectCorpusIsHeldOut)|TestDebtRegistersAreWellFormed|TestPhase20UnownedDebt(Population|Cap)' -count=1` — passed.
- `GOCACHE=/private/tmp/phase20-gocache GOMAXPROCS=4 go test ./cmd/lang-repair -run '^TestRepairDriverFixesEveryDefectClassSinglePass$' -count=1` — passed.
- Archived row-grade scan, generated debt view, source-derived population, and payload replay pins — passed.
- `GOCACHE=/private/tmp/phase20-gocache GOMAXPROCS=4 go test ./... -count=1` — passed across all packages.

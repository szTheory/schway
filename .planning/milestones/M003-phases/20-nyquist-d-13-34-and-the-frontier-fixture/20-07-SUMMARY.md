---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "07"
subsystem: archive-evidence-and-frontier
tags: [qlt-10, prc-02, groundedness, debt-cap, validation-maps]
requires:
  - phase: 20-02
    provides: corrected archived Phase 07/08/11 commands and current frontier cohort
  - phase: 20-03
    provides: zero-draft validation lifecycle and interim evidence scan
  - phase: 20-04
    provides: final cache source call sites and Darwin-only witnesses
  - phase: 20-05
    provides: source-derived M002/M003 debt population and seeded cap rule
  - phase: 20-09
    provides: exact repaired archived validation rows
provides:
  - Current maturity count, emitter call inventory, and six resolvable skip witnesses
  - Grade and non-inertness schema on validated Phase 17–19 maps and Phase 20 plan map
  - Closed stale archive reconciliation rows and a generated 54-entry evidence view
  - Live five-item open-unowned debt cap with D-13-34 still open
  - Passing groundedness lint against the current measured corpus
affects: [20-10, 20-08, 20-06, QLT-10, PRC-02]
actuals:
  tokens: 0 # Executor token telemetry was unavailable during shared-worktree execution.
  tasks: 2
  commits: 2
plan_head_before: 075f2cd
tech-stack:
  added: []
  patterns: [source-backed inventory reconciliation, generated evidence-view refresh, live debt-cap pin]
key-files:
  created: []
  modified:
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/EVIDENCE-RECONCILIATION.md
    - .planning/LANGUAGE-MATURITY.md
    - .planning/UNREACHABLE-CLAIMS.md
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-DEBT-BASELINE.md
    - testdata/phase16/public-emitter-consumers.json
    - internal/compiler/native/native_test.go
    - internal/compiler/session/session_phase20_cache_test.go
    - internal/compiler/session/session_test.go
    - internal/compiler/session/verification_groundedness_test.go
key-decisions:
  - "Keep D-14-50/51/52/54 closures tied to their direct replacement-test runs and exact row parsers; Plan 10 adds the provenance-bound corpus record as corroboration."
  - "Preserve the unfiltered full-suite and timing-script commands at their honest WIRED/REACHABLE ceilings with owned Phase 20 row exemptions, rather than replacing them with a narrower test command."
  - "Leave D-13-34 open and count it in the five-item population before the human decision."
requirements-completed: [QLT-10, PRC-02]
coverage:
  - id: I1
    description: Maturity counts, public emitter call positions, and new skip witnesses match the current source tree.
    requirement: QLT-10
    verification:
      - kind: unit
        ref: internal/compiler/session.TestLanguageMaturityCountsAreCurrent, TestPhase16PublicEmitterConsumerInventory, TestNoSuppressionOutlivesItsWitness
        status: pass
    human_judgment: false
  - id: I2
    description: The groundedness scan, pinned R1/R2/R3 frontier, and reconciliation ledger agree after Plan 02 and Phase 20 updates.
    requirement: QLT-10
    verification:
      - kind: unit
        ref: internal/compiler/session.TestVerificationGroundedness, TestVerificationGroundednessFrontierIsPinned, TestVerificationGroundednessThreeClassesAreEmpty, TestReconciliationVerdictsCarryTheirObligations
        status: pass
    human_judgment: false
  - id: D1
    description: The live source-derived open-unowned set is exactly five, including the pending D-13-34 decision.
    requirement: PRC-02
    verification:
      - kind: unit
        ref: internal/compiler/session.TestPhase20UnownedDebtPopulation, TestPhase20UnownedDebtCapRule
        status: pass
    human_judgment: false
duration: pending serialized closeout
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 07: Reconcile the Live Validation Frontier

The checked-in corpus now matches the source: 141 programs and 4,639 lines, the emitter registry points to the current line 501 call, and all six Phase 20/native skips resolve to named witnesses. Phase 08 and 11 archive-row findings were closed against Plan 02's corrected records, leaving 54 exact live reconciliation entries. The groundedness scan and its frontier tests pass.

## Accomplishments

- Added `Grade` and `Non-inertness` cells to the validated Phase 17–19 maps and completed the Phase 20 map's explicit no-twin cells.
- Updated four Phase 14 row-debt dispositions using the repaired exact commands and direct passing tests; the complete provenance-bound corpus record is deferred to Plan 10.
- Removed the stale Phase 08/11 reconciliation obligations, regenerated the reconciliation view from 66 to 54 current entries, and regenerated the unreachable-claims view from the source debt register.
- Closed D-14-70/71/81–90 against Plan 02's archive-row correction commit `c4278ea`.
- Reduced the live open-unowned set from nine to five without using the D-13-34 decision.

## Verification

- `go test` focused Phase 12, Phase 04, and Phase 06 named-test replacements — passed.
- `TestLanguageMaturityCountsAreCurrent`, `TestSelfDescribingDocsGuardIsNotInert`, `TestPhase16PublicEmitterConsumerInventory`, `TestNoSuppressionOutlivesItsWitness` — passed.
- `TestVerificationGroundedness`, `TestVerificationGroundednessFrontierIsPinned`, `TestVerificationGroundednessThreeClassesAreEmpty`, `TestReconciliationVerdictsCarryTheirObligations` — passed.
- `TestPhase20ValidationLifecycle`, `TestPhase20UnownedDebtPopulation`, `TestPhase20UnownedDebtCapRule`, `TestDebtRegistersAreWellFormed`, `TestEvidenceReconciliationViewIsCurrent`, `TestUnreachableClaimsView` — passed.
- `verify plan-structure` for Phase 20 Plans 07 and 10 — valid with no structural warnings.

## Commits

- `32ef547` — Refresh Phase 20 source inventory witnesses.
- `e516fce` — Reconcile Phase 20 validation debt and frontier.

Plan 10 now produces the final run record from the exact post-reconciliation 32-pair export and verifies the archived grade scan before Phase 08 timing begins.

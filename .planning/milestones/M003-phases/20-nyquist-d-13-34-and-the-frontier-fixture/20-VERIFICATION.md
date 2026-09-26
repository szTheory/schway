---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
verified: 2026-09-25
status: passed
score: 4/4 requirements
---

# Phase 20 Verification

## Goal

The measuring corpus reconciles against the language surface M003 built, and the refusal frontier is pinned as moved.

## Requirement verdicts

| Requirement | Verdict | Evidence |
|---|---|---|
| QLT-10 — honest, reconciled validation evidence | PASS | Plans 02, 03, 07, 09, and 10 reconciled the validation rows and exact frontier; the 33-pair archive passed `TestValidationRowGradesAreEarnedOverArchivedCorpus`. |
| QLT-11 — checksum frontier moves | PASS | Plan 01's `examples/checksum.lang` refusal is pinned and compared against the independently reproduced M003-open diagnostic; see `20-01-SUMMARY.md`. |
| QLT-12 — content-addressed closure and timing | PASS | Plan 04's unchanged 112-program closure reuses evidence and seeded input changes force recomputation; Plan 08 records three paired cold/warm runs and its unfiltered preflight. |
| PRC-02 — live open-unowned debt cap | PASS | `TestPhase20UnownedDebtPopulation` and `TestPhase20UnownedDebtCapRule` pass; the live set is four IDs. |

## Additional phase truths

- All ten plans have completion summaries; all 25 automated validation rows are marked passed.
- The user's blocking-human choice was `replace`. The move and borrow fixture summaries now differ from their derivation pairs under the unchanged five-component predicate, and both injectors plus the end-to-end repair-driver control pass.
- D-13-34 is `CLOSED(4658eb1)`. D-06-29 structural inference is restored for move and borrow, with the small-corpus limitation retained.
- `GOCACHE=/private/tmp/phase20-gocache GOMAXPROCS=4 go test ./... -count=1` passed across all packages after the implementation and evidence-pin updates.

## Result

Phase 20 meets its goal and all four requirements. The checksum frontier, archive, closure cache, timing evidence, generated views, debt population, selected D-13-34 outcome, and repository test suite are verified.

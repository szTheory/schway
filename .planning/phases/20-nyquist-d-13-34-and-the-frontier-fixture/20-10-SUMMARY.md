---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "10"
subsystem: validation-evidence
tags: [qlt-10, prc-02, archived-validation, corpus-record]
requires:
  - phase: 20-07
    provides: reconciled validation rows, debt ownership, and exact live frontier
  - phase: 20-09
    provides: repaired archived row citations and exact corpus inputs
provides:
  - Complete 33-pair archived validation corpus record and digest manifest
  - Corpus-backed corroboration for four archived debt closures
  - Confirmed current generated unreachable-claims view
affects: [20-08, 20-06, QLT-10, PRC-02]
actuals:
  tasks: 2
  pairs: 33
  recorder_elapsed: 142.108s
  failed_pairs: 0
plan_head_before: e359f17
commits: [01485e5, aa13977]
key-files:
  created: []
  modified:
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-VALIDATION.md
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-10-PLAN.md
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-PLAN-CHECK.md
key-decisions:
  - "The current corpus contains 33 exact package-pattern pairs; all completed without failures, and the manifest binds raw pair and record bytes."
  - "The exact archived grade scan passes. Four repaired row closures cite the complete run record at revision e359f17."
  - "The generated unreachable-claims view remained byte-current after the debt-detail corroboration was added."
verification:
  - "GOCACHE=/private/tmp/phase20-gocache go test ./internal/compiler/session -run '^TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1 — passed."
  - "GOCACHE=/private/tmp/phase20-gocache go test ./internal/compiler/session -run 'TestDebtRegistersAreWellFormed|TestReconciliationVerdictsCarryTheirObligations|TestUnreachableClaimsView|TestValidationRowGradesAreEarnedOverArchivedCorpus' -count=1 — passed."
  - "git diff --check — passed."
status: complete
completed: 2026-09-25
---

# Phase 20 Plan 10: Record the Exact Validation Corpus

The final post-reconciliation archive contains 33 exact package-pattern pairs. After the full-suite gate exposed a stale M004 test callsite, the provenance guard was aligned with the live `EmitNative` call and the corpus was regenerated. The final sequential recorder completed in 142.108 seconds with 33 pair-completion sentinels, one batch-completion sentinel, and no test or build failures. The manifest records source revision `aa13977`, the raw pair-export SHA-256 `580587ae5b1763daf70b774a5dc862e076f023ec3674a1e22e639355285c3465`, and run-record SHA-256 `a1db41c803d69ceb7650f0adfce818c532f953bf664e8674e509eb57814c8a63`.

## Accomplishments

- Replaced the older corpus artifact with the complete run record and updated its provenance manifest.
- Corrected validation row 20-04-01 to name the exact targeted native/cache test pattern, then regenerated the corpus from that row set.
- Aligned the M004 provenance guard with the refreshed emitter inventory's live `EmitNative` callsite at line 501.
- Added record revision and digest corroboration to D-14-50/51/52/54.
- Confirmed the unreachable-claims view remains byte-identical to the test renderer after the debt detail updates.
- Updated the Phase 20 validation map and plan-check inventory for this plan's actual six-file scope.

## Verification

- Archived validation grade scan passed against the full checked-in run record.
- Debt well-formedness, reconciliation obligations, generated-view equality, and archived grade scan passed together.
- The corpus artifact and planning/debt updates were committed as `01485e5`.

## Next

Plan 08 owns the unfiltered full-suite preflight and three paired cold/warm timing samples. Plan 06 remains the final human-gated D-13-34 checkpoint; prepare its measured alternatives after timing, then stop for the user's choice.

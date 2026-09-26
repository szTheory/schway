---
phase: 21-native-emission-ownership-and-resource-discharge-m004
plan: 04
subsystem: compiler-emission-evidence
tags: [go, cgen, resource-discharge, clang, lto, validation]
requires:
  - phase: 21-01
    provides: Machine-checkable resource-discharge contract and mutation checks
  - phase: 21-02
    provides: Retired legacy emitter bodies and AST refusal guard
  - phase: 21-03
    provides: Scoped emitted-fixture LTO comparison and evidence receipt
provides:
  - Historical emitter and LTO debt records aligned with executed evidence
  - Derived unreachable-claims view refreshed from source registers
  - Validation corpus record refreshed from the exact consumer-derived 33 package/pattern pairs
  - Phase 21 acceptance rows graded and backed by automated checks
affects: [phase-21-verification, phase-planning, emitter-debt, validation-evidence]
actuals:
  tokens: 12500
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [Evidence receipts stay bounded to measured fixture/compiler/host facts, generated views are re-derived from source registers]
key-files:
  created: []
  modified:
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md
    - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VALIDATION.md
    - .planning/UNREACHABLE-CLAIMS.md
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json
key-decisions:
  - "The LTO evidence supports semantic equality for one named fixture on Darwin arm64 with Apple Clang 21.0.0; it does not establish optimizer activity, performance, cleanup, or other host behavior."
  - "The recurring CI witness binds the tagged comparison to its receipt; the compiler comparison remains a one-shot cost lane."
  - "Historical foreign and by-pointer emitter debt remains family-specific and refused pending its own discharge proof and runtime evidence."
patterns-established:
  - "Refresh archived evidence records only from the validation consumer's exported package/pattern pairs, then bind both pair and raw-record digests."
requirements-completed: []
coverage:
  - id: D1
    description: Historical emitter debt, scoped LTO claims, and derived unreachable claims match current implementation and recorded evidence.
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestDebtRegistersAreWellFormed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestUnreachableClaimsViewIsCurrent"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase16EmitterCutsAreAmendedAndOwned"
        status: pass
      - kind: unit
        ref: "go test ./... -count=1 -p 2"
        status: pass
    human_judgment: false
  - id: D2
    description: Phase 21 acceptance rows and archived validation grades are machine-checked and earned.
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestValidationRowGradesAreEarnedOverArchivedCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestVerificationGroundednessThreeClassesAreEmpty"
        status: pass
    human_judgment: false
duration: 38min
completed: 2026-09-26
status: complete
---

# Phase 21 Plan 04 Summary

**Emitter debt, generated claims, and CI evidence now reflect the retired code and the exact scope of the measured LTO comparison.**

## Performance

- **Duration:** about 38 minutes
- **Started:** 2026-09-26T02:12:00Z
- **Completed:** 2026-09-26T02:50:32Z
- **Tasks:** 2
- **Files modified:** 10 in the final task commit, plus the summary

## Accomplishments

- Updated D-11-02, D-12-36, D-16-11..13, and D-14-45 to cite the actual retirement probe, preserve separate refusal gates, and bound LTO claims to the single recorded fixture/compiler/host.
- Re-derived `.planning/UNREACHABLE-CLAIMS.md` and refreshed Phase 21 validation status and the archived validation corpus record using all 33 consumer-derived package/pattern pairs.
- Added recurring receipt-binding and evidence-pin checks; the one-shot compiler comparison remains out of routine CI.

## Task Commits

1. **Task 1: Reconcile retired emitter debt** — `1397863` (docs/test)
2. **Task 2: Record scoped evidence and refresh claims** — `4c48228` (docs/test)

Plan metadata will be committed with this summary.

## Files Created/Modified

- `.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VALIDATION.md` — all six acceptance rows are exercised and pass.
- `.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESEARCH.md` — removed broad command-like claims from prose and made the contract probe precise.
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` and `.planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md` — evidence and family-specific reopening gates.
- `.planning/UNREACHABLE-CLAIMS.md` — generated view matching the updated debt registers.
- `testdata/phase16/validation-corpus-run-record.jsonl` and its manifest — refreshed complete consumer-derived corpus record.
- `internal/compiler/session/witness_registry_test.go` — corrected the pinned source location for the refusal witness.

## Decisions Made

- Kept the recurring structural and receipt-binding checks in standard CI while recording the Clang comparison as a bounded one-shot experiment.
- Narrowed the historical Phase 4 witness to the exact surviving foreign resource-ledger refusal test, avoiding a compound command that obscured its acceptance behavior.
- Preserved the foreign, exclusive by-pointer, and shared by-pointer future proof gates independently; no emitter family was admitted.

## Deviations from Plan

### Auto-fixed Issues

**1. Stale downstream evidence pins after emitter retirement**
- **Found during:** final plan verification
- **Issue:** Static checks still expected the old source line and old `-flto consequence` marker; a historic validation row cited a broad compound command, and new prose was misclassified by the groundedness parser.
- **Fix:** Updated the exact line pin and required debt marker, narrowed the witness command, and made research/validation prose unambiguous.
- **Verification:** Focused provenance, amendment, frontier, and class-empty tests passed.
- **Committed in:** `4c48228`

**2. Validation evidence producer inherited the wrong Go cache on the first attempt**
- **Found during:** corpus refresh
- **Issue:** The shell producer's child `go test` processes used the default cache and emitted sandbox cache-trim errors, making that first record invalid.
- **Fix:** Re-ran all 33 pairs with `GOCACHE` explicitly exported to each child, rebound the manifest to the exact exported pair bytes and completed record, then removed the failed attempt's stderr sidecar.
- **Verification:** Archived-grade consumer passed and confirmed all validation rows remain earned; full repository suite passed.
- **Committed in:** `4c48228`

**Total deviations:** 2 auto-fixed. Both were necessary to keep recurring evidence and source pins trustworthy.

## Issues Encountered

- A shell wrapper tried to assign zsh's read-only `$status` variable after a complete green suite. The suite was run directly afterward and returned exit code 0.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 21-04's source, evidence, and validation artifacts are complete; the full Go suite passes.
- No human UAT criterion remains for this plan. Phase-level automated verification is the remaining workflow closeout.

---
*Phase: 21-native-emission-ownership-and-resource-discharge-m004*
*Completed: 2026-09-26*

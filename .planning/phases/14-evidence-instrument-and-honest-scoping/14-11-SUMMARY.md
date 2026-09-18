---
phase: 14-evidence-instrument-and-honest-scoping
plan: 11
subsystem: testing
tags: [go, evidence-instrument, run-record, groundedness-lint, qlt02, reconciliation]

# Dependency graph
requires:
  - phase: 14-evidence-instrument-and-honest-scoping (plans 01-10)
    provides: the run-record substrate (evidence_grade_test.go), the static
      test index and groundedness lint (verification_groundedness_test.go),
      the reconciliation ledger and PHASE-14-DEBT.md apparatus
provides:
  - a completion-witnessed run record (per-pair and per-batch sentinels)
    that cannot report coverage it did not finish producing
  - a measured-and-asserted timing budget (evidenceRunRecordTimeout=480s,
    evidenceRunRecordMarginFraction=0.75) that fails closed on a
    near-timeout run instead of silently degrading to a lower grade ceiling
  - an anchored producer/consumer name contract (resolvedPkgPatterns) so
    the run-record batch executes exactly what deriveCeiling consults
  - a named self-citation refusal (evidenceRunRecordSelfCitedTests)
affects: [14-12, 14-13]

actuals:
  tokens: 10040
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Completion-witness sentinels with an Action vocabulary disjoint from
      the tool being wrapped, so a parser can never confuse a witness with
      a result"
    - "Shared admissibility predicate (runRecordAdmissible) consulted by
      both the production fatal gate and its own non-inertness proof"
    - "Producer batches built from resolved, anchored names instead of raw
      pattern text, so a runtime tool invocation cannot substring-match
      wider than the static resolution already computed"

key-files:
  created: []
  modified:
    - internal/compiler/session/evidence_grade_test.go
    - scripts/evidence-run-record.sh
    - internal/compiler/session/qlt02_budget_manifest.json
    - internal/compiler/session/session_phase6_budget.go
    - internal/compiler/session/verification_groundedness_test.go
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/EVIDENCE-RECONCILIATION.md

key-decisions:
  - "evidenceRunRecordTimeout raised 300s -> 480s with its doc comment
    restated to the actually-measured 269s-347s range (commit 0dcb460),
    replacing a stale ~136s claim WR-01 flagged"
  - "evidenceRunRecordMarginFraction=0.75: a run consuming more than three
    quarters of its own deadline is a named budget failure, never a grade"
  - "resolvedPkgPatterns replaces raw-pattern batching for the go-test-run
    shape; the assert-go-tests.sh shape stays routed through pkgPatternsFor
    unchanged (it already anchors correctly)"
  - "Two pre-existing, unrelated gaps in already-committed Phase 14
    documents (a stale non-inertness-spot-check citation in
    14-VERIFICATION.md, and three new unparseable prose fragments in this
    plan's own PLAN.md/VERIFICATION.md) were reconciled/pinned rather than
    left broken, since they blocked this plan's own required `go test
    ./...` green and the fix is the same additive, low-risk mechanism this
    phase built for exactly this purpose (Rule 3)"

patterns-established:
  - "A corpus-wide fail-closed test caches its failure reason in a package
    var alongside sync.Once, so every caller (not just the first) raises
    the same t.Fatalf instead of silently returning a partial result"

requirements-completed: [EVD-02]

coverage:
  - id: D1
    description: "A run record that did not finish covering every requested (package, pattern) pair is a named hard failure, not a silent WIRED-ceiling fallback"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestRunRecordCompletenessGuardIsNotInert"
        status: pass
    human_judgment: false
  - id: D2
    description: "The corpus-wide run record's measured wall-clock is recorded as a number in the tree, and a run consuming more than 75% of its timeout fails with a named budget message"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestRunRecordCompletenessGuardIsNotInert"
        status: pass
      - kind: other
        ref: "go test ./internal/compiler/session/ -run 'TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1 -v (measured elapsed 1m30.301553708s, logged and recorded in qlt02_budget_manifest.json)"
        status: pass
    human_judgment: false
  - id: D3
    description: "The run-record producer executes exactly the top-level test names the consumer will consult, anchored, so no cited pattern can substring-match an unrelated top-level test in the same package"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestEvidencePatternsResolveToAnchoredNames"
        status: pass
    human_judgment: false
  - id: D4
    description: "An evidence cell citing TestValidationRowGradesAreEarnedOverArchivedCorpus is refused by name as self-certification"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestEvidencePatternsResolveToAnchoredNames/a_cell_citing_the_corpus-wide_test_itself_is_refused_as_a_self-citation"
        status: pass
    human_judgment: false
  - id: D5
    description: "Each new guard in this plan has a seeded fault that makes it go red and an unmodified control that stays green"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestRunRecordCompletenessGuardIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestEvidencePatternsResolveToAnchoredNames"
        status: pass
    human_judgment: false

duration: 40min
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 11: Completion-Witnessed Run Record and Anchored Producer/Consumer Contract Summary

**A run record can no longer claim coverage it did not finish producing, a near-timeout run is refused rather than graded, and the run-record producer now executes exactly the anchored, resolved names the grader consults -- closing WR-01 and WR-02 from 14-REVIEW.md and dropping the corpus-wide run's measured cost from 269-347s to ~90s.**

## Performance

- **Duration:** ~40 min
- **Started:** 2026-09-18T13:18:00Z
- **Completed:** 2026-09-18T13:56:36Z
- **Tasks:** 3/3
- **Files modified:** 7

## Accomplishments

- Wired a completion witness (per-pair and per-batch JSON-line sentinels,
  `record_pair_complete`/`record_batch_complete`) end to end through
  `scripts/evidence-run-record.sh` and `evidence_grade_test.go`'s
  `runRecord`/`complete()`/`generateRunRecord`, proven by a real two-pair
  batch (`TestRunRecordCarriesACompletionWitness`).
- Raised `evidenceRunRecordTimeout` to the measured-honest 480s and added
  `evidenceRunRecordMarginFraction` (0.75); `corpusRunRecord` now
  `t.Fatalf`s on an incomplete OR near-timeout record via a shared
  `runRecordAdmissible` predicate, on every caller (cached failure reason,
  not swallowed inside `sync.Once`).
- Replaced raw-pattern batching with `resolvedPkgPatterns`: the run-record
  producer now executes an anchored alternation of names resolved through
  the SAME static-index matching `deriveCeiling` itself performs, and a
  self-citing evidence cell is refused by name
  (`evidenceRunRecordSelfCitedTests`/`selfCitationError`).
- Corpus-wide run dropped from 269-347s (commit 0dcb460's measured range,
  D-14-121's 300.11s near-miss) to a measured **90.30s**, recorded as
  `evidence_run_record_wall_clock_ns` in `qlt02_budget_manifest.json`
  (9.4x margin under the 480s budget).

## Task Commits

Each task was committed atomically:

1. **Task 1: End-to-end completion witness for one run-record batch** - `d81a991` (feat)
2. **Task 2: An incomplete or margin-less run record fails by name, and the timeout tells the truth** - `32bb781` (feat, tdd)
3. **Task 3: Producer and consumer consult one anchored name set; self-citation is refused** - `2dd7cd8` (feat, tdd; includes two Rule 3 auto-fixes, see Deviations)

**Plan metadata:** (this commit) `docs(14-11): complete completion-witnessed run record plan`

_Note: Tasks 2 and 3 carry `tdd="true"` in the plan but each landed as a single commit -- the new tests (`TestRunRecordCompletenessGuardIsNotInert`, `TestEvidencePatternsResolveToAnchoredNames`) were authored, run RED against the not-yet-existing predicate/function to confirm they fail for the right reason, then made GREEN in the same edit pass before committing; no separate RED commit was created since the plan did not require one and the behavior was net-new (not a refactor of already-green code)._

## Files Created/Modified

- `scripts/evidence-run-record.sh` - emits `record_pair_complete`/`record_batch_complete` JSON-line sentinels after each pair and the whole batch, via discrete `printf` argv (no `eval`, no shell-string interpolation)
- `internal/compiler/session/evidence_grade_test.go` - `runRecord.{coveredPairs,batchComplete,requestedPairs,elapsed,complete()}`, `runRecordAdmissible`, `resolvedPkgPatterns`, `evidenceRunRecordSelfCitedTests`/`selfCitationError`, `patternNames`, raised timeout/margin constants, three new tests (`TestRunRecordCarriesACompletionWitness`, `TestRunRecordCompletenessGuardIsNotInert`, `TestEvidencePatternsResolveToAnchoredNames`)
- `internal/compiler/session/qlt02_budget_manifest.json` - new `evidence_run_record_wall_clock_ns` observed row (90,301,553,708 ns)
- `internal/compiler/session/session_phase6_budget.go` - widened `QLT02MetricVocabulary` to admit the new metric name (Rule 3 fix, Task 3)
- `internal/compiler/session/verification_groundedness_test.go` - `pinnedFrontier` gains three unparseable prose-fragment findings from this plan's own PLAN/VERIFICATION docs (Rule 3 fix, Task 3)
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` - new debt row D-14-122 reconciling a stale R1 citation in `14-VERIFICATION.md:133` (Rule 3 fix, Task 3)
- `.planning/EVIDENCE-RECONCILIATION.md` - regenerated to match PHASE-14-DEBT.md's new reconciliation entry

## Decisions Made

- Doing BOTH the timeout raise (480s) and the batch cost reduction
  (anchored resolved names) rather than either alone, per the plan's own
  reversibility note -- the timeout raise alone could not have bought
  margin under Go's 600s per-package ceiling, and the measured result
  (90.30s, 18.8% of budget) needed both.
- Kept `pkgPatternsFor`'s `assert-go-tests.sh` branch byte-unchanged and
  routed it through `resolvedPkgPatterns` rather than re-deriving a second
  resolver, per the plan's explicit read_first constraint.
- `consolidatePkgPatterns` keeps wrapping each branch in its own parens
  (never one outer `^(...)$`) -- documented in its doc comment as
  load-bearing (breaks `classifyPerBranchGroundedness`'s per-branch
  splitter otherwise), not cosmetic.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Stale R1 citation in 14-VERIFICATION.md blocked TestReconciliationVerdictsCarryTheirObligations and cascaded into TestValidationRowGradesAreEarnedOverArchivedCorpus**
- **Found during:** Task 3, while re-measuring the corpus-wide elapsed after landing the anchoring fix
- **Issue:** `14-VERIFICATION.md:133`'s Non-Inertness Spot-Check table narrates an already-reverted fault-injection perturbation (`TestThisNameDoesNotExistAnywhereZZQQ`, deliberately fabricated, injected then `git checkout --`-reverted per that same row). The groundedness lint's full-corpus scan (widened past plan 14-10's own reconciliation pass, which predates this document's addition) correctly flagged the citation as an unreconciled R1 finding. This made `TestReconciliationVerdictsCarryTheirObligations` fail, which cascaded: `14-VALIDATION.md` row 14-10-T1's own evidence cites that same test, so `TestValidationRowGradesAreEarnedOverArchivedCorpus` derived `WIRED` instead of `EXERCISED` -- correctly, since the cited test genuinely failed for real when the run record actually finished (this plan's whole point).
- **Fix:** Added `PHASE-14-DEBT.md` row D-14-122, reconciling the citation as `renamed` to the real, resolving test the row's claim actually rests on (`TestVerificationGroundednessFrontierIsPinned`); regenerated `.planning/EVIDENCE-RECONCILIATION.md` to match.
- **Files modified:** `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md`, `.planning/EVIDENCE-RECONCILIATION.md`
- **Verification:** `go test ./internal/compiler/session/ -run 'TestReconciliationVerdictsCarryTheirObligations$|TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1 -v` exits 0
- **Committed in:** `2dd7cd8` (part of Task 3 commit)

**2. [Rule 3 - Blocking issue] New evidence_run_record_wall_clock_ns metric rejected by the closed QLT-02 metric vocabulary**
- **Found during:** Task 3, after adding the manifest row
- **Issue:** `QLT02MetricVocabulary()` is a closed list; the new metric name I added was outside it, failing `TestBudgetAuditRefusesUndeclaredMachine`, `TestQLT02InterproceduralGrowthExponent`, and (cascading) `TestVerifyPhase6ControlsAndWork`.
- **Fix:** Widened `QLT02MetricVocabulary()` to admit `evidence_run_record_wall_clock_ns`, following the exact precedent set when `suite_wall_clock_ns` was added.
- **Files modified:** `internal/compiler/session/session_phase6_budget.go`
- **Verification:** the three named tests exit 0
- **Committed in:** `2dd7cd8` (part of Task 3 commit)

**3. [Rule 3 - Blocking issue] Three new unparseable prose fragments in this plan's own PLAN.md/VERIFICATION.md broke the pinned groundedness frontier**
- **Found during:** Task 3, during the final `go test ./...` sweep
- **Issue:** `TestVerificationGroundednessFrontierIsPinned` requires exact set equality between the pinned frontier and a fresh measurement. This plan's own `14-11-PLAN.md` (trust-boundary and artifacts table cells naming `go test`/`go test -run` as bare prose, not literal invocations) and `14-VERIFICATION.md:86` (a prose cell describing what the lint detects) were already committed to the tree before this plan started and were never pinned -- pre-existing drift, unrelated to this plan's code, but blocking the required `go test ./...` green.
- **Fix:** Added the three new findings to `pinnedFrontier`, matching the exact shape of the existing STACK.md/SUMMARY.md/14-RESEARCH.md unparseable-prose entries already pinned there.
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `TestVerificationGroundednessFrontierIsPinned` exits 0
- **Committed in:** `2dd7cd8` (part of Task 3 commit)

---

**Total deviations:** 3 auto-fixed (all Rule 3 - blocking issue)
**Impact on plan:** All three were pre-existing gaps in already-committed Phase 14 artifacts (two documents, one code vocabulary), surfaced only because this plan's own fail-closed margin check now actually finishes and grades for real instead of silently degrading. None touched this plan's own core deliverable (the completion witness, margin gate, or anchored name contract); all were additive, low-risk corrections using the exact mechanisms (reconciliation ledger, pinned frontier, metric vocabulary) this phase already built for this purpose. No scope creep.

## Issues Encountered

Two synthetic test scenarios needed rework after their first draft failed for reasons unrelated to the code under test:

- Task 1's tracer test initially considered `internal/compiler/testsupport` (CLI-heavy, slow) before settling on `internal/compiler/core` (~1s) as the cheap two-pair fixture package.
- Task 3's "prefix-related patterns consolidate" subtest initially tried to parse the CONSOLIDATED (multi-branch-joined) pattern with the single-level `patternNames` helper, which only handles one `^(...)$` layer; rewritten to compile the consolidated pattern as a real regexp and assert what it actually matches/does not match, which is also the more faithful check (it verifies what the run-record script itself would do with that exact string).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plans 14-12 (PRC-01 exemption removal) and 14-13 depend on this plan's timeout raise and cost reduction landing first -- both now hold: the corpus-wide run finishes at ~90s (18.8% of the 480s budget), comfortably clear of D-14-121's 300.11s near-miss and the margin check that would now refuse it if it weren't. No blockers.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

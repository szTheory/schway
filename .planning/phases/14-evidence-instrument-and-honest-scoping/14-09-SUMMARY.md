---
phase: 14-evidence-instrument-and-honest-scoping
plan: 09
subsystem: evidence-instruments
tags: [go, testing, go-parser, evd-02, grade-cap, groundedness-lint, debt-register]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-01's static test index (buildTestIndex/testIndex) as the sole resolution primitive; plan 14-06's Status/Grade dual-keying on Tier-A detection (verdictColumnIndex), which is why this plan can retire the Status column without leaving any document briefly undetectable; plan 14-07's Grade/Witness closed vocabulary and witness-token grammar for *-DEBT.md registers (debtRegisterGrades, debtRegisterWitnessTokenProblem, the generated UNREACHABLE-CLAIMS.md view)"
provides:
  - "internal/compiler/session/evidence_grade_test.go (new file): the closed five-member grade vocabulary and total order, the derivation ladder (deriveCeiling) covering go-test -run patterns, the M001-era assert-go-tests.sh exact-name-list shape, compile-time evidence, callsite: tokens, and bare command-line invocations, the fail-closed run-record consumer, the corpus-wide production test TestValidationRowGradesAreEarnedOverArchivedCorpus, the file-scoped dated bar exemption map, and a three-fault non-inertness proof"
  - "scripts/evidence-run-record.sh (new file): the sequential run-record producer, documented as deliberately not parallel after a measured thrash regression"
  - "all fourteen *-VALIDATION.md documents migrated: Status retired, Grade + Non-inertness columns added, frontmatter carries evidence_vocabulary/graded_rows"
  - "PHASE-14-DEBT.md grows from 3 to 10 items: six rows recording EVD-02's first real findings (rows that derive below their shipped ✅ green verdict) plus one recording 14-VALIDATION.md's own unfilled template"
  - "the groundedness lint's pinned frontier re-measured over the migrated corpus (125 -> 124 entries, fully attributed)"
affects: ["14-10 (Nyquist reconciliation inherits two of this plan's six findings -- 09-VALIDATION.md:87/:95 -- as members of the same R2 frontier it already owns closing)", "any future plan editing a *-VALIDATION.md or PHASE-14-DEBT.md"]

actuals:
  tokens: 74200
  tasks: 3
  commits: 10

tech-stack:
  added: []
  patterns:
    - "Declared-grade-capped-by-derived-ceiling: the author writes a Grade cell, a Go test derives a ceiling from the row's own evidence cell, and declared > derived is a hard failure -- never a string-membership check alone"
    - "Two authored laws, never merged: *-VALIDATION.md row grades (this plan, EVD-02) and *-DEBT.md row grades (plan 14-07, EVD-03/04) share vocabulary text but never share package vars, so each law's own evolution stays independent"
    - "Fail-closed run-record consumption: absence of a run record caps every ceiling below EXERCISED rather than defaulting to a pass; the corpus-wide production test memoizes ONE real run-record generation per test binary via sync.Once"
    - "File-scoped, dated bar exemptions (never row-scoped), copying debtRegisterLandingPhaseExemptions' shape exactly"
    - "Below-shipped-verdict findings are recorded as new debt rows with a real witness probe, never suppressed or rewritten into the archived document"

key-files:
  created:
    - internal/compiler/session/evidence_grade_test.go
    - scripts/evidence-run-record.sh
  modified:
    - internal/compiler/session/verification_groundedness_test.go
    - internal/compiler/session/session_test.go
    - internal/compiler/session/witness_registry_test.go
    - .planning/milestones/M001-phases/01-canonical-pure-spine/01-VALIDATION.md
    - .planning/milestones/M001-phases/02-owned-values-and-abilities/02-VALIDATION.md
    - .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md
    - .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md
    - .planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-VALIDATION.md
    - .planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-VALIDATION.md
    - .planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-VALIDATION.md
    - .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md
    - .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
    - .planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-VALIDATION.md
    - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
    - .planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md
    - .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
    - .planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/UNREACHABLE-CLAIMS.md

key-decisions:
  - "Reused plan 14-01's static test index and plan 14-07's callsite-witness/exact-match machinery directly (buildTestIndex, debtRegisterCallsiteTokenPattern, debtRegisterCallSiteWitnessMatches) rather than re-deriving a second resolver -- the plan's own key_links obligation."
  - "Grade is authored, never derived automatically above EXERCISED: every migrated row's declared Grade was set to the derived ceiling itself (capped at EXERCISED, never MUTATION-KILLED), with Non-inertness uniformly '—'. This is a deliberate, disclosed migration-authoring choice: proving a genuine mutation-killed twin for ~230 rows was out of this plan's budget, and declaring at the ceiling (never above it) is always honest by construction, never an overclaim."
  - "Run-record generation for the ~230-row corpus is genuinely expensive: a bounded-parallelism attempt was tried and measured to THRASH (oversubscribing already-internally-parallel `go test ./internal/compiler/...` invocations, individual jobs stalling at ~0% CPU for tens of seconds) rather than speed up. Reverted to sequential; documented in the script's own header. Measured cost: several minutes for the full corpus, correcting the threat model's (T-14-56) original 'milliseconds' assumption -- disclosed, not hidden."
  - "Migration authoring used TWO successive real run-record generations (a ~230-row full-corpus pass, then a ~65-row targeted recheck of every row that initially derived below EXERCISED) rather than one, because the first pass's own run-record generation was itself truncated by an operational timeout mid-batch, understating which rows actually resolve. The recheck used a from-scratch, independently-run, complete record and is the one migration Grade values are authored from."
  - "12-VALIDATION.md's primary table (07-VALIDATION.md's Req-ID-keyed 6-column shape and the other 13 documents' canonical 10-column shape) is the ONLY table migrated per document. Three secondary supplement sub-tables under the same '## Per-Task Verification Map' heading -- 02-VALIDATION.md's two 'Post-Gate Supplement' tables and 11-VALIDATION.md's 'Requirement -> Test Map' table -- are explicitly out of this migration's scope and retain their original Status column. Recorded here, not silently limited."
  - "Discovered a pre-existing malformed row (01-VALIDATION.md rows 01-02-02 through 01-03-02, 3 of 8 rows) missing one column (Threat Ref) relative to the canonical 10-column header. Not fixed -- fixing would require inventing the missing value, and the plan's own constraint forbids altering any pre-existing cell. Graded DEFINED (the safest floor) since the Automated Command cell's real position is ambiguous for these three rows; documented here as a Rule-1-adjacent finding rather than silently normalized."
  - "Task 2's real corpus-wide grade derivation (TestValidationRowGradesAreEarnedOverArchivedCorpus) surfaced THREE more genuine dead-citation findings beyond the two anticipated ones (matching plan 14-01/14-06's already-pinned R2 frontier) once the ladder ran with a COMPLETE run record: 04-VALIDATION.md:74 and TWO rows in 06-VALIDATION.md (not one) cite tests retired or renamed in later phases. All three are recorded as debt (D-14-51, D-14-52, D-14-54) with a shared witness probe (TestValidationGradeCapArchivedDeadCitationsRemainAbsent) rather than silently re-graded upward or hidden."
  - "Two regressions in EXISTING tests were exposed by this plan's own edits and fixed as Rule 1 auto-fixes, not silently worked around: session_test.go's debtRegisterEmptyTableFixture hardcoded a stale 'items: 3' literal that PHASE-14-DEBT.md's own growing item count (3 -> 10) made a silent no-op; and .planning/UNREACHABLE-CLAIMS.md (plan 14-07's generated view) needed regeneration since six of this plan's seven new debt rows carry a probe: witness token and now qualify for that view."
  - "Two OTHER pre-existing tests (TestVerificationGroundedness's hardcoded 'must contain' violations, TestVerificationGroundednessGrepOverRealCorpus's hardcoded archival-breakage line number) carried hardcoded (file, line) expectations that plan 14-09's own frontmatter migration shifted by +2, beyond the pinnedFrontier literal itself. Found only by running the full TestVerificationGroundedness family, not assumed sufficient from the frontier re-pin alone."

patterns-established:
  - "A migration that touches many archived documents' line numbers must search for every OTHER hardcoded (file, line) assertion across the test suite, not just the one pinned-frontier literal that was expected to need re-measuring."
  - "When authoring ~200+ Grade values from a live run record, generate the record via a targeted, deduplicated batch of exactly the (package, pattern) pairs the rows cite -- never a bare `go test ./...` -- and verify the batch actually completed (no internal timeout truncation) before trusting any 'not confirmed' result as a real finding rather than a measurement gap."

requirements-completed: [EVD-02]

coverage:
  - id: D1
    description: "Every Per-Task Verification Map row carries a grade from the closed five-member vocabulary (DEFINED/WIRED/REACHABLE/EXERCISED/MUTATION-KILLED), asserted as a total order; an out-of-vocabulary or empty/absent Grade cell fails outright, never defaulted."
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeVocabularyIsClosed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeOrderIsTotal"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationRowGradesAreEarned"
        status: pass
    human_judgment: false
  - id: D2
    description: "A declared grade is mechanically capped by a ceiling derived from the row's own evidence cell -- across go-test -run patterns, the M001-era assert-go-tests.sh exact-name-list shape, compile-time evidence, callsite: tokens, and bare command-line invocations -- and the gate fails closed when EXERCISED-or-above is declared with no run record present."
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationRowGradesAreEarned"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeCapIsNotInert"
        status: pass
    human_judgment: false
  - id: D3
    description: "Derivation is total over all fourteen archived-and-live *-VALIDATION.md documents (every row classifies; none unclassifiable), and the >= EXERCISED satisfying bar is enforced only for phase 14 onward via a file-scoped, dated exemption map -- never row-scoped, mechanically asserted."
    requirement: "EVD-02"
    verification:
      - kind: integration
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationRowGradesAreEarnedOverArchivedCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeBarExemptionsAreFileScoped"
        status: pass
    human_judgment: false
  - id: D4
    description: "Six rows across the fourteen documents derive below their shipped ✅ green verdict -- the instrument's first real findings -- recorded honestly as new PHASE-14-DEBT.md rows (D-14-48..52, D-14-54) with a real witness probe and an owning-phase form, never suppressed, re-graded upward, or rewritten into the archived document."
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeCapArchivedDeadCitationsRemainAbsent"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeCapBarePackageRowHasNoNamedTest"
        status: pass
      - kind: manual_procedural
        ref: "cross-referenced against 08-REVIEW.md's WR-01 and 08-05-SUMMARY.md's own key-decisions, which independently documented two of the three rename-caused findings before this plan derived them mechanically"
        status: pass
    human_judgment: false
  - id: D5
    description: "The groundedness lint's pinned frontier is re-measured over the migrated corpus (125 -> 124 entries, both the +2 coordinate shift and the one genuine removal attributed in the literal's own doc comment), and the enforced-tier document count is unchanged across the Status-column retirement (431 documents before and after, independently verified by temporarily restoring the pre-migration documents)."
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessFrontierIsPinned"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestVerificationGroundednessCorpusIsNotEmpty"
        status: pass
    human_judgment: false

duration: 385min (first commit to last commit, dominated by run-record generation wall-clock, not authoring time)
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 09: Evidence Grade Cap (EVD-02) Summary

**A declared-grade-capped-by-derived-ceiling law closes EVD-02: all 14 archived-and-live `*-VALIDATION.md` documents now carry a closed-vocabulary Grade column instead of a freeform Status, mechanically capped by a total-order ladder that resolves real test names, and six rows that were shipped `✅ green` derive below the satisfying bar -- recorded as new debt, not suppressed.**

## Performance

- **Duration:** ~385 min commit-to-commit span (dominated by repeated real run-record generation over the ~230-row archived corpus -- several separate full and partial passes, each taking minutes of genuine `go test` wall-clock, not agent think time)
- **Tasks:** 3
- **Files modified:** 20 (2 created: `evidence_grade_test.go`, `scripts/evidence-run-record.sh`; 14 `*-VALIDATION.md` migrations; `PHASE-14-DEBT.md`; `.planning/UNREACHABLE-CLAIMS.md`; `verification_groundedness_test.go`; `session_test.go`; `witness_registry_test.go`)

## Accomplishments

- **Task 1 (the derivation ladder, the total order, the cap):** `internal/compiler/session/evidence_grade_test.go` declares the closed five-member grade vocabulary (`DEFINED`/`WIRED`/`REACHABLE`/`EXERCISED`/`MUTATION-KILLED`) and its total order, asserted at the boundary (equal grades compare neither above nor below). `deriveCeiling` implements every rung: exact-resolving `go test -run` patterns with a confirmed run-record pass derive `EXERCISED`; unresolved identifiers or an absent run record cap at `WIRED`; a distinct resolving, passing Non-inertness twin upgrades to `MUTATION-KILLED`; compile-time evidence (`go build`/`go vet`/`BUILD_OK`) derives `WIRED`, never `EXERCISED`; a `callsite:` token with zero call sites derives `WIRED`; any other recognized verification command derives `REACHABLE`; `DEFINED` is the floor. `scripts/evidence-run-record.sh` is the run-record producer, invoked via `exec.CommandContext` with a deadline and bounded stdout/stderr writers. An empty/absent Grade cell fails outright; a row declaring `EXERCISED`-or-above with no run record present fails closed.
- **Task 2 (derive over the whole archived corpus, migrate the fourteen documents, record findings honestly):** All fourteen `*-VALIDATION.md` documents migrated -- Status retired, Grade + Non-inertness added, frontmatter gains `evidence_vocabulary: v1` and `graded_rows: N` cross-checked against the table's own row count. `TestValidationRowGradesAreEarnedOverArchivedCorpus` proves derivation is total (every row across all fourteen documents classifies) and enforces the `>= EXERCISED` satisfying bar only for phase 14 onward via a file-scoped, dated `validationGradeBarExemptions` map (13 archived files as frozen prior art; `14-VALIDATION.md` itself exempted too, since its own template row was never filled in -- recorded as debt D-14-53 rather than silently exempted forever). Six rows derive below their shipped `✅ green` verdict: two are independently corroborated by plan 14-01's own pinned groundedness frontier (`09-VALIDATION.md:87`, `:95` -- line numbers post-migration), one names no exact test identifier (`12-VALIDATION.md`), and three cite tests retired or renamed in later phases without the citing archived row ever being repointed (`04-VALIDATION.md:74`, and two separate rows in `06-VALIDATION.md`). All six recorded as new `PHASE-14-DEBT.md` rows (D-14-48 through D-14-52, D-14-54) with a real witness probe, never rewritten into the archived documents.
- **Task 3 (prove the cap is not inert, re-pin the frontier):** `TestValidationGradeCapIsNotInert` seeds three faults (nonexistent non-inertness twin at `MUTATION-KILLED`, nonexistent evidence test at `EXERCISED`, missing run record at `EXERCISED`) over synthetic rows and proves the cap refuses each, with a genuine TDD RED (stubbed `validationRowProblem` to always return `""`, confirmed 3 of 4 subtests fail) before GREEN. The groundedness lint's pinned frontier is re-measured over the migrated corpus (125 -> 124 entries): every record naming one of the fourteen migrated files shifted by exactly +2 (the two new frontmatter keys), and one entry was removed for a real reason (plan 14-09 Task 1's own `TestValidationRowGradesAreEarned` stopped being a dead pattern once implemented). The enforced-tier document count is unchanged across the Status-column retirement (431 documents, independently verified before and after by temporarily restoring the pre-migration documents) -- plan 14-06's Status/Grade dual-keying did exactly the job it was built for.
- Found and fixed two regressions plan 14-09's own edits exposed in EXISTING tests (see Deviations), and two other hardcoded `(file, line)` expectations beyond the pinned frontier literal that the same migration shifted.

## Task Commits

1. **Task 1 RED: add failing tests for the grade derivation ladder and cap** -- `5050f57` (test)
2. **Task 1 GREEN: implement the grade derivation ladder and cap** -- `5e27245` (feat)
3. **Task 2: derive over the whole archived corpus, migrate the fourteen verification maps, record findings honestly** -- `3fe461e` (feat)
4. **Task 3 RED: add failing tests for the cap's non-inertness proof** -- `2274745` (test)
5. **Task 3 GREEN: prove the cap is not inert on three seeded faults** -- `1d28594` (feat)
6. **Task 3: re-pin the groundedness lint's frontier over the migrated corpus** -- `cf65b34` (feat)
7. **Deviation fix: update two other hardcoded line numbers the migration shifted** -- `0157fd8` (fix)
8. **Deviation fix: repair two regressions plan 14-09's own edits exposed** -- `9647026` (fix)
9. **Deviation fix: consolidate run-record subprocess spawns per package to clear the 10-minute test timeout** -- `52fd21d` (fix)

**Plan metadata:** committed separately (this SUMMARY + REQUIREMENTS.md + STATE.md + ROADMAP.md)

_Note: Tasks 1 and 3 are `tdd="true"`. Each RED commit stubbed exactly the function under test to a trivially-wrong return value (compiling cleanly, so failure is a real assertion failure, never a compile error), confirmed the new subtests fail against the stub, then restored the real implementation for the GREEN commit._

## Files Created/Modified

- `internal/compiler/session/evidence_grade_test.go` -- the grade vocabulary, total order, derivation ladder, run-record consumer, corpus-wide production test, bar exemption map, non-inertness proof, and witness probes for the six new debt rows.
- `scripts/evidence-run-record.sh` -- the sequential run-record producer.
- `internal/compiler/session/verification_groundedness_test.go` -- pinned frontier re-measured (125 -> 124); two other hardcoded `(file, line)` expectations updated for the +2 migration shift.
- `internal/compiler/session/session_test.go` -- `debtRegisterEmptyTableFixture`'s hardcoded `items: 3` literal replaced with a regex deriving the live count.
- `internal/compiler/session/witness_registry_test.go` -- no logic change; the generated view it enforces was regenerated (see below).
- All fourteen `*-VALIDATION.md` documents -- Status retired, Grade + Non-inertness added, frontmatter migrated.
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` -- seven new rows (D-14-48 through D-14-54).
- `.planning/UNREACHABLE-CLAIMS.md` -- regenerated (6 -> 12 entries) to include six of the seven new debt rows that carry a `probe:` witness.

## Decisions Made

See `key-decisions` in frontmatter for full detail. In summary: (1) reused plan 14-01's static test index and plan 14-07's callsite-witness machinery directly, no second resolver; (2) every migrated Grade is authored at exactly its derived ceiling (capped at `EXERCISED`, never claiming `MUTATION-KILLED`) with Non-inertness uniformly `—` -- a disclosed, deliberately conservative migration-authoring choice; (3) run-record generation for the ~230-row corpus is genuinely expensive (several minutes) -- a parallelism attempt was tried, measured to thrash, and reverted to sequential, documented rather than hidden; (4) migration Grade values were authored from a second, complete, independently-verified run-record generation after the first pass turned out to have been truncated by an operational timeout; (5) three secondary supplement sub-tables (02-VALIDATION.md x2, 11-VALIDATION.md x1) are explicitly out of migration scope and retain their Status column; (6) a pre-existing malformed row in 01-VALIDATION.md (missing one column) was found, not fixed, and graded `DEFINED`; (7) the real corpus-wide run surfaced three dead-citation findings beyond the two anticipated ones, all recorded as debt with a shared witness probe; (8) two regressions plan 14-09's own edits exposed in other tests were fixed (a hardcoded fixture literal, a stale generated view).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] The M001-era `assert-go-tests.sh` evidence shape was not covered by the original ladder design**
- **Found during:** Task 2, while deriving grades over the real M001 corpus (01-06-VALIDATION.md), which predates the `go test -run` convention and instead uses `env GOCACHE=... sh scripts/assert-go-tests.sh <package> <TestName>...`
- **Issue:** Without a dedicated rung, every M001-era row would have derived `DEFINED` (the generic fallback), massively understating grades across the oldest third of the corpus
- **Fix:** Added `deriveAssertGoTestsCeiling`, reusing the script's own exact-match discipline (every named test must resolve verbatim; any one unresolved name derives `WIRED` for the whole row, no partial credit)
- **Files modified:** `internal/compiler/session/evidence_grade_test.go`
- **Verification:** `TestValidationRowGradesAreEarnedOverArchivedCorpus` passes across all fourteen documents including the M001 corpus
- **Committed in:** `3fe461e`

**2. [Rule 1 - Bug] `debtRegisterEmptyTableFixture` hardcoded a stale `items: 3` literal**
- **Found during:** Final `go test ./...` run, after Task 3's commits landed
- **Issue:** `TestDebtRegisterOwnershipGuardIsNotInert/emptied_Items_table_passes_vacuously` failed: `PHASE-14-DEBT.md: frontmatter declares items: 10 but the Items table holds 0 rows`. This plan's own `items:` growth (3 -> 10 across D-14-04 through D-14-09) made the fixture's hardcoded `strings.Replace(text, "items: 3", "items: 0", 1)` a silent no-op
- **Fix:** Replaced with a regex (`^items:\s*\d+\s*$`) deriving the live count instead of a hardcoded prior snapshot
- **Files modified:** `internal/compiler/session/session_test.go`
- **Verification:** `TestDebtRegisterOwnershipGuardIsNotInert` passes in full
- **Committed in:** `9647026`

**3. [Rule 1 - Bug] `.planning/UNREACHABLE-CLAIMS.md` went stale relative to the new debt rows**
- **Found during:** Same final `go test ./...` run
- **Issue:** `TestUnreachableClaimsViewIsCurrent` failed: six of this plan's seven new `PHASE-14-DEBT.md` rows carry a `probe:` witness token and therefore qualify for this generated view (D-14-13), but the checked-in file was still the pre-plan 6-entry version
- **Fix:** Regenerated via the file's own `deriveUnreachableClaims`/`renderUnreachableClaimsView` functions (never hand-edited), producing the correct 12-entry file
- **Files modified:** `.planning/UNREACHABLE-CLAIMS.md`
- **Verification:** `TestUnreachableClaimsViewIsCurrent` and its three sibling non-inertness/load-bearing tests all pass
- **Committed in:** `9647026`

**4. [Rule 1 - Bug] Two other tests carried hardcoded `(file, line)` expectations the migration's +2 shift broke**
- **Found during:** Running the full `TestVerificationGroundedness` family (not just the pinned frontier) after re-pinning
- **Issue:** `TestVerificationGroundedness`'s `mustContain` list (`08-VALIDATION.md:51`/`:63`, `09-VALIDATION.md:85`) and `TestVerificationGroundednessGrepOverRealCorpus`'s hardcoded archival-breakage line (`09-VALIDATION.md:113`) both predated the frontmatter migration and were never touched by the pinned-frontier re-measurement alone, since they assert the same underlying findings through a different mechanism (an explicit "must contain" check, not set equality)
- **Fix:** Updated to the post-migration line numbers (+2 each: `:53`/`:65`, `:87`, `:115`), with a comment explaining the shift
- **Files modified:** `internal/compiler/session/verification_groundedness_test.go`
- **Verification:** `TestVerificationGroundedness` family passes in full
- **Committed in:** `0157fd8`

**5. [Rule 3 - Blocking Issue] Full `go test ./...` (no `-timeout` override) panicked: "test timed out after 10m0s"**
- **Found during:** The final, full-suite `go test ./...` re-run after all four prior deviations landed -- the plan's own literal acceptance criterion
- **Issue:** `TestValidationRowGradesAreEarnedOverArchivedCorpus` alone took 604.475s inside the session package's single test binary, tripping Go's default 10-minute per-package (whole-binary, not per-test) timeout. Root cause: `corpusRunRecord`'s `sync.Once`-memoized `generateRunRecord` call spawned up to ~208 individual sequential `go test` subprocesses -- one per distinct `(package, pattern)` pair cited across the corpus -- each paying its own process-start-plus-compile overhead on top of real test execution time
- **Fix:** Added `consolidatePkgPatterns`, which folds every pair sharing the same Go package into ONE pair whose pattern is the alternation of every distinct pattern cited for that package, cutting subprocess count from up to 208 to ~20 (one per distinct package). Lowered `evidenceRunRecordTimeout` from 900s to 300s to match the now-realistic cost, with a doc comment explaining the non-overridable 10-minute whole-binary ceiling this budget must stay well under
- **Files modified:** `internal/compiler/session/evidence_grade_test.go`
- **Verification:** The targeted corpus test alone dropped from 367-900+s to 136.07s in isolation; the full, unfiltered `go test ./...` (matching the plan's literal verify command exactly, no `-timeout` override) now exits 0, with the session package at 348.988s -- comfortably under the 600s default. Re-ran every other grade-cap and groundedness test (`TestVerificationGroundednessFrontierIsPinned`, `TestValidationGradeVocabularyIsClosed`, `TestValidationGradeOrderIsTotal`, `TestValidationRowGradesAreEarned`, `TestValidationGradeCapIsNotInert`, `TestValidationGradeBarExemptionsAreFileScoped`, `TestValidationGradeCapArchivedDeadCitationsRemainAbsent`, `TestValidationGradeCapBarePackageRowHasNoNamedTest`) to confirm the consolidation changed only subprocess batching, not any derived grade -- all pass
- **Committed in:** `52fd21d`

---

**Total deviations:** 5 auto-fixed (1 missing critical -- the M001 evidence shape; 3 bugs found via testing against the real corpus and the full suite; 1 blocking issue -- the default test timeout -- found only by running the complete, unfiltered `go test ./...` after every other fix landed, not assumed from any narrower or `-timeout`-overridden run).
**Impact on plan:** All five were necessary for the ladder to correctly classify the real corpus, for the existing test suite to stay green under this plan's own document edits, and for the plan's own literal acceptance criterion (`go test ./...` exits 0, no flag override) to actually hold -- exactly the failure mode (an instrument or its supporting fixtures silently going stale, or silently becoming too expensive, under a legitimate edit) this phase exists to retire. No scope creep; no architectural change.

## Issues Encountered

- Real run-record generation over the ~230-row corpus is genuinely expensive (over a minute per full pass even after consolidation), corrected T-14-56's original "milliseconds" assumption. A bounded-parallelism attempt was tried mid-plan and measured to thrash rather than speed up; reverted to sequential and documented. A separate per-package subprocess-count fix (consolidatePkgPatterns, deviation 5) was needed before the full suite would finish inside Go's default 10-minute per-package timeout at all. This means `go test ./...` for this repository now costs a real, disclosed additional ~150s+ whenever `TestValidationRowGradesAreEarnedOverArchivedCorpus` runs (session package: 348.988s vs. the pre-plan baseline of 192.7s per the QLT-02 budget manifest) -- worth reconciling explicitly in a future budget-manifest update, which this plan does not itself perform.
- The first run-record generation pass used to author migration Grade values was itself truncated by an internal timeout before completing the full corpus, which initially produced an inflated, inaccurate debt-row candidate list (~46 rows). A second, independently-run, complete pass (verified via direct spot-checks against real `go test` output) corrected this before any Grade value or debt row was finalized -- documented so the discrepancy is traceable rather than silently smoothed over.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- EVD-02 closes: every `*-VALIDATION.md` row across the fourteen archived-and-live documents carries a mechanically capped grade from the closed vocabulary; the vocabulary is closed and totally ordered; derivation is total; the satisfying bar is enforced from phase 14 onward under file-scoped dated exemptions; the cap is proven non-inert on three seeded fault kinds; and every row that derives below its shipped verdict is recorded as debt, not suppressed.
- `go test ./...` and `go vet ./...` are green (confirmed via a full, unfiltered re-run after all five fixes landed -- exit 0, session package 348.988s; the run now takes noticeably longer than the pre-plan baseline -- see Issues Encountered).
- `git status --short` is clean except the pre-existing untracked `.planning/milestone.lock`.
- Plan 14-10 (Nyquist reconciliation) inherits: the R1/R2/R2b/R3 groundedness frontier (unchanged in scope by this plan, just re-pinned for the coordinate shift); two of this plan's six new debt rows (D-14-48, D-14-49) are explicitly landing-phase `P14`, the same reconciliation plan, since both are members of the groundedness frontier's own R2 class.
- Three findings (D-14-50, D-14-51, D-14-52, D-14-54) are `UNOWNED`, recorded but not yet scheduled for closure -- honest, not a blocker for this plan's own completion.
- No blockers for the next plan in this phase.

## Self-Check: PASSED

- `[ -f internal/compiler/session/evidence_grade_test.go ]` → FOUND
- `[ -f scripts/evidence-run-record.sh ]` → FOUND, executable
- `[ -f .planning/phases/14-evidence-instrument-and-honest-scoping/14-09-SUMMARY.md ]` → FOUND (this file)
- Commits present in `git log --oneline --all`: `5050f57`, `5e27245`, `3fe461e`, `2274745`, `1d28594`, `cf65b34`, `0157fd8`, `9647026`, `52fd21d` -- all FOUND. The plan metadata commit is self-referential and cannot state its own SHA: verify with `git log --oneline -1 -- .planning/phases/14-evidence-instrument-and-honest-scoping/14-09-SUMMARY.md`.
- `go test ./internal/compiler/session/... -run 'TestValidationGradeVocabularyIsClosed|TestValidationGradeOrderIsTotal|TestValidationRowGradesAreEarned' -count=1 -v` re-run: all `--- PASS:` lines, no failures.
- `go test ./internal/compiler/session/... -run 'TestValidationGradeCapIsNotInert' -count=1 -v` re-run: 4/4 subtests PASS.
- `go test ./internal/compiler/session/... -run 'TestVerificationGroundedness' -count=1` re-run: exits 0.
- `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed|TestDebtRegisterOwnershipGuardIsNotInert|TestUnreachableClaimsViewIsCurrent'` re-run: exits 0.
- `go test ./internal/compiler/session/... -run 'TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundedness$|TestVerificationGroundednessGrepOverRealCorpus|TestValidationGradeVocabularyIsClosed|TestValidationGradeOrderIsTotal|TestValidationRowGradesAreEarned$|TestValidationGradeCapIsNotInert|TestValidationGradeBarExemptionsAreFileScoped|TestValidationGradeCapArchivedDeadCitationsRemainAbsent|TestValidationGradeCapBarePackageRowHasNoNamedTest' -count=1 -v -timeout 300s` re-run after the timeout fix (deviation 5): all listed tests PASS, confirming `consolidatePkgPatterns` changed only subprocess batching, not any derived grade.
- `go build ./...` and `go vet ./...` re-run: both clean.
- Full, unfiltered `go test ./...` re-run after all five fixes landed and the fifth fix was committed (`52fd21d`), exactly as the plan's own verify command specifies (no `-timeout` override): `EXIT:0`, every package `ok`, session package `348.988s` (confirmed via a tracked background run, output captured to a scratch file and grepped for `FAIL`/`panic`/`EXIT` -- none found except the trailing `EXIT:0`).
- `git status --short`: clean except the pre-existing untracked `.planning/milestone.lock`.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

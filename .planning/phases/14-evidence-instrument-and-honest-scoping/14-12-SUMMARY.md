---
phase: 14-evidence-instrument-and-honest-scoping
plan: 12
subsystem: testing
tags: [go, evidence-instrument, grade-bar, debt-register, qlt02]

# Dependency graph
requires:
  - phase: 14-evidence-instrument-and-honest-scoping (plan 11)
    provides: the completion-witnessed run record (evidenceRunRecordTimeout=480s,
      evidenceRunRecordMarginFraction=0.75, resolvedPkgPatterns anchoring) that
      makes a complete, honestly-timed corpus-wide run possible
  - phase: 14-evidence-instrument-and-honest-scoping (plan 13)
    provides: the build-constraint suppression guard's real test names
      (TestBuildConstraintsOutsideTheAllowlistAreRefused,
      TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile)
provides:
  - 14-VALIDATION.md removed from validationGradeBarExemptions; the
    corpus-wide >=EXERCISED bar confirmed for it with a complete run
    record (measured elapsed 90-92s against the 480s budget)
  - validationGradeBarRowExemptions, the row-scoped narrowing that
    replaces a file-scoped exemption for any document authored from
    Phase 14 on, each entry required to cite a debt row with a real
    P<NN>/CLOSED(sha) owner (never UNOWNED)
  - TestValidationGradeBarAppliesToPhase14, the permanent two-half guard
    against the exemption returning or the table being emptied
  - TestValidationGradeBarRowExemptionsAreOwned + rowExemptionProblems,
    the ownership check and its shared predicate
  - D-14-121 and D-14-53 transitioned from UNOWNED(none-yet-scheduled) to
    CLOSED(128ecec)
affects: []

actuals:
  tokens: 11688
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Row-scoped debt-witnessed narrowing (validationGradeBarRowExemptions)
      as the honest successor to a file-scoped exemption map -- every entry
      must resolve to a debt row with a real owner, never UNOWNED, checked
      by the same predicate (rowExemptionProblems) its own seeded-fault
      proof exercises"
    - "Two-half permanent guard (exemption absence + non-vacuity row floor)
      so a finding cannot be silently reopened by either of two structurally
      different regressions"

key-files:
  created: []
  modified:
    - internal/compiler/session/evidence_grade_test.go
    - .planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/UNREACHABLE-CLAIMS.md

key-decisions:
  - "Task 1 and Task 2 landed in one commit (128ecec): Task 1's own
    acceptance criteria required the corpus test to pass, and passing
    required Task 2's row-scoped exemption mechanism to already exist --
    splitting them would have left an intermediate commit red for a test
    the plan itself required green at Task 1's boundary."
  - "validationGradeBarRowExemptions ships with 5 entries, not empty as
    the plan's own Task 2 text expected -- real re-derivation against the
    live corpus falsified that expectation (5 pre-existing rows in
    14-VALIDATION.md's own table fail the new bar for reasons this plan's
    own never-rewrite-an-archived-row prohibition forbids fixing by
    editing the row). Each is narrowed via a new debt row (D-14-123..127,
    P14, witnessed by TestValidationGradeBarRowExemptionsAreOwned)."
  - "D-14-121's and D-14-53's Landing phase cells both close via
    CLOSED(128ecec) rather than a future P<NN>: the corpus-wide bar is
    confirmed and permanently guarded by this same plan, so there is no
    remaining work for a future phase to schedule."

requirements-completed: [EVD-02, PRC-01]

coverage:
  - id: D1
    description: "14-VALIDATION.md removed from validationGradeBarExemptions and the corpus-wide >=EXERCISED bar confirmed for it with a complete run record"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationRowGradesAreEarnedOverArchivedCorpus"
        status: pass
    human_judgment: false
  - id: D2
    description: "A permanent guard fails if the exemption returns or if 14-VALIDATION.md's table is emptied below its 31-row floor"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeBarAppliesToPhase14"
        status: pass
    human_judgment: false
  - id: D3
    description: "Any residual narrowing of the bar is row-scoped and cites a debt row with a real owner (never UNOWNED), proven with 3 seeded faults"
    requirement: "EVD-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationGradeBarRowExemptionsAreOwned"
        status: pass
    human_judgment: false
  - id: D4
    description: "D-14-121 and D-14-53 transition out of UNOWNED to CLOSED(<sha>) with a reason naming the measured margin"
    requirement: "PRC-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 12: EVD-02 Exemption Removal and PRC-01 Closure Summary

**14-VALIDATION.md's file-scoped grade-bar exemption is gone for good -- the corpus-wide `>=EXERCISED` bar now holds for it with a complete run record (90-92s against the 480s budget), guarded by a permanent two-half check, and D-14-121/D-14-53 both close `CLOSED(128ecec)` instead of staying `UNOWNED`.**

## Performance

- **Duration:** ~55 min
- **Started:** 2026-09-18T14:15:00Z (approx.)
- **Completed:** 2026-09-18T15:10:00Z (approx.)
- **Tasks:** 3/3
- **Files modified:** 4

## Accomplishments

- Removed `14-VALIDATION.md` from `validationGradeBarExemptions` (13 frozen M001/M002 files remain) and added 8 real Per-Task Verification Map rows (`14-11-T1..T3`, `14-13-T1..T2`, `14-12-T1..T3`), each evidence command executed individually and confirmed to exit 0 before being written. `graded_rows` moved 31 -> 39.
- `TestValidationRowGradesAreEarnedOverArchivedCorpus` passes for `14-VALIDATION.md` with a run record that reported complete and a measured elapsed of 90-92s against the `480s` budget across four independent runs during this plan (18.5%-19.3% of budget, comfortably inside the 75% margin -- nowhere near the original `300.11s`/`300s`-ceiling near-miss).
- Added `validationGradeBarRowExemptions` (the row-scoped narrowing successor to a file-scoped exemption), `rowExemptionProblems` (the one shared predicate), `phase14DebtLandingPhases` (a `PHASE-14-DEBT.md` lookup built on the existing `parseDebtRegisterTable`), `TestValidationGradeBarRowExemptionsAreOwned` (3 seeded subtests plus the shipped-map check), and `TestValidationGradeBarAppliesToPhase14` (the permanent two-half guard: exemption absence + a `phase14ValidationRowFloor=31` non-vacuity floor).
- Manual red proof: temporarily re-added the `14-VALIDATION.md` key to `validationGradeBarExemptions`, observed `TestValidationGradeBarAppliesToPhase14` fail naming the file (`"14-VALIDATION.md must never re-appear in validationGradeBarExemptions -- D-14-121's finding is permanently closed, not merely fixed once"`), reverted, confirmed `git status --short` clean.
- D-14-121 and D-14-53 both transitioned `UNOWNED(none-yet-scheduled)` -> `CLOSED(128ecec)`, Grade `WIRED`, Witness `probe:TestValidationGradeBarAppliesToPhase14`; their detail sections rewritten to record the actual root cause (plan 14-11's fix), the measured elapsed, and the budget/margin numbers. Zero `UNOWNED(none-yet-scheduled)` cells remain in `PHASE-14-DEBT.md` (was 4 before this plan).
- Regenerated `.planning/UNREACHABLE-CLAIMS.md` (12 -> 19 entries) as the five new row-scoped debt rows plus D-14-121/D-14-53 now carry `probe:` witnesses and qualify for that generated view.
- Full `go test ./... -count=1` green, 25 packages, confirmed after each of the two commits.

## Task Commits

Task 1 and Task 2 landed together (see Deviations below for why):

1. **Task 1 + Task 2: Remove the exemption, confirm the bar, add permanent guards** - `128ecec` (feat)
2. **Task 3: D-14-121 and D-14-53 leave UNOWNED with real reasons** - `85c4dee` (docs)

No separate plan-metadata commit for this SUMMARY -- it lands as this plan's own docs commit alongside STATE.md/ROADMAP.md per the standard execute-plan protocol (see final commit below).

## Files Created/Modified

- `internal/compiler/session/evidence_grade_test.go` - removed the `14-VALIDATION.md` file-scoped exemption entry and rewrote the map's doc comment; added `validationGradeBarRowExemptions`, `phase14DebtLandingPhases`, `rowExemptionProblems`, `TestValidationGradeBarRowExemptionsAreOwned`, `phase14ValidationRowFloor`, `TestValidationGradeBarAppliesToPhase14`; wired the row-scoped map into the corpus test's per-row bar check
- `.planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md` - added 8 new Per-Task Verification Map rows (`14-11-T1..T3`, `14-13-T1..T2`, `14-12-T1..T3`), `graded_rows` 31 -> 39
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` - added D-14-123..D-14-127 (the five row-scoped narrowings), transitioned D-14-121 and D-14-53's Landing phase cells and detail sections to `CLOSED(128ecec)`, `items:` 78 -> 83
- `.planning/UNREACHABLE-CLAIMS.md` - regenerated (byte-derived, never hand-edited) to include the 5 new row-scoped debt rows plus D-14-121/D-14-53's new `probe:` witnesses (12 -> 19 entries)

## Decisions Made

- **Task 1 and Task 2 combined into one commit.** Task 1's acceptance criteria required `TestValidationRowGradesAreEarnedOverArchivedCorpus` to exit 0 for `14-VALIDATION.md`. Once the file-scoped exemption was removed, re-deriving the bar over the real, live corpus showed five pre-existing rows in `14-VALIDATION.md`'s own table (`14-01-T2`, `14-02-T1`, `14-02-T2`, `14-03-T1`, `14-10-T3`) fail it -- two (`14-01-T2`, `14-10-T3`) because their declared grade (`WIRED`) understates what a now-complete run record proves (`EXERCISED`), and three (`14-02-T1`, `14-02-T2`, `14-03-T1`) because their evidence cell is structurally incapable of reaching `EXERCISED` (a `go run` invocation, or a `go test` invocation with no `-run` pattern). This plan's own prohibition forbids rewriting any of those five rows to force the bar to pass. The plan's own Task 1 text anticipates exactly this: "for an archived 14-01..14-10 row you may not rewrite, route it through Task 2's row-scoped narrowing with a debt witness." Making Task 1's own acceptance criteria pass therefore required building (a minimal but complete version of) Task 2's mechanism first -- so both tasks are one commit.
- **`validationGradeBarRowExemptions` ships with 5 entries, not empty.** Task 2's own text expected the map to ship empty (the planner assumed every pre-existing row would already clear the bar once the file exemption was removed). Real execution falsified that assumption for the five rows above. Each is narrowed via a new, real, P14-owned debt row (D-14-123..D-14-127), not left ownerless -- satisfying the must-have that every row-scoped entry cites a debt row naming an owner from the closed PRC-01 vocabulary.
- **D-14-121 and D-14-53 close `CLOSED(128ecec)`, not a future `P<NN>`.** Both rows' premises are fully retired by this same plan (the corpus-wide bar is confirmed, permanently guarded, and the table is filled) -- there is no remaining work to schedule to a future phase.
- **Two debt-row Item cells were phrased to avoid backtick-quoted, incomplete command fragments** (`go run ./cmd/lang --json check testdata/distinctness/spiral_full.lang` written in full rather than truncated with `...`; `go test`/`go test -run` written as plain prose rather than in backticks) after the groundedness lint correctly flagged the truncated/bare forms as unparseable R1 findings. This is prose-quality reconciliation, not a code change, and does not affect any row's declared grade or the finding it records.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug in plan's own expectation] `validationGradeBarRowExemptions` could not ship empty as Task 2's text expected**
- **Found during:** Task 1, immediately after removing the file-scoped exemption and re-running the corpus test
- **Issue:** Five pre-existing rows in `14-VALIDATION.md`'s own table (`14-01-T2`, `14-02-T1`, `14-02-T2`, `14-03-T1`, `14-10-T3`) fail the newly-applied `>=EXERCISED` bar. Two declare `WIRED` though a complete run record now derives `EXERCISED` for their cited tests; three are structurally capped below `EXERCISED` by their evidence cell's own shape. This plan's own prohibition forbids fixing either class by editing the row.
- **Fix:** Built `validationGradeBarRowExemptions` + `rowExemptionProblems` + `phase14DebtLandingPhases` + `TestValidationGradeBarRowExemptionsAreOwned` (originally specified as Task 2 deliverables) as part of Task 1's own commit, populated with 5 entries citing 5 new `PHASE-14-DEBT.md` rows (D-14-123..D-14-127, all `P14`, all witnessed by `TestValidationGradeBarRowExemptionsAreOwned`), exactly per Task 1's own contingency text.
- **Files modified:** `internal/compiler/session/evidence_grade_test.go`, `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md`
- **Verification:** `go test ./internal/compiler/session/ -run 'TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1 -v` exits 0; `TestValidationGradeBarRowExemptionsAreOwned` exits 0 with 3 seeded subtests plus the shipped-map check
- **Committed in:** `128ecec`

**2. [Rule 3 - Blocking issue] New probe-witnessed debt rows made `.planning/UNREACHABLE-CLAIMS.md` stale**
- **Found during:** Task 1, first `go test ./internal/compiler/session/ -run 'TestUnreachableClaimsViewIsCurrent$'` after adding D-14-123..D-14-127
- **Issue:** `TestUnreachableClaimsViewIsCurrent` byte-compares a regenerated view against the checked-in `.planning/UNREACHABLE-CLAIMS.md`. The five new debt rows carry `probe:TestValidationGradeBarRowExemptionsAreOwned` witnesses and now qualify for that view (any row whose Witness cell names a `probe:` token qualifies, regardless of Landing phase), so the checked-in file was out of date.
- **Fix:** Regenerated `.planning/UNREACHABLE-CLAIMS.md` from the live registers (12 -> 19 entries across this plan's two commits, as D-14-121/D-14-53 later also gained `probe:` witnesses in Task 3). Never hand-edited, per the file's own generation contract.
- **Files modified:** `.planning/UNREACHABLE-CLAIMS.md`
- **Verification:** `TestUnreachableClaimsViewIsCurrent` exits 0
- **Committed in:** `128ecec`, `85c4dee`

**3. [Rule 3 - Blocking issue] A truncated backtick-quoted command in a new debt-row Item cell triggered a false-positive groundedness-lint finding**
- **Found during:** Task 1, `go test ./...` sweep after adding D-14-124
- **Issue:** D-14-124's Item text originally quoted `` `go run ./cmd/lang --json check ...` `` and `` `go test -run` `` in backticks. `verificationCommandPattern` matches any backtick-delimited span starting with `go run`/`go test`; both spans were incomplete/unresolvable, producing two new `unparseable` R1 violations not in the pinned frontier (`TestVerificationGroundednessFrontierIsPinned` failed) and, via the same mechanism, an unreconciled live-finding failure in `TestReconciliationVerdictsCarryTheirObligations` (since the identical text is regenerated verbatim into `.planning/UNREACHABLE-CLAIMS.md`).
- **Fix:** Reworded the Item cell to spell out the real, full, resolving command (`go run ./cmd/lang --json check testdata/distinctness/spiral_full.lang`, matching this plan's own Detail-section text) instead of an ellipsis-truncated fragment, and moved the bare `go test`/`go test -run` mentions out of backticks into plain prose. No finding, code, or grade was altered -- purely a prose fix to avoid a false match on an incomplete command fragment.
- **Files modified:** `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md`, `.planning/UNREACHABLE-CLAIMS.md` (regenerated)
- **Verification:** `TestVerificationGroundednessFrontierIsPinned`, `TestReconciliationVerdictsCarryTheirObligations`, `TestUnreachableClaimsViewIsCurrent` all exit 0
- **Committed in:** `128ecec`

---

**Total deviations:** 3 auto-fixed (1 Rule 1 - plan's own falsified expectation, 2 Rule 3 - blocking issues)
**Impact on plan:** All three were necessary to make the plan's own stated acceptance criteria (a genuinely passing corpus-wide bar test, plus `go test ./...` green) achievable at all, given what real execution against the live corpus actually found. None weakened any grade, suppressed any finding, or rewrote an archived row -- every new finding got its own named, owned debt row. No scope creep.

## Issues Encountered

- Accidentally ran `git stash push -u` mid-session while investigating a test failure (a destructive-git-prohibition violation for this sequential, main-working-tree execution). Immediately recognized the mistake, ran `git stash pop` to restore the working tree from the correct stash entry (verified by name/message before popping), and confirmed `git status --short` showed the same in-progress changes as before. A separate, unrelated pre-existing stash entry from before this session was left untouched throughout. No work was lost; no further `git stash` commands were run.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- EVD-02 closed: the corpus-wide `>=EXERCISED` satisfying bar is confirmed for `14-VALIDATION.md` itself, reflexively, with a permanent guard against the exemption returning or the table being emptied.
- PRC-01 closed: D-14-121 and D-14-53 both carry `CLOSED(128ecec)` with a probe witness; zero `UNOWNED(none-yet-scheduled)` cells remain in `PHASE-14-DEBT.md`.
- `go test ./... -count=1` exits 0 (25 packages) after both commits in this plan.
- `14-VERIFICATION.md`'s one recorded gap (the corpus-wide bar was never confirmed for Phase 14's own artifact) is closed. This was the last open item for Phase 14 named in `.planning/STATE.md`'s Current Position section; Phase 14 has no other outstanding work known to this plan.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

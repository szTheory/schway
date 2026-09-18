---
phase: 14-evidence-instrument-and-honest-scoping
plan: 04
subsystem: process
tags: [go, debt-register, prc-01, evd-07, mechanized-vocabulary, git-blame-migration]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-02's diagnostic-distinctness debt rows (deferred here for registration)"
provides:
  - "internal/compiler/session/session_test.go's checkDebtRegister extended with a closed three-form owning-phase vocabulary (P<NN> | CLOSED(<sha>) | UNOWNED(<witness>)), a PHASE-<NN>-DEBT.md naming convention with a frozen-prior-art exemption list, and a first-recorded: milestone requirement per row"
  - "debtRegisterProblems -- a pure, non-fataling, deterministic variant of the well-formedness law that reports every problem in table order instead of failing on the first"
  - "all twelve pre-existing *-DEBT.md registers migrated to the closed vocabulary, cell-format only, item text/severity/identifier untouched"
  - ".planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md -- registers the -flto multi-function inertness claim (closing EVD-07's outstanding half) plus two debt rows plan 14-02 deferred"
affects: ["plan 14-07 (adds the Witness column and resolves UNOWNED(...) witness identifiers to executed probes)", "any future plan that opens a new *-DEBT.md register"]

actuals:
  tokens: 55498
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Pure, non-fataling variant of a *testing.T-based well-formedness law (debtRegisterProblems) so a caller can inspect exact refusal text without a nested subtest capture -- the same pattern Task 3's non-inertness proof needed"
    - "Cell-format-only migration of an owned register field via a scripted, git-blame-sourced closure-commit derivation, verified byte-identical elsewhere in the row via `git diff`"

key-files:
  created:
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
  modified:
    - internal/compiler/session/session_test.go
    - .planning/milestones/M001-phases/02-owned-values-and-abilities/02-DEBT.md
    - .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md
    - .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-DEBT.md
    - .planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md
    - .planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-DEBT.md
    - .planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md
    - .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md
    - .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md
    - .planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md
    - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
    - .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md
    - .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md

key-decisions:
  - "Migration mapping methodology (documented here since it is not in the plan text): a row's Landing phase cell maps to CLOSED(<sha>) only when the row's own text (or a same-corpus cross-reference, e.g. PHASE-09-DEBT.md's 'resolves PHASE-08-DEBT.md D-08-40') explicitly declares closure (RESOLUTION/REVERSAL/SUPERSESSION/Resolved/Superseded/CLOSED --/WITHDRAWN AS INCORRECT); to P<NN> when it names exactly one specific phase number with no closure or OPEN/UNOWNED/Not-scheduled signal; otherwise to UNOWNED(<slug>). CLOSED shas are the git-blame commit of the line that recorded the closure statement, not an invented value."
  - "D-09-53 (PHASE-09-DEBT.md) migrated to CLOSED, not P10 as its own literal cell text ('CARRIED -- landing phase Phase 10') would suggest read in isolation: PHASE-10-DEBT.md's D-10-27/D-10-30 explicitly REVERSE it, stating its premise was 'WITHDRAWN AS INCORRECT' and its real content re-filed as the new D-10-28. Treating D-09-53 as still-open would have mis-migrated a row whose underlying claim the corpus itself already retracted."
  - "D-13-02b and D-13-10a map to P17, and D-13-34 to UNOWNED(probe:TestPhase6HeldoutPairsAreAlphaRenamesOnly), per 14-CONTEXT.md's D-14-30 already-ratified day-one register contents -- not independently re-derived."
  - "The D-14-31 two-milestone-carry ratification requirement is enforced only against the LIVE phase tree (.planning/phases/**), never archived milestone registers (.planning/milestones/**): an archived register is frozen historical record, not an active re-deferral, and requiring a fabricated ratification line on it would falsify history the register exists to preserve. first-recorded: lines are still required everywhere (added to all twelve registers, satisfying Task 2's diff-shape acceptance criterion)."
  - "The Witness column and its requirement that an UNOWNED(...) identifier resolve to an executed probe are explicitly NOT built in this plan -- per the plan's own stated hand-off to 14-07 -- and a code comment states this in session_test.go so the two halves are not mistaken for one law."

patterns-established:
  - "A *testing.T-based well-formedness law should expose a pure, error-returning sibling (here debtRegisterProblems) rather than only a Fatalf-driven one, so later plans can seed faults and assert on exact refusal text without nested-subtest output capture."

requirements-completed: [PRC-01, EVD-07]

coverage:
  - id: D1
    description: "checkDebtRegister enforces the closed three-form owning-phase vocabulary (P<NN> | CLOSED(<sha>) | UNOWNED(<witness>)), trims before testing so absent/empty/whitespace-only cells fail identically, reports every ownerless row in table order deterministically, handles a zero-row Items table as a vacuous pass, and enforces the PHASE-<NN>-DEBT.md naming convention with a frozen-prior-art exemption list"
    requirement: "PRC-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegisterOwnershipGuardIsNotInert"
        status: pass
    human_judgment: false
  - id: D2
    description: "All twelve pre-existing *-DEBT.md registers migrated to the closed owning-phase vocabulary as a cell-format-only change (item text, severity, identifier byte-identical, verified via git diff), and the new PHASE-14-DEBT.md registers the -flto multi-function inertness claim (closing EVD-07's remaining task) plus two debt rows deferred from plan 14-02"
    requirement: "EVD-07"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
      - kind: other
        ref: "grep -rn -- \"-flto\" .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md"
        status: pass
    human_judgment: true
    rationale: "The migration mapping (which cells are CLOSED vs P<NN> vs UNOWNED) is an editorial judgment call over ~90 rows of free prose, applied via a documented mechanical rule set and cross-checked against git blame and same-corpus closure statements; a human should spot-check a sample against the recorded methodology before treating the mapping as final, even though the mechanical test proves the SHAPE is correct."
  - id: D3
    description: "The ownership-vocabulary guard is proven non-inert: red on an emptied, whitespace-only, and free-prose owning-phase cell (each naming the offending row), green on an unmodified register, and vacuously green on an empty-Items-table register"
    requirement: "PRC-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegisterOwnershipGuardIsNotInert"
        status: pass
    human_judgment: false

duration: 6min (commit-to-commit span; full go test ./... verification ran after)
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 04: Debt Register Owning-Phase Vocabulary Summary

**A closed three-form owning-phase vocabulary (`P<NN>` | `CLOSED(<sha>)` | `UNOWNED(<witness>)`) is now mechanically enforced on every `*-DEBT.md` register, all twelve pre-existing registers migrated cell-by-cell via a documented rule set with git-blame-sourced closure shas, and the `-flto` multi-function inertness claim finally has a debt row.**

## Performance

- **Duration:** 6 min (commit-to-commit span, 21:46:52 → 21:52:26 local); full-suite verification (`go test ./...`, ~220s) ran after the last commit
- **Started:** 2026-09-17T21:46:52-04:00
- **Completed:** 2026-09-17T21:52:26-04:00
- **Tasks:** 3
- **Files created/modified:** 14 (1 created, 13 modified)

## Accomplishments

- `internal/compiler/session/session_test.go`: `checkDebtRegister` extended in place (same function, same file, no new parser per D-14-24) with:
  - `debtRegisterOwningPhaseForm` -- a closed three-form regex vocabulary (`^P\d{2}$`, `^CLOSED\([0-9a-fA-F]{7,40}\)$`, `^UNOWNED\([A-Za-z0-9][A-Za-z0-9:_./-]*\)$`), checked against the TRIMMED cell so absent, empty, and whitespace-only cells fail identically.
  - `debtRegisterNamingPattern`/`debtRegisterNamingExemptions` -- any register created from M003 on must be named `PHASE-<NN>-DEBT.md`; the five frozen M001 files (`02`–`06-DEBT.md`) are the only exemptions, dated and file-scoped, mirroring `debtRegisterLandingPhaseExemptions`'s existing shape.
  - `debtRegisterFirstRecordedPattern` -- every row's detail section must carry a `first-recorded: M0NN` line (D-14-31); the two-milestone-carry ratification check is scoped to the live phase tree only (see Key Decisions).
  - `debtRegisterProblems` -- a new pure, non-fataling function that returns every problem found in table order instead of stopping at the first; `checkDebtRegister` is now a thin `*testing.T` wrapper around it, and `debtRegisterTable`'s zero-row case no longer panics/fatals (delegated to `parseDebtRegisterTable`, which treats zero rows as valid).
- All twelve pre-existing registers migrated: every free-prose Landing phase cell now reads as `P<NN>`, `CLOSED(<sha>)`, or `UNOWNED(<witness>)`. `git diff` confirms the change is confined to Landing phase cells, added `first-recorded:` lines, and (for the two column-less legacy files) `first-recorded:` lines only.
- New `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` (3 items): D-14-45 (`-flto` multi-function inertness, closing EVD-07's outstanding half), D-14-46 and D-14-47 (the two debt rows plan 14-02's SUMMARY deferred here).
- `TestDebtRegisterOwnershipGuardIsNotInert`: 6 subtests proving the guard is not inert -- 3 seeded-fault subtests (emptied/whitespace/free-prose cell, each refusing and naming row D-14-45), an unmodified-copy control (passes), an empty-Items-table control (passes vacuously), and a fixture-isolation check (nothing leaks into the live `.planning` tree).

## Task Commits

1. **Task 1: Close the owning-phase vocabulary inside `checkDebtRegister`** — `40a377c` (test, RED)
2. **Task 2: Migrate the twelve registers, open `PHASE-14-DEBT.md`, and register the `-flto` inertness row** — `5bf76d5` (feat, GREEN)
3. **Task 3: Prove the ownership gate fails on a seeded ownerless row** — `af86062` (test)

_Note: Task 1 is `tdd="true"`: `40a377c` is the RED commit (the extended law compiles and correctly fails against the twelve unmigrated registers' free-prose cells and missing `first-recorded:` lines -- confirmed by running the test before committing); `5bf76d5` is the GREEN commit that makes it pass by migrating the data, not the law._

## Files Created/Modified

- `internal/compiler/session/session_test.go` — closed owning-phase vocabulary, naming convention + exemption list, `first-recorded:`/two-milestone-carry guard, `debtRegisterProblems` pure variant, `TestDebtRegisterOwnershipGuardIsNotInert`.
- Twelve `*-DEBT.md` registers — Landing phase cells migrated, `first-recorded:` lines added.
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` — new register, 3 items.

## Decisions Made

See `key-decisions` in frontmatter. In summary: the migration mapping rule (CLOSED only on explicit closure language, P<NN> on a single named phase, UNOWNED otherwise) is documented there since the plan itself delegated exact per-row classification to execution; the two most consequential findings were (1) D-09-53's literal "CARRIED — landing phase Phase 10" cell text is superseded by PHASE-10-DEBT.md's own explicit reversal, so it migrated to CLOSED rather than P10, and (2) the D-14-31 two-milestone-carry guard is scoped to the live tree only, to avoid fabricating a ratification record for archived, frozen-history registers.

## Deviations from Plan

None — plan executed as written. The per-row Landing-phase classification required editorial judgment the plan text did not spell out row-by-row (unavoidable, given ~90 free-prose cells across twelve registers); that judgment is documented in Key Decisions and in `coverage` D2's `human_judgment: true` rationale rather than treated as a deviation, since the plan's own action text explicitly left the exact vocabulary assignment to the migration.

**Total deviations:** 0. **Impact:** None.

## Issues Encountered

None new. The one pre-existing, unrelated `go test ./...` failure (`TestSourceNeverSpawnsUnboundedProcesses` in `internal/compiler/native`, flagged in `.planning/phases/14-evidence-instrument-and-honest-scoping/deferred-items.md` by plan 14-02) is still present, still confined to the same two files neither touched by this plan, and is out of this plan's scope per the orchestrator's own instruction. Full-suite run for this plan: every other package `ok`, including `internal/compiler/session` (215.2s, all subtests including the new ones pass).

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- PRC-01 mechanized: every debt item now names an owning phase from a closed, non-inert vocabulary, and the register's own well-formedness law enforces it on every commit.
- EVD-07's outstanding half closed: the `-flto` multi-function inertness claim has a debt row (D-14-45) with an owning-phase cell (currently `UNOWNED(...)`, pending an owning phase decision).
- Plan 14-07 has clean ground to add the `Witness` column and resolve every `UNOWNED(...)` identifier to an executed probe -- this plan deliberately left that resolution unbuilt per the stated hand-off.
- No blockers for the next plan in this phase.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED

- FOUND: `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md`
- FOUND: commit `40a377c` (test, RED, Task 1)
- FOUND: commit `5bf76d5` (feat, GREEN, Task 2)
- FOUND: commit `af86062` (test, Task 3)
- `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -count=1 -v` prints 13 `--- PASS:` subtest lines (one per register)
- `go test ./internal/compiler/session/... -run TestDebtRegisterOwnershipGuardIsNotInert -count=1 -v` prints 6 `--- PASS:` subtest lines
- `grep -rn -- "-flto" .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` matches
- `go test ./...` reports exactly one FAIL package (`internal/compiler/native`), confirmed pre-existing and unrelated to this plan's diff; `internal/compiler/session` reports `ok` (215.2s)

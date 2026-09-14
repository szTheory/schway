---
phase: 13-agent-loop-for-interprocedural-defects
plan: 07
subsystem: testing
tags: [held-out-corpus, structural-distinctness, mutation-testing, validation, debt-register]

# Dependency graph
requires:
  - phase: 13-agent-loop-for-interprocedural-defects (plan 04)
    provides: "the D-13-26 topology-triple methodology and TestCorpusTopologyGuardIsNotInert this plan's testdata/phase6 predicate mirrors at intraprocedural scale"
  - phase: 13-agent-loop-for-interprocedural-defects (plan 06)
    provides: "the twin-pair criterion-3 evidence and D-13-10a's use_matching_argument withdrawal this plan's checkpoint ratifies as DX-07's final disposition"
provides:
  - "A retro-strengthened testdata/phase6 distinctness control (TestPhase6DefectCorpusIsHeldOut), replacing M001's byte-inequality-only check with an identifier-independent structural predicate, per D-13-33"
  - "A real, run, recorded finding: M001's own heldout_move_defect.lang/derivation_move_defect.lang and heldout_borrow_defect.lang/derivation_borrow_defect.lang pairs are structurally identical -- a genuine hole in shipped M001 evidence -- ratified as permanent debt (D-13-34) rather than fixed or hidden"
  - "PHASE-13-DEBT.md, this phase's first debt register, recording three developer-ratified terminal findings (D-13-02b, D-13-10a, D-13-34)"
  - "REQUIREMENTS.md traceability marking DX-06 and DX-07 Partial (not Complete, not a bare Pending) with pointers to their owning debt rows"
  - "13-VALIDATION.md filled in against what actually shipped: every Per-Task Verification Map row run for real, two ungrounded -run patterns found and corrected, a new QLT-08 Mutation-Kill Completeness matrix, and a new Unresolved edge-coverage assumptions section"
affects: []

# Actuals (#2632)
actuals:
  tokens: 15053
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Identifier-independent structural summary as a distinctness predicate, at whatever granularity the corpus actually has: testdata/phase13's (function count, call-edge count, hop distance) topology triple degenerates to a constant on testdata/phase6's zero-interprocedural fixtures, so the retro-strengthened predicate there is (bindingCount, matchArmCount, borrowCount, takeCount, maxDepth) instead -- the weaker substitute the plan's own must_haves required, verified necessary rather than assumed"
    - "Per-class subtests with a known-identical allowlist, not a blanket skip: TestPhase6DefectCorpusIsHeldOut splits into one subtest per class so the one pair that discriminates (match) keeps running and failing loudly on regression, while the two known-identical pairs (move, borrow) skip with a named debt pointer and a guard that fails loudly if a class silently drifts out of the known-identical set"
    - "Debt-register-as-adjudication-record: PHASE-13-DEBT.md's three rows are not merely open questions -- each documents a developer decision made at a blocking-human checkpoint, with the rejected alternatives and their reasons named alongside the chosen one, so a later reader sees why C and D were not picked, not just that B was"

key-files:
  created:
    - .planning/phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md
  modified:
    - internal/compiler/session/session_phase6_injectors_test.go
    - .planning/phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
    - .planning/REQUIREMENTS.md

key-decisions:
  - "D-13-33 adjudication: Option B (accept as M001 evidence debt), decided by the developer at Task 2's gate=\"blocking-human\" checkpoint. testdata/phase6's move and borrow class pairs are structurally identical on all five predicate components -- alpha-rename-only, exactly the M001 weakness D-13-33 exists to close -- and are recorded as debt (D-13-34) rather than fixed with replacement fixtures (option C, new M001 work re-opening a shipped milestone's corpus) or reversed by scoping the predicate down (option D, explicitly rejected during phase discussion)."
  - "DX-07 ratified Partial: two of three classes ship as genuinely repairable (move_after_interprocedural_loan backward-direction, wrap_call_in_try). use_matching_argument withdrawn per D-13-10a -- every Lang function shares one type fact, so the uniqueness gate's only possible match is always the argument's own already-passed place, making the Replacement a byte-identical no-op on every real trigger."
  - "DX-06 ratified Partial, on the D-12-43 precedent: B1 contract-violation blame is structurally unreachable at this language maturity (D-13-02b) -- sameType(ReturnType, Parameter.Type) is an admission precondition independent of any call, so a self-contradicting callee is refused before the interprocedural pass runs. The contract-boundary rule stays as correct, exhaustively-tested infrastructure for the first B1-shaped class; criterion 3's twin pair discriminates detection-site blame from caller blame, not B1 specifically."
  - "REQUIREMENTS.md's traceability table has no existing partial-status marker, so per the developer's explicit instruction, the disposition is recorded as row text (\"Partial -- ... see PHASE-13-DEBT.md D-...\") rather than inventing a new checkbox state or marking either requirement plainly Complete."

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "testdata/phase6's distinctness control retro-strengthened with an identifier-independent structural predicate, replacing byte-inequality per D-13-33"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestPhase6DefectCorpusIsHeldOut"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestPhase6DefectCorpusDistinctnessGuardIsNotInert"
        status: pass
    human_judgment: false
  - id: D2
    description: "D-13-33's escalation carried through: move/borrow class pairs found structurally identical (real M001 evidence hole), reported without weakening the predicate or editing shipped fixtures"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestPhase6DefectCorpusIsHeldOut -v -count=1 (move/borrow subtests explicit SKIP with D-13-34 reason)"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-13-33 adjudicated at a gate=\"blocking-human\" checkpoint -- developer chose Option B and the rationale is recorded"
    human_judgment: true
    rationale: "A blocking-human decision by design (PLAN.md Task 2) -- the executor cannot answer it; the developer's verbatim choice and rationale are recorded above and in PHASE-13-DEBT.md D-13-34."
  - id: D4
    description: "DX-06 and DX-07 ratified Partial with recorded terminal findings (D-13-02b, D-13-10a), REQUIREMENTS.md traceability updated accordingly, neither marked plainly Complete"
    human_judgment: true
    rationale: "Both dispositions were explicit developer ratifications relayed at the same checkpoint exchange, not something an automated check can classify as met/unmet on its own."
  - id: D5
    description: "13-VALIDATION.md's Per-Task Verification Map corrected against real test names and re-run for every row; QLT-08 completeness matrix and Unresolved edge-coverage assumptions sections added; wave_0_complete/nyquist_compliant set from evidence"
    requirement: "QLT-08"
    verification:
      - kind: integration
        ref: "go test ./... -count=1 (25 packages, exit 0)"
        status: pass
    human_judgment: false

# Metrics
duration: ~25min active (Task 1: ~10min before the checkpoint; Tasks 2-3: ~15min after the developer's checkpoint answer -- the 17:25-20:42 gap between commits is the blocking-human wait, not active execution time)
completed: 2026-09-13
status: complete
---

# Phase 13 Plan 07: Retro-Strengthen M001's Distinctness Control and Close QLT-08 Summary

**M001's `testdata/phase6` distinctness control replaced with a structural predicate that found a real hole (move/borrow pairs are alpha-renames), ratified as permanent debt; DX-06/DX-07 ratified Partial; 13-VALIDATION.md corrected and closed.**

## Performance

- **Duration:** ~25 min active work, spanning a blocking-human checkpoint pause
- **Started:** 2026-09-13T17:21:31-04:00 (base SHA `ad3877a`)
- **Completed:** 2026-09-13T20:52:33-04:00
- **Tasks:** 3 (Task 1 auto, Task 2 checkpoint:decision, Task 3 auto)
- **Files modified:** 4 (1 created, 3 modified)

## Accomplishments

- Replaced `TestPhase6DefectCorpusIsHeldOut`'s byte-inequality-only assertion with an
  identifier-independent structural predicate `(bindingCount, matchArmCount, borrowCount,
  takeCount, maxDepth)`, per D-13-33's instruction to retro-strengthen M001's own held-out
  control, not only Phase 13's. `testdata/phase6` has zero interprocedural fixtures, so
  D-13-26's topology triple degenerates completely there — this intraprocedural predicate is
  the mandatory weaker substitute.
- Ran the predicate and recorded the actual result rather than assuming one: the `match` class
  pair is structurally distinct (3 arms vs 2) and passes; the `move` and `borrow` class pairs
  are structurally IDENTICAL on every component — the fixtures differ only in identifier
  spelling (`item`→`buffer`, `moved_once`→`delivered` for move; `item`→`buffer`,
  `alpha`/`beta`→`first`/`second` for borrow), exactly M001's own alpha-rename weakness.
- Added `TestPhase6DefectCorpusDistinctnessGuardIsNotInert`, the required D-13-30(b) mutation
  kill at intraprocedural scale: an alpha-renamed copy of `derivation_match_defect.lang` proves
  structurally equal to the original, and the real `heldout_match_defect.lang` proves
  structurally distinct from it — confirming the predicate genuinely discriminates rather than
  passing vacuously.
- At Task 2's `gate="blocking-human"` checkpoint, the developer adjudicated **Option B**: keep
  the strengthened predicate unweakened, do not edit any `testdata/phase6/` fixture, and record
  the move/borrow finding as permanent M001 evidence debt. Applied by splitting
  `TestPhase6DefectCorpusIsHeldOut` into per-class subtests — `match` keeps running and passing,
  `move`/`borrow` explicitly `t.Skip` with a reason naming the debt row, guarded so a class
  silently drifting out of the known-identical set fails loudly instead of continuing to skip
  stale debt.
- Created `PHASE-13-DEBT.md`, recording three developer-ratified terminal findings as
  mechanically-checked debt-register rows (`TestDebtRegistersAreWellFormed` passes): D-13-02b
  (B1 contract-violation blame structurally unreachable), D-13-10a (`use_matching_argument`
  withdrawn as unrepairable), and D-13-34 (this plan's own testdata/phase6 finding).
- Updated `REQUIREMENTS.md`'s traceability table: DX-06 and DX-07 marked `Partial` with a
  pointer to their owning debt row — neither left as a bare `Pending` nor marked `Complete`.
- Task 3: filled in `13-VALIDATION.md`'s Per-Task Verification Map against what actually
  shipped across all seven plans, running every command for real. Found and corrected two
  instances of the same defect class (`go test -run` exits 0 when its pattern matches
  nothing): the plan-flagged D-13-09a row (`TestOrderingStability` → real name
  `TestInterproceduralDiagnosticOrderingStability`) and a second, previously unflagged instance
  in the pre-existing import-boundary row (`-run TestImportBoundary` matched only the not-inert
  companion test, never the actual lint `TestRepairDriverImportsStayOutsideInternal`). Added
  `## QLT-08 Mutation-Kill Completeness` (all four D-13-30 seams plus D-13-33's extension of
  seam (b), plus the one seam recorded as withdrawn rather than dropped) and `## Unresolved
  edge-coverage assumptions` (DX-06/DX-07, restated for phase verification). Set
  `wave_0_complete: true` and `nyquist_compliant: true` from the evidence. Full suite green:
  `go test ./...` — 25 packages, exit 0.

## Task Commits

1. **Task 1: Retro-strengthen `testdata/phase6` distinctness with a structural predicate** —
   `28183ad` (test)
2. **Task 2: Adjudicate the retro-strengthening result** (applied after the developer's
   checkpoint answer) — `c46504f` (test)
3. **Task 3: QLT-08 completeness matrix and validation sign-off** — `aee252c` (docs)

**Plan metadata:** this commit (docs, includes this SUMMARY + STATE.md + ROADMAP.md +
REQUIREMENTS.md — REQUIREMENTS.md's DX-06/DX-07 update landed with Task 2's commit above since
it was part of applying the checkpoint's ratified disposition)

## Files Created/Modified

- `internal/compiler/session/session_phase6_injectors_test.go` — strengthened
  `TestPhase6DefectCorpusIsHeldOut` (per-class subtests, structural predicate, two named
  `t.Skip`s), added `TestPhase6DefectCorpusDistinctnessGuardIsNotInert`,
  `phase6StructuralSummary`, `computePhase6StructuralSummary`, `summarizePhase6LinearBody`,
  `alphaRenamePhase6MatchDefect`
- `.planning/phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md` — new debt
  register, 3 items, `TestDebtRegistersAreWellFormed`-conformant
- `.planning/phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md` — Per-Task
  Verification Map filled in with real plan/task numbers and re-run commands, two ungrounded
  `-run` patterns corrected, new QLT-08 completeness and unresolved-assumptions sections,
  `wave_0_complete: true`, `nyquist_compliant: true`
- `.planning/REQUIREMENTS.md` — DX-06/DX-07 traceability rows marked `Partial` with debt-row
  pointers

## Decisions Made

See `key-decisions` in the frontmatter above; full rationale for each is recorded verbatim in
`PHASE-13-DEBT.md`'s `### D-13-02b`, `### D-13-10a`, and `### D-13-34` sections.

## D-13-33 adjudication

**Choice: Option B — accept the finding as M001 evidence debt.**

**Per-pair structural summary values, run and recorded (not assumed):**

| Class | heldout summary | derivation summary | Structurally distinct? |
|-------|------------------|---------------------|--------|
| match | `{bindingCount:0 matchArmCount:3 borrowCount:0 takeCount:0 maxDepth:1}` | `{bindingCount:0 matchArmCount:2 borrowCount:0 takeCount:0 maxDepth:1}` | Yes — differs on `matchArmCount` |
| move | `{bindingCount:2 matchArmCount:0 borrowCount:0 takeCount:2 maxDepth:1}` | `{bindingCount:2 matchArmCount:0 borrowCount:0 takeCount:2 maxDepth:1}` | **No — identical on every component** |
| borrow | `{bindingCount:3 matchArmCount:0 borrowCount:3 takeCount:0 maxDepth:1}` | `{bindingCount:3 matchArmCount:0 borrowCount:3 takeCount:0 maxDepth:1}` | **No — identical on every component** |

`TestPhase6DefectCorpusDistinctnessGuardIsNotInert` proved the predicate discriminating: an
alpha-renamed copy of `derivation_match_defect.lang` came out structurally EQUAL to the
original (proving the predicate is rename-invariant, as required), while the real
`heldout_match_defect.lang` came out structurally UNEQUAL to `derivation_match_defect.lang`
(proving the predicate genuinely discriminates on the one pair that currently passes, not
vacuously).

**Rationale, recorded verbatim from the developer:** the `move` and `borrow` pairs are
structurally identical on all five components and differ only in identifier spelling — exactly
the alpha-rename weakness D-13-33 named. Byte-inequality passed and always would have. This is
a genuine hole in shipped M001 evidence, and it is being surfaced permanently rather than closed
inside a Phase 13 budget: authoring replacement M001 fixtures (option C) is new work re-opening
a shipped milestone's corpus, and scoping the predicate down (option D) would reverse the
stricter branch the developer deliberately chose during discussion.

**Applied exactly as instructed:** the strengthened predicate stays in place, unweakened; no
`testdata/phase6/` fixture was edited (`git diff --quiet -- testdata/phase6/` exits 0); only the
`move` and `borrow` class-pair assertions are `t.Skip`'d, each with a reason naming the finding
and pointing at `PHASE-13-DEBT.md` D-13-34; the `match` pair and
`TestPhase6DefectCorpusDistinctnessGuardIsNotInert` continue to run and pass. Re-run:
`go test ./internal/compiler/session/... -run TestPhase6DefectCorpus -v -count=1` →
`TestPhase6DefectCorpusIsHeldOut` PASS (`match` subtest PASS, `move`/`borrow` subtests SKIP),
`TestPhase6DefectCorpusDistinctnessGuardIsNotInert` PASS.

Full record: `PHASE-13-DEBT.md` § `D-13-34`.

## Deviations from Plan

None — plan executed exactly as written, including its explicit instruction to escalate rather
than weaken the predicate when it turned green tests red (Task 1's `<action>`), which is exactly
what happened and was then adjudicated at Task 2's checkpoint as designed.

## Issues Encountered

None. The two grounding corrections in `13-VALIDATION.md` (the plan-flagged D-13-09a row and
the previously-unflagged import-boundary row) were found by systematically running every row's
command during Task 3, as the plan's action explicitly required ("Status from an actual run,
not from the SUMMARY's claim") — not unexpected obstacles, but exactly the kind of finding this
task exists to catch.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

Phase 13 is now complete (all 7 plans have SUMMARYs). `PHASE-13-DEBT.md` carries three
open-and-unowned findings (D-13-02b, D-13-10a, D-13-34), each with a stated reopening condition
and no phase currently claiming the reopening work — consistent with this project's established
practice of recording debt rather than inventing speculative owners. `13-VALIDATION.md` is
`wave_0_complete: true`, `nyquist_compliant: true`, with every row run and green (including two
deliberate, named, debt-linked skips). `go test ./...` is fully green (25 packages). M002's
seven-phase roadmap (Phases 07–13) is complete pending `/gsd-verify-work` and
`/gsd-complete-milestone`.

---
*Phase: 13-agent-loop-for-interprocedural-defects*
*Completed: 2026-09-13*

## Self-Check: PASSED

- All 5 key files verified present on disk (`[ -f ]`).
- All 3 task commits (`28183ad`, `c46504f`, `aee252c`) verified present in `git log --oneline --all`.
- `go test ./internal/compiler/session/... -run TestPhase6DefectCorpus -v -count=1` re-run: PASS.
- `go test ./... -count=1`: PASS (25 packages, exit 0).
- `git diff --quiet -- testdata/phase6/`: exit 0 (no shipped fixture edited).
- `TestDebtRegistersAreWellFormed`: PASS (`PHASE-13-DEBT.md` conforms).

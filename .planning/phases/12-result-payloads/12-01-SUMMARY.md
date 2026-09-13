---
phase: 12-result-payloads
plan: 01
subsystem: compiler-native-emission
tags: [cgen, debt-register, convergence-test, N=1-differential, D-11-02, D-12-31, D-12-36]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: emitProgram (multi-function native emission), the D-11-02 six-emitter deletion debt row, D-11-52's Match-body collision note
provides:
  - A permanent, committed measurement (TestN1ConvergenceDifferential) answering D-11-02's Q-05 gate
  - A ratified developer decision (RE-DEFER) on whether the inherited six-emitter deletion proceeds this phase
  - A stated reversal in PHASE-11-DEBT.md superseding the Phase-12 landing-phase commitment
  - PHASE-12-DEBT.md, the phase's mechanically-checked debt register (6 items)
affects: [12-02, 12-03, 12-04, 12-05, any future emitter-convergence phase]

# Actuals (#2632)
actuals:
  tokens: 42000
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Committed convergence-differential test as permanent regression evidence, not a throwaway spike (D-12-01 shape)"
    - "Stated-reversal debt-register precedent (D-09-08/D-09-30/D-10-27): supersede prior landing-phase text with a dated sub-paragraph, quote the superseded text verbatim, never silently rewrite"

key-files:
  created:
    - internal/compiler/cgen/cgen_n1_convergence_test.go
    - .planning/phases/12-result-payloads/PHASE-12-DEBT.md
  modified:
    - .planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md

key-decisions:
  - "D-12-31 branch ratified as RE-DEFER (not DELETE): the measured N=1 differential shows emitProgram refuses 4/5 single-function shapes and diverges on the 5th, so convergence is not cheap; D-12-36's fallback trigger fired."
  - "PHASE-11-DEBT.md's D-11-02 landing-phase commitment is formally superseded with a stated reversal, not silently edited — the original text is quoted verbatim in the new sub-paragraph."
  - "The six legacy emitters (emitLinear, emitLinearBorrowedByPointer, emitLinearBorrowedByPointerPlain, emitLinearForeign, emitBranch, emitMatch) stay in cgen.go, byte-untouched, with no currently-owned landing phase for their eventual deletion."

requirements-completed: [NAT-04, NAT-05, NAT-06, NAT-07]

coverage:
  - id: D1
    description: "TestN1ConvergenceDifferential permanently records the five-shape N=1 convergence measurement (D-11-02 Q-05 gate, D-12-31 step (i))"
    requirement: "NAT-04"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_n1_convergence_test.go#TestN1ConvergenceDifferential"
        status: pass
    human_judgment: false
  - id: D2
    description: "Developer ratified the D-12-31 branch (RE-DEFER over DELETE) at the Task 2 blocking-human checkpoint"
    requirement: "NAT-04"
    verification: []
    human_judgment: true
    rationale: "A developer decision recorded verbatim in this SUMMARY, not a mechanically-checkable fact — the checkpoint's whole purpose is human judgment on a one-way branch."
  - id: D3
    description: "PHASE-11-DEBT.md's D-11-02 stated reversal and PHASE-12-DEBT.md's six-item register are both mechanically well-formed"
    requirement: "NAT-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false

duration: 38min
completed: 2026-09-12
status: complete
---

# Phase 12 Plan 01: N=1 Convergence Differential and D-12-31 Branch Ratification Summary

**Committed a permanent N=1 emitter-convergence measurement, then the developer ratified RE-DEFER over DELETE for the inherited D-11-02 six-emitter deletion, recorded as a stated reversal in both phase debt registers.**

## Performance

- **Duration:** 38 min (Task 1 + continuation for Tasks 2-3)
- **Started:** 2026-09-12T21:15:00Z (approx., prior executor)
- **Completed:** 2026-09-12T21:53:00Z
- **Tasks:** 3
- **Files modified:** 3 (1 created test file, 1 created debt register, 1 modified debt register)

## Accomplishments

- `TestN1ConvergenceDifferential` permanently records that `emitProgram` and the legacy single-function emitter family have NOT converged at N=1: `emitProgram` refuses 4 of 5 representative single-function shapes outright (`testdata/phase1/toggle.lang`, `testdata/phase3/borrowed_view.lang`, `testdata/phase4/foreign_acquire_one.lang`, `testdata/phase4/defect_terminal.lang`) and produces textually- and structurally-divergent output on the 5th (`testdata/phase2/owned_transfer.lang`).
- The developer ratified the D-12-31 branch decision as **RE-DEFER**, firing D-12-36's pre-registered fallback rather than proceeding with the six-emitter deletion inside Phase 12.
- PHASE-11-DEBT.md's `D-11-02` row and detail section carry a stated reversal, quoting the superseded landing-phase text verbatim and citing `TestN1ConvergenceDifferential` as evidence, per the D-09-08/D-09-30/D-10-27 precedent.
- `PHASE-12-DEBT.md` was created and is mechanically well-formed (`TestDebtRegistersAreWellFormed` green), carrying D-12-24 (niche optimization, uninstantiable), D-12-30 (resource-payload three-part landing condition), D-12-21 (D-11-51 interaction flag), D-12-42 (QLT-09 deferral), D-12-04c (stale fixture header), and D-12-36 (the ratified re-deferral itself).

## Task Commits

Each task was committed atomically:

1. **Task 1: Commit the N=1 convergence differential as a permanent test** - `1bac938` (test)
2. **Task 2: Ratify the D-12-31 branch** - checkpoint:decision, no code commit (decision recorded in this SUMMARY, per plan)
3. **Task 3: Record the ratified outcome in PHASE-11-DEBT.md and create PHASE-12-DEBT.md** - `60c16d9` (docs)

**Plan metadata:** (this commit, following SUMMARY creation)

## Files Created/Modified

- `internal/compiler/cgen/cgen_n1_convergence_test.go` - Permanent five-fixture N=1 convergence differential test, package `cgen` internal test, no `session` import
- `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` - New phase debt register, 6 items, mechanically well-formed
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md` - `D-11-02` row's `Landing phase` cell updated; detail section gained a dated stated-reversal sub-paragraph

## Task 2 Checkpoint: Developer Decision

**Ratified branch: RE-DEFER**

**Developer's stated reasoning (verbatim):**

> RE-DEFER
>
> Fire D-12-36's pre-registered fallback: re-defer the six-emitter deletion with a stated reversal superseding D-11-02's landing-phase text in PHASE-11-DEBT.md. Justified by the measured differential — emitProgram cannot express a match body at all, so payload `case` arms stay exclusively legacy-family territory and the "two coexisting laws" tax D-12-31 feared does not materialize. Plans 02-05 as written are scoped to this branch, so the phase proceeds as planned.

**The DELETE option was explicitly presented and explicitly not chosen.** Its stated consequence (converging `emitProgram` with the legacy family — porting branch bodies, foreign-call block bodies, and both by-pointer lowering variants, ~1,500 lines of `cgen.go`, re-pinning four frozen golden-C digests — very likely requiring an additional phase or a phase split) was presented alongside RE-DEFER and rejected by the developer in favor of RE-DEFER.

**Measured differential facts that grounded the decision** (from Task 1's `TestN1ConvergenceDifferential`):

| Fixture | legacy `Emit` | `emitProgram` | identical? |
|---|---|---|---|
| `testdata/phase1/toggle.lang` | ok | refused: multi-function branch bodies are not supported | no |
| `testdata/phase2/owned_transfer.lang` | ok | ok | **no — diverges textually and structurally** |
| `testdata/phase3/borrowed_view.lang` | ok | refused: multi-function branch bodies are not supported | no |
| `testdata/phase4/foreign_acquire_one.lang` | ok | refused: multi-function foreign-call bodies are not supported | no |
| `testdata/phase4/defect_terminal.lang` | ok | refused: multi-function branch bodies are not supported | no |

Because the DELETE branch was not authorized, no follow-on phase or phase split was named — that acceptance criterion applies only to the DELETE branch.

## Decisions Made

- **D-12-31 branch: RE-DEFER, not DELETE.** See Task 2 checkpoint section above for full reasoning.
- **PHASE-11-DEBT.md's D-11-02 landing-phase commitment is superseded, not deleted.** The original text ("Phase 12 — a green N=1 convergence differential (Q-05), landing inside Phase 12's per-dispatch-site `Result` plans, never as a second sweep") is quoted verbatim in the stated-reversal sub-paragraph, per this project's D-09-08/D-09-30/D-10-27 precedent for how a reversal must be recorded (never a silent rewrite).
- **The six legacy emitters have no currently-owned landing phase.** `D-11-02`'s row and `D-12-36`'s row both record `OPEN and UNOWNED` — no phase currently claims the ~1,500-line convergence-porting work. A future phase would need to explicitly re-open this.

## Deviations from Plan

None - plan executed exactly as written. The RE-DEFER branch was one of the two pre-registered options in Task 2's `<options>` block, and Task 3's action text explicitly conditions the `D-12-36` row and the PHASE-11-DEBT.md stated-reversal on this exact outcome.

## Issues Encountered

None. The continuation picked up cleanly from the prior executor's Task 1 commit (`1bac938`, verified present with `git show --stat` before proceeding) and the human's RE-DEFER answer to the Task 2 checkpoint.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plans 02-05, as written, are scoped to the RE-DEFER branch and can proceed without modification.
- The six legacy emitters (`emitLinear`, `emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`, `emitLinearForeign`, `emitBranch`, `emitMatch`) remain byte-untouched in `cgen.go`; the four pinned golden-C digests are unaffected.
- `TestN1ConvergenceDifferential` will flip to a byte-identity assertion automatically whenever future convergence work lands — no test deletion needed, per its own doc comment.
- The re-deferred D-11-02/D-12-36 deletion has no owning phase. A future phase that wants to revisit this must explicitly claim it; it is not scheduled.

## Self-Check: PASSED

- FOUND: internal/compiler/cgen/cgen_n1_convergence_test.go
- FOUND: .planning/phases/12-result-payloads/PHASE-12-DEBT.md
- FOUND: .planning/phases/12-result-payloads/12-01-SUMMARY.md
- FOUND commit: 1bac938 (Task 1)
- FOUND commit: 60c16d9 (Task 3)
- FOUND commit: 5a94b42 (SUMMARY)

---
*Phase: 12-result-payloads*
*Completed: 2026-09-12*

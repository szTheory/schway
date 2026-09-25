---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "05"
subsystem: testing
tags: [prc-02, debt-registers, evidence, session-tests]
requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: PRC-01 debt register parser and generated unreachable-claims view
  - phase: 16-branch-match-emitter-port
    provides: public emitter cutover and M004 refusal evidence
provides:
  - Source-derived PRC-02 population with exact historical and live M003 cohorts
  - Evidence-backed Phase 20/21 dispositions and a nine-ID current set
  - Pure five-item cap predicate with a seeded six-item refusal control
affects: [20-06, 20-07, PRC-02]
actuals:
  tokens: 0 # Executor token telemetry was unavailable in the shared-worktree handoff.
  tasks: 3
  commits: 1 # One combined named-file implementation commit; Tasks 1 and 3 share session_test.go.
plan_head_before: 06aa0fe1189d7ee8cf2937c0636c4e1cc2e5c069
tech-stack:
  added: []
  patterns: [source-derived debt cohort, pure cap predicate, generated-view regeneration]
key-files:
  created:
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-DEBT-BASELINE.md
  modified:
    - internal/compiler/session/session_test.go
    - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
    - .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md
    - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
    - .planning/UNREACHABLE-CLAIMS.md
    - .planning/ROADMAP.md
key-decisions:
  - "Use the M002 audit's ten exact carry-forward IDs plus seven live M003 additions; count only current source-register UNOWNED dispositions."
  - "Assign residual legacy-emitter and LTO evidence work to Phase 21 without claiming the retained legacy functions were deleted or that LTO is non-inert."
  - "Keep D-10-C04's negative-control verdict review open and D-13-34 pending Plan 06's human decision."
requirements-completed: [PRC-02]
coverage:
  - id: D1
    description: The PRC-02 denominator is derived from the exact M002 carry-forward and current source-register dispositions.
    requirement: PRC-02
    verification:
      - kind: unit
        ref: internal/compiler/session.TestPhase20UnownedDebtPopulation
        status: pass
    human_judgment: false
  - id: D2
    description: The five-item cap is a live pure predicate that rejects a seeded sixth open-unowned row.
    requirement: PRC-02
    verification:
      - kind: unit
        ref: internal/compiler/session.TestPhase20UnownedDebtCapRule
        status: pass
    human_judgment: false
  - id: D3
    description: Register evidence and the generated unreachable-claims view remain consistent after the ownership updates.
    verification:
      - kind: unit
        ref: internal/compiler/session.TestDebtRegistersAreWellFormed, TestUnreachableClaimsViewIsCurrent
        status: pass
    human_judgment: false
duration: pending serialized closeout
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 05: Source-Derived Debt Cap Summary

**PRC-02 now derives its starting liability from ten audited carry-forward IDs plus seven live M003 additions, and a seeded sixth-item control proves the five-item cap predicate refuses overflow.**

## Accomplishments

- Added source-derived cohort/disposition parsing that rejects missing, duplicate, or unknown qualified provenance and checks the M002 audit's exact ten-ID table.
- Pinned the phase-start population at 13 unique open-unowned IDs and documented why the generated eight-row claims view and 40 raw archive cells have different scopes.
- Assigned residual emitter/convergence work to Phase 21 and D-11-27 applicability review to Phase 20 while preserving the separate, still-open D-10-C04 verdict review.
- Assigned D-14-45's measured one-TU/LTO evidence boundary to Phase 21. The evidence continues to say LTO is inert for emitted multi-function programs.
- Regenerated `.planning/UNREACHABLE-CLAIMS.md` through its existing derive/render test helpers and pinned the resulting current qualified set at nine IDs.
- Added a pure cap predicate and a seeded six-ID failure control. The four remaining QLT-10 repair IDs D-14-50/51/52/54 give Plan 07 a route from nine to five without relying on D-13-34's human decision.

## Verification

Passed with `GOCACHE=/private/tmp/phase20-gocache`:

- `go test ./internal/compiler/session -run 'TestDebtRegistersAreWellFormed|TestPhase20UnownedDebt(Population|CapRule)|TestUnreachableClaimsViewIsCurrent' -count=1`
- `go test ./internal/compiler/check -run '^TestInterproceduralDisclosedFieldSet$' -count=1`
- `go test ./internal/compiler/cgen -run '^TestSupersededEmitterDefinitionsRemoved$' -count=1`
- `go test ./internal/compiler/session -run '^TestPhase16ProductionPathsPreserveM004Refusal$' -count=1`
- `git diff --check` on Plan 05 files

The Phase 16 emitter test confirms removal of `emitLinear`, `emitBranch`, and `emitMatch`. Three retained implementations remain in `cgen.go`; they have no public production call path and their corresponding cut families are explicitly owned by Phase 21. The record does not claim six definitions were deleted. The Phase 16 production-refusal test and the checker control both pass; the checker still pins `core.callee_not_callable` for both historical negative controls. D-10-C04 remains the outstanding review of that diagnostic verdict.

## Task Commits

The shared `session_test.go` file is changed by Tasks 1 and 3, and the baseline records Task 2/3 dispositions. Per the orchestrator's shared-index serialization, the implementation changes were captured in one named-file Plan 05 commit (`4fe505a`) because the tasks share `session_test.go`; this summary will be committed separately after self-check.

## Deviations from Plan

**Evidence-scoped adjustment:** Phase 16 did not delete all six historical function definitions. The source and test show three old definitions remain, although the public path refuses the cut families. Their residual scope is assigned to the already-named Phase 21 owner; no deletion is claimed. D-11-27 is owned by P20 as an applicability check, while D-10-C04 stays open for its distinct human review.

## Next Phase Readiness

The current qualified open-unowned IDs are D-10-C04, D-12-43, D-13-34, D-14-46, D-14-47, D-14-50, D-14-51, D-14-52, and D-14-54. Plan 07's four validation citation repairs reduce this to five. D-13-34 remains pending the Plan 06 decision checkpoint.

## Self-Check

- [x] All three tasks are accounted for; overlapping edits share one implementation commit.
- [x] Required focused tests and generated-view checks passed.
- [x] D-13-34 and D-10-C04 remain visibly open as planned.
- [x] `git diff --check` passed on the Plan 05 changes.
- [x] Implementation commit `4fe505a` contains only Plan 05 declared files.

Executor token telemetry and elapsed duration were not available in the shared-worktree handoff; no estimate is substituted.

---
*Phase: 20-nyquist-d-13-34-and-the-frontier-fixture*
*Completed: 2026-09-25*

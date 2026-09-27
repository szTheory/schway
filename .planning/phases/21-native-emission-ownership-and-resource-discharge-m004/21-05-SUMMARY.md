---
phase: 21-native-emission-ownership-and-resource-discharge-m004
plan: 05
subsystem: compiler-emission
tags: [go, cgen, session, archived-artifacts, refusal-guards]
requires:
  - phase: 16
    provides: Human-selected cut-m004 disposition for the by-pointer emitter families
  - phase: 21-04
    provides: Archived emitter ownership records and the provisional Phase 21 roadmap entry
provides:
  - By-pointer whole-program refusal guard reads the archived Phase 16 decision
  - NAT-09 ownership guard joins archived M003 requirements and Phase 16 debt to the current provisional Phase 21 owner
affects: [phase-21-06, phase-21-verification, emitter-debt, archive-aware-tests]
actuals:
  tokens: 930
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [archive-aware testsupport.ProjectPath lookups, mutation-sensitive ownership guards]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/session/session_test.go
key-decisions:
  - "D-21-01 and D-21-02 remain design-only; foreign and by-pointer emitters stay refused until family-specific proof and witnesses exist."
patterns-established:
  - "Recurring planning guards resolve tracked M003 artifacts from their milestone archive."
  - "Seeded roadmap mutations assert that the fixture actually changed before testing rejection."
requirements-completed: []
coverage:
  - id: D1
    description: "The by-pointer disposition guard reads the archived Plan 16-05 decision and confirms cut-m004 refusal before C serialization."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/cgen -run '^TestProgramBorrowedByPointerDisposition$' -count=1 -v"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/cgen -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The emitter-cut ownership guard joins archived NAT-09, the Phase 16 debt register, and the provisional Phase 21 roadmap owner while rejecting seeded mutations."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestPhase16EmitterCutsAreAmendedAndOwned$' -count=1 -v"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session -count=1"
        status: fail
    human_judgment: true
    rationale: "The focused guard passes, but the full session suite still has post-archive corpus and reconciliation failures that require classification after remaining gap closure."
duration: 10min
completed: 2026-09-26
status: complete
plan_head_before: a467c55a5995f5d11bb29b4c62a4c7d0a6d1ad2a
---

# Phase 21 Plan 05: Archived Phase 16 Decision and Ownership Guards Summary

**Archive-aware by-pointer refusal and NAT-09 ownership guards now bind to tracked M003 records and Phase 21's provisional roadmap registration.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-09-26T23:53:47Z
- **Completed:** 2026-09-27T00:03:47Z
- **Tasks:** 2/2
- **Files modified:** 2

## Accomplishments

- Updated `TestProgramBorrowedByPointerDisposition` to read the tracked archived Plan 16-05 decision while retaining multi-function refusal and pre-serialization assertions.
- Updated `TestPhase16EmitterCutsAreAmendedAndOwned` to read archived M003 requirements, recognize the current shipped M003 and provisional M004 roadmap entries, and verify the seeded owner mutation changes the Phase 21 heading.
- Kept the foreign and by-pointer emitter families refused; no M004 requirement IDs or production emitter admission were added.

## Task Commits

1. **Task 1: Resolve the archived cut decision through the whole-program refusal guard** — `b9a7201` (test)
2. **Task 2: Bind emitter-cut ownership to archived NAT-09 and the current roadmap registration** — `1c42a7e` (test)

**Plan metadata:** summary commit recorded in the completion report.

## Files Created/Modified

- `internal/compiler/cgen/cgen_program_test.go` — points the by-pointer disposition guard at archived Plan 16-05 evidence.
- `internal/compiler/session/session_test.go` — binds the NAT-09 guard to archived M003 requirements and current provisional roadmap state.

## Decisions Made

- Preserved the Phase 21 design-only boundary: ownership records do not authorize foreign or by-pointer emitter admission without family-specific proof and witnesses.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestProgramBorrowedByPointerDisposition$' -count=1 -v` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase16EmitterCutsAreAmendedAndOwned$' -count=1 -v` — passed, including the four seeded rejection controls.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -count=1` — failed on remaining post-archive validation corpus and groundedness/reconciliation findings, including a stale Phase 20 debt artifact path. The focused ownership guard itself passed.
- `git diff --check` — passed before both task commits.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The full session package suite still fails on archive-dependent checks outside this plan's two target guards: the validation corpus pair digest is stale, the Phase 20 debt baseline path is unresolved, and archived validation/reconciliation findings remain. Plan 21-06 addresses the corpus and groundedness gap; these failures were left for their owning closure work. No changes were made to `STATE.md` or `ROADMAP.md`.

## Known Stubs

None found in the modified files.

## Next Phase Readiness

- Plan 21-05 is complete. Continue with Plan 21-06, then refresh Phase 21 verification while preserving the completed UAT.

## Self-Check: PASSED

- Summary file exists and both task commits are present in git history.
- Coverage metadata passes `uat classify-coverage`; production commit count is measured as 2 from the plan baseline.

---
*Phase: 21-native-emission-ownership-and-resource-discharge-m004*
*Completed: 2026-09-26*

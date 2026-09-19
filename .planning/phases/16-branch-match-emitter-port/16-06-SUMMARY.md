---
phase: 16-branch-match-emitter-port
plan: "06"
subsystem: native-emission-debt
tags: [c17, cgen, m004, debt-register, nat-09]
requires:
  - phase: 16-branch-match-emitter-port
    provides: Plan 16-05 cut-m004 decision and cross-host evidence disposition
provides:
  - Decision-linked whole-program refusal for every pointer-specialized family
  - D-10-60 NAT-09 amendment and bijective M004 emitter-family debt register
affects: [16-07, 16-08, 16-09, M004-native-emission]
tech-stack:
  added: []
  patterns: [decision-linked admission tests, shared debt-table parser, seeded amendment/debt mismatch control]
key-files:
  created:
    - .planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/cgen_n1_convergence_test.go
    - internal/compiler/session/session_test.go
    - .planning/REQUIREMENTS.md
key-decisions:
  - "Applied the human-selected cut-m004: no pointer-specialized family is admitted by emitProgram at any program cardinality."
  - "Recorded foreign, exclusive-pointer, and shared-pointer families as explicit M004 debt; one-TU pure-Lang -flto remains behaviorally inert evidence."
actuals:
  tokens: 4582
  tasks: 2
  commits: 3
commits: 3
plan_head_before: e7eb33edc4aab4af6b4fc3b7298a447d0702b793
duration: 12 min
completed: 2026-09-19
status: complete
---

# Phase 16 Plan 06: M004 Emitter Cuts Summary

**The program emitter now enforces the reviewed `cut-m004` boundary for every pointer-specialized body, while NAT-09 names each excluded family and its M004 reopening work.**

## Accomplishments

- Extended the whole-program pointer refusal to multi-function programs and linked it directly to the recorded Plan 16-05 decision.
- Added the D-10-60 NAT-09 amendment and a three-row Phase 16 debt register for foreign, exclusive-pointer, and shared-pointer lowering.
- Added a semantic session control that uses the shared debt-table parser and kills a seeded amendment/debt family mismatch.
- Updated the N=1 convergence characterization for match and defect bodies now admitted by `emitProgram`, while retaining their expected schema-output difference.

## Verification

- `go test ./internal/compiler/cgen -run 'TestProgramBorrowedByPointerDisposition|TestRestrictAdmissionRejectsExtensions|TestUnsupportedProgramShapePrecedesSchema2Preflight' -count=1 -v`
- `go test ./internal/compiler/cgen -count=1`
- `go test ./internal/compiler/session -run 'TestDebtRegistersAreWellFormed|TestPhase16EmitterCutsAreAmendedAndOwned' -count=1 -v`

## Task Commits

1. **Task 1: Implement the recorded exact-readonly admission or named cut** — `c33aebd`
2. **Task 2: Amend NAT-09 and create owned M004 debt rows** — `b98f283`
3. **Convergence repair required by the selected cut** — `065a202`

## Deviations from Plan

### Auto-fixed Issues

1. **[Rule 1 - Bug] Corrected the seeded mismatch mutation and convergence expectations**
   - **Found during:** Task 2 and package-wide cgen verification
   - **Issue:** The initial seeded replacement retained the expected-family substring, and the intentionally-red N=1 test plus two pointer-shaped Phase 11 gate users no longer matched the selected whole-program refusal boundary.
   - **Fix:** Seeded a genuinely missing family name, flipped the already-planned match/defect convergence expectations, and moved the two ordinary-output checks to an admitted multi-function fixture.
   - **Files modified:** `internal/compiler/session/session_test.go`, `internal/compiler/cgen/cgen_n1_convergence_test.go`, `internal/compiler/cgen/cgen_program_test.go`
   - **Commit:** `065a202`

## Known Stubs

None.

## Self-Check: PASSED

- Confirmed all six changed artifacts exist.
- Confirmed task commits `c33aebd`, `b98f283`, and `065a202` exist in git history.

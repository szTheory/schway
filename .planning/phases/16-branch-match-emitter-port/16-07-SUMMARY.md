---
phase: 16-branch-match-emitter-port
plan: "07"
subsystem: native-emission-evidence
tags: [c17, schema-2, provenance, differential, lto]
requires:
  - phase: 16-branch-match-emitter-port
    provides: Plan 16-06 cut-m004 admission boundary and ported branch/match emitter
provides:
  - Both-mode five-shape pre-cut protocol-divergence characterization
  - Four-entry stateful golden-C provenance ledger
  - Direct emitProgram schema-2 four-tier semantic differential
affects: [16-08, 16-09, native-emission]
actuals:
  tokens: 48905
  tasks: 3
  commits: 3
commits: 3
plan_head_before: 4a555808ac5513551ff4fef3fbb8bb93b4b958a9
tech-stack:
  added: []
  patterns: [pre-cut protocol characterization, stateful golden provenance, direct-emitter differential]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen_n1_convergence_test.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/core/core_test.go
    - internal/compiler/session/session_phase11_differential_test.go
    - internal/compiler/session/witness_registry_test.go
key-decisions:
  - "Record the real legacy protocol variants against direct schema-2 output before Plan 16-09, rather than assert false byte identity."
  - "Keep pre-cut golden entries free of guessed post-cut hashes; require one coherent post-cut transition."
  - "Use direct emitProgram(..., true) C for semantic tiers, with the bare-match return projected into schema-2 comparison facts."
requirements-completed: [NAT-08]
coverage:
  - id: D1
    description: "The five-shape characterization covers both public modes and records the intentional legacy-versus-direct protocol relation."
    requirement: NAT-08
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/cgen -run 'TestN1ConvergenceDifferential|TestProgramBranchValidationOrder|TestExistingEmittersAreByteIdentical' -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: "Golden-C provenance distinguishes real pre-cut state from the atomic post-cut baseline."
    requirement: NAT-08
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/core -run 'TestPreviousPhaseGoldenCUnchanged|TestPhase16GoldenChangeLedger|TestPhase16GoldenChangeLedgerRejectsFaults' -count=1 -v"
        status: pass
    human_judgment: false
  - id: D3
    description: "Direct-program schema-2 C agrees with the interpreter across O0, O3, and O3-LTO for the three admitted fixtures."
    requirement: NAT-08
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestPhase16DirectProgramFourTierDifferential/(Toggle|OwnedTransfer|BorrowedView)|TestPhase16EmitterPortSemanticGuardIsNotInert|TestLTOInertnessOnMultiFunctionEmission' -count=1 -v"
        status: pass
    human_judgment: false
duration: 32 min
completed: 2026-09-19
status: complete
---

# Phase 16 Plan 07: Pre-Cutover Emitter Evidence Summary

**The retained N=1 paths now document their real legacy protocol divergence from direct schema-2 emission, bind every frozen golden-C file to an honest cut state, and execute direct-program C through all four semantic tiers.**

## Accomplishments

- Expanded the five-row convergence characterization across `Emit` and `EmitNative`, with exact bytes and SHA-256 values in failure output while retaining foreign and M004 cut refusals.
- Added a four-entry golden ledger that rejects stale, duplicate, malformed, premature, and map-mismatched evidence without pre-pinning Plan 16-09 output.
- Added a narrow direct-program native seam, schema-2 bare-match return evidence, and direct C four-tier comparisons for Toggle, OwnedTransfer, and BorrowedView.

## Task Commits

1. **Task 1: Guard the retained five-shape pre-cutover characterization** — `bd80334` (test)
2. **Task 2: Establish stateful four-entry golden-C provenance without pre-pinning post-cut bytes** — `54e8df3` (test)
3. **Task 3: Exercise the port through the independent four-tier comparator** — `258aada` (feat)

## Decisions Made

- The current protocol relation is intentionally non-identical: match source mode is plain output, legacy native match is `/0`, legacy linear/branch modes are `/1`, and direct program emission is `/2`.
- A post-cut ledger may honestly retain unchanged digests, but only after all four rows, files, and map values transition coherently in Plan 16-09.
- The bare-match direct route records a schema-2 return event; its comparator path avoids pretending that the operation-only schema peer can classify an event with no `core.LinearOperation` source.

## Deviations from Plan

### Auto-fixed Issues

1. **[Rule 1 - Bug] Made direct bare-match schema-2 execution observable and valid**
   - **Found during:** Task 3
   - **Issue:** Direct `emitProgram(..., true)` emitted no event for a bare match, so Toggle could not pass native execution validation.
   - **Fix:** Emit a selected-arm schema-2 return event and include it in capacity and output-size derivation.
   - **Files modified:** `internal/compiler/cgen/cgen_program.go`, `internal/compiler/session/session_phase11_differential_test.go`
   - **Commit:** `258aada`

2. **[Rule 3 - Blocking verification debt] Replaced the obsolete multi-function match refusal witness**
   - **Found during:** Task 3 focused verification
   - **Issue:** `TestLTOInertnessOnMultiFunctionEmission` still required the refusal that Plan 16-03 intentionally removed.
   - **Fix:** Verify current schema-2 one-TU admission while retaining the explicit no-LTO-axis-movement limitation.
   - **Files modified:** `internal/compiler/session/witness_registry_test.go`
   - **Commit:** `258aada`

## Verification

- Focused cgen, core, and session gates passed.
- `go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/session -count=1` passed.

## Next Phase Readiness

Plan 16-09 alone owns the public dispatch flip, legacy-emitter deletion, and resulting byte-identity assertion. Pointer-specialized and foreign shapes remain explicit M004 exclusions.

## Self-Check: PASSED

- Found all five modified source/test files.
- Found task commits `bd80334`, `54e8df3`, and `258aada` in git history.

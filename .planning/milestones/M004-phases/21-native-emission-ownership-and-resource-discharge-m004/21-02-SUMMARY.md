---
phase: 21-native-emission-ownership-and-resource-discharge-m004
plan: 02
subsystem: compiler-emission
tags: [cgen, refusal-boundary, emitter-retirement, ast-regression]
requires:
  - phase: 16
    provides: Whole-program emission refusal boundary and family-specific reopening conditions
provides:
  - Retirement of three unreachable private foreign/by-pointer lowering bodies and legacy resource-ledger emitters
  - AST regression guard preserving sole public dispatch and absent private bodies
  - Current-tense documentation aligned with whole-program refusal behavior
affects: [phase-21, native-emission, verification]
actuals:
  tokens: 3600
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [AST-guarded production dispatch and emitter absence]
key-files:
  created:
    - internal/compiler/cgen/cgen_program_test.go
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/core/core.go
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interptestdirect/interptestdirect.go
    - internal/compiler/native/foreign_nonlocal.go
    - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
    - internal/compiler/session/session_phase5_alias.go
    - internal/compiler/session/session_test.go
key-decisions:
  - "Retire unreachable program-lowering bodies and their exclusive helpers while retaining metadata/classification APIs and all refusal boundaries."
  - "Treat archived foreign and by-pointer fixtures as historical evidence, not proof of current emitted behavior."
patterns-established:
  - "Recurring AST checks pin public dispatch and keep retired family bodies absent."
requirements-completed: []
coverage:
  - id: D1
    description: "Both public emitter APIs retain emitProgram as their sole lowering authority; the three legacy family bodies are absent."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/cgen -run 'TestPhase21LegacyEmitterBodiesRetired|TestPublicDispatchUsesOnlyEmitProgram|TestProgramBorrowedByPointerDisposition|TestProgramBranchValidationOrder' -count=1 -v"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/cgen -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Current foreign and by-pointer program shapes continue to fail at the structural refusal boundary before serialization."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/cgen -run 'TestProgramBorrowedByPointerDisposition|TestForeignLandingPadEmitterRemainsRefused|TestForeignResourceLedgerEmitterRemainsRefused|TestByPointerAttributeMetadataRemainsSeparateFromBodyAdmission' -count=1 -v"
        status: pass
    human_judgment: false
duration: 11min
completed: 2026-09-26
status: complete
plan_head_before: e9e08fd
commits: 2
---

# Phase 21 Plan 02: Emitter Retirement Summary

**Unreachable foreign and by-pointer program emitters are removed, with the single whole-program refusal path pinned by recurring AST and behavior checks.**

## Performance

- **Duration:** 11 min
- **Tasks:** 2/2
- **Files modified:** 10

## Accomplishments

- Added a red/green AST guard that pins `Emit` and `EmitNative` to `emitProgram` and rejects declarations for the three retired lowering bodies.
- Removed the unreachable foreign/by-pointer emitters, their legacy resource-ledger and nonlocal-pad output helpers, and helpers used only by those paths.
- Kept public foreign manifest utilities and all current unsupported-shape refusals intact; the full cgen package passes.
- Reworded comments and tests so archived fixtures are identified as historical evidence rather than current output.

## Task Commits

1. **Task 1: Retire legacy family bodies under an AST-pinned refusal boundary** — `e9e08fd` (test), `3204db7` (refactor)
2. **Task 2: Run focused refusal checks and the cgen package suite** — verified in `3204db7` (shared implementation commit; no additional source change was needed)

**Plan metadata:** created separately after production commits.

## Verification

- Focused dispatch, shape-refusal, and retired-body tests — passed.
- `GOCACHE=/private/tmp/ai-lang-phase21-gocache go test ./internal/compiler/cgen -count=1` — passed (`ok github.com/codename-lang/lang/internal/compiler/cgen`).
- `git diff --check` — passed.

## Deviations from Plan

- Updated `internal/compiler/interp/interp.go` and `internal/compiler/session/session_phase16_frozen_evidence_external_test.go` comments in addition to the listed files. Removing dead helpers made those references stale; changes clarify documentation only and do not alter behavior.

**Total deviations:** 1 documentation-only scope adjustment.
**Impact on plan:** Required to keep references accurate after helper retirement; no admission or runtime behavior changed.

## Issues Encountered

- No production callers of the retired bodies existed. The cgen suite passed after removal without fixes.

## Next Phase Readiness

- Plan 21-03 can now produce the scoped compiler comparison against stable emitter behavior while the refusal boundary is structurally pinned.

## Self-Check: PASSED

- The AST guard, focused refusal tests, and full cgen package suite pass.
- Both production commits are present after the recorded plan baseline.

---
*Phase: 21-native-emission-ownership-and-resource-discharge-m004*
*Completed: 2026-09-26*

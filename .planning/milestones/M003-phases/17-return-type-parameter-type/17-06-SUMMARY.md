---
phase: 17-return-type-parameter-type
plan: 06
subsystem: native C17 whole-program lowering
tags: [go, cgen, c17, native, differential-testing]
requires:
  - phase: 17-05
    provides: checked Resource-to-Result tracer and directional core facts
provides:
  - Directional C parameter/return type lowering through the sole production emitter
  - Public-emitter four-tier Resource-to-Result native differential
affects: [TYP-01, internal/compiler/cgen, internal/compiler/session]
tech-stack:
  added: []
  patterns: [parallel-C-type-tables, public-emitter-four-tier-differential, comparator-mutation-control]
key-files:
  created:
    - internal/compiler/session/session_phase17_native_test.go
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
decisions:
  - The C ABI name comes from each checked declared parameter or return type, while match representation remains derived from the checked input shape so native and interpreter execution agree.
  - Input-only generated type-name renderers are referenced from main to satisfy the strict native C warning policy without restoring an emitter family.
metrics:
  duration: 20 min
  completed: 2026-09-22
  tasks: 2
  files: 3
status: complete
plan_head_before: f291def2f919c702be61a299b514e690c402f986
commits: 4
actuals:
  tokens: 7188
  tasks: 2
  commits: 4
---

# Phase 17 Plan 06: Two-Type Native Lowering Summary

The sole whole-program C17 emitter now carries Resource parameter and Result return types independently from prototypes through entry rendering, with public-route execution agreeing across interpreter, O0, O3, and O3-LTO.

## Accomplishments

- Replaced the single whole-program type table with parameter and return C-type tables, including prototypes, definitions, OpCall targets, and entry input/output.
- Added exact generated-C guards and seeded prototype, call-target, and entry-output swaps.
- Added a public `EmitNative` four-tier differential for the checked-in Resource-to-Result tracer and a return-renderer mutation control that moves the terminal-outcome comparator axis.

## Task Commits

1. **Task 1: Carry a C parameter/return type pair through emitProgram** — `b49e8ae` (`test`), `0ac421b` (`feat`)
2. **Task 2: Prove four-tier semantic agreement for Resource -> Result** — `2903265` (`test`), `f2025fa` (`fix`)

## Verification

- Passed: `go test ./internal/compiler/cgen -run 'TestPhase17Program(TwoTypePrototype|TwoTypeDefinition|TwoTypeCall|TwoTypeEntryIO|TypePairMutation)' -count=1 -v`
- Passed: `go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase17(TwoTypeGeneratedC|TwoTypeFourTierDifferential|TwoTypeEmitterGuardIsNotInert)' -count=1 -v`
- Passed: `go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/session -run 'TestN1ConvergenceDifferential|TestPhase16GoldenChangeLedger|TestPhase16EmitterPortSemanticGuardIsNotInert|TestPhase16EmitterCutsAreAmendedAndOwned' -count=1 -v`

## Deviations from Plan

### Auto-fixed Issues

1. [Rule 1 - Bug] Restored strict-C compilation and execution for distinct entry input/output types.
- **Found during:** Task 2 four-tier native run.
- **Issue:** The Resource name renderer was unused under `-Werror`, and using the Result declaration's alternatives changed the existing match runtime representation.
- **Fix:** Referenced input-only type renderers from `main` and retained the checked match input representation beneath the distinct Result ABI type name.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Commit:** `f2025fa`

## Verification Notes

An unscoped `go test ./internal/compiler/cgen -count=1` also reports the existing frozen-evidence failure `TestLegacyEmitterEvidence: generated program digest mismatch for "phase5.enum_chain_byte_l0_e0_tfalse"`. The plan-required focused and aggregate regressions above pass; no Phase 16 frozen artifacts were changed.

## Self-Check: PASSED

- Found all three plan artifacts and task commits `b49e8ae`, `0ac421b`, `2903265`, and `f2025fa` in repository history.

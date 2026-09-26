---
phase: 21-native-emission-ownership-and-resource-discharge-m004
plan: 03
subsystem: compiler-evidence
tags: [clang, lto, semantic-comparison, provenance]
requires:
  - phase: 14
    provides: D-14-45 multi-function emission evidence target and honest scope
provides:
  - Opt-in direct emitted-C four-lane semantic comparison
  - Host/toolchain/fixture/source provenance receipt with explicit evidence ceiling
affects: [phase-21, native-emission, verification]
actuals:
  tokens: 2400
  tasks: 3
  commits: 2
tech-stack:
  added: []
  patterns: [Opt-in compiler evidence test reusing existing runner and comparator]
key-files:
  created:
    - internal/compiler/session/session_phase21_lto_evidence_test.go
    - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md
  modified: []
key-decisions:
  - "Use the existing Phase 14 multi-function fixture and run direct emitProgram output through the shared interpreter/-O0/-O3/-O3 -flto comparator."
  - "Keep compiler measurement opt-in; behavioral equality does not prove optimizer activity, performance, cleanup, or other hosts/toolchains."
patterns-established:
  - "Compiler evidence receipts include fixture and generated-source digests plus exact host and compiler identity."
requirements-completed: []
coverage:
  - id: D1
    description: "One checked emitted multi-function fixture is compared across interpreter, -O0, -O3, and -O3 -flto using the existing all-pairs comparator."
    verification:
      - kind: integration
        ref: "go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: "The comparator's independent seeded-disagreement control remains effective in the regular suite."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestPhase16EmitterPortSemanticGuardIsNotInert$' -count=1 -v"
        status: pass
    human_judgment: false
duration: 10min
completed: 2026-09-26
status: complete
plan_head_before: 50b6ec4
commits: 2
---

# Phase 21 Plan 03: Scoped LTO Evidence Summary

**A one-shot emitted multi-function comparison now records four-lane semantic agreement with exact fixture, generated-C, Clang, and host provenance.**

## Performance

- **Duration:** 10 min
- **Tasks:** 3/3
- **Files modified:** 2

## Accomplishments

- Added a `phase21_lto_evidence`-tagged test that directly emits the checked Phase 14 multi-function fixture, then reuses the established four-tier runner and all-pairs comparator.
- Recorded fixture and emitted-C SHA-256 digests, Apple Clang path/version, Darwin arm64 host, common flags, and all comparison lanes in `21-LTO-EVIDENCE.md`.
- Verified the existing independent semantic mutation control still rejects seeded disagreement.
- Kept the comparison opt-in; untagged recurring CI continues to exercise the cheap structural guards.

## Task Commits

1. **Task 1: Create a tagged test for the emitted multi-function LTO comparison** — `f5bfded` (test)
2. **Task 2: Run the scoped comparison and record its exact evidence boundary** — `a2d018c` (docs)
3. **Task 3: Recheck the existing comparator's independent negative control** — verified with the Plan 02 codebase; no new source change required.

## Verification

- `GOCACHE=/private/tmp/ai-lang-phase21-gocache go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v` — passed. Observed `darwin/arm64`, `/usr/bin/clang`, Apple Clang 21.0.0, fixture digest `a86d9e90be4866cb20626961fc3757604bea7ac68ffff27f4eab64cc51f3e556`, emitted-C digest `d096fca69195eb63b09566387690a7b40fdc9c364f55b0b4a24209a895ff6416`, and all-pairs semantic equality across interpreter, `-O0`, `-O3`, and `-O3 -flto`.
- `GOCACHE=/private/tmp/ai-lang-phase21-gocache go test ./internal/compiler/session -run '^TestPhase16EmitterPortSemanticGuardIsNotInert$' -count=1 -v` — passed.
- `git diff --check` — passed.

## Deviations from Plan

- `phase16CompareDirectProgramFourTiers` did not support this fixture's `Switch` input type. The new test reuses its underlying checked-fixture, direct-emitter, four-tier runner, schema projection, and comparator helpers with the fixture's explicit `Off` input, without duplicating comparator logic.
- The fixture source has a stale comment claiming the multi-function match shape is refused. The existing `TestLTOInertnessOnMultiFunctionEmission` and this direct-emitter run both prove it currently emits; no emitter behavior was changed.

**Total deviations:** 2, both limited to correcting the test harness path and documenting conflicting stale source prose.
**Impact on plan:** The exact required emitted fixture and existing comparator were used; no admission boundary changed.

## Issues Encountered

- The comparison completed successfully; no semantic disagreement or compiler failure occurred.

## Next Phase Readiness

- Plan 21-04 can reconcile the contract, emitter-retirement, compiler-measurement, CI, and debt claims before the phase-wide test run.

## Self-Check: PASSED

- The opt-in test, receipt, and recorded evidence agree; production code was unchanged.
- Both plan commits are present after baseline `50b6ec4`.

---
*Phase: 21-native-emission-ownership-and-resource-discharge-m004*
*Completed: 2026-09-26*

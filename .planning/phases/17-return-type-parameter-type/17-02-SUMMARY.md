---
phase: 17-return-type-parameter-type
plan: 02
subsystem: checker-and-reducer
tags: [ownership, type-facts, reduction]
requires:
  - phase: 17
    provides: directional checker type facts
provides:
  - Independent checker parameter Drops and return Fresh derivation
  - Reducer preservation of directional parameter and return facts
affects: [corevalidate, originvalidate, native-emitter]
actuals: {tokens: 0, tasks: 2, commits: 3}
tech-stack: {added: [], patterns: [directional-ability-derivation]}
key-files:
  created: []
  modified: [internal/compiler/check/check.go, internal/compiler/check/check_test.go, internal/compiler/reduce/reduce.go, internal/compiler/reduce/reduce_test.go]
key-decisions:
  - "Reduced foreign chains retain declared return facts while preserving legacy zero/one-fact layouts."
requirements-completed: [TYP-01, TYP-04]
coverage:
  - id: D1
    description: "Checker independently derives directional ownership abilities and kills a return-only fault."
    requirement: TYP-04
    verification:
      - {kind: unit, ref: "internal/compiler/check/check_test.go#TestPhase17CheckerDirectionalAbilities", status: pass}
    human_judgment: false
  - id: D2
    description: "Reducer preserves directional type facts independently of operation order."
    requirement: TYP-01
    verification:
      - {kind: unit, ref: "go test ./internal/compiler/reduce -count=1", status: pass}
    human_judgment: false
duration: 0 min
completed: 2026-09-22
status: complete
---

# Phase 17 Plan 02: Directional Ownership Facts Summary

**Checker-local Drops/Fresh derivation and reducer projections now preserve parameter and return contracts independently.**

## Accomplishments

- Added a return-only checker mutation control and directional ability tests.
- Removed reducer parameter-TypeID inheritance and retained return facts through foreign-chain reduction.

## Task Commits

1. `91a9d5b` — checker directional ability lookup.
2. `65465b4` — reducer parameter fact preservation.
3. `5590190` — foreign-chain return-fact retention.

## Verification

- `go test ./internal/compiler/check ./internal/compiler/reduce -count=1` — passed.

## Deviations from Plan

**[Rule 1 - correctness]** The broad reducer suite exposed a foreign-chain projection that retained a legacy failure fact instead of the declared return fact. The scoped reducer fix and regression test landed in `5590190`.

## Issues Encountered

None remaining.

## Next Phase Readiness

Independent core and origin validators can now implement and compare their own directional derivations.

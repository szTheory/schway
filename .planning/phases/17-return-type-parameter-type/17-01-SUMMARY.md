---
phase: 17-return-type-parameter-type
plan: 01
subsystem: checker
tags: [type-facts, call-contracts, source-fixtures]
requires:
  - phase: 16
    provides: sole native-emitter baseline
provides:
  - Source-reachable directional call-contract diagnostics
  - Separate parameter and return type facts at checker admission
affects: [reducer, independent validators, native emitter, repair]
actuals:
  tokens: 0
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [directional parameter-and-return type facts]
key-files:
  created:
    - testdata/phase17/return_type_tracer.lang
    - testdata/phase17/call_argument_type_mismatch.lang
    - testdata/phase17/call_return_type_unrepresentable.lang
    - internal/compiler/session/session_phase17_test.go
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
key-decisions:
  - "Call sources resolve against callee parameter facts; targets resolve against caller-visible callee return facts."
requirements-completed: [TYP-01, TYP-02, TYP-03]
coverage:
  - id: D1
    description: "Resource-to-Result source tracer passes checker admission."
    requirement: TYP-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase17_test.go#TestPhase17SourceFrontierMoved"
        status: pass
    human_judgment: false
  - id: D2
    description: "Argument and return call-contract failures are reached from source with stable causes."
    requirement: TYP-02
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/check ./internal/compiler/session -run TestPhase17(SourceFrontierMoved|TwoTypeAdmission|CallContractCauses)"
        status: pass
    human_judgment: false
duration: 0 min
completed: 2026-09-22
status: complete
---

# Phase 17 Plan 01: Source Frontiers and Directional Call Contracts Summary

**Parser-valid Resource-to-Result source fixtures now exercise checker admission and directional call contracts.**

## Accomplishments

- Added the fixture-first canonical tracer and two source-level diagnostic frontiers.
- Replaced the checker’s single-type admission with directional parameter and return facts.
- Moved source expectations to `check.call_argument_type_mismatch` and `check.call_return_type_unrepresentable` with ordered causes.

## Task Commits

1. Task 1 — `a6fb72e` `test(17-01): pin source frontier before widening`
2. Task 2 — `337ecc1` `feat(17-01): admit directional call contracts`

## Verification

- `go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase17(SourceFrontierMoved|TwoTypeAdmission|CallContractCauses)' -count=1 -v` — passed.
- `go test ./internal/compiler/check ./internal/compiler/session -count=1` — passed during task execution.

## Deviations from Plan

None - plan executed as specified.

## Issues Encountered

None.

## Next Phase Readiness

Directional checker facts are available for reducer and independent-peer work in Wave 2.

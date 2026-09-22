---
phase: 17-return-type-parameter-type
plan: 05
subsystem: compiler agreement and interpreter testing
tags: [go, type-facts, mutation-testing, interpreter]
requires:
  - phase: 17-02
    provides: checker directional type facts
  - phase: 17-03
    provides: corevalidate directional return derivation
  - phase: 17-04
    provides: originvalidate directional return derivation
provides:
  - Three-peer return-contract agreement gate with coordinated-fault oracle
  - Core directional call-fact assertions and deterministic interpreter tracer
affects: [phase-17-native-lowering, TYP-01, TYP-04]
actuals:
  tokens: 2886
  tasks: 2
  commits: 3
plan_head_before: defc33a737e99a7c65b8cdfc588b9d0bc64fb618
tech-stack:
  added: []
  patterns: [session-local declared-contract oracle, return-only peer mutation matrix]
key-files:
  created: []
  modified:
    - internal/compiler/session/session_phase17_test.go
    - internal/compiler/core/core_test.go
key-decisions:
  - "The coordinated-fault oracle reconstructs nominal Result freshness from source declaration and core function shape rather than consuming a peer contract."
  - "Call TypeIDs are caller-local directional facts; their shapes must agree with the callee parameter and return declarations."
requirements-completed: [TYP-01, TYP-04]
coverage:
  - id: D1
    description: "Checker, corevalidate, and originvalidate agree on directional return facts and every isolated or coordinated return fault is rejected."
    requirement: TYP-04
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/session -run TestPhase17... -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: "The Resource-to-Result tracer retains directional call facts and executes deterministically in the interpreter."
    requirement: TYP-01
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/core ./internal/compiler/session -run TestPhase17(TwoTypeCoreFacts|InterpreterTracer) -count=1 -v"
        status: pass
    human_judgment: false
duration: 0 min
completed: 2026-09-22
status: complete
---

# Phase 17 Plan 05: Three-Peer Return Agreement Summary

**Independent checker, core-validator, and origin-validator return facts now face a session-local oracle, while the Resource-to-Result tracer proves directional core facts and deterministic interpreter execution.**

## Accomplishments

- Added a four-case mutation matrix that activates each return-only peer fault and all three simultaneously, preserving serialized parameter facts in every case.
- Reconstructed the canonical nominal Result expectation in the session test from declared source/core structure, preventing coordinated faulty peers from self-certifying.
- Asserted caller-local source/target TypeIDs against the callee's Resource parameter and Result return shapes, then ran the multi-function tracer twice through `interp.Run` with byte-identical results.

## Task Commits

1. **Task 1: Build the three-peer return-ability mutation matrix** — `0113763` (`test`)
2. **Task 2: Run the canonical two-type tracer through checked core and interpreter** — `2d02f6f` (`test`)

## Verification

- `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/session -run 'TestPhase17(Checker(ReturnOnlyMutation|MutationControlBoundary)|CorePeer(ReturnOnlyMutation|IndependenceBoundary)|OriginPeer(ReturnOnlyMutation|IndependenceBoundary)|ThreePeerAgreement|EachReturnMutationFails|CoordinatedReturnMutationFails|ParameterFactsStayFixed)' -count=1 -v` — passed.
- `go test ./internal/compiler/core ./internal/compiler/session -run 'TestPhase17(TwoTypeCoreFacts|InterpreterTracer)' -count=1 -v` — passed.

## Decisions Made

- Kept the agreement oracle test-local and declaration-based, so no peer helper or peer-produced return contract participates in the coordinated-fault verdict.
- Treated TypeIDs as directional identities within the caller: the test compares their resolved shapes to the callee contract rather than incorrectly requiring function-local IDs to match across functions.

## Deviations from Plan

None - plan executed exactly as written.

## Next Phase Readiness

Native lowering can use this semantic and independent-peer evidence as the TYP-01/TYP-04 baseline.

## Self-Check: PASSED

- Found `internal/compiler/session/session_phase17_test.go` and `internal/compiler/core/core_test.go`.
- Found task commits `0113763` and `2d02f6f` in repository history.

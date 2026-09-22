---
phase: 15-event-identity-lang-execution-2
plan: "06"
status: complete
subsystem: session differential
tags: [session, comparator, execution-schema-2, executionpeer]
requires: [15-02, 15-03, 15-05]
provides: [schema-2-peer-gate, invocation-field-routing, bidirectional-peer-controls]
affects: [four-tier-differential, observability]
coverage:
  - id: D1
    description: "Schema 2 engine comparison fails closed through the independent peer, including producer and peer-boundary fault controls."
    requirement: OBS-03
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestSchema2ComparisonRequiresPeerVerdict"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase15_frontier_test.go#TestExecutionProducerFaultIsCaughtByPeer"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase15_frontier_test.go#TestExecutionPeerAcceptanceFaultIsCaughtByControl"
        status: pass
    human_judgment: false
  - id: D2
    description: "Program-aware comparison preserves Schema 0 and Schema 1 behavior without consulting the Schema 2 peer."
    requirement: OBS-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestPhase5CompareProgramEnginesPreservesLegacySchemas"
        status: pass
    human_judgment: false
tech-stack:
  added: []
  patterns: [program-aware-schema-2-peer-gate, disjoint-fault-seams]
key-files:
  modified:
    - internal/compiler/session/session_phase5_compare.go
    - internal/compiler/session/session_phase5_compare_test.go
    - internal/compiler/session/session_phase11_differential_test.go
    - internal/compiler/session/session_phase15_frontier_test.go
decisions:
  - "Keep legacy comparison unchanged and require independent peer validation only at the program-aware /2 four-tier gate."
  - "Test peer acceptance corruption at the session boundary, separately from producer-document corruption."
actuals:
  tokens: 2613
  tasks: 2
  commits: 2
plan_head_before: 0a1e7de
completed: 2026-09-19
---

# Phase 15 Plan 06: Session Peer Gate Summary

**Schema-2 four-tier comparison now validates every engine document with the independent causal peer before it can report agreement.**

## Accomplishments

- Routed `Execution.Events.Invocation` and `Execution.Events.CalleeFunctionID` through the fail-closed field-routing table; whole-event equality reports either mutation as `axis:event-order`.
- Added `Phase5CompareProgramEngines`, which validates each `/2` document before pair comparison while preserving the legacy comparator behavior for `/0` and `/1`.
- Routed the Phase 11 program-aware four-tier helper through that peer gate.
- Added opposite-direction controls: producer-corrupted evidence is refused by the intact peer, and a peer-boundary acceptance fault visibly blesses fixed bad bytes only until restoration rejects those exact bytes.

## Verification

Passed:

- `go test ./internal/compiler/session -run 'TestComparisonFieldRoutingIsExhaustive|TestInvocationFieldsAreCompared|TestSchema2ComparisonRequiresPeerVerdict' -count=1 -v`
- `go test ./internal/compiler/session -run 'TestExecutionProducerFaultIsCaughtByPeer|TestExecutionPeerAcceptanceFaultIsCaughtByControl' -count=1 -v`
- `go test ./internal/compiler/session/... -count=1` is running concurrently under the phase orchestrator at handoff time.

## Task Commits

1. `a6ecb26` — `feat(15-06): gate schema 2 comparison with peer`
2. `851ea3f` — `test(15-06): prove peer and producer independence`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Required integration] Added the program-aware Phase 11 helper to the changed files.**

- **Found during:** Task 1
- **Issue:** The plan's declared file list omitted `session_phase11_differential_test.go`, the only session call site holding both the checked `core.Program` and all four engine documents.
- **Fix:** Routed that helper through `Phase5CompareProgramEngines`, preserving the generic legacy comparator.
- **Files modified:** `internal/compiler/session/session_phase11_differential_test.go`
- **Commit:** `a6ecb26`

## Self-Check: PASSED

- Confirmed all four modified session files and this summary exist.
- Confirmed commits `a6ecb26` and `851ea3f` exist in Git history.

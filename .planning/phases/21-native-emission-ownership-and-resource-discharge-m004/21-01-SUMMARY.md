---
phase: 21-native-emission-ownership-and-resource-discharge-m004
plan: 01
subsystem: compiler-contracts
tags: [resource-discharge, foreign-ownership, structural-validation]
requires:
  - phase: 16
    provides: Whole-program emission refusal boundary and family-specific reopening conditions
provides:
  - Versioned machine-readable foreign-exit and discharge contract
  - Recurring mutation-sensitive structural validator
affects: [phase-21, native-emission, verification]
actuals:
  tokens: 2709
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [Versioned JSON contract with independent fail-closed validator]
key-files:
  created:
    - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json
    - internal/compiler/session/session_phase21_contract_test.go
  modified: []
key-decisions:
  - "Contract classification remains design-only; every cut emitter family stays unadmitted."
  - "Normal and error return carry discharge obligations and future runtime evidence; opaque exits are refused or outside cleanup guarantees."
patterns-established:
  - "Contract completeness and refusal invariants are validated from checked-in data and seeded in-memory mutations."
requirements-completed: []
coverage:
  - id: D1
    description: "Seven modeled foreign exits and all three cut emitter families have a machine-checked design-only contract."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract' -count=1 -v"
        status: pass
    human_judgment: false
duration: 4min
completed: 2026-09-26
status: complete
plan_head_before: 0ba6504bc71edaea9f9dc9bab11237b242bb72e4
commits: 2
---

# Phase 21 Plan 01: Resource Discharge Contract Summary

**A checked JSON contract now classifies all modeled foreign exits, pins per-family reopening requirements, and rejects mutations that weaken the refusal boundary.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-09-26T01:50:29Z
- **Completed:** 2026-09-26T01:53:49Z
- **Tasks:** 2/2
- **Files modified:** 2

## Accomplishments

- Added a versioned contract for normal return, error return, supported unwind, nonlocal transfer, defect, cancellation, and process termination.
- Kept D-16-11, D-16-12, and D-16-13 refused, preserving each family's distinct prerequisite, host evidence, and refusal fence.
- Added recurring structural validation with mutation controls for missing, unknown, duplicate, and misclassified exits; missing discharge evidence; and attempted family admission.

## Task Commits

1. **Task 1: Load and validate the resource-discharge contract** — `bd667b9` (feat)
2. **Task 2: Prove contract completeness and fail-closed behavior with seeded mutations** — `cf33387` (test)

## Verification

- `go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract$' -count=1 -v` — passed.
- `go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract' -count=1 -v` — passed, including all seven mutation subtests.

## Deviations from Plan

None.

## Self-Check: PASSED

- Both declared artifacts exist and are committed.
- The two task commits are present after the recorded plan baseline.
- The contract validator and all mutation controls pass.

---
phase: 16-branch-match-emitter-port
plan: "08"
subsystem: compiler-native-emission
tags: [decision, cutover, provenance]
requires: [16-07]
provides: [approved-one-way-dispatch-cutover]
affects: [16-09]
tech-stack:
  added: []
  patterns: [explicit-blocking-human-authorization, pre-cut-evidence-gate]
key-files:
  created: [.planning/phases/16-branch-match-emitter-port/16-08-SUMMARY.md]
  modified: []
decisions:
  - "approve-cutover authorizes only the Plan 16-09 atomic dispatch cut."
metrics:
  duration: "0m"
  completed: 2026-09-19
  tasks: 1
  commits: 1
status: complete
---

# Phase 16 Plan 08: Approved Atomic Cutover Summary

The required one-way cutover authorization was explicitly supplied as `approve-cutover` after all three pre-deletion evidence gates passed.

## Decision Recorded

**Decision:** `approve-cutover`

This authorizes only Plan 16-09 to flip both public dispatchers to `emitProgram`, delete exactly `emitMatch`, `emitBranch`, and `emitLinear` in the same atomic commit, and transition the affected goldens, digest map, and four-entry provenance ledger coherently. It does not authorize deleting foreign or pointer-specialized helpers, admitting cut-M004 exclusions, or treating pre-cut `/1` and direct `/2` bytes as equivalent.

## Pre-deletion Evidence

All required evidence was green before approval:

- `bd80334` — five-row pre-cut convergence characterization, preserving the measured legacy/direct divergence.
- `54e8df3` — four-entry golden provenance ledger and fault controls.
- `258aada` — four-tier direct program semantic guard, including the repaired bare-match event evidence.
- `b98f283` and `065a202` — NAT-09/M004 cut ownership and amendment controls.

The combined focused gate passed:

```text
go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/session -run 'TestN1ConvergenceDifferential|TestPhase16GoldenChangeLedger|TestPhase16EmitterPortSemanticGuardIsNotInert|TestPhase16EmitterCutsAreAmendedAndOwned' -count=1 -v
```

## Deviations from Plan

None - this decision record contains no source changes and preserves the plan's exact scope boundary.

## Self-Check: PASSED

- Approval and all three named evidence gates are recorded above.
- The authorization is constrained to the Plan 16-09 atomic cutover.

---
phase: 29-byte-views-and-the-adler-32-library
plan: 01
subsystem: compiler
tags: [byte-view, cgen, native-proxy, wave-0-gate, adler32]
requires: [29-00]
provides:
  - OpBytesLen / OpBytesAt arms in both C shells with a never-elided bounds if
  - SCHWAY_BYTE_VIEW typedef emitted only when a ByteView function is present
  - gate-only native proxy tests (literals and OOB at -O0, -O3, -O3 -flto)
  - TestPhase29ByteViewAbilityAgreement (D-17 / D-23)
  - 29-00-GATE-VERDICT.md with verdict PROCEED
affects: [29-02, 29-03]
tech-stack:
  added: []
  patterns: [gate-only disposable C driver, armed mismatch control, env-gated cost reading]
key-files:
  created:
    - internal/compiler/cgen/cgen_phase29_gate_test.go
    - internal/compiler/corevalidate/corevalidate_phase29_ability_internal_test.go
    - .planning/phases/29-byte-views-and-the-adler-32-library/29-00-GATE-VERDICT.md
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/session/session_phase29_gate_test.go
key-decisions:
  - "Gate verdict is PROCEED: slowest lane max at 65,536 is 225.78 ms (interpreter, -race) and 3.89 ms (native), far under the 4x D-13 STOP thresholds (48 s / 12 s)"
  - "The bounds check is an ordinary if before the load in both shells; OpBytesAt reserves only the terminal defect event"
requirements-completed: [BYTE-01, BYTE-02, APP-10]
status: complete
commits: 3
plan_head_before: f2a1a47fbbf1aba53ff4b2f7f4d14c8cb1bc52e9
plan_head_after: 1c6a7e365963640403bfec1aecaf32b24987a091
actuals:
  tokens: 35000
  tasks: 2
  commits: 3
duration: 25min
completed: 2026-10-11
---

# Phase 29 Plan 01: Native byte views and the gate verdict Summary

Both C shells now lower byte_len / byte_at over a SCHWAY_BYTE_VIEW parameter with a never-elided bounds `if`; the gate-only native proxy returns all seven pinned adler32 literals and the exit-65 OOB contract at -O0, -O3 and -O3 -flto, and the written gate verdict is PROCEED.

## Gate verdict: PROCEED

See 29-00-GATE-VERDICT.md. Flags (a), (b), (c), (e), (f) all pass; no lane is near the 4x budget.

## Tasks

1. cad95b7: cgen arms (schema2ExecutionDocumentSize, hasScalarOperations, emitProgramScalarFunction), SCHWAY_BYTE_VIEW typedef, linearInput / linearFixedNames, proxy vector and cost tests. `EmitApplication` accepted the two-function unpruned program directly (the ByteView function is an unreferenced callee of the U64 entry), so no unexported emitter was needed.
2. bf718b0: TestPhase29ByteViewAbilityAgreement with armed flip control. 1c6a7e3: TestPhase29GateCostReading and the verdict.

## Verification

- TestPhase29GateNativeProxyVectors, TestEmittedCExitLiteralsMatchTable, TestGeneratedIdentifierNamespacesStayConfined, TestPhase273CapacityContractLiteralsAgree pass.
- TestPhase16PublicEmitterConsumerInventory, TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity, TestAllOperationKindsHandledAtEverySite pass (no emitter row moved, untouched programs emit identical C).
- Cost readings (SCHWAY_SPIKE008=1) taken on interpreter plain and -race (no data race) and the three native lanes; numbers are in the verdict.
- Session meta-tests (Inventory, Groundedness, Reconciliation, ValidationLifecycle) pass.

## Deviations from Plan

None in scope. Note: the proxy vector test passed on first run once the cgen arms existed; the RED step was not committed separately.

## Known Stubs

None.

## Threat Flags

None. The proxy driver is test-only (T-29-04 mitigated by the bounds if; T-29-05 by the unsigned widen, asserted by the [0x80,0xFF] vector).

## Self-Check: PASSED

Created files exist; commits cad95b7, bf718b0, 1c6a7e3 are on the branch.

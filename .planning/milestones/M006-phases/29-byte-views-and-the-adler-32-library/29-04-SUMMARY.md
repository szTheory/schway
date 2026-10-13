---
phase: 29-byte-views-and-the-adler-32-library
plan: 04
subsystem: evidence
tags: [byte-views, oob-defect, execution-peer, schema-2, cgen]
requires: [29-01]
provides:
  - "ProjectExecutionSchema2 reason for OpBytesAt defects"
  - "executionpeer checkedBytesAt arm and function.defected classifier entry"
  - "interpreter boundary table and C bounds-before-load ordering tests"
affects: [29-10]
tech-stack:
  added: []
  patterns: ["independent peer arm beside checkedOverflow / checkedRemainderZero"]
key-files:
  created:
    - internal/compiler/session/session_phase29_evidence_test.go
    - internal/compiler/executionpeer/executionpeer_phase29_test.go
    - internal/compiler/interp/interp_phase29_oob_test.go
    - internal/compiler/cgen/cgen_phase29_oob_test.go
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/executionpeer/executionpeer.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "Peer arm checks fixed reason, source place, no target place, final position, empty defect outcome; independent of the projection"
requirements-completed: [BYTE-02]
status: complete
duration: 20 min
completed: 2026-10-11
commits: 2
plan_head_before: 9cd4809cb3139f90b73a776577bb58d1385be34c
plan_head_after: cef27156bfff220ded884ef044d82677f0b30fb2
actuals:
  tokens: 9000
  tasks: 2
  commits: 2
---

# Phase 29 Plan 04: Byte index defect evidence consumers Summary

An out-of-range `byte_at` is now a classified defect everywhere: interpreter evidence projects with the reason "byte index out of bounds" and the execution peer accepts it through an independent arm while refusing four forgeries; both C shells are pinned to test the bound before the load.

## Accomplishments

- `session.ProjectExecutionSchema2`: `case core.OpBytesAt` maps the defect event to the fixed reason.
- `executionpeer`: `checkedBytesAt` arm (reason, source place, no target place, final event, empty defect outcome), a follows-terminal refusal, and `OpBytesAt` in the function.defected classifier.
- `TestPhase29OOBEvidenceProjection`, `TestPhase29ExecutionPeerByteIndexDefect` (wrong reason, non-final defect, foreign place, non-empty value all refused), `TestPhase29InterpreterOOBBoundary` (index==len, len-1, 0 on empty, max U64 no wrap), `TestPhase29BoundsCheckPrecedesLoad` (textual order for every OpBytesAt in both shells, exact fputs/exit(65) and exit(0) terminals, no assert, const data typedef).

## Deviations from Plan

**1. [Rule 3 - Blocking] Application shell needs a U64 entry** - Found during Task 2: `emitProgramWithShell` refuses a ByteView entry in the application shell, so the cgen test source has a U64 `main` entry with the ByteView function emitted as a callee (same shape as the 29-01 gate test). Files: cgen_phase29_oob_test.go.

**2. [Rule 3] Emitter inventory row shifted** - `session.go:EmitNative:3042` became `3044` in testdata/phase16/public-emitter-consumers.json (amended into the Task 1 commit).

**3.** The bounds check and interpreter arm already existed from 29-00/29-01, as noted; this plan added evidence consumers and tests only. RED-first was not separately committed: the projection test and peer arm were written together (tests asserted against the new behaviour on first run).

**Total deviations:** 3 (2 blocking-fix, 1 process note). **Impact:** none on scope.

## Known Stubs

None.

## Self-Check: PASSED

Focused verify sets pass (session, executionpeer, interp, cgen); Inventory/Groundedness/Reconciliation/ValidationLifecycle tests pass in core and session; created files present; commits 694868f and cef2715 on the branch.

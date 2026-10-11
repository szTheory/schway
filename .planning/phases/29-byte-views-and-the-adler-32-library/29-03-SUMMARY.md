---
phase: 29-byte-views-and-the-adler-32-library
plan: 03
subsystem: compiler
tags: [byte-view, peers, seeded-faults, neighbor-refusals, interpreter-events]
requires: [29-00, 29-01]
provides:
  - independent bytes_len / bytes_at structural laws in corevalidate and originvalidate
  - default-off law toggles with export_test.go setters (byte-op, back-edge authority, return value, scalar analysis)
  - TestPhase29NeighborSeededFaults in corevalidate, originvalidate and pathoracle (rows n1..n4)
  - TestPhase29NeighborRefusals, TestPhase29PeerIndependence, TestPhase29NoLoopEventsStructure, TestPhase29NoLoopEventsInterpreter
affects: [29-09]
tech-stack:
  added: []
  patterns: [JSON-clone core forgery, default-off law toggle per peer, noLaw reason pinned as acceptance]
key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_phase29_test.go
    - internal/compiler/originvalidate/originvalidate_phase29_test.go
    - internal/compiler/pathoracle/pathoracle_phase29_test.go
    - internal/compiler/check/check_phase29_peers_test.go
    - internal/compiler/interp/interp_phase29_events_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/export_test.go
    - internal/compiler/check/scalar.go
key-decisions:
  - "No new refusal code: every asserted code already existed; originvalidate reuses core.scalar_operation_invalid for the byte-op law"
  - "pathoracle has no matching law for n1..n4 (it refuses every cyclic CFG identically); rows pinned as noLaw"
requirements-completed: [BYTE-01, BYTE-03]
status: complete
commits: 3
plan_head_before: 8b3b92b77c47252da4f1361cc0b890c4e2c674cd
plan_head_after: 066da6a99cce10002c9ff2feda47b7c615b1d6f1
actuals:
  tokens: 40000
  tasks: 3
  commits: 3
duration: 60min
completed: 2026-10-11
---

# Phase 29 Plan 03: Independent byte-op laws, loop neighbors, and event-free loop Summary

corevalidate and originvalidate now each refuse forged bytes_len / bytes_at shapes with their own law (armed by seeded-fault toggles), every loop-function neighbor keeps its existing code, and adler32's loop is proven to record no per-byte events against an armed let-in-loop mutant.

## Tasks

1. Tracer (7be0c05): corevalidate tightened its `case core.OpBytesLen, core.OpBytesAt` law (target is not the parameter place or the source) behind `disableByteOpLawForTest`; originvalidate gained `phase29ByteOpProblem` (sibling of `phase27ArithmeticProblem`, runs before the scalar fixed point). Eight forged rows per validator. pathoracle: acyclic ByteView body yields zero endpoints, forged unknown source yields `pathoracle.unknown_place`. This commit also carried the neighbor seeded-fault toggles and `TestPhase29NeighborSeededFaults` for all three peers.
2. Neighbors (0ee1c18): `TestPhase29NeighborRefusals` pins nine source-level neighbors plus a forged loan (category `loan`) and a forged ByteView copy across the back edge (category `loan_derived_provenance`); `scalarAuthorityCategory` learned `"ByteView"`.
3. Events (066da6a): structural zero-OpCopy walk of the loop plus armed `let step` mutant; interpreter event counts equal for 4,096 x 0x00 and 5,552 x 0xFF, mutant events grow; empty view returns 1 and ascending order pinned.

## Observed peer matrix (D-22), forged adler32 core

| Row | corevalidate (law, off result) | originvalidate (law, off result) | pathoracle |
|-----|-------------------------------|----------------------------------|-----------|
| n1 OpCopy of ByteView param | core.scalar_backedge_authority; off -> core.ability.copy_denied | core.scalar_analysis_invalid; off -> accepted | noLaw (cyclic refusal) |
| n2 OpScalarStore of ByteView | core.scalar_backedge_authority; off -> accepted | core.scalar_analysis_invalid; off -> accepted | noLaw |
| n3 return carrying ByteView | core.return_type_mismatch (block replay); off -> accepted | noLaw (accepts; provenance only) | noLaw |
| n4 ByteView place across back edge | core.scalar_backedge_authority; off -> core.ability.copy_denied | core.scalar_analysis_invalid; off -> accepted | noLaw |

No row is accepted by all three peers, and no new code was needed (D-22 untouched). No refusal row (R2) was accepted, and no move row (M1) was involved (D-24 not triggered).

## Deviations from Plan

**1. [Rule 1 - plan assumption] check's fixed point is not a type validator for byte ops.** The plan asked that check's fixed point refuse a type-forged bytes_at. It only analyses initialization; a forged core is never check's input. `TestPhase29PeerIndependence` pins this gap (fixed point returns "" on the forged core) and pins the source-level counterpart (`byte_at` over a U64 operand is refused `type.*` first). Adding a new check refusal code was out of scope.

**2. [Rule 1 - plan assumption] Disabling the corevalidate byte-op law does not always change the outcome.** For two rows (source is a U64 place, non-U64 RightID) a block-replay arm re-derives the same `core.scalar_operation_invalid`; for six rows the toggle changes the code (core.place_uninitialized or core.scalar_field_kind_exclusive). Armed rows assert the change; the two redundant rows are pinned as unchanged. `core.scalar_field_kind_exclusive` is reachable only with the structural law disabled (the structural law refuses first).

**3. [Rule 1] n3's refusing law is the block-replay returned-type check, not the match-arm return law** the plan's candidate list suggested; the toggle sits on that site.

**4. pathoracle has no matching law for any neighbor row.** It refuses every cyclic CFG identically (even the unforged loop), so rows carry a `noLaw` reason and a pinned unchanged outcome.

**5. Commit boundary.** The neighbor-row toggles and per-peer `TestPhase29NeighborSeededFaults` landed in the Task 1 commit (same files as the byte-op laws); Task 2's commit holds the check-side half.

**6. The mutant tail is shortened** (two doublings instead of sixteen) since only the loop body matters for the control.

## Deferred / notes

- Pre-existing scalar `var` formatter defect from 29-02 was not touched; no test here round-trips var bodies through the formatter.
- Not run: the full Go suite (hosted CI only). Run: the three validator packages unfiltered, check back-edge/fixed-point tests, session `TestPhase29NeighborBaseline`, and the core/session Inventory|Groundedness|Reconciliation|ValidationLifecycle meta-tests (all pass).

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

All five new test files exist; commits 7be0c05, 0ee1c18, 066da6a are ancestors of HEAD.

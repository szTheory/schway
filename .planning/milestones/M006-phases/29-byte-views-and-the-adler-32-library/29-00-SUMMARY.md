---
phase: 29-byte-views-and-the-adler-32-library
plan: 00
subsystem: compiler
tags: [byte-view, adler32, tracer, wave-0-gate, interpreter]
requires: []
provides:
  - core OpBytesLen / OpBytesAt (AllOperationKinds now 24)
  - byte_len(bytes) / byte_at(bytes, i) scalar atoms and the ByteView parameter type ({drop} only)
  - corevalidate, originvalidate and interpreter arms; adler32 published Callable
  - interp.RunByteViewForTest harness entry
  - testdata/phase29/adler32_library.schway registered in all corpus registers
affects: [29-01, 29-02, 29-03]
tech-stack:
  added: []
  patterns: [independently written peer arms, harness-only ForTest entry, pins committed before production edits]
key-files:
  created:
    - internal/compiler/session/session_phase29_gate_test.go
    - testdata/phase29/adler32_library.schway
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/check/scalar.go
    - internal/compiler/ability/ability.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/interp/interp.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase272_identity_test.go
    - internal/compiler/session/session_phase273_identity_test.go
    - testdata/phase16/public-emitter-consumers.json
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "ByteView type fact sits at :type:0 only for ByteView-parameter functions; U64-parameter functions keep their type-fact order so frozen goldens stay byte-identical"
  - "Both byte ops reuse SourceID/RightID/TargetID; no new LinearOperation field, no lexer change"
  - "Bounds comparison precedes every payload slice index in the interpreter; octet widens unsigned"
requirements-completed: [BYTE-01, BYTE-02, APP-10]
status: complete
commits: 4
plan_head_before: 058b1210c8490b730fb240721b1a6e3f20806bab
plan_head_after: 21dbd7d9f7c177d72f1f7d36d6a90396e872af51
actuals:
  tokens: 60000
  tasks: 2
  commits: 4
duration: 35min
completed: 2026-10-11
---

# Phase 29 Plan 00: Byte views tracer (Wave-0 gate, tracer half) Summary

adler32 written in ordinary Schway (`byte_len` / `byte_at` over a `ByteView` parameter) checks clean, is published Callable by originvalidate, and returns every pinned oracle literal on the interpreter; out-of-range reads are a defect.

## Tasks

1. Tracer. Pins committed first (94894b2, touches only the new gate test file), then the byte-view plumbing across core, parser, check, ability, corevalidate, originvalidate and interp plus the fixture and its registers (62065f8).
2. Interpreter Stage-A vector table, byte edges and the OOB defect (f7eb5c4). No interpreter fix was needed; every row passed on first run.
3. Follow-up (21dbd7d): the phase 16 emitter inventory pins `core_test.go` call lines, and registering the fixture shifted one by a line; re-pinned 1022 to 1023.

## Verification

- TestPhase29NeighborBaseline (4 rows), TestPhase29GateCallable, TestPhase29GateInterpreterVectors (12 rows plus OOB on empty and "a"): pass.
- TestPhase28EntryExitClassManifest, TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity, TestLanguageMaturityCountsAreCurrent (now 157 programs, 5,069 lines): pass.
- core: TestAllOperationKindsRegistered (24), TestAllOperationKindsHandledAtEverySite, TestPhase26CoreInventory pass. check back-edge/fixed-point tests, and the ability, corevalidate, originvalidate, interp and syntax packages, pass.
- Session meta-tests (Inventory, Groundedness, Reconciliation, ValidationLifecycle) pass after the phase16 re-pin.
- `git diff 058b121 -- examples/phase23 testdata/phase272 testdata/phase273 internal/compiler/cgen` is empty; `RunByteViewForTest` has no use under `cmd/`.

## Gate reading (tracer half)

Flag (a) holds: adler32 is Callable with a single ByteView parameter place admitted by both peers. Flag (e), interpreter half, holds: no event is attributed to either byte op on success. The cgen, native proxy, Stage-A cost reading and written verdict remain plan 29-01; no GATE-VERDICT file was written here.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Phase 16 emitter inventory line pin**
- **Found during:** session meta-test run before finishing
- **Issue:** adding the fixture to `exhaustiveDispatchFixtures` shifted the `cgen.Emit` call in core_test.go from line 1022 to 1023, failing TestPhase16PublicEmitterConsumerInventory.
- **Fix:** updated the registry entry line number.
- **Files modified:** testdata/phase16/public-emitter-consumers.json
- **Commit:** 21dbd7d

**Total deviations:** 1 auto-fixed (1 blocking). **Impact:** none on scope.

Acceptance note: `grep -c 'OpBytesLen\|OpBytesAt' core.go` counts lines and prints 3 (two constants plus the registry line), against the plan's "at least 4"; each name appears twice by occurrence, so the intent is met.

## Known Stubs

None.

## Threat Flags

None. RunByteViewForTest is harness-only (T-29-02) and the bounds comparison precedes the slice index (T-29-03).

## Self-Check: PASSED

Created files exist; commits 94894b2, 62065f8, f7eb5c4, 21dbd7d are on the branch.

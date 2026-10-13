---
phase: 29-byte-views-and-the-adler-32-library
plan: 06
subsystem: checker-and-validators
tags: [fixture-owner, byte-views, closed-shape, abilities, corpus-registers]
requires: [29-01, 29-02, 29-03, 29-05]
provides: [fixture-owner-caller-shape, fixture-bytes-abilities, adler32-fixture-program]
affects: [29-07, 29-08, 29-09, 29-10]
tech-stack:
  added: []
  patterns: [closed-shape matcher per peer, callee-contract-resolved OpCall, compiler-minted release]
key-files:
  created:
    - examples/phase29/adler32_fixture.schway
    - internal/compiler/check/check_phase29_fixture_test.go
    - internal/compiler/corevalidate/corevalidate_phase29_fixture_test.go
    - internal/compiler/corevalidate/corevalidate_phase29_fixture_internal_test.go
    - internal/compiler/originvalidate/originvalidate_phase29_fixture_test.go
    - internal/compiler/session/session_phase29_fixture_check_test.go
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/ability/ability.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/corevalidate/corevalidate_phase29_ability_internal_test.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase272_identity_test.go
    - internal/compiler/session/session_phase273_identity_test.go
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "The fixture shape is keyed by schway_fixture_bytes_acquire (check) and by the schway_fixture_bytes_ prefix or FixtureBytes type name (peers); no Phase 23-25 matcher is edited"
  - "Callee is resolved through resolveCallBinding (argument/return gates) and accepted when it takes one ByteView and returns U64; the name is not pinned"
  - "corevalidate exempts only the ByteView-to-U64 fixture OpCall from the generic call-keeps-argument-type law, since the fixture entry is U64 to U64"
requirements-completed: [APP-10, BYTE-03]
status: complete
actuals:
  tokens: 36000
  tasks: 3
  commits: 3
plan_head_before: db6915602c26fa3de7a69e9cab1eb6753c378b4f
plan_head_after: 5deb01d6025cf82ef6558a7925b922fe5342b27d
completed: 2026-10-11
---

# Phase 29 Plan 06: Fixture-owner caller shape Summary

The real fixture program (examples/phase29/adler32_fixture.schway) now checks clean through three independent closed-shape implementations: check hand-builds a five-operation main (acquire, view, contract-resolved OpCall carrying the ByteView, compiler-minted release, return), and corevalidate and originvalidate each replay and refuse forged variants on their own.

## Tasks

1. Tracer (0eb2dc2): fixture program, check dispatch branch `checkFixtureBytesEntry`, `case "FixtureBytes"` in ability.Derive and corevalidate.deriveAbility, corevalidate fixture foreign-operation pins plus `fixtureOwnerLifecycleValid`, originvalidate `fixtureOwnerOriginProblem`, registers (phase28 manifest, phase272/273 allow-lists, LANGUAGE-MATURITY counts 158 programs / 5,145 lines).
2. Shape faults per peer (ed6d2e3): check tests (core shape, callee renamed `checksum`, five seeded source faults), corevalidate forged-core table, originvalidate forged-core table, FixtureBytes ability agreement with armed mismatch control.
3. Frozen-shape regression (5deb01d): Phase 23 file_byte keeps its golden core hash; a PathToken program naming the fixture symbols is refused with a check.fixture_owner_* code.

## Observed classes (phase28 manifest, dated 2026-10-11)

Check 0, EntryShape single-root, Entry main, no dead functions; Build exit 3 native.build_failed, InterpreterRun exit 3 tool.run_failed, NativeRun exit 3 native.tool_failure. Plan 29-07 may change Build (and the run classes) when native lowering of the FixtureBytes owner lands.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Generic call-type law refused the fixture OpCall**
- **Found during:** Task 1
- **Issue:** corevalidate's replay law "a call keeps its argument's type when parameter and return types match" refused main (U64 to U64) calling a ByteView to U64 function.
- **Fix:** added `fixtureViewCall`, a narrow exemption (function carries fixture-owner operations, call is ByteView to U64).
- **Files modified:** internal/compiler/corevalidate/corevalidate.go
- **Commit:** 0eb2dc2

**2. [Rule 2 - Missing critical] Direct test for the lifecycle matcher**
- **Found during:** Task 2
- **Issue:** generic replay laws (core.callee_frame_not_drained, core.invalid_target, core.final_claim_mismatch, core.place_uninitialized) refuse several forged shapes before `fixtureOwnerLifecycleValid` runs, so the table test alone would not prove the matcher.
- **Fix:** added corevalidate_phase29_fixture_internal_test.go (TestPhase29FixtureLifecycleShapeDirect), not listed in the plan's files. The external table pins the observed first corevalidate code per fault.
- **Commit:** ed6d2e3

**Total deviations:** 2 auto-fixed. **Impact:** none on scope; one extra test file.

## Verification

- Task verify commands pass: TestPhase29FixtureChecks, TestPhase29GateCallable, TestPhase28EntryExitClassManifest, TestPhase272/273 corpus identity, TestLanguageMaturityCountsAreCurrent; the five TestPhase29Fixture tests across check, corevalidate, originvalidate; both ability agreement tests.
- Session meta-tests (Inventory|Groundedness|Reconciliation|ValidationLifecycle) pass across ./internal/...
- Acceptance greps pass; examples/phase23, testdata/phase272 and testdata/phase273 unchanged since the plan base.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

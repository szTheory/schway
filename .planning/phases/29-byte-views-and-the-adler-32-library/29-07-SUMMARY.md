---
phase: 29-byte-views-and-the-adler-32-library
plan: 07
subsystem: native-application-lowering
tags: [fixture-owner, byte-views, pathoracle, cgen, application-shell, runnable-gain]
requires: [29-04, 29-05, 29-06]
provides: [fixture-owner-path-replay, fixture-owner-admission, fixture-main-emission, native-fixture-runnable-gain]
affects: [29-08, 29-09, 29-10, 29-12]
tech-stack:
  added: []
  patterns: [independent closed replay, whole-contract pin at the serializer boundary, application-shell-only lowering]
key-files:
  created:
    - internal/compiler/pathoracle/pathoracle_phase29_fixture_test.go
    - internal/compiler/cgen/cgen_phase29_fixture_test.go
    - internal/compiler/session/session_phase29_native_fixture_test.go
    - cmd/schway/main_phase29_test.go
  modified:
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/session/session.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "The fixture main is lowered only by the application shell; the conformance shell refuses it with cgen.fixture_owner_shell_unsupported"
  - "The fixture owner is recognized by the schway_fixture_bytes_ symbol prefix or the FixtureBytes type name so forged variants are refused by the fixture validators, never admitted through Phase 23-25 shapes"
  - "Failed acquire exits 65 silently; an out-of-bounds read inside the callee exits 65 without release (D-20)"
requirements-completed: [APP-10, BYTE-02]
status: complete
actuals:
  tokens: 40000
  tasks: 3
  commits: 3
plan_head_before: d14d384b67bc32ead1357ff837e8efe05b2ef544
plan_head_after: 67e2a9960455e13d6bffa5dd578811153de398cd
completed: 2026-10-11
---

# Phase 29 Plan 07: Native fixture runnable gain Summary

`schway build examples/phase29/adler32_fixture.schway --manifest examples/phase29/fixture_bytes.bindings.json --output ART` now exits 0, and one artifact answers `schway app run ART -- N` with the pinned Adler-32 checksum for every case id 0..9 (exit 0, empty stderr); case 10 exits 65 with empty stdout.

## Tasks

1. Tracer (77416b6): `pathoracle.ValidateFixtureOwnerPaths` (independent acquire/borrow/call/release/return replay, codes `pathoracle.fixture_owner_*`), `fixtureOwnerEntry` branch in `session.BuildApplication` (replay, exactly one manifest, `validateLocalOperationBindings`), cgen `checkFixtureOwnerProgram` / `validateFixtureOwnerFunction` / `emitProgramFixtureBytesFunction` plus the widened adapter include gate. Tracer verified end-to-end: all ten vectors from one artifact, then expanded.
2. Faults and pins (e71c3ae): eight forged-core cases for the replay, emission order and D-20 pins, admission refusals (no manifest, missing symbol, extra symbol).
3. Public CLI (67e2a99): `TestPhase29FixtureAppCLI` covers build, ten case ids, unknown case, and the unchanged `schway run` classes.

## Observed classes

The phase28 manifest rows for the fixture are unchanged (Build exit 3 native.build_failed, InterpreterRun exit 3 tool.run_failed, NativeRun exit 3 native.tool_failure): that walker invokes build and run without a binding manifest, so Build still stops at the manifest requirement. The runnable gain is the manifest-bearing build route, pinned by `TestPhase29FixtureAppCLI`; the manifest comment was updated with that dated explanation. `schway run` and `cmd/schway/main.go` are untouched (D-16).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Emitter inventory row shifted**
- **Found during:** Task 1
- **Issue:** session.go gained five lines; `TestPhase16PublicEmitterConsumerInventory` flagged the EmitNative row (3044 to 3049).
- **Fix:** re-pinned the one shifted row.
- **Files modified:** testdata/phase16/public-emitter-consumers.json
- **Commit:** 77416b6

**2. [Plan wording] Emitter placement**
- The fixture validator and emitter live in cgen_program.go with the other application-shell emitters; cgen.go needed no edit.

**Total deviations:** 1 auto-fixed. **Impact:** none on scope.

## Verification

- Task verify commands pass: TestPhase29NativeFixtureVectors, TestPhase16PublicEmitterConsumerInventory, TestPhase28EntryExitClassManifest, TestPhase272/273 corpus identity, the three cgen literal/namespace tests, the three Task 2 tests, TestPhase29FixtureAppCLI.
- Session meta-tests (Inventory|Groundedness|Reconciliation|ValidationLifecycle) pass across ./internal/... and ./cmd/...
- `git diff` against the plan base is empty for internal/compiler/native, examples/phase23, testdata/phase272, testdata/phase273 and cmd/schway/main.go.

## Known Stubs

None.

## Threat Flags

None. T-29-21 (manifest symbol equality), T-29-22 (release before use) and T-29-23 (unknown case) are mitigated and pinned by tests.

## Self-Check: PASSED

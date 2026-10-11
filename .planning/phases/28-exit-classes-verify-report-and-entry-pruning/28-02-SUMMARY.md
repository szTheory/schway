---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 02
subsystem: cli
tags: [exit-codes, exitcode-table, ast-scan, generated-c-pin]
requires:
  - phase: 28-01
    provides: exit-pin inventory and entry/exit-class manifest baseline
provides:
  - internal/compiler/exitcode leaf package (single named exit-code table, D-07)
  - appRunExitCode pure decision function for app run (D-08)
  - TestNoDuplicatedExitLiterals static scan with negative-control self-test (D-11, D-21)
  - TestRepairExitMirrorMatchesTable pinned mirror for cmd/schway-repair (D-22)
  - TestEmittedCExitLiteralsMatchTable generated-C exit text pin
affects: [28-03, 28-04, 28-05, 28-06, 28-07]
tech-stack:
  added: []
  patterns: [leaf constant package with AST scan guard, pinned mirror for import-boundary-restricted binary]
key-files:
  created:
    - internal/compiler/exitcode/exitcode.go
    - internal/compiler/exitcode/exitcode_test.go
    - internal/compiler/exitcode/scan_test.go
    - cmd/schway/main_exitcode_test.go
    - internal/compiler/cgen/cgen_exit_literals_test.go
  modified:
    - internal/compiler/protocol/protocol.go
    - internal/compiler/native/native.go
    - internal/compiler/native/native_app.go
    - internal/compiler/native/sanitize.go
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_sanitize.go
    - cmd/schway/main.go
    - cmd/schway-repair/main.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "D-07 one-way door (65/69 published) recorded as pre-approved; values unchanged, native.EventCapacityExitStatus aliases exitcode.EventCapacity"
  - "schway-repair keeps literal consts as a pinned mirror verified by test (D-06-28 forbids internal imports)"
requirements-completed: [DX-17]
status: complete
commits: 3
plan_head_before: 44c42491fba9e3e22864aad18be28f0fd80b5c78
plan_head_after: 715a3f5b194ac829b0b6172ba8bf3479e2f39bb3
actuals:
  tokens: 22000
  tasks: 3
  commits: 3
duration: 25min
completed: 2026-10-10
---

# Phase 28 Plan 02: Exit-code table and literal scan Summary

One zero-import `exitcode` package now holds every exit value; protocol, native, session and cmd/schway read from it, and an AST scan with synthetic negative controls proves no exit literal remains. Behavior-neutral: all exit values unchanged.

## Tasks

1. Tracer (commit 2a1d4b2): exitcode package, `TestExitCodeTable`, `protocol.ExitCode` and `native.EventCapacityExitStatus` rewired. CLI taxonomy and the 69 contract tests stay green.
2. Rewire (commit f0bff83): `appRunExitCode` extracted from `runApplicationRun` with `TestAppRunExitCode` (13 table subtests plus `nonexistent-artifact`, exit 125 with `schway app run:` prefix). Success literals in native and session replaced. schway-repair const block documented as pinned mirror. stderr text unchanged.
3. Scan and pin (commit 715a3f5): `TestNoDuplicatedExitLiterals`, `TestExitScanDetectorFires` (9 firing rows plus clean control), `TestRepairExitMirrorMatchesTable`, `TestEmittedCExitLiteralsMatchTable` (exit set {0, 65, 69}; return set {0,1,2,64,65,70,71,74}). No cgen source change.

## Verification (all observed passing)

- `go test ./internal/compiler/exitcode -run '^(TestExitCodeTable|TestNoDuplicatedExitLiterals|TestExitScanDetectorFires|TestRepairExitMirrorMatchesTable)$' -count=1`
- `go test ./cmd/schway -run '^(TestAppRunExitCode|TestPhase22AppVerifyIndependentIdentityCases|TestPhase27PublicAppCLI|TestPhase273AppVerifyCapacityCLI)$' -count=1`
- `go test ./cmd/schway-repair -run '^(TestRepairDriverImportsStayOutsideInternal|TestImportBoundaryTestIsNotInert|TestPhase17RepairProtocolFields)$' -count=1`
- `go test ./internal/compiler/native -run '^(TestPhase273RunnerEventCapacityAdjudication|TestNonzeroExitIsDistinctFromSignal)$' -count=1`
- `go test ./internal/compiler/cgen -run '^(TestEmittedCExitLiteralsMatchTable|TestPhase273CapacityContractLiteralsAgree)$' -count=1`
- `go test ./internal/compiler/session -run '^(TestPhase273ConformanceCorpusByteIdentity|TestPhase272AcceptedCorpusByteIdentity|TestPhase28EntryExitClassManifest|TestPhase16PublicEmitterConsumerInventory)$' -count=1`
- `go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle' -count=1` (clean)
- `go list -f '{{len .Imports}}' ./internal/compiler/exitcode` prints 0; `go build ./...` and `go vet` on touched packages clean.

## Deviations from Plan

**1. [Rule 1 - Bug] Line-keyed emitter registry shifted by new import**
- **Found during:** Task 3 meta-test run
- **Issue:** Adding the exitcode import to session_phase5.go moved its `EmitNative` call from line 26 to 27, so `TestPhase16PublicEmitterConsumerInventory` failed (missing and stale registry rows). The plan's file list did not include the registry.
- **Fix:** Re-keyed the row to `session_phase5.go:EmitNative:27` in `testdata/phase16/public-emitter-consumers.json`.
- **Commit:** 715a3f5. Note: the inventory test was red at the intermediate Task 2 commit f0bff83 and green again from 715a3f5.

**2. [Process] No separate RED commit for Task 2.** The new test was written first and failed to compile (appRunExitCode undefined) but was committed together with the implementation, matching the plan's single commit message.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

All created files exist and commits 2a1d4b2, f0bff83, 715a3f5 are ancestors of HEAD.

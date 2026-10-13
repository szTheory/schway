---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 08
subsystem: session/native
tags: [gap-closure, G-28-3, native, relative-path, exit-classes]
requires: []
provides:
  - relative-path native command lane reaching Check/admitProgram
affects: [DX-17, DX-18]
tech-stack:
  added: []
  patterns: [no path arithmetic before Check in native lanes]
key-files:
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/native/native_phase28_test.go
  created:
    - internal/compiler/session/session_phase28_native_relpath_test.go
decisions:
  - "Removed unused build-root filepath.Rel from runNativeCommand and RunNativeFile; fixture arg is now filepath.ToSlash(path)."
metrics:
  tasks: 2
status: complete
commits: 2
plan_head_before: 01cb969a93a7ba192bcd392c0248903f8edbad5a
plan_head_after: 0d91b86e6a819c974f60fe3a56bd5463e0aa6196
actuals:
  tokens: 6000
  tasks: 2
  commits: 2
---

# Phase 28 Plan 08: Relative-path native run (G-28-3) Summary

Removed the fatal build-time-root `filepath.Rel` from both native file lanes so `schway run --engine=native <relative path>` reaches Check and admitProgram, exiting 2 for invalid sources and 0 for valid ones.

## Tasks

1. Tracer: session fix plus `TestPhase28NativeCommandRelativePath` (commit e3497d2). RED was observed first: all three subtests returned `native.tool_failure` / operational_failure. After the fix all pass. Shipped binary from repo root: duplicate_function_name relative exits 2 with `core.duplicate_function_id`; call_basic relative exits 0.
2. Shipped-binary regression (commit 0d91b86): factored `assertUniformRefusal`, added `relative_path/duplicate_function_name` and `relative_path/native_positive_control` rows to `TestExitClassesUniformAcrossSurfaces`. All absolute rows still pass.

## Verification

Focused session, native, cgen RunNativeFile callers and the four session meta-tests (TestEntryFunctionCalledOnlyViaResolveEntry, TestVerificationGroundedness, TestPhase20ValidationLifecycle, TestStaticTestIndexMatchesGoTestList) all pass. `grep` confirms no `filepath.Rel(` or `nat03ProjectRoot` remain in session.go.

## Deviations from Plan

The optional local experiment of reverting the session.go hunk to observe the CLI-level exit=3 was not run separately. The equivalent pre-fix failure was observed at session level (RED run, operational_failure / native.tool_failure, which the CLI maps to exit 3).

## Known Stubs

None.

## Self-Check: PASSED

Created files exist; commits e3497d2 and 0d91b86 are on the branch.

---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 05
subsystem: native-build-identity
tags: [identity, receipt, evidence, modules, MOD-06]
requires: [31-01]
provides: [module-digest identity, receipt module validation]
affects: [31-07]
key-files:
  created:
    - internal/compiler/session/session_moduledigests.go
    - internal/compiler/session/session_phase31_modules_test.go
    - internal/compiler/session/session_phase31_export_test.go
    - internal/compiler/native/native_phase31_modules_test.go
  modified:
    - internal/compiler/native/bindings.go
    - internal/compiler/native/native_app.go
    - internal/compiler/session/session.go
decisions:
  - "Modules added alongside SourceDigest (D-11/D-19); trailing omitempty keeps every no-import ID byte-identical."
  - "Import-order control compares InputID when the host clang is not reproducible (ExecutableDigest never zeroed)."
status: complete
commits: 3
plan_head_before: df5c175251979df44e4824b656355c023d7893f7
plan_head_after: 9302e53c
actuals:
  tokens: 30000
  tasks: 3
  commits: 3
---

# Phase 31 Plan 05: Module digests in identity Summary

Every module file that reached the compiler is now sha256-hashed from the single modload read into BuildIdentityInputs, BuildReceipt and ApplicationEvidenceReport as sorted `modules` rows (omitted entirely for no-import programs).

## What was built
- `native.ModuleDigest`, trailing `Modules` (omitempty) on identity, receipt, report; `Runner.BuildApplicationModules` (BuildApplication delegates with nil).
- `validateReceiptModules` in `readApplicationReceipt`: sorted/unique, >=2 rows, 64-hex digests, relative non-`..` paths, equals Identity.Modules, entry row links to SourceDigest; modules without identity refused.
- `checkSourcePathModules` / `moduleDigests` in session_moduledigests.go (one `loadSourcePath` call). session.go edited line-neutrally (5/5).
- Tests: omitempty invariance with non-omitempty control, 9 tampered-receipt cases, receipt size, checksum receipt rows, comment-only library edit, import order (incl. set-order via ModuleDigestsForTest), no-import, evidence, two-apps-same-library.

## Deviations from Plan

**1. [Rule 3 - Blocking] Host clang does not reproduce executables.** Two builds of the identical input yield different ExecutableDigest on this darwin host, so the import-order ExecutableDigest/ID equality (D-21) cannot hold. Per instruction ExecutableDigest is never zeroed; the test runs a same-dir rebuild control and, only when the toolchain is non-reproducible, compares the product's own `InputID()` (ExecutableDigest-free) instead of `ID()`, and logs it. On a reproducible toolchain the strict ExecutableDigest and ID checks apply. Commit 9302e53c. Reported for orchestrator awareness.

**2.** Receipt validation (Task 2's code) landed in Task 1's commit because native_app.go was edited together; Task 2's commit contains the tests.

## Known Stubs
None.

## Self-Check: PASSED

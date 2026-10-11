---
phase: 28-exit-classes-verify-report-and-entry-pruning
reviewed: 2026-10-10T00:00:00Z
depth: standard
files_reviewed: 48
files_reviewed_list:
  - README.md
  - cmd/schway-repair/main.go
  - cmd/schway/main.go
  - internal/compiler/callgraph/callgraph.go
  - internal/compiler/callgraph/prune.go
  - internal/compiler/exitcode/exitcode.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/sanitize.go
  - internal/compiler/protocol/protocol.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_admit.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_entry.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_mismatch.go
  - internal/compiler/session/session_phase5_sanitize.go
findings:
  critical: 0
  warning: 4
  info: 2
  total: 6
status: issues_found
---

# Phase 28: Code Review Report

**Reviewed:** 2026-10-10
**Depth:** standard
**Files Reviewed:** the production Go files above in full diff form. Test, testdata, and example files were skimmed for reliability only.

## Summary

I found no path where a refusal becomes exit 0 or a silent skip.

- Every command-file lane (check, interpreter run, native run, build, app verify) runs Check, then `admitProgram` on the full program, then `resolveEntry`. Admission therefore precedes pruning (D-18).
- `resolveEntry` errors that are not entry refusals propagate as errors, not as success.
- `app verify` maps `*SourceRefusalError` to exit 2 and keeps all other failures at exit 3.
- Exit-code constants route through the single `exitcode` table.

The remaining issues are robustness and diagnostic-quality defects, listed below.

## Warnings

### WR-01: Duplicate-`main` ambiguity reports the wrong roots and a misleading hint

**File:** `internal/compiler/callgraph/callgraph.go:315-326`, `internal/compiler/session/session_entry.go:52-55`
**Issue:** When two or more functions are named `main`, `EntryFunction` returns `&entryAmbiguousError{roots: roots}`. `roots` holds the in-degree-zero nodes, not the colliding `main` functions. The diagnostic can therefore name the wrong functions, or an empty list (for example when one `main` has a caller). It also appends "add `fn main` to disambiguate", which is wrong advice when `main` is exactly what is duplicated. The refusal itself is correct (fail-closed). The message misleads the user and the Phase 28 ambiguity test may not cover this case.
**Fix:** Build the error from the colliding IDs:
```go
if len(mains) > 1 {
    sort.Strings(mains)
    return core.Function{}, &entryAmbiguousError{roots: mains}
}
```
Then make `entryRefusalDiagnostic` omit the `fn main` hint when the roots are all named `main`. Alternatively, give the error type a reason field.

### WR-02: A source refusal can be masked as exit 3 when the report write fails

**File:** `internal/compiler/session/session_app_verify.go:219-223`
**Issue:** `VerifyApplicationCasesFile` returns the report-write error in preference to `verifyErr`. If writing the (empty, failing) report fails, the `*SourceRefusalError` is discarded and the CLI exits 3 instead of 2. This breaks the D-21 exit-class contract on that path. The same applies to `ErrApplicationVerificationReportOversize`.
**Fix:** Join the errors so `errors.As` still finds the refusal:
```go
if werr := writeApplicationVerificationReport(reportPath, report); werr != nil {
    if verifyErr != nil { return report, errors.Join(verifyErr, werr) }
    return report, werr
}
```

### WR-03: A stale report file survives a refused or oversize run

**File:** `internal/compiler/session/session_app_verify.go:636-644`, `internal/compiler/session/session_app_verify.go:219`
**Issue:** On `ErrApplicationVerificationReportOversize` the code writes no file ("no file is written when it fires"). A previous run's report at `reportPath` stays on disk, possibly with `verified: true`. The process exits non-zero, but a consumer that reads only the file would see a stale pass. This is a fail-open risk for report-reading automation.
**Fix:** Remove or truncate any existing `reportPath` before verification, or write a minimal fixed-size failure report on the oversize path.

### WR-04: A non-diagnostic `resolveEntry` error is silently tolerated in app verify

**File:** `internal/compiler/session/session_app_verify.go:275-283`
**Issue:** When `resolveEntry` returns a plain error (for example a `buildAdjacency` failure) with no diagnostics, `entryErr` is stashed and the unpruned program is used. Only source-kind cases check `entryErr`; they are marked `unsupported`. A cases file containing only `verifier_model_only` cases never consults `entryErr`, so it can end `Verified == true` for a source whose entry could not be resolved. Every other command treats this as an error.
**Fix:** Return `entryErr` immediately, as the other lanes do:
```go
if entryErr != nil { return report, entryErr }
```
Then drop the later `entryErr != nil` guard.

## Info

### IN-01: `appRunExitCode` collides with child exit statuses by design

**File:** `cmd/schway/main.go:257-273`
**Issue:** A child's own exit status passes through unchanged, so a child exiting 65, 124 or 125 is indistinguishable from the host's `AppInputRefused`, `AppTimeout` or `AppLaunchError`. D-08 appears to accept this. Confirm the README exit table says so explicitly.
**Fix:** Add a one-line note in the README exit table, or ignore if already documented.

### IN-02: `Phase4CheckedProgram` discards the pruned program

**File:** `internal/compiler/session/session.go:605-612`
**Issue:** `Phase4CheckedProgram` calls `resolveEntry` only for its refusal and its entry. It returns the unpruned program, which is intentional per the comment. Callers must now prune per engine, and nothing enforces that. A future caller that forgets will feed dead-code roots to an engine (see D-19 about a single in-degree-zero root).
**Fix:** Add a doc comment on the function stating that its return value is unpruned. Alternatively, export a typed wrapper that returns both forms.

---

_Reviewed: 2026-10-10_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

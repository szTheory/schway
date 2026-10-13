---
phase: 28-exit-classes-verify-report-and-entry-pruning
fixed_at: 2026-10-10T00:00:00Z
review_path: .planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-REVIEW.md
iteration: 1
findings_in_scope: 6
fixed: 6
skipped: 0
status: all_fixed
---

# Phase 28: Code Review Fix Report

**Fixed at:** 2026-10-10
**Source review:** .planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 6
- Fixed: 6 (IN-01 resolved as already documented, no code change)
- Skipped: 0

## Fixed Issues

### WR-01: Duplicate-`main` ambiguity reports the wrong roots and a misleading hint

**Files modified:** `internal/compiler/callgraph/callgraph.go`, `internal/compiler/callgraph/callgraph_entry_test.go`, `internal/compiler/session/session_entry.go`
**Commit:** 94fcbb3
**Applied fix:** The duplicate-main refusal now carries the sorted colliding `main` function IDs and a `duplicateMain` flag (exposed as `DuplicateMain()`). `Error()` and `entryRefusalDiagnostic` use a duplicate-main message ("keep exactly one `fn main`") instead of the misleading "add `fn main`" hint. Other ambiguity messages are unchanged. Added a unit test. Logic fix: requires human verification.

### WR-02: A source refusal can be masked as exit 3 when the report write fails

**Files modified:** `internal/compiler/session/session_app_verify.go`, `internal/compiler/session/session_phase28_appverify_test.go`
**Commit:** d665d48
**Applied fix:** `VerifyApplicationCasesFile` returns `errors.Join(verifyErr, writeErr)` so `errors.As` still finds `*SourceRefusalError` (CLI exit 2) and `errors.Is` still finds the oversize sentinel. Added a test using an unwritable report path. Logic fix: requires human verification.

### WR-03: A stale report file survives a refused or oversize run

**Files modified:** `internal/compiler/session/session_app_verify.go`, `internal/compiler/session/session_phase28_appverify_test.go`
**Commit:** 29cb469
**Applied fix:** New `removeStaleVerificationReport` deletes any pre-existing report at the start of `VerifyApplicationCasesFile` (skipped only if the report path is the same file as an input). Added a test that a stale `verified:true` report does not survive a refused run. Logic fix: requires human verification.

### WR-04: A non-diagnostic `resolveEntry` error is silently tolerated in app verify

**Files modified:** `internal/compiler/session/session_app_verify.go`
**Commit:** 3743ef3
**Applied fix:** `VerifyApplicationCases` now returns `entryErr` immediately after the diagnostics check; the later `entryErr != nil` guard and the unpruned-program fallback were removed. Logic fix: requires human verification.

### IN-01: `appRunExitCode` collides with child exit statuses by design

**Files modified:** none
**Commit:** none
**Applied fix:** No change needed. README.md (Exit codes section, after the table) already states that `schway app run` passes the child's exit status through unchanged, so it may coincide with a host code, and names the stderr prefix and `--report` as the disambiguators. Resolved as already documented.

### IN-02: `Phase4CheckedProgram` discards the pruned program

**Files modified:** `internal/compiler/session/session.go`, `testdata/phase16/public-emitter-consumers.json`
**Commit:** 8436391 (doc comment), 307e14a (registry line-number follow-up)
**Applied fix:** Added a doc comment stating the return value is unpruned and callers must prune per engine. The six added lines shifted `EmitNative` in `session.go`, which tripped `TestPhase16PublicEmitterConsumerInventory`; the registry entry was re-grounded from line 3041 to 3047.

## Verification

Gates ran in the isolated worktree (go has no node_modules dependency; not reproducible from the worktree after teardown, but the commits were fast-forwarded onto `phase26-validation`). `go build ./...`, `go vet` on callgraph, session and cmd, and the full `go test ./...` all pass. The only failure along the way was the Phase 16 inventory meta-test, fixed in 307e14a and re-run green together with the session, scripts and testsupport packages. `.planning/state.json` was left untouched and unstaged.

---

_Fixed: 2026-10-10_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_

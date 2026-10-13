---
phase: 28-exit-classes-verify-report-and-entry-pruning
reviewed: 2026-10-11T00:00:00Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - internal/compiler/callgraph/callgraph.go
  - internal/compiler/callgraph/callgraph_entry_test.go
  - internal/compiler/native/native_phase28_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_entry.go
  - internal/compiler/session/session_phase28_appverify_test.go
  - internal/compiler/session/session_phase28_native_relpath_test.go
  - testdata/phase16/public-emitter-consumers.json
findings:
  critical: 1
  warning: 1
  info: 0
  total: 2
status: issues_found
---

# Phase 28: Code Review Report (incremental, diff_base 1beb03a9)

**Reviewed:** 2026-10-11
**Depth:** standard
**Files Reviewed:** 9

## Summary

The G-28-3 change is sound. `Phase16ControlNativeC` ignores its `fixture` argument, so replacing the root-relative path with the caller's path has no effect on emitted C or evidence. `nat03ProjectRoot` is still used elsewhere, so removing the call leaves no dead code. The duplicate-main refusal, `removeStaleVerificationReport` and `errors.Join` changes are correct. `cmd/schway/main.go` uses `errors.As`, so exit 2 is preserved. The stale-report removal does not delete inputs, and a symlinked report path only loses the link.

One real regression remains. The emitter-consumer registry is out of sync with the source.

## Critical Issues

### CR-01: Emitter consumer registry line number is stale, so the inventory test fails at HEAD

**File:** `testdata/phase16/public-emitter-consumers.json:272`
**Issue:** The entry says `session.go:EmitNative:3047`, but the `cgen.EmitNative` call is at `internal/compiler/session/session.go:3042`. Line 3047 is `cleanRunner := runner`. Commit 307e14a ("re-ground emitter consumer registry line after doc comment") moved the pin to the wrong line. I ran `go test ./internal/compiler/session -run 'EmitterInventory|PublicEmitter|Phase16.*Emitter'` and it fails:
```
TestPhase16PublicEmitterConsumerInventory
  missing registry entry for source call internal/compiler/session/session.go:EmitNative:3042
  stale registry entry internal/compiler/session/session.go:EmitNative:3047
```
**Fix:** Change the line to 3042:
```json
"call": "internal/compiler/session/session.go:EmitNative:3042",
```
Then run `TestPhase16PublicEmitterConsumerInventory` after any further edit to `session.go`. The pin is line-based, so it breaks whenever lines shift above the call.

## Warnings

### WR-01: Registry pins are fragile and were re-grounded blind

**File:** `testdata/phase16/public-emitter-consumers.json:272`
**Issue:** The prior review already flagged this line (IN-02), and the follow-up fix still did not match the source. The session meta-tests were evidently not run after the doc-comment edit to `session.go`, even though the project memory says to run them before dispatch.
**Fix:** Run the inventory, groundedness and validation-lifecycle session tests as a gate before committing any `session.go` edit. Consider making the registry anchor on the function and call text rather than a raw line number.

---

_Reviewed: 2026-10-11_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

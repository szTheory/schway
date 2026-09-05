---
phase: 04-fallible-resources-and-c-boundary
fixed_at: 2026-09-05T13:26:52Z
review_path: .planning/phases/04-fallible-resources-and-c-boundary/04-REVIEW.md
iteration: 1
findings_in_scope: 4
fixed: 4
skipped: 0
status: all_fixed
---

# Phase 4: Code Review Fix Report

**Fixed at:** 2026-09-05T13:26:52Z
**Source review:** .planning/phases/04-fallible-resources-and-c-boundary/04-REVIEW.md
**Iteration:** 1

**Verification environment:** `workflow.use_worktrees` is `false` in `.planning/config.json`, so all edits, syntax/build checks, and commits below ran directly in the main checkout on branch `main` (no isolated worktree was created; there is nothing to fast-forward or clean up).

**Summary:**
- Findings in scope: 4 (critical_warning: CR-01, WR-01, WR-02, WR-03; IN-01/IN-02 out of scope for this fix_scope)
- Fixed: 4
- Skipped: 0

## Fixed Issues

### CR-01: `corevalidate.checkReleaseOrder` silently skips validation for any terminal block that doesn't have exactly one incoming edge

**Files modified:** `internal/compiler/corevalidate/corevalidate.go`
**Commit:** `16fb0c9`
**Applied fix:** The review's literal suggested fix ("hard-refuse any terminal block whose incoming-edge count isn't exactly 1") was tested first and found to be **too strict**: it broke two real, previously-passing test cases (`TestValidatorRederivesReleaseOrder`, `TestReleaseOrderValidationWorkSeries`) exercising `discard_because.lang`, a legitimate program where `discard <call> because "..."` merges its ok/err edges into the same successor block — a valid merge with nothing tracked to release on either path. Adapted the fix instead: `checkReleaseOrder` now walks backward from **every** incoming edge into a terminal block (not just refusing multi-edge blocks, and not just checking edge `[0]`), and verifies the block's single, fixed release list against each rederived expectation independently. A hand-corrupted merge of two acquisition chains with genuinely divergent completed-acquisition sets can match at most one of those expectations, so it is still caught — closing the vacuous-skip gap the review identified — while legitimate merges (like `discard`) where every incoming path expects the same (empty) release list continue to validate correctly. Verified: `go build ./...`, `go test ./internal/compiler/corevalidate/... ./internal/compiler/native/...` and the full `go test ./...` all pass.

### WR-01: `control:foreign.no_unproven_attributes` never scans two of the three generated inspectable layers

**Files modified:** `internal/compiler/session/session.go`, `internal/compiler/session/session_test.go`
**Commit:** `cb701b5`
**Applied fix:** Added `cgen.EmitForeignHeader(...)` and `cgen.EmitForeignConformance(..., native.ForeignResourcePrivateHeaderPath())` output (for both the tracer and release-fixture programs) to the `lane:foreign-no-unproven-attributes` production scan in `session.go`, and to `TestNoUnprovenAttributesEmitted`. Added a new dedicated mutation-kill test, `TestAttributeInjectionIntoHeaderOnlyMakesControlFail`, which injects a banned token into `EmitForeignHeader`'s output alone (leaving `emitLinearForeign`'s own extern declaration untouched) and proves the header artifact is now independently scanned rather than only incidentally caught via the two format strings' prior coincidental overlap. Verified: `go build ./...`, `go test ./internal/compiler/session/... ./internal/compiler/cgen/...`, full `go test ./...`, and `scripts/verify-phase4.sh` all pass — the gate's `lane:foreign-no-unproven-attributes` recomputed_work rose from 4 to 8, confirming the new artifacts are actually being scanned in the production lane, not just in tests.

### WR-02: `interp.RunLinearBlockDirect` is exported production API that exists solely to serve a test

**Files modified:** `internal/compiler/interp/interp.go`, `internal/compiler/interp/interptestdirect/interptestdirect.go` (new), `internal/compiler/session/session_test.go`
**Commit:** `e3fa508`
**Applied fix:** Moved `RunLinearBlockDirect` and its `foreignFailureLiteralDirect` helper out of `interp.go` into a new package, `internal/compiler/interp/interptestdirect`, so the "test-only" boundary is a compile-time import fact rather than a disclaimer in a doc comment. Because Go's `_test.go` convention cannot expose symbols across package boundaries (the caller, `session_test.go`, is in a different package than `interp`), a build-tag-guarded file was not viable without complicating every test invocation; a plain importable sibling package was the correct shape per the review's own alternative suggestion ("or into an internal test-support package under `interp`"). Following this project's own established D-12 convention (independently re-deriving shared facts across packages rather than coupling them through shared helpers — the same pattern `foreignFailureLiteralDirect`'s own original comment cites), the new package duplicates interp's small private `findFunction`/`ownedEvent`/`liveResourceList` helpers instead of exporting them from `interp`, so `interp`'s real production surface is not widened to compensate for shrinking it elsewhere. `session_test.go` now calls `interptestdirect.RunLinearBlockDirect`. Verified: `go build ./...`, `go vet ./...`, full `go test ./...` (including `internal/compiler/session`, which exercises this call site) all pass.

### WR-03: `evidence.Validate` compares `ForeignDigest` in constant time while every sibling digest comparison uses plain equality

**Files modified:** `internal/compiler/evidence/evidence.go`
**Commit:** `cbe1788`
**Applied fix:** Investigated the actual code at the cited lines and found the review's factual premise did not hold against the current file: `SourceDigest`, `CoreDigest`, `CDigest`, and `ID` are **already** compared with `subtle.ConstantTimeCompare` in this function (lines 383-391), identically to `ForeignDigest` — there is no plain-equality/constant-time asymmetry among the digest fields. The fields that genuinely use plain `!=`/`==` (`Schema`, `IDAlgorithm`, `SourceSchema`, `CompilerIdentity`, etc., in the `checks` slice at lines 342-356) are a different category — fixed schema/identity tags, not content digests — and are correctly, deliberately excluded from the constant-time treatment. Rather than skip the finding outright, applied the review's own offered alternative ("or document why this one field specifically needs constant-time comparison when its siblings don't"): added a comment above the `ForeignDigest` check explaining that all content-digest fields intentionally share the constant-time convention for uniformity (not because of a secret-bearing threat model), and that schema/identity tags are the deliberately-different, plain-compared category — so a future reader (or reviewer) doesn't rediscover this same false inconsistency. No comparison logic was changed. Verified: `go build ./...`, `go test ./internal/compiler/evidence/...`, full `go test ./...` pass.

## Skipped Issues

None — all 4 in-scope findings (CR-01, WR-01, WR-02, WR-03) were fixed. IN-01 and IN-02 were out of scope for `fix_scope: critical_warning` and were not attempted.

---

_Fixed: 2026-09-05T13:26:52Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_

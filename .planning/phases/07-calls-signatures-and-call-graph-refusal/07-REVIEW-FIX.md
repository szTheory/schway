---
phase: 07-calls-signatures-and-call-graph-refusal
fixed_at: 2026-09-09T00:00:00Z
review_path: .planning/phases/07-calls-signatures-and-call-graph-refusal/07-REVIEW.md
iteration: 1
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 07: Code Review Fix Report

**Fixed at:** 2026-09-09
**Source review:** .planning/phases/07-calls-signatures-and-call-graph-refusal/07-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 1 (WR-01 only -- `fix_scope: critical_warning`, 0 critical findings; IN-01/IN-02 are Info severity and out of scope for this run)
- Fixed: 1
- Skipped: 0

## Fixed Issues

### WR-01: `consumeCallArgument`'s ability re-derivation is keyed off the callee's return type, not the argument's own declared type, and this equivalence is not defended by any operation-kind-specific check

**Files modified:** `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/corevalidate/corevalidate_call_argument_consume_internal_test.go`
**Commit:** 3f274ad
**Applied fix:** Chose the review's second offered option (an explicit
`OpCall`-specific assertion tied to the D-07-09 constraint it depends on)
over deriving from `types[source.TypeID].Shape` directly, because it
produces a defense that is local to `consumeCallArgument` itself and
independently test-caught, per the review's own stated intent ("a fix
that makes the dependency explicit and test-caught is what is wanted").

- `consumeCallArgument`'s signature grew a `sourceTypeID string`
  parameter. Both call sites (`replayStraightLine` and `replayBlocks`,
  each already holding `source := places[operation.SourceID]` in scope)
  now pass `source.TypeID` explicitly instead of relying on the generic,
  not-OpCall-specific `source.TypeID == operation.TypeID` law checked
  earlier in the same loop.
- Inside `consumeCallArgument`, before the `v.derive` call, a new
  `v.check(operation.TypeID == sourceTypeID, "core.type_mismatch",
  operation.ID)` fails closed if the two ever diverge, with a comment
  tying the assertion to D-07-09 (a function's declared return type must
  equal its own declared parameter type) and explaining why it is not
  redundant with the pre-existing generic check.
- The doc comment above `consumeCallArgument` was extended to explain why
  the equivalence is a derived consequence of two other, separately
  checked facts (not an IR invariant of `OpCall` itself, unlike
  `OpCopy`/`OpMove`/`OpBorrow*`), and why this function now defends it
  locally rather than trusting the caller-side law.
- Updated the two existing internal tests
  (`TestConsumeCallArgumentDerivesFromOwnShapeNotRecordedAbilities`,
  `TestConsumeCallArgumentConsumesNonCopyableFromOwnShape`) to pass the
  new `sourceTypeID` parameter (equal to their existing `typeID`, since
  neither test previously modeled a source/operation type divergence).
- Added a new negative test,
  `TestConsumeCallArgumentRefusesWhenSourceTypeDivergesFromOperationType`,
  which drives `consumeCallArgument` directly with a deliberately
  diverging `sourceTypeID` and asserts it is refused fail-closed (not a
  silent consume) -- this is the "test-caught" half of the fix the review
  asked for.

**Constraints honored:**
- No shared helper, type, or constant introduced between `corevalidate`
  and `check`; the fix is entirely local to the `corevalidate` package,
  touching only `consumeCallArgument` and its two same-package call sites.
- No existing test weakened or deleted. The phase-07 control registry
  triple (`session.Phase7RequiredControls()`,
  `scripts/verify-phase7.sh`, `session_phase7_test.go`'s
  `controlsWithRecordedMutationKill`) was not touched and still agrees
  (24/24 controls, verified below).
- No fixture verdict changed: `call_argument_used_twice.lang` still
  refuses `ownership.use_after_move`;
  `call_argument_used_once.lang` still admits (both exercised end-to-end
  by the full test suite and `verify-phase7.sh` below).

## Verification

All commands run against the real tree (no worktree isolation was used --
`workflow.use_worktrees` is `false` in `.planning/config.json`, so the fix
was edited and committed directly on `main` per the documented opt-out):

- `go build ./...` -- pass, no output.
- `go vet ./...` -- pass, no output.
- `go test ./... -p 1` -- all packages `ok`, including
  `internal/compiler/corevalidate` (new/updated tests pass) and
  `internal/compiler/session` (274s, drives the native/CLI corpus).
- `sh scripts/verify-phase7.sh` -- exit code 0. Final lane
  (`lane:kind-exhaustive-dispatch-phase07`) reports `status: pass` with
  all 24 phase-07 controls present, including
  `control:call.argument_consumed_when_noncopyable` and
  `control:call.copyable_argument_not_consumed` (the two controls this
  fix's derivation path feeds).
- `git status --short go.mod go.sum` -- no output; `go.mod`/`go.sum`
  unmodified.

## Skipped Issues

None -- WR-01 was the entire in-scope finding set for this run
(`fix_scope: critical_warning`) and it was fixed. IN-01 and IN-02 are
Info-severity and deliberately out of scope; they were not attempted.

---

_Fixed: 2026-09-09_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_

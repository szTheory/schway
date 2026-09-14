---
phase: 07-calls-signatures-and-call-graph-refusal
fixed_at: 2026-09-09T20:48:58Z
review_path: .planning/phases/07-calls-signatures-and-call-graph-refusal/07-REVIEW.md
iteration: 1
findings_in_scope: 2
fixed: 2
skipped: 0
status: all_fixed
---

# Phase 07: Code Review Fix Report

**Fixed at:** 2026-09-09T20:48:58Z
**Source review:** .planning/phases/07-calls-signatures-and-call-graph-refusal/07-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 2 (`fix_scope: all` — both are Info-severity; REVIEW.md's frontmatter `status: clean` reflects 0 Critical/0 Warning, not "nothing to fix")
- Fixed: 2
- Skipped: 0

**Verification environment:** `workflow.use_worktrees` is `false` in `.planning/config.json`, so this run edited and committed directly in the main checkout (no isolated worktree was created; no fast-forward/cleanup tail applies). All numbers below are reproducible from this tree as-is.

## Fixed Issues

### IN-01: `originvalidate/export_test.go` still has a stray trailing blank line

**Files modified:** `internal/compiler/originvalidate/export_test.go`
**Commit:** `c8c3d8f`
**Applied fix:** Removed the extra blank line after the file's final `}` (confirmed via hex dump before/after: file previously ended `...7d0a 7d0a 0a` — `}\n}\n\n` — now ends `...7d0a 7d0a` — `}\n}\n`). File is gofmt-clean; no other content touched.

### IN-02: `peerJoinFails`'s determinism still depends on an unenforced cross-package sort-order coincidence

**Files modified:** `testdata/phase07/call_two_fallible_callees_disagree.lang` (new), `internal/compiler/corevalidate/corevalidate_foreign_closure_test.go`, `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md`
**Commit:** `590f680`
**Applied fix:** The existing `TestPeerFailsIsClosureDerived` only exercised `joinFails`/`peerJoinFails`'s identity/equal-merge cases (a single fallible chain via `call_fallible_foreign_reach.lang`), never a genuinely disagreeing multi-callee shape where the fold's result actually depends on discovery/sort order. Added a new fixture, `call_two_fallible_callees_disagree.lang`, whose `main` calls two callees (`tracer_a`, `tracer_b`) publishing different non-empty `Fails` values (`ErrA`, `ErrB`), and a new test, `TestPeerFailsAgreesOnDisagreeingMultiCalleeJoin`, that drives it through both `originvalidate.BuildInterface` (producer) and `corevalidate.Validate`'s `PeerSignatures()` (peer) and asserts the two `Fails` values agree — confirmed both sides publish `"ErrA"` (accumulator-wins: `main` calls `tracer_a` before `tracer_b`). This makes the previously-only-documented producer/peer sort-order coincidence test-caught, per the finding's own fix suggestion.

Per the project-context constraints for this fix: `corevalidate` was NOT given any new shared helper, type, or constant with `check`/`originvalidate` — the new test only *observes* agreement between the two independently-implemented sides, exactly the shape the existing `TestPeerFailsIsClosureDerived`/`TestPeerForeignReachIsClosureDerived` tests already use. `joinFails` itself was left unchanged (still deliberately not order-independent — D-07-53). A note was added to `PHASE-07-DEBT.md`'s existing D-07-53 entry (new paragraph, not an edit to the entry's existing text) stating that the producer/peer agreement is now test-caught; the debt's own scope (schema-level `Fails`-cannot-express-a-union imprecision, reopens on vocabulary/schema changes) is unchanged. No control-registry identifier was added (this finding did not warrant one), so the three-registry set-equality tests (`session.Phase7RequiredControls()`, `scripts/verify-phase7.sh`, `session_phase7_test.go`'s `controlsWithRecordedMutationKill`) required no changes and remain in sync.

## Verification

Ran to completion, in this order, on the resulting tree:

- `go build ./...` — clean, no output.
- `go vet ./...` — clean, no output.
- `go test ./... -p 1` — all packages pass (`ok` for every package with tests; no failures).
- `sh scripts/verify-phase7.sh` (~8-10 min: full suite plus `go test -race ./...`, then the phase 1-7 verify-lane gates) — exit 0. All required phase-07 controls present and passing, including `control:phase07.controls_are_mutation_killed`, `control:summary.fails_closure_derived`, `control:summary.foreign_reach_closure_derived`, and the rest of the `lane:kind-exhaustive-dispatch-phase07` control set.
- `go.mod` unmodified (not in `git status`); `go.sum` does not exist in this repository (no external dependencies) — no drift possible.

No fixture's verdict changed: `call_fallible_foreign_reach.lang` and every other existing `testdata/phase07/*.lang` fixture were left untouched; the new fixture (`call_two_fallible_callees_disagree.lang`) checks and corevalidates clean by design (a witness fixture, not a refusal fixture), consistent with the file's own doc comment.

---

_Fixed: 2026-09-09T20:48:58Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_

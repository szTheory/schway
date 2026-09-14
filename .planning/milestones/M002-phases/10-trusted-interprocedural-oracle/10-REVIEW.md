---
phase: 10-trusted-interprocedural-oracle
reviewed: 2026-09-11T00:00:00Z
depth: standard
files_reviewed: 27
files_reviewed_list:
  - internal/compiler/check/check.go
  - internal/compiler/check/check_ordering_stability_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go
  - internal/compiler/corevalidate/corevalidate_endpoint_test.go
  - internal/compiler/corevalidate/corevalidate_peer_origin_test.go
  - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/corevalidate/export_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_oracle_golden_test.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/originvalidate/export_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_closure_chain_test.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/export_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_compose.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_composition_depth_test.go
  - internal/compiler/session/session_peer_gate_test.go
  - internal/compiler/session/session_phase7.go
findings:
  critical: 0
  warning: 2
  info: 2
  total: 4
status: issues_found
---

# Phase 10: Code Review Report

**Reviewed:** 2026-09-11T00:00:00Z
**Depth:** standard
**Files Reviewed:** 27
**Status:** issues_found

## Summary

Reviewed the actual diff for Phase 10 (`git diff 55807cacbba5c4b66e033dcdb3cbd8ea6b4ae95d..HEAD`, since the `diff_base` passed in config points at the phase's own context-capture commit rather than its parent) across `check`, `core`, `corevalidate`, `interp`, `originvalidate`, `pathoracle`, and `session`.

The four re-derivers (`check`, `corevalidate`, `originvalidate`, `pathoracle`) each independently gained their own OpCall-consulting logic for this phase (frame drain, origin propagation, path composition), and in every case examined the new logic is a genuine re-derivation rather than one peer reading another's conclusion:

- `originvalidate.walkReturnOrigin`'s new `core.OpCall` case consults `calleeContracts` (built from `PublicOrigin`, the DECLARED contract) — never a peer's *computed* verdict.
- `corevalidate.peerCalleeFrameDrained` is a from-scratch structural predicate over `core.LinearOperation`, not shared with `interp`'s runtime drain order.
- `pathoracle`'s composition machinery (`pathoracle_compose.go`) recurses into `EnumeratePaths`/`linearizePath` again, never collapsing to a contract hop, preserving the "never converges, only replays paths" identity claim the package doc comment states explicitly.
- The three-way (four-way on accept) loan-endpoint/origin differentials in `session_peer_gate_test.go` and `corevalidate_endpoint_test.go` compare genuinely independently-computed sets (`function.Linear.LoanEndpoints`, `corevalidate.Result.LoanEndpoints()`, `pathoracle.RecomputeEndpoints`), not a value compared to itself through two names.

`MaxCallDepth` (128) and `MaxCompositionDepth` (3) are checked at correct, deliberate points in `interp.runFrameStack` and `pathoracle_compose.composeCall` respectively: the interp check occurs *after* `partitionFrameForCall` resolves the callee (so a bare-match-arm callee needing no frame is never wrongly refused), and both boundary values are exercised at both `n` and `n+1` in dedicated tests. `drainStackForAbruptExit` is invoked exactly once per abrupt-exit path (nonlocal-exit landing pad, depth-exceeded refusal) with no double-drain or double-count found. Fault-injection seams (`moveAsCopyForTest`, `frameDrainOrderForTest`, `forceContractHopForTest`, `disableOpCallOriginConsultForTest`, `parameterContractModeOverrideForTest`) are all restored via `defer` in every call site inspected, so no cross-test pollution risk was found.

Two warnings below are about an underdocumented ordering assumption and an accepted-but-newly-surfaced independence asymmetry; neither is a proven correctness defect but both are worth a second look.

## Warnings

### WR-01: `peerCalleeFrameDrained`'s forward escape-propagation relies on an unstated operation-ordering invariant

**File:** `internal/compiler/corevalidate/corevalidate.go:2850-2884` (the `reaches` propagation loop inside `peerCalleeFrameDrained`)
**Issue:** The escape-via-return check builds `reaches` by a single forward pass over `linear.Operations`:
```go
reaches := map[string]bool{acquisition.TargetID: true}
for _, operation := range linear.Operations {
    if operation.Kind != core.OpMove && operation.Kind != core.OpCopy {
        continue
    }
    if reaches[operation.SourceID] {
        reaches[operation.TargetID] = true
    }
}
```
This is a *single-pass* transitive-closure computation, not a fixpoint loop: it is only correct if every `OpMove`/`OpCopy` whose `SourceID` depends on the acquisition's `TargetID` appears *later* in `linear.Operations` than the operation that produced that `SourceID`. That happens to hold today because a place must be produced (as some earlier operation's `TargetID`) before it can be read as a later operation's `SourceID` — an invariant enforced elsewhere by "uninitialized place" checks — but this function does not check or assert that invariant itself, nor does it document the assumption the way `peerParameterEscapesOwned` (the function's own stated precedent) does: that function walks *backward* via a `sourceOf` map and is correct regardless of iteration order. `peerCalleeFrameDrained` instead depends silently on `linear.Operations`' declaration order matching dependency order. If a future block-lowering change ever produces a flat `Operations` list that is not already dependency-ordered (e.g. blocks emitted in a different traversal order than execution order for some CFG shape), this predicate could silently under-report an escape and admit a genuinely abandoned resource as drained — a false-negative in a fail-closed safety check.
**Fix:** Either document the relied-upon invariant explicitly at the top of `peerCalleeFrameDrained` (referencing whichever structural check already guarantees dependency-before-use ordering in `linear.Operations`), or make the function robust to arbitrary ordering the same way `peerParameterEscapesOwned` is (a backward walk from each `OpReturn`'s `SourceID` via a `sourceOf` map, rather than a forward single pass), so the two "trace forward from an acquisition" and "trace backward from a return" helpers do not silently differ in their tolerance for operation ordering.

### WR-02: `originvalidate`'s forbidden-import list for `callgraph` diverges from `corevalidate`'s equivalent guard, weakening one leg of the independence proof

**File:** `internal/compiler/originvalidate/originvalidate_test.go:27-35` (`originValidateForbiddenImports`)
**Issue:** The comment candidly discloses that `originvalidate`'s own forbidden-import guard deliberately excludes `/compiler/callgraph` (because `originvalidate.go` itself imports `callgraph` for `BuildInterface`'s topological ordering), while `corevalidate`'s equivalent guard (`corevalidate_test.go:157`) does forbid it. This means one of the four peers (`originvalidate`) is permitted to depend on a shared graph-ordering component that two of the peers explicitly forbid, and the asymmetry is enforced only by a code comment plus PHASE-10-DEBT.md, not by any test that would catch a *widening* of this exception (e.g. `originvalidate` later importing `callgraph`'s own cycle-detection logic and using it in place of pathoracle's/corevalidate's independent cycle guards, rather than merely its topological order). Since this phase explicitly calls out "does anything weaken the independence claim" as a review focus, this pre-existing-but-newly-relevant asymmetry between the four peers' forbidden-import lists is worth flagging even though it is documented as accepted debt rather than a silent gap.
**Fix:** No code change required if the debt entry is intentional and current, but the review should confirm PHASE-10-DEBT.md's entry for this asymmetry is still accurate post-Phase-10 (i.e., that `originvalidate`'s only use of `callgraph` remains the topological `Order` call and never reaches into `callgraph`'s cycle-detection internals), and consider adding a narrower guard that permits importing `callgraph.Order` specifically without permitting the whole package surface, to prevent silent widening.

## Info

### IN-01: `diff_base` in the review config points at a doc-only commit, not the phase's actual base

**File:** N/A (workflow config)
**Issue:** The `diff_base` value passed to this review (`9c145d76e33cbef4d543121904334d7b3be74466^`) resolves to `55807cacbba5c4b66e033dcdb3cbd8ea6b4ae95d`, which is correct, but `9c145d76e33cbef4d543121904334d7b3be74466` itself is `docs(10): capture phase context` — a documentation-only commit at the *start* of Phase 10, not its end. `git diff --stat` against that hash directly (without walking to its parent) produced no output at all, which could silently produce an empty/misleading review scope if a future invocation passes the phase's context-capture commit as `diff_base` without the `^` parent-selector already baked in by this run's config. This review used `55807cacbba5c4b66e033dcdb3cbd8ea6b4ae95d..HEAD` (with the `^` already applied) and confirmed a substantial, correctly-scoped diff.
**Fix:** No action needed for this review; noting for the workflow's own robustness that a `diff_base` resolving to zero changed files should be treated as a scoping error worth surfacing loudly, not silently accepted as "no files to review."

### IN-02: `MaxCallDepth` and `MaxPaths`/`MaxCompositionDepth` rationale comments are unusually long and repeat the same "direction is opposite" framing three times across two packages

**File:** `internal/compiler/interp/interp.go:13-32`, `internal/compiler/pathoracle/pathoracle_compose.go:21-85`
**Issue:** Not a functional defect, but the extensive prose-as-documentation style (multi-paragraph rationale blocks citing decision IDs like D-10-21/D-10-23/D-10-46/D-10-47 inline in production source) makes the actual code harder to scan, and the same "MaxPaths sits above its ceiling, MaxCallDepth/MaxCompositionDepth sit at their ceiling — do not pattern-match one onto the other" point is restated near-verbatim in three separate doc comments (`interp.go`, `pathoracle.go`, `pathoracle_compose.go`). This is a maintainability cost: a future change to either constant's rationale requires updating prose in multiple files to stay consistent, and the doc-comment volume makes it easy for a reviewer to miss the one line that is actually load-bearing (the boundary check's placement relative to callee resolution).
**Fix:** Consider consolidating the shared "asymmetric-direction" rationale into one location referenced by the others (e.g. a single doc comment on `MaxCompositionDepth` that `MaxCallDepth`'s comment cross-references, rather than restating), leaving each site's own comment focused on what is unique to it.

---

_Reviewed: 2026-09-11T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

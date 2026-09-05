---
phase: 04-fallible-resources-and-c-boundary
reviewed: 2026-09-05T00:00:00Z
depth: standard
files_reviewed: 48
files_reviewed_list:
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/evidence/evidence.go
  - internal/compiler/evidence/evidence_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/native/foreign_resource.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_conformance_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/native/symbols.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - native/lang_foreign_nonlocal.c
  - native/lang_foreign_resource.c
  - native/lang_foreign_resource_private.h
  - scripts/verify-phase4.sh
  - testdata/phase4/acquire_three_fail_second.lang
  - testdata/phase4/acquire_three_fail_third.lang
  - testdata/phase4/acquire_three_success.lang
  - testdata/phase4/defect_terminal.lang
  - testdata/phase4/discard_because.lang
  - testdata/phase4/fallible_call_unconsumed.lang
  - testdata/phase4/foreign_acquire_one.lang
  - testdata/phase4/foreign_call_target_not_foreign.lang
  - testdata/phase4/foreign_layout_mismatch.golden.c
  - testdata/phase4/foreign_origin_omitted.lang
  - testdata/phase4/foreign_unwind_undeclared.lang
  - testdata/phase4/nonlocal_exit_probe.lang
findings:
  critical: 1
  warning: 1
  info: 1
  total: 3
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-09-05
**Depth:** standard
**Files Reviewed:** 48
**Status:** issues_found

## Summary

This pass re-reviews the full Phase 4 file set with closest attention on the two newest gap-closure plans: 04-08 (`corevalidate.go`'s per-incoming-edge release-order rederivation and the new `core.terminal_block_unreachable` structural peer check in `blocksAndEdges`, plus `corevalidate_test.go`'s falsifiers) and 04-09 (`session_test.go`'s conformance-layer injection falsifier and the scanned-artifact-count pin). Both of those changes are sound and well-tested: `TestMergeTerminalBlockDivergentReleaseSetsRefused` / `TestMergeTerminalBlockAgreeingChainsAccepted` correctly falsify and confirm the fixed per-edge rederivation from CR-01 of the prior review pass, `TestTerminalBlockUnreachableRefused` correctly proves the new structural peer check, and `TestAttributeScanLaneCoversEveryInspectableLayer` correctly pins the production lane's scanned-artifact count to 8 so a dropped argument turns the test red rather than silently degrading coverage. The four findings recorded as fixed in `04-REVIEW-FIX.md` (CR-01, WR-01, WR-02, WR-03) all hold up against the current code; none are re-reported here.

One new critical gap survived both this pass and the prior one: `corevalidate.checkReleaseOrder`'s own backward-walk helper (`rederive`, corevalidate.go:1236-1255) has no cycle protection, unlike every other backward/forward graph walk in the same file (`blockReach`, `loanChainIndex.carriedLoans`) which explicitly guard against a cyclic declared graph and document why they must. A hand-corrupted `core.Program` with a cyclic chain of `"ok"`-pattern edges among non-terminal blocks causes `rederive`'s `for { ... }` loop to run forever, hanging validation — precisely the class of adversarial input this package's own comments (e.g. the `carriedLoans` doc comment) say a source-blind validator "must stay defined against."

One warning and one info item round out the pass.

## Critical Issues

### CR-01: `corevalidate.checkReleaseOrder`'s backward rederivation walk has no cycle guard and can loop forever on a corrupted core.Program

**File:** `internal/compiler/corevalidate/corevalidate.go:1236-1255` (the loop is the `rederive` closure; the vulnerable step is line 1251, `currentBlockID = edge.FromBlockID`)
**Issue:**

```go
rederive := func(startEdge core.Edge) []core.LinearOperation {
    var expected []core.LinearOperation
    includeThis := startEdge.Pattern == "ok"
    currentBlockID := startEdge.FromBlockID
    for {
        v.checks++
        if includeThis {
            if op, ok := callInBlock[currentBlockID]; ok {
                expected = append(expected, op)
            }
        }
        edge, ok := okEdgeInto[currentBlockID]
        if !ok {
            break
        }
        currentBlockID = edge.FromBlockID
        includeThis = true
    }
    return expected
}
```

`okEdgeInto` is `map[string]core.Edge` keyed by `ToBlockID`, built directly from `linear.Edges` with no acyclicity requirement enforced anywhere earlier in `Validate`. `blocksAndEdges` (corevalidate.go:376-476) checks edge/block ID uniqueness and referential closure — every edge's endpoints resolve to a declared block — but never checks that the block graph is acyclic. If a hand-corrupted (or buggy-producer) `core.Program` declares two blocks A and B with `"ok"` edges A→(something)→...→B and B→...→A (i.e., `okEdgeInto[A]`'s `FromBlockID` chain eventually reaches B, and `okEdgeInto[B]`'s chain reaches back to A), then `rederive`'s `for { }` loop bounces between them forever: `ok` stays `true` on every iteration (a cyclic chain never returns `!ok` from the map lookup), so the loop never hits its only `break`.

This is the same shape of adversarial input the file explicitly defends against elsewhere with visited-sets — `loanChainIndex.carriedLoans`'s doc comment states plainly: "A self-referencing or cyclic parent pointer can only arise from a corrupted core artifact... but this helper must never crash on one: a recursive walk would recurse forever... which is precisely the kind of adversarial input a source-blind validator must stay defined against"; `blockReach`'s doc comment makes the identical argument for its own BFS ("a visited-once frontier, never re-enqueued, so the computation terminates in bounded time even over an (illegitimately) cyclic declared graph"). `checkReleaseOrder`'s `rederive` walk is the one backward-graph-traversal in this file that lacks the equivalent guard, and it is reachable from `Validate` on any function whose `Linear.Blocks`/`Linear.Edges` are populated (any branch-shaped function, via `replayBlocks` → `checkReleaseOrder`) — not gated behind any earlier acyclicity check. `corevalidate_test.go` has extensive merge-terminal-block falsifiers (`TestMergeTerminalBlockDivergentReleaseSetsRefused`, `TestMergeTerminalBlockAgreeingChainsAccepted`) added by the 04-08 gap-closure plan, but no falsifier constructs a cyclic `"ok"`-edge chain, so this gap is untested and unfixed by that plan.

Because `corevalidate.Validate` is documented as source-blind and independent-of-the-checker by design (its entire reason for existing is to stay defined against a corrupted or adversarially hand-constructed `core.Program`, not just one honestly produced by `check.go`), an infinite loop here is a real denial-of-service against any caller that runs untrusted or fuzzed core artifacts through validation — the compiler process simply hangs, consuming CPU forever, with no timeout or bound anywhere in the call chain.

**Fix:** Add a visited-set to `rederive`, the same shape `carriedLoans` and `blockReach` already use, and treat re-visiting a block during the backward walk as a hard refusal (a cyclic release-order chain is itself evidence of a corrupted core artifact, not a shape to silently truncate):

```go
rederive := func(startEdge core.Edge) ([]core.LinearOperation, bool) {
    var expected []core.LinearOperation
    includeThis := startEdge.Pattern == "ok"
    currentBlockID := startEdge.FromBlockID
    visited := make(map[string]bool)
    for {
        if visited[currentBlockID] {
            return nil, false // cyclic ok-edge chain: refuse, do not loop forever
        }
        visited[currentBlockID] = true
        v.checks++
        if includeThis {
            if op, ok := callInBlock[currentBlockID]; ok {
                expected = append(expected, op)
            }
        }
        edge, ok := okEdgeInto[currentBlockID]
        if !ok {
            break
        }
        currentBlockID = edge.FromBlockID
        includeThis = true
    }
    return expected, true
}
```

and propagate the `false` result to `v.check(false, "core.release_order_cyclic", block.ID)` at each of the two call sites in `checkReleaseOrder`'s main loop (corevalidate.go:1288-1298). Add a falsifier constructing a two-block `"ok"`-edge cycle feeding into a terminal block's incoming edge, asserting the validator returns (does not hang) and refuses with the new code.

## Warnings

### WR-01: `TestReleaseOrderValidationWorkSeries`'s monotonic-work assertion does not by itself prove linearity, and is loosely coupled to the new per-incoming-edge cost

**File:** `internal/compiler/corevalidate/corevalidate_test.go:703-731`
**Issue:** The test's own doc comment claims it "proves the validator's release-order rederivation cost is linear in the number of blocks plus edges plus operations," but the test body only asserts the three-point `series` is strictly increasing (`series[index] <= series[index-1]` fails the test). A strictly-increasing series is consistent with linear, quadratic, or any other superlinear growth in the fixture sizes used — it does not distinguish "linear" from "quadratic in the number of incoming edges per terminal block," which is exactly the shape the 04-08 fix (walking every incoming edge into a terminal block, corevalidate.go:1288) newly introduces. This isn't a functional defect, but the doc comment overstates what the assertion actually establishes, and a future accidental quadratic blowup in `checkReleaseOrder` (e.g. from a terminal block with many incoming edges each triggering a full backward walk) would not be caught by this test.
**Fix:** Either soften the doc comment to describe what is actually checked ("proves the counted work does not regress to a constant, i.e. it grows with fixture size") or strengthen the assertion to fit at least four points and check a ratio bound consistent with linear growth (e.g. `series[i+1]-series[i]` staying within a bounded multiple of the size delta).

## Info

### IN-01: `checkReleaseOrder`'s new per-incoming-edge loop recomputes `actual` once but calls `rederive` (a full backward graph walk) once per incoming edge, which is fine functionally but makes the big-O of a single terminal block's check proportional to `incoming-edge-count × chain-depth` rather than `chain-depth` alone

**File:** `internal/compiler/corevalidate/corevalidate.go:1288-1298`
**Issue:** This is a correctness-required cost (per the CR-01 fix rationale in `04-REVIEW-FIX.md`: every incoming edge must be independently walked and compared, since a merge point can only be safely accepted if every path agrees), not a defect — flagged only as a note for anyone reading the `LinearWorkLimit`/`TestReleaseOrderValidationWorkSeries` cost-accounting comments elsewhere in the file, since a terminal block with many incoming merge edges (not exercised by any current fixture) will visibly increase `v.checks` beyond what the existing scale-shape formula (`LinearWorkLimit`) accounts for. Performance is out of scope for this review; this is purely a note that the two cost-tracking mechanisms (`LinearWorkLimit`'s formula and `TestReleaseOrderValidationWorkSeries`'s monotonic check) do not currently model this dimension.
**Fix:** None required; consider a fixture with a terminal block reached by 3+ incoming edges if `LinearWorkLimit`'s formula is ever asserted against a shape like this in the future.

---

_Reviewed: 2026-09-05_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

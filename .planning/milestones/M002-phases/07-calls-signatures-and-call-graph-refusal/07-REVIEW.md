---
phase: 07-calls-signatures-and-call-graph-refusal
reviewed: 2026-09-09T00:00:00Z
depth: standard
files_reviewed: 22
files_reviewed_list:
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_call_argument_consume_internal_test.go
  - internal/compiler/corevalidate/corevalidate_foreign_closure_test.go
  - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/corevalidate/export_test.go
  - internal/compiler/originvalidate/export_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_foreign_closure_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_peer_gate_test.go
  - internal/compiler/session/session_phase7.go
  - internal/compiler/session/session_phase7_export_test.go
  - internal/compiler/session/session_phase7_test.go
  - scripts/verify-phase7.sh
  - testdata/phase07/call_argument_used_once.lang
  - testdata/phase07/call_argument_used_twice.lang
  - testdata/phase07/call_fallible_foreign_reach.lang
  - testdata/phase07/duplicate_function_name.lang
findings:
  critical: 0
  warning: 0
  info: 2
  total: 2
status: clean
---

# Phase 07: Code Review Report

**Reviewed:** 2026-09-09
**Depth:** standard
**Files Reviewed:** 22
**Status:** clean (Info-only findings remain; no Critical or Warning findings)

## Summary

This is a re-review (iteration 2) of the same 22-file scope reviewed in
iteration 1. Iteration 1 found 0 Critical, 1 Warning (WR-01), 2 Info
(IN-01, IN-02). Commit `3f274ad` applied WR-01; commit `01cbe49` recorded
the fix report. This review verifies the WR-01 fix directly against the
current code and re-checks whether IN-01/IN-02 (deliberately left
unfixed, Info severity, out of `critical_warning` fix scope) are still
present.

**WR-01 verification (`internal/compiler/corevalidate/corevalidate.go`).**
Read `consumeCallArgument` (now at lines 2252-2278) and both call sites
(`replayStraightLine:1593`, `replayBlocks:1845`) directly:

- The function signature grew an explicit `sourceTypeID string` parameter.
  Both call sites pass `source.TypeID` — the argument place's own
  declared type, taken from a `source` value each call site already holds
  in scope before the `OpCall` case runs — not a value re-derived from
  `operation.TypeID`.
- Inside the function, the new assertion `v.check(operation.TypeID ==
  sourceTypeID, "core.type_mismatch", operation.ID)` runs and returns
  `false` on failure **before** `v.derive(types[operation.TypeID].Shape,
  0)` executes (line 2267 precedes line 2270) — correctly placed ahead of
  the derivation it protects, not after.
- The assertion fails closed: on divergence, the function returns `false`
  without touching `initialized[operation.SourceID]`, propagating a
  refusal rather than silently admitting or silently consuming.
- The diagnostic code reused (`"core.type_mismatch"`) is the same code
  already used by the pre-existing generic `source.TypeID ==
  operation.TypeID` check at lines 1492/1751 for the identical semantic
  property (type identity mismatch) — this is a sound reuse, not a
  collision with an unrelated meaning.
- The new test,
  `TestConsumeCallArgumentRefusesWhenSourceTypeDivergesFromOperationType`
  (`corevalidate_call_argument_consume_internal_test.go:89-107`), drives
  `consumeCallArgument` directly with a deliberately diverging
  `sourceTypeID` (`type:1`) against `operation.TypeID` (`type:0`) and
  asserts both that the call returns `false` and that
  `initialized[sourceID]` is left untouched — confirming the refusal is
  genuine and fail-closed, not merely a returned boolean that the caller
  ignores.
- Ran the three `consumeCallArgument`-adjacent tests directly
  (`go test ./internal/compiler/corevalidate/... -run
  TestConsumeCallArgument -v`): all three pass, including the new one.
  `go build ./...` and `go vet ./...` are clean.
- Because both call sites already pass `source.TypeID` (the same value
  the generic pre-existing law already forces to equal
  `operation.TypeID` before the `OpCall` switch case runs), the new
  assertion is inert in production today — it cannot change any
  fixture's verdict. It exists purely as a locally-owned, test-caught
  tripwire for exactly the future-drift scenario WR-01 described. This
  is the intended shape of the fix and introduces no over-refusal risk:
  no previously-admitted program can newly fail this check, since the
  values compared are identical by construction at both call sites today.

WR-01 is genuinely resolved. No new Critical or Warning finding was
introduced by the fix.

**IN-01 and IN-02 (Info, unfixed by design)** — re-confirmed still present,
unchanged, exactly as expected since the fixer's stated scope
(`fix_scope: critical_warning`) excluded them:

- IN-01: `internal/compiler/originvalidate/export_test.go` still ends
  with a stray trailing blank line (confirmed via hex dump of file tail:
  `...0a0a` — a blank line after the final `}`).
- IN-02: the `peerJoinFails`/`joinFails` cross-package sort-order
  coincidence (`corevalidate.go` `checkCallGraphAcyclic`'s adjacency
  build vs. `originvalidate.go`'s `calleeIDsForClosureDigest`) is still
  undefended by any dedicated same-callee-set-different-discovery-order
  test for `Fails` specifically. The referenced code (now around
  `corevalidate.go:624-635`) is unchanged since iteration 1.

Re-reporting both below at Info severity is correct and expected per the
iteration instructions — this is not a regression, and with only Info
findings remaining, `status: clean` applies for the loop's convergence
check.

## Info

### IN-01: `originvalidate/export_test.go` still has a stray trailing blank line

**File:** `internal/compiler/originvalidate/export_test.go:57` (EOF)
**Issue:** The file ends with an extra blank line after the final `}`
with no accompanying code change to justify it. Harmless, but it's dead
diff noise gofmt/goimports would not introduce on its own.
**Fix:** Remove the trailing blank line.

### IN-02: `peerJoinFails`'s determinism still depends on an unenforced cross-package sort-order coincidence

**File:** `internal/compiler/corevalidate/corevalidate.go:624-635`,
`internal/compiler/originvalidate/originvalidate.go:702-722`
**Issue:** `joinFails`/`peerJoinFails` are explicitly not
order-independent (the accumulator always wins over a later disagreeing
value; disclosed as D-07-53 debt). Producer/peer parity for `Fails`
depends on both sides folding over their respective callee lists in the
same order, which today holds only because two *separately implemented*
sorts (`calleeIDsForClosureDigest` in originvalidate,
`checkCallGraphAcyclic`'s adjacency build in corevalidate) happen to
agree — documented in a comment, not asserted by a dedicated test for
`Fails` divergence under differing dedup/traversal order between the two
sides. If either sort were dropped independently on one side, producer
and peer would silently compute different `Fails` values with no
diagnostic.
**Fix:** Add a peer-side test analogous to
`TestForeignJoinIsOrderIndependent` that drives a multi-callee,
disagreeing-`Fails` shape through both `originvalidate.BuildInterface`
and `corevalidate.Validate`'s peer signatures and asserts the two `Fails`
values agree.

---

_Reviewed: 2026-09-09_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

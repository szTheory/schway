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
  warning: 1
  info: 2
  total: 3
status: issues_found
---

# Phase 07: Code Review Report (Gap Closure: plans 07-10, 07-11, 07-12)

**Reviewed:** 2026-09-09
**Depth:** standard
**Files Reviewed:** 22
**Status:** issues_found

## Summary

This review covers only `git diff 6706526..HEAD` for phase-07's gap-closure
plans: 07-10 (CR-04/PVG-03, `lang check`/`interface` consult the
corevalidate peer and report its refusal instead of discarding it),
07-11 (CR-01/PVG-01, a call now moves a non-copyable argument so a
double-consume is refused), and 07-12 (CR-03/PVG-02, `Foreign`/`Fails` are
now closure-derived worst-case joins across the call graph, independently
on both the producer and the peer side).

I traced each of the four correctness properties named in the review scope
against the actual code, not just the doc comments claiming them:

1. **Bilateral independence.** Verified genuinely separate implementations
   at every point that matters: `check.resolveCallBinding`'s consume rule
   reads its own `TypeFact.Abilities` (a value it computed itself via
   `ability.DeriveSealed`), while `corevalidate.consumeCallArgument` calls
   its own `v.derive` over `Shape` — a structural re-derivation, never a
   read of the recorded `Abilities` list (proven directly by
   `corevalidate_call_argument_consume_internal_test.go`, which hand-builds
   a `TypeFact` whose recorded `Abilities` disagrees with its `Shape` and
   shows the peer follows `Shape`). Likewise `originvalidate.joinForeignReach`
   /`joinFails` and `corevalidate.peerJoinForeignReach`/`peerJoinFails` are
   textually distinct functions, walking distinct traversal state
   (`callgraph.Order`'s reverse postorder vs. `checkCallGraphAcyclic`'s own
   `v.peerPostorder`/`v.peerAdjacency`), with a dedicated bilateral test
   (`TestForeignClosureJoinPeersIndependent`) that disables each side's join
   in turn and shows the other side still derives the correct answer alone.
2. **Refusing union, never intersection.** `session.CheckCommandFile`,
   `InterfaceExportCommandFile`, and `InterfaceCoreCommandFile` all consult
   `corevalidate.Validate` in a fixed, documented precedence (check's own
   diagnostics first, then the peer, then `originvalidate.ValidatePublished`)
   and report whichever fires first; none of them read anything back from
   the peer into an admission decision. `peerRefusalDiagnostic` is
   confirmed to be pure formatting of the peer's own already-computed
   verdict, not a second predicate.
3. **Join correctness (07-12).** `joinReachPolicy`/`joinAllocatorName` are
   commutative (both branches are symmetric in `into`/`from`) and
   associative once a conflict or "forbidden" value appears (both are
   absorbing under a further fold), and `BuildInterface` explicitly runs
   `callgraph.Order` before computing any `ClosureDigest`, aborting with no
   digest at all on a cyclic graph. `TestForeignJoinIsOrderIndependent`
   exercises a genuine three-way disagreement (conflicting allocator names,
   a forbidden-vs-permitted Unwind, and an absent-vs-forbidden
   NonlocalExit) in both call orders and gets the same answer both times.
4. **No over-refusal.** `call_argument_used_once.lang` and the Byte/Byte
   double-call path are pinned by dedicated tests
   (`TestCallDoesNotConsumeCopyableArgument`,
   `TestCallArgumentConsumeOverRefusalMutationKilled`,
   `TestPeerAdmitsRepeatedCopyableCallArgument`), and
   `TestCallArgumentConsumptionUnchangedAcrossAcceptingCorpus` sweeps the
   whole previously-accepting fixture list and asserts byte-identical,
   deterministic `core.Program` output — the consume rule is proven to
   change checker *state* only, never emitted core, for every fixture that
   used to check clean.

`go build ./...` succeeds and `go test ./internal/compiler/{check,corevalidate,originvalidate,session}/...` passes in full (224s for the session package, which drives the native/CLI corpus). I did not find any BLOCKER-level defect in the reviewed diff. The three findings below are quality/robustness observations, not correctness failures.

## Warnings

### WR-01: `consumeCallArgument`'s ability re-derivation is keyed off the callee's return type, not the argument's own declared type, and this equivalence is not defended by any operation-kind-specific check

**File:** `internal/compiler/corevalidate/corevalidate.go:2234-2246` (also `internal/compiler/corevalidate/corevalidate.go:1593`, `:1845`)
**Issue:** `consumeCallArgument` derives the copy-ability decision from
`types[operation.TypeID].Shape` — but for an `OpCall`, `operation.TypeID` is
the call's *target* type (derived, on the check-production side, from the
**callee's declared return type**; see `check.go:1940` `derivedTypeID =
typeFact.ID`), not from the argument (`operation.SourceID`)'s own type. The
function's own doc comment argues this is safe because the generic
`source.TypeID == operation.TypeID` check (run for every operation kind,
`corevalidate.go:1492`/`:1751`) already forces the two to agree — but that
generic check is not specific to `OpCall`; it is the same law used for
`OpCopy`/`OpMove`/`OpBorrow*`, where `TargetID`'s type and `SourceID`'s type
are definitionally the same value. For `OpCall` this equivalence is not an
IR invariant of the operation itself, it is a *derived consequence* of two
other, separately-checked facts: `check.go`'s D-07-09 constraint that every
function's declared return type equals its own declared parameter type
(`sameType`, enforced once per function at admission time — `check.go:170,
1575, 1826`), plus `checkCallTypeContract`'s own argument/return contract
check, which must run and succeed *before* `consumeCallArgument` is called
(both replay sites do call it first). If a future change ever reordered
`checkCallTypeContract` and `consumeCallArgument`, relaxed the
callee-return-equals-callee-parameter language constraint, or allowed
`operation.TypeID` to diverge from `source.TypeID` for `OpCall` in some new
way, `consumeCallArgument` would silently decide copyability from the wrong
type with no local invariant to catch it — the generic
`source.TypeID==operation.TypeID` check only accidentally happens to make
this safe today, for `OpCall` specifically, because of a cross-function
language-surface constraint elsewhere in the codebase, not because of
anything local to this operation kind.
**Fix:** Either derive from `types[source.TypeID].Shape` directly (the
argument's own type, matching what the doc comment claims is being read),
or add an explicit assertion inside `consumeCallArgument` (or immediately
before its call site) that `operation.TypeID == source.TypeID` for
`OpCall` specifically, with a comment tying that assertion to the D-07-09
constraint it depends on — so a future change to either constraint fails a
test here instead of silently changing which type's abilities decide
consumption.

## Info

### IN-01: `originvalidate/export_test.go` diff adds a stray trailing blank line

**File:** `internal/compiler/originvalidate/export_test.go:57`
**Issue:** The gap-closure diff appends a trailing empty line at end of
file with no accompanying code change. Harmless, but it's dead diff noise
that gofmt/goimports would not normally introduce on its own.
**Fix:** Remove the trailing blank line, or fold it into a future edit to
this file.

### IN-02: `peerJoinFails`'s determinism depends on an unenforced cross-package sort-order coincidence between `corevalidate`'s adjacency build and `originvalidate`'s `calleeIDsForClosureDigest`

**File:** `internal/compiler/corevalidate/corevalidate.go:620-635`,
`internal/compiler/originvalidate/originvalidate.go:702-722`
**Issue:** `joinFails`/`peerJoinFails` are explicitly *not*
order-independent (`into` — the accumulator, i.e. whichever value was
folded first — always wins over a later disagreeing `from`; this is
disclosed as D-07-53 debt). Producer/peer parity for `Fails` therefore
depends entirely on both sides folding over their respective callee lists
in the *same* order. The code achieves this today because both
`calleeIDsForClosureDigest` (originvalidate) and `checkCallGraphAcyclic`'s
adjacency build (corevalidate) independently sort each function's own
callee-ID list before it is folded — but this is two *separately
implemented* sorts that happen to agree, documented only in a comment
(`corevalidate.go:624-629`) rather than asserted by any test that
specifically targets `Fails` divergence under differing dedup/traversal
order between the two sides (the existing
`TestForeignClosureJoinPeersIndependent`/mutation-matrix tests exercise
seam-disable divergence, not a same-callee-set-different-discovery-order
divergence for `Fails` specifically, the way `TestForeignJoinIsOrderIndependent`
does for the (order-independent) `Foreign` join). If either sort were ever
dropped independently on one side only, producer and peer would silently
compute *different* `Fails` values for the same program without either
side refusing — a real divergence with no diagnostic, since both sides
would still independently succeed, just disagree.
**Fix:** Add a peer-side test analogous to
`TestSecondPassResolvesProgramFunctionByID`/`TestForeignJoinIsOrderIndependent`
that drives the same multi-callee-with-disagreeing-`Fails` shape through
both `originvalidate.BuildInterface` and `corevalidate.Validate`'s peer
signatures and asserts the two `Fails` values agree — closing the gap
between "both sides sort, by inspection" and "both sides are proven to
agree even if one sort implementation regresses."

---

_Reviewed: 2026-09-09_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

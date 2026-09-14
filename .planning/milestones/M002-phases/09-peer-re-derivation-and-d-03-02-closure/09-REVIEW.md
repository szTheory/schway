---
phase: 09-peer-re-derivation-and-d-03-02-closure
reviewed: 2026-09-11T03:03:57Z
depth: deep
files_reviewed: 31
files_reviewed_list:
  - internal/compiler/check/check.go
  - internal/compiler/check/check_call_transfer_agreement_test.go
  - internal/compiler/check/check_disclosure_peer_test.go
  - internal/compiler/check/check_exclusive_test.go
  - internal/compiler/check/check_ordering_stability_test.go
  - internal/compiler/check/check_peer_liveness_seam_test.go
  - internal/compiler/check/check_peer_shape_differential_test.go
  - internal/compiler/check/check_shadow_subsumption_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/check/costcorpus_relocation_test.go
  - internal/compiler/check/costcorpus_test.go
  - internal/compiler/core/core_convention_absence_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go
  - internal/compiler/corevalidate/corevalidate_cycle_peer_test.go
  - internal/compiler/corevalidate/corevalidate_disclosure_test.go
  - internal/compiler/corevalidate/corevalidate_exclusive_test.go
  - internal/compiler/corevalidate/corevalidate_peer_cost_test.go
  - internal/compiler/corevalidate/corevalidate_peer_liveness.go
  - internal/compiler/corevalidate/corevalidate_peer_liveness_test.go
  - internal/compiler/corevalidate/corevalidate_peer_origin_test.go
  - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/corevalidate/export_test.go
  - internal/compiler/measure/statistics.go
  - internal/compiler/session/session_peer_gate_test.go
  - internal/compiler/session/session_phase6_budget.go
  - internal/compiler/session/session_phase6_budget_test.go
  - internal/compiler/session/session_phase6_risklanes.go
  - internal/compiler/syntax/syntax_convention_override_test.go
  - internal/compiler/testsupport/callgraphcorpus.go
findings:
  critical: 0
  warning: 1
  info: 2
  total: 3
status: issues_found
---

# Phase 09: Code Review Report

**Reviewed:** 2026-09-11T03:03:57Z
**Depth:** deep
**Files Reviewed:** 31
**Status:** issues_found

## Summary

Phase 09 lands a genuinely independent second admission layer inside
`corevalidate` and closes D-03-02 in `check`. I read every changed production
`.go` file in full, cross-referenced them against the ten `09-*-SUMMARY.md`
files and `PHASE-09-DEBT.md`'s 16 items, ran `go build ./...` and `go vet
./internal/compiler/...` (both clean), and ran the four directly affected
packages' test suites (`check`, `corevalidate`, `session`, `measure` — all
pass). I traced call chains across the `check`/`corevalidate`/`testsupport`
boundary and diffed the restructured `check.go` loan-liveness machinery
against the pre-Phase-09 tree line by line for the two codes D-09-47 named as
timing-dependent (`ownership.move_while_borrowed`, `ownership.borrow_conflict`).

**Verdict on each review priority, stated explicitly per the instructions:**

1. **Peer independence.** Clean. Production `corevalidate.go` and the new
   `corevalidate_peer_liveness.go` import only `core` (verified by direct
   inspection of both files' import blocks, not just by trusting the doc
   comments). `derivePeerLoanCarry` is a genuinely different algorithm from
   `check`'s `deriveFunctionUsesParam`/`buildInterproceduralSummaries`: a
   single forward pass over a callee-before-caller postorder with no
   fixpoint iteration, versus `check`'s backward-walked, memoized summary
   table over `callgraph.Order` (caller-before-callee, walked backward). The
   two do not share a helper, a data structure, or a traversal order. I
   found no drift toward "one algorithm wearing two names."
2. **Decorative gates.** Clean, and unusually well-disclosed. The synthetic
   call-graph corpus differential (`check_peer_shape_differential_test.go`)
   explicitly documents that its five shapes contain **zero** borrow
   operations (verified by inspecting `testsupport/callgraphcorpus.go`'s
   three builder templates — only `OpCall`/`OpCopy`/`OpReturn` are ever
   emitted) and is therefore vacuously true on its own; it then supplies a
   real mutation-kill test
   (`TestSyntheticShapeDifferentialMutationReintroducesDivergence`) that
   hand-builds one fixture with a real borrow and proves the differential's
   detection machinery actually fires, using
   `corevalidate.SetForcePeerLoanCarryTrueForTest` as the fault, with an
   explicit re-disengage-and-recheck-clean step. The bidirectional seam
   pair required by D-09-24/D-09-25
   (`TestPeerLoanCarrySeamDisabledCheckStillRefuses` /
   `TestInterproceduralLivenessSeamCheckDisabledCorevalidateStillRefuses`)
   is present in both directions and each asserts the full
   disagree-with-seam / agree-without-seam / unrelated-fact-unaffected
   triple, not merely "output changed." I did not find a test that would
   pass with its target mechanism deleted.
3. **The deletion's blast radius.** Clean. `computeLoanLastUses` and its
   `shadow:place:*` scaffolding are gone
   (`check.go` no longer defines them; `analyzeStraightLine`/
   `analyzeArmBody`'s `activeLoans`/`expiringLoans` are explicitly
   relabeled "evidence only" and no longer gate). `checkInterproceduralLoanLiveness`
   is confirmed as the sole decision point for both timing-dependent codes,
   with cause lists (`borrow_created_here`, `borrow_used_later`, `loan`,
   `owner`, `type` for `move_while_borrowed`; the same five for
   `borrow_conflict`) reproduced byte-for-byte against the pre-deletion tree
   (diffed directly, see below), so the `Code + Span + Causes` identity is
   preserved for these two codes even though the *decision timing* moved.
   `testOnlyForceUniformLoanJoin` is confirmed still consumed at
   `check.go:887`, gated to `function.Match != nil` — not orphaned.
   `LoanFinalUses` is confirmed still populated from `loanFinalUseEvidence`
   independently of the deleted timing index. The genuinely new fact this
   review surfaced — that Pattern B's twin split does not materialize
   because of a separate, pre-existing `deriveFunctionUsesParam` defect —
   is already fully disclosed as D-09-53 in `PHASE-09-DEBT.md` and is not
   re-reported here as new.
4. **Iterative, never recursive.** Clean. Every new function introduced this
   phase (`derivePeerLoanCarry`, `chainPeerLoanCarry`, `peerDeriveOriginFacts`,
   `peerOriginContained`, `peerForeignOriginOmitted`, `buildLoanChainIndex`'s
   extended form, `checkInterproceduralLoanLiveness`'s extension) is a single
   forward or backward scan over a slice, never self-referential. The
   existing `carriedLoans` iterative walk (with cycle-detection via a
   `visited` set, never recursion) is untouched.
   `TestPeerLoanCarryTerminatesOnCorruptedCoreArtifact` backs this with a
   real 5-second-timeout goroutine race against a self-referencing
   `core.Program`, not merely an assertion of memo-map non-recursion.
5. **Diagnostic identity.** Clean for every code this phase actually
   changed the decision timing of, with one stale doc-comment exception
   (see WR-01 below) that is a documentation defect, not an identity
   defect: `check.interprocedural_loan_liveness`'s builder function
   (`interproceduralLoanLivenessDiagnostic`) is byte-identical to the
   pre-Phase-09 tree. `check_ordering_stability_test.go` commits a literal
   pre-restructure baseline and re-walks the corpus, correctly classifying
   each observed identity/span change (and honestly reporting the Pattern B
   non-split as a diagnosed pre-existing defect rather than silently
   updating the table).
6. **Ordinary correctness.** No off-by-one, nil-deref, map-mutation-during-
   iteration, unhandled-error, or race issues found in the new production
   code. `go build ./...` and `go vet ./internal/compiler/...` are both
   clean; `go test` on `check`, `corevalidate`, `session`, and `measure`
   all pass.

One warning (a stale doc comment contradicting a locked decision) and two
info items are recorded below.

## Warnings

### WR-01: Stale doc comment still describes the superseded core.* promotion plan

**File:** `internal/compiler/check/check.go:820-829`
**Issue:** `checkInterproceduralLoanLiveness`'s doc comment (fact 4 of its
four-fact final-scope list) still reads:

```go
//  4. This code lives in the check.* namespace, not core.*. It is scheduled
//     for promotion to core.* in Phase 09 at the moment corevalidate
//     independently re-derives the same interprocedural fact (D-08-21) --
//     promoting it earlier, before a second derivation exists to validate
//     against, would ship an unvalidated single-source-of-truth move.
```

D-09-31 formally supersedes this exact commitment in writing ("Locked:
`check.interprocedural_loan_liveness` stays `check.*`. ... Neither is
renamed, aliased, or merged"), and `PHASE-09-DEBT.md` D-09-31 records the
supersession — but the source comment this decision was originally written
against was never updated to match. A reader who opens `check.go` directly
(rather than `09-CONTEXT.md` or `PHASE-09-DEBT.md` first) will read a
currently-false claim: that promotion is still scheduled and merely
awaiting `corevalidate`'s re-derivation, which as of this phase now exists.
This is exactly the "no future reader has to discover the change" failure
mode `09-CONTEXT.md` explicitly calls out for D-09-08 and D-09-31 elsewhere,
just not closed here. It carries no correctness risk (the code's actual
behavior — staying `check.*` — is correct), but it is a real, silently
misleading planning-document/source-of-truth mismatch, and it sits directly
inside the same doc comment block Phase 09 itself extended (facts 1-3, added
just above it, are current).
**Fix:**
```go
//  4. This code lives in the check.* namespace, not core.*, and stays there
//     permanently (D-09-31, superseding D-08-21's promotion commitment):
//     corevalidate's own independent liveness derivation
//     (corevalidate_peer_liveness.go) computes liveness through a
//     reachability closure, not a worklist, and therefore cannot fail the
//     same way this fixpoint does -- a shared core.* code would assert an
//     agreement the two mechanisms are structurally incapable of having.
//     corevalidate keeps core.move_while_borrowed with its own causes;
//     neither code is renamed, aliased, or merged.
```

## Info

### IN-01: `paramTrace`/`derived` maps in `derivePeerLoanCarry` never clear a place once set

**File:** `internal/compiler/corevalidate/corevalidate_peer_liveness.go:74-90`
**Issue:** `paramTrace[operation.TargetID] = true` and
`derived[operation.TargetID] = true` are only ever set, never cleared, for
the lifetime of one function's operation scan. This means if a place ID were
ever reused as the `TargetID` of two different operations within the same
function body (once as parameter-derived, later reassigned to something
that is not), the second write's absence of `paramTrace`/`derived` update
would leave the first write's `true` value in place — a "still tracked
after being reassigned to something else" false positive. This mirrors the
pre-existing, already-shipped pattern in `peerParameterEscapesOwned`
(`corevalidate.go:2210-2229`) and `peerDeriveOriginFacts`
(`corevalidate.go:2293-2336`, itself extended this phase), so it is not a
new invariant violation introduced by this file — core's SSA-like linear
place-ID scheme (each operation produces a fresh `TargetID`) is relied on
project-wide to make place-ID reuse within one function body impossible.
No test in this phase (or prior phases) constructs a core.Program that
violates that invariant to confirm `buildLoanChainIndex`'s consumers fail
closed rather than silently misattribute a loan if it ever were violated by
a corrupted artifact — a corrupted-artifact-with-reused-place-IDs case is a
narrower gap than the self-referencing-cycle case
`TestPeerLoanCarryTerminatesOnCorruptedCoreArtifact` already covers.
**Fix:** No action required unless the place-ID-uniqueness invariant is
ever relaxed elsewhere; if it is, add a targeted corrupted-artifact test
(mirroring the existing cycle test) asserting `derivePeerLoanCarry` fails
closed on place-ID reuse within one function, the same way it already fails
closed on an unresolvable `CalleeID`.

### IN-02: `lane:peer-closure-cost-scaling` is registered but not wired to any shipped verification call site

**File:** `internal/compiler/session/session_phase6_risklanes.go:290-304`
**Issue:** `liveLanesPhase9`'s own doc comment states plainly: "Not yet
wired to a shipped VerifyXxx addLane call site ... registered here solely
so risk_lanes.json's row for it satisfies TestRiskLaneRegistryLanesAreLive."
The peer's own `peer_closure_recomputed_work_growth_exponent` "hard" bound
(ratified in `qlt02_budget_manifest.json` at 1300 milliexponent) is
therefore enforced only by `corevalidate_peer_cost_test.go`'s own
`TestPeerClosureCostGrowthExponentWithinBound` test assertion at `go test`
time, never by a `session`-level gated lane invocation the way
`recomputed_work`'s bound is consulted through `EvaluateBudget` at
`session_phase6_verify.go:502`. This mirrors an already-existing,
already-disclosed precedent (`liveLanesPhase8`'s
`lane:interprocedural-cost-scaling` has the identical "not yet wired"
disposition, per that lane's own doc comment and `risk_lanes.json`'s
stated rationale for both), so this is not a new gap Phase 09 introduced —
it is Phase 09 following an established, deliberately-scoped pattern
exactly. Flagging it at Info rather than omitting it, per the instruction
to state explicitly rather than silently pass over a priority-adjacent
observation: the manifest row's "hard" `gate_type` could read as stronger
than what actually runs in the default CI/edit-loop path.
**Fix:** No action required for Phase 09 specifically (the pattern is
inherited and disclosed); if a future phase wires `lane:interprocedural-cost-scaling`
into a shipped gate, `lane:peer-closure-cost-scaling` should be wired
alongside it in the same change, since both bounds now share the identical
"declared hard, enforced only by direct test invocation" disposition.

---

_Reviewed: 2026-09-11T03:03:57Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: deep_

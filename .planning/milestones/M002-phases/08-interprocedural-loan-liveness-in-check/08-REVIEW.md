---
phase: 08-interprocedural-loan-liveness-in-check
reviewed: 2026-09-09T00:00:00Z
depth: standard
files_reviewed: 22
files_reviewed_list:
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/check/costcorpus_test.go
  - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
  - internal/compiler/measure/statistics.go
  - internal/compiler/measure/statistics_test.go
  - internal/compiler/session/qlt02_budget_manifest.json
  - internal/compiler/session/risk_lanes.json
  - internal/compiler/session/session_peer_gate_test.go
  - internal/compiler/session/session_phase6_budget.go
  - internal/compiler/session/session_phase6_budget_test.go
  - internal/compiler/session/session_phase6_risklanes.go
  - testdata/phase07/relay_escort_witness.lang
  - testdata/phase08/match_arm_call.lang
  - testdata/phase08/negative_control_fails.lang
  - testdata/phase08/negative_control_infallible.lang
  - testdata/phase08/relay_depth2_accept.lang
  - testdata/phase08/relay_depth2_refuse.lang
  - testdata/phase08/twin_a_accept.lang
  - testdata/phase08/twin_a_refuse.lang
  - testdata/phase08/twin_b_accept.lang
  - testdata/phase08/twin_b_refuse.lang
findings:
  critical: 1
  warning: 2
  info: 2
  total: 5
status: issues_found
---

# Phase 08: Code Review Report

**Reviewed:** 2026-09-09
**Depth:** standard
**Files Reviewed:** 22
**Status:** issues_found

## Summary

This phase adds interprocedural loan-liveness to `check`: a per-function summary table
(`buildInterproceduralSummaries`), a forward loan-canonicalization pass that propagates
loans across a callee's declared borrow-return contract (`derivePlaceLoans`), a backward
summary-aware liveness fixpoint (`blockLoanLiveness`/`loanLivenessFixpoint`), and the new
`check.interprocedural_loan_liveness` diagnostic. The design is heavily and precisely
documented, and the accompanying test corpus (twin pairs, negative controls, a
growth-exponent cost gate) is unusually thorough for the mechanism's actual surface area.

One genuine correctness defect was found: the new diagnostic-selection loop in
`checkInterproceduralLoanLiveness` iterates a Go map to pick which conflicting loan to
blame, without the deterministic tie-break the pre-existing `conflictingLoan` helper uses
for the exact same class of ambiguity (multiple concurrent shared borrows of the same
owner place). This directly contradicts the function's own doc comment claim of
"deterministic output" and this project's repeatedly-stated house style of never letting
Go map iteration leak into diagnostic content. Everything else reviewed — the bound
formula, the gate-eligibility chokepoints, `EvaluateBudget`'s verdict logic, and the
growth-exponent fit/stability tripwire — held up under scrutiny; only minor naming/quality
issues were found there.

## Critical Issues

### CR-01: Nondeterministic loan selection in the interprocedural liveness diagnostic

**File:** `internal/compiler/check/check.go:869-887`
**Issue:**

`checkInterproceduralLoanLiveness` picks which loan a conflicting `OpMove` is blamed for
by iterating `chain.borrowOperation`, a `map[string]core.LinearOperation`:

```go
for index, operation := range function.Linear.Operations {
    if operation.Kind != core.OpMove {
        continue
    }
    for loanID, borrow := range chain.borrowOperation {
        if borrow.SourceID != operation.SourceID {
            continue
        }
        lastUse, ok := lastUseIndexByLoan[loanID]
        if !ok || index >= lastUse {
            continue
        }
        if offendingIndex == -1 || index < offendingIndex {
            offendingIndex = index
            offendingMove = operation
            offendingLoanID = loanID
        }
    }
}
```

The outer loop deterministically picks the earliest offending `OpMove` by operation
index (as the function's own doc comment claims), but the inner loop's tie-break is
`index < offendingIndex` — a *strict* inequality. When **two or more distinct loans**
share the same `borrow.SourceID` (the same owner place) and both are still live at the
same offending `OpMove`, their `index` values are identical, so neither loan's
`offendingIndex` check ever wins over the other by index — whichever loan Go's map
iteration visits *first* is the one silently kept, and Go map iteration order is
randomized per run.

This is not a hypothetical: the codebase's own `conflictingLoan` helper (check.go:3524)
explicitly documents that **multiple concurrent shared borrows of the same owner place
are legal** ("Shared-plus-shared overlap is never a conflict") and, because of exactly
this ambiguity, deliberately sorts candidate loan IDs before picking one — "the same
deterministic-first-by-sorted-ID selection the existing move_while_borrowed gate already
uses (so two runs never disagree on which loan a diagnostic blames)". The new
interprocedural pass reintroduces the exact hazard that helper was written to avoid: two
`borrow buffer` (shared) loans both extended past the same later `take buffer` by the
same or different calls will cause the diagnostic's `borrow_created_here` span (cause 1)
and possibly `loan_extended_by_call` callee (cause 2, via a different `extendedByCall`
entry per loan) to vary nondeterministically between two runs of `lang check` over the
byte-identical source file.

This directly contradicts:
- The function's own doc comment: "choosing the earliest offending OpMove by operation
  index (deterministic output)" (check.go:782).
- The project's own stated house style throughout this phase (e.g. the sorted-name hash
  in `laneInputHash`, the sorted-ID tie-break in `conflictingLoan`, the explicit ordering
  discipline documented for `buildInterproceduralSummaries`) of never letting Go map
  iteration order leak into observable output.

**Fix:** Sort candidate loan IDs (or track the best-so-far with a deterministic
secondary key, e.g. lexicographically smallest `loanID`) before selecting among ties at
the same `offendingIndex`, mirroring `conflictingLoan`'s own pattern:

```go
for index, operation := range function.Linear.Operations {
    if operation.Kind != core.OpMove {
        continue
    }
    var candidateLoanIDs []string
    for loanID, borrow := range chain.borrowOperation {
        if borrow.SourceID != operation.SourceID {
            continue
        }
        lastUse, ok := lastUseIndexByLoan[loanID]
        if !ok || index >= lastUse {
            continue
        }
        candidateLoanIDs = append(candidateLoanIDs, loanID)
    }
    if len(candidateLoanIDs) == 0 {
        continue
    }
    sort.Strings(candidateLoanIDs)
    if offendingIndex == -1 || index < offendingIndex {
        offendingIndex = index
        offendingMove = operation
        offendingLoanID = candidateLoanIDs[0]
    }
}
```

A regression test exercising two concurrent shared loans on the same owner place, both
extended past the same move (one via a call, or both directly), run under `go test
-count=10` (or with `GODEBUG=` map randomization forced), would catch this; none of the
existing twin-pair fixtures construct two simultaneously-live loans on the same owner.

## Warnings

### WR-01: Stale test name after gate-eligible metric set was widened

**File:** `internal/compiler/session/session_phase6_budget_test.go:302`
**Issue:** `TestRecomputedWorkIsTheOnlyHardGate` was accurate before this phase, when
`recomputed_work` was the sole gate-eligible metric. Phase 08 Plan 05 (D-08-31/D-08-32)
widened `QLT02GateEligibleMetrics()`/`measure.GateEligibleMetrics()` to a two-element set
that also includes `recomputed_work_growth_exponent` (both files under review document
this widening explicitly), and `qlt02_budget_manifest.json` now carries a second `hard`
row for that metric. The test itself only exercises `recomputed_work`'s own
below/exact/above-ceiling behavior and does not assert exclusivity, so it still passes,
but its name now asserts something the codebase no longer claims to be true, which will
mislead a future reader auditing "which metrics can block".
**Fix:** Rename to something like `TestRecomputedWorkCeilingIsStrictlyEnforced`, and
leave the "only hard gate" claim to `TestGateEligibleMetricSetsAgreeAcrossChokepoints`
and `TestQLT02InterproceduralGrowthExponent`, which already test the actual (now
two-element) eligible set.

### WR-02: `derivePlaceLoans` recomputed three times per function in the same pass

**File:** `internal/compiler/check/check.go:834-864`, `1899`, `1991`
**Issue:** `checkInterproceduralLoanLiveness` calls `cfgBlocksForFunction` once, then
`loanLivenessFixpoint` (which internally calls `derivePlaceLoans` at line 1899),
`materializeLoanEndpoints` (which calls `derivePlaceLoans` again at line 1991), and then
calls `derivePlaceLoans` a third time directly at line 864 — all three over the identical
`function.Linear.Operations`/`summaries` pair. Per-call this is not asymptotically
expensive (it is one linear scan each), and the phase context explicitly puts pure
performance concerns out of scope, but three redundant re-derivations of the same
immutable value inside one function body is a maintainability smell: a future edit to
`derivePlaceLoans` that isn't purely referentially transparent (e.g. one that reads from
mutable state, or that this reviewer's CR-01 fix above turns into a sort-dependent
operation) risks the three call sites silently drifting out of sync with each other.
**Fix:** Compute `chain := derivePlaceLoans(...)` once in `checkInterproceduralLoanLiveness`
and thread it into `loanLivenessFixpoint`/`materializeLoanEndpoints` as a parameter,
rather than having each of the three sites re-derive it independently from the same
inputs.

## Info

### IN-01: `computeLoanLastUses`' shadow OpCall path silently assumes call arity 1

**File:** `internal/compiler/check/check.go:3735-3752`
**Issue:** The shadow-operation builder resolves a call binding's `sourcePlaceID` only
when `len(binding.RHS.Arguments) == 1`; for any other arity it falls through to the
`shadow:place:unknown:%d` sentinel, silently treating the call as if it referenced
nothing. This mirrors the real admission path's own arity-1 assumption
(`resolveCallBinding`'s `check.call_arity_unsupported` guard, per the comment), so it is
not reachable for any admitted program today, but the comment could more explicitly
state that this is fail-*safe* (the loan simply isn't tracked as used through the call,
which is the conservative direction) rather than leaving a reader to infer that from the
surrounding prose.
**Fix:** No code change required; consider a one-line comment addition noting the
fail-safe direction explicitly, for a future reader who has not internalized the
arity-1 admission invariant from the rest of the file.

### IN-02: `interproceduralSummary.usesParam`/`returnsBorrowOfParam` documented as
"exactly two consulted fields" but the summary struct actually carries three

**File:** `internal/compiler/check/check.go:499-524`, `793-799`
**Issue:** The doc comment on `checkInterproceduralLoanLiveness` (D-08-Task-2's "final
scope" block) states the law "consults EXACTLY two callee-signature bits":
`ReturnsBorrowOfParam` (`return.mode`) and `UsesParam` (`parameters[0].mode`) — but
`UsesParam` is not itself a callee-signature bit; it is a check-derived fact
(`deriveFunctionUsesParam`), and the actual third struct field, `parameterMode`, is
consulted purely for diagnostic *Detail* text, never for the liveness verdict itself.
This is a documentation-precision nit, not a functional gap (the disclosed-field-set
test, `TestInterproceduralDisclosedFieldSet`, is presumably what actually enforces the
closed set, and was not itself flagged as wrong here), but the prose conflates "consulted
for the verdict" with "consulted for the diagnostic's rendered text", which could confuse
a future auditor trying to verify T-08-04's body-blindness claim by reading the comment
alone rather than the code.
**Fix:** Clarify the comment to distinguish the two verdict-determining bits
(`returnsBorrowOfParam`, `usesParam`) from the diagnostic-only `parameterMode`/`returnMode`
detail strings.

---

_Reviewed: 2026-09-09_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

---
phase: 03-borrowed-views-and-cfg-lifetimes
recorded: 2026-09-04
code_head: 68ed4f7
status: accepted
disposition: carried-from-mid-phase-gate
items: 2
blocking: 0
---

# Phase 03: Mid-Phase Gate Debt Register

Recorded by 03-05's mandatory mid-phase gate (ROADMAP §Phase 3, "Decision,
2026-09-04"), adjudicating the two open items the phase handoff named. Neither
item is a correctness defect a program can observe executing today — both are
named here so 03-07 and the phase verifier see them before the phase closes,
per the gate's own instruction to record findings as dated debt rather than
let them pass unexamined.

## Items

| ID | Source | Threat/Req | Severity | Item |
|---|---|---|---|---|
| D-03-01 | 03-05 mid-phase gate, open item 1 | D-05/D-02-03 | warning | `loanLivenessFixpoint`'s linear cost never reaches the admission-deciding code path; `discoverLoanLastUses`' own quadratic propagation is still what accepts/rejects every program, and its work is still uncounted |
| D-03-02 | 03-05 mid-phase gate, open item 2 | OWN-04 | warning | A borrow-derived return with no declared `borrow(path)` origin exports a `FunctionSignature` indistinguishable from a fully-owned return |

## Detail

### D-03-01 — the checker's two liveness derivations decide different things, and the quadratic one still governs admission

**Finding.** Two liveness derivations coexist in `internal/compiler/check/check.go`, but they are not alternative answers to the same question — they answer *different* questions, and only one of them is load-bearing for accept/reject:

- `discoverLoanLastUses` (unchanged since Phase 2, transitive per-binding propagation, O(N²) in the worst case per D-02-03) is read by **both** `analyzeStraightLine` (`checkLinear`, line 1019) and `analyzeArmBody` (`checkBranch`'s per-arm helper, line 700). Its `loanUses[index].index` value feeds `loan.lastUse` directly, which drives `conflictingLoan`/`expiringLoans` — the actual accept/reject decision and the actual expiry timing, in **every** function this checker admits, straight-line or branch.
- `loanLivenessFixpoint`/`materializeLoanEndpoints` (03-03's new backward worklist dataflow) is called **only** inside `checkBranch`, **after** `analyzeArmBody` has already returned an admission verdict (`check.go:341-352`). Its sole output, `linear.LoanEndpoints`, is never read by any admission-deciding code — not by `conflictingLoan`, not by `expiringLoans`, not by anything in `analyzeArmBody` itself. It is a purely observational fact: the record 03-04's `corevalidate.recomputeLoanEndpoints` and this plan's `pathoracle` independently re-derive and compare against.

So "the checker's two derivations" do not disagree on any verdict, because only one of them ever produces a verdict. `discoverLoanLastUses` is single-sourced for admission across the whole checker. Two independent lines of evidence confirm it is *correct* for that role on every currently-reachable shape:

1. `TestOwnershipSequenceExhaustive` (extended this plan from 36 to 48 symbols to add the exclusive-borrow spelling — ~113,164 cases) proves `analyzeStraightLine`'s straight-line verdicts agree with an independently-implemented oracle (`oracleStraightLine`, materialized-edges-plus-fixed-point-closure — a different mechanism, not the same law restated) across the full alphabet, now including the five-row shared/exclusive conflict matrix.
2. `TestBranchSequenceExhaustive` (new this plan) proves `checkBranch`'s per-arm verdicts, produced through the real `Program(...)` entry point over 2,401 synthetic two-block programs, agree with the same `oracleStraightLine` differential applied independently per arm — legitimate because 03-01's per-arm aliasing makes each arm's admission decision structurally independent of its sibling (03-03-SUMMARY.md's own documented finding), so there is no cross-arm interaction `discoverLoanLastUses` could get wrong that this test would miss.
3. A straight-line function's CFG is a single block by construction (no branch exists), so "edge-specific last use" is vacuous for it — there is no divergent edge to place an endpoint on, and `discoverLoanLastUses`'s transitive scan reduces mathematically to the same live-set computation `loanLivenessFixpoint` performs on a degenerate one-block CFG (proven directly by 03-03's own `TestStraightLineEndpointsUnchanged`, which runs `loanLivenessFixpoint` on a single synthetic block wrapping a straight-line body's real operations and confirms it reproduces `discoverLoanLastUses`'s answer exactly).

**So OWN-03 genuinely holds** — for both straight-line and branch-shaped programs, under the law that is actually authoritative (`discoverLoanLastUses`), independently re-verified by this plan's third mechanism. The split between "the law that decides" and "the law that reports facts" is not a correctness gap.

**What is real debt.** D-05 required: *"OWN-03's CFG/edge-specific liveness should remove the quadratic factor [D-02-03] by construction... make `recomputed_work` count the propagation so the metric stops understating real cost."* Neither half of that requirement is achieved on the path that matters:

- `discoverLoanLastUses` is unchanged from Phase 2 and remains the transitive, potentially-quadratic law — and it is the ONLY law deciding admission, in both `checkLinear` and `checkBranch`. `loanLivenessFixpoint`'s linear cost is real, but it prices a computation (`LoanEndpoints`) that plays no role in whether a program is accepted, so its cheapness does not translate to a cheaper checker.
- `discoverLoanLastUses`'s own propagation work is still never counted: `analyzeStraightLine`'s `result.Work` is `typeNodeCount(...) + len(body.Bindings) + 1` plus one increment per binding in the main loop (`check.go:1017,1062`) — no increment anywhere inside `discoverLoanLastUses` itself (verified by `grep -n "Work++" check.go`: every hit is in the two `analyze*` functions, none inside `discoverLoanLastUses`). `analyzeArmBody` has the identical shape (`check.go:698,756`). `checkBranch` additionally adds `fixpoint.work` (03-03's honest, counted, LINEAR cost) to `work` — meaning a branch function's `recomputed_work` now UNDERSTATES its real cost by two separate amounts stacked together: `discoverLoanLastUses`'s still-uncounted quadratic propagation, plus the fact that the counted `fixpoint.work` prices a computation that isn't even on the admission path. The metric's understatement, which D-05 asked to fix, is unchanged for straight-line functions and arguably compounded for branch functions.

**Why not blocking.** Bounded by the same `MaxTokens`/parser limits as Phase 2; fails closed; no unbounded path; every currently-shipped fixture (Phase 1-3 corpus, all conflict-matrix fixtures, all generated/fuzzed cases) checks well within existing time budgets. This is a counted-work honesty gap and a residual cost-shape gap, not a live vulnerability or a wrong verdict.

**Phase 4+ fix.** Either (a) retire `discoverLoanLastUses` in favor of driving BOTH admission and endpoint-fact materialization from `loanLivenessFixpoint` directly (the literal "no longer on the production path" the 03-05 plan's own must_haves originally described for 03-03, deferred there to protect `TestOwnershipSequenceExhaustive` and 03-06's absence invariant), or (b) instrument `discoverLoanLastUses` to count its own transitive-scan work honestly, matching the `result.Work++`-per-operation convention `blockLoanLiveness` already established. (a) is the more complete fix and the one D-05 originally intended; (b) is the smaller one if (a) is deferred again.

### D-03-02 — an undeclared borrow-derived return exports as if fully owned

**Finding.** `originvalidate.ValidatePublished` (`internal/compiler/originvalidate/originvalidate.go:110-114`) begins:

```go
for _, function := range program.Functions {
    if function.PublicOrigin == nil {
        continue
    }
    ...
}
```

It never calls `RecomputeOrigin` — the independent, body-derived recomputation — for a function that declares NO origin. `RecomputeOrigin` is only ever invoked to CHECK a declaration that already exists; it is never used to DETECT that one is missing. `BuildInterface` (`originvalidate.go:159-184`) then copies `function.PublicOrigin` (nil) straight into the exported `FunctionSignature`, alongside the type's independently-derived `Abilities` (identical to a genuinely fresh owned value of the same type — no ability distinguishes "this value aliases the parameter" from "this value is freshly owned").

**Concrete reachable instance.** `internal/compiler/check/check_exclusive_test.go`'s `exclusive_borrow_clean` fixture:

```
export { fn relay }
fn relay(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let reviewed = borrow view
  view
}
```

`relay` is EXPORTED, checks with zero diagnostics, and its return (`view`) is exclusively-borrow-derived from `buffer` — yet its return type carries no `borrow(path)` annotation. `interface export` on this module would produce a `FunctionSignature{PublicOrigin: nil, Abilities: [drop, share, send, escape]}` for `relay`: structurally identical to a function that returns a brand-new, unrelated `Buffer` it owns outright. A separate-compilation consumer has no way to learn, from the signature alone, that the returned value aliases the caller's own argument.

**Why the grammar rule does not close this.** 03-06 replaced a `check.go` admission gate ("a borrow-derived return with no declared origin is rejected") with a PARSE-TIME grammar rule ("a `borrow` token in return-type position must be followed by `(path)`"). Those enforce different things: the grammar rule constrains the SPELLING of an annotation once someone chooses to write one; it does nothing to require that anyone write one in the first place. A function whose body happens to return a borrow-derived value, with a plain (unannotated) return type, is syntactically legal and was never within the grammar rule's reach — only the dropped `check.go` gate could have caught it, and 03-06 removed that gate rather than narrowing it, to protect `exclusive_borrow_clean`'s own already-shipped acceptance (see 03-06-SUMMARY.md, Deviations #1). The substitution is therefore NOT complete: the grammar rule and the dropped gate do not have the same reach, and 03-06's summary understated this when it framed the replacement as closing "the origin annotation is mandatory."

**Why not blocking today.** This language has no cross-function call construct yet — nothing in the executable semantics ever invokes one Lang function from another and observes both the alias and an independent mutation/move in the same run. The gap is real in the exported ARTIFACT (`interface export`'s output) but currently unreachable as an executable unsoundness, because there is no consumer that could act on the misleading signature. `originvalidate.KnownEscape` already names one coordinated frontend/summary lie as an accepted residual (`escape:coordinated-frontend-summary-lie`); this is a DIFFERENT, uncoordinated gap — the frontend is honest (it genuinely never received an annotation), the summary is honest about what it was given — the gap is that nothing ever asks whether an annotation SHOULD have existed.

**Fix, before Phase 4 introduces calls.** Restore an admission-time (or `ValidatePublished`-time) check that runs `RecomputeOrigin` unconditionally for every EXPORTED function — not gated on `PublicOrigin != nil` — and rejects (or requires an explicit, named "not part of the public surface" opt-out) when the recomputation finds a real origin the declaration omitted entirely. This is a narrower, better-scoped version of the gate 03-06 dropped: scoped to exported functions only (03-06's removed gate applied to every straight-line function, which is what broke `exclusive_borrow_clean`, a non-exported... actually `relay` in the fixture above IS exported, so scoping to "exported" alone does not by itself save `exclusive_borrow_clean`'s CHECK-time acceptance — the fix must live in `ValidatePublished`/`interface export`'s own admission path, not `check.go`'s, so `lang check` keeps accepting `exclusive_borrow_clean` exactly as today while `interface export` on the same source would newly refuse to publish it without an origin, which is the shape the ORIGINAL fixture never needed to satisfy since it never called `interface export`).

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Recorded: 2026-09-04 at `68ed4f7` (03-05, mid-phase gate)*

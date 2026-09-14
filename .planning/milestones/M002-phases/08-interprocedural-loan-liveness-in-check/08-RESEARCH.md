# Phase 08: Interprocedural Loan Liveness in `check` - Research

**Researched:** 2026-09-09
**Domain:** interprocedural summary-based dataflow analysis inside a Go compiler checker (IFDS-style memoized liveness, fail-closed fixpoint bounding, cost-gate instrumentation)
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

08-CONTEXT.md contains 40 lettered decisions (D-08-01 through D-08-40), all synthesized by advisor research and independently re-verified against the shipped tree by the orchestrator (three corrections recorded: D-08-13's bound axis, D-08-24's span limitation, D-08-32's metric-name chokepoints). This research does not re-litigate any of them. Load-bearing summary for the planner:

- **Two summary bits** (`UsesParam`, `ReturnsBorrowOfParam`), transitive through relays, derived once per function in `callgraph.Order`'s reverse postorder, memoized in-run on the existing `callSignatureTable` — no schema bump, no new package (D-08-01..D-08-06).
- **Mechanism placement:** the interprocedural fact enters through a **forward canonicalization pre-pass in `derivePlaceLoans`** (not the backward transfer function — that placement is a proven silent-failure design), plus **one added clause in `blockLoanLiveness`** for `UsesParam` (D-08-07, D-08-08).
- **D-07-49 entry defect must be fixed in BOTH admission paths**: `computeLoanLastUses` (AST-shadow, `check.go:2979`) and the real `derivePlaceLoans` path — a test must fail if only one is fixed (D-08-09).
- **Program-order invariant is load-bearing** for both correctness (canonicalization) and cost (192x penalty reversed) — state it in the derivation's doc comment and pin it with a test (D-08-11).
- **Memo is within-run only**, keyed by function ID, rebuilt every `check.Program` call, never persisted — test that it is never persisted (D-08-12).
- **The only genuine fixpoint is the intraprocedural worklist** inside `loanLivenessFixpoint`; the bound is `k × len(blocks) × distinctLoanCount`, derived not a flat constant, and is an internal-consistency assertion (no `.lang` program can force divergence — arity 1, no loops) (D-08-13..D-08-15).
- **Seam:** unexported package-level var in the `callReturnTypeDerivationSeam` shape (verified `check.go:493`), not the exported `pathoracle.TerminatorKindsOverride` shape (D-08-16).
- **Namespace:** `check.*` (not `core.*`) — Phase 09's peer uses a reachability closure, not a worklist, so it can't fail the same way (D-08-17).
- **Native-recursion → explicit stack conversion required in this phase** for `loanLivenessFixpoint`'s cycle pre-walk (`check.go:1157-1180`), mirroring D-07-18; new transitive `ReturnsBorrowOfParam` walk must be iterative from the start (D-08-19a). Promote `check.cfg_back_edge` to a coded diagnostic while in the same function (D-08-19b).
- **Diagnostic code:** mint `check.interprocedural_loan_liveness` (never reuse an `ownership.*` code) (D-08-20), promote to `core.*` only in Phase 09 (D-08-21).
- **Blame:** `Primary` = the failing operation (the `take`), not the call site or callee declaration (D-08-22). Fixed three-role cause template: borrow-created-here / loan-extended-by-call / callee-return-contract — no rotation needed, positions are role-keyed (D-08-23).
- **Cross-file spans impossible today** (`diagnostic.Span{Start,End int}`, no file field) — do not widen `Span` this phase; callee-contract cause is ID-only and spanless (D-08-24). `Repairs: nil` — the only valid fix lives in a different function than `Primary` (D-08-25).
- **Criterion-4 disclosure:** on refusal, name the one field consulted (`return.mode` or `parameters[0].mode`). On acceptance, criterion 4 is **NOT met this phase** — all three candidate vehicles (`ExplainSummary`, `cache`, a new `/1` sibling doc) verified to fail structurally; satisfy with a table-driven test as documentation-of-record, take the always-on runtime record question to the mid-phase gate (D-08-26).
- **OWN-09 document conflict flagged, not resolved**: interim rule for Phase 08 is `ownership.*` codes stay, interprocedural check runs **last**, after intraprocedural admission passes (D-08-27).
- **Criterion-1 corpus (D-08-28)** needs, or the gate is decorative: (1) both S-006 patterns as refuse/accept twins differing in exactly one callee contract field, (2) a depth-≥2 relay chain, (3) `testdata/phase07/relay_escort_witness.lang` flipping from clean to refused, (4) a negative control pair varying only `Fails`/`Foreign.*`, (5) a should-add both-match-arms fixture.
- **Cost gate (EFF-02):** deterministic work-counter is the **hard** gate; `elapsed_ns` is observed-only (D-08-29). Applying p50/p95/CoV to work-count samples is literal, not an invented parallel instrument (D-08-30). Bound: fitted growth exponent of work against **operation count** (not function count) ≤ 1.2, plus a ratio-stability tripwire at 15% (D-08-31). **Two hardcoded chokepoints must be widened** (`measure.Demote`'s `"recomputed_work"` string literal and `session.QLT02GateEligibleMetrics()`) or a new metric name is silently demoted to `observed` and can never block — this is a **named task with its own test obligation** (D-08-32). Manifest mechanics: `machine_id` required, no schema change needed (D-08-33).
- **Corpus generator:** port (don't import) the spike's `iplive/corpus.go` `Generate(shape, n)` — synthetic `core.Program` values directly, **not** through the real `.lang` parser (parse time would dominate and mask the curve) (D-08-34). Required shapes: `chain`, `diamond`, `dense`/`parser-shaped`, `forward` (star) — `tree` explicitly dropped (D-08-35). New `lane:interprocedural-cost-scaling` risk lane, not the default edit loop (D-08-36).
- **Cross-run cache is a declared non-goal** this phase (Phase 11/QLT-06); `PHASE-08-DEBT.md` must record it, plus the verified all-clear that `ClosureDigest` is not wired into `cache.Input` today (D-08-37).
- **2× scope-cut trigger**: if summary-table + liveness-law work exceeds ~2×, the cost-gate *instrument work* (chokepoint widening, generator, lane) renegotiates into Phase 09 — never the criterion-1 corpus, never the seeded mutation-kills, never the two-path D-07-49 fix (D-08-38).
- **`PHASE-08-DEBT.md` written at planning time** (not phase end), carrying at minimum D-08-15, D-08-26, D-08-27, D-08-37, D-08-38 (D-08-39).
- **Mid-phase gate agenda:** (a) OWN-09 conflict, (b) whether criterion-4's accepted-case disclosure becomes a runtime artifact, (c) whether both D-08-09 admission paths remain load-bearing (D-08-40).

### Claude's Discretion

- Plan decomposition and wave ordering, subject to two hard constraints: summary derivation runs **after** `callgraph.Order` proves acyclicity; D-08-32's chokepoint widening lands **before** any manifest row claims to be a hard gate.
- Exact Go identifier and file names for the summary bits, the memo table field, the seam, and the generator.
- Whether the summary derivation is a method on the existing table builder or a sibling function it calls.
- The exact corpus file names, module paths, and generated-shape size ladder (subject to D-08-35's required shape list).
- Whether the ratio-stability tripwire ships as a separate manifest row or an in-test assertion.

### Deferred Ideas (OUT OF SCOPE)

- `corevalidate`'s independent liveness peer, D-03-02 closure — Phase 09 (OWN-07, OWN-08); mechanism is a reachability closure, must not import `check`.
- Promotion of `check.interprocedural_loan_liveness` to `core.*` — Phase 09.
- OWN-09's retirement of the intraprocedural law — Phase 09 per REQUIREMENTS.md mapping (conflict flagged D-08-27, mid-phase gate agenda item).
- Criterion 4's accepted-program disclosure as a runtime artifact — deferred, mid-phase gate cost question.
- A persistent cross-run summary cache — Phase 11/QLT-06.
- Widening `diagnostic.Span` with a file/module field — not this phase.
- A third summary bit or per-arm summary bits — not this phase; reopen at arity>1 or Phase 12 `Result` match arms.
- Publishing the two liveness bits into `lang.interface/1` or `/2` — deliberately not done.
- The interpreter's fixed call-stack ceiling — SEM-08, Phase 10.
- Re-measuring cache invalidation fallout on production-shaped corpora — not worth the sweep cost; spike numbers stand.
- A `tree`-shaped cost corpus — dropped, dominated by `diamond`+`chain`.
- Match-arm call fixture for liveness — recommended regression coverage, not gate-blocking.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OWN-06 | `check` derives interprocedural loan liveness from callee signatures only — never by re-walking callee bodies — and terminates under a fail-closed iteration bound. | Verified code locations for the consumption-only per-call-site read (`callSignatureTable.lookup`, `core.FunctionSignature` has no Linear/Match field — quoted below); verified fail-closed bound mechanics in `loanLivenessFixpoint` (`check.go:1142-1225`, `work` counter already threaded); spike S-006 findings 1-3 establish the memoized, one-pass-RPO, canonicalization-pre-pass design is sound and cheap. |
| EFF-02 | Interprocedural admission cost is measured on realistic call-graph fan-out under the existing p50/p95/CoV protocol, stays within a declared bound, and is recorded in the feedback-budget manifest. | Verified `measure/statistics.go` is unit-agnostic (`[]int64` Samples); verified the two hardcoded `"recomputed_work"` chokepoints (`measure.Demote:131-135`, `session.QLT02GateEligibleMetrics():55-57`) that must widen or the gate is decorative; spike S-006's growth-exponent methodology and generator shape (`iplive/corpus.go`) to port. |
</phase_requirements>

## Summary

This phase is almost entirely pre-decided: 08-CONTEXT.md is the product of four parallel advisor-research passes plus orchestrator re-verification against the shipped tree, and every design question a planner would normally research (mechanism placement, memoization scheme, bound derivation, diagnostic identity, cost-gate shape) is already locked with cited evidence. The remaining research value for planning is **narrow and mechanical**: confirming the exact shape of the code the plan will touch (verified below with line numbers and verbatim quotes), confirming the test-file conventions this codebase already uses for this exact kind of control (fault-injection seams, work-ratio assertions, mutation-kill tests), and flagging the two places a plan could silently go wrong even while following the locked decisions correctly — the D-08-09 two-path fix and the D-08-32 two-chokepoint widening, both of which are structurally invisible in a diff that only touches one of the two sites.

There is no new external dependency in this phase (zero-external-production-dependency is a standing project constraint — `.planning/STANDING-VERDICTS.md`), so the Package Legitimacy Audit and Standard Stack sections below are correspondingly thin: this is pure Go-stdlib compiler-internals work extending an existing package.

**Primary recommendation:** Plan this phase as four waves matching the four CONTEXT.md deliverables in dependency order — (1) D-07-49 entry-defect fix in both admission paths + D-08-19a/b hardening, landed and tested first since it's the smallest, most mechanical, and unlocks the criterion-1 fixture flip; (2) the two summary bits + memo table + canonicalization pre-pass + `UsesParam` transfer clause (the core mechanism); (3) the fail-closed bound + seam + diagnostic identity/blame/causes; (4) the cost-gate instrumentation (chokepoint widening, generator port, lane, manifest ratification) — explicitly the wave the 2× scope-cut trigger may push into Phase 09, per D-08-38.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Two-bit summary derivation (`UsesParam`, `ReturnsBorrowOfParam`) | `check` (in-process compiler pass) | — | Consumes callee body once per function at signature-table build time; D-08-04/D-08-06 lock this in `check`, unexported, not a shared package |
| Interprocedural liveness admission (per call site) | `check` (`derivePlaceLoans` + `blockLoanLiveness`) | — | Existing intraprocedural dataflow engine extended in place, not a new engine (D-08-07/D-08-08) |
| Fail-closed iteration bound / refusal | `check` diagnostic layer | `diagnostic` package (Span/Cause/identity machinery) | New coded diagnostic `check.interprocedural_loan_liveness`; Span/Cause shape reused unmodified (D-08-24) |
| Independent peer re-derivation | `corevalidate` | — | Explicitly Phase 09, not this phase (OWN-07/D-08-06) |
| Cost measurement / gate | `internal/compiler/measure` + `internal/compiler/session` | test-only corpus generator | Two chokepoints in `measure` and `session` must widen for the gate to be live (D-08-32) |

## Standard Stack

No new external packages. This phase extends existing internal packages only:

| Package | Purpose in this phase | Why no alternative considered |
|---------|------------------------|-------------------------------|
| `internal/compiler/check` | All mechanism code (summary derivation, canonicalization pre-pass, transfer-function clause, bound, diagnostic) | Project standing verdict: zero external production dependencies; this is pure Go stdlib compiler-internals work |
| `internal/compiler/callgraph` | Reuses `Order` (RPO) unmodified | Already proven acyclic-then-ordered by Phase 07; free reuse (D-08-03) |
| `internal/compiler/measure` | `Samples`/`Summary`/`Demote` — extend, don't replace | Genuinely unit-agnostic; extending its one hardcoded chokepoint is required (D-08-32) |
| `internal/compiler/session` | `QLT02GateEligibleMetrics()`, `risk_lanes.json`, budget manifest | Same chokepoint-widening requirement |
| `internal/compiler/diagnostic` | `Diagnostic`/`Cause`/`Span` unmodified | D-08-24 forbids widening `Span` this phase |

**Installation:** none — no new dependencies to install.

## Package Legitimacy Audit

Not applicable. This phase adds zero external packages (project-wide zero-external-production-dependency standing verdict, `.planning/STANDING-VERDICTS.md`). No `go.mod` changes are anticipated; if a plan step proposes adding one, that is itself a red flag requiring a checkpoint against the standing verdict.

## Architecture Patterns

### System Architecture Diagram

```
core.Program (post-body-checked, immutable)
        |
        v
callgraph.Order  ---------------------------> reverse postorder (RPO), proven acyclic (Phase 07)
        |
        v
[NEW] per-function summary derivation (RPO order, callee-before-caller)
    reads: each function's OWN checked core.LinearOperation body (once)
    writes: UsesParam, ReturnsBorrowOfParam onto callSignatureTable entry
    memo: keyed by function ID, in-run only, never persisted
        |
        v
callSignatureTable (extended; still body-blind by TYPE -- holds core.FunctionSignature)
        |
        |  <-- per-call-site consumption reads ONLY this table from here on
        v
derivePlaceLoans (forward canonicalization pre-pass, program order)
    [NEW] OpCall branch: if callee.ReturnsBorrowOfParam ->
          alias call result's loan identity to argument's current loan
    else: existing fallthrough (fresh identity) -- unchanged
        |
        v
blockLoanLiveness (backward transfer function, per CFG block)
    [NEW] OpCall clause: if callee.UsesParam -> counts as a use of the
          argument's current loan chain (symmetric to a plain reference)
        |
        v
loanLivenessFixpoint (worklist over blocks; UNCHANGED shape)
    [NEW] fail-closed iteration bound: k * len(blocks) * distinctLoanCount
          exceeding it -> check.interprocedural_loan_liveness (bound path,
          reachable only via unexported test seam, not from any .lang source)
        |
        v
admission result: accept, OR refuse with check.interprocedural_loan_liveness
    Primary = the failing operation (e.g. the `take`)
    Causes  = [borrow_created_here, loan_extended_by_call, callee_return_contract]

Parallel, gated separately:
[NEW] cost-gate instrumentation
  synthetic core.Program generator (chain/diamond/dense/forward shapes)
        |
        v
  work counters (deterministic; wall-clock observed-only)
        |
        v
  measure.Summary (p50/p95/CoV, unit-agnostic) + fitted growth exponent vs. OPERATION count
        |
        v
  measure.Demote + session.QLT02GateEligibleMetrics()  [BOTH must widen to accept
        the new metric name, or it is silently demoted to "observed" and never gates]
        |
        v
  qlt02_budget_manifest.json row (hard gate) + risk_lanes.json entry (changed-risk lane)
```

### Recommended Project Structure

No new files/packages at the top level — this phase edits existing files in place:

```
internal/compiler/check/
├── check.go                    # extend: buildCallSignatureTable, derivePlaceLoans,
│                                #   blockLoanLiveness, loanLivenessFixpoint,
│                                #   computeLoanLastUses (D-07-49 fix, both paths)
├── check_test.go               # add: twin-pair corpus tests, mutation-kill tests,
│                                #   memo-never-persisted test, work-ratio test
internal/compiler/measure/
├── statistics.go               # widen: Demote's "recomputed_work" chokepoint
internal/compiler/session/
├── session_phase6_budget.go    # widen: QLT02GateEligibleMetrics()
├── qlt02_budget_manifest.json  # new row(s): the growth-exponent / work metric
├── risk_lanes.json             # new lane: interprocedural-cost-scaling
testdata/phase07/
├── relay_escort_witness.lang   # existing; assertion flips clean->refused
testdata/phase08/                # new: twin-pair corpus, negative-control pair,
│                                 #   depth->=2 relay chain, match-arm fixture
.planning/phases/08-.../
├── PHASE-08-DEBT.md            # written at planning time (D-08-39)
```

### Pattern 1: Forward canonicalization pre-pass for interprocedural aliasing

**What:** Rather than teaching the backward liveness transfer function about calls, bind the call result's loan identity to the argument's loan identity in the SAME forward pre-pass that already handles reborrow chains.

**When to use:** Any time an interprocedural fact changes what identity a downstream backward analysis should see, and the backward walk would otherwise reach the fact "too late" in its reverse traversal order.

**Example (verified in tree, `internal/compiler/check/check.go:1049-1068`):**
```go
// derivePlaceLoans -- existing reborrow-chain shape the OpCall branch mirrors
func derivePlaceLoans(operations []core.LinearOperation) placeLoanChain {
	chain := placeLoanChain{
		latestLoan: make(map[string]string, len(operations)),
		parentLoan: make(map[string]string, len(operations)),
	}
	for _, operation := range operations {
		inherited := chain.latestLoan[operation.SourceID]
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			chain.parentLoan[operation.LoanID] = inherited
			if operation.TargetID != "" {
				chain.latestLoan[operation.TargetID] = operation.LoanID
			}
			continue
		}
		if inherited != "" && operation.TargetID != "" {
			chain.latestLoan[operation.TargetID] = inherited
		}
	}
	return chain
}
```
D-08-07 locks the extension: add an `operation.Kind == core.OpCall` branch that, when the callee's `ReturnsBorrowOfParam` is true, sets `chain.latestLoan[operation.TargetID] = chain.latestLoan[operation.SourceID]` (same-loan-identity aliasing); otherwise the existing generic fallthrough already does the correct thing (fresh identity, since `inherited` for a fresh call result is `""`).

### Pattern 2: Symmetric "use" clause in a backward transfer function

**What:** A call that reads through its parameter counts as a use of the argument's loan chain, exactly like a plain reference operation already does.

**Example (verified in tree, `internal/compiler/check/check.go:1094-1099`, the exact loop the new clause joins):**
```go
for loan := chain.latestLoan[operation.SourceID]; loan != "" && !recorded[loan]; loan = chain.parentLoan[loan] {
    work++ // one unit per chain-ancestor step walked
    live[loan] = true
    recorded[loan] = true
    uses = append(uses, loanBlockUse{loanID: loan, operationIndex: index, operationID: operation.ID})
}
```
D-08-08 locks that an `OpCall` whose callee's `UsesParam` is true enters this same walk over `operation.SourceID` — no new loop shape, no new lattice.

### Pattern 3: Deterministic work-counter cost gate (never wall-clock for the hard gate)

**What:** Every unit of analysis work (transfer-function evaluation, worklist reinsertion, chain-ancestor step, summary derivation) increments a plain `int` counter returned alongside the result. Wall-clock is recorded separately and is *observed*, never *hard*.

**Example (verified in tree, `internal/compiler/check/check.go:1142` `loanLivenessFixpoint`'s existing counters, already present and already the precedent the summary derivation's own counter must follow):**
```go
work := 0
for len(queue) > 0 {
    ...
    work++ // one transfer-function evaluation
    ...
    _, newLiveIn, transferWork := blockLoanLiveness(block.operations, placeLoan, liveOut)
    work += transferWork
    if !loanSetsEqual(liveIn[id], newLiveIn) {
        ...
        work++ // one worklist reinsertion
    }
}
return loanLivenessResult{liveIn: liveIn, work: work}, nil
```

### Pattern 4: Fault-injection seam, unexported, same-package-only

**What:** A package-level `var` defaulting `false`/production-safe, flipped only by a same-package test that defers its own restore.

**Example (verified in tree, `internal/compiler/check/check.go:493-500`):**
```go
var callReturnTypeDerivationSeam = false
```
Read at `check.go:1928`; flipped and deferred-restored in `check_test.go:3412`. D-08-16 locks this exact shape (not `pathoracle.TerminatorKindsOverride`'s exported-var shape) for the new fail-closed-bound seam.

### Anti-Patterns to Avoid

- **Putting the interprocedural fact in the backward transfer function directly** — proven wrong by S-006 iteration 2: in `borrow; call; move; use(result)` the backward walk reaches the `move` before the call establishes the alias, so the refusal never fires. Passes every accept-case test and silently under-approximates.
- **A flat numeric iteration-bound constant** — rejected by D-08-14: equally unprovokable, no derivation trail, false-positives the day `blocks × loans` legitimately grows.
- **Embedding the bound's numeric value in a `Cause` detail or `Message`** — `Causes` participate in diagnostic ID identity (`diagnostic.go:96-112`); a retuned bound would silently move every existing diagnostic ID for that code (D-08-18).
- **Naming the new cost metric anything other than the literal string `"recomputed_work"`** without widening both chokepoints first — silently produces a decorative gate that reports `"gate_type": "hard"` in the manifest JSON but can never actually block (D-08-32).
- **Generating the cost corpus through the real `.lang` parser** at hundreds-of-functions scale — parse time would plausibly dominate and mask the exact curve the gate exists to see (D-08-34).
- **Fixing only one of the two D-07-49 admission paths** (`computeLoanLastUses` or `derivePlaceLoans`) — reproduces the literal D-02-03/D-03-01 failure shape and D-07-21's lesson ("fixing one of N independent peers is not fixing the item").

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Interprocedural summary propagation | A new whole-program dataflow engine or a shared `internal/compiler/summary` package | Extend the existing intraprocedural `loanLivenessFixpoint`/`derivePlaceLoans`/`blockLoanLiveness` in place | This is the IFDS/tabulation insight verbatim: the summary table *is* the algorithm, not a new engine (S-006 finding #3); ARCHITECTURE.md §4 rejects a shared library outright |
| Ordering functions for callee-before-caller derivation | A new topological sort or call-graph-level fixpoint | `callgraph.Order`'s existing reverse postorder | Already proven acyclic by Phase 07; free reuse; S-006 finding #2: "one pass with NO call-graph-level fixpoint at all" |
| Recursive graph traversal depth safety | Trusting Go's native call stack for the cycle pre-walk | Explicit-stack iterative traversal, per D-07-18's precedent in `callgraph.Order` | Depth is program-controlled (CFG block count); native recursion is Pitfall 4's stack-overflow failure mode |
| Cost regression detection | A wall-clock timing threshold | A deterministic work counter (already threaded through `loanLivenessResult.work`) plus `measure`'s existing p50/p95/CoV protocol | Wall-clock is confounded by machine noise the project has already named (`CoVDemotionThreshold`'s own comment); rustc/Z3 precedent for deterministic counters |

**Key insight:** every mechanism this phase needs already exists in the tree in miniature — reborrow-chain aliasing, backward-walk "use" clauses, deterministic work counters, unexported fault-injection seams, RPO ordering. The phase's actual work is *extending four existing functions with one clause each* and *wiring two existing-but-unconnected chokepoints*, not inventing new subsystems.

## Common Pitfalls

### Pitfall 1: Backward-transfer placement silently passes every accept-case test
**What goes wrong:** Putting the interprocedural clause in `blockLoanLiveness`'s backward walk instead of `derivePlaceLoans`'s forward pre-pass causes the refusal to never fire on the target unsafe pattern, while every hand-written accept fixture still passes.
**Why it happens:** The backward walk visits `move` before the `call` that establishes the alias in `borrow; call; move; use(result)`, so by the time the walk reaches the call, the loan already looks dead.
**How to avoid:** Follow D-08-07 exactly — canonicalization pre-pass in `derivePlaceLoans`, program order, before any backward walk begins.
**Warning signs:** A refuse-case fixture that checks clean; a twin pair where only the accept member is exercised by an existing test.

### Pitfall 2: A criterion-1 corpus that varies the wrong axis
**What goes wrong:** A twin pair that differs in the *caller's* shape (not the callee's declared contract field) passes even with the interprocedural law entirely deleted, because the existing intraprocedural `ownership.*` law already refuses same-function violations.
**Why it happens:** It's easy to write "refuse case" and "accept case" fixtures that differ in several places at once, especially when reusing existing borrow/take grammar.
**How to avoid:** D-08-28's mechanical check — the AST diff between the refusing and accepting twin members must touch only the callee's declared return/parameter type field, nothing else. Assert this mechanically in the corpus harness, not just by inspection.
**Warning signs:** Deleting the new interprocedural code and re-running the corpus still shows the same pass/fail pattern.

### Pitfall 3: Fixing one of two admission paths
**What goes wrong:** `computeLoanLastUses` (AST-shadow path) and the real `derivePlaceLoans`/`core.LinearOperation` path are two independent re-implementations of the same law; landing the `"call"` case in only one leaves the other silently unfixed.
**Why it happens:** They look similar and a diff touching one can appear complete in isolation, especially since only one may be "load-bearing" for a given fixture's outcome.
**How to avoid:** D-08-09 mandates a test that fails if only one path receives the fix — e.g. a fixture that only the AST-shadow path would catch and one only the real path would catch.
**Warning signs:** A green corpus that never actually exercises `computeLoanLastUses`'s call handling (check which admission path each fixture actually traverses, e.g. `check.go:1328` vs `:2529` call sites).

### Pitfall 4: Naming the cost metric anything but the exact literal `"recomputed_work"` without widening the chokepoints
**What goes wrong:** A manifest row with `"gate_type": "hard"` and a plausible-looking new metric name (e.g. `recomputed_work_per_op_exponent`) is silently demoted to `observed` at read time and can never block a build.
**Why it happens:** Two independent call sites hardcode the bare string `"recomputed_work"` (`measure.Demote:135`, `session.QLT02GateEligibleMetrics():56-57`); neither failure is visible from the manifest JSON alone — a reviewer reading only the JSON would conclude the gate is live.
**How to avoid:** D-08-32 mandates widening both chokepoints as a named task with its own test, re-deriving `TestDemoteHasExactlyOnePromotionPassthrough` so it still forbids a second promotion path.
**Warning signs:** `Demote` returning `VerdictObserved` for a row you believe should be `VerdictBlocking`; the `TestDemoteHasExactlyOnePromotionPassthrough` test count of `return VerdictBlocking` statements not increasing when it logically should have to.

### Pitfall 5: Reversed-order summary derivation reintroducing a 192x quadratic penalty
**What goes wrong:** A future refactor (or an unwary Phase-08 implementation choice) that iterates a function body out of program order before deriving its summary silently regresses from linear to quadratic in body length.
**Why it happens:** The two-pass-suffices property is entirely a consequence of program-order input; S-006 iteration 5 measured 4.0 work units/op flat in order vs. 12.4→767.0 reversed (192x at k=512).
**How to avoid:** D-08-11 — state the invariant in the derivation's doc comment and pin it with a test, following `TestReborrowChainWorkIsLinear`'s precedent (already in tree at `check_test.go:737-753`, verified above).
**Warning signs:** A cost-sweep result where the `forward` (star) shape's growth exponent unexpectedly climbs above ~1.1.

### Pitfall 6: Native recursion in the cycle pre-walk under adversarial straight-line bodies
**What goes wrong:** `loanLivenessFixpoint`'s existing cycle pre-walk (`var walk func(id string) error` at `check.go:1150-1165`, verified above) recurses natively over block successors; a sufficiently long straight-line CFG body can exhaust the native Go call stack.
**Why it happens:** Depth is bounded by CFG block count within one function body — smaller than `callgraph`'s roots, but "bounded by body size" is not the same guarantee as "safe", and a long straight-line body is adversarially reachable even without loops in the source language.
**How to avoid:** D-08-19a — convert to an explicit stack in this phase, mirroring `callgraph.Order`'s existing iterative three-color DFS (D-07-18 precedent). Do this even though the language has no loops — the CFG shape itself can still be made long via straight-line sequencing.
**Warning signs:** A stack-overflow panic (not a clean refusal) on a large synthetic fixture.

## Code Examples

### Existing signature table extension point (verified `internal/compiler/check/check.go:577-609`)
```go
type callSignatureTable struct {
	entries map[string]core.FunctionSignature
}

func (t callSignatureTable) lookup(calleeID string) (core.FunctionSignature, bool) {
	entry, ok := t.entries[calleeID]
	return entry, ok
}
```
The two new bits do NOT go on `core.FunctionSignature` (that would be a schema bump, D-08-04 forbids it) — they go on a sibling in-memory structure this same package owns, keyed alongside these entries.

### `core.FunctionSignature` has no body-fact field to reuse (verified `internal/compiler/core/core.go:181-191`)
```go
// closure. It deliberately has no Linear/Match field at all -- not merely an
// omitted one -- so a consumer decoding this type structurally cannot reach a
// body even by accident (the same argument style as /0).
...
type FunctionSignature struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Parameters []ParameterContract `json:"parameters"`
	Return ReturnContract `json:"return"`
	Abilities []Ability `json:"abilities"`
	Callable bool `json:"callable"`
	Fails string `json:"fails,omitempty"`
	Foreign ForeignReach `json:"foreign"`
	ClosureDigest string `json:"closure_digest"`
}
```
This confirms D-08-01's central falsification: there genuinely is no field here (and by explicit doc-comment design, never will be one) that could answer `UsesParam` from declared data alone.

### `ParameterContract.Mode` / `ReturnContract.Mode` — the two fields the liveness law consults (verified `internal/compiler/core/core.go:245-282`)
```go
type ParameterContract struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Mode string `json:"mode"`   // "owned", "shared", or "exclusive"
	Drops bool `json:"drops"`
}

type ReturnContract struct {
	Type string `json:"type"`
	Mode string `json:"mode"`   // "owned", "shared", or "exclusive"
	Paths []string `json:"paths"`
	Fresh bool `json:"fresh"`
}
```
D-08-23's cause 3 discloses exactly one of these `Mode` values per refusal — never a bitmask.

### The existing fixture that must flip from clean to refused (verified `internal/compiler/check/check_test.go:2338-2364`)
```go
func TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness(t *testing.T) {
	source := readPhase07Fixture(t, "relay_escort_witness.lang")
	result := Program(mustParseProgram(t, source))
	if len(result.Diagnostics) != 0 {
		t.Fatalf("expected zero error diagnostics (the D-03-02 interprocedural finding), got %+v", result.Diagnostics)
	}
	...
}
```
This test's own name and body assert the current (pre-Phase-08) clean-admission behavior. D-08-28.3 requires this to become a refusal assertion once the D-07-49 fix lands — renaming the test is itself part of the required work, not an incidental cleanup.

### `computeLoanLastUses`'s missing `"call"` case — the exact D-07-49 defect site (verified `internal/compiler/check/check.go:2996-3006`)
```go
kind := core.OpCopy
loanID := ""
switch binding.RHS.Kind {
case "take":
	kind = core.OpMove
case "borrow":
	kind = core.OpBorrowShared
	loanID = fmt.Sprintf("shadow:loan:%d", index)
case "borrow_mut":
	kind = core.OpBorrowExclusive
	loanID = fmt.Sprintf("shadow:loan:%d", index)
}
```
Confirmed: no `"call"` case exists; a call binding falls through to the `core.OpCopy` default, exactly as D-08-09 states. This is one of the two sites requiring the fix.

### Existing work-ratio pinning test pattern to replicate (verified `internal/compiler/check/check_test.go:737-753`)
```go
func TestReborrowChainWorkIsLinear(t *testing.T) {
	series := []int{10, 100, 1_000, 10_000}
	work := make([]int, len(series))
	for index, n := range series {
		block := cfgBlockSpec{id: "chain:block:ratio", operations: reborrowChainOperations(n), successors: nil}
		result, err := loanLivenessFixpoint("chain", []cfgBlockSpec{block})
		...
		work[index] = result.work
	}
	for index := 1; index < len(series); index++ {
		operationRatio := float64(series[index]) / float64(series[index-1])
		workRatio := float64(work[index]) / float64(work[index-1])
		if workRatio > operationRatio*2 {
			t.Fatalf(...)
		}
	}
}
```
Use this exact pattern (or a size-swept variant with real growth-exponent fitting per D-08-31) for both the new summary-derivation cost test and the interprocedural admission cost test.

### Cost-gate chokepoints to widen (verified `internal/compiler/measure/statistics.go:120-135` and `internal/compiler/session/session_phase6_budget.go:47-56`)
```go
// measure/statistics.go
func Demote(requested string, metric string, summary Summary, err error) string {
	if err != nil {
		return VerdictNotRatified
	}
	if metric != "recomputed_work" {   // <-- chokepoint 1
		return VerdictObserved
	}
	...
}

// session/session_phase6_budget.go
func QLT02GateEligibleMetrics() []string {
	return []string{"recomputed_work"}   // <-- chokepoint 2
}
```
Both must accept whatever new metric name(s) the growth-exponent gate introduces (e.g. widen the equality check to a set membership check), with `TestDemoteHasExactlyOnePromotionPassthrough` re-derived to still assert exactly one `return VerdictBlocking` statement.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|---------------|--------|
| Intraprocedural-only loan liveness (Phase 3, M001) | Interprocedural liveness via signature-mediated summaries | This phase (Phase 08) | Extends the same worklist/lattice machinery; no new dataflow engine, per S-006's validated finding |
| Naive per-call-site summary recomputation (considered, rejected by S-006) | Memoized, RPO-ordered, callee-before-caller derivation | Decided at S-006 (pre-phase spike), locked D-08-03 | 76x-and-widening cost separation at 512 functions; the naive arm exhausts a budget at 32 functions on a chain with zero sharing |

**Deprecated/outdated:**
- Backward-transfer-function placement for interprocedural facts: considered and explicitly rejected by S-006 iteration 2 as a silent-failure design (passes accept tests, never refuses).
- A persistent cross-run summary cache as an assumed win: explicitly not to be assumed until QLT-06 gates it (Phase 11); S-006 measured 92%/43%/100% invalidation fallout on realistic/mean/chain shapes.

## Assumptions Log

All claims above were either verified against the shipped tree in this research session (quoted verbatim with file:line) or are 08-CONTEXT.md's own already-independently-verified locked decisions (which the orchestrator re-verified against the tree before locking, per that document's own provenance statement). No new `[ASSUMED]` claims were introduced by this research pass.

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| — | (none) | — | — |

**This table is empty.** All claims in this research were verified this session (direct file reads, quoted verbatim) or inherited from 08-CONTEXT.md's own tree-verified decision record.

## Open Questions

1. **Whether the summary-bit storage is a method on `buildCallSignatureTable` or a sibling function it calls**
   - What we know: D-08-04/D-08-06 lock *where* (in-memory, in `check`, unexported) and *when* (post-body, RPO order) but not the exact Go shape.
   - What's unclear: Whether extending the existing `callSignatureTable` struct's entries (widening `core.FunctionSignature`-adjacent storage) or adding a parallel `map[string]interproceduralSummary` is cleaner given the immutability guarantee `callSignatureTable` currently has by construction (no setter after `buildCallSignatureTable` returns).
   - Recommendation: planner's discretion per CONTEXT.md; either is compatible with all locked decisions. A parallel map keyed identically to `callSignatureTable.entries` most cleanly preserves the "immutable once built" property without touching the existing struct's field list.

2. **Exact corpus size ladder for the growth-exponent fit**
   - What we know: D-08-35 names required shapes (chain, diamond, dense/parser-shaped, forward); the spike used up to 512 functions.
   - What's unclear: Whether the production gate's size ladder should match the spike's exactly or can be smaller if the fitted exponent stabilizes earlier (cheaper CI cost per D-08-36's "not the default edit loop" framing).
   - Recommendation: start with the spike's ladder (up to 512 functions) at the mid-phase gate; the ratio-stability tripwire (S vs 4S at ≤15%) gives an empirical signal for whether a smaller ladder already stabilizes.

## Environment Availability

Skipped — this phase has no external dependencies beyond the existing Go toolchain already used throughout the project (no new CLIs, services, or databases).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go's built-in `testing` package (stdlib) |
| Config file | none — standard `go test` |
| Quick run command | `go test ./internal/compiler/check/... -run TestLoanLiveness` (or the specific new test names) |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|--------------------|--------------|
| OWN-06 | refuse-case twin (Pattern A, ReturnsBorrowOfParam) | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessTwinPatternA` | ❌ Wave 2 |
| OWN-06 | refuse-case twin (Pattern B, UsesParam) | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessTwinPatternB` | ❌ Wave 2 |
| OWN-06 | depth-≥2 relay chain (transitive summary) | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessRelayDepth2` | ❌ Wave 2 |
| OWN-06 | `relay_escort_witness.lang` flips clean→refused | unit (existing test renamed/re-asserted) | `go test ./internal/compiler/check/... -run TestRelayEscortWitness` | ✅ exists (assertion must flip) |
| OWN-06 | negative control (Fails/Foreign.* don't affect verdict) | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessNegativeControl` | ❌ Wave 2 |
| OWN-06 | D-07-49 both admission paths fixed | unit (differential) | `go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree` | ❌ Wave 1 |
| OWN-06 | fail-closed iteration bound refusal (seeded seam) | unit + mutation-kill | `go test ./internal/compiler/check/... -run TestLoanLivenessBoundMutationKilled` | ❌ Wave 3 |
| OWN-06 | native-recursion→explicit-stack conversion doesn't change behavior | unit (existing tests re-run) | `go test ./internal/compiler/check/... -run TestLoanLivenessFixpoint` | ✅ exists (`check_test.go:146`), covers post-conversion too |
| OWN-06 | memo never persisted, rebuilt per invocation | unit | `go test ./internal/compiler/check/... -run TestSummaryMemoNeverPersisted` | ❌ Wave 2 |
| OWN-06 | program-order invariant pinned | unit | `go test ./internal/compiler/check/... -run TestSummaryDerivationRequiresProgramOrder` | ❌ Wave 2 (pattern exists: `TestReborrowChainWorkIsLinear`, `check_test.go:737`) |
| EFF-02 | growth exponent ≤1.2 vs operation count, all required shapes | integration | `go test ./internal/compiler/session/... -run TestQLT02InterproceduralGrowthExponent` | ❌ Wave 4 |
| EFF-02 | chokepoint widening preserves exactly-one-promotion-passthrough | unit | `go test ./internal/compiler/measure/... -run TestDemoteHasExactlyOnePromotionPassthrough` | ✅ exists, must be re-derived not just re-run |
| EFF-02 | manifest row ratified with machine_id | integration | `go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest` | ✅ existing framework, new row |

### Sampling Rate
- **Per task commit:** `go test ./internal/compiler/check/...`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd-verify-work`; additionally the mid-phase gate requires criterion 1's corpus and criterion 3's cost measurement adjudicated from code-level evidence before the liveness law is declared final.

### Wave 0 Gaps
None — existing test infrastructure (`check_test.go`'s established patterns for work-ratio assertions, fault-injection seams, and corpus-driven admission tests; `measure`/`session`'s existing manifest and demotion test scaffolding) covers all phase requirements. No new framework or shared fixture setup is needed before implementation begins.

## Security Domain

`security_enforcement: true`, `security_asvs_level: 1` (`.planning/config.json`). This phase is compiler-internals work with no network, storage, authentication, or user-session surface — the ASVS categories below are assessed for applicability, not blanket-included.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-------------------|
| V2 Authentication | No | No auth surface in this phase |
| V3 Session Management | No | No session surface |
| V4 Access Control | No | No access-control surface |
| V5 Input Validation | Yes (narrowly) | The fail-closed iteration bound IS the input-validation control for this phase — a program that could otherwise drive the worklist unboundedly is refused, not allowed to hang. Already-existing `.lang` source-shape validation (parser, checker) is unchanged by this phase. |
| V6 Cryptography | No | No cryptographic material in this phase (digests like `ClosureDigest` are integrity/staleness hashes, not authenticity — already the subject of D-07-13, not touched here) |

### Known Threat Patterns for {stack}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|----------------------|
| Unbounded worklist / analysis DoS (compiler hang on adversarial input) | Denial of Service | Fail-closed derived iteration bound (D-08-13/D-08-14), already the design this phase implements — the phase's own OWN-06 requirement is the mitigation |
| Native-stack exhaustion via recursive graph traversal | Denial of Service | Explicit-stack iterative traversal (D-08-19a), already locked as a required task |
| Diagnostic identity instability enabling agent-loop confusion (not a classic security threat, but a `Causes`-in-hash integrity concern for the AI-agent consumer) | Tampering (of trust in verifier feedback, not of data) | Never embed tunable numeric values in `Cause`/`Message` (D-08-18); already locked |

No new attack surface (network, file I/O beyond existing compiler CLI, external process invocation) is introduced by this phase.

## Sources

### Primary (HIGH confidence — verified in-tree this session)
- `internal/compiler/check/check.go:493-500,577-609,1049-1225,2979-3030` — signature table, canonicalization pre-pass, backward transfer, fixpoint, D-07-49 defect site (all read and quoted verbatim this session)
- `internal/compiler/check/check_test.go:716-753,2338-2364` — existing work-ratio test pattern, the fixture assertion that must flip
- `internal/compiler/core/core.go:181-282` — `FunctionSignature`/`ParameterContract`/`ReturnContract`, confirming D-08-01's falsification
- `internal/compiler/diagnostic/diagnostic.go:1-30,90-115` — `Span`/`Cause`/`Diagnostic`/identity hash construction
- `internal/compiler/measure/statistics.go:100-140` — `Demote`'s hardcoded chokepoint
- `internal/compiler/session/session_phase6_budget.go:40-70` — `QLT02GateEligibleMetrics()`'s hardcoded chokepoint
- `.planning/spikes/006-interprocedural-liveness-cost-scaling/README.md` — full spike investigation trail (VALIDATED verdict, all iterations)
- `.planning/config.json` — workflow flags (`nyquist_validation: true`, `security_enforcement: true`, `security_asvs_level: 1`)

### Secondary (MEDIUM confidence)
- `.planning/phases/08-interprocedural-loan-liveness-in-check/08-CONTEXT.md` — the 40-decision locked record, itself the product of orchestrator re-verification against the tree (three corrections recorded therein)
- `.planning/REQUIREMENTS.md`, `.planning/STATE.md` — requirement text and project history

### Tertiary (LOW confidence)
- None used — this phase's research surface is entirely covered by in-tree verification and the already-verified CONTEXT.md.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no new dependencies; all extension points verified in tree
- Architecture: HIGH — mechanism placement, ordering, and memoization scheme independently spike-validated (S-006) and re-verified against the tree by both this session and 08-CONTEXT.md's orchestrator pass
- Pitfalls: HIGH — every listed pitfall is either a proven failure mode from S-006's investigation trail or a structurally-verified chokepoint in the current tree, not a hypothetical

**Research date:** 2026-09-09
**Valid until:** Stable until the tree changes underneath these line numbers (compiler-internals research; treat as valid through Phase 08's execution, re-verify line numbers if execution is delayed past other intervening phase work landing in `check.go`)

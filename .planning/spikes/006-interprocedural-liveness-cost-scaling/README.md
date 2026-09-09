---
spike: 006
idea: ownership-kernel
name: interprocedural-liveness-cost-scaling
type: comparison
validates: "Given call graphs of increasing size and sharing, when loan liveness is derived from callee summaries, then a memoized derivation stays linear in program size while an unmemoized one does not, and both answer exactly as a context-sensitive expansion oracle does"
verdict: VALIDATED
related: [002, 003, 004]
tags: [ownership, liveness, interprocedural, summaries, cost, scaling, caching, oracle]
---

# Spike 006: Interprocedural liveness cost-scaling

Pre-phase spike **S-006**, recorded in `ROADMAP.md` under "Pre-Phase Spikes".
It is a **hard entry gate on the *planning* of Phase 08**, carries no M002
requirement, and produces throwaway workbench code.

## What this validates

The gate question, verbatim from the roadmap:

> Does summary-based liveness stay linear/sub-quadratic in call-graph size, or
> must a memoized summary cache be designed in from the start?

Given call graphs of increasing size and increasing callee sharing, when loan
liveness at a call site is decided from callee **summaries** only (OWN-06:
never by re-walking callee bodies), then the cost of deriving those summaries
should scale with the number of functions rather than with the number of paths
through the call graph — and every mechanism priced must first be shown to give
the same answer as an oracle that shares nothing with it.

It does **not** attempt: `corevalidate`'s independent peer derivation (OWN-07),
D-03-02 closure (OWN-08), recursion or cycle policy (that is S-007), Nyquist
fold-in cost (S-008), or anything native. It is a cost probe with a correctness
floor, not a design for Phase 08's checker.

## Research

Summary-based interprocedural analysis is a solved problem with a named
algorithm and a named failure mode, and both bear directly on this gate.

- **IFDS/tabulation (Reps–Horwitz–Sagiv)** is precisely "compute a procedure
  summary once, reuse it at every call site". Its whole point is that the
  summary edge between a call node and its return site is *tabulated* — the
  algorithm is defined by its memo table, not merely optimized by one. Its
  complexity bound is stated per call-graph edge, which is only meaningful
  because each procedure is summarized once.
- **Call-graph precision dominates cost.** Published measurements of real
  codebases find call sites with five or more targets routinely, and some with
  hundreds. Any per-call-site recomputation multiplies by that fan-out at every
  level.
- **Polonius is the cautionary precedent the roadmap already names.** The
  original Datalog formulation was "considerably slower than NLL" on some
  programs and had no path to stabilization; the re-engineered version scales
  because of *how* the facts are computed, not because the facts changed. Worst
  observed regression on the revised implementation is 2–3x.

| Approach | What it costs | Disposition |
|---|---|---|
| Re-derive a callee summary at every consultation | Number of consultations, which compounds per level | **priced arm** (`recompute`) |
| Re-derive per caller, remembering within one caller's analysis | Number of distinct call-graph paths | **priced arm** (`recompute+scratch`) |
| Derive once per function, remember for the run (IFDS tabulation) | Number of functions + edges | **priced arm** (`memoized`) |
| Inline everything, then analyze one monolithic CFG | Exponential in sharing | **independent oracle only** |
| Persistent cross-run summary cache keyed by content digest | Depends on invalidation fallout | measured here, not built |

Sources: [IFDS via graph reachability](https://research.cs.wisc.edu/wpis/papers/popl95.pdf) ·
[On-demand data-flow analysis](https://arxiv.org/pdf/2001.11070) ·
[Value contexts in Soot](https://arxiv.org/pdf/1304.6274) ·
[Polonius alpha on nightly](https://blog.rust-lang.org/2026/08/04/enabling-polonius-alpha-on-nightly/) ·
[Polonius current status](https://rust-lang.github.io/polonius/current_status.html)

## Gates

1. Every priced mechanism agrees with the expansion oracle on the hand-written
   fixtures and on generated graphs of all five shapes, before any cost number
   is quoted.
2. Hand-stated refusals — written by hand, not produced by either mechanism —
   match on all six fixtures.
3. Cost is swept across shapes with and without callee sharing, and the growth
   exponent is fitted against program size, not only function count.
4. Both fault seams (a dropped returned-borrow summary; an ID-keyed cache that
   never invalidates callers) produce a stable, small disagreement.
5. A cycle is refused fail-closed by the order, both mechanisms, and the oracle.
6. The cost of the cache itself — invalidation fallout from a one-function edit
   — is measured, not assumed.

## How to run

From this directory:

```sh
go test ./...
go test -race ./...
go test -cover ./...
go run ./cmd/iplive
go run ./cmd/iplive -inject-summary-drop -pretty=false
go run ./cmd/iplive -inject-stale-cache -pretty=false
```

Both injected-fault commands are expected to exit unsuccessfully after
reporting disagreements.

## What to expect

The ordinary run emits a compact JSON decision report and exits successfully:

- `"verdict": "agreed"` with no `disagreements` — 11 agreement cases (6
  fixtures + 5 generated shapes) x 2 mechanisms, each checked against the
  expansion oracle;
- six cost series (`chain`, `tree`, `diamond`, `dense`, `parser`, `forward`),
  each with three priced arms and a fitted growth exponent;
- a body-order probe and four cache-invalidation probes.

`-inject-summary-drop` exits 1 with 62 disagreements, the first being the
returned-borrow refusal `caller|caller:then:move|l` that the mechanism no
longer makes. `-inject-stale-cache` exits 1 with exactly 2 disagreements, both
from the stale-edit probe: after a leaf edit, the ID-keyed cache still reports
`caller|caller:move|l`, which the oracle no longer does.

`go test -cover` reports 92.1% of statements.

## Observability

Every number the report quotes is a **deterministic work counter**, not a
timing: one unit per transfer-function evaluation, one per worklist
reinsertion, one per summary consultation, one per operation visited during a
derivation. Wall-clock nanoseconds are recorded in a separate field and are
never used to classify a mechanism. Every conflict, operation, loan, block and
context carries a stable identity, and the oracle projects its context-
sensitive answer back onto source identities before comparing.

## Investigation trail

### Iteration 1 — make the summary observable before pricing it

A cost comparison between two mechanisms is worthless if either is wrong, and
"is it wrong" needs an answer the mechanism cannot supply. The workbench
therefore models exactly two summary bits — `UsesParam` (the call reads through
the argument loan) and `ReturnsBorrowOfParam` (the result aliases it) — and
gives each one a caller-side pattern whose *admission decision* flips on it:

- `borrow; call; move; use(result)` refuses iff the callee returns a borrow;
- `borrow; move; call` refuses iff the callee uses its parameter.

Both bits are transitive through relays, which is what makes a summary depend
on the whole reachable subgraph rather than on one body — the cost question in
miniature.

### Iteration 2 — a first design that put the interprocedural fact in the wrong place

The first transfer function handled a call by mapping *result* liveness back to
*argument* liveness while walking backward. It was wrong, and the fixtures said
so immediately: in `borrow; call; move; use(result)` the backward walk reaches
the `move` **before** the call that establishes the alias, so the loan was not
yet live and the refusal never fired.

The fix is a summary-consuming **canonicalization pre-pass** in program order: a
call whose callee returns a borrow of its parameter binds its result to *the
same loan identity* as the argument; a call returning an owned value binds its
result to a fresh identity no borrow ever creates. After that pass the backward
transfer has almost no interprocedural residue left — only `UsesParam`. This is
the same move `check` already makes with `derivePlaceLoans`, and it is what
made the whole per-function analysis stay ordinary spike-002-shaped dataflow.

### Iteration 3 — an oracle that shares nothing

The oracle inlines every call at its own call site into one monolithic CFG and
then runs a *naive round-robin* liveness over it — no worklist, no
predecessors, no summaries, no canonicalization special case (after inlining
there are no calls left to canonicalize). It is context-sensitive by
construction, so its per-context answers are unioned before comparison.

That union is the right ground truth only where a context-insensitive analyzer
could reach it, so the comparison scope is stated rather than assumed:
**conflicts project exactly**, and **local liveness projects exactly for loans
born in the same function**. A parameter loan is deliberately excluded — in a
context its liveness is decided by the caller's later uses, which a body-blind
analyzer cannot and must not see. All eleven agreement cases pass for all
mechanisms with that scope.

### Iteration 4 — the first cost sweep, and a result that was too strong

The naive `recompute` arm exhausted a four-million-unit budget at **32
functions on a plain chain** — a call graph with no sharing at all. That is
exponential in call *depth*, and it happens because a single derivation
consults each callee more than once (two call sites to the same callee, plus a
confirming fixpoint pass), so the multiplier compounds per level.

That is a true measurement but an unfair headline: it prices repetition, not
sharing. So a third, deliberately charitable arm was added —
`recompute+scratch`, which remembers a derivation for exactly as long as one
caller's analysis and forgets it afterwards. Its cost is the number of distinct
call-graph *paths*. Pricing all three separates "exponential because of
repetition" from "quadratic because of sharing", and the charitable arm is the
one the gate's verdict rests on.

### Iteration 5 — following the derivation into its own worst case

The memoized arm derives each summary by re-scanning a body until stable, which
raises an obvious worry: a body that threads one borrow through *k* successive
calls could need one pass per link. A dedicated `forward` shape (one caller,
`k` chained calls, star-shaped call graph) says it does not: **4.0 work units
per operation, flat from k=8 to k=512.** Two passes suffice, because a body in
program order propagates the whole chain in its first pass.

Listing the *same* chain in reverse order costs **12.4 → 767.0 work units per
operation** over the same range — a 192x penalty at k=512, and quadratic in
body length. So the linear result carries a precondition, and it is an
invariant Phase 08 inherits rather than a property of the mechanism:
**summary derivation must consume operations in program order.** Real
`core.LinearOperation` sequences are in program order today; any future pass
that reorders operations before summary derivation hands the bound back. Both
halves are pinned by tests that fail if either changes.

### Iteration 6 — pricing the cache itself

A cache is not free, and the honest second half of the gate is what one edit
costs once a cache exists. Editing a single leaf and evicting it plus every
transitive caller (what Phase 07's `ClosureDigest` chain forces) throws away:

| Shape | Worst edit evicts | Mean |
|---|---|---|
| chain (128 fns) | 128/128 (100%) | 64.5 |
| tree (128 fns) | 8/128 (6%) | 6.1 |
| dense (127 fns) | 125/127 (98%) | 63.0 |
| parser-shaped (128 fns) | 118/128 (92%) | 55.1 |

### Iteration 7 — proving the harness can fail

Two seams, each armed from the CLI and each independently mutation-killed:

1. **Summary drop** — the derivation forgets that a forwarded parameter comes
   back out through the return. 62 disagreements, including missing refusals:
   the mechanism under-approximates liveness and *admits* a move-while-borrowed.
2. **Stale ID-keyed cache** — the cache is keyed by function ID alone, so a
   callee edit never invalidates its callers. Invisible on a cold run by
   construction, so the harness replays the edit explicitly: warm the cache,
   edit the leaf, invalidate under the armed policy, re-check. Two
   disagreements; the caller keeps refusing a move the edited program permits.

Both are exactly the failure classes QLT-06 ("no interprocedural fact is
cacheable until a callee-changes-invalidates-caller regression test gates it")
and OWN-06 are written against.

## Results

**Verdict: VALIDATED — and the gate's answer is the second branch. A memoized
summary cache must be designed into Phase 08 from the start.**

Growth exponents are fitted against program operation count (the honest axis
for a claim about an analyzer; a corpus whose leaf-to-relay ratio drifts with
size moves the function-count fit without any mechanism changing).

| Shape | Largest measured | `memoized` | `recompute+scratch` | `recompute` |
|---|---|---|---|---|
| chain | 512 fns / 1,705 ops | **linear** (1.02) | sub-quadratic (1.99) | budget exhausted at 32 fns |
| tree | 255 fns / 616 ops | **linear** (1.03) | sub-quadratic (1.32) | sub-quadratic (1.95) |
| diamond | 29 fns / 113 ops | **linear** (1.04) | sub-quadratic (1.91) | budget exhausted at 25 fns |
| dense | 22 fns / 98 ops | **linear** (1.11) | super-quadratic (2.03) | worse than cubic (6.71) |
| parser-shaped | 512 fns / 2,328 ops | **linear** (1.06) | sub-quadratic (2.00) | budget exhausted at 32 fns |
| forward (star) | 512 fns / 515 ops | **linear** (1.04) | linear (1.04) | linear |

Concrete separation on the realistic shape: at 512 functions / 1,427 call edges,
`memoized` costs **13,441 work units (2.3 ms)** and derives each summary exactly
once (469 derivations across 512 functions -- every function that is a callee,
none of them twice). The charitable
`recompute+scratch` arm costs **1,019,403 units (21.9 ms) — 76x more — and the
multiplier doubles with every doubling of program size**. The naive arm never
finished.

**Findings for Phase 08:**

1. **The cache is not an optimization, it is the mechanism.** Memoized
   derivation is linear in program size on every shape tried, at a stable
   ~5 work units per operation across a 100x size range. Without it, cost is
   governed by call-graph *paths*, and paths are exponential in exactly the
   graphs real code produces (shared leaf utilities: `diamond`, `dense`,
   `parser`). This matches IFDS: tabulation is the algorithm's definition, not
   a tuning knob.
2. **Reverse postorder over the proven-acyclic call graph is sufficient.**
   Phase 07 already refuses cycles and already computes this order
   (`callgraph.Order`), so summaries can be derived callee-before-caller in one
   pass with no call-graph-level fixpoint at all. The workbench's memo is lazy
   rather than RPO-ordered and lands on the same count — one derivation per
   function — which means the ordering choice is free and the acyclicity
   guarantee is what is load-bearing.
3. **`loanLivenessFixpoint` extends directly, once summaries exist.** The
   interprocedural fact enters through a program-order canonicalization
   pre-pass plus a single extra clause in the transfer function. The backward
   worklist itself is unchanged from spike 002's shape. The new subsystem
   Phase 08 must build is the summary table and its invalidation, not a new
   dataflow engine.
4. **Program order is a load-bearing invariant.** Derivation is constant-pass
   only because bodies arrive in program order; reversed, the same derivation
   is quadratic in body length (192x at k=512). Phase 08 should state this
   where the derivation lives.
5. **Do not assume a persistent cross-run summary cache pays off.** On the
   realistic shape a single leaf edit invalidates 92% of the cache worst case
   and 43% on average; on a chain, 100%. Within-run memoization is mandatory;
   cross-run caching is a separate claim that must be measured against real
   edit patterns before EFF-02 quotes it — and QLT-06 already refuses to let an
   interprocedural fact be marked cacheable before a
   callee-changes-invalidates-caller test gates it.
6. **Feedback-budget signal for EFF-02.** At 512 functions / 2,328 operations
   the whole memoized interprocedural pass costs 2.3 ms in this workbench. The
   cost model to declare a bound against is `work ≈ 5 x operations`, not
   anything shaped like the call graph. This is a workbench figure on model
   bodies and must not be quoted as a production budget.

**Surprises.**

- The naive arm dies on a *chain* — no sharing required. Repetition alone
  compounds per level, which is why the charitable arm had to exist before any
  headline could be quoted.
- Memoized cost tracks operation count, not call-graph shape. The `dense` shape
  looked mildly superlinear (1.50) against function count and is 1.11 against
  operation count; the difference is entirely the corpus's leaf-to-relay ratio
  drifting with size. Fitting only against function count would have produced a
  wrong finding.
- The correctness fix in iteration 2 (canonicalize in program order) turned out
  to be the same assumption iteration 5 measured as a cost invariant. One
  assumption, two consequences, both now pinned by tests.

**Bounded claims / expected escape.** This workbench prices mechanisms and
catches *observable* drift. A summary defect in a function that no caller ever
probes with a move is semantically inert on this corpus, and no oracle here can
catch it — `TestInertSummaryDefectEscapesByDesign` states that escape
explicitly and fails if the corpus ever silently starts covering it. Also out
of scope by construction: recursion and cycle policy (S-007), the peer
re-derivation (OWN-07), field-sensitive places, higher-ranked origins, and any
production performance claim. Both the mechanism and the oracle assume bodies
in program order; neither would detect a mis-ordered body, and that is a
documented boundary, not a tested one.

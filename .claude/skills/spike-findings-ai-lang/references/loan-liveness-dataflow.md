# Loan Liveness Dataflow

How a loan's end is inferred — intraprocedurally over a CFG (spike 002) and
across call boundaries from callee summaries (spike 006). These are one
subsystem: spike 006's backward worklist is spike 002's, unchanged.

## Requirements

From the `ownership-kernel` idea (MANIFEST.md):

- Infer local loan endings across control flow without requiring source-level
  end markers on ordinary paths.
- Keep the production-shaped analysis polynomial while retaining bounded path
  expansion only as an independent oracle.
- Treat explicit source loan endings as outside the CFG-last-use slice rather
  than silently mixing manual and inferred policies.
- Memoize interprocedural summaries within a run: derive each function's summary
  once over the proven-acyclic call graph, never per call site. Cost must be
  declared against program operation count, not call-graph shape.
- Derive summaries from operations in program order. Any pass that reorders
  operations before summary derivation must restate the cost bound.
- Do not mark an interprocedural fact cacheable across runs before the
  invalidation fallout of a real callee edit is measured.

## How to Build It

### Intraprocedural: backward loan liveness (spike 002)

**Carry a finite set of live loan identities at each block entry and exit, and
run a reverse worklist to a fixpoint.** On the 1,501-block / 2,000-edge scale
case this converged in exactly **1,501 transfer evaluations** — one per block —
with no path enumeration.

**A loan's end is not always a program point. It can be a control-flow edge.**
Materialize two endpoint kinds:

- `point:then:0:view` — after an operation, when the final use is in a block
- `edge:entry:else:view` — on an edge, when only one successor keeps the loan live

Loop-carried uses stay live through the back edge and end on the **loop-exit
edge**. This is the consequential finding: the typed IR and machine diagnostics
need **stable edge identities** even though ordinary source contains no explicit
loan-ending syntax.

**Encode endpoint coordinates losslessly.** The first schema omitted operation
index zero, which silently made a valid coordinate unrepresentable in JSON.

### Interprocedural: summaries over the call graph (spike 006)

**Model the interprocedural fact as two summary bits, each with a caller-side
pattern whose admission decision flips on it:**

- `UsesParam` — the call reads through the argument loan.
  Discriminator: `borrow; move; call` refuses iff true.
- `ReturnsBorrowOfParam` — the result aliases the argument.
  Discriminator: `borrow; call; move; use(result)` refuses iff true.

Both bits are transitive through relays. A summary therefore depends on the
whole reachable subgraph — which is exactly the cost question.

**Put the interprocedural fact in a program-order canonicalization pre-pass, not
in the backward transfer.** Spike 006 iteration 2 got this wrong first: mapping
result liveness back to argument liveness during the backward walk reaches the
`move` *before* the call that establishes the alias, so the refusal never fires.
The fix:

- a call whose callee returns a borrow of its parameter binds its result to **the
  same loan identity** as the argument;
- a call returning an owned value binds its result to a **fresh identity no
  borrow ever creates**.

After that pass the backward transfer has almost no interprocedural residue —
only `UsesParam` — and stays spike-002-shaped. (This is the same move `check`
already makes with `derivePlaceLoans`.)

**Memoize. The cache is the mechanism, not an optimization.** Derive each
function's summary once and reuse it at every call site (IFDS tabulation).
Reverse postorder over the proven-acyclic call graph is sufficient — Phase 07
already refuses cycles and already computes `callgraph.Order`, so summaries
derive callee-before-caller in one pass with **no call-graph-level fixpoint at
all**. The workbench's lazy memo lands on the same count (469 derivations across
512 functions — every callee once, none twice), which means the ordering choice
is free and the **acyclicity guarantee is what is load-bearing**.

**Derive summaries in program order.** A body in program order propagates a
whole k-link borrow chain in its first pass: **4.0 work units per operation, flat
from k=8 to k=512**. The same chain listed in reverse costs **12.4 → 767.0 units
per operation — a 192x penalty at k=512, quadratic in body length.** State this
invariant where the derivation lives. Both halves are pinned by tests.

## What to Avoid

- **Lexical loan extent.** Rejected as anything but a control: it rejects safe
  branch-local release and adds ceremony.
- **Exhaustive path expansion in the analyzer.** Exponential, and loops need a
  bound. It belongs in the bounded oracle only.
- **Full origin/loan relations before local last use is settled**, and
  **general symbolic path conditions** — solver cost and diagnostic instability
  exceed the question.
- **Mixing explicit `end_loan` into the CFG-last-use slice.** Spike 002 *rejects*
  source-level `end_loan` operations so manual and inferred policies cannot
  jointly determine the result. A later surface policy may offer scoped escape
  hatches; ordinary source should not require one.
- **Letting the imported linear normalizer repair a missing CFG endpoint.** The
  path oracle needed a deliberately late terminal guard, or the fault injection
  would have been silently healed.
- **Quoting a cost number before the arm agrees with the oracle.** Gate 1 of
  spike 006: every priced mechanism agrees on every fixture and all five
  generated shapes *first*.
- **Pricing only the weakest opponent.** The naive `recompute` arm dies on a
  plain *chain* — no sharing at all — because one derivation consults each callee
  more than once and the multiplier compounds per level. That is a true
  measurement but an unfair headline. A charitable `recompute+scratch` arm
  (remembers within one caller's analysis) had to exist before any verdict
  rested on the comparison.
- **Fitting growth against function count.** The `dense` shape looks superlinear
  (1.50) against function count and is 1.11 against **operation count**; the
  difference is entirely the corpus's leaf-to-relay ratio drifting with size.
  Fitting the structural axis alone produces a wrong finding.
- **Assuming a cross-run summary cache pays off.** See Constraints.

## Constraints

**Cost, fitted against program operation count** (spike 006, largest measured):

| Shape | Largest | `memoized` | `recompute+scratch` | `recompute` |
|---|---|---|---|---|
| chain | 512 fns / 1,705 ops | **linear** (1.02) | sub-quadratic (1.99) | budget exhausted at 32 fns |
| tree | 255 fns / 616 ops | **linear** (1.03) | sub-quadratic (1.32) | sub-quadratic (1.95) |
| diamond | 29 fns / 113 ops | **linear** (1.04) | sub-quadratic (1.91) | budget exhausted at 25 fns |
| dense | 22 fns / 98 ops | **linear** (1.11) | super-quadratic (2.03) | worse than cubic (6.71) |
| parser-shaped | 512 fns / 2,328 ops | **linear** (1.06) | sub-quadratic (2.00) | budget exhausted at 32 fns |
| forward (star) | 512 fns / 515 ops | **linear** (1.04) | linear (1.04) | linear |

At 512 functions / 1,427 call edges: `memoized` costs **13,441 work units
(2.3 ms)**; `recompute+scratch` costs **1,019,403 units (21.9 ms) — 76x more,
and the multiplier doubles with every doubling of program size**. Cost model to
declare a bound against: **`work ≈ 5 x operations`**, not anything shaped like
the call graph.

**Cross-run cache invalidation fallout** — one leaf edit, evicting it plus every
transitive caller (what Phase 07's `ClosureDigest` chain forces):

| Shape | Worst edit evicts | Mean |
|---|---|---|
| chain (128 fns) | 128/128 (100%) | 64.5 |
| tree (128 fns) | 8/128 (6%) | 6.1 |
| dense (127 fns) | 125/127 (98%) | 63.0 |
| parser-shaped (128 fns) | 118/128 (92%) | 55.1 |

Within-run memoization is mandatory. **Cross-run caching is a separate claim**
that must be measured against real edit patterns before EFF-02 quotes it.

**Scope limits:**

- Spike 002's path evidence is exhaustive only within three visits per block.
- Spike 002 models **one static loan definition per CFG** — not fresh dynamic
  loan identities created per loop iteration.
- Spike 006's oracle comparison scope is stated, not assumed: **conflicts project
  exactly**, and **local liveness projects exactly for loans born in the same
  function**. A parameter loan is deliberately excluded — in a context its
  liveness is decided by the caller's later uses, which a body-blind analyzer
  cannot and must not see.
- Out of scope by construction in spike 006: recursion and cycle policy (S-007),
  peer re-derivation (OWN-07), field-sensitive places, higher-ranked origins, any
  production performance claim.
- **Both the mechanism and the oracle assume bodies in program order; neither
  would detect a mis-ordered body.** Documented boundary, not a tested one.
- All work numbers are **deterministic work counters** (one unit per transfer
  evaluation, worklist reinsertion, summary consultation, operation visited).
  Wall-clock nanoseconds live in a separate field and never classify a mechanism.

## Verdicts

Spike 002: **VALIDATED** — 9/9 fixtures, 100/100 exhaustive diamond products,
500/500 seeded randomized/metamorphic trials, 93.6% coverage; omitted edge
endpoints detected after 7 generated programs.

Spike 006: **VALIDATED, and the gate answers the second branch — a memoized
summary cache must be designed into Phase 08 from the start.** 11 agreement
cases x 2 mechanisms; 92.1% coverage; `-inject-summary-drop` yields 62
disagreements, `-inject-stale-cache` exactly 2.

## Origin

Synthesized from spikes: 002, 006
Source files available in: `sources/002-cfg-edge-last-use/`,
`sources/006-interprocedural-liveness-cost-scaling/`

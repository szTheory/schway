# Phase 07: Calls, Signatures, and Call-Graph Refusal - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-08
**Phase:** 07-calls-signatures-and-call-graph-refusal
**Areas discussed:** Call argument shape, Summary scope + schema, Cycle refusal shape, Stage 0 peer count
**Mode:** advisor (four parallel `gsd-advisor-researcher` agents, `minimal_decisive` calibration, technical framing)

**Developer's standing instruction, given at area selection and reaffirmed at
decision time:** fan out across all relevant stakeholder-role lenses, weigh
pros/cons/tradeoffs/anti-patterns/best-practices/footguns/lessons-learned, draw
on other products and ecosystems, run an adversarial pass, then synthesize a
one-shot recommendation for each point. On the second pass the developer
declined per-area selection and directed that the synthesized recommendations be
adopted as written ("i gave u general fan out prompt just fan out follow ur
recs").

---

## Call argument shape

| Option | Description | Selected |
|--------|-------------|----------|
| Arity-1 name argument | `let r = f(x)` where `x` is any in-scope binding name. No borrow/take/literal/nested-call expression at the call site; convention implicit from the callee signature; positional. Widens one predicate at `check.go:1316`. | ✓ |
| Arity-N positional | Widen the checker to N arguments now; unlocks Phase 11's two-`restrict` case but forces multi-argument loan-interaction rules into the phase that first makes `OpCall` real. | |

**Choice:** Arity-1 name argument (D-07-01 … D-07-07).
**Notes:** Verified in the shipped tree during discussion that
`testdata/phase3/shared_shared_accept.lang` already binds loans to names, so a
divergent cross-call loan state is expressible with **no new grammar** — the
finding that made arity-1 sufficient rather than merely cheap. Adversarial pass
produced two consequences accepted in writing rather than left to be discovered:
Phase 11's `restrict` is single-parameter-only and must scope its gate
accordingly (two-parameter case → M003), and Phase 08's gate wording must say
"call-graph depth ≥2" rather than "composition depth", since the language now has
A-normal-form calls. M001's D-02 (labeled call sites) recorded as **deferred, not
revoked** — the shipped tree is already positional, and at arity 1 labels are
pure ceremony.

---

## Summary scope + schema

| Option | Description | Selected |
|--------|-------------|----------|
| `/1` bump, arity-ready schema | `lang.interface/1` with required total fields; `/0` frozen behind a pinned legacy decoder; `Parameters []ParameterContract` while the checker still admits arity 1. | ✓ |
| `/1` bump, `Parameter` singular | Same required-field set and bump, but signature keeps a single `Parameter` matching the checker rule exactly; a second break (and a `/2`) required when arity widens. | |
| Additive on frozen `/0` | Append `omitempty` fields following the `LinearOperation` Phase 4 precedent; no literal-site churn, but a pre-Stage-0 summary decodes with Go zero values and admission proceeds on absent facts. | |

**Choice:** `/1` bump with arity-ready schema (D-07-08 … D-07-13).
**Notes:** The two researchers **conflicted** here — the call-shape researcher
argued for keeping `Parameter` singular (no schema change at all), the schema
researcher for `Parameters []ParameterContract` with a `/1` bump. Reconciled in
synthesis: the bump is forced by *required fields* regardless of arity, so the
arity-ready shape is free to take now with one producer rather than later across
five consumers — schema capacity is cheap, the checker rule is where cost lives,
and the checker rule stays tight. Decisive argument against additive growth: Go
zero values make `mode: ""` indistinguishable from "the producer never knew about
ownership", which is fail-open at exactly the boundary this phase exists to make
fail-closed. `ClosureDigest` (Merkle chain over callee summary digests, OCaml
`.cmi` / rustc SVH shape) added as the only thing that closes the
stale-but-self-consistent hole. Call-graph edges deliberately **excluded** from
the summary — an edge list would re-import the rejected Design A.

---

## Cycle refusal shape

| Option | Description | Selected |
|--------|-------------|----------|
| Adopt all five verdicts | (1) `check` refuses over its own completed in-memory `core.Program` before returning it. (2) One code `core.call_graph_cycle`, path as canonically-rotated bounded `Causes`, cap 32, no `Repairs`. (3) `corevalidate` re-derives, tested only on hand-built synthetic artifacts. (4) Depth unbounded if acyclic. (5) Iterative three-color DFS, roots = all declared functions, returns reverse postorder. | ✓ |
| Adopt, but three distinct codes | As above except direct/mutual/indirect each get their own diagnostic code. | |

**Choice:** All five verdicts as researched (D-07-14 … D-07-19).
**Notes:** Sub-decisions considered and rejected individually — an AST-side graph
built before `OpCall` emission (two node-identity rules that can drift
invisibly, and mutual cycles need a whole-program pass regardless); three
diagnostic codes (triples a versioned API surface for one fact and makes
*classification itself* something two peers can disagree about, against SEM-07's
singular wording); `corevalidate` trusting `check` (a forged core would walk into
`interp`/`cgen` and hang); a declared max call-graph depth (depth is not the
hazard; would false-positive on the parser-shaped programs S-007 flags); and
Tarjan/Kosaraju SCC condensation (every SCC is refused, so condensation buys
nothing, yields membership rather than a witness path, and adds lowlink
bookkeeping no test can distinguish correct from subtly-wrong).

**The highest-value finding of the whole discussion** came from this area's
adversarial pass: the gray/black distinction is the entire control, and a
straight deep chain **will not kill** the mutation that replaces the on-stack
gray check with a plain visited check — reverse postorder never revisits on a
chain. The "pathological-depth-but-acyclic" corpus in success criterion 3 must
therefore contain **diamonds and shared leaves**, or that gate is decorative.
Also surfaced: canonical cycle-path rotation is load-bearing rather than
cosmetic, because `Causes` participate in diagnostic ID identity
(`diagnostic.go:102-112`) and an un-rotated path yields N distinct IDs for one
cycle.

---

## Stage 0 peer count

| Option | Description | Selected |
|--------|-------------|----------|
| Two producers in Stage 0 | `originvalidate.BuildInterface` widened (producer) + `corevalidate` re-deriving the summary (peer), same plan, both replay sites, zero-divergence sweep over the M001 Phase 4 corpus plus three seeded faults and a static import-independence test. | ✓ |
| Single producer, peer to Phase 09 | Stage 0 ships one producer; the summary peer defers alongside the liveness peer, held by a `PHASE-07-DEBT.md` entry. | |

**Choice:** Two producers in Stage 0 (D-07-20 … D-07-25).
**Notes:** Two further options were considered and **explicitly rejected as
anti-independence**: a `check`-side producer (it is the admission consumer; a
summary built from its own state correlates with its own decision by
construction — Knight & Leveson found even independently *written* versions
correlate, and a version sharing state correlates fully) and a shared
`callsummary` package (a single implementation two callers agree with cannot
diverge). The strongest honest case for deferring was argued and answered: the
"nothing consumes it yet" objection is backwards, since the peer's value is
fixing the *shape* under adversarial agreement before consumers ossify it, and
the feared harness duplication is illusory — a corpus sweep over existing
functions is the seed of Phase 09's harness, not a competitor. Concrete leak
identified under single-producer trust: `originvalidate.go:427`'s exact-ID lookup
with a silent empty-abilities fallback would yield a self-consistent under-claimed
summary that nothing contradicts across four phases. Three seeded faults
specified, including a **bilateral** one whose gate must FAIL — proving the
comparison is not a value compared to itself.

---

## Claude's Discretion

The developer declined per-area selection and directed that the synthesized
recommendation be adopted for all four areas. Every decision in CONTEXT.md is
Claude's synthesis under that standing instruction, grounded in the four advisor
returns and verified against the shipped tree wherever a claim was checkable.
Remaining planner discretion is enumerated in CONTEXT.md's "Claude's Discretion"
subsection (plan decomposition, identifier and file names inside `callgraph`,
`ForeignReach`'s field typing, fixture naming).

Two cross-cutting items were added by synthesis rather than by any single
researcher: a Phase 07 scope-cut trigger (D-07-26 — the phase had none, while
Phases 08 and 09 both do) and the Phase 12 blindness guard (D-07-28 — enumerate
by `Kind == core.OpCall` across all blocks, provable today with a fixture calling
from both arms of an existing `match`).

## Deferred Ideas

- Arity-N calls and multi-argument loan interaction (schema is already
  arity-ready, so the future change is a checker predicate, not a `/2`).
- Labeled call-site arguments — M001 D-02, deferred not revoked.
- Two-parameter `restrict` evidence in `cgen` — M003 obligation; Phase 11 must
  scope its gate to single-parameter derivation.
- Interprocedural *liveness* re-derivation in `corevalidate` — Phase 09, held by
  a `PHASE-07-DEBT.md` entry written at planning time.
- A maximum call-graph depth / declared depth budget — the interpreter's fixed
  ceiling is SEM-08 in Phase 10; spike S-007 informs whether it is a constant or
  a budget.
- `cgen`'s `len(Functions) != 1` hard fail — Phase 11.
- Call-graph edges in the signature summary — permanently rejected, not deferred.
- Keyed/signed summary digests — out of scope; independent re-derivation is the
  forgery answer.

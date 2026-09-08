# Phase 07: Calls, Signatures, and Call-Graph Refusal - Context

**Gathered:** 2026-09-08
**Status:** Ready for planning

<domain>
## Phase Boundary

A Lang function can call another Lang function. The call is admitted from the
callee's **digest-bound signature summary alone** — no caller admission path
ever opens a callee body — and a program whose calls form a cycle is refused
by name instead of hanging.

Concretely, this phase delivers three staged things in this order:

1. **Stage 0** — widen `core.Interface` / `core.FunctionSignature` into the full
   interprocedural call contract, and land its independent peer re-derivation,
   **before** `core.OpCall` exists. Gated against the entire M001 Phase 4 corpus.
2. **Stage 1** — register `core.OpCall` in `AllOperationKinds()`, deliberately
   breaking all six exhaustive switches at once, and land the `check` producer
   path plus the five consumers.
3. **Stage 2** — a new `internal/compiler/callgraph` package, cycle refusal in
   two independent peers, and the two corpora that make the refusal gate real.

**Requirements:** SEM-04, SEM-05, SEM-06, SEM-07, QLT-08.

**Not this phase** (each has its own phase and its own gate): interprocedural
loan liveness (Phase 08), `corevalidate`'s *liveness* peer and D-03-02 closure
(Phase 09), the interpreter call stack and `originvalidate`/`pathoracle` origin
hops (Phase 10), multi-function C emission (Phase 11), `Result` payloads
(Phase 12), the agent loop (Phase 13).

</domain>

<decisions>
## Implementation Decisions

All four gray areas were researched in advisor mode (four parallel
`gsd-advisor-researcher` agents, `minimal_decisive` calibration) under an
explicit developer mandate to fan out across stakeholder-role lenses, run an
adversarial pass, draw on cross-ecosystem prior art, and synthesize one-shot
recommendations. The developer directed that the synthesized recommendation be
adopted in every case without further selection.

### Call surface — argument shape

- **D-07-01:** A Lang-to-Lang call is a **binding RHS with exactly one argument,
  and that argument is the name of an in-scope binding** (the function's own
  parameter, or any prior `let`). Shape: `let r = f(x)`. This widens the
  existing predicate at `check.go:1316` from *"the argument must equal
  `functionParameterName`"* to *"the argument must resolve to an in-scope
  binding"* — one predicate, not a new grammar.

  No `borrow` / `borrow mut` / `take` / literal / nested-call **expression** may
  appear at a call site. Arguments are **positional**, not labeled.
  — **Reversibility:** reversible — widening to arity-N later changes a checker
  predicate and adds an aliasing rule; the call-site grammar (positional names)
  does not change, and none of the 58 existing corpus programs need edits.

- **D-07-02:** The ownership convention (move / borrow-shared / borrow-exclusive)
  is read **implicitly from the callee's declared signature**, never written at
  the call site. This makes the standing anti-feature *"call-site override of a
  callee's declared ownership convention"* (the C++ overload-resolution
  cautionary case) **ungrammatical rather than merely rejected** — there is no
  syntax in which to express the override, so no diagnostic is needed to refuse
  it. Security consequence, recorded deliberately: a caller cannot weaken a
  callee's exclusivity claim, so the `restrict` facts Phase 11 emits into C17
  derive from a single authority.

- **D-07-03:** Divergent cross-call loan state is expressible **today** under
  D-07-01 with no new grammar, and this was verified in the shipped tree during
  this discussion: `testdata/phase3/shared_shared_accept.lang` already binds
  loans to names (`let first = borrow buffer`), so `let v = borrow buffer` then
  `f(v)` gives Phase 08 the shape its mandatory gate needs. Phase 09's
  recursion / diamond / deep-chain shapes and its *"argument-loan-state
  variation across call sites to the same callee"* are likewise all
  arity-1-expressible — the variation lives in the **argument binding's loan
  state**, not in argument count.

- **D-07-04:** accepted consequence, recorded in writing rather than discovered
  later — at arity 1, `cgen` can emit at most **one `restrict` per callee**.
  The attribute is therefore provably *correct* but never *load-bearing* — the
  point of `restrict` is asserting that two pointers do not alias. **Phase 11
  must scope its `restrict` evidence explicitly to single-parameter derivation
  and name the two-parameter case as an M003 obligation.** A Phase 11 gate that
  implies two-pointer coverage it does not have would be exactly the
  "inert tier treated as proven" failure the milestone already guards against.

- **D-07-05 (wording correction for downstream phases):** Phase 08's gate must
  say **"call-graph depth ≥2"**, not "composition depth". Under D-07-01 the
  language has A-normal form calls — composition depth comes from call-graph
  depth (`f→g→h`), not from expression nesting. The current ROADMAP wording
  would let a reviewer read nested-expression capability into a gate that does
  not test it.

- **D-07-06:** M001's D-02 (*"parameters are always labeled at call sites"*) is
  **deferred, not revoked.** The shipped tree is already positional
  (`try lang_res_open(request)`), and at arity 1 labels carry zero
  disambiguation and are pure ceremony. D-02 is reinstated at the moment arity
  widens past 1 — which is the only point at which labels earn their cost.

- **D-07-07 (deliberately rejected):** arity-N in Phase 07. It would force
  multi-argument **loan interaction** rules (two arguments naming the same
  place; one shared and one exclusive loan of one owner) into the very phase
  that first makes `OpCall` real — Swift-exclusivity / Rust-two-phase-borrow
  territory — and no current language construct can consume a second argument,
  so the capability would be exercisable only by synthetic fixtures. A
  half-specified aliasing rule that Phase 08's gate then certifies as sound is
  the concrete hazard.

- **Posture, stated so a reviewer does not have to discover it:** this is
  A-normal form. Calls are a binding RHS exactly as `try`, `borrow`, and `take`
  already are; the canonical formatter owns surface shape; syntax remains
  provisional while the typed-core identity is the asset.

### Callee signature summary — field set and schema identity

- **D-07-08:** Stage 0 mints **`lang.interface/1`**. `lang.interface/0` bytes
  stay frozen and are decoded only by a **pinned legacy decoder struct**
  (`InterfaceV0`) with a frozen-bytes fixture proving no `/0` byte moved. A `/0`
  summary is **decodable but never admissible for a call**.
  — **Reversibility:** one-way — this is a published versioned product-API
  schema; per the standing project rule, `/0` bytes are frozen once shipped and
  can only be superseded additively by a `/1`.

  **Why a bump rather than additive growth.** The in-tree additive precedent
  (`core.LinearOperation`'s Phase 4 `OkEdgeID` / `ErrEdgeID` / `ErrTargetID` /
  `ReleasesOperationID` / `Allocator`) governs *optional facts populated only on
  new kinds, where absence is meaningful and correct*. Stage 0 is the opposite:
  every function must now carry per-parameter ownership and a total return
  contract, so **absence must be an error**. Under additive `omitempty` growth a
  pre-Stage-0 `/0` summary decodes cleanly into the widened struct with Go zero
  values — `mode: ""` is indistinguishable from *"the producer never knew about
  ownership"* — and admission proceeds on absent facts. That is fail-open at the
  exact boundary this phase exists to make fail-closed. D-06-04's rule reads
  literally on this case (*"only extending an existing document's fields earns a
  `/1` bump"*), and the bump is empirically cheap here: **4 literal sites**
  versus M001's 12, moving **no `core.Program` byte** — which is precisely the
  property `InterfaceSchema`'s independent versioning was created to buy.

- **D-07-09:** Recommended field set for `lang.interface/1`. Required and total
  unless marked otherwise; **no field admission depends on may have a legal
  empty value**, except `Callable`, whose zero value `false` fails closed.

  ```go
  const (
      InterfaceSchema  = "lang.interface/0" // frozen; decodable, NEVER admissible for a call
      InterfaceSchema1 = "lang.interface/1"
  )

  type FunctionSignature struct {
      ID            string              `json:"id"`
      Name          string              `json:"name"`
      Parameters    []ParameterContract `json:"parameters"`
      Return        ReturnContract      `json:"return"`
      Abilities     []Ability           `json:"abilities"`
      Callable      bool                `json:"callable"`        // D-04-03: callable ⊆ publishable
      Fails         string              `json:"fails,omitempty"` // "" = infallible
      Foreign       ForeignReach        `json:"foreign"`         // closure-derived worst case
      ClosureDigest string              `json:"closure_digest"`
  }

  type ParameterContract struct {
      ID    string `json:"id"`
      Name  string `json:"name"`
      Type  string `json:"type"`
      Mode  string `json:"mode"`  // "owned" | "shared" | "exclusive" — no legal ""
      Drops bool   `json:"drops"` // callee discharges the drop obligation on an owned param
  }

  type ReturnContract struct {
      Type  string   `json:"type"`
      Mode  string   `json:"mode"`  // "owned" | "shared" | "exclusive"
      Paths []string `json:"paths"` // empty iff Mode == "owned"
      Fresh bool     `json:"fresh"` // caller inherits a new drop obligation
  }

  type ForeignReach struct{ Allocator, Unwind, NonlocalExit string } // "" = none reachable
  ```

  `PublicOrigin` is **subsumed** by `ReturnContract` — same two access modes, now
  total instead of optional. Keep the `PublicOrigin` type for `core.Function`;
  drop it from the `/1` signature.

- **D-07-10:** `Parameters` is a **slice** (`[]ParameterContract`) even though
  the **checker admits exactly arity 1** under D-07-01. This is the deliberate
  reconciliation of two researcher recommendations that conflicted: the `/1`
  bump is forced by *required fields* regardless of arity, so the arity-ready
  schema shape is free to take now, with **one producer**, rather than later
  across five consumers. Schema capacity is cheap; the checker rule is where the
  cost lives, and the checker rule stays tight.
  — **Reversibility:** one-way at the schema layer (a published `/1`), but it is
  the *cheap* direction — it avoids a `/2` at the moment arity widens.

- **D-07-11:** **Call-graph edges are deliberately excluded from the summary.**
  An edge list would make every consumer whole-program-aware and re-import the
  already-rejected Design A. Edges live in `callgraph`; the summary carries only
  closure-derived **scalars** (`Foreign`, `ClosureDigest`).

- **D-07-12:** Keep `Interface.CoreDigest` (binds a summary to its own unit's
  core — detects staleness *of the summary*) and **add a per-function
  `ClosureDigest`**: a Merkle chain over this signature plus its callees'
  `ClosureDigest`s. Shape precedent: OCaml `.cmi` CRC chains and rustc's SVH — a
  chain over callee **summary** digests, never over a callee body. This is the
  only thing that closes the stale-but-self-consistent hole the standing verdict
  names (*"per-unit hashes will silently serve stale verdicts when an unchanged
  caller's callee changes"* — Rust ThinLTO import maps, Bazel/Buck cross-module).

- **D-07-13 (must be stated in the doc comment, not left implicit):**
  `CoreDigest` and `ClosureDigest` prove **integrity, not authenticity**. They
  are unkeyed content hashes — anyone who can write the summary can write a
  consistent digest. They detect **staleness**, never **forgery**. The forgery
  answer is and remains independent re-derivation (D-07-19). A consumer must not
  be able to mistake the digest for a signature.

### Cycle refusal — staging, diagnostic identity, algorithm, and peer

All five sub-decisions resolve toward one shape: **one derivation, two
independent runs of it, over disjoint reachable input spaces.**

- **D-07-14 (staging):** `check` emits `OpCall` normally, then runs `callgraph`
  over its **own completed in-memory `core.Program`** and refuses **before
  returning it**. There is **no AST-side graph.**

  A cyclic `core.Program` therefore exists only as an ephemeral local inside
  `check` — never returned, serialized, cached, interpreted, or lowered. The
  real trust boundary is the *returned artifact*, not the constructed value.
  Rejected alternative: building the graph from AST/name-resolution so no
  `OpCall` is emitted. Mutual cycles are whole-program facts that cannot be
  decided at single-call admission anyway, so it needs a second pass regardless
  — and it would mean two graph builders with two node-identity rules that can
  drift (shadowing, foreign-vs-declared resolution) **invisibly**, because
  neither peer sees the other's input.

- **D-07-15 (diagnostic identity):** exactly **one** code,
  **`core.call_graph_cycle`**, in the existing `core.*` namespace. SEM-07 says
  *"a named refusal code"*, singular. Three codes would triple a versioned API
  surface for one fact and — the decisive objection — make **classification
  itself** something the two peers can disagree about (is a self-call inside a
  3-cycle "direct"?).

  Shape:
  - `Code`: `core.call_graph_cycle`
  - `Primary`: span of the `OpCall` operation whose edge **closes** the
    canonical cycle
  - `Message`: `"call graph contains a cycle; recursion is refused"`
  - `Causes`, ordered, **canonically rotated so the lexicographically smallest
    function ID is index 0**:
    - `{kind: "cycle_length", detail: "<n>"}`
    - `{kind: "cycle_member", detail: "<functionID>", span: <call-site span of
      its outgoing edge>}` × n
  - Bound: max **32** `cycle_member` causes; on overflow append
    `{kind: "truncated", detail: "truncated:core.call_cycle_bound"}` and drop
    the overflow, matching the `truncated:evidence.trace_bound` convention
  - `Repairs`: **none.** A cycle is not a local edit, is not one of
    `lang-repair`'s five defect classes, and is declared **explicitly
    non-repairable** this milestone.
  - Direct / mutual / indirect are read off `cycle_length` and membership, not
    off the code.
  - `corevalidate` emits the **same inert string constant** with its own message
    and no spans. The code is shared as a string; the derivation is not shared.

  Printing the full chain follows Go's import-cycle diagnostics and Rust E0391.
  — **Reversibility:** one-way — a diagnostic code and its cause shape are
  published agent-facing API consumed by `lang-repair` and `lang explain`.

- **D-07-16 (why rotation is load-bearing, not cosmetic):** `Causes` participate
  in diagnostic **ID identity** (`diagnostic.go:102-112`). Without canonical
  rotation the same cycle yields a **different `ID` per discovery start order** —
  an unstable versioned API. Required test: discover the same cycle from two
  different root orderings and assert **one identical ID**. Removing the
  rotation is a seeded mutation that must fail that test.

- **D-07-17 (depth):** call-graph depth is **unbounded provided the graph is
  acyclic.** No maximum call-graph depth is declared in Phase 07. Depth is not
  the hazard — compile-time traversal is an explicit heap worklist and costs no
  Go stack — and a static ceiling would false-positive on exactly the
  parser-shaped programs spike S-007 flags. The distinguisher is **on-stack
  (gray) re-entry, never visit count or depth.** The fixed documented ceiling
  stays where it belongs, with SEM-08's interpreter call stack in Phase 10.

- **D-07-18 (algorithm):** **iterative three-color (white / gray / black) DFS
  over an explicit stack**, roots = **all declared functions** (not just the
  entry point), returning `(reversePostorder, error)`.

  Chosen over Tarjan/Kosaraju SCC condensation deliberately: LLVM's
  `CallGraphSCCPass` computes SCCs because recursive SCCs must be *analyzed*;
  here **every** SCC is refused, so condensation buys nothing, yields membership
  rather than a witness path, and adds lowlink bookkeeping that no test can
  distinguish correct from subtly-wrong (the output is always "refuse"). The
  three-color DFS stack hands you the **witness path for free**, and the
  **reverse postorder is exactly the ordering Phase 08's summary fixpoint
  needs** on an accepted graph — also free.

- **D-07-19 (peer):** `corevalidate` runs **its own** three-color traversal over
  the checked artifact, tested **exclusively against hand-built synthetic
  `core.Program` values** — no parser, no `check`. Exact in-tree precedent one
  level down: `pathoracle`'s `backEdgeError` (`pathoracle.go:112-125`), which
  re-derives a rejection *"rather than trusting check.go's own
  loanLivenessFixpoint to have caught it first, so a corrupted or synthetic CFG
  that check.go never saw is still refused here."*

  Independence here comes from **disjoint reachable input spaces**
  (parser-reachable vs. synthetic), not from a different algorithm. Writing that
  test *through the parser* instead of by hand would make the peer inert — which
  is M001's recorded failure shape. Plan structure must therefore include a
  synthetic-artifact builder helper and a mutation matrix that disables **each
  peer's refusal independently** and asserts the other still refuses.

### Stage 0 peer set

- **D-07-20:** **Two producers land in Stage 0, in the same plan.**
  `originvalidate.BuildInterface` widened is the **producer**; `corevalidate`
  structurally **re-derives the signature summary** (parameter ownership
  requirement, return-origin contract, abilities) from the checked core and
  refuses on mismatch. Non-liveness only.

  Stage 0's entire justification for existing is that the summary shape is
  provable *"while the change is still a pure data-shape change with one
  producer — not after five consumers depend on it."* That same sentence is the
  argument for adding the peer **here**: the marginal cost of an independent
  re-derivation is minimal precisely when there are no calls, no cycles, and no
  liveness contract to model, and it rises monotonically afterward. The shadow
  corpus is the **existing M001 Phase 4 corpus** — already built.
  — **Reversibility:** costly — deferring the peer later converts a structural
  re-derivation over a finished corpus into an interprocedural re-derivation
  over a call graph, landing in the phase that already carries the declared 2×
  scope-cut trigger.

- **D-07-21:** The peer must be wired into **both** `corevalidate` replay sites —
  `replayStraightLine` (`corevalidate.go:903`) **and** `replayBlocks` (`:1098`) —
  with a test that fails if only one has it. Wiring one and not the other
  reproduces the literal D-02-03 / D-03-01 failure (*"a quadratic cost fix
  landed in the wrong half"*; Lesson 3: *"fixing one of N independent peers is
  not fixing the item"*).

- **D-07-22 (explicitly NOT shipped):**
  - **No `check`-side producer.** `check` is the admission *consumer*; a summary
    it produces from its own state is correlated with its own decision by
    construction. Knight & Leveson found independently *written* versions already
    correlate their failures; a version sharing state correlates fully. It buys
    agreement theater, not detection.
  - **No shared `callsummary` package.** A single implementation that two callers
    agree with **cannot diverge**. That is anti-independence by definition, and
    exactly what the seeded-fault discipline exists to catch.

- **D-07-23 (what single-producer trust would actually hide):**
  `originvalidate.go:427` looks up type facts by exact `function.ID+":type:0"`
  match with a **silent empty-abilities fallback**. Under one producer that
  fallback yields a *self-consistent under-claimed* summary, and every caller
  admission in Phases 08-11 would be decided against an artifact nobody
  contradicts — `escape:coordinated-source-to-core-false-claim` promoted from
  origins to the entire call contract.

- **D-07-24 (three seeded faults, all required — QLT-08):**
  1. **Producer-only.** Force `BuildInterface` into the empty-abilities fallback
     whenever the type-fact ID lookup misses. Gate requires the peer to **diverge
     and refuse**.
  2. **Two-site.** A branch-shaped function returning a borrowed origin on
     exactly one arm while the summary claims owned. Gate requires refusal — a
     peer wired only into `replayStraightLine` **cannot** produce it. This is the
     direct D-02-03 repeat detector.
  3. **Bilateral.** Inject fault (1) into producer **and** peer. The gate must
     **FAIL**, reporting *"no divergence detected under bilateral fault"* —
     proving the comparison is not a value compared to itself.

  Backed by a **static import-independence test**: `corevalidate` must not import
  `originvalidate` or any shared summary helper.

- **D-07-25 (Stage 0 gate wording):** Stage 0's gate is *"zero divergence between
  producer and peer over every function in the existing M001 Phase 4 corpus,
  with both replay shapes exercised, plus all three seeded faults behaving as
  specified."* Recursion / diamond / deep-chain call-graph shadow shapes belong
  to Phase 09 — Stage 0's corpus has no calls.

### Cross-cutting

- **D-07-26:** declare Phase 07's scope-cut trigger now, in writing. The
  milestone declares a 2× trigger for Phases 08 and 09; Phase 07 currently has
  none. Adopt: **if Stage 0 + Stage 1 exceed ~2× their initial plan estimate,
  Stage 2's `callgraph` cycle refusal renegotiates out to Phase 08 — never the
  peer (D-07-20) and never the seeded faults (D-07-24).** Key Lesson 4: declare
  deferred scope in writing at the moment it is decided, not under pressure.

- **D-07-27:** a `PHASE-07-DEBT.md` entry written at planning time, not at phase
  end. It names: the deferred item (*interprocedural **liveness**
  re-derivation*), its D-03-02 lineage, its closure phase (09), its closure gate
  verbatim, and — explicitly — that the **signature-summary peer is NOT part of
  this deferral.**

- **D-07-28:** Phase 12 blindness guard, provable today. The `callgraph`
  builder must enumerate operations by `Kind == core.OpCall` **across all
  blocks**, never by block position, so calls inside future `Result` match arms
  are picked up automatically (SEM-07 names *"cycles through `Result`
  matching"*). This is provable in Phase 07 with a fixture that calls from
  **both arms of an existing `match`** — no `Result` payloads needed.

### Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all
four gray areas without per-area selection ("just fan out, follow ur recs").
Every decision above is therefore Claude's synthesis under that standing
instruction, grounded in the four advisor research returns and verified against
the shipped tree where a claim was checkable. Planner discretion remains over:

- Plan decomposition within each Stage (the sizing note expects Stage 0 as one
  plan, then coordinated-change plans at the M001 5-consumer rate of
  70-95 min/plan, not the ~10 min/plan single-consumer rate).
- Exact Go identifier and file names inside the new `callgraph` package.
- Whether `ForeignReach`'s three fields are literal strings or an existing
  enumerated type already in `core`.
- The precise corpus file names and module paths for the new fixtures.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Standing — durable, read before planning any M002 phase
- `.planning/STANDING-VERDICTS.md` — dependency verdicts (the
  zero-external-production-dependency record is a **hard constraint**;
  `golang.org/x/tools/go/callgraph`, LLVM `CallGraphSCCPass`, and Rust
  NLL/Polonius are all **read, don't import**), design anti-features
  (call-site convention override, whole-program borrow inference, indirect
  dispatch, implicit ARC fallback, silent depth-based stack growth, a second
  coexisting liveness law), process anti-patterns, and load-bearing facts about
  the six dispatch sites, `cgen`'s single-function hard fail, and why `-flto` is
  load-bearing.
- `.planning/LANGUAGE-MATURITY.md` — what the language can actually express
  today. No arithmetic, no iteration, no collections, no Lang-to-Lang calls; 58
  single-function programs averaging ~28 lines. Guards against planning as if
  those exist.

### Phase-scoped research (milestone-scoped; regenerated by the next `/gsd-new-milestone`)
- `.planning/research/ARCHITECTURE.md` §1 — the six dispatch sites, file by
  file, with per-site entry points and the specific cost of `OpCall` at each.
- `.planning/research/ARCHITECTURE.md` §2 — Designs A / B / C for the signature
  contract. **Design C is adopted and settled**; A and B are rejected with
  reasons. Do not re-litigate the artifact's existence.
- `.planning/research/ARCHITECTURE.md` §3 — call-graph placement (why a new
  `internal/compiler/callgraph` package and not `core` or `session`), the
  `cache` content-binding interaction, and where the refusal code lives.
- `.planning/research/PITFALLS.md` Pitfall 4 — fixpoint non-termination, SCC
  mishandling, native stack blowup in the checker itself.
- `.planning/research/PITFALLS.md` Pitfall 5 — generator reachability; why
  call-graph *shape* needs its own reachability-register entries.
- `.planning/research/SUMMARY.md` "The Six Dispatch Sites".

### Project design corpus
- `wiki/semantic-kernel-contract.md` — the kernel contract this phase extends.
- `wiki/ownership-evidence-roadmap.md` — ownership evidence obligations.
- `wiki/compute-efficiency-constitution.md` — the cost constitution that names
  whole-program inference as the failure mode to avoid.

### Project state
- `.planning/PROJECT.md` — core value, constraints, Key Decisions table
  (including the D-05-32 deferral of `OpCall` out of M001).
- `.planning/REQUIREMENTS.md` — SEM-04, SEM-05, SEM-06, SEM-07, QLT-08 verbatim.
- `.planning/ROADMAP.md` "Phase 07" — goal, success criteria, riskiest
  assumption, gate, sizing note, and the pre-phase spike decision.
- `.planning/RETROSPECTIVE.md` Key Lessons 2, 3, 4 — two coexisting laws;
  fixing one of N peers; declare deferred scope in writing when decided.
- `.planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-CONTEXT.md`
  — D-04-01 (`OpForeignCall` only), D-04-02 (`core.call_target_not_foreign` as a
  fail-closed control in both layers), **D-04-03 (callable ⊆ publishable — the
  lift condition SEM-06 implements)**, and the recorded dangling-alias witness
  program that motivates this whole phase.

### Live code this phase edits (verified in-tree during discussion)
- `internal/compiler/core/core.go:148-171` — `Interface`, `InterfaceSchema`,
  `FunctionSignature`, `Parameter`.
- `internal/compiler/core/core.go:244-252` — `AllOperationKinds()` /
  `TerminatorKinds()`, the registry that breaks all six switches at once.
- `internal/compiler/check/check.go:1316` — `resolveForeignStep`, the predicate
  D-07-01 widens and the `core.call_target_not_foreign` refusal SEM-04 flips.
- `internal/compiler/corevalidate/corevalidate.go:903` (`replayStraightLine`),
  `:1098` (`replayBlocks`), `:359` (callee resolution / `isDeclaredFunctionName`).
- `internal/compiler/originvalidate/originvalidate.go:415-434` (`BuildInterface`),
  `:427` (the exact-ID lookup with the silent empty-abilities fallback).
- `internal/compiler/pathoracle/pathoracle.go:112-125` — `backEdgeError`, the
  synthetic-artifact peer precedent D-07-19 generalizes.
- `internal/compiler/diagnostic/diagnostic.go:96-112` — `Causes` participating
  in diagnostic ID identity (why D-07-16's rotation is load-bearing).
- `internal/compiler/protocol/protocol.go:125-147, 189-193` — truncation-code and
  bound conventions; the "never from a body field" summary consumer.
- `internal/compiler/session/session.go:1027-1143` — `RunInterfaceCommandFile`;
  `:2501-2574` — the CLI-observable `lane:kind-exhaustive-dispatch` sibling lane.
- `internal/compiler/core/core_test.go:266-280` —
  `TestAllOperationKindsHandledAtEverySite` (`control:kind.exhaustive_dispatch`).
- `internal/compiler/debugmap/debugmap.go:54` — an `InterfaceSchema` literal site
  the `/1` bump must touch.
- `internal/compiler/cgen/cgen.go:22,52` — the `len(Functions) != 1` hard fail.
  **Phase 11's problem, named here so Phase 07 does not accidentally touch it.**

### Fixtures
- `testdata/phase3/shared_shared_accept.lang`, `testdata/phase3/public_view.lang`
  — the existing borrow/take binding grammar D-07-01 reuses unchanged.
- `testdata/phase4/foreign_acquire_one.lang` — the shipped positional call
  syntax (`try lang_res_open(request)`) D-07-06 records as the real precedent.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`core.Interface` / `core.FunctionSignature`** — already body-stripped,
  digest-bound, and produced source-blind by `originvalidate.BuildInterface`.
  Stage 0 widens this artifact rather than inventing a summary type. Its own doc
  comment already argues the structural case (*"deliberately has no Linear/Match
  field at all — not merely an omitted one"*).
- **`pathoracle`** — the package-posture template for `callgraph`: owns one
  whole-function derivation, consumed only by tests and the verify harness, never
  by `check`/`corevalidate`/`interp`/`cgen`. Its `backEdgeError` is also the
  synthetic-artifact peer-testing template.
- **The existing borrow/take binding grammar** — `let first = borrow buffer`,
  `let delivered = take buffer`. Divergent cross-call loan state is expressible
  with zero grammar additions.
- **The M001 Phase 4 corpus** — Stage 0's shadow corpus, already built. No new
  generative machinery needed for Stage 0's gate.
- **`ForeignContract.Fails` vocabulary** — reused verbatim by
  `FunctionSignature.Fails`; no new failure vocabulary is minted.
- **The `/0`→`/1` bump discipline** — M001 already executed one across 12 literal
  sites with `/0` bytes provably frozen. This is a third of that.
- **`truncated:evidence.trace_bound`** — the bounded-output convention
  `truncated:core.call_cycle_bound` copies.

### Established Patterns
- **Six literal, test-enforced dispatch sites.** Registering one
  `OperationKind` costs a minimum of 8-10 independent edits and deliberately
  breaks all six exhaustive switches at once — a forcing function, converting
  "six sites, easy to forget one" into "six failures, impossible to forget one."
- **Fail-closed `default:` arms.** `corevalidate`'s switches reject unregistered
  kinds via `core.unknown_operation` rather than silently accepting.
- **Duplicated straight-line vs. branch replay.** Both `check` and `corevalidate`
  handle straight-line and arm-shaped bodies in genuinely different code, not a
  shared helper. Every new case lands **twice**, per engine.
- **Shared inert records only.** Peers may share a diagnostic *code string*;
  they must never share the logic that derives the fact.
- **Graph walks carry visited-set guards and named refusal codes rather than
  hanging** — already a recorded project pattern; Phase 07 applies it to a new
  graph.
- **Bounded output with stable truncation codes.**

### Integration Points
- `check.go:1316` `resolveForeignStep` — the existing refusal becomes a producer
  path. Not a greenfield addition.
- `AllOperationKinds()` — the single registry edit that cascades to all six sites
  plus both exhaustive-dispatch controls (in-process and CLI-observable).
- `originvalidate.BuildInterface` — the summary producer to widen.
- `corevalidate` `replayStraightLine` + `replayBlocks` — two `case core.OpCall:`
  additions **and** the summary peer, all landing in both.
- `cache.Input` — `ClosureDigest` is the named, digested input that keeps `cache`
  structurally unaware of what a call graph *means* while remaining correct at
  the new scale.
- `session.go` — `RunInterfaceCommandFile` (`/1` writer + pinned `/0` decoder)
  and the CLI-observable exhaustive-dispatch lane.

</code_context>

<specifics>
## Specific Ideas

- **The corpus that makes success criterion 3 real, stated precisely.** The
  "pathological-depth-but-acyclic" corpus **must contain diamonds and shared
  leaves** (A→B, A→C, B→D, C→D, deep and repeated, parser-shaped) — **not just a
  deep chain.** A straight A→B→C→…→Z chain **will not kill** the highest-value
  seeded mutation (swapping the on-stack *gray* check for a plain visited check),
  because reverse postorder never revisits on a chain. Without diamonds, that
  gate is decorative. This is the single most actionable finding of the whole
  discussion.
- **Three more fixtures the adversarial pass named as individually necessary:**
  1. **Self-call.** A plausible "optimization" — skipping edges where
     `callee == caller` — silently legalizes direct recursion. Needs its own
     fixture and its own seeded mutation.
  2. **Unreachable cycle.** A cycle among functions not reachable from the entry
     point. Roots must be **all declared functions**; `corevalidate`'s synthetic
     artifacts may have no entry point at all. Fail-closed.
  3. **Foreign-symbol shadowing.** A declared Lang function whose name shadows a
     foreign symbol could drop a graph edge and hide a cycle. Node identity must
     use the same precedence rule `check` and `corevalidate` already share
     (`isDeclaredFunctionName`, `corevalidate.go:359`), with an explicit fixture.
- **Unresolvable callee.** A forged artifact naming a nonexistent callee must
  refuse through the existing `core.unknown_*` / `core.call_target_not_foreign`
  path — **never be silently dropped from the graph.** Dropping an edge is how a
  cycle escapes.
- **Diagnostic-ID stability test.** Discover the same cycle from two different
  root orderings; assert one identical diagnostic `ID`.
- **DoS bound.** Dedupe edges **at graph-build time** so `E` is bounded by `V²`
  rather than by operation count. Traversal is `O(V+E)`. Bound the **diagnostic**,
  never the traversal.
- **Prior art worth reading, not importing:** OCaml `.cmi` CRC chains and rustc's
  SVH (the `ClosureDigest` shape); Swift `.swiftinterface` and library evolution
  (total vs. optional signature fields); protobuf/Avro required-vs-optional
  evolution rules (why `mode: ""` is the bug); Go's import-cycle and Rust E0391
  diagnostics (both print the full chain); Rust ThinLTO import-map invalidation
  and Bazel/Buck cross-module staleness (why per-unit keys fail); Knight &
  Leveson on correlated failure in independently written versions (why a
  `check`-side producer is not a peer).

</specifics>

<deferred>
## Deferred Ideas

- **Arity-N calls and multi-argument loan interaction** — D-07-07. Requires an
  argument-aliasing rule (two arguments naming the same place; one shared and one
  exclusive loan of one owner) with its own diagnostic family and QLT-08
  mutation-kill obligations. The `/1` schema is arity-ready (D-07-10) so the
  future change is a checker predicate plus an aliasing rule, not a `/2`.
- **Labeled call-site arguments (M001 D-02)** — D-07-06. Deferred, not revoked;
  reinstated when arity widens past 1.
- **Two-parameter `restrict` evidence in `cgen`** — D-07-04. An **M003
  obligation**, and Phase 11 must say so in its gate rather than implying
  coverage it does not have.
- **Interprocedural *liveness* re-derivation in `corevalidate`** — Phase 09, held
  by the `PHASE-07-DEBT.md` entry required by D-07-27. Explicitly **not** the
  signature-summary peer, which ships in Phase 07.
- **A maximum call-graph depth / declared depth budget** — rejected for Phase 07
  (D-07-17). The interpreter's fixed documented ceiling is SEM-08, Phase 10.
  Spike S-007 informs whether that ceiling is a constant or a declared budget.
- **`cgen`'s `len(Functions) != 1` hard fail** (`cgen.go:22,52`) — Phase 11.
  Phase 07 registers `core.OpCall` at the `cgen` dispatch site to keep the
  six-site control green, exactly as `interp.go:100-123` already documents doing
  for kinds no current program can produce. It does **not** build multi-function
  emission.
- **Call-graph edges in the signature summary** — permanently rejected
  (D-07-11), not deferred. It would re-import the already-rejected Design A.
- **Keyed/signed summary digests** — out of scope. D-07-13 records that content
  digests detect staleness, never forgery, and that independent re-derivation is
  the forgery answer.

</deferred>

---

## Amendments from 07-RESEARCH.md (2026-09-08, post-research)

Research falsified four factual anchors used above. **The decisions stand
unchanged; their cost and their file references do not.** Independently
re-verified in the tree before recording.

- **A-01 — D-07-01 is parser work, not one predicate.**
  `internal/compiler/syntax/parser.go:404-421` unconditionally refuses **any**
  bare call in a binding RHS (`syntax.fallible_call_not_consumed`) at *parse
  time*, before callee identity is known — deliberately, per D-04-06, "so the
  core IR never has to encode a fallible operation without a failure successor."
  `resolveForeignStep` is `hasTryCall`-gated and therefore **unreachable from a
  non-fallible call**. So `let r = f(x)` needs a new `ast.RHS.Kind` and the
  foreign-vs-Lang admission decision **relocated from parse time to check time**.
  D-07-01's *rule* is unchanged; its *sizing* is not — Stage 1 must include
  parser work, and the relocated refusal's diagnostic code identity is an open
  question the planner must decide explicitly.
- **A-02 — `cgen`'s `OpCall` case is unreachable this phase, and neither control
  requires it.** `cgen.Emit`/`EmitNative` hard-fail on `len(Functions) != 1`, and
  **both** exhaustive-dispatch controls gate the `cgen` site behind that same
  condition (`core_test.go:346-352` — verified: `if len(program.Functions) == 1 {
  cgen.Emit(...) }`; `session.go:2559`). No legal `OpCall`-bearing program can
  have exactly one function. Registering the case in `cgen`'s switches is
  **cheap hygiene for Phase 11, not a Phase 07 gate obligation** — and D-07-04's
  boundary (do not lift the hard fail) is unaffected.
- **A-03 — two named fixtures do not exist.** Neither
  `testdata/phase3/exclusive_borrow_clean` nor the `relay`/`escort`
  dangling-alias witness quoted in D-04-03's narrative is present anywhere in
  the tree. **SEM-06's negative control has no existing artifact to point at;**
  both must be authored from scratch as an early task, ahead of SEM-06's gate.
- **A-04 — two `/1`-bump site references in D-07-08 are wrong.**
  `debugmap.go:54` is a **comment**, not a functional site, and
  `session.RunInterfaceCommandFile` does not exist — the real function is
  `session.InterfaceExportCommandFile` (`session.go:1028`). The bump is *cheaper*
  than D-07-08 assumed; the `/1` decision itself is unaffected.
- **A-05 — the two exhaustive-dispatch controls are literal, phase-scoped
  fixture lists, not generic sweeps.** Phase 07 must add **its own** fixture list
  to `core_test.go` and **its own** lane block in `session.go` mirroring
  `:2501-2568`. This is a task, not a free consequence of registering the kind.
- **A-06 — open question for the planner (from research):** `ClosureDigest`'s
  base case for a zero-callee function. Recommendation on the table: hash of its
  own signature with an empty callee list. Decide it explicitly in Stage 0; do
  not leave it implicit.

---

*Phase: 07-calls-signatures-and-call-graph-refusal*
*Context gathered: 2026-09-08*
*Amended: 2026-09-08 after 07-RESEARCH.md*

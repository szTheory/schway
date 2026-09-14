# Phase 08: Interprocedural Loan Liveness in `check` - Context

**Gathered:** 2026-09-09
**Status:** Ready for planning

<domain>
## Phase Boundary

`check` decides, from a callee's signature summary alone, whether a loan taken in
the caller is still live across a call boundary — and the decision is **bounded**,
**terminating**, **measurably cheap**, and **readable after the fact**.

Concretely, this phase delivers four things:

1. **A two-bit liveness summary per function**, derived exactly once per function
   in `callgraph.Order`'s reverse postorder, memoized within the run, carried on
   the existing in-`check` signature table alongside `Callable`.
2. **The interprocedural clause in the sole liveness law** — a forward
   canonicalization branch in `derivePlaceLoans` plus one backward clause in
   `blockLoanLiveness`. Not a new dataflow engine.
3. **A fail-closed iteration bound** on the intraprocedural worklist with a named
   `check.*` refusal, provoked through an unexported seam.
4. **A cost gate** that is the only mechanism able to detect the phase's central
   failure mode — a fitted growth exponent over deterministic work counters,
   measured on a generated call-graph corpus, ratified into the QLT-02 budget
   manifest.

**Requirements:** OWN-06, EFF-02.

**Not this phase** (each has its own phase and its own gate): `corevalidate`'s
*liveness* peer and D-03-02 closure (Phase 09), OWN-09's retirement of the
intraprocedural law (Phase 09 — see D-08-19, this conflict is flagged not
resolved), the interpreter call stack (Phase 10), multi-function C emission
(Phase 11), cross-run summary caching / QLT-06 (Phase 11), `Result` payloads
(Phase 12), the agent loop (Phase 13).

</domain>

<decisions>
## Implementation Decisions

All four gray areas were researched in advisor mode (four parallel
`gsd-advisor-researcher` agents, `minimal_decisive` calibration) under the
developer's standing mandate to fan out across stakeholder-role lenses, run an
adversarial pass, draw on cross-ecosystem prior art, and synthesize one-shot
recommendations. The developer directed that the synthesized recommendation be
adopted in every case.

**Every factual claim below was independently re-verified against the shipped
tree by the orchestrator before being locked.** Three researcher claims required
correction; each correction is recorded inline at the decision it changes
(D-08-13, D-08-24, D-08-26). Where a researcher and the tree disagreed, the tree
won.

### Summary source, memoization, and where the interprocedural fact enters

- **D-08-01 (the falsification that decides the area):** deriving the liveness
  answer from the **already-declared `/1` fields alone is not viable**, and this
  was investigated rather than merely disfavoured. `ReturnContract.Mode` +
  `Paths` does declaratively state `ReturnsBorrowOfParam`. But **no field
  anywhere in `lang.interface/1` states `UsesParam`** — whether the callee reads
  *through* the parameter loan without that read escaping into the return.
  `ParameterContract.Mode` (`"owned"|"shared"|"exclusive"`) is a **declared
  ownership convention** read implicitly from the signature per D-07-02, not an
  observed body fact; `core.FunctionSignature`'s own doc comment records that it
  "deliberately has no Linear/Match field at all — not merely an omitted one",
  i.e. it was **built incapable** of carrying `UsesParam`.

  The only sound declared-only fallback is to assume `UsesParam == true` for
  every shared/exclusive parameter. That over-approximation refuses the safe
  program in `borrow; move; call` shape — which is **precisely the safe twin
  criterion 1 requires `check` to accept**. A declared-only design therefore
  fails the phase's own gate by construction.

- **D-08-02:** The two summary bits are **`UsesParam`** and
  **`ReturnsBorrowOfParam`**, matching spike S-006's validated model exactly.
  Both are **transitive through relays**, which is what makes a summary depend on
  the whole reachable subgraph rather than on one body — the cost question in
  miniature, and the reason memoization is load-bearing.

- **D-08-03:** Both bits are **derived once per function, in
  `callgraph.Order`'s reverse postorder** (callee-before-caller), **memoized
  within the run**. There is **no call-graph-level fixpoint** — S-006 finding #2:
  "summaries can be derived callee-before-caller in ONE PASS with NO
  call-graph-level fixpoint at all", and the workbench's lazy memo landed on the
  identical count (one derivation per function), which means **the ordering
  choice is free and the acyclicity guarantee is what is load-bearing**. Phase 07
  already computes this order for cycle refusal; RPO is free here.

- **D-08-04 (this is NOT a schema bump, and that is the point):** the two bits
  live on the **existing in-`check` `callSignatureTable`**, in memory. **No
  `lang.interface/2`, no new `/1` field, no new published document.**

  The justification is a precedent already shipped, not an argument: `Callable`
  is *already* a body-derived bit sitting on `callSignatureTable`, whose own doc
  comment records that the table is built "AFTER every function's body is checked
  (its `Callable` bit recomputes return origin from checked `core.LinearOperation`
  facts)". `UsesParam`/`ReturnsBorrowOfParam` are the **same shape of fact**:
  derived once per function from its own checked body, then consulted by every
  caller as a table field, never re-walked per call site. Taking this route
  avoids D-07-08's recorded one-way cost of a published schema bump entirely.
  — **Reversibility:** reversible — an in-memory unexported table field. This is
  deliberately the cheap direction; publishing these bits into `/1` later remains
  available and would then be a considered decision rather than an accident.

- **D-08-05 (OWN-06 gets a recorded interpretation, not a silent liberty):**
  OWN-06 reads *"`check` derives interprocedural loan liveness from callee
  signatures only — never by re-walking callee bodies."* Under D-08-03 the
  **production** of the two bits does walk each callee's own checked body once.

  The interpretation locked here — and it is **not new**, it is the reading
  D-07-34 already recorded for `Callable`: **"from callee signatures" governs
  *consumption*.** Each per-call-site admission reads only the table and
  structurally cannot reach a body (the table holds `core.FunctionSignature`,
  which has no `Linear`/`Match` field even to follow). The table itself is built
  once, callee-before-caller, ahead of any caller admission. That two-tier
  structure is exactly what `callSignatureTable` already is today.

  **Write this into the phase's decision record explicitly.** An undeclared
  version of this reading is the failure mode the debt register exists to
  prevent; a reviewer must not have to discover it.

- **D-08-06 (placement — where the table lives):** **inside `check`, unexported**,
  as an extension of `buildCallSignatureTable`. **Not** a new
  `internal/compiler/summary` package and **not** a shared helper.
  ARCHITECTURE.md §4 rejects a shared interprocedural-liveness library
  **outright** ("a shared implementation isn't a peer at all, it's a single point
  of failure wearing two names"), and D-07-22 already rejected a shared
  `callsummary` package for the same reason. Phase 09's `corevalidate` peer must
  derive its own two bits by its own materially different route — a reachability
  closure, not a worklist — and must not import `check`.

- **D-08-07 (mechanism placement — S-006 iteration 2 is direct authority, and
  getting this wrong silently disables the gate):** the interprocedural fact
  enters through a **forward canonicalization pre-pass in `derivePlaceLoans`
  (`check.go:1049`)**, **not** through a clause in the backward transfer
  function.

  The spike's first design put it in the backward transfer and it was wrong, and
  the fixtures said so immediately: in `borrow; call; move; use(result)` the
  backward walk reaches the **`move` before the call that establishes the
  alias**, so the loan is not yet live and **the refusal never fires**. A
  liveness law with this defect passes every accept-case test and silently admits
  the program the phase exists to refuse.

  `derivePlaceLoans` already has precisely the required shape for reborrow chains
  (`latestLoan[placeID]` / `parentLoan[loanID]`). The extension is additive and
  structurally identical:
  - `OpCall` where the callee's `ReturnsBorrowOfParam` is true → bind the call
    result's `TargetID` to the **same loan identity** the argument place carries
    (`chain.latestLoan[TargetID] = chain.latestLoan[SourceID]`), exactly as a
    reborrow would;
  - otherwise → the **existing generic fallthrough already does the right
    thing**, giving the result a fresh identity no borrow ever creates, which is
    S-006's canonicalization rule verbatim.

- **D-08-08:** `UsesParam` is consumed by **one added clause in
  `blockLoanLiveness`**: an `OpCall` whose callee's `UsesParam` is true counts as
  a **use of its argument's current loan chain**, symmetric to how a plain
  reference is already treated. This is what makes `borrow; move; call` refuse.
  **The backward worklist itself is otherwise unchanged** from spike 002's shape
  (S-006 finding #3) — the new subsystem this phase builds is the summary table,
  not a new dataflow engine.

- **D-08-09 (D-07-49 lands in BOTH paths, or it is not landed):** the entry
  defect is confirmed in tree — `computeLoanLastUses` (`check.go:2979`) switches
  on `binding.RHS.Kind` over exactly `"take"` / `"borrow"` / `"borrow_mut"` with
  **no `"call"` case**, so a call binding falls through to the `core.OpCopy`
  default and the argument loan's liveness is not extended across the call.

  There are **two admission paths** over the same law: `computeLoanLastUses`
  (AST-derived, over `ast.LinearBody` bindings, called at `check.go:1328` and
  `check.go:2529`) and the real `derivePlaceLoans` path (over emitted
  `core.LinearOperation`s carrying `CalleeID`). **Both must receive the fix, with
  a test that fails if only one has it.** Wiring one and not the other reproduces
  the literal D-02-03 / D-03-01 failure and D-07-21's lesson — *"fixing one of N
  independent peers is not fixing the item."* Which of the two is load-bearing
  for the criterion-1 gate is a planning-time determination; shipping only one is
  not.

- **D-08-10:** **Two bits are enough this phase.** `ParameterContract.Drops`
  already exists and answers a *consumption* question (does the callee discharge
  the owned parameter's drop obligation), not a liveness question — no third bit.
  No per-arm bit: calls inside `match` arms are a call-graph-edge concern already
  solved by D-07-28 (`callgraph` enumerates by `Kind == core.OpCall` across all
  blocks, never by block position), not a summary-bit concern. **Re-open the
  moment arity widens past 1 (D-07-07) or `Result` match arms can independently
  borrow-vs-move a parameter (Phase 12).**

- **D-08-11 (a load-bearing invariant that must be stated where the derivation
  lives, and pinned):** **summary derivation must consume operations in program
  order.** S-006 iteration 5 measured 4.0 work units per operation, flat from
  k=8 to k=512, in program order — and **12.4 → 767.0 per operation for the same
  chain listed in reverse, a 192× penalty at k=512, quadratic in body length.**
  Two passes suffice in program order because a body in program order propagates
  the whole chain in its first pass.

  This is an invariant Phase 08 **inherits**, not a property of the mechanism.
  Real `core.LinearOperation` sequences are in program order today; any future
  pass that reorders operations before summary derivation hands the bound back.
  State it in the derivation's doc comment and pin it with a test, following the
  in-tree precedent of `TestReborrowChainWorkIsLinear`.

- **D-08-12:** the memo is **within-run only**, keyed by function ID, rebuilt on
  every `check.Program` call, **never persisted**. A within-run memo has no
  invalidation problem by construction: the run observes one immutable
  `core.Program`, `callgraph.Order` has already proven it acyclic before any
  derivation happens (D-07-38's ordering constraint mirrored here), and RPO
  guarantees each function is derived exactly once, after all its callees, before
  any of its callers consult it. **Add a test asserting the memo is never
  persisted and is rebuilt per invocation** — that constraint is the thing under
  test, since honouring it is what keeps this phase clear of QLT-06 (see D-08-26).

### The fail-closed iteration bound and its named refusal

- **D-08-13 (the honest framing, and it changes what criterion 2 can mean):**
  there is **no call-graph fixpoint to bound** — S-006 finding #2 establishes RPO
  is one pass, so that axis is bounded by construction (a permutation of a finite
  function list) and has no fixpoint claim to assert. Summary derivation is two
  passes in program order (D-08-11), also bounded by construction. **The only
  genuine iterative fixpoint is the intraprocedural worklist** inside
  `loanLivenessFixpoint` (`check.go:1142`), which already maintains a `work`
  counter (one unit per transfer-function evaluation, one per worklist
  reinsertion) returned as `loanLivenessResult.work`, and today has **no bound at
  all** — it terminates purely by monotonicity over a finite lattice.

  Therefore the bound is an **internal-consistency assertion, not a DoS
  defense.** A monotone fixpoint over a finite lattice provably converges within
  lattice-height iterations. **No `.lang` program can force divergence**; only an
  implementation bug can — a non-monotone transfer function, a mutated lattice, a
  lost `queued` flag reintroducing infinite reinsertion. Stating this plainly is
  what keeps the resulting control from being a magic number nobody can audit.

- **D-08-14:** the bound is **derived, not a fixed constant**:
  `k × len(blocks) × distinctLoanCount`, justified in the doc comment by the
  lattice-height argument. A flat constant (e.g. `10_000`) is strictly worse
  here — equally unprovokable, with no derivation trail, and it would
  false-positive the day `blocks × loans` legitimately grows in a later milestone
  that adds iteration. This follows Z3's choice of a deterministic `rlimit` over
  a wall-clock timeout for reproducibility, rather than LLVM/GCC's folklore
  `--param` constants (which exist precisely because those analyses lack the
  lattice argument this one has).

- **D-08-15 (declared at planning time, per Key Lesson 4, not discovered
  mid-phase):** criterion 2's *"a program engineered to exceed that bound"* is
  **unsatisfiable from source in this language** — arity 1, no loops, no
  iteration, no collections, 58 programs averaging ~28 lines. The refusal is
  **unreachable from any legal source program** and its only living witness is a
  seeded seam.

  This is the **exact disposition D-07-47 already established** for
  `check.call_return_type_unrepresentable`, and it must be written into
  `PHASE-08-DEBT.md` **at planning time**, naming the item, why it is
  unreachable, and the seam that mutation-kills it.

- **D-08-16 (the seam is unexported — verified precedent):** provoke the bound
  through an **unexported package-level seam** in the `callReturnTypeDerivationSeam`
  shape, verified in tree at `check.go:493` (`var callReturnTypeDerivationSeam =
  false`, read at `check.go:1928`, flipped and deferred-restored in
  `check_test.go:3412`). **Not** `pathoracle.TerminatorKindsOverride`
  (`pathoracle.go:51,59`), which is **exported** — D-07-42 cites it as the shape
  precedent while explicitly forbidding multiplying exported mutable globals
  across production `check`/`corevalidate` paths, on `-race` and test-order
  grounds. `loanLivenessFixpoint` takes `[]cfgBlockSpec`, an unexported type, so
  any test exercising it is same-package by construction and the seam route fits
  naturally. QLT-08 applies: mutation-kill this control in the plan that
  introduces it.

- **D-08-17 (namespace, and why it differs from Phase 07's cycle choice):** the
  code lives in the **`check.*` namespace**, not `core.*`.

  D-07-15 chose `core.call_graph_cycle` precisely because `check` and
  `corevalidate` re-derive the *same structural fact* from the *same adjacency
  shape* and must agree. That does not hold here: Phase 09's peer computes
  liveness through a **reachability closure, not a worklist**, so it has a
  differently-shaped ceiling and **cannot fail the same way**. A shared `core.*`
  code would assert an agreement the two mechanisms are structurally incapable of
  having. `check.cfg_back_edge` is the in-tree precedent for a mechanism-local
  fact that correctly stayed `check.*`.

- **D-08-18 (a stability trap, stated so it is not rediscovered):** **the bound's
  numeric value must never appear in a `Cause` detail or in the `Message`.**
  `Causes` participate in diagnostic **ID identity**
  (`internal/compiler/diagnostic/diagnostic.go:96-112` — `Error` and
  `ErrorWithRepairs` fold `Code + Span + Causes` into the identity hash), so a
  bound value embedded in a cause would silently move every existing diagnostic
  ID for that code the first time the bound is retuned. This is the same class of
  trap D-07-16 and D-07-43 solved for cycle witness stability. The refusal is
  **not `Repairs`-eligible** — it is not one of `lang-repair`'s five defect
  classes, matching Phase 07's non-repairable disposition for cycles. `Primary`
  is the function's own declaration span, since a whole-function property has no
  single offending operation.

- **D-08-19a (native stack safety — do it here, not later):** convert
  `loanLivenessFixpoint`'s cycle pre-walk (`check.go:1157-1180`, a
  `var walk func(id string) error` calling itself over successors) from **native
  recursion to an explicit stack**, in this phase. `callgraph.Order` was
  deliberately written as an iterative explicit-stack three-color DFS (D-07-18)
  for exactly this Pitfall-4 reason. The change is small, mechanical, and sits
  inside the very function this phase is already editing; shipping the phase
  without it **while citing D-07-18 as prior art** is an inconsistency an auditor
  flags immediately. Depth here is bounded by CFG block count within one function
  body, which is smaller than `callgraph`'s roots — but "bounded by body size" is
  not the same guarantee, and a long straight-line body is adversarially
  reachable.

  Correspondingly: **the new transitive `ReturnsBorrowOfParam`-through-relays
  walk must be iterative from the start**, not converted after the fact.

- **D-08-19b (regularize while in the same function):** `check.cfg_back_edge` is
  today a bare `fmt.Errorf` string (`check.go:1163`), not a coded
  `diagnostic.Diagnostic`. Promote it to a coded diagnostic while this phase is
  already hardening the same pre-walk. This is the same file, the same fixpoint,
  and the same OWN-03 acyclicity story — not scope creep.

### Refusal identity, blame, and the criterion-4 disclosure

- **D-08-20 (mint, do not reuse):** an interprocedural loan-liveness refusal gets
  its **own code, `check.interprocedural_loan_liveness`**. Verified: every
  existing loan code is in the `ownership.*` namespace —
  `ownership.move_while_borrowed` (`check.go:1430`), `ownership.borrow_conflict`
  (`:2793`), `ownership.use_after_move` (`:511`),
  `ownership.borrow_requires_share` (`:814`),
  `ownership.transfer_requires_take` (`:1508`).

  Reusing one would force `cmd/lang-repair` to inspect `Causes[].Kind` **before**
  deciding whether to propose an edit — because `ownership.borrow_conflict`
  carries local repairs (`narrow_to_shared_borrow`,
  `create_loan_after_conflicting_loan_ends`, `check.go:2809-2825`) that are edits
  to the *same* function, whereas a cross-function conflict's fix is not. That is
  a strictly worse contract for the AI agent this project weights heaviest
  (*"AI effectiveness depends more on precise verifier feedback than exotic
  syntax"*), and it degrades `Code` as a reliable dispatch key.
  — **Reversibility:** one-way — a diagnostic code and its cause shape are
  published agent-facing API consumed by `lang-repair` and `lang explain`.

- **D-08-21 (namespace now, promotion later — the `core.call_graph_cycle`
  pattern, applied in order):** `check.*` **in Phase 08**, promoted to
  **`core.interprocedural_loan_liveness` in Phase 09** at the moment
  `corevalidate` independently re-derives the same fact and both peers must agree
  on-code. `core.call_graph_cycle` was minted `core.*` only because
  `callgraph.CycleError` was designed as a shared, peer-agnostic **witness type**
  available to both layers from day one; `check.cfg_back_edge` stayed local
  because it had one deriver. Phase 08 has one deriver. **Do not pre-commit the
  `core.*` string before Phase 09 has verified independent derivation produces the
  same semantics** — that is D-07-15's rework lesson.

- **D-08-22 (blame):** `Primary` is the **failing operation** — the `take` in
  `let v = borrow buffer; f(v); take buffer` — not the call site, not the loan's
  origin, not the callee's declaration. This matches how
  `ownership.move_while_borrowed` and `ownership.borrow_conflict` already point
  `Primary` at the operation under check rather than at the loan's origin
  (`borrowConflictDiagnostic`, `check.go:2793-2830`: span-bearing causes first,
  then ID-bearing detail causes). The call site and the borrow go into `Causes`.

- **D-08-23 (the cause chain, and why it needs no rotation):** a **fixed
  three-role template**, not a variable-depth DAG walk:

  1. `{Kind: "borrow_created_here", Span: <borrow span, caller>}`
  2. `{Kind: "loan_extended_by_call", Span: <call-site span, caller>, Detail: "<calleeID>"}`
  3. `{Kind: "callee_return_contract", Detail: "<calleeID>:return.mode=<Mode>"}`
     — or `":parameters[0].mode=<Mode>"` for the `borrow; move; call` pattern.

  **Determinism is by construction, not by normalization.** Positions are
  role-keyed and never selected from a candidate set, so **there is nothing to
  rotate** — this sidesteps the D-07-16/D-07-43 problem rather than re-solving
  it, and it structurally answers Pitfall 7's "the bounded cause DAG grows
  unbounded across function boundaries": the chain **never recurses into a middle
  relay's own causes**. It stops at the immediate caller's borrow, the immediate
  call, and one fact about the immediate callee's contract.

  If a future extension needs multi-hop relay chains, the bound is
  `MaxInterproceduralLoanCauses = 32` with
  `truncated:check.interprocedural_loan_bound`, ordered by `callgraph.Order`'s
  RPO position (inheriting determinism from an already-deterministic source
  rather than inventing an ordering). Both constants are declared **in `check`,
  not `protocol`** — the same import-cycle reason `callgraph.MaxCycleCauses` and
  `callgraph.TruncatedCycleBound` were declared locally at `callgraph.go:61,71`
  rather than in `protocol` (`protocol.go:137-144` records the cycle
  `check → protocol → interp → [test] → check`).

- **D-08-24 (a schema limit that must not be papered over — verified):**
  **cross-file spans are impossible today.** `diagnostic.Span` is
  `{Start, End int}` (`diagnostic.go:17-20`) with **no file or module field**, and
  `Cause.Span *Span` (`:22-26`) is likewise fileless. A callee declared in
  another module **cannot be pointed at**.

  **Do not widen `Span` in Phase 08.** D-07-08 records that a published schema
  change is one-way, and widening it on research-only justification is exactly
  the speculative commitment that rule exists to prevent. The callee-contract
  cause is therefore **ID-only and spanless** — already an established shape in
  this very file (`{Kind: "loan", Detail: blocking.id}`, `check.go:2809`).

- **D-08-25 (`Repairs: nil`, and the reason is a safety reason):** the only valid
  fix for a cross-function loan conflict lives in a **different function than
  `Primary` points at** — either shorten the loan in the caller before the call,
  or change the callee's declared return contract. Proposing either as a
  `diagnostic.Repair` on this diagnostic would have `lang-repair` **edit code the
  diagnostic does not point at**. That is unsafe by the same reasoning D-07-15
  used to declare cycles non-repairable, and it matches
  `ownership.borrow_requires_share` / `ownership.transfer_requires_take`'s
  existing classification-only precedent rather than `borrow_conflict`'s
  repair-bearing shape.

- **D-08-26 (criterion 4 — the disclosure, and the half of it that cannot be met
  this phase):**

  **On refusal:** the disclosure is cause 3 of D-08-23, naming the **one specific
  field** the answer depended on — `return.mode` **or** `parameters[0].mode`,
  never a bitmask and never "all of them". The liveness law consults exactly one
  bit per direction, so naming one field is both honest and sufficient.
  Body-blindness holds: `Mode` is copied verbatim from the declared AST type
  (`core.go`'s `ParameterContract` doc records "copied verbatim from
  `core.Function.Parameter`"), so disclosing it reveals nothing about the callee's
  control flow and cannot reconstruct body facts — SEM-05 is preserved.

  **On acceptance: criterion 4 is NOT met by any shipped runtime artifact in
  Phase 08, and this gap is declared, not papered over.** Verified against the
  tree, all three candidate vehicles fail:
  - `protocol.ExplainSummary` (`protocol.go:223-249`) is synthesized **from an
    existing diagnostic's flat `Causes` list**. There is no diagnostic on an
    accepted program, so `lang explain` has nothing to expand.
  - `internal/compiler/cache` is **structurally incapable of holding a verdict** —
    the guard is named literally, `TestCacheExportedSurfaceStoresNoVerdict`
    (`cache_test.go:120`).
  - A new sibling `lang.*/1` document contradicts `protocol.InterfaceSummary`'s
    own recorded anti-pattern (`protocol.go:149-152`): mirroring `/1` fields
    "would create a second schema that can drift from the first with no producer
    forcing them together". The established pattern is a deliberately lossy
    command-specific view, not a second copy.

  **Disposition:** satisfy the accepted case with a **table-driven test as
  documentation-of-record** (per accepted fixture, assert which specific
  `ParameterContract.Mode` / `ReturnContract.Mode` value admission consulted), and
  take the always-on `--explain`-style runtime record to the **mid-phase gate** as
  an explicit cost question. An always-on provenance record has a real measured
  cost under `wiki/compiler-and-feedback-latency.md`, and Phase 08's own goal is
  to *prove* cheapness first — that is the wrong place to spend the budget in the
  phase that establishes it. **This is a criterion-4 scope boundary and belongs in
  `PHASE-08-DEBT.md` at planning time.**

- **D-08-27 (OWN-09 — flagged for a human, deliberately not resolved by Claude):**
  the project's own documents disagree. **OWN-09's text** says the intraprocedural
  law is *"retired in the same phase the interprocedural law lands. One law, not
  two"* — which is **Phase 08**. **`REQUIREMENTS.md:173`** maps OWN-09 to
  **Phase 09**. Only a human decision closes this.

  **Interim rule locked for Phase 08 planning:** `ownership.*` codes are **not
  retired** in this phase, and the **interprocedural check runs last**, after
  intraprocedural admission has already passed — so the two never overlap in
  scope and no program can be judged by both laws for the same fact. Retirement
  lands in Phase 09 alongside the peer. Retiring the sole law **before** a second
  independent detector exists to validate the interprocedural law's completeness
  would violate RETROSPECTIVE Key Lesson 2 directly. **Surface this at the
  mid-phase gate as an open question.**

- **D-08-28 (the criterion-1 corpus, with a per-shape "or the gate is decorative"
  argument — this is the section a planner must not compress):**

  1. **Both S-006 caller-side patterns, each as a refuse/accept twin pair.**
     - Pattern A: `let v = borrow buffer; let r = f(v); take buffer` — refuses iff
       `f`'s `ReturnContract.Mode ∈ {shared, exclusive}`, accepts iff `"owned"`.
     - Pattern B: `borrow; move; call` — refuses iff the callee uses its
       parameter.
     **Required or the gate is decorative:** the twin must differ in **exactly the
     callee's declared contract field** and nothing else. A pair that varies the
     *caller's* shape instead would **pass with the interprocedural law entirely
     deleted**, because the existing intraprocedural `ownership.*` law already
     refuses same-function violations. Only a callee-contract-driven pair proves
     the interprocedural derivation ran at all.
  2. **A depth-≥2 relay chain, transitive through a pass-through middle**
     (`caller → relay → leaf`, where `relay`'s own `Return.Mode` forwards
     `leaf`'s). **Required or the gate is decorative:** a depth-1-only corpus
     cannot distinguish *"the derivation reads callee signatures"* from *"it
     happens to work for the one-hop case"*. This is the direct generalization of
     Phase 07's single most actionable finding (a straight chain could not kill
     the highest-value cycle mutation; diamonds were required).
  3. **`testdata/phase07/relay_escort_witness.lang` is the criterion-1 refusal
     fixture** — not a solved case. Its own header records that it is *"exactly
     the INTERPROCEDURAL half of D-03-02"* and that it **checks clean today only
     because `computeLoanLastUses` has no `"call"` case**. Landing D-08-09 flips
     it from clean to refused; that flip is a required assertion.
  4. **A negative control pair differing only in fields the law must NOT
     consult** — `Fails`, `Foreign.*`. Both members must resolve identically.
     **Required or criterion 4 is half-proven:** a corpus that only ever varies
     liveness-relevant fields never demonstrates the law correctly *ignores* the
     fields it must not read. This is the "which fields are **not** named" half of
     criterion 4 and it is cheap.
  5. **Call in both `match` arms** — **should-add, not gate-blocking.** D-07-28's
     both-arms guard was about call-graph *edge discovery* (miss an arm and cycle
     detection goes blind); liveness admission is per-call-site and arity-1
     regardless of arm, and `check` already walks both arms uniformly. Record it
     as a recommended regression fixture rather than asserting a requirement the
     evidence does not support.

### Cost gate (EFF-02, criterion 3 — the phase's riskiest-assumption gate)

- **D-08-29 (instrument — deterministic counters gate, wall-clock observes):**
  the **hard** gate is a **deterministic work counter**; `elapsed_ns` is recorded
  **observed-only**. The precedent already ships:
  `internal/compiler/session/qlt02_budget_manifest.json` carries
  `{"metric": "recomputed_work", "gate_type": "hard", "unit": "count"}` beside
  `{"metric": "elapsed_ns", "gate_type": "observed"}`.

  This is forced, not preferred. The roadmap states the failure mode **"does not
  fail a correctness test; it fails a cost curve"**, and S-006 states its own
  instrument rule verbatim: *"Wall-clock nanoseconds are recorded in a separate
  field and are never used to classify a mechanism."* Wall-clock is confounded by
  exactly the noise `measure/machine.go`'s own `CoVDemotionThreshold` comment
  names — battery state, P/E core scheduling, thermal throttling, single-host, no
  CI fleet — which would mask a super-linear curve at any corpus size a CI budget
  can afford. This mirrors rustc's own decision to gate perf regressions on
  instruction counts rather than wall time.

- **D-08-30 (EFF-02's wording gets a recorded interpretation):** verified —
  `measure/statistics.go`'s `Samples` is a bare `[]int64` and `Summary()` computes
  P50/P95/Mean/StdDev/CoV with **no nanosecond-specific logic anywhere**. The
  protocol is genuinely unit-agnostic, so applying it to **work-count samples**
  satisfies EFF-02's *"the existing p50/p95/CoV protocol"* **literally** and
  invents no parallel instrument. The **fitted growth exponent** is an
  *additional* hard control layered on the same deterministic metric across a size
  sweep, not a replacement. Record this in the decision log — a literal
  timing-only reading of EFF-02 is a live misreading that would make the gate
  decorative.

- **D-08-31 (the bound shape):** **fitted growth exponent of work against
  OPERATION count, bounded at ≤ 1.2**, plus a **ratio-stability tripwire**
  (`work/operations` at size S vs 4S differing by ≤ 15%, reusing
  `CoVDemotionThreshold`'s own 0.15 convention).

  **Operation count, not function count — this is not a stylistic preference.**
  S-006 records it under Surprises: the `dense` shape fits **1.50 against function
  count and 1.11 against operation count**, the difference being entirely the
  corpus's leaf-to-relay ratio drifting with size, and **"fitting only against
  function count would have produced a wrong finding."**

  1.2 gives headroom over the spike's memoized cluster (1.02–1.11, `dense` the
  tightest at 1.11) for real production-path overhead absent from the workbench,
  while staying far below the quadratic regimes the unmemoized arms hit
  (1.9–6.7). An **absolute ceiling is rejected as the sole bound** — a
  whole-program analysis passes under it, which is the exact decorative failure
  this gate exists to prevent. A ceiling may still be recorded as `observed`,
  mirroring `output_bytes`.

- **D-08-32 (the correction that changes the plan — highest-value finding in this
  area, verified at two independent chokepoints):** the naming of the new metric
  is **not free**. **Two places hardcode the literal string `"recomputed_work"`:**
  - `measure.Demote` (`internal/compiler/measure/statistics.go:131-135`) —
    `if metric != "recomputed_work" { return VerdictObserved }`, and it is
    **structurally guarded** by `TestDemoteHasExactlyOnePromotionPassthrough`,
    which asserts the function contains exactly one `return VerdictBlocking`.
  - `session.QLT02GateEligibleMetrics()`
    (`internal/compiler/session/session_phase6_budget.go:55-57`) — returns exactly
    `[]string{"recomputed_work"}`, and it is the sole production argument to
    `QLT02LaneFromRows` (`:256`), which feeds `AuditQLT02BudgetManifest`'s
    gate-eligibility check (`:126-164`).

  **Consequence:** any newly-named metric — including a
  `recomputed_work_per_op_exponent_*` row — is **silently demoted to `observed`**
  and produces a control that can never block. Shipping that would be a
  textbook decorative gate.

  **Therefore the plan must contain an explicit task** to widen **both**
  chokepoints and re-derive the `TestDemoteHasExactlyOnePromotionPassthrough`
  guard so it still forbids a *second* promotion path while permitting the
  widened eligible-metric set. This is a named task with its own test obligation,
  **not a free consequence of adding a manifest row.** Reusing the bare name
  `recomputed_work` is not an escape: duplicate `(machine_id, metric)` rows are
  themselves a manifest validation failure (`session_phase6_budget_test.go:138-141`).

- **D-08-33 (manifest mechanics):** every row requires `machine_id`, and
  `AuditQLT02BudgetManifest` **refuses** a row whose `machine_id` has no live probe
  counterpart for the invocation (`session_phase6_budget.go:~183`). A work-count
  row is therefore re-ratified per machine even though its **value** is
  machine-independent — that independence is a property of the metric, not
  something the schema needs to encode, so **no schema change is required**.
  `ratified_at` / `ratified_by_commit` apply unchanged; ratify at the mid-phase
  gate, before the liveness law is declared final.

- **D-08-34 (corpus generator — and the footgun that decides it):** generate
  **synthetic `core.Program` values directly**, following D-07-19's
  synthetic-artifact precedent. **Not** a `.lang` source generator through the
  real parser: at the sizes needed (hundreds of functions) **parse time would very
  plausibly dominate and mask the exact curve the gate exists to see.** That is a
  concrete, stateable risk, and it is why the source-generator route is rejected
  rather than merely deprioritized.

  Verified: **no `OpCall`-aware call-graph generator exists in the production
  tree.** `internal/compiler/reduce` is an HDD *reducer* for mismatches, not a
  generator. The shape to port is the spike's throwaway
  `.planning/spikes/006-interprocedural-liveness-cost-scaling/iplive/corpus.go:107`
  (`Generate(shape, n)`) and its `scaling.go` (`Growth.ExponentInOps`). PITFALLS
  Pitfall 5 applies directly — the intraprocedural generator systematically
  under-covers cross-function cases.

- **D-08-35 (required shapes, each with its decorative argument):**
  1. **`chain`** — pure repetition, **zero sharing**. Required because the naive
     arm **died on a chain at 32 functions** in the spike; without it a corpus
     cannot separate *"exponential because of repetition"* from *"quadratic
     because of sharing."*
  2. **`diamond`** — the minimal **shared-callee** shape. Required because
     without it a gate cannot distinguish an analysis that works when every callee
     has exactly one caller from one that genuinely reuses a per-function summary
     across callers — which **is** the load-bearing IFDS claim.
  3. **`dense` / `parser-shaped`** — the shapes whose **leaf-to-relay ratio drifts
     with size**. Required because they are what makes D-08-31's operation-count
     fit *demonstrably necessary*; a corpus of only constant-ratio shapes could
     never expose why fitting against function count is unsound. `parser-shaped`
     doubles as criterion 3's named "realistic fan-out" shape.
  4. **`forward` (star, one caller / k chained callees)** — required because it is
     the shape that **isolates the program-order invariant** (4.0 flat vs
     767.0 per operation reversed). Without it, a future pass that batches or
     reorders summary derivation could reintroduce a quadratic regression that
     every other shape is blind to.

  **`tree` is dropped** — dominated by `diamond` for sharing and `chain` for
  depth, adding sweep cost with no distinct decorative argument of its own.

- **D-08-36 (lane):** a **new `lane:interprocedural-cost-scaling`** entered in
  `internal/compiler/session/risk_lanes.json` (38 lanes exist today) as a
  **changed-risk** lane firing when `internal/compiler/check`'s interprocedural
  summary/liveness code or `internal/compiler/callgraph` changes. **Explicitly not
  the default edit loop** — a multi-shape size sweep is outside what
  `wiki/compiler-and-feedback-latency.md` permits the default loop to pay. Runs at
  the mid-phase gate and pre-merge.

- **D-08-37 (cross-run cache — declared non-goal, plus a verified all-clear):**
  record in `PHASE-08-DEBT.md` **at planning time**:

  > **NON-GOAL (Phase 08):** a persistent cross-run summary cache is explicitly
  > deferred to Phase 11 / QLT-06. Spike S-006 measured that a single leaf edit
  > invalidates **92%** of a call-graph-closure-keyed cache worst-case and **43%**
  > mean on a realistic parser-shaped corpus, and **100%** on a chain. Within-run
  > memoization is mandatory and sufficient for EFF-02; cross-run caching is a
  > separate, unproven claim that must not be assumed or quoted as a production
  > win until QLT-06's callee-changes-invalidates-caller regression test gates it.

  Do **not** re-measure on production shapes — the spike's numbers are sufficient
  for planning, and the milestone's declared 2× scope-cut trigger for Phases 08/09
  makes the marginal evidentiary value a poor trade.

  **Verified all-clear, recorded as a positive finding:** `ClosureDigest` is
  **not referenced anywhere in `internal/compiler/cache`**, and `cache.Input` is
  constructed **only in tests** (`cache_test.go:26,43,80,414`, all with unrelated
  literal names such as `"fixture_source"`). **There is no QLT-06 violation in
  flight.** `07-CONTEXT.md`'s line describing `cache.Input` as taking
  `ClosureDigest` as "the named, digested input" is a **statement of planning
  intent, not shipped wiring**. The non-goal above pre-empts a future wiring; it
  does not describe a current defect. (The chaining property itself *is* well
  tested — `ControlSummaryClosureDigestChained` in `session_phase7.go`, backed by
  `originvalidate_closure_chain_test.go` and
  `corevalidate_closure_chain_mutation_test.go`.)

### Cross-cutting

- **D-08-38 (scope-cut trigger, declared now in writing per Key Lesson 4):** the
  milestone declares a **2× trigger** for Phases 08 and 09. Adopt for Phase 08:
  **if the summary-table + liveness-law work exceeds ~2× its initial plan
  estimate, the cost-gate *instrument work* (D-08-32's chokepoint widening,
  D-08-34's generator, D-08-36's lane) renegotiates into Phase 09 — never the
  criterion-1 corpus (D-08-28), never the seeded mutation-kills (D-08-16), and
  never the two-path D-07-49 fix (D-08-09).** Criterion 3's measurement is the
  gate for the riskiest assumption, so cutting it is a *deferral with a named
  landing phase*, never a silent drop; if it is cut, the liveness law **may not be
  declared final** until it lands.

- **D-08-39:** a `PHASE-08-DEBT.md` written **at planning time**, not phase end,
  in the mechanically-checked debt-register format
  `TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go`)
  enforces. It must carry, at minimum: D-08-15 (bound unreachable from source),
  D-08-26 (criterion 4's accepted-program half), D-08-27 (the OWN-09 document
  conflict), D-08-37 (the cross-run-cache non-goal), and D-08-38 (the scope-cut
  trigger).

- **D-08-40 (three items go to the mid-phase gate as open questions, not as
  settled facts):** the mandatory mid-phase gate adjudicates criteria 1 and 3 from
  code-level evidence before the liveness law is declared final. Add to its
  agenda: **(a)** OWN-09's Phase-08-vs-09 conflict (D-08-27); **(b)** whether
  criterion 4's accepted-program disclosure should become a runtime artifact and
  what it costs (D-08-26); **(c)** whether D-08-09's two admission paths both
  remain load-bearing or one is vestigial and should be retired rather than
  duplicated.

### Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all
four gray areas ("that's my boilerplate prompt saying research and synthesize
best recommendation for everything"). Every decision above is therefore Claude's
synthesis under that instruction, grounded in the four advisor research returns
and **independently verified against the shipped tree**, with three researcher
claims corrected (D-08-13's bound axis, D-08-24's span limitation, D-08-32's
metric-name chokepoints). Planner discretion remains over:

- Plan decomposition and wave ordering, subject to two hard constraints: the
  summary derivation must run **after** `callgraph.Order` proves acyclicity
  (D-08-12, mirroring D-07-38), and D-08-32's chokepoint widening must land
  **before** any manifest row claims to be a hard gate.
- Exact Go identifier and file names for the summary bits, the memo table field,
  the seam, and the generator.
- Whether the summary derivation is a method on the existing table builder or a
  sibling function it calls.
- The exact corpus file names, module paths, and generated-shape size ladder
  (subject to D-08-35's required shape list).
- Whether the ratio-stability tripwire (D-08-31) ships as a separate manifest row
  or as an in-test assertion.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Hard entry gate — read first
- `.planning/spikes/006-interprocedural-liveness-cost-scaling/README.md` — **spike
  S-006, VALIDATED, the hard entry gate on this phase's planning.** Read the whole
  investigation trail, not just the Results. Iteration 2 (why the interprocedural
  fact must enter through a program-order canonicalization pre-pass, not the
  backward transfer), iteration 5 (the program-order invariant, 192× reversed),
  iteration 6 (cache invalidation fallout), the Observability section
  (deterministic work counters, never wall-clock), and Findings 1-6.
- `.planning/spikes/006-interprocedural-liveness-cost-scaling/iplive/corpus.go` and
  `scaling.go` — the throwaway `Generate(shape, n)` and `Growth.ExponentInOps`
  shapes to **port, not import** (separate module, throwaway workbench code).

### Standing — durable, read before planning any M002 phase
- `.planning/STANDING-VERDICTS.md` — dependency verdicts (the
  zero-external-production-dependency record is a **hard constraint**; IFDS
  literature, Polonius, LLVM `CallGraphSCCPass` are all **read, don't import**),
  and the named anti-features **"whole-program borrow inference"** and **"a second
  coexisting liveness law"** — both directly load-bearing here.
- `.planning/LANGUAGE-MATURITY.md` — what the language can actually express today.
  No arithmetic, no iteration, no collections; arity 1; A-normal form; 58 programs
  averaging ~28 lines. **Read before assuming any program shape is writable** —
  D-08-15 and D-08-34 both turn on this.
- `.planning/RETROSPECTIVE.md` Key Lessons 2 (two coexisting laws — D-08-27), 3
  (fixing one of N peers — D-08-09), 4 (declare deferred scope in writing when
  decided — D-08-15, D-08-37, D-08-38, D-08-39).

### Phase-scoped research (milestone-scoped)
- `.planning/research/ARCHITECTURE.md` §4 (lines ~420-505) — the three design
  options for keeping `check` and `corevalidate` independent. **Option 1
  (signature-mediated) is adopted; option 3 (a shared library) is rejected
  outright** and must not be re-litigated (D-08-06).
- `.planning/research/PITFALLS.md` Pitfall 1 — cross-function borrow/lifetime
  unsoundness at the call boundary (criterion 1).
- `.planning/research/PITFALLS.md` Pitfall 4 (lines ~301-400) — fixpoint
  non-termination, context-sensitivity blowup, **and native stack overflow in the
  checker itself** (D-08-13 through D-08-19a).
- `.planning/research/PITFALLS.md` Pitfall 5 — the intraprocedural generator
  under-covers cross-function cases (D-08-34).
- `.planning/research/PITFALLS.md` Pitfall 6 — per-unit cache invalidation misses
  call-graph facts (D-08-37).
- `.planning/research/PITFALLS.md` Pitfall 7 (lines ~594-684) — **interprocedural
  diagnostics blame the wrong function, and the bounded cause DAG grows unbounded
  across function boundaries** (D-08-22, D-08-23).

### Project design corpus
- `wiki/compute-efficiency-constitution.md` — the cost constitution that names
  whole-program inference as the failure mode. This is the document the phase's
  riskiest assumption cites.
- `wiki/compiler-and-feedback-latency.md` — what the default edit loop may pay
  (D-08-26, D-08-36).
- `wiki/ownership-evidence-roadmap.md` — ownership evidence obligations.
- `wiki/semantic-kernel-contract.md` — the kernel contract this phase extends.

### Project state
- `.planning/PROJECT.md` — core value, constraints, Key Decisions table (including
  the "Infer local borrow ends at CFG last use" row, marked
  **"⚠️ Revisit for the interprocedural half (D-03-02)"** — that is this phase).
- `.planning/REQUIREMENTS.md` — **OWN-06** and **EFF-02** verbatim; **OWN-07**,
  **OWN-08**, **OWN-09** (Phase 09) for the downstream contract; **QLT-06**
  (Phase 11); **QLT-08**'s mutation-kill discipline. Note line 173's OWN-09
  mapping, which **conflicts with OWN-09's own text** — see D-08-27.
- `.planning/ROADMAP.md` "Phase 08" (lines ~300-345) — goal, all four success
  criteria, riskiest assumption, the gate that catches it, the **mandatory
  mid-phase gate**, and the **2× scope-cut trigger**.
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/07-CONTEXT.md` —
  the full locked call contract, D-07-01..D-07-45. Especially **D-07-02**
  (ownership convention read implicitly from the signature — why
  `ParameterContract.Mode` is a declared convention, not an observed fact),
  **D-07-05** (call-graph depth, not composition depth), **D-07-09** (the `/1`
  field set), **D-07-15/16/43** (diagnostic identity and determinism),
  **D-07-17/18** (unbounded acyclic depth; iterative explicit-stack DFS),
  **D-07-21** (both replay sites or it isn't wired), **D-07-22** (no shared
  summary package), **D-07-28** (enumerate by `Kind`, never block position),
  **D-07-34** (the pre-body signature table and its already-recorded OWN-06-style
  reading), **D-07-35** (no `Span` on `core.LinearOperation`; `check` projects
  operation IDs to spans), **D-07-38** (acyclicity proven before any chaining),
  **D-07-42** (unexported fault-injection seams).
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md` —
  **D-07-49 is this phase's assigned entry defect**; **D-07-47** is the precedent
  D-08-15 follows; **D-07-51/WR-02** records that `spanByOperationID` already
  exists and is already used for per-edge spans (directly reusable for D-08-22).

### Live code this phase edits (verified in-tree during discussion)
- `internal/compiler/check/check.go:1049` — `derivePlaceLoans`, the forward
  canonicalization pre-pass D-08-07 extends. Already carries the `latestLoan` /
  `parentLoan` reborrow-chain shape the `OpCall` branch mirrors.
- `internal/compiler/check/check.go:1142-1225` — `loanLivenessFixpoint`: the
  worklist D-08-13 bounds, its `work` counter, and its **native-recursion
  pre-walk at `:1157-1180`** that D-08-19a converts to an explicit stack. The
  bare `check.cfg_back_edge` string is at `:1163` (D-08-19b).
- `internal/compiler/check/check.go:2979` — `computeLoanLastUses`, the AST-shadow
  path missing the `"call"` case (D-08-09); called at `:1328` and `:2529`.
- `internal/compiler/check/check.go:~577-600` — `callSignatureTable` /
  `buildCallSignatureTable`, the immutable body-blind table D-08-04 extends; its
  doc comment carries the "built AFTER every function's body is checked" precedent
  D-08-05 relies on.
- `internal/compiler/check/check.go:552-572` — `calleeContract` /
  `buildCalleeContracts`, the **separate, narrower, AST-derived** pre-body table.
  Not the same object as `core.FunctionSignature`; today's admission gate reads
  only this one.
- `internal/compiler/check/check.go:493,1928` — `callReturnTypeDerivationSeam`,
  the **unexported** seam shape D-08-16 copies (test usage at
  `check_test.go:3412`).
- `internal/compiler/check/check.go:1430,2793-2830,511,814,1508` — the existing
  `ownership.*` loan diagnostics and `borrowConflictDiagnostic`'s cause/repair
  shape (D-08-20, D-08-22, D-08-25).
- `internal/compiler/check/check.go:1870` — `resolveCallBinding`.
- `internal/compiler/callgraph/callgraph.go:294` — `Order`, returning reverse
  postorder over a proven-acyclic graph. **D-08-03 consumes this; it is free.**
  `:61,71` — `MaxCycleCauses` / `TruncatedCycleBound`, the local-declaration
  precedent D-08-23 follows.
- `internal/compiler/core/core.go` — `Interface`, `FunctionSignature`,
  `ParameterContract`, `ReturnContract`, `ForeignReach`. The absence of a
  `UsesParam` field is the fact D-08-01 turns on.
- `internal/compiler/diagnostic/diagnostic.go:17-26` — `Span{Start, End int}`,
  **no file field** (D-08-24); `:96-160` — `Causes` in ID identity (D-08-18,
  D-08-23).
- `internal/compiler/measure/statistics.go:106-145` — `Samples` (bare `[]int64`,
  unit-agnostic), `Summary`, and **`Demote`'s hardcoded
  `if metric != "recomputed_work"`** at `:135` (D-08-30, D-08-32).
- `internal/compiler/session/session_phase6_budget.go:55-57` —
  `QLT02GateEligibleMetrics()` returning exactly `[]string{"recomputed_work"}`;
  `:126-164` — `AuditQLT02BudgetManifest`; `:256` — the sole production call site
  (D-08-32, D-08-33).
- `internal/compiler/session/qlt02_budget_manifest.json` — the hard/observed split
  D-08-29 extends.
- `internal/compiler/session/risk_lanes.json` — 38 existing lanes; D-08-36 adds
  one.
- `internal/compiler/protocol/protocol.go:145-166` — `InterfaceSummary`'s
  "deliberately lossy… never a second copy of the `/1` artifact" anti-pattern
  (D-08-26); `:223-249` — `ExplainSummary`, synthesized only from an existing
  diagnostic's causes (D-08-26); `:137-144` — the import-cycle note (D-08-23).
- `internal/compiler/cache/cache_test.go:120` —
  `TestCacheExportedSurfaceStoresNoVerdict` (D-08-26, D-08-37).
- `internal/compiler/pathoracle/pathoracle.go:51,59` —
  `TerminatorKindsOverride`, **exported**; the shape D-08-16 deliberately does
  *not* copy.

### Fixtures
- `testdata/phase07/relay_escort_witness.lang` — **the criterion-1 refusal
  fixture** (D-08-28.3). Its header records it as D-03-02's interprocedural half,
  checking clean today only because of the missing `"call"` case.
- `testdata/phase3/shared_shared_accept.lang`, `testdata/phase3/public_view.lang`
  and the `public_view_*` family — the existing borrow/take binding grammar the
  twin pairs reuse unchanged.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`callgraph.Order` (`callgraph.go:294`)** — already returns reverse postorder
  over a proven-acyclic graph. S-006 finding #2: this is *exactly* the ordering
  the summary derivation needs, so callee-before-caller derivation costs nothing
  extra and no call-graph-level fixpoint is required at all.
- **`callSignatureTable` (`check.go:~577`)** — already immutable, already
  body-blind by construction (holds `core.FunctionSignature`, which has no
  `Linear`/`Match` field to follow even by pointer), and **already carries a
  body-derived bit** (`Callable`). The two liveness bits are the same shape of
  fact, so this is an extension of an existing pattern rather than a new channel.
- **`derivePlaceLoans` (`check.go:1049`)** — already implements the exact
  loan-identity aliasing mechanism the interprocedural canonicalization needs
  (`latestLoan` / `parentLoan` reborrow chains), with a generic fallthrough that
  already produces the correct "fresh identity" behaviour for the owned-return
  case. The `OpCall` branch is additive.
- **`callReturnTypeDerivationSeam` (`check.go:493`)** — the in-tree **unexported**
  fault-injection seam pattern, with an existing test that flips and restores it.
- **`spanByOperationID`** — already exists and is already used for per-edge cycle
  spans; directly reusable for projecting the call-site span into D-08-23's
  cause 2, given D-07-35 keeps spans off `core.LinearOperation`.
- **`measure.Samples` / `Summary`** — genuinely unit-agnostic (`[]int64`), so the
  p50/p95/CoV protocol applies to work counts with no modification.
- **The existing hard/observed manifest split** — `recomputed_work` hard beside
  `elapsed_ns` observed is the precedent D-08-29 extends rather than invents.
- **`loanLivenessResult.work` / `ownershipSupport.FixpointWork`** — a deterministic
  work counter already threaded through `check`'s call sites (`:1328`, `:2529`).
  It is computed today but does not reach any manifest; wiring it is the gap.

### Established Patterns
- **Consumption/production two-tier signature access.** D-07-34 already
  established that "admitted from the callee's signature alone" governs the
  per-call consumption path, with the table itself built once, post-body-check,
  ahead of admission. D-08-05 restates this reading for the liveness bits rather
  than inventing a new liberty.
- **Peers share an inert code string, never a derivation.** `core.call_graph_cycle`
  is emitted by both layers with their own messages and their own traversals; a
  shared *implementation* is rejected outright.
- **Mechanism-local vs peer-shared code namespaces.** `check.cfg_back_edge` stayed
  `check.*` (one deriver); `core.call_graph_cycle` is `core.*` (two peers, shared
  witness type from day one). D-08-21 places this phase's code by that same rule
  and schedules its promotion.
- **Bounded output with stable truncation codes**, declared in the package that
  owns them rather than in `protocol`, to dodge the
  `check → protocol → interp → [test] → check` import cycle.
- **Duplicated straight-line vs. branch replay.** Every new case lands **twice**,
  per engine — the reason D-08-09 requires both admission paths.
- **Fault-injection seams are unexported package-level vars exercised by
  same-package tests** (D-07-42), not exported globals.
- **Iterative explicit-stack graph traversal**, never native recursion, wherever
  depth is program-controlled (D-07-18).

### Integration Points
- `buildCallSignatureTable` — the producer to extend with the two bits.
- `callgraph.Order` — the ordering the derivation consumes; the derivation must
  run strictly after it proves acyclicity.
- `derivePlaceLoans` + `blockLoanLiveness` — the two clauses that make the
  interprocedural fact real, one forward and one backward.
- `computeLoanLastUses` **and** the `derivePlaceLoans` path — the two admission
  paths D-07-49 must land in.
- `measure.Demote` + `session.QLT02GateEligibleMetrics()` — the two hardcoded
  chokepoints D-08-32 must widen, or the cost gate cannot block.
- `qlt02_budget_manifest.json` + `risk_lanes.json` — where criterion 3's bound and
  its lane are recorded.

</code_context>

<specifics>
## Specific Ideas

- **The single most actionable finding of this discussion**, and the direct
  analogue of Phase 07's diamond insight: **a criterion-1 twin pair that varies
  the caller's shape rather than the callee's declared contract field would pass
  with the interprocedural law entirely deleted**, because the existing
  intraprocedural `ownership.*` law already refuses same-function violations. The
  twin must differ in **exactly one callee contract field**, and the corpus
  harness should assert that mechanically (the AST diff between refusing and
  accepting members touches only the callee's declared return/parameter type).
- **The second-most actionable:** naming the cost metric anything other than
  `recomputed_work` silently produces a gate that can never block, because
  **two** independent chokepoints hardcode that literal string. A reviewer reading
  only the manifest JSON would see a `"gate_type": "hard"` row and conclude the
  gate is live. Widening both, with the promotion-passthrough guard re-derived, is
  a named task.
- **Putting the interprocedural clause in the backward transfer function is the
  silent-failure design.** It passes every accept-case test and never fires on the
  refuse case, because the backward walk reaches the `move` before the call that
  establishes the alias. S-006 hit this in iteration 2 and only the fixtures
  caught it.
- **A negative-control twin pair varying only `Fails` / `Foreign.*`** is cheap and
  is the only thing that proves the law correctly *ignores* the fields it must not
  read — the "which fields are **not** named" half of criterion 4.
- **`relay_escort_witness.lang` already exists and already refuses at `lang check`
  through the Phase 07 peer** while `check` itself still admits it. That
  asymmetry is the phase's entry condition, and flipping it is a concrete,
  pre-authored assertion rather than a fixture to invent.
- **Prior art worth reading, not importing:** IFDS/tabulation (Reps-Horwitz-Sagiv)
  — the memo table *is* the algorithm's definition, not a tuning knob; Rust NLL vs
  Polonius, and specifically why the Datalog formulation had no path to
  stabilization while the re-engineered version scales; rustc's `measureme` and
  perf.rust-lang.org gating on **instruction counts rather than wall time** for
  precisely this noise reason; LLVM's compile-time-tracker; Z3's deterministic
  `rlimit` chosen over a timeout for reproducibility; Clang's
  `-Wthread-safety-analysis` as the closest analogue to an annotation-mediated
  cross-function answer, including how it blames the use site rather than the
  declaration; LSP `DiagnosticRelatedInformation` and TypeScript's "the expected
  type comes from property X declared here" as the cross-file blame convention;
  Swift `.swiftinterface` / library evolution and OCaml `.cmi` CRC chains for
  declared-vs-inferred summary boundaries.

</specifics>

<deferred>
## Deferred Ideas

- **`corevalidate`'s independent *liveness* peer, and D-03-02 closure** —
  Phase 09 (OWN-07, OWN-08). Held by `PHASE-07-DEBT.md`'s D-03-02 entry. Its
  mechanism is a reachability closure, deliberately not a worklist; it must not
  import `check` (D-08-06).
- **Promotion of `check.interprocedural_loan_liveness` to `core.*`** — Phase 09,
  at the moment the peer independently re-derives the fact (D-08-21).
- **OWN-09's retirement of the intraprocedural liveness law** — Phase 09 under
  `REQUIREMENTS.md`'s mapping, though OWN-09's own text says Phase 08. **Conflict
  flagged, not resolved** (D-08-27); on the mid-phase gate agenda.
- **Criterion 4's accepted-program disclosure as a runtime artifact** — deferred
  with a declared reason (D-08-26). No existing vehicle can carry it: `explain`
  needs a diagnostic, `cache` cannot hold a verdict, and a sibling schema
  contradicts `InterfaceSummary`'s recorded anti-pattern. Cost question goes to
  the mid-phase gate.
- **A persistent cross-run summary cache** — Phase 11 / QLT-06, with the non-goal
  and the spike's invalidation numbers recorded now (D-08-37).
- **Widening `diagnostic.Span` with a file/module field** — not this phase
  (D-08-24). It would be a one-way published-schema change on research-only
  justification; revisit when a cross-module callee span is genuinely needed.
- **A third summary bit, or per-arm summary bits** — not this phase (D-08-10).
  Re-open when arity widens past 1 (D-07-07) or when `Result` match arms can
  independently borrow-vs-move a parameter (Phase 12).
- **Publishing the two liveness bits into `lang.interface/1` or a `/2`** —
  deliberately not done (D-08-04). The in-memory table is the cheap direction and
  keeps the published schema untouched; publishing later remains available as a
  considered decision.
- **The interpreter's fixed documented call-stack ceiling** — SEM-08, Phase 10.
  D-07-17 already decided call-graph *depth* is unbounded provided the graph is
  acyclic; spike S-007 informs whether that ceiling is a constant or a declared
  budget.
- **Re-measuring cache invalidation fallout on production-shaped corpora** —
  explicitly not worth the sweep cost (D-08-37); the spike's 92%/43%/100% stand.
- **A `tree`-shaped corpus** — dropped as dominated by `diamond` and `chain`
  (D-08-35), not deferred.
- **Match-arm call fixture for liveness** — recommended regression coverage, not a
  gate requirement (D-08-28.5).

</deferred>

---

*Phase: 08-interprocedural-loan-liveness-in-check*
*Context gathered: 2026-09-09*
*Advisor mode: 4 parallel `gsd-advisor-researcher` agents, `minimal_decisive` calibration*
*All researcher claims re-verified against the shipped tree; 3 corrected (D-08-13, D-08-24, D-08-32)*

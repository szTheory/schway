# Phase 09: Peer Re-Derivation and D-03-02 Closure - Context

**Gathered:** 2026-09-10
**Status:** Ready for planning

<domain>
## Phase Boundary

A second, independently-implemented admission layer (`corevalidate`, a
reachability-closure validator) reaches the same interprocedural loan-liveness
answer as the first (`check`, an iterative worklist), **and the agreement is
proven load-bearing by a seeded fault that makes the two diverge**. D-03-02 —
the one debt item M001 knowingly carried past its milestone — closes in both
layers. The checker ends the phase with exactly one loan-liveness decision
point.

Concretely, this phase delivers six things:

1. **`corevalidate`'s own interprocedural liveness derivation** — peer-derived
   liveness bits produced by forward set-propagation over `Linear.Operations`,
   folded into the existing `peerPostorder`/`peerSignatures` substrate, and
   consulted by `buildLoanChainIndex` before it propagates a loan across an
   `OpCall` (OWN-07).
2. **Full closure of D-07-33's narrowed peer** — the peer re-derives all four
   `PublishProblemsFor` refusal classes, not just `core.origin_omitted`, so
   `Callable` agreement stops being vacuous on three of them (OWN-08).
3. **A single loan-liveness decision point in `check`** — `computeLoanLastUses`
   and its summary-blind shadow-place scaffolding are deleted; lowering makes no
   ownership decisions; one pass over the assembled `core.Program` decides
   (OWN-09).
4. **A zero-divergence shadow-run differential** over cycle, diamond, and
   deep-chain call-graph shapes, landed in the same plan as the peer, with a
   seeded endpoint-level fault plus the companion assertion that discriminates
   independence from mere change-detection (TRU-04).
5. **Call-site ownership transfer proven to have one meaning** across `check`
   and `corevalidate`, with call-site convention override proven non-expressible
   at the source layer and fail-closed at the core layer (OWN-05, partial — the
   `interp` peer is Phase 10).
6. **QLT-07's committed-vs-stretch status decided against a pre-registered
   threshold**, and — if committed — M001 Phase 3's loan-liveness Nyquist debt
   closed in a new `09-VALIDATION.md` section.

**Not this phase** (each has its own phase and its own gate): `interp`'s third
independent derivation of ownership transfer and the bounded interpreter call
stack (Phase 10, verified there by TRU-03); multi-function C emission and
interprocedural `-O3`/LTO equivalence (Phase 11); cross-run summary caching /
QLT-06 (Phase 11, per D-08-37); `Result` payloads (Phase 12); the agent loop
(Phase 13).

</domain>

<decisions>
## Implementation Decisions

All eight gray areas were researched in advisor mode (eight parallel
`gsd-advisor-researcher` agents, `minimal_decisive` calibration) under the
developer's standing mandate to fan out across stakeholder-role lenses, run an
adversarial pass, draw on cross-ecosystem prior art, and synthesize one-shot
recommendations. The developer directed that the synthesized recommendation be
adopted in every case.

**Every factual claim below was independently re-verified against the shipped
tree by the orchestrator before being locked.** Three researcher claims required
correction; each correction is recorded inline at the decision it changes
(D-09-05, D-09-19, D-09-24). Where a researcher and the tree disagreed, the tree
won.

**Two decisions here formally reverse commitments recorded in Phase 08**
(D-09-08 reverses D-08-41's disposition; D-09-30 supersedes D-08-21's promotion
commitment). Both reversals are stated as reversals, with the superseded text
named, so no future reader has to discover the change.

### `corevalidate`'s independent liveness derivation (OWN-07)

- **D-09-01 (the substrate already exists, and this is the finding that shapes
  the area):** `corevalidate` already ships a working peer-derivation substrate
  for a structurally analogous problem — Phase 07's signature/`Callable` peer.
  Verified in tree: `v.peerPostorder` (`corevalidate.go:138`, appended at
  `:472`) is built as a byproduct of `checkCallGraphAcyclic`'s **own**
  independent DFS, and `v.peerSignatures` (`:126`) is refined in a postorder
  loop (`:505-547`) that reads `v.peerSignatures[calleeID]` before the caller —
  i.e. an existing memoized, callee-before-caller peer table. The liveness bits
  are the same shape of fact. **Extend that substrate; do not build a second
  one.**

- **D-09-02 (derivation direction is the independence, not the file layout):**
  the peer's liveness bits are derived by **forward set-propagation** over
  `function.Linear.Operations`, matching the idiom `peerReturnDerivesFromBorrow`
  (`corevalidate.go:2099`) and `peerParameterEscapesOwned` (`:2061`) already
  use, and structurally opposite `check`'s backward memoized walk
  (`deriveFunctionUsesParam`, `check.go:686`). Independence is proven by
  **derivation method plus import boundary**, never by file or package
  location. A new `internal/compiler/corevalidate/...` subpackage buys nothing
  and doubles the ordering surface Phase 09's own risk statement warns about.

- **D-09-03 (the concrete defect, and its fixture-backed close):**
  `buildLoanChainIndex` (`corevalidate.go:1133`) today propagates
  `parent[TargetID] = SourceID` through **every** `core.OpCall` with no
  callee-signature consultation. It must consult the peer-derived
  "returns-borrow-of-param" bit before propagating. Landing this should retire
  **`testdata/phase07/twin_a_accept.lang`** and
  **`testdata/phase07/relay_depth2_accept.lang`** from
  `peerDivergenceExpected` (`internal/compiler/session/session_peer_gate_test.go`)
  — turning D-08-40's declared debt into a test-visible fact rather than a
  comment. **Retiring both entries is a required assertion of this phase**, not
  a side effect.

- **D-09-04 (placement):** the new derivation functions live **inside package
  `corevalidate`**, either in `corevalidate.go` or a sibling file in the same
  package (planner's discretion — `corevalidate.go` is already ~137KB, so a
  sibling `corevalidate_peer_liveness.go` is reasonable). The doc comment must
  state explicitly that what is reused from Phase 07 is **substrate** (ordering
  plumbing plus memoization), never **derivation** — this is precisely the
  conflation a reviewer will probe.

- **D-09-05 (CORRECTION — the shared-substrate line is already drawn and
  enforced, which answers the question the researcher left open):** the peer
  **may not** reuse `callgraph.Order`, and this is not a judgement call — it is
  a shipped test. `corevalidate_endpoint_internal_test.go:107` forbids
  `corevalidate` from importing `compiler/check`, `compiler/ast`,
  `compiler/originvalidate`, **and `compiler/callgraph`**. Verified: production
  `corevalidate.go` imports only `crypto/sha256`, `encoding/hex`,
  `encoding/json`, `fmt`, `reflect`, `sort`, and
  `github.com/codename-lang/lang/internal/compiler/core`. **Four** separate
  import-guard tests already exist (`corevalidate_cycle_peer_test.go:146`,
  `corevalidate_endpoint_internal_test.go:107`,
  `corevalidate_exclusive_test.go:216`, `corevalidate_test.go:148`).
  Consequence for planning: sub-question (c) of this area — "what structurally
  prevents drift" — is **already answered by shipped machinery**; the phase
  extends the guards' coverage to the new code rather than inventing a fourth
  enforcement mechanism.

- **D-09-06:** the peer's own bits are **within-run only**, rebuilt per
  `Validate` call, never persisted — mirroring D-08-12 on the `check` side and
  keeping the peer clear of QLT-06 (D-08-37's cross-run non-goal binds both
  layers, not just `check`).

### OWN-09 — the single decision point (and what "one law" actually means)

- **D-09-07 (THE CORRECTION THAT REDEFINES THIS REQUIREMENT — verified
  directly, and every planner must read it before touching OWN-09):** there was
  **never** a second loan-liveness law. Verified in tree: `computeLoanLastUses`
  is at **`check.go:3743`** (not `:2979` — `08-CONTEXT.md`'s D-08-09 line
  reference is **stale**, recorded here rather than left to mislead), and it
  calls the **same** `loanLivenessFixpoint`, fed (a) synthetic
  `shadow:place:parameter` / `shadow:place:<n>` operations built straight from
  `ast.LinearBody` before any `core.Function` exists, and (b) a permanent
  zero-value `interproceduralSummaryTable{}` — literally
  `loanLivenessFixpoint("shadow", []cfgBlockSpec{block}, interproceduralSummaryTable{}, diagnostic.Span{})`
  and the matching `materializeLoanEndpoints("shadow", ...)` call.

  **There is one algorithm with two callers, one of which is deliberately
  summary-blind.** OWN-09's "the intraprocedural law is DELETED, not left
  coexisting" therefore cannot mean deleting an algorithm — it means **deleting
  the early, summary-blind call site and its shadow-place scaffolding, and
  moving loan-liveness decision time from per-function lowering to
  post-assembly.** Any plan that reads OWN-09 as "delete a law" will either find
  nothing to delete or delete the wrong thing.

- **D-09-08 (this REVERSES Phase 08's D-08-41 disposition, deliberately and in
  writing):** D-08-41 adjudicated `computeLoanLastUses`' summary-blindness as
  "an accepted, permanent, disclosed scope limitation… Landing phase: Not
  scheduled." **That disposition is superseded.** `computeLoanLastUses` is
  **deleted** in Phase 09.

  The reason is the harm D-08-41 itself documented: the Pattern B twin pair
  `twin_b_refuse.lang` / `twin_b_accept.lang` cannot demonstrate a differing
  end-to-end CLI verdict, because the summary-blind shadow path refuses **both**
  members identically with `ownership.move_while_borrowed` before the
  interprocedural pass ever runs. A summary-blind decision point is actively
  **masking** the law the milestone exists to prove. Making it summary-aware
  instead would leave two call sites into one algorithm — exactly the
  coexistence OWN-09 forbids. Re-scoping "the intraprocedural law" to mean
  something narrower is goalpost-moving with no basis in the code.

  Cross-ecosystem precedent: rustc's `-Zborrowck=migrate` two-law period was
  explicitly **temporary**, ended at an edition boundary with future-compat
  lints as the off-ramp — not permanent coexistence. The status-quo option here
  is the documented "strangler fig that never finishes."
  — **Reversibility:** costly — restoring the shadow path would mean
  reconstructing the `shadow:place:*` synthesis and re-threading two call sites
  through lowering. Deliberately so: the point is that it does not come back.

- **D-09-09 (shape of the restructure):** `check` becomes a strict two-pass
  pipeline — lower AST→core with **no ownership decisions during lowering**,
  then decide loan liveness once over the assembled `core.Program` via the
  already-interprocedural fixpoint fed real summaries. The two call sites to
  retire are `analyzeStraightLine` (`check.go:3287`) and `analyzeArmBody`
  (`check.go:2074`), whose hand-rolled `activeLoans`/`expiringLoans` state
  machines currently consume `computeLoanLastUses`' index to raise
  `ownership.move_while_borrowed` (`check.go:2176`, `:3392`) during lowering.

- **D-09-10 (ORDERING — build-then-delete, and this is a safety constraint, not
  a preference):** (1) land `corevalidate`'s contract-aware peer **fully first**,
  so a second independent detector exists before anything is removed; (2) get
  TRU-04's zero-divergence gate green; (3) **only then** restructure and delete,
  in the same commit that flips lowering to defer to the single pass. Never an
  intermediate state where neither path is wired to decide. Deleting the sole
  decision point before a second independent detector validates the surviving
  law's completeness would violate RETROSPECTIVE Key Lesson 2 directly — this is
  D-08-27's own argument, and Phase 09 is the phase where its precondition is
  finally met.

- **D-09-11 (the corpus is too small to trust for a subsumption claim):** "zero
  divergence over existing fixtures" is a **weak** claim here — ~58 programs
  averaging ~28 lines, no loops, no arithmetic, arity 1. The shadow-run in step
  (2) of D-09-10 must be **deliberately widened past the committed corpus** with
  synthetic/mutation-generated programs that try to reach every `ownership.*`
  code through the AST-shadow path specifically. Otherwise "the deleted path's
  extra refusals were simply unreachable" is asserted, not proven.

- **D-09-12 (the published codes SURVIVE — unify the derivation, not the API):**
  all five `ownership.*` codes (`move_while_borrowed` `check.go:1430`,
  `borrow_conflict` `:2793` — the only repair-bearing one,
  `use_after_move` `:511`, `borrow_requires_share` `:814`,
  `transfer_requires_take` `:1508`) and `check.interprocedural_loan_liveness`
  stay exactly as they are. `cmd/lang-repair`'s dispatch and `lang explain` are
  unaffected as long as each code's **trigger condition** is preserved by the
  surviving single pass. Retiring a code is one-way and unforced here.
  — **Reversibility:** one-way if violated — diagnostic codes are published
  agent-facing API.

- **D-09-13 (the real risk of the restructure, named so it is designed for):**
  **diagnostic order and identity.** Today a per-function early exit can report
  a different *first* error than a whole-program late pass would (e.g.
  `name.unknown` in fn A vs `move_while_borrowed` in fn B). `Code + Span +
  Causes` fold into an identity hash
  (`internal/compiler/diagnostic/diagnostic.go:96-112`), so emission-order
  changes are observable. `LoanFinalUse` facts also move from pre-emission to
  post-lowering, touching `checkLinear`/`checkBranch`'s emission contract. This
  needs an explicit ordering-stability assertion, not a corpus replay.

- **D-09-14 (the document conflict, closed):** `.planning/REQUIREMENTS.md:173`'s
  **Phase 09 mapping stands as operative**. Phase 08 is shipped (`4ab5c65`);
  reopening it to relitigate OWN-09 is not viable. **OWN-09's own requirement
  text is what gets corrected** to say Phase 09, exactly as Phase 08's mid-phase
  gate anticipated ("Phase 09's own planning is where OWN-09's text should be
  corrected to match"). This closes D-08-27.

### D-07-33 — full closure of the narrowed `Callable` peer (OWN-08)

- **D-09-15 (in scope — and it was never a cut candidate):** D-07-33's own debt
  row already records **"Landing phase: Phase 09"** and "in the same plan that
  builds the peer's own origin recomputation." It is not sitting outside the
  scope-cut order awaiting a decision; it was committed in writing at Phase 07
  planning time per Key Lesson 4. REQUIREMENTS.md's cut order not listing it is
  consistent — it was never proposed for cutting. It needs **execution**, or a
  **new** cut decision with its own justification.

- **D-09-16 (the literal OWN-08 reading is REJECTED, and this is the adversarial
  crux of the area):** OWN-08 says "an exported borrow-derived return with no
  declared origin is refused… in both admission layers," and "no declared
  origin" maps exactly to `core.origin_omitted` — the one class the peer already
  covers. A literal reading therefore says OWN-08 closes without touching the
  other three classes.

  **That reading must be rejected.** `Callable` is defined project-wide as the
  **full** D-04-03 predicate (publication safety, i.e. `PublishProblemsFor`
  returns no problems across all four classes). D-07-33 itself names the
  narrowing as `escape:coordinated-source-to-core-false-claim` promoted from
  origins to the call contract. Declaring OWN-08 closed while three of four
  refusal classes stay **provably vacuous** on the peer side is Key Lesson 3's
  failure mode verbatim, and it would be D-03-02 repeating its own M001 history
  — deferred under a closure claim, for a second consecutive milestone, on the
  same item. **OWN-08 is honestly closed only when all four classes are
  independently re-derived by the peer.**

- **D-09-17 (the fear was larger than the work — verified):**
  `peerReturnDerivesFromBorrow` (`corevalidate.go:2099`) is **already** a forward
  set-propagation walk with `paramTrace` and `derived` maps, materially
  different from both `check` (which has no equivalent) and
  `originvalidate.walkReturnOrigin`'s backward per-return walk. It currently
  tracks **presence only** (`derived map[string]bool`) — never access mode,
  never paths. The extension is therefore an **access-mode payload on an
  existing walk**, not a file-sized duplicate of `originvalidate.go`.

- **D-09-18 (arity 1 collapses one of the three classes):** because arity is 1,
  `PublicOrigin.Paths` has exactly one possible value (`function.Parameter.Name`)
  whenever any origin is derivable at all — so "understatement of paths"
  degenerates to a boolean the peer **already computes** via
  `peerReturnDerivesFromBorrow`. The genuinely new derivation work is **access
  mode** (`shared`/`exclusive`), not path containment. **Re-open the moment
  arity widens past 1 (D-07-07).**

- **D-09-19 (CORRECTION — the peer performs a CONTAINMENT check, not a
  recomputation, and the distinction is what preserves independence):** the peer
  must differ from **both** `check` and `originvalidate` — `originvalidate` is
  itself already a third layer. The peer therefore compares its **declared**
  origin against its **own** forward-derived access/presence facts (a
  containment/refinement check), rather than re-implementing
  `RecomputeOrigin`/`RecomputeOriginPerReturn`'s backward combination law
  (`originvalidate.go:234`, `:133`). First-hop-wins direction must be **proven**
  equivalent to `RecomputeOriginPerReturn`'s rule by seeded-fault tests, not
  assumed.

- **D-09-20 (the foreign class is bounded, with a named precedent):** the peer's
  forward walk "deliberately never crosses an `OpForeignCall` hop." Crossing it
  does **not** reopen spike 005's PARTIAL FFI-provenance surface:
  `checkForeignOriginOmitted` (`originvalidate.go:290-346`) is already a ~55-line
  backward walk scoped to **one function's own** `ForeignContract.Alias` — proof
  that a narrow, function-local foreign-crossing check is possible without
  depending on spike 005's completion. The peer mirrors that **technique in
  spirit, never in code**.

- **D-09-21 (the fallback, declared now per Key Lesson 4):** if execution — not
  plan-time analysis — discovers the `OpForeignCall` extension is genuinely
  unbounded (e.g. it touches multi-hop foreign chains spike 005 did not clear),
  then and only then: split D-07-33 into two debt rows, close the two body-only
  classes (`origin_understated`, `origin_access_mismatch`) in Phase 09, cut
  foreign-origin-omitted peer re-derivation with a **named landing phase**, and
  **amend REQUIREMENTS.md to carve that class explicitly out of OWN-08's
  scope** — because OWN-08 is on the never-cut list, so narrowing any slice of
  its supporting peer requires explicit written sign-off, not a silent scope
  read.

### TRU-04 — the shadow-run zero-divergence gate

- **D-09-22 ("recursion" is a CATEGORY ERROR as written, and the disposition
  must be written into TRU-04's own text):** TRU-04 names "recursion" as a
  required call-graph shape, but Phase 07 shipped cycle refusal — `callgraph`'s
  iterative three-color DFS refuses `core.call_graph_cycle` **before any
  liveness derivation runs**. There is no admitted recursive program in this
  compiler, so a *liveness* differential over recursion cannot exist.

  **Disposition, to be written into TRU-04's definition-of-done (not left
  implicit in a test file):** TRU-04's "recursion" shape is satisfied as a
  **cycle-peer differential** — both `check`'s callgraph refusal and
  `corevalidate`'s independent refusal must agree on refusal-or-not **and** on
  the witness (cycle nodes / diagnostic code), for self-cycles, mutual cycles,
  and indirect (3+ hop) cycles. `corevalidate_cycle_peer_test.go` already builds
  exactly these synthetic programs. Writing this down matters because a future
  reader skimming "recursion" + "liveness differential" will otherwise assume
  the compiler admits recursive programs and go hunting for a fact that
  structurally cannot exist.

- **D-09-23 (vehicle — extend the trusted gate, feed it more shapes):** the
  primary vehicle is the existing corpus-wide
  `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` +
  `peerDivergenceExpected` (`session_peer_gate_test.go`), which already asserts
  the divergence set is **exact in both directions** (a new undeclared
  divergence fails; a stale entry that no longer diverges also fails) and has a
  track record of separating closed from declared-and-open. Feed it
  **additionally** from Phase 08's `generateCallGraphCorpus` (`chain`,
  `diamond`, `dense`, `parser-shaped`, `forward`) as a synthetic
  `core.Program` source. **Do not** stand up a second, independently-truthed
  harness — two gates that can drift on what "divergence" means is the same
  two-derivations-one-truth failure this phase exists to prevent. Keep the cost
  gate's growth-exponent fitting entirely separate: same generator, disjoint
  consumers. The real corpus alone is insufficient — D-08-41 already proved it
  structurally cannot demonstrate an intended split verdict.

- **D-09-24 (CORRECTION — the seam precedent is in `check`, and it is the exact
  companion-assertion shape, running the OPPOSITE direction from the
  recommendation):** verified in tree, the precedent is `verifyCallableRefusalSeam`
  (`check.go:1183`, read at `:1250`) with
  `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`
  (`check_test.go:3322`) — **in `check`, not `corevalidate_mutation_matrix_test.go`**
  as researched. That test disables **`check`'s** refusal and asserts
  **`corevalidate` still refuses**. So the shipped precedent already
  demonstrates the discrimination pattern, in the direction the research did not
  recommend. **Land seams in BOTH directions** — the precedent's direction is
  free (it already exists and generalizes), and the `corevalidate`-side seam is
  the sharper test of *this* phase's stated risk.

- **D-09-25 (the seam design, and the argument for why it discriminates —
  this is the section a planner must not compress):** a seeded fault that merely
  makes both peers diverge proves nothing. That is precisely the false positive
  **Knight & Leveson (1986)** documented for N-version programming: independently
  written implementations correlate on faults because the spec, the requirements
  understanding, or (here) the code path is shared. A seam inside code shared by
  both peers would also cause divergence.

  **The gate is therefore not "seed a fault, see if output changes." It is:
  seed a fault in peer X only, and assert peer Y's independently-derived answer
  for the SAME endpoint fact is UNCHANGED, and that the two now DISAGREE — plus,
  with the seam off, that they AGREE.** The companion assertion is what
  discriminates; a seeded fault without it is exactly the "vacuous agreement"
  ROADMAP names as the risk. **State this reasoning in the test's own doc
  comment**, following `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`'s
  existing precedent.

  What makes the seam structurally sound: the seams are unexported
  package-level vars, so a `corevalidate`-side seam is unreachable from
  `check`'s package by construction. The seam's very scoping is the proof the
  code paths are separate — reinforced by the four shipped import guards
  (D-09-05).

- **D-09-26 (mechanism):** unexported package-level bool + deferred restore,
  matching `loanLivenessBoundSeam` and `callReturnTypeDerivationSeam`. **Not**
  an external mutation-testing tool — those produce a broad score, not a single
  targeted witness-carrying fault, which is the wrong shape for a criterion that
  names ONE seeded fault as the gate. The existing mutation matrices
  (`corevalidate_mutation_matrix_test.go`,
  `corevalidate_closure_chain_mutation_test.go`) are a **supplement** under
  QLT-08, not a replacement for criterion 2's single discriminating fault.

- **D-09-27 ("same plan" and incremental discipline do not conflict):** write
  the differential **test-first against the known-divergent state** (the
  `twin_a_accept.lang` / `relay_depth2_accept.lang` entries that exist today),
  land the mechanism change, then the same test flips to asserting agreement —
  all inside one plan. That is ordinary red-green TDD scoped at the plan
  boundary. The gate is never "written after the thing it gates," and `main` is
  never red, because the plan is the atomic unit. Ordering tasks within the plan
  is the thing to get right; a skipped/pending assertion until the final task is
  the failure mode to avoid.

- **D-09-28 (CORRECTION + a cost the research did not price — the peer needs its
  own bound, and getting a new metric name wrong silently makes it decorative):**
  verified, `qlt02_budget_manifest.json` **already carries**
  `recomputed_work_growth_exponent` alongside `recomputed_work`, `elapsed_ns`,
  and `output_bytes`; **both** chokepoints were **already widened by Phase 08**
  (`measure.Demote`, `session.QLT02GateEligibleMetrics()`), with
  `TestGateEligibleMetricSetsAgreeAcrossChokepoints` proving they agree as sets
  — so no re-widening is needed to reuse that name.

  But the peer's algorithm is a **reachability closure**, not a bounded worklist
  fixpoint, so it is **not the same cost curve** and must not borrow `check`'s
  ratified bound (measuring one of N independent peers is not measuring both).
  It needs its own bound. And because duplicate `(machine_id, metric)` rows are
  themselves a manifest validation failure, a distinct bound requires a
  **distinct metric name** — which means **a third chokepoint widening**, with
  `TestGateEligibleMetricSetsAgreeAcrossChokepoints` re-derived so it still
  forbids a one-sided widening. **This is a named task with its own test
  obligation, not a free consequence of adding a manifest row** — the same trap
  D-08-32 caught in Phase 08. Ratify the row per machine at the mid-phase gate.

### The accepted-program disclosure (D-08-26) and the code-promotion commitment (D-08-21)

- **D-09-29 (permanent test-of-record — no runtime vehicle, ever, for this
  fact — and the reason is a verified property, not a budget excuse):** verified
  in tree, the accepted-side consulted field set is a **compile-time constant**.
  The consultation loop (`check.go:594-609`) fires exactly `return.mode` and
  `parameters[0].mode` for every function with a declared signature, corpus-wide,
  recorded through the `interproceduralConsultObserved` seam (`check.go:641`) and
  already proven a closed set by the shipped `TestInterproceduralDisclosedFieldSet`
  (`check_test.go:4898`).

  A runtime record would therefore **reprint a constant on every run at real
  `qlt02`-gated cost, with zero incremental per-program information** — the
  opposite of the refusal-side disclosure, which genuinely varies (direction,
  which callee). All three vehicles Phase 08 ruled out stay ruled out for their
  recorded structural reasons; the fourth candidate this research surfaced (an
  opt-in flag mirroring `evidence --validate`'s `Trace *TraceSummary` opt-in
  pattern, `protocol.go:223-230`) is technically buildable and still rejected on
  the constant-payload argument. **Replace D-08-26's "declared debt" framing with
  "resolved — no vehicle, by design"** so a future phase does not reopen it as
  unfinished.

- **D-09-30 (the peer discloses too, the two sets must match — and this RESOLVES
  the independence tension rather than dodging it):** `corevalidate` gets its
  **own independently-written** corpus-wide closed-field-set test, and a
  cross-peer test asserts the two sets are **identical**.

  The tension — "does forcing matching field sets force the peers toward one
  derivation?" — resolves once the layer is identified correctly. Matching
  consulted **field sets** is a **contract-conformance** check ("both peers read
  only what SEM-05's body-blind `ParameterContract.Mode` / `ReturnContract.Mode`
  legally permit"). It says nothing about worklist-vs-closure, traversal order,
  or internal data structures — exactly as the shipped `structuralFieldsEqual`
  and `ClosureDigest` byte-equality comparisons already prove *outputs* without
  constraining *implementation*.

  Far from weakening independence, this is **strictly stronger than
  verdict-agreement alone**: two independent derivations can agree on
  accept/reject while one secretly reads a third field — a body fact (breaking
  SEM-05) or `Return.Paths` instead of `Return.Mode`. Verdict comparison never
  catches that; a field-set mismatch catches it immediately, permanently, at zero
  runtime cost. **This is the independence instrument Phase 08's gate
  anticipated when it sent both disclosure questions here together.**

- **D-09-31 (D-08-21's promotion commitment is FORMALLY SUPERSEDED — the codes
  stay divergent):** Phase 08 committed in writing to promoting
  `check.interprocedural_loan_liveness` to `core.interprocedural_loan_liveness`
  in Phase 09, "at the moment `corevalidate` independently re-derives the same
  fact and both peers must agree on-code." **That trigger is retired as the
  promotion criterion, because it was never achievable — and D-08-17's own
  reasoning already said so.**

  D-08-17 argued the peer "computes liveness through a reachability closure, not
  a worklist, so it has a differently-shaped ceiling and **cannot fail the same
  way**. A shared `core.*` code would assert an agreement the two mechanisms are
  structurally incapable of having." That reasoning is general to the two
  mechanisms, not scoped to the iteration bound. The corpus already demonstrates
  it: `corevalidate` refuses the relay-escort witness via
  `core.move_while_borrowed` (`corevalidate.go:1551`) while `check` refuses the
  identical program via `check.interprocedural_loan_liveness`
  (`check.go:963-966`) — and `session_peer_gate_test.go` **celebrates** this as
  "BOTH sides refuse independently, via different codes, for the same program."

  **Locked: `check.interprocedural_loan_liveness` stays `check.*`.
  `corevalidate` keeps `core.move_while_borrowed` with its own existing causes.
  Neither is renamed, aliased, or merged.**
  — **Reversibility:** one-way if reversed later — promotion would move every
  existing diagnostic ID for the code, and diagnostic codes are published
  agent-facing API consumed by `lang-repair` and `lang explain`. Not promoting
  costs nothing and stays available.

- **D-09-32 (the cause template is NOT forced onto the peer — this is why
  promotion fails):** promoting would require either forcing `corevalidate` to
  synthesize `check`'s fixed three-role template
  (`borrow_created_here` / `loan_extended_by_call` / `callee_return_contract`)
  for which a reachability closure has **no natural "the call that extended the
  loan"** — reverse-engineering the worklist's shape into the peer, the literal
  anti-pattern this phase exists to prevent — or accepting the same code with
  different causes, which (per the `Code + Span + Causes` identity fold,
  `diagnostic.go:96-112`) produces **different diagnostic IDs for the same code
  depending on which layer refused**: a strictly worse dispatch contract for
  `lang-repair` than two distinct stable codes. `Repairs: nil` stays on
  `check`'s side, independently justified by D-08-25's safety argument, and is
  never imposed on the peer.

- **D-09-33 (in-tree precedent, and the ecosystem's answer):**
  divergent-codes-per-peer is **already shipped**, not novel:
  `check.call_argument_type_mismatch` / `core.CallArgumentTypeMismatch`
  (D-07-46, Phase 07) are a different-code pair for the same defect, deliberately
  "sharing no code." `core.call_graph_cycle` is the *shared*-code precedent, and
  it is justified by something absent here — a shared, peer-agnostic **witness
  type** (`callgraph.CycleError`) available to both layers from day one. The
  ecosystem agrees: SARIF separates `ruleId` from result fingerprinting;
  Postgres SQLSTATE classes track subsystem. **Cross-mechanism sameness is
  expressed as a documented relationship, never as a merged identifier** — meet
  the agent-facing "these are the same defect" need through `lang explain`
  cross-linking or a documented code-equivalence table.

### OWN-05 — one meaning for call-site transfer

- **D-09-34 (non-expressibility is true at the source layer — prove it as
  structural absence, not as a refusal):** verified — `core.LinearOperation`
  (read in full) carries no per-call convention-override field; its only Phase-07
  addition is `CalleeID`. `corevalidate.derivePeerSignature` (`corevalidate.go:2011-2015`)
  and `originvalidate.go:811` both hardcode `Mode: "owned"` with an explicit
  "D-07-01: today's grammar has exactly one parameter form" comment. Prove the
  absence the same way `core.FunctionSignature`'s body-blindness is proven — a
  grammar/parser-level test that no production admits a call-site convention
  annotation, plus a `core`-level test that no field can carry one.

- **D-09-35 (do NOT mint a new seam here — and the distinction from D-07-47 /
  D-08-15 is the point):** the one core-level field a hostile or corrupted
  producer could use to claim a non-`"owned"` convention is
  `ParameterContract.Mode`, and it is **already validated against its closed
  three-value set at decode time**. There is no undecoded slot for a call-site
  override to hide in. A seam-backed refusal for a field the wire format does
  not have is **dead weight invented to fill a shape that does not exist** — the
  opposite of D-07-47 and D-08-15, where a real reachable-looking code path
  existed and was proven closed. **Confirm and extend the existing closed-set
  decode check as the core-level fail-closed control instead.**

  The adversarial point this preserves: "not expressible in source" is **not**
  the same claim as "not expressible in core," and `corevalidate`'s whole role
  is validating a core it did not produce. Both claims must be stated
  separately — two paragraphs in the debt register, not one.

- **D-09-36 (the shared fact and where agreement is asserted):** the fact the
  two peers must independently agree on is the per-call-site move-vs-borrow
  classification and the resulting loan/ownership state after the call. This
  **folds into TRU-04's differential** (D-09-23) rather than needing a separate
  harness — `corevalidate.Result` already exposes `PeerSignatures()` and
  `PeerSiteCoverage()`, and `corevalidate_call_argument_consume_internal_test.go`
  already exists, so the question at planning time is what is *missing*, not
  what to build from scratch.

- **D-09-37 (OWN-05 must NOT be checked off Complete at Phase 09's end):**
  OWN-05's text names **three** derivers (`check`, `corevalidate`, `interp`);
  Phase 09 delivers two. `REQUIREMENTS.md:169` maps it to Phase 09 as a single
  row, currently `Pending`. This is the same class of conflict as D-08-27, and
  it gets the same treatment **now rather than later**: either split into
  **OWN-05a (Phase 09: `check` + `corevalidate`)** and **OWN-05b (Phase 10:
  `interp`, verified by TRU-03 / NAT-06)**, or keep one row marked
  **"Phase 09 (partial) / Phase 10 (interp peer)"**. Flipping it to `Complete`
  at Phase 09's end would be exactly the requirement-vs-code overclaim the debt
  registers exist to catch. ROADMAP.md's scope note already states the Phase 10
  split; the requirement text and mapping are what must catch up.

- **D-09-38 (do NOT prepare for Phase 10 by extracting a helper):** Phase 09
  must **not** extract a shared "convention classification" helper for `interp`
  to reuse later. `check` and `corevalidate` already each read
  `signature.Parameters[0].Mode` independently, sharing nothing beyond the struct
  shape. Phase 10's `interp` peer re-derives the move/borrow decision from its
  own view of the callee signature. A helper "ready for `interp`" is precisely
  ARCHITECTURE §4's "single point of failure wearing two names."

### S-008 and QLT-07's status

- **D-09-39 (S-008 is a CATEGORY ERROR as specified — replace it, in writing):**
  the project's own spike discipline requires **named competing mechanisms** and
  **an oracle that does not reuse the mechanism under test**. "Is closing a
  validation-document debt cheap when folded into open code" has neither — it is
  a scoping/estimation question, identical in kind to the sizing `plan-phase`
  already does, not to S-006's memoized-vs-unmemoized cost-scaling comparison.
  Forcing spike shape onto it means inventing artificial arms, which
  `CONVENTIONS.md` itself argues against ("a headline that only beats the weakest
  possible opponent is not a finding").

  Second, independent reason: a new `.planning/spikes/008-*` directory would
  **reproduce D-08-43's already-open, un-owned registry gap a second time** —
  `TestQLT01RegistryCoversAllFiveSpikes` already fails today because spike 006
  has a directory and no registry row, and no phase owns that maintenance.

  **Locked: S-008 is replaced by a bounded inventory/estimation pre-flight pass**
  run during Phase 09 planning, before any `09-PLAN.md` is drafted. Record this
  as an explicit process amendment to ROADMAP.md's spike table, not a silent
  substitution.

- **D-09-40 (the threshold is PRE-REGISTERED here, before the count is taken —
  this is what keeps the gate non-decorative):** enumerate **only** the
  loan-liveness-scoped rows of `03-VALIDATION.md`'s Per-Task Verification Map,
  Mutation-Kill Register, and Generator Reachability Register — i.e. the
  03-03 / 03-04 / 03-05 rows, **explicitly excluding 03-06 / 03-07's OWN-04
  origin rows**, which QLT-07's text does not claim — and classify each as
  already-satisfied-by-Phase-07/08-artifacts vs. net-new work.

  > **QLT-07 is committed (in scope) if and only if the inventory requires
  > ≤ 1 additional plan beyond what Phase 09 already needs for
  > OWN-05/07/08/09/TRU-04, AND opens zero packages or files that OWN-07 /
  > OWN-08's work in `check` / `corevalidate` does not already touch.**

  If either bound is exceeded, QLT-07 is **re-declared a stretch item in
  writing at that moment**, in ROADMAP.md's and REQUIREMENTS.md's scope-cut
  sections, carried as disclosed debt alongside QLT-09. This is **not** a repeat
  of D-03-02's defer-under-load pattern, because the criterion and threshold are
  fixed **now**, before planning starts — not discovered mid-phase under
  pressure.

- **D-09-41 (the archived M001 document is NEVER amended in place — and it can
  never honestly close in full):** `03-VALIDATION.md`'s `nyquist_compliant: false`,
  its unticked checkboxes, and its recorded acyclic-CFG scope limitation stay
  exactly as written. Its own text defers loop-carried loan liveness, loop-exit
  edges, and per-iteration loan identity to "Phase 4-or-later" — **and the
  language still has no loops**, so that clause is structurally un-closable
  today and the document's boolean can never legitimately flip to `true`. It
  must not be made to look as though it did.

- **D-09-42 (closure vehicle — a subset closes in Phase 09's own document):**
  QLT-07's closure is recorded in a new **`09-VALIDATION.md`** section named
  **"M001 Phase 3 Debt Closure (loan-liveness subset)"**, citing the specific
  03-03 / 03-04 / 03-05 rows satisfied, stating explicitly that OWN-04's rows and
  the loop-carried-liveness exclusion remain **outside** this closure, and
  carrying its own `nyquist_compliant` flag scoped to Phase 09. A **single
  non-mutating pointer line** is appended to `03-VALIDATION.md` — touching
  neither its frontmatter nor its checkboxes — noting that its loan-liveness
  subset is superseded by Phase 09's document as of that date. This is the
  supplemental-filing pattern: reference the original approval record, never
  rewrite it.

### Cross-cutting

- **D-09-43 (scope-cut trigger, declared now in writing per Key Lesson 4):** the
  milestone declares a **2× trigger** for Phases 08 and 09. Adopted for Phase 09:
  **if the peer-derivation + D-07-33-closure work exceeds ~2× its initial plan
  estimate, the cut order is (1) QLT-07 per D-09-40's threshold, then (2) the
  `OpForeignCall` slice of D-07-33 per D-09-21's named fallback. Never the
  criterion-1 differential corpus (D-09-23), never the seeded fault and its
  companion assertion (D-09-25), never the `computeLoanLastUses` deletion
  (D-09-08), and never the two `peerDivergenceExpected` retirements (D-09-03).**
  Any cut is a deferral with a named landing phase, never a silent drop.

- **D-09-44 (`PHASE-09-DEBT.md` written at PLANNING time, not phase end):** in
  the mechanically-checked format `TestDebtRegistersAreWellFormed` enforces
  (`items:` count matching the `## Items` table, one `### <ID>` detail section
  per row, severity from {blocker, warning, info}). It must carry, at minimum:
  D-09-08 (the D-08-41 reversal), D-09-13 (diagnostic ordering risk), D-09-21
  (the D-07-33 fallback and its trigger), D-09-31 (the D-08-21 supersession),
  D-09-37 (OWN-05's partial-completion disposition), D-09-40 (QLT-07's
  pre-registered threshold), and D-09-43 (the scope-cut trigger). It should also
  carry forward the resolutions of **D-08-26** (→ D-09-29/D-09-30, resolved),
  **D-08-27** (→ D-09-14, resolved), and **D-08-40** (→ D-09-03, resolved).

- **D-09-45 (three stale planning-document references, corrected here rather
  than left to mislead):** (a) `08-CONTEXT.md`'s D-08-09 cites
  `computeLoanLastUses` at `check.go:2979`; it is at **`check.go:3743`**.
  (b) `08-CONTEXT.md`'s D-08-03 and 08-01-PLAN's original doc comment assert
  `callgraph.Order` is callee-before-caller; it is **caller-before-callee** and
  `buildInterproceduralSummaries` walks it backward (this was already recorded as
  D-08-42 and is repeated here because Phase 09 planning will read those
  documents). (c) The `.planning/spikes` registry gap (D-08-43, spike 006 has no
  row) is **still open and still un-owned**; if Phase 09 touches the spike table
  at all under D-09-39, adding spike 006's row closes it cheaply — recommended,
  not required.

### Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all
eight gray areas, under the standing mandate to fan out across stakeholder-role
lenses, run an adversarial pass, draw on cross-ecosystem prior art, and
synthesize one-shot recommendations. Every decision above is therefore Claude's
synthesis under that instruction, grounded in the eight advisor research returns
and **independently verified against the shipped tree**, with three researcher
claims corrected (D-09-05's already-enforced `callgraph` import ban, D-09-24's
seam-precedent location and direction, D-09-28's already-widened chokepoints plus
the unpriced third widening).

Planner discretion remains over:

- Plan decomposition and wave ordering, subject to three hard constraints:
  **build-then-delete** (D-09-10), the differential landing in the **same plan**
  as the peer with the test written red-first (D-09-27), and the **third
  chokepoint widening landing before** any peer manifest row claims to be a hard
  gate (D-09-28).
- Whether the peer's new derivation lives in `corevalidate.go` or a sibling file
  in the same package (D-09-04).
- Exact Go identifier names for the peer's liveness bits, the seams, and the new
  metric.
- Whether OWN-05 is split into OWN-05a/OWN-05b or kept as one partial row
  (D-09-37) — either is honest; picking one is required.
- The order in which the four `PublishProblemsFor` classes are closed within
  D-07-33's work, subject to D-09-21's fallback trigger.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase charter and requirements

- `.planning/ROADMAP.md` — Phase 09 section (goal, five success criteria,
  riskiest assumption, the gate that catches it, scope note on `interp`/QLT-07);
  the Pre-Phase Spikes table (S-008's row, amended per D-09-39); the
  gate-level sequencing block (`07 → 08 → 09 → …`).
- `.planning/REQUIREMENTS.md` — OWN-05 (:36), OWN-07 (:42), OWN-08 (:45),
  OWN-09 (:48), TRU-04 (:71), QLT-07 (:87); the requirement→phase mapping table
  (:165-185, incl. OWN-05 at :169 and OWN-09 at :173); the Scope-Cut Order and
  never-cut list (:205-225).
- `.planning/PROJECT.md` — M002 charter, active requirements, out-of-scope list.
- `.planning/STATE.md` — current position, durable context pointers.

### The pitfall this phase exists to survive

- `.planning/research/PITFALLS.md` Pitfall 2 — "`check` and `corevalidate`
  silently drift on cross-function facts"; its "land both peers in the same plan"
  prescription; the D-02-03 "landed in the wrong half" precedent; the LLVM
  IR-verifier analogy. **Also Pitfall 1's line: "Do not resolve D-03-02 by
  pattern-matching the intraprocedural fix… the contract must be RE-DERIVED for
  the call-boundary case, not copied."**
- `.planning/RETROSPECTIVE.md` — Key Lesson 2 (coexisting laws, and deleting the
  one with coverage); Key Lesson 3 ("fixing one of N independent peers is not
  fixing the item"); Key Lesson 4 (declare deferred scope in writing at the
  moment it is decided).
- `.planning/research/ARCHITECTURE.md` §4 — the outright rejection of a shared
  interprocedural-liveness library ("a shared implementation isn't a peer at all,
  it's a single point of failure wearing two names"); Stage 3 and the
  `corevalidate` half of Stage 4.
- `.planning/research/SUMMARY.md` — the reconciled build order and the
  mid-phase-gate requirement.

### Inherited debt this phase must close or carry

- `.planning/phases/08-interprocedural-loan-liveness-in-check/PHASE-08-DEBT.md`
  — **D-08-26** (accepted-program disclosure → resolved by D-09-29/D-09-30),
  **D-08-27** (OWN-09 conflict → resolved by D-09-14), **D-08-40** (the two
  peer-divergence fixtures → resolved by D-09-03), **D-08-41** (Pattern B scope
  limit → **reversed** by D-09-08), **D-08-37** (cross-run cache non-goal →
  binds the peer too, D-09-06), **D-08-42** and **D-08-43** (stale doc refs and
  the spike-registry gap → D-09-45).
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md`
  — **D-03-02**'s closure gate quoted verbatim; **D-07-33** (the narrowed peer,
  its four refusal classes, and why it was narrowed) → D-09-15 through D-09-21;
  **D-07-46** (the divergent-code precedent) → D-09-33.
- `.planning/phases/08-interprocedural-loan-liveness-in-check/08-CONTEXT.md` —
  Phase 08's locked decisions, especially D-08-06 (no shared library), D-08-17
  (why the namespace differed), D-08-20/21 (the code and its promotion
  commitment), D-08-23 (the three-role cause template), D-08-25 (`Repairs: nil`),
  D-08-29/31/32 (the cost-gate instrument). **Note D-09-45: two line references
  in this document are stale.**
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/07-CONTEXT.md` —
  D-07-02 (declared convention read implicitly), D-07-15, D-07-22, D-07-42,
  D-07-47.

### Validation artifacts (QLT-07)

- `.planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md`
  — `nyquist_compliant: false`, the acyclic-CFG scope limitation, the Per-Task
  Verification Map / Mutation-Kill Register / Generator Reachability Register
  whose 03-03/03-04/03-05 rows D-09-40 scopes. **Read-only — never amended in
  place (D-09-41).**
- `.planning/phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md`
  and `07-VALIDATION.md` — the in-milestone shape a `09-VALIDATION.md` follows.
- `.planning/spikes/CONVENTIONS.md` and `.planning/spikes/MANIFEST.md` — the
  spike discipline D-09-39 argues S-008 fails, and the registry gap in D-09-45.
- `.planning/spikes/006-interprocedural-liveness-cost-scaling/README.md` —
  S-006's findings, which underwrite Phase 08's instrument and D-08-37's
  cache non-goal.

### Language and standing verdicts

- `.planning/LANGUAGE-MATURITY.md` — what the language can actually express
  (no arithmetic, no iteration, no collections, arity 1, ~58 programs averaging
  ~28 lines). **Load-bearing for D-09-11, D-09-18, D-09-22, and D-09-41.**
- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (deps,
  anti-features, the six dispatch sites, why `-flto` is load-bearing).

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`corevalidate`'s peer substrate** (`internal/compiler/corevalidate/corevalidate.go`):
  `v.peerPostorder` (:138, appended :472 — the peer's **own** DFS byproduct,
  callee-before-caller), `v.peerSignatures` (:126, refined in the postorder loop
  at :505-547), `derivePeerSignature` (:1990), `peerReturnDerivesFromBorrow`
  (:2099 — forward propagation, presence-only today), `peerParameterEscapesOwned`
  (:2061), `peerCallable` (:2349), `recordSummaryPeer` (:1942),
  `peerComputeClosureDigest` (:696). This is the substrate D-09-01 extends.
- **`buildLoanChainIndex` / `loanChainIndex`** (:1126-1200, `carriedLoans` :1160,
  `foldChain` :1188) — the site of D-09-03's defect.
- **`recomputeLoanEndpoints`** (:1270) and **`loanEndpointsMatch`** (:1405) — the
  peer's existing endpoint machinery, the natural home for TRU-04's seeded
  endpoint-level fault.
- **Seam precedent, verified:** `verifyCallableRefusalSeam` (`check.go:1183`,
  read at :1250) with `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses`
  (`check_test.go:3322`) — the exact companion-assertion shape D-09-25 requires.
  Also `callReturnTypeDerivationSeam` (`check.go:493`) and `loanLivenessBoundSeam`.
- **Disclosure seam:** `interproceduralConsultObserved` (`check.go:641`, fired at
  :599-606) and `TestInterproceduralDisclosedFieldSet` (`check_test.go:4898`) —
  the pattern D-09-30 mirrors on the peer side.
- **Corpus generator:** Phase 08's `generateCallGraphCorpus` (five shapes) and the
  least-squares growth-exponent fit — reused by D-09-23 as a program source only.
- **Peer-agreement comparison precedent:** `structuralFieldsEqual` and
  `ClosureDigest` byte-equality in `corevalidate_summary_peer_test.go` — proves
  outputs agree without constraining implementation, which is D-09-30's argument
  in shipped form.

### Established Patterns

- **Import independence is enforced, not asked for.** Four shipped guards forbid
  `corevalidate` from importing `check`, `ast`, `originvalidate`, and
  **`callgraph`** (`corevalidate_cycle_peer_test.go:146`,
  `corevalidate_endpoint_internal_test.go:107`, `corevalidate_exclusive_test.go:216`,
  `corevalidate_test.go:148`). Production `corevalidate.go` imports only `core`.
  Test files *may* import `check` as an ordinary dependency to drive comparisons
  (`corevalidate_test.go:15` documents exactly this distinction).
- **Exact-set divergence gating.** `peerDivergenceExpected` +
  `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`
  (`internal/compiler/session/session_peer_gate_test.go`) fails on **both** a new
  undeclared divergence and a stale entry that no longer diverges.
- **Unexported seam + deferred restore + both-directions mutation-kill** (QLT-08).
- **Debt registers are mechanically validated** (`TestDebtRegistersAreWellFormed`).
- **Gate-eligible metrics are a closed set declared at two independent
  chokepoints** (`measure.Demote`, `session.QLT02GateEligibleMetrics()`), proven
  to agree by `TestGateEligibleMetricSetsAgreeAcrossChokepoints` — deliberately
  duplicated so a one-sided widening fails.

### Integration Points

- `check.go:3743` `computeLoanLastUses` and its callers `analyzeStraightLine`
  (:3287) / `analyzeArmBody` (:2074) — deleted and restructured by D-09-08/09.
- `check.go:2176`, `:3392` — the `activeLoans`/`expiringLoans` state machines
  that consume the deleted index and raise `ownership.move_while_borrowed`.
- `corevalidate.go:1133` `buildLoanChainIndex` — the `OpCall` propagation fix.
- `corevalidate.go:505-547` — the postorder loop the new bits fold into.
- `internal/compiler/session/session_peer_gate_test.go` — two entries retired.
- `internal/compiler/session/qlt02_budget_manifest.json` (rows today:
  `recomputed_work`, `recomputed_work_growth_exponent`, `elapsed_ns`,
  `output_bytes`), `session_phase6_budget.go:55-57`,
  `measure/statistics.go:131-135`, `risk_lanes.json:232`
  (`lane:interprocedural-cost-scaling`) — D-09-28's third widening and the peer's
  own row.

</code_context>

<specifics>
## Specific Ideas

- **Write TRU-04's "recursion = cycle-peer differential" disposition into the
  requirement's own definition-of-done text**, not just a test comment
  (D-09-22). The failure mode being prevented is a future reader assuming this
  compiler admits recursive programs.
- **Write D-09-25's discrimination argument into the seeded-fault test's own doc
  comment** — that a seeded fault without the unseamed-peer companion assertion
  proves nothing, per Knight & Leveson's correlated-failure result. The existing
  `TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses` is the
  naming and shape precedent.
- **State D-09-07's finding prominently wherever OWN-09 is planned**: there is
  one algorithm with two callers, not two laws. A plan that reads OWN-09
  literally will delete the wrong thing.
- **`corevalidate`'s new derivation doc comment must distinguish reused
  *substrate* from reused *derivation*** (D-09-04) — the conflation a reviewer
  will probe first.

</specifics>

<deferred>
## Deferred Ideas

- **`interp`'s third independent derivation of ownership transfer** (OWN-05's
  remaining peer) — Phase 10, verified there by TRU-03 and in Phase 11 by
  NAT-06. Explicitly **not** prepared for by extracting a shared helper
  (D-09-38).
- **Persistent cross-run summary caching** — Phase 11 / QLT-06, per D-08-37's
  non-goal, which binds the peer as well as `check` (D-09-06).
- **A runtime accepted-program disclosure artifact** — closed as "no vehicle, by
  design" (D-09-29), not deferred. Reopen only if a future law's accepted-side
  consultation becomes genuinely per-program-variable, which it is not today.
- **Promotion of `check.interprocedural_loan_liveness` to `core.*`** — closed as
  superseded (D-09-31), not deferred. Reopen only if the two mechanisms are ever
  proven to refuse for the identical underlying fact with an identical natural
  cause shape.
- **`.planning/spikes` registry maintenance** (D-08-43, spike 006's missing row)
  — still un-owned by any phase; cheap to close if Phase 09 touches the spike
  table under D-09-39, but not a charter item (D-09-45c).
- **Re-opening the two-bit summary model** — the moment arity widens past 1
  (D-07-07) or `Result` match arms can independently borrow-vs-move a parameter
  (Phase 12), both D-08-10 and D-09-18 re-open.

</deferred>

---

*Phase: 09-Peer Re-Derivation and D-03-02 Closure*
*Context gathered: 2026-09-10*

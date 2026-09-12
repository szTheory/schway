# Phase 11: Multi-Function Native Emission and Interprocedural Equivalence - Context

**Gathered:** 2026-09-11
**Status:** Ready for planning

<domain>
## Phase Boundary

A multi-function Lang program lowers to readable C17, compiles, links, runs, and
agrees with the Phase 10 interpreter oracle on the five-axis comparator — and the
optimizer tier is proven non-inert rather than assumed.

This phase makes multi-function programs **executable**. Phases 07-10 made them
*admissible*: `check`, `corevalidate`, `originvalidate`, `pathoracle` and `interp`
all handle `OpCall` today, but `lang run` refuses a multi-function program on
**both** engines. That refusal is 32 non-test `len(Functions) != 1` guards across
6 files in 3 packages (`session` 26, `cgen` 4, `reduce` 2), not the two `cgen`
entry points the roadmap's criterion 1 names.

**Not in scope:** deleting the six single-function emitters (Phase 12, see D-11-02);
closing the `peerDeriveOriginFacts` `OpCall` gap beyond what Q-01 requires;
`Result` payloads (Phase 12); arithmetic, iteration, strings, arrays (on no roadmap).

</domain>

<decisions>
## Implementation Decisions

Seven decision points were fanned out to parallel researchers with an explicit
adversarial pass each. Four returned findings that change scope; two conflicted
and are adjudicated below; one requirement is weakened by evidence and flagged.

### Emission structure

- **D-11-01:** `cgen` gains a new **additive** whole-program path — `emitProgram`
  in a new file `internal/compiler/cgen/cgen_program.go` — reached only when
  `len(program.Functions) != 1` at `cgen.go:34` and `cgen.go:64`. The six existing
  whole-TU emitters (`emitLinear` 234, `emitLinearBorrowedByPointer` 456,
  `emitLinearBorrowedByPointerPlain` 649, `emitLinearForeign` 793, `emitBranch`
  1597, `emitMatch` 152) are left **byte-for-byte untouched**. Their *predicates*
  (`selectsByPointerLowering` 400, `selectsByPointerLoweringSharedOnly` 589,
  `functionHasDefect` 1559) and *support writers* (`emitEventSupport` 1507,
  `emitStreamingEventSupport` 1468, `emitDefectSupport` 1579, `resourceLedger`
  1057) are **called, never forked**.
  Rationale: each of the six writes a complete TU including its own
  `int main(int argc, char **argv)` — there is no per-function seam to loop over.
  Extracting one from all six puts `testdata/phase5/restrict_borrow.golden.c` and
  the four pinned digests at `core_test.go:156-159` at risk in the same diff that
  introduces calls; when a digest moves you cannot tell which change did it. The
  additive path freezes them **by construction** rather than by test.
  — **Reversibility:** costly — two emitter families coexist until Phase 12; the
  deletion is gated on the N=1 convergence differential (Q-05) and must land inside
  Phase 12's per-dispatch-site `Result` plans, never as a second sweep.

- **D-11-02:** Deletion of the six single-function paths is **declared debt now,
  scheduled for Phase 12**, with its gate written down: a green N=1 convergence
  differential. This is the build-then-delete pattern Phase 09 already ran
  deliberately (D-09-10: peer lands fully, zero-divergence differential goes green,
  only then the deletion). Record it in `PHASE-11-DEBT.md` at phase open, not at
  phase close.

- **D-11-03:** Body emitters emit **statements and events only**; the TU assembler
  owns includes, typedefs, every support block, the resource ledger, prototypes,
  and `main`. This is the split every real C backend draws (Nim `genProc` vs
  `genMainProc`; Vala `CCodeFunction` vs `CCodeFile`; Cython's `UtilityCode`
  require-set).

- **D-11-04:** **One** `emitCall` helper is the only writer of a Lang-to-Lang call
  anywhere in `cgen`, called from exactly two places (straight-line bodies and
  block bodies). Roadmap-mandated; do not let it become five copies across the
  switch sites. It never mints an attribute — `restrict` is a property of a
  *definition* and its prototype, never of a call site.

- **D-11-05:** Entry point resolves via a new `callgraph.EntryFunction(program)`
  returning the unique in-degree-zero root over the adjacency `buildAdjacency`
  (`callgraph.go:215`) already computes; zero-or-many roots is a **named
  fail-closed refusal**, never a guess. `session.go:510`, `:773`, `:886` switch
  from `Functions[0].Name` to the same resolver so the oracle and the binary can
  never disagree about which function *is* the program. For every existing
  single-function fixture the unique root **is** `Functions[0]`, so Phase 1
  evidence bytes do not move.
  — **Reversibility:** one-way — `session`'s three run sites and the oracle
  contract change together; reverting means re-establishing which function the
  interpreter was run against, and every multi-function fixture's evidence was
  produced under the new resolver.

- **D-11-06:** Do **not** add an `Exported`/`EntryFunctionID` field to
  `core.Program` this phase. Every existing fixture writes `export { fn main }`;
  any populated field (even `omitempty`) moves frozen core bytes and forces a
  `lang.core/2` bump against D-05-39. Instead assert `EntryFunction` agrees with
  the single `ast.Export` across the whole corpus — the same independent
  re-derivation pattern already in use.

- **D-11-07:** `singleForeignFunction` (1868) and `singleManifestFunction` (1908)
  stay as they are and stay scoped to `EmitForeignManifest`/`EmitForeignHeader`/
  `EmitForeignConformance`. With N functions they generalize by **refusing** a
  program with two foreign contracts, not by picking one. `lang.foreign/0` has no
  multi-symbol schema and must not be quietly widened. (These are additional
  single-function assumptions the `len(Functions) != 1` grep does **not** catch.)

- **D-11-08:** Name allocation is two-tier: **one** `cNames` allocates every type
  and function name first in `callgraph.Order` (deterministic); each function then
  gets a **fresh** `cNames` for locals, seeded with the reserved list *plus* every
  allocated global name. C block scope makes locals independent; the seed makes
  shadowing impossible; the second function's places stay `lang_value_x` rather
  than `lang_value_x_2`. `cgen_names_test.go` grows the multi-function case —
  this is what the honest-reservation property at `cgen.go:96-150` demands.

### Alias attributes — the phase's central scope change

- **D-11-09 (SCOPE REDUCTION, evidence-driven):** Phase 11 emits **zero
  call-boundary alias attributes**. The mid-phase gate's zero-attribute `-O0`
  milestone is the **terminal** state for this phase, not a waypoint.
  Rationale, established by proof rather than preference: **two pointers to one
  object cannot exist across a Lang call boundary in M002.** Lang functions take
  exactly one parameter (`core.Function.Parameter`), have no globals, no
  callbacks, no address-escaping foreign contracts (all `foreign C` params are
  `Byte` by value), and `check` already refuses passing one place to two calls
  (`testdata/phase07/call_argument_used_twice.lang`). A call-boundary `restrict`
  in M002 is therefore a promise about a hazard that **cannot exist** — it buys
  nothing and stakes the equivalence claim on LLVM's scoped-noalias machinery,
  whose known weakness kept Rust's `noalias` on `&mut` disabled across #31681 and
  #54878, was re-enabled on LLVM 12 (#82834), and immediately regressed (#84958).
  — **Reversibility:** reversible — this withholds an emission rule rather than
  establishing one. Adding call-boundary attributes later is additive work whose
  gate (a non-inert LTO control) this phase builds anyway.

- **D-11-10:** NAT-05 ("every aliasing or capture promise emitted at a call
  boundary names the checked fact it derives from") is consequently satisfied by
  an **explicit, visible empty set**: the emitted artifact carries a generated
  comment recording that the call-boundary attribute set is empty **and why**,
  citing D-11-09. This is a requirement **weakened by evidence** and must be
  written up as such in `11-VERIFICATION.md` — it reads as done and is not the
  same claim the roadmap's prose implies.

- **D-11-11:** The `EmittedAttribute` *discharge-pair* design — callee-side
  `justified_by` plus a caller-side `discharged_by` list (one entry per `OpCall`
  naming the operation, argument place and caller loan), refused on **equality,
  not containment**, because a *missing* discharge is the forgery shape — is
  **designed, recorded, and not built** in Phase 11. It is the correct design the
  moment D-11-09 is revisited. Its cheapest falsifier (Q-06) is still worth
  running this phase as evidence, because it costs about an hour.

- **D-11-12:** If call-boundary attributes are ever emitted, the channel is
  **sidecar normative + inline C comment generated from the identical struct
  value** — one construction, two renderers, plus a text↔manifest cross-scan lane.
  The repo already made this ruling at `cgen.go:1980-1983` (D-04-12): "a
  hand-written comment beside a generated JSON is a second source of truth that
  will drift." The inline comment is **not** a second knower — it is one
  derivation printed twice; the knower is `corevalidate.ValidateEmittedAttributes`
  with an `OpCall` arm re-deriving the discharge set from the program alone.
  Schema verdict: mint **`lang.attributes/0`**, do **not** bump `lang.foreign/0` —
  that manifest is already structurally single-function via `singleManifestFunction`
  and folding multi-function attribute entries into a manifest named "foreign"
  cements a misnomer into a schema name.

- **D-11-13 (PRE-EXISTING GAP, disclose):** `evidence.go:233` binds `ForeignDigest`
  only under `hasForeignContract`. A pure-Lang `restrict` claim is therefore **not
  digest-bound into evidence today** — D-05-01's sidecar is emitted on demand in a
  verify lane and never becomes content-bound evidence. Live now, not introduced by
  Phase 11. Record in `PHASE-11-DEBT.md`; closing it is `lang.attributes/0`'s job
  whenever D-11-11 is built.

### The mid-phase gate

- **D-11-14:** The gate is a **conjunction with a second independent knower**, not
  a fixture-selection assertion:
  1. `cgen.ScanForBannedAttributes` (`cgen.go:2084`) returns empty over every
     emitted artifact — reads the emitter's *output bytes*;
  2. `check`, on the **same corpus**, independently reports **N ≥ 1 functions that
     would have carried `restrict`** — reads the *program*, never `cgen`;
  3. structural floors: ≥2 functions, ≥1 call edge;
  4. interpreter ≡ `-O0`.
  Read aloud: *the corpus had real aliasing to promise about, we emitted zero
  promises, and it still compiled, linked, ran, and agreed.* Hard-wiring
  `ByPointer = false` alone cannot distinguish "we suppressed attributes" from
  "this program had nothing to say" — that is fixture selection wearing a
  different hat, and it is the M001 three-failure shape (*a green test whose
  reachable input space omitted the hard case*).

- **D-11-15:** The manifest's empty `emitted_attributes` is **explicitly not** the
  second knower — that is `cgen`'s own derivation read twice. Assert it as a
  consistency check only.

- **D-11-16:** Suppression is **token-local**, not a second emitter: the
  `Fprintf` at `cgen.go:498` keeps its single call site and gains one interpolated
  qualifier from one helper. Guarded by a **diff-locality test** — suppressed
  output must differ from justified output *only* in bytes belonging to
  `BannedOptimizerAttributes`. Any structural divergence fails red the moment it
  appears, which is what answers the "second convention to keep in sync" objection.
  `NoreturnExemption` (`cgen.go:2077`) is not an alias promise and stays emitted
  under both profiles — the repo's analogue of LLVM `OptBisect`'s never-skipped set.

- **D-11-17:** The attribute-suppression control is **permanent, declared so up
  front**, not a temporary flag. Its standing value is the inertness differential:
  attributes-off vs attributes-on vs interpreter at `-O3`/`-flto` tests `restrict`'s
  semantic-inertness claim directly and forever, and is the bisection tool for
  "our promise, or Clang?". Undeclared permanence is the flag-debt failure mode,
  not permanence itself. Prior art: LLVM shipped `-opt-bisect-limit`/`OptPassGate`
  rather than curating inputs, for exactly this reason.

- **D-11-18:** Anti-vacuity is mutation-killed: re-running the gate corpus under
  the justified profile must make the lane go **red**. A gate that passes under
  both profiles is measuring nothing. Mirrors the existing
  `TestAttributeInjectionMakesControlFail` pattern.

- **D-11-19:** The gate's verdict is a written adjudication in the phase directory
  recording corpus size, function count, call-edge count, N from `check`, artifacts
  scanned, and the `-O0` agreement result — following Phase 08's four-item gate
  adjudication and Phase 10's `DID NOT FIRE` precedent.

### NAT-07 — the composition-only negative control

- **D-11-20 (ROADMAP AMENDMENT REQUIRED):** Criterion 3's divergence signature
  `interpreter == -O0 != -O3` is inherited verbatim from M001's single-TU control
  and is **factually wrong for the interprocedural sequel**. Measured on this host
  (Apple clang 21.0.0, arm64-apple-darwin25.6.0), the composition-only divergence
  is `interpreter == -O0 == -O3 != (-O3 -flto)`. Demanding a non-LTO `-O3`
  divergence forces the control back into a single TU, destroying the very thing
  criterion 3 exists to prove.
  **Amend to:** *reproduces a divergence in which the interpreter and `-O0` agree
  and at least one optimized tier disagrees, with the `-flto` tier required to be
  the disagreeing one.*

- **D-11-21 (ROADMAP AMENDMENT REQUIRED):** "**fails red before its fix**"
  presupposes a shipped defect with a shipped fix. Per D-11-09 no such defect is
  constructible in M002.
  **Amend to:** *the control is demonstrated red under its injected mutation and
  green unmutated, with a full mutation-kill matrix over each injection site
  independently; an all-green matrix is a lane **failure**, not a pass.*

- **D-11-22 (ROADMAP ADDITION REQUIRED):** Add a clause criterion 3 currently
  lacks: *the control's red cell is re-measured against a recorded
  `clang --version`; a toolchain change that extinguishes the divergence is an
  escalation, not a pass.* Exploitation is **non-monotonic in inlining
  aggressiveness** — measured: `-O1` → correct, `-O2`/`-O3` single-TU → diverges,
  full-LTO 2-TU → inlined everything and constant-folded back to correct. Three
  behaviours for morally identical programs, one afternoon, one machine.

- **D-11-23:** The control is a **coordinated two-site mutation across two
  translation units**, extending the proven `AliasFactMutationRunner` shape
  (`session_phase5_alias.go:46,67,351`): the false `restrict` is injected on the
  callee TU's by-pointer parameter, the aliasing write in a third TU's function,
  and the caller binds probe == primary. Neither injection alone diverges; neither
  TU alone diverges; only `-flto` can inline across them. Yields a 4×3 matrix with
  exactly one red cell.

- **D-11-24 (ADJUDICATION — TU topology):** The **production corpus stays
  single-TU** (D-11-01's assembler). Multi-TU is required **only** for NAT-07's
  control, which ships as a **hand-written-C matrix test** ported into
  `internal/compiler/native/native_lto_test.go` (which already asserts `-flto`
  reaches both compile and link lines) — pinned to the recorded Clang version,
  **before** the emitter learns to write more than one source file.
  `native.Runner`'s single `program.c` (`native.go:153-158`) is **not** widened
  this phase. This directly satisfies M001's standing rule: *drive the shipped
  toolchain on hand-written programs, not only the gate's own corpus.*

- **D-11-25 (DECLARE IN WRITING):** Consequence of D-11-24 — because all Lang
  functions land in one TU, **`-flto` is inert by construction for Lang-to-Lang
  code** in criterion 2's corpus. The existing LTO lane's non-inertness is borrowed
  entirely from the *foreign* TU boundary (`inline_across_foreign.lang`). This must
  be stated in `11-VERIFICATION.md`, not left as an implied claim. NAT-07's control
  is what proves the *tier* can be exploited; it does not make the corpus's LTO
  tier meaningful.

- **D-11-26:** The control's honest scope, to be written into the phase record
  verbatim: it survives as **tier evidence plus a mutation-killed validator test**,
  and it does **not** survive as evidence that a shipped Lang `restrict` is sound.
  The `interpreterInput` stipulation inherited from M001 (the injected
  demonstration has no Lang-level meaning) is disclosed as a named residual
  alongside `EscapeCoordinatedSourceToCoreFalseClaim` — not laundered.

- **D-11-27:** NAT-07's design does **not** depend on reviewing the unreviewed
  D-09-51 negative-control flip (STATE.md:249), *because* D-11-09 deletes the
  emission rule that would have depended on it. Reject D-11-09 and that review
  becomes a hard precondition.

### The HDD reducer (QLT-05)

- **D-11-28:** **Never inline.** Full multi-function HDD with callee inlining is
  rejected on a point that reframes the requirement: inlining **destroys the
  defect class Phase 11 is about**. The call boundary *is* where the promise
  lives; inline the callee and the emitted attribute is deleted, guaranteeing
  slippage into a different bug. It also forces ID renumbering, which invalidates
  `Signature.OperationID` and makes the `CausalRole` fallback load-bearing — the
  exact relaxation that would make QLT-05 vacuous.

- **D-11-29:** `reduce` gains **exactly two** new whole-program moves, run
  **first** (coarse level before intra-function narrowing, per HDD's level-by-level
  descent):
  1. `drop-call-site` — rewrite the first eligible `OpCall` to `core.OpCopy` with
     the same `SourceID`/`TargetID`, clear `CalleeID`. No renumbering, no place
     churn, no block surgery.
  2. `drop-orphan-function` — delete the first non-entry function with zero
     in-edges over all `OpCall.CalleeID`.
  Order matters: a callee only becomes orphanable after its last call site is
  dropped. This is J-Reduce's dependency-graph closure expressed as two single-site
  moves. The existing five moves become per-function loops replacing the
  `p.Functions[0]` indexing at `reduce.go:184, 214, 266, 484, 584`.
  Verified by probe programs, not assumed: an uncalled unexported function checks
  clean (so function removal is never *required* for validity — call removal is);
  a devirtualized call checks clean; forward references check clean (so
  `ProjectSource` needs no topological sort and `reduce` never imports
  `callgraph`, keeping that package's consumer set pinned to `check`).

- **D-11-30:** `Reduce` takes a `Seed{Program, EntryFunctionID}` rather than a bare
  `core.Program`: `core.Function` carries no export/access field (`core.go:29-52`),
  only `core.FunctionSignature.Callable` does, and that lives in the interface
  artifact. The entry function is **caller-supplied**, never minted as a new core
  schema field.

- **D-11-31:** The flat `MaxReductionAttempts = 64` becomes a **derived** bound
  (`loanLivenessFixpoint` precedent): `AttemptsPerFunction*len(Functions) +
  callSiteCount`, with `AttemptsPerFunction = 64` preserving the constant verbatim
  so a one-function zero-call seed gets **exactly** 64 attempts and every existing
  single-function reduction's output is byte-identical. `Minimality` stays the
  closed two-value field it is today — `budget_exhausted` is not a refusal and no
  third state is introduced.

- **D-11-32:** QLT-05's re-verification is **strict field equality**, not
  `Interesting`-equality: `Axis`, `EnginePair`, `OperationID` (no `CausalRole`
  fallback) and foreign-call-sequence drift, re-run from a cold start. The
  asymmetry is the whole point — `Interesting` permits the relaxation *during the
  search* to tolerate shifted operations; inheriting it at re-verification would
  satisfy QLT-05 by construction and prove nothing. It is **safe** to forbid here
  because every operation ID is function-ID-prefixed (`…:fn:main:op:0`) and
  neither new move renumbers anything, so a genuine ID shift is impossible and
  observing one is a real reducer bug. This converts HDD's classic slippage trap
  into a falsifiable assertion.

- **D-11-33 (SILENT-SLIPPAGE HOLE, fix required):** `foreignCallSequenceFor`
  returns `nil` for a multi-function program (`session_phase5_mismatch.go:72-75`).
  Left as-is, the reducer could drop a call to a callee containing a foreign
  acquisition and drift-checking would compare `nil` to `nil`, see nothing, and
  **reduce to a different bug while passing re-verification clean**. Widen the walk
  across all functions in `callgraph` reverse-postorder, **and** add a *dynamic*
  second knower deriving the same sequence from the `-O0` run's event stream
  (which already carries `FunctionID` per event) — satisfying "never one derivation
  read twice" for the one fact that guards against slippage.

- **D-11-34:** Anti-vacuity: a reduction that drops **zero** calls and **zero**
  functions re-verifies perfectly and the lane goes green. The existing
  `control:reduce.no_progress` does **not** cover this — it was written against a
  single-function reducer. The multi-function gate fixture must be engineered so at
  least one function is **provably removable**, asserting `AppliedMoves` contains a
  `drop-orphan-function`. One assertion; it is the difference between a gate and a
  decoration.

- **D-11-35:** `RefusedShapes()` ships as an enumerated, test-asserted register of
  shapes the reducer provably declines (`Match`-bodied functions, type-changing
  calls, ADT-typed collapsed matches, recursive graphs), each naming the shape, its
  refusal ID, and *why*. Refusing more is acceptable here **because** the refusals
  are named. New refusal IDs follow the `check.loan_liveness_bound_exceeded`
  convention: `reduce.seed_shape_unsupported`, `reduce.projection_unsupported`,
  `reduce.reverification_signature_drift`.

- **D-11-36:** Flaky-predicate tolerance is **explicitly not built**. A `-O3`
  miscompile reproducing only sometimes would make the reducer non-deterministic
  and violate byte-for-byte determinism. Record it in `RefusedShapes()` as an
  unhandled case. `false_restrict_hoist.lang`'s divergence is a *seeded,
  deterministic* mutation, not a real nondeterministic miscompile. If Phase 11's
  engineered control turns out flaky, that is a criterion-3 problem surfacing in
  criterion 4 — escalate, do not absorb.

- **D-11-37:** No `lang.mismatch/1` bump. QLT-05 is a **gate** claim, not a
  document field; report it as a new lane control `control:reduce.reverified` on
  `LaneMismatchReduce`. `MismatchDocument`'s field set stays pinned at
  `lang.mismatch/0`.

### Caching (QLT-06)

- **D-11-38:** **Cache nothing new in Phase 11.** The decisive fact is not
  eviction cost — it is that `ArtifactSpec.FixtureSource` already hashes the
  **entire `.lang` source file**. Lang has one module, one file, no separate
  compilation units, so a whole-program hash **strictly dominates** any
  call-graph-closure key. A closure key could only ever admit *more* hits, each on
  strictly less evidence: it is a soundness-**loosening** change that buys nothing
  and costs complexity. Do **not** wire `ClosureDigest` into `cache.Input`.
  QLT-06's named failure mode — "per-unit hashes" — is not present in the repo and
  cannot be, because there are no units. Verified: `ClosureDigest` appears in
  `core` and `originvalidate` only; **no file under `internal/compiler/cache/`
  mentions it**.

- **D-11-39 (SPLIT THE ROW, OWN-05 precedent):**
  - **QLT-06a — Complete (Phase 07).** The callee-changes-invalidates-caller
    regression exists and is mutation-killed by two independent knowers:
    `TestCalleeChangeInvalidatesCallerClosureDigest`
    (`originvalidate_closure_chain_test.go:312`) and `corevalidate`'s independently
    re-derived sibling (`corevalidate_closure_chain_mutation_test.go:110`).
  - **QLT-06b — Complete-by-abstention (Phase 11), structurally gated.** No
    interprocedural fact is marked cacheable because none is cached. Discharged by
    a **structural test** asserting `cache.DeclaredInputNames()` is unchanged at
    seven names and `internal/compiler/cache` does not import
    `core`/`originvalidate`.
  A single row flipped Complete would read as "we built the closure-keyed cache,"
  which is false — the exact overclaim OWN-05's split exists to prevent.

- **D-11-40 (CORRECTION TO CARRIED EVIDENCE):** S-006's eviction figures
  (100% chain / 92% worst / 43% mean) are an **upper bound on a model, not a
  measurement**: the spike modelled an edit as unconditionally digest-moving.
  `ClosureDigest`'s preimage chains over **signature summaries, never bodies**
  (`originvalidate.go:561-590`), so a body-only edit preserving the signature does
  not move the caller's digest at all. That is *early cutoff* falling out of the
  trust boundary rather than bolted on. Do not re-quote those figures as measured.

- **D-11-41 (LIVE CACHE SOUNDNESS HOLE — fix before any cache work):**
  `internal/compiler/cgen/*.go` **is not a declared cache input.** The seven
  declared names cover `go_toolchain` and `mutation_runner_source`; neither covers
  `cgen`. Edit `cgen`, re-run with unchanged `.lang` fixtures, and `cache.Consult`
  (`session_phase6_verify.go:307`) serves a binary built by the **old** `cgen`,
  compared against the **new** interpreter. This is the ccache `__TIME__` class and
  it is **not** one of D-06-13's four declared escapes (hole 3 is *nondeterministic*
  codegen, not *changed* codegen). **Phase 11 is the phase that rewrites `cgen`.**
  Claimed, not yet confirmed — Q-02 settles it in an afternoon.

- **D-11-42:** Do not extend `SelectLanesForFixture`'s change-state to any new
  Phase 11 lane without the same scrutiny. Deferring a lane because its declared
  inputs did not move is the closest thing in this tree to a persisted verdict, and
  it shares D-11-41's hole.

### Shape reachability register (QLT-03)

- **D-11-43:** New **`qlt03_shape_register.json`** in `internal/compiler/session/`.
  Do **not** extend `qlt01_registry.json` — its row key is `spike_id` over an
  unrelated universe, and merging makes both audits ambiguous about what a missing
  row means. Reuse the *skeleton*: the discriminated-union row, the
  failure-with-control-ID shape, the free-text-placeholder ban, the
  cross-check-against-code idea.

- **D-11-44:** **Cells are enumerated mechanically; `reached` is recomputed
  mechanically; only `unreachable` is hand-written**, and every hand-written row
  names a falsifier test that goes red if the claim stops being true. Rows are an
  axis *product*, not a list, so a cell cannot be absent; an unclassified cell
  defaults to `gap` and **fails** the audit. Csmith's discipline (safety is a
  property of the *declared production space*, not of observed output) plus
  DO-178C §6.4.4.3's resolution rule (every uncovered item resolved — never left
  silent).

- **D-11-45:** Taxonomy — five axes, chosen because they are what NAT-06 and the
  call-boundary promise actually depend on: edge topology; **argument mode across
  the edge** (`move|copy|borrow_shared|borrow_exclusive`); callee body shape;
  **returned-value origin class** (`fresh_owned|param_forwarded|
  borrow_of_own_param|borrow_of_callee_result|foreign_derived`); composition-depth
  bucket. Do **not** take the full product (2,560 cells). Take **full enumeration
  of argument-mode × return-origin (20 cells)** — the pair the promise rests on —
  plus pairwise coverage over the rest. Target ≈40 rows.

- **D-11-46:** Negative claims carry **one admissible proof mechanism per class**,
  declared in the row; the audit **rejects** any class/mechanism pair not in the
  table. Classes: `grammar` (source fixture + exact refusal code), `refused`
  (construct the `core.Program`, assert the exact diagnostic), `generator`
  (enumeration over the generator's declared production space + committed op-kind
  literal), `bounded_exhaustive` (exhaustive construction below a declared,
  justified ceiling — the `pathoracle.MaxPaths` pattern), `gap` (no proof; must be
  a red test or a ratified waiver). **No free-text unreachable.**

- **D-11-47 (REJECTED AS THEATER):** "An assertion test that the shape never
  appears across the corpus" is **not** an admissible row mechanism. The generator
  is deterministic; asserting a fixed program has a fixed property is a tautology,
  and it goes vacuous the moment the generator widens while sweep sizes stay small.
  Keep it as a **drift detector only**. Likewise a prose structural argument alone
  is unfalsifiable — the *test* is the evidence.

- **D-11-48:** `witness` is not a boolean. A cell reached by <3 occurrences, in 1
  shape, at `max_depth == 1` is stamped **`reached_thin` and counts as a gap for
  gating**. This is swarm testing's finding — features compete for space and an
  all-features-on default systematically suppresses depth. Secondary defense: run
  the corpus in swarm configurations (omit shapes / omit op-kinds) to surface cells
  reachable only under omission.

- **D-11-49:** The register is the **honest home for disclosed trust gaps**,
  provided each gap row carries a falsifier that goes red when the gap closes —
  SPARK's justified-unproved-check pattern, not a TODO list. Its headline negative
  entry is the `peerDeriveOriginFacts` `OpCall` gap (see Critical Path below).

- **D-11-50:** Drift protection, in cost order: an **op-kind closure test** (assert
  the generator's emittable `core.OperationKind` set equals a committed literal —
  fires the instant anyone adds a borrow constructor, *before* any sweep runs);
  status recomputation at n ∈ {8, 24, 48}; axis-completeness cross-check against
  `CallGraphCorpusShapes()`; and a **mutation-kill** via a fault-injection seam in
  the `disablePeerOriginContainmentForTest` style, proving `AuditQLT03Register` can
  fail without editing the committed file.

### Claude's Discretion

The user's standing instruction is to research deeply, synthesize one coherent
opinionated recommendation, and proceed — escalating only genuinely high-impact
decisions. All 50 decisions above were taken under that mandate. Three warrant
explicit human attention at the planning checkpoint rather than silent adoption:

1. **D-11-09/D-11-10** — emitting zero call-boundary alias attributes materially
   weakens what NAT-05 claims, even though the evidence for it is strong.
2. **D-11-20/D-11-21/D-11-22** — three roadmap amendments to criterion 3.
3. **D-11-39** — splitting QLT-06 into a/b, one of which is satisfied by
   abstention.

</decisions>

<critical_path>
## Critical Path — one carry-forward item, four independent hits

`corevalidate.peerDeriveOriginFacts` **has no `core.OpCall` case**. It entered this
session as one of five disclosed Phase 10 carry-forwards. Four researchers hit it
independently, from four different directions:

| Route | How it bites |
|---|---|
| **Reducer (D)** | The core-level `OpCall`→`OpCopy` rewrite may be accepted or rejected on grounds unrelated to the rewrite. **Blocks QLT-05's central move** if it refuses. |
| **Mid-phase gate (C)** | A corpus with N ≥ 1 alias facts *and* ≥1 call edge may be **inexpressible**, making D-11-14's anti-vacuity floor unreachable. |
| **Shape register (F)** | The whole `borrow_of_callee_result` family is the register's headline `class: "refused"` row. |
| **Negative control (G)** | Reached it via the unreviewed D-09-51 flip; dissolved by D-11-09, but only because that decision deletes the dependent emission rule. |

**Consequence for planning:** if Q-01 fails, QLT-05 is blocked on closing a
**Phase 10 debt item, not on reducer design**. The honest split is then to keep
QLT-03 in Phase 11 and move the reducer half behind the `corevalidate` `OpCall`
closure, shipping `RefusedShapes()` as the named statement of what it does not yet
reduce. **Plan for both branches.**

</critical_path>

<experiments>
## Pre-Planning Experiments — run these first

Each is under an hour and each can invalidate a decision before a planner spends
real time on it. Q-01 and Q-02 are the two that must run before any plan is
written.

- **Q-01 (blocks D-11-29, the reducer's central move).** Build the `core.Program`
  for `testdata/phase07/deep_diamond_acyclic.lang` in a Go test, apply the
  `OpCall`→`OpCopy` rewrite by hand to one edge, run `corevalidate.Validate` and
  `check`. The *source*-level analogue is already verified to check clean; the
  **core-level** one is the fact that matters and is untested. If corevalidate
  rejects it, the design collapses to a much narrower refusal — see Critical Path.

- **Q-02 (confirms D-11-41, the live cache hole).** Set
  `Phase6CacheRootOverrideForTest` to a `t.TempDir()`, run
  `verifyPhase6NativeDifferentialLane` on one fixture to populate the store,
  capture `outcome.Key.ID`, perturb `cgen`'s emitted C for the **same** `.lang`
  source, assert the key is **unchanged** and the outcome is
  `StatusArtifactReused`. Passing means the stale-`cgen` false hit is real and
  demonstrated, and Phase 11 opens with a declared-input fix. Failing means the
  claim is wrong and the finding is withdrawn.

- **Q-03 (validates D-11-45's taxonomy; carries a falsifiable prediction).**
  `TestQLT03GeneratorOpKindClosure` — walk every
  `GenerateCallGraphCorpus(shape, n)` for all five `CallGraphCorpusShapes()` at
  n ∈ {8, 24, 48}; collect the `core.OperationKind` set and the reached
  (arg_mode, return_origin) cells. **Predicted:** kinds = `{call, copy, return}`
  exactly; the entire borrow column empty; `borrow_of_callee_result` unreached.
  Confirmation makes the register mostly transcription. A surprise means the axes
  are wrong and get replanned before any file is committed.

- **Q-04 (validates D-11-23/D-11-24).** Port the 4×3 hand-written-C matrix into
  `internal/compiler/native/native_lto_test.go`, pinned to the recorded Clang
  version, **before** any `cgen` work. An all-green matrix means the control is
  inert — do not ship it, escalate. A red cell at `-O3` *without* LTO means the
  TU split is not what buys LTO exclusivity — re-examine.

- **Q-05 (gates D-11-02's Phase 12 deletion).** Route every `testdata/phase1`–
  `phase5` fixture through `emitProgram`, compile at `-O0`, diff the
  `lang.execution/1` document against `interp.Run`. **Bytes are allowed to differ;
  the document is not.** If this cannot go green, the deletion path is dead and two
  conventions are permanent.

- **Q-06 (evidence for D-11-11, ~1 hour, no compiler).** Build the two-function
  analogue of `false_restrict_hoist.lang` — a caller passing the same place as
  argument at two `OpCall`s into a `restrict`-promising callee — and assert
  `corevalidate.ValidateEmittedAttributes` returns `core.attribute_unjustified`
  **before any C is emitted**. Run on the zero-attribute side of the mid-phase
  gate.

- **Q-07 (validates D-11-03).** Hand-write the target C for a three-function
  program (one foreign-shaped, one branch-shaped, one plain linear), compile with
  the installed Clang, run, diff against `interp`. If it cannot be hand-written
  readably, no emitter will generate it readably.

</experiments>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase charter and carried state
- `.planning/ROADMAP.md` § "Phase 11: Multi-Function Native Emission and
  Interprocedural Equivalence" — goal, 5 success criteria, mid-phase gate,
  riskiest assumption, and the 2026-09-11 32-guard scope input
- `.planning/STATE.md` § Blockers/Concerns "Phase 10 carry-forward — READ BEFORE
  PLANNING PHASE 11" — the five residual trust gaps, restated there because at
  this project's 200k context window the planner does **not** auto-load
  prior-phase files
- `.planning/REQUIREMENTS.md` — NAT-04, NAT-05, NAT-06, NAT-07, QLT-03, QLT-05,
  QLT-06; and the "Never cut" list (NAT-06, NAT-07 are on it)
- `.planning/LANGUAGE-MATURITY.md` § "The single-function guard inventory" —
  the 32-guard per-package breakdown with its re-verify command
- `.planning/STANDING-VERDICTS.md` — dependency adopt/reject verdicts and
  anti-features; the zero-external-dependency record is a hard constraint

### Phase 10 detail (inputs, not closed history)
- `.planning/phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` — debt
  register and the D-10-59 gate verdict
- `.planning/phases/10-trusted-interprocedural-oracle/deferred-items.md`
- `.planning/phases/10-trusted-interprocedural-oracle/10-REVIEW.md` — WR-01
  (`peerCalleeFrameDrained`'s unstated ordering assumption), WR-02/D-10-19 (the
  `callgraph` import asymmetry that makes peer independence review-enforced
  rather than mechanism-enforced — **NAT-06 leans on that independence**)
- `.planning/phases/10-trusted-interprocedural-oracle/10-VERIFICATION.md`,
  `10-VALIDATION.md`
- `.planning/phases/10-trusted-interprocedural-oracle/10-02-SUMMARY.md`
  § Deviations — the **unreviewed** negative-control verdict flip

### Research corpus
- `.planning/research/ARCHITECTURE.md` §6 — Stage 7 (multi-function C emission +
  call lowering, "the second-highest-risk integration point") and Stage 8's
  closing verification
- `.planning/research/PITFALLS.md` Pitfalls 3, 5, 6 — the three that became
  criteria 3, 4 and 5
- `.planning/spikes/006-interprocedural-liveness-cost-scaling/README.md` — but
  see D-11-40: its eviction figures are an upper bound on a model
- `.planning/spikes/MANIFEST.md`, `.planning/spikes/CONVENTIONS.md`
- ⚠ Spikes are unpackaged (no findings skill). S-006's cost findings already
  landed in Phase 08; treat the rest as raw.

### Wiki (intent and constraints)
- `wiki/compiler-and-feedback-latency.md` — feedback latency as a first-class
  constraint; governs D-11-38 and the register's cost
- `wiki/compute-efficiency-constitution.md`
- `wiki/semantic-kernel-contract.md`, `wiki/semantic-kernel-probes.md`
- `wiki/ownership-evidence-roadmap.md`
- `wiki/residual-uncertainty-register.md`
- ⚠ **`wiki/example-tour.md` is design fiction** — effect rows, `?` propagation,
  generics, `spec`/`property` blocks. None of those tokens are in the lexer. Read
  `internal/compiler/syntax/token.go` for what the language actually is.

### Code — the sites this phase touches
- `internal/compiler/cgen/cgen.go` — 29/59 (`Emit`/`EmitNative` guards), 152, 234,
  361 (`borrowByPointerMarker`), 400, 456, 498 (the `restrict` write), 589, 649,
  760-790 (`emittedAttributeForByPointerParameter`), 793, 1057, 1468, 1507, 1559,
  1579, 1597, 1685, 1830-1845 (`EmittedAttribute`), 1868, 1908, 1980-1983 (D-04-12
  two-sources-of-truth ruling), 2068 (`BannedOptimizerAttributes`), 2077
  (`NoreturnExemption`), 2084 (`ScanForBannedAttributes`), 2109
  (`JustifiableAttributes`), 2113 (`ScanForUnjustifiedAttributes`)
- `internal/compiler/callgraph/callgraph.go` — 215 (`buildAdjacency`), 233-245
  (`core.CallCalleeUnresolved`, D-07-45), 294 (`Order`)
- `internal/compiler/session/session.go` — 510, 773, 886 (the three run sites that
  must share the entry resolver)
- `internal/compiler/session/session_phase5_compare.go` — 20-24 (the five axes),
  71 (`Phase5CompareEngines`), 145 (`Phase5CompareDiagnosticIDs`)
- `internal/compiler/session/session_phase5_alias.go` — 46, 67, 351
  (`AliasFactMutationRunner`), ~147 (`injectRestrictIntoSignature`)
- `internal/compiler/session/session_phase5.go` — 44 (`control:alias.false_no_alias`),
  130-175 (valid/corrupt `AttributeClaim` pair), 226-231 (`inline_across_foreign.lang`)
- `internal/compiler/session/session_phase5_mismatch.go` — 39-43 (the three
  `control:reduce.*`), 72-75 (`foreignCallSequenceFor`, D-11-33), 117-126,
  161-204 (`mismatchPredicate`), 265/279 (the two production `Reduce` call sites),
  309-319
- `internal/compiler/session/session_phase6_verify.go` — 38, 103-121, 307
  (`cache.Consult`), 324 (`store.Put`), 409
- `internal/compiler/session/session_phase6_risklanes.go` — 315-340, 417, 443
- `internal/compiler/session/session_phase6_escapes.go` — 16-45 (D-06-13's four
  declared escapes; D-11-41 is **not** among them)
- `internal/compiler/reduce/reduce.go` — 31 (`MaxReductionAttempts`), 35-38
  (`Minimality`), 53-61 (`Moves`, order is the determinism guarantee), 183, 184,
  214, 266, 474-477, 484, 584, 641, 686
- `internal/compiler/reduce/predicate.go` — 63-66 (the `CausalRole` relaxation
  D-11-32 must not inherit)
- `internal/compiler/cache/cache.go` — 82-105 (`ComputeKey`), package doc (D-06-06,
  the verdict prohibition)
- `internal/compiler/cache/probe.go` — 33-50 (the seven declared input names),
  56-58 (three artifact kinds)
- `internal/compiler/core/core.go` — 29-52 (`Function`, no export field), 212
  (`Callable`), 222-240 (`ClosureDigest`; **234-239: staleness, not authenticity**),
  315-339 (the `/0` pinning pattern), 480-483, 590
- `internal/compiler/originvalidate/originvalidate.go` — 525-590
  (`ClosureDigestDomainSeparator`, the preimage — D-11-40's basis)
- `internal/compiler/corevalidate/corevalidate.go` — 2407-2450
  (`peerDeriveOriginFacts`, **the `OpCall` gap**), 2452-2460
  (`disablePeerOriginContainmentForTest`, the seam style D-11-50 copies), 3469
  (`AttributeClaim`), 3502 (`recomputeAliasJustifications`), 3573
  (`ValidateEmittedAttributes`)
- `internal/compiler/native/native.go` — 57-66 (`Options.LTO`), 153-158 (the single
  `program.c` D-11-24 does **not** widen), 163-197 (`ForeignSources` loop)
- `internal/compiler/native/native_lto_test.go` — 72-97 (Q-04's host)
- `internal/compiler/testsupport/callgraphcorpus.go` — `GenerateCallGraphCorpus`,
  `CallGraphCorpusShapes` (D-09-46)
- `internal/compiler/session/qlt01.go` — 32-39, 55, 84-90, 92-160 (the skeleton
  D-11-43 reuses)
- `internal/compiler/check/check.go` — 639-647
  (`buildInterproceduralSummariesObserved`)
- `internal/compiler/syntax/token.go` — the complete token set; the calibration
  artifact for what is expressible
- `qlt01_registry.json`, `qlt02_budget_manifest.json` — the two existing registries

### Fixtures
- `testdata/phase07/call_basic.lang`, `deep_diamond_acyclic.lang` (13 functions),
  `call_argument_used_twice.lang` (the refusal proving D-11-09),
  `cycle_unreachable.lang` (three roots)
- `testdata/phase10/relay_depth3_accept.lang` — its ~50-line header documents the
  `peerDeriveOriginFacts` gap
- `testdata/phase5/restrict_borrow.lang` + `.golden.c` (the only committed golden
  containing a `restrict` site), `false_restrict_hoist.lang`,
  `inline_across_foreign.lang`
- `testdata/phase08/negative_control_fails.lang`,
  `negative_control_infallible.lang` — the two whose verdict flip is unreviewed

### External prior art (surfaced by research; cite in plans where it settles a choice)
- LLVM `OptBisect` / `-opt-bisect-limit` — first-class transformation control
  instead of fixture curation (D-11-17)
- Rust `noalias` history: #31681, #54878, #82834 (re-enable on LLVM 12), #84958
  (immediate regression) — the direct precedent for D-11-09
- "Restrict-Qualified Pointers in LLVM" (Finkel 2017); "ptr_provenance and
  llvm.noalias: The Tale of Full Restrict" (2021) — why the promise becomes
  block-scoped metadata after inlining
- Csmith (PLDI'11) and its documented exclusion list; Swarm Testing (ISSTA'12) —
  D-11-44 and D-11-48
- "Build Systems à la Carte" — early cutoff, D-11-40
- rust-lang/rust#82920 — incremental-compilation miscompilation, D-11-38's
  asymmetry argument
- J-Reduce / "Binary Reduction of Dependency Graphs" (ESEC/FSE'19) — D-11-29
- MacIver, "Notes on Test-Case Reduction" — reduction slippage, D-11-32
- Nim `genProc`/`genMainProc`, Vala `CCodeFunction`/`CCodeFile`, Cython
  `UtilityCode` — the body-vs-TU split in D-11-03

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- **`callgraph.Order`** — iterative white/gray/black DFS, no native recursion,
  deterministic cycle-witness selection. Gives declaration order, prototype order
  and the reducer's reverse-postorder for free. Its doc comment pins its consumer
  set to `check` alone; D-11-29 deliberately avoids adding `reduce` to it.
- **`AliasFactMutationRunner` + `injectRestrictIntoSignature`** — already
  idempotent and marker-guarded, fail-closed via `native.alias_control_invalid`.
  D-11-23 extends the shape rather than inventing one.
- **`corevalidate.ValidateEmittedAttributes` / `AttributeClaim`** — the proven
  D-05-01 independent validator; the second knower for any attribute claim.
- **`testsupport.GenerateCallGraphCorpus` / `CallGraphCorpusShapes`** — five
  shapes, already relocated out of `check`'s test file (D-09-46) so `session` can
  consume it with no production import of `check`. **Known limitation: the five
  shapes contain no borrow op at all** (Phase 09 finding) — this is Q-03's
  predicted headline negative row.
- **`session.qlt01.go`'s audit skeleton** — discriminated-union rows,
  failure-with-control-ID, placeholder ban, cross-check-against-code.
- **`Phase6CacheRootOverrideForTest`** — the hook Q-02 needs.
- **`disablePeerOriginContainmentForTest`** — the `export_test.go`-only
  fault-injection seam style D-11-50 copies.
- **`native.Options.LTO`** — already built so "did `-flto` reach clang" and "did
  LTO change codegen" stay separately observable. Same epistemic move as D-11-17,
  one layer down.

### Established Patterns
- **Two independent knowers, never one derivation read twice.** Governs D-11-14,
  D-11-15, D-11-33, D-11-39. Where a design cannot name a second knower, it is not
  evidence.
- **Additive sibling over signature change** (D-04-20, `cgen.go:442-446`) — how
  D-11-16 adds a profile without touching ~20 existing call sites.
- **Declared bounds are themselves the claim** — `pathoracle.MaxPaths` (bound 4096
  against a reachable maximum of 64), `interp.MaxCallDepth = 128` set deliberately
  *below* the 1024-function structural ceiling. D-11-31 and D-11-46's
  `bounded_exhaustive` follow it.
- **Derived fail-closed bounds with named refusals** —
  `check.loan_liveness_bound_exceeded` and its `+1` floor lesson (a literal
  `blockCount*loanCount` formula computes zero for a loan-free function and would
  refuse nearly every legal program). D-11-31 inherits both the pattern and the
  warning.
- **Build-then-delete** (D-09-10) — peer lands fully, zero-divergence differential
  goes green, only then the deletion. D-11-02 is the same move.
- **Split a requirement rather than overclaim** (OWN-05a/05b) — D-11-39.
- **Schema `/0` pinning** (`core.InterfaceV0`, `FunctionSignatureV0`) — the
  fallback shape if `lang.attributes/0` is refused.
- **Standing rule from three M001 gate failures:** *a green test whose reachable
  input space omitted the hard case.* D-11-14, D-11-18, D-11-34, D-11-48 each
  exist to defeat one instance of it.
- **Drive the shipped binary on hand-written programs, not only the gate's own
  corpus.** D-11-24 and Q-04/Q-07 are direct applications.

### Integration Points
- `cgen.Emit`/`EmitNative` at 29/59 — the dispatch fork into `emitProgram`
- `session.go` 510/773/886 — the three run sites and the entry-resolver switch
- `session_phase5_mismatch.go` 265/279 — the two production `Reduce` call sites
- `session_phase6_verify.go` 307/324 — the single production `cache.Consult`/`Put`
- `cache.DeclaredInputNames()` — the closed seven-name list D-11-39b tests and
  D-11-41 must extend
- `qlt02_budget_manifest.json` — any new gate-eligible metric needs a **ratified
  row**, never a raised bound (`corevalidate_peer_cost_test.go:24` precedent)

</code_context>

<specifics>
## Specific Ideas

- **The user's research mandate, applied to every decision above:** fan out across
  all relevant stakeholder/role lenses, consider pros/cons/tradeoffs, anti-patterns,
  best practices, footguns and lessons learned, research external ecosystems, run
  an adversarial pass, then synthesize **one** decisive recommendation. No menus.
  Seven researchers ran that protocol; two were given standing permission to return
  "the requirement is wrong" and one exercised it (D-11-20/21/22).
- **Measured, not asserted.** G compiled and ran a 4×3 matrix on this host and
  dumped IR; D compiled probe programs to verify that orphan functions, devirtualized
  calls and forward references all check clean. Plans should preserve this habit:
  where a claim about Clang or about `check` is load-bearing, compile something.
- **Honest scope beats a green row.** Three decisions deliberately claim *less*
  than the roadmap's prose implies (D-11-10, D-11-25, D-11-39). That is the
  intended direction of travel for this project.

</specifics>

<deferred>
## Deferred Ideas

- **Deleting the six single-function emitters** — Phase 12, gated on Q-05
  (D-11-02).
- **Call-boundary alias attributes and the `EmittedAttribute` discharge-pair
  schema** — designed in D-11-11/D-11-12, built only when D-11-09 is revisited;
  earliest sensible home is a phase where Lang gains a construct that can create
  two pointers to one object. Nothing on the current roadmap does.
- **`lang.attributes/0`** — mint it with the discharge schema, not before; it also
  closes D-11-13's evidence-binding gap.
- **Per-function translation units / widening `native.Runner` beyond one
  `program.c`** — only if criterion 2's LTO tier must be non-inert over the corpus
  itself rather than over NAT-07's hand-written control (D-11-24/D-11-25).
- **A closure-keyed native cache** — revisit only if Lang gains separate
  compilation units (M003), which would make whole-program hashing stop dominating
  (D-11-38).
- **Reviewing the D-09-51 negative-control verdict flip** — dissolved as a Phase 11
  precondition by D-11-09, but still an open, unreviewed change to two fixtures'
  expected diagnostics. Belongs to whoever revisits SEM-06's gate ordering.
- **`peerCalleeFrameDrained`'s forward-pass ordering assumption** (10-REVIEW WR-01)
  and the **`callgraph` import asymmetry** (WR-02/D-10-19) — both remain
  review-enforced rather than mechanism-enforced. NAT-06 leans on the latter; if a
  plan needs that independence to be *mechanical*, it is new scope.
- **Flaky-predicate tolerance in the reducer** — explicitly not built (D-11-36).
- **Arithmetic, iteration, strings, arrays** — on no roadmap at all. M003 does not
  exist yet.

</deferred>

---

*Phase: 11-multi-function-native-emission-and-interprocedural-equivalence*
*Context gathered: 2026-09-11*

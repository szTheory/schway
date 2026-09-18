# Codename Lang

## What This Is

Codename Lang is a general-purpose programming language and toolchain for
AI-authored, human-audited production software. It aims to combine fast,
structured feedback with a small coherent semantic core, usable low-level
ownership and resource control, deterministic evidence, and a calm canonical
source form that remains readable when humans mostly review rather than type it.

The first two milestones were deliberately narrower than the full vision: prove
one real source-to-native semantic spine (M001), then prove it survives a
function boundary (M002), before expanding into effects, services, agents,
databases, GUIs, or an ecosystem. Lang-to-Lang calls now exist end to end; loops,
arithmetic, and generic user code still do not.

## Core Value

Give an AI agent and a human reviewer the shortest reliable path from intent to
sound, reproducible evidence without wasting iteration time or hiding runtime
costs.

## Current State

**Shipped M002 — Interprocedural Semantic Spine (2026-09-14).** Phases 07-13,
61 plans, 183 tasks over 6 days; 386 commits, 441 files changed,
+130,172 / -1,527 lines against `f2490a8`.

`OpCall` is a real `core.OperationKind` wired at all six dispatch sites
(`check`, `corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate`).
A multi-function Lang program parses, checks, is independently re-validated by
three non-importing peers, interprets, lowers to multi-function C17, and agrees
across interpreter / `-O0` / `-O3` / `-O3 -flto` on the five-axis comparator.
`Result` values carry payloads. `lang-repair` fixes interprocedural loan-liveness
defects through the JSON protocol alone.

**Phase 14 complete (2026-09-18) — Evidence Instrument and Honest Scoping.**
13 plans. The systemic finding below — *the instrument reports green because
something is wired, not because it runs* — now has a mechanism rather than a
set of patches: claims carry evidence grades with declared ceilings, every guard
carries a seeded-fault proof, the corpus-wide run record cannot report a result
it did not finish producing, and the `>=EXERCISED` bar is applied reflexively to
Phase 14's own validation artifact. Debt rows require a named owner; zero
`UNOWNED(none-yet-scheduled)` rows remain in `PHASE-14-DEBT.md`.

**Known gaps carried into M003** (corrected 2026-09-17 against the tree by the
M003 research fan-out — `.planning/research/M003/`; three claims previously
recorded here were wrong and are restated below):

- **DX-06 (partial, and worse than recorded)** — this section previously said
  the contract-boundary blame resolver was "built, tested, and compile-time
  exhaustive, but structurally unreachable," and that it would close
  automatically once `sameType(ReturnType, Parameter.Type)` lifted. **Both
  halves are wrong.** The *entire* blame subsystem — `resolveBlame`,
  `resolveCycleBlame`, `classifyDeclaredCause` — is test-only dead code: zero
  production call sites, and nothing ever sets `blameFact.Violated = true`
  (`classifyDeclaredCause` hardcodes `false`; `true` occurs only in
  `check_blame_test.go`). Lifting the invariant does **not** close it: a
  B1-shaped diagnostic needs a *user-declared* contract field that the
  declaring function's own admission cannot verify, and every
  `FunctionSignature` field today is producer-derived. B1's real home is
  separate compilation (M006). M003 corrects D-13-02b's reopening condition
  rather than restating the automatic-closure claim a third time.
- **DX-07 (partial)** — two of three new repairable interprocedural classes
  ship (`move_after_interprocedural_loan`, `wrap_call_in_try`);
  `use_matching_argument` was withdrawn empirically, its repair byte-identical
  to the original on every real trigger. See `PHASE-13-DEBT.md` D-13-10a. This
  one *does* close when the invariant lifts.
- **`-flto` is structurally inert on every cgen-emitted multi-function
  program** — one translation unit, no `restrict`, foreign refused. Both live
  non-inertness proofs are non-multi-function: `TestLTOTierIsNotInert` uses a
  single-function foreign fixture, and NAT-07's composition-only control is
  hand-written C explicitly not emitted by cgen (D-11-24). D-11-25 names this
  only in a test doc comment — it has **no debt row** and is **not** among the
  ten unowned items — while this document's NAT-07 bullet previously read as
  though the proof covered the interprocedural corpus. It does not.
- **Nyquist validation is partial, and "cheap to close" was only half true** —
  Phases 12 and 13 are literal frontmatter flips; Phase 07 is roughly an hour.
  But Phase 11 has six or more command cells elided with `…` and therefore
  un-runnable, and **three `go test -run` patterns across the VALIDATION corpus
  resolve to zero tests and exit 0 with `[no tests to run]`** — two in Phase 08
  (one naming a mechanism deleted in Phase 09) and one in **Phase 09**, a phase
  certified `validated` / `nyquist_compliant: true` on a green row whose
  evidence is a summary's prose. These are provenance defects in the measuring
  instrument, not coverage holes; the substance is intact.
- **10 open, unowned debt items** in three clusters: the single-function
  emitter work (D-11-02 → D-12-36 — a **port-then-delete**, not a deletion:
  `emitProgram` is a strict *subset* that refuses `Match`, `Blocks` and
  `ForeignContract`, emits zero `restrict`, and hardcodes
  `"live_resources":[]`), event identity (D-11-51 → D-12-21), and the
  single-type-per-function invariant (D-13-02b, D-13-10a).
- **D-13-34 points backwards** — M001's `testdata/phase6` move and borrow
  held-out pairs are alpha-renames of each other (match genuinely differs).
  Blast radius determined: DX-04's written claim survives on corpus-independent
  evidence; what is void is D-06-29's anti-overfitting inference for two of
  three source classes. M002 is unaffected.

**The systemic finding.** Those are not six unrelated defects. DX-06, the
`-flto` inertness, the zero-test VALIDATION patterns, and
`TestEveryMutationMovesItsClaimedAxis` silently skipping two rows are four
instances of one failure mode: *the instrument reports green because something
is wired, not because it runs.* It is the same mode the M002 audit caught in
the integration checker. M003 Phase 14 mechanizes the fix rather than patching
the instances.

## Current Milestone: M003 Computation and Honest Instruments

**Goal:** Prove a meaning survives *computation* — the first value Lang creates
rather than moves — on instruments that cannot report green for work that is
merely wired.

**Thesis:** M001 proved one meaning survives lowering. M002 proved it survives a
function boundary. M003 proves it survives computation, and makes the measuring
instruments honest enough that the claim means what it says.

**Target features:**

- Evidence instruments that grade at EXERCISED+ or name themselves unreachable —
  a groundedness lint, a closed evidence vocabulary
  (`DEFINED | WIRED | REACHABLE | EXERCISED | MUTATION-KILLED`), and
  `.planning/UNREACHABLE-CLAIMS.md` for criterion-shaped defects. The three
  wrong claims above are corrected here.
- Event identity that survives a shared-leaf diamond — context-path invocation
  identity under a `lang.execution/2` bump, uniqueness moved to
  `(invocation, id)`, independently re-derivable by a non-importing peer
  because the call graph is a guaranteed DAG.
- One emission law instead of two — branch/match ported into `emitProgram`,
  three emitters deleted, the other three formally cut by a D-10-60 amendment
  with M004 landings.
- A function's return type may differ from its parameter type.
- A branch that can discriminate a **computed** value — today `match` is a
  whole-function-body form whose scrutinee must be the function's own
  parameter, so branching on a computed value does not exist and cannot be
  reached by desugaring. Hard-gated on spike S-010.
- Literals and `OpConst` — Lang can name a value it was not given.
- A moved refusal frontier, pinned by a test.

**Explicitly out of this milestone:** `if` as surface syntax, comparison
operators, `Bool`, `OpBinary`, loops and back edges, arity-N, DX-06 closure,
the three cut emitter families, and DX-08/09/11/13.

**The governing gate — constructibility precondition.** No requirement is
admitted unless its `.lang` fixture is checked in *first* as a refused frontier
fixture with its diagnostic pinned by a test; the phase gate is that the pinned
diagnostic moved. This is the only gate that fires *before* the work, and it
would have refused comparison operators on day one — a `Bool` that cannot be
branched on is the DX-06 failure mode for the fourth time.

**Named risk.** Arithmetic is ownership-inert (~15% redesign risk), but **any
branch on a computed value is ~60%** against `loanLivenessFixpoint`, the
comparator, and `corevalidate`. Early warning: a `borrow` created before a
branch and live in one arm has no `LoanEndpoint` classification today. That is
why the branch phase is spike-gated, and why the bet is declared lost — with
M004 becoming an assurance-refactor milestone — if arithmetic forces a
*redesign* rather than an extension.

**WIP limits.** M003 may not close with more than 5 unowned debt items. Every
debt item names an owning phase when it is recorded.

## Milestone Arc (M004–M006)

Provisional, revised at each milestone boundary. Recorded here so the M003
research is not re-derived later.

| Milestone | Charter | Gate that becomes meaningful |
|---|---|---|
| M004 | Iteration | Back edges exist. Must first answer what replaces `pathoracle`, which refuses every CFG cycle by name (`pathoracle.cfg_back_edge`) — iteration does not make it expensive, it makes it impossible. A pre-phase spike is mandatory. |
| M005 | Aggregates and arity-N | D-12-43 finally becomes constructible — via a generalized-scrutinee arm returning a destructured payload, not via arithmetic alone. Arity 2 takes the `pathoracle` case space from ~12 to ~108. |
| M006 | Modules and separate compilation | DX-06's B1 blame becomes real: the first user-declared contract field the declaring function's own admission cannot verify. |

**First real program:** `examples/checksum.lang` — read a file through foreign
C, loop-accumulate a byte checksum, print it. Lands end of M005. Checked in
during M003 as a *refused* frontier fixture with its diagnostic pinned, so each
milestone must measurably move the refusal forward.

## Requirements

### Validated

- ✓ A clean checkout can format, check, interpret, and natively run a small
  canonical Lang program through one production command — M001 (Phase 1;
  `lang format` / `lang check` / `lang run --engine={interp,native}`).
- ✓ One portable typed core gives ownership, borrowing, cleanup, and foreign
  boundaries the same meaning in the checker, interpreter, and native path —
  M001 (Phases 2-4; `check`, `corevalidate`, `originvalidate`, `interp`, `cgen`
  independently derive the same facts).
- ✓ Interpreter, native `-O0`, and native `-O3` executions produce equivalent
  semantic outcomes and events for the milestone corpus — M001 (Phase 5;
  five-axis comparator, `-flto`-proven-non-inert `-O3` tier, ASan/UBSan lane,
  adversarial + bounded-enumeration corpora). Scope note: M001 ships without
  Lang-to-Lang calls, so this holds intraprocedurally by construction.
- ✓ Invalid programs fail early with stable, bounded, machine-readable cause
  graphs and calm human projections — M001 (Phases 1-3, 6; `lang.diagnostic`,
  `lang explain` cause DAG, rustfix-style `diagnostic.Repair` applicability).
- ✓ Compact content-bound evidence manifests make successful work reusable
  without treating frontend claims as proofs of source correctness — M001
  (Phases 1-2, 6; `evidence --validate`, `internal/compiler/cache` which is
  structurally incapable of holding a verdict).
- ✓ The default edit loop remains fast, deterministic, offline-capable, and
  measurable; expensive evidence runs only when its distinct question is
  relevant — M001 (Phase 6; `risk_lanes.json` changed-risk lane selection,
  `internal/compiler/measure` p50/p95/CoV protocol, `qlt02_budget_manifest.json`).

- ✓ Storable/matchable `Result` values with payload-carrying alternatives —
  M002 (Phase 12; `OpConstructPayload`/`OpDestructurePayload` with a single
  shared `core.AlternativeNameForPayloadType` derivation read by `check`,
  `interp`, and `cgen`, and affine drop obligation held on the unmoved
  alternative). Closes D-04-30, deferred from M001 Phase 4. Scope note: a
  `data` declaration whose alternatives collide on payload type is refused
  (`check.duplicate_payload_type`) rather than resolved — see D-12-44.

- ✓ Lang-to-Lang calls: `OpCall` is a real `OperationKind` at all six dispatch
  sites, gated on callable ⊆ publishable (D-04-03) — M002 (Phase 07;
  SEM-04/05/06/07, both exhaustive-dispatch controls green, digest-bound
  signature summaries, no caller admission ever reads a callee body).
- ✓ Call-graph construction refuses cycles — direct, mutual, and indirect —
  with a named code, never a hang; the interpreter executes calls on a bounded
  stack with a documented ceiling — M002 (Phases 07, 10; SEM-07, SEM-08).
- ✓ Interprocedural loan liveness is derived in both admission layers from
  signatures alone and independently re-derived by a non-importing peer —
  M002 (Phases 08-09; OWN-06/07/08/09, EFF-02). **Closes D-03-02**, the one
  debt item knowingly carried past M001.
- ✓ Cross-function loan-endpoint differentials rebuild Phase 3's exhaustive
  enumeration at a declared bounded composition depth, with `pathoracle` and
  `originvalidate` re-deriving without importing `check` or `corevalidate` —
  M002 (Phase 10; QLT-04, TRU-02, TRU-03, TRU-04; all three transitive-
  dependency guards clean).
- ✓ Interpreter, `-O0`, `-O3`, and `-O3 -flto` produce equivalent semantic
  outcomes and events for the interprocedural corpus, with the tier proven
  non-inert by an engineered composition-only divergence — M002 (Phase 11;
  NAT-04/05/06/07, QLT-03/05/06). This is the claim M001 could not make,
  because M001 ships without calls.
- ✓ `lang explain`'s cause DAG stays bounded across function boundaries and
  names the function each cause step belongs to — M002 (Phase 13; DX-05).
- ✓ Every shipped claim is graded at EXERCISED or above, or names itself
  unreachable with an unblocking trigger, so no instrument can report green for
  work that is merely wired — M003 (Phase 14; evidence grades with WIRED /
  EXERCISED ceilings, seeded-fault proofs on every guard, a completion-witnessed
  corpus-wide run record with a measured timing margin, a live `//go:build`
  suppression allowlist, and owner-required debt rows). Validated in Phase 14:
  Evidence Instrument and Honest Scoping.

### Active

M003 scope, confirmed 2026-09-17. Each line is a milestone-level intent;
`REQUIREMENTS.md` carries the testable REQ-IDs.

- [ ] Give event identity an owner: shared-leaf diamond call graphs collide
      (D-11-51), `OpCall` emits no event at all, and D-12-21 cannot close
      until both are fixed.
- [ ] Reduce two emission laws to one: port branch/match into `emitProgram`,
      delete three emitters, formally cut the other three under a D-10-60
      amendment rather than a silent third deferral.
- [ ] Allow a function's return type to differ from its parameter type.
      Closes DX-07 and D-13-10a. Does **not** close DX-06 — see Current State.
- [ ] Let a branch discriminate a computed value, by generalizing `match`'s
      scrutinee beyond the function's own parameter. Spike-gated.
- [ ] Let Lang name a value it was not given: literals and `OpConst`.
- [ ] Close the reconcilable Nyquist debt (07, 08, 11) and pin a refused
      frontier fixture whose diagnostic each milestone must move.

Deferred with a named landing, not dropped:

- Loops and back edges → M004, after a spike answers what replaces
  `pathoracle`.
- Arity-N and aggregates → M005, where D-12-43 becomes constructible.
- DX-06 / B1 blame → M006, with separate compilation.
- D-13-34's alpha-renamed held-out pairs → attached to the grammar-widening
  phase; the admissible space is 112 programs, so "structurally distinct" is
  cosmetic at this maturity.

### Out of Scope

- General effect rows, resumable handlers, and runtime dependency injection —
  preserve typed-core seams, then design from a working ownership spine.
- Async/await, actors, schedulers, backpressure, supervision, or managed cyclic
  heaps — defer until sequential resource semantics survive native lowering.
- Package registry, remote cache, LSP, MCP server, web/SQL/TLS/GUI kits, AI
  inference, and distributed workflows — later milestones built on the stable
  compiler service contract.
- Full generics, subtyping, variance, partial moves, borrow across suspension,
  reflection, macros, and user-defined precedence — they would broaden the
  search space before the kernel oracle is trustworthy.
- A permanent compiler host or final optimizing backend decision — Go 1.24 and
  readable C17/Clang remain reversible Stage 0 choices; neither M001 nor M002
  gave a reason to revisit either, and M002's multi-function C17 emission and
  `-flto` tier exercised the backend harder than M001 did.
- Recursive programs — `callgraph`'s three-color DFS refuses every cycle before
  any liveness derivation runs, so a liveness differential over an admitted
  recursive program structurally cannot exist (TRU-04, D-09-22). Revisit only
  alongside a bounded-recursion admission design.

## Context

**Shipped M001 — Source-to-Native Semantic Spine (2026-09-07).** 6 phases,
62 plans, 174 tasks over 5 days. Details in
`.planning/milestones/M001-ROADMAP.md`.

**Shipped M002 — Interprocedural Semantic Spine (2026-09-14).** 7 phases
(07-13), 61 plans, 183 tasks over 6 days; 386 commits, 441 files changed,
+130,172 / -1,527 lines, 61 new `.lang` fixtures. ~102,280 lines of Go across
25 packages; `go build ./...`, `go vet ./...`, and `go test ./...` are clean.
Details in `.planning/milestones/M002-ROADMAP.md`.

What exists now is a real vertical compiler that crosses function boundaries:
lossless canonical frontend → typed core with a per-function CFG and `OpCall` →
an affine ownership and borrow checker with interprocedural loan liveness → three
independent, non-importing validators (`corevalidate`, `originvalidate`,
`pathoracle`) → deterministic interpreter oracle on a bounded call stack →
multi-function readable C17 through Clang at `-O0`/`-O3`/`-flto`, with an
ASan+UBSan lane, a core-level HDD reducer, and a content-bound evidence/cache
layer keyed on the call-graph closure. `Result` values carry payloads.
`cmd/lang-repair` fixes interprocedural defects through the JSON protocol alone.

**M001's structural gap is closed.** `OpCall` landed at all six dispatch sites,
D-03-02 closed in Phase 09, and interprocedural `-O3`/LTO equivalence is now
inside the proof scope with an engineered composition-only negative control
proving the tier non-inert.

**M002's principal outstanding liability is its debt register, not its
requirements.** 77 items across 7 registers, 26 closed, 0 open blockers, but 10
open and unowned — which is why the milestone audit's status is `tech_debt`
rather than `passed`. The three clusters, their dependency chains, and the two
honest partials (DX-06, DX-07) are enumerated in `## Current State` above and in
full in `.planning/milestones/M002-MILESTONE-AUDIT.md`.

**One process observation worth carrying forward.** The M002 integration checker
returned an unqualified "all 31 requirements satisfied, ready for release,"
marking DX-06 and DX-07 `✓ WIRED` on the strength of wiring alone. Its
structural findings held under spot-check and were adopted; its
requirement-satisfaction column was rejected against the tree. An integration
checker that grades requirements from wiring will systematically convert an
honest partial into a false green, because wiring is exactly what a structurally
unreachable defect class still has.

The design corpus in `wiki/` records intent, primary-source research, synthesis,
decisions, examples, and open questions with explicit provenance. Five bounded
workbenches in `.planning/spikes/` tested local affine ownership, CFG
edge-specific loan liveness, separate-compilation origins and generic abilities,
independent typed-core certificate validation, and native C ABI/provenance/
cleanup hazards; their surviving contracts are now production code, and their
hazards are reconciled row-by-row in `qlt01_registry.json`.

Canonical planning inputs:

- `.planning/LANGUAGE-MATURITY.md` — where the language actually stands
  (expressiveness vs assurance) versus how the roadmap sounds. **Read this
  before answering any "how far along are we" question.**
- `.planning/STANDING-VERDICTS.md` — dependency verdicts, anti-features, and
  load-bearing facts already researched. Reopen only on new evidence.
- `wiki/real-frontend-core-ir-entry-plan.md`
- `wiki/semantic-kernel-contract.md`
- `wiki/semantic-kernel-probes.md`
- `wiki/ownership-evidence-roadmap.md`
- `wiki/compiler-and-feedback-latency.md`
- `wiki/compute-efficiency-constitution.md`
- `wiki/residual-uncertainty-register.md`
- `.planning/spikes/MANIFEST.md`

## Constraints

- **Feedback latency**: Optimize for the whole generate → verify → run → observe
  → repair loop; report cold and warm distributions rather than one flattering
  number.
- **Correctness**: Safe code has defined behavior; optimizer attributes and FFI
  guarantees must derive from checked facts.
- **Low-level capability**: Usable ownership, borrowing, deterministic cleanup,
  target layout, and C interop are milestone requirements, not post-v1 add-ons.
- **Runtime posture**: No mandatory global tracing heap or service runtime; the
  first native slice uses stack/inline values and explicit resources.
- **Bootstrap**: Go 1.24 standard library hosts Stage 0; readable C17 emitted to
  installed Clang is the development native path.
- **Dependencies**: Prefer standard-library-only and shallow audited boundaries;
  every dependency must earn more than copy-local implementation.
- **Source authority**: Text remains Git-friendly and formatter-owned; parsing is
  lossless enough to preserve comments and produce stable recovery.
- **Evidence**: Deterministic fixtures, negative controls, properties, mutation,
  differential execution, and sanitizers answer distinct questions and have
  explicit cost lanes.
- **Portability**: macOS and Linux are the initial host priorities; wire contracts
  and target facts must not accidentally encode the current Apple arm64 host.
- **Security**: Untrusted input, unsafe operations, FFI, secrets, and build
  authority stay explicit; no generic `untaint` or ambient authority escape.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Build M001 as vertical source-to-native slices | Cross-layer feedback reveals semantic mistakes sooner than disconnected compiler layers | ✓ Good — every phase surfaced consumer gaps (e.g. 03-02's native execution-document validator) that a layered build would have hidden until integration |
| Use Go 1.24 stdlib for Stage 0 | Fast builds, current spike evidence, simple offline onboarding, and low dependency depth | ✓ Good — zero external deps through 6 phases; `go/ast`/`go/parser` doubled as a structural-test substrate |
| Emit readable C17 through Clang first | Exposes real ABI and optimizer behavior while keeping backend replacement reversible | ✓ Good — real `-O3`/LTO/`restrict` divergences and ASan findings only exist because the backend is real |
| Make the deterministic interpreter the semantic oracle | Prevents C output or optimizer accidents from defining the language | ✓ Good — the five-axis comparator's authority rests on it; it caught two real `--engine=native` bugs in Phase 4 |
| Use affine values with separate copy/drop/share/send/escape abilities | Avoids conflating distinct semantic and cost properties | ✓ Good — abilities derived independently through aggregates and generics; one accepted override (OV-02-01) |
| Infer local borrow ends at CFG last use; declare public origins | Preserves local ergonomics without hiding public ABI relationships | ✓ Good — `loanLivenessFixpoint` became the sole liveness law after a 230,692-comparison shadow run, and M002 Phase 09 closed the interprocedural half (D-03-02) with an independently re-deriving peer |
| Validate compact typed-core facts independently at trust crossings | Gives cheap cache/release evidence without claiming source-to-core proof | ✓ Good — and its known escape is executable, not hypothetical (`escape:coordinated-source-to-core-false-claim`) |
| Keep syntax provisional but canonical | Enables real parser/formatter evidence without freezing decorative choices early | ✓ Good — losslessness and idempotence held through generated programs and malformed AI edits |
| Treat structured diagnostics and evidence as a versioned product API | AI effectiveness depends more on precise verifier feedback than exotic syntax | ✓ Good — the `/0`→`/1` schema bump landed across 12 literal sites with `/0` bytes provably frozen; `cmd/lang-repair` repairs 5 defect classes through the JSON protocol alone |
| Defer `OpCall` and interprocedural equivalence out of M001 | Landing a new `OperationKind` at six dispatch sites alongside alias facts, sanitizers, the reducer, and the QLT-01 registry matched the fingerprint of the failures that cost Phases 2-4 extra remediation rounds | ✓ Good — M002 spent 7 phases and 61 plans on exactly that one addition, which vindicates the deferral; `OpCall` is now real at all six sites |
| Ratify D-12-43: accept that the decisive wrong-slot value-divergence control is unconstructible at this language maturity | The hazard is independently policed at compile time by D-12-37, and the current grammar never lets a match arm's result expose raw payload bytes — so a wrong-slot write is invisible to every comparator axis by construction, not by weak testing. Escalated as a defect in the criterion rather than a downgraded assertion (D-12-41/D-11-36 precedent) | ⚠️ Revisit when the grammar widens to expose bound payload content, or when a payload-value-aware harness observation channel exists; ratified at a blocking-human checkpoint on 2026-09-13, no owning phase named |
| Cut M002 phases where a *gate* becomes meaningful, not where implementation could parallelize | A new `OperationKind` at six dispatch sites has no safe partial state; the useful boundary is "can this claim now be independently checked?" | ✓ Good — all 7 phases verified passed, and each gate caught something (Phase 09's requirement-vs-traceability defect, Phase 10's fixture escape hatch, Phase 12's blocker) |
| Split OWN-05 into OWN-05a/OWN-05b (D-09-37) rather than flip one row Complete on two of three derivers | A single-row overclaim is exactly the requirement-vs-code failure the debt registers exist to catch | ✓ Good — the split is why the milestone audit found zero orphaned or overclaimed requirements across three independent sources |
| Eliminate Phase 11's two human-judgment items with tests rather than adjudicate them | A debt note is read once; a test runs on every CI invocation | ✓ Good — `TestSeedEntryHazardIsReal` and `session_admission_divergence_test.go` both fail in *both* directions, catching a silently resolved divergence as well as a new one |
| Ratify DX-06 and DX-07 as honest partials rather than downgrade the criteria | Both are blocked by one root cause — `sameType(ReturnType, Parameter.Type)` at function admission — and both close automatically when it lifts; restating the criterion to match what shipped would hide a real language limit | ⚠️ Revisit in M003 — the blame resolver and its exhaustiveness guard are already built and waiting on the type-system widening |
| Reject the integration checker's requirement-satisfaction column while adopting its structural findings | It graded requirements from wiring, and wiring is exactly what a structurally unreachable defect class still has | ✓ Good — caught two would-be false greens (DX-06, DX-07) that contradicted both Phase 13's own verification and a `grep` of the tree |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Move disproved active requirements to Out of Scope with evidence.
2. Move verified requirements to Validated with a phase reference.
3. Add requirements exposed by reduced counterexamples.
4. Record decisions and their measured outcomes.
5. Check that the description and core value have not drifted.

**After each milestone:**
1. Review all sections against shipped behavior.
2. Reconfirm the core value.
3. Audit deferred scope and preservation seams.
4. Update context with adopter, performance, and correctness evidence.

---
*Last updated: 2026-09-18 after Phase 14 (Evidence Instrument and Honest Scoping) completed*

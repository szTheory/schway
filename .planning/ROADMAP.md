# Roadmap: Codename Lang

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1-6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07-13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- 🚧 **M003 — Computation and Honest Instruments** — Phases 14-20 (active)
- ◷ **M004 — Native Emission Ownership and Resource Discharge** — planned after M003; Phase 21 owns D-16-11 through D-16-13

## Overview

M001 proved one meaning survives lowering. M002 proved it survives a function
boundary. M003 proves it survives **computation** — the first value Lang creates
rather than moves — on instruments that cannot report green for work that is
merely wired.

Phase numbering **continues** from M002, which ended at Phase 13. M003 runs
Phases 14-20.

The phase structure below is **not re-derived here**. It is the ratified output
of `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md`, an arbitration pass that
resolved five inter-document contradictions against the tree (`908bac3`) and
spot-checked eight high-consequence claims — all eight verified, three stronger
than originally reported. Where that document disagrees with the six source
research documents, it wins. Its load-bearing correction is C3:

> **`match` is a whole-function-body form whose scrutinee must be the function's
> own parameter, and a linear body has no branch form at all.** Branching on a
> computed value therefore does not exist, is not reachable by desugaring, and
> is not cheap.

That single fact refutes ROADMAP-STRATEGY's "`match` already *is* `if`",
understates CONTROL-FLOW-ARITHMETIC's `if` cost by roughly a phase, and breaks
the arithmetic → comparison → `if` → D-12-43 chain three documents relied on.
It is why Phase 18 exists as its own spike-gated phase rather than as a
desugaring task, and why comparison operators and `Bool` are out of M003
entirely.

**Sizing.** Seven phases, ~48-57 plans. This is an M002-scale milestone and is
labelled as one. ROADMAP-STRATEGY's "40-45 plans, explicitly smaller than M002"
is not defensible against the measured emitter surface (869 emitter body lines
plus a 258-line branch-only helper) and is rejected.

## The Governing Gate — Constructibility Precondition

**This gate fires before the work, and it applies to every phase 14-20.**

> No requirement may be admitted into a phase unless a `.lang` fixture that
> exercises it end to end is **written first** and checked in as a *refused
> frontier fixture* with its refusing diagnostic pinned by a test. **The phase's
> own gate is that the pinned diagnostic moved.** A requirement whose fixture
> cannot be written is not deferred and not descoped — it is **refused at the
> gate**, before a plan exists.

Why this one and not a WIP limit, a debt-register schema, or Nyquist-as-a-gate:
none of those would have caught DX-06, D-13-10a, or D-12-43, because all three
were *built correctly and shipped unreachable*. Unreachability is discovered
after the work every single time this project has hit it. This gate converts
"is this claim constructible?" from a post-hoc audit finding into an entry
condition, and it is mechanically checkable — every requirement row cites a
fixture path and a pinned diagnostic code, and a test asserts the fixture exists
and currently produces that code.

Applied on day one it would have refused comparison operators, because no
fixture can be written that *uses* a `Bool` you cannot branch on. That would
have been the fifth instance of this project's signature failure mode.

**Phase-planning consequence.** Every phase 14-20 opens with a fixture-first
plan: check in that phase's refused frontier fixtures with pinned diagnostics
*before* any production code in the phase is planned. Every phase closes by
asserting the pinned diagnostics moved. This is a standing discipline, not a
Phase 14 deliverable.

## Pre-Phase Spikes

Both follow the existing `.planning/spikes/` discipline (falsifiable question,
named competing mechanisms, an oracle that does not reuse the mechanism under
test, at least one injected defect, **stop when the gate is answered**). Neither
is production code and neither carries a requirement.

| Spike | Question it must answer | Blocks | Runs |
|-------|-------------------------|--------|------|
| **S-010 — Does a loan crossing a branch point force a redesign?** Hand-write the fixture: a `borrow` binding created *before* a branch, live inside **exactly one** arm. Drive it through `loanLivenessFixpoint` (`check.go:2711-2790`) and `materializeLoanEndpoints` (`check.go:2794`). | If `loanLivenessFixpoint` accepts it and `materializeLoanEndpoints` classifies its endpoint **without a new `Kind`**, branching is an extension and Phase 18 proceeds. If either needs a **third `LoanEndpoint` kind** or a **per-arm ownership-state merge**, the falsifiable trigger has fired. | **Planning of Phase 18** — hard entry gate. Phase 18 may not be planned until S-010 has answered. | Before Phase 18 is planned; alongside Phases 14-15. |
| **S-009 — Fresh-agent authoring probe.** A fresh agent, with no development-session context, authors the frontier fixture from the JSON diagnostic/evidence protocol alone. Then re-run with the diagnostic **prose** scrambled to lorem ipsum, using M001 Phase 6's existing falsifiability control, to separate "the protocol carried it" from "the English carried it". | Can an agent that is not *this* agent author a small Lang function from the structured protocol alone? This is the project's central untested premise: the AI-authoring thesis has only ever been tested against the development agent. | Nothing hard. Informs Phase 14's DX work and, if negative, tilts later milestones toward corpus and examples. | **Early — alongside Phase 14**, before Phase 17 is planned. |

### S-010's contingency — recorded so it survives a context reset

**If S-010 comes back dirty** (a third `LoanEndpoint` kind or a per-arm
ownership-state merge is required), then, without further deliberation:

1. **Phase 18 is CUT.** CTL-01, CTL-02 and CTL-03 move to **M004**, bundled with
   loops — they are the same loan-across-control-flow problem, and M004 is where
   that problem was already going to be paid for.

2. **Phase 18 is REPLACED by `OpBinary` + arithmetic under optimization**
   (Alt-3 in ADVERSARIAL-SYNTHESIS): binary operators over the fixed-width
   unsigned type introduced in Phase 19, an overflow law, C promotion
   discipline, and a non-importing structural C scan. Arithmetic is
   ownership-inert (`loanLivenessBound = 4 × blocks × (loans+1)` has no term
   arithmetic moves), so the trigger probability there is ~15%, not ~60%.
   Note the ordering flip: the replacement phase *depends on* Phase 19's
   literals, so Phases 19 and 18 swap positions and the arithmetic phase
   executes last of the two.

3. **New requirement IDs are amended into `REQUIREMENTS.md`** for the
   replacement phase (`ARI-01`..) in the same pass that moves CTL-01..03 to the
   Deferred table. Coverage is re-validated at 33 rows; it does not silently
   drop to 30.

4. **Comparison operators and `Bool` stay out regardless.** They are inert
   without a branch on a computed value, and shipping them would be the fifth
   instance of the failure mode this milestone exists to retire.

5. **The milestone's honest headline becomes** "a program can name and compute a
   number" — less exciting, still the first computed value, and not a
   `tech_debt` closeout.

**If S-010 comes back clean**, Phase 18 proceeds as written and arithmetic stays
in M004. It does **not** get added to M003 as a P21/P22 — that would make an
8-phase milestone that should be split rather than crammed.

## Phases

- [x] **Phase 14: Evidence Instrument and Honest Scoping** - The instruments stop reporting green for work that is merely wired.
- [x] **Phase 15: Event Identity (`lang.execution/2`)** - Two activations of the same callee through a shared-leaf diamond become distinguishable. (completed 2026-09-19)
- [x] **Phase 16: Branch/Match Emitter Port** - One emission law lowers every admissible program; three emitters die, three are formally cut. (completed 2026-09-21)
- [x] **Phase 17: Return Type ≠ Parameter Type** - A function may return a type it was not given. (completed 2026-09-22)
- [x] **Phase 18: Branch on a Computed Value** - A branch discriminates a value the function computed. **HARD-GATED on spike S-010.** (completed 2026-09-24)
- [x] **Phase 19: Numeric Literals and `OpConst`** - Lang can name a value it was not given. (completed 2026-09-24)
- [ ] **Phase 20: Nyquist, D-13-34, and the Frontier Fixture** - The measuring corpus reconciles against the new surface and the refusal frontier is pinned as moved.

## Phase Details

**Standing refs — apply to every phase 14-20, read before planning any of them:**

- `.planning/STANDING-VERDICTS.md` — dependency verdicts (the zero-external-dep
  record is a hard constraint), design anti-features, process anti-patterns, and
  load-bearing facts about the six dispatch sites and why `-flto` is
  load-bearing. Reopen an entry only on new evidence.

- `.planning/LANGUAGE-MATURITY.md` — what the language can actually express
  today. Assurance stack ~70-75% built; language surface ~5-10%. Guards against
  planning as if arithmetic, iteration, collections, or a branch on a computed
  value already exist. They do not. Do **not** read `wiki/example-tour.md` as a
  description of the language.

- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` — **authoritative**. Its
  contradiction resolutions override the six source documents wherever they
  conflict.

Note: the six per-dimension documents under `.planning/research/M003/` are
milestone-scoped and regenerated by the next `/gsd-new-milestone`. The three
refs above are the durable ones.

---

### Phase 14: Evidence Instrument and Honest Scoping

**UI hint**: no
**Goal**: Every shipped claim is graded at EXERCISED or above, or names itself
unreachable with an unblocking trigger — so that no instrument in this project
can report green for something that is merely wired.
**Depends on**: M002 Phase 13 (complete)
**Requirements**: EVD-01, EVD-02, EVD-03, EVD-04, EVD-05, EVD-06, EVD-07,
EVD-08, DX-08, DX-09, PRC-01
**Success Criteria** (what must be TRUE):

  1. A CI lint over `.planning/**` reports **zero** `go test -run` patterns that
     resolve to zero tests and **zero** command cells elided with `…`; its own
     non-inertness control — a seeded dead pattern — makes the lint fail. Three
     such patterns exist today and exit 0, including one in a phase certified
     `nyquist_compliant: true`. *(EVD-01)*

  2. Every requirement and success-criterion row carries a grade from the closed
     vocabulary `DEFINED | WIRED | REACHABLE | EXERCISED | MUTATION-KILLED`; a
     grade outside the vocabulary fails a parser test, and a row below EXERCISED
     cannot be recorded as satisfied. *(EVD-02)*

  3. No claim or exclusion outlives its justification: `.planning/UNREACHABLE-CLAIMS.md`
     exists and holds each built-but-structurally-unreachable claim with its
     unblocking trigger; a guard fails when a cited gate has already closed; and
     the mutation axis-movement law has **exactly one** implementation with
     **zero** per-row exclusions (`Phase5AssertMutationMovesAnAxis` deleted, the
     one `Subjected: true` fixture-path exclusion removed, `retained_pointer`
     kept as a declared escape rather than a silent skip). *(EVD-03, EVD-04,
     EVD-05)*

  4. Self-describing documents are machine-checked and currently true:
     `LANGUAGE-MATURITY.md`'s counts are asserted by a test over its own
     re-verify greps (it says 32 guards; there are 26, and its corpus figures
     are stale in both directions); `qlt02_budget_manifest.json` carries an
     `observed` suite wall-clock row with the 192.7 s baseline recorded (docs
     claim ~60 s — a 3.2× error in an input to this project's own
     feedback-budget reasoning); PROJECT.md's DX-06 claim and NAT-07 `-flto`
     bullet state what the evidence supports; the `-flto` multi-function
     inertness has a debt row with an owning phase; and a well-formedness gate
     **fails** when any debt row is recorded without an owning phase.
     *(EVD-06, EVD-07, EVD-08, PRC-01)*

  5. Three structurally distinct defective programs mint three **distinct**
     result IDs and distinct diagnostic-ID sets — the reproduced `if`-spiral
     (`if v { v } else { v }` / `if v { v }` / `if v`) is pinned as the negative
     control — and `lang-repair --json` never returns `unrepairable` with an
     empty `diagnosis`. *(DX-08, DX-09)*

**Naming note**: REQUIREMENTS.md's **DX-08** and **DX-09** are
ADVERSARIAL-SYNTHESIS's **DX-10** (`diagnostic_distinctness`) and **DX-12**
(non-empty `diagnosis`) renumbered. The synthesis's *own* DX-08/DX-09 — the
`not_in_language` capability manifest and `surface.not_in_language` — are
**deferred**, per its C5 ruling. Read the requirement text, not the number, when
cross-referencing the research.
**Riskiest assumption**: that grading is mechanizable rather than a judgment
call, i.e. that "EXERCISED" can be decided by a test rather than by a reviewer.
**Gate that catches it being wrong**: EVD-02's parser test runs over the
existing M001/M002 requirement corpus before any M003 row is graded. If the
vocabulary cannot classify already-shipped rows without argument, the
vocabulary is wrong while it is still a data-shape change.
**Sizing note**: the lint is ~40 lines, offline and deterministic — the single
highest-ROI item across all six research documents. The cost is elsewhere: the
axis-movement law has two coexisting implementations, and the PROJECT.md
corrections touch three documents that each state the same wrong claim.
**Parallel**: S-009 and S-010 run alongside.
**Plans**: 13 plans (10 planned 2026-09-17; 3 gap-closure plans added
2026-09-18 after `14-VERIFICATION.md` scored 10/11 with one partial — the
above-estimate count reflects eleven requirements decomposing into
single-concern units, each with its own non-inertness proof, rather than
scope growth)

Plans:
**Wave 1**

- [x] 14-01-PLAN.md — tracer: the groundedness lint end-to-end, with its frontier pinned as an exact literal and three seeded-fault non-inertness proofs (EVD-01)
- [x] 14-02-PLAN.md — attach the discarded recovery extent as an identity-bearing cause; distinctness corpus, gate, and frozen pre-fix collision control (DX-08)
- [x] 14-03-PLAN.md — an `unrepairable` verdict explains itself: additive decline fields, closed decline vocabulary, guard and non-inertness proof (DX-09)
- [x] 14-04-PLAN.md — close the owning-phase vocabulary in the debt-register law, migrate twelve registers, register the `-flto` inertness row (PRC-01, EVD-07)
- [x] 14-05-PLAN.md — machine-check the maturity document's counts by independent re-derivation; record a cold suite wall-clock as an observed manifest row (EVD-06, EVD-08)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 14-06-PLAN.md — lint expansion: scope by illocutionary role, executed grep groundedness, per-branch detection, corpus floors, index accuracy control (EVD-01)
- [x] 14-07-PLAN.md — closed witness grammar, six executed probes, module-wide suppression enumerator, generated unreachable-claims view (EVD-03, EVD-04)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 14-08-PLAN.md — one axis-movement law, zero per-row exclusions, the superseded marker guard and its stale markers deleted (EVD-05)
- [x] 14-09-PLAN.md — closed grade vocabulary with a mechanically derived ceiling; fourteen verification maps migrated; the freeform status column retired (EVD-02)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 14-10-PLAN.md — reconcile every finding outside the archives under checked verdict obligations; generated reconciliation view; frontier emptied (EVD-03, EVD-01)

**Wave 5** *(gap closure — added 2026-09-18 from `14-VERIFICATION.md`; blocked on Wave 4 completion)*

- [x] 14-11-PLAN.md — a run record that cannot lie about completing: per-pair completion witness, honest 480s budget with a measured 75% margin assertion, anchored producer/consumer name contract, self-citation refused (EVD-02; closes WR-01, WR-02)
- [x] 14-13-PLAN.md — the `//go:build` suppression surface stops being inert: declared allowlist, enumeration ahead of the host's own constraint filter, a fourth seeded fault (EVD-04; closes WR-03)

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 14-12-PLAN.md — remove `14-VALIDATION.md`'s grade-bar exemption and confirm the corpus-wide bar with a complete run record and measured margin; row-scoped narrowing that names an owner; D-14-121 leaves UNOWNED (EVD-02, PRC-01)

---

### Phase 15: Event Identity (`lang.execution/2`)

**Goal**: Two activations of the same callee through a shared-leaf diamond are
distinguishable, and the causal edge between caller and callee is **observed**
rather than inferred.
**Depends on**: Phase 14 (the grading vocabulary and the lint must exist before
a new schema version mints new claims)
**Requirements**: OBS-01, OBS-02, OBS-03, OBS-04, NAT-10
**Success Criteria** (what must be TRUE):

  1. `testdata/phase11/multi_function_diamond_call.lang` — which today emits a
     byte-identical duplicate event ID on the interpreter and fails
     `native.invalid_execution` on the native path — runs on **all four tiers**
     (interpreter, `-O0`, `-O3`, `-O3 -flto`) and `Phase5CompareEngines` agrees.
     The `DiamondSharedLeaf` subtest's inverted assertion is **flipped, not
     deleted**, per its own written instruction, and fails if the collision
     returns. *(OBS-01, NAT-10)*

  2. An `OpCall` emits an observable event naming the caller and callee
     invocation, present in the execution document on every engine — today
     `OpCall` emits no event at all, so the causal edge is inferred. A fixture
     whose call is removed loses exactly that event. *(OBS-02)*

  3. A non-importing peer re-derives the admissible invocation set over its own
     traversal (sound because the call graph is a guaranteed DAG), and a seeded
     mismatch fails the differential **in both directions** — a new divergence
     and a silently-resolved one. *(OBS-03)*

  4. `/0` and `/1` execution-document golden bytes are unchanged; the required
     `invocation` field exists only under `lang.execution/2`; uniqueness is
     `(invocation, id)` rather than `id`. A test asserts the frozen `/0` and
     `/1` byte sets. *(OBS-04)*

  5. The static path table's node ceiling is **measured** against
     `testdata/phase07/deep_diamond_acyclic.lang` (13 functions) rather than
     guessed, and exceeding it fails closed with a named diagnostic instead of
     unfolding.

**Riskiest assumption**: EMISSION's static-path-table C mechanism for
`invocation` identity, and its 4096-node ceiling — both unmeasured and
self-rated MEDIUM in the research. The `2^depth` unfolding bound is real
arithmetic; whether it bites on the 13-function diamond is untested.
**Gate that catches it being wrong**: measure it in the phase's discuss gate,
before the emission mechanism is chosen. Criterion 5 is that measurement.
**Sizing note**: ~300-400 lines across `interp.go` frame threading,
`emitProgram`'s event derivation, a peer re-derivation and comparator field
routing. Ordered **before** the emitter port deliberately: port the branch arms
first and their event emission gets written twice, which is
STANDING-VERDICTS' own "two coexisting laws is a defect with a delayed fuse".
**Note on the dependency's true strength**: the port's authorizing corpus does
**not** need a re-invoked callee — a `main → callee-with-a-match` chain needs no
identity fix. The dependency caps corpus richness; it does not gate the port.
The ordering rests on the write-it-twice argument, not on a hard block.
**Closes**: D-11-51, D-12-21.
**Plans**: 20 plans

Plans:
**Wave 1**

- [x] 15-01-PLAN.md — confirm and trace the `/2` wire grammar through strict admission while freezing `/0` and `/1` and pinning the pre-fix diamond frontier

**Wave 2** *(blocked on Wave 1 completion; plans may execute in parallel)*

- [x] 15-02-PLAN.md — thread interpreter occurrence identity and emit caller-owned preorder call edges with projection-only removal semantics
- [x] 15-03-PLAN.md — independently re-derive invocation membership and validate exact observed causal structure in a non-importing peer
- [x] 15-04-PLAN.md — preflight native invocation expansion with the measured 61-node control and exact 4096/4097 non-inert boundary

**Wave 3** *(blocked on the named Wave 2 dependencies)*

- [x] 15-05-PLAN.md — emit static parent-indexed invocation tables and `/2` call evidence in readable C17 without moving legacy writers
- [x] 15-06-PLAN.md — route both new comparator fields, require peer validation, and prove producer/peer independence in both fault directions

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 15-07-PLAN.md — flip DiamondSharedLeaf into a peer-validated four-tier positive gate, retain the collision negative, and close inherited debt honestly

**Wave 5** *(gap closure; blocked on Plan 01 completion)*

- [x] 15-08-PLAN.md — automate the Schema 2 JSON-to-ToolError admission seam and publish Plan 01 coverage metadata (G-15-1)

**Wave 6** *(gap closure; blocked on Plans 06-08 completion)*

- [x] 15-09-PLAN.md — preserve legacy wrapper behavior, replace stale CI provenance with a durable aggregate, and publish peer/diamond coverage (G-15-11, G-15-12)

---

### Phase 16: Branch/Match Emitter Port

**Goal**: One emission law lowers every admissible program, instead of two laws
split by a function-count guard.
**Depends on**: Phase 15
**Requirements**: NAT-08, NAT-09
**Success Criteria** (what must be TRUE):

  1. `grep -c 'func emitMatch\|func emitBranch(\|func emitLinear('` returns
     **0**, and the deletion lands in the **same commit** that flips dispatch —
     Phase 09's pattern: three pre-deletion gates, a blocking
     `checkpoint:decision` plan, same-commit cutover. *(NAT-08)*

  2. `TestN1ConvergenceDifferential`'s measured table flips: `emitProgram` today
     **refuses** 4 of 5 shapes (`Match`, `Blocks`, `ForeignContract`) and
     differs in bytes on the 5th; after the port it reaches byte-identity on the
     three shapes in scope, and **each** moved golden-C digest — four are pinned
     by `previousPhaseGoldenCDigests` — carries a written justification rather
     than a silent re-baseline. *(NAT-08)*

  3. A recorded amendment, filed under D-10-60's own clause that a
     twice-deferred item *"must either be cut from the milestone explicitly via
     a REQUIREMENTS.md amendment, or becomes automatically never-cut"*, names
     `emitLinearForeign`, `emitLinearBorrowedByPointer` and `…Plain` as **cut**,
     names **M004** as their landing milestone, and states the `-flto`
     inertness consequence. Each is refiled as an M004 debt row with an owning
     phase. *(NAT-09)*

  4. No admissible `.lang` fixture in the corpus is routed by a function-count
     guard: `cgen.Emit`/`EmitNative`'s `len(program.Functions) != 1` fork is
     gone, and `emitProgram`'s hardcoded `"live_resources":[]` string literal is
     a derivation.

**Riskiest assumption**: that the port is ~600-800 lines and 8-10 plans. The
line counts are measured (869 emitter body + 258 `emitBranchOperations`); the
plan count is judgment, raised from EMISSION's 4-6 on the strength of
`emitBranchOperations`'s payload-arm handling plus four pinned golden digests.
**Gate that catches it being wrong**: the three pre-deletion gates land before
any deletion, so an under-estimate surfaces while both laws still exist and the
cutover has not happened.
**Sizing note**: this is a **port-then-delete**, not a deletion. TYPE-WIDENING's
claim that widening the return type forces the deletion is rejected (C1): a
single-function program with `return ≠ parameter` routes to
`emitLinear`/`emitBranch`, never to `emitProgram`, so widening before the port
means paying the two-type model **twice** and carrying it through ~870 lines of
legacy emitter while simultaneously converging them — the exact fingerprint
PROJECT.md's Key Decisions table blames for M001 Phases 2-4.
**A spike worth running before planning**: whether a callee-side `restrict` on a
`static` function inside one TU needs a caller-side discharge pair. If it does
not, `emitLinearBorrowedByPointer` moves from the cut half to the landed half
and the `-flto` inertness gap is de-fanged. Highest-value optional probe in the
milestone; not a hard gate.
**Closes**: D-11-02, D-12-36, D-11-27; honours D-10-60.
**Plans**: 9 plans

**Wave 1**

- [x] 16-01-PLAN.md — prove the ordinary-linear tracer and derived resource tail through `emitProgram`
- [x] 16-04-PLAN.md — check in the exact-shape `restrict` probe and exhaustive structural refusal matrix

**Wave 2**

- [x] 16-02-PLAN.md — port branch blocks while preserving Phase 15 validation order
- [x] 16-05-PLAN.md — decide the evidence-bounded by-pointer admission or M004 cut

**Wave 3**

- [x] 16-03-PLAN.md — port match payload and defect lowering through checker-owned layout facts

**Wave 4**

- [x] 16-06-PLAN.md — implement the pointer disposition and formally amend/own every M004 cut

**Wave 5**

- [x] 16-07-PLAN.md — establish byte, provenance-ledger, and independent four-tier convergence gates

**Wave 6**

- [x] 16-08-PLAN.md — obtain explicit approval for the one-way atomic authority cut

**Wave 7**

- [x] 16-09-PLAN.md — flip dispatch, delete the three superseded laws in the same commit, and close validation

**Wave 1 — UAT gap closure**

- [x] 16-16-PLAN.md — migrate Phase 11 by-pointer evidence to refusal-first frozen inputs
- [x] 16-19-PLAN.md — synchronize historical witnesses and make budget probes deterministic
- [x] 16-20-PLAN.md — refresh machine-checked language maturity corpus counts

**Wave 2**

- [x] 16-18-PLAN.md — reconcile the exact groundedness frontier and P20 ownership

**Wave 3**

- [x] 16-17-PLAN.md — regenerate the current validation corpus run record and manifest

---

### Phase 17: Return Type ≠ Parameter Type

**Goal**: A function may declare a return type different from its parameter
type, and a program relying on that checks, runs, and lowers.
**Depends on**: Phase 16 (so the two-type model is paid once, in one surviving
emitter family)
**Requirements**: TYP-01, TYP-02, TYP-03, TYP-04, TYP-05
**Success Criteria** (what must be TRUE):

  1. A checked-in `.lang` program whose function declares `-> T` with a
     different parameter type checks, interprets, lowers, and agrees across
     interpreter / `-O0` / `-O3` / `-O3 -flto`. *(TYP-01)*

  2. `check.call_argument_type_mismatch` fires **from a `.lang` fixture**, not
     from a seam-level unit test — the distinction that separates REACHABLE from
     EXERCISED under Phase 14's vocabulary. *(TYP-02)*

  3. `check.call_return_type_unrepresentable` fires from a `.lang` fixture.
     *(TYP-03)*

  4. Drop obligations and freshness for the returned value are derived from the
     **return type's own abilities**, independently in `check`, `corevalidate`
     and `originvalidate` — three derivations, not one inherited from the
     parameter — and a seeded divergence fails all three. *(TYP-04)*

  5. `lang-repair` reaches `repaired` on `use_matching_argument` against a
     **sealed held-out** fixture, closing DX-07 / D-13-10a. In the same phase,
     D-13-02b is ratified **permanent** with its reopening condition corrected
     to name a *user-declared contract field the declaring function's own
     admission cannot verify* — i.e. separate compilation, M006. *(TYP-05)*

**Explicitly does NOT close DX-06.** The entire blame subsystem —
`resolveBlame`, `resolveCycleBlame`, `classifyDeclaredCause` — is test-only dead
code with zero production call sites, and nothing ever sets
`blameFact.Violated = true` outside `check_blame_test.go`. Lifting the
single-type invariant is **necessary-not-sufficient**. PROJECT.md, the M002
audit's Cluster C, and PHASE-13-DEBT all previously stated otherwise; Phase 14
corrects them. Do not restate the "closes automatically" claim a third time.
**Riskiest assumption**: that the `Drops`/`Fresh` split can be derived three
times independently without the three derivations converging on a shared helper
— which would make "independently re-derived" a wiring claim.
**Gate that catches it being wrong**: the existing transitive-dependency import
guards, plus a seeded divergence that must fail all three peers separately.
**Sizing note**: touches three `sameType` heads, the match-arm alternative set,
`resolveCallBinding`'s single-constructor comparison, N type facts at three
minting sites (`functionID:type:N`), the `Drops`/`Fresh` split three times,
`reduce.go:769`'s silent TypeID fallback, and two C type names in the one
surviving emitter family.
**Parallel**: S-009's finding should be in hand before planning, since it bears
on how much `lang-repair` surface is worth extending.
**Plans**: 8-10 (TBD at `/gsd-plan-phase 17`)

---

### Phase 18: Branch on a Computed Value

> **HARD-GATED on spike S-010.** This phase may not be planned until S-010 has
> answered. If S-010 comes back dirty, this phase is **cut** and replaced per
> the contingency recorded under `## Pre-Phase Spikes` above.

**Goal**: A branch can discriminate a value the function **computed**, not only
the function's own parameter — by generalizing `match` to a terminal form of a
linear body over any in-scope place of a `data` type.
**Depends on**: Phase 17, and S-010
**Requirements**: CTL-01, CTL-02, CTL-03
**Success Criteria** (what must be TRUE):

  1. A `.lang` fixture binds a value with `let` and then matches on **that
     place** — not on `function.Parameter.Name`, which `check.go:258` requires
     today — and it checks, interprets, lowers, and agrees across all five
     comparator axes. *(CTL-01)*

  2. A fixture that calls a `Result`-returning callee and matches on the
     returned result agrees across all five axes. This is the first time M002's
     `Result`-payload work is usable *across a call*. *(CTL-02)*

  3. A match arm returns a **destructured payload place**, and a seeded
     wrong-slot write is observed on `axis:terminal-outcome` — **D-12-43
     constructed, not ratified as unconstructible**. Note this closes D-12-43
     via the generalized scrutinee, not via arithmetic: the synthesis refutes
     the "arithmetic in an arm makes D-12-43 constructible" inference, because
     D-12-43's reopening condition is exposing *payload bytes* on an arm's
     return path. *(CTL-03)*

  4. S-010's fixture is re-run against **production** code, not the spike
     workbench: a `borrow` created before the branch and live in exactly one arm
     is classified by `materializeLoanEndpoints` with no new `LoanEndpoint`
     kind, and `loanLivenessFixpoint` terminates within
     `loanLivenessBound = 4 × blocks × (loans+1)`.

**Riskiest assumption**: that generalizing the scrutinee is an *extension* of
loan-endpoint materialization rather than a *redesign*. Probability the
falsifiable trigger fires on this phase: **~60%** — far above arithmetic's
~15%. Two things true today stop being true the moment a `let` precedes a
branch: (a) `analyzeArmBody`'s soundness argument (`check.go:2872-2876`,
*"mutually exclusive control flow — the two arms' places never collide"*) rests
on each arm starting from the parameter alone; with a shared straight-line
prefix, siblings share live places; (b) `materializeLoanEndpoints`'s
`point`/`edge` dichotomy has no classification for a loan born in the prefix and
last-used in one arm but not the other.
**Gate that catches it being wrong**: S-010, run **before this phase is
planned**, so the finding arrives at a planning gate rather than mid-phase. The
single most likely path to another `tech_debt` closeout is a phase that attempts
this mid-flight, discovers it, descopes, and ships an inert `Bool`. That path is
closed by construction here.
**Sizing note**: a new AST shape (a branch in a linear body's terminal
position), a `checkBranch` re-architecture so the entry block can carry a
straight-line prefix, and an ownership question that does not exist today. This
is Austral Rule 3's territory arriving a full milestone before anyone budgeted
for it.
**Plans**: 8

Plans:

- [x] 18-01-PLAN.md — Pin computed-place, Result-call, payload, and production-loan source frontiers
- [x] 18-02-PLAN.md — Trace computed terminal match through source, peers, interpreter, and native emission
- [x] 18-03-PLAN.md — Independently validate computed places and payload origins
- [x] 18-04-PLAN.md — Match a Result-returning callee value across all five axes
- [x] 18-05-PLAN.md — Return destructured payload and kill the wrong-slot mutation
- [x] 18-06-PLAN.md — Prove bounded production borrow-across-branch liveness
- [x] 18-07-PLAN.md — Add peer, comparator, and mutation anti-vacuity controls
- [x] 18-08-PLAN.md — Measure verification cost and justify focused recurring CI

---

### Phase 19: Numeric Literals and `OpConst`

**Goal**: Lang can name a value it was not given — the first value the language
creates rather than moves.
**Depends on**: Phase 17. **Not** dependent on Phase 18; if Phase 18 is cut,
this phase is unaffected and executes in its place.
**Requirements**: VAL-01, VAL-02, VAL-03
**Success Criteria** (what must be TRUE):

  1. A numeric literal parses, formats idempotently (losslessness preserved),
     checks, interprets, and lowers; `lang run` on both engines produces the
     named value. One fixed-width unsigned type — no signed integers, no
     multiple widths, no floats, no numeric tower, which makes signed-overflow
     UB *unconstructible* rather than merely avoided. *(VAL-01)*

  2. `TestAllOperationKindsHandledAtEverySite` is green with `OpConst`
     registered, and **both** exhaustive-dispatch controls fail when any one of
     the six dispatch sites is removed. *(VAL-02)*

  3. A literal-bearing fixture agrees across interpreter, `-O0`, `-O3`, and
     `-O3 -flto`. *(VAL-03)*

  4. No serialized-execution golden moves except by reviewed intent: D-12-18's
     byte-identical scalar projection of `interp.value` (`{tag, payload string}`)
     survives the value-domain widening, or every moved golden is justified in
     writing.

**Riskiest assumption**: that a numeric domain can preserve D-12-18's
byte-identical scalar projection. If it cannot, every serialized-execution
golden in the corpus moves — painful, but still an extension, not a redesign.
This is the entire residual 15% behind the "arithmetic is ownership-inert"
finding.
**Gate that catches it being wrong**: criterion 4 is checked in the first
production plan — widen the interpreter value domain and re-run the golden
corpus before `OpConst` reaches any other dispatch site. The preceding fixture-
only plan pins the current refused frontier and its diagnostics.
**Sizing note**: 8-10 coordinated edits at six dispatch sites (the measured
project base rate for a new `OperationKind`) plus a new interpreter value
domain. Registering `OpConst` in `AllOperationKinds()` deliberately breaks all
six exhaustive switches at once — a forcing function, per Phase 07's precedent.
**Plans**: 7, in seven dependency-ordered waves. Plan 01 pins the refused
literal frontier; Plan 02 widens interpreter values and replays scalar goldens;
Plan 03 adds lossless syntax; Plan 04 checks and lowers U64; Plan 05 updates
independent core/path/origin peers; Plan 06 executes constants in the interpreter
and native C emitter; Plan 07 registers all consumers and proves four-tier
agreement. See `.planning/phases/19-numeric-literals-and-opconst/19-01-PLAN.md`
through `19-07-PLAN.md`.

---

### Phase 20: Nyquist, D-13-34, and the Frontier Fixture

**Goal**: The measuring corpus reconciles against the surface M003 actually
built, and the refusal frontier is pinned as having moved.
**Depends on**: Phases 14-19 (all). Nyquist reconciliation may **not** start
before Phase 14's lint exists — the lint must do the finding.
**Requirements**: QLT-10, QLT-11, QLT-12, PRC-02
**Success Criteria** (what must be TRUE):

  1. Zero VALIDATION files remain `status: draft`. Phases 07, 08 and 11
     reconcile against the post-M003 surface with **Phase 14's lint doing the
     finding** (Phase 11 has six or more `…`-elided, un-runnable command cells;
     Phase 08 has two zero-test `-run` patterns, one naming a mechanism deleted
     in Phase 09). Phases 12 and 13 are frontmatter flips done inline.
     *(QLT-10)*

  2. `examples/checksum.lang` — read a file through foreign C, loop-accumulate a
     byte checksum, print it — is checked in as a **refused** frontier fixture
     with its refusing diagnostic pinned by a test, and that test asserts the
     pinned diagnostic **differs** from the one pinned at M003 open. Each
     milestone must measurably move the refusal forward; landing is end of M005.
     *(QLT-11)*

  3. The enumerated-closure proof is content-addressed rather than re-run every
     commit: an unchanged 112-program closure re-uses its evidence, a seeded
     change forces a re-run, and suite wall-clock drops measurably against Phase
     14's recorded 192.7 s baseline. The subtest is 58.33 s — 30% of total wall
     clock — re-proving a frozen closure, twice under `-race`, on a language
     that cannot add two numbers. *(QLT-12)*

  4. Open, **unowned** debt items number **≤ 5**, machine-counted by PRC-01's
     well-formedness gate rather than by hand. M002 closed with 10.
     *(PRC-02)*

  5. D-13-34 is re-triggered now that a structurally distinct move/borrow
     program is constructible: M001's `testdata/phase6` held-out pairs are
     replaced with genuinely distinct programs, **or** the item is re-ratified
     with the reason recorded and an owning milestone named. D-06-29's
     anti-overfitting inference is restored for the two voided source classes,
     or explicitly written off.

**Riskiest assumption**: that reconciliation is cheap once the lint exists. It
was "cheap to close" in M002's telling and that was only half true.
**Gate that catches it being wrong**: the lint's finding count is taken at the
*start* of this phase, from Phase 14's instrument, so the size is measured
before plans are written rather than estimated.
**Sizing note**: the milestone's close condition lives here (PRC-02). If the
unowned count is above 5 when this phase opens, the overflow is worked here or
the milestone does not close.
**Plans**: 6 plans

Plans:
- [ ] 20-01-PLAN.md — Pin the refused checksum frontier and its M003-open comparison
- [ ] 20-02-PLAN.md — Reconcile Phases 07, 08, and 11 through the live lint
- [ ] 20-03-PLAN.md — Clear the remaining draft validation records and re-pin the frontier
- [ ] 20-04-PLAN.md — Reuse content-addressed closure artifacts with fresh judgments
- [ ] 20-05-PLAN.md — Derive and enforce the open-unowned debt cap
- [ ] 20-06-PLAN.md — Adjudicate D-13-34 and D-06-29 at a human decision gate

---

## Explicit Non-Goals

Recorded so they are not silently re-added. Each has a reason and, where
applicable, a named landing.

| Not in M003 | Reason |
|---|---|
| **`if` / `else` as surface syntax** | A second control-flow law over a construct that does not yet do the first law's job. Phase 18 generalizes the one law instead. |
| **Comparison operators** | Inert until Phase 18 lands and proves out. A comparison must return something matchable, and `match`'s scrutinee resolution is a `data`-type lookup — so comparison forces a `Bool` into existence. Shipping them earlier repeats DX-06. → M004, first item. |
| **`Bool`** | A `Bool` that cannot be branched on is the DX-06 failure mode for the fourth time. → M004, with comparison. |
| **`OpBinary` / arithmetic operators** | Phase 19's natural successor, not its companion. → M004 — **unless** S-010 comes back negative, in which case they replace Phase 18 and comparison stays out regardless. |
| **Loops, iteration, back edges** | `pathoracle.cfg_back_edge` refuses every CFG cycle *by name*; `EnumeratePaths`'s acyclic DFS, `loanLivenessBound`, `materializeLoanEndpoints`'s `point`/`edge` dichotomy, and `analyzeArmBody`'s soundness argument all fail under a back edge. Iteration does not make the oracle expensive, it makes it **impossible**. → M004, with a pathoracle-successor spike as a hard entry gate. |
| **Arity-N, aggregates** | `OpCall` must gain `SourceIDs []string` — the six-dispatch-site schema shape M002 spent seven phases and 61 plans doing exactly once. Reopens `pathoracle`'s collapsed `Mode` dimension from ≈12 to ≈108 cases, breaks `deriveReturnOrigin`'s hardcoded parameter path and three backward walks, and makes `restrict` soundness depend on a multi-argument disjointness rule that does not exist. → M005. |
| **DX-06 closure / B1 blame** | Needs a user-declared contract field the declaring function's own admission cannot verify; every `FunctionSignature` field today is producer-derived. → M006, with separate compilation. Ratified permanent as D-13-02b with a corrected reopening condition in Phase 17. |
| **Three cut emitter families** (`emitLinearForeign`, `emitLinearBorrowedByPointer`, `…Plain`) | No M003 consumer, and their port opens two designed-not-built subsystems (D-11-11/D-11-12 discharge pairs, D-10-34's multi-frame pad). Formally **cut** under D-10-60's own amendment clause in Phase 16, refiled with M004 landings and the LTO-inertness consequence named — not deferred a third time. |
| **Signed integers, multiple widths, floats, a numeric tower** | One fixed-width unsigned type makes signed-overflow UB unconstructible rather than merely avoided. |
| **`-fwrapv`** | Since GCC 8 it disables `-fsanitize=signed-integer-overflow`, blinding the existing UBSan lane. |
| **Type inference** | Inferred signatures make declared-contract blame incoherent — an anti-feature for a language whose product is attribution. |
| **Separate compilation, modules, a fourth peer, a second comparator** | Unchanged from M002. |
| **Content-addressed digest for event identity** | Categorically wrong here: the two colliding activations are content-identical, so a digest collides *by design*. |
| **DX capability manifest, `surface.not_in_language`, rendered diagnostics, repair provenance** | Contingent and decreasing in value; DX-11 would polish a surface this milestone is about to change. |
| **Any Nyquist reconciliation before Phase 14's lint exists** | The lint must do the finding. |

## Forward Arc (M004-M006)

Provisional, revised at each milestone boundary. Recorded here so the M003
research fan-out is not re-derived later.

| Milestone | Charter | Gate that becomes meaningful |
|---|---|---|
| **M004** | **Iteration** (plus comparison, `Bool`, and `OpBinary`; plus branching if S-010 came back dirty) | Back edges exist. Must first answer what replaces `pathoracle`, which refuses every CFG cycle by name. Under a back edge the set of *executions* is infinite while the set of acyclic path shapes stays finite — so the oracle's claim silently weakens with **no test going red**. Needs bounded unrolling to a declared depth or a loop-summary oracle. A pre-phase spike is **mandatory**, as a hard entry gate. |
| **M005** | **Aggregates and arity-N** | D-12-43 finally becomes constructible via a generalized-scrutinee arm returning a destructured payload — unless Phase 18 already constructs it, in which case M005 inherits a closed item. Arity 2 takes the `pathoracle` case space from ~12 to ~108. `examples/checksum.lang` lands here. |
| **M006** | **Modules and separate compilation** | DX-06's B1 blame becomes real: the first user-declared contract field the declaring function's own admission cannot verify. |

**The standing falsifiable trigger**, carried forward: if a milestone's work
requires a **redesign** — not an extension — of `loanLivenessFixpoint`, the
five-axis comparator, or `corevalidate`'s forward propagation, then the
assurance stack was calibrated against a toy, and the next milestone becomes an
assurance-refactor milestone rather than a feature milestone.

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 14. Evidence Instrument and Honest Scoping | 13/13 | Complete    | 2026-09-18 |
| 15. Event Identity (`lang.execution/2`) | 9/9 | Complete    | 2026-09-19 |
| 16. Branch/Match Emitter Port | 26/26 | Complete    | 2026-09-21 |
| 17. Return Type ≠ Parameter Type | 9/9 | Complete    | 2026-09-22 |
| 18. Branch on a Computed Value | 8/8 | Complete    | 2026-09-24 |
| 19. Numeric Literals and `OpConst` | 7/7 | Complete    | 2026-09-24 |
| 20. Nyquist, D-13-34, and the Frontier Fixture | 0/? | Not started | - |

## Requirement Coverage

**33 of 33 M003 requirements mapped to exactly one phase. No orphans, no
duplicates.** Full traceability table lives in `REQUIREMENTS.md`.

| Phase | Requirements | Count |
|-------|--------------|-------|
| 14 | EVD-01, EVD-02, EVD-03, EVD-04, EVD-05, EVD-06, EVD-07, EVD-08, DX-08, DX-09, PRC-01 | 11 |
| 15 | OBS-01, OBS-02, OBS-03, OBS-04, NAT-10 | 5 |
| 16 | NAT-08, NAT-09 | 2 |
| 17 | TYP-01, TYP-02, TYP-03, TYP-04, TYP-05 | 5 |
| 18 | CTL-01, CTL-02, CTL-03 | 3 |
| 19 | VAL-01, VAL-02, VAL-03 | 3 |
| 20 | QLT-10, QLT-11, QLT-12, PRC-02 | 4 |

### Deviations from the natural category mapping, justified

- **NAT-10 → Phase 15, not Phase 16.** NAT-10 reads "a re-invoking
  multi-function fixture is compared across interpreter, `-O0`, `-O3`, and
  `-O3 -flto`" — which is *verbatim* Phase 15's ratified gate
  (`multi_function_diamond_call.lang` passing `Phase5CompareEngines` across all
  four tiers). Mapping it to Phase 16 would mean Phase 15 satisfies a
  requirement it does not own, and Phase 16 claims credit for an already-green
  row. That is precisely the "green because something is wired" pattern this
  milestone exists to retire, so the requirement is filed where its gate fires.
  Phase 16 keeps NAT-08 and NAT-09, which are genuinely its own.

- **PRC-01 → Phase 14, PRC-02 → Phase 20.** PRC-01 (every debt row names an
  owning phase, enforced by a well-formedness gate) is an *instrument* and must
  exist before debt accrues, so it belongs with the evidence phase. PRC-02
  (≤ 5 open unowned items) is a *close condition* and can only be adjudicated at
  the end. They are the same discipline measured at two different times, which
  is why they are two rows.

- **DX-08, DX-09 → Phase 14.** Both are defects in a shipped versioned API
  measurable today, ~1 plan each, and both are instances of the milestone's
  central theme (a protocol that cannot distinguish three different programs is
  an instrument reporting green). They are plans inside the evidence phase, not
  a DX phase.

### Coverage under the S-010 contingency

If Phase 18 is cut, CTL-01/02/03 move to M004's Deferred table and are replaced
by new `ARI-NN` rows for the arithmetic phase, amended into `REQUIREMENTS.md` in
the same pass. Coverage is re-validated at 33 rows. It does not silently drop to
30, and the CTL rows do not linger as orphans.

## Archived Milestones

<details>
<summary>✅ M002 Interprocedural Semantic Spine (Phases 07-13) — SHIPPED 2026-09-14</summary>

- [x] Phase 07: Calls, Signatures, and Call-Graph Refusal (12/12 plans) — completed 2026-09-09
- [x] Phase 08: Interprocedural Loan Liveness in `check` (6/6 plans) — completed 2026-09-10
- [x] Phase 09: Peer Re-Derivation and D-03-02 Closure (10/10 plans) — completed 2026-09-10
- [x] Phase 10: Trusted Interprocedural Oracle (9/9 plans) — completed 2026-09-11
- [x] Phase 11: Multi-Function Native Emission and Interprocedural Equivalence (9/9 plans) — completed 2026-09-12
- [x] Phase 12: `Result` Payloads (8/8 plans) — completed 2026-09-13
- [x] Phase 13: Agent Loop for Interprocedural Defects (7/7 plans) — completed 2026-09-13

Full phase details: [`milestones/M002-ROADMAP.md`](milestones/M002-ROADMAP.md)
Requirements (29/31 complete, 2 ratified partial): [`milestones/M002-REQUIREMENTS.md`](milestones/M002-REQUIREMENTS.md)
Audit: [`milestones/M002-MILESTONE-AUDIT.md`](milestones/M002-MILESTONE-AUDIT.md)
Execution artifacts: `milestones/M002-phases/`

</details>

<details>
<summary>✅ M001 Source-to-Native Semantic Spine (Phases 1-6) — SHIPPED 2026-09-07</summary>

- [x] Phase 1: Canonical Pure Spine (3/3 plans) — completed 2026-09-03
- [x] Phase 2: Owned Values and Abilities (7/7 plans) — completed 2026-09-03
- [x] Phase 3: Borrowed Views and CFG Lifetimes (10/10 plans) — completed 2026-09-04
- [x] Phase 4: Fallible Resources and C Boundary (13/13 plans) — completed 2026-09-05
- [x] Phase 5: Native Equivalence and Adversarial Evidence (14/14 plans) — completed 2026-09-06
- [x] Phase 6: Agent Feedback and Performance Ratification (15/15 plans) — completed 2026-09-07

Full phase details: [`milestones/M001-ROADMAP.md`](milestones/M001-ROADMAP.md)
Requirements (28/28 complete): [`milestones/M001-REQUIREMENTS.md`](milestones/M001-REQUIREMENTS.md)
Audit: [`milestones/M001-MILESTONE-AUDIT.md`](milestones/M001-MILESTONE-AUDIT.md)
Execution artifacts: `milestones/M001-phases/`

</details>

## Known Constraint

`.planning/ROADMAP.md` is referenced by a doc-grep validation row that M002's
archival already broke (see `research/M003/EVIDENCE-AND-DEBT.md`). **Phase 14
owns the repair** — it is exactly the class of defect EVD-01's groundedness lint
is built to find, and fixing it by hand here would deprive the lint of its first
real finding.

---
*Roadmap created: 2026-09-17 for milestone M003*

### Phase 21: Native Emission Ownership and Resource Discharge (M004)

**Status**: Planned owner designation only; M003 Phases 14-20 complete before
M004 work is scheduled.
**Goal**: Design checked resource discharge and foreign-boundary ownership
contracts required before any M003-cut emitter family can be reconsidered.
**Owns**: D-16-11 (`emitLinearForeign`), D-16-12
(`emitLinearBorrowedByPointer`), and D-16-13
(`emitLinearBorrowedByPointerPlain`).
**Admission boundary**: This ownership assignment does not admit or reopen any
emitter family. Each family retains its own prerequisites, cross-host evidence,
refusal-fence requirements, and one-translation-unit/`-flto` limitation.

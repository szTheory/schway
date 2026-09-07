# Roadmap: Codename Lang

## Overview

M001 builds a real compiler spine by adding semantic pressure vertically rather
than completing disconnected frontend, checker, interpreter, and backend layers.
Phase 1 proves the smallest pure source-to-native path. Each following phase
extends that same path with ownership, borrowing, resources/FFI, adversarial
native evidence, and finally the bounded agent/human feedback service.

## Milestone

**M001 — Source-to-Native Semantic Spine**

## Phases

- [x] **Phase 1: Canonical Pure Spine** - Format, check, interpret, and natively run one nominal ADT transformation end to end.
- [x] **Phase 2: Owned Values and Abilities** - Add affine transfer and independently derived type abilities across the same spine.
- [x] **Phase 3: Borrowed Views and CFG Lifetimes** - Add shared/exclusive loans, edge-specific last use, and public borrow origins. (completed 2026-09-04)
- [x] **Phase 4: Fallible Resources and C Boundary** - Prove partial cleanup and typed foreign obligations through a real C call. (completed 2026-09-05)
- [x] **Phase 5: Native Equivalence and Adversarial Evidence** - Preserve semantics under optimization, sanitizers, and hostile mutations. (completed 2026-09-06)
- [ ] **Phase 6: Agent Feedback and Performance Ratification** - Expose bounded query/explain/verify/evidence protocols and ratify feedback budgets.

## Phase Details

### Phase 1: Canonical Pure Spine

**Goal**: A contributor can take one provisional canonical Lang program through lossless parsing, typed core, deterministic interpretation, readable C17, and native execution with one command.
**Depends on**: Nothing (first phase)
**Requirements**: FND-01, FND-02, FND-03, SYN-01, SYN-02, SYN-03, SYN-04, SEM-01, SEM-02, INT-01, NAT-01, DX-01
**Canonical refs**: `wiki/real-frontend-core-ir-entry-plan.md`, `wiki/semantic-kernel-contract.md`, `wiki/semantic-kernel-probes.md`, `wiki/compiler-and-feedback-latency.md`, `wiki/compute-efficiency-constitution.md`
**Success Criteria** (what must be TRUE):

  1. From a clean offline checkout, `lang format`, `lang check`, and both run engines process a pure nominal-ADT match fixture and produce matching results.
  2. Formatting is idempotent, comments survive, parse/format/parse preserves the semantic tree, and malformed syntax emits bounded stable diagnostics.
  3. The interpreter and Clang-built C17 path emit the same versioned semantic outcome for the fixture at `-O0` and `-O3`.
  4. A stale or mismatched compact manifest is rejected, while human and JSON command projections identify the same diagnostic/event IDs.

**Plans**: 3/3 plans executed

- [x] `01-01-PLAN.md` — source → typed core → interpreter/C17 tracer
- [x] `01-02-PLAN.md` — lossless canonical frontend and recovery pressure
- [x] `01-03-PLAN.md` — structured evidence, mutation controls, and phase gate

### Phase 2: Owned Values and Abilities

**Goal**: The end-to-end compiler distinguishes cheap implicit copy from explicit transfer and derives independent value abilities coherently.
**Depends on**: Phase 1
**Requirements**: OWN-01, OWN-02
**Canonical refs**: `wiki/ownership-lifetime-decisive-study.md`, `wiki/ownership-evidence-roadmap.md`, `.planning/spikes/001-ownership-kernel-workbench/README.md`, `.planning/spikes/003-public-origins-generic-abilities/README.md`
**Success Criteria** (what must be TRUE):

  1. A valid owned-buffer program transfers exactly once and has equivalent interpreter/native events.
  2. Use after move and move during an active loan are rejected with stable cause chains and smallest repair choices.
  3. Copy, drop, share, send, and escape abilities derive independently through representative aggregate and generic shapes.

**Plans**: 7/7 plans executed; verified 14/14 at `3399ddc` after seven review waves. Nine fail-closed controls; Phase 1 goldens byte-frozen. One accepted override (`02-OVERRIDES.md` OV-02-01: `Buffer` grants `share`); nine accepted debt items carried to Phase 3 (`02-DEBT.md`).

- [x] 02-01-PLAN.md
- [x] 02-02-PLAN.md
- [x] 02-03-PLAN.md
- [x] 02-04-PLAN.md
- [x] 02-05-PLAN.md
- [x] 02-06-PLAN.md
- [x] 02-07-PLAN.md — runtime-derived owned C facts and real backend causality control

### Phase 3: Borrowed Views and CFG Lifetimes

**Goal**: Local borrows remain ergonomic through precise last-use inference while public borrowed results remain explicit and separately checkable.
**Depends on**: Phase 2
**Requirements**: OWN-03, OWN-04
**Canonical refs**: `wiki/ownership-evidence-roadmap.md`, `.planning/spikes/002-cfg-edge-last-use/README.md`, `.planning/spikes/003-public-origins-generic-abilities/README.md`, `.planning/spikes/004-independent-certificate-checker/README.md`
**Success Criteria** (what must be TRUE):

  1. Shared and exclusive borrow conflicts are accepted or rejected identically by the core checker and bounded path oracle.
  2. A branch-specific last use ends a loan on the correct CFG edge without a manual scope block; omitting that edge is detected.
  3. A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection.
  4. Separate compilation rejects stale, omitted, or impossible public-origin summaries while retaining the coordinated-frontend-lie limitation explicitly.

**Scope note**: OWN-03 is scoped to **acyclic** (branch-only) CFG edges this phase. The language has no loop or recursion construct and this phase adds neither; loop-carried loan liveness is an explicit Phase 4+ follow-on.

**Plans**: 10/10 plans executed (03-01..03-07 executed; 03-08 and 03-09 are executed gap-closure plans; re-verification found a new multi-arm instance of the same origin-recomputation defect class — 03-10 is the third-round gap-closure plan)

- [x] 03-10-PLAN.md

- [x] 03-08-PLAN.md
- [x] 03-09-PLAN.md

**Wave 1**

- [x] 03-01-PLAN.md
- [ ] `03-01-PLAN.md` — reserved-identifier rename, match-arm bodies, and the first real CFG
- [ ] `03-02-PLAN.md` — exclusive loans and the shared/exclusive conflict matrix across all four operation-kind sites
- [ ] `03-03-PLAN.md` — backward worklist loan liveness with edge-specific endpoints and honest counted work
- [ ] `03-04-PLAN.md` — the independent validator's own CFG liveness derivation and endpoint mutation control
- [ ] `03-05-PLAN.md` — bounded path oracle, mutation kill, and generators that reach branching shapes
- [ ] `03-06-PLAN.md` — public borrow origins, body-blind summary verification, and separate compilation
- [ ] `03-07-PLAN.md` — bounded debug-lineage experiment, Phase 3 gate, and carried debt closure
- [ ] `03-08-PLAN.md` — gap closure (GAP 1 / CR-01): first-seen access guard in `RecomputeOrigin` plus the mixed-access reborrow-chain regression fixture
- [ ] `03-09-PLAN.md` — gap closure (GAP 2 / D-03-02): reject an omitted public origin on the publication path, wired as two new fail-closed controls
- [ ] `03-10-PLAN.md` — gap closure (SC3/SC4, multi-arm): walk every `OpReturn`, combine per-arm origins conservatively, two new fixtures and two new fail-closed controls

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 03-02-PLAN.md
- [x] 03-06-PLAN.md

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 03-03-PLAN.md

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 03-04-PLAN.md

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 03-05-PLAN.md

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 03-07-PLAN.md

**Cross-cutting constraints:**

- A public borrowed view names all verified field/alternative origins and access modes without downstream body inspection (ROADMAP SC3, verbatim — the truth 03-VERIFICATION.md falsified)

**Wave structure** (7 plans, 6 waves — two parallel tracks joining at the gate):

| Wave | Plans | Track |
|---|---|---|
| 1 | 03-01 | shared groundwork: `_LANG_` rename, arm bodies, first CFG |
| 2 | 03-02, **03-06** | OWN-03 chain starts; OWN-04 runs in parallel |
| 3 | 03-03 | OWN-03 |
| 4 | 03-04 | OWN-03 |
| 5 | 03-05 | OWN-03 terminal — **mid-phase gate here** |
| 6 | 03-07 | joins both tracks; carries the phase gate |

**Decision (2026-09-04): one phase, with a mandatory mid-phase gate at wave 5.**
The plan-checker recommended splitting at 03-01..05 (OWN-03) / 03-06..07 (OWN-04)
to obtain an intermediate verify gate, citing Phase 2's compounding-wave-defect
history. The split is not applied: 03-07 carries the gate for *both* tracks, so a
clean split would require replanning 03-07, not relabeling it. The checker's actual
concern — no verification between wave 1 and wave 7 — is closed instead by two
changes that cost nothing structurally:

1. **03-06 was re-pinned** from `depends_on: ["03-05"]` to `["03-01"]`, the
   checker's own minimum remedy. OWN-04 touches none of the CFG/liveness/oracle
   code, so it now runs as a parallel wave-2 track instead of serializing behind it.
2. **A mid-phase gate is required after 03-05** — the OWN-03 terminal plan — and
   before 03-07 assembles: run an independent code review and goal-backward
   verification scoped to OWN-03 (criteria 1 and 2) on the wave-5 tree. Wave 6 does
   not start until that gate is clean or its findings are recorded as dated debt.
   This is the intermediate ship/verify point the split would have bought.

**Cross-cutting constraints for every Phase 3 plan:**

- OWN-03 is acyclic (branch-only) this phase; loop-carried liveness is Phase 4+.
- `__LANG_` → `_LANG_` (D-02-05) lands as 03-01's first standalone commit, **before**
  any new C artifact is frozen.
- No `lang.core/2`. New core fields are additive and `omitempty`; Phase 1 and Phase 2
  goldens stay byte-identical.
- Any new `OperationKind` must be updated at **four** sites — `check`, `corevalidate`,
  `interp`, `cgen` (D-12a). A missed site silently drops the operation from the trace
  the O0/O3 differential compares.

### Phase 4: Fallible Resources and C Boundary

**Goal**: A noncopyable resource crosses one audited C boundary while partial initialization, failure propagation, and cleanup remain defined.
**Depends on**: Phase 3
**Requirements**: SEM-03, RES-01, FFI-01
**Canonical refs**: `wiki/semantic-kernel-contract.md`, `wiki/native-and-low-level-profile.md`, `wiki/memory-reclamation-policy.md`, `.planning/spikes/005-native-ffi-provenance-cleanup/README.md`
**Success Criteria** (what must be TRUE):

  1. A fallible two-step acquisition releases only initialized resources, exactly once, in reverse order on success and typed failure.
  2. Generated C declarations and adapters make target layout, allocator identity, alias/capture, callback retention, and unwind policy inspectable.
  3. Panic cannot cross the ordinary non-unwinding C boundary, and a foreign nonlocal exit cannot silently bypass Lang cleanup.
  4. Interpreter and native executions agree on primary failure and cleanup events.

**Plans**: 13/13 plans executed (04-01..04-07 executed; 04-08, 04-09 and 04-10 are executed gap-closure plans; the fourth-round re-verification confirmed two NEW gaps by direct code read — the interior-merge collapse in `checkReleaseOrder`'s `rederive` walk and the unsanitized `core.ForeignContract.Symbol` splice into generated C — so 04-11 and 04-12 are the fourth-round gap-closure plans, not yet executed)

- [x] 04-13-PLAN.md

- [x] 04-11-PLAN.md
- [x] 04-12-PLAN.md

- [x] 04-10-PLAN.md

- [ ] `04-11-PLAN.md` — gap closure (RES-01): plural `okEdgeInto`, per-candidate independent rederivation and agreement at interior merges, refused with `core.release_order_merge_mismatch`, falsified in both edge orderings
- [ ] `04-12-PLAN.md` — gap closure (FFI-01): audit `core.ForeignContract.Symbol` as a C identifier with `foreign.symbol_not_identifier` before any splice, plus cgen's own independent refusal on the three `EmitForeign*` entry points that never call `Validate`
- [ ] `04-13-PLAN.md` — gap closure (FFI-01): refuse a source-reachable hostile `foreign C { }` policy value at admission with `check.foreign_policy_value_unsafe`, audit `Allocator`/`Unwind`/`NonlocalExit` and every remaining spliced contract string in corevalidate (`foreign.policy_value_not_identifier`, `foreign.contract_field_not_c_safe`), and give cgen its own independent field guard in `singleForeignFunction`

- [x] 04-08-PLAN.md
- [x] 04-09-PLAN.md

**Wave 1**

- [ ] `04-08-PLAN.md` — gap closure: make the merge-terminal release-order rederivation falsifiable and add its structural peer check
- [ ] `04-09-PLAN.md` — gap closure: the conformance-layer attribute falsifier and a pin on the scan's artifact count
- [ ] `04-10-PLAN.md` — gap closure (CR-01): visited-set guard and `core.release_order_cyclic` refusal in `rederive`, with a cyclic-ok-edge falsifier asserting the validator returns rather than hangs
- [x] 04-01-PLAN.md
- [ ] `04-01-PLAN.md` — byte-identity pins, the operation-kind registry and six-site dispatch control, and the end-to-end fallible foreign call across the frozen C boundary
- [ ] `04-02-PLAN.md` — three-stage partial acquisition with materialized reverse-order release, an independent rederivation, and two mutations attacking different artifacts
- [ ] `04-03-PLAN.md` — the authoritative foreign contract, its three inspectable layers, the layout compile-time refusal, and zero optimizer-visible attributes
- [ ] `04-04-PLAN.md` — the reachable abort-only defect, the closed terminal-outcome axis, and the additive streaming event emitter
- [ ] `04-05-PLAN.md` — the process-root landing pad and static ledger, the undefined-symbol allowlist, and mutation-killed nonlocal-exit detection
- [ ] `04-06-PLAN.md` — every terminator walked by both independent analyses, foreign-return origin facts, and honest counted work
- [ ] `04-07-PLAN.md` — the Phase 4 gate and its contract test, three-engine agreement across every path, and the shipped-binary out-of-corpus exercise

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 04-02-PLAN.md

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 04-03-PLAN.md

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 04-04-PLAN.md

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 04-05-PLAN.md

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 04-06-PLAN.md

**Wave 7** *(blocked on Wave 6 completion)*

- [x] 04-07-PLAN.md

**Cross-cutting constraints:**

- One authoritative `core.ForeignContract` carries target layout, initialized state, allocator identity, capture and retention, aliasing, and unwind obligations, and all three inspectable layers are derived from it so no layer can invent a fact the contract does not carry (FFI-01, D-04-12, ROADMAP SC2).

**Waves**: 1→7, strictly sequential. Every plan touches the checker, the independent
validator, the code generator, or the session gate, so no two plans have disjoint
`files_modified`; parallelism is not available and is not claimed.

### Phase 5: Native Equivalence and Adversarial Evidence

**Goal**: The source-to-native subset survives optimization and deliberately seeded boundary defects with independently meaningful evidence.
**Depends on**: Phase 4
**Requirements**: INT-02, NAT-02, NAT-03, QLT-01
**Canonical refs**: `wiki/semantic-kernel-probes.md`, `wiki/ownership-evidence-roadmap.md`, `.planning/spikes/MANIFEST.md`, `.planning/spikes/004-independent-certificate-checker/README.md`, `.planning/spikes/005-native-ffi-provenance-cleanup/README.md`
**Success Criteria** (what must be TRUE):

  1. Interpreter, `-O0`, and `-O3` match on terminal outcome, semantic-event order, and live-resource state across the milestone corpus.
  2. All seven native hostile mutations and all applicable ownership/certificate mutations are detected by the intended independent lane.
  3. ASan/UBSan evidence is isolated from semantic equivalence evidence and catches retained-pointer lifetime defects.
  4. Any injected mismatch reports a minimized source/core case and causal event trace; the coordinated source-to-core false claim remains a documented escape.

**Plans**: 14/14 plans executed

Plans:
**Wave 1**

- [x] 05-01-PLAN.md — Phase 1-4 identity pin, by-pointer lowering tracer, `Alias` discharge (wave 1)
- [x] 05-02-PLAN.md — `discoverLoanLastUses` shadow-mode retirement (wave 1)
- [x] 05-03-PLAN.md — Sanitizer lane: isolated binary, pinned options, availability probe (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 05-04-PLAN.md — Alias-fact proof, `restrict` emission, manifest justification binding (wave 2)
- [x] 05-05-PLAN.md — Phase 5 corpus: adversarial subset plus bounded enumerated closure (wave 2)
- [x] 05-06-PLAN.md — `-O3 -flto` tier, five-axis comparator, fail-closed field routing (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 05-07-PLAN.md — `control:alias.false_no_alias` and the NAT-03 axis-movement table (wave 3)
- [x] 05-08-PLAN.md — Foreign dynamic fixtures: allocator mismatch, UAF, retained pointer (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 05-09-PLAN.md — MID-PHASE GATE: `Phase5RequiredControls()` and `scripts/verify-phase5.sh` (wave 4)

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 05-10-PLAN.md — Core reducer: five moves, strict predicate, bounded budget (wave 5)
- [x] 05-11-PLAN.md — QLT-01 control registry and executable audit (wave 5)

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 05-12-PLAN.md — `lang.mismatch/0` document and the three reducer mutation-kills (wave 6)
- [x] 05-13-PLAN.md — Demonstrable coordinated source-to-core false-claim escape (wave 6)

**Wave 7** *(blocked on Wave 6 completion)*

- [x] 05-14-PLAN.md — Final gate, M002 charter and debt-register entries (wave 7)

### Phase 6: Agent Feedback and Performance Ratification

**Goal**: AI agents and human reviewers can inspect, repair, and verify the compiler corpus through bounded stable protocols whose cost is measured and enforced.
**Depends on**: Phase 5
**Requirements**: FND-04, DX-02, DX-03, DX-04, QLT-02
**Canonical refs**: `wiki/compiler-and-feedback-latency.md`, `wiki/compute-efficiency-constitution.md`, `wiki/performance-observability-and-delivery.md`, `wiki/ai-native-runtime-and-evals.md`
**Success Criteria** (what must be TRUE):

  1. `explain` and `query` return bounded versioned facts and cause graphs by stable identity without requiring prose scraping or whole-program dumps.
  2. `verify` selects only evidence lanes relevant to changed risk and reports reused cache inputs; `evidence` validates compact manifests and expands traces on failure/request.
  3. An automated repair exercise fixes representative match, move, borrow, cleanup, and stale-evidence defects using only the supported command protocol.
  4. Declared-machine cold/warm feedback distributions, peak memory, output bytes, and affected work meet ratified budgets or identify a specific blocking regression.

**Plans**: 7/15 plans executed

Plans:
**Wave 1**

- [x] 06-01-PLAN.md — Wave 0: freeze prior-phase bytes, pin Metrics/Lane identity exclusion, pin the 12 lane-schema sites

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 06-02-PLAN.md — `explain`: bounded, deterministic cause DAG under the net-new `lang.explain/0`
- [x] 06-04-PLAN.md — `internal/compiler/cache`: declared-input artifact store that structurally cannot hold a verdict
- [x] 06-08-PLAN.md — `internal/compiler/measure`: leak-free machine probe, 20-sample p50/p95, CoV auto-demotion
- [x] 06-11-PLAN.md — `diagnostic.Repair` grows into an applicable edit; identity split held at `kind` only

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 06-03-PLAN.md — `query`: five-vocabulary join, cursor-paginated, under `lang.query/0`
- [x] 06-05-PLAN.md — `risk_lanes.json`, conservative widening, gitignored change-state file, bidirectional audit
- [ ] 06-12-PLAN.md — Five defect injectors, held-out corpus, and anti-theater guard 3 (marker mutation-kill)

**Wave 4** *(blocked on Wave 3 completion)*

- [ ] 06-06-PLAN.md — The one coordinated additive bump: `lang.command/1`, `lang.verify-lane/1`, seven reporting fields
- [ ] 06-13-PLAN.md — `cmd/lang-repair`: subprocess-only driver, import boundary, non-degenerate success oracle

**Wave 5** *(blocked on Wave 4 completion)*

- [ ] 06-07-PLAN.md — Wire cache + risk selection into `verify`; deferred lanes never render pass; `evidence` trace expansion
- [ ] 06-14-PLAN.md — Anti-theater guards 1 and 2: prose-scramble identical, vocabulary-removal goes RED

**Wave 6** *(blocked on Wave 5 completion)*

- [ ] 06-09-PLAN.md — `qlt02_budget_manifest.json`, executable audit, observation-only mode, the blocking rule

**Wave 7** *(blocked on Wave 6 completion)*

- [ ] 06-10-PLAN.md — Stage attribution without a tracer; peak-RSS gap recorded as a deliberate decision

**Wave 8** *(blocked on Wave 7 completion)*

- [ ] 06-15-PLAN.md — `scripts/verify-phase6.sh`, required-control set, escape register, 06-DEBT.md

### M002 Charter (deferred from Phase 5, D-05-32/D-05-33)

**`OpCall` (Lang-to-Lang calls) is M002's LEAD charter item — recorded explicitly
here so the deferral is an owned, scheduled decision, never an unowned carry.**
No phase in the current M001 roadmap owns this work: Phase 6 is agent feedback
and performance ratification, not language surface.

M002's lead item, named explicitly:

- `OpCall` itself — the new `OperationKind` at all six dispatch sites (D-12a).
- Interprocedural loan liveness in BOTH admission layers (`check` and
  `corevalidate`), rebuilt cross-function rather than the single-function
  differentials Phase 3/4/5 shipped.
- Call-graph construction and cycle refusal.
- A bounded interpreter call stack.
- The cross-function rebuild of Phase 3's exhaustive differentials.

**Gated on**: callable ⊆ publishable (D-04-03) — a function is callable only if
`originvalidate.ValidatePublished` would publish it. This lift condition is
unchanged by Phase 5.

**Why deferred out of M001**: pulling `OpCall` into Phase 5 would have landed a
new `OperationKind` at six dispatch sites simultaneously with alias-fact
emission, the first sanitizer lanes, the first reducer, and the QLT-01
registry — the exact fingerprint of the failure that cost Phase 2 a
remediation round, Phase 3 a mid-phase gate plus three gap-closure plans, and
Phase 4 thirteen plans and five review rounds (D-05-32).

**Accepted consequence, stated plainly**: **M001 ships without Lang-to-Lang
calls.** Interprocedural `-O3` equivalence is outside M001's proof scope BY
CONSTRUCTION, since M001 ships without calls — this is not an accident of
Phase 5's success-criteria wording, it is a direct structural consequence of
this deferral. `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-DEBT.md`'s
D-03-02 remains open past the milestone on this basis (see
`.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md`
for the verbatim carry-forward entry).

## Progress

**Execution Order:** Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6.

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Canonical Pure Spine | 3/3 | Complete | 2026-09-03 |
| 2. Owned Values and Abilities | 7/7 | Complete    | 2026-09-03 |
| 3. Borrowed Views and CFG Lifetimes | 10/10 | Complete    | 2026-09-04 |
| 4. Fallible Resources and C Boundary | 13/13 | Complete    | 2026-09-05 |
| 5. Native Equivalence and Adversarial Evidence | 14/14 | Complete    | 2026-09-06 |
| 6. Agent Feedback and Performance Ratification | 7/15 | In Progress|  |

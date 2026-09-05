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
- [ ] **Phase 4: Fallible Resources and C Boundary** - Prove partial cleanup and typed foreign obligations through a real C call.
- [ ] **Phase 5: Native Equivalence and Adversarial Evidence** - Preserve semantics under optimization, sanitizers, and hostile mutations.
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

**Plans**: 6/7 plans executed

- [x] 04-01-PLAN.md
- [x] 04-02-PLAN.md
- [x] 04-03-PLAN.md
- [x] 04-04-PLAN.md
- [x] 04-05-PLAN.md
- [x] 04-06-PLAN.md
- [ ] 04-07-PLAN.md

- [ ] `04-01-PLAN.md` — byte-identity pins, the operation-kind registry and six-site dispatch control, and the end-to-end fallible foreign call across the frozen C boundary
- [ ] `04-02-PLAN.md` — three-stage partial acquisition with materialized reverse-order release, an independent rederivation, and two mutations attacking different artifacts
- [ ] `04-03-PLAN.md` — the authoritative foreign contract, its three inspectable layers, the layout compile-time refusal, and zero optimizer-visible attributes
- [ ] `04-04-PLAN.md` — the reachable abort-only defect, the closed terminal-outcome axis, and the additive streaming event emitter
- [ ] `04-05-PLAN.md` — the process-root landing pad and static ledger, the undefined-symbol allowlist, and mutation-killed nonlocal-exit detection
- [ ] `04-06-PLAN.md` — every terminator walked by both independent analyses, foreign-return origin facts, and honest counted work
- [ ] `04-07-PLAN.md` — the Phase 4 gate and its contract test, three-engine agreement across every path, and the shipped-binary out-of-corpus exercise

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

**Plans**: TBD

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

**Plans**: TBD

## Progress

**Execution Order:** Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6.

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Canonical Pure Spine | 3/3 | Complete | 2026-09-03 |
| 2. Owned Values and Abilities | 7/7 | Complete    | 2026-09-03 |
| 3. Borrowed Views and CFG Lifetimes | 10/10 | Complete    | 2026-09-04 |
| 4. Fallible Resources and C Boundary | 6/7 | In Progress|  |
| 5. Native Equivalence and Adversarial Evidence | 0/TBD | Not started | - |
| 6. Agent Feedback and Performance Ratification | 0/TBD | Not started | - |

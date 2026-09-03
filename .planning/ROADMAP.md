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
- [ ] **Phase 2: Owned Values and Abilities** - Add affine transfer and independently derived type abilities across the same spine.
- [ ] **Phase 3: Borrowed Views and CFG Lifetimes** - Add shared/exclusive loans, edge-specific last use, and public borrow origins.
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

**Plans**: 7/7 plans executed; independent phase re-verification pending

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

**Plans**: TBD

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

**Plans**: TBD

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
| 2. Owned Values and Abilities | 7/7 | Awaiting re-verification |  |
| 3. Borrowed Views and CFG Lifetimes | 0/TBD | Not started | - |
| 4. Fallible Resources and C Boundary | 0/TBD | Not started | - |
| 5. Native Equivalence and Adversarial Evidence | 0/TBD | Not started | - |
| 6. Agent Feedback and Performance Ratification | 0/TBD | Not started | - |

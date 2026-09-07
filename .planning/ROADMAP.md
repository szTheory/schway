# Roadmap: Codename Lang

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1-6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)

## Phases

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

## Carried Into M002

M001 ships without Lang-to-Lang calls. `OpCall` was deliberately deferred rather
than landed alongside alias-fact emission, the first sanitizer lanes, the
reducer, and the QLT-01 registry (D-05-32). The following is M002's lead charter:

- `OpCall` as a real `OperationKind` at all six dispatch sites, gated on
  callable ⊆ publishable (D-04-03).
- Interprocedural loan liveness in BOTH admission layers (`check` and
  `corevalidate`), rebuilt cross-function — closes D-03-02, the one debt item
  knowingly left open past the milestone.
- Call-graph construction with cycle refusal, plus a bounded interpreter call
  stack.
- Cross-function rebuild of Phase 3's exhaustive loan-endpoint differentials.
- Interprocedural `-O3` equivalence, outside M001's proof scope by construction.
- Storable/matchable `Result` values and payload-carrying alternatives (D-04-30).

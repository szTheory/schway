# Roadmap: Codename Lang

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1-6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07-13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ⬜ **M003 — TBD** — run `/gsd-new-milestone` to define

## Phases

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

## Next

M003 is not yet defined. Run `/gsd-new-milestone` to run questioning → research
→ requirements → roadmap.

The debt M002 leaves unowned is the strongest available input to that
conversation — see `PROJECT.md` `## Next Milestone Goals` and
`milestones/M002-MILESTONE-AUDIT.md` §4. In short: the single-function emitter
deletion (a third deferral would violate the project's own D-10-60 rule), the
unowned event-identity chain (D-11-51 → D-12-21), and the single-type-per-
function invariant whose lifting closes DX-06 and DX-07 together.

---
*Last updated: 2026-09-14 after the M002 milestone*

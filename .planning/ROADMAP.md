# Roadmap: Codename Lang

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1–6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07–13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ✅ **M003 — Computation and Honest Instruments** — Phases 14–20 (shipped 2026-09-26; audit: tech debt) — [archive](milestones/M003-ROADMAP.md)
- ◷ **M004 — Native Emission Ownership and Resource Discharge** — provisional; formal requirements and full charter not yet ratified — [milestone kickoff](milestones/M003-ROADMAP.md)

## Phases

<details>
<summary>✅ M003 — Computation and Honest Instruments (Phases 14–20) — SHIPPED 2026-09-26</summary>

- [x] Phase 14: Evidence Instrument and Honest Scoping (13/13 plans) — completed 2026-09-18
- [x] Phase 15: Event Identity (`lang.execution/2`) (10/10 plans) — completed 2026-09-26
- [x] Phase 16: Branch/Match Emitter Port (26/26 plans) — completed 2026-09-25; verification refreshed 2026-09-26
- [x] Phase 17: Return Type ≠ Parameter Type (9/9 plans) — completed 2026-09-22
- [x] Phase 18: Branch on a Computed Value (10/10 plans) — completed 2026-09-25
- [x] Phase 19: Numeric Literals and `OpConst` (7/7 plans) — completed 2026-09-26; verification refreshed
- [x] Phase 20: Nyquist, D-13-34, and the Frontier Fixture (10/10 plans) — completed 2026-09-26

Full phase details: [M003 roadmap](milestones/M003-ROADMAP.md)  
Requirements: [M003 requirements](milestones/M003-REQUIREMENTS.md)  
Audit: [M003 milestone audit](milestones/M003-MILESTONE-AUDIT.md)  
Execution artifacts: `milestones/M003-phases/`

</details>

<details>
<summary>◷ M004 — Native Emission Ownership and Resource Discharge (provisional)</summary>

Phase 21's original implementation is complete (4/4 plans), and automated UAT
is complete (7/7). The refreshed verification reports `gaps_found` against the
post-M003 archive. Preserve UAT and do not replay the original plans or UAT.

The two gap-closure plans (21-05 and 21-06) are prepared and registered below.
`roadmap.get-phase 21` now resolves the existing phase. M004 remains provisional;
this registration does not ratify its full charter or requirements. Execute only
the gap-closure plans with `$gsd-execute-phase 21 --gaps-only`.

After verification passes, file Phase 21 under
`.planning/milestones/M004-phases/` before running M004 kickoff. The installed
`gsd-new-milestone` cleanup archives every directory still under
`.planning/phases/` as part of the outgoing milestone, regardless of roadmap
ownership; leaving Phase 21 there would place it in M003's archive.

The full M004 charter and requirements have not been ratified. The provisional
forward arc includes iteration, which requires a pre-phase spike to decide what
replaces `pathoracle`'s cycle refusal before loop planning.

</details>

### Phase 21: Native Emission Ownership and Resource Discharge

**Goal:** Close the archive-dependent findings in Phase 21 verification while preserving the completed implementation plans and UAT. M004 remains provisional; this entry registers only the already-existing gap-closure work.

**Plans:** 6 plans

Plans:
**Wave 1**

- [x] 21-01-PLAN.md — Define the checked resource-discharge contract.
- [x] 21-02-PLAN.md — Retire unreachable emitter bodies and pin refusal.
- [x] 21-03-PLAN.md — Record the scoped emitted-fixture LTO comparison.

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 21-04-PLAN.md — Reconcile emitter debt and evidence records.

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 21-05-PLAN.md — Restore archived Phase 16 decision and ownership guards.
- [ ] 21-06-PLAN.md — Refresh validation corpus and groundedness reconciliation.

## Next Action

Run `$gsd-execute-phase 21 --gaps-only`. Once verification passes, file Phase 21
under M004 as described above, then run `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"`
to formalize M004 scope and requirements.

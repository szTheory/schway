# Roadmap: Codename Lang

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1–6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07–13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ✅ **M003 — Computation and Honest Instruments** — Phases 14–20 (shipped 2026-09-26; audit: tech debt) — [archive](milestones/M003-ROADMAP.md)
- ◷ **M004 — Native Emission Ownership and Resource Discharge** — provisional; formal requirements and full charter not yet ratified

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

Phase 21 is complete: all six plans are complete, all seven UAT cases are
preserved, and current verification passes 6/6. Its execution archive is filed
under [M004 phase artifacts](milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/),
outside M003's outgoing cleanup. M003 remains the shipped outgoing milestone.

M004 is still provisional; this entry does not ratify its full charter or
requirements. No M004 requirement IDs have been assigned. The next action is to
run `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"` to
formalize M004's scope and requirements.

The full M004 charter and requirements have not been ratified. The provisional
forward arc includes iteration, which requires a pre-phase spike to decide what
replaces `pathoracle`'s cycle refusal before loop planning.

</details>

### Phase 21: Native Emission Ownership and Resource Discharge

**Goal:** Close the archive-dependent findings in Phase 21 verification while preserving the completed implementation plans and UAT. M004 remains provisional; this entry records the completed gap-closure work.

**Plans:** 6/6 complete · **UAT:** 7/7 preserved · **Verification:** 6/6 passed

Plans:
**Wave 1**

- [x] 21-01-PLAN.md — Define the checked resource-discharge contract.
- [x] 21-02-PLAN.md — Retire unreachable emitter bodies and pin refusal.
- [x] 21-03-PLAN.md — Record the scoped emitted-fixture LTO comparison.

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 21-04-PLAN.md — Reconcile emitter debt and evidence records.

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 21-05-PLAN.md — Restore archived Phase 16 decision and ownership guards.
- [x] 21-06-PLAN.md — Refresh validation corpus and groundedness reconciliation.

## Next Action

Run `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"` to
formalize the provisional M004 scope and requirements.

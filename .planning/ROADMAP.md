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

Phase 21 implementation is complete (4/4 plans) and automated UAT passed (7/7).
Its verification fingerprint is stale after the recurring LTO evidence row was
updated. Run `$gsd-execute-phase 21` to resume at verification gates; preserve
the completed UAT and do not replay plans or UAT.

After verification passes, file Phase 21 under
`.planning/milestones/M004-phases/` before running M004 kickoff. The installed
`gsd-new-milestone` cleanup archives every directory still under
`.planning/phases/` as part of the outgoing milestone, regardless of roadmap
ownership; leaving Phase 21 there would place it in M003's archive.

The full M004 charter and requirements have not been ratified. The provisional
forward arc includes iteration, which requires a pre-phase spike to decide what
replaces `pathoracle`'s cycle refusal before loop planning.

</details>

## Next Action

After a context reset, run `$gsd-execute-phase 21` to refresh the stale Phase 21
verification report. Once it passes, file that phase under M004 as described
above, then run `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"`
to formalize M004 scope and requirements.

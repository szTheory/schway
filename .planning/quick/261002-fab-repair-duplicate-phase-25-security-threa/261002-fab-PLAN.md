---
id: 261002-fab
phase: quick
plan: 261002-fab
type: execute
mode: quick-validate
status: planned
wave: 1
depends_on: []
files_modified:
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-PLAN.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md
autonomous: true
requirements: [EVD-10, DX-14]
must_haves:
  truths:
    - "Each Phase 25 plan threat-register row has one plan owner: 25-05 retains T-25-13..15, 25-06 owns T-25-16..18, and 25-07 owns T-25-19..20."
    - "Validation task T-25-10 cites its own 25-06 threat IDs, and 25-07 continues to reference canonical security controls T-25-14 and T-25-15 as audit evidence."
    - "The canonical Phase 25 security register remains 15/15 closed, SECURED at ASVS L1 with threats_open: 0; later plan-local IDs are mapped to those controls without adding audit findings."
    - "Phase 25's #4683 duplicate-threat gate clears, while the fresh Phase 23 verification report and current Phase 25 hosted evidence remain untouched."
  artifacts:
    - path: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md
      provides: "Unique threat-register identities T-25-16..18 for the evidence and documentation task"
    - path: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-PLAN.md
      provides: "Unique threat-register identities T-25-19..20 with canonical audit cross-references retained"
    - path: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
      provides: "T-25-10 validation trace to the 25-06 plan-local threats"
    - path: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md
      provides: "Concise later-plan-to-canonical-control mapping beside the unchanged 15-row audit"
  key_links:
    - from: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
      to: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md
      via: "Task T-25-10 Threat Ref names T-25-16, T-25-17, and T-25-18"
    - from: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-PLAN.md
      to: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md
      via: "Plan-local T-25-19/T-25-20 map to canonical T-25-14/T-25-15 evidence without new audit rows"
---

<objective>
Clear the Phase 25 duplicate security-threat-ID gate and restore exact traceability between the later plans, validation map, and canonical security audit.

Purpose: Let the already-complete phase reach its report-refresh gate without changing its security verdict or implementation evidence.
Output: Four corrected planning records; the Phase 25 verification report is refreshed by its owning workflow after this quick task.
</objective>

<execution_context>
@/Users/jon/.codex/gsd-core/workflows/execute-plan.md
@/Users/jon/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-05-PLAN.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-PLAN.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md

The live execute-phase gate reports duplicate IDs T-25-13 in plans 25-05/06 and T-25-14/T-25-15 in plans 25-05/06/07. Plan 25-05 is the canonical source for those three existing audit controls. Plan 25-07's `provides` link and action deliberately cite the canonical T-25-14/T-25-15 evidence; those are cross-references, not threat-register claims. Validation task T-25-10 belongs to plan 25-06. The 2026-10-02 Phase 25 security record has 15 closed rows and `threats_open: 0`, with hosted run 36971855722 closing all 18 family/host/lane evidence rows. The fresh Phase 23 report and unrelated working-tree edits are outside this task.

Discovery level 0: this is an exact planning-record correction using existing formats and no new package. The API coverage detector examined the task scope and returned `detected=false`; no coverage matrix is needed. The assumption-delta query for quick task 261002-fab returned `skipped: true, reason: phase_unresolved`, so it makes no identity-model finding. The schema detector found no matching file among the four planned edits, so no schema push applies. Security enforcement is ASVS L1 with high-severity findings blocking; this correction introduces no new finding.
</context>

<tasks>

<task type="auto">
  <name>Task 1: Give later Phase 25 plan registers unique threat IDs</name>
  <files>.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md, .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-PLAN.md</files>
  <action>In the STRIDE Threat Register table of 25-06 only, rename its three claimed rows in order from T-25-13/14/15 to T-25-16/17/18. In the STRIDE Threat Register table of 25-07 only, rename its two claimed rows from T-25-14/15 to T-25-19/20. Keep each row's category, component, severity, disposition, mitigation, and task evidence unchanged. Preserve 25-05's T-25-13..15 register. Preserve 25-07's frontmatter `provides` string and task action references to canonical T-25-14/T-25-15: they name existing security-audit controls being updated, not local register identities. Do not rewrite historical implementation claims or the hosted run receipt.</action>
  <verify><automated>python3 -c 'import pathlib,re; d=pathlib.Path(".planning/phases/25-separate-pointer-successors-and-integrated-utility"); p={n:(d/f"25-{n}-PLAN.md").read_text().split("## STRIDE Threat Register",1)[1].split("</threat_model>",1)[0] for n in ("05","06","07")}; got={n:re.findall(r"^\| (T-25-\d+) \|",s,re.M) for n,s in p.items()}; assert got["05"]==["T-25-13","T-25-14","T-25-15"]; assert got["06"]==["T-25-16","T-25-17","T-25-18"]; assert got["07"]==["T-25-19","T-25-20"]; ids=sum(got.values(),[]); assert len(ids)==len(set(ids))' && git diff --check HEAD -- .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-06-PLAN.md .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-07-PLAN.md</automated></verify>
  <done>The three plan registers claim T-25-13 through T-25-20 exactly once across plans 25-05/06/07, while 25-07's canonical audit cross-references still identify T-25-14/T-25-15.</done>
</task>

<task type="auto">
  <name>Task 2: Reconcile validation and security cross-references</name>
  <files>.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md, .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md</files>
  <action>In 25-VALIDATION.md, change only task T-25-10's Threat Ref cell to T-25-16, T-25-17, T-25-18, matching its clean-checkout, matrix-receipt, and living-claim responsibilities in 25-06. Keep earlier task rows and the hosted 18/18 receipt unchanged. In 25-SECURITY.md, add a short note outside the canonical Threat Register that maps later plan-local T-25-16→canonical T-25-13, T-25-17→T-25-14, T-25-18→T-25-15, T-25-19→T-25-14, and T-25-20→T-25-15; explain that these are traceability aliases for later plan work, not extra audited controls. Retain exactly the 15 canonical threat rows, the dated 15/15 audit trail, SECURED/ASVS L1 sign-off, `threats_open: 0`, and run 36971855722 evidence. Do not edit implementation, dependencies, CI, ROADMAP, PRODUCT-ROADMAP, LANGUAGE-MATURITY, REQUIREMENTS, or Phase 23/25 verification reports. After edits, use the live GSD Phase 25 security gate to confirm the duplicate-ID block has cleared; leave the report refresh to the owning phase workflow.</action>
  <verify><automated>python3 -c 'import pathlib,re; d=pathlib.Path(".planning/phases/25-separate-pointer-successors-and-integrated-utility"); v=(d/"25-VALIDATION.md").read_text(); s=(d/"25-SECURITY.md").read_text(); row=next(x for x in v.splitlines() if x.startswith("| T-25-10 |")); assert "| T-25-16, T-25-17, T-25-18 |" in row; register=s.split("## Threat Register",1)[1].split("## Accepted Risks Log",1)[0]; ids=re.findall(r"^\| (T-25-\d+) \|",register,re.M); assert ids==[f"T-25-{i:02d}" for i in range(1,16)]; assert "threats_open: 0" in s and "SECURED" in s and "ASVS L1" in s and "36971855722" in s; assert re.search(r"\| 2026-10-02 \| 15 \| 15 \| 0 \|",s)' && git diff --check HEAD -- .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-SECURITY.md</automated></verify>
  <done>Validation points to 25-06's own three threat IDs; security text explains all five later-plan aliases while its 15 canonical controls, verdict, and evidence remain unchanged; the live duplicate-ID gate no longer blocks Phase 25.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Plan registers to GSD security gate | Duplicate claimed IDs can prevent a completed phase from reaching verification. |
| Plan-local threats to canonical audit | Later task evidence could be miscounted as new findings or lose its link to the existing 15 controls. |
| Validation and hosted receipts to report refresh | A documentation correction could accidentally weaken the verdict or disturb fresh verification inputs. |

## STRIDE Threat Register

ASVS L1; high and critical open findings block. These quick-task risks describe the document repair and are not added to the Phase 25 canonical audit.

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| Q-261002-fab-01 | Tampering / Repudiation | Plan threat identities | high | mitigate | Task 1 makes claimed plan-register IDs disjoint and verifies exact ownership; Task 2 checks the GSD gate. |
| Q-261002-fab-02 | Repudiation | Audit traceability | medium | mitigate | Task 2 maps all later-plan IDs to existing canonical controls and checks the original 15 rows and 15/15 audit receipt. |
| Q-261002-fab-03 | Tampering | Fresh reports and hosted evidence | high | mitigate | The file allowlist excludes verification reports and living roadmap inputs; Task 2 asserts the existing verdict and hosted run, then reviews the scoped diff. |
</threat_model>

<verification>Run the two task checks and inspect `git diff --name-only HEAD --` with the four target paths; compare overall working-tree status to the starting status so unrelated pre-existing edits remain intact. Run the live Phase 25 security-threat uniqueness gate used by `$gsd-execute-phase 25`; stop this quick task on any remaining #4683 collision. Use the existing full uncached Go regression and focused Phase 25 native run as prior, revision-scoped evidence only; this documentation correction requires no UAT or repeated implementation suite. The subsequent Phase 25 workflow refreshes its report from the final covered inputs.</verification>

<success_criteria>All five later-plan register IDs are unique, validation and security records trace them to their owners and canonical controls, the hard gate clears, and the 15/15 SECURED ASVS L1 audit remains intact without verifier-input churn.</success_criteria>

<output>Create `.planning/quick/261002-fab-repair-duplicate-phase-25-security-threa/261002-fab-SUMMARY.md` after execution.</output>

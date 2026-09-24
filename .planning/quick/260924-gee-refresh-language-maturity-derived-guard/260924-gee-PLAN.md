---
id: 260924-gee
phase: quick
plan: 260924-gee
type: execute
mode: quick-full
status: planned
wave: 1
depends_on: []
files_modified:
  - .planning/LANGUAGE-MATURITY.md
autonomous: true
must_haves:
  truths:
    - "The current maturity snapshot states 19 non-test len(Functions) != 1 guards across 6 files in 1 package, and 46 guards including tests."
    - "The current corpus snapshot states 137 .lang programs and 4,578 lines, with an accurate rounded average and 193-line maximum."
    - "The current re-assessment date and nearby guard prose describe the Phase 18 fixture-era snapshot without altering historical assessments or language-capability claims."
    - "The independent maturity self-check and full Go suite pass."
  artifacts:
    - path: .planning/LANGUAGE-MATURITY.md
      provides: Machine-checked current guard and corpus inventory
  key_links:
    - from: .planning/LANGUAGE-MATURITY.md
      to: internal/compiler/session/self_describing_docs_test.go
      via: TestLanguageMaturityCountsAreCurrent independently derives guard and corpus values
---

<objective>
Refresh the current language-maturity inventory after Phase 18 added fixtures.

Purpose: Restore an accurate, dated document and the independent documentation self-check.
Output: A scoped edit to `.planning/LANGUAGE-MATURITY.md` and recorded verification results.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/LANGUAGE-MATURITY.md
@internal/compiler/session/self_describing_docs_test.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Reconcile the current maturity snapshot with the independent check</name>
  <files>.planning/LANGUAGE-MATURITY.md</files>
  <action>Use `TestLanguageMaturityCountsAreCurrent` as the authority for the current figures. In the existing guard-total sentence, retain 19 non-test guards across 6 files in 1 package and change the including-tests count from 45 to 46; keep the `session` table row at 19 and reconcile the nearby prose that currently says 20 `session` guards. In the current corpus sentence, change 133 programs and 4,478 lines to 137 programs and 4,578 lines, update the rounded average to about 33 lines, and retain the verified 193-line maximum. Add a 2026-09-24 re-assessment entry and date the current corpus and guard headings to this Phase 18 fixture-era check; distinguish that current date from the historical 2026-09-17 EVD-06 introduction of the machine check. Preserve prior dated snapshots, the historical 32-to-22 correction, and language-capability claims. Edit only the maturity document; leave source, tests, Phase 18 artifacts, and existing unrelated or untracked work untouched. Record the actual focused and full-suite results in the quick summary.</action>
  <verify>
    <automated>go test ./internal/compiler/session -run '^(TestLanguageMaturityCountsAreCurrent|TestSelfDescribingDocsGuardIsNotInert)$' -count=1</automated>
    <automated>go test ./...</automated>
    <automated>git diff --check -- .planning/LANGUAGE-MATURITY.md</automated>
  </verify>
  <done>The current inventory states 19 non-test guards across 6 files in 1 package, 46 including tests, a `session` row and prose both at 19, and 137 `.lang` programs totaling 4,578 lines (about 33 average; 193 maximum). The current assessment is dated 2026-09-24 while historical dates remain historical; both Go commands pass and the diff has no whitespace errors.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Repository tree → maturity document | Counts in prose can drift from the source tree and give reviewers false evidence. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QGEE-01 | Repudiation | Current maturity counts | medium | mitigate | Keep the independent AST and corpus self-check authoritative; run it and the full Go suite after the scoped document edit. |
| T-QGEE-02 | Tampering | Historical assessment record | low | mitigate | Update only the current snapshot and its date/prose, preserving the prior dated measurements and EVD-06 history. |
</threat_model>

<source_audit>

| Source | Item | Task | Status |
|--------|------|------|--------|
| GOAL | Refresh derived guard inventory and corpus counts after Phase 18 fixtures | 1 | COVERED |
| REQ | Machine check is authoritative; current values are 19/6/1/46 and 137/4,578 | 1 | COVERED |
| REQ | Correct current date and nearby stale prose | 1 | COVERED |
| REQ | Documentation only; preserve unrelated and untracked work | 1 | COVERED |
| REQ | Restore focused and full-suite green verification | 1 | COVERED |
| RESEARCH | No separate research artifact for this quick documentation task | — | EXCLUDED |
| CONTEXT | No separate discussion artifact; parent task supplies the constraints above | 1 | COVERED |
</source_audit>

<verification>
The executor runs the independent maturity self-check and its non-inert control, then the full Go suite, and inspects the maturity-document diff for scope and whitespace errors.
</verification>

<success_criteria>
The current maturity document agrees with the independent machine check, carries an accurate 2026-09-24 re-assessment, and the previously isolated stale-document failure is gone.
</success_criteria>

<output>
Create `.planning/quick/260924-gee-refresh-language-maturity-derived-guard/260924-gee-SUMMARY.md` when done.
</output>

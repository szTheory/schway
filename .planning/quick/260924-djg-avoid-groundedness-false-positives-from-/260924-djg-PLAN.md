---
id: 260924-djg
phase: quick
plan: 260924-djg
type: execute
mode: quick-full
status: planned
wave: 1
depends_on: []
files_modified:
  - .planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md
  - .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
autonomous: true
must_haves:
  truths:
    - "Every Go test command in Phase 18 research and validation guidance runs an existing package without filtering on a future test name."
    - "Phase 18's requirement-specific prospective acceptance checks remain in the execution plans, including the fixture-first and wrong-slot mutation checks."
    - "The research and validation documents continue to describe all CTL-01, CTL-02, CTL-03, and S-010 evidence obligations honestly."
    - "The full Go regression suite passes after the document correction."
  artifacts:
    - path: .planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md
      provides: Runnable current-package verification guidance
    - path: .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
      provides: Runnable current-package verification map and measurement command
  key_links:
    - from: .planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md
      to: .planning/phases/18-branch-on-a-computed-value/18-01-PLAN.md
      via: Current package commands in research; prospective named acceptance in the plan
    - from: .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
      to: .planning/phases/18-branch-on-a-computed-value/18-08-PLAN.md
      via: Existing package command in the draft evidence record; phase-local helper remains a planned deliverable
---

<objective>
Ground Phase 18's research and validation commands in packages that exist now.

Purpose: Stop groundedness scans from treating prospective named tests in preparatory documents as current executable evidence, while retaining the precise acceptance commands in Phase 18 execution plans.
Output: Corrected research and validation guidance, with a passing full Go regression run.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md
@.planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
@.planning/phases/18-branch-on-a-computed-value/18-01-PLAN.md
@.planning/phases/18-branch-on-a-computed-value/18-08-PLAN.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Ground Phase 18 preparatory verification commands</name>
  <files>.planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md, .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md</files>
  <action>In the four requirement rows of RESEARCH.md's Phase Requirements → Test Map and the matching four Wave 0 rows of VALIDATION.md's Per-Task Verification Map, replace commands filtered to prospective TestPhase18 names with package-level Go test commands over packages that already exist. Update VALIDATION.md's draft `focused_command` measurement field to an existing package-level Go test command. The two plan-08 helper-script rows currently refer to a future script: make their current Automated Command cells runnable package-level Go tests and describe the script as a plan-08 deliverable, with its precise planned verification retained in 18-08-PLAN.md. Preserve the CTL and S-010 behavior, threat, axis, Wave 0, latency, and CI acceptance descriptions; keep “File Exists” honest about missing fixtures and named tests. Leave every 18-*-PLAN.md, reconciliation ledger, source file, and test unchanged. Run the full Go suite once after these document edits, record its result in the quick summary, and do not claim the missing Phase 18 tests have already passed.</action>
  <verify>
    <automated>! rg -n 'go test.*-run' .planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md</automated>
    <automated>git diff --quiet -- '.planning/phases/18-branch-on-a-computed-value/18-*-PLAN.md'</automated>
    <automated>go test ./...</automated>
  </verify>
  <done>Both preparatory documents name only existing-package Go test commands, their evidence descriptions remain specific and prospective claims remain marked as such, Phase 18 plans retain named acceptance commands without edits, and the full Go regression suite passes.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Preparatory guidance → automated groundedness scan | A command in research or validation guidance can be mistaken for a current test result even when its named test does not exist yet. |
| Phase evidence → reconciliation ledger | A documentation fix must preserve the requirement evidence chain and must not silently change reconciliation dispositions. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QDJG-01 | Repudiation | Phase 18 validation guidance | medium | mitigate | Replace prospective named test commands in the two preparatory documents with runnable package commands; retain requirement-specific future checks in the execution plans and record the actual full-suite result. |
| T-QDJG-02 | Tampering | Reconciliation evidence | medium | mitigate | Limit edits to the two named Phase 18 documents; preserve CTL/S-010 behavior and axis claims, and leave reconciliation ledgers and groundedness tests untouched. |
| T-QDJG-03 | Tampering | Compiler input and native output | low | accept | This task changes documentation only; it adds no source parser, checker, interpreter, or native-code path. |
</threat_model>

<source_audit>

| Source | ID | Item | Task | Status | Notes |
|--------|----|------|------|--------|-------|
| GOAL | quick | Remove groundedness false positives from Phase 18 preparatory documents | 1 | COVERED | Package-level commands replace future-name filters. |
| REQ | docs-only | Change only research and validation guidance | 1 | COVERED | Files and action constrain edits. |
| REQ | acceptance | Keep precise prospective TestPhase18 acceptance commands in Phase 18 plans | 1 | COVERED | Plans are read for context and checked for no diff. |
| REQ | evidence | Preserve CTL-01/02/03, S-010, and reconciliation truthfulness | 1 | COVERED | Action preserves evidence prose and excludes ledger edits. |
| REQ | regression | Run full Go regression after the fix | 1 | COVERED | `go test ./...` is an automated task gate. |
| RESEARCH | Phase 18 | Four requirement map rows presently filter to future tests | 1 | COVERED | Replace only their command cells and keep Wave 0 status. |
| CONTEXT | user preference | Automate objective acceptance; avoid routine human UAT | 1 | COVERED | All verification is automated. |

</source_audit>

<verification>
The executor verifies that neither document contains a Go test filter for a future name, that Phase 18 plans have no diff, and that the full Go suite passes. The quick summary records actual command results without claiming prospective Phase 18 acceptance has already been achieved.
</verification>

<success_criteria>
The two preparatory documents provide runnable current verification commands without weakening the future Phase 18 acceptance contract or changing the compiler evidence ledger.
</success_criteria>

<output>
Create `.planning/quick/260924-djg-avoid-groundedness-false-positives-from-/260924-djg-SUMMARY.md` when done.
</output>

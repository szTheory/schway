---
phase: quick
plan: 260922-hfs
type: execute
wave: 1
depends_on: []
files_modified:
  - .planning/phases/17-return-type-parameter-type/17-07-PLAN.md
  - .planning/phases/17-return-type-parameter-type/17-VERIFICATION.md
autonomous: true
must_haves:
  truths:
    - "Phase 17 Plan 07 names the consolidated session_phase17_test.go artifact at every repair-corpus and held-out-seal declaration site."
    - "The existing Phase 17 repair-corpus and held-out-seal tests still pass from their consolidated location without moving or duplicating test code."
    - "A fresh Phase 17 verification reports passed with no open gaps after checking the corrected Plan 07 traceability against the repository."
  artifacts:
    - path: ".planning/phases/17-return-type-parameter-type/17-07-PLAN.md"
      provides: "Correct artifact and task-file traceability for the consolidated Phase 17 session tests"
      contains: "internal/compiler/session/session_phase17_test.go"
    - path: ".planning/phases/17-return-type-parameter-type/17-VERIFICATION.md"
      provides: "Fresh Phase 17 re-verification evidence and verdict"
      contains: "status: passed"
  key_links:
    - from: ".planning/phases/17-return-type-parameter-type/17-07-PLAN.md"
      to: "internal/compiler/session/session_phase17_test.go"
      via: "files_modified, must_haves.artifacts, and both task files declarations"
      pattern: "session_phase17_test\\.go"
    - from: ".planning/phases/17-return-type-parameter-type/17-VERIFICATION.md"
      to: ".planning/phases/17-return-type-parameter-type/17-07-PLAN.md"
      via: "fresh artifact and key-link verification after the scoped plan correction"
      pattern: "Re-verification.*Yes"
---

<objective>
Correct Phase 17 Plan 07 traceability to name the consolidated test artifact, then reverify Phase 17.

Purpose: Close the sole formal verification gap without manufacturing a duplicate test file or rewriting execution history.
Output: A scoped Plan 07 metadata correction and a fresh passing Phase 17 verification report.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@.planning/phases/17-return-type-parameter-type/17-07-PLAN.md
@.planning/phases/17-return-type-parameter-type/17-07-SUMMARY.md
@.planning/phases/17-return-type-parameter-type/17-VERIFICATION.md
@internal/compiler/session/session_phase17_test.go
</context>

<tasks>

<task type="auto">
  <name>Task 1: Align Plan 07 with the consolidated repair-test artifact</name>
  <files>.planning/phases/17-return-type-parameter-type/17-07-PLAN.md</files>
  <action>Make a traceability-only edit to Plan 07: update the test-file entry in `files_modified`, the corresponding `must_haves.artifacts` path, and both task-level `files` declarations to `internal/compiler/session/session_phase17_test.go`, which already contains the named repair-corpus and held-out-seal tests and is recorded as the actual modified file in 17-07-SUMMARY.md. Do not move, split, duplicate, or edit test code. Leave 17-07-SUMMARY.md unchanged so the original execution record remains intact, and preserve all other Plan 07 scope, decisions, commands, and acceptance criteria.</action>
  <verify>
    <automated>test "$(rg -c 'internal/compiler/session/session_phase17_test.go' .planning/phases/17-return-type-parameter-type/17-07-PLAN.md)" -eq 4 &amp;&amp; ! rg -q 'session_phase17_repair_test.go' .planning/phases/17-return-type-parameter-type/17-07-PLAN.md &amp;&amp; node ~/.codex/gsd-core/bin/gsd-tools.cjs query frontmatter.validate .planning/phases/17-return-type-parameter-type/17-07-PLAN.md --schema plan | jq -e '.valid == true' &gt;/dev/null &amp;&amp; node ~/.codex/gsd-core/bin/gsd-tools.cjs query verify.plan-structure .planning/phases/17-return-type-parameter-type/17-07-PLAN.md | jq -e '.valid == true' &gt;/dev/null &amp;&amp; go test ./internal/compiler/session -run 'TestPhase17(RepairCorpus|Heldout)' -count=1 -v</automated>
  </verify>
  <done>Plan 07 has exactly four declarations of the real consolidated test path, no stale path declaration, valid plan structure, and passing focused repair-corpus/seal tests; production files and the historical summary are untouched.</done>
</task>

<task type="auto">
  <name>Task 2: Reverify Phase 17 against the corrected traceability</name>
  <files>.planning/phases/17-return-type-parameter-type/17-VERIFICATION.md</files>
  <action>Run a fresh goal-backward Phase 17 verification after Task 1. Recheck every Phase 17 observable truth, required artifact, key link, and TYP-01 through TYP-05 using current repository evidence, with special attention to Plan 07 resolving to the consolidated session test file. Regenerate 17-VERIFICATION.md as a re-verification with a fresh timestamp and evidence-backed verdict. Record `status: passed` and an empty gaps list only when the focused Phase 17 commands pass and the corrected artifact exists; retain precise attribution for any repository-wide failures outside Phase 17 rather than treating them as this phase's failures. Do not edit production code, test code, summaries, requirements, or roadmap state during this task.</action>
  <verify>
    <automated>go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/cgen ./internal/compiler/session ./cmd/lang-repair -run 'TestPhase17|TestUnreachableClaimsGeneratedView|TestRepairDriverDecodesNoProseFields' -count=1 -v &amp;&amp; go test ./internal/compiler/core ./internal/compiler/reduce -run 'TestPhase17|Test.*Return.*Type|Test.*Type.*Fact' -count=1 -v &amp;&amp; test "$(node ~/.codex/gsd-core/bin/gsd-tools.cjs query frontmatter.get .planning/phases/17-return-type-parameter-type/17-VERIFICATION.md --pick status --raw)" = "passed" &amp;&amp; test "$(node ~/.codex/gsd-core/bin/gsd-tools.cjs query frontmatter.get .planning/phases/17-return-type-parameter-type/17-VERIFICATION.md --pick gaps --raw)" = "[]" &amp;&amp; rg -q '\*\*Re-verification:\*\* Yes' .planning/phases/17-return-type-parameter-type/17-VERIFICATION.md &amp;&amp; rg -q 'session_phase17_test.go' .planning/phases/17-return-type-parameter-type/17-VERIFICATION.md</automated>
  </verify>
  <done>The regenerated report is explicitly a re-verification, cites the real consolidated test artifact, reports all Phase 17 truths and TYP requirements satisfied, and carries `status: passed` with no gaps.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Historical plan metadata -> current verifier | The verifier relies on declared artifact paths to decide whether completed work still has its required evidence. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QHFS-01 | Tampering | Phase 17 Plan 07 traceability | medium | mitigate | Limit the edit to four path declarations, validate plan structure, and execute the named tests from the declared file. |
| T-QHFS-02 | Repudiation | Phase 17 execution history | low | accept | Git retains the plan correction history, while the already-correct 17-07-SUMMARY.md remains unchanged as the original execution record. |
</threat_model>

<source_audit>

| Source | ID | Item | Task | Status | Notes |
|--------|----|------|------|--------|-------|
| GOAL | quick | Correct Plan 07 artifact traceability and reverify Phase 17 | 1-2 | COVERED | Scoped metadata correction followed by fresh verification. |
| REQ | TYP-05 | Sealed held-out repair evidence remains verifiable | 1-2 | COVERED | Existing behavior is rerun; no feature scope is changed. |
| RESEARCH | — | No research input | — | EXCLUDED | Internal artifact-path correction uses established project patterns and adds no dependency. |
| CONTEXT | — | No quick-task CONTEXT.md | — | EXCLUDED | The user supplied the complete locked corrective scope directly. |

</source_audit>

<verification>
Validate both edited planning artifacts, run the focused repair-corpus/seal tests and the two established Phase 17 verification commands, and require the regenerated report to identify itself as a passing re-verification with no gaps.
</verification>

<success_criteria>
Plan 07 points to the existing consolidated repair-test file at all four declaration sites; no production, test, summary, requirement, or roadmap file changes; and Phase 17 has a fresh `status: passed` re-verification report with an empty gap list.
</success_criteria>

<output>
Create `.planning/quick/260922-hfs-correct-phase-17-plan-07-traceability-so/260922-hfs-SUMMARY.md` when done.
</output>

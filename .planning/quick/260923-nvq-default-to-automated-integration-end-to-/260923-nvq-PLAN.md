---
phase: quick
plan: 260923-nvq
type: execute
wave: 1
depends_on: []
files_modified:
  - .planning/PROJECT.md
autonomous: true
must_haves:
  truths:
    - "Codename Lang's durable project guidance tells GSD to automate verification wherever practical, with integration, end-to-end, smoke, and seam tests as named evidence forms."
    - "Checks with enough recurring regression value are assigned to CI when that value justifies their maintenance and runtime cost."
    - "Deterministically provable work targets zero human verification or UAT, while irreducible judgment, unavailable external access or hardware, and user-reserved decisions remain valid human handoffs."
    - "The automation preference preserves mandatory workflow gates, user authorization requirements, and explicit human-only acceptance decisions."
    - "Future phase plans must name concrete automated verification commands and place high-value recurring checks in CI."
  artifacts:
    - path: ".planning/PROJECT.md"
      provides: "Durable project-wide verification operating preference for GSD planning and execution"
      contains: "## Verification Operating Preference"
  key_links:
    - from: ".planning/PROJECT.md"
      to: "future GSD phase plans"
      via: "project guidance requiring concrete automated commands and justified recurring CI checks"
      pattern: "concrete automated verification commands"
    - from: ".planning/PROJECT.md"
      to: "human verification and UAT handoffs"
      via: "explicit deterministic-automation target and bounded human-only exceptions"
      pattern: "zero human verification"
---

<objective>
Record Codename Lang's durable preference for automated verification in the project charter.

Purpose: Make GSD consistently prefer deterministic integration, end-to-end, smoke, and seam evidence; retain valuable recurring checks in CI; and reserve human handoff for the cases that genuinely require it without overriding mandatory gates or user authority.
Output: A scoped `.planning/PROJECT.md` policy section that future phase planners and executors can apply directly.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/STATE.md
@AGENTS.md
</context>

<tasks>

<task type="auto">
  <name>Task 1: Add the durable automated-verification operating preference</name>
  <files>.planning/PROJECT.md</files>
  <action>Add one concise `## Verification Operating Preference` section near the project's foundational guidance, before milestone-specific state. State that GSD defaults to automating verification wherever practical and explicitly names integration, end-to-end, smoke, and seam tests as suitable evidence. Require recurring checks to move into CI when their ongoing regression value justifies maintenance and runtime cost. Set the target of zero human verification or UAT for claims that can be established deterministically. Limit human handoff to irreducible judgment, external access or hardware unavailable to the agent, and decisions reserved to the user. Clarify that this preference does not waive mandatory workflow gates, required user authorization, or acceptance decisions explicitly designated as human-only. Direct future phase plans to name the concrete automated commands that prove their acceptance criteria and to arrange for high-value recurring checks to run in CI. Preserve the rest of PROJECT.md verbatim apart from any final-document date annotation needed to identify this preference update.</action>
  <verify>
    <automated>test "$(rg -c '^## Verification Operating Preference$' .planning/PROJECT.md)" -eq 1 &amp;&amp; section="$(awk '/^## Verification Operating Preference$/{capture=1; next} /^## /{if(capture) exit} capture' .planning/PROJECT.md)" &amp;&amp; printf '%s\n' "$section" | rg -qi 'integration.*end-to-end.*smoke.*seam' &amp;&amp; printf '%s\n' "$section" | rg -qi 'CI.*regression value|regression value.*CI' &amp;&amp; printf '%s\n' "$section" | rg -qi 'zero human verification.*UAT|zero human verification or UAT' &amp;&amp; printf '%s\n' "$section" | rg -qi 'irreducible judgment' &amp;&amp; printf '%s\n' "$section" | rg -qi 'external access.*hardware|hardware.*external access' &amp;&amp; printf '%s\n' "$section" | rg -qi 'reserved to the user' &amp;&amp; printf '%s\n' "$section" | rg -qi 'mandatory workflow.*user authorization.*human-only' &amp;&amp; printf '%s\n' "$section" | rg -qi 'concrete automated verification commands' &amp;&amp; printf '%s\n' "$section" | rg -qi 'high-value recurring checks.*CI'</automated>
  </verify>
  <done>PROJECT.md contains exactly one durable verification-preference section covering all requested automation defaults, CI cost/value policy, zero-human-UAT target, permitted human handoffs, preserved gates and authority, and future-plan command/CI requirements; no source, test, configuration, roadmap, or state file is changed.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Project charter -> future GSD plans and execution | Agents treat PROJECT.md as durable authority when choosing verification commands, CI coverage, and human checkpoints. |
| Automated evidence -> human acceptance | Deterministic checks may replace routine UAT, but mandatory authorization and explicitly human-only decisions must remain under user control. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QNVQ-01 | Tampering | Verification operating preference | medium | mitigate | State both the automation default and its preserved mandatory gates in the same durable section, then verify every required clause from that bounded section. |
| T-QNVQ-02 | Repudiation | Future plan verification evidence | low | mitigate | Require future plans to name concrete automated commands and place justified recurring checks in CI so their evidence path is reviewable and repeatable. |
</threat_model>

<source_audit>

| Source | ID | Item | Task | Status | Notes |
|--------|----|------|------|--------|-------|
| GOAL | quick | Record a durable Codename Lang automated-verification preference in PROJECT.md | 1 | COVERED | The single document task promotes the requested preference into project-wide guidance. |
| REQ | quick-automation | Prefer practical integration, end-to-end, smoke, and seam automation; use CI when recurring value justifies cost | 1 | COVERED | Both evidence forms and the CI value/cost rule are required and verified. |
| REQ | quick-human-boundary | Target zero human verification/UAT for deterministic claims while retaining irreducible judgment, unavailable access/hardware, and user-reserved decisions | 1 | COVERED | The section defines the target and all three permitted handoff classes. |
| REQ | quick-gates | Preserve mandatory workflow gates, user authorization, and explicitly human-only acceptance | 1 | COVERED | The preference is framed as a default within existing authority constraints. |
| REQ | quick-future-plans | Future plans name concrete automated commands and ensure high-value recurring checks run in CI | 1 | COVERED | The plan action and automated gate require both directives. |
| RESEARCH | — | No external research input | — | EXCLUDED | This is an internal project-process preference with no new dependency or external integration. |
| CONTEXT | user-request | The user supplied the complete policy and its limits directly | 1 | COVERED | No deferred ideas or unresolved discretion remain. |

</source_audit>

<verification>
Extract only the new PROJECT.md section and require every requested policy clause to be present, including the automation forms, CI value/cost rule, zero-human-UAT target, bounded handoff cases, preserved gates and authority, and future-plan command/CI obligations.
</verification>

<success_criteria>
Codename Lang's PROJECT.md gives future GSD work a durable, precise automation default that is deterministically checkable and retains all mandatory human authority boundaries.
</success_criteria>

<output>
Create `.planning/quick/260923-nvq-default-to-automated-integration-end-to-/260923-nvq-SUMMARY.md` when done.
</output>

---
id: 261002-ahx
phase: quick
plan: 261002-ahx
type: execute
mode: quick-full
status: planned
verification_role: proposal
wave: 1
depends_on: []
files_modified:
  - .planning/PROJECT.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - internal/compiler/native/phase25_utility_test.go
autonomous: true
requirements: [DX-14, EVD-10]
estimate:
  tokens: 6500
  raw_tokens: 6500
  tasks: 1
  confidence: low
must_haves:
  truths:
    - Planning guidance makes early deterministic acceptance checks the default, aims for zero human UAT, and assigns recurring CI checks by regression value versus runtime and maintenance cost.
    - A focused standard-library Go test rejects missing local file targets linked from the Phase 24–25 README Evidence index, including a deliberately broken link in a negative control.
    - The new test runs under the existing native package Go suite used by CI without a new dependency or CI job.
    - Phase 25 validation assigns evidence-index file navigation to automation while retaining first-time clean-checkout comprehension as subjective human judgment.
  artifacts:
    - path: .planning/PROJECT.md
      provides: Durable verification operating preference in its existing section
    - path: internal/compiler/native/phase25_utility_test.go
      provides: Deterministic evidence-index link integrity test and negative control
    - path: .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
      provides: Updated verification ownership for navigation and comprehension
  key_links:
    - from: examples/phase24/README.md#Evidence-index
      to: internal/compiler/native/phase25_utility_test.go#TestPhase25EvidenceIndex
      via: Parse section-local relative Markdown file links and resolve each against the README directory
    - from: internal/compiler/native/phase25_utility_test.go#TestPhase25EvidenceIndex
      to: .github/workflows/ci.yml
      via: Existing go test ./... and go test -race ./... checks include the native package
---

<objective>
Make shift-left acceptance verification the durable planning default and close the mechanical part of the Phase 25 README navigation check.

Purpose: Reviewers can trust that evidence-index file links remain navigable after future edits, while the validation record accurately separates this deterministic property from subjective comprehension.
Output: A scoped PROJECT.md amendment, a focused Go test, and revised Phase 25 validation language.
</objective>

<execution_context>
@/Users/jon/.codex/gsd-core/workflows/execute-plan.md
@/Users/jon/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/PROJECT.md
@.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
@examples/phase24/README.md
@internal/compiler/native/phase25_utility_test.go
@.github/workflows/ci.yml

Planning source inspection: PROJECT.md already has a Verification Operating Preference section; amend it in place. The README Evidence index has local relative Markdown file links. TestPhase25EvidenceIndex currently checks required tokens only. Phase 25 validation calls direct evidence-index navigation a human check. Existing CI executes go test ./... and go test -race ./... on both hosts, so the native package test reaches those lanes without workflow edits. STATE.md says not to run project suites locally in this workspace and records GOCACHE=/tmp/ai-lang-verification-gocache for focused Go checks; use focused verification only during execution unless the active workflow explicitly requires more.

Discovery level 0: existing Go tests, repository path helper, and CI suite establish the implementation pattern. No external API, schema, package install, or new dependency is in scope. The API coverage detector returned detected=false. The assumption-delta detector reported a Phase 25 roadmap sentence about a by-value fallback, which is unrelated to this documentation/test-only quick task and introduces no architecture decision. Schema-relevant ORM file patterns are absent. Estimate calibration returned factor 1, sample_count 0, confidence low.
</context>

<tasks>

<task type="tracer" tdd="true">
  <name>Task 1: Enforce evidence-index link integrity and assign the remaining verification boundary</name>
  <files>.planning/PROJECT.md, .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md, internal/compiler/native/phase25_utility_test.go</files>
  <read_first>Read the existing PROJECT.md Verification Operating Preference, README Evidence index, TestPhase25EvidenceIndex, Phase 25 Manual-Only Verifications, and .github/workflows/ci.yml Go-suite steps. Preserve other planning history and Phase 25 hosted receipt statements.</read_first>
  <behavior>
    - Current README Evidence index passes when each local relative Markdown file link resolves to an existing file relative to examples/phase24/.
    - A test-local mutated index containing a missing relative file target fails the same validation helper with a diagnostic naming that target; this is a negative control, not a repository edit.
    - The check is scoped to the Evidence index and rejects an empty or linkless section so accidental section removal cannot pass vacuously.
  </behavior>
  <action>Extend the existing TestPhase25EvidenceIndex with a compact Go-standard-library helper that locates the heading-bounded Evidence index section, extracts relative Markdown file-link destinations there, strips an optional fragment from the file path, and resolves each path from the README directory using filepath.Clean and os.Stat. Ignore external and pure-fragment links if encountered; require at least one local file link and a regular-file target. Keep the current token assertions. Add a table/subtest or direct negative control using an in-memory broken destination to prove failure detection. Keep parsing narrowly tied to this README's ordinary inline Markdown links; do not add a Markdown dependency or generic parser. Amend the existing PROJECT.md preference to say acceptance criteria become early deterministic unit, seam, smoke, integration, or end-to-end checks with relevant failure controls; recurring CI inclusion is justified by regression value versus runtime/maintenance cost; aim for zero human UAT and hand off only genuinely subjective, external, or user-authority properties lacking repeatable proxies; favor standard-library/local code and no new dependency by default. Preserve its existing stale-report and authorization guidance. In Phase 25 validation, update the TestPhase25EvidenceIndex row to name file-link integrity, and revise Manual-Only Verifications so direct link navigation is owned by that automated check while first-time clean-checkout instruction comprehension remains explicitly subjective human review. Do not claim link resolution proves human comprehension or alter historical hosted results. The new native test is automatically included in existing CI Go suites; no CI file change is needed.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^TestPhase25EvidenceIndex$' ./internal/compiler/native</automated>
    <fails_when>The focused go test command exits nonzero, including when the link-missing negative control fails to detect its deliberately broken destination or its assertion logic is broken.</fails_when>
  </verify>
  <done>The focused test passes for the checked-in README and contains a negative control that catches a missing Evidence index target; the test runs via existing native-package CI suites. PROJECT.md states the durable shift-left preference in its existing section. The Phase 25 validation row and manual-only paragraph assign navigation to automation and first-time comprehension to subjective review, with no dependency or generic framework added.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Authored README to repository files | Relative links are editor-controlled text and can become stale after file moves. |
| Validation claim to CI evidence | Documentation must identify only properties the Go suite actually checks. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-ahx-01 | Tampering | README Evidence index | low | mitigate | Task 1 resolves section-local file links against the README directory and includes a broken-link negative control. |
| T-ahx-02 | Repudiation | Phase 25 validation record | low | mitigate | Task 1 names the automated test in the verification map and retains subjective comprehension as a separate manual criterion. |
</threat_model>

<verification>
The executor runs the focused native Go test, records the outcome in the quick summary, and confirms the existing CI go test ./... step includes this package. This plan does not request a local full-suite run or a fresh hosted receipt.
</verification>

<success_criteria>
- The existing PROJECT.md section states the requested verification posture and dependency preference without duplication.
- Every relative file link in the README Evidence index resolves, and a test-local missing target is rejected.
- Phase 25 validation distinguishes automated navigation integrity from subjective first-time comprehension.
- The focused Go test passes; no new dependency, CI job, or production behavior is introduced.
</success_criteria>

<source_audit>
Quick-scope GOAL: durable shift-left default and deterministic Phase 25 evidence-index navigation — Task 1. REQ: DX-14 clean-checkout evidence location and EVD-10 evidence navigation — Task 1. RESEARCH: existing native test and CI suite patterns, standard-library-only constraint — Task 1. CONTEXT: amend the existing preference; aim for zero human UAT; early deterministic checks with cost/value CI threshold; preserve subjective comprehension; no new dependency/framework — Task 1. No requested item is unplanned, and no deferred Phase 25 capability is introduced.
</source_audit>

<output>
Create .planning/quick/261002-ahx-record-shift-left-verification-as-the-de/261002-ahx-SUMMARY.md after execution with focused test results and the exact documentation changes. The orchestrator owns summary/state publication; this planning pass writes only this PLAN.md and does not commit it.
</output>

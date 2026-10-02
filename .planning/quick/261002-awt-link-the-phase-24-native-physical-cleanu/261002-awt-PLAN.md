---
id: 261002-awt
phase: quick
plan: 261002-awt
type: execute
mode: quick-full
status: planned
verification_role: proposal
wave: 1
depends_on: []
files_modified:
  - examples/phase24/README.md
  - internal/compiler/native/phase25_utility_test.go
autonomous: true
requirements: [EVD-09, RES-06]
estimate:
  tokens: 4500
  raw_tokens: 4500
  tasks: 1
  confidence: low
must_haves:
  truths:
    - The Phase 24 physical-cleanup claim links directly to the native observer source and the Phase 24 validation record, with hosted run 36856048690 visible.
    - TestPhase25EvidenceIndex fails if either required Markdown link is removed, changed, or points to a missing or non-file local target.
    - The existing Phase 25 evidence-index checks and historical Phase 24 receipt remain intact.
  artifacts:
    - path: examples/phase24/README.md
      provides: Direct observer and hosted validation links beside the physical-cleanup claim
    - path: internal/compiler/native/phase25_utility_test.go
      provides: Focused standard-library contract for both required links and resolved targets
  key_links:
    - from: examples/phase24/README.md
      to: internal/compiler/native/phase24_observer_test.go
      via: Relative Markdown link beside the native observer claim
    - from: examples/phase24/README.md
      to: .planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md
      via: Relative Markdown link to hosted run 36856048690's recorded receipt
---

<objective>
Make the Phase 24 physical-cleanup claim directly navigable to its owning native observer and hosted validation receipt, with both links pinned by the existing README contract test.

Purpose: A reader can inspect the physical observer and the two-host receipt from the claim itself; a broken or removed evidence link fails the focused CI-covered test.
Output: Two README links and one focused test amendment.
</objective>

<execution_context>
@/Users/jon/.codex/gsd-core/workflows/execute-plan.md
@/Users/jon/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/PRODUCT-ROADMAP.md
@.planning/LANGUAGE-MATURITY.md
@examples/phase24/README.md
@internal/compiler/native/phase25_utility_test.go
@internal/compiler/native/phase24_observer_test.go
@.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md

Source inspection: README lines 69–73 link only the transfer and error programs beside the native physical-cleanup claim. The owner tests are TestPhase24ObserverPublicLifecycleAndTypedError and TestPhase24ObserverReachedPhysicalDestructorControls. Phase 24 validation records hosted run 36856048690 on Ubuntu and macOS. TestPhase25EvidenceIndex already reads this README, and phase25CheckEvidenceIndexLinks checks local files only under the separate Evidence index heading. The working tree has an unrelated change to Phase 25 verification; preserve it.

Discovery level 0: the existing README and Go standard-library test supply the pattern; no new package, API integration, schema file, or runtime behavior is involved. API coverage detector returned detected=false. Assumption-delta query skipped with phase_unresolved because this is a quick task, so no architectural finding is asserted. The active schema-push pattern scan found no matching files. Estimate calibration factor 1, sample_count 0, confidence low. PRODUCT-ROADMAP and LANGUAGE-MATURITY retain their ranked Phase 26 computation, bounded checksum, and consumer-led JSON thresholds; this evidence-navigation correction changes none of those facts.
</context>

<tasks>

<task type="tracer">
  <name>Task 1: Link the physical observer and hosted receipt, then pin the links</name>
  <files>examples/phase24/README.md, internal/compiler/native/phase25_utility_test.go</files>
  <read_first>Read the Phase 24 transfer and cleanup paragraph, TestPhase25EvidenceIndex and its existing link helper, the two named observer tests, and the Final Phase 24 Hosted Receipt section of 24-VALIDATION.md.</read_first>
  <action>In the Phase 24 transfer and cleanup paragraph, make “native observer” a direct relative Markdown link to ../../internal/compiler/native/phase24_observer_test.go and add a direct relative Markdown link to ../../.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md beside the physical-cleanup claim. Identify the validation link as the hosted receipt for run 36856048690, retaining the distinction between physical observer evidence and model-only replay. In TestPhase25EvidenceIndex, require both exact Markdown label/destination pairs in this paragraph, resolve each destination relative to examples/phase24, and require os.Stat to find a regular file. Read the linked validation record and require the existing hosted run ID 36856048690 so a misplaced receipt is caught. Keep the existing Evidence index checker and broken-link negative control. Use the existing standard-library imports and focused test; do not broaden the Markdown parser or change the observer, validation record, CI, or other README sections.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test -v -count=1 -run '^TestPhase25EvidenceIndex$' ./internal/compiler/native</automated>
    <fails_when>non-zero exit, or output lacks "--- PASS: TestPhase25EvidenceIndex"</fails_when>
  </verify>
  <done>The claim has both direct local links and names hosted run 36856048690; TestPhase25EvidenceIndex requires each link, proves both targets are regular files, and confirms the receipt ID in the validation record. Its focused command passes, and the change is confined to the two authorized files.</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| README claim to repository evidence | Authored Markdown points reviewers to local source and a historical hosted receipt; no new runtime or security boundary is introduced. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-awt-01 | Tampering | Phase 24 evidence links | low | mitigate | Task 1 pins the exact links, checks regular-file targets, and confirms the historical run ID through the focused contract test. |
</threat_model>

<verification>
Run only the focused TestPhase25EvidenceIndex command. Existing CI's go test ./... includes internal/compiler/native; no full-suite rerun or human UAT is needed for this objective documentation contract. The hosted Phase 24 run is historical evidence, not a new execution receipt.
</verification>

<success_criteria>
- Both evidence targets are directly linked from the Phase 24 physical-cleanup paragraph, and run 36856048690 remains explicit.
- The focused test catches missing or mistyped link literals, missing or non-file targets, and a validation record without the receipt ID.
- The existing Evidence index check and its negative control pass unchanged.
</success_criteria>

<source_audit>
GOAL: README claim-to-evidence navigation and a deterministic regression guard are both Task 1. REQ: EVD-09 physical-observer evidence and RES-06 typed-error cleanup are linked, without changing their implementation. RESEARCH: no external dependency or integration is required; the existing Go test and local validation record provide the implementation pattern. CONTEXT: both named observer tests, hosted run 36856048690, focused standard-library verification, no broad parser/full suite/UAT, and exact two-file edit scope are Task 1. No quick-scope item remains unplanned.
</source_audit>

<output>
Create .planning/quick/261002-awt-link-the-phase-24-native-physical-cleanu/261002-awt-SUMMARY.md after execution with the focused test outcome. The orchestrator owns summary and state publication; this planning pass writes only this PLAN.md.
</output>

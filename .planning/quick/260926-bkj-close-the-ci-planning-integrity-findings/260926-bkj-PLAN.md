---
id: 260926-bkj
phase: quick
plan: 260926-bkj
type: execute
mode: quick
status: complete
wave: 1
depends_on: []
files_modified:
  - .planning/ROADMAP.md
  - .planning/STATE.md
  - .planning/state.json
  - .planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md
  - .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-PLAN.md
  - .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-SUMMARY.md
autonomous: true
must_haves:
  truths:
    - "Phase 15's completed 12/12 automated UAT remains unchanged and is not presented again to a human."
    - "Phase 15's verification report is fresh against the post-Phase-21 shared compiler code and reports the current automated evidence accurately."
    - "The one-shot tagged LTO comparison remains recorded as bounded evidence while recurring CI checks the untagged receipt-binding guard."
    - "The M004 roadmap states that Phases 14-20 in M003 complete before Phase 21 starts."
    - "Future GSD work defaults to recurring, machine-checkable verification where it has recurring value, and does not repeat a routed command without changed evidence or state."
    - "The four planning-integrity guards and Phase 15-specific evidence selectors pass after the documentation corrections."
  artifacts:
    - path: .planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md
      provides: Fresh Phase 15 goal verification and current content fingerprint
    - path: .planning/STATE.md
      provides: Durable shift-left, zero-routine-UAT, and no-loop workflow defaults with a live next action
    - path: .planning/ROADMAP.md
      provides: Explicit M003 completion prerequisite for Phase 21
    - path: .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md
      provides: Correct distinction between one-shot tagged evidence and recurring CI evidence
  key_links:
    - from: .planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md
      to: .planning/phases/15-event-identity-lang-execution-2/15-UAT.md
      via: Preserve completed automated UAT as the acceptance record; refresh only verifier evidence
    - from: .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md
      to: .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md
      via: Keep the measured tagged comparison in its scoped receipt and use an untagged binding test for recurring CI

<objective>
Refresh Phase 15 verification without repeating completed UAT, repair two CI-detected planning evidence defects, and record a durable GSD shift-left/no-loop policy.

Purpose: Keep objective recurring evidence automated while limiting human verification to criteria that cannot be checked reliably by software.
Output: Fresh Phase 15 verification, corrected planning evidence, durable workflow guidance, and a quick-task summary of the observed test results.
</objective>

<tasks>

<task type="auto">
  <name>Task 1: Correct CI evidence lane and M004 ordering statement</name>
  <files>.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md, .planning/ROADMAP.md</files>
  <action>In Phase 21's Behavioral Spot-Checks, leave the actual tagged LTO run and its bounded result in 21-LTO-EVIDENCE.md and the existing report narrative, but make the recurring test row invoke the untagged TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison guard that binds that receipt to its tagged source. Keep the comparison one-shot and keep all evidence limitations. In ROADMAP.md's milestone list, explicitly state the exact prerequisite that M003 Phases 14-20 complete before M004 Phase 21 starts. Do not change the groundedness implementation, reconciliation ledger, Phase 16 debt table, or any source code.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison$' -count=1 -v</automated>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase16EmitterCutsAreAmendedAndOwned$' -count=1 -v</automated>
  </verify>
  <done>The recurring report row names the untagged receipt-binding test, the one-off measured command remains accurately recorded, and the Phase 21 owner/prerequisite guard passes.</done>
</task>

<task type="auto">
  <name>Task 2: Refresh Phase 15 verifier evidence from current automated tests</name>
  <files>.planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md</files>
  <action>Keep 15-UAT.md byte-for-byte unchanged. Re-run the planning-integrity guards discovered by the full suite and the Phase 15 deterministic execution, native admission, interpreter, peer, cgen, CI aggregate, legacy-wrapper, diamond, and collision selectors from the existing report. After the relevant checks pass, recompute covered_files/covered_digest using gsd-tools verification.fingerprint; update the verifier timestamp and add a concise refresh note identifying the four shared cgen/interpreter files changed by Phase 21 and the current bounded test evidence. Preserve the existing 7/7 goal conclusions only when their current evidence supports them; record any repository-wide issue outside Phase 15 without misclassifying it as a Phase 15 failure.</action>
  <verify>
    <automated>go test ./internal/compiler/session -run 'TestReconciliationVerdictsCarryTheirObligations|TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundednessThreeClassesAreEmpty|TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison' -count=1 -v</automated>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./...</automated>
    <automated>go test ./internal/compiler/execution ./internal/compiler/native ./internal/compiler/interp ./internal/compiler/executionpeer ./internal/compiler/cgen -run 'TestInvocationGrammar|TestExecutionLegacyBytesFrozen|TestValidateExecutionSchema2|TestInvocationThreadsThroughAllEventPaths|TestExecutionSchemaSelectionPreservesLegacy|TestFunctionCalledPreorderAndOwnership|TestFunctionCalledProjectionRemoval|TestRejectedCallEmitsNoCalledEdge|TestIndependentEntryResolution|TestInvocationMembershipTraversal|TestExecutionPeerImportBoundary|TestFullCoverageControlIsSeparate|TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable|TestInvocationPathTableDeepDiamondMeasures61|TestInvocationPreflightOrdering|TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestParentIndexedChildLookup|TestInvocationTableEmissionIsDeterministic|TestProgramWritesExecutionSchema2|TestNativeFunctionCalledPreorder|TestNativeFunctionCalledProjectionRemoval|TestLegacyEventWritersFrozen' -count=1 -v</automated>
    <automated>go test ./internal/compiler/session -run 'TestSchema2ComparisonRequiresPeerVerdict|TestPhase5CompareProgramEnginesPreservesLegacySchemas|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert|TestCIWorkflowRunsCurrentAggregateGate|TestCIWorkflowSelectionPinsPackageOwnership' -count=1 -v</automated>
    <automated>node "$HOME/.codex/gsd-core/bin/gsd-tools.cjs" query verification.status .planning/phases/15-event-identity-lang-execution-2 --raw</automated>
  </verify>
  <done>Phase 15's live verifier status is passed with a current content fingerprint, its 12 automated UAT passes remain intact, and no human checkpoint was introduced.</done>
</task>

<task type="auto">
  <name>Task 3: Record durable shift-left defaults and the non-loop route</name>
  <files>.planning/STATE.md, .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-SUMMARY.md</files>
  <action>Add durable GSD guidance to STATE.md: prefer deterministic machine-checkable acceptance and recurring CI evidence when repeated value justifies its cost; use appropriate unit, seam, smoke, integration, or e2e checks; reserve human UAT for subjective, external, or otherwise non-automatable criteria. For completed UAT plus stale verification, never rerun verify-work unchanged: inspect init.execute-phase, use execute-phase only when it has zero incomplete plans so it routes directly to the verifier, and re-query init.progress after the report refresh. Replace the stale next-action pointer with the resolver-backed next command. Record this quick task and its actual pass/fail/cache evidence in SUMMARY.md and STATE.md.</action>
  <verify>
    <automated>git diff --check</automated>
    <automated>node "$HOME/.codex/gsd-core/bin/gsd-tools.cjs" query init.progress --raw</automated>
  </verify>
  <done>STATE.md contains durable shift-left/no-human-UAT-by-default and no-loop instructions, plus the exact resolver-backed next GSD action and a task-index row.</done>
</task>

</tasks>

<verification>
The quick summary records the initial full-suite result, the four exact planning-integrity failures, the sandbox Go-cache adjustment, focused post-fix guards, Phase 15 selectors, and the refreshed fingerprint/status. The completed UAT remains 12/12 automated.
</verification>

<success_criteria>
No completed UAT is repeated, current Phase 15 automated evidence is fresh, planning-integrity CI passes its exact findings, and future GSD runs have a durable shift-left/no-loop default.
</success_criteria>

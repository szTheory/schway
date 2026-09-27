---
id: 260926-uzs
phase: quick
plan: 260926-uzs
type: execute
mode: quick
status: complete
wave: 1
depends_on: []
files_modified:
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/
  - .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/
  - internal/compiler/session/session_phase21_contract_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
  - .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-PLAN.md
  - .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-SUMMARY.md
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
  - .planning/ROADMAP.md
  - .planning/STATE.md
autonomous: true
must_haves:
  truths:
    - "All six completed Phase 21 plans and summaries, the passed verification, and all seven completed UAT cases reside under M004-phases; no Phase 21 directory remains for M003 cleanup to capture."
    - "Every operational reader of Phase 21 contract, LTO receipt, verification report, and M003 debt resolves the archived path while preserving the existing refusal and bounded-evidence claims."
    - "Groundedness frontier and ownership records match the measured moved corpus; the checked-in validation record is a real completed run over the consumer's current pair request with matching digests."
    - "Phase 21 verification is fresh for the archived file inventory and the original UAT is byte-for-byte unchanged."
    - "ROADMAP and STATE report 6/6 plans, 7/7 UAT, 6/6 verification, provisional M004 scope, and the exact next command `$gsd-new-milestone \"Native Emission Ownership and Resource Discharge\"`."
  artifacts:
    - path: .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/
      provides: Complete Phase 21 execution, UAT, and verification archive
    - path: internal/compiler/session/session_phase21_contract_test.go
      provides: Live archive path to checked contract and LTO receipt
    - path: internal/compiler/session/verification_groundedness_test.go
      provides: Measured path-specific frontier and R2b ownership
    - path: testdata/phase16/validation-corpus-run-record.manifest.json
      provides: Digest and completion envelope for regenerated JSONL evidence
    - path: .planning/STATE.md
      provides: Accurate milestone handoff and next command
    - path: .planning/ROADMAP.md
      provides: Completed provisional Phase 21 entry and next action
  key_links:
    - from: internal/compiler/session/session_phase21_contract_test.go
      to: .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json
      via: testsupport.ProjectPath archive lookup
    - from: internal/compiler/session/verification_groundedness_test.go
      to: .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md
      via: Full file, line, command, classification finding identity and P21 landing
    - from: testdata/phase16/validation-corpus-run-record.manifest.json
      to: testdata/phase16/validation-corpus-run-record.jsonl
      via: Exact consumer pair digest, raw record digest, and completion witnesses
    - from: .planning/ROADMAP.md
      to: .planning/STATE.md
      via: Same completed Phase 21 status and M004 kickoff command
---

<objective>
File completed Phase 21 under the provisional M004 archive and repair all live path, evidence, verification, and handoff records affected by the move.

Purpose: Prevent the next milestone command from misarchiving Phase 21 under M003 while keeping its existing checks honest.
Output: A complete M004 phase archive, passing archive-sensitive guards, fresh Phase 21 verification, and a single kickoff command.
</objective>

<tasks>

<task type="auto">
  <name>Archive Phase 21 and retarget its operational references</name>
  <files>.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/, .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/, internal/compiler/session/session_phase21_contract_test.go, .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md, .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-PLAN.md, .planning/quick/260926-bkj-close-the-ci-planning-integrity-findings/260926-bkj-SUMMARY.md</files>
  <action>Before moving or committing the directory, record the pre-move UAT blob ID with `git rev-parse HEAD:.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-UAT.md`; compare it with `git hash-object` on the archived UAT during this task's verification, before the move commit changes HEAD. Move the entire Phase 21 directory to `.planning/milestones/M004-phases/` as one archive unit. Inventory every reference to its former `.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/` path before editing. Retarget operational links in the moved plans, summaries, verifier inventory, review metadata, M003 Phase 14 debt register, and earlier quick-task artifacts when they identify a current file. Preserve the six plan and summary records, their substantive findings, the contract JSON and bounded LTO receipt; keep `21-UAT.md` byte-for-byte identical to its pre-move Git blob. Update both ProjectPath reads in `session_phase21_contract_test.go` to the M004 archive. Leave historical source-path statements that explicitly describe where work originally occurred only if they are clearly marked as history; do not leave a live path pointing to a nonexistent file. Do not change compiler production behavior or expand ownership/resource admission.</action>
  <verify>
    <automated>test -d .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004</automated>
    <automated>test ! -e .planning/phases/21-native-emission-ownership-and-resource-discharge-m004</automated>
    <automated>test "$(rg --files .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004 | rg -c '/21-[0-9][0-9]-PLAN[.]md$' | awk -F: '{s+=$NF} END {print s+0}')" -eq 6</automated>
    <automated>test "$(rg --files .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004 | rg -c '/21-[0-9][0-9]-SUMMARY[.]md$' | awk -F: '{s+=$NF} END {print s+0}')" -eq 6</automated>
    <automated>test "$(git hash-object .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-UAT.md)" = "$(git rev-parse HEAD:.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-UAT.md)"</automated>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase21ResourceDischargeContract|TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison)$' -count=1</automated>
  </verify>
  <done>The complete directory is in M004-phases, the old directory is absent, all six summary files and UAT remain present, UAT matches its pre-move Git blob, and both contract/receipt guards resolve the archive.</done>
</task>

<task type="auto">
  <name>Reconcile moved evidence and update verification findings</name>
  <files>internal/compiler/session/verification_groundedness_test.go, .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md, testdata/phase16/validation-corpus-run-record.jsonl, testdata/phase16/validation-corpus-run-record.manifest.json, .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md</files>
  <action>Measure the groundedness corpus after the move. Re-key only the affected pinnedFrontier and r2bLandingPhases records by full file/line/command/classification identity, retaining existing R1/R2/R3 zero gates, R2b obligations, ownership vocabulary, stale-entry rejection, and corpus floors. Reconcile any affected M003 Phase 14 debt citation from the measured identity. Export the exact current package/pattern request through TestExportValidationCorpusPairs with AI_LANG_EVIDENCE_PAIR_OUTPUT, pass each pair as discrete argv to `scripts/evidence-run-record.sh` with `GOCACHE=/tmp/ai-lang-verification-gocache`, and replace the checked-in JSONL only with actual producer output. Refresh the manifest from exported pair bytes, raw record bytes, measured elapsed time, UTC timestamp, and current revision; keep the existing consumer's digest, completion, tamper, and grade laws. Update Phase 21 verifier path inventory and factual evidence narrative while preserving the original seven UAT outcomes and bounded LTO claims. The final fingerprint follows handoff edits in Task 3. Do not recast historical raw log lines as current paths.</action>
  <verify><automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestValidationCorpusPairExportMatchesConsumer|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestCheckedInCorpusRecordRejectsTamperingAndVacuity|TestVerificationGroundednessThreeClassesAreEmpty|TestVerificationGroundednessFrontierIsPinned|TestReconciliationVerdictsCarryTheirObligations|TestEvidenceReconciliationViewIsCurrent)$' -count=1</automated></verify>
  <done>The consumer accepts the genuine completed record and rejects stale/tampered controls; measured groundedness matches its pinned frontier and ownership; the verifier records the archived file inventory.</done>
</task>

<task type="auto">
  <name>Correct the M004 kickoff handoff</name>
  <files>.planning/ROADMAP.md, .planning/STATE.md, .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md</files>
  <action>Replace stale gap execution and pre-archive directions with the completed Phase 21 result: 6/6 plans, 7/7 preserved UAT, and 6/6 current verification. Point ROADMAP to the M004 phase archive and remove the misleading M003 roadmap link from its provisional M004 milestone row; do not invent an M004 roadmap or ratified requirement IDs. Retain M003 as the shipped outgoing milestone and explain that Phase 21 is already filed in M004-phases, outside outgoing cleanup. Update STATE's current focus, M004 handoff, Session Continuity next command, and Operator Next Steps to the exact `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"`. After ROADMAP and STATE edits, recompute Phase 21 `covered_files` and `covered_digest` through `gsd-tools query verification.fingerprint` for the archived directory, refresh the verifier timestamp, and query `verification.status` and `init.progress`. If the router still treats the archived phase as in progress due resolver limitations, record that specific discrepancy without changing the verified facts or routing back to completed UAT/gap plans.</action>
  <verify>
    <automated>git diff --check</automated>
    <automated>rg -Fq '$gsd-new-milestone "Native Emission Ownership and Resource Discharge"' .planning/ROADMAP.md</automated>
    <automated>rg -Fq '$gsd-new-milestone "Native Emission Ownership and Resource Discharge"' .planning/STATE.md</automated>
    <automated>node ~/.codex/gsd-core/bin/gsd-tools.cjs query verification.status .planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004 --raw</automated>
    <automated>node ~/.codex/gsd-core/bin/gsd-tools.cjs query init.progress --raw</automated>
  </verify>
  <done>ROADMAP and STATE agree on the completed archived phase and exact kickoff command; M004 remains provisional, and any remaining router discrepancy is explained from fresh query output.</done>
</task>

</tasks>

## Source coverage audit

| Source | Item | Coverage |
|--------|------|----------|
| GOAL | Archive completed Phase 21 under M004 before kickoff and preserve current checks | Tasks 1–3 |
| REQ | M004 has no ratified requirements or Phase 21 requirement IDs | Excluded; no requirement invented |
| RESEARCH | Existing Phase 21 contract, refusal boundaries, bounded LTO receipt, evidence-grade producer/consumer, and groundedness controls | Tasks 1–2 preserve and rebind existing artifacts |
| CONTEXT | D-21-01/D-21-02 structural discharge and refusal, D-21-03 bounded emitted-fixture evidence | Tasks 1–2 preserve contract and receipt without new compiler claims |
| HANDOFF | Completed 6/6 plans, 7/7 UAT, 6/6 verification; physical directory still in `.planning/phases/`; exact M004 kickoff command | Tasks 1 and 3 |

<threat_model>
## ASVS level and blocking threshold

ASVS Level 1; high-severity threats block completion.

## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Archived files to compiler tests | Test consumers trust file paths and contents under `.planning/milestones/`. |
| Producer JSONL to evidence grade | Recorded test results must correspond to exact consumer requests and completion witnesses. |
| Measured groundedness to exception ledger | Only fully identified findings may receive historical ownership or reconciliation. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-UZS-01 | Tampering | Archive migration | high | mitigate | Compare UAT with the pre-move Git blob and retain all six plans and summaries; retarget operational references only. |
| T-UZS-02 | Repudiation | Evidence run record | high | mitigate | Re-export consumer pairs, run the existing producer, bind exact pair and raw-record digests, and check completion witnesses. |
| T-UZS-03 | Elevation of privilege | Groundedness exceptions | high | mitigate | Match findings by complete identity and retain both-direction stale-entry and owner checks. |
| T-UZS-04 | Information disclosure | LTO receipt | low | accept | Existing receipt contains bounded compiler/host measurements and no secret material; preserve its claim limits. |
</threat_model>

<verification>
Run the focused guards in task order, then check the archived UAT against its pre-move Git blob, all six summaries, the old-directory absence, current `verification.status`, and the exact kickoff command in both handoff files. Preserve the measured evidence record and any current router discrepancy in the quick-task summary at execution time.
</verification>

<success_criteria>
Phase 21 is completely archived under M004, all archive-sensitive checks remain grounded and pass, the verifier is fresh, and the handoff points only to `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"`.
</success_criteria>

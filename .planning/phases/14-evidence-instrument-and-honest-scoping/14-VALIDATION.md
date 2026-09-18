---
phase: "14"
slug: "evidence-instrument-and-honest-scoping"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-17"
evidence_vocabulary: v1
graded_rows: 39
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — `go.mod` at repo root |
| **Quick run command** | `go test ./<changed-package>/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~192.7s full suite (recorded baseline; EVD-06 re-measures and records it as an `observed` row — do NOT copy the stale ~60s figure from docs) |

**Baseline:** `go test ./...` is green (exit 0, 25 packages) on the current tree
before any Phase 14 change — confirmed during research. Any red is caused by
this phase.

---

## Sampling Rate

- **After every task commit:** Run the quick command for the touched package
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 200 seconds (full suite); < 10s for a single package

---

## Per-Task Verification Map

> Seeded by `/gsd-plan-phase`; the planner fills one row per task from PLAN.md.
> **Phase-specific law (the whole point of this phase):** every `Automated Command`
> in this table MUST be executed and MUST resolve to at least one test before it is
> recorded here. A `go test -run` pattern that matches zero tests is the exact defect
> EVD-01 exists to catch — authoring one here would be self-refuting.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| 14-01-T1 | 01 | 1 | EVD-01 | — | End-to-end groundedness lint over one Tier-A document class | unit | `go test ./internal/compiler/session/... -run 'TestVerificationGroundedness$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-01-T2 | 01 | 1 | EVD-01 | — | Pin the measured violation frontier as an exact committed literal | unit | `go test ./internal/compiler/session/... -run 'TestVerificationGroundednessFrontierIsPinned' -count=1 -v` | ✅ exists | WIRED | — |
| 14-01-T3 | 01 | 1 | EVD-01 | — | Prove the lint is not inert with a seeded dead pattern | unit | `go test ./internal/compiler/session/... -run 'TestVerificationGroundednessIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-02-T1 | 02 | 2 | DX-08 | — | Build the distinctness corpus and freeze the pre-fix collision control | unit | `go run ./cmd/lang --json check testdata/distinctness/spiral_full.lang` | ✅ exists | REACHABLE | — |
| 14-02-T2 | 02 | 2 | DX-08 | — | Attach the discarded recovery extent as a skipped_region cause | unit | `go test ./internal/compiler/syntax/... ./internal/compiler/diagnostic/... ./internal/compiler/check/... -count=1` | ✅ exists | WIRED | — |
| 14-02-T3 | 02 | 2 | DX-08 | — | Ship the distinctness gate with its corpus predicate and frozen control | unit | `go test ./internal/compiler/check/... -run 'TestDistinctnessCorpusMembersAreStructurallyDistinct\|TestDiagnosticDistinctnessIsOne\|TestDiagnosticDistinctnessGuardIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-03-T1 | 03 | 2 | DX-09 | — | Carry the decline reason through selectRepair onto the Outcome envelope | unit | `go test ./cmd/lang-repair/... -count=1` | ✅ exists | WIRED | — |
| 14-03-T2 | 03 | 2 | DX-09 | — | Guard that no unrepairable outcome ever carries an empty diagnosis | unit | `go test ./cmd/lang-repair/... -run 'TestUnrepairableAlwaysCarriesDiagnosis' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-03-T3 | 03 | 2 | DX-09 | — | Prove the diagnosis guard is not inert with a stand-in binary | unit | `go test ./cmd/lang-repair/... -run 'TestUnrepairableDiagnosisGuardIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-04-T1 | 04 | 3 | PRC-01 | — | Close the owning-phase vocabulary inside checkDebtRegister | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-04-T2 | 04 | 3 | PRC-01 | — | Migrate the twelve registers, open PHASE-14-DEBT.md | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed' -count=1` | ✅ exists | EXERCISED | — |
| 14-04-T3 | 04 | 3 | PRC-01 | — | Prove the ownership gate fails on a seeded ownerless row | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegisterOwnershipGuardIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-05-T1 | 05 | 3 | EVD-06 | — | Machine-check LANGUAGE-MATURITY.md's counts by independent re-derivation | unit | `go test ./internal/compiler/session/... -run 'TestLanguageMaturityCountsAreCurrent' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-05-T2 | 05 | 3 | EVD-08 | — | Measure the suite wall-clock cold and record an observed manifest row | unit | `go test ./internal/compiler/session/... -run 'TestBudgetManifestLoads\|TestQLT02BudgetManifestFileUnchangedDuringAudit' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-05-T3 | 05 | 3 | EVD-06 | — | Prove both self-checks are not inert | unit | `go test ./internal/compiler/session/... -run 'TestSelfDescribingDocsGuardIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-06-T1 | 06 | 4 | EVD-01 | — | Scope by illocutionary role, with promotion and no filename escape | unit | `go test ./internal/compiler/session/... -run 'TestVerificationGroundedness' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-06-T2 | 06 | 4 | EVD-01 | — | Execute grep-shaped commands, detect and pin per-branch groundedness | unit | `go test ./internal/compiler/session/... -run 'TestVerificationGroundedness' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-06-T3 | 06 | 4 | EVD-01 | — | Install corpus floors, index accuracy control, re-measured frontier pin | unit | `go test ./internal/compiler/session/... -run 'TestStaticTestIndexMatchesGoTestList' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-07-T1 | 07 | 5 | EVD-03 | — | Close the witness grammar and escape registry inside the register law | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-07-T2 | 07 | 5 | EVD-03 | — | Write the executed probes and populate the day-one register rows | unit | `go test ./internal/compiler/session/... -run 'TestB1BlameIsStructurallyUnreachable\|TestD1243ControlIsUnconstructible\|TestPhase6HeldoutPairsAreAlphaRenamesOnly' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-07-T3 | 07 | 5 | EVD-04 | — | Enumerate every suppression surface and require a resolvable witness | unit | `go test ./internal/compiler/session/... -run 'TestNoSuppressionOutlivesItsWitness' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-07-T4 | 07 | 5 | EVD-03 | — | Generate UNREACHABLE-CLAIMS.md as a byte-compared view | unit | `go test ./internal/compiler/session/... -run 'TestUnreachableClaimsViewIsCurrent' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-08-T1 | 08 | 6 | EVD-05 | — | Collapse the two axis-movement implementations into one | unit | `go test ./internal/compiler/session/... -run 'TestEveryMutationMovesItsClaimedAxis\|TestAssertMutationMovesAnAxisFailsOnAMislabeledAxis\|TestNAT03MutationTableHasSevenRows' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-08-T2 | 08 | 6 | EVD-05 | — | Remove the per-row exclusion, admit the unsubjected row with an escape | unit | `go test ./internal/compiler/session/... -run 'TestEveryMutationMovesItsClaimedAxis' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-08-T3 | 08 | 6 | EVD-05 | — | Delete the superseded marker guard and the surviving stale prose | unit | `go test ./internal/compiler/session/... -run 'TestNoSuppressionOutlivesItsWitness\|TestSuppressionWitnessGuardIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-09-T1 | 09 | 7 | EVD-02 | — | Build the derivation ladder, the total order, and the cap | unit | `go test ./internal/compiler/session/... -run 'TestValidationGradeVocabularyIsClosed$\|TestValidationGradeOrderIsTotal$\|TestValidationRowGradesAreEarned$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-09-T2 | 09 | 7 | EVD-02 | — | Derive over the archived corpus, migrate the fourteen verification maps | unit | `go test ./internal/compiler/session/... -run 'TestValidationRowGradesAreEarned$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-09-T3 | 09 | 7 | EVD-02 | — | Prove the cap is not inert, re-pin the lint frontier | unit | `go test ./internal/compiler/session/... -run 'TestValidationGradeCapIsNotInert' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-10-T1 | 10 | 8 | EVD-03 | — | Author reconciliation entries under a closed verdict vocabulary | unit | `go test ./internal/compiler/session/... -run 'TestReconciliationVerdictsCarryTheirObligations' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-10-T2 | 10 | 8 | EVD-03 | — | Generate the reconciliation view and couple archive edits to it | unit | `go test ./internal/compiler/session/... -run 'TestEvidenceReconciliationViewIsCurrent' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-10-T3 | 10 | 8 | EVD-01 | — | Empty three of four frontier classes and close the phase's evidence | unit | `go test ./internal/compiler/session/... -run 'TestVerificationGroundednessThreeClassesAreEmpty' -count=1 -v` | ✅ exists | WIRED | — |
| 14-11-T1 | 11 | 9 | EVD-02 | — | End-to-end completion witness for one run-record batch | unit | `go test ./internal/compiler/session/... -run 'TestRunRecordCarriesACompletionWitness$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-11-T2 | 11 | 9 | EVD-02 | — | An incomplete or margin-less run record fails by name, and the timeout tells the truth | unit | `go test ./internal/compiler/session/... -run 'TestRunRecordCompletenessGuardIsNotInert$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-11-T3 | 11 | 9 | EVD-02 | — | Producer and consumer consult one anchored name set; self-citation is refused | unit | `go test ./internal/compiler/session/... -run 'TestEvidencePatternsResolveToAnchoredNames$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-13-T1 | 13 | 9 | EVD-04 | — | One unallowlisted build constraint, end to end, goes red | unit | `go test ./internal/compiler/session/... -run 'TestBuildConstraintsOutsideTheAllowlistAreRefused$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-13-T2 | 13 | 9 | EVD-04 | — | The scan sees constrained-out files, and the new branch carries its own seeded fault | unit | `go test ./internal/compiler/session/... -run 'TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-12-T1 | 12 | 9 | EVD-02 | — | The exemption cannot silently return | unit | `go test ./internal/compiler/session/... -run 'TestValidationGradeBarAppliesToPhase14$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-12-T2 | 12 | 9 | EVD-02/PRC-01 | — | Any residual narrowing names an owner | unit | `go test ./internal/compiler/session/... -run 'TestValidationGradeBarRowExemptionsAreOwned$' -count=1 -v` | ✅ exists | EXERCISED | — |
| 14-12-T3 | 12 | 9 | PRC-01 | — | D-14-121 leaves UNOWNED with a real reason | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed$' -count=1 -v` | ✅ exists | EXERCISED | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Filled by the planner. Existing `go test` infrastructure covers all phase
requirements — no framework install is needed; Wave 0 is limited to any new
test files the plans introduce.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | — |

*Target: all phase behaviors have automated verification. A row added here needs
an explicit justification — this phase's premise is that reviewer judgment is the
failure mode being engineered out.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Every `<automated>` command was executed and matched ≥1 test (no dead `-run` patterns)
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 200s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

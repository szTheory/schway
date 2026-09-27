---
phase: "21"
slug: "native-emission-ownership-and-resource-discharge-m004"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-25"
---

# Phase 21 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing package and test runner |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./internal/compiler/session ./internal/compiler/cgen -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | Measure in Phase 21; do not assume a value |

---

## Sampling Rate

- **After every task commit:** Run the task's focused Go test command.
- **After every plan wave:** Run `go test ./... -count=1`.
- **Before `$gsd-verify-work`:** Full suite must be green.
- **Max feedback latency:** Record measured focused and full-suite durations during execution.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|-------|---------------|--------|
| 21-01-01 | 01 | 1 | Phase goal | — | All modeled exits have an explicit, checked disposition; admitted contract paths require discharge evidence | unit | `go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-02-01 | 02 | 1 | Phase goal | — | Legacy foreign/by-pointer lowering bodies are absent while public emission remains single-law and fail-closed | unit | `go test ./internal/compiler/cgen -run '^TestPhase21LegacyEmitterBodiesRetired$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-02-02 | 02 | 1 | Phase goal | — | Existing whole-program foreign and by-pointer refusal gates remain green | unit | `go test ./internal/compiler/cgen -run 'TestPublicDispatchUsesOnlyEmitProgram|TestProgramBorrowedByPointerDisposition|TestProgramBranchValidationOrder' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-03-01 | 03 | 1 | Phase goal | — | The tagged compiler comparison is bound to its recorded receipt by a cheap recurring structural test; the one-shot four-lane run is recorded below | unit | `go test ./internal/compiler/session -run '^TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-03-02 | 03 | 1 | Phase goal | — | Existing seeded semantic control still detects comparator divergence | unit | `go test ./internal/compiler/session -run '^TestPhase16EmitterPortSemanticGuardIsNotInert$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-04-01 | 04 | 2 | Phase goal | — | Debt records and generated claims view reflect measured/retired evidence without widening the claim | unit | `go test ./internal/compiler/session -run 'TestDebtRegistersAreWellFormed|TestUnreachableClaimsViewIsCurrent' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-05-01 | 05 | 3 | Phase goal | — | Archived Phase 16 cut decision resolves and unsupported by-pointer lowering refuses before serialization | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestProgramBorrowedByPointerDisposition$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-05-02 | 05 | 3 | Phase goal | — | NAT-09 family ownership matches archived debt and current provisional roadmap registration; seeded mutations are rejected | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase16EmitterCutsAreAmendedAndOwned$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-06-01 | 06 | 3 | Phase goal | — | Validation corpus exactly matches the consumer request and rejects tampering or vacuity | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestValidationCorpusPairExportMatchesConsumer|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestCheckedInCorpusRecordRejectsTamperingAndVacuity)$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |
| 21-06-02 | 06 | 3 | Phase goal | — | Groundedness findings and reconciliation obligations match the archived M003 corpus | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestVerificationGroundednessThreeClassesAreEmpty|TestVerificationGroundednessFrontierIsPinned|TestReconciliationVerdictsCarryTheirObligations|TestEvidenceReconciliationViewIsCurrent)$' -count=1 -v` | ✅ | EXERCISED | — | ✅ passed |

---

## Wave 0 Requirements

- Phase 21 plans create the contract artifact, structural tests, one-shot build-tagged LTO comparison, and retirement guard before their focused commands run.
- Existing Go and Clang toolchain are present; no new framework or package installation is required.

---

## Manual-Only Verifications

All phase behaviors have automated verification. Human judgment is reserved for an explicit user decision if a later emitter-admission phase proposes changing the locked refusal boundary.

## One-Shot Compiler Evidence

`go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v` passed on 2026-09-26; exact fixture, emitted-C, Clang, host, and lane facts are recorded in `21-LTO-EVIDENCE.md`. The tagged comparison is intentionally not repeated in ordinary CI; `TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison` checks its source/receipt binding on every normal session-suite run.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency measured and recorded
- [x] `nyquist_compliant: true` set in frontmatter after execution

**Approval:** automated verification complete; no human-only checks remain for this phase.

## Validation Audit 2026-09-27

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

The four tasks added by gap-closure Plans 21-05 and 21-06 are now mapped to their focused automated checks. The review-fix regression test `TestSchema2PayloadOutcomeBoundIsPreflighted` also passes in the full suite and asserts refusal before C serialization. Full `go test ./...` and `go build ./...` pass on the current tree.

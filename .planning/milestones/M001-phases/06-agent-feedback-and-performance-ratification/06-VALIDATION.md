---
phase: "6"
slug: "agent-feedback-and-performance-ratification"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-06"
evidence_vocabulary: v1
graded_rows: 30
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go's built-in `testing` package — no third-party test framework anywhere in the tree (`go.mod` has no `require` block) |
| **Config file** | none — `go test` driven directly; `scripts/assert-go-tests.sh` wraps exact-test selection with a nonexistent-sentinel guard that refuses a test name the package does not actually declare |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/<pkg>/... <ExactTestName> [...]` |
| **Full suite command** | `go test ./... && go test -race ./... && go vet ./...` |
| **Estimated runtime** | ~60–90 seconds for the full suite; per-task exact-selection runs are seconds |

**Why `assert-go-tests.sh` and not bare `go test -run`:** a bare `-run` regex that matches
nothing exits 0. The wrapper lists the package's real targets first and fails when a requested
name was never discovered, so a typo or a deleted test can never render as green. Every Phase 6
task's `<automated>` command must go through it, matching the Phase 5 precedent verbatim.

---

## Sampling Rate

- **After every task commit:** Run the task's exact-selection `assert-go-tests.sh` command
- **After every plan wave:** Run `go test ./... && go test -race ./... && go vet ./...`
- **Before `/gsd-verify-work`:** Full suite green **and** `sh scripts/verify-phase6.sh` green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

Task IDs are assigned by the planner; this table records the requirement→test-type→command
binding each task must satisfy. The planner fills task IDs and plan/wave columns.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| 06-01-T2 | 06-01 | 1 | FND-04 | T-06-01 | Metrics/Lane fields never reach content identity, so measurement cannot alter semantic output | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/protocol/... TestMetricsAndLaneFieldsExcludedFromIdentity TestIdentityFieldEnumerationIsExhaustive` | ❌ W0 | EXERCISED | — |
| 06-01-T1 | 06-01 | 1 | FND-04, DX-03, QLT-02 | T-06-02 | Prior-phase bytes frozen through Phase 5 before any schema bump lands | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/core/... TestPreviousPhaseCoreBytesUnchanged TestPreviousPhaseManifestIDsUnchanged TestPreviousPhaseGoldenCUnchanged` | ✅ | EXERCISED | — |
| 06-01-T3 | 06-01 | 1 | DX-03, QLT-02 | T-06-03 | The 12 lane-schema literal sites are pinned so the coordinated bump cannot half-land | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestLaneSchemaLiteralSiteCountIsPinned` | ❌ W0 | EXERCISED | — |
| 06-02-T2 | 06-02 | 2 | DX-02 | T-06-EXPLAIN-01 | Bounded output; truncation code emitted rather than unbounded expansion | unit + CLI golden | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestExplainRespectsDepthAndNodeBudget TestExplainTruncationCodeIsStable TestExplainZeroCauseDiagnosticReturnsSingleNode TestExplainEdgeKindVocabularyIsClosed` | ❌ W0 | EXERCISED | — |
| 06-02-T3 | 06-02 | 2 | DX-02 | — | Synthesized cause DAG is deterministic across cold invocations | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestExplainCauseGraphIsDeterministic TestExplainNodeOrderIsStableOnTies` | ❌ W0 | EXERCISED | — |
| 06-03-T2 | 06-03 | 3 | DX-02 | T-06-QUERY-03 | One joined addressing surface; no sixth vocabulary; honest absence for an unknown ID | unit + CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestQueryResolvesEveryStableIDVocabulary TestQueryMintsNoSixthVocabulary TestQueryUnknownIDReportsNotCaptured TestQueryKindFilterVocabularyIsClosed` | ❌ W0 | EXERCISED | — |
| 06-03-T3 | 06-03 | 3 | DX-02 | T-06-QUERY-01 | Cursor pagination is bounded, lossless, and stably ordered | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestQueryCursorPaginationIsBounded TestQueryPagesConcatenateExactlyOnce TestQueryResultOrderIsStableOnTies TestQueryMalformedCursorIsUsageError` | ❌ W0 | EXERCISED | — |
| 06-04-T2 | 06-04 | 2 | DX-03 | T-06-CACHE-03 | Undeclared cache input cannot produce a reuse; the Clang digest is not just a version string | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/cache/... TestCacheKeyCoversEveryDeclaredInput TestMutationRunnerSourceHashIsADeclaredInput TestClangDigestIsNotJustTheVersionString TestClangDigestProbeIsBounded` | ❌ W0 | EXERCISED | — |
| 06-04-T3 | 06-04 | 2 | DX-03 | T-06-CACHE-04 | Fail-closed to not_cacheable; a cold store can never report reuse | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/cache/... TestCacheFailsClosedToNotCacheable TestCacheEqualDeclaredInputsShareOneEntry TestCacheEmptyStoreReportsCold TestCacheStatusVocabularyIsClosed TestCacheMismatchedMetaIsTreatedAsAbsent` | ❌ W0 | EXERCISED | — |
| 06-05-T2 | 06-05 | 3 | DX-03 | T-06-RISK-01 | Risk-to-lane selection widens on unclassified input; first run is cold | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestRiskLaneUnclassifiedFixtureWidensToAllLanes TestRiskLaneUndeclaredInputIsNotCacheable TestChangeStateFirstRunIsColdAndSelectsAll TestChangeStateCorruptFileWidens TestChangeStateFileIsLocalAndGitignored` | ❌ W0 | EXERCISED | — |
| 06-05-T3 | 06-05 | 3 | DX-03 | T-06-RISK-02 | The declared table cannot drift from the code that emits the lanes, in either direction | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestRiskLaneAuditRefusesStaleLaneID TestRiskLaneAuditRefusesEmptyRegistry TestRiskLaneAuditRefusesUndeclaredLane TestRiskLaneAuditRefusesUnknownDeclaredInput TestRiskLaneDoubleMatchSelectsOnce TestRiskLaneSelectionOrderIsStable` | ❌ W0 | EXERCISED | — |
| 06-06-T1 | 06-06 | 4 | FND-04, DX-03, QLT-02 | T-06-BUMP-02 | Both schemas bump to /1 together; all 12 lane sites move at once | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestLaneSchemaLiteralSiteCountIsPinned` | ❌ W0 | EXERCISED | — |
| 06-06-T3 | 06-06 | 4 | FND-04, DX-03, QLT-02 | T-06-BUMP-01 | /0 bytes frozen and prior-phase evidence byte-identical after the bump | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/protocol/... TestSchemaZeroConstantsStillExist TestPhase1EvidenceBytesUnchangedAfterBump TestCommandSchemaIsVersionOne` | ❌ W0 | EXERCISED | — |
| 06-07-T2 | 06-07 | 5 | DX-03, FND-04 | T-06-VERIFY-02 | A deferred lane never renders pass; no live lane goes silently missing | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestDeferredLaneNeverRendersPass TestEveryLiveLaneIsAccountedFor TestCacheInputsReusedCountIsCounted TestFirstColdRunReportsRecomputedNotReused TestLaneStatusVocabularyIsClosed` | ❌ W0 | EXERCISED | — |
| 06-07-T3 | 06-07 | 5 | DX-03 | T-06-VERIFY-04 | evidence validates compact and expands traces on failure or request, verdict unaffected | unit + CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestEvidenceValidateIsCompactByDefault TestEvidenceExpandsTraceOnFailure TestEvidenceExpandsTraceOnRequest TestExpansionNeverChangesTheVerdict TestEvidenceTraceRespectsOutputBound` | ❌ W0 | EXERCISED | — |
| 06-08-T1 | 06-08 | 2 | QLT-02, FND-04 | T-06-MEASURE-01 | machine_id carries six declared facts and no host fingerprint | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/measure/... TestMachineIDIsDerivedFromDeclaredFacts TestMachineIDExcludesHostFingerprints TestMachineProbeIsBounded TestMachineFactFieldsAreClosed` | ❌ W0 | EXERCISED | — |
| 06-08-T3 | 06-08 | 2 | QLT-02 | T-06-MEASURE-04 | CoV auto-demotion: a noisy metric becomes observed, never blocking | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/measure/... TestCoVAboveThresholdDemotesToObserved TestCoVDemotionNeverPromotesToBlocking TestOnlyRecomputedWorkIsGateEligible TestVerdictVocabularyIsClosed TestUndeterminableCoVIsNotRatified` | ❌ W0 | WIRED | — |
| 06-09-T2 | 06-09 | 6 | QLT-02 | T-06-BUDGET-01 | The budget audit refuses a manifest whose declared machines do not match live probe output | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestBudgetAuditRefusesUndeclaredMachine TestBudgetAuditRefusesEmptyManifest TestBudgetAuditRefusesDuplicateRow TestBudgetAuditRefusesIneligibleHardGate TestUndeclaredMachineRunsObservationOnly TestUndeclaredMachineNeverWritesManifest` | ❌ W0 | EXERCISED | — |
| 06-09-T3 | 06-09 | 6 | QLT-02, FND-04 | T-06-BUDGET-03 | recomputed_work is the only hard gate; a wall-clock overrun must cite a stage work delta | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestRecomputedWorkIsTheOnlyHardGate TestWallClockObservationIsNeverBlocking TestBlockingRegressionCitesStageWorkDelta TestBudgetCeilingIsStrictlyExceeded TestEvaluateBudgetAgreesWithDemote TestUncitableObservationReportsHonestly` | ❌ W0 | WIRED | — |
| 06-10-T2 | 06-10 | 7 | FND-04 | T-06-STAGE-01 | Stage timing is single-gated and identity-free; unobserved runs are byte-stable | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestStageTimingIsGatedOnObserveTiming TestStageBreakdownAbsentWhenUnobserved TestTimingEnvironmentIsReadInExactlyOnePlace` | ❌ W0 | EXERCISED | — |
| 06-10-T3 | 06-10 | 7 | FND-04 | T-06-STAGE-03 | peak_rss stays unavailable and the reason is recorded in code | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/protocol/... TestPeakRSSStaysUnavailable TestPeakRSSUnavailabilityIsDocumented TestNoGetrusageAnywhere` | ❌ W0 | EXERCISED | — |
| 06-11-T2 | 06-11 | 2 | DX-04 | T-06-REPAIR-01 | A coordinate shift cannot move a published diagnostic ID | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/diagnostic/... TestRepairIdentityUsesKindOnly TestRepairSpanShiftDoesNotChangeDiagnosticID TestRepairFieldEnumerationIsExhaustive` | ❌ W0 | EXERCISED | — |
| 06-11-T3 | 06-11 | 2 | DX-04 | T-06-REPAIR-02 | Only a complete MachineApplicable repair is driver-eligible; /0 stays frozen | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/diagnostic/... TestOnlyMachineApplicableIsDriverEligible TestDiagnosticZeroBytesUnchanged` | ❌ W0 | EXERCISED | — |
| 06-12-T3 | 06-12 | 3 | DX-04 | T-06-INJECT-01 | **Marker mutation-kill:** each injector refuses when its target marker disappears | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestEveryInjectorRefusesWhenMarkerDisappears TestInjectorMarkerCountGuardIsNotInert TestInjectorRefusalPropagatesToExerciseFailure` | ❌ W0 | EXERCISED | — |
| 06-13-T2 | 06-13 | 4 | DX-04 | T-06-BOUNDARY-01 | Repair driver import boundary: the build fails if it imports the compiler's private tree | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/lang-repair/... TestRepairDriverImportsStayOutsideInternal TestRepairDriverNeverOpensSourceOutsideSpan TestRepairDriverSpawnsOnlyTheLangBinary TestImportBoundaryTestIsNotInert` | ❌ W0 | EXERCISED | — |
| 06-13-T3 | 06-13 | 4 | DX-04 | T-06-BOUNDARY-04 | Single-pass repair of all five classes; the degenerate delete-the-code repair is rejected | subprocess CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/lang-repair/... TestRepairDriverFixesEveryDefectClassSinglePass TestRepairOracleRejectsDeleteTheCode TestStaleEvidenceRepairRebindsManifest TestUnrepairableDefectFailsTheGate TestRepairSelectionIsSpecifiedOnTies` | ❌ W0 | EXERCISED | — |
| 06-14-T2 | 06-14 | 5 | DX-04 | T-06-THEATER-01 | **Prose-scramble guard:** identical outcome with all prose replaced by lorem ipsum | subprocess CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/lang-repair/... TestProseScrambleLeavesRepairBehaviourIdentical TestProseScrambleFixtureKeepsStructuredFieldsIntact` | ❌ W0 | EXERCISED | — |
| 06-14-T3 | 06-14 | 5 | DX-04 | T-06-THEATER-02 | **Vocabulary-removal guard:** the driver goes RED when the structured channel is stripped and prose is left intact | subprocess CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/lang-repair/... TestVocabularyRemovalDrivesTheDriverRed TestVocabularyRemovalGuardIsNotInert` | ❌ W0 | EXERCISED | — |
| 06-15-T2 | 06-15 | 8 | all five | T-06-GATE-01 | Script and Go control sets cannot drift apart, in either direction | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestPhase6RequiredControlsMatchScript TestPhase6BoundsMatchScript TestPhase6VerifierScriptContract TestPhase6ScriptInvokesNoPriorGate TestPhase6SamplingLoopMatchesGoStatistics` | ❌ W0 | EXERCISED | — |
| 06-15-T3 | 06-15 | 8 | all five | T-06-GATE-02 | Every accepted escape is declared, visible, and never presented as a solved control | unit + shell | `sh scripts/verify-phase6.sh` | ❌ W0 | REACHABLE | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/compiler/protocol/protocol_test.go` — add `TestMetricsAndLaneFieldsExcludedFromIdentity`
      **before any new `Metrics`/`Lane` field is added.** Research confirmed `Result.Finalize()`'s
      identity struct already excludes `Metrics` entirely and reduces `Lane` to `ID+":"+Status`,
      so D-06-32 holds today — this test is what keeps it holding once fields are added.
- [ ] Re-pin prior-phase frozen bytes first (the Phase 5 `05-01` precedent):
      `TestPreviousPhaseCoreBytesUnchanged`, `TestPreviousPhaseManifestIDsUnchanged`,
      `TestPreviousPhaseGoldenCUnchanged` must be green before the coordinated
      `lang.command/0`→`/1` and `lang.verify-lane/0`→`/1` bump lands.
- [ ] `internal/compiler/cache/` — brand-new package, zero existing tests; needs `cache_test.go`
      from the first task that creates it.
- [ ] `internal/compiler/session/session_phase6_*_test.go` — sibling test files per the Phase 5
      convention, one per new surface (explain/query, cache wiring, budget).
- [ ] A Go home for p50/p95/CoV. Research found the "Phase 2 20-sample machinery" is **shell**
      (`scripts/verify-phase2.sh`'s `observe()`), not a Go library — no Go package computes these
      today. D-06-19's CoV auto-demotion needs new, unit-testable Go code.
- [ ] `cmd/<repair-driver>/` — new binary plus its import-boundary test.
- [ ] Held-out `.lang` defect fixture corpus (D-06-29), structurally separate from any fixtures
      used to hand-derive the driver's kind→edit mapping. None exists yet.
- [ ] `scripts/verify-phase6.sh` — following `verify-phase5.sh`'s exact shape: self-test sentinel,
      `GOCACHE` pinned to a disposable temp dir, `go test ./...` → race → vet → build, per-corpus
      JSON captured then grepped for required `control:*` strings, expected-escapes grepped
      separately, then this phase's own 20-sample `observe()`-style loop.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| The **recorded agent-legibility exercise** (D-06-30): repair rounds against the `p95: 2` intent, token/tool-call cost, and whether protocol-only access sufficed | DX-04 | Explicitly **non-gating by decision**. A live LLM in the suite would break the offline / deterministic / stdlib-only constraints (D-06-23). Runs on a per-milestone or on-demand cadence. | Drive the shipped `lang` binary through the five defect classes using only `--json` output and the structured `Repair` records. Record rounds, tokens, and tool calls. Report honestly-unmeasured values rather than fabricating them. Store alongside other phase evidence artifacts. |
| Wall-clock cold/warm p50/p95 **ratification** on a newly declared machine | QLT-02 | Ratifying a *new* `machine_id` into `qlt02_budget_manifest.json` is a reviewed human act by design — the JSON diff is the review surface (D-06-16), and an auto-ratifying gate would defeat it. The *audit* of an already-ratified manifest is automated. | Run the 20-sample loop on the target host, inspect the reported distribution and computed `machine_id`, then commit the manifest row with `ratified_at` and `ratified_by_commit`. On an undeclared host the run is observation-only (`ratified: false`) and writes nothing (D-06-18). |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

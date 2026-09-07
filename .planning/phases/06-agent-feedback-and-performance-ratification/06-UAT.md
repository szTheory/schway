---
status: complete
phase: 06-agent-feedback-and-performance-ratification
source: 06-01-SUMMARY.md, 06-02-SUMMARY.md, 06-03-SUMMARY.md, 06-04-SUMMARY.md, 06-05-SUMMARY.md, 06-06-SUMMARY.md, 06-07-SUMMARY.md, 06-08-SUMMARY.md, 06-09-SUMMARY.md, 06-10-SUMMARY.md, 06-11-SUMMARY.md, 06-12-SUMMARY.md, 06-13-SUMMARY.md, 06-14-SUMMARY.md, 06-15-SUMMARY.md
started: 2026-09-07T21:05:22Z
updated: 2026-09-07T21:05:22Z
---

## Current Test

[testing complete]

## Tests

<!-- Every deliverable below is deterministically covered by a passing automated
     test (uat.classify-coverage reports all_auto_covered=true, present=0,
     errors=0 for all 15 SUMMARYs). No human checkpoint was required.
     The five CLI behaviors that previously needed hands-on confirmation are now
     asserted through the shipped binary in
     internal/compiler/testsupport/cli_phase6_test.go, and the Phase 6 release
     gate runs on every push via .github/workflows/ci.yml's phase-gate job,
     pinned by session_phase6_test.go#TestCIWorkflowRunsPhase6Gate. -->

### 1. Phase 1-5 core bytes, manifest IDs, and generated-C goldens are pinned against any Phase 6 perturbation
expected: Phase 1-5 core bytes, manifest IDs, and generated-C goldens are pinned against any Phase 6 perturbation
result: pass
source: automated
coverage_id: D1 (06-01-SUMMARY.md)
verification: unit:internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged; unit:internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged; unit:internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged

### 2. D-06-32 (no Metrics/Lane field reaches Result.Finalize() identity) is an executable, self-invalidating pin
expected: D-06-32 (no Metrics/Lane field reaches Result.Finalize() identity) is an executable, self-invalidating pin
result: pass
source: automated
coverage_id: D2 (06-01-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestMetricsAndLaneFieldsExcludedFromIdentity; unit:internal/compiler/protocol/protocol_test.go#TestIdentityFieldEnumerationIsExhaustive

### 3. The 12-site lang.verify-lane/0 literal count is pinned per-file so the 06-06 coordinated bump cannot half-land
expected: The 12-site lang.verify-lane/0 literal count is pinned per-file so the 06-06 coordinated bump cannot half-land
result: pass
source: automated
coverage_id: D3 (06-01-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_pin_test.go#TestLaneSchemaLiteralSiteCountIsPinned

### 4. lang explain returns a bounded synthesized cause DAG addressed by a stable diagnostic ID under the net-new lang.explain/0 schema, without dumping the whole program
expected: lang explain returns a bounded synthesized cause DAG addressed by a stable diagnostic ID under the net-new lang.explain/0 schema, without dumping the whole program
result: pass
source: automated
coverage_id: D1 (06-02-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainSummarySchemaIsMinted; e2e:internal/compiler/testsupport/cli_phase6_test.go#TestExplainCLIReturnsBoundedCauseDAG

### 5. The cause DAG is computed fresh per cold invocation and never persisted; two cold invocations on the same input produce byte-identical output
expected: The cause DAG is computed fresh per cold invocation and never persisted; two cold invocations on the same input produce byte-identical output
result: pass
source: automated
coverage_id: D2 (06-02-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainCauseGraphIsDeterministic; unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainSynthesisOpensNoWritePath

### 6. --depth defaults to 3, a hard node-count budget applies, and exceeding either truncates with a stable code rather than growing
expected: --depth defaults to 3, a hard node-count budget applies, and exceeding either truncates with a stable code rather than growing
result: pass
source: automated
coverage_id: D3 (06-02-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainRespectsDepthAndNodeBudget; unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainTruncationCodeIsStable

### 7. A zero-cause diagnostic returns a single root node and no edges (honest empty result); equal-comparing nodes have a specified, stable order across cold invocations
expected: A zero-cause diagnostic returns a single root node and no edges (honest empty result); equal-comparing nodes have a specified, stable order across cold invocations
result: pass
source: automated
coverage_id: D4 (06-02-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainZeroCauseDiagnosticReturnsSingleNode; unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainNodeOrderIsStableOnTies

### 8. Node availability reuses the existing not_captured/optimized_out vocabulary rather than fabricating a value
expected: Node availability reuses the existing not_captured/optimized_out vocabulary rather than fabricating a value
result: pass
source: automated
coverage_id: D5 (06-02-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_explain_test.go#TestExplainSummarySchemaIsMinted

### 9. lang query is one joined addressing surface resolving any of the five stable ID vocabularies already shipped in the tree, under the net-new lang.query/0 schema, without minting a sixth
expected: lang query is one joined addressing surface resolving any of the five stable ID vocabularies already shipped in the tree, under the net-new lang.query/0 schema, without minting a sixth
result: pass
source: automated
coverage_id: D1 (06-03-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_query_test.go#TestQueryResolvesEveryStableIDVocabulary; unit:internal/compiler/session/session_phase6_query_test.go#TestQueryMintsNoSixthVocabulary; e2e:internal/compiler/testsupport/cli_phase6_test.go#TestQueryCLIResolvesEveryStableIDVocabulary; e2e:internal/compiler/testsupport/cli_phase6_test.go#TestQueryCLIMintsNoSixthVocabulary

### 10. An ID that resolves to nothing returns an honest not_captured fact, never an error and never a fabricated value, across all five vocabularies
expected: An ID that resolves to nothing returns an honest not_captured fact, never an error and never a fabricated value, across all five vocabularies
result: pass
source: automated
coverage_id: D2 (06-03-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_query_test.go#TestQueryUnknownIDReportsNotCaptured; e2e:internal/compiler/testsupport/cli_phase6_test.go#TestQueryCLIUnknownIDReportsNotCaptured

### 11. Results are bounded lists paginated by --cursor, never by --depth; pages concatenate exactly once with no loss and no duplication, and a malformed cursor is a usage error
expected: Results are bounded lists paginated by --cursor, never by --depth; pages concatenate exactly once with no loss and no duplication, and a malformed cursor is a usage error
result: pass
source: automated
coverage_id: D3 (06-03-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_query_test.go#TestQueryCursorPaginationIsBounded; unit:internal/compiler/session/session_phase6_query_test.go#TestQueryPagesConcatenateExactlyOnce; unit:internal/compiler/session/session_phase6_query_test.go#TestQueryMalformedCursorIsUsageError; unit:internal/compiler/session/session_phase6_query_test.go#TestQueryDepthDoesNotPaginate

### 12. Equal-comparing query results have a specified, stable order across two cold invocations
expected: Equal-comparing query results have a specified, stable order across two cold invocations
result: pass
source: automated
coverage_id: D4 (06-03-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_query_test.go#TestQueryResultOrderIsStableOnTies

### 13. --kind filters the fact list by D-06-05's closed five-value vocabulary; an unrecognized kind is a usage error, never a silent no-op
expected: --kind filters the fact list by D-06-05's closed five-value vocabulary; an unrecognized kind is a usage error, never a silent no-op
result: pass
source: automated
coverage_id: D5 (06-03-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_query_test.go#TestQueryKindFilterVocabularyIsClosed

### 14. The cache stores only expensive intermediate artifacts and never a verdict; nothing in the package can record or return a pass/fail judgement (D-06-06)
expected: The cache stores only expensive intermediate artifacts and never a verdict; nothing in the package can record or return a pass/fail judgement (D-06-06)
result: pass
source: automated
coverage_id: D1 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/cache_test.go#TestCacheExportedSurfaceStoresNoVerdict; unit:internal/compiler/cache/cache_test.go#TestCacheImportsStayIndependent

### 15. A cache key is a SHA-256 over an explicitly declared input list, and meta.json records that full list so a reviewer can see exactly what the key covered (D-06-07, D-06-09)
expected: A cache key is a SHA-256 over an explicitly declared input list, and meta.json records that full list so a reviewer can see exactly what the key covered (D-06-07, D-06-09)
result: pass
source: automated
coverage_id: D2 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/cache_test.go#TestCacheRoundTripsArtifactAndMeta; unit:internal/compiler/cache/probe_test.go#TestCacheKeyCoversEveryDeclaredInput

### 16. A mutant binary built by a changed mutation runner misses the cache, because the mutation runner's own source hash is a declared input (D-06-07)
expected: A mutant binary built by a changed mutation runner misses the cache, because the mutation runner's own source hash is a declared input (D-06-07)
result: pass
source: automated
coverage_id: D3 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/probe_test.go#TestMutationRunnerSourceHashIsADeclaredInput

### 17. When any declared input cannot be computed (e.g. the Clang probe fails), the entry is not_cacheable for that run; ambiguity and silence resolve to run it, never to skip it (D-06-11)
expected: When any declared input cannot be computed (e.g. the Clang probe fails), the entry is not_cacheable for that run; ambiguity and silence resolve to run it, never to skip it (D-06-11)
result: pass
source: automated
coverage_id: D4 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/cache_test.go#TestCacheFailsClosedToNotCacheable; unit:internal/compiler/cache/probe_test.go#TestClangDigestProbeIsBounded

### 18. Two artifacts whose declared input lists are byte-identical resolve to the same cache entry -- a merge by construction, not a collision (FND-04 adjacency edge)
expected: Two artifacts whose declared input lists are byte-identical resolve to the same cache entry -- a merge by construction, not a collision (FND-04 adjacency edge)
result: pass
source: automated
coverage_id: D5 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/cache_test.go#TestCacheEqualDeclaredInputsShareOneEntry

### 19. A first run against an empty cache directory reports every artifact as recomputed and never as reused (FND-04 empty-input edge)
expected: A first run against an empty cache directory reports every artifact as recomputed and never as reused (FND-04 empty-input edge)
result: pass
source: automated
coverage_id: D6 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/cache_test.go#TestCacheEmptyStoreReportsCold

### 20. The clang_identity input is a real probed digest of the binary's own bytes, not merely its --version string, and the probe is bounded
expected: The clang_identity input is a real probed digest of the binary's own bytes, not merely its --version string, and the probe is bounded
result: pass
source: automated
coverage_id: D7 (06-04-SUMMARY.md)
verification: unit:internal/compiler/cache/probe_test.go#TestClangDigestIsNotJustTheVersionString; unit:internal/compiler/cache/probe_test.go#TestClangDigestProbeIsBounded

### 21. Risk-to-lane selection is a checked-in declared table (risk_lanes.json), not a derived dependency graph, with an executable audit cross-checking it against the live lane-ID set (D-06-10)
expected: Risk-to-lane selection is a checked-in declared table (risk_lanes.json), not a derived dependency graph, with an executable audit cross-checking it against the live lane-ID set (D-06-10)
result: pass
source: automated
coverage_id: D1 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneRegistryLoads; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditPassesCheckedInRegistry; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneVerifyRegistryRunsCleanly

### 22. A fixture whose kind is not exhaustively classified, or whose declared inputs cannot be fully computed, selects ALL its lanes and is reported not_cacheable for that run (D-06-11)
expected: A fixture whose kind is not exhaustively classified, or whose declared inputs cannot be fully computed, selects ALL its lanes and is reported not_cacheable for that run (D-06-11)
result: pass
source: automated
coverage_id: D2 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneUnclassifiedFixtureWidensToAllLanes; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneUndeclaredInputIsNotCacheable

### 23. Changed is measured against a gitignored local state file recording last-seen input hash per (fixture x lane); the first run on any machine is cold and is reported as cold (D-06-08)
expected: Changed is measured against a gitignored local state file recording last-seen input hash per (fixture x lane); the first run on any machine is cold and is reported as cold (D-06-08)
result: pass
source: automated
coverage_id: D3 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateFirstRunIsColdAndSelectsAll; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateCorruptFileWidens; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateFileIsLocalAndGitignored; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneChangeStateRecordsOnePerFixtureLane

### 24. A lane matched by two different risk-table rows is selected exactly once, with a single selection_reason (FND-04 adjacency edge)
expected: A lane matched by two different risk-table rows is selected exactly once, with a single selection_reason (FND-04 adjacency edge)
result: pass
source: automated
coverage_id: D4 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneDoubleMatchSelectsOnce

### 25. The very first run with no state file selects every lane and reports every one as cold (FND-04 empty-input edge)
expected: The very first run with no state file selects every lane and reports every one as cold (FND-04 empty-input edge)
result: pass
source: automated
coverage_id: D5 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateFirstRunIsColdAndSelectsAll

### 26. When two lanes compare equal on the selection key, selection output order is specified and stable across two cold invocations (FND-04 ordering edge)
expected: When two lanes compare equal on the selection key, selection output order is specified and stable across two cold invocations (FND-04 ordering edge)
result: pass
source: automated
coverage_id: D6 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneSelectionOrderIsStable; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneSelectLanesOutputHasNoDuplicates

### 27. The audit is bidirectional: a declared row referencing a nonexistent lane or input name fails it, and a live lane with no row also fails it
expected: The audit is bidirectional: a declared row referencing a nonexistent lane or input name fails it, and a live lane with no row also fails it
result: pass
source: automated
coverage_id: D7 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesStaleLaneID; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesEmptyRegistry; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesUndeclaredLane; unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesUnknownDeclaredInput

### 28. session_phase6_risklanes.go has no dependency on internal/compiler/protocol -- this file's wiring is independent of how a later plan attaches Selection to protocol.Lane
expected: session_phase6_risklanes.go has no dependency on internal/compiler/protocol -- this file's wiring is independent of how a later plan attaches Selection to protocol.Lane
result: pass
source: automated
coverage_id: D8 (06-05-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneFileDoesNotImportProtocol

### 29. One coordinated additive bump lands in one commit: lang.command/0 to /1 and lang.verify-lane/0 to /1, with all /0 bytes frozen byte-for-byte
expected: One coordinated additive bump lands in one commit: lang.command/0 to /1 and lang.verify-lane/0 to /1, with all /0 bytes frozen byte-for-byte
result: pass
source: automated
coverage_id: D1 (06-06-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestCommandSchemaIsVersionOne; unit:internal/compiler/protocol/protocol_test.go#TestSchemaZeroConstantsStillExist; e2e:internal/compiler/testsupport/cli_phase6_test.go#TestVerifyPhase1CLIEmitsCoordinatedSchemaBump

### 30. All 12 hardcoded lane-schema literal sites move together; a half-landed bump fails the site-count pin from 06-01
expected: All 12 hardcoded lane-schema literal sites move together; a half-landed bump fails the site-count pin from 06-01
result: pass
source: automated
coverage_id: D2 (06-06-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_pin_test.go#TestLaneSchemaLiteralSiteCountIsPinned

### 31. protocol.Lane gains cache_status, selection_reason, machine_id, gate_verdict, cold_or_warm, stage_breakdown; protocol.Metrics gains cache_inputs_reused_count
expected: protocol.Lane gains cache_status, selection_reason, machine_id, gate_verdict, cold_or_warm, stage_breakdown; protocol.Metrics gains cache_inputs_reused_count
result: pass
source: automated
coverage_id: D3 (06-06-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestIdentityFieldEnumerationIsExhaustive; unit:internal/compiler/protocol/protocol_test.go#TestNewLaneAndMetricsFieldsAreAdditive; unit:internal/compiler/protocol/protocol_test.go#TestLaneVocabulariesAreClosed; unit:internal/compiler/protocol/protocol_test.go#TestCacheInputsReusedCountSharesUnitWithRecomputedWork

### 32. Every new Metrics and Lane field stays outside Result.Finalize()'s identity struct, so measurement cannot change semantic output
expected: Every new Metrics and Lane field stays outside Result.Finalize()'s identity struct, so measurement cannot change semantic output
result: pass
source: automated
coverage_id: D4 (06-06-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestMetricsAndLaneFieldsExcludedFromIdentity

### 33. Phase 1 evidence bytes remain byte-identical after the bump
expected: Phase 1 evidence bytes remain byte-identical after the bump
result: pass
source: automated
coverage_id: D5 (06-06-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestPhase1EvidenceBytesUnchangedAfterBump; unit:internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged; unit:internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged; unit:internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged

### 34. cache_status is drawn from the four-value artifact vocabulary and never uses hit or miss terminology
expected: cache_status is drawn from the four-value artifact vocabulary and never uses hit or miss terminology
result: pass
source: automated
coverage_id: D6 (06-06-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestLaneVocabulariesAreClosed

### 35. VerifyPhase6ChangedRisk selects lane:native-differential by real changed risk, compiles and caches its artifact, and always re-runs the checker and the interpreter-vs-native comparator fresh -- a cache hit changes only what is skipped, never what is asserted (D-06-06)
expected: VerifyPhase6ChangedRisk selects lane:native-differential by real changed risk, compiles and caches its artifact, and always re-runs the checker and the interpreter-vs-native comparator fresh -- a cache hit changes only what is skipped, never what is asserted (D-06-06)
result: pass
source: automated
coverage_id: D1 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_verify_test.go#TestVerifyPhase6ChangedRiskRunsOneSelectedLane; unit:internal/compiler/session/session_phase6_verify_test.go#TestNativeDifferentialAssertionOutsideCacheBranch

### 36. Every lane VerifyPhase6ChangedRisk emits carries a schema-1 lane with CacheStatus drawn from the closed four-value artifact vocabulary and SelectionReason beginning with one of the three declared prefixes
expected: Every lane VerifyPhase6ChangedRisk emits carries a schema-1 lane with CacheStatus drawn from the closed four-value artifact vocabulary and SelectionReason beginning with one of the three declared prefixes
result: pass
source: automated
coverage_id: D2 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_verify_test.go#TestVerifyReportsCacheStatusVocabulary

### 37. A lane that was not run renders status deferred and is structurally incapable of rendering pass; no lane silently goes missing from the Result
expected: A lane that was not run renders status deferred and is structurally incapable of rendering pass; no lane silently goes missing from the Result
result: pass
source: automated
coverage_id: D3 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_verify_test.go#TestDeferredLaneNeverRendersPass; unit:internal/compiler/session/session_phase6_verify_test.go#TestEveryLiveLaneIsAccountedFor

### 38. cache_inputs_reused_count is counted in the same integer unit as recomputed_work and reflects artifacts actually reused this run; a cold run reports zero reused and never artifact_reused
expected: cache_inputs_reused_count is counted in the same integer unit as recomputed_work and reflects artifacts actually reused this run; a cold run reports zero reused and never artifact_reused
result: pass
source: automated
coverage_id: D4 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_verify_test.go#TestCacheInputsReusedCountIsCounted; unit:internal/compiler/session/session_phase6_verify_test.go#TestFirstColdRunReportsRecomputedNotReused

### 39. The lane-status vocabulary this verify path emits is closed, and protocol.LaneStatuses() names it exhaustively
expected: The lane-status vocabulary this verify path emits is closed, and protocol.LaneStatuses() names it exhaustively
result: pass
source: automated
coverage_id: D5 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_verify_test.go#TestLaneStatusVocabularyIsClosed

### 40. evidence --validate returns the compact summary by default, expands the trace on failure or --expand, and expansion never changes the verdict (Status or diagnostic codes)
expected: evidence --validate returns the compact summary by default, expands the trace on failure or --expand, and expansion never changes the verdict (Status or diagnostic codes)
result: pass
source: automated
coverage_id: D6 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceValidateIsCompactByDefault; unit:internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceExpandsTraceOnFailure; unit:internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceExpandsTraceOnRequest; unit:internal/compiler/session/session_phase6_evidence_test.go#TestExpansionNeverChangesTheVerdict

### 41. The expanded evidence trace respects the existing 64 KiB-plus-one output bound and truncates with a stable code rather than growing unboundedly
expected: The expanded evidence trace respects the existing 64 KiB-plus-one output bound and truncates with a stable code rather than growing unboundedly
result: pass
source: automated
coverage_id: D7 (06-07-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceTraceRespectsOutputBound

### 42. internal/compiler/session/session.go is unchanged by this plan
expected: internal/compiler/session/session.go is unchanged by this plan
result: pass
source: automated
coverage_id: D8 (06-07-SUMMARY.md)
verification: other:git diff --stat 612f7b5..HEAD -- internal/compiler/session/session.go (empty)

### 43. Declared-machine probe produces a leak-free machine_id from six declared facts, with bounded subprocess probes for clang_version and cpu_model
expected: Declared-machine probe produces a leak-free machine_id from six declared facts, with bounded subprocess probes for clang_version and cpu_model
result: pass
source: automated
coverage_id: D1 (06-08-SUMMARY.md)
verification: unit:internal/compiler/measure/machine_test.go#TestMachineIDIsDerivedFromDeclaredFacts; unit:internal/compiler/measure/machine_test.go#TestMachineIDExcludesHostFingerprints; unit:internal/compiler/measure/machine_test.go#TestMachineProbeIsBounded; unit:internal/compiler/measure/machine_test.go#TestMachineFactFieldsAreClosed

### 44. 20-warm-sample p50/p95 statistics matching scripts/verify-phase2.sh's sorted-index protocol exactly, refusing empty/short/non-positive sample sets
expected: 20-warm-sample p50/p95 statistics matching scripts/verify-phase2.sh's sorted-index protocol exactly, refusing empty/short/non-positive sample sets
result: pass
source: automated
coverage_id: D2 (06-08-SUMMARY.md)
verification: unit:internal/compiler/measure/statistics_test.go#TestWarmSamplePercentilesMatchShellPrecedent; unit:internal/compiler/measure/statistics_test.go#TestPercentileSelectionIsStableOnTies; unit:internal/compiler/measure/statistics_test.go#TestSampleStatisticsRefuseShortSampleSets; unit:internal/compiler/measure/statistics_test.go#TestSampleStatisticsRefuseEmptyAndNonPositive; unit:internal/compiler/measure/statistics_test.go#TestSampleCountConstantsAreExported

### 45. CoV auto-demotion rule mechanically quarantines unstable metrics to observed, restricts blocking eligibility to recomputed_work, and cannot promote to blocking
expected: CoV auto-demotion rule mechanically quarantines unstable metrics to observed, restricts blocking eligibility to recomputed_work, and cannot promote to blocking
result: pass
source: automated
coverage_id: D3 (06-08-SUMMARY.md)
verification: unit:internal/compiler/measure/statistics_test.go#TestCoVAboveThresholdDemotesToObserved; unit:internal/compiler/measure/statistics_test.go#TestCoVDemotionNeverPromotesToBlocking; unit:internal/compiler/measure/statistics_test.go#TestOnlyRecomputedWorkIsGateEligible; unit:internal/compiler/measure/statistics_test.go#TestVerdictVocabularyIsClosed; unit:internal/compiler/measure/statistics_test.go#TestUndeterminableCoVIsNotRatified; unit:internal/compiler/measure/statistics_test.go#TestDemoteHasExactlyOnePromotionPassthrough

### 46. A checked-in qlt02_budget_manifest.json with the plan's exact seven-field row shape, an executable audit cross-checking declared machine_ids against a live probe
expected: A checked-in qlt02_budget_manifest.json with the plan's exact seven-field row shape, an executable audit cross-checking declared machine_ids against a live probe
result: pass
source: automated
coverage_id: D1 (06-09-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetManifestLoads; unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetLaneCarriesMachineIDAndVerdict; unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesUndeclaredMachine; unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesEmptyManifest; unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesDuplicateRow; unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesIneligibleHardGate

### 47. Observation-only mode on an undeclared machine: full distribution + machine_id reported, ratified false, no manifest write, no manifest-consulting budget lookup
expected: Observation-only mode on an undeclared machine: full distribution + machine_id reported, ratified false, no manifest write, no manifest-consulting budget lookup
result: pass
source: automated
coverage_id: D2 (06-09-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_budget_test.go#TestUndeclaredMachineRunsObservationOnly; unit:internal/compiler/session/session_phase6_budget_test.go#TestUndeclaredMachineNeverWritesManifest; unit:internal/compiler/session/session_phase6_budget_test.go#TestNoFunctionWritesTheBudgetManifest

### 48. recomputed_work is the only hard gate; strictly-exceeds not exactly-equal; wall-clock/output-bytes observations are never blocking on their own and must cite a stage work delta or report uncitable
expected: recomputed_work is the only hard gate; strictly-exceeds not exactly-equal; wall-clock/output-bytes observations are never blocking on their own and must cite a stage work delta or report uncitable
result: pass
source: automated
coverage_id: D3 (06-09-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_budget_test.go#TestRecomputedWorkIsTheOnlyHardGate; unit:internal/compiler/session/session_phase6_budget_test.go#TestBudgetCeilingIsStrictlyExceeded; unit:internal/compiler/session/session_phase6_budget_test.go#TestWallClockObservationIsNeverBlocking; unit:internal/compiler/session/session_phase6_budget_test.go#TestUncitableObservationReportsHonestly; unit:internal/compiler/session/session_phase6_budget_test.go#TestEvaluateBudgetAgreesWithDemote; unit:internal/compiler/session/session_phase6_budget_test.go#TestBlockingRegressionCitesStageWorkDelta

### 49. session.StageRecorder records explicit start/stop timestamp pairs at D-06-21's five fixed pipeline stages (parse, check, lower, native_compile, link) with no tracing runtime, span model, or global registry
expected: session.StageRecorder records explicit start/stop timestamp pairs at D-06-21's five fixed pipeline stages (parse, check, lower, native_compile, link) with no tracing runtime, span model, or global registry
result: pass
source: automated
coverage_id: D1 (06-10-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_stages_test.go#TestStageBreakdownCoversEveryPipelineStage; unit:internal/compiler/session/session_phase6_stages_test.go#TestStageRecorderRefusesUnknownStage; unit:internal/compiler/session/session_phase6_stages_test.go#TestStageRecorderFileDeclaresNoMutableStateOrGoroutines

### 50. Stage timing is single-gated on LANG_OBSERVE_TIMING via TimingObservationEnabled, the sole reader of that env var; a run with the switch unset emits no stage_breakdown at all, never an array of zeros
expected: Stage timing is single-gated on LANG_OBSERVE_TIMING via TimingObservationEnabled, the sole reader of that env var; a run with the switch unset emits no stage_breakdown at all, never an array of zeros
result: pass
source: automated
coverage_id: D2 (06-10-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_stages_test.go#TestStageTimingIsGatedOnObserveTiming; unit:internal/compiler/session/session_phase6_stages_test.go#TestStageBreakdownAbsentWhenUnobserved; unit:internal/compiler/session/session_phase6_stages_test.go#TestTimingEnvironmentIsReadInExactlyOnePlace

### 51. stage_breakdown never enters Result.Finalize()'s identity struct, so stage attribution cannot change semantic output
expected: stage_breakdown never enters Result.Finalize()'s identity struct, so stage attribution cannot change semantic output
result: pass
source: automated
coverage_id: D3 (06-10-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestStageBreakdownExcludedFromIdentity; unit:internal/compiler/protocol/protocol_test.go#TestMetricsAndLaneFieldsExcludedFromIdentity; unit:internal/compiler/protocol/protocol_test.go#TestIdentityFieldEnumerationIsExhaustive

### 52. peak_rss_status remains unavailable for M001; the reason is recorded in a code comment at protocol.PeakRSSUnavailable and mechanically enforced -- no raw literal, no getrusage call, no PeakRSSBytes assignment anywhere in the tree
expected: peak_rss_status remains unavailable for M001; the reason is recorded in a code comment at protocol.PeakRSSUnavailable and mechanically enforced -- no raw literal, no getrusage call, no PeakRSSBytes assignment anywhere in the tree
result: pass
source: automated
coverage_id: D4 (06-10-SUMMARY.md)
verification: unit:internal/compiler/protocol/protocol_test.go#TestPeakRSSStaysUnavailable; unit:internal/compiler/protocol/protocol_test.go#TestPeakRSSUnavailabilityIsDocumented; unit:internal/compiler/protocol/protocol_test.go#TestNoGetrusageAnywhere

### 53. diagnostic.Repair carries optional Span, Replacement, and a closed Applicability enum, all non-identity-bearing
expected: diagnostic.Repair carries optional Span, Replacement, and a closed Applicability enum, all non-identity-bearing
result: pass
source: automated
coverage_id: D1 (06-11-SUMMARY.md)
verification: unit:internal/compiler/diagnostic/diagnostic_test.go#TestRepairCarriesSpanReplacementApplicability; unit:internal/compiler/diagnostic/diagnostic_test.go#TestApplicabilityVocabularyIsClosed; unit:internal/compiler/diagnostic/diagnostic_test.go#TestEmptyApplicabilityDefaultsToUnspecified; other:lang --json check testdata/phase2/implicit_noncopy.lang (shipped binary, manual run)

### 54. The identity split holds: span/replacement/applicability never move a diagnostic ID; kind does; the field set is exhaustively enumerated
expected: The identity split holds: span/replacement/applicability never move a diagnostic ID; kind does; the field set is exhaustively enumerated
result: pass
source: automated
coverage_id: D2 (06-11-SUMMARY.md)
verification: unit:internal/compiler/diagnostic/diagnostic_test.go#TestRepairIdentityUsesKindOnly; unit:internal/compiler/diagnostic/diagnostic_test.go#TestRepairSpanShiftDoesNotChangeDiagnosticID; unit:internal/compiler/diagnostic/diagnostic_test.go#TestRepairFieldEnumerationIsExhaustive; other:manual mutation-kill: folding Span into ErrorWithRepairs's identity struct flips TestRepairSpanShiftDoesNotChangeDiagnosticID; adding an unused Repair field flips TestRepairFieldEnumerationIsExhaustive; unit:internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged; unit:internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged

### 55. Only a complete MachineApplicable repair is driver-eligible; lang.diagnostic/0 stays frozen and repair-free
expected: Only a complete MachineApplicable repair is driver-eligible; lang.diagnostic/0 stays frozen and repair-free
result: pass
source: automated
coverage_id: D3 (06-11-SUMMARY.md)
verification: unit:internal/compiler/diagnostic/diagnostic_test.go#TestOnlyMachineApplicableIsDriverEligible; unit:internal/compiler/diagnostic/diagnostic_test.go#TestMachineApplicableWithoutMaterialIsNotEligible; unit:internal/compiler/diagnostic/diagnostic_test.go#TestDiagnosticZeroBytesUnchanged

### 56. Five defect injectors exist, one per named class, each answering to session.Injector
expected: Five defect injectors exist, one per named class, each answering to session.Injector
result: pass
source: automated
coverage_id: D1 (06-12-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_injectors_test.go#TestMatchDefectInjectorProducesExactlyOneDefect; unit:internal/compiler/session/session_phase6_injectors_test.go#TestMoveDefectInjectorProducesExactlyOneDefect; unit:internal/compiler/session/session_phase6_injectors_test.go#TestBorrowDefectInjectorProducesExactlyOneDefect; unit:internal/compiler/session/session_phase6_injectors_test.go#TestCleanupInjectorReusesReleaseOmissionRunner; unit:internal/compiler/session/session_phase6_injectors_test.go#TestStaleEvidenceInjectorBreaksManifestBinding

### 57. Each of the four source-granularity injectors produces exactly one mechanical change, and every injector refuses (phase6.injector_target_missing) rather than silently passing when its target marker disappears
expected: Each of the four source-granularity injectors produces exactly one mechanical change, and every injector refuses (phase6.injector_target_missing) rather than silently passing when its target marker disappears
result: pass
source: automated
coverage_id: D2 (06-12-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_injectors_test.go#TestEveryInjectorProducesExactlyOneMechanicalChange; unit:internal/compiler/session/session_phase6_injectors_test.go#TestInjectorTargetChoiceIsSpecified; unit:internal/compiler/session/session_phase6_injectors_test.go#TestEveryInjectorRefusesWhenMarkerDisappears; unit:internal/compiler/session/session_phase6_injectors_test.go#TestInjectorMarkerCountGuardIsNotInert; unit:internal/compiler/session/session_phase6_injectors_test.go#TestInjectorRefusalPropagatesToExerciseFailure; other:go test -race ./internal/compiler/session/... and go test -race ./... (full suite)

### 58. The held-out corpus is structurally distinct from the derivation corpus (D-06-29)
expected: The held-out corpus is structurally distinct from the derivation corpus (D-06-29)
result: pass
source: automated
coverage_id: D3 (06-12-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_injectors_test.go#TestPhase6DefectCorpusIsHeldOut

### 59. cmd/lang-repair driver: spawns the shipped lang binary via --json check FILE, decodes only repairs[] (never message/detail), splices a MachineApplicable repair's replacement into its span, and re-verifies clean -- single pass
expected: cmd/lang-repair driver: spawns the shipped lang binary via --json check FILE, decodes only repairs[] (never message/detail), splices a MachineApplicable repair's replacement into its span, and re-verifies clean -- single pass
result: pass
source: automated
coverage_id: D1 (06-13-SUMMARY.md)
verification: unit:cmd/lang-repair/repair_test.go#TestRepairDriverFixesOneDefectEndToEnd; unit:cmd/lang-repair/repair_test.go#TestRepairDriverDecodesNoProseFields

### 60. Structural import-boundary lint fails the build if cmd/lang-repair ever imports internal/, or opens a .lang file outside the span it was handed
expected: Structural import-boundary lint fails the build if cmd/lang-repair ever imports internal/, or opens a .lang file outside the span it was handed
result: pass
source: automated
coverage_id: D2 (06-13-SUMMARY.md)
verification: unit:cmd/lang-repair/import_boundary_test.go#TestRepairDriverImportsStayOutsideInternal; unit:cmd/lang-repair/import_boundary_test.go#TestImportBoundaryTestIsNotInert; unit:cmd/lang-repair/import_boundary_test.go#TestRepairDriverNeverOpensSourceOutsideSpan

### 61. Single-pass CI gate fixes match/move/borrow/cleanup/stale-evidence, each bounded and O(1); the oracle rejects a clean-but-degenerate repair; unrepairable and tied-selection edges are covered
expected: Single-pass CI gate fixes match/move/borrow/cleanup/stale-evidence, each bounded and O(1); the oracle rejects a clean-but-degenerate repair; unrepairable and tied-selection edges are covered
result: pass
source: automated
coverage_id: D3 (06-13-SUMMARY.md)
verification: unit:cmd/lang-repair/repair_test.go#TestRepairDriverFixesEveryDefectClassSinglePass; unit:cmd/lang-repair/repair_test.go#TestRepairOracleRejectsDeleteTheCode; unit:cmd/lang-repair/repair_test.go#TestStaleEvidenceRepairRebindsManifest; unit:cmd/lang-repair/repair_test.go#TestUnrepairableDefectFailsTheGate; unit:cmd/lang-repair/repair_test.go#TestRepairSelectionIsSpecifiedOnTies

### 62. Fixture-substitution subprocess: a content-hash-keyed stand-in lets the real driver be driven against arbitrary/mutated lang --json check documents with zero driver code changes, proven faithful against the real binary end to end
expected: Fixture-substitution subprocess: a content-hash-keyed stand-in lets the real driver be driven against arbitrary/mutated lang --json check documents with zero driver code changes, proven faithful against the real binary end to end
result: pass
source: automated
coverage_id: D1 (06-14-SUMMARY.md)
verification: unit:cmd/lang-repair/antitheater_test.go#TestFixtureSubstitutionIsFaithful; unit:cmd/lang-repair/antitheater_test.go#TestBaselineCaptureContainsEligibleRepair; unit:cmd/lang-repair/antitheater_test.go#TestMutateCaptureIdentityRoundTripIsByteStable

### 63. Anti-theater guard 1 (prose-scramble): scrambling every message/detail string leaves repair behaviour (bytes, exit code, reported outcome) identical across match/move/borrow; cleanup/stale-evidence proven structurally prose-independent; the scrambler itself is verified bidirectionally
expected: Anti-theater guard 1 (prose-scramble): scrambling every message/detail string leaves repair behaviour (bytes, exit code, reported outcome) identical across match/move/borrow; cleanup/stale-evidence proven structurally prose-independent; the scrambler itself is verified bidirectionally
result: pass
source: automated
coverage_id: D2 (06-14-SUMMARY.md)
verification: unit:cmd/lang-repair/antitheater_test.go#TestProseScrambleLeavesRepairBehaviourIdentical; unit:cmd/lang-repair/antitheater_test.go#TestProseScrambleFixtureKeepsStructuredFieldsIntact; other:Live falsification: temporarily made selectRepair require message text containing \"ownership transferred\"; match/move/borrow subtests failed as expected; reverted (repair.go confirmed byte-identical by diff)

### 64. Anti-theater guard 2 (vocabulary removal): stripping repairs[]/kind/span/replacement with prose intact drives the driver RED; a green control on the identical harness proves the guard is not inert
expected: Anti-theater guard 2 (vocabulary removal): stripping repairs[]/kind/span/replacement with prose intact drives the driver RED; a green control on the identical harness proves the guard is not inert
result: pass
source: automated
coverage_id: D3 (06-14-SUMMARY.md)
verification: unit:cmd/lang-repair/antitheater_test.go#TestVocabularyRemovalDrivesTheDriverRed; unit:cmd/lang-repair/antitheater_test.go#TestVocabularyRemovalGuardIsNotInert; other:Live falsification: strip_kind subtest failed before the driverEligible Kind-check fix (deviation below), passed after

### 65. The differential-behaviour fallback oracle lives CI-gate-side, in-process, in internal/compiler/session -- never inside the driver, which may not import internal/compiler/reduce
expected: The differential-behaviour fallback oracle lives CI-gate-side, in-process, in internal/compiler/session -- never inside the driver, which may not import internal/compiler/reduce
result: pass
source: automated
coverage_id: D4 (06-14-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_oracle_test.go#TestDifferentialFallbackOracleRunsInProcess; other:go test ./... && go test -race ./... && go vet ./... (full suite)

### 66. scripts/verify-phase6.sh is a runnable Phase 6 release-cost lane that runs green end-to-end, and is executed on every push and pull request on both host priorities
expected: scripts/verify-phase6.sh is a runnable Phase 6 release-cost lane that runs green end-to-end, and is executed on every push and pull request on both host priorities
result: pass
source: automated
coverage_id: D1 (06-15-SUMMARY.md)
verification: e2e:.github/workflows/ci.yml#phase-gate -- runs `sh scripts/verify-phase6.sh` on ubuntu-latest and macos-latest for every push and pull request; unit:internal/compiler/session/session_phase6_test.go#TestCIWorkflowRunsPhase6Gate; unit:internal/compiler/session/session_phase6_test.go#TestPhase6VerifierScriptContract; unit:internal/compiler/session/session_phase6_test.go#TestPhase6ScriptInvokesNoPriorGate

### 67. The gate script and its Go sources cannot drift apart in either direction: required controls, pinned bounds, and declared escapes are each asserted against the script's own text
expected: The gate script and its Go sources cannot drift apart in either direction: required controls, pinned bounds, and declared escapes are each asserted against the script's own text
result: pass
source: automated
coverage_id: D2 (06-15-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_test.go#TestPhase6RequiredControlsMatchScript; unit:internal/compiler/session/session_phase6_test.go#TestPhase6BoundsMatchScript; unit:internal/compiler/session/session_phase6_escapes_test.go#TestPhase6EscapeGrepsMatchScript; unit:internal/compiler/session/session_phase6_escapes_test.go#TestPhase6EscapesAreNeverPresentedAsControls

### 68. The Phase 6 control-and-work gate, its corpus dispatch, and the `lang stats` sampling seam are each reachable and correct through the shipped binary, not only in process
expected: The Phase 6 control-and-work gate, its corpus dispatch, and the `lang stats` sampling seam are each reachable and correct through the shipped binary, not only in process
result: pass
source: automated
coverage_id: D3 (06-15-SUMMARY.md)
verification: unit:internal/compiler/session/session_phase6_test.go#TestVerifyPhase6ControlsAndWork; unit:cmd/lang/main_test.go#TestPhase6CorpusDispatchRequiresMarker; unit:internal/compiler/session/session_phase6_test.go#TestPhase6SamplingLoopMatchesGoStatistics; e2e:internal/compiler/testsupport/cli_phase6_test.go#TestStatsCLIComputesGoStatistics

## Summary

total: 68
passed: 68
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

[none]

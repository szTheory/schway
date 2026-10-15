---
phase: 36
slug: scalar-helper-calls
status: validated
nyquist_compliant: true
wave_0_complete: true
created: 2026-10-14
validated: 2026-10-15
source_revision: d117ca4058fbf6a727cb5aa634768cf6194bac34
---

# Phase 36 — Validation Audit

The pre-execution draft has been reconciled against all eight executable plans and summaries. Every one of the 21 task rows has named behavioral evidence with actual RUN/PASS output. No implementation or fixture gap needed another test. The Phase 36 validation map is complete; phase completion and requirement checkboxes remain with the parent workflow.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 testing; installed C17 compiler for native behavioral cases |
| Config | go.mod; no new package or dependency |
| Cache | GOCACHE=/tmp/ai-lang-verification-gocache |
| Focused runner | Exact named selectors with -count=1 -v -timeout=600s |
| Owner receipts | Fixed Phase 36 owner lists and captured argv/log hashes in the manifests below |
| Host/toolchain in accepted receipt | macOS arm64; Go 1.24.0; Apple Clang 21.0.1 |

The source-bound final capture at revision d117ca4058fbf6a727cb5aa634768cf6194bac34 passed 45/45 expected roots across 12 packages. The same revision's compatibility capture passed 15/15 roots across session, cgen, and cmd/schway. I checked each manifest log against its SHA-256 and verified its anchored root RUN/PASS lines, exit code, and absence of FAIL/SKIP. Current HEAD is 7dafaa2db0629735e36b5552c4074b351d00513c; a Git comparison confirmed compiler, command, example, fixture, and dependency sources are unchanged from the captured revision.

- [Final owner-results.json](/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json)
- [Compatibility owner-results.json](/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/owner-results.json)

I also ran the remaining inherited selectors omitted by those fixed owner lists as exact focused commands at the current source state. The 32 named test invocations below all had one root RUN and one root PASS, exit 0, and no skipped or failed root. One selector, TestInvocationPreflightGuardIsNotInert, was also present in the compatibility capture.

## Per-Task Validation Map

| Task ID | Plan / wave | Requirement | Threat | Current behavioral evidence | Type | Status |
|---|---|---|---|---|---|---|
| 36-01-01 | 01 / 1 | CALL-02 | T-36-01 | internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaseline — literal outcomes for 163 old tracked sources across the three emitter routes | integration | green |
| 36-01-02 | 01 / 1 | CALL-02 | T-36-01 | internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaselineControls — byte drift, omitted row, false exemption, old-source ledger and path containment controls | integration | green |
| 36-02-01 | 02 / 2 | CALL-01, CALL-02 | T-36-02 | internal/compiler/syntax/parser_phase36_test.go#TestPhase36ScalarCallGrammar; #TestPhase36ScalarCallDepthAndRecovery; #TestPhase36ScalarCallFormatting; parser_phase34_recovery_test.go#TestPhase34RecoveryCorpus; #TestPhase34RecoveryControls | unit / focused integration | green |
| 36-02-02 | 02 / 2 | CALL-01, CALL-02 | T-36-03 | internal/compiler/check/check_phase36_scalar_calls_test.go#TestPhase36ScalarCallLoweringAndTypePairs — four signature pairs, ordered computed arguments and fresh typed results | unit / focused integration | green |
| 36-02-03 | 02 / 2 | CALL-01, CALL-02 | T-36-03, T-36-04 | internal/compiler/check/check_phase36_scalar_calls_test.go#TestPhase36CallSourceIntentControls; #TestPhase36CallImportAndRefusalBoundaries; inherited Phase30/34 import, Bool-tail and arity guards | unit / focused integration | green |
| 36-03-01 | 03 / 3 | CALL-01, CALL-02 | T-36-05 | internal/compiler/corevalidate/corevalidate_phase36_test.go#TestPhase36RawCallShapeAndFreshness; #TestPhase36RawCallRefusalRemovalControls; corevalidate_phase36_legacy_internal_test.go#TestPhase36CallCFGApplicability; #TestPhase36LegacyFileOwnerCalls | unit / focused integration | green |
| 36-03-02 | 03 / 3 | CALL-01, CALL-02 | T-36-06 | internal/compiler/originvalidate/originvalidate_phase36_test.go#TestPhase36OriginCallFreshnessAndTransfer; #TestPhase36OriginCallRefusalRemovalControls | unit / focused integration | green |
| 36-03-03 | 03 / 3 | CALL-01, CALL-02 | T-36-07 | internal/compiler/pathoracle/pathoracle_phase36_test.go#TestPhase36ScalarCallAuthorityBoundaries; pathoracle_phase36_borrow_test.go#TestPhase36ScalarCallLocalBorrowCopies; #TestPhase36ScalarCallLocalBorrowControls; pathoracle_phase35_test.go#TestPhase35PureOperationAncestry | unit / focused integration | green |
| 36-04-01 | 04 / 4 | CALL-01, CALL-02 | T-36-08 | interp_phase36_test.go#TestPhase36InterpreterRepeatedCallTracer; executionpeer_phase36_test.go#TestPhase36DynamicCallOrdinalControls; #TestPhase36CallCausalAttributionControls | unit / focused integration | green |
| 36-04-02 | 04 / 4 | CALL-01, CALL-02 | T-36-08, T-36-09 | interp_phase36_test.go#TestPhase36InterpreterCallAnswersAndDefects; executionpeer_phase36_test.go#TestPhase36DynamicCallOrdinalControls; #TestPhase36CallCausalAttributionControls; executionpeer_test.go#TestPhase26PeerOccurrenceOrder; #TestValidateObservedCausalStructure; #TestExecutionPeerFailuresAreActionable | unit / focused integration | green |
| 36-04-03 | 04 / 4 | CALL-01, CALL-02 | T-36-10, T-36-11 | interp_phase36_test.go#TestPhase36InterpreterEvidenceBoundaries; executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls — exact/plus-one capture, ordinal ceiling, continued work and incomplete-prefix refusal | unit / focused integration | green |
| 36-05-01 | 05 / 4 | CALL-01, CALL-02 | T-36-12, T-36-13 | internal/compiler/cgen/cgen_phase36_test.go#TestPhase36NativeDynamicInvocationPaths — O0/O3/O3-LTO answers and retained canonical nested/repeated paths | integration | green |
| 36-05-02 | 05 / 4 | CALL-01, CALL-02 | T-36-14, T-36-15 | cgen_phase36_test.go#TestPhase36CallEventCapacityBoundary; #TestPhase36NativeOrdinalAndPathBoundaries; native_phase36_test.go#TestPhase36CaptureOutcomeSeparation; native_phase33_capture_test.go#TestPhase33CaptureReadBounded | integration | green |
| 36-05-03 | 05 / 4 | CALL-01, CALL-02 | T-36-12, T-36-13, T-36-16 | cgen_phase36_test.go#TestPhase36EmitterCallInventoryControls; cgen_phase35_test.go#TestPhase35EmitterAdmissionGates; #TestPhase35UnsignedShiftEmissionAndEventProjection; cgen_program_test.go#TestInvocationPathTableBoundary; #TestInvocationPreflightGuardIsNotInert; #TestProgramInvocationIndexThreading; #TestNativeFunctionCalledPreorder; cgen_phase273_test.go#TestPhase273NamedCapacityExitGate; #TestPhase273CapacityContractLiteralsAgree; core_test.go#TestAllOperationKindsHandledAtEverySite | unit / focused integration | green |
| 36-06-01 | 06 / 5 | CALL-01, CALL-02 | T-36-17 | session_phase36_calls_test.go#TestPhase36HelperAnswersAndDefects; #TestPhase36HelperOptimizationMatrix; #TestPhase36HelperUBSan — local/imported independent literals, four pairs, dynamic defects, O0/O3/O3-LTO and UBSan | integration | green |
| 36-06-02 | 06 / 5 | CALL-01, CALL-02 | T-36-18, T-36-19 | main_phase36_test.go#TestPhase36PublicHelperConformance; main_phase321_run_input_test.go#TestPhase321RunInputRequiredCLI; session_phase36_calls_test.go#TestPhase36ByteViewCallerArgumentBoundary; #TestPhase36LegacyByteViewCalleeBounds; #TestPhase36NativeCallCaptureOutcomeSeparation; session_phase36_byteview_test.go#TestPhase36NativeByteViewCallerArgumentBoundary; #TestPhase36NativeByteViewLoopHelpers; #TestPhase36NativeFileByteViewLoopHelpers; #TestPhase36ByteViewArgumentOrderMutationControls; inherited Phase29 fixture/bounds roots | integration | green |
| 36-06-03 | 06 / 5 | CALL-01, CALL-02 | T-36-17 | session_phase36_peer_mutation_test.go#TestPhase36EngineCallMutationControls; session_phase36_calls_test.go#TestPhase36HelperAnswersAndDefects — same-typed argument/result, swallowed overflow and remainder controls at engine assertions | integration | green |
| 36-07-01 | 07 / 6 | CALL-01, CALL-02 | T-36-20 | surfacemanifest/manifest_test.go#TestPhase36SurfaceDocument; session_phase36_surface_test.go#TestPhase36SurfaceManifestCallWitnesses — complete source witnesses and exact U64/Bool pairs/contexts | unit / integration | green |
| 36-07-02 | 07 / 6 | CALL-01, CALL-02 | T-36-21 | session_phase36_surface_test.go#TestPhase36SurfaceManifestCallCensus; #TestPhase36SurfaceManifestCallControls; #TestPhase36SurfaceLimitConsumers; all named Phase34/35 witness, census, control, document and limit roots | unit / integration | green |
| 36-08-01 | 08 / 7 | CALL-01, CALL-02 | T-36-22 | [36-08-PLAN.md Task 1](36-08-PLAN.md) fixed compatibility command; [current compatibility manifest](/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/owner-results.json) — 15/15 roots, frozen C/ledger and old safety routes | integration | green |
| 36-08-02 | 08 / 7 | CALL-01, CALL-02 | T-36-23 | [36-08-PLAN.md Task 2](36-08-PLAN.md) amended fixed final capture and document checks; [actual final manifest](/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json) — 45/45 roots across 12 packages; dated acceptance and living amendments checked | integration / artifact audit | green |

## Supplemental Inherited Selector Commands

These commands filled the inherited-selector omissions in the pre-execution map. All listed roots produced RUN/PASS at the current source state.

| Owner package | Exact command | Roots |
|---|---|---:|
| syntax | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase34RecoveryCorpus|TestPhase34RecoveryControls)$' -count=1 -v -timeout=600s | 2 |
| check | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase30PrivateNameRefused|TestPhase30ImportCollisions|TestPhase30ImportItemOrderIrrelevant|TestPhase34BoolSpellingsRemainIdentifiers|TestPhase34CallArityBoundaries)$' -count=1 -v -timeout=600s | 5 |
| pathoracle | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/pathoracle -run '^TestPhase35PureOperationAncestry$' -count=1 -v -timeout=600s | 1 |
| executionpeer | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^(TestPhase26PeerOccurrenceOrder|TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable)$' -count=1 -v -timeout=600s | 3 |
| native | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase33CaptureReadBounded$' -count=1 -v -timeout=600s | 1 |
| cgen | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase35EmitterAdmissionGates|TestPhase35UnsignedShiftEmissionAndEventProjection|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestNativeFunctionCalledPreorder|TestPhase273NamedCapacityExitGate|TestPhase273CapacityContractLiteralsAgree)$' -count=1 -v -timeout=600s | 7 |
| core | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/core -run '^TestAllOperationKindsHandledAtEverySite$' -count=1 -v -timeout=600s | 1 |
| cmd/schway | GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase321RunInputRequiredCLI$' -count=1 -v -timeout=600s | 1 |
| session | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls|TestPhase34SurfaceLimitConsumers|TestPhase35SurfaceManifestWitnesses|TestPhase35SurfaceManifestCensus|TestPhase35SurfaceManifestControls|TestPhase35SurfaceLimitConsumers)$' -count=1 -v -timeout=600s | 8 |
| surfacemanifest | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest -run '^(TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase35SurfaceDocument)$' -count=1 -v -timeout=600s | 3 |

## Requirement Classification and Evidence Boundaries

| Requirement | Validation classification | Evidence basis |
|---|---|---|
| CALL-01 | COVERED for the dated D-18 scope | Source grammar and all four typed pairs; local/imported runnable helper; literal engine answers and overflow/remainder defects; old owning-wrapper ByteView callee bounds and separate before-entry caller bounds |
| CALL-02 | COVERED for the dated D-18 scope | Checker source-intent controls; independent raw core/origin contracts; engine mutation controls; causal peer attribution; unchanged old generated-C corpus |

The Phase 36 behavior coverage is complete. These separate acceptance flags remain unresolved and must not be represented as test coverage or wired enforcement:

- CALL-01/CALL-02 remain marked unclassified/unresolved because the source specification has no classification for those edge rows.
- The two descriptorless projectProhibitions projections remain flagged unverified; there is no wired descriptor or enforcement claim.
- Hosted two-host binding remains a Phase 38 obligation. Human readability and whole-loop cold/warm performance distributions are unmeasured.

The checker and tests establish behavior at their actual owners. The causal peer does not establish arithmetic or source intent; the native/interpreter literal assertions do. The old ByteView callee-bounds compatibility test is separate from the new caller argument before-entry test.

## Audit Evidence

Current source inspection: internal/compiler/check/scalar.go; internal/compiler/interp/interp.go; internal/compiler/executionpeer/executionpeer.go; internal/compiler/cgen/cgen.go and cgen_program.go; examples/helper_calls/crc_step.schway and importer.schway.

Fresh checks: source-bound owner captures at d117ca4058fbf6a727cb5aa634768cf6194bac34 (45 final roots; 15 compatibility roots), plus the exact supplemental inherited selector commands above (32 invocations, all PASS). The actual owner manifests preserve full argv, log digests, roots, and exit codes.

Historical only: TDD RED/restore results and prior debug resolutions are historical receipts described by the eight summaries and resolved debug reports. They explain why specific controls exist; current PASS claims above come from fresh owner logs and supplemental command outputs.

## Sign-Off

- [x] All 21 task IDs, plan waves and threat references match the executable plans.
- [x] Each task's named behavior has a present behavioral test or fixed owner-capture assertion with an observable failing direction.
- [x] Every current fixed-owner selector produced actual root RUN/PASS; no zero-match or skipped result was credited.
- [x] All inherited selectors omitted from owner captures were executed with exact focused commands.
- [x] Mutation tests reached named semantic assertions after fresh no-op and restored positives.
- [x] Only after these checks: status validated, nyquist_compliant true, and wave_0_complete true.

No new test file was required. The existing suite contains the behavioral controls; no implementation bug was found, so nothing was escalated.

## Manual-Only Verifications

None expected for the deterministic criteria audited here. This sign-off does not claim readability review, complete feedback-loop distributions, hosted evidence, a full suite, or Phase 36 completion.

## Validation Audit 2026-10-15

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 21 |
| Escalated | 0 |

## Security audit coverage correction — 2026-10-15

The initial Nyquist audit marked Task36-05-03 covered but missed two explicitly promised adversarial controls. The independent security audit identified returned/reused stack-pointer lifetime and missing/extra actual emitter-site controls. Test-only04ce6fa1 added runtime lifetime and census controls; independent re-audit required50b90a30 to count actual call expressions rather than comments. The original audit is retained as history, not a claim that it discovered these gaps.

Both gaps are now resolved within TestPhase36EmitterCallInventoryControls, with clean/no-op/restored native positives, actual compiled stack-pointer mutant returning10 and a named retained-path failure, plus eight actual-call/marker/event/reservation census mutations. No new selector, dependency or implementation file was introduced. Task36-05-03 remains green with strengthened evidence. Current final capture /tmp/schway36-final-83c1f58a1720-jikc_g2_/owner-results.json observes45/45 roots across12 packages including the expanded control root, binds source/tests to83c1f58a1720770e77216ee82737efa3acdbeeaa, and was independently verified by the security auditor. All21 task rows remain covered; no manual-only gap or human UAT is inferred.

## Validation Audit 2026-10-15

| Metric | Count |
|---|---|
| Gaps found | 2 |
| Resolved | 2 |
| Escalated | 0 |

## Final regression integration — 2026-10-15

The bounded Phase33–35 sample passed80 compiled root occurrences across13 packages atb51043c6566f, including inherited frozen-C controls, actual admission mutants and global summary coverage. The sample caught and corrected exact helper route-inventory rows, Phase34 new-fixture ledger composition, the overlapping shift-law mutation operator, and missing Plan08 coverage metadata. See36-REGRESSION.md and the resolved phase36-regression-consumers debug record. These are recorded follow-up gaps to the earlier audit, not evidence that the original audit found them. Production validation and frozen expected data stay unchanged. All21 task rows remain covered by their named owners plus the corrected integration controls.


## Successor verification refresh — 2026-10-15

All eight plans and 21 task rows remain covered. Fresh exact owner selection passed 84/84 roots in 13 packages, with zero missing/fail/skip; commands, current stable input hashes and raw log hashes are in [36-REFRESH-RUN.json](36-REFRESH-RUN.json). This is a focused local run, not a whole-package/project suite or hosted receipt.

The first 19 task commands supplied Go selectors. The two plan-08 Python commands are historical source-bound capture producers that would assert current HEAD and append to the old acceptance receipt. They were not replayed. Their fixed compatibility/final owner lists were parsed from the original Python AST, merged with the other task selectors and freshly run. The original read-only artifact assertions also passed against those observed owners while retaining the original receipt revision as historical; `/tmp/schway36-artifact-grounding.json` records the distinct revisions. No historical receipt was relabeled current.

Both existing test-tier prohibition descriptors returned canonical green, with machine-proven violation-fixture RED and clean/default GREEN; see [36-REFRESH-PROHIBITIONS.json](36-REFRESH-PROHIBITIONS.json). No gap, new test, implementation change or human waiver was needed. The earlier validation/source receipt remains historical at its stated revision.

## Validation Audit 2026-10-15

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

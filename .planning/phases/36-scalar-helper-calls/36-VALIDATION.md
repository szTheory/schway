---
phase: 36
slug: scalar-helper-calls
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-10-14
---

# Phase 36 — Validation Strategy

This is the pre-execution sampling contract, derived from 36-RESEARCH's
Validation Architecture. New selectors below are planned, not existing/passing
evidence. The planner must refine task IDs/waves/threat references and ensure
each test is created before its selector can earn credit.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 `testing`, existing local compiler/application runners |
| Config | `go.mod`; no new package |
| Quick command, after test creation | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase36ScalarCallGrammar$' -count=1 -v -timeout=600s` |
| Phase verification | Exact named selectors from finalized plans; no local whole-package/full-suite or broad cross-program native matrix |
| Runtime | Unmeasured until focused execution; 600s per-command timeout is a cap, not a latency measurement |

## Sampling Rate

- Each task: exact owner selector and applicable positive/no-op/restored control.
- Each wave: completed cross-owner helper answer, defect, identity and capacity selectors.
- Phase gate: every finalized plan selector and affected inherited guard must produce actual root RUN/PASS output; zero-test success, skips and unrelated mutant failures do not count.
- Existing recurring CI owns broad checks; final two-host receipts remain Phase 38. No human-only acceptance is expected for these deterministic contracts.

## Per-Task Verification Map

Task IDs, waves and threat IDs below match the executable plans. Commands are planned samples, not passing implementation evidence. All new TestPhase36 roots are missing at planning time and are created by their owning task before first use. Exact root RUN/PASS, non-skip native lanes and each task's stated failing direction are mandatory.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---|---|---|---|---|---|---|---|---|---|
| 36-01-01 | 01 | 1 | CALL-02 | T-36-01 | Literal old C/inventory | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-01-02 | 01 | 1 | CALL-02 | T-36-01 | Reached byte/omission/false-exemption controls | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline\|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-02-01 | 02 | 2 | CALL-01, CALL-02 | T-36-02 | Bounded source grammar/recovery | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase36ScalarCallGrammar\|TestPhase36ScalarCallDepthAndRecovery\|TestPhase36ScalarCallFormatting\|TestPhase34RecoveryCorpus\|TestPhase34RecoveryControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-02-02 | 02 | 2 | CALL-01, CALL-02 | T-36-03 | Exact typed ordered source call | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase36ScalarCallLoweringAndTypePairs$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-02-03 | 02 | 2 | CALL-01, CALL-02 | T-36-03, T-36-04 | Source-intent/import/cycle refusals | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase36CallSourceIntentControls\|TestPhase36CallImportAndRefusalBoundaries\|TestPhase30PrivateNameRefused\|TestPhase30ImportCollisions\|TestPhase30ImportItemOrderIrrelevant\|TestPhase34BoolSpellingsRemainIdentifiers\|TestPhase34CallArityBoundaries)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-03-01 | 03 | 3 | CALL-01, CALL-02 | T-36-05 | Independent raw call shape/freshness | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase36RawCallShapeAndFreshness\|TestPhase36RawCallRefusalRemovalControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-03-02 | 03 | 3 | CALL-01, CALL-02 | T-36-06 | Independent origin unknown transfer | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase36OriginCallFreshnessAndTransfer\|TestPhase36OriginCallRefusalRemovalControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-03-03 | 03 | 3 | CALL-01, CALL-02 | T-36-07 | Narrow scalar authority/loan boundaries | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/pathoracle -run '^(TestPhase36ScalarCallAuthorityBoundaries\|TestPhase35PureOperationAncestry)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-04-01 | 04 | 4 | CALL-01, CALL-02 | T-36-08 | Interpreter repeated path plus independent peer | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/executionpeer -run '^(TestPhase36InterpreterRepeatedCallTracer\|TestPhase36DynamicCallOrdinalControls\|TestPhase36CallCausalAttributionControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-04-02 | 04 | 4 | CALL-01, CALL-02 | T-36-08, T-36-09 | Causal/terminal forgeries separate from numbers | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/executionpeer -run '^(TestPhase36InterpreterCallAnswersAndDefects\|TestPhase36DynamicCallOrdinalControls\|TestPhase36CallCausalAttributionControls\|TestPhase26PeerOccurrenceOrder\|TestValidateObservedCausalStructure\|TestExecutionPeerFailuresAreActionable)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-04-03 | 04 | 4 | CALL-01, CALL-02 | T-36-10, T-36-11 | Bounded evidence with actual outcome retained | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/executionpeer -run '^(TestPhase36InterpreterEvidenceBoundaries\|TestPhase36InterpreterRepeatedCallTracer\|TestPhase36DynamicCallOrdinalControls\|TestPhase36CallCausalAttributionControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-05-01 | 05 | 4 | CALL-01, CALL-02 | T-36-12, T-36-13 | Native value and stable paths | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase36NativeDynamicInvocationPaths$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-05-02 | 05 | 4 | CALL-01, CALL-02 | T-36-14, T-36-15 | Event/path/ordinal/capture-outcome boundaries | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/native -run '^(TestPhase36CallEventCapacityBoundary\|TestPhase36NativeOrdinalAndPathBoundaries\|TestPhase36CaptureOutcomeSeparation\|TestPhase33CaptureReadBounded)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-05-03 | 05 | 4 | CALL-01, CALL-02 | T-36-12, T-36-13, T-36-16 | Emitter admission/inventory/semantic guard controls | unit + focused integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/core -run '^(TestPhase36EmitterCallInventoryControls\|TestPhase35EmitterAdmissionGates\|TestPhase35UnsignedShiftEmissionAndEventProjection\|TestInvocationPathTableBoundary\|TestInvocationPreflightGuardIsNotInert\|TestProgramInvocationIndexThreading\|TestNativeFunctionCalledPreorder\|TestPhase273NamedCapacityExitGate\|TestPhase273CapacityContractLiteralsAgree\|TestAllOperationKindsHandledAtEverySite)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-06-01 | 06 | 5 | CALL-01, CALL-02 | T-36-17 | Local/importer answers and dynamic defects per lane | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperAnswersAndDefects\|TestPhase36HelperOptimizationMatrix\|TestPhase36HelperUBSan)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-06-02 | 06 | 5 | CALL-01, CALL-02 | T-36-18, T-36-19 | Public input, distinct old/new bounds, capture | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./internal/compiler/session -run '^(TestPhase36PublicHelperConformance\|TestPhase321RunInputRequiredCLI\|TestPhase36ByteViewCallerArgumentBoundary\|TestPhase36LegacyByteViewCalleeBounds\|TestPhase36NativeCallCaptureOutcomeSeparation\|TestPhase29InterpreterFixtureVectors\|TestPhase29InterpreterDefectWithLiveOwner\|TestPhase29NativeFixtureVectors)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-06-03 | 06 | 5 | CALL-01, CALL-02 | T-36-17 | Reached per-engine argument/result/defect mutants | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36EngineCallMutationControls\|TestPhase36HelperAnswersAndDefects)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-07-01 | 07 | 6 | CALL-01, CALL-02 | T-36-20 | Independent exact source-free manifest | unit + integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest ./internal/compiler/session -run '^(TestPhase36SurfaceDocument\|TestPhase36SurfaceManifestCallWitnesses)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-07-02 | 07 | 6 | CALL-01, CALL-02 | T-36-21 | Finite census/limits/bidirectional controls | unit + integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/surfacemanifest -run '^(TestPhase36SurfaceManifestCallCensus\|TestPhase36SurfaceManifestCallControls\|TestPhase36SurfaceLimitConsumers\|TestPhase36SurfaceDocument\|TestPhase34SurfaceManifestWitnesses\|TestPhase34SurfaceManifestCensus\|TestPhase34SurfaceManifestControls\|TestPhase34SurfaceLimitConsumers\|TestPhase35SurfaceManifestWitnesses\|TestPhase35SurfaceManifestCensus\|TestPhase35SurfaceManifestControls\|TestPhase35SurfaceLimitConsumers\|TestPhase34SurfaceDocument\|TestPhase34SurfaceLimits\|TestPhase35SurfaceDocument)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-08-01 | 08 | 7 | CALL-01, CALL-02 | T-36-22 | Frozen final C/ledger/legacy safety | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/cgen -run '^(TestPhase36NoCallGeneratedCByteBaseline\|TestPhase36NoCallGeneratedCByteBaselineControls\|TestPhase36LegacyByteViewCalleeBounds\|TestPhase29InterpreterFixtureVectors\|TestPhase29InterpreterDefectWithLiveOwner\|TestPhase29NativeFixtureVectors\|TestPhase29FixtureMainEmission\|TestPhase273RealCapacityExit\|TestPhase273AppVerifySumToNWithinCapacity\|TestPhase273AppVerifySumToNExhausted\|TestInvocationPathTableBoundary\|TestInvocationPreflightGuardIsNotInert)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |
| 36-08-02 | 08 | 7 | CALL-01, CALL-02 | T-36-23 | Per-knower evidence and living transition | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperAnswersAndDefects\|TestPhase36SurfaceManifestCallWitnesses\|TestPhase36SurfaceManifestCallControls)$' -count=1 -v -timeout=600s` | New roots W0 missing; inherited roots source-inspected | pending |

## Wave 0 Requirements

- [ ] Commit literal enumerated pre-change C corpus and reached comparator controls before compiler edits.
- [ ] Create grammar/lowering/all-four-signature positive and negative fixtures.
- [ ] Create independent raw core/origin refusal-removal tests with no-op/restored baselines.
- [ ] Create repeated/nested/two-site invocation fixtures and forged trace controls.
- [ ] Prove bounded native path-storage constructibility, exact identity and exhaustion before committing to its representation.
- [ ] Add engine exact-answer and dynamic-callee defect controls at O0/O3/O3-LTO; preserve separate legacy ByteView bounds regression.
- [ ] Add manifest witness/census and restriction drift controls.
- [ ] Resolve existing inherited selector names from RESEARCH/source and retain only affected focused guards in finalized plans.

### Exact missing selector ownership

Every new path below is planned creation, not an existing source witness. No implementation selector was run during planning. "Wave 0" means create the named test in its owning task before changing behavior/claiming its verify; it does not authorize source changes before the baseline safety commit.

| Owner task | New file | Missing root selectors to create |
|---|---|---|
| 36-01-01/02 | internal/compiler/session/session_phase36_prechange_test.go | TestPhase36NoCallGeneratedCByteBaseline; TestPhase36NoCallGeneratedCByteBaselineControls |
| 36-02-01 | internal/compiler/syntax/parser_phase36_test.go | TestPhase36ScalarCallGrammar; TestPhase36ScalarCallDepthAndRecovery; TestPhase36ScalarCallFormatting |
| 36-02-02/03 | internal/compiler/check/check_phase36_scalar_calls_test.go | TestPhase36ScalarCallLoweringAndTypePairs; TestPhase36CallSourceIntentControls; TestPhase36CallImportAndRefusalBoundaries |
| 36-03-01 | internal/compiler/corevalidate/corevalidate_phase36_test.go | TestPhase36RawCallShapeAndFreshness; TestPhase36RawCallRefusalRemovalControls |
| 36-03-02 | internal/compiler/originvalidate/originvalidate_phase36_test.go | TestPhase36OriginCallFreshnessAndTransfer; TestPhase36OriginCallRefusalRemovalControls |
| 36-03-03 | internal/compiler/pathoracle/pathoracle_phase36_test.go | TestPhase36ScalarCallAuthorityBoundaries |
| 36-04-01/02/03 | internal/compiler/interp/interp_phase36_test.go | TestPhase36InterpreterRepeatedCallTracer; TestPhase36InterpreterCallAnswersAndDefects; TestPhase36InterpreterEvidenceBoundaries |
| 36-04-01/02/03 | internal/compiler/executionpeer/executionpeer_phase36_test.go | TestPhase36DynamicCallOrdinalControls; TestPhase36CallCausalAttributionControls |
| 36-05-01/02/03 | internal/compiler/cgen/cgen_phase36_test.go | TestPhase36NativeDynamicInvocationPaths; TestPhase36CallEventCapacityBoundary; TestPhase36NativeOrdinalAndPathBoundaries; TestPhase36EmitterCallInventoryControls |
| 36-05-02 | internal/compiler/native/native_phase36_test.go | TestPhase36CaptureOutcomeSeparation |
| 36-06-01/02 | internal/compiler/session/session_phase36_calls_test.go | TestPhase36HelperAnswersAndDefects; TestPhase36HelperOptimizationMatrix; TestPhase36HelperUBSan; TestPhase36ByteViewCallerArgumentBoundary; TestPhase36LegacyByteViewCalleeBounds; TestPhase36NativeCallCaptureOutcomeSeparation |
| 36-06-03 | internal/compiler/session/session_phase36_peer_mutation_test.go | TestPhase36EngineCallMutationControls |
| 36-06-02 | cmd/schway/main_phase36_test.go | TestPhase36PublicHelperConformance |
| 36-07-01 | internal/compiler/surfacemanifest/manifest_test.go (existing file, new root) | TestPhase36SurfaceDocument |
| 36-07-01/02 | internal/compiler/session/session_phase36_surface_test.go | TestPhase36SurfaceManifestCallWitnesses; TestPhase36SurfaceManifestCallCensus; TestPhase36SurfaceManifestCallControls; TestPhase36SurfaceLimitConsumers |

### Hard ordering and boundary rules

- Baseline literal corpus plus reached drift/omission/false-exemption controls must be committed before any compiler edit. Frozen old expected C and Phase34/35 witness/census bytes stay immutable. Exact new-source and newly admitted old-refusal rows use a bounded path/hash/route/reason ledger, never broad exemptions.
- The planning-only C constructibility probe is recorded in 36-CONSTRUCTIBILITY.md: O0 ASan/UBSan and O3-LTO scratch exit 0, deliberate conformance capacity exit 69. This is not a generated-C/feature pass. Actual tools newly observed were Go 1.24.0 darwin/arm64 and Apple Clang 21.0.0; research's different Clang label is historical text.
- Interpreter source audit found no event-retention cap in the current driver; 36-04 owns a new-call-only bounded collector and actual result plus named incomplete-evidence error. Native 36-05 owns stable append-only path bytes, per-frame/site checked ordinals and capture-disabled execution. Neither uses static 4096 as a dynamic semantic work limit.
- Every real call/return/failure/defect/event and invocation text counts toward its capture budget. Event, path, encoded-byte and ordinal boundaries get separate exact/plus-one controls. 16 MiB execution documents, 64 KiB app capture, application output and depth128 remain distinct.
- New overflow/remainder rows require dynamic U64/Bool callee execution, actual reason/65 and absent continuation. TestPhase36LegacyByteViewCalleeBounds explicitly executes old owning-wrapper/ByteView OOB on interpreter fixture/file and native application C at O0/O3/O3-LTO; old vector selectors alone do not prove this. TestPhase36ByteViewCallerArgumentBoundary separately proves caller OOB prevents helper entry.
- Check owns actual source selection; core/origin own valid raw contracts; each engine owns numeric answer/defect; executionpeer owns only observed causal attribution. Fresh no-op/restored positives plus the intended named failure are required for each seeded semantic/control mutation.
- Native optimization/sanitizer lanes must report actual execution; zero-match or skip does not count. Full suites, whole-package runs, broad native cross-program matrices, push and hosted dispatch remain prohibited locally. No human UAT substitutes for these deterministic checks.

### Explicit unresolved planning flags

Spec-less CALL-01/CALL-02 edge rows remain unclassified/unresolved; no SPEC classification is invented. Plan36-08 carries the two descriptorless projectProhibitions projections as flagged unverified (optional evidence cannot replace actual outcomes; causal acceptance cannot claim numeric/source-intent proof). There is no wired descriptor/enforcement claim. The source/requirement/research/D-01..D-26 coverage audit lives in 36-08-PLAN.md and covers all in-scope items; these flags remain separate from implementation evidence.

## Manual-Only Verifications

None expected. Human readability, whole-loop performance distributions and
external hosted receipts are not inferred from these automated tests.

## Validation Sign-Off

- [ ] Final task/wave/threat mapping matches executable plans.
- [ ] Every runnable verify has a stated observable failing direction.
- [ ] Missing tests created before their selectors earn credit.
- [ ] Every requirement and declared mitigation has named passing evidence.
- [ ] Mutation failures reach the intended assertion and restored positives pass.
- [ ] `status: validated`, `nyquist_compliant: true`, and `wave_0_complete: true` only after earned audit evidence.

**Approval:** pending implementation and validation audit.

## Planning verification receipt — 2026-10-14

Planning checks only: all eight plan frontmatter/structure queries returned valid with zero warnings; the static ownership/selector audit found all inherited roots in their declared packages, every task file in frontmatter, no task above five files, no same-wave file conflict, all26 D IDs in meaningful action bodies and23 unique non-SC threat IDs. Calibration returned factor1, sample_count0, confidence low; emitted token estimates retain raw projections unchanged. The source audit is in 36-08-PLAN.md.

The external-integration detector examined the eight plans plus Phase36 roadmap section and returned detected:false (its negative-result exit1 is expected). No external service/database/AI/UI surface is introduced. The assumption-delta detector returned detected:true for "another"; 36-04 records promotion of dynamic parent/site activation identity, retaining static ordinal-zero paths as first activations. Graphify is configured disabled and no project skill directory was found. The planning prohibition projection used the existing projectProhibitions contract once; no descriptors were invented. The two spec-less edge rows remain explicitly unresolved.

At this preplanning transition the relevant z1g/trz/z1s seed metadata and full seed/appraisal text were consulted against current source anchors already named in CONTEXT. All three retain dormant disposition: no measured crc_step navigation gap, unexplained failure beyond bounded capture, or concrete source-form dispute is established. The eight plans retain helper composition36 → file CRC/second consumer/interpreter CLI/guide37 → final audited two-host binding38. Task36-08-02 owns the next evidence-bound living amendment; historical claims are not rewritten.

No Phase36 Go/compiler behavior selector ran during planning. New selectors/fixtures above remain missing and pending implementation. Only the expressly scoped scratch C storage probe and independent scalar arithmetic calculation were executed; their limited results cannot close implementation, security, validation, readability, performance or hosted evidence gates.

---
status: complete
phase: 04-fallible-resources-and-c-boundary
source: 04-01-SUMMARY.md, 04-02-SUMMARY.md, 04-03-SUMMARY.md, 04-04-SUMMARY.md, 04-05-SUMMARY.md, 04-06-SUMMARY.md, 04-07-SUMMARY.md
started: 2026-09-05T14:26:54Z
updated: 2026-09-05T14:42:56Z
---

## Current Test

[testing complete]

## Tests

### 1. core.AllOperationKinds()/TerminatorKinds() registry; six-site exhaustive-dispatch control green on the five pre-Phase-4 kinds
expected: core.AllOperationKinds()/TerminatorKinds() registry; six-site exhaustive-dispatch control green on the five pre-Phase-4 kinds
result: pass
source: automated
coverage_id: 04-01/D1
verification: internal/compiler/core/core_test.go#TestAllOperationKindsRegistered; internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite

### 2. Pre-Phase-4 core bytes and evidence manifest IDs pinned byte-for-byte against a phase-start golden
expected: Pre-Phase-4 core bytes and evidence manifest IDs pinned byte-for-byte against a phase-start golden
result: pass
source: automated
coverage_id: 04-01/D2
verification: internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged; internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged

### 3. native.validateExecution carries an explicit expected-terminal-outcome axis; every original Phase 1-3 rejection ground still rejects
expected: native.validateExecution carries an explicit expected-terminal-outcome axis; every original Phase 1-3 rejection ground still rejects
result: pass
source: automated
coverage_id: 04-01/D3
verification: internal/compiler/native/native_test.go#TestValidateExecutionStillRejectsOldGrounds; internal/compiler/native/native_test.go#TestValidateExecutionTypedFailureDiscriminates

### 4. `foreign C { }` block and `try <callee>(<args>)` parse losslessly, format to a fixed point, and reparse to the same semantic token projection; a bare fallible call is refused at parse time
expected: `foreign C { }` block and `try <callee>(<args>)` parse losslessly, format to a fixed point, and reparse to the same semantic token projection; a bare fallible call is refused at parse time
result: pass
source: automated
coverage_id: 04-01/D4
verification: internal/compiler/syntax/syntax_test.go#TestForeignCallRoundTrips; internal/compiler/syntax/syntax_test.go#TestFallibleCallUnconsumedRejected

### 5. One Lang program declaring a foreign C symbol, calling it fallibly through try, checks into a three-block/two-edge CFG with a populated ForeignContract
expected: One Lang program declaring a foreign C symbol, calling it fallibly through try, checks into a three-block/two-edge CFG with a populated ForeignContract
result: pass
source: automated
coverage_id: 04-01/D5
verification: internal/compiler/check/check_test.go#TestForeignCallLowersToOkAndErrEdges

### 6. Interpreter and Clang-built native code (-O0 and -O3) agree on terminal outcome, ordered events, and live-resource state for the tracer fixture
expected: Interpreter and Clang-built native code (-O0 and -O3) agree on terminal outcome, ordered events, and live-resource state for the tracer fixture
result: pass
source: automated
coverage_id: 04-01/D6
verification: internal/compiler/session/session_test.go#TestForeignCallInterpreterNative

### 7. Two inadmissible foreign shapes (missing unwind/nonlocal_exit policy; call target resolving to a Lang function) are refused independently by check and corevalidate, and visible to the session gate as required negative controls
expected: Two inadmissible foreign shapes (missing unwind/nonlocal_exit policy; call target resolving to a Lang function) are refused independently by check and corevalidate, and visible to the session gate as required negative controls
result: pass
source: automated
coverage_id: 04-01/D7
verification: internal/compiler/check/check_test.go#TestUnwindPolicyUndeclaredRejected; internal/compiler/check/check_test.go#TestCallTargetNotForeignRejected; internal/compiler/corevalidate/corevalidate_test.go#TestForeignRefusalsAreIndependentlyDerived; internal/compiler/session/session_test.go#TestVerifyPhase4ForeignControls

### 8. The shipped ./cmd/lang binary runs the tracer fixture and one hand-written, non-corpus program cleanly through format --check, check, run --engine=interpreter, and run --engine=native
expected: The shipped ./cmd/lang binary runs the tracer fixture and one hand-written, non-corpus program cleanly through format --check, check, run --engine=interpreter, and run --engine=native
result: pass
source: automated
coverage_id: 04-01/D8
verification: internal/compiler/native/native_test.go#TestShippedBinaryFourSubcommandCorpusMatrix; internal/compiler/native/native_test.go#TestShippedBinaryExercisesEveryPhase4Behavior; .github/workflows/ci.yml (checks + phase-4 gate, ubuntu-latest and macos-latest)

### 9. Three-stage partial initialization releases only completed acquisitions, exactly once, in reverse order, on success and on both failure stages
expected: Three-stage partial initialization releases only completed acquisitions, exactly once, in reverse order, on success and on both failure stages
result: pass
source: automated
coverage_id: 04-02/D1
verification: internal/compiler/check/check_test.go#TestThreeAcquisitionReleaseOrder; internal/compiler/check/check_test.go#TestPartialAcquisitionReleasesOnlyCompleted; internal/compiler/session/session_test.go#TestReleaseInterpreterNative

### 10. discard ... because is the second admissible fallible-call consumer, round-tripping through format and carrying a required non-empty rationale into the core artifact
expected: discard ... because is the second admissible fallible-call consumer, round-tripping through format and carrying a required non-empty rationale into the core artifact
result: pass
source: automated
coverage_id: 04-02/D2
verification: internal/compiler/check/check_test.go#TestDiscardBecauseRoundTrips; internal/compiler/check/check_test.go#TestDiscardRationaleRequired

### 11. A second, independently written derivation of the release order (backward from failure/return edges) agrees with the shipped artifact on every fixture and disagrees with every mutation of it, at honestly counted, increasing cost across sizes
expected: A second, independently written derivation of the release order (backward from failure/return edges) agrees with the shipped artifact on every fixture and disagrees with every mutation of it, at honestly counted, increasing cost across sizes
result: pass
source: automated
coverage_id: 04-02/D3
verification: internal/compiler/corevalidate/corevalidate_test.go#TestValidatorRederivesReleaseOrder; internal/compiler/corevalidate/corevalidate_test.go#TestReleaseOrderMutationMatrix; internal/compiler/corevalidate/corevalidate_test.go#TestReleaseOrderValidationWorkSeries

### 12. Release order and release set are each falsified by a mutation, the two mutations attack different artifacts, and the three-acquisition requirement is demonstrated by a test rather than argued in prose
expected: Release order and release set are each falsified by a mutation, the two mutations attack different artifacts, and the three-acquisition requirement is demonstrated by a test rather than argued in prose
result: pass
source: automated
coverage_id: 04-02/D4
verification: internal/compiler/session/session_test.go#TestReleaseOmissionMutationIsMismatch; internal/compiler/session/session_test.go#TestReleaseTranspositionMutationIsMismatch; internal/compiler/session/session_test.go#TestTwoAcquisitionTranspositionIsIndistinguishable; internal/compiler/session/session_test.go#TestReleaseMutationsAttackDifferentArtifacts; internal/compiler/session/session_test.go#TestVerifyPhase4ReleaseControls

### 13. The shipped ./cmd/lang binary runs all four Phase 4 plan-02 fixtures and one hand-written, non-corpus three-acquisition program cleanly through format --check, check, run --engine=interpreter, and run --engine=native
expected: The shipped ./cmd/lang binary runs all four Phase 4 plan-02 fixtures and one hand-written, non-corpus three-acquisition program cleanly through format --check, check, run --engine=interpreter, and run --engine=native
result: pass
source: automated
coverage_id: 04-02/D5
verification: internal/compiler/native/native_test.go#TestShippedBinaryFourSubcommandCorpusMatrix; internal/compiler/native/native_test.go#TestShippedBinaryExercisesEveryPhase4Behavior; .github/workflows/ci.yml (checks + phase-4 gate, ubuntu-latest and macos-latest)

### 14. core.ForeignContract carries every FFI-01 obligation category (layout, initialized state, allocator, capture, retention, aliasing, unwind) and is content-bound into evidence without moving any pre-existing manifest identifier
expected: core.ForeignContract carries every FFI-01 obligation category (layout, initialized state, allocator, capture, retention, aliasing, unwind) and is content-bound into evidence without moving any pre-existing manifest identifier
result: pass
source: automated
coverage_id: 04-03/D1
verification: internal/compiler/check/check_test.go#TestForeignContractCarriesEveryObligation; internal/compiler/check/check_test.go#TestForeignFieldsAreOmittedWhenAbsent; internal/compiler/evidence/evidence_test.go#TestForeignSidecarManifestDigestBinds; internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged; internal/compiler/corevalidate/corevalidate_test.go#TestAllocatorIdentityMismatchRejected; internal/compiler/corevalidate/corevalidate_test.go#TestForeignContractInternallyValidated

### 15. Three inspectable layers (sidecar manifest, generated header, conformance unit) are all generated from one contract, and the conformance unit compiles on its own, never linked, over the real frozen private header
expected: Three inspectable layers (sidecar manifest, generated header, conformance unit) are all generated from one contract, and the conformance unit compiles on its own, never linked, over the real frozen private header
result: pass
source: automated
coverage_id: 04-03/D2
verification: internal/compiler/cgen/cgen_test.go#TestGeneratedForeignHeaderNamesAreAllocated; internal/compiler/cgen/cgen_test.go#TestObligationCommentsAreGeneratedFromJSON; internal/compiler/cgen/cgen_test.go#TestConformanceUnitAssertsEveryField; internal/compiler/native/native_conformance_test.go#TestConformanceUnitCompilesSeparately; internal/compiler/cgen/cgen_test.go#TestPrivateHeaderIsIncludedOnlyByConformanceUnit

### 16. A field transposition in a frozen boundary fixture is a compile-time refusal, and zero optimizer-visible attributes are emitted, mutation-killed by injection
expected: A field transposition in a frozen boundary fixture is a compile-time refusal, and zero optimizer-visible attributes are emitted, mutation-killed by injection
result: pass
source: automated
coverage_id: 04-03/D3
verification: internal/compiler/session/session_test.go#TestLayoutMutationIsCompileTimeRefusal; internal/compiler/session/session_test.go#TestLayoutMutationAttacksFrozenFixtureOnly; internal/compiler/session/session_test.go#TestNoUnprovenAttributesEmitted; internal/compiler/session/session_test.go#TestAttributeInjectionMakesControlFail; internal/compiler/session/session_test.go#TestNoreturnExemptionIsNamed; internal/compiler/session/session_test.go#TestVerifyPhase4ForeignLayoutControls

### 17. A real, reachable, deliberately terminal defect operation exists (core.OpDefect), admissible only in a match arm's terminal position, aborting via a generated _Noreturn lang_defect function whose every path ends in abort(), with no route into typed_failure
expected: A real, reachable, deliberately terminal defect operation exists (core.OpDefect), admissible only in a match arm's terminal position, aborting via a generated _Noreturn lang_defect function whose every path ends in abort(), with no route into typed_failure
result: pass
source: automated
coverage_id: 04-04/D1
verification: internal/compiler/syntax/syntax_test.go#TestDefectTerminatorRoundTrips; internal/compiler/check/check_test.go#TestDefectLowersToTerminalOutcome; internal/compiler/core/core_test.go#TestCancelledOutcomeIsUnconstructible; internal/compiler/core/core_test.go#TestNoErrorValueConstructorExists; internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite; internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged

### 18. An additive streaming event emitter writes each event at the point it occurs for every foreign-acquiring function, leaving every other emitter (and every Phase 1/2/3 golden) byte-identical; native.go reports a terminal-record-absence code distinct from cap truncation
expected: An additive streaming event emitter writes each event at the point it occurs for every foreign-acquiring function, leaving every other emitter (and every Phase 1/2/3 golden) byte-identical; native.go reports a terminal-record-absence code distinct from cap truncation
result: pass
source: automated
coverage_id: 04-04/D2
verification: internal/compiler/cgen/cgen_test.go#TestStreamingEmitterWritesAtPointOfOccurrence; internal/compiler/cgen/cgen_test.go#TestExistingEmittersAreByteIdentical; internal/compiler/native/native_test.go#TestAbortingProgramStreamsEventsBeforeDying; internal/compiler/native/native_test.go#TestTerminalRecordAbsenceIsHardFailure; internal/compiler/native/native_test.go#TestTruncationAndAbsenceReportDistinctCodes

### 19. A dying program is adjudicated by its real wait status (never a hardcoded exit code), and the rule that a defect path runs no cleanup is an enforced, mutation-killed control alongside a real interpreter/native signal-adjudication differential
expected: A dying program is adjudicated by its real wait status (never a hardcoded exit code), and the rule that a defect path runs no cleanup is an enforced, mutation-killed control alongside a real interpreter/native signal-adjudication differential
result: pass
source: automated
coverage_id: 04-04/D3
verification: internal/compiler/native/native_test.go#TestAbortSignalAdjudicatedByWaitStatus; internal/compiler/native/native_test.go#TestNonzeroExitIsDistinctFromSignal; internal/compiler/native/native_test.go#TestDefectExpectationRejectsReturnedDocument; internal/compiler/session/session_test.go#TestNoReleaseAfterDefect; internal/compiler/session/session_test.go#TestVerifyPhase4DefectControls

### 20. One process-root setjmp landing pad, reading a static-storage resource ledger, emits one foreign.nonlocal_exit event plus one resource.leaked event per still-live acquisition, then a function.defected terminal record, then aborts -- with a second frozen foreign TU giving it a reachable witness through the real generated pipeline, and the interpreter agreeing byte-for-byte via a shared documented convention
expected: One process-root setjmp landing pad, reading a static-storage resource ledger, emits one foreign.nonlocal_exit event plus one resource.leaked event per still-live acquisition, then a function.defected terminal record, then aborts -- with a second frozen foreign TU giving it a reachable witness through the real generated pipeline, and the interpreter agreeing byte-for-byte via a shared documented convention
result: pass
source: automated
coverage_id: 04-05/D1
verification: internal/compiler/cgen/cgen_test.go#TestExactlyOneLandingPadIsInstalled; internal/compiler/cgen/cgen_test.go#TestLedgerIsStaticStorage; internal/compiler/session/session_test.go#TestNonlocalExitEmitsLeakPerLiveAcquisition; internal/compiler/session/session_test.go#TestPadRunsNoRelease; internal/compiler/session/session_test.go#TestNonlocalExitProbeInterpreterNative

### 21. An nm -u undefined-symbol allowlist control: a positive, rationale-carrying list where any new undefined symbol fails until a human adds it, an honest operational status when the listing tool is absent, and no dependence on unwind-section presence
expected: An nm -u undefined-symbol allowlist control: a positive, rationale-carrying list where any new undefined symbol fails until a human adds it, an honest operational status when the listing tool is absent, and no dependence on unwind-section presence
result: pass
source: automated
coverage_id: 04-05/D2
verification: internal/compiler/native/symbols_test.go#TestUndefinedSymbolAllowlistRejectsNewSymbol; internal/compiler/native/symbols_test.go#TestUndefinedSymbolNormalizationHandlesBothFormats; internal/compiler/native/symbols_test.go#TestMissingSymbolToolReportsOperational; internal/compiler/native/symbols_test.go#TestUnwindControlDoesNotInspectSections; internal/compiler/session/session_test.go#TestVerifyPhase4UnwindControl

### 22. control:foreign.nonlocal_exit_undetected is mutation-killed two different ways (pad-span omission; ledger-population omission surfacing specifically as a leak-count disagreement), and the detection mechanism's reachable input space -- including two named, never-claimed-covered blind spots -- is recorded rather than implied covered
expected: control:foreign.nonlocal_exit_undetected is mutation-killed two different ways (pad-span omission; ledger-population omission surfacing specifically as a leak-count disagreement), and the detection mechanism's reachable input space -- including two named, never-claimed-covered blind spots -- is recorded rather than implied covered
result: pass
source: automated
coverage_id: 04-05/D3
verification: internal/compiler/session/session_test.go#TestNonlocalExitDetectionIsMutationKilled; internal/compiler/session/session_test.go#TestLeakCountMatchesLiveAcquisitions; internal/compiler/cgen/cgen_test.go#TestNonlocalExitReachabilityIsRecorded; internal/compiler/cgen/cgen_test.go#TestBlindSpotsAreNamedNotClaimed; internal/compiler/session/session_test.go#TestVerifyPhase4NonlocalExitControl

### 23. originvalidate and pathoracle both walk every terminator (return, typed failure, defect), read from core.TerminatorKinds(), mutation-killed independently in both packages
expected: originvalidate and pathoracle both walk every terminator (return, typed failure, defect), read from core.TerminatorKinds(), mutation-killed independently in both packages
result: pass
source: automated
coverage_id: 04-06/D1
verification: internal/compiler/originvalidate/originvalidate_test.go#TestOriginWalksEveryTerminator; internal/compiler/originvalidate/originvalidate_test.go#TestTerminatorSetReadFromRegistry; internal/compiler/originvalidate/originvalidate_test.go#TestTerminatorWalkMutationKilled; internal/compiler/pathoracle/pathoracle_test.go#TestPathOracleClosesOnEveryTerminator; internal/compiler/pathoracle/pathoracle_test.go#TestFailureOnlyPathIsChecked; internal/compiler/pathoracle/pathoracle_test.go#TestTerminatorWalkMutationKilled; internal/compiler/session/session_test.go#TestVerifyPhase4TerminatorControl

### 24. A foreign call declared to borrow/retain its argument is recognised as borrow-derived; an undeclared origin on such a return is refused with core.foreign_origin_omitted
expected: A foreign call declared to borrow/retain its argument is recognised as borrow-derived; an undeclared origin on such a return is refused with core.foreign_origin_omitted
result: pass
source: automated
coverage_id: 04-06/D2
verification: internal/compiler/originvalidate/originvalidate_test.go#TestForeignBorrowDerivedReturnRecognised; internal/compiler/originvalidate/originvalidate_test.go#TestForeignOriginOmittedRejected; internal/compiler/session/session_test.go#TestVerifyPhase4ForeignOriginControl; cmd/lang check testdata/phase4/foreign_origin_omitted.lang (exit 2, core.foreign_origin_omitted)

### 25. originvalidate.ValidatePublished runs on lang check and lang run, not only interface export; testdata/phase3/public_view_omitted.lang, which checked clean before this plan, is now refused
expected: originvalidate.ValidatePublished runs on lang check and lang run, not only interface export; testdata/phase3/public_view_omitted.lang, which checked clean before this plan, is now refused
result: pass
source: automated
coverage_id: 04-06/D3
verification: internal/compiler/session/session_test.go#TestPublishedOriginValidatedOnCheckAndRun; internal/compiler/originvalidate/originvalidate_test.go#TestOriginValidatorImportsStayIndependent

### 26. discoverLoanLastUses counts its own transitive-scan work; growth series over three chain lengths; before/after values recorded; no verdict changes over the whole corpus
expected: discoverLoanLastUses counts its own transitive-scan work; growth series over three chain lengths; before/after values recorded; no verdict changes over the whole corpus
result: pass
source: automated
coverage_id: 04-06/D4
verification: internal/compiler/check/check_test.go#TestLastUseDiscoveryWorkIsCounted; internal/compiler/check/check_test.go#TestLastUseDiscoveryWorkSeries; internal/compiler/check/check_test.go#TestCheckerVerdictsUnchanged; internal/compiler/check/check_test.go#TestOwnershipWorkSeries

### 27. 04-DEBT.md records the deferred discoverLoanLastUses retirement, the deferred payload-carrying-alternative work, and the accepted residual limitations, each dated with identifier/severity/source/landing phase
expected: 04-DEBT.md records the deferred discoverLoanLastUses retirement, the deferred payload-carrying-alternative work, and the accepted residual limitations, each dated with identifier/severity/source/landing phase
result: pass
source: automated
coverage_id: 04-06/D5
verification: internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed

### 28. scripts/verify-phase4.sh: a bounded gate that builds the shipped binary once, verifies Phase 1-4 corpora with it, and requires every Phase 4 control identifier and expected escape by exact text, contract-tested against its own script text
expected: scripts/verify-phase4.sh: a bounded gate that builds the shipped binary once, verifies Phase 1-4 corpora with it, and requires every Phase 4 control identifier and expected escape by exact text, contract-tested against its own script text
result: pass
source: automated
coverage_id: 04-07/D1
verification: internal/compiler/session/session_test.go#TestVerifyPhase4ControlsAndWork; internal/compiler/session/session_test.go#TestVerifyPhase4CLI; internal/compiler/session/session_test.go#TestPhase4VerifierScriptContract; internal/compiler/session/session_test.go#TestPhase4RequiredControlsMatchScript; internal/compiler/session/session_test.go#TestExpectedEscapesAreVisibleNotSolved; sh scripts/verify-phase4.sh

### 29. All three engines (interpreter, -O0, -O3) genuinely agree on terminal outcome, ordered events, and live-resource state across all five Phase 4 path shapes: success, second-stage typed failure, third-stage typed failure, defect, and nonlocal exit
expected: All three engines (interpreter, -O0, -O3) genuinely agree on terminal outcome, ordered events, and live-resource state across all five Phase 4 path shapes: success, second-stage typed failure, third-stage typed failure, defect, and nonlocal exit
result: pass
source: automated
coverage_id: 04-07/D2
verification: internal/compiler/session/session_test.go#TestPhase4CorpusThreeEngineAgreement; internal/compiler/session/session_test.go#TestPhase4DifferentialNamesFirstDisagreement; internal/compiler/evidence/evidence_test.go#TestForeignDigestMismatchRefused; internal/compiler/session/session_test.go#TestReleaseOmissionMutationIsMismatch; internal/compiler/session/session_test.go#TestReleaseTranspositionMutationIsMismatch

### 30. Every Phase 4 behavior is demonstrated on the shipped binary against a hand-written, out-of-corpus program; the debt and escape register is finalized and no test/comment/fixture claims coverage of a named residual
expected: Every Phase 4 behavior is demonstrated on the shipped binary against a hand-written, out-of-corpus program; the debt and escape register is finalized and no test/comment/fixture claims coverage of a named residual
result: pass
source: automated
coverage_id: 04-07/D3
verification: internal/compiler/native/native_test.go#TestShippedBinaryExercisesEveryPhase4Behavior; internal/compiler/session/session_test.go#TestNoCoverageClaimedForNamedResiduals; internal/compiler/session/session_test.go#TestPhase4ReachabilityRecordIsComplete; internal/compiler/native/native_test.go#TestOutOfCorpusSourcesAreGenuinelyNovel; internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed

## Summary

total: 30
passed: 30
issues: 0
pending: 0
skipped: 0
blocked: 0

## Notes

- Zero human checkpoints. Every Phase 4 deliverable is covered by a passing automated control, run in CI on Linux and macOS (`.github/workflows/ci.yml`).
- The four deliverables that previously required a human reading (04-01 D8, 04-02 D5, 04-06 D5, 04-07 D3) were mechanized during this session rather than answered by hand — see the Automation Record below.
- 04-07's `coverage:` block used inline flow mappings (`{kind: ..., ref: ...}`) that `uat classify-coverage` could not parse, which silently demoted two fully-automated deliverables (D1, D2) into human checkpoints. The block was reformatted to block mappings; no control changed.

## Automation Record

| Was | Now asserted by | Runs in CI |
|---|---|---|
| 04-01 D8 / 04-02 D5 — shipped binary driven by hand through four subcommands, exit codes retyped into the SUMMARY | `native_test.go#TestShippedBinaryFourSubcommandCorpusMatrix` (exhaustive over `testdata/phase4`, fails on an unlisted fixture) + `#TestShippedBinaryExercisesEveryPhase4Behavior` (now covers `format --check` on all eight out-of-corpus programs) | yes |
| 04-06 D5 — debt-register completeness read by a human | `session_test.go#TestDebtRegistersAreWellFormed` (every phase's register: declared count, closed severity vocabulary, required columns, row↔detail-section pairing) | yes |
| 04-07 D3 — out-of-corpus novelty judged by a human | `native_test.go#TestOutOfCorpusSourcesAreGenuinelyNovel` (no hand-written program is byte-identical to any corpus fixture) | yes |

**Deliberately left un-mechanized** (recorded, not claimed): whether an out-of-corpus program is *interestingly* novel rather than merely non-identical, and whether debt-register prose is honestly phrased. The dangerous direction of the latter — a named residual described as covered — is already mechanized by `TestNoCoverageClaimedForNamedResiduals`.

## Re-Verification

- **2026-09-05T14:42:56Z** — re-ran every covering control from a clean runtime state; no test, fixture, or
  source file was changed by this session.
  - `go test ./...` — all 16 test packages pass (`native` 13.7s, `session` 38.2s).
  - `sh scripts/verify-phase4.sh` — exit 0. All four phase corpora verify `status: pass`
    on the freshly built binary; Phase 4's lane set reports 14 lanes green, including
    `lane:foreign-no-unproven-attributes`, `lane:release-omitted`,
    `lane:release-order-transposed`, `lane:defect-signal-adjudicated`,
    `lane:nonlocal-exit-undetected`, and `lane:kind-exhaustive-dispatch`.
  - `uat classify-coverage` re-run on all seven SUMMARYs: `mode: coverage`,
    30/30 `all_auto_covered: true`, zero human checkpoints, zero parse errors.
  - `verify:pre` gate `api-coverage.verify-pre`: `block: false` (no external-API
    integration detected).

  UAT result is unchanged: 30 passed, 0 issues. Note that phase advancement remains
  blocked by 04-VERIFICATION.md (`status: gaps_found`) — see Gaps below.

## Gaps

UAT found no gaps. Phase-goal verification did — these two are recorded in
04-VERIFICATION.md, not discovered by this UAT session, and are reproduced here
so `/gsd-plan-phase 04 --gaps` and the completion predicate agree on what is open.

- gap_id: G-04-V1
  truth: "`corevalidate` independently rederives the expected release order by walking backward from each failure edge over the block and edge graph, and compares (D-04-07, D-12a)."
  status: failed
  severity: major
  source: 04-VERIFICATION.md
  reason: "checkReleaseOrder silently skips any terminal block whose incoming-edge count != 1 instead of refusing the shape; no other check bounds terminal blocks to one incoming edge, so a merge-point terminal block would go entirely unchecked by the control D-04-07 requires to be an independent rederivation."
  artifacts:
    - path: internal/compiler/corevalidate/corevalidate.go
      issue: "checkReleaseOrder (~line 1242) skips rather than refuses on unexpected terminal incoming-edge count"
  missing:
    - "Treat an unexpected incoming-edge count on a terminal block as a hard refusal in checkReleaseOrder"
    - "Add a structural check that every OpReturn/OpFail-terminated block has exactly one incoming edge, independent of checkReleaseOrder"

- gap_id: G-04-V2
  truth: "One authoritative `core.ForeignContract` carries the boundary facts and all three inspectable layers derive from it, enforced by `control:foreign.no_unproven_attributes` scanning all emitted C (FFI-01, D-04-12, D-04-13)."
  status: failed
  severity: major
  source: 04-VERIFICATION.md
  reason: "lane:foreign-no-unproven-attributes scans only cgen.Emit and cgen.EmitForeignManifest output. EmitForeignHeader and EmitForeignConformance — two of the three named inspectable layers — are never passed to ScanForBannedAttributes anywhere in the repo, so the control asserts coverage it does not have."
  artifacts:
    - path: internal/compiler/session/session.go
      issue: "lane:foreign-no-unproven-attributes (~lines 2178-2212) never scans EmitForeignHeader/EmitForeignConformance output"
    - path: internal/compiler/cgen/cgen.go
      issue: "EmitForeignHeader (~line 1270) and EmitForeignConformance (~line 1327) are unscanned inspectable layers"
  missing:
    - "Feed cgen.EmitForeignHeader and cgen.EmitForeignConformance output into the lane:foreign-no-unproven-attributes scan"
    - "Add a mutation-kill test injecting a banned token into EmitForeignHeader only, to prove the two artifacts are independently covered"

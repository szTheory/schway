---
phase: "04"
slug: "fallible-resources-and-c-boundary"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-04"
---

# Phase 4 — Validation Strategy

> Feedback contract for proving one fallible foreign call, a three-stage resource lifecycle with reverse-order release, an audited and layout-proven C boundary, a real terminal defect, and detected foreign nonlocal exit — across the real compiler spine and the shipped binary.
>
> Every checkbox in this document is deliberately **unchecked**. A box is ticked only when execution has supplied the evidence named beside it. `nyquist_compliant` and `wave_0_complete` stay `false` until then.

## Scope Limitations Recorded Up Front

Four fences are declared before any evidence is collected, and no test, comment, fixture, or report in this phase may claim past them.

1. **Foreign calls only.** There is no Lang-to-Lang call construct this phase (D-04-01). Every admission-layer differential written here is single-function by construction, and rebuilding them cross-function is the next phase's cost when calls arrive. The lift condition is recorded as callable being a subset of publishable (D-04-03).
2. **Typed failure is an edge, not a value.** No `Result` type, no generics, no payload-carrying alternatives; the nullary-alternative shape is unchanged (D-04-04). The payload-carrying-alternative work is recorded as deferred debt, not discovered later.
3. **Abort-only panic.** The defect terminator has no catch, no containment, no unwinding, and no cleanup (D-04-15). Containment, isolation boundaries, and cancellation are a later milestone; `cancelled` is reserved and provably unconstructible so adding it later is a reviewed change rather than a quiet retrofit.
4. **One small by-value record on one host.** Unions, bit fields, vectors, variadics, packed and aligned records, and aggregate-passing thresholds remain open, as do the other object format, the other architecture, and the other compilers. Spike 005's Known Limitations fence this and Phase 4 inherits it verbatim.

Success criterion 3's second half is delivered as **detection, not prevention**. Prevention against opaque C is unachievable and claiming it would be an unproven promise of the same class as a false optimizer attribute. Prevention *is* delivered statically at the declaration layer, where a foreign declaration that does not state its policies is refused admission with no default (D-04-16).

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 `testing`, `testing/quick`, native fuzz seeds, black-box `os/exec` |
| **Config file** | `go.mod` — no new module requirement is permitted this phase |
| **Exact-target selector** | `scripts/assert-go-tests.sh` (existing; `--self-test` proves the selector is not vacuous before any targeted command is trusted) |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./...` |
| **Full suite command** | `sh scripts/verify-phase4.sh` |
| **Native toolchain** | Apple Clang 21.0.0 arm64, invoked with the existing pinned flag set; the frozen foreign translation unit and the generated conformance unit are separate invocations |
| **Symbol tool** | `nm` for the undefined-symbol allowlist; its absence must report operational, never pass |
| **Feedback budget** | quick under 30 seconds; full under 150 seconds on the current macOS arm64 host — observed, never ratified as a release SLO |

## Sampling Rate

- **After every task commit:** run that task's exact `<automated>` command, then `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./...`.
- **After every plan wave:** `go test -race ./...` and `go vet ./...`, plus `sh scripts/verify-phase3.sh` to prove Phase 3 non-regression while Phase 4's own gate does not yet exist.
- **Before phase verification:** `sh scripts/verify-phase4.sh` must exit zero and must report every required Phase 4 control identifier and every expected escape.
- **Investigation only:** fuzzing is time-bounded and never enters the deterministic default gate.
- **Max feedback latency:** the per-task command must complete in under 60 seconds; a command that does not is split rather than tolerated.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | SEM-03/FFI-01 | T-04-01/T-04-06 | Pre-Phase-4 bytes and manifest identifiers are pinned; one kind registry backs a six-site dispatch control; the native harness can express a non-return terminal outcome without loosening any prior rejection | regression pin + dispatch table | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestPreviousPhaseCoreBytesUnchanged TestPreviousPhaseManifestIDsUnchanged TestAllOperationKindsRegistered TestAllOperationKindsHandledAtEverySite TestValidateExecutionStillRejectsOldGrounds` | ✅ Wave 0 | ✅ green |
| 04-01-02 | 01 | 1 | SEM-03 | — | The one-way edge-versus-value door is confirmed before the tracer commits the shape | decision checkpoint (no automated verification by design) | none — `checkpoint:decision`, gated on a human choice | n/a | ✅ green |
| 04-01-03 | 01 | 1 | SEM-03/FFI-01 | T-04-02/T-04-04 | One fallible foreign call travels source to native across a frozen, never-parsed C boundary and agrees across engines | vertical tracer | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestForeignCallRoundTrips TestForeignCallLowersToOkAndErrEdges TestFallibleCallUnconsumedRejected TestForeignCallInterpreterNative TestAllOperationKindsHandledAtEverySite TestPreviousPhaseCoreBytesUnchanged` | ✅ Wave 0 | ✅ green |
| 04-01-04 | 01 | 1 | FFI-01 | T-04-02/T-04-03/T-04-05 | A policy-less foreign declaration and a Lang-targeted call are each refused by two independently derived layers, and the surface is capped fail-closed | compile-reject matrix + independence | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestUnwindPolicyUndeclaredRejected TestCallTargetNotForeignRejected TestForeignRefusalsAreIndependentlyDerived TestValidatorImportsStayIndependent TestForeignAdmissionCapsRejectFailClosed TestVerifyPhase4ForeignControls` | ✅ Wave 0 | ✅ green |
| 04-02-01 | 02 | 2 | RES-01/SEM-03 | T-04-08/T-04-10/T-04-13 | Three-stage partial initialization releases only completed acquisitions, exactly once, in reverse order, on success and both failure stages | differential over three engines | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestThreeAcquisitionReleaseOrder TestPartialAcquisitionReleasesOnlyCompleted TestDiscardBecauseRoundTrips TestDiscardRationaleRequired TestReleaseInterpreterNative TestAllOperationKindsHandledAtEverySite` | ✅ Wave 0 | ✅ green |
| 04-02-02 | 02 | 2 | RES-01 | T-04-11/T-04-12 | A second derivation walks backward from failure edges and disagrees with every mutation of the materialized order | independent re-derivation | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestValidatorRederivesReleaseOrder TestReleaseOrderMutationMatrix TestValidatorImportsStayIndependent TestCoreValidationWorkSeries` | ✅ Wave 0 | ✅ green |
| 04-02-03 | 02 | 2 | RES-01 | T-04-08/T-04-09 | Transposition and omission each turn the differential red, attacking two different artifacts | mutation kill | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestReleaseOmissionMutationIsMismatch TestReleaseTranspositionMutationIsMismatch TestTwoAcquisitionTranspositionIsIndistinguishable TestReleaseMutationsAttackDifferentArtifacts TestVerifyPhase4ReleaseControls` | ✅ Wave 0 | ✅ green |
| 04-03-01 | 03 | 3 | FFI-01 | T-04-18 | One authoritative contract carries every obligation and binds a sidecar digest without moving a pre-existing identifier | schema additivity + identity pin | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestForeignContractCarriesEveryObligation TestForeignFieldsAreOmittedWhenAbsent TestForeignSidecarManifestDigestBinds TestPreviousPhaseManifestIDsUnchanged TestAllocatorIdentityMismatchRejected TestForeignContractInternallyValidated` | ✅ Wave 0 | ✅ green |
| 04-03-02 | 03 | 3 | FFI-01 | T-04-16/T-04-17/T-04-19 | Three inspectable layers are generated from one contract, and the conformance unit compiles on its own and is never linked | generated-artifact contract | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestGeneratedForeignHeaderNamesAreAllocated TestObligationCommentsAreGeneratedFromJSON TestConformanceUnitAssertsEveryField TestConformanceUnitCompilesSeparately TestPrivateHeaderIsIncludedOnlyByConformanceUnit` | ✅ Wave 0 | ✅ green |
| 04-03-03 | 03 | 3 | FFI-01 | T-04-14/T-04-15 | A field transposition in the frozen boundary is a compile-time refusal, and zero optimizer-visible attributes are emitted | mutation kill (two artifacts) | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestLayoutMutationIsCompileTimeRefusal TestLayoutMutationAttacksFrozenFixtureOnly TestNoUnprovenAttributesEmitted TestAttributeInjectionMakesControlFail TestNoreturnExemptionIsNamed TestVerifyPhase4ForeignLayoutControls` | ✅ Wave 0 | ✅ green |
| 04-04-01 | 04 | 4 | SEM-03 | T-04-20/T-04-25 | A reachable abort-only defect exists, the outcome axis is closed, and no IR route carries a panic into typed failure | structural contract | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestDefectTerminatorRoundTrips TestDefectLowersToTerminalOutcome TestCancelledOutcomeIsUnconstructible TestNoErrorValueConstructorExists TestAllOperationKindsHandledAtEverySite TestPreviousPhaseCoreBytesUnchanged` | ✅ Wave 0 | ✅ green |
| 04-04-02 | 04 | 4 | SEM-03/RES-01 | T-04-21/T-04-22 | An aborting process emits every event it produced plus a terminal record, and a missing terminal record fails with its own code | streaming-emitter contract | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestStreamingEmitterWritesAtPointOfOccurrence TestAbortingProgramStreamsEventsBeforeDying TestTerminalRecordAbsenceIsHardFailure TestTruncationAndAbsenceReportDistinctCodes TestExistingEmittersAreByteIdentical` | ✅ Wave 0 | ✅ green |
| 04-04-03 | 04 | 4 | SEM-03/RES-01 | T-04-23/T-04-24/T-04-26 | Signals are adjudicated by wait status, and no release runs on the defect path | bounded process/IO + mutation kill | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestAbortSignalAdjudicatedByWaitStatus TestNonzeroExitIsDistinctFromSignal TestDefectExpectationRejectsReturnedDocument TestNoReleaseAfterDefect TestVerifyPhase4DefectControls` | ✅ Wave 0 | ✅ green |
| 04-05-01 | 05 | 5 | FFI-01/RES-01 | T-04-27/T-04-28/T-04-29 | One process-root pad names every leaked acquisition from static storage and terminates as a defect, running no release | vertical probe + cost constraint | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestExactlyOneLandingPadIsInstalled TestLedgerIsStaticStorage TestNonlocalExitEmitsLeakPerLiveAcquisition TestPadRunsNoRelease TestNonlocalExitProbeInterpreterNative` | ✅ Wave 0 | ✅ green |
| 04-05-02 | 05 | 5 | FFI-01 | T-04-30/T-04-31 | Any unaudited undefined symbol fails the gate, and a host without the listing tool reports operational rather than passing | allowlist control + honest degradation | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestUndefinedSymbolAllowlistRejectsNewSymbol TestUndefinedSymbolNormalizationHandlesBothFormats TestMissingSymbolToolReportsOperational TestUnwindControlDoesNotInspectSections TestVerifyPhase4UnwindControl` | ✅ Wave 0 | ✅ green |
| 04-05-03 | 05 | 5 | FFI-01 | T-04-32/T-04-33 | Detection is falsified by two independent mutations, and every blind spot is named rather than implied covered | mutation kill + reachability record | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestNonlocalExitDetectionIsMutationKilled TestLeakCountMatchesLiveAcquisitions TestNonlocalExitReachabilityIsRecorded TestBlindSpotsAreNamedNotClaimed TestVerifyPhase4NonlocalExitControl` | ✅ Wave 0 | ✅ green |
| 04-06-01 | 06 | 6 | SEM-03 | T-04-34/T-04-35 | Both independent analyses recognise every terminator, read the set from one registry, and are mutation-killed per package | walk-completeness control | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOriginWalksEveryTerminator TestPathOracleClosesOnEveryTerminator TestTerminatorSetReadFromRegistry TestTerminatorWalkMutationKilled TestFailureOnlyPathIsChecked TestVerifyPhase4TerminatorControl` | ✅ Wave 0 | ✅ green |
| 04-06-02 | 06 | 6 | FFI-01 | T-04-36/T-04-37 | A foreign return that borrows or retains must declare its origin, and published-origin validation runs wherever a program is admitted | differential + CLI wiring | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestForeignBorrowDerivedReturnRecognised TestForeignOriginOmittedRejected TestPublishedOriginValidatedOnCheckAndRun TestOriginValidatorImportsStayIndependent TestVerifyPhase4ForeignOriginControl` | ✅ Wave 0 | ✅ green |
| 04-06-03 | 06 | 6 | SEM-03 | T-04-38/T-04-39/T-04-40 | The checker reports what it costs, no verdict moved, and everything deferred is dated | counted-work series + verdict pin | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/check TestCheckerVerdictsUnchanged TestLastUseDiscoveryWorkIsCounted TestLastUseDiscoveryWorkSeries TestOwnershipWorkSeries` | ✅ Wave 0 | ✅ green |
| 04-07-01 | 07 | 7 | SEM-03/RES-01/FFI-01 | T-04-41/T-04-42/T-04-44 | One gate requires every control by exact identifier and proves three prior phases by running their corpora with this binary | phase gate + contract test | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestVerifyPhase4ControlsAndWork TestVerifyPhase4CLI TestPhase4VerifierScriptContract TestPhase4RequiredControlsMatchScript TestExpectedEscapesAreVisibleNotSolved TestVerifyPhase3ControlsAndWork` plus `sh scripts/verify-phase4.sh` | ✅ Wave 0 | ✅ green |
| 04-07-02 | 07 | 7 | SEM-03/RES-01/FFI-01 | T-04-43/T-04-46 | All three engines agree on all five Phase 4 path shapes, falsified by the two release mutations | three-engine differential | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestPhase4CorpusThreeEngineAgreement TestPhase4DifferentialNamesFirstDisagreement TestForeignDigestMismatchRefused TestReleaseOmissionMutationIsMismatch TestReleaseTranspositionMutationIsMismatch` | ✅ Wave 0 | ✅ green |
| 04-07-03 | 07 | 7 | SEM-03/RES-01/FFI-01 | T-04-45/T-04-47 | Every behavior is demonstrated on the shipped binary on programs the gate has never seen, and no residual is claimed as covered | shipped-binary exercise + register closure | `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestShippedBinaryExercisesEveryPhase4Behavior TestNoCoverageClaimedForNamedResiduals TestPhase4ReachabilityRecordIsComplete` | ✅ Wave 0 | ✅ green |

*Status: ✅ green · ✅ green · ❌ red · ⚠️ flaky*

Sampling continuity: the only task without an `<automated>` command is `04-01-02`, a `checkpoint:decision` that gates a human choice and produces no artifact. It is preceded and followed immediately by tasks carrying full automated verification, so no three consecutive tasks lack automated feedback.

## Required Negative Controls

The Phase 4 verifier must fail closed unless it observes all of these exact identifiers. Each must be observed by running its fixture or mutation, must carry nonzero recomputed work in its lane, and must be asserted both at the session layer and through the shipped CLI.

```text
control:kind.exhaustive_dispatch
control:foreign.call_target_not_foreign
control:foreign.unwind_policy_undeclared
control:resource.release_omitted
control:resource.release_order_transposed
control:foreign.layout_mismatch
control:foreign.no_unproven_attributes
control:defect.no_release_on_defect
control:foreign.unwind_forbidden
control:foreign.nonlocal_exit_undetected
control:terminator.walk_incomplete
control:origin.foreign_origin_omitted
```

The Phase 4 gate must additionally re-observe every Phase 3, Phase 2, and Phase 1 control by running those corpora with the same once-built binary. Non-regression is proven by running the older corpora, never by invoking an older gate script.

The expected escapes must appear under expected escapes and must never appear as a detected control:

```text
escape:coordinated-foreign-boundary-lie
escape:nonlocal-exit-below-the-pad
escape:foreign-process-exit
```

The two carried Phase 3 escapes remain declared and unchanged.

## Wave 0 Requirements

Every item below must exist and be red-then-green before the task that depends on it is considered done.

- [x] Pre-Phase-4 byte-identity pins for the Phase 1, 2, and 3 core bytes and evidence manifest identifiers, written and green **before** any new operation kind lands (D-04-23).
- [x] `core.AllOperationKinds()` and `core.TerminatorKinds()` registries — neither exists today; both back the exhaustive-dispatch and terminator-walk controls.
- [x] The six-site dispatch table, green on the existing five kinds before the tracer widens them.
- [x] `native.validateExecution`'s expected-terminal-outcome axis plus its old-grounds regression test — this blocks every other native task until it lands.
- [x] `native/lang_foreign_resource.c` and `native/lang_foreign_resource_private.h`, hand-written and byte-frozen from their first commit.
- [x] `native/lang_foreign_nonlocal.c`, the second frozen unit that performs the nonlocal exit so the pad has a reachable witness.
- [x] `testdata/phase4/` corpus: `foreign_acquire_one.lang`, `foreign_call_target_not_foreign.lang`, `foreign_unwind_undeclared.lang`, `fallible_call_unconsumed.lang`, `acquire_three_success.lang`, `acquire_three_fail_second.lang`, `acquire_three_fail_third.lang`, `discard_because.lang`, `defect_terminal.lang`, `nonlocal_exit_probe.lang`, `foreign_origin_omitted.lang`, and the frozen `foreign_layout_mismatch.golden.c`.
- [x] Each fixture is led by a comment stating the law under test, is canonical under the formatter fixed point, and is under the source byte limit.
- [x] Checker verdict-pinning test over the whole existing corpus, written and green **before** the work-counting change in 04-06.
- [x] `scripts/verify-phase4.sh` created as a peer of, never an extension of, the Phase 3 script, with its contract test.
- [x] Existing `scripts/assert-go-tests.sh --self-test` run once at the top of the Phase 4 gate before any targeted command is trusted.

## Mutation-Kill Register

Per D-04-21, a differential is not evidence until reverting the production hunk makes it fail. Each row must be demonstrated in a throwaway detached worktree and its failing output pasted verbatim into the owning plan's summary.

| Oracle / differential | Artifact attacked | Production hunk to revert | Expected failure | Owning plan | Status |
|---|---|---|---|---|---|
| Release-order transposition | the checker's materialized order in the core artifact | the reverse-order release materialization | the three-engine differential disagrees, naming the two transposed operations | 04-02 | ✅ green |
| Release omission | the emitter's own generated C | the emitted release line bearing the release marker | the differential reports a live resource at termination | 04-02 | ✅ green |
| Two-acquisition transposition (negative result) | the same core artifact, two-acquisition shape | seeded, not a revert | the differential stays **green**, proving a two-stage fixture cannot falsify order | 04-02 | ✅ green |
| Validator release-order rederivation | the core artifact | the backward-from-failure-edge rederivation | a moved, dropped, duplicated, or invented release is accepted | 04-02 | ✅ green |
| Foreign layout conformance | the frozen foreign translation unit | seeded field transposition, not a revert | the conformance unit is refused at compile time under the existing warning-as-error flags | 04-03 | ✅ green |
| Zero optimizer attributes | the emitter's own output | seeded attribute injection, not a revert | the attribute scan finds a banned token and the control turns red | 04-03 | ✅ green |
| No release on defect | the emitter's defect path | seeded release emission on the defect path | a release event appears after the defect terminal record | 04-04 | ✅ green |
| Nonlocal-exit detection, pad removal | the emitter | the landing-pad installation | the probe produces no nonlocal-exit event and no leak events | 04-05 | ✅ green |
| Nonlocal-exit detection, ledger drop | the emitter | one ledger population site | the leak count understates the live set, reported as a leak-count disagreement | 04-05 | ✅ green |
| Terminator walk, origin analysis | the origin walker | the typed-failure member of the recognised terminator set | the control names the missing member and one fixture's origin fact goes unchecked | 04-06 | ✅ green |
| Terminator walk, path oracle | the path oracle | the typed-failure member of the recognised terminator set | the control names the missing member and one fixture's path fact goes unchecked | 04-06 | ✅ green |
| Counted-work honesty | the transitive last-use scan | the new per-operation work increments | the chain series stops growing and the assertion no longer distinguishes | 04-06 | ✅ green |
| Validator merge-terminal rederivation | the core artifact's merge terminal block | the per-incoming-edge rederivation in checkReleaseOrder (16fb0c9) | a hand-corrupted merge of two acquisition chains with divergent completed-acquisition sets is silently accepted instead of refused | 04-08 | ✅ green |
| Terminal block reachability | the core artifact's block/edge graph | the core.terminal_block_unreachable refusal in blocksAndEdges | a non-entry terminal block with zero incoming edges is silently accepted (falls back to a different, non-structural refusal) instead of refused by the peer check | 04-08 | ✅ green |
| Zero optimizer attributes, conformance layer | the generated conformance TU's own added text (the region after the embedded header) | the tracerConformance/releaseConformance arguments added to session.go's ScanForBannedAttributes call by commit cb701b5 | a banned token injected into the conformance unit's own added text is silently accepted instead of detected, with the untouched header and compiled-program C staying clean | 04-09 | ✅ green |
| Attribute-scan artifact count | the production lane's ScanForBannedAttributes argument list | any single scan argument (header or conformance), plus the lane's declared work value | the lane's RecomputedWork drops below 8 (two fixture programs x four artifacts each) while still reporting "pass", instead of the pinned-count test turning red | 04-09 | ✅ green |
| Rederivation cycle guard | the core artifact's declared ok-edge chain (a hand-corrupted two-block cycle) | the visited-set guard and core.release_order_cyclic refusal in checkReleaseOrder's rederive closure | the backward walk never terminates, so validation hangs instead of refusing — observed as the falsifier's bounded-deadline fatal (or the harness test timeout), and, under the weaker truncation mutation, as a missing core.release_order_cyclic code | 04-10 | ✅ green |
| Validator interior-merge rederivation | the core artifact's declared ok-edge graph (a second ok edge into a non-terminal block from a source with a different completed-acquisition history) | the plural okEdgeInto map and the per-candidate agreement comparison with its core.release_order_merge_mismatch refusal in checkReleaseOrder's rederive closure | the discarded history is never rederived, so the corrupted program is silently accepted in one edge ordering and refused with the wrong code in the other; under the weaker comparison-only mutation the merge is walked but disagreement is accepted; under the shared-visited-set mutation a legitimate reconvergence is misrefused as a cycle | 04-11 | ✅ green |
| Foreign symbol identifier audit at the C boundary | the core artifact's ForeignContract.Symbol field, spliced raw into the extern declaration, the call-expression callee, and the generated header's symbol comment | the foreign.symbol_not_identifier audit in corevalidate's ForeignContract block and cgen's own validForeignSymbol guards in emitLinearForeign and singleForeignFunction | a Symbol carrying C syntax metacharacters validates and is emitted verbatim into generated C, injecting arbitrary top-level source into a translation unit the compiler then compiles and executes; reverting only the cgen guards still injects through the three exported EmitForeign entry points that never call Validate; weakening either predicate to a non-empty check restores the original gap | 04-12 | ✅ green |
| Foreign contract field audit at the C boundary, source-reachable path | an ordinary Lang author's own `foreign C { }` policy value, and every remaining `core.ForeignContract` string field spliced into the generated header's comment block and the conformance unit's `_Static_assert` operands | the `check.foreign_policy_value_unsafe` admission refusal in `collectForeignSymbols`, the `foreign.policy_value_not_identifier` and `foreign.contract_field_not_c_safe` audits in corevalidate's `OpForeignCall` block, and cgen's own `unsafeForeignContractField` guard in `singleForeignFunction` | an honest `.lang` source declaring `allocator: "*/ int injected(void){return 1;} /*"` passes `session.Check` with zero diagnostics and `EmitForeignHeader` emits a live top-level C function definition outside any comment into a unit `native.Runner.CompileConformanceUnit` then compiles; reverting only the corevalidate audits still admits a corrupted `core.Program`; reverting only the cgen guard still injects through the three exported `EmitForeign*` entry points that never call `Validate`; weakening any predicate to a non-empty or always-true check restores the original gap | 04-13 | ✅ green |

The two release mutations and the layout mutation attack **three different artifacts** — the checker's materialized order, the emitter's own output, and the frozen fixture. An author keeping any two aligned cannot satisfy all three.

## Generator and Probe Reachability Register

Per D-04-21, coverage is not the question; reachability is. Each generator or probe must carry a comment stating what it reaches and what it does not, and this table must be completed from those comments during execution.

| Generator / probe | Must reach | Known not to reach | Owning plan | Status |
|---|---|---|---|---|
| Three-acquisition fixture family | success, second-stage failure, and third-stage failure paths; release sets of size 0, 1, and 2 on failure | four or more acquisitions; a partially initialized field within a single acquisition; any loop-carried acquisition, since no loop construct exists | 04-02 | ✅ green |
| Six-site dispatch table | every declared operation kind at every one of the six sites, including both switches in each two-switch package | a seventh consumer added without registering it in the table — the table must itself enumerate sites, not be enumerated by hand | 04-01 | ✅ green |
| Conformance-unit assertions | every declared record field's size, alignment, and offset, plus record-level size and alignment | unions, bit fields, vectors, variadics, packed and aligned records; the other object format and architecture | 04-03 | ✅ green |
| Defect witness | a defect terminator in a match arm, requiring no call surface | a defect raised from inside a foreign call; any contained or caught defect, which does not exist this phase | 04-04 | ✅ green |
| Nonlocal-exit probe | a foreign nonlocal transfer to the process-root pad with at least one acquisition live | a landing point established below the pad through a foreign-invoked callback; a foreign process-exit call; a signal raised inside foreign code | 04-05 | ✅ green |
| Terminator-walk fixtures | a path whose only exit is a typed failure, and a path whose only exit is a defect | a terminator kind not yet declared — which is exactly what the registry-equality control exists to catch | 04-06 | ✅ green |
| Undefined-symbol allowlist | the linked binary's full undefined set on this host, underscore-normalized | the other object format's naming convention, which is normalized for but not exercised on this host | 04-05 | ✅ green |

## Shipped-Binary Register

Per D-04-21, the gate only ever sees what ships with it. Each row must be exercised against a freshly built `./cmd/lang` on a hand-written program that is **not** in any corpus.

| Behaviour | Commands | Owning plan | Status |
|---|---|---|---|
| Fallible foreign call through `try` | `format --check`, `check`, `run --engine=interpreter`, `run --engine=native` | 04-01 | ✅ green |
| Policy-less foreign declaration refusal | `check` with JSON projection, exit code and diagnostic code asserted | 04-01 | ✅ green |
| Lang-targeted call refusal | `check` with JSON projection, exit code and diagnostic code asserted | 04-01 | ✅ green |
| Three-stage acquisition with reverse-order release | `check`, both engines, event order asserted | 04-02 | ✅ green |
| `discard ... because` consumer | `format --check`, `check`, both engines | 04-02 | ✅ green |
| Defect terminator | `check`, `run --engine=interpreter`, `run --engine=native` with wait-status assertion | 04-04 (native run --engine=native multi-arm bug fixed 04-07) | ✅ green |
| Nonlocal-exit probe | `check`, both engines, leak events and terminal record asserted | 04-05 (native run --engine=native single-defect-input bug fixed 04-07) | ✅ green |
| Foreign-origin refusal | `check` with JSON projection, exit code and diagnostic code asserted | 04-06 | ✅ green |

## Property and Cost Lanes

- Ordinary `go test` exercises the existing exhaustive ownership and branch sequence enumerations unchanged, plus the new six-site dispatch table and the terminator-set equality assertions.
- Scalability is asserted by counted work — the validator's release-order rederivation at three artifact sizes, and the checker's transitive scan at three chain lengths — never by wall clock.
- Investigation command: the existing fuzz targets, time-bounded, never in the default gate.
- The conformance compile, the program compile, the symbol listing, and the program run are four separate bounded, timed invocations; none shares a writer or a timeout context with another.

## Manual-Only Verifications

None. Every Phase 4 semantic behaviour has an automated verification. Human review remains valuable for the frozen translation unit's C readability, the generated obligation comment block's usefulness to a reviewer, and the honesty of the unchecked-obligations wording, but none of the three is accepted as a substitute for an automated check.

The one non-automated gate in the phase is `04-01-02`, a `checkpoint:decision`. It gates a human choice about a one-way door and produces no artifact to verify; it is not a manual verification of built behaviour.

## Threat Model

Phase 2's and Phase 3's registers remain in force unchanged and are re-proven by running those corpora with the Phase 4 binary. The rows below are the Phase 4 additions; each is carried in full in its owning plan's `<threat_model>`.

| ID | Threat | Failure mode | Required mitigation and falsifier |
|----|--------|--------------|------------------------------------|
| T-04-01 | A new operation kind unhandled at one of six dispatch sites | the operation vanishes from the trace the differential compares | one registry, one table-driven control over all six sites, both switches per two-switch package |
| T-04-02 | A foreign declaration admitted without stating its policies | an unaudited boundary passes as audited | mandatory policy with no default, refused independently by two layers |
| T-04-03 | A call target resolving to a Lang function | the deliberately absent call surface arrives by accident | hard refusal at both layers, independently derived |
| T-04-04 | Quarantine breached by header ingestion | the compiler's contract becomes a restatement of the foreign side | the private header is reachable only from the generated conformance unit, asserted by test |
| T-04-05 | Foreign declaration surface blowup | unbounded admission cost | declared, commented caps rejecting fail-closed |
| T-04-06 | The native harness silently rejecting or silently accepting a new terminal outcome | every Phase 4 native task is blocked or vacuously green | explicit expected-outcome axis with an old-grounds regression test |
| T-04-07 | A coordinated edit across the frozen unit, the declaration, and the conformance expectations | self-consistent false evidence | **accepted residual** — declared as a named expected escape, asserted visible, never claimed solved |
| T-04-08 | Release order derived by the emitter rather than consumed | cleanup order is decided twice and only agrees by luck | one materialized order; three-acquisition fixture plus transposition mutation |
| T-04-09 | A release silently dropped from the emitted C | a resource leaks with no evidence | omission mutation against the emitter's output, exact-one-marker fail-closed |
| T-04-10 | Releasing an incomplete acquisition, or releasing twice | undefined behaviour presented as cleanup | failure blocks release only completed undischarged acquisitions; validator rederives and refuses |
| T-04-11 | Two derivations agreeing because they share a law | a differential agrees on the wrong answer | different mechanism classes; import independence asserted; revert-and-fail recorded |
| T-04-12 | Unbounded validator cost on the new facts | the gate becomes the bottleneck it was built to bound | per-inspection counting, limit re-derived not bumped, linear series |
| T-04-13 | Nondeterministic release sequence from map iteration | the same input produces two answers | explicit accumulation ordinal; explicit sort where a map is unavoidable |
| T-04-14 | Foreign record layout disagreement | silent runtime corruption, as spike 005 iteration 2 demonstrated | paired size, alignment, and offset assertions compiled under warning-as-error; transposition is a required control |
| T-04-15 | An unproven optimizer attribute | wrong code at higher optimization, as spike 005 iteration 4 demonstrated | zero attributes emitted, scan over all emitted C and the manifest field, mutation-killed by injection |
| T-04-16 | Obligation comments drifting from the authoritative contract | a reviewer trusts a stale claim | comments generated from the serialized contract; a contract mutation must move its comment |
| T-04-18 | Evidence identity moving for pre-existing programs | every prior manifest becomes unverifiable | the identity function forks rather than grows; pre-Phase-4 identifiers pinned |
| T-04-20 | A panic path reaching typed failure | panic erased as an ordinary error | no error-value constructor; typed failure producible only from an err edge |
| T-04-21 | Evidence erased by a dying process | the criterion is unfalsifiable on exactly the paths it is about | additive streaming emitter; terminal record last; absence is a hard failure |
| T-04-22 | An absent terminal record tolerated as truncation | a failing run looks like a short output | distinct codes for absence and for cap truncation, asserted against each other |
| T-04-23 | Cleanup attempted from indeterminate state | a leak converted into a use-after-free | enforced no-release-on-defect control, mutation-killed |
| T-04-24 | A shell's numeric encoding standing in for a signal | a wrong adjudication that happens to pass on one host | wait-status adjudication with no numeric literal in the signal case |
| T-04-25 | Cancellation quietly retrofitted as an error variant | the cancellation seam closes without review | reserved and unconstructible, asserted over reachable constructors |
| T-04-27 | A foreign nonlocal exit bypassing cleanup unnoticed | resources leak with no record | one process-root pad emitting a nonlocal-exit event plus a leak per live acquisition, mutation-killed twice |
| T-04-29 | A landing pad per call or per borrow | the cost constraint the wiki sets is violated | exactly one installation, asserted across one, two, and three acquisitions |
| T-04-30 | An unaudited symbol linking into the binary | an unwind or exit path arrives unreviewed | positive allowlist with per-entry rationale |
| T-04-31 | A control passing on a host that cannot answer its question | absence of a tool read as absence of a problem | operational tool-missing status distinct from pass and from fail |
| T-04-32 | A green control whose reachable input space omits the hard case | the failure mode this project has already paid for three times | reachability register completed from in-code comments, each naming an unreached shape |
| T-04-33 | A landing point below the pad, or a foreign process-exit call | a nonlocal exit the pad cannot observe | **accepted residual** — named expected escapes, recorded in the unchecked-obligations list, never claimed covered |
| T-04-34 | An analysis walking one terminator rather than every terminator | the failure path is invisible while the optimization differential still passes | registry-driven membership test in both packages; set-equality control mutation-killed per package |
| T-04-36 | An undeclared borrow-derived foreign return | an alias fact silently absent at a second surface | refusal derived from the core artifact alone; published-origin validation reaching every admission path |
| T-04-38 | An understated cost metric | a bound appears satisfied because the expensive step counts nothing | per-operation counting, growth series at three lengths, before and after values recorded |
| T-04-40 | Compounding defects from restructuring a file while it gains four operation kinds | the shape two prior phases already hit | the derivation retirement is deferred and recorded; only the counting change lands, behind a verdict pin |
| T-04-41 | A gate appearing to prove more than it runs | reviewers trust a claim the script does not execute | contract test over the script's own text; set equality between the session list and the script |
| T-04-42 | Non-regression proven by invoking an older gate | an older script's own staleness is inherited | one binary runs all four corpora; the contract test forbids invoking a prior gate |
| T-04-45 | A behavior proven only in the harness | the shipped binary diverges from what the gate tested | out-of-corpus exercise per behavior with subcommands, exit codes, and diagnostic codes recorded |
| T-04-SC | Supply chain, package installs | an unaudited dependency enters the toolchain | **not applicable by construction** — this phase installs no external packages; any future third-party dependency must run the Package Legitimacy Gate at that time |

## Debt Closure Register

Carried from `03-DEBT.md`. Per the standing rule, no debt-cleanup plan exists; each item lands in the plan that already touches its files.

| Debt ID | Landing plan | Closure evidence required | Status |
|---|---|---|---|
| D-03-01 (metric-honesty half) | 04-06 | per-operation counting inside the transitive scan; growth series at three chain lengths; before and after values recorded; no verdict moved | ✅ green |
| D-03-01 (retirement half) | deferred to the next phase | re-recorded as dated open debt in `04-DEBT.md` with the compounding-wave rationale | ✅ green |
| D-03-02 | closed by construction this phase | no Lang-to-Lang call surface exists, so the defect class has no consumer; the lift condition is recorded as callable being a subset of publishable | ✅ green |
| WR-01 | 04-06 | published-origin validation invoked from the check and run command paths, with a fixture refused there that previously passed | ✅ green |

## Validation Sign-Off

- [x] All tasks have an immediate `<automated>` verification or an explicit Wave 0 dependency; the single exception is the `checkpoint:decision`, which gates a human choice and produces no artifact.
- [x] No three consecutive tasks lack automated feedback.
- [x] All Wave 0 references exist and are green.
- [x] Every row of the Mutation-Kill Register is demonstrated, with failing output recorded verbatim in the owning plan's summary — including the deliberately-green two-acquisition row.
- [x] Every row of the Generator and Probe Reachability Register is completed from an in-code comment, each naming at least one unreached shape.
- [x] Every row of the Shipped-Binary Register is exercised on a program that is in no corpus, with subcommands, exit codes, and diagnostic codes recorded.
- [x] Every row of the Debt Closure Register is closed or explicitly re-carried with a dated reason.
- [x] No watch-mode or unbounded-fuzz flag enters a default command.
- [x] `scripts/verify-phase1.sh`, `scripts/verify-phase2.sh`, and `scripts/verify-phase3.sh` are byte-identical to their state at the phase start, and all three still exit zero.
- [x] `git diff <phase-start>..HEAD -- testdata/phase1 testdata/phase2 testdata/phase3` is empty; every committed generated-C golden is byte-identical.
- [x] `go.mod` gained no requirement.
- [x] `native/lang_foreign_resource.c`, `native/lang_foreign_resource_private.h`, and `native/lang_foreign_nonlocal.c` are byte-identical to their first committed state at the end of the phase.
- [x] Measured warm feedback latency for the Phase 4 fixtures is reported as an observation, not ratified as a release SLO.
- [x] `nyquist_compliant: true` and `wave_0_complete: true` are set only after execution supplies the evidence above.

**Approval:** granted 2026-09-05 by plan 04-07 (Task 3), the phase's closing plan. Every checkbox above was ticked only after `go test ./...`, `go test -race ./...`, `go vet ./...`, and `sh scripts/verify-phase4.sh` were run and observed to exit zero, and the out-of-corpus shipped-binary and reachability-record tests (TestShippedBinaryExercisesEveryPhase4Behavior, TestPhase4ReachabilityRecordIsComplete, TestNoCoverageClaimedForNamedResiduals) were run and observed to pass.
</content>

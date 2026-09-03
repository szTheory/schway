---
phase: "02"
slug: "owned-values-and-abilities"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-03"
---

# Phase 2 — Validation Strategy

> Feedback contract for proving affine ownership and independent abilities across the real compiler spine.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 `testing`, `testing/quick`, native fuzz seeds, black-box `os/exec` |
| **Config file** | `go.mod` |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` |
| **Full suite command** | `sh scripts/verify-phase2.sh` |
| **Feedback budget** | quick <15 seconds; full <60 seconds on the current macOS arm64 host (observe, do not ratify as release SLO) |

## Sampling Rate

- **After every task commit:** run the task's targeted command, then `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...`.
- **After every plan wave:** run `env GOCACHE=/tmp/ai-lang-phase2-cache go test -race ./...` and `env GOCACHE=/tmp/ai-lang-phase2-cache go vet ./...`.
- **Before phase verification:** `sh scripts/verify-phase2.sh` must be green and must report every required negative-control ID.
- **Investigation only:** active fuzzing is time-bounded and never enters the deterministic default gate.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | OWN-01/02 | T-02-02/T-02-04 | Test selection fails on zero matches before explicit Buffer transfer and Byte copy reach interpreter facts | harness self-test + vertical tracer | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestOwnedTransferInterpreter TestImplicitByteCopy` | ✅ | ✅ green |
| 02-01-02 | 01 | 1 | OWN-01/02 | T-02-03 | Match/linear union, IDs, and feature schemas preserve Phase 1 | core/session contract | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestClosedBodyUnion TestLinearIdentityStability TestFeatureSpecificCoreExecutionSchemas` | ✅ | ✅ green |
| 02-02-01 | 02 | 2 | OWN-02 | T-02-02/T-02-03 | Source Box/Pair shapes produce independently derived typed-core facts | source-to-core contract | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestSourceBoxPairAbilityFacts TestSourceCannotGrantAbilityRoots` | ✅ | ✅ green |
| 02-02-02 | 02 | 2 | OWN-02 | T-02-01/T-02-02 | Test-private masks exhaust the unexported production combiner and the sealed TypeRef path demonstrably uses it | exhaustive/non-tautology unit | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/ability TestBoxAllAbilityMasks TestPairAllAbilityMasks TestTypeRefDerivationUsesStructuralCombiner TestNegativeWitnessPath TestArbitraryMasksRemainTestPrivate` | ✅ | ✅ green |
| 02-02-03 | 02 | 2 | OWN-02 | T-02-01/T-02-03 | Generic syntax is lossless, canonical, recoverable, and bounded | syntax property | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/syntax TestOwnershipRoundTrip TestOwnershipRecovery TestTypeApplicationLimits FuzzParseFormat` | ✅ | ✅ green |
| 02-03-01 | 03 | 3 | OWN-01 | T-02-04/T-02-05 | Three ownership failures carry causal, version-compatible diagnostics | compile-reject contract | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestTransferRequiresTake TestUseAfterMoveDiagnostic TestMoveWhileBorrowedDiagnostic TestDiagnosticSchemaCompatibility TestHumanJSONMixedDiagnosticVersionParity TestPhase1DiagnosticGoldenUnchanged` | ✅ | ✅ green |
| 02-03-02 | 03 | 3 | OWN-01 | T-02-01/T-02-05 | Static support facts match an independent model across shadowing and multiple owners/loans with linear counted work | exhaustive/model property | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/check TestOwnershipSequenceExhaustive TestOwnershipOracleTracksLoansPerOwner TestOwnershipWorkSeries FuzzOwnershipLinear` | ✅ | ✅ green |
| 02-04-01 | 04 | 4 | OWN-01/02 | T-02-02/T-02-03/T-02-05 | Forged abilities and illegal operations cannot reach execution | core mutation | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnershipMutationMatrix TestUnvalidatedCoreCannotExecute TestArbitraryMaskCannotEnterCoreValidation` | ✅ | ✅ green |
| 02-04-02 | 04 | 4 | OWN-01/02 | T-02-01/T-02-08 | Validator work is bounded and its source-translation limitation is explicit | scale/claim contract | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestCoreValidationWorkSeries TestOwnedClaimReorderRejected TestCoordinatedSourceCoreEscapeIsNamed` | ✅ | ✅ green |
| 02-05-01 | 05 | 5 | OWN-01 | T-02-06/T-02-07 | Compile/run stdout and stderr are independently bounded; O0/O3 agree | native differential | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestNativeStreamsIndependentlyBounded TestNativeNeverUsesCombinedOutput TestOwnedTransferInterpreterNative TestNativeToolFailureIsOperational` | ✅ | ✅ green |
| 02-05-02 | 05 | 5 | OWN-01 | T-02-06/T-02-07 | Strict decoder and full execution comparison reject malformed or reordered facts | decoder/mutation | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestExecutionDecoderRejectsMalformedOutput TestOwnedEventReorderIsMismatch TestOwnedExecutionFieldMutationMatrix` | ✅ | ✅ green |
| 02-06-01 | 06 | 6 | OWN-01/02 | T-02-03/T-02-08 | Evidence binds validated core and ordered executions without overclaiming | evidence mutation | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnedEvidenceBindings TestOwnedEvidenceMutationMatrix TestOwnershipProjectionIdentityParity TestPhase1EvidenceGoldenUnchanged TestCoordinatedSourceCoreEscapeIsNamed` | ✅ | ✅ green |
| 02-06-02 | 06 | 6 | OWN-01/02 | T-02-01..T-02-08 | One bounded gate proves test discovery, Phase 1 regression, controls, work, and cost | phase gate | `sh scripts/verify-phase2.sh` | ✅ | ✅ green |
| 02-07-01 | 07 | 7 | OWN-01 | T-02-01/T-02-03/T-02-06/T-02-07 | Executed C operations and returned runtime state cause the bounded native execution document | native causality integration | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestLinearCSerializesRuntimeState TestOwnedTransferInterpreterNative TestPhase1EvidenceGoldenUnchanged` | ✅ | ✅ green |
| 02-07-02 | 07 | 7 | OWN-01 | T-02-07/T-02-08 | Exact-one emitted-C mutation runs through real O0/O3 and must produce semantic mismatch exit 4 plus a nonzero-work control | backend mutation integration | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork TestVerifyPhase2CLI TestPhase1EvidenceGoldenUnchanged` | ✅ | ✅ green |

## Required Negative Controls

The phase verifier fails closed unless it observes all of these exact IDs:

```text
control:ownership.use_after_move
control:ownership.move_while_borrowed
control:ownership.transfer_requires_take
control:ability.forged_copy
control:core.duplicate_operation_id
control:interpreter-o0-o3-owned
control:evidence.core_mismatch
control:backend.runtime_causality
```

It also records `escape:coordinated-source-core-lie` as a known trust-boundary limitation, never as a detected mutation.

## Wave 0 Requirements

- [x] Add canonical fixtures under `testdata/phase2/`: owned transfer, implicit copy, use after move, move while borrowed, and implicit noncopyable read.
- [x] Add a canonical source fixture covering nested `Box`/`Pair` ability derivation.
- [x] Add `scripts/assert-go-tests.sh` first; its `--self-test` path must route a guaranteed nonexistent name through the real selector, require nonzero exit, then prove a known existing test before any planned targeted command is trusted.
- [x] Add ownership syntax fixed-point, lossless round-trip, bounded-recovery, and fuzz seeds.
- [x] Add exhaustive independent ability-oracle tests before implementing ability derivation.
- [x] Add core-validator mutation skeletons before connecting execution.
- [x] Add strict native execution-decoder malformed/unknown/trailing/oversized controls.
- [x] Add `scripts/verify-phase2.sh` as one non-duplicating deterministic gate.

## Property and Fuzz Lanes

- Ordinary `go test` exhausts short ownership-operation sequences, all 32 leaf ability masks, all 1,024 `Pair` mask pairs, bounded generated type trees, and persisted fuzz seeds.
- Investigation command: `go test ./internal/compiler/check -fuzz=FuzzOwnershipLinear -fuzztime=30s`.
- Fuzz targets allocate all state per invocation, perform no native compilation, and cap type depth and operation count before evaluation.
- Checker scalability is asserted by counted operations at 10, 100, 1,000, and 10,000 operations—not wall-clock thresholds.

## Manual-Only Verifications

All Phase 2 semantic behaviors have automated verification. Human review remains useful for C readability and diagnostic prose, but neither is accepted as a substitute for an automated semantic check.

## Threat Model

| ID | Threat | Failure mode | Required mitigation and falsifier |
|----|--------|--------------|------------------------------------|
| T-02-01 | Adversarial nested types or long bodies | compiler work or recursion becomes unbounded | explicit caps, memoization, counted linear-work tests |
| T-02-02 | Forged positive ability | unauthorized copy/share/send/escape reaches execution | sealed primitive rules plus independent core recomputation and mutation |
| T-02-03 | Ambiguous body representation or duplicate IDs | consumers infer different meaning | closed match/linear variant validation and uniqueness controls |
| T-02-04 | Implicit noncopyable transfer | ordinary-looking read silently consumes authority | explicit `take` requirement and compile-reject control |
| T-02-05 | Corrupted move/loan core | use-after-move or move-during-live-loan executes | independent validator and interpreter state checks |
| T-02-06 | Compiler or native child hangs or floods/malforms output | denial of service or accepted false facts | timeout, four independent compile/run stdout/stderr caps, stdout-only strict decoder, no `CombinedOutput` |
| T-02-07 | Backend re-infers source intent | optimizer-dependent semantic drift | C generation consumes explicit operations; exact ordered-event differential |
| T-02-08 | Digest is mistaken for translation proof | internally consistent frontend lie is overclaimed | evidence labels content binding only and records expected escape |

## Validation Sign-Off

- [x] All tasks have an immediate `<automated>` verification or an explicit Wave 0 dependency.
- [x] No three consecutive tasks lack automated feedback.
- [x] All Wave 0 references exist and are green.
- [x] No watch-mode or unbounded-fuzz flags enter default commands.
- [x] Measured warm command feedback latency is reported and within the provisional budget; the whole clean-cache gate remains an observational lane rather than a ratified release SLO.
- [x] `nyquist_compliant: true` and `wave_0_complete: true` are set only after execution supplies evidence.

**Approval:** validated against reviewed `HEAD` `8a4b9f5` from fresh task selectors, negative harness probes, and the complete Phase 2 gate on 2026-09-03

## Measured Validation Observations

Fresh `sh scripts/verify-phase2.sh` execution exited 0. Phase 1 remained green with five lanes and 18 recomputed work units. Phase 2 remained green with five lanes, all eight required controls, 51 recomputed work units, and the expected coordinated-source/core escape kept outside detected controls.

| Command lane | Warm samples | p50 | p95 | Min | Max | Output bytes | Work |
|--------------|-------------:|----:|----:|----:|----:|-------------:|-----:|
| format | 20 | 38,292 ns | 45,958 ns | 34,000 ns | 48,292 ns | 248 | 1 |
| check | 20 | 47,916 ns | 61,250 ns | 44,875 ns | 90,416 ns | 302 | 52 |
| interpreter | 20 | 173,500 ns | 193,208 ns | 163,375 ns | 202,125 ns | 902 | 1 |
| native | 20 | 462,216,125 ns | 506,134,083 ns | 447,545,417 ns | 512,507,583 ns | 905 | 3 |
| full verify | 20 | 912,927,375 ns | 1,181,393,417 ns | 670,283,792 ns | 1,417,613,792 ns | 1,710 | 51 |

Peak RSS was reported honestly as unavailable for every lane. All values are operational observations excluded from semantic identity.

## Validation Audit 2026-09-03

| Metric | Count |
|--------|------:|
| Actual tasks mapped | 15 |
| Mapped targeted selectors green | 14/14 |
| Review-supplement selectors green | 1/1 |
| Negative harness probes correct | 3/3 |
| Full phase gates executed | 1 |
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
| Manual-only requirements | 0 |

Audit notes:

- `HEAD` was exactly `8a4b9f5f2fa03d9c0d76dbe0ba0917f1a73db916` (`fix(02): WR-01 strengthen independent ownership oracle`) for this audit.
- A deliberately absent target was rejected as `target not discovered` with exit 1; a malformed target was rejected as `invalid exact Go target` with exit 2; an empty target list was rejected with usage and exit 2.
- All mapped task selectors from 02-01-01 through 02-07-02 passed. The strengthened 02-03-02 selector additionally executes `TestOwnershipOracleTracksLoansPerOwner`, covering independent backward name resolution, shadowing, and multiple owner/loan relationships.
- The backend causality test uses an exact-one emitted-C mutation, real Clang at both `-O0` and `-O3`, and asserts `native.engine_mismatch` with exit 4.
- Bounded syntax recovery, 10/100/1,000/10,000 checker work, 101/1,001/10,001 validator work, four independent native streams, strict one-document decoding, and the 64 KiB generated-output ceiling all have active passing controls.
- The complete gate passed `go test ./...`, `go test -race ./...`, `go vet ./...`, Phase 1 verification, and Phase 2 verification. Phase 1 retained five green lanes and 18 work units; Phase 2 retained five green lanes, all eight exact controls, and 51 work units.
- Phase 1 schema, diagnostic, execution, and evidence regressions are exercised by task selectors and the complete gate. No Phase 2 requirement depends on manual-only verification.

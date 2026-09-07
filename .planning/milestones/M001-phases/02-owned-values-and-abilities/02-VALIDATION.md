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
| 02-05-01 | 05 | 5 | OWN-01 | T-02-06/T-02-07 | Compile/run stdout and stderr are independently bounded (guard is now repo-wide, not native.go-only); O0/O3 agree | native differential | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestNativeStreamsIndependentlyBounded TestSourceNeverSpawnsUnboundedProcesses TestOwnedTransferInterpreterNative TestNativeToolFailureIsOperational` | ✅ | ✅ green |
| 02-05-02 | 05 | 5 | OWN-01 | T-02-06/T-02-07 | Strict decoder and full execution comparison reject malformed or reordered facts | decoder/mutation | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestExecutionDecoderRejectsMalformedOutput TestOwnedEventReorderIsMismatch TestOwnedExecutionFieldMutationMatrix` | ✅ | ✅ green |
| 02-06-01 | 06 | 6 | OWN-01/02 | T-02-03/T-02-08 | Evidence binds validated core and ordered executions without overclaiming | evidence mutation | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnedEvidenceBindings TestOwnedEvidenceMutationMatrix TestOwnershipProjectionIdentityParity TestPhase1EvidenceGoldenUnchanged TestCoordinatedSourceCoreEscapeIsNamed` | ✅ | ✅ green |
| 02-06-02 | 06 | 6 | OWN-01/02 | T-02-01..T-02-08 | One bounded gate proves test discovery, Phase 1 regression, controls, work, and cost | phase gate | `sh scripts/verify-phase2.sh` | ✅ | ✅ green |
| 02-07-01 | 07 | 7 | OWN-01 | T-02-01/T-02-03/T-02-06/T-02-07 | Executed C operations and returned runtime state cause the bounded native execution document | native causality integration | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestLinearCSerializesRuntimeState TestOwnedTransferInterpreterNative TestPhase1EvidenceGoldenUnchanged` | ✅ | ✅ green |
| 02-07-02 | 07 | 7 | OWN-01 | T-02-07/T-02-08 | Exact-one emitted-C mutation runs through real O0/O3 and must produce semantic mismatch exit 4 plus a nonzero-work control | backend mutation integration | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork TestVerifyPhase2CLI TestPhase1EvidenceGoldenUnchanged` | ✅ | ✅ green |

### Post-Gate Supplement (added by the 2026-09-03 re-audit at `cbba405`)

Three behaviors introduced or altered by `91a6206`, `a8c14da`, and `cbba405` had passing tests but no task-mapped selector. Each row below points at a test that already exists in the tree; none were invented.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-07-01-S1 | 07 | 7 | OWN-01 | T-02-01/T-02-07 | Generated C identifiers are globally unique across categories: Unicode locals (`α`/`β`), reserved-looking `__1`, shadowed `value`, case/suffix variant neighbours (`a`/`A`/`A_1`), and the cross-category `Thing`/`thing`/`thing_LANG_THING` case all compile under real Clang at `-O0` and `-O3` | native collision integration | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestNativeIdentifiersRemainCollisionFree TestNativeToggleO0O3` | ✅ | ✅ green |
| 02-06-01-S2 | 06 | 6 | OWN-01/02 | T-02-03/T-02-08 | Frozen Phase 1 and pre-`91a6206` Phase 2 goldens keep byte identity: `testdata/phase1/generated.golden.c`, `testdata/phase1/evidence.golden.json`, and `testdata/phase2/evidence.golden.json` | golden byte identity | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/evidence TestCanonicalEvidence TestPhase1EvidenceGoldenUnchanged TestOwnedEvidenceBindings` | ✅ | ✅ green |
| 02-06-01-S3 | 06 | 6 | OWN-01/02 | T-02-06/T-02-08 | Both evidence tool-identity probes (`clang --version` and `clang -dumpmachine`) are independently bounded on stdout overflow, stderr overflow, and timeout | bounded child-process probe | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/evidence TestDefaultFactsBoundsBothToolProbes` | ✅ | ✅ green |

### Post-Gate Supplement 2 (added by the 2026-09-03 re-audit at `da75e95`)

Nine commits landed after `cbba405`. Two of them fixed CRITICAL blockers that the prior 18/18-green audit did not surface. Seven behaviours introduced or hardened by those commits had passing tests but no task-mapped selector. Every row below points at a test that already exists in the tree; none were invented, and each was independently checked for falsifiability by mutation or revert in a throwaway worktree.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-02-03-S4 | 02 | 2 | OWN-01/02 | T-02-03/T-02-05 | `lang format` output over the **linear** surface reparses and preserves token identity: generated `{Byte, Buffer, Box<Byte>, Box<Box<Byte>>, Pair<Byte, Buffer>}` x `{implicit copy, take, borrow}` x `{0,1,2,3}` bindings round-trip, and two generic-type-plus-binding fuzz seeds are pinned in `FuzzParseFormat`. Both generated properties assert `semanticTokens` identity, so a formatter that fuses or drops an identifier fails even when the result still parses | generated syntax property + fuzz seeds | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/syntax TestGeneratedLinearRoundTrips TestGeneratedRoundTrips FuzzParseFormat` | ✅ | ✅ green |
| 02-06-01-S5 | 06 | 6 | OWN-01/02 | T-02-03/T-02-08 | `evidence.Build` fails closed with `evidence.canonical_unstable` when the canonical bytes do not reparse or are not a formatter fixed point, instead of binding a manifest to a corrupted projection or fabricating a source diagnostic at offsets the user never wrote. Exercised through injected format/parse seams, so it does not depend on a live formatter defect | trust-boundary fail-closed unit | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/evidence TestCanonicalRoundTripFailsClosed TestCanonicalRoundTripAdmitsValidPrograms` | ✅ | ✅ green |
| 02-03-02-S6 | 03 | 3 | OWN-01 | T-02-05 | Loan liveness is **transitive** at both admission layers and at the source surface: `let c = borrow b` (one hop) must reject the same move the direct form rejects. The check-layer oracle was re-derived — it now materializes the derivation relation as an explicit edge set and closes it by order-independent fixed-point iteration, a method production (single forward stream, inherited association, relation never materialized) cannot mirror | differential + independent-model property + fixture control | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnershipSequenceExhaustive TestTransitiveLoanBlocksMove TestLoanLivenessIsTransitive` | ✅ | ✅ green |
| 02-06-02-S7 | 06 | 6 | OWN-01 | T-02-05/T-02-08 | The ninth negative control `control:ownership.move_while_reborrowed` (fixture `testdata/phase2/reborrow_while_moved.lang`) is required by the session fail-closed control set **and** asserted through the shipped CLI, so the two layers cannot disagree about the required nine | gate/CLI control parity | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestVerifyPhase2ControlsAndWork TestVerifyPhase2CLI` | ✅ | ✅ green |
| 02-07-01-S8 | 07 | 7 | OWN-01 | T-02-01/T-02-07 | The generated-C ordinary identifier namespace is closed by two invariants rather than by a fixture list: PREFIX CONFINEMENT (every allocated identifier is `^LANG_[A-Z0-9_]*$` or `^lang_value_[A-Za-z0-9_]*$`, and the collision suffix never leaves the namespace or re-issues a reserved name) and HONEST RESERVATION (the fixed-identifier set is derived from a disjoint-name differential of two generated programs, not hand-listed, and the test fails rather than passing vacuously if the differential separates nothing) | emitted-artifact invariant | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen TestGeneratedIdentifierNamespacesStayConfined TestReservedSetsCoverTheirOwnNamespace TestIdentifierPrefixInvariance` | ✅ | ✅ green |
| 02-05-01-S9 | 05 | 5 | OWN-01 | T-02-06 | No Go file anywhere in the module spawns an unbounded child: the scan walks every `*.go` in the module and rejects both the merged-output helper (one unbounded buffer for both streams) and the context-free spawn constructor (a child with no deadline). Needles are assembled at runtime so the scanner never matches itself; the allowlist is empty; the scan fails rather than passing vacuously if it finds no Go sources | repo-wide source invariant | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/native TestSourceNeverSpawnsUnboundedProcesses` | ✅ | ✅ green |
| 02-02-02-S10 | 02 | 2 | OWN-02 | T-02-02/T-02-04 | `borrow` is gated on `AbilityShare` at the check layer, agreeing with the independent `corevalidate` denial of `OpBorrowShared` without `AbilityShare`; the gate reports `ownership.borrow_requires_share` with repairs and emits no core it cannot authorize. Because `Buffer` now grants `share`, the law has no source-reachable negative, so `TestShareIsUniversallyGrantedAfterBufferShare` is an executable tripwire that fails the moment any source-reachable type withholds `share` and a real negative fixture becomes required | ability gate + executable tripwire | `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestShareIsUniversallyGrantedAfterBufferShare TestOwnershipSequenceExhaustive` | ✅ | ✅ green |

### Why the prior audit passed while blockers existed

The `cbba405` audit reported 18/18 green and was, row by row, honest: every selector it ran did pass. It was nonetheless wrong at the level that matters, because the *sampling map* had two holes. Both blockers sat squarely inside those holes, so no amount of re-running the existing rows could ever have surfaced them.

**Hole 1 — the generator never reached the surface under test.** `TestGeneratedRoundTrips` asserts exactly the property `lang format` was violating: that formatter output reparses. But its `generatedProgram` only ever emitted Phase 1 `data`/`match` modules. Phase 2 added the linear surface and the property never followed it there. The single generic fixture, `ability_shapes.lang`, has no bindings, so even the hand-written corpus could not reach the `generic return type + binding` shape. The result: `lang format` silently emitted corrupted, NON-REPARSABLE source at exit 0 (identifiers fused, e.g. `take vww`) while a green property test claimed the opposite. *Closed by:* `generatedLinearProgram` and `TestGeneratedLinearRoundTrips` (02-02-03-S4) walking the linear/generic cross-product, two pinned fuzz seeds, and a `semanticTokens` token-identity assertion in **both** generated properties so a fusing formatter fails even when its output happens to reparse. Independently confirmed load-bearing: with `format.go` reverted to `890bcd0^`, the property and both new seeds fail with `syntax.expected_linear_result`.

**Hole 2 — the independent oracle encoded the same wrong law.** Loan liveness was not transitive: one hop of reborrow (`let c = borrow b`) let through a move the direct form correctly rejects. The differential test *did* run, and it *did* agree — because the oracle computed liveness the same way production did. A differential is only as independent as its second derivation; this one was a paraphrase, not an independent model, so the two sides agreed on the wrong answer. *Closed by:* re-deriving the oracle by an explicitly different method (materialize the derivation relation as an edge set, close it by order-independent fixed-point iteration, then reduce to the maximum ordinal) where production streams forward once carrying an inherited association and never materializes the relation — neither derivation is obtainable from the other by renaming. Reinforced at the other admission layer (`TestTransitiveLoanBlocksMove` in `corevalidate`), at the source surface (`TestLoanLivenessIsTransitive`), and by a ninth negative control with its own fixture (02-03-02-S6, 02-06-02-S7). Independently confirmed load-bearing: with `check.go` and `corevalidate.go` reverted to `3d9493a^`, all three fail.

**The generalizable lesson.** Both holes are the same failure in two costumes: *a green test whose reachable input space does not contain the requirement's hard case.* Neither would have been caught by re-running rows, adding rows for existing tests, or measuring coverage of the production code — the production code was fully executed in both cases. They are caught only by asking, per row, "what input would falsify this, and can the harness actually generate it?" This audit therefore mutation-checked or reverted every new row rather than accepting a passing run as evidence.

## Required Negative Controls

The phase verifier fails closed unless it observes all of these exact IDs (nine as of `3d9493a`/`da75e95`; the ninth is asserted at both the session layer and through the shipped CLI):

```text
control:ownership.use_after_move
control:ownership.move_while_borrowed
control:ownership.move_while_reborrowed
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

**Approval:** validated against reviewed code `HEAD` `da75e95` from fresh-cache task selectors, ten post-gate supplement selectors (three from the `cbba405` re-audit, seven added here), three negative harness probes, and the complete Phase 2 gate on 2026-09-03. One stale selector (02-05-01) was found red and re-pointed at its strictly stronger successor before approval.

## Measured Validation Observations

Fresh `sh scripts/verify-phase2.sh` execution at `da75e95` with `GOCACHE=/tmp/ai-lang-renyquist` exited 0. Phase 1 remained green with five lanes and 18 recomputed work units. Phase 2 was green with five lanes, all **nine** required controls, **52** recomputed work units, and the coordinated-source/core escape reported only under `expected_escapes`, never as a detected control.

| Command lane | Warm samples | p50 | p95 | Min | Max | Output bytes | Work |
|--------------|-------------:|----:|----:|----:|----:|-------------:|-----:|
| format | 20 | 52,459 ns | 62,250 ns | 40,750 ns | 70,209 ns | 248 | 1 |
| check | 20 | 67,209 ns | 145,625 ns | 54,417 ns | 166,792 ns | 302 | 52 |
| interpreter | 20 | 207,292 ns | 275,875 ns | 187,084 ns | 327,875 ns | 902 | 1 |
| native | 20 | 397,719,125 ns | 499,719,542 ns | 355,868,083 ns | 514,990,958 ns | 905 | 3 |
| full verify | 20 | 862,666,916 ns | 1,166,832,791 ns | 769,027,583 ns | 1,191,176,917 ns | 1,752 | 52 |

Peak RSS was reported honestly as unavailable for every lane. All values are operational observations excluded from semantic identity and are never ratified as release SLOs. These samples were taken on a host running concurrent agent workloads, so they are very likely inflated — in particular `check` p95 (145,625 ns) is more than double its own p50, which is a scheduling artefact and not a signal about the checker. Treat every number here as an upper bound observed under contention.

## Validation Audit 2026-09-03 (re-pinned to `da75e95`)

| Metric | Count |
|--------|------:|
| Actual tasks mapped | 15 |
| Post-gate supplement rows mapped | 10 (3 from `cbba405`, 7 new) |
| Mapped targeted selectors green | 14/14 (13/14 as written; 02-05-01 was red and required correction) |
| Supplement selectors green | 10/10 |
| Full phase gates executed and green | 1/1 |
| Negative harness probes correct | 3/3 |
| Total selector executions green | 25/25 after correction (24/25 as the map stood at audit start) |
| Gaps found | 8 |
| Resolved | 8 |
| Escalated | 0 |
| Manual-only requirements | 0 |

Audit notes:

- Code `HEAD` was exactly `da75e95133fa9f90ebb74bef9cb382cf9e5baf3e` (`test(02): assert the reborrow control at the CLI layer`). Nine commits landed after the previously audited `cbba405`: `890bcd0`, `2d98a78`, `d9b370f`, `2f2e2e2`, `8d87a1a`, `dd9c0a8`, `2549311`, `3d9493a`, `da75e95`. All are covered here.
- Every selector in the Per-Task Verification Map and both Post-Gate Supplements was re-executed **verbatim** against `GOCACHE=/tmp/ai-lang-renyquist`. No status was changed without running the row.
- **Gap 1 (resolved, and the only red row).** Selector 02-05-01 named `TestNativeNeverUsesCombinedOutput`, which `8d87a1a` deleted. The harness behaved correctly and failed closed: `assert-go-tests: target not discovered: TestNativeNeverUsesCombinedOutput`, exit 1. This is a live demonstration that the harness catches map rot rather than silently passing a vanished target. The row was re-pointed at the strictly stronger successor `TestSourceNeverSpawnsUnboundedProcesses` (module-wide rather than `native.go`-only, and covering the context-free-spawn pattern in addition to merged output). Re-run green. The assertion was strengthened, never weakened.
- **Gaps 2–8 (resolved).** Seven behaviours from the nine new commits had passing tests but no task-mapped selector: linear/generic formatter round-trip with token identity, evidence canonical-projection fail-closed, transitive loan liveness at both admission layers, the ninth control's session/CLI parity, the cgen prefix-confinement and honest-reservation invariants, the repo-wide unbounded-spawn guard, and the `borrow`-on-`AbilityShare` gate. Each is mapped in Post-Gate Supplement 2 (02-02-03-S4, 02-06-01-S5, 02-03-02-S6, 02-06-02-S7, 02-07-01-S8, 02-05-01-S9, 02-02-02-S10). Every row points at a test that already exists in the tree; nothing was invented and no new test was written by this audit.
- **Falsifiability was independently re-established, not taken on trust.** In a throwaway detached worktree: reverting `syntax/format.go` to `890bcd0^` failed `TestGeneratedLinearRoundTrips` and both new `FuzzParseFormat` seeds with `syntax.expected_linear_result` on the fused identifier; reverting `check/check.go` and `corevalidate/corevalidate.go` to `3d9493a^` failed all three transitive-liveness tests, with the check-layer oracle disagreeing with production on `LoanFinalUses` and `ActiveLoans`; mutating `cLocal`'s prefix from `lang_value_` to `v_` failed both `TestIdentifierPrefixInvariance` and `TestGeneratedIdentifierNamespacesStayConfined`; and planting a `CombinedOutput` spawn in a new `internal/compiler/cgen` file was caught by `TestSourceNeverSpawnsUnboundedProcesses` for both patterns, going green again on removal.
- Negative harness probes on `scripts/assert-go-tests.sh` all behaved correctly: an absent target failed `target not discovered` with exit 1; a malformed target failed `invalid exact Go target: Test*Glob` with exit 2; an empty target list printed usage and failed with exit 2.
- `sh scripts/verify-phase2.sh` exited 0. Phase 1 reported five green lanes and 18 recomputed work units. Phase 2 reported five green lanes, **52** recomputed work units (was 51), and all **nine** required control IDs: `control:ownership.use_after_move`, `control:ownership.move_while_borrowed`, `control:ownership.move_while_reborrowed`, `control:ownership.transfer_requires_take`, `control:ability.forged_copy`, `control:core.duplicate_operation_id`, `control:interpreter-o0-o3-owned`, `control:evidence.core_mismatch`, `control:backend.runtime_causality`. `escape:coordinated-source-core-lie` appeared only in `expected_escapes` and never as a detected control.
- Known residual weakness, recorded rather than hidden: `TestGeneratedIdentifierNamespacesStayConfined` reaches `matchFixedNames`/`linearFixedNames` through an `export_test.go` seam introduced by the same commit as the invariant, so it cannot be falsified by reverting `cgen.go` alone (that combination does not build). It was falsified by direct source mutation instead. Any future test that ships its own seam should be mutation-checked the same way; a revert-based falsifier is not available for it.
- Second residual weakness: `TestShareIsUniversallyGrantedAfterBufferShare` guards a law that currently has no source-reachable negative, because `Buffer` gained `share` in `2549311`. The `borrow` gate's negative path is therefore reached only through a deliberately synthetic `nonShareableTypeFact`. That is honest and the tripwire fires the moment a source-reachable type withholds `share`, but until such a type exists the gate is not exercised end to end from source.
- The backend causality test still uses an exact-one emitted-C mutation, real Clang at both `-O0` and `-O3`, and asserts `native.engine_mismatch` with exit 4. Bounded syntax recovery, 10/100/1,000/10,000 checker work, 101/1,001/10,001 validator work, four independent native streams, strict one-document decoding, and the 64 KiB generated-output ceiling all retain active passing controls.
- Phase 1 schema, diagnostic, execution, and evidence regressions remain exercised by task selectors, both supplements, and the complete gate. `git diff cbba405 da75e95 -- testdata/phase1/` is empty, so every frozen Phase 1 golden is byte-identical to the previously reviewed state; `testdata/phase2/evidence.golden.json` changed by exactly one line under `2549311` (`Buffer` gaining `share`), which is the intended semantic change and is asserted by `TestCanonicalEvidence`.
- No Phase 2 requirement depends on manual-only verification. No gap was escalated.


## Close-Out Addendum (2026-09-04, `3399ddc`)

This audit was pinned to `da75e95`. One commit landed after it: `3399ddc`, fixing a
formatter regression that a concurrent deep review caught — a `//` comment trailing a
declaration header erased brace classification and produced non-reparsable output at
exit 0.

That defect is a direct instance of the sampling hole this audit already named: the
generator emitted comments only before `module` and before a binding, never trailing
a header, which is the one placement the classifier was sensitive to. Both generators
now emit that placement (`caseID%11` on the `fn` header, `caseID%13` on the `match`
header), and the coverage was mutation-killed — reverting `format.go` alone fails
`TestGeneratedLinearRoundTrips` with `borrow hold1hold2`.

Re-run at `3399ddc`: `go test ./...`, `go test -race ./...`, `go vet ./...`, and
`sh scripts/verify-phase2.sh` all exit 0; Phase 1 work 18; Phase 2 work 52; nine
controls; frozen goldens byte-identical.

The audit's own conclusion stands and is worth repeating: the sampling map is
adequate for the known blockers and better *in kind*, but is not proven adequate
against unknown ones. Wave 7 confirmed that caution was justified. Remaining
validation-relevant debt is recorded in `02-DEBT.md` (D-02-01 spawn-guard strength,
D-02-04 missing `native.timeout` falsifier) plus the two residual weaknesses this
file already records.

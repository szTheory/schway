# Phase 36 Acceptance Evidence

Evidence date: 2026-10-15
Source revision: 5474f16ecd3e80cf7cdafec73173b2af48c6556f
Host/toolchain: macOS arm64; Go 1.24.0; Apple Clang 21.0.1
Application/conformance routes: Public `schway` command conformance, interpreter application and fixture routes, native application C, and direct local/importer helper conformance are separate observations; no hosted receipt is claimed.

## Acceptance interpretation

The dated clarification in D-18 preserves the original CALL-01/CALL-02 wording while assigning each claim to its actual knower. New U64/Bool callee defects are overflow and remainder. The legacy owning-wrapper-to-ByteView callee bounds defect remains a distinct compatibility regression. A ByteView caller that computes a helper argument has a separate before-entry bounds obligation; an out-of-bounds `byte_at` does not enter the scalar helper. The checker proves source argument/result lowering, raw peers independently check their IR contracts, engines prove literal values and application outcomes, and the execution peer proves causal attribution without proving arithmetic or source intent.

CALL-01 and CALL-02 remain `unclassified/unresolved` because no SPEC classifies the listed edges. The two descriptorless prohibitions remain `unverified`; this receipt does not claim wired enforcement. Hosted two-host binding belongs to Phase 38. Human readability and full generate/verify/run/observe/repair cold/warm distributions remain unmeasured.

## Claim evidence and result

| Claim | Evidence | Result |
|---|---|---|
| frozen-c | `internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaseline` | PASS — frozen no-call C comparison passed for the tracked source inventory and emitter routes. |
| old-refusal-inventory | `internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaseline` | PASS — frozen emitting and non-emitting rows remained enumerated and matched their recorded outcomes. |
| new-fixture-ledger | `internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaselineControls` | PASS — the three exact helper example rows and omission/false-exemption controls passed. |
| legacy-byteview-bounds | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36LegacyByteViewCalleeBounds` | PASS — interpreter and native O0/O3/O3-LTO owning-wrapper bounds regressions passed. |
| static-and-runtime-capacity | `internal/compiler/cgen/cgen_program_test.go#TestInvocationPathTableBoundary`; `internal/compiler/cgen/cgen_phase36_test.go#TestPhase36CallEventCapacityBoundary` | PASS — static invocation-path and runtime event/output exact/plus-one boundaries passed. |
| route-input-and-capture-separation | `cmd/schway/main_phase36_test.go#TestPhase36PublicHelperConformance`; `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeCallCaptureOutcomeSeparation` | PASS — public interpreter/native routes and disabled/exhausted capture outcome controls passed. |
| helper-values | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperAnswersAndDefects` | Awaiting fixed final capture |
| callee-defects | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperOptimizationMatrix` | Awaiting fixed final capture |
| caller-bounds-before-entry | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36ByteViewCallerArgumentBoundary`; `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeByteViewCallerArgumentBoundary` | Awaiting fixed final capture |
| independent-raw-contracts | `internal/compiler/corevalidate/corevalidate_phase36_test.go#TestPhase36RawCallShapeAndFreshness`; `internal/compiler/originvalidate/originvalidate_phase36_test.go#TestPhase36OriginCallFreshnessAndTransfer` | Awaiting fixed final capture |
| causal-only | `internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls` | Awaiting fixed final capture |
| capture-outcome-separation | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeCallCaptureOutcomeSeparation` | Awaiting fixed final capture |
| manifest | `internal/compiler/session/session_phase36_surface_test.go#TestPhase36SurfaceManifestCallWitnesses` | Awaiting fixed final capture |

## Evidence categories

Source inspection: Current checker and engine boundaries are in `internal/compiler/check/scalar.go`, `internal/compiler/interp/interp.go`, `internal/compiler/executionpeer/executionpeer.go`, and `internal/compiler/cgen/cgen.go`; local/importer programs are `examples/helper_calls/crc_step.schway` and `examples/helper_calls/importer.schway`. Inspection is not a fresh execution claim.

New focused checks: Compatibility capture passed 15 root selectors in three owner packages with exit 0 and no FAIL/SKIP: `internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaseline`, `internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaselineControls`, `internal/compiler/session/session_phase36_calls_test.go#TestPhase36LegacyByteViewCalleeBounds`, `internal/compiler/cgen/cgen_program_test.go#TestInvocationPathTableBoundary`, `internal/compiler/cgen/cgen_phase36_test.go#TestPhase36CallEventCapacityBoundary`, `cmd/schway/main_phase36_test.go#TestPhase36PublicHelperConformance`, and `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeCallCaptureOutcomeSeparation`. Scratch constructibility probes in `36-VALIDATION.md` are planning-checker observations, not compiler feature evidence.

Historical receipts: Historical only; not fresh Phase 36 evidence. [Phase 35 verification](.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md) passed on 2026-10-14 and remains scoped to its recorded revision and host.

## Requirement, decision, and threat mapping

| Owner | Evidence represented here |
|---|---|
| CALL-01 | `helper-values`, `callee-defects`, `caller-bounds-before-entry`, `route-input-and-capture-separation`; D-18 keeps the old ByteView callee bounds regression separate from helper argument bounds. |
| CALL-02 | `frozen-c`, `new-fixture-ledger`, `independent-raw-contracts`, `causal-only`, `capture-outcome-separation`, and `manifest`; no peer is credited with numeric or source-intent proof it does not derive. |
| D-15/D-18 | Numeric engine results, causal attribution, source lowering, and the two bounds obligations have separate named owners. |
| D-19/D-22/D-24/D-25 | Frozen bytes and old rows stay immutable; only exact fixed owner commands and their named roots count. |
| D-21/D-26 | Public/application routes are distinct; next capabilities and dormant seeds are recorded in both dated living amendments. |
| T-36-22 | Frozen C, omission, false-exemption, and exact fixture-ledger controls are owned by the compatibility capture. |
| T-36-23 | Claim-to-owner mappings, source/fresh/historical categories, and unresolved flags remain explicit. |
| T-36-SC | No package installation or new dependency is part of this plan. |

## Fixed owner captures

Only the exact package and selector lists embedded in `36-08-PLAN.md` are valid for these receipts. Each successful capture records the full argv, `GOCACHE`, source HEAD, clean tracked-source binding, exit code, expected root selectors, capture time, bounded combined output, and SHA-256 in its read-only `owner-results.json` and logs. The checker rejects missing, duplicate, skipped, failed, unowned, or prose-invented roots.

Observed owner commands (compatibility): PASS — 15/15 expected root selectors across the session, cgen, and public command owners; each command exited 0. Full argv, roots, hashes and logs are in [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/owner-results.json](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/owner-results.json) and the linked logs appended below.

Observed owner commands (final): awaiting the fixed final capture.

Observed owner commands (compatibility, HEAD 5474f16ecd3e80cf7cdafec73173b2af48c6556f): [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/owner-results.json](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/owner-results.json)
- [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/1-session.log](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/1-session.log), sha256 532732e9b01633f18df764db593f14c1c7ff6b506055e2ccdf4cbc11b554d41b
- [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/2-cgen.log](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/2-cgen.log), sha256 6244a598ae46877ea49d8883dc1fb84c08c7df8947e98a92da3abf0581a4828b
- [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/3-schway.log](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/3-schway.log), sha256 6c03673c479a60f57a4921e81df6d02ba9c7bb2ec59d8b2468317da6f63957c8

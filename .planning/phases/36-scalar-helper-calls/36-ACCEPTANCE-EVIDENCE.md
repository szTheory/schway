# Phase 36 Acceptance Evidence

Evidence date: 2026-10-15
Source revision: 83c1f58a1720770e77216ee82737efa3acdbeeaa
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
| helper-values | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperAnswersAndDefects` | PASS — local/imported four-step answers and the U64/Bool value cases passed; overflow and zero-remainder defects were reached. |
| callee-defects | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36HelperOptimizationMatrix` | PASS — local/imported O0, O3, and O3-LTO answer and defect rows passed; UBSan passed separately. |
| caller-bounds-before-entry | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36ByteViewCallerArgumentBoundary`; `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeByteViewCallerArgumentBoundary` | PASS — interpreter OOB arguments stopped before helper entry; native first/last/in-range/past-end/empty rows passed at O0, O3, and O3-LTO. |
| independent-raw-contracts | `internal/compiler/corevalidate/corevalidate_phase36_test.go#TestPhase36RawCallShapeAndFreshness`; `internal/compiler/originvalidate/originvalidate_phase36_test.go#TestPhase36OriginCallFreshnessAndTransfer` | PASS — core and origin owners independently accepted exact calls and unknown-result transfer; their malformed-field, signature, and target-freshness controls passed. |
| causal-only | `internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls` | PASS — ordinal gaps/duplicates/order, wrong parent/site/callee/places/type, missing call/return, and repeated-terminal controls passed without numeric replay. |
| capture-outcome-separation | `internal/compiler/session/session_phase36_calls_test.go#TestPhase36NativeCallCaptureOutcomeSeparation` | PASS — disabled/exhausted native evidence preserved successful work and late defects; the native owner also passed its exhaustion/disabled cases. |
| manifest | `internal/compiler/session/session_phase36_surface_test.go#TestPhase36SurfaceManifestCallWitnesses` | PASS — witness, census, copied-data controls, limit consumers, and source-free document owners passed. |

## Evidence categories

Source inspection: Current checker and engine boundaries are in `internal/compiler/check/scalar.go`, `internal/compiler/interp/interp.go`, `internal/compiler/executionpeer/executionpeer.go`, and `internal/compiler/cgen/cgen.go`; local/importer programs are `examples/helper_calls/crc_step.schway` and `examples/helper_calls/importer.schway`. Inspection is not a fresh execution claim.

New focused checks: Compatibility capture passed 15/15 expected root selectors across the session, cgen, and public command owners. The amended final capture passed 45/45 expected roots across 12 owner packages with exit 0 and no FAIL/SKIP; it includes `internal/compiler/check/check_phase36_scalar_calls_test.go#TestPhase36CallSourceIntentControls`, `internal/compiler/corevalidate/corevalidate_phase36_legacy_internal_test.go#TestPhase36CallCFGApplicability`, `internal/compiler/originvalidate/originvalidate_phase36_test.go#TestPhase36OriginCallFreshnessAndTransfer`, `internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go#TestPhase36ScalarCallLocalBorrowControls`, `internal/compiler/interp/interp_phase36_test.go#TestPhase36InterpreterCallAnswersAndDefects`, `internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls`, and `internal/compiler/session/session_phase36_peer_mutation_test.go#TestPhase36EngineCallMutationControls`. Logs below bind the complete selector set and output digests. Scratch constructibility probes in `36-VALIDATION.md` are planning-checker observations, not compiler feature evidence.

Historical receipts: Historical only; not fresh Phase 36 evidence. [Phase 35 verification](.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VERIFICATION.md) passed on 2026-10-14 and remains scoped to its recorded revision and host.

Resolved integration-debug records are historical explanations for the final owner set: [legacy call routes](../../debug/resolved/phase36-legacy-call-routes.md), [borrow call paths](../../debug/resolved/phase36-borrow-call-paths.md), and [ByteView helper composition](../../debug/resolved/phase36-byteview-helper.md). They do not substitute for the amended fixed final capture below.

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

Prior final capture (source HEAD d117ca4058fbf6a727cb5aa634768cf6194bac34): PASS — 45/45 expected roots across 12 packages; all commands exited 0 with no FAIL/SKIP. This remains an accurate receipt for that revision; the audit-follow-up captures below bind added controls to their respective test revisions.

## Security-audit follow-up — 2026-10-15

Test-only commit `04ce6fa1` expands `internal/compiler/cgen/cgen_phase36_test.go#TestPhase36EmitterCallInventoryControls`. Its returned stack-path mutant redirects the invocation-table pointer to a local array that a loop reuses; emitted native code still compiles, runs, and returns 10, while the retained-path assertion identifies the missing first canonical child path. Clean, stack-copy no-op, and restored positives retain the first and final paths. The same selector applies six region-scoped removals/additions across call sites, `function.called` events, and invocation reservations, checks named missing/extra findings, and verifies the emitted region is clean again after each mutation.

Test-only commit `50b90a30` corrects the T-36-16 census control: it counts actual `SCHWAY_` invocations after stripping comments and strings, and derives expected callees from checked `core.OpCall` contracts plus emitted definitions. It catches both a marker-preserving removal of a real call and an unmarked extra call. The prior capture at HEAD `eddb535ae9fa10f3700143e02f7fc5c11106e8d5` remains accurate for the commands it ran, but predates this census correction and does not evidence the corrected T-36-16 controls. The fixed final owner command below captures the corrected selector at source/test HEAD `83c1f58a1720770e77216ee82737efa3acdbeeaa`; intervening `05-summary` changes are planning-document-only. The independent security review remains open pending re-audit.

Observed owner commands (compatibility, HEAD 5474f16ecd3e80cf7cdafec73173b2af48c6556f): [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/owner-results.json](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/owner-results.json)
- [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/1-session.log](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/1-session.log), sha256 532732e9b01633f18df764db593f14c1c7ff6b506055e2ccdf4cbc11b554d41b
- [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/2-cgen.log](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/2-cgen.log), sha256 6244a598ae46877ea49d8883dc1fb84c08c7df8947e98a92da3abf0581a4828b
- [/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/3-schway.log](/tmp/schway36-compatibility-5474f16ecd3e-1fge23bk/3-schway.log), sha256 6c03673c479a60f57a4921e81df6d02ba9c7bb2ec59d8b2468317da6f63957c8

Observed owner commands (compatibility, HEAD d117ca4058fbf6a727cb5aa634768cf6194bac34): [/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/owner-results.json](/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/owner-results.json)
- [/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/1-session.log](/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/1-session.log), sha256 63d3d4c60cb8dbe3900965d49da2276ed8abec885eac2d6869f99588daaec42f
- [/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/2-cgen.log](/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/2-cgen.log), sha256 ec3f6a6e0270b2ac449e15d37961c2be49a1cd44a971b148bacd7a185def6e62
- [/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/3-schway.log](/tmp/schway36-compatibility-d117ca4058fb-w_dxuf4j/3-schway.log), sha256 756ca3ffd983acf25e7e7480d866957eccdf6ea118bf30da0fd2bacb6e8a3157

Observed owner commands (final, HEAD d117ca4058fbf6a727cb5aa634768cf6194bac34): [/tmp/schway36-final-d117ca4058fb-lo7ctopk/owner-results.json](/tmp/schway36-final-d117ca4058fb-lo7ctopk/owner-results.json)
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/1-syntax.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/1-syntax.log), sha256 cadb6d9849e08ab930ecbeb98b50dadb003b7f9a07b7371456fe3fe52e77cfee
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/2-check.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/2-check.log), sha256 e3185cc9810966b9a34c4ae3ec12398adfd8ed79158ce3f5bb65547fcec0ad48
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/3-corevalidate.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/3-corevalidate.log), sha256 bce07457cef51c1033500f61bfa251e36008c21c83ab9ba6c205d97d9b21cd80
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/4-originvalidate.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/4-originvalidate.log), sha256 1b6ccfbfaba61226276a876b844722d5c3215db559428cb3759b67e0e49024cb
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/5-pathoracle.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/5-pathoracle.log), sha256 67e3baf4746c26d0c037ba06e48c1d56ba23650c0f49d09a03a6f5f8aca33fe6
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/6-interp.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/6-interp.log), sha256 4db77a0eec93657b49e94a1cde16f19e40d504c9cedf85909f9629944d0e1c43
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/7-executionpeer.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/7-executionpeer.log), sha256 956c1331ad69710b33a1feb1a4ad8b216d7fde25f994c6c30a4286182bb53796
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/8-cgen.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/8-cgen.log), sha256 b458bacdf344d3ed2d422b1abb83e19981b2bd118cd3371b047a560ffd3d5c18
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/9-native.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/9-native.log), sha256 e65ee6108c7e84ba8831a61283a8550e1ac8c31fb27fe1aa9359eb9f66676ab4
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/10-session.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/10-session.log), sha256 6200805b42c102197041eaf701deb80b2d2696cf9f9fbe2e49ebdf585b408738
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/11-schway.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/11-schway.log), sha256 86e6b8c08db6b5226d30becda8daa256e0c96fc0b8ba5544662a0697d0087286
- [/tmp/schway36-final-d117ca4058fb-lo7ctopk/12-surfacemanifest.log](/tmp/schway36-final-d117ca4058fb-lo7ctopk/12-surfacemanifest.log), sha256 4e59c6ea26ff967f54daff0c35cb86e5a557af9076b767d68312911fce5c84c2

Observed owner commands (final, HEAD d117ca4058fbf6a727cb5aa634768cf6194bac34): [/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json](/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json)
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/1-syntax.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/1-syntax.log), sha256 7a2415e7b66318ab720c1031fe4f72801b69fad21b40d8488f19cf1f55f01290
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/2-check.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/2-check.log), sha256 1a59756433e4a5289ffae18984bd1f44a1f5a8b521a1c68ec48e4e5b8de288f9
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/3-corevalidate.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/3-corevalidate.log), sha256 bce07457cef51c1033500f61bfa251e36008c21c83ab9ba6c205d97d9b21cd80
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/4-originvalidate.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/4-originvalidate.log), sha256 1b6ccfbfaba61226276a876b844722d5c3215db559428cb3759b67e0e49024cb
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/5-pathoracle.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/5-pathoracle.log), sha256 67e3baf4746c26d0c037ba06e48c1d56ba23650c0f49d09a03a6f5f8aca33fe6
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/6-interp.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/6-interp.log), sha256 19a4c89b05255c3b20c3c8c52d00df4ff37c4a6ef388389947a626694e2e091b
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/7-executionpeer.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/7-executionpeer.log), sha256 9707b4078e9b37b9402bf44b0c068d22a830eb3372fd110bd6c4316e9fad011a
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/8-cgen.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/8-cgen.log), sha256 5d71daaa631836b82a8a1ae0ebb6bcf746d11dad9af1433465a9a22f6b3f27bc
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/9-native.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/9-native.log), sha256 a07bfccccbf0b6ee8258dc7ac966b481d10867d82735ed9ee26c5e135c5c54a0
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/10-session.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/10-session.log), sha256 c2a81284bb4a3d7ecbc1f18c7a33ee5b1889e0516e5a112516f7036778b536de
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/11-schway.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/11-schway.log), sha256 db8d312e2401a130701cc0c0c3f35f3bcfe09d862042a7f04a47d5368612ab01
- [/tmp/schway36-final-d117ca4058fb-evx1nc2r/12-surfacemanifest.log](/tmp/schway36-final-d117ca4058fb-evx1nc2r/12-surfacemanifest.log), sha256 4e59c6ea26ff967f54daff0c35cb86e5a557af9076b767d68312911fce5c84c2

Observed owner commands (final, HEAD eddb535ae9fa10f3700143e02f7fc5c11106e8d5): [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/owner-results.json](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/owner-results.json)
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/1-syntax.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/1-syntax.log), sha256 48f10e44cfab0d212c0f724a78a811534098ab8e3855fc16ccd8ebb72ab90e9a
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/2-check.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/2-check.log), sha256 f67033844e27ad3909a09065ad132a763929bca838bb469b2b7d83a9191ad3df
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/3-corevalidate.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/3-corevalidate.log), sha256 b04b46c47119e5ddd7232f44653d9ad91c8d845a889e286542acfcfa99f2acf7
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/4-originvalidate.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/4-originvalidate.log), sha256 1b6ccfbfaba61226276a876b844722d5c3215db559428cb3759b67e0e49024cb
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/5-pathoracle.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/5-pathoracle.log), sha256 67e3baf4746c26d0c037ba06e48c1d56ba23650c0f49d09a03a6f5f8aca33fe6
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/6-interp.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/6-interp.log), sha256 19a4c89b05255c3b20c3c8c52d00df4ff37c4a6ef388389947a626694e2e091b
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/7-executionpeer.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/7-executionpeer.log), sha256 c8b09ee197361df7aba8780871937005c5f6f18925f5465344117564909d45b7
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/8-cgen.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/8-cgen.log), sha256 76af4b7f1d70098c756e346d441b97ecd7ea8a04b648c15433d1f055099ea0af
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/9-native.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/9-native.log), sha256 86cc37c9dcece2884ce6c6bd9368ff9333398c16135fe1d52265a831289434f7
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/10-session.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/10-session.log), sha256 47d71ecbe86e549040e28a1be493fbfe053c6950fe16c0f347299da52130473d
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/11-schway.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/11-schway.log), sha256 893d663a014a3594ccd646861e2583898cc3ceb16f2936e0191c8da6a28919e6
- [/tmp/schway36-final-eddb535ae9fa-g7t84c5o/12-surfacemanifest.log](/tmp/schway36-final-eddb535ae9fa-g7t84c5o/12-surfacemanifest.log), sha256 4e59c6ea26ff967f54daff0c35cb86e5a557af9076b767d68312911fce5c84c2

Observed owner commands (final, HEAD 83c1f58a1720770e77216ee82737efa3acdbeeaa): [/tmp/schway36-final-83c1f58a1720-jikc_g2_/owner-results.json](/tmp/schway36-final-83c1f58a1720-jikc_g2_/owner-results.json)
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/1-syntax.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/1-syntax.log), sha256 7a2415e7b66318ab720c1031fe4f72801b69fad21b40d8488f19cf1f55f01290
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/2-check.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/2-check.log), sha256 4933328e7b113d51773095ca8a143f818cce28de4947422dff733eb0171cfb01
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/3-corevalidate.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/3-corevalidate.log), sha256 4030f781cd00d164bbdd77dc76c1266a189239669da51d321dcd94d1ecfb47f5
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/4-originvalidate.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/4-originvalidate.log), sha256 1b6ccfbfaba61226276a876b844722d5c3215db559428cb3759b67e0e49024cb
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/5-pathoracle.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/5-pathoracle.log), sha256 67e3baf4746c26d0c037ba06e48c1d56ba23650c0f49d09a03a6f5f8aca33fe6
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/6-interp.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/6-interp.log), sha256 660825616fe4bc6fa1fe9ddd41278e7111f161ee240299f2c861e870d229d421
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/7-executionpeer.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/7-executionpeer.log), sha256 9707b4078e9b37b9402bf44b0c068d22a830eb3372fd110bd6c4316e9fad011a
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/8-cgen.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/8-cgen.log), sha256 336c9a2e14d6f468bc32a3c097fc7d18151fb68227f5795693ab098a51235c6d
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/9-native.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/9-native.log), sha256 21398503e6f7ce5743a46d84ba36a4644f34e25ff27ad9c9fb47093b87cfbded
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/10-session.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/10-session.log), sha256 02d683d13617df2ee0ac09b5d3352a1ec903ee2413ced56e2bc493cd3ed593f4
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/11-schway.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/11-schway.log), sha256 3bd8e51717a94ac23dbd09b39850cc28d72befbdc27cc595cdf0c432212e83ad
- [/tmp/schway36-final-83c1f58a1720-jikc_g2_/12-surfacemanifest.log](/tmp/schway36-final-83c1f58a1720-jikc_g2_/12-surfacemanifest.log), sha256 4e59c6ea26ff967f54daff0c35cb86e5a557af9076b767d68312911fce5c84c2

---
status: resolved
trigger: "Phase36 native plan found three frozen no-call corpus core-validation refusal-route shifts before C generation."
created: 2026-10-14
updated: 2026-10-14
---

## Current Focus

bug_class: Bohrbug (deterministic route mismatch reproduced)
hypothesis: Confirmed overbroad scalar-call applicability across the legacy file-owner and fallible PathToken call families; scoped correction is self-verified.
test: Final restored exact frozen routes, new regression roots and raw/lifecycle controls pass; seven independent fix-site overlays are semantically killed.
expecting: Resolved under the user's existing authorization for researched routine fixes; no human-verification claim is made.
next_action: Resume Plan36-06; this debug session's three-route implementation blocker is cleared.

reasoning_checkpoint:
  hypothesis: b717e4f5 misclassifies legacy owning wrapper and PathToken error calls as scalar helper CFGs, and omits the established file-owner ByteView lifecycle at the mixed-call gate.
  confirming_evidence:
    - Exact current frozen baseline refuses three previously emitting application rows; all routes and controls pass with only corevalidate.go overlaid from b717e4f5^.
    - Agent-authored TestPhase36CallCFGApplicability fails on file view/result/tail, PathToken probe and fixture view classification; TestPhase36LegacyFileOwnerCalls fails both five/six-operation admitted owner bodies with core.scalar_call_context_invalid.
  falsification_test: If narrowing call-only CFG classification and adding the independently checked full file-owner lifecycle fails to restore frozen routes, or accepts malformed scalar/resource/back-edge controls, the hypothesis/fix is insufficient.
  fix_rationale: Require both U64/Bool operands and a scalar-compatible caller formal for call-only scalar analysis; explicit scalar operations still trigger all resource/back-edge laws. Allow the existing ByteView-to-U64 file wrapper only when its entire pre-existing lifecycle matcher succeeds. Preserve exact callee typing and initialized arguments independently.
  blind_spots: No full-suite or cross-host execution authorized; validate focused ownership, peer and frozen-emission neighbors locally. Origin/path peers may reveal independent applicability bugs after the core fix.
  candidate_causes:
    - code: overbroad scalar applicability and missing file-owner lifecycle branch (confirmed by one-file overlay).
    - data: stale or unsound frozen fixtures (contradicted by unchanged sources and restored exact outputs under pre-feature corevalidate).
    - environment: emitter/toolchain drift (contradicted by same-host one-file overlay restoring all frozen bytes/routes).
  and_gate: No simultaneous runtime preconditions are needed; two independent new applicability mistakes cause different legacy rows to fail. Both must be corrected for the reported corpus to pass.

## Symptoms

expected: Existing sources without new scalar helper calls retain their admitted/refused emitter-route behavior and generated C bytes; malformed new scalar calls remain refused independently.
actual: Plan36-05 reports three pre-emission core-validator refusal-route shifts in examples/checksum/adler_low.schway, examples/checksum/checksum.schway, and examples/phase24/error.schway. Actual generated-C bytes otherwise match.
errors: TestPhase36NoCallGeneratedCByteBaseline remains red for the three rows; full exact output must be reproduced rather than inferred from this report.
timeline: Frozen baseline committed by Plan36-01 before parser/checker, core/origin/path, interpreter/peer, and native changes in Plans36-02..05. Current HEAD 4b9b5d46.
reproduction: GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s

## Evidence

- Source reports only: 36-05-SUMMARY.md, phase deferred-items.md and WINDOWS entry13 record the residual. Parent has not yet rerun the failing command.
- Native task selectors passed per executor; no phase completion or blanket compatibility pass is claimed.
- 2026-10-14: Resumed under authorized Phase36 debug workflow; root-pin script passed before first write. Existing unrelated edits are .planning/config.json, .planning/state.json, .planning/milestone.lock, and .planning/research/.cache/. No project-local skills found. KB offers the general pattern of overbroad new-call rules misclassifying legacy call/provenance families (phase25), not a confirmed diagnosis. SBFL skipped: no per-test coverage spectrum available for this focused run.
- 2026-10-14: Exact baseline+Controls both FAIL (2.319s). checksum/adler_low and checksum/checksum: emit and emit_native change from cgen.file_bytes_owner_shell_unsupported to core.scalar_call_context_invalid, and emit_application changes emitted to refused. phase24/error: cgen.phase24_error_application_shell_unsupported changes to core.scalar_backedge_authority, and emit_application changes emitted to refused. Controls stop at the clean-baseline gate; negative controls are not yet reached. This disproves refusal-text-only drift.
- 2026-10-14: Bounded diff b717e4f5^..b717e4f5 adds mixed scalar/non-scalar refusal allowing only fixtureViewCall (not existing fileBytesOwnerLifecycleValid), and hasScalarCFG selects every OpCall with any scalar source/result. Existing independent file-bytes lifecycle matcher admits PathToken acquire/borrow/view-call/release/optional scalar-tail/return; Phase24 has dedicated error-edge and cleanup matchers. No evidence of unsound original fixtures.
- 2026-10-14: Temporary Go overlay containing only pre-feature corevalidate.go from b717e4f5^ restores exact frozen baseline and Controls (both root RUN/PASS, 2.409s). This localizes the regression to corevalidate, not source lowering, C bytes or host configuration.
- 2026-10-14: New agent-authored TestPhase36CallCFGApplicability and TestPhase36LegacyFileOwnerCalls run RED with semantic assertions (no compile failure). Four classifier rows misclassified; both full raw owner-call positives refused. Oracle is derived from D08/D10/D19 and existing independently enforced lifecycle contracts; boundary neighbors include all four scalar pairs, Byte/ByteView callers, zero calls, explicit scalar branch, owning and fallible calls.
- 2026-10-14: Minimal two-site correction applied in corevalidate.go. Eight exact core roots PASS (0.043s): TestPhase36CallCFGApplicability, TestPhase36LegacyFileOwnerCalls, TestPhase36RawCallShapeAndFreshness, TestPhase36RawCallRefusalRemovalControls, TestPhase31FileBytesSeededFaults, TestPhase29ReleaseBeforeViewUseRefused, TestPhase29MoveOwnerWhileViewLive, TestPhase29OwnerLoopBlockRefused. Actual RUN/PASS includes wrong argument/result for all four pairs, foreign/fallible fields, missing IDs, Unit/ByteView/resource results, prior constant/copy/store targets, 5/6-op owners, dropped/early release, extra/wrong trailing calls, owner moves and R1/R2/M1 back edges.
- 2026-10-14: First correction is insufficient: frozen rerun restores checksum rows but Phase24/error now reaches core.scalar_call_context_invalid instead of the prior masking backedge failure. TestPhase24ActivationPeerDerivesRepeatedAcquisitions independently observes core=false, origin=false, path=false. Source inspection confirms origin/path callTouchesScalar gates also misclassify the existing PathToken-to-U64 fallible probe. Expand only to these evidenced contract gates, preserving their independently existing repeated-owner cleanup validators.
- 2026-10-14: Revert via temporary Go overlay of unmodified HEAD corevalidate.go reproduces all original three frozen failures (2.543s), without disturbing shared working-tree edits. Reapply verification is pending the second correction. Origin/path new raw controls and Phase35 ancestry all PASS independently on current first correction, including ByteView scalar-call and resource-result refusals.
- 2026-10-14: Second correction retains only the matching PathToken caller/argument/callee to U64 result with all three fallible fields, in each peer's own gate. Existing independent owner/error validators remain active. Exact frozen baseline+Controls and Phase24 activation/cleanup plus Phase29 fixture roots all PASS (2.362s), including all seven Phase24 lifecycle mutants and Phase23/29 separation. No frozen corpus/source changed.
- 2026-10-14: Held-out same-named TestPhase31FileBytesSeededFaults in originvalidate fails positive-6op with core.scalar_analysis_invalid and masks several expected core.file_bytes_owner_origin refusals. Core/path counterparts pass. originCallHasScalarEndpoints still selects U64/U64 tails in PathToken wrappers, and new call-contract validation precedes the old owner matcher. Newly authored TestPhase36LegacyFallibleCallRoutes PASS including each missing fallible field, wrong argument, wrong cleanup and restored baselines.
- 2026-10-14: Origin's caller-scope and refusal-order correction restores TestPhase31FileBytesSeededFaults plus both Phase36 origin raw roots (all RUN/PASS, 0.024s). Newly authored TestPhase36LegacyFileOwnerPeerRoutes passes both real checksum sources through all peers and pins old owner refusal precedence on malformed release/trailing-owner arguments.
- 2026-10-14: Seven single-site semantic overlay mutants killed by new agent-authored regressions (each compiled and reached target assertion): file-owner disables full lifecycle bridge -> TestPhase36LegacyFileOwnerCalls fails; cfg restores broad scalar-result trigger -> TestPhase36CallCFGApplicability fails; probe-core/probe-origin/probe-path remove only that peer's PathToken applicability -> TestPhase36LegacyFallibleCallRoutes fails with precisely core/origin/path false respectively; origin-cfg removes caller-formal scope -> TestPhase36LegacyFileOwnerPeerRoutes/adler_low fails only origin; origin-order restores earlier scalar-contract gate -> same test fails exact owner refusal oracle. Temporary overlays /tmp/schway36-v2-{file-owner,cfg,probe-core,probe-origin,probe-path,origin-cfg,origin-order}.json. No repository source mutation during controls. No Stryker configured (Go standard-library repo); direct bounded Go-overlay mutation used instead, 7/7 semantic kills, zero compile/vet-only kills.
- 2026-10-14T23:39-04:00: Final restored RUN/PASS: session seven selected roots, 2.489s; corevalidate selected roots, 0.042s; originvalidate selected roots, 0.026s; pathoracle selected roots, 0.026s. All commands used prescribed GOCACHE, exact named roots, -count=1 -v -timeout=600s. Frozen baseline Controls reached changed-byte, omitted-row, false route exemption, old-source exemption and unsafe-ledger controls and restored baselines. Logs /tmp/schway36-final-session.log and /tmp/schway36-final-validators.log contain commands and full verbose output. Darwin arm64 / Go 1.24.0; no broad/native/hosted/cross-host claim. git diff --check passed. No commits made; manager owns root-pin and exact-file commit.

- timestamp: 2026-10-14
  checked: Session manager independently reran both exact Final exact commands after the investigator handoff; session seven roots PASS (2.336s), core PASS (0.044s), origin PASS (0.027s), path PASS (0.026s), all selected raw/lifecycle controls reached. Root pin and git diff --check passed before archive/commit. Resolution uses existing routine-fix authorization and does not claim human verification.

reasoning_checkpoint_origin_owner:
  hypothesis: Origin's call-only scalar trigger independently selects the file-owner scalar tail, and its new contract gate masks the established owner-lifecycle oracle.
  confirming_evidence: ["Origin TestPhase31FileBytesSeededFaults positive-6op fails with core.scalar_analysis_invalid while matching core/path tests pass", "PublishProblemsFor runs new call contract/analysis before fileBytesOwnerOriginProblem; originCallHasScalarEndpoints lacks caller signature scope"]
  falsification_test: Scope only call-triggered scalar analysis and place new contract checking after the unchanged lifecycle block; all Phase31 positives/faults must regain original verdict/code and Phase36 raw controls must pass.
  fix_rationale: Keep PathToken owner bodies on their existing independent lifecycle/origin walk and retain all new scalar helper checks for U64/Bool/Byte/ByteView callers.
  blind_spots: No full suite or cross-host run; focused neighboring gates only.
  candidate_causes: ["code: scalar-call trigger/order", "data: corrupt owner fixture (core/path and frozen routes contradict)"]
  and_gate: No; the admission failure and refusal precedence share the overbroad trigger/order, with distinct minimal corrections.

reasoning_checkpoint_phase24:
  hypothesis: The legacy PathToken-to-U64 probe is independently misclassified as a U64/Bool helper by core, origin and path contract gates solely because its result is scalar.
  confirming_evidence:
    - Exact frozen first-correction rerun now exposes core.scalar_call_context_invalid at phase24/error while checksum rows disappear from failures.
    - Existing activation test independently reports core=false/origin=false/path=false; all three gate bodies require a scalar argument for every scalar-result call.
  falsification_test: Restoring only the exact legacy typed fallible PathToken-to-U64 family must pass all three independent activation peers while cleanup/error mutations stay refused and new scalar-call refusal controls remain reached.
  fix_rationale: Exclude only the old matching PathToken caller/argument/callee and U64 result with complete fallible fields from the new scalar helper contract. Existing complete owner/error-edge validators still run in each peer; this is not permission for U64/Bool helpers to consume resources.
  blind_spots: Full suite and other hosts not run; frozen exact corpus plus Phase24 mutations and new raw peer controls cover the bounded compatibility claim.
  candidate_causes: ["code: scalar-result-only trigger applied to old fallible probe", "data: malformed old cleanup (contradicted by pre-feature overlay and unchanged source)"]
  and_gate: No; this is a later independent gate masked by the original CFG refusal.

## Eliminated

- hypothesis: Only refusal wording changed with no admitted behavior lost.
  evidence: Exact frozen application routes change emitted to refused for all three source rows.
  timestamp: 2026-10-14

## Resolution

root_cause: Phase36 scalar-call applicability used scalar source/result membership as a proxy for the new U64/Bool helper family. Core omitted the existing file-owner ByteView lifecycle; core and origin call-only CFG triggers captured PathToken ownership bodies/tails; core/origin/path contract gates misclassified the existing fallible PathToken-to-U64 probe. Origin's new gate also masked established owner-lifecycle refusals.
fix: Scope call-only scalar CFG triggers in core/origin to scalar-compatible caller formals and U64/Bool endpoints; retain core's full file-owner lifecycle at ByteView mixed calls; keep the exact typed PathToken-to-U64 fallible probe on independent legacy owner/error replay in each peer; run origin's helper contract after its existing owner lifecycle refusal. No checker, fixture, corpus/golden, schema, dependency, or emitter edit.
oracle_type: derived (D08/D10/D19 plus pre-existing independent owner/error lifecycle contracts and frozen literal byte/route oracle)
verification:
  target_test: { result: pass, roots: [TestPhase36NoCallGeneratedCByteBaseline, TestPhase36NoCallGeneratedCByteBaselineControls, TestPhase36CallCFGApplicability, TestPhase36LegacyFileOwnerCalls, TestPhase36LegacyFallibleCallRoutes, TestPhase36LegacyFileOwnerPeerRoutes] }
  mutation_check: { result: skipped, reason_if_skipped: "Stryker is absent/unconfigured in this Go standard-library repository", supplementary_check: "seven bounded direct fix-site Go-overlay semantic mutants killed by agent-authored regressions", mutant_killed: "7/7" }
  no_op_deletion: { result: pass, deletion_justified_by_rca: true, evidence: "No behavior deletion or assertion weakening; two call-only classifiers narrowed, complete legacy lifecycle retained, three exact legacy fallible-call applicability cases retained, origin's unchanged lifecycle oracle ordered ahead of new helper gate." }
  adjacent_tests: { result: pass, suites_run: ["corevalidate raw call and Phase29/31 ownership roots", "originvalidate raw call and Phase31 ownership roots", "pathoracle raw authority and Phase29/31 ownership and Phase35 ancestry roots", "session Phase24 activation/cleanup and Phase29 fixture roots"] }
  revert_and_reconfirm: { result: pass, bug_returned_on_revert: true, fixed_on_reapply: true, method: "Temporary Go overlay of HEAD corevalidate reproduced all original route failures (2.543s); restored final sources pass exact frozen baseline+Controls (2.489s session selection). Seven one-site reversions additionally fail their driving regressions." }
  guardrail_verdict: accepted
  disposition: resolved under existing routine-fix authorization; session manager independently reran both exact final selectors successfully before archive/commit; no human-confirmation claim
files_changed: [internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_phase36_legacy_internal_test.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/pathoracle/pathoracle.go, internal/compiler/session/session_phase36_legacy_calls_test.go, .planning/debug/resolved/phase36-legacy-call-routes.md, .planning/phases/36-scalar-helper-calls/deferred-items.md, .planning/phases/36-scalar-helper-calls/36-05-SUMMARY.md, .planning/WINDOWS.md]
cycles: 1 investigation + 3 bounded fix rounds (core applicability, peer PathToken gate, origin owner trigger/refusal precedence)

### Final exact commands

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase36LegacyFallibleCallRoutes|TestPhase36LegacyFileOwnerPeerRoutes|TestPhase24ActivationPeerDerivesRepeatedAcquisitions|TestPhase24CleanupPeerRejectsOwnershipAndOrderMutations|TestPhase29FixtureChecks)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^(TestPhase36CallCFGApplicability|TestPhase36LegacyFileOwnerCalls|TestPhase36RawCallShapeAndFreshness|TestPhase36RawCallRefusalRemovalControls|TestPhase31FileBytesSeededFaults|TestPhase29ReleaseBeforeViewUseRefused|TestPhase29MoveOwnerWhileViewLive|TestPhase29OwnerLoopBlockRefused|TestPhase36OriginCallFreshnessAndTransfer|TestPhase36OriginCallRefusalRemovalControls|TestPhase36ScalarCallAuthorityBoundaries|TestPhase35PureOperationAncestry)$' -count=1 -v -timeout=600s
```

The second command intentionally selects the identically named Phase31 seeded
fault root in all three independent validator packages. All three ran and passed.

## Prevention

why_chain:
  - Legacy applications stopped before emission because new scalar-only validators selected any call with a scalar endpoint.
  - A U64 result was mistaken for evidence of a U64/Bool scalar argument/helper; old PathToken and owning-ByteView calls also return U64.
  - New raw helper tests covered four scalar pairs and malformed fields, but did not compose both existing owner wrappers, the optional scalar tail and the fallible PathToken probe before implementation acceptance.
why_not_caught: The Phase36-01 frozen no-call gate did catch the route regression; Plan36-03's focused helper controls did not run the existing Phase24 activation and all three Phase31 seeded-fault consumers. Plan36-05 recorded the failing frozen gate as an ownership-scoped residual rather than erasing it.
recurrence_guard: TestPhase36CallCFGApplicability and TestPhase36LegacyFileOwnerCalls in corevalidate; TestPhase36LegacyFallibleCallRoutes and TestPhase36LegacyFileOwnerPeerRoutes in session. Retain immutable frozen baseline+Controls and the existing Phase24/29/31 independent lifecycle mutants. Seven fix-site overlay mutations demonstrate these new guards are load-bearing.
scope: Local deterministic source/IR/byte-route validation only; no broad native matrix, full suite, hosted run or cross-host claim.

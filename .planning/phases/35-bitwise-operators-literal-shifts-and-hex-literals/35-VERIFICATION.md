---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
verified: 2026-10-15T21:37:03Z
status: passed
score: "38/38 truths verified"
covered_files:
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-03-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-03-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-04-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-04-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-05-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-05-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-06-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-06-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-07-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-07-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-PLAN.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-SUMMARY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-ADMISSION-MUTATION-EVIDENCE.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-CONTEXT.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-MUTATION-EVIDENCE.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-PROHIBITION-ENFORCEMENT.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-PROHIBITION-WIRING.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH-FINAL-RUN.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH-FIX.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH-GATES.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH-REGRESSION.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REFRESH-RUN.json"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REGRESSION.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW-DISPOSITION.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW-FIX.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-SECURITY.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-TRANSITION-CHECK.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-UI-REVIEW.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VALIDATION.md"
  - ".planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md"
  - ".planning/phases/36-scalar-helper-calls/36-REGRESSION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-DISPOSITION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-ACCEPTANCE-EVIDENCE.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REMEDIATION-EVIDENCE.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REVIEW-DISPOSITION.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REVIEW.md"
  - "cmd/schway/main_phase32_routes_test.go"
  - "cmd/schway/main_phase35_test.go"
  - "cmd/schway/main_phase37_file_run_test.go"
  - "docs/agent-surface.md"
  - "examples/bitwise.schway"
  - "examples/checksum/adler32.schway"
  - "internal/compiler/cgen/cgen_phase35_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/check/check.go"
  - "internal/compiler/check/check_phase35_baseline_test.go"
  - "internal/compiler/check/check_phase35_numeric_test.go"
  - "internal/compiler/check/check_phase35_test.go"
  - "internal/compiler/check/check_phase36_scalar_calls_test.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/core/core.go"
  - "internal/compiler/core/core_phase35_test.go"
  - "internal/compiler/core/core_test.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/corevalidate/corevalidate_phase35_test.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase35_test.go"
  - "internal/compiler/interp/filebytes_reader.go"
  - "internal/compiler/interp/interp.go"
  - "internal/compiler/interp/interp_phase35_test.go"
  - "internal/compiler/interp/interp_phase37_test.go"
  - "internal/compiler/originvalidate/originvalidate.go"
  - "internal/compiler/originvalidate/originvalidate_phase35_test.go"
  - "internal/compiler/pathoracle/pathoracle.go"
  - "internal/compiler/pathoracle/pathoracle_phase35_test.go"
  - "internal/compiler/session/session_phase272_identity_test.go"
  - "internal/compiler/session/session_phase273_identity_test.go"
  - "internal/compiler/session/session_phase34_compat_test.go"
  - "internal/compiler/session/session_phase34_limit_consumers_test.go"
  - "internal/compiler/session/session_phase34_prechange_test.go"
  - "internal/compiler/session/session_phase34_presentation_test.go"
  - "internal/compiler/session/session_phase34_surface_test.go"
  - "internal/compiler/session/session_phase35_adler_test.go"
  - "internal/compiler/session/session_phase35_admission_mutation_test.go"
  - "internal/compiler/session/session_phase35_closeout_test.go"
  - "internal/compiler/session/session_phase35_peer_mutation_test.go"
  - "internal/compiler/session/session_phase35_prechange_test.go"
  - "internal/compiler/session/session_phase35_surface_test.go"
  - "internal/compiler/session/session_phase35_values_test.go"
  - "internal/compiler/session/session_phase36_prechange_test.go"
  - "internal/compiler/surface/refusals.go"
  - "internal/compiler/surfacemanifest/manifest.go"
  - "internal/compiler/surfacemanifest/manifest_test.go"
  - "internal/compiler/syntax/format.go"
  - "internal/compiler/syntax/numeric_phase35_test.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase34_recovery_test.go"
  - "internal/compiler/syntax/parser_phase34_refusal_test.go"
  - "internal/compiler/syntax/parser_phase35_baseline_test.go"
  - "internal/compiler/syntax/parser_phase35_test.go"
  - "internal/compiler/syntax/parser_phase36_test.go"
  - "testdata/phase16/public-emitter-consumers.json"
  - "testdata/phase35/numeric-delta.json"
  - "testdata/phase35/numeric-policy-baseline.json"
  - "testdata/phase35/prechange-corpus.json"
  - "testdata/phase35/prohibition-causal-bad.txt"
  - "testdata/phase35/prohibition-evidence.test.cjs"
  - "testdata/phase35/prohibition-lanes-bad.txt"
  - "testdata/phase35/recovery-after.json"
  - "testdata/phase35/shift-sites.json"
  - "testdata/phase35/surface-census.json"
  - "testdata/phase35/surface-witnesses.json"
  - "testdata/phase35/value-vectors.json"
  - "testdata/phase36/new-fixture-ledger.json"
  - "testdata/phase37/new-fixture-ledger.json"
covered_digest: "v3:sha256:787e2bb5d94b4ee7d8c706b041d32b137724dece6ab45c88e4652ad2148b9f1c"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: human_needed
  previous_score: 38/38
  gaps_closed:
    - "BIT-01 test-tier prohibition is enforced with a known-bad causal-evidence fixture that fails and clean evidence that passes."
    - "BIT-02 test-tier prohibition is enforced with a known-bad freshness claim fixture that fails and clean evidence that passes."
  gaps_remaining: []
  regressions: []
  scope: "Phase37 test-tier prohibition wiring and current-head evidence refresh; prior human_needed finding closed by explicit plan descriptors and fail-first enforcement, not an override."
---

# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals Verification Report

**Phase Goal:** Ordinary Schway scalar code can use U64 ^, &, |, literal-amount << and >>, and hex literals, with one defined meaning on every engine and optimization level, no C17 shift undefined behavior, and no new defect path.

**Verified:** 2026-10-15T21:37:03Z
**Status:** passed
**Re-verification:** Yes — after the test-tier prohibition wiring closure; the initial Phase 35, Phase 36 refresh, and pre-wiring Phase 37 refresh remain preserved below.

## Goal Achievement

Current source inspection confirms the complete route from scalar parsing and checking through independent raw-core validators to interpreter and C execution. Shift operations carry canonical immediate counts 0..63; the C emitter uses unsigned widened shift operands and literal counts; surface and public witness paths are connected. Independent expected values and wrong-answer controls establish numeric outcomes, while executionpeer coverage is explicitly limited to event-free causal admission.

I derived 38 observable truths from the five roadmap criteria plus the plan truths (34 plan-specific, with one duplicate of roadmap criterion R4). All eight artifact queries pass: 48 declarations across 45 unique paths. The plans declare 24 key links; the generic checker recognizes one lexically, and source/test/fixture tracing confirms all 24. The requirement union is BIT-01 through BIT-04, each mapped and satisfied; no Phase35 requirement is orphaned. The context decision check reports 23/23 honored.

At the prior Phase37 refresh, two plan prohibitions remained an escalation item because the plan had no explicit verification tier. The dated amendment to 35-08-PLAN.md now declares both as resolved test-tier checks. The canonical prohibition-enforcement path parsed those descriptors, ran their violation fixtures and clean controls, and returned green for both; the previous human_needed item is closed by that mechanical evidence, not by an override.

I retained the earlier goal-backward inventory: five roadmap criteria, 34 plan-specific truths with one duplicate, 48 artifact declarations over 45 unique paths, 24 declared links, the BIT-01/02/03/04 requirement set, and 23 context decisions/prohibitions. The two prohibition checks now have concrete test-tier descriptors and executable negative controls. All 38 unique truths remain verified; no plan-mapped requirement is orphaned.

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| R1 | U64 ^, & and \| have literal expected values in the interpreter and native code at O0, O3 and O3-LTO; checker/corevalidate/originvalidate independently derive transfers; independent value controls and per-op transfer mutants catch wrong answers; executionpeer checks causal admission and forged events only. | ✓ VERIFIED | TestPhase35BitwiseOptimizationMatrix, TestPhase35EngineWrongAnswerControls, fresh TestPhase35PeerTransferMutationControls, TestPhase35PureOperationCausalControls. Executionpeer is not counted as numeric evidence. |
| R2 | Literal << and >> counts 0, 31, 32 and 63 are defined on both engines at each optimization level, left shift drops high bits, invalid or nonliteral counts are refused at check/core/origin, and generated C has no invalid shift amount. | ✓ VERIFIED | TestPhase35BitwiseOptimizationMatrix, TestPhase35ShiftUBSan, TestPhase35ShiftAdmission, TestPhase35ShiftImmediateAdmission, TestPhase35UnsignedShiftEmissionAndEventProjection and TestPhase35ShiftSiteControls; fresh admission mutation and target-freshness controls. |
| R3 | Hex U64 literals accept case-insensitive hex digits on the scalar route, preserve source spelling, and reject malformed or overflowing forms with named diagnostics. | ✓ VERIFIED | TestPhase35NumericOwnership, TestPhase35NumericRoutePolicy, TestPhase35ScalarGrammar, TestPhase35FormatterRoundTrip and TestPhase35OperatorDocumentation. |
| R4 | Adler uses b << 16 and all pinned Adler/checksum/adler_low results remain unchanged on both engines, including the existing Phase29 and Phase31 fixtures. | ✓ VERIFIED | TestPhase35ExistingAdlerAnswers, TestPhase35AdlerWrongShiftControl, TestPhase29InterpreterFixtureVectors, TestPhase29NativeFixtureVectors and TestPhase31CaseMatrix. |
| R5 | Programs without the new operators/literals retain byte-identical generated C, and the Phase34 surface manifest meta-test still passes with the new admitted operators and refusals registered. | ✓ VERIFIED | TestPhase35PreChangeEmitterCorpus, TestPhase35PreChangeCorpusControls, TestPhase34SurfaceManifestControls and TestPhase35SurfaceManifestControls. Baseline corpus predates the first feature commit. |
| P01-T1 | Before compiler edits, a revision-pinned old-C/core/identity corpus covered tracked source, all three emitter entry points and explicit refusals. | ✓ VERIFIED | TestPhase35PreChangeEmitterCorpus and TestPhase35PreChangeCorpusControls compare the frozen corpus; the recorded prechange revision precedes the first feature commit. |
| P01-T2 | Legacy numeric policy remains: scalar 010 is 8, 0B10 is 2, legacy 010 is 10, legacy 0B10 is refused, and Bool identifiers remain legal. | ✓ VERIFIED | TestPhase35NumericPolicyBaseline, TestPhase35NumericRoutePolicy, TestPhase34BoolSpellingsRemainIdentifiers, and Phase19 literal admission/range/refusal selectors. |
| P01-T3 | Previous nested-index admission is recorded separately from the new depth-64 limit, without changing old successful goldens. | ✓ VERIFIED | TestPhase35NestedIndexBaseline, TestPhase35ExpressionDepth, TestPhase34RecoveryCorpus and TestPhase34RecoveryControls. |
| P02-T1 | x ^ (y << 3) & 0xFF parses as x ^ ((y << 3) & 0xFF), preserving the literal-number leaf through grouping. | ✓ VERIFIED | TestPhase35ScalarGrammar, TestPhase35FormatterRoundTrip and TestFormatLetBinaryExpressions; the parser/formatter source preserves the transparent grouped literal. |
| P02-T2 | Depth 64 is accepted and 65 refused across groups/indices; shift token adjacency, formatter boundaries and sibling recovery remain defined. | ✓ VERIFIED | TestPhase35ExpressionDepth, TestPhase34RecoveryCorpus, TestPhase34RecoveryControls, TestPhase35FormatterPreservesSeparatedShiftTokens and TestPhase35GroupedNegativeShiftAmountRefusal. |
| P02-T3 | Uppercase legacy hex is admitted narrowly, original spelling is preserved, malformed ownership is full-token and overflow diagnostics remain route-specific. | ✓ VERIFIED | TestPhase35NumericOwnership, TestPhase35NumericRoutePolicy, TestPhase35ScalarGrammar and TestPhase35FormatterRoundTrip, with Phase19 numeric regression selectors. |
| P03-T1 | Five pure exact-width operations lower to core; shifts carry only canonical decimal ConstU64 counts 0..63 and no right operand ID. | ✓ VERIFIED | TestPhase35CoreShape, TestAllOperationKindsRegistered, TestPhase35ShiftAdmission and TestPhase35ShiftImmediateAdmission. |
| P03-T2 | Non-number shift amounts are refused even when constant-known, with lexical and overflow ownership retained. | ✓ VERIFIED | TestPhase35ShiftAdmission and TestPhase35NumericRoutePolicy exercise source refusal and route-specific diagnostics. |
| P03-T3 | Exactly one unused Byte formal is supported for U64/Bool local work; Byte use/read/copy/conversion/operand remains refused, with no arity or U64 input-required widening. | ✓ VERIFIED | TestPhase35UnusedByteSourceAdmission, TestPhase34BoolSpellingsRemainIdentifiers, TestPhase34CallArityBoundaries, TestPhase321RunInputRequiredCLI and TestPhase35PublicConformanceWitness. |
| P03-T4 | Checker transfer faults for all five operations are caught through independent scalar facts and safety sinks. | ✓ VERIFIED | Fresh TestPhase35PeerTransferMutationControls caught 15 actual implementation mutants across check, corevalidate and originvalidate. |
| P04-T1 | Both raw-IR peers independently reject invalid shift shape, type and range without claiming source provenance. | ✓ VERIFIED | TestPhase35ShiftImmediateAdmission in both validator packages, plus the fresh 26-control admission mutation harness. |
| P04-T2 | Check, corevalidate and originvalidate independently derive all five value transfers; corevalidate and origin transfer faults are reached through their real analysis paths. | ✓ VERIFIED | Fresh TestPhase35PeerTransferMutationControls; TestPhase35ScalarTransferMutants and TestPhase35AnalysisTriggers. |
| P04-T3 | Each raw-IR peer accepts only an unused Byte formal with U64/Bool locals and independently rejects Byte use. | ✓ VERIFIED | TestPhase35UnusedByteIRAdmission in both validator packages and TestPhase35AdmissionRefusalMutationControls. Whole-body formal reference checking and peer-specific refusals are exercised. |
| P04-T4 | Old core/ID/schema bytes remain stable; wrong-kind ConstU64 and shift RightID are rejected. | ✓ VERIFIED | TestPhase35PreChangeEmitterCorpus, TestPhase35PreChangeCorpusControls, TestPhase35ShiftImmediateAdmission and TestPhase35AdmissionRefusalMutationControls. |
| P05-T1 | Interpreter computes independently pinned exact-width values for all five operations without a new event, defect or reason. | ✓ VERIFIED | TestPhase35BitwiseShiftVectorsEventFree, TestPhase35InterpreterWrongAnswerControls and TestPhase35BitwiseOptimizationMatrix compare against authored literals. |
| P05-T2 | Byte fixture input remains Byte at the unused bridge and produces U64 locals without conversion. | ✓ VERIFIED | TestPhase35UnusedByteRuntime, TestPhase35PublicConformanceWitness and TestPhase35NativeShiftTracer inspect and run the existing Byte-input fixture. |
| P05-T3 | Executionpeer admits event-free operations and rejects fabricated per-op step/defect events, without numeric replay. | ✓ VERIFIED | TestPhase35PureOperationCausalControls checks the event-free and fabricated-event cases for each op; it is reported as causal evidence only. |
| P05-T4 | Pathoracle recognizes all five value operations and preserves loan-free ancestry. | ✓ VERIFIED | TestPhase35PureOperationAncestry exercises the new kinds through the ancestry dispatch. |
| P06-T1 | C shifts use literal counts below 64, unsigned non-promoted operands and exact-width result casts; an explicit positive site count prevents vacuous scanning. | ✓ VERIFIED | TestPhase35UnsignedShiftEmissionAndEventProjection and TestPhase35ShiftSiteControls; both expected sites are found and generated C/native boundary vectors pass. |
| P06-T2 | Event projection, schema size/capacity and scalar selectors recognize all five operations without adding events or optimizer facts. | ✓ VERIFIED | TestPhase35UnsignedShiftEmissionAndEventProjection, TestPhase35NativeShiftTracer and TestAllOperationKindsHandledAtEverySite. |
| P06-T3 | The unused-Byte conformance signature retains Byte input/U64 result and U64/Bool locals, without defaults or Byte conversion. | ✓ VERIFIED | TestPhase35NativeShiftTracer, TestPhase35PublicConformanceWitness and TestPhase35UnusedByteRuntime. |
| P06-T4 | Old generated C and core/schema identifiers remain byte-identical apart from declared feature/resource deltas. | ✓ VERIFIED | TestPhase35PreChangeEmitterCorpus and TestPhase35PreChangeCorpusControls. The only declared example source rewrite is the Adler compression; emitted baseline fixtures remain frozen. |
| P07-T1 | Public run conformance yields value 264 for x=0x100,y=0x21; alternate final-mask association yields 8; application routes separately print 264 plus newline. | ✓ VERIFIED | TestPhase35PublicConformanceWitness and TestPhase35FocusedValueAnswers compare outcome values and application stdout on interpreter/native routes. |
| P07-T2 | Literal vectors pass O0/O3/O3-LTO with per-op wrong-answer controls and focused shift UBSan; agreement alone is not treated as correctness. | ✓ VERIFIED | TestPhase35BitwiseOptimizationMatrix ran 33 authored rows in each native optimization lane; TestPhase35ShiftUBSan ran all 16 shift rows; TestPhase35EngineWrongAnswerControls caught five operator controls. |
| P07-T3 | Versioned schway.surface/1 keeps ten top-level fields and source-free discovery while documenting operators, precedence, counts, range, low-64 shift and hex facts. | ✓ VERIFIED | TestPhase35SurfaceManifestWitnesses, TestPhase35SurfaceManifestCensus, TestPhase35SurfaceManifestControls, TestPhase35SurfaceDocument and TestPhase35OperatorDocumentation. |
| P07-T4 | Finite admission, limits and census plus removal/false-admission/missing-refusal/hint/weakened-restriction controls remain reached. | ✓ VERIFIED | TestPhase35SurfaceManifestCensus, TestPhase35SurfaceManifestControls, TestPhase35SurfaceLimitConsumers and fresh TestPhase35AdmissionRefusalMutationControls. |
| P07-T5 | Operator adjacency, empty/missing operand recovery, equal operands and left association are pinned. | ✓ VERIFIED | TestPhase35ScalarGrammar, TestPhase34RecoveryCorpus, TestPhase34RecoveryControls, TestPhase35GroupedNegativeShiftAmountRefusal and TestPhase35FormatterPreservesSeparatedShiftTokens. |
| P07-T6 | Negative and 64 counts refuse, 0 and 63 succeed; grouped number literals remain literal, computed/empty/missing values refuse, and low-64/logical shifts are exact. | ✓ VERIFIED | TestPhase35ShiftAdmission, TestPhase35ShiftImmediateAdmission, TestPhase35BitwiseOptimizationMatrix, TestPhase35ShiftUBSan and TestPhase35GroupedNegativeShiftAmountRefusal. |
| P07-T7 | Maximum U64 and bounded leading-zero hex succeed; max+1 refuses without rounding, signed conversion or digit-count cutoff. | ✓ VERIFIED | TestPhase35NumericOwnership, TestPhase35NumericRoutePolicy, TestPhase35BitwiseOptimizationMatrix and TestPhase35OperatorDocumentation. |
| P08-T1 | Adler uses b<<16 while existing Phase29/31 checksum/adler_low outputs remain unchanged on both engines. | ✓ VERIFIED (deduplicated under R4) | Same roadmap truth as R4; TestPhase35ExistingAdlerAnswers, TestPhase29InterpreterFixtureVectors, TestPhase29NativeFixtureVectors and TestPhase31CaseMatrix provide the evidence. |
| P08-T2 | Empty, one-byte and boundary Adler cases preserve results, imports and cleanup; wrong-shift control reaches the comparator. | ✓ VERIFIED | TestPhase35ExistingAdlerAnswers and TestPhase35AdlerWrongShiftControl; the shift-15 copied control reported 150143896 against independently pinned 300286872. |
| P08-T3 | Summary coverage points to actual tests and closed coverage kinds; evidence distinguishes fresh runs, source inspection, historical receipts and incomplete hosts. | ✓ VERIFIED | TestPhase35EvidenceGrounding and TestSummaryCoverageGroundedness; both fresh selectors passed, and this report separates evidence classes. |
| P08-T4 | Dormant seeds z1g/trz/z1s retain named source witnesses, consumers and triggers; living documents update only inspected facts. | ✓ VERIFIED | 35-CONTEXT.md, 35-RESEARCH.md, 35-TRANSITION-CHECK.md and the checked .planning/PRODUCT-ROADMAP.md/.planning/LANGUAGE-MATURITY.md entries record the dormant dispositions and three scoped next capability owners. |
| P08-T5 | Adler and resource deltas are explicit while old C/core/ID/schema/numeric baselines remain immutable. | ✓ VERIFIED | 35-COMPATIBILITY-DELTA.md plus TestPhase35PreChangeEmitterCorpus, TestPhase35PreChangeCorpusControls, TestPhase35EvidenceGrounding and TestSummaryCoverageGroundedness. |

**Score:** 38/38 unique truths verified (0 present, behavior-unverified). All 39 declared entries are accounted for; P08-T1 duplicates roadmap truth R4 and is counted once.

### Declared Prohibitions

| Prohibition | Status | Evidence |
|---|---|---|
| Do not present engine agreement or causal event admission as independent proof of pure numeric results. | ✓ VERIFIED | The plan declares a test-tier node-test descriptor. The known-bad causal-evidence fixture fails at the expected scope assertion; the clean evidence and default subject pass. Canonical GSD enforcement reports failFirst=true, passed=true, flagged=false. Numeric evidence remains separately grounded in literal vectors, wrong-answer controls and peer transfer mutants. |
| Do not describe historical hosted receipts or skipped optimization/sanitizer lanes as fresh Phase35 evidence. | ✓ VERIFIED | The plan declares a test-tier node-test descriptor. Its known-bad freshness-claim fixture fails at the expected scope assertion; clean evidence and default subject pass. Canonical enforcement reports failFirst=true, passed=true, flagged=false. The current final receipt is distinguished from historical receipts below. |

## Required Artifacts

All 48 artifact declarations across the eight plans passed the artifact verifier's existence and substantive-content checks. Implementation sources were compiled and exercised by the named package tests; fixtures flow into their consumer tests; documentation and evidence ledgers are checked by closeout selectors or inspected with their source cross-references. Duplicate paths appear under each plan that declares them.

| Plan | Artifact | Expected | Status | Details |
|---|---|---|---|---|
| 35-01 | internal/compiler/session/session_phase35_prechange_test.go | Frozen baseline evidence | ✓ VERIFIED | Executed in prechange corpus and controls selectors. |
| 35-01 | testdata/phase35/prechange-corpus.json | Frozen old emitter/core/identity outcomes | ✓ VERIFIED | Read by the baseline corpus comparator. |
| 35-01 | testdata/phase35/numeric-policy-baseline.json | Old numeric route policy | ✓ VERIFIED | Read by numeric policy baseline and route-policy tests. |
| 35-02 | internal/compiler/syntax/parser.go | Scalar grammar and precedence | ✓ VERIFIED | Compiled and exercised by grammar, grouping, shift and depth tests. |
| 35-02 | internal/compiler/syntax/format.go | Lossless/source-spelling format behavior | ✓ VERIFIED | Round-trip and separated-shift formatter tests pass. |
| 35-02 | internal/compiler/surface/refusals.go | Named refusal codes and hints | ✓ VERIFIED | Admission/surface tests verify registered refusal vocabulary. |
| 35-02 | internal/compiler/syntax/lexer.go | Numeric candidate/token ownership | ✓ VERIFIED | Numeric ownership and route policy tests pass. |
| 35-02 | internal/compiler/check/check.go | Numeric conversion and checker policy | ✓ VERIFIED | Compiled and used by numeric and shift admission tests. |
| 35-02 | internal/compiler/syntax/parser_phase35_test.go | Grammar, grouping and depth evidence | ✓ VERIFIED | Exact named grammar/depth selectors pass. |
| 35-02 | internal/compiler/syntax/numeric_phase35_test.go | Literal boundaries and spelling evidence | ✓ VERIFIED | Numeric ownership/route tests pass. |
| 35-02 | internal/compiler/check/check_phase35_numeric_test.go | Checker numeric route evidence | ✓ VERIFIED | Numeric route policy selector passes. |
| 35-02 | testdata/phase35/numeric-delta.json | Scoped numeric delta | ✓ VERIFIED | Consumed by the numeric compatibility evidence. |
| 35-03 | internal/compiler/core/core.go | Five operation kinds and canonical immediates | ✓ VERIFIED | Core shape and exhaustive operation registration selectors pass. |
| 35-03 | internal/compiler/check/scalar.go | Source transfer and immediate lowering | ✓ VERIFIED | Shift/source admission and fresh transfer mutants exercise it. |
| 35-03 | internal/compiler/check/check.go | Scalar checker integration | ✓ VERIFIED | Compiled and used by source admission and public witness tests. |
| 35-03 | internal/compiler/core/core_phase35_test.go | Producer/core contract tests | ✓ VERIFIED | Core shape, registration and shift admission selectors pass. |
| 35-03 | internal/compiler/check/check_phase35_test.go | Source admission and transfer tests | ✓ VERIFIED | Shift, unused-Byte and scalar transfer selectors pass. |
| 35-04 | internal/compiler/corevalidate/corevalidate.go | Independent raw-core validation | ✓ VERIFIED | Raw-IR admission, analysis triggers, freshness and mutation tests pass. |
| 35-04 | internal/compiler/originvalidate/originvalidate.go | Independent origin validation | ✓ VERIFIED | Raw-IR admission, analysis triggers, freshness and mutation tests pass. |
| 35-04 | internal/compiler/corevalidate/corevalidate_phase35_test.go | Core validator controls | ✓ VERIFIED | Shift/Byte admission and producer-shape selectors pass. |
| 35-04 | internal/compiler/originvalidate/originvalidate_phase35_test.go | Origin validator controls | ✓ VERIFIED | Shift/Byte admission and producer-shape selectors pass. |
| 35-05 | internal/compiler/interp/interp.go | Runtime value semantics | ✓ VERIFIED | Literal vectors and wrong-answer controls exercise interpretation. |
| 35-05 | internal/compiler/pathoracle/pathoracle.go | Operation ancestry classification | ✓ VERIFIED | Pure-operation ancestry selector passes. |
| 35-05 | internal/compiler/executionpeer/executionpeer.go | Causal event-free validation | ✓ VERIFIED | Per-op admission and forged-event controls pass. |
| 35-05 | internal/compiler/interp/interp_phase35_test.go | Runtime/vector controls | ✓ VERIFIED | Exact value, no-event, wrong-answer and Byte runtime selectors pass. |
| 35-05 | internal/compiler/pathoracle/pathoracle_phase35_test.go | Ancestry tests | ✓ VERIFIED | New kinds resolve through the ancestry dispatch. |
| 35-05 | internal/compiler/executionpeer/executionpeer_phase35_test.go | Causal controls | ✓ VERIFIED | Per-op causal tests pass; not used as numeric evidence. |
| 35-05 | examples/bitwise.schway | Public scalar/unused-Byte witness | ✓ VERIFIED | Consumed by the public conformance test and reaches both engines. |
| 35-05 | internal/compiler/core/core_test.go | Exhaustive fixture registration | ✓ VERIFIED | TestAllOperationKindsHandledAtEverySite passes. |
| 35-05 | testdata/phase35/value-vectors.json | Independent literal expected values | ✓ VERIFIED | Read by interpreter, optimization, UBSan and wrong-answer controls. |
| 35-06 | internal/compiler/cgen/cgen.go | Existing emitter entry points | ✓ VERIFIED | Frozen corpus and native tracer exercise Emit, EmitNative and EmitApplication. |
| 35-06 | internal/compiler/cgen/cgen_program.go | Scalar C generation and event projection | ✓ VERIFIED | C site, event projection and native answer selectors pass. |
| 35-06 | internal/compiler/cgen/cgen_phase35_test.go | C shift site evidence | ✓ VERIFIED | Native tracer, admission gate, emission and site-control selectors pass. |
| 35-06 | testdata/phase35/shift-sites.json | Positive C shift site inventory | ✓ VERIFIED | Static scan finds both required sites; removal controls are reached. |
| 35-07 | internal/compiler/surfacemanifest/manifest.go | Versioned surface manifest | ✓ VERIFIED | Manifest witness/census/control tests and Phase34 meta-tests pass. |
| 35-07 | examples/bitwise.schway | Runnable bitwise witness | ✓ VERIFIED | Public conformance route produces the expected outcome. |
| 35-07 | cmd/schway/main_phase35_test.go | Runnable CLI witness | ✓ VERIFIED | TestPhase35PublicConformanceWitness passes. |
| 35-07 | internal/compiler/session/session_phase35_values_test.go | Native vector, optimization and UBSan checks | ✓ VERIFIED | 33 vector rows and all 16 shift UBSan rows pass without skips. |
| 35-07 | testdata/phase35/value-vectors.json | Independent vector fixture | ✓ VERIFIED | Consumer tests compare each engine against decimal literals. |
| 35-07 | internal/compiler/session/session_phase35_surface_test.go | Surface manifest acceptance/controls | ✓ VERIFIED | Witness, census, controls and limit-consumer selectors pass. |
| 35-07 | testdata/phase35/surface-witnesses.json | Public surface witness values | ✓ VERIFIED | Read by manifest witness checks. |
| 35-07 | testdata/phase35/surface-census.json | Finite public surface census | ✓ VERIFIED | Read by manifest census and mutation controls. |
| 35-08 | examples/checksum/adler32.schway | Adler consumer using b<<16 | ✓ VERIFIED | Imported by fresh Adler answer tests; old expected results pass. |
| 35-08 | docs/agent-surface.md | User-facing operator semantics | ✓ VERIFIED | Operator documentation and public document value selectors pass. |
| 35-08 | internal/compiler/session/session_phase35_adler_test.go | Adler answers and wrong-shift control | ✓ VERIFIED | Both exact named selectors pass. |
| 35-08 | internal/compiler/session/session_phase35_closeout_test.go | Evidence/document grounding | ✓ VERIFIED | Evidence grounding and summary coverage selectors pass. |
| 35-08 | .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md | Scoped compatibility ledger | ✓ VERIFIED | Cross-checked against source and frozen baselines. |
| 35-08 | .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md | Fresh/history/incomplete evidence ledger | ✓ VERIFIED | Named receipts resolve to actual tests; freshness controls pass. |
| 35-08 acceptance | testdata/phase35/prohibition-evidence.test.cjs | Existing Go closeout validator exposed through a test-tier TAP adapter | ✓ VERIFIED | Adapter calls TestPhase35EvidenceGrounding and asserts exact RUN/PASS with no fail or skip; it contains no duplicate evidence rules. |
| 35-08 acceptance | testdata/phase35/prohibition-causal-bad.txt | Known-bad BIT-01 false-evidence control | ✓ VERIFIED | One false causal/numeric claim is rejected by the existing Go evidence validator. |
| 35-08 acceptance | testdata/phase35/prohibition-lanes-bad.txt | Known-bad BIT-02 freshness-claim control | ✓ VERIFIED | One false optimization/sanitizer freshness claim is rejected by the existing Go evidence validator. |
| 35-08 acceptance | 35-PROHIBITION-ENFORCEMENT.json | Canonical test-tier enforcement outcomes | ✓ VERIFIED | Both projected descriptors pass fail-first, violation-fixture, clean-control and default-subject checks. |
| 35-08 acceptance | 35-REFRESH-FINAL-RUN.json | Current focused regression receipt | ✓ VERIFIED | 122 current input hashes and 16 raw-log hashes match; 175 root executions are RUN/PASS with zero missing, failed or skipped roots. |

## Key Link Verification

The generic key-link query recognized only 1 of 24 relationships because most are expressed through domain-specific tests or fixtures rather than a literal import pattern. I traced the source/fixture path through the consumer tests for all 24 links; none remains unconnected.

| From | To | Via | Status | Details |
|---|---|---|---|---|
| session_phase35_prechange_test.go | cgen.go | Frozen source inventory and Emit/EmitNative/EmitApplication outcomes | ✓ WIRED | TestPhase35PreChangeEmitterCorpus and TestPhase35PreChangeCorpusControls. |
| check_phase35_baseline_test.go | check/scalar.go | Numeric policy and nested-index baseline | ✓ WIRED | TestPhase35NumericPolicyBaseline and TestPhase35NestedIndexBaseline. |
| syntax/parser.go | syntax/format.go | Lossless operator tokens to AST and formatter reparse | ✓ WIRED | TestPhase35ScalarGrammar, TestPhase35FormatterRoundTrip and separated-token repair selector. |
| syntax/lexer.go | check/check.go | Raw numeric token/span to route-local conversion | ✓ WIRED | TestPhase35NumericOwnership and TestPhase35NumericRoutePolicy. |
| syntax/parser.go | surface/refusals.go | Depth/invalid-shift recovery to named refusal | ✓ WIRED | TestPhase35ExpressionDepth, TestPhase35GroupedNegativeShiftAmountRefusal and surface manifest tests. |
| syntax/parser.go | check/scalar.go | Transparent numeric leaf to literal-only shift admission | ✓ WIRED | TestPhase35ShiftAdmission and TestPhase35NumericOwnership. |
| check/scalar.go | core/core.go | Five operation registrations and canonical immediate | ✓ WIRED | TestPhase35CoreShape, TestAllOperationKindsRegistered and TestPhase35ShiftAdmission. |
| check/scalar.go | check/scalar.go | Unused Byte formal refusal at source | ✓ WIRED | TestPhase35UnusedByteSourceAdmission and TestPhase35PublicConformanceWitness. |
| core/core.go | corevalidate/corevalidate.go | Raw immediate laws and independent value analysis | ✓ WIRED | TestPhase35ShiftImmediateAdmission, TestPhase35AnalysisTriggers and fresh peer mutations. |
| core/core.go | originvalidate/originvalidate.go | Raw immediate laws and independent origin analysis | ✓ WIRED | TestPhase35ShiftImmediateAdmission, TestPhase35AnalysisTriggers and fresh peer mutations. |
| originvalidate/originvalidate.go | originvalidate/originvalidate.go | Independent transfer fault to named safety sink | ✓ WIRED | TestPhase35ScalarTransferMutants and fresh TestPhase35PeerTransferMutationControls. |
| corevalidate/corevalidate.go | interp/interp.go | Validated core to runtime values/events | ✓ WIRED | TestPhase35BitwiseShiftVectorsEventFree and TestPhase35InterpreterWrongAnswerControls. |
| core/core.go | pathoracle/pathoracle.go | New operation ancestry dispatch | ✓ WIRED | TestPhase35PureOperationAncestry. |
| interp/interp.go | executionpeer/executionpeer.go | Actual execution document to causal classification | ✓ WIRED | TestPhase35PureOperationCausalControls; this proves causal admission only. |
| value-vectors.json | interp_phase35_test.go | Decimal literals to interpreter comparator | ✓ WIRED | TestPhase35BitwiseShiftVectorsEventFree and TestPhase35InterpreterWrongAnswerControls. |
| core/core.go | cgen_program.go | Validated immediate through emitter to unsigned C | ✓ WIRED | TestPhase35UnsignedShiftEmissionAndEventProjection and TestPhase35ShiftSiteControls. |
| cgen_program.go | native.go | C output and event projection to native execution | ✓ WIRED | TestPhase35BitwiseOptimizationMatrix, TestPhase35ShiftUBSan and TestPhase35EngineWrongAnswerControls. |
| core/core.go | cgen_program.go | Existing unused Byte signature through native fixture | ✓ WIRED | TestPhase35NativeShiftTracer and TestPhase35PublicConformanceWitness. |
| examples/bitwise.schway | main_phase35_test.go | Public run on the checked Byte fixture | ✓ WIRED | TestPhase35PublicConformanceWitness returns value 264. |
| value-vectors.json | session_phase35_values_test.go | Literal vectors through O0/O3/O3-LTO/UBSan | ✓ WIRED | TestPhase35BitwiseOptimizationMatrix and TestPhase35ShiftUBSan. |
| manifest.go | session_phase35_surface_test.go | Complete source/census and copied drift controls | ✓ WIRED | TestPhase35SurfaceManifestWitnesses, TestPhase35SurfaceManifestCensus and TestPhase35SurfaceManifestControls. |
| examples/checksum/adler32.schway | session_phase35_adler_test.go | b<<16 through imported checksum/adler_low consumers | ✓ WIRED | TestPhase35ExistingAdlerAnswers and TestPhase35AdlerWrongShiftControl. |
| 35-EVIDENCE.md | session_phase35_closeout_test.go | Actual test refs and closed coverage vocabulary | ✓ WIRED | TestPhase35EvidenceGrounding and TestSummaryCoverageGroundedness. |
| 35-EVIDENCE.md | .planning/LANGUAGE-MATURITY.md | Source facts/receipts and scoped capability owners | ✓ WIRED | Cross-reference inspection plus TestPhase35EvidenceGrounding. |
| 35-08-PLAN.md | prohibition-evidence.test.cjs | Canonical descriptor projection to fail-first test-tier enforcement | ✓ WIRED | parseMustHavesBlock and descriptorFromProjection produce the exact requests; each check runs its violation fixture, clean fixture, and default subject. |

## Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| examples/bitwise.schway | x, y, witness | Program source flows through CLI, checking, interpreter/native runners | Yes; returned outcome 264 and application output 264 newline are compared to literal expectations | ✓ FLOWING |
| testdata/phase35/value-vectors.json | Operation, operands, expected decimal | Independently authored fixture read by value, optimization and sanitizer tests | Yes; each vector is executed by interpreter and generated native C | ✓ FLOWING |
| examples/checksum/adler32.schway | Adler result and imported checksum/adler_low values | Imported Schway modules run through existing interpreter/native consumers | Yes; pinned literal outputs and cleanup outcomes are compared | ✓ FLOWING |
| testdata/phase35/surface-witnesses.json and surface-census.json | Expected operator/manifest rows | Independent fixture files read by session surface tests against manifest generation | Yes; generated manifest is compared with witness/census, including copied drift controls | ✓ FLOWING |
| testdata/phase35/prechange-corpus.json | Frozen emitter output/refusal outcomes | Revision-pinned baseline consumed by session corpus test | Yes; current emitter outputs are compared byte-for-byte to old expected data | ✓ FLOWING |

## Behavioral Spot-Checks

The 35-REFRESH-RUN and 35-REFRESH-REGRESSION receipts are preserved as history, but are not treated as current input bindings because the closeout test changed after those runs. The current Phase37 follow-up relies on 35-REFRESH-FINAL-RUN.json and the separate prohibition-enforcement evidence. No full local package/project suite, broad native matrix, hosted run, push, or UAT replay was run.

### Current final regression receipt

| Receipt | Bound inputs and logs | Result |
|---|---|---|
| 35-REFRESH-FINAL-RUN.json | 122 current input hashes; 16 raw-log hashes; 175 package-root executions, 165 distinct root test names across 16 packages. All roots are RUN/PASS; no missing, failed or skipped roots. | ✓ PASS |

I independently checked that all 122 input hashes and 16 raw-log hashes match current bytes. The receipt source_head is c1136586224e06d300b0b96b123db8af05a4be9b; current HEAD is d9f08d495476eb2fb2e522f84ef2c7150c85afd6, and the intervening tracked changes are limited to planning evidence documents. The receipt records all 15 arithmetic and 26 admission source mutants caught. Its feature-only native evidence covers 33 independently authored vectors at O0/O3/O3-LTO and all 16 shift UBSan vectors; every lane passed without skips. The exact multiset includes 175 root executions; repeated test names in separate package roots are preserved by the receipt.

### Test-tier prohibition enforcement

I projected both 35-08-PLAN.md prohibition descriptors through the canonical parseMustHavesBlock and descriptorFromProjection helpers and compared the resulting requests with the saved requests. They matched exactly. Each canonical check returned green, with located=true, flagged=false, failFirst=true, passed=true, and failFirstProof=violation-fixture.

~~~sh
source /tmp/schway34-gsd-run.sh
gsd_run check prohibition-enforcement /tmp/schway35-prohibition-0.json
gsd_run check prohibition-enforcement /tmp/schway35-prohibition-1.json
~~~

The node TAP adapter invokes the existing Go TestPhase35EvidenceGrounding test; it adds no evidence-validation rules. Each known-bad fixture fails the Go test at its intended false-claim/scope assertion. The unchanged clean evidence and the default subject each return exit 0 with one passing test and no skip. Thus each descriptor is fail-first against its own violation fixture, while the ordinary accepted subject stays green.

~~~sh
GOCACHE=/tmp/ai-lang-verification-gocache GSD_PROHIB_SUBJECT=testdata/phase35/prohibition-causal-bad.txt node --test --test-reporter=tap -- testdata/phase35/prohibition-evidence.test.cjs
GOCACHE=/tmp/ai-lang-verification-gocache GSD_PROHIB_SUBJECT=testdata/phase35/prohibition-lanes-bad.txt node --test --test-reporter=tap -- testdata/phase35/prohibition-evidence.test.cjs
GOCACHE=/tmp/ai-lang-verification-gocache GSD_PROHIB_SUBJECT=.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md node --test --test-reporter=tap -- testdata/phase35/prohibition-evidence.test.cjs
GOCACHE=/tmp/ai-lang-verification-gocache node --test --test-reporter=tap -- testdata/phase35/prohibition-evidence.test.cjs
~~~

### Earlier bounded implementation checks retained from the 21:17Z refresh

The following named implementation selectors passed during the pre-wiring Phase37 refresh. They are preserved with their original commands as prior checks; the current final regression receipt above supplies the current source/log-bound regression. No compiler or runtime implementation changed in the wiring follow-up.

**Evidence grounding**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35EvidenceGrounding|TestSummaryCoverageGroundedness)$' -count=1 -v -timeout=600s
```

✓ PASS; false-evidence controls and shared summary coverage validator pass.

**Grammar, recovery and literal ownership**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase35ScalarGrammar|TestPhase35ExpressionDepth|TestPhase35FormatterRoundTrip|TestPhase35NumericOwnership|TestPhase35GroupedNegativeShiftAmountRefusal)$' -count=1 -v -timeout=600s
```

✓ PASS; precedence, bounded grouping, formatter, malformed/overflowing hex and negative shift controls pass.

**Checker lowering/transfers**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase35ShiftAdmission|TestPhase35ScalarTransferMutants|TestPhase35NumericRoutePolicy)$' -count=1 -v -timeout=600s
```

✓ PASS; literal-only shifts, route-local numeric policy and value-transfer controls pass.

**Core validator contract/analysis**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s
```

✓ PASS; raw immediate laws, transfer paths and Byte controls pass.

**Origin validator contract/analysis**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s
```

✓ PASS; independent raw immediate laws and transfer paths pass.

**Interpreter**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase35BitwiseShiftVectorsEventFree|TestPhase35InterpreterWrongAnswerControls|TestPhase35UnusedByteRuntime)$' -count=1 -v -timeout=600s
```

✓ PASS; literal vectors, wrong-answer controls and event-free outcomes pass.

**Executionpeer**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^TestPhase35PureOperationCausalControls$' -count=1 -v -timeout=600s
```

✓ PASS; fabricated event controls pass; this is causal evidence only.

**C emission**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase35UnsignedShiftEmissionAndEventProjection|TestPhase35ShiftSiteControls)$' -count=1 -v -timeout=600s
```

✓ PASS; unsigned shifts, positive site count and dynamic/64/signed controls pass.

**Public application and Adler**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase35FocusedValueAnswers|TestPhase35ExistingAdlerAnswers|TestPhase35AdlerWrongShiftControl)$' -count=1 -v -timeout=600s
```

✓ PASS; interpreter/native public output and pinned Adler values pass; wrong-shift control is reached.

**CLI witness**

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase35PublicConformanceWitness$' -count=1 -v -timeout=600s
```

✓ PASS; public interpreter/native witness returns 264 and the alternate association control returns 8.

### Final report-grounding checks

After writing this report, run these exact named tests against the rendered Phase35 verification file. Each must appear as RUN and PASS with no skip.

~~~sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundednessThreeClassesAreEmpty|TestPhase35EvidenceGrounding)$' -count=1 -v -timeout=600s
~~~

Result: ✓ PASS. All three exact roots appeared RUN and PASS with no skip against the rendered report.

No server, broad test suite, broad native matrix, hosted job, push or UAT replay was run.

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| None declared or implied by this compiler-feature phase | N/A | No phase probe path appears in the plans/summaries and there is no conventional scripts/tests/probe-*.sh for this phase. | N/A |

## Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| BIT-01 | 35-01 through 35-08 | U64 bitwise operators have specified values and independent transfer/engine/causal evidence. | ✓ SATISFIED | R1, P03-T4, P04-T2, P05-T1/P05-T3, P07-T1/P07-T2; named vector, mutation and causal controls above. |
| BIT-02 | 35-01 through 35-08 | Literal shifts 0..63 are defined; invalid counts refuse; generated C contains no unguarded shift. | ✓ SATISFIED | R2, P03-T1/P03-T2, P04-T1, P06-T1, P07-T6; fresh refusal/freshness controls, optimization vectors and UBSan. |
| BIT-03 | 35-01, 35-02, 35-07, 35-08 | Hex U64 spelling, exact parsing, overflow/malformed refusal and formatter preservation. | ✓ SATISFIED | R3, P02-T3 and P07-T7; numeric ownership, route policy, grammar and formatter selectors. |
| BIT-04 | 35-01, 35-08 | Adler shift compresses sixteen doublings without changing pinned outputs. | ✓ SATISFIED | R4, P08-T1/P08-T2; fresh Adler, Phase29 interpreter/native and Phase31 matrix tests. |
| DX-22 dynamic-shift sub-item | Phase34-owned; delegated to BIT-02 | Dynamic shift amounts retain the Phase34 refusal family and are registered. | ✓ SATISFIED | TestPhase35ShiftAdmission, TestPhase35SurfaceManifestControls and TestPhase34SurfaceManifestControls. DX-22 ownership remains Phase34; Phase35 closes its explicitly pending literal-shift part. |

No Phase35-mapped requirement is orphaned. The DX-22 entry is acknowledged as a Phase34-owned delegated sub-item, not re-owned by this phase.

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| 53 plan-listed implementation/test/fixture paths | — | No TBD/FIXME/XXX, TODO/HACK/PLACEHOLDER, unfinished-placeholder phrases, empty implementation, hardcoded-empty output, or console.log-only implementation patterns. | — | No stub or unreferenced debt marker was found in the current phase scope. Test setup defaults and intentional empty derivations were not misclassified as production stubs. |

The 62-file code review is clean with zero findings; the two earlier WR-01/WR-02 warnings are recorded fixed with no open findings. The authored security register is 24/24 closed at ASVS L1 with zero open items. The UI gate is inapplicable because this phase adds no frontend; no graphical review or human-appearance claim is made.

## Advisory (New Scope, Unevidenced)

None. Re-verification surfaced no unevidenced new-scope finding.

## Human Verification Required

None. Both previously flagged test-tier items now have canonical fail-first enforcement, violation fixtures that fail at their intended scope assertions, and clean/default controls that pass. The feature has no visual or external-service behavior requiring subjective review.

## Gaps Summary

No observable truth, artifact, key link, requirement, or behavior-dependent claim failed. All 38 truths are verified; BIT-01 through BIT-04 are satisfied; all 48 declared artifacts, 24 declared links, and 23 context decisions/prohibitions are accounted for. The earlier two-item human gate is closed by executable test-tier enforcement with negative and clean controls. No gaps or human verification items remain.


---

Verified: 2026-10-14T23:30:56Z  
Verifier: gsd-verifier agent

## Preserved Phase 36 successor refresh — 2026-10-15T07:22:07Z

This bounded amendment revalidated the Phase 35 goal after the Phase 36 successor made the former fingerprint stale. It preserves the dated initial Phase 35 verification and review/disposition artifacts as history; it did not replay old plans, prior Phase 35 tests, UAT, or phase.complete.

The Phase 36 prior-phase receipt in `.planning/phases/36-scalar-helper-calls/36-REGRESSION.md` binds source/test state to `b51043c6566f793893c8c4e143493e7ec3af5bcd`, after the last compiler/test edit. I checked the owner-results manifest and all 13 raw log hashes; every expected root was anchored as RUN/PASS, all 13 commands exited 0, with 80 occurrences across 69 selected names and 13 packages, and no FAIL/SKIP. Phase 35 roots include shift admission/site controls, event-free interpreter vectors, native shift tracing, optimization matrix, engine wrong-answer controls, transfer mutants, and `TestPhase35ShiftUBSan`. The final acceptance receipt in `36-ACCEPTANCE-EVIDENCE.md` binds a clean source tree at `979357388171339005fd6dbd65aa760810ae9384`; 45 actual roots across 12 packages and matching hashes provide current integrated successor evidence. A tracked source/test comparison through current `HEAD 66cebc9f` found no changes after the regression revision; later commits are docs-only. Neither receipt is a full suite or hosted run.

The Phase 36 successor expands scalar calls, checker/call execution, and emission paths. Current Phase 35 regression roots still exercise bitwise values, admission, optimizer levels, transfer authority, and C17 shift safety; those implementation and refusal boundaries remain covered by current source and named controls. The Phase 36 52-file review is clean with zero findings, and its disposition records zero open findings. This is successor-delta coverage only; Phase 35's original review and disposition remain historical.

All eight Phase 35 plan artifact queries passed (48 declared artifact occurrences); prior source traces and named controls for Phase 35 links remain applicable. BIT-01 through BIT-04 remain satisfied with no new defect path. Nyquist remains validated; the authored L1 register retains 24/24 closed items and zero open, reused for applicability rather than a new full security audit. The UI safety gate remains inapplicable because there is no frontend. No new goal gap or human-verification item is outstanding; the score remains **38/38**.

---

_Verified: 2026-10-15T07:22:07Z_
_Verifier: the agent (gsd-verifier)_


## Phase 37 current-head refresh — 2026-10-15T21:17:32Z

This dated refresh preserves the initial Phase 35 report (2026-10-14T23:30:56Z) and the Phase 36 successor-staled refresh (2026-10-15T07:22:07Z) above. It independently re-established the phase contract from the roadmap, all eight plan truth/artifact/link declarations, requirements traceability, current source paths and current Phase35/36/37 receipts. The Phase36 helper and Phase37 public interpreter-file route are included as legitimate current consumers; frozen Phase35 fixtures remain interpreted through their explicit Phase36/37 ledgers.

The current checkout is HEAD `de6cde3d56be24382e38eb3b69f23e28e3721d65`. The run receipt records source_head `d48227882579947201f607814585c42e6ad29bd9`; the regression receipt records `6c5739bb88c150d2aedf8bdb6c4ecde0b6e6061e`. Those older identifiers are not used as a proxy for freshness: all 62 run inputs, 117 regression inputs, and all 29 raw-log hashes were independently compared to current bytes and match. The only current source-tree addition in the earlier closeout repair is the shared-validator-required synthetic coverage description in `session_phase35_closeout_test.go`; compiler/runtime implementation is unchanged by this verifier refresh.

The current source and proof path supports the 38/38 goal truths: checker, corevalidate and originvalidate retain independent value transfer/admission; executionpeer validates causal structure only; interpreter and native expected answers are compared to independent literals; the C emission tests require a positive shift-site count and unsigned widened operands; refusal, high-bit truncation, wrong-answer and false-evidence controls are reached. No C17 undefined-shift path or new defect/event path was found. The present focused commands passed as recorded above, and the two hash-bound receipts cover the broader named evidence without replaying the expensive matrix.

The remaining human decision is limited to the two unresolved 35-08 prohibitions with `verification: null` (BIT-01 and BIT-02). The verification gate cannot silently mark them resolved. The behavior score remains 38/38; status is `human_needed`. The root living documents and configuration are excluded from this fingerprint because they are outside the implementation/evidence slice and may be reconciled after phase acceptance.


## Phase 37 test-tier prohibition refresh — 2026-10-15T21:37:03Z

This amendment preserves the dated Phase37 pre-wiring report immediately above. Its 38/38 truth result was accurate, and its `human_needed` status correctly identified that the two plan prohibitions lacked explicit verification tiers at that time. That acceptance gap is now closed by the dated 35-08-PLAN.md amendment and executable enforcement; no override or human judgment was applied.

The amended descriptors are test-tier checks for BIT-01's causal-evidence boundary and BIT-02's freshness-claim boundary. I independently derived both requests using the canonical prohibition parser/projection helpers and reran the canonical prohibition-enforcement check for each. Both returned green with fail-first proof from their violation fixture. The existing Go evidence-grounding test rejected each bad fixture at the expected scope assertion, and accepted the unchanged clean evidence and default subject. The TAP adapter only invokes that Go test and asserts the exact RUN/PASS, no-fail/no-skip result.

For current regression acceptance I used 35-REFRESH-FINAL-RUN.json, not the older run/regression receipts whose test metadata input had since changed. All 122 current input hashes and 16 raw-log hashes match; all 175 package-root executions are RUN/PASS across 16 packages, with 165 distinct root test names and no missing/failing/skipped roots. The receipt includes 15 arithmetic and 26 admission source mutants caught, plus 33 vector rows at O0/O3/O3-LTO and 16 shift UBSan rows without skips. Its source_head is c1136586224e06d300b0b96b123db8af05a4be9b; the later commits through current HEAD d9f08d495476eb2fb2e522f84ef2c7150c85afd6 changed only planning evidence documents, and the receipt inputs/logs still match current bytes.

The current score remains 38/38. The derived inventory is five roadmap criteria, 34 unique plan-specific truths after the one duplicate, 48 artifact declarations across 45 unique paths, 24 declared links, 23 context decisions/prohibitions, and requirements BIT-01 through BIT-04 with no orphan. The four added enforcement artifacts are listed separately from the 48 plan declarations. The 66-file review/follow-up union has zero findings, the security register remains 24/24 closed at ASVS L1, and UI review remains inapplicable. No C17 shift undefined behavior, new defect path, gap, or human-verification item remains.

---

_Verified: 2026-10-15T21:37:03Z_
_Verifier: the agent (gsd-verifier)_

---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
verified: 2026-10-14T23:30:56Z
status: passed
score: 38/38 must-haves verified
covered_files:
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-02-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-03-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-03-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-04-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-04-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-05-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-05-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-06-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-06-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-07-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-07-SUMMARY.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-PLAN.md
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-SUMMARY.md
  - cmd/schway/main_phase32_routes_test.go
  - cmd/schway/main_phase35_test.go
  - docs/agent-surface.md
  - examples/bitwise.schway
  - examples/checksum/adler32.schway
  - internal/compiler/cgen/cgen_phase35_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase35_baseline_test.go
  - internal/compiler/check/check_phase35_numeric_test.go
  - internal/compiler/check/check_phase35_test.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_phase35_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_phase35_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_phase35_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase35_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_phase35_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_phase35_test.go
  - internal/compiler/session/session_phase272_identity_test.go
  - internal/compiler/session/session_phase273_identity_test.go
  - internal/compiler/session/session_phase34_compat_test.go
  - internal/compiler/session/session_phase34_limit_consumers_test.go
  - internal/compiler/session/session_phase34_prechange_test.go
  - internal/compiler/session/session_phase34_presentation_test.go
  - internal/compiler/session/session_phase34_surface_test.go
  - internal/compiler/session/session_phase35_adler_test.go
  - internal/compiler/session/session_phase35_admission_mutation_test.go
  - internal/compiler/session/session_phase35_closeout_test.go
  - internal/compiler/session/session_phase35_peer_mutation_test.go
  - internal/compiler/session/session_phase35_prechange_test.go
  - internal/compiler/session/session_phase35_surface_test.go
  - internal/compiler/session/session_phase35_values_test.go
  - internal/compiler/surface/refusals.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/surfacemanifest/manifest_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/numeric_phase35_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase34_recovery_test.go
  - internal/compiler/syntax/parser_phase34_refusal_test.go
  - internal/compiler/syntax/parser_phase35_baseline_test.go
  - internal/compiler/syntax/parser_phase35_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase35/numeric-delta.json
  - testdata/phase35/numeric-policy-baseline.json
  - testdata/phase35/prechange-corpus.json
  - testdata/phase35/recovery-after.json
  - testdata/phase35/shift-sites.json
  - testdata/phase35/surface-census.json
  - testdata/phase35/surface-witnesses.json
  - testdata/phase35/value-vectors.json
covered_digest: "v3:sha256:46c44e74287e82aa7d732c2d781fd33567acdfcd1f5b3906b9ba408ee6dc2f0b"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals Verification Report

**Phase Goal:** Ordinary Schway scalar code can use U64 ^, &, |, literal-amount << and >>, and hex literals, with one defined meaning on every engine and optimization level, no C17 shift undefined behavior, and no new defect path.

**Verified:** 2026-10-14T23:30:56Z  
**Status:** passed  
**Re-verification:** No — initial verification; no previous 35-VERIFICATION.md existed.

## Goal Achievement

The phase goal is present in the actual compiler and consumer path: lexer/parser and formatter preserve the new source forms; checking lowers five exact-width operations and admits only literal shift counts from 0 through 63; both independent raw-IR validators enforce the shift shape and independently derive value transfers; the interpreter and C emitter execute the operations; surface metadata and public examples expose the behavior. Focused executions exercise the real interpreter and native paths and compare them to independently authored decimal answers.

The two plan prohibition entries were explicitly inspected. Numeric correctness is supported by literal answer vectors, per-engine wrong-answer controls, and transfer mutants in the value-deriving peers. Executionpeer evidence is described only as causal/event-free admission. Fresh local optimization and UBSan runs are distinguished from historical receipts; no hosted or skipped lane is presented as fresh. No synthetic check descriptor or automatic prohibition test result was invented.

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
| Do not present engine agreement or causal event admission as independent proof of pure numeric results. | ✓ VERIFIED | Independent literal vectors, named wrong-answer controls and the value-peer transfer mutation harness provide numeric evidence. Executionpeer controls are described only as causal/event-free checks. No check descriptor was invented. |
| Do not describe historical hosted receipts or skipped optimization/sanitizer lanes as fresh Phase35 evidence. | ✓ VERIFIED | The fresh local O0/O3/O3-LTO and UBSan runs are recorded separately below. The report makes no hosted or cross-host claim and records no required lane as skipped. No check descriptor was invented. |

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

## Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| examples/bitwise.schway | x, y, witness | Program source flows through CLI, checking, interpreter/native runners | Yes; returned outcome 264 and application output 264 newline are compared to literal expectations | ✓ FLOWING |
| testdata/phase35/value-vectors.json | Operation, operands, expected decimal | Independently authored fixture read by value, optimization and sanitizer tests | Yes; each vector is executed by interpreter and generated native C | ✓ FLOWING |
| examples/checksum/adler32.schway | Adler result and imported checksum/adler_low values | Imported Schway modules run through existing interpreter/native consumers | Yes; pinned literal outputs and cleanup outcomes are compared | ✓ FLOWING |
| testdata/phase35/surface-witnesses.json and surface-census.json | Expected operator/manifest rows | Independent fixture files read by session surface tests against manifest generation | Yes; generated manifest is compared with witness/census, including copied drift controls | ✓ FLOWING |
| testdata/phase35/prechange-corpus.json | Frozen emitter output/refusal outcomes | Revision-pinned baseline consumed by session corpus test | Yes; current emitter outputs are compared byte-for-byte to old expected data | ✓ FLOWING |

## Behavioral Spot-Checks

All 22 exact automated selectors from the eight phase plans were run individually with GOCACHE=/tmp/ai-lang-verification-gocache, -count=1, -v and -timeout=600s. The runs covered 15 Go test binaries across 13 import paths (internal/external test packages are separate binaries). They produced 88 root RUN and 88 root PASS records across 66 distinct top-level names, with zero missing expected name multiplicities, failures or skips.

| Plan command | Named tests observed | Result |
|---|---|---|
| 35-01 task selectors | TestPhase35PreChangeEmitterCorpus; TestPhase35PreChangeCorpusControls | PASS, 2/2 |
| 35-01 regression selectors | TestPhase35NumericPolicyBaseline; TestPhase35NestedIndexBaseline | PASS, 2/2 |
| 35-02 grammar selectors | TestPhase35ScalarGrammar; TestPhase35FormatterRoundTrip; TestFormatLetBinaryExpressions | PASS, 3/3 |
| 35-02 recovery/depth selectors | TestPhase34RecoveryCorpus; TestPhase34RecoveryControls; TestPhase35ExpressionDepth; TestPhase35FormatterRoundTrip | PASS, 4/4 |
| 35-02 numeric selectors | TestPhase35NumericOwnership; Phase19 numeric roundtrip/token/malformed/admission/range/refusal selectors; TestPhase35NumericRoutePolicy | PASS, 9/9 |
| 35-03 core selectors | TestPhase35CoreShape; TestAllOperationKindsRegistered; TestPhase35ShiftAdmission | PASS, 3/3 |
| 35-03 source/Byte selectors | TestPhase35ShiftAdmission; TestPhase35UnusedByteSourceAdmission; TestPhase34BoolSpellingsRemainIdentifiers; TestPhase34CallArityBoundaries | PASS, 4/4 |
| 35-03 checker selectors | TestPhase35ShiftAdmission; TestPhase35NumericRoutePolicy; TestPhase35ScalarTransferMutants; TestPhase35UnusedByteSourceAdmission | PASS, 5/5 |
| 35-04 raw core admission | TestPhase35ShiftImmediateAdmission; TestPhase35UnusedByteIRAdmission | PASS, 4/4 across both peers |
| 35-04 validator analysis | TestPhase35ShiftImmediateAdmission; TestPhase35UnusedByteIRAdmission; TestPhase35ScalarTransferMutants; TestPhase35AnalysisTriggers; Phase19 independent OpConst controls | PASS, 6/6 |
| 35-04 origin/core analysis | TestPhase35ShiftImmediateAdmission; TestPhase35UnusedByteIRAdmission; TestPhase35ScalarTransferMutants; TestPhase35AnalysisTriggers | PASS, 4/4 |
| 35-05 interpreter/pathoracle | TestPhase35BitwiseShiftVectorsEventFree; TestPhase35PureOperationAncestry | PASS, 2/2 |
| 35-05 interpreter/Byte | TestPhase35BitwiseShiftVectorsEventFree; TestPhase35InterpreterWrongAnswerControls; TestPhase35UnusedByteRuntime | PASS, 3/3 |
| 35-05 executionpeer | TestPhase35PureOperationCausalControls | PASS, 1/1 |
| 35-06 C tracer/gates | TestPhase35NativeShiftTracer; TestPhase35EmitterAdmissionGates | PASS, 2/2 |
| 35-06 C shift sites | TestPhase35NativeShiftTracer; TestPhase35EmitterAdmissionGates; TestPhase35UnsignedShiftEmissionAndEventProjection; TestPhase35ShiftSiteControls | PASS, 4/4 |
| 35-07 public CLI/core | TestPhase321RunInputRequiredCLI; TestPhase35PublicConformanceWitness; TestPhase35FocusedValueAnswers; TestAllOperationKindsHandledAtEverySite | PASS, 4/4 |
| 35-07 native/value/surface | TestPhase35FocusedValueAnswers; TestPhase35BitwiseOptimizationMatrix; TestPhase35ShiftUBSan; TestPhase35EngineWrongAnswerControls; Phase34 and Phase35 surface document/limit selectors | PASS, 7/7 |
| 35-07 manifest controls | Phase34 surface limit/manifest regression selectors and Phase35 surface witness/census/control/limit selectors | PASS, 8/8 |
| 35-08 Adler/regressions | TestPhase29InterpreterFixtureVectors; TestPhase29NativeFixtureVectors; TestPhase35ExistingAdlerAnswers; TestPhase35AdlerWrongShiftControl; TestPhase31CaseMatrix | PASS, 5/5 |
| 35-08 docs/public projection | TestPhase35OperatorDocumentation; TestPhase35PublicDocumentValues | PASS, 2/2 |
| 35-08 evidence grounding | TestPhase35EvidenceGrounding; TestPhase35PreChangeEmitterCorpus; TestPhase35PreChangeCorpusControls; TestSummaryCoverageGroundedness | PASS, 4/4 |

Additional post-repair executions:

| Behavior | Command | Result | Status |
|---|---|---|---|
| Check/corevalidate/originvalidate wrong-transfer mutants | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35PeerTransferMutationControls$' -count=1 -v -timeout=600s | Exit 0; 15 actual implementation mutants were caught at named sinks; 3 no-op overlays were controlled and all changes restored. | ✓ PASS |
| Admission refusal mutants | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$' -count=1 -v -timeout=600s | Exit 0; 26 controls at 55 replacement sites, 46 intended failures; 3 no-op and 3 restored controls. | ✓ PASS |
| Prior static write target refusals | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^TestPhase35ShiftTargetFreshness$' -count=1 -v -timeout=600s | Exit 0; both peers pass eight left/right prior-write cases. | ✓ PASS |
| Legitimate shift producers | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^TestPhase35ShiftProducerShapes$' -count=1 -v -timeout=600s | Exit 0; assignment, branch and loop producers pass in both peers. | ✓ PASS |
| Parser/formatter repairs | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase35GroupedNegativeShiftAmountRefusal|TestPhase35FormatterPreservesSeparatedShiftTokens|TestPhase35ScalarGrammar|TestPhase35ExpressionDepth|TestPhase35FormatterRoundTrip|TestPhase35RecoveryDeltaControls)$' -count=1 -v -timeout=600s | Exit 0; all six named roots pass, including grouped-negative and separated-token cases. | ✓ PASS |
| Native optimization matrix | TestPhase35BitwiseOptimizationMatrix from plan command 18 | 33 independently authored literal rows execute at O0, O3 and O3-LTO; 99 native executions and interpreter comparisons pass. | ✓ PASS |
| Shift undefined-behavior sanitizer | TestPhase35ShiftUBSan from plan command 18 | All 16 shift fixture rows pass with Clang UBSan; no skip. | ✓ PASS |
| Generated C site scan and runtime | TestPhase35NativeShiftTracer, TestPhase35UnsignedShiftEmissionAndEventProjection and TestPhase35ShiftSiteControls | Both expected C sites are found; dynamic/64/signed/removed-site controls are reached; native boundary values pass. | ✓ PASS |

No full workspace suite, broad package union, build/vet, cross-program matrix, hosted dispatch or install was run.

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
| None | — | No TBD, FIXME or XXX markers in the 58 reviewed/changed source, test, fixture and contract paths. | — | The broad empty-return scan found intentional empty derivations for current resource ownership and call-boundary attributes, plus test helpers; none feeds the new scalar value path as a stub. |

The warning-level scan also matched an existing test comment that uses “placeholder” to describe an unused fixture field. It is not an unfinished implementation. No phase feature value flows from a hardcoded empty fallback.

## Human Verification Required

None. This phase adds no visual interface, external service flow or subjective runtime interaction that remains unverified. Current execution is local to macOS arm64 with Go 1.24 and Apple Clang 21. The separate Phase38 hosted/two-host receipts remain pending and are not claimed here. Whole-loop cold/warm feedback distributions and human readability were not measured; they are not Phase35 acceptance criteria.

## Gaps Summary

No must-have failed or remains behavior-unverified. The parser/formatter repairs, inherited validator registry wiring, and prior-static-write shift-target refusal were checked against current source, and their relevant tests plus the two current-snapshot mutation harnesses passed. The focused native matrix and UBSan lanes executed without skips.

No new Go dependency or backend was introduced: go.mod remains standard-library-only, and the existing C17/Clang native route is reused. The locally observed host and toolchain are recorded as scope limits; this verification does not claim cross-host results or fresh hosted receipts.

---

Verified: 2026-10-14T23:30:56Z  
Verifier: gsd-verifier agent

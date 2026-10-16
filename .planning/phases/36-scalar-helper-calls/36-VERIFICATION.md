---
phase: 36-scalar-helper-calls
verified: 2026-10-16T21:21:58Z
status: passed
score: "32/32 must-haves verified"
covered_files:
  - ".planning/LANGUAGE-MATURITY.md"
  - ".planning/PRODUCT-ROADMAP.md"
  - ".planning/REQUIREMENTS.md"
  - ".planning/ROADMAP.md"
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REFRESH3-FIX.md"
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REVIEW-DISPOSITION.md"
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-01-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-01-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-02-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-02-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-03-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-03-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-04-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-04-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-05-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-05-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-06-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-06-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-07-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-07-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-08-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-08-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md"
  - ".planning/phases/36-scalar-helper-calls/36-371-REFRESH-GATES.md"
  - ".planning/phases/36-scalar-helper-calls/36-371-REFRESH-RUN.json"
  - ".planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md"
  - ".planning/phases/36-scalar-helper-calls/36-CONTEXT.md"
  - ".planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH3-GATES.md"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH3-PROHIBITIONS.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH3-REGRESSION.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH3-RUN.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH3-TRANSITION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-DISPOSITION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-SECURITY.md"
  - ".planning/phases/36-scalar-helper-calls/36-UI-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-VALIDATION.md"
  - ".planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs"
  - ".planning/phases/36-scalar-helper-calls/prohibition-causal-violation.json"
  - ".planning/phases/36-scalar-helper-calls/prohibition-clean.json"
  - ".planning/phases/36-scalar-helper-calls/prohibition-outcome-violation.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-REVIEW.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-SECURITY.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-VALIDATION.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-COMPLETION-2026-10-16.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-DISPOSITION-AMENDMENT-2026-10-16.md"
  - "cmd/schway/main.go"
  - "cmd/schway/main_phase36_test.go"
  - "examples/helper_calls/crc_step.schway"
  - "examples/helper_calls/importer.schway"
  - "examples/helper_calls/public.schway"
  - "internal/compiler/ast/ast.go"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_exit_literals_test.go"
  - "internal/compiler/cgen/cgen_phase273_test.go"
  - "internal/compiler/cgen/cgen_phase33_scalar_defect_test.go"
  - "internal/compiler/cgen/cgen_phase35_test.go"
  - "internal/compiler/cgen/cgen_phase36_caller_shapes_test.go"
  - "internal/compiler/cgen/cgen_phase36_test.go"
  - "internal/compiler/cgen/cgen_phase371_legacy_defect_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/cgen/cgen_program_test.go"
  - "internal/compiler/check/check.go"
  - "internal/compiler/check/check_phase34_refusal_test.go"
  - "internal/compiler/check/check_phase36_scalar_calls_test.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/core/core.go"
  - "internal/compiler/core/core_test.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/corevalidate/corevalidate_phase36_legacy_internal_test.go"
  - "internal/compiler/corevalidate/corevalidate_phase36_test.go"
  - "internal/compiler/execution/defect_reasons.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go"
  - "internal/compiler/executionpeer/executionpeer_phase36_test.go"
  - "internal/compiler/executionpeer/executionpeer_phase371_test.go"
  - "internal/compiler/interp/interp.go"
  - "internal/compiler/interp/interp_phase36_test.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/native/native_app.go"
  - "internal/compiler/native/native_phase36_test.go"
  - "internal/compiler/originvalidate/originvalidate.go"
  - "internal/compiler/originvalidate/originvalidate_phase36_test.go"
  - "internal/compiler/pathoracle/pathoracle.go"
  - "internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go"
  - "internal/compiler/pathoracle/pathoracle_phase36_test.go"
  - "internal/compiler/session/session.go"
  - "internal/compiler/session/session_modules.go"
  - "internal/compiler/session/session_phase16_emitter_inventory_test.go"
  - "internal/compiler/session/session_phase34_limit_consumers_test.go"
  - "internal/compiler/session/session_phase34_surface_test.go"
  - "internal/compiler/session/session_phase35_surface_test.go"
  - "internal/compiler/session/session_phase36_byteview_test.go"
  - "internal/compiler/session/session_phase36_calls_test.go"
  - "internal/compiler/session/session_phase36_legacy_calls_test.go"
  - "internal/compiler/session/session_phase36_peer_mutation_test.go"
  - "internal/compiler/session/session_phase36_prechange_test.go"
  - "internal/compiler/session/session_phase36_surface_test.go"
  - "internal/compiler/session/session_phase371_compatibility_test.go"
  - "internal/compiler/session/session_phase371_defect_reason_test.go"
  - "internal/compiler/session/verification_groundedness_test.go"
  - "internal/compiler/surface/refusals.go"
  - "internal/compiler/surfacemanifest/manifest.go"
  - "internal/compiler/surfacemanifest/manifest_test.go"
  - "internal/compiler/syntax/format.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase36_test.go"
  - "testdata/phase16/public-emitter-consumers.json"
  - "testdata/phase36/helper-values.json"
  - "testdata/phase36/new-fixture-ledger.json"
  - "testdata/phase36/prechange-corpus.json"
  - "testdata/phase36/surface-census.json"
  - "testdata/phase36/surface-witnesses.json"
  - "testdata/phase371/compatibility-amendments.json"
covered_digest: "v3:sha256:407593878a8526430b8ea914cfbe52f56abb68d1905e15df666c361f380c5dcb"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: "32/32"
  gaps_closed: []
  gaps_remaining: []
  regressions: []
decision_coverage:
  honored: 26
  total: 26
  not_honored: []
---

# Phase 36: Scalar Helper Calls Verification Report

**Phase Goal:** A scalar function calls another scalar function with one U64 or Bool argument and uses its result, including imported modules, with the same defect behavior on both engines and independent call derivation by every peer.

**Verified:** 2026-10-16T21:21:58Z
**Status:** passed
**Verification mode:** Initial-mode refresh. The previous report had no gaps section, so all roadmap and plan truths were re-established rather than presumed.

## Goal Achievement

The source path is present from direct scalar-call syntax through checking, typed core, independent validation, interpreter execution, native emission, and execution-peer attribution. Local and imported helper programs consume returned values. The scalar fixture pins its expected result independently. Dynamic callee defects retain their exact reason on both engines; the application routes exit 65. Every peer derives its own portion of call safety and causality; execution-peer acceptance is not used as evidence of numeric correctness or source intent.

The previous report is preserved byte-for-byte at 36-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md. The current Phase 37.1 source delta intersecting Phase 36 was reviewed directly: cgen_program.go, corevalidate.go, executionpeer.go, parser.go, executionpeer_phase33_terminal_test.go, session_phase36_prechange_test.go, and the test-only cgen_exit_literals_test.go. The native defect shell now emits a structured terminal defect for conformance capture while the application shell preserves the defect reason and exit 65. Raw UTF-8 defect reasons are validated before clone normalization. The strict compatibility comparator continues to compare exact emitted bytes and does not normalize them or add fixture-ledger rows.

The eight Phase 36 plans are complete. The fresh current-gate receipts report a bounded 250-selector registry across 16 groups, 251 root events, 240 distinct names, and zero missing, failing, or skipped roots. They also report five formerly failing selectors and 26 previously unrun selectors passing, with unchanged green roots reused against their recorded basis. A separate source-bound D18 receipt records four green unflagged proof subjects and a 19-root union with 254 subtests and no skips. The explicit capture-pressure root remains excluded. These scoped receipts are not a full acceptance inventory or a package/workspace suite.

| # | Observable truth | Status | Evidence |
|---:|---|---|---|
| 1 | One-argument U64 and Bool scalar calls use their result locally and through imports; recursion and mutual recursion remain refused. | ✓ VERIFIED | TestPhase36HelperAnswersAndDefects, TestPhase36HelperOptimizationMatrix, TestPhase36CallImportAndRefusalBoundaries; parser, checker, and refusal paths inspected. |
| 2 | Callee defects preserve their reason and exit 65 on both engines; ByteView bounds remain a separate obligation. | ✓ VERIFIED | TestPhase36HelperAnswersAndDefects and current TestPhase33ScalarExplicitDefectShells receipt; interpreter/native outcome tests and D-18 boundary inspected. |
| 3 | Checker, corevalidate, originvalidate, pathoracle, and executionpeer independently derive call constraints; argument/result/defect mutations are caught. | ✓ VERIFIED | TestPhase36CallSourceIntentControls, TestPhase36RawCallShapeAndFreshness, TestPhase36OriginCallFreshnessAndTransfer, TestPhase36CallCausalAttributionControls. |
| 4 | No-call C output retains literal bytes and the source-free manifest records bounded arity-one calls without recursion. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaseline, TestPhase36SurfaceManifestCallWitnesses; current exact comparator and manifest controls inspected. |
| 5 | Literal baseline and controls predate compiler edits. | ✓ VERIFIED | Git ancestry and TestPhase36NoCallGeneratedCByteBaselineControls. |
| 6 | Frozen pre-change programs have applicable emitter routes accounted for, including refusals. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaseline checks source inventory and route outcomes. |
| 7 | Existing golden bytes remain immutable and new fixtures have exact ledger rows. | ✓ VERIFIED | Baseline omission/false-exemption controls and current fixture-ledger checks. |
| 8 | Direct one-argument calls parse, recover, and format within the existing depth bound. | ✓ VERIFIED | Parser implementation and TestPhase36ScalarCallFormatting; parser roots appear in current regression receipt. |
| 9 | All four U64/Bool signature pairs lower to typed fresh results, including initialized Bool identifier tails. | ✓ VERIFIED | TestPhase36ScalarCallLoweringAndTypePairs and helper answer tests. |
| 10 | Arguments execute once in left-to-right order; unused-Byte and recursion refusals remain. | ✓ VERIFIED | Checker lowering inspection, source-intent, caller-boundary, and refusal tests. |
| 11 | Raw core calls independently validate callee, signature, initialized argument, result type, and fresh target. | ✓ VERIFIED | TestPhase36RawCallShapeAndFreshness and refusal-removal controls. |
| 12 | Origin analysis transfers calls as initialized, value-unknown scalars through branches and loops. | ✓ VERIFIED | TestPhase36OriginCallFreshnessAndTransfer and refusal-removal controls. |
| 13 | Scalar calls add no loan ancestry or resource/back-edge authority. | ✓ VERIFIED | TestPhase36ScalarCallAuthorityBoundaries; pathoracle treats a call as a value boundary. |
| 14 | Repeated calls receive contiguous per-parent/site ordinals; child counters reset under a fresh parent. | ✓ VERIFIED | Interpreter tracer and ordinal/dynamic native path controls. |
| 15 | Interpreter results/defects and execution-peer causal attribution remain separate claims. | ✓ VERIFIED | Interpreter answer/defect tests plus causal attribution controls; peer does not assert arithmetic. |
| 16 | Evidence retention, path/counter bounds, and exhaustion remain distinct from numeric execution. | ✓ VERIFIED | Interpreter evidence boundaries, event-capacity, and native capture outcome tests. |
| 17 | Native calls use sequenced typed arguments/results and stable dynamic paths. | ✓ VERIFIED | Native invocation tests and C emitter inspection. |
| 18 | Repeated-call captures and exact event/path/ordinal boundaries have named controls. | ✓ VERIFIED | Native dynamic invocation, capacity, ordinal, and path-boundary tests. |
| 19 | Optional capture exhaustion preserves application output, exit, and later defects. | ✓ VERIFIED | TestPhase36NativeCallCaptureOutcomeSeparation and wired fail-first prohibition controls. |
| 20 | Old C output and affected emitter inventory are independently checked. | ✓ VERIFIED | No-call byte baseline and emitter-inventory controls; current EmitNative consumer registration is included. |
| 21 | Local/imported repeated crc_step helpers produce pinned answers on both engines. | ✓ VERIFIED | Helper answer/optimization tests and independent helper-values.json fixture. |
| 22 | Four type pairs, nested/computed arguments, arithmetic use, mutation, and Bool conditions have discriminating witnesses. | ✓ VERIFIED | Helper answer/type-pair roots; fixture values were computed independently, not by a Schway compiler or engine. |
| 23 | Dynamic defects preserve exact reasons and exit 65 across native optimization lanes; bounds remain distinct. | ✓ VERIFIED | Optimization matrix, legacy bounds, and caller-before-entry controls; current dynamic-defect root in D18 receipt. |
| 24 | Public conformance retains its input contract and capture stays separate from process outcome. | ✓ VERIFIED | Public helper conformance and native capture-outcome roots. |
| 25 | Source-free manifest states bounded arity-one contexts and all four type pairs. | ✓ VERIFIED | Surface manifest tests and independently authored witnesses. |
| 26 | Complete-program witnesses and finite source/census checks cover manifest claims. | ✓ VERIFIED | Surface witness/census tests and current census receipt inputs. |
| 27 | Admission, refusal, hint, restriction, and census controls are bidirectional; prior witness bytes stay frozen. | ✓ VERIFIED | Manifest call controls and surface limit-consumer tests. |
| 28 | Integrated no-call programs retain literal C bytes through frozen routes. | ✓ VERIFIED | Current TestPhase36NoCallGeneratedCByteBaseline root. |
| 29 | Acceptance evidence separates source, raw contract, numeric, causal, and bounds proofs. | ✓ VERIFIED | Acceptance evidence and distinct owner tests/receipts. |
| 30 | Current inspection, fresh evidence, and historical/hosted receipts are distinguished; phase ownership is preserved. | ✓ VERIFIED | Transition and acceptance artifacts; this report makes no hosted-evidence claim. |
| 31 | Optional evidence capture exhaustion preserves application stdout, stderr, exit status, and later defects. | ✓ VERIFIED | Native capture-outcome roots and the canonical outcome prohibition producer; this claim is separate from numeric or peer proof. |
| 32 | Causal-peer acceptance does not prove scalar arithmetic or source argument/result intent. | ✓ VERIFIED | Canonical causal prohibition producer and controls; numeric and source-intent tests own those separate claims. |

**Score:** 32/32 truths verified; behavior-unverified: 0.

## Required Artifacts

All 24 declarations across the eight plans exist and are substantive. Implementation/test artifacts are wired through the named consumers below. The two queried generic key-link results could not resolve the plans' descriptive from-labels as paths; all 19 links were therefore traced manually against source and tests.

| Plan | Artifact | Status | Verification |
|---|---|---|---|
| 36-01 | testdata/phase36/prechange-corpus.json | ✓ VERIFIED | Literal frozen source/emitted-byte rows and explicit route outcomes are consumed by the comparator. |
| 36-01 | internal/compiler/session/session_phase36_prechange_test.go | ✓ VERIFIED | Re-discovers inventory and compares live emission against the corpus with drift/omission controls. |
| 36-01 | testdata/phase36/new-fixture-ledger.json | ✓ VERIFIED | Explicit ledger is consumed by baseline controls; no silent fixture exemption. |
| 36-02 | internal/compiler/syntax/parser_phase36_test.go | ✓ VERIFIED | Call grammar, depth, recovery, and formatter tests exercise parser/formatter. |
| 36-02 | internal/compiler/check/check_phase36_scalar_calls_test.go | ✓ VERIFIED | Source intent, signature, import, and refusal controls exercise checker lowering. |
| 36-03 | internal/compiler/corevalidate/corevalidate_phase36_test.go | ✓ VERIFIED | Raw OpCall contract and freshness controls exercise independent validator. |
| 36-03 | internal/compiler/originvalidate/originvalidate_phase36_test.go | ✓ VERIFIED | Independent origin and unknown-result transfer controls exercise CFG analysis. |
| 36-03 | internal/compiler/pathoracle/pathoracle_phase36_test.go | ✓ VERIFIED | Call ancestry and retained authority controls exercise path analysis. |
| 36-04 | internal/compiler/interp/interp_phase36_test.go | ✓ VERIFIED | Interpreter answers, defects, activation paths, and evidence boundaries are exercised. |
| 36-04 | internal/compiler/executionpeer/executionpeer_phase36_test.go | ✓ VERIFIED | Independent ordinal, ancestry, and causal attribution controls are exercised. |
| 36-05 | internal/compiler/cgen/cgen_phase36_test.go | ✓ VERIFIED | Native call identity, emitter inventory, and capacity behavior are exercised. |
| 36-05 | internal/compiler/native/native_phase36_test.go | ✓ VERIFIED | Actual process outcomes and optional capture status are compared separately. |
| 36-06 | examples/helper_calls/crc_step.schway | ✓ VERIFIED | Local repeated-call program parses, checks, and runs through both app engines. |
| 36-06 | examples/helper_calls/importer.schway | ✓ VERIFIED | Closed-module import program runs through both app engines. |
| 36-06 | examples/helper_calls/public.schway | ✓ VERIFIED | Public conformance wrapper retains its existing input contract. |
| 36-06 | testdata/phase36/helper-values.json | ✓ VERIFIED | Literal independent answers and defect expectations feed exact behavioral assertions. |
| 36-06 | internal/compiler/session/session_phase36_calls_test.go | ✓ VERIFIED | Local/imported answers, optimization lanes, defect behavior, and capture boundary are driven through both engines. |
| 36-07 | testdata/phase36/surface-witnesses.json | ✓ VERIFIED | Independent complete-program acceptance/refusal cases feed manifest controls. |
| 36-07 | testdata/phase36/surface-census.json | ✓ VERIFIED | Exact source/limit census is consumed by finite census tests. |
| 36-07 | internal/compiler/session/session_phase36_surface_test.go | ✓ VERIFIED | Manifest witnesses, census, and bidirectional controls are exercised. |
| 36-08 | .planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md | ✓ VERIFIED | Separates proof owners, tests, limitations, and evidence dates. |
| 36-08 | .planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md | ✓ VERIFIED | Producer-backed test-tier prohibitions and their controls are documented and consumed. |
| 36-08 | .planning/PRODUCT-ROADMAP.md | ✓ VERIFIED | Dated maturity/next-capability evidence remains in its project owner document. |
| 36-08 | .planning/LANGUAGE-MATURITY.md | ✓ VERIFIED | Dated capability evidence remains in its project owner document. |

## Key Link Verification

| Plan | From → To | Status | Trace |
|---|---|---|---|
| 36-01 | pre-change comparator → prechange-corpus.json | ✓ WIRED | Comparator loads literal expected bytes and checks live emitter/module routes. |
| 36-02 | syntax scalarExprAtom → checker scalar expression builder | ✓ WIRED | Call AST uses direct callee name and its sole argument; checker resolves and lowers it. |
| 36-02 | checker signature resolution → core.OpCall | ✓ WIRED | Stable callee ID, initialized argument, fresh target, and declared result type are emitted. |
| 36-03 | corevalidate raw OpCall → declared callee/function/place/type tables | ✓ WIRED | Core validator derives the call contract without checker summary. |
| 36-03 | origin scalar CFG transfer → call-result target state | ✓ WIRED | Result is initialized as a value-unknown declared scalar through control flow. |
| 36-03 | pathoracle scalar-call trigger → loan ancestry | ✓ WIRED | Typed scalar-safe calls retain owner and loan restrictions. |
| 36-04 | interpreter dynamic frame → InvocationSegment.Ordinal | ✓ WIRED | One reserved child activation is reused by call event and child frame. |
| 36-04 | executionpeer observed event stack → own ordinal map/core topology | ✓ WIRED | Peer derives activation ancestry and ordinal independently; no producer verdict or scalar replay. |
| 36-04 | bounded interpreter collector → result plus evidence error | ✓ WIRED | Incomplete evidence is separate from numeric/output/defect result. |
| 36-05 | new scalar-call emitter branch → invocation strings/counters | ✓ WIRED | One bounded reservation is reused through final serialization. |
| 36-05 | evidence budget guards → conformance/app outcome | ✓ WIRED | Capture failure is distinct from app streams and result. |
| 36-05 | old emission branch → frozen literal corpus | ✓ WIRED | Exact baseline comparison preserves no-call output bytes. |
| 36-06 | local/importer source → interpreter and native retained app runners | ✓ WIRED | Literal stdout/stderr/exit expectations drive both engines. |
| 36-06 | helper callee defects → app outcome and causal capture | ✓ WIRED | Exact reason/exit and invocation attribution have separate assertions. |
| 36-06 | public unused-Byte wrapper → public run engines | ✓ WIRED | Existing conformance route is exercised without a new U64 flag/default input. |
| 36-07 | source-free manifest → independent full-program witnesses | ✓ WIRED | Hand-authored semantic expectations are checked against actual admission. |
| 36-07 | finite source census → Phase 34/35/36 consumers | ✓ WIRED | Exact changed branch and claim supersession is checked without blanket omission. |
| 36-08 | frozen corpus → current emitter outputs | ✓ WIRED | Literal comparator and active drift/omission/false-exemption controls cover routes. |
| 36-08 | acceptance evidence → named owner selectors/literal expectations | ✓ WIRED | Distinct proof owners and exact selector references prevent causal-to-numeric promotion. |

## Data-Flow Trace

No rendered UI or database value applies. Helper values flow from local or imported source through parser and typed checker output into interpreter frames or emitted C calls, then to captured result/output. helper-values.json supplies independently computed expected values. The execution peer derives call identity, path, and ordinal; it does not establish arithmetic correctness.

| Artifact | Data | Source and flow | Status |
|---|---|---|---|
| Helper examples | U64/Bool arguments and returned values | Source → parser/checker → typed core → interpreter/native → captured output | ✓ FLOWING |
| Interpreter | Return value or callee defect | Dynamic frame executes callee and propagates result/reason | ✓ FLOWING |
| C generator/native app | Typed call result and process outcome | Generated call → native process stdout/stderr/exit | ✓ FLOWING |
| Execution peer | Call identity and causal path | Runtime event sequence → independently derived path/ordinal | ✓ FLOWING |
| Surface manifest | Bounded source-free call contract | Manifest data → surface tests and CLI route | ✓ FLOWING |

## Behavioral Spot-Checks

| Behavior | Command or receipt | Result | Status |
|---|---|---|---|
| Local/imported helper answers and defects | GOCACHE=/tmp/ai-lang-verification-gocache-count1-v120 go test ./internal/compiler/session -run '^TestPhase36HelperAnswersAndDefects$' -count=1 -timeout=600s | PASS; package result 1.220s | ✓ PASS |
| Dynamic explicit defect terminal behavior | Current D18 bounded receipt: TestPhase33ScalarExplicitDefectShells | Interpreter reason/peer validation, native conformance at O0/O3/O3-LTO, and app stderr/exit 65 all pass | ✓ PASS (recorded receipt) |
| Parser/core/origin/peer, path/ordinal/capacity, capture, bounds, and baseline controls | Current bounded registry and source-bound D18 receipts | No missing, failing, or skipped roots in the recorded scope | ✓ PASS (recorded receipts) |

No suite or service was started. The direct spot-check was a single named selector.

## Test Quality Audit

The positive arithmetic and engine-equivalence claim is owned by TestPhase36HelperAnswersAndDefects and the optimization matrix, with literal expected values whose fixture provenance is independent of the Schway compiler/engines. Parser, raw-core, origin, path, source-intent, and causal-peer controls are separate tests with mutations aimed at their own asserted property. The no-call literal comparator has drift, omission, and false-exemption controls. The two test-tier prohibitions have wired producers and clean/violation fixtures: capture exhaustion cannot suppress actual process outcomes, and causal-peer acceptance cannot stand in for numeric/source-intent proof. The current bounded receipts exclude skipped or missing roots from credit. No test count is used as a substitute for these assertion-specific witnesses.

## Requirements Coverage

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| CALL-01 | 36-02 through 36-06 | One-argument U64/Bool calls use returned values across local/imported helper paths, preserving defect behavior and recursion refusal. | ✓ SATISFIED | Both-engine answer/defect tests and parser/checker/runtime/native evidence. |
| CALL-02 | 36-01 through 36-08 | Every peer independently derives call constraints; compatibility and bounded public surface are checked. | ✓ SATISFIED | Source/core/origin/path/peer tests, mutation controls, manifest and literal baseline. |

No phase-assigned requirement is orphaned. D-18 keeps helper arithmetic defects distinct from legacy ByteView callee bounds and caller-before-entry argument bounds.

## Probe Execution

N/A. No Phase 36 plan declares a probe and no conventional compiler-phase probe is present.

## Anti-Patterns Found

| File | Pattern | Severity | Impact |
|---|---|---|---|
| — | No unreferenced TBD/FIXME/XXX markers, stubs, empty implementations, or console-only implementations found in phase-touched code. | — | A lexical placeholder match is a comment about a dead/unused field, not an implementation stub. |

## Security and UI Gates

The refreshed security delta check reports 13/13 checks closed with zero open threats, supplementing the existing Phase 36 authored register, which remains at zero open findings. The independent review of the eight-file regression delta is clean; the prior unchanged production review remains clean and the dated disposition amendment preserves original findings/history. UI review is not applicable to this compiler-only phase.

## Human Verification Required

None. The phase’s asserted runtime behavior has named automated evidence. No visual, hosted-service, Linux, latency, physical-cleanup, or human-readability claim is made. Existing UAT/dispositions are preserved and were not replayed; no new human approval is asserted.

## Gaps Summary

No Phase 36 must-have gaps remain. Local and imported U64/Bool helper calls use returned values, preserve callee defect reason/exit behavior on both engines, and retain independent checker, validator, and execution-peer derivation. Phase 38 and EVD-17 remain with their future owner and are outside this completed phase’s contract.

---

Verified: 2026-10-16T21:21:58Z
Verifier: gsd-verifier

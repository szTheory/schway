---
phase: 36-scalar-helper-calls
verified: 2026-10-16T04:07:18Z
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
  - "cmd/schway/main.go"
  - "cmd/schway/main_phase36_test.go"
  - "examples/helper_calls/crc_step.schway"
  - "examples/helper_calls/importer.schway"
  - "examples/helper_calls/public.schway"
  - "internal/compiler/ast/ast.go"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_phase273_test.go"
  - "internal/compiler/cgen/cgen_phase33_scalar_defect_test.go"
  - "internal/compiler/cgen/cgen_phase35_test.go"
  - "internal/compiler/cgen/cgen_phase36_test.go"
  - "internal/compiler/cgen/cgen_phase36_caller_shapes_test.go"
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
covered_digest: "v3:sha256:c90fbf587b476354621ea29e16323bb1d73741c6e287b39b0e28c656ddaf3dc0"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 26
  total: 26
  not_honored: []
---

# Phase 36: Scalar Helper Calls Verification Report

**Phase Goal:** A scalar function calls another scalar function with one U64 or Bool argument and uses its result, including imported modules, with the same defect behavior on both engines and independent call derivation by every peer.

**Verified:** 2026-10-16T04:09:33Z
**Status:** passed
**Verification mode:** Full independent verification. The prior report had no `gaps:` section, so its previously passed truths were rechecked rather than presumed.

## Goal Achievement

The goal is present in the parser, checker, typed core, independent validators, interpreter, native C emission, and execution peer. I inspected those paths and the current tests; current owner receipts and raw logs bind to HEAD `52cdc9954297e0d002d4fa8ca20fe444ddb32f85`. All 32 must-haves (four roadmap truths, 26 distinct plan truths, and two test-tier prohibitions) are verified.

| # | Truth | Status | Evidence |
|---:|---|---|---|
| 1 | One-argument U64 and Bool scalar calls use their result locally and through imports; recursion and mutual recursion remain refused. | ✓ VERIFIED | `TestPhase36HelperAnswersAndDefects`, `TestPhase36HelperOptimizationMatrix`, `TestPhase36CallImportAndRefusalBoundaries`; parser, checker, and refusal paths inspected. |
| 2 | Callee defects preserve reason and exit 65 on both engines; legacy ByteView bounds remain a distinct obligation. | ✓ VERIFIED | Fresh owner/regression roots cover helper defects, legacy ByteView callee bounds, and caller-before-entry controls; D-18 keeps these boundaries separate. |
| 3 | Check, corevalidate, originvalidate, and executionpeer independently derive calls; argument/result/defect mutations are caught. | ✓ VERIFIED | `TestPhase36CallSourceIntentControls`, `TestPhase36RawCallShapeAndFreshness`, `TestPhase36OriginCallFreshnessAndTransfer`, `TestPhase36CallCausalAttributionControls`. |
| 4 | No-call C output retains literal bytes and the source-free manifest records bounded arity-one calls without recursion. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline`, `TestPhase36SurfaceManifestCallWitnesses`, baseline ancestry and surface controls. |
| 5 | Literal baseline and controls predate compiler edits. | ✓ VERIFIED | Git ancestry check and `TestPhase36NoCallGeneratedCByteBaselineControls`. |
| 6 | Frozen pre-change programs have applicable emitter routes accounted for, including refusals. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline` checks source inventory and route outcomes. |
| 7 | Existing golden bytes remain immutable and new fixtures have exact ledger rows. | ✓ VERIFIED | Baseline omission/false-exemption controls and current fixture ledger checks. |
| 8 | Direct one-argument calls parse, recover, and format within the existing depth bound. | ✓ VERIFIED | Parser implementation and `TestPhase36ScalarCallFormatting`; parser roots are in fresh regression receipt. |
| 9 | All four U64/Bool signature pairs lower to typed fresh results, including initialized Bool identifier tails. | ✓ VERIFIED | `TestPhase36ScalarCallLoweringAndTypePairs` and helper answer tests. |
| 10 | Arguments execute once in left-to-right order; unused-Byte and recursion refusals remain. | ✓ VERIFIED | Checker lowering inspection, source-intent, caller-boundary, and refusal tests. |
| 11 | Raw core calls independently validate callee, signature, initialized argument, result type, and fresh target. | ✓ VERIFIED | `TestPhase36RawCallShapeAndFreshness` and refusal-removal controls. |
| 12 | Origin analysis transfers calls as initialized, value-unknown scalars through branches and loops. | ✓ VERIFIED | `TestPhase36OriginCallFreshnessAndTransfer` and refusal-removal controls. |
| 13 | Scalar calls add no loan ancestry or resource/back-edge authority. | ✓ VERIFIED | `TestPhase36ScalarCallAuthorityBoundaries`; pathoracle treats a call as a value boundary. |
| 14 | Repeated calls receive contiguous per-parent/site ordinals; child counters reset under a fresh parent. | ✓ VERIFIED | Interpreter tracer and ordinal/dynamic native path controls. |
| 15 | Interpreter results/defects and execution-peer causal attribution remain separate claims. | ✓ VERIFIED | Interpreter answer/defect tests plus causal attribution tests; peer does not assert arithmetic. |
| 16 | Evidence retention, path/counter bounds, and exhaustion remain distinct from numeric execution. | ✓ VERIFIED | Interpreter evidence boundaries, event-capacity, and native capture outcome tests. |
| 17 | Native calls use sequenced typed arguments/results and stable dynamic paths. | ✓ VERIFIED | Native invocation tests and C emitter inspection. |
| 18 | Repeated-call captures and exact event/path/ordinal boundaries have named controls. | ✓ VERIFIED | Native dynamic invocation, capacity, ordinal, and path boundary tests. |
| 19 | Optional capture exhaustion preserves application output, exit, and later defects. | ✓ VERIFIED | `TestPhase36NativeCallCaptureOutcomeSeparation`. |
| 20 | Old C output and affected emitter inventory are independently checked. | ✓ VERIFIED | No-call byte baseline and emitter inventory controls; current `EmitNative` consumer registration is included. |
| 21 | Local/imported repeated `crc_step` helpers produce pinned answers on both engines. | ✓ VERIFIED | Helper answer/optimization tests and independent `helper-values.json` values. |
| 22 | Four type pairs, nested/computed arguments, arithmetic use, mutation, and Bool conditions have discriminating witnesses. | ✓ VERIFIED | Helper answer and type-pair roots; fixture values were independently computed, not compiler-produced. |
| 23 | Dynamic defects preserve exact reasons and exit 65 across native optimization lanes; bounds remain distinct. | ✓ VERIFIED | Fresh optimization matrix, legacy bounds, and caller-before-entry controls. |
| 24 | Public conformance retains its input contract and capture stays separate from process outcome. | ✓ VERIFIED | Public helper conformance and native capture outcome roots. |
| 25 | Source-free manifest states bounded arity-one contexts and all four type pairs. | ✓ VERIFIED | Surface manifest tests and independently authored witnesses. |
| 26 | Complete-program witnesses and finite source/census checks cover manifest claims. | ✓ VERIFIED | Surface witness/census tests and current census receipt inputs. |
| 27 | Admission, refusal, hint, restriction, and census controls are bidirectional; prior witness bytes stay frozen. | ✓ VERIFIED | Manifest call controls and surface limit-consumer tests. |
| 28 | Integrated no-call programs retain literal C bytes through frozen routes. | ✓ VERIFIED | Fresh `TestPhase36NoCallGeneratedCByteBaseline` root. |
| 29 | Acceptance evidence separates source, raw contract, numeric, causal, and bounds proofs. | ✓ VERIFIED | Acceptance evidence and distinct owner tests/receipts. |
| 30 | Current inspection, fresh evidence, and historical/hosted receipts are distinguished; phase ownership is preserved. | ✓ VERIFIED | Transition and acceptance artifacts; this report makes no hosted evidence claim. |
| 31 | Optional evidence capture exhaustion preserves application stdout, stderr, exit status, and later defects. | ✓ VERIFIED | Fresh native capture outcome roots check preserved process results and late defects; this claim is separate from peer or numeric proof. |
| 32 | Causal-peer acceptance does not prove scalar arithmetic or source argument/result intent. | ✓ VERIFIED | Canonical causal prohibition is green with fail-first proof; separate numeric and source-intent tests own those claims. |

**Score:** 32/32 truths verified; behavior-unverified: 0.

## Fresh Evidence and Scope

The current acceptance receipt covers 86/86 roots in 13 packages (14 commands, 34 plan-selector groups); the bounded regression receipt covers 252 actual root executions across 16 packages, 240 distinct test names and 251 unique package/name selectors. Both have zero missing, failing, or skipped roots. I independently matched receipt heads to the current HEAD, checked every listed post-input hash against current files, and matched raw log hashes and root results. The duplicate internal/external checker root explains 252 executions versus 251 selectors. No project/package-wide suite was run.

The test-tier prohibitions' four direct controls have the expected red/red/green/green exit results and matching logs. Canonical producers for both prohibitions independently return `status: green`, `flagged: false`, `kind: node-test`, and wired `failFirst` evidence. All 26 context decisions are honored (`decision_coverage` above).

Since the prior report, the relevant source delta is the dynamic scalar `OpDefect` terminal path in `cgen_program.go` and `TestPhase33ScalarExplicitDefectShells`. In the dynamic invocation shell used to capture native conformance events, `OpDefect` writes the structured terminal defect event/reason and then exits 0 so the wrapper can parse it. In the application shell, `OpDefect` follows the application-failure path, writes `schway app: <reason>` to stderr, and exits 65. I ran that exact named test with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, and a bounded timeout; it passed. The test checks interpreter event/reason and executionpeer validation, native `ExpectDefect` runs at O0/O3/O3-LTO for helper and entry defects, and the application shell's stderr/exit-65 result. This closes the Phase 36 dynamic native defect claim.

Phase 33 CR-03/04/06 remain deferred BLOCKER M007 audit obligations; they are neither fixed nor accepted here. CR-05 is the repaired dynamic scalar caller path. Legacy no-call C-byte parity is not evidence of legacy no-call native defect parity. These carried obligations are outside the Phase 36 helper-call contract. The current review has no new finding; WR-02 remains a separate deferred warning about an obsolete same-type validator comment.

## Required Artifacts

The canonical artifact query passed all 24 declarations across eight plans. The 19 semantic key links could not be resolved by the generic path-based query because their `from` fields are descriptive labels, not relative paths; I manually traced every link through implementations and named tests.

| Plan | Artifacts | Status | Link trace |
|---|---:|---|---|
| 36-01 | 3/3 | ✓ VERIFIED | Frozen C bytes and route inventory are read by comparator. |
| 36-02 | 2/2 | ✓ VERIFIED | Parser call AST reaches checker lowering and typed core. |
| 36-03 | 3/3 | ✓ VERIFIED | Core, origin, and path validators derive their own call constraints. |
| 36-04 | 2/2 | ✓ VERIFIED | Interpreter frames/events feed independent peer topology checks. |
| 36-05 | 2/2 | ✓ VERIFIED | Native emission and capture boundary tests exercise generated code. |
| 36-06 | 5/5 | ✓ VERIFIED | Local/imported sources reach interpreter/native and public runner. |
| 36-07 | 3/3 | ✓ VERIFIED | Independent witnesses and census are consumed by manifest controls. |
| 36-08 | 4/4 | ✓ VERIFIED | Acceptance evidence binds selectors; prohibition receipts are independently checked. |

| Plan | Links | Status | Trace |
|---|---:|---|---|
| 36-01 | 1/1 | ✓ WIRED | Baseline comparator reads literal bytes and checks emitter/application/module routes. |
| 36-02 | 2/2 | ✓ WIRED | Parser creates call AST; checker resolves signatures and lowers ordered typed calls. |
| 36-03 | 3/3 | ✓ WIRED | Independent validators consume core facts, not checker verdicts. |
| 36-04 | 3/3 | ✓ WIRED | Runtime ordinals, peer ancestry, and capture completeness are separately exercised. |
| 36-05 | 3/3 | ✓ WIRED | Native event/path bounds and application-outcome separation are tested. |
| 36-06 | 3/3 | ✓ WIRED | Local/imported calls reach both engines; defects and causal proof are distinct. |
| 36-07 | 2/2 | ✓ WIRED | Source-free manifest is checked against independent witnesses and census. |
| 36-08 | 2/2 | ✓ WIRED | Exact owner selectors and distinct proof claims are referenced by evidence. |

## Data Flow

No rendered UI or database value applies. Program values flow from local/imported source through parsing, typed checker lowering, and either interpreter frames or emitted C calls to captured output. `helper-values.json` pins independently computed expected values. The peer only derives call identity/path/ordinal; it does not establish arithmetic correctness.

| Artifact | Data | Source and flow | Status |
|---|---|---|---|
| Helper examples | U64/Bool inputs and returned helper values | Source → parser/checker → core → interpreter/native → captured output | ✓ FLOWING |
| Interpreter | Return value or callee defect | Dynamic frame executes callee and propagates result/reason | ✓ FLOWING |
| C generator/native app | Typed call result and process output/status | Generated invocation → actual process stdout/stderr/exit | ✓ FLOWING |
| Execution peer | Call identity and causal path | Runtime event sequence → independently derived path/ordinal | ✓ FLOWING |
| Surface manifest | Bounded source-free call contract | Manifest data → surface tests and CLI route | ✓ FLOWING |

## Behavioral Spot-Checks

Fresh receipts provide named behavioral roots for answer values, imported modules, defects, recursion/refusal boundaries, all four signature pairs, peer mutations, path/ordinal/capacity limits, capture separation, ByteView bounds, and literal no-call C parity. The direct post-receipt spot-check was:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase33ScalarExplicitDefectShells$' -count=1 -timeout=600s
```

Result: PASS (1.011s). This was a named selector only; no suite was run.

## Requirements Coverage

| Requirement | Source | Status | Evidence |
|---|---|---|---|
| CALL-01 | Plans 36-02 through 36-06 | ✓ SATISFIED | One-argument U64/Bool calls, local/imported use, same-engine result and defect behavior, recursion refusal. |
| CALL-02 | Plans 36-03 through 36-08 | ✓ SATISFIED | Independent call derivation by core/origin/execution peers, mutation controls, and bounded surface contract. |

The D-18 clarification scopes ByteView bounds separately: helper U64/Bool defects cover overflow/remainder; legacy ByteView callee bounds and caller-before-entry argument bounds retain their own tests. No phase-assigned requirement is orphaned.

## Probe Execution

N/A. No phase plan declares a probe and no conventional `scripts/*/tests/probe-*.sh` probe exists for this compiler phase.

## Anti-Patterns Found

| File | Pattern | Severity | Impact |
|---|---|---|---|
| — | No blocking debt markers, stubs, empty implementations, or console-only implementations found in phase-touched code. | — | A lexical `placeholder` match is a comment about a dead/unused field; it is not a stub. |

## Security and UI Gates

The current security artifact reports 24/24 authored Level 1 checks closed with zero open threats; no new trust boundary was introduced. UI review is N/A for this compiler-only phase. Current standard review reports no new issue.

## Human Verification Required

None. Specified runtime behaviors are covered by named tests and fresh bounded receipts. No visual, service, latency, hosted, Linux, or human-readability claim is made.

## Gaps Summary

No Phase 36 must-have gaps remain. The scalar helper path has independent source/core/origin/peer derivation, local and imported U64/Bool calls, and matching interpreter/native defect behavior with current evidence. Phase 33 M007 obligations remain explicitly deferred outside this phase's contract.

---

_Verified: 2026-10-16T04:09:33Z_
_Verifier: gsd-verifier_

---
phase: 36-scalar-helper-calls
verified: 2026-10-16T00:07:46Z
status: gaps_found
score: "32/32 must-haves verified"
covered_files:
  - ".planning/REQUIREMENTS.md"
  - ".planning/ROADMAP.md"
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
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH-PROHIBITIONS.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH-REGRESSION.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH-RUN.json"
  - ".planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs"
  - "cmd/schway/main_phase36_test.go"
  - "examples/helper_calls/crc_step.schway"
  - "examples/helper_calls/importer.schway"
  - "examples/helper_calls/public.schway"
  - "internal/compiler/cgen/cgen_phase36_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/check/check_phase36_scalar_calls_test.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/corevalidate/corevalidate_phase36_legacy_internal_test.go"
  - "internal/compiler/corevalidate/corevalidate_phase36_test.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase36_test.go"
  - "internal/compiler/interp/interp.go"
  - "internal/compiler/interp/interp_phase36_test.go"
  - "internal/compiler/native/native_phase36_test.go"
  - "internal/compiler/originvalidate/originvalidate.go"
  - "internal/compiler/originvalidate/originvalidate_phase36_test.go"
  - "internal/compiler/pathoracle/pathoracle.go"
  - "internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go"
  - "internal/compiler/pathoracle/pathoracle_phase36_test.go"
  - "internal/compiler/session/session_phase36_byteview_test.go"
  - "internal/compiler/session/session_phase36_calls_test.go"
  - "internal/compiler/session/session_phase36_peer_mutation_test.go"
  - "internal/compiler/session/session_phase36_surface_test.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase36_test.go"
  - "testdata/phase36/helper-values.json"
  - "testdata/phase36/surface-census.json"
  - "testdata/phase36/surface-witnesses.json"
covered_digest: "v3:sha256:31f5932a47bd7dcba04361735613ec7a6fad31d0a65c05fe89c2fc6fdbfe1806"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 26
  total: 26
  not_honored: []
gaps:
  - truth: "Phase-modified roadmap entries replace TBD plan counts for completed phases with the recorded completed counts."
    status: failed
    reason: "The modified roadmap still labels completed Phases 34, 35, 36, and 37 as Plans: TBD, despite its progress table recording 8/8, 8/8, 8/8, and 9/9. These are unreferenced debt markers under the Phase 36 anti-pattern gate."
    artifacts:
      - path: ".planning/ROADMAP.md"
        issue: "Lines 162, 207, 236, and 278 contain stale TBD plan counts for completed phases; the Phase 36 summary lists this roadmap as modified."
    missing:
      - "Replace or remove the four completed-phase TBD plan counts so the roadmap agrees with its completion table."
---

# Phase 36: Scalar Helper Calls Verification Report

**Phase Goal:** A scalar function can be factored into small helpers: it calls another scalar function with one U64 or Bool argument and uses the result, including calls into imported modules, with the same defect behavior on both engines and every peer re-deriving the call independently.

**Verified:** 2026-10-16T00:07:46Z  
**Status:** gaps_found  
**Verification mode:** Initial under Step 0; the prior report had no top-level `gaps:` section. Its claims and fingerprint were treated as historical context, not evidence.

## Goal Achievement

### Observable Truths

The four roadmap success criteria, all distinct plan-frontmatter truths, and both test-tier prohibitions produce 32 must-haves. Roadmap criteria were kept even where plans add detail.

| # | Truth | Status | Current evidence |
|---:|---|---|---|
| 1 | One-argument U64 and Bool scalar calls use their result locally and through imports; recursion and mutual recursion remain refused. | ✓ VERIFIED | `TestPhase36HelperAnswersAndDefects`, `TestPhase36HelperOptimizationMatrix`, and `TestPhase36CallImportAndRefusalBoundaries`; current parser, checker, and call-graph refusal paths inspected. |
| 2 | Callee defects preserve the callee reason and exit 65 on both engines, with the source-grounded ByteView bounds case kept distinct from U64/Bool helper defects. | ✓ VERIFIED | `TestPhase36HelperOptimizationMatrix`, `TestPhase36LegacyByteViewCalleeBounds`, caller-before-entry controls, and REQUIREMENTS.md's dated CALL clarification. U64/Bool helper cases cover overflow and remainder-by-zero; legacy ByteView owns callee bounds. |
| 3 | Check, corevalidate, originvalidate, and executionpeer independently derive calls; argument, result, and defect mutations are caught. | ✓ VERIFIED | `TestPhase36CallSourceIntentControls`, `TestPhase36RawCallShapeAndFreshness`, `TestPhase36OriginCallFreshnessAndTransfer`, `TestPhase36CallCausalAttributionControls`, and `TestPhase36EngineCallMutationControls`. |
| 4 | No-call programs retain literal generated-C bytes and the source-free manifest records arity one with no recursion. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline`, `TestPhase36SurfaceManifestCallWitnesses`, and the baseline ancestry check below. |
| 5 | Literal C baseline and controls predate compiler edits. | ✓ VERIFIED | Git ancestry places baseline `9b8285de` before parser implementation `fe174ce3`; `TestPhase36NoCallGeneratedCByteBaselineControls` passes. |
| 6 | Tracked pre-change programs have applicable emitter routes accounted for, including refusals. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline` checks the frozen source inventory and route outcomes. |
| 7 | Earlier golden bytes stay immutable and new fixtures have exact ledger rows. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaselineControls`; the new-fixture ledger checks helper examples and omission/false-exemption controls. |
| 8 | Direct one-argument calls parse, recover, and format within the existing 64-level bound. | ✓ VERIFIED | Current `parser.go` call parser and `parser_phase36_test.go`; bounded grammar/recovery tests are present and covered by the regression receipt. |
| 9 | All four U64/Bool signature pairs lower to typed fresh call results, including initialized Bool identifier tails. | ✓ VERIFIED | `TestPhase36ScalarCallLoweringAndTypePairs` and `TestPhase36HelperAnswersAndDefects`. |
| 10 | Arguments execute once in left-to-right order; unused-Byte and whole-program recursion refusals remain. | ✓ VERIFIED | `TestPhase36CallSourceIntentControls`, caller argument-boundary tests, and `TestPhase36CallImportAndRefusalBoundaries`; ordered lowering is explicit in `check/scalar.go`. |
| 11 | Raw core calls independently validate callee, signature, initialized argument, result type, and fresh target. | ✓ VERIFIED | `TestPhase36RawCallShapeAndFreshness` plus raw-call refusal-removal controls; `corevalidate` derives its own contract. |
| 12 | Origin analysis transfers calls as initialized, value-unknown scalars through branches and loops. | ✓ VERIFIED | `TestPhase36OriginCallFreshnessAndTransfer` and refusal-removal controls. |
| 13 | Scalar calls add no loan ancestry or resource/back-edge authority. | ✓ VERIFIED | `TestPhase36ScalarCallAuthorityBoundaries` and `TestPhase36ScalarCallLocalBorrowControls`; `pathoracle` treats calls as value boundaries. |
| 14 | Repeated calls receive contiguous per-parent/per-site ordinals, and child counters reset under a fresh parent. | ✓ VERIFIED | `TestPhase36DynamicCallOrdinalControls`, `TestPhase36InterpreterRepeatedCallTracer`, and `TestPhase36NativeDynamicInvocationPaths`. |
| 15 | Interpreter results/defects and execution-peer causal attribution remain separate claims. | ✓ VERIFIED | `TestPhase36InterpreterCallAnswersAndDefects` and `TestPhase36CallCausalAttributionControls`; causal acceptance does not assert arithmetic. |
| 16 | Evidence retention, path/counter bounds, and exhaustion remain distinct from numeric work. | ✓ VERIFIED | `TestPhase36InterpreterEvidenceBoundaries`, `TestPhase36CallEventCapacityBoundary`, and `TestPhase36NativeCallCaptureOutcomeSeparation`. |
| 17 | Native calls use sequenced typed arguments/results and stable dynamic paths. | ✓ VERIFIED | `TestPhase36NativeDynamicInvocationPaths`, emitter inventory controls, and engine mutation coverage for O0/O3/O3-LTO. |
| 18 | Repeated-call captures and exact event/path/ordinal boundaries have named controls. | ✓ VERIFIED | `TestPhase36NativeDynamicInvocationPaths`, event capacity controls, and native ordinal/path boundary controls. |
| 19 | Optional capture exhaustion preserves application stdout, stderr, exit, and later defects. | ✓ VERIFIED | `TestPhase36NativeCallCaptureOutcomeSeparation` separately checks successful work and late-defect outcomes under disabled and exhausted capture. |
| 20 | Old generated-C output and affected emitter inventories are independently checked. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline` and emitter call inventory controls. |
| 21 | Local and imported repeated crc_step helpers produce pinned answers on both engines. | ✓ VERIFIED | `TestPhase36HelperAnswersAndDefects`, `TestPhase36HelperOptimizationMatrix`, `crc_step.schway`, and `importer.schway`. |
| 22 | Four signature pairs, nested/computed arguments, arithmetic use, mutable assignments, and Bool conditions have discriminating witnesses. | ✓ VERIFIED | Helper answer tests and type-pair tests; `helper-values.json` identifies independently calculated expected values. |
| 23 | Dynamic defects retain exact reasons and exit 65 across native optimization lanes; bounds obligations stay distinct. | ✓ VERIFIED | `TestPhase36HelperOptimizationMatrix`, legacy ByteView callee bounds, and caller-before-entry controls. |
| 24 | Public conformance retains its input contract and native capture stays separate from process outcome. | ✓ VERIFIED | `TestPhase36PublicHelperConformance` and `TestPhase36NativeCallCaptureOutcomeSeparation`. |
| 25 | Source-free manifest states bounded arity-one contexts and all four U64/Bool pairs. | ✓ VERIFIED | `TestPhase36SurfaceManifestCallWitnesses`; independently authored `surface-witnesses.json`. |
| 26 | Complete-program witnesses and finite source/census checks cover manifest claims. | ✓ VERIFIED | `TestPhase36SurfaceManifestCallWitnesses` and `TestPhase36SurfaceManifestCallCensus`. |
| 27 | Missing/false admission, refusal, hint, restriction, and census controls are bidirectional while prior witness bytes stay frozen. | ✓ VERIFIED | `TestPhase36SurfaceManifestCallControls` and `TestPhase36SurfaceLimitConsumers`. |
| 28 | Integrated no-call programs retain literal generated-C bytes through frozen routes. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline`; the exact baseline selector passed during this verification. |
| 29 | Acceptance evidence separates source, raw-contract, numeric, causal, and both bounds proofs. | ✓ VERIFIED | `36-ACCEPTANCE-EVIDENCE.md` and the source/peer/engine/bounds tests cited above. |
| 30 | Current inspection, fresh evidence, and historical/hosted receipts are labeled separately; later-phase ownership is preserved. | ✓ VERIFIED | `36-ACCEPTANCE-EVIDENCE.md`, `36-CONTEXT.md`, and `36-TRANSITION-CHECK.md`; this verification makes no hosted-evidence claim. |
| 31 | Optional evidence exhaustion cannot turn success into application success or hide later application output/defect/exit. | ✓ VERIFIED | Independently rerun canonical outcome-prohibition producer is green/unflagged with `failFirst=true`, `passed=true`, and `failFirstProof=violation-fixture`; violation, clean, and default controls passed. |
| 32 | Causal-peer acceptance does not prove scalar arithmetic or source argument/result intent. | ✓ VERIFIED | Independently rerun causal-prohibition producer has the same green fail-first disposition; numeric-limit mutation fails the named peer assertion, while source and arithmetic claims remain separately tested. |

**Score:** 32/32 must-haves verified. The implementation goal is achieved. Overall status remains `gaps_found` because the Step 7 debt-marker gate found stale completed-phase plan counts in a Phase 36-modified roadmap file.

### Prohibition Gate Assessment

Both Plan 08 test-tier prohibitions were independently checked through the canonical producer. I did not rely on caller-supplied `failFirst` fields: the producer invoked the closed-subject adapter, required the exact source assertions, exercised the violation fixture, then checked clean and default controls. The adapter uses shell-free child execution, exact test roots, 20-second child bounds, and fails closed on skips or infrastructure errors. No compiler implementation, fixture, or dependency changed for this proof bridge.

| Prohibition | Current proof | Status |
|---|---|---|
| Do not let optional evidence exhaustion turn success into application success or suppress application output, exit, or a later defect. | Outcome violation fixture kills both named capture assertions; clean/default controls pass. Canonical producer returns green, unflagged, fail-first proof from the violation fixture. | ✓ VERIFIED |
| Do not claim causal execution-peer acceptance proves arithmetic or source argument/result intent. | Numeric-limit mutation kills the named peer assertion; clean/default controls pass. Numeric and source-intent claims are tested by their owning peers. | ✓ VERIFIED |

Exact independent commands:

```sh
node /Users/jon/.codex/gsd-core/bin/gsd-tools.cjs check prohibition-enforcement /tmp/schway36-prohibition-0.json --raw
node /Users/jon/.codex/gsd-core/bin/gsd-tools.cjs check prohibition-enforcement /tmp/schway36-prohibition-1.json --raw
```

Both returned exit 0 and `status=green`, `flagged=false`, `tier=test`, `kind=node-test`, `failFirst=true`, `passed=true`, and `failFirstProof=violation-fixture`. These are local proof results, not hosted evidence.

## Required Artifacts

The canonical `verify.artifacts` query passed all 24 declared artifacts across eight plans. Each artifact exists and has no stub issue. The plan-declared paths are grouped below; source and test wiring was inspected separately.

| Plan | Artifacts | Status | Wiring/details |
|---|---|---|---|
| 36-01 | `testdata/phase36/prechange-corpus.json`; `internal/compiler/session/session_phase36_prechange_test.go`; `testdata/phase36/new-fixture-ledger.json` | ✓ VERIFIED (3/3) | Comparator consumes frozen bytes and route inventory. |
| 36-02 | `internal/compiler/syntax/parser_phase36_test.go`; `internal/compiler/check/check_phase36_scalar_calls_test.go` | ✓ VERIFIED (2/2) | Parser call AST reaches checker lowering and typed core. |
| 36-03 | `internal/compiler/corevalidate/corevalidate_phase36_test.go`; `internal/compiler/originvalidate/originvalidate_phase36_test.go`; `internal/compiler/pathoracle/pathoracle_phase36_test.go` | ✓ VERIFIED (3/3) | Independent analyses consume core facts without trusting checker verdicts. |
| 36-04 | `internal/compiler/interp/interp_phase36_test.go`; `internal/compiler/executionpeer/executionpeer_phase36_test.go` | ✓ VERIFIED (2/2) | Runtime frames/events and independent peer ancestry are exercised. |
| 36-05 | `internal/compiler/cgen/cgen_phase36_test.go`; `internal/compiler/native/native_phase36_test.go` | ✓ VERIFIED (2/2) | C emitter and native capture boundary are connected. |
| 36-06 | `examples/helper_calls/crc_step.schway`; `examples/helper_calls/importer.schway`; `examples/helper_calls/public.schway`; `testdata/phase36/helper-values.json`; `internal/compiler/session/session_phase36_calls_test.go` | ✓ VERIFIED (5/5) | Programs reach interpreter, native, and public conformance routes. |
| 36-07 | `testdata/phase36/surface-witnesses.json`; `testdata/phase36/surface-census.json`; `internal/compiler/session/session_phase36_surface_test.go` | ✓ VERIFIED (3/3) | Independent witnesses and census are consumed by manifest controls. |
| 36-08 | `36-ACCEPTANCE-EVIDENCE.md`; `36-PROHIBITION-EVIDENCE.md`; `.planning/PRODUCT-ROADMAP.md`; `.planning/LANGUAGE-MATURITY.md` | ✓ VERIFIED (4/4) | Evidence and living planning records exist; the current producer results were separately rerun. |

The prohibition adapter and request files also exist and are substantive; the adapter is included in the fresh evidence fingerprint.

## Key Link Verification

`verify.key-links` could not resolve the plans' semantic `from` labels as file paths, so its negative result was not treated as proof that links are broken. I traced every declared connection through implementation and named tests. All 19 of 19 links are wired.

| Plan | Links | Evidence |
|---|---:|---|
| 36-01 | 1/1 | Comparator reads frozen source bytes and checks current emitter/application/module routes. |
| 36-02 | 2/2 | Parser emits call AST; checker resolves the callee/signature and lowers ordered typed OpCalls with fresh result targets. |
| 36-03 | 3/3 | Corevalidator, originvalidator, and pathoracle independently derive their call contracts from core facts. |
| 36-04 | 3/3 | Interpreter creates per-parent ordinals; executionpeer derives its own topology; collector completeness stays separate from program outcome. |
| 36-05 | 3/3 | Native emitter reserves bounded event/path capacity; capacity failure stays separate from application exit; old routes compare literal bytes. |
| 36-06 | 3/3 | Local and imported source reach both runners; defects and causal capture are distinct; public route keeps its input contract. |
| 36-07 | 2/2 | Source-free manifest is checked against independent witnesses and finite route/census consumers. |
| 36-08 | 2/2 | Acceptance evidence cites exact owner selectors and separates proof types. |

## Data-Flow Trace (Level 4)

This phase has no rendered UI or database-backed values. Its program data-flow does reach user-visible output: source calls parse into call expressions, checker lowering resolves an imported or local function and emits a typed core call, then either the interpreter evaluates that call or cgen emits a typed C invocation. Both paths return the result to assignment/expression use and the public runner captures stdout, stderr, and exit. Independent fixture values are documented as Python integer arithmetic and were not generated by the Schway compiler or either engine.

| Artifact | Data | Source and flow | Real data? | Status |
|---|---|---|---|---|
| `examples/helper_calls/crc_step.schway` and `importer.schway` | U64/Bool inputs and helper results | Source parser → checker → core call → interpreter/native execution → captured output | Yes; example source and independently computed expected values | ✓ FLOWING |
| `internal/compiler/interp/interp.go` | Callee return or defect | Runtime call frames execute the callee and return its actual result/reason to caller | Yes; execution result | ✓ FLOWING |
| `internal/compiler/cgen/cgen_program.go` | Typed call result | Emits a dynamic native invocation and records its own call path/outcome | Yes; generated program result | ✓ FLOWING |
| `internal/compiler/executionpeer/executionpeer.go` | Call identity and causal path | Independent call events and ordinals; does not invent numeric result proof | Yes; runtime event stream | ✓ FLOWING |

## Behavioral Spot-Checks

All checks below are exact named tests; no full workspace or package suite was run. Each ran with `GOCACHE=/tmp/ai-lang-verification-gocache`. Most completed in under 2.5 seconds. The isolated engine mutation control took 23.247 seconds because it compiles the native optimization matrix; a concurrent attempt hit the 20-second harness bound without an assertion failure, then the same named test passed in isolation under a 35-second outer bound.

| Behavior | Named check | Result | Status |
|---|---|---|---|
| Parser/checker calls, imports/refusals, type pairs, and source-intent mutation | `TestPhase36HelperOptimizationMatrix`; `TestPhase36CallSourceIntentControls`; `TestPhase36CallImportAndRefusalBoundaries`; `TestPhase36ScalarCallLoweringAndTypePairs` | PASS; typed local/imported calls and refusal/mutation controls exercised. | ✓ PASS |
| Independent raw-core, origin, and causal peers | `TestPhase36RawCallShapeAndFreshness`; `TestPhase36OriginCallFreshnessAndTransfer`; `TestPhase36CallCausalAttributionControls` | PASS; callee/signature/freshness and independent causal controls exercised. | ✓ PASS |
| Engine defects and mutation controls | `TestPhase36HelperAnswersAndDefects`; `TestPhase36EngineCallMutationControls`; `TestPhase36LegacyByteViewCalleeBounds` | PASS; exact output/reason/exit behavior and native optimization lanes checked. | ✓ PASS |
| No-call bytes and native capture outcome | `TestPhase36NoCallGeneratedCByteBaseline`; `TestPhase36NoCallGeneratedCByteBaselineControls`; `TestPhase36NativeCallCaptureOutcomeSeparation` | PASS; frozen bytes and process/capture outcomes remain distinct. | ✓ PASS |
| Runtime invocation and source-free manifest | `TestPhase36InterpreterCallAnswersAndDefects`; `TestPhase36NativeDynamicInvocationPaths`; `TestPhase36SurfaceManifestCallWitnesses`; `TestPhase36SurfaceManifestCallCensus`; `TestPhase36SurfaceManifestCallControls` | PASS; runtime answers and bounded manifest claims are exercised. | ✓ PASS |
| Public route and transition evidence | `TestPhase36PublicHelperConformance`; `TestPhase37Phase36AdditionLedgerControls` | PASS; public helper contract and Phase 36 ledger guard remain active. | ✓ PASS |

Exact named-test commands:

```sh
source /tmp/schway34-root-pin.sh
source /tmp/schway34-gsd-run.sh
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36HelperOptimizationMatrix$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase36CallSourceIntentControls$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^TestPhase36RawCallShapeAndFreshness$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^TestPhase36OriginCallFreshnessAndTransfer$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -timeout=20s
gsd_run run-with-timeout 35 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36EngineCallMutationControls$' -count=1 -timeout=600s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NativeCallCaptureOutcomeSeparation$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36SurfaceManifestCallWitnesses$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase36CallImportAndRefusalBoundaries$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36HelperAnswersAndDefects$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36LegacyByteViewCalleeBounds$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaselineControls$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase36ScalarCallLoweringAndTypePairs$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^TestPhase36InterpreterCallAnswersAndDefects$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase36NativeDynamicInvocationPaths$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36SurfaceManifestCallCensus$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36SurfaceManifestCallControls$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase36PublicHelperConformance$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37Phase36AdditionLedgerControls$' -count=1 -timeout=20s
```

### Freshness and Regression Receipts

I inspected the receipt hashes and raw logs rather than relying on the receipt summary text. In `36-REFRESH-RUN.json`, all 82 bound inputs and all 13 Go-log hashes match; its 19 Go commands covered 84 exact roots across 13 packages, with no missing, failed, or skipped roots. The two historical Plan 08 Python capture producers were represented through extracted owner lists only; neither was replayed and no archived capture was appended.

In `36-REFRESH-REGRESSION.json`, all 172 bound inputs and 16 raw-log hashes match; logs contain 167 distinct test names and 179 root executions, all passing, with no missing/fail/skip. The receipt is bound to source revision `534fac4d0ec9e2ff6ade8da20a4627b888f12792`. The relevant source, tests, examples, fixtures, and dependency files are unchanged from that revision through current HEAD `d52cc00726221401f3a7688d8af0567aeeed6b9f`. These historical receipts were inspected, not replayed.

The current 36-REFRESH-RUN owner set binds 82 inputs to source revision `3d647e44554cdce5a12525b81da1add1ff2f8f1e`; the read-only grounding record preserves the historical receipt revision separately from current owners. Production code, fixtures, and dependencies were unchanged throughout this verification.

## Verification Report Validation

After writing this report, both required groundedness guards passed, and the directly applicable Phase 36 baseline, manifest-control, and Phase 37 addition-ledger guards passed. The status readback below is the canonical status-owner result.

```sh
source /tmp/schway34-root-pin.sh
source /tmp/schway34-gsd-run.sh
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestVerificationGroundednessFrontierIsPinned$' -count=1 -timeout=60s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1 -timeout=60s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaselineControls$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36SurfaceManifestCallControls$' -count=1 -timeout=20s
gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37Phase36AdditionLedgerControls$' -count=1 -timeout=20s
gsd_run query verification.status .planning/phases/36-scalar-helper-calls --pick status
```

The test commands all passed; the status owner returned `gaps_found` with exit 0.

## Probe Execution

Not applicable. Phase 36 is a language feature phase, not a migration/tooling phase, and its plan/success criteria declare no probes.

## Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| CALL-01 | 36-01 through 36-08 | One-argument U64/Bool helper calls, imported calls, result use, recursion refusal, and callee defects on both engines. The dated clarification scopes helper defects to overflow/remainder and assigns ByteView bounds to the legacy callee regression plus a separate caller-before-entry control. | ✓ SATISFIED | Truths 1–2, 8–10, and 21–24; exact local/imported outputs and defect reasons/exits passed. |
| CALL-02 | 36-01 through 36-08 | Independent call derivation, argument/result/defect mutant detection, and no-call generated-C compatibility. | ✓ SATISFIED | Truths 3–4, 11–20, and 25–28; no-call baseline and all relevant peer controls passed. |

Both requirements are checked complete in REQUIREMENTS.md and mapped to Phase 36. No phase-mapped requirement is orphaned.

## Decision Coverage

The canonical decision-coverage query reports 26/26 honored, with no unhonored CONTEXT decisions. The 26 decisions were read from `36-CONTEXT.md`; the implementation and corresponding tests uphold the selected bounds, separation of proof types, and ownership boundaries.

## Anti-Patterns and Review Findings

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `.planning/ROADMAP.md` | 162 | `Plans: TBD` for completed Phase 34 | 🛑 BLOCKER | Progress table says 8/8 Complete; stale marker is not tied to follow-up work. |
| `.planning/ROADMAP.md` | 207 | `Plans: TBD` for completed Phase 35 | 🛑 BLOCKER | Progress table says 8/8 Complete; stale marker is not tied to follow-up work. |
| `.planning/ROADMAP.md` | 236 | `Plans: TBD` for completed Phase 36 | 🛑 BLOCKER | Progress table says 8/8 Complete; stale marker is not tied to follow-up work. |
| `.planning/ROADMAP.md` | 278 | `Plans: TBD` for completed Phase 37 | 🛑 BLOCKER | Progress table says 9/9 Complete; stale marker is not tied to follow-up work. |
| `internal/compiler/corevalidate/corevalidate.go` | 4248 | WR-02 stale explanatory comment | ⚠️ WARNING | Comment describes obsolete same-type-only call validation; current implementation derives from `sourceTypeID` and separate contracts correctly. Tests and cross-typed controls support runtime behavior. Review disposition defers the comment to the core-validator maintainer's next semantic edit; it is not fixed and review is not clean. |

The four roadmap markers are grouped as one planning-metadata gap in frontmatter. The later Phase 38 `TBD` and `0/TBD` entries describe not-started work and are not treated as completed-phase debt. The Phase 36 summaries name `.planning/ROADMAP.md` as modified, which brings its stale completed-phase markers into the explicit debt-marker scan. The roadmap's progress table currently gives the counts that should replace the stale markers. No later phase explicitly owns that correction, so it is not deferred.

No implementation TBD/FIXME/XXX debt markers, functional stubs, empty handlers, or hardcoded empty values flowing to runtime output were found in the implementation paths reviewed. The two plan prohibitions passed their canonical producer controls. Security review records 24/24 authored high/L1 threats closed and zero open; this is not an L2/L3 claim. The UI applicability gate reports no frontend files or UI spec, so no visual review is needed.

The current review disposition records WR-01 fixed and WR-02 deferred, with zero open of two total. The warning is documentation-only and does not contradict the observable call truths; it remains visible for maintainers. `36-REFRESH-GATES.md` records that changed review/disposition bytes make earlier Phase 33 and 35 fingerprints stale, while Phase 34 remains passed; those prior reports/digests were not modified.

## Human Verification Required

None. The goal's output, defects, refusals, and independent-peer invariants have named deterministic tests. No UI, hosted, or external-service behavior is part of this phase acceptance.

## Gaps Summary

The codebase achieves the Phase 36 scalar helper-call goal: local and imported U64/Bool helper calls return used values, interpreter/native defect behavior matches, call derivation is independently checked, and no-call generated C remains byte-identical. The two test-tier prohibitions also pass independent canonical enforcement.

The phase gate remains `gaps_found` because the Phase 36-modified `.planning/ROADMAP.md` leaves plan-count fields as `TBD` for completed Phases 34–37 despite recording exact completion counts in the progress table. Fix those four planning entries to clear the debt-marker blocker. No implementation gap or human verification item remains.

---

_Verified: 2026-10-16T00:07:46Z_  
_Verifier: the agent (gsd-verifier)_

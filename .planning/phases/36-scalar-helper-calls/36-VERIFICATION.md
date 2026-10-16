---
phase: 36-scalar-helper-calls
verified: 2026-10-16T02:06:09Z
status: gaps_found
score: "32/32 must-haves verified"
covered_files:
  - ".planning/LANGUAGE-MATURITY.md"
  - ".planning/PRODUCT-ROADMAP.md"
  - ".planning/REQUIREMENTS.md"
  - ".planning/ROADMAP.md"
  - ".planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-REFRESH2-FIX.md"
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
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-FIX.md"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-GATES.md"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-PROHIBITIONS.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-REGRESSION.json"
  - ".planning/phases/36-scalar-helper-calls/36-REFRESH2-RUN.json"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-DISPOSITION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-SUPPLEMENT.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-SECURITY.md"
  - ".planning/phases/36-scalar-helper-calls/36-UI-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-VALIDATION.md"
  - ".planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs"
  - ".planning/phases/36-scalar-helper-calls/prohibition-causal-violation.json"
  - ".planning/phases/36-scalar-helper-calls/prohibition-clean.json"
  - ".planning/phases/36-scalar-helper-calls/prohibition-outcome-violation.json"
  - "cmd/schway/main_phase36_test.go"
  - "internal/compiler/ast/ast.go"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_phase36_caller_shapes_test.go"
  - "internal/compiler/cgen/cgen_phase36_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/check/check.go"
  - "internal/compiler/check/check_phase34_refusal_test.go"
  - "internal/compiler/check/check_phase36_scalar_calls_test.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/core/core.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/corevalidate/corevalidate_phase36_legacy_internal_test.go"
  - "internal/compiler/corevalidate/corevalidate_phase36_test.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase36_test.go"
  - "internal/compiler/interp/interp.go"
  - "internal/compiler/interp/interp_phase36_test.go"
  - "internal/compiler/native/native.go"
  - "internal/compiler/native/native_phase36_test.go"
  - "internal/compiler/originvalidate/originvalidate.go"
  - "internal/compiler/originvalidate/originvalidate_phase36_test.go"
  - "internal/compiler/pathoracle/pathoracle.go"
  - "internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go"
  - "internal/compiler/pathoracle/pathoracle_phase36_test.go"
  - "internal/compiler/session/session_phase16_emitter_inventory_test.go"
  - "internal/compiler/session/session_phase34_surface_test.go"
  - "internal/compiler/session/session_phase35_surface_test.go"
  - "internal/compiler/session/session_phase36_byteview_test.go"
  - "internal/compiler/session/session_phase36_calls_test.go"
  - "internal/compiler/session/session_phase36_peer_mutation_test.go"
  - "internal/compiler/session/session_phase36_prechange_test.go"
  - "internal/compiler/session/session_phase36_surface_test.go"
  - "internal/compiler/session/verification_groundedness_test.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase36_test.go"
  - "testdata/phase16/public-emitter-consumers.json"
  - "testdata/phase36/helper-values.json"
  - "testdata/phase36/new-fixture-ledger.json"
  - "testdata/phase36/prechange-corpus.json"
  - "testdata/phase36/surface-census.json"
  - "testdata/phase36/surface-witnesses.json"
covered_digest: "v3:sha256:212690faea1ee94bcc41805a1a31856ebe6423fb4d9835e33aeea22fc28ac4cc"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 26
  total: 26
  not_honored: []
gaps:
  - truth: "Phase-modified roadmap has no unreferenced TBD/FIXME/XXX debt markers."
    status: failed
    reason: "The Phase 36 summary lists .planning/ROADMAP.md as modified, and that file contains two TBD markers at lines 317 and 330. They explicitly describe unstarted Phase 38 work, but neither marker line carries an issue, PR, or DEF-* follow-up reference required by the debt-marker gate."
    artifacts:
      - path: ".planning/ROADMAP.md"
        issue: "Unreferenced TBD markers remain in the phase-modified roadmap file."
    missing:
      - "Replace or formally track both Phase 38 plan-count placeholders so the modified roadmap contains no unresolved debt markers."
---

# Phase 36: Scalar Helper Calls Verification Report

**Phase Goal:** A scalar function can be factored into small helpers: it calls another scalar function with one U64 or Bool argument and uses the result, including calls into imported modules, with the same defect behavior on both engines and every peer re-deriving the call independently.

**Verified:** 2026-10-16T02:06:09Z
**Status:** gaps_found
**Verification mode:** Initial verification. The prior report has no top-level `gaps:` section, so its historical re-verification metadata did not put this refresh into gap-closure mode.

## Goal Achievement

The four roadmap success criteria, all 26 distinct plan-frontmatter truths, and both Plan 08 test-tier prohibitions produce 32 must-haves. All 32 are verified. The overall status is `gaps_found` only because Step 7's debt-marker gate also applies to the phase-modified roadmap.

| # | Truth | Status | Evidence |
|---:|---|---|---|
| 1 | One-argument U64 and Bool scalar calls use their result locally and through imports; recursion and mutual recursion remain refused. | ✓ VERIFIED | `TestPhase36HelperAnswersAndDefects`, `TestPhase36HelperOptimizationMatrix`, and `TestPhase36CallImportAndRefusalBoundaries`; current parser, checker, and call-graph refusal paths inspected. |
| 2 | Callee defects preserve the callee reason and exit 65 on both engines, while the legacy ByteView bounds case remains distinct from U64/Bool helper defects. | ✓ VERIFIED | `TestPhase36HelperOptimizationMatrix`, the legacy ByteView callee test, and separate caller-before-entry controls; requirements clarification D-18 keeps bounds outside the new U64/Bool value-call contract. |
| 3 | Check, corevalidate, originvalidate, and executionpeer independently derive calls; argument, result, and defect mutations are caught. | ✓ VERIFIED | `TestPhase36CallSourceIntentControls`, `TestPhase36RawCallShapeAndFreshness`, `TestPhase36OriginCallFreshnessAndTransfer`, and `TestPhase36CallCausalAttributionControls`. |
| 4 | No-call programs retain literal generated-C bytes and the source-free manifest records arity one with no recursion. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline`, `TestPhase36SurfaceManifestCallWitnesses`, and baseline ancestry controls. |
| 5 | Literal C baseline and controls predate compiler edits. | ✓ VERIFIED | Git ancestry places baseline `9b8285de` before parser implementation `fe174ce3`; `TestPhase36NoCallGeneratedCByteBaselineControls` passes. |
| 6 | Tracked pre-change programs have applicable emitter routes accounted for, including refusals. | ✓ VERIFIED | Frozen source inventory and route outcomes are checked by `TestPhase36NoCallGeneratedCByteBaseline`. |
| 7 | Earlier golden bytes stay immutable and new fixtures have exact ledger rows. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaselineControls`; the new-fixture ledger has omission and false-exemption controls. |
| 8 | Direct one-argument calls parse, recover, and format within the existing 64-level bound. | ✓ VERIFIED | `parser.go` and parser phase tests contain the bounded grammar and recovery behavior; current regression logs include their exact roots. |
| 9 | All four U64/Bool signature pairs lower to typed fresh call results, including initialized Bool identifier tails. | ✓ VERIFIED | `TestPhase36ScalarCallLoweringAndTypePairs` and helper answer tests. |
| 10 | Arguments execute once in left-to-right order; unused-Byte and whole-program recursion refusals remain. | ✓ VERIFIED | `TestPhase36CallSourceIntentControls`, caller argument-boundary controls, and `TestPhase36CallImportAndRefusalBoundaries`; ordered lowering is explicit in `check/scalar.go`. |
| 11 | Raw core calls independently validate callee, signature, initialized argument, result type, and fresh target. | ✓ VERIFIED | `TestPhase36RawCallShapeAndFreshness` plus refusal-removal controls; corevalidate derives its own contract. |
| 12 | Origin analysis transfers calls as initialized, value-unknown scalars through branches and loops. | ✓ VERIFIED | `TestPhase36OriginCallFreshnessAndTransfer` and refusal-removal controls. |
| 13 | Scalar calls add no loan ancestry or resource/back-edge authority. | ✓ VERIFIED | `TestPhase36ScalarCallAuthorityBoundaries` and `TestPhase36ScalarCallLocalBorrowControls`; pathoracle treats calls as value boundaries. |
| 14 | Repeated calls receive contiguous per-parent/per-site ordinals, and child counters reset under a fresh parent. | ✓ VERIFIED | Interpreter repeated-call tracer, dynamic ordinal controls, and native dynamic invocation path tests. |
| 15 | Interpreter results/defects and execution-peer causal attribution remain separate claims. | ✓ VERIFIED | Interpreter answer/defect tests and `TestPhase36CallCausalAttributionControls`; peer acceptance does not assert arithmetic. |
| 16 | Evidence retention, path/counter bounds, and exhaustion remain distinct from numeric work. | ✓ VERIFIED | Interpreter evidence-boundary, event-capacity, and native capture-outcome tests. |
| 17 | Native calls use sequenced typed arguments/results and stable dynamic paths. | ✓ VERIFIED | Native invocation tests, emitter inventory controls, and optimization-lane mutation coverage. |
| 18 | Repeated-call captures and exact event/path/ordinal boundaries have named controls. | ✓ VERIFIED | Native dynamic invocation, capacity, ordinal, and path-boundary controls. |
| 19 | Optional capture exhaustion preserves application stdout, stderr, exit, and later defects. | ✓ VERIFIED | `TestPhase36NativeCallCaptureOutcomeSeparation` checks successful work and late defects under disabled and exhausted capture. |
| 20 | Old generated-C output and affected emitter inventories are independently checked. | ✓ VERIFIED | No-call byte baseline and emitter call inventory controls; current inventory contains the newly registered Phase 36 emitter-test consumer. |
| 21 | Local and imported repeated crc_step helpers produce pinned answers on both engines. | ✓ VERIFIED | Helper answer and optimization matrix tests, `crc_step.schway`, `importer.schway`, and independently computed `helper-values.json`. |
| 22 | Four signature pairs, nested/computed arguments, arithmetic use, mutable assignments, and Bool conditions have discriminating witnesses. | ✓ VERIFIED | Helper answer and type-pair tests; expected values are independently recorded, not produced by Schway. |
| 23 | Dynamic defects retain exact reasons and exit 65 across native optimization lanes; bounds obligations stay distinct. | ✓ VERIFIED | Optimization matrix, legacy ByteView callee bounds, and caller-before-entry controls. |
| 24 | Public conformance retains its input contract and native capture stays separate from process outcome. | ✓ VERIFIED | `TestPhase36PublicHelperConformance` and native capture outcome tests. |
| 25 | Source-free manifest states bounded arity-one contexts and all four U64/Bool pairs. | ✓ VERIFIED | `TestPhase36SurfaceManifestCallWitnesses` and independently authored `surface-witnesses.json`. |
| 26 | Complete-program witnesses and finite source/census checks cover manifest claims. | ✓ VERIFIED | Manifest witness and census tests, with current census inputs bound by the regression receipt. |
| 27 | Missing/false admission, refusal, hint, restriction, and census controls are bidirectional while prior witness bytes stay frozen. | ✓ VERIFIED | Manifest call controls and surface limit consumer tests. |
| 28 | Integrated no-call programs retain literal generated-C bytes through frozen routes. | ✓ VERIFIED | `TestPhase36NoCallGeneratedCByteBaseline` passed in this verification. |
| 29 | Acceptance evidence separates source, raw-contract, numeric, causal, and both bounds proofs. | ✓ VERIFIED | `36-ACCEPTANCE-EVIDENCE.md` and the separately owned source, peer, engine, and bounds tests. |
| 30 | Current inspection, fresh evidence, and historical/hosted receipts are labeled separately; later-phase ownership is preserved. | ✓ VERIFIED | Acceptance evidence, context, and transition records. This report makes no hosted-evidence claim. |
| 31 | Optional evidence exhaustion cannot turn success into application success or hide later output, defects, or exit status. | ✓ VERIFIED | Current canonical prohibition request returns green/unflagged with `failFirst=true`, `passed=true`, and `failFirstProof=violation-fixture`; violation, clean, and default controls are bound. |
| 32 | Causal-peer acceptance does not prove scalar arithmetic or source argument/result intent. | ✓ VERIFIED | Current canonical causal-prohibition request has the same green fail-first disposition; numeric mutation and source-intent claims have separate test owners. |

**Score:** 32/32 must-haves verified; behavior-unverified: 0.

## Required Artifacts

The canonical artifact query passed all 24 declared artifacts across eight plans: each exists and has no stub issue. Wiring was traced separately.

| Plan | Artifacts | Status | Wiring |
|---|---|---|---|
| 36-01 | Pre-change corpus, literal-byte comparator, and new-fixture ledger | ✓ VERIFIED (3/3) | Comparator consumes frozen bytes and route inventory. |
| 36-02 | Parser and scalar-call checker tests | ✓ VERIFIED (2/2) | Parser call AST reaches checker lowering and typed core. |
| 36-03 | Core, origin, and path validator tests | ✓ VERIFIED (3/3) | Independent analyses consume core facts without trusting the checker verdict. |
| 36-04 | Interpreter and execution-peer tests | ✓ VERIFIED (2/2) | Runtime frames/events and independent peer ancestry are exercised. |
| 36-05 | C generator and native tests | ✓ VERIFIED (2/2) | C emission and native capture boundaries are connected. |
| 36-06 | Local/imported helper programs, expected values, and runner tests | ✓ VERIFIED (5/5) | Programs reach interpreter, native, and public-conformance routes. |
| 36-07 | Surface census, witnesses, and consumer tests | ✓ VERIFIED (3/3) | Independent witnesses and census are consumed by manifest controls. |
| 36-08 | Acceptance/prohibition evidence and living planning records | ✓ VERIFIED (4/4) | Evidence sources exist; both prohibition producers were independently checked in this verification. |

The cgen repair is narrowly scoped: `isDynamicScalarFunctionBody` admits the existing unused-Byte-to-U64 caller and ByteView-to-U64/Unit caller forms; the Byte formal must be unused across the six reference fields, and accepted helper callee signatures remain U64/Bool. `TestPhase36UnusedByteStructuredHelper` returns 11 in interpreter/native O0/O3/O3-LTO and validates its causal peer. `TestPhase36ByteViewUnitHelperEmission` compiles and links an uncalled Unit-returning body, then runs only an unrelated U64 entry that prints 7. It does not prove execution of that Unit body or a ByteView value-call entry. The repair test failed both exact emission assertions over the pre-fix source overlay and passes on the repaired source.

The Phase 16 emitter-consumer inventory initially found 116 source callsites against 115 registered consumers. Commit `0e56a3be` adds the new `EmitNative` consumer row; the inventory and its omission/stale/duplicate/invalid/missing controls pass on the final registry. The initial red is retained in the gate record and is not counted as a pass.

## Key Link Verification

The generic `verify.key-links` query could not resolve plan `from` labels because they are semantic labels rather than relative file paths. I manually traced every declared link through implementation and named tests; all 19 of 19 are wired.

| Plan | Links | Trace |
|---|---:|---|
| 36-01 | 1/1 | Baseline comparator reads frozen bytes and checks current emitter/application/module routes. |
| 36-02 | 2/2 | Parser emits call AST; checker resolves the callee and signature, then lowers ordered typed calls with fresh result targets. |
| 36-03 | 3/3 | Corevalidator, originvalidator, and pathoracle independently derive call contracts from core facts. |
| 36-04 | 3/3 | Interpreter creates per-parent ordinals; executionpeer derives topology; capture completeness remains separate from program outcome. |
| 36-05 | 3/3 | Native emitter reserves bounded event/path capacity; capture failure stays separate from application exit; old routes compare literal bytes. |
| 36-06 | 3/3 | Local and imported sources reach both runners; defects and causal capture are separate; public route retains its input contract. |
| 36-07 | 2/2 | Source-free manifest is checked against independent witnesses and finite route/census consumers. |
| 36-08 | 2/2 | Acceptance evidence points to exact owner selectors and separates proof types. |

## Data-Flow Trace

There is no rendered UI or database-backed value in this phase. Program data reaches user-visible output: parser call expressions are resolved by checker lowering into typed core calls, then the interpreter executes the callee or cgen emits a typed C invocation. The result flows into assignment/expression use, while the public runner captures stdout, stderr, and exit. `helper-values.json` records independent Python arithmetic rather than engine-generated expected values.

| Artifact | Data | Source and flow | Status |
|---|---|---|---|
| Helper example modules | U64/Bool inputs and helper results | Source → parser → checker → core call → interpreter/native → captured output | ✓ FLOWING |
| `interp.go` | Callee return or defect | Runtime call frames execute the callee and return its result/reason to the caller | ✓ FLOWING |
| `cgen_program.go` | Typed native call result | Emits the invocation and captures the generated program's result | ✓ FLOWING |
| `executionpeer.go` | Call identity and causal path | Derives path and ordinal from runtime events, without proving scalar arithmetic | ✓ FLOWING |

## Behavioral Spot-Checks

These were exact named tests, run with `GOCACHE=/tmp/ai-lang-verification-gocache`; no full project or package suite was run.

| Behavior | Named roots | Result |
|---|---|---|
| Structured caller emission regression, including the uncalled ByteView Unit body shape | `TestPhase36UnusedByteStructuredHelper`, `TestPhase36ByteViewUnitHelperEmission` | PASS |
| Helper answers/defects, exact reasons and exits, imports, and refusal boundaries | `TestPhase36HelperAnswersAndDefects`, `TestPhase36HelperOptimizationMatrix`, `TestPhase36CallImportAndRefusalBoundaries` | PASS |
| Argument/result intent and independent raw/origin/causal derivation | `TestPhase36CallSourceIntentControls`, `TestPhase36RawCallShapeAndFreshness`, `TestPhase36OriginCallFreshnessAndTransfer`, `TestPhase36CallCausalAttributionControls` | PASS |
| Legacy ByteView bounds and caller-before-entry distinction | `TestPhase36LegacyByteViewCalleeBounds`, `TestPhase36ByteViewCallerArgumentBoundary`, `TestPhase36NativeByteViewCallerArgumentBoundary` | PASS; one native-lane group took about 12 seconds |
| Frozen no-call bytes and Phase 16 inventory mutation controls | `TestPhase36NoCallGeneratedCByteBaseline`, `TestPhase16PublicEmitterConsumerInventory`, `TestPhase16EmitterInventoryMutationControls` | PASS |
| Type-pair lowering | `TestPhase36ScalarCallLoweringAndTypePairs` | PASS |

The exact direct selectors were:

```sh
source /tmp/schway34-root-pin.sh
source /tmp/schway34-gsd-run.sh @opengsd/gsd-core1.16.0
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/cgen -run '^(TestPhase36UnusedByteStructuredHelper|TestPhase36ByteViewUnitHelperEmission)$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/session -run '^TestPhase36HelperAnswersAndDefects$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/session -run '^TestPhase36HelperOptimizationMatrix$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/check -run '^(TestPhase36CallSourceIntentControls|TestPhase36CallImportAndRefusalBoundaries|TestPhase36ScalarCallLoweringAndTypePairs)$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/corevalidate -run '^TestPhase36RawCallShapeAndFreshness$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/originvalidate -run '^TestPhase36OriginCallFreshnessAndTransfer$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/session -run '^(TestPhase36LegacyByteViewCalleeBounds|TestPhase36ByteViewCallerArgumentBoundary|TestPhase36NativeByteViewCallerArgumentBoundary)$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=20s
GOCACHE=/tmp/ai-lang-verification-gocache gsd_run run-with-timeout 30 -- go test ./internal/compiler/session -run '^(TestPhase16PublicEmitterConsumerInventory|TestPhase16EmitterInventoryMutationControls)$' -count=1 -v -timeout=20s
```

## Freshness and Regression Evidence

I checked the receipt input hashes and raw logs against the current files. The owner-validation receipt contains 86 executed roots across 13 packages, 85 current input hashes, and 14 logs; there are no missing, failing, or skipped roots. The final union regression receipt contains 251 root executions and 239 distinct test names across 16 packages, 179 input hashes, and 16 logs; all final logs are newly run with no missing, failing, or skipped roots. The final union is the exact current Phase 36 root set plus the repair roots, Phase 35 refresh roots (including earlier phases 33/34), and historical Phase 36 names. No receipt is treated as proof for a source whose hash differs.

The two Phase 08 prohibition requests were rechecked independently through the canonical enforcement owner. Each returns `green`, `flagged=false`, tier `test`, kind `node-test`, `failFirst=true`, `passed=true`, and `failFirstProof=violation-fixture`. Raw result hashes and all 22 bound source/fixture inputs match the final prohibition receipt. These are local checks, not hosted evidence.

## Requirements Coverage

| Requirement | Source plans | Status | Evidence |
|---|---|---|---|
| CALL-01 | 36-01 through 36-08 | ✓ SATISFIED | Local/imported U64/Bool results, recursion refusal, and exact engine defects verified in truths 1–2, 8–10, and 21–24. D-18 scopes ByteView bounds to the legacy callee and caller-before-entry controls. |
| CALL-02 | 36-01 through 36-08 | ✓ SATISFIED | Independent call derivation, argument/result/defect mutation controls, and no-call byte baseline verified in truths 3–4 and 11–20, 25–28. |

Both requirements are mapped to Phase 36 in `REQUIREMENTS.md`; no phase-mapped requirement is orphaned. The canonical decision coverage query reports 26/26 honored, with no unhonored CONTEXT decisions.

## Anti-Patterns and Review Findings

| File | Line | Finding | Severity | Impact |
|---|---:|---|---|---|
| `.planning/ROADMAP.md` | 317, 330 | Unreferenced `TBD` markers in Phase 38 plan-count fields | 🛑 BLOCKER | Phase 36 touched the roadmap, and the verifier debt-marker rule blocks unresolved markers in a phase-modified file. The entries correctly describe Phase 38 as not started, but the marker lines have no issue/PR/DEF-* reference. |
| `internal/compiler/corevalidate/corevalidate.go` | 4248 | WR-02 stale explanatory comment | ⚠️ WARNING | Comment describes the obsolete same-type-only invariant; implementation derives the argument type from `sourceTypeID`. Disposition defers the comment to the core-validator maintainer's next semantic edit; review is not clean. |

No runtime stub, empty handler, or empty value flowing to observable output was found. The grep matches for empty slices are intentional empty-result/diagnostic helpers and do not stand in for missing call behavior. The roadmap markers predate the current implementation changes and are about the explicitly unstarted Phase 38 planning work, but Step 7 applies to the modified file and does not exempt them for phase scope.

The Phase 33 review separately retains CR-03 and CR-04 as blockers for the execution-peer maintainer's mandatory M007 audit before the Phase 38 freeze. I inspected the overlap: those findings concern legacy nested defect-unwind and typed-failure-through-return behavior, while Phase 36's new scalar-call obligations and D-18 bounds controls are distinct. They are not closed or accepted risk by this report, and this is not a global-clean verdict.

The scoped Phase 36 security review records 24/24 authored L1/high threats closed; that is not a broader security certification. The UI gate is not applicable because the phase contains no frontend files or UI spec.

## Human Verification Required

None. Named deterministic tests exercise the phase's observable call, output, defect, refusal, and independent-peer behavior. There is no visual, hosted, or external-service behavior in this phase goal.

## Gaps Summary

The code achieves the scalar-helper-call goal: local and imported U64/Bool helper calls return used values, interpreter/native defect behavior matches, peers independently derive call facts, and no-call generated C remains byte-identical. All 32 must-haves pass, with no behavior-unverified truth. The overall verifier status is nevertheless `gaps_found` because the phase-modified roadmap retains two unreferenced Phase 38 `TBD` markers at lines 317 and 330, which the mandatory debt-marker rule treats as a blocker. The review also retains WR-02 and the separate Phase 33 M007 audit debt described above.

---

_Verified: 2026-10-16T02:06:09Z_
_Verifier: the agent (gsd-verifier)_

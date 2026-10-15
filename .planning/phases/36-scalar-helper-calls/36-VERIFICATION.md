---
phase: 36-scalar-helper-calls
verified: 2026-10-15T08:01:22Z
status: passed
score: "32/32 must-haves verified"
covered_files:
  - ".planning/phases/36-scalar-helper-calls/36-01-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-01-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-02-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-02-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-03-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-03-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-03-pathoracle-red.json"
  - ".planning/phases/36-scalar-helper-calls/36-04-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-04-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-05-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-05-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-05-task1-red-evidence.json"
  - ".planning/phases/36-scalar-helper-calls/36-05-task2-red-evidence.json"
  - ".planning/phases/36-scalar-helper-calls/36-05-task3-red-evidence.json"
  - ".planning/phases/36-scalar-helper-calls/36-06-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-06-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-06-task1-red-evidence.json"
  - ".planning/phases/36-scalar-helper-calls/36-06-task2-red-evidence.json"
  - ".planning/phases/36-scalar-helper-calls/36-06-task3-red-evidence.json"
  - ".planning/phases/36-scalar-helper-calls/36-07-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-07-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-08-PLAN.md"
  - ".planning/phases/36-scalar-helper-calls/36-08-SUMMARY.md"
  - ".planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md"
  - ".planning/phases/36-scalar-helper-calls/36-CONSTRUCTIBILITY.md"
  - ".planning/phases/36-scalar-helper-calls/36-CONTEXT.md"
  - ".planning/phases/36-scalar-helper-calls/36-DISCUSSION-LOG.md"
  - ".planning/phases/36-scalar-helper-calls/36-PATTERNS.md"
  - ".planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md"
  - ".planning/phases/36-scalar-helper-calls/36-REGRESSION.md"
  - ".planning/phases/36-scalar-helper-calls/36-RESEARCH.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-DISPOSITION.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-FIX.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW-SUPPLEMENT.md"
  - ".planning/phases/36-scalar-helper-calls/36-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-SECURITY.md"
  - ".planning/phases/36-scalar-helper-calls/36-TRANSITION-CHECK.md"
  - ".planning/phases/36-scalar-helper-calls/36-UI-REVIEW.md"
  - ".planning/phases/36-scalar-helper-calls/36-VALIDATION.md"
  - ".planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs"
  - ".planning/phases/36-scalar-helper-calls/prohibition-causal-violation.json"
  - ".planning/phases/36-scalar-helper-calls/prohibition-clean.json"
  - ".planning/phases/36-scalar-helper-calls/prohibition-outcome-violation.json"
  - "cmd/schway/main_phase32_routes_test.go"
  - "cmd/schway/main_phase36_test.go"
  - "examples/helper_calls/crc_step.schway"
  - "examples/helper_calls/importer.schway"
  - "examples/helper_calls/public.schway"
  - "internal/compiler/ast/ast.go"
  - "internal/compiler/cgen/cgen.go"
  - "internal/compiler/cgen/cgen_phase36_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/check/check.go"
  - "internal/compiler/check/check_phase34_refusal_test.go"
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
  - "internal/compiler/session/session_phase34_prechange_test.go"
  - "internal/compiler/session/session_phase34_surface_test.go"
  - "internal/compiler/session/session_phase35_admission_mutation_test.go"
  - "internal/compiler/session/session_phase35_surface_test.go"
  - "internal/compiler/session/session_phase36_byteview_test.go"
  - "internal/compiler/session/session_phase36_calls_test.go"
  - "internal/compiler/session/session_phase36_legacy_calls_test.go"
  - "internal/compiler/session/session_phase36_peer_mutation_test.go"
  - "internal/compiler/session/session_phase36_prechange_test.go"
  - "internal/compiler/session/session_phase36_surface_test.go"
  - "internal/compiler/surfacemanifest/manifest.go"
  - "internal/compiler/surfacemanifest/manifest_test.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase36_test.go"
  - "testdata/phase36/helper-values.json"
  - "testdata/phase36/new-fixture-ledger.json"
  - "testdata/phase36/prechange-corpus.json"
  - "testdata/phase36/surface-census.json"
  - "testdata/phase36/surface-witnesses.json"
covered_digest: "v3:sha256:d70149f072a81475f3adb8caa33acc62ab7f56102bc2e674bdbdf570939af00a"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: "30/30 observable truths; two prohibition gates unverified"
  gaps_closed:
    - "MUST NOT turn optional evidence exhaustion into application success or suppress the application's actual defect, output, or exit."
    - "MUST NOT present causal execution-peer acceptance as proof of scalar arithmetic or source argument/result intent."
  gaps_remaining: []
  regressions: []

---

# Phase 36: Scalar Helper Calls Verification Report

**Phase Goal:** A scalar function can be factored into small helpers: it calls another scalar function with one U64 or Bool argument and uses the result, including calls into imported modules, with the same defect behavior on both engines and every peer re-deriving the call independently.

**Verified:** 2026-10-15T08:01:22Z
**Status:** passed
**Re-verification:** Yes — after closing the two prohibition gaps

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---:|---|---|---|
| 1 | Local and imported scalar helpers use one U64/Bool argument and print pinned results on both engines; recursion remains refused. | ✓ VERIFIED | TestPhase36HelperAnswersAndDefects, TestPhase36HelperOptimizationMatrix, TestPhase36CallImportAndRefusalBoundaries. |
| 2 | Overflow, zero remainder, and source-grounded ByteView bounds cases retain exact reason/exit behavior on both engines. | ✓ VERIFIED | TestPhase36HelperOptimizationMatrix, TestPhase36LegacyByteViewCalleeBounds, TestPhase36ByteViewCallerArgumentBoundary, TestPhase36NativeByteViewCallerArgumentBoundary; D-18 keeps callee/caller bounds distinct. |
| 3 | Checker, corevalidator, originvalidator, and execution peer independently derive calls; argument/result/defect mutants are caught. | ✓ VERIFIED | TestPhase36CallSourceIntentControls, TestPhase36RawCallShapeAndFreshness, TestPhase36OriginCallFreshnessAndTransfer, TestPhase36CallCausalAttributionControls, TestPhase36EngineCallMutationControls. |
| 4 | No-call programs retain literal generated-C bytes, and the source-free manifest records arity one/no recursion. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaseline, TestPhase36SurfaceManifestCallWitnesses; baseline commit 9b8285de predates parser/compiler implementation. |
| 5 | Literal C baseline and live controls predate compiler edits. | ✓ VERIFIED | Git history places baseline commit 9b8285de before first Phase 36 parser implementation fe174ce3; TestPhase36NoCallGeneratedCByteBaselineControls. |
| 6 | Every tracked pre-change program has applicable emitter routes accounted for, including refusals. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaseline checks frozen inventory and route outcomes. |
| 7 | Earlier golden bytes stay immutable; new fixtures use exact ledger rows. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaselineControls; new-fixture-ledger.json checks the exact helper examples and omission/false-exemption controls. |
| 8 | Direct one-argument calls parse, recover, and format within the existing 64-level bound. | ✓ VERIFIED | TestPhase36ScalarCallGrammar, TestPhase36ScalarCallDepthAndRecovery, TestPhase36ScalarCallFormatting. |
| 9 | All four U64/Bool signature pairs lower to typed fresh call results, including initialized Bool identifier tails. | ✓ VERIFIED | TestPhase36ScalarCallLoweringAndTypePairs; TestPhase36HelperAnswersAndDefects. |
| 10 | Arguments execute once; siblings are ordered left-to-right; unused-Byte and whole-program recursion refusals remain. | ✓ VERIFIED | TestPhase36CallSourceIntentControls, TestPhase36ByteViewCallerArgumentBoundary, TestPhase36CallImportAndRefusalBoundaries; ordered core lowering inspected in check/scalar.go. |
| 11 | Raw core calls independently validate callee, signature, initialized argument, result type, and fresh target. | ✓ VERIFIED | TestPhase36RawCallShapeAndFreshness, TestPhase36RawCallRefusalRemovalControls. |
| 12 | Origin analysis transfers calls as initialized, value-unknown scalars through branch/loop assignments. | ✓ VERIFIED | TestPhase36OriginCallFreshnessAndTransfer, TestPhase36OriginCallRefusalRemovalControls. |
| 13 | Scalar calls add no loan ancestry or resource/back-edge authority. | ✓ VERIFIED | TestPhase36ScalarCallAuthorityBoundaries, TestPhase36ScalarCallLocalBorrowControls. |
| 14 | Repeated calls receive contiguous per-parent/per-site ordinals; child counters reset with a fresh parent. | ✓ VERIFIED | TestPhase36DynamicCallOrdinalControls, TestPhase36InterpreterRepeatedCallTracer, TestPhase36NativeDynamicInvocationPaths. |
| 15 | Interpreter results/defects and execution-peer causal attribution are separate claims. | ✓ VERIFIED | TestPhase36InterpreterCallAnswersAndDefects, TestPhase36CallCausalAttributionControls; plausible false-success remains accepted by design, so causal acceptance is not numeric proof. |
| 16 | Evidence retention, path/counter bounds, and exhaustion remain distinct from numeric work. | ✓ VERIFIED | TestPhase36InterpreterEvidenceBoundaries, TestPhase36CallEventCapacityBoundary, TestPhase36NativeCallCaptureOutcomeSeparation. |
| 17 | Native calls use sequenced typed arguments/results and stable dynamic paths. | ✓ VERIFIED | TestPhase36NativeDynamicInvocationPaths, TestPhase36EmitterCallInventoryControls; engine mutation capture covers O0/O3/O3-LTO. |
| 18 | Repeated-call captures and exact event/path/ordinal boundaries have named controls. | ✓ VERIFIED | TestPhase36NativeDynamicInvocationPaths, TestPhase36CallEventCapacityBoundary, TestPhase36NativeOrdinalAndPathBoundaries. |
| 19 | Optional capture exhaustion preserves application stdout, stderr, exit, and later defects. | ✓ VERIFIED | TestPhase36NativeCallCaptureOutcomeSeparation, TestPhase36CaptureOutcomeSeparation; success/late-defect capacity and disabled cases assert process/capture outcomes separately. |
| 20 | Old generated-C output and affected emitter inventories remain independently checked. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaseline, TestPhase36EmitterCallInventoryControls. |
| 21 | Local and imported repeated crc_step helpers produce pinned answers on both engines. | ✓ VERIFIED | TestPhase36HelperAnswersAndDefects, TestPhase36HelperOptimizationMatrix; examples/helper_calls/crc_step.schway and importer.schway. |
| 22 | Four signature pairs, nested/computed arguments, arithmetic use, mutable assignments, and Bool conditions have discriminating witnesses. | ✓ VERIFIED | TestPhase36HelperAnswersAndDefects, TestPhase36ScalarCallLoweringAndTypePairs; helper-values.json carries independently pinned values. |
| 23 | Dynamic defects retain exact reasons and exit 65 across O0/O3/O3-LTO; bounds obligations stay distinct. | ✓ VERIFIED | TestPhase36HelperOptimizationMatrix, TestPhase36LegacyByteViewCalleeBounds, caller-before-entry tests; sanitizer evidence is separate. |
| 24 | Public conformance retains its input contract and native capture remains separate from process outcome. | ✓ VERIFIED | TestPhase36PublicHelperConformance, TestPhase36NativeCallCaptureOutcomeSeparation. |
| 25 | Source-free manifest states bounded arity-one contexts and all four U64/Bool pairs. | ✓ VERIFIED | TestPhase36SurfaceManifestCallWitnesses; independently authored surface-witnesses.json. |
| 26 | Complete-program witnesses and finite source/census checks cover manifest claims. | ✓ VERIFIED | TestPhase36SurfaceManifestCallWitnesses, TestPhase36SurfaceManifestCallCensus. |
| 27 | Missing/false admission, refusal, hint, restriction, and census controls are bidirectional. | ✓ VERIFIED | TestPhase36SurfaceManifestCallControls, TestPhase36SurfaceLimitConsumers; old witness bytes remain frozen. |
| 28 | Integrated no-call programs retain literal-byte generated C through frozen routes. | ✓ VERIFIED | TestPhase36NoCallGeneratedCByteBaseline; final focused capture includes the exact baseline selector. |
| 29 | Acceptance evidence separates source, raw-contract, numeric, causal proof and both bounds obligations. | ✓ VERIFIED | 36-ACCEPTANCE-EVIDENCE.md and source/peer/engine/bounds tests. |
| 30 | Current inspection, fresh evidence, and historical/hosted receipts are separately labeled; next-phase ownership is preserved. | ✓ VERIFIED | 36-ACCEPTANCE-EVIDENCE.md, 36-CONTEXT.md, dated living records; no hosted claim is made for this local capture. |
| 31 | Optional evidence exhaustion cannot convert success into application success or hide the application's later defect/output/exit. | ✓ VERIFIED | Canonical outcome request returned green/unflagged with failFirstProof=violation-fixture; violation overlay kills both named capture assertions; clean/default controls pass. |
| 32 | Causal-peer acceptance is not proof of numeric arithmetic or source argument/result intent. | ✓ VERIFIED | Canonical causal request returned green/unflagged with failFirstProof=violation-fixture; numeric-limit mutation kills the named peer assertion; clean/default controls pass.

**Score:** 32/32 must-haves verified. The 30 runtime/contract truths remain verified; this re-verification closes the two prohibition-evidence gaps.

### Prohibition Gate Assessment

Both descriptors now resolve as test-tier prohibitions through the canonical producer. I compared each request field-for-field against its matching Plan 08 flat projection: statement, resolved/test status, node-test kind/target, statement-specific violation fixture, and shared clean fixture. I then reran the requests sequentially.

This supersedes the initial 2026-10-15T07:31:00Z gap verdict: at that point the two items were descriptorless and only their Go-level behavior tests existed, so the shared producer correctly had no machine evidence to green. The re-verification below closes those exact gaps without changing the original 30-truth evidence.

| Prohibition | Current proof | Gate disposition |
|---|---|---|
| MUST NOT turn optional evidence exhaustion into application success or suppress the application's actual defect, output, or exit. | Outcome violation overlay fails the exact existing assertions for changed successful work and hidden later defect; clean and default controls pass. | ✓ VERIFIED — producer status green, flagged false, tier test, kind node-test, failFirst=true, passed=true, failFirstProof=violation-fixture. |
| MUST NOT present causal execution-peer acceptance as proof of scalar arithmetic or source argument/result intent. | Numeric-limit overlay fails the named false-success assertion; clean and default controls pass. Source intent and arithmetic remain owned by separate checker/engine tests. | ✓ VERIFIED — producer status green, flagged false, tier test, kind node-test, failFirst=true, passed=true, failFirstProof=violation-fixture. |

Exact producer requests and outputs are recorded in 36-PROHIBITION-EVIDENCE.md. The adapter accepts only closed subjects and exact Go roots/assertions; malformed subjects, skips, unrelated failures, source-anchor drift, and infrastructure failures cannot count as red or green. The clean fixture exercises the same subject path, and the producer separately confirms default behavior. The seven-file review supplement reports zero findings and scopes its review to this bridge. No compiler/runtime source, Go test, corpus fixture, dependency, or public interface changed.

## Required Artifacts

The original artifact query passed 23/23 plan-declared artifacts. Updated Plan 08 artifact verification passed 4/4; together the current plan-declared total is 24/24. Paths are grouped by plan; source/import and consumer traces verified wiring.

| Plan | Artifacts | Status | Wiring/details |
|---|---|---|---|
| 36-01 | testdata/phase36/prechange-corpus.json; internal/compiler/session/session_phase36_prechange_test.go; testdata/phase36/new-fixture-ledger.json | ✓ VERIFIED (3/3) | Comparator consumes frozen bytes and route inventory. |
| 36-02 | internal/compiler/syntax/parser_phase36_test.go; internal/compiler/check/check_phase36_scalar_calls_test.go | ✓ VERIFIED (2/2) | Parser AST call reaches checker lowering and typed core. |
| 36-03 | internal/compiler/corevalidate/corevalidate_phase36_test.go; internal/compiler/originvalidate/originvalidate_phase36_test.go; internal/compiler/pathoracle/pathoracle_phase36_test.go | ✓ VERIFIED (3/3) | Independent validators consume core facts without a shared checker verdict. |
| 36-04 | internal/compiler/interp/interp_phase36_test.go; internal/compiler/executionpeer/executionpeer_phase36_test.go | ✓ VERIFIED (2/2) | Runtime frames/events and independent peer ancestry are exercised. |
| 36-05 | internal/compiler/cgen/cgen_phase36_test.go; internal/compiler/native/native_phase36_test.go | ✓ VERIFIED (2/2) | C emitter and native application capture boundary are connected. |
| 36-06 | examples/helper_calls/crc_step.schway; examples/helper_calls/importer.schway; examples/helper_calls/public.schway; testdata/phase36/helper-values.json; internal/compiler/session/session_phase36_calls_test.go | ✓ VERIFIED (5/5) | Programs reach interpreter, native, and public conformance routes. |
| 36-07 | testdata/phase36/surface-witnesses.json; testdata/phase36/surface-census.json; internal/compiler/session/session_phase36_surface_test.go | ✓ VERIFIED (3/3) | Independent witness and exact census documents are consumed by manifest controls. |
| 36-08 | .planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md; .planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md; .planning/PRODUCT-ROADMAP.md; .planning/LANGUAGE-MATURITY.md | ✓ VERIFIED (4/4) | Acceptance/prohibition receipts and dated living records exist; updated declared evidence artifact passed. |
| Proof bridge | 36-prohibition-check.test.cjs; prohibition-clean.json; prohibition-outcome-violation.json; prohibition-causal-violation.json | ✓ VERIFIED | Both canonical producer requests ran violation, clean, and default subjects; review supplement found zero findings. |

## Key Link Verification

The link query could not interpret semantic labels such as “checker signature resolution” as file paths. I manually traced all 19 declared links to their concrete owners:

| Plan | Links verified | Evidence |
|---|---:|---|
| 36-01 | 1/1 | Comparator reads frozen pre-change corpus and compares current emitter/application/module routes. |
| 36-02 | 2/2 | Parser emits call AST; checker resolves callee/signature and lowers ordered typed OpCalls with fresh targets. |
| 36-03 | 3/3 | Corevalidator, originvalidator, and pathoracle derive constraints from core facts and preserve authority boundaries. |
| 36-04 | 3/3 | Interpreter creates per-parent ordinals; executionpeer derives its own ordinal/topology; collector incompleteness does not replace program outcome. |
| 36-05 | 3/3 | Native emitter reserves bounded paths/events; capacity guards separate conformance exit 69 from app capture; old routes compare literal bytes. |
| 36-06 | 3/3 | Local/importer sources reach existing runners; defects and causal capture are separate; public route keeps input contract. |
| 36-07 | 2/2 | Source-free manifest is checked against independently authored witnesses and finite route/census consumers. |
| 36-08 | 2/2 | Frozen corpus controls compare current outputs; acceptance evidence cites exact owner selectors and separates proof kinds. |

### Data-Flow Trace (Level 4)

N/A for rendered application data: this compiler/runtime phase has no dynamic UI or database-backed rendering. Source flows through parser, checked core, interpreter or generated C/native execution; literal fixtures and exact output expectations anchor results.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Local/imported helper answers and defects | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36HelperAnswersAndDefects$' -count=1 -v -timeout=600s | PASS; exact values, output, and defect reasons/exits. | ✓ PASS |
| Source-intent argument/result controls | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase36CallSourceIntentControls$' -count=1 -v -timeout=600s | PASS; argument and result mutants are discriminated by checker. | ✓ PASS |
| Causal-peer negative controls | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -v -timeout=600s | PASS; malformed causal traces fail while plausible false-success remains accepted by design. | ✓ PASS |
| Evidence-exhaustion outcome separation | GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NativeCallCaptureOutcomeSeparation$' -count=1 -v -timeout=600s | PASS; process output/exit and later defect survive disabled/exhausted capture. | ✓ PASS |
| Outcome prohibition producer | gsd_run check prohibition-enforcement /tmp/schway36-prohibition-proof-1aa485499637/outcome-request.json --raw | PASS; exact Plan 08 projection; green, unflagged, failFirst=true, passed=true, failFirstProof=violation-fixture. | ✓ PASS |
| Causal prohibition producer | gsd_run check prohibition-enforcement /tmp/schway36-prohibition-proof-1aa485499637/causal-request.json --raw | PASS; exact Plan 08 projection; green, unflagged, failFirst=true, passed=true, failFirstProof=violation-fixture. | ✓ PASS |
| Final Phase 36 owner capture | /tmp/schway36-final-979357388171-x4n9kf3c/owner-results.json | 45 expected roots across 12 packages; 12/12 commands exit 0, matching log SHA-256 values, no FAIL/SKIP; bound to 979357388171339005fd6dbd65aa760810ae9384. Current HEAD 2548fd95d51f3e24805254a5279081622928e30f has no source/test diff from captured revision. | ✓ PASS |
| Prior-phase scoped regression | /tmp/schway36-regression-b51043c6566f-_h2nianh/owner-results.json | 80 root occurrences / 69 names across 13 packages; source/test unchanged since recorded revision. | ✓ PASS (revision-bound) |

No full suite, whole-package run, broad cross-program native matrix, hosted dispatch, or push was run.

### Probe Execution

No phase-declared or conventional scripts/*/tests/probe-*.sh probes were found; not applicable.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| CALL-01 | 36-01..36-08 | One-argument U64/Bool calls, imported helpers, result use, recursion refusal, and defect behavior. | ✓ SATISFIED | Truths 1–2, 8–10, and 21–24; D-18 bounds cases remain distinct. |
| CALL-02 | 36-01..36-08 | Independent call derivation, mutant detection, and no-call generated-C compatibility. | ✓ SATISFIED | Truths 3–4, 11–20, and 25–28; prohibition dispositions are separately verified by the canonical producer. |

The REQUIREMENTS.md checkboxes remain pending because the phase orchestrator owns tracker updates; implementation evidence satisfies both descriptions. This scoped re-verification changes neither requirement ownership nor source-grounded classification.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unreferenced TBD/FIXME/XXX debt markers or functional stubs found. Two comments use “placeholder” for an intentionally unused typed C place and a one-byte ABI field. | ℹ️ Info | Neither comment is unfinished code. |

The 52-file code review is clean with zero open dispositions; WR-01 is fixed. Security audit reports 24/24 mitigations closed and zero open threats. UI review is N/A because this phase adds no frontend.

### Human Verification Required

None. Runtime claims and both scope prohibitions have deterministic evidence; no user-facing manual check is needed.

### Gaps Summary

No current gaps remain. The initial verification dated 2026-10-15T07:31:00Z passed all 30 observable runtime/plan truths but recorded two descriptorless prohibition gates as partial gaps. This re-verification closes both with the Plan 08 projected descriptors, exact fail-first violation fixtures, clean controls, and green canonical producer outputs. The prior 30-truth evidence and initial gap verdict remain in this report as history. Current score is 32/32 and status is passed.

The original 45-root feature capture remains bound to 979357388171339005fd6dbd65aa760810ae9384. No tracked compiler/test/corpus/dependency path changed since that capture. The added proof bridge is limited to the named assertions and local Node/Go toolchain recorded in 36-PROHIBITION-EVIDENCE.md; it does not claim broader production-source or hosted validation.

## Post-completion metadata refresh — 2026-10-15

`phase.complete 36` ran once after the 32/32 passed verification, closed CALL-01/CALL-02, and advanced the tracker. The subsequent completion-path correction expanded three RED-receipt paths and the `examples/phase24/error.schway` reference in 36-05-SUMMARY.md; each referenced file exists. The diff from completion commit `5a08f5e5` to metadata commit `6b5ff67c` contains planning-document changes and an empty Phase 37 `.gitkeep`; there are no compiler, test, fixture, example-content, or dependency changes. The transition record was reviewed and added to this report's covered inputs. Because implementation evidence is unchanged, the runtime producers, feature capture, and regressions were not rerun.

---

_Verified: 2026-10-15T08:01:22Z_
_Verifier: gsd-verifier_

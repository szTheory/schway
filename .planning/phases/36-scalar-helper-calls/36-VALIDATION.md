---
phase: 36
slug: scalar-helper-calls
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-10-14
---

# Phase 36 — Validation Strategy

This is the pre-execution sampling contract, derived from 36-RESEARCH's
Validation Architecture. New selectors below are planned, not existing/passing
evidence. The planner must refine task IDs/waves/threat references and ensure
each test is created before its selector can earn credit.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 `testing`, existing local compiler/application runners |
| Config | `go.mod`; no new package |
| Quick command, after test creation | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase36ScalarCallGrammar$' -count=1 -v -timeout=600s` |
| Phase verification | Exact named selectors from finalized plans; no local whole-package/full-suite or broad cross-program native matrix |
| Runtime | Unmeasured until focused execution; 600s per-command timeout is a cap, not a latency measurement |

## Sampling Rate

- Each task: exact owner selector and applicable positive/no-op/restored control.
- Each wave: completed cross-owner helper answer, defect, identity and capacity selectors.
- Phase gate: every finalized plan selector and affected inherited guard must produce actual root RUN/PASS output; zero-test success, skips and unrelated mutant failures do not count.
- Existing recurring CI owns broad checks; final two-host receipts remain Phase 38. No human-only acceptance is expected for these deterministic contracts.

## Per-Task Verification Map

Task/wave placement remains pending planner output. `—` threat references are
unassigned, not declarations of security coverage.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---|---|---|---|---|---|---|---|---|---|
| 36-baseline | pending | first | CALL-02 | — | No-new-call C bytes immutable; drift/omission/false-exemption controls reached | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-grammar | pending | pending | CALL-01 | — | Arity/type/64-depth/recovery bounded; Bool identifier tails only | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase36ScalarCallGrammar$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-lowering | pending | pending | CALL-01 | — | Correct source argument/result; all four type pairs; no unused Byte escape | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase36ScalarCallLoweringAndTypePairs$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-core | pending | pending | CALL-02 | — | Independent typed initialized argument and fresh result, value unknown | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^TestPhase36RawCallShapeAndFreshness$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-origin | pending | pending | CALL-02 | — | Independent origin/authority/result transfer, malformed raw call refused | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^TestPhase36OriginCallFreshnessAndTransfer$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-causal | pending | pending | CALL-02 | — | Dynamic ordinals, parent/site/callee/terminal attribution checked | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^(TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls)$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-native | pending | pending | CALL-02 | — | Stable bounded dynamic path storage, event reservation and named exhaustion | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase36NativeDynamicInvocationPaths|TestPhase36CallEventCapacityBoundary)$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-outcomes | pending | pending | CALL-01, CALL-02 | — | Both-engine answers/defects; argument OOB precedes helper; optional capture separate | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36HelperAnswersAndDefects|TestPhase36ByteViewCallerArgumentBoundary|TestPhase36NativeCallCaptureOutcomeSeparation)$' -count=1 -v -timeout=600s` | W0 missing | pending |
| 36-surface | pending | pending | CALL-02 | — | Independent complete-program witnesses and two-way manifest drift controls | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36SurfaceManifestCallWitnesses|TestPhase36SurfaceManifestCallCensus|TestPhase36SurfaceManifestCallControls)$' -count=1 -v -timeout=600s` | W0 missing | pending |

## Wave 0 Requirements

- [ ] Commit literal enumerated pre-change C corpus and reached comparator controls before compiler edits.
- [ ] Create grammar/lowering/all-four-signature positive and negative fixtures.
- [ ] Create independent raw core/origin refusal-removal tests with no-op/restored baselines.
- [ ] Create repeated/nested/two-site invocation fixtures and forged trace controls.
- [ ] Prove bounded native path-storage constructibility, exact identity and exhaustion before committing to its representation.
- [ ] Add engine exact-answer and dynamic-callee defect controls at O0/O3/O3-LTO; preserve separate legacy ByteView bounds regression.
- [ ] Add manifest witness/census and restriction drift controls.
- [ ] Resolve existing inherited selector names from RESEARCH/source and retain only affected focused guards in finalized plans.

## Manual-Only Verifications

None expected. Human readability, whole-loop performance distributions and
external hosted receipts are not inferred from these automated tests.

## Validation Sign-Off

- [ ] Final task/wave/threat mapping matches executable plans.
- [ ] Every runnable verify has a stated observable failing direction.
- [ ] Missing tests created before their selectors earn credit.
- [ ] Every requirement and declared mitigation has named passing evidence.
- [ ] Mutation failures reach the intended assertion and restored positives pass.
- [ ] `status: validated`, `nyquist_compliant: true`, and `wave_0_complete: true` only after earned audit evidence.

**Approval:** pending implementation and validation audit.

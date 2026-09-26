---
phase: "19"
slug: "numeric-literals-and-opconst"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-24"
validated: "2026-09-26"
---

# Phase 19 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package (Go 1.24) |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/core` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | Full suite passed in 184.8s on 2026-09-26. |

The current `STATE.md` records unrelated existing red Phase 11 gate fixtures,
stale corpus counts/validation digest, and unpinned groundedness findings. Plans
must distinguish those known baseline issues from Phase 19 regressions.

## Sampling Rate

- **After each touched package task:** Run focused tests for the changed syntax,
  checker/core, validator, interpreter, or C emitter package.
- **After each plan wave:** Run that wave's focused gate. The four-tier literal
  comparison becomes runnable after Wave 7 integrates both engines and all
  independent peers. Run `go test ./...` at the phase gate, reporting known
  unrelated baseline failures separately.
- **Before `$gsd-verify-work`:** Run the full suite and the phase-specific
  evidence gates.
- **Max feedback latency:** Measure during the phase; do not claim a budget
  before collecting timings.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | VAL-01 | T-19-02 | Pin the refused source fixture before production edits. | source frontier | `go test ./internal/compiler/session -run 'TestPhase19LiteralFrontier' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-01-02 | 01 | 1 | VAL-01 | T-19-02 | Pin separate malformed and overflow refusals. | source frontier | `go test ./internal/compiler/session -run 'TestPhase19(LiteralFrontier|NumericRefusalFrontiers)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-02-01 | 02 | 2 | Roadmap criterion 4 | T-19-01 | Widen interpreter values and replay old scalar documents before OpConst routing. | regression | `go test ./internal/compiler/interp ./internal/compiler/session -run 'TestPhase19(ScalarProjection|LiteralFrontier)|TestPayloadCorpusCharacterizationReplay' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-02-02 | 02 | 2 | VAL-03 | T-19-01 | Prove the scalar replay detects a seeded projection fault. | mutation | `go test ./internal/compiler/session -run 'TestPhase19ScalarGate|TestPayloadCorpusCharacterizationReplayMutationKilled|TestPayloadCorpusCharacterizationReplay' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-03-01 | 03 | 3 | VAL-01 | T-19-03 | Refuse malformed complete numeric tokens. | unit | `go test ./internal/compiler/syntax -run 'TestPhase19Numeric(Token|Malformed)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-03-02 | 03 | 3 | VAL-01 | T-19-03 | Parse and format literal bindings without changing spelling. | round trip | `go test ./internal/compiler/syntax ./internal/compiler/session -run 'TestPhase19(Numeric|LiteralFrontier)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-04-01 | 04 | 4 | VAL-01, VAL-02 | T-19-06 | Define fixed U64 ability and kind-exclusive OpConst shape. | unit | `go test ./internal/compiler/ability ./internal/compiler/core -run 'TestPhase19(U64Ability|OpConstShape)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-04-02 | 04 | 4 | VAL-01 | T-19-05 | Refuse U64 overflow before typed OpConst. | checker | `go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase19(LiteralAdmission|LiteralRange|LiteralType|LiteralFrontier|NumericRefusalFrontiers)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-05-01 | 05 | 5 | VAL-02 | T-19-07 | Independently admit only canonical U64 core facts. | peer + forged core | `go test ./internal/compiler/corevalidate -run 'TestPhase19(OpConst|U64|Forged)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-05-02 | 05 | 5 | VAL-02 | T-19-08 | Stop path and origin traversal at source-free constant roots. | peer | `go test ./internal/compiler/pathoracle ./internal/compiler/originvalidate -run 'TestPhase19(OpConst|ConstantOrigin|ConstantPath)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-06-01 | 06 | 6 | VAL-01 | T-19-10 | Execute OpConst in the interpreter with canonical decimal output. | interpreter | `go test ./internal/compiler/interp ./internal/compiler/session -run 'TestPhase19(OpConstInterpreter|ScalarProjection)|TestPayloadCorpusCharacterizationReplay' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-06-02 | 06 | 6 | VAL-01 | T-19-09 | Emit exact-width U64 native C and fail closed on unsupported targets. | native | `go test ./internal/compiler/cgen -run 'TestPhase19(U64Native|OpConst|ExactWidth)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-07-01 | 07 | 7 | VAL-02 | T-19-12 | Exercise OpConst at all six consumers with both mutation-sensitive controls. | integration + mutation | `go test ./internal/compiler/core ./internal/compiler/session -run 'Test(AllOperationKinds|Phase7DispatchControlsMutationKilled|Phase19Dispatch)' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |
| 19-07-02 | 07 | 7 | VAL-03 | T-19-13 | Observe exact U64 result across four tiers and five axes. | differential | `go test ./internal/compiler/session ./internal/compiler/core -run 'TestPhase19(FourTier|LiteralRun|WrongResult|Dispatch)|TestPayloadCorpusCharacterizationReplay|TestAllOperationKindsHandledAtEverySite' -count=1` | ✅ | EXERCISED | — | ✅ PASS (full suite) |

## Wave 0 Requirements

- [x] Wave 1: check in the literal, overflow, and malformed fixtures and pin their current production diagnostic before any compiler production edit.
- [x] Wave 2: widen interpreter values and replay D-12-18 scalar goldens before `OpConst` reaches another dispatch site.
- [x] Add accepted/refused decimal, hexadecimal, binary, separator, formatting,
  maximum-U64, and overflow cases in their owning package tests.
- [x] Add a real literal-bearing fixture that drives `OpConst` through all six
  required dispatch consumers and both exhaustive-dispatch controls.
- [x] Add a four-tier observable literal result using the existing five-axis
  comparator.
- [x] Identify and run the existing D-12-18 scalar golden replay before
  broadening operation dispatch.

## Manual-Only Verifications

All phase behaviors have automated verification. Human review remains useful
for any intentionally moved golden and its written rationale.

## Validation Sign-Off

- [x] All 14 plan tasks have an automated verify command or explicit Wave 0 dependency.
- [x] Sampling continuity: no 3 consecutive tasks without automated verification.
- [x] Wave 0 covers all test seams needed by plan tasks.
- [x] No watch-mode flags.
- [x] Feedback latency measured and recorded (full suite: 184.8s).
- [x] `nyquist_compliant: true` set after validation passes.

**Approval:** verified 2026-09-26

## Validation Audit 2026-09-26

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

All 14 plan tasks declare automated verification. The nine original map entries
and five previously omitted task entries are now recorded. The full suite passed
with `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` (184.8s); its
one-shot command has no watch mode. No new tests were needed.

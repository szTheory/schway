---
phase: "19"
slug: "numeric-literals-and-opconst"
status: validated
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-24"
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
| **Estimated runtime** | Not measured for this phase; measure focused and full lanes during execution. |

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

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | VAL-01 | T-19-02 | Pin the refused source fixture before production edits. | source frontier | `go test ./internal/compiler/session -run 'TestPhase19LiteralFrontier' -count=1` | ❌ W1 | ⬜ pending |
| 19-01-02 | 01 | 1 | VAL-01 | T-19-02 | Pin separate malformed and overflow refusals. | source frontier | `go test ./internal/compiler/session -run 'TestPhase19(LiteralFrontier|NumericRefusalFrontiers)' -count=1` | ❌ W1 | ⬜ pending |
| 19-02-01 | 02 | 2 | Roadmap criterion 4 | T-19-01 | Widen interpreter values and replay old scalar documents before OpConst routing. | regression | `go test ./internal/compiler/interp ./internal/compiler/session -run 'TestPhase19(ScalarProjection|LiteralFrontier)|TestPayloadCorpusCharacterizationReplay' -count=1` | ✅ replay precedent; ❌ new test | ⬜ pending |
| 19-03-01 | 03 | 3 | VAL-01 | T-19-03 | Refuse malformed complete numeric tokens. | unit | `go test ./internal/compiler/syntax -run 'TestPhase19Numeric(Token|Malformed)' -count=1` | ❌ W3 | ⬜ pending |
| 19-04-02 | 04 | 4 | VAL-01 | T-19-05 | Refuse U64 overflow before typed OpConst. | checker | `go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase19(LiteralAdmission|LiteralRange|LiteralType|LiteralFrontier|NumericRefusalFrontiers)' -count=1` | ❌ W4 | ⬜ pending |
| 19-05-01 | 05 | 5 | VAL-02 | T-19-07 | Independently admit only canonical U64 core facts. | peer + forged core | `go test ./internal/compiler/corevalidate -run 'TestPhase19(OpConst|U64|Forged)' -count=1` | ❌ W5 | ⬜ pending |
| 19-06-02 | 06 | 6 | VAL-01 | T-19-09 | Emit exact-width U64 native C and fail closed on unsupported targets. | native | `go test ./internal/compiler/cgen -run 'TestPhase19(U64Native|OpConst|ExactWidth)' -count=1` | ❌ W6 | ⬜ pending |
| 19-07-01 | 07 | 7 | VAL-02 | T-19-12 | Exercise OpConst at all six consumers with both mutation-sensitive controls. | integration + mutation | `go test ./internal/compiler/core ./internal/compiler/session -run 'Test(AllOperationKinds|Phase7DispatchControlsMutationKilled|Phase19Dispatch)' -count=1` | ❌ W7 | ⬜ pending |
| 19-07-02 | 07 | 7 | VAL-03 | T-19-13 | Observe exact U64 result across four tiers and five axes. | differential | `go test ./internal/compiler/session ./internal/compiler/core -run 'TestPhase19(FourTier|LiteralRun|WrongResult|Dispatch)|TestPayloadCorpusCharacterizationReplay|TestAllOperationKindsHandledAtEverySite' -count=1` | ❌ W7 | ⬜ pending |

## Wave 0 Requirements

- [ ] Wave 1: check in the literal, overflow, and malformed fixtures and pin their current production diagnostic before any compiler production edit.
- [ ] Wave 2: widen interpreter values and replay D-12-18 scalar goldens before `OpConst` reaches another dispatch site.
- [ ] Add accepted/refused decimal, hexadecimal, binary, separator, formatting,
  maximum-U64, and overflow cases in their owning package tests.
- [ ] Add a real literal-bearing fixture that drives `OpConst` through all six
  required dispatch consumers and both exhaustive-dispatch controls.
- [ ] Add a four-tier observable literal result using the existing five-axis
  comparator.
- [ ] Identify and run the existing D-12-18 scalar golden replay before
  broadening operation dispatch.

## Manual-Only Verifications

All phase behaviors have automated verification. Human review remains useful
for any intentionally moved golden and its written rationale.

## Validation Sign-Off

- [ ] All plan tasks have an automated verify command or explicit Wave 0 dependency.
- [ ] Sampling continuity: no 3 consecutive tasks without automated verification.
- [ ] Wave 0 covers all test seams needed by plan tasks.
- [ ] No watch-mode flags.
- [ ] Feedback latency measured and recorded.
- [ ] `nyquist_compliant: true` set after validation passes.

**Approval:** pending

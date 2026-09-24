---
phase: "19"
slug: "numeric-literals-and-opconst"
status: draft
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
- **After each plan wave:** Run the focused phase integration and four-tier
  literal comparison; run `go test ./...` at the phase gate, reporting known
  unrelated baseline failures separately.
- **Before `$gsd-verify-work`:** Run the full suite and the phase-specific
  evidence gates.
- **Max feedback latency:** Measure during the phase; do not claim a budget
  before collecting timings.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | VAL-01 | T-19-01 | Reject malformed/out-of-range source literals before executable core. | unit + integration | `go test ./internal/compiler/syntax ./internal/compiler/check` | ❌ W0 | ⬜ pending |
| 19-02-01 | 02 | 1 | VAL-02 | — | Every independent peer handles admitted `OpConst`; removal controls detect omissions. | integration + mutation control | `go test ./internal/compiler/core ./internal/compiler/corevalidate ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/interp ./internal/compiler/cgen` | ❌ W0 | ⬜ pending |
| 19-03-01 | 03 | 2 | VAL-03 | T-19-01 | Exact U64 values are observed consistently across interpreter and native tiers. | differential | `go test ./internal/compiler/session -run 'Phase19|FourTier|Phase5' -count=1` | ❌ W0 | ⬜ pending |
| 19-01-02 | 01 | 1 | Roadmap criterion 4 | T-19-01 | Existing scalar execution projection remains byte-identical unless a movement is justified. | regression | `go test ./internal/compiler/session -run 'PayloadCorpusCharacterizationReplay' -count=1` | ✅ precedent | ⬜ pending |

Plan/task identifiers above are provisional sampling slots. The planner may
renumber or consolidate them, but every phase requirement and the scalar golden
gate must remain covered.

## Wave 0 Requirements

- [ ] Add accepted/refused decimal, hexadecimal, binary, separator, formatting,
  maximum-U64, and overflow fixtures.
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

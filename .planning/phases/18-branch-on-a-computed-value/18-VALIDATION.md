---
phase: "18"
slug: "branch-on-a-computed-value"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-24"
---

# Phase 18 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package (Go 1.24) |
| **Config file** | `go.mod`; `.github/workflows/ci.yml` |
| **Quick run command** | `go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -count=1` |
| **Full suite command** | `go test ./...`; phase gate also runs `go vet ./...`, `go build ./...`, `go test -race ./...`, and native differential checks with installed Clang |
| **Estimated runtime** | Measure in Wave 0 and record; do not assume a latency budget without evidence |

---

## Sampling Rate

- **After every task commit:** Run the narrow owning-package test(s) for changed parser, checker, validator, interpreter, or backend behavior.
- **After every plan wave:** Run the focused Phase 18 source/differential/mutation checks and `go test ./...`.
- **Before `$gsd-verify-work`:** Run `go vet ./...`, `go build ./...`, `go test ./...`, `go test -race ./...`, and the native differential evidence with installed Clang.
- **Max feedback latency:** Measure the focused and full-suite durations in Wave 0; keep the recurring focused CI lane bounded and justify any broader recurring lane by regression value and runtime.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| Wave 0 | 01 | 0 | CTL-01 | T-18-01 | Pin the existing computed-scrutinee refusal; independent validators fail closed for malformed or forged core | fixture, unit | `go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase18' -count=1` | ❌ W0 | ⬜ pending |
| Wave 0 | 01 | 0 | CTL-02 | T-18-02 | A Result-returning callee's computed value reaches all five comparator axes through independent peer admission | integration, differential, smoke | `go test ./internal/compiler/session -run 'TestPhase18.*Result|TestPhase18.*Computed' -count=1` | ❌ W0 | ⬜ pending |
| Wave 0 | 01 | 0 | CTL-03 | T-18-03 | Payload place is returned; injected wrong-slot corruption diverges at `axis:terminal-outcome` | integration, mutation-kill | `go test ./internal/compiler/session -run 'TestPhase18.*Payload|TestPhase18.*WrongSlot' -count=1` | ❌ W0 | ⬜ pending |
| Wave 0 | 01 | 0 | CTL-01 | T-18-01 | A pre-match loan live in one arm is classified with existing endpoints and bounded fixpoint work | source integration, ownership | `go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase18.*Loan|TestEdgeSpecificLiveOut' -count=1` | ❌ W0 | ⬜ pending |

Threat refs:

- **T-18-01:** Untrusted input causes unbounded parser, CFG, or liveness work; preserve limits and fail closed at `4 × blocks × (loans+1)`.
- **T-18-02:** Checker and independent admission peers derive computed-place facts differently; independently validate place, type, and ownership.
- **T-18-03:** Native code writes the wrong payload slot but the comparator misses it; require mutation injection and terminal-outcome divergence.

The five-axis acceptance must use the full session/program comparator and peer-validation route. `Phase5CompareEngines` alone covers only four execution-bearing axes; diagnostic-ID refusal comparison is separate.

---

## Wave 0 Requirements

- [ ] Add and pin the refused CTL-01 computed-place `.lang` frontier fixture before changing production admission.
- [ ] Add source fixtures for a Result-returning computed match, a destructured payload-place return, and a pre-branch loan live in one arm.
- [ ] Add focused requirement, independent-peer, comparator-axis, and wrong-slot mutation-kill tests.
- [ ] Confirm the full five-axis evidence route and measure focused/full test latency.
- [ ] Add or extend recurring CI coverage only where the focused regression value justifies maintenance and runtime cost; use existing macOS/Linux CI lanes.

---

## Manual-Only Verifications

All objective phase behaviors have automated verification. Human review may assess code/API readability, but no acceptance criterion is delegated to manual UAT when source fixtures, peer checks, differential execution, mutation controls, and CI can establish it.

---

## Validation Sign-Off

- [ ] All plan tasks have an `<automated>` verify command or explicit Wave 0 dependency.
- [ ] Sampling continuity: no 3 consecutive tasks without automated verification.
- [ ] Wave 0 covers all missing tests and fixtures above.
- [ ] No watch-mode flags.
- [ ] Feedback latency is measured and documented; recurring CI cost is justified.
- [ ] `nyquist_compliant: true` set after Wave 0 and plan mapping are confirmed.

**Approval:** pending plan and Wave 0 validation.

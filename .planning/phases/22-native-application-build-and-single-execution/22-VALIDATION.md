---
phase: "22"
slug: "native-application-build-and-single-execution"
status: in_progress
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-27"
---

# Phase 22 — Validation Strategy

> Seeded from `22-RESEARCH.md` § Validation Architecture. Execution evidence is recorded as each dependent plan completes; planning itself made no acceptance claims.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` package; no external test framework |
| **Config file** | `go.mod`; no test-runner configuration |
| **Quick run command** | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./internal/compiler/session ./internal/compiler/cgen ./cmd/lang` |
| **Full suite command** | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` |
| **Estimated runtime** | Not measured for this phase; preserve cold and warm samples when recording the phase feedback loop |

---

## Sampling Rate

- **After every task commit:** Run the changed package tests and the focused CLI/session tests for the touched boundary.
- **After every plan wave:** Run the quick command and the documented retained-artifact CLI smoke cases for the wave.
- **Before `$gsd-verify-work`:** Run the full suite and the end-to-end identity witness, including relocation, malformed input, launch count, streams, process outcomes, and evidence states.
- **Feedback latency:** Record cold and warm distributions for the relevant author/check/build/run/observe loop; do not substitute a single timing sample.

---

## Per-Requirement Verification Map

Task and plan IDs are assigned by the planner; the planner or `$gsd-validate-phase` maps these requirement rows onto executable tasks.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 22-01-T1, 22-01-T2, 22-02-T2 | 22-01, 22-02 | 1, 2 | APP-02 | T-22-01, T-22-03 | Build does not execute the app; retained artifact runs after temporary build cleanup and checkout relocation | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./internal/compiler/session ./cmd/lang` | ❌ Wave 0/new tests | ⬜ pending |
| 22-01-T1, 22-01-T2 | 22-01 | 1 | APP-03 | T-22-04 | Bounded decimal U64 accepts `7` and `42`; malformed, overlong, extra, and overflowing input is rejected before app effects | unit + CLI integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/session ./cmd/lang` | ✅ tests exist | ✅ green |
| 22-01-T1, 22-01-T2, 22-03-T1 | 22-01, 22-03 | 1, 3 | APP-04 | T-22-02 | One application request starts exactly one selected artifact; verification replay remains explicit | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./internal/compiler/session ./cmd/lang` | ❌ Wave 0/new tests | ⬜ pending |
| 22-01-T1, 22-01-T2 | 22-01 | 1 | APP-05 | T-22-02 | Application stdout/stderr pass through; nonzero exit, signal, timeout, and launch failure remain distinct from compiler protocol errors | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./cmd/lang` | ✅ tests exist | ✅ green |
| 22-03-T1 | 22-03 | 3 | APP-06 | T-22-05 | Separate evidence reports distinguish complete, disabled, incomplete, and capacity-exhausted states; incomplete evidence never verifies success | unit + mutation/control | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/native` | ❌ Wave 0/new tests | ⬜ pending |
| 22-02-T1, 22-02-T2 | 22-02 | 2 | FFI-02 | T-22-01, T-22-03, T-22-06 | Declared local inputs resolve after relocation; missing/incompatible inputs fail; every relevant input changes build/evidence identity | unit + integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./internal/compiler/session` | ❌ Wave 0/new tests | ⬜ pending |
| 22-03-T2 | 22-03 | 3 | EVD-11 | T-22-05, T-22-06 | Isolated identity replay checks independent answers; a separate verifier-only `fixture.value` model case consumes and compares declared U64 42, while missing/duplicate/unconsumed/mismatched outcomes fail closed and reports deny host-IO/cleanup claims | integration + positive model witness + negative controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang -run '^TestPhase22(AppVerify|Replay)' -count=1 -v` | ❌ Wave 0/new tests | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] 22-01-T1/T2: identity Lang source and independent expected pairs `7 → 7` and `42 → 42`.
- [ ] 22-02-T1/T2: versioned local C binding manifest plus relocation fixture for source, header, symbol, ABI, and runtime dependency declarations.
- [x] 22-01-T2 and 22-02-T2: side-effect marker proving build performs no app launch, plus a process counter proving one launch per app-run request.
- [ ] 22-03-T1: evidence schema/state tests for disabled, incomplete, complete, and capacity exhaustion, including missing and partial report writes.
- [ ] 22-03-T2: positive verifier-only modeled `fixture.value` outcome consumption/comparison plus missing, duplicate, unconsumed, and mismatched outcome controls; this does not execute Lang FFI or local C and does not claim host IO or cleanup.
- [ ] 22-02-T2: build identity invalidation cases for Lang source, C source, header, manifest ABI/symbols, flags/compiler/target, and runtime dependency declarations.
- [ ] No framework installation is required; the Go standard library test framework is already present.

---

## Manual-Only Verifications

All Phase 22 behaviors are intended to have automated verification. Any host or toolchain lane unavailable during execution must be recorded as unavailable and must not be represented as passing evidence.

---

## Security Domain

Security enforcement is enabled at ASVS Level 1; high severity blocks. The plans should carry the corresponding STRIDE threat register and link each mitigation to a concrete task.

| Threat Ref | STRIDE concern | Validation evidence |
|------------|----------------|---------------------|
| T-22-01 | Manifest path traversal or symlink escape; undeclared local build authority | Relocation and path-boundary tests; identity includes normalized declared inputs |
| T-22-02 | Shell injection or repeated application effects | argv-based Clang/process invocation; no-shell tests and exact launch counter |
| T-22-03 | Stale artifact/evidence after source, header, ABI, flag, or dependency changes | Build identity mutation matrix changes the identity for each relevant input |
| T-22-04 | Malformed or oversized input | Length bound, digit validation, U64 overflow control before body execution |
| T-22-05 | Truncated or missing evidence reported as verified success | Sidecar state/control tests keep non-complete states from successful verification |
| T-22-06 | Trusted C violates its declared ABI or behavior | Keep trust boundary and declared inputs reviewable; do not claim arbitrary C correctness |

---

## Validation Sign-Off

- [ ] Every plan task has an automated verification or an explicit Wave 0 dependency.
- [ ] No three consecutive tasks lack automated verification.
- [ ] Wave 0 covers all validation gaps.
- [ ] No watch-mode flags.
- [ ] Cold and warm feedback samples have explicit scope.
- [ ] `nyquist_compliant: true` is set only after validation evidence is complete.

**Approval:** pending

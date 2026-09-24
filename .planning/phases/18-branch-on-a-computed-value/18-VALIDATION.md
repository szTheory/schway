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
| Wave 0 | 01 | 0 | CTL-01 | T-18-01 | Pin the existing computed-scrutinee refusal; independent validators fail closed for malformed or forged core | fixture, unit | `go test ./internal/compiler/check ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |
| Wave 0 | 01 | 0 | CTL-02 | T-18-02 | A Result-returning callee's computed value reaches all five comparator axes through independent peer admission | integration, differential, smoke | `go test ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |
| Wave 0 | 01 | 0 | CTL-03 | T-18-03 | Payload place is returned; injected wrong-slot corruption diverges at `axis:terminal-outcome` | integration, mutation-kill | `go test ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |
| Wave 0 | 01 | 0 | CTL-01 | T-18-01 | A pre-match loan live in one arm is classified with existing endpoints and bounded fixpoint work | source integration, ownership | `go test ./internal/compiler/check ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |
| 18-08-T1 | 08 | 6 | CTL-01, CTL-02, CTL-03 | T-18-01 | Repeated cold/warm distributions and host/tool provenance are recorded and internally consistent; plan 08 will deliver the phase-local evidence verifier | latency evidence | `go test ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |
| 18-08-T2 | 08 | 6 | CTL-01, CTL-02, CTL-03 | T-18-03 | CI disposition is evidence-backed; added command runs in both matrix hosts or unchanged CI blob is proven; plan 08 will deliver the phase-local evidence verifier | CI configuration evidence | `go test ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |

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

## Phase 18 Feedback and CI Evidence Record

Plan 08 fills the fenced key/value record below after measuring each named lane
with at least three cold and three warm runs. Times are seconds; include each
sample and matching count/min/median/max. The helper validates arithmetic and
provenance without imposing a time threshold. Record actual host, Go, Clang,
and pre-decision workflow blob hash; do not leave placeholders in final evidence.

```text
phase18_evidence_version=1
host=
go_version=
clang_version=
focused_command=go test ./internal/compiler/session -count=1
focused_cold_seconds=
focused_cold_count=
focused_cold_min_seconds=
focused_cold_median_seconds=
focused_cold_max_seconds=
focused_warm_seconds=
focused_warm_count=
focused_warm_min_seconds=
focused_warm_median_seconds=
focused_warm_max_seconds=
vet_command=go vet ./...
vet_cold_seconds=
vet_cold_count=
vet_cold_min_seconds=
vet_cold_median_seconds=
vet_cold_max_seconds=
vet_warm_seconds=
vet_warm_count=
vet_warm_min_seconds=
vet_warm_median_seconds=
vet_warm_max_seconds=
build_command=go build ./...
build_cold_seconds=
build_cold_count=
build_cold_min_seconds=
build_cold_median_seconds=
build_cold_max_seconds=
build_warm_seconds=
build_warm_count=
build_warm_min_seconds=
build_warm_median_seconds=
build_warm_max_seconds=
full_test_command=go test ./... -count=1
full_test_cold_seconds=
full_test_cold_count=
full_test_cold_min_seconds=
full_test_cold_median_seconds=
full_test_cold_max_seconds=
full_test_warm_seconds=
full_test_warm_count=
full_test_warm_min_seconds=
full_test_warm_median_seconds=
full_test_warm_max_seconds=
race_command=go test -race ./... -count=1
race_cold_seconds=
race_cold_count=
race_cold_min_seconds=
race_cold_median_seconds=
race_cold_max_seconds=
race_warm_seconds=
race_warm_count=
race_warm_min_seconds=
race_warm_median_seconds=
race_warm_max_seconds=
ci_before_blob=
ci_disposition=pending
ci_command=
ci_decision_rationale=
```

The verifier is phase-local at
`.planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh`.
`--measurements` validates all five lane distributions, actual host/tool
versions, and `ci_before_blob` against `git hash-object .github/workflows/ci.yml`.
`--final` additionally requires `ci_disposition=added` or `not_added`; for
`added`, the exact `ci_command` must appear in the existing `checks` job whose
matrix includes Ubuntu and macOS. For `not_added`, the workflow blob must still
match `ci_before_blob`, and the rationale must explain the measured tradeoff.

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
- [ ] The phase-local evidence verifier passes in both measurement and final-disposition modes.
- [ ] `nyquist_compliant: true` set after Wave 0 and plan mapping are confirmed.

**Approval:** pending plan and Wave 0 validation.

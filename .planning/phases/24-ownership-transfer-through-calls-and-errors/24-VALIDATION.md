---
phase: "24"
slug: "ownership-transfer-through-calls-and-errors"
status: complete
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-30"
updated: "2026-10-01"
---

# Phase 24 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard-library `testing`; native cases compile with installed Clang. |
| **Config file** | `go.mod`; no external test framework is required. |
| **Quick run command** | `sh scripts/verify-phase24.sh` through the existing hosted Ubuntu/macOS evidence aggregate. No project tests may run in this checkout. |
| **Full suite command** | Existing GitHub Actions evidence aggregate, full `go test ./...`, and `go test -race -timeout=20m ./...` on Linux and macOS. Run only through the approved hosted CI workflow. The explicit race timeout accommodates the session package's hosted macOS race runtime while retaining a finite limit. |
| **Measured runtime** | Focused Phase 24 aggregate: 25s Linux/x86_64 and 22s Darwin/arm64. Full hosted gates: 5m44s Linux and 11m30s macOS; evidence aggregates: 8m23s Linux and 13m37s macOS. |

## Sampling Rate

- **After each implementation wave:** Run the relevant Phase 24 focused group through hosted CI on the designated Linux/macOS evidence lane; do not run project tests locally.
- **Before phase verification:** Require the Phase 24 aggregate and existing full/race suite receipts on both required hosts.
- **Max feedback latency:** 13m37s for the macOS evidence aggregate; the focused Phase 24 groups took 22s.

## Per-Requirement Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| T-24-01 | 24-01 | 1 | RES-05 / OWN-10 | T-24-01, T-24-03 | Helper acquisition and owning return preserve obligation; copy, stale use and owning entry result refuse with source spans. | Checker source refusal | `scripts/verify-phase24.sh` — source-refusal group | ✅ present | ✅ pass |
| T-24-02 | 24-01 | 1 | RES-05 / OWN-10 | T-24-01 | Three independent peers rederive acquisition and reject missing, duplicate and wrong release despite checker events. | Independent peer mutations | `scripts/verify-phase24.sh` — independent-peer group | ✅ present | ✅ pass |
| T-24-03 | 24-01 | 1 | RES-05 / OWN-10 | T-24-01, T-24-03 | Ordinary app on caller-selected 0x41/0x42 yields independently fixed 65/66, use after return, exact paired release and emitter refusal. | C emitter / native app | `scripts/verify-phase24.sh` — emitter-admission and native-public-app groups | ✅ present | ✅ pass |
| T-24-04 | 24-02 | 2 | RES-06 / OWN-11 / OWN-12 | T-24-02, T-24-05 | Same helper site succeeds three times before real 0x43 typed failure; failed acquisition creates no owner and source cleanup records C,B,A. | Checker source/error path | `scripts/verify-phase24.sh` — source-refusal and native-public-app groups | ✅ present | ✅ pass |
| T-24-05 | 24-02 | 2 | RES-06 / OWN-11 / OWN-12 | T-24-01, T-24-02, T-24-05 | Independent peers reject activation collision, all-release deletion, wrong pair/order, phantom failed owner and caller/callee scope confusion. | Independent peer mutations | `scripts/verify-phase24.sh` — independent-peer group | ✅ present | ✅ pass |
| T-24-06 | 24-02 | 2 | RES-06 / OWN-12 | T-24-03, T-24-05 | Interpreter and sole C serializer execute typed error and generated reverse cleanup via ordinary app; unsupported shapes refuse before serialization. | Interpreter/C emitter/native app | `scripts/verify-phase24.sh` — model-only, emitter-admission, and native-public-app groups | ✅ present | ✅ pass |
| T-24-07 | 24-03 | 3 | EVD-09 / OWN-11 / RES-06 | T-24-01, T-24-02, T-24-03, T-24-05 | Independent observer proves actual allocation, post-transfer use, C,B,A physical destruction and zero outstanding; five reached controls fail despite plausible events. | Native physical observer | `scripts/verify-phase24.sh` — native-observer group | ✅ present | ✅ pass |
| T-24-08 | 24-03 | 3 | EVD-09 and EVD-11 extension | T-24-02, T-24-04, T-24-06 | Deterministic model-only outcomes retain false physical/IO claims; one focused aggregate carries every Phase 24 group and host identity on Linux/macOS. | Session replay / hosted aggregate | `scripts/verify-phase24.sh` — model-only and public-contract groups; CI 36856048690 | ✅ present | ✅ pass |

*Plan 24-01 selectors were included in the full `go test ./...` suites on both hosted platforms; the full/race suites and regression aggregate passed in run [36780855499](https://github.com/szTheory/schway/actions/runs/36780855499). The focused selectors were not dispatched as separate commands. No project tests run in this checkout. Later plan selectors activate when their implementations and tests land; `scripts/verify-phase24.sh` remains a Plan 24-03 deliverable.*

## Plan 24-01 Hosted Receipt

| Host | Full checks (vet, build, tests, race) | Existing evidence aggregate |
|------|---------------------------------------|-----------------------------|
| Ubuntu | pass, 5m 03s | pass, 8m 46s |
| macOS | pass, 13m 42s | pass, 9m 43s |

The passing retry used HEAD `3b6a2da4` and completed on 2026-09-30. Run `36777742560` previously hit Go's default 10-minute package timeout while the macOS session package was running under race instrumentation; the active test had only just started, and the log contained no race-detector report. The existing CI race command now sets `-timeout=20m`. The retry passed `go vet`, `go build`, full tests and race tests on both hosts, `scripts/verify-phase23.sh`, `scripts/verify-phase6.sh`, and the existing Phase 15 evidence seams. All test execution remained on hosted CI.

## Final Phase 24 Hosted Receipt

Run [36856048690](https://github.com/szTheory/schway/actions/runs/36856048690) validated source revision `ed94ef7972b25deb85ab90fafbf3403dc31629f5` with clean source trees on both hosts. The optional validation-corpus job was skipped by manual dispatch and is outside this phase's acceptance criteria.

| Host | Target / compiler | Full checks (vet, build, tests, race) | Evidence aggregate | Phase 24 focused receipt |
|------|-------------------|--------------------------------------|-------------------|--------------------------|
| Ubuntu Linux/x86_64 | `x86_64-pc-linux-gnu`; Go 1.24.13; Clang 18.1.3 | pass, 5m44s | pass, 8m23s | pass, 25s |
| macOS Darwin/arm64 | `arm64-apple-darwin25.6.0`; Go 1.24.13; Apple Clang 21.0.0 | pass, 11m30s | pass, 13m37s | pass, 22s |

The existing aggregate also passed the Phase 23 and Phase 6 scripts and Phase 15 native/session evidence seams on both hosts. The focused Phase 24 receipt includes source refusal, independent peers, model replay, emitter admission, native apps, physical observer controls, and the public contract. No project tests or scripts were run locally.

## Wave 0 Requirements

No separate infrastructure-install wave was required: Go testing and the existing hosted CI lanes were present, and no new dependency was planned. Phase 24 added the fixtures, independent acquisition-derived validators, model evidence, emitter admission/cleanup, physical observer controls, and one named aggregate. No checks were executed as part of research or planning.

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None planned | — | The phase claims are intended to have automated checker, peer, model, and physical-observer evidence. Review hosted receipts as evidence artifacts; do not substitute a model replay for physical cleanup proof. | — |

## Validation Sign-Off

- [x] Every plan task has an automated verify command or an explicit hosted-CI dependency.
- [x] Sampling continuity: no three consecutive tasks without automated evidence.
- [x] All five reached physical mutation controls are covered by the independent observer.
- [x] Focused validation is owned by one Phase 24 aggregate and does not duplicate the full suite.
- [x] No local project tests are run in this checkout.
- [x] Hosted receipts cover both Linux and macOS before verification.
- [x] `nyquist_compliant: true` set in frontmatter after validation.

**Approval:** hosted validation complete; goal verification is the remaining phase gate.

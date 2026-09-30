---
phase: "24"
slug: "ownership-transfer-through-calls-and-errors"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-30"
---

# Phase 24 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard-library `testing`; native cases compile with installed Clang. |
| **Config file** | `go.mod`; no external test framework is required. |
| **Quick run command** | `sh scripts/verify-phase24.sh` on the designated hosted CI lane after that phase runner is added. No project tests may run in this checkout. |
| **Full suite command** | Existing GitHub Actions evidence aggregate, full `go test ./...`, and `go test -race ./...` on Linux and macOS. Run only through the approved hosted CI workflow. |
| **Estimated runtime** | Not measured for Phase 24. Capture elapsed time in the first hosted receipt; preserve the existing lane and avoid a duplicate full-suite job. |

## Sampling Rate

- **After each implementation wave:** Run the relevant Phase 24 focused group through hosted CI on the designated Linux/macOS evidence lane; do not run project tests locally.
- **Before phase verification:** Require the Phase 24 aggregate and existing full/race suite receipts on both required hosts.
- **Max feedback latency:** Not yet measured. Record actual focused-group and aggregate durations in host receipts.

## Per-Requirement Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| T-24-01 | 24-01 | 1 | RES-05 / OWN-10 | T-24-01, T-24-03 | Helper acquisition and owning return preserve obligation; copy, stale use and owning entry result refuse with source spans. | Checker source refusal | `go test -count=1 -run '^TestPhase24(SourceTransfer\|SourceRefusal)$' ./internal/compiler/check` | ❌ planned | ⬜ pending |
| T-24-02 | 24-01 | 1 | RES-05 / OWN-10 | T-24-01 | Three independent peers rederive acquisition and reject missing, duplicate and wrong release despite checker events. | Independent peer mutations | `go test -count=1 -run '^TestPhase24TransferPeer' ./internal/compiler/session` | ❌ planned | ⬜ pending |
| T-24-03 | 24-01 | 1 | RES-05 / OWN-10 | T-24-01, T-24-03 | Ordinary app on caller-selected 0x41/0x42 yields independently fixed 65/66, use after return, exact paired release and emitter refusal. | C emitter / native app | `go test -count=1 -run '^TestPhase24(EmitterTransfer\|PositiveTransfer)' ./internal/compiler/cgen ./internal/compiler/native` | ❌ planned | ⬜ pending |
| T-24-04 | 24-02 | 2 | RES-06 / OWN-11 / OWN-12 | T-24-02, T-24-05 | Same helper site succeeds three times before real 0x43 typed failure; failed acquisition creates no owner and source cleanup records C,B,A. | Checker source/error path | `go test -count=1 -run '^TestPhase24(ErrorSource\|RepeatedHelperSource)' ./internal/compiler/check` | ❌ planned | ⬜ pending |
| T-24-05 | 24-02 | 2 | RES-06 / OWN-11 / OWN-12 | T-24-01, T-24-02, T-24-05 | Independent peers reject activation collision, all-release deletion, wrong pair/order, phantom failed owner and caller/callee scope confusion. | Independent peer mutations | `go test -count=1 -run '^TestPhase24(ActivationPeer\|CleanupPeer)' ./internal/compiler/session` | ❌ planned | ⬜ pending |
| T-24-06 | 24-02 | 2 | RES-06 / OWN-12 | T-24-03, T-24-05 | Interpreter and sole C serializer execute typed error and generated reverse cleanup via ordinary app; unsupported shapes refuse before serialization. | Interpreter/C emitter/native app | `go test -count=1 -run '^TestPhase24(EmitterError\|NativeError)' ./internal/compiler/cgen ./internal/compiler/native` | ❌ planned | ⬜ pending |
| T-24-07 | 24-03 | 3 | EVD-09 / OWN-11 / RES-06 | T-24-01, T-24-02, T-24-03, T-24-05 | Independent observer proves actual allocation, post-transfer use, C,B,A physical destruction and zero outstanding; five reached controls fail despite plausible events. | Native physical observer | `go test -count=1 -run '^TestPhase24(Observer\|PhysicalDestructorControl)' ./internal/compiler/native ./internal/compiler/cgen` | ❌ planned | ⬜ pending |
| T-24-08 | 24-03 | 3 | EVD-09 and EVD-11 extension | T-24-02, T-24-04, T-24-06 | Deterministic model-only outcomes retain false physical/IO claims; one focused aggregate carries every Phase 24 group and host identity on Linux/macOS. | Session replay / hosted aggregate | `sh scripts/verify-phase24.sh` | ❌ planned | ⬜ pending |

*All commands above are planned for the existing hosted CI lanes only; none has been run in this checkout. `sh scripts/verify-phase24.sh` is a planned focused runner and does not yet exist. Task T-24-08 must include every listed focused group.*

## Wave 0 Requirements

No separate infrastructure-install wave is required: Go testing and the existing hosted CI lanes are present, and no new dependency is planned. The phase plans must add or extend the fixtures, independent acquisition-derived validators, model evidence, emitter admission/cleanup, physical observer controls, and one named Phase 24 CI aggregate before declaring this strategy complete. No checks have been executed as part of research or planning.

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None planned | — | The phase claims are intended to have automated checker, peer, model, and physical-observer evidence. Review hosted receipts as evidence artifacts; do not substitute a model replay for physical cleanup proof. | — |

## Validation Sign-Off

- [ ] Every plan task has an automated verify command or an explicit hosted-CI dependency.
- [ ] Sampling continuity: no three consecutive tasks without automated evidence.
- [ ] All five reached physical mutation controls are covered by the independent observer.
- [ ] Focused validation is owned by one Phase 24 aggregate and does not duplicate the full suite.
- [ ] No local project tests are run in this checkout.
- [ ] Hosted receipts cover both Linux and macOS before verification.
- [ ] `nyquist_compliant: true` set in frontmatter after validation.

**Approval:** pending plan and checker.

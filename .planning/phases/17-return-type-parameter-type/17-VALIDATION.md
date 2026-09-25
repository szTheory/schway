---
phase: "17"
slug: "return-type-parameter-type"
status: validated
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-22"
---

# Phase 17 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/cgen ./internal/compiler/session ./cmd/lang-repair -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~192 seconds |

## Sampling Rate

- **After every task commit:** Run the owning focused `go test` command.
- **After every plan wave:** Run `go test ./... -count=1`.
- **Before `$gsd-verify-work`:** Full suite, four-tier tracer, fixture diagnostic movement, mutation kills, and sealed repair evidence must be green.
- **Max feedback latency:** 192 seconds for the full suite; focused package tests per task.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 17-01-01 | 01 | 1 | TYP-02, TYP-03 | — | Source frontier diagnostics move only to named call-contract causes. | source/integration | `go test ./internal/compiler/check -run '^TestPhase17UseMatchingArgument$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 17-02-01 | 02 | 2 | TYP-01, TYP-04 | — | Parameter and return facts are directional and peers independently derive abilities. | unit/mutation/import | `go test ./internal/compiler/corevalidate -run '^TestPhase17CorePeerDirectionalCallContract$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 17-03-01 | 03 | 3 | TYP-01 | — | The sole emitter preserves distinct parameter and return C types across all tiers. | C structure/differential | `go test ./internal/compiler/session -run '^TestPhase17TwoTypeFourTierDifferential$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 17-04-01 | 04 | 4 | TYP-05 | — | A sealed held-out source repairs through protocol fields without fixture-path coupling. | subprocess/integration | `go test ./internal/compiler/session -run '^TestPhase17RepairCorpusReachable$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

## Wave 0 Requirements

- [ ] Parser-valid refused-frontier fixtures with old diagnostic pins and movement assertions.
- [ ] Canonical nominal two-type tracer, generated-C structure assertions, and four-tier comparison.
- [ ] Per-layer return-only and coordinated mutations plus import-boundary guards.
- [ ] Distinct derivation/held-out repair fixtures, SHA-256 seal, and protocol-field red controls.

## Manual-Only Verifications

All phase behaviors have automated verification.

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies.
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify.
- [ ] Wave 0 covers all MISSING references.
- [ ] No watch-mode flags.
- [ ] Feedback latency is bounded by the recorded full-suite baseline.
- [ ] `nyquist_compliant: true` set in frontmatter.

**Approval:** pending

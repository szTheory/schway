---
phase: "16"
slug: "branch-match-emitter-port"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-19"
---

# Phase 16 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` |
| **Config file** | `go.mod` |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase16-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen/... 'TestN1ConvergenceDifferential|Test.*Program'` |
| **Full suite command** | `env GOCACHE=/tmp/ai-lang-phase16-cache go test ./...` |
| **Estimated runtime** | ~192 seconds |

## Sampling Rate

- **After every task commit:** Run the focused `scripts/assert-go-tests.sh` cgen command for the affected behavior.
- **After every plan wave:** Run `env GOCACHE=/tmp/ai-lang-phase16-cache go test ./...`.
- **Before `$gsd-verify-work`:** The full suite must be green; run the exact-shape `restrict` lane on macOS and Linux before any by-pointer admission decision.
- **Max feedback latency:** 30 seconds for focused cgen checks.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 16-01-01 | 01 | 1 | NAT-08 | T-16-01 | N=1 legacy and whole-program paths are byte-identical for scoped fixtures in both public modes. | unit + golden | `env GOCACHE=/tmp/ai-lang-phase16-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen/... 'TestN1ConvergenceDifferential|Test.*Program'` | ❌ W0 expansion | ⬜ pending |
| 16-01-02 | 01 | 1 | NAT-08 | T-16-02 | Admission preserves graph/entry → shape → preflight → serialization ordering. | unit + mutation | `env GOCACHE=/tmp/ai-lang-phase16-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen/... 'Test.*Program|TestN1ConvergenceDifferential'` | ❌ W0 port cases | ⬜ pending |
| 16-02-01 | 02 | 2 | NAT-08 | T-16-03 | Golden-change ledger bijects with the digest map and rejects stale or duplicate entries. | unit | `go test ./internal/compiler/core/... -run 'TestPreviousPhaseGoldenCUnchanged|Test.*Golden.*Ledger' -count=1` | ❌ W0 ledger | ⬜ pending |
| 16-03-01 | 03 | 3 | NAT-08 | T-16-04 | Public native dispatch has no function-count route and production lowering uses `emitProgram`. | structural + unit | `go test ./internal/compiler/cgen/... -run 'Test.*Dispatch|TestN1ConvergenceDifferential' -count=1` | ❌ W0 structural control | ⬜ pending |
| 16-04-01 | 04 | 4 | NAT-08 | T-16-05 | Interpreter, `-O0`, `-O3`, and `-O3 -flto` remain semantically aligned. | integration | `go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential -count=1` | ✅ | ⬜ pending |
| 16-05-01 | 05 | 4 | NAT-09 | T-16-06 | Amendment and M004 debt records preserve owner, prerequisite, reopening, and LTO consequence. | document + unit | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -count=1` | ❌ W0 Phase-16 rows | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

## Wave 0 Requirements

- [ ] Add both-mode N=1 byte-identity rows and retained scoped-refusal rows.
- [ ] Add golden-change ledger parser, bijection, current-digest, duplicate, and stale negative controls.
- [ ] Add a structural control proving public dispatch has no function-count route and only `emitProgram` is a production route.
- [ ] Add Phase-16 debt-register and amendment assertions for M004 owner, prerequisite, reopening condition, and LTO consequence.
- [ ] Add an exact-shape `restrict` probe harness and extension-refusal matrix if D-16-07 is considered for admission.

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Linux exact-shape `restrict` lane | NAT-08 | The current research environment lacks Linux host evidence. | Run the unchanged one-TU microprogram at `-O0`, `-O3`, and `-O3 -flto` with its sanitizer lane on Linux; record the result before admitting by-pointer lowering. |

## Validation Sign-Off

- [ ] All tasks have `<automated>` verification or Wave 0 dependencies.
- [ ] Sampling continuity has no three consecutive tasks without automated verification.
- [ ] Wave 0 covers all missing verification references.
- [ ] No watch-mode flags.
- [ ] Focused feedback latency is under 30 seconds.
- [ ] `nyquist_compliant: true` set in frontmatter.

**Approval:** pending

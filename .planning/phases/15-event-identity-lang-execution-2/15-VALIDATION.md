---
phase: "15"
slug: "event-identity-lang-execution-2"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-19"
---

# Phase 15 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 standard `testing` package |
| **Config file** | none |
| **Quick run command** | `go test ./internal/compiler/{execution,interp,cgen,native,session} -run 'Test(.*Invocation.*|.*DiamondSharedLeaf.*|.*PathTable.*|.*ComparisonFieldRouting.*)' -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | Measure during execution; Phase 14 observed the full suite near 193 seconds on its host |

---

## Sampling Rate

- **After every task commit:** Run the narrow package/test command named by that task.
- **After every plan wave:** Run `go test ./...`.
- **Before `$gsd-verify-work`:** Full suite must be green.
- **Max feedback latency:** Record cold and warm distributions; do not substitute a guessed single number.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 15-01-01 | 01 | 1 | OBS-04 | T-15-02 | Reject ambiguous or non-canonical `/2` identity | unit/golden | `go test ./internal/compiler/execution ./internal/compiler/native -count=1` | ❌ W0 `/2` cases | ⬜ pending |
| 15-02-01 | 02 | 1 | OBS-01, OBS-02 | T-15-03 | Interpreter emits correctly owned preorder edges | unit | `go test ./internal/compiler/interp -count=1` | ❌ W0 | ⬜ pending |
| 15-03-01 | 03 | 1 | OBS-03 | T-15-03 | Independent peer rejects forged causal structure | peer/mutation | `go test ./internal/compiler/executionpeer -count=1` | ❌ W0 | ⬜ pending |
| 15-04-01 | 04 | 2 | OBS-01, OBS-04 | T-15-01 | Native preflight bounds expansion before allocation | boundary/mutation | `go test ./internal/compiler/cgen -count=1` | ❌ W0 | ⬜ pending |
| 15-05-01 | 05 | 2 | OBS-02, OBS-03 | T-15-03, T-15-04 | Comparator routes new fields and both fault directions fail | differential | `go test ./internal/compiler/session -count=1` | ❌ W0 | ⬜ pending |
| 15-06-01 | 06 | 3 | OBS-01, NAT-10 | T-15-04 | Shared-leaf fixture agrees across four tiers and collision return fails | integration | `go test ./internal/compiler/session -run '^TestPhase11InterproceduralDifferential$' -count=1 -v` | ✅ expectation flips | ⬜ pending |

*Task/plan allocation is provisional until PLAN.md files are finalized. Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky.*

---

## Wave 0 Requirements

- [ ] Invocation grammar canonical/non-canonical byte cases: UTF-8, `%`, `/`, `#`, empty fields, and malformed ordinals.
- [ ] Frozen `/0` and `/1` canonical-byte corpus pin before producer changes.
- [ ] Peer structural import guard and named actionable error assertions.
- [ ] Static-table 61/4096/4097 boundary plus preflight-bypass mutation proof.
- [ ] `function.called` projection removal control and strict preorder assertion.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Threat Coverage

| Ref | Threat | Mitigation / evidence |
|-----|--------|-----------------------|
| T-15-01 | Exponential path-table expansion DoS | Checked/saturating preflight; 4096 accepted, 4097 named refusal, no override; mutation-kill bypass. |
| T-15-02 | Ambiguous delimiter or non-canonical path | One strict byte escape and parser; reject alternate spellings. |
| T-15-03 | Forged/reparented causal edge | Independent peer checks membership, binding, ownership, and preorder. |
| T-15-04 | Silent evidence weakening | Frozen legacy bytes, fail-closed comparator routing, flipped collision control, bidirectional seeded faults. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency distributions recorded honestly
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

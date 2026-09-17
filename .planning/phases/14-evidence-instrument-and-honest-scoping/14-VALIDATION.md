---
phase: "14"
slug: "evidence-instrument-and-honest-scoping"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-17"
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) |
| **Config file** | none — `go.mod` at repo root |
| **Quick run command** | `go test ./<changed-package>/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~192.7s full suite (recorded baseline; EVD-06 re-measures and records it as an `observed` row — do NOT copy the stale ~60s figure from docs) |

**Baseline:** `go test ./...` is green (exit 0, 25 packages) on the current tree
before any Phase 14 change — confirmed during research. Any red is caused by
this phase.

---

## Sampling Rate

- **After every task commit:** Run the quick command for the touched package
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 200 seconds (full suite); < 10s for a single package

---

## Per-Task Verification Map

> Seeded by `/gsd-plan-phase`; the planner fills one row per task from PLAN.md.
> **Phase-specific law (the whole point of this phase):** every `Automated Command`
> in this table MUST be executed and MUST resolve to at least one test before it is
> recorded here. A `go test -run` pattern that matches zero tests is the exact defect
> EVD-01 exists to catch — authoring one here would be self-refuting.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 14-01-01 | 01 | 1 | EVD-01 | — | N/A | unit | `{command}` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Filled by the planner. Existing `go test` infrastructure covers all phase
requirements — no framework install is needed; Wave 0 is limited to any new
test files the plans introduce.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | — |

*Target: all phase behaviors have automated verification. A row added here needs
an explicit justification — this phase's premise is that reviewer judgment is the
failure mode being engineered out.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Every `<automated>` command was executed and matched ≥1 test (no dead `-run` patterns)
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 200s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

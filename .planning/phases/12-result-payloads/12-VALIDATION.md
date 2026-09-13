---
phase: "12"
slug: "result-payloads"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-12"
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` (`go test`) — no third-party framework anywhere in `internal/compiler` |
| **Config file** | none — plain `_test.go` files colocated with their packages |
| **Quick run command** | `go test ./internal/compiler/<package>/... -run <TestName> -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60 seconds (full suite) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/compiler/<touched-package>/... -count=1`, plus `go vet ./...` and `gofmt -l .`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite green, plus the four pinned golden-C digests (`core_test.go:156-159`) byte-identical
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 12-01-01 | 01 | 1 | RES-02 | — | N/A | unit | `{planner fills}` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

*Seeded as draft by plan-phase; the planner's PLAN.md task IDs and `<automated>` commands populate this table.*

---

## Wave 0 Requirements

- [ ] `internal/compiler/core/` — fixtures exercising `OpConstructPayload` / `OpDestructurePayload` at all six `TestAllOperationKindsHandledAtEverySite` dispatch sites (RES-02)
- [ ] Corpus characterization replay (D-12-19) — byte-identical execution documents across the whole pre-Phase-12 fixture corpus
- [ ] `testdata/phase12/` — frozen struct-shape mutation fixture(s) for D-12-37's `control:foreign.layout_mismatch`-style control, plus its `_Static_assert`-killing companion test
- [ ] Reverted-production-hunk mutation test for `cgen`'s match-arm/constructor codegen (D-12-38), plus the unmutated cross-fixture companion run
- [ ] N=1 convergence differential test gating the D-11-02 six-emitter deletion (D-12-31 step i)
- [ ] Framework install: none needed — `go test` is already the only test framework

**Already green (not Wave 0):**
- `go test ./internal/compiler/corevalidate/... -run TestC03ResultPayloadOriginAcrossOpCall` — the criterion-3 pre-flight probe (BRANCH A accepted)
- `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed` — mechanism exists; the D-12-24 debt entry is Wave 0 *content*, not new test code

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Niche optimization is designed-but-uninstantiable | RES-03 | No niche exists with only `Byte`/`Buffer` as payload types — an absence-of-applicable-input finding, not a runnable assertion | Confirm the debt register (D-12-24) records the absence explicitly rather than silently skipping |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

---
phase: "12"
slug: "result-payloads"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: true
created: "2026-09-12"
evidence_vocabulary: v1
graded_rows: 15
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

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| 12-01-01 | 01 | 1 | NAT-04..NAT-07 | — | N=1 convergence differential committed permanently | unit | `go test ./internal/compiler/cgen/... -run 'TestN1ConvergenceDifferential' -v -count=1` | ✅ | EXERCISED | — |
| 12-01-02 | 01 | 1 | NAT-04..NAT-07 | — | D-12-31 branch ratified (checkpoint:decision) | manual | N/A — developer ratification recorded in `12-01-SUMMARY.md` | ✅ | DEFINED | — |
| 12-01-03 | 01 | 1 | RES-02, RES-03 | — | Ratified outcome + niche finding recorded in mechanically-checked debt registers | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed' -v -count=1` | ✅ | EXERCISED | — |
| 12-02-01 | 02 | 2 | RES-02, RES-03 | — | Published surface ratified (checkpoint:decision) | manual | N/A — developer ratification recorded in `12-02-SUMMARY.md` | ✅ | DEFINED | — |
| 12-02-02 | 02 | 2 | RES-02 | T-12-32 | Two new `core.OperationKind`s, AST binder, parser grammar | unit | `go test ./internal/compiler/core/... -run 'TestAllOperationKinds' -v -count=1` | ✅ | EXERCISED | — |
| 12-02-03 | 02 | 2 | RES-02, RES-03 | T-12-31..T-12-34 | Payload round-trip: interpreter/-O0/-O3 agreement on the tracer fixture | unit | `go test ./internal/compiler/cgen/... -run 'TestPayloadTracer' -v -count=1` | ✅ | EXERCISED | — |
| 12-03-01 | 03 | 3 | RES-02 | — | Three named pattern refusals (arity mismatch, binder-on-nullary, missing-binder) | unit | `go test ./internal/compiler/check/... -run 'TestPayloadPatternRefusals' -v -count=1` | ✅ | EXERCISED | — |
| 12-03-02 | 03 | 3 | RES-02 | T-12-27..T-12-30 (12-CONTEXT) | D-12-27's fail-closed resource-payload refusal | unit | `go test ./internal/compiler/check/... -run 'TestResourcePayloadRefused' -v -count=1` | ✅ | EXERCISED | — |
| 12-03-03 | 03 | 3 | RES-02 | — | Affine drop obligation witness + real exhaustive-dispatch coverage | unit | `go test ./internal/compiler/core/... -run 'TestAllOperationKindsHandledAtEverySite' -v -count=1` | ✅ | EXERCISED | — |
| 12-04-01 | 04 | 4 | RES-02 | — | `originvalidate`'s payload arm; `corevalidate`'s corpus-wide `AlternativeDetails` invariant | unit | `go test ./internal/compiler/originvalidate/... ./internal/compiler/corevalidate/... -count=1` | ✅ | WIRED | — |
| 12-04-02 | 04 | 4 | RES-03 | T-12-36 (12-CONTEXT) | Corpus characterization replay — interp widening moved zero bytes | unit | `go test ./internal/compiler/session/... -run 'TestPayloadCorpusCharacterizationReplay' -v -count=1` | ✅ | EXERCISED | — |
| 12-04-03 | 04 | 4 | RES-03 | — | `pathoracle` termination on payload shapes; D-12-04c header correction | unit | `go test ./internal/compiler/pathoracle/... -run 'TestPayloadPathEnumerationTerminates' -v -count=1` | ✅ | EXERCISED | — |
| 12-05-01 | 05 | 5 | RES-03 | T-12-31, T-12-32 | D-12-37 frozen-fixture payload layout control (necessary, not sufficient) | unit | `go test ./internal/compiler/session/... -run 'TestPayloadLayoutMutationRefused' -v -count=1` | ✅ | EXERCISED | — |
| 12-05-02 | 05 | 5 | RES-03 | T-12-33, T-12-34, T-12-37 | D-12-38 decisive slot-swap mutation control — escalated as D-12-43 finding | unit | `go test ./internal/compiler/session/... -run 'TestPayloadSlotSwapMutationKilled' -v -count=1` | ✅ | EXERCISED | — |
| 12-05-03 | 05 | 5 | RES-03 | T-12-35 | Criterion 2's findings (niche uninstantiability, "one meaning" scope) recorded in `PHASE-12-DEBT.md` | unit | `go test ./internal/compiler/session/... -run 'TestDebtRegistersAreWellFormed' -v -count=1` | ✅ | EXERCISED | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/compiler/core/` — fixtures exercising `OpConstructPayload` / `OpDestructurePayload` at all six `TestAllOperationKindsHandledAtEverySite` dispatch sites (RES-02) — plan 02/03
- [x] Corpus characterization replay (D-12-19) — byte-identical execution documents across the whole pre-Phase-12 fixture corpus — plan 04
- [x] `testdata/phase12/` — frozen struct-shape mutation fixture(s) for D-12-37's `control:foreign.layout_mismatch`-style control, plus its `_Static_assert`-killing companion test — plan 05 Task 1
- [x] Reverted-production-hunk mutation test for `cgen`'s match-arm/constructor codegen (D-12-38), plus the unmutated cross-fixture companion run — plan 05 Task 2 (built; escalated as D-12-43 since the decisive value-divergence form is unconstructible against the current representation — see `PHASE-12-DEBT.md`)
- [x] N=1 convergence differential test gating the D-11-02 six-emitter deletion (D-12-31 step i) — plan 01 Task 1
- [x] Framework install: none needed — `go test` is already the only test framework

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

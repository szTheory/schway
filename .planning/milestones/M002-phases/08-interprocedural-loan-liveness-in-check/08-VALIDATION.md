---
phase: "08"
slug: "interprocedural-loan-liveness-in-check"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: true
created: "2026-09-09"
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `08-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go's built-in `testing` package (stdlib) |
| **Config file** | none — standard `go test` |
| **Quick run command** | `go test ./internal/compiler/check/...` |
| **Full suite command** | `go test ./... && go vet ./...` |
| **Estimated runtime** | ~60 seconds (full suite) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/compiler/check/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green, AND the mid-phase gate
  (criterion 1's adversarial composition-depth corpus + criterion 3's cost
  measurement) adjudicated from code-level evidence before the liveness law is
  declared final.
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

> Task IDs are assigned by the planner; the rows below are the requirement-level
> verification contract every task must map onto. `File Exists` marks whether the
> named test already exists in-tree (✅) or is authored by this phase (❌ + wave).

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | TBD | 1 | OWN-06 | — | Both admission paths derive identical loan last-uses (D-07-49 entry defect fixed in both) | unit (differential) | `go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree` | ❌ Wave 1 | ⬜ pending |
| TBD | TBD | 1 | OWN-06 | — | Native-recursion→explicit-stack conversion preserves fixpoint behavior | unit (existing) | `go test ./internal/compiler/check/... -run TestLoanLivenessFixpoint` | ✅ exists | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | Refuse-case twin Pattern A (ReturnsBorrowOfParam) refused, safe twin admitted | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessTwinPatternA` | ❌ Wave 2 | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | Refuse-case twin Pattern B (UsesParam) refused, safe twin admitted | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessTwinPatternB` | ❌ Wave 2 | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | Composition-depth-≥2 relay chain resolved transitively via summaries | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessRelayDepth2` | ❌ Wave 2 | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | `relay_escort_witness.lang` flips clean→refused | unit (existing, assertion flips) | `go test ./internal/compiler/check/... -run TestRelayEscortWitness` | ✅ exists | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | Negative control: `Fails`/`Foreign.*` summary fields do not move the verdict | unit | `go test ./internal/compiler/check/... -run TestInterproceduralLivenessNegativeControl` | ❌ Wave 2 | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | Summary memo is rebuilt per invocation, never persisted across sessions | unit | `go test ./internal/compiler/check/... -run TestSummaryMemoNeverPersisted` | ❌ Wave 2 | ⬜ pending |
| TBD | TBD | 2 | OWN-06 | — | Summary derivation requires program order (invariant pinned) | unit | `go test ./internal/compiler/check/... -run TestSummaryDerivationRequiresProgramOrder` | ❌ Wave 2 | ⬜ pending |
| TBD | TBD | 3 | OWN-06 | — | Fail-closed iteration bound produces a named refusal, not a hang or silent under-approximation; seeded seam mutation is killed | unit + mutation-kill | `go test ./internal/compiler/check/... -run TestLoanLivenessBoundMutationKilled` | ❌ Wave 3 | ⬜ pending |
| TBD | TBD | 4 | EFF-02 | — | Growth exponent ≤ 1.2 vs operation count across all required call-graph shapes | integration | `go test ./internal/compiler/session/... -run TestQLT02InterproceduralGrowthExponent` | ❌ Wave 4 | ⬜ pending |
| TBD | TBD | 4 | EFF-02 | — | Chokepoint widening preserves exactly-one-promotion-passthrough (re-derived, not merely re-run) | unit (existing) | `go test ./internal/compiler/measure/... -run TestDemoteHasExactlyOnePromotionPassthrough` | ✅ exists | ⬜ pending |
| TBD | TBD | 4 | EFF-02 | — | Feedback-budget manifest row ratified with `machine_id` | integration | `go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest` | ✅ framework exists, new row | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements. `check_test.go` already
supplies the established patterns for work-ratio assertions, unexported
fault-injection seams, and corpus-driven admission tests; `measure`/`session`
already supply the manifest and demotion test scaffolding. No new framework
install or shared fixture setup is required before implementation begins.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Criterion 4 disclosure is human-readable — a reader can tell, from the shipped artifact, which callee-signature fields the liveness answer depended on at a given call site | OWN-06 | Legibility of a rendered artifact is a judgement about the reader's experience, not an assertion about bytes; the field *set* is machine-checked, its readability is not | Run `check` on the depth-≥2 relay corpus, open the emitted artifact, and confirm that for a chosen call site the disclosed callee-signature fields are named and attributable without consulting source |
| Mid-phase gate adjudication — criterion 1's adversarial corpus and criterion 3's cost curve reviewed from code-level evidence before the liveness law is declared final and before `corevalidate`'s peer is planned | OWN-06, EFF-02 | A gate is a human decision to proceed, take debt, or stop; it consumes automated evidence but is not itself an assertion | Present the corpus results and the growth-exponent fit; adjudicate each open item explicitly as resolved or recorded debt — never assumed safe |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references (none — infrastructure exists)
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

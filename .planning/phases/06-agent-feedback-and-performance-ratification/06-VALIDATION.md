---
phase: "6"
slug: "agent-feedback-and-performance-ratification"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-06"
---

# Phase 6 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go's built-in `testing` package — no third-party test framework anywhere in the tree (`go.mod` has no `require` block) |
| **Config file** | none — `go test` driven directly; `scripts/assert-go-tests.sh` wraps exact-test selection with a nonexistent-sentinel guard that refuses a test name the package does not actually declare |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/<pkg>/... <ExactTestName> [...]` |
| **Full suite command** | `go test ./... && go test -race ./... && go vet ./...` |
| **Estimated runtime** | ~60–90 seconds for the full suite; per-task exact-selection runs are seconds |

**Why `assert-go-tests.sh` and not bare `go test -run`:** a bare `-run` regex that matches
nothing exits 0. The wrapper lists the package's real targets first and fails when a requested
name was never discovered, so a typo or a deleted test can never render as green. Every Phase 6
task's `<automated>` command must go through it, matching the Phase 5 precedent verbatim.

---

## Sampling Rate

- **After every task commit:** Run the task's exact-selection `assert-go-tests.sh` command
- **After every plan wave:** Run `go test ./... && go test -race ./... && go vet ./...`
- **Before `/gsd-verify-work`:** Full suite green **and** `sh scripts/verify-phase6.sh` green
- **Max feedback latency:** 90 seconds

---

## Per-Task Verification Map

Task IDs are assigned by the planner; this table records the requirement→test-type→command
binding each task must satisfy. The planner fills task IDs and plan/wave columns.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | TBD | 0 | FND-04 | — | Metrics/Lane fields never reach content identity, so measurement cannot alter semantic output | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/protocol/... TestMetricsAndLaneFieldsExcludedFromIdentity` | ❌ W0 | ⬜ pending |
| TBD | TBD | 0 | FND-04, DX-03, QLT-02 | — | Prior-phase bytes frozen before any schema bump lands | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestPreviousPhaseCoreBytesUnchanged TestPreviousPhaseManifestIDsUnchanged TestPreviousPhaseGoldenCUnchanged` | ✅ | ⬜ pending |
| TBD | TBD | — | DX-02 | T-06-EXPLAIN | Bounded output; truncation code emitted rather than unbounded expansion | unit + CLI golden | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... <Explain/Query test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-02 | — | Synthesized cause DAG is deterministic across cold invocations | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... <determinism test name>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-03 | T-06-CACHE | Undeclared cache input cannot produce a hit; deferred lane never renders `pass` | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/cache/... <cache key + fail-closed test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-03 | — | Risk→lane table selection widens on unclassified input | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... <risk-lane selection + widening test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | QLT-02 | — | Budget manifest audit refuses a manifest whose declared machines don't match live probe output | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... <budget audit test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | QLT-02 | — | CoV auto-demotion: a noisy metric becomes `observed`, never `blocking` | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/<stats pkg>/... <CoV demotion test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-04 | T-06-REPAIR | Driver repairs via structured channel only; never scrapes prose | subprocess CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/... <repair driver test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-04 | — | **Prose-scramble guard:** identical outcome with all `message`/`detail` replaced by lorem ipsum | subprocess CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/... <prose scramble test name>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-04 | — | **Vocabulary-removal guard:** driver goes RED when `repairs[]`/`kind`/`span`/`replacement` are stripped and prose is left intact | subprocess CLI | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/... <vocabulary removal test name>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-04 | — | **Marker mutation-kill:** each injector refuses when its target marker disappears | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./internal/compiler/session/... <marker guard test names>` | ❌ W0 | ⬜ pending |
| TBD | TBD | — | DX-04 | T-06-BOUNDARY | Repair driver import boundary: build fails if it imports anything under `internal/` | unit | `env GOCACHE=/tmp/ai-lang-phase6-cache sh scripts/assert-go-tests.sh ./cmd/... <import boundary test name>` | ❌ W0 | ⬜ pending |
| TBD | TBD | final | all five | — | Phase gate | shell | `sh scripts/verify-phase6.sh` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/compiler/protocol/protocol_test.go` — add `TestMetricsAndLaneFieldsExcludedFromIdentity`
      **before any new `Metrics`/`Lane` field is added.** Research confirmed `Result.Finalize()`'s
      identity struct already excludes `Metrics` entirely and reduces `Lane` to `ID+":"+Status`,
      so D-06-32 holds today — this test is what keeps it holding once fields are added.
- [ ] Re-pin prior-phase frozen bytes first (the Phase 5 `05-01` precedent):
      `TestPreviousPhaseCoreBytesUnchanged`, `TestPreviousPhaseManifestIDsUnchanged`,
      `TestPreviousPhaseGoldenCUnchanged` must be green before the coordinated
      `lang.command/0`→`/1` and `lang.verify-lane/0`→`/1` bump lands.
- [ ] `internal/compiler/cache/` — brand-new package, zero existing tests; needs `cache_test.go`
      from the first task that creates it.
- [ ] `internal/compiler/session/session_phase6_*_test.go` — sibling test files per the Phase 5
      convention, one per new surface (explain/query, cache wiring, budget).
- [ ] A Go home for p50/p95/CoV. Research found the "Phase 2 20-sample machinery" is **shell**
      (`scripts/verify-phase2.sh`'s `observe()`), not a Go library — no Go package computes these
      today. D-06-19's CoV auto-demotion needs new, unit-testable Go code.
- [ ] `cmd/<repair-driver>/` — new binary plus its import-boundary test.
- [ ] Held-out `.lang` defect fixture corpus (D-06-29), structurally separate from any fixtures
      used to hand-derive the driver's kind→edit mapping. None exists yet.
- [ ] `scripts/verify-phase6.sh` — following `verify-phase5.sh`'s exact shape: self-test sentinel,
      `GOCACHE` pinned to a disposable temp dir, `go test ./...` → race → vet → build, per-corpus
      JSON captured then grepped for required `control:*` strings, expected-escapes grepped
      separately, then this phase's own 20-sample `observe()`-style loop.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| The **recorded agent-legibility exercise** (D-06-30): repair rounds against the `p95: 2` intent, token/tool-call cost, and whether protocol-only access sufficed | DX-04 | Explicitly **non-gating by decision**. A live LLM in the suite would break the offline / deterministic / stdlib-only constraints (D-06-23). Runs on a per-milestone or on-demand cadence. | Drive the shipped `lang` binary through the five defect classes using only `--json` output and the structured `Repair` records. Record rounds, tokens, and tool calls. Report honestly-unmeasured values rather than fabricating them. Store alongside other phase evidence artifacts. |
| Wall-clock cold/warm p50/p95 **ratification** on a newly declared machine | QLT-02 | Ratifying a *new* `machine_id` into `qlt02_budget_manifest.json` is a reviewed human act by design — the JSON diff is the review surface (D-06-16), and an auto-ratifying gate would defeat it. The *audit* of an already-ratified manifest is automated. | Run the 20-sample loop on the target host, inspect the reported distribution and computed `machine_id`, then commit the manifest row with `ratified_at` and `ratified_by_commit`. On an undeclared host the run is observation-only (`ratified: false`) and writes nothing (D-06-18). |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

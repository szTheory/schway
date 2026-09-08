---
phase: "07"
slug: "calls-signatures-and-call-graph-refusal"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-08"
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded from `07-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` (`go test ./...`) — no external test framework in this repo, and none is to be added (zero-external-production-dependency record; `pgregory.net/rapid` is test-only and optional) |
| **Config file** | none — table-driven Go tests, fixtures under `testdata/phaseN/` |
| **Quick run command** | `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/core/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | quick ~seconds; full suite dominated by native `-O0`/`-O3`/`-flto` lanes |

---

## Sampling Rate

- **After every task commit:** targeted package tests for the file(s) touched — `go test ./internal/compiler/<package>/...`
- **After every plan wave:** `go test ./...` (the project documents no separate slow/fast tiers)
- **Before `/gsd-verify-work`:** full suite green, **plus both exhaustive-dispatch controls** — the in-process `core_test.go` control and the CLI-observable `lane:kind-exhaustive-dispatch` equivalent, each carrying its own new Phase 07 fixture list (see Wave 0; these are literal phase-scoped lists, not generic sweeps)
- **Max feedback latency:** quick command, seconds

---

## Per-Task Verification Map

Populated by the planner per task. Requirement-level map seeded from research:

| Req ID | Behavior | Test Type | Automated Command | File Exists |
|--------|----------|-----------|-------------------|-------------|
| SEM-04 | `OpCall` handled at all six dispatch sites; both exhaustive-dispatch controls green | unit + CLI | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandledAtEverySite` and the session lane assertions | ❌ W0 — needs new Phase 07 fixtures added to **both** lists |
| SEM-05 | Digest-bound `lang.interface/1` carries everything a caller needs; no caller path reads a callee body | unit | new `originvalidate` / `corevalidate` tests over the Stage 0 gate corpus (producer-vs-peer zero-divergence sweep) | ❌ W0 |
| SEM-06 | callable ⊆ publishable; refusal carries a stable `core.*` code | unit | new `check` / `originvalidate` test using a clean-but-uncallable fixture | ❌ W0 — **the fixture does not exist and must be authored first** |
| SEM-07 | Direct / mutual / indirect cycles refused, bounded, never hang | unit | new `internal/compiler/callgraph` tests + `check`-level integration over diamond/shared-leaf and deep-chain corpora | ❌ W0 — package does not exist yet |
| QLT-08 | Every new interprocedural control observed to fail against a seeded mutation in the plan that introduced it | unit | `TestXMutationKilled` / `TestXMutationMatrix` naming convention, e.g. `go test ./internal/compiler/callgraph/... -run MutationKilled` | ❌ W0 — **no QLT-08-scoped registry exists**; `session/qlt01_registry.json` is a different M001-era spike-descendant registry and is **not** reusable here |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `testdata/phase07/*.lang` — the two-function call corpus: basic call; **diamond + shared leaves, deep and repeated, parser-shaped** (load-bearing — a chain-only corpus cannot kill the gray/black mutation, per CONTEXT.md `<specifics>`); self-call; unreachable cycle; foreign-symbol shadowing; unresolvable-callee forged artifact; **calls from both arms of an existing `match`** (D-07-28 `Result`-blindness guard)
- [ ] A clean-but-uncallable fixture for SEM-06 — **does not exist under any name today** (CONTEXT.md A-03)
- [ ] The `relay` / `escort` two-function dangling-alias witness quoted in D-04-03's narrative — **does not exist under any name today** (CONTEXT.md A-03)
- [ ] `internal/compiler/callgraph/callgraph.go` + `callgraph_test.go` — the package does not exist
- [ ] A new Phase-07-scoped block in `session.go` mirroring `:2501-2568`, with its own fixture list and required-kinds check
- [ ] A synthetic `core.Program` builder helper for `corevalidate`'s peer tests — **hand-built, never through the parser** (D-07-19; building them through the parser is what would make the peer inert)
- [ ] A pinned `InterfaceV0` decode struct + frozen-bytes fixture proving no `lang.interface/0` byte moved (D-07-08)
- [ ] Framework install: **none.** Stdlib `testing` covers everything.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Bilateral seeded fault (D-07-24 fault 3) | QLT-08 | The gate must **FAIL** when the fault is injected into producer *and* peer — a passing automated assertion of a deliberately-failing gate is itself the thing under test | Inject the empty-abilities fallback into both `BuildInterface` and the `corevalidate` peer; confirm the sweep reports "no divergence detected under bilateral fault" rather than green |

*All other phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references (note: **three** named artifacts in CONTEXT.md/D-04-03 do not exist and must be authored)
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s for the quick command
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

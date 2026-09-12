---
phase: "11"
slug: "multi-function-native-emission-and-interprocedural-equivalen"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-11"
---

# Phase 11 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded by `/gsd-plan-phase 11` from `11-RESEARCH.md` § Validation Architecture.
> The Per-Task Verification Map is filled in by the planner / validate-phase once
> PLAN.md task IDs exist.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go's built-in `testing` package (`go test`) — no external test framework; consistent with this repo's zero-external-dependency record |
| **Config file** | none — no test-runner config beyond the module's own `go.mod` |
| **Quick run command** | `go test ./internal/compiler/<touched-package>/...` (package-scoped) |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60 seconds (full suite); package-scoped runs are seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/compiler/<touched-package>/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd-verify-work`:** Full suite must be green, AND the mid-phase gate's
  four-part conjunction (D-11-14) must be green before any attribute-emission work
  is attempted. Per D-11-09 no attribute emission is planned this phase, but the
  gate's structural floor — ≥2 functions, ≥1 call edge, N≥1 would-have-carried-`restrict`
  — must still be demonstrated non-vacuous.
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| *(pending — filled from PLAN.md task IDs)* | | | | | | | | | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

### Requirement → Test Map (from RESEARCH.md)

| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|--------------------|-------------|
| NAT-04 | `cgen.Emit`/`EmitNative` accept N>1 functions and emit valid C17 | unit + integration | `go test ./internal/compiler/cgen/... -run TestEmitProgram -v` | ❌ Wave 0 |
| NAT-05 | Emitted artifact carries a generated comment naming the empty attribute set and citing D-11-09 | unit | `go test ./internal/compiler/cgen/... -run TestEmittedAttributeSetIsExplicitlyEmpty -v` | ❌ Wave 0 |
| NAT-06 | Interpreter/`-O0`/`-O3`/`-flto` agree on the interprocedural corpus (five-axis comparator) | integration/differential | `go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential -v` | ❌ Wave 0 |
| NAT-07 | 4×3 mutation matrix, one red cell, mutation-killed | integration (hand-written C, no `cgen`) | `go test ./internal/compiler/native/... -run TestCompositionOnlyLTODivergence -v` | ❌ Wave 0 (Q-04 builds this pre-planning per D-11-24) |
| QLT-03 | `qlt03_shape_register.json` audit, zero blank/unclassified cells | unit | `go test ./internal/compiler/session/... -run TestQLT03Register -v` | ❌ Wave 0 |
| QLT-05 | `reduce` re-verification on a multi-function seed, strict field equality | unit + differential | `go test ./internal/compiler/reduce/... -run TestReduceMultiFunctionSeed -v` and `go test ./internal/compiler/session/... -run TestQLT05Reverification -v` | ❌ Wave 0 |
| QLT-06 | `cache.DeclaredInputNames()` stable (or explicitly widened for D-11-41); `cache` does not import `core`/`originvalidate` | unit (structural) | `go test ./internal/compiler/cache/... -run TestDeclaredInputNamesStableAndNoInterproceduralImport -v` | ❌ Wave 0 |

---

## Wave 0 Requirements

- [ ] `internal/compiler/cgen/cgen_program_test.go` — covers NAT-04, NAT-05
- [ ] `internal/compiler/callgraph/callgraph_entry_test.go` (or extend `callgraph_test.go`) — covers `EntryFunction`'s zero-or-many-roots fail-closed refusal (D-11-05)
- [ ] `internal/compiler/reduce/reduce_multifunction_test.go` — covers QLT-05, the two new moves, `RefusedShapes()`
- [ ] `internal/compiler/session/qlt03_shape_register_test.go` — covers QLT-03
- [ ] `internal/compiler/session/session_phase11_gate_test.go` — covers the mid-phase gate's four-part conjunction (D-11-14/D-11-18/D-11-19)
- [ ] `internal/compiler/native/native_lto_test.go` extension — covers NAT-07's 4×3 matrix
- [ ] `internal/compiler/cache/probe_test.go` extension — covers the `DeclaredInputNames()` eighth-entry fix (D-11-41)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Emitted C17 is *readable* (the roadmap goal's qualitative half) | NAT-04 | Readability is not mechanically checkable; the automated half is "compiles under `-std=c17 -Wall -Werror`" | Inspect the emitted `.c` for a multi-function fixture; confirm one function per Lang function, provenance comments present, no macro soup |
| `clang --version` on the execution host matches the version the LTO divergence was characterized against | NAT-07 | Host-dependent; the divergence is a property of a specific LLVM version | Capture `clang --version` output into the NAT-07 evidence record before accepting the red cell |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending

---
phase: "11"
slug: "multi-function-native-emission-and-interprocedural-equivalen"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false   # gaps mapped to plan tasks by /gsd-plan-phase 11 (see § Wave 0 Requirements)
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
| 11-01 T1 | 11-01 | 1 | QLT-05 | T-11-02 | spike verdict is a single-branch committed test, not prose | unit | `go test ./internal/compiler/corevalidate/... -run 'TestQ01' -v -count=1` | ❌ creates | ⬜ pending |
| 11-01 T2 | 11-01 | 1 | QLT-06 | T-11-01 | stale-artifact reuse is demonstrated or withdrawn | unit | `go test ./internal/compiler/session/... -run 'TestQ02' -v -count=1` | ❌ creates | ⬜ pending |
| 11-01 T3 | 11-01 | 1 | QLT-05, QLT-06 | T-11-02 | recorded-not-built decisions carry closing conditions | doc | `grep -c -E 'D-11-(02\|07\|11\|12\|13\|27\|36\|40\|42)' …/PHASE-11-DEBT.md` | ❌ creates | ⬜ pending |
| 11-02 T1 | 11-02 | 1 | NAT-07 | T-11-04 | false `restrict` confined to hand-written test C | integration | `go test ./internal/compiler/native/... -run 'TestCompositionOnlyLTODivergence' -v -count=1` | ❌ creates | ⬜ pending |
| 11-02 T2 | 11-02 | 1 | NAT-07 | T-11-03 | divergence pinned to a recorded toolchain identity | manual + cmd | `clang --version` | ❌ creates | ⬜ pending |
| 11-02 T3 | 11-02 | 1 | NAT-07 | T-11-03 | criterion-3 amendments applied or declined, never silent | checkpoint | `grep -c -E 'flto\|escalation, not a pass' .planning/ROADMAP.md` | ✅ exists | ⬜ pending |
| 11-03 T1 | 11-03 | 2 | NAT-04 | T-11-07, T-11-08 | one entry resolver for oracle and binary; validation not bypassed | integration (tracer) | `go test ./internal/compiler/cgen/... -run 'TestEmitProgram' -v -count=1` | ❌ creates | ⬜ pending |
| 11-03 T2 | 11-03 | 2 | NAT-04 | T-11-07 | zero-or-many roots is a named fail-closed refusal | unit | `go test ./internal/compiler/callgraph/... -run 'TestEntryFunction…' -v -count=1` | ❌ creates | ⬜ pending |
| 11-03 T3 | 11-03 | 2 | NAT-04 | T-11-06 | two-tier name allocation preserves prefix confinement | unit | `go test ./internal/compiler/cgen/... -run 'TestMultiFunctionNameAllocation\|TestEmitProgram' -v -count=1` | ❌ creates | ⬜ pending |
| 11-04 T1 | 11-04 | 3 | NAT-05 | T-11-11 | empty attribute set is generated, not a literal | unit | `go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSet…' -v -count=1` | ❌ creates | ⬜ pending |
| 11-04 T2 | 11-04 | 3 | NAT-04, NAT-05 | T-11-09, T-11-10 | gate non-vacuous, mutation-killed, diff-local suppression | integration | `go test ./internal/compiler/session/... -run 'TestPhase11' -v -count=1` | ❌ creates | ⬜ pending |
| 11-04 T3 | 11-04 | 3 | NAT-05 | T-11-09 | gate verdict recorded with every conjunct's value | checkpoint | `grep -c -E 'function count\|call-edge count\|N =…' …/11-MIDPHASE-GATE.md` | ❌ creates | ⬜ pending |
| 11-05 T1 | 11-05 | 4 | NAT-06 | T-11-14 | every guard has a recorded WIDENED/KEPT disposition | structural | `awk … \| wc -l` + `go test ./internal/compiler/session/... -count=1` | ✅ exists | ⬜ pending |
| 11-05 T2 | 11-05 | 4 | NAT-06 | T-11-12 | four-tier agreement on all five comparator axes | differential | `go test ./internal/compiler/session/... -run 'TestPhase11InterproceduralDifferential' -v -count=1` | ❌ creates | ⬜ pending |
| 11-05 T3 | 11-05 | 4 | NAT-06 | T-11-13 | LTO inertness and lane-deferral declared, not implied | doc | `grep -c 'D-11-25' …/session_phase11_differential_test.go` | ❌ creates | ⬜ pending |
| 11-06 T1 | 11-06 | 4 | QLT-03 | T-11-15 | op-kind closure pinned against a committed literal | unit | `go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure…' -v -count=1` | ❌ creates | ⬜ pending |
| 11-06 T2 | 11-06 | 4 | QLT-03 | T-11-16 | closed proof-mechanism set; no free-text negatives | unit | `go test ./internal/compiler/session/... -run 'TestQLT03Register' -v -count=1` | ❌ creates | ⬜ pending |
| 11-06 T3 | 11-06 | 4 | QLT-03 | T-11-15 | audit provably able to fail without editing its data file | unit | `go test ./internal/compiler/session/... -run 'TestQLT03' -v -count=2` | ❌ creates | ⬜ pending |
| 11-07 T1 | 11-07 | 5 | QLT-06 | T-11-18 | `cgen` source declared; key actually moves | unit | `go test ./internal/compiler/cache/... ./internal/compiler/session/... -run 'TestDeclaredInputNames\|TestQ02' -v -count=1` | ✅ exists | ⬜ pending |
| 11-07 T2 | 11-07 | 5 | QLT-06 | T-11-19, T-11-20, T-11-21 | dual import scans with negative controls; nothing cached | structural | `go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames\|TestCache…\|TestNoClosureDigestInCache' -v -count=1` | ✅ exists | ⬜ pending |
| 11-07 T3 | 11-07 | 5 | QLT-06 | T-11-20 | split row recorded; abstention argued | doc | `grep -c -E 'QLT-06a\|QLT-06b\|strictly dominates…' …/11-QLT06-ABSTENTION.md` | ❌ creates | ⬜ pending |
| 11-08 T1 | 11-08 | 5 | QLT-05 | T-11-23, T-11-24 | derived bound has a floor; `reduce` stays out of `callgraph` | unit | `go test ./internal/compiler/reduce/... -run 'TestReduce\|TestDerivedAttemptBound' -v -count=1` | ❌ creates | ⬜ pending |
| 11-08 T2 | 11-08 | 5 | QLT-05 | T-11-22 | no move renumbers; no move inlines | unit | `go test ./internal/compiler/reduce/... -run 'TestDropCallSite\|TestDropOrphanFunction\|…' -v -count=1` | ❌ creates | ⬜ pending |
| 11-08 T3 | 11-08 | 5 | QLT-05 | T-11-22 | every declined shape is named and exercised | unit | `go test ./internal/compiler/reduce/... -run 'TestRefusedShape\|TestNoFlakyPredicateTolerance' -v -count=1` | ❌ creates | ⬜ pending |
| 11-09 T1 | 11-09 | 6 | QLT-05 | T-11-25 | foreign-call sequence derived twice, independently | unit | `go test ./internal/compiler/session/... -run 'TestForeignCallSequence' -v -count=1` | ✅ exists | ⬜ pending |
| 11-09 T2 | 11-09 | 6 | QLT-05 | T-11-26 | strict field equality; `CausalRole` drift rejected | differential | `go test ./internal/compiler/session/... -run 'TestQLT05\|TestMismatchDocumentSchemaUnchanged' -v -count=1` | ❌ creates | ⬜ pending |
| 11-09 T3 | 11-09 | 6 | QLT-05 | T-11-27 | gate non-vacuous on both sides | unit + doc | `go test ./internal/compiler/session/... -run 'TestQLT05Gate\|TestQLT05EmptyReduction' -v -count=1` | ❌ creates | ⬜ pending |

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

Every gap below is owned by a named plan task. No task in any plan carries a
`MISSING — Wave 0 …` sentinel, because each test file is created by the same task
whose behaviour it proves.

- [ ] `internal/compiler/cgen/cgen_program_test.go` — covers NAT-04, NAT-05 → **11-03 T1/T3, 11-04 T1**
- [ ] `internal/compiler/callgraph/callgraph_entry_test.go` — covers `EntryFunction`'s zero-or-many-roots fail-closed refusal (D-11-05) → **11-03 T2**
- [ ] `internal/compiler/reduce/reduce_multifunction_test.go` — covers QLT-05, the two new moves, `RefusedShapes()` → **11-08 T1/T2/T3**
- [ ] `internal/compiler/session/qlt03_shape_register_test.go` — covers QLT-03 → **11-06 T1/T2/T3**
- [ ] `internal/compiler/session/session_phase11_gate_test.go` — covers the mid-phase gate's four-part conjunction (D-11-14/D-11-18/D-11-19) → **11-04 T2**
- [ ] `internal/compiler/native/native_lto_test.go` extension — covers NAT-07's 4×3 matrix → **11-02 T1** (runs in wave 1, before any `cgen` work, per D-11-24)
- [ ] `internal/compiler/cache/probe_test.go` extension — covers the `DeclaredInputNames()` eighth-entry fix (D-11-41) → **11-07 T1/T2**

Additional Wave 0 items added by planning (the three mandated pre-planning
experiments, each with a stated branch rather than an assumed outcome):

- [ ] `internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go` — Q-01 → **11-01 T1**; branch A proceeds with `drop-call-site` as a live reducer move, branch B ships it as a named `RefusedShapes()` entry behind the `peerDeriveOriginFacts` `OpCall` gap
- [ ] `internal/compiler/session/session_phase6_cache_hole_test.go` — Q-02 → **11-01 T2**; branch A ships the eighth declared cache input, branch B records the claim as withdrawn
- [ ] `internal/compiler/native/native_lto_test.go` — Q-04 → **11-02 T1**; an all-green matrix is a lane failure and an escalation, never a pass
- [ ] `internal/compiler/session/qlt03_shape_register_test.go` — Q-03 → **11-06 T1**; a contradicted prediction revises the taxonomy before any register row is committed

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

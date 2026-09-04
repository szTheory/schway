---
gsd_state_version: 1.0
current_phase: 3
current_phase_name: Borrowed Views and CFG Lifetimes
status: planning
stopped_at: Phase 02 complete, ready to plan Phase 3
last_updated: "2026-09-04T01:50:25.190Z"
last_activity: 2026-09-04
last_activity_desc: Phase 02 complete, transitioned to Phase 3
state_head: 6e3ef81d84b7b1311b452eca5dd17c4acf302710
progress:
  total_phases: 6
  completed_phases: 2
  total_plans: 10
  completed_plans: 10
  percent: 33
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-03)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** Phase 02 — Owned Values and Abilities

## Current Position

Phase: 3 — Borrowed Views and CFG Lifetimes
Plan: Not started
Status: Ready to plan
Last activity: 2026-09-03 — Phase 02 complete, transitioned to Phase 3

Progress: ██████████ [███░░░░░░░] 33%

## Performance Metrics

**Velocity:**

- Total plans completed: 7
- Average duration: 11 min
- Total execution time: 105 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 02 | 7 | - | - |
**Per-Plan Metrics:**

| Plan | Duration | Tasks | Files |
|------|----------|-------|-------|
| Phase 01 P01 | 10 min | 3 tasks | 18 files |
| Phase 01 P02 | 4 min | 3 tasks | 9 files |
| Phase 01 P03 | 16 min | 3 tasks | 14 files |
| Phase 02 P01 | 10 min | 2 tasks | 14 files |
| Phase 02 P02 | 9 min | 3 tasks | 6 files |
| Phase 02 P03 | 12 min | 2 tasks | 8 files |
| Phase 02 P04 | 13 min | 2 tasks | 4 files |
| Phase 02 P05 | 12min | 2 tasks | 8 files |
| Phase 02 P06 | 10 min | 2 tasks | 9 files |
| Phase 02 P07 | 9 min | 2 tasks | 8 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md and the provenance-rich wiki research ledger.

- M001 uses vertical source-to-native slices.
- Stage 0 uses Go 1.24 stdlib and readable C17/Clang reversibly.
- The deterministic interpreter is the semantic oracle.
- Syntax remains provisional; stable typed-core/evidence identities are the asset.
- Phase 1 command/evidence contracts are executable; Phase 2 should extend the same vertical path with affine ownership and independent abilities.
- [Phase 02]: Linear type, place, operation, and point IDs use function-local semantic ordinals, never source offsets.
- [Phase 02]: Phase 1 match artifacts retain lang.core/0 and lang.execution/0 while owned linear artifacts use /1.
- [Phase 02]: Ownership token normalization lives at the parser boundary so the Phase 1 lexer remains unchanged.
- [Phase 02]: Box and Pair use a request-local package-private structural conjunction with no arbitrary production masks.
- [Phase 02]: Generic parser limits are depth 64 and 4,096 type nodes with declaration-bounded recovery.
- [Phase 02]: Legacy Error construction remains diagnostic/0; only repair-bearing ownership diagnostics select diagnostic/1.
- [Phase 02]: Straight-line shared loans expire immediately after their precomputed final use through a linear indexed schedule.
- [Phase 02]: Ownership checker work counts type nodes, the complete last-use scan, and the inspected forward prefix.
- [Phase 02]: Validator ability and transition authorization is independently implemented from checker and interpreter behavior, sharing only inert core records.
- [Phase 02]: The owned final return operation is the explicit final claim, and canonical validation work is exactly 16n+13.
- [Phase 02]: Each compile/run stdout/stderr stream has an independent 64 KiB max-plus-one bound and stable stage/stream truncation code.
- [Phase 02]: Native execution decodes stdout only as exactly one strict execution document; successful-run stderr is operational failure.
- [Phase 02]: Interpreter/O0/O3 equality covers complete ordered semantic execution facts while excluding physical observations.
- [Phase 02]: Legacy match Emit bytes remain frozen for Phase 1 evidence while native execution uses dedicated strict-JSON emission.
- [Phase 02]: Owned evidence selects lang.evidence/1 after independent core admission while Phase 1 evidence remains byte-identical on /0.
- [Phase 02]: SHA-256 is content identity only; the coordinated source/core lie is an expected escape, never a detected control.
- [Phase 02]: The bounded phase gate runs shared test/race/vet work once and reports five 20-sample warm distributions without ratifying an SLO.
- [Phase 02]: Generated linear C records executed operation events and serializes the returned runtime place through one counted 64 KiB output layer.
- [Phase 02]: The exact-one owned backend mutation must run at O0 and O3 and produce semantic mismatch exit 4 before control:backend.runtime_causality is admitted.
- [Phase 02]: Generated C uses one global ordinary-identifier allocator that preserves legacy source-derived names when unique and adds deterministic category/ordinal suffixes only for actual collisions.
- [Phase 02]: Compiler-spawned tool identity probes have independent 64 KiB-plus-one stdout/stderr bounds and five-second deadlines.
- [Phase 02]: `Buffer` grants `share` (OV-02-01), resolving a self-contradiction between the ability table and the shipped move-while-borrowed control fixture; `Buffer` remains noncopyable.
- [Phase 02]: `borrow` is gated on `AbilityShare` in the checker as defence in depth; the gate is unreachable from source today and guarded by a self-invalidating enumeration test.
- [Phase 02]: Loan liveness is transitive across reborrows and copies-of-loans in both admission layers, with the test oracle re-derived by fixed-point closure so it cannot mirror the production law.
- [Phase 02]: The formatter classifies an opening brace by the declaration keyword that opened the line, and that classification survives a trailing comment.

### Pending Todos

None yet.

### Blockers/Concerns

- Baseline machines for ratified feedback budgets remain to be chosen before Phase 6.
- Nine Phase 02 debt items are carried into Phase 3; see `.planning/phases/02-owned-values-and-abilities/02-DEBT.md`. Two are deadline-bearing: D-02-05 (`__LANG_` → `_LANG_` before any further C artifact is frozen) and D-02-03 (the Θ(N²) checker cost, which OWN-03's CFG liveness should remove anyway).
- Process debt adopted as standing rules after three gate failures shared one shape — a green test whose reachable input space omitted the hard case: mutation-kill every differential, interrogate what inputs a property test actually reaches, and drive the shipped binary on hand-written programs rather than only the gate's own corpus.

### Roadmap Evolution

- Phase 1 edited: removed generic web-app MVP mode; retained tracer-first vertical planning

## Deferred Items

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| Runtime | Effects, async, actors, scheduling, managed heaps | Deferred | Initialization | Post-M001 |
| Ecosystem | Packages and first-party application kits | Deferred | Initialization | Post-M001 |

## Session Continuity

Last session: 2026-09-04
Stopped at: Phase 02 complete, ready to plan Phase 3
Resume file: none — start Phase 03 with `/gsd-plan-phase 03`

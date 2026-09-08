---
gsd_state_version: 1.0
milestone: M002
milestone_name: Interprocedural Semantic Spine
current_phase: 07
current_phase_name: Calls, Signatures, and Call-Graph Refusal
status: roadmapped
stopped_at: Phase 07 context gathered
last_updated: "2026-09-08T18:47:42.233Z"
last_activity: 2026-09-08
last_activity_desc: M002 roadmap created (7 phases, 30/30 requirements mapped)
state_head: f68327c1f58b23caac81d5b41adfe09dd7abfa9d
progress:
  total_phases: 7
  completed_phases: 0
  total_plans: 5
  completed_plans: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-07)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** M002 — Interprocedural Semantic Spine. Land `OpCall` at all
six dispatch sites and prove every M001 guarantee survives a function boundary.

**Durable context (survives context clears — read before re-deriving):**

- `.planning/LANGUAGE-MATURITY.md` — the language is far less expressive than
  the roadmap vocabulary implies: no arithmetic, no iteration, no Lang-to-Lang
  calls yet. Assurance stack ~60-70% built; language surface ~5-10%.
- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (deps,
  anti-features, the six dispatch sites, why `-flto` is load-bearing).

## Current Position

Phase: 07 (Calls, Signatures, and Call-Graph Refusal) — READY TO EXECUTE
Plan: —
Status: Roadmap created; awaiting phase discussion/planning
Last activity: 2026-09-08 — M002 roadmap created (7 phases, 30/30 requirements mapped)

Progress: [--------------------] 0% (0/7 phases)

## M002 Phase Map

| Phase | Name | Requirements | Status |
|-------|------|--------------|--------|
| 07 | Calls, Signatures, and Call-Graph Refusal | 5 | Not started |
| 08 | Interprocedural Loan Liveness in `check` | 2 | Not started |
| 09 | Peer Re-Derivation and D-03-02 Closure | 6 | Not started |
| 10 | Trusted Interprocedural Oracle | 5 | Not started |
| 11 | Multi-Function Native Emission and Equivalence | 7 | Not started |
| 12 | `Result` Payloads | 2 | Not started |
| 13 | Agent Loop for Interprocedural Defects | 3 | Not started |

Phase numbering continues from M001 (which ended at Phase 06). Full detail:
`.planning/ROADMAP.md`.

## Performance Metrics

**Velocity:**

- Total plans completed: 59
- Average duration: 11 min
- Total execution time: 105 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 02 | 7 | - | - |
| 03 | 10 | - | - |
| 04 | 13 | - | - |
| 05 | 14 | - | - |
| 06 | 15 | - | - |
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
| Phase 03 P01 | 95 min | 3 tasks | 19 files |
| Phase 03 P02 | 70 min | 3 tasks | 26 files |
| Phase 03 P06 | 70 min | 3 tasks | 17 files |
| Phase 03 P03 | 100 min | 3 tasks | 6 files |
| Phase 03 P04 | 95 min | 3 tasks | 7 files |
| Phase 03 P05 | 95 min | 3 tasks | 7 files |
| Phase 03-borrowed-views-and-cfg-lifetimes P07 | 130min | 3 tasks | 13 files |
| Phase 03 P08 | 11 min | 2 tasks | 4 files |
| Phase 03 P09 | 22 min | 3 tasks | 8 files |
| Phase 03 P10 | 45 min | 3 tasks | 9 files |
| Phase 04 P01 | ~5h | 4 tasks | 25 files |
| Phase 04 P02 | 3h | 3 tasks | 16 files |
| Phase 04 P03 | ~2h | 3 tasks | 15 files |
| Phase 04 P04 | ~2h | 3 tasks | 19 files |
| Phase 04-fallible-resources-and-c-boundary P05 | ~2h | 3 tasks | 11 files |
| Phase 04 P06 | ~2h | 3 tasks | 9 files |
| Phase 04 P07 | 55 min | 3 tasks | 7 files |
| Phase 04 P08 | 30 min | 3 tasks | 3 files |
| Phase 04 P09 | 25 min | 2 tasks | 2 files |
| Phase 04 P10 | 20min | 2 tasks | 3 files |
| Phase 04 P11 | 35 min | 2 tasks | 3 files |
| Phase 04 P12 | 45min | 3 tasks | 5 files |
| Phase 04 P13 | 30 min | 4 tasks | 10 files |
| Phase 05 P01 | 110 min | 3 tasks | 8 files |
| Phase 05 P02 | ~70min | 3 tasks | 3 files |
| Phase 05 P03 | 35 min | 3 tasks | 2 files |
| Phase 05 P04 | 90 min | 3 tasks | 8 files |
| Phase 05 P05 | ~140min | 3 tasks | 8 files |
| Phase 05 P06 | 75 min | 3 tasks | 5 files |
| Phase 05 P07 | 180 min | 3 tasks | 4 files |
| Phase 05 P08 | 95min | 3 tasks | 13 files |
| Phase 05 P09 | 36min | 3 tasks | 6 files |
| Phase 05 P10 | ~110min | 3 tasks | 4 files |
| Phase 05 P11 | 45min | 3 tasks | 3 files |
| Phase 05 P12 | 95min | 3 tasks | 8 files |
| Phase 05 P13 | ~50min | 2 tasks | 5 files |
| Phase 05 P14 | ~35 min (continuation) | 3 tasks | 5 files |
| Phase 06-agent-feedback-and-performance-ratification P01 | 40min | 3 tasks | 3 files |
| Phase 06 P02 | 55 min | 3 tasks | 6 files |
| Phase 06-agent-feedback-and-performance-ratification P04 | 55 min | 3 tasks | 4 files |
| Phase 06 P08 | 35min | 3 tasks | 4 files |
| Phase 06-agent-feedback-and-performance-ratification P11 | 18 min | 3 tasks | 3 files |
| Phase 06-agent-feedback-and-performance-ratification P03 | 70min | 3 tasks | 4 files |
| Phase 06 P05 | 55min | 3 tasks | 4 files |
| Phase 06 P12 | ~35 min | 3 tasks | 13 files |
| Phase 06 P06 | 65min | 3 tasks | 8 files |
| Phase 06 P13 | 55min | 3 tasks | 6 files |
| Phase 06 P07 | 95min | 3 tasks | 8 files |
| Phase 06-agent-feedback-and-performance-ratification P14 | 50 min | 3 tasks | 11 files |
| Phase 06 P09 | 55min | 3 tasks | 4 files |
| Phase 06 P10 | 55min | 3 tasks | 9 files |
| Phase 06 P15 | ~40 min | 3 tasks | 12 files |

## Accumulated Context

### Decisions

Full decision log lives in PROJECT.md (Key Decisions) and the provenance-rich
`wiki/` research ledger. Per-phase decisions from M001 are archived with their
phase artifacts under `.planning/milestones/M001-phases/`.

Standing architectural commitments carried into M002:

- Vertical source-to-native slices, not layered compiler construction.
- Go 1.24 stdlib for Stage 0; readable C17 through Clang as the reversible
  native path.
- The deterministic interpreter is the semantic oracle.
- Syntax stays provisional; stable typed-core, diagnostic, and evidence
  identities are the asset.
- Trust-crossing facts are re-derived independently (`corevalidate`,
  `originvalidate`), never trusted from the producer.

### Pending Todos

Three bounded pre-phase spikes, recorded in ROADMAP.md ("Pre-Phase Spikes"):

- **S-006 interprocedural liveness cost-scaling probe** — hard entry gate on
  *planning* Phase 08. Deliberately a pre-phase spike rather than a phase or a
  Phase 07 gate: it carries no requirement, produces throwaway workbench code,
  and its finding is Phase-08-blocking, not Phase-07-blocking.
- **S-007 recursive / mutually-recursive stress corpus** — runs alongside
  Phase 07, non-blocking; informs whether the call-stack ceiling is a fixed
  constant or a declared budget.
- **S-008 Nyquist fold-in cost measurement** — runs before Phase 09 planning;
  determines whether QLT-07 stays committed or becomes a declared stretch item.

Also: a `Result` payload / interprocedural-origin interaction probe must be
answered before Phase 12 is planned.

### Blockers/Concerns

- **D-03-02 (open past M001):** an exported borrow-derived return with no
  declared origin exports indistinguishable from a fully-owned return, in the
  INTERPROCEDURAL half of the hazard. The single-function half closed in
  Phase 3. Owned by M002's `OpCall` charter (D-05-32/D-05-33).
- **M001 ships without Lang-to-Lang calls.** Interprocedural `-O3` equivalence
  is outside M001's proof scope by construction, not by omission.
- Remaining Phase 2/4 debt is registered in the archived `*-DEBT.md` files and
  summarized in `.planning/milestones/M001-MILESTONE-AUDIT.md`: five open
  Phase 2 items (info/warning level), D-04-30 (storable/matchable `Result`
  values, deferred to M002), and D-04-31's four accepted residual limitations.
- Nyquist validation is compliant for Phases 1, 2, and 4; Phases 3, 5, and 6
  are not-validated. Overall status: partial.
- **Standing process rules** adopted after three M001 gate failures that shared
  one shape — a green test whose reachable input space omitted the hard case:
  mutation-kill every differential, interrogate what inputs a property test
  actually reaches, and drive the shipped binary on hand-written programs
  rather than only the gate's own corpus.

### Roadmap Evolution

- Phase 1 edited: removed generic web-app MVP mode; retained tracer-first
  vertical planning.
- `OpCall` and interprocedural equivalence deferred out of M001 (Phase 5,
  D-05-32) and promoted to M002's lead charter.
- **M002 roadmap (2026-09-08):** adopted `research/SUMMARY.md`'s reconciled
  build order with **one documented departure** — SUMMARY's Phase 4 is split
  into Phases 10 and 11 at the ARCHITECTURE.md Stage 6 / Stage 7 boundary.
  Reason: as one phase it would carry 12 of 30 requirements and two brand-new
  subsystems (interpreter `Frame` stack; multi-function C emission) at M001's
  recorded 70-95 min/plan coordinated-change rate, and it would hold the
  milestone's central risk unadjudicated for its whole length. The "cgen must
  differential-test against a trusted interpreter" dependency SUMMARY itself
  names is a gate, so it is used as one.
- **Scope-cut order carried into the roadmap up front** (Phase 12 → QLT-07 →
  Phase 13 narrowing), with the explicit trigger: if Phase 08 or 09 exceeds
  ~2x its initial plan estimate, renegotiate Phase 12 out to M003 immediately
  rather than adding plans.
- **Two mandatory mid-phase gates** recorded, following M001 Phase 3's
  precedent: Phase 08 (interprocedural loan liveness) and Phase 11
  (multi-function C emission before alias-attribute call lowering).

## Deferred Items

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| Runtime | Effects, async, actors, scheduling, managed heaps | Deferred | Initialization | Post-M001 |
| Ecosystem | Packages and first-party application kits | Deferred | Initialization | Post-M001 |

## Session Continuity

Last session: 2026-09-08T17:47:28.537Z
Stopped at: Phase 07 context gathered
durable context recorded in LANGUAGE-MATURITY.md and STANDING-VERDICTS.md
Resume file: .planning/phases/07-calls-signatures-and-call-graph-refusal/07-CONTEXT.md
Next command: `/gsd-discuss-phase 07`

## Operator Next Steps

- Review `.planning/ROADMAP.md` (note the documented departure from the
  research-reconciled 6-phase order, and the pre-phase spike decision).
- Kick off S-006 and S-007 alongside Phase 07.
- Discuss and plan Phase 07 with /gsd-discuss-phase.

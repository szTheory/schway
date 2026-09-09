---
gsd_state_version: 1.0
milestone: M002
milestone_name: Interprocedural Semantic Spine
current_phase: 07
current_phase_name: Calls, Signatures, and Call-Graph Refusal
status: executing
stopped_at: Completed 07-10-PLAN.md
last_updated: "2026-09-09T17:35:35.352Z"
last_activity: 2026-09-09
last_activity_desc: Phase 07 execution started
state_head: 045973aff94ab2b3e5f6f75369702aa5bd0b5fc2
progress:
  total_phases: 7
  completed_phases: 0
  total_plans: 12
  completed_plans: 10
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-07)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** Phase 07 — Calls, Signatures, and Call-Graph Refusal
six dispatch sites and prove every M001 guarantee survives a function boundary.

**Durable context (survives context clears — read before re-deriving):**

- `.planning/LANGUAGE-MATURITY.md` — the language is far less expressive than
  the roadmap vocabulary implies: no arithmetic, no iteration, no Lang-to-Lang
  calls yet. Assurance stack ~60-70% built; language surface ~5-10%.
- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (deps,
  anti-features, the six dispatch sites, why `-flto` is load-bearing).

## Current Position

Phase: 07 (Calls, Signatures, and Call-Graph Refusal) — EXECUTING
Plan: 2 of 12
Status: Ready to execute
Last activity: 2026-09-09 — Phase 07 execution started

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
| Phase 07 P01 | 55min | 3 tasks | 8 files |
| Phase 07 P02 | 90 min | 3 tasks | 11 files |
| Phase 07 P03 | 55 min | 3 tasks | 16 files |
| Phase 07 P04 | ~70 min | 3 tasks | 14 files |
| Phase 07 P05 | ~140min | 3 tasks | 10 files |
| Phase 07 P06 | 95min | 3 tasks | 12 files |
| Phase 07 P07 | 150 min | 3 tasks | 13 files |
| Phase 07 P08 | 64min | 2 tasks | 11 files |
| Phase 07 P09 | 90 min | 4 tasks | 11 files |
| Phase 07 P10 | 105min | 3 tasks | 9 files |

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
- [Phase 07]: 07-01: DecodeInterface routes CheckSummary; /0 stays decodable but never admissible for a call — D-07-36 closes codex's HIGH finding that strict /1 decoding was unspecified and /0 dispatch was unwired
- [Phase 07]: Callable is publication safety (D-04-03), not export membership: PublishProblemsFor extracted, corevalidate summary peer wired at both replay sites, narrowed to core.origin_omitted (D-07-33) — 07-02 review-driven decisions D-07-31/D-07-32/D-07-33/D-07-20/D-07-21/D-07-22 landed exactly as ratified at the checkpoint
- [Phase 07]: core.LinearOperation.CalleeID added additive+omitempty per D-07-29; the parser change and the foreign-bare-call refusal relocation to check (D-07-40) landed together in Task 1 since they are inseparable within one green commit.
- [Phase 07]: 07-04: phase07 lane fires all three Phase 07 controls (phase07_in_process, phase07_lane, dispatch.recognized_not_executed) since the lane's own fixtures genuinely re-prove all three claims through the CLI path; cmd/lang main.go isPhase7Corpus dispatch added as Rule 3 blocking fix (not in plan's files_modified) since the lane is not CLI-observable without it
- [Phase 07]: 07-05: check.Program builds an immutable post-body signature table (reusing core.FunctionSignature, structurally body-free) and refuses a call to a non-Callable callee with core.callee_not_callable; corevalidate independently re-derives the same refusal via peerCallable. Discovered relay_escort_witness.lang checks clean while corevalidate refuses (core.move_while_borrowed) -- the interprocedural half of D-03-02, left open for Phase 08/09. — Two genuinely independent SEM-06 derivations (check's table consult, corevalidate's own peerCallable) satisfy QLT-08 without collapsing to one derivation read twice; the D-03-02 divergence is documented as a named, tested finding rather than hidden.
- [Phase 07]: 07-06: callgraph.Order (iterative white/gray/black DFS, no native recursion, roots = all declared functions, edges from CalleeID only) wired into check before it returns a core.Program; canonical rotation plus cross-cycle lexicographically-smallest witness selection makes the core.call_graph_cycle diagnostic ID deterministic regardless of discovery order (D-07-16/D-07-43); diagnostic bounded at 32 cycle_member causes while the traversal itself stays unbounded; spans projected from OpCall operation IDs via new check-side emission bookkeeping, no Span added to core.LinearOperation (D-07-35); three fault-injection seams (gray-vs-visited, self-edge, unresolved-edge-drop) each independently mutation-killed against a diamond/shared-leaf corpus a chain fixture could never kill.
- [Phase 07]: 07-07: corevalidate grows its own independently-written whole-program cycle peer (checkCallGraphAcyclic) over disjoint synthetic-only input space; run() split into structural-then-replay passes to place it correctly; remaining SEM-07 corpus (indirect/unreachable/match-arm/shadowing) landed; bilateral independent-disable proof and phase-wide completeness matrix close QLT-08 for Phase 07.
- [Phase 07]: [Phase 07] 07-08: ClosureDigest chained over callee summary digests only after callgraph.Order proves acyclicity; corevalidate independently re-derives the chain over its own postorder, matching the producer byte-for-byte wherever D-07-33's narrowed Callable scope already agrees; callee-changes-invalidates-caller mutation-killed on both sides; SEM-05 closed.
- [Phase 07]: Checkpoint auto-ratified: accepted all four proposed diagnostic code strings (check.call_argument_type_mismatch, check.call_return_type_unrepresentable, core.CallArgumentTypeMismatch, core.CallReturnTypeMismatch) and both ordered Causes shapes verbatim.
- [Phase 07]: 07-09: check.resolveCallBinding gates argument-type match and derives OpCall's TargetID.TypeID from the callee's declared return contract (fail-closed); corevalidate independently re-derives both refusals with no shared helper. Closes 07-VERIFICATION.md's single FAILED truth and 07-REVIEW.md CR-01.
- [Phase 07]: 07-10: CheckCommandFile consults corevalidate.Validate (refusing union, fixed precedence check-then-peer-then-originvalidate); interface export/core report a peer refusal as StatusInvalid with the peer's own code instead of tool.operation_failed. Closes 07-REVIEW.md CR-04/PVG-03.

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

Last session: 2026-09-09T17:35:35.334Z
Stopped at: Completed 07-10-PLAN.md
durable context recorded in LANGUAGE-MATURITY.md and STANDING-VERDICTS.md
Resume file: None
Next command: `/gsd-discuss-phase 07`

## Operator Next Steps

- Review `.planning/ROADMAP.md` (note the documented departure from the
  research-reconciled 6-phase order, and the pre-phase spike decision).
- Kick off S-006 and S-007 alongside Phase 07.
- Discuss and plan Phase 07 with /gsd-discuss-phase.

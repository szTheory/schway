---
gsd_state_version: "1.0"
milestone: M004
milestone_name: Native Emission Ownership and Resource Discharge
status: planning
last_updated: "2026-09-27T13:42:25.661Z"
last_activity: 2026-09-27
progress:
  total_phases: 0
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-27)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** Define M004 requirements and roadmap; new implementation starts at Phase 22

**Durable context (survives context clears — read before re-deriving):**

- `.planning/LANGUAGE-MATURITY.md` — current source-grounded capability snapshot:
  calls, computed Result payloads and U64 constants execute; arithmetic/loops and
  ordinary application IO do not. Native run is still a canned-input differential
  harness. Foreign/by-pointer program families remain refused. No completeness
  percentages are meaningful. The wiki tour remains a design target.
- `.planning/PRODUCT-ROADMAP.md` — living near/mid/long capability order and
  current three recommendations. AGENTS.md requires proactive review at planning
  transitions. M004 is real application/resource ownership; FizzBuzz follows.

- `.planning/STANDING-VERDICTS.md` — already-researched verdicts (deps,
  anti-features, the six dispatch sites, why `-flto` is load-bearing).

- **GSD no-loop rule:** Before repeating a routed command, compare its inputs
  and expected state transition with the previous attempt. Rerun only when the
  evidence, inputs, or requested action changed. If the same blocker remains,
  stop and reconcile the conflicting source or identify the exact owner of the
  required update; do not present the same command as progress. When STATE,
  ROADMAP, and live GSD routing disagree, inspect the structured resolver,
  record the discrepancy in the handoff, and use the next action that advances
  the earliest real blocker. Report what changed and what evidence proves it.
  Before recommending a verification command, inspect the existing UAT and
  VERIFICATION artifacts and the canonical status. If UAT is already complete
  and only the report fingerprint is stale, refresh the verifier report while
  preserving UAT; do not rerun UAT or echo a stale router command. After closing
  a blocker, re-query `init.progress` and update STATE/handoff from its result;
  never carry forward an old next-action pointer.

- **Shift-left verification default:** Turn objective acceptance criteria into
  deterministic checks at the lowest layer that proves the claim: unit, seam,
  smoke, integration, or end-to-end. Put a check in recurring CI when its
  regression value justifies its runtime and maintenance cost. Keep expensive
  measurements bounded to the runs that need them, with a cheaper recurring
  structural or receipt-binding guard when that provides continuing value.
  Target zero human UAT when automated evidence covers the criteria; hand off
  only irreducibly subjective, external, or user-authority decisions.

- **Stale-report route:** A complete UAT and a stale verification fingerprint
  are different states. `$gsd-verify-work` handles UAT and can route back to
  itself when the report is stale; it does not refresh that report. When
  `init.execute-phase N` reports zero incomplete plans, use
  `$gsd-execute-phase N` to resume at the phase gates and verifier without
  replaying plans. Preserve completed UAT, then re-query `init.progress` and
  update this file with the exact next route.

- **Go test cache in this workspace:** The default Go cache path is outside
  the writable sandbox. Prefix Go test commands with
  `GOCACHE=/tmp/ai-lang-verification-gocache`.

## Current Position

Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-09-27 — Milestone M004 started

## M003 Closeout (archived)

M003 shipped on 2026-09-26: Phases 14–20, 7 phases, 85 plans, 114 tasks, and
33/33 requirements with all seven phase verifications passing. The audit status
is `tech_debt`, with partial Nyquist coverage in Phases 14, 17, and 18 and four
open unowned debt items within the five-item cap. Phase 21 is the provisional
M004 follow-on and is now filed under `.planning/milestones/M004-phases/`,
outside M003's outgoing cleanup. The user chose to skip quick-task archival;
completed quick tasks remain under `.planning/quick/`.

- Full phase history: `.planning/milestones/M003-phases/`
- Roadmap: `.planning/milestones/M003-ROADMAP.md`
- Requirements: `.planning/milestones/M003-REQUIREMENTS.md`
- Audit: `.planning/milestones/M003-MILESTONE-AUDIT.md`

## M004 Handoff

M004 opened 2026-09-27 under the user's explicit authorization to apply the
second specialist review automatically. The SDK switched STATE/state.json from
M003 to M004; outgoing phase cleanup found zero physical phase directories.
New phases start at 22. Phase 21 remains completed, archived prework with six
plans and seven preserved UAT cases; do not replay it.

The accepted scope is a single-run native application boundary, a real live
foreign allocation owned through Lang, acquisition-based obligations and
per-operation contracts, transfer/error cleanup, and bounded shared/exclusive
read-copy pointers with no additional alias attributes. Arithmetic/loops and
FizzBuzz move to the following milestone. See PROJECT and PRODUCT-ROADMAP.

Phase 21's archive is
`.planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/`.
Its historical verification passed 6/6 at its recorded revision. This kickoff
amends the D-12-43 debt disposition and historical Phase 16 ownership guard;
those inputs occur in archived Phase 16/21 fingerprints. Their receipts remain
historical, not freshly verified at this revision. Focused planning/debt checks
are separate evidence; completed UAT is preserved.

D-12-43's Phase 18 wrong-slot witness is now reflected in the authored debt and
canonical generated view. Current qualified unowned debt is three:
D-10-C04, D-14-46, D-14-47. M003's four-item closeout count remains historical.

Next command after roadmap creation: `$gsd-plan-phase 22`.
The installed resolver may not identify latest completed `M00x` milestones;
M003 is the shipped predecessor and Phase 21 is archived M004 prework. Do not
reset numbering or infer that archived Phase 21 must run again.

## M002 Phase Map

| Phase | Name | Requirements | Status |
|-------|------|--------------|--------|
| 07 | Calls, Signatures, and Call-Graph Refusal | 5 | Complete |
| 08 | Interprocedural Loan Liveness in `check` | 2 | Complete |
| 09 | Peer Re-Derivation and D-03-02 Closure | 6 | Complete |
| 10 | Trusted Interprocedural Oracle | 6 | Complete |
| 11 | Multi-Function Native Emission and Equivalence | 7 | Complete |
| 12 | `Result` Payloads | 2 | Complete |
| 13 | Agent Loop for Interprocedural Defects | 3 | Complete (DX-06/DX-07 partial) |

Phase numbering continues from M001 (which ended at Phase 06). Numbering continues
into M003 — never restart at 01. Full detail: `.planning/milestones/M002-ROADMAP.md`;
execution artifacts in `.planning/milestones/M002-phases/`.

## Performance Metrics

**Velocity:**

- Total plans completed: 194
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
| 07 | 12 | - | - |
| 08 | 6 | - | - |
| 09 | 10 | - | - |
| 10 | 9 | - | - |
| 11 | 9 | - | - |
| 12 | 8 | - | - |
| 14 | 13 | - | - |
| 15 | 10 | - | - |
| 16 | 26 | - | - |
| 17 | 9 | - | - |
| 18 | 10 | - | - |
| 19 | 7 | - | - |
| 21 | 6 | - | - |
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
| Phase 07 P11 | 90 min | 3 tasks | 12 files |
| Phase 07 P12 | 55min | 3 tasks | 12 files |
| Phase 08 P01 | ~33min | 3 tasks | 5 files |
| Phase 08 P02 | ~50 min | 3 tasks | 2 files |
| Phase 08 P03 | 30 min | 3 tasks | 12 files |
| Phase 08 P04 | ~55min | 3 tasks | 3 files |
| Phase 08 P05 | 65min | 3 tasks | 8 files |
| Phase 08 P06 | ~45min | 3 tasks | 4 files |
| Phase 09 P01 | 33min | 3 tasks | 9 files |
| Phase 09 P02 | 45min | 3 tasks | 7 files |
| Phase 09 P03 | 55min | 3 tasks | 5 files |
| Phase 09 P04 | 55min | 3 tasks | 2 files |
| Phase 09 P05 | 55min | 3 tasks | 3 files |
| Phase 09 P07 | 55min | 3 tasks | 2 files |
| Phase 09 P06 | 50min | 3 tasks | 7 files |
| Phase 09 P09 | ~5h | 4 tasks | 10 files |
| Phase 09 P10 | 50min | 3 tasks | 5 files |
| Phase 10-trusted-interprocedural-oracle P01 | 55 min | 3 tasks | 4 files |
| Phase 10 P02 | 95min | 3 tasks | 11 files |
| Phase 10-trusted-interprocedural-oracle P03 | 100min | 3 tasks | 10 files |
| Phase 10-trusted-interprocedural-oracle P04 | 45min | 3 tasks | 3 files |
| Phase 10 P05 | 85 min | 2 tasks | 5 files |
| Phase 10-trusted-interprocedural-oracle P06 | 40min | 3 tasks | 4 files |
| Phase 10-trusted-interprocedural-oracle P07 | 65 min | 3 tasks | 6 files |
| Phase 10 P08 | 140min | 3 tasks | 5 files |
| Phase 10-trusted-interprocedural-oracle P09 | 70min | 3 tasks | 9 files |
| Phase 11 P01 | 25 min | 3 tasks | 3 files |
| Phase 11 P02 | 55min | 3 tasks | 3 files |
| Phase 11 P03 | 70 min | 3 tasks | 13 files |
| Phase 11-multi-function-native-emission-and-interprocedural-equivalen P04 | 95min | 3 tasks | 9 files |
| Phase 11 P05 | 100min | 3 tasks | 12 files |
| Phase 11 P06 | 105min | 3 tasks | 4 files |
| Phase 11 P08 | 70min | 3 tasks | 6 files |
| Phase 11 P07 | 45 min | 3 tasks | 7 files |
| Phase 11-multi-function-native-emission-and-interprocedural-equivalen P09 | 65min | 3 tasks | 6 files |
| Phase 12 P01 | 38 min | 3 tasks | 3 files |
| Phase 12 P02 | 95min | 3 tasks | 13 files |
| Phase 12 P03 | 130min | 3 tasks | 11 files |
| Phase 12 P04 | 24 min | 3 tasks | 9 files |
| Phase 12-result-payloads P05 | 55min | 3 tasks | 6 files |
| Phase 12-result-payloads P06 | 35 min | 3 tasks | 3 files |
| Phase 12-result-payloads P07 | 45min | 3 tasks | 6 files |
| Phase 12 P08 | 25 min | 3 tasks | 1 files |
| Phase 13-agent-loop-for-interprocedural-defects P01 | 27 min | 3 tasks | 7 files |
| Phase 13 P02 | 58min | 3 tasks | 2 files |
| Phase 13 P03 | 62min | 3 tasks | 5 files |
| Phase 13 P04 | 60min | 3 tasks | 12 files |
| Phase 13 P05 | 55min | 3 tasks | 5 files |
| Phase 13 P06 | 95min | 3 tasks | 8 files |
| Phase 13 P07 | ~25min active | 3 tasks | 5 files |
| Phase 14 P01 | 15min | 3 tasks | 1 files |
| Phase 14 P02 | 15 min | 3 tasks | 17 files |
| Phase 14 P03 | 9 min | 3 tasks | 2 files |
| Phase 14-evidence-instrument-and-honest-scoping P04 | 6 min | 3 tasks | 14 files |
| Phase 14 P05 | 55min | 3 tasks | 4 files |
| Phase 14 P06 | 55min | 3 tasks | 1 files |
| Phase 14-evidence-instrument-and-honest-scoping P07 | 26 min | 4 tasks | 19 files |
| Phase 14-evidence-instrument-and-honest-scoping P08 | 9 min | 3 tasks | 5 files |
| Phase 14 P09 | 385min | 3 tasks | 20 files |
| Phase 14 P11 | 40min | 3 tasks | 7 files |
| Phase 14 P13 | 15min | 2 tasks | 1 files |
| Phase 14 P12 | 55min | 3 tasks | 4 files |
| Phase 16 P01 | 2 min | 2 tasks | 3 files |
| Phase 16 P04 | 2 min | 2 tasks | 2 files |
| Phase 16-branch-match-emitter-port P02 | 8m 26s | 2 tasks | 3 files |
| Phase 16 P05 | 3 min | 1 tasks | 1 files |
| Phase 16-branch-match-emitter-port P03 | 4min | 2 tasks | 2 files |
| Phase 16-branch-match-emitter-port P06 | 12 min | 2 tasks | 6 files |
| Phase 16-branch-match-emitter-port P15 | ~15 min | 2 tasks | 2 files |
| Phase 16 P19 | 15 min | 2 tasks | 2 files |
| Phase 16 P20 | 9 min | 1 tasks | 1 files |
| Phase 16 P18 | 50min | 2 tasks | 5 files |
| Phase 16 P17 | 12min | 2 tasks | 3 files |
| Phase 16 P16 | 4min | 1 tasks | 5 files |
| Phase 16 P21 | 20 min | 2 tasks | 6 files |
| Phase 18 P1 | 15min | 3 tasks | 6 files |
| Phase 18 P2 | 31min | 2 tasks | 11 files |
| Phase 18 P03 | 23min | 2 tasks | 4 files |
| Phase 18 P4 | 35m | 3 tasks | 11 files |
| Phase 18 P5 | 24min | 3 tasks | 9 files |
| Phase 18 P6 | 11min | 3 tasks | 6 files |
| Phase 18 P07 | 3m | 2 tasks | 2 files |
| Phase 18 P08 | 64m | 2 tasks | 2 files |
| Phase 19 P02 | 4min | 2 tasks | 4 files |
| Phase 19 P03 | 9min | 2 tasks | 8 files |
| Phase 19 P4 | 10m | 2 tasks | 7 files |
| Phase 19 P05 | 13min | 2 tasks | 8 files |
| Phase 19 P6 | 9min | 2 tasks | 5 files |
| Phase 19 P07 | 17 | 2 tasks | 7 files |
| Phase 20 P01 | 5min | 2 tasks | 3 files |
| Phase 18 P09 | 63min | 2 tasks | 12 files |
| Phase 21 P01 | 4min | 2 tasks | 2 files |
| Phase 21 P02 | 11 | 2 tasks | 10 files |
| Phase 21 P03 | 10 | 3 tasks | 2 files |
| Phase 18 P10 | 8min | 2 tasks | 4 files |

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
- [Phase 07]: 07-11: consume-on-call closes CR-01/PVG-01 -- resolveCallBinding consumes call arguments (ability-decided), corevalidate independently re-derives from the core artifact alone; D-07-07 corrected, D-07-52 records the accepted implicit-transfer residual — Checkpoint auto-ratified under auto_advance/yolo mode
- [Phase 07]: 07-12: Foreign/Fails made closure-derived over the proven-acyclic call graph in both originvalidate (joinForeignReach/joinFails) and corevalidate (peerJoinForeignReach/peerJoinFails, independently written, no shared helper beyond core.ForeignReachConflict); closes 07-REVIEW.md CR-03/PVG-02 and incidentally IN-01 (index-coupling fix). Phase 07 has no open post-verification gaps.
- [Phase 08]: 08-01: interprocedural fact enters through derivePlaceLoans' forward canonicalization pass (D-08-07), never the backward transfer function; new post-acyclicity check.interprocedural_loan_liveness law closes D-03-02's interprocedural half on relay_escort_witness.lang, with both admission paths proven to agree via a mutation-tested differential (D-08-09).
- [Phase 08]: 08-02: fixed a real callgraph.Order ordering-direction bug in buildInterproceduralSummaries (its result is caller-before-callee, not callee-before-caller as previously assumed) -- required for UsesParam transitivity to hold; deriveFunctionUsesParam's fixpoint-loop shape makes the program-order-vs-reversed cost claim genuinely falsifiable (171.3x observed multiplier at k=512, matching spike S-006's 192x finding).
- [Phase 08]: 08-03: fixture-header + peerDivergenceExpected register the corevalidate residual for the two ACCEPT twins (check admits, corevalidate's still-intraprocedural loanChainIndex refuses via core.move_while_borrowed) as concrete input for Phase 09's peer re-derivation.
- [Phase 08]: 08-03: Pattern B's real .lang twin pair cannot demonstrate a differing end-to-end CLI verdict (computeLoanLastUses' summary-blind AST-shadow path refuses both members identically, ownership.move_while_borrowed) -- tests assert the TRUE observed outcome; the contract-driven backward gate itself stays proven at the checked-core level (08-02's own synthetic test).
- [Phase 08]: 08-04: loanLivenessFixpoint's cycle pre-walk converted to an explicit-stack DFS ported from callgraph.Order, its cycle refusal promoted to a coded check.cfg_back_edge diagnostic, and the worklist given a derived fail-closed bound (factor*blocks*(loans+1), the +1 floor a Rule 1 auto-fix over the plan's literal formula) with a named check.loan_liveness_bound_exceeded refusal, mutation-killed via loanLivenessBoundSeam and proven identity-stable against its own retuning — Success criterion 2 required the only genuinely iterative fixpoint in Phase 08 to terminate under a fail-closed bound with a named refusal rather than a hang or silent truncation; the literal blockCount*distinctLoanCount formula computes zero for any loan-free function and would have refused nearly every legal program, caught by reasoning through the existing suite before committing
- [Phase 08]: 08-05: interprocedural cost gate shipped -- five-shape synthetic call-graph corpus fits buildInterproceduralSummaries' own deterministic work counter to <=1.2 growth exponent against operation count; both hardcoded gate-eligibility chokepoints (measure.Demote, session.QLT02GateEligibleMetrics) widened together and proven to agree; a latent EvaluateBudget bug (literal-string gating instead of set membership) surfaced by that widening was Rule-1-fixed so the new metric genuinely blocks on its bound; ratified hard manifest row + lane:interprocedural-cost-scaling changed-risk lane land the gate in the feedback-budget ledger
- [Phase 08]: 08-06: Mid-phase gate adjudicated all four agenda items from re-run evidence (a/b carried as named debt, c/d resolved), scope-cut trigger did not fire, liveness law declared final; debt register expanded 5->9 items; manifest row re-ratified at gate's own commit; OWN-06/EFF-02 confirmed complete (found already marked by 08-03/08-05 ahead of the gate, a process-ordering deviation documented, not reverted since correct in substance).
- [Phase 09]: 09-01: corevalidate's own forward-derived interprocedural loan-carry peer (derivePeerLoanCarry/chainPeerLoanCarry) folded into the existing v.peerPostorder substrate; buildLoanChainIndex's OpCall consult retires both peerDivergenceExpected accept-twin divergences (D-08-40 resolved); bidirectional seeded-fault companion assertions prove the agreement load-bearing. Discovered and documented (not fixed) a pre-existing originvalidate OpCall-transparency defect this fix unmasks (PHASE-09-DEBT.md D-09-51) -- the full CLI's clean-check claim for both fixtures is not satisfied end-to-end, a separate validator's pre-existing bug.
- [Phase 09]: [Phase 09]: 09-02: generateCallGraphCorpus relocated verbatim from check's costcorpus_test.go to testsupport.GenerateCallGraphCorpus/CallGraphCorpusShapes (D-09-46), unblocking session/corevalidate consumption with zero production import of check; a third gate-eligible metric (peer_closure_recomputed_work_growth_exponent) declared at both chokepoints plus the QLT-02 vocabulary with no manifest row yet (D-09-28); spike-006's QLT-01 registry gap closed (waived, since AllShippedControlIDs predates Phase 08), making go test ./... unconditionally green for the rest of Phase 09 (D-08-43 resolved).
- [Phase 09]: [Phase 09] 09-03: peerDeriveOriginFacts adds an access-mode payload to peerReturnDerivesFromBorrow's existing forward walk; peerOriginContained performs a containment check (never a recomputation of originvalidate's backward combination law); peerForeignOriginOmitted independently re-derives the fourth class with a bounded, function-local walk. peerCallable now consults all four PublishProblemsFor classes, closing D-07-33; TestPeerDoesNotRederiveNarrowedClasses flipped to TestPeerRederivesFormerlyNarrowedClasses. D-09-21's fallback did not fire. OWN-08 stays Pending in REQUIREMENTS.md since 09-08's mid-phase gate also carries it.
- [Phase 09]: [Phase 09]: 09-04: synthetic-shape zero-divergence differential (package check, D-09-50's vehicle split) proves TRU-04 criterion 1 over diamond/deep-chain/dense/parser-shaped/forward shapes; discovered the five shapes never contain a borrow op, making the natural sweep provably vacuous, so Task 3's mutation-kill hand-built a twin_a_accept.lang-shaped fixture instead. TRU-04's 'recursion' shape settled as a cycle-peer witness-agreement differential (self/mutual/indirect), never a liveness one -- category-error disposition recorded in code (D-09-22).
- [Phase 09]: OWN-05 stays Pending after 09-05 (D-09-37): only check+corevalidate proven this plan; interp is Phase 10
- [Phase 09]: No new seam minted for core-layer convention override (D-09-35): existing ParameterContract.Mode closed-set decode check confirmed as the fail-closed control
- [Phase 09]: [Phase 09] 09-07: D-09-49 Q1 settled by enumeration (no fixture exhibits the interaction across all 6 move_while_borrowed fixtures); shadowPathReachableCodes={move_while_borrowed,borrow_conflict} proven by source scan + testOnlyForceUniformLoanJoin perturbation, not merely predicted
- [Phase 09]: [Phase 09]: 09-06: peer's cost sweep injects a genuine borrow into every generated function (raw corpus has none, D-09-04) via injectBorrowForLoanCarry, fits flat-linear against a self-derived 1300 milli-exponent bound (all five shapes ~999-1000), and mutation-kills it via peerClosureUnmemoizedSeamForTest (foldChain memo-write skip) on the 'forward' shape (0.999->1.396 exponent, 4.58x checks at n=512); Result.PeerConsultedFields() closes D-08-26's positive half, proven identical to check's set (D-09-30). No qlt02_budget_manifest.json row added (D-09-28 -- ratified at 09-08's gate).
- [Phase 09]: Deleted computeLoanLastUses; checkInterproceduralLoanLiveness is now the sole loan-liveness decision point (D-09-07/D-09-08/D-09-09)
- [Phase 09]: Pattern B's twin split does not materialize; recorded as new debt D-09-53 (pre-existing deriveFunctionUsesParam defect), not fixed in this plan
- [Phase 09]: OWN-05 split into OWN-05a (Phase 09, Complete) and OWN-05b (Phase 10, Pending); QLT-07 closed for the loan-liveness subset, ratified in 09-VALIDATION.md; OWN-09/TRU-04 requirement texts corrected in place
- [Phase 10]: 10-01: interp executes calls across an explicit []frame heap stack (D-10-21); OWN-05b falls out of observable execution, ability-aware per D-07-11; move-as-copy mutant proven via a synthetic core.Program since no legal Lang source can express the shape
- [Phase 10]: OpCall origin walk now consults a callee's DECLARED PublicOrigin via a narrow, unexported calleeOriginFact map built by exported BuildCalleeOriginFacts — Closes D-09-51; mirrors corevalidate.buildLoanChainIndex's shipped peerLoanCarry precedent, threaded through RecomputeOriginPerReturn/RecomputeOrigin/PublishProblemsFor without widening ValidatePublished/BuildInterface's own signatures
- [Phase 10]: Corrected two check-package negative-control fixtures/tests (negative_control_fails.lang, negative_control_infallible.lang) whose expected diagnostic depended on the same pre-existing D-09-51 transparent-walk defect — relay's declared borrow(buffer) return was never honestly derivable given leaf's genuinely owned return; check's own SEM-06 Callable gate now correctly fires first with core.callee_not_callable
- [Phase ?]: 10-03: pathoracle composes callee paths across OpCall via its own EnumeratePaths (composeCall/composeCarriesOwnLoan), re-deriving per-path (never per-function-contract) whether a loan survives the call; MaxCompositionDepth + compositionCycleError guard recursion independently of MaxPaths/EnumeratePaths' own back-edge guard; D-10-14 discriminating fixture (two-arm callee, one borrows one owns) proven via the D-10-55 seeded contract-hop fault, stitched across two separately-checked fixtures since the shape cannot be one joint Callable program
- [Phase 10]: [Phase 10]: 10-04: MaxCallDepth = 128 declared deliberately BELOW the real 1024-function structural ceiling (inverse of pathoracle.MaxPaths' above-maximum direction); depth-exceeded refusal modeled as a typed Outcome/Event (never a bare error); a subprocess probe (first in this repo) directly observes the host-stack limit and the language-level call-depth bound are structurally unrelated
- [Phase 10]: 10-05: interp's cross-frame drain order (D-10-33/D-10-35) plus corevalidate's per-declaration callee-frame-drain invariant (D-10-34) are the two independent knowers of drop order; cgen's multi-frame pad stays named Phase 11 debt
- [Phase 10]: D-10-59 trigger measured at plan 10-06's gate: 0.19x (71,773 actual vs 375,000 estimated across plans 10-01-10-05) — DID NOT FIRE, no cut taken — All four hard ordering constraints confirmed against git history; plans 10-07/10-08 proceed unchanged
- [Phase 10-trusted-interprocedural-oracle]: Composition depth declared at 3 (D-10-46) with a bidirectional gate (D-10-50); a previously-undocumented corevalidate peerDeriveOriginFacts gap (no OpCall case) discovered and reported per plan 10-07's own escape hatch.
- [Phase 10]: Result.LoanEndpoints() captures loanEndpointsMatch's computed value in-flight during Validate rather than recomputing via a fresh throwaway validator, avoiding a silent divergence from missing interprocedural loan-carry facts
- [Phase 10]: The three-way endpoint comparator skips a CFG-carrying function when corevalidate did not fully validate the program, since corevalidate's fail-fast replay may never reach that function's own loan-endpoint check
- [Phase 10]: D-10-55/D-10-41 seeded-fault companion tests landed in pathoracle_test.go and corevalidate_endpoint_test.go respectively (not session_peer_gate_test.go) because both fault seams are _test.go-only symbols invisible outside their own package's test binary
- [Phase 10]: Phase 10 plan 10-09: interp is STABLE per the roadmap's Phase 11 precondition -- a 6-program byte-for-byte golden corpus (testdata/phase10/interp_oracle/), determinism proven at -count=10 over the full corpus, and the structural coverage floor (D-10-58) shipped as three enumerated tables in 10-VALIDATION.md with zero blank cells. SEM-08/SEM-09/OWN-05b all complete; Phase 10 done.
- [Phase 11]: 11-01: Q-01=BRANCH A (accepted), Q-02=BRANCH A (hole reproduces) -- committed single-branch verdict tests decide plan 11-08's reducer proceeds unnarrowed and plan 11-07 ships a real eighth cache input — Both spikes settled as committed Go tests, not prose, per D-11-29/D-11-41; PHASE-11-DEBT.md opened at phase start with nine recorded-not-built decisions and five Phase 10 carry-forward items
- [Phase 11]: Q-04 topology found empirically: composition-only LTO divergence requires a 4-way TU split (callee, writer, a SEPARATE coordination wrapper calling the callee twice, caller) -- merging wrapper+main lets the compiler prove pointer identity and refuse the hoist at every tier — Five of six tried topologies either diverged without LTO or never diverged at all on Apple clang 21.0.0
- [Phase 11]: Task 3 ratification auto-selected Option A (approve all five items: D-11-20/21/22 roadmap amendments, D-11-09/10 zero-attribute terminal state, D-11-39 QLT-06 split) since it carried no gate=blocking-human — Auto-mode checkpoint protocol; flagged human_judgment:true in SUMMARY coverage for human review
- [Phase 11]: 11-03: callgraph.EntryFunction resolves via in-degree-zero roots with a closure-size tie-break (never Functions[0]); emitProgram is the new whole-program C17 assembler with emitCall as sole Lang-to-Lang call writer — Reconciles the plan's own conflicting must_haves (unreachable function must not block resolution vs never guess); native.go's execution validator widened (Rule 3) for legitimate multi-FunctionID/non-terminal-return events, a pre-existing single-function assumption Phase 11 first exercises
- [Phase 11]: Mid-phase gate ratified PASS (11-MIDPHASE-GATE.md): zero-attribute state terminal for Phase 11, NAT-05 written up as requirement weakened by evidence; waves 4-6 admitted.
- [Phase 11]: 11-05: session's three CLI run sites were already widened at plan 11-03; Task 1 re-verified the 29-site baseline and KEPT every remaining single-function guard as legitimately fixed-fixture-bound (Phase 2-7) or deferred to plan 11-08 -- zero additional widening needed. — Guard ledger's mechanical awk re-verification (29, not the ROADMAP's expected 32) plus row-by-row disposition in 11-GUARD-LEDGER.md
- [Phase 11]: 11-05: DiamondSharedLeaf discovered a genuine event-ID collision (D-11-51) when a callee is invoked from two static call sites; DivergingCallee is not expressible in a multi-function program this phase (D-11-52, extends D-11-02). Both recorded as new PHASE-11-DEBT.md debt, not silently fixed or narrowed. — Real fixes touch interp.go/cgen_program.go/emitMatch's multi-function generalization, outside this plan's files_modified and each large enough to need its own reviewed plan
- [Phase 11]: Q-01 BRANCH A: drop-call-site ships as a live whole-program reduce move, not narrowed behind RefusedShapes(). — corevalidate accepts the core-level OpCall-to-OpCopy rewrite (11-01 spike).
- [Phase 11]: 11-07: cache.DeclaredInputNames() widened to eight (cgen_source appended, D-11-41 closed); QLT-06 split QLT-06a/QLT-06b recorded in 11-QLT06-ABSTENTION.md per OWN-05a/05b precedent — Q-02 BRANCH A confirmed the stale-cgen cache-reuse hole reproduces; whole-program FixtureSource hash strictly dominates any closure key so QLT-06b is discharged by abstention, structurally gated by dual import scans
- [Phase 11]: 11-09: foreignCallSequenceFor widened to a two-independent-knower guard (static callgraph.Order walk + dynamic -O0 event-stream walk), closing the multi-function nil-compares-nil slippage hole (D-11-33); QLT05Reverify implements strict, cold-start re-verification with no positional/causal-role fallback, reported as control:reduce.reverified (D-11-32/D-11-37); anti-vacuity gate proven on both sides via testdata/phase11/multi_function_reduce_gate.lang (D-11-34); 11-VERIFICATION-INPUTS.md collects the phase's four evidence-weakened claims. QLT-05 closed. Phase 11 complete (last plan, wave 6).
- [Phase 11]: Phase 11 UAT closed with ZERO human verification: both items 11-VERIFICATION.md routed to a human were converted into committed tests instead of being answered once. WR-01 is FIXED not accepted — reduce.Seed.Validate (internal/compiler/reduce/seed_validate.go) refuses a multi-function seed whose EntryFunctionID is empty or unmatched with core.seed_entry_invalid, mirroring callgraph.EntryFunction's own fail-closed shape; TestSeedEntryHazardIsReal is an anti-vacuity control that fails if the deletion hazard ever stops existing. The 11-MIDPHASE-GATE.md CLI-check divergence is PINNED not filed as debt — session_admission_divergence_test.go sweeps every committed .lang fixture through both admission surfaces and asserts the divergence set exactly, failing on a new divergence AND on a silently resolved one (mutation-checked both ways). — Operator directive: shift left, automate the world, target 0 human UAT, in CI only where value recurs. Both land in CI unchanged (`go test ./...` already runs on both hosts); no debt rows added, because a debt note is read once and a test is checked forever.
- [Phase 11]: 11-SECURITY.md written at phase close: 28 threats (T-11-01..T-11-27 plus T-11-SC), all closed, threats_open 0 at ASVS L1 with block_on high. Register was authored at plan time (all 9 plans carry <threat_model>), so this verified a pre-declared register rather than reconstructing one. Every mitigate row was closed by locating its named control and RUNNING it, not by reading mitigation prose — the Verification Evidence table names each one, so deleting or renaming a control breaks the audit. Note: Phase 10 by contrast ran with security_enforcement=true and never produced 10-SECURITY.md.
- [Phase 12]: D-12-31 branch ratified as RE-DEFER: the measured N=1 differential shows emitProgram refuses 4/5 single-function shapes and diverges on the 5th, so D-12-36's fallback trigger fired.
- [Phase 12]: PHASE-11-DEBT.md's D-11-02 landing-phase commitment is superseded with a stated reversal quoting the original text verbatim, per the D-09-08/D-09-30/D-10-27 precedent.
- [Phase 12]: The six legacy cgen emitters stay in cgen.go, byte-untouched, with no currently-owned landing phase for their eventual deletion.
- [Phase 12]: Ratified as proposed: source spelling Ok(Buffer)/Ok(v)=>/Ok(v), six diagnostic codes, one-globbing-test D-12-19 replay
- [Phase 12]: check.payload_arity_mismatch required an additive ast.MatchArm.ConstructBinder field (Rule 2) since the collapsed single Binder field made the refusal structurally unconstructible
- [Phase 12]: D-12-27's resource-payload refusal matches on foreign-return-type name equality (deliberately over-inclusive, fail-closed) since this project's type system has no ability-derived resource provenance marker
- [Phase 12]: originvalidate's RecomputeOriginPerReturn needed a sourceOf-indexing fix (PayloadTargetID, not TargetID) for OpDestructurePayload before its new origin-walk arm could be reachable at all
- [Phase 12]: Cross-package test-only mutation-kill seams must be production-visible functions (corevalidate.SetDisableCyclePeerForTest's D-07-42 shape), never export_test.go symbols, which are invisible outside the defining package's own test binary
- [Phase 12]: 12-06: check.duplicate_payload_type closes CR-01's source-layer half (D-12-44); a control asserts the code directly since interp/cgen share the flawed derivation; WR-01's nullary-binder message now names the construction side, code unchanged.
- [Phase 12]: PayloadSlotSwapInjectedWriteCount is production-visible (cgen.go), not export_test.go — cross-package need mirrors SetPayloadSlotSwapForTest's own precedent
- [Phase 12]: D-12-43 ratified at plan 12-08 Task 1's blocking-human checkpoint (developer chose 'ratify') — Closes 12-VERIFICATION.md's human-verification item 2; no enabling work scheduled, no future phase named as owner
- [Phase 12]: D-12-44 recorded: CR-01's disposition is FIX (check.duplicate_payload_type + shared core resolver), restriction on the source language, lifting condition named as GEN-01 — CR-01 was the code review's sole CRITICAL finding and the verifier's only hard gap
- [Phase 12]: D-12-45 recorded: WR-01/WR-02/IN-01 dispositions, all closed, none deferred — Every secondary review finding must have an explicit recorded disposition
- [Phase 13]: 13-01: check.interprocedural_loan_liveness's move_after_interprocedural_loan repair is emitted ONLY for the BACKWARD direction (call is the loan's own recorded last use) — the FORWARD direction (loan propagated through the call onto a place read still later) is not fixed by swapping the move and call statements, discovered empirically by splicing the repair onto real testdata/phase07-08 fixtures both ways. A `callIsLastUse` gate added to `interproceduralLoanLivenessDiagnostic`. Re-pinned FIVE (not the plan's stated four) check_ordering_stability_test.go rows — phase08/twin_b_accept.lang also carries this code. Fixed a latent syntax/parser.go call-binding Span truncation (Rule 1) the new repair's :stmt span channel exposed.
- [Phase 13]: Blame resolver (resolveBlame/resolveCycleBlame) built and exhaustively tested but not wired into the eight existing diagnostic emission sites -- D-13-04 verified empirically that all eight already select B2's Primary span, so wiring is a no-op today; the resolver is ready infrastructure for the first B1-shaped defect class. — Avoids touching identity-bearing Primary spans on published diagnostics for zero behavioral gain, per D-13-04's empirical verification (TestBlameMovesNoPrimarySpanToday).
- [Phase 13]: checkCallGraphAcyclic routed through the new calleeBeforeCallerOrder helper instead of calling callgraph.Order directly, making it check.go's sole callgraph.Order call site. — Required to satisfy the plan's 'exactly one non-comment callgraph.Order call site' acceptance criterion, and strengthens D-13-03's 'no new ordering authority' claim by construction; behavior-preserving since only the error is ever consulted.
- [Phase 13]: 13-03: peer-disagreement refusal scoped to core-claims-and-AST-disagrees, not core's mere absence — core.Program.Functions only includes cleanly-checked functions; treating absence as disagreement broke most of the existing explain corpus
- [Phase 13]: D-13-02a resolved: no B1-shaped interprocedural diagnostic is constructible at this language maturity (sameType is a precondition of every interprocedural pass, checked before any call site is reached).
- [Phase 13]: Set A repair-emission logic implemented with the uniqueness-gate comparison target changed from contract.ParameterType to the caller's own type fact — the literal plan-specified comparison is mathematically unsatisfiable at this language's single-type-per-function maturity.
- [Phase 13]: 13-06: DX-06/DX-07 requirement checkboxes deliberately left unmarked (Pending) despite the plan completing -- D-13-10a resolved use_matching_argument as unrepairable (2 of 3 repairable classes, not 3) and D-13-28's twin pair proved only detection-site-vs-position-hardcoded blame, not the B1-shaped discrimination DX-06's literal text describes (D-13-02b: unconstructible at this maturity). Full evidence in 13-06-SUMMARY.md; final ratification deferred to 13-07's checkpoint per D-13-10a's own closing instruction.
- [Phase 13]: 13-07: D-13-33 adjudicated Option B (developer, blocking-human checkpoint) -- testdata/phase6 move/borrow pairs are structurally identical (alpha-rename only), predicate kept unweakened, carried as permanent M001 evidence debt (D-13-34, PHASE-13-DEBT.md) rather than fixed or reversed
- [Phase 13]: 13-07: DX-06 and DX-07 ratified Partial (not Complete) -- B1 contract-violation blame structurally unreachable at this maturity (D-13-02b, D-12-43 precedent); use_matching_argument withdrawn as unrepairable, every Lang function sharing one type fact (D-13-10a). Both recorded in PHASE-13-DEBT.md, REQUIREMENTS.md traceability updated
- [Phase 13]: 13-07: 13-VALIDATION.md corrected against real test names -- two instances of the same go-test-run-matches-nothing defect class found and fixed (D-13-09a's TestOrderingStability -> TestInterproceduralDiagnosticOrderingStability; import-boundary row's TestImportBoundary pattern, which missed the actual lint TestRepairDriverImportsStayOutsideInternal). wave_0_complete/nyquist_compliant set true from evidence; go test ./... green (25 packages)
- [Phase ?]: Groundedness lint (EVD-01) shipped end-to-end: static Go test index over *_test.go, Tier-A document scanner via phaseArtifactGlob, R1/R2/unparseable classification, 26-record pinned violation frontier, three-fault non-inertness proof. Fixed a real backtick-awareness bug in table-row splitting and a Contains-vs-HasPrefix substring bug found via testing against the live corpus. — Both bugs were caught by running the lint against the real .planning/** corpus rather than trusting the plan's literal algorithm description -- exactly the discipline this phase exists to install.
- [Phase ?]: 14-02: recoverRegion advances past the unexpected token before measuring the discarded extent, so causes[0].span never overlaps primary_span; the third spiral member's skipped region lands at exactly one token.
- [Phase ?]: 14-02: skipped_region cause attached only when recoverRegion actually discarded >=1 token, confining published-ID churn to fixtures that reach the declaration-recovery arm with a non-empty discard.
- [Phase ?]: 14-02: diagnostic_distinctness_test.go lives in package check_test (not package check) since session.CheckCommandFile would otherwise create a check->session->check import cycle.
- [Phase ?]: [Phase 14]: 14-03: DX-09 closed -- unrepairable decline now carries DiagnosisCodes/DeclineReason/BestApplicability, populated via a new classifyDecline sibling (not a widened selectRepair, to keep antitheater_test.go/repair_test.go byte-unchanged); DiagnosisCodes is a comparable diagnosisCodeList string type (custom JSON marshaling) since a bare []string broke Outcome's existing == comparison in antitheater_test.go; no_diagnostics decline sets diagnosis_code to the reason itself rather than leaving it empty, honoring the plan's unqualified never-empty truth
- [Phase ?]: 14-04: mechanized PRC-01's closed owning-phase vocabulary (P<NN>|CLOSED(sha)|UNOWNED(witness)) inside checkDebtRegister and migrated all twelve pre-existing debt registers to it cell-format-only, via a documented rule (CLOSED only on explicit closure language, P<NN> on a single named phase, else UNOWNED); D-09-53 migrated CLOSED (not the P10 its own cell text implies) since PHASE-10-DEBT.md's D-10-27/D-10-30 explicitly withdrew its premise — PRC-01 requires every debt item to name a resolvable owning phase; the prior law only checked non-empty
- [Phase ?]: 14-04: opened PHASE-14-DEBT.md and closed EVD-07's outstanding half by registering the -flto multi-function inertness claim (D-14-45) with an owning-phase cell; PROJECT.md's DX-06/-flto text confirmed already correct (commit d21db90), not re-edited — the debt row was the one remaining task; the text correction had already landed
- [Phase 14]: 14-05: Machine-checked LANGUAGE-MATURITY.md via independent go/parser AST re-derivation (never shelling out to the doc's own awk command); corrected guard total 32->22 and removed the stale reduce-package row (Phase 11 already widened reduce.Reduce to accept multi-function seeds). Recorded a real cold go test ./... wall-clock baseline (191.89s) as an observed qlt02_budget_manifest.json row.
- [Phase ?]: 14-06: D-14-11 role-based document scoping wired with promotion; grep execution (R3) and per-branch groundedness (R2b) wired; frontier re-pinned 26->125 entries (full .planning/**/*.md scope, no documents yet declare the proposal exemption)
- [Phase 14-evidence-instrument-and-honest-scoping]: 14-07: closed four-kind witness grammar (probe:/callsite:/escape:/env:) mechanized in checkDebtRegister; four executed probes back D-13-02b/D-13-10a/D-13-34/D-14-45/D-12-43/D-11-02; module-wide suppression enumerator requires every t.Skip to cite a resolvable witness (scoped to PENDING-only for the live module's blanket comment/string-literal surface after a full-grammar attempt produced 115 false positives against this codebase's own unrelated escape:/D-XX-NN conventions); .planning/UNREACHABLE-CLAIMS.md generated and byte-compared, holding exactly the six known qualifying rows. EVD-03/EVD-04 complete. 87 debt-register rows across 02-DEBT.md..PHASE-10-DEBT.md remain unmigrated to Grade/Witness, recorded as debt.
- [Phase ?]: Task 2's per-row-exclusion removal shipped as a single test commit (no separate GREEN): Task 1's collapse already made the real mutation table satisfy the check; RED was demonstrated by temporarily clearing the real EscapeID field, observing failure, then reverting before commit. — Data already valid post-Task1, but genuine RED evidence was still produced rather than asserted from memory.
- [Phase ?]: Found and fixed two stale PENDING-05-08 prose occurrences beyond the plan's named five sites (a test's own negative-assertion literal, and a witness_registry_test.go hand-off comment), since the plan's acceptance criterion is a repo-wide grep, not a fixed site list. — Rule 2 - missing critical: satisfying the literal 'marker survives nowhere' truth required a full repo scan, not just the five research-identified sites.
- [Phase ?]: EVD-02 grade cap: declared grade authored at its derived ceiling (capped, never MUTATION-KILLED) across all 14 migrated VALIDATION docs; 6 rows derive below shipped verdict, recorded as PHASE-14-DEBT.md rows D-14-48..54 rather than suppressed
- [Phase ?]: Run-record generation for the ~230-row corpus is genuinely several minutes (not milliseconds as T-14-56 first assumed); a parallelism attempt thrashed and was reverted to sequential, documented in scripts/evidence-run-record.sh
- [Phase 14]: 14-10 (Nyquist reconciliation, closes the phase): 66 runnability/groundedness/grep findings reconciled under a closed four-verdict vocabulary (renamed/superseded/obsolete-by-design/under-scoped) as debt rows with a witness in PHASE-14-DEBT.md (D-14-55..D-14-120), never rewriting the archived documents they were found in; the pinned groundedness frontier emptied for those three classes (124->58 entries), per-branch (R2b, 10 entries) stays pinned and owned by P20 (ROADMAP QLT-10). `.planning/EVIDENCE-RECONCILIATION.md` generated and byte-compared; `scripts/assert-reconciliation-touched.sh` couples future archive edits to it. Filling in 14-VALIDATION.md's own Per-Task Verification Map (31 real rows) exposed a genuine performance regression -- an unanchored `TestValidationRowGradesAreEarned` pattern substring-matched the expensive `TestValidationRowGradesAreEarnedOverArchivedCorpus` itself, recursively re-invoking it inside the run-record generator and pushing the session package past Go's 600s default timeout; fixed by anchoring each alternation branch with a trailing `$` (not wrapping the whole group in `^(...)$`, which breaks the R2b per-branch splitter). Attempted removing 14-VALIDATION.md's stale grade-bar exemption; reverted after it exposed the same corpus-wide timeout risk on nine unrelated files under this specific machine's load -- recorded as debt D-14-121, not landed under uncertainty.
- [Phase ?]: 14-11: raised evidenceRunRecordTimeout 300s->480s with a measured-honest doc comment, added a 0.75 margin fraction that fails closed on a near-timeout run, and replaced raw-pattern run-record batching with resolvedPkgPatterns (anchored, index-resolved names) -- closing WR-01/WR-02; corpus-wide measured elapsed dropped 269-347s -> ~90s
- [Phase ?]: 14-13: buildConstraintAllowlist (darwin/linux/amd64/arm64/cgo) closes EVD-04's inert //go:build branch; scanSuppressionSurfaces reordered so MatchFile gates only AST passes, never the textual constraint pass, closing T-14-13-02's constraint-hides-itself hole
- [Phase ?]: 14-12: removed 14-VALIDATION.md's file-scoped grade-bar exemption, confirmed the corpus-wide >=EXERCISED bar with a complete run record (90-92s vs 480s budget), added the row-scoped validationGradeBarRowExemptions narrowing (5 entries, each debt-witnessed) since Task 2's own ship-empty expectation was falsified by real execution, and closed D-14-121/D-14-53 as CLOSED(128ecec). EVD-02 and PRC-01 complete.
- [Phase 15]: Ratified Phase 15 /2 invocation grammar and caller-owned callee_function_id contract before publication.
- [Phase 16]: Preserved public dispatch while proving direct ordinary-linear emitProgram behavior.
- [Phase 16]: Derived schema-2 live_resources from emitProgram without introducing a resource ledger.
- [Phase 16]: Keep restrict admission evidence test-local; no production routing changes before Plan 05 human disposition.
- [Phase 16]: Linux restrict-probe lane is unresolved on macOS and cannot be counted as success.
- [Phase 16]: Admit nullary branch blocks through emitProgram while retaining explicit refusals for match-only, payload, foreign, and N=1 by-pointer shapes.
- [Phase 16]: Keep main as the sole schema-2 document writer; branch helpers only record events and return C values.
- [Phase 16]: Selected cut-m004: unavailable Linux evidence means macOS-only probe results cannot admit emitLinearBorrowedByPointer or any by-pointer family in M003.
- [Phase 16]: emitProgram reads checker-owned payload layout and shared alternative identity; defect evidence is written before the narrow abort helper.
- [Phase 16]: Applied cut-m004: no pointer-specialized family is admitted by emitProgram at any program cardinality.
- [Phase 16]: Recorded foreign and by-pointer emitter families as M004 debt; one-TU pure-Lang -flto remains behaviorally inert evidence.
- [Phase 16]: AST-derived public-emitter calls are authoritative; registry rows classify each exactly once.
- [Phase 16]: Mutation controls require intended validator error classes, not generic cardinality drift.
- [Phase 16]: 16-19: replaced stale archived witness names with the current Phase 17 witness and used deterministic injected MachineFacts for budget audit tests; preserved the existing generated-view disposition rules.
- [Phase 18]: 16-18: Keep the Phase 16 research command explicit and executable while preserving its characterization purpose.
- [Phase 18]: 16-18: Pin every current groundedness finding and assign all surviving R2b rows to P20 under QLT-10; preserve exact-set equality and empty R1/R2/R3 gates.
- [Phase 16]: Refresh the Phase 16 validation record only from the live consumer-derived pair export and actual sequential producer output.
- [Phase 16]: Bind each checked-in record body to exact pair bytes and require persisted per-pair and batch completion witnesses, including under re-digested mutations.
- [Phase 18]: Phase 16 Plan 21: Keep Phase 11 assurance test-only; require exact cut-m004 refusal before loading provenance-checked frozen C.
- [Phase 18]: Phase 16 Plan 21: Keep zero-call entry ambiguity as an independently typed refusal, without by-pointer frozen-evidence mapping.
- [Phase 18]: Phase 18's current source frontier is the parser diagnostic syntax.expected_linear_result at terminal match; preserve it until computed match is admitted.
- [Phase 18]: Represent a linear prefix followed by one terminal match in the existing body and branch CFG.
- [Phase 18]: Keep CTL-01 open until full Phase 18 acceptance is verified.
- [Phase 18]: Core computed scrutinee IDs must resolve to a definition in the shared branch entry prefix; ID-less matches remain parameter-only.
- [Phase 18]: Origin derivation is bounded to earlier operations in the return arm and shared entry prefix, using block facts even if Match metadata is absent.
- [Phase 18]: Computed Result match arms use an explicit edge-bound typed value place when the return type differs from the scrutinee type.
- [Phase 18]: Serialize payload-bearing terminal ADTs with their runtime payload while retaining tag-only branch dispatch.
- [Phase 18]: Seed interpreter data inputs with the canonical payload values emitted by native entry setup.
- [Phase 18]: Keep CTL-01 through CTL-03 open until full Phase 18 verification.
- [Phase 18]: 18-06: Keep S-010 loan endpoints within existing point and edge kinds, with the computed-match prefix participating in bounded CFG liveness.
- [Phase 18]: Plan 18-08 kept CI unchanged because existing macOS/Linux full and race suites already cover Phase 18 and duplicate focused coverage costs about two minutes per host.
- [Phase 19]: 19-02: store U64 as a discriminated interpreter value and serialize it as canonical decimal while preserving prior scalar projections
- [Phase 19]: 19-02: require D-12-18 corpus replay and seeded projection mutation evidence before OpConst routing
- [Phase 19]: Numeric syntax remains lossless in RHS.Source; checker owns numeric interpretation.
- [Phase 19]: Lexer consumes adjacent Unicode letters and digits in malformed numeric candidates.
- [Phase 19]: U64 is a zero-argument structural scalar granting copy, drop, share, send, and escape abilities.
- [Phase 19]: OpConst carries only a canonical decimal U64 semantic value in an omitempty ConstU64 field; source spelling stays in syntax/AST.
- [Phase 19]: The checker uses checked 64-bit radix conversion, rejects overflow at the literal span before core creation, and applies no contextual numeric conversion.
- [Phase 19]: Corevalidation independently parses canonical decimal OpConst values and re-derives U64 abilities.
- [Phase 19]: Path and origin analysis treat OpConst as a source-free root.
- [Phase 19]: Supplemental branch U64 facts use the type:u64 identity and are independently constrained to zero-argument U64.
- [Phase 19]: Phase 19-06 emits stdint exact-width support only when the checked program uses U64.
- [Phase 19]: Native U64 decimal input and JSON string output use checked bounded digit-by-digit conversion.
- [Phase 19]: Use the existing schema-2 comparison projection for the interpreter before comparing all four execution tiers.
- [Phase 19]: Keep stale Phase 11 maturity and reconciliation findings as regression debt outside Plan 19-07.
- [Phase 20]: M003-open checksum refusal is treated as a historical reconstruction, not a contemporaneous fixture pin.
- [Phase 18]: Keep cache.input_undeclared stable while unwrapping typed Clang probe failures through Cause.
- [Phase 18]: Use a finite 30-second default per subprocess; explicit runner timeouts and parent cancellation remain authoritative.
- [Phase 18]: Keep Plan 08 CI disposition unchanged because existing macOS and Linux full and race jobs cover Phase 18.
- [Phase 18]: Regenerate the stale validation corpus from the live pair exporter and actual sequential test results.
- [Phase 21]: Phase 21 Plan 01: exit and discharge rules are machine-checkable; contract structure does not admit emitters or prove runtime cleanup.
- [Phase 21]: Retire unreachable program-lowering bodies and their exclusive helpers while retaining metadata/classification APIs and all refusal boundaries.
- [Phase 21]: Treat archived foreign and by-pointer fixtures as historical evidence, not proof of current emitted behavior.
- [Phase 21]: Use the existing Phase 14 multi-function fixture and run direct emitProgram output through the shared interpreter/-O0/-O3/-O3 -flto comparator.
- [Phase 21]: Keep compiler measurement opt-in; behavioral equality does not prove optimizer activity, performance, cleanup, or other hosts/toolchains.
- [Phase 18]: Stream schema-2 terminal tags and payloads through the bounded JSON writer without tag-sized storage.
- [Phase 18]: Emit terminal scratch buffers only for payload types present in the branch.

### Pending Todos

Cleared at the M002 close. All three M002 pre-phase spikes (S-006 interprocedural
liveness cost-scaling, S-007 recursive stress corpus, S-008 Nyquist fold-in cost)
are answered and their gates released; the `Result` payload / interprocedural-
origin interaction probe was answered before Phase 12 planning. Archived detail:
`.planning/spikes/` and `.planning/milestones/M002-ROADMAP.md`.

Carried into M003 as cheap, unowned cleanup:

- `/gsd-validate-phase 07`, `08`, `11`, `12`, `13` — each has a VALIDATION.md
  that `validate-phase` never reconciled. Phase 13's `nyquist_compliant: true`
  was genuinely earned (plan 13-07 re-ran every Per-Task Verification Map row
  and fixed two ungrounded `-run` patterns); only the lifecycle marker is stale.

- `/gsd-secure-phase 10` — Phase 10 ran with `workflow.security_enforcement=true`
  but produced no `10-SECURITY.md`.

- D-13-34 — M001's `testdata/phase6` move and borrow held-out/derivation pairs
  are alpha-renames of each other, not structurally distinct programs. A hole in
  *shipped M001* evidence, found only because Phase 13 took the stricter
  retro-strengthening branch and reported what it found.

### Blockers/Concerns

**No open blockers.** M002's only `blocker`-severity item ever filed (D-12-44,
CR-01's duplicate payload-type ambiguity) closed in plans 12-06 and 12-07.

**Closed at the M002 boundary** — do not re-derive these as open: D-03-02 (closed
in Phase 09, both admission layers), D-04-30 (`Result` payloads, closed in
Phase 12), "M001 ships without Lang-to-Lang calls" (closed in Phases 07-11), and
plan 09-09's architectural halt (resolved; `computeLoanLastUses` deleted under
D-09-08 authorization).

**10 open, unowned debt items carry into M003.** Full table and cluster analysis:
`.planning/milestones/M002-MILESTONE-AUDIT.md` §4. Three clusters, not ten
independent problems:

1. **Single-function emitter deletion** (D-11-02 → D-12-36, plus D-11-27).
   Deferred in Phase 11, re-deferred in Phase 12 with the reversal stated openly.
   **A third deferral would violate D-10-60, a no-third-deferral rule this
   milestone wrote for itself.** M003 must either land the deletion or retire
   D-10-60 explicitly.

2. **Event identity** (D-11-51 → D-12-21). Shared-leaf diamond call graphs
   collide on event identity; D-12-21 cannot close until D-11-51 does. A real
   dependency chain sitting unowned across a milestone boundary.

3. **The single-type-per-function invariant** (D-13-02b, D-13-10a, plus Phase
   13's twin-pair sub-finding). One root cause:
   `sameType(ReturnType, Parameter.Type)` is enforced at every function's
   admission, so a callee cannot contradict its own contract and the real fix
   always lands in the caller. These close together, automatically, the moment
   return type may differ from parameter type. The blame resolver and its
   exhaustiveness guard are already built and waiting.

Plus D-10-C04 / D-12-43, and the three acknowledged Phase 10 deferred items
recorded under `## Deferred Items` below.

**Residual Phase 10 trust gaps still open** (authoritative detail now in
`.planning/milestones/M002-phases/10-trusted-interprocedural-oracle/`):

- `corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case — fail-closed
  (conservative, not unsound), but it constrained which fixtures Phases 10-11
  could express. Acknowledged deferred item.

- 10-REVIEW.md WR-01 — `corevalidate.peerCalleeFrameDrained` walks FORWARD over
  `linear.Operations`, correct only under an unstated, unenforced
  declaration-order assumption; its own cited precedent walks BACKWARD.

- 10-REVIEW.md WR-02 / D-10-19 — the four peers' independence is enforced by
  hand-curated per-package import lists, and `originvalidate`'s permits
  `callgraph` while `corevalidate`'s forbids it. Defensible by review, not by
  mechanism — and NAT-06 leans on that independence.

- The D-09-51 negative-control verdict flip (`negative_control_fails.lang`,
  `negative_control_infallible.lang` moved from
  `check.interprocedural_loan_liveness` to `core.callee_not_callable`) was
  flagged for human review and **that review still has not happened** — this is
  D-10-C04, and D-11-27 records that no phase claims it.

- `interp.Run`'s `!function.HasClosedBody()` guard is provably unreachable —
  dead defensive code, harmless, recorded so a future reader does not mistake it
  for live protection.

**Standing process rules** adopted after three M001 gate failures that shared one
shape — a green test whose reachable input space omitted the hard case:
mutation-kill every differential, interrogate what inputs a property test
actually reaches, and drive the shipped binary on hand-written programs rather
than only the gate's own corpus. M002 added a fourth: **an integration checker
that grades requirements from wiring will convert an honest partial into a false
green**, because wiring is exactly what a structurally unreachable defect class
still has. Grade requirements against the tree, not the wiring diagram.

- The Phase 16 gate's earlier full-suite and Phase 18 pathoracle findings were
  resolved by later M003 work. Treat those old blocker notes as historical; the
  M003 phase artifacts and audit are the current record.

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

- **M002 closed 2026-09-14.** The scope-cut trigger never fired: Phase 12 stayed
  in, QLT-07 closed for its loan-liveness subset, and Phase 13 shipped without
  narrowing — its two shortfalls (DX-06, DX-07) are structural limits of the
  current type system, ratified as terminal findings, not scope cuts. Both
  mandatory mid-phase gates ran and were ratified in writing. ROADMAP.md is now
  a milestone index; M002's full phase detail lives in
  `.planning/milestones/M002-ROADMAP.md`.

### Quick Tasks Completed

| # | Description | Date | Commit | Status | Directory |
|---|-------------|------|--------|--------|-----------|
| 260922-hfs | Correct Phase 17 Plan 07 traceability and reverify Phase 17 | 2026-09-22 | c7f3692 | passed | [260922-hfs-correct-phase-17-plan-07-traceability-so](./quick/260922-hfs-correct-phase-17-plan-07-traceability-so/) |
| 260923-nvq | Automate verification by default and record project preference | 2026-09-23 | dd257c5 | passed | [260923-nvq-default-to-automated-integration-end-to-](./quick/260923-nvq-default-to-automated-integration-end-to-/) |
| 260924-djg | Avoid groundedness false positives from Phase 18 preparatory commands | 2026-09-24 | c349f52 | passed | [260924-djg-avoid-groundedness-false-positives-from](./quick/260924-djg-avoid-groundedness-false-positives-from-/) |
| 260924-gee | Refresh derived language maturity guard and corpus counts | 2026-09-24 | f5bd7ae | passed | [260924-gee-refresh-language-maturity-derived-guard](./quick/260924-gee-refresh-language-maturity-derived-guard/) |
| 260924-k1v | Reconcile pathoracle loan endpoints for Phase 18 computed match | 2026-09-24 | bc71633 | passed | [260924-k1v-reconcile-pathoracle-loan-endpoints-for-](./quick/260924-k1v-reconcile-pathoracle-loan-endpoints-for-/) |
| 260924-kto | Restore historical core compatibility and Phase 18 liveness acceptance after full-suite regressions | 2026-09-24 | e91628c | passed | [260924-kto-restore-historical-core-compatibility-an](./quick/260924-kto-restore-historical-core-compatibility-an/) |
| 260924-tsl | Fix Phase 19 full-suite regressions | 2026-09-24 | ed237ab | passed | [260924-tsl-fix-phase-19-full-suite-regressions-refr](./quick/260924-tsl-fix-phase-19-full-suite-regressions-refr/) |
| 8 | Complete Phase 18 automated UAT with zero manual checks | 2026-09-25 | dfa5598 | passed | — |
| 9 | Record GSD no-loop rule: compare routed command inputs and state transitions before rerun; resolve unchanged blockers rather than repeating commands. | 2026-09-25 | 550bd0f | — | — |
| 10 | Refresh stale Phase 15 and 19 verification and route GSD to Phase 21 | 2026-09-25 | 0124de9 | passed | — |
| 260926-bkj | Refresh Phase 15 evidence, repair CI planning guards, and record shift-left/no-loop defaults | 2026-09-26 | 27cb078 | passed | [260926-bkj-close-the-ci-planning-integrity-findings](./quick/260926-bkj-close-the-ci-planning-integrity-findings/) |
| 260926-ewj | Fix Phase 18 regression gate failures: preserve schema-1 emitter bytes and refresh corpus snapshot | 2026-09-26 | 7be22ab | passed | [260926-ewj-fix-phase-18-regression-gate-failures-pr](./quick/260926-ewj-fix-phase-18-regression-gate-failures-pr/) |
| 260926-ffr | Ensure Phase 18 payload mutation seams restore after runner failure | 2026-09-26 | fe92bc3 | passed | [260926-ffr-ensure-phase-18-payload-mutation-seams-r](./quick/260926-ffr-ensure-phase-18-payload-mutation-seams-r/) |
| 260926-uzs | Archive completed Phase 21 under provisional M004 and refresh the verification/evidence handoff | 2026-09-27 | 397753f | Verified | [260926-uzs-prepare-the-completed-phase-21-for-the-m](./quick/260926-uzs-prepare-the-completed-phase-21-for-the-m/) |

## Deferred Items

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| deferred_items | Phase 10/deferred-items.md: originvalidate_test.go transitiveImportsViolation spawns `go list -deps` via bare exec.Command with unbounded .Output() — violates TestSourceNeverSpawnsUnboundedProcesses (D-02-01) | acknowledged | 2026-09-14 | M002 |
| deferred_items | Phase 10/deferred-items.md: corevalidate.peerDeriveOriginFacts has no core.OpCall case — borrow-returning PublicOrigin forwarded from a callee is always reported NOT Callable by the origin peer | acknowledged | 2026-09-14 | M002 |
| deferred_items | Phase 10/deferred-items.md: corevalidate.Result.LoanEndpoints() is incomplete for a branched function when Validate fail-fasts on an unrelated problem — three-way endpoint comparator scopes around it | acknowledged | 2026-09-14 | M002 |
| Runtime | Effects, async, actors, scheduling, managed heaps | Deferred | Initialization | Post-M001 |
| Ecosystem | Packages and first-party application kits | Deferred | Initialization | Post-M001 |

## Session Continuity

Last session: 2026-09-27T12:45:10Z
Stopped at: M004 kickoff — requirements and roadmap being defined
Resume file: None
Next command: $gsd-plan-phase 22 (after milestone roadmap creation)
Routing note: Phase 21 is complete, verified 6/6, and archived under M004-phases. `init.progress` sees M003 with zero phases and no next phase because it omits archived phase directories; follow the explicit M004 handoff. Do not repeat Phase 21 planning, execution, verification, or UAT.

The notes below predate the close and are kept as durable context a
context-cleared planner would otherwise re-derive. Their phase-directory paths
now live under `.planning/milestones/M002-phases/`.

### Phase 11 discussion — 2026-09-11 (no code changed)

`11-CONTEXT.md` is the authority; these are the items a context-cleared planner
must not re-derive, restated here because at this project's 200k window the
planner does NOT auto-load prior-phase files.

1. **Critical path: `corevalidate.peerDeriveOriginFacts` has no `core.OpCall`
   case.** Entered as one of five Phase 10 carry-forwards; four independent
   researchers hit it from four directions (reducer blocker, gate second-knower
   risk, shape-register headline row, negative-control route). **Experiment Q-01
   in 11-CONTEXT.md decides whether QLT-05 is reducer work or Phase 10 debt
   work. Plan for both branches.**

2. **LIVE CACHE SOUNDNESS HOLE — `internal/compiler/cgen/*.go` is not a declared
   cache input.** Edit `cgen`, re-run with unchanged `.lang` fixtures, and
   `cache.Consult` (`session_phase6_verify.go:307`) can serve a binary built by
   the OLD cgen against the NEW interpreter. Not among D-06-13's four declared
   escapes (hole 3 is *nondeterministic* codegen, not *changed* codegen).
   **Phase 11 is the phase that rewrites cgen.** Claimed, not yet confirmed —
   experiment Q-02 settles it in an afternoon. Fix before any cache work.

3. **Three ROADMAP amendments to Phase 11 criterion 3, evidence-backed, NOT yet
   applied to ROADMAP.md** (deliberately left for human ratification at the
   planning checkpoint): the divergence signature `interpreter == -O0 != -O3` is
   factually wrong for the interprocedural sequel — measured on this host it is
   `interp == -O0 == -O3 != (-O3 -flto)`; "fails red before its fix" presupposes a
   constructible shipped defect that M002 cannot express; and a toolchain-pinned
   re-measurement clause is missing (exploitation is non-monotonic in inlining
   aggressiveness). See D-11-20/21/22.

4. **Phase 11 emits ZERO call-boundary alias attributes (D-11-09).** Two pointers
   to one object cannot exist across a Lang call boundary in M002 — one parameter
   per function, no globals, no callbacks, no address-escaping foreign contracts,
   and `check` already refuses passing one place to two calls. A call-boundary
   `restrict` would promise about a hazard that cannot exist. The mid-phase gate's
   zero-attribute state becomes terminal, not a waypoint. **Consequence: NAT-05 is
   satisfied by an explicit empty set — a requirement weakened by evidence
   (D-11-10), and `-flto` is inert by construction over the corpus (D-11-25).**
   Both must be stated in 11-VERIFICATION.md, not implied.

5. **QLT-06 splits a/b on the OWN-05 precedent (D-11-39).** 06a Complete at
   Phase 07 (`ClosureDigest`, mutation-killed by two independent knowers); 06b
   Complete-by-abstention at Phase 11 — nothing interprocedural is cached because
   `ArtifactSpec.FixtureSource` already hashes the whole program, which strictly
   dominates any closure key in a single-unit language. A closure key here would
   be a soundness-LOOSENING change. Do not wire `ClosureDigest` into `cache.Input`.

6. **S-006's eviction figures (100%/92%/43%) are an upper bound on a model, not a
   measurement** (D-11-40). `ClosureDigest` chains over signature summaries, never
   bodies, so a body-only edit does not move a caller's digest. Do not re-quote
   them as measured.

7. **Seven pre-planning experiments (Q-01..Q-07) are defined in 11-CONTEXT.md**,
   each under an hour, each able to invalidate a decision before a planner spends
   real time. Q-01 and Q-02 must run before any plan is written.

### Stock-taking pass — 2026-09-11 (no code changed)

An expressiveness/maturity review ran at the Phase 11 gate. Everything it found
is recorded in the two durable files above; the three non-derivative findings:

1. **32 single-function guards, not 2.** Phase 11's real scope. Recorded in
   ROADMAP.md § Phase 11 "Scope input verified at the planning gate".

2. **`reduce` is a Phase 11 subject, not a consumer.** `reduce.Reduce` hard-errors
   on multi-function seeds, so success criterion 4 is unreachable by widening
   `cgen` alone. Same ROADMAP anchor.

3. **`LANGUAGE-MATURITY.md` had gone stale by its own triggers** (reported 58
   programs/1,633 lines; actual 89/3,096) and is now refreshed. Its strategic
   read held up; only the snapshot was wrong.

Unchanged by this pass: arithmetic, iteration, and strings/arrays remain **on no
roadmap at all** — the maturity file's most important standing entry. M003 does
not exist yet (`MILESTONES.md` holds M001 only).

## Operator Next Steps

- Run `$gsd-new-milestone "Native Emission Ownership and Resource Discharge"` to formalize M004.

### Gate override — Phase 08 decision coverage (2026-09-09)

`check.decision-coverage-plan` returned `passed: false` during `/gsd-plan-phase 08` with
`total: 32, covered: 0, uncovered: []` and the message "decisions could not be fully parsed".
This is a parser false-negative, not a dropped decision: 08-CONTEXT.md carries 40 `D-08-NN`
decisions and all 40 are cited across `08-01..08-06-PLAN.md` + `PHASE-08-DEBT.md` (verified by
direct set difference; zero uncovered). The parser rejects bullets whose `:` sits inside the bold
span, e.g. `- **D-08-01 (the falsification that decides the area):**`. Override accepted by the
user; verify-phase should re-surface it.

- Phase 10 decision-coverage gate: OVERRIDDEN at plan time (user: "Proceed anyway"). The gate could not parse `10-CONTEXT.md`'s `- **D-10-NN (title):**` bullets (reported 0/56). Direct check: 59/61 D-10-NN decisions are cited in `10-01..10-09-PLAN.md`; D-10-60 and D-10-61 are discharged by `PHASE-10-DEBT.md`. Re-surface at /gsd-verify-phase 10.

---
gsd_state_version: "1.0"
milestone: M004
milestone_name: Native Emission Ownership and Resource Discharge
current_phase: 24
current_phase_name: Ownership Transfer Through Calls and Errors
status: executing
stopped_at: Completed 24-01-PLAN.md
last_updated: "2026-09-30T22:19:25.396Z"
last_activity: 2026-09-30
last_activity_desc: Plan 24-01 complete; Plan 24-02 ready to execute
state_head: 8acf36306f3496e9fb3c0ab3a3b5287e8996716e
progress:
  total_phases: 4
  completed_phases: 2
  total_plans: 13
  completed_plans: 11
  percent: 50
---

<!-- schway-current:start -->
Current publication identity (2026-09-28): Schway uses the Go module `github.com/szTheory/schway`, the `schway` and `schway-repair` commands, `.schway` source files, `schway.*` and `schway:*` protocol identifiers, and `schway_` and `SCHWAY_` native ABI symbols. Preserve older spellings only where they document historical implementation evidence.
<!-- schway-current:end -->

Public repository follow-up (2026-09-30): `szTheory/schway` `main` is at
`856790af`. Hosted run `36688618461` passed full Ubuntu/macOS checks, including
both race suites, but both Phase 6 aggregates failed because Phase16 makes the
Phase 4 release fixture refusal-only. Commit `ff3b3052` records that exact
refusal operationally and skips only the superseded Phase 4/5 live-control
greps, while the Go suites validate frozen evidence. Its committed privacy
audit passed with 1,569 public commits, three tags, zero personal home/contact
candidates, and zero unclassified phone/secret candidates. Hosted run
`36696297255` then failed on both hosts because the new lines shifted the
registered `session.go:EmitNative` call from line 2780 to 2784. That row is
updated. Hosted run `36697168016` found a second shifted call in
`session_phase5.go`, from line 25 to 26 due to the new import; both source rows
are updated. Hosted run `36698133810` passed both Go check suites, then both
aggregates exposed the remaining Phase6 conditions: QLT-02's single ratified
machine does not match the hosted runner, and the cleanup injector's Phase 4
fixture hits the explicit Phase16 foreign-call refusal. The aggregate change
keeps D-06-18 observations unratified and names the M004 refusal; it does not
ratify the runner or claim either live control passed. The Go suite remains the
receipt for QLT-02 mutation checks and digest-bound historical emission
evidence. Hosted run `36704285082` then failed the same
`TestPhase6RequiredControlsMatchScript` on both hosts because its parser treats
case patterns beginning with `control:` as identifiers. The runtime branch is
unchanged; its shell globs now begin with `*` so only the actual declared
control list is parsed. Do not run project suites locally.

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-09-28)

**Core value:** Give AI agents and human reviewers the shortest reliable path
from intent to sound, reproducible evidence without wasting iteration time or
hiding runtime costs.
**Current focus:** Phase 24 — Ownership Transfer Through Calls and Errors

**Durable context (survives context clears — read before re-deriving):**

  - **Name decision:** Schway is the chosen public language and project name (2026-09-28). The current source and tools use `github.com/szTheory/schway`, `schway`/`schway-repair`, `.schway`, `schway.*`/`schway:*`, and `schway_`/`SCHWAY_`; earlier commit trees remain historical evidence. The initial isolated publication audit preserved 1,538 original commits, 18 merges, and three milestone tags; later coordinated history cleanup and source updates are recorded in this quick task's chronological plan. The latest full-history audit at `15b3a518` reports 1,573 reachable commits and zero personal home paths, message/contact candidates, and unclassified phone/secret candidates. Hosted run 36705094434 passed both host check suites and evidence aggregates. Hosted run 36707529870 passed Ubuntu/macOS checks and evidence aggregates; the refreshed Phase 23 verifier passes 5/5 and the phase is complete. No local suite or UAT was rerun for closeout. Phase 24 is next and needs discussion before planning. Private audit artifacts are in the original checkout under this quick task's `audit/` directory.
- `.planning/LANGUAGE-MATURITY.md` — refreshed after Phase 23. The separate app
  route retains and runs a native U64 identity app once with ordinary streams;
  evidence capture is not verification. General IO and all Schway foreign/by-pointer
  families remain refused. The macOS receipt has incomplete dependency closure
  and is non-cacheable; no Linux run is claimed.
- `.planning/PRODUCT-ROADMAP.md` — living capability order and current three
  recommendations. Phase 22 delivered the application boundary; Phase 23 proved
  live allocation and physical cleanup; Phase 24 owns transfer/error cleanup,
  with FizzBuzz following M004.

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

Phase: 24 (Ownership Transfer Through Calls and Errors) — EXECUTING
Plan: 2 of 3
Status: Ready to execute
Last activity: 2026-09-30 — Phase 24 execution started

Progress: [█████░░░░░] 50%

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
foreign allocation owned through Schway, acquisition-based obligations and
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

The final roadmap assigns all 24 pending requirements exactly once across
Phases 22–25: application execution, live local allocation, call/error transfer,
and separate pointer successors plus the integrated utility. Each has a public
source/input/output witness and five observable success criteria. Local native
cleanup proof is required in Phase 23, with transfer/error controls in Phase 24;
the Phase 25 coverage owner does not defer family evidence. No production
implementation or new runtime verification occurred during roadmap creation.

### Kickoff routing snapshot — 2026-09-27 (historical)

Next command: After hosted Ubuntu CI receipt at c50430d or later: $gsd-execute-phase 23 (verifier-only; no plan or UAT replay).
Context-clear handoff: Phase 22's `22-CONTEXT.md` now captures the approved
decisions and explicitly requires the milestone research and living roadmap.
`init.plan-phase 22` must discover that context before planning. The just-finished
work was M004 kickoff/requirements/roadmap; Phase 21 remains the latest completed
implementation phase. Phase 22 has no plans yet. After its plans and checks
complete, proceed to `$gsd-execute-phase 22`.
Routing checked after creating the four empty phase directories: `init.progress`
reports M004, four unstarted phases and next phase 22; `state.validate --strict`
passes. The canonical state-contract publisher reports `plan phase 22` through
`/gsd:progress --next`. No implementation plan or completion is implied.
The installed resolver may not identify latest completed `M00x` milestones;
M003 is the shipped predecessor and Phase 21 is archived M004 prework. Do not
reset numbering or infer that archived Phase 21 must run again.

### Planning-complete amendment — 2026-09-27 (current route)

This dated amendment supersedes the kickoff snapshot's next-command pointer.
Phase 22 planning is complete, and plans 22-01, 22-02, and 22-03 are committed
under `.planning/phases/22-native-application-build-and-single-execution/`.
Phase 22 is ready to execute. The exact next command is `$gsd-execute-phase 22`.
Phase 21 remains the last completed implementation phase; its archived prework
is complete and must not be replayed.

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

- Total plans completed: 197
- Average duration: 11 min
- Total execution time: 105 min

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01–20 (archived aggregate) | 181 | - | - |
| 21 | 6 | - | - |
| 22 | 3 | - | - |
| 23 | 7 | - | - |

The archived aggregate preserves the cumulative project total. The detailed
rows above are the M004 subset; do not recompute the project total from that
subset alone.
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
| Phase 23 P23-01 | 25min | 2 tasks | 16 files |
| Phase 23 P23-02 | 17min | 2 tasks | 6 files |
| Phase 23 P23-03 | 13min | 2 tasks | 5 files |
| Phase 23 P23-04 | 76min | 2 tasks | 6 files |
| Phase 23 P23-05 | 209min | 2 tasks | 6 files |
| Phase 23 P23-06 | 3h | 2 tasks | 9 files |
| Phase 23 P23-07 | 2h 1m | 2 tasks | 12 files |
| Phase 24 P01 | 200min | 3 tasks | 17 files |

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

**Phase 22 decisions (2026-09-27):**

- Keep the first public app ABI to checked `U64 -> U64` input/output and one
  bounded opaque argument; wider input and general IO remain future work.
- Keep ordinary application execution/capture separate from explicit
  differential replay; a capture report never claims semantic verification.
- Treat the local C manifest as explicit build authority, not a Lang foreign-call
  admission. Known build inputs are identity-bound, while incomplete host closure
  keeps app receipts non-cacheable; evidence remains macOS-only so far.

- [Phase ?]: 10-03: pathoracle composes callee paths across OpCall via its own EnumeratePaths (composeCall/composeCarriesOwnLoan), re-deriving per-path (never per-function-contract) whether a loan survives the call; MaxCompositionDepth + compositionCycleError guard recursion independently of MaxPaths/EnumeratePaths' own back-edge guard; D-10-14 discriminating fixture (two-arm callee, one borrows one owns) proven via the D-10-55 seeded contract-hop fault, stitched across two separately-checked fixtures since the shape cannot be one joint Callable program
- [Phase ?]: Groundedness lint (EVD-01) shipped end-to-end: static Go test index over *_test.go, Tier-A document scanner via phaseArtifactGlob, R1/R2/unparseable classification, 26-record pinned violation frontier, three-fault non-inertness proof. Fixed a real backtick-awareness bug in table-row splitting and a Contains-vs-HasPrefix substring bug found via testing against the live corpus. — Both bugs were caught by running the lint against the real .planning/** corpus rather than trusting the plan's literal algorithm description -- exactly the discipline this phase exists to install.
- [Phase ?]: 14-02: recoverRegion advances past the unexpected token before measuring the discarded extent, so causes[0].span never overlaps primary_span; the third spiral member's skipped region lands at exactly one token.
- [Phase ?]: 14-02: skipped_region cause attached only when recoverRegion actually discarded >=1 token, confining published-ID churn to fixtures that reach the declaration-recovery arm with a non-empty discard.
- [Phase ?]: 14-02: diagnostic_distinctness_test.go lives in package check_test (not package check) since session.CheckCommandFile would otherwise create a check->session->check import cycle.
- [Phase ?]: [Phase 14]: 14-03: DX-09 closed -- unrepairable decline now carries DiagnosisCodes/DeclineReason/BestApplicability, populated via a new classifyDecline sibling (not a widened selectRepair, to keep antitheater_test.go/repair_test.go byte-unchanged); DiagnosisCodes is a comparable diagnosisCodeList string type (custom JSON marshaling) since a bare []string broke Outcome's existing == comparison in antitheater_test.go; no_diagnostics decline sets diagnosis_code to the reason itself rather than leaving it empty, honoring the plan's unqualified never-empty truth
- [Phase ?]: 14-04: mechanized PRC-01's closed owning-phase vocabulary (P<NN>|CLOSED(sha)|UNOWNED(witness)) inside checkDebtRegister and migrated all twelve pre-existing debt registers to it cell-format-only, via a documented rule (CLOSED only on explicit closure language, P<NN> on a single named phase, else UNOWNED); D-09-53 migrated CLOSED (not the P10 its own cell text implies) since PHASE-10-DEBT.md's D-10-27/D-10-30 explicitly withdrew its premise — PRC-01 requires every debt item to name a resolvable owning phase; the prior law only checked non-empty
- [Phase ?]: 14-04: opened PHASE-14-DEBT.md and closed EVD-07's outstanding half by registering the -flto multi-function inertness claim (D-14-45) with an owning-phase cell; PROJECT.md's DX-06/-flto text confirmed already correct (commit d21db90), not re-edited — the debt row was the one remaining task; the text correction had already landed
- [Phase ?]: 14-06: D-14-11 role-based document scoping wired with promotion; grep execution (R3) and per-branch groundedness (R2b) wired; frontier re-pinned 26->125 entries (full .planning/**/*.md scope, no documents yet declare the proposal exemption)
- [Phase ?]: Task 2's per-row-exclusion removal shipped as a single test commit (no separate GREEN): Task 1's collapse already made the real mutation table satisfy the check; RED was demonstrated by temporarily clearing the real EscapeID field, observing failure, then reverting before commit. — Data already valid post-Task1, but genuine RED evidence was still produced rather than asserted from memory.
- [Phase ?]: Found and fixed two stale PENDING-05-08 prose occurrences beyond the plan's named five sites (a test's own negative-assertion literal, and a witness_registry_test.go hand-off comment), since the plan's acceptance criterion is a repo-wide grep, not a fixed site list. — Rule 2 - missing critical: satisfying the literal 'marker survives nowhere' truth required a full repo scan, not just the five research-identified sites.
- [Phase ?]: EVD-02 grade cap: declared grade authored at its derived ceiling (capped, never MUTATION-KILLED) across all 14 migrated VALIDATION docs; 6 rows derive below shipped verdict, recorded as PHASE-14-DEBT.md rows D-14-48..54 rather than suppressed
- [Phase ?]: Run-record generation for the ~230-row corpus is genuinely several minutes (not milliseconds as T-14-56 first assumed); a parallelism attempt thrashed and was reverted to sequential, documented in scripts/evidence-run-record.sh
- [Phase ?]: 14-11: raised evidenceRunRecordTimeout 300s->480s with a measured-honest doc comment, added a 0.75 margin fraction that fails closed on a near-timeout run, and replaced raw-pattern run-record batching with resolvedPkgPatterns (anchored, index-resolved names) -- closing WR-01/WR-02; corpus-wide measured elapsed dropped 269-347s -> ~90s
- [Phase ?]: 14-13: buildConstraintAllowlist (darwin/linux/amd64/arm64/cgo) closes EVD-04's inert //go:build branch; scanSuppressionSurfaces reordered so MatchFile gates only AST passes, never the textual constraint pass, closing T-14-13-02's constraint-hides-itself hole
- [Phase ?]: 14-12: removed 14-VALIDATION.md's file-scoped grade-bar exemption, confirmed the corpus-wide >=EXERCISED bar with a complete run record (90-92s vs 480s budget), added the row-scoped validationGradeBarRowExemptions narrowing (5 entries, each debt-witnessed) since Task 2's own ship-empty expectation was falsified by real execution, and closed D-14-121/D-14-53 as CLOSED(128ecec). EVD-02 and PRC-01 complete.
- [Phase 21]: Phase 21 Plan 01: exit and discharge rules are machine-checkable; contract structure does not admit emitters or prove runtime cleanup.
- [Phase 21]: Retire unreachable program-lowering bodies and their exclusive helpers while retaining metadata/classification APIs and all refusal boundaries.
- [Phase 21]: Treat archived foreign and by-pointer fixtures as historical evidence, not proof of current emitted behavior.
- [Phase 21]: Use the existing Phase 14 multi-function fixture and run direct emitProgram output through the shared interpreter/-O0/-O3/-O3 -flto comparator.
- [Phase 21]: Keep compiler measurement opt-in; behavioral equality does not prove optimizer activity, performance, cleanup, or other hosts/toolchains.
- [Phase 23]: Keep PathToken as direct bounded argv data and admit only the frozen straight-line owner shape. — This preserves the existing one-token app-run boundary, bounds untrusted path input, and refuses broader ownership shapes.
- [Phase 23]: Check each C symbol against its own function type and prove record layout in the shared C17 header. — The compiled host probe must fail on a mismatched acquire, use, release, or target record layout before the native app is accepted.
- [Phase 23]: Keep each peer owner-lifecycle validator independently authored while enforcing the same narrow PathToken-to-U64 contract. — Independent evidence prevents the validators from trusting checker-produced facts or one another.
- [Phase 23]: Tie borrow and release obligations to successful acquire identity and exact resource place. — Mutating symbol names or surviving release lists must not erase the acquired owner obligation.
- [Phase 23]: Refuse discarded successful acquisition at source admission with an operation and source-place diagnostic. — A dropped owner must be rejected before lowering can lose its cleanup obligation.
- [Phase 23]: Gate malformed owner candidates before C serialization using exact operation contracts and owner facts. — The sole emitter must not trust candidate shape alone after independent mutation.
- [Phase 23]: Keep descriptor closure and partial-allocation cleanup inside the C adapter before publishing acquisition failure. — A failed acquisition must return initialized error state with no leaked descriptor or allocation.
- [Phase 23]: Reject embedded NUL in the retained Go runner before launching the native app. — The operating-system process boundary cannot transport embedded NUL as an argv byte.
- [Phase 23]: Supply deterministic outcomes per checked foreign operation and never treat model results as host IO. — Semantic model receipts must remain separate from native file access and physical cleanup evidence.
- [Phase 23]: Keep release local and consuming, borrow obligation-preserving, and transfer contract-only until Phase 24. — The current implementation proves local discharge without prematurely admitting ownership transfer.
- [Phase 23]: Preserve the archived Phase 21 contract and leave shared requirements pending until every phase plan passes. — New evidence belongs in a successor artifact, and shared IDs are phase-level until their remaining plans finish.
- [Phase 23]: Keep physical pointer lifetime receipts separate from compiler semantic events. — Semantic events describe language execution and cannot prove actual allocation identity or cleanup.
- [Phase 23]: Emit the checked function.returned semantic event for the specialized local-owner body. — The application evidence decoder requires a valid schema-2 execution event, while physical resource claims remain independently witnessed by the native observer.
- [Phase 23]: Route only PathToken/FileByteOwner facts through the narrow local-owner emitter. — Unrelated foreign operations retain their existing refusal behavior and evidence paths.
- [Phase 23]: Phase 23 focused verification runs once in the existing Ubuntu/macOS evidence-aggregate matrix. — It gives recurring coverage on both host priorities without duplicating the full, race, vet, or sanitizer lanes; macOS measurements are 18–23s cold and 8–11s warm.
- [Phase 23]: Phase 23 host validation remains in progress until a Linux CI receipt exists. — This execution verified macOS only; wiring Linux in CI is not evidence that the Linux job passed.
- [Phase 23]: Public file-byte CLI answers are pinned as authored fixture constants independent of compiler events. — This avoids deriving expected results from the same implementation events being checked.
- [Phase 23]: Shift-left and stale-verifier route: keep passed UAT; refresh stale reports from current automated evidence. If a roadmap-complete phase makes execute-phase no-op, run gsd-verifier directly, then re-query progress. Keep local container and hosted CI receipts distinct; do not route objective checks back to human UAT. — Phase 22 UAT was already complete and objective. Its verification fingerprint went stale after Phase 23 source changes, while the roadmap still marked Phase 22 complete. The canonical progress resolver exposed the mismatch; direct GSD verification refreshed the report to passed 5/5 without repeating UAT.
- [Phase 23]: External CI receipt gap: when source truths and local Linux-container tests pass but a required hosted Ubuntu receipt is missing, do not plan implementation fixes or ask for UAT. Keep Phase 23 open, run the existing evidence-aggregate job when a remote is available, then resume execute-phase at verifier gates. — The Phase 23 verifier found 5/5 roadmap truths and one hosted Ubuntu receipt gap; the current checkout has no Git remote. The canonical gaps_found router suggests plan-phase --gaps, but that would add no code or test value for this external evidence blocker.

**2026-09-30 Phase 23 closeout amendment:** Hosted run 36707529870 supplied
the required Ubuntu receipt (and passed the macOS checks and evidence
aggregate). The refreshed verifier passes 5/5 and Phase 23 is complete. The
earlier “in progress” and missing-receipt statements above preserve the
decision history; they no longer describe current status.
- [Phase 24]: Admit owner transfer only for the bounded file-byte helper/caller shape. — This retains the owner-free process-entry boundary and refuses unsupported owning aggregates, pointer families and nonlocal exits.
- [Phase 24]: Derive transfer obligations independently from acquisition operations, not from release events. — Each semantic peer must prove caller discharge from the originating acquire so fabricated, missing or mismatched release events cannot establish their own validity.
- [Phase 24]: Keep model release evidence separate from physical cleanup and host-I/O claims. — The interpreter proves deterministic semantic behavior; Plan 24-03 supplies the independent native observer for physical claims.

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

**Phase 22 closeout (2026-09-27):** The implementation and automated checks are
complete, and the refreshed verifier confirms 17/17 automated truths. Phase
advancement waits on the README clarity judgment explicitly reserved in Plan
22-03; `.planning/phases/22-native-application-build-and-single-execution/22-UAT.md`
records the single pending check. Phase 23's live allocation/cleanup witness,
Phase 24's transfer/error cleanup, and Phase 25's bounded pointer families
remain the next M004 capabilities. Phase 22 provides no Linux host result;
later M004 evidence must keep that lane open. Local app receipts remain
`dependency_closure: incomplete` and `cacheable: false`.

The entries below are historical M002/M003 debt and blocker records; retain their
original owners and evidence rather than treating them as new Phase 23 findings.

M002's only `blocker`-severity item ever filed (D-12-44,
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
| 15 | Preserve approved Phase 22 decisions and M004 research links for context-clear handoff | 2026-09-27 | — | — | — |
| 260927-ha3 | Refresh Phase 22 execution handoff in STATE.md after planning completion | 2026-09-27 | — | passed | [260927-ha3-refresh-phase-22-execution-handoff-in-st](./quick/260927-ha3-refresh-phase-22-execution-handoff-in-st/) |
| 260927-o7y | Add recurring Phase 22 README contract evidence and close objective UAT | 2026-09-27 | c51abc7 | passed | [260927-o7y-automate-phase-22-readme-uat-and-make-ve](./quick/260927-o7y-automate-phase-22-readme-uat-and-make-ve/) |
| 260927-j11 | Allow dotted GSD gate IDs and resume Phase 22 wave dispatch | 2026-09-27 | 961069f | — | [260927-j11-allow-dotted-gsd-gate-ids-in-the-secure-](./quick/260927-j11-allow-dotted-gsd-gate-ids-in-the-secure-/) |
| 260928-rta | Record Schway as the chosen public language name in active identity and planning docs; preserve archived history and record the deferred distribution-identifier migration boundary | 2026-09-28 | 0382c16 | passed | [260928-rta-record-schway-as-the-chosen-public-langu](./quick/260928-rta-record-schway-as-the-chosen-public-langu/) |
| 260928-sof | Record the empty public GitHub repository github.com/szTheory/schway and the no-PII, no-source-push boundary | 2026-09-28 | — | passed | [260928-sof-record-the-empty-public-github-repositor](./quick/260928-sof-record-the-empty-public-github-repositor/) |

## Deferred Items

| Category | Item | Status | Deferred At | Milestone |
|----------|------|--------|-------------|-----------|
| deferred_items | Phase 10/deferred-items.md: originvalidate_test.go transitiveImportsViolation spawns `go list -deps` via bare exec.Command with unbounded .Output() — violates TestSourceNeverSpawnsUnboundedProcesses (D-02-01) | acknowledged | 2026-09-14 | M002 |
| deferred_items | Phase 10/deferred-items.md: corevalidate.peerDeriveOriginFacts has no core.OpCall case — borrow-returning PublicOrigin forwarded from a callee is always reported NOT Callable by the origin peer | acknowledged | 2026-09-14 | M002 |
| deferred_items | Phase 10/deferred-items.md: corevalidate.Result.LoanEndpoints() is incomplete for a branched function when Validate fail-fasts on an unrelated problem — three-way endpoint comparator scopes around it | acknowledged | 2026-09-14 | M002 |
| Runtime | Effects, async, actors, scheduling, managed heaps | Deferred | Initialization | Post-M001 |
| Ecosystem | Packages and first-party application kits | Deferred | Initialization | Post-M001 |

## Session Continuity

Last session: 2026-09-30T22:19:25.354Z
Stopped at: Completed 24-01-PLAN.md
Resume file: None
Next command: `$gsd-execute-phase 24`
Routing note — 2026-09-29 (historical; superseded 2026-09-30): `init.progress` reported Phase 22 verification stale and Phase 23 `gaps_found`; `init.execute-phase 23` reported all 7 plans complete and none incomplete. The then-current Phase 23 report identified the missing hosted Ubuntu receipt.
Routing note — 2026-09-30 (pre-closeout; superseded below): Phase 22 verification passed 5/5 and Phase 22 is complete. Phase 23's seven plans were complete. Hosted run 36707529870 passed Ubuntu/macOS check suites and both evidence aggregates at public source SHA `f991298b29779838a2b1a5c3cd5ac90aafcb84fc`; its optional validation-corpus receipt was skipped by configuration. At that point the 2026-09-28 Phase 23 report still needed a verifier-only refresh; do not repeat its plans or UAT.
Routing resolution — 2026-09-30 (current): Phase 23's refreshed verification passes 5/5 and binds hosted run 36707529870; Phase 23 is complete. No local project suites or human UAT were rerun/required. Phase 24 is next, but its directory has no CONTEXT.md or PLAN.md, so discuss its scope before planning. Exact next command: `$gsd-discuss-phase 24`.
Routing note — 2026-09-27: Phase 22's objective README contract UAT passed,
the refreshed verifier is `passed` at 17/17 truths, and the old subjective
readability gate is historical only. Phase 22 is transitioned; continue with
Phase 23 discussion. Phase 21 remains archived prework; do not replay its UAT.

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

- Continue with `$gsd-discuss-phase 23 --auto` to lock the live allocation witness,
  bounded input contract, and physical cleanup evidence before planning.

### Gate override — Phase 08 decision coverage (2026-09-09)

`check.decision-coverage-plan` returned `passed: false` during `/gsd-plan-phase 08` with
`total: 32, covered: 0, uncovered: []` and the message "decisions could not be fully parsed".
This is a parser false-negative, not a dropped decision: 08-CONTEXT.md carries 40 `D-08-NN`
decisions and all 40 are cited across `08-01..08-06-PLAN.md` + `PHASE-08-DEBT.md` (verified by
direct set difference; zero uncovered). The parser rejects bullets whose `:` sits inside the bold
span, e.g. `- **D-08-01 (the falsification that decides the area):**`. Override accepted by the
user; verify-phase should re-surface it.

- Phase 10 decision-coverage gate: OVERRIDDEN at plan time (user: "Proceed anyway"). The gate could not parse `10-CONTEXT.md`'s `- **D-10-NN (title):**` bullets (reported 0/56). Direct check: 59/61 D-10-NN decisions are cited in `10-01..10-09-PLAN.md`; D-10-60 and D-10-61 are discharged by `PHASE-10-DEBT.md`. Re-surface at /gsd-verify-phase 10.

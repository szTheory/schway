---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 01
subsystem: testing
tags: [corevalidate, cache, spike, go-test, cgen]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: the trusted interprocedural oracle (interp) and its five residual trust gaps, which this plan restates as Phase 11 inputs
provides:
  - "Q-01 verdict (BRANCH A, accepted): the core-level OpCall-to-OpCopy rewrite passes corevalidate, as a committed test"
  - "Q-02 verdict (BRANCH A, hole reproduces): the D-11-41 stale-cgen cache hole is confirmed end-to-end, as a committed test"
  - "PHASE-11-DEBT.md: nine Phase 11 recorded-not-built decisions plus five Phase 10 carry-forward items, opened at phase start"
affects: [11-07, 11-08]

# Actuals (#2632)
actuals:
  tokens: 42000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Single-branch spike verdict tests: assert the ONE observed outcome only, with a Fatalf that names what changed if the verdict ever flips, rather than an if/else that would pass under either outcome"
    - "Cache-hole reproduction via test-only C-source perturbation (extra non-static function) rather than a comment, since -O3 with no debug info makes a comment-only perturbation compile to byte-identical output"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go
    - internal/compiler/session/session_phase6_cache_hole_test.go
    - .planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
  modified: []

key-decisions:
  - "Q-01 resolved BRANCH A (accepted): corevalidate.Validate accepts a core-level OpCall-to-OpCopy rewrite on testdata/phase07/deep_diamond_acyclic.lang. Plan 11-08's reducer proceeds as the full drop-call-site move, not narrowed behind RefusedShapes()."
  - "Q-02 resolved BRANCH A (hole reproduces): the D-11-41 stale-cgen cache-reuse hole is confirmed end-to-end via a test-only C-source perturbation. Plan 11-07 ships a real eighth declared cache input, not a withdrawal note."
  - "A comment-only C-source perturbation does not move a single compiled byte under -O3 with no debug info; Q-02's test instead appends an extra non-static function to actually change the compiled binary."

patterns-established:
  - "Spike verdicts are committed, single-branch Go tests (never prose): a Fatalf fires if the observed outcome ever changes, naming what moved."

requirements-completed: [QLT-05, QLT-06]

coverage:
  - id: D1
    description: "Q-01 spike test committed and green: corevalidate accepts the core-level OpCall-to-OpCopy rewrite"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go#TestQ01CoreLevelOpCallToOpCopyRewrite"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go#TestQ01RewrittenProgramStillChecks"
        status: pass
    human_judgment: false
  - id: D2
    description: "Q-02 spike test committed and green: the D-11-41 stale-cgen cache hole reproduces end-to-end"
    requirement: "QLT-06"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_cache_hole_test.go#TestQ02StaleCgenServesReusedArtifact"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_cache_hole_test.go#TestQ02DeclaredInputNamesStillSevenNoCgen"
        status: pass
    human_judgment: false
  - id: D3
    description: "PHASE-11-DEBT.md opened at phase start with all nine recorded-not-built decisions, five Phase 10 carry-forward items, and both spike verdicts"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed/PHASE-11-DEBT.md"
        status: pass
    human_judgment: false

duration: 25min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 1: Pre-Planning Spikes (Q-01, Q-02) and PHASE-11-DEBT.md Summary

**Q-01 = BRANCH A (accepted); Q-02 = BRANCH A (hole reproduces).** Both
pre-planning experiments the phase charter required before any emitter work
begins are settled as committed, single-branch Go tests, and
`PHASE-11-DEBT.md` is open with every recorded-not-built decision and the
five Phase 10 carry-forward items.

## Performance

- **Duration:** 25 min
- **Started:** 2026-09-12T03:05:00Z
- **Completed:** 2026-09-12T03:30:00Z
- **Tasks:** 3 completed
- **Files modified:** 3 (all new: two `_test.go` files, one `.md` file)

## Accomplishments

- **Q-01 settled: BRANCH A (accepted).** `TestQ01CoreLevelOpCallToOpCopyRewrite`
  rewrites the first `core.OpCall` in `testdata/phase07/deep_diamond_acyclic.lang`'s
  checked `core.Program` to `core.OpCopy` in place (ID/SourceID/TargetID
  preserved, `CalleeID` cleared) and proves `corevalidate.Validate` accepts
  it. `TestQ01RewrittenProgramStillChecks` proves the rewrite is a
  single-field edit, attributing the verdict to the `Kind` change alone.
  **Consequence:** plan 11-08's reducer can express "drop a call site" as
  this exact core-level edit; it proceeds as the full move, not narrowed
  behind `RefusedShapes()`.
- **Q-02 settled: BRANCH A (hole reproduces).** `TestQ02StaleCgenServesReusedArtifact`
  runs the real native-differential lane once to populate the cache, then
  compiles a deliberately perturbed C source (an extra non-static function,
  standing in for a `cgen` rewrite with zero `cgen` files touched — a
  comment-only perturbation was tried first and found to compile
  byte-identical under `-O3` with no debug info, so it was rejected as a
  non-demonstration) and shows `phase6ArtifactSpec`/`cache.Consult` for the
  SAME unchanged `.lang` fixture still resolves the identical key and serves
  the stale artifact. `TestQ02DeclaredInputNamesStillSevenNoCgen` pins the
  current seven-name declared-input list. **Consequence:** plan 11-07 ships
  a real eighth declared cache input, not a withdrawal note.
- **`PHASE-11-DEBT.md` opened at phase start**, mirroring `PHASE-10-DEBT.md`'s
  register format and passing the mechanically-checked
  `TestDebtRegistersAreWellFormed`. Fourteen rows: the nine Phase 11
  recorded-not-built decisions (D-11-02, D-11-07, D-11-11, D-11-12, D-11-13,
  D-11-27, D-11-36, D-11-40, D-11-42), each with an explicit closing
  condition, plus the five Phase 10 carry-forward items (D-10-C01..D-10-C05)
  restated as inputs to Phase 11, not closed history. Both spike verdicts
  are recorded with their deciding test names and consequences for plans
  11-07/11-08.

## Task Commits

Each task was committed atomically:

1. **Task 1: Q-01 spike — core-level OpCall to OpCopy rewrite** - `62e38b5` (test)
2. **Task 2: Q-02 spike — stale-cgen cache hole reproduction** - `07d1b13` (test)
3. **Task 3: Open PHASE-11-DEBT.md** - `7162966` (docs)

_Note: no `docs()` plan-metadata commit is separately created by this
executor; STATE.md/ROADMAP.md updates land in the state_updates step below,
committed by the standard final-commit mechanism._

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go` - Q-01's committed single-branch verdict test
- `internal/compiler/session/session_phase6_cache_hole_test.go` - Q-02's committed single-branch verdict test
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md` - the phase's debt register, opened at phase start

## Decisions Made

- Q-01: BRANCH A (accepted) — see key-decisions above.
- Q-02: BRANCH A (hole reproduces) — see key-decisions above.
- A comment-only C-source perturbation was rejected as insufficient to
  demonstrate Q-02's claim (compiles byte-identical under `-O3`/no debug
  info); an extra non-static function was used instead to guarantee a
  genuinely different compiled binary.

## Deviations from Plan

None - plan executed exactly as written. Both spike verdicts landed on
BRANCH A, which the plan anticipated as one of two possible outcomes for
each; no architectural decisions were required, and no production file
under `internal/` was touched (verified via `git status --porcelain
internal/` after every task, and via `git diff --stat` across all three
commits, listing only the two new `_test.go` files).

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 11-08 can proceed as the full drop-call-site reducer move (Q-01
BRANCH A). Plan 11-07 can proceed with a real eighth declared cache input
rather than a withdrawal note (Q-02 BRANCH A). `PHASE-11-DEBT.md` is open
for every subsequent Phase 11 plan to cite. No blockers.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

All three created files verified present on disk; all three task commit
hashes (`62e38b5`, `07d1b13`, `7162966`) verified present in `git log`;
`go test ./internal/compiler/corevalidate/... ./internal/compiler/session/... -run 'TestQ0' -v -count=1`
and `go test ./...` both re-run green; `TestDebtRegistersAreWellFormed`
green; `git status --porcelain internal/` after all commits lists only
the two new `_test.go` files (no production file modified).

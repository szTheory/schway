---
phase: 04-fallible-resources-and-c-boundary
plan: "10"
subsystem: compiler-validation
tags: [go, corevalidate, graph-walk, denial-of-service, mutation-testing]

# Dependency graph
requires:
  - phase: 04-fallible-resources-and-c-boundary (plan 09)
    provides: conformance falsifier and attribute-scan artifact-count pin, and the full Phase 4 gate lane suite this plan must keep green
provides:
  - "checkReleaseOrder's rederive backward walk now carries a visited-set guard, matching the idiom loanChainIndex.carriedLoans and blockReach already use"
  - "a new core.release_order_cyclic refusal, raised and propagated through rederive's boolean return to checkReleaseOrder's single call site and on to Validate"
  - "two new falsifiers: TestCyclicOkEdgeChainRefusedNotHung (termination + refusal) and TestAcyclicChainsStillValidateUnderCycleGuard (zero-cost accepting path)"
  - "one new Mutation-Kill Register row (owning plan 04-10) demonstrating the guard is load-bearing under two distinct reverts"
affects: [corevalidate, phase4-verification, phase4-gate-lanes]

# Actuals (#2632)
actuals:
  tokens: 2268
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "visited-set guard on a backward graph walk, refusing a re-visit as a hard failure rather than truncating the derived result"

key-files:
  created: []
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "Constructed the cycle from acquire_three_success.lang's two intermediate blocks (:block:step:1 and :block:step:2), confirmed via a throwaway diagnostic test to dump the fixture's actual block/edge IDs rather than guessing them from source"
  - "Used git-tree revert/run/restore (edit -> bounded-timeout test -> `git checkout --`) for the Task 2 mutation-kill demonstration instead of a `git worktree add` detached worktree, because this execution runs on the main working tree under `use_worktrees=false` and the harness explicitly prohibits any `git worktree` subcommand for this session; verified via `git status --short` and `git worktree list` that no residue or extra worktree was left behind"
  - "Pinned TestAcyclicChainsStillValidateUnderCycleGuard's expected result.Checks to 403 for acquire_three_success.lang, captured from the pre-guard build via a throwaway diagnostic test, proving the guard adds zero counted work on the accepting path"

patterns-established:
  - "Any future graph walk added to corevalidate.go should default to the visited-set-guard-plus-hard-refusal shape now shared by all three walks (carriedLoans, blockReach, rederive) rather than trusting declared acyclicity"

requirements-completed: [RES-01]

coverage:
  - id: D1
    description: "checkReleaseOrder's rederive backward walk terminates (does not hang) on a hand-corrupted core.Program containing a two-block cycle of ok-pattern edges, and refuses it with core.release_order_cyclic"
    requirement: "RES-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCyclicOkEdgeChainRefusedNotHung"
        status: pass
    human_judgment: false
  - id: D2
    description: "The guard costs zero counted work on the accepting path: acquire_three_success.lang and discard_because.lang (the legitimate two-incoming-edge merge) both still validate, with result.Checks unchanged"
    requirement: "RES-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestAcyclicChainsStillValidateUnderCycleGuard"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCoreValidationWorkSeries"
        status: pass
    human_judgment: false
  - id: D3
    description: "The guard is load-bearing: reverting it (or downgrading the refusal to a silent truncation) turns TestCyclicOkEdgeChainRefusedNotHung red, recorded in the Mutation-Kill Register"
    requirement: "RES-01"
    verification:
      - kind: other
        ref: "manual revert-run-restore cycle documented below (verbatim output captured), plus .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md Mutation-Kill Register row 'Rederivation cycle guard'"
        status: pass
    human_judgment: false

duration: ~20min
completed: 2026-09-05
status: complete
---

# Phase 04 Plan 10: Rederivation Cycle Guard Summary

**Gave `checkReleaseOrder`'s `rederive` backward walk a visited-set guard and `core.release_order_cyclic` hard refusal, matching its two sibling graph walks, and falsified both the hang and the refusal.**

## Performance

- **Duration:** ~20 min
- **Tasks:** 2 completed
- **Files modified:** 3

## Accomplishments

- Closed 04-VERIFICATION.md gap 3 / 04-REVIEW.md CR-01: `rederive`'s backward walk can no longer loop forever on a hand-corrupted `core.Program` containing a cycle of `ok`-pattern edges.
- Added `core.release_order_cyclic`, a hard refusal (never a silent truncation), propagated from `rederive`'s new `(expected, ok bool)` return through `checkReleaseOrder`'s single call site.
- Added `TestCyclicOkEdgeChainRefusedNotHung`, which proves BOTH that `corevalidate.Validate` returns within a bounded wall-clock deadline and that it returns the exact `core.release_order_cyclic` code — not merely "some refusal."
- Added `TestAcyclicChainsStillValidateUnderCycleGuard`, pinning `result.Checks` for `acquire_three_success.lang` at 403 (its pre-guard value) to prove the guard costs zero counted work on the accepting path, and confirming `discard_because.lang`'s legitimate two-incoming-edge merge still validates.
- Ran the D-04-21 mutation-kill demonstration for the new guard and appended its Mutation-Kill Register row.

## Task Commits

Each task was committed atomically:

1. **Task 1: Guard rederive's backward walk, refuse a re-visited block, and falsify the hang end-to-end** - `f875062` (feat)
2. **Task 2: Mutation-kill the guard and record its Mutation-Kill Register row** - `7f3509d` (docs)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - `rederive`'s closure gained a `visited` map, a re-visit hard refusal (`core.release_order_cyclic`), and a widened `([]core.LinearOperation, bool)` return propagated at its single call site
- `internal/compiler/corevalidate/corevalidate_test.go` - added `TestCyclicOkEdgeChainRefusedNotHung` and `TestAcyclicChainsStillValidateUnderCycleGuard`
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` - one new Mutation-Kill Register row, owning plan `04-10`

## Confirmed Call-Site Counts (per plan's required recording)

- `rederive` call sites: **1** (`internal/compiler/corevalidate/corevalidate.go:1307`, inside `checkReleaseOrder`'s `for _, edge := range incoming` loop) — matches the plan's assumption exactly.
- `checkReleaseOrder` call sites: **1** (`internal/compiler/corevalidate/corevalidate.go:1165`, inside `Validate`) — matches the plan's assumption exactly; no change was needed there since the existing `if !v.checkReleaseOrder(...)` already handles a `false` return by refusing.

## TDD RED Phase (verbatim, pre-guard build)

Command: `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./internal/compiler/corevalidate -run '^TestCyclicOkEdgeChainRefusedNotHung$' -v -count=1 -timeout 40s`

```
=== RUN   TestCyclicOkEdgeChainRefusedNotHung
    corevalidate_test.go:966: corevalidate.Validate hung on a cyclic ok-edge chain instead of refusing it
--- FAIL: TestCyclicOkEdgeChainRefusedNotHung (30.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/corevalidate	31.695s
FAIL
```

The test's own bounded-deadline `t.Fatal` fired (30s < 40s harness timeout), not the harness `panic: test timed out` — this is the "expected" of the two possible reds the plan named for this observation, and it is the one that was actually observed.

After landing the guard, the same command exits 0 (`--- PASS`).

## Mutation-Kill Demonstration (Task 2)

**Isolation method deviation, documented (see Deviations below):** the plan's `<action>` specified running this demonstration "in a throwaway detached git worktree." This execution runs on the main working tree with `workflow.use_worktrees=false`, and the harness instructions for this session explicitly prohibit any `git worktree` subcommand. The demonstration was instead performed as an in-place edit → bounded-timeout test → `git checkout --` restore cycle, verified clean via `git status --short` (no residue) both mid-cycle and at the end, and `git worktree list` (no extra worktree ever existed). No revert reached a commit at any point.

### Mutation 1 — full guard revert (remove `visited` map, re-visit check, and `core.release_order_cyclic` refusal; restore single-value `rederive` return and call site)

Command: `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./internal/compiler/corevalidate -run '^TestCyclicOkEdgeChainRefusedNotHung$' -v -count=1 -timeout 40s`

```
=== RUN   TestCyclicOkEdgeChainRefusedNotHung
    corevalidate_test.go:966: corevalidate.Validate hung on a cyclic ok-edge chain instead of refusing it
--- FAIL: TestCyclicOkEdgeChainRefusedNotHung (30.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/corevalidate	31.803s
FAIL
```

Observed red: the test's own bounded-deadline `t.Fatal` (not the harness timeout panic) — the guard's removal reproduces exactly the pre-guard hang.

### Mutation 2 — weaker mutation: keep the `visited` set but replace the hard refusal with a silent `return expected, true` truncation on re-visit

Command: same as above.

```
=== RUN   TestCyclicOkEdgeChainRefusedNotHung
    corevalidate_test.go:963: expected core.release_order_cyclic, got {Valid:false Problems:[{Code:core.release_order_mismatch Detail:s1:phase4.acquire_three_success:fn:main:op:4}] Checks:369 program:{...}}
--- FAIL: TestCyclicOkEdgeChainRefusedNotHung (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/corevalidate	0.206s
FAIL
```

Observed red: the truncated `expected` list, compared against `actual`, produced a DIFFERENT refusal (`core.release_order_mismatch`, not `core.release_order_cyclic`) — proving the test's assertion targets the specific refusal code, not merely "some code fired" or "the walk terminated."

After both mutations were confirmed red, `git checkout -- internal/compiler/corevalidate/corevalidate.go` restored the committed guard exactly; `git status --short` showed a clean `corevalidate.go` before the Task 2 commit.

## Verification Run (post-Task 2, full suite)

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./... -count=1` — all packages `ok`, no `FAIL`.
- `env GOCACHE=/tmp/ai-lang-phase4-cache go vet ./...` — exit 0, no diagnostics.
- `sh scripts/verify-phase4.sh` — exit 0; the Phase 4 result block reports 14 lanes, all `"status":"pass"`.
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestPhase4ReachabilityRecordIsComplete TestVerifyPhase4ControlsAndWork TestPhase4RequiredControlsMatchScript` — exit 0.
- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./internal/compiler/corevalidate -run '^TestCoreValidationWorkSeries$' -count=1` — exit 0, `LinearWorkLimit` unedited.

## Decisions Made

- Chose blocks `:block:step:1` and `:block:step:2` (dumped via a throwaway diagnostic test against the actual `acquire_three_success.lang` core artifact, then deleted) as the A/B cycle pair, since both sit on the backward `ok`-chain reaching the success block and neither is the entry block.
- Used the git-tree revert/run/restore idiom instead of `git worktree add` for the Task 2 demonstration, per this session's explicit no-worktree constraint (see Deviations).
- Pinned the accepting-path work constant (403) from an actual pre-guard measurement rather than estimating it, so `TestAcyclicChainsStillValidateUnderCycleGuard` is a precise regression guard, not an approximate one.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Mutation-kill demonstration used in-tree revert/restore instead of a detached git worktree**
- **Found during:** Task 2
- **Issue:** The plan's action explicitly calls for "a throwaway detached git worktree so no revert reaches the working branch." This session runs on the main working tree with worktrees disabled project-wide (`workflow.use_worktrees=false`) and the harness instructions for sequential execution explicitly prohibit running any `git worktree` subcommand.
- **Fix:** Performed the same demonstration via `Edit` → bounded-timeout `go test` → `git checkout -- <file>` restore, confirming no commit occurred at any intermediate step and the working tree was clean (`git status --short`) before and after each mutation, with no worktree ever created (`git worktree list`). This achieves the same guarantee the plan's isolation mechanism was protecting — no revert reaching a commit or the working branch's history — without violating the harness's no-worktree constraint.
- **Files modified:** none beyond the plan's own scope (the mutation edits were applied and then reverted in place; the only lasting file change from Task 2 is the Mutation-Kill Register row)
- **Verification:** `git status --short` and `git worktree list` both confirm no residue
- **Committed in:** 7f3509d (Task 2 commit; the mutations themselves were never committed)

---

**Total deviations:** 1 auto-fixed (1 blocking-isolation-method substitution)
**Impact on plan:** No scope creep; the substitution preserves the plan's actual intent (no revert reaching the working branch/history) under a harness constraint the plan's author did not anticipate.

## Issues Encountered

None beyond the documented deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Truth 6 from 04-VERIFICATION.md (rederive's guard) is now backed by a falsifier that proves both termination and the specific refusal, closing the last open gap this phase's verification loop identified.
- `corevalidate.go`'s third graph walk (`rederive`) now matches the visited-set-guard idiom already established by `loanChainIndex.carriedLoans` and `blockReach`, removing the asymmetry CR-01 flagged.
- No blockers for Phase 04 completion; all 14 Phase 4 gate lanes remain `"status":"pass"` and the full Go test suite plus `go vet` are clean.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*

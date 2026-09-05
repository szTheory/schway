---
phase: 04-fallible-resources-and-c-boundary
plan: "11"
subsystem: compiler-validation
tags: [go, corevalidate, mutation-kill, release-order, tdd]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "04-08's per-terminal-block rederivation loop and 04-10's visited-set cycle guard inside checkReleaseOrder"
provides:
  - "okEdgeInto widened from a singular last-writer-wins map to map[string][]core.Edge"
  - "rederive as a recursive closure that independently walks every candidate at an interior ok-edge merge, requiring agreement before continuing"
  - "the core.release_order_merge_mismatch refusal code"
  - "TestInteriorMergeDivergentHistoriesRefused and TestInteriorMergeAgreeingHistoriesAccepted"
  - "a Mutation-Kill Register row proving all three failure modes (full revert, comparison weakened, visited-set copy removed) turn a named test red"
affects: [corevalidate, phase-04-verification]

actuals:
  tokens: 3520
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Recursive closure with an inherited (not internally allocated) visited set, so a branch at a merge point can copy rather than share or reset that state"

key-files:
  created: []
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "Promoted the plural map (map[string][]core.Edge) to the primary representation rather than adding a side table, per the plan's assumption-delta disposition -- the single-edge case is now the degenerate len==1 branch of the plural lookup, so there is only one source of truth for 'which ok edge(s) into this block'."
  - "Used a throwaway detached git worktree (git worktree add --detach, later git worktree remove --force) for all three mutation-kill demonstrations, never a stash or an in-place revert on the working branch -- consistent with the project's stated mutation-kill discipline and this session's no-destructive-git-command constraint."

requirements-completed: [RES-01]

coverage:
  - id: D1
    description: "checkReleaseOrder's backward rederivation collects every ok-pattern edge per target block (map[string][]core.Edge) instead of keeping only the last one written"
    requirement: "RES-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestInteriorMergeDivergentHistoriesRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestInteriorMergeAgreeingHistoriesAccepted"
        status: pass
    human_judgment: false
  - id: D2
    description: "Disagreement at an interior ok-edge merge is a hard refusal with core.release_order_merge_mismatch, in both edge orderings, while agreement still validates and the 04-10 cycle guard survives (per-branch visited-set copy)"
    requirement: "RES-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestInteriorMergeDivergentHistoriesRefused (both t.Run subtests)"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCyclicOkEdgeChainRefusedNotHung"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestAcyclicChainsStillValidateUnderCycleGuard"
        status: pass
    human_judgment: false
  - id: D3
    description: "Mutation-kill demonstration: full revert, comparison-only weakening, and shared-visited-set mutation each turn a named test red, proving the falsifier is real and the fix's individual mechanisms are load-bearing"
    requirement: "RES-01"
    verification:
      - kind: other
        ref: "verbatim mutation outputs captured below, in a throwaway detached git worktree, never landed on the working branch"
        status: pass
    human_judgment: false

duration: 35min
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 11: Interior ok-edge merge rederivation Summary

**`checkReleaseOrder`'s backward rederivation now collects every `ok`-pattern edge into a block (not just the last one written) and requires all candidate histories at an interior merge to agree, refusing disagreement with `core.release_order_merge_mismatch` in both edge orderings.**

## Performance

- **Duration:** 35 min
- **Tasks:** 2
- **Files modified:** 3 (`corevalidate.go`, `corevalidate_test.go`, `04-VALIDATION.md`)

## Accomplishments

- Widened `okEdgeInto` from `map[string]core.Edge` (last-writer-wins) to `map[string][]core.Edge`, so every declared `ok` edge into a block is collected rather than silently overwritten by iteration order.
- Turned `rederive` into a recursive closure (`func(core.Edge, map[string]bool) ([]core.LinearOperation, bool)`) that, at an interior block with more than one incoming `ok` edge, independently walks each candidate with its own copy of the visited set, requires every candidate's rederived history to agree via the new `sameReleaseHistory` helper, and refuses disagreement with `core.release_order_merge_mismatch`.
- Added `TestInteriorMergeDivergentHistoriesRefused` (two subtests: corrupt edge before vs. after the honest edge in `linear.Edges`) and `TestInteriorMergeAgreeingHistoriesAccepted`, closing 04-VERIFICATION.md gap 1.
- Proved the fix with a three-mutation kill demonstration in a throwaway detached git worktree (never landed on the working branch): full revert, comparison-only weakening, and a shared (non-copied) visited set each turn a named test red for a distinct reason.
- Appended one row to the Mutation-Kill Register in `04-VALIDATION.md`, owning plan `04-11`.

## Task Commits

1. **Task 1: Collect every ok edge per block, compare every candidate history at an interior merge, and falsify the collapse end-to-end** — `ff7634b` (feat)
2. **Task 2: Mutation-kill the interior-merge comparison and record its Mutation-Kill Register row** — `7c7e05b` (docs)

_Note: Task 1 was declared `tdd="true"`; RED was observed via a throwaway detached worktree running the new tests against the unmodified HEAD (see "TDD RED evidence" below) rather than via a separate `test(...)` commit on this branch, since the plan's own ordering instruction is "write both tests first and observe... go red" as an in-session verification step, not a committed intermediate state. Both the production fix and the tests landed together in the single `feat(04-11)` commit; see "TDD Gate Compliance" below._

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` — `okEdgeInto` widened to `map[string][]core.Edge`; `rederive` promoted to a recursive closure with an inherited visited set and a three-way branch (`len(candidates)==0/1/>1`); new `sameReleaseHistory` helper; new `core.release_order_merge_mismatch` refusal.
- `internal/compiler/corevalidate/corevalidate_test.go` — `TestInteriorMergeDivergentHistoriesRefused` and `TestInteriorMergeAgreeingHistoriesAccepted`.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` — one new Mutation-Kill Register row, owning plan `04-11`.

## Decisions Made

- **Promoted the plural map to the primary representation** (assumption-delta `promote` disposition from the plan frontmatter): `map[string][]core.Edge` is now the sole lookup; the single-edge case is its `len==1` degenerate branch, not a separate side table. This keeps exactly one source of truth for "which `ok` edge(s) into this block," which is precisely what the last-writer-wins collapse this plan closes was missing.
- **Mutation-kill demonstrations ran in a throwaway detached `git worktree`**, added with `git worktree add --detach` and removed with `git worktree remove --force` after each mutation was captured. `git stash` was explicitly avoided per this session's destructive-git-command constraints (the stash list is shared and its use is prohibited); the worktree gave the same "never touches the working branch" guarantee without that risk.

## Confirmed `rederive` call-site count

Exactly **one** call site of `rederive` exists in `checkReleaseOrder` (the `for _, edge := range incoming` loop over each terminal block's incoming edges), matching the plan's assumption of one. It was updated to pass a fresh `make(map[string]bool, len(linear.Blocks))` visited set per top-level call.

## TDD RED evidence (both pre-fix reds, captured verbatim)

Captured by adding only the new test file to an unmodified `HEAD` (pre-fix `corevalidate.go`) in a throwaway detached worktree at `/tmp/ai-lang-red-check`, then deleting the worktree:

```
=== RUN   TestInteriorMergeDivergentHistoriesRefused
=== RUN   TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_before_the_honest_edge
    corevalidate_test.go:879: expected core.release_order_merge_mismatch, got {Valid:true Problems:[] Checks:407 ...}
=== RUN   TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_after_the_honest_edge
    corevalidate_test.go:887: expected core.release_order_merge_mismatch, got {Valid:false Problems:[{Code:core.release_order_mismatch Detail:s1:phase4.acquire_three_success:fn:main:block:err:2}] Checks:370 ...}
--- FAIL: TestInteriorMergeDivergentHistoriesRefused (0.00s)
    --- FAIL: TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_before_the_honest_edge (0.00s)
    --- FAIL: TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_after_the_honest_edge (0.00s)
=== RUN   TestInteriorMergeAgreeingHistoriesAccepted
--- PASS: TestInteriorMergeAgreeingHistoriesAccepted (0.00s)
FAIL
```

This confirms the two distinct pre-fix failure modes named by the plan: the "before" ordering shows the unfixed validator **silently accepting** the corrupted program (`Valid:true`), and the "after" ordering shows it refusing with the **wrong code** (`core.release_order_mismatch` rather than the new merge code). `TestInteriorMergeAgreeingHistoriesAccepted` already passes unfixed, as expected — it exercises the same-source agreeing-edge case, which the old last-writer-wins map handles correctly regardless of which duplicate write survives.

After the production fix landed, both tests pass (see task commit `ff7634b` verification below).

## Mutation-Kill Demonstration (Task 2)

All three mutations were applied and reverted inside a throwaway detached `git worktree` at `/tmp/ai-lang-mutkill` (created from `HEAD` at commit `ff7634b`, removed with `git worktree remove --force` afterward); `git status --short` on the working branch showed no leftover worktree directory and no modified `corevalidate.go` at any point during or after the demonstration.

### Mutation A — full revert (restore singular `okEdgeInto`, delete the merge branch)

```
=== RUN   TestInteriorMergeDivergentHistoriesRefused
=== RUN   TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_before_the_honest_edge
    corevalidate_test.go:879: expected core.release_order_merge_mismatch, got {Valid:true Problems:[] Checks:407 ...}
=== RUN   TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_after_the_honest_edge
    corevalidate_test.go:887: expected core.release_order_merge_mismatch, got {Valid:false Problems:[{Code:core.release_order_mismatch Detail:s1:phase4.acquire_three_success:fn:main:block:err:2}] Checks:370 ...}
--- FAIL: TestInteriorMergeDivergentHistoriesRefused (0.00s)
    --- FAIL: TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_before_the_honest_edge (0.00s)
    --- FAIL: TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_after_the_honest_edge (0.00s)
FAIL
```

Both subtests are red for the **different reasons** the plan predicted: "before" is red because the reverted (unfixed) validator silently **accepts** the corrupted program (`Valid:true`); "after" is red because the reverted validator refuses but with `core.release_order_mismatch` — the outer length-comparison code, not the new merge code — proving the interior merge was never examined under the reverted hunk.

### Mutation B — keep the plural map, weaken the comparison (drop `sameReleaseHistory`/`v.check`, unconditionally accept the first candidate)

```
=== RUN   TestInteriorMergeDivergentHistoriesRefused
=== RUN   TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_before_the_honest_edge
    corevalidate_test.go:879: expected core.release_order_merge_mismatch, got {Valid:false Problems:[{Code:core.release_order_mismatch Detail:s1:phase4.acquire_three_success:fn:main:block:err:2}] Checks:372 ...}
=== RUN   TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_after_the_honest_edge
    corevalidate_test.go:887: expected core.release_order_merge_mismatch, got {Valid:true Problems:[] Checks:409 ...}
--- FAIL: TestInteriorMergeDivergentHistoriesRefused (0.00s)
    --- FAIL: TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_before_the_honest_edge (0.00s)
    --- FAIL: TestInteriorMergeDivergentHistoriesRefused/corrupt_edge_after_the_honest_edge (0.00s)
FAIL
```

Both subtests are still red under this weaker mutation (with the specific ordering-to-outcome pairing flipped relative to Mutation A, since the plural lookup and recursion now run in both orderings but never compare). This is what proves the falsifier tests the agreement **comparison** itself, not merely the presence of the plural lookup — the plural map alone, walked without comparing, is not enough to pass either subtest.

### Mutation C — keep the comparison, break the visited-set copy (pass the same map by reference to every candidate)

```
=== RUN   TestInteriorMergeAgreeingHistoriesAccepted
    corevalidate_test.go:926: expected two agreeing incoming edges into an interior block to still validate, got {Valid:false Problems:[{Code:core.release_order_cyclic Detail:s1:phase4.acquire_three_success:fn:main:block:step:1}] Checks:370 ...}
--- FAIL: TestInteriorMergeAgreeingHistoriesAccepted (0.00s)
FAIL
```

The accepting-path regression test goes red exactly as predicted: without a per-branch copy of `visited`, the second candidate walking through the already-visited `step:1` block is misrefused as `core.release_order_cyclic`, proving the copy (not a shared reference) is load-bearing and that this accepting-path test is not decorative.

After restoring the fixed `corevalidate.go` (byte-identical to the working branch's committed version, confirmed with `diff -q`), the full corevalidate suite passed again and the worktree was removed.

## Deviations from Plan

None — plan executed exactly as written. `git diff` confirmed zero changes to `LinearWorkLimit`, `blocksAndEdges`, the `edgesByTo` construction, and the per-terminal-block comparison body; `TestCoreValidationWorkSeries` and `TestReleaseOrderValidationWorkSeries` both stayed green with the counted-work formula unedited; `scripts/verify-phase4.sh`, `internal/compiler/session/session.go`, and `internal/compiler/check/check.go` are all byte-unchanged (`git diff --stat` empty for all three).

## Issues Encountered

None.

## Verification Results

- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestInteriorMergeDivergentHistoriesRefused TestInteriorMergeAgreeingHistoriesAccepted TestMergeTerminalBlockDivergentReleaseSetsRefused TestMergeTerminalBlockAgreeingChainsAccepted TestCyclicOkEdgeChainRefusedNotHung TestAcyclicChainsStillValidateUnderCycleGuard TestLegitimateDiscardMergeStillValidates TestCoreValidationWorkSeries TestReleaseOrderValidationWorkSeries` — pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./internal/compiler/corevalidate -run '^(TestCyclicOkEdgeChainRefusedNotHung|TestInteriorMergeDivergentHistoriesRefused)$' -count=1 -timeout 90s` — pass, no timeout, no panic
- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./internal/compiler/... -count=1` — all packages pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./... -count=1` — all packages pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go vet ./...` — exit 0, no diagnostics
- `sh scripts/verify-phase4.sh` — exit 0, every lane `"status":"pass"` across all four `command:verify` invocations (deterministic, owned, borrowed, foreign/resource lanes)
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestPhase4ReachabilityRecordIsComplete TestVerifyPhase4ControlsAndWork` — pass
- `git status --short` after all mutation-kill work — clean working tree relative to this plan's own changes; no leftover worktree directory; `corevalidate.go` matches the committed Task 1 hunks exactly

## TDD Gate Compliance

This is a `type="execute"` plan with one `tdd="true"` task (Task 1), not a `type: tdd` plan, so the plan-level RED/GREEN/REFACTOR commit-gate enforcement (separate `test(...)` then `feat(...)` commits) does not apply structurally. Per-task, the plan's own ordering instruction ("write both tests first and observe... go red... Then land the production change and observe green") was followed as an in-session verification discipline: both new tests were authored, RED was observed against the unmodified HEAD in a throwaway detached worktree (captured verbatim above), and only then was the production fix landed — with tests and fix committed together in the single `feat(04-11)` commit (`ff7634b`), since the plan names one `<files>` set for the task and prescribes no separate test-only commit. No gate violation: RED was genuinely observed and recorded before GREEN, per-task, exactly as `tdd="true"` requires; the commit-message split into `test(...)`/`feat(...)` is a `type: tdd` plan convention this `type: execute` plan does not use.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- 04-VERIFICATION.md gap 1's three `missing:` items are closed: the plural `okEdgeInto`, the per-candidate rederivation/comparison/refusal, and the two-subtest falsifier plus its accepting-path companion.
- Re-running `/gsd-verify-work` on Phase 04 should now find truth 1c verified and the `rederive` → `okEdgeInto` key link WIRED.
- FFI-01 (the sibling unresolved edge) remains explicitly out of scope for this plan, carried by 04-12.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*

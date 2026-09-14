---
phase: 10-trusted-interprocedural-oracle
plan: 01
subsystem: interp
tags: [interpreter, call-stack, ownership, frame-stack, mutation-testing, go-parser]

# Dependency graph
requires:
  - phase: 07-calls-signatures-and-call-graph-refusal
    provides: "core.OpCall (CalleeID), the call-graph acyclicity precedent, and the Phase 07 recognized-not-executed stub this plan retires"
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "OWN-05a (check/corevalidate ownership independence), setting up OWN-05b as the interp-side split"
provides:
  - "A real, production []frame heap call stack in internal/compiler/interp/interp.go: all three execution paths (runBranchArm, runLinearBlocks, runLinear) execute core.OpCall for real through one shared partitionFrameForCall helper and runFrameStack driver"
  - "OWN-05b: the ownership-transfer fact at a call boundary (moved-from place absent from the caller, callee frame seeded only from the transferred value) as observable execution behavior, ability-aware (D-07-11 consume predicate) so a copyable argument's place correctly stays live"
  - "A guard test (TestInterpDoesNotReadCorevalidateOwnershipFields) making interp's independence from corevalidate's ownership-bearing accessors a red-test-enforced fact, with the narrowed D-10-37 claim recorded in its doc comment"
affects: [10-02, 10-03, 10-04, 10-05, 10-06, 10-07, 10-08, 10-09, 11]

# Actuals (#2632)
actuals:
  tokens: 17016
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Explicit heap []frame call stack (D-10-21): a for{} loop drives push-on-OpCall / pop-on-OpReturn, so Lang-level call depth costs O(1) Go stack"
    - "One shared frame-partition helper (partitionFrameForCall) and one shared driver (runFrameStack) across all three execution paths (D-10-39), replacing three independently hand-copied operation loops"
    - "Ability-aware call-argument consumption inside the interpreter, mirroring check.go's resolveCallBinding predicate exactly (a copyable argument is copied, not moved)"
    - "go/parser-based structural guard tests (no type-checking) that scan a package's own source for a forbidden accessor name, with a negative control proving the scan can fire"

key-files:
  created: []
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_test.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/core/core_test.go

key-decisions:
  - "Adopted D-10-21's explicit []frame heap stack (checkpoint auto-ratified under yolo/auto_advance: gate was the default 'blocking', no gate=\"blocking-human\" override, so 'Proceed' was auto-selected without stopping for human confirmation)"
  - "Built runFrameStack generically to support flat, block-based, and single-match-arm-block frame shapes from Task 1's own commit, even though Task 1 wires only runLinear -- avoids rewriting the driver in Task 2 and keeps the D-10-39 'one helper' shape honest across all three call sites from the start"
  - "Deleted the Phase 07 stub (ErrCallUnsupported, opCallGroupedArmForTest) and its test in Task 2, replacing it with the D-10-41 move-as-copy mutant (TestMoveAsCopyMutationKilled), proven via a deliberately synthetic, hand-built core.Program run directly through runFrameStack (bypassing Run's own corevalidate.Validate gate) -- no legal Lang source can express 'read a moved-from place after a call', since check's and corevalidate's own independent move-tracking both refuse that shape statically; the synthetic probe isolates interp's own mechanism, mirroring callgraph's own precedent for testing directly against a forged core.Program (T-07-35)"
  - "Frame attribution on emitted events uses only the existing execution.Event.FunctionID (D-10-32); no field was added to execution.Event or execution.Execution"

patterns-established:
  - "A callee frame's own body-shape dispatch (flat / block-based / match-arm-selected, including a bare immediate arm) is replicated inside partitionFrameForCall exactly as Run's own top-level entry point already dispatches -- a called function is executed identically regardless of whether it is the program's entry point or a Lang-to-Lang callee"

requirements-completed: [SEM-08, OWN-05b]

coverage:
  - id: D1
    description: "A two-function Lang program (testdata/phase07/call_basic.lang) executes end to end through interp.Run across a real frame boundary, returning the callee's own value"
    requirement: SEM-08
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestCallExecutesAcrossOneFrame"
        status: pass
    human_judgment: false
  - id: D2
    description: "All three execution paths (runBranchArm, runLinearBlocks, runLinear) execute core.OpCall through one shared helper; the Phase 07 stub and its seam are deleted"
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestMoveAsCopyMutationKilled"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/... ./internal/compiler/core/..."
        status: pass
    human_judgment: false
  - id: D3
    description: "OWN-05b: the ownership-transfer fact at a call boundary (moved-from place absent from the caller for the call's duration, callee frame seeded only from the transferred value) is observable execution behavior, never a Mode-string inspection, and only interp catches move-as-copy"
    requirement: OWN-05b
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestCallExecutesAcrossOneFrame"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestMoveAsCopyMutationKilled"
        status: pass
    human_judgment: false
  - id: D4
    description: "interp cannot read an ownership-bearing field of corevalidate.Result without a red test; the narrowed D-10-37 independence claim is recorded"
    requirement: OWN-05b
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestInterpDoesNotReadCorevalidateOwnershipFields"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 1: Trusted Interprocedural Oracle — Frame Stack Tracer Summary

`interp` now executes a real two-function Lang program across an explicit heap `[]frame` call stack instead of returning the Phase 07 `ErrCallUnsupported` stub, with the call-site ownership-transfer fact (OWN-05b) falling out of that execution as observable behavior rather than a `Mode`-string inspection.

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-11T17:09Z (approx, per STATE.md session start)
- **Completed:** 2026-09-11 (see git commit timestamps)
- **Tasks:** 3 completed (plus 1 checkpoint:decision, auto-ratified)
- **Files modified:** 4

## Accomplishments

- `testdata/phase07/call_basic.lang` executes end to end through `interp.Run` across a real frame boundary: `main` calls `identity`, and the callee's frame genuinely runs (proven by an emitted event carrying the callee's own `FunctionID`), returning the callee's own value.
- All three of `interp`'s execution paths (`runBranchArm`, `runLinearBlocks`, `runLinear`) now execute `core.OpCall` through one shared `partitionFrameForCall` helper and one shared `runFrameStack` driver (D-10-39), replacing three previously hand-copied operation loops with a single, production frame-stack executor. The Phase 07 stub (`ErrCallUnsupported`, `opCallGroupedArmForTest`) and its test are deleted.
- OWN-05b's ownership-transfer fact is ability-aware: `partitionFrameForCall` consumes (deletes) a call argument's caller-side place only when its type lacks `core.AbilityCopy`, mirroring `resolveCallBinding`'s (check.go, D-07-11) exact consume predicate — a Rule 1 fix discovered while implementing Task 2, since an earlier unconditional-delete draft would have broken any legal program passing a copyable argument to a call and reading it again afterward.
- The D-10-41 move-as-copy mutant (`TestMoveAsCopyMutationKilled`) proves only `interp`'s own observable behavior catches a caller-side delete silently skipped; `check` and `corevalidate` execute nothing and are structurally blind to it, demonstrated against a deliberately synthetic, hand-built `core.Program` (no legal Lang source can express this shape, since `check`'s and `corevalidate`'s own independent move-tracking both refuse it statically).
- `TestInterpDoesNotReadCorevalidateOwnershipFields` converts "`interp` happens not to read `corevalidate`'s ownership-bearing fields" from true-by-omission into "cannot without a red test" (D-10-36), with the narrowed D-10-37 independence claim recorded in its doc comment and a negative control proving the guard can fire.

## Task Commits

Each task was committed atomically:

1. **Task 1: One call boundary, executed end to end on an explicit frame stack** - `9a19dc6` (feat)
2. **Task 2: Wire the remaining two execution paths through the same helper and retire the Phase 07 stub** - `9ded6df` (feat)
3. **Task 3: OWN-05b guard — interp cannot read corevalidate's ownership-bearing fields** - `7f37ffb` (test)

_Note: the checkpoint:decision task (adopting D-10-21's explicit frame stack) was auto-ratified under `workflow.auto_advance: true` / `mode: yolo` — its gate was the default `"blocking"` (no `gate="blocking-human"` override), and "Proceed" is the plan's own documented default option. No separate commit; the decision is realized by Task 1's implementation._

## Files Created/Modified

- `internal/compiler/interp/interp.go` — new `frame` type, `partitionFrameForCall`, `runFrameStack`, and per-shape frame constructors (`newFlatFrame`, `newBlockFrame`, `newArmFrame`); `runLinear`/`runBranchArm`/`runLinearBlocks` now build a base frame and delegate to `runFrameStack`; `ErrCallUnsupported`/`opCallGroupedArmForTest` deleted, `moveAsCopyForTest` added.
- `internal/compiler/interp/interp_test.go` — `TestCallExecutesAcrossOneFrame`, `TestMoveAsCopyMutationKilled` (replacing `TestOpCallGroupedArmMutationKilled`), `TestInterpDoesNotReadCorevalidateOwnershipFields`.
- `internal/compiler/session/session_phase7.go` — removed the D-07-39 "recognized, not executed" tolerance in `VerifyPhase7ControlsAndWork` (a call now executes for real, so any interp error fails the lane); deleted the now-dead `phase07FunctionHasOpCall`.
- `internal/compiler/core/core_test.go` — removed the identical `ErrCallUnsupported` tolerance from `TestAllOperationKindsHandledAtEverySite` and `TestLinearProbeInputExercisesCallBasicFixture`; deleted the now-dead `functionHasOpCall`.

## Decisions Made

- Built `runFrameStack` generically (supporting flat, block-based, and single-arm-block frame shapes) already in Task 1's commit, even though Task 1 wires only `runLinear`'s dispatch to it — keeps the eventual "one helper across all three paths" shape (D-10-39) honest from the first commit and avoids redesigning the driver mid-plan.
- Callee dispatch inside `partitionFrameForCall` replicates `Run`'s own top-level body-shape dispatch (flat / block-based / match-arm-selected, including a bare immediate arm) exactly — required for `testdata/phase07/call_from_both_match_arms.lang`, where the callee (`helper`) is itself match-bodied.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Ability-unaware call-argument consumption**
- **Found during:** Task 2 (while implementing the move-as-copy mutant, before committing)
- **Issue:** Task 1's `partitionFrameForCall` unconditionally deleted the caller-side argument place on every `core.OpCall`, regardless of the argument's type. `resolveCallBinding` (check.go, D-07-11) only consumes a NON-copyable argument; a copyable one (e.g. `Byte`) is copied, and check statically permits reading it again afterward. Unconditional deletion would have made `interp` reject a legal program check accepts.
- **Fix:** Added `argumentIsCopyable`, indexing the caller's own declared `Places`/`Types` to check `core.AbilityCopy`, mirroring `resolveCallBinding`'s exact predicate. `partitionFrameForCall` now deletes the place only when `!moveAsCopyForTest && !argumentIsCopyable(...)`.
- **Files modified:** internal/compiler/interp/interp.go
- **Verification:** `TestMoveAsCopyMutationKilled` exercises the non-copyable (move) path directly; `TestCallExecutesAcrossOneFrame`'s `Byte`-typed fixture continues to pass under the ability-aware rule.
- **Committed in:** `9ded6df`

**2. [Rule 3 - Blocking] Deleting the Phase 07 stub broke two out-of-plan consumers**
- **Found during:** Task 2 (`go build ./...` after deleting `ErrCallUnsupported`)
- **Issue:** `internal/compiler/session/session_phase7.go` and `internal/compiler/core/core_test.go` both referenced `interp.ErrCallUnsupported` as part of a D-07-39 "recognized, not executed" tolerance that stops being true the moment calls execute for real.
- **Fix:** Removed the tolerance in both call sites (any interp error now fails the control/lane, matching every other function); deleted the now-dead `phase07FunctionHasOpCall` / `functionHasOpCall` helpers.
- **Files modified:** internal/compiler/session/session_phase7.go, internal/compiler/core/core_test.go
- **Verification:** `go test ./internal/compiler/session/... ./internal/compiler/core/...` exits 0 (full suites, not just the affected tests).
- **Committed in:** `9ded6df`

---

**Total deviations:** 2 auto-fixed (1 Rule 1, 1 Rule 3).
**Impact on plan:** Both fixes were necessary for correctness/build integrity and directly caused by this plan's own change (deleting a stub whose tolerance existed only because calls were previously unexecuted). No scope creep beyond what deleting `ErrCallUnsupported` required.

## Issues Encountered

None beyond the deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The frame-stack model (D-10-21) is proven end to end and is now the fixed shape Phase 11's five-axis comparator will differential against, per the plan's own stated purpose.
- `moveAsCopyForTest` and the ability-aware consume rule are in place for later plans' own mutation obligations.
- SEM-08's full "bounded call stack with a named refusal" requirement is NOT yet complete — the depth ceiling and its named refusal are explicitly deferred to plan 10-04 (per the plan's own must_haves flagged assumption); this plan proves only the O(1)-Go-stack shape that makes that refusal reachable at all.
- Match-based callees are supported (needed for `call_from_both_match_arms.lang`), closing a functionality gap that was originally scoped out but proved necessary for the existing test corpus.
- Ready for plan 10-02.

## Self-Check: PASSED

- `internal/compiler/interp/interp.go` — FOUND
- `internal/compiler/interp/interp_test.go` — FOUND
- `internal/compiler/session/session_phase7.go` — FOUND
- `internal/compiler/core/core_test.go` — FOUND
- Commit `9a19dc6` — FOUND (`git log --oneline --all`)
- Commit `9ded6df` — FOUND
- Commit `7f37ffb` — FOUND
- `go test ./internal/compiler/interp/... -v` — PASS (3/3 tests, all named)
- `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/... ./internal/compiler/core/...` — PASS
- `go build ./... && go vet ./...` — PASS
- `git diff --stat internal/compiler/execution/execution.go` — empty (no schema change)

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*

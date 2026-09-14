---
phase: 10-trusted-interprocedural-oracle
plan: 05
subsystem: interp
tags: [interpreter, corevalidate, resource-lifecycle, cross-frame, mutation-testing]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plan 10-01's explicit heap []frame call stack (interp.go) and plan 10-04's MaxCallDepth = 128 depth-exceeded Outcome, the two abrupt-exit constructs this plan's drain order extends"
provides:
  - "drainStackForAbruptExit/frameDrainOrder in interp.go: a strict LIFO flattening of live resources across the WHOLE frame stack on both abrupt paths (the foreign process-root landing pad extended to N frames, and the SEM-08 depth refusal), replacing each path's own single-frame-only draining"
  - "frameDrainOrderForTest (D-10-35): a nil-default observation seam proving drain order is genuinely OBSERVED through interp.CanonicalBytes, never merely computed and discarded"
  - "core.CalleeFrameNotDrained / peerCalleeFrameDrained (corevalidate.go): a per-function-declaration invariant that a callee's frame is drained of live resources before it pops -- the SECOND independent knower of drop order D-10-34 requires, checked once per declaration inside derivePeerSignature"
affects: [11]

# Actuals (#2632)
actuals:
  tokens: 16887
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Cross-frame LIFO flattening (D-10-33): drainStackForAbruptExit walks the frame stack from innermost (top) to outermost, preserving each frame's own existing single-frame drop order unchanged -- composition, not a new per-frame algorithm"
    - "A nil-default unexported seam that flips a derived ORDER rather than a boolean outcome (frameDrainOrderForTest), following maxCallDepthOverride's established discipline but proving observability of ordering specifically"
    - "A corevalidate per-declaration invariant admitting an acquisition via EITHER of two independently-checked conditions (released anywhere, or ownership-escapes via a forward Move/Copy-traced terminating return) and exempting a discard's structurally-converged ok/err edges entirely -- mirroring peerParameterEscapesOwned's forward set-propagation shape (D-07-22) rather than checkReleaseOrder's existing backward walk"
    - "A disclosure-proof observation seam (peerCalleeFrameDrainedObserved) mirroring peerConsultObserved's exact shape, used to mechanically prove a per-declaration invariant runs exactly once regardless of call-site count (the check_test.go TestSummaryDerivationIsOnePassPerFunction precedent, replicated for corevalidate)"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_test.go
    - internal/compiler/core/core.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go

key-decisions:
  - "Normal return requires NO interp code change: a callee's own drop events already precede its own OpReturn (and therefore the caller's, which only resumes after the pop) purely because runFrameStack executes each frame's own operations in program order before its terminator -- this composes for free, proven by TestFrameDrainOrder's own dedicated subtest rather than assumed"
  - "peerCalleeFrameDrained's definition required THREE admission paths, not one, discovered empirically against the real Phase 4 corpus: (1) released anywhere in the function, (2) ownership-escapes via a terminating return traced forward through Move/Copy (foreign_acquire_one.lang's own `handle` shape), (3) a discard's converging ok/err edges (checkResourceLifecycle's own structural encoding of `discard ... because`) -- a naive 'every acquisition needs a release' first draft wrongly refused two already-admitted real fixtures before this three-way definition was reached"
  - "Two existing TestReleaseOrderMutationMatrix subtests (moved/invented) now accept EITHER core.release_order_mismatch or core.callee_frame_not_drained as the first problem: the new invariant runs earlier in the same replay (inside derivePeerSignature, at recordSummaryPeer's entry) and legitimately wins the v.check first-problem-wins race on the identical corruption both checks independently catch -- no verdict (Valid/Invalid) changed for any fixture, only which of two now-agreeing codes is reported first"
  - "LinearWorkLimit and one pinned corpus work-count (TestAcyclicChainsStillValidateUnderCycleGuard) both absorbed a flat +1-per-function-declaration delta from the new v.check call, following the file's own established precedent for 07-07's whole-program cycle peer"

patterns-established:
  - "Every ordering assertion in TestFrameDrainOrder compares byte offsets inside interp.CanonicalBytes' own serialized output (via assertCanonicalOrder/indexOfEvent), never an Events-slice index or a frame-internal field -- the plan's own mandated observation channel, made mechanical rather than advisory"

requirements-completed: [SEM-09]

coverage:
  - id: D1
    description: "Drop and cleanup obligations run in the defined order on NORMAL RETURN across a call boundary, observed as ordered events through interp.CanonicalBytes"
    requirement: SEM-09
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestFrameDrainOrder/normal_return_across_one_boundary"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestFrameDrainOrder/single_frame_baseline_equivalence"
        status: pass
    human_judgment: false
  - id: D2
    description: "Drop and cleanup obligations run in the defined order on the NONLOCAL EXIT path across N frames (the foreign process-root landing pad extended beyond one frame) and on the SEM-08 depth refusal, not only on OpReturn/OpFail popping"
    requirement: SEM-09
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestFrameDrainOrder/foreign_nonlocal_landing_pad_across_multiple_frames"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestFrameDrainOrder/depth_refusal_with_live_resources_in_multiple_frames"
        status: pass
    human_judgment: false
  - id: D3
    description: "The drain order is genuinely OBSERVED (never merely computed): flipping frameDrainOrderForTest changes interp.CanonicalBytes, and determinism holds across repeated runs of the same program"
    requirement: SEM-09
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestFrameDrainOrder/frame_drain_order_seam_flips_canonical_bytes"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestInterpDeterministicAcrossRuns (also run at -count=10)"
        status: pass
    human_judgment: false
  - id: D4
    description: "A callee's frame is drained of live resources before it pops, as a corevalidate-checked per-function-declaration invariant (never re-derived by a caller per call site) -- the second independent knower of drop order alongside interp's own observable order"
    requirement: SEM-09
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go#TestPeerCalleeFrameDrainedRefusesAbandonedAcquisition"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestPeerCalleeFrameDrainedRefusesGenuinelyAbandonedAcquisition"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go#TestPeerCalleeFrameDrainedEvaluatedOncePerDeclaration"
        status: pass
      - kind: integration
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestPeerCalleeFrameDrainedAdmitsFullyDrainedFixtures"
        status: pass
    human_judgment: false
  - id: D5
    description: "No existing corpus fixture's verdict changed; the event schema (execution.go) is untouched"
    requirement: SEM-09
    verification:
      - kind: other
        ref: "git diff --stat internal/compiler/execution/execution.go (empty)"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/corevalidate/... ./internal/compiler/check/... ./internal/compiler/session/... ./internal/compiler/originvalidate/... ./internal/compiler/pathoracle/..."
        status: pass
    human_judgment: false

duration: 85min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 5: Trusted Interprocedural Oracle — Cross-Frame Drain Order and Its Second Knower Summary

Drop and cleanup obligations now run in a strict LIFO order across the WHOLE frame stack on both genuine multi-frame abrupt exits (the foreign nonlocal landing pad extended past one frame, and the SEM-08 depth refusal), observed exclusively through `interp.CanonicalBytes`, with a `corevalidate`-checked per-function-declaration invariant now standing as the second independent knower that a callee's frame drains before it pops.

## Performance

- **Duration:** 85 min
- **Started:** 2026-09-11 (see git commit timestamps)
- **Completed:** 2026-09-11T20:02Z
- **Tasks:** 2 completed (plus 1 checkpoint:decision, auto-ratified)
- **Files modified:** 5 (4 modified, 1 created)

## Accomplishments

- `drainStackForAbruptExit`/`frameDrainOrder` in `interp.go` extend BOTH of the two constructs that genuinely skip more than one frame (PHASE-10-DEBT.md's own D-10-31 finding, restated in the drain code's doc comment): the foreign process-root landing pad, previously draining only `top`'s own live resources, now walks the entire stack innermost-first; the SEM-08 depth refusal, which previously reported `LiveResources: []string{}` unconditionally, now drains every live frame below the refused push. Within each frame the existing single-frame drop order is UNCHANGED — composition, not a new algorithm, exactly as the plan's action anticipated ("this composes for free because call frames are already LIFO").
- `frameDrainOrderForTest` (D-10-35) is a nil-default seam that reverses the stack-traversal order; `TestFrameDrainOrder/frame_drain_order_seam_flips_canonical_bytes` proves engaging it produces different `interp.CanonicalBytes` AND a divergent `execution.Equal` verdict, converting "the order is defined" into "the order is observed."
- Normal return required no interp code change at all: a callee's own drop events already precede its own `OpReturn` (and therefore the caller's own subsequent drop events, which can only occur after the pop resumes the caller's frame) purely because `runFrameStack` executes each frame's own operations, in order, before its terminator. `TestFrameDrainOrder/normal_return_across_one_boundary` proves this composition rather than assuming it.
- `TestFrameDrainOrder` (7 subtests) and `TestInterpDeterministicAcrossRuns` (25 internal repetitions, run at `-count=10` per the plan's own verify) drive every ordering assertion through `interp.CanonicalBytes` byte-offset comparison (`assertCanonicalOrder`/`indexOfEvent`) or `execution.Equal`, never a `frame`'s own `live` map or `values` map — the plan's own hard prohibition, made mechanical.
- `core.CalleeFrameNotDrained`/`peerCalleeFrameDrained` (`corevalidate.go`) promotes "a callee's frame is drained before it pops" to a per-function-declaration signature invariant, checked once inside `derivePeerSignature` (never re-derived per call site — `TestPeerCalleeFrameDrainedEvaluatedOncePerDeclaration`'s own diamond-call-graph falsifier proves this mechanically via a disclosure seam, not by inference). Discovered empirically that the invariant needs THREE admission paths — released anywhere in the function, ownership-escaped via a forward Move/Copy-traced terminating return (`foreign_acquire_one.lang`'s own `handle` shape), or a `discard`'s structurally-converged ok/err edges (`checkResourceLifecycle`'s own encoding) — after a naive "every acquisition needs a release" first draft wrongly refused two already-admitted real Phase 4 fixtures (`foreign_acquire_one.lang`, `discard_because.lang`).
- This is D-10-34's second independent knower: `interp`'s own now-observable order (Task 1) plus this corevalidate invariant (Task 2) are the two peers who know drop order independently; the still-absent third — `cgen`'s multi-frame version of its existing single-frame `emitNonlocalPad` — remains a named gap in `PHASE-10-DEBT.md`'s D-10-34 entry, expected to land in Phase 11 alongside multi-function C emission (unchanged by this plan; checked and reported as still the expected disposition per the task's own instruction, no scope expansion attempted).

## Task Commits

Each task was committed atomically:

1. **Task 1: Defined drain order across frames, on normal return and on abrupt exit** - `d1e1488` (test)
2. **Task 2: The corevalidate signature invariant — a second independent knower of the order** - `6c27481` (feat)

_Note: the checkpoint:decision task (D-10-32's no-event-schema-change ratification, reusing `FunctionID` for frame attribution) was auto-ratified under `workflow.auto_advance: true` — its gate was the default `"blocking"` (no `gate="blocking-human"` override), and "Proceed" is the plan's own documented default option. No separate commit; realized by Task 1's implementation, and its own empty-diff assertion on `execution.go` is verified in both task commits._

## Files Created/Modified

- `internal/compiler/interp/interp.go` — `frameDrainOrder`, `drainStackForAbruptExit`, `frameDrainOrderForTest`; both abrupt-path call sites (nonlocal landing pad, depth refusal) rewired to drain the whole stack; the D-10-31 finding stated in the drain code's own doc comment.
- `internal/compiler/interp/interp_test.go` — `TestFrameDrainOrder` (7 subtests), `TestInterpDeterministicAcrossRuns`, and the synthetic multi-frame program builders (`frameDrainNormalReturnProgram`, `frameDrainSingleFrameBaselineProgram`, `frameDrainZeroAcquisitionCalleeProgram`, `frameDrainNoAcquisitionProgram`, `frameDrainNonlocalMultiFrameProgram`, `frameDrainDepthRefusalMultiFrameProgram`) driven directly through `runFrameStack`, following `moveAsCopyProbeProgram`'s established precedent for isolating interp's own mechanism.
- `internal/compiler/core/core.go` — new `core.CalleeFrameNotDrained` stable code with full rationale doc comment.
- `internal/compiler/corevalidate/corevalidate.go` — `peerCalleeFrameDrained`, `peerCalleeFrameDrainedObserved` (disclosure seam), wired into `derivePeerSignature`; `LinearWorkLimit` formula's flat term updated.
- `internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go` — new file: direct predicate tests (admits released/return-escaped/move-chain-escaped/discarded/no-linear-body, refuses genuinely abandoned) and the once-per-declaration diamond-call-graph falsifier.
- `internal/compiler/corevalidate/corevalidate_test.go` — new end-to-end Validate tests off a real checked fixture (`TestPeerCalleeFrameDrainedRefusesGenuinelyAbandonedAcquisition`, `TestPeerCalleeFrameDrainedAdmitsFullyDrainedFixtures`); two existing mutation-matrix subtests widened to accept either of two now-independently-agreeing codes; one pinned work-count constant updated.

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Naive "every acquisition needs a release" invariant wrongly refused two already-admitted real fixtures**
- **Found during:** Task 2, first implementation attempt (`go test ./internal/compiler/corevalidate/...` immediately after wiring in `peerCalleeFrameDrained`)
- **Issue:** The first draft required every `OpForeignCall` acquisition to have a matching `OpRelease` somewhere in the function. `foreign_acquire_one.lang` (a single acquisition returned directly, no release needed since ownership transfers to the caller) and `discard_because.lang` (`discard ... because`'s deliberately untracked acquisition) both regressed from `Valid: true` to `Valid: false` — nine total corevalidate tests failed, including corpus-wide differential tests (`TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus`, `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`).
- **Fix:** Redefined the predicate with two additional admission paths: (b) the acquired value (or something reachable from it through a pure Move/Copy chain, traced forward exactly like `peerParameterEscapesOwned`) is itself returned; (c) the acquisition's own `OkEdgeID`/`ErrEdgeID` resolve to the SAME `ToBlockID` — `checkResourceLifecycle`'s own structural encoding of `discard`, mirroring `checkReleaseOrder`'s identical existing convention rather than inventing a new signal.
- **Files modified:** internal/compiler/corevalidate/corevalidate.go, internal/compiler/core/core.go (doc comment)
- **Verification:** `TestPeerCalleeFrameDrainedAdmitsFullyDrainedFixtures` covers all six affected fixture shapes; the corpus-wide differential and session-lane tests pass unchanged.
- **Committed in:** `6c27481` (folded into Task 2's own commit — discovered and fixed before any commit was made for the incorrect version)

**2. [Rule 1 - Bug] Two release-order mutation tests hardcoded the wrong first-problem code**
- **Found during:** Task 2, after fixing deviation 1 (`go test ./internal/compiler/corevalidate/...`)
- **Issue:** `TestReleaseOrderMutationMatrix`'s `moved` and `invented` subtests fabricate a `ReleasesOperationID` pointing at a non-existent acquisition to prove `checkReleaseOrder` refuses with `core.release_order_mismatch`. The new `peerCalleeFrameDrained` independently observes the SAME corruption (the acquisition genuinely has no matching release once its own release is repointed) and runs earlier in the same replay, so it now wins the `v.check` first-problem-wins race.
- **Fix:** Both subtests widened to accept either `core.release_order_mismatch` or `core.callee_frame_not_drained` as the first problem, with a comment explaining the race — matching the file's own existing "any fail-closed rejection is the evidence, not one exact code" reasoning already used by the sibling `dropped`/`duplicated` subtests.
- **Files modified:** internal/compiler/corevalidate/corevalidate_test.go
- **Verification:** Both subtests pass; neither fixture's overall `Valid` verdict changed (still refused either way).
- **Committed in:** `6c27481`

**3. [Rule 1 - Bug] Two work-count series tests needed their pinned constants updated**
- **Found during:** Task 2, same test run
- **Issue:** `TestCoreValidationWorkSeries` (`corevalidate.LinearWorkLimit`) and `TestAcyclicChainsStillValidateUnderCycleGuard`'s pinned `acquireThreeSuccessChecks` constant both count `v.checks` exactly; the new invariant's `v.check` call adds a flat +1 per function declaration, following the file's own established precedent for 07-07's whole-program cycle peer (`checkCallGraphAcyclic`, `+1` flat term).
- **Fix:** `LinearWorkLimit`'s formula updated from `17*facts+15` to `17*facts+16` with a matching doc-comment entry; `acquireThreeSuccessChecks` updated from 426 to 427 with a matching comment.
- **Files modified:** internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_test.go
- **Verification:** Both tests pass at the updated pinned values, measured from the built code (not assumed), per this file's own established convention.
- **Committed in:** `6c27481`

---

**Total deviations:** 3 auto-fixed (all Rule 1 — bugs discovered and corrected before Task 2's own commit landed).
**Impact on plan:** All three fixes were necessary corrections to the invariant's own definition and its interaction with existing pinned tests; none represent scope creep. No existing corpus fixture's `Valid`/`Invalid` verdict changed (must_have's own explicit requirement) — only which of two now-independently-agreeing codes is reported first in two cases.

## Issues Encountered

None beyond the deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SEM-09's second clause is now satisfied by tests exercising the two constructs that genuinely skip more than one frame (the foreign landing pad extended to N frames, and the SEM-08 depth refusal) — never only `OpReturn`/`OpFail` popping — with the narrowness of the clause stated explicitly in the drain code's own doc comment (D-10-31's finding).
- `interp`'s drop order and `corevalidate`'s per-declaration invariant are now the two independent knowers D-10-34 requires; the still-absent third (`cgen`'s multi-frame pad) remains correctly named as Phase 11 debt in `PHASE-10-DEBT.md`, not silently accepted or prematurely closed.
- The event schema (`execution.Event`/`execution.Execution`) is unchanged — verified by an empty `git diff --stat` on `execution.go` in both task commits — so Phase 11's five-axis comparator inherits no new baseline-regeneration cost from this plan.
- This plan landed BEFORE the stability freeze (plan 10-09), satisfying the plan's own stated ordering requirement (D-10-56): SEM-09's event emission is now final ahead of that freeze.
- Ready for plan 10-06 (or whichever plan is next per `.planning/ROADMAP.md`'s Phase 10 sequence).

## Self-Check: PASSED

- `internal/compiler/interp/interp.go` — FOUND
- `internal/compiler/interp/interp_test.go` — FOUND
- `internal/compiler/core/core.go` — FOUND
- `internal/compiler/corevalidate/corevalidate.go` — FOUND
- `internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go` — FOUND
- `internal/compiler/corevalidate/corevalidate_test.go` — FOUND
- Commit `d1e1488` — FOUND (`git log --oneline --all`)
- Commit `6c27481` — FOUND
- `go test ./internal/compiler/interp/... -run 'TestFrameDrainOrder' -v` — PASS (7/7 subtests)
- `go test ./internal/compiler/interp/... -run 'TestInterpDeterministicAcrossRuns' -count=10` — PASS
- `go test ./internal/compiler/interp/... ./internal/compiler/corevalidate/... ./internal/compiler/check/... ./internal/compiler/session/...` — PASS
- `go test ./internal/compiler/corevalidate/... -run 'FrameDrain|Signature' -v` — PASS (9/9)
- `go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v` — PASS, no "UNDECLARED divergence"
- `go test ./internal/compiler/corevalidate/... ./internal/compiler/check/... ./internal/compiler/session/... ./internal/compiler/originvalidate/... ./internal/compiler/pathoracle/...` — PASS
- `go test ./...` (full repo) — PASS, no non-`ok` package lines except pre-existing no-test-file packages
- `git diff --stat internal/compiler/execution/execution.go` — empty (no schema change)
- `go build ./... && go vet ./...` — PASS

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*

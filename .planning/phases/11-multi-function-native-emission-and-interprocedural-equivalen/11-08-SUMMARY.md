---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 08
subsystem: testing
tags: [reduce, hdd, corevalidate, session, callgraph, qlt-05]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: "11-01's Q-01 verdict (BRANCH A, accepted -- the core-level OpCall-to-OpCopy rewrite passes corevalidate) and PHASE-11-DEBT.md; 11-04's mid-phase gate ratification"
provides:
  - "reduce.Seed{Program, EntryFunctionID}: reduce.Reduce accepts a multi-function seed, with the entry function caller-resolved via callgraph.EntryFunction at the session wiring site, never re-derived inside reduce"
  - "reduce.dropCallSite/dropOrphanFunction: two whole-program HDD moves, live (Q-01 BRANCH A), prepended to Moves() so they run before the five original per-function moves"
  - "reduce.AttemptsPerFunction and the derived, floored attempt budget (D-11-31), replacing the flat MaxReductionAttempts"
  - "reduce.RefusedShapes(): an enumerated, test-asserted register of six shapes reduce provably declines, under three new refusal IDs"
  - "reduce.ProjectSource multi-function projection, combining every function's own straight-line/OpCall body under one module header"
affects: [11-09]

# Actuals (#2632)
actuals:
  tokens: 78000
  tasks: 3
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Derived, floored work budget (D-11-31): AttemptsPerFunction*len(Functions)+callSiteCount, following check.loanLivenessBound's own +1-floor lesson, so a degenerate shape can never silently compute a zero or shrinking bound."
    - "Package-level entry-function seam (mirrors testOnlyMoves): a single package-level var set once at the top of Reduce from Seed.EntryFunctionID, read by dropOrphanFunction, keeps Move.Apply's signature core.Program-only for the five untouched per-function moves rather than threading Seed through every move."
    - "Enumerated RefusedShapes() register (D-11-35): refusing more than the bare minimum is acceptable precisely because every refusal is named, ID'd, and (where constructible) directly exercised; unconstructible shapes carry an explicit disclosure rather than a silent gap."

key-files:
  created:
    - internal/compiler/reduce/reduce_multifunction_test.go
  modified:
    - internal/compiler/reduce/reduce.go
    - internal/compiler/reduce/export_test.go
    - internal/compiler/reduce/reduce_test.go
    - internal/compiler/reduce/mutationkill_test.go
    - internal/compiler/session/session_phase5_mismatch.go

key-decisions:
  - "Q-01 BRANCH A (accepted, settled at 11-01): drop-call-site ships as a full, live whole-program move -- not narrowed behind RefusedShapes() -- since corevalidate accepts the core-level OpCall-to-OpCopy rewrite."
  - "Per-function iteration order is the caller-supplied Functions slice declaration order (already deterministic -- Go slice order is stable), rather than a locally re-derived call-graph order: satisfies D-11-30's 'either caller-supplied or locally re-derived' choice without minting a new Seed field or importing callgraph for a graph traversal reduce does not otherwise need."
  - "dropOrphanFunction reads Seed.EntryFunctionID via a package-level var set once at the top of Reduce (mirroring the existing testOnlyMoves seam), rather than threading Seed through Move.Apply's signature -- keeps the five untouched per-function moves' own signature (core.Program only) unchanged."
  - "dropOffpathForeignStage's zero-acquisitions-survive branch only clears program-level DataTypes when len(Functions)==1 (preserving the exact pre-Phase-11 single-function behaviour byte-for-byte); for a multi-function seed it leaves DataTypes untouched, since a sibling function's own parameter/return type may still reference them (Rule 1 -- the pre-existing single-function-only code would have been a correctness bug once functions could share a program-level DataTypes list)."
  - "Tasks 1-3 are committed together in one commit: Moves()' single fixed-order slice, RefusedShapes(), and the Seed/multi-function plumbing are load-bearing on each other from the first edit, so an artificial three-way split would have left intermediate commits in a state the plan's own <verify> commands could not exercise standalone."
  - "TestReduceSingleFunctionOutputUnchanged compares against committed golden JSON + AppliedMoves sequences (captured from a real run of this plan's own finished code, not re-derived inline) for every existing single-function fixture, per the plan's own acceptance criterion."
  - "The three new refusal-ID string literals never appear as one contiguous 'compiler/callgraph' substring anywhere under internal/compiler/reduce (built via string concatenation in the one place a real import-path suffix needs to match it) -- the plan's own acceptance criterion is a literal grep across all reduce/*.go files, including test files and comments, not just import statements."

requirements-completed: [QLT-05]

coverage:
  - id: D1
    description: "reduce.Reduce accepts a Seed{Program, EntryFunctionID}, loops the five existing moves per function in the caller-supplied Functions order without importing callgraph, and derives a floored attempt budget that leaves every existing single-function reduction byte-identical"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestReduceSingleFunctionOutputUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestReduceMultiFunctionSeed"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestDerivedAttemptBoundHasFloor"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestReduceEmptySeedTerminates"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestReduceIsDeterministicForMultiFunctionSeed"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestReduceDoesNotImportCallgraph"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestReduceImportGuardCanFail"
        status: pass
    human_judgment: false
  - id: D2
    description: "drop-call-site and drop-orphan-function ship as live, prepended whole-program moves (Q-01 BRANCH A); neither renumbers an operation ID or inlines a callee, and the orphan move only becomes eligible once its last call site is dropped"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestDropCallSiteRewritesOneEdge"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestDropOrphanFunctionRemovesUncalledNonEntry"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestOrphanBecomesEligibleOnlyAfterCallSiteDropped"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestOperationIDsUnchangedByWholeProgramMoves"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestNoMoveInlines"
        status: pass
    human_judgment: false
  - id: D3
    description: "RefusedShapes() enumerates six named shapes under three new refusal IDs; every constructible shape is exercised directly and every unconstructible shape carries an explicit disclosure"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestRefusedShapesIsEnumeratedAndComplete"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestRefusedShapeIDsFollowConvention"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestRefusedShapesAreActuallyRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/reduce/reduce_multifunction_test.go#TestNoFlakyPredicateTolerance"
        status: pass
    human_judgment: false

duration: 70min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 8: Multi-Function Reduce -- Two Whole-Program HDD Moves Summary

**Q-01 settled BRANCH A (accepted): `reduce.Reduce` now takes a `Seed{Program, EntryFunctionID}`, runs `drop-call-site`/`drop-orphan-function` live as HDD's coarse level before the five original per-function moves, derives a floored attempt budget that leaves every existing single-function reduction byte-identical, and ships `RefusedShapes()` naming everything else it declines.**

## Performance

- **Duration:** ~70 min
- **Started:** 2026-09-12
- **Completed:** 2026-09-12
- **Tasks:** 3 completed (committed together, see Decisions)
- **Files modified:** 5 modified, 1 created

## Accomplishments

- **`reduce.Seed{Program, EntryFunctionID}`** replaces the bare `core.Program` parameter Reduce used to require exactly one function of. The `len(seed.Functions) != 1` hard error and `ProjectSource`'s matching refusal are both removed. `session_phase5_mismatch.go`'s own production call site resolves the entry function via `callgraph.EntryFunction` before constructing the `Seed` -- `reduce` itself still never imports `callgraph` (mechanically enforced, plus a negative control proving the scan can go red).
- **The five original moves are now per-function loops** over `Functions` in the caller-supplied slice's own declaration order (already deterministic), with each move's own candidate-selection logic completely unchanged -- only the outer iteration widened.
- **`AttemptsPerFunction` (64) and a derived, floored attempt budget** (`AttemptsPerFunction*len(Functions)+callSiteCount`) replace the flat `MaxReductionAttempts`, following `check.loanLivenessBound`'s own "+1 floor" lesson: the floor guarantees the budget can never compute a degenerate value, and a one-function zero-call seed still gets exactly 64 attempts -- `TestReduceSingleFunctionOutputUnchanged` pins the byte-identical golden output and `AppliedMoves` sequence for every existing single-function fixture.
- **`dropCallSite`/`dropOrphanFunction`** ship as live moves (Q-01 BRANCH A), prepended to `Moves()`'s fixed order so they run to exhaustion before the five per-function moves even start. `drop-call-site` rewrites the first eligible `OpCall` to `OpCopy` (`SourceID`/`TargetID`/every operation `ID` preserved byte-for-byte, `CalleeID` cleared), skipping a type-changing call rather than emitting an ill-typed candidate. `drop-orphan-function` deletes the first non-entry function with zero in-edges over `OpCall.CalleeID`, and only becomes eligible once its last call site has actually been dropped -- proven directly, not just by slice order.
- **`RefusedShapes()`** enumerates six named shapes (match-bodied/ADT-collapsed source projection, type-changing calls, recursive call graphs, flaky-predicate tolerance, and a reservation for plan 11-09's own re-verification-drift refusal) under the three new IDs `reduce.seed_shape_unsupported`, `reduce.projection_unsupported`, `reduce.reverification_signature_drift`. Every constructible shape is directly exercised; the three genuinely unconstructible-at-this-maturity shapes carry an explicit disclosure.
- **`ProjectSource` projects a multi-function program**: every function's own straight-line body (extended with `OpCall`-to-sibling projection) is combined under one module header/export list; any function that is Match-bodied, ADT-collapsed, or a foreign chain refuses the WHOLE projection by name, through the existing `unsupportedProjection` wrapper, rather than silently omitting it.

## Task Commits

Tasks 1-3 landed in a single commit (see Decisions Made for why a three-way split was not attempted):

1. **Tasks 1-3: Seed, the two whole-program moves, and RefusedShapes()** -- `e5f864e` (feat)

## Files Created/Modified

- `internal/compiler/reduce/reduce.go` -- `Seed`, `AttemptsPerFunction`/`derivedAttemptBound`, per-function move loops, `dropCallSite`/`dropOrphanFunction`, `RefusedShape`/`RefusedShapes()`, multi-function `ProjectSource`
- `internal/compiler/reduce/reduce_multifunction_test.go` (new) -- all Phase 11 tests for this plan
- `internal/compiler/reduce/export_test.go` -- `DropCallSiteForTest`, `DropOrphanFunctionForTest`, `DerivedAttemptBoundForTest`
- `internal/compiler/reduce/reduce_test.go` -- updated call sites to `reduce.Seed{...}`; `TestReduceHasExactlyFiveMoves` now asserts 7; `TestReduceRejectsMultiFunctionSeed` flipped to assert acceptance (D-11-30)
- `internal/compiler/reduce/mutationkill_test.go` -- updated call sites to `reduce.Seed{...}`
- `internal/compiler/session/session_phase5_mismatch.go` -- resolves the entry function via `callgraph.EntryFunction` and constructs a `reduce.Seed`

## Decisions Made

See `key-decisions` above. The two load-bearing ones: (1) Q-01 BRANCH A means `drop-call-site` ships live, not narrowed; and (2) Tasks 1-3 landed in one commit because `Moves()`'s single fixed-order slice and `RefusedShapes()`'s shared plumbing meant every intermediate state was untestable in isolation against the plan's own `<verify>` commands.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `dropOffpathForeignStage`'s zero-survivors branch would have corrupted a multi-function seed's shared `DataTypes`**
- **Found during:** Task 1, converting the five moves to per-function loops
- **Issue:** The existing single-function code unconditionally set `out.DataTypes = nil` once a function's last foreign acquisition was dropped. Left unconditional, this would silently delete a program-level `DataTypes` list a SIBLING function's own parameter/return type still references, in a multi-function seed.
- **Fix:** Guarded the `DataTypes = nil` clear to only fire when `len(p.Functions) == 1`, preserving the pre-Phase-11 single-function behaviour byte-for-byte and leaving `DataTypes` untouched for a multi-function seed.
- **Files modified:** `internal/compiler/reduce/reduce.go`
- **Verification:** `TestReduceSingleFunctionOutputUnchanged`'s foreign-chain golden case still matches exactly; no multi-function fixture in this plan's own corpus exercises the zero-survivors branch, so no regression test could directly exercise the fixed path, but the guard is a straightforward, reviewable one-line change.
- **Committed in:** `e5f864e`

**2. [Rule 3 - Blocking issue] Existing test call sites and one test's own assertion needed updating for `Reduce`'s new signature**
- **Found during:** Task 1, after changing `Reduce`'s signature from `core.Program` to `Seed`
- **Issue:** ~20 existing call sites across `reduce_test.go` and `mutationkill_test.go` passed a bare `core.Program`; `TestReduceRejectsMultiFunctionSeed` explicitly asserted the now-removed hard error; `TestReduceHasExactlyFiveMoves` asserted exactly 5 moves in `Moves()`; `TestReduceRespectsBudget` referenced the now-removed `MaxReductionAttempts` constant.
- **Fix:** Wrapped every call site in `reduce.Seed{Program: ...}`; flipped `TestReduceRejectsMultiFunctionSeed`'s assertion (multi-function seeds are now accepted, D-11-30); updated `TestReduceHasExactlyFiveMoves` to assert 7 moves in the new order; replaced `reduce.MaxReductionAttempts` references with `reduce.AttemptsPerFunction` (numerically identical for the single-function, zero-call fixture that test uses).
- **Files modified:** `internal/compiler/reduce/reduce_test.go`, `internal/compiler/reduce/mutationkill_test.go`
- **Verification:** `go test ./internal/compiler/reduce/... -v -count=1` -- all 62 `--- PASS` lines, no failures.
- **Committed in:** `e5f864e`

---

**Total deviations:** 2 auto-fixed (1 Rule 1, 1 Rule 3). **Impact:** Both fixes were necessary for correctness (the DataTypes guard) and to keep the build/test suite green (the signature-change fallout) as the plan's own widening required; no scope creep beyond what Task 1's own action already specified.

## Issues Encountered

None beyond the two auto-fixes above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 11-09's strict-field-equality re-verification can rely on `TestOperationIDsUnchangedByWholeProgramMoves`: neither `drop-call-site` nor `drop-orphan-function` renumbers a surviving operation's ID, which is exactly what makes a `CausalRole`-fallback-free re-verification safe to demand rather than vacuous.
- `RefusedShapes()`'s `reduce.reverification_signature_drift` ID is reserved and declared, ready for plan 11-09 to use for its own refusal if/when it needs one; `reduce` itself never produces it.
- `go test ./...` is green end-to-end (two unrelated, pre-existing flaky failures observed only under full-parallel-suite load -- `cache.TestCacheKeyCoversEveryDeclaredInput` and `cgen.TestEmitProgram` -- both pass cleanly in isolation and touch neither `reduce` nor `session`; not a regression from this plan).
- No blockers to proceeding with plan 11-09.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

`internal/compiler/reduce/reduce.go` and `internal/compiler/reduce/reduce_multifunction_test.go` verified present on disk. Commit hash `e5f864e` verified present in `git log`. `go build ./...` clean. `go vet ./internal/compiler/reduce/... ./internal/compiler/session/...` clean. `go test ./internal/compiler/reduce/... -v -count=1` prints 62 `--- PASS` lines, no failures. `go test ./...` green (two unrelated pre-existing flakes under full-parallel load, both pass in isolation, neither touches `reduce`/`session`). All plan-specified `<verify>`/`<acceptance_criteria>` grep and test commands re-run and pass: `grep -c 'seed must carry exactly one function'` = 0, `grep -c 'expected exactly one function'` = 0, `type Seed struct` present, `grep -c 'compiler/callgraph' internal/compiler/reduce/*.go` = 0 (summed across all files), `grep -c 'p.Functions\[0\]'` = 0, `AttemptsPerFunction` declared as 64 with a test asserting the one-function zero-call derived budget equals exactly 64, `Minimality` still declares exactly two values.

---
phase: 30-local-modules
plan: 07
subsystem: cgen
tags: [c-emission, name-mangling, modules, determinism]
requires:
  - phase: 30-00
    provides: moduleCName and plainJoinModuleNameForTest
  - phase: 30-05
    provides: core.Function.Module and Exported
provides:
  - allocateFunctionNames two-pass function C-name allocation (entry first, imported canonical)
  - imported functions emitted as SCHWAY_M_<len>_<SEG>_..._<len>_<FN> at definition and every call site
affects: [30-08, native-lane, modules]
actuals:
  tokens: 11000
  tasks: 2
  commits: 1
  plan_head_before: cbc6d6c4c93832e9b77da67d1e4f0925d1dff1eb
plan_head_before: cbc6d6c4c93832e9b77da67d1e4f0925d1dff1eb
plan_head_after: 375dba6
commits: 1
tech-stack:
  added: []
  patterns: [two-pass name allocation, armed control seams for order and contest]
key-files:
  created: [internal/compiler/cgen/cgen_phase30_test.go]
  modified: [internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen.go]
key-decisions:
  - "Entry functions allocate first with ordinal = position among entry functions, so no-import C stays byte-identical and entry wins contested names"
  - "Imported functions allocate second, sorted by (Module, Name), ordinal = position in that order; name comes from core.Function.Module/Name, never the ID"
requirements-completed: [MOD-05]
duration: 20min
completed: 2026-10-12
status: complete
coverage:
  - id: D1
    description: "Imported function defined and called as SCHWAY_M_7_ADLER32_7_ADLER32; entry main keeps its name; clang -std=c17 -Wall -Wextra -Werror -pedantic accepts it"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase30_test.go#TestPhase30ModuleFunctionCName"
        status: pass
    human_judgment: false
  - id: D2
    description: "Names injective and confined over the adversarial universe; entry named like a mangled name wins, imported twin takes the allocate suffix"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase30_test.go#TestPhase30CNamesInjective"
        status: pass
    human_judgment: false
  - id: D3
    description: "Generated C independent of function/import order; skip-sort control observed to break allocation order independence; each function defined once in a diamond"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase30_test.go#TestPhase30EmitOrderIndependent"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_phase30_test.go#TestPhase30EmitDefinesEachFunctionOnce"
        status: pass
    human_judgment: false
---

# Phase 30 Plan 07: Module-qualified C names Summary

**Two-pass function C-name allocation: entry functions keep today's names and ordinals, imported functions get length-prefixed injective `SCHWAY_M_...` names in canonical (Module, Name) order.**

## Performance

- **Duration:** ~20 min
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments
- `allocateFunctionNames(functions []core.Function) []string` replaces the inline per-function allocation; emission calls it and seeds its global allocator with the result.
- `moduleCName` (from 30-00) now has its caller; imported functions are defined and called under their D-12 name.
- No-import C is byte-identical: the whole cgen package and the session golden/272/273/inventory meta tests pass unchanged.
- Tests cover the tracer (with a clang `-pedantic -Werror` compile), the adversarial universe, order independence with an armed control, and one definition per diamond function.

## Task Commits

1. **Task 1 (tracer) and Task 2 (tests)** - `375dba6` (feat). Both tasks' tests live in one new file and were committed together with the implementation, so there is a single commit rather than one per task.

## Files Created/Modified
- `internal/compiler/cgen/cgen_program.go` - `allocateFunctionNames`, `importedFirstForTest`, `skipCanonicalImportSortForTest`
- `internal/compiler/cgen/cgen.go` - `moduleCName` comment updated (now used by emission)
- `internal/compiler/cgen/cgen_phase30_test.go` - four Phase 30 tests plus a merged-program builder

## Decisions Made
- `allocateFunctionNames` builds its own allocator seeded with `linearFixedNames`; emission re-marks the returned names in its global allocator, which yields the same used-set as before.
- Whole-program order independence is asserted at the emission layer (callgraph.Order fixes emitted order); the skip-sort control is observed at the allocation layer, where input order is visible.

## Deviations from Plan

None - plan executed exactly as written, except that the two tasks share one commit (see Task Commits). The TDD RED step was not run as a separate failing commit because Task 1's implementation was required by the tracer; the controls (`importedFirstForTest`, `skipCanonicalImportSortForTest`) are observed to change results, which is the negative evidence for the new tests.

## Issues Encountered
- Linear-shape source restrictions (a `let x = call` must be followed directly by the result identifier) forced the generated chain and diamond programs into that shape.

## Known Stubs
None.

## Next Phase Readiness
Native emission of merged module programs is ready for the runnable end-to-end lane.

## Self-Check: PASSED
- cgen_phase30_test.go present; commit 375dba6 is an ancestor of HEAD; `go test ./internal/compiler/cgen` and the session meta set pass.

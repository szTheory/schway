---
phase: 30-local-modules
plan: 06
subsystem: compiler-consumers
tags: [go, corevalidate, callgraph, interp, debugmap, modules, visibility, entry-selection]
requires:
  - phase: 30-local-modules
    provides: 30-05 core.Function.Module/Exported, core.Program.Imports, core.ModuleImport; 30-02 debugmap Entry.Unit field; 30-00 no-import golden
provides:
  - corevalidate independent cross-module visibility law (core.callee_not_exported, core.callee_not_imported) at both OpCall replay sites
  - cloneProgram preserving module facts, with a production control seam
  - callgraph.EntryFunction restricted to entry-module functions
  - interp.findFunction preferring the entry-module function
  - debugmap.Build joined by (module, name) with Entry.Unit set
affects: [30-07, 30-08, 30-09]
actuals:
  tokens: 30000
  tasks: 3
  commits: 3
plan_head_before: de5311b61605d246d850438aa7761c043c5ede4f
plan_head_after: c6492c0315791492495b7441cac1a9a52743904f
commits: 3
tech-stack:
  added: []
  patterns:
    - "visibility law written from core facts only; corevalidate never imports check"
    - "unexported ...ForTest seams inside production paths, each with an armed control observed to break the claim"
key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_phase30_internal_test.go
    - internal/compiler/callgraph/callgraph_phase30_test.go
    - internal/compiler/interp/interp_phase30_test.go
    - internal/compiler/debugmap/debugmap_phase30_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/callgraph/callgraph.go
    - internal/compiler/callgraph/callgraph_entry_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/debugmap/debugmap.go
key-decisions:
  - "visibilityAllows is called directly at both OpCall replay sites, fed by a per-validator importFacts set built from Program.Imports only, so the sites cannot drift and check is never consulted."
  - "The visibility codes are string constants local to corevalidate (calleeNotExportedCode, calleeNotImportedCode), so core.go is untouched."
  - "An imported caller calling the entry module is refused as core.callee_not_exported because the entry callee's Module is empty."
  - "A coordinated lie (matching Exported and Imports facts) is admitted by design; it is the documented frontend trust boundary, cross-checked later by Phase 31 module digests."
requirements-completed: [MOD-04, MOD-05]
duration: 25min
completed: 2026-10-12
status: complete
coverage:
  - id: D1
    description: "corevalidate alone refuses forged cross-module calls (private callee, stripped or wrong import fact, library caller into the entry module) at both replay sites and admits valid and same-module calls"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_phase30_internal_test.go#TestPhase30CorevalidateVisibilityPeer"
        status: pass
    human_judgment: false
  - id: D2
    description: "Admission preserves Function.Module, Function.Exported and Program.Imports through cloneProgram; the lossy-clone mutant is observed"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_phase30_internal_test.go#TestPhase30ValidatedProgramKeepsModuleFacts"
        status: pass
    human_judgment: false
  - id: D3
    description: "EntryFunction ignores a library main and unimported library exports; findFunction prefers the entry-module function"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_phase30_test.go#TestPhase30EntryIgnoresLibraryMain"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_phase30_test.go#TestPhase30FindFunctionPrefersEntryModule"
        status: pass
    human_judgment: false
  - id: D4
    description: "Debug map joins by (module, name) with Entry.Unit set; no-import bytes unchanged"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/debugmap/debugmap_phase30_test.go#TestPhase30DebugMapUnit"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase30PreChangeGolden"
        status: pass
    human_judgment: false
---

# Phase 30 Plan 06: Module-aware consumers Summary

**corevalidate re-derives cross-module visibility independently of check at both OpCall replay sites, and entry selection, interpreter lookup and the debug-map join are now keyed on the entry module or (module, name).**

## Performance

- **Duration:** about 25 min
- **Tasks:** 3
- **Files modified:** 9 (4 created)

## Accomplishments

- `visibilityAllows` refuses a cross-module call as `core.callee_not_exported` unless the callee has a non-empty Module and Exported, and as `core.callee_not_imported` unless Program.Imports holds `{From: caller.Module, Module: callee.Module, Name: callee.Name}`. Absent facts refuse. A library caller into the entry module is refused.
- `TestPhase30CorevalidateVisibilityPeer` runs 8 rows at both the straight-line and the block-shaped replay site; each refused row is re-run under `disableVisibilityPeerForTest` and observed admitted.
- `dropModuleFactsInCloneForTest` inside production `cloneProgram` makes the lossy-clone mutant observable on Module, Exported and Imports.
- `EntryFunction` keeps only `Module == ""` functions as `main` candidates and roots; with the filter off, a library `main` yields the duplicate-main refusal and an unimported export ties with the entry.
- `findFunction` returns the entry-module match first; with the preference off the earlier library `main` wins.
- `debugmap.Build` joins by `funcKey{module, name}` and sets `Entry.Unit`; the bare-name control moves lib `helper`'s source span.

## Task Commits

1. **Task 1: corevalidate visibility law and clone control** - `dbae43f` (feat)
2. **Task 2: entry selection and interpreter lookup** - `b2dddaa` (feat)
3. **Task 3: debug-map (module, name) join** - `c6492c0` (feat)

## Decisions Made

See key-decisions. The tests were authored with the implementation within each task rather than as a separate failing-test commit, so there is no standalone RED commit; every control was nonetheless observed to flip its claim.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] TestNoCoreProgramEntryField rejected the Phase 30 Exported field**
- **Found during:** Task 2 (running `go test ./internal/compiler/callgraph`)
- **Issue:** The D-11-06 source-scan guard failed because 30-05 added `core.Function.Exported` as a module-visibility fact; the guard forbade any field of that name.
- **Fix:** The guard still forbids `EntryFunctionID`, and admits `Exported` only when its JSON tag is `exported,omitempty`, so no-import core bytes stay frozen (D-15, which TestPhase30PreChangeGolden confirms).
- **Files modified:** internal/compiler/callgraph/callgraph_entry_test.go
- **Verification:** callgraph package green.
- **Committed in:** b2dddaa

**Total deviations:** 1 auto-fixed (1 blocking). **Impact on plan:** a pre-existing guard needed to reflect the 30-05 field; no behavior change.

## Issues Encountered

The first block-shaped fixture failed `core.release_order_indeterminate` because a terminal block needs an incoming edge; the fixture now has an entry block, a return block and an edge.

## Known Stubs

None.

## Threat Flags

None. The mitigations for T-30-24, T-30-25 and T-30-26 are implemented and armed as planned.

## Next Phase Readiness

30-07 (cgen) and 30-08 (session) can rely on module facts surviving `Validate(...).Program()`, on entry resolution ignoring library mains, and on debug-map entries carrying their unit.

## Self-Check: PASSED

Created files exist; commits dbae43f, b2dddaa, c6492c0 are ancestors of HEAD; the corevalidate, callgraph, interp and debugmap packages and the session meta set pass.

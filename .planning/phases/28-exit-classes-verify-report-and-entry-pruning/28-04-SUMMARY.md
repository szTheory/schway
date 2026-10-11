---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 04
subsystem: compiler-entry
tags: [entry-resolution, pruning, main-precedence, core.entry_ambiguous, DX-18]
requires:
  - phase: 28-01
    provides: exit-pin inventory and class manifest baseline
  - phase: 28-03
    provides: app verify report v2
provides:
  - callgraph.EntryFunction main precedence (D-01) and callgraph.PruneToReachable (D-04/D-05)
  - session resolveEntry / entryOf / entryRefusalDiagnostic single resolution point (D-04, D-18)
  - core.entry_ambiguous as a source refusal (exit 2) on check, build, run on both engines (D-02)
  - examples/phase28/unused_function.schway runnable gain (ROADMAP SC3)
affects: [28-05, 28-06, 28-07]
status: complete
commits: 3
plan_head_before: 29c92e1a1eea3c2cefeb5cb1bb794176df303f4b
plan_head_after: e1dbb3c7eab7c6e890ba0939ead06c0c8c00fbfd
actuals:
  tokens: 21000
  tasks: 2
  commits: 3
tech-stack:
  added: []
  patterns: [single entry-resolution helper, AST meta-test pinning a call site, prune-after-validate]
key-files:
  created:
    - internal/compiler/callgraph/prune.go
    - internal/compiler/session/session_entry.go
    - internal/compiler/session/session_entry_inventory_test.go
    - internal/compiler/session/session_phase28_prune_test.go
    - internal/compiler/executionpeer/executionpeer_phase28_internal_test.go
    - internal/compiler/native/native_phase28_test.go
    - examples/phase28/unused_function.schway
  modified:
    - internal/compiler/callgraph/callgraph.go
    - internal/compiler/callgraph/callgraph_entry_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_app_verify.go
    - internal/compiler/session/session_phase5_mismatch.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "D-01 main precedence: the function named exactly main is the entry; two mains refuse; without main the strictly-larger-closure tiebreak is unchanged"
  - "Phase4CheckedProgram resolves through resolveEntry for ambiguity but returns the CHECKED (unpruned) program (plan pitfall 7: the Phase 11 differential looks up dead functions by name)"
  - "Check refuses ambiguity after corevalidate/originvalidate/pathoracle (D-18 order)"
---

# Phase 28 Plan 04: Entry resolution and pruning Summary

`main` now wins entry resolution, unreachable functions are pruned after validation and before emission, and an ambiguous entry is a `core.entry_ambiguous` source refusal (exit 2) on check, build and both run engines; a program with an uncalled function now checks, builds and runs.

## Tasks

| Task | Name | Commits |
| ---- | ---- | ------- |
| 1 (RED) | main precedence and prune contract tests | d11c3d0 |
| 1 | callgraph main precedence, PruneToReachable, resolveEntry, BuildApplication, new example | b58194e |
| 2 | every call site on resolveEntry, check refusal, meta-test, corpus/golden/peer/CLI tests | e1dbb3c |

## Observed values (advisory check, no divergence from the plan)

- `examples/phase28/unused_function.schway`: `check` exit 0, `build` exit 0, `app run -- 7` prints `7`, `-- 0` prints `0` (exit 0).
- `multi_function_unreachable` and `multi_function_zero_call`: exactly one execution, `returned:"7"`, on `run --engine=interpreter` and `run --engine=native`.
- `payload_borrow_interaction`: check, build, run interpreter, run native each exit 2 with first diagnostic `core.entry_ambiguous` naming `choose` and `identity` and "add `fn main`".
- All accepted-corpus programs that resolve an entry prune to exactly one root and re-resolve to the same entry; the peer `resolveEntry` returns the same entry with nil error on every pruned program, and still refuses the unpruned `multi_function_zero_call` (negative control).
- Pruned emission equals the frozen phase272/phase273 goldens except the allow-listed files (gate corpus, unreachable, the two phase13 heldout twins, zero_call, new example).

## Deviations from Plan

1. **[Rule 1 - Bug] Phase4CheckedProgram returns the unpruned program.** First attempt returned the pruned program; `TestPhase11InterproceduralDifferential/UnreachableFunction` (not in the handoff list) failed because it looks up `orphan` in the checked program (plan pitfall 7). Fixed by resolving through `resolveEntry` but returning the checked program.
2. **[Rule 3] Line shifts re-keyed** in `testdata/phase16/public-emitter-consumers.json` (session.go `EmitNative`) in both task commits.
3. **[Rule 1] Test-shape fix:** a `len(pruned.Functions) != 1` comparison in a new callgraph test inflated the self-describing guard count; rewritten as an ID-slice comparison.
4. Heldout twin allow-list paths are `testdata/phase13/...` (the plan text said phase11 context); corrected by the first test run.

## Handoff to 28-05: tests red after the final commit (all deliberate, plan 28-05 flips them)

Verified by focused runs (session package filtered by name, plus cgen, callgraph, interp, executionpeer, reduce):

- `TestEmitProgramZeroCallEdges` (package cgen)
- `TestPhase11InterproceduralDifferential/ZeroCallEdges` (package session)
- `TestPayloadCorpusCharacterizationReplay/testdata/phase11/multi_function_zero_call.schway`
- `TestCLICheckAdmissionDivergenceIsExactlyKnown`
- `TestPhase272AcceptedCorpusByteIdentity`
- `TestPhase273ConformanceCorpusByteIdentity`
- `TestPhase28EntryExitClassManifest`
- `TestSelfDescribingDocsGuardIsNotInert` (corpus census: stated 155 programs / 5009 lines, re-derived 156 / 5026 after the new example)

No red test outside this list was observed. The session meta-tests (`go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle' -count=1`) pass clean. `TestNoDuplicatedExitLiterals` passes.

Passing new/related tests: `TestEntryFunctionMainPrecedence`, `TestPruneToReachable`, `TestEntryFunctionEmitsUnreachableFunctions`, `TestEntryFunctionCalledOnlyViaResolveEntry`, `TestPhase28PruneCorpusProperties`, `TestPhase28PrunedEmitMatchesGolden`, `TestPhase28PeerEntryRefusalUnreachableAfterPrune`, `TestUnreachableFunctionProgramRuns`, `TestIndependentEntryResolution`, `TestExclusiveBorrowInterpreterNative`, `TestPhase16PublicEmitterConsumerInventory`, `TestSessionRunSitesDoNotIndexFunctionsZero`.

## Known Stubs

None.

## Threat Flags

None. T-28-11 (dead code validated before pruning) holds by construction: resolveEntry runs after the checker and admission in every path.

## Self-Check: PASSED

Created files exist; commits d11c3d0, b58194e, e1dbb3c are ancestors of HEAD.

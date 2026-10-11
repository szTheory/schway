---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 05
subsystem: test-pins
tags: [pin-flips, frozen-goldens, class-manifest, census, DX-18]
requires:
  - phase: 28-04
    provides: main precedence, PruneToReachable, core.entry_ambiguous refusal
provides:
  - zero-call pins flipped to main precedence (emitter, four-tier differential)
  - phase272/273 identity allow-lists over frozen goldens (D-23)
  - payload replay baseline, admission divergence row, class manifest changes, census
affects: [28-06, 28-07]
status: complete
commits: 2
plan_head_before: 06f2ccc8856cb00d06eb73c0e8a36afd866a5835
plan_head_after: 1ed357659acf4199e9ac9a4c030bbb11fcdca928
actuals:
  tokens: 14000
  tasks: 2
  commits: 2
key-files:
  modified:
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/session/session_phase11_differential_test.go
    - internal/compiler/session/session_phase272_identity_test.go
    - internal/compiler/session/session_phase273_identity_test.go
    - internal/compiler/session/session_payload_replay_test.go
    - internal/compiler/session/session_admission_divergence_test.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/witness_registry_test.go
    - testdata/phase16/public-emitter-consumers.json
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "Goldens (phase272, phase273, phase28 manifest) stay byte-frozen; every flip lives in a named allow-list map with a dated reason"
  - "ZeroCallEdges runs the four tiers on the checked program and validates the independent peer against the pruned program"
---

# Phase 28 Plan 05: Pin flips over frozen goldens Summary

Every pin handed off by 28-04 is green: multi_function_zero_call now resolves to `main` and agrees on all tiers, payload_borrow_interaction is a recorded `core.entry_ambiguous` divergence, and the frozen corpus goldens plus the Phase 28 class manifest absorb the changes only through named allow-list entries.

## Tasks

| Task | Name | Commit |
| ---- | ---- | ------ |
| 1 (tracer) | zero-call flip through emitter, differential, both identity tests | 965cce9 |
| 2 | payload replay, admission divergence, manifest changes, census, inventory re-key | 1ed3576 |

## Observed values

- Starting state confirmed red for exactly the 8 handoff tests; none outside the list.
- zero_call, phase272 (EmitApplication): hash empty, error `only a linear U64-to-U64 or U64-to-Unit entry is supported` (golden: `callgraph.entry_ambiguous`). phase273 (EmitNative): hash `6652285100a4cf459a9e6ff53cf84b444ad403b88245229fb211994d49b8f058` (golden: entry_ambiguous error).
- Payload replay baseline digest for zero_call: `332bbf8297dc56afad24c4af48b3cadd93dd75561228d5bb3281bc65b124ee66`.
- Census: 156 programs, 5,026 lines; guard total stays 55 including tests.
- Manifest changes: zero_call (interpreter run exit 0, native run native.tool_missing, entry main); payload_borrow_interaction (all four surfaces exit 2 core.entry_ambiguous); new example unused_function.

## Deviations from Plan

1. **[Rule 1/3] gate_corpus manifest change (not predicted by plan).** `multi_function_gate_corpus` NativeRun moved `native.tool_failure` -> `native.tool_missing`: the dead function `touch` is pruned, so native run gets past emission to the (deliberately missing) clang. Explained by D-04/D-05 pruning, so added to `phase28ClassChanges`; check/build/interpreter still `core.origin_omitted` (D-18).
2. **[Rule 3] ZeroCallEdges uses checked program for tiers.** Running the interpreter on the pruned (single-function) program emits schema /1 while native emits /2, so the tiers run the checked program and only the independent peer validates against the pruned program (the peer refuses the unpruned two-root program). `phase11CompareFourTiers` is unchanged.
3. **[Rule 3] Emitter inventory and witness registry.** Removing the ZeroCallEdges `EmitNative` refusal call shifted/removed pinned sites. Re-keyed `testdata/phase16/public-emitter-consumers.json` and kept the `typed-refusal` row alive by retaining a refusal control in ZeroCallEdges (payload_borrow_interaction -> `core.EntryAmbiguous`), re-keying `witness_registry_test.go` (:391 -> :424). This file was outside the plan's file list.
4. **Intermediate red.** The Task 1 commit left `TestPhase16PublicEmitterConsumerInventory` red (line shifts); fixed in the Task 2 commit. Final state is clean.
5. Guard-count: a `len(pruned.Functions) != 1` in the new emitter test bumped the self-describing guard total; rewritten as an ID-slice comparison (same fix as 28-04).
6. D-20 re-check: `TestPhase16ProductionPathsPreserveM004Refusal` and the phase271 admission tests pass unchanged; no edit was needed.
7. Planned `TestPhase28PrunedEmitMatchesGolden` and the identity tests resolve to real tests; phase27x goldens and testdata/phase28 untouched (`git diff --name-only 06f2ccc -- testdata/phase272 testdata/phase273 testdata/phase28` prints nothing).

## Verification

- `go test ./internal/compiler/session -run '^(TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase28EntryExitClassManifest|TestPayloadCorpusCharacterizationReplay|TestCLICheckAdmissionDivergenceIsExactlyKnown|TestPhase11InterproceduralDifferential|TestSelfDescribingDocsGuardIsNotInert)$' -count=1` passes.
- `go test ./internal/compiler/cgen -run '^TestEmitProgramZeroCallEdges$' -count=1` passes.
- `go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle' -count=1` passes clean.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Commits 965cce9 and 1ed3576 are ancestors of HEAD; all modified files exist.

---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 06
subsystem: session-admission
tags: [admission, exit-classes, DX-17, native-run, pin-flips]
requires:
  - phase: 28-05
    provides: frozen goldens with named allow-lists, class manifest change mechanism
provides:
  - admitProgram single admission sequence (corevalidate, originvalidate, pathoracle)
  - runNativeChecked split; native runs admit before resolveEntry, cgen or clang
  - build, interpreter application, both run engines report core/origin refusals as diagnostics (exit 2)
  - TestExitClassesUniformAcrossSurfaces shipped-binary pin
affects: [28-07]
status: complete
commits: 3
plan_head_before: 42561f9cd806fd8e3f840df22ab658a3ba9ab39f
plan_head_after: c4d23a36bda7b59b690631ae4946e8c353fa00e8
actuals:
  tokens: 16000
  tasks: 3
  commits: 3
key-files:
  created:
    - internal/compiler/session/session_admit.go
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/session/session_phase271_admission_test.go
    - internal/compiler/session/session_phase16_production_paths_test.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/native/native_test.go
    - internal/compiler/native/native_phase28_test.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "Admission is wired only in the command-file layer (D-19); RunInterpreter, RunNative and runNativeChecked run no origin admission"
  - "admitPublishedApplication retained only for VerifyApplicationCases (plan 28-07 replaces it)"
---

# Phase 28 Plan 06: Uniform exit classes Summary

`admitProgram` now gates check, build, `RunInterpreterApplication`, `RunInterpreterCommandFile` and `RunNativeCommandFile` with one corevalidate, originvalidate, pathoracle sequence on the full program before entry pruning, so the same source gets exit 2 and the same code on every surface and native runs never reach clang for a refused source.

## Tasks

| Task | Name | Commit |
| ---- | ---- | ------ |
| 1 (tracer) | admitProgram, runNativeChecked split, native admits before any tool | 11de921 |
| 2 | build, interpreter application and command file return diagnostics; remaining pin flips | 1a4dab8 |
| 3 | TestExitClassesUniformAcrossSurfaces | c4d23a3 |

## Observed values

- `TestExitClassesUniformAcrossSurfaces` (clang present, not skipped): duplicate_function_name, foreign_origin_omitted and payload_borrow_interaction each exit 2 with the expected first code on check, build, run interpreter and run native; build leaves no artifact.
- `TestPublishedOriginValidatedOnCheckAndRun` now asserts StatusInvalid `core.foreign_origin_omitted` on native and a counting runner with zero invocations.
- Class manifest changes recorded in `phase28ClassChanges`: 13 entries touched in total. 10 origin-refused fixtures plus foreign_origin_omitted: NativeRun 3 -> 2 with the check code; `multi_function_gate_corpus` NativeRun moved tool_missing -> origin_omitted (D-18, supersedes 28-05); `duplicate_function_name` Build/InterpreterRun/NativeRun 3 -> 2 `core.duplicate_function_id`; `shared_copy_accept` InterpreterRun 3 `tool.run_failed` -> 2 `core.origin_omitted` and NativeRun -> 2.

## Deviations from Plan

1. **[Rule 3] Pins pulled into Task 1.** `TestPhase16ProductionPathsPreserveM004Refusal` (restrict_borrow -> acquire_three_success), the manifest NativeRun changes and the emitter-inventory re-key (`session.go:EmitNative` 3096 -> 3113 -> 3088 in Task 2) were done in the commit that caused them, to avoid red intermediate commits. Task 2 added the build/interpreter manifest changes.
2. **Manifest changes beyond the plan's prediction.** The plan expected duplicate_function_name, foreign_origin_omitted and restrict_borrow; the observed set also includes the other origin-refused fixtures (phase07 clean_but_unpublishable, phase10 compose_per_path_borrow_callee_accept, phase25 exclusive_escape_reject and shared_copy_accept, phase3 public_view_*, phase5 false_restrict_hoist). All are explained by D-09/D-10; none are unexplained.
3. The phase271 rows live in `TestApplicationRoutesApplySameAdmissionLayers` (not matched by the plan's `TestPhase271.*` regex); it was run explicitly and passes.
4. `admitProgram` includes pathoracle, so BuildApplication and RunInterpreterApplication now also surface `core.pathoracle_refused` (previously only core and origin layers).
5. `RunNativeFile` is unchanged and still used by exported callers; the command-file lane uses a new unexported `runNativeCommand`, and `runInterpreterCommand` / `runInterpreterResolved` serve the interpreter lanes.

## Verification

- `go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle|TestPhase28|TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestCLICheckAdmissionDivergenceIsExactlyKnown|TestPhase16' -count=1` passes.
- `go test ./cmd/schway -count=1` and `go test ./internal/compiler/exitcode -count=1` pass.
- `go test ./internal/compiler/native -run '^(TestExitClassesUniformAcrossSurfaces|TestUnreachableFunctionProgramRuns|TestShippedBinaryFourSubcommandCorpusMatrix)$' -count=1 -v` passes.
- `go test ./internal/compiler/session -run '^(TestNativeToolFailureIsOperational|TestOwnedBackendMutationIsMismatch|TestApplicationRoutesApplySameAdmissionLayers|TestExclusiveBorrowInterpreterNative)$' -count=1` passes; `go test ./internal/compiler/callgraph -run '^TestSessionRunSitesDoNotIndexFunctionsZero$' -count=1` passes.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Commits 11de921, 1a4dab8 and c4d23a3 are ancestors of HEAD; session_admit.go exists.

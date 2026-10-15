---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "03"
subsystem: cli
tags: [go, interpreter, native, file-io, argv, authority]

# Dependency graph
requires:
  - phase: 37-02
    provides: fully admitted interpreter file-application API and checked file outcomes
provides:
  - separator-marked source file-run forms for interpreter and native engines
  - exact CLI outcome tests and pre-prune whole-program authority controls
affects: [37-04, interpreter-cli, native-cli, application-file-authority]

# Actuals (#2632), measured as chars/4 over the realized plan diff.
actuals:
  tokens: 8544.5
  tasks: 2
  commits: 3
plan_head_before: 40e692f9b4c865bf622fe0a06bef2d6ed58b2948
plan_head_after: cd7f50d52bf45da476e7589123e70ee336ce731e

# Tech tracking
tech-stack:
  added: []
  patterns:
    - early separator dispatch preserves one opaque application argument
    - native source execution builds in a private temporary directory and runs once
    - full-program admission is tested before entry pruning with a reader sentinel

key-files:
  created:
    - cmd/schway/main_phase37_file_run_test.go
    - internal/compiler/session/session_phase37_authority_test.go
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-03-task1-red-evidence.json
  modified:
    - cmd/schway/main.go
    - internal/compiler/session/session_file_application.go

key-decisions:
  - "Keep the no-tail command route unchanged and recognize separator-marked source runs before global option extraction."
  - "Keep interpreter IO on the fixed checked Go reader; require an explicit manifest only for native build authority."
  - "Leave APP-14, APP-13, and MOD-07 pending parent phase acceptance."

requirements-completed: []

coverage:
  - id: D1
    description: "Public interpreter and native source forms run an admitted CRC file program with one opaque path argument and fixed outcome tuples."
    verification:
      - kind: integration
        ref: "go test ./cmd/schway -run '^(TestPhase37PublicFileRunTracer|TestPhase37FileRunGrammar|TestPhase37OpaqueFilePaths|TestPhase37FileRunInputBoundary)$' -count=1 -v -timeout=600s"
        status: pass
      - kind: integration
        ref: "go test ./cmd/schway -run '^(TestPhase37FileRunOutcomes|TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s"
        status: pass
    human_judgment: true
    rationale: "This plan supplies the CLI slice; parent phase acceptance must consider it with the consumer and guide deliverables in the remaining plan."
  - id: D2
    description: "Whole merged source and raw-core authority checks reject neighboring programs before any host read, including an invalid imported function that entry pruning would omit."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^TestPhase37FileApplicationAuthority$' -count=1 -v -timeout=600s"
        status: pass
    human_judgment: true
    rationale: "The parent phase still owns acceptance of the complete capability and its downstream guide evidence."

# Metrics
duration: 18min
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 03: Public Source File-Run Route Summary

**Interpreter and native `schway run` source forms preserve opaque file paths, fixed outcome classes, and full-program authority admission.**

## Performance

- **Duration:** 18 min
- **Started:** 2026-10-15T15:24:23Z
- **Completed:** 2026-10-15T15:41:56Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added early separator-marked dispatch for `run --engine=interpreter SOURCE -- PATH` and `run --engine=native SOURCE --manifest MANIFEST -- PATH`; the single suffix token remains opaque, including option-shaped values and a literal `--`.
- Exercised both engines against the CRC fixture and verified `(0, "3421780262\n", "")`, no interpreter Clang invocation, one native run without a banner, and private native temporary-build cleanup.
- Pinned empty, 4096-byte, 4097-byte, NUL, typed file error, checked-defect, source-refusal, and missing source/tool/manifest outcomes; preserved the no-tail PathToken and conformance routes.
- Proved the full merged independent admission runs before entry pruning: an invalid unused imported function and eight neighboring/raw-core controls were refused, with the reader sentinel remaining zero for refusals and reaching one for the valid positive control. Both module-root traversal and symlink escape were refused.

## Verification

All focused roots passed with no skips, using `GOCACHE=/tmp/ai-lang-verification-gocache`:

1. `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileRunTracer|TestPhase37FileRunGrammar|TestPhase37OpaqueFilePaths|TestPhase37FileRunInputBoundary)$' -count=1 -v -timeout=600s`
2. `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37FileApplicationAuthority$' -count=1 -v -timeout=600s`
3. `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37FileRunOutcomes|TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s`

The plan's task-1 TDD RED evidence is recorded in `37-03-task1-red-evidence.json`: the tracer root failed against the old usage-only behavior, then passed after the route was implemented. Task 2's authority and outcome behavior already existed after task 1 and the 37-02 session API, so there was no valid semantic RED for its additional test coverage; early fixture/source-shape mistakes were corrected before recording final passing evidence.

## Task Commits

1. **Task 1 RED:** `9bf7c3d6` (`test(37-03): add source file run tracer tests`)
2. **Task 1 GREEN:** `7163bc49` (`feat(37-03): add public source file-run route`)
3. **Task 2:** `cd7f50d5` (`test(37-03): cover file run authority and outcomes`)

**Plan metadata:** committed with the GSD state and roadmap tracking update.

## Files Created/Modified

- `cmd/schway/main.go` - early source-run grammar and interpreter/native route.
- `cmd/schway/main_phase37_file_run_test.go` - tracer, grammar, opaque paths, bounds, and outcome tuples.
- `internal/compiler/session/session_phase37_authority_test.go` - neighboring-source and raw-core gates plus read-count controls.
- `internal/compiler/session/session_file_application.go` - package-private runner seam for proving the pre-read boundary.
- `37-03-task1-red-evidence.json` - machine-readable TDD RED evidence.

## Decisions Made

- Preserve the established no-tail route and dispatch the new grammar before global extractors.
- Keep the interpreter path on the checked fixed reader; native execution alone accepts an explicit build manifest.
- Keep APP-14, APP-13, and MOD-07 pending parent phase acceptance; this summary records the witnesses without marking those broader requirements complete.

## Deviations from Plan

### Plan-order TDD overlap

Task 2 was marked TDD, but its production admission boundary had already been implemented by plan 37-02 and the task 1 route. The task adds independent verification coverage rather than new behavior, and its exact roots passed before any code change for task 2. No artificial failing test was introduced. A package-private `runAdmittedFileApplication` seam was added to prove that rejected inputs cause zero file reads while a valid positive control reaches the runner once.

**Total deviations:** 1 plan-order verification adjustment. No product behavior was expanded beyond the plan.

## Issues Encountered

Initial task-2 test fixtures did not match the supported source/defect shapes. They were revised to use an imported borrowed function omitted by entry pruning and a real checked U64 overflow; the final exact roots pass.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The CLI and authority witnesses for plan 37-03 are ready for parent phase acceptance. Plan 37-04 remains responsible for the executable guide evidence; phase requirements stay pending until the parent accepts the complete phase.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

## Self-Check: PASSED

- Summary and all five plan output files exist.
- Task commits `9bf7c3d6`, `7163bc49`, and `cd7f50d5` are ancestors of HEAD.
- Focused verification roots passed; no stub patterns or whitespace errors were found.

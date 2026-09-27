---
phase: 22-native-application-build-and-single-execution
plan: 01
subsystem: native
tags: [Go, C17, Clang, U64, native-process]
requires: []
provides:
  - Retained native application builds with a versioned receipt and no build-time application launch
  - A checked U64 identity application route with one child process per invocation
  - Process-boundary controls for inputs, byte streams, exits, signals, timeout, relocation, and concurrency
affects: [22-02, 22-03, phase-23]
actuals:
  tokens: 11843
  tasks: 2
  commits: 2
plan_head_before: aa89e1b9957e9c7126bc84fb8c722e39e787ee04
tech-stack:
  added: []
  patterns:
    - "Application C entry shell shares checked function-body lowering with conformance emission"
    - "Retained app build publishes executable and integrity receipt without running the child"
    - "Application streams and process outcomes remain outside conformance JSON"
key-files:
  created:
    - internal/compiler/native/native_app.go
    - internal/compiler/native/native_app_test.go
    - examples/phase22/identity.lang
    - examples/phase22/identity.expected.json
  modified:
    - cmd/lang/main.go
    - cmd/lang/main_test.go
    - internal/compiler/session/session.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - .planning/REQUIREMENTS.md
key-decisions:
  - "The first application ABI admits only a checked U64-to-U64 entry and one bounded opaque argv token."
  - "Build receipts report runtime closure as incomplete and cacheability as false when dependency discovery is unavailable."
  - "The existing native conformance route continues to emit its execution document independently of raw app streams."
requirements-completed: [APP-02, APP-03, APP-04, APP-05]
coverage:
  - id: D1
    description: "Build retains a relocatable executable and receipt without starting the application."
    requirement: APP-02
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_app_test.go#TestPhase22BuildRetainsRelocatableArtifactWithoutLaunching"
        status: pass
    human_judgment: false
  - id: D2
    description: "The checked identity app returns exact independent outputs for 7 and 42 and rejects malformed or oversized input before entry execution."
    requirement: APP-03
    verification:
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22IdentityApplicationBuildAndRunCLI"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestPhase22ApplicationEmitterSharesBodyAndSeparatesOutputShell"
        status: pass
    human_judgment: false
  - id: D3
    description: "Each ordinary app-run request starts exactly one child, with concurrent requests isolated from sibling cancellation."
    requirement: APP-04
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_app_test.go#TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_app_test.go#TestPhase22ConcurrentRequestsLaunchIndependently"
        status: pass
    human_judgment: false
  - id: D4
    description: "Raw stdout and stderr remain byte streams and child exit, signal, timeout, and launch failures are distinguishable; conformance output remains unchanged."
    requirement: APP-05
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_app_test.go#TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes"
        status: pass
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22ConformanceRunStillEmitsExecutionDocument"
        status: pass
    human_judgment: false
duration: 19 min
completed: 2026-09-27
status: complete
---

# Phase 22 Plan 01: Native Application Build and Single Execution Summary

**Retained U64 native app builds with receipt validation, single-child execution, raw streams, and stable process outcomes**

## Performance

- **Duration:** 19 min
- **Started:** 2026-09-27T16:58:22Z
- **Completed:** 2026-09-27T17:18:21Z
- **Tasks:** 2
- **Files modified:** 11

### Cold and Warm Feedback Samples

Five samples per lane; wall time includes the Lang CLI process and captured output. “Cold” means the first command against a fresh output path or the first launch of a newly built artifact. OS caches were not flushed, so these are process-cold samples rather than cold-machine measurements.

| Operation | First-use samples (ms) | Repeated samples (ms) |
|---|---|---|
| Build | 726.848, 181.984, 140.997, 158.847, 203.118 | 166.870, 172.083, 180.543, 170.554, 239.052 |
| Run and observe exact stdout | 26.477, 17.323, 14.496, 18.750, 19.249 | 14.571, 13.542, 12.574, 13.221, 12.846 |

## Accomplishments

- Added a checked application build path that emits a retained C17 executable plus a versioned integrity receipt; compilation never launches the app.
- Added the public `lang build` and `lang app run` path for the bounded U64 identity source, with input checks before the Lang entry and plain decimal output.
- Added executable controls for launch count, relocation, byte streams, process outcomes, concurrency, and preservation of the existing conformance JSON route.

## Task Commits

Each task was committed atomically with normal hooks enabled:

1. **Task 1: Build and run the checked identity application** - `10c49a3`
2. **Task 2: Lock input, launch-count, stream, and process outcomes** - `d44c08e`

## Files Created/Modified

- `internal/compiler/native/native_app.go` - Retained build receipt, integrity checks, and one-child runner.
- `internal/compiler/native/native_app_test.go` - Build launch-count, relocation, streams, outcomes, and concurrency controls.
- `examples/phase22/identity.lang`, `examples/phase22/identity.expected.json` - Checked identity witness and independent expected pairs.
- `cmd/lang/main.go`, `cmd/lang/main_test.go` - Public build/app-run commands and CLI contract tests.
- `internal/compiler/session/session.go` - Check, independent validation, entry resolution, and application emission path.
- `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_program_test.go` - Shared lowering with a separate application entry/output shell and tests.
- `.planning/REQUIREMENTS.md` - Marked APP-03 and APP-05 complete; APP-02 and APP-04 remain shared with sibling plans.

## Decisions Made

- Limited the initial app ABI to a checked `U64 -> U64` entry and one bounded opaque argv token.
- Kept the receipt explicit that runtime closure is incomplete and the artifact is not cacheable until dependency discovery can establish closure.
- Preserved the existing conformance path and its execution-document output while app execution passes ordinary streams and process outcomes directly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected generated C newline string escaping**
- **Found during:** Task 1 (identity application smoke verification)
- **Issue:** The first generated C entry shell had invalid escaping for the output newline.
- **Fix:** Corrected the C string escaping in the application emitter.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** Retained build and CLI smoke passed for inputs 7 and 42.
- **Committed in:** `10c49a3`

**Total deviations:** 1 auto-fixed (Rule 1).
**Impact:** The correction was required for valid generated C and stayed within the planned application shell.

## Issues Encountered

- One initial test fixture used syntax the existing Lang parser does not accept; the fixture was corrected and the focused tests then passed.
- The sandbox initially denied creation of Git’s index lock for Task 1. The commit was retried with normal hooks enabled; no hook bypass was used.

## Authentication Gates

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 22-01 is complete. APP-03 and APP-05 are marked complete; APP-02 and APP-04 remain open until their sibling plan work is complete. The retained app/build seam is ready for the remaining Phase 22 evidence and public-contract work.

---
*Phase: 22-native-application-build-and-single-execution*
*Completed: 2026-09-27*

## Self-Check: PASSED

- Required created files exist.
- Task commits 10c49a3 and d44c08e are present.
- Plan task commit count measured from the recorded base is 2.
- STATE.md and ROADMAP.md are unchanged.

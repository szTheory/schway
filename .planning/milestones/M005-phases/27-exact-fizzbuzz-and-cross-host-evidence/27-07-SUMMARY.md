---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 07
subsystem: application
tags: [output-bounds, interpreter, c17, native-runner, evidence]

requires:
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    provides: Literal text writes, guarded FizzBuzz, and retained native app route
provides:
  - Incremental 65,536-byte application stdout ceiling shared by interpreter and generated C
  - Exact-limit and crossing-write behavior for explicit FizzBuzz output and automatic U64 return output
  - Independent application stdout and evidence-document capacity outcomes
affects: [phase-27-parity, native-output, app-evidence]

actuals:
  tokens: 9111
  tasks: 2
  commits: 1

tech-stack:
  added: []
  patterns: [whole-chunk-output-accounting, independent-evidence-capacity, retained-app-boundary-tests]
key-files:
  created:
    - internal/compiler/cgen/cgen_phase27_output_test.go
    - internal/compiler/interp/interp_phase27_output_test.go
    - internal/compiler/session/session_phase27_output_test.go
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/execution/execution.go
    - internal/compiler/interp/interp.go
    - internal/compiler/native/native_app.go
    - internal/compiler/native/native_app_test.go
    - cmd/schway/main_test.go
key-decisions:
  - "Count each complete text, decimal U64, or LF chunk before writing, and reject the first crossing chunk while retaining earlier completed output."
  - "Keep the application stdout limit distinct from evidence serialization capacity; report evidence exhaustion without changing a successful child result."
  - "Treat evidence documents exactly equal to their configured capacity as complete; only documents larger than capacity are exhausted."
patterns-established:
  - "Interpreter and generated C use matching write chunk boundaries, including legacy automatic decimal return output."
  - "Public app tests assert child status, stdout, stderr, and evidence/tool errors independently."
requirements-completed: [APP-09, APP-08, EVD-12]
coverage:
  - id: D1
    description: "The interpreter and generated C accept exact-limit output and reject a crossing whole chunk while retaining earlier completed chunks."
    requirement: APP-09
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/session -run '^TestPhase27Output(ExactLimit|CrossingWrite|EvidenceSeparation|ProductionLimit)$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Evidence capacity is independent from application stdout, and an evidence document exactly at capacity remains valid."
    requirement: EVD-12
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase27OutputEvidenceExactLimit$' -count=1"
        status: pass
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase27OutputEvidenceSeparation$' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: "Retained applications preserve public input-boundary behavior and expose output-limit failure as exit 65 with the pinned stderr reason."
    requirement: APP-09
    verification:
      - kind: e2e
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./internal/compiler/native ./internal/compiler/session -run '^TestPhase27(AppBoundary|OutputPublicStatus|OutputCrossingWrite|DynamicZero)$|^TestPhase26PublicAppCLI$' -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 43min
completed: 2026-10-03
tasks: 2
files_modified: 10
status: complete
---

# Phase 27 Plan 07: Bounded Application Output Summary

**FizzBuzz and sum output now share a 65,536-byte per-write ceiling, with evidence-capacity failures kept separate from child process results.**

## Performance

- **Duration:** 43 min
- **Started:** 2026-10-03T16:37:59-04:00
- **Completed:** 2026-10-03T17:20:49-04:00
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments

- Added a distinct 65,536-byte application stdout limit in the interpreter and emitted C. Each complete text literal, decimal U64 value, LF, and legacy automatic return-output chunk is checked before writing.
- Pinned exact-limit success and crossing-write failure for FizzBuzz and sum. Crossing chunks are omitted, earlier bytes remain, and native applications return exit 65 with the exact output-limit stderr.
- Verified that evidence capacity can be exhausted while application stdout and the child exit remain successful, and that evidence exactly at its capacity is accepted.
- Added public CLI boundary coverage for FizzBuzz inputs 0 through 1000 and pre-output rejection at 1001; retained sum boundary and dynamic-zero controls also pass.

## Task Commits

1. **Tasks 1–2: Enforce bounded application output and pin outcomes** - `817396f` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `internal/compiler/execution/execution.go` - Defines the application stdout ceiling independently from evidence capacity.
- `internal/compiler/interp/interp.go` - Counts complete output chunks, preserves prior bytes, and returns a stable defect on crossing.
- `internal/compiler/cgen/cgen.go` and `internal/compiler/cgen/cgen_program.go` - Emit bounded native app output, deterministic exit 65 failures, and an evidence-capacity marker.
- `internal/compiler/native/native_app.go` - Accepts evidence documents exactly at capacity and rejects only larger captures.
- `internal/compiler/cgen/cgen_phase27_output_test.go`, `internal/compiler/interp/interp_phase27_output_test.go`, and `internal/compiler/session/session_phase27_output_test.go` - Cover exact and crossing output, evidence separation, and the production limit.
- `internal/compiler/native/native_app_test.go` and `cmd/schway/main_test.go` - Cover exact evidence capacity and retained public CLI outcomes.

## Decisions Made

- Kept the application byte ceiling separate from both execution-document and evidence-document limits.
- Used a subtraction-safe remaining-capacity check and charged bytes only after a complete chunk is written.
- Changed native evidence decoding to accept an exactly full valid document; the former greater-than-or-equal check incorrectly classified that boundary as exhausted.

## Deviations from Plan

- Added `MaxApplicationOutputBytes` to the execution contract and an exact-capacity native evidence test alongside the planned app tests. These make the production ceiling explicit and catch an existing exact-boundary decoder mismatch.
- `session.go` required no implementation change: its existing application builder already uses the C emitter, and the new session integration test exercises the evidence seam through that path.

**Total deviations:** 1 necessary correctness update.

## Issues Encountered

- A focused exact-capacity evidence test first exposed that the native reader used a greater-than-or-equal boundary check. After changing it to reject only captures larger than capacity, the exact-capacity test passed.
- Focused selectors passed. No full project suite was run; the package-level times shown in Go output included 1.195s for the cgen output tests, 0.021s for interpreter output tests, 0.212s for session output tests, and 1.123s for the public CLI selector.

## Next Plan Readiness

- Plan 27-08 can validate the public CLI and process outcome contract against the completed FizzBuzz and sum output boundaries.
- Output and evidence capacity now have independent controls for the remaining cross-host and documentation evidence plans.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*

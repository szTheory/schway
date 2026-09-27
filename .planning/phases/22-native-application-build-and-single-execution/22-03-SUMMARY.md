---
phase: 22-native-application-build-and-single-execution
plan: 03
subsystem: application
tags: [Go, C17, Clang, application-evidence, differential-verification]
requires:
  - phase: 22-01
    provides: Retained U64 application builds and one-child application execution
  - phase: 22-02
    provides: Build/input identities, receipt validation, and explicit local C build authority
provides:
  - Bounded same-run compiler-event sidecars with closed disabled, complete, incomplete, and capacity-exhausted states
  - Explicit replay verification of isolated source inputs against independent answers and interpreter/O0/O3 traces
  - A verifier-only scripted outcome adapter with explicit host-IO and physical-cleanup limits
affects: [phase-23, APP-04, APP-06, EVD-11]
actuals:
  tokens: 19219
  tasks: 2
  commits: 3
  plan_head_before: e5d65f4ea13fcfba92e34ba1f62c0f855c760c11
tech-stack:
  added: []
  patterns:
    - Keep same-run application capture on a private channel, separate from ordinary output streams
    - Replay explicit cases through the checked core, shared C emitter, and independent expected values
    - Mark host dependency closure incomplete and disable cache reuse when provenance is unknown
key-files:
  created:
    - internal/compiler/session/session_app_verify.go
    - examples/phase22/identity.cases.json
    - examples/phase22/README.md
  modified:
    - cmd/lang/main.go
    - cmd/lang/main_test.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/execution/execution.go
    - internal/compiler/native/native.go
    - internal/compiler/native/native_app.go
    - internal/compiler/native/native_app_test.go
key-decisions:
  - "A complete same-run event capture is an execution record only; its report always has verified=false."
  - "Only the explicit app verify route runs interpreter/O0/O3 replay, with independent checked-in answers."
  - "The fixture.value scripted outcome is confined to a verifier-only adapter and never enters Lang lowering or local C execution."
  - "Verification reports retain incomplete dependency closure and cacheable=false because host toolchain closure is not captured."
requirements-completed: [APP-04, APP-06, EVD-11]
coverage:
  - id: D1
    description: "One retained app process can emit bounded compiler events to a separate status-bearing report while preserving ordinary streams and outcome."
    requirement: APP-04
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_app_test.go#TestPhase22EvidenceDisabledCompleteAndStreamIsolation"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/native_app_test.go#TestPhase22EvidenceMissingPartialAndCapacityControls"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/native_app_test.go#TestPhase22EvidenceReportWriteFailurePreservesAppOutcome"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestPhase22ApplicationEmitterSharesBodyAndSeparatesOutputShell"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang build examples/phase22/identity.lang --output /tmp/ai-lang-phase22-03-identity; go run ./cmd/lang app run /tmp/ai-lang-phase22-03-identity --report /tmp/ai-lang-phase22-03-evidence.json --evidence=events -- 42"
        status: pass
    human_judgment: false
  - id: D2
    description: "Explicit verification compares independent 7/42 answers and ordered interpreter/O0/O3 execution evidence; verifier-only modeled outcomes fail closed."
    requirement: EVD-11
    verification:
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22AppVerifyIndependentIdentityCases"
        status: pass
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22AppVerifyWrongExpectedAnswerControl"
        status: pass
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22AppVerifyModelOnlyOutcomes"
        status: pass
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22AppVerifyRejectsOrdinaryForeignScriptsAndLocalC"
        status: pass
      - kind: integration
        ref: "cmd/lang/main_test.go#TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute"
        status: pass
      - kind: other
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -count=1"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang app verify examples/phase22/identity.lang --cases examples/phase22/identity.cases.json --report /tmp/ai-lang-phase22-03-verification.json"
        status: pass
    human_judgment: false
  - id: D3
    description: "The Phase 22 README documents build, run, evidence, replay, input bounds, local C authority, and modeled-world limits."
    verification: []
    human_judgment: true
    rationale: "The accuracy is checked against implementation, but clarity of the public contract benefits from human review."
duration: 43 min
completed: 2026-09-27
status: complete
---

# Phase 22 Plan 03: Single-Run Evidence and Explicit Replay Summary

**Bounded app event sidecars and a separate 7/42 interpreter/O0/O3 verifier with an isolated scripted model case**

## Performance

- **Duration:** 43 min
- **Started:** 2026-09-27T19:08:01Z
- **Completed:** 2026-09-27T19:51:29Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments

- Added optional `--report` and `--evidence=events` to ordinary app runs. A private bounded channel captures checked events from the same child, with explicit capture status and `verified:false`; stdout, stderr, and child outcome remain on their existing path.
- Added `lang app verify SOURCE --cases CASES --report REPORT`, comparing each independent identity input against the interpreter and generated-C `-O0`/`-O3` traces and the fixture's literal expected terminal value.
- Added a closed verifier-only `fixture.value` model adapter plus missing, duplicate, unconsumed, and expected-value mismatch controls. Model reports make no host-IO or physical-cleanup claim.
- Documented the public build, app-run, evidence, replay, and local C manifest contracts in the Phase 22 example README.

## Task Commits

1. **22-03-T1: Capture compiler events from the one application process** — `144abc0` (`feat`).
2. **22-03 corrective fix: Use a portable stdout fallback for generated capture code** — `214c6cb` (`fix`).
3. **22-03-T2: Verify replayable inputs against independent answers** — `c91eca6` (`feat`).

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase22Evidence' -count=1 -v` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang -run '^TestPhase22' -count=1 -v` — passed, including positive replay, wrong-answer, malformed model-outcome, local C refusal, and ordinary-run routing controls.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -count=1` — passed.
- Built `examples/phase22/identity.lang`, ran input `42` with same-run evidence and observed stdout `42` plus a separate `lang.app-evidence/1` report with `capture_status: complete` and `verified: false`; the uncaptured input `7` printed `7`.
- Ran the documented `lang app verify` command; both cases passed with interpreter/O0/O3 execution documents and independent expected values.
- `git diff --check` — passed.

## Decisions Made

- Reused the established `session.go` checked-core, projection, emission, and comparison APIs. The new replay service is in the planned `session_app_verify.go` file.
- Kept ordinary app execution single-run and separate from explicit differential replay. Evidence capture cannot establish semantic verification.
- Made the generated-C verifier non-cacheable and recorded dependency closure as incomplete because the installed host SDK/linker/runtime are not fully identified.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Dated artifact correction (2026-09-27): the plan's read-first list named a separate session-entry source that was not present**
- **Found during:** Task 2 read-first gate.
- **Issue:** Phase 22-02 had already placed the application entry in the existing `session.go`, so a separate session-entry source was unnecessary.
- **Fix:** Reused the existing session APIs and implemented this plan's verifier in `session_app_verify.go`, avoiding a duplicate application entry.
- **Files modified:** `internal/compiler/session/session_app_verify.go`.
- **Verification:** Full session package and Phase 22 CLI tests passed.
- **Committed in:** `c91eca6`.

**2. [Rule 1 - Bug] macOS rejected the generated global stdout initializer**
- **Found during:** Plan-level documented app build after Task 1.
- **Issue:** The generated `static FILE *` initialized from `stdout` was not a compile-time constant with the macOS SDK headers.
- **Fix:** Initialize the pointer to `NULL` and select `stdout` at runtime unless evidence capture switches to the private file; added an emitter regression assertion.
- **Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_program_test.go`.
- **Verification:** Focused emitter tests, full cgen suite, and actual retained app build/run with event capture passed.
- **Committed in:** `214c6cb`.

**Total deviations:** 2 auto-fixed (1 Rule 1 bug, 1 Rule 3 blocking path correction). **Impact:** Both changes were needed to complete the planned verifier and compile the generated application on this host; no language semantics or dependencies were added.

## Issues Encountered

The first generated app build exposed the macOS initializer issue listed above. It was fixed and the documented build/run/verify path then passed. No remaining blockers.

## User Setup Required

None.

## Next Phase Readiness

The explicit verifier and same-run report formats are ready for subsequent resource-evidence work. This worktree exercised macOS; the phase-level Linux host lane and full Go suite remain for the orchestrator. Host dependency closure remains explicitly incomplete and reports are not cacheable.

## Self-Check: PASSED

- All 12 source, test, and example files listed in the plan diff exist.
- Task commits `144abc0`, `214c6cb`, and `c91eca6` exist; the persisted plan ledger measures three commits before this summary.
- Required native evidence, cgen, CLI, and session checks passed, as did documented app build/run/verify commands.
- No stubs, skipped tests, or unrun plan verifications were found.
- `STATE.md` and `ROADMAP.md` were not changed.

---
*Phase: 22-native-application-build-and-single-execution*
*Completed: 2026-09-27*

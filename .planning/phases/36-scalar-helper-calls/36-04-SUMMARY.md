---
phase: 36-scalar-helper-calls
plan: "04"
subsystem: compiler-runtime
tags: [interpreter, executionpeer, scalar-calls, bounded-evidence]

requires:
  - phase: 36-03
    provides: Typed scalar call contracts and independent raw/source validation.
provides:
  - Interpreter execution for all four scalar parameter/result pairs with fresh dynamic activation ordinals.
  - Independent causal validation of repeated and nested invocation paths without numeric replay.
  - Bounded call-evidence capture that preserves actual results and output when evidence is exhausted.
affects: [36-05, interpreter, executionpeer, CALL-01, CALL-02]

actuals:
  tokens: 18197
  tasks: 3
  commits: 8
commits: 8
plan_head_before: d5efeeda04e78ff2d4e76d1baedf8572f0fbcedf
plan_head_after: 35fb0fe259dee872090a1c8bdc47580ce672eefc

tech-stack:
  added: []
  patterns:
    - Per-frame and per-static-site activation ordinals, reset for each dynamic parent.
    - Independent event-stack causal validation, separate from interpreter arithmetic.
    - Bounded evidence collection with explicit incomplete-capture error and preserved execution result.

key-files:
  created:
    - internal/compiler/interp/interp_phase36_test.go
    - internal/compiler/executionpeer/executionpeer_phase36_test.go
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/executionpeer/executionpeer.go

key-decisions:
  - "Dynamic invocation identity uses a monotonically assigned ordinal per parent frame and static call site; each new parent begins at ordinal zero."
  - "The execution peer derives call ordinals and terminal ancestry from observed events and checked core facts; it does not evaluate scalar arithmetic."
  - "Evidence exhaustion reports incomplete capture while returning the actual interpreter execution and application output."

requirements-completed: []

coverage:
  - id: D1
    description: "The interpreter executes repeated scalar helpers with literal U64/Bool results and attributes later overflow/remainder defects to the correct activation."
    requirement: CALL-01
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_phase36_test.go#TestPhase36InterpreterRepeatedCallTracer"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_phase36_test.go#TestPhase36InterpreterCallAnswersAndDefects"
        status: pass
    human_judgment: false
  - id: D2
    description: "The execution peer independently validates repeated dynamic ordinals, parent/site identity, call-result shape, and paired causal terminals without claiming numeric proof."
    requirement: CALL-02
    verification:
      - kind: unit
        ref: "internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36DynamicCallOrdinalControls"
        status: pass
      - kind: unit
        ref: "internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls"
        status: pass
    human_judgment: false
  - id: D3
    description: "Call evidence retention is bounded and fail-closed while exact-path, exact-event, exhaustion, ordinal-limit, and more-than-4096-work controls preserve the real output and terminal reason."
    requirement: CALL-02
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_phase36_test.go#TestPhase36InterpreterEvidenceBoundaries"
        status: pass
      - kind: unit
        ref: "internal/compiler/executionpeer/executionpeer_phase36_test.go#TestPhase36CallCausalAttributionControls"
        status: pass
    human_judgment: false

duration: 27min
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 04: Scalar Helper Runtime Calls Summary

Repeated scalar helper calls now execute to real U64 and Bool values with canonical per-parent activation paths, independent causal checks, and bounded evidence capture that preserves actual outcomes on exhaustion.

## Performance

- **Duration:** 27 min
- **Started:** 2026-10-15T02:26:13Z
- **Completed:** 2026-10-15T02:53:10Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- The interpreter runs all four U64/Bool parameter and result combinations, including nested/computed arguments, mutable assignment, Boolean branching, helper-local loops, distinct call sites, and fresh nested-parent ordinals.
- The execution peer independently accepts valid repeated traces and rejects ordinal gaps, duplicates, reordering, wrong parent/site/callee, forged source/target/result shape, missing call/return, repeated terminals, and post-terminal events. A plausible false-success trace remains a causal success only; numeric correctness belongs to the interpreter assertions.
- A bounded collector checks event/path budgets before retention, enforces a portable ordinal ceiling, and returns `ErrEvidenceCapacityExhausted` with the actual execution/output. More than 4,096 dynamic calls remain executable, and a later real defect remains observable after capture stops.

## Verification

Each task's exact focused selector produced fresh root RUN/PASS evidence. The final task selector was rerun after the last test commit; `git diff --check` also passed.

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/executionpeer -run '^(TestPhase36InterpreterRepeatedCallTracer|TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/executionpeer -run '^(TestPhase36InterpreterCallAnswersAndDefects|TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls|TestPhase26PeerOccurrenceOrder|TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/executionpeer -run '^(TestPhase36InterpreterEvidenceBoundaries|TestPhase36InterpreterRepeatedCallTracer|TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls)$' -count=1 -v -timeout=600s`

No full suite, whole-package test, native matrix, or new dependency was used. The execution peer's structural verdict is not evidence of arithmetic correctness.

## Task Commits

1. **Task 1: Repeated helper tracer and activation identity** — `d5828817` (test), `f6d6bdb2` (feat).
2. **Task 2: Numeric answers and independent causal controls** — `f23a06b0` (test), `7c82478a` (fix).
3. **Task 3: Bounded evidence and preserved outcomes** — `5ccaced4` (test), `fd6d3b2a` (feat), `7d42ebbe` (fix), `35fb0fe2` (test).

The separate plan-metadata commit records this summary and canonical planning state.

## Files Created/Modified

- `internal/compiler/interp/interp.go` — executes scalar calls, assigns dynamic activation identity, bounds captured evidence, and preserves results on evidence exhaustion.
- `internal/compiler/interp/interp_phase36_test.go` — interpreter value, defect, repeated-call, and evidence-boundary controls.
- `internal/compiler/executionpeer/executionpeer.go` — independently validates dynamic causal identity with input-bounded state.
- `internal/compiler/executionpeer/executionpeer_phase36_test.go` — positive and forged causal-trace controls.

## Decisions Made

- Dynamic ordinals belong to a parent frame and static call site; separate sites do not share counters, and fresh parents restart at zero.
- The peer validates call structure and causal attribution from its own observed stack and checked core topology, without numeric replay.
- Evidence exhaustion separates incomplete capture from execution semantics and retains the real result, output, and terminal reason.

## Deviations from Plan

None — the plan was executed as written.

## Issues Encountered

The evidence-exhaustion boundary exposed caller wrappers that could drop the actual execution on the evidence-error path. The wrappers were corrected within the planned preservation contract and verified by the fixture-route boundary test.

## User Setup Required

None — no external service or local setup is required.

## Next Phase Readiness

Plan 36-04 provides the interpreter and execution-peer contracts required by the remaining Phase 36 plans. CALL-01 and CALL-02 remain open until the full phase-level obligations are complete.

---
*Phase: 36-scalar-helper-calls*
*Completed: 2026-10-15*

## Self-Check: PASSED

- Summary file exists at the required phase path.
- All eight task commits are ancestors of the current plan head.
- `git diff --check` passed.

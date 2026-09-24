---
phase: 18-branch-on-a-computed-value
plan: 04
subsystem: compiler
tags: [checker, core, interpreter, cgen, differential-testing]
requires:
  - phase: 17-return-type-parameter-type
    provides: "One-argument call contracts with distinct parameter and return data types"
provides:
  - "Typed arm value places for computed matches whose return type differs from the scrutinee"
  - "Independent core admission, interpreter prefix dispatch, and native Result return coverage"
  - "Result-returning call source fixture with independent peers, four execution tiers, and diagnostic-ID comparison"
affects: [phase-18-verification, CTL-02]
actuals:
  tokens: 10824
  tasks: 3
  commits: 0
tech-stack:
  added: []
  patterns: ["Edge-bound typed branch result places", "Dispatch computed-match callees from their entry block"]
key-files:
  created: [internal/compiler/corevalidate/corevalidate_phase18_test.go]
  modified: [.planning/LANGUAGE-MATURITY.md, internal/compiler/core/core.go, internal/compiler/check/check.go, internal/compiler/check/check_phase18_test.go, internal/compiler/corevalidate/corevalidate.go, internal/compiler/interp/interp.go, internal/compiler/interp/interp_test.go, internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen_program_test.go, internal/compiler/session/session_phase18_test.go, internal/compiler/session/session_peer_gate_test.go, testdata/phase18/result_computed_match.lang, testdata/phase16/public-emitter-consumers.json]
key-decisions:
  - "Represent a bare branch result with an explicit MatchArm.ValuePlaceID when the Result type differs from the match scrutinee type."
  - "Use the existing function return ABI while deriving match and local C types from checked place types."
requirements-completed: []
coverage:
  - id: D1
    description: "Resource-to-Result computed-match calls are independently admitted and return Accepted through interpreter and native optimization tiers."
    requirement: CTL-02
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check -run 'TestPhase18ResultComputedMatchChecker|TestPhase18ComputedScrutinee' -count=1"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/cgen -run 'TestPhase18ResultArmValuePlace|TestPhase18ComputedScrutinee|TestPhase18CallComputedMatchPrefix|TestCallFromBothMatchArmsAcrossFrames|TestPhase17ProgramTwoType' -count=1"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase18ResultComputedMatchNativeReturn|TestPhase18ResultComputedMatchAdmission|TestPhase18ResultComputedMatch|TestPhase18FiveAxis' -count=1"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestPhase18' -count=1"
        status: pass
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-gocache go test ./..."
        status: pass
    human_judgment: false
duration: 45min
completed: 2026-09-24
status: complete
---

# Phase 18 Plan 04: Result Call Computed Match Summary

**A Result-returning callee now carries its selected alternative through checked core, interpreter execution, and native code emission.**

## Performance

- **Duration:** approximately 45 minutes
- **Tasks:** 3
- **Files modified:** 14

## Accomplishments

- Added `MatchArm.ValuePlaceID` for a selected arm's explicit return-typed value, and made the checker derive the caller's match type from the computed call target.
- Extended independent core validation to check the arm alternative, unique Result-typed place, and matching `OpReturn`; added forged-source, alternative, place, and type controls.
- Made called computed-match functions execute the entry prefix before arm dispatch, including one-arm branches, then bind the selected arm value.
- Updated native emission to switch and declare locals using checked place types while retaining the callee's Resource parameter and Result return ABI.
- Added source-to-native and five-axis evidence; interpreter, `-O0`, `-O3`, and `-O3 -flto` all return `Accepted` for `Raw`.
- Updated the machine-checked corpus line count, computed-match probe inputs in the cross-corpus peer gate, and the public emitter inventory after adding the native regression.
- Re-ran the complete Go suite successfully after these integration updates.

## Task Commits

Task commits are deferred to the parent GSD commit helper as requested. No commits were created by this executor.

## Files Created/Modified

- `internal/compiler/core/core.go` — explicit arm return place identity.
- `internal/compiler/check/check.go` and `check_phase18_test.go` — computed match typing and refusal controls.
- `internal/compiler/corevalidate/corevalidate.go` and `corevalidate_phase18_test.go` — independent typed arm admission and forgery rejection.
- `internal/compiler/interp/interp.go` and `interp_test.go` — entry-prefix execution and selected value binding.
- `internal/compiler/cgen/cgen_program.go` and `cgen_program_test.go` — typed native switching, values, and direct compilation regression.
- `internal/compiler/session/session_phase18_test.go` and `testdata/phase18/result_computed_match.lang` — peer, four-tier, and diagnostic-ID acceptance evidence.
- `internal/compiler/session/session_peer_gate_test.go` — computed-match fixtures are probed using their function parameter alternatives.
- `testdata/phase16/public-emitter-consumers.json` — updated source-derived emitter locations and added native consumer.
- `.planning/LANGUAGE-MATURITY.md` — corrected derived corpus line count.

## Decisions Made

- The arm value is edge-bound and initialized when that arm is selected; no extra source operation is introduced.
- The fixture retains the single-argument `Resource -> Result` call contract from Phase 17.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Corrected peer initialization placement for edge-bound arm values**
- **Found during:** Task 2
- **Issue:** Core replay initialized the new value in straight-line replay, so valid branch `OpReturn` sources remained uninitialized.
- **Fix:** Initialize the arm value place in branch replay before validating its operations.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** Core peer and forged-return regressions pass.

**2. [Rule 1 - Bug] Dispatch single-arm computed callees and bind selected return value**
- **Found during:** Task 2
- **Issue:** Interpreter dispatch only selected match edges with multiple successors, skipping the single-arm producer's edge-bound value.
- **Fix:** Run match edge selection for a computed-match entry block even when it has one successor, and bind the arm's declared value place.
- **Files modified:** `internal/compiler/interp/interp.go`
- **Verification:** `TestPhase18CallComputedMatchPrefix` returns `Accepted`.

**3. [Rule 1 - Bug] Emit copies using their checked source type**
- **Found during:** Task 2
- **Issue:** Native code declared Result call-target copies using the branch parameter's Resource enum type.
- **Fix:** Select the operation's native branch type from its checked `TypeID`.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** Direct Clang `-O0` regression and full tier comparison pass.

**4. [Rule 1 - Bug] Use valid input probes for computed-match corpus fixtures**
- **Found during:** Full-suite verification
- **Issue:** The corpus gate passed computed arm labels as function inputs even when the function parameter had a different ADT type.
- **Fix:** Probe computed-match functions with alternatives from their declared parameter type.
- **Files modified:** `internal/compiler/session/session_peer_gate_test.go`
- **Verification:** The corpus peer test and full suite pass.

**5. [Rule 3 - Blocking issue] Refresh derived corpus count and emitter inventory**
- **Found during:** Full-suite verification
- **Issue:** The added fixture and native emitter regression made the recorded line count and source-call registry stale.
- **Fix:** Update the corpus count from 4,578 to 4,582 and refresh affected emitter call locations, including the new consumer.
- **Files modified:** `.planning/LANGUAGE-MATURITY.md`, `testdata/phase16/public-emitter-consumers.json`
- **Verification:** The maturity count and emitter inventory tests pass.

Task-level TDD commits were not created because this child executor was explicitly instructed to leave commits to the parent helper. Focused tests and the full suite were run after implementation.

## Issues Encountered

The first native compile correctly rejected enum conversions between Resource and Result. Type selection now follows checked place facts, and the compiler accepts the generated C.

## Next Phase Readiness

Plan acceptance evidence is complete. CTL-02 remains open pending full Phase 18 verification, as directed.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-24*

## Self-Check: PASSED

- Summary file exists at the required plan path.
- All four required verification commands passed.
- The complete `GOCACHE=/tmp/ai-lang-gocache go test ./...` suite passed after fixing the reported integration regressions.
- Task code and summary remain uncommitted as directed; the parent GSD helper will commit the scoped paths.

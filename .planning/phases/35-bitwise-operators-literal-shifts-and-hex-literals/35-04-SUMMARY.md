---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "04"
subsystem: compiler-validation
tags: [Go, core-ir, scalar-validation, bitwise, literal-shifts]

requires:
  - phase: 35-03
    provides: Registered U64 scalar operations, canonical literal shift immediates, and the narrow unused-Byte source bridge
provides:
  - Independent raw-core shift and unused-Byte laws in corevalidate and originvalidate
  - Independent core and origin transfer checks for all five scalar operations through their analysis routes
  - Reached and restored shape, type, immediate, trigger, and wrong-result controls for both validator peers
affects: [35-05, 35-06, 35-07, corevalidate, originvalidate]

actuals:
  tokens: 21818
  tasks: 3
  commits: 4
commits: 4
plan_head_before: e072741f121747c12eb485e0cb02d5cbe1ebf22a
plan_head_after: 727780c3070ac0d68d37d8e6688585e5dbd9e533

tech-stack:
  added: []
  patterns:
    - Each validator peer derives raw shape, types, immediate range, and transfer facts independently.
    - Transfer mutants must reach a peer-specific safety sink or observation, then pass again after restoration.

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_phase35_test.go
    - internal/compiler/originvalidate/originvalidate_phase35_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go

key-decisions:
  - "Keep corevalidation and origin validation as separate proof authorities; neither relies on source-token provenance or the other peer's transfer implementation."
  - "Admit only canonical decimal shift counts from 0 through 63 with empty RightID, and permit Byte solely as an unused formal in the narrow U64 bridge."
  - "Use reached static safety sinks for validator transfer faults; do not add execution-result replay or runtime events."

patterns-established:
  - "Exercise pure scalar operations through straight-line and CFG/fixed-point triggers, including publication and per-return recomputation in origin validation."
  - "Keep existing OpConst and Byte linear contracts intact while independently rejecting malformed new raw-core forms."

requirements-completed: [BIT-01, BIT-02]
coverage:
  - id: D1
    description: Both validator peers independently admit valid raw shift immediates and reject malformed shift and used-Byte raw IR.
    requirement: BIT-02
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s"
        status: pass
    human_judgment: false
  - id: D2
    description: Core validation reaches all five independent value transfers and its shape/type/range faults through straight-line and CFG analysis routes.
    requirement: BIT-01
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission|TestPhase19OpConstIndependentAdmission|TestPhase19ForgedOpConstFactsRefused)$' -count=1 -v -timeout=600s"
        status: pass
    human_judgment: false
  - id: D3
    description: Origin validation independently reaches all five transfers through publication and return recomputation, with no-remainder trigger and restored-fault controls.
    requirement: BIT-01
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/originvalidate -run '^(TestPhase35ShiftImmediateAdmission|TestPhase35ScalarTransferMutants|TestPhase35AnalysisTriggers|TestPhase35UnusedByteIRAdmission)$' -count=1 -v -timeout=600s"
        status: pass
    human_judgment: false

duration: 1h 04m
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 04: Independent Scalar IR Validation Summary

**Core and origin validation now independently enforce raw U64 shift and unused-Byte laws and derive all five scalar transfers through their own analysis paths.**

## Performance

- **Duration:** 1h 04m, including recovery from the preserved compatibility baseline
- **Started:** 2026-10-14T15:23:11Z
- **Completed:** 2026-10-14T16:26:51Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Added independent corevalidate and originvalidate checks for canonical decimal shift counts 0 through 63, empty unrelated fields, initialized U64 sources, fresh U64 targets, and the unused-Byte bridge; existing OpConst and Byte linear behavior remains covered by compatibility controls.
- Reached XOR, AND, OR, shift-left, and shift-right transfers through core straight-line and CFG/fixed-point routes, with restored result and admission mutants.
- Routed origin validation through pure-operation triggers, publication, and per-return recomputation without relying on a remainder operation; each peer's tests observe its own facts or named safety sink.

## Verification Evidence

All invocations used `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, verbose output, and a 600-second timeout. The selected package-qualified top-level executions all printed RUN and PASS:

- Task 1 tracer: two Phase 35 selectors in each validator package (4 executions); rerun after its commit.
- Task 2 core selector: four Phase 35 selectors and two Phase 19 OpConst compatibility selectors (6 executions).
- Task 3 origin selector: four Phase 35 selectors (4 executions).
- Recovery compatibility baseline: `TestPhase35CorePeerBaseline` and `TestPhase35OriginPeerBaseline` (2 executions), freshly rerun during closeout.

Total: 16 package-qualified top-level test executions across the plan's focused selectors and baseline check.

## Task Commits

1. **Compatibility baseline preserved during recovery** - `a929f41f` (`test(35-04): pin validator compatibility baseline`)
2. **Task 1: Validate one raw literal shift and unused-Byte bridge through both peers** - `26e4b509` (`feat(35-04): admit raw U64 shifts and unused Byte bridge`)
3. **Task 2: Complete core field/range laws and all scalar transfer routes** - `18461724` (`test(35-04): cover core scalar transfer routes`)
4. **Task 3: Complete origin laws, exact transfers and publication reachability** - `727780c3` (`test(35-04): cover origin transfer publication routes`)

**Plan metadata commit:** recorded separately after state and roadmap updates.

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - independent raw shift/Byte admission and five-operation transfer laws across core analysis paths.
- `internal/compiler/corevalidate/corevalidate_phase35_test.go` - core admission, route, trigger, and restored-mutant controls.
- `internal/compiler/originvalidate/originvalidate.go` - independent origin transfer, trigger, publication, and return-recomputation paths.
- `internal/compiler/originvalidate/originvalidate_phase35_test.go` - origin admission, publication, trigger, and restored-mutant controls.

## Decisions Made

- Kept validator proof authority independent from source AST provenance, checker admission, and the other validator peer.
- Retained the D-18 boundary: these validators establish value facts and obligations; executionpeer remains a causal checker, with no result replay or new runtime event.
- Used the external `originvalidate_test` package for tests that call the exported API and import `check`, avoiding an import cycle.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Corrected `ParseUint` error handling in the inherited core fixed-point transfer.**
- **Found during:** Task 1 tracer verification
- **Issue:** The second return from `strconv.ParseUint` is an `error`, but the inherited fixed-point admission path treated it as a boolean, preventing the validators from compiling.
- **Fix:** Check the parse error explicitly while retaining canonical decimal and range checks.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** Task 1 tracer selector passed in both packages and was rerun after commit.
- **Committed in:** `26e4b509`

**Total deviations:** 1 auto-fixed blocking issue. No semantic scope expansion.

## Issues Encountered

The origin test package uses the external test package because importing `check` from an internal `originvalidate` test would create a package import cycle. Tests exercise only the exported publication and recomputation APIs.

## User Setup Required

None - no external service or package setup is required.

## Next Phase Readiness

The two validator authorities now provide focused independent evidence for the downstream execution and native plans. Phase 35-04 does not claim interpreter, native, public conformance, or whole-phase completion.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*

## Self-Check: PASSED

- Summary file exists at the required plan path.
- Compatibility baseline and all three task commits are ancestors of the current `HEAD`.
- The committed-source stub-pattern scan found no TODO, FIXME, placeholder, or coming-soon stub markers in the four plan-owned source/test files.

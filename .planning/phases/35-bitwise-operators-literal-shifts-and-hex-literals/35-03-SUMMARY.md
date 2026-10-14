---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "03"
subsystem: compiler
tags: [go, scalar-checker, core-ir, bitwise, literal-shifts]
requires:
  - phase: 35-02
    provides: grouped scalar grammar, hex numeric admission, and frozen compatibility boundaries
provides:
  - Five registered U64 bitwise and literal-shift core operations with stable schemas
  - Canonical decimal shift immediates restricted to literal values 0 through 63
  - Structured scalar admission for an unused Byte formal with U64 and Bool locals
  - Independent checker transfer controls for all five operations
affects: [35-04, 35-05, 35-06, 35-07, 35-08, scalar-core, bitwise]
actuals:
  tokens: 10225.75
  tasks: 3
  commits: 3
commits: 3
plan_head_before: 783a6c101d905c60a2a14e1e03ff49da989ef85c
plan_head_after: e412841f90fb2a3276643d92392b6498fb1f7aee
tech-stack:
  added: []
  patterns:
    - Independent fixed-point transfers are observed through a static zero-remainder safety sink.
    - Byte input admission scans the full structured source body before lowering and leaves the formal unused in core operations.
key-files:
  created:
    - internal/compiler/core/core_phase35_test.go
    - internal/compiler/check/check_phase35_test.go
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/scalar.go
key-decisions:
  - "Keep shifts immediate-only: grouped number leaves become canonical decimal ConstU64 values and never produce RHS places."
  - "Admit Byte only for the new structured U64-returning bridge after a full-body no-use prepass; preserve statement-free legacy Byte routing and existing U64/ByteView type-fact order."
  - "Keep checker value transfers independent from the producer; transfer tests use checker-owned facts and a static safety sink, without execution-result replay."
patterns-established:
  - "A source refusal must remain singular when expression lowering has already reported the more specific operator cause."
  - "Fault copied core observations serially, require the named sink verdict to change, then rerun the unmodified observation."
requirements-completed: [BIT-01, BIT-02, BIT-03]
coverage:
  - id: D1
    description: Five closed core operation kinds are registered while existing schema, IDs, and ConstU64 serialization remain stable.
    requirement: BIT-01
    verification:
      - kind: unit
        ref: internal/compiler/core/core_phase35_test.go#TestPhase35CoreShape
        status: pass
      - kind: unit
        ref: internal/compiler/core/core_test.go#TestAllOperationKindsRegistered
        status: pass
    human_judgment: false
  - id: D2
    description: Grouped literal shifts lower to canonical immediates; dynamic, computed, negative, and out-of-range amounts retain their owning refusal.
    requirement: BIT-02
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase35_test.go#TestPhase35ShiftAdmission
        status: pass
      - kind: unit
        ref: internal/compiler/check/check_phase35_test.go#TestPhase35NumericRoutePolicy
        status: pass
    human_judgment: false
  - id: D3
    description: The unused-Byte witness admits U64 and Bool locals while rejecting Byte reads, copies, stores, builtins, and unreachable uses; the legacy Byte literal route remains stable.
    requirement: BIT-01
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase35_test.go#TestPhase35UnusedByteSourceAdmission
        status: pass
    human_judgment: false
  - id: D4
    description: Each bitwise or shift transfer reaches an independent checker fact and discriminating copied-core fault control.
    requirement: BIT-01
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase35_test.go#TestPhase35ScalarTransferMutants
        status: pass
    human_judgment: false
  - id: D5
    description: Existing numeric route conversion and malformed-token ownership remain pinned.
    requirement: BIT-03
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase35_numeric_test.go#TestPhase35NumericRoutePolicy
        status: pass
    human_judgment: false
duration: 11min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 03: Literal Shifts, Bitwise Core, and Narrow Byte Bridge Summary

**Five U64 bitwise and shift operations now lower to stable core, with literal-only shifts, unused-Byte admission, and independent checker controls.**

## Performance

- **Duration:** 11 minutes
- **Started:** 2026-10-14T15:04:39Z
- **Completed:** 2026-10-14T15:16:07Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Registered xor, and, or, left shift, and right shift in the existing operation registry without changing core schemas or IDs. `ConstU64` remains limited to constants and the two shift immediates.
- Lowered grouped integer shift leaves to canonical decimal immediates in `0..63`, with empty `RightID`; names, computed expressions, negative syntax, and rejected ranges retain their exact owner-specific refusal.
- Added the narrow structured Byte bridge. A pre-lowering scan rejects every formal use, including nested unreachable source and implicit captured operations, while statement-free legacy Byte bindings keep their established route.
- Exercised all five checker transfers through a static zero-remainder sink. Peer-operation and shift-count faults are applied only to copied core observations and are followed by a clean rerun.

## Task Commits

1. **Task 1: Check one grouped literal shift from source through canonical core** — `dee2d8d0` (`feat`)
2. **Task 2: Admit only a provably unused Byte entry without widening scalar input** — `194a127f` (`feat`)
3. **Task 3: Reach independent checker result and literal-refusal mutants** — `e412841f` (`test`)

The summary and state metadata commit is recorded separately after closeout.

## Files Created/Modified

- `internal/compiler/core/core.go` — registered the five exact-width operators and documented their kind-specific immediate contract.
- `internal/compiler/core/core_test.go` — updated the registry test's declared operation count from 24 to 29.
- `internal/compiler/core/core_phase35_test.go` — pinned registry shape, stable schema/IDs, and serialization.
- `internal/compiler/check/check.go` — documented the structured scalar dispatch boundary while preserving legacy binding routing.
- `internal/compiler/check/scalar.go` — added operator lowering/transfers, canonical shift validation, and the strict unused-Byte source law.
- `internal/compiler/check/check_phase35_test.go` — added source admission, compatibility, numeric-owner, and checker-transfer controls.

## Decisions Made

- Shift amounts remain source-only literals. They are not lowered into ordinary constant places, and dynamic shifts do not gain runtime checks or defect behavior.
- The unused Byte bridge adds a Byte type fact only on its new structured route. Existing U64 and ByteView type-fact order and statement-free Byte behavior remain unchanged.
- Checker facts are verified by the checker fixed point. The execution peer's causal role is not extended into result replay, and this plan makes no native or interpreter result claim.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical] Updated the stale registry test count**
- **Found during:** Task 1
- **Issue:** Adding the five closed operation kinds made the existing registry selector's hard-coded count of 24 stale.
- **Fix:** Updated only the expected count to 29 and its corresponding comment.
- **Files modified:** `internal/compiler/core/core_test.go`
- **Verification:** `TestAllOperationKindsRegistered` passed in the exact core/check selector.
- **Committed in:** `dee2d8d0`

**2. [Rule 1 - Bug] Preserved singular shift-source refusals**
- **Found during:** Task 1
- **Issue:** A dedicated unsupported shift amount also fell through to the generic scalar-binding refusal.
- **Fix:** Skip generic binding validation after expression lowering has already emitted a specific diagnostic.
- **Files modified:** `internal/compiler/check/scalar.go`, `internal/compiler/check/check_phase35_test.go`
- **Verification:** `TestPhase35ShiftAdmission` requires exactly one diagnostic with the registered hint.
- **Committed in:** `dee2d8d0`

**3. [Rule 1 - Bug] Made the excluded Byte/Unit signature diagnostic precise**
- **Found during:** Task 3
- **Issue:** The new Byte-to-Unit refusal inherited generic wording that said the input must be U64.
- **Fix:** Kept the signature refusal code and made its message state that the unused Byte bridge requires a U64 result; pinned that case.
- **Files modified:** `internal/compiler/check/scalar.go`, `internal/compiler/check/check_phase35_test.go`
- **Verification:** The Task 2 and Task 3 exact selectors passed after the refinement.
- **Committed in:** `e412841f`

**Total deviations:** 3 auto-fixed issues (1 Rule 2, 2 Rule 1). All were directly required for accurate source behavior or the required registry selector.

## Issues Encountered

Two initial negative fixtures used constructs the parser itself rejects (`if` without `else` and a terminal defect in the structured scalar grammar). They were replaced with a parser-valid `if/else` case and a synthetic AST control for the implicit defect source; the required selector then passed.

## User Setup Required

None — no external service configuration is required.

## Next Phase Readiness

The core and checker contracts are ready for dependent Phase 35 work. Runtime, native C, and independent peer behavior remain owned by later plans and were not asserted here. Only the plan's named focused Go selectors ran; no whole-package or native suite was run.

## Self-Check: PASSED

- Summary file exists at the plan's canonical path.
- Task commits `dee2d8d0`, `194a127f`, and `e412841f` are ancestors of HEAD.
- The plan ledger measures three task commits from `783a6c101d905c60a2a14e1e03ff49da989ef85c` through `e412841f90fb2a3276643d92392b6498fb1f7aee`.
- No stubs, skipped tests, or unrun plan verifications were introduced.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*

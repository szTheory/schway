---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "02"
subsystem: compiler
tags: [borrow-checking, independent-validation, loan-liveness, pointer-successors]
plan_head_before: d97b573d8ef59eda3c1b010c7e1b9899e2f43757
commits: 5

# Dependency graph
requires:
  - phase: 25-01
    provides: Checked shared U64 borrow/copy core witness for production pointer lowering.
provides:
  - Core and origin peers independently accept shared and exclusive read/copy families, compatible shared readers, and sequential shared-to-exclusive use.
  - Core and origin peer controls reject live shared/exclusive conflicts and origin/access-family mutations.
  - Path oracle rejects unknown terminal source places and recomputes endpoints without consuming declared endpoint claims.
affects: [25-03, pointer-successors, borrow-validation]

# Actuals (#2632)
actuals:
  tokens: 3981
  tasks: 2
  commits: 5

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Origin validation derives loan owners, access families, and last uses from its own linear operations.
    - Path replay validates terminal source-place membership before deriving loan liveness.

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_pointer_successor_test.go
    - internal/compiler/originvalidate/originvalidate_pointer_successor_test.go
    - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/pathoracle/pathoracle.go

key-decisions:
  - "Keep corevalidate, originvalidate, and pathoracle as independent derivations; the peer tests feed each validator core operations directly."
  - "Derive bounded linear origin-peer conflicts from the peer's own loan chains and last uses, allowing compatible reborrows and ended shared loans."
  - "Reject unknown terminal source places in path replay so a forged source cannot erase a loan's last use."

patterns-established:
  - "Peer mutation controls alter core family/access facts after checker admission, then ask each validator to derive its own result."
  - "Path endpoint tests mutate serialized endpoint claims and confirm pathoracle recomputes from the function CFG and operations."

requirements-completed: []
coverage:
  - id: D1
    description: Independent core and origin derivations distinguish compatible shared loans, sequential shared-to-exclusive loans, live conflicts, and family/access mutations.
    requirement: NAT-11
    verification:
      - kind: unit
        ref: "go test -count=1 -run '^TestPhase25(PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate"
        status: pass
    human_judgment: false
  - id: D2
    description: Path oracle independently validates source-place references and derives loan endpoints without trusting serialized endpoint claims.
    requirement: NAT-13
    verification:
      - kind: unit
        ref: "go test -count=1 -run '^TestPhase25PointerPath' ./internal/compiler/pathoracle"
        status: pass
    human_judgment: false

# Metrics
duration: 13min
completed: 2026-10-01
status: complete
---

# Phase 25 Plan 02: Independent Pointer-Family Validation Summary

**Core, origin, and path peers now derive pointer-family conflicts and loan boundaries independently from core operations.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-10-01T20:04:00Z
- **Completed:** 2026-10-01T20:17:12Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added peer tests for compatible shared readers, sequential shared-to-exclusive use, a live incompatible access, an undeclared returned loan, and a family/access-mode mutation.
- Added originvalidate's own linear loan owner, access-family, and last-use derivation. It permits valid reborrows and sequential access while rejecting overlapping shared/exclusive loans.
- Hardened pathoracle against a forged source place that would otherwise silently lose its inherited loan, and proved endpoint claims are recomputed from operations and CFG paths.

## Task Commits

1. **Task 1: Derive family and origin obligations independently in core peers** - `e081b70` (test), `64526f5` (test), `6c89a17` (fix)
2. **Task 2: Derive shared and exclusive loan lifetimes along paths** - `0005407` (test), `0ae7bb9` (fix)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate_pointer_successor_test.go` - shared/exclusive core family and conflict controls.
- `internal/compiler/originvalidate/originvalidate_pointer_successor_test.go` - origin, escape, compatible-family, and conflict controls.
- `internal/compiler/originvalidate/originvalidate.go` - independent linear loan conflict derivation.
- `internal/compiler/pathoracle/pathoracle_pointer_successor_test.go` - endpoint and forged-source path controls.
- `internal/compiler/pathoracle/pathoracle.go` - source-place validation in independent path replay.

## Decisions Made

- Kept the three peer implementations independent and used checked core only as their common input.
- Kept requirement IDs pending because the remaining Phase 25 plans still own source admission, production emission, native receipts, and integrated utility acceptance.
- Added no dependencies, consistent with D-25-08.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Reject unknown source places during path replay**
- **Found during:** Task 2 (Derive shared and exclusive loan lifetimes along paths)
- **Issue:** An unknown terminal source was treated as carrying no loans, allowing malformed core to hide the last use from the independent oracle.
- **Fix:** Index each function's own places and reject non-constant operations that name an unknown source place before replaying loan state.
- **Files modified:** `internal/compiler/pathoracle/pathoracle.go`, `internal/compiler/pathoracle/pathoracle_pointer_successor_test.go`
- **Verification:** `go test -count=1 -run '^TestPhase25PointerPath' ./internal/compiler/pathoracle`
- **Committed in:** `0ae7bb9`

**Total deviations:** 1 auto-fixed (Rule 1)
**Impact on plan:** The fix closes the path oracle gap directly exposed by the required retained/forged-loan control; scope remains within the plan's bounded path validation.

## TDD Notes

- Task 1's initial peer controls passed against existing corevalidate behavior. The additional origin conflict control failed before the origin peer derivation was added, then passed after the fix.
- Task 2's forged-source control failed before the path source-place guard, then passed after the fix.

## Issues Encountered

- The origin peer initially treated a valid exclusive reborrow as a conflict. The derived family walk was refined to allow reborrows through an existing exclusive loan while still rejecting an exclusive child of a live shared loan.

## User Setup Required

None.

## Self-Check: PASSED

- All three created test files and both modified implementation files exist.
- Task commits `e081b70`, `64526f5`, `6c89a17`, `0005407`, and `0ae7bb9` exist in Git history.
- The plan commit ledger records five commits before the metadata closeout commit.

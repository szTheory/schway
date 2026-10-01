---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "03"
subsystem: compiler
tags: [ownership, borrowing, corevalidate, originvalidate, pathoracle]
requires:
  - phase: 25-02
    provides: Independent pointer-family successor facts and validation patterns
provides:
  - Distinct exclusive U64 read/copy helper and bounded transfer caller composition
  - Independent peer derivation that terminates origin propagation at copied U64 values while preserving loan liveness
  - Separate shared/exclusive conflict and exclusive escape source controls
affects: [phase-25-native-emission, pointer-family-validation]
actuals:
  tokens: 6782
  tasks: 2
  commits: 5
  plan_head_before: ac41bf83ca23075cb2c9d779e5e898d326a008c9
tech-stack:
  added: []
  patterns:
    - Narrow U64 copy boundary independently derived by core and origin validators
    - Exact source recognizer for transfer, typed use, shared copy, exclusive copy, and return
key-files:
  created:
    - internal/compiler/session/session_pointer_successor_test.go
    - testdata/phase25/exclusive_copy_accept.schway
    - testdata/phase25/exclusive_conflict_reject.schway
    - testdata/phase25/exclusive_escape_reject.schway
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_pointer_successor_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_pointer_successor_test.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_pointer_successor_test.go
key-decisions:
  - "A copied U64 is an ordinary value and ends origin propagation; the copy operation remains part of borrow access and liveness analysis."
  - "Shared and exclusive borrow contracts remain distinct, and caller recognition stays limited to the exact composed source shape."
patterns-established:
  - "Each independent validator derives the U64 copy boundary from its own core facts."
  - "Conflict, escape, and sequential-ended-loan witnesses have separate source outcomes."
requirements-completed: [NAT-12, NAT-13, DX-15]
coverage:
  - id: D1
    description: "Exclusive U64 copy helper composes after transfer and typed fallible use, with both helper results feeding the final return."
    requirement: NAT-12
    verification:
      - kind: unit
        ref: "go test -count=1 -run '^TestPhase25(ExclusiveSource|TransferCallerComposition)' ./internal/compiler/check"
        status: pass
      - kind: unit
        ref: "go test -count=1 -run '^TestPhase25(U64CopyOrigin|PointerFamily|PointerOrigin)' ./internal/compiler/corevalidate ./internal/compiler/originvalidate"
        status: pass
    human_judgment: false
  - id: D2
    description: "Shared and exclusive conflicts, exclusive escape, and independent sequential-loan validation have distinct outcomes."
    requirement: DX-15
    verification:
      - kind: integration
        ref: "go test -count=1 -run '^TestPhase25(FamilyConflict|FamilyEscape|IndependentPeer)' ./internal/compiler/session"
        status: pass
    human_judgment: false
duration: 45m
completed: 2026-10-01
status: complete
---

# Phase 25 Plan 03: Separate Pointer Successors and Integrated Utility Summary

The checker admits an exclusive read/copy helper and the bounded transfer → typed use → shared copy → exclusive copy flow, while copied U64 values stop carrying borrow origin and continue to participate in loan liveness checks.

## Performance

- **Duration:** 45m
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments

- Added a distinct `borrow mut value` source helper that returns a copied U64, and extended the exact caller recognizer to require transfer acquisition, fallible typed use, shared helper call, exclusive helper call, and final return in order.
- Refined corevalidate and originvalidate independently so an exact Copy-capable U64 copy ends origin propagation without removing the copy use from conflict/liveness analysis.
- Added separate shared/exclusive conflict witnesses, an exclusive publication escape witness, and independent corevalidate, originvalidate, and pathoracle checks for the sequential ended-loan positive case.

## Task Commits

1. **Task 1 TDD red:** `014c2d1` — specify exclusive source and caller composition.
2. **Task 1 peer TDD red:** `1ba5e20` — specify copied-U64 origin boundary.
3. **Task 1 implementation:** `0c820e3` — admit exclusive U64 copy composition.
4. **Task 2:** `35b10de` — pin exclusive conflicts and escapes.

The plan was amended during execution by the parent workflow in `3b27857`; this is included in the measured `actuals.commits` count. The final metadata commit is created after summary verification.

## Decisions Made

- A copied U64 is an ordinary value with no propagated borrow origin. Its OpCopy still extends the source loan’s access/liveness, so conflict detection continues through the copy.
- The caller recognizer remains bounded to the exact transfer/use/shared/exclusive/return composition, and shared and exclusive borrow contracts remain distinct.

## Issues Encountered

- The first escape witness returned a U64, which is copied by value and correctly did not represent a borrowed escape. The fixture was tightened to return a Buffer borrow; `CheckCommandFile` then refused publication with `core.origin_omitted`.
- That origin peer diagnostic currently has a zero primary span. Conflict diagnostics do carry source spans and causes, which the session tests assert. The zero-span peer diagnostic does not indicate source-checker attribution loss; it is the current publication-peer diagnostic surface.

## Deviations from Plan

None. The peer-refinement files were added to the authorized plan scope by the parent plan amendment before implementation.

## User Setup Required

None.

## Next Phase Readiness

The source/checker and independent-peer evidence is ready for the next Phase 25 plan to verify native helper emission and integrated runtime behavior.

---
*Phase: 25-separate-pointer-successors-and-integrated-utility*
*Completed: 2026-10-01*

## Self-Check: PASSED

- Summary file exists at the plan's required path.
- Task commits `014c2d1`, `1ba5e20`, `0c820e3`, and `35b10de` are present in Git history.

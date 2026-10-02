---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "01"
subsystem: compiler
tags: [shared-borrow, pointer-abi, c17, manifest]
plan_head_before: f7fd5964bdb6cf26340e212e654a2cd6b44b4051
commits: 3

# Dependency graph
requires: []
provides:
  - An ordinary Schway U64 shared borrow/copy source witness with an explicit checked OpBorrowShared loan fact.
  - Exact-shape production C lowering to a const U64 pointer parameter and pointer-address argument.
  - A pointer manifest entry derived from the same checked ABI fact, with no emitted optimizer attributes.
affects: [25-02, cgen, native-emission, pointer-abi]

# Actuals (#2632)
actuals:
  tokens: 6391
  tasks: 2
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Keep pointer parameter, call argument, body copy, and manifest metadata tied to one exact checked ABI fact.
    - Refuse wider pointer chains before production C serialization.

key-files:
  created:
    - internal/compiler/check/check_pointer_successor_test.go
    - internal/compiler/cgen/cgen_pointer_successor_test.go
    - internal/compiler/native/phase25_pointer_successor_test.go
    - testdata/phase25/shared_copy_accept.schway
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go

key-decisions:
  - "The existing checker already records shared family identity as OpBorrowShared with a loan ID, so this plan adds a source witness instead of changing general checker admission."
  - "Only the exact U64 shared-borrow-to-return shape receives the const uint64_t pointer ABI; unsupported chains remain refused."
  - "The C parameter, call-site address, scalar copy, and manifest pointer entry use one checked ABI fact; no restrict, noalias, capture, alignment, or ownership promise is emitted."

requirements-completed: []
coverage:
  - id: D1
    description: Ordinary shared U64 borrow/copy source is accepted with family-specific checked facts.
    requirement: NAT-11
    verification:
      - kind: unit
        ref: "go test -count=1 -run '^TestPhase25Shared(PointerSuccessor|Source)' ./internal/compiler/check"
        status: pass
    human_judgment: false
  - id: D2
    description: Exact shared helper lowers to a const pointer C parameter, matching address argument, and matching manifest entry with no emitted attributes.
    requirement: NAT-13
    verification:
      - kind: integration
        ref: "go test -count=1 -run '^TestPhase25Shared(PointerABI|PointerManifest|NativeShape)' ./internal/compiler/cgen ./internal/compiler/native"
        status: pass
    human_judgment: false

---

# Phase 25 Plan 01: Separate Pointer Successors and Integrated Utility Summary

**A checked shared U64 read/copy helper now emits a `const uint64_t *` ABI and matching manifest entry without unsupported attribute promises.**

## Performance

- **Duration:** 12 min
- **Completed:** 2026-10-01
- **Tasks:** 2
- **Files modified or created:** 6

## Accomplishments

- Added a source fixture and checker assertion proving the ordinary helper carries an explicit shared-borrow operation and loan ID.
- Lowered the exact shared shape to a pointer parameter, passed the caller value by address, and copied through the pointer in generated C.
- Added a manifest pointer entry from the same ABI fact and verified the emitted attribute list stays empty.
- Added a negative control proving a wider shared-borrow chain is refused before C serialization.
- Confirmed the emitted C compiles under installed Clang in C17 syntax-check mode.

## Task Commits

1. **Task 1: Admit the checked shared read/copy source witness** - `4220d98` (test)
2. **Task 2 RED: Specify shared pointer ABI contract** - `6826ffe` (test)
3. **Task 2 GREEN: Lower checked shared U64 helper by pointer** - `65bfadc` (feat)

## Decisions Made

- The existing checker already records the shared family as `OpBorrowShared` with a loan ID, so the plan adds a positive source witness without widening generic checker rules.
- Pointer lowering is limited to the exact U64 borrow-and-return shape; a longer chain remains outside the serializer's admitted subset.
- The pointer declaration, argument form, dereference copy, and manifest record are derived from one checked ABI fact. The record claims shared access only and carries no optimizer or ownership promise.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The broader legacy foreign manifest contains empty/unverified fields named `capture` and similar obligations. The focused assertion checks the new pointer entry's exact fields and empty emitted-attribute list, so it distinguishes historical manifest fields from new ABI promises.
- NAT-11 and NAT-13 also appear in later Phase 25 plans and their full acceptance criteria are not complete here, so they remain pending in REQUIREMENTS.md despite the partial plan citations.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The shared helper ABI pattern is available for the separately checked exclusive family and later integrated utility work.
- Independent conflict/escape evidence and the integrated native utility remain owned by later Phase 25 plans.

## Self-Check: PASSED

- All six created or modified implementation and evidence files exist.
- Task commits `4220d98`, `6826ffe`, and `65bfadc` exist in Git history.
- The changed files contain no TODO, FIXME, placeholder, coming-soon, or not-available stubs.

---
*Phase: 25-separate-pointer-successors-and-integrated-utility*
*Completed: 2026-10-01*

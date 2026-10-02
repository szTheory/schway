---
phase: 23-live-local-allocation-and-discharge
plan: 02
subsystem: compiler-validation
tags: [go, ownership, local-allocation, file-byte, independent-checkers]

requires: []
provides:
  - Core, origin, and path peers seed local owner obligations from successful acquire operations.
  - Borrowed use preserves the acquired owner obligation, including when the use operation carries an error outcome.
  - Exact consuming releases discharge only their matching acquired owners; malformed and incomplete lifecycles are rejected.
affects: [phase-23, FFI-03, RES-04, RES-09]

actuals:
  tokens: 10268
  tasks: 2
  commits: 3
  plan_head_before: 581754eea346360ace666787398a95d1a5bafd7d

tech-stack:
  added: []
  patterns:
    - "Independent peers derive owner obligations from each operation's checked acquire facts."
    - "Reached source mutations exercise lifecycle and per-operation contract rejection."

key-files:
  created: []
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/pathoracle_test.go

key-decisions:
  - "Keep each peer's owner-lifecycle validation independently authored while enforcing the same narrow PathToken-to-U64 operation contract."
  - "Tie borrow and release obligations to the successful acquire identity and exact resource place, rather than sibling symbol names."

requirements-completed: [FFI-03, RES-04, RES-09]

coverage:
  - id: D1
    description: "Core, origin, and path peers seed and discharge local owner obligations from each acquire and exact consuming release."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^TestPhase23' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Reached lifecycle and per-operation mutations are rejected by all three independent peers."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run '^TestPhase23ResourceMutation' -count=1"
        status: pass
    human_judgment: false

duration: 17min
completed: 2026-09-28
status: complete
---

# Phase 23 Plan 02: Acquisition-Seeded Owner Validation Summary

**Core, origin, and path peers now seed FileByteOwner obligations from each acquire operation and reject tested omissions, mispairings, and per-operation contract mutations.**

## Performance

- **Duration:** 17 min
- **Started:** 2026-09-28T03:03:07Z
- **Completed:** 2026-09-28 03:20:27Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Made each independent peer validate the fixed acquire, borrowed use, consuming release, and returned-value contract.
- Preserved owner obligations through borrowed use, including the `UseError` outcome, and required exact matching release discharge.
- Added reached mutation controls for omitted, premature, duplicate, wrong-owner, fabricated, and contract-mismatched releases and operations.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add acquisition-seeded owner lifecycle validation** - `bb8dbfc` (test), `7626e83` (feat)
2. **Task 2: Add reached resource lifecycle mutation controls** - `0c915d2` (test)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - Enforces the operation-specific local owner lifecycle in the core peer.
- `internal/compiler/corevalidate/corevalidate_test.go` - Adds acquisition behavior and mutation controls.
- `internal/compiler/originvalidate/originvalidate.go` - Independently checks owner identity, use, release, and return origin.
- `internal/compiler/originvalidate/originvalidate_test.go` - Adds origin-peer lifecycle and mutation controls.
- `internal/compiler/pathoracle/pathoracle.go` - Independently validates the supported straight-line local owner path.
- `internal/compiler/pathoracle/pathoracle_test.go` - Adds path-peer lifecycle and mutation controls.

## Decisions Made

- Each peer retains its own validation logic while enforcing the same narrow operation contract.
- A local borrow's scalar result does not replace or discharge the resource owner; only the matching consume operation does.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

The mutation controls were already rejected after Task 1's validator changes, so Task 2 required no additional production patch. All targeted mutation tests passed.

## User Setup Required

None - no external service configuration required.

## Requirements

The plan declares `FFI-03`, `RES-04`, and `RES-09`. The shared-ID readiness check reported 0/3 ready because sibling plans in this phase also declare these requirements; `REQUIREMENTS.md` remains unchanged for the orchestrator's phase-level update.

## Self-Check: PASSED

- All six modified source and test files exist.
- All three task commits are present, and the measured plan commit count is 3.
- Shared `.planning/STATE.md` and `.planning/ROADMAP.md` are unchanged.
- `git diff --check` passes.

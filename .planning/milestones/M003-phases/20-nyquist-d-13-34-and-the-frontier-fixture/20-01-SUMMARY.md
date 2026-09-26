---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "01"
subsystem: testing
tags: [frontier-fixture, diagnostics, sha256, session-check]
requires: []
provides:
  - Refused checksum-intent fixture bound to an exact source digest and current checker diagnostic
  - Historical M003-open diagnostic comparison for identical bytes, documented as a reconstruction
affects: [phase-20, checksum-frontier, qlt-11]
actuals:
  tokens: 1268
  tasks: 2
  commits: 2
  plan_head_before: 86338f049e17044763c08b0828d747b5f4d6f71e
tech-stack:
  added: []
  patterns: ["Pin rejected compiler fixtures by digest, executable intent witnesses, and structured diagnostic code/span"]
key-files:
  created:
    - examples/checksum.lang
    - internal/compiler/session/session_phase20_test.go
  modified:
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-CHECKSUM-BASELINE.md
key-decisions:
  - "Treat the M003-open result as a historical reconstruction, since no contemporaneous checksum fixture or pin existed."
patterns-established:
  - "A frontier fixture must assert the exact checked-in path, digest, intent witnesses, and first production diagnostic."
requirements-completed: [QLT-11]
coverage:
  - id: D1
    description: "Checksum-intent source remains refused at a stable current diagnostic, and its refusal is later than the reconstructed M003-open refusal for identical bytes."
    requirement: QLT-11
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase20_test.go#TestPhase20ChecksumFrontier"
        status: pass
    human_judgment: false
duration: 5min
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 01: Checksum Frontier Fixture Summary

**A 663-byte checksum-intent Lang fixture is pinned at the current loop refusal and compared to the reconstructed M003-open numeric-literal refusal.**

## Performance

- **Duration:** 5 min
- **Started:** 2026-09-25T02:52:45Z
- **Completed:** 2026-09-25T02:57:45Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added `examples/checksum.lang` with SHA-256 `ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154`.
- Added `TestPhase20ChecksumFrontier`, which checks executable intent witnesses and confirms `syntax.expected_rbrace` at bytes 521–527.
- Compared against the historical reconstruction at revision `d21db90e67750bb19976c4206f4c23c68cd06207`, which reported `syntax.unexpected_byte` at bytes 512–513 for the identical bytes. The record explicitly says no original M003-open fixture or diagnostic pin existed.

## Task Commits

Each task was committed atomically:

1. **Task 1: Check in a real checksum-intent source path with its current refusal** - `8368e34` (feat)
2. **Task 2: Establish and compare the M003-open diagnostic** - `18b4582` (test)

## Files Created/Modified

- `examples/checksum.lang` - Refused future checksum program used as a frontier fixture.
- `internal/compiler/session/session_phase20_test.go` - Digest, intent, current diagnostic, and historical movement assertions.
- `.planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-CHECKSUM-BASELINE.md` - Recorded pinned comparison and evidence limit.

## Decisions Made

- The historical M003-open diagnostic is labeled as a reconstruction, not as a contemporaneous pin.
- The fixture remains refused; this plan does not change parser or checker behavior.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The repository did not yet have an `examples/` directory; created it to place the planned fixture.
- The default sandbox could not write the `.git` index. Both task commits succeeded through the authorized GSD commit command with elevated sandbox access; no unrelated files were staged.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The checksum frontier is now a deterministic, checker-backed fixture. Later plans can build on it while preserving the refusal until the planned language surface is implemented.

---
*Phase: 20-nyquist-d-13-34-and-the-frontier-fixture*
*Completed: 2026-09-25*

## Self-Check: PASSED

- All three planned artifacts exist.
- Task commits `8368e34` and `18b4582` exist in Git history.
- The focused test passed after both task commits. A later rerun was blocked by concurrent uncommitted imports in `internal/compiler/native/native.go`, owned outside this plan; those changes were left untouched.

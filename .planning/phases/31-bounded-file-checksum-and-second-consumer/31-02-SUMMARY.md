---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 02
subsystem: compiler/interp
tags: [interpreter, oracle, host-reader, adler32]
status: complete
requires:
  - phase: 31-bounded-file-checksum-and-second-consumer
    provides: 31-01 checksum.schway with the FileBytes owner shape
provides:
  - interp.RunWithFileBytes (path-keyed host reader, validated FileError alternative)
  - testsupport.ChecksumReadFile, ChecksumLimit, ChecksumOracleRows (14 D-12 rows)
  - TestPhase31ChecksumInterpreter (wikipedia, missing, unknown-alternative-refused, trailing-call-tempdir, skip-isregular-control)
affects: [31-03, 31-04, 31-06, 31-07, 31-08]
actuals:
  tokens: 14000
  tasks: 2
  commits: 2
plan_head_before: db746f479c547ecd3196cc5a8aea11a42d57db99
plan_head_after: 6c0a3530
commits: 2
tech-stack:
  added: []
  patterns:
    - independent copy of the fixture route (fileForeignCall) rather than a shared helper
key-files:
  created:
    - internal/compiler/testsupport/checksum_oracle.go
    - internal/compiler/session/session_phase31_interp_test.go
  modified:
    - internal/compiler/interp/interp.go
key-decisions:
  - "Interpreter refuses (error, not Outcome) a reader alternative that is not in the Fails data type; the fixture route's Alternatives[0] behavior is not copied (D-09)"
requirements-completed: []
---

# Phase 31 Plan 02: Interpreter file route Summary

The interpreter now reaches 300286872 for a Wikipedia file through `RunWithFileBytes`, reading via an independent Go reader (`O_NONBLOCK`, `IsRegular`, limit+1 `io.ReadAll`) that never touches the C adapter, and the D-12 oracle table is committed.

## Accomplishments

- `RunWithFileBytes` plus a `fileBytes` frame field, a dispatch branch before the fixture one, propagation to callee frames, and a `terminalLiveResources` branch. No existing fixtureBytes line was removed.
- Subtests pass: Wikipedia returns `300286872` with no live resources and exactly one release event; a missing file gives `typed_failure FileUnavailable`; an unknown alternative name is refused; the 6-operation trailing `low16` probe returns `920`; the control reader without `IsRegular` returns `1` on `/dev/null` while `ChecksumReadFile` gives `FileUnavailable`.
- `ChecksumOracleRows()` holds 14 rows with a Python command and `Python 3.14.4` on each. The 65,537-byte rows are `FileTooLarge` with Want 0, and the would-be checksum literal is not in the Go source.
- Fixture-route regressions (`Phase29|Interp`) and the inventory/groundedness tests still pass.

## Deviations from Plan

None - plan executed as written. The extra `unknown-alternative-refused` subtest covers D-09 directly.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Files exist; commits c604563b and 6c0a3530 are in HEAD.

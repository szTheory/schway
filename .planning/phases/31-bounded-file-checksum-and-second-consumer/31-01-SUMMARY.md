---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 01
subsystem: compiler/file-bytes-owner
tags: [tracer, ffi, adapter, owner-shape, adler32, native]
status: complete
requires:
  - phase: 31-bounded-file-checksum-and-second-consumer
    provides: 31-00 moved Adler-32 library and CLI in examples/checksum
provides:
  - copy-local silent bounded file adapter triple (schway_file_bytes_acquire/view/release) wired through schway.local-c/1
  - examples/checksum/checksum.schway (fn main(path: PathToken) -> U64)
  - third owner shape (FileBytes) in ability, check, corevalidate, originvalidate, pathoracle, cgen and the session build route
  - D-16 optional single trailing U64 -> U64 entry-module call admitted by every matcher
  - D-17 generated FileError.<Alternative> stderr table with exit 65
affects: [31-02, 31-03, 31-04, 31-06, 31-07, 31-08]
actuals:
  tokens: 60000
  tasks: 2
  commits: 2
plan_head_before: 0983ddd8cd4b4c1609d85a4ea4dd73e7957d6777
plan_head_after: 673d8906
commits: 2
tech-stack:
  added: []
  patterns:
    - independent per-package copies of the owner matcher (no shared helpers between peers)
    - compiler-minted release at operation index 3, never a source call
key-files:
  created:
    - examples/checksum/checksum.schway
    - examples/checksum/adapter.c
    - examples/checksum/adapter.h
    - examples/checksum/file_bytes.bindings.json
    - internal/compiler/session/session_filebytes.go
    - cmd/schway/main_phase31_test.go
  modified:
    - internal/compiler/ability/ability.go
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase30_prechange_test.go
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "D-16 core layout: with a trailing call, place:4 is an unread typed placeholder (_released) because a target must be place:<op index + 1> and places must be contiguous; the trailing call is op 4 targeting place:5"
  - "Entry-module trailing callee is recognised by ID prefix (entry function ID minus its name) in every peer; check uses the callee-contract table, the others use the core function IDs"
requirements-completed: []
---

# Phase 31 Plan 01: Bounded file checksum tracer Summary

A file containing `Wikipedia` prints `300286872` through `schway build --manifest` and `schway app run` on native, via a new FileBytes owner shape accepted independently by five peers, with a silent adapter and a generated `FileError.<Alternative>` stderr table.

## Accomplishments

- Tracer: `TestPhase31ChecksumPublicRoute` passes all five subtests (wikipedia, missing, trailing-call-tempdir, second-trailing-call-refused, mutated-library).
- Missing file prints exactly `FileError.FileUnavailable\n`, exit 65, empty stdout, path-free.
- 6-operation shape (trailing `low16`) prints `920`; a second trailing call is refused with `check.file_bytes_owner_shape_unsupported` (exit 2).
- Adapter is silent (no stdio calls), uses `O_NONBLOCK` + `fstat`/`S_ISREG`, a limit+1 read probe with EINTR retry, never reads `st_size`, and shrinks to `max(1,n)`.
- Registers updated (phase28 row, phase30 key, corpus counts 162 / 5,332). Phase 23 and Phase 29 matchers, adapters and tests are byte-unchanged; `session.go` edit is a single line-neutral callee swap.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Trailing-call place numbering rejected by core validation**
- **Found during:** Task 1 (trailing-call-tempdir)
- **Issue:** `core.place_order` / `core.invalid_target`: places must be contiguous and a call target must be `place:<op index + 1>`.
- **Fix:** insert an unread `_released` placeholder place at `place:4` so the trailing call (op 4) targets `place:5`.
- **Files modified:** internal/compiler/check/check.go
- **Commit:** aaa4ec83

None otherwise - plan executed as written.

## Known Stubs

None.

## Threat Flags

None beyond the plan's threat register (T-31-03..06 mitigated as specified; FIFO deadline rows and the size mutant kill land in plans 31-06 and 31-04).

## Self-Check: PASSED

All created files exist; commits aaa4ec83 and 673d8906 are in HEAD.

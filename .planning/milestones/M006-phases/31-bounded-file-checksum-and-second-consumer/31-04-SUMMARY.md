---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 04
subsystem: native-adapter-fault-injection
tags: [fault-injection, asan, bounded-read, APP-12, APP-11]
requires: [31-01, 31-02]
provides: [counting observer, release-exactly-once proof, bounded-read mutant controls]
affects: [31-08]
key-files:
  created:
    - internal/compiler/native/testdata/phase31_observer.c
    - internal/compiler/native/native_phase31_fault_test.go
  modified: []
decisions:
  - "Hooks are inserted into a TempDir copy of examples/checksum; the shipped adapter is never edited."
  - "The ASan double-release lane builds without observer hooks so the real free runs twice."
status: complete
commits: 3
plan_head_before: 8bae8c06642d7da27842bc35bc12b4433a349063
plan_head_after: 9469f46a8d803c1d1566d7f82c3fc66c5d2d5054
actuals:
  tokens: 25000
  tasks: 3
  commits: 3
---

# Phase 31 Plan 04: Adapter fault injection Summary

A counting observer linked into a TempDir copy of the checksum adapter proves every path (success, oversize, malloc NULL, realloc NULL, fstat fail, read EIO, close fail) frees each allocation exactly once with no double free. EINTR and one-byte reads are retried, and a 1 MiB file is refused after reading at most 65,537 bytes.

## What was built
- `phase31_observer.c`: wrappers for open/fstat/read/malloc/realloc/free/close, a 64-entry live-pointer table, env-selected faults (`SCHWAY_PHASE31_FAULT`), and an atexit receipt (`SCHWAY_PHASE31_RECEIPT`). Compiles under `-std=c17 -Wall -Wextra -Werror -pedantic`.
- `TestPhase31ReleaseExactlyOnce`: 11 subtests. Expected success values come from `testsupport.ChecksumOracleRows`.
- `TestPhase31BoundedReadMutants`, each mutant observed to fail against its faithful control:
  - growth-after-fstat: faithful reports FileTooLarge.
  - stat-size mutant: prints a checksum on the growth row.
  - `>=` limit mutant: refuses the 65,536-byte file.
  - no-S_ISREG mutant: prints `1` for /dev/null and for a FIFO.
  - no-O_NONBLOCK mutant: hangs on the FIFO until the 5 s context deadline (RunTimedOut).
  - ASan double-release mutant: stderr contains `AddressSanitizer: attempting double-free`. The faithful build under ASan gives exactly `FileError.FileTooLarge`.

## Deviations from Plan
None. The plan was executed as written. A failure of the hooked build on the first try was not seen.

## Known Stubs
None.

## Threat Flags
None.

## Self-Check: PASSED
Files and commits e0f192c6, 23f6c97f, 9469f46a exist; `TestPhase31ReleaseExactlyOnce` and `TestPhase31BoundedReadMutants` pass locally with ASan present (no skip); `TestSourceNeverSpawnsUnboundedProcesses` and `TestNoSuppressionOutlivesItsWitness` pass.

---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "04"
subsystem: testing
tags: [interpreter, native, file-bytes, crc32, bounds, cleanup, asan, ubsan]

requires:
  - phase: 37-02
    provides: bounded file-bytes consumer and interpreter application route
provides:
  - zero-progress reader failures, exact cap-plus-one allocation, and descriptor fault coverage
  - repeated-call capture-retention assertions with exact terminal outcomes
  - physical CRC adapter read/close/allocation/free receipts and reached bound/free mutants
  - focused ASan/UBSan evidence for generated CRC success and adapter cap-plus-one cleanup
affects: [phase-37, interpreter-file-application, native-file-bytes-adapter]

actuals:
  tokens: 7416
  tasks: 2
  commits: 3
  plan_head_before: 3ce9c6d5d8e0f75c4b5214ed76eef0a488ec4b35
  plan_head_after: 23f4c90f6a87cd8ec99962cf5c7259b4c01170a7

tech-stack:
  added: []
  patterns:
    - private descriptor/allocation seams for bounded interpreter reads
    - private copied C adapter with syscall and allocator receipts for native controls

key-files:
  created:
    - internal/compiler/native/native_phase37_filebytes_test.go
  modified:
    - internal/compiler/interp/filebytes_reader.go
    - internal/compiler/interp/filebytes_reader_test.go
    - internal/compiler/interp/interp_phase37_test.go

key-decisions:
  - "Treat a zero-byte, nil-error read as FileUnavailable; it is not EOF evidence."
  - "Allocate exactly cap+1 bytes before the first descriptor read and independently observe the allocation/read order."
  - "Keep native physical close/free receipts distinct from interpreter semantic owner-release counts."
  - "Run focused ASan/UBSan against the generated CRC consumer and use a returning C caller for its typed cap-plus-one adapter error path."

patterns-established:
  - "Private native mutants must compile and reach exact physical assertions before they count as killed."
  - "Record sanitizer host availability in a named subtest; this harness makes no Darwin leak-check claim."

requirements-completed: [] # APP-13 and APP-14 remain pending parent acceptance.
coverage:
  - id: D1
    description: "Interpreter file reads enforce the real cap, descriptor type, bounded allocation, and once-only close behavior."
    verification:
      - kind: unit
        ref: "internal/compiler/interp/filebytes_reader_test.go#TestPhase37FileReaderFaults"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/filebytes_reader_test.go#TestPhase37FileReaderBoundControls"
        status: pass
    human_judgment: false
  - id: D2
    description: "Capture-disabled repeated calls retain no optional trace or identity state while preserving output and terminal outcomes."
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_phase37_test.go#TestPhase37CaptureRetentionControls"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_phase37_test.go#TestPhase37CaptureDisabledOutcomes"
        status: pass
    human_judgment: false
  - id: D3
    description: "The actual CRC adapter closes and frees on success and typed error paths, and reached mutants expose missing cleanup or a one-byte-short read."
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_phase37_filebytes_test.go#TestPhase37CRC32PhysicalCleanup"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/native_phase37_filebytes_test.go#TestPhase37CRC32BoundControls"
        status: pass
    human_judgment: false
  - id: D4
    description: "Focused ASan/UBSan runs cover the generated CRC consumer and actual adapter cap-plus-one cleanup on this host."
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_phase37_filebytes_test.go#TestPhase37CRC32Sanitizer"
        status: pass
    human_judgment: false

duration: 11m
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 04: Reader Bounds and CRC Physical Cleanup Summary

**The interpreter reader now rejects zero-progress reads and proves the cap-plus-one boundary; the native CRC route records physical close/free evidence and passes its focused sanitizer lane.**

## Performance

- **Duration:** 11 minutes
- **Started:** 2026-10-15T15:44:53Z
- **Completed:** 2026-10-15T15:56:07Z
- **Tasks:** 2
- **Files modified:** 5, including this summary

## Accomplishments

- Added fault and bound controls for open/stat/read/close errors, EINTR, short and zero-progress reads, forged descriptor sizes, FIFO/device refusal, symlink handling, and replacement after open. The buffer is exactly 65,537 bytes and its allocation is observed before the first read.
- Exercised capture-disabled execution through three repeated scalar calls. Optional event, identity, ordinal, and evidence stores remain zero; the result is `8`, and existing typed-error, prior-output, checked-defect, and capture-exhaustion outcomes remain pinned.
- Built private copies of the actual CRC adapter with physical syscall and allocator observers. Success, stat/read/close failures, oversize precedence, and EINTR retry record exact reads, closes, allocations, and frees. A missing-free mutant and a one-byte-short request mutant both build and reach their discriminating assertions.
- The combined ASan/UBSan capability probe passed. Sanitized generated CRC runs passed for a small input and the exact 65,536-byte boundary; a returning C caller exercised the adapter's actual 65,537-byte `FileTooLarge` cleanup path. Leak detection is disabled by the established harness, so no Darwin leak claim is made.

## Task Commits

1. **Task 1: Reach reader fault and boundary controls** - `b5e2ee7a` (test)
2. **Task 1: Enforce reader progress and allocation bounds** - `17723f11` (fix)
3. **Task 2: Observe physical CRC cleanup and sanitizer bounds** - `23f4c90f` (test)

## Files Created/Modified

- `internal/compiler/interp/filebytes_reader.go` - adds a narrow allocator seam and maps zero-progress reads to `FileUnavailable`.
- `internal/compiler/interp/filebytes_reader_test.go` - covers bounded allocations, descriptor fault precedence, path replacement, and nonblocking special-file refusal.
- `internal/compiler/interp/interp_phase37_test.go` - extends disabled-capture checks through repeated calls and preserves distinct semantic cleanup outcomes.
- `internal/compiler/native/native_phase37_filebytes_test.go` - adds private adapter copies, physical receipts, killed mutants, and focused sanitizer runs.

## Decisions Made

- A zero-byte read without an error cannot establish EOF for this descriptor contract, so the reader fails closed with `FileUnavailable`.
- Native syscall and allocator receipts prove physical cleanup; interpreter probes continue to report semantic owner events separately.
- The sanitizer's typed-error branch uses a small returning C caller because the generated application intentionally exits with code 65 for `FileError`; the generated CRC consumer itself is also run under sanitizers on success and exact-cap inputs.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Reject zero-progress descriptor reads**
- **Found during:** Task 1 (reader fault and boundary controls)
- **Issue:** A descriptor returning `(0, nil)` was treated like EOF and could return partial data as successful input.
- **Fix:** Convert zero progress into a read failure while preserving once-only close and typed `FileUnavailable` behavior.
- **Files modified:** `internal/compiler/interp/filebytes_reader.go`, `internal/compiler/interp/filebytes_reader_test.go`
- **Verification:** The new zero-progress case failed before the fix and passed with the exact focused interpreter command.
- **Committed in:** `17723f11`

**Total deviations:** 1 auto-fixed bug.
**Impact on plan:** The fix closes a correctness gap directly in the reader contract; no scope expansion.

## Issues Encountered

- The sanitizer-generated C source referred to the manifest include label `local/adapter.h`; the private sanitizer source now resolves that include to the copied header's absolute path.
- The initial short-read mutant changed the loop condition while leaving the first request at 65,537 bytes, so it was not a one-byte-short control. The control was corrected to cap the requested byte count; the reached run now records 65,536 of 65,537 bytes and is killed by the exact boundary assertion.

## User Setup Required

None. Clang and the combined sanitizer runtime were available on this host.

## Next Phase Readiness

The plan's focused reader and native evidence is ready for the remaining Phase 37 plans. APP-13 and APP-14 stay pending parent acceptance; this summary does not mark requirements complete.

## Self-Check: PASSED

- All four plan-owned source/test files exist, and the three task commits are ancestors of the plan head recorded above.
- The focused interpreter and native verification commands each passed after the final task commit; every named root ran and passed.
- The source inventory assertions confirmed the tracked CRC adapter, header, manifest, entry/library sources, and reused Phase 31 observer were unchanged by private controls.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

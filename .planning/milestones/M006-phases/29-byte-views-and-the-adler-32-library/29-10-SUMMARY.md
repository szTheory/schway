---
phase: 29-byte-views-and-the-adler-32-library
plan: 10
subsystem: native-evidence
tags: [adler32, differential, mutation, oob-control, sanitizer, d-07, d-19, d-20, d-21]
requires: [29-07, 29-08]
provides: [both-engine-pinned-literals, d07-mutants-killed, oob-control-all-opt-levels, asan-lanes]
affects: [29-11, 29-12]
tech-stack:
  added: []
  patterns: [test-only package-level seams swapped and restored, exact-once mutation anchors, per-claim asan subtests]
key-files:
  created:
    - internal/compiler/native/export_phase29_test.go
    - internal/compiler/native/native_phase29_test.go
    - examples/phase29/byte_index_oob.schway
  modified:
    - internal/compiler/native/native_app.go
    - examples/phase29/README.md
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase272_identity_test.go
    - internal/compiler/session/session_phase273_identity_test.go
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "The C-mutation hook is applied at the top of BuildApplication so the receipt digests describe the mutated C that was actually compiled"
  - "The ASan lane passes -fno-sanitize-ignorelist: the bindings dependency check refuses clang's asan_ignorelist.txt as an out-of-root header"
  - "LeakSanitizer is unsupported on darwin, so the lane requests detect_leaks=0 there and records no leak detection in the README"
requirements-completed: [APP-10, BYTE-01, BYTE-02]
status: complete
actuals:
  tokens: 40000
  tasks: 3
  commits: 3
plan_head_before: ea66167e27d32894213ceed65a9bfbb25378c58b
plan_head_after: ec63996e4a3f45a59bad3b2fdb0d5ac5795aab1e
completed: 2026-10-11
---

# Phase 29 Plan 10: Adler-32 both-engine proof, mutants and OOB control Summary

All ten pinned Adler-32 literals match on the interpreter and on native builds at -O0, -O3 and -O3 -flto; the wrong-modulus, off-by-one-start, swapped-halves and signed-char mutants are killed; the out-of-bounds control exits 65 with the exact stderr at every optimization level and as the D-19 defect in the interpreter.

## Accomplishments

- Test-only seams (`SetApplicationFlagsForTest`, `SetApplicationCSourceMutationForTest`) in export_phase29_test.go with a nil-by-default hook in native_app.go; production `applicationFlags` unchanged and no non-test caller exists.
- `byte_index_oob.schway` registered in phase28NewlyAddedEntries, phase272NewlyAcceptedPaths, phase273NewlyAcceptedPaths; LANGUAGE-MATURITY corpus sentence updated to 159 programs, 5,196 lines.
- `TestPhase29AdlerBothEngines`, `TestPhase29AdlerMutantsKilled`, `TestPhase29OOBApplicationLanes` (with `asan/vectors` and `asan/oob-control` subtests).
- README: build/run commands, D-20 note, "Sanitizer evidence" table.

## Mutant observations (case 5 / case 6, pinned 4052720524 / 33620352)

- wrong modulus: 3690503073 / 33620352 (killed on case 5 only; case 6 unchanged)
- off-by-one start: 1443076749 / 16777472
- swapped halves: 2609705359 / 25166337
- signed char (C text): 2459237135 / 27525442, exit 0, matching the hand calculation

## Sanitizer evidence

Host darwin/arm64, Apple clang 21.0.0: all three lanes ran with no report (ASan available). LeakSanitizer is unsupported on darwin, so no leak detection ran here; the vector lane would request `detect_leaks=1` on a Linux host. No PARTIAL marker is needed for plan 29-11 on this host, but 29-11 should carry the darwin no-leak-detection note.

## Deviations from Plan

**1. [Rule 3 - Blocking] ASan build refused by the bindings header check**
- **Found during:** Task 3
- **Issue:** with -fsanitize=address the binding probe lists clang's asan_ignorelist.txt as an undeclared header (`native.bindings_invalid`).
- **Fix:** the test lane adds `-fno-sanitize-ignorelist`; production code untouched.
- **Files modified:** internal/compiler/native/native_phase29_test.go
- **Commit:** ec63996

**2. [Rule 1 - Plan precision] detect_leaks=1 not applied on darwin**
- LeakSanitizer is unsupported on darwin; the lane uses detect_leaks=0 there (documented in README). Linux hosts request detect_leaks=1 for the vector lane.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Created files exist; commits f8e5f2b, fd22584, ec63996 are ancestors of HEAD; focused tests and session meta-tests pass.

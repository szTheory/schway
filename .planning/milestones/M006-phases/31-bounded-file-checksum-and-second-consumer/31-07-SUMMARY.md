---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 07
subsystem: testing
tags: [checksum, adler_low, single-copy, dual-flip, module-identity]
requires: [31-01, 31-02, 31-05]
provides:
  - examples/checksum/adler_low.schway, the second consumer of the one adler32 library (920 for Wikipedia)
  - TestPhase31AdlerLowPublicRoute
  - TestPhase31SingleCopyAndDualFlip (8 subtests)
affects: [31-08, 31-09]
key-files:
  created:
    - examples/checksum/adler_low.schway
    - cmd/schway/main_phase31_adler_low_test.go
    - internal/compiler/session/session_phase31_single_copy_test.go
  modified:
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase30_prechange_test.go
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "Dual flip compares receipt Modules rows, BuildID inequality and evidence Modules; never ExecutableDigest equality across builds (host clang is non-reproducible)"
status: complete
commits: 3
plan_head_before: c057eef4cfa52a29ca5ee46fe27a4227f59a87ed
plan_head_after: 7de1a3ee
actuals:
  tokens: 30000
  tasks: 3
  commits: 3
---

# Phase 31 Plan 07: adler_low second consumer Summary

A second ordinary application, `adler_low`, imports the single `adler32` library and prints its low 16 bits (920 for Wikipedia against checksum's 300286872). A static scan with a sha256-frozen allow-list, per-app merged-program checks, a copy-detection control, a dual flip and a B-only reverse control show the library is imported, not copied.

## Accomplishments

- `adler_low.schway`: checksum shape plus a trailing `low16(sum)`; no loop copy, library unchanged.
- Public route: all adler_low oracle rows (empty 1, 0x00 1, 65536x0xFF 3826, 65537x0xFF FileTooLarge exit 65, 0x80, 0x80 0xFF) and `same-file-different-result`.
- Single copy: exactly three files carry the loop (library plus two hash-frozen Phase 29 fixtures); each app merges exactly one `adler32` in module `adler32`; a pasted copy is found by the scan and refused with `module.import_collision`.
- Dual flip (`var a = 2`): both outputs change on native and interpreter, both BuildIDs move, the adler32 Modules row is new and identical in both receipts, entry rows unchanged, and evidence Modules equal receipt Modules for pristine and mutated artifacts of both apps.
- Reverse control (`var b = 1`): checksum changes, adler_low stays 920 on both engines.
- Registers: phase28NewlyAddedEntries row, phase30NewCommittedPaths key, LANGUAGE-MATURITY corpus counts (163 programs, 5,389 lines).

## Deviations from Plan

None in substance. The first task commit initially omitted the LANGUAGE-MATURITY edit (a failed `sed` on macOS); it was corrected by amending that unpushed commit before any later work.

## Known Stubs

None.

## Self-Check: PASSED

Files exist, the three task commits are in history, and the focused verify commands for all three tasks pass.

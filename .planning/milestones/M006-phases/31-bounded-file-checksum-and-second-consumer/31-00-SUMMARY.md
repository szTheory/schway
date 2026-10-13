---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 00
subsystem: examples/modules
tags: [move, re-pin, adler32, local-modules]
requires: []
provides:
  - examples/checksum/ holding the single Adler-32 library and the Phase 30 CLI
affects: [31-01 onward consumers that import the library]
tech-stack:
  added: []
  patterns: [atomic git mv + re-pin commit, sha256 identity guard with non-inert control]
key-files:
  created:
    - internal/compiler/session/session_phase31_move_test.go
  modified:
    - examples/checksum/adler32.schway
    - examples/checksum/adler32_cli.schway
    - examples/checksum/README.md
    - cmd/schway/main_phase30_test.go
    - internal/compiler/session/session_phase30_adler_test.go
    - internal/compiler/session/session_phase272_identity_test.go
    - internal/compiler/session/session_phase273_identity_test.go
    - internal/compiler/session/session_phase30_prechange_test.go
    - internal/compiler/session/session_phase28_manifest_test.go
    - .planning/LANGUAGE-MATURITY.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/todos/pending/2026-10-12-formatter-merges-trailing-result-into-var-assignment.md
decisions:
  - "D-01: library and CLI live in examples/checksum/; library bytes unchanged (sha256 317ac139...)"
metrics:
  duration: short
  completed: 2026-10-12
status: complete
actuals:
  tokens: 9000
  tasks: 2
  commits: 2
plan_head_before: 70710df6478958da7027510e208d6874c614d527
plan_head_after: 62be3803d7b283ef6b226225bbe9537ce0ff8a70
---

# Phase 31 Plan 00: Move Adler-32 library and CLI into examples/checksum Summary

One rename commit moves the Adler-32 library, Phase 30 CLI and README into `examples/checksum/` and re-pins every live path site; a second commit adds a sha256 guard test with a non-inert control.

## Commits
- 2b7a0c18: refactor(31-00): move Adler-32 library and CLI into examples/checksum (D-01)
- 62be3803: test(31-00): pin the moved Adler-32 library bytes (D-01)

## Verification
- TestPhase30AdlerModuleAppCLI (native, not skipped), TestPhase30AdlerModuleInterpreter, TestPhase30LibraryIsSingleCopy, TestPhase30PreChangeGolden, TestPhase28EntryExitClassManifest, the 27.2/27.3 identity walks and TestPhase20ChecksumFrontier pass.
- TestPhase31MovedLibraryIdentity and its non-inert subtest pass; the session meta set (inventory, groundedness, reconciliation, validation lifecycle, maturity counts) is green.
- Library sha256 unchanged; CLI diff is a rename with one changed line (module header).
- `git grep` for the old directory across cmd, internal, examples, LANGUAGE-MATURITY and pending todos returns no match.

## Deviations from Plan
- [Rule 3 - Blocking] After `git mv`, an empty untracked `examples/phase30/` directory remained on disk and made the guard test's "old directory is gone" check fail; removed with `rmdir` (not tracked by git).
- The README title and its inline `examples.phase30.adler32_cli` header mention were also re-pathed (the README's first line was retitled per the plan).
- The PRODUCT-ROADMAP amendment was appended to the same line as the "What 30 changed" paragraph, leaving the historical build command text intact.

## Known Stubs
None.

## Self-Check: PASSED

---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 08
subsystem: testing
tags: [evd-14, oracle, mutants, asan, signed-char, inventory]
requires: [31-04, 31-06, 31-07]
provides:
  - TestPhase31OracleRows and checksumOracleRowProblems (Python command plus Go hash/adler32 cross-check)
  - TestPhase31Controls (8 subtests) and TestPhase31ControlInventory
  - TestPhase31LoopNeighborRefusals (positive control plus nine CLI refusals)
affects: [31-09]
key-files:
  created:
    - internal/compiler/testsupport/checksum_oracle_test.go
    - internal/compiler/native/native_phase31_controls_test.go
    - cmd/schway/main_phase31_neighbors_test.go
  modified:
    - internal/compiler/interp/interp.go
key-decisions:
  - "Off-by-one mutant is `n = n + 1` after `var n = byte_len(bytes)`; the language has no <= operator"
  - "Inventory counts a subtest literal as a t.Run first argument or the first field of a table row, and requires the function to call t.Run"
status: complete
commits: 3
plan_head_before: 46f5c4ab0c3f61facfd0a7c55664624148753bf0
plan_head_after: d26be97191992af93318fc909a6c6b642c7f482a
actuals:
  tokens: 22000
  tasks: 3
  commits: 3
---

# Phase 31 Plan 08: EVD-14 controls Summary

Every pinned checksum answer is cross-checked against Go `hash/adler32` beside its recorded Python command, each of the eight EVD-14 controls is observed to fail in a subtest, and an inventory proves the one-to-one wiring.

## Accomplishments

- Oracle rows: 14 rows, every Cmd begins `python3 -c 'import zlib;`, every Want equals `hash/adler32` (`% 65536` for adler_low). The 65,537x0xFF rows are FileTooLarge with `hash/adler32` 2273841137 and 4081, matching what the recorded Cmd prints; an armed blank-Cmd control reports exactly one problem.
- Controls on native and interpreter (copy of examples/checksum): wrong modulus, wrong initial state and swapped halves each change an answer on both engines; the unmutated control matches. The off-by-one bound gives exit 65 with the exact stderr natively and the byte-index defect in the interpreter. The signed-char retype changes 0x80/0xFF results; the `-fsigned-char` lane yields 8454273, 33620352, 2006388466. With the bounds check deleted, ASan reports `heap-buffer-overflow` and not the checked failure.
- Nine loop neighbours refused by `schway check` with exit 2 and their Phase 29 first codes; var-first positive control exits 0.
- Inventory: eight bullets of EVD-14 map to eight unique existing subtests, including `TestPhase31SingleCopyAndDualFlip/dual-flip` for the library mutation; missing and duplicate mappings are each reported.

## Deviations from Plan

**1. [Rule 3 - Blocking] No `<=` operator.** `while i <= n` is a syntax error. Replaced by `n = n + 1` after the length is read, which has identical behavior. Commit db5025e1.

**2. [Rule 1 - Bug] Interpreter file route rejected the byte-index defect.** `RunWithFileBytes` returned a "live local resources" error on the abort-like defect, unlike the fixture route. Added the same `isByteIndexDefect` exemption in `internal/compiler/interp/interp.go` (outside the plan's file list). Interp tests pass. Commit db5025e1.

**3.** The inventory scan also accepts the first field of a table row as the subtest literal, because `wrong-*`, `swapped-halves` and the neighbour rows reach `t.Run` through a table variable.

## Known Stubs

None.

## Self-Check: PASSED

Three task commits exist (d3a6b94c, db5025e1, d26be971); the three created test files exist; all target tests pass locally, ASan lane included (not skipped).

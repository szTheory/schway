---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 01
subsystem: syntax/formatter
tags: [dx-19, formatter, lossless, golden-allow-list]
requires: []
provides:
  - second newline guard in syntax.Format keeping a trailing var-body result on its own line
  - TestPhase32FormatVarBody (table, armed pre-fix control, five-copy corpus check)
  - phase30FormatChanges dated allow-list with stale-entry control
affects: [32-09]
key-files:
  created:
    - internal/compiler/syntax/syntax_phase32_format_test.go
    - .planning/todos/completed/2026-10-12-formatter-merges-trailing-result-into-var-assignment.md
  modified:
    - internal/compiler/syntax/format.go
    - internal/compiler/session/session_phase30_prechange_test.go
key-decisions:
  - "Guard fires for an identifier after an identifier, number or right paren across a source newline, in function bodies, excluding match context (D-10c)"
  - "Three frozen Phase 30 format hashes are carried by phase30FormatChanges; the golden stays byte-frozen (D-06, D-15)"
requirements-completed: [DX-19]
status: complete
commits: 2
plan_head_before: 48e667ffe38a6607cf22cef12be1cf448085a7cc
plan_head_after: 207028b171f1ccb36c48879820620184b23413c1
actuals:
  tokens: 47000
  tasks: 2
  commits: 2
coverage:
  - id: D-10
    description: "schway format keeps a trailing var-body result identifier on its own line; output reparses, is a fixed point, and preserves semantic tokens, checked on three terminal shapes and five committed sources as in-memory copies"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase32_format_test.go#TestPhase32FormatVarBody"
        status: pass
    human_judgment: false
  - id: D-10c
    description: "The frozen Phase 30 pre-change golden passes through a dated phase30FormatChanges allow-list whose stale-entry control is observed to fail"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_prechange_test.go#TestPhase30PreChangeGolden"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase30_prechange_test.go#TestPhase30PreChangeGoldenIsNotInert"
        status: pass
    human_judgment: false
duration: 20min
completed: 2026-10-12
---

# Phase 32 Plan 01: Formatter var-body fix Summary

A second newline guard in `syntax.Format` stops `schway format` from fusing a bare trailing result identifier into the preceding `var` assignment (`r = r + a` / `r` no longer becomes `r = r + ar`), with the three moved Phase 30 format hashes carried by a dated allow-list.

## Accomplishments

- Guard in `format.go` before `f.token(token, next)`: identifier token, `index > 0`, line open, inside a function body, context not `match`, previous token an identifier, number or `)`, and a newline in the source gap. Match-arm outputs are unchanged.
- `TestPhase32FormatVarBody`: identifier, number and call-paren terminals; an armed control showing the pre-fix merged text has different semantic tokens; the five D-10c sources formatted as in-memory copies (reparse clean, fixed point, equal semantic tokens, equal check diagnostic codes). RED was observed before the guard: `var x = byte_at(bytes, n) x` and four of the five copies failed.
- `phase30FormatChanges` holds the 3 new format_sha256 values read from the real fix; an entry equal to the golden is reported stale, and the `format-change-stale` subtest observes that report. `testdata/phase30/prechange-corpus.json`, `examples` and `testdata` are byte-unchanged.
- The pending todo moved to `todos/completed` with a Resolution section.

## Deviations from Plan

None - plan executed exactly as written. Between Task 1 and Task 2 the session package failed only in `TestPhase30PreChangeGolden` (the 3 expected hash problems) and its baseline-dependent `TestPhase30PreChangeGoldenIsNotInert`; both are green after Task 2.

## Verification

- syntax, cmd/schway, check, cgen, core, evidence, cmd/schway-repair: ok; `TestShippedBinaryExercisesEveryPhase4Behavior` passes.
- Full session package ok; session meta set ok.
- `git diff 45a526b8 -- testdata examples` is empty.

## Known Stubs

None.

## Self-Check: PASSED

Created files exist; commits 45a526b8 and 207028b1 are on the branch.

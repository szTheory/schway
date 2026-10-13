---
phase: 29-byte-views-and-the-adler-32-library
plan: 02
subsystem: compiler
tags: [byte-view, parser, diagnostics, formatter, reserved-names]
requires: [29-00, 29-01]
provides:
  - syntax.byte_len_arity and syntax.byte_at_arity named refusals with recovery to the closing paren
  - check.reserved_builtin_name for user fn byte_len / byte_at
  - let-first ByteView bodies reach the scalar parser (scalarLetStart / scalarBodyStart widened)
  - formatter goldens for byte read spacing; D-18 data-keyword pin
affects: [29-03]
tech-stack:
  added: []
  patterns: [name-recognized builtin atom, named arity refusal]
key-files:
  created:
    - internal/compiler/syntax/syntax_phase29_test.go
    - internal/compiler/check/check_phase29_surface_test.go
  modified:
    - internal/compiler/syntax/parser.go
    - internal/compiler/ast/ast.go
    - internal/compiler/check/check.go
key-decisions:
  - "No machine-applicable repair added; schway-repair decline-vocabulary tests pass unchanged"
  - "format.go needed no change for let-bound byte reads"
requirements-completed: [BYTE-01]
status: complete
commits: 2
plan_head_before: e393b4415879b2b3de73a26fe554ed3ede3f826f
plan_head_after: 6785f450adc631cd0d2292c4bd5d613fe0a96c06
actuals:
  tokens: 9000
  tasks: 2
  commits: 2
duration: 15min
completed: 2026-10-11
---

# Phase 29 Plan 02: Byte read surface Summary

byte_len and byte_at now have named arity refusals, reserved builtin names, let-first ByteView scalar bodies and formatter goldens, with no lexer, corpus or format.go change.

## Tasks

1. f39a0de: parser arity branches (`byteReadArity` helper recovers to the closing paren), scalarLetStart admits `let x = byte_len(` / `byte_at(` only, scalarBodyStart admits a ByteView parameter, ast.Expr Kind doc lists byte_len / byte_at. Tests: ParseByteReads, ByteReadArity, LetFirstByteViewBody, DataParameterRefused.
2. 6785f45: check.Program refuses a function named byte_len / byte_at with check.reserved_builtin_name; TestPhase29FormatByteReads (idempotent goldens).

## Verification

Focused syntax and check packages, schway-repair decline vocabulary tests, and session Inventory/Groundedness/Reconciliation/ValidationLifecycle plus corpus identity and class-manifest tests all pass. lexer.go and testdata/examples untouched.

## Deviations from Plan

**1. [Rule 3 - Adaptation] Formatter goldens use `let` bindings.** The formatter does not format Phase 26 scalar `var` bodies today (pre-existing: `var x = 1` then `x` renders `var x = 1x`, and `else` lands on its own line). Fixing that is outside this plan (only spacing inside the two builtin call forms may change), so goldens use `let x = byte_at(bytes,i)` shapes, which already format correctly. Logged to the pre-existing limitation list below.

**2. [Minor] Reserved-name diagnostic span** uses the whole function declaration span, since ast.FuncDecl does not retain a name-only span. Check tests are in package check_test (the directory convention) rather than package check.

**3. [Minor]** The reserved-name refusal returns before the mixed-body check, so it is always the first diagnostic. Added TestPhase29ByteReadNotResolvedAsUserCall (byte_len over a U64 operand is not accepted).

## Deferred / pre-existing

- Formatter mishandles scalar `var` statement bodies (missing newline before the result, `else` on its own line). Not caused by this plan; any plan that needs canonical formatting of var-based byte-view bodies (29-03 adler32 library) should address it first.

## Known Stubs

None.

## Threat Flags

None. T-29-07 mitigated (reserved name test), T-29-08 mitigated (arity recovery tests pin the first code).

## Self-Check: PASSED

Created test files exist; commits f39a0de and 6785f45 are on the branch.

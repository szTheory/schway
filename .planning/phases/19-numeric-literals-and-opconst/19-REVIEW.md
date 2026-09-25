---
phase: 19-numeric-literals-and-opconst
reviewed: 2026-09-24T00:00:00Z
depth: standard
files_reviewed: 33
files_reviewed_list:
  - internal/compiler/ability/ability.go
  - internal/compiler/ability/ability_test.go
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase19_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_convention_absence_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_phase19_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase19_integration_test.go
  - internal/compiler/interp/interp_phase19_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_phase19_test.go
  - internal/compiler/session/session_phase7.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase19_test.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase19/literal_malformed.lang
  - testdata/phase19/literal_overflow.lang
  - testdata/phase19/literal_tracer.lang
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
status: issues_found
---

# Phase 19: Code Review Report

**Reviewed:** 2026-09-24T00:00:00Z  
**Depth:** standard  
**Files Reviewed:** 33  
**Status:** issues_found

## Summary

Reviewed the Phase 19 compiler, validator, interpreter, emitter, syntax, and fixture changes. I found two input handling inconsistencies: uppercase radix prefixes are accepted by lexing but rejected during checking, and the interpreter accepts a signed decimal U64 input that the generated native program rejects.

## Warnings

### WR-01: Uppercase radix prefixes pass lexing but fail numeric conversion

**Classification:** WARNING  
**File:** `internal/compiler/syntax/lexer.go:164-167`; `internal/compiler/check/check.go:5420-5424`  
**Issue:** `validNumericLiteral` accepts `0X` and `0B`, but `parseU64Literal` recognizes only lowercase `0x` and `0b`. Thus a literal such as `0XFF` tokenizes as valid, then is rejected by the checker as `check.literal_out_of_range`. This makes the lexer and semantic admission rules disagree and reports a false range error for valid digits.  
**Fix:** Normalize the prefix or recognize both cases in `parseU64Literal`, or reject uppercase prefixes in the lexer if the language intends lowercase-only syntax.

### WR-02: Interpreter and native U64 input parsing differ for leading plus

**Classification:** WARNING  
**File:** `internal/compiler/interp/interp.go:331-337`  
**Issue:** Go's `strconv.ParseUint` accepts a leading `+`, so `interp.Run(..., "+42")` accepts the input. The emitted C parser accepts only ASCII decimal digits and returns invalid-input for the same string. The two execution backends therefore disagree on whether this U64 input is valid.  
**Fix:** Enforce the native parser's digit-only grammar before calling `ParseUint`, or change the native parser to accept the same optional leading plus sign.

---

_Reviewed: 2026-09-24T00:00:00Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

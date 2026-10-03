---
phase: 26-checked-scalar-sum
reviewed: 2026-10-03T08:41:27Z
depth: standard
files_reviewed: 45
files_reviewed_list:
  - .github/workflows/ci.yml
  - cmd/schway/main_test.go
  - examples/phase24/README.md
  - examples/phase26/checked_add_overflow.schway
  - examples/sum_to_n.schway
  - internal/compiler/ability/ability.go
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_convention_absence_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_cycle_peer_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/execution/execution_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/native/phase25_utility_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/session/session_phase23_contract_test.go
  - internal/compiler/session/session_phase26_event_test.go
  - internal/compiler/session/session_phase26_overflow_test.go
  - internal/compiler/session/session_phase26_test.go
  - internal/compiler/session/session_phase5_compare.go
  - internal/compiler/session/session_phase5_compare_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
findings:
  critical: 1
  warning: 0
  info: 0
  total: 1
status: issues_found
---

# Phase 26: Code Review Report

**Reviewed:** 2026-10-03T08:41:27Z  
**Depth:** standard  
**Files Reviewed:** 45  
**Status:** issues_found

## Summary

Reviewed all 45 files in the resolved Phase 26 scope, including scalar parsing and lowering, fixed-point validation, interpreter/native execution, event handling, and application evidence. Checked U64 overflow produces a defect event in both execution paths, but the independent event peer cannot classify that event and the native size preflight incorrectly treats the overflowing operation as event-free. No tests or native programs were run under project policy.

## Critical Issues

### CR-01: Checked-add overflow events are rejected by the independent peer

**File:** `internal/compiler/executionpeer/executionpeer.go:451`  
**Issue:** On overflow, the interpreter emits `function.defected` with ID `<OpAddChecked ID>:event:defected` (`internal/compiler/interp/interp.go:1253-1259`). The non-application C execution path emits the same terminal event (`internal/compiler/cgen/cgen_program.go:2564-2572`, `2255-2272`). The peer only classifies `:event:defected` when the underlying operation is `OpDefect`, so validating either engine's direct checked-add overflow execution fails with `unknown_kind`. Also, `schema2ExecutionDocumentSize` skips every `OpAddChecked` as event-free (`cgen_program.go:202`) even though the overflow path appends a terminal event, so its capacity preflight undercounts overflow executions and can let an oversized document reach a runtime abort. Together these make valid overflow evidence unusable and can turn capacity exhaustion into a process abort instead of a clean refusal.
**Fix:** In `classify`, recognize `OpAddChecked`'s `:event:defected` event as the fixed checked-overflow defect, and validate its reason and source/type attribution against the operation. Add a peer acceptance case for a reached overflow plus controls that reject a forged overflow event on non-overflowing execution.

## Warnings

## Info

---

_Reviewed: 2026-10-03T08:41:27Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

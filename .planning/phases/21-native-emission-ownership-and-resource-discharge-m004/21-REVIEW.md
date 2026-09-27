---
phase: 21-native-emission-ownership-and-resource-discharge-m004
reviewed: 2026-09-27T00:49:08Z
depth: standard
files_reviewed: 14
files_reviewed_list:
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/core/core.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interptestdirect/interptestdirect.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
  - internal/compiler/session/session_phase21_contract_test.go
  - internal/compiler/session/session_phase21_lto_evidence_test.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/verification_groundedness_test.go
findings:
  critical: 1
  warning: 0
  info: 0
  total: 1
status: issues_found
---

# Phase 21: Code Review Report

**Reviewed:** 2026-09-27T00:49:08Z  
**Depth:** standard  
**Files Reviewed:** 14  
**Status:** issues_found

## Summary

Reviewed the Phase 21 source scope and the whole-program emitter implementation it dispatches to. The schema-2 output-size preflight does not account for bytes added when a returned payload-bearing ADT is serialized. An input at the configured output limit can therefore begin emitting a document and then exit with status 74 partway through, instead of being refused before execution output starts.

## Critical Issues

### CR-01: [BLOCKER] Payload return bytes are omitted from output-size preflight

**File:** `internal/compiler/cgen/cgen_program.go:194-213`
**Issue:** For a match entry, `schema2ExecutionDocumentSize` estimates the returned value using only the longest raw `arm.Value`. The terminal writer in `emitProgramTerminalValueWriter` serializes payload-bearing results as the alternative name followed by `:` and the payload representation (up to eight hex bytes for Buffer or a decimal Byte value). Those bytes are absent from this estimate. As a result, a program whose actual document is just over `executionOutputLimit` can pass the preflight and begin writing; `lang_write_bytes` then fails partway through and `main` returns 74 with truncated, invalid JSON. This violates the preflight’s guarantee that oversized execution documents are rejected before C serialization/runtime output.

**Fix:** Compute the worst-case returned value using the same payload encoding contract as `lang_write_terminal_value` (including separator, payload width, and JSON escaping), or serialize a bounded canonical worst-case `execution.Execution` with the actual possible terminal values before comparing against the limit. Keep this calculation shared with the terminal writer’s representation so the two cannot drift.

## Warnings

## Info

---

_Reviewed: 2026-09-27T00:49:08Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

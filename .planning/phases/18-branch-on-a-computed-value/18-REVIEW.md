---
phase: 18-branch-on-a-computed-value
reviewed: 2026-09-26T15:12:00Z
depth: deep
files_reviewed: 4
files_reviewed_list:
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/session/session_phase18_payload_test.go
  - testdata/phase18/long_payload_return.lang
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 18: Code Review Report

**Reviewed:** 2026-09-26T15:12:00Z  
**Depth:** deep  
**Files Reviewed:** 4  
**Status:** clean

## Summary

Re-reviewed the emitter repair, the long-tag source witness, and its session regressions against Phase 18 Plan 18-10 and the earlier review. CR-01 is resolved: terminal tags and payloads no longer pass through a fixed 128-byte tag buffer. WR-01 is resolved: both mutation controls defer restoring the global seam immediately after installation and retain the eager restore after the runner returns; the captured prior value makes the second restore harmless. The schema-1 support writer remains unchanged; schema-2 streams terminal output through the shared escaping routine and bounded byte writer. No remaining findings.

All reviewed files meet quality standards. No issues found.

## Review Checks

- The prior CR-01 fixed `rendered[128]` formatting buffer is absent. `emitProgramTerminalValueWriter` streams the selected tag via `lang_write_json_string_content`, then writes the colon and selected runtime payload through `lang_write_bytes`; the Byte formatting scratch is four bytes, sufficient for three decimal digits plus the terminator.
- The terminal path retains unknown-tag/nested-tag validation, the Buffer length bound, lowercase hexadecimal payload output, and failure propagation from the bounded output writer.
- `emitEventSupport` and its schema-1 JSON writer are unchanged. The refactor is confined to `emitEventSupportSchema2`; the added reserved-name entries cannot change the existing allocator namespaces because generated names are prefix-confined.
- The long-tag test asserts the exact 152-character alternative and exact `tag:01020304` terminal outcome across interpreter and three native tiers, then checks the complete engine comparator. Both mutation tests assert a positive injected-write count and the exact terminal-outcome disagreement axis; each now defers restoration immediately and also restores eagerly on success.

---

_Reviewed: 2026-09-26T15:12:00Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: deep_

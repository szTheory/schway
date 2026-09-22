---
phase: 15-event-identity-lang-execution-2
plan: "08"
subsystem: native schema-admission evidence
tags: [execution-schema-2, native-decoder, tool-error, coverage]
requires:
  - phase: 15-event-identity-lang-execution-2
    provides: Schema 2 invocation identity and semantic admission rules
provides:
  - Canonical-JSON-to-ToolError Schema 2 admission seam regression
  - Machine-readable automated OBS-04 coverage for Plan 01
affects: [native-execution-validation, verify-work, phase-15-uat]
tech-stack:
  added: []
  patterns: [canonical-wire-seam-test, refusal-specific-tool-error-assertions, machine-readable-coverage]
key-files:
  created: [.planning/phases/15-event-identity-lang-execution-2/15-08-SUMMARY.md]
  modified:
    - internal/compiler/native/native_test.go
    - .planning/phases/15-event-identity-lang-execution-2/15-01-SUMMARY.md
key-decisions:
  - "Exercise decodeExecution from execution.CanonicalBytes so JSON decoding, strict fields, semantic validation, and ToolError wrapping form one admission oracle."
  - "Publish OBS-04 as deterministic integration coverage rather than route it to human UAT."
actuals:
  tokens: 2050
  tasks: 2
  commits: 3
plan_head_before: 14282b1f8b55045ffcd7151246b30e91fb888120
duration: 7min
completed: 2026-09-22
status: complete
---

# Phase 15 Plan 08: Schema Admission Coverage Summary

Canonical Schema 2 execution JSON now has a recurring native wire-seam regression that pins its ToolError classification and refusal-specific diagnostics, with Plan 01 published as fully automated OBS-04 coverage.

## Accomplishments

- Added `TestDecodeExecutionSchema2AdmissionSeam`, which serializes a canonical multi-function document with `execution.CanonicalBytes`, decodes it through `decodeExecution`, and preserves invocation-qualified event identity.
- Pinned `native.invalid_execution` plus distinct diagnostic context for non-canonical invocation, duplicate `(invocation,id)`, unknown event kind, and missing or extra caller-owned callee identity.
- Added Plan 01 coverage metadata that names the integration seam and is accepted by the UAT coverage classifier with no human judgment.

## Verification

Passed:

- `go test ./internal/compiler/native -run 'TestDecodeExecutionSchema2AdmissionSeam|TestValidateExecutionSchema2' -count=1 -v`
- `node ~/.codex/gsd-core/bin/gsd-tools.cjs uat classify-coverage --summary .planning/phases/15-event-identity-lang-execution-2/15-01-SUMMARY.md` — `mode: coverage`, `total: 1`, `all_auto_covered: true`, no errors.

## Task Commits

1. `e1e1e63` — `test(15-08): cover schema 2 decoder admission seam`
2. `f33f7ae` — `docs(15-08): publish schema admission coverage`

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- Confirmed `internal/compiler/native/native_test.go`, Plan 01 coverage metadata, and this summary exist.
- Confirmed task commits `e1e1e63` and `f33f7ae` exist in Git history.

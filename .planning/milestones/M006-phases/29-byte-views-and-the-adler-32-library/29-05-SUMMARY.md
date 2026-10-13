---
phase: 29-byte-views-and-the-adler-32-library
plan: 05
subsystem: native-fixtures
tags: [fixture-owner, c-adapter, byte-views, differential]
requires: [29-01]
provides: [fixture-bytes-owner, go-fixture-twin, pinned-expected-checksums]
affects: [29-07, 29-10]
key-files:
  created:
    - examples/phase29/adapter.h
    - examples/phase29/adapter.c
    - examples/phase29/fixture_bytes.bindings.json
    - examples/phase29/README.md
    - internal/compiler/testsupport/phase29_fixture_bytes.go
    - internal/compiler/native/native_phase29_fixture_bytes_test.go
  modified: []
key-decisions:
  - "Case ids 0..9 follow the plan table; unknown ids refused with status 1 and a zeroed owner"
requirements-completed: [APP-10]
status: complete
actuals:
  tokens: 14000
  tasks: 2
  commits: 2
plan_head_before: a2ee6b24685f8ced818952dac49aa984b6bb9f04
plan_head_after: b501c8b3d187dd0984b96c3819abb9ec20463fbf
completed: 2026-10-11
---

# Phase 29 Plan 05: Fixture bytes owner Summary

Test-only C fixture owner (acquire/view/release keyed by case id 0..9) with an independently written Go twin, proven byte-identical across all ten cases, plus unknown-case refusal and an ASan release lane.

## Tasks

1. Tracer: adapter.h/.c, manifest, Go generator, equality test (d867bc6).
2. Unknown-case refusal (10 and 2^64-1), ASan acquire/release over every case, README (b501c8b).

## Deviations from Plan

None - plan executed exactly as written.

## Verification

- TestPhase29FixtureBytesGeneratorsAgree and TestPhase29FixtureBytesUnknownCaseRefused (refusal and asan subtests) pass; ASan was available, so no PARTIAL sanitizer note is needed.
- Session meta-tests (Inventory|Groundedness|Reconciliation|ValidationLifecycle) pass across ./internal/...
- Acceptance greps: no schway_file_byte or adler wording in the adapter; manifest check passes; README has no command lines; examples/phase23 untouched.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

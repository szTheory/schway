---
phase: 28-exit-classes-verify-report-and-entry-pruning
plan: 01
subsystem: testing
tags: [corpus-manifest, exit-classes, entry-resolution, baseline]
requires: []
provides:
  - testdata/phase28/entry-exit-class-manifest.json behavior-class golden at PLAN_BASE
  - phase28ClassChanges / phase28NewlyAddedEntries allow-lists for later Phase 28 plans
  - 28-01-EXIT-PIN-INVENTORY.md exit-pin inventory and corpus audit
affects: [28-02, 28-03, 28-04, 28-05, 28-06, 28-07]
tech-stack:
  added: []
  patterns: [env-guarded golden writer, missing-clang runner sweep, allow-list with stale guard]
key-files:
  created:
    - internal/compiler/session/session_phase28_manifest_test.go
    - testdata/phase28/entry-exit-class-manifest.json
    - .planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-01-EXIT-PIN-INVENTORY.md
  modified: []
key-decisions:
  - "Manifest sweep uses a missing-clang runner so it is clang-free and identical on macOS and Linux"
  - "phase28MinCorpus set to 155, the observed count at generation"
requirements-completed: [DX-17, DX-18]
status: complete
commits: 2
plan_head_before: c2a6f719ec4b823b330781949b7d37682ef0f9c2
plan_head_after: 7126713
actuals:
  tokens: 30000
  tasks: 2
  commits: 2
duration: 30min
completed: 2026-10-10
---

# Phase 28 Plan 01: Entry and exit-class manifest baseline Summary

A 155-file behavior-class golden (check, build, interpreter run, native run, entry shape, entry, dead functions) generated at PLAN_BASE c2a6f71, plus an exit-pin inventory and corpus audit; no production source changed.

## Tasks

1. Tracer: `TestPhase28EntryExitClassManifest` with golden `testdata/phase28/entry-exit-class-manifest.json` (commit ae2a817). Revision field is PLAN_BASE. Entry shapes: single-root 87, unchecked 62, multi-root-main 5, multi-root-no-main 1.
2. Exit-pin inventory and baseline-run record (commit 7126713).

## Verification

- Focused run of TestPhase28EntryExitClassManifest, TestPhase273ConformanceCorpusByteIdentity, TestPhase272AcceptedCorpusByteIdentity and TestPhase16PublicEmitterConsumerInventory: all pass at PLAN_BASE.
- Session meta-tests (Inventory, Groundedness, Reconciliation, ValidationLifecycle) pass clean after each commit.
- `go vet ./internal/compiler/session` clean; no `cgen.` calls in the new test file.

## Research cross-check (all matched)

- payload_borrow_interaction: multi-root-no-main, core.entry_ambiguous, check exit 0.
- multi_function_zero_call: multi-root-main, core.entry_ambiguous.
- duplicate_function_name: check exit 2 core.duplicate_function_id; build and interpreter run exit 3.
- foreign_origin_omitted: native run exit 3.
- Dead-function programs: exactly the four research named. Tiebreak-without-main resolving programs: 0.

## Observations worth carrying forward

- Under the missing-clang runner, native-run failures classify as exit 3 `native.tool_failure` and build failures as exit 3 `native.build_failed`; interpreter run on an unresolved entry is exit 3 `tool.run_failed`. These are the recorded baseline values, not predictions.
- examples/checksum.schway is `unchecked` (check refuses it) at PLAN_BASE.

## Deviations from Plan

None - plan executed as written. A BSD `sed` cleanup of an unused import failed mid-task and was redone with a Python edit before the commit; no effect on committed content.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

Created files and both commits (ae2a817, 7126713) verified present.

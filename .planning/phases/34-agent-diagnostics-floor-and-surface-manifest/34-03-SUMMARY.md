---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "03"
subsystem: diagnostics
tags: [go, diagnostics, module-loader, source-snapshot, json]

# Dependency graph
requires:
  - phase: 34-01
    provides: structured diagnostic identity, units, causes, and repairs
  - phase: 34-02
    provides: captured-source coordinate mapping and bounded human rendering
provides:
  - imported diagnostics presented from the same module snapshots retained by checking
  - source labels and coordinates selected independently for each effective unit
  - focused evidence for immutable snapshots, nested spans, and unmappable peers
affects: [34-04, 34-05, 34-06, 34-07, 34-08, check CLI, diagnostic clients]

actuals:
  tokens: 26331.75
  tasks: 2
  commits: 4
commits: 4
plan_head_before: 7d466a4882403dd8f1b49f94f1fd6da60cd24abd
plan_head_after: 506f469a69d348bcffedc19d0a4228a803a968ee

# Tech tracking
tech-stack:
  added: []
  patterns:
    - per-invocation presentation sidecar joins loader-confirmed source labels with the parsed snapshots
    - diagnostic coordinates are projected from captured bytes without changing IDs or edit spans

key-files:
  created:
    - internal/compiler/session/session_presentation.go
    - internal/compiler/session/session_phase34_presentation_test.go
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-task1-tdd-red-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-task1-tdd-green-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-task1-tracer-final-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-task2-unexpected-green-evidence.json
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_modules.go
    - internal/compiler/protocol/presentation.go

key-decisions:
  - "Keep presentation invocation-local and attach the loader's existing bytes and relative labels without another read or parse."
  - "Use each primary, cause, or repair EffectiveUnit to choose source bytes; retain dotted JSON units and byte-only diagnostic identity."
  - "Task 2 needed no additional production change because Task 1's loaded-source sidecar and the Phase 34-02 projector already satisfied its expanded behavior contract."

requirements-completed: [DX-20]

coverage:
  - id: D1
    description: Public check locates imported syntax and checker diagnostics from their captured source while preserving dotted units and relative labels.
    requirement: DX-20
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase34_presentation_test.go#TestPhase34ImportedDiagnosticPresentation
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_phase34_presentation_test.go#TestPhase34PresentationUsesCapturedSource
        status: pass
    human_judgment: false
  - id: D2
    description: Loaded snapshots locate imported header and origin-peer causes, map nested spans by effective unit, and keep output stable after files are removed.
    requirement: DX-20
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase34_presentation_test.go#TestPhase34PresentationUsesCapturedSource
        status: pass
    human_judgment: false
  - id: D3
    description: Wrong-unit substitution, missing sources, invalid nested spans, nil spans, and source-less peer refusals remain visible as negative controls without fabricated locations.
    requirement: DX-20
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase34_presentation_test.go#TestPhase34PresentationUsesCapturedSource
        status: pass
    human_judgment: false

duration: 14 min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 03: Imported Diagnostic Presentation Summary

**Schway check renders imported diagnostics from retained module snapshots with relative labels, unit-aware spans, and stable text and JSON projections.**

## Performance

- **Duration:** 14 min from the first task commit through summary creation.
- **Started:** 2026-10-13T22:12:31Z (first task commit; the executor start timestamp was not retained across context compaction).
- **Completed:** 2026-10-13T22:26:22Z.
- **Tasks:** 2.
- **Files modified:** 13, including focused evidence records and JUnit reports.

## Accomplishments

- Carried the loaded module inventory into `CheckResult.Presentation`, including parse-error returns and peer-refusal results, without adding a file read or reparse.
- Rendered imported syntax, checker, path-mismatch, and origin-peer diagnostics against loader-confirmed relative labels and the bytes for each effective unit; preserved byte spans, dotted units, and diagnostic identity.
- Proved file deletion after checking leaves text and JSON unchanged, the wrong-unit inventory substitution breaks the exact imported-label control, reads remain one per unit, and unavailable peer positions stay unlocated.

## Task Commits

Each task's changes were committed atomically:

1. **Task 1: Imported syntax error through public checked result** — `18e10fd5` RED test/evidence, `b4bb996d` corrected source-coordinate expectation/evidence, `812b8a94` production implementation and GREEN evidence.
2. **Task 2: Multi-unit nested spans, peer provenance and immutable snapshots** — `506f469a` focused tests and evidence; no additional production change was required after the unexpected-green inspection.

The plan metadata commit is recorded after the self-check.

## Files Created/Modified

- `internal/compiler/session/session_presentation.go` — Joins loader-confirmed labels and parsed bytes into a per-invocation presentation inventory.
- `internal/compiler/session/session.go` — Carries the presentation sidecar on check results and checked command results.
- `internal/compiler/session/session_modules.go` — Retains loaded trees and snapshots on checker and loader diagnostic paths.
- `internal/compiler/protocol/presentation.go` — Preserves the established `unit=... [byte-range]` text prefix while appending the new source location.
- `internal/compiler/session/session_phase34_presentation_test.go` — End-to-end imported diagnostic, nested effective-unit, mutation, and negative-control tests.
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-task*-*-evidence.json` and matching `.junit.xml` — Actual Go reporter outputs with raw events, exit codes, freshness, source, and semantic assessment.

## Decisions Made

- Kept the sidecar internal to checked results; presentation data remains outside core IR, JSON schema, and diagnostic/result identity.
- Used the existing loader source inventory and parse trees as the only source of bytes and import labels; the CLI path remains authoritative for the entry label.
- Reused Phase 34-02's generic span projector for nested causes and repairs rather than adding session-specific formatting or semantic rules.

## TDD Gate Compliance

- **Task 1 RED:** The real `TestPhase34ImportedDiagnosticPresentation` run failed at the intended assertion because the imported primary location was `0:0` rather than `8:15`. `gsd_run check tdd-red-evidence` returned `RED_EVIDENCE_OK` for the preserved record.
- **Task 1 GREEN:** The focused selector for `TestPhase34ImportedDiagnosticPresentation`, `TestPhase30SingleIOSite`, and `TestPhase30ImportedFileDiagnosticNamesFile` passed with exit 0, 12 named events, and 1,436.976 ms. The prior single-I/O and imported-file controls passed.
- **Tracer feedback:** The same focused verification passed after commit `812b8a94`, with exit 0, 12 named events, and 369.063 ms.
- **Task 2 unexpected green:** The real focused selector for both Phase 34 tests passed with exit 0, four named events, and 1,795.581 ms before any Task 2 production edit. Inspection confirmed that the Task 1 sidecar plus Phase 34-02's `EffectiveUnit` projection already supplied the planned behavior. The record is marked `unexpected_green`; no RED was fabricated and no duplicate production code was added. Its file-deletion and wrong-unit controls passed, with the original inventory restored and one loader read per unit.
- The standard-library reporter ran actual `go test -json` commands and retained raw events, command exit, duration, report freshness, reporter source, and semantic assessment.

## Deviations from Plan

**1. [Rule 1 - Bug] Preserved the existing human diagnostic prefix**
- **Found during:** Task 1.
- **Issue:** Adding the imported `file:line:column` projection displaced the established `unit=lib [byte-range]` text that an existing compatibility test checks.
- **Fix:** Kept the unit and byte range before the new source location in `protocol/presentation.go`.
- **Files modified:** `internal/compiler/protocol/presentation.go`.
- **Verification:** `TestPhase30ImportedFileDiagnosticNamesFile` passed in the focused GREEN and tracer runs.
- **Committed in:** `812b8a94`.

**2. Task 2 was already satisfied by the Task 1 implementation.** The target and its path-mismatch, origin-peer, nested-span, mutation, and wrong-unit controls all passed before any Task 2 production change. The test and unexpected-green evidence were committed without manufacturing a failing assertion.

**Total deviations:** 1 auto-fixed compatibility issue and 1 already-satisfied Task 2 implementation contract.
**Impact on plan:** The production scope stayed within the snapshot-presentation contract; Task 2 added focused evidence without duplicating existing projection logic.

## Issues Encountered

- The installed evidence classifier did not consume Go's native JSON event stream. The existing standard-library reporter from Phase 34-02 was reused; it generated JUnit from the real test events and retained the original event stream.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Imported diagnostics are now available through the checked command result. The follow-on diagnostic consumers can use the same invocation-local source inventory, and no blocker remains for this plan.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*

## Self-Check: PASSED

- All created source, test, reporter-evidence, and summary files exist.
- All four plan task commits are ancestors of `HEAD`.
- The latest commit contains no file deletions.

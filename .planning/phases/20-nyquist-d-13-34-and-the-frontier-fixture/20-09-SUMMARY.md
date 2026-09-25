---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "09"
subsystem: validation-evidence
tags: [qlt-10, prc-02, archived-validation, evidence-grade]
requires:
  - phase: 20-02
    provides: corrected Phase 07/08/11 rows for post-reconciliation evidence scanning
  - phase: 20-03
    provides: zero-draft lifecycle enforcement and interim groundedness findings
provides:
  - Four archived rows rewritten to exact current tests
  - A live row-reading grade probe for Phase 12's repaired citation
affects: [20-07, 20-10, QLT-10, PRC-02]
actuals:
  tokens: 0 # Executor token telemetry was unavailable in the shared-worktree handoff.
  tasks: 1
  commits: 2
plan_head_before: bee2c21
tech-stack:
  added: []
  patterns: [live archived-row parser assertion]
key-files:
  created: []
  modified:
    - .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md
    - .planning/milestones/M001-phases/06-agent-feedback-and-performance-ratification/06-VALIDATION.md
    - .planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md
    - internal/compiler/session/evidence_grade_test.go
key-decisions:
  - "The final corpus record moves to Plan 10 after Phase 20 resolves its live status and grade-map failures; recording earlier produced a complete but correctly rejected artifact."
requirements-completed: [QLT-10]
coverage:
  - id: A1
    description: Phase 12's broad package citation is replaced by six exact current test names and the probe reads the live row.
    requirement: QLT-10
    verification:
      - kind: unit
        ref: internal/compiler/session.TestValidationGradeCapBarePackageRowHasNoNamedTest
        status: pass
    human_judgment: false
  - id: A2
    description: Retired Phase 04 citations are absent while replacement commands for Phases 04 and 06 resolve.
    requirement: QLT-10
    verification:
      - kind: unit
        ref: internal/compiler/session.TestValidationGradeCapArchivedDeadCitationsRemainAbsent
        status: pass
    human_judgment: false
duration: pending serialized closeout
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 09: Repair Archived Validation Citations

Four archived rows now name current tests, and the Phase 12 grade probe reads the exact live row rather than a copied command string. The Phase 20 wide corpus record is deferred to Plan 10 because other active validation tables still need status and grade-contract reconciliation.

## Accomplishments

- Replaced the dead Phase 04 names and broad Phase 12 package command with exact live tests.
- Updated Phase 06 rows to the current metric-gate tests.
- Made the grade probe assert the six exact test identifiers parsed from the archived row.
- Verified the retired-citation control and live-row parser probe.

## Verification

- `GOCACHE=/private/tmp/phase20-gocache go test ./internal/compiler/session -run 'TestValidationGradeCapBarePackageRowHasNoNamedTest|TestValidationGradeCapArchivedDeadCitationsRemainAbsent' -count=1` — passed.
- Phase 09 and Phase 10 plan structures — valid with no warnings.

## Plan status

Plan 09 source edits are committed as `bee2c21` and `e359ceb`. The final provenance-bound corpus record and its archive debt closures are intentionally scheduled after Plan 07 in Plan 10.

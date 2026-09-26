---
phase: 15-event-identity-lang-execution-2
plan: "09"
subsystem: session differential and CI evidence
tags: [session, execution-schema-2, ci, coverage, four-tier-differential]
requires:
  - phase: 15-06
    provides: schema-2-peer-gate and legacy comparator routing
  - phase: 15-07
    provides: four-tier diamond and collision controls
  - phase: 15-08
    provides: schema-2 decoder admission seam
provides:
  - direct legacy program-aware wrapper preservation regression
  - cross-platform current aggregate CI evidence lane
  - machine-readable all-automated coverage for peer and diamond gates
affects: [verify-work, phase-15-uat, ci-provenance]
actuals:
  tokens: 3089
  tasks: 2
  commits: 3
tech-stack:
  added: []
  patterns: [non-recursive-ci-source-pin, focused-cross-platform-evidence-aggregate, machine-readable-coverage]
key-files:
  created:
    - .planning/phases/15-event-identity-lang-execution-2/15-09-SUMMARY.md
  modified:
    - internal/compiler/session/session_phase5_compare_test.go
    - internal/compiler/session/session_phase6_test.go
    - .github/workflows/ci.yml
    - .planning/phases/15-event-identity-lang-execution-2/15-06-SUMMARY.md
    - .planning/phases/15-event-identity-lang-execution-2/15-07-SUMMARY.md
key-decisions:
  - "Retain scripts/verify-phase6.sh as historical CI baseline while a separate focused aggregate supplies current Phase 15 provenance."
  - "Use source inspection rather than recursively executing CI because the historical baseline already invokes the full Go suite."
requirements-completed: [OBS-01, OBS-02, OBS-03, OBS-04, NAT-10]
coverage:
  - id: D1
    description: "Schema 0 and Schema 1 program-aware wrapper behavior remains comparator-equivalent and bypasses the Schema 2 peer."
    requirement: OBS-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestPhase5CompareProgramEnginesPreservesLegacySchemas"
        status: pass
    human_judgment: false
  - id: D2
    description: "Current cross-platform CI provenance names schema admission, peer, legacy-wrapper, diamond, and collision seams while retaining the historical baseline."
    requirement: NAT-10
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestCIWorkflowRunsCurrentAggregateGate"
        status: pass
    human_judgment: false
duration: 12min
completed: 2026-09-22
status: complete
commits: 3
plan_head_before: f57cb0022699dfc50d9427cabe7f2aa37b32c94b
---

# Phase 15 Plan 09: Gap-Closure Evidence Summary

The legacy comparator boundary, Schema 2 peer controls, and four-tier shared-leaf diamond are now explicitly pinned as automated cross-platform CI evidence.

## Performance

- **Duration:** 12 min
- **Started:** 2026-09-22T09:52:44Z
- **Completed:** 2026-09-22T10:04:00Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added direct Schema 0/1 wrapper tests that prove peer bypass and comparator-equivalent matching and disagreement behavior.
- Replaced the stale current-phase presentation with a durable aggregate that preserves `scripts/verify-phase6.sh` and runs five focused Phase 15 seams on Ubuntu and macOS.
- Added valid all-automated coverage blocks to Plans 06 and 07, removing their deterministic human-UAT fallback.

## Verification

- `go test ./internal/compiler/session -run 'TestDecodeExecutionSchema2AdmissionSeam|TestSchema2ComparisonRequiresPeerVerdict|TestPhase5CompareProgramEnginesPreservesLegacySchemas|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert' -count=1 -v` passed.
- `go test ./internal/compiler/session -run 'TestCIWorkflowRunsCurrentAggregateGate|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert' -count=1 -v` passed.
- `uat classify-coverage` classified both amended Plan 06 and Plan 07 summaries as `coverage`, with two valid automated entries each and `all_auto_covered: true`.

## Task Commits

1. **Task 1: Trace legacy wrapper preservation from comparator through CI and coverage classification** - `4de4c95` (`test`)
2. **Task 2: Extend the durable aggregate with the four-tier diamond and pin its provenance** - `2ce44fd` (`test`, RED) and `4b6255c` (`feat`, GREEN)

## Decisions Made

- Kept the Phase 6 script as historical baseline coverage; the focused aggregate adds current evidence rather than duplicating its full-suite invocation.
- Kept CI provenance enforcement non-recursive by inspecting checked-in workflow text.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Known Stubs

None.

## Self-Check: PASSED

- Confirmed all five modified deliverables and this summary exist.
- Confirmed task commits `4de4c95`, `2ce44fd`, and `4b6255c` exist in Git history.

---

*Phase: 15-event-identity-lang-execution-2*
*Completed: 2026-09-22*

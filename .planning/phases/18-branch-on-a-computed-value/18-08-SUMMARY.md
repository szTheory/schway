---
phase: 18-branch-on-a-computed-value
plan: 08
subsystem: testing
tags: [go, ci, validation, latency, phase18]
requires:
  - phase: 18-branch-on-a-computed-value
    provides: Phase 18 source fixtures, independent validators, differential checks, and mutation controls
provides:
  - Repeated cold/warm timing distributions for five verification lanes
  - Standard-library evidence verifier for distributions, provenance, and CI disposition
  - Measured decision to retain existing full and race CI coverage without a duplicate focused step
affects: [phase-18-verification, ci, validation]
actuals:
  tokens: 2311
  tasks: 2
  commits: 2
  plan_head_before: 479750b0aba7c68c59ace4a301a34864362267b0
tech-stack:
  added: []
  patterns: [phase-local standard-library validation verifier, isolated cold and shared warm Go cache measurements]
key-files:
  created:
    - .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh
  modified:
    - .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
key-decisions:
  - "Do not add a focused CI command: the existing Ubuntu and macOS jobs already run the complete test and race suites, and the duplicate focused native session command takes about two minutes per host."
patterns-established:
  - "Record sample vectors and matching distribution statistics alongside host/tool provenance and the pre-decision CI blob."
requirements-completed: []
coverage:
  - id: D1
    description: "Five verification lanes have three cold and three warm passing samples, with internally checked distributions and provenance."
    verification:
      - kind: unit
        ref: "bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh --measurements"
        status: pass
      - kind: integration
        ref: "bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh --final"
        status: pass
    human_judgment: false
  - id: D2
    description: "The CI disposition is justified from measured cost, and the existing macOS/Linux full and race lanes remain byte-identical."
    verification:
      - kind: unit
        ref: "bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh --final"
        status: pass
    human_judgment: false
duration: 64min
completed: 2026-09-24
status: complete
---

# Phase 18 Plan 08: Validation Cost and CI Disposition Summary

Five verification lanes now have repeatable cold/warm distributions, and the measured overlap supports keeping the existing full and race CI lanes without adding duplicate focused coverage.

## Performance

- **Duration:** 64 minutes
- **Started:** 2026-09-24T19:17:38Z
- **Completed:** 2026-09-24T20:21:22Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Recorded three cold and three warm successful samples for focused Phase 18 session tests, `go vet ./...`, `go build ./...`, `go test ./... -count=1`, and `go test -race ./... -count=1`.
- Added a standard-library-only verifier that checks exact commands, sample counts, numeric distribution consistency, host/tool provenance, the pre-decision CI blob, and the final disposition.
- Kept `.github/workflows/ci.yml` byte-identical. Both Ubuntu and macOS checks already run the full Go suite and race suite, which include the Phase 18 tests.

## Task Commits

1. **Task 1: Measure focused and full verification feedback cost** - `43ef196` (test)
2. **Task 2: Add only justified recurring focused coverage to existing CI** - `04a590f` (docs; no CI change was justified)

**Plan metadata:** recorded after state updates.

## Files Created/Modified

- `.planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md` — all five cold/warm sample vectors, statistics, provenance, CI hash, and measured decision.
- `.planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh` — measurement and final-disposition verifier.

## Decisions Made

The focused session command has a cold median of 123.411s and a warm median of 114.901s. The current CI checks already run `go test ./...` and `go test -race ./...` on Ubuntu and macOS, and both execute the Phase 18 session package. Adding a separate focused command would repeat about two minutes of coverage per host without a distinct acceptance signal.

## Deviations from Plan

The environment denied writes to the default Go build cache under the user Library directory. Cold samples therefore used fresh isolated caches and warm samples used the writable shared cache at `/tmp/ai-lang-gocache`. The initial warmup permission failure was not counted; all required recorded samples and warmups subsequently passed. No external dependency or runtime threshold was added.

## Issues Encountered

The GSD commit helper initially could not create `.git/index.lock` under the workspace sandbox. Retrying the same scoped GSD operation with repository-write escalation succeeded. No unrelated files were staged.

The GSD roadmap progress updater reported `missing_phase_details` because `ROADMAP.md` has no writable Phase 18 entry. It left the roadmap unchanged; the canonical execution position is recorded in STATE.md.

## User Setup Required

None.

## Next Phase Readiness

Plan 08 is complete. CTL-01 through CTL-03 remain open until phase-level verification checks the phase acceptance evidence; this plan does not independently close those requirements.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-24*

## Self-Check: PASSED

The summary file exists, both per-task commits are present, both evidence verifier modes pass, and the workflow remains byte-identical to the recorded CI blob.

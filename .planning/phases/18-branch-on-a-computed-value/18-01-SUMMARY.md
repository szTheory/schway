---
phase: 18-branch-on-a-computed-value
plan: 01
subsystem: compiler-testing
tags: [go, parser, computed-match, source-fixtures, ownership]
requires: []
provides:
  - Four Phase 18 source witnesses for computed match, Result return, payload return, and loan liveness
  - Baseline parser-refusal tests that pin the diagnostic at the terminal match token
affects: [phase-18-plans-02-through-08]
actuals:
  tokens: 1768
  tasks: 3
  commits: 3
tech-stack:
  added: []
  patterns: [source-fixture frontier tests, production-path diagnostic span pinning]
key-files:
  created:
    - testdata/phase18/computed_match.lang
    - testdata/phase18/result_computed_match.lang
    - testdata/phase18/payload_return.lang
    - testdata/phase18/loan_across_branch.lang
    - internal/compiler/check/check_phase18_test.go
    - internal/compiler/session/session_phase18_test.go
  modified: []
key-decisions:
  - "Pin the current parser refusal syntax.expected_linear_result at the terminal match token because the existing grammar rejects the new source shape before checker admission."
patterns-established:
  - "Each computed-match fixture test loads the source and checks the first diagnostic code and exact token span."
requirements-completed: []
coverage:
  - id: D1
    description: "Computed terminal-match refusal is pinned through the source parser/check entry point."
    requirement: CTL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check -run TestPhase18ComputedScrutineeFrontierPinned -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Result-returning and payload-place source witnesses independently reach the computed-match parser frontier."
    requirement: CTL-02
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestPhase18(ResultFixtureFrontier|PayloadFixtureFrontier)' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: "Production source witnesses the pre-match borrow used in exactly one branch arm."
    requirement: CTL-01
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/check -run TestPhase18LoanAcrossBranchFixture -count=1"
        status: pass
    human_judgment: false
duration: 15min
completed: 2026-09-24
status: complete
plan_head_before: 6fb842ac5ae1c1c6ec5483b48a9b74bd2f6dc2cc
commits: 3
---

# Phase 18 Plan 01: Source Witnesses Summary

**Four source fixtures and focused tests pin the pre-implementation terminal-match boundary and the production S-010 borrow shape.**

## Performance

- **Duration:** approximately 15 minutes; the executor did not capture a monotonic start timestamp
- **Started:** 2026-09-24 (time not recorded)
- **Completed:** 2026-09-24
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added the computed local-place match frontier and pinned its current diagnostic code and exact `match` token span.
- Added separate Result-returning and payload-place source witnesses with tests that prove each file is loaded and reaches the same frontier.
- Added a production `.lang` witness for a borrow created before a terminal branch and consumed in exactly one arm.

## Task Commits

1. **Task 1: Pin the refused computed-place frontier** - `4f40dce` (`test`)
2. **Task 2: Add isolated CTL-02 and CTL-03 source witnesses** - `a228400` (`test`)
3. **Task 3: Add the production S-010 source witness** - `6b28787` (`test`)

## Files Created/Modified

- `testdata/phase18/computed_match.lang` - computed local-place match frontier.
- `testdata/phase18/result_computed_match.lang` - match on a Lang callee's Result return.
- `testdata/phase18/payload_return.lang` - return a destructured payload place from the single-alternative branch.
- `testdata/phase18/loan_across_branch.lang` - pre-match borrow used by one arm.
- `internal/compiler/check/check_phase18_test.go` - frontier and S-010 source checks.
- `internal/compiler/session/session_phase18_test.go` - independent Result and payload frontier checks.

## Decisions Made

- Preserve the actual parser boundary as the baseline. The existing parser does not yet allow a terminal `match` after linear bindings, so the production source path reports `syntax.expected_linear_result` at `match` before checker admission can report `name.unknown_scrutinee`.
- Keep each witness independent so later parser/checker changes can move its frontier without conflating CTL-02, CTL-03, and S-010 evidence.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Pinned the reachable parser frontier instead of an unreachable checker diagnostic**
- **Found during:** Task 1
- **Issue:** `linearBody()` requires a terminal identifier. Source with a local binding followed by `match computed` is refused by parsing before `check.Program` runs, so `name.unknown_scrutinee` cannot be observed on this source path yet.
- **Fix:** Recorded the real `syntax.expected_linear_result` diagnostic and asserted its span covers the terminal `match` token for the primary fixture and all three source witnesses.
- **Files modified:** `internal/compiler/check/check_phase18_test.go`, `internal/compiler/session/session_phase18_test.go`, and the four Phase 18 fixtures.
- **Verification:** Both focused package commands passed; see exact invocations in `coverage`.
- **Committed in:** `4f40dce`, `a228400`, `6b28787`.

**Total deviations:** 1 auto-fixed (Rule 3). **Impact:** The fixture-first gate is now grounded in the production parser. Later plans must move this parser refusal before checker-level computed-place behavior can be proven.

## Issues Encountered

- The default Go build cache under `~/Library/Caches` was not writable in the sandbox. Setting `GOCACHE=/tmp/ai-lang-gocache` allowed the focused tests to run; no repository change was needed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 18-02 can proceed with parser admission work. CTL-01 through CTL-03 are not yet satisfied end to end; these source witnesses are baseline controls only. Continue to keep deterministic acceptance automated and reserve human UAT for irreducible judgment or unavailable external access.

## Self-Check: PASSED

- All four fixture files exist and are non-empty.
- Focused checker and session tests pass.
- Task commits `4f40dce`, `a228400`, and `6b28787` exist in history.
- The three recorded task commits match the persisted plan-head ledger count.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-24*

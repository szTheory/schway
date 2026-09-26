---
phase: 18-branch-on-a-computed-value
plan: 06
subsystem: compiler
tags: [computed-match, ownership, loan-liveness, cfg, interpreter, native-c]
requires:
  - phase: 18-02
    provides: Production computed-match prefix construction.
  - phase: 18-03
    provides: Independent peers for computed scrutinees and provenance.
provides:
  - Production source proof that a pre-match borrow live in one arm uses existing point and edge endpoint kinds.
  - Prefix-aware arm place visibility and CFG liveness with the declared fail-closed bound retained.
  - Four-tier peer/interpreter/native execution for the S-010 source topology.
affects: [phase-18-verification, CTL-01, S-010, loan-liveness]
actuals:
  tokens: 3354
  tasks: 3
  commits: 1
tech-stack:
  added: []
  patterns: ["Computed-match entry prefix participates in the ownership CFG", "Each branch arm receives isolated state seeded from shared prefix places"]
key-files:
  created:
    - internal/compiler/session/session_phase18_loan_test.go
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_phase18_test.go
    - internal/compiler/check/check_test.go
    - testdata/phase18/loan_across_branch.lang
key-decisions:
  - "Keep loan endpoints within the existing point and edge vocabulary; classify divergence on the actual computed-match entry edges."
  - "Keep phase requirements open until the complete Phase 18 verification pass."
patterns-established:
  - "Source-level ownership controls assert the arm-specific endpoint topology, including both-arm and neither-arm controls."
  - "S-010 acceptance includes independent peer admission and interpreter/native differential execution."
requirements-completed: []
coverage:
  - id: D1
    description: "A source borrow created before computed match is used in exactly one arm and materializes one point endpoint plus an edge endpoint on the unused sibling."
    requirement: CTL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check -run 'TestPhase18.*Loan|TestEdgeSpecificLiveOut' -count=1"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/check -run 'TestPhase18.*Loan|TestLoanLiveness(Fixpoint|Bound|Cycle)' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The same accepted S-010 source passes independent peers and agrees across the interpreter, -O0, -O3, and -O3 -flto for both alternatives."
    requirement: CTL-01
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run '^TestPhase18LoanAcrossBranchProduction$' -count=1"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase18' -count=1"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang check testdata/phase18/loan_across_branch.lang (recomputed_work=173)"
        status: pass
    human_judgment: false
metrics:
  duration: 11min
  completed_date: 2026-09-24
status: complete
plan_head_before: 76c7bd1b78560a207dcde9fe59b76e5520d85e6b
commits: 1
---

# Phase 18 Plan 06: Production Loan Across a Computed Branch

**The production checker now carries a prefix-created borrow through computed-match fan-out and proves existing point/edge endpoints across peers and execution engines.**

## Performance

- **Duration:** approximately 11 minutes
- **Started:** 2026-09-24T18:04:00Z (approximate)
- **Completed:** 2026-09-24T18:15:45Z
- **Tasks:** 3 of 3
- **Files modified:** 5 tracked implementation/test/fixture files and 1 new integration test, plus this summary

## Accomplishments

- Made the shared linear prefix part of the production loan-liveness CFG, so a borrow created before computed `match` flows through the real arm fan-out and endpoints refer to actual match edges.
- Seeded each arm's local analysis with shared prefix places and operations while keeping the scrutinee alias arm-local. The positive source witness records one `point` use and one `edge` endpoint on the unused `Off` sibling; both-arm and neither-arm controls reject the one-arm edge classification.
- Re-ran S-010 from source through core and origin peers, interpreter, and native `-O0`, `-O3`, and `-O3 -flto` execution for both alternatives. The source ownership command reports 173 units of recomputed work.
- Retained the existing `4 × blocks × (distinct loans + 1)` work bound, cycle refusal, and zero-valued failure result; existing deterministic boundary tests pass.

## Files Created/Modified

- `internal/compiler/check/check.go` — prefix-aware arm place state and entry-block liveness.
- `internal/compiler/check/check_phase18_test.go` — production endpoint assertions and exclusive-use controls.
- `internal/compiler/check/check_test.go` — adapted the direct internal analyzer call to its prefix inputs.
- `internal/compiler/session/session_phase18_loan_test.go` — independent peer and four-tier S-010 run.
- `testdata/phase18/loan_across_branch.lang` — source witness with a pre-match borrow and computed scrutinee.

## Decisions Made

- Preserve existing `point` and `edge` endpoint kinds and use the match's real entry-edge identities.
- Keep CTL-01 through CTL-03 open until complete Phase 18 verification.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Prefix places and operations were absent from branch arm ownership state and liveness CFG.**
- **Found during:** Task 1
- **Issue:** The fixture parsed, but production checking could not resolve the prefix-created borrow inside an arm, and the liveness graph omitted the computed-match entry prefix and fan-out.
- **Fix:** Seeded arm analysis with the shared prefix facts and included the entry block in the existing bounded liveness fixpoint; edge materialization uses the actual match-edge IDs.
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_phase18_test.go`, `testdata/phase18/loan_across_branch.lang`
- **Verification:** Focused checker/session tests, all Phase 18 checker/session tests, source ownership command, peer validation, and four execution tiers passed.
- **Commit:** Pending parent GSD commit.

**Total deviations:** 1 auto-fixed bug. No scope expansion.

## Issues Encountered

The original fixture included an unnecessary identity call that was refused on the source path. Replacing it with a direct computed local retained the intended computed-match topology and isolated the S-010 ownership question.

## User Setup Required

None.

## Next Phase Readiness

The production S-010 gate is clean for this topology. Requirements remain open for the phase-level verification across all Phase 18 plans.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-24*

## Self-Check: PASSED

- Summary and all planned source/test/fixture files are present.
- Focused checker, liveness-bound, source ownership, and Phase 18 checker/session verification passed.
- Implementation and state are recorded in commit `81f030a` (`feat(18-06): validate loans across computed match`).

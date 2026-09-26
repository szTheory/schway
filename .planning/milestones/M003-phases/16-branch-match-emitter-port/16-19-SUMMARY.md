---
phase: 16-branch-match-emitter-port
plan: 19
subsystem: testing
tags: [evidence, debt-register, machine-probe, budget-audit]

requires:
  - phase: 17-return-type-parameter-type
    provides: Current B1 boundary witness used by the archive reconciliation
provides:
  - Archived Phase 13 probe references that resolve to current evidence while retaining historical decisions
  - Deterministic budget-lane assertions independent of restricted host probe permissions
affects: [phase-16-evidence, debt-registers, budget-audit]

actuals:
  tokens: 3127
  tasks: 2
  commits: 4
plan_head_before: 7c4411313e590c78af3d291acd75ed0e5080d834
commits: 3

tech-stack:
  added: []
  patterns:
    - Inject deterministic MachineFacts into budget audit tests while retaining commandFactory for subprocess probe controls

key-files:
  created:
    - .planning/phases/16-branch-match-emitter-port/16-19-SUMMARY.md
  modified:
    - .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md
    - internal/compiler/session/session_phase6_budget_test.go

key-decisions:
  - "Keep the existing generated UNREACHABLE-CLAIMS disposition logic; the synchronized archived witness source now resolves to the current Phase 17 test."
  - "Use fixture-derived MachineFacts and the existing QLT02LaneFromRows seam for budget assertions; keep actual subprocess behavior in bounded commandFactory tests."

requirements-completed: [NAT-08, NAT-09]
coverage:
  - id: D1
    description: Archived Phase 13 witness citations resolve to the Phase 17 B1 witness and the generated claims view remains current.
    requirement: NAT-08
    verification:
      - kind: unit
        ref: go test ./internal/compiler/session -run '^(TestDebtRegistersAreWellFormed|TestUnreachableClaimsViewIsCurrent|TestNoSuppressionOutlivesItsWitness)$' -count=1 -v
        status: pass
    human_judgment: false
  - id: D2
    description: Budget audit tests use deterministic machine facts while bounded probe success and failure classes stay covered.
    requirement: NAT-09
    verification:
      - kind: unit
        ref: go test ./internal/compiler/measure ./internal/compiler/session -run '^(TestMachineProbeIsBounded|TestBudgetLaneCarriesMachineIDAndVerdict|TestBudgetAuditRefusesUndeclaredMachine|TestQLT02InterproceduralGrowthExponent)$' -count=1 -v
        status: pass
    human_judgment: false

duration: 15min
completed: 2026-09-23
status: complete
---

# Phase 16 Plan 19: Evidence Reference and Host Probe Gap Closure Summary

**Phase 13 debt rows now cite the current Phase 17 witness, and budget lane tests use injected machine facts while real probe controls remain deterministic.**

## Performance

- **Duration:** 15 min
- **Started:** Approximately 2026-09-23T22:28:23Z
- **Completed:** 2026-09-23T22:43:23Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Replaced the obsolete Phase 13 test name in D-13-02b and D-13-10a with `TestPhase17B1RequiresUnverifiableDeclaredContract`, documenting the successor witness provenance and preserving the historical disposition text.
- Changed the session budget tests to construct the QLT-02 audit lane from deterministic `measure.MachineFacts`; host probe outcomes remain exercised by bounded fake commands through `commandFactory`.
- Confirmed the generated `.planning/UNREACHABLE-CLAIMS.md` view is current through its existing source-derived test. The view already represented the Phase 17 disposition and required no independent edit.

## Task Commits

1. **Task 1: Reconcile archived Phase 13 probe citations with the current witness** - `67cd99e` (`fix`)
2. **Task 2: Make budget assertions deterministic across restricted host probes** - `fd6a898` (`test`)
3. **Task 2 follow-up: Clarify injected budget audit evidence wording** - `904822b` (`test`)

## Files Created/Modified

- `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md` - Current witness identifiers and provenance note.
- `internal/compiler/session/session_phase6_budget_test.go` - Fixed machine facts feed the budget lane and audit assertions.

## Decisions Made

- Kept the test generator’s special Phase 17 disposition logic. Editing the generated view independently would violate its source-authoritative process.
- Used the package’s existing lane builder with injected facts; no production probe behavior, bounds, privacy rules, or error vocabulary changed.

## Deviations from Plan

**1. The plan’s Phase 13 debt path was stale.** The archive lives at `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md`; the file was updated at that current location.

**2. The full repository verification gate remains red outside this plan’s scope.** The focused plan tests pass, while `go test ./...` fails on unrelated baseline drift. See Issues Encountered.

**3. Roadmap progress updater did not recognize Phase 16’s existing section.** `roadmap.update-plan-progress 16` returned `missing_phase_details` and left ROADMAP.md unchanged; the existing row remains 15/15 Complete. This is recorded for follow-up rather than manually rewriting roadmap history during this gap closure.

**Total deviations:** 3 (1 path correction, 1 unrelated full-suite failure, 1 updater limitation)  
**Impact on plan:** Both scoped gap closures are verified. Broader baseline failures and roadmap bookkeeping remain visible.

## Issues Encountered

- The default Go build cache was not writable under the sandbox, so Go commands used `GOCACHE=/tmp/ai-lang-go-cache`.
- `go test ./...` failed in existing session checks: Phase 11 native gate fixtures are refused as unsupported by-pointer bodies; LANGUAGE-MATURITY.md reports corpus counts 128/4311 instead of 133/4478; a validation-corpus digest is stale; and the pinned verification-groundedness frontier reports additional R2b/unparseable findings. These failures involve files outside this plan’s declared scope and were not changed.
- The checkout contains pre-existing user modifications and untracked historical Phase 16 fixtures; they were preserved.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The Phase 16 UAT evidence and probe-related tests are deterministic and the stale archive witness identifiers resolve. The full repository gate still needs the unrelated fixture and groundedness drift reconciled; the cross-phase WINDOWS ledger records this failure.

## Self-Check
**Result:** PASSED — changed files exist and all three task commits are present in git history.

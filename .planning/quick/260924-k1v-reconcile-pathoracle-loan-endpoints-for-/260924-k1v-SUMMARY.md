---
phase: quick
plan: 260924-k1v
subsystem: compiler pathoracle
tags: [loan-liveness, pathoracle, phase-18, differential-testing]
requires:
  - phase: 18
    provides: computed-match CFG and accepted loan_across_branch fixture
provides:
  - Path-specific loan endpoint synthesis with independent checker and corevalidate agreement
  - Current Phase 18 corpus line count
affects: [pathoracle, phase-18 verification, language maturity]
tech-stack:
  added: []
  patterns: ["Reduce concrete path replay outcomes while preserving same-block inconsistency refusal"]
key-files:
  created:
    - .planning/quick/260924-k1v-reconcile-pathoracle-loan-endpoints-for-/260924-k1v-SUMMARY.md
    - .planning/quick/260924-k1v-reconcile-pathoracle-loan-endpoints-for-/260924-k1v-VERIFICATION.md
  modified:
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/pathoracle_test.go
    - .planning/LANGUAGE-MATURITY.md
    - .planning/STATE.md
key-decisions:
  - "Distinct branch blocks may hold different final references; conflicting death positions within the same block remain fail-closed."
  - "The session corpus gate already includes the Phase 18 fixture, so no session test change was needed."
requirements-completed: []
coverage:
  - id: D1
    description: "Pathoracle agrees with checker and corevalidate on the computed-match loan fixture, including its On point and Off edge."
    verification:
      - kind: test
        ref: "TestOracleAgreesWithComputedMatchLoan"
        status: pass
    human_judgment: false
  - id: D2
    description: "Path enumeration caps, malformed terminal refusal, composition guards, corpus agreement, and maturity count are verified."
    verification:
      - kind: test
        ref: "Focused pathoracle/session tests and complete pathoracle/session packages"
        status: pass
    human_judgment: false
  - id: D3
    description: "The repository-wide Go suite passes."
    verification:
      - kind: test
        ref: "GOCACHE=/tmp/ai-lang-gocache go test ./..."
        status: fail
    human_judgment: false
duration: ~12min
completed: 2026-09-24
status: complete
---

# Quick Plan 260924-k1v: Reconcile Pathoracle Loan Endpoints Summary

**Pathoracle now independently produces the accepted computed-match loan's On-arm point and Off-arm edge endpoints, while the maturity snapshot reports 4,572 corpus lines.**

## Accomplishments

- Replaced the global one-death-per-loan assumption with a bounded reduction over enumerated paths and their replayed loan states.
- Preserved concrete-path enumeration, work accounting, composition handling, and fail-closed caps; conflicting death positions within one block still return `pathoracle.inconsistent_path_death`.
- Added fixture assertions for whole-value agreement with checker and corevalidate, the On-arm move endpoint, and the declared Off successor edge.
- Confirmed the existing corpus-wide three-way gate already includes the Phase 18 fixture; no session test changes were needed.
- Updated the current maturity sentence from 4,577 to the independently derived 4,572 lines.

## Verification

- Focused pathoracle regression and retained terminal, path-cap, composition-depth, composition-cycle, and per-path composition controls — PASS.
- Focused corpus agreement and maturity checks — PASS.
- Full `internal/compiler/pathoracle` and `internal/compiler/session` packages — PASS.
- `git diff --check` on planned files — PASS.
- `GOCACHE=/tmp/ai-lang-gocache go test ./...` — FAIL; the failing package checks and exact observations are recorded in the verification report for parent triage.

## Deviations from Plan

- `internal/compiler/session/session_peer_gate_test.go` was left untouched because the existing corpus gate already walks and checks this fixture.
- No commit was created by this executor; the parent executor owns the GSD scoped commit.

## Changed Files

- `internal/compiler/pathoracle/pathoracle.go`
- `internal/compiler/pathoracle/pathoracle_test.go`
- `.planning/LANGUAGE-MATURITY.md`
- `.planning/STATE.md`
- This summary and the companion verification report.

## Self-Check: PASSED

- Summary and verification report exist.
- No task commit was expected from this executor; parent handles the scoped commit.

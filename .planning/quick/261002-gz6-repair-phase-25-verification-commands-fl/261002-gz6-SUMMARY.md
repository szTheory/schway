---
id: 261002-gz6
phase: quick
plan: 261002-gz6
subsystem: verification
tags: [phase-25, groundedness, ci]
status: complete
completed: 2026-10-02
tasks: 1
commits: 0
key-files:
  modified:
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VERIFICATION.md
  created:
    - .planning/quick/261002-gz6-repair-phase-25-verification-commands-fl/261002-gz6-PLAN.md
    - .planning/quick/261002-gz6-repair-phase-25-verification-commands-fl/261002-gz6-SUMMARY.md
    - .planning/quick/261002-gz6-repair-phase-25-verification-commands-fl/261002-gz6-VERIFICATION.md
key-decisions:
  - "Use a single anchored TestPhase25 prefix for the three Phase 25 package checks so the per-branch audit can verify them without multi-alternative false findings."
---

# Quick Task 261002-gz6 Summary

**The Phase 25 spot-check commands now stay within the repository's pinned verification-groundedness frontier.**

## Accomplishments

- Replaced three disjunctive test selectors in the Phase 25 verification report with the package-scoped `^TestPhase25` selector.
- Preserved the reported checks' scope while making every selector auditable by the independent per-branch test-grounding checks.
- Kept the Phase 25 goal status, M004 completion, and the recorded `$gsd-new-milestone` next command unchanged.

## Verification

- All three corrected package commands passed: `internal/compiler/native`, `internal/compiler/check`, and `internal/compiler/cgen`.
- The groundedness test group passed with the corrected report.
- `GOCACHE=/tmp/schway-phase-handoff-gocache go test -count=1 ./...` passed locally on macOS/arm64.
- `git diff --check` and strict GSD state validation passed.
- Hosted run 37031590255 exposed the stale grouped selectors on both CI hosts; its other packages passed. Fresh hosted checks and evidence aggregates are required on the follow-up PR before merge.

## Deviations

The Phase 6 and Phase 25 evidence scripts were not repeated locally: the follow-up PR's Ubuntu/macOS CI reruns those aggregate gates, while the focused groundedness checks and uncached full Go suite were run locally.

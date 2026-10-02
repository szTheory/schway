---
phase: quick
plan: 261002-awt
subsystem: testing
tags: [phase24, evidence, native-cleanup]
requires: [Phase 24 observer and hosted validation evidence]
provides:
  - Direct README links from the physical-cleanup claim to its native observer and hosted receipt.
  - A focused contract that checks exact links, regular-file targets, and hosted run identity.
affects: [Phase 25 evidence index]
actuals:
  tokens: 885
  tasks: 1
  commits: 2
plan_head_before: 380f6ad5cb3655b216a59eb4fc944134d7d46e6e
commits: 2
tech-stack:
  added: []
  patterns: [Focused standard-library README evidence contract]
key-files:
  created: []
  modified:
    - examples/phase24/README.md
    - internal/compiler/native/phase25_utility_test.go
key-decisions:
  - "Keep physical observer evidence separate from model-only replay and retain hosted run 36856048690 explicitly."
requirements-completed: [EVD-09, RES-06]
duration: 2min
completed: 2026-10-02
status: complete
---

# Quick Task 261002-awt Summary

**Phase 24 physical-cleanup claims now link directly to the native observer and the hosted two-host validation receipt.**

## Accomplishments

- Added direct relative links to `phase24_observer_test.go` and `24-VALIDATION.md`, naming hosted run `36856048690` beside the cleanup claim.
- Extended `TestPhase25EvidenceIndex` to require both exact links in the Phase 24 transfer and cleanup paragraph, verify each target is a regular file, and confirm the receipt contains the hosted run ID.
- Preserved the existing evidence-index link checker and missing-link negative control.

## Verification

`GOCACHE=/tmp/ai-lang-verification-gocache go test -v -count=1 -run '^TestPhase25EvidenceIndex$' ./internal/compiler/native`

Result: PASS — output included `--- PASS: TestPhase25EvidenceIndex`.

## Task Commits

- `2f91e99` — `docs(261002-awt): link Phase 24 physical cleanup evidence`
- `48a1fbb` — `test(261002-awt): scope Phase 24 links to cleanup paragraph`

## Deviations from Plan

None. The test assertion was scoped to the named evidence paragraph so moving a required link elsewhere in the README fails the contract.

## Scope and Working Tree

Only `examples/phase24/README.md` and `internal/compiler/native/phase25_utility_test.go` were included in implementation commits. The pre-existing edit to `.planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VERIFICATION.md` remains untouched and uncommitted.

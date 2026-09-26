---
phase: 18
plan: 07
subsystem: compiler-validation
tags: [computed-match, peer-validation, comparator, mutation]
dependency_graph:
  requires: [18-04, 18-05, 18-06]
  provides: [independent-origin-refusal, phase18-comparator-non-vacuity]
  affects: [corevalidate, originvalidate, session]
tech_stack:
  added: []
  patterns: [seeded peer refusals, explicit required-engine gate, exact-axis comparator controls]
key_files:
  created:
    - internal/compiler/session/session_phase18_adversarial_test.go
  modified:
    - internal/compiler/originvalidate/originvalidate_phase18_test.go
decisions:
  - Requirements remain open pending final phase verification.
metrics:
  completed_date: 2026-09-24
  duration: "~3 minutes"
status: complete
actuals:
  tokens: 1150
  tasks: 2
  commits: 0
  plan_head_before: 537d4599f181d56abe2f72b9fbb13b1e245806a6
---

# Phase 18 Plan 07: Adversarial Validation Summary

Added controls that make independent origin admission, complete engine participation, peer validation, and each execution comparator axis observable. The existing wrong-slot control continues to require a positive injection count and an exact terminal-outcome disagreement, alongside an agreeing unmutated payload baseline.

## Tasks Completed

1. Added a control that first obtains checker admission, then forges a published origin and requires the independent origin peer to reject it.
2. Added a four-engine acceptance gate, a peer-call count control, and seeded divergences for terminal outcome, event order, resource ledger, and exit status/signal.

## Verification

- `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run 'TestPhase18.*(Forged|Mutation|Origin|Place)' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run 'TestPhase18(ComparatorControl|ComparatorAxes|FiveAxis|PayloadWrongSlot|WrongSlotMutation)' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/session -count=1` — corevalidate and originvalidate passed; session failed in unrelated maturity-count checks and a Phase 18 plan 18-06 loan fixture check described under Deferred Issues.

## Deviations from Plan

None. No production code changes were needed; the existing Phase 18 wrong-slot seam and payload baseline already supplied the required positive mutation and agreement controls.

## Deferred Issues

- `internal/compiler/session/self_describing_docs_test.go` reports that `.planning/LANGUAGE-MATURITY.md:87` states a corpus line count of 4577 while the re-derived value is 4572. This also fails two guard self-tests that expect the unchanged document copy to pass.
- `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` reports `pathoracle.inconsistent_path_death` for `testdata/phase18/loan_across_branch.lang`, which belongs to the completed plan 18-06 rather than this plan's owned files.

## Self-Check: PASSED

The changed test files exist and the focused controls pass. Changes remain uncommitted as requested by the parent executor; no task or metadata commit was made here.

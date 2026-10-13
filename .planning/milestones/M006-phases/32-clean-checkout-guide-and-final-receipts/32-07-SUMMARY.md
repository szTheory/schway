---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 07
subsystem: cmd/schway guide doc-test
tags: [dx-19, doc-test, negative-controls, native-policy]
requires: [32-04]
provides:
  - phase32GuideStaticProblems pure static guide check gating TestPhase32GuideCleanCheckout
  - TestPhase32GuideParserNegativeControls (five in-memory mutations)
  - TestPhase32NativePolicy (hosted fail-not-skip pin)
affects: [32-09]
key-files:
  created: []
  modified:
    - cmd/schway/main_phase32_guide_test.go
key-decisions:
  - "Static check runs before the clang skip, so a drifted guide fails even on hosts without clang"
  - "Existing phase32NativePolicy also treats CI != '' as hosted; kept (superset of the plan's GITHUB_ACTIONS / SCHWAY_REQUIRE_NATIVE rows)"
requirements-completed: [DX-19]
status: complete
commits: 1
plan_head_before: b188b825f9cd1f6e2c46f5af26c0ad36bd572654
plan_head_after: f108bc6e
actuals:
  tokens: 6000
  tasks: 2
  commits: 1
coverage:
  - id: D-04 / D-01a
    description: "The guide checker rejects a changed value, a renamed command, a dropped program fence, a repo-root output target and a dropped table row on in-memory copies; the real README passes the static check before anything executes"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_guide_test.go#TestPhase32GuideParserNegativeControls"
        status: pass
      - kind: unit
        ref: "cmd/schway/main_phase32_guide_test.go#TestPhase32GuideCleanCheckout"
        status: pass
    human_judgment: false
  - id: D-04a
    description: "phase32NativePolicy fails a missing clang under GITHUB_ACTIONS=true or SCHWAY_REQUIRE_NATIVE=1, skips otherwise, and never fails when clang is present"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_guide_test.go#TestPhase32NativePolicy"
        status: pass
    human_judgment: false
duration: 8min
completed: 2026-10-12
---

# Phase 32 Plan 07: Guide checker negative controls Summary

`phase32GuideStaticProblems` now gates the clean-checkout guide test, five in-memory mutation controls prove it rejects drift, and `TestPhase32NativePolicy` pins fail-not-skip for hosted runners without clang.

## Accomplishments

- Pure static check (parse, anti-vacuity, D-01a output targets, `./schway app run ` / `./schway build ` command shapes, JSON and oracle cross-check); real README yields no problems.
- `TestPhase32GuideParserNegativeControls`: mutated-value, renamed-command, dropped-fence, outside-target, dropped-row, each with an anchor-present fatal so none passes vacuously.
- `TestPhase32NativePolicy`: four-row table with a fake getenv.

## Deviations from Plan

**[Process] Single commit for both tasks.** Both tasks edit the same file; Task 2 only adds a test over a function written in 32-04, so the work was committed once (f108bc6e) rather than as two commits. No code impact.

**Total deviations:** 1 (process only).

## Verification

Guide tests (no SKIP), spawn lint, whole `./cmd/schway` package, the session meta set, and `TestValidationRowGradesAreEarnedOverArchivedCorpus` all pass. Acceptance greps: `t.Parallel` 0, `trimpath` 0.

## Known Stubs

None.

## Self-Check: PASSED

---
phase: 26-checked-scalar-sum
plan: 01
subsystem: compiler
tags: [scalar-cfg, u64, bool, native-app, github-actions]
requires:
  - phase: 25
    provides: Existing U64 values, typed CFG, interpreter, C17 backend, and public app runner.
provides:
  - Structured scalar if/else and while lowering with bounded independent CFG validation.
  - Checked U64 sum source witness with interpreter and native app evidence.
affects: [phase-26-follow-on-plans, phase-27, compiler-checker, native-app]
actuals:
  tokens: 17053
  tasks: 2
  commits: 15
plan_head_before: b65e839519f6db6e1d8c6c22dfcd984ce1b90e40
tech-stack:
  added: []
  patterns: [bounded scalar CFG fixed-point analysis, revision-bound focused hosted validation]
key-files:
  created: [examples/sum_to_n.schway, internal/compiler/check/scalar.go]
  modified: [.github/workflows/ci.yml, internal/compiler/ast/ast.go, internal/compiler/syntax/token.go, internal/compiler/syntax/lexer.go, internal/compiler/syntax/parser.go, internal/compiler/syntax/format.go, internal/compiler/core/core.go, internal/compiler/core/core_test.go, internal/compiler/ability/ability.go, internal/compiler/check/check.go, internal/compiler/corevalidate/corevalidate.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/interp/interp.go, internal/compiler/cgen/cgen_program.go, internal/compiler/session/session_phase26_test.go]
key-decisions:
  - "Keep scalar if/else and while as explicit CFG blocks with typed operations."
  - "Use bounded, independently derived scalar fixed-point analyses; ownership and provenance remain outside admitted loop carries."
  - "Keep the path oracle acyclic and use focused hosted Linux/macOS jobs for this plan's suite."
patterns-established:
  - "Checker and peer validators must independently converge under the deterministic transfer budget and fail closed on exhaustion."
  - "Native app defects report bounded stderr and nonzero exit before producing stdout."
requirements-completed: [U64-01, FLOW-01, FLOW-02, APP-07]
coverage:
  - id: D1
    description: Checked-in sum_to_n source runs through the public app route for inputs 0, 10, and 1000, and rejects 1001.
    requirement: APP-07
    verification:
      - kind: integration
        ref: "GitHub Actions run 37073989672, Phase 26 compiler and native app regressions (ubuntu-latest and macos-latest)"
        status: pass
    human_judgment: false
  - id: D2
    description: Scalar branching, checked arithmetic, syntax recovery, core inventory, and bounded validator behavior are covered by focused regression tests.
    requirement: FLOW-01
    verification:
      - kind: unit
        ref: "GitHub Actions run 37073989672, TestPhase26* and TestAllOperationKindsHandledAtEverySite"
        status: pass
    human_judgment: false
metrics:
  duration: 100min
completed: 2026-10-02
status: complete
---

# Phase 26 Plan 01: Checked Scalar Sum Summary

**A source-authored checked sum now flows through scalar CFG validation, interpretation, C17 generation, Clang, and the public app runner.**

## Performance

- **Duration:** 100 min
- **Started:** 2026-10-02T21:03:00Z (approximate)
- **Completed:** 2026-10-02T22:46:25Z
- **Tasks:** 2
- **Files modified:** 17

## Accomplishments

- Added `examples/sum_to_n.schway` and source syntax for mutable U64/Bool locals, assignment, `if/else`, pretested `while`, `<`, and checked `+`.
- Lowered structured scalar control into explicit typed CFG operations and added bounded scalar facts in the checker plus independent core and origin validators.
- Matched interpreter and native C17 behavior through `schway app run`; added exact application, boundary, syntax, operator-span, zero-iteration, and core-inventory tests.
- Added a manually dispatched focused GitHub Actions lane and verified the selected compiler/native regressions on both Ubuntu and macOS at revision `b2034083c2274a2d1d301394def0835047dd5eda`.

## Task Commits

Task 1 followed TDD: source witness first, then the failing integration test, implementation, and fixes discovered by hosted validation. Task 2 added focused acceptance and negative-control coverage.

1. `30c480d` witness; `3eb5a6d` hosted RED test; `7ea10fe` scalar CFG lowering.
2. Follow-up correctness fixes: `adeb557`, `1f0375b`, `3a0cf7e`, `75d1dd9`, `9754f44`, `39590ed`, `7ea12f4`, `70a0bc4`, `d4703a7`, `00457df`, `6fd2c36`, `b203408`.

**Plan commits measured from ledger:** 15 task commits. The documentation/tracking closeout commit follows this summary and is outside the plan's measured task commit count.

## Files Created/Modified

- `examples/sum_to_n.schway` - ordinary source witness for the checked scalar sum.
- `internal/compiler/check/scalar.go` - scalar CFG lowering and bounded fixed-point analysis.
- `internal/compiler/{ast,syntax,core,ability,check,corevalidate,originvalidate,interp,cgen,session}` - syntax, typed operations, independent validation, execution, and integration coverage.
- `.github/workflows/ci.yml` - focused hosted validation lane, dispatched for Linux and macOS.

## Decisions Made

- Scalar loop admission is limited to U64/Bool state and uses a deterministic 65,536-transfer cap; authority-bearing values stay refused across back edges.
- The path oracle remains an acyclic oracle rather than becoming a loop proof.
- Project suites/native execution remain hosted due the standing project constraint; local work was limited to source inspection, formatting, and Git checks.

## Deviations from Plan

### Auto-fixed Issues

Hosted compilation and integration surfaced issues directly tied to the new feature: Bool capability registration, nested formatter boundaries, definite initialization, block replay/claims, schema event projection, and generated C entry-label/legacy helper handling. These were fixed in the relevant task commits and revalidated on both hosts.

**1. [Rule 3 - Blocking issue] Added focused hosted validation support**
- **Found during:** Task 1 verification
- **Issue:** The full CI lane included unrelated existing failures and the project forbids local project-suite execution, so it could not provide isolated revision-bound evidence for this feature.
- **Fix:** Added a manually dispatched focused matrix job for the Phase 26 compiler/native app regressions on Ubuntu and macOS.
- **Files modified:** `.github/workflows/ci.yml`
- **Verification:** GitHub Actions run `37073989672`, SHA `b2034083c2274a2d1d301394def0835047dd5eda`; both focused host jobs passed.
- **Committed in:** `6fd2c36`.

**Total deviations:** 1 focused validation addition plus directly related correctness fixes. **Impact:** Required to produce trustworthy hosted evidence; no new dependency or backend.

## Issues Encountered

The initial hosted RED run `37065010477` failed the intended `TestPhase26TracerAppSum10` witness because the syntax and lowering were not yet implemented. The final focused run `37073989672` passed on both host priorities. Other unrelated checks in the broader workflow were not treated as Phase 26 evidence or repaired here.

## TDD Gate Compliance

- RED witness commit: `3eb5a6d`; hosted run `37065010477` showed the targeted test failing on the unsupported feature, and `check tdd-red-evidence` accepted the revision-bound receipt.
- GREEN: final hosted run `37073989672` passed `TestPhase26*` and `TestAllOperationKindsHandledAtEverySite` on Ubuntu and macOS.
- No local Go tests, builds, or native verification scripts were run.

## User Setup Required

None.

## Next Phase Readiness

Plan 26-01 is complete. Phase 26 remains active; its other plans and ownership/refusal obligations remain in scope. No Phase 27 work was started.

## Self-Check: PASSED

- Summary file exists and includes the plan's four requirement IDs.
- The 15 task commits are measured from ledger base `b65e839519f6db6e1d8c6c22dfcd984ce1b90e40` through `b2034083c2274a2d1d301394def0835047dd5eda`.
- Hosted validation receipt is tied to that exact head SHA and both focused jobs succeeded.

---
*Phase: 26-checked-scalar-sum*
*Completed: 2026-10-02*

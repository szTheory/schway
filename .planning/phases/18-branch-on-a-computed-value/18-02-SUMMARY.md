---
phase: 18-branch-on-a-computed-value
plan: 02
subsystem: compiler
tags: [parser, checker, core, interpreter, native-c, differential-testing]
requires:
  - phase: 18-01
    provides: pinned computed-terminal-match parser refusal and Wave 0 evidence
provides:
  - AST and parser representation for a linear prefix with one terminal match
  - Checker and peer admission support for computed-place branch entry
  - Interpreter, production C emitter, and source-to-native differential tracer
affects: [phase-18, computed-match, checker, corevalidate, originvalidate, interpreter, cgen]
actuals:
  tokens: 8571
  tasks: 2
  commits: 4
tech-stack:
  added: []
patterns: [terminal-match-on-linear-body, entry-block-prefix-execution, stable-scrutinee-place-id, computed-place-negative-controls]
key-files:
  created: []
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/syntax/parser_phase18_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_phase18_test.go
    - internal/compiler/core/core.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session_phase18_test.go
    - testdata/phase18/computed_match.lang
key-decisions:
  - "Represent the new surface form as a terminal match attached to the existing linear body, and carry its computed place ID into core; preserve the single branch law."
  - "Execute computed-prefix operations in the existing entry block before match dispatch in both interpreter and native emitter."
  - "Keep CTL-01 open until the complete Phase 18 acceptance is verified."
patterns-established:
  - "Checker-produced prefix operations feed the existing branch CFG entry block."
  - "Computed terminal scrutinee refusals use a stable checker diagnostic at the match boundary."
requirements-completed: []
duration: 38min
completed: 2026-09-24
status: complete
plan_head_before: 561597d2868bb598c9a0346698321a6fc1a4d9ee
commits: 4
coverage:
  - id: D1
    description: "Computed terminal match flows from production source parsing and checking through independent admission, interpreter, emitted native C, and five-axis agreement."
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run TestPhase18ComputedSourceFourTierDifferential -count=1"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/session -run 'TestPhase18Computed|TestGeneratedBranchBodiesRoundTrip' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Generated round-trip coverage preserves computed terminal matches and stable refusal diagnostics reject out-of-scope, shadowed, and non-data scrutinees."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/syntax ./internal/compiler/check -run 'Test(GeneratedBranchBodiesRoundTrip|Phase18Computed|Phase18LoanAcrossBranchFixture)' -count=1"
        status: pass
    human_judgment: false
metrics:
  duration: 38min
  completed_date: 2026-09-24
---

# Phase 18 Plan 02: Branch on a Computed Value Summary

**A computed terminal match now runs through the existing branch CFG, interpreter, production C emitter, and five-axis source-to-native evidence path.**

## Performance

- **Duration:** 38 minutes
- **Started:** 2026-09-24T15:08:29Z
- **Completed:** 2026-09-24T15:46:37Z
- **Tasks:** 2 of 2
- **Files modified:** 11

## Accomplishments

- Added a terminal-match AST/parser form that preserves source spans and keeps the linear prefix in the existing function body.
- Lowered prefix operations into the branch entry block and carried a stable computed place ID through core admission, interpreter dispatch, and production C emission.
- Added a production fixture exercised by the complete five-axis comparator, plus generated round-trip and refusal controls for missing, shadowed, and non-data scrutinees.

## Task Commits

1. **Task 1: Trace computed terminal match from source through native evidence** — `df8b893` (RED test), `e38c1c4` (feat).
2. **Task 2: Preserve source fidelity and computed-place negative controls** — `33f8c2e` (test), `170bc81` (stable place ID fix).

The plan metadata commit is recorded after this summary is verified.

## Decisions Made

- Added `TerminalMatch` to the existing linear body and reused the branch CFG entry block; no second branch representation was introduced.
- Carried an optional `ScrutineeID` for computed matches so peers and execution paths use place identity; ordinary parameter matches omit it, preserving prior core bytes.
- Refused a terminal match that refers to the parameter instead of a computed local, retaining the existing ordinary parameter-match path.
- Left `requirements-completed` empty because later Phase 18 plans still need to establish full CTL-01 acceptance.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing integration functionality] Wired computed-prefix execution into existing interpreter and native emitter paths.**
- **Found during:** Task 1
- **Issue:** The plan's `files_modified` omitted these execution paths, but the promised end-to-end behavior could not work without running entry-block prefix operations before dispatch.
- **Fix:** Reused existing branch execution and C emission paths for prefix operations and scrutinee selection.
- **Files modified:** `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen_program.go`
- **Verification:** The computed source four-tier differential test passed.
- **Committed in:** `e38c1c4`

**2. [Rule 1 - Diagnostic boundary] Normalized an out-of-scope terminal scrutinee to the stable checker diagnostic.**
- **Found during:** Task 2
- **Issue:** The synthetic linear result surfaced as generic `name.unknown` before the checker could report the terminal match's place refusal.
- **Fix:** Convert that terminal-result failure to `name.unknown_scrutinee` at the match-body span.
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_phase18_test.go`
- **Verification:** Negative controls assert stable diagnostic codes; targeted tests passed.
- **Committed in:** `e38c1c4`, `33f8c2e`

**3. [Rule 1 - Test assumption] Updated the loan fixture test to locate its intended `select` function.**
- **Found during:** Task 2
- **Issue:** The fixture includes an earlier helper function, so selecting `Funcs[0]` inspected a body without a terminal match.
- **Fix:** Locate `select` by name before asserting terminal-match parsing.
- **Files modified:** `internal/compiler/check/check_phase18_test.go`
- **Verification:** `TestPhase18LoanAcrossBranchFixture` passed.
- **Committed in:** `33f8c2e`

**4. [Rule 2 - Stable identity] Preserved the resolved place ID through core and execution.**
- **Found during:** Task 2 verification
- **Issue:** Name-based scrutinee rediscovery did not satisfy the plan's stable-place-identity contract and could select an ambiguous same-name place.
- **Fix:** Added an optional core `ScrutineeID`; peers verify that exact place ID, and interpreter/native dispatch use it. Parameter-only matches omit the field so prior core evidence remains byte-identical.
- **Files modified:** `internal/compiler/core/core.go`, `internal/compiler/check/check.go`, `internal/compiler/check/check_phase18_test.go`, `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen_program.go`
- **Verification:** Focused eight-package Phase 18 tests passed; core, emitter, interpreter, checker, and validators passed full package tests. The full session package reports the unrelated maturity-document drift below.
- **Committed in:** `170bc81`

## Verification

- Passed the Task 1 focused command for syntax, checker, both admission peers, and session differential coverage.
- Passed the Task 2 focused syntax/checker command, including generated round-trip and negative controls.
- Passed focused Phase 18 tests across core, C emitter, interpreter, checker, core validator, origin validator, session, and syntax after adding stable scrutinee IDs.
- Attempted `go test ./...` and a full `-count=1` run of affected packages; both fail only in `internal/compiler/session` because pre-existing `LANGUAGE-MATURITY.md` counts are stale: tests/guard count 45 vs 46, corpus program count 133 vs 137, and corpus line count 4478 vs 4578. This documentation-count drift is unrelated to Plan 02 and was left untouched.

## Deferred Issues

- Full-suite session tests remain blocked by the stale counts in `.planning/LANGUAGE-MATURITY.md`; refresh that document through its owning workflow before using the full-suite result as a clean baseline.

## Self-Check: PASSED

- Summary file exists at the required plan path.
- Task commits `df8b893`, `e38c1c4`, and `33f8c2e` are present in git history.

---
phase: 26-checked-scalar-sum
plan: 05
subsystem: compiler
tags: [schway, scalar-loop, checked-u64, native-app, validation]
dependency_graph:
  requires: [26-02, 26-03, 26-04]
  provides: [source-authored immutable scalar copies in loops, pinned sum oracles and mutations, public app CLI evidence, repeated-copy event evidence]
  affects: [26-checked-scalar-sum, 27-exact-fizzbuzz]
tech_stack:
  added: []
  patterns: [immutable scalar let lowers through OpCopy, source-to-app exact stream oracle, hosted validation selectors by package and test]
key_files:
  created: [internal/compiler/session/session_phase26_event_test.go]
  modified: [.github/workflows/ci.yml, examples/sum_to_n.schway, internal/compiler/check/scalar.go, internal/compiler/syntax/parser.go, internal/compiler/session/session_phase26_test.go, internal/compiler/native/native_app_test.go, cmd/schway/main_test.go, internal/compiler/syntax/syntax_test.go, .planning/phases/26-checked-scalar-sum/26-VALIDATION.md, .planning/phases/26-checked-scalar-sum/26-WAVE-GATES.md, .planning/LANGUAGE-MATURITY.md, testdata/phase16/validation-corpus-run-record.jsonl, testdata/phase16/validation-corpus-run-record.manifest.json]
decisions:
  - Immutable scalar `let` snapshots copy the value into a fresh place, preserving value semantics.
  - Keep frontend routing narrow so unsupported generic call lets remain on the generic parser path.
  - Convert legacy interpreter event documents using the existing schema-2 projection before independent peer validation.
  - Keep app outcome shape unchanged and test capacity exhaustion through the existing evidence-limit seam.
  - Split validation selectors by package/test and grade rows from executed hosted evidence.
plan_head_before: 4439e4ab2f51f5e22d8808c5724098acd2e9c935
commits: 10
actuals:
  tokens: 1573916
  tasks: 2
  commits: 10
requirements-completed: [U64-01, FLOW-01, FLOW-02, APP-07]
coverage:
  - id: D1
    description: Independent literal sum answers, no-output rejection at 1001, and reached accumulation/iteration mutation controls.
    requirement: U64-01
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase26_test.go#TestPhase26ExactSumMatrix
        status: pass
    human_judgment: false
  - id: D2
    description: Ordinary source-authored `let snapshot = i` produces repeated copy events accepted by independent peers and the public app path.
    requirement: FLOW-01
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase26_event_test.go#TestPhase26SourceRepeatedCopyEvidence
        status: pass
    human_judgment: false
  - id: D3
    description: Public app build/run, exact output and overflow tuples, bounded evidence exhaustion, and syntax recovery are covered.
    requirement: APP-07
    verification:
      - kind: e2e
        ref: cmd/schway/main_test.go#TestPhase26PublicAppCLI
        status: pass
    human_judgment: false
metrics:
  duration: 217min
  completed: 2026-10-03
  status: complete
  started_at: 2026-10-02T23:01:12-04:00
  completed_at: 2026-10-03T02:38:33-04:00
  tasks: 2
  files_modified: 14
---

# Phase 26 Plan 05: Checked Scalar Sum Summary

**The source-authored scalar sum now passes through immutable loop snapshots, independent event peers, exact public app oracles, and hosted validation on Linux and macOS.**

## Performance

- **Duration:** 217 min
- **Started:** 2026-10-02T23:01:12-04:00
- **Completed:** 2026-10-03T02:38:33-04:00
- **Tasks:** 2
- **Files modified:** 14

## Accomplishments

- Pinned hand-derived `0`, `55`, and `500500` outcomes for interpreter and native execution, plus the 1001 rejection and reached accumulation/iteration mutation controls.
- Added public `schway app build` / `app run` exact stream checks, direct overflow behavior, forced evidence-capacity exhaustion, and syntax recovery/round-trip controls.
- Enabled the real source `let snapshot = i` loop path as an immutable copy and verified same-site ordinals 0/1/2 through checked source, interpreter projection, native capture, and independent peers.
- Completed hosted CI with run 37101934669 on Ubuntu and macOS, including vet, builds, all implementation and race suites, and Phase 23/24/25 aggregates.

## Task Commits

1. **Task 1: Pin sum values and kill reached mutations** — `fb1f61e` (test), `a671308` (feat), `d3a809f` (fix).
2. **Task 2: Exercise public CLI and evidence separation** — `ced8eba` (test), `19058e6` (fix).

Plan execution also required validation/receipt integration commits: `677a453`, `46a4dfd`, `72a8174`, `b308651`, and `1dc5e63`.

## Files Created/Modified

- `internal/compiler/check/scalar.go` and `internal/compiler/syntax/parser.go` — lower the restricted scalar immutable let to a value copy while keeping unsupported generic syntax on its existing route.
- `examples/sum_to_n.schway` — exercise the snapshot directly in the ordinary loop source.
- `internal/compiler/session/session_phase26_test.go` and `session_phase26_event_test.go` — literal answers, mutations, source-authored events, peer verification, and app evidence.
- `cmd/schway/main_test.go`, `internal/compiler/native/native_app_test.go`, and `internal/compiler/syntax/syntax_test.go` — public CLI, bounded capture, and parser recovery controls.
- `.github/workflows/ci.yml` — install Linux libc++ in the focused job and run command/syntax controls before broad session tests.
- `.planning/phases/26-checked-scalar-sum/26-VALIDATION.md` and `26-WAVE-GATES.md` — grounded validation status and hosted run receipts.
- `.planning/LANGUAGE-MATURITY.md` — refresh source corpus line count.
- `testdata/phase16/validation-corpus-run-record.jsonl` and its manifest — install the actual hosted sequential receipt artifact for the 35-pair selector set.

## Decisions Made

- Scalar immutable bindings are copied into a fresh place; they do not alias mutable loop variables.
- Parser routing only selects the scalar let form where its signature and expression are supported; generic function-call lets retain generic parsing.
- The legacy interpreter event document uses its existing schema-2 projection before execution-peer validation, preserving peer independence.
- Evidence capacity uses the existing runner limit seam; child stdout, stderr, and exit status remain intact when evidence is incomplete.
- The receipt was regenerated from the hosted recorder after the selector set changed from 34 to 35 pairs; the checked-in record was downloaded and installed from the workflow artifact.

## TDD Gate Compliance

Task 1 recorded an honest RED: hosted run 37091806637 failed because the ordinary source `let snapshot = i` was not routed through the frontend/lowerer. Commits `a671308` and `d3a809f` supplied the narrow fix; the later focused hosted runs passed. Task 2 recorded source-to-peer and public app regressions before its fixes; hosted run 37092941501 exposed the schema-projection gap, which was corrected in `19058e6` and passed in subsequent focused runs. The project forbids local Go tests and program runs, so RED/GREEN evidence came from hosted actions. No synthetic local failure was claimed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical Functionality] Added the minimum source parser/lowering path for immutable scalar let copies.**

- **Found during:** Task 1 source witness acceptance.
- **Issue:** The required ordinary `let snapshot = i` source path was refused before it could produce the planned copy evidence.
- **Fix:** Added narrow parser routing and scalar lowering that copies the value into a fresh place; generic call lets retain their existing route.
- **Files modified:** `internal/compiler/syntax/parser.go`, `internal/compiler/check/scalar.go`, `examples/sum_to_n.schway`.
- **Verification:** Actual source regression and full hosted run 37101934669.
- **Committed in:** `a671308`, `d3a809f`.

**2. [Rule 3 - Blocking] Installed the focused job's Linux libc++ dependency.**

- **Found during:** Task 2 focused hosted CI setup.
- **Issue:** Focused Linux session sanitizer coverage requires libc++.
- **Fix:** Reused the existing full-job Ubuntu installation step in the focused job.
- **Files modified:** `.github/workflows/ci.yml`.
- **Verification:** Focused run 37101757588 and full run 37101934669 passed both hosts.
- **Committed in:** `ced8eba`.

**3. [Rule 3 - Blocking] Reconciled validation selectors, grade fields, and the generated hosted receipt.**

- **Found during:** Phase validation integration.
- **Issue:** Mixed package selector groups, missing grade metadata, and a changed pair set prevented active validation checks from passing.
- **Fix:** Split selectors by package/test, added grounded grade fields, reran the hosted recorder, and installed the produced 35-pair artifact.
- **Files modified:** `26-VALIDATION.md`, validation receipt JSONL and manifest.
- **Verification:** Focused 37101757588 and full 37101934669 passed.
- **Committed in:** `677a453`, `46a4dfd`, `72a8174`, `b308651`, `1dc5e63`.

**Total deviations:** 3 auto-fixed. These close source-path and validation blockers already within Plan 05 acceptance; no new backend or public result shape was added.

## Issues Encountered

The first full gate after receipt refresh identified the missing `Grade` and `Non-inertness` columns. They were added and validated in the final focused and full hosted runs. Local source formatting and compile-only build checks passed; local test execution remained prohibited by project policy.

## User Setup Required

None.

## Next Phase Readiness

Plan 05 and the Phase 26 hosted gate are complete. The phase remains in progress until independent review and goal-backward verification by the orchestrator. No Phase 27 behavior was introduced.

---
*Phase: 26-checked-scalar-sum*
*Completed: 2026-10-03*

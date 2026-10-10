---
phase: 26-checked-scalar-sum
plan: 02
subsystem: compiler
tags: [u64, overflow, interpreter, c17, native, ci]

requires:
  - phase: 26-checked-scalar-sum
    provides: checked scalar sum foundation and Phase 26 integration repairs
provides:
  - Checked U64 addition defects before result assignment in interpreter and generated C
  - Pinned app child-exit, stderr, tool-error, and evidence-capacity channels
  - Focused hosted CI lane covering implemented Phase 26 packages
affects: [26-checked-scalar-sum, language-maturity, ci]

actuals:
  tokens: 5999
  tasks: 2
  commits: 8
plan_head_before: 30c82faa8165cbf87a5e6c5f114a711e94a8b5d8
commits: 8

tech-stack:
  added: []
  patterns:
    - Check unsigned overflow before assigning an arithmetic result
    - Keep process outcome, tool error, and evidence capture status in separate channels

key-files:
  created:
    - examples/phase26/checked_add_overflow.schway
    - internal/compiler/session/session_phase26_overflow_test.go
  modified:
    - .github/workflows/ci.yml
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_test.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/native/native_app_test.go
    - .planning/LANGUAGE-MATURITY.md

key-decisions:
  - "Use the existing OutcomeDefect and process-outcome contracts; do not add a public outcome type."
  - "Verify execution on the hosted phase26_focused lane; do not run project suites locally."

patterns-established:
  - "Overflow checks precede U64 result assignment in both semantic engines."
  - "Native wrong-result mutation is compared against an independently authored expected value."

requirements-completed: [U64-01, APP-07]

coverage:
  - id: D1
    description: "Near-MAX U64 addition is exact in range and becomes an attributed noncatchable defect before wrapping."
    requirement: U64-01
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestPhase26CheckedAddInterpreter"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestPhase26CheckedAddC17"
        status: pass
    human_judgment: false
  - id: D2
    description: "Native application overflow has a deterministic child exit and diagnostic, distinct from tool and evidence failures."
    requirement: APP-07
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase26_overflow_test.go#TestPhase26OverflowProcessOutcome"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/native_app_test.go#TestPhase26FailureChannels"
        status: pass
      - kind: integration
        ref: "GitHub Actions phase26_focused run 37081675751, Ubuntu and macOS"
        status: pass
    human_judgment: false

duration: 93min
completed: 2026-10-02
status: complete
---

# Phase 26 Plan 02: Checked Scalar Sum Summary

**U64 addition now rejects reached overflow before producing a wrapped result, with matching interpreter and native app evidence.**

## Performance

- **Duration:** 93 min (estimated from the prior plan completion at 22:47Z to hosted verification at 00:20Z)
- **Started:** 2026-10-02T22:47:00Z (estimated)
- **Completed:** 2026-10-03T00:20:00Z
- **Tasks:** 2
- **Files modified:** 11 plan-delivery files, including the corpus census update

## Accomplishments

- Added a direct near-MAX arithmetic source witness and checked overflow in the interpreter and C17 emitter before result assignment.
- Pinned the interpreter's empty `OutcomeDefect` and attributed `function.defected` event, plus native exit 65, empty stdout, and exact bounded stderr `schway: U64 addition overflow\n`.
- Added native assertions separating child exit, ToolError, and evidence-capacity exhaustion; a reached `+ 1` generated-C mutant is rejected against an independently authored expected output.
- Expanded `phase26_focused` to include `interp`, `cgen`, and `native` alongside `session` and `core`; refreshed the language corpus count to 153 programs / 4,963 lines.

## Hosted Verification

The focused GitHub Actions run **37081675751** passed on **Ubuntu and macOS** at revision `1677ca3` (`test(26-02): kill wrong-result native mutation`). Task 1's valid RED run **37075505541** failed on both hosts at the intended interpreter and C-emitter assertions before the implementation commits. Its persisted evidence passed `gsd-tools check tdd-red-evidence`. The task 2 candidate assertions passed on first hosted execution; source inspection confirmed the process/tool/capture separation already existed, so the task added regression coverage rather than manufacturing a production behavior change. No project suites, native verification scripts, or test commands were run locally. `ci_monitor.cjs` was absent from both the repository and skill installation; monitoring and hosted logs used the `gh` CLI fallback.

The broader workflow run **37074516513** had failures documented in `26-WAVE-GATES.md`; remaining draft/undefined future-plan selectors are not covered by this focused receipt and remain for plans 03–05. No claim is made that the full Phase 26 suite passed.

## Task Commits

1. **Task 1: Reach a direct checked-add overflow in both semantic engines** — `4f4ecc1`, `c74d783`, `8f2a5be`, `66d1695` (RED witness/evidence, implementation, and call-site repair).
2. **Task 2: Pin child process overflow apart from tool and evidence failures** — `830a906`, `280ad94`, `1677ca3` (test contract, channel assertions, and wrong-result mutation control).
3. **Corpus census refresh** — `055f343`.

The eight commits measured from `plan_head_before` include the seven task commits and corpus census update. The summary/state/roadmap closeout commit is recorded separately after those artifacts are updated.

## Decisions Made

- Reused existing `OutcomeDefect`, `RunOutcome`, and capture status contracts rather than adding a public outcome type.
- Kept local verification to source inspection, formatting, and Git checks; hosted CI ran all focused Phase 26 tests on both supported hosts.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Corrected the scalar overflow witness and application shell call path.**
- **Found during:** Task 1
- **Issue:** The first candidate source witness was not a valid scalar statement, and the scalar fallback had a legacy emitter call missing the application shell argument.
- **Fix:** Reworked the witness into a scalar statement, updated its RED evidence, and passed the shell argument through the fallback.
- **Files modified:** `examples/phase26/checked_add_overflow.schway`, `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_program_test.go`, RED evidence records.
- **Verification:** Valid hosted RED run 37075505541 failed at the intended behavior assertions, then focused hosted GREEN run 37076057779 passed both hosts.
- **Committed in:** `c74d783`, `66d1695`.

**2. [Rule 2 - Missing critical test coverage] Added an explicit native wrong-result mutation control.**
- **Found during:** Task 2
- **Issue:** Initial Phase 26 channel assertions passed because the underlying channels were already implemented; the tests did not yet prove they would reject a native wrong result.
- **Fix:** Mutated emitted addition to add one and compared execution to the independent exact expected output.
- **Files modified:** `internal/compiler/native/native_app_test.go`.
- **Verification:** Focused hosted run 37081675751 passed both hosts with the mutation control.
- **Committed in:** `1677ca3`.

**Total deviations:** 2 auto-fixed. **Impact:** Both close execution-evidence gaps; no public API or plan scope expansion.

## Issues Encountered

- `ci_monitor.cjs` was unavailable, so `gh run watch` and `gh api` were used to inspect hosted runs/logs.
- The initial malformed witness was caught by hosted RED evidence and corrected before implementation; no invalid RED evidence was accepted.

## User Setup Required

None.

## Next Phase Readiness

The checked-add slice and focused cross-host lane are ready for plan 03. Broader workflow blockers and future test selectors remain as recorded in `26-WAVE-GATES.md`; this plan does not mark Phase 26 complete.

## TDD Gate Compliance

- Task 1: valid intentional RED established on Ubuntu and macOS; evidence files recorded and accepted by `check tdd-red-evidence`; implementation reached hosted GREEN.
- Task 2: assertions were unexpectedly green on first run because the distinction behavior pre-existed. Source inspection and existing related tests documented this, then an independent wrong-result mutation was added and the expanded focused lane passed.
- Local test/native execution was omitted per the inherited constraint; the hosted focused lane is the verification receipt.

## Self-Check: PASSED

The summary file exists, all eight plan-range commits are present, and `git diff --check` is clean. The pre-existing untracked `.planning/milestone.lock` and `.planning/research/.cache/` remain untouched.

---
*Phase: 26-checked-scalar-sum*
*Completed: 2026-10-03*

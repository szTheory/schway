---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "01"
subsystem: testing
tags: [Go, compiler-baseline, numeric-literals, parser, cgen]

requires:
  - phase: 34
    provides: inherited C-output and Bool-identifier compatibility baselines
provides:
  - Revision-pinned tracked-source, checked-core, semantic-ID, and three-emitter baseline corpus
  - Route-local numeric policy and Bool identity observations
  - Read-only nested byte_at admission history at depths 63, 64, 65, and 128
affects: [Phase 35 implementation plans]

actuals:
  tokens: 1432357
  tasks: 2
  commits: 2
commits: 2
plan_head_before: 1fd69337d980c4156914d47669a520fc8b04cc96
plan_head_after: 8fd061ed5197b99a7d7a2cedc34ca491b9e82559

tech-stack:
  added: []
  patterns:
    - Read-only revision-pinned compiler evidence fixtures
key-files:
  created:
    - internal/compiler/session/session_phase35_prechange_test.go
    - testdata/phase35/prechange-corpus.json
    - internal/compiler/check/check_phase35_baseline_test.go
    - internal/compiler/syntax/parser_phase35_baseline_test.go
    - testdata/phase35/numeric-policy-baseline.json
  modified: []
key-decisions:
  - "Keep legacy and scalar numeric conversion results route-local, including their existing disagreement and refusals."
  - "Treat nested-index results as immutable historical evidence; measure any future depth refusal separately."
patterns-established:
  - "Baseline fixtures pin their source revision and preserve source bytes, diagnostics, core identities, and available C output."
requirements-completed: [] # Baseline evidence only; BIT-01 through BIT-04 remain pending implementation.
coverage:
  - id: D1
    description: "Revision-pinned tracked-source corpus captures checking, core identity, and explicit outcomes from all three C emitter routes, with reached drift controls."
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeEmitterCorpus
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeCorpusControls
        status: pass
    human_judgment: false
  - id: D2
    description: "Route-local numeric and Bool behavior plus actual recursive byte_at admission are frozen with diagnostic and emitted-C evidence."
    verification:
      - kind: unit
        ref: internal/compiler/check/check_phase35_baseline_test.go#TestPhase35NumericPolicyBaseline
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_phase35_baseline_test.go#TestPhase35NestedIndexBaseline
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase34_compat_test.go#TestPhase34PreChangeCProgramGolden
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase34_compat_test.go#TestPhase34PreChangeCGoldenIsNotInert
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase34_compat_test.go#TestPhase34BoolNameBaseline
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase34_compat_test.go#TestPhase34IdentityBaseline
        status: pass
    human_judgment: false

duration: 24min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 01: Pre-change Compatibility Baselines Summary

**Tracked-source emitter goldens, route-specific numeric policy, and the actual pre-bound nested-index admission are frozen before Phase 35 compiler changes.**

## Performance

- **Duration:** 24 min
- **Started:** 2026-10-14T13:34:19Z
- **Completed:** 2026-10-14T13:58:51Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Pinned all tracked `.schway` inputs to revision `1fd69337` with literal source, core identity, and explicit `Emit`, `EmitNative`, and `EmitApplication` outcomes. The copied-byte, identity, outcome, source-row, and emitter-row controls are reached and restored.
- Recorded twelve scalar/legacy numeric cases and four legal Bool parameter/local identity cases. The live regression gate preserves legacy `010 = 10`, scalar `010 = 8`, scalar `0B10 = 2`, and the legacy `0B10` refusal.
- Captured actual `byte_at` nesting at depths 63, 64, 65, and 128 against revision `c4a2eec1`. Every case parsed and checked, and all three emitter routes produced C; the complete outputs and core identities remain in the read-only fixture.

## Task Commits

1. **Task 1: Commit one source-to-C baseline path, then its complete frozen inventory** - `c4a2eec1` (`test(35-01): freeze pre-change emitter baseline`)
2. **Task 2: Freeze route-local numeric policy and characterize recursive admission** - `8fd061ed` (`test(35-01): freeze numeric and index baselines`)

## Files Created/Modified

- `internal/compiler/session/session_phase35_prechange_test.go` - compares the pinned source/core/emitter observations and reaches copied-row drift controls.
- `testdata/phase35/prechange-corpus.json` - literal source inventory, core/identity witnesses, and all three emitter outcomes.
- `internal/compiler/check/check_phase35_baseline_test.go` - validates historical numeric policy and live stable numeric/Bool behavior.
- `internal/compiler/syntax/parser_phase35_baseline_test.go` - validates the historical nested-index fixture without rerunning it against future parser behavior.
- `testdata/phase35/numeric-policy-baseline.json` - route-local literal/Bool observations and complete before-state C output for nested indexes.

## Decisions Made

- Kept the old legacy and scalar conversion results separate: legacy `010` is decimal 10 and refuses `0B10`/`0X1A`; scalar `010` is 8 and admits uppercase `0B10`/`0X1A` as 2/26. Lowercase binary/hex and separators are also recorded per route.
- Preserved the observed nested-index successes as historical data. A later depth limit must have a separate after-state measurement and must not rewrite these rows.
- Left BIT-01 through BIT-04 pending because this plan creates evidence and no feature implementation.

## Deviations from Plan

None - plan executed as written. No production compiler files changed.

## Issues Encountered

The first nested-index probe used a tail expression outside the scalar block shape and reached the existing generic-block parser refusal. The fixture was corrected to bind `byte_at(...)` in a scalar `let` and return the binding; the captured source then parsed, checked, and emitted at all four measured depths.

## User Setup Required

None - no external service configuration is required.

## Next Phase Readiness

Both baseline commits precede all Phase 35 implementation work. The next plans can compare changes against the frozen source and route policies. The depth history establishes that 63, 64, 65, and 128 calls were all admitted before a new bound; any new refusal needs its own after-state evidence. BIT-01 through BIT-04 remain pending.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*

## Self-Check: PASSED

All five plan output files exist, and both task commits are ancestors of the current plan branch.

---
phase: 19-numeric-literals-and-opconst
plan: 05
subsystem: core-validation
tags: [U64, OpConst, corevalidate, pathoracle, originvalidate]
requires:
  - phase: 19-04
    provides: typed canonical U64 constants in checked core
provides:
  - Independent U64 ability and OpConst admission with forged-core refusals
  - Source-free constant roots in path and origin analysis
  - Literal admission in match entry prefixes and arm bodies
affects: [phase-19-interpreter, phase-19-native-emission, phase-19-dispatch]
actuals:
  tokens: 5502
  tasks: 2
  commits: 3
commits: 3
plan_head_before: 719074af745269a67fc67167eda2a9a783d99c41
tech-stack:
  added: []
  patterns: [independent U64 re-derivation, source-free constant provenance roots]
key-files:
  created: [internal/compiler/corevalidate/corevalidate_phase19_test.go]
  modified: [internal/compiler/corevalidate/corevalidate.go, internal/compiler/check/check.go, internal/compiler/check/check_test.go, internal/compiler/pathoracle/pathoracle.go, internal/compiler/pathoracle/pathoracle_test.go, internal/compiler/originvalidate/originvalidate.go, internal/compiler/originvalidate/originvalidate_test.go]
key-decisions:
  - "Core admission validates canonical decimal U64 values independently and requires kind-exclusive fields."
  - "A constant root terminates origin traversal and carries no path-oracle loan chain."
  - "Branch U64 facts retain the stable type:u64 identity; peer type ordering recognizes only the zero-argument U64 fact with that identity."
requirements-completed: [VAL-01, VAL-02]
coverage:
  - id: D1
    description: "Corevalidate independently admits typed canonical constants and rejects forged facts."
    requirement: VAL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/corevalidate -run 'TestPhase19(OpConst|U64|Forged)' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Path and origin peers treat constants as source-free roots without inherited loans or parameter origins."
    requirement: VAL-02
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/pathoracle ./internal/compiler/originvalidate -run 'TestPhase19(OpConst|ConstantOrigin|ConstantPath)' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: "U64 literals are available in match entry prefixes and arm bodies."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check -count=1"
        status: pass
    human_judgment: false
duration: 13min
completed: 2026-09-25
status: complete
---

# Phase 19 Plan 05: Independent `OpConst` Peers Summary

**Corevalidation now independently admits canonical U64 constants, while path and origin analysis stop at each source-free constant root.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-09-25T00:16:56Z
- **Completed:** 2026-09-25T00:29:12Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Re-derived U64 abilities and admitted only canonical decimal `OpConst` values with valid U64 targets and no forged source, loan, call, or payload facts.
- Added constant-root handling to path and origin peers, including a mutation proving a supplied source cannot inject loan inheritance or parameter provenance.
- Covered checked constants in a Phase 18 entry prefix and arm body.

## Task Commits

1. **Task 1: Re-derive constant and U64 invariants in corevalidate** — `1b23bf4` (RED controls), `289c906` (implementation).
2. **Task 2: Recompute path and origin facts for constant roots** — `32f6e16` (implementation and peer controls).

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` and `corevalidate_phase19_test.go` — independent U64 and OpConst admission and forged-core controls.
- `internal/compiler/check/check.go` and `check_test.go` — seed U64 type facts across match prefixes/arms and validate arm results against the declared return type.
- `internal/compiler/pathoracle/pathoracle.go` and `pathoracle_test.go` — make OpConst a source-free path root and test loan-inheritance mutation.
- `internal/compiler/originvalidate/originvalidate.go` and `originvalidate_test.go` — terminate provenance walks at constants and test a forged borrow source.

## Decisions Made

- Corevalidation parses the canonical decimal representation itself using bounded unsigned parsing and round-trip formatting.
- Constants terminate provenance walks and carry an empty path-oracle loan chain.
- Supplemental U64 type facts preserve the `type:u64` identity; corevalidate admits that identity only for a zero-argument U64 fact.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Seed and validate U64 facts in branch bodies**
- **Found during:** Task 2 (Recompute path and origin facts for constant roots)
- **Issue:** Match entry prefixes and arm bodies did not provide a U64 type fact, so legal Phase 18 placements were rejected before reaching the peers. Arm result checking also treated a supplemental U64 fact at ordinal one as the declared return type when parameter and return types were identical.
- **Fix:** Seed the U64 fact from prefix and arm bodies before analysis, preserve its stable `type:u64` identity, teach corevalidate to recognize that supplemental fact, and compare arm results directly to the declared return type.
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_test.go`, `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** Focused peer commands and complete `corevalidate`, `pathoracle`, `originvalidate`, and `check` package suites passed.
- **Committed in:** `32f6e16`

**Total deviations:** 1 auto-fixed (Rule 2). **Impact:** Enables the branch placements explicitly required by this plan and prevents U64 from being mistaken for a function return type.

## Issues Encountered

- The first full `corevalidate` run caught exact work-budget regressions from per-operation validation checks. The kind-specific facts were consolidated into existing counted checks; the complete package suite then passed without changing the established linear work formula.
- The repository build cache required `GOCACHE=/private/tmp/ai-lang-go-cache` in this environment.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 19-06 can lower the independently admitted U64 constants to exact-width native C; plan 19-07 can then register and exercise all six operation consumers.

---
*Phase: 19-numeric-literals-and-opconst*
*Completed: 2026-09-25*

## Self-Check: PASSED

- All created and modified source/test files exist.
- Measured task commits: 3 from `719074af745269a67fc67167eda2a9a783d99c41` to `32f6e16`.
- Focused peer checks, full affected package suites, and `git diff --check` passed.

---
phase: 19-numeric-literals-and-opconst
plan: 04
subsystem: checker-core
tags: [U64, numeric-literals, OpConst, type-checking]
requires:
  - phase: 19-03
    provides: lossless numeric RHS syntax and source spans
provides:
  - Structural U64 ability derivation and kind-distinct OpConst core shape
  - Checked radix conversion, U64 range refusal, and direct literal lowering
affects: [phase-19-core-validation, phase-19-interpreter, phase-19-native-emission]
actuals:
  tokens: 4473.25
  tasks: 2
  commits: 4
tech-stack:
  added: []
  patterns: [checked ParseUint radix conversion, canonical decimal core values]
key-files:
  created: [internal/compiler/check/check_phase19_test.go]
  modified: [internal/compiler/ability/ability.go, internal/compiler/ability/ability_test.go, internal/compiler/core/core.go, internal/compiler/core/core_test.go, internal/compiler/check/check.go, internal/compiler/session/session_phase19_test.go]
key-decisions:
  - "U64 is a zero-argument scalar type granting copy, drop, share, send, and escape."
  - "OpConst stores canonical decimal U64 digits in an omitempty field; source spelling remains in syntax/AST."
  - "Literal range is checked before core admission, and literals do not contextually convert to declared return types."
patterns-established:
  - "Numeric literal lowering: parse radix digits with strconv.ParseUint at 64 bits, then store strconv.FormatUint decimal output."
requirements-completed: []
coverage:
  - id: D1
    description: U64 ability and OpConst preserve canonical semantic identity and legacy JSON.
    requirement: VAL-01
    verification:
      - kind: unit
        ref: go test ./internal/compiler/ability ./internal/compiler/core -run 'TestPhase19(U64Ability|OpConstShape)' -count=1
        status: pass
    human_judgment: false
  - id: D2
    description: Valid direct literals lower to typed OpConst facts and overflow/type mismatches are refused.
    requirement: VAL-01
    verification:
      - kind: unit
        ref: go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase19(LiteralAdmission|LiteralRange|LiteralType|LiteralFrontier|NumericRefusalFrontiers)' -count=1
        status: pass
    human_judgment: false
metrics:
  duration: 10min
  completed: 2026-09-25
  commits: 4
  plan_head_before: f646b1a2655a789f61031dbbc3d435bb37c1f1f5
status: complete
---

# Phase 19 Plan 04: U64 Literal Admission Summary

**Direct numeric bindings now become fresh typed `OpConst` places carrying canonical U64 values, with overflow and return-type mismatches rejected before executable core.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-09-25T00:03:00Z
- **Completed:** 2026-09-25T00:13:27Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Added the zero-argument U64 scalar to structural ability derivation and defined `OpConst` with a kind-exclusive canonical decimal payload.
- Lowered decimal, hexadecimal, binary, zero, and maximum-U64 literal bindings to fresh typed places; rejected max+1 at the complete literal span.
- Moved the Phase 19 valid-source frontier from checker refusal to a typed constant operation while preserving malformed-literal syntax diagnostics.

## Task Commits

1. **Task 1: Define the U64 type fact and kind-exclusive constant representation** — `88de2d4` (RED tests), `e03ed84` (GREEN implementation).
2. **Task 2: Check range and lower direct literal bindings** — `a88a038` (RED tests), `e051fd0` (GREEN implementation).

**Plan metadata:** pending final metadata commit.

## Files Created/Modified

- `internal/compiler/ability/ability.go` and `ability_test.go` — U64 scalar abilities and zero-argument constraint.
- `internal/compiler/core/core.go` and `core_test.go` — kind-distinct `OpConst`, canonical `ConstU64`, and unchanged legacy JSON when unused.
- `internal/compiler/check/check.go` and `check_phase19_test.go` — checked conversion, typed lowering, range/type tests.
- `internal/compiler/session/session_phase19_test.go` — source frontier and refusal expectation updates.

## Decisions Made

- `ConstU64` is canonical base-10 digits, including the non-empty value `"0"`; `omitempty` preserves old operation JSON when the field is unused.
- U64 function boundaries use exact type identity, with no literal inference or conversion.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Lower literals in the separate match-arm body analyzer**
- **Found during:** Task 2 (Check range and lower direct literal bindings)
- **Issue:** Match-arm linear bodies use a distinct checker path; without matching literal handling, the already-parsed numeric RHS would be treated as a place name there.
- **Fix:** Added the same checked U64 conversion and fresh `OpConst` lowering to that checker path.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** Focused checker and session Phase 19 tests passed.
- **Committed in:** `e051fd0`

**Total deviations:** 1 auto-fixed (Rule 2). **Impact:** Applies the same checker contract to both existing linear-body analysis paths.

## Issues Encountered

- The default Go build cache was inaccessible in the sandbox. Focused tests passed with `GOCACHE=/private/tmp/ai-lang-go-cache`.
- The broader `go test ./internal/compiler/session -count=1` failure recorded by Plan 19-03 remains a pre-existing phase regression: stale LANGUAGE-MATURITY corpus counts (137/4572 vs 140/4602) and unreconciled findings from Phase 19 validation command text. It is outside these source changes and remains unresolved.
- `roadmap.update-plan-progress 19` returned `missing_phase_details`; `ROADMAP.md` has no writable Phase 19 progress entry. It was left unchanged as instructed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The independent core validator and runtime/backend plans can now consume the typed constant contract. `VAL-02` remains pending until all six dispatch sites and the exhaustive controls are complete; no requirement was newly marked complete by this plan.

---
*Phase: 19-numeric-literals-and-opconst*
*Completed: 2026-09-25*

## Self-Check: PASSED

- All seven planned source/test files exist and are included in the four task commits.
- The task commit range measures four commits from `plan_head_before` to `HEAD`.
- The focused ability/core/checker/session verification commands passed.

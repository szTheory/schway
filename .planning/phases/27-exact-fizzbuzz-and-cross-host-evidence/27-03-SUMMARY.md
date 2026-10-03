---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 03
subsystem: compiler
tags: [parser, checker, u64, remainder, scalar-analysis]

# Dependency graph
requires:
  - phase: 27-02
    provides: Runnable Unit application path and exact-output witness for the next arithmetic slice.
provides:
  - Canonical U64 equality and remainder source operators with typed core operations.
  - Source-span diagnostics for literal and scalar-propagated zero remainder divisors.
  - Admission of unknown divisors for the guarded engine implementation in Plan 27-04.
affects: [27-04, U64-02, phase27-evidence]

# Actuals (#2632)
actuals:
  tokens: 4934
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Equality and remainder use the existing binary-expression AST and scalar fixed-point domain.
    - Known-zero diagnostics retain the divisor expression's source span through operation IDs.

key-files:
  created:
    - internal/compiler/session/session_phase27_remainder_test.go
  modified:
    - internal/compiler/syntax/token.go
    - internal/compiler/syntax/lexer.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/core/core.go
    - internal/compiler/check/scalar.go

key-decisions:
  - "Use == for U64 equality and % for U64 remainder, preserving the existing binary AST shape and canonical formatter spacing."
  - "Reject literal and propagated zero divisors during checking; leave unknown divisors for the runtime guard implemented by Plan 27-04."
  - "Keep U64-02's spec-less edge probe unclassified; this summary records only the plan's locked typed and static-zero acceptance."

patterns-established:
  - "Scalar operation diagnostics can map a fixed-point fact back to the original operand span using the operation ID."

requirements-completed: [U64-02]

coverage:
  - id: D1
    description: "Ordinary source parses, formats, and types U64 equality and remainder with the expected Bool and U64 results; invalid operand types are rejected and recovery preserves source spans."
    requirement: U64-02
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax ./internal/compiler/session -run '^TestPhase27(OperatorSyntax|StaticZero)$|^TestPhase26FrontendRecovery$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Literal and propagated zero remainder divisors are refused at their source spans; unknown divisors are admitted for the runtime guard, while Phase 26 fixed-point and back-edge controls continue to pass."
    requirement: U64-02
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check ./internal/compiler/session -run '^TestPhase27(ScalarZero|StaticZero)$|^TestPhase26(FixedPoint|AnalysisExhaustion|BackEdgeAuthority|ScalarCopyCycleBoundary)$' -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 8 min
completed: 2026-10-03
status: complete
---

# Phase 27 Plan 03: U64 Equality and Remainder Admission Summary

**U64 equality and remainder now have canonical typed source forms, with source-spanned refusal for every statically known zero divisor.**

## Performance

- **Duration:** 8 min
- **Started:** 2026-10-03T14:55:00-04:00
- **Completed:** 2026-10-03T15:03:00-04:00
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Added `==` and `%` lexing, precedence, formatting, typed core operations, and source-level type checks.
- Rejected literal and scalar-propagated zero divisors at their source spans; admitted an unknown divisor for the following runtime-guard plan.
- Kept the Phase 26 scalar fixed-point, analysis-exhaustion, and unsupported back-edge controls passing.

## Task Commits

1. **Task 1: Parse, format and type U64 equality and remainder** - `c0ca874` (feat)
2. **Task 2: Preserve scalar zero facts and loop authority** - `17fd082` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `internal/compiler/syntax/token.go` - Equality and remainder tokens.
- `internal/compiler/syntax/lexer.go` - `==` and `%` token recognition.
- `internal/compiler/syntax/parser.go` - Binary operator precedence and AST mapping.
- `internal/compiler/syntax/format.go` - Canonical spacing for both operators.
- `internal/compiler/core/core.go` - Typed U64 equality and remainder operations.
- `internal/compiler/check/scalar.go` - Operator transfer and propagated-zero diagnostics.
- `internal/compiler/session/session_phase27_remainder_test.go` - Syntax, typing, zero-divisor, and unknown-divisor controls.

## Decisions Made

- Reused the binary-expression AST and scalar fixed-point domain, keeping the feature within the existing checker architecture.
- Preserved an unknown divisor as an admitted operation for Plan 27-04 to guard at runtime; no engine semantics are claimed by this plan.
- Kept the U64-02 spec-less edge probe explicitly unclassified.

## Deviations from Plan

None - plan executed as written.

## Issues Encountered

- The unknown-divisor test uses a guarded control-flow shape because the independent validator refuses a terminal CFG with no incoming edge. This keeps the checker admission case within the existing CFG contract.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 27-04 can implement independent core validation and guarded interpreter/native remainder evaluation.
- Focused Plan 27-03 selectors and Phase 26 scalar controls passed. No full project suite was run.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*

---
phase: 19-numeric-literals-and-opconst
plan: 03
subsystem: syntax
tags: [lexer, numeric-literals, parser, formatter]
requires:
  - phase: 19-02
    provides: widened scalar projection and the literal frontier fixture
provides:
  - Bounded full-candidate numeric tokenization for decimal, hexadecimal, and binary spellings
  - Distinct literal RHS nodes retaining exact spelling and byte span
  - Idempotent formatting that preserves numeric spelling
affects: [phase-19-checker-admission, numeric-literal-syntax]
actuals:
  tokens: 3032.5
  tasks: 2
  commits: 5
tech-stack:
  added: []
  patterns: [complete-candidate numeric lexing, source-preserving literal RHS]
key-files:
  created: [internal/compiler/syntax/parser_phase19_test.go]
  modified: [internal/compiler/syntax/token.go, internal/compiler/syntax/lexer.go, internal/compiler/syntax/syntax_test.go, internal/compiler/syntax/parser.go, internal/compiler/syntax/format.go, internal/compiler/ast/ast.go, internal/compiler/session/session_phase19_test.go]
key-decisions:
  - "Retain numeric spelling in RHS.Source and defer numeric meaning to checker admission."
  - "Consume Unicode letter and digit suffix candidates as part of the rejected token."
requirements-completed: [VAL-01]
commits: 5
plan_head_before: 0a436e9f647958bf2bb13b744d4921e0a437b6f4
duration: 9min
completed: 2026-09-25
status: complete
---

# Phase 19 Plan 03: Numeric Literal Syntax Summary

Decimal, hexadecimal, and binary integer spellings now survive lexing, parsing, and canonical formatting without losing their original text or byte spans; malformed candidates are refused as whole tokens.

## Performance

- **Duration:** 9 min
- **Started:** 2026-09-24T23:53:10Z
- **Completed:** 2026-09-25T00:01:49Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Added a bounded forward scan that validates decimal, `0x`, and `0b` candidates and enforces radix-valid digits with separators only between digits.
- Added literal RHS parsing at the direct `let` operand, retaining exact spelling and source span; canonical formatting remains idempotent.
- Advanced the Phase 19 tracer from a syntax refusal to the checker’s `type.unknown` refusal, keeping the literal outside executable core.

## Task Commits

1. **Task 1: Tokenize complete numeric spellings and reject malformed tails** - `d8ea2ea` (RED tests), `7ca8e6f` (implementation), `3d43d67` (Unicode suffix boundary fix)
2. **Task 2: Parse literal bindings and preserve spellings under canonical format** - `7b82507` (RED tests), `2b72d15` (implementation)

**Plan metadata:** `99d73ad` (docs: complete plan).

## Files Created/Modified

- `internal/compiler/syntax/token.go` — added the numeric token kind.
- `internal/compiler/syntax/lexer.go` — validates complete numeric candidates and reports full-span malformed-token diagnostics.
- `internal/compiler/syntax/syntax_test.go` — covers accepted spellings and malformed complete candidates.
- `internal/compiler/syntax/parser.go` and `internal/compiler/ast/ast.go` — added distinct literal RHS representation with source spelling/span.
- `internal/compiler/syntax/format.go` — emits the original numeric text and ends literal binding lines consistently.
- `internal/compiler/syntax/parser_phase19_test.go` — covers literal RHS parsing and formatting across radices and separators.
- `internal/compiler/session/session_phase19_test.go` — pins the moved tracer/checker frontier and malformed-token span.

## Decisions Made

- Numeric syntax is preserved as source text; range interpretation remains the checker’s responsibility.
- Numeric candidate scanning includes Unicode letters and digits so an adjacent suffix cannot be split off and accepted independently.

## Deviations from Plan

**1. [Rule 2 - Input-boundary correctness] Consume Unicode suffix candidates in the same malformed token**
- **Found during:** Task 1
- **Issue:** An ASCII-only candidate scan would leave a Unicode letter after a numeric spelling as a separate token, weakening complete-token refusal.
- **Fix:** Extended the bounded candidate scan to consume Unicode letters and digits and added a `42λ` negative control.
- **Files modified:** `internal/compiler/syntax/lexer.go`, `internal/compiler/syntax/syntax_test.go`
- **Verification:** `TestPhase19NumericMalformed`
- **Committed in:** `3d43d67`

**Total deviations:** 1 auto-fixed (Rule 2). **Impact:** Tightens the planned strict token boundary.

## Issues Encountered

- The focused plan checks passed. A broader `go test ./internal/compiler/session -count=1` ran but failed on existing repository-wide doc and groundedness pins: stale LANGUAGE-MATURITY corpus counts (137/4572 vs 140/4602) and unreconciled findings from Phase 19 validation command text. These failures are outside this plan’s source changes.
- The sandbox required an alternate Go build cache at `/private/tmp/ai-lang-go-cache`; the default user cache was not writable.

## Next Phase Readiness

Plan 19-04 can add checker admission and range diagnostics using the literal RHS kind and source span. The syntax/checker boundary is now explicit.

---
*Phase: 19-numeric-literals-and-opconst*
*Completed: 2026-09-25*

## Self-Check: PASSED

- SUMMARY file exists.
- All five task commits are present in Git history.
- `git diff --check` passed for the plan commits.

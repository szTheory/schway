---
phase: 36-scalar-helper-calls
plan: "02"
subsystem: compiler
tags: [parser, ast, scalar-checker, OpCall]

# Dependency graph
requires:
  - phase: 36-scalar-helper-calls/36-01
    provides: frozen no-call C corpus and existing scalar/module contracts
provides:
  - Bounded arity-one named scalar-call AST, parser recovery, and formatter round trips
  - Ordered typed OpCall lowering for all U64/Bool argument/result pairs
  - Import, recursion, signature, and source-intent controls for the checker path
affects: [36-scalar-helper-calls/36-03, 36-scalar-helper-calls/36-04, 36-scalar-helper-calls/36-05, 37-crc32-consumer]

# Actuals measured from the plan-start ledger through the last task commit.
actuals:
  tokens: 13852
  tasks: 3
  commits: 8
commits: 8
plan_head_before: 85238933a99a45708c5d5eda5712fc9cffab0777
plan_head_after: 7a6d870246cc9b851017539d23c105407000e1c7

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Expr.Kind call stores the direct callee in Name and its sole argument in Left.
    - Scalar calls evaluate arguments before emitting a fresh typed OpCall result.

key-files:
  created:
    - internal/compiler/syntax/parser_phase36_test.go
    - internal/compiler/check/check_phase36_scalar_calls_test.go
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/check/check.go
    - internal/compiler/check/scalar.go
    - internal/compiler/check/check_phase34_refusal_test.go

key-decisions:
  - "Reuse existing Expr fields and core.OpCall; add no AST field or core schema field."
  - "Treat call results as initialized, value-unknown locals with type from the declared callee result."

patterns-established:
  - "Nested and sibling call arguments are lowered left-to-right into explicit places."
  - "Source-intent controls use clean, no-op, mutant, and restored Go source overlays."

# CALL-01 and CALL-02 remain open until the remaining Phase 36 runtime and
# independent-peer obligations are complete.
requirements-completed: []
coverage:
  - id: D1
    description: "Direct bounded scalar calls parse, recover, and round-trip through canonical formatting."
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/parser_phase36_test.go#TestPhase36ScalarCallGrammar"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/parser_phase36_test.go#TestPhase36ScalarCallDepthAndRecovery"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/parser_phase36_test.go#TestPhase36ScalarCallFormatting"
        status: pass
    human_judgment: false
  - id: D2
    description: "All four U64/Bool signatures lower to ordered typed calls with fresh result places."
    requirement: CALL-01
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase36_scalar_calls_test.go#TestPhase36ScalarCallLoweringAndTypePairs"
        status: pass
    human_judgment: false
  - id: D3
    description: "Imported calls, refusal boundaries, recursion, and checker source-intent controls remain closed."
    requirement: CALL-02
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_phase36_scalar_calls_test.go#TestPhase36CallImportAndRefusalBoundaries"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_phase36_scalar_calls_test.go#TestPhase36CallSourceIntentControls"
        status: pass
    human_judgment: false

# Metrics
duration: 23min
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 02: Scalar Helper Calls Summary

**Bounded one-argument scalar calls now lower U64/Bool helper pairs into ordered fresh `OpCall` results, with import and source-intent refusal controls.**

## Performance

- **Duration:** 23 min
- **Started:** 2026-10-15T01:32:12Z
- **Completed:** 2026-10-15T01:55:06Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Added `Expr.Kind == "call"` using the existing `Name` and `Left` fields, with exactly one bounded nested argument, delimiter-aware recovery, and formatter round-trip coverage.
- Lowered U64→U64, U64→Bool, Bool→U64, and Bool→Bool calls into the existing typed `OpCall`; verified left-to-right nested and sibling work, assignments, conditions, byte indexes, and unknown dynamic results.
- Proved imported stable callee identity and retained private-name, collision, forbidden-signature, wrong-type, unknown-argument, immutable-assignment, and whole-program recursion refusals.

## TDD Evidence

- **Task 1 RED/GREEN:** The grammar, depth/recovery, and formatter selectors first failed on the absent call atom, then passed after the parser and AST update. A second focused failure found builtin initializer routing had moved behind named-call lookahead; its regression test and fix are included in the Task 1 commits.
- **Task 2 RED/GREEN:** `TestPhase36ScalarCallLoweringAndTypePairs` failed while the checker rejected Bool signatures and call atoms, then passed after contract-based `OpCall` lowering and value-unknown call-result transfer were added.
- **Task 3 mutation controls:** With a Go overlay selecting the caller parameter instead of the computed argument, `TestPhase36CallSourceIntentChild` failed on `SOURCE_INTENT_ARGUMENT_MISMATCH`; a separate return-type overlay failed on `SOURCE_INTENT_RESULT_MISMATCH`. The clean, no-op-overlay, and restored-overlay runs passed. The RED report was accepted as `RED_EVIDENCE_OK`.

## Task Commits

1. **Task 1: Parse, format and recover one nested named-call expression** — `385eb9b8` (test), `fe174ce3` (feat), `3977fa0b` (test), `a55e09d0` (fix).
2. **Task 2: Lower all four signatures, exact argument work and fresh result places** — `9aa79286` (test), `1ba2af1f` (feat).
3. **Task 3: Reach source-intent mutants and imported/cyclic refusal controls** — `2f6d61bd`, `7a6d8702` (test).

Plan metadata is committed separately after the canonical state update.

## Files Created/Modified

- `internal/compiler/ast/ast.go` — documents the call expression shape without adding AST fields.
- `internal/compiler/syntax/parser.go` — parses bounded named calls and preserves builtin initializer routing.
- `internal/compiler/syntax/parser_phase36_test.go` — grammar, arity, depth, recovery, and formatting controls.
- `internal/compiler/check/check.go` and `internal/compiler/check/scalar.go` — pass resolved callee contracts and lower typed calls into existing core.
- `internal/compiler/check/check_phase36_scalar_calls_test.go` — call-pair, source-order, import, refusal, recursion, and source-intent controls.
- `internal/compiler/check/check_phase34_refusal_test.go` — updates the now-superseded one-argument rejection expectation while retaining the two-argument refusal.

## Decisions Made

- Reused the existing AST and core call contracts; no AST field, IR kind, schema, or dependency was added.
- Kept call results initialized but value-unknown so local constant and zero-divisor analysis does not assume a helper's return value.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Preserve builtin byte-call initializer routing**
- **Found during:** Task 1
- **Issue:** Generic named-call lookahead intercepted `byte_at(view, index_of(1))` in a binding initializer before the builtin parser route.
- **Fix:** Keep `byte_len` and `byte_at` precedence ahead of generic named-call routing and pin the nested byte-index initializer.
- **Files modified:** `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/parser_phase36_test.go`
- **Verification:** `TestPhase36ByteAtIndexCallLetInitializer` and the full Task 1 focused syntax selector set passed.
- **Committed in:** `a55e09d0`

**2. [Rule 1 - Bug] Update the superseded Phase 34 helper-call expectation**
- **Found during:** Task 3
- **Issue:** `TestPhase34CallArityBoundaries` still expected a one-argument helper call in a scalar expression to be rejected, contradicting the Phase 36 source contract.
- **Fix:** Assert that the one-argument expression emits one `OpCall`; the existing two-argument refusal remains asserted.
- **Files modified:** `internal/compiler/check/check_phase34_refusal_test.go`
- **Verification:** The exact Task 3 selector set, including `TestPhase34CallArityBoundaries`, passed.
- **Committed in:** `2f6d61bd`

**Total deviations:** 2 auto-fixed (2 Rule 1).
**Impact on plan:** Both changes preserve required call routing and keep existing arity tests aligned with the newly admitted syntax.

## Issues Encountered

The Phase 34 one-argument expectation was stale after this plan introduced scalar call atoms; the focused test now checks the new behavior while continuing to reject arity two.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

This plan establishes the bounded text-to-checker path. Remaining Phase 36 plans own independent peer checks, interpreter/native behavior, repeated-call evidence, and the runnable CRC helper. CALL-01 and CALL-02 remain open until those phase obligations pass.

---
*Phase: 36-scalar-helper-calls*
*Completed: 2026-10-15*

## Self-Check: PASSED

The summary file exists, and all eight task commits are ancestors of the current plan head.

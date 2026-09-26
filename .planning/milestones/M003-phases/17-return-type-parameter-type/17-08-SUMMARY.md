---
phase: 17-return-type-parameter-type
plan: 08
subsystem: checker-repair-protocol
tags: [checker, diagnostics, repair, type-facts]
requires:
  - phase: 17-02
    provides: directional caller type facts
  - phase: 17-07
    provides: sealed derivation repair corpus
provides:
  - Source-reachable use_matching_argument repair emission
  - Exact argument-token repair spans and fail-closed candidate selection
affects: [17-09 repair-driver-evidence]
actuals:
  tokens: 5035
  tasks: 2
  commits: 4
plan_head_before: 308506a91c463a2afd44a43148d473744d6cc15a
tech-stack:
  added: []
  patterns: [per-place fact candidate partition, token-span source edit]
key-files:
  created: []
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_repair_emission_test.go
    - internal/compiler/check/check_test.go
key-decisions:
  - "Candidate selection resolves every place's own TypeID against the caller fact set."
  - "Machine edits use parser-preserved argument token spans rather than reconstructed offsets."
requirements-completed: [TYP-05]
coverage:
  - id: D1
    description: "Only one initialized, distinct place with the callee parameter type yields use_matching_argument."
    requirement: TYP-05
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check -run TestPhase17UseMatchingArgument(UniqueRealCandidate|ZeroCandidates|AmbiguousCandidates|UninitializedCandidate|NoOpCandidate) -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: "The repair is a machine-applicable exact-token source edit whose protocol fields are mutation-killed."
    requirement: TYP-05
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check -run TestPhase17UseMatchingArgument(ProtocolFields|FieldControlsAreNotInert|HistoricalNoOpStaysWithdrawn) -count=1 -v"
        status: pass
    human_judgment: false
duration: 15min
completed: 2026-09-22
status: complete
---

# Phase 17 Plan 08: Real Matching-Argument Repair Summary

**The checker now offers a single machine-applicable argument replacement only when a distinct initialized caller place uniquely satisfies the callee's parameter contract.**

## Accomplishments

- Selected candidates from their individual directional type facts, rejecting zero, ambiguous, uninitialized, and byte-identical alternatives.
- Preserved exact parser argument spans so the repair replaces only the incorrect argument token.
- Proved the derivation fixture's repair protocol and source splice, while retaining Phase 13's injected no-op control with no advertised repair.

## Task Commits

1. **Task 1: Reopen only the unique real candidate partition**
   - `f69a68e` — failing candidate-partition tests.
   - `3604830` — unique, source-reachable repair emission.
2. **Task 2: Bind exact protocol fields to the source call argument**
   - `2032dfd` — protocol, splice, mutation, and historical-control tests.

## Verification

- `go test ./internal/compiler/check -run 'TestPhase17UseMatchingArgument(UniqueRealCandidate|ZeroCandidates|AmbiguousCandidates|UninitializedCandidate|NoOpCandidate)' -count=1 -v` — passed.
- `go test ./internal/compiler/check -run 'TestPhase17UseMatchingArgument(ProtocolFields|FieldControlsAreNotInert|HistoricalNoOpStaysWithdrawn)' -count=1 -v` — passed.
- `go test ./internal/compiler/check -count=1` — passed.

## Decisions Made

- Candidate uniqueness is evaluated against `contract.ParameterType`, using each in-scope place's resolved fact rather than a function-wide parameter fact.
- The AST now retains argument spans because a token-exact source edit cannot safely be reconstructed from the callee and statement spans.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - correctness] Preserved argument token spans in the AST.**
- **Found during:** Task 1.
- **Issue:** The existing AST retained only the callee and whole binding spans, so no checker repair could prove an exact argument-only splice.
- **Fix:** Added `RHS.ArgumentSpans` and populated it in the parser for all call forms.
- **Files modified:** `internal/compiler/ast/ast.go`, `internal/compiler/syntax/parser.go`.
- **Verification:** Full checker and syntax package tests passed before the Task 1 commit.
- **Committed in:** `3604830`.

**Total deviations:** 1 auto-fixed (Rule 2).

## Known Stubs

None.

## Self-Check: PASSED

- All five production/test files listed above exist.
- Task commits `f69a68e`, `3604830`, and `2032dfd` exist in git history.

## Next Phase Readiness

Plan 09 can exercise the emitted repair through the sealed corpus and driver protocol without production references to held-out bytes or paths.

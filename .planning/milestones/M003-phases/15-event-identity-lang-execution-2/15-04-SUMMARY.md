---
phase: 15-event-identity-lang-execution-2
plan: "04"
subsystem: native-emission
tags: [cgen, callgraph, invocation-identity, bounded-preflight]
requires:
  - phase: 15-event-identity-lang-execution-2
    provides: lang.execution/2 admission contract
provides:
  - Bounded deterministic native invocation-node preflight
  - Stable 4096-node overflow diagnostic and mutation proof
affects: [native-emission, invocation-path-table]
actuals:
  tokens: 3500
  tasks: 2
  commits: 0
tech-stack:
  added: []
  patterns: [bounded preflight before serialization, test-only non-inertness seam]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/export_test.go
key-decisions:
  - "Count static activation occurrences independently of pathoracle and refuse node 4097 before C serialization."
  - "Carry deterministic parent and incoming-call witnesses in the overflow diagnostic."
patterns-established:
  - "Native amplification guards run after callgraph and entry validation, before generated C construction."
requirements-completed: [OBS-04]
coverage:
  - id: D1
    description: "The checked four-diamond source fixture unfolds to exactly 61 invocation nodes after graph and entry validation."
    requirement: OBS-04
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestInvocationPathTableDeepDiamondMeasures61"
        status: pass
    human_judgment: false
  - id: D2
    description: "Native preflight admits 4096 nodes, refuses the attempted 4097th with stable context, and cannot be bypassed inertly."
    requirement: OBS-04
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestInvocationPathTableBoundary"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestInvocationPreflightGuardIsNotInert"
        status: pass
    human_judgment: false
status: complete
---

# Phase 15 Plan 04 Summary

Native emission now unfolds and bounds invocation occurrences before generated-C serialization.

## Accomplishments

- Added a fixed, independent 4096-node invocation preflight immediately after callgraph and entry validation.
- Pinned the tracked `deep_diamond_acyclic.lang` fixture at the source-derived recurrence `1, 5, 13, 29, 61`.
- Added exact 4096/4097 synthetic controls, a stable `cgen.invocation_path_table_exceeded` witness, and a bypass mutation test that reaches serialization only when the guard is disabled.

## Verification

Passed with a writable temporary Go build cache:

- `go test ./internal/compiler/cgen -run 'TestInvocationPathTableDeepDiamondMeasures61|TestInvocationPreflightOrdering' -count=1 -v`
- `go test ./internal/compiler/cgen -run 'TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert' -count=1 -v`
- `go test ./internal/compiler/cgen -count=1`

## Task Commits

No commits were possible in this sandbox: Git could not create `.git/worktrees/agent-p04/index.lock` (`operation not permitted`). The implementation and this summary remain staged only in the working tree for a writable Git session.

## Deviations from Plan

None in implementation scope. The requested atomic commits were blocked by worktree Git metadata permissions.

## Files Modified

- `internal/compiler/cgen/cgen_program.go` — bounded deterministic invocation preflight and typed overflow refusal.
- `internal/compiler/cgen/cgen_program_test.go` — fixture, ordering, boundary, and mutation-kill controls.
- `internal/compiler/cgen/export_test.go` — test-only preflight observation hooks.

## Next Phase Readiness

The invocation preflight metadata and exact cap are available for the subsequent static path-table emission work. No production override, environment control, or adaptive heuristic exists.

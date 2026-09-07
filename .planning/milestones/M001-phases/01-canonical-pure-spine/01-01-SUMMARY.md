---
phase: 01-canonical-pure-spine
plan: "01"
subsystem: compiler
tags: [go, parser, typed-core, interpreter, c17, clang]
requires: []
provides:
  - "Real S1 source-to-typed-core compiler path"
  - "Deterministic semantic interpreter over checked core"
  - "Readable C17 emission with Clang O0/O3 equivalence"
affects: [frontend, ownership, evidence, native]
actuals:
  tokens: 10076
  tasks: 3
  commits: 3
tech-stack:
  added: ["Go 1.24 standard library", "C17/Clang host tool"]
  patterns: ["session composition root", "typed core shared by independent engines", "argv-only isolated native execution"]
key-files:
  created: [go.mod, cmd/lang/main.go, internal/compiler/syntax/parser.go, internal/compiler/core/core.go, internal/compiler/interp/interp.go, internal/compiler/cgen/cgen.go, internal/compiler/native/native.go]
  modified: [internal/compiler/session/session.go, internal/compiler/session/session_test.go]
key-decisions:
  - "The interpreter is the semantic oracle; generated C is checked against it at both O0 and O3."
  - "Clang is invoked by literal argv in a fresh temporary directory and tool failure is separate from invalid source."
  - "S1 semantic IDs use an explicit provisional namespace rather than Go object or file identities."
patterns-established:
  - "Compiler stage DAG: syntax/AST -> checking/core -> interpreter and C backend -> session."
  - "Invalid checked programs never enter an execution engine."
requirements-completed: [FND-01, SYN-01, SEM-01, SEM-02, INT-01, NAT-01]
coverage:
  - id: D1
    description: "S1 source parses and checks into immutable nominal typed core while the missing-arm control is rejected."
    requirement: SEM-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestTogglePipeline and TestNonExhaustiveMatch"
        status: pass
    human_judgment: false
  - id: D2
    description: "The deterministic interpreter maps Off to On and On to Off through checked match arms."
    requirement: INT-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestInterpreterDeterministic"
        status: pass
    human_judgment: false
  - id: D3
    description: "Readable C17 compiled at O0 and O3 agrees with the interpreter and missing Clang remains operational failure."
    requirement: NAT-01
    verification:
      - kind: e2e
        ref: "internal/compiler/session/session_test.go#TestNativeToggleO0O3 and TestNativeToolFailureIsOperational"
        status: pass
    human_judgment: false
duration: 10min
completed: 2026-09-03
status: complete
---

# Phase 1 Plan 1: Executable Semantic Tracer Summary

**A real braced Lang program now checks into typed core, executes deterministically, and agrees with readable C17 at Clang `-O0` and `-O3`.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-09-03T18:47:00Z
- **Completed:** 2026-09-03T18:57:08Z
- **Tasks:** 3
- **Files modified:** 18

## Accomplishments

- Parsed the S1 module/export/data/function/match surface into a lossless token tree and projected immutable nominal typed core with stable provisional IDs.
- Rejected a missing match alternative before execution and ran valid core through a deterministic semantic event oracle.
- Emitted readable C17, compiled it through shell-free isolated Clang invocations, and matched interpreter results at `-O0` and `-O3`.

## Task Commits

1. **Task 1: Source, syntax, and typed core** — `032b837`
2. **Task 2: Deterministic semantic interpreter** — `9f0d1a1`
3. **Task 3: C17 and native equivalence** — `3283775`

## Files Created/Modified

- `cmd/lang/main.go` — root check and interpreter/native run adapter.
- `internal/compiler/syntax/*` — UTF-8 lexer, lossless token tree, and S1 parser.
- `internal/compiler/core/core.go` — versioned serializable typed-core records.
- `internal/compiler/check/check.go` — nominal type and exhaustive-match checking.
- `internal/compiler/interp/interp.go` — deterministic outcome/event oracle.
- `internal/compiler/cgen/cgen.go` — deterministic readable C17 emission.
- `internal/compiler/native/native.go` — bounded Clang compilation/execution without a shell.
- `internal/compiler/session/*` — composition root and end-to-end tests.

## Decisions Made

- Kept all mutable compiler state request-local, so check/run can be safely parallelized later.
- Normalized native output back into the semantic event model rather than treating C stdout bytes as language semantics.
- Deferred formatter recovery and complete structured CLI projection to their planned waves instead of enlarging this tracer.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The sandbox denied the default Go build cache path; verification used the explicit disposable cache `/tmp/ai-lang-gocache`. Product code and semantics were unaffected.

## User Setup Required

None - native execution detects the existing Clang host tool and interpreter execution remains available independently.

## Next Phase Readiness

The vertical compiler path is operational. Plan 01-02 can now harden the exact same syntax path with canonical formatting, comments, bounded recovery, and property evidence.

## Self-Check: PASSED

- All named task tests passed.
- `go test ./...` and `go vet ./...` passed with the explicit sandbox cache.
- Every key created file exists and all three task commits are present.

---
*Phase: 01-canonical-pure-spine*
*Completed: 2026-09-03*

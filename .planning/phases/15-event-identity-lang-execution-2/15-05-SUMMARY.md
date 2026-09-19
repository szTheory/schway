---
phase: 15-event-identity-lang-execution-2
plan: "05"
subsystem: native evidence emission
tags: [cgen, execution-schema-2, invocation, call-edges]
requires:
  - phase: 15-event-identity-lang-execution-2
    provides: bounded deterministic invocation preflight and interpreter /2 semantics
provides:
  - Static parent-indexed invocation and child-index tables in generated C
  - Native lang.execution/2 events with caller-owned function.called edges
affects: [native-emission, execution-validation, session-differential]
actuals:
  tokens: 7126
  tasks: 2
  commits: 4
plan_head_before: 26ac070
tech-stack:
  added: []
  patterns: [static occurrence tables, uniform unsigned invocation ABI, schema-specific frozen event writers]
key-files:
  created: [.planning/phases/15-event-identity-lang-execution-2/15-05-SUMMARY.md]
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/cgen_names_test.go
key-decisions:
  - "Generate literal invocation strings and call-site tables from the same bounded preorder nodes."
  - "Keep /2 support separate from legacy writer generation so /0 and /1 bytes remain frozen."
patterns-established:
  - "Generated internal C functions uniformly receive an unsigned invocation index."
requirements-completed: [OBS-01, OBS-02, OBS-04, NAT-10]
coverage:
  - id: D1
    description: "Static native invocation tables preserve distinct activation occurrences and thread indices through internal calls."
    requirement: NAT-10
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestProgramInvocationIndexThreading; TestParentIndexedChildLookup; TestInvocationTableEmissionIsDeterministic"
        status: pass
    human_judgment: false
  - id: D2
    description: "Native multi-function execution writes interpreter-equivalent /2 invocation evidence and preorder call edges."
    requirement: OBS-02
    verification:
      - kind: integration
        ref: "internal/compiler/cgen/cgen_program_test.go#TestNativeFunctionCalledPreorder; TestNativeFunctionCalledProjectionRemoval"
        status: pass
    human_judgment: false
  - id: D3
    description: "Legacy C event writers retain exact frozen output while schema selection is explicit."
    requirement: OBS-01
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestLegacyEventWritersFrozen"
        status: pass
    human_judgment: false
duration: 14min
completed: 2026-09-19
status: complete
---

# Phase 15 Plan 05: Native Invocation Identity Summary

**Native multi-function C now carries static per-occurrence identities and emits interpreter-equivalent `/2` call evidence without changing legacy writer bytes.**

## Accomplishments

- Rendered the bounded preflight preorder into literal invocation strings and parent-indexed child tables; no runtime path construction, allocation, or hashing is introduced.
- Threaded an unsigned invocation index through every generated internal prototype, definition, entry call, and `OpCall`.
- Added schema-specific `/2` event support with invocation on every native event and caller-owned `function.called` records immediately before the child call.
- Proved the shared-leaf diamond, preorder edge order, projection-only removal control, deterministic emission, and frozen legacy bytes.

## Verification

Passed:

- `go test ./internal/compiler/cgen -run 'TestProgramInvocationIndexThreading|TestParentIndexedChildLookup|TestInvocationTableEmissionIsDeterministic' -count=1 -v`
- `go test ./internal/compiler/cgen -run 'TestProgramWritesExecutionSchema2|TestNativeFunctionCalledPreorder|TestNativeFunctionCalledProjectionRemoval|TestLegacyEventWritersFrozen' -count=1 -v`
- `go test ./internal/compiler/cgen -count=1`

## Task Commits

1. `25265ba` — `test(15-05): add failing invocation table tests`
2. `c93f382` — `feat(15-05): thread static invocation indices`
3. `60d3566` — `test(15-05): add failing schema 2 native evidence tests`
4. `07511cf` — `feat(15-05): emit schema 2 native call evidence`

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Build correctness] Kept all emitted function definitions warning-free under `-Werror`.**

- **Found during:** Task 1
- **Issue:** Leaf functions do not consult the threaded index and the literal invocation table was initially not read before Task 2, triggering unused diagnostics.
- **Fix:** Added harmless explicit references while preserving the uniform ABI and static table.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`

**2. [Rule 1 - Regression coverage] Updated pre-existing C signature matchers.**

- **Found during:** Task 2 verification
- **Issue:** Name-allocation tests pinned the former one-argument generated-function ABI.
- **Fix:** Updated their prototype and definition patterns to include the required unsigned invocation index.
- **Files modified:** `internal/compiler/cgen/cgen_names_test.go`

## Next Phase Readiness

The native and interpreter producers now agree on `/2` occurrence identity and causal call edges. Downstream validation and differential work can consume the same literal, deterministic event facts.

## Self-Check: PASSED

- Confirmed all modified CGen files and this summary exist.
- Confirmed task commits `25265ba`, `c93f382`, `60d3566`, and `07511cf` exist in Git history.

---
*Phase: 15-event-identity-lang-execution-2*
*Completed: 2026-09-19*

---
phase: 16-branch-match-emitter-port
plan: 01
subsystem: native-emission
tags: [go, c17, cgen, schema-2, execution-evidence]
requires:
  - phase: 15-event-identity-lang-execution-2
    provides: schema-2 invocation and event serialization contracts
provides:
  - Direct ordinary-linear tracer through emitProgram
  - Derived schema-2 live-resource serialization with a mutation control
affects: [16-02, cgen, native-execution]
actuals:
  tokens: 2759
  tasks: 2
  commits: 5
commits: 5
plan_head_before: 20f6ae0020c5cc2faabb3166102f018d04682271
tech-stack:
  added: []
  patterns: [emitter-owned derived document tails, direct native tracer tests]
key-files:
  created: []
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/export_test.go
key-decisions:
  - "Keep the public Emit and EmitNative dispatch unchanged while proving direct emitProgram behavior."
  - "Serialize live_resources from one emitter-owned derivation, currently empty for admitted ordinary-linear shapes."
requirements-completed: [NAT-08]
coverage:
  - id: D1
    description: Direct emitProgram lowers and executes the ordinary tracked-transfer fixture as a schema-2 document.
    requirement: NAT-08
    verification:
      - kind: integration
        ref: internal/compiler/cgen/cgen_program_test.go#TestProgramOrdinaryLinearTracer
        status: pass
    human_judgment: false
  - id: D2
    description: Schema-2 live_resources is rendered from a derived value and a seed changes generated serialization.
    requirement: NAT-08
    verification:
      - kind: unit
        ref: internal/compiler/cgen/cgen_program_test.go#TestProgramLiveResourcesAreDerived
        status: pass
      - kind: unit
        ref: internal/compiler/cgen/cgen_program_test.go#TestProgramLiveResourceDerivationIsNotInert
        status: pass
    human_judgment: false
duration: 2 min
completed: 2026-09-19
status: complete
---

# Phase 16 Plan 01: Ordinary Linear Emitter Tracer Summary

**Direct `emitProgram` now executes the tracked ordinary-linear transfer as schema-2 evidence, with its `live_resources` tail serialized from an emitter-owned derivation.**

## Performance

- **Duration:** 2 min
- **Started:** 2026-09-19T21:18:20Z
- **Completed:** 2026-09-19T21:20:49Z
- **Tasks:** 2/2
- **Files modified:** 3

## Accomplishments

- Added an end-to-end direct-emitter tracer that checks the fixture, runs one generated C17 translation unit, and compares the schema-2 operation sequence to the independent interpreter sequence after applying the documented schema-2 invocation projection.
- Retained the Phase 15 ordering: graph/entry validation, supported-shape validation, invocation/output preflight, then serialization.
- Replaced the fixed schema-2 resource suffix with a derived collection and a focused mutation control proving a seeded value reaches generated serialization.

## Task Commits

1. **Task 1: Lower one ordinary linear program end to end through `emitProgram`** — `fba5ab0` (test)
2. **Task 2: Derive `live_resources` in the surviving emitter** — `528388d` (feat)

## Files Created/Modified

- `internal/compiler/cgen/cgen_program.go` — derives resource output once for preflight and serialization, then emits the schema-2 resource writer.
- `internal/compiler/cgen/cgen_program_test.go` — supplies the direct ordinary-linear tracer and derived-tail mutation evidence.
- `internal/compiler/cgen/export_test.go` — exposes the narrow test-only resource-derivation seam.

## Decisions Made

- Kept dispatch untouched; this plan proves the surviving emitter directly and leaves public cutover to the designated later plan.
- Modelled currently admitted ordinary-linear resource output as an explicit empty derivation, without importing a foreign resource ledger or ownership semantics.

## TDD Gate Compliance

The planned RED assertions failed intentionally in Go's test runner before implementation: the direct tracer exposed the schema-1/schema-2 projection difference, and the derived-tail test exposed the fixed literal. The runtime `tdd-red-evidence` checker only parses TAP/Node summaries and reports Go test output as `zero_tests_discovered`, so it could not mechanically certify the otherwise target-specific Go assertion. The focused GREEN and full package commands passed after implementation.

## Deviations from Plan

None - plan executed within the specified emitter and test scope. A narrow `export_test.go` accessor was added solely for the required mutation control.

## Issues Encountered

- A seeded non-empty collection is intentionally rejected by the native schema-2 returned-outcome validator; the mutation control therefore verifies generated serialization rather than attempting an invalid runtime document. Ordinary fixtures continue to derive and execute with the valid empty collection.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The ordinary linear tracer and derived resource tail provide the surviving-emitter baseline for the subsequent admission and convergence plans.

## Self-Check: PASSED

- Found `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_program_test.go`, and `internal/compiler/cgen/export_test.go`.
- Found task commits `fba5ab0` and `528388d` in git history.

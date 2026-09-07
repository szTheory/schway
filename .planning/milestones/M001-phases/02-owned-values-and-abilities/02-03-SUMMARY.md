---
phase: 02-owned-values-and-abilities
plan: "03"
subsystem: compiler
tags: [go, affine-ownership, diagnostics, property-testing, fuzzing, bounded-work]

requires:
  - phase: 02-owned-values-and-abilities
    plan: "02"
    provides: Sealed structural ability facts, linear typed core, and exact test selection
provides:
  - Three distinct ownership rejections with causal diagnostic/1 facts and lawful repairs
  - Byte-compatible Phase 1 diagnostic/0 identity under the unchanged command/0 envelope
  - Straight-line loan final-use discovery with independently checked intermediate states
  - Exact linear work accounting through 10,000 operations and bounded deterministic fuzz seeds
affects: [02-04, core-validation, owned-native-execution, evidence]

actuals:
  tokens: 8807
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns: [feature-specific diagnostic schemas, last-use prepass, forward affine state, independent state-machine oracle, counted linear work]

key-files:
  created:
    - internal/compiler/check/check_test.go
    - testdata/phase2/use_after_move.lang
    - testdata/phase2/move_while_borrowed.lang
    - testdata/phase2/implicit_noncopy.lang
  modified:
    - internal/compiler/diagnostic/diagnostic.go
    - internal/compiler/check/check.go
    - internal/compiler/session/session_test.go
    - internal/compiler/testsupport/cli_test.go

key-decisions:
  - "Legacy Error construction remains exactly diagnostic/0; only the explicit repair-bearing ownership constructor selects diagnostic/1."
  - "Straight-line shared loans expire immediately after their precomputed final use, with an indexed expiry schedule that keeps forward checking linear."
  - "Checker work is the exact sum of structural type nodes, one complete last-use scan, and the inspected forward prefix."

patterns-established:
  - "Causal ownership failures: primary spans identify the rejected use while ordered causes retain declaration, transfer, loan, type, and stable semantic identities."
  - "Independent support proof: tests duplicate the small ownership law and compare operations, loan endpoints, and every successful intermediate state."

requirements-progressed: [OWN-01]

coverage:
  - id: D1
    description: "Use after move, move during a future-used shared loan, and bare noncopy transfer fail before execution with distinct /1 causes and lawful sorted repairs while Phase 1 stays /0."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestTransferRequiresTake,TestUseAfterMoveDiagnostic,TestMoveWhileBorrowedDiagnostic,TestDiagnosticSchemaCompatibility,TestPhase1DiagnosticGoldenUnchanged"
        status: pass
      - kind: e2e
        ref: "internal/compiler/testsupport/cli_test.go#TestHumanJSONMixedDiagnosticVersionParity"
        status: pass
    human_judgment: false
  - id: D2
    description: "An independent straight-line model agrees on legality, operation facts, loan final uses, intermediate states, and exact linear work through 10,000 operations."
    requirement: OWN-01
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestOwnershipSequenceExhaustive,TestOwnershipWorkSeries,FuzzOwnershipLinear"
        status: pass
      - kind: other
        ref: "env GOCACHE=/tmp/ai-lang-phase2-cache go test ./..."
        status: pass
    human_judgment: false

duration: 12min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 03: Causal Ownership Rejection Summary

**Three affine ownership violations now fail before execution with stable repairable `/1` diagnostics, while an independent state model proves straight-line loan endpoints and exact linear checker work without changing Phase 1 `/0` identities.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-09-03T20:44:43Z
- **Completed:** 2026-09-03T20:56:35Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Added canonical use-after-move, future-used-loan conflict, and implicit noncopy-transfer fixtures whose invalid programs produce no execution events.
- Added feature-specific `lang.diagnostic/1` construction with semantic causes and canonically sorted repair kinds, leaving the legacy constructor and Phase 1 golden exactly on `/0`.
- Implemented a linear last-use prepass and forward affine state machine with stable loan IDs, expiry facts, intermediate initialized/active-loan snapshots, and prefix-aware work counts.
- Independently exhausted short sequences, pinned the minimal declaration/borrow/move/later-read witness, checked 10/100/1,000/10,000 operation scales, and added bounded request-local fuzz seeds.

## Task Commits

Each TDD task was committed with a RED contract followed by its GREEN implementation:

1. **Task 02-03-01 RED: ownership diagnostic contracts** - `860a85f` (test)
2. **Task 02-03-01 GREEN: causal ownership rejection** - `81ccac6` (feat)
3. **Task 02-03-02 RED: independent ownership support oracle** - `cb50e07` (test)
4. **Task 02-03-02 GREEN: bounded ownership support facts** - `d51e5f8` (feat)

## Files Created/Modified

- `internal/compiler/diagnostic/diagnostic.go` - Coexisting `/0` and repair-bearing `/1` constructors with prose-independent identity.
- `internal/compiler/check/check.go` - Straight-line final-use discovery, affine transitions, causal failures, support snapshots, and counted work.
- `internal/compiler/check/check_test.go` - Independent exhaustive oracle, minimal loan witness, four scale points, and capped fuzz seeds.
- `internal/compiler/session/session_test.go` - Compile-reject, no-execution, schema compatibility, lawful repair, and Phase 1 golden contracts.
- `internal/compiler/testsupport/cli_test.go` - Mixed diagnostic-version identity parity through human and JSON command projections.
- `testdata/phase2/use_after_move.lang` - Explicit second transfer from a moved owner.
- `testdata/phase2/move_while_borrowed.lang` - Four-step future-used shared-loan conflict witness.
- `testdata/phase2/implicit_noncopy.lang` - Bare Buffer binding that requires explicit `take`.

## Decisions Made

- Repair detail remains display/support data; schema, code, primary span, ordered semantic causes, and sorted repair kinds define `/1` identity.
- A borrow with no later use expires after its creation operation, so the following owner move is legal; a later read keeps the loan active and rejects the intervening move.
- Work accounting uses semantic units instead of wall time: type nodes plus a complete linear final-use scan plus the actually inspected forward prefix.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The independent oracle initially used nil for an empty operation slice while production intentionally returned an initialized empty slice. The oracle was normalized to the same inert empty fact before GREEN; no semantic expectation changed.

## Verification Evidence

- The Task 1 exact fail-on-zero selector passed all six named ownership, compatibility, projection, and golden tests.
- The Task 2 exact fail-on-zero selector passed exhaustive sequence comparison, 10/100/1,000/10,000 work counts, and `FuzzOwnershipLinear` seeds.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` exited 0 across every package.
- `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/verify-phase1.sh` exited 0 after tests, race, vet, and all five Phase 1 verification lanes; it reported 18 recomputed work units.
- `git diff --check` reported no whitespace errors after the final implementation change.

## Known Stubs

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02-04 can independently validate materialized ownership operations, ability facts, IDs, and support claims before either execution engine consumes them.
- `OWN-01` is progressed by this plan but remains pending until final Phase 2 verification; `OWN-01` and `OWN-02` were not marked complete here.
- No blockers remain for sequential Phase 2 execution.

## Self-Check: PASSED

All eight implementation/test/fixture files, this summary, and all four TDD commits were found. The summary contains `requirements-progressed: [OWN-01]` and no requirement-completion claim.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*

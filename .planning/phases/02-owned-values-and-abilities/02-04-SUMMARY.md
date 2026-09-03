---
phase: 02-owned-values-and-abilities
plan: "04"
subsystem: compiler
tags: [go, typed-core, validation, affine-ownership, mutation-testing, bounded-work]

requires:
  - phase: 02-owned-values-and-abilities
    plan: "03"
    provides: Materialized sealed ability facts, linear ownership operations, and causal affine rejection
provides:
  - Source-blind typed-core validation before interpreter and session engine entry
  - Independent sealed ability recomputation and affine move/shared-loan replay
  - One-boundary mutation controls for identities, references, abilities, transitions, and final claims
  - Exact 16n+13 validation work over 101, 1,001, and 10,001 canonical facts
  - Executable expected escape for a coordinated source-to-core lie
affects: [02-05, native-owned-execution, evidence, trust-boundaries]

actuals:
  tokens: 8015
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns: [source-blind validation, independently duplicated semantic law, content-owned trust-boundary copy, exact counted work]

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/session/session.go

key-decisions:
  - "The existing terminal return operation is the owned core's explicit final claim; validation requires it to be unique, last, initialized, type-consistent, and identity-ordered."
  - "Validator ability and transition authorization is independently implemented from checker and interpreter behavior, sharing only inert core records."
  - "The canonical scale shape has an exact counted validation formula of 16n+13, with loan final-use consolidation remaining linear."
  - "Internally consistent false producer facts remain the named expected escape: escape:coordinated-source-core-lie."

patterns-established:
  - "Fail-closed engine boundary: session validates a content-owned core copy before dispatch, and direct interpreter entry independently refuses invalid raw core."
  - "Honest assurance claim: core validation proves internal consistency and artifact integrity, never source translation truth."

requirements-progressed: [OWN-01, OWN-02]

coverage:
  - id: D1
    description: "A source-blind validator independently rejects forged abilities, ambiguous identities, invalid references, illegal affine transitions, and inconsistent final claims before interpreter execution."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestOwnershipMutationMatrix,TestUnvalidatedCoreCannotExecute,TestArbitraryMaskCannotEnterCoreValidation"
        status: pass
      - kind: other
        ref: "env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOwnershipMutationMatrix TestUnvalidatedCoreCannotExecute TestArbitraryMaskCannotEnterCoreValidation"
        status: pass
    human_judgment: false
  - id: D2
    description: "Validation performs exact linear counted work, preserves semantic order, owns its input storage, and records the coordinated source-to-core lie only as an expected escape."
    requirement: OWN-02
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCoreValidationWorkSeries,TestOwnedClaimReorderRejected,TestCoordinatedSourceCoreEscapeIsNamed,TestValidationOwnsTrustBoundaryCopy"
        status: pass
      - kind: other
        ref: "env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestCoreValidationWorkSeries TestOwnedClaimReorderRejected TestCoordinatedSourceCoreEscapeIsNamed"
        status: pass
    human_judgment: false

duration: 13min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 04: Independent Owned Core Validation Summary

**Owned typed core now crosses an independent, content-owning semantic validator before execution, with exact linear work and an executable statement of the coordinated-producer escape.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-09-03T21:01:29Z
- **Completed:** 2026-09-03T21:14:13Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added a source/checker/engine-blind validator that recomputes Byte, Buffer, Box, and Pair abilities from sealed TypeRefs and independently replays copy, move, shared-loan, and return transitions.
- Added fail-closed schema, closed-body, stable-ID, point-order, reference, match-arm, ability-witness, operation, and final-claim checks over content-owned core copies.
- Gated both session engine paths and direct interpreter entry so forged or unvalidated core cannot execute.
- Proved exact `16n+13` work at 101, 1,001, and 10,001 operation/final-claim facts and kept loan final-use consolidation linear.
- Demonstrated that a self-consistent coordinated source/core falsehood is accepted and recorded only as `escape:coordinated-source-core-lie`.

## Task Commits

Each TDD task was committed with a RED contract followed by its GREEN implementation:

1. **Task 02-04-01 RED: owned core mutation contracts** - `b5013bb` (test)
2. **Task 02-04-01 GREEN: independent execution gate** - `9b0e17b` (feat)
3. **Task 02-04-02 RED: validator bound and escape contracts** - `eacbd65` (test)
4. **Task 02-04-02 GREEN: counted validation and named escape** - `e603534` (feat)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - Source-blind schema/reference/ability/transition/final-claim validator with content ownership and exact work accounting.
- `internal/compiler/corevalidate/corevalidate_test.go` - One-boundary mutation matrix, alias controls, semantic reorder control, three-point scale series, and expected escape fixture.
- `internal/compiler/interp/interp.go` - Direct raw-core entry validation before the interpreter's private dynamic replay.
- `internal/compiler/session/session.go` - Explicit validation and deep-copy gate between checking and interpreter/native engine orchestration.

## Decisions Made

- The final `return` operation is the current core's explicit final claim, so no out-of-scope checker representation change was needed.
- The validator duplicates the small ability and transition laws; it does not import producer or engine packages and does not share authorization helpers.
- Operation order is semantic and identity-bearing; validation never normalizes reordered operations or final claims.
- The expected coordinated-producer escape remains evidence of an assurance boundary, not a detected negative control.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Closed the legacy match-reference validation seam**

- **Found during:** Final threat-surface scan after Task 02-04-02
- **Issue:** Initial validation strictly checked owned linear facts but the legacy match branch only checked nonempty identities, leaving invalid type, scrutinee, alternative, or exhaustiveness references fail-open at the now-shared interpreter boundary.
- **Fix:** Added independent data-type, alternative, scrutinee, arm, and exhaustive final-claim validation for `lang.core/0` without changing its schema or execution facts.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** Both exact Plan 02-04 selectors, full Go suite, and `scripts/verify-phase1.sh` passed.
- **Committed in:** `e603534`

---

**Total deviations:** 1 auto-fixed missing critical validation
**Impact on plan:** The adjustment makes the shared execution boundary uniformly fail closed and preserves Phase 1 byte-compatible evidence.

## Issues Encountered

- The exact scale constant was corrected as validation coverage expanded from `16n+12` to `16n+13`; all three scale points now assert equality to the final formula.
- The combined regression command exceeded one 30-second tool response window during its race lane; the continuing process and subsequent standalone runs completed successfully with explicit exit code 0.

## Verification Evidence

- Both exact fail-on-zero selectors passed all six plan-named tests.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` passed across every package.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test -race ./...` and `go vet ./...` passed.
- `env GOCACHE=/tmp/ai-lang-phase2-cache sh scripts/verify-phase1.sh` passed tests, race, vet, and all five Phase 1 lanes with 18 recomputed work units.
- Forbidden-import and transition-helper scans confirmed `corevalidate` shares only inert `core` records.
- `git diff --check` reported no whitespace errors.

## Known Stubs

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02-05 can consume only validated owned core while adding strict native execution decoding and full interpreter/O0/O3 event comparison.
- `OWN-01` and `OWN-02` are progressed but remain pending until the remaining Phase 2 plans and phase verification complete.
- No blockers remain for sequential Phase 2 execution.

## Self-Check: PASSED

All four changed implementation/test files, this summary, and all four RED/GREEN commits were found. The summary contains `requirements-progressed: [OWN-01, OWN-02]`; `REQUIREMENTS.md` and `.planning/milestone.lock` were not modified.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*

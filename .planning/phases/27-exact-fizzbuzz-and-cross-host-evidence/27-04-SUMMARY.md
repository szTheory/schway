---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 04
subsystem: compiler
tags: [u64, equality, remainder, interpreter, c17, validation]

# Dependency graph
requires:
  - phase: 27-03
    provides: Typed U64 equality and remainder operations, static zero-divisor refusal, and dynamic-divisor admission.
provides:
  - Independent core and origin validation for U64 equality/remainder shapes and propagated zero facts.
  - Exact nonzero interpreter/native arithmetic and a reached zero-divisor failure with stable attribution.
  - Public application exit 65 with exact stderr for dynamic remainder by zero.
affects: [U64-02, 27-05, phase27-evidence]

# Actuals (#2632)
actuals:
  tokens: 10524
  tasks: 2
  commits: 1

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Independent peers reconstruct scalar constants, Bool values, and initialized states rather than trusting checker allow bits.
    - C17 remainder lowering branches on a zero divisor before evaluating `%`, with separate public-process and execution-document failure channels.

key-files:
  created: []
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/executionpeer/executionpeer.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase27_remainder_test.go

key-decisions:
  - "Keep equality and remainder shape/fact derivation independent in both validation peers; reject any divisor proven zero while allowing unknown runtime divisors."
  - "Use the same fixed execution reason for interpreter and native evidence, while public applications report the specified diagnostic and exit status."
  - "Preserve U64-02's spec-less edge probe as unclassified; this summary records the locked implementation evidence only."

patterns-established:
  - "Execution peers can admit a checked arithmetic defect only with the operation's fixed reason, source/type attribution, empty target, and terminal outcome."
  - "Native zero guards are tested structurally and by executing the zero path, so source order and reached behavior have separate evidence."

requirements-completed: [U64-02]

coverage:
  - id: D1
    description: "Dynamic nonzero U64 remainder and equality produce exact results in interpreter and native applications; independently validated forged operand/result shapes and known-zero facts are rejected."
    requirement: U64-02
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/session -run '^TestPhase27(EqualityRemainder|ForgedRemainder)$|^TestPhase26(BackEdgeAuthority|ScalarCopyCycleBoundary)$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "A reached dynamic zero divisor is guarded before native `%`; interpreter/native execution documents agree, and the public app exits 65 with empty stdout and exact diagnostic stderr."
    requirement: U64-02
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/session -run '^TestPhase27(DynamicZero|ForgedRemainder)$|^TestPhase26(AnalysisExhaustion|OverflowProcessOutcome)$' -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 38min
completed: 2026-10-03
status: complete
---

# Phase 27 Plan 04: Independent Remainder Validation and Guarded Execution

**U64 equality and remainder now run through independent validators and both engines, with a reached zero guard and exact public failure outcome.**

## Performance

- **Duration:** 38 min
- **Started:** 2026-10-03T15:04:00-04:00
- **Completed:** 2026-10-03T15:42:00-04:00
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Added independent typed/fixed-point admission for equality and remainder, including rejection of forged shapes and statically known zero divisors.
- Ran dynamic nonzero equality/remainder through interpreter and native application paths with exact output.
- Reached dynamic zero in both engines, proved generated C checks before `%`, matched execution documents, and pinned public exit 65 with exact stderr and no stdout.
- Preserved the Phase 26 loop-authority, analysis-exhaustion, and overflow controls.

## Task Commits

The two interdependent tasks were committed together as one cross-layer implementation change:

1. **Tasks 1–2: Validate and execute checked equality/remainder** - `db01336` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - Independent equality/remainder shape, initialization, and zero-fact derivation.
- `internal/compiler/originvalidate/originvalidate.go` - Separate origin-peer arithmetic and scalar fixed-point validation.
- `internal/compiler/interp/interp.go` - Equality/remainder execution and stable zero-divisor defect.
- `internal/compiler/cgen/cgen_program.go` - Equality lowering, pre-`%` zero guard, and public/execution failure paths.
- `internal/compiler/executionpeer/executionpeer.go` - Remainder-defect attribution and terminal validation.
- `internal/compiler/session/session.go` - Schema 2 reason projection for remainder defects.
- `internal/compiler/session/session_phase27_remainder_test.go` - Positive results, forged-core controls, reached-zero execution, and exact public outcome.

## Decisions Made

- Kept each validation peer's arithmetic facts independent so checker-produced metadata is not treated as proof.
- Kept unknown divisors admissible for a runtime guard, while rejecting every divisor independently proven to be zero.
- Used separate failure channels for machine-readable conformance evidence and the public app's exact diagnostic/status contract.

## Deviations from Plan

The plan's `files_modified` list omitted `executionpeer.go` and `session.go`. Both were required to preserve the requested conformance-document route for the new terminal defect: the projector supplies the stable reason, and the execution peer validates its attribution. The interpreter/native execution-document equality test and peer validation pass.

**Total deviations:** 1 necessary file-list extension. No scope expansion.

## Issues Encountered

- Dynamic-zero conformance evidence required threading the new fixed reason through schema 2 projection and independent execution-peer admission; this was completed and covered by `TestPhase27DynamicZero`.

## User Setup Required

None - installed Clang was used by the focused native application tests; no external service configuration is required.

## Next Phase Readiness

- Plan 27-05 can consume the now-runnable arithmetic path.
- Both plan acceptance selectors passed, along with focused Phase 26 regression selectors. No full project suite was run.
- U64-02's spec-less edge probe remains unclassified; the broader phase evidence and documentation contract are still pending.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*

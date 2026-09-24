---
phase: 18-branch-on-a-computed-value
plan: 03
subsystem: compiler-validation
tags: [go, typed-core, corevalidate, originvalidate, computed-match]
requires:
  - phase: 18-02
    provides: computed terminal match representation and source-to-core prefix operations
provides:
  - Core admission requires computed scrutinees to be defined by the shared entry prefix
  - Origin derivation scopes branch returns to their own arm and the shared prefix
  - Peer-local controls for malformed scrutinees and forged payload origins
affects: [phase-18, computed-match, corevalidate, originvalidate]
actuals:
  tokens: 4472
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [entry-prefix place derivation, return-arm-scoped origin walk]
key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_phase18_test.go
    - internal/compiler/originvalidate/originvalidate_phase18_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
key-decisions:
  - "Core computed scrutinee IDs must resolve to a place produced by an operation in the branching entry block; ID-less legacy matches remain parameter-only."
  - "Origin walks may follow earlier definitions in the return's own block and shared entry block, and do not use Match metadata to decide whether block scoping applies."
requirements-completed: []
coverage:
  - id: D1
    description: "Core admission accepts a valid computed prefix and rejects an arm-local, forged, or wrong-type scrutinee."
    requirement: CTL-01
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/corevalidate -run 'Test(Phase18|CoreValidateBranch)' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Origin admission derives a payload return through a computed alias and refuses understated or sibling-arm payload origins."
    requirement: CTL-03
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/originvalidate -run 'Test(Phase18|OriginValidate.*Payload)' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: "The accepted computed-match source fixture passes origin admission and the Phase 18 quick validation packages remain green."
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -count=1"
        status: pass
      - kind: integration
        ref: "go test ./..."
        status: pass
    human_judgment: false
duration: 23min
completed: 2026-09-24
status: complete
plan_head_before: be25df0e5bb7a37f6ef5dd603fe2b53aaa006ddc
commits: 2
---

# Phase 18 Plan 03: Independent Admission Peers Summary

**Core admission now proves computed scrutinees come from the shared entry prefix, while origin admission confines payload provenance to the return arm and that prefix.**

## Performance

- **Duration:** approximately 23 minutes
- **Started:** 2026-09-24 (time not captured at executor start)
- **Completed:** 2026-09-24
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added an independent core check that resolves an explicit computed scrutinee ID to a definition in the branch entry block; existing replay continues to validate type, ownership, and global operation/place ordinals.
- Added negative controls for an arm-local place, a forged place ID, and a computed place claiming a non-data type, while retaining parameter-scrutinee controls.
- Scoped origin walks to earlier operations in the returning arm and the shared entry prefix, preventing an understated claim or sibling-arm definition from lending origin evidence to a return.
- Added a source acceptance control for the computed-match fixture and synthetic computed-payload origin controls, including a sibling-arm mutation with the Match summary removed.

## Task Commits

1. **Task 1: Independently validate computed scrutinee place and type in core** - `8201c82` (RED control).
2. **Plan implementation and metadata** - `6195d66` (`feat(18-03): harden computed match admission peers`).

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - derives computed-place admission from entry-block definitions.
- `internal/compiler/corevalidate/corevalidate_phase18_test.go` - independent acceptance and forgery controls.
- `internal/compiler/originvalidate/originvalidate.go` - bounds origin derivation to preceding, arm-local and entry-prefix operations.
- `internal/compiler/originvalidate/originvalidate_phase18_test.go` - computed payload, understated-origin, sibling-arm, and source acceptance controls.

## Decisions Made

- Retained ID-less parameter matches as the legacy path; computed local matches must carry an explicit place ID.
- Treat block membership as origin-scope evidence even if a forged core omits the Match summary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Compatibility bug] Kept legacy parameter scrutinee resolution parameter-only.**
- **Found during:** Task 1 validation
- **Issue:** Looking up ID-less scrutinees across the flattened place list selected same-named arm aliases and rejected valid existing parameter matches.
- **Fix:** ID-less matches resolve only to the function parameter; computed matches require an explicit ID and entry-prefix producer.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** The core peer tests and Phase 18 quick validation passed after the adjustment.
- **Commit:** `6195d66`.

**2. [Rule 2 - Independent origin scope] Bound each branch return to its own definitions.**
- **Found during:** Task 2
- **Issue:** A whole-function source map could let a malformed return trace through a later or sibling-arm payload definition.
- **Fix:** Require every traversed definition to precede the return and reside in its block or the shared entry block.
- **Files modified:** `internal/compiler/originvalidate/originvalidate.go`, `internal/compiler/originvalidate/originvalidate_phase18_test.go`
- **Verification:** Origin-focused controls and the quick validation passed.
- **Commit:** `6195d66`.

**Total deviations:** 2 auto-fixed (1 Rule 1, 1 Rule 2). **Impact:** Both changes close peer correctness gaps exposed by Phase 18 computed-place inputs.

## Verification

- Passed `go test ./internal/compiler/corevalidate -run 'Test(Phase18|CoreValidateBranch)' -count=1`.
- Passed `go test ./internal/compiler/originvalidate -run 'Test(Phase18|OriginValidate.*Payload)' -count=1`.
- Passed the Phase 18 quick validation command: `go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -count=1`.
- Passed the full Go suite after the final origin-scope tightening; `GOCACHE=/tmp/ai-lang-gocache go test ./...` exited 0 (session package: 126.138s).
- Confirmed the new peer tests use the public source/session entry point only and import no checker derivation helpers.

## Issues Encountered

- The executor environment could not create `.git/index.lock`; the parent committed the scoped implementation and plan artifacts through the GSD helper.
- The CTL-03 payload fixture is not yet checker-valid at this plan boundary, so the valid-source origin acceptance control uses the already accepted computed-match fixture; synthetic core tests separately exercise computed payload provenance. Later phase plans retain full CTL acceptance responsibility.

## Next Phase Readiness

Both peers now have independent computed-place boundaries and automated malformed-core controls. CTL-01, CTL-02, and CTL-03 remain open until phase-level verification proves the complete requirements.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-24*

## Self-Check: PASSED

- The summary and both Phase 18 peer test files exist.
- The RED control commit `8201c82` and implementation/metadata commit `6195d66` are recorded.
- Focused core, focused origin, quick validation, and full `go test ./...` results are recorded above.

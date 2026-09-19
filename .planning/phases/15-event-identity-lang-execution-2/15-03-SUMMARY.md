---
phase: 15-event-identity-lang-execution-2
plan: 03
status: complete
subsystem: execution
tags: [execution, invocation, causal-validation, peer]
requires: [15-01]
provides: [executionpeer]
affects: [session]
key_files:
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_test.go
coverage:
  - id: D1
    description: Independent bounded invocation membership and entry resolution
    requirement: OBS-03
    verification:
      - kind: unit
        ref: internal/compiler/executionpeer/executionpeer_test.go#TestIndependentEntryResolution
        status: pass
      - kind: unit
        ref: internal/compiler/executionpeer/executionpeer_test.go#TestInvocationMembershipTraversal
        status: pass
    human_judgment: false
  - id: D2
    description: Observed causal structure, ownership, and preorder validation
    requirement: OBS-04
    verification:
      - kind: unit
        ref: internal/compiler/executionpeer/executionpeer_test.go#TestValidateObservedCausalStructure
        status: pass
      - kind: unit
        ref: internal/compiler/executionpeer/executionpeer_test.go#TestExecutionPeerFailuresAreActionable
        status: pass
    human_judgment: false
---

# Phase 15 Plan 03 Summary

Implemented an independent `lang.execution/2` causal-document peer.

## Delivered

- Added independent entry resolution, bounded deterministic invocation unfolding, and a 4096-node refusal.
- Added observed-only validation for canonical membership, pair uniqueness, owner/function consistency, caller-owned call edges, and depth-first preorder.
- Kept static full-set equality in the separate `ValidateFullCoverage` control.
- Added structural import-boundary, diamond/shared-leaf, boundary, and actionable mutation tests.

## Verification

Passed:

`GOCACHE=/private/tmp/ai-lang-go-build go test ./internal/compiler/executionpeer -run 'TestIndependentEntryResolution|TestInvocationMembershipTraversal|TestExecutionPeerImportBoundary|TestFullCoverageControlIsSeparate' -count=1 -v`

`GOCACHE=/private/tmp/ai-lang-go-build go test ./internal/compiler/executionpeer -run 'TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable' -count=1 -v`

`GOCACHE=/private/tmp/ai-lang-go-build go test ./internal/compiler/executionpeer -count=1`

## Issues Encountered

The sandbox exposes this worktree's Git metadata read-only, so `git commit` cannot create the worktree index lock. The implementation and this summary are ready to commit from a Git-authorized environment.

## User Setup Required

None.

## Next Phase Readiness

The session comparator can now call `executionpeer.Validate` for program-aware `/2` admission.

---
*Phase: 15-event-identity-lang-execution-2*
*Completed: 2026-09-19*

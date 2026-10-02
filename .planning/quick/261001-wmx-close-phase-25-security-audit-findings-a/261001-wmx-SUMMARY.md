---
id: 261001-wmx
phase: quick
plan: 261001-wmx
status: complete
plan_head_before: 6f3c071f508288539978cdd8804a426c3b540f42
commits: 6
key-files:
  created: []
  modified:
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_pointer_successor_test.go
    - internal/compiler/session/session_pointer_successor_identity_test.go
    - internal/compiler/originvalidate/originvalidate.go
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VERIFICATION.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md
decisions:
  - Preserve CFG endpoint synthesis as-is and derive bounded straight-line loan facts independently from operations and place facts.
  - Keep EVD-10 pending until hosted macOS/Linux family-by-lane receipts exist.
metrics:
  completed: 2026-10-02
  tasks: 3
  commits: 5
---

# Quick Task 261001-wmx Summary

Phase 25 now has independent straight-line pointer-path refusal evidence for a real shared/exclusive overlap and a genuine borrowed-result escape, and command-level escape refusals point to both the escaping result and its borrow origin.

## Completed Tasks

| Task | Result | Commit |
| --- | --- | --- |
| 1. Refuse live pointer overlap and escape in the independent path oracle | Added operation-derived loan ancestry, use, access-family, and terminal escape validation; endpoint claims are ignored. Added empty/forged endpoint controls, positive shared-access controls, independent peer mutation checks, command refusal checks, and an observation test pinning actual command peer admission order. | `b78952c`, `f56b135`, `b19446f` |
| 2. Preserve source attribution on the actual escape refusal | Origin problems carry structured function and return-operation identity. The command projects the checked source to a primary escaping-result span and borrow-origin cause span without parsing peer prose or changing diagnostic schema/repair behavior. | `7472fbe` |
| 3. Reconcile Phase 25 claims | Updated only the named Phase 25 validation/verification and living roadmap/maturity evidence claims, distinguishing source inspection, current local checks, and historical receipts. | `3e6da03` |

## Verification

All three exact automated task commands passed:

- Task 1: `go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25PointerPath|TestPhase25UtilityOwnerTransfer|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25IndependentPeer|TestPhase25FamilyConflict)$' ./internal/compiler/pathoracle ./internal/compiler/session`
- Task 2: `go test -count=1 -run '^(TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyEscape|TestPhase25FamilyConflict|TestPhase25UnsupportedPointerShape|TestPhase25StructuredDiagnosticWireSchema|TestPhase25OwnershipDiagnosticBoundary)$' ./internal/compiler/session ./internal/compiler/diagnostic`
- Task 3: `go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyConflict|TestPhase25StructuredDiagnosticWireSchema|TestLanguageMaturityCountsAreCurrent)$' ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/session ./internal/compiler/diagnostic`

Additional checks passed:

- `GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^TestPhase25' ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/session ./internal/compiler/native`
- `GOCACHE=/private/tmp/schway-gocache go test -count=1 -run '^(TestPhase25LivingRoadmap|TestLanguageMaturityCountsAreCurrent)$' ./internal/compiler/native ./internal/compiler/session`
- The exact task 1, task 2, and task 3 commands were also rerun as needed during integration; the added peer-observation test passed directly and with the task 1 gate.
- `git diff --check`
- After the focused suite exposed legacy interprocedural regressions, commit `95ce75b` narrowed only local `OpCall` result ancestry and declared borrowed-return handling, with controls for deferred call provenance, declared return, overlap across calls, and owned-result escape inside call-containing functions.
- Fresh uncached repository-wide validation passed: `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...`.
- The exact Task 1, Task 2, and Task 3 validation commands were rerun after the integration repair; all exited 0.

The first post-edit Go invocation used the sandbox's default cache outside writable roots and failed with an access-denied cache error. Rerunning with the task-local `/private/tmp/schway-gocache` passed. Git metadata writes likewise required the managed elevated execution path; all five requested commits completed successfully.

## Evidence Boundary and Deferred Items

The source and local checks support the named overlap, escape, peer-independence, command-admission, source-attribution, schema, function-identity, and no-unsafe-repair claims. Hosted EVD-10 macOS/Linux family-by-lane evidence remains pending and is not claimed as complete. No dependencies were added. No stubs, skipped tests, or unrun plan verifications remain.

## Self-Check: PASSED

- All five original task commits plus the integration repair commit exist in Git history and match the per-task change groups above.
- The summary file exists at the requested quick-task path.
- `git status --short` contains only the quick-task artifact directory (PLAN.md and this SUMMARY.md); both are intentionally uncommitted.

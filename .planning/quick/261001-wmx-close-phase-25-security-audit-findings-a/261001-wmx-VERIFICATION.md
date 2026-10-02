---
phase: 261001-wmx
verified: "2026-10-02T06:00:16.881Z"
status: passed
score: 4/4 must-haves verified
covered_files:
  - .planning/LANGUAGE-MATURITY.md
  - .planning/PRODUCT-ROADMAP.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VERIFICATION.md
  - .planning/quick/261001-wmx-close-phase-25-security-audit-findings-a/261001-wmx-PLAN.md
  - .planning/quick/261001-wmx-close-phase-25-security-audit-findings-a/261001-wmx-SUMMARY.md
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_pointer_successor_identity_test.go
  - internal/compiler/session/session_pointer_successor_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - testdata/phase16/public-emitter-consumers.json
covered_digest: "v1:sha256:2823262b7add9ef0874a5a5199ca66e5ba2836e62765d645de90a59ebfbb33b2"
behavior_unverified: 0
overrides_applied: 0
---

# Quick Task 261001-wmx Verification Report

**Goal:** Close Phase 25 findings T-25-05/T-25-08 with independent overlap and escape controls, close T-25-09/T-25-12 with actual source-attributed command refusals and accurate evidence docs, while leaving hosted EVD-10 open.
**Verified:** 2026-10-02T06:00:16.881Z
**Status:** passed

## Goal Achievement

| # | Must-have truth | Status | Evidence |
|---|---|---|---|
| 1 | T-25-05/T-25-08: the independent path oracle rejects a real live shared/exclusive overlap and valid borrowed-result escape despite empty or forged endpoints; all three peers and command admission reject mutations while positive access sequences pass. | ✓ VERIFIED | `validateStraightLinePointerLoans` derives ancestry, last use, access family, and return escape from operations/place facts. It does not read `LoanEndpoints`; `RecomputeEndpoints` still returns early for straight-line bodies. The two named oracle tests mutate accepted checked cores and try both empty and forged endpoints. `TestPhase25IndependentPeerMutations` separately calls corevalidate, originvalidate, and pathoracle on each mutation and on three positive cases. `TestPhase25CheckCommandPeerGate` checks accepted and refused command sources, and `TestPhase25CheckCommandPeerObservation` asserts command peer order. Exact Task 1 command passed. |
| 2 | Actual `schway check` escape refusals identify the escaping result and borrow cause with stable diagnostic schema, deterministic multi-function attribution, and no unsafe repair. | ✓ VERIFIED | `CheckCommandFile` calls `originPeerRefusalDiagnostic` for origin refusals. It selects by structured function/return operation identity, resolves the result through checked core places and tokens, and falls back to a spanless refusal when facts are unavailable. `TestPhase25EscapeDiagnosticSourceAttribution` calls `CheckCommandFile` for exclusive and shared source cases and checks exact primary/cause bytes, nonempty spans, `core.origin_omitted`, schema, and zero repairs. `TestPhase25EscapeDiagnosticFunctionIdentity` checks a similar harmless return and changed peer prose. Exact Task 2 command passed. |
| 3 | T-25-09/T-25-12 diagnostic code, schema, identity rules, and refusal precedence remain valid without unsafe repair suggestions. | ✓ VERIFIED | The named command tests assert `core.origin_omitted`, schema, identity stability, and no repairs. `TestPhase25FamilyConflict` preserves `ownership.borrow_conflict` with an offending-borrow span; the focused command also runs unsupported-shape, structured-schema, and ownership-boundary tests. Exact Task 2 and Task 3 commands passed. |
| 4 | Phase 25 evidence claims are limited to supported controls and host receipts; EVD-10 remains open pending hosted macOS/Linux family-by-lane receipts. | ✓ VERIFIED | The Phase 25 validation and verification records name the new controls and distinguish fresh local checks from prior native receipts. The verification keeps the hosted matrix gap open and the original Phase 25 score at 4/5. PRODUCT-ROADMAP and LANGUAGE-MATURITY retain hosted receipts as pending. The source and evidence-index contract test `TestLanguageMaturityCountsAreCurrent` passed in the exact Task 3 command. |

**Score:** 4/4 must-haves verified; behavior-unverified: 0.

## Required Artifacts

| Artifact | Status | Evidence |
|---|---|---|
| `internal/compiler/pathoracle/pathoracle_pointer_successor_test.go` | ✓ VERIFIED | Separate overlap and real borrowed-result escape tests; both endpoint mutations exercised. |
| `internal/compiler/session/session_pointer_successor_test.go` | ✓ VERIFIED | Per-peer mutation assertions, command-gate cases, source attribution, conflict and diagnostic contract checks. |
| `internal/compiler/session/session.go` | ✓ VERIFIED | Command admission and structured source projection are wired into `CheckCommandFile`. |

## Key Links

| From | To | Via | Status | Evidence |
|---|---|---|---|---|
| `pathoracle.ValidateLocalOwnerPaths` | `session.CheckCommandFile` | independent application admission after core and origin peers | WIRED | Source order is corevalidate → originvalidate → pathoracle; observation test asserts accepted command invokes peers in that order. |
| `originvalidate.Problem` | command diagnostic | structured `FunctionID` and `ReturnOperationID` | WIRED | Session resolves the identified return operation and derives primary/cause spans from checked core and source tokens; it does not parse `Problem.Detail`. |
| command escape refusal | `testdata/phase25/exclusive_escape_reject.schway` and shared escape source | `CheckCommandFile` test cases | WIRED | The attribution test runs both actual command cases and compares selected source bytes. |

## Behavioral Spot-Checks

| Plan task | Exact command | Result | Status |
|---|---|---|---|
| Task 1 | `GOCACHE=/private/tmp/schway-verifier-gocache go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25PointerPath|TestPhase25UtilityOwnerTransfer|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25IndependentPeer|TestPhase25FamilyConflict)$' ./internal/compiler/pathoracle ./internal/compiler/session` | Both packages passed. | ✓ PASS |
| Task 2 | `GOCACHE=/private/tmp/schway-verifier-gocache go test -count=1 -run '^(TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyEscape|TestPhase25FamilyConflict|TestPhase25UnsupportedPointerShape|TestPhase25StructuredDiagnosticWireSchema|TestPhase25OwnershipDiagnosticBoundary)$' ./internal/compiler/session ./internal/compiler/diagnostic` | Both packages passed. | ✓ PASS |
| Task 3 | `GOCACHE=/private/tmp/schway-verifier-gocache go test -count=1 -run '^(TestPhase25PointerPathLiveOverlap|TestPhase25PointerPathBorrowedResultEscape|TestPhase25IndependentPeerMutations|TestPhase25CheckCommandPeerGate|TestPhase25EscapeDiagnosticSourceAttribution|TestPhase25EscapeDiagnosticFunctionIdentity|TestPhase25FamilyConflict|TestPhase25StructuredDiagnosticWireSchema|TestLanguageMaturityCountsAreCurrent)$' ./internal/compiler/pathoracle ./internal/compiler/originvalidate ./internal/compiler/session ./internal/compiler/diagnostic` | All packages passed; originvalidate reported no tests matching the regex, as expected because its relevant peer is called by the session test. | ✓ PASS |
| Post-integration regression | `GOCACHE=/private/tmp/schway-gocache go test -count=1 ./...` | Fresh uncached repository-wide suite passed, including the OpCall and declared-borrow-return boundaries, Phase 25 repair/reverify fixtures, and downstream CLI admission. | ✓ PASS |

Hosted EVD-10 remains intentionally open and is outside this quick task’s closure claim.

## Anti-patterns

No unreferenced `TBD`, `FIXME`, or `XXX` markers or implementation stubs were found in the changed implementation, tests, or Phase 25 evidence files inspected.

## Gaps Summary

No gaps found against this quick task’s must-haves. The Phase 25 hosted macOS/Linux family-by-lane evidence remains open as explicitly required; this report does not treat that open Phase 25 acceptance item as closed.

---

_Verified: 2026-10-02T04:47:15Z_  
_Verifier: gsd-verifier_

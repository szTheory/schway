---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "04"
subsystem: compiler
tags: [ownership, pointer-abi, corevalidate, interpreter, diagnostics]
requires:
  - phase: 25-03
    provides: checked shared/exclusive U64 copy families and integrated source composition
provides:
  - Exclusive U64 copy family emitted with pointer ABI and matching manifest
  - Exact independent owner-transfer proof for shared then exclusive copy helpers
  - Interpreter model proof for 0x41/0x42 results and 0x43 typed failure ordering
affects: [phase-25-plan-05, ownership, native-emission]
actuals:
  tokens: 10843
  tasks: 2
  commits: 8
plan_head_before: 9c427a20dbe21202ed74fa1858c8bf72e8aef993
commits: 8
tech-stack:
  added: []
  patterns: [checked-fact-driven pointer ABI, exact local helper-chain validation, modeled typed-failure cleanup traversal]
key-files:
  created: [internal/compiler/interp/interp_pointer_successor_test.go]
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_pointer_successor_test.go
    - internal/compiler/native/phase25_pointer_successor_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/diagnostic/diagnostic_test.go
    - internal/compiler/session/session_pointer_successor_test.go
key-decisions:
  - "Only explicit Foreign borrow/consume contracts exempt frame-drain tracking; nil or unknown contracts remain conservative."
  - "Owner transfer accepts either the prior direct result or the exact independently checked shared-copy then exclusive-copy U64 chain, after matching owner release."
  - "Pointer-body forms not reliably expressible in current source remain tested at production serializer admission; source tests cover expressible ownership boundaries."
patterns-established:
  - "Derive pointer declaration, call lowering, and manifest from the same checked ABI fact without unsupported optimizer or ownership attributes."
  - "On a modeled foreign failure, skip successful-path value operations and calls while continuing cleanup and the enclosing typed failure."
requirements-completed: []
coverage:
  - id: D1
    description: "Exclusive U64 helper uses actual pointer C ABI with a matching manifest and serializer refusal controls."
    verification:
      - kind: unit
        ref: "TestPhase25ExclusivePointerABI / TestPhase25ExclusivePointerManifest / TestPhase25ExclusivePointerWrongResult / TestPhase25ExclusivePointerRefusal"
        status: pass
    human_judgment: false
  - id: D2
    description: "Interpreter composition returns 65 and 66; modeled 0x43 failure occurs before helpers and drains the owner."
    verification:
      - kind: integration
        ref: "TestPhase25InterpreterComposition / TestPhase25ErrorBeforeHelpers"
        status: pass
    human_judgment: false
  - id: D3
    description: "Independent validation admits only the exact helper chain and source diagnostics retain stable boundaries."
    verification:
      - kind: unit
        ref: "TestPeerCalleeFrameDrained / TestPhase25OwnerTransferExactHelperChain / TestPhase25StructuredDiagnosticWireSchema / TestPhase25UnsupportedPointerShape / TestPhase25OwnershipDiagnosticBoundary"
        status: pass
    human_judgment: false
duration: 52min
completed: 2026-10-01
status: complete
---

# Phase 25 Plan 04: Separate Pointer Successors and Integrated Utility Summary

The exclusive U64 read/copy helper now uses a real pointer ABI, and the integrated model validates and executes the exact acquire/use/shared-copy/exclusive-copy/return flow.

## Performance

- **Duration:** 52 minutes
- **Started:** 2026-10-01T16:30:00-04:00 (approximate; first task commit at 16:52)
- **Completed:** 2026-10-01T17:22:00-04:00
- **Tasks:** 2
- **Files modified:** 10 plan-owned files

## Accomplishments

- Emitted the exclusive U64 read/copy helper as `uint64_t *`, with call-site address-taking and a matching manifest entry; verified the forbidden optimizer and ownership attributes remain absent.
- Extended independent core validation to admit only two locally proven helper bodies in shared-then-exclusive order, with matching scalar result flow and owner release before return. Direct Phase 24 return remains admitted.
- Corrected frame-drain classification for explicit foreign borrow/consume modes while retaining conservative handling for legacy missing contracts and unknown modes, and rejecting abandoned acquisitions.
- Executed model outcomes 65 and 66; the modeled 0x43 use failure returns its inherited typed error before either helper and leaves no live resources.
- Kept source-expressible ownership refusals source-attributed and locked the structured diagnostic wire fields without suggesting unproven repairs. Mutation, forwarding, retention, callback, nonlocal-exit, and wider-pointer serializer controls remain at the production-emission boundary.

## Task Commits

1. **Task 1: Emit the exclusive family as an actual pointer ABI** — `8fc40ec` (RED), `f073b2f` (GREEN).
2. **Task 2: Match interpreter composition and stable diagnostic boundaries** — `9f28abf` (initial regression test), `46bac97` (restored prior test content and added mode cases), `7dcec96` (implementation and integration tests).

Plan-scope amendments were checker-verified and committed as `6a8e7b6`, `499e7e9`, and `5ccab2d`.

## Files Created/Modified

- `internal/compiler/cgen/cgen.go`, `cgen_program.go` — checked exclusive pointer ABI derivation and lowering.
- `internal/compiler/cgen/cgen_pointer_successor_test.go`, `internal/compiler/native/phase25_pointer_successor_test.go` — ABI, manifest, wrong-result, and fail-before-serialization controls.
- `internal/compiler/corevalidate/corevalidate.go` — conservative frame drain and exact owner-transfer result-chain validation.
- `internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go` — explicit borrow/consume, legacy/unknown, and abandoned-acquisition boundaries.
- `internal/compiler/interp/interp.go`, `interp_pointer_successor_test.go` — modeled failure continuation and integrated success/error ordering.
- `internal/compiler/diagnostic/diagnostic_test.go`, `internal/compiler/session/session_pointer_successor_test.go` — stable wire schema and source-level ownership boundary controls.

## Decisions Made

- Explicit `Foreign` modes `borrow` and `consume` do not seed frame-owned acquisition tracking. Missing or unknown contracts continue to be treated conservatively.
- The owner-transfer peer recognizes only the exact two-call U64 flow through independent shared-borrow/copy/return and exclusive-borrow/copy/return helpers; it does not follow arbitrary call results.
- The current source language cannot reliably express several pointer-body mutations independently. Their negative controls remain at the production serializer seam, while source tests cover conflicts, returned escape, moved-from use, and discarded acquisition.
- NAT-12, NAT-13, and DX-15 remain pending for Plan 25-05’s phase-level acceptance and evidence.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Modeled foreign failure evaluated a successful-path call**
- **Found during:** Task 2
- **Issue:** After a borrowed foreign use selected its error edge, the interpreter tried to read the uninitialized scalar result while entering the shared helper.
- **Fix:** Skip successful-path value operations and calls during modeled failure traversal while still running release and the terminal typed failure.
- **Files modified:** `internal/compiler/interp/interp.go`
- **Verification:** The focused error-order test passed and observed neither helper before the typed failure.
- **Committed in:** `7dcec96`

**2. [Rule 1 - Bug] Regression test edit initially replaced pre-existing cases**
- **Found during:** Task 2
- **Issue:** The frame-drain test file already contained 271 lines; an initial add-file operation replaced that content.
- **Fix:** Restored the file from its parent revision, verified existing content was preserved, appended the explicit-mode cases, and committed the correction. The final diff keeps all prior tests.
- **Files modified:** `internal/compiler/corevalidate/corevalidate_callee_frame_drain_internal_test.go`
- **Verification:** `TestPeerCalleeFrameDrained` passed with explicit borrow/consume, legacy nil contract, unknown mode, and abandoned acquire cases.
- **Committed in:** `46bac97`

**Total deviations:** 2 auto-fixed issues (2 Rule 1 bugs).
**Impact on plan:** Both fixes were required to complete the integrated behavior and preserve existing regression coverage; scope remained within Plan 25-04.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^TestPhase25Exclusive(PointerABI|PointerManifest|WrongResult|Refusal)' ./internal/compiler/cgen ./internal/compiler/native` — passed (`cgen`, `native`).
- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^(TestPeerCalleeFrameDrained|TestPhase25(OwnerTransfer|InterpreterComposition|ErrorBeforeHelpers|StructuredDiagnostic|UnsupportedPointerShape|OwnershipDiagnosticBoundary))' ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/diagnostic ./internal/compiler/session` — passed all four packages.

These were the plan’s focused checks; no project-wide suites or CI were run.

## Self-Check: PASSED

The summary file exists, all listed task commits are present, and the measured plan commit count is 8 from the recorded baseline.

## Issues Encountered

The plan checker requested two narrow wording/scope amendments around the owner-transfer proof and source-versus-serializer refusal boundary. Both were checker-approved before closeout. The task commit count is measured from plan baseline `9c427a20dbe21202ed74fa1858c8bf72e8aef993`: 8 commits, including the task commits and plan-scope amendments.

## User Setup Required

None.

## Next Phase Readiness

Plan 25-04’s implementation and focused checks are complete. Plan 25-05 retains phase-level acceptance for NAT-12, NAT-13, and DX-15, including any remaining independent evidence.

---
*Phase: 25-separate-pointer-successors-and-integrated-utility*
*Completed: 2026-10-01*

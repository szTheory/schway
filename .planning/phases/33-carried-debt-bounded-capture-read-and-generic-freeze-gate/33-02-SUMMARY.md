---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
plan: 02
subsystem: testing
tags: [go, interpreter, native-parity, defects]
requires:
  - phase: 33-01
    provides: reopened M006 receipt boundary and bounded capture baseline
provides:
  - Fixture and file-route native/interpreter defect parity pins
  - Shared defect reason constants used by interpreter, peer validation, codegen, and session emission
  - Owner-scoped, function-scoped abort-like interpreter defect predicate with killed widening mutants
affects: [33-03, interpreter, execution-reasons, native-parity]
actuals:
  tokens: 8673
  tasks: 3
  commits: 4
plan_head_before: 3f101a3d7d6d16f592c3084bd3dcaf0110d7bc37
plan_head_after: d4e6282b2c964cf737da2234967407505cf415dc
tech-stack:
  added: []
  patterns:
    - Share literal defect reasons through the execution package without changing emitted C bytes
    - Resolve abort attribution in the final event function and require its exact owner resource
key-files:
  created:
    - internal/compiler/native/native_phase33_defect_parity_test.go
    - internal/compiler/execution/defect_reasons.go
    - internal/compiler/interp/interp_phase33_defect_test.go
    - .planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate/33-02-task3-red-evidence.json
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_phase321_defect_test.go
    - internal/compiler/executionpeer/executionpeer.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session.go
key-decisions:
  - "Keep abort-like acceptance limited to the exact acquired owner, the event function's operation, and that operation kind's exact reason."
  - "Keep defect reason literals centralized in execution while concatenating them into byte-identical C text."
patterns-established:
  - "Parity tests independently pin native exit/stderr and interpreter event/resource evidence on fixture and file routes."
requirements-completed: [QLT-14]
coverage:
  - id: D1
    description: Both native routes and interpreter routes agree on four pinned abort-like defect reasons and the single acquired owner.
    requirement: QLT-14
    verification:
      - kind: integration
        ref: internal/compiler/native/native_phase33_defect_parity_test.go#TestPhase33DefectParityFixtureRoute
        status: pass
      - kind: integration
        ref: internal/compiler/native/native_phase33_defect_parity_test.go#TestPhase33DefectParityFileRoute
        status: pass
    human_judgment: false
  - id: D2
    description: Interpreter defect acceptance checks exact owner, event function, operation kind, and reason bytes; both widening seams are mutation-tested.
    requirement: QLT-14
    verification:
      - kind: unit
        ref: internal/compiler/interp/interp_phase33_defect_test.go#TestPhase33AbortLikeNarrowed
        status: pass
      - kind: unit
        ref: internal/compiler/interp/interp_phase33_defect_test.go#TestPhase33AbortLikeWideningMutantKilled
        status: pass
    human_judgment: false
  - id: D3
    description: Shared reason constants preserve generated C identity and the session emitter inventory.
    requirement: QLT-14
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase272_identity_test.go#TestPhase272AcceptedCorpusByteIdentity
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase16_emitter_inventory_test.go#TestPhase16PublicEmitterConsumerInventory
        status: pass
    human_judgment: false
duration: 11min
completed: 2026-10-13
status: complete
---

# Phase 33 Plan 02: Abort-Like Defect Narrowing Summary

**Abort-like interpreter defects now require one exact live owner and a matching operation reason, while native parity and emitted C bytes stay pinned.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-10-13T16:53:52Z
- **Completed:** 2026-10-13T17:04:51Z
- **Tasks:** 3
- **Files modified:** 10, including the phase Windows ledger deviation entry

## Accomplishments

- Added literal parity pins for byte-index bounds, checked-add overflow, remainder by zero, and user defects on fixture-byte and file-byte routes.
- Moved the three fixed reason strings into `execution` and used them in interpreter, peer validation, code generation, and session emission. Byte-identity corpora and the public-emitter inventory passed.
- Narrowed `isAbortLikeDefect` to an exact one-owner live set, the function named by the final event, the operation kind, and exact output bytes. Ported the Phase 32.1 predicate tests and proved both widening seams are load-bearing.

## Task Commits

1. **Task 1: Pin abort-like defect parity on both routes** — `a7e46bef` (`test(33-02): pin abort-like defect parity on both routes`)
2. **Task 2: Shared defect-reason constants** — `b4897964` (`refactor(33-02): centralize defect reason strings`)
3. **Task 3 RED: Narrowing refusal matrix and evidence** — `48ab450d` (`test(33-02): add failing abort-like narrowing matrix`)
4. **Task 3 GREEN: Narrow predicate and kill widening mutants** — `d4e6282b` (`fix(33-02): narrow abort-like interpreter defects`)

## Files Created/Modified

- `internal/compiler/native/native_phase33_defect_parity_test.go` — literal native/interpreter parity on both acquisition routes.
- `internal/compiler/execution/defect_reasons.go` — shared exact defect reasons.
- `internal/compiler/interp/interp.go` — owner capture, exact reason matching, and final-function operation lookup.
- `internal/compiler/interp/interp_phase321_defect_test.go` — Phase 32.1 test rows updated for owner and function identity.
- `internal/compiler/interp/interp_phase33_defect_test.go` — narrow/refuse matrix and two widening mutants.
- `internal/compiler/executionpeer/executionpeer.go`, `internal/compiler/cgen/cgen_program.go`, and `internal/compiler/session/session.go` — shared constants; `session.go` retained its line count and inventory positions.
- `33-02-task3-red-evidence.json` — classifier-approved RED report for the extra-live-resource assertion.
- `.planning/WINDOWS.md` — recorded the acceptance-count mismatch described below.

## Decisions Made

- Defect event reasons are shared as execution constants, while C output is assembled from the same bytes without adding an escaping path.
- The live-resource exception remains narrower than native process exit behavior: any resource beyond the acquired owner is refused by the interpreter.

## RED and Mutant Evidence

- The initial narrowing test run failed on the intended five assertions: wrong reason, extra owner, missing owner, duplicate owner, and wrong event function. The four accepted cases and existing refused-terminal controls passed.
- The Go test text format was unsupported by the RED evidence checker. A local adapter rendered `go test -json` pass/fail/skip events as TAP; `gsd_run check tdd-red-evidence` returned `RED_EVIDENCE_OK` for `TestPhase33AbortLikeNarrowed/refuse/extra_live_resource`. The report is preserved at `33-02-task3-red-evidence.json`.
- After implementation, the narrowed matrix and the Phase 32.1 predicate tests passed. The mutant test widened each seam in turn, observed acceptance of its invalid row, restored the production seam, and observed refusal again.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run 'TestPhase321DefectParityWithLiveOwner|TestPhase33DefectParity' -count=1 -v` — passed; four subtests per new route and no skips.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase28PrunedEmitMatchesGolden|TestEmittedCExitLiteralsMatchTable|TestPhase16PublicEmitterConsumerInventory' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./internal/compiler/executionpeer ./internal/compiler/interp -run 'TestPhase33DefectParity|TestPhase321|Phase29' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run 'TestPhase321AbortLikeDefect|TestPhase33AbortLike' -count=1 -v` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native ./internal/compiler/session -run 'TestPhase33DefectParity|TestPhase321DefectParityWithLiveOwner|TestPhase29' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase28PrunedEmitMatchesGolden|TestEmittedCExitLiteralsMatchTable|TestPhase16PublicEmitterConsumerInventory|TestPhase32ReceiptLivingClaims' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run 'Phase29|Phase321' -count=1` — passed.
- Literal scan found exactly three non-test reason declarations, all in `defect_reasons.go`; cgen has nine constant references; `session.go` numstat is 3 additions and 3 removals. `git diff --check` passed.

## Deviations from Plan

### Plan acceptance mismatch

The plan's grep criterion expects two occurrences of the live-resource error prefix in `interp.go`. The required implementation correctly retains the same text in `RunWithForeignOutcomes` and both routed callers, so the source has three occurrences. No error text or caller was changed; the mismatch is recorded in `.planning/WINDOWS.md`.

**Total deviations:** 1 plan-acceptance mismatch. The implementation and all focused checks passed.

## Issues Encountered

- The first RED evidence submission used Go's plain text output, which the checker rejected as an unsupported format. The same target was rerun through a TAP adapter over Go JSON events and the classifier approved it before GREEN.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 33-02 is complete. Plan 33-03 can proceed with the narrowed interpreter predicate and centralized reason strings. No full local project suites, hosted receipts, or CI workflows were run.

---
*Phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate*
*Completed: 2026-10-13*

## Self-Check: PASSED

All created key files exist. Task commits `a7e46bef`, `b4897964`, `48ab450d`, and `d4e6282b` are ancestors of HEAD. The plan ledger measures four commits from `3f101a3d` through `d4e6282b`.

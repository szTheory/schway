---
phase: 24-ownership-transfer-through-calls-and-errors
plan: "02"
subsystem: compiler
tags: [ownership, typed-errors, call-transfer, core-validation, interpreter, c17]

# Dependency graph
requires:
  - phase: 24-01
    provides: Bounded local file-byte acquisition, release contracts, and the ordinary application path.
provides:
  - Repeated calls to one owner-returning helper with activation-qualified resource identity.
  - Typed-error cleanup derived across callee and caller frames.
  - Interpreter, emitter, and native application support for the bounded Phase 24 error witness.
affects: [24-03, corevalidate, originvalidate, pathoracle, interpreter, native-emission]

# Actuals (#2632)
actuals:
  tokens: 24483
  tasks: 3
  commits: 7

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Derive transferred owner obligations from acquisition and owning-return facts in independent peers.
    - Keep modeled typed-error outcomes separate from physical cleanup evidence.

key-files:
  created:
    - examples/phase24/error.schway
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session_phase24_error_test.go
    - .planning/LANGUAGE-MATURITY.md
    - testdata/phase16/public-emitter-consumers.json

key-decisions:
  - "Resource identity combines the static acquisition operation with the dynamic helper-call activation; repeated calls do not collapse into one owner."
  - "Phase 24 admits one bounded PathToken helper/use/error shape and refuses unsupported control transfers before native serialization."
  - "Independent release-order validation follows successful owning calls through the callee's returned acquisition and caller result place."
  - "Interpreter and generated event evidence do not claim physical destruction; Plan 24-03 supplies an independent observer."

patterns-established:
  - "A failed acquisition creates no owner obligation; only the successful call edge contributes to reverse-completion cleanup."
  - "Typed errors drain the exiting frame before propagating the original error to the caller frame."

requirements-completed: [RES-05, RES-06, OWN-10, OWN-11, OWN-12]
coverage:
  - id: D1
    description: Repeated owner-returning helper calls produce distinct activation-qualified acquisitions and preserve the typed error source path.
    requirement: OWN-11
    verification:
      - kind: integration
        ref: "CI 36798020384: Ubuntu and macOS go test ./...; includes TestPhase24ActivationPeerDerivesRepeatedAcquisitions"
        status: pass
    human_judgment: false
  - id: D2
    description: Independent core, origin, and path peers accept the valid cleanup proof and reject ownership/order mutations.
    requirement: RES-06
    verification:
      - kind: unit
        ref: "CI 36798020384: TestPhase24CleanupPeerRejectsOwnershipAndOrderMutations on Ubuntu and macOS"
        status: pass
    human_judgment: false
  - id: D3
    description: Interpreter, emitter, and ordinary native app agree on the propagated typed error and admitted cleanup path.
    requirement: OWN-12
    verification:
      - kind: integration
        ref: "CI 36798020384: TestPhase24RepeatedHelperTypedErrorDrainsEachActivation, TestPhase24EmitterError*, and TestPhase24NativeErrorApplication on Ubuntu and macOS"
        status: pass
    human_judgment: false

# Metrics
duration: 41min
completed: 2026-09-30
status: complete
---

# Phase 24 Plan 02: Repeated Owner Transfer and Typed Errors Summary

**Repeated helper acquisitions now survive owning returns as distinct resources, and a later typed error drains callee then caller obligations in reverse completion order.**

## Performance

- **Duration:** 41 minutes
- **Started:** 2026-09-30T20:19:39-04:00
- **Completed:** 2026-09-30T21:00:17-04:00
- **Tasks:** 3
- **Files modified:** 15

## Accomplishments

- Added `examples/phase24/error.schway`: `main` acquires twice through one helper, `probe` acquires a third time, and the 0x43 use path propagates `UseError.UnsupportedByte`.
- Extended source admission, interpreter execution, and C17 emission for the bounded call/error shape; cleanup preserves the original typed error and releases still-owned resources in reverse successful-acquisition completion order.
- Extended independent core, origin, and path proofs to validate activation-qualified repeated owners, caller/callee frame boundaries, and typed-error cleanup mutations.

## Task Commits

1. **Task 1: Admit repeated helper activations and typed-error source path** — `0a7c5655` (`feat(24-02): admit repeated helper typed errors`)
2. **Task 2: Independently derive activation and cleanup across call/error paths** — `a491c939` (`feat(24-02): validate repeated acquisition error cleanup`)
3. **Task 3: Execute the post-acquisition typed error through the ordinary app** — `b06d42f7` (`feat(24-02): execute propagated typed errors with cleanup`)
4. **Task 3 fix: Keep the emitter test helper package-local** — `778dd272`
5. **Cross-task fix: Reconcile proof identifiers and tracked fixture metadata** — `72807362`
6. **Task 2 fix: Derive cleanup across owning calls** — `a4b522c5`
7. **Task 2 fix: Accept typed-call edge terminators** — `e8e0d691`

## Files Created/Modified

- `examples/phase24/error.schway` — repeated helper acquisition and propagated typed-error witness.
- `internal/compiler/check/check.go`, `check_test.go` — bounded source admission, successful-completion cleanup ledger, and source tests.
- `internal/compiler/core/core.go` — typed call/error edge facts.
- `internal/compiler/corevalidate/corevalidate.go`, `originvalidate/originvalidate.go`, `pathoracle/pathoracle.go` — independent activation, transfer, edge, and cleanup validation.
- `internal/compiler/session/session_phase24_error_test.go` — peer agreement and ownership/error mutation matrix.
- `internal/compiler/interp/interp.go`, `interp_test.go` — dynamic activation tracking and original typed-error propagation.
- `internal/compiler/cgen/cgen_program.go`, `cgen_program_test.go` — fail-closed Phase 24 validation and bounded typed-error C lowering.
- `internal/compiler/native/native_app_test.go` — ordinary typed-error native application execution.
- `.planning/LANGUAGE-MATURITY.md` — refreshed corpus totals from hosted evidence: 147 programs and 4,870 lines.
- `testdata/phase16/public-emitter-consumers.json` — shifted Phase 16 source-line pins after the test import change.

## Decisions Made

- Use `(static acquisition operation, dynamic call activation)` as the resource identity through repeated owning helper returns.
- Keep Phase 24 source admission and native emission bounded to the named fixture shape; unsupported shapes remain refused.
- Have core release-order validation derive returned resources from the callee's own acquisition and return, then check the caller's result place on success edges.
- Preserve evidence boundaries: model output and generated release events are not physical-cleanup proof; that proof is Plan 24-03's responsibility.

## Deviations from Plan

### Auto-fixed Issues

1. **Test visibility:** an external emitter test referenced a package-private helper. The test now has a package-local helper, without widening production API surface (`778dd272`).
2. **Proof and inventory alignment:** the hosted run exposed a custom linear-body ID mismatch, changed line pins in the Phase 16 emitter inventory, and stale corpus totals. These were reconciled against the generated proof IDs and hosted corpus evidence (`72807362`).
3. **Owning-call release derivation:** the first full hosted run showed that core release-order validation followed direct foreign acquisitions but omitted an owner returned by a helper call. It now derives the call's owner event from the callee acquisition and returned place (`a4b522c5`).
4. **Typed-call CFG terminator:** the next hosted run showed that a fallible `OpCall` with success/error successors was still treated as a function-terminal block. The validator now recognizes the declared call edges (`e8e0d691`).

**Total deviations:** 4 auto-fixed. All were required for correctness or repository gates; the plan scope did not expand.

## Issues Encountered

- Hosted CI run `36795775442` caught a test-only helper visibility error; `36795959701` also caught proof-ID and tracked-inventory drift. Both were corrected before the later clean runs.
- Hosted CI run `36796619535` caught the missing owning-call cleanup derivation. Run `36797648335` then caught the call-edge terminator omission after the first fix. Both findings were fixed and rerun through hosted CI.
- Hosted CI run `36798020384` passed on the final commit `e8e0d69112e15cca2a97cd7a17e3f2b89f17822a`: Ubuntu and macOS full checks, race suites, and current-evidence aggregates all succeeded. The validation-corpus receipt was skipped because this was a manual dispatch; no corpus receipt is claimed from that skip.
- No local project tests or scripts were run; the plan requires hosted-only automated validation.

## User Setup Required

None.

## Next Phase Readiness

Plan 24-02 is complete and Plan 24-03 is ready. The remaining Phase 24 proof is an independent native observer for physical allocation/use/free behavior, five reached negative controls, model-only replay flags, and a focused two-host Phase 24 aggregate. Physical cleanup is not claimed by this summary.

---
*Phase: 24-ownership-transfer-through-calls-and-errors*
*Completed: 2026-09-30*

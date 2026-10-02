---
phase: 24-ownership-transfer-through-calls-and-errors
plan: "01"
subsystem: compiler/resource-lifecycle
tags: [ownership, call-return, C17, interpreter, hosted-CI]

# Dependency graph
requires:
  - phase: "23"
    provides: ["live file-byte acquisition and physical cleanup evidence"]
provides:
  - "Bounded helper-acquired owner transfer through owning return and caller use"
  - "Independent acquisition-derived transfer validation in three peers"
  - "Interpreter and native application witness that releases after caller use"
affects: ["24-02 repeated activations and typed-error cleanup", "24-03 physical observer and Phase 24 aggregate"]

# Actuals (#2632)
actuals:
  tokens: 20311
  tasks: 3
  commits: 10

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Resource obligations are derived from acquisition operations and carried across owning returns."
    - "Modelled release events remain separate from physical cleanup and host-I/O claims."
key-files:
  created:
    - examples/phase24/transfer.schway
    - internal/compiler/session/session_phase24_transfer_test.go
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/core/core.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/native/native_app_test.go
    - .planning/LANGUAGE-MATURITY.md
    - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md
    - .github/workflows/ci.yml
key-decisions:
  - "Admit only the bounded file-byte helper transfer shape, with no owning result from process entry."
  - "Keep foreign operation contracts local to each operation and retain emitProgram as the sole C serializer."
  - "Run project validation only on hosted Linux and macOS CI; allow 20 minutes for the full hosted race suite."

requirements-completed: [RES-05, OWN-10]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: "A helper returns a live file-byte owner to its caller; invalid copy, stale use and entry-result escape are rejected with source attribution."
    requirement: RES-05
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestPhase24SourceTransfer and TestPhase24SourceRefusal; full hosted go test ./..."
        status: pass
    human_judgment: false
  - id: D2
    description: "Three independent peers derive the transfer obligation from acquisition and reject missing, duplicate, mismatched and escaping releases."
    requirement: OWN-10
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase24_transfer_test.go#TestPhase24TransferPeer; full hosted go test ./..."
        status: pass
    human_judgment: false
  - id: D3
    description: "The caller-selected 0x41 and 0x42 files produce 65 and 66 through the ordinary app, with borrowed use before one paired release."
    requirement: RES-05
    verification:
      - kind: integration
        ref: "internal/compiler/cgen/cgen_program_test.go#TestPhase24EmitterTransfer and internal/compiler/native/native_app_test.go#TestPhase24PositiveTransferNativeApplication; full hosted go test ./..."
        status: pass
    human_judgment: false

# Metrics
duration: 200min
completed: 2026-09-30
status: complete
---

# Phase 24 Plan 01: Helper Owner Transfer Summary

**A helper-returned file-byte owner stays live through caller use and releases once after returning 65 or 66.**

## Performance

- **Duration:** 200 min
- **Started:** 2026-09-30T18:34:03Z
- **Completed:** 2026-09-30T21:53:40Z
- **Tasks:** 3
- **Files modified:** 17 implementation, CI and validation files

## Accomplishments

- Added an ordinary Schway witness that acquires a file-byte owner in a helper, returns it to `main`, borrows it for use, and releases it after the use.
- Added source-attributed refusal for owner copy, stale use and an owning process-entry result.
- Extended `corevalidate`, `originvalidate` and `pathoracle` to derive and independently check transfer obligations from acquisition operations.
- Carried the same owner through interpreter call/return and the bounded C emitter; the native application returns the independently fixed results 65 and 66 for caller-selected 0x41 and 0x42 files.
- Kept model release evidence distinct from physical cleanup and host-I/O evidence reserved for the later observer plan.

## Task Commits

1. **T-24-01: Admit a helper-acquired owner returned to the caller** — `f55d791b`, `d5a5e80c`, `b0eafa4c`.
2. **T-24-02: Independently validate call/return ownership transfer** — `97360862`, `926ce878`, `101a1722`.
3. **T-24-03: Run and emit the positive transferred-owner application** — `ec113f94`, `9962d911`, `41f9908f`.
4. **Hosted race-suite timeout adjustment** — `3b6a2da4`.

**Plan metadata:** to be committed with this summary.

## Files Created/Modified

- `examples/phase24/transfer.schway` — positive helper-acquire, owning-return and caller-use witness.
- `internal/compiler/check/` — bounded transfer admission and source refusal cases.
- `internal/compiler/core/` — typed transfer facts attached to operations.
- `internal/compiler/corevalidate/`, `internal/compiler/originvalidate/`, `internal/compiler/pathoracle/` — independent acquisition-derived peer validation.
- `internal/compiler/session/session_phase24_transfer_test.go` — valid-transfer and mutation evidence across peers.
- `internal/compiler/interp/interp.go` — owner identity survives helper return and caller borrow/release.
- `internal/compiler/cgen/` and `internal/compiler/native/native_app_test.go` — fail-closed C admission and ordinary 65/66 app witness.
- `.planning/LANGUAGE-MATURITY.md` — refreshed corpus and guard census from current source witnesses.
- `.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md` — hosted result and host-duration receipt.
- `.github/workflows/ci.yml` — explicit 20-minute race-test timeout for the full hosted suite.

## Decisions Made

- Kept the admitted behavior to the exact bounded file-byte helper/caller shape; process entry still returns no owner.
- Preserved each foreign operation's own ABI and release contract and extended the existing `emitProgram` serializer only.
- Recorded a 20-minute Go test timeout after the macOS race suite exceeded Go's default 10-minute package timeout. The complete retry passed.
- No local project test or build commands were run. The outgoing public diff had zero added-line PII-pattern matches and zero Gitleaks findings.

## Deviations from Plan

### Auto-fixed Issues

1. The emitter initially compared function-local type identifiers across helper and caller. Admission now checks each function's type constructors in its own scope.
2. Model execution needed to follow the foreign helper's declared success/error CFG edges and carry its activation and owner identity through return.
3. Entry-owner admission was tightened to refuse `main`; otherwise an owning helper shape could accidentally apply to the process boundary.
4. Hosted checks exposed stale maturity counts and emitter-test assumptions about generated function order; the census and structural assertions were updated to match current source and emitter output.
5. The first complete macOS race run reached the default 10-minute Go package timeout. CI now uses `-timeout=20m` for that same full suite; the subsequent Linux/macOS run passed.

**Total deviations:** 5 auto-fixed issues. No scope expansion; each correction preserved the bounded transfer contract or made its hosted evidence accurate.

## Issues Encountered

Hosted CI identified compiler errors and mismatches in function-local type facts, foreign CFG routing, test assumptions about generated order and release-event identity. Those were corrected in the commits above. Final full checks, race checks and both existing evidence aggregates passed on Linux and macOS in [run 36780855499](https://github.com/szTheory/schway/actions/runs/36780855499).

## User Setup Required

None.

## Next Phase Readiness

Ready for Plan 24-02. Single-activation owner transfer is implemented and covered; repeated helper activations, typed-error cleanup and reverse-order discharge remain unproven and are the next planned work.

---
*Phase: 24-ownership-transfer-through-calls-and-errors*
*Completed: 2026-09-30*

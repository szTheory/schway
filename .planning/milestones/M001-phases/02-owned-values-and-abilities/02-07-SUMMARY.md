---
phase: 02-owned-values-and-abilities
plan: "07"
subsystem: compiler
tags: [go, c17, native-execution, runtime-causality, mutation-control, tdd]

requires:
  - phase: 02-owned-values-and-abilities
    plan: "06"
    provides: Owned evidence and the bounded Phase 1+2 verification gate
provides:
  - Runtime-derived owned C outcome and ordered event documents with counted output
  - Real O0/O3 emitted-C mutation proving semantic mismatch exit 4
  - Eight exact nonzero-work Phase 2 controls including backend runtime causality
affects: [phase-2-reverification, owned-native-semantics, phase-5-adversarial-evidence]

actuals:
  tokens: 7894
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns: [operation-site runtime event recording, returned-place serialization, exact-one emitted-C mutation, fail-closed causal control]

key-files:
  created:
    - internal/compiler/cgen/cgen_test.go
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/testsupport/cli_test.go
    - testdata/phase2/owned_transfer.golden.c
    - testdata/phase2/evidence.golden.json
    - scripts/verify-phase2.sh

key-decisions:
  - "Generated linear C records event occurrence and order at executed operation sites, then serializes the terminal value from the returned C place."
  - "All generated-C document writes share one 64 KiB counted writer; JSON strings, Byte decimals, and fixed-buffer hex values are centralized helpers."
  - "The backend causality control replaces exactly one canonical owned-transfer value site and must execute through real Clang at both O0 and O3 before mismatch evidence is admitted."

patterns-established:
  - "Causal native facts: checked metadata may be static, but event occurrence/order and outcome values come from executed C state."
  - "Backend mutation controls: fail closed on a missing or duplicate mutation site and distinguish operational control failure from semantic mismatch."

requirements-progressed: [OWN-01]

coverage:
  - id: D1
    description: "Owned Buffer and Byte native executions serialize the returned C place and operation-site runtime events through one bounded output layer without Go-side linear replay."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/cgen/cgen_test.go#TestLinearCSerializesRuntimeState"
        status: pass
      - kind: e2e
        ref: "internal/compiler/session/session_test.go#TestOwnedTransferInterpreterNative"
        status: pass
    human_judgment: false
  - id: D2
    description: "A real exact-one emitted-C value mutation runs at O0 and O3, disagrees with the interpreter, and is required as control:backend.runtime_causality with nonzero work."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestOwnedBackendMutationIsMismatch,TestVerifyPhase2ControlsAndWork"
        status: pass
      - kind: e2e
        ref: "scripts/verify-phase2.sh"
        status: pass
    human_judgment: false

duration: 9min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 07: Native Runtime Causality Gap Closure Summary

**Executed linear C state now causes the owned execution document, and a real O0/O3 backend mutation proves drift becomes semantic mismatch exit 4.**

## Performance

- **Duration:** 9 min
- **Started:** 2026-09-03T22:06:12Z
- **Completed:** 2026-09-03T22:14:53Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Removed `linearExecution` as an independent Go replay and made generated C record copy/move/return events in executed operation order before serializing the returned runtime place.
- Added a centralized generated-C output layer for counted writes, JSON string escaping, Byte decimal formatting, and fixed-buffer hex formatting with a deterministic 64 KiB ceiling.
- Added a fail-closed exact-one C mutation seam around the real shell-free native runner; both O0 and O3 corrupt the returned Buffer and deterministically produce `native.engine_mismatch`, semantic-mismatch status, and exit 4.
- Extended Phase 2 verification from seven to eight exact controls while preserving all prior controls, strict stream boundaries, Phase 1 bytes, and the expected coordinated-lie escape.

## Task Commits

Each TDD task has a RED contract followed by its GREEN implementation:

1. **Task 02-07-01 RED: causal linear C output contract** - `f2fe11e` (test)
2. **Task 02-07-01 GREEN: runtime-derived owned execution document** - `75130d7` (feat)
3. **Task 02-07-02 RED: real backend causality control contract** - `4981acf` (test)
4. **Task 02-07-02 GREEN: O0/O3 emitted-C mutation control** - `0cf3565` (feat)

## Files Created/Modified

- `internal/compiler/cgen/cgen.go` - Operation-site event recording, runtime terminal serialization, bounded writers, and removal of Go-side linear replay.
- `internal/compiler/cgen/cgen_test.go` - Buffer move and Byte copy contracts for causal runtime serialization and absence of a constant execution document.
- `internal/compiler/session/session.go` - Exact-one backend mutation runner, O0/O3 exercise tracking, mismatch admission, and the eighth required control.
- `internal/compiler/session/session_test.go` - Runtime-derived owned C assertions, real mutation mismatch/exit taxonomy, and exact control-set coverage.
- `internal/compiler/testsupport/cli_test.go` - Shipped CLI and script requirements for backend runtime causality.
- `testdata/phase2/owned_transfer.golden.c` - Reviewable runtime-derived generated C17.
- `testdata/phase2/evidence.golden.json` - Causally updated owned C digest and evidence identity only.
- `scripts/verify-phase2.sh` - Exact discovery and explicit enforcement of the backend runtime-causality control.

## Decisions Made

- Runtime event records store only checked literal metadata; their existence and order are determined by the C statements that actually execute.
- Output overflow or event-capacity failure exits the child with stable nonzero result 74 before a document can exceed the bound.
- The canonical owned-transfer assignment is the deliberate mutation seam; absence or duplication is `native.backend_control_invalid`, never a passing mutation.
- Phase 1 match emission and evidence remain unchanged; only causally affected Phase 2 C/evidence goldens were regenerated.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The first generated outcome prefix omitted the JSON value's opening quote. The strict native decoder rejected it immediately; the emitted prefix was corrected before the GREEN verification and commits.

## Verification Evidence

- Task 1 exact selector passed `TestLinearCSerializesRuntimeState`, `TestOwnedTransferInterpreterNative`, and `TestPhase1EvidenceGoldenUnchanged` with fail-on-zero discovery.
- Task 2 exact selector passed `TestOwnedBackendMutationIsMismatch`, `TestVerifyPhase2ControlsAndWork`, `TestVerifyPhase2CLI`, and `TestPhase1EvidenceGoldenUnchanged` with fail-on-zero discovery.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...`, `go test -race ./...`, and `go vet ./...` all exited 0.
- `scripts/verify-phase1.sh` passed five lanes with 18 recomputed work units and unchanged Phase 1 evidence.
- `scripts/verify-phase2.sh` exited 0 with all eight controls present; `lane:owned-backend-causality` passed with work 2 and total Phase 2 work 51.
- Full-gate warm observations (20 each): format p50/p95 `50,750/60,666 ns`; check `69,541/80,250 ns`; interpreter `199,459/304,208 ns`; native `368,623,083/394,563,208 ns`; full verify `787,742,000/834,836,834 ns`.

## Known Stubs

None.

## Threat Flags

None. The runtime document path and backend mutation seam implement planned mitigations T-02-01, T-02-03, T-02-06, T-02-07, and T-02-08 without adding a new trust boundary.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The causal OWN-01 gap is closed and ready for independent Phase 2 re-verification.
- `OWN-01` remains progressed, not completed; `REQUIREMENTS.md` is intentionally unchanged.
- Phase 02 remains in `verifying` status until the independent verifier accepts the repaired native-causality truth.

## Self-Check: PASSED

All eight changed implementation/test/golden/script files, this summary, STATE, ROADMAP, and all four RED/GREEN commits were found. The complete Phase 1 and Phase 2 gates passed; `requirements-progressed: [OWN-01]` is recorded while `REQUIREMENTS.md` and `.planning/milestone.lock` remain untouched.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*

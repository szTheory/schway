---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
plan: 04
subsystem: testing
tags: [go, posix-sh, receipts, freeze-gate, migration]
requires:
  - phase: 33-03
    provides: Generic freeze and receipt gates with the M006 manifest
provides:
  - Generic freeze and receipt tests retain the Phase 32 refusal cases under generic names
  - Literal differential goldens preserve legacy gate outcomes after script retirement
  - Phase 32 producer, route audit, and case table tests read the M006 manifest
  - Tracked-file control prevents retired gate callers from returning
affects: [phase-33, phase-38, evidence-gates, receipt-production]
actuals:
  tokens: 28518
  tasks: 3
  commits: 3
plan_head_before: 446b705d752a0ed361b47c7dc72554a3bed24d20
plan_head_after: a414dea1af6f171b00791d77bc382e0ee31cf931
tech-stack:
  added: []
  patterns:
    - Freeze retired-script behavior as byte-level exit/stdout/stderr literals
    - Use one milestone manifest as the producer and test source for programs and case rows
key-files:
  created:
    - internal/compiler/session/session_gate_retire_test.go
  modified:
    - internal/compiler/session/session_freeze_gate_test.go
    - internal/compiler/session/session_receipt_gate_test.go
    - internal/compiler/session/session_gate_differential_test.go
    - scripts/verify-phase32.sh
    - cmd/schway/main_phase32_routes_test.go
    - .planning/milestones/M006-phases/32.1-address-tech-debt-cli-exit-classes-and-module-aware-routes/32.1-VALIDATION.md
  deleted:
    - internal/compiler/session/session_phase32_freeze_test.go
    - internal/compiler/session/session_phase32_gate_test.go
    - scripts/check-phase32-freeze.sh
    - scripts/check-phase32-receipts.sh
key-decisions:
  - Keep M006 program and case data in the manifest shared by the producer and its consumer tests.
  - Normalize generated fixture commit IDs as @SHA@ in pass-line goldens, alongside temporary paths as @TMP@.
requirements-completed: [EVD-16]
coverage:
  - id: D1
    description: Freeze and receipt refusal cases run under generic test names, and all 45 old/new differential rows matched before the retired scripts were removed.
    requirement: EVD-16
    verification:
      - kind: integration
        ref: internal/compiler/session/session_freeze_gate_test.go#TestFreezeGate
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_receipt_gate_test.go#TestReceiptGate
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_gate_differential_test.go#TestFreezeGateDifferential
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_gate_differential_test.go#TestReceiptGateDifferential
        status: pass
    human_judgment: false
  - id: D2
    description: Producer and route audit use M006.manifest, manifest bytes are pinned, archived logs contain every case row, and no tracked caller reintroduces retired scripts.
    requirement: EVD-16
    verification:
      - kind: integration
        ref: internal/compiler/session/session_receipt_gate_test.go#TestPhase32ProducerCaseTable
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_gate_retire_test.go#TestNoCheckPhase32Caller
        status: pass
      - kind: integration
        ref: cmd/schway/main_phase32_routes_test.go#TestPhase32PublicRouteAudit
        status: pass
    human_judgment: false
duration: 12min
completed: 2026-10-13
status: complete
---

# Phase 33 Plan 04: Generic Gate Migration Summary

**The Phase 32 producer and tests now share one M006 manifest, while literal goldens preserve every legacy gate result after the old scripts are removed.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-10-13T17:28:06Z
- **Completed:** 2026-10-13T17:40:09Z
- **Tasks:** 3
- **Files modified:** 11 (including four deletions)
- **Budget:** 11-file plan scope; estimated 90,000 tokens versus 28,518 actual tokens (114,072 diff bytes / 4).

## Accomplishments

- Ported the freeze and receipt test cases under `TestFreezeGate` and `TestReceiptGate`; moved `TestPhase32ProducerCaseTable` into the receipt test file and amended the archived validation command on its existing line with a dated note.
- Captured and checked 45 old-versus-generic script rows before deletion. The differential tests now compare literal exit, stdout, and stderr goldens, normalizing fixture paths as `@TMP@` and generated commit IDs as `@SHA@`.
- Deleted the two phase-specific gate scripts and their old test files. Added `TestNoCheckPhase32Caller`, including an armed in-memory control.
- Made `verify-phase32.sh` read its program list and case rows from `M006.manifest`, and its digest path set from `codedigest.sh --paths`. The producer and route-audit tests now use that same manifest; the test pins its SHA-256 and checks each case row in both archived logs.

## Task Commits

1. **Task 1: Cut the gate test suite over to generic scripts** — `d07d4c2b` (`test(33-04): port phase32 gate cases to generic gates`)
2. **Task 2: Freeze differential goldens and retire old scripts** — `1dda5a0b` (`test(33-04): freeze generic gate differential goldens`)
3. **Task 3: Repoint producer and table-coupled tests to the manifest** — `a414dea1` (`feat(33-04): read Phase 32 producer inputs from manifest`)

## Verification

- Focused session tests passed for freeze/receipt behavior, differential goldens, archived M006 receipt controls, no-caller scan, producer table, codedigest, living claims, CI host matrix, and hosted CI environment controls.
- Session groundedness, validation lifecycle, static test index, and emitter inventory meta-tests passed.
- `TestPhase32PublicRouteAudit` passed.
- `sh -n scripts/verify-phase32.sh` passed. Static checks found no `check-phase32` references outside `.planning/`, no inline producer case/program table, and no CI workflow change. The producer retains the same six `run_control` occurrences as its baseline.
- Gate subtest output contained 25 freeze subtests and 29 receipt subtests, including the preserved ported cases.
- No full project suite, hosted receipt, CI workflow, or archived receipt/log edit was performed.

## Deviations from Plan

### Auto-adjusted test normalization

The pass rows print fixture-generated Git object IDs. Those IDs changed when the old script stopped being part of the fixture repository. The goldens normalize these generated IDs to `@SHA@` in addition to the planned `@TMP@` path normalization so the remaining output stays byte-for-byte stable and reviewable.

## Decisions Made

- The M006 manifest is the shared source for producer programs and case rows; the route audit and producer-table test read it directly.
- Literal differential results remain in catalog order and keep each refusal reason visible in the expected stderr.

## Self-Check: PASSED

All created key files exist, all three task commits are ancestors of HEAD, and the plan ledger measures three commits from `446b705d752a0ed361b47c7dc72554a3bed24d20` through `a414dea1af6f171b00791d77bc382e0ee31cf931`.

---
*Phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate*
*Completed: 2026-10-13*

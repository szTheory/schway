---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 08
subsystem: scripts/receipts
tags: [evd-15, receipts, codedigest, two-host-gate, case-table]
requires: [32-02, 32-05, 32-06, 32-07]
provides:
  - scripts/verify-phase32.sh one-host producer (39-row shared case table, seven programs, four controls)
  - scripts/check-phase32-receipts.sh two-host consistency gate bound to the codedigest
  - TestPhase32ReceiptGate (21 subtests) and TestPhase32ProducerCaseTable
  - phase32_programs= coupling between producer, route audit and phase32ReceiptPrograms
affects: [32-09, 32-11]
key-files:
  created:
    - scripts/verify-phase32.sh
    - scripts/check-phase32-receipts.sh
    - internal/compiler/session/session_phase32_gate_test.go
  modified:
    - cmd/schway/main_phase32_routes_test.go
key-decisions:
  - "The gate reads the case table and phase32_programs from the producer script, so producer and gate share one table; the Go test keeps an independent literal list and proves agreement"
  - "Two logs at different digests report log_digest_differs before any per-log digest_mismatch; a missing log is host_missing"
  - "Case row grammar carries program=/status= so the 32-05 phase32ReceiptProblems validator reads the same logs"
requirements-completed: [EVD-15]
status: complete
commits: 2
plan_head_before: 774c0d4c995e27194ea08710a57f39059e11caa3
plan_head_after: ecdf2526e8f73a06066fa770ed6f912c538604c6
actuals:
  tokens: 10300
  tasks: 2
  commits: 2
coverage:
  - id: D-08b / D-09
    description: "The two-host gate accepts an independent literal macOS + Linux pair and rejects a mutated digest, a code edit, a dirty tree, a dropped host, a duplicate host, dropped program and OOB rows, a swapped programs= list, an unknown row, altered cases and controls, a misplaced result and differing log digests; a .planning-only HEAD move keeps a receipt current"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_gate_test.go#TestPhase32ReceiptGate"
        status: pass
    human_judgment: false
  - id: EVD-15 programs
    description: "The producer case table holds 39 rows over seven programs, each agreeing with checksum.expected.json, the ChecksumOracleRows adler_low values, Phase29FixtureExpected, the BYTE-02 stderr literal and the Phase 27 fixtures"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_gate_test.go#TestPhase32ProducerCaseTable"
        status: pass
    human_judgment: false
  - id: D-10b
    description: "The route audit's covered program set equals the producer's phase32_programs= line, and a producer line with fizzbuzz removed is reported"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_routes_test.go#TestPhase32PublicRouteAudit"
        status: pass
    human_judgment: false
duration: 4min
completed: 2026-10-12
---

# Phase 32 Plan 08: Receipt producer and two-host gate Summary

`scripts/verify-phase32.sh` records, at one codedigest, 39 case rows over seven programs plus four non-skippable controls, and `scripts/check-phase32-receipts.sh` accepts only a consistent macOS + Linux pair at that digest.

## Accomplishments

- Tracer: one checksum row travelled from a real local producer run through the shared table into the gate and the Go gate test (commit 9083c997).
- Producer: identity (host, arch, go, clang_target, headsha, codedigest, tree over the digest paths), toolchain, 39 case rows, controls guide/route-audit/codedigest/receipt-rules (fail on SKIP or no tests), start/end tree comparison (`tree_moved`), and a result row carrying `programs=`. The D-09a non-claim (default build only; optimization matrix bound by the hosted go test lanes) is in the header.
- Gate: named rejections for digest, tree, host, program, case, control, unknown-row, ordering and failing-status; header keeps "This gate checks consistency, not authenticity".
- Tests: `TestPhase32ReceiptGate` (21 subtests, independent literal log builder), `TestPhase32ProducerCaseTable` (8 subtests), and the `programs-line-drift` armed subtest in `TestPhase32PublicRouteAudit`.
- phase27 scripts and `session_phase27_host_test.go` are byte-untouched; `TestPhase27HostReceiptGate`, `TestPhase27CIHostMatrix` and `TestPhase27GuideContract` pass.

## Local producer observation (not a receipt)

Run on the planning host (macOS arm64, clean digest-path tree) at HEAD ecdf2526; 39 case rows passed in about 10 s. Last row:

`phase32 host result host=macOS codedigest=715b4137f99d90aaa5479bbda757d763afb4b199b404770f1f3644cb78b137dc tree=clean cases=39 controls=4 programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz status=pass`

Feeding that log and a host-renamed copy to the gate printed `status=pass hosts=macOS,Linux ... cases=39`. That second log was not produced on Linux, so this is a gate smoke check only.

## Deviations from Plan

None - plan executed exactly as written. Two small choices inside the plan's latitude: the gate checks `log_digest_differs` before per-log digest checks (so the two named reasons are distinguishable), and every program binary is built to `$verify_tmp/<program>` (the plan's `adler32-module` output name for adler32_cli is cosmetic).

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- Files present: scripts/verify-phase32.sh, scripts/check-phase32-receipts.sh, internal/compiler/session/session_phase32_gate_test.go, cmd/schway/main_phase32_routes_test.go.
- Commits present: 9083c997, ecdf2526.
- Acceptance: 39 table rows, 11 adler32_fixture rows, 2 byte_index_oob rows, 5 run_control occurrences, 21 TestPhase32ReceiptGate subtests passing, 0 phase27-prefixed rows, phase27 files unchanged.
- Verification: session meta set, TestValidationRowGradesAreEarnedOverArchivedCorpus and `go test ./cmd/schway` green; full producer run printed exactly 39 pass lines.

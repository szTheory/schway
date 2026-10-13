---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 11
subsystem: evidence/receipts
tags: [evd-15, dx-19, receipts, m006-closure]
requires: ["32-09", "32-10"]
provides:
  - 32-RECEIPT.json bound to codedigest c1b4e85d (run 38315799626)
  - receipts/receipt-macos-latest.log and receipt-ubuntu-latest.log
  - bound marker lines in LANGUAGE-MATURITY.md, PRODUCT-ROADMAP.md, STATE.md
affects: []
key-files:
  created:
    - .planning/phases/32-clean-checkout-guide-and-final-receipts/32-RECEIPT.json
    - .planning/phases/32-clean-checkout-guide-and-final-receipts/receipts/receipt-macos-latest.log
    - .planning/phases/32-clean-checkout-guide-and-final-receipts/receipts/receipt-ubuntu-latest.log
  modified:
    - .planning/LANGUAGE-MATURITY.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/STATE.md
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md
    - .planning/phases/32-clean-checkout-guide-and-final-receipts/32-VALIDATION.md
key-decisions:
  - "Ingest is one .planning-only commit, so the codedigest does not move and the receipt stays current"
requirements-completed: [EVD-15, DX-19]
status: complete
commits: 1
plan_head_before: 410ad481e486f03ad2f46043460e7abf69b2c295
plan_head_after: f1c8216cbf11f777d423fb17c8932176cde90f18
actuals:
  tokens: 12000
  tasks: 2
  commits: 1
coverage:
  - id: SC3-EVD-15-receipt
    description: "The committed hosted macOS and Linux logs and 32-RECEIPT.json are accepted by the receipt gate at the frozen digest, and the living claims are enforced against the real receipt"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_receipt_test.go#TestPhase32ReceiptLivingClaims"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase32_gate_test.go#TestPhase32ReceiptGate"
        status: pass
    human_judgment: false
  - id: D-10a-bound-markers
    description: "The three living docs carry the bound marker at the receipt digest; a later code change must flip them to reopened"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_receipt_test.go#TestPhase32ReceiptLivingClaims"
        status: pass
    human_judgment: false
  - id: D-08c-freeze-binding
    description: "The run headSha and the post-ingest HEAD are bound to the freeze; the .planning-only HEAD move is accepted and a synthetic sha is rejected"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_freeze_test.go#TestPhase32FreezeGate"
        status: pass
    human_judgment: false
  - id: D-11-coverage-groundedness
    description: "Every Phase 32 SUMMARY coverage block resolves"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "internal/compiler/session/summary_coverage_test.go#TestSummaryCoverageGroundedness"
        status: pass
    human_judgment: false
duration: 20min
completed: 2026-10-12
---

# Phase 32 Plan 11: Receipt ingest and M006 closure Summary

Ingested the hosted macOS and Linux receipts for the frozen code digest, bound the living docs to them, and closed DX-19, EVD-15 and Phase 32.

## Accomplishments

- Hosted run 38315799626 (workflow_dispatch, headSha 6c304271) concluded success; checks and evidence aggregate green on ubuntu-latest and macos-latest, input-gated receipt jobs skipped as expected.
- Freeze gate passed for the run headSha and for HEAD; the armed control with a synthetic sha was rejected with `reason=sha_mismatch`.
- Both logs were downloaded into fresh scratch dirs, gated with `PHASE32_EXPECTED_CODEDIGEST` at c1b4e85d... (status=pass, 39 cases, 7 programs, 4 controls per host), and committed byte-for-byte.
- `32-RECEIPT.json` (schema `schway.m006-receipt/1`, state current) records both hosts, seven programs passing on both, and the optimization-matrix non-claim verbatim.
- Marker line added to LANGUAGE-MATURITY.md, PRODUCT-ROADMAP.md and STATE.md with an explanatory sentence.
- REQUIREMENTS.md (DX-19, EVD-15 complete), ROADMAP.md (Phase 32 11/11 Complete 2026-10-12) and 32-VALIDATION.md (validated, nyquist_compliant true) updated.
- Ingest commit f1c8216c touches only `.planning/`; after it the freeze gate passes, the codedigest is unchanged, and the full `internal/compiler/session` and `cmd/schway` packages pass.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Next

`/gsd-verify-work 32`, then `/gsd-complete-milestone`.

## Self-Check: PASSED

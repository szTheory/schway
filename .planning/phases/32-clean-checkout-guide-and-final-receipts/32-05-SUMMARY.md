---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 05
subsystem: receipts
tags: [evd-15, codedigest, receipt-validator, living-claims]
requires: ["32-03"]
provides:
  - scripts/codedigest.sh (sh code digest over HEAD, one line of 64 hex)
  - phase32CodeDigest, phase32DigestPaths, phase32ReceiptPrograms, phase32ReceiptProblems, phase32OptimizationMatrixNonClaim
  - TestPhase32CodeDigest and TestPhase32ReceiptLivingClaims
affects: [32-08, 32-09, 32-11]
key-files:
  created:
    - scripts/codedigest.sh
    - internal/compiler/session/session_phase32_receipt_test.go
key-decisions:
  - "Receipt revision is a code digest of ls-tree output minus .md lines; sh and Go implement it twice and a test proves equality"
  - "Validator is pure over (docs, receipt bytes, logs, computed digest); absence of receipt and marker passes as an ordinary pass with no t.Skip (D-08c)"
  - "Digest drift alone never fails; only a bound marker that disagrees with the receipt or the computed digest does (D-10a)"
requirements-completed: []
status: complete
commits: 2
plan_head_before: dbd94f5353dd3c48a33b311fa88c4781bf8860c3
plan_head_after: 59ddfd797cba3ff44a4701f1e257ab61ea9f1a5d
actuals:
  tokens: 58000
  tasks: 2
  commits: 2
coverage:
  - id: D-08-D-08a
    description: "codedigest.sh and the Go digest agree on the real checkout and in temp repos; a code edit moves the digest while a .md edit and a .planning edit do not"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_receipt_test.go#TestPhase32CodeDigest"
        status: pass
    human_judgment: false
  - id: D-09-D-10a-D-08c
    description: "Receipt validator with named reasons and armed in-memory controls; reopen rule for bound versus reopened markers; passes with no receipt and no marker without t.Skip"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_receipt_test.go#TestPhase32ReceiptLivingClaims"
        status: pass
    human_judgment: false
duration: 25min
completed: 2026-10-12
---

# Phase 32 Plan 05: Receipt-binding primitives Summary

A receipt now binds to a code digest computed identically by `scripts/codedigest.sh` and a Go helper, and a pure validator enforces the receipt and living-claim rules once a receipt is ingested.

## Accomplishments

- `scripts/codedigest.sh`: POSIX sh, shasum with sha256sum fallback, named failures `unavailable_tool_sha256` and `git_ls_tree_failed`.
- `TestPhase32CodeDigest`: `sh-equals-go` on the real root, plus `armed-code-edit`, `md-edit-ignored`, `planning-edit-ignored` in temp git repos (each also checks sh equals Go).
- `phase32ReceiptProblems` with named reasons (claim_without_receipt, malformed_marker, stale_bound_claim, log_sha256_mismatch_*, log_digest_differs_receipt_*, tree_not_clean_*, host_missing, host_result_not_passing_*, program_missing_*, nonclaim_missing_optimization_matrix, schema_invalid and others).
- `TestPhase32ReceiptLivingClaims`: a real-checkout subtest (passes now: no receipt, no marker, no skip) and 21 armed in-memory rows.

## Observation (not a receipt)

`sh scripts/codedigest.sh` at the final code commit 59ddfd79: `e10ddcff21d94c90f96e03ed8bf93d4559739bc92a544acaf4369678df6aa51e`.

## Deviations from Plan

None. Two small additions beyond the listed reasons: `job_not_success_*`, `program_not_passing_*`, `unexpected_host_*`, `unexpected_program_*` and `log_missing_*` guard the same receipt fields.

## Known Stubs

None.

## Verification

Both Phase 32 receipt tests, the session meta set (including `TestSummaryCoverageGroundedness`), `TestValidationRowGradesAreEarnedOverArchivedCorpus`, the static test index check and the native spawn lint pass locally. EVD-15 is not marked complete (32-11 closes it).

## Self-Check: PASSED

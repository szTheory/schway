---
created: 2026-10-12T00:00:00Z
title: Phase 32 code-review hardening (deferred past the M006 receipt freeze)
area: evidence
severity: minor
files:
  - scripts/check-phase32-freeze.sh
  - scripts/codedigest.sh
  - scripts/check-phase32-receipts.sh
  - internal/compiler/session/session_phase32_receipt_test.go
  - internal/compiler/syntax/syntax_phase32_format_test.go
  - internal/compiler/session/evidence_grade_test.go
---

## Problem

`32-REVIEW.md` has 2 Warning and 5 Info findings. Each one fails closed or is a hardening or coverage gap, and none affects the M006 receipt (run 38315799626). They were deferred because the code was frozen at 079e112f, and any fix touches digest paths and reopens the receipts (D-10a).

- WR-01: drop `*.md` from the freeze gate's digest-path comparison so it matches `codedigest.sh`.
- WR-02: `codedigest.sh` should cd to `git rev-parse --show-toplevel`, fail on an empty `ls-tree` listing, and gain an armed control.
- IN-01: the receipt gate's pass line should say whether the expected digest came from the environment or was recomputed.
- IN-02: the living-claim marker scan should also catch indented or bulleted claims, and duplicate host rows in the receipt should be rejected.
- IN-03: add negative-context formatter tests for match arms, nested arms and if/while bodies.
- IN-04: `retiredCitationProblem` should require the specific citing row to be superseded.
- IN-05: document that the `buildTestIndex` process-wide cache is read-only, or key it per root.

## Solution

Land these in the first M007 phase that touches the evidence gates. Before that work closes, reopen the M006 hosted-receipt claim per D-10a and re-bind it, or mark it "reopened".

---
phase: "33"
review: 33-REVIEW.md
fixed: 3
skipped: 0
status: complete
---

# Phase 33 — Review Fix

## Fixed Issues

### CR-01: Empty manifest controls bypass receipt control validation

The reproduced bypass is a delimiter-only or whitespace-only declaration; an entirely empty declaration was already refused downstream. `TestReceiptGateMalformedControls` demonstrated exit 0 for arbitrary control rows on both hosts under the malformed declarations before the fix. Commit `1e207367` preserves that failing regression. Commit `99b72703` rejects missing controls and empty or whitespace-only comma-separated entries before log validation.

The six malformed-list subtests and the existing receipt/freeze/golden/producer/living-claim controls passed in a fresh focused session run (28.712s). Shell syntax and `git diff --check` passed. No archived receipt or log changed and no hosted receipt was bound.

## Skipped Issues

None.

## Fixed Issues — Phase 37 successor refresh, 2026-10-15

### WR-01: Coverage checker accepts incomplete verification records

Verification items are now separate records with required nonempty kind, ref
and status fields. The existing kind enum and canonical pass/fail/unknown
status vocabulary remain explicit; only pass satisfies test-backed entries.
Human checkpoints keep their rationale and valid non-passing statuses.
The original parser falsely accepted nine of 21 malformed controls, including
split items whose aggregate fields looked complete. The repaired parser passed
all 29 controls (eight valid, 21 invalid) and all 60 current summaries.
RED: `/tmp/schway33-wr01-red.log`, exit 1. GREEN:
`/tmp/schway33-wr01-green.log`, exit 0, package 0.112 seconds.
Only `internal/compiler/session/summary_coverage_test.go` changed for WR-01.

### WR-02: Receipt gate allows undeclared case programs

The manifest program inventory and case-program inventory must match in both
directions using literal field membership. Empty, whitespace-containing or
duplicate program members are rejected; reordered valid inventories remain
accepted. The pre-existing missing-declared-program check remains in its
original refusal position. The new reverse refusal is
`case_program_not_declared_inventory_extra`; the forward refusal remains
`program_without_table_entry_fizzbuzz`.

The RED control built matching table/log/result rows for 40 cases and reached
a false pass on the original gate, proving this was a consistency gap. The
repaired gate passed five exact roots, including archived M006, differential,
malformed-control and new program-inventory checks, without skips (26.219
seconds). RED: `/tmp/schway33-wr02-red.log`, exit 1. GREEN:
`/tmp/schway33-wr02-green.log`, exit 0. `sh -n` and diff checks passed.
Only `scripts/check-receipts.sh` and
`internal/compiler/session/session_receipt_gate_test.go` changed for WR-02.
No frozen golden, hosted log, corpus receipt or dependency changed.

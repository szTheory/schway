---
phase: "33"
review: 33-REVIEW.md
fixed: 5
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

## Fixed Issues

Phase 37 successor refresh, 2026-10-15.

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

## Fixed Issues

Closure boundary cases, 2026-10-15.

### WR-03: Empty case program bypasses the reverse inventory check

The receipt gate now emits an explicit full reason for malformed case records
rather than using an empty field as a no-mismatch sentinel. Required fields
are checked before inventory membership. The producer's real grammar is seven
logical fields: STDERR consumes the remainder and may contain spaces. The
repair preserves that grammar instead of imposing a false seven-token ceiling.
Seven malformed rows with matching 40-case logs reached false passes before
the repair; both multiword stderr controls remain accepted. The repaired six
receipt roots passed without skips (40.881 seconds); shell syntax and diff
checks passed. RED: `/tmp/schway33-wr03-red.log`, exit 1. GREEN:
`/tmp/schway33-wr03-green.log`, exit 0. New refusal: `case_table_malformed`.

### WR-04: Coverage checker still accepts incomplete entry metadata

The local canonical SUMMARY subset now checks nonempty string IDs and
descriptions, unique IDs, strict boolean human_judgment, nonblank checkpoint
rationale and optional-but-list verification shape. Scalar or malformed list
items fail closed; each mapping retains the required item-field checks from
WR-01. The guard is an explicit indentation/scalar subset, not a full YAML
implementation. Current summaries retain their valid objective or human
checkpoint classification. Reached malformed-entry controls and the final
current-corpus outcome are recorded in the final closure receipt. RED:
`/tmp/schway33-wr04-red.log`; GREEN: `/tmp/schway33-wr04-green.log`. No
existing summary needs a metadata rewrite for this repair.

Final WR-04 result: 71 controls (12 valid, 59 invalid) and all 60 current
summaries pass. The primary RED reproduced 29 false accepts and one valid
reordered-entry false refusal; the final boundary RED reached five additional
malformed-shape accepts. Final GREEN package time was 0.246 seconds.
The intermediate boundary RED is `/tmp/schway33-wr04-shape-red.log`.

---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
reviewed: 2026-10-15T19:42:42Z
depth: standard
files_reviewed: 25
files_reviewed_list:
  - cmd/schway/main_phase321_run_input_test.go
  - cmd/schway/main_phase32_routes_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/execution/defect_reasons.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase321_defect_test.go
  - internal/compiler/interp/interp_phase33_defect_test.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_phase33_capture_test.go
  - internal/compiler/native/native_phase33_defect_parity_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_freeze_gate_test.go
  - internal/compiler/session/session_gate_differential_test.go
  - internal/compiler/session/session_gate_retire_test.go
  - internal/compiler/session/session_phase32_receipt_test.go
  - internal/compiler/session/session_phase37_validation_test.go
  - internal/compiler/session/session_receipt_gate_test.go
  - internal/compiler/session/summary_coverage_test.go
  - scripts/check-freeze.sh
  - scripts/check-receipts.sh
  - scripts/codedigest.sh
  - scripts/verify-phase32.sh
  - testdata/phase16/public-emitter-consumers.json
  - testdata/receipts/M006.manifest
findings:
  critical: 0
  warning: 4
  info: 0
  total: 4
status: issues_found
---

# Phase 33: Code Review Report

**Reviewed:** 2026-10-15T19:42:42Z
**Depth:** standard
**Files Reviewed:** 25
**Status:** issues_found

## Summary

The review covers source snapshot `537e90ac` on `phase26-validation` and used the explicit 25-file scope override. The canonical resolver's degraded 77-file fallback was not used, and this report does not claim an incremental commit union. I reviewed the bounded capture reader, the single-owner/reason defect rule including its disabled-capture path, the generic POSIX freeze and receipt gates, and the current registry and metadata repairs. Historical CR-01, “Empty manifest controls bypass receipt control validation,” was fixed by commit `99b72703`; it is not reused here. No tests were run in this review, so validation receipts elsewhere remain historical or parent-reported evidence.

## Narrative Findings (AI reviewer)

### WR-01: Coverage checker accepts incomplete verification records

**File:** `internal/compiler/session/summary_coverage_test.go:84-104`
**Severity:** WARNING
**Issue:** The checker validates only values collected in `e.kinds` and `e.statuses`; it never requires either slice to contain one value for each verification reference. A `human_judgment: false` entry with a non-empty ref but no `kind:` or `status:` therefore passes this guard. The coverage consumer requires each verification item to provide kind, ref, and status, so this test can let malformed SUMMARY metadata reach downstream verification.
**Fix:** Track each verification item as its own record and reject a missing or empty `kind`, `ref`, or `status`; require every status to be `pass` for test-backed entries. Add controls that remove `kind:` and `status:` from an otherwise valid item.

### WR-02: Receipt gate allows undeclared case programs

**File:** `scripts/check-receipts.sh:86-95`
**Severity:** WARNING
**Issue:** The loop checks that every name in `programs=` appears in the case table and in each log, then checks every table row against the logs. It never checks the reverse relationship: a case row can name a program omitted from `programs=`. With `programs=checksum`, a matching checksum case and an additional matching case for another program can pass, while the result rows and pass summary still declare only `checksum`. That leaves the generic receipt's declared program inventory inconsistent with its case coverage.
**Fix:** Parse the program names used by the case table and require their unique set to equal the `programs=` set before validating logs; add a refusal control for an extra case-table program absent from `programs=`.

---

_Reviewed: 2026-10-15T19:42:42Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

## Closure Review — 2026-10-15T19:52:40Z

**Source:** `eb04b1a4`. This closure pass reviewed exactly `internal/compiler/session/summary_coverage_test.go` (fix commit `858453ae`), `scripts/check-receipts.sh`, and `internal/compiler/session/session_receipt_gate_test.go` (fix commit `67eee80d`). It does not refresh or claim a new review of the original 25-file scope; the original issued review had two warnings; frontmatter now counts the four cumulative findings, including this dated closure pass.

**WR-01:** Closed by source inspection. Verification items are now tracked separately and each item must have an allowed nonempty kind and status plus a nonempty ref; automated entries require pass. The invalid controls cover missing fields and fields split across different items.

**WR-02:** Closed by source inspection. The manifest program list is validated for empty, whitespace and duplicate members; the gate checks declared programs against case rows and logs, then checks each case program against the declaration. The new control constructs matching case and result rows for an undeclared program, and the existing missing-declared-program refusal remains before the reverse check.

The fix artifacts report 29 coverage controls and 60 summaries passing for WR-01, and a 40-case false-pass reproducer followed by five focused passing receipt roots for WR-02. The final parent-run receipt records 48/48 exact roots across six packages and six passing metadata roots with no missing, failed or skipped roots. These are parent-reported checks, not runs performed during this closure review.

**Effective original warnings remaining:** 0 of 2. Source inspection also found two new false-pass cases in the same scoped files:

### WR-03: Empty case program bypasses the reverse inventory check

**File:** `scripts/check-receipts.sh:101-111`
**Severity:** WARNING
**Issue:** The reverse check prints `$3` for an undeclared case program, then treats an empty command-substitution result as “no undeclared program.” A manifest row beginning `# case ` with no program therefore evades that check. If both logs contain the corresponding empty-program case row and matching counts, the later exact-row comparison can accept it. This row does not satisfy the documented `# case PROGRAM KEY EXIT STDOUT STDERR` grammar.
**Fix:** Validate each case row's required fields, including a nonempty program, before inventory comparison; use an explicit found flag or sentinel so an empty undeclared value cannot be confused with no mismatch.

### WR-04: Coverage checker still accepts incomplete entry metadata

**File:** `internal/compiler/session/summary_coverage_test.go:49-52,108-132`
**Severity:** WARNING
**Issue:** The parser does not retain or validate an entry description, and it never requires a nonempty or unique entry ID. A whitespace-only rationale also passes because the check compares only with the empty string. The coverage consumer requires nonempty unique IDs, a nonempty description, and a nonblank rationale for human checkpoints, so the groundedness test still accepts entries that the consumer rejects.
**Fix:** Parse the description, require nonempty trimmed IDs and descriptions, reject duplicate IDs, and trim the rationale check. Add controls for missing/blank descriptions, empty/duplicate IDs, and whitespace-only rationale.

**Effective warning count after closure inspection:** 2 new warnings (WR-03 and WR-04); WR-01 and WR-02 are closed. No tests were run during this closure review.

---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
reviewed: 2026-10-13T18:00:00Z
depth: standard
files_reviewed: 23
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
  - internal/compiler/session/session_receipt_gate_test.go
  - internal/compiler/session/summary_coverage_test.go
  - scripts/check-freeze.sh
  - scripts/check-receipts.sh
  - scripts/codedigest.sh
  - scripts/verify-phase32.sh
  - testdata/receipts/M006.manifest
findings:
  critical: 1
  warning: 0
  info: 0
  total: 1
status: issues_found
---

# Phase 33: Code Review Report

**Reviewed:** 2026-10-13T18:00:00Z  
**Depth:** standard  
**Files Reviewed:** 23  
**Status:** issues_found

## Summary

Reviewed the 23 files in the canonical Phase 33 scope, including the interpreter/native boundary and the generic receipt and freeze gates. The receipt gate can accept a log with an arbitrary control row when the manifest declares an empty controls list, so malformed future milestone manifests can bypass the control checks.

## Critical Issues

### CR-01: Empty manifest controls bypass receipt control validation

**Classification:** BLOCKER  
**File:** `scripts/check-receipts.sh:39-44`
**Issue:** `control_count` is calculated with `awk -F, '{print NF}'`. For an empty `controls=` value, awk still reports one field, while the validation loop at lines 99-101 iterates over no declared control names. The gate then requires exactly one row beginning with the manifest label's `control ` prefix but never checks its name, package, or tests. A two-host receipt pair with any single control row can therefore pass without running any declared control. The current M006 manifest has nonempty controls, but this generic gate is intended for later milestones and should reject an empty declaration rather than silently waive the controls.

**Fix:** Reject an empty `controls` value before deriving the count, and reject empty entries in the comma-separated list before checking rows. For example:

```sh
[ -n "$controls" ] || reject controls_line_missing
control_count=$(printf '%s' "$controls" | awk -F, 'NF { for (i=1; i<=NF; i++) if ($i == "") exit 1; print NF }') || reject controls_line_malformed
```

## Footer

_Reviewed: 2026-10-13T18:00:00Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

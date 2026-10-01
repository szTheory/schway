---
phase: 24-ownership-transfer-through-calls-and-errors
reviewed: 2026-10-01T11:22:25Z
depth: standard
files_reviewed: 7
files_reviewed_list:
  - internal/compiler/native/phase24_observer_test.go
  - internal/compiler/native/testdata/phase24_observer.c
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/session/session_phase24_model_test.go
  - examples/phase24/README.md
  - scripts/verify-phase24.sh
  - .github/workflows/ci.yml
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 24: Code Review Report

**Reviewed:** 2026-10-01T11:22:25Z  
**Depth:** standard  
**Files Reviewed:** 7  
**Status:** issues_found

## Summary

Reviewed the Phase 24 observer, model and emitter tests, example contract, focused verifier, and CI wiring. The model cleanup assertion does not check which owner each release event discharged, so it can accept an incorrect release ordering while claiming to verify the expected C/B/A cleanup sequence.

## Warnings

### WR-01: Model cleanup check does not verify owner identity or order

**File:** `internal/compiler/native/phase24_observer_test.go:413-421`  
**Issue:** The typed-error replay collects only `FunctionID` for `resource.released` and compares `probe, main, main`. Swapping the releases of caller owners A and B would still produce the same function sequence, so this assertion cannot establish the stated helper-C, caller-B, caller-A order or catch a model regression that discharges the wrong owner. `execution.Event` also carries `SourcePlace` and `Invocation`, which allow the test to check the relevant owner place and activation.  
**Fix:** Assert the release events' `SourcePlace` sequence against the owner places established by the checked fixture, and verify the expected invocation IDs as well as their functions.

---

_Reviewed: 2026-10-01T11:22:25Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

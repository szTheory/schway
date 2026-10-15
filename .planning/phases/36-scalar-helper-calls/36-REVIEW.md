---
phase: 36-scalar-helper-calls
reviewed: 2026-10-15T23:46:42Z
depth: standard
files_reviewed: 53
files_reviewed_list:
  - cmd/schway/main_phase32_routes_test.go
  - cmd/schway/main_phase36_test.go
  - examples/helper_calls/crc_step.schway
  - examples/helper_calls/importer.schway
  - examples/helper_calls/public.schway
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_phase273_test.go
  - internal/compiler/cgen/cgen_phase35_test.go
  - internal/compiler/cgen/cgen_phase36_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase34_refusal_test.go
  - internal/compiler/check/check_phase36_scalar_calls_test.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_phase36_legacy_internal_test.go
  - internal/compiler/corevalidate/corevalidate_phase36_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_phase36_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase36_test.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_phase36_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_phase36_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_phase36_borrow_test.go
  - internal/compiler/pathoracle/pathoracle_phase36_test.go
  - internal/compiler/session/session_phase34_limit_consumers_test.go
  - internal/compiler/session/session_phase34_prechange_test.go
  - internal/compiler/session/session_phase34_surface_test.go
  - internal/compiler/session/session_phase35_admission_mutation_test.go
  - internal/compiler/session/session_phase35_closeout_test.go
  - internal/compiler/session/session_phase35_surface_test.go
  - internal/compiler/session/session_phase36_byteview_test.go
  - internal/compiler/session/session_phase36_calls_test.go
  - internal/compiler/session/session_phase36_legacy_calls_test.go
  - internal/compiler/session/session_phase36_peer_mutation_test.go
  - internal/compiler/session/session_phase36_prechange_test.go
  - internal/compiler/session/session_phase36_surface_test.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/surfacemanifest/manifest_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase36_test.go
  - testdata/phase36/helper-values.json
  - testdata/phase36/new-fixture-ledger.json
  - testdata/phase36/prechange-corpus.json
  - testdata/phase36/surface-census.json
  - testdata/phase36/surface-witnesses.json
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 36: Code Review Report

**Reviewed:** 2026-10-15T23:46:42Z
**Depth:** standard
**Files Reviewed:** 53
**Status:** issues_found

## Summary

The 53-path review found one maintainability defect in the raw call validator's ownership commentary: it still describes the pre-Phase-36 same-type call invariant even though scalar calls now allow distinct argument and result types. The implementation derives ownership from the argument's own type, so this is not a runtime failure. The supplied refresh artifacts report 84 passing exact roots across 13 packages and both prohibition checks green; these were treated as supporting evidence, not substitutes for source review. The earlier WR-01 attribution finding remains a fixed historical item and is not reopened.

## Narrative Findings (AI reviewer)

### Warnings

### WR-02: Raw call ownership comment states an obsolete same-type invariant

**Classification:** WARNING
**File:** `internal/compiler/corevalidate/corevalidate.go:4248-4270`
**Issue:** This comment says every operation's `operation.TypeID` equals the source argument type and that a callee's declared return type equals its parameter type. Phase 36 admits all U64/Bool parameter/result pairs, including cross-typed calls, and `consumeCallArgument` now correctly derives copy/move behavior from `sourceTypeID` while the separate contract check compares the argument and result independently. The stale claims also contradict the current comment at lines 4286-4296, so a future change could incorrectly reintroduce result-type-based ownership derivation or break cross-typed calls.
**Fix:** Rewrite the commentary to state that call ownership is derived from `sourceTypeID`; the independent call-contract checks separately validate source against the callee parameter and target against the callee result. Remove the old same-type assertions or label them as historical pre-Phase-36 context.

---

_Reviewed: 2026-10-15T23:46:42Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

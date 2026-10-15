---
phase: 36-scalar-helper-calls
reviewed: 2026-10-15T06:27:35Z
depth: standard
files_reviewed: 49
files_reviewed_list:
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
  - internal/compiler/session/session_phase34_surface_test.go
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
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 36: Code Review Report

**Reviewed:** 2026-10-15T06:27:35Z
**Depth:** standard
**Files Reviewed:** 49
**Status:** clean

## Summary

The fixed peer now reports the expected callee for truncated dynamic invocations at ordinal one, and regression controls cover both pending-child and unterminated-activation cases. No additional concrete bugs, security issues, or quality defects were found in the supplied scope.

## Narrative Findings (AI reviewer)

All reviewed files meet quality standards. No issues found.

---

_Reviewed: 2026-10-15T06:27:35Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_


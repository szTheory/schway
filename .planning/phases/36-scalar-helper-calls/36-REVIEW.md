---
phase: 36-scalar-helper-calls
reviewed: 2026-10-15T06:21:47Z
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
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 36: Code Review Report

**Reviewed:** 2026-10-15T06:21:47Z
**Depth:** standard
**Files Reviewed:** 49
**Status:** issues_found

## Summary

The scalar-call parser, checker, independent core peers, interpreter, native emitter, and legacy ByteView composition were reviewed at standard depth. One warning remains: truncated traces after repeated dynamic calls lose the expected callee identity in the peer error.

## Narrative Findings (AI reviewer)

### Warnings

#### WR-01: Truncated repeated-call traces omit the expected callee

**Classification:** WARNING
**File:** `internal/compiler/executionpeer/executionpeer.go:681-684`
**Issue:** When a dynamic call stream is truncated after a repeated call site has created a child with ordinal greater than zero, the child path is present in `dynamicOccurrences` but absent from the statically unfolded `i.occurrences` map. Both terminal error branches read the static map, so `orphan_child` (and the nonempty-stack refusal for a truncated dynamic activation) return an empty `ExpectedFunction`. Consumers lose the callee identity precisely when reporting an incomplete causal trace.
**Fix:** Resolve the pending/stack invocation from `dynamicOccurrences` when `i.dynamicCallIdentity` is true, and from `i.occurrences` otherwise; use that selected occurrence's `functionID` in both refusal paths.

---

_Reviewed: 2026-10-15T06:21:47Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_


---
phase: 17-return-type-parameter-type
reviewed: 2026-09-22T00:00:00-04:00
depth: standard
files_reviewed: 26
files_reviewed_list:
  - cmd/lang-repair/antitheater_test.go
  - cmd/lang-repair/import_boundary_test.go
  - cmd/lang-repair/repair_test.go
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_repair_emission_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/reduce/reduce.go
  - internal/compiler/reduce/reduce_test.go
  - internal/compiler/session/session_phase17_native_test.go
  - internal/compiler/session/session_phase17_test.go
  - internal/compiler/session/witness_registry_test.go
  - internal/compiler/syntax/parser.go
  - testdata/phase17/HELDOUT.sha256
  - testdata/phase17/call_argument_type_mismatch.lang
  - testdata/phase17/call_return_type_unrepresentable.lang
  - testdata/phase17/derivation_call_argument_mismatch.lang
  - testdata/phase17/heldout_call_argument_mismatch.lang
  - testdata/phase17/return_type_tracer.lang
findings:
  critical: 2
  warning: 0
  info: 0
  total: 2
status: issues_found
---

# Phase 17: Code Review Report

**Reviewed:** 2026-09-22T00:00:00-04:00
**Depth:** standard
**Files Reviewed:** 26
**Status:** issues_found

## Summary

The review covered all source artifacts identified by the nine Phase 17 summaries and cross-checked them against the phase commit range `f1201fbc..163fc8b`. Directional call facts are added, but the function's declared return type is not enforced at a terminal return and the C emitter does not model the declared return type's alternatives. The current one-alternative tracer masks both failures.

## Critical Issues

### CR-01: Terminal returns are not checked against the declared return contract

**File:** `internal/compiler/check/check.go:2222-2230`, `internal/compiler/corevalidate/corevalidate.go:2069-2076`

**Issue:** Phase 17 creates `type:1` for the declared return type, but an `OpReturn` continues to use the source place's TypeID. Neither the checker nor core validator requires that TypeID to be the function's declared return fact. Consequently, the checked-in `classify(Resource) -> Result` fixture returns its `Resource` parameter (`testdata/phase17/return_type_tracer.lang:16-20`) and is accepted as a `Result`. This lets callers receive a value whose runtime alternatives belong to a different nominal type than the declared call contract. The native branch lowering then paper-overcasts that value to the return C type at `cgen_program.go:1035-1040`, which is invalid for distinct payload structs and semantically wrong even when enum representations happen to compile.

**Fix:** At every ordinary `OpReturn`, require the returned place's fact to equal the function's `ID + ":type:1"` fact (and independently enforce this in corevalidate and originvalidate). If a distinct return is meant to be constructed from a match arm, lower an explicit construction/conversion operation with a destination carrying `type:1`; otherwise reject returning the parameter-typed source. Add source and native negative tests for `Resource -> Result { ... resource }`, including payload-bearing types.

### CR-02: Native return layouts are built from parameter alternatives, so valid distinct-result matches fail or emit the wrong value

**File:** `internal/compiler/cgen/cgen_program.go:558-582`, `internal/compiler/cgen/cgen_program.go:944-955`

**Issue:** `branchTypeFor` caches by nominal `typeName`, but the Phase 17 return path calls it as `branchTypeFor(function.ReturnType, function.Parameter.Type)`. Thus a `Result` C type is populated with `Resource` alternatives. The checker correctly accepts a bare match arm only when `arm.Value` is an alternative of the declared return data type (`check.go:302-355`), but native lowering looks up that value in the parameter-derived map. With more than one result alternative, `EmitNative` returns “match arm names unknown alternative”; with one alternative, the fallback at lines 949-950 silently emits the parameter alternative instead. C layout and behavior therefore depend on whether a nominal type is first encountered as a parameter or return type, rather than its declaration.

**Fix:** Construct each branch type's layout and alternative map from its own declared nominal type: call `branchTypeFor(function.ReturnType, function.ReturnType)`, remove the one-alternative fallback, and introduce an explicit, checked conversion only where the core operation actually requires one. Add a multi-alternative `Input -> Output` bare-match fixture and verify interpreter/O0/O3/O3-LTO all return the selected `Output` alternative.

---

_Reviewed: 2026-09-22T00:00:00-04:00_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

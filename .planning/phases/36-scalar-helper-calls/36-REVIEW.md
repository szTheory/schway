---
phase: 36-scalar-helper-calls
reviewed: 2026-10-16T01:51:52Z
depth: standard
files_reviewed: 56
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
  - internal/compiler/cgen/cgen_phase36_caller_shapes_test.go
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
  - internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go
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
  - testdata/phase16/public-emitter-consumers.json
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

**Reviewed:** 2026-10-16T01:51:52Z
**Depth:** standard
**Files Reviewed:** 56 cumulative unique paths
**Status:** issues_found

## Summary

The initial standard pass reviewed the explicit 54-path scope and found CR-01, a native-emission failure for checker-admitted structured scalar helper callers, plus the deferred documentation warning WR-02. The dated closure review rechecked cgen and the new caller-shape tests, confirmed CR-01's repair with the two exact regression roots, and leaves WR-02 as the only current finding. The cumulative inventory contains 56 unique paths: cgen_program.go was in the initial scope; the caller-shape test and public-emitter registry add one path each.

The parent-supplied post-fix refresh artifacts at d69380edd8ab276539940fcc0ed0d1a383b07786 record 86 roots across 13 packages with no missing, failed, or skipped roots; both canonical prohibition checks are green. I independently ran the two new focused roots and they passed. The five-root safety run and original-source overlay RED are parent-supplied evidence, not rerun in this closure. No full package or project suite was run.

Phase 33 execution-peer findings CR-03 (legacy nested defects rejected) and CR-04 (propagated typed failures rejected) remain separate, pre-existing blockers deferred for the M007 audit. They do not overlap the scalar helper-call path and are not counted as Phase 36 findings. The proof-bridge supplement remains outside the source scope and was not rerun here.

## Narrative Findings (AI reviewer)

### Warnings

### WR-02: Raw call ownership comment states an obsolete same-type invariant

**Classification:** WARNING
**File:** internal/compiler/corevalidate/corevalidate.go:4248-4270
**Issue:** The comment still states that a call's operation.TypeID is the argument type and that a callee's declared return type equals its parameter type. Phase 36 admits cross-typed U64/Bool calls, and consumeCallArgument correctly derives copy/move behavior from sourceTypeID while separate checks validate the argument and result contracts. The stale description contradicts the current code comments at lines 4286-4296 and risks guiding a future change toward deriving ownership from the result type. This finding remains deferred in 36-REVIEW-DISPOSITION.md; the current implementation behavior is correct.
**Fix:** When the core-validator's next semantic edit is made, rewrite this commentary to say that call ownership is derived from sourceTypeID, while the independent call-contract checks validate the source against the callee parameter and the target against the callee result.

## Dated Closure Review — 2026-10-16T01:46:23Z

**Closure scope:** internal/compiler/cgen/cgen_program.go and internal/compiler/cgen/cgen_phase36_caller_shapes_test.go. The first path overlaps the initial 54; the second adds the 55th cumulative unique path. The initial review timestamp was 2026-10-16T01:35:55Z.

### CR-01 (closed): Checker-admitted scalar CFG callers outside cgen selection failed native emission

**Initial classification:** BLOCKER
**Initial file:** internal/compiler/cgen/cgen_program.go:2674-2676, 2937-2939
**Initial issue:** The checker admitted an unused Byte-to-U64 structured scalar body, and the parser routed a body beginning with var to scalar CFG parsing regardless of parameter type. A body such as var result = crc_step(1) lowered to OpCall plus OpScalarStore while leaving the Byte formal unused. With no other dynamic-eligible call in the program, hasScalarValueCalls did not enable dynamic invocation because isDynamicScalarFunctionBody excluded Byte callers. hasScalarOperations still routed the function to emitProgramScalarFunction, whose OpCall arm refused emission. The same mismatch affected checker-admitted ByteView-to-Unit bodies containing helper calls because the classifier admitted ByteView only for U64 results. An unrelated eligible scalar call could mask the failure by enabling the global dynamic flag.
**Initial fix suggestion:** Include every checker-admitted scalar CFG caller shape that may contain OpCall in cgen's dynamic body classifier, prove the unused-Byte formal is not referenced, and retain the U64/Bool callee signature restriction; otherwise reject unsupported caller shapes in the checker.

**Closure:** Fixed in d69380edd8ab276539940fcc0ed0d1a383b07786. The classifier now admits unused Byte-to-U64 and ByteView-to-Unit caller shapes. For the unused Byte bridge it checks all six operation place-reference fields (SourceID, RightID, TargetID, StoreTargetID, ErrTargetID, PayloadTargetID) before skipping the Byte formal; helper callee signatures remain U64/Bool pairs. No checker or peer admission changed.

**Evidence and limits:** The parent-supplied original-source overlay log shows both new named roots fail at the expected native-emission assertions with the scalar-call-support error. I independently ran GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^(TestPhase36UnusedByteStructuredHelper|TestPhase36ByteViewUnitHelperEmission)$' -count=1 -timeout=600s; both passed. The Byte test independently checks interpreter result 11 and native O0/O3/O3-LTO result 11 with causal validation. The ByteView Unit test retains an uncalled Unit body in the translation unit, compiles and links it, then runs the unrelated U64 entry for result 7 and newline; it makes no runtime claim about executing that Unit body or introducing a ByteView entry/value-call route. The parent-supplied five-root safety log and the 86-root/13-package refresh plus both fresh green prohibition checks provide additional evidence; they were not rerun in this closure.

## Dated Closure Supplement — 2026-10-16T01:51:52Z

**Scope:** `testdata/phase16/public-emitter-consumers.json`, the 56th cumulative unique path. The preceding 54-path source review and CR-01 closure evidence remain as recorded above.

**Inventory gap (resolved):** The initial source/registry union had 116 AST-discovered public emitter callsites and 115 registry rows; this was the sole initial union failure. The new row at `testdata/phase16/public-emitter-consumers.json:5-7` records `internal/compiler/cgen/cgen_phase36_caller_shapes_test.go:EmitNative:41`, classification `admitted-dynamic-schema2`, and witness `probe:TestPhase36UnusedByteStructuredHelper`. I inspected the source call at line 41 and its enclosing test: it checks the admitted unused-Byte helper, calls `EmitNative`, and exercises the resulting native artifact. The inventory scanner parses Go AST calls and records the source path, emitter method, and token line, so the registry location matches the actual callsite and the witness matches the test that covers it.

**Evidence and limits:** The parent-supplied `/tmp/schway36-r2-inventory-green.log` records passing `TestPhase16PublicEmitterConsumerInventory` and `TestPhase16EmitterInventoryMutationControls`, including missing, stale, duplicate, invalid-classification, and missing-refusal-witness controls. I inspected the current row, callsite, and inventory-matching code; I did not rerun these tests or the behavioral roots in this supplement. The initial union failure is resolved bookkeeping, not a new code finding. WR-02 remains the only current finding; CR-01 remains closed with its prior reproduction limits and history intact.

---

_Reviewed: 2026-10-16T01:51:52Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

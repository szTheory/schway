---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
reviewed: 2026-10-16T03:25:03Z
depth: standard
files_reviewed: 29
files_reviewed_list:
  - docs/agent-surface.md
  - examples/bitwise.schway
  - examples/checksum/adler32.schway
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_phase33_scalar_defect_test.go
  - internal/compiler/cgen/cgen_phase36_caller_shapes_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go
  - internal/compiler/executionpeer/executionpeer_phase35_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session_phase35_admission_mutation_test.go
  - internal/compiler/session/session_phase35_closeout_test.go
  - internal/compiler/session/session_phase35_peer_mutation_test.go
  - internal/compiler/session/session_phase35_prechange_test.go
  - internal/compiler/session/session_phase35_surface_test.go
  - internal/compiler/session/session_phase35_values_test.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - testdata/phase35/prohibition-evidence.test.cjs
  - testdata/phase16/public-emitter-consumers.json
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 35: Code Review Report

**Reviewed:** 2026-10-16T03:25:03Z
**Depth:** standard
**Files Reviewed:** 29
**Status:** clean for the Phase 35 scope

## Summary

This standard-depth freshness review covers the 26 paths in the previous report plus three explicit additions, for 29 unique paths at source HEAD 9bc84d95da311beb24cf3e5621f1483346eaf01b. Between the previous report's source HEAD f19721428ec02008b17c3dbbb13a4169f035ef89 and this HEAD, the in-scope source/test changes are limited to cgen_program.go and the two new cgen regression files; the public-emitter registry adds their two EmitNative callsites. The earlier 66-path union remains historical and was not reviewed again as a union.

The current emitter admits the unused-Byte caller shape only when its Byte formal is absent from every operation place-reference field, while retaining the old route for other shapes. Its added native defect terminal path applies only when dynamic invocation support is active; application shells and no-call native emission retain their legacy failure route. The Phase33 explicit-defect regression and Phase36 caller-shape regressions passed at current HEAD with the exact three named cgen roots. The parent-provided current-head Phase35 refresh receipt records all 22 task selectors freshly run, covering 66 distinct names and 72 roots across 13 packages, with zero missing, failed, or skipped roots. That receipt includes the frozen pre-change emitter corpus and controls, shift site and unsigned-result checks, event-free bitwise vectors, causal mutation controls, and value/refusal witnesses. These executions were run by the parent, not by this reviewer; I ran no broad suite.

This report's clean status is limited to Phase 35 behavior. Existing Phase 33 CR-03 and CR-04 remain deferred BLOCKERs for the mandatory M007 audit before Phase 38. CR-03 concerns legacy nested-defect unwinding at executionpeer.go:693-699; CR-04 concerns classifying modeled typed failure propagated through OpReturn at executionpeer.go:735-741. Neither affects the event-free bitwise/shift operations, literal shift rules, hex handling, or Adler witness. Phase 33 CR-06 also remains a deferred BLOCKER: no-call native scalar explicit defects still follow the legacy application-failure path at cgen_program.go:3008-3010. The frozen no-call C corpus receipt protects generated-C compatibility and does not establish native defect-document parity. These are existing external audit obligations, not new Phase 35 findings, and are not counted in the frontmatter. Phase 33 CR-05 is fixed: the current dynamic native path emits the structured defect terminal, and the current-head explicit-defect regression passes for entry and helper callers across O0, O3, O3-LTO, and application-shell behavior.

Phase 36 WR-02 remains a deferred comment warning at corevalidate.go:4254-4270: the prose describes an obsolete same-type invariant, while current implementation and tests derive from sourceTypeID correctly. It is unrelated to Phase 35 semantics and remains external documentation debt for the next semantic edit. Phase 35 WR-01 and WR-02 remain fixed in 35-REVIEW-DISPOSITION.md. Current evidence is local targeted validation; it is not a hosted receipt, two-host receipt, or whole-project suite.

## Narrative Findings (AI reviewer)

No new Phase 35-specific finding was confirmed in the current 29-path scope. Carried Phase 33 blockers CR-03/04/06 and Phase 36 warning WR-02 remain in their existing review/disposition records; this report does not close or hide that external debt. Phase 33 CR-05 remains fixed.

## Historical Review Passes

The prior report's 26-path freshness pass at f19721428ec02008b17c3dbbb13a4169f035ef89 is superseded by this 29-path review. The original 62-path standard-depth pass was recorded at 2026-10-15T21:02:28Z and reported no confirmed correctness, security, or maintainability defect. Its historical refresh receipt at d48227882579947201f607814585c42e6ad29 was not evidence from the current source head.

The prior follow-up at HEAD 04d1407a reviewed five wiring paths, with one overlap against the original 62, producing the historical 66-path union. It separately recorded four targeted adapter runs: both known-bad fixtures failed as expected, while the clean fixture and normal subject passed. That history remains distinct from this freshness pass and its parent-provided current receipt.

---

_Reviewed: 2026-10-16T03:25:03Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

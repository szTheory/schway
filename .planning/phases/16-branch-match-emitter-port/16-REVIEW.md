---
phase: 16-branch-match-emitter-port
reviewed: 2026-09-21T21:50:52Z
depth: deep
files_reviewed: 33
files_reviewed_list:
  - cmd/lang-repair/repair_test.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_n1_convergence_test.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen_restrict_probe_test.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/cgen/legacy_emitter_evidence_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/native/foreign_retained_test.go
  - internal/compiler/native/native_lto_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_payload_control_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase16_emitter_inventory_test.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_phase5_corpus_test.go
  - internal/compiler/session/session_phase5_mismatch.go
  - internal/compiler/session/session_phase5_sanitize.go
  - internal/compiler/session/session_phase6.go
  - internal/compiler/session/session_phase6_injectors_test.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/witness_registry_test.go
  - testdata/phase16/file-frozen-evidence.json
  - testdata/phase16/generated-frozen-evidence.json
  - testdata/phase16/public-emitter-consumers.json
findings:
  critical: 1
  warning: 1
  info: 0
  total: 2
status: issues_found
---

# Phase 16: Code Review Report

**Reviewed:** 2026-09-21T21:50:52Z
**Depth:** deep
**Files Reviewed:** 33
**Status:** issues_found

## Summary

The Phase 16 public emitter cutover, branch/match lowering, schema-2 projection, and the migrated evidence controls were reviewed across their call chains. The focused cgen/core package gate completed successfully, and the changed Go files are formatted with no diff whitespace errors. However, the evidence migration introduces a production-reachable alternate lowering path for inputs the public emitter intentionally refuses, and that path does not bind file-backed artifacts to the supplied core program. These invalidate the stated sole-emitter/M004-refusal boundary.

## Narrative Findings (AI reviewer)

## Critical Issues

### CR-01: Production session helper bypasses the M004 public-emitter refusal

**File:** `internal/compiler/session/session_phase5.go:113`
**Issue:** `Phase16ControlNativeC` is compiled into the non-test `session` package and is called by non-test session command/gate code (for example `session.go:427`, `session.go:1065`, `session.go:2528`, `session.go:2877`, and `session_phase5.go:578`). When `cgen.EmitNative` refuses a recognized M004 fixture, the helper reads and returns historical C instead (lines 128-155); its file/generated fallbacks do the same (lines 158-226). Thus any production caller of `session` can obtain executable C for a public-emitter-refused foreign or by-pointer program simply by supplying an approved fixture identity. This is a hidden alternate native-emission authority, directly violating the phase's requirement that the public emitter be the sole production dispatcher and that M004 inputs be refusal-only.

**Fix:** Keep frozen-artifact loading strictly test-only. Move the fallback router and its callers into `_test.go` helpers (or inject an artifact reader only into test control code), and make non-test session paths return the `cgen.EmitNative` refusal unchanged. Production commands that need historical evidence should report the refusal plus provenance metadata, never return historical C as a compilable artifact.

### WR-01: File-backed frozen evidence is not bound to the supplied core program

**File:** `internal/compiler/session/session_phase5.go:158`
**Issue:** `phase16FileControlNativeC` validates only the fixture file's SHA-256 and the historical artifact's SHA-256 (lines 170-185). It neither re-derives `program` from that fixture nor records/verifies a canonical checked-program digest. Consequently, a caller can pass an unrelated, manually constructed or drifted `core.Program` that happens to receive a by-pointer/foreign refusal along with the path of a valid frozen fixture and receive that fixture's historical C. This can conceal checker/core-IR drift and makes the artifact evidence attest to the path string rather than the actual compilation input.

**Fix:** Add a canonical program digest to each file-backed manifest record and verify it against `json.Marshal(program)` before loading the artifact. Prefer an API that accepts fixture source, parses/checks/validates it internally, and refuses unless the resulting program digest matches the record; also reject duplicate fixture/program records.

---

_Reviewed: 2026-09-21T21:50:52Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: deep_

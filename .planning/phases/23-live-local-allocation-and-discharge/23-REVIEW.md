---
phase: 23-live-local-allocation-and-discharge
reviewed: 2026-09-28T15:34:45Z
depth: standard
files_reviewed: 22
files_reviewed_list:
  - internal/compiler/check/check.go
  - internal/compiler/ability/ability.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/bindings.go
  - internal/compiler/native/native_app.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - examples/phase23/file_byte.lang
  - examples/phase23/file_byte.bindings.json
  - examples/phase23/README.md
  - testdata/phase23/discard_owner.lang
  - cmd/lang/main.go
  - scripts/verify-phase23.sh
  - .github/workflows/ci.yml
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 23: Code Review Report

**Reviewed:** 2026-09-28T15:34:45Z
**Depth:** standard
**Files Reviewed:** 22
**Status:** issues_found

## Summary

Reviewed the Phase 23 production, example, build, CLI, and CI paths identified by the seven plan and summary artifacts, cross-checked against the Phase 23 commit range. The acquisition and owner lifecycle paths are narrowly admitted and the generated release precedes reporting use failure. One cleanup robustness issue remains in the C adapter's acquisition failure paths.

## Warnings

### WR-01: Acquisition failure paths ignore descriptor close failures

**File:** `examples/phase23/adapter.c:82-118`
**Issue:** Several paths close an opened descriptor and discard the return value, including `fstat`/non-regular rejection, allocation failure, empty input, read failure, overlong input, and trailing-read failure. If `close()` itself fails while the descriptor remains open, the function returns its original error and loses the descriptor, allowing repeated failures to leak descriptors. This also means these paths cannot support the documented assertion that acquisition failures close their descriptors.
**Fix:** Centralize failure cleanup so it records the close result and reports `LANG_FILE_BYTE_CLOSE_FAILED` when cleanup fails (while preserving the primary failure in a separate internal cause if needed). Do not retry `close()` blindly after an error because descriptor state can be platform-dependent; make the cleanup outcome explicit and ensure any retry policy is platform-specific.

---

_Reviewed: 2026-09-28T15:34:45Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_

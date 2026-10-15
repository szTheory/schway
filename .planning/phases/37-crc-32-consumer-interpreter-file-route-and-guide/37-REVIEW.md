---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
reviewed: 2026-10-15T18:54:26Z
depth: standard
files_reviewed: 39
files_reviewed_list:
  - README.md
  - cmd/schway/main.go
  - cmd/schway/main_phase32_routes_test.go
  - cmd/schway/main_phase34_routes_test.go
  - cmd/schway/main_phase37_crc32_controls_test.go
  - cmd/schway/main_phase37_file_run_test.go
  - cmd/schway/main_phase37_guide_test.go
  - cmd/schway/main_phase37_inventory_test.go
  - cmd/schway/main_phase37_matrix_test.go
  - examples/crc32/README.md
  - examples/crc32/adapter.c
  - examples/crc32/adapter.h
  - examples/crc32/crc32.expected.json
  - examples/crc32/crc32.schway
  - examples/crc32/crc_low.schway
  - examples/crc32/file_bytes.bindings.json
  - examples/crc32/lib/crc32.schway
  - internal/compiler/interp/filebytes_reader.go
  - internal/compiler/interp/filebytes_reader_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase37_test.go
  - internal/compiler/native/native_phase37_filebytes_test.go
  - internal/compiler/session/session_file_application.go
  - internal/compiler/session/session_phase28_exitdocs_test.go
  - internal/compiler/session/session_phase34_limit_consumers_test.go
  - internal/compiler/session/session_phase34_prechange_test.go
  - internal/compiler/session/session_phase34_surface_test.go
  - internal/compiler/session/session_phase36_prechange_test.go
  - internal/compiler/session/session_phase37_authority_test.go
  - internal/compiler/session/session_phase37_crc32_test.go
  - internal/compiler/session/session_phase37_file_application_test.go
  - internal/compiler/session/session_phase37_reuse_test.go
  - internal/compiler/session/session_phase37_surface_test.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/surfacemanifest/manifest_test.go
  - internal/compiler/testsupport/crc32_oracle.go
  - internal/compiler/testsupport/crc32_oracle_test.go
  - testdata/phase37/new-fixture-ledger.json
  - testdata/phase37/surface-witnesses.json
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 37: Code Review Report

**Reviewed:** 2026-10-15T18:54:26Z  
**Depth:** standard  
**Files Reviewed:** 39  
**Status:** issues_found

## Summary

Reviewed the 39-file source scope, including the CRC consumer, interpreter file route, bounded reader, native adapter, CLI parsing, surface facts, and their focused evidence. One user-facing CLI usage defect remains: the error syntax suggests an interpreter manifest is accepted, although the parser rejects it. No critical issues were found.

## Narrative Findings (AI reviewer)

### WR-01: Usage Syntax Suggests Interpreter Accepts a Manifest

**Classification:** WARNING  
**File:** `cmd/schway/main.go:183`  
**Issue:** `fileRunUsage` places `[--manifest MANIFEST]` after a shared engine alternative, so users are told the option is available with `--engine=interpreter`. `parseFileRun` rejects that combination at line 265, and the grammar test explicitly treats it as invalid. Following the usage text therefore produces another usage error instead of a supported command.
**Fix:** Show the two productions separately, for example:

```text
usage: schway run --engine=interpreter SOURCE -- PATH | run --engine=native SOURCE --manifest MANIFEST -- PATH
```

Update the expected usage string in `main_phase37_file_run_test.go` with the corrected grammar.

---

_Reviewed: 2026-10-15T18:54:26Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_

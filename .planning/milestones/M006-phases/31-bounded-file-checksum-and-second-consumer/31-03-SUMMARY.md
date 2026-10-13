---
phase: 31-bounded-file-checksum-and-second-consumer
plan: 03
subsystem: compiler/file-bytes-owner-peers
tags: [peers, seeded-faults, owner-shape, cgen, stderr-table]
status: complete
requires:
  - phase: 31-bounded-file-checksum-and-second-consumer
    provides: 31-01 third owner shape and D-16 trailing call in every peer
provides:
  - per-peer TestPhase31FileBytesSeededFaults (2 positive controls + 9 fault rows) in check, corevalidate, originvalidate, pathoracle, cgen
  - TestPhase31FileBytesNeighborRefusals (check) and TestPhase31FileErrorStderrTable (cgen)
affects: [31-04, 31-07, 31-08]
actuals:
  tokens: 30000
  tasks: 2
  commits: 2
plan_head_before: b773977c338d8c8f430eeb3a0cbc3917c8f7f1a5
plan_head_after: cfa1d77a
commits: 2
tech-stack:
  added: []
  patterns:
    - each peer forges and refuses alone; no helper shared between packages (phase31Peer prefix)
    - generic-law-first packages (corevalidate, cgen) have their own matcher called directly
key-files:
  created:
    - internal/compiler/check/check_phase31_filebytes_test.go
    - internal/compiler/corevalidate/corevalidate_phase31_filebytes_internal_test.go
    - internal/compiler/originvalidate/originvalidate_phase31_filebytes_test.go
    - internal/compiler/pathoracle/pathoracle_phase31_filebytes_test.go
    - internal/compiler/cgen/cgen_phase31_filebytes_test.go
    - internal/compiler/cgen/cgen_phase31_export_test.go
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/cgen/cgen_program.go
key-decisions:
  - "corevalidate and cgen run generic replay laws first (first-problem-only), so their seeded-fault rows assert Validate/EmitApplication refuse AND call the peer's own file-bytes matcher directly, which must refuse alone"
  - "check rows use checkFileBytesOwnerPostAssembly for core-shaped faults and mutated inline source for source-expressible faults; release-before-view-call is refused with check.release_while_view_live (the law that runs first by design)"
requirements-completed: []
---

# Phase 31 Plan 03: Independent peer refusals Summary

Every peer (check, corevalidate, originvalidate, pathoracle, cgen) now refuses all nine seeded file-bytes faults alone, accepts the 5- and 6-operation controls, and cgen's FileError stderr table is shown to derive from `program.DataTypes`.

## Accomplishments

- 55 passing `TestPhase31FileBytesSeededFaults/*` subtests (5 packages x 11); the Phase 23/29, peer and loan-liveness tests and the whole `./internal/compiler/...` suite stay green.
- `TestPhase31FileBytesNeighborRefusals`: second trailing call, trailing call taking the view, trailing callee returning Unit all give `check.file_bytes_owner_shape_unsupported`.
- `TestPhase31FileErrorStderrTable`: exactly three `FileError.` literals, each in its `status == N` branch with newline write and `exit(65)`; the reversed-alternatives forge puts `AllocFailed` at status 1.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] originvalidate accepted a trailing callee that is not U64 -> U64**
- **Found during:** Task 2 (trailing-callee-not-u64)
- **Issue:** the function-only matcher keyed the callee by entry-module ID prefix and never inspected its types.
- **Fix:** program-level `fileBytesTrailingCalleeProblem` in `ValidatePublished`, code `core.file_bytes_owner_origin`.
- **Files modified:** internal/compiler/originvalidate/originvalidate.go
- **Commit:** 3cc5bb5b

**2. [Rule 1 - Bug] cgen's file-bytes matcher accepted any trailing callee**
- **Issue:** `validateFileBytesOwnerFunction` sees one function, so callee types and module were never checked in cgen.
- **Fix:** `validateFileBytesTrailingCallee` called from `checkFileBytesOwnerProgram`.
- **Files modified:** internal/compiler/cgen/cgen_program.go
- **Commit:** 3cc5bb5b

**3. [Rule 3 - Blocking] cgen matcher unreachable from cgen_test**
- `EmitApplication` validates the core first, so generic codes pre-empt cgen's matcher. Added a test-only seam `cgen_phase31_export_test.go` (not in the plan's file list) so cgen's own matcher is exercised alone.

### Process note
The tracer task and the full table were committed together as one test commit (plus one fix commit) rather than two task commits, because the fixes surfaced only when the full table ran.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

All six test files exist; commits 3cc5bb5b and cfa1d77a are in HEAD.

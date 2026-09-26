---
phase: 16-branch-match-emitter-port
plan: 14
subsystem: session native-emission boundary
tags: [native, m004, provenance, tests]
requires: [16-13]
provides: [terminal-public-m004-refusal, test-only-frozen-evidence]
affects: [internal/compiler/session, testdata/phase16]
tech-stack:
  added: []
  patterns: [canonical-program-digest, refusal-first-evidence]
key-files:
  created:
    - internal/compiler/session/session_phase16_control_test.go
    - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
  modified:
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_corpus_test.go
    - internal/compiler/session/session_test.go
    - testdata/phase16/file-frozen-evidence.json
decisions:
  - Phase16ControlNativeC delegates only to cgen.EmitNative; a fixture cannot authorize historical C in production.
  - File frozen evidence is test-only and binds the supplied Program's canonical bytes before reading artifacts.
metrics:
  tasks: 3
  status: complete
status: complete
plan_head_before: 2022f8b2b1d5cc5228cdcb01e0885058fec5d2f9
commits: 3
actuals:
  tasks: 3
  commits: 3
---

# Phase 16 Plan 14: Terminal M004 Session Boundary Summary

Production session native-C requests now preserve cgen's terminal M004 refusal, while historical file artifacts remain runnable only through a test-only loader tied to the exact rechecked program.

## Completed Tasks

1. Replaced the session fallback with direct `cgen.EmitNative` delegation and pinned admitted/refusal behavior.
2. Added `ProgramSHA256` provenance to every file-backed frozen-evidence record and migrated external historical controls to a test-only canonical-program loader.
3. Added a source-derived guard that rejects frozen-evidence selectors in non-test session sources.

## Verification

- Passed: `go test ./internal/compiler/session -run '^(TestPhase16ControlNativeCAdmittedUsesPublicEmitter|TestPhase16ControlNativeCPreservesM004Refusal)$' -count=1 -v`
- Passed: `go test ./internal/compiler/session -run '^(TestPhase16FileFrozenEvidenceBindsCanonicalProgram|TestPhase16FileFrozenEvidenceRejectsProgramSubstitution|TestPhase16M004CorpusRefusal|TestNonlocalExitProbeInterpreterNative)$' -count=1 -v`
- Passed: `go test ./internal/compiler/session -run '^(TestPhase16ProductionSourcesCannotLoadFrozenC|TestPhase16ControlNativeCAdmittedUsesPublicEmitter|TestPhase16ControlNativeCPreservesM004Refusal)$' -count=1 -v`

## Deviations from Plan

### Auto-fixed Issues

1. [Rule 1 - Bug] The historical-C fallback was compiled into production session code.
- **Fix:** Removed all fallback routing and evidence-file readers from the production boundary.

## Deferred Issues

`go test ./internal/compiler/session -run '^TestPhase16' -count=1` remains red on the pre-existing public-emitter inventory drift: source=77, registry=80. This plan deliberately does not alter the separately owned registry.

## Self-Check: PASSED

All task commits and created test files exist.

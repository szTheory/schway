---
phase: 16-branch-match-emitter-port
plan: "13"
subsystem: compiler-native-evidence
tags: [native, schema2, m004, validation, provenance]
requires: [16-11, 16-12]
provides: [fail-closed-m004-witness-provenance, reproducible-phase16-validation]
affects: [phase-16-verification, m004-emitter-debt]
tech-stack:
  added: []
  patterns: [refusal-first-frozen-provenance, dynamic-schema2-admission, mutation-killed-validation-chain]
key-files:
  created: [testdata/phase16/validation-corpus-run-record.jsonl, testdata/phase16/validation-corpus-run-record.manifest.json]
  modified: [internal/compiler/session/witness_registry_test.go, internal/compiler/session/evidence_grade_test.go, .planning/phases/16-branch-match-emitter-port/16-VALIDATION.md]
decisions:
  - "Admitted controls retain dynamic schema-2 comparison; frozen artifacts are usable only after the current public M004 refusal."
  - "The public consumer inventory, witness probe, legacy manifest, and file provenance are one fail-closed evidence chain."
metrics:
  duration: "~3h"
  completed: 2026-09-21
  tasks: 2
  commits: 10
  plan_head_before: 3cc1bd4e6665f6c39ce557cfd9261b9c2118d7a8
actuals:
  tokens: 546417
  tasks: 2
  commits: 10
requirements-completed: [NAT-08, NAT-09]
status: complete
---

# Phase 16 Plan 13: Evidence, Witness, and Full-Suite Validation Summary

Phase 16 now distinguishes dynamic schema-2 admission from explicit M004 refusal-first frozen evidence, with mutation-killed registry/probe/provenance linkage and reproducible gate records.

## Accomplishments

- Restored dynamic schema-2 evidence for admitted core, native, quality, and session controls without reopening any legacy emitter dispatch.
- Bound all cut controls to current public refusal, digest-bound frozen artifacts, and the named `probe:TestPhase16M004CorpusRefusal` witness.
- Added a quality-control mutation suite that rejects missing citations, stale witnesses, changed file provenance, and a representative M004 classification drift.
- Published dated focused and full package-gate records, retaining cutover `0607486` and the explicit M004-only boundary.

## Task Commits

1. **Task 1: Link one quality/control exception through registry, witness, and frozen provenance** — `0cbfe0b`, `9a1de4c`, `8733e5b`, `c9e587a`, `68aa421`, `834e038`, `1401a25`, `5344bf7`, `6827f8f`.
2. **Task 2: Publish complete Phase 16 validation and run the final package gate** — `1218c57`.

## Verification

- `go test ./internal/compiler/session -run '^TestPhase16M004ProvenanceRegistryRejectsFaults$' -count=1` — PASS (RED compile failure observed before the helper was added, then GREEN).
- `go test ./internal/compiler/core ./internal/compiler/session -run 'Test.*(QLT|Admission|Payload|Witness|EmitterInventory|PreviousPhaseCore)' -count=1` — PASS.
- `go test ./internal/compiler/cgen -run 'Test(LegacyEmitterEvidence|FileFrozenEvidenceRejectsFaults|GeneratedFrozenEvidenceRejectsFaults)' -count=1` — PASS.
- `go test ./internal/compiler/session -run 'Test(ValidationRowGradesAreEarnedOverArchivedCorpus|Phase16M004ProvenanceRegistryRejectsFaults)$' -count=1` — PASS.
- `go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/native ./internal/compiler/session -count=1` — PASS; cgen, core, and native reported PASS directly before the session portion completed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Cross-checked the M004 consumer registry with both current refusal evidence and frozen file provenance.**
- **Found during:** Task 1.
- **Issue:** The registry verified a named probe independently, but did not make a representative M004 classification fail when its frozen provenance chain drifted.
- **Fix:** Added a shared in-memory validator and mutation cases for missing/stale citation, changed digest reference, and classification drift.
- **Files modified:** `internal/compiler/session/witness_registry_test.go`.
- **Commit:** `6827f8f`.

## Issues Encountered

- The final package command's session package runs substantially longer than cgen/core/native; it completed after those packages reported PASS. No legacy dispatch was restored and M004 was not broadened.

## Next Phase Readiness

The Phase 16 evidence boundary is explicit: admitted fixtures have dynamic schema-2 evidence, while M004 remains public-refusal-only debt backed by immutable provenance.

## Self-Check: PASSED

- Task commits `0cbfe0b`, `9a1de4c`, `8733e5b`, `c9e587a`, `68aa421`, `834e038`, `1401a25`, `5344bf7`, `6827f8f`, and `1218c57` exist.
- `internal/compiler/session/witness_registry_test.go`, `testdata/phase16/public-emitter-consumers.json`, and `16-VALIDATION.md` exist.

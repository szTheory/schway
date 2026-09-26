---
phase: 16-branch-match-emitter-port
plan: "11"
subsystem: compiler-native-evidence
tags: [native, cgen, frozen-artifacts, provenance, m004]
requires: [16-10]
provides: [fail-closed-legacy-emitter-registry]
affects: [16-12, 16-13]
tech-stack:
  added: []
  patterns: [digest-bound-artifact-bijection, refusal-first-frozen-evidence]
key-files:
  created: []
  modified: [internal/compiler/cgen/legacy_emitter_evidence_test.go]
decisions:
  - "Historical C is admissible only as a digest-bound artifact paired one-to-one with a current public refusal."
  - "Foreign and by-pointer M004 refusal identities are validated from their evidence family, not merely recorded as arbitrary strings."
metrics:
  duration: "~15m"
  completed: 2026-09-20
  tasks: 2
  commits: 2
  plan_head_before: 3852ea4906513ef618a0d7a12896fd072ba34408
actuals:
  tokens: 2437
  tasks: 2
  commits: 2
status: complete
---

# Phase 16 Plan 11: Native and Cgen Frozen Artifact Evidence Summary

The cut native/Cgen consumers retain their current public M004 refusals while historical C evidence is accepted only through a provenance-complete, digest-checked registry.

## Completed Tasks

1. Validated the existing refusal-first native/Cgen cut controls against a complete artifact/evidence registry: matching schemas and cut commit, unique fixture and artifact identities, artifact and fixture SHA-256 values, witness presence, exact artifact-to-fixture pairing, and the family-specific M004 refusal identity.
2. Added fail-closed mutation coverage for duplicate or omitted entries, stale digest, swapped artifact, changed fixture identity, changed refusal code, and altered frozen artifact bytes.

## Verification

- `go test ./internal/compiler/cgen -run 'TestLegacyEmitterEvidence$' -count=1` — passed.
- `go test ./internal/compiler/native -run 'Test.*(Foreign|Retained|LTO|Symbol)' -count=1` — passed.
- `go test ./internal/compiler/cgen -run 'TestLegacyEmitterEvidence' -count=1` — passed.
- `go test ./internal/compiler/cgen ./internal/compiler/native -count=1` — native passed; cgen failed in pre-existing `TestNativeFunctionCalledPreorder`, `TestNativeFunctionCalledProjectionRemoval`, and `TestEmitProgramEndToEndAgreesWithInterpreterAtO0`. Each reports a schema-2 invocation projection mismatch in `cgen_program_test.go`; this plan only changes the evidence test and does not alter the compared emitter or session projection.
- `git diff --check` — passed before each task commit.

## Decisions Made

- A non-empty registry field is insufficient evidence: provenance includes cut revision, fixture content, immutable artifact bytes, a named witness, and the refusal the public API currently returns.
- The five existing historical consumers remain frozen evidence; no alternative public lowering path is introduced for their cut-M004 inputs.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Made the artifact/evidence registry fail closed.**
- **Found during:** Task 1
- **Issue:** The existing test checked artifact hashes independently but did not verify that every evidence record was unique, provenance-complete, paired with its own artifact, or tied to its declared M004 refusal.
- **Fix:** Added a registry validator and targeted mutation controls.
- **Files modified:** `internal/compiler/cgen/legacy_emitter_evidence_test.go`
- **Commits:** `a9a5d6f`, `68d180c`

## Deferred Issues

- The package-wide Cgen gate has three pre-existing schema-2 invocation-projection failures described in Verification. They are outside the Plan 16-11 evidence boundary and were not changed.

## Self-Check: PASSED

- Task commits `a9a5d6f` and `68d180c` exist.
- `internal/compiler/cgen/legacy_emitter_evidence_test.go` and both Phase 16 registry files exist.

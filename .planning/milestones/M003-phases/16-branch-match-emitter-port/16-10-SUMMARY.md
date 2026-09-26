---
phase: 16-branch-match-emitter-port
plan: "10"
subsystem: compiler-session-evidence
tags: [schema-2, differential, emitter-inventory, m004]
requires: [16-09]
provides: [independent-schema2-projection, public-emitter-consumer-registry]
affects: [16-11, 16-12, 16-13]
tech-stack:
  added: []
  patterns: [checked-fact-projection, fail-closed-ast-inventory, named-refusal-witness]
key-files:
  created: [internal/compiler/session/session_phase16_emitter_inventory_test.go]
  modified: [internal/compiler/session/session.go, internal/compiler/session/session_phase5_corpus_test.go, internal/compiler/session/session_phase11_differential_test.go, internal/compiler/session/witness_registry_test.go, testdata/phase16/public-emitter-consumers.json]
decisions:
  - "Schema-2 expectations never copy invocation or defect facts from native output."
  - "Every direct public cgen emitter call receives exactly one checked disposition."
metrics:
  duration: "~20m"
  completed: 2026-09-20
  tasks: 2
  commits: 2
  plan_head_before: 140403e
actuals:
  tokens: 8000
  tasks: 2
  commits: 2
status: complete
---

# Phase 16 Plan 10: Schema-2 Projection and Emitter Inventory Summary

Schema-2 comparison expected evidence now comes from checked program facts and interpreter output, while a 95-entry source-derived registry fail-closes every direct public emitter consumer.

## Completed Tasks

1. Removed the native-to-expected invocation copy in `RunNative`, derives defect reasons from checked core operations, and added mutation coverage for native invocation/reason changes. `acquire_three_success.lang` now follows a named refusal-only corpus witness.
2. Replaced the count-only inventory with a bijective `/2` registry. The AST scan handles aliases and same-package calls, rejects cgen dot imports, excludes non-cgen local shadows, and rejects stale, duplicate, missing, or unclassified records. Refusal records require a `probe:` witness.

## Verification

- `go test ./internal/compiler/session -run 'TestPhase5Corpus|TestPhase11Differential|TestPhase16Schema2Projection' -count=1` — passed.
- `go test ./internal/compiler/session -run 'TestPhase16.*Emitter.*Inventory|TestWitnessRegistry|TestPhase16EmitterInventoryRefusalWitnessesResolve' -count=1` — passed.
- `git diff --check` — passed.

## Decisions Made

- Expected comparison data is not repaired from observed native documents; a changed invocation or defect reason remains a mismatch.
- The registry records all direct public consumers now; Plans 16-11 through 16-13 extend refusal rows with their frozen-artifact provenance migration.

## Deviations from Plan

None - plan executed exactly as written.

## Self-Check: PASSED

- Task commits `20227c6` and `09703a4` exist.
- All six planned production/test/registry files exist and are represented above.

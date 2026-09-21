---
phase: 16-branch-match-emitter-port
plan: 15
subsystem: session public-emitter inventory
tags: [native, cgen, provenance, tests]
requires: [16-14]
provides: [bijective-public-emitter-registry, mutation-killed-inventory-gate]
affects: [internal/compiler/session, testdata/phase16]
tech-stack:
  added: []
  patterns: [AST-derived-call-inventory, pure-bijection-validator, in-memory-negative-controls]
key-files:
  created: []
  modified:
    - internal/compiler/session/session_phase16_emitter_inventory_test.go
    - testdata/phase16/public-emitter-consumers.json
decisions:
  - The AST-derived source inventory is authoritative; registry rows must classify it exactly once.
  - Mutation controls assert their intended validator error class rather than accepting generic cardinality drift.
metrics:
  tasks: 2
  status: complete
status: complete
plan_head_before: f7e6c6a5d9c1b62b2bf7c035edb2333a798d7310
commits: 3
actuals:
  tokens: 2353
  tasks: 2
  commits: 3
---

# Phase 16 Plan 15: Public Emitter Registry Reconciliation Summary

The public-emitter registry now exactly mirrors the current AST-derived direct-call inventory, and shared validation rejects each seeded form of registry drift with an explicit error class.

## Completed Tasks

1. Refactored source-to-registry comparison into a deterministic pure validator and reconciled the canonical registry to 77 current calls.
2. Added independent in-memory controls for duplicate, stale, missing, invalid-classification, and missing-refusal-witness rows.

## Verification

- Passed: `go test ./internal/compiler/session -run '^TestPhase16PublicEmitterConsumerInventory$' -count=1 -v`
- Passed: `go test ./internal/compiler/session -run '^(TestPhase16PublicEmitterConsumerInventory|TestPhase16EmitterInventoryMutationControls)$' -count=1 -v`
- Package gate run: `go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/native ./internal/compiler/session -count=1`
  - `cgen` and `core` passed.
  - `native` remains red in pre-existing `TestShippedBinaryFourSubcommandCorpusMatrix`: Phase 4 fixtures return `native.tool_failure` for native execution. This plan changes only the session inventory test and its JSON registry, so the unrelated native corpus failure was not changed.

## Deviations from Plan

### Auto-fixed Issues

1. [Rule 1 - Bug] The first mutation-control draft left an unused fixture variable.
- **Found during:** Task 2 focused test run
- **Fix:** Removed the unused variable before rerunning the focused suite.
- **Files modified:** `internal/compiler/session/session_phase16_emitter_inventory_test.go`
- **Commit:** `dc3fb3a`

## Deferred Issues

The phase package gate is blocked by the unrelated native corpus failure described above.

## Self-Check: PASSED

The reconciled registry, shared validator, mutation controls, and all three task commits exist.

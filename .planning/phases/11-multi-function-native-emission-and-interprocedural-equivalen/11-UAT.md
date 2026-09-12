---
status: testing
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
source: [11-VERIFICATION.md]
started: 2026-09-12T10:19:31Z
updated: 2026-09-12T10:19:31Z
---

## Current Test

number: 1
name: Decide on WR-01 — reduce.Reduce trusts Seed.EntryFunctionID without validating it
expected: |
  Either accept the current state (the single production caller,
  session_phase5_mismatch.go:386-390, derives the ID correctly via
  callgraph.EntryFunction) as sufficient for M002, or require a follow-up plan
  implementing 11-REVIEW.md's suggested Seed.Validate() fail-closed guard before
  QLT-05 is treated as hardened against misuse rather than only against its one
  current caller.
awaiting: user response

## Tests

### 1. Decide on WR-01 — reduce.Reduce trusts Seed.EntryFunctionID without validating it
expected: Accept the residual risk for M002, or require a follow-up plan adding a fail-closed Seed.Validate() guard. Reproduce with: go test ./internal/compiler/reduce/... -run TestReduceRejectsMultiFunctionSeed -v — then inspect reduce.go:202-254 and :320-347, where dropOrphanFunction relies on EntryFunctionID to exempt the real entry point from deletion.
result: [pending]

### 2. Decide on the mid-phase gate's disclosed CLI-check divergence
expected: Decide whether the pre-existing check-command vs session.Check split needs its own tracked debt item, independent of Phase 11. Reproduce with: go run ./cmd/lang --json check testdata/phase11/multi_function_gate_corpus.lang (reports status:invalid / core.origin_omitted) versus session.Check + corevalidate.Validate, which accept the identical program cleanly. The same divergence reproduces on the already-shipped Phase 5 fixture testdata/phase5/restrict_borrow.lang, confirming it predates Phase 11.
result: [pending]

## Summary

total: 2
passed: 0
issues: 0
pending: 2
skipped: 0
blocked: 0

## Gaps

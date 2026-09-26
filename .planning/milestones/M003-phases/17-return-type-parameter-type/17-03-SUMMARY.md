---
phase: 17-return-type-parameter-type
plan: 03
subsystem: corevalidate
tags: [independent-derivation, type-facts, mutation]
requires: [{phase: 17, provides: directional checker contracts}]
provides: [independent core validator directional return derivation]
affects: [originvalidate, peer-agreement]
actuals: {tokens: 0, tasks: 2, commits: 1}
tech-stack: {added: [], patterns: [independent-return-ability-lookup]}
key-files:
  created: []
  modified: [internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_test.go]
key-decisions: ["Core validation derives return facts locally rather than consuming checker output."]
requirements-completed: [TYP-04]
coverage:
  - id: D1
    description: "Core validator independently derives and mutation-kills directional return facts."
    requirement: TYP-04
    verification: [{kind: unit, ref: "go test ./internal/compiler/corevalidate -count=1", status: pass}]
    human_judgment: false
duration: 0 min
completed: 2026-09-22
status: complete
---

# Phase 17 Plan 03: Independent Core Validation Summary

**Core validation now derives parameter and return ownership facts independently and rejects directional contract drift.**

## Task Commits

1. `210f11f` — independent core peer return facts.

## Verification

- `go test ./internal/compiler/corevalidate -count=1` — passed.

## Deviations from Plan

None.

## Next Phase Readiness

Origin validation can complete the third independent derivation.

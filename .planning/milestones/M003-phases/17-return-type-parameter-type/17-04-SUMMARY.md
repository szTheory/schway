---
phase: 17-return-type-parameter-type
plan: 04
subsystem: compiler validation
tags: [go, ownership, originvalidate, interface-contracts, mutation-testing]
requires:
  - phase: 17-01
    provides: canonical Resource-to-Result source tracer
provides:
  - independently-derived parameter and return ownership facts in originvalidate
  - return-only mutation seam and structural peer-boundary controls
affects: [17-05 agreement gate, session directional-contract evidence]
actuals:
  tokens: 2962
  tasks: 2
  commits: 4
tech-stack:
  added: []
  patterns: [separate local type-fact lookups, fact-free idempotent test fault]
key-files:
  created: []
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
key-decisions:
  - "Origin validation looks up parameter type:0 and return type:1 independently."
  - "The exported fault control carries no derived fact and only changes the return lookup."
patterns-established:
  - "Independent peers prove directional contracts with a return-only mutation and synthetic import control."
requirements-completed: [TYP-04]
coverage:
  - id: D1
    description: Origin peer independently publishes directional parameter and return contracts.
    requirement: TYP-04
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/originvalidate -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Return-only fault and import boundary are observable and mutation-killed.
    requirement: TYP-04
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestPhase17OriginPeerReturnOnlyMutation"
        status: pass
    human_judgment: false
duration: 10min
completed: 2026-09-22
status: complete
---

# Phase 17 Plan 04 Summary

**Origin validation now derives parameter `Drops` and return `Fresh` from separate local type-fact lookups.**

## Accomplishments

- Split the origin peer's parameter and return ability derivations without importing another derivation peer.
- Added return-only fault injection with an idempotent restoration closure.
- Added mutation and structural-boundary tests, including a seeded prohibited import control.

## Task Commits

1. Task 1 — `5a4eeb5`, `8cca354`
2. Task 2 — `207c924`, `470202f`

## Verification

- `go test ./internal/compiler/originvalidate -run 'TestPhase17OriginPeer(ReturnOnlyMutation|IndependenceBoundary)' -count=1 -v` — pass
- `go test ./internal/compiler/originvalidate -count=1` — pass

## Deviations from Plan

None.

## Self-Check: PASSED

---
phase: 17-return-type-parameter-type
plan: 07
subsystem: repair evidence
tags: [fixtures, heldout, sha256, repair]
requires:
  - phase: 17-01
    provides: widened call-contract diagnostics
provides:
  - sealed Phase 17 derivation and held-out call-mismatch corpus
affects: [17-08 repair emission, 17-09 protocol evidence]
actuals:
  tokens: 4000
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [topology-distinct held-out corpus, exhaustive byte seal]
key-files:
  created:
    - testdata/phase17/derivation_call_argument_mismatch.lang
    - testdata/phase17/heldout_call_argument_mismatch.lang
    - testdata/phase17/HELDOUT.sha256
  modified:
    - internal/compiler/session/session_phase17_test.go
key-decisions:
  - "Held-out repair evidence is sealed before repair-emission work."
requirements-completed: [TYP-05]
coverage:
  - id: D1
    description: Distinct reachable derivation and held-out repair fixtures.
    requirement: TYP-05
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run TestPhase17RepairCorpus -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Exhaustive held-out SHA-256 seal.
    requirement: TYP-05
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run TestPhase17Heldout -count=1"
        status: pass
    human_judgment: false
status: complete
---

# Phase 17 Plan 07 Summary

**Phase 17 now has a structurally distinct, byte-sealed held-out repair corpus before repair emission.**

## Task Commits

1. Task 1 — `caf61b3`
2. Task 2 — `4200fb8`

## Verification

- Focused repair-corpus tests — pass.
- Focused seal tests — pass.
- Full `./internal/compiler/session` suite has existing unrelated failures in historical evidence/native-emission guards; the new focused tests pass.

## Self-Check: PASSED

---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "08"
subsystem: closure-performance
tags: [qlt-12, cache, full-suite-timing]
requires:
  - phase: 20-10
    provides: final corpus evidence and reconciled validation tree
  - phase: 20-07
    provides: exact frontier, debt cap, maturity, emitter, and skip-witness guards
provides:
  - Green unfiltered full-suite preflight on the reconciled tree
  - Three paired cold/warm full-suite timings and closure-cache reuse measurements
affects: [20-06, QLT-12]
actuals:
  tasks: 2
  preflight_seconds: 179.035
  timing_protocol_seconds: 1507.346
  suite_samples: 6
  closure_cache_reused: 112
source_revision: d842177
key-files:
  created:
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-CLOSURE-TIMING.md
  modified:
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-VALIDATION.md
    - .planning/phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-08-PLAN.md
key-decisions:
  - "Use GOMAXPROCS=4 consistently to keep native subprocess tests from timing out under package contention."
  - "The sandbox denied sysctl; the CPU model Apple M5 Pro was confirmed by a direct host query and supplied through a temporary shim during timing. The resulting machine ID exactly matches the Phase 14 baseline."
  - "Retain the high 466.305-second cold sample; report three-sample spread and avoid percentile claims."
verification:
  - "GOMAXPROCS=4 GOCACHE=/private/tmp/phase20-gocache go test ./... -count=1 — passed in 179.035 seconds before timing."
  - "GOMAXPROCS=4 GOCACHE=/private/tmp/phase20-gocache go run scripts/phase20-closure-timing.go — passed; six full-suite samples and the closure cold/warm pair completed."
  - "go test ./scripts -run '^TestPhase20ClosureTimingHarness$' -count=1 — passed."
  - "Session reconciliation, frontier, and archived-grade checks — passed after report cleanup."
status: complete
completed: 2026-09-25
---

# Phase 20 Plan 08: Measure Closure Cache Timing

The unfiltered preflight passed on revision `d842177` with Go 1.24.0, Apple Clang 21, darwin/arm64, and `GOMAXPROCS=4`. It took 179.035 seconds. The timing helper then completed six full-suite runs and the separate enumerated-closure pair in 1507.346 seconds. The real host machine identity was independently confirmed as `machine:4797d76b7863`, matching Phase 14.

## Results

- Cold full-suite samples: 152.363 s, 129.425 s, 466.305 s; median 152.363 s.
- Warm full-suite samples: 120.374 s, 139.389 s, 381.854 s; median 139.389 s.
- Paired cold-minus-warm deltas: 31.990 s, -9.963 s, 84.451 s.
- Warm median is below the paired cold median and both Phase 14 references (192.7 s and 191.89 s).
- Enumerated closure cache: cold 112 recomputed; warm 112 reused.

The three samples vary widely, including one slow pair. The result meets the plan’s median comparison, but it does not establish a high-confidence percentile or a stable per-run speedup.

## Verification

- The timing report retains all six suite sample values, closure-cache counts, host/toolchain identity, and raw subprocess outputs.
- The timing harness test passed.
- Reconciliation, pinned frontier, empty violation classes, and archived validation-grade checks passed after the report’s prewarm command was removed from the claim table and its success/duration recorded as prewarm metadata.

## Next

Plan 06 is the final human checkpoint. Prepare the D-13-34 replacement alternatives from the measured frontier and phase evidence, then stop for the user to choose.

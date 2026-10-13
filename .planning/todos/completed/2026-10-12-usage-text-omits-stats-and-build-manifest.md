---
created: 2026-10-12T00:00:00Z
title: usage text omits the stats route and build --manifest
area: cli
severity: minor
files:
  - cmd/schway/main.go
resolved: 2026-10-12 (Phase 32.1, plan 32.1-07)
---

## Problem

`usageResult()` in `cmd/schway/main.go` lists the public routes, but two routes
that `run()` dispatches are absent from it:

- `schway stats` (reads decimal values from stdin and prints statistics).
- `schway build SOURCE --output ARTIFACT --manifest MANIFEST` (and the
  `--manifest MANIFEST --output ARTIFACT` order). The usage text shows only
  `build SOURCE --output ARTIFACT`.

A user reading the usage text cannot discover either route. The built-artifact
manifest route is the one the final receipts rely on, so omitting it from the
usage text is a documentation gap on a public surface.

## Notes

Phase 32 audited this and did not change the CLI (D-10d).
`TestPhase32PublicRouteAudit` (`cmd/schway/main_phase32_routes_test.go`)
records the omission set in `phase32UsageOmissions` and fails if the set
changes. When the usage text is fixed, empty that set in the same change.

---
phase: quick
plan: 260924-kto
subsystem: compiler-checker
tags: [core-identity, computed-match, loan-endpoints]
requires: []
provides:
  - Computed-terminal-only Result arm return-place lowering with legacy core identities preserved.
  - Valid S-010 both-arm CFG control and strengthened neither-arm source control.
affects: [phase18, core-serialization, loan-liveness]
tech-stack:
  added: []
  patterns: [explicit computed-prefix provenance gates new lowering]
key-files:
  created: [.planning/quick/260924-kto-restore-historical-core-compatibility-an/260924-kto-VERIFICATION.md]
  modified: [internal/compiler/check/check.go, internal/compiler/check/check_phase18_test.go, internal/compiler/check/check_test.go]
decisions:
  - "Legacy matches keep their historical match-edge IDs and source-type result facts; computed terminal matches use Result-type facts."
  - "The valid both-arm endpoint control is expressed as a synthetic CFG because duplicating a source take is rejected as use-after-move."
metrics:
  duration: "~15 minutes"
  completed: 2026-09-24
status: complete
---

# Quick 260924-kto Summary

**Computed Result matches retain selected-arm value places while historical Phase 3–5 core bytes and evidence manifest IDs remain pinned.**

## Accomplishments

- Restricted Result-specific bare-arm return places, value lookup, and return facts to computed terminal matches with a non-empty prefix.
- Restored legacy edge IDs for ordinary matches, preserving all seven Phase 3–5 byte and manifest pins without changing their golden values.
- Replaced the invalid both-arm source anti-control with a valid two-successor CFG using the same pre-branch loan in each arm.
- Left the payload tracer timeout unchanged after five isolated runs passed.

## Verification

- Historical core bytes and manifest IDs: passed.
- Phase 18 Result computed-match and S-010 controls: passed.
- Isolated payload tracer, `-count=5`: passed in 5.270s.
- Full suite, `GOCACHE=/tmp/ai-lang-gocache go test ./...`: passed.
- Scoped `git diff --check`: passed.

## Deviations from Plan

**1. [Rule 1 - Bug] Preserved legacy match-edge IDs outside computed terminal matches**
- **Found during:** Task 1
- **Issue:** Phase 18 edge ID changes also altered the seven historical core and manifest pins.
- **Fix:** Kept the new pattern-based edge IDs for computed terminal matches and restored the prior ordinal edge IDs for ordinary matches. Kept legacy source-type arm facts unless the terminal match has a non-empty prefix.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** Both historical pin tests pass with the original pinned values.

**2. [Rule 2 - Missing Critical] Replaced invalid S-010 both-arm source control**
- **Found during:** Task 2
- **Issue:** A second `take view` is rejected as use-after-move and cannot demonstrate both-arm liveness.
- **Fix:** Added a synthetic two-successor CFG asserting the same loan is used in both arms, each arm receives a point endpoint, and no diverging edge endpoint exists.
- **Files modified:** `internal/compiler/check/check_phase18_test.go`, `internal/compiler/check/check_test.go`
- **Verification:** Focused Phase 18 endpoint controls pass.

## Commits

No commits created by this executor; the parent agent owns GSD commits.

## Self-Check: PASSED

Summary and verification artifacts exist. Historical pins and the full suite passed. No golden pins or tracer timeout were changed.

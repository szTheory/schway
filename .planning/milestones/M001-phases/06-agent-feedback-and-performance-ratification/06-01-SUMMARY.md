---
phase: 06-agent-feedback-and-performance-ratification
plan: 01
subsystem: testing
tags: [go-testing, reflection, ast-parsing, evidence, content-identity]

# Dependency graph
requires:
  - phase: 05-native-o3-and-restrict-safety
    provides: "protocol.Result.Finalize() identity struct, Phase 5 accepting corpus and restrict_borrow.golden.c, the 12-site lang.verify-lane/0 Lane composite literals"
provides:
  - "Widened prior-phase byte pin covering Phase 1-5 core bytes, manifest IDs, and generated-C goldens"
  - "Self-invalidating executable pin that D-06-32 (no Metrics/Lane field enters Result.Finalize() identity) holds today and fails loudly the moment a new field is added without re-confirming exclusion"
  - "Pinned count (12) and per-file breakdown of the lang.verify-lane/0 schema literal sites the 06-06 coordinated bump must move all at once"
affects: [06-06-lane-schema-bump, phase-6-verification, all-later-phase-6-plans]

# Actuals (#2632)
actuals:
  tokens: 4038
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "reflect-based sentinel-field mutation to prove a field is excluded from a content-identity computation"
    - "go/parser full-mode AST literal-counting pin (vs. grep/string-count) for a coordinated multi-site literal bump"

key-files:
  created:
    - internal/compiler/session/session_phase6_pin_test.go
  modified:
    - internal/compiler/core/core_test.go
    - internal/compiler/protocol/protocol_test.go

key-decisions:
  - "All 10 non-coordinated-lie fixtures under testdata/phase5/ are accepting (zero diagnostics) under evidence.Build with pinnedFacts; all 10 were widened into pinnedFixtures rather than a curated subset, since no phase5CorpusMatrix helper exists to define an authoritative accepting subset the way native_test.go does for Phase 4."
  - "No production code changed in Task 2 -- Result.Finalize()'s identity struct already excluded Metrics entirely and reduced Lane to ID+\":\"+Status before this plan ran; the task was purely to make that fact an executable, self-invalidating pin."

requirements-completed: [FND-04, DX-03, QLT-02]

coverage:
  - id: D1
    description: "Phase 1-5 core bytes, manifest IDs, and generated-C goldens are pinned against any Phase 6 perturbation"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "D-06-32 (no Metrics/Lane field reaches Result.Finalize() identity) is an executable, self-invalidating pin"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestMetricsAndLaneFieldsExcludedFromIdentity"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestIdentityFieldEnumerationIsExhaustive"
        status: pass
    human_judgment: false
  - id: D3
    description: "The 12-site lang.verify-lane/0 literal count is pinned per-file so the 06-06 coordinated bump cannot half-land"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_pin_test.go#TestLaneSchemaLiteralSiteCountIsPinned"
        status: pass
    human_judgment: false

duration: 40min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 1: Prior-Phase Freeze and D-06-32 Identity Pin Summary

**Widened the Phase 1-5 core-byte/manifest/golden-C pin, added a self-invalidating reflection-based test proving no `protocol.Metrics`/`Lane` field reaches `Result.Finalize()`'s content identity, and pinned the 12-site `lang.verify-lane/0` literal count via `go/parser` AST traversal ahead of Phase 6's coordinated schema bump.**

## Performance

- **Duration:** 40 min
- **Started:** 2026-09-06T22:26Z (approx)
- **Completed:** 2026-09-06T23:06Z
- **Tasks:** 3
- **Files modified:** 3 (2 modified, 1 created)

## Accomplishments
- `pinnedFixtures` in `internal/compiler/core/core_test.go` widened with all 10 accepting `testdata/phase5/*.lang` fixtures (each fixture's real `evidence.Build` CoreSHA256 and ManifestID computed against the current tree), excluding the declared `coordinated_lie.lang` escape pair.
- `previousPhaseGoldenCDigests` widened with `testdata/phase5/restrict_borrow.golden.c`, and `TestPreviousPhaseGoldenCUnchanged`'s `phaseDir` loop now covers `phase5`.
- Doc comments on all three prior-phase pin tests updated to name D-06-31/D-06-32 and "before the coordinated lang.command and lang.verify-lane bump lands" as the guarded change.
- `TestMetricsAndLaneFieldsExcludedFromIdentity` added: reflection-based sentinel mutation proves every `Metrics` field, and every `Lane` field except `ID`/`Status`, leaves `Result.Finalize().ID` unchanged, with positive control assertions that `Lane.ID` and `Lane.Status` each DO change the result ID (so the test cannot pass vacuously).
- `TestIdentityFieldEnumerationIsExhaustive` added: reflects over `Metrics`/`Lane` field-name sets and fails loudly the moment a new field (e.g. 06-06's `cache_status`, `machine_id`, etc.) is added without the reader re-confirming identity exclusion.
- New `internal/compiler/session/session_phase6_pin_test.go` with `TestLaneSchemaLiteralSiteCountIsPinned`: a `go/parser` full-mode AST scan (not grep/`strings.Count`) over every non-test `.go` file in `internal/compiler/session`, pinning the exact per-file breakdown (`session.go` x8, `session_phase5.go` x1, `session_phase5_mismatch.go` x2, `session_phase5_sanitize.go` x1 = 12 total) of `"lang.verify-lane/0"` string literals.

## Task Commits

Each task was committed atomically:

1. **Task 1: Widen the prior-phase byte pin through Phase 5, end-to-end** - `1244a2b` (test)
2. **Task 2: Pin D-06-32 — Metrics and Lane fields never reach content identity** - `30dea3c` (test)
3. **Task 3: Pin the lane-schema literal site count before the coordinated bump** - `0e7f17d` (test)

**Plan metadata:** (this commit)

_Note: Tasks 2 and 3 carried `tdd="true"` but produced a single `test(...)` commit each, not a RED/GREEN pair — the pinned behavior (D-06-32's identity exclusion) already existed in `Result.Finalize()` before this plan ran, so there was no implementation step to add. This was verified explicitly per the plan's project-critical-context instruction: editing `Finalize()`'s identity struct was never in scope, only writing the pin._

## Files Created/Modified
- `internal/compiler/core/core_test.go` - Widened `pinnedFixtures` (Phase 5) and `previousPhaseGoldenCDigests` (restrict_borrow.golden.c); widened `TestPreviousPhaseGoldenCUnchanged`'s phase-directory scan to include `phase5`; updated doc comments to name D-06-31/D-06-32.
- `internal/compiler/protocol/protocol_test.go` - Added `TestMetricsAndLaneFieldsExcludedFromIdentity` and `TestIdentityFieldEnumerationIsExhaustive`, plus a shared `setSentinelFields` reflection helper.
- `internal/compiler/session/session_phase6_pin_test.go` (new) - `TestLaneSchemaLiteralSiteCountIsPinned`, an AST-based pin of the 12-site lane-schema literal count.

## Decisions Made
- All 10 non-`coordinated_lie` `testdata/phase5/*.lang` fixtures were confirmed accepting (zero diagnostics) by probing each with `evidence.Build`/`pinnedFacts` before writing the pin table, since no `phase5CorpusMatrix()`-style helper exists yet to define an authoritative accepting subset (unlike Phase 4's `native_test.go`). All 10 were pinned.
- No changes were made to `protocol.go`'s `Finalize()` — confirmed via a scripted before/after diff test (temporarily adding `RecomputedWork` to the identity struct) that the new exclusion test correctly fails when that boundary is violated, then reverted.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- The Phase 1-5 byte/manifest/golden pin, the D-06-32 identity-exclusion pin, and the 12-site lane-schema literal pin are all in place and green on the unmodified tree, and were each independently confirmed to go RED under a seeded violation matching the plan's acceptance criteria.
- 06-06 (the coordinated `lang.command`/`lang.verify-lane` schema bump) can now use `TestLaneSchemaLiteralSiteCountIsPinned` and `TestIdentityFieldEnumerationIsExhaustive` as its own completion gates.
- No blockers for subsequent Phase 6 plans.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: internal/compiler/session/session_phase6_pin_test.go
- FOUND: commit 1244a2b (Task 1)
- FOUND: commit 30dea3c (Task 2)
- FOUND: commit 0e7f17d (Task 3)

---
phase: 17-return-type-parameter-type
plan: 09
subsystem: repair protocol and blame-boundary evidence
tags: [go, repair, diagnostics, held-out, witnesses]
requires:
  - phase: 17-05
    provides: two-type semantic tracer
  - phase: 17-08
    provides: source-reachable use_matching_argument repair
provides:
  - sealed held-out repair driver evidence with protocol-field controls
  - M006-only executable boundary for unwired B1 blame
affects: [TYP-05, DX-06, D-13-02b, D-13-10a]
tech-stack:
  added: []
  patterns: [content-hash-keyed repair stand-in, byte-compared current debt disposition, exhaustive signature authority map]
key-files:
  created: []
  modified:
    - cmd/lang-repair/repair_test.go
    - cmd/lang-repair/antitheater_test.go
    - cmd/lang-repair/import_boundary_test.go
    - internal/compiler/session/witness_registry_test.go
    - .planning/UNREACHABLE-CLAIMS.md
decisions:
  - "The checked-in Phase 13 debt register remains immutable provenance; the generated current view applies the Phase 17 M006-only disposition."
  - "A signature field must be explicitly classified by declaring-function verification before it can be treated as outside the B1 boundary."
metrics:
  duration: 18 min
  completed: 2026-09-22
status: complete
actuals:
  tokens: 6457
  tasks: 2
  commits: 2
plan_head_before: abed5846215336a60f4e1b68222ea36f7a4cfb3f
---

# Phase 17 Plan 09: Sealed Repair and M006 Boundary Summary

**The sealed Phase 17 held-out mismatch now repairs through structured JSON alone, while B1 blame stays explicitly unwired until M006 separate compilation creates a declarer-unverifiable contract.**

## Accomplishments

- Ran the real repair driver on a temporary held-out copy and asserted its exact diagnosis, kind, span, replacement, repaired bytes, two driver checks, and independent clean third check.
- Added prose-scramble, stripped-field, and corrupted-field controls; production command sources are scanned for held-out fixture coupling.
- Replaced the obsolete `sameType` witness with admitted two-type evidence, a non-test `resolveBlame` call-site scan, and exhaustive `FunctionSignature` authority classification.
- Regenerated the current unreachable view: D-13-10a is absent after sealed repair evidence, and D-13-02b names only M006/separate compilation while historical Phase 13 debt bytes stay unchanged.

## Task Commits

1. **Task 1: Drive the sealed held-out repair through protocol-only subprocesses** — `a9d9d21` (`test`)
2. **Task 2: Replace the obsolete sameType witness with the permanent M006 boundary** — `163fc8b` (`test`)

## Verification

- `go test ./cmd/lang-repair ./internal/compiler/session -run 'TestPhase17(SealedHeldoutRepairs|RepairProtocolFields|RepairFieldControls|HeldoutPathsAbsentFromProduction)|TestRepairDriverDecodesNoProseFields' -count=1 -v` — passed; the session package had no matching test in this command.
- `go test ./internal/compiler/session -run 'TestPhase17(B1RequiresUnverifiableDeclaredContract|BlameBoundaryViewIsCurrent|UseMatchingArgumentDebtIsClosed)|TestUnreachableClaimsGeneratedView' -count=1 -v` — passed.
- `go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase17(TwoTypeGeneratedC|TwoTypeFourTierDifferential|TwoTypeEmitterGuardIsNotInert)' -count=1 -v` — passed.
- `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/cgen ./internal/compiler/session ./cmd/lang-repair -count=1` — blocked by pre-existing unrelated failures in corevalidate closure digests and cgen legacy emitter evidence.
- `go test ./... -count=1` — blocked by the same unrelated cgen/corevalidate failures plus existing core historical-byte and evidence-golden mismatches.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - correctness] Added explicit corrupted-field controls to the shared JSON mutation harness.**
- **Found during:** Task 1
- **Issue:** Existing controls proved field removal but did not cover changed kind, span, and replacement values on the Phase 17 repair.
- **Fix:** Added deterministic corrupt-field modes and a bounded-splice rejection assertion.
- **Files modified:** `cmd/lang-repair/antitheater_test.go`
- **Commit:** `a9d9d21`

## Deferred Issues

- Repository-wide verification remains red outside this plan’s files: cgen legacy-emitter digest evidence, corevalidate closure-digest agreement, core historical-byte manifests, and evidence goldens. These failures were present in shared concurrent phase work and were not modified here.

## Known Stubs

None.

## Self-Check: PASSED

- Found all five modified task artifacts.
- Found task commits `a9d9d21` and `163fc8b` in repository history.
- Confirmed `PHASE-13-DEBT.md` has no diff.

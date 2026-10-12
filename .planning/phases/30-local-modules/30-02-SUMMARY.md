---
phase: 30-local-modules
plan: 02
subsystem: diagnostics-protocol
tags: [go, diagnostic, protocol, repair-driver, unit-attribution]
requires:
  - phase: 30-local-modules
    provides: 30-00 gate verdict PROCEED (omitempty keeps IDs)
provides:
  - Diagnostic/Cause/Repair Unit field (json unit, omitempty) with InUnit and EffectiveUnit
  - unit= token in the human diagnostic line
  - DebugMapEntry.Unit and debugmap.Entry.Unit fields
  - schway-repair cross-unit decline
affects: [30-03, 30-04, 30-06, 30-08]
actuals:
  tokens: 9000
  tasks: 2
  commits: 2
plan_head_before: b917b37201a066485f5bc86aa54bd35b135957cd
plan_head_after: ddc7cc7c8aed03b1b82b88250ceaa6bc1c0d07d3
tech-stack:
  added: []
  patterns: ["omitempty unit attribution keeps pre-phase bytes and IDs", "unexported negative-control seams (unitOutOfIdentityForTest, alwaysPrintUnitForTest, ignoreUnitForTest)"]
key-files:
  created:
    - internal/compiler/diagnostic/diagnostic_phase30_test.go
    - internal/compiler/protocol/protocol_phase30_test.go
    - internal/compiler/protocol/export_test.go
    - cmd/schway-repair/repair_phase30_test.go
  modified:
    - internal/compiler/diagnostic/diagnostic.go
    - internal/compiler/diagnostic/diagnostic_test.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/debugmap/debugmap.go
    - cmd/schway-repair/repair.go
key-decisions:
  - "D-11/D-23 applied as recorded: Unit is the dotted module name, spans stay file-local, empty means entry module"
  - "Repair.Unit is non-identity-bearing; the diagnostic Unit carries identity"
requirements-completed: [MOD-05]
coverage:
  - id: D1
    description: "Unit-bearing diagnostics get distinct IDs; empty-unit IDs and JSON are byte-identical"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/diagnostic#TestPhase30UnitIdentity"
        status: pass
      - kind: unit
        ref: "internal/compiler/diagnostic#TestPhase30UnitOmitempty"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase30PreChangeGolden"
        status: pass
    human_judgment: false
  - id: D2
    description: "Human line prints unit= only when set; JSON carries unit only when set"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/protocol#TestPhase30UnitProjection"
        status: pass
    human_judgment: false
  - id: D3
    description: "Repair driver declines a repair attributed to another unit and leaves the entry file unchanged"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "cmd/schway-repair#TestPhase30RepairDriverDeclinesCrossUnit"
        status: pass
    human_judgment: false
duration: 8min
completed: 2026-10-12
status: complete
---

# Phase 30 Plan 02: Unit attribution through the diagnostic protocol Summary

**Unit-attributed diagnostics: `unit` in the ID identity, JSON and human line, plus a repair driver that never splices a cross-unit repair into the entry file.**

## Performance

- **Duration:** about 8 min
- **Started:** 2026-10-12T15:09:35Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments
- `Diagnostic`, `Cause` and `Repair` carry `Unit` (`json:"unit,omitempty"`). Both anonymous identity structs carry it too, with `json:",omitempty"`. The pinned ID `diagnostic:d34396fcc8ea6530a1f04dc3` is unchanged.
- `Diagnostic.InUnit` recomputes the ID for schema /0 and /1. `EffectiveUnit` on Cause and Repair gives the inheritance rule.
- The human line is `ID CODE unit=<module> [s:e]: message` when a unit is set, and byte-identical to the old line when it is empty.
- `DebugMapEntry.Unit` and `debugmap.Entry.Unit` exist as fields only. Plan 30-06 sets them; plan 30-08 copies the field in the session projection.
- `schway-repair` decodes `unit` on diagnostics and repairs and declines cross-unit repairs with `repair.none_eligible`. No internal import was added, and no prose field is decoded.

## Task Commits

1. **Task 1: Tracer, unit-bearing diagnostic through ID, JSON and human line** - `6100696` (feat)
2. **Task 2: Debug-map unit field and repair driver cross-unit decline** - `ddc7cc7` (feat)

The tracer `<verify>` was re-run end to end: the diagnostic and protocol tests and the session byte-identity set all pass.

## Decisions Made
- `classifyDecline` needed no change. A document whose only repairs are cross-unit already classifies as `none_eligible`, because repairs exist but none pass `selectRepair`. `TestPhase30RepairDriverDeclinesCrossUnit` asserts that outcome.
- `DebugMapEntry.Unit` in `protocol.go` landed in the Task 1 commit, not Task 2. It sits next to the projection test and is purely additive.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added `internal/compiler/protocol/export_test.go`**
- **Found during:** Task 1
- **Issue:** The plan puts the test in `package protocol_test` but the seam `alwaysPrintUnitForTest` is unexported, so the external test package could not arm the control.
- **Fix:** A small test-only export setter, `SetAlwaysPrintUnitForTest`, in package protocol.
- **Files modified:** `internal/compiler/protocol/export_test.go`
- **Committed in:** `6100696`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Test-only addition, no production behavior change.

## Issues Encountered
None. `gofmt -l` lists several unrelated pre-existing files (corevalidate, evidence, measure, pathoracle, reduce, session tests). Those are out of scope and untouched.

## Registry gap for a later plan
30-01's new codes and repair kinds are not registered in diagnostic or protocol, because this plan's scope does not cover registries. They are `syntax.expected_import_name`, `syntax.import_after_export`, `module.duplicate_import`, `module.imports_limit`, `move_import_above_export` and `delete_duplicate_import`. A later plan that owns code and repair-kind registries must register them. A grep of `internal/compiler/diagnostic` and `internal/compiler/protocol` finds no registry entries for them today.

## Known Stubs
None. `debugmap.Entry.Unit` and `DebugMapEntry.Unit` are intentionally unset until plans 30-06 and 30-08.

## Threat Flags
None beyond the plan's threat model. T-30-07, T-30-08 and T-30-09 are mitigated and test-pinned, each with an armed negative control.

## Next Phase Readiness
Plans 30-03 and later can stamp `Diagnostic.InUnit(module)`. The wire shape is fixed.

## Self-Check: PASSED
All five modified and four created files exist, and commits `6100696` and `ddc7cc7` are in HEAD. The session meta set (Inventory, Groundedness, ValidationLifecycle, PreChangeGolden, corpus byte identity) passes.

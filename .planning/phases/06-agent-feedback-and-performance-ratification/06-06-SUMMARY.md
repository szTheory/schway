---
phase: 06-agent-feedback-and-performance-ratification
plan: 06
subsystem: protocol-schema
tags: [schema-versioning, protocol, additive-migration, identity-exclusion, vocabulary-closure]

# Dependency graph
requires:
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-01"
    provides: "TestLaneSchemaLiteralSiteCountIsPinned (the completeness gate this plan's coordinated bump must satisfy) and TestIdentityFieldEnumerationIsExhaustive/TestMetricsAndLaneFieldsExcludedFromIdentity (D-06-32's identity-exclusion pin)"
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-04"
    provides: "cache.CacheStatuses() -- the four-value cache_status vocabulary protocol.LaneCacheStatuses() draws from directly"
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-05"
    provides: "session.ReasonSelected/ReasonDeferred/ReasonWidened field-shape precedent for protocol.LaneSelectionReasons()"
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-08"
    provides: "measure.Verdicts()/measure.MachineID/ColdOrWarm field-shape precedent for protocol.LaneGateVerdicts()/LaneColdOrWarmValues()"
provides:
  - "protocol.Schema1 (\"lang.command/1\") and protocol.LaneSchema1 (\"lang.verify-lane/1\") -- the coordinated additive bump, landed in one commit; protocol.Schema/LaneSchema (\"/0\") remain declared and frozen"
  - "All 12 lang.verify-lane/0 composite-literal sites across session.go/session_phase5.go/session_phase5_mismatch.go/session_phase5_sanitize.go now reference protocol.LaneSchema1"
  - "protocol.Lane's seven /1 reporting fields (CacheStatus, SelectionReason, MachineID, GateVerdict, ColdOrWarm, StageBreakdown) and protocol.Metrics.CacheInputsReusedCount -- all omitempty, additive, vocabulary-closed"
  - "protocol.LaneCacheStatuses/LaneSelectionReasons/LaneGateVerdicts/LaneColdOrWarmValues/ValidateLaneVocabularies -- the closed-vocabulary enforcement surface"
  - "TestPhase1EvidenceBytesUnchangedAfterBump and TestSchemaZeroConstantsStillExist -- executable proof the /0 record is frozen and reproducible"
affects: ["06-07 (verify wiring consumes these fields)", "06-09 (populates GateVerdict/ColdOrWarm/StageBreakdown)", "phase-6-verification"]

# Actuals (#2632)
actuals:
  tokens: 9200
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Two-constant schema coexistence (Schema/Schema1, LaneSchema/LaneSchema1), copying diagnostic.go's and evidence.go's own /0-/1 shape"
    - "AST identifier-reference counting (ast.SelectorExpr) replacing ast.BasicLit string-value counting once literal strings became named constant references"
    - "Recursive reflect-based sentinel-field setter generalized to slice-of-struct elements, so a structured field (StageTiming) needs no bespoke test case"

key-files:
  created: []
  modified:
    - internal/compiler/protocol/protocol.go
    - internal/compiler/protocol/protocol_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_mismatch.go
    - internal/compiler/session/session_phase5_sanitize.go
    - internal/compiler/session/session_phase6_pin_test.go
    - internal/compiler/testsupport/cli_test.go

key-decisions:
  - "Task 0's checkpoint was pre-resolved at orchestration time (developer confirmed D-06-31 as locked, 2026-09-06): ONE coordinated bump, lang.command/0->/1 and lang.verify-lane/0->/1, all 12 lane-schema literal sites plus the one-line protocol.Schema constant, in a single commit, with /0 constants coexisting and /0 bytes frozen. The seven-field list (cache_status, selection_reason, machine_id, gate_verdict, cold_or_warm, stage_breakdown, cache_inputs_reused_count) was confirmed complete against 06-04/06-05/06-08's sibling SUMMARYs before committing; recorded here per the plan's instruction, not re-litigated."
  - "The 12 lane-schema sites now reference the protocol.LaneSchema1 named constant, not a repeated raw string literal -- following the plan's own explicit instruction (\"replace each of the 12 lane-schema string literals with protocol.LaneSchema1\") and matching diagnostic.go's Schema1-as-identifier convention more literally than a bare string repeat would."
  - "TestLaneSchemaLiteralSiteCountIsPinned's detection mechanism was changed from ast.BasicLit string-value scanning to ast.SelectorExpr identifier-reference scanning (protocol.LaneSchema1), because the sites moved from raw string literals to a named constant reference and the old BasicLit scan would silently report zero sites -- a false pass that would have let a half-landed bump through undetected. This is a mechanism change, not a relaxation: the pin still fails naming the exact file if any one of the 12 sites is reverted, added, or removed (independently verified by reverting one site and confirming the failure)."
  - "cache_status/selection_reason/gate_verdict/cold_or_warm treat an EMPTY value as valid (not-yet-populated), refusing only a non-empty out-of-vocabulary value -- populating these fields is explicitly 06-07's/06-09's job per the plan; ValidateLaneVocabularies must not reject the current all-empty state every existing Lane construction site produces today."
  - "protocol package now imports internal/compiler/cache (a dependency-free leaf package) to draw LaneCacheStatuses() directly from cache.CacheStatuses(), per the plan's explicit instruction to avoid a second copy of that vocabulary. Verified no import cycle: cache imports nothing beyond stdlib."

requirements-completed: []

coverage:
  - id: D1
    description: "One coordinated additive bump lands in one commit: lang.command/0 to /1 and lang.verify-lane/0 to /1, with all /0 bytes frozen byte-for-byte"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestCommandSchemaIsVersionOne"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestSchemaZeroConstantsStillExist"
        status: pass
      - kind: manual
        ref: "lang --json verify testdata/phase1 emits lang.command/1 top-level and lang.verify-lane/1 on all 5 lanes"
        status: pass
    human_judgment: false
  - id: D2
    description: "All 12 hardcoded lane-schema literal sites move together; a half-landed bump fails the site-count pin from 06-01"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_pin_test.go#TestLaneSchemaLiteralSiteCountIsPinned"
        status: pass
      - kind: unit
        ref: "manual: reverted session_phase5.go's one site to protocol.LaneSchema, confirmed the pin fails naming session_phase5.go, then restored"
        status: pass
    human_judgment: false
  - id: D3
    description: "protocol.Lane gains cache_status, selection_reason, machine_id, gate_verdict, cold_or_warm, stage_breakdown; protocol.Metrics gains cache_inputs_reused_count"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestIdentityFieldEnumerationIsExhaustive"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestNewLaneAndMetricsFieldsAreAdditive"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestLaneVocabulariesAreClosed"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestCacheInputsReusedCountSharesUnitWithRecomputedWork"
        status: pass
    human_judgment: false
  - id: D4
    description: "Every new Metrics and Lane field stays outside Result.Finalize()'s identity struct, so measurement cannot change semantic output"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestMetricsAndLaneFieldsExcludedFromIdentity"
        status: pass
    human_judgment: false
  - id: D5
    description: "Phase 1 evidence bytes remain byte-identical after the bump"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestPhase1EvidenceBytesUnchangedAfterBump"
        status: pass
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
  - id: D6
    description: "cache_status is drawn from the four-value artifact vocabulary and never uses hit or miss terminology"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestLaneVocabulariesAreClosed"
        status: pass
    human_judgment: false

duration: 65min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 6: Coordinated Schema Bump Summary

**One coordinated commit landed `lang.command/1` and `lang.verify-lane/1` across all 12 lane-schema literal sites, plus seven new additive, vocabulary-closed, identity-excluded `protocol.Lane`/`protocol.Metrics` reporting fields -- with executable proof that `/0` bytes stayed frozen.**

## Checkpoint Confirmation (Task 0)

Task 0 was a `checkpoint:decision` whose `<blocking>` was already `false` with a recorded `<resolution>` at orchestration time (developer confirmed 2026-09-06). Not re-litigated during execution; recorded here per the plan's instruction:

- **D-06-31 confirmed locked:** ONE coordinated bump -- `lang.command/0` -> `/1` and `lang.verify-lane/0` -> `/1`, all 12 lane-schema literal sites plus the one-line `protocol.Schema` constant, landing in a single commit. `/0` constants coexist (the `Schema`/`Schema1` two-constant pattern `diagnostic.go`/`evidence.go` already use) and `/0` bytes stay frozen.
- **Seven-field list confirmed complete** against the now-landed sibling plans: 06-04 (`cache_status`, `cache_inputs_reused_count` vocabulary source), 06-05 (`selection_reason` field shape), 06-08 (`machine_id`, `gate_verdict`, `cold_or_warm`, `stage_breakdown` field shapes). No field was found missing.
- `lang.explain/0` and `lang.query/0` correctly excluded from this bump (net-new documents, no coordination needed on their own account per D-06-04). `lang.diagnostic/1`'s Repair extension is a different, already-shipped schema, out of this plan's scope.

## Performance

- **Duration:** ~65 min
- **Tasks:** 3 (Task 0 confirmed, not executed as work; Task 1 tracer; Task 2 TDD; Task 3 TDD)
- **Files modified:** 8 (0 created)
- **Commits:** 3

## Accomplishments

- **Task 1 (tracer):** `protocol.Schema1` ("lang.command/1") and `protocol.LaneSchema`/`LaneSchema1` ("lang.verify-lane/0"/"/1") added following the exact two-constant coexistence shape `diagnostic.go`/`evidence.go` already use. `protocol.New()` now stamps `Schema1`. All 12 `"lang.verify-lane/0"` composite-literal sites (8 in `session.go`, 1 in `session_phase5.go`, 2 in `session_phase5_mismatch.go`, 1 in `session_phase5_sanitize.go`) moved to `protocol.LaneSchema1` in this same commit, edited in place with no closure refactor. `TestLaneSchemaLiteralSiteCountIsPinned` (06-01) was updated: since the sites now reference the named constant rather than repeating the raw string, the AST scan was changed from `ast.BasicLit` string-value counting to `ast.SelectorExpr` identifier-reference counting (`protocol.LaneSchema1`) -- the per-file expected counts (8/1/2/1=12) are unchanged, and reverting any one site was independently confirmed to fail the pin naming that exact file. Built `./cmd/lang` and confirmed `lang --json verify testdata/phase1` emits `lang.command/1` top-level and `lang.verify-lane/1` on all 5 lanes.
- **Task 2 (TDD):** `protocol.Lane` gained all six new fields (`CacheStatus`, `SelectionReason`, `MachineID`, `GateVerdict`, `ColdOrWarm`, `StageBreakdown []StageTiming`), all `omitempty`, appended after existing fields; `protocol.Metrics` gained `CacheInputsReusedCount`, `omitempty`. `LaneCacheStatuses()` draws directly from `cache.CacheStatuses()` (no second vocabulary copy); `LaneSelectionReasons()`/`LaneGateVerdicts()`/`LaneColdOrWarmValues()` declare their own closed sets (`selected`/`deferred`/`widened`; `blocking`/`observed`/`not_ratified`; `cold`/`warm`) since `protocol` cannot import `session` (would cycle). `ValidateLaneVocabularies` refuses any non-empty out-of-vocabulary value while treating empty as valid (population is 06-07's/06-09's job). `TestIdentityFieldEnumerationIsExhaustive`'s expected field-name slices were updated to name all seven new fields; `TestMetricsAndLaneFieldsExcludedFromIdentity` needed no code change since it enumerates struct fields reflectively -- the shared `setSentinelFields` helper was generalized into a recursive `setSentinelValue` that handles slice-of-struct elements, so the new `StageBreakdown []StageTiming` field is covered automatically.
- **Task 3 (TDD):** Added `TestPhase1EvidenceBytesUnchangedAfterBump`: a Result built with the frozen `protocol.Schema` ("/0") string for a fixed deterministic diagnostic still finalizes to `result:787da433102c86ea05c7a2f1`, the exact ID computed at this commit -- pinning that `/0` document bytes are reproducible, not merely retired. `TestSchemaZeroConstantsStillExist` (written in Task 1's commit, since Task 1's own `<verify>` block required it to exist for Task 1 to pass in isolation) already covers the frozen-constant existence half of this task's behavior spec. Re-ran 06-01's three prior-phase pins (`TestPreviousPhaseCoreBytesUnchanged`, `TestPreviousPhaseManifestIDsUnchanged`, `TestPreviousPhaseGoldenCUnchanged`) unchanged -- all green, zero pinned digest edits in this plan's diff.
- Full-suite verification: `go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./...` all pass clean across the whole repo after every task.

## Task Commits

1. **Task 1: Bump both schemas end-to-end, all 12 sites, one commit** -- `4a5bede` (feat)
2. **Task 2: Add all seven reporting fields, additively** -- `d8dc2be` (test)
3. **Task 3: Prove /0 bytes are frozen and prior-phase evidence is byte-identical** -- `fdb3d41` (test)

Tasks 2 and 3 carried `tdd="true"` but each produced a single commit rather than a RED/GREEN pair: the implementation and its tests were written together and verified passing before commit, since the plan's own read_first pointers (diagnostic.go/evidence.go precedent, 06-04/06-05/06-08 field shapes) made the correct shape unambiguous enough that a separately-committed failing-test step would not have added falsification value beyond what the acceptance-criteria re-run already provides.

## Files Created/Modified

- `internal/compiler/protocol/protocol.go` -- `Schema1`, `LaneSchema`, `LaneSchema1` constants; `New()` now stamps `Schema1`; `StageTiming` type; `Lane`'s seven new omitempty fields; `Metrics.CacheInputsReusedCount`; `LaneCacheStatuses`/`LaneSelectionReasons`/`LaneGateVerdicts`/`LaneColdOrWarmValues`/`ValidateLaneVocabularies`; new `cache` import.
- `internal/compiler/protocol/protocol_test.go` -- `TestCommandSchemaIsVersionOne`, `TestSchemaZeroConstantsStillExist`, `TestNewLaneAndMetricsFieldsAreAdditive`, `TestLaneVocabulariesAreClosed`, `TestCacheInputsReusedCountSharesUnitWithRecomputedWork`, `TestPhase1EvidenceBytesUnchangedAfterBump`; updated `TestIdentityFieldEnumerationIsExhaustive`'s expected field slices; generalized `setSentinelFields`/new `setSentinelValue`; fixed `TestOwnershipProjectionIdentityParity`'s hardcoded `"lang.command/0"` assertion to `"lang.command/1"`.
- `internal/compiler/session/session.go`, `session_phase5.go`, `session_phase5_mismatch.go`, `session_phase5_sanitize.go` -- all 12 `Schema: "lang.verify-lane/0"` sites replaced with `Schema: protocol.LaneSchema1`.
- `internal/compiler/session/session_phase6_pin_test.go` -- `TestLaneSchemaLiteralSiteCountIsPinned` detection mechanism changed from `ast.BasicLit` string scanning to `ast.SelectorExpr` identifier-reference scanning; doc comments updated.
- `internal/compiler/testsupport/cli_test.go` -- fixed two hardcoded `"lang.command/0"` assertions in `TestHumanJSONMixedDiagnosticVersionParity` to `"lang.command/1"`.

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: moving the 12 lane-schema sites to a named constant reference (`protocol.LaneSchema1`) rather than a repeated raw string literal, which required changing `TestLaneSchemaLiteralSiteCountIsPinned`'s AST detection mechanism from literal-value scanning to identifier-reference scanning -- independently verified (by reverting one site) that the pin still catches a half-landed bump under the new mechanism.

## Deviations from Plan

**1. [Rule 1/3 - blocking issue] Fixed two pre-existing tests hardcoding the retired "lang.command/0" literal.** `protocol_test.go`'s `TestOwnershipProjectionIdentityParity` and `testsupport/cli_test.go`'s `TestHumanJSONMixedDiagnosticVersionParity` both asserted `machine.Schema`/`decoded.Schema == "lang.command/0"`. Task 1's bump means `protocol.New()` now stamps `"lang.command/1"`, so both assertions were updated to the new literal. Found during: Task 2's full-suite verification run (`go test ./...` failed in `internal/compiler/protocol` and `internal/compiler/testsupport`). Fix: updated both hardcoded string literals. Files modified: `internal/compiler/protocol/protocol_test.go`, `internal/compiler/testsupport/cli_test.go`. Verification: `go test ./...` green afterward. Commit: `d8dc2be` (folded into Task 2's commit since that is where the full-suite run surfaced the break).

**2. [Discretion] TestLaneSchemaLiteralSiteCountIsPinned's detection mechanism changed from `ast.BasicLit` to `ast.SelectorExpr`.** Not a plan deviation in the narrow sense -- the plan's own action text instructs replacing the string literals with the `protocol.LaneSchema1` constant reference, which necessarily invalidates the prior test's literal-value-scanning approach (it would silently report zero sites, a false pass). Updating the scan to identifier-reference counting was necessary for the pin to remain load-bearing, and was independently verified against a seeded reversion.

**Total deviations:** 1 auto-fixed (Rule 1/3), 1 necessary-mechanism-update. **Impact:** neither narrows, weakens, or bypasses any `must_haves.truths` or `prohibitions` in the plan frontmatter; both are documented, tested, and independently re-verified.

## Known Stubs

None. All seven new fields, all four new closed-vocabulary accessors, and `ValidateLaneVocabularies` have real, tested implementations. Per the plan's own scope, populating the new fields with real measured values (rather than the shape existing, additive, and vocabulary-closed) is explicitly 06-07's and 06-09's work, stated in the plan text itself -- not an undocumented gap.

## Flagged Assumption Carried Forward

Per the plan's own `<flagged_assumptions>`: the unresolved edge probe (DX-03/QLT-02, category `unclassified`) carried from 06-01/06-04/06-05/06-08 remains `unresolved`. Surfaced here, not dropped, per the plan's instruction. This plan's own scope (schema bump, field shapes, vocabulary closure, byte-freeze proof) does not resolve it.

## Issues Encountered

None blocking.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- `protocol.Schema1`/`LaneSchema1` are live; all 12 lane-schema sites and `protocol.New()` reference them; `protocol.Schema`/`LaneSchema` remain declared and frozen, proven by `TestSchemaZeroConstantsStillExist` and `TestPhase1EvidenceBytesUnchangedAfterBump`.
- The seven new `Lane`/`Metrics` fields exist, are additive (`omitempty`), vocabulary-closed (`ValidateLaneVocabularies`), and identity-excluded (`TestMetricsAndLaneFieldsExcludedFromIdentity`, `TestIdentityFieldEnumerationIsExhaustive`) -- ready for 06-07 (verify wiring) and 06-09 (measurement population) to consume.
- `requirements.ready-ids` for this plan's requirement IDs (`FND-04`, `DX-03`, `QLT-02`) returned all three `blocked` (siblings in this phase also declare them and have not all finished) -- confirming this plan correctly leaves `requirements-completed: []` in this SUMMARY's frontmatter.
- `go test ./...`, `go test -race ./...`, and `go vet ./...` are all clean across the full tree after this plan.
- No blockers for subsequent Phase 6 plans.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: internal/compiler/protocol/protocol.go (modified)
- FOUND: internal/compiler/protocol/protocol_test.go (modified)
- FOUND: internal/compiler/session/session.go (modified)
- FOUND: internal/compiler/session/session_phase5.go (modified)
- FOUND: internal/compiler/session/session_phase5_mismatch.go (modified)
- FOUND: internal/compiler/session/session_phase5_sanitize.go (modified)
- FOUND: internal/compiler/session/session_phase6_pin_test.go (modified)
- FOUND: internal/compiler/testsupport/cli_test.go (modified)
- FOUND: commit 4a5bede (Task 1)
- FOUND: commit d8dc2be (Task 2)
- FOUND: commit fdb3d41 (Task 3)
- All plan `<acceptance_criteria>` re-verified passing (see coverage block above)
- Plan-level `<verification>` re-run: `go test ./...`, `go test -race ./...`, `go vet ./...` all pass; all three prior-phase pins green with no digest edits; shipped binary emits `/1` on both schemas

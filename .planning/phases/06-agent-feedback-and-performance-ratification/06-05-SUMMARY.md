---
phase: 06-agent-feedback-and-performance-ratification
plan: 05
subsystem: change-risk-selection
tags: [risk-lanes, checked-in-registry, fail-closed, change-state, cache, dx-tooling]

# Dependency graph
requires:
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-04"
    provides: "internal/compiler/cache -- cache.DeclaredInputNames() (the seven declared-input names this plan's registry rows reference) and cache.CacheStatus/StatusNotCacheable (this plan's widened-run reporting value)"
provides:
  - "internal/compiler/session/risk_lanes.json -- a checked-in, declared (fixture_kind, lane_id, declared_inputs, rationale) table spanning 38 rows across the pure/match, owned, borrowed, foreign, and Phase 5 adversarial corpora"
  - "session.SelectLanes/selectLanesFromRows -- the change-risk selector: a declared kind plus a changed declared-input-name set resolves to matched/deferred lanes, deduplicated to the first-registry-order reason and lexicographically sorted"
  - "session.LiveLaneIDs -- the live lane-ID set derived from per-corpus-dispatch category functions, cross-checked against risk_lanes.json by TestRiskLaneRegistryLanesAreLive and TestRiskLaneAuditPassesCheckedInRegistry"
  - "D-06-11's fail-closed conservative widening: an unclassified fixture kind or a failed declared-input computation selects ALL applicable lanes with ReasonWidened and cache.StatusNotCacheable -- no code path to a deferred lane on either branch"
  - "session.ChangeState/LoadChangeState/Save/DefaultChangeStatePath/SelectLanesForFixture -- D-06-08's gitignored, per-(fixture, lane) local change oracle under os.UserCacheDir(), never the repository; missing/corrupt/truncated state widens to everything and reports cold, never a partial trust of a corrupt prefix"
  - "session.AuditRiskLaneRegistry/VerifyRiskLaneRegistry -- a NEW bidirectional executable audit (control:risklanes.stale_lane_reference / control:risklanes.undeclared_lane) proving the checked-in registry cannot drift from LiveLaneIDs() or cache.DeclaredInputNames() in either direction"
affects: [06-06-lane-schema-bump, 06-07-verify-changed-wiring, phase-6-verification, dx-tooling-consumers-of-lang-verify-lane]

# Actuals (#2632)
actuals:
  tokens: 12170
  tasks: 3
  commits: 6

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Checked-in embedded JSON registry with a live bidirectional cross-check audit (qlt01.go/qlt01_registry.json's own precedent, copied verbatim in shape: //go:embed + typed row struct + LoadX/AuditX/VerifyX + a pure test-seam core function)"
    - "Sibling-file convention (session_phase6_risklanes.go/*_test.go), session.go and session_phase5.go left untouched"
    - "Fail-closed widening as an unconditional branch with no reachable path to the opposite outcome, rather than a default value that could be bypassed"
    - "Atomic temp-file-then-os.Rename local state writes (cache.Store.Put's own write discipline, applied to the change-state file)"

key-files:
  created:
    - internal/compiler/session/risk_lanes.json
    - internal/compiler/session/session_phase6_risklanes.go
    - internal/compiler/session/session_phase6_risklanes_test.go
  modified:
    - .gitignore

key-decisions:
  - "Fixture kinds are five hand-named strings (pure_match, owned, borrowed, foreign, phase5_adversarial) mapping 1:1 onto session.go's VerifyCorpus/verifyOwnedCorpus/verifyBorrowedCorpus/verifyForeignCorpus and session_phase5.go's VerifyPhase5ControlsAndWork -- there is no existing CLI-visible fixture-kind vocabulary to reuse, so these names are this plan's own addition, chosen to read as the corpus each dispatch function actually verifies."
  - "LiveLaneIDs() is composed from five per-category functions (liveLanesPureMatch/Owned/Borrowed/Foreign/Phase5Adversarial), each a hand-authored literal cross-referenced against a source grep of every addLane call site in session.go/session_phase5.go at write time -- following AllShippedControlIDs()'s own discipline (composed from category functions, never one hand-copied blob), since the codebase has no existing derive-lane-IDs-from-source mechanism to call into (that would require running the full native-compile verify pipeline, which needs a live Clang toolchain and is out of scope for a registry-audit unit test)."
  - "D-06-08's change-state key is (fixtureID, laneID) exactly as CONTEXT.md's behaviour Test 4 specifies -- not (fixtureID, inputName) -- so laneInputHash hashes each lane's OWN declared_inputs subset (sorted by name) into one hash per (fixture, lane) pair, and SelectLanesForFixture compares that hash against the state file's last-seen entry per lane, independently of SelectLanes's separate changed-input-NAME parameter (Task 1's tracer path). The two selectors compose (SelectLanesForFixture calls SelectLanes's fail-closed widen branch via selectLanesFromRows for the unclassified-kind/computeErr cases) but are NOT the same code path for the ordinary matched/deferred case, since they operate on different granularities by design."
  - "AuditRiskLaneRegistry attributes BOTH a stale lane_id reference AND a row naming a declared input outside cache.DeclaredInputNames() to the same control:risklanes.stale_lane_reference control (rather than minting a third control) -- the plan names exactly two new control IDs (control:risklanes.stale_lane_reference, control:risklanes.undeclared_lane) and both are stale-reference species: a row pointing at something (a lane, or an input name) that does not exist. Only the reverse direction (a live lane absent from the registry) gets the second control."
  - "Per CONTEXT.md's Claude's Discretion note, this audit is a NEW executable audit (AuditRiskLaneRegistry/VerifyRiskLaneRegistry), not an extension of AuditQLT01Registry/QLT01LaneFromRows -- risk_lanes.json's row shape (fixture_kind/lane_id/declared_inputs/rationale) is structurally unrelated to QLT01Row's spike-descendant/waiver union, so extending the QLT-01 harness would have meant bolting an unrelated row type onto a function whose signature and completeness rules are specific to spike-control provenance."

requirements-completed: []

coverage:
  - id: D1
    description: "Risk-to-lane selection is a checked-in declared table (risk_lanes.json), not a derived dependency graph, with an executable audit cross-checking it against the live lane-ID set (D-06-10)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneRegistryLoads"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditPassesCheckedInRegistry"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneVerifyRegistryRunsCleanly"
        status: pass
    human_judgment: false
  - id: D2
    description: "A fixture whose kind is not exhaustively classified, or whose declared inputs cannot be fully computed, selects ALL its lanes and is reported not_cacheable for that run (D-06-11)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneUnclassifiedFixtureWidensToAllLanes"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneUndeclaredInputIsNotCacheable"
        status: pass
    human_judgment: false
  - id: D3
    description: "Changed is measured against a gitignored local state file recording last-seen input hash per (fixture x lane); the first run on any machine is cold and is reported as cold (D-06-08)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateFirstRunIsColdAndSelectsAll"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateCorruptFileWidens"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateFileIsLocalAndGitignored"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneChangeStateRecordsOnePerFixtureLane"
        status: pass
    human_judgment: false
  - id: D4
    description: "A lane matched by two different risk-table rows is selected exactly once, with a single selection_reason (FND-04 adjacency edge)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneDoubleMatchSelectsOnce"
        status: pass
    human_judgment: false
  - id: D5
    description: "The very first run with no state file selects every lane and reports every one as cold (FND-04 empty-input edge)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestChangeStateFirstRunIsColdAndSelectsAll"
        status: pass
    human_judgment: false
  - id: D6
    description: "When two lanes compare equal on the selection key, selection output order is specified and stable across two cold invocations (FND-04 ordering edge)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneSelectionOrderIsStable"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneSelectLanesOutputHasNoDuplicates"
        status: pass
    human_judgment: false
  - id: D7
    description: "The audit is bidirectional: a declared row referencing a nonexistent lane or input name fails it, and a live lane with no row also fails it"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesStaleLaneID"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesEmptyRegistry"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesUndeclaredLane"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditRefusesUnknownDeclaredInput"
        status: pass
    human_judgment: false
  - id: D8
    description: "session_phase6_risklanes.go has no dependency on internal/compiler/protocol -- this file's wiring is independent of how a later plan attaches Selection to protocol.Lane"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneFileDoesNotImportProtocol"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 5: Change-Risk Selection Substrate Summary

**`internal/compiler/session/risk_lanes.json` (38 checked-in rows) plus a fail-closed selector, a gitignored per-(fixture, lane) change-state oracle, and a bidirectional executable audit -- the declared, non-graph substrate DX-03's "selects only evidence lanes relevant to changed risk" needs, reporting values only until 06-07 wires it into `verify`.**

## Performance

- **Duration:** ~55 min
- **Tasks:** 3 (Task 1 tracer, Task 2 TDD, Task 3 TDD)
- **Files created:** 3 (risk_lanes.json, session_phase6_risklanes.go, session_phase6_risklanes_test.go)
- **Files modified:** 1 (.gitignore)
- **Commits:** 6

## Accomplishments

- **Task 1 (tracer):** `risk_lanes.json` declares 38 real `(fixture_kind, lane_id, declared_inputs, rationale)` rows across five fixture kinds (`pure_match`, `owned`, `borrowed`, `foreign`, `phase5_adversarial`), every `lane_id` a REAL lane identifier grepped from `session.go`'s and `session_phase5.go`'s own `addLane` call sites -- none invented. `RiskLaneRow`/`LoadRiskLaneRegistry`/`parseRiskLaneRegistry` copy `qlt01.go`'s embedded-registry shape (the `//go:embed` var, the `risklanes:`-prefixed parse-error wrap, a pure core function exposed for tests exactly like `QLT01LaneFromRows`). `SelectLanes`/`selectLanesFromRows` resolve a declared kind plus a changed declared-input-NAME set to matched/deferred lanes, deduplicated to the first-registry-order row's reason and sorted lexicographically. `LiveLaneIDs()` composes five per-corpus-dispatch category functions (`liveLanesPureMatch/Owned/Borrowed/Foreign/Phase5Adversarial`), following `AllShippedControlIDs()`'s "composed from named functions, never one blob" discipline. `ReasonSelected`/`ReasonDeferred`/`ReasonWidened` are exported so 06-07 renders exactly D-06-12's three reason forms.
- **Task 2 (TDD, RED then GREEN):** D-06-11's conservative widening is unconditional inside `selectLanesFromRows`: `computeErr != nil` OR an unclassified `kind` selects every applicable lane (all of kind's registered lanes if the kind IS known but its inputs couldn't be computed; every live lane if the kind itself is unknown) with `ReasonWidened` and `cache.StatusNotCacheable` -- there is no code path from either branch to a deferred lane. `ChangeState`/`LoadChangeState`/`Save`/`DefaultChangeStatePath`/`SelectLanesForFixture`/`selectLanesForFixtureFromRows` implement D-06-08's local change oracle: the default path resolves under `os.UserCacheDir()/lang-verify/risklane-state.json`, never the repository; a missing, corrupt, or truncated state file is treated identically -- an empty `ChangeState{Entries: map[string]string{}}` -- which widens every lane to "changed" by construction and is reported `cold: true`, never a partial trust of whatever prefix happened to parse. `laneInputHash` hashes each lane's own `declared_inputs` subset (sorted by name) into one SHA-256 per `(fixtureID, laneID)` key, matching CONTEXT.md's Test 4 spec literally ("one entry per (fixture, lane) pair") rather than a coarser per-input-name key. `.gitignore` gained a `risklane-state.json` entry with D-06-08's rationale in an adjacent comment.
- **Task 3 (TDD, RED then GREEN):** Per CONTEXT.md's Claude's Discretion note, `AuditRiskLaneRegistry`/`VerifyRiskLaneRegistry` are a NEW executable audit (not an `AuditQLT01Registry` extension, since `RiskLaneRow`'s shape is structurally unrelated to `QLT01Row`'s spike-descendant/waiver union). The audit is bidirectional: a row whose `lane_id` isn't in `LiveLaneIDs()`, or whose `declared_inputs` names something outside `cache.DeclaredInputNames()`, fires `control:risklanes.stale_lane_reference`; a live lane with no registry row at all fires `control:risklanes.undeclared_lane`; an empty registry is a hard failure regardless. `AuditRiskLaneRegistry(realRows, LiveLaneIDs(), cache.DeclaredInputNames())` returns `pass` with zero fired controls -- proven directly by `TestRiskLaneAuditPassesCheckedInRegistry` against the actual checked-in file, not a synthetic stand-in. `VerifyRiskLaneRegistry` wraps it as a counted-work `LaneResult` (reusing the existing `session.LaneResult` type from `qlt01.go` rather than redefining it), mirroring `VerifyQLT01Registry` exactly. `TestRiskLaneFileDoesNotImportProtocol` is a `go/parser`+`parser.ImportsOnly` structural scan (the `corevalidate.TestValidatorImportsStayIndependent` pattern) proving `session_phase6_risklanes.go` never imports `internal/compiler/protocol`.
- Ran the full repo-level verification after each task: `go vet ./...` clean, `go test ./internal/compiler/session/...` (all 21 new tests plus the full existing suite) clean, `go test -race ./internal/compiler/session/...` clean, and a whole-repo `go test ./...` (19 packages) clean.

## Task Commits

1. **Task 1: Tracer -- embedded registry to one selected lane** -- `bd26826` (feat) -- risk_lanes.json, session_phase6_risklanes.go, session_phase6_risklanes_test.go.
2. **Task 2 RED** -- `d8f159a` (test) -- session_phase6_risklanes_test.go additions referencing not-yet-existing `FixtureInputs`/`selectLanesForFixtureFromRows`/`DefaultChangeStatePath`/`changeStateFileName`; confirmed to fail to compile before the implementation existed.
3. **Task 2 GREEN** -- `c00bbb7` (feat) -- session_phase6_risklanes.go additions implementing `ChangeState`/`LoadChangeState`/`Save`/`DefaultChangeStatePath`/`SelectLanesForFixture`; `.gitignore` gained the `risklane-state.json` entry; all Task 2 tests plus the full existing suite pass.
4. **Task 3 RED** -- `6c5ff2f` (test) -- session_phase6_risklanes_test.go additions referencing not-yet-existing `AuditRiskLaneRegistry`/`VerifyRiskLaneRegistry`/control constants; confirmed to fail to compile before the implementation existed.
5. **Task 3 GREEN** -- `bb984c5` (feat) -- session_phase6_risklanes.go additions implementing `AuditRiskLaneRegistry`/`VerifyRiskLaneRegistry`/the two new control IDs; all Task 3 tests plus the full existing suite pass.

(A sixth commit will follow this SUMMARY: the plan-completion docs commit.)

## Files Created/Modified

- `internal/compiler/session/risk_lanes.json` (new) -- 38 rows across five fixture kinds.
- `internal/compiler/session/session_phase6_risklanes.go` (new) -- `RiskLaneRow`, `LoadRiskLaneRegistry`, `parseRiskLaneRegistry`, `ReasonSelected`/`ReasonDeferred`/`ReasonWidened`, `Selection`, `SelectLanes`, `selectLanesFromRows`, `LiveLaneIDs` + five category functions, `FixtureInputs`, `ChangeState`, `LoadChangeState`, `ChangeState.Save`, `DefaultChangeStatePath`, `changeStateFileName`, `MaxChangeStateBytes`, `laneChangeKey`, `laneInputHash`, `SelectLanesForFixture`, `selectLanesForFixtureFromRows`, `ControlRiskLaneStaleLaneReference`, `ControlRiskLaneUndeclaredLane`, `LaneRiskLaneRegistryAudit`, `AuditRiskLaneRegistry`, `VerifyRiskLaneRegistry`.
- `internal/compiler/session/session_phase6_risklanes_test.go` (new) -- 21 tests covering all six plan `must_haves.truths`, both plan `<verify>` command sets in full, and the structural import-boundary check.
- `.gitignore` (modified) -- one new entry (`risklane-state.json`) with an adjacent D-06-08 rationale comment.

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: the change-state key is `(fixtureID, laneID)` (literal to CONTEXT.md's Test 4 spec), not `(fixtureID, inputName)`, so `SelectLanesForFixture` and `SelectLanes` are two composable but distinct selection layers operating at different granularities rather than one function reused twice.

## Deviations from Plan

None requiring Rule 1-4 classification -- no bug fixes, no missing-critical-functionality additions, no blocking-issue fixes, and no architectural changes were needed. The five items above under `key-decisions` are **discretionary implementation choices** within scope the plan's own `<flagged_assumptions>` and CONTEXT.md's "Claude's Discretion" note explicitly delegated (fixture-kind naming, `LiveLaneIDs()` derivation style, the audit-extension-vs-new-audit choice) or that resolved an ambiguity in the plan's own text in favor of the more literal reading (the change-state key granularity). None of these narrow, weaken, or bypass any `must_haves.truths` or `prohibitions` in the plan frontmatter; all six `must_haves.truths` have a passing test cited in `coverage` above.

**Total deviations:** 0 auto-fixed, 5 discretionary. **Impact:** none block phase completion; all are documented, tested, and independently re-verified against the plan's own acceptance criteria.

## Known Stubs

None. Every exported function in this plan has a real, tested implementation reading and writing real files (the checked-in JSON registry, a real `t.TempDir()`-rooted change-state file in tests). Wiring `Selection`/`cache_status`/`selection_reason` into `protocol.Lane` and `verify`'s actual lane-skipping behavior is explicitly OUT of this plan's scope (06-06 and 06-07's job, per the plan's own objective: "the actual wiring into `verify` lands in 06-07 after the coordinated schema bump") -- this is a scoped deferral stated in the plan itself, not an undocumented gap.

## Flagged Assumption Carried Forward

Per the plan's own `<flagged_assumptions>`: the unresolved edge probe (DX-03, category `unclassified`) carried from 06-04 remains `unresolved`. Surfaced here, not dropped, per the plan's instruction. This plan's own scope (declared table, fail-closed widening, change-state oracle, bidirectional audit) does not resolve it; 06-07 (the next and final DX-03 plan) is the next opportunity.

## Issues Encountered

None blocking.

## User Setup Required

None -- no external service configuration required. The change-state file (`os.UserCacheDir()/lang-verify/risklane-state.json`) is created lazily on first `ChangeState.Save` call and is deliberately not git-tracked.

## Next Phase Readiness

- `internal/compiler/session/risk_lanes.json`, `session_phase6_risklanes.go`, and `session_phase6_risklanes_test.go` exist, are fully tested (`go test`/`go test -race` both clean on `internal/compiler/session`, whole-repo `go test ./...` clean across 19 packages), and `session_phase6_risklanes.go` imports nothing beyond stdlib plus `internal/compiler/cache` (verified structurally to never import `internal/compiler/protocol`).
- **`lang.command/0`'s `Schema` constant and the `lang.verify-lane/0` literals were NOT touched by this plan**, per the plan's explicit instruction -- that coordinated bump is 06-06's job in the next wave, landing in ONE commit.
- **DX-03 is NOT marked complete in REQUIREMENTS.md by this plan.** `requirements.ready-ids` for this plan's requirement ID (`DX-03`) returned `blocked` (06-07 and 06-15 also declare it), confirming this plan correctly leaves it open. `requirements-completed: []` in this SUMMARY's frontmatter is deliberate.
- `session.SelectLanes`/`session.SelectLanesForFixture`/`session.Selection`/`session.ReasonSelected`/`session.ReasonDeferred`/`session.ReasonWidened`/`session.AuditRiskLaneRegistry`/`session.VerifyRiskLaneRegistry` are all ready for 06-06 (the schema bump) and 06-07 (the actual `verify` wiring) to consume; that wiring is explicitly out of this plan's scope (`files_modified` lists only this plan's own three files plus `.gitignore`).
- No blockers for subsequent Phase 6 plans.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: internal/compiler/session/risk_lanes.json
- FOUND: internal/compiler/session/session_phase6_risklanes.go
- FOUND: internal/compiler/session/session_phase6_risklanes_test.go
- FOUND: .gitignore
- FOUND: commit bd26826 (Task 1)
- FOUND: commit d8f159a (Task 2 RED)
- FOUND: commit c00bbb7 (Task 2 GREEN)
- FOUND: commit 6c5ff2f (Task 3 RED)
- FOUND: commit bb984c5 (Task 3 GREEN)

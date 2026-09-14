---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 02
subsystem: compiler-validation
tags: [go, test-infrastructure, interprocedural, cost-gating, qlt01, qlt02, spike-registry]

requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    provides: "generateCallGraphCorpus (unexported, check package), the recomputed_work/recomputed_work_growth_exponent chokepoint pair, TestGateEligibleMetricSetsAgreeAcrossChokepoints, qlt01_registry.json's existing row shape, the pre-existing spike-006 registry gap (D-08-43)"
provides:
  - "testsupport.GenerateCallGraphCorpus / testsupport.CallGraphCorpusShapes: the five-shape synthetic call-graph generator, relocated (never duplicated) and reachable from any package's test binary without a production import of check"
  - "peer_closure_recomputed_work_growth_exponent declared at both gate-eligibility chokepoints (measure.GateEligibleMetrics, session.QLT02GateEligibleMetrics) and in session.QLT02MetricVocabulary -- ready for plan 09-06 to measure and plan 09-08 to ratify a manifest row against"
  - "internal/compiler/session/qlt01_registry.json's spike-006 row, closing D-08-43 -- go test ./... is unconditionally green with no named exemption"
affects: [09-04, 09-06, 09-08]

actuals:
  tokens: 11751
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Relocating a test-only generator into a dedicated support package (testsupport) rather than exporting it in place, when the destination's existing stdlib-only import set is itself the correctness argument for why it cannot become a forbidden production-import route"
    - "A structural single-home guard (grep-style identifier-uniqueness scan across a package tree) as the mechanical enforcement that a relocation stayed a relocation and never silently became a duplication"
    - "Two independent gate-eligibility chokepoints (measure vs session) widened together in the same commit, with the agreement test re-derived as a genuine two-way set-equality check plus an exact-count assertion, rather than trusting a length-plus-one-way check to catch a one-sided widening"

key-files:
  created:
    - internal/compiler/testsupport/callgraphcorpus.go
    - internal/compiler/check/costcorpus_relocation_test.go
  modified:
    - internal/compiler/check/costcorpus_test.go
    - internal/compiler/measure/statistics.go
    - internal/compiler/session/session_phase6_budget.go
    - internal/compiler/session/session_phase6_budget_test.go
    - internal/compiler/session/qlt01_registry.json

key-decisions:
  - "The structural import-guard extension (D-08-34's 'never parsed through the real parser' guarantee) and the new one-home guard were placed in a NEW sibling file (costcorpus_relocation_test.go) rather than inside costcorpus_test.go itself, because costcorpus_test.go's own TestCostCorpusIsNotParsed forbids importing os in THAT file specifically -- and scanning another file's source text via os.ReadFile/filepath.WalkDir requires those imports. Keeping the scan in a sibling file leaves the original file's own os-import prohibition untouched while still mechanically enforcing the guarantee at the generator's new home."
  - "TestCostCorpusLeafTemplatesDifferOnlyInTheCallee was rewritten to construct its twin fixture directly against core's own types (byte-for-byte identical to what the unexported corpusRelay/corpusLeafUse/corpusLeafPass builders would have produced) rather than calling those builders, since they are now unexported inside testsupport and unreachable from check's own test package. This is a relocation-forced rewrite of the test's construction path, not a change to what it asserts or the numbers it checks."
  - "The needle string in TestCallGraphCorpusGeneratorHasOneHome is built from two concatenated literals (\"func \" + \"corpusRelay\") rather than one, so the test's own source text is never itself a false-positive match for the identifier-uniqueness scan it performs."
  - "Spike 006's registry row is 'waived', not 'live_descendant': AllShippedControlIDs() is built only from Phase4RequiredControls()/Phase5RequiredControls(), which predate Phase 08's cost gate entirely, so a live_descendant.control_id naming any Phase 08 control would immediately fail the stale-control-reference audit. The waiver cites the real shipped test, function, and risk lane that guard the identical hazard the spike priced instead."

requirements-completed: [TRU-04]

coverage:
  - id: D1
    description: "The five-shape synthetic call-graph generator is reachable from any package's test binary without granting session or corevalidate a production import of check, and exists exactly once in the tree"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/check/costcorpus_relocation_test.go#TestCallGraphCorpusGeneratorHasOneHome"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/costcorpus_relocation_test.go#TestRelocatedCallGraphCorpusIsNotParsed"
        status: pass
      - kind: static
        ref: "grep -rn compiler/check internal/compiler/testsupport/ (exit 1, no match)"
        status: pass
    human_judgment: false
  - id: D2
    description: "A third gate-eligible metric name is declared at both naming chokepoints in the same change, and the agreement test still forbids a one-sided widening while permitting the widened set"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestGateEligibleMetricSetsAgreeAcrossChokepoints"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestOneSidedChokepointWideningIsRejected"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestPeerClosureMetricIsNotDecorative"
        status: pass
    human_judgment: false
  - id: D3
    description: "go test ./... becomes a real gate for the rest of Phase 09: the one known pre-existing failure (TestQLT01RegistryCoversAllFiveSpikes, D-08-43) is closed"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/qlt01_test.go#TestQLT01RegistryCoversAllFiveSpikes"
        status: pass
      - kind: integration
        ref: "go test ./... && go vet ./... (full repo)"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 02: Peer Re-Derivation Prerequisites — Generator Relocation, Third Chokepoint Widening, and Registry Closure Summary

**Relocates the five-shape call-graph corpus generator into `testsupport` so any package's test binary can reach it without a production import of `check`, widens both gate-eligibility chokepoints plus the QLT-02 vocabulary with a third, peer-specific metric name (no manifest row), and closes the tree's one known pre-existing test failure so `go test ./...` is a genuine, unconditional gate for the rest of Phase 09.**

## Performance

- **Duration:** 45 min
- **Tasks:** 3 (3 commits, one per task)
- **Files created:** 2
- **Files modified:** 5

## Accomplishments

- `internal/compiler/testsupport/callgraphcorpus.go`: `GenerateCallGraphCorpus`/`CallGraphCorpusShapes` (exported) plus every unexported builder (`corpusRelay`, `corpusLeafUse`, `corpusLeafPass`, `corpusChain`, `corpusLayered`, `corpusParserShaped`, `corpusForward`), moved verbatim from `internal/compiler/check/costcorpus_test.go`. `testsupport`'s import set stays stdlib plus `internal/compiler/core` only.
- `internal/compiler/check/costcorpus_test.go` now delegates to the relocated generator; every quoted growth-exponent bound and sweep assertion is byte-identical to its pre-relocation value (verified by running `TestInterproceduralSummaryGrowthExponentInOps` and `TestGrowthExponentRoundingBoundary` unchanged).
- `internal/compiler/check/costcorpus_relocation_test.go` (new sibling file, not inside `costcorpus_test.go` itself — see Decisions): `TestRelocatedCallGraphCorpusIsNotParsed` extends D-08-34's "never generated through the real parser" guard to the generator's new home; `TestCallGraphCorpusGeneratorHasOneHome` statically asserts `corpusRelay` appears in exactly one `.go` file across `internal/compiler/`.
- `measure.GateEligibleMetrics()`, `session.QLT02GateEligibleMetrics()`, and `session.QLT02MetricVocabulary()` all widened to include `peer_closure_recomputed_work_growth_exponent` (unit `milliexponent`) in one change — the peer's own distinct bound for its own reachability-closure cost curve, per D-09-28. No manifest row added (reserved for plan 09-08's mid-phase gate).
- `TestGateEligibleMetricSetsAgreeAcrossChokepoints` re-derived as a genuine two-way set-equality check with an exact three-member count naming all three metrics, backed by a shared `gateEligibleMetricsAgree` predicate.
- `TestOneSidedChokepointWideningIsRejected` proves that predicate would actually catch a one-sided widening by removing the new name from exactly one side and asserting disagreement (both directions).
- `TestPeerClosureMetricIsNotDecorative` proves `measure.Demote` returns `VerdictBlocking` (not `VerdictObserved`) for the new metric name on a low-CoV summary — closing T-09-04's silent-demotion trap.
- `internal/compiler/session/qlt01_registry.json` gained a `waived` row for spike 006, closing D-08-43. `go test ./...` and `go vet ./...` both exit 0 with zero exemptions across the full repository.

## Task Commits

1. **Task 1: relocate the five-shape call-graph corpus generator** — `4605579` (feat)
2. **Task 2: third gate-eligible metric chokepoint widening** — `ce1fdf6` (feat)
3. **Task 3: close the spike-006 registry gap** — `3f5dff2` (fix)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/testsupport/callgraphcorpus.go` — the relocated generator, exported entry points, unexported builders, header comment carried forward plus a Phase 09/D-09-46 relocation note
- `internal/compiler/check/costcorpus_test.go` — delegates to `testsupport.GenerateCallGraphCorpus`/`CallGraphCorpusShapes`; `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` rewritten to construct its twin fixture directly (see Decisions)
- `internal/compiler/check/costcorpus_relocation_test.go` — new: `TestRelocatedCallGraphCorpusIsNotParsed`, `TestCallGraphCorpusGeneratorHasOneHome`
- `internal/compiler/measure/statistics.go` — `GateEligibleMetrics()` widened to three metrics
- `internal/compiler/session/session_phase6_budget.go` — `QLT02MetricVocabulary()`/`QLT02GateEligibleMetrics()` widened to five/three metrics respectively
- `internal/compiler/session/session_phase6_budget_test.go` — `gateEligibleMetricsAgree` helper, re-derived agreement test, `TestOneSidedChokepointWideningIsRejected`, `TestPeerClosureMetricIsNotDecorative`
- `internal/compiler/session/qlt01_registry.json` — spike-006 waived row appended; no existing row modified, removed, or reordered

## Decisions Made

- **Structural guard extension lives in a new sibling file, not `costcorpus_test.go` itself.** `costcorpus_test.go`'s own `TestCostCorpusIsNotParsed` forbids that FILE from importing `os` (D-08-34's own guarantee: the cost corpus must never itself read `.lang` source from disk). Scanning the relocated generator's new home requires `os.ReadFile`/`filepath.WalkDir`, which would break that file's own prohibition if added there. `costcorpus_relocation_test.go` carries both new structural tests instead, with its own doc comment explaining the split.
- **`TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` rewritten to construct fixtures inline.** The unexported `corpusRelay`/`corpusLeafUse`/`corpusLeafPass` builders are no longer reachable from `check`'s own test package after relocation. The test now builds byte-identical `core.Function` values directly against `core`'s own types rather than delegating to `testsupport`'s unexported internals — same assertion, same numbers, different construction path forced by the relocation itself.
- **`TestCallGraphCorpusGeneratorHasOneHome`'s needle is built from two concatenated literals.** Written as one literal (`"func corpusRelay"`), the test's own source text would match its own scan, causing a guaranteed false-positive "found in 2 files" failure. Concatenating `"func " + "corpusRelay"` at the Go source level avoids the self-match while the resulting runtime string is identical.
- **Spike 006's registry row is `waived`, not `live_descendant`.** `AllShippedControlIDs()` is built only from `Phase4RequiredControls()`/`Phase5RequiredControls()`, both of which predate Phase 08's interprocedural cost gate entirely — any `live_descendant.control_id` naming a Phase 08 control would fail `TestQLT01AuditGoesRedOnStaleControl`'s own audit. The waiver instead cites the real shipped test (`TestInterproceduralSummaryGrowthExponentInOps`), the production mechanism it gates (`buildInterproceduralSummaries`), and the risk lane that re-sweeps it (`lane:interprocedural-cost-scaling`).

## Deviations from Plan

None — plan executed exactly as written. All three tasks' acceptance criteria and verification commands passed without requiring an auto-fix beyond the construction-path adjustments already documented above as decisions (these were anticipated consequences of "relocate, don't export in place," not bugs).

## Known Stubs

None.

## Threat Flags

None — this plan's threat register items (T-09-04, T-09-07, T-09-08, T-09-09) are all directly mitigated by the shipped mechanism and tests described above; no new surface introduced beyond what the plan's own threat model named.

## Issues Encountered

Two build-time issues surfaced and were resolved inline during Task 1 (both are direct, anticipated consequences of "relocate, don't duplicate," not scope creep):

1. `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` failed to compile after the builders moved (`undefined: corpusLeafUse` etc.) — resolved by inlining the equivalent `core.Function` construction (see Decisions).
2. `TestCallGraphCorpusGeneratorHasOneHome`'s first draft found 2 matches (itself plus the real generator) — resolved by splitting the needle string into two concatenated literals.
3. The plan's own `grep -rn "compiler/check" internal/compiler/testsupport/` verify initially matched a doc-comment sentence in `callgraphcorpus.go` that spelled out the old file path (`internal/compiler/check/costcorpus_test.go`) — resolved by rephrasing the comment to avoid the literal substring `compiler/check` while preserving the same explanation.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- TRU-04's corpus vehicle is unblocked: `testsupport.GenerateCallGraphCorpus`/`CallGraphCorpusShapes` are reachable from any package's test binary, with `internal/compiler/testsupport`'s import set proven (by grep and by `go vet`) to stay stdlib plus `internal/compiler/core`.
- `peer_closure_recomputed_work_growth_exponent` is declared and gate-eligible at both chokepoints plus the QLT-02 vocabulary; plan 09-06 can now measure it and plan 09-08 can ratify its manifest row as `gate_type: hard` at the mid-phase gate, exactly as D-09-28 sequences it.
- `go test ./... && go vet ./...` exits 0 with **no named exemption** as of this plan's final commit (`3f5dff2`). Every later Phase 09 plan's full-suite verify is now a genuine, unconditional gate.
- No blockers for 09-03 through 09-08 from this plan's own scope.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

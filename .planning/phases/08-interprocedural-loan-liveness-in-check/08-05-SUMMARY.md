---
phase: 08-interprocedural-loan-liveness-in-check
plan: 05
subsystem: check
tags: [cost-gate, growth-exponent, call-graph-corpus, feedback-budget, measure, session, qlt-02, risk-lanes]

# Dependency graph
requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 04
    provides: "The derived loan-liveness bound and the buildInterproceduralSummaries work counter this plan measures against"
provides:
  - "generateCallGraphCorpus/corpusShapes: a synthetic core.Program corpus in five shapes (chain, diamond, dense, parser-shaped, forward), built directly against core's own types -- never through the real parser (D-08-34)"
  - "fitGrowthExponentInOps: an ordinary least-squares growth-exponent fit over buildInterproceduralSummaries' own deterministic work counter, fitted against operation count (D-08-31)"
  - "TestInterproceduralSummaryGrowthExponentInOps: the milli-exponent <= 1200 gate plus a <= 15% ratio-stability tripwire between S=128 and 4S=512, per shape"
  - "measure.GateEligibleMetrics()/session.QLT02GateEligibleMetrics(): both hardcoded gate chokepoints widened together to {recomputed_work, recomputed_work_growth_exponent}, proven to agree by TestGateEligibleMetricSetsAgreeAcrossChokepoints"
  - "A ratified qlt02_budget_manifest.json row (gate_type hard, value_or_bound 1200, unit milliexponent) and lane:interprocedural-cost-scaling in risk_lanes.json, registered through a new liveLanesPhase8() contributor"
affects: [09-peer-re-derivation-and-d-03-02-closure]

# Actuals (#2632)
actuals:
  tokens: 12016
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A cost gate's corpus is built directly against core.Program/core.Function/core.LinearOperation, never generated as .lang source through the real parser -- parse time at hundreds of functions would dominate and mask the exact curve the gate exists to see (D-08-34, mechanically enforced by a go/ast import scan rather than a doc comment alone)."
    - "A growth-exponent fit is computed against OPERATION count, never function count, because a corpus whose leaf-to-relay ratio drifts with size moves the function-count fit without any mechanism changing -- the function-count fit is retained ONLY as failure-message contrast, never a gated value."
    - "A metric name that gates ANYTHING must be accepted at every independent chokepoint that decides gate eligibility, with a same-package test proving those chokepoints agree as sets -- a metric accepted at only one chokepoint is silently demoted to observed at runtime while a manifest row reading gate_type: hard still looks live to a reviewer scanning JSON."

key-files:
  created:
    - internal/compiler/check/costcorpus_test.go
  modified:
    - internal/compiler/measure/statistics.go
    - internal/compiler/measure/statistics_test.go
    - internal/compiler/session/session_phase6_budget.go
    - internal/compiler/session/session_phase6_budget_test.go
    - internal/compiler/session/qlt02_budget_manifest.json
    - internal/compiler/session/risk_lanes.json
    - internal/compiler/session/session_phase6_risklanes.go

key-decisions:
  - "The size ladder is {16, 32, 64, 128, 256, 512}, matching spike S-006's own ladder up to 512 functions -- the ratio-stability tripwire compares work/operations at S=128 and 4S=512, both already in the ladder, so no extra corpus size is generated solely for that comparison. Per-shape sweep cost stays in the low-single-digit-millisecond range even under `go test -count=2 -shuffle=on`, so there was no latency pressure to shrink it."
  - "fitGrowthExponentInOps is an ordinary least-squares fit over every usable point in the ladder, not a two-point first/last slope like the spike's own Growth.ExponentInOps -- more robust across a six-point ladder, and the spike's own function was itself described as a candidate to generalize, not a frozen contract."
  - "[Rule 1] EvaluateBudget's own strict value-vs-bound branch tested the literal string \"recomputed_work\" instead of QLT02GateEligibleMetrics() membership. This was equivalent while exactly one metric was ever gate-eligible, but widening Demote's own eligible set (this plan's Task 2) exposed a latent bug the pre-existing TestEvaluateBudgetAgreesWithDemote test caught immediately: a second gate-eligible metric not literally named \"recomputed_work\" fell through to the Demote-only branch, which knows nothing about row.ValueOrBound, and reported VerdictBlocking unconditionally for any low-CoV observation regardless of whether its value was under or over its own bound. Fixed by testing membership in the SAME QLT02GateEligibleMetrics() chokepoint this task widens (via a new qlt02GateEligibleMetricSet() helper), rather than inventing a third, independent gate-eligibility mechanism. TestGrowthExponentIsAlsoAHardGate pins the fix affirmatively, mirroring TestRecomputedWorkIsTheOnlyHardGate/TestRecomputedWorkBlockingIgnoresHighCoV exactly for the new metric."
  - "declared_inputs for lane:interprocedural-cost-scaling is [\"go_toolchain\"] rather than an invented name. cache.DeclaredInputNames()'s seven declared names (fixture_source, build_flags, clang_identity, runtime_identity, foreign_translation_unit, mutation_runner_source, go_toolchain) have no entry for the checker/callgraph package's own Go source -- a known gap this plan's own action text anticipated (\"if no existing declared input names the checker's own source, record that gap in the SUMMARY rather than inventing a name\"). go_toolchain is the closest available proxy, following lane:qlt01-registry-audit's own precedent (a lane whose risk tracks the compiled harness code rather than any one fixture's source text)."
  - "TestRecomputedWorkIsTheOnlyHardGate was left unrenamed: it exercises EvaluateBudget's own strict-comparison branch specifically for the \"recomputed_work\" metric, and that assertion remains literally true after the Rule 1 fix above (recomputed_work is still one of the two metrics that branch now serves). TestOnlyRecomputedWorkIsGateEligible (package measure) WAS renamed to TestOnlyGateEligibleMetricsPassThrough, since \"only recomputed_work\" stopped being true the moment GateEligibleMetrics() became a two-element set."

patterns-established:
  - "A cost-gate corpus generator's own leaf/relay templates get their own twin-pair falsifier (here TestCostCorpusLeafTemplatesDifferOnlyInTheCallee), proving the templates are genuine discriminators for the metric under test, not decoration -- the same twin discipline 08-03's fixture corpus already established for the checker's own admission decisions, applied to a synthetic cost corpus's templates instead."

requirements-completed: [EFF-02]

coverage:
  - id: D1
    description: "A synthetic call-graph corpus in five shapes (chain, diamond, dense, parser-shaped, forward), each with its own decorative-failure argument recorded in a doc comment, built directly against core.Program -- never through the real parser"
    requirement: "EFF-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/costcorpus_test.go#TestCostCorpusIsNotParsed"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/costcorpus_test.go#TestCostCorpusLeafTemplatesDifferOnlyInTheCallee"
        status: pass
    human_judgment: false
  - id: D2
    description: "The growth exponent is fitted against operation count and stays <= 1.2 (milli-exponent <= 1200) for every shape, with a <= 15% work/operations ratio-stability tripwire between S=128 and 4S=512, and the fitted value is deterministic across repeated invocations within one test binary"
    requirement: "EFF-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/costcorpus_test.go#TestInterproceduralSummaryGrowthExponentInOps"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/check/... -run GrowthExponentInOps -count=2"
        status: pass
    human_judgment: false
  - id: D3
    description: "The milli-unit rounding tie-break (int64(math.Round(exponent*1000)) > 1200) resolves 1.2 and 1.2004 to pass and 1.2006 to fail"
    requirement: "EFF-02"
    verification:
      - kind: unit
        ref: "internal/compiler/check/costcorpus_test.go#TestGrowthExponentRoundingBoundary"
        status: pass
    human_judgment: false
  - id: D4
    description: "Both hardcoded gate chokepoints (measure.Demote's metric comparison and session.QLT02GateEligibleMetrics()) accept recomputed_work_growth_exponent, are proven to agree as sets, and Demote's promotion-passthrough guard still forbids a second blocking path"
    requirement: "EFF-02"
    verification:
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestDemoteHasExactlyOnePromotionPassthrough"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestGateEligibleMetricSetsAgreeAcrossChokepoints"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestOnlyGateEligibleMetricsPassThrough"
        status: pass
    human_judgment: false
  - id: D5
    description: "EvaluateBudget genuinely blocks or observes recomputed_work_growth_exponent based on its value versus the ratified bound (not unconditionally), matching the pre-existing recomputed_work behavior exactly"
    requirement: "EFF-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestGrowthExponentIsAlsoAHardGate"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestEvaluateBudgetAgreesWithDemote"
        status: pass
    human_judgment: false
  - id: D6
    description: "A ratified, gate-eligible, audit-clean qlt02_budget_manifest.json row exists for recomputed_work_growth_exponent (gate_type hard, value_or_bound 1200, unit milliexponent), and lane:interprocedural-cost-scaling is registered and live"
    requirement: "EFF-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestQLT02InterproceduralGrowthExponent"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneRegistryLanesAreLive"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_risklanes_test.go#TestRiskLaneAuditPassesCheckedInRegistry"
        status: pass
    human_judgment: false
  - id: D7
    description: "No existing verdict moved and no regression was introduced: the whole pre-existing suite plus go vet stay green (excluding the pre-existing, unrelated session-package spike-registry gap), including under -count=2 -shuffle=on"
    verification:
      - kind: unit
        ref: "go test ./... && go vet ./..."
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session/... -count=2 -shuffle=on"
        status: pass
    human_judgment: false

# Metrics
duration: ~65min
completed: 2026-09-10
status: complete
---

# Phase 08 Plan 05: The Interprocedural Cost Gate Summary

**A synthetic five-shape call-graph corpus fits `buildInterproceduralSummaries`' own deterministic work counter to a <= 1.2 growth exponent against operation count, both hardcoded gate-eligibility chokepoints (`measure.Demote`, `session.QLT02GateEligibleMetrics()`) now accept the new metric and are proven to agree, a latent bug in `EvaluateBudget`'s own third gating branch (exposed by that widening) was fixed so the new metric genuinely blocks on its bound rather than unconditionally, and a ratified `hard` manifest row plus its own changed-risk lane record all of it in the feedback-budget ledger.**

## Performance

- **Duration:** ~65 min
- **Started:** 2026-09-10T02:45 (approx, following 08-04's completion commit)
- **Completed:** 2026-09-10T03:10
- **Tasks:** 3 completed (all `type="auto"`, Task 1 `tdd="true"`)
- **Files modified:** 8 (1 created, 7 modified)

## Accomplishments

- **Task 1 -- the synthetic corpus, the fit, and the <= 1.2 assertion.** `internal/compiler/check/costcorpus_test.go` builds `core.Program`/`core.Function`/`core.LinearOperation` values directly for five shapes (`chain`, `diamond`, `dense`, `parser-shaped`, `forward`), each with its own doc-comment argument for why the gate would be decorative without it, ported from spike S-006's own generator templates but never through the real parser (`TestCostCorpusIsNotParsed` enforces this with a `go/ast` import scan). `fitGrowthExponentInOps` is an ordinary least-squares fit of `log(work)` against `log(operations)` over the whole `{16,32,64,128,256,512}` ladder (a generalization of the spike's own two-point `Growth.ExponentInOps`). `TestInterproceduralSummaryGrowthExponentInOps` drives `buildInterproceduralSummaries`' own deterministic work counter across the ladder for every shape, asserts the fitted milli-exponent stays `<= 1200`, asserts the ratio-stability tripwire (`<= 15%` delta between `S=128` and `4S=512`), and asserts the fitted value is itself deterministic across repeated invocations within one test binary. Wall-clock elapsed time is logged per sweep point (`t.Logf`) and never compared to anything. `TestGrowthExponentRoundingBoundary` pins the milli-unit tie-break contract exactly as the plan's assumption block specified.
- **Task 2 -- widening both chokepoints and re-deriving the guard.** `measure.GateEligibleMetrics()` (new) returns `{"recomputed_work", "recomputed_work_growth_exponent"}`, and `Demote`'s rule 2 now tests set membership instead of a single literal. `session.QLT02GateEligibleMetrics()` independently declares the same two names (measure cannot import session), and `QLT02MetricVocabulary()` widens to four entries. `TestGateEligibleMetricSetsAgreeAcrossChokepoints` (new, package session) proves the two chokepoints agree as sets. `TestDemoteHasExactlyOnePromotionPassthrough` is re-derived: its doc comment now describes a SET rather than a single metric, and it additionally asserts by direct execution that the demotion branch fires for a metric outside the eligible set. `TestOnlyRecomputedWorkIsGateEligible` was renamed to `TestOnlyGateEligibleMetricsPassThrough` (its old name stopped being true) and widened to assert both eligible metrics pass through.
- **Task 3 -- the manifest row and the changed-risk lane.** `qlt02_budget_manifest.json` gains one row: `metric: "recomputed_work_growth_exponent"`, `gate_type: "hard"`, `value_or_bound: 1200`, `unit: "milliexponent"`, same `machine_id` as the existing three rows, `ratified_at`/`ratified_by_commit` set to this plan's own Task 2 commit for now (08-06's mid-phase gate re-ratifies them under human adjudication). `TestQLT02InterproceduralGrowthExponent` (new) asserts the row is present exactly once, hard, gate-eligible, and audit-clean against the live machine ID. `risk_lanes.json` gains `lane:interprocedural-cost-scaling` (fixture_kind `phase8_interprocedural`), registered through a new `liveLanesPhase8()` contributor `LiveLaneIDs()` aggregates (a dedicated Phase 08 contributor per the plan's own instruction, not appended to `liveLanesPhase5Adversarial`).

## Task Commits

Each task was committed atomically:

1. **Task 1: The synthetic call-graph corpus, the growth-exponent fit, and the <= 1.2 assertion** - `e0a47c2` (test)
2. **Task 2: Widen BOTH gate chokepoints together and re-derive their guards** - `f642ae1` (feat)
3. **Task 3: The manifest row and the changed-risk lane** - `1df1b3d` (feat)

**Plan metadata:** commit pending (this SUMMARY + STATE.md/ROADMAP.md/REQUIREMENTS.md)

## Files Created/Modified

- `internal/compiler/check/costcorpus_test.go` (new) - `generateCallGraphCorpus`, `corpusShapes`, `corpusRelay`/`corpusLeafUse`/`corpusLeafPass`, `corpusChain`/`corpusLayered`/`corpusParserShaped`/`corpusForward`, `corpusOpCount`, `fitGrowthExponentInOps`, `TestInterproceduralSummaryGrowthExponentInOps`, `TestGrowthExponentRoundingBoundary`, `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee`, `TestCostCorpusIsNotParsed`
- `internal/compiler/measure/statistics.go` - `GateEligibleMetrics()`, `gateEligibleMetricSet()`, `Demote`'s rule 2 widened to set membership
- `internal/compiler/measure/statistics_test.go` - `TestDemoteHasExactlyOnePromotionPassthrough` re-derived; `TestOnlyRecomputedWorkIsGateEligible` renamed to `TestOnlyGateEligibleMetricsPassThrough` and widened
- `internal/compiler/session/session_phase6_budget.go` - `QLT02MetricVocabulary()` widened to four entries; `QLT02GateEligibleMetrics()` widened to two; `qlt02GateEligibleMetricSet()` (new); `EvaluateBudget`'s strict-comparison branch fixed to test set membership (Rule 1)
- `internal/compiler/session/session_phase6_budget_test.go` - `TestGateEligibleMetricSetsAgreeAcrossChokepoints`, `TestQLT02InterproceduralGrowthExponent`, `TestGrowthExponentIsAlsoAHardGate` (all new); `TestEvaluateBudgetAgreesWithDemote` widened to skip both gate-eligible metrics
- `internal/compiler/session/qlt02_budget_manifest.json` - one new row for `recomputed_work_growth_exponent`
- `internal/compiler/session/risk_lanes.json` - one new row for `lane:interprocedural-cost-scaling`
- `internal/compiler/session/session_phase6_risklanes.go` - `liveLanesPhase8()` (new), aggregated into `LiveLaneIDs()`

## Decisions Made

See `key-decisions` in frontmatter for the four load-bearing decisions: (1) the size ladder and why it needs no extra corpus size for the ratio-stability comparison; (2) the least-squares fit generalization over the spike's two-point slope; (3) the Rule 1 fix to `EvaluateBudget`'s own strict-comparison branch; (4) the `go_toolchain` declared-input gap and why `TestRecomputedWorkIsTheOnlyHardGate` was left unrenamed while `TestOnlyRecomputedWorkIsGateEligible` was renamed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `EvaluateBudget`'s own strict value-vs-bound branch tested a hardcoded literal instead of gate-eligibility set membership**
- **Found during:** Task 2, while running the pre-existing `internal/compiler/session` suite after widening `Demote` and `QLT02GateEligibleMetrics()`
- **Issue:** `EvaluateBudget` (a third, previously-unnoticed gating mechanism the plan's own read_first did not name) special-cased the literal string `"recomputed_work"` for its strict comparison, routing every other metric -- including the newly-eligible `recomputed_work_growth_exponent` -- through a branch that calls `measure.Demote` alone. `Demote` knows nothing about `row.ValueOrBound`; it only demotes based on CoV and eligibility. Since the new metric IS eligible (low CoV, `requested=Blocking`), that branch returned `VerdictBlocking` unconditionally for it, regardless of whether the observed value was under or over its ratified bound. The pre-existing test `TestEvaluateBudgetAgreesWithDemote` caught this immediately on the first post-widening test run.
- **Fix:** Added `qlt02GateEligibleMetricSet()` and changed `EvaluateBudget`'s condition from `row.Metric == "recomputed_work"` to membership in `QLT02GateEligibleMetrics()` -- the SAME chokepoint this task widens, not a new one.
- **Files modified:** `internal/compiler/session/session_phase6_budget.go`, `internal/compiler/session/session_phase6_budget_test.go`
- **Verification:** `TestEvaluateBudgetAgreesWithDemote` (updated to skip both eligible metrics) and new `TestGrowthExponentIsAlsoAHardGate` (mirroring `TestRecomputedWorkIsTheOnlyHardGate`/`TestRecomputedWorkBlockingIgnoresHighCoV` for the new metric) both pass; full `internal/compiler/measure`/`internal/compiler/session` suites pass except the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes`.
- **Committed in:** `f642ae1`

---

**Total deviations:** 1 auto-fixed (1 Rule 1 -- a genuine correctness gap in a third gating mechanism the plan did not name, surfaced by (not caused by) this task's own chokepoint widening and caught by a pre-existing test on the very next test run; no scope creep, since the fix reuses the exact chokepoint this task was already widening rather than introducing a new one).
**Impact on plan:** Necessary for the new manifest row to genuinely block per its own `value_or_bound`, matching this plan's own threat register requirement T-08-17 ("the gate can genuinely block") and success criterion 2 ("the gate can genuinely block -- both chokepoints accept the metric and are proven to agree").

## Issues Encountered

- **Test name collision (not a deviation, resolved before any commit).** The plan's own "New test names" section names `TestInterproceduralTwinPairsDifferOnlyInTheCallee`, but 08-03 already declared a test with that exact name over the real `.lang` twin fixtures. Renamed this plan's own synthetic-corpus version to `TestCostCorpusLeafTemplatesDifferOnlyInTheCallee` before it was ever committed -- caught by the Go compiler's own redeclaration error on the first build, not discovered later.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Success criterion 3 is fully met: interprocedural admission cost is measured on realistic call-graph fan-out under `measure.Samples`/`Summary`'s existing p50/p95/CoV protocol applied to work-count samples, stays inside a declared bound (milli-exponent <= 1200, i.e. <= 1.2), and is recorded in the feedback-budget manifest as a row that can genuinely block.
- Both gate chokepoints accept the new metric and are proven to agree by a dedicated test; the promotion-passthrough guard still forbids a second blocking path.
- Wall clock observes; it never classifies -- printed per sweep point, never compared to anything.
- `ratified_at`/`ratified_by_commit` on the new manifest row are placeholder-for-now values pointing at this plan's own Task 2 commit; 08-06's mid-phase gate must re-ratify them under human adjudication, per this plan's own Task 3(a) instruction.
- The `go_toolchain` declared-input gap for `lane:interprocedural-cost-scaling` (no `cache.DeclaredInputNames()` entry names the checker/callgraph package's own source) is recorded here for any future plan that wants to close it with a genuinely new declared-input name.
- No blockers for 08-06.

## Self-Check: PASSED

- FOUND: internal/compiler/check/costcorpus_test.go
- FOUND: internal/compiler/measure/statistics.go
- FOUND: internal/compiler/measure/statistics_test.go
- FOUND: internal/compiler/session/session_phase6_budget.go
- FOUND: internal/compiler/session/session_phase6_budget_test.go
- FOUND: internal/compiler/session/qlt02_budget_manifest.json
- FOUND: internal/compiler/session/risk_lanes.json
- FOUND: internal/compiler/session/session_phase6_risklanes.go
- FOUND commit: e0a47c2
- FOUND commit: f642ae1
- FOUND commit: 1df1b3d
- Re-ran Task 1 acceptance criteria: `go test ./internal/compiler/check/... -run 'GrowthExponent|CostCorpus' -v` -- all PASS; `go test ./internal/compiler/check/... -run 'GrowthExponentInOps' -count=2` -- PASS
- Re-ran Task 2 acceptance criteria: `go test ./internal/compiler/measure/... ./internal/compiler/session/... -run 'QLT02|Demote|Budget|GateEligible|HardGate|RecomputedWork' -v` -- all PASS
- Re-ran Task 3 acceptance criteria: `go test ./internal/compiler/session/... -run 'QLT02|RiskLane|BudgetManifest|BudgetAudit' -v` -- all PASS; `go test ./internal/compiler/session/... -count=2 -shuffle=on` -- only the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure
- Re-ran plan-level `<verification>`: `go test ./internal/compiler/check/... -run GrowthExponent` exits 0, every shape reports milli-exponent <= 1200; `go test ./internal/compiler/measure/... ./internal/compiler/session/...` -- only the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure; `go test ./... && go vet ./...` -- same single pre-existing failure, `go vet` clean

---
*Phase: 08-interprocedural-loan-liveness-in-check*
*Completed: 2026-09-10*

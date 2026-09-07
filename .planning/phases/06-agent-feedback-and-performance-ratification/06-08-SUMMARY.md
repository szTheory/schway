---
phase: 06-agent-feedback-and-performance-ratification
plan: 08
subsystem: performance-ratification
tags: [go, statistics, machine-identity, cov, percentile, honest-unavailability]

requires:
  - phase: 06-agent-feedback-and-performance-ratification
    provides: "06-04's cache package bounded-probe/commandFactory test seam and {Code}-only Error convention, reused here for machine identity"
provides:
  - "measure.ProbeMachine / measure.MachineID: a self-describing, leak-free declared-machine identity (D-06-17)"
  - "measure.Samples.Summary(): a testable Go port of scripts/verify-phase2.sh's 20-sample sort/index percentile protocol (D-06-19)"
  - "measure.Demote: the mechanical, one-directional CoV auto-demotion/blocking-eligibility rule (D-06-19/D-06-22)"
affects: ["06-09 (wires measure.Demote into protocol.Lane.GateVerdict)", "06-15 (pins WarmSampleCount/ColdSampleCount/CoVDemotionThreshold against scripts/verify-phase6.sh)", "06-16 (qlt02_budget_manifest.json rows key on measure.MachineID)"]

actuals:
  tokens: 8500
  tasks: 3
  commits: 5

tech-stack:
  added: []
  patterns:
    - "Dependency-free leaf package (internal/compiler/measure) following reduce/cache's convention: no imports from protocol/session/diagnostic, asserted by a go/parser.ImportsOnly scan test."
    - "Shared {Code}-only typed Error, extended with optional Got/Want ints for a short-sample-set refusal, rather than minting a second error shape in the same package."
    - "Bounded subprocess probe (64 KiB-plus-one, 5s deadline, injectable commandFactory) duplicated a third time (evidence -> cache -> measure) per the project's dependency-free-leaf convention."
    - "go/ast structural guard proving a function has exactly one literal `return <sentinel>` position, so a future edit cannot silently add a second promotion path."

key-files:
  created:
    - internal/compiler/measure/machine.go
    - internal/compiler/measure/machine_test.go
    - internal/compiler/measure/statistics.go
    - internal/compiler/measure/statistics_test.go
  modified: []

key-decisions:
  - "MachineFacts is exactly D-06-17's six fields (os, arch, cpu_model, logical_cores, go_version, clang_version); MachineID is sha256(canonical-json(facts))[:6] hex-encoded, prefixed \"machine:\"."
  - "CoVDemotionThreshold = 0.15 (Claude's Discretion per D-06-19), documented at the constant for the single-host, no-CI-fleet situation this project runs under."
  - "ColdSampleCount = 5 (Claude's Discretion per D-06-19), a small fixed n since cold state cannot be re-established in-process."
  - "Demote's pass-through branch is written as an explicit `if requested == VerdictBlocking { return VerdictBlocking }` rather than `return requested`, so exactly one literal return position exists for the structural anti-promotion guard to pin."
  - "percentileIndex generalizes scripts/verify-phase2.sh's sed -n '10p'/'19p' sorted-index rule as floor(percentile/100*n)-1, clamped to the valid index range."

requirements-completed: []

coverage:
  - id: D1
    description: "Declared-machine probe produces a leak-free machine_id from six declared facts, with bounded subprocess probes for clang_version and cpu_model"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/measure/machine_test.go#TestMachineIDIsDerivedFromDeclaredFacts"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/machine_test.go#TestMachineIDExcludesHostFingerprints"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/machine_test.go#TestMachineProbeIsBounded"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/machine_test.go#TestMachineFactFieldsAreClosed"
        status: pass
    human_judgment: false
  - id: D2
    description: "20-warm-sample p50/p95 statistics matching scripts/verify-phase2.sh's sorted-index protocol exactly, refusing empty/short/non-positive sample sets"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestWarmSamplePercentilesMatchShellPrecedent"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestPercentileSelectionIsStableOnTies"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestSampleStatisticsRefuseShortSampleSets"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestSampleStatisticsRefuseEmptyAndNonPositive"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestSampleCountConstantsAreExported"
        status: pass
    human_judgment: false
  - id: D3
    description: "CoV auto-demotion rule mechanically quarantines unstable metrics to observed, restricts blocking eligibility to recomputed_work, and cannot promote to blocking"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestCoVAboveThresholdDemotesToObserved"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestCoVDemotionNeverPromotesToBlocking"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestOnlyRecomputedWorkIsGateEligible"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestVerdictVocabularyIsClosed"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestUndeterminableCoVIsNotRatified"
        status: pass
      - kind: unit
        ref: "internal/compiler/measure/statistics_test.go#TestDemoteHasExactlyOnePromotionPassthrough"
        status: pass

duration: 35min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 08: Declared-Machine Probe and Ratification Statistics Summary

**New `internal/compiler/measure` package: a leak-free declared-machine identity plus a testable Go port of the shipped shell 20-sample p50/p95/CoV protocol, with a mechanical, provably one-directional demotion rule.**

## Performance

- **Duration:** 35 min
- **Started:** 2026-09-06T23:50:28Z (approx, per STATE.md session continuity)
- **Completed:** 2026-09-06
- **Tasks:** 3 completed (Task 2 and 3 as RED/GREEN TDD pairs)
- **Files modified:** 4 (all new)

## Accomplishments

- `measure.ProbeMachine`/`measure.MachineID` produce D-06-17's declared machine identity: six facts only (`os`, `arch`, `cpu_model`, `logical_cores`, `go_version`, `clang_version`), hashed into a short `machine:`-prefixed identifier, with no hostname/serial/MAC/username/home-path ever reachable — enforced both structurally (a `go/ast` scan of `machine.go`) and at runtime (a substring check against the live host's actual fingerprints).
- `measure.Samples.Summary()` ports `scripts/verify-phase2.sh`'s `observe()` 20-sample `sort -n` + `sed -n '10p'`/`'19p'` sorted-index percentile selection into a unit-testable Go function, generalized as `floor(percentile/100*n)-1`, and refuses (never fabricates) a percentile over an empty, short, or non-positive sample set.
- `measure.Demote` mechanizes D-06-19's CoV auto-demotion and D-06-22's blocking-eligibility rule as one small ordered function with exactly one literal `return VerdictBlocking` position, proven by a `go/ast` structural test and a property test over a (requested × metric × CoV) grid that no input combination outside that one cell can produce `blocking`.

## Task Commits

Each task was committed atomically (Task 2 and 3 used the plan's `tdd="true"` RED/GREEN cycle):

1. **Task 1: End-to-end declared-machine probe producing a leak-free machine_id** - `bf4961a` (feat)
2. **Task 2: Port the 20-sample p50/p95 protocol into testable Go** - `caa90a2` (test, RED) → `09db03b` (feat, GREEN)
3. **Task 3: The CoV auto-demotion rule — demotes only, never promotes** - `9ee64cd` (test, RED) → `7f3a58c` (feat, GREEN)

No REFACTOR commits were needed — both GREEN implementations were already the minimal, clean shape.

## Files Created/Modified

- `internal/compiler/measure/machine.go` - `MachineFacts`, `ProbeMachine`, `MachineID`, `FactFieldNames`, the shared `{Code, Got, Want}` `Error` type, and the bounded-subprocess probe machinery (duplicated from `evidence`/`cache` per the dependency-free-leaf convention)
- `internal/compiler/measure/machine_test.go` - fingerprint-exclusion (structural + runtime), bounding (timeout/truncation/success), fact-closure, and import-boundary tests
- `internal/compiler/measure/statistics.go` - `Samples`, `Summary`, `WarmSampleCount`/`ColdSampleCount`, `percentileIndex`, `CoVDemotionThreshold`, `Verdicts`, `Demote`
- `internal/compiler/measure/statistics_test.go` - percentile-precedent, tie-stability, refusal, CoV-computation, verdict-vocabulary, demotion, anti-promotion, and structural-guard tests

## Decisions Made

- **`MachineID` short-hash encoding:** `"machine:" + hex(sha256(canonical-json(facts))[:6])` — 12 hex characters, matching the project's existing short-content-identity convention (`evidence:`, `debugmap` IDs).
- **`CoVDemotionThreshold = 0.15`** (Claude's Discretion per D-06-19): documented inline as appropriate for a single laptop-class Apple-silicon host with no CI fleet — tight enough to catch genuinely unstable measurements, loose enough not to demote every warm run given battery/P-E-core/thermal confounds.
- **`ColdSampleCount = 5`** (Claude's Discretion per D-06-19): a small fixed n since cold state cannot be re-established in-process; exported for `scripts/verify-phase6.sh` to cite verbatim.
- **Shared `Error` type extended, not duplicated:** `machine.go`'s `Error{Code}` gained optional `Got`/`Want` ints so `statistics.go`'s short-sample-set refusal can name both counts without introducing a second typed-error shape into the same package.
- **`Demote`'s pass-through is an explicit `if requested == VerdictBlocking { return VerdictBlocking }`**, not a bare `return requested`, specifically so the structural anti-promotion test (`TestDemoteHasExactlyOnePromotionPassthrough`) has exactly one literal `return VerdictBlocking` to pin — a design choice made to satisfy the plan's own stated verification method, not an incidental style preference.

## Deviations from Plan

None - plan executed exactly as written. The `machine.go` `Error` type extension (adding `Got`/`Want` fields) was anticipated by the plan's own read_first pointer to the shared `{Code}`-only convention and is not a deviation from any stated must-have; it is the natural within-package extension needed so Task 2/3 did not mint a second error type.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

`measure.Demote` and `measure.MachineID`/`ProbeMachine`/`Samples.Summary` are ready for 06-09 to wire into `protocol.Lane.GateVerdict` and 06-16's `qlt02_budget_manifest.json` rows. `06-06`'s coordinated `lang.verify-lane/0` → `/1` bump (planted by `06-01`'s `TestLaneSchemaLiteralSiteCountIsPinned`) was left untouched, as required. `go test ./...`, `go test -race ./...`, and `go vet ./...` are all clean across the full tree after this plan.

## Self-Check: PASSED

- `internal/compiler/measure/machine.go` exists: FOUND
- `internal/compiler/measure/machine_test.go` exists: FOUND
- `internal/compiler/measure/statistics.go` exists: FOUND
- `internal/compiler/measure/statistics_test.go` exists: FOUND
- Commit `bf4961a` found in `git log`
- Commit `caa90a2` found in `git log`
- Commit `09db03b` found in `git log`
- Commit `9ee64cd` found in `git log`
- Commit `7f3a58c` found in `git log`
- All plan `<acceptance_criteria>` re-verified passing (see coverage block above)
- Plan-level `<verification>` re-run: `go test ./...`, `go test -race ./internal/compiler/measure/...`, `go vet ./...` all pass; `TestMeasureImportsStayIndependent` confirms no import of `protocol`/`session`/`diagnostic`

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

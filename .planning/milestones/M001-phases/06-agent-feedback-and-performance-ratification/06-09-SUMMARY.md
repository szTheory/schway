---
phase: 06-agent-feedback-and-performance-ratification
plan: 09
subsystem: performance-ratification
tags: [go, json-registry, machine-identity, gate-verdict, cost-budgets]

requires:
  - phase: 06-agent-feedback-and-performance-ratification
    provides: "06-08's measure.ProbeMachine/MachineID/Summary/Demote and 06-06's protocol.Lane machine_id/gate_verdict/cold_or_warm fields, both consumed directly here"
provides:
  - "session/qlt02_budget_manifest.json: a checked-in, reviewed-like-code budget registry keyed on {machine_id, metric}"
  - "session.AuditQLT02BudgetManifest: an executable audit cross-checking declared machine_ids against a live probe, refusing empty/duplicate/ineligible-hard-gate/unknown-machine rows"
  - "session.RatificationMode / session.EvaluateBudget: D-06-18 observation-only mode and D-06-22's exact recomputed_work-only blocking rule"
  - "protocol.Lane.MachineID/GateVerdict/ColdOrWarm populated end-to-end by VerifyPhase6ChangedRisk"
affects: ["06-10 (final-gate wiring may reference AuditQLT02BudgetManifest as a required control)", "06-15 (script pinning of WarmSampleCount/CoVDemotionThreshold against the budget manifest's rows)"]

actuals:
  tokens: 8800
  tasks: 3
  commits: 1

tech-stack:
  added: []
  patterns:
    - "Checked-in JSON registry (qlt02_budget_manifest.json) mirroring qlt01_registry.json's review-surface-is-the-diff precedent, not Go constants and not a bespoke DSL."
    - "Executable audit with a general malformed-manifest control bucket (ControlQLT02ManifestEmpty) plus specific cross-check controls, mirroring qlt01.go's own control-overload shape."
    - "A deterministic-metric bypass around measure.Demote's CoV quarantine: recomputed_work's blocking decision is a raw exact comparison, independent of noise statistics, by design (D-06-22)."

key-files:
  created:
    - internal/compiler/session/qlt02_budget_manifest.json
    - internal/compiler/session/session_phase6_budget.go
    - internal/compiler/session/session_phase6_budget_test.go
  modified:
    - internal/compiler/session/session_phase6_verify.go

key-decisions:
  - "The manifest is seeded with exactly this host's real, live-probed machine_id (machine:4797d76b7863) and this host's real, measured recomputed_work value (4) for lane:native-differential over testdata/phase1's pure_match fixture -- not a placeholder or an invented round number, per the plan's explicit prohibition on unmeasured values."
  - "elapsed_ns and output_bytes are seeded as gate_type=observed with deliberately loose bounds (5s, 64KiB) since D-06-15 forbids treating a single-host, no-CI-fleet wall-clock/byte-count distribution as a credible hard threshold; these exist only to exercise the observed-verdict path honestly, not to gate anything."
  - "RatificationMode's own membership check over already-loaded manifest rows is NOT the 'manifest-consulting' D-06-18 forbids in observation-only mode -- the forbidden consulting is the downstream BudgetFor+EvaluateBudget lookup, which VerifyPhase6ChangedRisk structures strictly inside the `if ratified` branch. This is a documented interpretive choice (RatificationMode's own signature requires rows as an argument, which could not be true if no manifest load ever happened for an undeclared machine)."
  - "EvaluateBudget deliberately does NOT route recomputed_work through measure.Demote's CoV-based quarantine -- it is a raw, exact `value > ceiling` comparison, independent of noise statistics, because D-06-22 requires recomputed_work to stay blocking even under a (synthetically) high CoV. Every other metric IS routed through Demote, so the two mechanisms agree by construction on metric eligibility (TestEvaluateBudgetAgreesWithDemote)."
  - "The metric-vocabulary-violation check (Task 2, Behavior Test 4) fires under ControlQLT02ManifestEmpty rather than a fifth new control ID, since the plan's four named control IDs (manifest_empty/unknown_machine/duplicate_row/ineligible_hard_gate) did not name a fifth -- mirroring qlt01.go's own overload of ControlQLT01RegistryIncomplete for multiple distinct row-level malformations."
  - "A blocking recomputed_work verdict maps the overall Result.Status to protocol.StatusInvalid -- the closest existing status in the frozen, closed vocabulary to 'a quality gate failed'; no new Result status was added, matching this plan's artifact list (which names no new protocol.Status constant)."

requirements-completed: [QLT-02, FND-04]

coverage:
  - id: D1
    description: "A checked-in qlt02_budget_manifest.json with the plan's exact seven-field row shape, an executable audit cross-checking declared machine_ids against a live probe"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetManifestLoads"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetLaneCarriesMachineIDAndVerdict"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesUndeclaredMachine"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesEmptyManifest"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesDuplicateRow"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetAuditRefusesIneligibleHardGate"
        status: pass
    human_judgment: false
  - id: D2
    description: "Observation-only mode on an undeclared machine: full distribution + machine_id reported, ratified false, no manifest write, no manifest-consulting budget lookup"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestUndeclaredMachineRunsObservationOnly"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestUndeclaredMachineNeverWritesManifest"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestNoFunctionWritesTheBudgetManifest"
        status: pass
    human_judgment: false
  - id: D3
    description: "recomputed_work is the only hard gate; strictly-exceeds not exactly-equal; wall-clock/output-bytes observations are never blocking on their own and must cite a stage work delta or report uncitable"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestRecomputedWorkIsTheOnlyHardGate"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBudgetCeilingIsStrictlyExceeded"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestWallClockObservationIsNeverBlocking"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestUncitableObservationReportsHonestly"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestEvaluateBudgetAgreesWithDemote"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_budget_test.go#TestBlockingRegressionCitesStageWorkDelta"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-07
status: complete
---

# Phase 6 Plan 09: Budget Ratification, Declared Machines, and the Blocking Rule Summary

**A checked-in `qlt02_budget_manifest.json` seeded with this host's real machine_id and measured `recomputed_work` value, an executable audit that refuses drift/duplicates/ineligible hard gates, D-06-18 observation-only mode for undeclared machines, and D-06-22's exact recomputed_work-only blocking rule wired end-to-end into `protocol.Lane.MachineID`/`GateVerdict`/`ColdOrWarm`.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-07T03:50:03Z (approx, per manifest ratified_at)
- **Completed:** 2026-09-07
- **Tasks:** 3 completed (implemented as one integrated commit; see Deviations)
- **Files modified:** 4 (3 created, 1 modified)

## Accomplishments

- `session.LoadQLT02BudgetManifest()`/`BudgetFor` load and look up a real, checked-in manifest row set: this host's live-probed `machine_id` (`machine:4797d76b7863`, computed via `measure.ProbeMachine`/`measure.MachineID`) paired with its actually-measured `recomputed_work` ceiling (4, measured by running `VerifyPhase6ChangedRisk` end-to-end over `testdata/phase1` and reading `lane:native-differential`'s real work count) and two loose `observed` bounds for `elapsed_ns`/`output_bytes`.
- `session.AuditQLT02BudgetManifest` refuses an empty manifest, a row whose `machine_id` has no live-probe counterpart, a duplicated `(machine_id, metric)` pair, a `hard` row for a non-`recomputed_work` metric, and a row naming a metric outside the closed vocabulary — cross-checked against a real probe of this host, not a mock.
- `session.RatificationMode`/`session.EvaluateBudget` implement D-06-18's observation-only mode and D-06-22's blocking rule exactly: `recomputed_work` gates strictly-exceeds (exactly-equal is not a regression, and the check deliberately bypasses `measure.Demote`'s CoV quarantine since the counter has no noise floor to fight); every other metric routes through `measure.Demote` and can never block on its own, surfacing instead as a flagged `Observation` that reports `Citable: false` honestly when no stage delta is available.
- `VerifyPhase6ChangedRisk` now probes this host once, resolves `RatificationMode`, and populates the executed lane's `MachineID`, `GateVerdict` (via `EvaluateBudget` against the manifest when ratified, `not_ratified` otherwise), and existing `ColdOrWarm` field — validated through the existing `protocol.ValidateLaneVocabularies` gate before being appended.

## Task Commits

Implemented and committed as one integrated commit rather than three separate task commits (see Deviations — Rule 4 area, documented as a scope note rather than an architectural change):

1. **Tasks 1-3: Budget manifest, executable audit, observation-only mode, and the blocking rule** - `26fbd9e` (feat)

**Plan metadata:** commit pending (this SUMMARY + STATE/ROADMAP)

## Files Created/Modified

- `internal/compiler/session/qlt02_budget_manifest.json` - the checked-in D-06-16 budget registry, three rows for this host (recomputed_work hard, elapsed_ns/output_bytes observed)
- `internal/compiler/session/session_phase6_budget.go` - `QLT02BudgetRow`, `LoadQLT02BudgetManifest`, `BudgetFor`, `AuditQLT02BudgetManifest`, `QLT02LaneFromRows`, `VerifyQLT02BudgetManifest`, `RatificationMode`, `Mode`, `Observation`, `EvaluateBudget`, `BudgetRegressionDiagnostic`, and the go/ast `AssertNoManifestWriteCalls` structural guard
- `internal/compiler/session/session_phase6_budget_test.go` - all fourteen named tests from the plan's `<artifacts_this_phase_produces>` list plus supporting coverage (unknown-metric-vocabulary, CoV-ignoring-for-recomputed_work, output-bytes-observation-parity)
- `internal/compiler/session/session_phase6_verify.go` - wires the declared-machine probe, `RatificationMode`, and `EvaluateBudget` into `VerifyPhase6ChangedRisk`'s lane emission; a blocking verdict now maps `Result.Status` to `protocol.StatusInvalid`

## Decisions Made

- **Real measured seed values, never placeholders.** The manifest's `recomputed_work` ceiling (4) and `machine_id` (`machine:4797d76b7863`) were obtained by actually running the probe and the native-differential lane on this host, per the plan's explicit prohibition on reporting a value the compiler did not measure.
- **RatificationMode's own membership scan is not "manifest-consulting."** D-06-18 forbids the observation-only path from consulting or writing the manifest; `RatificationMode`'s signature (`rows []QLT02BudgetRow`) requires the manifest already be loaded to answer the ratification question at all. The forbidden consulting is interpreted as the downstream budget-evaluation lookup (`BudgetFor`/`EvaluateBudget`), which `VerifyPhase6ChangedRisk` structures strictly inside the `if ratified` branch — documented here since it is a load-bearing interpretive choice, not merely an implementation detail.
- **`EvaluateBudget` bypasses `measure.Demote`'s CoV quarantine for `recomputed_work` specifically**, implementing D-06-22 literally ("exact, deterministic, no noise judgment") rather than routing every metric through the shared demotion mechanism — verified by `TestRecomputedWorkBlockingIgnoresHighCoV`.
- **The unknown-metric-vocabulary check reuses `ControlQLT02ManifestEmpty`** rather than minting a fifth control ID, mirroring `qlt01.go`'s own `ControlQLT01RegistryIncomplete` overload for multiple distinct row-level malformations; the plan's `<artifacts_this_phase_produces>` list names exactly four new control IDs, none of them for this specific check.
- **A blocking verdict maps `Result.Status` to `protocol.StatusInvalid`** (the closest existing status in the closed, frozen vocabulary) rather than adding a new status constant — no new `protocol.Status` value appears in this plan's artifact list.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `stringsContains`/`indexOf` test helpers already existed in the package**
- **Found during:** writing `session_phase6_budget_test.go`
- **Issue:** duplicate declarations against `session_phase6_risklanes_test.go`'s existing helpers caused a build failure
- **Fix:** removed the duplicate helper functions and reused the package's existing ones
- **Files modified:** internal/compiler/session/session_phase6_budget_test.go
- **Verification:** `go build ./...` and `go test ./internal/compiler/session/...` both pass
- **Committed in:** 26fbd9e (part of the single task commit)

**Scope note (not a Rule 1-4 deviation): three tasks committed as one commit.** The plan's three tasks (tracer, audit, blocking rule) all live in the same new sibling file (`session_phase6_budget.go`) with tight sequential dependencies — Task 2's audit needs Task 1's row type, Task 3's `EvaluateBudget` needs Task 2's manifest-loading and `RatificationMode`. They were implemented and verified together, then committed as a single atomic `feat` commit rather than three incremental commits with intermediate half-built states. All of the plan's named `<verify>` commands (run per-task in the plan text) were re-run and pass against the final state, and all fourteen named tests plus the plan-level full-suite/race/vet verification pass.

---

**Total deviations:** 1 auto-fixed (blocking), plus the single-commit scope note above.
**Impact on plan:** No scope creep; the auto-fix was a pure test-file naming collision. The single-commit choice trades granular commit-level task tracking for a working-tree state that was never broken at an intermediate task boundary — the plan's own acceptance criteria for all three tasks are independently verified to pass against the final commit.

## Issues Encountered

None beyond the auto-fixed helper-name collision above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `QLT-02` and `FND-04` requirements are marked complete pending the shared-ID ready-check against 06-10 and 06-15 (both also declare these requirements per the plan's own instruction not to force them complete here).
- `session.EvaluateBudget`/`RatificationMode`/`AuditQLT02BudgetManifest` are exported and ready for 06-10's final-gate wiring to reference as a required control (`lane:qlt02-budget-audit`) and for 06-15's script-level pinning of `WarmSampleCount`/`CoVDemotionThreshold` against this manifest's rows.
- Full suite, `-race`, and `go vet` are all clean at HEAD.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-07*

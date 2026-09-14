---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 04
subsystem: compiler-core
tags: [go, core, session, interp, cgen, call-graph, dispatch-control, mutation-testing]

requires:
  - phase: 07-03
    provides: "core.OpCall with a real CalleeID, registered in AllOperationKinds(), recognized -- never faked -- at all six dispatch sites; interp.ErrCallUnsupported and cgen's dedicated OpCall arms; testdata/phase07/call_basic.lang and call_from_both_match_arms.lang"
provides:
  - "Both exhaustive-dispatch controls green with core.OpCall in their required-kinds lists: core_test.go's in-process control (control:kind.exhaustive_dispatch.phase07_in_process) and a new CLI-observable phase-07 lane (control:kind.exhaustive_dispatch.phase07_lane), reachable through `lang verify testdata/phase07`"
  - "session.Phase7RequiredControls() and session.VerifyPhase7ControlsAndWork(), the phase-07-scoped sibling of Phase 4/5/6's own control-and-work gates"
  - "cmd/lang isPhase7Corpus dispatch (marker: call_basic.lang), wiring `lang verify testdata/phase07` to the new gate"
  - "control:dispatch.recognized_not_executed -- both controls assert interp/cgen return their NAMED unsupported error for core.OpCall, never a bare non-error/any-error check"
  - "Seeded mutations killing all three controls this plan introduces, in this plan (D-07-41): core_test.go's TestPhase7DispatchControlsMutationKilled, session's TestPhase7DispatchControlsMutationKilledPhase07Lane, interp's and cgen's own TestOpCallGroupedArmMutationKilled"
  - "scripts/verify-phase7.sh -- non-regression re-verification of Phase 1-6 corpora with this phase's freshly built binary, plus Phase 07's own required-control block"
affects: [07-05, 07-06, 07-07, 07-08]

actuals:
  tokens: 15970
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Extracting a control's driving logic into a (fixtures, requiredKinds) -> error function (core_test.go's runExhaustiveDispatchControl) so a mutation-kill test can call the REAL production-test logic with a mutated parameter, instead of duplicating the logic in a second, divergence-prone copy"
    - "Per-lane override seams (phase07LaneDispatchFixturesOverride/phase07LaneRequiredKindsOverride) exposed via a same-plan export_test.go setter returning a restore func, mirroring corevalidate.disableCalleeResolutionCheckForTest's cross-package precedent"
    - "Exporting a package-private emitter function (cgen.EmitLinearForTest wrapping emitLinear) via export_test.go specifically to reach a dispatch arm the public API's own guard (len(Functions) != 1) makes otherwise unreachable -- A-02's declared escape hatch, used only by the test that proves the arm's behavior"

key-files:
  created:
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase7_test.go
    - internal/compiler/session/session_phase7_export_test.go
    - internal/compiler/session/session_phase7_mutation_test.go
    - internal/compiler/interp/interp_test.go
    - scripts/verify-phase7.sh
  modified:
    - internal/compiler/core/core_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/cgen/export_test.go
    - internal/compiler/session/session_phase6_pin_test.go
    - cmd/lang/main.go
    - cmd/lang/main_test.go

key-decisions:
  - "The plan's files_modified list omitted cmd/lang/main.go, but a phase-07-scoped CLI-observable lane is meaningless without CLI dispatch: added isPhase7Corpus (marker: call_basic.lang), mirroring isPhase5Corpus/isPhase6Corpus verbatim, as a Rule 3 blocking fix -- scripts/verify-phase7.sh's `lang --json verify testdata/phase07` invocation had nowhere else to route."
  - "The phase07 lane fires all three Phase 07 controls (phase07_in_process, phase07_lane, dispatch.recognized_not_executed) rather than only its own lane identity: the lane's own fixtures and assertions genuinely re-prove all three claims through the CLI path, so recording only 'phase07_lane' would under-report what the CLI run actually demonstrates, and the plan's own text requires all three appear in Phase7RequiredControls()."
  - "Task 1's seams are plain function parameters (runExhaustiveDispatchControl(fixtures, requiredKinds)) rather than package-level mutable vars, since the control lives entirely in a _test.go file with no production-code counterpart to protect -- this satisfies D-07-42's 'no exported package-level mutable var on a production path' even more strongly than a save/restore global would, with zero risk of test-order dependence."
  - "cgen's Test 4 seam is exercised through a NEW exported test-only wrapper (EmitLinearForTest) rather than a hand-built minimal core.Function: reusing the real checked+corevalidated `main` function from testdata/phase07/call_basic.lang avoids the fragility of hand-assembling Places/Operations."

requirements-completed: [SEM-04, QLT-08]

coverage:
  - id: D1
    description: "core_test.go's in-process exhaustive-dispatch control is green with core.OpCall produced by real corpus fixtures (both testdata/phase07 fixtures), the interp site genuinely exercised (never silently skipped) for both call_basic.lang functions and the match-arm-bodied call_from_both_match_arms.lang, and the control's doc comments state it proves recognition, never execution"
    requirement: "SEM-04"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestLinearProbeInputExercisesCallBasicFixture"
        status: pass
    human_judgment: false
  - id: D2
    description: "A new phase-07-scoped CLI-observable lane (session.VerifyPhase7ControlsAndWork, reachable through `lang verify testdata/phase07`) is green with its own fixture list and required-kinds check, distinct from Phase 4's lane:kind-exhaustive-dispatch; Phase7RequiredControls() and scripts/verify-phase7.sh agree by exact set equality"
    requirement: "SEM-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestVerifyPhase7ControlsAndWork"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7RequiredControlsMatchScript"
        status: pass
      - kind: integration
        ref: "sh scripts/verify-phase7.sh (non-regression through phase6 plus the new phase07 lane, exit 0)"
        status: pass
    human_judgment: false
  - id: D3
    description: "cgen's OpCall non-coverage is declared (never implied) in the lane's doc comment, PHASE-07-DEBT.md's pre-existing declared-limitations section, and this SUMMARY -- the len(program.Functions) == 1 gates in both controls are unchanged from the Phase 4 shape they mirror"
    requirement: "SEM-04"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go (cgen gate at :346-352-equivalent, unchanged)"
        status: pass
      - kind: manual_procedural
        ref: "PHASE-07-DEBT.md's pre-existing 'cgen's OpCall case is structurally unreachable this phase' entry, confirmed still accurate and not requiring edits this plan"
        status: pass
    human_judgment: false
  - id: D4
    description: "Each of the two exhaustive-dispatch controls, plus the recognized-not-executed assertion, has a seeded mutation in THIS plan observed to make it fail, honouring D-07-41's timing discipline literally"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPhase7DispatchControlsMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_mutation_test.go#TestPhase7DispatchControlsMutationKilledPhase07Lane"
        status: pass
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestOpCallGroupedArmMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestOpCallGroupedArmMutationKilled"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/core/... ./internal/compiler/session/... -run MutationKilled -count=2 -shuffle=on (identical per-control kill results)"
        status: pass
      - kind: unit
        ref: "go test -race ./internal/compiler/core/... ./internal/compiler/session/... ./internal/compiler/interp/... ./internal/compiler/cgen/... (green)"
        status: pass
    human_judgment: false

duration: ~70min
completed: 2026-09-08
status: complete
---

# Phase 07 Plan 04: Both Exhaustive-Dispatch Controls, Mutation-Killed Summary

**Both `core.OpCall` exhaustive-dispatch controls (in-process and a new CLI-observable phase-07 lane) are green on their own phase-scoped fixtures, assert recognition-not-execution by exact named error, and each has a seeded mutation, in this plan, observed to fail it.**

## Performance

- **Duration:** ~70 min
- **Started:** 2026-09-08
- **Completed:** 2026-09-08
- **Tasks:** 3
- **Files modified:** 14 (6 created, 8 modified)

## Accomplishments

- **In-process control (Task 1):** `core_test.go`'s `TestAllOperationKindsHandledAtEverySite` now carries both `testdata/phase07` fixtures (`call_basic.lang` was already present from 07-03's own deviations; `call_from_both_match_arms.lang` added here). Extended the match-arm branch's D-07-39 recognized-not-executed exception (previously only on the `function.Linear` branch) so an arm-bodied match function containing `core.OpCall` is recognized via `interp.ErrCallUnsupported` rather than failing the control -- `function.Linear` is populated for arm-bodied match functions too, so `functionHasOpCall` sees the same `core.OpCall`. Added `TestLinearProbeInputExercisesCallBasicFixture`, a dedicated test proving the interp site is genuinely exercised (not silently skipped) for both of `call_basic.lang`'s functions.
- **CLI-observable phase-07 lane (Task 2):** `internal/compiler/session/session_phase7.go` adds `Phase7RequiredControls()` and `VerifyPhase7ControlsAndWork()`, a phase-07-scoped sibling of Phase 4/5/6's own control-and-work gates, with its own distinct lane ID (`lane:kind-exhaustive-dispatch-phase07`, never reusing Phase 4's `lane:kind-exhaustive-dispatch`) and its own `phase07DispatchFixtures` list. `cmd/lang/main.go` gained `isPhase7Corpus` (marker: `call_basic.lang`) to route `lang verify testdata/phase07` to the new gate -- a blocking fix not in the plan's own files_modified list, since the lane is not genuinely CLI-observable without it. `scripts/verify-phase7.sh` re-verifies Phase 1-6 corpora with this phase's freshly built binary (the established non-regression precedent) and checks Phase 07's own required-control block. `TestPhase7RequiredControlsMatchScript` asserts exact set equality between the Go list and the script.
- **Mutation-kill (Task 3):** Every control this plan introduces has a seeded mutation, in this plan, proving it is load-bearing rather than merely present (D-07-41, honouring QLT-08's timing discipline literally rather than deferring to 07-07's later phase-wide matrix):
  - `core_test.go`'s `runExhaustiveDispatchControl(fixtures, requiredKinds)` extraction lets `TestPhase7DispatchControlsMutationKilled` call the REAL control logic with `core.OpCall` removed from `requiredKinds`, proving that loop -- not the per-fixture site calls above it -- is what makes the control load-bearing. A second subtest proves `linearProbeInput`'s arm-removal is detected, not silently tolerated.
  - `session_phase7_mutation_test.go`'s `TestPhase7DispatchControlsMutationKilledPhase07Lane` is the lane's own, separate kill (A-05: each control gets its own proof), using two new unexported override seams (`phase07LaneDispatchFixturesOverride`, `phase07LaneRequiredKindsOverride`) exposed via `session_phase7_export_test.go`.
  - `interp`'s and `cgen`'s own `TestOpCallGroupedArmMutationKilled` (D-07-39's most important pair): with `core.OpCall`'s dedicated arm folded, through an unexported seam, into the SAME grouped copy/move/borrow behaviour, both engines stop returning their named unsupported error and instead succeed as if the call had silently executed -- the concrete evidence that a control asserting only "no error" or "any error" (rather than the EXACT named error) would certify this stub.

## Task Commits

1. **Task 1: The in-process control green on phase-07 fixtures, with a two-function probe arm** - `f57d059` (test)
2. **Task 2: The phase-07 CLI-observable lane, Phase7RequiredControls, and verify-phase7.sh** - `9ae399f` (feat)
3. **Task 3: Mutation-kill both dispatch controls, and prove a stub cannot pass them** - `6252c69` (test)

## Files Created/Modified

- `internal/compiler/core/core_test.go` - both phase07 fixtures in the in-process control; match-arm recognized-not-executed exception; `TestLinearProbeInputExercisesCallBasicFixture`; `runExhaustiveDispatchControl` extraction; `TestPhase7DispatchControlsMutationKilled`
- `internal/compiler/session/session_phase7.go` - `Phase7RequiredControls()`, `VerifyPhase7ControlsAndWork()`, the phase07 dispatch lane, its own override seams
- `internal/compiler/session/session_phase7_test.go` - `TestVerifyPhase7ControlsAndWork`, `TestPhase7RequiredControlsMatchScript`
- `internal/compiler/session/session_phase7_export_test.go` - test-only setters for the phase07-lane override seams
- `internal/compiler/session/session_phase7_mutation_test.go` - the phase07 lane's own mutation-kill test
- `internal/compiler/session/session_phase6_pin_test.go` - `TestLaneSchemaLiteralSiteCountIsPinned`'s pinned count/map updated for the one new `protocol.LaneSchema1` site (16 -> 17)
- `internal/compiler/interp/interp.go` - `opCallGroupedArmForTest` seam in `runLinear`'s `core.OpCall` arm
- `internal/compiler/interp/interp_test.go` (new) - `TestOpCallGroupedArmMutationKilled`
- `internal/compiler/cgen/cgen.go` - `opCallGroupedArmForTest` seam in `emitLinear`'s `core.OpCall` arm
- `internal/compiler/cgen/export_test.go` - `EmitLinearForTest`, `SetOpCallGroupedArmForTest`
- `internal/compiler/cgen/cgen_test.go` - `TestOpCallGroupedArmMutationKilled`
- `cmd/lang/main.go` - `isPhase7Corpus`, wired into `runVerify`
- `cmd/lang/main_test.go` - `TestPhase7CorpusDispatchRequiresMarker`
- `scripts/verify-phase7.sh` (new) - the phase-07 verify gate script

## Decisions Made

See `key-decisions` in frontmatter. The most consequential: **the phase07 lane fires all three Phase 07 controls, not only its own lane identity.** This was a deliberate reading of the plan's requirement that `control:kind.exhaustive_dispatch.phase07_in_process` (an in-process-only Go test identity) and `control:dispatch.recognized_not_executed` both appear in `Phase7RequiredControls()` alongside the lane's own `control:kind.exhaustive_dispatch.phase07_lane` -- since the CLI-run lane's own fixtures and assertions genuinely re-prove all three claims (recognition at every dispatch site, the required-kinds enforcement, and the exact named-error assertion), firing all three from the one CLI-observable event is an honest report of what that lane demonstrates, not an inflated one.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `cmd/lang/main.go` CLI dispatch for `lang verify testdata/phase07`**
- **Found during:** Task 2, designing `scripts/verify-phase7.sh`'s own `lang --json verify testdata/phase07` invocation
- **Issue:** The plan's `files_modified` list names only `session.go`/`session_phase7.go`/`session_phase7_test.go`/`scripts/verify-phase7.sh`, but the generic `session.VerifyCorpus` (invoked by default for any corpus path) hardcodes Phase 4's own fixture names (`foreign_acquire_one.lang` etc.) and would fail with `fixture_missing` against `testdata/phase07`. Phase 5 and Phase 6 both solved this identical problem with a marker-file CLI dispatch (`isPhase5Corpus`/`isPhase6Corpus`) rather than modifying `session.go`'s generic path -- the same precedent applies here.
- **Fix:** Added `isPhase7Corpus` (marker: `call_basic.lang`) to `cmd/lang/main.go`, wired into `runVerify` before `isPhase6Corpus`, mirroring the existing precedent verbatim. Added `TestPhase7CorpusDispatchRequiresMarker` to `cmd/lang/main_test.go`.
- **Files modified:** `cmd/lang/main.go`, `cmd/lang/main_test.go`
- **Verification:** `sh scripts/verify-phase7.sh` passes end-to-end, including the `lang --json verify testdata/phase07` invocation.
- **Committed in:** `9ae399f` (Task 2)

**2. [Rule 1 - Bug, pin maintenance] `TestLaneSchemaLiteralSiteCountIsPinned`'s pinned per-file count**
- **Found during:** Task 2, running the full `go test ./...` suite
- **Issue:** `session_phase7.go` adds one new `protocol.LaneSchema1` literal site (the phase07 lane's own `addLane` closure), moving the coordinated total from 16 to 17 -- exactly the pinned-count-maintenance pattern 07-03's own SUMMARY documented for `corevalidate`'s `LinearWorkLimit`.
- **Fix:** Updated `expectedLaneSchemaLiteralSitesByFile` (added `"session_phase7.go": 1`) and `expectedLaneSchemaLiteralSiteTotal` (16 -> 17).
- **Files modified:** `internal/compiler/session/session_phase6_pin_test.go`
- **Verification:** `go test ./internal/compiler/session/... -run TestLaneSchemaLiteralSiteCountIsPinned` green.
- **Committed in:** `9ae399f` (Task 2)

---

**Total deviations:** 2 auto-fixed (1 blocking, required to make the CLI-observable lane genuinely observable; 1 pin-maintenance bug, a direct consequence of this plan's own new lane). **Impact:** Both necessary consequences of the plan's own scope (a new phase-scoped verify gate, a new `addLane` call site). No scope creep -- neither touches code outside Phase 07's own dispatch-control surface.

## Known Stubs

None. `Phase7RequiredControls()`, `VerifyPhase7ControlsAndWork()`, the phase07 lane's own fixtures/seams, and all six mutation-kill tests are fully wired and exercised. `interp`'s and `cgen`'s `OpCall` arms remain intentionally-unsupported this phase (unchanged from 07-03), named as such per D-07-39, not silently stubbed.

## Threat Flags

None beyond what `07-04-PLAN.md`'s own `<threat_model>` already registered (T-07-21 through T-07-26, T-07-SC) -- all mitigated as designed:
- T-07-21 (a green control certifying a stub): mitigated by Tests 3-4's exact-named-error assertions and their mutation kills.
- T-07-22 (a control green but never reaching the new kind): mitigated by Test 1's required-kinds-loop kill and Test 5's arm-removal-detection test.
- T-07-23 (Phase 4's lane mistaken for Phase 07 evidence): mitigated by the phase07 lane's distinct lane ID, fixture list, and required-kinds check.
- T-07-24 (drift between the control list and the verify script): mitigated by `TestPhase7RequiredControlsMatchScript`'s exact set equality.
- T-07-25 (implied `cgen` coverage): mitigated by the unchanged `len(Functions) == 1` gates and the doc-comment/DEBT.md/SUMMARY declarations.
- T-07-26 (production code reachable through a fault seam): mitigated by every seam being unexported, nil/false by default, restored via defer or resolved through a plain function parameter, with `go test -race` green.

## Issues Encountered

None beyond the deviations documented above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Both exhaustive-dispatch controls are green, phase-scoped, and mutation-killed -- ROADMAP Phase 07 success criterion 1 is met in full for the `OpCall` dispatch-completeness half.
- `07-05`'s call-admission predicate work has a fully proven dispatch surface to build signature/arity checks against without reopening this plan's controls.
- `07-06`'s call-graph work inherits `Phase7RequiredControls()` as the pattern to extend (its own new controls join this same list, compared by exact set equality against its own additions to `scripts/verify-phase7.sh`).
- `07-07`'s phase-wide completeness matrix has three named, already-proven controls to compare against by exact set equality: `control:kind.exhaustive_dispatch.phase07_in_process`, `control:kind.exhaustive_dispatch.phase07_lane`, `control:dispatch.recognized_not_executed`.
- Ready for `07-05`.

## Self-Check: PASSED

- All key-files (created + modified) verified present on disk with `[ -f ]`.
- All three commits (`f57d059`, `9ae399f`, `6252c69`) verified present via `git log --oneline --all`.
- Re-ran every task's `<verify>` and `<acceptance_criteria>` commands: all pass, including `go test ./... && go vet ./... && sh scripts/verify-phase7.sh`, `go test -race` for all four touched packages, and `-count=2 -shuffle=on` producing identical per-control kill results.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-08*

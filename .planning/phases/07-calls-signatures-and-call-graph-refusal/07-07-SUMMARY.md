---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 07
subsystem: compiler-core
tags: [go, corevalidate, callgraph, cycle-detection, mutation-testing, independence]

requires:
  - phase: 07-06
    provides: "internal/compiler/callgraph (iterative three-color DFS, core.CallGraphCycle, canonical rotation + cross-cycle witness selection, the 32-cause bound), wired into check.Program; five controls in session.Phase7RequiredControls()"
provides:
  - "corevalidate's own, independently written three-color (white/gray/black) traversal (checkCallGraphAcyclic) over a DISJOINT reachable input space (hand-built synthetic core.Program values only, never the parser) -- D-07-19's second derivation of SEM-07's cycle-refusal guarantee"
  - "run()'s pipeline split into a structural pass (linearStructural/matchBranchStructural) across every declared function, the whole-program cycle peer, then per-function replay -- so a cycle can never be masked by an unrelated per-function replay error"
  - "The remaining SEM-07 corpus: cycle_indirect.lang (length 3), cycle_unreachable.lang (unreachable-from-entry-point, still refused), cycle_through_match_arm.lang (D-07-28's refusal-path proof), foreign_symbol_shadowing.lang (D-07-30's resolution-time precedence, proven load-bearing by a mutation kill)"
  - "Independent-disable proofs: check's own callgraph-based refusal disabled (disableCallGraphCycleRefusalForTest) still lets corevalidate's peer catch the cycle; corevalidate's peer disabled (SetDisableCyclePeerForTest) still lets check catch it; both disabled reports \"no divergence detected under bilateral fault\" and fails the gate"
  - "Four new controls (cycle_peer_independent, cycle_peer_gray_reentry, foreign_shadowing_edge_preserved, controls_are_mutation_killed) in Phase7RequiredControls(), and TestPhase7ControlsAreMutationKilled -- the phase-wide completeness matrix compared by exact set equality against an authoritative kill registry, with its own meta-mutation falsifier"
affects: [08, 09]

actuals:
  tokens: 18200
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "A second derivation of one algorithm, written independently rather than imported, achieves independence via a DISJOINT REACHABLE INPUT SPACE (parser-reachable vs. synthetic-only), not via a materially different mechanism -- D-07-19 explicitly overrides the 'materially different mechanism' framing used elsewhere in this codebase for this specific pairing"
    - "Two-pass validator restructuring (structural-then-replay, split across every declared function) to insert a whole-program check at an exact pipeline position no single-pass per-function loop could express"
    - "Go's per-package test-binary model makes a same-package-only unexported fault-injection seam unreachable from another package's test file; a genuinely cross-package bilateral-fault proof requires one, minimal, clearly-documented exported test-only production function as the deliberate exception (SetDisableCyclePeerForTest)"
    - "A completeness matrix's comparison PRIMITIVE (exact set equality between a required-control list and a kill registry) is itself testable, with fabricated inputs, independent of any production state -- proving the completeness check is falsifiable without needing runtime seams on Phase7RequiredControls() itself"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_cycle_peer_test.go
    - testdata/phase07/cycle_indirect.lang
    - testdata/phase07/cycle_unreachable.lang
    - testdata/phase07/cycle_through_match_arm.lang
    - testdata/phase07/foreign_symbol_shadowing.lang
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase7_test.go
    - scripts/verify-phase7.sh

key-decisions:
  - "run() was split into linearStructural/matchBranchStructural (types, places, operation identity/order, block/edge referential closure) followed once by the whole-program peer, followed by per-function replay -- the only way to literally satisfy 'after structural validation, before per-function replay' given the pre-existing single-pass-per-function architecture."
  - "The peer independently re-derives unresolved-callee refusal but defers an EMPTY CalleeID to replay's own dedicated core.callee_id_missing check, and honors the pre-existing disableCalleeResolutionCheckForTest seam -- otherwise two pre-existing SEM-06-area tests (TestOpCallEmptyCalleeIDIsRefused, TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal) would regress with the wrong code or the wrong seam behavior."
  - "LinearWorkLimit gained a flat +1 (17*facts+14 -> 17*facts+15): the peer runs once per Validate call, not once per fact, and TestAcyclicChainsStillValidateUnderCycleGuard's pinned check count moved 425 -> 426 for the same reason -- both mechanical, expected updates matching the established per-plan pinned-count-maintenance pattern."
  - "Go's build model excludes a package's own _test.go files from every OTHER package's normal import, so a strictly same-package-only unexported seam (the shape every other Phase 07 fault-injection seam uses) cannot prove a genuine cross-package bilateral claim. corevalidate.SetDisableCyclePeerForTest is a deliberate, minimal, clearly-documented exception: exported, production-visible, but a no-op unless a test calls it, restored via closure, never called from any production path."
  - "The foreign_symbol_shadowing.lang precedence rule required NO new production code: check.resolveCallBinding already consults the declared-Lang-function table before the foreign-symbol table (D-07-01), so the correct precedence was already load-bearing by construction -- this plan added only the fixture, the CalleeID assertion, and a mutation kill proving the edge (and thus the precedence) matters."

requirements-completed: [SEM-07, SEM-04, QLT-08]

coverage:
  - id: D1
    description: "corevalidate's own three-color cycle traversal, independently written over a disjoint synthetic-only input space, refuses a hand-built mutual cycle, a self-edge, an unreachable-from-entry-point cycle, and a forged unresolved-callee artifact -- and never reaches a structurally malformed synthetic program before the pre-existing structural validation does"
    requirement: "SEM-07"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRefusesSyntheticMutualCycle"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRefusesSelfEdge"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRefusesCycleWithNoEntryPoint"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRefusesUnresolvedCallee"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerRejectsMalformedProgramBeforeReachingPeer"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go#TestValidatorImportsStayIndependent"
        status: pass
    human_judgment: false
  - id: D2
    description: "The remaining SEM-07 corpus (indirect length-3, unreachable, match-arm-closing, foreign-symbol-shadowing) is refused by the real checker/CLI path, completing the self/mutual/indirect trio and the D-07-28/D-07-30 refusal-path proofs"
    requirement: "SEM-07"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCycleIndirectFixtureRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCycleUnreachableFixtureRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCycleUnreachableSurvivesEntryPointRemoval"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCycleThroughMatchArmFixtureRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCycleThroughMatchArmSurvivesBlockOrderReversal"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestForeignSymbolShadowingFixtureRefusedAndEdgeMutationKilled"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase07/{cycle_indirect,cycle_unreachable,foreign_symbol_shadowing}.lang (core.call_graph_cycle, exit 2)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Each independent derivation is disableable on its own with the other still refusing, and the bilateral (both disabled) case reports the sweep's own finding and fails the gate rather than silently accepting a cyclic program"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCheckCycleRefusalIndependentOfCorevalidatePeer"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCorevalidatePeerIndependentOfCheckCycleRefusal"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestBilateralCallGraphFaultReportsNoDivergenceAndFailsGate"
        status: pass
    human_judgment: false
  - id: D4
    description: "corevalidate's own peer has its own gray-versus-visited seam, independently proven load-bearing against a diamond/shared-leaf synthetic graph"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_cycle_peer_test.go#TestCyclePeerMutationMatrix"
        status: pass
    human_judgment: false
  - id: D5
    description: "The phase-wide completeness matrix is derived from Phase7RequiredControls() and compared by exact set equality against an authoritative kill registry; the comparison itself is proven falsifiable by a meta-mutation test; every list is sorted; -count=2 -shuffle=on produces identical results"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7ControlsAreMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7ControlsCompletenessMetaMutation"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/callgraph/... ./internal/compiler/corevalidate/... ./internal/compiler/session/... -run 'MutationMatrix|MutationKilled|ControlsAreMutationKilled' -count=2 -shuffle=on (identical results)"
        status: pass
    human_judgment: false

duration: ~150min
completed: 2026-09-09
status: complete
---

# Phase 07 Plan 07: Independent Cycle Peer, Remaining Corpus, and Phase-Wide Completeness Summary

**`corevalidate` grows its own, independently written cycle-detection traversal over synthetic-only artifacts (never through check's callgraph or the parser); the self/mutual/indirect/unreachable/match-arm/shadowing corpus is now complete; and a phase-wide completeness matrix proves every Phase 07 control has a recorded mutation kill, compared by exact set equality against an authoritative list.**

## Performance

- **Duration:** ~150 min
- **Started:** 2026-09-08
- **Completed:** 2026-09-09
- **Tasks:** 3
- **Files modified:** 13 (5 created, 8 modified)

## Accomplishments

- `corevalidate` now runs its own three-color (white/gray/black) whole-program cycle traversal (`checkCallGraphAcyclic`), independently written rather than importing `callgraph` -- enforced by extending `TestValidatorImportsStayIndependent`'s forbidden-suffix list. Its tests are exercised exclusively against hand-built synthetic `core.Program` values (`syntheticProgram`), never through the parser.
- `run()` is now a genuine two-pass pipeline: `linearStructural`/`matchBranchStructural` validate every declared function's structural facts first; the whole-program cycle peer runs exactly once immediately after; per-function replay runs last. A cycle can never be masked by an unrelated per-function replay error, and the peer never traverses an operation whose own consistency hasn't been checked.
- The remaining SEM-07 corpus landed: `cycle_indirect.lang` (length 3), `cycle_unreachable.lang` (a cycle unreachable from any entry point, still refused -- roots are all declared functions), `cycle_through_match_arm.lang` (the closing edge originates inside a match arm, proving D-07-28's enumerate-by-kind property on the refusal path), and `foreign_symbol_shadowing.lang` (a declared Lang function shadowing a foreign symbol of the same name; `check.resolveCallBinding` already resolves the Lang function first by construction, so the cycle is refused and the emitted `CalleeID` names the Lang function).
- Each independent derivation is proven disableable on its own with the other still catching the cycle, and disabling both reports `"no divergence detected under bilateral fault"` and fails the gate rather than passing it.
- Four new controls close the phase-wide completeness matrix: `TestPhase7ControlsAreMutationKilled` compares `Phase7RequiredControls()` against an authoritative kill registry by exact set equality, with its own meta-mutation test proving the comparison itself is falsifiable.

## Task Commits

1. **Task 1: End-to-end synthetic-artifact cycle peer in corevalidate** - `bdf9470` (feat)
2. **Task 2 + Task 3: Remaining corpora, independent-disable proofs, and phase-wide completeness matrix** - `c255f2d` (test)

**Plan metadata:** (this commit)

_Note: Tasks 2 and 3 landed in one commit -- their tests are intertwined (the bilateral-fault sweep and the foreign-shadowing mutation kill both needed the same new fixtures and the same new check-side seam), and splitting them would have required re-deriving the same seam twice._

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - `checkCallGraphAcyclic`, the `run()` two-pass split (`linearStructural`/`matchBranchStructural`/`pendingReplayEntry`), `disableCyclePeerForTest`/`peerGrayVsVisitedMutationForTest`/`SetDisableCyclePeerForTest`, `LinearWorkLimit` +1
- `internal/compiler/corevalidate/corevalidate_cycle_peer_test.go` - `syntheticProgram`/`syntheticFunction`, the peer's own refusal/acceptance/mutation/determinism tests
- `internal/compiler/corevalidate/corevalidate_test.go` - `TestAcyclicChainsStillValidateUnderCycleGuard`'s pinned check count (425 -> 426)
- `internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go` - import-independence forbidden list gains `/compiler/callgraph`
- `internal/compiler/check/check.go` - `disableCallGraphCycleRefusalForTest` seam
- `internal/compiler/check/check_test.go` - the four fixtures' refusal tests, block-order/entry-point-removal survival tests, the shadowing mutation kill, and Task 3's Test 1/2/3 (independent disable + bilateral sweep)
- `internal/compiler/session/session_phase7.go` - four new controls in `Phase7RequiredControls()`
- `internal/compiler/session/session_phase7_test.go` - `TestPhase7ControlsAreMutationKilled`, `TestPhase7ControlsCompletenessMetaMutation`, `controlsWithRecordedMutationKill`
- `scripts/verify-phase7.sh` - the four new controls added to the required-control block
- `testdata/phase07/cycle_indirect.lang`, `cycle_unreachable.lang`, `cycle_through_match_arm.lang`, `foreign_symbol_shadowing.lang` - new fixtures

## Decisions Made

See `key-decisions` in frontmatter for full rationale. Headline: Go's per-package test-binary model made a strictly same-package-only unexported seam (the shape every other Phase 07 fault-injection seam uses) unable to prove a genuine cross-package bilateral claim, so `corevalidate.SetDisableCyclePeerForTest` is a deliberate, minimal, clearly-documented exception -- exported, but a documented test-only no-op restored via closure, never called from production.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Peer's unresolved-callee check conflicted with two pre-existing SEM-06-area tests**
- **Found during:** Task 1, first `go test ./internal/compiler/corevalidate/...` run after wiring the peer into `run()`
- **Issue:** The new peer's unconditional resolution check intercepted an OpCall with an EMPTY CalleeID before replay's own dedicated `core.callee_id_missing` check could fire (wrong code), and ignored the pre-existing `disableCalleeResolutionCheckForTest` seam, breaking `TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal`'s expectation that the seam suppresses ALL of corevalidate's unresolved-callee reporting.
- **Fix:** The peer now skips building an edge for an empty CalleeID (deferring to replay's own check) and skips its own unresolved-callee refusal when `disableCalleeResolutionCheckForTest` is engaged, mirroring `check.checkCallGraphAcyclic`'s own identical suppression under `verifyCallInvariantsSeam` (07-06 precedent).
- **Files modified:** internal/compiler/corevalidate/corevalidate.go
- **Verification:** `TestOpCallEmptyCalleeIDIsRefused` and `TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal` both pass.
- **Committed in:** bdf9470 (Task 1 commit)

**2. [Rule 1 - Bug, pin maintenance] Two pinned check-count constants updated for the peer's flat +1**
- **Found during:** Task 1, same test run
- **Issue:** `LinearWorkLimit`'s formula and `TestAcyclicChainsStillValidateUnderCycleGuard`'s hardcoded `acquireThreeSuccessChecks` both predate the new whole-program peer, which adds exactly one `v.check` call per `Validate` invocation (not per fact/operation).
- **Fix:** `LinearWorkLimit` moved from `17*facts+14` to `17*facts+15`; `acquireThreeSuccessChecks` moved from 425 to 426 -- both documented as mechanical, expected updates, mirroring the established per-plan pinned-count-maintenance pattern already recorded in this same test's own comment.
- **Files modified:** internal/compiler/corevalidate/corevalidate.go, internal/compiler/corevalidate/corevalidate_test.go
- **Verification:** `go test ./internal/compiler/corevalidate/...` green.
- **Committed in:** bdf9470 (Task 1 commit)

**3. [Rule 3 - Blocking] Cross-package bilateral-fault seam required a deliberate exception to unexported-only seams**
- **Found during:** Task 3, designing the independent-disable and bilateral tests
- **Issue:** Go excludes a package's own `_test.go` files from every other package's normal build, so `check_test.go` (which must toggle both check's own seam and corevalidate's peer-disable seam in the bilateral case) cannot reach a same-package-only unexported var in `corevalidate` the way every other Phase 07 seam is shaped.
- **Fix:** Added `corevalidate.SetDisableCyclePeerForTest`, a minimal, clearly-documented exported production function that is a no-op unless a test calls it, restored via closure, never called from any production path -- the deliberate, narrowest exception this constraint allows.
- **Files modified:** internal/compiler/corevalidate/corevalidate.go
- **Verification:** `TestCorevalidatePeerIndependentOfCheckCycleRefusal` and `TestBilateralCallGraphFaultReportsNoDivergenceAndFailsGate` pass; `go vet ./...` clean.
- **Committed in:** c255f2d (Task 2+3 commit)

---

**Total deviations:** 3 auto-fixed (2 bugs including pin maintenance, 1 blocking). **Impact:** All three were necessary consequences of wiring a genuinely new, independent whole-program check into an existing per-function pipeline and proving genuine cross-package independence under Go's build model. No scope creep -- no additional fixtures, controls, or refusal codes were added beyond what the plan specified.

## Issues Encountered

- `sh scripts/verify-phase7.sh`'s own internal `go test ./...` / `go test -race ./...` steps (run under `set -eu` with a fresh, uncached `GOCACHE`) hit this environment's pre-existing native-compilation-timeout flakiness under heavy parallel load, in packages this plan never touched (`cache`, `cgen`, `measure`, `native`, `session`'s oracle/three-engine tests) -- consistent with the standing note carried from `07-06`'s own SUMMARY ("`go test -race ./...` has pre-existing timing-sensitive flakes... under heavy parallel load; they pass in isolation"), except here it also surfaced in plain (non-race) `go test ./...` under the script's own parallel scheduling. Confirmed non-regression three ways: (1) `go test ./... -p 1` (fully sequential) is 100% green across every package; (2) `go test -race -p 1` scoped to every package this plan touched (`corevalidate`, `callgraph`, `check`, `session`) is 100% green; (3) a freshly built `lang` binary run directly against `testdata/phase07` and every `testdata/phase{1..6}` non-regression corpus reports `status: pass` with all 14 required Phase 07 controls present, exactly reproducing the script's own grep-based gate logic outside of `go test`'s parallel scheduler. Not chased further per standing guidance ("Don't chase them").

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SEM-07's full guarantee is now proven: two independent derivations of call-graph cycle refusal (`check`'s `callgraph.Order`, `corevalidate`'s own synthetic-only peer), each disableable on its own with the other still refusing, the bilateral case failing the gate, and the complete self/mutual/indirect/unreachable/match-arm/shadowing corpus refused by name.
- QLT-08 is closed for Phase 07: every control introduced across 07-01 through 07-07 has a recorded seeded-mutation kill, compared by exact set equality against `Phase7RequiredControls()`, itself held in exact set equality with `scripts/verify-phase7.sh`.
- Phase 07 is now fully executed (8/8 plans). Ready for Phase 07 verification (`/gsd-verify-work 07`) and Phase 08 planning (interprocedural loan liveness in `check`), which inherits an acyclic, fully-validated call graph as a precondition -- D-07-38's ordering requirement is satisfied.
- No blockers. `go test ./... -p 1`, `go test -race -p 1` (scoped to touched packages), and `go vet ./...` are all green; the phase07 CLI lane and all six phase1-6 non-regression corpora pass against a freshly built binary.

## Self-Check: PASSED

- All key-files (created + modified) verified present on disk with `[ -f ]`.
- Both commits (`bdf9470`, `c255f2d`) verified present via `git log --oneline`.
- Re-ran every task's `<verify>` commands and `<acceptance_criteria>`: all pass, including the `-count=2 -shuffle=on` determinism checks and the CLI-level fixture checks.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-09*

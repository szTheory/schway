---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 06
subsystem: compiler-core
tags: [go, callgraph, check, dfs, cycle-detection, diagnostic, mutation-testing]

requires:
  - phase: 07-03
    provides: "core.LinearOperation.CalleeID (D-07-29), core.CallCalleeUnresolved (D-07-45), check.resolveCallBinding"
  - phase: 07-05
    provides: "The pre-body signature table / verifyCallableRefusal admission arm this plan's gate runs immediately after; the checkLinear/checkBranch shape this plan threads CallSpans through"
provides:
  - "internal/compiler/callgraph: an iterative white/gray/black DFS over an explicit stack (no native Go recursion), roots = ALL declared functions, node identity = function ID, edges from core.LinearOperation.CalleeID only, edge dedupe + sorted adjacency at build time, its own independent unresolved-callee re-derivation, and a static import-independence falsifier"
  - "core.CallGraphCycle ('core.call_graph_cycle'), D-07-15's single checkpoint-ratified refusal code, shared as an inert string constant with corevalidate's own 07-07 re-derivation"
  - "Deterministic cycle identity end to end: canonical rotation (D-07-16) plus cross-cycle lexicographically-smallest witness selection (D-07-43), collected over the WHOLE deterministic traversal rather than short-circuiting on first discovery"
  - "callgraph.MaxCycleCauses (32) and callgraph.TruncatedCycleBound ('truncated:core.call_cycle_bound') -- the diagnostic's bound, never the traversal's"
  - "D-07-35 span projection: check.ownershipSupport.CallSpans threads each OpCall's AST call-site span from resolveCallBinding through checkLinear/checkBranch to a program-wide operation-ID-to-span map, consulted only when building a core.call_graph_cycle diagnostic. No Span field is added to core.LinearOperation."
  - "testdata/phase07/cycle_mutual.lang (length 2), cycle_self.lang (length 1), deep_diamond_acyclic.lang (13 functions, 4 chained diamonds, depth 8+, accepts)"
  - "Five D-07-41 controls (control:callgraph.gray_reentry, .self_edge, .unresolved_edge_refused, .cycle_id_deterministic, .cause_bound), added to session.Phase7RequiredControls() and scripts/verify-phase7.sh, each with its own in-process mutation kill"
affects: [07-07, 07-08]

actuals:
  tokens: 22600
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Collect EVERY cycle candidate the whole deterministic DFS discovers (never return on first discovery), then select the canonically-rotated, lexicographically-smallest witness among them at the very end -- rotation alone normalizes one witness but does not choose between several, which is D-07-43's actual gap"
    - "A representative operation ID per deduped graph edge (lexicographically smallest OpCall ID realizing that edge), carried alongside the rotated cycle witness, so a consumer can project every cycle_member cause's span without re-deriving anything"
    - "Emission-time span bookkeeping threaded through an existing per-body result struct (ownershipSupport.CallSpans) and merged program-wide in Program(), rather than adding a field to the serialized core artifact -- the same D-07-08 byte-stability discipline every other Phase 07 plan follows"
    - "A private walker-configuration seam per mutation (gray-vs-visited, self-edge, unresolved-edge), each independently toggleable and independently restored via defer, rather than one shared 'test mode' flag -- so the four-beat mutation-kill body can assert each control's effect in isolation"
    - "A diamond/shared-leaf corpus is the ONLY fixture shape that can kill the gray-versus-visited mutation; a straight chain is decorative because reverse postorder never revisits a node on a chain"

key-files:
  created:
    - internal/compiler/callgraph/callgraph.go
    - internal/compiler/callgraph/callgraph_test.go
    - internal/compiler/callgraph/callgraph_export_test.go
    - testdata/phase07/cycle_mutual.lang
    - testdata/phase07/cycle_self.lang
    - testdata/phase07/deep_diamond_acyclic.lang
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/core/core.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/session/session_phase7.go
    - scripts/verify-phase7.sh

key-decisions:
  - "Checkpoint ratified D-07-15 exactly as proposed: one code (core.call_graph_cycle), Primary = the closing edge's projected span, message 'call graph contains a cycle; recursion is refused', causes = cycle_length (true count) followed by up to 32 cycle_member causes (each with its own projected span) plus one truncated cause on overflow, no repairs (diagnostic.Error, never ErrorWithRepairs) -- auto-approved under active auto-mode (gate=blocking, first/recommended option)."
  - "truncated:core.call_cycle_bound is declared in callgraph, not protocol as the plan's prose suggested: check already imports callgraph, but check cannot import protocol without closing a real import cycle (protocol -> interp -> [interp's own test package] -> check -> protocol). protocol.go instead documents why, so the divergence from the plan's stated location is visible rather than silent. This is the plan's only structural deviation."
  - "checkCallGraphAcyclic's own unresolved-callee sub-case is suppressed when verifyCallInvariantsSeam (07-03/07-05's pre-existing seam) is engaged, keeping check's independent callgraph-side re-derivation failing TOGETHER with verifyCallInvariants' own suppression under the same seeded mutation -- otherwise TestVerifyCallInvariantsSeamRestoresBothRefusals (a pre-existing test) would regress, since callgraph.Order now ALSO independently re-derives the same fact D-07-42 requires it to re-derive for defense in depth against a forged core.Program."
  - "Witness selection is proven with a discovery-order-driven fixture (a shared root fanning out, in sorted-adjacency order, to a self-cycle discovered FIRST and a smaller lexicographic 2-cycle discovered SECOND), not mere declaration-order permutation of core.Program.Functions -- because Order always computes its own sorted roots internally, so permuting the Functions slice alone never actually varies discovery order for two disjoint self-contained cycles (every function is independently a root, so a self-contained cycle's own smallest member is always visited as a root before any other member gets a chance to be reached first). This was discovered empirically when the first attempt at the mutation-kill test failed to distinguish the seam at all."
  - "The mutation matrix's self_edge and restored_after_every_override subtests exercise a synthetic core.Program mirroring cycle_self.lang's shape, not the fixture parsed through check.Program -- check.Program itself now refuses a real cycle before returning a core.Program, so there is no other way to hand callgraph.Order a checked cycle_self.lang program to run the seam against."

patterns-established:
  - "Fault-injection seams on a PRODUCTION dependency (unlike pathoracle, which is test/verify-only): unexported bool vars, false by default, exercised only through a companion _export_test.go's setter functions that return a restore closure, mirroring session_phase7_export_test.go's established shape."

requirements-completed: [SEM-07, QLT-08]

coverage:
  - id: D1
    description: "callgraph.Order: iterative white/gray/black DFS over an explicit stack, roots = all declared functions, edges from CalleeID only, refuses a length-2 cycle end to end via check before check.Program returns"
    requirement: "SEM-07"
    verification:
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestOrderDetectsMutualCycle"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallGraphCycleClearsReturnedProgram"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase07/cycle_mutual.lang (core.call_graph_cycle, exit 2, terminates)"
        status: pass
    human_judgment: false
  - id: D2
    description: "Deterministic cycle identity: canonical rotation plus cross-cycle lexicographically-smallest witness selection, each independently proven load-bearing by its own seeded mutation"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestCycleIdentityStableAcrossDeclarationOrder"
        status: pass
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestTwoDistinctCyclesSelectsLexicographicallySmallestWitness"
        status: pass
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestRotationRemovalMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestWitnessSelectionRemovalMutationKilled"
        status: pass
    human_judgment: false
  - id: D3
    description: "The 32-cause bound with stable truncation code applies only to the diagnostic; the traversal itself stays unbounded on a 200-node acyclic diamond-laden graph"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallGraphCycleBoundedAt32Causes"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallGraphCycleTruncatesAt33Members"
        status: pass
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestOrderNeverBoundsAnAcyclicTraversal"
        status: pass
    human_judgment: false
  - id: D4
    description: "D-07-35 span projection: causes and Primary carry real, non-zero spans projected from operation IDs; core.LinearOperation has no Span field"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallGraphCycleProjectsSpansFromOperationIDs"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLinearOperationHasNoSpanField"
        status: pass
    human_judgment: false
  - id: D5
    description: "The diamond/shared-leaf corpus and the two headline mutation kills (gray-versus-visited, self-edge), plus the unresolved-edge-drop control, landed in this plan alongside the control they gate (D-07-41)"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestDeepDiamondAcyclicHasFourDiamondsAndDepthEight"
        status: pass
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_test.go#TestCallGraphMutationMatrix"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase07/deep_diamond_acyclic.lang (status pass, bounded)"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase07/cycle_self.lang (core.call_graph_cycle, cycle_length \"1\")"
        status: pass
    human_judgment: false

duration: 95min
completed: 2026-09-08
status: complete
---

# Phase 07 Plan 06: Call-Graph Acyclicity Refusal Summary

**A new `internal/compiler/callgraph` package runs an iterative three-color DFS with no native recursion, refuses any call-graph cycle by name (`core.call_graph_cycle`) with a deterministic ID regardless of discovery order, and proves it can't be defeated by three specific seeded mutations, including the one a chain-only corpus could never catch.**

## Performance

- **Duration:** ~95 min
- **Tasks:** 3 (checkpoint decision + 2 code tasks; the plan's own numbering treats the ratified checkpoint as Task 0)
- **Files created:** 6
- **Files modified:** 6

## Accomplishments

- `callgraph.Order` refuses a call-graph cycle end to end, wired into `check.Program` immediately before it returns a `core.Program`, with the cyclic program cleared to its zero value rather than returned (D-07-14).
- Cycle identity is deterministic end to end: canonical rotation (D-07-16) puts the lexicographically smallest function ID at index 0, and cross-cycle witness selection (D-07-43) picks the smallest-rotation candidate among every cycle the whole traversal discovers -- proven with a fixture specifically engineered so discovery order and lexicographic order disagree.
- The diagnostic is bounded (32 `cycle_member` causes, then `truncated:core.call_cycle_bound`); the traversal is not -- verified against a 200-node synthetic diamond graph and the real `deep_diamond_acyclic.lang` fixture (13 functions, 4 chained diamonds, depth 8).
- Every cause and the diagnostic's own `Primary` carry real spans, projected from `OpCall` operation IDs through a new emission-time bookkeeping map (`ownershipSupport.CallSpans`) -- `core.LinearOperation` gained no `Span` field.
- Three fault-injection seams (gray-vs-visited, self-edge, unresolved-edge-drop) each independently proven load-bearing by a dedicated mutation-kill subtest, landing alongside the controls they gate per D-07-41.

## Task Commits

1. **Task 1: callgraph package + wiring + cycle_mutual.lang** - `c7efb65` (feat)
2. **Task 2: deterministic identity, 32-cause bound, span projection** - `e044642` (feat)
3. **Task 3: diamond corpus, cycle_self.lang, mutation matrix, session wiring** - `7a1b5a8` (test)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/callgraph/callgraph.go` - the package: `Order`, `cycleError`, `unresolvedCalleeError`, three fault-injection seam pairs, `MaxCycleCauses`, `TruncatedCycleBound`
- `internal/compiler/callgraph/callgraph_test.go` - unit tests, mutation matrix, diamond/depth assertion
- `internal/compiler/callgraph/callgraph_export_test.go` - seam setters for the external test package
- `internal/compiler/check/check.go` - `checkCallGraphAcyclic` gate, `CallSpans` threading through `ownershipSupport`/`checkLinear`/`checkBranch`/`Program`
- `internal/compiler/check/check_test.go` - 32/33-bound tests, span-projection test, no-Span-field reflection test
- `internal/compiler/core/core.go` - `core.CallGraphCycle` constant
- `internal/compiler/protocol/protocol.go` - documents why the truncation code moved to `callgraph` (import-cycle avoidance)
- `internal/compiler/session/session_phase7.go` - five new controls, `deep_diamond_acyclic.lang` in `phase07DispatchFixtures`
- `scripts/verify-phase7.sh` - mirrors the five new controls (deviation, see below)
- `testdata/phase07/cycle_mutual.lang`, `cycle_self.lang`, `deep_diamond_acyclic.lang` - new fixtures

## Decisions Made

See `key-decisions` in frontmatter for full rationale. Headline: the plan's stated location for `truncated:core.call_cycle_bound` (protocol.go) turned out to close a real import cycle once actually attempted (`check -> protocol -> interp -> [interp's own test package] -> check`), so the constant was placed in `callgraph` instead, alongside its sibling bound `MaxCycleCauses`, with the reasoning documented in `protocol.go` itself.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `truncated:core.call_cycle_bound` moved from `protocol` to `callgraph`**
- **Found during:** Task 2, first build attempt after adding the constant to `protocol.go` as the plan's prose specified
- **Issue:** `check` importing `protocol` closes a real import cycle: `protocol` imports `interp`, and `internal/compiler/interp`'s own (non-external) test package imports `check`, which would import `protocol` -- `go vet`/`go test` failed to build with "import cycle not allowed in test"
- **Fix:** Declared `callgraph.TruncatedCycleBound` instead (check already imports callgraph for every other purpose this plan needs), and left a doc comment in `protocol.go` explaining the divergence rather than silently placing it there with no trace
- **Files modified:** internal/compiler/callgraph/callgraph.go, internal/compiler/protocol/protocol.go, internal/compiler/check/check.go
- **Verification:** `go build ./...` and `go vet ./...` clean; `go test ./internal/compiler/check/... ./internal/compiler/callgraph/... ./internal/compiler/protocol/...` green
- **Committed in:** e044642 (Task 2 commit)

**2. [Rule 3 - Blocking] `scripts/verify-phase7.sh` updated to list the five new controls**
- **Found during:** Task 3, after adding the five controls to `session.Phase7RequiredControls()`
- **Issue:** `scripts/verify-phase7.sh` is not in this plan's `files_modified`, but `TestPhase7RequiredControlsMatchScript` asserts the script's own required-control grep block is set-equal to `Phase7RequiredControls()` in both directions -- leaving the script unchanged would have made that pre-existing test fail
- **Fix:** Added the same five `control:callgraph.*` identifiers to the script's existing `for control in ...` block
- **Files modified:** scripts/verify-phase7.sh
- **Verification:** `go test ./internal/compiler/session/... -run TestPhase7RequiredControlsMatchScript` passes; `sh scripts/verify-phase7.sh` reaches its own control-presence grep block without failing on this check
- **Committed in:** 7a1b5a8 (Task 3 commit)

**3. [Rule 1 - Bug] `checkCallGraphAcyclic`'s unresolved-callee case suppressed under `verifyCallInvariantsSeam`**
- **Found during:** Task 1, first full `go test ./internal/compiler/check/...` run after wiring `callgraph.Order` into `check.Program`
- **Issue:** `TestVerifyCallInvariantsSeamRestoresBothRefusals` (pre-existing, from 07-03/07-05) expects the seam to suppress ALL resolves-to-a-declared-function refusals; `callgraph.Order`'s own independent unresolved-callee re-derivation (needed for defense in depth against a forged `core.Program`, per the threat register) was refusing regardless of the seam, regressing that test
- **Fix:** `checkCallGraphAcyclic` now returns `nil` (no diagnostic) for its own unresolved-callee sub-case specifically when `verifyCallInvariantsSeam` is engaged, keeping every admission arm the seam names failing together under one seeded mutation
- **Files modified:** internal/compiler/check/check.go
- **Verification:** `TestVerifyCallInvariantsSeamRestoresBothRefusals` passes; `TestOrderRefusesUnresolvedCallee` (callgraph's own, seam-independent) still passes since it calls `callgraph.Order` directly, never through `check`
- **Committed in:** c7efb65 (Task 1 commit)

---

**Total deviations:** 3 auto-fixed (2 blocking, 1 bug)
**Impact on plan:** All three were necessary for correctness (avoiding a real import cycle, keeping a pre-existing seam test green, keeping the script/session control registries in sync). No scope creep -- no additional fixtures, controls, or refusal codes were added beyond what the plan specified.

## Issues Encountered

None beyond the deviations above. The `-race`/parallel-load timing flakes in `cache`, `measure`, and `cgen` (probe-timeout tests, documented as pre-existing in this plan's own prior-wave context) surfaced twice during `sh scripts/verify-phase7.sh` and one earlier full-suite run; both flaked tests (`TestCacheKeyCoversEveryDeclaredInput`, `TestPhase5ByPointerLoweringThreeEngineAgreement`) were re-run in isolation and passed, confirming they are not regressions from this plan's changes.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `internal/compiler/callgraph` and its five controls are ready for `07-07` to consume: `corevalidate`'s own independent traversal re-derives the same refusal (`core.call_graph_cycle` as a shared inert string constant, per this plan's own doc comments), exercises the `isDeclaredFunctionName` shadowing-precedence fixture this plan deliberately left untouched, and adds the `cycle_indirect.lang` (length >= 3) fixture this plan's naming convention reserved.
- D-07-38's ordering requirement (`07-06` and `07-07` land before `07-08`'s closure-digest chaining) is satisfied: acyclicity is proven here, and `07-07` closes the second peer before the digest chain -- which requires a DAG -- is built.
- No blockers. `go test ./...`, `go vet ./...`, and `sh scripts/verify-phase7.sh` are green (modulo the two documented pre-existing timing flakes, neither of which touches this plan's files).

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-08*

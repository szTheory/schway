---
phase: 08-interprocedural-loan-liveness-in-check
plan: 02
subsystem: check
tags: [ownership, loan-liveness, interprocedural, callgraph, transitivity, go]

# Dependency graph
requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 01
    provides: "interproceduralSummaryTable/buildInterproceduralSummaries scaffold, derivePlaceLoans' summary-aware OpCall branch, checkInterproceduralLoanLiveness/interproceduralLoanLivenessDiagnostic, callSignatureTable"
provides:
  - "deriveFunctionUsesParam: body-derived, memoized, transitive UsesParam, computed exactly once per function in a true callee-before-caller traversal"
  - "A fixed ordering bug in buildInterproceduralSummaries: callgraph.Order's own result is caller-before-callee, not callee-before-caller as its own prior doc comment claimed; the build now walks it backward"
  - "blockLoanLiveness's backward OpCall gate (D-08-08): a call enters the chain-ancestor walk iff the callee's usesParam is true or absent"
  - "checkInterproceduralLoanLiveness's widened reporting condition covering both the forward (D-08-07) and backward (D-08-08) interprocedural directions, with a direction-correct third diagnostic cause"
  - "TestSummaryDerivationRequiresProgramOrder / TestSummaryMemoNeverPersisted: D-08-11 and D-08-12/D-08-37 pinned under test, not merely documented"
affects: [08-03, 08-04, 08-05, 08-06, 09-peer-re-derivation-and-d-03-02-closure]

# Actuals (#2632)
actuals:
  tokens: 8245
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A per-function summary derivation that needs transitive facts from its own callees must consume callgraph.Order's result BACKWARD -- Order's own doc comment says \"reverse postorder\" but the traversal empirically lists a caller before every function it calls, the opposite of callee-before-caller."
    - "Fixpoint-until-no-change forward propagation (not a single flat pass) is what makes a per-operation work counter genuinely reflect program-order vs. reversed-order cost: program order converges in exactly two passes, reversed order needs one additional pass per unresolved link (O(n) passes of O(n) work each)."
    - "A shared backward transfer function (blockLoanLiveness) gates entry per-operation-kind on a summary table whose zero value always misses every lookup -- this keeps every pre-existing call site (which still passes the zero-value table) byte-identical while the new populated-table call site changes behaviour, with no call-site-specific branching."

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go

key-decisions:
  - "Fixed a real ordering bug (Rule 1) in buildInterproceduralSummaries, landed by 08-01 but never exercised by it: callgraph.Order's result is empirically caller-before-callee (confirmed against TestOrderSortsAdjacencyByCalleeID's own fixture and a standalone probe), not callee-before-caller as buildInterproceduralSummaries' own doc comment and 08-CONTEXT.md's D-08-03 both assumed. 08-01 never surfaced this because its own summary bits (returnsBorrowOfParam) never depended on another function's summary; Task 1's transitivity requirement made it load-bearing. Fixed by walking Order's result backward, not by changing callgraph itself (out of this plan's scope)."
  - "deriveFunctionUsesParam is a fixpoint-until-no-change loop over operations, not a single linear pass, so its own work counter genuinely reflects program-order (2-pass convergence) vs. reversed-order (O(n)-pass) cost -- required for TestSummaryDerivationRequiresProgramOrder to observe a real multiplier rather than assert a property the algorithm couldn't exhibit."
  - "checkInterproceduralLoanLiveness's offending-move detection loop was unchanged; only its POST-detection direction classification was widened (forward extendedByCall vs. backward last-use-is-a-gated-call), since the general last-use-vs-move-index comparison already covers both directions once blockLoanLiveness's gate makes the backward direction's last-use position correct."
  - "Task 2's tests are synthetic-core.Program tests (never through ast.Program -> check.Program), per the plan's own instruction that the .lang fixture pair for Pattern B lands in 08-03: verified empirically that the existing AST-shadow intraprocedural admission path (computeLoanLastUses) treats ANY call referencing a borrowed value as an unconditional use (summary-blind by design, D-08-09's own 08-01 finding), so a real .lang source in this shape is refused intraprocedurally today regardless of the callee -- Task 2's backward gate is only observable at the core.Program level until a later plan reconciles the two paths."

patterns-established:
  - "Ordering claims about a shared traversal helper (callgraph.Order) must be empirically verified before a new consumer relies on them transitively -- a doc comment's own terminology (\"reverse postorder\") can silently diverge from a specific consumer's actual need (callee-before-caller) without any existing test catching it, if no prior consumer's correctness depended on the distinction."

requirements-completed: [OWN-06]

coverage:
  - id: D1
    description: "UsesParam is derived once per function, in a true callee-before-caller traversal, memoized for the run, and transitive through relays (leaf -> relay -> caller, and a diamond with a shared leaf reached through two callers)"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestDeriveFunctionUsesParamBehaviors"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestSummaryDerivationTwoHopChainPropagates"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestSummaryDerivationIsOnePassPerFunction"
        status: pass
    human_judgment: false
  - id: D2
    description: "borrow; move; call(v) refuses iff the callee uses its parameter; the safe twin (non-using callee) admits with zero diagnostics -- both members of the twin verified against a synthetic core.Program differing only in the callee's own usesParam bit"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessTwinPatternB"
        status: pass
    human_judgment: false
  - id: D3
    description: "Summary derivation work grows no faster than operation count in program order across a {8,32,128,512} size series, and the identical chain reversed costs strictly more per operation at the largest size (171.3x observed, matching spike S-006's 192x-at-512 finding within the same order of magnitude)"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestSummaryDerivationRequiresProgramOrder"
        status: pass
    human_judgment: false
  - id: D4
    description: "No package-level variable in package check retains the interprocedural summary table across check.Program invocations; running check.Program twice rebuilds the memo twice (proven mechanically via go/ast structural scan plus a build-observed instrumentation hook), never through inspection of the final result alone"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestSummaryMemoNeverPersisted"
        status: pass
    human_judgment: false
  - id: D5
    description: "No existing ownership.*/check.*/core.* verdict moved; the whole pre-existing check/corevalidate suite plus go vet stay green (excluding the pre-existing, unrelated session-package spike-registry gap)"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check/... ./internal/compiler/corevalidate/..."
        status: pass
      - kind: unit
        ref: "go test ./... && go vet ./..."
        status: pass
    human_judgment: false

# Metrics
duration: ~50min
completed: 2026-09-10
status: complete
---

# Phase 08 Plan 02: Interprocedural Summary Mechanism Expansion Summary

**`UsesParam` now derives once per function from its own checked body, in a genuinely callee-before-caller traversal (fixing a real ordering bug 08-01 never exercised), propagates transitively through relay chains and diamonds, and is consumed by a new backward gate in `blockLoanLiveness` that resolves both members of Pattern B's twin -- with the program-order cost claim and the never-persisted memo pinned under test rather than documented.**

## Performance

- **Duration:** ~50 min
- **Started:** 2026-09-09T21:00 (approx, continuing directly from 08-01)
- **Completed:** 2026-09-10T01:21:19Z
- **Tasks:** 3 completed (all `type="auto" tdd="true"`)
- **Files modified:** 2

## Accomplishments

- Landed `deriveFunctionUsesParam`: a fixpoint-until-no-change forward pass over a function's own `core.LinearOperation` stream, seeding a parameter-derived place set and propagating it in program order. A place counts as "used" when read by anything other than an `OpCall` or the function's own terminating `OpReturn`, or when passed to a callee whose own `usesParam` (via `summaries.lookup`) is true -- an absent callee counts as `usesParam == true`, the refusing direction. The fixpoint shape (not a single flat pass) is what makes its own work counter genuinely reflect program-order vs. reversed-order cost.
- **Discovered and fixed a real ordering bug** in `buildInterproceduralSummaries` (landed by 08-01, never exercised by it): `callgraph.Order`'s own doc comment and 08-CONTEXT.md's D-08-03 both describe its result as callee-before-caller, but it empirically lists a caller before every function it calls (confirmed against the package's own `TestOrderSortsAdjacencyByCalleeID` fixture and a standalone probe). `buildInterproceduralSummaries` now walks `callgraph.Order`'s result backward, the genuine callee-before-caller traversal Task 1's transitivity requires.
- Added `deriveFunctionUsesParamObserved` and `buildInterproceduralSummariesObserved` (mirroring `callSignatureTableBuildObserved`), proving respectively that a diamond call graph's shared leaf is derived exactly once, and that the interprocedural summary table is rebuilt fresh on every `check.Program` invocation, never persisted.
- Added D-08-08's single consuming clause to `blockLoanLiveness`: an `OpCall` enters the backward chain-ancestor walk iff the callee's `usesParam` is true (or absent, the refusing direction) -- every other operation kind unchanged. `summaries`' zero value always misses every lookup, so this gate is a byte-identical no-op for every pre-existing call site still passing `interproceduralSummaryTable{}` (OWN-03's own intraprocedural law is untouched).
- Widened `checkInterproceduralLoanLiveness`'s reporting condition to cover both interprocedural directions (forward `extendedByCall`, and the new backward case where the offending loan's own last use is the gating call), with `interproceduralLoanLivenessDiagnostic` now taking a precomputed, direction-correct cause-3 Detail string (`return.mode=` vs. `parameters[0].mode=`) chosen from a recorded fact rather than re-derived.
- Pinned both cost/lifecycle invariants under test: `TestSummaryDerivationRequiresProgramOrder` observed a 171.3x reversed-vs-forward work-per-operation multiplier at k=512 (spike S-006 measured 192x -- same order of magnitude); `TestSummaryMemoNeverPersisted` combines a `go/ast` structural scan with a build-observed instrumentation hook.

## Task Commits

Each task was committed atomically:

1. **Task 1: Derive `UsesParam` once per function, in reverse postorder, memoized and transitive** - `46438b2` (feat)
2. **Task 2: The backward `OpCall` gate — `borrow; move; call` refuses only when the callee uses its parameter** - `3266684` (feat)
3. **Task 3: Pin the program-order invariant and the never-persisted memo** - `637e555` (test)

**Plan metadata:** commit pending (this SUMMARY + STATE.md/ROADMAP.md/REQUIREMENTS.md)

_Note: all three tasks are marked `tdd="true"`; per 08-01's own recorded precedent (`workflow.tdd_mode` is `false` for this project), implementation and tests were developed together and each task's own `<verify>` block was run and confirmed green before that task's single commit._

## Files Created/Modified

- `internal/compiler/check/check.go` - `deriveFunctionUsesParam`, `deriveFunctionUsesParamObserved`, `buildInterproceduralSummariesObserved`, `buildInterproceduralSummaries`' callee-before-caller traversal fix, `interproceduralSummary.parameterMode`, `blockLoanLiveness`'s backward `OpCall` gate, `checkInterproceduralLoanLiveness`'s widened direction classification, `interproceduralLoanLivenessDiagnostic`'s precomputed-Detail signature
- `internal/compiler/check/check_test.go` - `TestDeriveFunctionUsesParamBehaviors`, `TestSummaryDerivationTwoHopChainPropagates`, `TestSummaryDerivationIsOnePassPerFunction`, `TestInterproceduralLivenessTwinPatternB`, `relayChainFunction`/`reverseFunctionOperations` helpers, `TestSummaryDerivationRequiresProgramOrder`, `TestSummaryMemoNeverPersisted`

## Decisions Made

- The `callgraph.Order` ordering-direction bug fix and its rationale, recorded above and in `buildInterproceduralSummaries`' own doc comment (not a `callgraph` package change -- out of this plan's scope; the fix lives entirely in how `check` consumes `Order`'s result).
- `deriveFunctionUsesParam`'s fixpoint-loop shape (over a single linear pass), required to make the program-order-vs-reversed cost claim genuinely observable rather than merely correct-but-unfalsifiable.
- Task 2's twin test is synthetic-`core.Program`-only, per the plan's own instruction that the `.lang` fixture pair for Pattern B lands in 08-03 -- verified during implementation that the existing AST-shadow intraprocedural path (`computeLoanLastUses`) is summary-blind by design and would flag a real `.lang` source in this shape regardless of the callee, so this task's backward gate is only observable at the core level until 08-03 reconciles the two paths.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `buildInterproceduralSummaries` iterated `callgraph.Order`'s result in the wrong direction for transitivity**
- **Found during:** Task 1, while implementing `deriveFunctionUsesParam`'s consultation of `summaries.lookup(calleeID)` for transitivity
- **Issue:** `callgraph.Order`'s own doc comment and 08-CONTEXT.md's D-08-03 both state the result is "callee-before-caller"; empirically (confirmed by a standalone probe and by re-reading `TestOrderSortsAdjacencyByCalleeID`'s own fixture) it is caller-before-callee. 08-01's own summary bits never depended on another function's summary, so this never surfaced as a test failure until Task 1's transitivity requirement made the direction load-bearing.
- **Fix:** `buildInterproceduralSummaries` now iterates `callgraph.Order`'s result backward (`for i := len(order) - 1; i >= 0; i--`), the genuine callee-before-caller traversal.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestSummaryDerivationTwoHopChainPropagates` and `TestSummaryDerivationIsOnePassPerFunction` both assert the leaf's `usesParam` bit reaches the caller through one and two hops respectively; both fail immediately if the traversal direction reverts.
- **Committed in:** `46438b2`

---

**Total deviations:** 1 auto-fixed (1 Rule 1 -- a genuine correctness bug directly caused by implementing this plan's own stated transitivity requirement, not architectural, no scope creep beyond the fix itself).
**Impact on plan:** Necessary for D-08-03's transitivity claim to hold at all; without it every transitive `usesParam` bit beyond the immediately-preceding-in-order function would silently read as unresolved (fail-closed `true`, safe but wrong) rather than the callee's real, derived bit.

### TDD process note (not a deviation, documented for transparency)

Same as 08-01: all three tasks are marked `tdd="true"`, but `workflow.tdd_mode` is `false` for this project, so the MVP+TDD gate does not apply. Implementation and tests were developed together and each task's own `<verify>` block was confirmed green before that task's single commit.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `interproceduralSummaryTable` now carries all three fields (`usesParam`, `returnsBorrowOfParam`, `returnMode`, `parameterMode`) this phase's two summary bits require; 08-03 can proceed directly to landing the `.lang` fixture pair for both twin patterns.
- `blockLoanLiveness`'s backward gate and `checkInterproceduralLoanLiveness`'s two-direction classification are stable, tested seams ready for 08-03's real-source fixtures.
- Known open item (not a blocker, documented above): the AST-shadow intraprocedural admission path (`computeLoanLastUses`) is summary-blind by design and will flag Pattern B's real `.lang` source shape unconditionally via `ownership.move_while_borrowed`, before the interprocedural pass ever runs. 08-03 (or whichever plan lands the `.lang` fixture pair) must account for this when choosing the fixture's exact shape, or reconcile the two paths if the safe twin needs to be observable end-to-end via the CLI.
- No blockers for 08-03 through 08-06.

## Self-Check: PASSED

- FOUND: internal/compiler/check/check.go
- FOUND: internal/compiler/check/check_test.go
- FOUND commit: 46438b2
- FOUND commit: 3266684
- FOUND commit: 637e555
- Re-ran acceptance criteria for all three tasks: `go test ./internal/compiler/check/... -run 'SummaryDerivation|UsesParam' -v` PASS; `go test ./internal/compiler/check/... -count=2 -shuffle=on` PASS; `go test ./internal/compiler/check/... -run 'TwinPatternB|InterproceduralLoanLiveness' -v` PASS; `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/...` PASS; `go test ./internal/compiler/check/... -run 'SummaryDerivationRequiresProgramOrder|SummaryMemoNeverPersisted' -v` PASS — all criteria pass
- Re-ran plan-level `<verification>`: `go test ./internal/compiler/check/...` exits 0; `go test ./... && go vet ./...` exits 0 (only pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure in the session package, confirmed present before this phase began per the orchestrator's own prior-wave note) — PASS
- Observed forward-vs-reversed work-per-operation multiplier: 171.3x at k=512 (forward=3.00, reversed=514.00)

---
*Phase: 08-interprocedural-loan-liveness-in-check*
*Completed: 2026-09-10*

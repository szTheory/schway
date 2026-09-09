---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 08
subsystem: semantic-core
tags: [closure-digest, merkle-chain, call-graph, corevalidate, originvalidate, mutation-testing]

requires:
  - phase: 07-calls-signatures-and-call-graph-refusal
    provides: "07-06's callgraph.Order (proven-acyclic reverse postorder) and 07-07's corevalidate cycle peer / phase-wide completeness matrix"
provides:
  - "ClosureDigest as a real Merkle chain over callee SUMMARY digests, computed only over a call graph already proven acyclic"
  - "The callee-changes-invalidates-caller regression that is the whole reason ClosureDigest exists, with a mutation-kill proof on both the producer and the independent peer"
  - "corevalidate's own independent chained-digest re-derivation, matching originvalidate's byte-for-byte wherever Callable is already known to agree"
affects: [phase-08-interprocedural-loan-liveness, phase-09-peer-rederivation-and-d-03-02-closure]

actuals:
  tokens: 16800
  tasks: 2
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Chain-over-summary-digests: originvalidate.BuildInterface runs callgraph.Order first and refuses to compute any digest on a non-DAG; corevalidate's own checkCallGraphAcyclic doubles as the source of its own independent postorder for the peer's chain, so the acyclicity proof and the digest ordering come from the SAME traversal on each side"
    - "Byte-identical-preimage-by-duplicated-literal: corevalidate never imports originvalidate, so its peer digest builder duplicates the domain separator and pair-struct shape verbatim (matching core.CallGraphCycle's own precedent for an inert string shared across independent derivations) rather than sharing code"

key-files:
  created:
    - internal/compiler/originvalidate/originvalidate_closure_chain_test.go
    - internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/export_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase7_test.go
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - scripts/verify-phase7.sh

key-decisions:
  - "callgraph.Order's returned array is CALLERS-first, callees-last (its own doc comment and TestOrderSortsAdjacencyByCalleeID's worked example both confirm this) -- so chaining bottom-up (callee digest available before its caller reads it) requires walking that array BACKWARD, not forward. Documented explicitly in both BuildInterface and corevalidate's mirror to prevent a future reader from iterating it the wrong direction."
  - "corevalidate's checkCallGraphAcyclic now also records its own adjacency (deduped, matching callgraph.buildAdjacency's dedup discipline) and its own raw DFS postorder (callee-before-caller, no reversal needed) as a byproduct of the SAME traversal that already proves acyclicity -- no second graph pass for the peer's chain."
  - "The peer's ClosureDigest is only asserted equal to the producer's for functions where Callable is already known to agree (D-07-33's narrowed scope, reusing TestSummaryPeerCallableAgreesOnOriginOmittedClass's exact scoping) -- ClosureDigest hashes the WHOLE FunctionSignature including Callable, so a function in D-07-33's documented Callable-divergence classes would otherwise fail a byte-for-byte comparison for a reason this plan does not own."

requirements-completed: [SEM-05, QLT-08]

coverage:
  - id: D1
    description: "ClosureDigest is a real Merkle chain over callee summary digests, computed only over a call graph already proven acyclic (D-07-38), with leaves unperturbed and no edge list published in the /1 summary"
    requirement: SEM-05
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestComputationOrderIsCalleeBeforeCaller"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestZeroCalleeLeafUnperturbedByContext"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestSharedLeafConsistentAcrossParents"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestBuildInterfaceRefusesCyclicGraphComputingNoDigest"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestDeterministicAcrossCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestSummaryCarriesNoEdgeList"
        status: pass
    human_judgment: false
  - id: D2
    description: "Changing a callee changes its caller's ClosureDigest (never an unrelated function's); the property is re-derived by an independent peer; both mutation-kill seams (empty-callee-pairs, discovery-order) are proven load-bearing on both the producer and the peer"
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestCalleeChangeInvalidatesCallerClosureDigest"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestUnrelatedFunctionChangeDoesNotInvalidateCallerClosureDigest"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestEmptyCalleesMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_closure_chain_test.go#TestClosureDigestDiscoveryOrderMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go#TestPeerClosureDigestEmptyCalleesMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go#TestPeerClosureDigestDiscoveryOrderMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestSummaryPeerClosureDigestMatchesProducerAcrossCorpus"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7ControlsAreMutationKilled"
        status: pass
    human_judgment: false

duration: 64min
completed: 2026-09-09
status: complete
---

# Phase 7 Plan 08: ClosureDigest Chaining and the Callee-Changes-Invalidates-Caller Regression Summary

**`ClosureDigest` is a real Merkle chain over callee summary digests, computed in `callgraph.Order`'s reverse postorder (walked backward for bottom-up processing) only after `callgraph.Order` has proven the call graph acyclic, with an independent `corevalidate` re-derivation over its own traversal — closing SEM-05 for Phase 07.**

## Performance

- **Duration:** 64 min
- **Started:** 2026-09-09T03:15:21Z
- **Completed:** 2026-09-09T04:19:36Z
- **Tasks:** 2 completed
- **Files modified:** 11 (2 new test files, 9 modified)

## Accomplishments

- `originvalidate.BuildInterface` runs `callgraph.Order` before computing any `ClosureDigest`; a refused (cyclic or unresolved-callee) graph returns that error unchanged with zero digests computed.
- The chaining arm supplies real, deduped, sorted `(ID, ClosureDigest)` callee pairs to `07-01`'s unchanged canonical preimage function, processed in callee-before-caller order (the reverse of `callgraph.Order`'s own caller-first return array).
- `corevalidate` grows its own independent chained-digest re-derivation: `checkCallGraphAcyclic`'s existing traversal now also records its deduped adjacency and raw postorder as a free byproduct, feeding a duplicated (never shared) preimage builder that matches the producer's digest byte-for-byte wherever `Callable` is already known to agree.
- The callee-changes-invalidates-caller regression — the entire reason `ClosureDigest` exists — is landed and proven on both the producer and the peer, with two independent fault-injection seams (empty-callee-pairs, discovery-order) each mutation-killed on both sides.
- Fixed a regression this plan's own change would otherwise introduce at `check`'s admission path: `buildCallSignatureTable` now propagates `BuildInterface`'s new cycle-refusal error instead of silently falling back to an empty table (which would mask `check`'s own `core.call_graph_cycle` diagnostic behind a spurious `core.callee_not_callable`).

## Task Commits

Each task was committed atomically:

1. **Task 1: The chaining arm — callee digests in reverse postorder over a proven DAG** - `97b7855` (feat)
2. **Task 2: Callee-changes-invalidates-caller, re-derived by the peer, with its mutation kill** - `5f33da3` (feat)

## Files Created/Modified

- `internal/compiler/originvalidate/originvalidate.go` - `BuildInterface` runs `callgraph.Order` first, chains `ClosureDigest` bottom-up, two fault seams + ordering-instrumentation hook
- `internal/compiler/originvalidate/originvalidate_closure_chain_test.go` - Task 1 + Task 2 producer-side tests (ordering, leaf invariance, shared-leaf propagation, cyclic refusal, determinism, no-edge-list, callee-change/unrelated-change, both mutation kills)
- `internal/compiler/originvalidate/export_test.go` - test-only setters for the two new seams and the ordering-observer hook
- `internal/compiler/corevalidate/corevalidate.go` - `checkCallGraphAcyclic` records its own deduped adjacency + raw postorder; `chainPeerClosureDigests` and its own duplicated preimage builder
- `internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go` - peer-side mutation kills for both fault seams
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go` - `TestSummaryPeerClosureDigestMatchesProducerAcrossCorpus` (Task 2 Test 3)
- `internal/compiler/corevalidate/export_test.go` - test-only setters for the peer-side seams
- `internal/compiler/session/session_phase7.go` - `control:summary.closure_digest_chained` / `control:summary.closure_digest_reverse_postorder` added to `Phase7RequiredControls()`
- `internal/compiler/session/session_phase7_test.go` - the two new controls added to the completeness matrix
- `internal/compiler/core/core.go` - `ClosureDigest`'s doc comment updated: chaining widens staleness detection, adds no authenticity
- `internal/compiler/check/check.go` / `check_test.go` - `buildCallSignatureTable` now returns `(table, error)`; the call site skips `verifyCallableRefusal` on error rather than admitting the empty-table fallback (Rule 3 fix, see Deviations)
- `scripts/verify-phase7.sh` - the two new controls added to the phase07 required-control grep list, kept in lockstep with `Phase7RequiredControls()`

## Decisions Made

- **`callgraph.Order`'s return direction.** Its own doc comment and `TestOrderSortsAdjacencyByCalleeID` both establish that the returned "reverse postorder" array lists callers first and callees last — the standard reverse-postorder/topological-order convention (source before dependents). Chaining needs the *opposite* direction (a callee's digest must exist before its caller reads it), which is exactly that array walked backward — recovering the traversal's raw DFS postorder, where a node is appended only once every callee it can reach has already finished. Both `originvalidate.BuildInterface` and `corevalidate.chainPeerClosureDigests` document this explicitly, since the plan's own phrasing ("iterate the reverse postorder: every callee is finished before its caller") is only satisfied by the backward walk, not a literal forward iteration of the array — a genuine terminology trap worth flagging for future readers.
- **Peer-producer digest comparison scope.** `TestSummaryPeerClosureDigestMatchesProducerAcrossCorpus` reuses `TestSummaryPeerCallableAgreesOnOriginOmittedClass`'s exact narrowed scope (D-07-33) rather than asserting whole-corpus equality: `ClosureDigest` hashes the entire `FunctionSignature`, `Callable` included, so a function in D-07-33's already-documented, narrowed-away Callable-divergence classes would trivially fail a byte-for-byte digest comparison for a reason this plan does not introduce and does not own. This surfaced as a real, non-obvious test failure during implementation (see Deviations) before the scoping was corrected.
- **Two independent adjacency builds, one dedup discipline.** `corevalidate`'s `checkCallGraphAcyclic` now dedupes edges (a caller calling the same callee via two `OpCall`s contributes one edge) — a discipline it did not need for cycle detection alone, but does need so `chainPeerClosureDigests`'s callee-pair list matches `originvalidate`'s own deduped pairs exactly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `buildCallSignatureTable` masked the correct cycle diagnostic**
- **Found during:** Task 1, after adding `callgraph.Order` to `BuildInterface`
- **Issue:** `check.Program` calls `buildCallSignatureTable` (which calls `originvalidate.BuildInterface`) *before* its own dedicated `callgraph.Order`-based cycle gate, as part of the "call" admission arm's `Callable` consult. Once `BuildInterface` itself started refusing cyclic graphs, `buildCallSignatureTable`'s pre-existing empty-table fallback (`if err != nil { return callSignatureTable{} }`, previously reachable only via an impossible `json.Marshal` error) became reachable for every genuinely cyclic program. An empty table makes `verifyCallableRefusal` refuse the FIRST `core.OpCall` it finds with `core.callee_not_callable`, populating `result.Diagnostics` and causing the dedicated cycle gate (`len(result.Diagnostics) == 0` guarded) to be skipped entirely — replacing the correct `core.call_graph_cycle` diagnostic (with its causes and spans) with a wrong, generic one. Confirmed by running the full sequential suite mid-implementation: 13 `check` package tests failed with exactly this symptom.
- **Fix:** `buildCallSignatureTable` now returns `(callSignatureTable, error)`; the call site skips `verifyCallableRefusal` entirely when `BuildInterface` errored (reachable only for a cycle, since `verifyCallInvariants` already required every `CalleeID` to resolve to a declared function before this point), falling through with zero diagnostics so the dedicated gate below runs and refuses the cycle properly.
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_test.go`
- **Verification:** All 13 previously-failing `check` package cycle tests pass; full sequential `go test ./... -p 1` is green.
- **Committed in:** `97b7855` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking).
**Impact on plan:** Necessary to prevent this plan's own change from silently regressing an existing, tested diagnostic. No scope creep — the fix is scoped exactly to the reachable-error path `BuildInterface`'s new behavior introduced.

## Issues Encountered

- The plan's Task 1 acceptance criterion named `testdata/phase07/cycle_mutual.lang`'s checked program, "constructed in test scope, since `check` refuses it," as the cyclic-graph fixture for `TestBuildInterfaceRefusesCyclicGraphComputingNoDigest`. `check`'s own cycle-bypass seam (`disableCallGraphCycleRefusalForTest`) is unexported and same-package-only, unreachable from `originvalidate_test`. Used a hand-built, forged two-function mutual-cycle `core.Program` instead (mirroring `callgraph_test.go`'s and `corevalidate_mutation_matrix_test.go`'s own precedent for exercising these packages against synthetic programs no parser would produce) — the identical behavioral property is proven; only the specific fixture source differs from the plan's suggestion.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SEM-05 is closed for Phase 07: the callee signature summary is a digest-bound artifact whose `ClosureDigest` is a Merkle chain over callee summary digests, computed only over a call graph already proven acyclic, with no caller admission path reading a callee body.
- QLT-08 holds for both controls this plan introduces (`control:summary.closure_digest_chained`, `control:summary.closure_digest_reverse_postorder`), and `07-07`'s phase-wide completeness matrix (`TestPhase7ControlsAreMutationKilled`, `TestPhase7RequiredControlsMatchScript`) covers them.
- Named forward obligation, not built here: QLT-06 ("no interprocedural fact is marked cacheable until a callee-changes-invalidates-caller regression test gates it; interprocedural cache keys derive from the call-graph closure, not per-unit hashes") is a **later milestone requirement**. This plan builds the regression QLT-06 will require and the closure-derived key it will use, but does **not** mark any interprocedural fact cacheable and does **not** modify `cache.Input` — confirmed unmodified (`git status` shows no changes under `internal/compiler/cache/`).
- Phase 07 is now complete (8/8 plans). Ready for `/gsd-verify-work 07` and Phase 08 planning (interprocedural loan liveness in `check`), which S-006's cost-scaling probe gates.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-09*

## Self-Check: PASSED

- All created/modified files verified present on disk.
- Both task commits (`97b7855`, `5f33da3`) verified present in `git log`.
- Full sequential `go test ./... -p 1` green (exit 0, all packages `ok`).
- `go test -race` green for originvalidate, corevalidate, check, session, callgraph.
- `go vet ./...` clean.
- `sh scripts/verify-phase7.sh`'s only failures were the pre-existing, documented parallel-load flakes (cache/measure/cgen/native probe-timeout, lang-repair timeout) named in this plan's prior-wave context, plus one transient `verify testdata/phase1` timeout under that same parallel load — reproduced clean in isolation (`/tmp/lang-check --json verify testdata/phase1` exits 0).

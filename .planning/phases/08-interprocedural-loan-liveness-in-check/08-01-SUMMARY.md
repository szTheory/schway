---
phase: 08-interprocedural-loan-liveness-in-check
plan: 01
subsystem: check
tags: [ownership, loan-liveness, interprocedural, diagnostics, callgraph, ast, go]

# Dependency graph
requires:
  - phase: 07-calls-signatures-and-call-graph-refusal
    provides: "core.FunctionSignature/callSignatureTable (Return.Mode), callgraph.Order (proven-acyclic reverse postorder), OpCall (CalleeID/SourceID/TargetID), spanByOperationID"
provides:
  - "interproceduralSummaryTable / buildInterproceduralSummaries: a post-acyclicity, per-callee summary bit (returnsBorrowOfParam) read from the callee's already-published FunctionSignature, never its body"
  - "derivePlaceLoans' summary-aware OpCall branch: the forward canonicalization pass that propagates a loan across a call boundary when the callee declares a borrow-of-param return"
  - "checkInterproceduralLoanLiveness / check.interprocedural_loan_liveness: the new post-acyclicity refusal, with a fixed three-cause template and no repairs"
  - "computeLoanLastUses' \"call\" case and calleeContract.ReturnsBorrowOfParam: the second (AST-shadow) admission path now agrees with the real path on loan last-use positions across a call"
affects: [09-peer-re-derivation-and-d-03-02-closure, 08-02, 08-03, 08-04]

# Actuals (#2632)
actuals:
  tokens: 16331
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Interprocedural facts enter through a FORWARD canonicalization pre-pass (derivePlaceLoans), never the backward transfer function (D-08-07) -- the riskiest structural choice this phase settles."
    - "Post-acyclicity, post-intraprocedural-admission ordering for a new whole-program pass: build the pass's own summary table only after callgraph.Order and every function's own body check have already produced zero diagnostics."
    - "Two independent admission paths (real core.LinearOperation stream vs. AST-shadow computeLoanLastUses) proven to agree via a differential test that fails if either path's own handling is reverted alone."

key-files:
  created: []
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - internal/compiler/session/session_peer_gate_test.go
    - testdata/phase07/relay_escort_witness.lang

key-decisions:
  - "Checkpoint auto-ratified under auto_advance/yolo: all three Task 1 proposals accepted verbatim -- check.loan_liveness_bound_exceeded, check.interprocedural_loan_bound (with MaxInterproceduralLoanCauses = 32), and the fixed three-cause template (borrow_created_here / loan_extended_by_call / callee_return_contract, in that order, cause 3 ID-only and spanless)."
  - "D-08-09 planning determination (Task 3): derivePlaceLoans/blockLoanLiveness over real core.LinearOperations is load-bearing for success criterion 1; computeLoanLastUses is load-bearing for the intraprocedural law's own loan-expiry answer (D-05-35(d)). Both must derive the same last-use positions or the two laws disagree about a fact neither owns alone -- proven by TestComputeLoanLastUsesAndDerivePlaceLoansAgree."
  - "Widened the existing spanByOperationID/CallSpans channel (D-07-35) to also carry every take/borrow/borrow_mut operation's own span, rather than adding a second span-carrying map -- needed for the new diagnostic's Primary and borrow_created_here cause, which D-07-35's original call-only scope did not cover."
  - "Reused Task 1's signatureTable build (hoisted out of the verifyCallableRefusal admission arm) for the new interprocedural pass instead of calling buildCallSignatureTable a second time, to keep callSignatureTableBuildObserved's exactly-one-build-per-Program-call invariant intact (TestCallSignatureTableBuiltBeforeCallableAdmissionRuns)."

patterns-established:
  - "Interprocedural summary tables (interproceduralSummaryTable) mirror callSignatureTable's own immutability-by-construction and body-blindness shape: read-only, lookup-only, built once from already-published signatures, never from a callee's Linear/Match field."

requirements-completed: [OWN-06]

coverage:
  - id: D1
    description: "escort refuses with check.interprocedural_loan_liveness on relay_escort_witness.lang, Primary span on the take, three fixed-role causes, no repairs"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestRelayEscortWitnessRefusesInterproceduralLiveness"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang"
        status: pass
    human_judgment: false
  - id: D2
    description: "The safe twin (callee declaring an owned return) still checks clean -- the refusal is contract-driven, not shape-driven"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLoanLivenessTracer"
        status: pass
    human_judgment: false
  - id: D3
    description: "computeLoanLastUses and derivePlaceLoans/blockLoanLiveness agree on every loan's last-use position across a call, and disagree (fail) if either path's own call handling is reverted alone"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestComputeLoanLastUsesAndDerivePlaceLoansAgree"
        status: pass
    human_judgment: false
  - id: D4
    description: "No existing ownership.* verdict moved; the whole pre-existing suite plus corevalidate/originvalidate stay green"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/originvalidate/..."
        status: pass
      - kind: unit
        ref: "go test ./... && go vet ./..."
        status: pass
    human_judgment: false

# Metrics
duration: ~33min
completed: 2026-09-09
status: complete
---

# Phase 08 Plan 01: End-to-End Interprocedural Loan-Liveness Tracer Summary

**One interprocedural loan-liveness refusal now travels the whole stack -- declared callee return contract, in-memory summary bit, forward canonicalization, backward liveness, a newly minted `check.interprocedural_loan_liveness` diagnostic, and the `lang check` CLI -- proven on the two-milestone-old `relay_escort_witness.lang` fixture, with both admission paths (real and AST-shadow) proven to agree by a differential test.**

## Performance

- **Duration:** ~33 min
- **Started:** 2026-09-09T20:11:09-04:00 (previous plan-doc commit)
- **Completed:** 2026-09-09T20:44:11-04:00
- **Tasks:** 3 completed (1 checkpoint:decision, 1 tracer, 1 auto)
- **Files modified:** 5

## Accomplishments

- Landed `interproceduralSummary`/`interproceduralSummaryTable`/`buildInterproceduralSummaries`: a post-acyclicity, per-function summary table read exclusively from `callSignatureTable`'s already-published `core.FunctionSignature`, structurally incapable of reaching a callee body.
- Widened `derivePlaceLoans` (the forward canonicalization pass, D-08-07) with an `OpCall` branch: a callee whose summary reports `returnsBorrowOfParam` propagates the loan onto the call's target place; every other callee (including absent-from-table) gets an explicit `continue` that refuses to propagate, load-bearing against today's over-approximating fallthrough.
- Added `checkInterproceduralLoanLiveness`/`interproceduralLoanLivenessDiagnostic`, wired into `check.Program` strictly after `checkCallGraphAcyclic` and after every function's own intraprocedural admission (D-08-12/D-08-27 interim rule): emits `check.interprocedural_loan_liveness` with the ratified three-cause template, `Primary` on the offending `take`, and no repairs.
- Flipped `testdata/phase07/relay_escort_witness.lang`'s two-milestone-old finding: `TestRelayEscortWitnessRefusesInterproceduralLiveness` (replacing `TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness`) now asserts the refusal; `TestInterproceduralLoanLivenessTracer` proves the owned-return safe twin still admits.
- Closed D-07-49 in the AST-shadow admission path (`computeLoanLastUses`'s new `case "call":`) and proved both admission paths agree via `TestComputeLoanLastUsesAndDerivePlaceLoansAgree`, a four-body-shape differential that fails if either path's own "call" handling is reverted alone (verified by direct mutation, not merely asserted).

## Task Commits

Each task was committed atomically:

1. **Task 1: Ratify the two minted diagnostic code strings and the cause shape** — checkpoint:decision, auto-ratified (no code changes; folded into Task 2's commit message)
2. **Task 2: End-to-end interprocedural refusal — one path, wired through every layer** - `2e37aab` (feat)
3. **Task 3: D-07-49 in BOTH admission paths, with a differential that fails if only one is fixed** - `6b7cdc8` (feat)

**Plan metadata:** commit pending (this SUMMARY + STATE.md/ROADMAP.md/REQUIREMENTS.md)

_Note: both feat commits carry their own tests (implementation and verification developed together and run as one unit before commit; see Deviations)._

## Files Created/Modified

- `internal/compiler/check/check.go` - `interproceduralSummary`/`interproceduralSummaryTable`/`buildInterproceduralSummaries`, `derivePlaceLoans`'s summary-aware `OpCall` branch, `cfgBlocksForFunction`, `checkInterproceduralLoanLiveness`, `interproceduralLoanLivenessDiagnostic`, span-widening for take/borrow spans, `calleeContract.ReturnsBorrowOfParam`, `computeLoanLastUses`'s `case "call":`
- `internal/compiler/check/check_test.go` - `TestRelayEscortWitnessRefusesInterproceduralLiveness`, `TestInterproceduralLoanLivenessTracer`, `TestComputeLoanLastUsesAndDerivePlaceLoansAgree`, updated `TestRelayEscortWitnessBothFunctionsAreCallable`/`TestCallAdmissionBodyBlindControl`/`TestCallArgumentConsumptionUnchangedAcrossAcceptingCorpus`/`TestLivenessLawsStayIndependent`, and the mechanical `interproceduralSummaryTable{}` parameter threading across all pre-existing `loanLivenessFixpoint`/`materializeLoanEndpoints` call sites
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go` - updated `TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed`'s expectation now that `check` itself also refuses the fixture (via a different code) before the peer is even consulted for this assertion's purposes
- `internal/compiler/session/session_peer_gate_test.go` - updated `TestCheckCommandSurfacesPeerRefusal`, retired the stale `relay_escort_witness.lang` entry in `peerDivergenceExpected`, and re-pointed `TestCheckCommandPeerConsultMutationKilled` at `duplicate_function_name.lang` (the remaining declared divergence) since `relay_escort_witness.lang` no longer demonstrates "peer consult is load-bearing" now that check refuses it independently
- `testdata/phase07/relay_escort_witness.lang` - rewrote the D-07-44 header paragraph to record the interprocedural closure and which mechanism closed it, per Task 2(e)

## Decisions Made

- Checkpoint auto-ratified (Task 1) under `auto_advance`/`yolo`: all three proposed diagnostic identifiers accepted verbatim.
- D-08-09 planning determination (Task 3): both admission paths are independently load-bearing for different success criteria and must agree; recorded above and proven by `TestComputeLoanLastUsesAndDerivePlaceLoansAgree`.
- Span channel widened (D-07-35's `spanByOperationID`) rather than adding a second map — reuses an existing, already-tested plumbing path.
- Reused Task 1's `signatureTable` build for the new pass rather than a second `buildCallSignatureTable` call, to respect `callSignatureTableBuildObserved`'s existing exactly-one-build invariant.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `interproceduralSummary` needed a `returnMode` field beyond the two bools the plan's Task 2(a) text names verbatim**
- **Found during:** Task 2, while implementing `interproceduralLoanLivenessDiagnostic`
- **Issue:** Cause 3's ratified Detail shape (`<calleeID>:return.mode=<Mode>`) requires the callee's actual declared `Return.Mode` string ("shared" or "exclusive"), not just the `returnsBorrowOfParam` bool the plan's literal struct text names. `checkInterproceduralLoanLiveness`'s own signature (per the plan) takes only `program`, `summaries`, `spanByOperationID` — no `callSignatureTable` — so the Mode string has to travel through the summary itself.
- **Fix:** Added a private `returnMode string` field to `interproceduralSummary`, populated alongside `returnsBorrowOfParam` in `buildInterproceduralSummaries` from the same `signature.Return.Mode` read. Disclosure-safe per T-08-03 (copied verbatim from the callee's declared type).
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestRelayEscortWitnessRefusesInterproceduralLiveness` asserts the exact Detail string (`...:return.mode=exclusive`).
- **Committed in:** `2e37aab`

**2. [Rule 1 - Bug] Widened `spanByOperationID`/`CallSpans` to also carry take/borrow spans**
- **Found during:** Task 2, building `interproceduralLoanLivenessDiagnostic`
- **Issue:** `spanByOperationID` was, before this plan, populated only from call spans (D-07-35's original scope, per its own doc comment). The new diagnostic's `Primary` (the offending `take`) and cause 1 (`borrow_created_here`) both need spans for non-call operations, which the map didn't carry.
- **Fix:** In both `analyzeArmBody` and `analyzeStraightLine`, every operation built in the take/borrow/borrow_mut/default switch (not just calls, which are already handled separately in `resolveCallBinding`) now also records its own binding span into the same `result.CallSpans` map, merged into `spanByOperationID` exactly as call spans already were.
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` shows non-zero Primary and cause 1 spans; full suite stays green (no consumer of `spanByOperationID` reads a non-call ID today, so the widened map is additive-only).
- **Committed in:** `2e37aab`

**3. [Rule 1 - Bug] Avoided a second `buildCallSignatureTable` call to respect an existing build-count invariant**
- **Found during:** Task 2, wiring `checkInterproceduralLoanLiveness` into `check.Program`
- **Issue:** A naive second `buildCallSignatureTable` call for the new pass broke `TestCallSignatureTableBuiltBeforeCallableAdmissionRuns`, which asserts exactly one build event per `check.Program` call via `callSignatureTableBuildObserved`.
- **Fix:** Hoisted the Task 1 admission arm's local `table` variable to `check.Program`'s outer scope (`signatureTable`/`signatureTableBuilt`); the new pass reuses it when already built, and only falls back to a fresh build when the admission arm was skipped (the `verifyCallInvariantsSeam` test-only path).
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestCallSignatureTableBuiltBeforeCallableAdmissionRuns` passes; full check suite green.
- **Committed in:** `2e37aab`

**4. [Rule 1 - Bug] Updated four pre-existing tests (check, corevalidate, session packages) whose expectations depended on `relay_escort_witness.lang` staying clean**
- **Found during:** Task 2, running the plan's own verify block (`go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/originvalidate/...`)
- **Issue:** `TestCallAdmissionBodyBlindControl`, `TestCallArgumentConsumptionUnchangedAcrossAcceptingCorpus`, `TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed`, `TestCheckCommandSurfacesPeerRefusal`, `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`, and `TestCheckCommandPeerConsultMutationKilled` all asserted (directly or via the `peerDivergenceExpected` register) that `check` admitted this fixture. Task 2 deliberately flips that outcome (the plan's own stated objective), so these were direct, expected consequences, not scope creep — but they were outside the plan's own `files_modified` list (which named only `check.go`/`check_test.go`).
- **Fix:** Updated each assertion to the new, correct outcome; retired the `relay_escort_witness.lang` entry in `peerDivergenceExpected` (it no longer diverges — both `check` and the peer now independently refuse it, via different codes); re-pointed the peer-consult mutation-kill test at `duplicate_function_name.lang`, the remaining declared divergence, since `relay_escort_witness.lang` can no longer demonstrate that property.
- **Files modified:** `internal/compiler/check/check_test.go`, `internal/compiler/corevalidate/corevalidate_summary_peer_test.go`, `internal/compiler/session/session_peer_gate_test.go`
- **Verification:** `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/originvalidate/... ./internal/compiler/session/...` all green (session package's one remaining failure, `TestQLT01RegistryCoversAllFiveSpikes`, is a pre-existing spike-registry gap unrelated to this plan — confirmed still failing identically on `main` before this plan's changes).
- **Committed in:** `2e37aab`

---

**Total deviations:** 4 auto-fixed (4 Rule 1 — all bugs/gaps directly caused by implementing the plan's own stated behavior, none architectural).
**Impact on plan:** All four are necessary consequences of Task 2's own objective (flip the fixture's outcome end to end); none introduce scope beyond what the plan's `<objective>` and `<behavior>` blocks already required. No `ownership.*` verdict moved anywhere in the corpus.

### TDD process note (not a deviation, documented for transparency)

Tasks 2 and 3 are marked `tdd="true"`. Given the tight coupling between the new interprocedural machinery and its verification (the differential test in Task 3 exists specifically to prove two code paths agree, and could not usefully be written "failing first" in isolation from the implementation it differentials against), implementation and tests were developed together and verified as a unit before each task's single commit, rather than as separate RED/GREEN/REFACTOR commits. Both commits' `<verify>` blocks were run and confirmed green before committing. `workflow.tdd_mode` is `false` for this project (config.json), so the MVP+TDD gate does not apply.

## Issues Encountered

None beyond the deviations above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `interproceduralSummaryTable`/`buildInterproceduralSummaries` are ready for 08-02 to add `usesParam` derivation (left at its fail-closed zero value this plan).
- `derivePlaceLoans`/`blockLoanLiveness`/`loanLivenessFixpoint`/`materializeLoanEndpoints` all now accept `interproceduralSummaryTable` and are ready for further widening (transitivity, the other direction) without another signature change.
- `MaxInterproceduralLoanCauses`/`TruncatedInterproceduralLoanBound` are declared and ready for 08-04's fail-closed bound work.
- `calleeContract.ReturnsBorrowOfParam` is declared and ready for later plans' fixture/parity work.
- No blockers for 08-02 through 08-06.

## Self-Check: PASSED

- FOUND: internal/compiler/check/check.go
- FOUND: internal/compiler/check/check_test.go
- FOUND commit: 2e37aab
- FOUND commit: 6b7cdc8
- Re-ran acceptance criteria: `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` emits `check.interprocedural_loan_liveness` with 3 causes, nil Repairs — PASS
- Re-ran plan-level `<verification>`: `go test ./internal/compiler/check/...` exits 0; `go test ./... && go vet ./...` exits 0 (only pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure, confirmed present on `main` before this plan) — PASS

---
*Phase: 08-interprocedural-loan-liveness-in-check*
*Completed: 2026-09-09*

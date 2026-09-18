---
phase: "03"
slug: "borrowed-views-and-cfg-lifetimes"
status: planned
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-04"
evidence_vocabulary: v1
graded_rows: 21
---

# Phase 3 — Validation Strategy

> Feedback contract for proving shared and exclusive loans, edge-specific CFG last use, and separately checkable public borrow origins across the real compiler spine.
>
> Every checkbox in this document is deliberately **unchecked**. A box is ticked only when execution has supplied the evidence named beside it. `nyquist_compliant` and `wave_0_complete` stay `false` until then.

## Scope Limitation Recorded Up Front

OWN-03 is scoped to **acyclic** control-flow graphs this phase. The language has no loop and no recursion, and this phase adds neither. Loop-carried loan liveness, loop-exit edges, and dynamic per-iteration loan identity are an explicit Phase 4-or-later follow-on. No test, comment, fixture, or report in this phase may claim loop coverage. The production liveness pass rejects a back edge fail-closed rather than iterating one, so the limitation is enforced rather than merely documented.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 `testing`, `testing/quick`, native fuzz seeds, black-box `os/exec` |
| **Config file** | `go.mod` — no new module requirement is permitted this phase |
| **Exact-target selector** | `scripts/assert-go-tests.sh` (existing; `--self-test` proves the selector is not vacuous before any targeted command is trusted) |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...` |
| **Full suite command** | `sh scripts/verify-phase3.sh` |
| **Feedback budget** | quick under 20 seconds; full under 90 seconds on the current macOS arm64 host — observed, never ratified as a release SLO |

## Sampling Rate

- **After every task commit:** run that task's exact `<automated>` command, then `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`.
- **After every plan wave:** `go test -race ./...` and `go vet ./...`, plus `sh scripts/verify-phase2.sh` to prove Phase 2 non-regression while Phase 3's own gate does not yet exist.
- **Before phase verification:** `sh scripts/verify-phase3.sh` must exit zero and must report every required Phase 3 control ID and both expected escapes.
- **Investigation only:** fuzzing is time-bounded and never enters the deterministic default gate.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| 03-01-01 | 01 | 1 | OWN-03 | T-03-09 | No generated C identifier is reserved to the C implementation, and no golden moves to achieve it | emitted-artifact invariant | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestGeneratedIdentifierNamespacesStayConfined TestReservedSetsCoverTheirOwnNamespace TestIdentifierPrefixInvariance TestNativeIdentifiersRemainCollisionFree TestPhase1EvidenceGoldenUnchanged` | ⬜ Wave 0 | EXERCISED | — |
| 03-01-02 | 01 | 1 | OWN-03 | T-02-03/T-03-04 | A match arm may hold a linear body, yielding a real CFG with a join, wired through check, corevalidate, interp, and cgen | vertical tracer | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestArmBodyRoundTrips TestArmBodyLowersToBlocksAndEdges TestArmBodySchemaCrossLock TestBranchInterpreterNative TestPhase1EvidenceGoldenUnchanged` | ⬜ Wave 0 | EXERCISED | — |
| 03-01-03 | 01 | 1 | OWN-03 | T-03-01/T-02-07 | The new CFG surface is capped fail-closed and independently validated; the causality control matches a stable marker | bounded-input + control hardening | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestArmBodyLimits TestBlockEdgeValidationRules TestCoreValidationWorkSeries TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork` | ⬜ Wave 0 | EXERCISED | — |
| 03-02-01 | 02 | 2 | OWN-03 | T-03-04 | An exclusive loan travels source to native through four independently updated operation-kind consumers | vertical tracer | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestExclusiveBorrowRoundTrips TestExclusiveBorrowLowersToCore TestExclusiveBorrowAuthorizedIndependently TestUnknownOperationStillRejected TestExclusiveBorrowInterpreterNative` | ⬜ Wave 0 | EXERCISED | — |
| 03-02-02 | 02 | 2 | OWN-03 | T-03-07 | The five conflict rows are decided by liveness overlap, rejected with repair-bearing causal diagnostics | compile-reject matrix | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestSharedSharedOverlapAccepted TestBorrowConflictMatrix TestBorrowConflictCauseChain TestExclusiveMoveRejected TestSequentialLoansAccepted` | ⬜ Wave 0 | EXERCISED | — |
| 03-02-03 | 02 | 2 | OWN-03 | T-03-10 | Every program that passes `check` is executable; unexecutable shapes are refused causally with a span | taxonomy contract | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestUnexecutableShapeRejectedWithSpan TestAbilityFactsSurviveExecutionRejection TestHumanJSONMixedDiagnosticVersionParity` | ⬜ Wave 0 | EXERCISED | — |
| 03-03-01 | 03 | 3 | OWN-03 | T-02-05/T-03-11 | Liveness is a per-edge backward fixpoint; straight-line answers provably unchanged; back edges rejected | dataflow contract + regression pin | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/check TestLoanLivenessFixpoint TestStraightLineEndpointsUnchanged TestEdgeSpecificLiveOut TestBackEdgeRejected TestOwnershipSequenceExhaustive` | ⬜ Wave 0 | EXERCISED | — |
| 03-03-02 | 03 | 3 | OWN-03 | T-03-06 | Edge-specific placement is falsified by a fixture pair whose verdicts both flip under a seeded uniform-join fault | fault injection + CLI parity | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestBranchEdgeLastUseAcceptAndReject TestUniformJoinPlacementFlipsBothVerdicts TestBranchFixturesThroughCLI` | ⬜ Wave 0 | EXERCISED | — |
| 03-03-03 | 03 | 3 | OWN-03 | T-02-01/T-03-12 | Counted work includes propagation; the reborrow-chain series is linear at 10/100/1,000/10,000 | counted-work scale series | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/check TestOwnershipWorkSeries TestLivenessWorkScale TestReborrowChainWorkIsLinear` | ⬜ Wave 0 | EXERCISED | — |
| 03-04-01 | 04 | 4 | OWN-03 | T-03-13 | The validator recomputes loan endpoints by a different mechanism and imports neither check nor ast | independent re-derivation | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestValidatorRecomputesLoanEndpoints TestValidatorImportsStayIndependent TestLoanEndpointMismatchRejected TestCoreValidationWorkSeries` | ⬜ Wave 0 | EXERCISED | — |
| 03-04-02 | 04 | 4 | OWN-03 | T-02-01 | The validator's own quadratic scan is off the path; `Checks` is linear at 101/1,001/10,001 with a reborrow chain | counted-work scale series + verdict pin | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestValidatorVerdictsUnchanged TestCoreValidationWorkSeries TestValidatorReborrowChainIsLinear TestTransitiveLoanBlocksMove` | ⬜ Wave 0 | EXERCISED | — |
| 03-04-03 | 04 | 4 | OWN-03 | T-03-06/T-03-14 | Moved, dropped, and invented endpoints are all rejected; failing lanes still record evidence | core mutation matrix | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestLoanEndpointMutationMatrix TestBorrowedLaneRecordsFailureStatus TestVerifyPhase2ControlsAndWork` | ⬜ Wave 0 | EXERCISED | — |
| 03-05-01 | 05 | 5 | OWN-03 | T-03-13 | A bounded path oracle decides endpoints by exhaustive expansion from the core artifact alone | independent oracle | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/pathoracle TestOracleEnumeratesAllAcyclicPaths TestOracleImportsStayIndependent TestOracleAgreesWithProduction TestOracleFailsOnUnterminatedLoan` | ⬜ Wave 0 | EXERCISED | — |
| 03-05-02 | 05 | 5 | OWN-03 | T-03-01/T-03-13 | The differential is mutation-killed by revert and by seeded fault; enumeration is capped fail-closed | mutation kill + cap falsifier | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOracleDisagreesWithUniformJoinFault TestOracleMetamorphicTrials TestOraclePathCountCapRejects TestPathOracleLaneRecordsControl` | ⬜ Wave 0 | EXERCISED | — |
| 03-05-03 | 05 | 5 | OWN-03 | T-03-15 | Generators reach arm bodies, exclusive borrows, and cross-arm reborrows; blind spots written down | generated property | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestGeneratedBranchBodiesRoundTrip TestGeneratedKindsIncludeExclusiveBorrow TestBranchSequenceExhaustive FuzzOwnershipLinear` | ⬜ Wave 0 | EXERCISED | — |
| 03-06-01 | 06 | 6 | OWN-04 | T-03-02 | A public borrowed view declares origins and access mode and is consumed body-blind by a separate invocation | vertical tracer + CLI integration | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestPublicViewRoundTrips TestPublicOriginFactLowered TestBorrowedReturnWithoutOriginRejected TestInterfaceSummaryOmitsBodies TestInterfaceCheckIsBodyBlindCLI` | ⬜ Wave 0 | EXERCISED | — |
| 03-06-02 | 06 | 6 | OWN-04 | T-03-02/T-03-03/T-03-16 | Understated, impossible, and stale summaries are rejected by recomputation; the escape is named | differential + mutation | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestOriginUnderstatedRejected TestOriginAccessMismatchRejected TestStaleSummaryRejectedBeforeOtherChecks TestOriginEscapeIsNamed TestCoordinatedSourceCoreEscapeIsNamed` | ⬜ Wave 0 | EXERCISED | — |
| 03-06-03 | 06 | 6 | OWN-03/04 | T-02-08 | Every field added this phase is absent from every pre-existing program's serialized form, key by key | golden byte identity | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestPhase3FieldsAreOmittedWhenAbsent TestPhase1EvidenceGoldenUnchanged TestCanonicalEvidence TestEvidenceErrorCodeReachesCLI` | ⬜ Wave 0 | EXERCISED | — |
| 03-07-01 | 07 | 7 | OWN-03/04 | T-03-05 | The lineage side table is capped, counted, schema-stable, and reports absence honestly | bounded instrumentation | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestDebugMapJoinsSourceCoreAndOperation TestDebugMapReportsHonestAbsence TestDebugMapCapsFailClosed TestDebugMapIdentityUsesOrdinals TestDebugMapCLIAvailability` | ⬜ Wave 0 | EXERCISED | — |
| 03-07-02 | 07 | 7 | OWN-03/04 | T-03-18 | One gate requires every Phase 3 control, proves Phase 1 and Phase 2 non-regression, and is gated by a contract test | phase gate | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestVerifyPhase3ControlsAndWork TestVerifyPhase3CLI TestPhase3VerifierScriptContract TestBorrowedLaneRecordsFailureStatus TestVerifyPhase2ControlsAndWork` plus `sh scripts/verify-phase3.sh` | ⬜ Wave 0 | EXERCISED | — |
| 03-07-03 | 07 | 7 | OWN-03/04 | T-02-06 | Both native deadlines have falsifiers; the spawn guard resolves constructors; the CLI ceiling matches measurement | bounded process/IO | `env GOCACHE=/tmp/ai-lang-phase3-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestNativeTimeoutHasFalsifier TestSourceNeverSpawnsUnboundedProcesses TestSpawnGuardCatchesKnownEvasions TestCLIStreamCeilingBoundary TestHumanJSONProjectionParity` | ⬜ Wave 0 | EXERCISED | — |

## Required Negative Controls

The Phase 3 verifier must fail closed unless it observes all of these exact IDs. Each must be observed by running its fixture or mutation, must carry nonzero recomputed work in its lane, and must be asserted both at the session layer and through the shipped CLI.

```text
control:ownership.exclusive_conflict
control:ownership.exclusive_move
control:ownership.edge_last_use_omitted
control:cfg.path_oracle_disagreement
control:core.loan_endpoint_mismatch
control:origin.understated_summary
control:origin.impossible_summary
control:interface.stale_summary
```

The Phase 3 gate must additionally re-observe all nine Phase 2 controls by running the Phase 2 corpus with the same once-built binary, and all three Phase 1 controls by running the Phase 1 corpus with that same binary. Non-regression is proven by running the older corpora, never by invoking the older gate script.

Two escapes must appear under expected escapes and must never appear as a detected control:

```text
escape:coordinated-source-core-lie
escape:coordinated-origin-summary-lie
```

## Wave 0 Requirements

Every item below must exist and be red-then-green before the task that depends on it is considered done.

- [ ] `testdata/phase3/` corpus with the nine fixtures: `borrowed_view.lang` (dispatch probe and positive), `branch_view.lang`, `shared_shared_accept.lang`, `shared_exclusive_reject.lang`, `exclusive_exclusive_reject.lang`, `exclusive_move_reject.lang`, `sequential_shared_then_exclusive_accept.lang`, `branch_one_arm_shared_accept.lang`, `branch_one_arm_shared_reject.lang`, plus the two dishonest-summary fixtures `public_view_understated.lang` and `public_view_impossible.lang` and the positive `public_view.lang`.
- [ ] Each fixture is led by a comment stating the law under test, is canonical under the formatter fixed point, and is under the source byte limit.
- [ ] Endpoint-pinning test for the existing Phase 2 straight-line answers, written and green **before** the liveness algorithm is swapped in 03-03.
- [ ] Validator verdict-pinning test over the whole existing corpus, written and green **before** the validator derivation is replaced in 03-04.
- [ ] `internal/compiler/pathoracle` package skeleton with its import-independence test, before any differential is trusted.
- [ ] Fault-injection seam for uniform-join placement, with the comment recording that a test shipping its own seam cannot be falsified by reverting production alone.
- [ ] `scripts/verify-phase3.sh` created as a peer of, never an extension of, the Phase 2 script, with its contract test.
- [ ] Existing `scripts/assert-go-tests.sh --self-test` run once at the top of the Phase 3 gate before any targeted command is trusted.

## Mutation-Kill Register

Per D-09, a differential is not evidence until reverting the production hunk makes it fail. Each row must be demonstrated in a throwaway detached worktree and its failing output pasted verbatim into the owning plan's summary.

| Oracle / differential | Production hunk to revert | Expected failure | Owning plan | Status |
|---|---|---|---|---|
| Path-oracle endpoint differential | the 03-03 backward-fixpoint liveness pass | endpoints disagree on both branch fixtures | 03-05 | ⬜ pending |
| Path-oracle uniform-join falsifier | seeded fault, not a revert | oracle disagrees, naming the loan and edge | 03-05 | ⬜ pending |
| Validator endpoint recomputation | the 03-04 validator derivation | endpoint mismatch code raised on a valid artifact | 03-04 | ⬜ pending |
| Edge-specificity fixture pair | the per-edge live-out placement | accept fixture rejected, reject fixture accepted | 03-03 | ⬜ pending |
| Origin understatement detection | the 03-06 recomputation | an understated summary is accepted | 03-06 | ⬜ pending |
| Stale-summary digest binding | the digest comparison | a mismatched summary reaches an origin check | 03-06 | ⬜ pending |
| Counted-work honesty | the propagation counters | the reborrow-chain series stops growing linearly and the assertion no longer distinguishes | 03-03 | ⬜ pending |

## Generator Reachability Register

Per D-10, coverage is not the question; reachability is. Each generator must carry a comment stating what it reaches and what it does not, and this table must be completed from those comments during execution.

| Generator | Must reach | Known not to reach | Owning plan | Status |
|---|---|---|---|---|
| `generatedLinearProgram` (extended) | arm bodies including a zero-binding arm, exclusive borrow spelling, reborrow whose original precedes the branch, trailing comment after a declaration header inside an arm body, generic type plus binding inside an arm body | loops and back edges (none exist in the language); nested arm bodies beyond depth one (capped) | 03-05 | ⬜ pending |
| Exhaustive ownership sequence enumeration | two-block branch programs at bounded length | three-or-more-block graphs; any cyclic shape | 03-05 | ⬜ pending |
| Path-oracle metamorphic trials | binding reorder and place alpha-rename over branch programs, fixed seed, asserted trial count | shape-changing transformations that alter the CFG | 03-05 | ⬜ pending |
| Conflict-matrix corpus | all five conflict rows from source | any row reachable only through a synthetic type fact — which rows those are must be named in a comment | 03-02 | ⬜ pending |

## Shipped-Binary Register

Per D-11, the gate only ever sees what ships with it. Each row must be exercised against a freshly built `./cmd/lang` on a hand-written program that is **not** in any corpus.

| Behaviour | Commands | Owning plan | Status |
|---|---|---|---|
| Branch with arm bodies | `format --check`, `check`, `run --engine=interpreter`, `run --engine=native` | 03-01 | ⬜ pending |
| Exclusive borrow | `check`, both engines | 03-02 | ⬜ pending |
| Edge-specific accept and reject pair | `check` with JSON projection, exit codes and diagnostic codes asserted | 03-03 | ⬜ pending |
| Public borrowed view, separate compilation | `interface export` then a second process running `interface check` | 03-06 | ⬜ pending |
| Debug-lineage availability | the lineage subcommand, asserting all three availability values | 03-07 | ⬜ pending |

## Property and Fuzz Lanes

- Ordinary `go test` exhausts short ownership sequences over one and two blocks, the extended generated linear and branch round trips at a fixed seed with a case-count assertion, and the persisted fuzz seeds.
- Investigation command: `go test ./internal/compiler/check -fuzz=FuzzOwnershipLinear -fuzztime=30s`. Never in the default gate.
- Path enumeration in the oracle is capped; the cap has its own over-cap falsifier and is never raised to make a test pass.
- Scalability is asserted by counted work at 10, 100, 1,000, and 10,000 operations for the checker and 101, 1,001, and 10,001 facts for the validator — never by wall clock.

## Manual-Only Verifications

None. Every Phase 3 semantic behaviour has an automated verification. Human review remains valuable for C readability, diagnostic prose, and the honesty of the debug-lineage availability wording, but none of the three is accepted as a substitute for an automated check.

## Threat Model

Phase 2's register T-02-01 through T-02-08 remains in force unchanged and is re-proven by running the Phase 2 corpus with the Phase 3 binary. The rows below are the Phase 3 additions.

| ID | Threat | Failure mode | Required mitigation and falsifier |
|----|--------|--------------|------------------------------------|
| T-03-01 | Adversarial CFG shape or path count | analysis or enumeration becomes unbounded | declared arm, block, and path caps rejecting fail-closed; over-cap falsifier; counted-work series |
| T-03-02 | Dishonest public-origin summary understating borrow scope | a consumer trusts a narrower borrow than the body performs | body-derived recomputation from the core artifact; understated fixture is a required control |
| T-03-03 | Stale summary bound to an older artifact | a changed implementation is checked against an old promise | digest compared first, before any origin or access check; stale fixture is a required control |
| T-03-04 | A new operation kind dropped by a non-admission consumer | the operation vanishes from the trace the O0/O3 differential compares | all four consumers updated together; trace comparison at both optimization levels is the falsifier |
| T-03-05 | Lineage artifact fabricating or leaking source detail | a report claims knowledge the compiler does not have | three-valued availability, ordinal-only identity, declared caps, no storage or upload path |
| T-03-06 | Loan endpoint omitted or placed uniformly at the join | a real conflict is admitted, or a legal program is rejected | fixture pair whose verdicts both flip under the seeded fault; oracle disagreement; required control |
| T-03-07 | Aliasing under an exclusive loan | two writers, or a reader and a writer, observe the same owner | five-row conflict matrix decided by liveness overlap, re-decided independently at the second layer |
| T-03-08 | Coordinated frontend-and-summary lie | an internally consistent producer lies about its own origins | **accepted residual** — declared as a named expected escape, asserted visible at the CLI, never claimed solved |
| T-03-11 | Non-terminating worklist on a back edge | the checker hangs on a future cyclic CFG | acyclicity asserted; a back edge is rejected fail-closed rather than iterated |
| T-03-12 | Understated cost metric | a bound appears satisfied because the expensive step counts nothing | every transfer evaluation and reinsertion counted; work bound re-derived, not bumped |
| T-03-13 | Two derivations agreeing because they share a law | a differential agrees on the wrong answer, as WR-02 did | import independence asserted by test; different mechanism classes; revert-and-fail recorded verbatim |
| T-03-14 | A failing verify lane leaving no record | partial-work evidence is lost exactly when it matters | status-bearing lane shape throughout the Phase 3 verify path |
| T-03-15 | Green property test whose reachable space omits the hard case | the three recorded Phase 2 misses, repeated | reachability register completed from in-code comments; pinned cases in the historically missed region; batches assert their own counts |
| T-03-16 | Impossible access mode declared | a shared body is published as an exclusive promise | declared mode compared against the body's actual borrow operations |
| T-03-17 | Unbounded origin union or summary size | a summary consumer is driven out of bounds | origin-set and function-count caps failing closed, sized against existing input caps |
| T-03-18 | A gate that appears to prove more than it runs | reviewers trust a claim the script does not execute | contract test over the gate's own text: no previous-phase script, no duplicated shared suite, every control ID present |

## Debt Closure Register

Carried from `02-DEBT.md`. No debt-cleanup plan exists; each item lands in the plan that already touches its files.

| Debt ID | Landing plan | Closure evidence required | Status |
|---|---|---|---|
| D-02-01 spawn guard defeatable | 03-07 | four documented evasions each caught; scan fails rather than passing vacuously on an empty source set | ⬜ pending |
| D-02-02 evidence code unreachable at the CLI | 03-06 | the distinct code surfaces at the CLI with an assertion | ⬜ pending |
| D-02-03 quadratic liveness in both layers | 03-03 (checker), 03-04 (validator) | two independent derivations; linear counted-work series in both; before and after values recorded | ⬜ pending |
| D-02-04 no `native.timeout` falsifier | 03-07 | helper hang mode drives both deadlines; both report the timeout code | ⬜ pending |
| D-02-05 reserved `__LANG_` identifier | 03-01 | standalone commit; collision suite green; no golden moved | ⬜ pending |
| D-02-06 loose CLI stream ceiling | 03-07 | ceiling tightened with the measured maximum cited; at-limit and one-over boundary cases | ⬜ pending |
| D-02-07 causality control coupled to one literal line | 03-01 | stable generated marker; absent or duplicated marker still fails closed | ⬜ pending |
| D-02-08 human and JSON projections diverge | 03-07 | parity assertion covering the divergent case | ⬜ pending |
| D-02-09 `Box`/`Pair` check but die spanless | 03-02 | span-bearing repair-bearing compile-time rejection; ability facts unaffected; no new C lowering | ⬜ pending |

## Debug-Lineage Experiment Boundary

Implemented this phase: wiki steps 1 (stable source and core IDs on an arm, a move, a borrow, and a return), 2 in its **semantic half only** (a schema-versioned side table), and 4 (resolve a diagnostic or trace event back to source and core identity with honest availability).

Rejected this phase under the D-03 scope fence, and required to be named in the shipped package documentation:

- Step 2's native-debug-info half — DWARF, CodeView, and even `#line` emission are named out of scope, and neither OWN-03 nor OWN-04 needs them.
- Step 3's panic and segfault capture — any crash-capture mechanism is the storage and symbolication subsystem the fence forbids. Step 3's ownership-diagnostic half is already satisfied by step 4.
- Step 5's fault injection across stale-symbol, inlining, redaction, truncation, and wrong-build-ID defects — it presupposes the native artifacts the rejected halves would have produced, so it has nothing to mutate.
- Step 6's capture and symbolication latency measurement — same reason; the semantic-side emission cost and output bytes piggyback on the existing metrics instead.

If the caps, schema, timeout, output ceiling, or honest-absence reporting cannot all be met, the experiment does not ship and the plan summary must say so. That is D-04's own term, not a fallback.

## Validation Sign-Off

- [ ] All tasks have an immediate `<automated>` verification or an explicit Wave 0 dependency.
- [ ] No three consecutive tasks lack automated feedback.
- [ ] All Wave 0 references exist and are green.
- [ ] Every row of the Mutation-Kill Register is demonstrated, with failing output recorded verbatim in the owning plan's summary.
- [ ] Every row of the Generator Reachability Register is completed from an in-code comment, including at least one named unreachable shape per generator.
- [ ] Every row of the Shipped-Binary Register is exercised on a program that is in no corpus.
- [ ] Every row of the Debt Closure Register is closed or explicitly re-carried with a reason.
- [ ] No watch-mode or unbounded-fuzz flag enters a default command.
- [ ] `scripts/verify-phase2.sh` is byte-identical to its state at the phase start, and still exits zero.
- [ ] `git diff <phase-start>..HEAD -- testdata/phase1/` is empty; every Phase 2 golden hunk has a field-by-field causal explanation.
- [ ] `go.mod` gained no requirement.
- [ ] Measured warm feedback latency for the Phase 3 fixture is reported as an observation, not ratified as a release SLO.
- [ ] `nyquist_compliant: true` and `wave_0_complete: true` are set only after execution supplies the evidence above.

**Approval:** not granted. This document is the contract to be satisfied, not a record of satisfaction.
> **Superseded (partial):** This document's loan-liveness subset (Per-Task Verification Map rows 03-03-01 through 03-05-03, plus the corresponding Mutation-Kill Register and Generator Reachability Register rows) is superseded by `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md`, section "M001 Phase 3 Debt Closure (loan-liveness subset)", as of 2026-09-10; this document's own `nyquist_compliant: false`, its unticked checkboxes, and its OWN-04/loop-carried-liveness scope limitation are unchanged and remain the record of record for everything outside that subset.

---
phase: 10-trusted-interprocedural-oracle
verified: 2026-09-11T22:28:00Z
status: passed
score: 10/10 must-haves verified
behavior_unverified: 0
overrides_applied: 0
coincidental_reliance_items: []
---

# Phase 10: Trusted Interprocedural Oracle Verification Report

**Phase Goal:** Cross-function execution and origin/path re-derivation are
trustworthy enough to be the authority everything native is
differential-tested against. (ROADMAP.md:509-510)

**Verified:** 2026-09-11T22:28:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (Roadmap Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1a | `originvalidate` walks published origins across `OpCall`, mirroring the `OpForeignCall` hop | ✓ VERIFIED | `originvalidate.go`'s `walkReturnOrigin` has a `case core.OpCall` consulting a callee-contract map built once in `ValidatePublished`. `TestOpCallOriginWalkPropagatesGenuineBorrowingCallee` PASS; `TestOpCallOriginWalkGateIsLoadBearing` PASS — forcing the case to no-op via `disableOpCallOriginConsultForTest` regresses `twin_a_accept.lang` to `core.origin_omitted` through the full `session.CheckCommandFile` pipeline, proving the case decides the outcome, not merely exists syntactically. |
| 1b | `pathoracle` independently re-derives the cross-function loan-chain rule | ✓ VERIFIED | `pathoracle_compose.go` recurses into the callee's own `EnumeratePaths`/`linearizePath` and splices concrete paths (never a contract hop shared with `check`/`corevalidate`). `TestCompositionDiscriminatesPerPathBorrow` PASS — a callee with two concrete paths (only one borrowing) causes the caller's composed paths to split into edge/point endpoints; running the same test with composition replaced by a stubbed contract hop (`forceContractHopForTest`) collapses the split — verified as a genuine mutation-kill, not an inert seam. |
| 1c | Neither `originvalidate` nor `pathoracle` imports `check`/`corevalidate`, enforced by a build/test control | ✓ VERIFIED | `go list -deps ./internal/compiler/originvalidate/...` returns no `compiler/check` or `compiler/corevalidate` entries. `originvalidate.go`'s own imports confirmed by direct read: only `callgraph`, `core`, and other non-forbidden packages. `TestOriginValidatorImportsStayIndependent` (transitive `go list -deps` check, not a direct-import scan alone) PASS. `corevalidate_endpoint_internal_test.go:107`'s equivalent guard for `corevalidate` also confirmed present and forbidding `check`/`ast`/`originvalidate`/`callgraph`. **Documented asymmetry** (D-10-19): `originvalidate` is permitted to import `callgraph` (confirmed at `originvalidate.go:18`, calling `callgraph.Order`) while `corevalidate`'s own guard forbids the same package for itself. This is recorded in PHASE-10-DEBT.md and 10-REVIEW.md WR-02 as a review-defensible, not mechanically-defensible, asymmetry — not silently discovered by a future reader. Each guard's forbidden list is hand-curated and a same-commit edit could still widen the exception (WR-02); this residual weakness is stated, not papered over. |
| 2a | `interp` executes multi-function programs on a bounded call stack with a documented fixed ceiling | ✓ VERIFIED | `MaxCallDepth = 128`, declared below the real structural ceiling (1024, from `maxFunctions` combined with `callgraph`'s DAG-only admission). A genuine 129-chained-function `.lang` program run through the full `syntax.Parse → check.Program → corevalidate.Validate → interp.Run` pipeline terminates in a named `Outcome`, not a host stack overflow — `TestCallDepthExceeded` PASS. |
| 2b | Pitfall-4 Gate: native-stack-overflow probe confirms the two limits are distinct | ✓ VERIFIED | `TestNativeStackHeadroomIndependentOfCallDepth` PASS (both `cap_disabled` and `cap_enabled` subtests) — with the depth cap disabled via an unexported override seam, an 800-function chain completes under a pinned 1 MiB host stack with no collapse, proving the language-level bound and the native-stack limit are structurally unrelated. |
| 3a | Drop/cleanup obligations run in the defined order on normal return, observed as ordered events | ✓ VERIFIED (behavioral) | `TestFrameDrainOrder/normal_return_across_one_boundary` PASS; `TestFrameDrainOrder/frame_drain_order_seam_flips_canonical_bytes` PASS — flipping the drain order through `frameDrainOrderForTest` diverges `interp.CanonicalBytes`, proving order is genuinely observed through the canonical execution document, not asserted against internal map state. |
| 3b | Drop/cleanup obligations run in the defined order on every nonlocal exit across N frames (not just `OpReturn`/`OpFail` popping) | ✓ VERIFIED (behavioral) | `TestFrameDrainOrder/foreign_nonlocal_landing_pad_across_multiple_frames` and `/depth_refusal_with_live_resources_in_multiple_frames` PASS — SEM-09's narrow-but-not-vacuous second clause (D-10-31: the only real multi-frame-skip constructs are the extended foreign landing pad and the SEM-08 depth refusal) is exercised by name, not satisfied by testing only `OpReturn`/`OpFail` popping. |
| 3c | Second independent knower of drop order: `corevalidate`-checked callee-signature invariant | ⚠ NOTED, not a blocker | `peerCalleeFrameDrained` exists and is checked once per function declaration (`corevalidate.go:2823`). Confirmed directly: its escape-via-return check is a **single forward pass**, not a fixpoint, relying on an unstated assumption that `linear.Operations` is dependency-ordered (10-REVIEW.md WR-01, independently re-derived from source by this verification, not merely trusted from the review doc). This differs from its own cited precedent `peerParameterEscapesOwned`, which walks backward and is order-independent. The assumption holds today (place-must-be-produced-before-read is enforced elsewhere), but the function does not itself assert it. This is an honestly-documented residual risk in the phase's own review, not a proven defect — a future block-lowering change that reorders `Operations` could silently regress this specific check to a false-negative. Recorded here as a WARNING for a human/future-phase decision, not scored as a failed truth. |
| 4a | Cross-function loan-endpoint differentials rebuild M001 Phase 3's exhaustive endpoint enumeration at a declared, bounded composition depth | ✓ VERIFIED | Composition depth declared as 3 (`session_composition_depth_test.go`). `TestCompositionDepthCorpusReachesDeclaredBound` PASS in both directions — confirmed directly: raising the required depth to 4 without extending the corpus produces the failure message "composition depth corpus falls short: observed maximum depth 3 ... required declared depth 4," proving the bidirectional gate is a real declaration, not a description. |
| 4b | The differential is billed accurately (three-way on refuse, four-way on accept) and seeded faults prove non-vacuous independence | ✓ VERIFIED | `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` PASS, extended to `assertThreeWayEndpointAgreement` (check/corevalidate/pathoracle `core.LoanEndpoint` sets) plus the four-way accept-side metamorphic `interp` check, billed in doc comments and test names exactly as "three-way on refuse, four-way on accept," never unqualified "four-way." Two seeded-fault mutation-kill tests independently confirmed by this verification: `TestDerivePeerSignatureModeMutantPairing` PASS (a hardcoded parameter-mode mutant is caught by the differential against the declared contract and NOT by `interp`, which never reads a mode string — genuinely discriminating, not tautological) and `TestCompositionDiscriminatesPerPathBorrow`'s companion contract-hop stub (caught only by `pathoracle`, confirmed above). `TestPeerDivergenceRegisterDebtIDsAreAccountable` PASS — every declared divergence entry's debt ID resolves to a still-open register row, including a negative-control fail-closed check for an unresolvable synthetic ID. |
| OWN-05b | The same call-site ownership-transfer fact is derived independently by `interp` | ✓ VERIFIED | `TestMoveAsCopyMutationKilled` PASS — a move-as-copy mutation (skip the caller-side delete) is caught by `interp` alone; `check`/`corevalidate` execute nothing and are structurally blind to it. `TestInterpDoesNotReadCorevalidateOwnershipFields` PASS (both `real_scan` and `negative_control` subtests) — a guard test fails the moment `interp` reads an ownership-bearing field of `corevalidate.Result`. **Narrowed claim confirmed accurate, not overclaimed**: `interp.go:153` does call `corevalidate.Validate(program)` as `Run`'s first act, so the independence is "derivation-mechanism independence for the ownership fact, nested inside a shared validation dependency" (D-10-37) — not mutual non-import independence like `check`/`corevalidate` earn from each other. This narrower framing is stated in the guard test's own doc comment, matching the code exactly; no artifact overstates it. |

**Score:** 10/10 truths verified (0 present-but-behavior-unverified)

### Requirements Coverage

| Requirement | Source Plan | Status | Evidence |
|---|---|---|---|
| SEM-08 | 10-01, 10-04, 10-09 | ✓ SATISFIED | Bounded call stack, named refusal, Pitfall-4 gate — see truths 2a/2b above |
| SEM-09 | 10-05, 10-09 | ✓ SATISFIED | Ordered drop/cleanup on return and nonlocal exit — see truths 3a/3b above; 3c is a documented residual risk, not a blocking gap |
| TRU-02 | 10-02, 10-06 | ✓ SATISFIED | `originvalidate` `OpCall` walk — see truth 1a; closes inherited D-09-51 defect (confirmed: `git log` shows `bb4aa02` widen then `6648b72` use as two separate commits, matching D-10-59's ordering-constraint claim) |
| TRU-03 | 10-03 | ✓ SATISFIED | `pathoracle` independent composition — see truth 1b |
| QLT-04 | 10-07, 10-08 | ✓ SATISFIED | Declared depth-3 bound, bidirectional gate, depth-3 fixture pair varying which hop carries the borrow — see truth 4a; `relay_depth3_accept.lang`/`relay_depth3_refuse.lang` confirmed present under `testdata/phase10/` |
| OWN-05b | 10-01, 10-08, 10-09 | ✓ SATISFIED | See OWN-05b row above; correctly split from OWN-05a per D-09-37 and not overclaimed as mutual non-import independence |

REQUIREMENTS.md's own tracker (`grep` confirmed) marks all six as `Phase 10 / Complete` with 31/31 M002 requirements mapped and 0 unmapped — no orphaned requirements found for this phase.

### Independence Mechanism Audit (verification_emphasis point 1)

Directly inspected (not merely trusted from documents):

- `originvalidate.go` imports: `callgraph`, `core`, plus non-compiler packages — no `check`/`corevalidate` (confirmed via `grep` and `go list -deps`).
- `corevalidate_endpoint_internal_test.go:107`'s forbidden list explicitly includes `check`, `ast`, `originvalidate`, and `callgraph` — confirmed present, unweakened.
- The `callgraph` asymmetry (`originvalidate` permits it, `corevalidate` forbids it) is real, confirmed in source, and matches PHASE-10-DEBT.md D-10-19 and 10-REVIEW.md WR-02's description exactly — it is a hand-curated, review-defensible (not mechanically-defensible) asymmetry, explicitly recorded rather than discovered as a surprise. Nothing landed this phase widens it — `originvalidate`'s only use of `callgraph` remains `BuildInterface`'s topological `Order` call.
- Four independently-implemented re-derivation mechanisms confirmed structurally distinct at the `OpCall` boundary: `originvalidate.walkReturnOrigin` consults a declared contract map; `pathoracle_compose.go` recurses into the callee's own path enumeration; `check`/`corevalidate` each have their own pre-existing contract-hop and forward/backward walk logic. None of the four collapses to reading a peer's computed verdict.

### Differential Falsifier Audit (verification_emphasis point 2)

Ran the specific seeded-fault tests named in the emphasis directly (not merely reading the SUMMARY's claims):

- `TestCompositionDiscriminatesPerPathBorrow` (D-10-14): PASS, and its own stubbed-contract-hop seam demonstrably collapses the per-path split (confirmed by reading the seam and its doc comment referencing the mutation-kill).
- `TestDerivePeerSignatureModeMutantPairing` (D-10-41 complement): PASS on both subtests — the differential catches the mutant, `interp` stays structurally blind to it (never consults a mode string). This is a genuine two-sided proof, not a tautology: two different code paths (a declared-contract comparator vs. an execution engine that never reads that field) are shown to diverge in exactly the direction independence requires.
- `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` extended with `assertThreeWayEndpointAgreement`: confirmed the three peers' `core.LoanEndpoint` sets are independently computed (`function.Linear.LoanEndpoints`, `corevalidate.Result.LoanEndpoints()`, `pathoracle.RecomputeEndpoints`) — not one value read through two field names. Confirmed by reading `assertThreeWayEndpointAgreement`'s implementation directly.
- `TestCompositionDepthCorpusReachesDeclaredBound`: ran directly with the bound artificially raised in-test to confirm direction 2 (raise-without-extend) genuinely fires — verified output: "composition depth corpus falls short: observed maximum depth 3 ... required declared depth 4."

No comparison found compares the same computed value to itself through two names.

### Known Open Items (verification_emphasis point 3) — judged, not silently resolved or treated as automatic failure

1. **`corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case** (deferred-items.md, plan 10-07 finding). Confirmed directly by reading the function (`corevalidate.go:2407-2447`): the switch only handles `OpBorrowShared`/`OpBorrowExclusive`/`OpMove`/`OpCopy`, no `OpCall` arm. **Judgment: does not undermine the phase goal.** This causes `corevalidate` to *over-refuse* (fail closed) a class of borrow-forwarding programs that `check` would admit — a conservative-direction gap, not a permissive one. It narrows what the compiler currently accepts; it does not let an unsound program through any peer. Correctly filed as an open, unassigned item (candidate: Phase 11 or M003), not silently dropped.

2. **D-10-27 reversal of D-09-53.** Confirmed directly: `deriveFunctionUsesParam`'s doc comment (`check.go:692-705`) and `twin_b_accept.lang`'s header both state the narrowed, re-executed finding accurately — the OpReturn exemption covers only a zero-hop direct return, and applying D-09-53's originally-proposed narrowing was verified (by this phase's own planner, and the reasoning is internally consistent and falsifiable) to break 8 currently-passing tests. Zero logic change landed alongside this correction (confirmed: only comment/fixture-header diffs in the relevant commit). This reversal is procedurally sound and does not weaken the oracle.

3. **`interp.Run`'s `!function.HasClosedBody()` guard is provably unreachable.** Confirmed: `Run` (`interp.go:152-164`) calls `corevalidate.Validate` first, which already refuses an invalid body union as `core.invalid_body` before the guard's own check is ever reached. `TestRunRefusesInvalidBodyUnion` PASS, documented in 10-VALIDATION.md as defense-in-depth, not a live path. Judgment: honestly disclosed, harmless (fail-closed redundancy, not a false sense of coverage).

4. **10-REVIEW.md WR-01** (`peerCalleeFrameDrained`'s single-forward-pass ordering assumption) — see truth row 3c above. Judgment: a real, documented residual risk on a fail-closed safety check, currently correct under an invariant enforced elsewhere but not asserted locally. Not a proven defect today; worth tracking if block-lowering ever changes operation ordering. Recorded as a warning, not a blocker, because it is disclosed with a concrete reproduction path rather than glossed over.

5. **10-REVIEW.md WR-02** (`callgraph` import asymmetry) — see Independence Mechanism Audit above. Judgment: accepted, review-defensible debt, accurately re-confirmed against current source by this verification.

None of these five items were found to be silently resolved, silently dropped, or overstated as fixed. All five are consistent between the debt register, the code review, and direct source inspection performed independently by this verification.

### Anti-Patterns Found

No `TBD`/`FIXME`/`XXX` debt markers found in any `.go` file changed this phase (`git diff` name-only list checked via `grep`). No stub `return null`/empty-body patterns found in the reviewed re-deriver logic — all four `OpCall`-handling arms inspected contain substantive, differentiated logic.

### Behavioral Spot-Checks / Targeted Test Runs

| Behavior | Command | Result | Status |
|---|---|---|---|
| Origin walk OpCall hop + load-bearing gate | `go test ./internal/compiler/originvalidate/... -run 'TestOpCallOriginWalk\|Imports.*Independent'` | 3/3 PASS | ✓ |
| Pathoracle per-path discrimination | `go test ./internal/compiler/pathoracle/... -run TestCompositionDiscriminatesPerPathBorrow` | PASS | ✓ |
| Move-as-copy mutant, depth exceeded, frame drain, native-stack probe, OWN-05b guard | `go test ./internal/compiler/interp/...` (named tests) | 5/5 groups PASS | ✓ |
| Cross-function endpoint differential, composition depth bidirectional gate | `go test ./internal/compiler/session/...` (named tests) | 3/3 PASS, direction-2 shortfall message confirmed | ✓ |
| Parameter-mode mutant pairing | `go test ./internal/compiler/corevalidate/... -run TestDerivePeerSignatureModeMutantPairing` | PASS | ✓ |
| Interp oracle golden corpus + determinism | `go test ./internal/compiler/interp/... -run 'TestInterpOracleGoldenCorpus\|TestInterpOracleCorpusDeterministic\|TestInterpDeterministicAcrossRuns'` | PASS | ✓ |
| Debt-ID accountability of divergence register | `go test ./internal/compiler/session/... -run TestPeerDivergenceRegisterDebtIDsAreAccountable` | PASS | ✓ |
| Import independence (`go list -deps`) | `go list -deps ./internal/compiler/originvalidate/...` grepped for `check`/`corevalidate` | 0 hits | ✓ |
| `interp` reads no ownership-bearing `corevalidate.Result` field | `grep` + `TestInterpDoesNotReadCorevalidateOwnershipFields` | 0 hits, guard PASS | ✓ |

### Human Verification Required

None. All must-haves resolved to VERIFIED with direct evidence; the two review warnings (WR-01, WR-02) are already-adjudicated, disclosed debt with concrete reproduction paths and named landing conditions, not open questions requiring a human decision to close this phase.

### Gaps Summary

No gaps found. All four ROADMAP.md success criteria and all six requirement IDs (SEM-08, SEM-09, TRU-02, TRU-03, QLT-04, OWN-05b) are backed by passing, independently-confirmed tests whose seeded faults genuinely falsify the claims they exist to prove — not tautological self-comparisons. The independence mechanism (forbidden-import guards) is enforced by `go list -deps` transitive checking, confirmed directly against source, with one honestly-documented, review-accepted asymmetry (`callgraph`) that does not weaken the actual guarantee (transitive dependence on `check`/`corevalidate` themselves). Known open items (missing `OpCall` case in `corevalidate`'s origin peer, the D-09-53 reversal, the unreachable `HasClosedBody` guard, and the two code-review warnings) are all judged not to undermine the phase's trustworthiness goal: they are either conservative-direction gaps (fail closed, never fail open), procedurally sound reversals backed by re-executed evidence, or disclosed residual risks with a stated reproduction path rather than a silent assumption. The phase goal — that `interp`, `originvalidate`, and `pathoracle` are trustworthy enough to be the authority Phase 11's native codegen is differential-tested against — is achieved on the evidence available today.

---

_Verified: 2026-09-11T22:28:00Z_
_Verifier: Claude (gsd-verifier)_

---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 03
subsystem: compiler-validation
tags: [go, ownership, origin, corevalidate, originvalidate, publication-safety, differential-testing, fault-injection]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "09-01's peer interprocedural loan-carry derivation (derivePeerLoanCarry/chainPeerLoanCarry), landed in the same package this plan extends"
  - phase: 07-calls-signatures-and-call-graph-refusal
    provides: "peerReturnDerivesFromBorrow/peerParameterEscapesOwned's forward set-propagation shape, peerCallable's D-07-33 narrowing, TestPeerDoesNotRederiveNarrowedClasses"
provides:
  - "peerDeriveOriginFacts: peerReturnDerivesFromBorrow's access-mode-carrying generalization -- the SAME single forward pass over function.Linear.Operations, now carrying shared/exclusive per place, guarded first-hop-wins per place"
  - "peerOriginContained: a CONTAINMENT check of the declared PublicOrigin against the peer's own forward-derived facts -- never a recomputation of originvalidate's backward combination law"
  - "peerForeignOriginOmitted: a bounded, function-local re-derivation of the fourth class, scoped exactly as narrowly as originvalidate.checkForeignOriginOmitted's own precedent"
  - "peerCallable now consults all four PublishProblemsFor classes; TestPeerRederivesFormerlyNarrowedClasses (flipped from TestPeerDoesNotRederiveNarrowedClasses) asserts real re-derivation on all three formerly-vacuous classes"
  - "TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn proves the peer's forward derivation equals RecomputeOrigin's answer across the real corpus, rather than assuming it"
affects: [09-06, 09-07, 09-08, 09-09]

actuals:
  tokens: 10385
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Access-mode payload on an EXISTING forward set-propagation walk (widening a `map[string]bool` presence map to a `map[string]string` mode map, guarded 'if not already set') as the standard shape for extending a peer's boolean fact into a richer one without introducing a second walk"
    - "Containment check (declared fact vs. peer's own derived fact) as the peer-independence-preserving alternative to a producer's backward recomputation, when the producer is itself already a third validation layer"
    - "Equivalence-by-construction proof: declare EXACTLY the producer's own recomputed answer and assert the peer's independent check accepts it, rather than exposing the peer's internal derivation to an external test package"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_peer_origin_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go

key-decisions:
  - "The plan's own <behavior> prose for Task 1's Test 2 (\"exclusive borrow then shared borrow of the already-derived place derives exclusive\") was inverted relative to originvalidate.RecomputeOriginPerReturn's own documented and mechanical 'first-seen-hop-wins' rule (the walk is BACKWARD from the return, so the hop CLOSEST to the return wins, not the farthest). Verified directly against testdata/phase3/public_view_mixed_access.lang -- a real, committed, checked-clean fixture with exactly this two-hop shape, whose own doc comment confirms the correct answer is 'shared' (the closer hop), not 'exclusive'. Implemented the mechanically-correct, producer-equivalent behavior (matching RecomputeOrigin, proven generally by TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn) and wrote Task 1's own test to assert that behavior, documenting the plan-prose discrepancy here per the plan's own escape hatch ('report it, do not adjust the peer to match by construction')."
  - "TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn scopes its corpus walk to corpusFixtures() (the package's own established phase1-4/phase07 corpus) and additionally excludes any function whose body crosses an OpCall (D-09-51): originvalidate.walkReturnOrigin has no OpCall case and walks transparently through a call boundary (a pre-existing, previously-documented defect), so RecomputeOrigin is not a reliable oracle for a cross-function-derived declaration. Verified concretely: testdata/phase08/negative_control_infallible.lang's `relay` derives its origin through a call to `leaf`, and RecomputeOrigin (walking through the call transparently) disagrees with the peer's own intraprocedural-only derivation for exactly that function -- confirming the exclusion is load-bearing, not precautionary. This is the plan's own D-09-20 function-local boundary applied to the equivalence proof itself, not a narrowing of OWN-08's scope (the excluded functions never declare Callable via this path in the shipped corpus)."
  - "D-09-21's OpForeignCall fallback did NOT fire: the foreign-origin-omitted class was fully bounded, exactly as plan-time analysis (D-09-20) predicted. peerForeignOriginOmitted's own forward walk, seeded from the OpForeignCall target and propagated through Move/Copy/Borrow hops, needed no multi-hop foreign-chain reasoning."
  - "OWN-08 is NOT marked complete in REQUIREMENTS.md (per this plan's own requirement_marking_rule): plan 09-08's mid-phase gate also carries it, so the checkbox stays Pending until that gate runs. `roadmap update-plan-progress` was still run to record this plan's own completion."
  - "Rule 1 fix, out of this plan's originally-listed files: corevalidate_closure_chain_mutation_test.go's callBasicVariant helper fabricated a PublicOrigin on `identity` (an owned passthrough whose body never borrows anything) purely to make ClosureDigest react to a signature change. Phase 09's new containment check correctly refuses this dishonest declaration (Callable == false), which then made `main`'s call to `identity` fail corevalidate's own callee-Callable gate, breaking TestPeerClosureDigestEmptyCalleesMutationKilled AND (via a t.Fatalf that skipped an un-deferred restore(), leaking corevalidate.SetClosureDigestEmptyCalleesForTest(true) into the next test) TestPeerClosureDigestDiscoveryOrderMutationKilled and four foreign-closure tests. Fixed by mutating `identity`'s ForeignContract instead of PublicOrigin -- a signature-affecting change fully decoupled from origin correctness, since no OpForeignCall operation exists in identity's body to trigger any of corevalidate's ForeignContract field-shape checks."

requirements-completed: []

coverage:
  - id: D1
    description: "peerDeriveOriginFacts widens peerReturnDerivesFromBorrow's existing forward walk with an access-mode payload (shared/exclusive), first-hop-wins guarded per place, structurally opposite originvalidate's backward per-return walk plus combination law"
    requirement: "OWN-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_origin_test.go#TestPeerCallableContainmentMatchesPublishProblemsFor"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_origin_test.go#TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn"
        status: pass
    human_judgment: false
  - id: D2
    description: "peerOriginContained performs a containment check of the declared PublicOrigin against the peer's own derived facts, closing core.origin_understated and core.origin_access_mismatch"
    requirement: "OWN-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_origin_test.go#TestPeerCallableContainmentMatchesPublishProblemsFor"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestPeerRederivesFormerlyNarrowedClasses"
        status: pass
    human_judgment: false
  - id: D3
    description: "peerForeignOriginOmitted independently re-derives the fourth class (foreign-origin-omitted) with a bounded, function-local walk, and does not over-refuse a non-borrow/retain alias"
    requirement: "OWN-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_origin_test.go#TestPeerForeignOriginOmittedIsRederived"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_origin_test.go#TestPeerForeignOriginOmittedDoesNotOverRefuse"
        status: pass
    human_judgment: false
  - id: D4
    description: "Each newly-re-derived class has a mutation-kill proving it can fail: disabling the containment/foreign seams reintroduces the false agreement, restoring them re-refuses"
    requirement: "OWN-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_origin_test.go#TestPeerOriginContainmentDisabledFalselyAgreesAgain"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-03-02's interprocedural half is refused in both admission layers, each with its own code, on testdata/phase07/relay_escort_witness.lang -- verified via the package tests and the shipped CLI"
    requirement: "OWN-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 03: Peer Full Re-Derivation and D-03-02 Closure Summary

**`corevalidate`'s `Callable` peer now independently re-derives all four `PublishProblemsFor` refusal classes via an access-mode payload on its existing forward walk and a bounded foreign-origin check, closing D-07-33 without ever recomputing `originvalidate`'s backward combination law.**

## Performance

- **Duration:** 55 min
- **Tasks:** 3 (landed as 2 commits — see Decisions)
- **Files created:** 1
- **Files modified:** 4

## Accomplishments

- `peerOriginFact`/`peerDeriveOriginFacts` (`corevalidate.go`): the SAME single forward pass `peerReturnDerivesFromBorrow` already performed, now carrying `shared`/`exclusive` per place instead of presence-only, with every write guarded "if not already set" so first-hop-wins holds structurally.
- `peerOriginContained`: a containment check of the declared `PublicOrigin` against the peer's own derived facts — domain check, path coverage, access agreement — never a recomputation of `originvalidate.RecomputeOrigin`/`RecomputeOriginPerReturn`'s backward combination law.
- `peerForeignOriginOmitted`: a bounded, function-local forward walk seeded from the function's own `OpForeignCall` target, mirroring `originvalidate.checkForeignOriginOmitted`'s narrowness in spirit, never in code. D-09-21's fallback did not fire — the class was fully bounded as predicted.
- `peerCallable` now consults all four classes (foreign-origin-omitted first, matching `PublishProblemsFor`'s own early return, then origin-omitted, then containment for the declared-origin branch), replacing the unconditional `return true`.
- `TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn` proves equivalence to `RecomputeOrigin` across the real corpus by declaring exactly the recomputed answer and observing the peer accept it — not by exposing the peer's internal derivation to the external test package.
- `TestPeerRederivesFormerlyNarrowedClasses` (renamed and flipped from `TestPeerDoesNotRederiveNarrowedClasses`): all three previously-vacuous mutation subtests now assert real refusal.
- Two new QLT-08 fault-injection seams (`disablePeerOriginContainmentForTest`, `disablePeerForeignOriginPeerForTest`) with `TestPeerOriginContainmentDisabledFalselyAgreesAgain` proving both are load-bearing.
- D-03-02's interprocedural half re-confirmed closed in both admission layers: `check` refuses `testdata/phase07/relay_escort_witness.lang` with `check.interprocedural_loan_liveness`, `corevalidate` independently refuses it with `core.move_while_borrowed` (both via package tests and the shipped `lang check` CLI).

## Task Commits

Each numbered task in the plan maps to these commits (Tasks 1+2 landed together — see Decisions Made):

1. **Task 1 + Task 2: access-mode payload, containment check, bounded foreign class** - `68dad8d` (feat)
2. **Task 3: flip the narrowed-classes test to assert real re-derivation** - `82fb37a` (test)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` — `peerOriginFact`, `peerDeriveOriginFacts`, `peerOriginContained`, `peerForeignOriginOmitted`, `disablePeerOriginContainmentForTest`, `disablePeerForeignOriginPeerForTest`, `peerCallable` rewired to consult all four classes, `peerReturnDerivesFromBorrow` now a thin wrapper
- `internal/compiler/corevalidate/corevalidate_peer_origin_test.go` — new file (external test package `corevalidate_test`): `TestPeerCallableContainmentMatchesPublishProblemsFor`, `TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn`, `TestPeerForeignOriginOmittedIsRederived`, `TestPeerForeignOriginOmittedDoesNotOverRefuse`, `TestPeerOriginContainmentDisabledFalselyAgreesAgain`
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go` — `TestPeerDoesNotRederiveNarrowedClasses` renamed to `TestPeerRederivesFormerlyNarrowedClasses`, all three assertions inverted
- `internal/compiler/corevalidate/export_test.go` — `SetDisablePeerOriginContainmentForTest`, `SetDisableForeignOriginPeerForTest`
- `internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go` — `callBasicVariant` mutates `ForeignContract` instead of a dishonest `PublicOrigin` (Rule 1 fix, see Deviations)

## Decisions Made

See `key-decisions` in frontmatter for the full account of:
- The plan-prose vs. mechanical-equivalence discrepancy in Task 1's Test 2, resolved in favor of the mechanically-correct, producer-equivalent behavior.
- `TestPeerFirstHopWinsMatchesRecomputeOriginPerReturn`'s corpus scoping (phase1-4/phase07, no `OpCall`-crossing functions) and the concrete D-09-51-caused disagreement this exclusion prevents.
- D-09-21's fallback not firing.
- OWN-08 staying `Pending` in REQUIREMENTS.md per this plan's own marking rule.
- The Rule 1 fix to `callBasicVariant`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `callBasicVariant`'s fabricated `PublicOrigin` broke four unrelated tests once the containment check landed**
- **Found during:** Task 1, running the full `corevalidate` package suite after wiring `peerOriginContained` into `peerCallable`
- **Issue:** `corevalidate_closure_chain_mutation_test.go`'s `callBasicVariant(t, true)` attached a `PublicOrigin{Access: "shared"}` to `identity` (an owned passthrough function whose body never borrows anything) purely to make `identity`'s peer signature — and therefore its `ClosureDigest` — differ from the unmutated variant. Phase 09's new containment check correctly refuses this as dishonest (`Callable == false`), which then made `main`'s call to `identity` fail corevalidate's own callee-Callable gate (`core.callee_not_callable`), so `peerClosureDigestFor`'s `t.Fatalf("expected the program to validate...")` fired. Because that `t.Fatalf` ran BEFORE the test's own `restore()` call (not deferred), `corevalidate.SetClosureDigestEmptyCalleesForTest(true)`'s seam leaked into every subsequent test in the package (`TestPeerClosureDigestDiscoveryOrderMutationKilled` and four `corevalidate_foreign_closure_test.go` tests), all of which failed for the same root cause.
- **Fix:** Changed `callBasicVariant` to attach a synthetic `ForeignContract` to `identity` instead of a `PublicOrigin` — a signature-affecting, `ClosureDigest`-relevant change with zero coupling to origin correctness, since `identity`'s body contains no `OpForeignCall` operation to trigger any of corevalidate's ForeignContract field-shape checks.
- **Files modified:** `internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go`
- **Verification:** `go test ./internal/compiler/corevalidate/...` fully green (all five previously-failing tests pass)
- **Committed in:** `68dad8d`

**Total deviations:** 1 auto-fixed (Rule 1), plus the documented plan-prose correction (Test 2's expected access mode) recorded above as a key decision rather than a "fix" (the plan's own prose was the thing needing correction, not the implementation).

## Known Stubs

None.

## Threat Flags

None — this plan's threat register items (T-09-01, T-09-10, T-09-02, T-09-11) are all directly mitigated by the shipped containment check, foreign-origin check, and their mutation-kill tests.

## Issues Encountered

See Deviations above. The `callBasicVariant` breakage was diagnosed by isolating each failing test (running in isolation passed; running the full package failed), confirming a cross-test seam leak rather than five independent new bugs.

## User Setup Required

None.

## Next Phase Readiness

- OWN-08's supporting mechanism (the peer's full four-class re-derivation) is complete and load-bearing (mutation-killed). The requirement itself stays `Pending` in REQUIREMENTS.md until plan 09-08's mid-phase gate, per this plan's own carried-requirement rule.
- D-03-02 is closed in both admission layers for the interprocedural half (re-confirmed, not newly closed by this plan — that closure landed in 07-05/08-02).
- No blockers for 09-06 through 09-09.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

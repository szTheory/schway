---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 01
subsystem: compiler-validation
tags: [go, ownership, loan-liveness, corevalidate, check, interprocedural, differential-testing, fault-injection]

requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    provides: "check.interprocedural_loan_liveness law, buildInterproceduralSummaries/deriveFunctionUsesParam, the two peerDivergenceExpected accept-twin fixtures this plan retires"
  - phase: 07-calls-signatures-and-call-graph-refusal
    provides: "v.peerPostorder/v.peerSignatures substrate (checkCallGraphAcyclic), chainPeerClosureDigests' postorder consumption idiom, peerReturnDerivesFromBorrow/peerParameterEscapesOwned's forward set-propagation shape, verifyCallableRefusalSeam + its companion-assertion test shape"
provides:
  - "corevalidate's own interprocedural loan-liveness derivation (derivePeerLoanCarry, chainPeerLoanCarry), a forward set-propagation walk folded into the existing peer postorder substrate, structurally opposite check's backward memoized walk"
  - "buildLoanChainIndex's OpCall consult, breaking the loan chain at a call boundary when the callee does not declare a borrow-of-parameter return"
  - "Both testdata/phase08/twin_a_accept.lang and testdata/phase08/relay_depth2_accept.lang retired from peerDivergenceExpected -- D-08-40 resolved"
  - "Bidirectional seeded-fault companion assertions (D-09-24/D-09-25) proving the two peers' agreement is load-bearing, not vacuous"
  - "PHASE-09-DEBT.md D-09-51: a newly-discovered, pre-existing originvalidate defect this plan's fix unmasks (documented, not fixed -- out of scope)"
affects: [09-02, 09-03, 09-06, 09-07, 09-08, 09-09]

actuals:
  tokens: 11106
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Forward set-propagation over function.Linear.Operations, folded into an existing postorder substrate, as the standard shape for a NEW corevalidate peer fact (never a backward walk, never a callgraph import)"
    - "Cross-package fault-injection seam exported from a PRODUCTION file (not export_test.go) when a same-package-external test in a DIFFERENT package needs to engage it -- the SetDisableCyclePeerForTest exception, now with two more instances"
    - "Bidirectional companion-assertion seam pairs (X-disabled/Y-still-refuses in both directions) as the discriminating evidence for peer independence, per Knight & Leveson (1986)"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_peer_liveness.go
    - internal/compiler/corevalidate/corevalidate_peer_liveness_test.go
    - internal/compiler/check/check_peer_liveness_seam_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/corevalidate/corevalidate_exclusive_test.go
    - internal/compiler/check/check.go
    - internal/compiler/session/session_peer_gate_test.go
    - .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md

key-decisions:
  - "Tasks 1 and 2 landed in a single commit: the multi-hop transitivity claim (Task 2) is proven by the exact same single forward pass Task 1 authors, with zero additional production code -- splitting the commit would separate a mechanism from half of its own falsifying tests."
  - "The two new cross-package fault-injection seams (SetDisablePeerLoanCarryConsultForTest, SetForcePeerLoanCarryTrueForTest) were moved from export_test.go (as the plan specified) into the production file corevalidate_peer_liveness.go, mirroring the existing SetDisableCyclePeerForTest exception (D-07-42): export_test.go symbols are invisible outside a package's own test binary, and check's own companion-assertion test needed to reach them from a different package."
  - "A new check.go seam (disableInterproceduralLoanLivenessForTest) was added -- not listed in the plan's files_modified -- because no existing seam could make check.Program admit a program its interprocedural liveness law would otherwise refuse; without it, the existing-direction companion assertion (D-09-24) could not be written at all."
  - "Discovered and documented (PHASE-09-DEBT.md D-09-51), rather than fixed, a pre-existing originvalidate defect: walkReturnOrigin has no OpCall case and walks transparently through a call ignoring the callee's own declared return contract, causing a false-positive core.origin_omitted refusal on both retired fixtures when run through the FULL CLI (session.CheckCommandFile, which calls originvalidate.ValidatePublished after corevalidate.Validate). This was never observable before because corevalidate's own (now-fixed) bug always refused first. Fixing it requires threading whole-program callee-lookup access through RecomputeOriginPerReturn/walkReturnOrigin's exported signatures -- a cross-cutting API change to a file not in this plan's scope, judged Rule 4 (architectural) rather than a bounded inline fix."

requirements-completed: [OWN-07, TRU-04]

coverage:
  - id: D1
    description: "corevalidate derives its own interprocedural loan-carry fact by forward set-propagation over function.Linear.Operations, consuming callee facts through its own v.peerPostorder byproduct -- never callgraph.Order, never an import of check"
    requirement: "OWN-07"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_liveness_test.go#TestPeerLoanCarryDerivesForwardFromOperations"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_liveness_test.go#TestPeerLivenessFileImportsStayIndependent"
        status: pass
    human_judgment: false
  - id: D2
    description: "buildLoanChainIndex consults the peer-derived fact before propagating a loan across a core.OpCall, breaking the caller's loan chain at the call boundary when the callee does not return a borrow of its parameter"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_liveness_test.go#TestPeerLoanCarryPropagatesAcrossTwoCallHops"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_peer_gate_test.go#TestNoUndeclaredCheckPeerDivergenceAcrossCorpus"
        status: pass
    human_judgment: false
  - id: D3
    description: "Both peerDivergenceExpected accept-twin fixtures (twin_a_accept.lang, relay_depth2_accept.lang) are retired, and the exact-set gate proves the retirement in both directions -- resolving D-08-40"
    requirement: "TRU-04"
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_peer_gate_test.go#TestNoUndeclaredCheckPeerDivergenceAcrossCorpus"
        status: pass
    human_judgment: false
  - id: D4
    description: "A seeded fault in one peer makes the two diverge while the other peer's independently-derived answer is unchanged, asserted in both directions with the discrimination argument in the test's own doc comment"
    requirement: "OWN-07"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_peer_liveness_seam_test.go#TestPeerLoanCarrySeamDisabledCheckStillRefuses"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_peer_liveness_seam_test.go#TestInterproceduralLivenessSeamCheckDisabledCorevalidateStillRefuses"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_liveness_test.go#TestPeerLoanCarryForcedTrueReintroducesRetiredDivergence"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_liveness_test.go#TestPeerLoanCarryConsultDisabledReintroducesRetiredDivergence"
        status: pass
    human_judgment: false
  - id: D5
    description: "The full CLI (`lang check`) reports a clean check for both retired fixtures end-to-end"
    verification: []
    human_judgment: true
    rationale: "NOT satisfied as literally written -- see Deviations/D-09-51. The full CLI now reports core.origin_omitted (a different, pre-existing, previously-masked originvalidate defect) instead of core.move_while_borrowed for both fixtures. corevalidate's own divergence is genuinely closed; a human/planning decision is needed on whether and when to fix originvalidate's OpCall-transparent origin walk."

duration: 33min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 01: Peer Interprocedural Loan-Carry Derivation Summary

**`corevalidate` now derives its own interprocedural loan-liveness fact by a forward set-propagation walk over its existing postorder substrate, structurally opposite `check`'s backward memoized walk, closing both `peerDivergenceExpected` accept-twin divergences and proving the agreement load-bearing with bidirectional seeded-fault companion assertions.**

## Performance

- **Duration:** 33 min
- **Tasks:** 3 (landed as 2 commits — see Decisions)
- **Files created:** 3
- **Files modified:** 6 (including one debt-register doc)

## Accomplishments

- `derivePeerLoanCarry`/`chainPeerLoanCarry` (`corevalidate_peer_liveness.go`): a single forward pass over `function.Linear.Operations`, folded into the existing `v.peerPostorder` callee-before-caller substrate, imports only `core`.
- `buildLoanChainIndex`'s new `OpCall` consult: a callee that does not declare a borrow-of-parameter return breaks the caller's loan chain at the call boundary.
- Red-first differential (D-09-27): captured the genuine `UNDECLARED divergence: testdata/phase08/twin_a_accept.lang` failure against the known-divergent state, then landed the mechanism and confirmed green — see RED Output below.
- Both `testdata/phase08/twin_a_accept.lang` and `testdata/phase08/relay_depth2_accept.lang` retired from `peerDivergenceExpected`; `relay_depth2_refuse.lang` confirmed to still refuse (the consult did not break every chain).
- Bidirectional seeded-fault companion assertions land in both directions (new: corevalidate-side seam, check asserted unaffected and still-disagreeing/agreeing correctly; existing: check-side seam generalized from Callable to liveness), each with the Knight & Leveson (1986) discrimination argument stated in its own doc comment.
- Two QLT-08 mutation-kills prove both new fault-injection seams are independently load-bearing.
- Import guards (`corevalidate_test.go`, `corevalidate_exclusive_test.go`) extended to also scan the new sibling production file.
- Discovered and fully documented (not fixed) a pre-existing, previously-masked `originvalidate` defect (PHASE-09-DEBT.md D-09-51).

## RED Output (captured before the mechanism landed, D-09-27)

```
=== RUN   TestNoUndeclaredCheckPeerDivergenceAcrossCorpus
    session_peer_gate_test.go:267: UNDECLARED divergence: testdata/phase08/twin_a_accept.lang (check admits, corevalidate refuses with core.move_while_borrowed) is not in peerDivergenceExpected
--- FAIL: TestNoUndeclaredCheckPeerDivergenceAcrossCorpus (0.07s)
FAIL
```

Captured by: reverting to the pre-plan tree (`git stash -u`), removing the `twin_a_accept.lang` entry from `peerDivergenceExpected`, running `go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v`, then restoring. GREEN was then confirmed with the real mechanism landed and both entries removed (`--- PASS`).

## Task Commits

Each numbered task in the plan maps to these commits (Tasks 1+2 landed together — see Decisions Made):

1. **Task 1 + Task 2: peer loan-carry derivation, OpCall consult, both fixture retirements, multi-hop transitivity** - `703fedc` (feat)
2. **Task 3: seeded fault in both directions with companion assertions** - `ed7ba91` (test)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate_peer_liveness.go` - `peerLoanCarryFact`, `derivePeerLoanCarry` (forward set-propagation), `chainPeerLoanCarry` (postorder fold-in), the two fault-injection seams and their cross-package setters
- `internal/compiler/corevalidate/corevalidate_peer_liveness_test.go` - Tasks 1-3's own tests (7 total): derivation shape, memoization ordering, corrupted-artifact termination, two-hop transitivity + its ordering falsifier, import independence, two mutation-kills
- `internal/compiler/corevalidate/corevalidate.go` - `validator.peerLoanCarry` field, `chainPeerLoanCarry()` wired into `run()` immediately after `checkCallGraphAcyclic` (NOT at `chainPeerClosureDigests`'s later call site — see Deviations), `buildLoanChainIndex`'s new third parameter and `OpCall` branch, all three call sites updated
- `internal/compiler/corevalidate/corevalidate_test.go` / `corevalidate_exclusive_test.go` - import guards 3/4 extended to scan `corevalidate_peer_liveness.go`
- `internal/compiler/check/check.go` - `disableInterproceduralLoanLivenessForTest` seam (new, not in plan's `files_modified` — see Deviations)
- `internal/compiler/check/check_peer_liveness_seam_test.go` - both companion-assertion tests
- `internal/compiler/session/session_peer_gate_test.go` - both fixture entries retired from `peerDivergenceExpected`
- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md` - D-09-51 appended (execution-time finding)

## Decisions Made

- **Tasks 1+2 combined into one commit.** Task 2's multi-hop transitivity claim requires zero additional production code beyond Task 1's single forward pass (composition falls out of the existing callee-before-caller postorder for free) — splitting the commit would separate the mechanism from half its own falsifying tests with no meaningful intermediate state.
- **Fault-injection setters moved from `export_test.go` to the production file.** The plan specified `export_test.go` for `SetDisablePeerLoanCarryConsultForTest`/`SetForcePeerLoanCarryTrueForTest`, but `export_test.go` symbols are invisible outside `corevalidate`'s own test binary (Go excludes `_test.go` files from normal imports) — and `check`'s own companion-assertion test needs to call them from a different package. Followed the exact precedent `SetDisableCyclePeerForTest` already established (D-07-42) instead: a documented, test-only-no-op, production-visible exception.
- **Added a new `check.go` seam not in the plan's file list.** No existing seam could make `check.Program` admit a program its interprocedural liveness law would otherwise refuse (needed for D-09-24's existing-direction companion assertion). Added `disableInterproceduralLoanLivenessForTest`, mirroring `verifyCallableRefusalSeam`'s exact shape, generalized to the liveness fact.
- **`chainPeerLoanCarry()` is called immediately after `checkCallGraphAcyclic`, not at `chainPeerClosureDigests`'s later call site.** The plan's action text said to call it "from the same place... where chainPeerClosureDigests is invoked" — but `buildLoanChainIndex` (which reads `v.peerLoanCarry`) is invoked from *inside* the pending-replay loop, which runs *before* `chainPeerClosureDigests`'s own call site. Calling it at the later point would read a nil map on every replay. Corrected the ordering to the actually-correct dependency point, documented in `run()`'s own comment.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fault-injection setters relocated from `export_test.go` to the production file**
- **Found during:** Task 3, writing `check_peer_liveness_seam_test.go`
- **Issue:** As specified, the setters would not compile — `export_test.go`'s symbols do not exist outside `corevalidate`'s own test binary, and `check`'s test needs cross-package access
- **Fix:** Declared `SetDisablePeerLoanCarryConsultForTest`/`SetForcePeerLoanCarryTrueForTest` in `corevalidate_peer_liveness.go` instead, following `SetDisableCyclePeerForTest`'s own documented precedent
- **Files modified:** `internal/compiler/corevalidate/corevalidate_peer_liveness.go`, `internal/compiler/corevalidate/export_test.go` (net no-op after add-then-revert)
- **Verification:** `go test ./internal/compiler/check/... -run 'SeamDisabled|StillRefuses' -v` passes
- **Committed in:** `ed7ba91`

**2. [Rule 3 - Blocking] Added `check.go`'s `disableInterproceduralLoanLivenessForTest` seam**
- **Found during:** Task 3, writing the existing-direction companion assertion
- **Issue:** No shipped seam could make `check.Program` admit a program its own interprocedural liveness pass would otherwise refuse — the existing-direction test literally could not be written without one
- **Fix:** Added a new unexported package-level bool, read at the call site immediately before `checkInterproceduralLoanLiveness` is invoked, mirroring `verifyCallableRefusalSeam`'s shape and doc-comment style
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `TestInterproceduralLivenessSeamCheckDisabledCorevalidateStillRefuses` passes; full check/corevalidate/session suite green under `-count=2 -shuffle=on` (no seam leakage)
- **Committed in:** `ed7ba91`

**3. [Rule 1 - Bug, ordering] Corrected `chainPeerLoanCarry()`'s call site in `run()`**
- **Found during:** Task 1, wiring the mechanism into `run()`
- **Issue:** The plan's literal instruction (call it where `chainPeerClosureDigests` is invoked) would read `v.peerLoanCarry` as `nil` on every replay, since `buildLoanChainIndex` consumes it from *inside* the pending-replay loop that runs *before* `chainPeerClosureDigests`'s own call site
- **Fix:** Call `v.chainPeerLoanCarry()` immediately after `checkCallGraphAcyclic` succeeds, before the pending-replay loop, with an explanatory comment
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** All `PeerLoanCarry`-scoped tests pass; the two fixture retirements are correct in both directions
- **Committed in:** `703fedc`

---

**Total deviations:** 3 auto-fixed (2 blocking, 1 ordering bug) — all necessary to make the plan's own required tests compile and pass. No scope creep beyond what Task 3 itself required.

### Major Finding, Documented Not Fixed (see PHASE-09-DEBT.md D-09-51)

**[Rule 4 - Architectural, out of scope] `originvalidate`'s OpCall-transparent origin walk unmasked**

Once `buildLoanChainIndex`'s consult lands and `corevalidate.Validate` correctly stops refusing both retired fixtures, running either through the FULL CLI (`go run ./cmd/lang --json check ...`, which additionally calls `originvalidate.ValidatePublished` after `check` and `corevalidate` both pass) surfaces a DIFFERENT, pre-existing refusal: `core.origin_omitted`.

- **Root cause (verified directly):** `originvalidate.walkReturnOrigin` (`originvalidate.go:171-218`) has no `case core.OpCall` in its backward-walk switch, so it treats a call boundary as fully transparent — walking straight through to the call's own argument's provenance, never consulting the callee's own declared return contract. For `twin_a_accept.lang`, `escortee` declares a plain owned return, so `escort`'s actual return is a genuinely fresh owned value with no alias risk, but `originvalidate` still reports it as borrow-derived.
- **Confirmed NOT introduced by this plan:** `check.Program` alone returns zero diagnostics for the fixture; `originvalidate.ValidatePublished` called independently (bypassing corevalidate entirely) reproduces the exact same `core.origin_omitted` problem regardless of this plan's changes. It was simply unreachable through the full CLI before, because `corevalidate`'s own (now-fixed) bug always refused first in `session.CheckCommandFile`'s fixed precedence.
- **Confirmed NOT to affect the actual required gate:** `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` compares `check.Program` against `corevalidate.Validate` only — it never calls `originvalidate.ValidatePublished` — so it passes cleanly and D-09-03's retirement is genuine and complete on its own terms.
- **Why not fixed here:** a real fix requires threading whole-`core.Program` (or a precomputed callee-return-mode table) access through `RecomputeOriginPerReturn`/`walkReturnOrigin`/`RecomputeOrigin`'s exported signatures — a cross-cutting API change touching `BuildInterface`, `ValidatePublished`, and every existing `originvalidate_test.go` call site that constructs these calls directly against a bare `core.Function`. `originvalidate.go` is not in this plan's `files_modified`, and the blast radius is judged Rule 4 territory (significant structural modification with broad impact on an unrelated validator's own extensive test suite), not a bounded inline fix available mid-task.
- **Consequence:** the plan's literal `<verify>`/`<acceptance_criteria>` CLI-clean-check claims for both retired fixtures are NOT satisfied end-to-end. Recorded explicitly (PHASE-09-DEBT.md D-09-51, and coverage item D5 above) rather than silently declared done. No landing phase is yet assigned; needs a scoped follow-up plan or a Phase 09 mid-phase gate (09-07/09-08) agenda item.

## Known Stubs

None.

## Threat Flags

None — this plan's threat register items (T-09-01, T-09-02, T-09-05, T-09-06) are all directly mitigated by the shipped mechanism and tests; D-09-51 is a finding in a DIFFERENT validator's pre-existing code, not new surface this plan introduced.

## Issues Encountered

See Deviations above — all three auto-fixes were discovered and resolved inline during Task 3. The D-09-51 finding was investigated thoroughly (direct experiments against both the original and fixed tree, isolating `check.Program` from `originvalidate.ValidatePublished`) before concluding it was out of scope, rather than assumed.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- A second, independently-implemented interprocedural loan-liveness detector now exists (D-09-10's first half) — plan 09-02 (which closes the pre-existing `TestQLT01RegistryCoversAllFiveSpikes` exemption and widens TRU-04's corpus vehicle) and later plans that restructure/delete `check`'s intraprocedural law (D-09-08/D-09-09/D-09-10's second half) can now proceed.
- **Carried forward for a future plan or the mid-phase gate (09-07/09-08):** PHASE-09-DEBT.md D-09-51, the originvalidate OpCall-transparency false positive. Not blocking this plan's own success criteria, but should be triaged before the phase closes — it affects the SAME two fixtures this plan's own success criteria name.
- No blockers for 09-02 through 09-06; all of this plan's own must-have truths, artifacts, and prohibitions are satisfied (see coverage D1-D4; D5 is the one open item, explicitly a human/planning decision, not an execution blocker).

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

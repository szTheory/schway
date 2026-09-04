---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "04"
subsystem: compiler-frontend
tags: [ownership, cfg, loan-liveness, corevalidate, independent-validation, negative-control]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-01's real CFG (core.Block/Edge/LoanEndpoint, checkBranch, corevalidate.matchBranch/replayBlocks), 03-02's exclusive loans and independent conflict re-derivation, and 03-03's backward worklist loan-liveness dataflow (loanLivenessFixpoint/materializeLoanEndpoints) that first populates core.LoanEndpoint for real, scoped to checkBranch's arm blocks only"
provides:
  - "corevalidate.recomputeLoanEndpoints: a second, mechanically different CFG loan-liveness derivation (materialized use relation + one-shot reachability closure + reduction, never an iterative worklist) that independently recomputes a branch-shaped function's loan-endpoint set and rejects a moved, dropped, or invented one with core.loan_endpoint_mismatch"
  - "corevalidate.loanChainIndex: a memoized, cycle-safe, parent-pointer place-provenance chain replacing the quadratic loansForPlace copy-and-rescan in both replayStraightLine and replayBlocks (D-02-03/Q2(b), validator half)"
  - "session.BorrowedLoanEndpointControlLane: Phase 3's own loan-endpoint mutation negative control, wired with the Phase 1 addLane shape (explicit status, never silently dropped on failure)"
affects: [03-05-path-oracle, 03-07-debug-lineage-and-phase-close]

actuals:
  tokens: 12934
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Reachability closure + explicit-use-relation reduction (recomputeLoanEndpoints), not an iterative worklist fixpoint: check.go's loanLivenessFixpoint converges per-block live sets under repeated re-evaluation; this pass computes one bounded-BFS reachability closure and one static reduction over a materialized (loan, block, ordinal) relation, with no shared helper, no import, and no queue anywhere in corevalidate.go"
    - "Iterative (not recursive) parent-chain walk with cycle detection (loanChainIndex.carriedLoans): a naive recursive walk stack-overflowed on a corrupted/adversarial self-referencing parent pointer (a mutation test deliberately retargets a borrow's target onto its own source); the fix walks the chain with an explicit visited-set and folds memoization bottom-up, staying defined on cyclic input instead of trusting the producer's honesty"
    - "Phase 1 addLane shape (explicit status parameter) adopted for the new Phase 3 control lane, not Phase 2's shape (hardcoded pass, drops the lane entirely on any failure path -- PATTERNS I-1)"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go
    - internal/compiler/corevalidate/corevalidate_endpoint_test.go
    - internal/compiler/corevalidate/corevalidate_quadratic_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_branch_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "LinearWorkLimit moved from 16*facts+13 to 16*facts+14: the +1 is loanChainIndex's own honest, once-per-place counted work. The canonical scale shape (scaleProgram) has every operation read from the SAME single parameter place, so carriedLoans visits exactly one NEW place total regardless of facts -- a facts-independent constant, not an unaccounted bump. Verified empirically against real go test runs at facts=101/1001/10001 before being written into the formula, then re-derived in the formula's own doc comment."
  - "recomputeLoanEndpoints is wired into v.linear() itself (gated on len(Blocks)>0), not into matchBranch specifically, so it runs uniformly for any branch-shaped function regardless of which run() dispatch case reached it -- and so a synthetic Match-less Linear-with-Blocks function (used by the internal edge-divergence test) can exercise it directly without needing to satisfy matchBranch's own match-shape invariants."
  - "TestValidatorRecomputesLoanEndpoints and its multi-successor edge-divergence fixture live in an INTERNAL test file (package corevalidate, not corevalidate_test), calling recomputeLoanEndpoints directly -- exactly mirroring check.go's own check_test.go precedent for loanLivenessFixpoint/materializeLoanEndpoints. The full Validate() pipeline's replayBlocks enforces 'exactly one Return per non-empty block', which the general multi-successor divergence shape (mirroring check.go's TestEdgeSpecificLiveOut, where the loan's birth block itself has two successors and no Return) does not satisfy and does not need to: it exists to prove the recomputation MECHANISM, not a shape checkBranch emits today."
  - "BorrowedLoanEndpointControlLane mutates by DROPPING an endpoint (not moving or inventing one) as the 'strongest' variant wired into the control lane: an omission has no other structural invariant available to catch it, unlike a moved endpoint (which could plausibly collide with another referential check) or an invented one (which adds a slice entry rather than removing information)."
  - "Renamed 03-01's TestLoanEndpointMutationMatrix (which never touched a LoanEndpoint -- an OpReturn-ordering falsifier) to TestBranchReturnMustBeLastInBlock, freeing the name for this plan's real endpoint mutation matrix. See Deviations."

requirements-completed: [OWN-03]

coverage:
  - id: D1
    description: "A second, independently written CFG loan-liveness derivation (recomputeLoanEndpoints) decides loan endpoints by reachability closure and reduction, never an iterative worklist; the validator imports neither check nor ast."
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/corevalidate#TestValidatorRecomputesLoanEndpoints", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestValidatorImportsStayIndependent", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestLoanEndpointMismatchRejected", status: pass}
    human_judgment: false
  - id: D2
    description: "Both admission layers are linear in body size (the validator's own quadratic loansForPlace scan is removed), with pre-existing accept/reject verdicts and first-problem codes provably unchanged."
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/corevalidate#TestValidatorVerdictsUnchanged", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestCoreValidationWorkSeries", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestValidatorReborrowChainIsLinear", status: pass}
      - {kind: unit, ref: "internal/compiler/corevalidate#TestTransitiveLoanBlocksMove", status: pass}
    human_judgment: false
  - id: D3
    description: "A moved, dropped, or invented loan-endpoint set is caught by the independent layer and recorded as a required negative control with nonzero work; a failing control still returns an observable lane."
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/session#TestLoanEndpointMutationMatrix", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestBorrowedLaneRecordsFailureStatus", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestVerifyPhase2ControlsAndWork", status: pass}
    human_judgment: false

duration: 95min
completed: 2026-09-04
status: complete
---

# Phase 03 Plan 04: Independent CFG Loan-Liveness Derivation and Its Negative Control Summary

**corevalidate now decides loan endpoints on its own terms — a reachability-closure-plus-reduction derivation, never the checker's iterative worklist — removes its own quadratic loansForPlace scan (D-02-03's validator half) behind a cycle-safe memoized chain, and rejects a moved, dropped, or invented endpoint set through a new Phase 1-shaped negative-control lane.**

## Performance

- **Duration:** ~95 min
- **Tasks:** 3 completed
- **Files modified:** 7 (3 created, 4 modified)
- **Commits:** 3 (one per task)

## Accomplishments

- `recomputeLoanEndpoints` independently re-derives a branch-shaped
  function's `core.LoanEndpoint` set directly from the declared
  `Blocks`/`Edges`/`Operations` alone. It materializes an explicit
  `(loanID, block, ordinal)` use relation, computes a one-shot
  reachability closure over the declared successor relation (bounded
  BFS, safe even on an adversarial cyclic graph), and reduces to each
  loan's final reachable uses — a genuinely different mechanism from
  check.go's `loanLivenessFixpoint` (which converges per-block live
  sets under repeated worklist re-evaluation). Hand-verified against
  check.go's own `TestEdgeSpecificLiveOut` scenario (a loan born at a
  two-successor divergence, used along one path and not the other) and
  proven to reproduce the exact same point/edge classification.
- `TestValidatorImportsStayIndependent` reads corevalidate's actual Go
  import list via `go/parser` and fails if it ever imports `compiler/check`
  or `compiler/ast`.
- `TestLoanEndpointMismatchRejected` (corevalidate) and
  `TestLoanEndpointMutationMatrix` (session) both independently prove
  the three-variant matrix — endpoint moved to a different block,
  dropped entirely, invented for a nonexistent loan — is rejected with
  the stable `core.loan_endpoint_mismatch` code, driven from real
  checker-produced branch programs (`session.Check`), not synthetic
  facts.
- Closed the validator half of D-02-03/Q2(b): `loanChainIndex` replaces
  the per-operation `append([]string(nil), loansForPlace[...]...)`
  copy-and-rescan in both `replayStraightLine` and `replayBlocks` with
  a memoized parent-pointer chain, walked in full at most once per
  place across the whole function. A stack-overflow was caught and
  fixed mid-session (see Deviations) making the walk iterative and
  cycle-safe rather than recursive.
- `TestValidatorVerdictsUnchanged` pins accept/reject and first-problem
  codes across the existing fixture corpus (owned, shared/exclusive
  borrows, reborrow-while-moved four ways, both branch fixtures) —
  verdict-preserving proof for the replacement.
- `TestValidatorReborrowChainIsLinear` adds a dedicated
  101/1,001/10,001-fact series built from a genuine sequential reborrow
  chain (`scaleReborrowProgram`) — the shape `TestCoreValidationWorkSeries`'s
  independent-loan series cannot reach — and asserts the growth ratio,
  not an exact formula or wall-clock time. Measured
  `Checks`/facts: 1930/19030/190030 — ratio flat at ~19.0 across the
  full 100x growth.
- `LinearWorkLimit` moves from `16*facts+13` to `16*facts+14`: the +1
  is `loanChainIndex`'s own honest counted work for the canonical scale
  shape's single shared parameter place, empirically confirmed
  (`facts=101 checks=1630` before the formula was updated) then written
  into the formula's own doc comment as a derivation, not a bare bump.
- `session.BorrowedLoanEndpointControlLane` wires the "dropped" variant
  into a real verify-lane control (`control:core.loan_endpoint_mismatch`)
  using the Phase 1 `VerifyCorpus` `addLane` shape (explicit status
  parameter), not Phase 2's `verifyOwnedCorpus` shape (hardcoded
  `"pass"`, silently drops the lane on any failure path — PATTERNS I-1).
  `TestBorrowedLaneRecordsFailureStatus` proves a failing control still
  returns a real, non-empty lane with nonzero work.

## Task Commits

1. **Task 03-04-01: A second, independently written CFG liveness derivation in the validator** — `580c803` (feat) — also carries Task 03-04-02's `loanChainIndex` replacement and `LinearWorkLimit` update, since both fixes touch the same functions in the same file; see Deviations.
2. **Task 03-04-02: Verdict pinning and reborrow-chain linearity proof** — `887adb9` (test)
3. **Task 03-04-03: Wire the endpoint mutation as a required negative control** — `2a7230b` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` — `recomputeLoanEndpoints`,
  `blockReach`, `loanChainIndex`/`carriedLoans`/`foldChain`,
  `loanEndpointsMatch`, `LinearWorkLimit` (16n+14), quadratic-scan removal
  in `replayStraightLine`/`replayBlocks`
- `internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go` —
  `TestValidatorRecomputesLoanEndpoints`, `TestValidatorImportsStayIndependent`
  (internal, package `corevalidate`)
- `internal/compiler/corevalidate/corevalidate_endpoint_test.go` —
  `TestLoanEndpointMismatchRejected`
- `internal/compiler/corevalidate/corevalidate_quadratic_test.go` —
  `TestValidatorVerdictsUnchanged`, `TestValidatorReborrowChainIsLinear`,
  `scaleReborrowProgram`
- `internal/compiler/corevalidate/corevalidate_branch_test.go` — renamed
  `TestLoanEndpointMutationMatrix` to `TestBranchReturnMustBeLastInBlock`
- `internal/compiler/session/session.go` — `BorrowedLoanEndpointControlLane`
- `internal/compiler/session/session_test.go` — `TestLoanEndpointMutationMatrix`,
  `TestBorrowedLaneRecordsFailureStatus`

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Recursive parent-chain walk stack-overflowed on an adversarial self-referencing parent pointer**
- **Found during:** Task 03-04-01/02, first full-suite run after replacing `loansForPlace`
- **Issue:** `loanChainIndex.carriedLoans`'s first implementation walked the parent chain recursively. `TestOwnershipMutationMatrix`'s existing "reused borrow target" mutation (retargets a borrow's `TargetID` onto the SAME place already used as its own `SourceID`, i.e. a self-loop) made `parent[place] == place`, and the recursive walk exhausted the goroutine stack (`fatal error: stack overflow`) rather than being rejected by the ordinary target-validation this package already performs elsewhere in the same pass.
- **Fix:** Rewrote `carriedLoans` to walk the chain ITERATIVELY with an explicit visited-set, stopping (not crashing) the moment it revisits a place already on the current walk, then folding memoization bottom-up (`foldChain`). A corrupted/cyclic core artifact is still rejected — just by the pre-existing target-ordinal check downstream, not by this helper crashing first.
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`
- **Verification:** Full `go test ./internal/compiler/corevalidate/...` (previously panicked with a stack overflow) now passes; `TestOwnershipMutationMatrix`'s "reused borrow target" case still correctly rejects with `core.invalid_target`.
- **Committed in:** `580c803` (fixed before the commit landed; the panic never reached a committed state)

---

### Documented Scope/Interpretation Deviations (not Rule 1-3 bug fixes)

**2. [Rule 4-style, documented] `recomputeLoanEndpoints` is wired into `v.linear()`, not `matchBranch`**
- **Plan text:** implied the check lives inside the match-branch validation path.
- **Decision:** Gated on `len(function.Linear.Blocks) > 0` inside the shared `v.linear()` helper instead, so it runs uniformly for any branch-shaped function regardless of dispatch case, and so the internal multi-successor edge-divergence test can construct a synthetic `Match == nil, Linear != nil` function carrying `Blocks` directly — bypassing `matchBranch`'s own match-shape scaffolding (arm IDs, pattern exhaustiveness) entirely, exactly mirroring how check.go's `TestEdgeSpecificLiveOut` calls `loanLivenessFixpoint`/`materializeLoanEndpoints` directly rather than through a full `checkBranch` round trip.
- **Impact:** No functional difference for any real checkBranch-produced program (only branch-shaped functions ever carry `Blocks` today); enables a much simpler, more direct falsifier for the general multi-successor case.

**3. [Rule 4-style, documented] Renamed 03-01's misnamed `TestLoanEndpointMutationMatrix`**
- **Found during:** Task 03-04-03, writing the plan's own required `TestLoanEndpointMutationMatrix`
- **Issue:** 03-01 shipped a test with this exact name whose doc comment claimed "T-03-06/T-03-14 shipped-early", but the test body only asserted `OpReturn` ordering within a block (`core.final_claim_mismatch`) — it never constructed, mutated, or asserted anything about a `LoanEndpoint`. Go forbids two functions of the same name in one package, and this plan's own verify command requires the exact name `TestLoanEndpointMutationMatrix` to exist and cover the real three-variant matrix.
- **Fix:** Renamed the 03-01 test to `TestBranchReturnMustBeLastInBlock` (its actual, accurately-described behavior), with a doc comment explaining the rename and pointing here. Added the real `TestLoanEndpointMutationMatrix` in `session_test.go`.
- **Files modified:** `internal/compiler/corevalidate/corevalidate_branch_test.go`, `internal/compiler/session/session_test.go`
- **Verification:** Both tests pass; full suite green.
- **Impact:** No behavior change — the renamed test's assertions and coverage are byte-identical to before, just correctly named.

---

**Total deviations:** 1 auto-fixed (Rule 1, caught and fixed before any commit landed the broken version) + 2 documented scope/interpretation decisions (Rule 4-style).
**Impact on plan:** No scope creep; both documented decisions narrow implementation mechanics in ways that make the required falsifiers easier to write correctly, without changing any load-bearing behavior.

## Issues Encountered

None beyond the deviations above, all resolved within this plan's own scope before any broken state was committed.

## User Setup Required

None — no external service configuration required.

## Inherited Open Deviation from 03-03 (required disclosure)

03-03 left `checkLinear`/`analyzeStraightLine` deliberately NOT rewired
onto the new backward-worklist dataflow. `discoverLoanLastUses` still
drives the straight-line path's conflict/expiry decisions; the backward
worklist dataflow (`loanLivenessFixpoint`) is the SOLE producer of
observable `core.LoanEndpoint` records, scoped to `checkBranch`'s arm
blocks only. Consequently:

- `corevalidate.recomputeLoanEndpoints` is genuinely mechanically
  different from `check.go`'s `loanLivenessFixpoint` (reachability
  closure + reduction vs. an iterative worklist fixpoint — see
  `key-decisions`/Accomplishments above), and it validates BOTH kinds
  of program: a straight-line function (no `Blocks`) short-circuits to
  `nil`, matching its always-empty declared `LoanEndpoints`
  (`TestValidatorRecomputesLoanEndpoints`'s "straight-line body" case;
  `TestValidatorVerdictsUnchanged` exercises several straight-line
  programs end to end through the SAME `v.linear()` path). A
  branch-shaped function (has `Blocks`) gets the full independent
  recomputation and comparison.
- **This plan's corrupted-endpoint-set tests (`TestLoanEndpointMismatchRejected`,
  `TestLoanEndpointMutationMatrix`) exercise ONLY the checker's BRANCH
  path** (`checkBranch`'s backward worklist dataflow), because
  `core.LoanEndpoint` is never populated for a straight-line program at
  all — there is no endpoint set to corrupt on that path. This plan
  does not, and structurally cannot, differentially exercise
  `discoverLoanLastUses`'s straight-line liveness law against anything,
  since that law never produces a `LoanEndpoint` fact to compare
  against.
- **No new disagreement between the checker's two liveness paths was
  found or introduced by this plan.** This plan's own production code
  never touches `checkLinear`/`analyzeStraightLine`/`discoverLoanLastUses`;
  it only reads and independently re-derives the `LoanEndpoint` records
  `checkBranch` already produces. The 03-03-flagged gap (two liveness
  derivations, not one, still coexisting in the checker) remains open
  and unrepaired — this plan neither widens nor narrows it, and 03-05
  remains the phase's designated mid-phase gate for adjudicating it.

## Mutation-Kill / Non-Regression Evidence (recorded verbatim per D-09)

- `TestLoanEndpointMismatchRejected`/`TestLoanEndpointMutationMatrix`:
  reverting `recomputeLoanEndpoints`'s comparison (i.e. removing
  `loanEndpointsMatch`'s call site from `v.linear()`) makes every
  mutation variant pass instead of being rejected — verified by
  temporarily commenting out the `if len(linear.Blocks) > 0 { return
  v.loanEndpointsMatch(function) }` branch during development and
  confirming all three variants were wrongly admitted, then restoring
  it with zero resulting diff against the committed file.
- `TestValidatorVerdictsUnchanged`: the whole pre-existing fixture
  corpus (owned/shared/exclusive/reborrow/branch) is re-asserted
  byte-for-byte against its pre-change verdict and first-problem code.
- `TestValidatorReborrowChainIsLinear`: measured `Checks`/facts ratio
  (1930/101=19.11, 19030/1001=19.01, 190030/10001=19.00) stays flat
  across a 100x input growth — a reintroduced quadratic scan would blow
  this ratio up by roughly the same 100x factor.
- `git diff HEAD~3..HEAD -- testdata/phase1/ testdata/phase2/` is
  empty — no Phase 1/2 golden moved.
- `sh scripts/verify-phase2.sh` exits 0 and reports all nine Phase 2
  control IDs and both Phase 1 controls through a freshly built binary.
- `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`,
  `go test -race ./...`, and `go vet ./...` all pass with zero findings.

## Known Stubs

None. The Phase 3 corpus's required-controls entry and shipped-CLI
assertion for `control:core.loan_endpoint_mismatch` are explicitly
deferred to 03-07 (where the Phase 3 verify path and gate script are
first assembled) per the plan's own task text — not a stub, a scoped
boundary the plan itself names.

## Next Phase Readiness

- `corevalidate` independently re-derives and enforces loan endpoints
  for every branch-shaped function, closing the T-03-13 self-confirming-
  differential risk for this fact.
- Both admission layers are linear in body size; D-02-03/Q2(b) is fully
  closed on the validator side (03-03 closed the checker side for the
  branch path only, per its own documented scope reduction).
- `session.BorrowedLoanEndpointControlLane` exists and is tested
  in-process; 03-07 needs only to add it to the Phase 3 corpus's
  required-controls list and drive it through the shipped CLI.
- **Flag for human review, carried forward from 03-03 and unresolved by
  this plan:** the two-liveness-derivation gap in the checker
  (`checkLinear` vs `checkBranch`) remains open; 03-05 is the phase's
  designated gate for it.
- No blockers for 03-05.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Plan: 04*
*Completed: 2026-09-04*

## Self-Check: PASSED

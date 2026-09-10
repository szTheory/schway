---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 06
subsystem: compiler-validation
tags: [go, corevalidate, cost-scaling, qlt02, disclosure, sem-05, mutation-kill]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "09-02: testsupport.GenerateCallGraphCorpus/CallGraphCorpusShapes and peer_closure_recomputed_work_growth_exponent declared at both gate-eligibility chokepoints (no manifest row); 09-04: the finding that the five relocated corpus shapes contain zero borrow operations"
provides:
  - "corevalidate's own counted-work cost lane (Result.Checks, driven through the existing loanChainIndex/foldChain/carriedLoans machinery) measured against operation count, on the peer's OWN bound (1300 milli-exponent), never check's ratified recomputed_work_growth_exponent"
  - "peerClosureUnmemoizedSeamForTest (corevalidate.go) and its mutation-kill, proving the bound has been seen to fail (0.999 -> 1.396 exponent, 4.58x more checks at n=512) when memoization is disabled on the one shape (forward) whose per-function chain depth grows with corpus size"
  - "lane:peer-closure-cost-scaling (risk_lanes.json + liveLanesPhase9), disjoint from lane:interprocedural-cost-scaling"
  - "Result.PeerConsultedFields() -- corevalidate's own disclosed callee-signature-field-set accessor, proven identical to check's interproceduralConsultObserved set (D-09-30), closing D-08-26's positive half"
affects: [09-08, 09-10]

actuals:
  tokens: 13400
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Injecting a genuine core.OpBorrowShared into every function of an otherwise borrow-free synthetic corpus (injectBorrowForLoanCarry), rather than sweeping the corpus as generated, so a cost-scaling sweep over loan-carry machinery is not decorative over a code path that never actually carries a loan"
    - "A memo-write skip (not a memo-read skip) as the minimal single-line mutation that defeats a memoization scheme while leaving every other invariant (bornAt, parent, the walk's own cycle guard) untouched"
    - "A per-function summary-derivation site (derivePeerSignature) as the peer-side analog of check's per-function interproceduralConsultObserved recording point -- both peers record two field names once per declared function, unconditionally, regardless of whether that function is ever called, even though check READS a stored interface value and corevalidate independently RE-DERIVES the equivalent fact"

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_peer_cost_test.go
    - internal/compiler/corevalidate/corevalidate_disclosure_test.go
    - internal/compiler/check/check_disclosure_peer_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/session/risk_lanes.json
    - internal/compiler/session/session_phase6_risklanes.go

key-decisions:
  - "The cost sweep does NOT run over testsupport's five call-graph shapes as generated. 09-04 found they contain zero core.OpBorrowShared/OpBorrowExclusive operations anywhere -- a naive sweep would exercise buildLoanChainIndex's bookkeeping but bornAt would never populate and carriedLoans would return an empty loan list on every call, measuring a code path that never carries a loan. injectBorrowForLoanCarry (corevalidate_peer_cost_test.go) prepends a genuine OpBorrowShared of the function's own parameter to every generated function and rewires every existing operation that read the raw parameter to read the freshly-borrowed place instead, so ReturnsBorrowOfParam is genuinely true for every function and buildLoanChainIndex's OpCall branch genuinely chains a call's result back through its argument's loan ancestry."
  - "PublicOrigin is declared only on LEAF functions (no OpCall of their own), never on relay functions. A leaf function's return directly derives from its own borrow (peerReturnDerivesFromBorrow sees the direct chain), so SEM-06's Callable-is-publication-safety law requires a declared PublicOrigin or the function is refused core.callee_not_callable. A relay function's return instead derives from an OpCall's result; peerDeriveOriginFacts/peerReturnDerivesFromBorrow do not trace through OpCall at all (a documented, pre-existing scope narrowing unrelated to this plan), so a relay function is already Callable with PublicOrigin left nil -- declaring one on a relay function would instead make peerOriginContained's !fact.Derived branch refuse it."
  - "The corpus this test measures actually re-derives loan-carry through recomputeLoanEndpoints's block-based endpoint machinery only when a function has Blocks; testsupport's generator builds Blocks-less (straight-line) functions, so what this sweep genuinely exercises is buildLoanChainIndex/carriedLoans/foldChain as consulted from replayStraightLine's own per-operation loop (corevalidate.go:1492-1501), not recomputeLoanEndpoints/blockReach. Documented explicitly in the test file's own header comment and restated here so a future reader does not conflate the two."
  - "TestPeerClosureCostUnmemoizedSeamExceedsBound sweeps ONLY the 'forward' shape, not all five. corpusForward is the one shape whose per-function chain depth grows with corpus size (a single caller threading one borrow through k sequential calls); the other four shapes' per-function chain depth stays O(1) by this test's own construction (every function borrows its own parameter exactly once, one hop from that borrow to every use), so disabling memoization cannot make them quadratic -- there is nothing deep to re-walk. Asserting the seam's effect only where it can structurally show up mirrors 09-04's own mutation-kill discipline (a hand-picked fixture, not the whole corpus)."
  - "peerClosureUnmemoizedSeamForTest skips foldChain's memo WRITE, not carriedLoans' memo READ. A write-skip is the minimal change that defeats memoization while leaving bornAt, parent, and the walk's own cycle guard untouched; a read-skip would additionally have to reimplement 'pretend this specific lookup missed' without disturbing the walk that follows it."
  - "derivePeerSignature was converted from a free function to a validator method (both call sites updated) so it can call the new v.recordPeerConsult choke point, which accumulates consulted field names onto the validator (surfaced via Result.PeerConsultedFields()) and optionally forwards to a same-package test hook (peerConsultObserved), mirroring check's own interproceduralConsultObserved recording point. Unlike check's seam (test-only, nil in production), corevalidate's own accumulation runs unconditionally in production, because PeerConsultedFields must be readable from an ordinary Validate call, not only from a test-installed hook."
  - "The two disclosed fields ('return.mode', 'parameters[0].mode') are recorded once per declared function inside derivePeerSignature, mirroring check's own per-function (not per-call-site) recording discipline exactly: check reads functionID's OWN declared return.mode/parameters[0].mode from originvalidate's interface to build that function's OWN interproceduralSummary entry (later consulted by its callers via summaries.lookup); corevalidate independently re-derives the equivalent facts (returnContract.Mode/parameterContract.Mode) for the same function, for the same reason (a later caller's own liveness/loan-carry fact reads v.peerLoanCarry[calleeID] and v.peerSignatures[calleeID], populated by this exact call)."

requirements-completed: []

coverage:
  - id: D1
    description: "The peer's re-derivation cost is measured on its own instrument (corevalidate.Result.Checks) against its own declared bound (1300 milli-exponent), fitted against operation count over five call-graph shapes each genuinely carrying a loan, never against check's ratified bound"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_cost_test.go#TestPeerClosureCostGrowthExponentWithinBound"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_cost_test.go#TestPeerClosureCostRatioStabilityTripwire"
        status: pass
      - kind: static
        ref: "git diff internal/compiler/session/qlt02_budget_manifest.json (no output -- untouched)"
        status: pass
    human_judgment: false
  - id: D2
    description: "The peer's bound has been seen to fail: an unmemoized seam reproduces a genuinely super-linear walk on the one shape whose chain depth grows with size, in both directions (seam on exceeds, seam off holds)"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_peer_cost_test.go#TestPeerClosureCostUnmemoizedSeamExceedsBound"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/corevalidate/... -count=2 -shuffle=on"
        status: pass
    human_judgment: false
  - id: D3
    description: "corevalidate has its own independently-written corpus-wide closed-consulted-field-set test, and a cross-peer test asserts the two peers' consulted field sets are identical across more than one program shape"
    requirement: "TRU-04"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_disclosure_test.go#TestPeerDisclosedFieldSet"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_disclosure_peer_test.go#TestPeerAndCheckDisclosedFieldSetsAreIdentical"
        status: pass
    human_judgment: false
  - id: D4
    description: "go test ./... && go vet ./... exits 0"
    verification:
      - kind: integration
        ref: "go test ./... && go vet ./... (full repo)"
        status: pass
    human_judgment: false

duration: 50min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 06: Peer Closure Cost and Disclosure Summary

**Measures corevalidate's own reachability-closure cost against a bound derived from its own observed curve (never borrowing check's ratified worklist bound), proves the bound genuinely fails under an unmemoized seam, and closes D-08-26's positive half by proving both admission peers disclose the identical closed set of callee-signature fields.**

## Performance

- **Duration:** ~50 min
- **Tasks:** 3 (3 commits, one per task)
- **Files created:** 3
- **Files modified:** 4

## Accomplishments

- `internal/compiler/corevalidate/corevalidate_peer_cost_test.go` (new, package `corevalidate_test`): `injectBorrowForLoanCarry` prepends a genuine `core.OpBorrowShared` to every function in `testsupport.CallGraphCorpusShapes()`'s five generated shapes and rewires every reference to the raw parameter to the freshly-borrowed place — required because 09-04 found the raw corpus contains zero borrow operations, which would have made a naive sweep decorative (constraint 5). `completePeerCostProgramForCorevalidate` performs the same structural completion `check_peer_shape_differential_test.go`'s own `completeSyntheticProgramForCorevalidate` does (independently written, since that helper is unexported and unreachable across the package boundary), plus declares `PublicOrigin` only on leaf functions (no `OpCall`) so `peerCallable`'s pre-existing SEM-06 gate does not refuse relay functions.
- `TestPeerClosureCostGrowthExponentWithinBound`: all five shapes fit flat-linear (milli-exponent ~999-1000) against a self-derived 1300 bound, driven off `corevalidate.Result.Checks` (never wall clock, logged only via `t.Logf`).
- `TestPeerClosureCostRatioStabilityTripwire`: `checks/operations` at S=128 vs 4S=512 stays within 0.1% for every shape.
- `TestPeerClosureCostUnmemoizedSeamExceedsBound`: with the new `peerClosureUnmemoizedSeamForTest` engaged (skips `foldChain`'s memo write), the "forward" shape's fitted exponent rises from 0.999 to 1.396 (exceeding the 1300 bound) with `checks` growing to 168683 vs 36844 at n=512 — a 4.58x multiplier — while the production path (seam off) stays within bound. `foldChain`'s memo-write skip was chosen over a `carriedLoans` memo-read skip as the minimal mutation that defeats memoization while leaving `bornAt`/`parent`/the cycle guard untouched.
- `lane:peer-closure-cost-scaling` added to `risk_lanes.json` (disjoint from `lane:interprocedural-cost-scaling`, naming `corevalidate`'s own source paths) and to `session.liveLanesPhase9()` so the registry audit recognizes it.
- `derivePeerSignature` converted to a validator method; a new `v.recordPeerConsult` choke point records `"return.mode"`/`"parameters[0].mode"` once per declared function, accumulated on the validator and surfaced via the new `Result.PeerConsultedFields()` accessor (fresh sorted copy, mirrors `PeerSignatures`'s own contract).
- `internal/compiler/corevalidate/corevalidate_disclosure_test.go` (new): `TestPeerDisclosedFieldSet`, written independently of `check`'s own `TestInterproceduralDisclosedFieldSet`, asserts the peer's consulted set is exactly `{"return.mode", "parameters[0].mode"}` over the real `testdata/phase08` `.lang` corpus plus the synthetic call-graph shapes.
- `internal/compiler/check/check_disclosure_peer_test.go` (new, package `check`): `TestPeerAndCheckDisclosedFieldSetsAreIdentical` proves the two peers' sets are identical sorted sets over a real fixture and a synthetic shape.

## Task Commits

1. **Task 1: measure the peer's closure cost curve against its own declared bound** — `699c436` (feat)
2. **Task 2: mutation-kill the peer's bound and give it a changed-risk lane** — `b36693c` (feat)
3. **Task 3: the peer discloses too — closed field sets asserted identical** — `7574cd7` (feat)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate_peer_cost_test.go` — the cost sweep, borrow injection, growth-exponent fit, ratio tripwire, and the mutation-kill test
- `internal/compiler/corevalidate/corevalidate.go` — `peerClosureUnmemoizedSeamForTest` (beside `foldChain`), `peerConsultObserved`/`recordPeerConsult`, `derivePeerSignature` → method, `Result.PeerConsultedFields()`
- `internal/compiler/corevalidate/export_test.go` — `SetPeerClosureUnmemoizedSeamForTest`, `SetPeerConsultObservedForTest`
- `internal/compiler/corevalidate/corevalidate_disclosure_test.go` — `TestPeerDisclosedFieldSet`
- `internal/compiler/check/check_disclosure_peer_test.go` — `TestPeerAndCheckDisclosedFieldSetsAreIdentical`
- `internal/compiler/session/risk_lanes.json` — `lane:peer-closure-cost-scaling` row (existing `lane:interprocedural-cost-scaling` row unmodified)
- `internal/compiler/session/session_phase6_risklanes.go` — `liveLanesPhase9()`, wired into `LiveLaneIDs()`

## Decisions Made

See `key-decisions` in frontmatter for full detail. Summary:

- The sweep injects a genuine borrow into every corpus function rather than sweeping the borrow-free corpus as generated (constraint 5 honesty requirement).
- `PublicOrigin` is declared only on leaf functions, never relay functions, to stay compatible with `peerCallable`'s pre-existing (and unrelated) scope narrowing around `OpCall`.
- The mutation-kill sweeps only the "forward" shape, since it is the one shape whose chain depth structurally grows with corpus size.
- `foldChain`'s memo write (not `carriedLoans`' memo read) is the mutation surface.
- `derivePeerSignature`'s existing per-function derivation is the disclosure recording point, mirroring `check`'s own per-function recording discipline even though the two peers read vs. re-derive the underlying fact by materially different mechanisms.

## Honesty Disclosure (phase constraint 5)

**What this plan's cost sweep exercises:** `buildLoanChainIndex`/`carriedLoans`/`foldChain` as consulted from `replayStraightLine`'s own per-operation loop, over programs where every function genuinely borrows and returns a borrow of its own parameter (via `injectBorrowForLoanCarry`), so loan-carry propagation across the call graph is genuinely non-vacuous.

**What it does NOT exercise:** `recomputeLoanEndpoints`/`blockReach` (the block-based endpoint-recomputation closure) — `testsupport`'s generator builds straight-line (Blocks-less) functions, so that code path never runs for this corpus. A reader should not read this plan's flat-linear finding as bounding the block-based closure's own cost curve; it bounds the place-provenance/loan-chain walk that both straight-line and block-shaped replay share.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Borrow-injected corpus programs needed PublicOrigin declared, but only on leaf functions**
- **Found during:** Task 1, first run of `TestPeerClosureCostGrowthExponentWithinBound`
- **Issue:** After injecting a genuine borrow into every function, `peerCallable`'s SEM-06 gate refused every callee as `core.callee_not_callable`: declaring `PublicOrigin` unconditionally on every function (including relay functions, whose return derives from an `OpCall` result, not a direct borrow) caused `peerOriginContained`'s `!fact.Derived` branch to refuse relay functions, since `peerDeriveOriginFacts` does not trace through `OpCall`.
- **Fix:** Declare `PublicOrigin` only on functions with no `OpCall` of their own (leaf functions); relay functions stay `PublicOrigin == nil`, where `!peerReturnDerivesFromBorrow(function)` (also untraced through `OpCall`) correctly evaluates to `Callable == true`.
- **Files modified:** `internal/compiler/corevalidate/corevalidate_peer_cost_test.go`
- **Verification:** `TestPeerClosureCostGrowthExponentWithinBound` passes across all five shapes
- **Committed in:** `699c436`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary to make the honest (borrow-carrying) sweep structurally valid under corevalidate's own pre-existing SEM-06 gate. No scope creep — no production behavior changed beyond the plan's own scope (the seam and the disclosure accessor).

## Issues Encountered

See Deviations above. No other issues.

## User Setup Required

None — no external service configuration required.

## Known Stubs

None.

## Threat Flags

None — this plan's threat register items (T-09-04, T-09-17, T-09-18, T-09-19, T-09-20) are all directly mitigated by the shipped mechanisms described above; no new surface introduced beyond what the plan's own threat model named.

## Next Phase Readiness

- `peer_closure_recomputed_work_growth_exponent` has a measured value (1300 milli-exponent bound, all five shapes observed at ~999-1000) ready for plan 09-08's mid-phase gate to ratify into `qlt02_budget_manifest.json` at that gate's own commit. **No manifest row was added by this plan** (D-09-28, D-08-33's precedent).
- D-08-26 is now fully resolved: the refusal half by `check`'s own shipped diagnostic (Phase 08), the accepted-program half's negative side as "no vehicle, by design" (D-09-29), and its positive side by this plan's `Result.PeerConsultedFields()` plus the cross-peer identity test (D-09-30).
- TRU-04 stays `Pending` in REQUIREMENTS.md per this plan's own requirement-marking rule (also carried by 09-01, 09-02, 09-04, 09-08, 09-10).
- No blockers for 09-07 through 09-10 from this plan's own scope.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

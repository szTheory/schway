---
phase: 10-trusted-interprocedural-oracle
plan: 02
subsystem: originvalidate
tags: [ownership, origin-validation, interprocedural, call-graph, go-parser, mutation-testing]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plan 10-01's real []frame call-stack executor (core.OpCall as a genuinely executed operation), the fixed ground this plan's origin re-derivation now trusts"
provides:
  - "originvalidate.walkReturnOrigin consults a callee's DECLARED return contract at an OpCall hop via a new, narrow map[string]calleeOriginFact parameter (BuildCalleeOriginFacts), closing the inherited D-09-51 defect: testdata/phase08/twin_a_accept.lang and relay_depth2_accept.lang now admit cleanly through the full `lang check` CLI"
  - "A D-10-08 nil-default gate seam (disableOpCallOriginConsultForTest / SetDisableOpCallOriginConsultForTest) proving the new case is load-bearing, not merely present, plus a positive-direction test proving the fix narrows rather than disables propagation"
  - "originvalidate's import-independence guards hardened from direct-import-only to also transitive (go list -deps), with corevalidate added to the forbidden set (D-10-16/D-10-17)"
affects: [10-03, 10-04, 10-05, 10-06, 10-07, 10-08, 10-09, 11]

# Actuals (#2632)
actuals:
  tokens: 16082
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Narrow callee-contract map (calleeOriginFact: access + derived-ness only) threaded as an explicit added parameter through a chain of functions, mirroring corevalidate.buildLoanChainIndex's shipped peerLoanCarry precedent (D-10-03) -- a caller never gains the ability to reach a callee's body, a property enforced by a signature-text go/parser assertion with a negative control, not merely a doc comment"
    - "Widen-then-use commit pair (Google-LSC style, D-10-06): commit 1 threads a new parameter with zero behavior change (verified suite-green + fixture-unchanged); commit 2 adds the one case that consumes it and is the actual behavior change"
    - "D-10-08 nil-default gate seam proving a switch case is load-bearing by observing a real regression when disabled, not merely that the case is syntactically present"
    - "Transitive go list -deps import-independence guard layered on top of (not replacing) an existing go/parser direct-import scan, proven load-bearing by a negative control over a synthetic dependency list"

key-files:
  created: []
  modified:
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/originvalidate/originvalidate_closure_chain_test.go
    - internal/compiler/originvalidate/export_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/core/core_test.go
    - internal/compiler/check/check_test.go
    - internal/compiler/check/check_ordering_stability_test.go
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - internal/compiler/corevalidate/corevalidate_peer_origin_test.go

key-decisions:
  - "calleeOriginFact stays unexported (per plan), but BuildCalleeOriginFacts (the map's builder) is exported: every external call site that calls RecomputeOriginPerReturn/RecomputeOrigin/PublishProblemsFor directly (not through ValidatePublished) needs to build the identical map from its own core.Program, and Go permits an exported function to return an unexported element type -- callers hold and pass it back via `:=` type inference without ever spelling the type name"
  - "OpCall's declared-contract source is core.Function.PublicOrigin (Access+Paths), read directly -- never RecomputeOrigin's own re-derived answer -- matching corevalidate.buildLoanChainIndex's own consult of a DECLARED (not re-derived) peer fact"
  - "The plan's own blast-radius estimate (files_modified: 6 files) undercounted by 5: RecomputeOrigin's and PublishProblemsFor's widened signatures are also called directly by check/check_test.go and two corevalidate test files, none of which the plan anticipated. All were fixed under Rule 3 (blocking -- the build would not compile otherwise)"

patterns-established:
  - "A callee-contract map's element type can stay unexported while its builder function is exported, letting outside packages construct and re-pass it without ever being able to add a body-reachable field to it -- a body-blindness guarantee enforced by the type system rather than a per-caller promise"

requirements-completed: [TRU-02]

coverage:
  - id: D1
    description: "walkReturnOrigin consults a callee's DECLARED return contract at an OpCall hop instead of walking transparently through into the argument's own provenance"
    requirement: TRU-02
    verification:
      - kind: unit
        ref: "originvalidate_test.go#TestOpCallOriginWalkGateIsLoadBearing"
        status: pass
      - kind: unit
        ref: "originvalidate_test.go#TestOpCallOriginWalkPropagatesGenuineBorrowingCallee"
        status: pass
    human_judgment: false
  - id: D2
    description: "go run ./cmd/lang --json check testdata/phase08/twin_a_accept.lang exits 0 with zero diagnostics through the full CLI pipeline, closing the inherited D-09-51 defect"
    requirement: TRU-02
    verification:
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase08/twin_a_accept.lang"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase08/relay_depth2_accept.lang"
        status: pass
    human_judgment: false
  - id: D3
    description: "The callee-contract map cannot be used to reach a callee's body -- enforced mechanically, not by comment"
    verification:
      - kind: unit
        ref: "originvalidate_test.go#TestWalkReturnOriginSignatureStaysBodyBlind"
        status: pass
    human_judgment: false
  - id: D4
    description: "originvalidate transitively imports neither check nor corevalidate, enforced by a go list -deps check that fails CI"
    requirement: TRU-02
    verification:
      - kind: unit
        ref: "originvalidate_test.go#TestOriginValidatorImportsStayIndependent"
        status: pass
      - kind: unit
        ref: "originvalidate_test.go#TestOriginValidateImportsNeitherCheckNorAst"
        status: pass
      - kind: unit
        ref: "originvalidate_test.go#TestTransitiveImportsGuardCanFail"
        status: pass
    human_judgment: false
  - id: D5
    description: "Two check-package tests/fixtures that were unknowingly relying on the same pre-existing D-09-51 transparent-walk defect are corrected to the newly-honest verdict, with a written justification rather than a silent regeneration"
    verification:
      - kind: unit
        ref: "check_test.go#TestInterproceduralLivenessNegativeControl"
        status: pass
      - kind: unit
        ref: "check_ordering_stability_test.go#TestInterproceduralDiagnosticOrderingStability"
        status: pass
    human_judgment: true
    rationale: "The correctness of switching these two fixtures' expected diagnostic from check.interprocedural_loan_liveness to core.callee_not_callable rests on a chain of reasoning (relay's declared borrow(buffer) return was never honestly derivable given leaf's genuinely owned return in either fixture) documented in this plan's Deviations section and inline in both test files; a human reviewer should confirm that reasoning before treating it as settled."

duration: 95min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 2: Trusted Interprocedural Oracle — Callee-Contract Origin Walk Summary

`originvalidate.walkReturnOrigin` now consults a callee's declared return contract at every `core.OpCall` hop through a narrow, body-blind `map[string]calleeOriginFact`, closing the inherited D-09-51 defect end-to-end through the full `lang check` CLI, at the cost of correcting two check-package fixtures that had unknowingly depended on the same transparent-walk bug for two phases.

## Performance

- **Duration:** 95 min
- **Started:** 2026-09-11 (see git commit timestamps)
- **Completed:** 2026-09-11
- **Tasks:** 3 completed
- **Files modified:** 11

## Accomplishments

- Added `calleeOriginFact` (unexported: access mode + derived-ness only) and `BuildCalleeOriginFacts`, threaded as an explicit added parameter through `walkReturnOrigin` → `RecomputeOriginPerReturn` → `RecomputeOrigin` → `PublishProblemsFor`, built once per program in `ValidatePublished` and `BuildInterface` (both signatures unchanged, per plan). Landed as a behavior-neutral commit (Task 1), verified suite-green and `twin_a_accept.lang` still refusing, before Task 2's actual behavior change.
- Added `case core.OpCall` to `walkReturnOrigin`: it looks up the callee's DECLARED `PublicOrigin` and propagates the borrow access mode only when the callee genuinely returns a borrow of its own parameter, breaking the chain otherwise — closing D-09-51. `testdata/phase08/twin_a_accept.lang` and `relay_depth2_accept.lang` now report `"status":"pass"` with zero diagnostics through the full `lang check` CLI (both were previously refused with `core.origin_omitted`).
- A D-10-08 nil-default gate seam (`disableOpCallOriginConsultForTest`, exposed via `export_test.go`'s `SetDisableOpCallOriginConsultForTest`) and its test (`TestOpCallOriginWalkGateIsLoadBearing`) prove the case is load-bearing: engaging the seam regresses `twin_a_accept.lang` back to exactly `core.origin_omitted`, run through `session.CheckCommandFile`'s full fixed-precedence pipeline (never a direct package call). A positive-direction test (`TestOpCallOriginWalkPropagatesGenuineBorrowingCallee`) proves the fix narrows rather than disables: a genuinely borrow-returning callee still propagates.
- `TestWalkReturnOriginSignatureStaysBodyBlind` mechanically asserts, via `go/parser`, that `walkReturnOrigin`/`RecomputeOriginPerReturn`/`RecomputeOrigin` never gain a `core.Function`- or `core.Program`-shaped parameter besides the function each already analyzes, with a negative control over a synthetic leaking signature.
- Both existing import-independence guards now also forbid `/compiler/corevalidate` (D-10-16) and both were hardened from a direct-import-only `go/parser` scan to ALSO run a transitive `go list -deps` check (D-10-17), proven load-bearing by `TestTransitiveImportsGuardCanFail`'s negative control. The two guards deliberately stay separate (D-10-18) and `/compiler/callgraph` is deliberately NOT added to the forbidden list (D-10-19); both decisions and the guard's residual same-commit bypassability (D-10-20) are recorded in comments citing `PHASE-10-DEBT.md`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Widen the origin walk with a narrow callee-contract map** - `bb4aa02` (feat)
2. **Task 2: Consult the callee contract at the OpCall hop** - `6648b72` (feat)
3. **Task 3: Harden originvalidate's import guards — add corevalidate, upgrade to transitive** - `7d7553b` (test)

## Files Created/Modified

- `internal/compiler/originvalidate/originvalidate.go` — `calleeOriginFact`, `BuildCalleeOriginFacts`, `disableOpCallOriginConsultForTest`, `case core.OpCall`; widened `walkReturnOrigin`/`RecomputeOriginPerReturn`/`RecomputeOrigin`/`PublishProblemsFor`.
- `internal/compiler/originvalidate/export_test.go` — `SetDisableOpCallOriginConsultForTest`.
- `internal/compiler/originvalidate/originvalidate_test.go` — new tests for all three tasks (signature-blindness, gate, positive-direction, transitive import guards); updated ~16 pre-existing call sites to the widened signatures.
- `internal/compiler/originvalidate/originvalidate_closure_chain_test.go` — `foreignReachOnlyCallProgramForTest`, replacing `TestClosureDigestEmptyCalleesMutationKilled`'s use of the shared origin-toggling helper (see Deviations).
- `internal/compiler/session/session.go`, `internal/compiler/session/session_phase7.go` — build `calleeContracts` once per dispatch fixture, pass to `RecomputeOriginPerReturn`.
- `internal/compiler/core/core_test.go` — same, for `runExhaustiveDispatchControl`.
- `internal/compiler/check/check_test.go` — pass `originvalidate.BuildCalleeOriginFacts(program)` at its own `PublishProblemsFor` call site; updated `TestInterproceduralLivenessNegativeControl`'s expected diagnostic (see Deviations).
- `internal/compiler/check/check_ordering_stability_test.go` — updated the literal pre-restructure baseline table for `negative_control_fails.lang`/`negative_control_infallible.lang` with a written justification.
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go`, `internal/compiler/corevalidate/corevalidate_peer_origin_test.go` — pass `originvalidate.BuildCalleeOriginFacts(program)` at each of their own `PublishProblemsFor`/`RecomputeOrigin` call sites.

## Decisions Made

- `BuildCalleeOriginFacts` is exported even though `calleeOriginFact` stays unexported, so every external call site that invokes `RecomputeOriginPerReturn`/`RecomputeOrigin`/`PublishProblemsFor` directly can build the identical map from its own `core.Program` — Go permits this (an exported function may return an unexported element type; callers use `:=` and never spell the type name).
- The OpCall case's declared-contract source is `core.Function.PublicOrigin` read directly, never `RecomputeOrigin`'s own re-derived answer — mirroring `corevalidate.buildLoanChainIndex`'s consult of a DECLARED (not re-derived) peer fact, and keeping the map buildable in one pass with no fixpoint iteration.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Plan's file list undercounted the widened signatures' blast radius**
- **Found during:** Task 1, `go vet ./...`
- **Issue:** `RecomputeOrigin`'s and `PublishProblemsFor`'s widened signatures are also called directly by `internal/compiler/check/check_test.go` and two corevalidate test files (`corevalidate_summary_peer_test.go`, `corevalidate_peer_origin_test.go`) — none of which the plan's `files_modified` frontmatter listed. The build would not compile without updating them.
- **Fix:** Each call site now builds `originvalidate.BuildCalleeOriginFacts(program)` from the `core.Program` already in scope and passes it through.
- **Files modified:** `internal/compiler/check/check_test.go`, `internal/compiler/corevalidate/corevalidate_summary_peer_test.go`, `internal/compiler/corevalidate/corevalidate_peer_origin_test.go`
- **Verification:** `go build ./...` and `go vet ./...` both exit 0; full package suites pass.
- **Committed in:** `bb4aa02`

**2. [Rule 1 - Bug] Two check-package fixtures were unknowingly relying on the same pre-existing D-09-51 transparent-walk defect**
- **Found during:** Task 2, `go test ./...`
- **Issue:** `testdata/phase08/negative_control_fails.lang` and `negative_control_infallible.lang` both declare `relay -> borrow(buffer) Buffer`, but in EITHER fixture `leaf` genuinely returns an OWNED value (an ordinary Lang function in the infallible member, a freshly-allocated foreign-call result in the fails member) — neither `leaf`'s body nor its declared contract derives from a borrow of its own parameter at all. Before this plan's Task 2, `originvalidate.walkReturnOrigin` walked TRANSPARENTLY through relay's `leaf(borrowed)` OpCall, so relay's dishonest `borrow(buffer)` declaration was never independently caught by origin re-validation, and `check.interprocedural_loan_liveness` (a separate law that trusts relay's declared contract) was the first diagnostic ever reported for `caller`. With the fix, `originvalidate.PublishProblemsFor` correctly re-derives relay's origin as `core.origin_understated`, so `check`'s own self-consistency gate (SEM-06's `verifyCallableRefusal`, D-07-31/D-07-34) now refuses `caller`'s call to `relay` with `core.callee_not_callable` BEFORE the interprocedural loan-liveness law ever runs (`check.Program`'s diagnostic-accumulation order always runs `verifyCallableRefusal` first). A sibling fixture, `relay_depth2_refuse.lang`, explicitly documents in its own header comment that it was engineered to "route around" this exact "pre-existing gate" — these two negative-control fixtures were built on the same assumption without documenting it.
- **Fix:** Updated `TestInterproceduralLivenessNegativeControl` (`check_test.go`) to expect `core.callee_not_callable` with a single `{Kind: "callee", Detail: <relay's function ID>}` cause for both members (the test's original intent — that Fails/ForeignReach do not affect the verdict — still holds and is still demonstrated, just via a different diagnostic). Updated `TestInterproceduralDiagnosticOrderingStability`'s literal baseline table (`check_ordering_stability_test.go`) for both fixtures with a written justification, per that test's own stated contract ("a difference here must be justified in the owning plan's SUMMARY, not silently regenerated").
- **Files modified:** `internal/compiler/check/check_test.go`, `internal/compiler/check/check_ordering_stability_test.go`
- **Verification:** `go test ./internal/compiler/check/...` exits 0; manually probed both fixtures' new `(code, id)` tuples before writing them into the baseline table.
- **Committed in:** `6648b72`

**3. [Rule 1 - Bug] A shared test helper's mutation confounded Callable with ClosureDigest chaining**
- **Found during:** Task 2, `go test ./internal/compiler/originvalidate/...`
- **Issue:** `TestClosureDigestEmptyCalleesMutationKilled` used `straightLineCallProgramForTest`'s `calleeHasDeclaredOrigin` toggle (its shared helper across several ClosureDigest tests) to vary "the callee's published signature." That helper's own doc comment stated the toggle affects ONLY the callee's own signature — true before this plan, but no longer: once `walkReturnOrigin` gained a real `core.OpCall` case, toggling the callee's declared origin ALSO flips the CALLER's own `Callable` bit (the caller's body directly forwards its own parameter into the call, so a callee that genuinely declares a borrow-of-parameter return makes the caller's own undeclared return newly derived, i.e. `core.origin_omitted`). Since `Callable` is part of `FunctionSignature` (hence the `ClosureDigest` preimage), this second, previously-uncontrolled variable made the test's `closureDigestEmptyCalleesOverride`-engaged assertion (digests must be IDENTICAL) fail even though the calleePairs/Foreign/Fails chaining property under test was working correctly.
- **Fix:** Added `foreignReachOnlyCallProgramForTest`, a local variant that keeps `PublicOrigin` nil in both variants and instead toggles `callee.ForeignContract` (changing the callee's Foreign/Fails and hence its own `ClosureDigest`, without ever touching origin). Used only by this one test; every other test using `straightLineCallProgramForTest`'s origin toggle is unaffected and unchanged.
- **Files modified:** `internal/compiler/originvalidate/originvalidate_closure_chain_test.go`
- **Verification:** `go test ./internal/compiler/originvalidate/...` exits 0, including `TestClosureDigestEmptyCalleesMutationKilled`.
- **Committed in:** `6648b72`

---

**Total deviations:** 3 auto-fixed (1 Rule 3, 2 Rule 1).
**Impact on plan:** All three were necessary consequences of correctly closing D-09-51 system-wide (originvalidate is consulted from check's own self-consistency gate, not only from the two CLI fixtures the plan explicitly named) or of keeping the build compiling. No scope creep: no new feature or architectural surface was added beyond what the plan's Tasks 1-3 specify; every fix is a correction to a pre-existing test/fixture that had (knowingly or not) depended on the exact defect this plan closes.

## Issues Encountered

One transient, unrelated test failure (`TestShippedBinaryExercisesEveryPhase4Behavior/Fallible_foreign_call_through_try` in `internal/compiler/native`) surfaced during a full `go test ./...` run as a `native.timeout` — confirmed as pre-existing environment flakiness (native-compiler timeout under parallel test load), not caused by this plan's changes: re-ran in isolation both before and after this plan's commits and it passed in under 2s each time.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Success Criterion 1's first half (the `OpCall` origin walk) is closed; the callee-contract map pattern (D-10-03) and the widen-then-use commit discipline (D-10-06) are established precedents for any later plan needing a similar cross-function, body-blind fact.
- The must_haves' own flagged assumption ("EDGE (TRU-02, unclassified — UNRESOLVED)... reviewer must confirm manually that the OpCall origin hop has no unconsidered edge beyond the OpForeignCall-mirror inexactness") is PARTIALLY addressed by this plan's own discovery: the "unconsidered edge" materialized as check's OWN internal Callable gate (SEM-06) also being downstream of `originvalidate.PublishProblemsFor`, affecting two pre-existing negative-control fixtures. A human reviewer should still confirm no OTHER downstream consumer of `originvalidate`'s recomputed origin (beyond `check`'s Callable gate and the two CLI fixtures) is similarly affected — a full corpus walk (`TestPublishProblemsForMatchesValidatedAcrossCorpus`, `TestInterproceduralDiagnosticOrderingStability`) found none beyond the two documented above, but a full-repository audit was not separately commissioned.
- Ready for plan 10-03.

## Self-Check: PASSED

- `internal/compiler/originvalidate/originvalidate.go` — FOUND
- `internal/compiler/originvalidate/originvalidate_test.go` — FOUND
- `internal/compiler/originvalidate/originvalidate_closure_chain_test.go` — FOUND
- `internal/compiler/originvalidate/export_test.go` — FOUND
- `internal/compiler/session/session.go` — FOUND
- `internal/compiler/session/session_phase7.go` — FOUND
- Commit `bb4aa02` — FOUND (`git log --oneline --all`)
- Commit `6648b72` — FOUND
- Commit `7d7553b` — FOUND
- `go test ./internal/compiler/originvalidate/... -v` — PASS (all tests, all named)
- `go test ./internal/compiler/originvalidate/... ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/session/... ./internal/compiler/core/...` — PASS
- `go build ./... && go vet ./...` — PASS
- `go run ./cmd/lang --json check testdata/phase08/twin_a_accept.lang` — exit 0, `"status":"pass"`, empty diagnostics
- `go run ./cmd/lang --json check testdata/phase08/relay_depth2_accept.lang` — exit 0, `"status":"pass"`, empty diagnostics
- `go list -deps ./internal/compiler/originvalidate | grep -c 'compiler/corevalidate\|compiler/check\|compiler/ast'` — prints `0`

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*

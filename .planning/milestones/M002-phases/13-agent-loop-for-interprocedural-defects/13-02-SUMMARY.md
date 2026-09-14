---
phase: 13-agent-loop-for-interprocedural-defects
plan: 02
subsystem: check
tags: [blame-attribution, interprocedural, diagnostics, contract-boundary]

# Dependency graph
requires:
  - phase: 13-agent-loop-for-interprocedural-defects
    provides: "13-01's move_after_interprocedural_loan repair and the check.go/check_ordering_stability_test.go state it left check.go in (the same file this plan edits)"
provides:
  - "D-13-01..D-13-08's contract-boundary blame resolver (resolveBlame, resolveCycleBlame, classifyDeclaredCause) in internal/compiler/check, new infrastructure ready for a future B1-shaped defect class"
  - "The shared calleeBeforeCallerOrder helper (D-13-03) as check.go's ONE non-comment callgraph.Order call site"
  - "functionOwnerIndex/buildFunctionByOperationID (D-13-02's resolver data structure)"
  - "D-13-07's compile-time exhaustiveness guard over the five declared-contract fields, demonstrated to fail the build when an arm is deleted"
  - "Empirical confirmation (TestBlameMovesNoPrimarySpanToday) that adopting the blame rule moves zero Primary spans on all eight currently-shipped interprocedural codes"
affects: [13-03, 13-04, 13-05, 13-06, 13-07]

# Actuals (#2632)
actuals:
  tokens: 14583
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Compile-time exhaustiveness guard via a fixed-size array literal assigned from a \"...\"-length keyed literal: deleting one keyed entry shrinks the inferred length below the fixed target, producing a genuine Go array-length type error rather than a silently-passing switch"
    - "Declared-fact classification is scoped to a NAMED, small vocabulary (D-13-02's five FunctionSignature fields) rather than attempting to classify every diagnostic.Cause kind the package emits -- causes about a different contract dimension entirely (type contracts, arity contracts) are out of scope for this resolver, not silently folded into \"unclassified\""
    - "One shared ordering helper (calleeBeforeCallerOrder) consumed by both the pre-existing summary builder and the new blame resolver's tie-break, so \"no new ordering authority is introduced\" holds by construction (single call site), not by review"

key-files:
  created:
    - internal/compiler/check/check_blame_test.go
  modified:
    - internal/compiler/check/check.go

key-decisions:
  - "The blame resolver (resolveBlame/resolveCycleBlame/classifyDeclaredCause) is built and exhaustively tested as standalone infrastructure but is NOT wired into any of the eight existing diagnostic emission sites' Primary-span computation. D-13-04 (verified empirically by TestBlameMovesNoPrimarySpanToday) confirms all eight already select the Primary span B2 would select, so wiring would be a no-op today while carrying real risk (touching identity-bearing Primary spans on published diagnostics). The resolver is ready for the first B1-shaped defect class a future plan introduces."
  - "classifyDeclaredCause recognizes exactly two cause shapes as D-13-02 declared facts today: callee_return_contract (Return field, parsed per D-13-02's own instruction) and callee-on-core.callee_not_callable (Callable field, settled B2). Every other cause kind (declared_parameter_type, declared_arity, argument_type, place, cycle_member, etc.) is a DIFFERENT contract dimension (type contract, arity contract, use-site fact) that D-13-02's five-field ownership-contract vocabulary was never written to cover, so classifyDeclaredCause reports these out of scope (ok == false) rather than folding them into D-13-07's unclassified/blame_undetermined trigger. This is what keeps all eight shipped codes landing cleanly in B2 rather than spuriously reaching blame_undetermined."
  - "A non-Callable callee is settled as CALLER MISUSE (B2), not a callee contract violation (B1) -- 13-CONTEXT.md's own \"Claude's Discretion\" list named this the one genuinely defensible-both-ways field. Recorded as a comment at verifyCallableRefusal's own emission site (check.go) so the shipped B2 behavior is a considered decision, not an artifact of implementation order."
  - "checkCallGraphAcyclic is now routed through calleeBeforeCallerOrder rather than calling callgraph.Order directly, even though it only ever consults the returned error. This was needed to satisfy the plan's own acceptance criterion (\"exactly one non-comment callgraph.Order call site in check.go\") -- the plan's read_first section did not flag checkCallGraphAcyclic's own pre-existing Order() call as a second site, but the criterion's grep is over the WHOLE file. Routing it through the shared helper makes the criterion literally true and further strengthens D-13-03's \"no new ordering authority\" claim (propagated error is byte-identical either way)."
  - "check.call_arity_unsupported and check.foreign_call_shape_unsupported have no dedicated testdata/phase07|phase08|phase4 fixture (confirmed by corpus search) -- TestBlameMovesNoPrimarySpanToday drives both from small inline .lang sources instead of a testdata file, and check.call_return_type_unrepresentable (structurally unreachable from legal source, per check.go's own existing doc comment) is exercised directly against resolveCallBinding, mirroring the package's own established callReturnTypeDerivationSeam convention (TestCallReturnTypeDerivationMutationKilled)."

requirements-completed: [DX-06]

coverage:
  - id: D1
    description: "resolveBlame implements D-13-03's B1/B2/B3 rule with D-13-07's fail-open closure (blame_undetermined), classifying real fixture-derived causes from five of the eight shipped codes into B2 and covering B1/B3/undetermined behavior via synthetic cases"
    requirement: "DX-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_blame_test.go#TestBlameClassifiesEveryInterproceduralCode"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_blame_test.go#TestBlameUndeterminedPublishesBothSitesUnapplied"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_blame_test.go#TestBlameTieBreakIsCalleeBeforeCaller"
        status: pass
    human_judgment: false
  - id: D2
    description: "core.call_graph_cycle is exempt from B3 (D-13-05): Primary stays at the closing operation's own span, every cycle member is published as a secondary cause via resolveCycleBlame, never resolveBlame"
    requirement: "DX-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_blame_test.go#TestBlameCycleIsExemptFromB3"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-13-07's compile-time exhaustiveness guard genuinely fails the build when one declared-contract-field arm is deleted"
    requirement: "DX-06"
    verification:
      - kind: other
        ref: "scratch deletion of blameFieldWitness's declaredFieldAbilities entry + go build, reverted (see Deviations/verbatim error below)"
        status: pass
    human_judgment: false
  - id: D4
    description: "Adopting the blame rule moves zero Primary spans on all eight currently-shipped interprocedural codes (D-13-04), verified empirically against real (or, for two codes with no corpus fixture, inline) source"
    requirement: "DX-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_blame_test.go#TestBlameMovesNoPrimarySpanToday"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_ordering_stability_test.go#TestInterproceduralDiagnosticOrderingStability"
        status: pass
    human_judgment: false
  - id: D5
    description: "internal/compiler/corevalidate carries no blame derivation (D-13-06); the peer gate agrees on refusal, never on blame site"
    requirement: "DX-06"
    verification:
      - kind: other
        ref: "grep -rn 'blame' internal/compiler/corevalidate/ --include='*.go' | grep -v '_test.go' | grep -v '//' (zero matches)"
        status: pass
    human_judgment: false

# Metrics
duration: 58min
completed: 2026-09-13
status: complete
---

# Phase 13 Plan 02: Contract-Boundary Blame Resolver Summary

**D-13-01..D-13-08's B1/B2/B3 blame resolver (`resolveBlame`/`resolveCycleBlame`/`classifyDeclaredCause`) lands in `internal/compiler/check` with a compile-time exhaustiveness guard over the five declared-contract fields, a `blame_undetermined` fail-open closure, and empirical proof it moves zero Primary spans on any of the eight currently-shipped interprocedural codes.**

## Performance

- **Duration:** 58 min
- **Started:** 2026-09-13T18:23:00Z (approx, session start)
- **Completed:** 2026-09-13T19:21:19Z
- **Tasks:** 3
- **Files modified:** 2 (1 modified, 1 created)

## Accomplishments
- Built `functionOwnerIndex`/`buildFunctionByOperationID` (D-13-02's resolver data structure: operation ID → owning function ID, total and exact, refuses on duplicate operation IDs rather than guessing) and `calleeBeforeCallerOrder` (D-13-03's shared ordering helper — `callgraph.Order` reversed, called exactly once)
- Refactored `buildInterproceduralSummaries` onto `calleeBeforeCallerOrder` and routed `checkCallGraphAcyclic` through the same helper, so `check.go` carries exactly ONE non-comment `callgraph.Order(` call site — "no new ordering authority is introduced" (D-13-03) holds by construction, verified via `grep`
- Implemented the B1/B2/B3 rule (`resolveBlame`) with D-13-07's MANDATORY compile-time exhaustiveness guard (`blameFieldWitness`, a fixed-size array literal that fails to compile if a declared-contract-field arm is deleted — demonstrated this session, error text below) and the `blame_undetermined` fail-open closure (never falls through to B2 on an unclassified declared fact)
- Implemented D-13-05's `core.call_graph_cycle` exemption (`resolveCycleBlame`) as a separate, named code path
- Settled the one open discretion item (non-`Callable` callee is caller misuse, B2) with a rationale comment at `verifyCallableRefusal`'s own emission site
- Empirically verified D-13-04's "zero Primary spans move" claim via `TestBlameMovesNoPrimarySpanToday`, computing each of the eight codes' expected Primary span independently from the parsed AST rather than reading it back from the diagnostic

## Task Commits

Each task was committed atomically:

1. **Task 1: `functionByOperationID` + `calleeBeforeCallerOrder`** - `6967f1b` (feat)
2. **Task 2: The B1/B2/B3 rule and the `blame_undetermined` fail-open closure** - `eb8c22f` (feat)
3. **Task 3: Empirically verify D-13-04's zero-Primary-movement claim** - `3a13274` (test)

**Plan metadata:** (this commit)

## Files Created/Modified
- `internal/compiler/check/check.go` - Adds `functionOwnerIndex`/`buildFunctionByOperationID`, `calleeBeforeCallerOrder`, the B1/B2/B3 blame resolver (`resolveBlame`, `resolveCycleBlame`, `classifyDeclaredCause`, `blameFieldWitness`'s compile-time exhaustiveness guard, `blameUndeterminedRepairs`), refactors `buildInterproceduralSummaries` and `checkCallGraphAcyclic` onto the shared ordering helper, and adds a settled-decision comment at `verifyCallableRefusal`
- `internal/compiler/check/check_blame_test.go` - New file: 10 `TestBlame*`-prefixed tests covering both resolver inputs, the B1/B2/B3 rule against five real fixture-derived diagnostics, the cycle exemption, the tie-break, the undetermined closure, and the empirical zero-Primary-movement verification

## Decisions Made
See `key-decisions` in frontmatter above for the full rationale on each; summarized:
1. The resolver is built and tested but not wired into the eight existing emission sites (D-13-04 makes this a verified no-op today)
2. `classifyDeclaredCause`'s vocabulary is deliberately scoped to D-13-02's five named fields, not every cause kind the package emits
3. Non-`Callable` callee settled as B2 (caller misuse), recorded at the emission site
4. `checkCallGraphAcyclic` routed through `calleeBeforeCallerOrder` to make the "exactly one call site" acceptance criterion literally true
5. Two codes with no corpus fixture use inline sources; the structurally-unreachable code uses the existing seam convention

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `checkCallGraphAcyclic`'s own pre-existing `callgraph.Order` call was a second call site**
- **Found during:** Task 1, running the plan's own acceptance-criteria grep
- **Issue:** The plan's acceptance criterion states `grep -n 'callgraph\.Order(' internal/compiler/check/check.go | grep -v '//'` must report exactly one non-comment call site, but `checkCallGraphAcyclic` (pre-existing, for cycle detection) already called `callgraph.Order` directly, independent of the reversal `calleeBeforeCallerOrder` introduces. The plan's `read_first` section only named `buildInterproceduralSummaries`' own reversal as the concern (13-RESEARCH.md Code Example 4); it did not flag this second, orthogonal call site.
- **Fix:** Routed `checkCallGraphAcyclic` through `calleeBeforeCallerOrder` too (it only ever consults the returned error, never the order value, so this is behavior-preserving — the propagated error is byte-identical either way). This makes `calleeBeforeCallerOrder` check.go's sole `callgraph.Order` caller, satisfying the criterion literally and further strengthening D-13-03's "no new ordering authority" claim.
- **Files modified:** internal/compiler/check/check.go
- **Verification:** `grep -n 'callgraph\.Order(' internal/compiler/check/check.go | grep -v '//' | wc -l` now reports `1`; full `go test ./internal/compiler/check/...` and `go test ./...` both green.
- **Committed in:** 6967f1b (Task 1 commit)

**2. [Rule 1 - Bug] `resolveBlame`'s B3 tie-break comparator ranked an unordered candidate first**
- **Found during:** Task 2, running `TestBlameTieBreakIsCalleeBeforeCaller`'s defensive-path assertion
- **Issue:** The sort comparator's fallback branch (`declarationIndex[left] < declarationIndex[right]`) was reached whenever either candidate was absent from `order`, and Go's zero-value default (`0`) for a map miss made an absent candidate sort BEFORE a present one whenever the present one's declaration index was `> 0` — the opposite of the intended "ordered candidate always outranks an unordered one" rule.
- **Fix:** Added an explicit `if leftOK != rightOK { return leftOK }` branch before the declaration-index fallback.
- **Files modified:** internal/compiler/check/check.go
- **Verification:** `TestBlameTieBreakIsCalleeBeforeCaller` passes, including its defensive-path assertion.
- **Committed in:** eb8c22f (Task 2 commit)

**3. [Rule 1 - Bug] `resolveBlame`'s undetermined-detection sentinel was defeated by an empty `FunctionID`**
- **Found during:** Task 2, running `TestBlameUndeterminedPublishesBothSitesUnapplied`
- **Issue:** The original implementation used `undeterminedFunctionID == ""` as its own "not yet found" sentinel, but a genuinely-unclassified `blameFact` (produced when `classifyDeclaredCause` recognizes a cause's KIND but fails to parse its Detail) carries `FunctionID: ""` by construction — so the sentinel check silently treated "found, but empty" the same as "not found", and `resolveBlame` fell through to `blameCaller` (B2) instead of `blameUndetermined`. This was the exact fail-open failure mode D-13-07 exists to prevent, caught by the resolver's own test before ever reaching production.
- **Fix:** Replaced the string-sentinel with an explicit `undeterminedFound bool`.
- **Files modified:** internal/compiler/check/check.go
- **Verification:** `TestBlameUndeterminedPublishesBothSitesUnapplied` passes; `blame_undetermined` now fires correctly.
- **Committed in:** eb8c22f (Task 2 commit)

**4. [Rule 1 - Bug] Task 3's independent Primary-span derivation used the wrong AST field for non-call bindings**
- **Found during:** Task 3, first run of `TestBlameMovesNoPrimarySpanToday`
- **Issue:** The independent span-computation helper initially returned `Binding.Span` (the whole statement, e.g. `let delivered = take buffer`), but `check.go` itself records `spanByOperationID[operation.ID] = binding.RHS.Span` (just the RHS token, e.g. `take buffer`) for every binding kind — so the test's own "independent" expectation was wider than the diagnostic's actual Primary and failed on `check.interprocedural_loan_liveness`'s own real fixture.
- **Fix:** Changed `findBindingSpan` to return `binding.RHS.Span`, matching what check.go's own emission sites actually record (still an independent re-scan of the parsed AST, not a read-back from the diagnostic or from `spanByOperationID` itself).
- **Files modified:** internal/compiler/check/check_blame_test.go
- **Verification:** `TestBlameMovesNoPrimarySpanToday` passes for all eight codes.
- **Committed in:** 3a13274 (Task 3 commit)

---

**Total deviations:** 4 auto-fixed (4 Rule 1 bugs, all caught by this plan's own tests before landing)
**Impact on plan:** All four fixes were necessary for the resolver's own correctness claims (the fail-open closure genuinely closing, the tie-break genuinely tie-breaking) and for the plan's own acceptance criteria to hold literally. No scope creep.

## D-13-07 Compile-Time Exhaustiveness Guard — Scratch Deletion Demonstration

Per the plan's explicit instruction, the guard was demonstrated this session by deleting one entry (`declaredFieldAbilities: {},`) from `blameFieldWitness`'s array literal in a scratch edit, running `go build ./...`, and reverting:

```
internal/compiler/check/check.go:634:61: cannot use [...]struct{}{…} (value of type [4]struct{}) as [5]struct{} value in variable declaration
```

The build failed with a genuine Go type error (array-length mismatch), not merely a test failure — confirmed the guard is real, then the scratch edit was reverted (`diff` against the pre-edit file confirmed byte-identical restoration) before any commit.

## Issues Encountered
None beyond the four deviations above, all caught and fixed by this plan's own tests before any commit landed.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- The blame resolver (`resolveBlame`/`resolveCycleBlame`/`classifyDeclaredCause`) exists, is compile-time exhaustive, and is empirically confirmed to reproduce all eight shipped codes' existing B2 behavior — ready for plans 13-03/13-05/13-06 to introduce the first genuinely B1-shaped defect class without needing to build any of this scaffolding themselves.
- `calleeBeforeCallerOrder` is the package's sole `callgraph.Order` caller; any future plan needing callee-before-caller order should call it rather than reversing `callgraph.Order` inline again.
- `classifyDeclaredCause`'s two-shape vocabulary (callee_return_contract, callee-on-callee_not_callable) is deliberately narrow — a future plan introducing a NEW declared-contract-field cause (e.g. for Parameters[].Drops or Abilities) must add a new case there, and per D-13-07 that new case is what the compile-time exhaustiveness guard exists to keep honest.
- No blockers for 13-03.

## Self-Check: PASSED

- `internal/compiler/check/check.go`, `internal/compiler/check/check_blame_test.go` exist and carry the described changes (`git show 6967f1b`, `eb8c22f`, `3a13274`).
- `git log --oneline --all --grep="13-02"` returns this plan's metadata commit once created; the three task commits (`6967f1b`, `eb8c22f`, `3a13274`) are present in `git log --oneline -5`.
- All plan-level `<verification>` commands re-run clean: `go build ./... && go vet ./internal/compiler/check/...`; `go test ./internal/compiler/check/... -run TestBlame -v -count=1` (10/10 pass); the scratch deletion of one declared-contract-field arm fails `go build` with the exact error quoted above, reverted; `grep -rn 'blame' internal/compiler/corevalidate/ --include='*.go' | grep -v '_test.go' | grep -v '//'` returns zero matches; `go test ./...` green.

---
*Phase: 13-agent-loop-for-interprocedural-defects*
*Completed: 2026-09-13*

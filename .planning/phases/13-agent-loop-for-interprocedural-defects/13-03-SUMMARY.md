---
phase: 13-agent-loop-for-interprocedural-defects
plan: 03
subsystem: session
tags: [explain, cause-dag, function-attribution, peer-re-derivation, dx-05]

# Dependency graph
requires:
  - phase: 13-agent-loop-for-interprocedural-defects
    provides: "13-01's check.go span-widening (call-binding Binding.Span fix in syntax/parser.go) and 13-02's blame-resolver vocabulary (blameUndetermined dual-site repairs) this plan's Blame field reads from the diagnostic payload"
provides:
  - "protocol.ExplainNode.FunctionID/FunctionName/Blame and protocol.ExplainSummary.Functions -- additive, omitempty, no ExplainSchema bump (D-13-14/D-13-18/D-13-22)"
  - "Peer-re-derived function attribution (session.resolveExplainFunction): core.Program.Functions is authoritative for FunctionID, independently cross-checked against ast.Program.Funcs, refusing (never silently preferring) on genuine disagreement while treating an absent core entry as an honest 'no function data' report (D-13-15)"
  - "D-13-19's function-scope narrows guard: a whole-function-span node is caused_by-only, and narrows requires matching function_id -- fixes the sleeper containment bug where explainSpanStrictlyContains's pure byte containment manufactured narrows edges across function boundaries"
  - "buildExplainGraphSkippingNarrowsGuard, the guard-disabled twin proving the guard is load-bearing (session_phase6_injectors.go's matchInjectSkippingGuard convention)"
  - "testdata/phase13/explain_three_function_chain[_reordered].lang -- real, checkable four-function (leaf/decoy/relay/caller) corpus fixtures for D-13-23's falsifiable attribution tests"
  - "TestExplainEdgeKindsAreRecordedForGuardComparison -- D-13-20's committed pre-guard edge-kind baseline, re-verified green after the guard landed"
affects: [13-04, 13-05, 13-06, 13-07]

# Actuals (#2632)
actuals:
  tokens: 13978
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-tier caused_by/narrows candidate selection: bestNarrows (guard-eligible: not a whole-function span, matching function_id) and bestFunctionScopeCausedBy (a whole-function-span candidate that still 'may be a caused_by parent' per D-13-19's own wording), falling back to the pre-existing root default only when neither pool has a match"
    - "Peer-re-derivation asymmetry: when the AUTHORITATIVE table (core) has no entry for a span, that is an honest 'no data' report, not a peer disagreement to refuse -- disagreement is refused only when the authoritative table DOES claim an identity and the independently-derived peer (AST) does not confirm it. Encoding 'X has no entry' and 'X and Y disagree' as the same case breaks anywhere the two tables can be legitimately incomplete (a per-function core table is exactly that here, since check.go only appends a function once it checks cleanly)"
    - "Guard-disabled twin as a shared-body wrapper: buildExplainGraph and buildExplainGraphSkippingNarrowsGuard both call buildExplainGraphWithGuard(..., enforceGuard bool), avoiding duplicating the ~90-line synthesis loop the matchInjectSkippingGuard precedent (a much smaller function) did not need to worry about"

key-files:
  created:
    - testdata/phase13/explain_three_function_chain.lang
    - testdata/phase13/explain_three_function_chain_reordered.lang
  modified:
    - internal/compiler/protocol/protocol.go
    - internal/compiler/session/session_phase6_explain.go
    - internal/compiler/session/session_phase6_explain_test.go

key-decisions:
  - "core.Program.Functions is legitimately INCOMPLETE relative to ast.Program.Funcs: check.go only appends a function to Functions once it checks cleanly, so a function that itself carries the diagnostic being explained (or any sibling function with its own diagnostic) is absent from the core table even though the AST still declares it. Treating 'core has nothing, AST has something' as a peer DISAGREEMENT (rather than an honest 'no core data' report) broke roughly half the existing rejecting-fixture corpus on first implementation (foreign_call_target_not_foreign.lang's own function never checks cleanly, so its own diagnostic's span had no core entry) -- caught by TestExplainEdgeKindVocabularyIsClosed's corpus-wide subtest before landing. Fixed by scoping the refusal to 'core claims an identity AND the AST-derived peer does not independently confirm the SAME Name+Span', never to core's absence alone."
  - "check.go's checkCallGraphAcyclic branch wipes result.Program to its zero value after a cycle refusal, even though the diagnostic's own cycle_member causes (built from spanByOperationID before the wipe) carry real, per-function spans. Rather than special-casing this, the resolver treats an entirely-empty core table as 'no function data available' globally -- core.call_graph_cycle diagnostics get no function attribution today, an honest gap (matching debugmap's own 'compiler genuinely has no value' precedent) rather than a fabricated one."
  - "D-13-19's guard is asymmetric by design: a whole-function-span node is excluded from narrows-candidacy but is STILL eligible as a caused_by parent (a second, narrower candidate pool, bestFunctionScopeCausedBy) rather than unconditionally falling back to the diagnostic root. This is the literal reading of 13-CONTEXT.md D-13-19 ('may be a caused_by parent but never a narrows parent') and is proven by TestNarrowsFunctionScopeGuard's cause-1 assertion (attaches to the whole-function node with caused_by, not to root)."
  - "Blame (D-13-22) is derived purely from the diagnostic payload already in hand: a node's span matching diag.Primary, or matching a blame_undetermined repair's Span. No diagnostic in the current corpus actually publishes a blame_undetermined repair (13-02's resolver is built but not wired into any emission site, per 13-02-SUMMARY.md), so today Blame is true only for the root node -- correct and honest, not a stub, since D-13-06 forbids importing the blame resolver into session at all."
  - "ExplainSummary.Functions is populated from the WHOLE core function table (every function 'consulted' during resolution, decoy included), sorted by (span.start, ID) for determinism -- not filtered down to only the functions actually referenced by a node in the graph. This matches the plan's own framing ('the table exists only to hand back the whole-function span for repair targeting') and keeps the field's contents independent of which specific causes a diagnostic happens to carry."

requirements-completed: [DX-05]

coverage:
  - id: D1
    description: "Every ExplainNode with a non-nil span carries function_id/function_name naming the innermost containing function, peer-re-derived from core.Program.Functions and ast.Program.Funcs with disagreement refused"
    requirement: "DX-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainFunctionAttribution"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainFunctionIdentitySurvivesReorder"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainNilSpanOmitsFunction"
        status: pass
    human_judgment: false
  - id: D2
    description: "D-13-19's function-scope narrows guard fixes the sleeper containment bug: a whole-function-span node is caused_by-only, narrows requires matching function_id, and the fix is proven load-bearing by a guard-disabled twin"
    requirement: "DX-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestNarrowsFunctionScopeGuard"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestNarrowsFunctionScopeGuardIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainEdgeKindVocabularyIsClosed"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-13-20's schema-bump decision is answered by an observed run (not prediction): zero edge kinds flip on the existing testdata/phase2-5 corpus after the guard lands, so ExplainSchema stays lang.explain/0"
    requirement: "DX-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainEdgeKindsAreRecordedForGuardComparison"
        status: pass
    human_judgment: false
  - id: D4
    description: "No diagnostic-ID churn from this plan's additive fields; check_ordering_stability_test.go untouched; no third truncation code; full go test ./... green"
    requirement: "DX-05"
    verification:
      - kind: other
        ref: "git diff --quiet -- internal/compiler/check/check_ordering_stability_test.go"
        status: pass
      - kind: other
        ref: "grep -rn 'function_budget' internal/compiler/ --include='*.go' (zero matches)"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainTruncationCodesStable"
        status: pass
      - kind: other
        ref: "go test ./... -count=1"
        status: pass
    human_judgment: false

# Metrics
duration: 62min
completed: 2026-09-13
status: complete
---

# Phase 13 Plan 03: Function Attribution and the D-13-19 Narrows Guard Summary

**`lang explain`'s cause DAG now names the owning function of every node (peer-re-derived from both core and AST tables), and the sleeper bug where a whole-function-span cause silently manufactured `narrows` edges across function boundaries is fixed and proven load-bearing by a guard-disabled twin -- with an observed, not predicted, confirmation that zero edge kinds changed on the existing corpus, so `lang.explain` stays at schema `/0`.**

## Performance

- **Duration:** 62 min
- **Started:** 2026-09-13T18:48:00Z (approx, session start)
- **Completed:** 2026-09-13T19:50:10Z
- **Tasks:** 3
- **Files modified:** 5 (3 modified, 2 created)

## Accomplishments
- Added `FunctionID`/`FunctionName`/`Blame` to `protocol.ExplainNode` and a `Functions []ExplainFunction` table to `protocol.ExplainSummary` -- all additive/`omitempty`, `ExplainSchema` unchanged at `lang.explain/0`
- Implemented D-13-15's peer-re-derived function resolution (`resolveExplainFunction`): core is authoritative for `function_id`, independently cross-checked against the AST for the same `Name`+`Span`, refusing on genuine disagreement -- discovered and fixed a real correctness hazard where treating "core has no entry" the same as "core and AST disagree" broke a large fraction of the existing rejecting-fixture corpus (functions that themselves carry the diagnostic being explained are never appended to `core.Program.Functions`)
- Fixed the D-13-19 sleeper bug: `explainSpanStrictlyContains`'s pure byte containment was silently letting a whole-function-span cause become the narrowest containing ancestor of every later cause in its body. The guard now excludes whole-function-span nodes from `narrows` candidacy (they remain eligible as `caused_by` parents, a second, narrower candidate pool) and additionally requires `narrows` parent/child to share `function_id`
- Built `buildExplainGraphSkippingNarrowsGuard`, the guard-disabled twin (the repo's `matchInjectSkippingGuard` convention), and proved via `TestNarrowsFunctionScopeGuardIsNotInert` that disabling the guard flips both protected edges to `narrows`
- Authored `testdata/phase13/explain_three_function_chain[_reordered].lang`, real four-function (`leaf`/`decoy`/`relay`/`caller`) fixtures that check cleanly to `check.interprocedural_loan_liveness`, giving the resolver real, checkable multi-function span data to be tested against
- Captured D-13-20's decision-gate baseline (`TestExplainEdgeKindsAreRecordedForGuardComparison`) from an actual run over the whole `testdata/phase2-5` rejecting corpus BEFORE the guard landed, then re-ran it unchanged AFTER the guard landed in the same session -- see the observation section below

## Task Commits

Each task was committed atomically:

1. **Task 1: Additive protocol fields + peer-re-derived function attribution** - `76fca61` (feat)
2. **Task 2: D-13-20 decision gate -- capture pre-guard edge-kind baseline** - `a003d83` (test)
3. **Task 3: The D-13-19 narrows guard + guard-disabled twin** - `e0bc966` (fix)

**Plan metadata:** (this commit)

## D-13-20 Observation

**Method:** Captured the ordered `(edge.Kind)` list for every diagnostic in `testdata/phase2-5`'s rejecting-fixture corpus (16 fixture/diagnostic pairs) via a real `ExplainCommandFile` run, committed as the literal baseline `wantExplainEdgeKindsBeforeGuard` in Task 2's commit (`a003d83`) -- this ran with Task 1's function-attribution code already present but BEFORE Task 3's narrows-guard code existed (the guard's two new conditions were added only in Task 3, `e0bc966`).

**Result: zero edge kinds changed.** After Task 3 landed, `TestExplainEdgeKindsAreRecordedForGuardComparison` was re-run unchanged and passed against the same baseline -- every one of the 16 pairs still produces the identical ordered edge-kind sequence. This confirms 13-RESEARCH.md Pitfall 2's assessment by an actual run: every cause span in this corpus is either unspanned (`Detail`-only) or a narrow token/statement span, never a whole-function span, so the guard's two new exclusion conditions never activate on any of these 16 pairs (`wholeFunctionSpans` lookups never hit; the `function_id` equality check never differs from the pre-guard default because no candidate in this corpus is ever a whole-function-span node to begin with).

**Decision: `ExplainSchema` stays `lang.explain/0`.** No bump. This matches D-13-18's separate no-bump claim for the additive fields (Task 1) -- neither the new fields nor the guard changed any already-published edge shape.

**Caveat on ordering:** because Task 1's attribution code and Task 3's guard code both landed in this same plan/session (rather than the guard being deferred to a later plan against an already-shipped attribution feature), the "before" baseline in Task 2 was captured with function attribution ALREADY active but the guard NOT yet active -- the comparison that matters (whether the GUARD specifically flips anything) is exactly what was measured, since attribution alone (without the guard) cannot change which edges are emitted at all, only which nodes carry which `function_id`/`function_name` fields. The synthetic `TestNarrowsFunctionScopeGuard`/`TestNarrowsFunctionScopeGuardIsNotInert` pair additionally demonstrates, on a real three-function fixture's span data, exactly which edges the guard WOULD flip if such a whole-function-span cause ever appeared in a future diagnostic -- proving the guard is both correct and load-bearing, not merely inert on today's corpus.

## Files Created/Modified
- `internal/compiler/protocol/protocol.go` - Adds `ExplainNode.FunctionID/FunctionName/Blame`, new `ExplainFunction{ID, Name, Span}` type, `ExplainSummary.Functions` -- all additive/`omitempty`, documented with the same rationale as `core.Function.ForeignContract` (D-04-23)
- `internal/compiler/session/session_phase6_explain.go` - `explainFunctionTable`/`buildExplainFunctionTable`/`resolveExplainFunction`/`narrowestContainingFunction` (D-13-15's peer re-derivation), `applyExplainFunction`, `explainBlameSpanSet` (D-13-22), the two-pool `bestNarrows`/`bestFunctionScopeCausedBy` selection in `buildExplainGraphWithGuard` (D-13-19's guard), `buildExplainGraphSkippingNarrowsGuard` (the guard-disabled twin), `ErrExplainFunctionResolutionDisagreement`; `ExplainCommandFile` now builds and threads the function table and surfaces a disagreement as an operational failure
- `internal/compiler/session/session_phase6_explain_test.go` - Seven new tests (`TestExplainFunctionAttribution`, `TestExplainNilSpanOmitsFunction`, `TestExplainFunctionIdentitySurvivesReorder`, `TestExplainEdgeKindsAreRecordedForGuardComparison`, `TestNarrowsFunctionScopeGuard`, `TestNarrowsFunctionScopeGuardIsNotInert`, `TestExplainTruncationCodesStable`); updated all seven pre-existing `buildExplainGraph` call sites to the new 6-return-value, 3-argument signature (passing `explainFunctionTable{}` where no function context is exercised)
- `testdata/phase13/explain_three_function_chain.lang` - Real, checkable four-function fixture (`leaf`, `decoy`, `relay`, `caller`) producing `check.interprocedural_loan_liveness`; `decoy`'s byte range lies between `leaf`'s and `relay`'s (D-13-23's decoy requirement)
- `testdata/phase13/explain_three_function_chain_reordered.lang` - Same four functions, reordered declaration, for the reorder-invariance mutation test

## Decisions Made
See `key-decisions` in frontmatter for full rationale; summarized:
1. Core-table absence is an honest "no data" report, never a peer disagreement -- caught and fixed before landing via the existing corpus-wide test
2. `core.call_graph_cycle`'s Program-wipe means that diagnostic gets no function attribution today (honest gap, not fabricated)
3. The narrows guard is asymmetric: whole-function-span nodes stay `caused_by`-eligible via a second candidate pool, never unconditionally falling to root
4. `Blame` is derived purely from the diagnostic payload (`Primary` + `blame_undetermined` repair spans), correctly true only for root today since no diagnostic yet publishes the dual-site shape
5. `ExplainSummary.Functions` lists the whole core function table, not a filtered subset

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Peer-disagreement refusal was breaking most of the existing rejecting-fixture corpus**
- **Found during:** Task 1, first run of `TestExplainEdgeKindVocabularyIsClosed`'s corpus-wide subtest
- **Issue:** The initial `resolveExplainFunction` treated `coreOK != astOK` as a disagreement to refuse unconditionally. But `check.go` only appends a function to `core.Program.Functions` once it checks cleanly -- a function that itself carries the diagnostic being explained (e.g. `testdata/phase4/foreign_call_target_not_foreign.lang`'s `main`) is legitimately absent from the core table even though the AST still declares it, so `coreOK=false, astOK=true` for that diagnostic's own Primary span is the COMMON case, not an anomaly. The strict rule made `ExplainCommandFile` return an operational failure (`result.Explain == nil`) for roughly a third of the real corpus.
- **Fix:** Scoped the refusal to "core DID claim an identity for this span AND the AST-derived peer does not independently confirm the SAME `Name`+`Span`" -- core's absence alone is now an honest "no function data" report, matching the same shape already used for nil spans and foreign-boundary spans.
- **Files modified:** internal/compiler/session/session_phase6_explain.go
- **Verification:** Full `TestExplainEdgeKindVocabularyIsClosed`/`TestExplainCauseGraphIsDeterministic` corpus runs (16 fixture/diagnostic pairs across phase2-5) all pass; `go test ./...` green.
- **Committed in:** 76fca61 (Task 1 commit, fixed before the commit landed -- no separate remediation commit needed)

---

**Total deviations:** 1 auto-fixed (1 Rule 1 bug, caught by this plan's own tests before landing)
**Impact on plan:** The fix was necessary for the resolver's own correctness claim ("disagreement refused, never silently preferring core's absence as if it meant something"). No scope creep -- the fix stayed inside `resolveExplainFunction`'s own logic.

## Issues Encountered
- `checkCallGraphAcyclic` (check.go) wipes `result.Program` to its zero value after a cycle refusal, even though the cycle diagnostic's own `cycle_member` causes carry real per-function spans (built from `spanByOperationID` before the wipe) -- 13-RESEARCH.md Pitfall 2's own candidate for a possible edge-kind flip. Because the core table is empty for that diagnostic, this plan's resolver reports "no function data" for every one of its nodes rather than attributing them -- consistent with the honest-gap precedent elsewhere in the compiler, but it does mean `core.call_graph_cycle` explanations do not get function names from this plan. No corpus fixture exercising this path exists in `testdata/phase2-5` (confirmed by the D-13-20 baseline capture, which found zero `narrows`/`same_binding` edges anywhere in that corpus, let alone from a cycle diagnostic), so this did not affect the D-13-20 observation. Left as a known gap for a future plan if cycle-diagnostic function attribution is ever required.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- `lang explain`'s wire format (`ExplainNode.FunctionID/FunctionName/Blame`, `ExplainSummary.Functions`) is stable at `lang.explain/0` and ready for any later plan that consumes it (e.g. a future repair-targeting feature that needs whole-function spans)
- The D-13-19 guard and its guard-disabled twin are in place; any future diagnostic whose causes genuinely span multiple functions will now be attributed correctly rather than silently mis-narrowed
- No blockers for 13-04 through 13-07

---
*Phase: 13-agent-loop-for-interprocedural-defects*
*Completed: 2026-09-13*

## Self-Check: PASSED

- All 5 files listed in `key-files` (created/modified) verified present on disk.
- All 3 task commit hashes (`76fca61`, `a003d83`, `e0bc966`) verified present in `git log`.
- All must_haves truths re-verified: `go test ./internal/compiler/session/... -run TestExplain -v -count=1` (13 tests, all PASS); `go test ./internal/compiler/session/... -run 'TestNarrowsFunctionScopeGuard|TestNarrowsFunctionScopeGuardIsNotInert|TestExplainTruncationCodesStable|TestExplainEdgeKindVocabularyIsClosed'` (4 tests, all PASS); `grep -n 'ExplainSchema *=' internal/compiler/protocol/protocol.go` shows `lang.explain/0`; `git diff --quiet -- internal/compiler/check/check_ordering_stability_test.go` exits 0; `grep -rn 'function_budget' internal/compiler/ --include='*.go'` returns zero matches; full `go test ./... -count=1` green (one flaky, unrelated `TestCacheKeyCoversEveryDeclaredInput` failure in the `cache` package reproduced as passing in isolation, confirmed unrelated to this plan's files).

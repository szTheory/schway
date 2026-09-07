---
phase: 06-agent-feedback-and-performance-ratification
reviewed: 2026-09-07T05:04:40Z
depth: standard
files_reviewed: 25
files_reviewed_list:
  - internal/compiler/protocol/protocol.go
  - internal/compiler/diagnostic/diagnostic.go
  - internal/compiler/check/check.go
  - internal/compiler/cache/cache.go
  - internal/compiler/cache/probe.go
  - internal/compiler/measure/machine.go
  - internal/compiler/measure/statistics.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_mismatch.go
  - internal/compiler/session/session_phase5_sanitize.go
  - internal/compiler/session/session_phase6.go
  - internal/compiler/session/session_phase6_explain.go
  - internal/compiler/session/session_phase6_query.go
  - internal/compiler/session/session_phase6_risklanes.go
  - internal/compiler/session/session_phase6_verify.go
  - internal/compiler/session/session_phase6_evidence.go
  - internal/compiler/session/session_phase6_budget.go
  - internal/compiler/session/session_phase6_stages.go
  - internal/compiler/session/session_phase6_injectors.go
  - internal/compiler/session/session_phase6_escapes.go
  - cmd/lang/main.go
  - cmd/lang-repair/main.go
  - cmd/lang-repair/repair.go
  - scripts/verify-phase6.sh
findings:
  critical: 1
  warning: 2
  info: 0
  total: 3
status: issues_found
---

# Phase 6: Code Review Report

**Reviewed:** 2026-09-07T05:04:40Z
**Depth:** standard
**Files Reviewed:** 25
**Status:** issues_found

## Summary

This is a mature, carefully-engineered phase: fail-closed discipline is
genuinely load-bearing throughout (`cache.Consult`, `SelectLanesForFixture`,
`RatificationMode`, `phase6AddDeferredLane`, the five injectors' marker
guards), the anti-theater guards (prose-scramble, vocabulary-removal,
marker-mutation-kill) are real and demonstrably bidirectional rather than
theater themselves, and the schema/identity discipline (D-06-31/D-06-32) is
correctly enforced in `protocol.Result.Finalize()`. I traced the cache key
computation, the risk-lane selector, the budget-ratification gate, the
explain/query synthesis and pagination, and the repair driver's subprocess
and splice boundary looking for the specific failure shapes named in the
review brief (stale-green gates, permissive-on-ambiguity paths, cursor
pagination that drops/duplicates, cache escapes reported as pass). Nearly
everything held up under adversarial reading.

Two real gaps surfaced. One is a metrics-honesty bug in the actual CI-gated
code path (a counter that is structurally always zero despite the field
existing specifically to report a real count — this violates the phase's own
stated purpose more than most bugs would, since metric honesty is this
phase's explicit deliverable). The other is a decision (D-06-19's cold/warm
distribution reporting) that is only half-implemented: the cold-sample
machinery is declared but never wired into any collection path, and the
shared statistics helper cannot even summarize a cold-sized sample set if one
were collected. A third, lower-confidence item notes a latent (currently
unreachable) pagination fragility in `query`'s cursor scheme.

## Critical Issues

### CR-01: `cache_inputs_reused_count` is dead code in the actual CI-gated verify path — always reports 0 even on a real cache hit

**File:** `internal/compiler/session/session_phase6.go:153-155`
**Issue:** `VerifyPhase6ControlsAndWork` — the function `lang verify testdata/phase6`
actually dispatches to (via `isPhase6Corpus` in `cmd/lang/main.go`, and the
only phase-6 path `scripts/verify-phase6.sh` drives end-to-end) — records
`lane.CacheStatus` on the native-differential lane correctly, but then does
this instead of incrementing the metric:

```go
if lane.CacheStatus != "" {
    result.Metrics.CacheInputsReusedCount += 0 // status recorded on the lane itself; no double count here
}
```

This branch is a no-op: it adds zero regardless of whether `lane.CacheStatus`
is `artifact_reused`, `artifact_recomputed`, or anything else. Compare the
sibling path `VerifyPhase6ChangedRisk` (`session_phase6_verify.go:516-518`),
which does this correctly:

```go
if lane.CacheStatus == string(cache.StatusArtifactReused) {
    result.Metrics.CacheInputsReusedCount++
}
```

D-06-12 requires `protocol.Metrics.cache_inputs_reused_count` to be
"denominated in the same counted-unit currency as `recomputed_work`" — i.e.
it must actually count reuse. In `VerifyPhase6ControlsAndWork`, it never
will, on any input, including a genuine warm-cache run where the artifact was
reused. This is the one lane driven by the real, always-run CI gate
(`scripts/verify-phase6.sh`'s `phase6.json`), so the field this phase
introduced specifically to report cache honesty is silently wrong in
precisely the code path that ships. `internal/compiler/session/session_phase6_verify_test.go`'s
`TestCacheInputsReusedCountIsCounted` only exercises the sibling
`VerifyPhase6ChangedRisk` function, so this dead branch has no test coverage
at all.

**Fix:**
```go
if lane.CacheStatus == string(cache.StatusArtifactReused) {
    result.Metrics.CacheInputsReusedCount++
}
```
(mirroring the already-correct code in `session_phase6_verify.go`), plus a
test that asserts `VerifyPhase6ControlsAndWork`'s own
`Metrics.CacheInputsReusedCount` is nonzero on a warm second run, the same
shape `TestCacheInputsReusedCountIsCounted` already proves for the sibling
function.

## Warnings

### WR-01: D-06-19's cold-sample distribution is declared but never produced anywhere

**File:** `internal/compiler/measure/statistics.go:14-19`, `scripts/verify-phase6.sh:140-154`
**Issue:** D-06-19 requires: "20 warm samples, a small fixed n for cold
(cold state cannot be re-established in-process), reporting p50 and p95" —
i.e. both a warm *and* a cold distribution should be reportable. The code
declares `ColdSampleCount = 5` with a doc comment describing its intended
use, but:

1. No production code path anywhere in the tree (`cmd/lang/main.go`,
   `scripts/verify-phase6.sh`, or any `session_phase6_*.go` file) ever
   collects `ColdSampleCount` samples or calls `Summary()` on a cold-sized
   set. `scripts/verify-phase6.sh`'s `observe()` function drives only
   `phase6_warm_sample_count=20` samples through `lang stats`
   (`cmd/lang/main.go:194-220`, `runStats`).
2. Even if a caller collected 5 cold samples and piped them through `lang
   stats`, `Samples.Summary()` unconditionally refuses any sample set
   shorter than `WarmSampleCount` (20):
   ```go
   if len(s) < WarmSampleCount {
       return Summary{}, &Error{Code: "measure.stats_short_sample_set", ...}
   }
   ```
   There is no alternate code path that accepts `ColdSampleCount`-sized
   input — the one shared statistics helper structurally cannot summarize a
   cold distribution as designed. `ColdSampleCount` is referenced only in
   `internal/compiler/measure/statistics_test.go`, which tests the constant's
   value in isolation, never an actual cold-sampling flow.

This means the ratification story this phase ships only ever reports a warm
distribution; the cold half of D-06-19 is unimplemented despite being coded
as if it were (a declared, documented, but functionally inert constant).

**Fix:** Either wire an actual cold-sample collection loop (a fresh-process
invocation per sample, since "cold state cannot be re-established
in-process") into `scripts/verify-phase6.sh` and give `Samples.Summary()` (or
a sibling function) a path that accepts `ColdSampleCount`-sized input without
refusing it, or — if cold-distribution reporting is being deferred — say so
explicitly in the phase's debt/escape register (`session_phase6_escapes.go`)
rather than leaving an apparently-wired-up constant that nothing calls.

### WR-02: `query`'s cursor pagination uses strict-inequality tie-breaking that would silently drop ties, if any resolver ever produced them

**File:** `internal/compiler/session/session_phase6_query.go:500-543`
**Issue:** `paginateQueryFacts` finds the next page's start via
`sort.Search` on `queryFactAfterKey`, which reports whether a fact sorts
*strictly* after the cursor's key (`(vocabulary, span.start, span.end, id)`).
If two facts in the same result set ever compared equal on that full tuple
(same vocabulary, same span, same ID), and the page boundary fell between
them, the later duplicate(s) would never be selected on the next page — they
are not strictly greater than the cursor key, so `sort.Search`'s boundary
would skip them permanently. This is exactly the drop/duplicate failure
shape flagged as high-value in the review brief.

As implemented today, this is **not reachable**: every one of the five
vocabulary resolvers (`resolveQueryDiagnosticFacts`, `resolveQueryDebugMapFacts`,
`resolveQueryControlFacts`, `resolveQueryLaneFacts`, `resolveQueryEvidenceFacts`)
either returns exactly one fact, or (the diagnostic resolver) returns facts
whose IDs are index-suffixed (`fmt.Sprintf("%s:cause:%d", ...)`) and therefore
always unique within one call. So no current input can trigger the drop. It
is a latent fragility, not a live bug — flagged so a future sixth vocabulary
or a change to one of the five resolvers that could emit two facts with an
identical (vocabulary, span, id) tuple doesn't silently reintroduce a
data-loss bug in the pagination contract `TestQueryPagesConcatenateExactlyOnce`
is supposed to guarantee.

**Fix:** No change required now given the closed, currently-unique-ID
resolver set; if a future resolver could plausibly emit duplicate-keyed
facts, either dedupe before pagination or extend `queryCursorKey` with a
tie-breaking ordinal so `queryFactAfterKey` never needs strict equality to
mean "already emitted."

---

_Reviewed: 2026-09-07T05:04:40Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_

---
phase: 06-agent-feedback-and-performance-ratification
plan: 07
subsystem: verify-wiring
tags: [cache, risk-lanes, evidence, verify, dx-tooling, go-ast-structural-tests]

# Dependency graph
requires:
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-04"
    provides: "internal/compiler/cache -- Consult/Outcome/CacheStatus/Store/Put/ArtifactSpec/InputsFor, the artifact cache this plan wires into a real verify path"
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-05"
    provides: "session.SelectLanesForFixture/SelectLanes/LiveLaneIDs/ReasonSelected/ReasonDeferred/ReasonWidened, the change-risk selector this plan consumes directly"
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-06"
    provides: "protocol.LaneSchema1 and the seven /1 reporting fields (cache_status, selection_reason, cache_inputs_reused_count, ...) this plan populates for the first time"
provides:
  - "session.VerifyPhase6ChangedRisk -- a real verify path wiring the change-risk selector and the artifact cache together: lane:native-differential is selected by changed risk, its compiled artifact is cached/reused via cache.Consult/Put, and the checker plus the interpreter-vs-native comparator ALWAYS re-run fresh outside the cache branch (D-06-06), proven by a go/ast structural test rather than review"
  - "protocol.StatusDeferred and protocol.LaneStatuses() -- the closed lane-status vocabulary; a lane not run this invocation renders deferred, structurally never pass (phase6AddDeferredLane hardcodes the status and takes no status parameter)"
  - "protocol.TraceSummary/TraceEntry and protocol.EvidenceSummary.Trace -- session.ValidateEvidenceExpanded validates the compact manifest by default and expands a per-input recorded-vs-recomputed digest trace on validation failure or the new `lang evidence --validate --expand` flag, bounded at 64 KiB-plus-one with a stable truncated:evidence.trace_bound code"
  - "protocol.ValidateLaneVocabularies's selection_reason check widened to accept session.Selection.Reasons's own '<vocab>: detail' form, not only the bare vocabulary word"
affects: ["06-15 (final-gate ratification of DX-03/FND-04)", "phase-6-verification"]

# Actuals (#2632)
actuals:
  tokens: 20300
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Sibling-file convention extended a third time (session_phase6_verify.go/session_phase6_evidence.go alongside session_phase5*.go, session_phase6_risklanes.go) -- session.go and session_phase5.go stay untouched throughout this plan"
    - "go/ast structural proof that an assertion sits outside a cache-outcome branch (TestNativeDifferentialAssertionOutsideCacheBranch), extending the project's existing go/ast-structural-invariant convention (corevalidate.TestValidatorImportsStayIndependent, cache.TestCacheExportedSurfaceStoresNoVerdict) to control-flow shape, not just import/identifier scanning"
    - "Test-only override seams (Phase6CacheRootOverrideForTest/Phase6ChangeStatePathOverrideForTest/Phase6FixtureKindOverrideForTest) mirroring Phase5ClangPathOverrideForTest's precedent, since VerifyPhase6ChangedRisk's plan-specified signature has no room for a cache-root, state-path, or kind override parameter"
    - "A minimal, local compile-and-run pair (phase6CompileBinary/phase6RunCompiledBinary) duplicated rather than reusing native.Runner.Run, because Run's monolithic shape always compiles internally and has no seam for 'run this binary I already have from the cache' -- and native.go is outside this plan's file scope"

key-files:
  created:
    - internal/compiler/session/session_phase6_verify.go
    - internal/compiler/session/session_phase6_verify_test.go
    - internal/compiler/session/session_phase6_evidence.go
    - internal/compiler/session/session_phase6_evidence_test.go
  modified:
    - internal/compiler/protocol/protocol.go
    - internal/compiler/protocol/protocol_test.go
    - internal/compiler/session/session_phase6_pin_test.go
    - cmd/lang/main.go

key-decisions:
  - "[Rule 1/3 deviation] protocol.ValidateLaneVocabularies's selection_reason check accepted ONLY the bare vocabulary word ('selected'/'deferred'/'widened'), but session.Selection.Reasons (06-05, already shipped) always appends explanatory detail ('selected: pure_match matched', 'deferred: no declared dependency', 'widened: undeclared-input risk'). Every real Lane this plan emits failed vocabulary validation until this was fixed. Widened the check (new unexported selectionReasonInVocabulary) to accept either the bare word or a '<vocab>: ' prefix, matching this plan's own acceptance criterion ('SelectionReason begins with one of the three declared reason prefixes') literally. Verified against 06-06's existing TestLaneVocabulariesAreClosed, which still rejects a genuinely out-of-vocabulary value like 'maybe'."
  - "[Rule 1/3 deviation] TestLaneSchemaLiteralSiteCountIsPinned (06-06, session_phase6_pin_test.go) pins an exact repo-wide total of protocol.LaneSchema1 references. This plan's 3 new lane:native-differential/addDeferredLane composite literals are brand-new production sites (never a /0 literal to migrate), not a half-landed bump, but the pin's original implementation could not distinguish the two. Updated the pinned map (added session_phase6_verify.go: 3) and total (12 -> 15), documenting in-line why these 3 are a different species from the original 12 the 06-06 bump moved."
  - "[Discretion] VerifyPhase6ChangedRisk fully wires exactly ONE lane (lane:native-differential) end-to-end through the cache; every other lane in scope for the fixture's classified kind renders deferred via phase6AddDeferredLane -- including lanes the selector itself marked 'selected' but this plan did not wire an executor for. This is honest, not evasive: Status=deferred means 'not run this invocation' regardless of why, while SelectionReason still carries the selector's own real reason, so a consumer can distinguish 'the selector deferred it' from 'the selector selected it but no runner exists yet for it in this function.' This keeps T-06-VERIFY-02's prohibition (never render not-run as pass) true for every lane this function does not execute, without requiring this single plan to reimplement all ~33 corpus lanes' full execution logic through the cache."
  - "[Discretion] TestEveryLiveLaneIsAccountedFor asserts the Result's lane-ID union equals the GLOBAL LiveLaneIDs() (33 lanes), achieved by forcing an unclassified fixture kind via a new Phase6FixtureKindOverrideForTest test seam -- routing every lane through D-06-11's widen-to-everything branch. On an ORDINARY classified corpus (pure_match), the accounting universe is that kind's own 5 registry rows, not the global 33; that is the selector's own established, tested (06-05) contract, not a gap this plan introduces."
  - "The mutation_runner_source declared input (D-06-07, always-required and non-empty) has no real per-mutant runner in this generic verify path the way NAT03Mutations does; phase6SelfSourcePath() uses runtime.Caller(0) to supply this file's own absolute path as an explicit, stable, cwd-independent placeholder rather than leaving it empty (which InputsFor refuses outright, D-06-11)."

requirements-completed: []

coverage:
  - id: D1
    description: "VerifyPhase6ChangedRisk selects lane:native-differential by real changed risk, compiles and caches its artifact, and always re-runs the checker and the interpreter-vs-native comparator fresh -- a cache hit changes only what is skipped, never what is asserted (D-06-06)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestVerifyPhase6ChangedRiskRunsOneSelectedLane"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestNativeDifferentialAssertionOutsideCacheBranch"
        status: pass
    human_judgment: false
  - id: D2
    description: "Every lane VerifyPhase6ChangedRisk emits carries a schema-1 lane with CacheStatus drawn from the closed four-value artifact vocabulary and SelectionReason beginning with one of the three declared prefixes"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestVerifyReportsCacheStatusVocabulary"
        status: pass
    human_judgment: false
  - id: D3
    description: "A lane that was not run renders status deferred and is structurally incapable of rendering pass; no lane silently goes missing from the Result"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestDeferredLaneNeverRendersPass"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestEveryLiveLaneIsAccountedFor"
        status: pass
    human_judgment: false
  - id: D4
    description: "cache_inputs_reused_count is counted in the same integer unit as recomputed_work and reflects artifacts actually reused this run; a cold run reports zero reused and never artifact_reused"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestCacheInputsReusedCountIsCounted"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestFirstColdRunReportsRecomputedNotReused"
        status: pass
    human_judgment: false
  - id: D5
    description: "The lane-status vocabulary this verify path emits is closed, and protocol.LaneStatuses() names it exhaustively"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_verify_test.go#TestLaneStatusVocabularyIsClosed"
        status: pass
    human_judgment: false
  - id: D6
    description: "evidence --validate returns the compact summary by default, expands the trace on failure or --expand, and expansion never changes the verdict (Status or diagnostic codes)"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceValidateIsCompactByDefault"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceExpandsTraceOnFailure"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceExpandsTraceOnRequest"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_evidence_test.go#TestExpansionNeverChangesTheVerdict"
        status: pass
    human_judgment: false
  - id: D7
    description: "The expanded evidence trace respects the existing 64 KiB-plus-one output bound and truncates with a stable code rather than growing unboundedly"
    requirement: "DX-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_evidence_test.go#TestEvidenceTraceRespectsOutputBound"
        status: pass
    human_judgment: false
  - id: D8
    description: "internal/compiler/session/session.go is unchanged by this plan"
    requirement: "DX-03"
    verification:
      - kind: other
        ref: "git diff --stat 612f7b5..HEAD -- internal/compiler/session/session.go (empty)"
        status: pass
    human_judgment: false

duration: 95min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 7: Cache and Risk-Lane Selector Wired Into a Real Verify Path Summary

**`VerifyPhase6ChangedRisk` selects `lane:native-differential` by real changed risk, compiles and caches its artifact through `cache.Consult`/`Put`, and always re-runs the checker and comparator fresh outside the cache branch (proven by a go/ast structural test); `evidence --validate` now validates compact by default and expands a bounded per-input digest trace on failure or `--expand`.**

## Performance

- **Duration:** ~95 min
- **Tasks:** 3 (Task 1 tracer, Task 2 TDD, Task 3 TDD)
- **Files created:** 4 (session_phase6_verify.go, session_phase6_verify_test.go, session_phase6_evidence.go, session_phase6_evidence_test.go)
- **Files modified:** 4 (protocol.go, protocol_test.go [self-check only, no new tests added], session_phase6_pin_test.go, cmd/lang/main.go)
- **Commits:** 3

## Accomplishments

- **Task 1 (tracer):** `VerifyPhase6ChangedRisk(ctx, corpus, runner)` classifies the fixture kind (mirroring `VerifyCorpus`'s own file-marker dispatch, duplicated rather than shared per the "session.go untouched" constraint), computes the seven declared inputs via `cache.InputsFor`, and calls `SelectLanesForFixture` (06-05) to get a real `Selection` plus the cold/warm state. For `lane:native-differential`, `verifyPhase6NativeDifferentialLane` runs the checker (`Check`) and builds the interpreter/native comparison inputs BEFORE ever consulting the cache; `cache.Consult` then resolves either a reused artifact (written to a temp executable via `phase6WriteExecutable`) or triggers a fresh `phase6CompileBinary` (a minimal local clang invocation, cached via `store.Put` when the outcome was recomputed) -- and the interpreter-vs-native comparator loop runs unconditionally AFTER that if/else, never nested inside it. `TestNativeDifferentialAssertionOutsideCacheBranch` is a `go/ast` scan proving this structurally: it flags every `if` whose condition mentions `outcome` and asserts no `execution.Equal` call falls inside one. `TestVerifyPhase6ChangedRiskRunsOneSelectedLane` proves a real corpus (`testdata/phase1`) run passes with `cache_status: artifact_recomputed` on a cold run, and a second run over the SAME cache root (with a fresh, deliberately cold change-state path so the lane is selected again rather than deferred) reuses the artifact (`artifact_reused`) while the lane still reports pass. `TestVerifyReportsCacheStatusVocabulary` confirms every emitted lane's `Schema`/`CacheStatus`/`SelectionReason` passes `protocol.ValidateLaneVocabularies`.
- **Task 2 (TDD):** `protocol.StatusDeferred` and `protocol.LaneStatuses()` (the closed lane-status vocabulary) were added to `protocol.go`. `phase6AddDeferredLane(result, id, reason, coldOrWarm)` hardcodes `protocol.StatusDeferred` and takes no status parameter at all -- structurally incapable of rendering pass. `TestDeferredLaneNeverRendersPass` proves both halves: the runtime property (a cold pure_match run always has at least one deferred lane, since only `lane:native-differential` is wired through this function) and the structural one (a `go/ast` scan confirms `phase6AddDeferredLane` never references `protocol.StatusPass` and has no `status` parameter). **Demonstrated once during execution:** `phase6AddDeferredLane`'s hardcoded status was temporarily changed to `protocol.StatusPass`, `TestDeferredLaneNeverRendersPass` was re-run and confirmed to FAIL (`no deferred lane observed on a cold pure_match run`), then the change was reverted and the diff confirmed clean (`git diff` empty) before re-running the full suite green. `TestEveryLiveLaneIsAccountedFor` forces an unclassified fixture kind via a new `Phase6FixtureKindOverrideForTest` seam, routing every one of `LiveLaneIDs()`'s 33 lanes through D-06-11's widen-to-everything branch, and asserts the Result's lane-ID union is exactly `LiveLaneIDs()` with no duplicates and no silent absence. `TestCacheInputsReusedCountIsCounted`/`TestFirstColdRunReportsRecomputedNotReused` prove the counted-unit and cold-run properties.
- **Task 3 (TDD):** `session_phase6_evidence.go`'s `ValidateEvidenceExpanded` is a sibling of the shipped `ValidateEvidenceCommandFile` (session.go untouched): it always attaches a compact `EvidenceSummary`, and attaches a `TraceSummary` (`phase6EvidenceTrace`, an independent recomputation of `evidence.Build` cross-checked against every digest-bearing manifest field: `source_digest`/`core_digest`/`c_digest`/`id`/`foreign_digest`/each `execution_digest[i]`) exactly when validation fails OR the new `expand` argument is true -- never on a plain passing, non-expand-requested call. `phase6BoundTraceEntries` bounds the trace at 64 KiB-plus-one, reporting `truncated:evidence.trace_bound` and stopping rather than growing unboundedly (`TestEvidenceTraceRespectsOutputBound` proves this with 5000 synthetic entries). `TestExpansionNeverChangesTheVerdict` runs the SAME stale manifest through both `expand=false` and `expand=true` and asserts identical `Status` and diagnostic code lists. `cmd/lang/main.go` gained `--expand` via a new `extractExpand` helper matching `extractJSON`'s own boolean-flag-stripping shape, wired into `evidence --validate`'s dispatch and `usageResult()`. Manually verified against the shipped binary: `lang evidence testdata/phase1/toggle.lang > manifest.json`, then `lang --json evidence --validate manifest.json testdata/phase1/toggle.lang` (compact, no `trace` key) and `lang --json evidence --validate --expand manifest.json testdata/phase1/toggle.lang` (a `trace` key with 4 entries, all `match: true`).
- Ran the full plan-level verification after every task: all three `<verify>` command blocks (via `scripts/assert-go-tests.sh`, `GOCACHE=/tmp/ai-lang-phase6-cache`) pass; `go build ./...`, `go vet ./...`, `go test ./...` (18 packages), and `go test -race ./internal/compiler/session/... ./internal/compiler/protocol/...` all clean; the shipped `lang` binary manually exercised for both `evidence` and `evidence --validate [--expand]`.

## Task Commits

1. **Task 1: One lane, selected by changed risk, run over a cached artifact, end-to-end** -- `a593b9a` (feat) -- session_phase6_verify.go, session_phase6_verify_test.go, protocol.go (StatusDeferred/LaneStatuses added here since phase6AddDeferredLane needed them from the start), session_phase6_pin_test.go (pin update deviation).
2. **Task 2: A lane that was not run is deferred, structurally never pass** -- `c5ff8bd` (test) -- session_phase6_verify_test.go additions only (the structural primitives already existed from Task 1's own implementation of `phase6AddDeferredLane`/`protocol.StatusDeferred`; this commit is the test proof, including the demonstrated-and-reverted mutation described above).
3. **Task 3: evidence validates compact, expands traces on failure or request** -- `a0bfab0` (feat) -- session_phase6_evidence.go, session_phase6_evidence_test.go, protocol.go (TraceSummary/TraceEntry/EvidenceSummary.Trace), cmd/lang/main.go (--expand flag).

(A fourth commit will follow this SUMMARY: the plan-completion docs commit.)

## Files Created/Modified

- `internal/compiler/session/session_phase6_verify.go` (new) -- `VerifyPhase6ChangedRisk`, `verifyPhase6NativeDifferentialLane`, `phase6AddDeferredLane`, `classifyPhase6FixtureKind`, `phase6ArtifactSpec`, `phase6CompileBinary`, `phase6WriteExecutable`, `phase6RunCompiledBinary`, `phase6SelfSourcePath`, `phase6Store`, `phase6ChangeStatePath`, plus the three test-only override vars.
- `internal/compiler/session/session_phase6_verify_test.go` (new) -- `TestVerifyPhase6ChangedRiskRunsOneSelectedLane`, `TestVerifyReportsCacheStatusVocabulary`, `TestNativeDifferentialAssertionOutsideCacheBranch`, `TestDeferredLaneNeverRendersPass`, `TestEveryLiveLaneIsAccountedFor`, `TestCacheInputsReusedCountIsCounted`, `TestFirstColdRunReportsRecomputedNotReused`, `TestLaneStatusVocabularyIsClosed`.
- `internal/compiler/session/session_phase6_evidence.go` (new) -- `ValidateEvidenceExpanded`, `phase6EvidenceTrace`, `phase6BoundTraceEntries`, `MaxEvidenceTraceBytes`, `TruncatedEvidenceTraceBound`.
- `internal/compiler/session/session_phase6_evidence_test.go` (new) -- `TestEvidenceValidateIsCompactByDefault`, `TestEvidenceExpandsTraceOnFailure`, `TestEvidenceExpandsTraceOnRequest`, `TestExpansionNeverChangesTheVerdict`, `TestEvidenceTraceRespectsOutputBound`.
- `internal/compiler/protocol/protocol.go` (modified) -- `StatusDeferred`, `LaneStatuses()`, `TraceEntry`, `TraceSummary`, `EvidenceSummary.Trace`, and `selectionReasonInVocabulary` (the ValidateLaneVocabularies fix).
- `internal/compiler/session/session_phase6_pin_test.go` (modified) -- pinned map/total updated (12 -> 15) to account for this plan's 3 new, legitimate lane-schema sites.
- `cmd/lang/main.go` (modified) -- `extractExpand`, `runEvidenceValidation` gained an `expand` parameter, `usageResult()` updated.

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: (1) the `ValidateLaneVocabularies` fix, without which every real lane this plan emits would fail its own vocabulary check; (2) scoping full lane execution to exactly one lane (`lane:native-differential`) while every other lane in scope renders honestly deferred, keeping the deferred-never-pass safety property true without requiring this plan to reimplement all ~33 corpus lanes' execution through the cache.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1/3 - Blocking] protocol.ValidateLaneVocabularies rejected every real SelectionReason this plan produces**
- **Found during:** Task 1's first `go test` run of `TestVerifyPhase6ChangedRiskRunsOneSelectedLane`
- **Issue:** `ValidateLaneVocabularies`'s `selection_reason` check used exact-string equality against `LaneSelectionReasons()` (`"selected"`/`"deferred"`/`"widened"`), but `session.Selection.Reasons` (06-05, already shipped) always appends explanatory detail (e.g. `"selected: pure_match matched"`). Every lane this plan emits failed vocabulary validation.
- **Fix:** Added `selectionReasonInVocabulary`, accepting either the bare word or a `"<vocab>:"` prefix.
- **Files modified:** `internal/compiler/protocol/protocol.go`
- **Verification:** `TestVerifyReportsCacheStatusVocabulary` passes; 06-06's own `TestLaneVocabulariesAreClosed` re-run and still rejects a genuinely out-of-vocabulary value (`"maybe"`).
- **Commit:** `a593b9a` (Task 1 commit)

**2. [Rule 1/3 - Blocking] TestLaneSchemaLiteralSiteCountIsPinned failed after Task 1's new file**
- **Found during:** Task 1's first whole-repo `go test ./...` run
- **Issue:** 06-06's pin counts every `protocol.LaneSchema1` reference repo-wide and asserts a fixed total of 12 across 4 named files. This plan's `session_phase6_verify.go` legitimately adds 3 new lane-schema composite literals (never a `/0` literal to migrate), which the pin could not distinguish from a half-landed bump.
- **Fix:** Updated the pinned per-file map (added `session_phase6_verify.go: 3`) and total (12 -> 15), with an in-line comment explaining these 3 are a different species from the original 12.
- **Files modified:** `internal/compiler/session/session_phase6_pin_test.go`
- **Verification:** `go test ./internal/compiler/session/...` clean; the pin still fails if any of the ORIGINAL 12 sites is reverted (not re-tested here, but the detection mechanism is unchanged from 06-06's own verified behavior).
- **Commit:** `a593b9a` (Task 1 commit)

---

**Total deviations:** 2 auto-fixed (both Rule 1/3, blocking issues on prior plans' shipped code discovered by this plan's own real usage). **Impact:** both necessary for this plan's tests to pass at all; neither weakens any prior plan's own structural guarantees (the vocabulary check still refuses genuinely out-of-vocabulary values; the pin still catches drift in the original 12 sites).

## Known Stubs

**Every other lane besides `lane:native-differential`, for the classified kind in scope, renders `deferred` rather than being executed by `VerifyPhase6ChangedRisk`.** This is a documented, scoped choice (see key-decisions), not a silent gap: `Status: deferred` accurately reports "not run this invocation" for every one of these lanes, `SelectionReason` still carries the selector's real reason, and no lane is ever rendered as a false pass. Wiring the remaining lanes' full execution through this cache-aware path (if ever desired) is future work, not claimed as done here. `VerifyPhase6ChangedRisk` is also not yet dispatched from `cmd/lang/main.go`'s `verify` CLI command -- that wiring was not in this plan's `files_modified` scope (only the `evidence --validate --expand` CLI surface was), and this plan's `<verification>` clause about "the shipped binary reports cache status..." was satisfied for the `evidence` half via a real shipped-binary run; the `verify`-side cache/selection reporting was verified via direct Go test invocation of `VerifyPhase6ChangedRisk`; equally real evidence, but not via the CLI's `verify` subcommand specifically.

## Flagged Assumption Carried Forward

Per the plan's own `<flagged_assumptions>`: the unresolved edge probe (DX-03, category `unclassified`) carried from 06-04/06-05/06-06 remains `unresolved`. Surfaced here, not dropped, per the plan's instruction. This plan's own scope (real verify wiring, deferred-never-pass, evidence expansion) does not resolve it.

## Issues Encountered

None blocking, beyond the two Rule 1/3 deviations documented above (both discovered and fixed within this plan's own execution, not carried forward as open issues).

## User Setup Required

None -- no external service configuration required. The artifact cache and change-state file are both created lazily under `os.UserCacheDir()` on first use (06-04/06-05's own precedent), unaffected by this plan.

## Next Phase Readiness

- `requirements.ready-ids` for this plan's requirement IDs (`DX-03`, `FND-04`) returned both `blocked` (06-15 also declares them and has not finished) -- confirming this plan correctly leaves `requirements-completed: []` in this SUMMARY's frontmatter, per the plan's own explicit instruction not to force these complete.
- `go build ./...`, `go vet ./...`, `go test ./...` (18 packages), and `go test -race ./internal/compiler/session/... ./internal/compiler/protocol/...` are all clean after this plan.
- `session.VerifyPhase6ChangedRisk`, `session.ValidateEvidenceExpanded`, `protocol.LaneStatuses`, `protocol.StatusDeferred`, and `protocol.TraceSummary` are all ready for 06-15 (the final-gate ratification plan) to consume or extend.
- No blockers for subsequent Phase 6 plans.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: internal/compiler/session/session_phase6_verify.go
- FOUND: internal/compiler/session/session_phase6_verify_test.go
- FOUND: internal/compiler/session/session_phase6_evidence.go
- FOUND: internal/compiler/session/session_phase6_evidence_test.go
- FOUND: internal/compiler/protocol/protocol.go (modified)
- FOUND: internal/compiler/session/session_phase6_pin_test.go (modified)
- FOUND: cmd/lang/main.go (modified)
- FOUND: commit a593b9a (Task 1)
- FOUND: commit c5ff8bd (Task 2)
- FOUND: commit a0bfab0 (Task 3)
- internal/compiler/session/session.go confirmed unchanged (git diff --stat empty)
- All plan `<acceptance_criteria>` re-verified passing (see coverage block above)
- Plan-level `<verification>` re-run: all three `<verify>` command sets pass via scripts/assert-go-tests.sh; `go test ./...`, `go test -race`, `go vet` all clean; shipped binary manually exercised for evidence/evidence --validate/--expand

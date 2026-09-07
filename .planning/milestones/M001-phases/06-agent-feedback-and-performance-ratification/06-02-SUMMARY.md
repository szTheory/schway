---
phase: 06-agent-feedback-and-performance-ratification
plan: 02
subsystem: cli-protocol
tags: [diagnostics, cause-graph, cli, schema-versioning, determinism]

# Dependency graph
requires:
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-01"
    provides: "Prior-phase byte/manifest/golden pin, D-06-32 identity-exclusion pin, and the 12-site lang.verify-lane/0 literal pin this plan built on top of without perturbing"
provides:
  - "lang.explain/0: a net-new, versioned command schema for a bounded, deterministic cause DAG synthesized from diagnostic.Diagnostic.Causes"
  - "protocol.Result.Explain (+ExplainID folded into Finalize()'s identity), mirroring the existing DebugMap/DebugMapID pattern"
  - "session.ExplainCommandFile: the CLI seam for `lang explain SRC ID [--depth=N] [--json]`, re-deriving the diagnostic from source on every cold invocation (no persisted store)"
  - "A closed 3-value edge-kind vocabulary (caused_by/narrows/same_binding) with depth and node-count bounding and stable truncation codes"
affects: [06-06-lane-schema-bump, phase-6-verification, dx-tooling-consumers-of-lang-explain]

# Actuals (#2632)
actuals:
  tokens: 9569
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "internal (package session, not session_test) test file for a synthesizer whose unexported helpers need direct construction of exact depth/span/binding shapes no real in-tree fixture happens to produce"
    - "dynamic corpus discovery (parse+check every .lang under testdata/<phase>, keep only the rejecting ones) rather than a hand-curated fixture list, matching the project's standing 'reachable input space must be the real corpus' rule"

key-files:
  created:
    - internal/compiler/session/session_phase6_explain.go
    - internal/compiler/session/session_phase6_explain_test.go
    - testdata/phase5/explain_use_after_move.lang
  modified:
    - internal/compiler/protocol/protocol.go
    - internal/compiler/protocol/protocol_test.go
    - cmd/lang/main.go

key-decisions:
  - "Task 0's checkpoint was pre-resolved at orchestration time (developer confirmed D-06-04 as locked: net-new lang.explain/0 and lang.query/0, not folded into lang.command/0, no /1 bump on their own account; grammar is `lang explain SRC ID [--depth=N] [--json]`). Recorded here, not re-litigated."
  - "buildExplainGraph's edge derivation, depth cutoff, and node budget were implemented as one cohesive unit in Task 1's commit (a graph builder incapable of bounding/typing edges would already be an incomplete implementation of the tracer's own real-fixture acceptance criteria); Task 2/3's tdd=\"true\" tasks discharged as pure test-addition commits over that already-correct implementation, following 06-01's own documented precedent for justified consolidation when tasks share one small file's helpers."
  - "session_phase6_explain_test.go is `package session` (internal), the only internal-package test file in this directory (every sibling is `package session_test`) -- Task 2/3's <behavior> bullets need to construct exact depth/span/binding shapes no real check.go Cause-construction site happens to produce (no two Cause entries in any real diagnostic share a Kind+Detail pair), so the synthesizer is exercised directly. Schema minting (Task 1) and cross-corpus determinism (Task 3) still drive session.ExplainCommandFile end-to-end against real fixtures."
  - "testdata/phase5 has zero check-time-rejecting fixtures (the phase 5 corpus tests runtime sanitizer/mutation properties, not checker rejection) -- confirmed by dynamically scanning every testdata/phase5/*.lang file rather than assumed. testdata/phase5/explain_use_after_move.lang was added: a minimal, honestly-labeled negative fixture (identical shape to testdata/phase2/use_after_move.lang, distinct module name) so the corpus-wide determinism sweep gets real phase5 coverage instead of silently having none."
  - "Edge-kind correlation keys on diagnostic.Cause{Kind, Detail} pairs restricted to the four Kind values (place/owner/loan/transfer_target) verified against every Cause construction site in check.go to actually carry a binding/place identity in Detail -- `type` and `declared_here`/`moved_here` do not correlate, since their Detail/Span do not name a reusable binding identity."

requirements-completed: [DX-02]

coverage:
  - id: D1
    description: "lang explain returns a bounded synthesized cause DAG addressed by a stable diagnostic ID under the net-new lang.explain/0 schema, without dumping the whole program"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainSummarySchemaIsMinted"
        status: pass
      - kind: e2e
        ref: "internal/compiler/testsupport/cli_phase6_test.go#TestExplainCLIReturnsBoundedCauseDAG"
        status: pass
    human_judgment: false
  - id: D2
    description: "The cause DAG is computed fresh per cold invocation and never persisted; two cold invocations on the same input produce byte-identical output"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainCauseGraphIsDeterministic"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainSynthesisOpensNoWritePath"
        status: pass
    human_judgment: false
  - id: D3
    description: "--depth defaults to 3, a hard node-count budget applies, and exceeding either truncates with a stable code rather than growing"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainRespectsDepthAndNodeBudget"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainTruncationCodeIsStable"
        status: pass
    human_judgment: false
  - id: D4
    description: "A zero-cause diagnostic returns a single root node and no edges (honest empty result); equal-comparing nodes have a specified, stable order across cold invocations"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainZeroCauseDiagnosticReturnsSingleNode"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainNodeOrderIsStableOnTies"
        status: pass
    human_judgment: false
  - id: D5
    description: "Node availability reuses the existing not_captured/optimized_out vocabulary rather than fabricating a value"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_explain_test.go#TestExplainSummarySchemaIsMinted"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-06
status: complete
---

# Phase 6 Plan 2: Mint `lang.explain/0` — Bounded, Deterministic Cause Graphs Summary

**`lang explain SRC ID [--depth=N] [--json]` synthesizes a bounded, deterministic cause DAG from a diagnostic's existing flat `Causes` list under the net-new `lang.explain/0` schema, with typed `caused_by`/`narrows`/`same_binding` edges and stable `truncated:explain.depth` / `truncated:explain.node_budget` codes.**

## Checkpoint Confirmation (Task 0)

Task 0 was a `checkpoint:decision` whose `<blocking>` was already `false` with a recorded `<resolution>` at orchestration time (developer confirmed 2026-09-06). Not re-litigated during execution; recorded here per the plan's instruction:

- **D-06-04 confirmed locked:** `explain` and `query` are net-new documents getting net-new `lang.explain/0` / `lang.query/0` schemas, each with a content-derived `id` following the existing `Finalize()` pattern. They do NOT fold into `lang.command/0` and do NOT trigger that constant's coordinated `/1` bump on their own account (that bump is 06-06's job in wave 4). Standing rule: a new capability gets a new `/0` schema; only extending an existing document's fields earns a `/1`.
- **Command grammar confirmed:** `lang explain SRC ID [--depth=N] [--json]`, matching the shipped `lang debug-map SRC [QUERY]` precedent. The `SRC` operand is required because D-06-02 forbids any persisted store or daemon a bare ID could resolve against — every invocation re-derives the diagnostic from source.

## Performance

- **Duration:** ~55 min
- **Tasks:** 3 (Task 0 confirmed, not executed as work)
- **Files modified:** 6 (3 created, 3 modified)

## Accomplishments

- `protocol.go`: minted `ExplainSchema = "lang.explain/0"`, the closed edge-kind vocabulary (`EdgeCausedBy`/`EdgeNarrows`/`EdgeSameBinding`), `ExplainDefaultDepth = 3`, `ExplainMaxNodes = 4096`, and the `ExplainNode`/`ExplainEdge`/`ExplainSummary` types. `Result.Explain *ExplainSummary` folds into `Finalize()` as a 24-hex-character `ExplainID`, mirroring `DebugMapID` exactly; `Metrics`/`Lane` were left untouched (06-06's territory, and `TestMetricsAndLaneFieldsExcludedFromIdentity`/`TestIdentityFieldEnumerationIsExhaustive` from 06-01 still pass unmodified).
- `session_phase6_explain.go`: `ExplainCommandFile(path, diagnosticID, depth)` re-parses and re-checks `path` on every call (no persisted state), locates the requested diagnostic among whichever diagnostics that pass produced, and calls `buildExplainGraph` to synthesize the DAG. `buildExplainGraph` attaches each `Cause` to the most specific correlated ancestor already in the graph: `same_binding` when two causes share a `place`/`owner`/`loan`/`transfer_target` Kind+Detail identity, `narrows` when a cause's span is strictly contained in an ancestor's span, `caused_by` otherwise. Exceeding `maxDepth` or `ExplainMaxNodes` stops expansion with a stable truncation code. Final output is sorted by `(depth, span.start, span.end, id)` for nodes and `(from, to, kind)` for edges.
- `cmd/lang/main.go`: added the `explain` dispatch arm (`runExplain`) and an `extractDepth` helper mirroring `extractJSON`'s strip-and-report shape for `--depth=N`; extended `usageResult()`'s grammar string.
- Verified end-to-end against the shipped binary: `lang --json explain testdata/phase2/use_after_move.lang <id>` emits `"schema":"lang.explain/0"` with a non-empty `nodes` array; `--depth=1` on the same diagnostic against a synthetic multi-level chain (unit-level) truncates with `truncated:explain.depth`.
- Determinism (D-06-02's obligation) discharged as an executable claim across every check-time-rejecting fixture found (dynamically) under `testdata/phase2` through `testdata/phase5` — 15 diagnostics across 11 fixtures, plus a new `testdata/phase5/explain_use_after_move.lang` fixture added specifically because phase 5's real corpus had zero check-rejecting fixtures on its own.
- Manually verified `buildExplainGraph`'s final `sort.SliceStable` is load-bearing, not decorative: constructed a diagnostic where raw insertion order is `[root(d0), cause0(d1), cause1(d2), cause2(d1)]` (cause2 attaches via `narrows` directly to root, landing at depth 1 even though it is inserted after cause1's depth-2 `same_binding` attachment). With the sort removed, `TestExplainNodeOrderIsStableOnTies` observably regresses from the expected `[root, cause0, cause2, cause1]` to the raw insertion order `[root, cause0, cause1, cause2]` — confirmed live during this task, then the sort was restored and the test re-verified green.

## Task Commits

Each task was committed atomically:

1. **Task 1: End-to-end `lang explain`** — `28301ae` (feat) — protocol.go, session_phase6_explain.go, cmd/lang/main.go, plus Task-1-scoped tests (`TestExplainSummarySchemaIsMinted`, `TestExplainDiagnosticNotFoundIsOperational`, `TestExplainFoldsIntoResultIdentity`).
2. **Task 2: Typed edges, depth and node budget, stable truncation** — `1ad092f` (test) — the full behavior test suite for edge derivation/bounding/truncation, over the already-implemented synthesizer from Task 1.
3. **Task 3: Discharge the determinism obligation** — `4a4b5c9` (test) — cross-corpus determinism test, the not-inert ordering pin, the non-persistence structural assertion, and the new `testdata/phase5` fixture.

_Note: Tasks 2 and 3 carried `tdd="true"` but produced single `test(...)` commits rather than RED/GREEN pairs. `buildExplainGraph`'s edge-typing, depth cutoff, and node-budget logic were implemented as one cohesive unit alongside Task 1's base graph builder (a builder that didn't bound/type edges would already fail Task 1's own real-fixture acceptance criteria), so there was no separate implementation step to add once Task 2/3's tests were written — mirroring 06-01's own documented precedent for justified single-commit consolidation when tasks share one small file's helpers._

## Files Created/Modified

- `internal/compiler/protocol/protocol.go` — `ExplainSchema`, `EdgeCausedBy`/`EdgeNarrows`/`EdgeSameBinding`, `ExplainDefaultDepth`, `ExplainMaxNodes`, `ExplainNode`/`ExplainEdge`/`ExplainSummary` types, `Result.Explain` field, `ExplainID` folded into `Finalize()`, `Explain` render block in `human()`.
- `internal/compiler/protocol/protocol_test.go` — `TestExplainFoldsIntoResultIdentity` (nil Explain omits the JSON key; non-nil Explain both serializes and perturbs `Result.ID`).
- `internal/compiler/session/session_phase6_explain.go` (new) — `ExplainCommandFile`, `ErrExplainDiagnosticNotFound`, `resolveExplainDepth`, `buildExplainGraph` and its correlation/span-containment/sort helpers.
- `internal/compiler/session/session_phase6_explain_test.go` (new) — 9 tests: schema minting, not-found handling, depth/budget/truncation, zero-cause empty result, closed edge-kind vocabulary (unit + real corpus), cross-corpus determinism, not-inert ordering, non-persistence structural check.
- `cmd/lang/main.go` — `explain` dispatch arm, `runExplain`, `extractDepth`, updated usage string.
- `testdata/phase5/explain_use_after_move.lang` (new) — minimal check-rejecting fixture giving the phase5 corpus real determinism-test coverage.

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: the internal (`package session`) test-file convention deviation, and the addition of a new phase5 testdata fixture, both explained above with full reasoning.

## Deviations from Plan

**1. [Discretion] Internal test package instead of external `session_test`.** `session_phase6_explain_test.go` is `package session`, while every other test file in the directory is `package session_test`. Required because `buildExplainGraph`/`resolveExplainDepth` are intentionally unexported per the plan's own artifact list (only `session.ExplainCommandFile` is named as new exported surface), and Tasks 2/3's `<behavior>` bullets need exact depth/span/binding shapes that no real in-tree `check.go` Cause-construction site happens to produce. Go permits both test-package styles to coexist in one directory; this was verified to compile and pass alongside every existing `session_test`-package file with no collisions.

**2. [Rule 2 - missing coverage] Added `testdata/phase5/explain_use_after_move.lang`.** Discovered while building the determinism test that `testdata/phase5`'s real corpus has zero check-time-rejecting fixtures (Phase 5's corpus exercises runtime sanitizer/mutation properties, never checker rejection) — confirmed by dynamically scanning every `.lang` file rather than assuming. Rather than silently having the cross-corpus determinism sweep produce zero phase5 coverage, added one minimal, honestly-labeled negative fixture (same shape as `testdata/phase2/use_after_move.lang`, distinct module name `owned.phase5_explain_use_after_move`) so `testdata/phase5` is genuinely swept, not skipped. Verified this addition does not perturb any existing phase5 fixture-count assumption: `TestPhase5CorpusIncludesEveryPriorPhase` and `TestPhase5AdversarialSubsetIsComplete` glob only `phase1`-`phase4` for their "no drift" checks; `TestPhase5CorpusThreeEngineAgreement` explicitly `t.Skipf`s any fixture the checker refuses (a reject-program is out of scope for that differential by design, D-05-20); `core_test.go`'s `pinnedFixtures` is a hardcoded `{path, sha}` slice unaffected by new directory contents; full `go test ./...` was re-run clean after adding the fixture.

**3. [Discretion] Correlation-kind allowlist for `same_binding`.** The plan left the exact correlation mechanism to Claude's discretion ("Correlation uses span and binding facts already present on `diagnostic.Cause` and the checked program"). Implemented as a closed allowlist of four `Cause.Kind` values (`place`, `owner`, `loan`, `transfer_target`) verified directly against every `diagnostic.Cause{Kind: ...}` construction site in `check.go` to actually carry a reusable binding/place identity in `Detail`; `type`, `declared_here`, `moved_here`, and other Kind values never correlate, since their Detail/Span do not name a reusable binding.

**Total deviations:** 2 auto-fixed/discretionary (Rule 2 + Claude's-discretion items), 1 stylistic (internal test package). **Impact:** none block phase completion; all are documented, tested, and independently re-verified not to regress existing suites.

## Issues Encountered

None blocking. One self-correction during Task 3: the first draft of `TestExplainNodeOrderIsStableOnTies` used a `Kind: "type"` correlation pair that (correctly) does NOT correlate under the allowlist above, so the constructed "tie" scenario accidentally produced identical sorted/unsorted output and did not demonstrate the sort's necessity. Caught by manually running the sort-removed probe before finalizing the test, redesigned using `Kind: "place"` (an allowlisted correlation kind) to produce a genuine out-of-insertion-order depth interleaving, then re-verified both that the corrected test fails without the sort and passes with it restored.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `lang.explain/0` is minted, tested end-to-end (unit + real shipped-binary check), and does not touch `lang.command/0`'s `Schema` constant or `Metrics`/`Lane` — the 06-06 coordinated bump remains untouched and unblocked.
- `protocol.ExplainSchema`, `EdgeCausedBy`/`EdgeNarrows`/`EdgeSameBinding`, `ExplainDefaultDepth`, `ExplainMaxNodes` are exported constants any later plan (including 06-06) can reference without re-deriving values.
- No blockers for subsequent Phase 6 plans. `lang.query/0` (the other net-new schema D-06-04 confirmed) is out of this plan's scope and remains open for a later plan.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: internal/compiler/session/session_phase6_explain.go
- FOUND: internal/compiler/session/session_phase6_explain_test.go
- FOUND: testdata/phase5/explain_use_after_move.lang
- FOUND: commit 28301ae (Task 1)
- FOUND: commit 1ad092f (Task 2)
- FOUND: commit 4a4b5c9 (Task 3)

---
phase: 06-agent-feedback-and-performance-ratification
plan: 03
subsystem: cli-protocol
tags: [addressing, cli, schema-versioning, pagination, determinism]

# Dependency graph
requires:
  - phase: 06-agent-feedback-and-performance-ratification
    plan: "06-02"
    provides: "lang.explain/0 precedent (schema-minting shape, ExplainID Finalize() fold, CLI dispatch/flag-strip pattern), and the developer-confirmed D-06-05 SRC-first command grammar this plan's `query` arm reuses verbatim"
provides:
  - "lang.query/0: a net-new, versioned command schema for one joined addressing surface over all five stable ID vocabularies already shipped in the tree"
  - "session.QueryCommandFile / session.QueryVocabularies / session.QueryKinds: the CLI seam, the closed 5-vocabulary audit surface, and the closed --kind vocabulary"
  - "protocol.Result.Query (+QueryID folded into Finalize()'s identity), mirroring the existing Explain/DebugMap pattern"
  - "A bounded, cursor-paginated fact list (protocol.QueryMaxFactsPerPage=64) with a stable truncated:query.page_bound code and a specified stable sort order on ties"
affects: [06-06-lane-schema-bump, phase-6-verification, dx-tooling-consumers-of-lang-query]

# Actuals (#2632)
actuals:
  tokens: 12649
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "single closed dispatch table (map[string]resolver) as the one source of truth for both the classifier's routing decision and the QueryVocabularies() audit list, so the two structurally cannot drift apart -- verified live by adding a fake sixth entry and watching the audit test name it, then reverting"
    - "opaque content-derived cursor (checksum + base64 JSON of the last emitted fact's sort key), never a raw offset -- decodable to re-seek by binary search over the same stable sort order rather than by position"

key-files:
  created:
    - internal/compiler/session/session_phase6_query.go
    - internal/compiler/session/session_phase6_query_test.go
  modified:
    - internal/compiler/protocol/protocol.go
    - cmd/lang/main.go

key-decisions:
  - "QueryVocabularies() returns the FIVE COARSE category names D-06-01 lists (diagnostic, debugmap, evidence, control, lane); QueryFact.Vocabulary carries a FINER label per fact (operation_id/core_id/point_id under the debugmap category, plus the diagnostic/evidence/control/lane names, plus the 'unrecognized' sentinel). This reconciles Task 1's literal instruction (fact.Vocabulary = 'operation_id') with Task 2's 'exactly five vocabularies' + 'no sixth vocabulary' audit requirement: the closed-5 contract lives at the dispatch-table/category level, the fine label is per-fact reporting detail within the debugmap category."
  - "core_id has no stable address prefix of its own (only functionID, an arbitrary program identifier) -- an address is only ever classified core_id when it matches a REAL entry.CoreID in the freshly-built debugmap for SRC. An address containing ':op:' or ':point:' is always routed as operation_id/point_id (even absent, honestly not_captured), but a bare address matching neither shape AND no real core ID is reported as the 'unrecognized' sentinel rather than a fabricated core_id guess -- this is the one place D-06-01's 'never a sixth vocabulary, never a guess' constraint required a real-corpus lookup rather than a shape rule alone."
  - "lane: vocabulary resolves against a live `verify` run over SRC (session.VerifyCorpusFile), so for this one vocabulary SRC is expected to be a corpus directory, exactly matching `lang verify CORPUS`'s own SRC operand -- consistent with every other vocabulary's 'SRC is whatever `lang verify`/`check`/`evidence` would take to reproduce this fact from scratch,' never a persisted store (D-06-02's discipline extends to query)."
  - "evidence vocabulary is implemented by calling the already-shipped session.EvidenceCommandFile directly rather than re-deriving Facts/Build a second way -- one producer of evidence identity, not two independently-maintained ones."
  - "--kind (D-06-05: symbol/type/ownership/dependency/test) semantics are Claude's Discretion (the grammar names the five kinds but not their meaning). Mapped onto REAL existing vocabulary rather than an invented taxonomy: 'type' reuses the literal diagnostic.Cause Kind value check.go already emits for type-mismatch causes; 'ownership' reuses explain's own existing place/owner/loan/transfer_target correlation-kind set plus debugmap's move/borrow_shared/borrow_exclusive/copy operation kinds; 'dependency' is the three debugmap-joined fine vocabularies (core_id/operation_id/point_id); 'test' is verify's lane facts; 'symbol' is a resolved evidence identity. Recorded here as a discretionary interpretation, not a locked contract -- a future plan needing sharper --kind semantics should treat this mapping as a starting point, not as D-06-05 itself."
  - "protocol.QueryMaxFactsPerPage = 64, not ExplainMaxNodes's 4096: query's own bounding discipline in D-06-03 is 'smallest sufficient context by default' (the compute-efficiency constitution's own phrase), so its default page is an order of magnitude below explain/debug-map's full-artifact caps rather than matching them. Claude's Discretion, recorded so a future plan does not need to re-derive the reasoning."
  - "Tasks 2 and 3 (tdd=\"true\") produced test-only commits over Task 1's already-complete, already-verified implementation -- following 06-02's own documented precedent for justified single-file/task consolidation when later tasks' behavior tests exercise one small file's shared helpers rather than requiring separate new production code. The five-vocabulary dispatcher, --kind filter, and cursor pagination were all implemented as one cohesive unit in Task 1's commit because a dispatcher that could not resolve every vocabulary, filter by kind, or page its output would already fail Task 1's own 'production quality, not a throwaway' tracer requirement."

requirements-completed: []

coverage:
  - id: D1
    description: "lang query is one joined addressing surface resolving any of the five stable ID vocabularies already shipped in the tree, under the net-new lang.query/0 schema, without minting a sixth"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryResolvesEveryStableIDVocabulary"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryMintsNoSixthVocabulary"
        status: pass
      - kind: command
        ref: "lang --json query testdata/phase2/owned_transfer.lang <real operation id>"
        status: pass
    human_judgment: false
  - id: D2
    description: "An ID that resolves to nothing returns an honest not_captured fact, never an error and never a fabricated value, across all five vocabularies"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryUnknownIDReportsNotCaptured"
        status: pass
      - kind: command
        ref: "lang --json query SRC <unknown id for each of the 5 vocabularies> (manually verified this session for diagnostic:, control:, lane:, evidence:, and operation_id shapes)"
        status: pass
    human_judgment: false
  - id: D3
    description: "Results are bounded lists paginated by --cursor, never by --depth; pages concatenate exactly once with no loss and no duplication, and a malformed cursor is a usage error"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryCursorPaginationIsBounded"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryPagesConcatenateExactlyOnce"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryMalformedCursorIsUsageError"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryDepthDoesNotPaginate"
        status: pass
    human_judgment: false
  - id: D4
    description: "Equal-comparing query results have a specified, stable order across two cold invocations"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryResultOrderIsStableOnTies"
        status: pass
    human_judgment: false
  - id: D5
    description: "--kind filters the fact list by D-06-05's closed five-value vocabulary; an unrecognized kind is a usage error, never a silent no-op"
    requirement: "DX-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_query_test.go#TestQueryKindFilterVocabularyIsClosed"
        status: pass
    human_judgment: false

duration: 70min
completed: 2026-09-07
status: complete
---

# Phase 6 Plan 3: Mint `lang.query/0` — One Joined Addressing Surface Over Five Stable ID Vocabularies Summary

**`lang query SRC ID_OR_PATTERN [--kind=K] [--cursor=C] [--json]` resolves any of `diagnostic:`, debugmap's `core_id`/`operation_id`/`point_id`, evidence IDs/digests, `control:*`, or `lane:*` through one closed dispatch table under the net-new `lang.query/0` schema, bounded and cursor-paginated, reporting honest `not_captured` for absence and an `unrecognized` sentinel for a genuinely unmatched address — no sixth vocabulary minted.**

## Performance

- **Duration:** ~70 min
- **Tasks:** 3 (all complete)
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments

- `protocol.go`: minted `QuerySchema = "lang.query/0"`, `QueryMaxFactsPerPage = 64`, and the `QueryFact`/`QuerySummary` types. `Result.Query *QuerySummary` folds into `Finalize()` as a 24-hex-character `QueryID`, mirroring `ExplainID`/`DebugMapID` exactly; `Metrics`/`Lane` were left untouched (06-06's territory) and `TestMetricsAndLaneFieldsExcludedFromIdentity`/`TestIdentityFieldEnumerationIsExhaustive` still pass unmodified. Added a `human()` render block.
- `session_phase6_query.go`: `QueryCommandFile(path, address, options)` classifies `address` by prefix/shape into one of five coarse categories (`diagnostic`, `debugmap`, `evidence`, `control`, `lane`) via a single `queryVocabularyDispatch` map that also backs `QueryVocabularies()` — the classifier and the closed-set audit read the same table, so they cannot independently drift. Each category resolver re-derives its answer from `SRC` fresh on every cold invocation (no persisted store, matching `explain`/`debug-map`'s own discipline):
  - `diagnostic:` — reuses `findExplainDiagnostic`, flattens the diagnostic root plus its `Causes` into facts (no graph — that's `explain`'s job).
  - debugmap `core_id`/`operation_id`/`point_id` — builds the map exactly like `debug-map`, extends `debugmap.Resolve`'s single-field lookup shape to `PointID`/`CoreID` entirely in this session-layer file (per the plan's instruction, `debugmap.Resolve` itself is untouched).
  - `evidence:`/64-hex digest — calls the already-shipped `EvidenceCommandFile` and matches against the manifest's own ID or content digest.
  - `control:*` — matches against `AllShippedControlIDs()`, the same live-derived set QLT-01's registry audit cross-checks against.
  - `lane:*` — matches against a real `VerifyCorpusFile` run's `Lanes`, so `SRC` is a corpus directory for this one vocabulary (matching `lang verify CORPUS`'s own operand).
  An address matching none of the five is reported as the honest `unrecognized` sentinel, never fabricated as one of the five. `--kind` filters facts by D-06-05's closed five-value vocabulary (mapped onto real existing Kind/vocabulary fields — see key-decisions); an unrecognized kind is `protocol.StatusUsage` with `tool.query_unknown_kind`. Facts are sorted by `(vocabulary, span.start, span.end, id)` before an opaque, content-derived cursor (checksum + base64 of the last-emitted fact's sort key) pages them at `QueryMaxFactsPerPage` (64); a malformed cursor is `protocol.StatusUsage` with `tool.query_malformed_cursor`, never a silent first page.
- `cmd/lang/main.go`: added the `query` dispatch arm (`runQuery`) and `extractKindAndCursor` mirroring `extractDepth`/`extractJSON`'s strip-and-report shape for `--kind=`/`--cursor=`; extended `usageResult()`'s grammar string.
- Verified end-to-end against the shipped binary for **all five vocabularies**: a real `operation_id`/`core_id`/`point_id` from `testdata/phase2/owned_transfer.lang`'s debug map, a real `diagnostic:` ID and its causes from `testdata/phase2/use_after_move.lang`, a real `evidence:` manifest ID from `testdata/phase1/toggle.lang`, a real `control:alias.false_no_alias` ID, and a real `lane:deterministic` ID from a `verify` run over `testdata/phase1` — each returns `"availability":"available"`. Re-ran every one with a made-up ID of the matching shape and confirmed `"availability":"not_captured"` with the pass exit code (never an error); a fully unstructured address returns the `"unrecognized"` sentinel.
- Manually demonstrated (during Task 2) that the no-sixth-vocabulary audit is load-bearing: added a `"sixth"` entry to `queryVocabularyDispatch`, re-ran `TestQueryMintsNoSixthVocabulary`, confirmed it failed naming `sixth` in the vocabulary list, then reverted and re-verified green.
- Manually demonstrated (during Task 3) that `sortQueryFacts`'s stable-tie ordering is load-bearing: temporarily emptied the sort function's body, re-ran `TestQueryResultOrderIsStableOnTies`, confirmed it went red (`order = [z-last a-first m-middle], want [a-first m-middle z-last]`), then restored the sort and re-verified green.
- `go build ./...`, `go vet ./...`, `go test ./...`, and `go test -race ./internal/compiler/session/... ./internal/compiler/protocol/...` are all clean at HEAD after this plan.

## Task Commits

Each task was committed atomically:

1. **Task 1: End-to-end `lang query` resolving one vocabulary** — `49384c2` (feat) — protocol.go, session_phase6_query.go (full cohesive implementation of all five vocabularies, --kind filter, and pagination), cmd/lang/main.go, plus Task-1-scoped tests (`TestQuerySummarySchemaIsMinted`, `TestQueryUnknownIDReportsNotCaptured`).
2. **Task 2: The five-vocabulary join dispatcher, with a no-sixth-vocabulary audit** — `517f5cd` (test) — the full behavior test suite for every vocabulary plus the closed-set audit and `--kind` validation, over the already-implemented dispatcher from Task 1.
3. **Task 3: Cursor pagination, bounding, and stable order on ties** — `6759a1a` (test) — pagination/concatenation/malformed-cursor/stable-order/depth-non-pagination tests, over the already-implemented pagination from Task 1.

_Note: Tasks 2 and 3 carried `tdd="true"` but produced single `test(...)` commits rather than RED/GREEN pairs, mirroring 06-02's own documented precedent for justified single-commit consolidation: the five-vocabulary dispatcher, `--kind` filter, and cursor pagination were implemented as one cohesive unit in Task 1's commit, since a dispatcher that could not resolve every vocabulary or page its output would already have failed Task 1's own "production quality, not a throwaway" tracer requirement. There was no separate implementation step to add once Task 2/3's tests were written._

## Files Created/Modified

- `internal/compiler/protocol/protocol.go` — `QuerySchema`, `QueryMaxFactsPerPage`, `QueryFact`/`QuerySummary` types, `Result.Query` field, `QueryID` folded into `Finalize()`, `Query` render block in `human()`.
- `internal/compiler/session/session_phase6_query.go` (new) — `QueryCommandFile`, `QueryOptions`, `QueryVocabularies`, `QueryKinds`, the closed `queryVocabularyDispatch` table, per-vocabulary resolvers, `--kind` filtering, `sortQueryFacts`/cursor pagination helpers.
- `internal/compiler/session/session_phase6_query_test.go` (new) — 11 tests: schema minting, honest-absence (recognized and unrecognized), all five vocabularies against real fixtures, the no-sixth-vocabulary audit, `--kind` filtering, cursor pagination bounding/concatenation/malformed-cursor, stable order on ties, and depth-does-not-paginate.
- `cmd/lang/main.go` — `query` dispatch arm, `runQuery`, `extractKindAndCursor`, updated usage string.

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: the two-level vocabulary naming (coarse `QueryVocabularies()` category vs. fine per-fact `Vocabulary` label), the corpus-lookup-based `core_id` recognition rule (no stable prefix of its own), and the `--kind` semantic mapping onto real existing Kind/vocabulary fields rather than an invented taxonomy.

## Deviations from Plan

**1. [Discretion] `--kind` semantics mapped onto existing fields, not a new taxonomy.** D-06-05 names the five `--kind` values (`symbol`/`type`/`ownership`/`dependency`/`test`) but leaves their meaning to Claude's Discretion. Rather than inventing new classification logic, each kind was mapped onto real, already-emitted vocabulary: `type` matches the literal `diagnostic.Cause.Kind == "type"` value `check.go` already produces; `ownership` reuses `explain`'s own `place`/`owner`/`loan`/`transfer_target` correlation-kind set plus debugmap's `move`/`borrow_shared`/`borrow_exclusive`/`copy` operation kinds; `dependency` is the three debugmap-joined fine vocabularies; `test` is verify's lane facts; `symbol` is a resolved evidence identity. Documented in `key-decisions` as a discretionary starting interpretation, not a locked contract.

**2. [Discretion] `QueryMaxFactsPerPage = 64`, not `ExplainMaxNodes`'s 4096.** The plan named "the closest in-tree bound" as guidance; `explain`/`debug-map`'s 4096 is a full-artifact cap, while `query`'s own bounding discipline (D-06-03) is explicitly "smallest sufficient context by default." Chose an order of magnitude below the existing caps rather than matching them, and recorded the reasoning so a future plan does not need to re-derive it.

**3. [Discretion] `core_id` recognition requires a real corpus match, not a shape rule.** Unlike `operation_id`/`point_id` (both contain a stable `:op:`/`:point:` marker), a bare `core_id` is just a function's own identifier with no distinguishing prefix. Rather than guessing (which risks fabricating a "core_id" classification for garbage input), an address only classifies as `core_id` when it equals a real `entry.CoreID` in the freshly-built debugmap for `SRC`; otherwise it falls through to the honest `unrecognized` sentinel. This is the one vocabulary where "recognized shape" and "recognized value" collapse into the same check.

**Total deviations:** 3 discretionary (all Claude's-Discretion items the plan explicitly delegated). **Impact:** none block phase completion; all are documented, tested, and independently re-verified not to regress existing suites.

## Issues Encountered

None blocking. One design tension resolved during Task 2 planning: the plan's Task 1 action text says "vocabulary set to `operation_id`" for a `QueryFact`, while Task 2's acceptance criteria requires `QueryVocabularies()` to return "exactly five" entries and pins the closed set to D-06-01's five coarse names (`diagnostic`, debugmap's three ID kinds treated as one, `evidence`, `control`, `lane`). Resolved by splitting the concept in two: `QueryVocabularies()`/the dispatch table operate at the coarse (5-entry) category level for the "no sixth vocabulary" audit, while `QueryFact.Vocabulary` carries the finer per-fact label (`operation_id`/`core_id`/`point_id` within the `debugmap` category) for reporting precision. Recorded in `key-decisions` so this is not re-litigated.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- `lang.query/0` is minted, tested end-to-end (unit + real shipped-binary check across all five vocabularies), and does not touch `lang.command/0`'s `Schema` constant or `Metrics`/`Lane` — the 06-06 coordinated bump remains untouched and unblocked.
- `protocol.QuerySchema`, `QueryMaxFactsPerPage`, `QueryFact`, `QuerySummary` are exported for any later plan (including 06-06) to reference without re-deriving values.
- DX-02's remaining unresolved edge probe (carried from 06-02, category `unclassified`) remains unresolved and is surfaced here again, not dropped. Per the plan's success criteria, DX-02 is NOT marked complete by this plan alone — 06-02 already completed it in isolation and this plan's own frontmatter declares no new `requirements-completed` entries, since DX-02 was already satisfied and 06-15 also declares it (shared-ID gate, `requirements.ready-ids`).
- No blockers for subsequent Phase 6 plans.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-07*

## Self-Check: PASSED

- FOUND: internal/compiler/session/session_phase6_query.go
- FOUND: internal/compiler/session/session_phase6_query_test.go
- FOUND: commit 49384c2 (Task 1)
- FOUND: commit 517f5cd (Task 2)
- FOUND: commit 6759a1a (Task 3)

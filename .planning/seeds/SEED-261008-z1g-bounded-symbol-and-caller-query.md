---
id: SEED-261008-z1g
status: dormant
planted: 2026-10-08
planted_during: M005/Phase 27 quick task
trigger_when: A named reviewer or agent task shows an accuracy or material cost gap with current bounded schway query and ordinary search.
scope: Extend the current source-bound query seam only for one named consumer; do not begin with a general program database.
---

# SEED-261008-z1g: Bounded symbol and caller query

## Why This Matters

An agent reviewing `examples/phase24/transfer.schway` may need a helper’s declared contract, definition, and every direct call occurrence. Current source query has five stable-ID vocabularies, 64-fact pagination, honest `not_captured`, and cold rechecking, but no named-symbol or direct-caller lookup. This is a bounded gap, not evidence for replacing LSP or introducing SQLite/Datalog.

## When to Surface

**Trigger:** A measured reviewer/agent task fails or costs materially more because current `schway query` and ordinary search cannot supply the answer efficiently and accurately. Official LSP workspace symbols and call hierarchy should be considered; Schway is not asserted to implement LSP.

## Smallest Consumer and Obligations

Start with one source unit and one query: helper name → declaration/definition → all direct callsite occurrences. Reuse `internal/compiler/session/session_phase6_query.go`, stable IDs/spans, `core.Interface`, and the existing source digest. `internal/compiler/callgraph/callgraph.go` deduplicates caller/callee adjacency and stores a representative operation; enumerate operations separately before claiming all callsites. Bind each page/cursor to source revision and query, reject stale input, and report incompleteness explicitly.

## Independent Answers and Negative Controls

Compare literal expected results from independently inspected source against query output. Include repeated calls from one caller, overloaded/colliding names if admitted, unknown IDs, malformed cursors, changed source between pages, and truncation. Never let truncated output claim a complete callsite set. Compare against `rg` and current JSON behavior for answer bytes, tool calls, successful task completion, and cold/warm p50/p95 across source sizes. Reject added machinery if search is as reliable and cheaper or independent expected answers cannot be established.

## Cost, Owner, and Next Action

**Cost:** A source-only, bounded query experiment with explicit recheck and pagination cost; no persistent database, schema, or general query language.  
**Owner/next action:** The compiler/query maintainer for the selected consumer phase should, at a relevant planning transition, name the consumer and capture a current failure or measured task before promoting through roadmap/requirements ownership.

## Supersession and Breadcrumbs

Supersede if existing query or an available LSP capability closes the need. Reject if exactness or source binding cannot be preserved. Regression anchors: `TestQueryPagesConcatenateExactlyOnce` and `TestQueryUnknownIDReportsNotCaptured` (source inspected, not rerun). Before adding concurrent automatic repair, investigate a source-digest precondition in `cmd/schway-repair/repair.go:applyRepair`; a bounds check after rereading does not establish that the original source is unchanged. Related current homes: `.planning/PRODUCT-ROADMAP.md` (locality and byte/module reuse) and `D-13-02b` (separate compilation). Source appraisal: [Dashbit assessment](../sources/DASHBIT-AI-ERA.md). Code: `internal/compiler/session/session_phase6_query.go`, `internal/compiler/callgraph/callgraph.go`, `internal/compiler/cache/probe_test.go` (`TestNoClosureDigestInCache`), `cmd/schway-repair/repair.go`. LSP references: [workspace/symbol](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.17/workspace/symbol.md), [call hierarchy](https://github.com/microsoft/language-server-protocol/blob/gh-pages/_specifications/lsp/3.17/language/callHierarchy.md).

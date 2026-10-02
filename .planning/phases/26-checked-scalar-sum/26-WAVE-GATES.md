# Phase 26 wave gates

## Wave 1 — 2026-10-02

Plan 26-01's focused hosted run 37073989672 passes on Ubuntu and macOS at
`b2034083c2274a2d1d301394def0835047dd5eda`. Summary and task commits exist;
the source witness and scalar checker helper exist; the summary self-check
passes. Schema drift and UI gates report no block; the codebase drift gate
has no structure map to inspect.

The broader integration run 37074516513 at
`1de217f878a6253415c1524d2190d40cf310086f` passes vet and build on both hosts,
but fails the full suites and evidence aggregates. No full-suite pass or race
pass is claimed. Failures identify:

- Four new scalar-operation fields absent from the closed core inventory.
- The emitter consumer moving from line 979 to 1006 in core_test.go.
- The new source witness changing the corpus to 152 programs / 4,948 lines.
- M004 archive moves absent from the Phase 23 contract lookup, Phase 24
  evidence links, and Phase 23/25 groundedness locator records.
- Phase 26's still-draft validation document and test selectors owned by the
  four remaining plans.

The first four categories are repaired in this wave's integration correction.
D-14-144 receives a dated locator amendment; historical validation documents
and receipts are preserved. Its generated reconciliation row is re-derived
from the authored record. The repairs have source/diff checks only until the
next hosted run.

The active phase's unimplemented test selectors and draft status remain
visible failures. Execution continues to their named owning plans under the
project's autonomous execution setting; no test is suppressed or marked
passed. The final phase gate must rerun full hosted validation after those
plans and the validation document are complete. Wave 1 full-gate failure
count: 1.

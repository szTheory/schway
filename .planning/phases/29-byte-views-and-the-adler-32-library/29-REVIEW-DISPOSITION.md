---
phase: 29
review: 29-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: open
    title: "Fixture-owner admission is advertised as test-only but reachable from any user program and manifest"
  - id: WR-02
    severity: warning
    disposition: open
    title: "A `ByteView` entry passes `check` but cgen emits an ill-formed initializer and the interpreter cannot bind the input"
  - id: WR-03
    severity: warning
    disposition: open
    title: "Interpreter and native disagree on non-bounds defects with a live fixture owner"
  - id: WR-04
    severity: warning
    disposition: open
    title: "New law-disable switches and a C-source rewrite hook live in production packages"
  - id: IN-01
    severity: info
    disposition: open
    title: "Fixture admission code in check is unreachable defense"
  - id: IN-02
    severity: info
    disposition: open
    title: "Check's scalar fixed point has no operand-type law for byte ops"
  - id: IN-03
    severity: info
    disposition: open
    title: "The fixture lifecycle pin lives only in straight-line replay and flat-order laws assume program order"
  - id: IN-04
    severity: info
    disposition: open
    title: "The shipped `adler32_fixture.schway` check-passes but cannot be run through `run` or `evidence`"
open: 8
total: 8
recorded: 2026-10-12T00:20:20.257Z
---

# Phase 29: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | open | - |
| WR-02 | warning | open | - |
| WR-03 | warning | open | - |
| WR-04 | warning | open | - |
| IN-01 | info | open | - |
| IN-02 | info | open | - |
| IN-03 | info | open | - |
| IN-04 | info | open | - |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.

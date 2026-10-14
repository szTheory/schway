---
phase: 34
review: 34-REVIEW.md
titles: json
findings:
  - id: WR-01
    severity: warning
    disposition: fixed
    title: "Source-checking routes are missing from limit applicability"
  - id: WR-02
    severity: warning
    disposition: fixed
    title: "Linear-binding limit reports per-body scope while enforcing an aggregate"
  - id: WR-03
    severity: warning
    disposition: fixed
    title: "Surface limit witnesses do not independently check route behavior"
open: 0
total: 3
recorded: 2026-10-14T23:39:49.827Z
---

# Phase 34: Code Review Disposition

| Finding | Severity | Disposition | Source |
|---------|----------|-------------|--------|
| WR-01 | warning | fixed | 34-REVIEW-FIX.md (not in the current review) |
| WR-02 | warning | fixed | 34-REVIEW-FIX.md (not in the current review) |
| WR-03 | warning | fixed | 34-REVIEW-FIX.md (not in the current review) |

Dispositions: `open` (recorded, not yet triaged), `fixed`, `skipped`, `deferred`.
Set `deferred` by hand and put the reason in the Source cell; both are preserved. A `|` in the reason is kept as prose and escaped on the next run.
Re-running the gate keeps every row it can. A row the current review no longer reports is kept and its Source cell flagged, so a finding does not leave this record silently. ONE exception: when a finding id is REUSED by a different finding, the earlier decision cannot keep a row — the id is taken — and it is dropped. A RECORDED decision (anything but `open`) is named on the console when that happens; a row still at `open` is replaced silently, because `open` records no decision to lose.
